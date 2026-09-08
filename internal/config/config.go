// Package config is the hand-editable preference file, kept out of SQLite
// so a log database can move machines without dragging display choices
// with it, and so a person can change units with a text editor.
//
// Paths follow XDG: config under $XDG_CONFIG_HOME (or ~/.config), the
// default database path advertised from here under $XDG_DATA_HOME (or
// ~/.local/share). A missing file is not an error — first run is kilograms.
// Invalid display_unit is an error so "lbs" cannot be stored as kg.
//
// Color schemes and NO_COLOR are specified in idea.md but not parsed here
// yet. This package does not open the database.
package config

import (
	"fmt"
	"os"
	"path/filepath"

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
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// DefaultDBPath lives beside ConfigDir's sibling data dir so -db is optional
// for the common local-file case.
func DefaultDBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbFileName), nil
}

// Load reads path. A missing file returns Default. An invalid display_unit
// fails so we never treat "lbs" as kilograms.
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
	return cfg, nil
}

// Write creates parent directories because first-run ~/.config/hdtools may
// not exist yet. Overwriting is intentional: this is the save path for
// future in-app preference edits.
func Write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	body, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
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
