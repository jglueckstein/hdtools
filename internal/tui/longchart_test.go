package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func nov1990() time.Time {
	return time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
}

func TestLongChartOpensFromListAndEscapes(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View().Content)
	if !strings.Contains(view, "Quarterly") {
		t.Fatalf("want quarterly: %q", view)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View().Content), "daily log") {
		t.Fatalf("esc not list: %q", app.View().Content)
	}
}

func TestLongChartOpensFromMonthAndEscapes(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "m", "l")
	if !strings.Contains(visible(app.View().Content), "Quarterly") {
		t.Fatalf("want quarterly: %q", app.View().Content)
	}
	press(app, "esc")
	if !strings.Contains(visible(app.View().Content), "November 1990") || strings.Contains(visible(app.View().Content), "Quarterly") {
		t.Fatalf("esc not month sheet: %q", app.View().Content)
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
	if strings.Contains(visible(app.View().Content), "Quarterly") {
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
	if !strings.Contains(visible(app.View().Content), "Quarterly") {
		t.Fatalf("want quarterly: %q", app.View().Content)
	}
	press(app, "esc")
	v := visible(app.View().Content)
	if !strings.Contains(v, "November 1990") || strings.Contains(v, "Quarterly") {
		t.Fatalf("esc not monthly chart: %q", app.View().Content)
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
		if !strings.Contains(visible(app.View().Content), kind) {
			t.Fatalf("after ] want %s: %q", kind, app.View().Content)
		}
	}
	press(app, "[")
	if !strings.Contains(visible(app.View().Content), "Complete") {
		t.Fatalf("after [ want Complete: %q", app.View().Content)
	}
}

func TestLongChartTwoLinesNoMarksOrStems(t *testing.T) {
	enableChroma(t)
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	assertLongPicture(t, app.View().Content)
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
	view := app.View().Content
	written := writtenANSI(t, view)
	if !hasIndexedForeground(written, 2) {
		t.Fatalf("missing green daily: %q", written)
	}
	if !hasIndexedForeground(written, 3) {
		t.Fatalf("missing yellow trend: %q", written)
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
	view := app.View().Content
	if hasChromaticSGR(view) {
		t.Fatalf("chroma under NO_COLOR: %q", view)
	}
	assertLongPicture(t, view)
}

func TestLongChartEmpty(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View().Content)
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
	view := visible(app.View().Content)
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
	view := visible(app.View().Content)
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
	if !strings.Contains(visible(app.View().Content), "lb") {
		t.Fatalf("axis not lb: %q", app.View().Content)
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

func quarterlyLossView(t *testing.T, today time.Time, cfg config.Config) string {
	t.Helper()
	freezeToday(t, today)
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	upsertKG(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), 70)
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	press(app, "l")
	return quarterlyView(t, app)
}

func TestLongChartFitOnOnlinePicture(t *testing.T) {
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	t.Run("kilograms", func(t *testing.T) {
		view := quarterlyLossView(t, today, config.Default())
		vis := visible(view)
		assertHeadingShape(t, view, "Quarterly")
		if strings.Contains(vis, "September 1990") || strings.Contains(vis, "November 1990") {
			t.Fatalf("heading still names the span: %q", vis)
		}
		if strings.Contains(vis, "Monthly loss:") {
			t.Fatalf("says Monthly loss: %q", vis)
		}
		assertLongPicture(t, view)
		assertTickLabels(t, view, []string{
			"81", "80", "79", "78", "77", "76", "75", "74", "73", "72", "71", "70", "69",
		})
		assertCaption(t, view, "Loss: 0.5 kg   Daily deficit: 421 cal")
	})
	t.Run("pounds", func(t *testing.T) {
		cfg := config.Default()
		cfg.DisplayUnit = units.Pound
		view := quarterlyLossView(t, today, cfg)
		assertCaption(t, view, "Loss: 1.1 lb   Daily deficit: 421 cal")
	})
}

func TestLongChartOmitsShortSeries(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 1, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := quarterlyView(t, app)
	vis := visible(view)
	if strings.Contains(vis, "Loss:") || strings.Contains(vis, "Daily deficit:") {
		t.Fatalf("caption on one trend value: %q", vis)
	}
}

func TestLongChartTitleKindAndSpan(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l", "]", "]")
	view := visible(app.View().Content)
	assertHeadingShape(t, app.View().Content, "Annual")
	if !strings.Contains(view, "Dec 89") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("want end months: %q", view)
	}
	if strings.Contains(view, "December 1989") || strings.Contains(view, "November 1990") {
		t.Fatalf("heading still names the span: %q", view)
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
	view := visible(app.View().Content)
	assertHeadingShape(t, app.View().Content, "Complete")
	if !strings.Contains(view, "Apr 89") {
		t.Fatalf("want Apr 89: %q", view)
	}
	if strings.Contains(view, "April 1989") {
		t.Fatalf("heading still names the span: %q", view)
	}
}

