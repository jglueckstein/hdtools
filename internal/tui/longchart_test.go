package tui

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func nov1990() time.Time {
	return time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
}

func plotHas(view string, ch rune) bool {
	for _, line := range strings.Split(visible(view), "\n") {
		rs := []rune(line)
		if len(rs) > 6 && unicode.IsDigit(rs[4]) {
			for _, r := range rs[6:] {
				if r == ch {
					return true
				}
			}
		}
	}
	return false
}

func longPlotCols(view string) int {
	for _, line := range strings.Split(visible(view), "\n") {
		if !strings.ContainsAny(line, `/\`) && !strings.Contains(line, "-") {
			continue
		}
		// Gutter is "%5.1f-" (6 runes).
		rs := []rune(line)
		if len(rs) > 6 && unicode.IsDigit(rs[4]) {
			return len(rs) - 6
		}
	}
	return 0
}

func TestLongChartOpensFromListAndEscapes(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View())
	if !strings.Contains(view, "Quarterly") {
		t.Fatalf("want quarterly: %q", view)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View()), "daily log") {
		t.Fatalf("esc not list: %q", app.View())
	}
}

func TestLongChartOpensFromMonthAndEscapes(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "l")
	if !strings.Contains(visible(app.View()), "Quarterly") {
		t.Fatalf("want quarterly: %q", app.View())
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View()), "November 1990") || strings.Contains(visible(app.View()), "Quarterly") {
		t.Fatalf("esc not month sheet: %q", app.View())
	}
}

func TestLongChartLWhileEditingIsText(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m")
	app.month.col = colNote
	app.month.beginEdit("")
	press(app, "l")
	if strings.Contains(visible(app.View()), "Quarterly") {
		t.Fatal("l opened long chart while editing")
	}
	if !strings.Contains(app.month.input.Value(), "l") {
		t.Fatalf("note = %q", app.month.input.Value())
	}
}

func TestLongChartOpensFromMonthlyChartAndEscapes(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c", "l")
	if !strings.Contains(visible(app.View()), "Quarterly") {
		t.Fatalf("want quarterly: %q", app.View())
	}
	press(app, "esc")
	v := visible(app.View())
	if !strings.Contains(v, "November 1990") || strings.Contains(v, "Quarterly") {
		t.Fatalf("esc not monthly chart: %q", app.View())
	}
}

func TestLongChartCyclesKinds(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	want := []string{"Semiannual", "Annual", "Complete", "Quarterly"}
	for _, kind := range want {
		press(app, "]")
		if !strings.Contains(visible(app.View()), kind) {
			t.Fatalf("after ] want %s: %q", kind, app.View())
		}
	}
	press(app, "[")
	if !strings.Contains(visible(app.View()), "Complete") {
		t.Fatalf("after [ want Complete: %q", app.View())
	}
}

func TestLongChartTwoLinesNoMarksOrStems(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View())
	if !strings.ContainsAny(view, `/\-`) {
		t.Fatalf("missing path: %q", view)
	}
	if plotHas(view, 'o') {
		t.Fatalf("daily mark on long chart: %q", view)
	}
	if plotHas(view, '|') {
		t.Fatalf("stem on long chart: %q", view)
	}
}

func TestLongChartWeightAndTrendColors(t *testing.T) {
	enableChroma(t)
	freezeToday(t, nov1990())
	store := openStore(t)
	cfg := schemeCfg(map[string]string{"weight": "green", "trend": "yellow"})
	w := 80.0
	w2 := 79.0
	for _, spec := range []struct {
		day time.Time
		w   float64
	}{
		{time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), w},
		{time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), w2},
	} {
		log, err := dailylog.New(spec.day, &spec.w, 8, 0, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(context.Background(), log); err != nil {
			t.Fatal(err)
		}
	}
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "l")
	view := app.View()
	if !hasIndexedForeground(view, 2) {
		t.Fatalf("missing green daily: %q", view)
	}
	if !hasIndexedForeground(view, 3) {
		t.Fatalf("missing yellow trend: %q", view)
	}
	if !hasSGRCode(view, 1) {
		t.Fatalf("missing bold trend: %q", view)
	}
}

func TestLongChartNoColorKeepsTwoLines(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := app.View()
	vis := visible(view)
	if !strings.ContainsAny(vis, `/\-`) {
		t.Fatalf("missing path: %q", view)
	}
	if hasChromaticSGR(view) {
		t.Fatalf("chroma under NO_COLOR: %q", view)
	}
	if !hasSGRCode(view, 1) {
		t.Fatalf("trend not bold: %q", view)
	}
}

func TestLongChartEmpty(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View())
	if !strings.Contains(view, "empty") {
		t.Fatalf("empty copy missing: %q", view)
	}
	if strings.Contains(view, "Loss:") {
		t.Fatalf("loss on empty: %q", view)
	}
}

func TestLongChartQuarterlyEmptyDespiteLogs(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	log, err := dailylog.New(time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC), nil, 8, 0, false, "note")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View())
	if !strings.Contains(view, "empty") {
		t.Fatalf("want empty quarterly: %q", view)
	}
	if strings.Contains(view, "Loss:") {
		t.Fatalf("loss on empty quarterly: %q", view)
	}
}

func TestLongChartBracketsDoNotChangeSheetMonth(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "l", "]", "esc")
	view := visible(app.View())
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("sheet month moved: %q", view)
	}
}

func TestLongChartAxisUsesDisplayUnit(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "l")
	if !strings.Contains(visible(app.View()), "lb") {
		t.Fatalf("axis not lb: %q", app.View())
	}
}

func TestLongChartKeysDoNotWrite(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l", "x", "tab", " ")
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("wrote rows: %d", len(logs))
	}
}

func TestLongChartLossAndDeficit(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 29, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	w80, w70 := 80.0, 70.0
	first, err := dailylog.New(time.Date(1990, 9, 1, 0, 0, 0, 0, time.UTC), &w80, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	last, err := dailylog.New(time.Date(1990, 11, 29, 0, 0, 0, 0, time.UTC), &w70, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), last); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View())
	if !strings.Contains(view, "Loss:") {
		t.Fatalf("missing Loss: %q", view)
	}
	if !strings.Contains(view, "1.0 kg") {
		t.Fatalf("want Loss 1.0 kg: %q", view)
	}
	if !strings.Contains(view, "86") {
		t.Fatalf("want 86 cal: %q", view)
	}
}

func TestLongChartTitleKindAndSpan(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l", "]", "]")
	view := visible(app.View())
	if !strings.Contains(view, "Annual") {
		t.Fatalf("want Annual: %q", view)
	}
	if !strings.Contains(view, "December 1989") || !strings.Contains(view, "November 1990") {
		t.Fatalf("want span months: %q", view)
	}
	if strings.Contains(view, "hdtools") {
		t.Fatalf("hdtools in long chart: %q", view)
	}
}

func TestLongChartCompleteStartsAtFirstLog(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	w := 80.0
	first, err := dailylog.New(time.Date(1989, 4, 15, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	app := twoDayApp(t, store)
	press(app, "l", "]", "]", "]")
	view := visible(app.View())
	if !strings.Contains(view, "Complete") {
		t.Fatalf("want Complete: %q", view)
	}
	if !strings.Contains(view, "April 1989") {
		t.Fatalf("want April 1989: %q", view)
	}
}

func TestLongChartAnnualIsBucketed(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l", "]", "]")
	view := app.View()
	n := longPlotCols(view)
	if n != 72 {
		t.Fatalf("plot cols = %d, want 72\n%s", n, visible(view))
	}
	if plotHas(view, 'o') {
		t.Fatalf("daily mark: %q", view)
	}
}

func TestLongChartQuarterlyClipsAtToday(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View())
	if !strings.Contains(view, "September 1990") || !strings.Contains(view, "November 1990") {
		t.Fatalf("quarterly span: %q", view)
	}
}

func TestLongChartPastDatabaseUsesLatestLog(t *testing.T) {
	freezeToday(t, time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View())
	if strings.Contains(view, "empty") {
		t.Fatalf("empty in 2026: %q", view)
	}
	if !strings.Contains(view, "September 1990") || !strings.Contains(view, "November 1990") {
		t.Fatalf("want 1990 quarter: %q", view)
	}
}

func TestLongChartHelpSaysKind(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	if !strings.Contains(visible(app.View()), "[ ] kind") {
		t.Fatalf("help cue missing: %q", app.View())
	}
}

func TestLongChartXLabelsQuarterlyOneLine(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View())
	for _, want := range []string{"Sep 90", "Oct 90", "Nov 90"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in %q", want, view)
		}
	}
}

func TestLongChartXLabelsCompleteTwoLine(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	w := 80.0
	first, err := dailylog.New(time.Date(1989, 4, 15, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	app := twoDayApp(t, store)
	press(app, "l", "]", "]", "]")
	view := visible(app.View())
	if strings.Contains(view, "Apr 89") {
		t.Fatalf("one-line label on dense complete: %q", view)
	}
	if !strings.Contains(view, "Apr") || !strings.Contains(view, "89") {
		t.Fatalf("want two-line Apr/89: %q", view)
	}
}

func TestLongChartXLabelsSkipMonths(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	w := 80.0
	first, err := dailylog.New(time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	app := twoDayApp(t, store)
	press(app, "l", "]", "]", "]")
	ticks := placeLongXLabels(monthTicks(sheetRange(app.logs, time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC)), 72), 72)
	if len(ticks) == 0 {
		t.Fatal("no X labels")
	}
	months := 0
	for d := time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC); !d.After(time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC)); d = d.AddDate(0, 1, 0) {
		months++
	}
	if len(ticks) >= months {
		t.Fatalf("placed %d labels for %d months, want skips", len(ticks), months)
	}
	for i := 1; i < len(ticks); i++ {
		if ticks[i].col < ticks[i-1].col+3 {
			t.Fatalf("overlap %v then %v", ticks[i-1], ticks[i])
		}
		if !ticks[i].two {
			t.Fatalf("expected two-line at col %d", ticks[i].col)
		}
	}
}

func TestLongChartXLabelsMidMonthStart(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	w := 80.0
	first, err := dailylog.New(time.Date(1989, 4, 15, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	app := twoDayApp(t, store)
	press(app, "l", "]", "]", "]")
	view := visible(app.View())
	if !strings.Contains(view, "Apr") || !strings.Contains(view, "89") {
		t.Fatalf("want Apr 89 at start: %q", view)
	}
}

func TestLongChartEmptyHasNoMonthXLabels(t *testing.T) {
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View())
	for _, m := range []string{"Jan ", "Feb ", "Mar ", "Apr ", "May ", "Jun ", "Jul ", "Aug ", "Sep ", "Oct ", "Nov ", "Dec "} {
		if strings.Contains(view, m) {
			t.Fatalf("X label %q on empty: %q", m, view)
		}
	}
}

func TestPlaceLongXLabelsModes(t *testing.T) {
	t.Parallel()
	one := placeLongXLabels([]monthTick{
		{col: 0, mon: "Sep", yy: "90"},
		{col: 30, mon: "Oct", yy: "90"},
		{col: 61, mon: "Nov", yy: "90"},
	}, 71)
	if len(one) != 3 || one[0].two || one[1].two {
		t.Fatalf("one-line: %#v", one)
	}
	two := placeLongXLabels([]monthTick{
		{col: 0, mon: "Apr", yy: "89"},
		{col: 4, mon: "May", yy: "89"},
		{col: 8, mon: "Jun", yy: "89"},
	}, 72)
	if len(two) < 2 || !two[0].two || !two[1].two {
		t.Fatalf("two-line: %#v", two)
	}
	skip := placeLongXLabels([]monthTick{
		{col: 0, mon: "Jan", yy: "88"},
		{col: 2, mon: "Feb", yy: "88"},
		{col: 4, mon: "Mar", yy: "88"},
		{col: 6, mon: "Apr", yy: "88"},
	}, 72)
	if len(skip) >= 4 {
		t.Fatalf("want skips: %#v", skip)
	}
	if skip[0].col != 0 || !skip[0].two {
		t.Fatalf("first %#v", skip[0])
	}
}
