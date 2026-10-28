package logpdf

import (
	"bytes"
	"compress/zlib"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestWriteLogPDFHeaderFitsColumns(t *testing.T) {
	t.Parallel()
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Unit:  units.Kilogram,
	})))
	wantHeaderInk(t, draw)
}

func TestWriteLogPDFPortraitOnePage(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 8, 0, false, "")
	log.Trend = 80
	path := writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
	})
	raw := assertPortrait(t, path)
	assertBuiltInHelvetica(t, raw)
	draw := scanSheet(t, raw)
	// One rule under each of the 30 day rows, and rules on the column boundaries.
	if draw.horiz+draw.rects < 30 {
		t.Fatalf("horizontal rules = %d, want at least one per November day", draw.horiz+draw.rects)
	}
	if draw.vert+draw.rects < 8 {
		t.Fatalf("vertical rules = %d, want the column boundaries", draw.vert+draw.rects)
	}
}

func TestWriteLogPDFHeading(t *testing.T) {
	t.Parallel()
	for _, unit := range []units.Unit{units.Pound, units.Kilogram, units.Stone} {
		t.Run(string(unit), func(t *testing.T) {
			t.Parallel()
			path := writeSheet(t, Options{Year: 1990, Month: time.November, Unit: unit})
			draw := scanSheet(t, readPDF(t, path))
			want := "November 1990 (" + string(unit) + ")"
			if !hasLiteral(draw.spans, want) {
				t.Fatalf("heading literals = %q, want %q", spanTexts(draw.spans), want)
			}
		})
	}
}

func TestWriteLogPDFOmitsChrome(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 8, 1000, true, "walk")
	log.Trend = 80
	path := writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
	})
	raw := readPDF(t, path)
	draw := scanSheet(t, raw)
	blob := strings.Join(spanTexts(draw.spans), "\n") + "\n" + string(pdfStreamContent(t, raw))
	for _, forbidden := range []string{
		"arrows move",
		"Monthly loss",
		"Daily deficit",
		"(empty month)",
		"hdtools",
		".db",
	} {
		if strings.Contains(blob, forbidden) {
			t.Fatalf("page contains %q", forbidden)
		}
	}
}

func TestWriteLogPDFCalendarDays(t *testing.T) {
	t.Parallel()
	t.Run("november-1990", func(t *testing.T) {
		t.Parallel()
		log := mustLog(t, 1990, time.November, 1, kg(80), 8, 0, false, "")
		log.Trend = 80
		raw := readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.November,
			Logs:  []dailylog.DailyLog{log},
			Unit:  units.Kilogram,
		}))
		draw := scanSheet(t, raw)
		wantHeader(t, draw.spans)
		if !hasDay(draw.spans, "1") {
			t.Fatal("day 1 is not a row")
		}
		wantCell(t, draw.spans, "1", "weekday", "Th")
		wantCell(t, draw.spans, "3", "weekday", "Sa")
		for d := 1; d <= 30; d++ {
			if !hasDay(draw.spans, strconv.Itoa(d)) {
				t.Fatalf("day %d is not a row", d)
			}
		}
		wantNoLiteral(t, draw.spans, "31")
	})
	t.Run("february-1990", func(t *testing.T) {
		t.Parallel()
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.February,
			Unit:  units.Kilogram,
		})))
		if !hasDay(draw.spans, "28") {
			t.Fatal("day 28 is not a row")
		}
		wantNoLiteral(t, draw.spans, "29")
	})
	t.Run("february-1992", func(t *testing.T) {
		t.Parallel()
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1992,
			Month: time.February,
			Unit:  units.Kilogram,
		})))
		if !hasDay(draw.spans, "29") {
			t.Fatal("day 29 is not a row")
		}
	})
	t.Run("january-1991", func(t *testing.T) {
		t.Parallel()
		path := writeSheet(t, Options{Year: 1991, Month: time.January, Unit: units.Kilogram})
		raw := assertPortrait(t, path)
		draw := scanSheet(t, raw)
		for d := 1; d <= 31; d++ {
			if !hasDay(draw.spans, strconv.Itoa(d)) {
				t.Fatalf("day %d is not a row", d)
			}
		}
		sp, ok := dayBaseline(draw.spans, "31")
		if !ok {
			t.Fatal("day 31 has no text baseline; extracting the string \"31\" is not enough")
		}
		if sp.x < 0 || sp.x > 612 || sp.y < 0 || sp.y > 792 {
			t.Fatalf("day 31 baseline (%.2f, %.2f) is outside the 612 by 792 media box", sp.x, sp.y)
		}
	})
}

