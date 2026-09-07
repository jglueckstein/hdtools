package config

import (
	"os"
	"path/filepath"
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
