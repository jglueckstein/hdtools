// Command hdtools is the Hacker's Diet TUI.
//
// main only resolves the SQLite path, opens the store, and runs Bubble Tea.
// Screen behaviour lives in internal/tui so the binary stays a wiring layer.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "hdtools: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("hdtools", flag.ContinueOnError)
	dbFlag := fs.String("db", "", "SQLite database path (default $HDTOOLS_DB or ~/.hdtools/hdtools.db)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	path, err := resolveDBPath(*dbFlag)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	store, err := dailylog.Open(path)
	if err != nil {
		return err
	}

	p := tea.NewProgram(tui.New(store, path))
	_, runErr := p.Run()
	closeErr := store.Close()
	if runErr != nil {
		return fmt.Errorf("run tui: %w", runErr)
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func resolveDBPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	if env := os.Getenv("HDTOOLS_DB"); env != "" {
		return env, nil
	}
	path, err := tui.DefaultDBPath()
	if err != nil {
		return "", fmt.Errorf("resolve database path: %w", err)
	}
	return path, nil
}
