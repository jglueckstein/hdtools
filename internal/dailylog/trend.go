package dailylog

// The trend is derived from the weight series, never stored, so a backdated
// edit cannot leave a stale moving average. Arithmetic matches the book's
// pencil procedure (10% smoothing, round the increment to one decimal)
// rather than a higher-precision EMA that would disagree with a paper log.
//
// This file does not persist, chart, or convert units.

import (
	"fmt"
	"math"
)

// smoothingP is the Hacker's Diet "10% smoothing": each day's trend moves
// one tenth of the way from yesterday's trend toward today's weight.
// The book chose 10% so a pencil-and-paper user can shift the decimal
// instead of dividing.
const smoothingP = 0.1

// ApplyTrend fills Trend on a chronological copy of logs using the
// exponentially smoothed moving average from The Hacker's Diet:
//
//	T[n] = T[n-1] + 0.1 * (W[n] - T[n-1])
//
// with the increment rounded to one decimal place before it is added, matching
// the pencil-and-paper procedure (shift decimal, round, add).
//
// carry is yesterday's trend (the "Trend carry forward" cell). If carry is
// nil, the first logged weight becomes the first trend, as on day one of a
// brand-new logbook. Days with no weight keep the previous trend.
func ApplyTrend(logs []DailyLog, carry *float64) ([]DailyLog, error) {
	out := make([]DailyLog, len(logs))
	copy(out, logs)

	var prev *float64
	if carry != nil {
		t := round1(*carry)
		prev = &t
	}

	for i := range out {
		if err := out[i].Validate(); err != nil {
			return nil, fmt.Errorf("apply trend: day %s: %w", out[i].Day.Format("2006-01-02"), err)
		}
		if i > 0 {
			if !out[i].Day.After(out[i-1].Day) {
				if out[i].Day.Equal(out[i-1].Day) {
					return nil, fmt.Errorf("apply trend: %w", ErrDuplicateDay)
				}
				return nil, fmt.Errorf("apply trend: %w", ErrUnsortedDays)
			}
		}

		switch {
		case out[i].Weight == nil:
			if prev == nil {
				return nil, fmt.Errorf("apply trend: %w", ErrNoTrendStart)
			}
			out[i].Trend = *prev
		case prev == nil:
			t := round1(*out[i].Weight)
			out[i].Trend = t
			prev = &out[i].Trend
		default:
			out[i].Trend = nextTrend(*prev, *out[i].Weight)
			prev = &out[i].Trend
		}
	}
	return out, nil
}

// nextTrend is the book's three-step arithmetic: difference, shift/round to
// one decimal, add to yesterday's trend. Using math.Round (half away from
// zero) matches "if the second decimal place is 5 or greater".
func nextTrend(prev, weight float64) float64 {
	increment := round1(smoothingP * (weight - prev))
	return round1(prev + increment)
}

func round1(x float64) float64 {
	return math.Round(x*10) / 10
}
