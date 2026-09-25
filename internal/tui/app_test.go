package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
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

func TestGotoTodayListJumpsToClosest(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	june := time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC)
	seedDays(t, store, june, today)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, june)
	press(app, "t")
	assertSelectedDay(t, app, today)
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("rows = %d, want 2", len(logs))
	}
}

func TestGotoTodayListDoesNotCreateToday(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	first := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	eighth := time.Date(1990, 11, 8, 0, 0, 0, 0, time.UTC)
	seedDays(t, store, first, eighth)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, first)
	press(app, "t")
	assertSelectedDay(t, app, eighth)
	if _, err := store.Get(context.Background(), time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("created 10 November 1990")
	} else if !errors.Is(err, dailylog.ErrNotFound) {
		t.Fatal(err)
	}
}

func TestGotoTodayListNearestFuture(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	twelfth := time.Date(1990, 11, 12, 0, 0, 0, 0, time.UTC)
	twentieth := time.Date(1990, 11, 20, 0, 0, 0, 0, time.UTC)
	seedDays(t, store, twelfth, twentieth)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, twentieth)
	press(app, "t")
	assertSelectedDay(t, app, twelfth)
}

func TestGotoTodayListTiePrefersPast(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	ninth := time.Date(1990, 11, 9, 0, 0, 0, 0, time.UTC)
	eleventh := time.Date(1990, 11, 11, 0, 0, 0, 0, time.UTC)
	seedDays(t, store, ninth, eleventh)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, eleventh)
	press(app, "t")
	assertSelectedDay(t, app, ninth)
}

func TestGotoTodayListNearerFutureBeatsLastPast(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	first := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	eleventh := time.Date(1990, 11, 11, 0, 0, 0, 0, time.UTC)
	seedDays(t, store, first, eleventh)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, first)
	press(app, "t")
	assertSelectedDay(t, app, eleventh)
}

func TestGotoTodayListEmptyDoesNothing(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "t")
	if app.screen != screenList {
		t.Fatalf("screen = %v, want list", app.screen)
	}
	if len(app.logs) != 0 {
		t.Fatalf("logs = %d, want empty", len(app.logs))
	}
	if !strings.Contains(visible(app.View()), "no entries yet") {
		t.Fatalf("View = %q", app.View())
	}
	if !helpHasKey(visible(app.View()), "t") {
		t.Fatal("empty-list help does not mention t")
	}
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 0 {
		t.Fatalf("rows = %d, want 0", len(logs))
	}
}

func TestGotoTodayListDoesNotChangeMonthModel(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m")
	press(app, "[")
	if app.month.month != time.October {
		t.Fatalf("month = %s, want October", app.month.month)
	}
	press(app, "esc")
	press(app, "t")
	if app.month.year != 1990 || app.month.month != time.October {
		t.Fatalf("after list t: %s %d, want October 1990", app.month.month, app.month.year)
	}
	assertSelectedDay(t, app, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
}

func TestGotoTodayFormIsText(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	june := time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC)
	seedDays(t, store,
		june,
		time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, june)
	press(app, "enter")
	for i := 0; i < fieldNote; i++ {
		press(app, "tab")
	}
	press(app, "t")
	if app.screen != screenForm {
		t.Fatalf("screen = %v, want form", app.screen)
	}
	if !strings.Contains(app.form.inputs[4].Value(), "t") {
		t.Fatalf("note = %q, want t", app.form.inputs[4].Value())
	}
	press(app, "esc")
	assertSelectedDay(t, app, june)
}

func TestGotoTodayHelpMentionsT(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, time.Date(1990, 11, 10, 0, 0, 0, 0, time.UTC))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	if !helpHasKey(visible(app.View()), "t") {
		t.Fatal("list help does not mention t")
	}
	press(app, "m")
	if !helpHasKey(visible(app.View()), "t") {
		t.Fatal("month help does not mention t")
	}
}

func TestListDeltaPositive(t *testing.T) {
	t.Parallel()
	app := listWithPair(t, 80.5, 80.0, config.Default())
	header, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "+0.5" {
		t.Fatalf("delta = %q, want +0.5\nheader %q\ndata   %q", cell, header, data)
	}
}

func TestListDeltaNegative(t *testing.T) {
	t.Parallel()
	app := listWithPair(t, 79.5, 80.0, config.Default())
	_, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "-0.5" {
		t.Fatalf("delta = %q, want -0.5\ndata %q", cell, data)
	}
}

func TestListDeltaZero(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	seedWeighIn(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80.0)
	seedWeighIn(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), 80.0)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	_, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "0.0" {
		t.Fatalf("delta = %q, want 0.0\ndata %q", cell, data)
	}
	if strings.Contains(cell, "+") || strings.Contains(cell, "-") {
		t.Fatalf("zero delta must not be signed: %q", cell)
	}
}

