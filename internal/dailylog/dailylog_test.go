package dailylog

import (
	"errors"
	"testing"
	"time"
)

func TestNewRejectsInvalidFields(t *testing.T) {
	t.Parallel()
	day := date(1990, 11, 4)
	w := 171.5

	cases := []struct {
		name  string
		day   time.Time
		w     *float64
		sleep float64
		steps int
		want  error
	}{
		{name: "zero day", want: ErrZeroDay},
		{name: "non-positive weight", day: day, w: ptr(0), want: ErrInvalidWeight},
		{name: "negative sleep", day: day, w: &w, sleep: -1, want: ErrNegativeSleep},
		{name: "negative steps", day: day, w: &w, steps: -1, want: ErrNegativeSteps},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := New(tc.day, tc.w, tc.sleep, tc.steps, false)
			if !errors.Is(err, tc.want) {
				t.Fatalf("New() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestNewNormalizesToUTCDate(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("west", -8*3600)
	got, err := New(time.Date(1990, 11, 4, 23, 30, 0, 0, loc), ptr(171.5), 7.5, 8000, true)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	want := time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC)
	if !got.Day.Equal(want) {
		t.Fatalf("Day = %v, want %v", got.Day, want)
	}
	if !got.Workout || got.SleepHours != 7.5 || got.Steps != 8000 {
		t.Fatalf("habit fields = %+v", got)
	}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func ptr(w float64) *float64 {
	return &w
}
