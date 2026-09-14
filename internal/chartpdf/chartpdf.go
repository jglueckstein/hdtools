// Package chartpdf writes one landscape US Letter page of a monthly
// chart: title box, daily marks, stems, trend path, day numbers, and
// the same loss copy as the TUI.
//
// Geometry is vector, not character cells. Data rules (clip, empty,
// Y pad, analysis omission) live in chartspan; this package paints.
// It does not open SQLite, import the TUI, or honour NO_COLOR — a
// PDF is a print artefact and is always colour.
package chartpdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/jglueckstein/hdtools/internal/chartspan"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

// Options is the already-trended month the writer paints. Logs are
// the full series (ApplyTrend already applied); the writer clips
// through chartspan. Today is the same local today the TUI uses.
type Options struct {
	Year   int
	Month  time.Month
	Logs   []dailylog.DailyLog
	Unit   units.Unit
	Colors map[string]string
	Today  time.Time
}

const (
	pageW = 792.0
	pageH = 612.0

	titleFillR, titleFillG, titleFillB             = 0x00, 0x00, 0xFF
	titleStrokeR, titleStrokeG, titleStrokeB       = 0xFF, 0x00, 0x00
	titleTextR, titleTextG, titleTextB             = 0xFF, 0xFF, 0x00
	stemR, stemG, stemB                            = 0x00, 0x80, 0x00
	defaultWeightR, defaultWeightG, defaultWeightB = 0x00, 0x00, 0x80
	defaultTrendR, defaultTrendG, defaultTrendB    = 0x80, 0x00, 0x00

	markSize = 3.5
)

// Write paints opts onto a one-page landscape Letter PDF at path,
// mode 0600. The bytes go to a same-dir temp created 0600, then
// Rename onto path, so a failed write leaves the previous file and
// the new file is owner-only from the first byte (like the log
// store). os.Create on the destination would truncate first and
// default to 0666.
func Write(path string, opts Options) error {
	span := chartspan.Clip(opts.Logs, opts.Year, opts.Month, opts.Today)

	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("write chart pdf %s: %w", path, err)
		}
	}

	pdf := fpdf.New("L", "pt", "Letter", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)
	paint(pdf, span, opts)
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return fmt.Errorf("write chart pdf %s: %w", path, err)
	}
	if err := installOwnerOnly(path, dir, buf.Bytes()); err != nil {
		return fmt.Errorf("write chart pdf %s: %w", path, err)
	}
	return nil
}

// installOwnerOnly writes data to a 0600 temp in dir, then renames
// it onto path. The destination is not truncated until rename
// succeeds.
func installOwnerOnly(path, dir string, data []byte) error {
	tmp, err := os.CreateTemp(dir, ".hdtools-chart-*.pdf")
	if err != nil {
		return err
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
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func paint(pdf *fpdf.Fpdf, span chartspan.Span, opts Options) {
	title := fmt.Sprintf("%s %d", opts.Month.String(), opts.Year)
	paintTitle(pdf, title)
	if span.Empty() {
		paintEmpty(pdf)
		return
	}
	paintChart(pdf, span, opts)
}

func paintTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 16)
	padX, padY := 12.0, 8.0
	textW := pdf.GetStringWidth(title)
	boxW := textW + 2*padX
	boxH := 16.0 + 2*padY
	boxX := (pageW - boxW) / 2
	boxY := 28.0
	pdf.SetFillColor(titleFillR, titleFillG, titleFillB)
	pdf.SetDrawColor(titleStrokeR, titleStrokeG, titleStrokeB)
	pdf.SetLineWidth(1.5)
	pdf.Rect(boxX, boxY, boxW, boxH, "FD")
	pdf.SetTextColor(titleTextR, titleTextG, titleTextB)
	pdf.Text(boxX+padX, boxY+padY+16, title)
	pdf.SetTextColor(0, 0, 0)
}

func paintEmpty(pdf *fpdf.Fpdf) {
	pdf.SetFont("Helvetica", "", 14)
	msg := "(empty month)"
	x := (pageW - pdf.GetStringWidth(msg)) / 2
	pdf.SetTextColor(0, 0, 0)
	pdf.Text(x, pageH/2, msg)
}

