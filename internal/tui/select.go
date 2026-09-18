package tui

// Closest-to-today list selection: which row commands act on after load.
// First load and out-of-range reload share this scan so the monthly chart
// follows the same row. Calendar dates only; ties prefer on-or-before
// today. This file does not create a today row and does not bind Goto
// Today.

import (
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"
)

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func absDays(a, b time.Time) int {
	n := int(dateOnly(a).Sub(dateOnly(b)).Hours() / 24)
	if n < 0 {
		return -n
	}
	return n
}

// closestLogIndex is the row with the smallest calendar-day distance
// to today. Equal distance prefers a day that is not after today; if
// both are after today, the earlier of those.
func closestLogIndex(logs []dailylog.DailyLog, today time.Time) int {
	if len(logs) == 0 {
		return 0
	}
	today = dateOnly(today)
	best := 0
	for i := 1; i < len(logs); i++ {
		if closerDay(logs[i].Day, logs[best].Day, today) {
			best = i
		}
	}
	return best
}

func closerDay(cand, best, today time.Time) bool {
	cd := absDays(cand, today)
	bd := absDays(best, today)
	if cd != bd {
		return cd < bd
	}
	candOK := !dateOnly(cand).After(today)
	bestOK := !dateOnly(best).After(today)
	if candOK != bestOK {
		return candOK
	}
	return dateOnly(cand).Before(dateOnly(best))
}

func indexOfDay(logs []dailylog.DailyLog, day time.Time) int {
	day = dateOnly(day)
	for i, log := range logs {
		if dateOnly(log.Day).Equal(day) {
			return i
		}
	}
	return -1
}
