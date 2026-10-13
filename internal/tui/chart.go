package tui

// The monthly chart screen is one sitting of the book: the keys, the
// PDF write, the database line, and Monthly Loss / Daily Deficit.
// The on-screen picture is the Online chart in onlinechart.go. The
// yellow-on-blue title box stays in this file because the long-term
// chart still uses it. Clip, empty, the ±2 lb pad, and the caption
// live in chartspan so the PDF cannot drift. p writes that PDF
// through chartpdf into the configured directory, not the process
// working directory. This file does not import a PDF library and
// does not paint the long-term chart.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jglueckstein/hdtools/internal/chartpdf"
	"github.com/jglueckstein/hdtools/internal/chartspan"
	"github.com/jglueckstein/hdtools/internal/config"
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

func (a *App) updateChart(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

// writeChartPDF uses the same directory as an empty -o. A failed
// resolve or write leaves the chart on screen and does not fall
// back to the working directory.
func (a *App) writeChartPDF() {
	dir, err := config.ResolvePDFDir(a.cfg)
	if err != nil {
		a.err = fmt.Errorf("write chart pdf: %w", err)
		a.status = ""
		return
	}
	name := fmt.Sprintf("%04d-%02d-chart.pdf", a.month.year, a.month.month)
	path := filepath.Join(dir, name)
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
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", onlineHeading(a.month.month, a.month.year, p))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render(fmt.Sprintf("db: %s   weight %s", a.dbPath, unit)))

	if len(span.Points) == 0 || span.Empty() {
		fmt.Fprintf(&b, "%s\n", p.muted.Render("(empty month)"))
		b.WriteString(a.chartFooter())
		return b.String()
	}

	b.WriteString(renderOnline(span, a.month.year, a.month.month, unit, p))
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

func yPad(unit units.Unit) float64 {
	return chartspan.YPad(unit)
}

func applyYPad(ymin, ymax float64, unit units.Unit) (float64, float64) {
	p := yPad(unit)
	return ymin - p, ymax + p
}
