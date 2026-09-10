package tui

import (
	"context"
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

func openStore(t *testing.T) *dailylog.Store {
	t.Helper()
	s, err := dailylog.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