func TestLongChartAnnualKeepsWeightLine(t *testing.T) {
	freezeToday(t, time.Date(1990, 12, 15, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	upsertKG(t, store, time.Date(1990, 11, 30, 0, 0, 0, 0, time.UTC), 90)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l", "]", "]")
	view := app.View().Content
	assertHeadingShape(t, view, "Annual")
	assertLongPicture(t, view)
	assertTickLabels(t, view, []string{
		"91", "90", "89", "88", "87", "86", "85", "84", "83", "82", "81", "80", "79",
	})
}

func TestLongChartAnnualHeadingIsTheWindow(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l", "]", "]")
	view := app.View().Content
	assertLongPicture(t, view)
	if strings.Contains(visible(view), "December 1989") {
		t.Fatalf("heading still names the span: %q", visible(view))
	}
}

func TestLongChartQuarterlyClipsAtToday(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View().Content)
	if !strings.Contains(view, "Sep 90") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("quarterly span: %q", view)
	}
	if strings.Contains(view, "September 1990") || strings.Contains(view, "November 1990") {
		t.Fatalf("heading still names the span: %q", view)
	}
}

func TestLongChartPastDatabaseUsesLatestLog(t *testing.T) {
	freezeToday(t, time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View().Content)
	if strings.Contains(view, "empty") {
		t.Fatalf("empty in 2026: %q", view)
	}
	if !strings.Contains(view, "Sep 90") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("want 1990 quarter: %q", view)
	}
	if strings.Contains(view, "September 2026") || strings.Contains(view, "2026") {
		t.Fatalf("followed the clock instead of the log: %q", view)
	}
}

func TestLongChartHelpSaysKind(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	if !strings.Contains(visible(app.View().Content), "[ ] kind") {
		t.Fatalf("help cue missing: %q", app.View().Content)
	}
}

func TestLongChartXLabelsQuarterlyOneLine(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := visible(app.View().Content)
	for _, want := range []string{"Sep 90", "Oct 90", "Nov 90"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in %q", want, view)
		}
	}
}

func TestLongChartXLabelsCompleteOneLine(t *testing.T) {
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
	view := visible(app.View().Content)
	if !strings.Contains(view, "Apr 89") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("want one-line Apr 89 and Nov 90: %q", view)
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
	view := visible(app.View().Content)
	if !strings.Contains(view, "Jan 88") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("want the end months: %q", view)
	}
	if strings.Contains(view, "Feb 88") {
		t.Fatalf("February overlaps January: %q", view)
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
	view := visible(app.View().Content)
	if !strings.Contains(view, "Apr 89") {
		t.Fatalf("want Apr 89 at start: %q", view)
	}
}

func TestLongChartEmptyHasNoMonthXLabels(t *testing.T) {
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := visible(app.View().Content)
	for _, m := range []string{"Jan ", "Feb ", "Mar ", "Apr ", "May ", "Jun ", "Jul ", "Aug ", "Sep ", "Oct ", "Nov ", "Dec "} {
		if strings.Contains(view, m) {
			t.Fatalf("X label %q on empty: %q", m, view)
		}
	}
}

func quarterlyView(t *testing.T, app *App) string {
	t.Helper()
	view := app.View().Content
	vis := visible(view)
	if !strings.Contains(vis, "Quarterly") || strings.Contains(vis, "daily log") {
		t.Fatalf("quarterly chart not shown: %q", vis)
	}
	return view
}

func s6Quarterly(t *testing.T) *App {
	t.Helper()
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	// 80 then 90 keeps the weight line and the trend on different rows.
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	upsertKG(t, store, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC), 90)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	return app
}

func TestLongChartWideQuarterlyStays62(t *testing.T) {
	app := s6Quarterly(t)
	plot := assertLongPicture(t, quarterlyView(t, app))
	if len(plot.rows) != onlinePlotRows || len(plot.rows[0]) != onlinePlotCols {
		t.Fatalf("canvas = %d x %d, want %d x %d", len(plot.rows[0]), len(plot.rows), onlinePlotCols, onlinePlotRows)
	}
}

func TestLongChartFittedQuarterlyStays62(t *testing.T) {
	freezeToday(t, time.Date(1990, 12, 15, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 10, 1, 0, 0, 0, 0, time.UTC), 80)
	upsertKG(t, store, time.Date(1990, 11, 30, 0, 0, 0, 0, time.UTC), 90)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	assertLongPicture(t, quarterlyView(t, app))
}

func TestLongChartNoColorOnSeparatedSeries(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := s6Quarterly(t)
	view := quarterlyView(t, app)
	if hasChromaticSGR(view) {
		t.Fatalf("chromatic SGR under NO_COLOR: %q", view)
	}
	assertLongPicture(t, view)
}
