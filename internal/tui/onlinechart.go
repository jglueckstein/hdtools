package tui

// The monthly Online picture is one calendar month of The Hacker's Diet
// website chart. It lives in this file so chart.go can keep the keys,
// the PDF, and the caption. The long-term screen reuses this canvas
// from longonline.go. app.go is already at the file-size line.
//
// The plot is 62 columns by 16 rows, plus a five-column gutter, and it
// does not follow the terminal. Y is the display unit. The scale runs
// one step past the plotted min and max, with no ±2 lb pad. Each
// calendar day gets one share of the 62 columns from the length of the
// month alone, so a day is the same width on the 10th as in a finished
// month. Days after today stay empty slots. The trend is the ApplyTrend
// value already on the log, drawn as Braille, including carry onto blank
// days. A weighed day puts an open diamond in the first column of its
// share. The stem is the line between that day's weight and that day's
// trend, and only when those two land on different rows. Trend is drawn
// first, then stems, then diamonds, so the diamond keeps its own cell.
//
// This file does not resize, does not change the long-term gutter or
// the PDF, does not fit a regression, and does not add colour roles.
// It does not turn on mouse, zoom, a grid, or a legend.

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/canvas"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/graph"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/runes"
	"github.com/jglueckstein/hdtools/internal/chartspan"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	monthCols   = 62
	monthRows   = 16
	monthGutter = 5
	monthLine   = monthGutter + monthCols
	// Open diamond, not the filled U+25C6 the website uses for a flag.
	diamondRune = '\u25c7'
	// Keep a tick that float error put a hair outside the closed interval.
	onlineAxisEps = 1e-6
	// A stone step is one pound. units keeps this factor private.
	lbPerStone = 14.0
)

type onlineTick struct {
	value float64
	label string
}

type monthShare struct {
	starts []int
	widths []int
}

func (s monthShare) bounds(day int) (int, int) {
	if day < 0 || day >= len(s.starts) {
		return 0, 0
	}
	return s.starts[day], s.starts[day] + s.widths[day]
}

// onlineHeading is the month and year in the existing title role.
// The words are centered on the 67-column plot line. There is no box
// and no filled background. The long-term window name uses the same
// line, from longonline.go.
func onlineHeading(month time.Month, year int, p palette) string {
	label := p.title.Render(monthYearLabel(month, year))
	return lipgloss.PlaceHorizontal(monthLine, lipgloss.Center, label)
}

// renderOnline paints one non-empty month. The caller already dropped
// an empty span, so this function does not print "(empty month)".
func renderOnline(span chartspan.Span, year int, month time.Month, unit units.Unit, p palette) string {
	samples, min, max, ok := onlineSamples(span, unit)
	if !ok {
		return ""
	}
	lo, hi, ticks := onlineAxis(min, max, unit)
	placed := placeOnlineTicks(ticks, lo, hi)
	share := monthShares(daysInMonth(year, month))
	m := canvas.New(monthCols, monthRows)
	paintTrend(&m, samples, share, lo, hi, p.trend)
	paintStems(&m, samples, share, lo, hi)
	paintDiamonds(&m, samples, share, lo, hi, p.weight)

	var b strings.Builder
	writeOnlineRows(&b, &m, placed)
	b.WriteString(onlineDayLabels(samples, share))
	b.WriteByte('\n')
	return b.String()
}

type onlineSample struct {
	day       int
	weight    float64
	trend     float64
	hasWeight bool
	hasTrend  bool
}

func onlineSamples(span chartspan.Span, unit units.Unit) ([]onlineSample, float64, float64, bool) {
	out := make([]onlineSample, 0, len(span.Points))
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
	for _, p := range span.Points {
		s := onlineSample{day: p.Day.Day() - 1}
		if p.Weight != nil {
			if v, err := units.FromKG(*p.Weight, unit); err == nil {
				s.weight = v
				s.hasWeight = true
				add(v)
			}
		}
		if p.HasTrend {
			if v, err := units.FromKG(p.Trend, unit); err == nil {
				s.trend = v
				s.hasTrend = true
				add(v)
			}
		}
		out = append(out, s)
	}
	return out, min, max, ok
}

