package dailylog

import (
	"fmt"
	"testing"
	"time"
)

func TestMonthlyBalanceLossAndDeficit(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := MonthlyBalance(80, 79, 30)
	if !ok {
		t.Fatal("ok = false")
	}
	if loss != 1 {
		t.Fatalf("loss = %v, want 1", loss)
	}
	if kcal != 257 {
		t.Fatalf("kcal = %d, want 257", kcal)
	}
}

func TestMonthlyBalanceGainIsNegative(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := MonthlyBalance(79, 80, 30)
	if !ok {
		t.Fatal("ok = false")
	}
	if loss != -1 {
		t.Fatalf("loss = %v, want -1", loss)
	}
	if kcal != -257 {
		t.Fatalf("kcal = %d, want -257", kcal)
	}
}

func TestMonthlyBalanceElapsedDays(t *testing.T) {
	t.Parallel()
	_, kcal, ok := MonthlyBalance(80, 79, 10)
	if !ok {
		t.Fatal("ok = false")
	}
	if kcal != 772 {
		t.Fatalf("kcal = %d, want 772", kcal)
	}
}

func TestMonthlyBalanceRejectsNonPositiveDays(t *testing.T) {
	t.Parallel()
	if _, _, ok := MonthlyBalance(80, 79, 0); ok {
		t.Fatal("days=0 should not be ok")
	}
}

func utc(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func weighed(t *testing.T, day time.Time, kg float64) DailyLog {
	t.Helper()
	log, err := New(day, &kg, 8, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	return log
}

func trended(t *testing.T, logs ...DailyLog) []DailyLog {
	t.Helper()
	out, err := ApplyTrend(logs, nil)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func wantRate(t *testing.T, loss float64, kcal int, ok bool, lossText string, wantKcal int) {
	t.Helper()
	if !ok {
		t.Fatal("ok = false")
	}
	if got := fmt.Sprintf("%.1f", loss); got != lossText {
		t.Fatalf("loss = %s, want %s", got, lossText)
	}
	if kcal != wantKcal {
		t.Fatalf("kcal = %d, want %d", kcal, wantKcal)
	}
}

func TestLastSevenFixtureA(t *testing.T) {
	t.Parallel()
	logs := trended(t,
		weighed(t, utc(1990, time.November, 1), 80),
		weighed(t, utc(1990, time.November, 2), 70),
	)
	got := LastSeven(logs, utc(1990, time.November, 10))
	if len(got) != 7 {
		t.Fatalf("len = %d, want the seven blank days 4 through 10, not day 1's 80: %v", len(got), got)
	}
	for _, v := range got {
		if v != 79 {
			t.Fatalf("trends = %v, want seven 79s", got)
		}
	}
}

func TestWindowRateFixtureA(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := WindowRate([]float64{79, 79, 79, 79, 79, 79, 79})
	wantRate(t, loss, kcal, ok, "0.0", 0)
	if kcal == 772 {
		t.Fatal("kcal used the pencil 772")
	}
}

func TestWindowRateFixtureB(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := WindowRate([]float64{80, 79, 79})
	wantRate(t, loss, kcal, ok, "1.0", 3858)
	if kcal == 257 || kcal == 2572 {
		t.Fatalf("kcal = %d, want 3858", kcal)
	}
}

func TestWindowRateGain(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := WindowRate([]float64{70, 71, 71})
	wantRate(t, loss, kcal, ok, "-1.0", -3858)
}

func TestLastSevenCarry(t *testing.T) {
	t.Parallel()
	logs := trended(t,
		weighed(t, utc(1990, time.October, 30), 80),
		weighed(t, utc(1990, time.October, 31), 70),
	)
	got := LastSeven(logs, utc(1990, time.November, 4))
	want := []float64{80, 79, 79, 79, 79, 79}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want six values and no seventh: %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("trends = %v, want %v", got, want)
		}
	}
}

func TestWindowRateCarry(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := WindowRate([]float64{80, 79, 79, 79, 79, 79})
	wantRate(t, loss, kcal, ok, "1.0", 1543)
	if fmt.Sprintf("%.1f", loss) == "0.7" || kcal == 1286 {
		t.Fatalf("loss %v kcal %d, want 1.0 kg and 1543 cal", loss, kcal)
	}
}

func TestFitRateFixtureC(t *testing.T) {
	t.Parallel()
	x := make([]float64, 30)
	y := make([]float64, 30)
	y[0] = 80
	for i := 1; i < 30; i++ {
		x[i] = float64(i)
		y[i] = 79
	}
	loss, kcal, ok := FitRate(x, y)
	if !ok {
		t.Fatal("ok = false")
	}
	// 0.1935… is the 30-day product. The gap from day 1 to day 30 is 29.
	if loss >= 0.19 {
		t.Fatalf("loss = %v, want the 29-day gap below 0.19", loss)
	}
	wantRate(t, loss, kcal, ok, "0.2", 50)
	if kcal == 257 {
		t.Fatal("kcal used the pencil 257")
	}
}

func TestFitRateLateStart(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := FitRate([]float64{28, 29}, []float64{80, 79})
	wantRate(t, loss, kcal, ok, "1.0", 7716)
	if fmt.Sprintf("%.1f", loss) == "30.0" {
		t.Fatalf("loss spread across 30 days: %v", loss)
	}
}

func TestLastSevenLateStart(t *testing.T) {
	t.Parallel()
	logs := trended(t,
		weighed(t, utc(1990, time.November, 29), 80),
		weighed(t, utc(1990, time.November, 30), 70),
	)
	got := LastSeven(logs, utc(1990, time.November, 30))
	want := []float64{80, 79}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want the two days and a stop before day 28: %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("trends = %v, want %v", got, want)
		}
	}
}

