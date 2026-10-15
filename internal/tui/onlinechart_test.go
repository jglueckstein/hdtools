package tui

// Online chart tests are the RED contract for the monthly Hacker's Diet
// Online picture. They drive App.Update and read View. They do not need
// a terminal. A plot row is 67 visible columns, measured with
// lipgloss.Width: a 5-column gutter, then 62 plot columns. Plot column
// 1 is the first column after the gutter.
//
// The spec does not say which days of a 28- or 30-day month receive a
// remainder column, so the share search accepts every even split. It
// does not treat the long-term chart as 62 or 67 columns, and it does
// not widen isPlotGutter. mustPlot, isPlotGutter, and chartYBounds stay
// on the long-term %5.1f- grid.

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	onlineGutterCols = 5
	onlinePlotCols   = 62
	onlineLineCols   = onlineGutterCols + onlinePlotCols
	onlinePlotRows   = 16
	openDiamond      = '\u25c7'
	filledDiamond    = '\u25c6'
)

type onlinePlot struct {
	rows        [][]cellStyle
	tickLabels  []string
	tickGutters []string
	tickRows    []int
}

type logPoint struct {
	day     time.Time
	kg      float64
	workout bool
}

type dayShare struct {
	starts []int
	nDays  int
	cols   int
}

func (s dayShare) bounds(day int) (int, int) {
	lo := s.starts[day]
	hi := s.cols
	if day+1 < s.nDays {
		hi = s.starts[day+1]
	}
	return lo, hi
}

