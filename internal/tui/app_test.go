package tui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestDefaultDBPathUsesHome(t *testing.T) {
	t.Parallel()
	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath: %v", err)
	}
	if filepath.Base(got) != "hdtools.db" {
		t.Fatalf("path = %q, want .../hdtools.db", got)
	}
	if !strings.Contains(got, "hdtools") {
		t.Fatalf("path = %q, want an hdtools directory", got)
	}
}

func TestViewEmptyStore(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "/tmp/test.db", config.Default())
	view := app.View()
	if !strings.Contains(view, "no entries yet") {
		t.Fatalf("View = %q", view)
	}
	if !strings.Contains(view, "/tmp/test.db") {
		t.Fatalf("View missing db path: %q", view)
	}
}

func TestViewListsTrendedLogs(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	ctx := context.Background()
	day := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	w := 172.5
	log, err := dailylog.New(day, &w, 8, 1000, true, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, log); err != nil {
		t.Fatal(err)
	}

	app := New(store, "mem.db", config.Default())
	msg := app.load()
	loaded, ok := msg.(loadedMsg)
	if !ok {
		t.Fatalf("load() = %T, want loadedMsg", msg)
	}
	model, _ := app.Update(loaded)
	view := model.View()
	if !strings.Contains(view, "1990-11-01") || !strings.Contains(view, "172.5") {
		t.Fatalf("View = %q", view)
	}
}

func TestNewOpensForm(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	view := model.View()
	if !strings.Contains(view, "day form") {
		t.Fatalf("View = %q", view)
	}
}

func TestListShowsPoundsWhenConfigured(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	ctx := context.Background()
	day := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	w := 80.0
	log, err := dailylog.New(day, &w, 0, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Config{DisplayUnit: units.Pound})
	msg := app.load()
	loaded, ok := msg.(loadedMsg)
	if !ok {
		t.Fatalf("load() = %T", msg)
	}
	model, _ := app.Update(loaded)
	view := model.View()
	lb, err := units.FromKG(80, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	shown := fmt.Sprintf("%.1f", lb)
	if !strings.Contains(view, shown) {
		t.Fatalf("View missing %s: %q", shown, view)
	}
}

func TestMOpensMonthSheet(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	ctx := context.Background()
	day := time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC)
	w := 171.5
	log, err := dailylog.New(day, &w, 0, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", config.Default())
	loaded := app.load()
	app.Update(loaded)
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	view := model.View()
	if !strings.Contains(view, "November 1990") {
		t.Fatalf("View = %q", view)
	}
	if !strings.Contains(view, "30") {
		t.Fatalf("expected 30 days in November: %q", view)
	}
}

func TestMonthCellSaveWritesWeight(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.month = newMonth(time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	app.month.col = colWeight
	app.month.beginEdit("80.0")
	msg := app.saveMonthCell()
	if _, ok := msg.(savedMsg); !ok {
		t.Fatalf("save = %#v", msg)
	}
	got, err := store.Get(context.Background(), time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != 80 {
		t.Fatalf("stored %+v", got)
	}
}

func TestMonthTabWhileEditingSavesAndAdvances(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.screen = screenMonth
	app.month = newMonth(time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	app.month.col = colWeight
	app.month.beginEdit("80.0")
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("tab while editing should save")
	}
	app.Update(cmd())
	got, err := store.Get(context.Background(), time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != 80 {
		t.Fatalf("stored %+v", got)
	}
	if app.month.editing {
		t.Fatal("still editing")
	}
	if app.month.col != colSleep || app.month.day != 4 {
		t.Fatalf("focus day %d col %d, want sleep on day 4", app.month.day, app.month.col)
	}
}

func TestMonthTabInvalidDoesNotAdvance(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.screen = screenMonth
	app.month = newMonth(time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	app.month.col = colWeight
	app.month.beginEdit("nope")
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("tab should attempt save")
	}
	app.Update(cmd())
	if !app.month.editing || app.month.col != colWeight {
		t.Fatalf("editing=%v col=%d", app.month.editing, app.month.col)
	}
	if _, err := store.Get(context.Background(), time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("should not have stored a day")
	}
}

func TestMonthTabWhenNotEditingMoves(t *testing.T) {
	t.Parallel()
	app := New(nil, "mem.db", config.Default())
	app.screen = screenMonth
	app.month = newMonth(time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC))
	app.month.col = colWeight
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if app.month.col != colSleep {
		t.Fatalf("col = %d, want sleep", app.month.col)
	}
}

func TestQuitKeys(t *testing.T) {
	t.Parallel()
	app := New(nil, "", config.Default())
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should quit")
	}
}

func TestLoadSelectsToday(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 20, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
}

func TestLoadSelectsNearestPast(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 8, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 8, 0, 0, 0, 0, time.UTC))
}

