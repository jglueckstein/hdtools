package tui

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func hasSGRCode(s string, code int) bool {
	re := regexp.MustCompile(fmt.Sprintf(`\x1b\[([0-9;]*;)?%d[;m]`, code))
	return re.MatchString(s)
}

func hasIndexedForeground(s string, n int) bool {
	if n >= 0 && n < 8 && hasSGRCode(s, 30+n) {
		return true
	}
	if n >= 8 && n < 16 && hasSGRCode(s, 90+n-8) {
		return true
	}
	return strings.Contains(s, fmt.Sprintf("38;5;%d", n))
}

func hasTruecolorForeground(s string, r, g, b int) bool {
	return strings.Contains(s, fmt.Sprintf("38;2;%d;%d;%d", r, g, b))
}

func hasChromaticSGR(s string) bool {
	return regexp.MustCompile(`\x1b\[(?:3[0-7]|9[0-7]|4[0-7]|10[0-7]|38;|48;)`).MatchString(s)
}

// writtenColor is the SGR a terminal of that profile would show.
// The model text may stay full fidelity; this writer downsamples once.
func writtenColor(t *testing.T, s string, profile colorprofile.Profile) string {
	t.Helper()
	var buf bytes.Buffer
	w := colorprofile.Writer{Forward: &buf, Profile: profile}
	if _, err := w.WriteString(s); err != nil {
		t.Fatalf("write color profile: %v", err)
	}
	return buf.String()
}

func writtenANSI(t *testing.T, s string) string {
	t.Helper()
	return writtenColor(t, s, colorprofile.ANSI)
}

func writtenTrueColor(t *testing.T, s string) string {
	t.Helper()
	return writtenColor(t, s, colorprofile.TrueColor)
}

func schemeCfg(colors map[string]string) config.Config {
	return config.Config{DisplayUnit: units.Kilogram, Colors: colors}
}

func enableChroma(t *testing.T) {
	t.Helper()
	t.Setenv("NO_COLOR", "")
}

func weighedList(t *testing.T, cfg config.Config) *App {
	t.Helper()
	store := openStore(t)
	ctx := context.Background()
	w := 171.5
	log, err := dailylog.New(time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), &w, 8, 1000, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, log); err != nil {
		t.Fatal(err)
	}
	app := New(store, "mem.db", cfg)
	app.Update(app.load())
	return app
}

func selectedLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(visible(line), "> ") && strings.Contains(visible(line), "171.5") {
			return line
		}
	}
	return ""
}

func TestViewUsesSchemeOnList(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "green",
		"trend":  "yellow",
	}))
	view := app.View().Content
	if !hasIndexedForeground(writtenANSI(t, view), 2) {
		t.Fatalf("list missing green weight SGR: %q", view)
	}
	if !hasIndexedForeground(writtenANSI(t, view), 3) {
		t.Fatalf("list missing yellow trend SGR: %q", view)
	}
	if !hasIndexedForeground(writtenANSI(t, view), 7) {
		t.Fatalf("list missing default white header SGR: %q", view)
	}
}

func TestViewUsesSchemeOnMonth(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "green",
		"trend":  "yellow",
	}))
	app.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	view := app.View().Content
	if !hasIndexedForeground(writtenANSI(t, view), 2) {
		t.Fatalf("month missing green weight SGR: %q", view)
	}
	if !hasIndexedForeground(writtenANSI(t, view), 3) {
		t.Fatalf("month missing yellow trend SGR: %q", view)
	}
}

func TestFormLabelsUseTitleRole(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{"title": "magenta"}))
	model, _ := app.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	view := model.View().Content
	if !hasIndexedForeground(writtenANSI(t, view), 5) {
		t.Fatalf("form missing magenta title SGR: %q", view)
	}
	if !strings.Contains(visible(view), ">") {
		t.Fatalf("form missing focus mark: %q", view)
	}
}

func TestSelectionReverseWithOptionalForeground(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight":    "green",
		"selection": "yellow",
	}))
	line := selectedLine(app.View().Content)
	if line == "" {
		t.Fatalf("no selected weight line in %q", app.View().Content)
	}
	if !hasSGRCode(line, 7) {
		t.Fatalf("selected row missing reverse: %q", line)
	}
	if !hasIndexedForeground(writtenANSI(t, line), 3) {
		t.Fatalf("selected row missing yellow foreground: %q", line)
	}
	if !strings.Contains(visible(line), ">") {
		t.Fatalf("selected row missing >: %q", line)
	}
}

func TestMonthSelectionRestylesFocusedCellOnly(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight":    "green",
		"trend":     "cyan",
		"selection": "yellow",
	}))
	app.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	view := app.View().Content
	var line string
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(visible(l), ">") && strings.Contains(visible(l), "171.5") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("no focused month row in %q", view)
	}
	if !hasSGRCode(line, 7) {
		t.Fatalf("focused cell missing reverse: %q", line)
	}
	if !hasIndexedForeground(writtenANSI(t, line), 3) {
		t.Fatalf("focused weight missing yellow: %q", line)
	}
	if !hasIndexedForeground(writtenANSI(t, line), 6) {
		t.Fatalf("trend cell missing cyan: %q", line)
	}
}

func TestFormFocusNotSelectionRole(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"title":     "magenta",
		"selection": "yellow",
	}))
	model, _ := app.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	view := model.View().Content
	if !hasIndexedForeground(writtenANSI(t, view), 5) {
		t.Fatalf("form labels missing magenta: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		vis := visible(line)
		if strings.Contains(vis, "weight") && strings.Contains(vis, "optional") && hasSGRCode(line, 7) {
			t.Fatalf("form label used reverse/selection: %q", line)
		}
	}
	if !strings.Contains(visible(view), ">") || !hasSGRCode(view, 1) {
		t.Fatalf("form focus missing > plus bold: %q", view)
	}
}

