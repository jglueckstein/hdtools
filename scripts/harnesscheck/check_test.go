package main

// The fixtures are the cases the pull-request scripts must catch,
// and the cases they must leave alone. They are not a scan of the
// real tree. That scan is the scripts themselves.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckNaming(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "internal/tui/monthsave.go", "package tui\n")
	writeFile(t, root, "internal/tui/month_save.go", "package tui\n")
	writeFile(t, root, "internal/tui/tui_test.go", "package tui_test\n")
	writeFile(t, root, "internal/bad/bad.go", "package Bad\n")

	got, err := checkNaming(root)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(got, "\n")
	for _, want := range []string{
		"month_save.go: file name contains '_'",
		"package Bad",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "monthsave.go") || strings.Contains(text, "tui_test") {
		t.Fatalf("flagged an allowed name:\n%s", text)
	}
}

func TestCheckLayout(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "cmd/hdtools/main.go", "package main\n")
	writeFile(t, root, "internal/tui/app_test.go", "package tui\n")
	writeFile(t, root, "scripts/harnesscheck/main.go", "package main\n")
	writeFile(t, root, "other.go", "package other\n")

	got, err := checkLayout(root)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(got, "\n")
	if !strings.Contains(text, "other.go: production file") {
		t.Fatalf("missing other.go finding:\n%s", text)
	}
	if strings.Contains(text, "main.go") || strings.Contains(text, "app_test.go") || strings.Contains(text, "harnesscheck") {
		t.Fatalf("flagged an allowed path:\n%s", text)
	}
}

func TestCheckLayoutRequiresMain(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	got, err := checkLayout(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.Contains(got[0], "cmd/hdtools/main.go: missing") {
		t.Fatalf("got %v", got)
	}
}

func TestCheckErrors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "internal/config/config.go", "package config\nfunc f() error {\n\terr := g()\n\treturn err\n}\nfunc g() error { return nil }\n")
	writeFile(t, root, "internal/config/ok.go", "package config\nimport \"fmt\"\nfunc h() error {\n\terr := g()\n\treturn fmt.Errorf(\"h: %w\", err)\n}\n")
	writeFile(t, root, "cmd/hdtools/main_test.go", "package main\nfunc t() error {\n\terr := g()\n\treturn err\n}\n")

	got, err := checkErrors(root)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(got, "\n")
	if !strings.Contains(text, "config.go") || !strings.Contains(text, "returned err is not wrapped") {
		t.Fatalf("missing bare return:\n%s", text)
	}
	if strings.Contains(text, "ok.go") || strings.Contains(text, "main_test.go") {
		t.Fatalf("flagged a wrapped return or a test:\n%s", text)
	}
}

func TestCheckDocs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "internal/units/units.go", "package units\nfunc Exported() {}\n// Kept documents the value.\nconst Kept = 1\n")
	got, err := checkDocs(root)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(got, "\n")
	if !strings.Contains(text, "Exported") {
		t.Fatalf("missing Exported:\n%s", text)
	}
	if strings.Contains(text, "Kept") {
		t.Fatalf("flagged a documented const:\n%s", text)
	}
}

func TestCheckPreamble(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "internal/tui/bare.go", "package tui\nfunc f() {}\n")
	writeFile(t, root, "internal/tui/after.go", "package tui\n\n// The sheet save lives here and does not draw.\n\nimport \"fmt\"\nfunc f() { fmt.Println() }\n")
	writeFile(t, root, "internal/tui/before.go", "// Wiring only.\npackage tui\n")
	got, err := checkPreamble(root)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(got, "\n")
	if !strings.Contains(text, "bare.go: no opening comment") {
		t.Fatalf("missing bare.go:\n%s", text)
	}
	if strings.Contains(text, "after.go") || strings.Contains(text, "before.go") {
		t.Fatalf("flagged a file that opens with a comment:\n%s", text)
	}
}
