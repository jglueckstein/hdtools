package dailylog

// The screen's two rates live here so the monthly chart and the
// long-term chart share one arithmetic owner. MonthlyBalance remains
// the pencil identity the PDF still calls. LastSeven is the current
// month's window. WindowRate turns that window into a loss and a
// calorie figure. FitRate is ordinary least squares through trend
// values whose x is the caller's calendar offset.
//
// Printed loss is the change over the days that entered the rate:
// first minus last for the window, and the daily rate times the gap
// from the first x to the last x for a fit. Calories are that daily
// rate in pounds times 3500, rounded half away from zero, with no
// second division. Display units are the caller's job.
//
// This file does not plot, persist, fit raw weights, or import a
// regression module. poundInKG matches internal/units so this
// package does not import it.

import (
	"math"
	"time"
)

const (
	poundInKG    = 0.45359237
	kcalPerPound = 3500
)

// MonthlyBalance reports loss in kilograms (first minus last) and the
// average daily calorie deficit. ok is false when days is not positive.
func MonthlyBalance(firstTrendKG, lastTrendKG float64, days int) (lossKG float64, dailyKCal int, ok bool) {
	if days <= 0 {
		return 0, 0, false
	}
	lossKG = firstTrendKG - lastTrendKG
	lb := lossKG / poundInKG
	dailyKCal = int(math.Round(lb * kcalPerPound / float64(days)))
	return lossKG, dailyKCal, true
}

// kcalFromDailyKG is 3500 times the daily rate in pounds, rounded
// half away from zero. The caller already has kilograms per day, so
// this step does not divide by the window or the span again.
func kcalFromDailyKG(dailyKG float64) int {
	return int(math.Round(dailyKG / poundInKG * float64(kcalPerPound)))
}

// LastSeven collects up to seven trend values ending on end, oldest
// first. Logs are chronological. A day with no row still counts when
// an earlier weight carries a trend onto it. The walk stops at seven
// values or at the first day with no trend, so a gap is not skipped
// in search of an earlier weigh-in. It does not recompute ApplyTrend.
func LastSeven(logs []DailyLog, end time.Time) []float64 {
	day := calendarDay(end)
	var newestFirst []float64
	for len(newestFirst) < 7 {
		trend, ok := trendOn(logs, day)
		if !ok {
			break
		}
		newestFirst = append(newestFirst, trend)
		day = day.AddDate(0, 0, -1)
	}
	for i, j := 0, len(newestFirst)-1; i < j; i, j = i+1, j-1 {
		newestFirst[i], newestFirst[j] = newestFirst[j], newestFirst[i]
	}
	return newestFirst
}

// trendOn is the month sheet's carry: the last log on or before day
// with a weight or a computed trend. Dates are calendar days, so a
// noon timestamp and midnight on that day are the same day.
func trendOn(logs []DailyLog, day time.Time) (float64, bool) {
	day = calendarDay(day)
	var trend float64
	found := false
	for _, log := range logs {
		if calendarDay(log.Day).After(day) {
			break
		}
		if log.Weight != nil || log.Trend != 0 {
			trend = log.Trend
			found = true
		}
	}
	return trend, found
}

// WindowRate is the current month's pencil difference. The printed
// loss is first minus last. The daily rate divides by the calendar
// gap, one less than the count, because the window is consecutive.
// ok is false when fewer than two values are available.
func WindowRate(trends []float64) (lossKG float64, kcal int, ok bool) {
	n := len(trends)
	if n < 2 {
		return 0, 0, false
	}
	lossKG = trends[0] - trends[n-1]
	daily := lossKG / float64(n-1)
	return lossKG, kcalFromDailyKG(daily), true
}

// FitRate fits a line through trendKG at the given calendar offsets.
// The points are not renumbered: a missing day stays a hole in x.
// The printed loss is the daily rate times the gap from the first
// offset to the last, so days with no trend are not part of the
// change. ok is false when fewer than two points are available, when
// that gap is not positive, or when the fit denominator is zero.
func FitRate(x, trendKG []float64) (lossKG float64, kcal int, ok bool) {
	n := len(x)
	if n < 2 || n != len(trendKG) {
		return 0, 0, false
	}
	gap := x[n-1] - x[0]
	if gap <= 0 {
		return 0, 0, false
	}
	var sumX, sumY, sumX2, sumXY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += trendKG[i]
		sumX2 += x[i] * x[i]
		sumXY += x[i] * trendKG[i]
	}
	nf := float64(n)
	den := nf*sumX2 - sumX*sumX
	if den == 0 {
		return 0, 0, false
	}
	daily := -(nf*sumXY - sumX*sumY) / den
	lossKG = daily * gap
	// A flat series has an exact +0 numerator. Negating it stores a
	// negative zero, and %.1f prints that as -0.0.
	if lossKG == 0 {
		lossKG = 0
	}
	return lossKG, kcalFromDailyKG(daily), true
}
