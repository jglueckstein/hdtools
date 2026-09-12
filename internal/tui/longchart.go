package tui

// The long-term chart is the book's WEIGHT-menu picture: quarterly,
// semiannual, annual, or complete history, ending at the latest log.
// Daily weight is a thin line only when one column per day fits;
// otherwise the plot is trend-only. This file does not open SQLite
// or edit the log.

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const defaultLongCols = 72

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

func (a *App) updateLong(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (a *App) longPlotWidth() int {
	if a.termCols > 6 {
		w := a.termCols - 6
		if w < 8 {
			return 8
		}
		return w
	}
	return defaultLongCols
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

func bucketSpan(span []sheetDay, w int) []sheetDay {
	d := len(span)
	if d <= w {
		return span
	}
	out := make([]sheetDay, w)
	for i := 0; i < w; i++ {
		lo := i * d / w
		hi := (i + 1) * d / w
		if hi <= lo {
			hi = lo + 1
		}
		if hi > d {
			hi = d
		}
		row := sheetDay{Day: span[lo].Day}
		for _, s := range span[lo:hi] {
			if s.HasTrend {
				row.Trend = s.Trend
				row.HasTrend = true
				row.Day = s.Day
			}
		}
		out[i] = row
	}
	return out
}

func (a *App) longView() string {
	p := a.pal
	unit := a.cfg.DisplayUnit
	help := p.help.Render("esc back   [ ] kind   q quit")
	ey, em, lastDay, ok := a.longEnd()
	if !ok {
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n", titleBox(a.longKind.name(), 40))
		fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty span)"), help)
		return b.String()
	}
	start := longStart(a.longKind, ey, em, a.logs)
	end := time.Date(ey, em, lastDay, 0, 0, 0, 0, time.UTC)
	if end.Before(start) {
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n", titleBox(a.longKind.name(), 40))
		fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty span)"), help)
		return b.String()
	}
	span := sheetRange(a.logs, start, end)
	title := fmt.Sprintf("%s  %s–%s", a.longKind.name(), monthYearLabel(start.Month(), start.Year()), monthYearLabel(em, ey))
	w := a.longPlotWidth()
	plotWidth := 6 + w
	if len(span) <= w {
		plotWidth = 6 + len(span)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", titleBox(title, plotWidth))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))
	if chartEmpty(span) {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty span)"), help)
		return b.String()
	}
	daily := len(span) <= w
	cols := bucketSpan(span, w)
	b.WriteString(renderLongPlot(cols, unit, p, daily))
	if line, ok := longAnalysis(span, unit); ok {
		fmt.Fprintf(&b, "\n%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", help)
	return b.String()
}

func longAnalysis(span []sheetDay, unit units.Unit) (string, bool) {
	if len(span) == 0 || !span[0].HasTrend || !span[len(span)-1].HasTrend {
		return "", false
	}
	lossKG, kcal, ok := dailylog.MonthlyBalance(span[0].Trend, span[len(span)-1].Trend, len(span))
	if !ok {
		return "", false
	}
	loss, err := units.FromKG(lossKG, unit)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("Loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal), true
}

func renderLongPlot(span []sheetDay, unit units.Unit, p palette, daily bool) string {
	n := len(span)
	ymin, ymax, ok := longYRange(span, unit, daily)
	if !ok {
		return ""
	}
	grid := make([][]rune, plotRows)
	styleDaily := make([][]bool, plotRows)
	for r := range grid {
		grid[r] = make([]rune, n)
		styleDaily[r] = make([]bool, n)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	yToRow := func(y float64) int {
		if ymax == ymin {
			return plotRows / 2
		}
		t := (ymax - y) / (ymax - ymin)
		r := int(math.Round(t * float64(plotRows-1)))
		if r < 0 {
			r = 0
		}
		if r >= plotRows {
			r = plotRows - 1
		}
		return r
	}
	paintPath := func(useWeight bool) {
		var prevRow int
		var hasPrev bool
		for i, d := range span {
			var kg float64
			if useWeight {
				if d.Log.Weight == nil {
					hasPrev = false
					continue
				}
				kg = *d.Log.Weight
			} else {
				if !d.HasTrend {
					hasPrev = false
					continue
				}
				kg = d.Trend
			}
			y, err := units.FromKG(kg, unit)
			if err != nil {
				hasPrev = false
				continue
			}
			r := yToRow(y)
			g := '-'
			if hasPrev {
				if r > prevRow {
					g = '\\'
				} else if r < prevRow {
					g = '/'
				}
			}
			grid[r][i] = g
			styleDaily[r][i] = useWeight
			prevRow = r
			hasPrev = true
		}
	}
	if daily {
		paintPath(true)
	}
	paintPath(false)
	var b strings.Builder
	for r := 0; r < plotRows; r++ {
		frac := 0.0
		if plotRows > 1 {
			frac = float64(r) / float64(plotRows-1)
		}
		labelY := ymax - frac*(ymax-ymin)
		fmt.Fprintf(&b, "%5.1f-", labelY)
		for c := 0; c < n; c++ {
			ch := grid[r][c]
			s := string(ch)
			switch ch {
			case '-', '/', '\\':
				if styleDaily[r][c] {
					s = p.weight.Render(s)
				} else {
					s = p.trend.Render(s)
				}
			}
			b.WriteString(s)
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s%d", strings.Repeat(" ", 6), 1)
	if n > 1 {
		last := fmt.Sprintf("%d", n)
		pad := n - 1 - len(last)
		if pad < 1 {
			pad = 1
		}
		fmt.Fprintf(&b, "%s%s", strings.Repeat(" ", pad), last)
	}
	b.WriteByte('\n')
	return b.String()
}

func longYRange(span []sheetDay, unit units.Unit, daily bool) (ymin, ymax float64, ok bool) {
	first := true
	add := func(kg float64) {
		v, err := units.FromKG(kg, unit)
		if err != nil {
			return
		}
		if first {
			ymin, ymax, first, ok = v, v, false, true
			return
		}
		if v < ymin {
			ymin = v
		}
		if v > ymax {
			ymax = v
		}
	}
	for _, d := range span {
		if daily && d.Log.Weight != nil {
			add(*d.Log.Weight)
		}
		if d.HasTrend {
			add(d.Trend)
		}
	}
	if !ok {
		return 0, 0, false
	}
	if ymax == ymin {
		ymin -= 0.5
		ymax += 0.5
		return ymin, ymax, true
	}
	pad := (ymax - ymin) * 0.05
	return ymin - pad, ymax + pad, true
}