func TestWriteLogPDFIncludesDaysAfterToday(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 15, kg(81), 8, 0, false, "")
	log.Trend = 81
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
		Today: time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC),
	})))
	wantCell(t, draw.spans, "15", "weight", "81.0")
	if !hasDay(draw.spans, "30") {
		t.Fatal("day 30 is not a row; the sheet stopped at today")
	}
}

func TestWriteLogPDFCarriedTrend(t *testing.T) {
	t.Parallel()
	t.Run("from-october", func(t *testing.T) {
		t.Parallel()
		log := mustLog(t, 1990, time.October, 31, kg(80), 0, 0, false, "")
		log.Trend = 80
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.November,
			Logs:  []dailylog.DailyLog{log},
			Unit:  units.Kilogram,
		})))
		wantCell(t, draw.spans, "1", "trend", "80.0")
		wantBlank(t, draw.spans, "1", "weight")
		wantBlank(t, draw.spans, "1", "delta")
	})
	t.Run("after-today", func(t *testing.T) {
		t.Parallel()
		log := mustLog(t, 1990, time.November, 10, kg(80), 0, 0, false, "")
		log.Trend = 80
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.November,
			Logs:  []dailylog.DailyLog{log},
			Unit:  units.Kilogram,
			Today: time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC),
		})))
		wantCell(t, draw.spans, "15", "trend", "80.0")
		wantBlank(t, draw.spans, "15", "weight")
	})
}

func TestWriteLogPDFEmptySeries(t *testing.T) {
	t.Parallel()
	raw := readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Unit:  units.Kilogram,
	}))
	draw := scanSheet(t, raw)
	for d := 1; d <= 30; d++ {
		day := strconv.Itoa(d)
		if !hasDay(draw.spans, day) {
			t.Fatalf("day %d is not a row", d)
		}
		for _, col := range []string{"weight", "trend", "delta", "sleep", "steps", "workout", "note"} {
			wantBlank(t, draw.spans, day, col)
		}
	}
	wantNoPlaceholders(t, draw.spans)
}

func TestWriteLogPDFDeltaKg(t *testing.T) {
	t.Parallel()
	up := mustLog(t, 1990, time.November, 1, kg(80.5), 0, 0, false, "")
	up.Trend = 80
	down := mustLog(t, 1990, time.November, 2, kg(79.5), 0, 0, false, "")
	down.Trend = 80
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{up, down},
		Unit:  units.Kilogram,
	})))
	wantCell(t, draw.spans, "1", "weight", "80.5")
	wantCell(t, draw.spans, "1", "trend", "80.0")
	wantCell(t, draw.spans, "1", "delta", "+0.5")
	wantCell(t, draw.spans, "2", "weight", "79.5")
	wantCell(t, draw.spans, "2", "trend", "80.0")
	wantCell(t, draw.spans, "2", "delta", "-0.5")
}

func TestWriteLogPDFDeltaFromPaintedPounds(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 0, 0, false, "")
	log.Trend = 79.9
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Pound,
	})))
	wantCell(t, draw.spans, "1", "weight", "176.4")
	wantCell(t, draw.spans, "1", "trend", "176.1")
	got, ok := rowCells(draw.spans, "1")["delta"]
	if got == "+0.2" {
		t.Fatal("delta subtracted the kilogram gap and converted, want the painted cells")
	}
	if !ok || got != "+0.3" {
		t.Fatalf("delta = %q, want +0.3", got)
	}
}

func TestWriteLogPDFFirstWeighInPounds(t *testing.T) {
	t.Parallel()
	stored, err := units.ToKG(176.5, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	log := mustLog(t, 1990, time.November, 1, &stored, 0, 0, false, "")
	trended, err := dailylog.ApplyTrend([]dailylog.DailyLog{log}, nil)
	if err != nil {
		t.Fatal(err)
	}
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  trended,
		Unit:  units.Pound,
	})))
	wantCell(t, draw.spans, "1", "weight", "176.5")
	wantCell(t, draw.spans, "1", "trend", "176.6")
	wantCell(t, draw.spans, "1", "delta", "-0.1")
}

