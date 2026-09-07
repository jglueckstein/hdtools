// Package tui is the Bubble Tea interface for hdtools.
//
// The binary stays a thin Open-and-Run loop. This package owns screen state
// and talks to the daily log store; it does not open SQLite itself, so tests
// can inject a memory-backed store and so a later remote database can reuse
// the same Model.
//
// Editing a day, monthly charts, and meal planning are out of scope here.
package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/dailylog"
)

// App is the root Bubble Tea model: a read-only list of daily logs plus the
// path of the database they came from.
type App struct {
	store  *dailylog.Store
	dbPath string
	logs   []dailylog.DailyLog
	err    error
}

type loadedMsg struct {
	logs []dailylog.DailyLog
}

type loadErrMsg struct {
	err error
}

// New returns an App that will load logs from store on Init.
func New(store *dailylog.Store, dbPath string) *App {
	return &App{store: store, dbPath: dbPath}
}

// DefaultDBPath is the local SQLite file when the user has not chosen a path.
// $HOME/.hdtools/hdtools.db keeps the database out of the working directory
// so running the TUI from a source checkout cannot drop a db next to go.mod.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".hdtools", "hdtools.db"), nil
}

// Init loads the log series from the store.
func (a *App) Init() tea.Cmd {
	return a.load
}

func (a *App) load() tea.Msg {
	logs, err := a.store.All(context.Background())
	if err != nil {
		return loadErrMsg{err: err}
	}
	trended, err := dailylog.ApplyTrend(logs, nil)
	if err != nil {
		return loadErrMsg{err: err}
	}
	return loadedMsg{logs: trended}
}

// Update handles quit keys and the initial load result.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		a.logs = msg.logs
		a.err = nil
	case loadErrMsg:
		a.err = msg.err
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

// View renders the database path and the daily log table.
func (a *App) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "hdtools — daily log\n")
	fmt.Fprintf(&b, "db: %s\n\n", a.dbPath)
	if a.err != nil {
		fmt.Fprintf(&b, "error: %v\n\nq quit\n", a.err)
		return b.String()
	}
	if len(a.logs) == 0 {
		fmt.Fprintf(&b, "(no entries yet)\n\nq quit\n")
		return b.String()
	}
	fmt.Fprintf(&b, "  date        weight   trend   sleep  steps  workout\n")
	for _, log := range a.logs {
		weight := "—"
		if log.Weight != nil {
			weight = fmt.Sprintf("%6.1f", *log.Weight)
		}
		workout := "no"
		if log.Workout {
			workout = "yes"
		}
		fmt.Fprintf(&b, "  %s  %7s  %5.1f  %5.1f  %5d  %s\n",
			log.Day.Format("2006-01-02"),
			weight,
			log.Trend,
			log.SleepHours,
			log.Steps,
			workout,
		)
	}
	fmt.Fprintf(&b, "\nq quit\n")
	return b.String()
}