// onlineAxis builds the one-step scale. Kilograms and pounds step by
// one. Stone steps by one pound and prints that pound as one decimal
// of stones, so two pounds can share a label and both still appear.
// The ends are not snapped to a further whole number.
func onlineAxis(min, max float64, unit units.Unit) (float64, float64, []onlineTick) {
	if unit == units.Stone {
		return stoneAxis(min, max)
	}
	lo := min - 1
	hi := max + 1
	start := int(math.Ceil(lo - onlineAxisEps))
	end := int(math.Floor(hi + onlineAxisEps))
	ticks := make([]onlineTick, 0, end-start+1)
	for v := end; v >= start; v-- {
		ticks = append(ticks, onlineTick{value: float64(v), label: strconv.Itoa(v)})
	}
	return lo, hi, ticks
}

func stoneAxis(min, max float64) (float64, float64, []onlineTick) {
	minLb := min * lbPerStone
	maxLb := max * lbPerStone
	// The canvas is one pound past the data even when that end is
	// not a whole pound. Ticks are the whole pounds inside it.
	// Using those ticks as the canvas would snap inward.
	lo := (minLb - 1) / lbPerStone
	hi := (maxLb + 1) / lbPerStone
	low := int(math.Ceil(minLb - 1 - onlineAxisEps))
	high := int(math.Floor(maxLb + 1 + onlineAxisEps))
	if high < low {
		high = low
	}
	ticks := make([]onlineTick, 0, high-low+1)
	for lb := high; lb >= low; lb-- {
		v := float64(lb) / lbPerStone
		ticks = append(ticks, onlineTick{value: v, label: fmt.Sprintf("%.1f", v)})
	}
	return lo, hi, ticks
}

// onlineRow maps a display value onto a character row. Row 0 is the
// high end, so the top gutter is the high tick.
func onlineRow(v, lo, hi float64) int {
	if monthRows <= 1 || hi == lo {
		return 0
	}
	r := int(math.Round((hi - v) / (hi - lo) * float64(monthRows-1)))
	if r < 0 {
		return 0
	}
	if r >= monthRows {
		return monthRows - 1
	}
	return r
}

// placeOnlineTicks keeps the end ticks and then as many middle ticks
// as the 16 rows can hold. A tick never shares a row. Ends are placed
// first so a crowded scale drops a middle label instead of 39 or 121.
func placeOnlineTicks(ticks []onlineTick, lo, hi float64) map[int]string {
	out := make(map[int]string, monthRows)
	if len(ticks) == 0 {
		return out
	}
	place := func(t onlineTick) {
		row := onlineRow(t.value, lo, hi)
		if _, exists := out[row]; exists {
			return
		}
		out[row] = t.label
	}
	place(ticks[0])
	if len(ticks) > 1 {
		place(ticks[len(ticks)-1])
	}
	if len(ticks) > 2 {
		for _, t := range ticks[1 : len(ticks)-1] {
			place(t)
		}
	}
	return out
}

// monthShares spreads 62 columns across the calendar days. The extra
// columns go to the earliest days. A 31-day month has no extra, so it
// is two columns each. The tests accept any even split; this one leaves
// the tail of an unfinished month empty.
func monthShares(nDays int) monthShare {
	if nDays <= 0 {
		return monthShare{}
	}
	base := monthCols / nDays
	rem := monthCols % nDays
	s := monthShare{
		starts: make([]int, nDays),
		widths: make([]int, nDays),
	}
	col := 0
	for d := 0; d < nDays; d++ {
		w := base
		if d < rem {
			w++
		}
		s.starts[d] = col
		s.widths[d] = w
		col += w
	}
	return s
}

// paintTrend draws ApplyTrend as Braille, including blank days that
// only carry. Dots are set on the character row onlineRow chose.
// BrailleGrid.GridPoint would use the four-dot cell and land beside
// the tick rows. The other columns of the day keep the trend after
// the stem replaces the diamond's column.
func paintTrend(m *canvas.Model, samples []onlineSample, share monthShare, lo, hi float64, style lipgloss.Style) {
	grid := graph.NewBrailleGrid(monthCols, monthRows, 0, 1, lo, hi)
	prevCol := -1
	prevRow := 0
	for _, s := range samples {
		if !s.hasTrend {
			prevCol = -1
			continue
		}
		row := onlineRow(s.trend, lo, hi)
		c0, c1 := share.bounds(s.day)
		for c := c0; c < c1; c++ {
			lightTrend(grid, c, row)
		}
		if prevCol >= 0 {
			from := canvas.Point{X: prevCol, Y: prevRow}
			to := canvas.Point{X: c0, Y: row}
			for _, p := range graph.GetLinePoints(from, to) {
				lightTrend(grid, p.X, p.Y)
			}
		}
		if c1 > c0 {
			prevCol = c1 - 1
			prevRow = row
		}
	}
	graph.DrawBraillePatterns(m, canvas.Point{}, grid.BraillePatterns(), style)
}

