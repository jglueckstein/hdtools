package config

// [colors] is a sparse overlay: only present, valid roles are kept so
// Write does not freeze today's defaults into the file. Invalid values
// are dropped, not rewritten as the built-in name. This file does not
// paint the TUI; it only canonicalizes what the file said.

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	RoleTitle     = "title"
	RoleMuted     = "muted"
	RoleHeader    = "header"
	RoleHelp      = "help"
	RoleWeight    = "weight"
	RoleTrend     = "trend"
	RoleDeltaPos  = "delta-pos"
	RoleDeltaNeg  = "delta-neg"
	RoleDeltaZero = "delta-zero"
	RoleSelection = "selection"
	RoleError     = "error"
	RoleStatus    = "status"
)

var colorRoles = map[string]struct{}{
	RoleTitle:     {},
	RoleMuted:     {},
	RoleHeader:    {},
	RoleHelp:      {},
	RoleWeight:    {},
	RoleTrend:     {},
	RoleDeltaPos:  {},
	RoleDeltaNeg:  {},
	RoleDeltaZero: {},
	RoleSelection: {},
	RoleError:     {},
	RoleStatus:    {},
}

var ansiNames = [...]string{
	"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"bright-black", "bright-red", "bright-green", "bright-yellow",
	"bright-blue", "bright-magenta", "bright-cyan", "bright-white",
}

var namedANSI = map[string]string{
	"black": "black", "red": "red", "green": "green", "yellow": "yellow",
	"blue": "blue", "magenta": "magenta", "cyan": "cyan", "white": "white",
	"bright-black": "bright-black", "gray": "bright-black", "grey": "bright-black",
	"bright-red": "bright-red", "bright-green": "bright-green",
	"bright-yellow": "bright-yellow", "bright-blue": "bright-blue",
	"bright-magenta": "bright-magenta", "bright-cyan": "bright-cyan",
	"bright-white": "bright-white",
}

// parseColorOverlay keeps valid role→color entries and drops everything
// else. A non-table colors value is treated as a missing table so a typo
// cannot fail the load.
func parseColorOverlay(v any) map[string]string {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string)
	for key, raw := range m {
		if _, known := colorRoles[key]; !known {
			continue
		}
		canon, ok := canonicalColor(raw)
		if !ok {
			continue
		}
		out[key] = canon
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func canonicalColor(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return parseColorString(t)
	case int64:
		return ansiIndex(t)
	case uint64:
		if t > 15 {
			return "", false
		}
		return ansiIndex(int64(t))
	case int:
		return ansiIndex(int64(t))
	default:
		return "", false
	}
}

func ansiIndex(n int64) (string, bool) {
	if n < 0 || n > 15 {
		return "", false
	}
	return ansiNames[n], true
}

func parseColorString(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "#") {
		return parseHex(s)
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ansiIndex(n)
	}
	name := normalizeColorName(s)
	canon, ok := namedANSI[name]
	return canon, ok
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

func parseHex(s string) (string, bool) {
	h := strings.ToLower(s[1:])
	switch len(h) {
	case 3:
		if !isHex(h) {
			return "", false
		}
		return fmt.Sprintf("#%c%c%c%c%c%c", h[0], h[0], h[1], h[1], h[2], h[2]), true
	case 6:
		if !isHex(h) {
			return "", false
		}
		return "#" + h, true
	default:
		return "", false
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