func ymd(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func loadOnline(t *testing.T, today time.Time, cfg config.Config, points ...logPoint) *App {
	t.Helper()
	freezeToday(t, today)
	store := openStore(t)
	ctx := context.Background()
	for _, p := range points {
		kg := p.kg
		log, err := dailylog.New(p.day, &kg, 8, 0, p.workout, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	return app
}

func showOnline(t *testing.T, today time.Time, cfg config.Config, points ...logPoint) *App {
	t.Helper()
	app := loadOnline(t, today, cfg, points...)
	press(app, "c")
	return app
}

func kgPoint(month time.Month, day int, kg float64) logPoint {
	return logPoint{day: ymd(1990, month, day), kg: kg}
}

func isStemRune(r rune) bool {
	if r == ' ' || r == 0 {
		return false
	}
	return !isBraille(r) && r != openDiamond && r != filledDiamond
}

func columnCells(line string) []cellStyle {
	styled := styledRunes(line)
	var out []cellStyle
	for _, c := range styled {
		w := lipgloss.Width(string(c.r))
		if w < 1 {
			w = 1
		}
		for i := 0; i < w; i++ {
			out = append(out, c)
		}
	}
	return out
}

func onlineScaleLine(line string) ([]cellStyle, bool) {
	if lipgloss.Width(line) != onlineLineCols {
		return nil, false
	}
	cells := columnCells(line)
	if len(cells) != onlineLineCols || cells[onlineGutterCols-1].r != '-' {
		return nil, false
	}
	return cells, true
}

func countOnlinePlotRows(view string) int {
	n := 0
	for _, line := range strings.Split(view, "\n") {
		if _, ok := onlineScaleLine(line); ok {
			n++
		}
	}
	return n
}

func parseOnlinePlot(view string) (onlinePlot, error) {
	var p onlinePlot
	for _, line := range strings.Split(view, "\n") {
		cells, ok := onlineScaleLine(line)
		if !ok {
			continue
		}
		gutter := make([]rune, onlineGutterCols)
		for i := 0; i < onlineGutterCols; i++ {
			gutter[i] = cells[i].r
		}
		field := strings.TrimSpace(string(gutter[:onlineGutterCols-1]))
		row := make([]cellStyle, onlinePlotCols)
		copy(row, cells[onlineGutterCols:])
		p.rows = append(p.rows, row)
		if field == "" {
			continue
		}
		if len(strings.Fields(field)) != 1 {
			return onlinePlot{}, fmt.Errorf("gutter %q holds more than one tick", string(gutter))
		}
		p.tickLabels = append(p.tickLabels, field)
		p.tickGutters = append(p.tickGutters, string(gutter))
		p.tickRows = append(p.tickRows, len(p.rows)-1)
	}
	return p, nil
}

func mustOnlinePlot(t *testing.T, view string) onlinePlot {
	t.Helper()
	p, err := parseOnlinePlot(view)
	if err != nil {
		t.Fatalf("monthly online plot: %v", err)
	}
	if len(p.rows) != onlinePlotRows {
		t.Fatalf("online plot rows = %d, want %d at %d visible columns; the monthly screen is still the braille-cell chart\n%s",
			len(p.rows), onlinePlotRows, onlineLineCols, visible(view))
	}
	for _, row := range p.rows {
		if len(row) != onlinePlotCols {
			t.Fatalf("plot columns = %d, want %d", len(row), onlinePlotCols)
		}
	}
	return p
}

func (p onlinePlot) ticksShareRow() bool {
	seen := map[int]bool{}
	for _, row := range p.tickRows {
		if seen[row] {
			return true
		}
		seen[row] = true
	}
	return false
}

func (p onlinePlot) slotHas(lo, hi int, pred func(cellStyle) bool) bool {
	for _, row := range p.rows {
		for c := lo; c < hi && c < len(row); c++ {
			if pred(row[c]) {
				return true
			}
		}
	}
	return false
}

func (p onlinePlot) slotHasBraille(lo, hi int) bool {
	return p.slotHas(lo, hi, func(c cellStyle) bool { return isBraille(c.r) })
}

func (p onlinePlot) slotHasDiamond(lo, hi int) bool {
	return p.slotHas(lo, hi, func(c cellStyle) bool { return c.r == openDiamond })
}

func (p onlinePlot) slotHasStem(lo, hi int) bool {
	return p.slotHas(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
}

func (p onlinePlot) hasBraille() bool {
	return p.slotHasBraille(0, onlinePlotCols)
}

func (p onlinePlot) ink(lo, hi int) bool {
	return p.slotHas(lo, hi, func(c cellStyle) bool { return c.r != ' ' && c.r != 0 })
}

func (p onlinePlot) columnsBlank(lo, hi int) bool {
	return !p.ink(lo, hi)
}

func (p onlinePlot) diamondCols() []int {
	seen := map[int]bool{}
	var cols []int
	for _, row := range p.rows {
		for c, cell := range row {
			if cell.r == openDiamond && !seen[c] {
				seen[c] = true
				cols = append(cols, c)
			}
		}
	}
	sort.Ints(cols)
	return cols
}

func (p onlinePlot) cellsWhere(lo, hi int, pred func(cellStyle) bool) []cellStyle {
	var out []cellStyle
	for _, row := range p.rows {
		for c := lo; c < hi && c < len(row); c++ {
			if pred(row[c]) {
				out = append(out, row[c])
			}
		}
	}
	return out
}

func (p onlinePlot) stemBetween(lo, hi int) bool {
	var dRows, bRows, sRows []int
	for r, row := range p.rows {
		d, b, s := false, false, false
		for c := lo; c < hi && c < len(row); c++ {
			switch {
			case row[c].r == openDiamond:
				d = true
			case isBraille(row[c].r):
				b = true
			case isStemRune(row[c].r):
				s = true
			}
		}
		if d {
			dRows = append(dRows, r)
		}
		if b {
			bRows = append(bRows, r)
		}
		if s {
			sRows = append(sRows, r)
		}
	}
	for _, d := range dRows {
		for _, b := range bRows {
			loR, hiR := d, b
			if loR > hiR {
				loR, hiR = hiR, loR
			}
			for _, s := range sRows {
				if s > loR && s < hiR {
					return true
				}
			}
		}
	}
	return false
}

func chooseDays(n, k int) [][]int {
	if k < 0 || k > n {
		return nil
	}
	var out [][]int
	var rec func(start int, cur []int)
	rec = func(start int, cur []int) {
		if len(cur) == k {
			out = append(out, append([]int(nil), cur...))
			return
		}
		need := k - len(cur)
		for i := start; i <= n-need; i++ {
			next := make([]int, len(cur)+1)
			copy(next, cur)
			next[len(cur)] = i
			rec(i+1, next)
		}
	}
	rec(0, nil)
	return out
}

func allEvenShares(nDays, cols int) []dayShare {
	if nDays <= 0 || cols < nDays {
		return nil
	}
	base := cols / nDays
	rem := cols % nDays
	out := make([]dayShare, 0, 1)
	for _, extra := range chooseDays(nDays, rem) {
		wide := make([]bool, nDays)
		for _, d := range extra {
			wide[d] = true
		}
		starts := make([]int, nDays)
		col := 0
		for d := 0; d < nDays; d++ {
			starts[d] = col
			w := base
			if wide[d] {
				w++
			}
			col += w
		}
		out = append(out, dayShare{starts: starts, nDays: nDays, cols: cols})
	}
	return out
}

func (p onlinePlot) diamondsOnDays(s dayShare, weighed map[int]bool) bool {
	for d := 0; d < s.nDays; d++ {
		lo, hi := s.bounds(d)
		if p.slotHasDiamond(lo, hi) != weighed[d] {
			return false
		}
	}
	return true
}

func (p onlinePlot) brailleDayCount(s dayShare) int {
	n := 0
	for d := 0; d < s.nDays; d++ {
		lo, hi := s.bounds(d)
		if p.slotHasBraille(lo, hi) {
			n++
		}
	}
	return n
}

// bestDayShares keeps the even splits that put diamonds only on weighed
// days and cover the most days with Braille. Callers assert on every
// kept split so a remainder day is not pinned.
func bestDayShares(t *testing.T, p onlinePlot, nDays int, weighed []int) []dayShare {
	t.Helper()
	want := make(map[int]bool, len(weighed))
	for _, d := range weighed {
		want[d] = true
	}
	var best []dayShare
	score := -1
	for _, s := range allEvenShares(nDays, onlinePlotCols) {
		if !p.diamondsOnDays(s, want) {
			continue
		}
		n := p.brailleDayCount(s)
		if n > score {
			score = n
			best = []dayShare{s}
		} else if n == score {
			best = append(best, s)
		}
	}
	if len(best) == 0 {
		t.Fatalf("no even share of %d columns across %d days fits the diamonds", onlinePlotCols, nDays)
	}
	return best
}

func assertNoStem(t *testing.T, p onlinePlot, shares []dayShare, day int) {
	t.Helper()
	for _, s := range shares {
		lo, hi := s.bounds(day)
		if p.slotHasStem(lo, hi) {
			t.Fatalf("day %d columns [%d,%d) have a stem rune, want none", day+1, lo+1, hi+1)
		}
	}
}

func assertBlankDay(t *testing.T, p onlinePlot, shares []dayShare, day int) {
	t.Helper()
	for _, s := range shares {
		lo, hi := s.bounds(day)
		if !p.slotHasBraille(lo, hi) || p.slotHasDiamond(lo, hi) || p.slotHasStem(lo, hi) {
			t.Fatalf("day %d columns [%d,%d) braille %v diamond %v stem %v, want braille only",
				day+1, lo+1, hi+1, p.slotHasBraille(lo, hi), p.slotHasDiamond(lo, hi), p.slotHasStem(lo, hi))
		}
	}
}

// assertFullStemBar rejects the library half-tips. A stem that joins
// the diamond to the trend is U+2502 on every cell of the segment.
func assertFullStemBar(t *testing.T, p onlinePlot, shares []dayShare, day int) {
	t.Helper()
	for _, s := range shares {
		lo, hi := s.bounds(day)
		stems := p.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
		if len(stems) == 0 {
			t.Fatalf("day %d has no stem bar", day+1)
		}
		for _, c := range stems {
			if c.r != '\u2502' {
				t.Fatalf("day %d stem %q, want U+2502", day+1, string(c.r))
			}
		}
	}
}

// assertStemReachesTrend requires that bar on the trend row in the
// diamond's column. A one-row gap has no row between the two marks,
// so the between-row check cannot see it.
func assertStemReachesTrend(t *testing.T, p onlinePlot, shares []dayShare, day int) {
	t.Helper()
	for _, s := range shares {
		lo, hi := s.bounds(day)
		dRow, dCol := -1, -1
		for r, row := range p.rows {
			for c := lo; c < hi && c < len(row); c++ {
				if row[c].r == openDiamond {
					dRow, dCol = r, c
				}
			}
		}
		if dCol < 0 {
			t.Fatalf("day %d has no diamond", day+1)
		}
		trendRow := -1
		for r, row := range p.rows {
			for c := lo; c < hi && c < len(row); c++ {
				if c == dCol || !isBraille(row[c].r) {
					continue
				}
				if trendRow >= 0 && trendRow != r {
					t.Fatalf("day %d braille on rows %d and %d", day+1, trendRow, r)
				}
				trendRow = r
			}
		}
		if trendRow < 0 {
			t.Fatalf("day %d has no trend beside the diamond", day+1)
		}
		gap := dRow - trendRow
		if gap < 0 {
			gap = -gap
		}
		if gap != 1 {
			t.Fatalf("day %d diamond row %d trend row %d, want adjacent rows", day+1, dRow, trendRow)
		}
		if got := p.rows[trendRow][dCol].r; got != '\u2502' {
			t.Fatalf("day %d trend cell %q, want U+2502", day+1, string(got))
		}
	}
}

func assertStemBetween(t *testing.T, p onlinePlot, shares []dayShare, day int) {
	t.Helper()
	for _, s := range shares {
		lo, hi := s.bounds(day)
		if !p.stemBetween(lo, hi) {
			t.Fatalf("day %d columns [%d,%d) have no stem row between the diamond and the trend", day+1, lo+1, hi+1)
		}
		for _, c := range p.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) }) {
			if isBraille(c.r) || c.r == openDiamond || c.r == filledDiamond {
				t.Fatalf("day %d stem rune %q, want neither braille nor a diamond", day+1, string(c.r))
			}
		}
	}
}

func assertTickLabels(t *testing.T, view string, topToBottom []string) onlinePlot {
	t.Helper()
	p := mustOnlinePlot(t, view)
	if p.ticksShareRow() {
		t.Fatalf("tick rows %v, two labels share a row", p.tickRows)
	}
	if strings.Join(p.tickLabels, ",") != strings.Join(topToBottom, ",") {
		t.Fatalf("ticks top to bottom = %v, want %v", p.tickLabels, topToBottom)
	}
	return p
}

func assertGutter(t *testing.T, p onlinePlot, label, want string) {
	t.Helper()
	for i, lab := range p.tickLabels {
		if lab != label {
			continue
		}
		if p.tickGutters[i] != want {
			t.Fatalf("gutter for %s = %q, want %q", label, p.tickGutters[i], want)
		}
		return
	}
	t.Fatalf("tick %s missing from %v", label, p.tickLabels)
}

func assertNoFilledDiamond(t *testing.T, view string) {
	t.Helper()
	if strings.ContainsRune(visible(view), filledDiamond) {
		t.Fatal("view has a filled diamond")
	}
}

// plotLabelCol is the plot column where text starts on the day-label
// line. The five gutter columns are not plot columns.
func plotLabelCol(line, text string) int {
	if len(line) < onlineGutterCols {
		return -1
	}
	return strings.Index(line[onlineGutterCols:], text)
}

func dayInts(line string) []int {
	var out []int
	for i := 0; i < len(line); {
		if line[i] < '0' || line[i] > '9' {
			i++
			continue
		}
		n := 0
		for i < len(line) && line[i] >= '0' && line[i] <= '9' {
			n = n*10 + int(line[i]-'0')
			i++
		}
		out = append(out, n)
	}
	return out
}

func hasDay(days []int, want int) bool {
	for _, d := range days {
		if d == want {
			return true
		}
	}
	return false
}

func dayLabelLine(view string) string {
	lines := strings.Split(view, "\n")
	last := -1
	for i, line := range lines {
		if _, ok := onlineScaleLine(line); ok {
			last = i
		}
	}
	if last < 0 {
		return ""
	}
	for _, line := range lines[last+1:] {
		vis := strings.TrimSpace(visible(line))
		if vis == "" {
			continue
		}
		low := strings.ToLower(vis)
		if strings.Contains(low, "loss") || strings.Contains(low, "deficit") || strings.Contains(low, "quit") || strings.Contains(low, "error") || strings.Contains(low, "db:") {
			continue
		}
		return visible(line)
	}
	return ""
}

func captionLine(view string) string {
	for _, line := range strings.Split(visible(view), "\n") {
		low := strings.ToLower(line)
		if strings.Contains(low, "loss") && strings.Contains(low, "cal") {
			return line
		}
	}
	return ""
}

func assertCaption(t *testing.T, view, want string) {
	t.Helper()
	if got := strings.TrimSpace(captionLine(view)); got != want {
		t.Fatalf("caption %q, want %q", got, want)
	}
}

func headingRaw(view, label string) string {
	for _, line := range strings.Split(view, "\n") {
		vis := visible(line)
		if strings.Contains(vis, label) && !strings.Contains(vis, "hdtools") && !strings.Contains(vis, "db:") {
			return line
		}
	}
	return ""
}

func hasBackgroundSGR(s string) bool {
	for i := 0; i < len(s); {
		if s[i] != 0x1b || i+1 >= len(s) || s[i+1] != '[' {
			i++
			continue
		}
		j := i + 2
		for j < len(s) && s[j] != 'm' {
			j++
		}
		if j >= len(s) {
			return false
		}
		for _, p := range strings.Split(s[i+2:j], ";") {
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil {
				continue
			}
			if n == 48 || (n >= 40 && n <= 47) || (n >= 100 && n <= 107) {
				return true
			}
		}
		i = j + 1
	}
	return false
}

func headingHasBorder(raw string) bool {
	for _, c := range columnCells(raw) {
		if c.r >= 0x2500 && c.r <= 0x257F {
			return true
		}
	}
	return false
}

func headingTextBold(raw, label string) bool {
	cells := columnCells(raw)
	vis := visible(raw)
	start := strings.Index(vis, label)
	if start < 0 {
		return false
	}
	cell := lipgloss.Width(vis[:start])
	width := lipgloss.Width(label)
	if cell < 0 || cell+width > len(cells) {
		return false
	}
	for i := cell; i < cell+width; i++ {
		if !cells[i].bold {
			return false
		}
	}
	return true
}

func yellowOnBlue(t *testing.T, raw string) bool {
	t.Helper()
	written := writtenANSI(t, raw)
	return hasIndexedForeground(written, 3) && (hasSGRCode(raw, 44) || hasSGRCode(written, 44))
}

// assertHeadingShape is the plain month-year line: bold, no box, centered
// on the 67-column plot line. Colour is the caller's job, because
// NO_COLOR keeps the bold and drops the chroma.
func assertHeadingShape(t *testing.T, view, label string) string {
	t.Helper()
	raw := headingRaw(view, label)
	if raw == "" {
		t.Fatalf("heading %q missing: %q", label, visible(view))
	}
	if strings.Contains(visible(raw), "hdtools") {
		t.Fatalf("heading says hdtools: %q", visible(raw))
	}
	if yellowOnBlue(t, raw) {
		t.Fatal("heading is still yellow on blue")
	}
	if headingHasBorder(raw) {
		t.Fatalf("heading still has a border: %q", visible(raw))
	}
	if hasBackgroundSGR(raw) {
		t.Fatal("heading still has a filled background")
	}
	if !headingTextBold(raw, label) {
		t.Fatal("heading is not bold")
	}
	vis := visible(raw)
	if lipgloss.Width(vis) != onlineLineCols {
		t.Fatalf("heading width = %d, want %d visible columns", lipgloss.Width(vis), onlineLineCols)
	}
	idx := strings.Index(vis, label)
	left := lipgloss.Width(vis[:idx])
	right := lipgloss.Width(vis) - left - lipgloss.Width(label)
	if left <= 0 || right <= 0 || absInt(left-right) > 1 {
		t.Fatalf("heading pads %d and %d, want the words centered with space on both sides", left, right)
	}
	return raw
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func assertDiamondFG(t *testing.T, cells []cellStyle, index int) {
	t.Helper()
	if len(cells) == 0 {
		t.Fatal("diamond cell missing")
	}
	for _, c := range cells {
		if c.r != openDiamond {
			t.Fatalf("diamond rune %q, want U+25C7", string(c.r))
		}
		if !cellHasFG(c, index) {
			t.Fatalf("diamond seq %q, want colour index %d", c.seq, index)
		}
		if index != 2 && cellHasFG(c, 2) {
			t.Fatalf("diamond seq %q is green, want the weight colour", c.seq)
		}
	}
}

func assertCellsFG(t *testing.T, cells []cellStyle, index int, what string) {
	t.Helper()
	if len(cells) == 0 {
		t.Fatalf("no %s cell", what)
	}
	for _, c := range cells {
		if !cellHasFG(c, index) {
			t.Fatalf("%s seq %q, want colour index %d", what, c.seq, index)
		}
	}
}

func novemberPlot(t *testing.T, today time.Time, cfg config.Config, points ...logPoint) (string, onlinePlot) {
	t.Helper()
	app := showOnline(t, today, cfg, points...)
	view := monthlyView(t, app)
	return view, mustOnlinePlot(t, view)
}

func TestOnlineDiamondAndTrend(t *testing.T) {
	view, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82))
	assertNoFilledDiamond(t, view)
	cols := plot.diamondCols()
	if len(cols) != 2 || cols[0] >= cols[1] {
		t.Fatalf("diamond columns %v, want two open diamonds, day 1 then day 30", cols)
	}
	if !plot.slotHasBraille(cols[0]+1, cols[1]) {
		t.Fatal("no braille trend on the days between the diamonds")
	}
}

