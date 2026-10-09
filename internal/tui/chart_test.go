package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func press(app *App, keys ...string) {
	for _, k := range keys {
		switch k {
		case "esc":
			app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		case "tab":
			app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		case "enter":
			app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		case "[":
			app.Update(tea.KeyPressMsg{Code: '[', Text: "["})
		case "]":
			app.Update(tea.KeyPressMsg{Code: ']', Text: "]"})
		case " ":
			app.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
		default:
			for _, r := range k {
				app.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
			}
		}
	}
}

func freezeToday(t *testing.T, tm time.Time) {
	t.Helper()
	old := nowFn
	nowFn = func() time.Time { return tm }
	t.Cleanup(func() { nowFn = old })
}

func TestChartOpensFromMonthAndEscapes(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "c")
	view := visible(app.View().Content)
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing: %q", app.View().Content)
	}
	if strings.Contains(view, "arrows move") {
		t.Fatalf("still on month sheet: %q", view)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View().Content), "arrows move") {
		t.Fatalf("esc did not return to month sheet: %q", app.View().Content)
	}
}

func TestChartOpensFromListAndEscapes(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := visible(app.View().Content)
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing: %q", app.View().Content)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View().Content), "daily log") {
		t.Fatalf("esc did not return to list: %q", app.View().Content)
	}
}

func TestChartFollowsSelectedRow(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	view := visible(app.View().Content)
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing November 1990: %q", app.View().Content)
	}
	if strings.Contains(view, "June 1990") {
		t.Fatalf("chart still on June 1990: %q", app.View().Content)
	}
}

func TestChartDailyMarksAndTrendPath(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View().Content
	assertNoFilledDiamond(t, view)
	plot := mustOnlinePlot(t, view)
	if len(plot.diamondCols()) < 2 {
		t.Fatalf("diamonds = %d, want open diamonds on the weighed days, not braille dot columns", len(plot.diamondCols()))
	}
	if !plot.hasBraille() {
		t.Fatal("missing braille trend")
	}
}

func TestChartOmitsDailyMarkOnBlankDay(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	ctx := context.Background()
	for _, spec := range []struct {
		d int
		w float64
	}{{1, 80}, {4, 79}} {
		w := spec.w
		log, err := dailylog.New(time.Date(1990, 11, spec.d, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	// Days 2 and 3 have no weigh-in. They keep day 1's trend and
	// must not grow a diamond or a stem.
	plot := mustOnlinePlot(t, app.View().Content)
	shares := bestDayShares(t, plot, 30, []int{0, 3})
	assertBlankDay(t, plot, shares, 1)
	assertBlankDay(t, plot, shares, 2)
}

func TestChartUsesWeightAndTrendColors(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	app.cfg = schemeCfg(map[string]string{"weight": "magenta", "trend": "yellow"})
	app.pal = newPalette(app.cfg)
	press(app, "c")
	view := app.View().Content
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	plot := mustOnlinePlot(t, writtenANSI(t, view))
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	for _, s := range shares {
		lo, hi := s.bounds(0)
		assertDiamondFG(t, plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }), 5)
		lo, hi = s.bounds(1)
		// The stem is green. It must not take the diamond's cell.
		assertDiamondFG(t, plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }), 5)
		lo, hi = s.bounds(2)
		assertCellsFG(t, plot.cellsWhere(lo, hi, func(c cellStyle) bool { return isBraille(c.r) }), 3, "blank-day trend")
	}
}

func TestChartNoColorKeepsTwoSeries(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View().Content
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	plot := mustOnlinePlot(t, view)
	if len(plot.diamondCols()) == 0 || !plot.hasBraille() {
		t.Fatal("NO_COLOR dropped the diamond or the braille trend")
	}
}

func TestChartEmptyMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	press(app, "c")
	view := strings.ToLower(visible(app.View().Content))
	if !strings.Contains(view, "empty") {
		t.Fatalf("empty state missing: %q", app.View().Content)
	}
	if strings.Contains(view, "deficit") || strings.Contains(view, " cal") {
		t.Fatalf("analysis shown on empty month: %q", app.View().Content)
	}
	if n := countOnlinePlotRows(app.View().Content); n != 0 {
		t.Fatalf("empty month has %d plot rows, want no 62-column plot", n)
	}
}

func TestChartCarryOnlyIsNotEmpty(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 10, 31, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m", "]", "c")
	view := visible(app.View().Content)
	if strings.Contains(view, "arrows move") {
		t.Fatalf("still on month sheet: %q", view)
	}
	if strings.Contains(strings.ToLower(view), "empty") {
		t.Fatalf("carry-only treated as empty: %q", view)
	}
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("want November chart: %q", view)
	}
	plot := mustOnlinePlot(t, app.View().Content)
	if !plot.hasBraille() {
		t.Fatal("carry-only plot has no braille trend")
	}
	if len(plot.diamondCols()) != 0 {
		t.Fatal("carry-only plot has a diamond")
	}
}

func TestChartBracketChangesMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c", "[")
	if !strings.Contains(visible(app.View().Content), "October 1990") {
		t.Fatalf("after [: %q", app.View().Content)
	}
	press(app, "]")
	if !strings.Contains(visible(app.View().Content), "November 1990") {
		t.Fatalf("after ]: %q", app.View().Content)
	}
}

func TestChartEscKeepsBrowsedMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "c", "[", "esc")
	if !strings.Contains(visible(app.View().Content), "October 1990") {
		t.Fatalf("sheet after esc: %q", app.View().Content)
	}
}

func TestChartCWhileEditingIsText(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m")
	app.month.col = colNote
	app.month.beginEdit("")
	press(app, "c")
	if strings.Contains(visible(app.View().Content), "November 1990") && !strings.Contains(visible(app.View().Content), "arrows move") {
		t.Fatal("c opened the chart while editing")
	}
	if !strings.Contains(app.month.input.Value(), "c") {
		t.Fatalf("input = %q, want c", app.month.input.Value())
	}
}

func TestChartAxisUsesDisplayUnit(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "c")
	view := visible(app.View().Content)
	if strings.Contains(view, "80.0") {
		t.Fatalf("still showing kg: %q", view)
	}
	if !strings.Contains(view, "lb") {
		t.Fatalf("missing lb: %q", view)
	}
}

func TestChartKeysDoNotWrite(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c", "x")
	app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
}

func TestChartShowsMonthlyLossAndDeficit(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := strings.ToLower(visible(app.View().Content))
	if !strings.Contains(view, "loss") || !strings.Contains(view, "cal") {
		t.Fatalf("analysis missing: %q", app.View().Content)
	}
}

func TestChartLossUsesDisplayUnit(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	ctx := context.Background()
	for i, w := range []float64{80, 79} {
		ww := w
		log, err := dailylog.New(time.Date(1990, 11, 1+i*29, 0, 0, 0, 0, time.UTC), &ww, 8, 0, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "c")
	view := visible(app.View().Content)
	if !strings.Contains(view, "cal") {
		t.Fatalf("deficit missing: %q", view)
	}
	if strings.Contains(view, " kg") {
		t.Fatalf("loss still in kg: %q", view)
	}
	if !strings.Contains(view, "lb") {
		t.Fatalf("loss not in lb: %q", view)
	}
}

func TestChartCurrentMonthDividesByElapsedDays(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	view := visible(app.View().Content)
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("chart missing: %q", app.View().Content)
	}
	if strings.Contains(view, " 11") && strings.Contains(view, " 30") {
		t.Fatalf("future days plotted: %q", view)
	}
}