func TestWriteLogPDFExactZero(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 1, 0, false, "")
	log.Trend = 80
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
	})))
	wantCell(t, draw.spans, "1", "delta", "0.0")
	wantNoLiteral(t, draw.spans, "+0.0")
	wantNoLiteral(t, draw.spans, "-0.0")
}

func TestWriteLogPDFDeltaBlank(t *testing.T) {
	t.Parallel()
	first := mustLog(t, 1990, time.November, 5, kg(80), 0, 0, false, "")
	first.Trend = 80
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{first},
		Unit:  units.Kilogram,
	})))
	wantBlank(t, draw.spans, "1", "trend")
	wantBlank(t, draw.spans, "1", "delta")
	wantCell(t, draw.spans, "6", "trend", "80.0")
	wantBlank(t, draw.spans, "6", "weight")
	wantBlank(t, draw.spans, "6", "delta")
}

func TestWriteLogPDFStoneOneDecimal(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 0, 0, false, "")
	log.Trend = 80
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Stone,
	})))
	wantCell(t, draw.spans, "1", "weight", "12.6")
	wantCell(t, draw.spans, "1", "trend", "12.6")
	wantCell(t, draw.spans, "1", "delta", "0.0")
}

func TestWriteLogPDFSleepStepsWorkout(t *testing.T) {
	t.Parallel()
	zeros := mustLog(t, 1990, time.November, 1, kg(80), 0, 0, false, "")
	zeros.Trend = 80
	full := mustLog(t, 1990, time.November, 2, kg(80), 7.5, 100000, true, "")
	full.Trend = 80
	gap := mustLog(t, 1990, time.November, 3, nil, 8, 100, false, "")
	raw := readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{zeros, full, gap},
		Unit:  units.Kilogram,
	}))
	wantOnePage(t, raw)
	draw := scanSheet(t, raw)
	wantCell(t, draw.spans, "1", "sleep", "0.0")
	wantCell(t, draw.spans, "1", "steps", "0")
	wantCell(t, draw.spans, "1", "workout", "no")
	wantCell(t, draw.spans, "2", "sleep", "7.5")
	wantCell(t, draw.spans, "2", "steps", "100000")
	wantCell(t, draw.spans, "2", "workout", "yes")
	wantBlank(t, draw.spans, "3", "weight")
	wantCell(t, draw.spans, "3", "sleep", "8.0")
	wantCell(t, draw.spans, "3", "steps", "100")
	wantCell(t, draw.spans, "3", "workout", "no")
}

func TestWriteLogPDFBlankDayEmpty(t *testing.T) {
	t.Parallel()
	log := mustLog(t, 1990, time.November, 1, kg(80), 8, 100, true, "walk")
	log.Trend = 80
	draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
	})))
	if !hasDay(draw.spans, "3") {
		t.Fatal("day 3 is not a row")
	}
	for _, col := range []string{"sleep", "steps", "workout", "note"} {
		got, ok := rowCells(draw.spans, "3")[col]
		if ok {
			t.Fatalf("day 3 %s = %q, want no text (not 0.0, 0, no, -, or an em dash)", col, got)
		}
	}
}

