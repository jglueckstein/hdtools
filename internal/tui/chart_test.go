package tui

import (
	"context"
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

func TestChartDailyMarksAndTrendPath(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "c")
	view := visible(app.View())
	if !strings.Contains(view, "o") {
		t.Fatalf("missing daily mark: %q", view)
	}
	if !strings.ContainsAny(view, "-/\\│─┌┐└┘") {
		t.Fatalf("missing trend path: %q", view)
	}
	plotRows := 0
	for _, line := range strings.Split(view, "\n") {
		if strings.ContainsAny(line, "o-/\\") {
			plotRows++
		}
	}
	if plotRows < 8 {
		t.Fatalf("plot rows = %d, want at least 8", plotRows)
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
	if !hasIndexedForeground(view, 2) {
		t.Fatalf("missing green weight: %q", view)
	}
	if !hasIndexedForeground(view, 3) {
		t.Fatalf("missing yellow trend: %q", view)
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
	if !strings.Contains(vis, "o") || !strings.ContainsAny(vis, "-/\\") {
		t.Fatalf("series missing under NO_COLOR: %q", vis)
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
}

func TestChartDailyMarkWinsSharedCell(t *testing.T) {
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
	if strings.Contains(view, "daily log") {
		t.Fatalf("chart missing: %q", app.View())
	}
	if !strings.Contains(view, "o") {
		t.Fatalf("daily mark missing on coincident day: %q", view)
	}
}
