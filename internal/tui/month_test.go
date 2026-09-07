package tui

import (
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestBuildMonthSheetFillsBlankDaysAndCarriesTrend(t *testing.T) {
	t.Parallel()
	carry := 173.6
	logs := []dailylog.DailyLog{
		{Day: dateUTC(1990, 11, 1), Weight: ptrFloat(172.5)},
		{Day: dateUTC(1990, 11, 2), Weight: ptrFloat(171.5)},
		{Day: dateUTC(1990, 11, 4), Weight: ptrFloat(171.5)},
	}
	trended, err := dailylog.ApplyTrend(logs, &carry)
	if err != nil {
		t.Fatal(err)
	}
	sheet := buildMonthSheet(trended, 1990, time.November)
	if len(sheet) != 30 {
		t.Fatalf("len = %d, want 30", len(sheet))
	}
	if !sheet[0].HasEntry || sheet[2].HasEntry {
		t.Fatalf("day 1 has entry=%v day 3 has entry=%v", sheet[0].HasEntry, sheet[2].HasEntry)
	}
	if !sheet[2].HasTrend || sheet[2].Trend != trended[1].Trend {
		t.Fatalf("day 3 trend = %v want carry %v", sheet[2].Trend, trended[1].Trend)
	}
}

func TestDaysInMonthFebruaryLeap(t *testing.T) {
	t.Parallel()
	if daysInMonth(1992, time.February) != 29 {
		t.Fatal("1992 should be a leap February")
	}
	if daysInMonth(1990, time.February) != 28 {
		t.Fatal("1990 February has 28 days")
	}
}

func TestMonthNavigationClampsDay(t *testing.T) {
	t.Parallel()
	m := newMonth(dateUTC(1990, 3, 31))
	m.prevMonth()
	if m.month != time.February || m.day != 28 {
		t.Fatalf("got %s %d", m.month, m.day)
	}
}

func TestPatchCellConvertsPounds(t *testing.T) {
	t.Parallel()
	log := dailylog.DailyLog{Day: dateUTC(1990, 11, 4)}
	got, err := patchCell(log, colWeight, "171.5", units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	want, err := units.ToKG(171.5, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != want {
		t.Fatalf("weight = %v want %v", got.Weight, want)
	}
}

func TestPatchCellClearsWeight(t *testing.T) {
	t.Parallel()
	w := 80.0
	log := dailylog.DailyLog{Day: dateUTC(1990, 11, 4), Weight: &w}
	got, err := patchCell(log, colWeight, "", units.Kilogram)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight != nil {
		t.Fatalf("weight = %v, want nil", *got.Weight)
	}
}

func dateUTC(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func ptrFloat(w float64) *float64 {
	return &w
}
