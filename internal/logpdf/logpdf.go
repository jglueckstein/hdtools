// Package logpdf writes one portrait US Letter page of a filled monthly
// log: a heading, a ruled table, and one row for every calendar day.
//
// The sheet is its own writer because the chart PDF is a landscape plot
// that stops at today, and this page is a portrait table that keeps the
// days after today. The numbers follow the month sheet (display unit,
// painted delta, carried trend) without importing the TUI, so a paper
// change cannot move the screen and a screen change cannot move the
// paper. The tests are what keep the two in step.
//
// Write does not open SQLite, resolve pdf_dir, read colours, call
// ApplyTrend, or clip at Today. A blank cell is omitted rather than
// drawn as a hyphen. Day 31 stays on this one page.
package logpdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

// Options is the month to paint. Logs are the full series and are
// already trended when a trend exists. Today is part of the call so a
// test can prove the page does not stop on that day. The painter does
// not read it: a zero Today drops no rows either.
type Options struct {
	Year  int
	Month time.Month
	Logs  []dailylog.DailyLog
	Unit  units.Unit
	Today time.Time
}

// Column widths sum to the 540 pt content width inside a 36 pt margin.
// Weekday is 36 pt because Helvetica 8 draws the header word "weekday"
// at 31.57 pt, which crosses a 28 pt rule into weight. Those 8 pt come
// from weight. A one-decimal weight still fits in 40 pt. The note's
// inner width stays that column minus 2 pt of pad on each side, which
// is the width the note tests measure before they cut.
const (
	margin    = 36.0
	rowPitch  = 18.0
	headingPt = 12.0
	bodyPt    = 8.0
	cellPad   = 2.0
	noteInner = 220.0 - 2*cellPad
)

var colWidth = [...]float64{28, 36, 40, 48, 44, 36, 52, 36, 220}

// Write paints opts onto one portrait Letter PDF at path, mode 0600.
// The bytes go to a same-directory temp, then Rename onto path, so a
// failed write leaves the previous file alone. os.Create on the
// destination would truncate it first and start from 0666.
func Write(path string, opts Options) error {
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("write log pdf %s: %w", path, err)
		}
	}

	pdf := fpdf.New("P", "pt", "Letter", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	if err := paint(pdf, opts); err != nil {
		return fmt.Errorf("write log pdf %s: %w", path, err)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return fmt.Errorf("write log pdf %s: %w", path, err)
	}
	if err := installOwnerOnly(path, dir, buf.Bytes()); err != nil {
		return fmt.Errorf("write log pdf %s: %w", path, err)
	}
	return nil
}

// paint keeps the heading, the header, and 31 day rows inside the
// media box. Pitch is fixed so day 31's baseline is still on this page.
// Today is intentionally unused: the chart clip does not apply here.
func paint(pdf *fpdf.Fpdf, opts Options) error {
	pdf.SetFont("Helvetica", "", headingPt)
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("select helvetica: %w", err)
	}
	title := fmt.Sprintf("%s %d (%s)", opts.Month, opts.Year, opts.Unit)
	headingY := margin + headingPt
	pdf.Text(margin+cellPad, headingY, title)

	pdf.SetFont("Helvetica", "", bodyPt)
	headerY := headingY + rowPitch
	headers := []string{"day", "weekday", "weight", "trend", "delta", "sleep", "steps", "workout", "note"}
	for i, header := range headers {
		pdf.Text(colX(i), headerY, header)
	}
	hrule(pdf, headerY)

	byDay := indexLogs(opts.Logs)
	n := daysInMonth(opts.Year, opts.Month)
	for d := 1; d <= n; d++ {
		day := time.Date(opts.Year, opts.Month, d, 0, 0, 0, 0, time.UTC)
		y := headerY + float64(d)*rowPitch
		if err := paintDay(pdf, opts, byDay, day, y); err != nil {
			return fmt.Errorf("paint %s: %w", day.Format(time.DateOnly), err)
		}
		hrule(pdf, y)
	}

	// One stroke per boundary, including the outer edges, so the rules
	// are vector lines rather than a filled box or a character grid.
	top := headerY - bodyPt
	bot := headerY + float64(n)*rowPitch + bodyPt
	x := margin
	pdf.Line(x, top, x, bot)
	for _, w := range colWidth {
		x += w
		pdf.Line(x, top, x, bot)
	}
	return pdf.Error()
}

