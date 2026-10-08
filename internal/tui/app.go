// Package tui is the screen: list, day form, monthly sheet, and charts.
//
// It owns Bubble Tea state and talks to dailylog.Store but never opens
// SQLite or imports database/sql, so tests inject a file and a later
// remote database can reuse the same Model. Missing days are
// dailylog.ErrNotFound. Meal planning is out of scope. Monthly chart
// PDF export goes through internal/chartpdf; this package must not
// import a PDF library.
//
// Color comes from a palette built at New from config.toml and NO_COLOR.
// Column geometry lives in layout.go so headers stay over numbers after
// ANSI codes are applied.
package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

type screen int

const (
	screenList screen = iota
	screenForm
	screenMonth
	screenChart
	screenLong
)

// App is the root Bubble Tea model: a log list, a day form, a monthly
// sheet, charts, and the display unit from config.toml.
type App struct {
	store      *dailylog.Store
	cfg        config.Config
	dbPath     string
	logs       []dailylog.DailyLog
	cursor     int
	screen     screen
	form       formModel
	month      monthModel
	afterSave  screen
	afterChart screen
	afterLong  screen
	longKind   longKind
	termCols   int
	err        error
	status     string
	pal        palette
	// selectDay, when set, is the calendar day the next loadedMsg
	// should land on (the day a form just wrote).
	selectDay time.Time
	// monthSaveCancel is bumped when Esc cancels an edit, so a
	// vertical save command that has not run yet writes nothing.
	// monthSaveEdit is bumped when that save's move is applied, so
	// a second savedMsg from the same edit does not skip a day.
	monthSaveCancel atomic.Uint64
	monthSaveEdit   atomic.Uint64
}

type loadedMsg struct {
	logs []dailylog.DailyLog
}

type loadErrMsg struct {
	err error
}

type savedMsg struct {
	advance   bool
	dayDelta  int       // month cell: +1 down, -1 up, 0 stay. Never combined with advance.
	day       time.Time // form save: the written day; zero for month-cell save
	origin    time.Time // vertical save: the day captured when the key was handled
	cancelGen uint64
	editGen   uint64
	guard     bool // vertical save: drop the move if Esc or an earlier twin won
}

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
		pal:       newPalette(cfg),
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
	return savedMsg{day: log.Day}
}

// placeCursor sets the list cursor after a load. A form save lands on
// the written day; otherwise keep the previous calendar day if it is
// still present; otherwise closest to local today.
func (a *App) placeCursor(logs []dailylog.DailyLog) {
	keep := time.Time{}
	if a.cursor >= 0 && a.cursor < len(a.logs) {
		keep = a.logs[a.cursor].Day
	}
	want := a.selectDay
	a.selectDay = time.Time{}
	a.logs = logs
	if len(a.logs) == 0 {
		a.cursor = 0
		return
	}
	if !want.IsZero() {
		if i := indexOfDay(a.logs, want); i >= 0 {
			a.cursor = i
			return
		}
	} else if !keep.IsZero() {
		if i := indexOfDay(a.logs, keep); i >= 0 {
			a.cursor = i
			return
		}
	}
	a.cursor = closestLogIndex(a.logs, localToday())
}

// Update handles list navigation, the day form, and load/save results.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		a.placeCursor(msg.logs)
		a.err = nil
		return a, nil
	case savedMsg:
		// A stale vertical save must not cancel a newer edit or
		// move the cursor a second time.
		if msg.guard && !a.verticalSaveCurrent(msg) {
			return a, nil
		}
		a.month.cancelEdit()
		if msg.advance {
			a.month.nextCell()
		} else if msg.dayDelta != 0 {
			if !msg.origin.IsZero() {
				a.month.year, a.month.month, a.month.day = msg.origin.Date()
			}
			a.month.moveDay(msg.dayDelta)
			if msg.guard {
				a.monthSaveEdit.Add(1)
			}
		}
		if !msg.day.IsZero() {
			a.selectDay = msg.day
		}
		a.screen = a.afterSave
		a.status = "saved"
		a.err = nil
		return a, a.load
	case loadErrMsg:
		// A failed reload must not apply a form-save day to a later
		// month-cell load (keep-or-closest, not a sticky want).
		a.selectDay = time.Time{}
		if a.screen == screenForm {
			a.form.err = msg.err.Error()
			a.err = nil
			return a, nil
		}
		if a.screen == screenMonth {
			a.month.err = msg.err.Error()
			a.err = nil
			return a, nil
		}
		a.err = msg.err
		return a, nil
	case tea.WindowSizeMsg:
		a.termCols = msg.Width
		return a, nil
	case tea.PasteMsg:
		return a.paste(msg)
	case tea.KeyPressMsg:
		switch a.screen {
		case screenForm:
			return a.updateForm(msg)
		case screenMonth:
			return a.updateMonth(msg)
		case screenChart:
			return a.updateChart(msg)
		case screenLong:
			return a.updateLong(msg)
		default:
			return a.updateList(msg)
		}
	}
	return a, nil
}

