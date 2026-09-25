package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func writeTOML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingColorsUsesDefaults(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, "display_unit = \"kg\"\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Colors) != 0 {
		t.Fatalf("Colors = %#v, want empty overlay", cfg.Colors)
	}
}

func TestLoadEmptyColorsTable(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, "display_unit = \"kg\"\n\n[colors]\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Colors) != 0 {
		t.Fatalf("Colors = %#v, want empty overlay", cfg.Colors)
	}
}

func TestLoadColorNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, role, want string
	}{
		{`weight = "green"`, "weight", "green"},
		{`weight = "Blue"`, "weight", "blue"},
		{`trend = "BRIGHT-RED"`, "trend", "bright-red"},
		{`muted = "bright black"`, "muted", "bright-black"},
		{`muted = "gray"`, "muted", "bright-black"},
		{`muted = "grey"`, "muted", "bright-black"},
		{`weight = "4"`, "weight", "blue"},
		{`weight = 4`, "weight", "blue"},
		{`title = 6`, "title", "cyan"},
		{`weight = "#00ff00"`, "weight", "#00ff00"},
		{`weight = "#0f0"`, "weight", "#00ff00"},
		{`trend = "#0D47A1"`, "trend", "#0d47a1"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			path := writeTOML(t, "display_unit = \"kg\"\n\n[colors]\n"+tc.in+"\n")
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Colors[tc.role] != tc.want {
				t.Fatalf("Colors[%q] = %q, want %q (map %#v)", tc.role, cfg.Colors[tc.role], tc.want, cfg.Colors)
			}
		})
	}
}

func TestLoadInvalidColorFallsBack(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, `display_unit = "kg"

[colors]
weight = "chartreuse"
trend = "yellow"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Colors["weight"]; ok {
		t.Fatalf("invalid weight kept: %#v", cfg.Colors)
	}
	if cfg.Colors["trend"] != "yellow" {
		t.Fatalf("trend = %q, want yellow", cfg.Colors["trend"])
	}
}

func TestInvalidDeltaColorDropped(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, `display_unit = "kg"

[colors]
delta-pos = "chartreuse"
trend = "yellow"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Colors["delta-pos"]; ok {
		t.Fatalf("invalid delta-pos kept: %#v", cfg.Colors)
	}
	if cfg.Colors["trend"] != "yellow" {
		t.Fatalf("trend = %q, want yellow", cfg.Colors["trend"])
	}
}

func TestLoadMalformedHexAndSequencesFallBack(t *testing.T) {
	t.Parallel()
	for _, in := range []string{
		`weight = "ff0000"`,
		`weight = "#00ff00ff"`,
		`weight = "\x1b[34m"`,
		`weight = ""`,
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			path := writeTOML(t, "display_unit = \"kg\"\n\n[colors]\n"+in+"\ntrend = \"yellow\"\n")
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := cfg.Colors["weight"]; ok {
				t.Fatalf("invalid %s kept: %#v", in, cfg.Colors)
			}
			if cfg.Colors["trend"] != "yellow" {
				t.Fatalf("trend = %q, want yellow", cfg.Colors["trend"])
			}
		})
	}
}

func TestLoadNonStringRoleFallsBack(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, `display_unit = "kg"

[colors]
weight = true
trend = "yellow"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Colors["weight"]; ok {
		t.Fatalf("bool weight kept: %#v", cfg.Colors)
	}
	if cfg.Colors["trend"] != "yellow" {
		t.Fatalf("trend = %q, want yellow", cfg.Colors["trend"])
	}
}

func TestLoadColorsNotATableUsesDefaults(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, "display_unit = \"kg\"\ncolors = \"red\"\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Colors) != 0 {
		t.Fatalf("Colors = %#v, want empty overlay", cfg.Colors)
	}
}

func TestLoadUnknownColorKeyIgnored(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, `display_unit = "kg"

[colors]
accent = "cyan"
title = "magenta"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Colors["accent"]; ok {
		t.Fatalf("accent kept: %#v", cfg.Colors)
	}
	if cfg.Colors["title"] != "magenta" {
		t.Fatalf("title = %q, want magenta", cfg.Colors["title"])
	}
}

func TestLoadRejectsUnknownUnitWithColors(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, "display_unit = \"stone\"\n\n[colors]\nweight = \"green\"\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestEnsureDoesNotWriteColorsTable(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Ensure(path); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "[colors]") {
		t.Fatalf("first-run file has [colors]: %s", body)
	}
}

func TestWriteRoundTripCustomColors(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := Config{
		DisplayUnit: units.Kilogram,
		Colors:      map[string]string{"weight": "green", "trend": "#ff0000"},
	}
	if err := Write(path, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Colors["weight"] != "green" || got.Colors["trend"] != "#ff0000" {
		t.Fatalf("Colors = %#v", got.Colors)
	}
}

func TestWriteRoundTripIsSparse(t *testing.T) {
	t.Parallel()
	path := writeTOML(t, `display_unit = "kg"

[colors]
weight = "green"
trend = "chartreuse"
accent = "cyan"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Write(path, cfg); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "weight") || !strings.Contains(s, "green") {
		t.Fatalf("missing weight=green: %s", s)
	}
	if strings.Contains(s, "trend") || strings.Contains(s, "chartreuse") {
		t.Fatalf("invalid trend rewritten: %s", s)
	}
	if strings.Contains(s, "accent") {
		t.Fatalf("unknown accent rewritten: %s", s)
	}
	if strings.Contains(s, "title") {
		t.Fatalf("omitted title materialized: %s", s)
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
