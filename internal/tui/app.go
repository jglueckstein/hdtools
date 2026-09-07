// Package tui is the Bubble Tea interface for hdtools.
//
// The binary stays a thin Open-and-Run loop. This package owns screen state
// and talks to the daily log store; it does not open SQLite itself, so tests
// can inject a memory-backed store and so a later remote database can reuse
// the same Model.
//
// The list is the home screen. A single-day form creates and edits rows.
// Monthly grids, charts, and meal planning are out of scope here.
package tui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

type screen int

const (
	screenList screen = iota
	screenForm
)

// App is the root Bubble Tea model: a log list, a day form, and the
// display unit from config.toml.
type App struct {
	store  *dailylog.Store
	cfg    config.Config
	dbPath string
	logs   []dailylog.DailyLog
	cursor int
	screen screen
	form   formModel
	err    error
	status string
}

type loadedMsg struct {
	logs []dailylog.DailyLog
}

type loadErrMsg struct {
	err error
}

type savedMsg struct{}

// DefaultDBPath is ~/.hdtools/hdtools.db.
func DefaultDBPath() (string, error) {
	return config.DefaultDBPath()
}

// New returns an App that will load logs from store on Init.
func New(store *dailylog.Store, dbPath string, cfg config.Config) *App {
	return &App{
		store:  store,
		cfg:    cfg,
		dbPath: dbPath,
		form:   newForm(cfg.DisplayUnit),
	}
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
	if len(logs) == 0 {
		return loadedMsg{logs: logs}
	}
	trended, err := dailylog.ApplyTrend(logs, nil)
	if err != nil {
		return loadErrMsg{err: err}
	}
	return loadedMsg{logs: trended}
}

func (a *App) saveForm() tea.Msg {
	log, err := a.form.parse()
	if err != nil {
		return loadErrMsg{err: err}
	}
	if err := a.store.Upsert(context.Background(), log); err != nil {
		return loadErrMsg{err: err}
	}
	return savedMsg{}
}

// Update handles list navigation, the day form, and load/save results.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		a.logs = msg.logs
		a.err = nil
		if a.cursor >= len(a.logs) {
			a.cursor = 0
		}
		return a, nil
	case savedMsg:
		a.screen = screenList
		a.status = "saved"
		a.err = nil
		return a, a.load
	case loadErrMsg:
		if a.screen == screenForm {
			a.form.err = msg.err.Error()
			a.err = nil
			return a, nil
		}
		a.err = msg.err
		return a, nil
	case tea.KeyMsg:
		if a.screen == screenForm {
			return a.updateForm(msg)
		}
		return a.updateList(msg)
	}
	return a, nil
}

func (a *App) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "n":
		a.openForm(localToday())
		return a, nil
	case "enter":
		if len(a.logs) == 0 {
			return a, nil
		}
		a.openForm(a.logs[a.cursor].Day)
		return a, nil
	case "up", "k":
		if a.cursor > 0 {
			a.cursor--
		}
	case "down", "j":
		if a.cursor < len(a.logs)-1 {
			a.cursor++
		}
	}
	return a, nil
}

func (a *App) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.screen = screenList
		a.form.err = ""
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "enter":
		a.form.err = ""
		return a, a.saveForm
	}
	cmd := a.form.update(msg)
	return a, cmd
}

func (a *App) openForm(day time.Time) {
	a.form = newForm(a.cfg.DisplayUnit)
	log, err := a.store.Get(context.Background(), day)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.form.loadNew(day)
		} else {
			a.err = err
			return
		}
	} else {
		a.form.load(log)
	}
	a.screen = screenForm
	a.status = ""
	a.err = nil
}

// View renders the list or the day form.
func (a *App) View() string {
	if a.screen == screenForm {
		return a.form.view()
	}
	return a.listView()
}

func (a *App) listView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "hdtools — daily log  (weight %s)\n", a.cfg.DisplayUnit)
	fmt.Fprintf(&b, "db: %s\n\n", a.dbPath)
	if a.err != nil {
		fmt.Fprintf(&b, "error: %v\n\nq quit\n", a.err)
		return b.String()
	}
	if len(a.logs) == 0 {
		fmt.Fprintf(&b, "(no entries yet)\n\nn new day   q quit\n")
		return b.String()
	}
	fmt.Fprintf(&b, "    date        weight   trend   sleep  steps  workout  note\n")
	for i, log := range a.logs {
		mark := "  "
		if i == a.cursor {
			mark = "> "
		}
		weight := "—"
		trend := "—"
		if log.Weight != nil {
			w, err := units.FromKG(*log.Weight, a.cfg.DisplayUnit)
			if err == nil {
				weight = fmt.Sprintf("%6.1f", w)
			}
		}
		if t, err := units.FromKG(log.Trend, a.cfg.DisplayUnit); err == nil && (log.Weight != nil || log.Trend != 0) {
			trend = fmt.Sprintf("%5.1f", t)
		}
		workout := "no"
		if log.Workout {
			workout = "yes"
		}
		fmt.Fprintf(&b, "%s%s  %7s  %5s  %5.1f  %5d  %-7s  %s\n",
			mark,
			log.Day.Format("2006-01-02"),
			weight,
			trend,
			log.SleepHours,
			log.Steps,
			workout,
			log.Note,
		)
	}
	if a.status != "" {
		fmt.Fprintf(&b, "\n%s\n", a.status)
	}
	fmt.Fprintf(&b, "\nn new   enter edit   q quit\n")
	return b.String()
}
