package tui

import "github.com/charmbracelet/lipgloss"

// Palette is the default 16-color scheme from idea.md: cyan chrome, blue
// daily weight, red trend (the book's chart colors), reverse for
// selection. Named ANSI indexes, not truecolor, so a 16-color terminal
// still works. ">" and column headers remain so color is never the only
// cue. User-configurable schemes are specified but not loaded yet.

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")) // cyan
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // bright black
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7")) // white
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")) // red
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green
	// Daily weight is the noisy measurement (blue in the book's charts).
	weightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")) // blue
	// Trend is the signal (red and heavier in the book's charts).
	trendStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("1")) // red
	selectedStyle = lipgloss.NewStyle().
			Reverse(true)
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))
	focusLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))
)
