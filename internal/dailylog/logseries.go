package dailylog

// The log PDF and the month-sheet key share one preparation of the
// series. ApplyTrend refuses a leading weightless row, and the TUI
// treats that as a load error. The page does not: order is checked
// first, then the trend starts at the first weigh-in. An empty log
// and a month with no weigh-in still print.
//
// This file does not change ApplyTrend, store a trend, or write a
// PDF. ErrNoTrendStart is not a failure of this preparation.

import (
	"errors"
	"fmt"
)

// RejectUnsorted fails a duplicate day and a series that goes
// backwards, including when no row has a weigh-in. ApplyTrend is
// skipped in that case, so it would not see the duplicate.
func RejectUnsorted(logs []DailyLog) error {
	for i := 1; i < len(logs); i++ {
		if logs[i].Day.After(logs[i-1].Day) {
			continue
		}
		if logs[i].Day.Equal(logs[i-1].Day) {
			return ErrDuplicateDay
		}
		return ErrUnsortedDays
	}
	return nil
}

// TrendFromFirstWeighIn leaves leading weightless rows untrended.
// ErrNoTrendStart is not an export failure: the page still prints.
func TrendFromFirstWeighIn(logs []DailyLog) ([]DailyLog, error) {
	start := -1
	for i := range logs {
		if logs[i].Weight != nil {
			start = i
			break
		}
	}
	if start < 0 {
		return logs, nil
	}
	trended, err := ApplyTrend(logs[start:], nil)
	if err != nil {
		if errors.Is(err, ErrNoTrendStart) {
			return logs, nil
		}
		return nil, fmt.Errorf("trend suffix: %w", err)
	}
	if start == 0 {
		return trended, nil
	}
	out := make([]DailyLog, 0, len(logs))
	out = append(out, logs[:start]...)
	out = append(out, trended...)
	return out, nil
}