func TestNoColorStripsChromaKeepsStructure(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "green",
		"trend":  "yellow",
	}))
	view := app.View().Content
	if hasChromaticSGR(view) {
		t.Fatalf("NO_COLOR left chromatic SGR: %q", view)
	}
	vis := visible(view)
	if !strings.Contains(vis, "weight") || !strings.Contains(vis, "171.5") || !strings.Contains(vis, ">") {
		t.Fatalf("NO_COLOR dropped structure: %q", vis)
	}
	if !hasSGRCode(view, 7) {
		t.Fatalf("NO_COLOR dropped reverse: %q", view)
	}
	if !hasSGRCode(view, 1) {
		t.Fatalf("NO_COLOR dropped bold: %q", view)
	}
}

// TestNoColorStripsFieldChroma is the day form and the month cell.
// The palette does not paint those widgets. Their default greys are
// still chromatic model text, and S13 forbids that under NO_COLOR.
func TestNoColorStripsFieldChroma(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app, _ := newNovemberSheet(t, 4, colNote)
	app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if app.screen != screenForm {
		t.Fatal("day form did not open")
	}
	view := app.View().Content
	if hasChromaticSGR(view) {
		t.Fatalf("NO_COLOR left field chroma: %q", view)
	}
	vis := visible(view)
	if !strings.Contains(vis, "weight") || !strings.Contains(vis, ">") {
		t.Fatalf("NO_COLOR dropped the form: %q", vis)
	}
	app.form.inputs[fieldWeight].SetValue("80")
	view = app.View().Content
	if hasChromaticSGR(view) {
		t.Fatalf("NO_COLOR left blurred field chroma: %q", view)
	}
	if !strings.Contains(visible(view), "80") || !strings.Contains(visible(view), ">") {
		t.Fatalf("NO_COLOR dropped the blurred value: %q", visible(view))
	}

	sheet, _ := newNovemberSheet(t, 4, colNote)
	sheet.month.beginEdit("café")
	view = sheet.View().Content
	if hasChromaticSGR(view) {
		t.Fatalf("NO_COLOR left month-cell chroma: %q", view)
	}
	if !strings.Contains(visible(view), "café") {
		t.Fatalf("NO_COLOR dropped the month cell: %q", visible(view))
	}
}

func TestFormKeepsWidgetColorWhenNoColorEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	app, _ := newNovemberSheet(t, 4, colNote)
	app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !hasChromaticSGR(app.View().Content) {
		t.Fatal("empty NO_COLOR stripped the field colour")
	}
}

func TestEmptyNoColorKeepsDefaultChroma(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	app := weighedList(t, config.Default())
	view := app.View().Content
	if !hasIndexedForeground(writtenANSI(t, view), 4) {
		t.Fatalf("empty NO_COLOR missing default blue weight: %q", view)
	}
	if !hasIndexedForeground(writtenANSI(t, view), 1) {
		t.Fatalf("empty NO_COLOR missing default red trend: %q", view)
	}
}

func TestTrendIsBoldWithAndWithoutColor(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, config.Default())
	if !hasSGRCode(app.View().Content, 1) {
		t.Fatalf("trend not bold with color: %q", app.View().Content)
	}
	t.Setenv("NO_COLOR", "1")
	app = weighedList(t, config.Default())
	if !hasSGRCode(app.View().Content, 1) {
		t.Fatalf("trend not bold under NO_COLOR: %q", app.View().Content)
	}
}

func TestListAlignsWithCustomWeightColor(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "green",
		"trend":  "yellow",
	}))
	view := visible(app.View().Content)
	var header, data string
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "weight") && strings.Contains(line, "trend") {
			header = line
		}
		if strings.Contains(line, "171.5") {
			data = line
		}
	}
	if header == "" || data == "" {
		t.Fatalf("missing header or data in %q", view)
	}
	assertRightAligned(t, header, "weight", data, "171.5")
}

func TestViewUsesHexOnTruecolor(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "#00ff00",
		"trend":  "#0D47A1",
	}))
	written := writtenTrueColor(t, app.View().Content)
	if !hasTruecolorForeground(written, 0, 255, 0) {
		t.Fatalf("missing #00ff00: %q", written)
	}
	if !hasTruecolorForeground(written, 0x0d, 0x47, 0xa1) {
		t.Fatalf("missing #0D47A1: %q", written)
	}
}

func TestHexDownshiftsOnSixteenColor(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{"weight": "#00ff00"}))
	written := writtenANSI(t, app.View().Content)
	if hasTruecolorForeground(written, 0, 255, 0) {
		t.Fatalf("16-color profile emitted 38;2: %q", written)
	}
	if !hasIndexedForeground(written, 2) && !hasIndexedForeground(written, 10) {
		t.Fatalf("#00ff00 did not downshift to green or bright-green: %q", written)
	}
}

func TestNamedAndHexMix(t *testing.T) {
	enableChroma(t)
	app := weighedList(t, schemeCfg(map[string]string{
		"weight": "blue",
		"trend":  "#ff0000",
	}))
	written := writtenTrueColor(t, app.View().Content)
	if !hasIndexedForeground(written, 4) {
		t.Fatalf("named blue missing: %q", written)
	}
	if !hasTruecolorForeground(written, 255, 0, 0) {
		t.Fatalf("hex red missing: %q", written)
	}
}
