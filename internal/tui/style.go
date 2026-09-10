package tui

// Palette turns the config overlay into lipgloss styles. Package-level
// vars cannot honour a per-process scheme or NO_COLOR, so each App
// builds one palette at New. Bold and reverse stay here, not in
// config. This file does not parse TOML.

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/muesli/termenv"
)

type palette struct {
	title, muted, header, help lipgloss.Style
	weight, trend              lipgloss.Style
	error, status              lipgloss.Style
	label, focusLabel          lipgloss.Style
	selected                   lipgloss.Style
	selectionFG                bool
}

var nameIndex = map[string]string{
	"black": "0", "red": "1", "green": "2", "yellow": "3",
	"blue": "4", "magenta": "5", "cyan": "6", "white": "7",
	"bright-black": "8", "bright-red": "9", "bright-green": "10",
	"bright-yellow": "11", "bright-blue": "12", "bright-magenta": "13",
	"bright-cyan": "14", "bright-white": "15",
}

var defaultColors = map[string]string{
	config.RoleTitle:  "cyan",
	config.RoleMuted:  "bright-black",
	config.RoleHeader: "white",
	config.RoleHelp:   "bright-black",
	config.RoleWeight: "blue",
	config.RoleTrend:  "red",
	config.RoleError:  "red",
	config.RoleStatus: "green",
}

func newPalette(cfg config.Config) palette {
	// Ascii emits no SGR at all, including reverse and bold. Tests and
	// non-TTY View() still need chroma (or attributes under NO_COLOR).
	if lipgloss.ColorProfile() == termenv.Ascii {
		lipgloss.SetColorProfile(termenv.ANSI)
	}
	noColor := os.Getenv("NO_COLOR") != ""
	fg := func(role string, bold bool) lipgloss.Style {
		s := lipgloss.NewStyle()
		if bold {
			s = s.Bold(true)
		}
		if noColor {
			return s
		}
		return s.Foreground(lipglossColor(roleColor(cfg, role)))
	}
	sel := lipgloss.NewStyle().Reverse(true)
	_, hasSel := cfg.Colors[config.RoleSelection]
	if !noColor && hasSel {
		sel = sel.Foreground(lipglossColor(cfg.Colors[config.RoleSelection]))
	}
	return palette{
		title:       fg(config.RoleTitle, true),
		muted:       fg(config.RoleMuted, false),
		header:      fg(config.RoleHeader, true),
		help:        fg(config.RoleHelp, false),
		weight:      fg(config.RoleWeight, false),
		trend:       fg(config.RoleTrend, true),
		error:       fg(config.RoleError, false),
		status:      fg(config.RoleStatus, false),
		label:       fg(config.RoleTitle, false),
		focusLabel:  fg(config.RoleTitle, true),
		selected:    sel,
		selectionFG: !noColor && hasSel,
	}
}

func roleColor(cfg config.Config, role string) string {
	if cfg.Colors != nil {
		if v, ok := cfg.Colors[role]; ok {
			return v
		}
	}
	return defaultColors[role]
}

func lipglossColor(canon string) lipgloss.Color {
	if len(canon) > 0 && canon[0] == '#' {
		return lipgloss.Color(canon)
	}
	if idx, ok := nameIndex[canon]; ok {
		return lipgloss.Color(idx)
	}
	return lipgloss.Color("")
}
