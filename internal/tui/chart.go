package tui

// The monthly chart is the book's signal/noise plot for one sitting:
// a month-year title box, daily weight as marks, green stems from
// each mark to the trend (floats and sinkers), the trend path, and
// Monthly Loss and Daily Deficit from first and last trend of the
// plotted span. Title-box and stem colours are Excel defaults, not
// [colors] roles. The dots are Braille runes from braille.go; this
// file still owns the title, the Y labels, and the loss line. Clip,
// empty, Y pad, and analysis live in chartspan so the PDF cannot
// drift. p writes that picture through chartpdf; this file does not
// import a PDF library.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jglueckstein/hdtools/internal/chartpdf"
	"github.com/jglueckstein/hdtools/internal/chartspan"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	plotRows  = 8
	chartHelp = "esc back   [ ] month   l long   p pdf   q quit"
)

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
	case "l":
		a.openLong()
		return a, nil
	case "p":
		a.writeChartPDF()
		return a, nil
	}
	return a, nil
}

func (a *App) writeChartPDF() {
	cwd, err := os.Getwd()
	if err != nil {
		a.err = fmt.Errorf("write chart pdf: %w", err)
		a.status = ""
		return
	}
	name := fmt.Sprintf("%04d-%02d-chart.pdf", a.month.year, a.month.month)
	path := filepath.Join(cwd, name)
	err = chartpdf.Write(path, chartpdf.Options{
		Year:   a.month.year,
		Month:  a.month.month,
		Logs:   a.logs,
		Unit:   a.cfg.DisplayUnit,
		Colors: a.cfg.Colors,
		Today:  localToday(),
	})
	if err != nil {
		a.err = fmt.Errorf("write chart pdf: %w", err)
		a.status = ""
		return
	}
	a.err = nil
	a.status = path
}

func monthYearLabel(month time.Month, year int) string {
	return fmt.Sprintf("%s %d", month.String(), year)
}

func titleBox(text string, plotWidth int) string {
	s := lipgloss.NewStyle().Padding(0, 1)
	if os.Getenv("NO_COLOR") == "" {
		s = s.
			Foreground(lipgloss.Color("3")).
			Background(lipgloss.Color("4")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("1"))
	}
	box := s.Render(text)
	if plotWidth < 1 {
		return box
	}
	return lipgloss.PlaceHorizontal(plotWidth, lipgloss.Center, box)
}

func monthYearBox(month time.Month, year int, plotWidth int) string {
	return titleBox(monthYearLabel(month, year), plotWidth)
}

func stemCell(ch string) string {
	if os.Getenv("NO_COLOR") != "" {
		return ch
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(ch)
}

func (a *App) chartView() string {
	p := a.pal
	unit := a.cfg.DisplayUnit
	span := chartspan.Clip(a.logs, a.month.year, a.month.month, localToday())
	last := len(span.Points)
	plotWidth := 6 + last
	if last <= 0 {
		plotWidth = 6 + daysInMonth(a.month.year, a.month.month)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", monthYearBox(a.month.month, a.month.year, plotWidth))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))

	if last <= 0 || span.Empty() {
		fmt.Fprintf(&b, "%s\n", p.muted.Render("(empty month)"))
		b.WriteString(a.chartFooter())
		return b.String()
	}

	b.WriteString(renderPlot(span, unit, p))
	if line, ok := span.Analysis(unit); ok {
		fmt.Fprintf(&b, "\n%s\n", line)
	}
	b.WriteString(a.chartFooter())
	return b.String()
}

func (a *App) chartFooter() string {
	p := a.pal
	var b strings.Builder
	if a.err != nil {
		fmt.Fprintf(&b, "\n%s\n", p.error.Render("error: "+a.err.Error()))
	}
	if a.status != "" {
		fmt.Fprintf(&b, "\n%s\n", p.status.Render(a.status))
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render(chartHelp))
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

func renderPlot(span chartspan.Span, unit units.Unit, p palette) string {
	n := len(span.Points)
	ymin, ymax, ok := chartspan.YRange(span, unit)
	if !ok {
		return ""
	}
	grid := newBrailleGrid(n, true)
	for i, d := range span.Points {
		var wy, ty float64
		hw, ht := false, false
		if d.Weight != nil {
			y, err := units.FromKG(*d.Weight, unit)
			if err == nil {
				wy, hw = y, true
			}
		}
		if d.HasTrend {
			y, err := units.FromKG(d.Trend, unit)
			if err == nil {
				ty, ht = y, true
			}
		}
		grid.plotMonthly(i, wy, hw, ty, ht, ymin, ymax)
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
			b.WriteString(grid.paint(r, c, p))
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s%d", strings.Repeat(" ", 6), 1)
	if n > 1 {
		last := fmt.Sprintf("%d", span.Points[n-1].Day.Day())
		pad := n - 1 - len(last)
		if pad < 1 {
			pad = 1
		}
		fmt.Fprintf(&b, "%s%s", strings.Repeat(" ", pad), last)
	}
	b.WriteByte('\n')
	return b.String()
}

func yPad(unit units.Unit) float64 {
	return chartspan.YPad(unit)
}

func applyYPad(ymin, ymax float64, unit units.Unit) (float64, float64) {
	p := yPad(unit)
	return ymin - p, ymax + p
}