func TestWriteLogPDFNoteCut(t *testing.T) {
	t.Parallel()
	t.Run("walk", func(t *testing.T) {
		t.Parallel()
		draw := scanNote(t, "walk")
		wantCell(t, draw.spans, "1", "note", "walk")
	})
	t.Run("parentheses", func(t *testing.T) {
		t.Parallel()
		draw := scanNote(t, "say (hi)")
		wantCell(t, draw.spans, "1", "note", "say (hi)")
	})
	t.Run("long", func(t *testing.T) {
		t.Parallel()
		note := strings.Repeat("x", 200)
		raw := readPDF(t, writeSheet(t, noteOpts(t, note)))
		wantOnePage(t, raw)
		draw := scanSheet(t, raw)
		got, ok := rowCells(draw.spans, "1")["note"]
		if !ok || !strings.HasSuffix(got, "...") || strings.Contains(got, "\u2026") {
			t.Fatalf("note = %q, want a cut ending with three periods", got)
		}
		body := strings.TrimSuffix(got, "...")
		if !strings.HasPrefix(note, body) || strings.Contains(got, note) {
			t.Fatalf("note = %q, want a prefix of the stored note plus ellipsis", got)
		}
		for _, s := range draw.spans {
			if s.text != got && strings.Contains(s.text, "xxx") {
				t.Fatalf("dropped tail %q is elsewhere on the page", s.text)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		draw := scanNote(t, "")
		wantBlank(t, draw.spans, "1", "note")
	})
	t.Run("snowman-suffix", func(t *testing.T) {
		t.Parallel()
		draw := scanNote(t, "say (hi)\u2603")
		wantCell(t, draw.spans, "1", "note", "say (hi)")
	})
	t.Run("snowman-inside-fit", func(t *testing.T) {
		t.Parallel()
		// Drop the snowman before measuring. The x run fits. The run
		// plus the width of "." does not, so counting the snowman as a
		// dot ellipsizes a note that should be printed whole.
		run := fittingXRun(t)
		mid := len(run) / 2
		draw := scanNote(t, run[:mid]+"\u2603"+run[mid:])
		wantCell(t, draw.spans, "1", "note", run)
	})
}

func TestWriteLogPDFSkipsUndrawable(t *testing.T) {
	t.Parallel()
	t.Run("cafe", func(t *testing.T) {
		t.Parallel()
		raw := readPDF(t, writeSheet(t, noteOpts(t, "café")))
		wantOnePage(t, raw)
		content := pdfStreamContent(t, raw)
		if !bytes.Contains(content, []byte{0xE9}) {
			t.Fatal("café did not contribute byte 0xE9 in the content stream")
		}
		draw := scanSheet(t, raw)
		wantCell(t, draw.spans, "1", "note", "caf\xe9")
	})
	t.Run("snowman-between", func(t *testing.T) {
		t.Parallel()
		draw := scanNote(t, "a\u2603b")
		got, ok := rowCells(draw.spans, "1")["note"]
		if got == "a.b" || got == "a?b" {
			t.Fatalf("snowman replaced with a mark: %q", got)
		}
		if !ok || got != "ab" {
			t.Fatalf("note = %q, want ab", got)
		}
	})
	t.Run("snowman-only", func(t *testing.T) {
		t.Parallel()
		raw := readPDF(t, writeSheet(t, noteOpts(t, "\u2603")))
		wantOnePage(t, raw)
		draw := scanSheet(t, raw)
		wantBlank(t, draw.spans, "1", "note")
	})
	t.Run("newline", func(t *testing.T) {
		t.Parallel()
		raw := readPDF(t, writeSheet(t, noteOpts(t, "ab\ncd")))
		wantOnePage(t, raw)
		draw := scanSheet(t, raw)
		wantCell(t, draw.spans, "1", "note", "abcd")
	})
}

func TestWriteLogPDFFailedWriteLeavesOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(path, []byte("OLD"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	err := Write(path, Options{Year: 1990, Month: time.November, Unit: units.Kilogram})
	if err == nil {
		t.Fatal("want write error in a read-only directory")
	}
	if !strings.Contains(err.Error(), "create temp log") {
		t.Fatalf("error %q does not name the temp-file step", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "OLD" {
		t.Fatalf("destination = %q, want OLD", got)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".hdtools-log-") {
			t.Fatalf("temp left behind: %s", entry.Name())
		}
	}
}

func TestWriteLogPDFNoWeighIn(t *testing.T) {
	t.Parallel()
	t.Run("no-weights", func(t *testing.T) {
		t.Parallel()
		a := mustLog(t, 1990, time.November, 2, nil, 8, 100, false, "rest")
		b := mustLog(t, 1990, time.November, 4, nil, 7.5, 100000, true, "walk")
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.November,
			Logs:  []dailylog.DailyLog{a, b},
			Unit:  units.Kilogram,
		})))
		for d := 1; d <= 30; d++ {
			day := strconv.Itoa(d)
			wantBlank(t, draw.spans, day, "weight")
			wantBlank(t, draw.spans, day, "trend")
			wantBlank(t, draw.spans, day, "delta")
		}
		wantCell(t, draw.spans, "2", "sleep", "8.0")
		wantCell(t, draw.spans, "2", "steps", "100")
		wantCell(t, draw.spans, "2", "workout", "no")
		wantCell(t, draw.spans, "2", "note", "rest")
		wantCell(t, draw.spans, "4", "sleep", "7.5")
		wantCell(t, draw.spans, "4", "steps", "100000")
		wantCell(t, draw.spans, "4", "workout", "yes")
		wantCell(t, draw.spans, "4", "note", "walk")
	})
	t.Run("leading-days", func(t *testing.T) {
		t.Parallel()
		gap := mustLog(t, 1990, time.November, 1, nil, 6, 10, false, "early")
		first := mustLog(t, 1990, time.November, 3, kg(80), 0, 0, false, "")
		first.Trend = 80
		draw := scanSheet(t, readPDF(t, writeSheet(t, Options{
			Year:  1990,
			Month: time.November,
			Logs:  []dailylog.DailyLog{gap, first},
			Unit:  units.Kilogram,
		})))
		wantBlank(t, draw.spans, "1", "trend")
		wantBlank(t, draw.spans, "1", "delta")
		wantBlank(t, draw.spans, "2", "trend")
		wantBlank(t, draw.spans, "2", "delta")
		wantCell(t, draw.spans, "3", "trend", "80.0")
	})
}