func TestWindowRateConsecutivePair(t *testing.T) {
	t.Parallel()
	loss, kcal, ok := WindowRate([]float64{80, 79})
	wantRate(t, loss, kcal, ok, "1.0", 7716)
	if kcal == 3858 {
		t.Fatal("divisor was the count of two values")
	}
}

func TestFitRateQuarterlySpan(t *testing.T) {
	t.Parallel()
	x := make([]float64, 10)
	y := make([]float64, 10)
	y[0] = 80
	for i := 1; i < 10; i++ {
		x[i] = float64(i)
		y[i] = 79
	}
	loss, kcal, ok := FitRate(x, y)
	if !ok {
		t.Fatal("ok = false")
	}
	// Ten trend days print 0.545 kg. Seventy-one window days print 3.9 kg.
	if loss >= 0.5 {
		t.Fatalf("loss = %v, want the nine-day gap below 0.5", loss)
	}
	wantRate(t, loss, kcal, ok, "0.5", 421)
	if fmt.Sprintf("%.1f", loss) == "3.9" {
		t.Fatal("loss used the 71-day window")
	}
}

func TestWindowRateOmitsFewerThanTwo(t *testing.T) {
	t.Parallel()
	if _, _, ok := WindowRate([]float64{79}); ok {
		t.Fatal("one trend value should not be ok")
	}
}

func TestFitRateOmitsFewerThanTwo(t *testing.T) {
	t.Parallel()
	if _, _, ok := FitRate([]float64{0}, []float64{80}); ok {
		t.Fatal("one point should not be ok")
	}
	if _, _, ok := FitRate([]float64{4, 4}, []float64{80, 79}); ok {
		t.Fatal("zero gap should not be ok")
	}
	if _, _, ok := FitRate([]float64{2, 2, 2}, []float64{80, 79, 79}); ok {
		t.Fatal("zero denominator should not be ok")
	}
}

func TestFitRateUnroundedLoss(t *testing.T) {
	t.Parallel()
	x := make([]float64, 30)
	y := make([]float64, 30)
	for i := 0; i < 30; i++ {
		x[i] = float64(i)
		y[i] = 69
		if i < 14 {
			y[i] = 70
		}
	}
	loss, kcal, ok := FitRate(x, y)
	if !ok {
		t.Fatal("ok = false")
	}
	if loss <= 1.44 {
		t.Fatalf("loss = %v, want the unrounded value above 1.44", loss)
	}
	wantRate(t, loss, kcal, ok, "1.4", 385)
	if fmt.Sprintf("%.1f", loss) == "1.5" {
		t.Fatal("loss used a 30-day product")
	}
}
