package tui

// Braille dots are the on-screen chart paint. A day or a bucket stays
// one character column; the two dot columns are the two horizontal
// positions, and four dots replace the old eight inclusive rows. Both
// the monthly chart and the long-term chart call this file so the bin
// rule cannot drift, and so neither of those files absorbs the grid.
//
// Bin group 0 sits on the bottom row, toward Ymin, because weight
// still has to increase upward. On the monthly chart a stem dot
// wins the cell, even when a mark or a trend dot shares it. A cell
// with no stem keeps the mark, then the trend. On the long-term
// chart the trend wins, and that cell stays bold. Stems are monthly
// only, and only in the mark's own column. This file does not
// choose the Y pad, the span, the buckets, or the PDF.

import "math"

const (
	brailleDotRows = 4
	// Four dots in each of the eight character rows.
	brailleBins = plotRows * brailleDotRows
)

const (
	dotMark = iota
	dotStem
	dotTrend
	dotPath
	dotKindCount
)

// dotMasks keeps each series in its own mask. The rune is their union,
// but the colour is not: one series has to win the whole cell because
// a terminal glyph has one foreground.
type dotMasks struct {
	bits [dotKindCount]int
}

type brailleGrid struct {
	cells   [][]dotMasks
	cols    int
	monthly bool
}

func newBrailleGrid(cols int, monthly bool) *brailleGrid {
	cells := make([][]dotMasks, plotRows)
	for r := range cells {
		cells[r] = make([]dotMasks, cols)
	}
	return &brailleGrid{cells: cells, cols: cols, monthly: monthly}
}

// plotMonthly lights one day. The mark is the left dot. The trend
// takes both dots so a carry-forward day still has two columns. The
// stem is the left dots strictly between those bins; the same bin
// means the mark is already on the trend and a stem would be a smear.
func (g *brailleGrid) plotMonthly(col int, weightY float64, hasWeight bool, trendY float64, hasTrend bool, ymin, ymax float64) {
	var markBin, trendBin int
	if hasWeight {
		markBin = brailleBin(weightY, ymin, ymax, brailleBins)
		g.light(col, markBin, 0, dotMark)
	}
	if hasTrend {
		trendBin = brailleBin(trendY, ymin, ymax, brailleBins)
		g.light(col, trendBin, 0, dotTrend)
		g.light(col, trendBin, 1, dotTrend)
	}
	if !hasWeight || !hasTrend || markBin == trendBin {
		return
	}
	lo, hi := markBin, trendBin
	if lo > hi {
		lo, hi = hi, lo
	}
	for b := lo + 1; b < hi; b++ {
		g.light(col, b, 0, dotStem)
	}
}

// plotLong lights a long-term column. The weight path is the left dot
// only, the thin line, and only when the span is still one day per
// column. The trend uses both dots and outranks the weight path when
// they share a cell. There is no stem.
func (g *brailleGrid) plotLong(col int, weightY float64, hasWeight bool, trendY float64, hasTrend bool, ymin, ymax float64) {
	if hasWeight {
		g.light(col, brailleBin(weightY, ymin, ymax, brailleBins), 0, dotPath)
	}
	if !hasTrend {
		return
	}
	b := brailleBin(trendY, ymin, ymax, brailleBins)
	g.light(col, b, 0, dotTrend)
	g.light(col, b, 1, dotTrend)
}

func (g *brailleGrid) paint(row, col int, p palette) string {
	if row < 0 || row >= len(g.cells) || col < 0 || col >= g.cols {
		return " "
	}
	d := g.cells[row][col]
	mask := d.mask()
	if mask == 0 {
		return " "
	}
	ch := string(rune(0x2800 + mask))
	switch d.tone(g.monthly) {
	case toneWeight:
		return p.weight.Render(ch)
	case toneStem:
		return stemCell(ch)
	case toneTrend:
		return p.trend.Render(ch)
	default:
		return " "
	}
}

const (
	toneBlank = iota
	toneWeight
	toneStem
	toneTrend
)

// tone is the one-colour rule. Monthly: any stem dot paints the
// cell stem green, so the stem is one colour even when a mark or
// a trend dot shares the cell. A cell with no stem keeps the mark,
// then the trend. Long-term charts have no stem: a trend dot takes
// the trend role (already bold on the palette), and a weight-only
// cell stays the weight role and is not bold.
func (d dotMasks) tone(monthly bool) int {
	if d.mask() == 0 {
		return toneBlank
	}
	if monthly {
		if d.bits[dotStem] != 0 {
			return toneStem
		}
		if d.bits[dotMark] != 0 {
			return toneWeight
		}
		return toneTrend
	}
	if d.bits[dotTrend] != 0 {
		return toneTrend
	}
	return toneWeight
}

func (d dotMasks) mask() int {
	m := 0
	for _, b := range d.bits {
		m |= b
	}
	return m
}

// light places one dot. side 0 is the left column. Bin group 0 is the
// bottom character row, and bin mod 4 == 0 is the bottom dot, so a
// larger weight is higher on the screen.
func (g *brailleGrid) light(col, bin, side, kind int) {
	if col < 0 || col >= g.cols || bin < 0 || kind < 0 || kind >= dotKindCount {
		return
	}
	group := bin / brailleDotRows
	if group >= plotRows {
		return
	}
	screenRow := (plotRows - 1) - group
	dotFromTop := (brailleDotRows - 1) - (bin % brailleDotRows)
	bit := brailleBit(side, dotFromTop)
	if bit == 0 {
		return
	}
	g.cells[screenRow][col].bits[kind] |= bit
}

// brailleBin is the half-open quantizer. Ymax has to land in the last
// bin; floor of a half-open span would otherwise put it past N-1 and
// the top row would be empty. A flat span uses the middle bin because
// there is no direction to rank.
func brailleBin(v, ymin, ymax float64, n int) int {
	if n <= 0 {
		return 0
	}
	if ymin == ymax {
		return n / 2
	}
	if v >= ymax {
		return n - 1
	}
	bin := int(math.Floor((v - ymin) / (ymax - ymin) * float64(n)))
	if bin < 0 {
		return 0
	}
	if bin > n-1 {
		return n - 1
	}
	return bin
}

// brailleBit is the Unicode Braille mask. Dot row 0 is the top of the
// cell. The left column is dots 1, 2, 3, 7 and the right is 4, 5, 6, 8,
// which is the standard U+2800 bit order, not a top-to-bottom count.
func brailleBit(col, rowFromTop int) int {
	left := [...]int{0x01, 0x02, 0x04, 0x40}
	right := [...]int{0x08, 0x10, 0x20, 0x80}
	if rowFromTop < 0 || rowFromTop >= len(left) {
		return 0
	}
	if col == 0 {
		return left[rowFromTop]
	}
	if col == 1 {
		return right[rowFromTop]
	}
	return 0
}
