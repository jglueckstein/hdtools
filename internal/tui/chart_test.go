package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func press(app *App, keys ...string) {
	for _, k := range keys {
		switch k {
		case "esc":
			app.Update(tea.KeyMsg{Type: tea.KeyEsc})
		case "tab":
			app.Update(tea.KeyMsg{Type: tea.KeyTab})
		case "enter":
			app.Update(tea.KeyMsg{Type: tea.KeyEnter})
		case "[":
			app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
		case "]":
			app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
		default:
			for _, r := range k {
				app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
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
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing: %q", app.View())
	}
	if strings.Contains(view, "arrows move") {
		t.Fatalf("still on month sheet: %q", view)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View()), "arrows move") {
		t.Fatalf("esc did not return to month sheet: %q", app.View())
	}
}

func TestChartOpensFromListAndEscapes(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing: %q", app.View())
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View()), "daily log") {
		t.Fatalf("esc did not return to list: %q", app.View())
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
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart title missing November 1990: %q", app.View())
	}
	if strings.Contains(view, "June 1990") {
		t.Fatalf("chart still on June 1990: %q", app.View())
	}
}

func TestChartDailyMarksAndTrendPath(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	rows := mustPlot(t, app.View())
	var left, right bool
	for _, row := range rows {
		for _, c := range row {
			if c.r == ' ' {
				continue
			}
			if !isBraille(c.r) {
				t.Fatalf("plot paints %q, want braille dots for the mark and the trend", string(c.r))
			}
			m := maskOf(c.r)
			if m&sideBits(true) != 0 {
				left = true
			}
			if m&sideBits(false) != 0 {
				right = true
			}
		}
	}
	if !left || !right {
		t.Fatalf("mark and trend dots missing: left %v right %v", left, right)
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
	view := visible(app.View())
	if !strings.Contains(view, "1") || !strings.Contains(view, "30") {
		t.Fatalf("axis labels missing: %q", view)
	}
}

func TestChartUsesWeightAndTrendColors(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	app.cfg = schemeCfg(map[string]string{"weight": "green", "trend": "yellow"})
	app.pal = newPalette(app.cfg)
	press(app, "c")
	view := app.View()
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	if !hasFGOnBraille(view, 2) {
		t.Fatalf("missing green weight on a braille cell")
	}
	if !hasFGOnBraille(view, 3) {
		t.Fatalf("missing yellow trend on a braille cell")
	}
}

func TestChartNoColorKeepsTwoSeries(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View()
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	vis := visible(view)
	if !strings.Contains(vis, "November 1990") {
		t.Fatalf("title missing under NO_COLOR: %q", vis)
	}
	mark, trend, _ := twoDaySeriesBins()
	rows := mustPlot(t, view)
	if !hasLeftBin(rows, 1, mark) || !hasRightBin(rows, 1, trend) {
		t.Fatalf("series missing under NO_COLOR: day 2 column %q", columnGlyphs(rows, 1))
	}
}

func TestChartEmptyMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	press(app, "c")
	view := strings.ToLower(visible(app.View()))
	if !strings.Contains(view, "empty") {
		t.Fatalf("empty state missing: %q", app.View())
	}
	if strings.Contains(view, "deficit") || strings.Contains(view, " cal") {
		t.Fatalf("analysis shown on empty month: %q", app.View())
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
	view := visible(app.View())
	if strings.Contains(view, "arrows move") {
		t.Fatalf("still on month sheet: %q", view)
	}
	if strings.Contains(strings.ToLower(view), "empty") {
		t.Fatalf("carry-only treated as empty: %q", view)
	}
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("want November chart: %q", view)
	}
}

func TestChartBracketChangesMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c", "[")
	if !strings.Contains(visible(app.View()), "October 1990") {
		t.Fatalf("after [: %q", app.View())
	}
	press(app, "]")
	if !strings.Contains(visible(app.View()), "November 1990") {
		t.Fatalf("after ]: %q", app.View())
	}
}

