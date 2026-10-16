package tui

// The long-term Online picture is one WEIGHT-menu window of The
// Hacker's Diet website chart. It lives here so longchart.go can keep
// the keys and the fit, and so the monthly painter does not grow a
// second screen. The canvas, the one-step scale, and the Braille
// trend are the monthly ones. The series that changes is the weight:
// a thin line through weighed days, not a diamond and not a stem.
// A window longer than 62 days still draws both series, by sharing
// those columns instead of dropping the weight line. Days after today
// are not in the window.
//
// The heading is the window name only. Month labels are one line
// under the plot, and a label that would overlap is skipped. This
// file does not fit a regression, does not write a PDF, and does not
// add a colour role or a custom date range.

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/canvas"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/graph"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/runes"
	"github.com/jglueckstein/hdtools/internal/units"
)

// monthLabelWidth is "Sep 90". A second line is not used, so a name
// that does not fit is skipped rather than stacked.
const monthLabelWidth = 6

// longHeading is the window name on the same 67-column line as the
// monthly heading. The date range stays in the month labels. A box
// would put the span back into the title.
func longHeading(name string, p palette) string {
	return lipgloss.PlaceHorizontal(monthLine, lipgloss.Center, p.title.Render(name))
}

// renderLongOnline paints one non-empty window. The caller already
// dropped an empty span. Trend goes down first so the weight line
// keeps a shared cell, which is the monthly diamond-wins order
// without the diamond.
func renderLongOnline(span []sheetDay, unit units.Unit, p palette) string {
	samples, min, max, ok := longSamples(span, unit)
	if !ok {
		return ""
	}
	lo, hi, ticks := onlineAxis(min, max, unit)
	placed := placeOnlineTicks(ticks, lo, hi)
	share := longShare(len(span))
	m := canvas.New(monthCols, monthRows)
	paintTrend(&m, samples, share, lo, hi, p.trend)
	paintWeightLine(&m, samples, share, lo, hi, p.weight)

	var b strings.Builder
	writeOnlineRows(&b, &m, placed)
	if labels := longMonthLabels(span, share); labels != "" {
		b.WriteString(labels)
		b.WriteByte('\n')
	}
	return b.String()
}

func longSamples(span []sheetDay, unit units.Unit) ([]onlineSample, float64, float64, bool) {
	out := make([]onlineSample, 0, len(span))
	var min, max float64
	ok := false
	add := func(v float64) {
		if !ok || v < min {
			min = v
		}
		if !ok || v > max {
			max = v
		}
		ok = true
	}
	for i, d := range span {
		s := onlineSample{day: i}
		if d.Log.Weight != nil {
			if v, err := units.FromKG(*d.Log.Weight, unit); err == nil {
				s.weight = v
				s.hasWeight = true
				add(v)
			}
		}
		if d.HasTrend {
			if v, err := units.FromKG(d.Trend, unit); err == nil {
				s.trend = v
				s.hasTrend = true
				add(v)
			}
		}
		out = append(out, s)
	}
	return out, min, max, ok
}

// longShare uses the monthly widths when the window fits in 62
// columns. A longer window still gives every day a column, repeating
// one when it must, so a later weigh-in is drawn. monthShares would
// give those later days a width of zero and drop them.
func longShare(n int) monthShare {
	if n <= monthCols {
		return monthShares(n)
	}
	s := monthShare{
		starts: make([]int, n),
		widths: make([]int, n),
	}
	for i := 0; i < n; i++ {
		c := i * monthCols / n
		if c >= monthCols {
			c = monthCols - 1
		}
		s.starts[i] = c
		s.widths[i] = 1
	}
	return s
}

// paintWeightLine draws the weighed days as one thin polyline.
// DrawLinePoints paints nothing from a single point, so each weighed
// day sets a horizontal stroke first and a lone weigh-in stays
// visible. Segments join those days in order. A day with no weight
// adds no point.
func paintWeightLine(m *canvas.Model, samples []onlineSample, share monthShare, lo, hi float64, style lipgloss.Style) {
	var pts []canvas.Point
	for _, s := range samples {
		if !s.hasWeight {
			continue
		}
		c0, c1 := share.bounds(s.day)
		if c1 <= c0 {
			continue
		}
		pt := canvas.Point{X: c0, Y: onlineRow(s.weight, lo, hi)}
		pts = append(pts, pt)
		m.SetRuneWithStyle(pt, runes.LineHorizontal, style)
	}
	for i := 1; i < len(pts); i++ {
		graph.DrawLinePoints(m, graph.GetLinePoints(pts[i-1], pts[i]), runes.ThinLineStyle, style)
	}
}

type monthMark struct {
	col  int
	text string
}

func monthMarks(span []sheetDay, share monthShare) []monthMark {
	var marks []monthMark
	for i, row := range span {
		if i > 0 && row.Day.Day() != 1 {
			continue
		}
		c0, c1 := share.bounds(i)
		if c1 <= c0 {
			continue
		}
		marks = append(marks, monthMark{col: c0, text: row.Day.Format("Jan 06")})
	}
	return marks
}

// placeMonthLabels keeps the first month and, when it fits, the last.
// Middles that would run into either name are skipped. The last name
// slides left when its column would run off the 62-column plot, so
// the window's final month stays readable.
func placeMonthLabels(marks []monthMark) []monthMark {
	if len(marks) == 0 {
		return nil
	}
	first := marks[0]
	first.col = visibleLabelCol(first.col)
	out := []monthMark{first}
	if len(marks) == 1 {
		return out
	}
	last := marks[len(marks)-1]
	last.col = visibleLabelCol(last.col)
	reserveLast := last.col >= first.col+monthLabelWidth
	right := monthCols
	if reserveLast {
		right = last.col
	}
	prevEnd := first.col + monthLabelWidth
	for _, m := range marks[1 : len(marks)-1] {
		if m.col+monthLabelWidth > monthCols {
			continue
		}
		if m.col >= prevEnd && m.col+monthLabelWidth <= right {
			out = append(out, m)
			prevEnd = m.col + monthLabelWidth
		}
	}
	if reserveLast {
		out = append(out, last)
	}
	return out
}

func visibleLabelCol(col int) int {
	if col < 0 {
		return 0
	}
	if col+monthLabelWidth > monthCols {
		col = monthCols - monthLabelWidth
	}
	if col < 0 {
		return 0
	}
	return col
}

func longMonthLabels(span []sheetDay, share monthShare) string {
	placed := placeMonthLabels(monthMarks(span, share))
	if len(placed) == 0 {
		return ""
	}
	buf := make([]rune, monthCols)
	for i := range buf {
		buf[i] = ' '
	}
	for _, m := range placed {
		for i, r := range m.text {
			at := m.col + i
			if at >= 0 && at < len(buf) {
				buf[at] = r
			}
		}
	}
	return strings.Repeat(" ", monthGutter) + string(buf)
}
