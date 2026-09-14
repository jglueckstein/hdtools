// Command hdtools is the process that turns flags and XDG paths into a
// running TUI.
//
// Flag, env, and directory creation live here because they are process
// concerns. The TUI must not import os.Args or open SQLite itself: tests
// inject a store, and a later remote database should reuse the same Model.
// This file does not parse keystrokes or SQL.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/chartpdf"
	"github.com/jglueckstein/hdtools/internal/config"
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
	dbFlag := fs.String("db", "", "SQLite database path (default $HDTOOLS_DB or $XDG_DATA_HOME/hdtools/hdtools.db)")
	configFlag := fs.String("config", "", "config file path (default $HDTOOLS_CONFIG or $XDG_CONFIG_HOME/hdtools/config.toml)")
	chartPDF := fs.String("chart-pdf", "", "write a monthly chart PDF for YYYY-MM and exit")
	outFlag := fs.String("o", "", "output path for -chart-pdf (default YYYY-MM-chart.pdf)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	if *chartPDF != "" {
		if _, _, err := parseMonth(*chartPDF); err != nil {
			return fmt.Errorf("chart-pdf: %w", err)
		}
	}

	cfgPath, err := resolveConfigPath(*configFlag)
	if err != nil {
		return fmt.Errorf("config path: %w", err)
	}
	if err := config.Ensure(cfgPath); err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	path, err := resolveDBPath(*dbFlag)
	if err != nil {
		return fmt.Errorf("database path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	store, err := dailylog.Open(path)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	if *chartPDF != "" {
		err := exportChartPDF(store, cfg, *chartPDF, *outFlag)
		closeErr := store.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return fmt.Errorf("close store: %w", closeErr)
		}
		return nil
	}

	p := tea.NewProgram(tui.New(store, path, cfg))
	_, runErr := p.Run()
	closeErr := store.Close()
	if runErr != nil {
		return fmt.Errorf("run tui: %w", runErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close store: %w", closeErr)
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
	path, err := config.DefaultDBPath()
	if err != nil {
		return "", fmt.Errorf("resolve database path: %w", err)
	}
	return path, nil
}

func exportChartPDF(store *dailylog.Store, cfg config.Config, monthStr, out string) error {
	year, month, err := parseMonth(monthStr)
	if err != nil {
		return fmt.Errorf("chart-pdf: %w", err)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		return fmt.Errorf("chart-pdf: load logs: %w", err)
	}
	if len(logs) > 0 {
		logs, err = dailylog.ApplyTrend(logs, nil)
		if err != nil {
			return fmt.Errorf("chart-pdf: apply trend: %w", err)
		}
	}
	if out == "" {
		out = fmt.Sprintf("%04d-%02d-chart.pdf", year, month)
	}
	now := time.Now()
	if err := chartpdf.Write(out, chartpdf.Options{
		Year:   year,
		Month:  month,
		Logs:   logs,
		Unit:   cfg.DisplayUnit,
		Colors: cfg.Colors,
		Today:  now,
	}); err != nil {
		return fmt.Errorf("chart-pdf: %w", err)
	}
	return nil
}

// parseMonth turns -chart-pdf YYYY-MM into a calendar month. Invalid
// values must fail closed so a typo cannot start the TUI or write a file.
func parseMonth(s string) (int, time.Month, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil || t.Format("2006-01") != s {
		return 0, 0, fmt.Errorf("month %q: want YYYY-MM", s)
	}
	y, m, _ := t.Date()
	return y, m, nil
}

func resolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	if env := os.Getenv("HDTOOLS_CONFIG"); env != "" {
		return env, nil
	}
	path, err := config.DefaultPath()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}
	return path, nil
}