func TestOnlineBlankDayHasTrendNoDiamond(t *testing.T) {
	_, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82))
	shares := bestDayShares(t, plot, 30, []int{0, 29})
	assertBlankDay(t, plot, shares, 1)
}

func TestOnlineStemFloatAndSinker(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		enableChroma(t)
		view, _ := novemberPlot(t, dec151990(), config.Default(),
			kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 100))
		plot := mustOnlinePlot(t, writtenANSI(t, view))
		shares := bestDayShares(t, plot, 30, []int{0, 1})
		assertNoStem(t, plot, shares, 0)
		assertStemBetween(t, plot, shares, 1)
		assertFullStemBar(t, plot, shares, 1)
		for _, s := range shares {
			lo, hi := s.bounds(1)
			assertDiamondFG(t, plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }), 4)
			stems := plot.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
			assertCellsFG(t, stems, 2, "float stem")
			for _, c := range stems {
				if isBraille(c.r) {
					t.Fatalf("float stem %q is braille", string(c.r))
				}
			}
		}
	})
	t.Run("sinker", func(t *testing.T) {
		enableChroma(t)
		view, _ := novemberPlot(t, dec151990(), config.Default(),
			kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 70))
		plot := mustOnlinePlot(t, writtenANSI(t, view))
		shares := bestDayShares(t, plot, 30, []int{0, 1})
		assertStemBetween(t, plot, shares, 1)
		assertFullStemBar(t, plot, shares, 1)
		for _, s := range shares {
			lo, hi := s.bounds(1)
			assertDiamondFG(t, plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }), 4)
			stems := plot.cellsWhere(lo, hi, func(c cellStyle) bool { return isStemRune(c.r) })
			assertCellsFG(t, stems, 2, "sinker stem")
		}
	})
}

