package chartpdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestWritePDFHeaderAndOnePage(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := Write(path, novemberPastOpts(t, kgLogs(t, 80, 79))); err != nil {
		t.Fatal(err)
	}
	assertLetterLandscapePDF(t, path)
}

func TestWritePDFFailedWriteLeavesOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(path, []byte("OLD"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	err := Write(path, novemberPastOpts(t, kgLogs(t, 80, 79)))
	if err == nil {
		t.Fatal("want write error in a read-only directory")
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "OLD" {
		t.Fatalf("destination overwritten on failure: %q", got)
	}
}

func TestWritePDFTitle(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := Write(path, novemberPastOpts(t, kgLogs(t, 80, 79))); err != nil {
		t.Fatal(err)
	}
	text := pdfExtractedText(t, path)
	if !strings.Contains(text, "November 1990") {
		t.Fatalf("missing title: %q", text)
	}
	if strings.Contains(text, "hdtools —") {
		t.Fatalf("TUI chrome in PDF: %q", text)
	}
}

func TestWritePDFYRangeKG(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	opts := novemberPastOpts(t, kgLogs(t, 80, 81))
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	text := pdfExtractedText(t, path)
	p, err := units.ToKG(2, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	lo := fmt.Sprintf("%.1f", 80-p)
	hi := fmt.Sprintf("%.1f", 81+p)
	if !strings.Contains(text, lo) || !strings.Contains(text, hi) {
		t.Fatalf("Y labels missing %s and %s in %q", lo, hi, text)
	}
}

func TestWritePDFLossAndDeficit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	opts := novemberPastOpts(t, endpointTrends(t, 80, 79))
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	text := pdfExtractedText(t, path)
	want := "Monthly loss: 1.0 kg   Daily deficit: 257 cal"
	if !strings.Contains(text, want) {
		t.Fatalf("missing %q in %q", want, text)
	}
}

func TestWritePDFEmptyMonth(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	opts := Options{
		Year:  1990,
		Month: time.June,
		Unit:  units.Kilogram,
		Today: time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC),
	}
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	assertLetterLandscapePDF(t, path)
	text := pdfExtractedText(t, path)
	if !strings.Contains(text, "June 1990") {
		t.Fatalf("missing title: %q", text)
	}
	if !strings.Contains(text, "(empty month)") {
		t.Fatalf("missing empty copy: %q", text)
	}
	if strings.Contains(text, "Monthly loss:") {
		t.Fatalf("analysis on empty month: %q", text)
	}
}

func TestWritePDFCarryOnlyNotEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 5, 31, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	trended, err := dailylog.ApplyTrend([]dailylog.DailyLog{log}, nil)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Year:  1990,
		Month: time.June,
		Logs:  trended,
		Unit:  units.Kilogram,
		Today: time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC),
	}
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	raw := readPDF(t, path)
	text := pdfTextFrom(raw)
	if strings.Contains(text, "(empty month)") {
		t.Fatalf("carry-only treated as empty: %q", text)
	}
	if !strings.Contains(text, "Monthly loss: 0.0 kg   Daily deficit: 0 cal") {
		t.Fatalf("missing zero analysis: %q", text)
	}
	if pdfStrokeCount(pdfContent(raw)) < 1 {
		t.Fatal("missing trend path on carry-only month")
	}
}

func TestWritePDFDisplayUnitLB(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	opts := novemberPastOpts(t, kgLogs(t, 80, 80))
	opts.Unit = units.Pound
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	text := pdfExtractedText(t, path)
	lb, err := units.FromKG(80, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	shown := fmt.Sprintf("%.1f", lb)
	if !strings.Contains(text, shown) {
		t.Fatalf("missing pound label %s in %q", shown, text)
	}
	if strings.Contains(text, "80.0") {
		t.Fatalf("axis still in kg: %q", text)
	}
}

func TestWritePDFCurrentMonthClipsAtToday(t *testing.T) {
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "out.pdf")
	opts := Options{
		Year:  1990,
		Month: time.November,
		Logs:  endpointTrendsOnDays(t, 80, 79, 1, 10),
		Unit:  units.Kilogram,
		Today: today,
	}
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	raw := readPDF(t, path)
	parts := pdfLiteralStrings(pdfContent(raw))
	text := strings.Join(parts, " ")
	has10, has30 := false, false
	for _, s := range parts {
		if s == "10" {
			has10 = true
		}
		if s == "30" {
			has30 = true
		}
	}
	if !has10 {
		t.Fatalf("missing day-number label 10 in %q", text)
	}
	if has30 {
		t.Fatalf("day-number label 30 after today in %q", text)
	}
	if !strings.Contains(text, "Daily deficit: 772 cal") {
		t.Fatalf("deficit should divide by 10: %q", text)
	}
	if strings.Contains(text, "Daily deficit: 257 cal") {
		t.Fatalf("deficit divided by 30: %q", text)
	}
}