func hrule(pdf *fpdf.Fpdf, baseline float64) {
	y := baseline + bodyPt
	pdf.Line(margin, y, margin+contentWidth(), y)
}

func contentWidth() float64 {
	var w float64
	for _, c := range colWidth {
		w += c
	}
	return w
}

func colX(i int) float64 {
	x := margin
	for j := 0; j < i; j++ {
		x += colWidth[j]
	}
	return x + cellPad
}

// paintDay emits one text operation per filled cell, in column order.
// A blank cell is skipped so the page has no hyphen, em dash, or the
// word blank standing in for missing data.
func paintDay(pdf *fpdf.Fpdf, opts Options, byDay map[string]dailylog.DailyLog, day time.Time, y float64) error {
	log, hasRow := byDay[dayKey(day)]
	trend, hasTrend := trendForDay(opts.Logs, log, hasRow, day)

	var text [9]string
	text[0] = strconv.Itoa(day.Day())
	text[1] = day.Weekday().String()[:2]

	var weightPaint, trendPaint float64
	var haveWeight bool
	if hasRow && log.Weight != nil {
		s, v, err := painted(*log.Weight, opts.Unit)
		if err != nil {
			return fmt.Errorf("weight cell: %w", err)
		}
		text[2] = s
		weightPaint = v
		haveWeight = true
	}
	if hasTrend {
		s, v, err := painted(trend, opts.Unit)
		if err != nil {
			return fmt.Errorf("trend cell: %w", err)
		}
		text[3] = s
		trendPaint = v
	}
	if haveWeight && hasTrend {
		text[4] = deltaText(weightPaint, trendPaint)
	}
	if hasRow {
		text[5] = fmt.Sprintf("%.1f", log.SleepHours)
		text[6] = strconv.Itoa(log.Steps)
		text[7] = "no"
		if log.Workout {
			text[7] = "yes"
		}
		text[8] = fitNote(pdf, encodeNote(log.Note))
	}
	for i, cell := range text {
		if cell == "" {
			continue
		}
		pdf.Text(colX(i), y, cell)
	}
	return nil
}

// trendForDay matches the month sheet. A logged day uses its own trend
// only when that row has a weight or a non-zero trend. A day with no
// row uses the latest earlier log that passes the same test. A logged
// day that itself has no trend does not inherit one.
func trendForDay(logs []dailylog.DailyLog, log dailylog.DailyLog, hasRow bool, day time.Time) (float64, bool) {
	if hasRow {
		if log.Weight != nil || log.Trend != 0 {
			return log.Trend, true
		}
		return 0, false
	}
	return carriedTrend(logs, day)
}

func carriedTrend(logs []dailylog.DailyLog, day time.Time) (float64, bool) {
	day = utcDay(day)
	var trend float64
	found := false
	for _, log := range logs {
		if utcDay(log.Day).After(day) {
			break
		}
		if log.Weight != nil || log.Trend != 0 {
			trend = log.Trend
			found = true
		}
	}
	return trend, found
}

func indexLogs(logs []dailylog.DailyLog) map[string]dailylog.DailyLog {
	byDay := make(map[string]dailylog.DailyLog, len(logs))
	for _, log := range logs {
		byDay[dayKey(log.Day)] = log
	}
	return byDay
}

func dayKey(t time.Time) string {
	return utcDay(t).Format(time.DateOnly)
}

func utcDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// painted is the one-decimal cell the month sheet shows. Delta subtracts
// these parsed texts, not the kilogram residual, so a pound display
// cannot disagree with the two numbers beside it.
func painted(kg float64, unit units.Unit) (string, float64, error) {
	v, err := units.FromKG(kg, unit)
	if err != nil {
		return "", 0, fmt.Errorf("display %s: %w", unit, err)
	}
	s := fmt.Sprintf("%.1f", v)
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", 0, fmt.Errorf("parse painted %q: %w", s, err)
	}
	return s, n, nil
}

// deltaText uses the sign of the painted numbers. %.1f of an exact zero
// can be "-0.0", and a positive residual has no plus of its own.
func deltaText(weight, trend float64) string {
	s := fmt.Sprintf("%.1f", weight-trend)
	if s == "0.0" || s == "-0.0" {
		return "0.0"
	}
	if weight > trend {
		return "+" + s
	}
	return s
}

// encodeNote drops runes the built-in Helvetica cannot draw, before the
// column measures the note. Printable ASCII stays. A Windows-1252 glyph
// other than .notdef stays as that byte, so é is 0xE9 in the stream.
// fpdf's translator would substitute a dot; this does not.
func encodeNote(note string) string {
	var b []byte
	for _, r := range note {
		if r >= 0x20 && r <= 0x7E {
			b = append(b, byte(r))
			continue
		}
		if c, ok := cp1252Byte(r); ok {
			b = append(b, c)
		}
	}
	return string(b)
}

func cp1252Byte(r rune) (byte, bool) {
	if r >= 0xA0 && r <= 0xFF {
		return byte(r), true
	}
	b, ok := cp1252Extra[r]
	return b, ok
}

// cp1252Extra is Windows-1252 bytes 0x80–0x9F that name a glyph.
// 0x81, 0x8D, 0x8F, 0x90, and 0x9D are .notdef and are absent, matching
// charmap.Windows1252. Controls are not in this table.
var cp1252Extra = map[rune]byte{
	0x20AC: 0x80,
	0x201A: 0x82,
	0x0192: 0x83,
	0x201E: 0x84,
	0x2026: 0x85,
	0x2020: 0x86,
	0x2021: 0x87,
	0x02C6: 0x88,
	0x2030: 0x89,
	0x0160: 0x8A,
	0x2039: 0x8B,
	0x0152: 0x8C,
	0x017D: 0x8E,
	0x2018: 0x91,
	0x2019: 0x92,
	0x201C: 0x93,
	0x201D: 0x94,
	0x2022: 0x95,
	0x2013: 0x96,
	0x2014: 0x97,
	0x02DC: 0x98,
	0x2122: 0x99,
	0x0161: 0x9A,
	0x203A: 0x9B,
	0x0153: 0x9C,
	0x017E: 0x9E,
	0x0178: 0x9F,
}

// fitNote cuts only after undrawable runes are gone. The ellipsis is
// three ASCII periods. The core font has no U+2026, and a snowman must
// not be counted as a dot that forces a cut.
func fitNote(pdf *fpdf.Fpdf, s string) string {
	if s == "" {
		return ""
	}
	if pdf.GetStringWidth(s) <= noteInner {
		return s
	}
	const ell = "..."
	best := ""
	for n := 0; n <= len(s); n++ {
		cand := s[:n] + ell
		if pdf.GetStringWidth(cand) > noteInner {
			break
		}
		best = cand
	}
	return best
}

// installOwnerOnly writes data to a 0600 temp in dir, then renames it
// onto path. The destination is not truncated until rename succeeds.
// The steps are local: calling the chart writer would paint a chart.
func installOwnerOnly(path, dir string, data []byte) error {
	tmp, err := os.CreateTemp(dir, ".hdtools-log-*.pdf")
	if err != nil {
		return fmt.Errorf("create temp log: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod log: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write log: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close log: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename log: %w", err)
	}
	cleanup = false
	return nil
}
