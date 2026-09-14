// Package chartspan is the single owner of the monthly chart's data
// rules: last-plotted-day clip, the empty predicate, the ±2 lb Y pad,
// and whether Monthly Loss / Daily Deficit may be shown.
//
// Both the TUI and the PDF writer call this package so a clip or
// omission rule cannot drift between screen and paper. Trend is
// already on the logs (ApplyTrend happens before Clip). This package
// does not open SQLite, paint characters or vectors, or import the
// TUI or a PDF library.
package chartspan

import (
	"fmt"
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

// Point is one day of a plotted monthly span. Weight is the daily
// mark when that day was logged; HasTrend includes carry-forward
// onto a day with no row.
type Point struct {
	Day      time.Time
	Weight   *float64
	Trend    float64
	HasTrend bool
}

// Span is day 1 through last plotted day of one calendar month.
type Span struct {
	Points []Point
}

// LastPlottedDay is the last calendar day the monthly chart may paint:
// today when year-month contains today, otherwise the month's last
// day. A future month returns 0 so trend is never plotted forward.
func LastPlottedDay(today time.Time, year int, month time.Month) int {
	ty, tm, td := today.Date()
	if year > ty || (year == ty && month > tm) {
		return 0
	}
	if year == ty && month == tm {
		return td
	}
	return daysInMonth(year, month)
}

// Clip projects a full ApplyTrend series onto day 1 through last
// plotted day, carrying trend onto blank days. It does not recompute
// trend, open SQLite, or paint.
func Clip(logs []dailylog.DailyLog, year int, month time.Month, today time.Time) Span {
	last := LastPlottedDay(today, year, month)
	if last <= 0 {
		return Span{}
	}
	byDay := make(map[string]dailylog.DailyLog, len(logs))
	for _, log := range logs {
		byDay[log.Day.Format(time.DateOnly)] = log
	}
	out := make([]Point, last)
	for i := 0; i < last; i++ {
		day := time.Date(year, month, i+1, 0, 0, 0, 0, time.UTC)
		p := Point{Day: day}
		if log, ok := byDay[day.Format(time.DateOnly)]; ok {
			p.Weight = log.Weight
			p.Trend = log.Trend
			p.HasTrend = log.Weight != nil || log.Trend != 0
		} else if t, ok := lastTrendOnOrBefore(logs, day); ok {
			p.Trend = t
			p.HasTrend = true
		}
		out[i] = p
	}
	return Span{Points: out}
}

// Empty is true when the plotted span has no daily marks and no
// trend, including no carry. A carry-only month is not empty.
func (s Span) Empty() bool {
	for _, p := range s.Points {
		if p.Weight != nil || p.HasTrend {
			return false
		}
	}
	return true
}

// Analysis is the TUI copy of Monthly Loss and Daily Deficit when
// both endpoints have trend. Otherwise both numbers are omitted.
func (s Span) Analysis(unit units.Unit) (string, bool) {
	n := len(s.Points)
	if n == 0 || !s.Points[0].HasTrend || !s.Points[n-1].HasTrend {
		return "", false
	}
	lossKG, kcal, ok := dailylog.MonthlyBalance(s.Points[0].Trend, s.Points[n-1].Trend, n)
	if !ok {
		return "", false
	}
	loss, err := units.FromKG(lossKG, unit)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("Monthly loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal), true
}

// YPad is 2 lb expressed in unit, the book's margin around min and max.
func YPad(unit units.Unit) float64 {
	kg, err := units.ToKG(2, units.Pound)
	if err != nil {
		return 0
	}
	p, err := units.FromKG(kg, unit)
	if err != nil {
		return 0
	}
	return p
}

// YRange is min−P through max+P of the plotted weights and trend, in
// unit. ok is false when the span has nothing to scale.
func YRange(s Span, unit units.Unit) (ymin, ymax float64, ok bool) {
	first := true
	add := func(kg float64) {
		v, err := units.FromKG(kg, unit)
		if err != nil {
			return
		}
		if first {
			ymin, ymax, first, ok = v, v, false, true
			return
		}
		if v < ymin {
			ymin = v
		}
		if v > ymax {
			ymax = v
		}
	}
	for _, p := range s.Points {
		if p.Weight != nil {
			add(*p.Weight)
		}
		if p.HasTrend {
			add(p.Trend)
		}
	}
	if !ok {
		return 0, 0, false
	}
	pad := YPad(unit)
	return ymin - pad, ymax + pad, true
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// lastTrendOnOrBefore matches the month sheet: a log on or before day
// with a weight or a computed trend is carry, so a blank June still
// plots May's trend.
func lastTrendOnOrBefore(logs []dailylog.DailyLog, day time.Time) (float64, bool) {
	var trend float64
	found := false
	for _, log := range logs {
		if log.Day.After(day) {
			break
		}
		if log.Weight != nil || log.Trend != 0 {
			trend = log.Trend
			found = true
		}
	}
	return trend, found
}