func (a *App) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "n":
		a.openForm(localToday(), screenList)
		return a, nil
	case "m":
		a.openMonth()
		return a, nil
	case "t":
		if len(a.logs) == 0 {
			return a, nil
		}
		a.cursor = closestLogIndex(a.logs, localToday())
		return a, nil
	case "c":
		a.openChart()
		return a, nil
	case "l":
		a.openLong()
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

func (a *App) updateForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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
		if errors.Is(err, dailylog.ErrNotFound) {
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

// spaceBar is the space bar. Its printed name is "space"; the text is
// still one space. Matching either the word or the old name " " would
// miss the key or insert the word into a field.
func spaceBar(msg tea.KeyPressMsg) bool {
	return msg.Code == tea.KeySpace || msg.Text == " "
}

func (a *App) updateMonth(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if a.month.editing {
		switch msg.String() {
		case "esc":
			a.month.cancelEdit()
			a.monthSaveCancel.Add(1)
			return a, nil
		case "enter", "down":
			return a, a.beginVerticalMonthSave(1)
		case "up":
			return a, a.beginVerticalMonthSave(-1)
		case "tab":
			return a, a.saveMonthCellAndAdvance
		case "ctrl+c":
			return a, tea.Quit
		}
		var cmd tea.Cmd
		a.month.input, cmd = a.month.input.Update(msg)
		return a, cmd
	}
	// A space bar's text is one character. Check it before a typed
	// character starts an edit, or an idle weight cell would open.
	if spaceBar(msg) {
		if a.month.col == colWorkout {
			return a, a.saveMonthCell
		}
		return a, nil
	}
	switch msg.String() {
	case "esc":
		a.screen = screenList
		return a, nil
	case "q", "ctrl+c":
		return a, tea.Quit
	case "tab":
		a.month.nextCell()
		return a, nil
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
	case "c":
		a.openChart()
		return a, nil
	case "l":
		a.openLong()
		return a, nil
	case "t":
		a.month = newMonth(localToday())
		return a, nil
	case "enter":
		a.openForm(a.month.cursorDay(), screenMonth)
		return a, nil
	case "n":
		a.openForm(a.month.cursorDay(), screenMonth)
		return a, nil
	default:
		// len is bytes. é is one character and two bytes, and the
		// previous gate counted runes.
		if a.month.col != colWorkout && len([]rune(msg.Text)) == 1 {
			a.month.beginEdit(msg.Text)
		}
	}
	return a, nil
}

// View renders the list, the day form, the month sheet, or a chart.
// Content is that text. Colour in it may stay full fidelity; the
// program writer downsamples once.
func (a *App) View() tea.View {
	switch a.screen {
	case screenForm:
		return tea.NewView(a.form.view(a.pal))
	case screenMonth:
		sheet := buildMonthSheet(a.logs, a.month.year, a.month.month)
		return tea.NewView(a.month.view(sheet, a.cfg.DisplayUnit, a.dbPath, a.status, a.pal))
	case screenChart:
		return tea.NewView(a.chartView())
	case screenLong:
		return tea.NewView(a.longView())
	default:
		return tea.NewView(a.listView())
	}
}

func (a *App) listView() string {
	p := a.pal
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", p.title.Render(fmt.Sprintf("hdtools — daily log  (weight %s)", a.cfg.DisplayUnit)))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render("db: "+a.dbPath))
	if a.err != nil {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.error.Render("error: "+a.err.Error()), p.help.Render("q quit"))
		return b.String()
	}
	if len(a.logs) == 0 {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(no entries yet)"), p.help.Render("n new day   m month   l long   t today   q quit"))
		return b.String()
	}
	fmt.Fprintf(&b, "%s\n", p.header.Render(listHeader()))
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
		selected := i == a.cursor
		showTrend := log.Weight != nil || log.Trend != 0
		delta := formatDelta(log.Weight, log.Trend, showTrend, a.cfg.DisplayUnit)
		var weightCell, trendCell, deltaCell string
		if selected && p.selectionFG {
			weightCell = visPad(weight, wWeight, true)
			trendCell = visPad(trend, wTrend, true)
			deltaCell = visPad(delta, wDelta, true)
		} else {
			weightCell = visPad(p.weight.Render(weight), wWeight, true)
			trendCell = visPad(p.trend.Render(trend), wTrend, true)
			deltaCell = styleDelta(delta, p)
		}
		line := visPad(mark, wMark, false) + joinCols(
			visPad(log.Day.Format("2006-01-02"), wDate, false),
			weightCell,
			trendCell,
			deltaCell,
			visPad(fmt.Sprintf("%.1f", log.SleepHours), wSleep, true),
			visPad(fmt.Sprintf("%d", log.Steps), wSteps, true),
			visPad(workout, wWorkout, false),
			log.Note,
		)
		if selected {
			line = p.selected.Render(line)
		}
		fmt.Fprintf(&b, "%s\n", line)
	}
	if a.status != "" {
		fmt.Fprintf(&b, "\n%s\n", p.status.Render(a.status))
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render("n new   enter edit   m month   c chart   l long   t today   q quit"))
	return b.String()
}