func TestChartFlatSeriesHasScale(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	if !strings.Contains(visible(app.View().Content), "November 1990") {
		t.Fatalf("chart missing: %q", app.View().Content)
	}
	plot := assertTickLabels(t, app.View().Content, []string{"81", "80", "79"})
	for _, lab := range plot.tickLabels {
		if strings.Contains(lab, ".") {
			t.Fatalf("tick %q, want 79, 80, 81 and not a ±2 lb span", lab)
		}
	}
}

func chartYBounds(view string) (ymin, ymax float64, ok bool) {
	var ys []float64
	for _, line := range strings.Split(visible(view), "\n") {
		if len(line) < 6 || line[5] != '-' {
			continue
		}
		var v float64
		if _, err := fmt.Sscanf(strings.TrimSpace(line[:5]), "%f", &v); err != nil {
			continue
		}
		ys = append(ys, v)
	}
	if len(ys) < 2 {
		return 0, 0, false
	}
	return ys[len(ys)-1], ys[0], true
}

func upsertKG(t *testing.T, store *dailylog.Store, day time.Time, kg float64) {
	t.Helper()
	log, err := dailylog.New(day, &kg, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
}

func TestChartYRangePounds(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	loKG, err := units.ToKG(170, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	hiKG, err := units.ToKG(175, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), loKG)
	upsertKG(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), hiKG)
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "c")
	assertTickLabels(t, app.View().Content, []string{"176", "175", "174", "173", "172", "171", "170", "169"})
	press(app, "l")
	ymin, ymax, ok := chartYBounds(app.View().Content)
	if !ok {
		t.Fatalf("long-term no Y scale: %q", app.View().Content)
	}
	if ymin < 167.9 || ymin > 168.2 || ymax < 176.8 || ymax > 177.2 {
		t.Fatalf("long-term Y = [%.2f, %.2f], want [168, 177]", ymin, ymax)
	}
}

func TestChartYRangeKilograms(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	upsertKG(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), 81)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	assertTickLabels(t, app.View().Content, []string{"82", "81", "80", "79"})
	press(app, "l")
	p := yPad(units.Kilogram)
	ymin, ymax, ok := chartYBounds(app.View().Content)
	if !ok {
		t.Fatal("long-term no Y scale")
	}
	if ymin < 80-p-0.05 || ymin > 80-p+0.05 || ymax < 81+p-0.05 || ymax > 81+p+0.05 {
		t.Fatalf("long-term Y = [%.3f, %.3f], want [%.3f, %.3f]", ymin, ymax, 80-p, 81+p)
	}
}

func TestChartYRangeStone(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	loKG, err := units.ToKG(12.0, units.Stone)
	if err != nil {
		t.Fatal(err)
	}
	hiKG, err := units.ToKG(12.5, units.Stone)
	if err != nil {
		t.Fatal(err)
	}
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), loKG)
	upsertKG(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), hiKG)
	cfg := config.Default()
	cfg.DisplayUnit = units.Stone
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "c")
	plot := assertTickLabels(t, app.View().Content, []string{
		"12.6", "12.5", "12.4", "12.4", "12.3", "12.2", "12.1", "12.1", "12.0", "11.9",
	})
	for _, lab := range plot.tickLabels {
		if strings.Contains(lab, "st") || strings.Contains(lab, "lb") {
			t.Fatalf("tick %q, want one decimal of stones and not a ±2 lb pad", lab)
		}
	}
}

