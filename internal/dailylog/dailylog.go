// Package dailylog is the daily weight-and-habit record from The Hacker's Diet.
//
// A day's entry is the source of truth for the scale reading and the extra
// fields this project tracks (sleep, steps, workout). The trend number is not
// stored here as an independent fact: it is derived from the weight series so
// that editing an earlier day cannot leave a stale moving average behind.
//
// This package does not own the TUI, monthly charts, or calorie accounting.
// Those read DailyLog values; they do not invent a second log format.
package dailylog

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for validation. Callers compare with errors.Is.
var (
	ErrZeroDay       = errors.New("day is the zero time")
	ErrInvalidWeight = errors.New("weight must be greater than zero when logged")
	ErrNegativeSleep = errors.New("sleep hours must not be negative")
	ErrNegativeSteps = errors.New("steps must not be negative")
	ErrUnsortedDays  = errors.New("daily logs must be in chronological order")
	ErrDuplicateDay  = errors.New("duplicate day in daily log series")
	ErrNoTrendStart  = errors.New("no weight and no carry-forward to start the trend")
	// ErrNotFound is Store.Get when that calendar day has no row. The TUI
	// compares with errors.Is so it never imports database/sql.
	ErrNotFound = errors.New("daily log not found")
)

// DailyLog is one calendar day's log line: the paper sheet's Weight column,
// plus the extra columns this project records by default.
//
// Weight is optional because the book treats travel, forgotten weigh-ins, and
// other gaps as a non-numeric note rather than a zero. A missing weight does
// not move the trend; it carries yesterday's trend forward.
//
// Trend is filled by ApplyTrend. A zero Trend on a value that has never been
// through ApplyTrend means "not computed yet", not "the person weighs nothing".
type DailyLog struct {
	Day        time.Time
	Weight     *float64
	Trend      float64
	SleepHours float64
	Steps      int
	Workout    bool
	Note       string
}

// New is the only constructor so every write path (form, month cell, tests)
// shares calendar-day normalisation and validation. Callers must not
// assemble DailyLog literals and Upsert them without going through New,
// except the empty-day placeholder the month sheet uses before patchCell.
func New(day time.Time, weight *float64, sleepHours float64, steps int, workout bool, note string) (DailyLog, error) {
	log := DailyLog{
		Day:        calendarDay(day),
		Weight:     weight,
		SleepHours: sleepHours,
		Steps:      steps,
		Workout:    workout,
		Note:       note,
	}
	if err := log.Validate(); err != nil {
		return DailyLog{}, fmt.Errorf("new daily log: %w", err)
	}
	return log, nil
}

// Validate is the gate for both New and Upsert so a store write cannot
// bypass the same rules the form uses.
func (d DailyLog) Validate() error {
	if d.Day.IsZero() {
		return ErrZeroDay
	}
	if d.Weight != nil && *d.Weight <= 0 {
		return ErrInvalidWeight
	}
	if d.SleepHours < 0 {
		return ErrNegativeSleep
	}
	if d.Steps < 0 {
		return ErrNegativeSteps
	}
	return nil
}

func calendarDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