func TestListDeltaRoundedZero(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	// 80.04 vs ApplyTrend's first-day round1(80.04)=80.0; difference rounds to 0.0.
	seedWeighIn(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), 80.04)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	_, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "0.0" {
		t.Fatalf("delta = %q, want 0.0\ndata %q", cell, data)
	}
	if strings.Contains(cell, "+") || strings.Contains(cell, "-") {
		t.Fatalf("rounded-zero delta must not be signed: %q", cell)
	}
}

func TestListDeltaHalfwayMatchesPaintedCells(t *testing.T) {
	t.Parallel()
	// 80.05 kg: %.1f is 80.0, math.Round is 80.1. Delta must match paint.
	app := listWithPair(t, 80.05, 80.00, config.Default())
	view := visible(app.View())
	if !strings.Contains(view, "80.0") {
		t.Fatalf("missing painted 80.0: %q", view)
	}
	_, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "0.0" {
		t.Fatalf("delta = %q, want 0.0 (painted 80.0−80.0), not +0.1 from math.Round\ndata %q", cell, data)
	}
}

func TestListDeltaFirstWeighInIsZero(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	seedWeighIn(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80.0)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	if len(app.logs) != 1 || app.logs[0].Weight == nil || app.logs[0].Trend != *app.logs[0].Weight {
		t.Fatalf("first weigh-in trend = %v weight = %v", app.logs[0].Trend, app.logs[0].Weight)
	}
	_, data, cell := listDeltaCell(t, app, "1990-11-01")
	if cell != "0.0" {
		t.Fatalf("delta = %q, want 0.0\ndata %q", cell, data)
	}
}

func TestListDeltaDisplayedCellsAddUp(t *testing.T) {
	t.Parallel()
	app := listWithPair(t, 80.0, 79.9, config.Config{DisplayUnit: units.Pound})
	view := visible(app.View())
	if !strings.Contains(view, "176.4") || !strings.Contains(view, "176.1") {
		t.Fatalf("list missing displayed lb cells: %q", view)
	}
	_, data, cell := listDeltaCell(t, app, "1990-11-02")
	if cell != "+0.3" {
		t.Fatalf("delta = %q, want +0.3 (displayed 176.4−176.1), not kg residual\ndata %q", cell, data)
	}
	if cell == "+0.2" || strings.Contains(data, "+0.2") {
		t.Fatalf("delta used kilogram residual conversion: %q", data)
	}
}

func TestListDeltaFirstWeighInLBIsDisplayedSubtraction(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	kg, err := units.ToKG(176.5, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	seedWeighIn(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), kg)
	app := New(store, "mem.db", config.Config{DisplayUnit: units.Pound})
	app.Update(app.load())
	view := visible(app.View())
	if !strings.Contains(view, "176.5") {
		t.Fatalf("missing painted 176.5: %q", view)
	}
	_, data, cell := listDeltaCell(t, app, "1990-11-01")
	if cell != "-0.1" {
		t.Fatalf("delta = %q, want -0.1 (176.5−176.6)\ndata %q", cell, data)
	}
}

func TestListDeltaNOCOLORKeepsSign(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := listWithPair(t, 80.5, 80.0, config.Default())
	view := app.View()
	vis := visible(view)
	if !strings.Contains(vis, "+0.5") {
		t.Fatalf("NO_COLOR dropped signed delta: %q", vis)
	}
	for _, line := range strings.Split(view, "\n") {
		if !strings.Contains(visible(line), "+0.5") {
			continue
		}
		if strings.Contains(line, "38;2;") || strings.Contains(line, "38;5;") {
			t.Fatalf("delta cell has chromatic ANSI under NO_COLOR: %q", line)
		}
		if hasChromaticSGR(line) {
			t.Fatalf("delta cell has chromatic SGR under NO_COLOR: %q", line)
		}
	}
}

func TestListDeltaCustomColorPainted(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	cfg := loadConfigTOML(t, `display_unit = "kg"

[colors]
delta-pos = "magenta"
`)
	app := listWithPair(t, 80.5, 80.0, cfg)
	view := app.View()
	vis := visible(view)
	if !strings.Contains(vis, "+0.5") {
		t.Fatalf("visible missing +0.5: %q", vis)
	}
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "+0.5") {
			line = l
			break
		}
	}
	if !hasIndexedForeground(line, 5) {
		t.Fatalf("+0.5 not magenta: %q", line)
	}
	if hasIndexedForeground(line, 3) {
		t.Fatalf("+0.5 still default yellow: %q", line)
	}
}

