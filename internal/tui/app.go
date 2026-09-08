// Package tui is the Bubble Tea interface for hdtools.
//
// The binary stays a thin Open-and-Run loop. This package owns screen state
// and talks to the daily log store; it does not open SQLite itself, so tests
// can inject a memory-backed store and so a later remote database can reuse
// the same Model.
//
// The list is the home screen. A single-day form creates and edits rows.
// The month sheet is the paper log: every day of the month, cell editing,
// trend carry-forward. Charts and meal planning are out of scope here.
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
	screenMonth
)

// App is the root Bubble Tea model: a log list, a day form, and the
// display unit from config.toml.
type App struct {
	store     *dailylog.Store
	cfg       config.Config
	dbPath    string
	logs      []dailylog.DailyLog
	cursor    int
	screen    screen
	form      formModel
	month     monthModel
	afterSave screen
	err       error
	status    string
}

type loadedMsg struct {
	logs []dailylog.DailyLog
}

type loadErrMsg struct {
	err error
}

type savedMsg struct{}

// DefaultDBPath is $XDG_DATA_HOME/hdtools/hdtools.db.
func DefaultDBPath() (string, error) {
	return config.DefaultDBPath()
}

// New returns an App that will load logs from store on Init.
func New(store *dailylog.Store, dbPath string, cfg config.Config) *App {
	return &App{
		store:     store,
		cfg:       cfg,
		dbPath:    dbPath,
		form:      newForm(cfg.DisplayUnit),
		month:     newMonth(localToday()),
		afterSave: screenList,
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
		a.month.cancelEdit()
		a.screen = a.afterSave
		a.status = "saved"
		a.err = nil
		return a, a.load
	case loadErrMsg:
		if a.screen == screenForm {
			a.form.err = msg.err.Error()
			a.err = nil
			return a, nil
		}
		if a.screen == screenMonth {
			a.month.err = msg.err.Error()
			a.month.editing = false
			a.err = nil
			return a, nil
		}
		a.err = msg.err
		return a, nil
	case tea.KeyMsg:
		switch a.screen {
		case screenForm:
			return a.updateForm(msg)
		case screenMonth:
			return a.updateMonth(msg)
		default:
			return a.updateList(msg)
		}
	}
	return a, nil
}

func (a *App) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "n":
		a.openForm(localToday(), screenList)
		return a, nil
	case "m":
		a.openMonth()
		return a, nil
	case "enter":
		if len(a.logs) == 0 {
			return a, nil
		}
		a.openForm(a.logs[a.cursor].Day, screenList)
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
		a.screen = a.afterSave
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

func (a *App) openForm(day time.Time, returnTo screen) {
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
	a.afterSave = returnTo
	a.screen = screenForm
	a.status = ""
	a.err = nil
}

func (a *App) openMonth() {
	if len(a.logs) > 0 && a.cursor >= 0 && a.cursor < len(a.logs) {
		a.month = newMonth(a.logs[a.cursor].Day)
	} else {
		a.month = newMonth(localToday())
	}
	a.afterSave = screenMonth
	a.screen = screenMonth
	a.status = ""
	a.err = nil
}

func (a *App) updateMonth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.month.editing {
		switch msg.String() {
		case "esc":
			a.month.cancelEdit()
			return a, nil
		case "enter":
			return a, a.saveMonthCell
		case "ctrl+c":
			return a, tea.Quit
		}
		var cmd tea.Cmd
		a.month.input, cmd = a.month.input.Update(msg)
		return a, cmd
	}
	switch msg.String() {
	case "esc":
		a.screen = screenList
		return a, nil
	case "q", "ctrl+c":
		return a, tea.Quit
	case "left":
		a.month.col--
		a.month.clamp()
	case "right":
		a.month.col++
		a.month.clamp()
	case "up":
		a.month.day--
		a.month.clamp()
	case "down":
		a.month.day++
		a.month.clamp()
	case "[":
		a.month.prevMonth()
	case "]":
		a.month.nextMonth()
	case "enter":
		a.openForm(a.month.cursorDay(), screenMonth)
		return a, nil
	case "n":
		a.openForm(a.month.cursorDay(), screenMonth)
		return a, nil
	case " ":
		if a.month.col == colWorkout {
			return a, a.saveMonthCell
		}
	default:
		if a.month.col != colWorkout && len(msg.Runes) == 1 && msg.Type == tea.KeyRunes {
			a.month.beginEdit(string(msg.Runes))
		}
	}
	return a, nil
}

func (a *App) saveMonthCell() tea.Msg {
	day := a.month.cursorDay()
	log, err := a.store.Get(context.Background(), day)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return loadErrMsg{err: err}
		}
		log = dailylog.DailyLog{Day: day}
	}
	if a.month.col == colWorkout {
		log.Workout = !log.Workout
		log, err = dailylog.New(log.Day, log.Weight, log.SleepHours, log.Steps, log.Workout, log.Note)
	} else {
		log, err = patchCell(log, a.month.col, a.month.input.Value(), a.cfg.DisplayUnit)
	}
	if err != nil {
		return loadErrMsg{err: err}
	}
	if err := a.store.Upsert(context.Background(), log); err != nil {
		return loadErrMsg{err: err}
	}
	return savedMsg{}
}

// View renders the list or the day form.
func (a *App) View() string {
	switch a.screen {
	case screenForm:
		return a.form.view()
	case screenMonth:
		sheet := buildMonthSheet(a.logs, a.month.year, a.month.month)
		return a.month.view(sheet, a.cfg.DisplayUnit, a.dbPath, a.status)
	default:
		return a.listView()
	}
}

func (a *App) listView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", titleStyle.Render(fmt.Sprintf("hdtools — daily log  (weight %s)", a.cfg.DisplayUnit)))
	fmt.Fprintf(&b, "%s\n\n", mutedStyle.Render("db: "+a.dbPath))
	if a.err != nil {
		fmt.Fprintf(&b, "%s\n\n%s\n", errorStyle.Render("error: "+a.err.Error()), helpStyle.Render("q quit"))
		return b.String()
	}
	if len(a.logs) == 0 {
		fmt.Fprintf(&b, "%s\n\n%s\n", mutedStyle.Render("(no entries yet)"), helpStyle.Render("n new day   m month   q quit"))
		return b.String()
	}
	fmt.Fprintf(&b, "%s\n", headerStyle.Render(listHeader()))
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
				weight = fmt.Sprintf("%.1f", w)
			}
		}
		if t, err := units.FromKG(log.Trend, a.cfg.DisplayUnit); err == nil && (log.Weight != nil || log.Trend != 0) {
			trend = fmt.Sprintf("%.1f", t)
		}
		workout := "no"
		if log.Workout {
			workout = "yes"
		}
		line := visPad(mark, wMark, false) + joinCols(
			visPad(log.Day.Format("2006-01-02"), wDate, false),
			visPad(weightStyle.Render(weight), wWeight, true),
			visPad(trendStyle.Render(trend), wTrend, true),
			visPad(fmt.Sprintf("%.1f", log.SleepHours), wSleep, true),
			visPad(fmt.Sprintf("%d", log.Steps), wSteps, true),
			visPad(workout, wWorkout, false),
			log.Note,
		)
		if i == a.cursor {
			line = selectedStyle.Render(line)
		}
		fmt.Fprintf(&b, "%s\n", line)
	}
	if a.status != "" {
		fmt.Fprintf(&b, "\n%s\n", statusStyle.Render(a.status))
	}
	fmt.Fprintf(&b, "\n%s\n", helpStyle.Render("n new   enter edit   m month   q quit"))
	return b.String()
}
