package tui

// formatDelta is the list and month residual: the two on-screen cells
// subtracted, not a kilogram residual converted afterward. Round to one
// decimal then pick the sign so +0.0 never appears. Empty when there is
// no weigh-in or no trend. This file does not paint or store a column.

import (
	"fmt"
	"strconv"

	"github.com/jglueckstein/hdtools/internal/units"
)

func paint1(x float64) float64 {
	s := fmt.Sprintf("%.1f", x)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return x
	}
	return v
}

func formatDelta(weight *float64, trend float64, showTrend bool, unit units.Unit) string {
	if weight == nil || !showTrend {
		return ""
	}
	dw, errW := units.FromKG(*weight, unit)
	dt, errT := units.FromKG(trend, unit)
	if errW != nil || errT != nil {
		return ""
	}
	// Same one-decimal as fmt.Sprintf("%.1f") on the weight and trend
	// cells (half-even), not math.Round, so the three numbers add.
	s := fmt.Sprintf("%.1f", paint1(dw)-paint1(dt))
	if s == "0.0" || s == "-0.0" {
		return "0.0"
	}
	if paint1(dw) > paint1(dt) {
		return "+" + s
	}
	return s
}

func styleDelta(s string, p palette) string {
	if s == "" {
		return visPad("", wDelta, true)
	}
	st := p.deltaZero
	switch {
	case s[0] == '+':
		st = p.deltaPos
	case s[0] == '-':
		st = p.deltaNeg
	}
	return visPad(st.Render(s), wDelta, true)
}
