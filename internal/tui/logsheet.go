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
// a save result that was already in flight does not cover it.
//
// This file does not paint, import a PDF library, call the command,
// or change ApplyTrend. It does not prompt for a path, and -o stays
// a command flag. Chart p is not this key.

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/logpdf"
)

// logPDFGate is the in-flight hold for the month-sheet key.
// save counts month-cell save commands that have not finished.
// A count, not a boolean, because a stale save must not release
// a newer one. The command goroutine decrements a drop or a
// failed save, so the count is atomic. reload is set when
// savedMsg starts the load, until that load is applied or fails.
// path is the file this key wrote while it is still the status line.
type logPDFGate struct {
	save   atomic.Uint64
	reload bool
	path   string
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
	a.err = nil
	a.status = path
	a.logPDF.path = path
}

// failLogPDF keeps the cause and names this writer. The month view
// paints month.err, not a.err, so a leftover a.err would show up on
// the list after Esc and hide the failure here.
func (a *App) failLogPDF(err error) {
	a.month.err = fmt.Errorf("write log pdf: %w", err).Error()
	a.err = nil
	a.status = ""
	a.logPDF.path = ""
}

// armLogSave counts the save when the command is returned. The next
// key is not ordered behind that command. A dropped or failed save
// releases only itself here, so Update can keep ignoring the drop.
func (a *App) armLogSave(cmd tea.Cmd) tea.Cmd {
	a.logPDF.save.Add(1)
	return func() tea.Msg {
		msg := cmd()
		switch msg.(type) {
		case monthSaveDroppedMsg, loadErrMsg:
			a.decLogSave()
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
// save, which never took a count. A path this key already showed
// stays on the status line; the load still runs.
func (a *App) noteMonthSaved(monthCell bool) {
	if monthCell {
		a.decLogSave()
	}
	a.logPDF.reload = true
	a.err = nil
	if a.logPDF.path != "" && a.status == a.logPDF.path {
		return
	}
	a.status = "saved"
	a.logPDF.path = ""
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
}
