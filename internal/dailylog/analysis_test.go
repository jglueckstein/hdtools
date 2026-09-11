package dailylog

import "testing"

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
