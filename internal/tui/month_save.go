package tui

// Month-cell saves. Tab still reads the live cell and advances with
// nextCell. Enter, Down, and Up snapshot the cell when the key is
// handled, because Bubble Tea runs the command later and a second key
// can change the cursor first. This file does not draw the sheet, and
// it does not clear the sticky "saved" status (that line is deferred).

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/dailylog"
)

// monthSaveDroppedMsg is a vertical save that Esc cancelled before the
// command wrote. Update ignores it, so the cancelled edit stays put.
type monthSaveDroppedMsg struct{}

// saveMonthCell writes whatever the sheet is focused on now. Tab and
// workout Space use it. Enter, Down, and Up do not: they snapshot first.
func (a *App) saveMonthCell() tea.Msg {
	return a.saveMonthCellAt(a.month.cursorDay(), a.month.col, a.month.input.Value())
}

func (a *App) saveMonthCellAt(day time.Time, col int, text string) tea.Msg {
	log, err := a.store.Get(context.Background(), day)
	if err != nil {
		if !errors.Is(err, dailylog.ErrNotFound) {
			return loadErrMsg{err: err}
		}
		log = dailylog.DailyLog{Day: day}
	}
	if col == colWorkout {
		log.Workout = !log.Workout
		log, err = dailylog.New(log.Day, log.Weight, log.SleepHours, log.Steps, log.Workout, log.Note)
	} else {
		log, err = patchCell(log, col, text, a.cfg.DisplayUnit)
	}
	if err != nil {
		return loadErrMsg{err: err}
	}
	if err := a.store.Upsert(context.Background(), log); err != nil {
		return loadErrMsg{err: err}
	}
	return savedMsg{}
}

func (a *App) saveMonthCellAndAdvance() tea.Msg {
	return a.finishMonthSave(true, 0)
}

// finishMonthSave is Tab's save-then-advance. It reads the live cell.
// A failed save returns unchanged, so the cell stays put. Vertical
// movement must not come through here: that would call nextCell.
func (a *App) finishMonthSave(advance bool, dayDelta int) tea.Msg {
	msg := a.saveMonthCell()
	saved, ok := msg.(savedMsg)
	if !ok {
		return msg
	}
	saved.advance = advance
	saved.dayDelta = dayDelta
	return saved
}

// beginVerticalMonthSave captures the cell on the key. The command
// runs later. Esc bumps monthSaveCancel so this command writes nothing.
// The move is applied once per edit generation.
func (a *App) beginVerticalMonthSave(dayDelta int) tea.Cmd {
	day := a.month.cursorDay()
	col := a.month.col
	text := a.month.input.Value()
	cancelGen := a.monthSaveCancel.Load()
	editGen := a.monthSaveEdit.Load()
	return func() tea.Msg {
		if cancelGen != a.monthSaveCancel.Load() {
			return monthSaveDroppedMsg{}
		}
		msg := a.saveMonthCellAt(day, col, text)
		saved, ok := msg.(savedMsg)
		if !ok {
			return msg
		}
		saved.dayDelta = dayDelta
		saved.origin = day
		saved.cancelGen = cancelGen
		saved.editGen = editGen
		saved.guard = true
		return saved
	}
}

func (a *App) verticalSaveCurrent(msg savedMsg) bool {
	return msg.cancelGen == a.monthSaveCancel.Load() && msg.editGen == a.monthSaveEdit.Load()
}
