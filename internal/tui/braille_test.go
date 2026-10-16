package tui

// These helpers read styled runes back out of a chart view. The
// Online pictures use them. The old eight-row bin oracle left with
// the Braille-cell grid. This file does not paint.

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	brailleRows = 8
	brailleDots = 4
	brailleN    = brailleRows * brailleDots
)

// Top of the cell is index 0, matching Decision 1.
var (
	brailleLeftFromTop  = []int{0x01, 0x02, 0x04, 0x40}
	brailleRightFromTop = []int{0x08, 0x10, 0x20, 0x80}
)

type cellStyle struct {
	r    rune
	bold bool
	seq  string
}

// isPlotGutter is true for a "%5.1f-" Y label. Braille lines do not
// contain o, -, /, \, or |, so those glyphs cannot mark a plot row.
func isPlotGutter(rs []rune) bool {
	if len(rs) < 6 || rs[5] != '-' {
		return false
	}
	var v float64
	_, err := fmt.Sscanf(strings.TrimSpace(string(rs[:5])), "%f", &v)
	return err == nil
}

func isBraille(r rune) bool {
	return r >= 0x2800 && r <= 0x28FF
}

func maskOf(r rune) int {
	if !isBraille(r) {
		return 0
	}
	return int(r - 0x2800)
}

func sideBits(left bool) int {
	bits := brailleLeftFromTop
	if !left {
		bits = brailleRightFromTop
	}
	m := 0
	for _, b := range bits {
		m |= b
	}
	return m
}

func styledRunes(s string) []cellStyle {
	bold := false
	var fg []int
	var out []cellStyle
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j >= len(s) {
				break
			}
			bold, fg = applySGR(bold, fg, s[i+2:j])
			i = j + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size <= 0 {
			break
		}
		out = append(out, cellStyle{r: r, bold: bold, seq: styleSeq(bold, fg)})
		i += size
	}
	return out
}

func applySGR(bold bool, fg []int, body string) (bool, []int) {
	if body == "" {
		return false, nil
	}
	parts := strings.Split(body, ";")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			nums = append(nums, 0)
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			bold = false
			fg = nil
		case n == 1:
			bold = true
		case n == 22:
			bold = false
		case n == 39:
			fg = nil
		case (n >= 30 && n <= 37) || (n >= 90 && n <= 97):
			fg = []int{n}
		case n == 38 && i+2 < len(nums) && nums[i+1] == 5:
			fg = []int{38, 5, nums[i+2]}
			i += 2
		case n == 38 && i+4 < len(nums) && nums[i+1] == 2:
			fg = []int{38, 2, nums[i+2], nums[i+3], nums[i+4]}
			i += 4
		}
	}
	return bold, fg
}

func styleSeq(bold bool, fg []int) string {
	var parts []string
	if bold {
		parts = append(parts, "1")
	}
	for _, n := range fg {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ";")
}

func cellHasFG(c cellStyle, n int) bool {
	if c.seq == "" {
		return false
	}
	return hasIndexedForeground("\x1b["+c.seq+"m", n)
}

func plotCells(raw string) [][]cellStyle {
	var rows [][]cellStyle
	for _, line := range strings.Split(raw, "\n") {
		cells := styledRunes(line)
		if len(cells) < 6 {
			continue
		}
		rs := make([]rune, len(cells))
		for i, c := range cells {
			rs[i] = c.r
		}
		if !isPlotGutter(rs) {
			continue
		}
		rows = append(rows, cells[6:])
	}
	return rows
}

func mustPlot(t *testing.T, view string) [][]cellStyle {
	t.Helper()
	rows := plotCells(view)
	if len(rows) != brailleRows {
		t.Fatalf("plot rows = %d, want %d\n%s", len(rows), brailleRows, visible(view))
	}
	return rows
}

func plotWidth(rows [][]cellStyle) int {
	if len(rows) == 0 {
		return 0
	}
	w := len(rows[0])
	for _, row := range rows {
		if len(row) != w {
			return -1
		}
	}
	return w
}

func columnGlyphs(rows [][]cellStyle, day int) string {
	var b strings.Builder
	for _, row := range rows {
		if day < 0 || day >= len(row) || row[day].r == ' ' || row[day].r == 0 {
			b.WriteRune('·')
			continue
		}
		b.WriteRune(row[day].r)
	}
	return b.String()
}

func firstPlotGlyph(rows [][]cellStyle) rune {
	for _, row := range rows {
		for _, c := range row {
			if c.r != ' ' && c.r != 0 {
				return c.r
			}
		}
	}
	return 0
}

// dotBins reads one side of a cell. Bin group 0 is the bottom row,
// toward Ymin, and bin mod 4 == 0 is the bottom dot of the cell.
func dotBins(screenRow, mask int, left bool) []int {
	bits := brailleLeftFromTop
	if !left {
		bits = brailleRightFromTop
	}
	group := (brailleRows - 1) - screenRow
	var out []int
	for dotRow, bit := range bits {
		if mask&bit == 0 {
			continue
		}
		out = append(out, group*brailleDots+(brailleDots-1-dotRow))
	}
	return out
}

func binsOn(rows [][]cellStyle, day int, left bool) []int {
	var out []int
	for sr, row := range rows {
		if day < 0 || day >= len(row) {
			continue
		}
		out = append(out, dotBins(sr, maskOf(row[day].r), left)...)
	}
	return out
}

func leftBins(rows [][]cellStyle, day int) []int {
	return binsOn(rows, day, true)
}

func rightBins(rows [][]cellStyle, day int) []int {
	return binsOn(rows, day, false)
}

func binRow(rows [][]cellStyle, day, bin int, left bool) (cellStyle, int, bool) {
	for sr, row := range rows {
		if day < 0 || day >= len(row) {
			continue
		}
		for _, b := range dotBins(sr, maskOf(row[day].r), left) {
			if b == bin {
				return row[day], sr, true
			}
		}
	}
	return cellStyle{}, -1, false
}

func hasLeftBin(rows [][]cellStyle, day, bin int) bool {
	_, _, ok := binRow(rows, day, bin, true)
	return ok
}

func hasRightBin(rows [][]cellStyle, day, bin int) bool {
	_, _, ok := binRow(rows, day, bin, false)
	return ok
}

func hasBin(rows [][]cellStyle, day, bin int) bool {
	return hasLeftBin(rows, day, bin) || hasRightBin(rows, day, bin)
}

func columnHasBothSides(rows [][]cellStyle, day int) bool {
	for _, row := range rows {
		if day < 0 || day >= len(row) {
			continue
		}
		m := maskOf(row[day].r)
		if m&sideBits(true) != 0 && m&sideBits(false) != 0 {
			return true
		}
	}
	return false
}

func sameSet(got []int, want ...int) bool {
	if len(got) != len(want) {
		return false
	}
	have := make(map[int]int, len(got))
	for _, g := range got {
		have[g]++
	}
	for _, w := range want {
		have[w]--
		if have[w] < 0 {
			return false
		}
	}
	for _, n := range have {
		if n != 0 {
			return false
		}
	}
	return true
}