func TestOnlineStemJoinsAdjacentRows(t *testing.T) {
	// 80 then 80.1 kg. Day 2's trend stays 80.0. On the 79 to 81.1
	// canvas those two values round to neighboring rows, so the stem
	// has no row between the diamond and the trend.
	_, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 80.1))
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	assertNoStem(t, plot, shares, 0)
	assertStemReachesTrend(t, plot, shares, 1)
}

func TestOnlineWorkoutSameDiamond(t *testing.T) {
	enableChroma(t)
	view, plot := novemberPlot(t, dec151990(), config.Default(),
		logPoint{day: ymd(1990, time.November, 1), kg: 80, workout: true},
		kgPoint(time.November, 2, 80))
	ansi := mustOnlinePlot(t, writtenANSI(t, view))
	assertNoFilledDiamond(t, view)
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	for _, day := range []int{0, 1} {
		for _, s := range shares {
			lo, hi := s.bounds(day)
			extra := plot.cellsWhere(lo, hi, func(c cellStyle) bool {
				return c.r != ' ' && c.r != 0 && !isBraille(c.r) && c.r != openDiamond
			})
			if len(extra) != 0 {
				t.Fatalf("day %d has an extra mark %q", day+1, string(extra[0].r))
			}
			assertDiamondFG(t, ansi.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }), 4)
		}
	}
}

