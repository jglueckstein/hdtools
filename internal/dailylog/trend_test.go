package dailylog

import (
	"errors"
	"testing"
)

func TestApplyTrendMatchesPencilExample(t *testing.T) {
	t.Parallel()
	// Pencil and Paper, November 1990: carry 173.6, then 172.5, 171.5, 172, 171.5.
	// The book publishes 173.5, 173.3, 173.2, 173.0 as the trend column.
	carry := 173.6
	logs := []DailyLog{
		{Day: date(1990, 11, 1), Weight: ptr(172.5)},
		{Day: date(1990, 11, 2), Weight: ptr(171.5)},
		{Day: date(1990, 11, 3), Weight: ptr(172)},
		{Day: date(1990, 11, 4), Weight: ptr(171.5)},
	}
	got, err := ApplyTrend(logs, &carry)
	if err != nil {
		t.Fatalf("ApplyTrend() unexpected error: %v", err)
	}
	want := []float64{173.5, 173.3, 173.2, 173.0}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Trend != want[i] {
			t.Errorf("day %s trend = %v, want %v", got[i].Day.Format("2006-01-02"), got[i].Trend, want[i])
		}
	}
}

func TestApplyTrendFirstDayCopiesWeight(t *testing.T) {
	t.Parallel()
	logs := []DailyLog{{Day: date(1989, 4, 1), Weight: ptr(145.5)}}
	got, err := ApplyTrend(logs, nil)
	if err != nil {
		t.Fatalf("ApplyTrend() unexpected error: %v", err)
	}
	if got[0].Trend != 145.5 {
		t.Fatalf("Trend = %v, want 145.5", got[0].Trend)
	}
}

func TestApplyTrendMissingWeightCarriesForward(t *testing.T) {
	t.Parallel()
	logs := []DailyLog{
		{Day: date(1989, 4, 9), Weight: ptr(145.5)},
		{Day: date(1989, 4, 10), Weight: nil}, // travel
		{Day: date(1989, 4, 11), Weight: nil},
		{Day: date(1989, 4, 12), Weight: ptr(145.5)},
	}
	got, err := ApplyTrend(logs, nil)
	if err != nil {
		t.Fatalf("ApplyTrend() unexpected error: %v", err)
	}
	if got[1].Trend != 145.5 || got[2].Trend != 145.5 {
		t.Fatalf("travel days should keep 145.5, got %v then %v", got[1].Trend, got[2].Trend)
	}
	if got[3].Trend != 145.5 {
		t.Fatalf("day after travel trend = %v, want 145.5 (weight equalled carried trend)", got[3].Trend)
	}
}

func TestApplyTrendRejectsBadSeries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		logs  []DailyLog
		carry *float64
		want  error
	}{
		{
			name: "unsorted",
			logs: []DailyLog{
				{Day: date(1990, 11, 2), Weight: ptr(171.5)},
				{Day: date(1990, 11, 1), Weight: ptr(172.5)},
			},
			want: ErrUnsortedDays,
		},
		{
			name: "duplicate",
			logs: []DailyLog{
				{Day: date(1990, 11, 1), Weight: ptr(172.5)},
				{Day: date(1990, 11, 1), Weight: ptr(171.5)},
			},
			want: ErrDuplicateDay,
		},
		{
			name: "no start",
			logs: []DailyLog{{Day: date(1990, 11, 1)}},
			want: ErrNoTrendStart,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ApplyTrend(tc.logs, tc.carry)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}
