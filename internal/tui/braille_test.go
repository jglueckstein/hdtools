package tui

// These helpers read styled runes back out of a chart view. The
// Online pictures use them. The old eight-row bin oracle left with
// the Braille-cell grid. This file does not paint.

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type cellStyle struct {
	r    rune
	bold bool
	seq  string
}

func isBraille(r rune) bool {
	return r >= 0x2800 && r <= 0x28FF
}

func styledRunes(s string) []cellStyle {
	bold := false
	var fg []int
	var out []cellStyle
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j >= len(s) {
				break
			}
			bold, fg = applySGR(bold, fg, s[i+2:j])
			i = j + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size <= 0 {
			break
		}
		out = append(out, cellStyle{r: r, bold: bold, seq: styleSeq(bold, fg)})
		i += size
	}
	return out
}

func applySGR(bold bool, fg []int, body string) (bool, []int) {
	if body == "" {
		return false, nil
	}
	parts := strings.Split(body, ";")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			nums = append(nums, 0)
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			bold = false
			fg = nil
		case n == 1:
			bold = true
		case n == 22:
			bold = false
		case n == 39:
			fg = nil
		case (n >= 30 && n <= 37) || (n >= 90 && n <= 97):
			fg = []int{n}
		case n == 38 && i+2 < len(nums) && nums[i+1] == 5:
			fg = []int{38, 5, nums[i+2]}
			i += 2
		case n == 38 && i+4 < len(nums) && nums[i+1] == 2:
			fg = []int{38, 2, nums[i+2], nums[i+3], nums[i+4]}
			i += 4
		}
	}
	return bold, fg
}

func styleSeq(bold bool, fg []int) string {
	var parts []string
	if bold {
		parts = append(parts, "1")
	}
	for _, n := range fg {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ";")
}

func cellHasFG(c cellStyle, n int) bool {
	if c.seq == "" {
		return false
	}
	return hasIndexedForeground("\x1b["+c.seq+"m", n)
}
