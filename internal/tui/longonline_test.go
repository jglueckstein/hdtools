package tui

// These tests are the long-term Online picture. They read the same
// 67-column frame as the monthly chart: a Braille trend, a thin
// weight line, a plain window name, and one-line month labels.
// They do not open a terminal and they do not score the fit.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/ntcharts/v2/canvas/runes"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
)

// assertLongPicture requires both series on the monthly canvas.
// A line rune is the weight. Braille is the trend, and it stays
// bold. Anything else in a plot cell is the old mark or stem.
func assertLongPicture(t *testing.T, view string) onlinePlot {
	t.Helper()
	plot := mustOnlinePlot(t, view)
	if plot.slotHasDiamond(0, onlinePlotCols) || strings.ContainsRune(visible(view), filledDiamond) {
		t.Fatal("long-term chart has a diamond")
	}
	line := false
	boldTrend := false
	for _, row := range plot.rows {
		for _, c := range row {
			switch {
			case c.r == ' ' || c.r == 0:
				continue
			case isBraille(c.r):
				if !c.bold {
					t.Fatal("braille trend is not bold")
				}
				boldTrend = true
			case runes.IsLine(c.r):
				if c.bold {
					t.Fatal("weight line is bold")
				}
				line = true
			default:
				t.Fatalf("plot rune %q, want a weight line or braille", string(c.r))
			}
		}
	}
	if !line {
		t.Fatal("weight line missing")
	}
	if !boldTrend {
		t.Fatal("trend is not braille")
	}
	return plot
}

func TestLongChartQuarterlyIsOnlinePicture(t *testing.T) {
	enableChroma(t)
	freezeToday(t, nov1990())
	store := openStore(t)
	app := twoDayApp(t, store)
	press(app, "l")
	view := app.View().Content
	assertHeadingShape(t, view, "Quarterly")
	vis := visible(view)
	if strings.Contains(vis, "September") || strings.Contains(vis, "November 1990") {
		t.Fatalf("heading still names the span: %q", vis)
	}
	for _, want := range []string{"Sep 90", "Oct 90", "Nov 90"} {
		if !strings.Contains(vis, want) {
			t.Fatalf("missing %q in %q", want, vis)
		}
	}
	assertLongPicture(t, view)
	// 172.5 kg and 171.5 kg, one step past the plotted values.
	assertTickLabels(t, view, []string{"173", "172", "171"})
}

func TestLongChartSingleWeighInDrawsLine(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 1, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 80)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := quarterlyView(t, app)
	if strings.Contains(visible(view), "Loss:") {
		t.Fatalf("caption on one trend value: %q", visible(view))
	}
	plot := mustOnlinePlot(t, view)
	if !plotHasLine(plot) {
		t.Fatal("single weigh-in drew no weight line")
	}
	if plot.slotHasDiamond(0, onlinePlotCols) {
		t.Fatal("single weigh-in is a diamond")
	}
}

func plotHasLine(plot onlinePlot) bool {
	for _, row := range plot.rows {
		for _, c := range row {
			if runes.IsLine(c.r) {
				return true
			}
		}
	}
	return false
}

func TestLongChartFutureWeighInStaysOut(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	upsertKG(t, store, time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), 172.5)
	upsertKG(t, store, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), 171.5)
	upsertKG(t, store, time.Date(1990, 11, 20, 0, 0, 0, 0, time.UTC), 50)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "l")
	view := app.View().Content
	assertTickLabels(t, view, []string{"173", "172", "171"})
	if strings.Contains(visible(view), "Dec ") {
		t.Fatalf("window passed today: %q", visible(view))
	}
}

func TestLongChartCompleteSkipsOverlappingMonths(t *testing.T) {
	freezeToday(t, nov1990())
	store := openStore(t)
	w := 80.0
	first, err := dailylog.New(time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	app := twoDayApp(t, store)
	press(app, "l", "]", "]", "]")
	view := visible(app.View().Content)
	assertHeadingShape(t, app.View().Content, "Complete")
	if !strings.Contains(view, "Jan 88") || !strings.Contains(view, "Nov 90") {
		t.Fatalf("want the end months: %q", view)
	}
	if strings.Contains(view, "Feb 88") {
		t.Fatalf("February overlaps January: %q", view)
	}
}
