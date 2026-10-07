// Package config is the hand-editable preference file, kept out of SQLite
// so a log database can move machines without dragging display choices
// with it, and so a person can change units with a text editor.
//
// Paths follow XDG: config under $XDG_CONFIG_HOME (or ~/.config), the
// default database path advertised from here under $XDG_DATA_HOME (or
// ~/.local/share). A missing file is not an error — first run is kilograms.
// Invalid display_unit is an error so "lbs" cannot be stored as kg.
//
// [colors] is a sparse overlay of valid roles; invalid values fall back
// silently so a typo cannot keep someone out of the log. Invalid
// display_unit still fails. pdf_dir is a directory for chart PDFs:
// absolute, or a leading ~/ for the home directory. Any other relative
// value fails the export, and a non-string value does not fail Load,
// because a bad directory cannot store pounds as kilograms. This
// package does not open the database and does not write the PDF.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jglueckstein/hdtools/internal/units"
	"github.com/pelletier/go-toml/v2"
)

const (
	appName    = "hdtools"
	fileName   = "config.toml"
	dbFileName = "hdtools.db"
)

// Config is the on-disk preference file. Unknown keys are ignored so adding
// fields later does not break older files.
type Config struct {
	DisplayUnit units.Unit `toml:"display_unit"`
	// PDFDir is the chart-PDF directory. Empty means DataDir.
	// A non-string value in the file leaves this empty and sets
	// pdfDirBad so the export can fail without blocking Load.
	PDFDir string `toml:"pdf_dir,omitempty"`
	// Colors is the sparse overlay of valid [colors] roles. Omitted roles
	// use the TUI built-in default. Load fills this from the file.
	Colors map[string]string `toml:"colors,omitempty"`
	// pdfDirBad is set when pdf_dir was present and not a string.
	// It is not a file field: Write must not persist it.
	pdfDirBad bool
}

// Default is first-run preferences: kilograms, matching storage.
func Default() Config {
	return Config{DisplayUnit: units.Kilogram}
}

// ConfigDir uses os.UserConfigDir so we inherit the platform XDG/macOS/Windows
// mapping instead of hard-coding ~/.config.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config directory: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// DataDir reads XDG_DATA_HOME itself because Go has no UserDataDir helper.
// An empty variable must fall back to ~/.local/share per the spec.
func DataDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, appName), nil
}

// DefaultPath is the file Ensure will create on first run.
func DefaultPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", fmt.Errorf("default config path: %w", err)
	}
	return filepath.Join(dir, fileName), nil
}

// DefaultDBPath lives beside ConfigDir's sibling data dir so -db is optional
// for the common local-file case.
func DefaultDBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", fmt.Errorf("default database path: %w", err)
	}
	return filepath.Join(dir, dbFileName), nil
}

// ResolvePDFDir chooses the directory for a chart PDF when -o is empty.
// An absolute pdf_dir is used as written. A leading ~/ expands to the
// home directory and nothing else is expanded, so a stored path does
// not follow the directory the process was started from. Any other
// relative value is an error. An empty pdf_dir is DataDir. A result
// that is not absolute fails here, without rewriting DataDir, so a
// relative base cannot create the chart under the working directory.
func ResolvePDFDir(cfg Config) (string, error) {
	dir, err := chosenPDFDir(cfg)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("pdf directory %q is not absolute", dir)
	}
	return dir, nil
}

func chosenPDFDir(cfg Config) (string, error) {
	if cfg.pdfDirBad {
		return "", fmt.Errorf("pdf_dir must be a string")
	}
	dir := strings.TrimSpace(cfg.PDFDir)
	if dir == "" {
		return DataDir()
	}
	if filepath.IsAbs(dir) {
		return dir, nil
	}
	if rest, ok := strings.CutPrefix(dir, "~/"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory: %w", err)
		}
		return filepath.Join(home, rest), nil
	}
	return "", fmt.Errorf("pdf_dir %q must be absolute or start with ~/", dir)
}

// Load reads path. A missing file returns Default. An invalid display_unit
// fails so we never treat "lbs" as kilograms. pdf_dir is decoded as a
// raw value so a number does not fail the whole file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var raw struct {
		DisplayUnit string `toml:"display_unit"`
		Colors      any    `toml:"colors"`
		PDFDir      any    `toml:"pdf_dir"`
	}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg := Default()
	if raw.DisplayUnit != "" {
		u, err := units.Parse(raw.DisplayUnit)
		if err != nil {
			return Config{}, fmt.Errorf("config %s: %w", path, err)
		}
		cfg.DisplayUnit = u
	}
	cfg.Colors = parseColorOverlay(raw.Colors)
	switch v := raw.PDFDir.(type) {
	case nil:
	case string:
		cfg.PDFDir = strings.TrimSpace(v)
	default:
		cfg.pdfDirBad = true
	}
	return cfg, nil
}

// Write creates parent directories because first-run ~/.config/hdtools may
// not exist yet. Overwriting is intentional: this is the save path for
// future in-app preference edits. The file is 0600 so display preferences
// are not world-readable.
func Write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	body, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("chmod config %s: %w", path, err)
	}
	return nil
}

// Ensure is first-run only: a user who set display_unit = "lb" must not
// have that file replaced with kilograms on the next launch.
func Ensure(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat config %s: %w", path, err)
	}
	return Write(path, Default())
}