func TestOnlinePlainHeading(t *testing.T) {
	points := []logPoint{kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82)}
	t.Run("default", func(t *testing.T) {
		enableChroma(t)
		view, _ := novemberPlot(t, dec151990(), config.Default(), points...)
		raw := assertHeadingShape(t, view, "November 1990")
		if !hasIndexedForeground(writtenANSI(t, raw), 6) {
			t.Fatal("heading is not the default cyan title role")
		}
	})
	t.Run("magenta", func(t *testing.T) {
		enableChroma(t)
		cfg := schemeCfg(map[string]string{"title": "magenta"})
		view, plot := novemberPlot(t, dec151990(), cfg, points...)
		raw := assertHeadingShape(t, view, "November 1990")
		if !hasIndexedForeground(writtenANSI(t, raw), 5) {
			t.Fatal("title magenta did not change the heading")
		}
		ansi := mustOnlinePlot(t, writtenANSI(t, view))
		shares := bestDayShares(t, plot, 30, []int{0, 29})
		for _, s := range shares {
			lo, hi := s.bounds(2)
			assertCellsFG(t, ansi.cellsWhere(lo, hi, func(c cellStyle) bool { return isBraille(c.r) }), 1, "blank-day trend")
		}
	})
	t.Run("nocolor", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		view, _ := novemberPlot(t, dec151990(), config.Default(), points...)
		if !strings.Contains(visible(view), "November 1990") {
			t.Fatalf("month missing under NO_COLOR: %q", visible(view))
		}
		if hasChromaticSGR(view) {
			t.Fatal("chromatic SGR under NO_COLOR")
		}
		raw := headingRaw(view, "November 1990")
		if raw == "" || !headingTextBold(raw, "November 1990") {
			t.Fatal("NO_COLOR dropped the heading bold")
		}
	})
}

func TestOnlineAxisKilograms(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82))
	plot := assertTickLabels(t, view, []string{"83", "82", "81", "80", "79"})
	assertGutter(t, plot, "79", "  79-")
}

func TestOnlineAxisFlat(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(), kgPoint(time.November, 1, 80))
	plot := assertTickLabels(t, view, []string{"81", "80", "79"})
	for _, lab := range plot.tickLabels {
		if strings.Contains(lab, ".") {
			t.Fatalf("tick %q, want 79, 80, 81 and not a ±2 lb band", lab)
		}
	}
}

func TestOnlineAxisOffGrid(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80.4), kgPoint(time.November, 30, 82.4))
	plot := assertTickLabels(t, view, []string{"83", "82", "81", "80"})
	for _, lab := range plot.tickLabels {
		if lab == "79" || lab == "84" {
			t.Fatalf("tick %s is outside 80 through 83", lab)
		}
	}
}

func TestOnlineAxisPounds(t *testing.T) {
	lo, err := units.ToKG(170, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	hi, err := units.ToKG(175, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	view, _ := novemberPlot(t, dec151990(), cfg,
		logPoint{day: ymd(1990, time.November, 1), kg: lo},
		logPoint{day: ymd(1990, time.November, 30), kg: hi})
	plot := assertTickLabels(t, view, []string{"176", "175", "174", "173", "172", "171", "170", "169"})
	for _, lab := range plot.tickLabels {
		if strings.Contains(lab, ".") {
			t.Fatalf("pound tick %q has a decimal", lab)
		}
	}
}

func TestOnlineAxisStone(t *testing.T) {
	lo, err := units.ToKG(12.0, units.Stone)
	if err != nil {
		t.Fatal(err)
	}
	hi, err := units.ToKG(12.5, units.Stone)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.DisplayUnit = units.Stone
	view, _ := novemberPlot(t, dec151990(), cfg,
		logPoint{day: ymd(1990, time.November, 1), kg: lo},
		logPoint{day: ymd(1990, time.November, 30), kg: hi})
	// Bottom to top is 11.9, 12.0, 12.1, 12.1, 12.2, 12.3, 12.4, 12.4, 12.5, 12.6.
	plot := assertTickLabels(t, view, []string{
		"12.6", "12.5", "12.4", "12.4", "12.3", "12.2", "12.1", "12.1", "12.0", "11.9",
	})
	assertGutter(t, plot, "12.6", "12.6-")
	for _, lab := range plot.tickLabels {
		if strings.Contains(lab, "st") || strings.Contains(lab, "lb") {
			t.Fatalf("tick %q is a stones-and-pounds pair", lab)
		}
	}
}

func TestOnlineStoneCanvasIsOnePoundPastData(t *testing.T) {
	// 80 kg is about 176.37 lb, not a whole pound. The canvas is one
	// pound past that value. Snapping onto the 176 and 177 ticks
	// leaves less than a pound of room.
	v, err := units.FromKG(80, units.Stone)
	if err != nil {
		t.Fatal(err)
	}
	lo, hi, ticks := stoneAxis(v, v)
	minLb := v * lbPerStone
	wantLo := (minLb - 1) / lbPerStone
	wantHi := (minLb + 1) / lbPerStone
	const eps = 1e-9
	if lo < wantLo-eps || lo > wantLo+eps || hi < wantHi-eps || hi > wantHi+eps {
		t.Fatalf("canvas [%v, %v], want one pound past the data [%v, %v]", lo, hi, wantLo, wantHi)
	}
	if len(ticks) != 2 || ticks[0].label != "12.6" || ticks[1].label != "12.6" {
		t.Fatalf("ticks = %+v, want the pounds inside that scale, both labeled 12.6", ticks)
	}
}

func TestOnlineCrowdedEndTicks(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 40), kgPoint(time.November, 30, 120))
	plot := mustOnlinePlot(t, view)
	if plot.ticksShareRow() {
		t.Fatal("two tick labels share a row")
	}
	labels := plot.tickLabels
	if len(labels) < 2 || labels[0] != "121" || labels[len(labels)-1] != "39" {
		t.Fatalf("end ticks = %v, want top 121 and bottom 39", labels)
	}
	assertGutter(t, plot, "121", " 121-")
	for _, lab := range labels {
		if lab == "38" || lab == "122" {
			t.Fatalf("tick %s is outside one step beyond 40..120", lab)
		}
	}
	// 39 through 121 is 83 ticks. Sixteen rows cannot show them all.
	if len(labels) >= 121-39+1 {
		t.Fatal("every middle tick is shown; the 16-row plot must omit at least one")
	}
}