func lightTrend(grid *graph.BrailleGrid, col, row int) {
	if col < 0 || col >= monthCols || row < 0 || row >= monthRows {
		return
	}
	// One dot is enough for the rune. The top of the character cell
	// is the row the tick scale uses.
	grid.Set(canvas.Point{X: col * 2, Y: row * 4})
}

// paintStems draws the segment from that day's weight to that day's
// trend. GetLinePoints stays inside the segment, where a vertical
// helper would paint past it. The same row means no stem, including
// equal values. The library's own line leaves a half-tip on each
// end. That tip sits in the wrong half of the trend cell, and a
// one-row gap has no middle bar once the diamond covers the weight
// cell, so every cell of the segment is the full vertical.
func paintStems(m *canvas.Model, samples []onlineSample, share monthShare, lo, hi float64) {
	style := onlineStemStyle()
	for _, s := range samples {
		if !s.hasWeight || !s.hasTrend {
			continue
		}
		weightRow := onlineRow(s.weight, lo, hi)
		trendRow := onlineRow(s.trend, lo, hi)
		if weightRow == trendRow {
			continue
		}
		c0, c1 := share.bounds(s.day)
		if c1 <= c0 {
			continue
		}
		from := canvas.Point{X: c0, Y: weightRow}
		to := canvas.Point{X: c0, Y: trendRow}
		for _, p := range graph.GetLinePoints(from, to) {
			m.SetRuneWithStyle(p, runes.LineVertical, style)
		}
	}
}

// onlineStemStyle is chrome green, the monthly stem, and not a
// [colors] role. NO_COLOR drops the foreground and keeps the rune.
func onlineStemStyle() lipgloss.Style {
	if os.Getenv("NO_COLOR") != "" {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
}

func paintDiamonds(m *canvas.Model, samples []onlineSample, share monthShare, lo, hi float64, style lipgloss.Style) {
	for _, s := range samples {
		if !s.hasWeight {
			continue
		}
		c0, c1 := share.bounds(s.day)
		if c1 <= c0 {
			continue
		}
		row := onlineRow(s.weight, lo, hi)
		m.SetRuneWithStyle(canvas.Point{X: c0, Y: row}, diamondRune, style)
	}
}

func writeOnlineRows(b *strings.Builder, m *canvas.Model, ticks map[int]string) {
	for row := 0; row < monthRows; row++ {
		fmt.Fprintf(b, "%4s-", ticks[row])
		for x := 0; x < monthCols; x++ {
			cell := m.Cell(canvas.Point{X: x, Y: row})
			if cell.Rune == 0 {
				b.WriteByte(' ')
				continue
			}
			b.WriteString(cell.Style.Render(string(cell.Rune)))
		}
		b.WriteByte('\n')
	}
}

// onlineDayLabels names the first and last days that were drawn, in
// that day's columns. Clip begins on day 1, so a series that starts
// later must not label those blank days, and the empty tail of the
// current month stays blank. The line is not a plot row: its fifth
// cell is a space.
func onlineDayLabels(samples []onlineSample, share monthShare) string {
	first := -1
	last := -1
	for _, s := range samples {
		if !s.hasWeight && !s.hasTrend {
			continue
		}
		if first < 0 {
			first = s.day
		}
		last = s.day
	}
	if first < 0 {
		return ""
	}
	buf := make([]rune, monthCols)
	for i := range buf {
		buf[i] = ' '
	}
	put := func(day int) {
		text := strconv.Itoa(day + 1)
		col, _ := share.bounds(day)
		for i, r := range text {
			if col+i >= 0 && col+i < len(buf) {
				buf[col+i] = r
			}
		}
	}
	put(first)
	if last != first {
		put(last)
	}
	return strings.Repeat(" ", monthGutter) + string(buf)
}
