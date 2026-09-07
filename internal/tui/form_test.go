package tui

import (
	"math"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/units"
)

func TestParseFormStoresPoundsAsKilograms(t *testing.T) {
	t.Parallel()
	log, err := parseForm("1990-11-04", "171.5", "7.5", "1000", "Travel", true, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	if log.Weight == nil {
		t.Fatal("weight missing")
	}
	want, err := units.ToKG(171.5, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(*log.Weight-want) > 1e-9 {
		t.Fatalf("stored kg = %v, want %v", *log.Weight, want)
	}
	if log.Note != "Travel" || !log.Workout || log.Steps != 1000 {
		t.Fatalf("log = %+v", log)
	}
	if !log.Day.Equal(time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("day = %v", log.Day)
	}
}

func TestParseFormEmptyWeightAllowed(t *testing.T) {
	t.Parallel()
	log, err := parseForm("1989-04-10", "", "", "", "Travel", false, units.Kilogram)
	if err != nil {
		t.Fatal(err)
	}
	if log.Weight != nil {
		t.Fatalf("weight = %v", *log.Weight)
	}
	if log.Note != "Travel" {
		t.Fatalf("note = %q", log.Note)
	}
}

func TestParseFormRejectsBadDate(t *testing.T) {
	t.Parallel()
	if _, err := parseForm("11/4/1990", "80", "", "", "", false, units.Kilogram); err == nil {
		t.Fatal("expected date error")
	}
}
