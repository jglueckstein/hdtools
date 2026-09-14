package chartpdf

// This file maps [colors] role values to print RGB so a PDF matches
// Excel, not whatever the current TTY mapped "blue" to. Hex is the
// bytes as written; named 16-color values (and 0–15 aliases) use a
// fixed VGA table. Title-box chrome and stem green stay as Excel
// constants in chartpdf.go. This file does not paint, open SQLite,
// or honour NO_COLOR.

import (
	"strconv"
	"strings"
	"unicode"
)

// VGA RGB for the TUI's 16-color names (and 0–15 aliases). Paper uses
// this table rather than the current TTY palette so a print matches
// Excel, not whatever the terminal happened to map "blue" to.
var vgaRGB = [16][3]int{
	{0x00, 0x00, 0x00}, // black
	{0x80, 0x00, 0x00}, // red
	{0x00, 0x80, 0x00}, // green
	{0x80, 0x80, 0x00}, // yellow
	{0x00, 0x00, 0x80}, // blue
	{0x80, 0x00, 0x80}, // magenta
	{0x00, 0x80, 0x80}, // cyan
	{0xC0, 0xC0, 0xC0}, // white
	{0x80, 0x80, 0x80}, // bright-black
	{0xFF, 0x00, 0x00}, // bright-red
	{0x00, 0xFF, 0x00}, // bright-green
	{0xFF, 0xFF, 0x00}, // bright-yellow
	{0x00, 0x00, 0xFF}, // bright-blue
	{0xFF, 0x00, 0xFF}, // bright-magenta
	{0x00, 0xFF, 0xFF}, // bright-cyan
	{0xFF, 0xFF, 0xFF}, // bright-white
}

var vgaNames = map[string]int{
	"black": 0, "red": 1, "green": 2, "yellow": 3,
	"blue": 4, "magenta": 5, "cyan": 6, "white": 7,
	"bright-black": 8, "bright-red": 9, "bright-green": 10,
	"bright-yellow": 11, "bright-blue": 12, "bright-magenta": 13,
	"bright-cyan": 14, "bright-white": 15,
}

// ColorToRGB turns a [colors] value into print RGB. Hex is the bytes
// as written (#rgb doubles each digit). Named 16-color values (and
// 0–15 aliases) use the VGA table from the spec, not a terminal
// palette, so paper matches Excel rather than the current TTY.
// Title chrome and stem green are fixed Excel RGB, not [colors] keys.
func ColorToRGB(s string) (r, g, b int, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, 0, false
	}
	if strings.HasPrefix(s, "#") {
		return parseHexRGB(s)
	}
	if n, err := strconv.Atoi(s); err == nil {
		if n < 0 || n > 15 {
			return 0, 0, 0, false
		}
		c := vgaRGB[n]
		return c[0], c[1], c[2], true
	}
	idx, ok := vgaNames[normalizeColorName(s)]
	if !ok {
		return 0, 0, 0, false
	}
	c := vgaRGB[idx]
	return c[0], c[1], c[2], true
}

func parseHexRGB(s string) (r, g, b int, ok bool) {
	h := strings.ToLower(s[1:])
	switch len(h) {
	case 3:
		if !isHex(h) {
			return 0, 0, 0, false
		}
		return hexByte(h[0], h[0]), hexByte(h[1], h[1]), hexByte(h[2], h[2]), true
	case 6:
		if !isHex(h) {
			return 0, 0, 0, false
		}
		return hexByte(h[0], h[1]), hexByte(h[2], h[3]), hexByte(h[4], h[5]), true
	default:
		return 0, 0, 0, false
	}
}

func hexByte(a, b byte) int {
	return hexNibble(a)<<4 | hexNibble(b)
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	default:
		return 0
	}
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func normalizeColorName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case unicode.IsSpace(r) || r == '-':
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		default:
			b.WriteRune(r)
			prevHyphen = false
		}
	}
	return b.String()
}
