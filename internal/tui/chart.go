package tui

// The monthly chart is the book's signal/noise plot for one sitting:
// daily weight as marks, trend as a path, Monthly Loss and Daily
// Deficit from first and last trend of the plotted span. This file
// does not open SQLite or edit the log.

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const plotRows = 8

func lastPlottedDay(year int, month time.Month) int {
	today := localToday()
	ty, tm, td := today.Date()
	if year > ty || (year == ty && month > tm) {
		return 0
	}
	n := daysInMonth(year, month)
	if year == ty && month == tm && td < n {
		return td
	}
	return n
}

func (a *App) openChart() {
	if a.screen == screenList {
		if len(a.logs) > 0 && a.cursor >= 0 && a.cursor < len(a.logs) {
			a.month = newMonth(a.logs[a.cursor].Day)
		} else {
			a.month = newMonth(localToday())
		}
	}
	a.afterChart = a.screen
	a.screen = screenChart
	a.status = ""
	a.err = nil
}

func (a *App) updateChart(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.screen = a.afterChart
		return a, nil
	case "q", "ctrl+c":
		return a, tea.Quit
	case "[":
		a.month.prevMonth()
	case "]":
		a.month.nextMonth()
	}
	return a, nil
}

func (a *App) chartView() string {
	p := a.pal
	unit := a.cfg.DisplayUnit
	title := fmt.Sprintf("hdtools — %s %d  (weight %s)", a.month.month.String(), a.month.year, unit)
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", p.title.Render(title))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render("db: "+a.dbPath))

	last := lastPlottedDay(a.month.year, a.month.month)
	sheet := buildMonthSheet(a.logs, a.month.year, a.month.month)
	if last <= 0 {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty month)"), p.help.Render("esc back   [ ] month   q quit"))
		return b.String()
	}
	span := sheet[:last]
	if chartEmpty(span) {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(empty month)"), p.help.Render("esc back   [ ] month   q quit"))
		return b.String()
	}

	b.WriteString(renderPlot(span, unit, p))
	if line, ok := analysisLine(span, unit); ok {
		fmt.Fprintf(&b, "\n%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render("esc back   [ ] month   q quit"))
	return b.String()
}

func chartEmpty(span []sheetDay) bool {
	for _, d := range span {
		if d.Log.Weight != nil || d.HasTrend {
			return false
		}
	}
	return true
}

func analysisLine(span []sheetDay, unit units.Unit) (string, bool) {
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
	return fmt.Sprintf("Monthly loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal), true
}

func renderPlot(span []sheetDay, unit units.Unit, p palette) string {
	n := len(span)
	ymin, ymax, ok := yRange(span, unit)
	if !ok {
		return ""
	}
	grid := make([][]rune, plotRows)
	for r := range grid {
		grid[r] = make([]rune, n)
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
	var prevRow int
	var hasPrev bool
	for i, d := range span {
		if !d.HasTrend {
			continue
		}
		y, err := units.FromKG(d.Trend, unit)
		if err != nil {
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
		prevRow = r
		hasPrev = true
	}
	for i, d := range span {
		if d.Log.Weight == nil {
			continue
		}
		y, err := units.FromKG(*d.Log.Weight, unit)
		if err != nil {
			continue
		}
		grid[yToRow(y)][i] = 'o'
	}

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
			case 'o':
				s = p.weight.Render(s)
			case '-', '/', '\\':
				s = p.trend.Render(s)
			}
			b.WriteString(s)
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s%d", strings.Repeat(" ", 6), 1)
	if n > 1 {
		last := fmt.Sprintf("%d", span[n-1].Day.Day())
		pad := n - 1 - len(last)
		if pad < 1 {
			pad = 1
		}
		fmt.Fprintf(&b, "%s%s", strings.Repeat(" ", pad), last)
	}
	b.WriteByte('\n')
	return b.String()
}

func yRange(span []sheetDay, unit units.Unit) (ymin, ymax float64, ok bool) {
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
		if d.Log.Weight != nil {
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
