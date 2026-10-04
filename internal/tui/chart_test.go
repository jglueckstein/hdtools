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
	raw := app.View()
	view := visible(raw)
	if !strings.Contains(view, "1") || !strings.Contains(view, "30") {
		t.Fatalf("axis labels missing: %q", view)
	}
	// Days 2 and 3 have no weigh-in. They keep day 1's trend and
	// must not grow a mark or a stem.
	rows := mustPlot(t, raw)
	pad := yPad(units.Kilogram)
	trendBin := wantBrailleBin(80, 79-pad, 80+pad, brailleN)
	for _, day := range []int{1, 2} {
		left, right := leftBins(rows, day), rightBins(rows, day)
		if !sameSet(left, trendBin) || !sameSet(right, trendBin) {
			t.Fatalf("blank day %d left %v right %v, want only carried trend bin %d; column %q",
				day+1, left, right, trendBin, columnGlyphs(rows, day))
		}
	}
}

func TestChartUsesWeightAndTrendColors(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	app.cfg = schemeCfg(map[string]string{"weight": "magenta", "trend": "yellow"})
	app.pal = newPalette(app.cfg)
	press(app, "c")
	view := app.View()
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	rows := mustPlot(t, view)
	// Day 1 sits on the trend, so that cell has a mark and no stem.
	// Stem chrome is green; magenta is the weight role.
	_, _, shared := twoDaySeriesBins()
	mark, _, ok := binRow(rows, 0, shared, true)
	if !ok || !cellHasFG(mark, 5) || cellHasFG(mark, 2) {
		t.Fatalf("coincident mark seq %q, want magenta weight", mark.seq)
	}
	if !sameSet(leftBins(rows, 0), shared) || !sameSet(rightBins(rows, 0), shared) {
		t.Fatalf("day 1 left %v right %v, want bin %d and no stem", leftBins(rows, 0), rightBins(rows, 0), shared)
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
			// A pure stem cell has left dots and no right dot.
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
	cwd := t.TempDir()
	t.Chdir(cwd)
	dataDir, dataPDF := useDataHome(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	applyP(t, app)
	assertPDFNotCwd(t, cwd, dataPDF)
	view := visible(app.View())
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
	view := visible(app.View())
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
	view := visible(app.View())
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
	if !strings.Contains(visible(app.View()), "November 1990") || strings.Contains(visible(app.View()), "arrows move") {
		t.Fatalf("chart did not open: %q", app.View())
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
	view := visible(app.View())
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
	if app.screen != screenList || !strings.Contains(visible(app.View()), "daily log") {
		t.Fatalf("left the list: %q", app.View())
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
	if app.screen != screenLong || !strings.Contains(visible(app.View()), "Quarterly") {
		t.Fatalf("left the long-term chart: %q", app.View())
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
	if strings.Contains(visible(app.View()), "November 1990") && !strings.Contains(visible(app.View()), "arrows move") {
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
		t.Fatalf("left the form: %q", app.View())
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
	if !strings.Contains(visible(app.View()), "November 1990") || strings.Contains(visible(app.View()), "arrows move") {
		t.Fatalf("chart did not open: %q", app.View())
	}
	applyP(t, app)
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, dataPDF)
	view := visible(app.View())
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
	view := visible(app.View())
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

func TestChartBrailleStemIsGreen(t *testing.T) {
	enableChroma(t)
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.55}, kgOnDay{3, 80.85}, kgOnDay{30, 82})
	rows := mustPlot(t, monthlyView(t, app))
	// Stems 9–11 share the trend's cell (bin 8). That rune is the
	// stem, so it is green rather than the trend role.
	stem, _, ok := binRow(rows, 1, 8, false)
	if !ok || !cellHasFG(stem, 2) || cellHasFG(stem, 1) {
		t.Fatalf("S2 day 2 stem cell seq %q, want stem green", stem.seq)
	}
	mark, _, ok := binRow(rows, 1, 12, true)
	if !ok || !cellHasFG(mark, 4) {
		t.Fatalf("S2 day 2 mark cell seq %q, want weight role", mark.seq)
	}
	for _, row := range rows {
		c := row[10]
		if !isBraille(c.r) || maskOf(c.r) == 0 {
			continue
		}
		if !cellHasFG(c, 1) {
			t.Fatalf("carry day cell seq %q, want trend role", c.seq)
		}
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
	enableChroma(t)
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.50}, kgOnDay{30, 82})
	rows := mustPlot(t, monthlyView(t, app))
	if !sameSet(leftBins(rows, 1), 8, 9, 10, 11) || !sameSet(rightBins(rows, 1), 8) {
		t.Fatalf("day 2 left %v right %v, want stem dots at 9 and 10 with mark 11 and trend 8 in one row; column %q",
			leftBins(rows, 1), rightBins(rows, 1), columnGlyphs(rows, 1))
	}
	// Bins 8 through 11 share one rune. The stem wins that cell.
	shared, _, ok := binRow(rows, 1, 11, true)
	if !ok || !cellHasFG(shared, 2) || cellHasFG(shared, 4) {
		t.Fatalf("S4 shared cell seq %q, want stem green", shared.seq)
	}
	if !sameSet(leftBins(rows, 0), 7) || !sameSet(rightBins(rows, 0), 7) {
		t.Fatalf("day 1 left %v right %v, want bin 7 and no stem; column %q",
			leftBins(rows, 0), rightBins(rows, 0), columnGlyphs(rows, 0))
	}
}

func TestChartBrailleNoColorCarryIsBold(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := openNovemberChart(t, dec151990(), kgOnDay{1, 80})
	view := monthlyView(t, app)
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	rows := mustPlot(t, view)
	lit := func(day int) []cellStyle {
		var out []cellStyle
		for _, row := range rows {
			c := row[day]
			if isBraille(c.r) && maskOf(c.r) != 0 {
				out = append(out, c)
			}
		}
		return out
	}
	weighIn, carry := lit(0), lit(1)
	if len(weighIn) != 1 || len(carry) != 1 {
		t.Fatalf("weigh-in cells %d, carry cells %d, want one each; columns %q %q",
			len(weighIn), len(carry), columnGlyphs(rows, 0), columnGlyphs(rows, 1))
	}
	got, want := maskOf(weighIn[0].r), maskOf(carry[0].r)
	if got != want || got&sideBits(true) == 0 || got&sideBits(false) == 0 {
		t.Fatalf("masks weigh-in %#x carry %#x, want the same two-dot rune", got, want)
	}
	if weighIn[0].bold || !carry[0].bold {
		t.Fatalf("bold weigh-in %v carry %v, want the carry day only", weighIn[0].bold, carry[0].bold)
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