func paintChart(pdf *fpdf.Fpdf, span chartspan.Span, opts Options) {
	ymin, ymax, ok := chartspan.YRange(span, opts.Unit)
	if !ok {
		return
	}
	plotLeft, plotRight := 72.0, pageW-36.0
	plotTop, plotBottom := 88.0, 500.0
	plotW := plotRight - plotLeft
	plotH := plotBottom - plotTop
	n := len(span.Points)
	if n < 1 {
		return
	}

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(0, 0, 0)
	const ticks = 5
	for i := 0; i < ticks; i++ {
		frac := float64(i) / float64(ticks-1)
		v := ymax - frac*(ymax-ymin)
		label := fmt.Sprintf("%.1f", v)
		y := yAt(v, ymin, ymax, plotTop, plotH)
		pdf.Text(plotLeft-8-pdf.GetStringWidth(label), y+3, label)
	}

	pdf.SetFont("Helvetica", "", 8)
	for i, p := range span.Points {
		s := fmt.Sprintf("%d", p.Day.Day())
		x := xAt(i, n, plotLeft, plotW) - pdf.GetStringWidth(s)/2
		pdf.Text(x, plotBottom+14, s)
	}

	if line, ok := span.Analysis(opts.Unit); ok {
		pdf.SetFont("Helvetica", "", 11)
		pdf.Text(plotLeft, plotBottom+40, line)
	}

	tr, tg, tb := seriesRGB(opts.Colors, "trend", defaultTrendR, defaultTrendG, defaultTrendB)
	wr, wg, wb := seriesRGB(opts.Colors, "weight", defaultWeightR, defaultWeightG, defaultWeightB)

	// Trend first, then stems, then daily marks on top — the book's
	// diamond sits on the line where they meet.
	pdf.SetDrawColor(tr, tg, tb)
	pdf.SetLineWidth(1.2)
	started := false
	for i, p := range span.Points {
		if !p.HasTrend {
			if started {
				pdf.DrawPath("D")
				started = false
			}
			continue
		}
		v, ok := displayKG(p.Trend, opts.Unit)
		if !ok {
			continue
		}
		x := xAt(i, n, plotLeft, plotW)
		y := yAt(v, ymin, ymax, plotTop, plotH)
		if !started {
			pdf.MoveTo(x, y)
			started = true
			continue
		}
		pdf.LineTo(x, y)
	}
	if started {
		pdf.DrawPath("D")
	}

	pdf.SetDrawColor(stemR, stemG, stemB)
	pdf.SetLineWidth(0.8)
	for i, p := range span.Points {
		if p.Weight == nil || !p.HasTrend {
			continue
		}
		wv, wok := displayKG(*p.Weight, opts.Unit)
		tv, tok := displayKG(p.Trend, opts.Unit)
		if !wok || !tok {
			continue
		}
		yw := yAt(wv, ymin, ymax, plotTop, plotH)
		yt := yAt(tv, ymin, ymax, plotTop, plotH)
		if yw == yt {
			continue
		}
		x := xAt(i, n, plotLeft, plotW)
		pdf.MoveTo(x, yw)
		pdf.LineTo(x, yt)
		pdf.DrawPath("D")
	}

	pdf.SetDrawColor(wr, wg, wb)
	pdf.SetFillColor(wr, wg, wb)
	for i, p := range span.Points {
		if p.Weight == nil {
			continue
		}
		v, ok := displayKG(*p.Weight, opts.Unit)
		if !ok {
			continue
		}
		x := xAt(i, n, plotLeft, plotW)
		y := yAt(v, ymin, ymax, plotTop, plotH)
		diamond(pdf, x, y, markSize)
	}
}

func seriesRGB(colors map[string]string, role string, defR, defG, defB int) (r, g, b int) {
	if colors != nil {
		if s, ok := colors[role]; ok {
			if r, g, b, ok := ColorToRGB(s); ok {
				return r, g, b
			}
		}
	}
	return defR, defG, defB
}

func displayKG(kg float64, unit units.Unit) (float64, bool) {
	v, err := units.FromKG(kg, unit)
	if err != nil {
		return 0, false
	}
	return v, true
}

func xAt(i, n int, plotLeft, plotW float64) float64 {
	if n <= 0 {
		return plotLeft
	}
	return plotLeft + (float64(i)+0.5)*plotW/float64(n)
}

func yAt(v, ymin, ymax, plotTop, plotH float64) float64 {
	if ymax == ymin {
		return plotTop + plotH/2
	}
	t := (ymax - v) / (ymax - ymin)
	return plotTop + t*plotH
}

func diamond(pdf *fpdf.Fpdf, x, y, r float64) {
	pdf.MoveTo(x, y-r)
	pdf.LineTo(x+r, y)
	pdf.LineTo(x, y+r)
	pdf.LineTo(x-r, y)
	pdf.ClosePath()
	pdf.DrawPath("FD")
}
