package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jglueckstein/hdtools/internal/units"
)

func TestLoadMissingFileUsesKilograms(t *testing.T) {
	t.Parallel()
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DisplayUnit != units.Kilogram {
		t.Fatalf("DisplayUnit = %q", cfg.DisplayUnit)
	}
}

func TestLoadAndWriteRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Write(path, Config{DisplayUnit: units.Pound}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DisplayUnit != units.Pound {
		t.Fatalf("DisplayUnit = %q", cfg.DisplayUnit)
	}
}

func TestLoadRejectsUnknownUnit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("display_unit = \"stone\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultPathUsesXDGConfigHome(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "hdtools", "config.toml")
	if got != want {
		t.Fatalf("DefaultPath = %q, want %q", got, want)
	}
}

func TestDefaultDBPathUsesXDGDataHome(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	got, err := DefaultDBPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "hdtools", "hdtools.db")
	if got != want {
		t.Fatalf("DefaultDBPath = %q, want %q", got, want)
	}
}

func TestDefaultDBPathFallsBackToLocalShare(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	got, err := DefaultDBPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".local", "share", "hdtools", "hdtools.db")
	if got != want {
		t.Fatalf("DefaultDBPath = %q, want %q", got, want)
	}
}

func TestWriteCreatesPrivateFile(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not POSIX on Windows")
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Write(path, Default()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("config mode = %o, want 0600", perm)
	}
}

func TestEnsureDoesNotOverwrite(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Write(path, Config{DisplayUnit: units.Stone}); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(path); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DisplayUnit != units.Stone {
		t.Fatalf("Ensure overwrote unit: %q", cfg.DisplayUnit)
	}
}
