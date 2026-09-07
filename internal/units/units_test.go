package units

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()
	got, err := Parse("KG")
	if err != nil || got != Kilogram {
		t.Fatalf("Parse(KG) = %q, %v", got, err)
	}
	if _, err := Parse("stone"); err == nil {
		t.Fatal("Parse(stone) should fail; only st is accepted")
	}
}

func TestRoundTripPoundAndStone(t *testing.T) {
	t.Parallel()
	const kg = 80.0
	lb, err := FromKG(kg, Pound)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ToKG(lb, Pound)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(back-kg) > 1e-9 {
		t.Fatalf("lb round-trip %v -> %v", kg, back)
	}
	st, err := FromKG(kg, Stone)
	if err != nil {
		t.Fatal(err)
	}
	back, err = ToKG(st, Stone)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(back-kg) > 1e-9 {
		t.Fatalf("st round-trip %v -> %v", kg, back)
	}
}

func TestKnownPound(t *testing.T) {
	t.Parallel()
	// 1 lb is defined as 0.45359237 kg.
	kg, err := ToKG(1, Pound)
	if err != nil {
		t.Fatal(err)
	}
	if kg != 0.45359237 {
		t.Fatalf("1 lb = %v kg", kg)
	}
}
