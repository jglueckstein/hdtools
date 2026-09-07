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
