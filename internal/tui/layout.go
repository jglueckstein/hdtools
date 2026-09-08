package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	wMark    = 2
	wDate    = 10 // 2006-01-02
	wDay     = 4  // month sheet day number, matches "date"
	wWeekday = 2
	wWeight  = 7
	wTrend   = 6
	wSleep   = 5
	wSteps   = 6
	wWorkout = 7
	colGap   = "  "
)

// visPad pads s to width display cells, ignoring ANSI so colored numbers
// still sit under the header.
func visPad(s string, width int, right bool) string {
	n := width - lipgloss.Width(s)
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	if right {
		return pad + s
	}
	return s + pad
}

func joinCols(cols ...string) string {
	return strings.Join(cols, colGap)
}

func listHeader() string {
	return visPad("", wMark, false) + joinCols(
		visPad("date", wDate, false),
		visPad("weight", wWeight, true),
		visPad("trend", wTrend, true),
		visPad("sleep", wSleep, true),
		visPad("steps", wSteps, true),
		visPad("workout", wWorkout, false),
		"note",
	)
}

func monthHeader() string {
	return visPad("", wMark, false) + joinCols(
		visPad("date", wDay, true),
		visPad("wd", wWeekday, false),
		visPad("weight", wWeight, true),
		visPad("trend", wTrend, true),
		visPad("sleep", wSleep, true),
		visPad("steps", wSteps, true),
		visPad("workout", wWorkout, false),
		"note",
	)
}
