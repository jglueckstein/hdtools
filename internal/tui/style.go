package tui

import "github.com/charmbracelet/lipgloss"

// Palette uses the 16 ANSI colors so a basic terminal still gets structure
// and meaning. Selection also keeps ">" / "[ ]", and trend vs weight stay
// labeled in the header, so color is never the only cue.
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
