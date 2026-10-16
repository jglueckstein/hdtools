package tui

// The long-term chart is the book's WEIGHT menu: quarterly,
// semiannual, annual, or complete history, ending at the latest log.
// The picture is the Online canvas in longonline.go. This file keeps
// the keys, the four windows, and the straight-line fit. The fit
// reads the day span, so a shared column is not a regression sample.
// The fit itself lives in dailylog. This file does not open SQLite,
// does not edit the log, and does not follow the terminal width.

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

type longKind int

const (
	longQuarterly longKind = iota
	longSemiannual
	longAnnual
	longComplete
	longKindCount
)

func (k longKind) name() string {
	switch k {
	case longSemiannual:
		return "Semiannual"
	case longAnnual:
		return "Annual"
	case longComplete:
		return "Complete"
	default:
		return "Quarterly"
	}
}

func (a *App) openLong() {
	a.afterLong = a.screen
	a.longKind = longQuarterly
	a.screen = screenLong
	a.status = ""
	a.err = nil
}

func (a *App) updateLong(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.screen = a.afterLong
		return a, nil
	case "q", "ctrl+c":
		return a, tea.Quit
	case "[":
		a.longKind = (a.longKind + longKindCount - 1) % longKindCount
	case "]":
		a.longKind = (a.longKind + 1) % longKindCount
	}
	return a, nil
}

func (a *App) longEnd() (year int, month time.Month, lastDay int, ok bool) {
	if len(a.logs) == 0 {
		return 0, 0, 0, false
	}
	latest := a.logs[0].Day
	for _, log := range a.logs {
		if log.Day.After(latest) {
			latest = log.Day
		}
	}
	year, month, _ = latest.Date()
	lastDay = daysInMonth(year, month)
	ty, tm, td := localToday().Date()
	if year == ty && month == tm && td < lastDay {
		lastDay = td
	}
	return year, month, lastDay, true
}

func longStart(kind longKind, endYear int, endMonth time.Month, logs []dailylog.DailyLog) time.Time {
	endFirst := time.Date(endYear, endMonth, 1, 0, 0, 0, 0, time.UTC)
	switch kind {
	case longSemiannual:
		return endFirst.AddDate(0, -5, 0)
	case longAnnual:
		return endFirst.AddDate(0, -11, 0)
	case longComplete:
		earliest := logs[0].Day
		for _, log := range logs {
			if log.Day.Before(earliest) {
				earliest = log.Day
			}
		}
		return earliest
	default:
		return endFirst.AddDate(0, -2, 0)
	}
}

func sheetRange(logs []dailylog.DailyLog, start, end time.Time) []sheetDay {
	n := int(end.Sub(start).Hours()/24) + 1
	if n < 1 {
		return nil
	}
	out := make([]sheetDay, n)
	byDay := make(map[string]dailylog.DailyLog, len(logs))
	for _, log := range logs {
		byDay[log.Day.Format(time.DateOnly)] = log
	}
	for i := 0; i < n; i++ {
		day := start.AddDate(0, 0, i)
		row := sheetDay{Day: day}
		if log, ok := byDay[day.Format(time.DateOnly)]; ok {
			row.Log = log
			row.HasEntry = true
			row.Trend = log.Trend
			row.HasTrend = log.Weight != nil || log.Trend != 0
		} else if t, ok := lastTrendOnOrBefore(logs, day); ok {
			row.Trend = t
			row.HasTrend = true
		}
		out[i] = row
	}
	return out
}

func (a *App) longView() string {
	p := a.pal
	unit := a.cfg.DisplayUnit
	help := p.help.Render("esc back   [ ] kind   q quit")
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", longHeading(a.longKind.name(), p))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))
	ey, em, lastDay, ok := a.longEnd()
	if !ok {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty span)"), help)
		return b.String()
	}
	start := longStart(a.longKind, ey, em, a.logs)
	end := time.Date(ey, em, lastDay, 0, 0, 0, 0, time.UTC)
	span := sheetRange(a.logs, start, end)
	if end.Before(start) || chartEmpty(span) {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty span)"), help)
		return b.String()
	}
	b.WriteString(renderLongOnline(span, unit, p))
	if line, ok := longAnalysis(span, unit); ok {
		fmt.Fprintf(&b, "\n%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", help)
	return b.String()
}

// longAnalysis fits the day span. x is the index in that slice, so a
// leading day with no trend keeps the later offsets and is left out
// of the sample. A window that ends today still uses the fit. The
// caption is omitted only when the fit itself has no rate.
func longAnalysis(span []sheetDay, unit units.Unit) (string, bool) {
	var x, y []float64
	for i, d := range span {
		if !d.HasTrend {
			continue
		}
		x = append(x, float64(i))
		y = append(y, d.Trend)
	}
	lossKG, kcal, ok := dailylog.FitRate(x, y)
	if !ok {
		return "", false
	}
	loss, err := units.FromKG(lossKG, unit)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("Loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal), true
}