func TestWritePDFOmitsAnalysisWithoutEndpoints(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 5, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	trended, err := dailylog.ApplyTrend([]dailylog.DailyLog{log}, nil)
	if err != nil {
		t.Fatal(err)
	}
	opts := novemberPastOpts(t, trended)
	if err := Write(path, opts); err != nil {
		t.Fatal(err)
	}
	text := pdfExtractedText(t, path)
	if strings.Contains(text, "Monthly loss:") {
		t.Fatalf("analysis without both endpoints: %q", text)
	}
}

func TestWritePDFVectorsFromFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := Write(path, novemberPastOpts(t, kgLogs(t, 80, 81))); err != nil {
		t.Fatal(err)
	}
	content := pdfContent(readPDF(t, path))
	if !pdfHasToken(content, "m") || !pdfHasToken(content, "l") {
		t.Fatal("PDF missing path moveto/lineto (not a vector chart)")
	}
	if n := pdfStrokeCount(content); n < 2 {
		t.Fatalf("strokes = %d, want a trend stroke and a stem stroke", n)
	}
}

func TestColorToRGB(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		r, g, b int
	}{
		{"#c0a080", 192, 160, 128},
		{"#c0a", 204, 0, 170},
		{"blue", 0, 0, 128},
		{"bright-red", 255, 0, 0},
	}
	for _, tc := range cases {
		r, g, b, ok := ColorToRGB(tc.in)
		if !ok || r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("ColorToRGB(%q) = %d,%d,%d ok=%v, want %d,%d,%d",
				tc.in, r, g, b, ok, tc.r, tc.g, tc.b)
		}
	}
}

func novemberPastOpts(t *testing.T, logs []dailylog.DailyLog) Options {
	t.Helper()
	return Options{
		Year:  1990,
		Month: time.November,
		Logs:  logs,
		Unit:  units.Kilogram,
		Today: time.Date(1990, 12, 1, 12, 0, 0, 0, time.UTC),
	}
}

func kgLogs(t *testing.T, first, last float64) []dailylog.DailyLog {
	t.Helper()
	a, err := dailylog.New(time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), &first, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := dailylog.New(time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), &last, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	out, err := dailylog.ApplyTrend([]dailylog.DailyLog{a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func endpointTrends(t *testing.T, first, last float64) []dailylog.DailyLog {
	t.Helper()
	return endpointTrendsOnDays(t, first, last, 1, 30)
}

func endpointTrendsOnDays(t *testing.T, first, last float64, d1, d2 int) []dailylog.DailyLog {
	t.Helper()
	a, err := dailylog.New(time.Date(1990, 11, d1, 0, 0, 0, 0, time.UTC), &first, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := dailylog.New(time.Date(1990, 11, d2, 0, 0, 0, 0, time.UTC), &last, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	a.Trend = first
	b.Trend = last
	return []dailylog.DailyLog{a, b}
}

func assertLetterLandscapePDF(t *testing.T, path string) {
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
	if math.Abs(w-792) > 0.5 || math.Abs(h-612) > 0.5 {
		t.Fatalf("MediaBox = %.2f×%.2f, want 792×612", w, h)
	}
	if n := pdfPageCount(raw); n != 1 {
		t.Fatalf("pages = %d, want 1", n)
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

func pdfExtractedText(t *testing.T, path string) string {
	t.Helper()
	return pdfTextFrom(readPDF(t, path))
}

func pdfTextFrom(raw []byte) string {
	parts := pdfLiteralStrings(raw)
	parts = append(parts, pdfLiteralStrings(pdfContent(raw))...)
	return strings.Join(parts, " ")
}

var mediaBoxRe = regexp.MustCompile(`/MediaBox\s*\[\s*[\d.]+\s+[\d.]+\s+([\d.]+)\s+([\d.]+)\s*\]`)

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

func pdfContent(raw []byte) []byte {
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
			if dec, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
				plain, err := io.ReadAll(dec)
				_ = dec.Close()
				if err == nil {
					data = plain
				}
			}
		}
		buf.Write(data)
		buf.WriteByte('\n')
		rest = rest[j+len("endstream"):]
	}
	return buf.Bytes()
}

func pdfLiteralStrings(content []byte) []string {
	var out []string
	for i := 0; i < len(content); i++ {
		if content[i] != '(' {
			continue
		}
		i++
		var b strings.Builder
		for i < len(content) {
			if content[i] == '\\' && i+1 < len(content) {
				i++
				switch content[i] {
				case 'n':
					b.WriteByte('\n')
				case '(':
					b.WriteByte('(')
				case ')':
					b.WriteByte(')')
				case '\\':
					b.WriteByte('\\')
				default:
					b.WriteByte(content[i])
				}
				i++
				continue
			}
			if content[i] == ')' {
				break
			}
			b.WriteByte(content[i])
			i++
		}
		out = append(out, b.String())
	}
	return out
}

func pdfHasToken(content []byte, tok string) bool {
	for _, f := range bytes.Fields(content) {
		if string(f) == tok {
			return true
		}
	}
	return false
}

func pdfStrokeCount(content []byte) int {
	n := 0
	for _, f := range bytes.Fields(content) {
		if string(f) == "S" || string(f) == "s" {
			n++
		}
	}
	return n
}
