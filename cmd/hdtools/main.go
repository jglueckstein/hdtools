// Command hdtools is the process that turns flags and XDG paths into a
// running TUI, or into one PDF when -chart-pdf or -log-pdf is set.
//
// Flag, env, and directory creation live here because they are process
// concerns. The TUI must not import os.Args or open SQLite itself: tests
// inject a store, and a later remote database should reuse the same Model.
// This file does not parse keystrokes, SQL, or paint a page. It only
// chooses the path, loads the series, and calls the writer.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/chartpdf"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/logpdf"
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
	logPDF := fs.String("log-pdf", "", "write a filled monthly log sheet for YYYY-MM and exit")
	outFlag := fs.String("o", "", "output file for -chart-pdf and -log-pdf (default YYYY-MM-chart.pdf or YYYY-MM-log.pdf in the data directory)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	// Both flags are rejected before a month check so a bad month cannot
	// hide the clash, and before config or the database so a rejected
	// export leaves neither behind.
	if *logPDF != "" && *chartPDF != "" {
		return fmt.Errorf("log-pdf and chart-pdf cannot both be set")
	}

	if *chartPDF != "" {
		if _, _, err := parseMonth(*chartPDF); err != nil {
			return fmt.Errorf("chart-pdf: %w", err)
		}
	}
	if *logPDF != "" {
		if _, _, err := parseMonth(*logPDF); err != nil {
			return fmt.Errorf("log-pdf: %w", err)
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

	if *logPDF != "" {
		err := exportLogPDF(store, cfg, *logPDF, *outFlag)
		closeErr := store.Close()
		if err != nil {
			return fmt.Errorf("log pdf: %w", err)
		}
		if closeErr != nil {
			return fmt.Errorf("close store: %w", closeErr)
		}
		return nil
	}

	if *chartPDF != "" {
		err := exportChartPDF(store, cfg, *chartPDF, *outFlag)
		closeErr := store.Close()
		if err != nil {
			return fmt.Errorf("chart pdf: %w", err)
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

// exportLogPDF loads the whole series. ApplyTrend refuses a leading
// weightless row, and the TUI treats that as a load error. This export
// does not: order is checked first, then trend starts at the first
// weigh-in. An empty log and a month with no weigh-in still print.
func exportLogPDF(store *dailylog.Store, cfg config.Config, monthStr, out string) error {
	year, month, err := parseMonth(monthStr)
	if err != nil {
		return fmt.Errorf("log-pdf: %w", err)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		return fmt.Errorf("log-pdf: load logs: %w", err)
	}
	if err := rejectUnsorted(logs); err != nil {
		return fmt.Errorf("log-pdf: %w", err)
	}
	logs, err = trendFromFirstWeighIn(logs)
	if err != nil {
		return fmt.Errorf("log-pdf: apply trend: %w", err)
	}
	if out == "" {
		dir, err := config.ResolvePDFDir(cfg)
		if err != nil {
			return fmt.Errorf("log-pdf: %w", err)
		}
		out = filepath.Join(dir, fmt.Sprintf("%04d-%02d-log.pdf", year, month))
	}
	if err := logpdf.Write(out, logpdf.Options{
		Year:  year,
		Month: month,
		Logs:  logs,
		Unit:  cfg.DisplayUnit,
		Today: time.Now(),
	}); err != nil {
		return fmt.Errorf("log-pdf: %w", err)
	}
	fmt.Println(out)
	return nil
}

// rejectUnsorted fails a duplicate day and a series that goes backwards,
// including when no row has a weigh-in. ApplyTrend is skipped in that
// case, so it would not see the duplicate.
func rejectUnsorted(logs []dailylog.DailyLog) error {
	for i := 1; i < len(logs); i++ {
		if logs[i].Day.After(logs[i-1].Day) {
			continue
		}
		if logs[i].Day.Equal(logs[i-1].Day) {
			return dailylog.ErrDuplicateDay
		}
		return dailylog.ErrUnsortedDays
	}
	return nil
}

// trendFromFirstWeighIn leaves leading weightless rows untrended.
// ErrNoTrendStart is not an export failure: the page still prints.
func trendFromFirstWeighIn(logs []dailylog.DailyLog) ([]dailylog.DailyLog, error) {
	start := -1
	for i := range logs {
		if logs[i].Weight != nil {
			start = i
			break
		}
	}
	if start < 0 {
		return logs, nil
	}
	trended, err := dailylog.ApplyTrend(logs[start:], nil)
	if err != nil {
		if errors.Is(err, dailylog.ErrNoTrendStart) {
			return logs, nil
		}
		return nil, fmt.Errorf("trend suffix: %w", err)
	}
	if start == 0 {
		return trended, nil
	}
	out := make([]dailylog.DailyLog, 0, len(logs))
	out = append(out, logs[:start]...)
	out = append(out, trended...)
	return out, nil
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
		dir, err := config.ResolvePDFDir(cfg)
		if err != nil {
			return fmt.Errorf("chart-pdf: %w", err)
		}
		out = filepath.Join(dir, fmt.Sprintf("%04d-%02d-chart.pdf", year, month))
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
	fmt.Println(out)
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