func TestOnlineDayLabels(t *testing.T) {
	view, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82))
	days := dayInts(dayLabelLine(view))
	if !hasDay(days, 1) || !hasDay(days, 30) {
		t.Fatalf("day labels %v, want 1 and 30", days)
	}
	if len(plot.rows) == 0 || len(plot.rows[0]) != onlinePlotCols {
		t.Fatalf("plot columns = %d, want 62", len(plot.rows[0]))
	}
	// The first and last day slots reach the ends of the 62 columns.
	if !plot.ink(0, 3) || !plot.ink(onlinePlotCols-3, onlinePlotCols) {
		t.Fatal("day slots do not use the full 62 plot columns")
	}
}

func TestOnlineCurrentMonthEndsToday(t *testing.T) {
	today := ymd(1990, time.November, 10)
	view, plot := novemberPlot(t, today, config.Default(),
		kgPoint(time.November, 1, 80),
		kgPoint(time.November, 2, 70),
		kgPoint(time.November, 15, 60))
	if cols := plot.diamondCols(); len(cols) != 2 {
		t.Fatalf("diamonds = %d, want two (day 15 is after today)", len(cols))
	}
	days := dayInts(dayLabelLine(view))
	if !hasDay(days, 1) || !hasDay(days, 10) {
		t.Fatalf("day labels %v, want 1 and 10", days)
	}
	if hasDay(days, 15) || hasDay(days, 30) {
		t.Fatalf("day labels %v include a day that is not plotted", days)
	}
	// A 30-day month gives day 10 columns 18 through 21 at most.
	// The two digits must sit in that share, not on the empty tail.
	col := plotLabelCol(dayLabelLine(view), "10")
	if col < 18 || col > 20 {
		t.Fatalf("day 10 label column %d, want it under that day", col)
	}
	// Ten days of a 30-day month take at most 22 columns when both
	// remainder columns fall on days 1 through 10.
	if !plot.columnsBlank(22, onlinePlotCols) {
		t.Fatal("plot columns 23 through 62 are not empty")
	}
	assertCaption(t, view, "Monthly loss: 0.0 kg   Daily deficit: 0 cal")
}

func TestOnlineDayLabelsSkipUnplottedStart(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 5, 80), kgPoint(time.November, 20, 82))
	line := dayLabelLine(view)
	days := dayInts(line)
	if !hasDay(days, 5) || !hasDay(days, 30) {
		t.Fatalf("day labels %v, want first plotted day 5 and last plotted day 30", days)
	}
	if hasDay(days, 1) {
		t.Fatalf("day labels %v include unplotted day 1", days)
	}
	// Day 5 starts in columns 8 through 10. Column 0 is day 1.
	col := plotLabelCol(line, "5")
	if col < 8 || col > 10 {
		t.Fatalf("day 5 label column %d, want it under that day", col)
	}
}

func TestOnlineFebruaryAnd31DayPlotWidth(t *testing.T) {
	app := showOnline(t, ymd(1990, time.March, 15), config.Default(),
		kgPoint(time.January, 1, 80),
		kgPoint(time.January, 31, 80),
		kgPoint(time.February, 1, 80))
	t.Run("february", func(t *testing.T) {
		view := app.View().Content
		if !strings.Contains(visible(view), "February 1990") {
			t.Fatalf("february chart missing: %q", visible(view))
		}
		plot := mustOnlinePlot(t, view)
		if len(plot.rows[0]) != onlinePlotCols {
			t.Fatalf("february plot columns = %d, want 62", len(plot.rows[0]))
		}
	})
	t.Run("january", func(t *testing.T) {
		press(app, "[")
		view := app.View().Content
		if !strings.Contains(visible(view), "January 1990") {
			t.Fatalf("january chart missing: %q", visible(view))
		}
		plot := mustOnlinePlot(t, view)
		cols := plot.diamondCols()
		if len(cols) != 2 || cols[0] > 1 || cols[1] < 60 || cols[1] != cols[0]+60 {
			got := make([]int, len(cols))
			for i, c := range cols {
				got[i] = c + 1
			}
			t.Fatalf("diamonds on plot columns %v, want day 1 in 1-2 and day 31 sixty columns later", got)
		}
	})
}

func TestOnlinePlotHeight(t *testing.T) {
	t.Run("kilograms", func(t *testing.T) {
		view, _ := novemberPlot(t, dec151990(), config.Default(),
			kgPoint(time.November, 1, 80), kgPoint(time.November, 30, 82))
		assertTickLabels(t, view, []string{"83", "82", "81", "80", "79"})
	})
	t.Run("stone", func(t *testing.T) {
		lo, err := units.ToKG(12.0, units.Stone)
		if err != nil {
			t.Fatal(err)
		}
		hi, err := units.ToKG(12.5, units.Stone)
		if err != nil {
			t.Fatal(err)
		}
		cfg := config.Default()
		cfg.DisplayUnit = units.Stone
		view, _ := novemberPlot(t, dec151990(), cfg,
			logPoint{day: ymd(1990, time.November, 1), kg: lo},
			logPoint{day: ymd(1990, time.November, 30), kg: hi})
		assertTickLabels(t, view, []string{
			"12.6", "12.5", "12.4", "12.4", "12.3", "12.2", "12.1", "12.1", "12.0", "11.9",
		})
	})
	t.Run("crowded", func(t *testing.T) {
		view, _ := novemberPlot(t, dec151990(), config.Default(),
			kgPoint(time.November, 1, 40), kgPoint(time.November, 30, 120))
		plot := mustOnlinePlot(t, view)
		labels := plot.tickLabels
		if len(labels) < 2 || labels[0] != "121" || labels[len(labels)-1] != "39" {
			t.Fatalf("end ticks = %v, want 121 and 39", labels)
		}
		if len(labels) >= 121-39+1 {
			t.Fatal("wide range showed every middle tick")
		}
	})
	t.Run("carry", func(t *testing.T) {
		app := loadOnline(t, dec151990(), config.Default(), kgPoint(time.October, 31, 80))
		press(app, "m", "]", "c")
		view := monthlyView(t, app)
		plot := mustOnlinePlot(t, view)
		if !plot.hasBraille() {
			t.Fatal("carry-only plot has no braille trend")
		}
		if len(plot.diamondCols()) != 0 {
			t.Fatal("carry-only plot has a diamond")
		}
		assertCaption(t, view, "Monthly loss: 0.0 kg   Daily deficit: 0 cal")
	})
	t.Run("empty", func(t *testing.T) {
		app := loadOnline(t, dec151990(), config.Default())
		press(app, "c")
		view := app.View().Content
		if !strings.Contains(strings.ToLower(visible(view)), "empty") {
			t.Fatalf("empty state missing: %q", visible(view))
		}
		if n := countOnlinePlotRows(view); n != 0 {
			t.Fatalf("empty month has %d plot rows, want no 16-row plot", n)
		}
	})
}

