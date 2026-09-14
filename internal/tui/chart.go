package tui

// The monthly chart is the book's signal/noise plot for one sitting:
// a month-year title box, daily weight as marks, green stems from
// each mark to the trend (floats and sinkers), the trend path, and
// Monthly Loss and Daily Deficit from first and last trend of the
// plotted span. Title-box and stem colours are Excel defaults, not
// [colors] roles. Clip, empty, Y pad, and analysis live in
// chartspan so the PDF cannot drift. p writes that picture through
// chartpdf; this file does not import a PDF library.

import (
	"fmt"
	"math"
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
	for i, d := range span.Points {
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
	for i, d := range span.Points {
		if d.Weight == nil || !d.HasTrend {
			continue
		}
		wy, errW := units.FromKG(*d.Weight, unit)
		ty, errT := units.FromKG(d.Trend, unit)
		if errW != nil || errT != nil {
			continue
		}
		wr, tr := yToRow(wy), yToRow(ty)
		if wr == tr {
			continue
		}
		lo, hi := wr, tr
		if lo > hi {
			lo, hi = hi, lo
		}
		for r := lo + 1; r < hi; r++ {
			grid[r][i] = '|'
		}
	}
	for i, d := range span.Points {
		if d.Weight == nil {
			continue
		}
		y, err := units.FromKG(*d.Weight, unit)
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
			case '|':
				s = stemCell(s)
			}
			b.WriteString(s)
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
