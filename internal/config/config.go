// Package config loads display preferences from a TOML file under the XDG
// config directory, not from SQLite. That keeps the log portable and the
// file hand-editable.
//
// Defaults follow the XDG Base Directory Specification: config lives under
// $XDG_CONFIG_HOME (or ~/.config), and the SQLite file lives under
// $XDG_DATA_HOME (or ~/.local/share). A missing config file is not an error:
// first run uses kilograms until the user writes display_unit.
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

// ConfigDir is $XDG_CONFIG_HOME/hdtools, or ~/.config/hdtools.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config directory: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// DataDir is $XDG_DATA_HOME/hdtools, or ~/.local/share/hdtools.
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

// DefaultPath is $XDG_CONFIG_HOME/hdtools/config.toml.
func DefaultPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// DefaultDBPath is $XDG_DATA_HOME/hdtools/hdtools.db.
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

// Write creates parent directories and saves cfg so the user can edit it.
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

// Ensure writes Default to path only when the file does not exist.
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
