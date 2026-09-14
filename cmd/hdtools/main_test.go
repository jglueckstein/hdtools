package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
)

func TestResolveDBPathPrefersFlagThenEnv(t *testing.T) {
	t.Setenv("HDTOOLS_DB", "/from/env.db")
	got, err := resolveDBPath("/from/flag.db")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/from/flag.db" {
		t.Fatalf("got %q, want flag path", got)
	}

	got, err = resolveDBPath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/from/env.db" {
		t.Fatalf("got %q, want env path", got)
	}
}

func TestResolveDBPathDefault(t *testing.T) {
	t.Setenv("HDTOOLS_DB", "")
	got, err := resolveDBPath("")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "hdtools.db" {
		t.Fatalf("got %q", got)
	}
}

func TestParseMonthRejects(t *testing.T) {
	for _, in := range []string{"1990-13", "banana"} {
		t.Run(in, func(t *testing.T) {
			if _, _, err := parseMonth(in); err == nil {
				t.Fatalf("parseMonth(%q) succeeded, want error", in)
			}
			dir := t.TempDir()
			t.Chdir(dir)
			dbPath := filepath.Join(dir, "t.db")
			cfgPath := filepath.Join(dir, "config.toml")
			out := filepath.Join(dir, "out.pdf")
			err := runCLI(t, "-chart-pdf", in, "-o", out, "-db", dbPath, "-config", cfgPath)
			if err == nil {
				t.Fatal("want non-zero exit")
			}
			if strings.Contains(err.Error(), "not implemented") {
				t.Fatalf("want rejection of %q, got stub: %v", in, err)
			}
			if _, statErr := os.Stat(out); statErr == nil {
				t.Fatal("wrote a PDF for an invalid month")
			}
			if _, statErr := os.Stat(cfgPath); statErr == nil {
				t.Fatal("created config for an invalid month")
			}
			if _, statErr := os.Stat(dbPath); statErr == nil {
				t.Fatal("created database for an invalid month")
			}
		})
	}
}

func TestChartPDFFlagWriteFailure(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	dbPath, cfgPath := cliPaths(t, dir)
	seedNovember(t, dbPath)
	out := filepath.Join(dir, "out.pdf")
	if err := os.Mkdir(out, 0o700); err != nil {
		t.Fatal(err)
	}
	err := runCLI(t, "-chart-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatal("want non-zero exit")
	}
	info, statErr := os.Stat(out)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !info.IsDir() {
		t.Fatal("replaced the unwritable path with a PDF")
	}
}

func TestChartPDFFlagSkipsTUI(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	dbPath, cfgPath := cliPaths(t, dir)
	seedNovember(t, dbPath)
	out := filepath.Join(dir, "out.pdf")
	if err := runCLI(t, "-chart-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 4 || string(raw[:4]) != "%PDF" {
		t.Fatalf("header = %q, want %%PDF", raw[:min(8, len(raw))])
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
}

func cliPaths(t *testing.T, dir string) (dbPath, cfgPath string) {
	t.Helper()
	dbPath = filepath.Join(dir, "t.db")
	cfgPath = filepath.Join(dir, "config.toml")
	if err := config.Write(cfgPath, config.Default()); err != nil {
		t.Fatal(err)
	}
	return dbPath, cfgPath
}

func seedNovember(t *testing.T, dbPath string) {
	t.Helper()
	store, err := dailylog.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for i, w := range []float64{80, 79} {
		ww := w
		log, err := dailylog.New(time.Date(1990, 11, 1+i, 0, 0, 0, 0, time.UTC), &ww, 8, 0, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Upsert(context.Background(), log); err != nil {
			t.Fatal(err)
		}
	}
}

func runCLI(t *testing.T, args ...string) error {
	t.Helper()
	errc := make(chan error, 1)
	go func() { errc <- run(args) }()
	select {
	case err := <-errc:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("CLI did not return; TUI probably started")
		return nil
	}
}
