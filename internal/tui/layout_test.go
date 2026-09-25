package tui

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visible(s string) string {
	return ansiSeq.ReplaceAllString(s, "")
}

func TestVisPadIgnoresANSI(t *testing.T) {
	t.Parallel()
	plain := visPad("172.5", 7, true)
	styled := visPad(newPalette(config.Default()).weight.Render("172.5"), 7, true)
	if lipgloss.Width(plain) != 7 || lipgloss.Width(styled) != 7 {
		t.Fatalf("plain %d styled %d", lipgloss.Width(plain), lipgloss.Width(styled))
	}
	if !strings.HasSuffix(visible(styled), "172.5") {
		t.Fatalf("styled = %q", styled)
	}
}

func columnStart(header, label string) int {
	return strings.Index(header, label)
}

func TestListHeaderAlignsWithWeightAndTrend(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	view := visible(app.View())
	lines := strings.Split(view, "\n")
	var header, data string
	for _, line := range lines {
		if strings.Contains(line, "weight") && strings.Contains(line, "trend") {
			header = line
		}
		if strings.Contains(line, "1990-11-02") {
			data = line
		}
	}
	if header == "" || data == "" {
		t.Fatalf("missing header or data in %q", view)
	}
	assertRightAligned(t, header, "weight", data, "171.5")
	assertRightAligned(t, header, "trend", data, "172.4")
}

func TestMonthHeaderAlignsWithWeightAndTrend(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	app := twoDayApp(t, store)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	view := visible(app.View())
	lines := strings.Split(view, "\n")
	var header, data string
	for _, line := range lines {
		if strings.Contains(line, "weight") && strings.Contains(line, "trend") {
			header = line
		}
		if strings.Contains(line, "171.5") && !strings.Contains(line, "weight") {
			data = line
		}
	}
	if header == "" || data == "" {
		t.Fatalf("missing header or data in %q", view)
	}
	assertRightAligned(t, header, "weight", data, "171.5")
	assertRightAligned(t, header, "trend", data, "172.4")
}

func TestDeltaColumnAligns(t *testing.T) {
	t.Parallel()
	app := listWithPair(t, 80.5, 80.0, config.Default())
	view := visible(app.View())
	header := headerLine(view)
	data := dataRow(view, "1990-11-02")
	if header == "" || data == "" {
		t.Fatalf("missing header or data in %q", view)
	}
	assertDeltaRightOfTrend(t, header)
	assertRightAligned(t, header, "delta", data, "+0.5")
}

func twoDayApp(t *testing.T, store *dailylog.Store) *App {
	t.Helper()
	ctx := context.Background()
	for i, w := range []float64{172.5, 171.5} {
		log, err := dailylog.New(time.Date(1990, 11, 1+i, 0, 0, 0, 0, time.UTC), &w, 8, 1000, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	return app
}

func assertRightAligned(t *testing.T, header, label, data, value string) {
	t.Helper()
	h := columnStart(header, label)
	d := strings.Index(data, value)
	if h < 0 || d < 0 {
		t.Fatalf("label %q or value %q missing\nheader %q\ndata   %q", label, value, header, data)
	}
	// Right-aligned: last character of the header label and of the value share a column.
	hEnd := h + len(label) - 1
	dEnd := d + len(value) - 1
	if hEnd != dEnd {
		t.Fatalf("%s header ends at %d, %s ends at %d\nheader %q\ndata   %q", label, hEnd, value, dEnd, header, data)
	}
}

func headerLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "weight") && strings.Contains(line, "trend") {
			return line
		}
	}
	return ""
}

func dataRow(view, needle string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, needle) && !strings.Contains(line, "weight") {
			return line
		}
	}
	return ""
}

func cursorRow(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "> ") {
			return line
		}
	}
	return ""
}

func assertDeltaRightOfTrend(t *testing.T, header string) {
	t.Helper()
	ti := strings.Index(header, "trend")
	di := strings.Index(header, "delta")
	if ti < 0 || di < 0 {
		t.Fatalf("header missing trend or delta: %q", header)
	}
	if di <= ti {
		t.Fatalf("delta is not right of trend: %q", header)
	}
}

func cellUnder(t *testing.T, header, data, label string) string {
	t.Helper()
	h := strings.Index(header, label)
	if h < 0 {
		t.Fatalf("header missing %q\nheader %q\ndata   %q", label, header, data)
	}
	start := lipgloss.Width(header[:h])
	end := start + lipgloss.Width(label)
	return strings.TrimSpace(visCut(data, start, end))
}

func visCut(s string, start, end int) string {
	var b strings.Builder
	col := 0
	for _, r := range s {
		w := lipgloss.Width(string(r))
		if col+w > end {
			break
		}
		if col >= start {
			b.WriteRune(r)
		}
		col += w
	}
	return b.String()
}
