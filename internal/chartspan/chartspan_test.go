package chartspan

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestLastPlottedDayClipsAtToday(t *testing.T) {
	t.Parallel()
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	got := LastPlottedDay(today, 1990, time.November)
	if got != 10 {
		t.Fatalf("LastPlottedDay = %d, want 10", got)
	}
}

func TestLastPlottedDayPastMonthUsesLastCalendarDay(t *testing.T) {
	t.Parallel()
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	got := LastPlottedDay(today, 1990, time.June)
	if got != 30 {
		t.Fatalf("LastPlottedDay = %d, want 30", got)
	}
}

func TestSpanEmptyMonth(t *testing.T) {
	t.Parallel()
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	span := Clip(nil, 1990, time.June, today)
	if !span.Empty() {
		t.Fatal("June with no logs and no carry should be empty")
	}
}

func TestSpanCarryOnlyIsNotEmpty(t *testing.T) {
	t.Parallel()
	today := time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 5, 31, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	trended, err := dailylog.ApplyTrend([]dailylog.DailyLog{log}, nil)
	if err != nil {
		t.Fatal(err)
	}
	span := Clip(trended, 1990, time.June, today)
	if span.Empty() {
		t.Fatal("carry-only June treated as empty")
	}
	if len(span.Points) != 30 {
		t.Fatalf("points = %d, want 30", len(span.Points))
	}
	for i, p := range span.Points {
		if p.Weight != nil {
			t.Fatalf("day %d has a daily mark", i+1)
		}
		if !p.HasTrend {
			t.Fatalf("day %d missing carried trend", i+1)
		}
	}
}

func TestSpanOmitsAnalysisWithoutEndpoints(t *testing.T) {
	t.Parallel()
	today := time.Date(1990, 12, 1, 12, 0, 0, 0, time.UTC)
	w := 80.0
	log, err := dailylog.New(time.Date(1990, 11, 5, 0, 0, 0, 0, time.UTC), &w, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	trended, err := dailylog.ApplyTrend([]dailylog.DailyLog{log}, nil)
	if err != nil {
		t.Fatal(err)
	}
	span := Clip(trended, 1990, time.November, today)
	if len(span.Points) != 30 {
		t.Fatalf("points = %d, want 30", len(span.Points))
	}
	if span.Empty() {
		t.Fatal("November with a weigh-in should not be empty")
	}
	if span.Points[0].HasTrend {
		t.Fatal("day 1 should have no trend")
	}
	line, ok := span.Analysis(units.Kilogram)
	if ok || strings.Contains(line, "Monthly loss:") {
		t.Fatalf("analysis without both endpoints: %q", line)
	}
}

func TestSpanAnalysisBothEndpoints(t *testing.T) {
	t.Parallel()
	w1, w30 := 80.0, 79.0
	points := make([]Point, 30)
	for i := range points {
		points[i] = Point{
			Day:      time.Date(1990, 11, i+1, 0, 0, 0, 0, time.UTC),
			Trend:    80,
			HasTrend: true,
		}
	}
	points[0].Weight = &w1
	points[29].Weight = &w30
	points[29].Trend = 79
	line, ok := Span{Points: points}.Analysis(units.Kilogram)
	if !ok {
		t.Fatal("omitted analysis with both endpoints")
	}
	want := "Monthly loss: 1.0 kg   Daily deficit: 257 cal"
	if line != want {
		t.Fatalf("got %q, want %q", line, want)
	}
}

func TestYPad(t *testing.T) {
	t.Parallel()
	kg, err := units.ToKG(2, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []units.Unit{units.Kilogram, units.Pound, units.Stone} {
		want, err := units.FromKG(kg, u)
		if err != nil {
			t.Fatal(err)
		}
		got := YPad(u)
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("YPad(%s) = %v, want %v", u, got, want)
		}
	}
}

func TestSpanYRangeKG(t *testing.T) {
	t.Parallel()
	w80, w81 := 80.0, 81.0
	span := Span{Points: []Point{
		{Day: time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC), Weight: &w80, Trend: 80, HasTrend: true},
		{Day: time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), Weight: &w81, Trend: 81, HasTrend: true},
	}}
	ymin, ymax, ok := YRange(span, units.Kilogram)
	if !ok {
		t.Fatal("YRange ok = false")
	}
	pKG, err := units.ToKG(2, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	p, err := units.FromKG(pKG, units.Kilogram)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(ymin-(80-p)) > 1e-9 || math.Abs(ymax-(81+p)) > 1e-9 {
		t.Fatalf("YRange = [%.6f, %.6f], want [%.6f, %.6f]", ymin, ymax, 80-p, 81+p)
	}
}
