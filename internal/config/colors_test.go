package config

import (
	"os"
	"strings"
	"testing"
)

func TestColorRoleNamesHaveDocComment(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("colors.go")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(body), "\n")
	constIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "const (" {
			continue
		}
		if i+1 < len(lines) && strings.Contains(lines[i+1], "RoleTitle") {
			constIdx = i
			break
		}
	}
	if constIdx < 1 {
		t.Fatal("color role const block not found")
	}
	doc := strings.TrimSpace(lines[constIdx-1])
	if !strings.HasPrefix(doc, "//") || !strings.Contains(strings.ToLower(doc), "role") {
		t.Fatalf("color role const block doc = %q, want a comment that says what the role names are", doc)
	}
}