func writeSheet(t *testing.T, opts Options) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sheet.pdf")
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	return path
}

func noteOpts(t *testing.T, note string) Options {
	t.Helper()
	log := mustLog(t, 1990, time.November, 1, kg(80), 0, 0, false, note)
	log.Trend = 80
	return Options{
		Year:  1990,
		Month: time.November,
		Logs:  []dailylog.DailyLog{log},
		Unit:  units.Kilogram,
	}
}

func scanNote(t *testing.T, note string) pdfDraw {
	t.Helper()
	return scanSheet(t, readPDF(t, writeSheet(t, noteOpts(t, note))))
}

func mustLog(t *testing.T, year int, month time.Month, day int, weight *float64, sleep float64, steps int, workout bool, note string) dailylog.DailyLog {
	t.Helper()
	log, err := dailylog.New(time.Date(year, month, day, 0, 0, 0, 0, time.UTC), weight, sleep, steps, workout, note)
	if err != nil {
		t.Fatal(err)
	}
	return log
}

func kg(v float64) *float64 { return &v }

// fittingXRun is the longest run of "x" in 8 pt Helvetica that fits the
// note column's inner width. One more dot does not fit, which is the
// snowman-as-dot trap.
func fittingXRun(t *testing.T) string {
	t.Helper()
	pdf := fpdf.New("P", "pt", "Letter", "")
	pdf.SetFont("Helvetica", "", 8)
	if err := pdf.Error(); err != nil {
		t.Fatal(err)
	}
	const inner = 220 - 4
	dot := pdf.GetStringWidth(".")
	n := 0
	for {
		if pdf.GetStringWidth(strings.Repeat("x", n+1)) > inner {
			break
		}
		n++
		if n > 500 {
			t.Fatal("note column fits more than 500 x")
		}
	}
	if n == 0 {
		t.Fatal("no x fits the note column")
	}
	run := strings.Repeat("x", n)
	if pdf.GetStringWidth(run)+dot <= inner {
		t.Fatalf("run of %d x plus a dot still fits", n)
	}
	return run
}

func assertPortrait(t *testing.T, path string) []byte {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
	raw := readPDF(t, path)
	if !bytes.HasPrefix(raw, []byte("%PDF")) {
		t.Fatalf("header = %q, want %%PDF", raw[:min(8, len(raw))])
	}
	w, h, ok := pdfMediaBox(raw)
	if !ok {
		t.Fatal("missing MediaBox")
	}
	if math.Abs(w-612) > 0.5 || math.Abs(h-792) > 0.5 {
		t.Fatalf("MediaBox = %.2f×%.2f, want 612×792", w, h)
	}
	wantOnePage(t, raw)
	return raw
}

func wantOnePage(t *testing.T, raw []byte) {
	t.Helper()
	if n := pdfPageCount(raw); n != 1 {
		t.Fatalf("pages = %d, want 1", n)
	}
}

func assertBuiltInHelvetica(t *testing.T, raw []byte) {
	t.Helper()
	if !bytes.Contains(raw, []byte("/Helvetica")) {
		t.Fatal("missing built-in /Helvetica")
	}
	if bytes.Contains(raw, []byte("/FontFile")) {
		t.Fatal("embedded font program")
	}
	if imageXObjectRe.Match(raw) {
		t.Fatal("embedded image")
	}
}