func chartTitleLine(view string) string {
	for _, line := range strings.Split(visible(view), "\n") {
		if strings.Contains(line, "November 1990") && !strings.Contains(line, "hdtools") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func plotColumn(view string, dayIndex int) string {
	var b strings.Builder
	for _, line := range strings.Split(visible(view), "\n") {
		// Day columns start after the "%5.1f-" gutter. A Braille line
		// has no o-/\|, so the gutter is what marks the plot.
		vis := []rune(line)
		if !isPlotGutter(vis) {
			continue
		}
		col := 6 + dayIndex
		if col >= 0 && col < len(vis) {
			b.WriteRune(vis[col])
		}
	}
	return b.String()
}

func TestChartTitleIsMonthYear(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	line := chartTitleLine(app.View().Content)
	if line == "" || strings.Contains(line, "hdtools") {
		t.Fatalf("title line = %q, want November 1990 without hdtools —", line)
	}
	if !strings.Contains(line, "November 1990") {
		t.Fatalf("title line = %q", line)
	}
}

func TestChartTitleBoxColours(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View().Content
	raw := assertHeadingShape(t, view, "November 1990")
	if !hasIndexedForeground(writtenANSI(t, raw), 6) {
		t.Fatal("heading is not the default cyan title role")
	}
}

func TestChartTitleSurvivesNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View().Content
	if chartTitleLine(view) == "" {
		t.Fatalf("title missing under NO_COLOR: %q", view)
	}
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
}

func TestChartStemJoinsMarkToTrend(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	plot := mustOnlinePlot(t, app.View().Content)
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	assertStemBetween(t, plot, shares, 1)
}

func TestChartNoStemWhenMarkOnTrend(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	plot := mustOnlinePlot(t, app.View().Content)
	if len(plot.diamondCols()) == 0 {
		t.Fatal("missing open diamond on the weighed day")
	}
	shares := bestDayShares(t, plot, 30, []int{0})
	assertNoStem(t, plot, shares, 0)
}

func TestChartStemGreen(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	plot := mustOnlinePlot(t, writtenANSI(t, app.View().Content))
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	found := false
	for _, s := range shares {
		lo, hi := s.bounds(1)
		stems := plot.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
		if len(stems) == 0 {
			t.Fatal("day 2 has no stem rune")
		}
		for _, c := range stems {
			if !cellHasFG(c, 2) {
				t.Fatalf("stem seq %q, want green", c.seq)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("no green stem rune")
	}
}

func TestChartDailyMarkWinsSharedCell(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	view := app.View().Content
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	plot := mustOnlinePlot(t, writtenANSI(t, view))
	diamonds := plot.cellsWhere(0, onlinePlotCols, func(c cellStyle) bool { return c.r == openDiamond })
	if len(diamonds) == 0 {
		t.Fatal("missing open diamond on the weight")
	}
	assertDiamondFG(t, diamonds, 4)
}

func TestChartPWritesPDF(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, dataPDF := useDataHome(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	assertPDFNotCwd(t, cwd, dataPDF)
	view := visible(app.View().Content)
	if !strings.Contains(view, dataDir) || !strings.Contains(view, chartPDFFile) {
		t.Fatalf("status missing data directory path: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
}

func TestChartPUsesPDFDir(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	_, dataPDF := useDataHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	press(app, "c")
	applyP(t, app)
	pdf := filepath.Join(pdfDir, chartPDFFile)
	assertPDFNotCwd(t, cwd, pdf)
	view := visible(app.View().Content)
	if !strings.Contains(view, pdfDir) || !strings.Contains(view, chartPDFFile) {
		t.Fatalf("status missing pdf_dir path: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
	assertNoPDF(t, dataPDF)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
}

func TestChartPWriteFailure(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	_, dataPDF := useDataHome(t)
	if err := os.MkdirAll(dataPDF, 0o700); err != nil {
		t.Fatal(err)
	}
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	info, err := os.Stat(dataPDF)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("destination directory was replaced")
	}
	view := visible(app.View().Content)
	if strings.Contains(view, chartPDFFile) && !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("claimed saved path on write failure: %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "error") && app.err == nil && !strings.Contains(strings.ToLower(app.status), "error") {
		t.Fatalf("missing error status: view=%q status=%q err=%v", view, app.status, app.err)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") || strings.Contains(view, "daily log") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPPDFDirIsFile(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	_, dataPDF := useDataHome(t)
	pdfDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(pdfDir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	press(app, "c")
	if !strings.Contains(visible(app.View().Content), "November 1990") || strings.Contains(visible(app.View().Content), "arrows move") {
		t.Fatalf("chart did not open: %q", app.View().Content)
	}
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, dataPDF)
	info, err := os.Stat(pdfDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatal("replaced the pdf_dir file with a directory")
	}
	view := visible(app.View().Content)
	if strings.Contains(view, chartPDFFile) && !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("claimed saved path on write failure: %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "error") && app.err == nil {
		t.Fatalf("missing error: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") || strings.Contains(view, "daily log") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPIgnoredOnList(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, _ := useDataHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	applyP(t, app)
	assertNoPDFIn(t, cwd, dataDir, pdfDir)
	if app.screen != screenList || !strings.Contains(visible(app.View().Content), "daily log") {
		t.Fatalf("left the list: %q", app.View().Content)
	}
	press(app, "m")
	applyP(t, app)
	assertNoPDFIn(t, cwd, dataDir, pdfDir)
	if app.screen != screenMonth || !app.month.editing || !strings.Contains(app.month.input.Value(), "p") {
		t.Fatalf("p on the month sheet did not stay text, screen=%v editing=%v input=%q", app.screen, app.month.editing, app.month.input.Value())
	}
	press(app, "esc", "l")
	applyP(t, app)
	assertNoPDFIn(t, cwd, dataDir, pdfDir)
	if app.screen != screenLong || !strings.Contains(visible(app.View().Content), "Quarterly") {
		t.Fatalf("left the long-term chart: %q", app.View().Content)
	}
}

func TestChartPWhileEditingIsText(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, _ := useDataHome(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m")
	app.month.col = colNote
	app.month.beginEdit("")
	press(app, "p")
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, filepath.Join(dataDir, chartPDFFile))
	if !strings.Contains(app.month.input.Value(), "p") {
		t.Fatalf("input = %q, want p", app.month.input.Value())
	}
	if strings.Contains(visible(app.View().Content), "November 1990") && !strings.Contains(visible(app.View().Content), "arrows move") {
		t.Fatal("p opened the chart while editing")
	}
}

func TestChartPIgnoredOnForm(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, _ := useDataHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	press(app, "n")
	if app.screen != screenForm {
		t.Fatalf("screen = %v, want form", app.screen)
	}
	press(app, "tab", "tab", "tab", "tab", "tab")
	applyP(t, app)
	assertNoPDFIn(t, cwd, dataDir, pdfDir)
	if app.screen != screenForm {
		t.Fatalf("left the form: %q", app.View().Content)
	}
	if !strings.Contains(app.form.inputs[4].Value(), "p") {
		t.Fatalf("note = %q, want p", app.form.inputs[4].Value())
	}
}

func TestChartPNonStringPDFDir(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	_, dataPDF := useDataHome(t)
	app := appWithConfig(t, loadConfigTOML(t, "display_unit = \"kg\"\npdf_dir = 3\n"))
	press(app, "c")
	if !strings.Contains(visible(app.View().Content), "November 1990") || strings.Contains(visible(app.View().Content), "arrows move") {
		t.Fatalf("chart did not open: %q", app.View().Content)
	}
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, dataPDF)
	view := visible(app.View().Content)
	if !strings.Contains(view, "pdf_dir") {
		t.Fatalf("error does not name pdf_dir: %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("missing error: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPOverwrite(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, dataPDF := useDataHome(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "1990-11-chart") || strings.HasSuffix(entry.Name(), ".pdf") {
			names = append(names, entry.Name())
		}
	}
	if len(names) != 1 || names[0] != chartPDFFile {
		t.Fatalf("chart files = %v, want [%s]", names, chartPDFFile)
	}
	info, err := os.Stat(dataPDF)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
}

func TestChartPNoHomeFails(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	store := openStore(t)
	app := twoDayApp(t, store)
	app.dbPath = filepath.Join(t.TempDir(), "explicit.db")
	press(app, "c")
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	view := visible(app.View().Content)
	if !strings.Contains(view, "home directory") {
		t.Fatalf("error missing home directory: %q", view)
	}
	if strings.Contains(view, chartPDFFile) && !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("claimed saved path: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPRelativeDataHomeFails(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "data")
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, filepath.Join(cwd, "data", "hdtools", chartPDFFile))
	view := visible(app.View().Content)
	if !strings.Contains(view, "not absolute") {
		t.Fatalf("error missing non-absolute directory: %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("missing error: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPDoesNotWriteLogs(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	_, dataPDF := useDataHome(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	assertPDFNotCwd(t, cwd, dataPDF)
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
}

const chartPDFFile = "1990-11-chart.pdf"

func useDataHome(t *testing.T) (dir, pdf string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	dir = filepath.Join(root, "hdtools")
	pdf = filepath.Join(dir, chartPDFFile)
	return dir, pdf
}

func configWithPDFDir(t *testing.T, pdfDir string) config.Config {
	t.Helper()
	body := "display_unit = \"kg\"\npdf_dir = " + strconv.Quote(pdfDir) + "\n"
	return loadConfigTOML(t, body)
}

func appWithConfig(t *testing.T, cfg config.Config) *App {
	t.Helper()
	app := twoDayApp(t, openStore(t))
	app.cfg = cfg
	return app
}

func assertNoPDF(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("wrote %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func assertNoPDFIn(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		assertNoPDF(t, filepath.Join(dir, chartPDFFile))
	}
}

func assertPDFNotCwd(t *testing.T, cwd, path string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(cwd, chartPDFFile)); err == nil {
		t.Fatalf("wrote %s in the working directory, want %s", chartPDFFile, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
}

func applyP(t *testing.T, app *App) {
	t.Helper()
	_, cmd := app.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	if cmd == nil {
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Fatal("p quit")
	}
	if msg != nil {
		app.Update(msg)
	}
}

func TestGotoTodayIgnoredOnChart(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	press(app, "[")
	if app.month.month != time.October {
		t.Fatalf("after [: %s, want October", app.month.month)
	}
	press(app, "t")
	if app.screen != screenChart {
		t.Fatalf("screen = %v, want chart", app.screen)
	}
	if app.month.month != time.October {
		t.Fatalf("chart month = %s, want October", app.month.month)
	}
	view := visible(app.View().Content)
	if strings.Contains(view, "daily log") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
	press(app, "l")
	kind := app.longKind
	press(app, "]")
	if app.longKind == kind {
		t.Fatal("] did not change long-term kind")
	}
	kind = app.longKind
	press(app, "t")
	if app.screen != screenLong {
		t.Fatalf("screen = %v, want long", app.screen)
	}
	if app.longKind != kind {
		t.Fatalf("longKind = %v, want %v", app.longKind, kind)
	}
	if !strings.Contains(visible(app.View().Content), app.longKind.name()) {
		t.Fatalf("left the long-term chart: %q", app.View().Content)
	}
	logs, err = store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
}

type kgOnDay struct {
	day int
	kg  float64
}

func openNovemberChart(t *testing.T, today time.Time, points ...kgOnDay) *App {
	t.Helper()
	freezeToday(t, today)
	store := openStore(t)
	for _, p := range points {
		upsertKG(t, store, time.Date(1990, 11, p.day, 0, 0, 0, 0, time.UTC), p.kg)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "c")
	return app
}

func monthlyView(t *testing.T, app *App) string {
	t.Helper()
	view := app.View().Content
	vis := visible(view)
	if strings.Contains(vis, "arrows move") || strings.Contains(vis, "daily log") || !strings.Contains(vis, "November 1990") {
		t.Fatalf("monthly chart not shown: %q", vis)
	}
	return view
}

func dec151990() time.Time {
	return time.Date(1990, 12, 15, 12, 0, 0, 0, time.UTC)
}

func TestChartBrailleBothColumnsNovember(t *testing.T) {
	app := openNovemberChart(t, dec151990(), kgOnDay{1, 80}, kgOnDay{30, 82})
	view := monthlyView(t, app)
	plot := mustOnlinePlot(t, view)
	assertNoFilledDiamond(t, view)
	cols := plot.diamondCols()
	if len(cols) != 2 {
		t.Fatalf("diamonds = %d, want two open diamonds on the 62-column plot", len(cols))
	}
	if !plot.slotHasBraille(cols[0]+1, cols[1]) {
		t.Fatal("no braille on the days between the diamonds")
	}
}

func TestChartBrailleCloseMarksBins(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	view := monthlyView(t, app)
	plot := assertTickLabels(t, view, []string{"83", "82", "81", "80", "79"})
	if len(plot.diamondCols()) != 4 {
		t.Fatalf("diamonds = %d, want one open diamond on each weighed day", len(plot.diamondCols()))
	}
}

func TestChartBrailleStemIsGreen(t *testing.T) {
	enableChroma(t)
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 100}, kgOnDay{30, 82})
	plot := mustOnlinePlot(t, writtenANSI(t, monthlyView(t, app)))
	shares := bestDayShares(t, plot, 30, []int{0, 1, 29})
	for _, s := range shares {
		lo, hi := s.bounds(1)
		stems := plot.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
		if len(stems) == 0 {
			t.Fatal("day 2 has no stem rune")
		}
		assertCellsFG(t, stems, 2, "stem")
	}
}

func TestChartBrailleEightByFour(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	plot := mustOnlinePlot(t, monthlyView(t, app))
	if len(plot.rows) != onlinePlotRows {
		t.Fatalf("plot rows = %d, want %d, not an eight-by-four braille grid", len(plot.rows), onlinePlotRows)
	}
	if len(plot.diamondCols()) == 0 {
		t.Fatal("plot has no open diamond; still an all-braille grid")
	}
}

func TestChartBrailleStemInsideRow(t *testing.T) {
	enableChroma(t)
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 100}, kgOnDay{30, 82})
	plot := mustOnlinePlot(t, monthlyView(t, app))
	shares := bestDayShares(t, plot, 30, []int{0, 1, 29})
	assertNoStem(t, plot, shares, 0)
	assertStemBetween(t, plot, shares, 1)
}

func TestChartBrailleNoColorCarryIsBold(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := openNovemberChart(t, dec151990(), kgOnDay{1, 80})
	view := monthlyView(t, app)
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	plot := mustOnlinePlot(t, view)
	shares := bestDayShares(t, plot, 30, []int{0})
	for _, s := range shares {
		lo, hi := s.bounds(0)
		diamonds := plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond })
		if len(diamonds) == 0 {
			t.Fatal("weigh-in is not an open diamond")
		}
		for _, c := range diamonds {
			if c.bold || isBraille(c.r) {
				t.Fatalf("diamond bold %v rune %q, want a non-bold open diamond", c.bold, string(c.r))
			}
		}
		bLo, bHi := s.bounds(1)
		blank := plot.cellsWhere(bLo, bHi, func(c cellStyle) bool { return isBraille(c.r) })
		if len(blank) == 0 {
			t.Fatal("blank day has no braille trend")
		}
		for _, c := range blank {
			if !c.bold {
				t.Fatal("blank-day trend is not bold")
			}
		}
	}
}

func TestChartBrailleNoColorKeepsDots(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	view := monthlyView(t, app)
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	plot := mustOnlinePlot(t, view)
	if len(plot.diamondCols()) == 0 || !plot.hasBraille() {
		t.Fatal("NO_COLOR dropped the diamond or the braille trend")
	}
	if !strings.Contains(visible(view), "November 1990") {
		t.Fatalf("title missing: %q", visible(view))
	}
}
