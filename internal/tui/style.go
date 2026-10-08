package tui

// Palette turns the config overlay into lipgloss styles. Package-level
// vars cannot honour a per-process scheme or NO_COLOR, so each App
// builds one palette at New. Bold and reverse stay here, not in
// config. Widget text fields call quietInput because their default
// greys are not palette roles, and a non-empty NO_COLOR still has to
// leave the model text without chromatic colour. This file does not
// parse TOML and does not choose key behaviour.

import (
	"image/color"
	"os"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/jglueckstein/hdtools/internal/config"
)

type palette struct {
	title, muted, header, help    lipgloss.Style
	weight, trend                 lipgloss.Style
	deltaPos, deltaNeg, deltaZero lipgloss.Style
	error, status                 lipgloss.Style
	label, focusLabel             lipgloss.Style
	selected                      lipgloss.Style
	selectionFG                   bool
}

var nameIndex = map[string]string{
	"black": "0", "red": "1", "green": "2", "yellow": "3",
	"blue": "4", "magenta": "5", "cyan": "6", "white": "7",
	"bright-black": "8", "bright-red": "9", "bright-green": "10",
	"bright-yellow": "11", "bright-blue": "12", "bright-magenta": "13",
	"bright-cyan": "14", "bright-white": "15",
}

var defaultColors = map[string]string{
	config.RoleTitle:     "cyan",
	config.RoleMuted:     "bright-black",
	config.RoleHeader:    "white",
	config.RoleHelp:      "bright-black",
	config.RoleWeight:    "blue",
	config.RoleTrend:     "red",
	config.RoleDeltaPos:  "yellow",
	config.RoleDeltaNeg:  "green",
	config.RoleDeltaZero: "white",
	config.RoleError:     "red",
	config.RoleStatus:    "green",
}

func newPalette(cfg config.Config) palette {
	// Render leaves full-fidelity colour in the model text. The
	// terminal writer downsamples once on the way out. A second
	// conversion here would paint the screen string twice.
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
		deltaPos:    fg(config.RoleDeltaPos, false),
		deltaNeg:    fg(config.RoleDeltaNeg, false),
		deltaZero:   fg(config.RoleDeltaZero, false),
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

// quietInput drops the widget's default foregrounds when NO_COLOR is
// set. Those defaults are 256-color greys on the placeholder and the
// blurred value, and they are painted into View before the palette
// runs. An empty NO_COLOR leaves the defaults, which is the coloured
// field the screens already had.
func quietInput(ti *textinput.Model) {
	if os.Getenv("NO_COLOR") == "" {
		return
	}
	ti.SetStyles(plainInputStyles())
}

func plainInputStyles() textinput.Styles {
	s := textinput.DefaultDarkStyles()
	plain := lipgloss.NewStyle()
	s.Focused.Text = plain
	s.Focused.Placeholder = plain
	s.Focused.Suggestion = plain
	s.Focused.Prompt = plain
	s.Blurred.Text = plain
	s.Blurred.Placeholder = plain
	s.Blurred.Suggestion = plain
	s.Blurred.Prompt = plain
	// The virtual cursor copies this colour into the model text.
	// NoColor keeps the reverse block and drops the grey.
	s.Cursor.Color = lipgloss.NoColor{}
	return s
}

func lipglossColor(canon string) color.Color {
	if len(canon) > 0 && canon[0] == '#' {
		return lipgloss.Color(canon)
	}
	if idx, ok := nameIndex[canon]; ok {
		return lipgloss.Color(idx)
	}
	return lipgloss.Color("")
}