func TestListDeltaCustomNegColorPainted(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	cfg := loadConfigTOML(t, `display_unit = "kg"

[colors]
delta-neg = "magenta"
`)
	app := listWithPair(t, 79.5, 80.0, cfg)
	view := app.View()
	if !strings.Contains(visible(view), "-0.5") {
		t.Fatalf("visible missing -0.5: %q", visible(view))
	}
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "-0.5") {
			line = l
			break
		}
	}
	if !hasIndexedForeground(line, 5) {
		t.Fatalf("-0.5 not magenta: %q", line)
	}
}

func TestListDeltaCustomZeroColorPainted(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	cfg := loadConfigTOML(t, `display_unit = "kg"

[colors]
delta-zero = "magenta"
`)
	app := listWithPair(t, 80.0, 80.0, cfg)
	view := app.View()
	if !strings.Contains(visible(view), "0.0") {
		t.Fatalf("visible missing 0.0: %q", visible(view))
	}
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "0.0") && strings.Contains(visible(l), "1990-11-02") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("no data row: %q", visible(view))
	}
	if !hasIndexedForeground(line, 5) {
		t.Fatalf("0.0 not magenta: %q", line)
	}
}

func TestInvalidDeltaNegColorDropped(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	cfg := loadConfigTOML(t, `display_unit = "kg"

[colors]
delta-neg = "chartreuse"
`)
	app := listWithPair(t, 79.5, 80.0, cfg)
	if app.err != nil {
		t.Fatalf("TUI did not open: %v", app.err)
	}
	view := app.View()
	vis := visible(view)
	if !strings.Contains(vis, "-0.5") {
		t.Fatalf("visible missing -0.5: %q", vis)
	}
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "-0.5") {
			line = l
			break
		}
	}
	if !hasIndexedForeground(line, 2) {
		t.Fatalf("-0.5 not default green: %q", line)
	}
}

func TestListDeltaSelectionIncludesReverse(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	app := listWithPair(t, 80.5, 80.0, config.Default())
	view := app.View()
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "+0.5") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("no +0.5 line: %q", visible(view))
	}
	if !hasSGRCode(line, 7) {
		t.Fatalf("selected delta row missing reverse: %q", line)
	}
}

func TestInvalidDeltaColorDropped(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	cfg := loadConfigTOML(t, `display_unit = "kg"

[colors]
delta-pos = "chartreuse"
`)
	app := listWithPair(t, 80.5, 80.0, cfg)
	if app.err != nil {
		t.Fatalf("TUI did not open: %v", app.err)
	}
	view := app.View()
	vis := visible(view)
	if strings.Contains(vis, "error:") {
		t.Fatalf("TUI error: %q", vis)
	}
	if !strings.Contains(vis, "1990-11-02") {
		t.Fatalf("log did not open: %q", vis)
	}
	if !strings.Contains(vis, "+0.5") {
		t.Fatalf("visible missing +0.5: %q", vis)
	}
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), "+0.5") {
			line = l
			break
		}
	}
	if !hasIndexedForeground(line, 3) {
		t.Fatalf("+0.5 not default yellow: %q", line)
	}
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

func helpHasKey(view, key string) bool {
	lines := strings.Split(view, "\n")
	help := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			help = s
			break
		}
	}
	for _, item := range strings.Split(help, "   ") {
		item = strings.TrimSpace(item)
		if item == key || strings.HasPrefix(item, key+" ") {
			return true
		}
	}
	return false
}

func seedWeighIn(t *testing.T, store *dailylog.Store, day time.Time, kg float64) {
	t.Helper()
	w := kg
	log, err := dailylog.New(day, &w, 8, 1000, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
}

// listWithPair loads one weigh-in so ApplyTrend runs, then stamps Weight
// and Trend. Smoothing cannot produce every spec pair (80.5 vs 80.0);
// the delta cell is FromKG(weight)−FromKG(trend) of the displayed row.
func listWithPair(t *testing.T, weight, trend float64, cfg config.Config) *App {
	t.Helper()
	store := openStore(t)
	seedWeighIn(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), weight)
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	if len(app.logs) != 1 {
		t.Fatalf("logs = %d, want 1", len(app.logs))
	}
	w := weight
	app.logs[0].Weight = &w
	app.logs[0].Trend = trend
	return app
}

func loadConfigTOML(t *testing.T, body string) config.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func listDeltaCell(t *testing.T, app *App, day string) (header, data, cell string) {
	t.Helper()
	view := visible(app.View())
	header = headerLine(view)
	data = dataRow(view, day)
	if header == "" || data == "" {
		t.Fatalf("missing header or data in %q", view)
	}
	assertDeltaRightOfTrend(t, header)
	cell = cellUnder(t, header, data, "delta")
	return header, data, cell
}
