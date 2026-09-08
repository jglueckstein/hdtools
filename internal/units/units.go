// Package units converts body weight between the canonical store unit (kilograms)
// and the units a person may want on screen: kg, lb, and st.
//
// Meal-planning grams are a different problem and do not belong here. The
// international pound (0.45359237 kg) and 14 lb per stone match the usual
// avoirdupois values The Hacker's Diet worksheets assume.
package units

import (
	"fmt"
	"strings"
)

// Unit is a display unit for body weight. Values in the database are always
// kilograms; Unit only affects what the TUI shows and what the form accepts.
type Unit string

const (
	// Kilogram is the store unit and the first-run display default.
	Kilogram Unit = "kg"
	// Pound is the avoirdupois pound.
	Pound Unit = "lb"
	// Stone is 14 avoirdupois pounds, entered and shown as a single decimal.
	Stone Unit = "st"

	poundInKG      = 0.45359237
	poundsPerStone = 14
)

// Parse accepts "kg", "lb", or "st" (any case). Anything else is an error so
// a typo in config.toml cannot silently store pounds as kilograms.
func Parse(s string) (Unit, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "kg":
		return Kilogram, nil
	case "lb":
		return Pound, nil
	case "st":
		return Stone, nil
	default:
		return "", fmt.Errorf("weight unit %q: want kg, lb, or st", s)
	}
}

// ToKG is the only conversion allowed on the save path. Display-unit values
// must never be written to SQLite, or a later unit change would rewrite
// history.
func ToKG(value float64, u Unit) (float64, error) {
	switch u {
	case Kilogram:
		return value, nil
	case Pound:
		return value * poundInKG, nil
	case Stone:
		return value * poundsPerStone * poundInKG, nil
	default:
		return 0, fmt.Errorf("weight unit %q: want kg, lb, or st", u)
	}
}

// FromKG is display-only. Using it before Upsert would store pounds as if
// they were kilograms.
func FromKG(kg float64, u Unit) (float64, error) {
	switch u {
	case Kilogram:
		return kg, nil
	case Pound:
		return kg / poundInKG, nil
	case Stone:
		return kg / (poundsPerStone * poundInKG), nil
	default:
		return 0, fmt.Errorf("weight unit %q: want kg, lb, or st", u)
	}
}