func TestOnlineCarryFromPreviousMonth(t *testing.T) {
	view, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.October, 31, 90), kgPoint(time.November, 1, 80))
	shares := bestDayShares(t, plot, 30, []int{0})
	for _, s := range shares {
		lo, hi := s.bounds(0)
		if !plot.slotHasStem(lo, hi) || !plot.slotHasDiamond(lo, hi) {
			t.Fatal("1 November has no diamond and stem")
		}
	}
	assertBlankDay(t, plot, shares, 2)
	labels := plot.tickLabels
	if len(labels) < 2 || labels[0] != "90" || labels[len(labels)-1] != "79" {
		t.Fatalf("end ticks = %v, want top 90 and bottom 79", labels)
	}
	for _, lab := range labels {
		if lab == "91" {
			t.Fatal("91 is an axis label; October's weight was plotted")
		}
	}
	if strings.Contains(visible(view), "October") {
		t.Fatalf("October is plotted on the November chart: %q", visible(view))
	}
}

func TestOnlineTrendNotRawFit(t *testing.T) {
	_, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 100))
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	assertNoStem(t, plot, shares, 0)
	for _, s := range shares {
		lo, hi := s.bounds(1)
		if !plot.slotHasStem(lo, hi) || !plot.slotHasDiamond(lo, hi) {
			t.Fatal("day 2 has no diamond and stem")
		}
	}
	labels := plot.tickLabels
	if len(labels) < 2 || labels[0] != "101" || labels[len(labels)-1] != "79" {
		t.Fatalf("end ticks = %v, want top 101 and bottom 79", labels)
	}
}

func colouredNovember(t *testing.T, cfg config.Config) onlinePlot {
	t.Helper()
	view, _ := novemberPlot(t, dec151990(), cfg,
		kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 100))
	return mustOnlinePlot(t, writtenANSI(t, view))
}

func assertSeriesColours(t *testing.T, plot onlinePlot, diamond, trend, stem int) {
	t.Helper()
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	for _, s := range shares {
		dLo, dHi := s.bounds(1)
		bLo, bHi := s.bounds(2)
		assertDiamondFG(t, plot.cellsWhere(dLo, dHi, func(c cellStyle) bool { return c.r == openDiamond }), diamond)
		assertCellsFG(t, plot.cellsWhere(bLo, bHi, func(c cellStyle) bool { return isBraille(c.r) }), trend, "blank-day trend")
		assertCellsFG(t, plot.cellsWhere(dLo, dHi, func(c cellStyle) bool { return isStemRune(c.r) }), stem, "stem")
	}
}

func TestOnlineDefaultColours(t *testing.T) {
	enableChroma(t)
	assertSeriesColours(t, colouredNovember(t, config.Default()), 4, 1, 2)
}

func TestOnlineColourOverrides(t *testing.T) {
	enableChroma(t)
	cfg := schemeCfg(map[string]string{"weight": "magenta", "trend": "yellow"})
	assertSeriesColours(t, colouredNovember(t, cfg), 5, 3, 2)
}

func TestOnlineIgnoredColourKeys(t *testing.T) {
	enableChroma(t)
	cfg := loadConfigTOML(t, `
display_unit = "kg"

[colors]
diamond = "yellow"
stem = "magenta"
trend = "not-a-colour"
`)
	assertSeriesColours(t, colouredNovember(t, cfg), 4, 1, 2)
}

func TestOnlineNoColorKeepsShapes(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	view, plot := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 100))
	if hasChromaticSGR(view) {
		t.Fatal("chromatic SGR under NO_COLOR")
	}
	if len(plot.diamondCols()) == 0 || !plot.hasBraille() {
		t.Fatal("NO_COLOR dropped the diamond or the braille trend")
	}
	shares := bestDayShares(t, plot, 30, []int{0, 1})
	for _, s := range shares {
		lo, hi := s.bounds(1)
		if !plot.slotHasStem(lo, hi) {
			t.Fatal("NO_COLOR dropped the stem")
		}
		for _, c := range plot.cellsWhere(lo, hi, func(c cellStyle) bool { return c.r == openDiamond }) {
			if c.bold {
				t.Fatal("diamond is bold")
			}
		}
		bLo, bHi := s.bounds(2)
		blank := plot.cellsWhere(bLo, bHi, func(c cellStyle) bool { return isBraille(c.r) })
		if len(blank) == 0 {
			t.Fatal("blank day has no braille trend")
		}
		for _, c := range blank {
			if !c.bold {
				t.Fatal("blank-day trend is not bold")
			}
		}
	}
}

func TestOnlineHexDownshift(t *testing.T) {
	enableChroma(t)
	cfg := schemeCfg(map[string]string{"trend": "#123456"})
	view, _ := novemberPlot(t, dec151990(), cfg,
		kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 100))
	truePlot := mustOnlinePlot(t, writtenTrueColor(t, view))
	shares := bestDayShares(t, truePlot, 30, []int{0, 1})
	found := false
	for _, s := range shares {
		lo, hi := s.bounds(2)
		for _, c := range truePlot.cellsWhere(lo, hi, func(c cellStyle) bool { return isBraille(c.r) }) {
			if strings.Contains(c.seq, "38;2;18;52;86") {
				found = true
			}
		}
	}
	if !found {
		t.Fatal(`truecolor write of the blank-day trend lacks 38;2;18;52;86`)
	}
	if strings.Contains(writtenANSI(t, view), "38;2") {
		t.Fatal("16-color write still contains 38;2")
	}
}

