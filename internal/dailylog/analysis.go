package dailylog

// MonthlyBalance is the book's pencil identity: first trend minus last,
// converted to pounds, times 3500 kcal, divided by days in the plotted
// span. This file does not plot, persist, or convert display units.
//
// poundInKG matches internal/units so this package does not import it.

import "math"

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