func readPDF(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func scanSheet(t *testing.T, raw []byte) pdfDraw {
	t.Helper()
	return scanPDF(pdfStreamContent(t, raw))
}

func wantHeader(t *testing.T, spans []pdfSpan) {
	t.Helper()
	var y float64
	found := false
	for _, s := range spans {
		if s.text == "day" {
			y = s.y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("header has no day")
	}
	var row []pdfSpan
	for _, s := range spans {
		if math.Abs(s.y-y) < 0.8 && s.text != "" {
			row = append(row, s)
		}
	}
	sort.Slice(row, func(i, j int) bool { return row[i].x < row[j].x })
	got := spanTexts(row)
	want := []string{"day", "weekday", "weight", "trend", "delta", "sleep", "steps", "workout", "note"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("header = %q, want %q", got, want)
	}
}

// wantHeaderInk fails when a header word's ink crosses the next
// vertical rule. Classifying the cell by the baseline's start
// misses that overlap.
func wantHeaderInk(t *testing.T, draw pdfDraw) {
	t.Helper()
	var y float64
	found := false
	for _, s := range draw.spans {
		if s.text == "day" {
			y = s.y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("header has no day")
	}
	rules := vertRules(draw.vertX)
	if len(rules) < 2 {
		t.Fatalf("vertical rules = %v, want the column boundaries", rules)
	}
	pdf := fpdf.New("P", "pt", "Letter", "")
	pdf.SetFont("Helvetica", "", 8)
	if err := pdf.Error(); err != nil {
		t.Fatal(err)
	}
	for _, s := range draw.spans {
		if math.Abs(s.y-y) >= 0.8 || s.text == "" {
			continue
		}
		end := s.x + pdf.GetStringWidth(s.text)
		next, ok := nextRule(rules, s.x)
		if !ok {
			t.Fatalf("header %q at x %.2f has no rule to its right", s.text, s.x)
		}
		if end >= next {
			t.Fatalf("header %q ends at %.2f, on or past the rule at %.2f", s.text, end, next)
		}
	}
}

func vertRules(xs []float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	out := []float64{sorted[0]}
	for _, x := range sorted[1:] {
		if math.Abs(x-out[len(out)-1]) < 0.2 {
			continue
		}
		out = append(out, x)
	}
	return out
}

func nextRule(rules []float64, x float64) (float64, bool) {
	for _, r := range rules {
		if r > x+0.5 {
			return r, true
		}
	}
	return 0, false
}

func wantCell(t *testing.T, spans []pdfSpan, day, col, want string) {
	t.Helper()
	got, ok := rowCells(spans, day)[col]
	if !ok {
		t.Fatalf("day %s %s has no text, want %q", day, col, want)
	}
	if got != want {
		t.Fatalf("day %s %s = %q, want %q", day, col, got, want)
	}
}

func wantBlank(t *testing.T, spans []pdfSpan, day, col string) {
	t.Helper()
	if !hasDay(spans, day) {
		t.Fatalf("day %s is not a row", day)
	}
	if got, ok := rowCells(spans, day)[col]; ok {
		t.Fatalf("day %s %s = %q, want no text", day, col, got)
	}
}

func hasDay(spans []pdfSpan, day string) bool {
	return rowCells(spans, day) != nil
}

func wantNoLiteral(t *testing.T, spans []pdfSpan, lit string) {
	t.Helper()
	for _, s := range spans {
		if s.text == lit {
			t.Fatalf("literal %q is on the page at (%.2f, %.2f)", lit, s.x, s.y)
		}
	}
}

func wantNoPlaceholders(t *testing.T, spans []pdfSpan) {
	t.Helper()
	for _, s := range spans {
		switch s.text {
		case "(empty month)", "-", "—", "–", "blank":
			t.Fatalf("placeholder %q at (%.2f, %.2f)", s.text, s.x, s.y)
		}
	}
}

func hasLiteral(spans []pdfSpan, lit string) bool {
	for _, s := range spans {
		if s.text == lit {
			return true
		}
	}
	return false
}

func spanTexts(spans []pdfSpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.text)
	}
	return out
}

func rowCells(spans []pdfSpan, day string) map[string]string {
	var y float64
	found := false
	for _, s := range spans {
		col, ok := columnOf(s.x)
		if ok && col == "day" && s.text == day {
			y = s.y
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	cells := map[string]string{}
	for _, s := range spans {
		if math.Abs(s.y-y) >= 0.8 || s.text == "" {
			continue
		}
		col, ok := columnOf(s.x)
		if !ok {
			continue
		}
		if prev, exists := cells[col]; exists {
			cells[col] = prev + "|" + s.text
		} else {
			cells[col] = s.text
		}
	}
	return cells
}

func dayBaseline(spans []pdfSpan, day string) (pdfSpan, bool) {
	for _, s := range spans {
		col, ok := columnOf(s.x)
		if ok && col == "day" && s.text == day && s.pos {
			return s, true
		}
	}
	return pdfSpan{}, false
}

// columnOf maps a text baseline onto the sheet geometry: 36 pt margin,
// then day 28, weekday 36, weight 40, trend 48, delta 44, sleep 36,
// steps 52, workout 36, note 220. A blank cell draws nothing, so the
// column is which band the baseline sits in.
func columnOf(x float64) (string, bool) {
	cols := []struct {
		name string
		w    float64
	}{
		{"day", 28},
		{"weekday", 36},
		{"weight", 40},
		{"trend", 48},
		{"delta", 44},
		{"sleep", 36},
		{"steps", 52},
		{"workout", 36},
		{"note", 220},
	}
	left := 36.0
	for _, c := range cols {
		right := left + c.w
		if x >= left-1 && x < right-0.5 {
			return c.name, true
		}
		left = right
	}
	return "", false
}

type pdfSpan struct {
	text string
	x, y float64
	pos  bool
}

type pdfDraw struct {
	spans       []pdfSpan
	horiz, vert int
	vertX       []float64
	rects       int
}

func scanPDF(content []byte) pdfDraw {
	var d pdfDraw
	var nums []float64
	var x, y float64
	var havePos bool
	var cx, cy float64
	var haveC bool
	for i := 0; i < len(content); {
		c := content[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == 0:
			i++
		case c == '%':
			for i < len(content) && content[i] != '\n' && content[i] != '\r' {
				i++
			}
		case c == '(':
			s, next := readPDFString(content, i)
			i = next
			j := skipPDFSpace(content, i)
			if hasPDFOp(content[j:], "Tj") || hasPDFOp(content[j:], "TJ") {
				op := "Tj"
				if hasPDFOp(content[j:], "TJ") && !hasPDFOp(content[j:], "Tj") {
					op = "TJ"
				}
				d.spans = append(d.spans, pdfSpan{text: s, x: x, y: y, pos: havePos})
				i = j + len(op)
			}
		case c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9'):
			n, next, ok := readPDFNum(content, i)
			if !ok {
				i++
				continue
			}
			nums = append(nums, n)
			i = next
		case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			start := i
			i++
			for i < len(content) && isPDFOpByte(content[i]) {
				i++
			}
			applyPDFOp(&d, string(content[start:i]), &nums, &x, &y, &havePos, &cx, &cy, &haveC)
		case c == '/':
			i++
			for i < len(content) && !isPDFDelim(content[i]) {
				i++
			}
		case c == '<':
			i = skipPDFAngle(content, i)
		default:
			i++
		}
	}
	return d
}

func applyPDFOp(d *pdfDraw, op string, nums *[]float64, x, y *float64, havePos *bool, cx, cy *float64, haveC *bool) {
	switch op {
	case "Td", "TD":
		if len(*nums) >= 2 {
			*x = (*nums)[len(*nums)-2]
			*y = (*nums)[len(*nums)-1]
			*havePos = true
		}
	case "Tm":
		if len(*nums) >= 6 {
			*x = (*nums)[len(*nums)-2]
			*y = (*nums)[len(*nums)-1]
			*havePos = true
		}
	case "m":
		if len(*nums) >= 2 {
			*cx = (*nums)[len(*nums)-2]
			*cy = (*nums)[len(*nums)-1]
			*haveC = true
		}
	case "l":
		if *haveC && len(*nums) >= 2 {
			x2 := (*nums)[len(*nums)-2]
			y2 := (*nums)[len(*nums)-1]
			if math.Abs(y2-*cy) < 0.2 && math.Abs(x2-*cx) > 1 {
				d.horiz++
			}
			if math.Abs(x2-*cx) < 0.2 && math.Abs(y2-*cy) > 1 {
				d.vert++
				d.vertX = append(d.vertX, *cx)
			}
			*cx, *cy = x2, y2
		}
	case "re":
		d.rects++
	}
	*nums = nil
}

func readPDFString(content []byte, i int) (string, int) {
	i++
	var b strings.Builder
	for i < len(content) {
		if content[i] == '\\' && i+1 < len(content) {
			i++
			c := content[i]
			switch c {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case '(', ')', '\\':
				b.WriteByte(c)
			default:
				if c >= '0' && c <= '7' {
					oct := int(c - '0')
					for n := 0; n < 2 && i+1 < len(content) && content[i+1] >= '0' && content[i+1] <= '7'; n++ {
						i++
						oct = oct*8 + int(content[i]-'0')
					}
					b.WriteByte(byte(oct))
				} else {
					b.WriteByte(c)
				}
			}
			i++
			continue
		}
		if content[i] == ')' {
			return b.String(), i + 1
		}
		b.WriteByte(content[i])
		i++
	}
	return b.String(), i
}

func readPDFNum(content []byte, i int) (float64, int, bool) {
	start := i
	if content[i] == '+' || content[i] == '-' {
		i++
		if i >= len(content) {
			return 0, start, false
		}
	}
	saw := false
	for i < len(content) && content[i] >= '0' && content[i] <= '9' {
		saw = true
		i++
	}
	if i < len(content) && content[i] == '.' {
		i++
		for i < len(content) && content[i] >= '0' && content[i] <= '9' {
			saw = true
			i++
		}
	}
	if !saw {
		return 0, start, false
	}
	v, err := strconv.ParseFloat(string(content[start:i]), 64)
	if err != nil {
		return 0, start, false
	}
	return v, i, true
}

func skipPDFSpace(content []byte, i int) int {
	for i < len(content) {
		switch content[i] {
		case ' ', '\t', '\n', '\r', '\f', 0:
			i++
		default:
			return i
		}
	}
	return i
}

func hasPDFOp(content []byte, op string) bool {
	if !bytes.HasPrefix(content, []byte(op)) {
		return false
	}
	if len(content) == len(op) {
		return true
	}
	return isPDFDelim(content[len(op)])
}

func isPDFOpByte(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '*'
}

func isPDFDelim(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\f', 0, '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

func skipPDFAngle(content []byte, i int) int {
	if i+1 < len(content) && content[i+1] == '<' {
		i += 2
		for i+1 < len(content) && !(content[i] == '>' && content[i+1] == '>') {
			i++
		}
		if i+1 < len(content) {
			return i + 2
		}
		return len(content)
	}
	i++
	for i < len(content) && content[i] != '>' {
		i++
	}
	if i < len(content) {
		i++
	}
	return i
}

var mediaBoxRe = regexp.MustCompile(`/MediaBox\s*\[\s*[\d.]+\s+[\d.]+\s+([\d.]+)\s+([\d.]+)\s*\]`)

var imageXObjectRe = regexp.MustCompile(`/Subtype\s*/Image\b`)

func pdfMediaBox(raw []byte) (w, h float64, ok bool) {
	m := mediaBoxRe.FindSubmatch(raw)
	if m == nil {
		return 0, 0, false
	}
	w, err1 := strconv.ParseFloat(string(m[1]), 64)
	h, err2 := strconv.ParseFloat(string(m[2]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return w, h, true
}

var pageTypeRe = regexp.MustCompile(`/Type\s*/Page\b`)

func pdfPageCount(raw []byte) int {
	return len(pageTypeRe.FindAll(raw, -1))
}

func pdfStreamContent(t *testing.T, raw []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	rest := raw
	for {
		i := bytes.Index(rest, []byte("stream"))
		if i < 0 {
			break
		}
		dictStart := bytes.LastIndex(rest[:i], []byte("<<"))
		dict := []byte{}
		if dictStart >= 0 {
			dict = rest[dictStart:i]
		}
		rest = rest[i+len("stream"):]
		if len(rest) > 0 && rest[0] == '\r' {
			rest = rest[1:]
		}
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		j := bytes.Index(rest, []byte("endstream"))
		if j < 0 {
			break
		}
		data := rest[:j]
		if bytes.HasSuffix(data, []byte("\r\n")) {
			data = data[:len(data)-2]
		} else if bytes.HasSuffix(data, []byte("\n")) {
			data = data[:len(data)-1]
		}
		if bytes.Contains(dict, []byte("FlateDecode")) {
			dec, err := zlib.NewReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("inflate content: %v", err)
			}
			plain, err := io.ReadAll(dec)
			_ = dec.Close()
			if err != nil {
				t.Fatalf("inflate content: %v", err)
			}
			data = plain
		}
		buf.Write(data)
		buf.WriteByte('\n')
		rest = rest[j+len("endstream"):]
	}
	return buf.Bytes()
}