func TestChartEscKeepsBrowsedMonth(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "c", "[", "esc")
	if !strings.Contains(visible(app.View()), "October 1990") {
		t.Fatalf("sheet after esc: %q", app.View())
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
	if strings.Contains(visible(app.View()), "November 1990") && !strings.Contains(visible(app.View()), "arrows move") {
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
	view := visible(app.View())
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
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
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
	view := strings.ToLower(visible(app.View()))
	if !strings.Contains(view, "loss") || !strings.Contains(view, "cal") {
		t.Fatalf("analysis missing: %q", app.View())
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
	view := visible(app.View())
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
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("chart missing: %q", app.View())
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
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("chart missing: %q", app.View())
	}
	lo, hi, ok := chartYBounds(view)
	if !ok {
		t.Fatalf("no Y scale: %q", view)
	}
	p := yPad(units.Kilogram)
	if hi-lo < 2*p-0.15 || hi-lo > 2*p+0.15 {
		t.Fatalf("Y span = %.2f, want ~%.2f (2P)", hi-lo, 2*p)
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
	ymin, ymax, ok := chartYBounds(app.View())
	if !ok {
		t.Fatalf("no Y scale: %q", app.View())
	}
	if ymin < 167.9 || ymin > 168.2 || ymax < 176.8 || ymax > 177.2 {
		t.Fatalf("monthly Y = [%.2f, %.2f], want [168, 177]", ymin, ymax)
	}
	press(app, "l")
	ymin, ymax, ok = chartYBounds(app.View())
	if !ok {
		t.Fatalf("long-term no Y scale: %q", app.View())
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
	p := yPad(units.Kilogram)
	ymin, ymax, ok := chartYBounds(app.View())
	if !ok {
		t.Fatal("no Y scale")
	}
	if ymin < 80-p-0.05 || ymin > 80-p+0.05 || ymax < 81+p-0.05 || ymax > 81+p+0.05 {
		t.Fatalf("monthly Y = [%.3f, %.3f], want [%.3f, %.3f]", ymin, ymax, 80-p, 81+p)
	}
	press(app, "l")
	ymin, ymax, ok = chartYBounds(app.View())
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
	p := yPad(units.Stone)
	ymin, ymax, ok := chartYBounds(app.View())
	if !ok {
		t.Fatal("no Y scale")
	}
	if ymin < 12-p-0.05 || ymin > 12-p+0.05 || ymax < 12.5+p-0.05 || ymax > 12.5+p+0.05 {
		t.Fatalf("monthly Y = [%.3f, %.3f], want [%.3f, %.3f]", ymin, ymax, 12-p, 12.5+p)
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
	line := chartTitleLine(app.View())
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
	view := app.View()
	if chartTitleLine(view) == "" {
		t.Fatalf("title missing: %q", view)
	}
	if !hasIndexedForeground(view, 3) {
		t.Fatalf("missing yellow title foreground: %q", view)
	}
	if !hasSGRCode(view, 44) {
		t.Fatalf("missing blue title background: %q", view)
	}
}

func TestChartTitleSurvivesNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View()
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
	view := app.View()
	rows := mustPlot(t, view)
	mark, trend, _ := twoDaySeriesBins()
	lo, hi := mark, trend
	if lo > hi {
		lo, hi = hi, lo
	}
	var missing []int
	for b := lo + 1; b < hi; b++ {
		if !hasLeftBin(rows, 1, b) {
			missing = append(missing, b)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("day 2 missing stem bins %v between %d and %d; plotColumn %q", missing, lo, hi, plotColumn(view, 1))
	}
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
	view := app.View()
	rows := mustPlot(t, view)
	left, right := leftBins(rows, 0), rightBins(rows, 0)
	if len(left) != 1 || !sameSet(right, left...) {
		t.Fatalf("coincident day left %v right %v, want one shared bin and no stem; plotColumn %q", left, right, plotColumn(view, 0))
	}
}

func TestChartStemGreen(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := app.View()
	rows := mustPlot(t, view)
	found := false
	for _, row := range rows {
		for _, c := range row {
			m := maskOf(c.r)
			// A stem cell has left dots and no right dot. A mark in
			// the same cell would take the weight role instead.
			if !isBraille(c.r) || m&sideBits(true) == 0 || m&sideBits(false) != 0 {
				continue
			}
			if cellHasFG(c, 2) {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no green stem braille cell without a mark dot; first glyph %q", string(firstPlotGlyph(rows)))
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
	view := app.View()
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	rows := mustPlot(t, view)
	found := false
	for _, row := range rows {
		c := row[0]
		m := maskOf(c.r)
		if !isBraille(c.r) || m&sideBits(true) == 0 {
			continue
		}
		found = true
		// Default weight role is blue. The mark wins this shared dot.
		if !cellHasFG(c, 4) {
			t.Fatalf("shared dot cell seq %q, want weight color", c.seq)
		}
	}
	if !found {
		t.Fatalf("shared dot missing; day 1 column %q", columnGlyphs(rows, 0))
	}
}

func TestChartPWritesPDF(t *testing.T) {
	t.Chdir(t.TempDir())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	info, err := os.Stat("1990-11-chart.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
	view := visible(app.View())
	if !strings.Contains(view, "1990-11-chart.pdf") {
		t.Fatalf("status missing path: %q", view)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPWriteFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("1990-11-chart.pdf", 0o700); err != nil {
		t.Fatal(err)
	}
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	view := visible(app.View())
	if strings.Contains(view, "1990-11-chart.pdf") && !strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("claimed saved path on write failure: %q", view)
	}
	if !strings.Contains(strings.ToLower(view), "error") && app.err == nil && !strings.Contains(strings.ToLower(app.status), "error") {
		t.Fatalf("missing error status: view=%q status=%q err=%v", view, app.status, app.err)
	}
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") || strings.Contains(view, "daily log") {
		t.Fatalf("left the monthly chart: %q", view)
	}
}

func TestChartPIgnoredOnList(t *testing.T) {
	t.Chdir(t.TempDir())
	store := openStore(t)
	app := twoDayApp(t, store)
	applyP(t, app)
	if _, err := os.Stat("1990-11-chart.pdf"); err == nil {
		t.Fatal("p on the list wrote a PDF")
	}
	press(app, "m")
	applyP(t, app)
	if _, err := os.Stat("1990-11-chart.pdf"); err == nil {
		t.Fatal("p on the month sheet wrote a PDF")
	}
	press(app, "esc", "l")
	applyP(t, app)
	if _, err := os.Stat("1990-11-chart.pdf"); err == nil {
		t.Fatal("p on the long-term chart wrote a PDF")
	}
}

func TestChartPWhileEditingIsText(t *testing.T) {
	t.Chdir(t.TempDir())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m")
	app.month.col = colNote
	app.month.beginEdit("")
	press(app, "p")
	if _, err := os.Stat("1990-11-chart.pdf"); err == nil {
		t.Fatal("p while editing wrote a PDF")
	}
	if !strings.Contains(app.month.input.Value(), "p") {
		t.Fatalf("input = %q, want p", app.month.input.Value())
	}
	if strings.Contains(visible(app.View()), "November 1990") && !strings.Contains(visible(app.View()), "arrows move") {
		t.Fatal("p opened the chart while editing")
	}
}

func TestChartPDoesNotWriteLogs(t *testing.T) {
	t.Chdir(t.TempDir())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	if _, err := os.Stat("1990-11-chart.pdf"); err != nil {
		t.Fatalf("pdf missing: %v", err)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
}

func applyP(t *testing.T, app *App) {
	t.Helper()
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
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
	view := visible(app.View())
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
	if !strings.Contains(visible(app.View()), app.longKind.name()) {
		t.Fatalf("left the long-term chart: %q", app.View())
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
	view := app.View()
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
	rows := mustPlot(t, monthlyView(t, app))
	if w := plotWidth(rows); w != 30 {
		t.Fatalf("day columns = %d, want 30", w)
	}
	for day := 0; day < 30; day++ {
		both := false
		for _, row := range rows {
			r := row[day].r
			if r == ' ' {
				continue
			}
			if !isBraille(r) {
				t.Fatalf("day %d paints %q, want one braille rune with a left and a right dot", day+1, string(r))
			}
			m := maskOf(r)
			if m&sideBits(true) != 0 && m&sideBits(false) != 0 {
				both = true
			}
		}
		if !both {
			t.Fatalf("day %d has no braille rune with both dot columns; column %q", day+1, columnGlyphs(rows, day))
		}
	}
}

func TestChartBrailleCloseMarksBins(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	view := monthlyView(t, app)
	ymin, ymax, ok := chartYBounds(view)
	pad := yPad(units.Kilogram)
	if !ok || ymin < 80-pad-0.06 || ymin > 80-pad+0.06 || ymax < 82+pad-0.06 || ymax > 82+pad+0.06 {
		t.Fatalf("Y labels [%.2f, %.2f], want padded 80..82 kg", ymin, ymax)
	}
	rows := mustPlot(t, view)
	if !hasLeftBin(rows, 1, 12) || !hasLeftBin(rows, 2, 14) || hasLeftBin(rows, 1, 14) {
		t.Fatalf("day 2 left %v, day 3 left %v, want bins 12 and 14; day2 %q day3 %q",
			leftBins(rows, 1), leftBins(rows, 2), columnGlyphs(rows, 1), columnGlyphs(rows, 2))
	}
}

func TestChartBrailleEightByFour(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	rows := mustPlot(t, monthlyView(t, app))
	painted := 0
	for _, row := range rows {
		for _, c := range row {
			if c.r == ' ' {
				continue
			}
			painted++
			if !isBraille(c.r) {
				t.Fatalf("plot cell %q is not a braille rune with 4 vertical dots", string(c.r))
			}
		}
	}
	if painted == 0 {
		t.Fatal("plot has no braille cells for 8×4 bins")
	}
}

func TestChartBrailleStemInsideRow(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.50}, kgOnDay{30, 82})
	rows := mustPlot(t, monthlyView(t, app))
	if !sameSet(leftBins(rows, 1), 8, 9, 10, 11) || !sameSet(rightBins(rows, 1), 8) {
		t.Fatalf("day 2 left %v right %v, want stem dots at 9 and 10 with mark 11 and trend 8 in one row; column %q",
			leftBins(rows, 1), rightBins(rows, 1), columnGlyphs(rows, 1))
	}
	if !sameSet(leftBins(rows, 0), 7) || !sameSet(rightBins(rows, 0), 7) {
		t.Fatalf("day 1 left %v right %v, want bin 7 and no stem; column %q",
			leftBins(rows, 0), rightBins(rows, 0), columnGlyphs(rows, 0))
	}
}

func TestChartBrailleNoColorKeepsDots(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	view := monthlyView(t, app)
	rows := mustPlot(t, view)
	_, markRow, markOK := binRow(rows, 1, 12, true)
	_, trendRow, trendOK := binRow(rows, 1, 8, false)
	if !markOK || !trendOK || markRow == trendRow {
		t.Fatalf("day 2 mark row %d ok %v, trend row %d ok %v, want different braille cells; column %q",
			markRow, markOK, trendRow, trendOK, columnGlyphs(rows, 1))
	}
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	if !strings.Contains(visible(view), "November 1990") {
		t.Fatalf("title missing: %q", visible(view))
	}
}