func TestOnlineFixtureC(t *testing.T) {
	points := []logPoint{kgPoint(time.November, 1, 80), kgPoint(time.November, 2, 70)}
	t.Run("kilograms", func(t *testing.T) {
		view, _ := novemberPlot(t, dec151990(), config.Default(), points...)
		assertCaption(t, view, "Monthly loss: 0.2 kg   Daily deficit: 50 cal")
	})
	t.Run("pounds", func(t *testing.T) {
		cfg := config.Default()
		cfg.DisplayUnit = units.Pound
		view, _ := novemberPlot(t, dec151990(), cfg, points...)
		assertCaption(t, view, "Monthly loss: 0.4 lb   Daily deficit: 50 cal")
		if strings.Contains(captionLine(view), " kg") {
			t.Fatalf("caption still in kg: %q", captionLine(view))
		}
	})
}

func TestOnlineLongTermKeepsPad(t *testing.T) {
	lo, err := units.ToKG(170, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	hi, err := units.ToKG(175, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.DisplayUnit = units.Pound
	app := showOnline(t, dec151990(), cfg,
		logPoint{day: ymd(1990, time.November, 1), kg: lo},
		logPoint{day: ymd(1990, time.November, 2), kg: hi})
	monthly := monthlyView(t, app)
	assertTickLabels(t, monthly, []string{"176", "175", "174", "173", "172", "171", "170", "169"})
	press(app, "l")
	view := app.View().Content
	vis := visible(view)
	if strings.ContainsRune(vis, openDiamond) || strings.ContainsRune(vis, filledDiamond) {
		t.Fatal("long-term chart has a diamond")
	}
	rows := mustPlot(t, view)
	for _, row := range rows {
		for _, c := range row {
			if c.r == ' ' || c.r == 0 {
				continue
			}
			if !isBraille(c.r) {
				t.Fatalf("long-term rune %q, want braille and no stem", string(c.r))
			}
		}
	}
	ymin, ymax, ok := chartYBounds(view)
	if !ok || ymin < 167.9 || ymin > 168.2 || ymax < 176.8 || ymax > 177.2 {
		t.Fatalf("long-term Y = [%.2f, %.2f], want about [168, 177]", ymin, ymax)
	}
	// Headless tests leave the long-term width at its 72-column default.
	if w := plotWidth(rows); w == onlinePlotCols || w == onlineLineCols || w != 72 {
		t.Fatalf("long-term width = %d, want the current 72 columns, not 62 or 67", w)
	}
	for _, line := range strings.Split(view, "\n") {
		if _, ok := onlineScaleLine(line); ok {
			t.Fatal("long-term row uses the monthly 67-column gutter")
		}
	}
}

func TestOnlineFixtureA(t *testing.T) {
	view, _ := novemberPlot(t, ymd(1990, time.November, 10), config.Default(),
		kgPoint(time.November, 1, 80),
		kgPoint(time.November, 2, 70))
	assertCaption(t, view, "Monthly loss: 0.0 kg   Daily deficit: 0 cal")
}

func TestOnlineFixtureB(t *testing.T) {
	points := []logPoint{
		kgPoint(time.November, 8, 80),
		kgPoint(time.November, 9, 70),
	}
	today := ymd(1990, time.November, 10)
	t.Run("kilograms", func(t *testing.T) {
		view, _ := novemberPlot(t, today, config.Default(), points...)
		assertCaption(t, view, "Monthly loss: 1.0 kg   Daily deficit: 3858 cal")
	})
	t.Run("pounds", func(t *testing.T) {
		cfg := config.Default()
		cfg.DisplayUnit = units.Pound
		view, _ := novemberPlot(t, today, cfg, points...)
		assertCaption(t, view, "Monthly loss: 2.2 lb   Daily deficit: 3858 cal")
		if strings.Contains(captionLine(view), " kg") {
			t.Fatalf("caption still in kg: %q", captionLine(view))
		}
	})
}

func TestOnlineGainIsNegative(t *testing.T) {
	view, _ := novemberPlot(t, ymd(1990, time.November, 10), config.Default(),
		kgPoint(time.November, 8, 70),
		kgPoint(time.November, 9, 80))
	assertCaption(t, view, "Monthly loss: -1.0 kg   Daily deficit: -3858 cal")
}

func TestOnlineCarryFeedsRateOnly(t *testing.T) {
	app := showOnline(t, ymd(1990, time.November, 4), config.Default(),
		kgPoint(time.October, 30, 80),
		kgPoint(time.October, 31, 70))
	press(app, "]")
	view := monthlyView(t, app)
	plot := mustOnlinePlot(t, view)
	if len(plot.diamondCols()) != 0 {
		t.Fatal("November has a diamond")
	}
	assertCaption(t, view, "Monthly loss: 1.0 kg   Daily deficit: 1543 cal")
}

func TestOnlineFitWithoutFirstDay(t *testing.T) {
	view, _ := novemberPlot(t, dec151990(), config.Default(),
		kgPoint(time.November, 29, 80),
		kgPoint(time.November, 30, 70))
	assertCaption(t, view, "Monthly loss: 1.0 kg   Daily deficit: 7716 cal")
}

func TestOnlineCurrentMonthLateStart(t *testing.T) {
	view, _ := novemberPlot(t, ymd(1990, time.November, 30), config.Default(),
		kgPoint(time.November, 29, 80),
		kgPoint(time.November, 30, 70))
	assertCaption(t, view, "Monthly loss: 1.0 kg   Daily deficit: 7716 cal")
}

func TestOnlineOmitsShortSeries(t *testing.T) {
	view, _ := novemberPlot(t, ymd(1990, time.November, 1), config.Default(),
		kgPoint(time.November, 1, 80))
	vis := visible(view)
	if strings.Contains(vis, "Monthly loss:") || strings.Contains(vis, "Daily deficit:") {
		t.Fatalf("caption on one trend value: %q", vis)
	}
}

func TestOnlineUnroundedPounds(t *testing.T) {
	points := []logPoint{
		kgPoint(time.November, 1, 70),
		kgPoint(time.November, 15, 60),
	}
	t.Run("kilograms", func(t *testing.T) {
		view, _ := novemberPlot(t, dec151990(), config.Default(), points...)
		assertCaption(t, view, "Monthly loss: 1.4 kg   Daily deficit: 385 cal")
	})
	t.Run("pounds", func(t *testing.T) {
		cfg := config.Default()
		cfg.DisplayUnit = units.Pound
		view, _ := novemberPlot(t, dec151990(), cfg, points...)
		assertCaption(t, view, "Monthly loss: 3.2 lb   Daily deficit: 385 cal")
	})
}
