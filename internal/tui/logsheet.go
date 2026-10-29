package tui

// The month sheet's shift-P writes the filled log the command writes.
// The key loads the store and trends from the first weigh-in, then
// calls the existing writer. The rows on screen stay as they are,
// including after a rejected reload. app.go only recognizes the
// letter and asks this file what a late save may do to the status
// line, because the write would push that file over 500 lines.
//
// A month-cell save and the reload it starts are in flight until
// their result arrives. Shift-P returns during that gap and leaves
// the status line alone. The word "saved" is not that mark: a sticky
// saved after the reload has finished is still a page, and a failed
// write clears it. Once this key has put a path on the status line,
// a save that was already armed does not cover it. A month-cell save
// armed after that path sets the status to saved and clears the path.
// The file is not rewritten until the next shift-P. [, ], and t clear
// the path when they land on a different month. The file stays where
// it was written.
//
// This file does not paint, import a PDF library, call the command,
// or change ApplyTrend. It does not prompt for a path, and -o stays
// a command flag. Chart p is not this key.

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/logpdf"
)

// logPDFGate is the in-flight hold for the month-sheet key.
// save counts month-cell save commands that have not finished.
// A count, not a boolean, because a stale save must not release
// a newer one. The command goroutine decrements a drop or a
// failed save, so the count is atomic. next stamps each armed
// save. keepBelow is the highest id already armed when the path
// was written; a later id is a new edit. reload is set when
// savedMsg starts the load, until that load is applied or fails.
// path is the file this key wrote, and year/month are the sheet
// that file belongs to.
type logPDFGate struct {
	save      atomic.Uint64
	next      atomic.Uint64
	keepBelow uint64
	reload    bool
	path      string
	year      int
	month     time.Month
}

// writeMonthLogPDF writes YYYY-MM-log.pdf for the month on screen.
// The series comes from the store, not from the rows on screen.
func (a *App) writeMonthLogPDF() {
	if a.logPDF.save.Load() > 0 || a.logPDF.reload {
		return
	}
	logs, err := a.store.All(context.Background())
	if err != nil {
		a.failLogPDF(fmt.Errorf("load logs: %w", err))
		return
	}
	if err := dailylog.RejectUnsorted(logs); err != nil {
		a.failLogPDF(err)
		return
	}
	logs, err = dailylog.TrendFromFirstWeighIn(logs)
	if err != nil {
		a.failLogPDF(err)
		return
	}
	dir, err := config.ResolvePDFDir(a.cfg)
	if err != nil {
		a.failLogPDF(err)
		return
	}
	path := filepath.Join(dir, fmt.Sprintf("%04d-%02d-log.pdf", a.month.year, a.month.month))
	err = logpdf.Write(path, logpdf.Options{
		Year:  a.month.year,
		Month: a.month.month,
		Logs:  logs,
		Unit:  a.cfg.DisplayUnit,
		Today: localToday(),
	})
	if err != nil {
		a.failLogPDF(err)
		return
	}
	a.month.err = ""
	a.month.holdErr = false
	a.err = nil
	a.status = path
	a.logPDF.path = path
	a.logPDF.year = a.month.year
	a.logPDF.month = a.month.month
	a.logPDF.keepBelow = a.logPDF.next.Load()
}

// failLogPDF keeps the cause and names this writer. The month view
// paints month.err, not a.err, so a leftover a.err would show up on
// the list after Esc and hide the failure here.
func (a *App) failLogPDF(err error) {
	a.month.err = fmt.Errorf("write log pdf: %w", err).Error()
	a.month.holdErr = true
	a.err = nil
	a.status = ""
	a.logPDF.path = ""
	a.logPDF.keepBelow = 0
}

// armLogSave counts the save when the command is returned. The next
// key is not ordered behind that command. A dropped or failed save
// releases only itself here, so Update can keep ignoring the drop.
func (a *App) armLogSave(cmd tea.Cmd) tea.Cmd {
	id := a.logPDF.next.Add(1)
	a.logPDF.save.Add(1)
	return func() tea.Msg {
		msg := cmd()
		switch m := msg.(type) {
		case monthSaveDroppedMsg, loadErrMsg:
			a.decLogSave()
		case savedMsg:
			m.id = id
			msg = m
		}
		return msg
	}
}

// decLogSave releases one month-cell save. A message that was never
// armed, or that is delivered twice, does not drive the count negative.
func (a *App) decLogSave() {
	for {
		n := a.logPDF.save.Load()
		if n == 0 {
			return
		}
		if a.logPDF.save.CompareAndSwap(n, n-1) {
			return
		}
	}
}

// clearLogSave releases a stale month-cell save that will not reload.
func (a *App) clearLogSave() {
	a.decLogSave()
}

// noteMonthSaved starts the reload. monthCell is false for a form
// save, which never took a count. A form save leaves a path this
// key already showed. A month-cell save already armed when the path
// was written leaves that path up. A month-cell save armed later
// sets the status to saved and clears the path. This does not
// rewrite the PDF.
func (a *App) noteMonthSaved(monthCell bool, saveID uint64) {
	if monthCell {
		a.decLogSave()
	}
	a.logPDF.reload = true
	a.err = nil
	if a.keepsLogPath(monthCell, saveID) {
		return
	}
	a.status = "saved"
	a.logPDF.path = ""
	a.logPDF.keepBelow = 0
}

// keepsLogPath reports whether this save result may leave the PDF
// path on the status line. A form save does. A month-cell save does
// only when it was armed before the path was written.
func (a *App) keepsLogPath(monthCell bool, saveID uint64) bool {
	if a.logPDF.path == "" || a.status != a.logPDF.path {
		return false
	}
	if !monthCell {
		return true
	}
	return saveID != 0 && saveID <= a.logPDF.keepBelow
}

// stepShownMonth moves off the month on screen. The log path is
// cleared when the landing month is not the one the file was
// written for.
func (a *App) stepShownMonth(prev bool) {
	if prev {
		a.month.prevMonth()
	} else {
		a.month.nextMonth()
	}
	a.noteMonthShown()
}

// noteMonthShown drops a log path that names a month the sheet is
// no longer showing. A same-month t keeps it. Arrows do not call
// this, and the file stays on disk.
func (a *App) noteMonthShown() {
	if a.logPDF.path == "" {
		return
	}
	if a.logPDF.year == a.month.year && a.logPDF.month == a.month.month {
		return
	}
	if a.status == a.logPDF.path {
		a.status = ""
	}
	a.logPDF.path = ""
	a.logPDF.keepBelow = 0
}

// clearLogReload opens the key again once the reloaded rows are in.
func (a *App) clearLogReload() {
	a.logPDF.reload = false
}

// clearLogFlight opens the key after a reload fails. A failed
// month-cell save already released its own count. Sticky saved
// stays until a later write clears it.
func (a *App) clearLogFlight() {
	a.logPDF.reload = false
	// A load error is a failed save or reload, not the export
	// failure that has to survive typing.
	a.month.holdErr = false
}