func TestLoadSelectsNearestFuture(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 12, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 20, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 12, 0, 0, 0, 0, time.UTC))
}

func TestLoadTiePrefersPast(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 9, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 11, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 9, 0, 0, 0, 0, time.UTC))
}

func TestLoadEmptyHasNoSelection(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	if len(app.logs) != 0 {
		t.Fatalf("logs = %d, want empty", len(app.logs))
	}
	view := visible(app.View())
	if !strings.Contains(view, "no entries yet") {
		t.Fatalf("View = %q", view)
	}
}

func TestLoadSelectsAcrossMonths(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
}

func TestLoadNearerFutureBeatsLastPast(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 11, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 11, 0, 0, 0, 0, time.UTC))
}

func TestSaveKeepsNonClosestDay(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	first := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	seedDays(t, store,
		first,
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, first)
	press(app, "enter")
	saved := app.saveForm()
	if _, ok := saved.(savedMsg); !ok {
		t.Fatalf("save = %#v", saved)
	}
	app.Update(saved)
	app.Update(app.load())
	assertSelectedDay(t, app, first)
}

func TestSaveFromNewDaySelectsWrittenDay(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, time.Date(1990, 11, 8, 0, 0, 0, 0, time.UTC))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "n")
	saved := app.saveForm()
	if _, ok := saved.(savedMsg); !ok {
		t.Fatalf("save = %#v", saved)
	}
	app.Update(saved)
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
}

func TestLoadOutOfRangeSelectsClosest(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	app.cursor = len(app.logs) + 1
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
}

func TestMonthCellSaveKeepsNonClosestDay(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	first := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	seedDays(t, store,
		first,
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, first)
	app.month = newMonth(time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC))
	app.month.col = colWeight
	app.month.beginEdit("80.0")
	saved := app.saveMonthCell()
	if _, ok := saved.(savedMsg); !ok {
		t.Fatalf("save = %#v", saved)
	}
	app.Update(saved)
	app.Update(app.load())
	assertSelectedDay(t, app, first)
}

func TestLoadSelectsLocalCivilDate(t *testing.T) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	// 9 Nov 20:00 in LA is 10 Nov UTC; closest must follow local date.
	freezeToday(t, time.Date(1990, 11, 9, 20, 0, 0, 0, loc))
	store := openStore(t)
	seedDays(t, store,
		time.Date(1990, 11, 9, 0, 0, 0, 0, time.UTC),
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	assertSelectedDay(t, app, time.Date(1990, 11, 9, 0, 0, 0, 0, time.UTC))
}

func TestFailedLoadClearsSelectDay(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	first := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	seedDays(t, store,
		first,
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "n")
	saved := app.saveForm()
	if _, ok := saved.(savedMsg); !ok {
		t.Fatalf("save = %#v", saved)
	}
	app.Update(saved)
	app.Update(loadErrMsg{err: errors.New("load failed")})
	selectDay(t, app, first)
	app.Update(savedMsg{})
	app.Update(app.load())
	assertSelectedDay(t, app, first)
}

func openStore(t *testing.T) *dailylog.Store {
	t.Helper()
	s, err := dailylog.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func seedDays(t *testing.T, store *dailylog.Store, days ...time.Time) {
	t.Helper()
	ctx := context.Background()
	w := 80.0
	for _, day := range days {
		log, err := dailylog.New(day, &w, 8, 0, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
}

func selectDay(t *testing.T, app *App, day time.Time) {
	t.Helper()
	for i, log := range app.logs {
		if log.Day.Equal(day) {
			app.cursor = i
			return
		}
	}
	t.Fatalf("day %s not in logs", day.Format(time.DateOnly))
}

func assertSelectedDay(t *testing.T, app *App, want time.Time) {
	t.Helper()
	if len(app.logs) == 0 {
		t.Fatal("no logs loaded")
	}
	if app.cursor < 0 || app.cursor >= len(app.logs) {
		t.Fatalf("cursor %d out of range (len %d)", app.cursor, len(app.logs))
	}
	got := app.logs[app.cursor].Day
	if !got.Equal(want) {
		t.Fatalf("selected day = %s, want %s", got.Format(time.DateOnly), want.Format(time.DateOnly))
	}
}
