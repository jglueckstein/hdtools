package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"database/sql"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
	if !strings.Contains(err.Error(), "chart pdf:") {
		t.Fatalf("error %q does not name the chart-pdf step in run", err)
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
	_, err := runCLIOutput(t, args...)
	return err
}

// runCLIOutput runs the CLI and returns standard output. Success and
// failure both have to be observable: a chart export prints the path
// it wrote, and a failed export prints nothing.
func runCLIOutput(t *testing.T, args ...string) (string, error) {
	t.Helper()
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatal(pipeErr)
	}
	orig := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	errc := make(chan error, 1)
	go func() { errc <- run(args) }()

	var runErr error
	select {
	case runErr = <-errc:
	case <-time.After(5 * time.Second):
		os.Stdout = orig
		_ = w.Close()
		_ = r.Close()
		t.Fatal("CLI did not return; TUI probably started")
	}
	os.Stdout = orig
	if err := w.Close(); err != nil {
		_ = r.Close()
		t.Fatal(err)
	}
	out, readErr := io.ReadAll(r)
	_ = r.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(out), runErr
}

func isolateCLI(t *testing.T) string {
	t.Helper()
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("HDTOOLS_DB", "")
	t.Setenv("HDTOOLS_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return cwd
}

func cliDataHome(t *testing.T) (dir, pdf string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	dir = filepath.Join(root, "hdtools")
	pdf = filepath.Join(dir, chartPDFName)
	return dir, pdf
}

const chartPDFName = "1990-11-chart.pdf"
const logPDFName = "1990-11-log.pdf"

func writeConfigFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("wrote %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func assertChartFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory", path)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 4 || string(raw[:4]) != "%PDF" {
		t.Fatalf("header = %q, want %%PDF", raw[:min(8, len(raw))])
	}
}

func assertPDFNotCwd(t *testing.T, cwd, path string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(cwd, chartPDFName)); err == nil {
		t.Fatalf("wrote %s in the working directory, want %s", chartPDFName, path)
	}
	assertChartFile(t, path)
}

func wantStdoutLine(t *testing.T, stdout, path string) {
	t.Helper()
	if stdout != path+"\n" {
		t.Fatalf("stdout = %q, want one line %q", stdout, path)
	}
}

func wantNoStdoutPath(t *testing.T, stdout string) {
	t.Helper()
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("stdout = %q, want no path", stdout)
	}
}

func TestChartPDFDefaultWritesDataDir(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliDataHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertPDFNotCwd(t, cwd, dataPDF)
	wantStdoutLine(t, stdout, dataPDF)
}

func TestChartPDFDefaultUsesStandInHome(t *testing.T) {
	cwd := isolateCLI(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(home, ".local", "share", "hdtools", chartPDFName)
	assertPDFNotCwd(t, cwd, pdf)
	wantStdoutLine(t, stdout, pdf)
}

func TestChartPDFDirAbsolute(t *testing.T) {
	cases := []struct {
		name  string
		value func(dir string) string
	}{
		{name: "absolute", value: func(dir string) string { return dir }},
		{name: "trailing-slash", value: func(dir string) string { return dir + "/" }},
		{name: "padded", value: func(dir string) string { return "  " + dir + "  " }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cwd := isolateCLI(t)
			_, dataPDF := cliDataHome(t)
			pdfDir := filepath.Join(t.TempDir(), "charts")
			dbPath, cfgPath := cliPaths(t, cwd)
			body := "display_unit = \"kg\"\npdf_dir = " + strconv.Quote(tc.value(pdfDir)) + "\n"
			writeConfigFile(t, cfgPath, body)
			seedNovember(t, dbPath)
			stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
			if err != nil {
				t.Fatal(err)
			}
			pdf := filepath.Join(pdfDir, chartPDFName)
			assertPDFNotCwd(t, cwd, pdf)
			assertNoFile(t, dataPDF)
			wantStdoutLine(t, stdout, pdf)
		})
	}
}

func TestChartPDFDirWithoutHome(t *testing.T) {
	cwd := isolateCLI(t)
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	pdfDir := filepath.Join(t.TempDir(), "charts")
	dbPath := filepath.Join(cwd, "t.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(pdfDir, chartPDFName)
	assertPDFNotCwd(t, cwd, pdf)
	wantStdoutLine(t, stdout, pdf)
}

func TestChartPDFBlankPDFDirWritesDataDir(t *testing.T) {
	colorDir := filepath.Join(t.TempDir(), "from-colors")
	cases := []struct {
		name string
		body string
		skip string
	}{
		{name: "omitted", body: "display_unit = \"kg\"\n"},
		{name: "empty", body: "display_unit = \"kg\"\npdf_dir = \"\"\n"},
		{name: "whitespace", body: "display_unit = \"kg\"\npdf_dir = \"   \"\n"},
		{name: "unknown-key", body: "display_unit = \"kg\"\nnot_a_dir = \"x\"\n"},
		{
			name: "colors",
			body: "display_unit = \"kg\"\n\n[colors]\npdf_dir = " + strconv.Quote(colorDir) + "\nweight = \"green\"\n",
			skip: colorDir,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cwd := isolateCLI(t)
			_, dataPDF := cliDataHome(t)
			dbPath := filepath.Join(cwd, "t.db")
			cfgPath := filepath.Join(cwd, "config.toml")
			writeConfigFile(t, cfgPath, tc.body)
			seedNovember(t, dbPath)
			stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
			if err != nil {
				t.Fatal(err)
			}
			assertPDFNotCwd(t, cwd, dataPDF)
			wantStdoutLine(t, stdout, dataPDF)
			if tc.skip != "" {
				assertNoFile(t, filepath.Join(tc.skip, chartPDFName))
			}
		})
	}
}

func TestChartPDFOAbsoluteWins(t *testing.T) {
	t.Run("absolute", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		pdfDir := filepath.Join(t.TempDir(), "charts")
		out := filepath.Join(t.TempDir(), "out.pdf")
		dbPath, cfgPath := cliPaths(t, cwd)
		writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertChartFile(t, out)
		assertNoFile(t, dataPDF)
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		if _, statErr := os.Stat(pdfDir); !os.IsNotExist(statErr) {
			t.Fatalf("pdf_dir %s was created: %v", pdfDir, statErr)
		}
		wantStdoutLine(t, stdout, out)
	})
	t.Run("no-home", func(t *testing.T) {
		cwd := isolateCLI(t)
		t.Setenv("HOME", "")
		t.Setenv("XDG_DATA_HOME", "")
		pdfDir := filepath.Join(t.TempDir(), "charts")
		out := filepath.Join(t.TempDir(), "out.pdf")
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertChartFile(t, out)
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		if _, statErr := os.Stat(pdfDir); !os.IsNotExist(statErr) {
			t.Fatalf("pdf_dir %s was created: %v", pdfDir, statErr)
		}
		wantStdoutLine(t, stdout, out)
	})
}

func TestChartPDFRelativeOStaysInCwd(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliDataHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-o", "out.pdf", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, filepath.Join(cwd, "out.pdf"))
	wantStdoutLine(t, stdout, "out.pdf")
	assertNoFile(t, filepath.Join(pdfDir, "out.pdf"))
	assertNoFile(t, filepath.Join(pdfDir, chartPDFName))
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
}

func TestChartPDFDefaultFailureNoCwdCopy(t *testing.T) {
	t.Run("path-is-directory", func(t *testing.T) {
		cwd := isolateCLI(t)
		dataDir, dataPDF := cliDataHome(t)
		if err := os.MkdirAll(dataPDF, 0o700); err != nil {
			t.Fatal(err)
		}
		dbPath, cfgPath := cliPaths(t, cwd)
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
		if err == nil {
			t.Fatalf("export succeeded, want error; stdout %q", stdout)
		}
		wantNoStdoutPath(t, stdout)
		info, statErr := os.Stat(dataPDF)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if !info.IsDir() {
			t.Fatalf("%s is no longer a directory", dataPDF)
		}
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		if _, statErr := os.Stat(dataDir); statErr != nil {
			t.Fatal(statErr)
		}
	})
	t.Run("pdf-dir-is-file", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		pdfDir := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(pdfDir, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
		if err == nil {
			t.Fatalf("export succeeded, want error; stdout %q", stdout)
		}
		wantNoStdoutPath(t, stdout)
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		assertNoFile(t, dataPDF)
		info, statErr := os.Stat(pdfDir)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.IsDir() {
			t.Fatal("replaced the pdf_dir file with a directory")
		}
	})
}

func TestChartPDFDatabasePathDoesNotMovePDF(t *testing.T) {
	t.Run("db-flag", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		dbPath := filepath.Join(t.TempDir(), "elsewhere.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		if err := config.Write(cfgPath, config.Default()); err != nil {
			t.Fatal(err)
		}
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertPDFNotCwd(t, cwd, dataPDF)
		assertNoFile(t, filepath.Join(filepath.Dir(dbPath), chartPDFName))
		wantStdoutLine(t, stdout, dataPDF)
	})
	t.Run("hdtools-db", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		dbPath := filepath.Join(t.TempDir(), "elsewhere.db")
		t.Setenv("HDTOOLS_DB", dbPath)
		cfgPath := filepath.Join(cwd, "config.toml")
		if err := config.Write(cfgPath, config.Default()); err != nil {
			t.Fatal(err)
		}
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertPDFNotCwd(t, cwd, dataPDF)
		assertNoFile(t, filepath.Join(filepath.Dir(dbPath), chartPDFName))
		wantStdoutLine(t, stdout, dataPDF)
	})
}

func TestChartPDFNonStringPDFDir(t *testing.T) {
	body := "display_unit = \"kg\"\npdf_dir = 3\n"
	t.Run("no-o", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, body)
		if _, err := config.Load(cfgPath); err != nil {
			t.Fatalf("Load: %v", err)
		}
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
		if err == nil {
			t.Fatalf("export succeeded, want error naming pdf_dir; stdout %q", stdout)
		}
		if !strings.Contains(err.Error(), "pdf_dir") {
			t.Fatalf("error %q does not name pdf_dir", err)
		}
		wantNoStdoutPath(t, stdout)
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		assertNoFile(t, dataPDF)
	})
	t.Run("o", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliDataHome(t)
		out := filepath.Join(t.TempDir(), "out.pdf")
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, body)
		if _, err := config.Load(cfgPath); err != nil {
			t.Fatalf("Load: %v", err)
		}
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertChartFile(t, out)
		assertNoFile(t, filepath.Join(cwd, chartPDFName))
		assertNoFile(t, dataPDF)
		wantStdoutLine(t, stdout, out)
	})
}

func TestChartPDFOverwriteDefaultPath(t *testing.T) {
	cwd := isolateCLI(t)
	dataDir, dataPDF := cliDataHome(t)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dataPDF, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	for range 2 {
		if _, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath); err != nil {
			t.Fatal(err)
		}
	}
	assertPDFNotCwd(t, cwd, dataPDF)
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "1990-11-chart") || strings.HasSuffix(entry.Name(), ".pdf") {
			names = append(names, entry.Name())
		}
	}
	if len(names) != 1 || names[0] != chartPDFName {
		t.Fatalf("chart files = %v, want [%s]", names, chartPDFName)
	}
}

func TestChartPDFNoHomeFails(t *testing.T) {
	cwd := isolateCLI(t)
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	dbPath := filepath.Join(t.TempDir(), "elsewhere.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	if err := config.Write(cfgPath, config.Default()); err != nil {
		t.Fatal(err)
	}
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		note := "no working-directory PDF"
		if _, statErr := os.Stat(filepath.Join(cwd, chartPDFName)); statErr == nil {
			note = "wrote " + chartPDFName + " in the working directory"
		}
		t.Fatalf("export succeeded (%s), want an error naming the home directory; stdout %q", note, stdout)
	}
	if !strings.Contains(err.Error(), "home directory") {
		t.Fatalf("error %q does not name the home directory", err)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
}

func TestChartPDFHomePrefix(t *testing.T) {
	cwd := isolateCLI(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	_, dataPDF := cliDataHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = \"~/charts\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(home, "charts", chartPDFName)
	assertPDFNotCwd(t, cwd, pdf)
	assertNoFile(t, dataPDF)
	wantStdoutLine(t, stdout, pdf)
}

func TestChartPDFRelativeDirFails(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "charts", value: "charts"},
		{name: "dot", value: "."},
		{name: "charts-slash", value: "charts/"},
		{name: "dollar", value: "$XDG_DATA_HOME/charts"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cwd := isolateCLI(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			dataRoot := t.TempDir()
			t.Setenv("XDG_DATA_HOME", dataRoot)
			dbPath, cfgPath := cliPaths(t, cwd)
			writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(tc.value)+"\n")
			seedNovember(t, dbPath)
			stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
			if err == nil {
				t.Fatalf("export succeeded, want error; stdout %q", stdout)
			}
			wantNoStdoutPath(t, stdout)
			assertNoFile(t, filepath.Join(cwd, chartPDFName))
			assertNoFile(t, filepath.Join(cwd, "charts", chartPDFName))
			assertNoFile(t, filepath.Join(home, "charts", chartPDFName))
			assertNoFile(t, filepath.Join(home, ".local", "share", "hdtools", chartPDFName))
			assertNoFile(t, filepath.Join(dataRoot, "hdtools", chartPDFName))
			assertNoFile(t, filepath.Join(dataRoot, "charts", chartPDFName))
		})
	}
}

func TestChartPDFHomePrefixNoHome(t *testing.T) {
	cwd := isolateCLI(t)
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	dbPath := filepath.Join(cwd, "t.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = \"~/charts\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		note := "no working-directory PDF"
		if _, statErr := os.Stat(filepath.Join(cwd, chartPDFName)); statErr == nil {
			note = "wrote " + chartPDFName + " in the working directory"
		}
		t.Fatalf("export succeeded (%s), want error; stdout %q", note, stdout)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
	assertNoFile(t, filepath.Join(cwd, "charts", chartPDFName))
}

func TestChartPDFRelativeDataHomeFails(t *testing.T) {
	cwd := isolateCLI(t)
	t.Setenv("XDG_DATA_HOME", "data")
	dbPath := filepath.Join(cwd, "t.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatalf("export succeeded, want error; stdout %q", stdout)
	}
	if !strings.Contains(err.Error(), "not absolute") {
		t.Fatalf("error = %q, want a non-absolute directory", err)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
	assertNoFile(t, filepath.Join(cwd, "data", "hdtools", chartPDFName))
}

func TestChartPDFHomePrefixRelativeHomeFails(t *testing.T) {
	cwd := isolateCLI(t)
	t.Setenv("HOME", "home")
	t.Setenv("XDG_DATA_HOME", "")
	dbPath := filepath.Join(cwd, "t.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = \"~/charts\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatalf("export succeeded, want error; stdout %q", stdout)
	}
	if !strings.Contains(err.Error(), "not absolute") {
		t.Fatalf("error = %q, want a non-absolute directory", err)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
	assertNoFile(t, filepath.Join(cwd, "home", "charts", chartPDFName))
}

// TestModuleRequiresCharmV2 is the library line for the monthly Online
// chart. Docs name ntcharts in prose, so the scan stays on go.mod plus
// Go files under cmd and internal. The module root is this file's tree:
// other tests in this package change directory.
func TestModuleRequiresCharmV2(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	var problems []string
	if got := goVersion(text); got != "1.27.0" {
		problems = append(problems, fmt.Sprintf("go line = %q, want 1.27.0", got))
	}
	mods := requireModules(text)
	for _, want := range []string{
		"charm.land/bubbletea/v2",
		"charm.land/lipgloss/v2",
		"charm.land/bubbles/v2",
		"github.com/NimbleMarkets/ntcharts/v2",
	} {
		if !moduleListed(mods, want) {
			problems = append(problems, "require block missing "+want)
		}
	}
	for _, mod := range mods {
		switch mod {
		case "github.com/charmbracelet/bubbletea",
			"github.com/charmbracelet/lipgloss",
			"github.com/charmbracelet/bubbles":
			problems = append(problems, "require block still requires "+mod)
		}
		if mod == "github.com/sajari/regression" || strings.HasPrefix(mod, "github.com/sajari/regression/") {
			problems = append(problems, "require block requires "+mod)
		}
		if strings.Contains(mod, "ntcharts") && mod != "github.com/NimbleMarkets/ntcharts/v2" {
			problems = append(problems, "require block names ntcharts module "+mod)
		}
	}
	ntchartsImport := false
	for _, imp := range goImports(t, root) {
		for _, bad := range []string{
			"github.com/charmbracelet/bubbletea",
			"github.com/charmbracelet/lipgloss",
			"github.com/charmbracelet/bubbles",
		} {
			if imp.path == bad || strings.HasPrefix(imp.path, bad+"/") {
				problems = append(problems, imp.file+" imports "+imp.path)
			}
		}
		if imp.path == "github.com/sajari/regression" || strings.HasPrefix(imp.path, "github.com/sajari/regression/") {
			problems = append(problems, imp.file+" imports "+imp.path)
		}
		if strings.Contains(imp.path, "bubblezone") {
			problems = append(problems, imp.file+" imports "+imp.path)
		}
		if strings.Contains(imp.path, "ntcharts") {
			if imp.path == "github.com/NimbleMarkets/ntcharts/v2" || strings.HasPrefix(imp.path, "github.com/NimbleMarkets/ntcharts/v2/") {
				ntchartsImport = true
			} else {
				problems = append(problems, imp.file+" imports ntcharts "+imp.path)
			}
		}
	}
	if !ntchartsImport {
		problems = append(problems, "no import of github.com/NimbleMarkets/ntcharts/v2 under cmd/ or internal/")
	}
	if len(problems) > 0 {
		t.Fatalf("charm v2 module contract:\n%s", strings.Join(problems, "\n"))
	}
}

// TestProgramIsModelOnly reads main.go. The screen starts from the model
// alone: an alternate screen or a mouse mode would add pointer behaviour
// the monthly chart does not have.
func TestProgramIsModelOnly(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "cmd", "hdtools", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Contains(text, "AltScreen") || strings.Contains(text, "MouseMode") {
		t.Fatal("main.go mentions AltScreen or MouseMode")
	}
	args := programArgs(text)
	if len(args) == 0 {
		t.Fatal("main.go does not call tea.NewProgram")
	}
	for _, arg := range args {
		if commaOutside(arg) {
			t.Fatalf("tea.NewProgram(%s), want the model only", arg)
		}
	}
}

func programArgs(text string) []string {
	var args []string
	rest := text
	const key = "tea.NewProgram("
	for {
		i := strings.Index(rest, key)
		if i < 0 {
			return args
		}
		arg, next, ok := programArgAt(rest[i+len(key):])
		if !ok {
			return args
		}
		args = append(args, arg)
		rest = next
	}
}

func programArgAt(rest string) (arg, next string, ok bool) {
	depth := 1
	for j, r := range rest {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return rest[:j], rest[j+len(string(r)):], true
			}
		}
	}
	return "", "", false
}

func commaOutside(arg string) bool {
	depth := 0
	for _, r := range arg {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("module root: caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root: go.mod not found")
		}
		dir = parent
	}
}

func goVersion(goMod string) string {
	for _, line := range strings.Split(goMod, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

func requireModules(goMod string) []string {
	var mods []string
	inBlock := false
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "//") {
			continue
		}
		switch {
		case trim == "require (":
			inBlock = true
		case inBlock && trim == ")":
			inBlock = false
		case inBlock:
			mods = append(mods, strings.Fields(trim)[0])
		case strings.HasPrefix(trim, "require "):
			rest := strings.TrimSpace(strings.TrimPrefix(trim, "require"))
			if rest != "" {
				mods = append(mods, strings.Fields(rest)[0])
			}
		}
	}
	return mods
}

func moduleListed(mods []string, want string) bool {
	for _, mod := range mods {
		if mod == want {
			return true
		}
	}
	return false
}

type scannedImport struct {
	file string
	path string
}

func goImports(t *testing.T, root string) []scannedImport {
	t.Helper()
	var found []scannedImport
	for _, dir := range []string{"cmd", "internal"} {
		start := filepath.Join(root, dir)
		err := filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != start && (d.Name() == "vendor" || d.Name() == "testdata") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			for _, imp := range file.Imports {
				p, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				found = append(found, scannedImport{file: rel, path: p})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return found
}

func TestLogPDFDefaultWritesDataDir(t *testing.T) {
	cwd := isolateCLI(t)
	dataDir, logPDF := cliLogHome(t)
	t.Setenv("NO_COLOR", "1")
	dbPath := filepath.Join(cwd, "t.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\n\n[colors]\nweight = \"green\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, logPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
	assertNoFile(t, filepath.Join(dataDir, chartPDFName))
	wantNoLogTemp(t, dataDir)
	wantNoLogTemp(t, cwd)
	wantPortraitBox(t, logPDF)
	draw := scanSheet(t, readPDF(t, logPDF))
	if !hasLiteral(draw.spans, "November 1990 (kg)") {
		t.Fatalf("heading literals = %q, want November 1990 (kg)", spanTexts(draw.spans))
	}
	wantStdoutLine(t, stdout, logPDF)
}

func TestLogPDFDirAbsolute(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliLogHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(pdfDir, logPDFName)
	assertChartFile(t, pdf)
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	wantStdoutLine(t, stdout, pdf)
}

func TestLogPDFHomePrefix(t *testing.T) {
	cwd := isolateCLI(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	_, dataPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = \"~/charts\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(home, "charts", logPDFName)
	assertChartFile(t, pdf)
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	wantStdoutLine(t, stdout, pdf)
}

func TestLogPDFRelativeDirFails(t *testing.T) {
	cwd := isolateCLI(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	dataDir, dataPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = \"charts\"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatalf("export succeeded, want error; stdout %q", stdout)
	}
	if !strings.Contains(err.Error(), "pdf_dir") {
		t.Fatalf("error %q does not name pdf_dir", err)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	assertNoFile(t, filepath.Join(cwd, "charts", logPDFName))
	assertNoFile(t, filepath.Join(home, "charts", logPDFName))
	assertNoFile(t, filepath.Join(home, ".local", "share", "hdtools", logPDFName))
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(dataDir, chartPDFName))
}

func TestLogPDFNonStringPDFDir(t *testing.T) {
	body := "display_unit = \"kg\"\npdf_dir = 3\n"
	t.Run("no-o", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliLogHome(t)
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, body)
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
		if err == nil {
			t.Fatalf("export succeeded, want error naming pdf_dir; stdout %q", stdout)
		}
		if !strings.Contains(err.Error(), "pdf_dir") {
			t.Fatalf("error %q does not name pdf_dir", err)
		}
		wantNoStdoutPath(t, stdout)
		assertNoFile(t, filepath.Join(cwd, logPDFName))
		assertNoFile(t, dataPDF)
	})
	t.Run("o", func(t *testing.T) {
		cwd := isolateCLI(t)
		_, dataPDF := cliLogHome(t)
		out := filepath.Join(t.TempDir(), "out.pdf")
		dbPath := filepath.Join(cwd, "t.db")
		cfgPath := filepath.Join(cwd, "config.toml")
		writeConfigFile(t, cfgPath, body)
		seedNovember(t, dbPath)
		stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		assertChartFile(t, out)
		assertNoFile(t, filepath.Join(cwd, logPDFName))
		assertNoFile(t, dataPDF)
		wantStdoutLine(t, stdout, out)
	})
}

func TestLogPDFOAbsoluteWins(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliLogHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	out := filepath.Join(t.TempDir(), "out.pdf")
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-o", out, "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, out)
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	if _, statErr := os.Stat(pdfDir); !os.IsNotExist(statErr) {
		t.Fatalf("pdf_dir %s was created: %v", pdfDir, statErr)
	}
	wantStdoutLine(t, stdout, out)
}

func TestLogPDFRelativeOStaysInCwd(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliLogHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"kg\"\npdf_dir = "+strconv.Quote(pdfDir)+"\n")
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-o", "out.pdf", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, filepath.Join(cwd, "out.pdf"))
	wantStdoutLine(t, stdout, "out.pdf")
	assertNoFile(t, filepath.Join(pdfDir, "out.pdf"))
	assertNoFile(t, filepath.Join(pdfDir, logPDFName))
	assertNoFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
}

func TestLogPDFDatabasePathDoesNotMovePDF(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliLogHome(t)
	dbPath := filepath.Join(t.TempDir(), "elsewhere.db")
	cfgPath := filepath.Join(cwd, "config.toml")
	if err := config.Write(cfgPath, config.Default()); err != nil {
		t.Fatal(err)
	}
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	assertNoFile(t, filepath.Join(filepath.Dir(dbPath), logPDFName))
	assertNoFile(t, filepath.Join(filepath.Dir(dbPath), chartPDFName))
	wantStdoutLine(t, stdout, dataPDF)
}

func TestLogPDFFailureNoCwdCopy(t *testing.T) {
	cwd := isolateCLI(t)
	_, dataPDF := cliLogHome(t)
	if err := os.MkdirAll(dataPDF, 0o700); err != nil {
		t.Fatal(err)
	}
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatalf("export succeeded, want error; stdout %q", stdout)
	}
	if !strings.Contains(err.Error(), "log pdf:") {
		t.Fatalf("error %q does not name the log-pdf step in run", err)
	}
	wantNoStdoutPath(t, stdout)
	info, statErr := os.Stat(dataPDF)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !info.IsDir() {
		t.Fatalf("%s is no longer a directory", dataPDF)
	}
	assertNoFile(t, filepath.Join(cwd, logPDFName))
}

func TestLogPDFOverwrite(t *testing.T) {
	cwd := isolateCLI(t)
	dataDir, dataPDF := cliLogHome(t)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dataPDF, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	for range 2 {
		if _, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath); err != nil {
			t.Fatal(err)
		}
	}
	assertChartFile(t, dataPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	assertNoFile(t, filepath.Join(dataDir, chartPDFName))
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "1990-11") || strings.HasSuffix(name, ".pdf") || strings.HasPrefix(name, ".hdtools-log-") || strings.HasSuffix(name, ".bak") || strings.HasSuffix(name, "~") {
			names = append(names, name)
		}
	}
	if len(names) != 1 || names[0] != logPDFName {
		t.Fatalf("log files = %v, want [%s]", names, logPDFName)
	}
}

func TestLogPDFRejectsBadMonth(t *testing.T) {
	for _, month := range []string{"1990-13", "banana"} {
		t.Run(month, func(t *testing.T) {
			cwd := isolateCLI(t)
			dbPath := filepath.Join(cwd, "t.db")
			cfgPath := filepath.Join(cwd, "config.toml")
			out := filepath.Join(cwd, "out.pdf")
			stdout, err := runCLIOutput(t, "-log-pdf", month, "-o", out, "-db", dbPath, "-config", cfgPath)
			if err == nil || !strings.Contains(err.Error(), "want YYYY-MM") {
				t.Fatalf("error %q, want YYYY-MM", err)
			}
			wantNoStdoutPath(t, stdout)
			assertNoFile(t, out)
			assertNoFile(t, cfgPath)
			assertNoFile(t, dbPath)
		})
	}
}

func TestLogPDFBothFlagsWriteNothing(t *testing.T) {
	t.Run("both-months", func(t *testing.T) {
		assertLogAndChartWriteNothing(t, "1990-11", "1990-11")
	})
	t.Run("banana", func(t *testing.T) {
		assertLogAndChartWriteNothing(t, "banana", "1990-11")
	})
}

func assertLogAndChartWriteNothing(t *testing.T, logMonth, chartMonth string) {
	t.Helper()
	cwd := isolateCLI(t)
	dataDir, logPDF := cliLogHome(t)
	cfgPath := filepath.Join(cwd, "config.toml")
	dbPath := filepath.Join(cwd, "t.db")
	out := filepath.Join(cwd, "out.pdf")
	stdout, err := runCLIOutput(t, "-log-pdf", logMonth, "-chart-pdf", chartMonth, "-o", out, "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatal("want non-zero before config and the database")
	}
	if !strings.Contains(err.Error(), "log-pdf") || !strings.Contains(err.Error(), "chart-pdf") {
		t.Fatalf("error %q does not name both export flags", err)
	}
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, out)
	assertNoFile(t, logPDF)
	assertNoFile(t, filepath.Join(dataDir, chartPDFName))
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	assertNoFile(t, filepath.Join(cwd, chartPDFName))
	assertNoFile(t, cfgPath)
	assertNoFile(t, dbPath)
}

func TestOAloneStartsTUIWritesNothing(t *testing.T) {
	cwd := isolateCLI(t)
	_, logPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	out := filepath.Join(cwd, "out.pdf")

	// runCLIOutput treats a live TUI as a fatal. -o alone is that TUI,
	// so a one-second wait is the signal, and a returned error must not
	// be an export. Stdin is closed so a non-TTY run can return.
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	origIn := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() { os.Stdin = origIn })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origOut })

	errc := make(chan error, 1)
	go func() { errc <- run([]string{"-o", out, "-db", dbPath, "-config", cfgPath}) }()

	var runErr error
	returned := false
	select {
	case runErr = <-errc:
		returned = true
	case <-time.After(time.Second):
	}
	os.Stdout = origOut
	if returned {
		_ = w.Close()
	}
	_ = r.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	outb, _ := io.ReadAll(r)
	_ = r.Close()
	if returned {
		if runErr == nil {
			t.Fatal("run returned nil; -o alone must start the TUI")
		}
		msg := runErr.Error()
		if strings.Contains(msg, "log pdf") || strings.Contains(msg, "chart pdf") {
			t.Fatalf("export error %q; -o alone must not export", runErr)
		}
	}
	wantNoStdoutPath(t, string(outb))
	assertNoFile(t, out)
	assertNoFile(t, logPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
}

func TestChartPDFDoesNotWriteLogPDF(t *testing.T) {
	cwd := isolateCLI(t)
	dataDir, logPDF := cliLogHome(t)
	chartPDF := filepath.Join(dataDir, chartPDFName)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedNovember(t, dbPath)
	stdout, err := runCLIOutput(t, "-chart-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	assertChartFile(t, chartPDF)
	assertNoFile(t, logPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
	w, h, ok := pdfMediaBox(readPDF(t, chartPDF))
	if !ok || math.Abs(w-792) > 0.5 || math.Abs(h-612) > 0.5 {
		t.Fatalf("MediaBox = %.2f×%.2f, want 792×612", w, h)
	}
	wantStdoutLine(t, stdout, chartPDF)
}

func TestLogPDFDoesNotRewriteKilograms(t *testing.T) {
	cwd := isolateCLI(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	writeConfigFile(t, cfgPath, "display_unit = \"lb\"\n")
	day := time.Date(1990, 11, 1, 0, 0, 0, 0, time.UTC)
	seedLog(t, dbPath, day, kgPtr(80), 0, 0, false, "")
	if _, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath); err != nil {
		t.Fatal(err)
	}
	store, err := dailylog.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	got, err := store.Get(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || math.Abs(*got.Weight-80) > 1e-9 {
		t.Fatalf("stored weight = %v, want 80 kg", got.Weight)
	}
}

func TestLogPDFCarryFromEarlierMonth(t *testing.T) {
	cwd := isolateCLI(t)
	_, logPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedLog(t, dbPath, time.Date(1990, 10, 31, 0, 0, 0, 0, time.UTC), kgPtr(80), 0, 0, false, "")
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	wantStdoutLine(t, stdout, logPDF)
	draw := scanSheet(t, readPDF(t, logPDF))
	wantCell(t, draw.spans, "1", "trend", "80.0")
	wantBlank(t, draw.spans, "1", "weight")
}

func TestLogPDFEmptyLog(t *testing.T) {
	cwd := isolateCLI(t)
	_, logPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	wantStdoutLine(t, stdout, logPDF)
	raw := readPDF(t, logPDF)
	wantPortraitBox(t, logPDF)
	wantOnePage(t, raw)
	draw := scanSheet(t, raw)
	if !hasDay(draw.spans, "30") {
		t.Fatal("day 30 is not a row")
	}
	wantNoLiteral(t, draw.spans, "(empty month)")
}

func TestLogPDFNoWeighIn(t *testing.T) {
	cwd := isolateCLI(t)
	_, logPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedLog(t, dbPath, time.Date(1990, 11, 2, 0, 0, 0, 0, time.UTC), nil, 8, 100, false, "rest")
	seedLog(t, dbPath, time.Date(1990, 11, 4, 0, 0, 0, 0, time.UTC), nil, 7.5, 100000, true, "walk")
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	wantStdoutLine(t, stdout, logPDF)
	draw := scanSheet(t, readPDF(t, logPDF))
	wantCell(t, draw.spans, "2", "sleep", "8.0")
	wantCell(t, draw.spans, "2", "steps", "100")
	wantCell(t, draw.spans, "2", "workout", "no")
	wantCell(t, draw.spans, "2", "note", "rest")
	wantBlank(t, draw.spans, "2", "weight")
	wantBlank(t, draw.spans, "2", "trend")
	wantBlank(t, draw.spans, "2", "delta")
	wantCell(t, draw.spans, "4", "sleep", "7.5")
	wantCell(t, draw.spans, "4", "note", "walk")
	for d := 1; d <= 30; d++ {
		day := strconv.Itoa(d)
		wantBlank(t, draw.spans, day, "weight")
		wantBlank(t, draw.spans, day, "trend")
		wantBlank(t, draw.spans, day, "delta")
	}
}

func TestLogPDFTrendError(t *testing.T) {
	t.Run("duplicate-day", func(t *testing.T) {
		w80 := 80.0
		w81 := 81.0
		assertTrendOrderFails(t, []rawDay{
			{day: "1990-11-01", weight: &w80},
			{day: "1990-11-01", weight: &w81},
		})
	})
	t.Run("duplicate-nil-weight", func(t *testing.T) {
		assertTrendOrderFails(t, []rawDay{
			{day: "1990-11-01", sleep: 8, steps: 100, note: "a"},
			{day: "1990-11-01", sleep: 7, steps: 50, note: "b"},
		})
	})
}

func assertTrendOrderFails(t *testing.T, rows []rawDay) {
	t.Helper()
	cwd := isolateCLI(t)
	_, logPDF := cliLogHome(t)
	dbPath, cfgPath := cliPaths(t, cwd)
	seedDuplicateDays(t, dbPath, rows)
	stdout, err := runCLIOutput(t, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	if err == nil {
		t.Fatalf("export succeeded, want a series-order error; stdout %q", stdout)
	}
	wantSeriesOrderError(t, err)
	wantNoStdoutPath(t, stdout)
	assertNoFile(t, logPDF)
	assertNoFile(t, filepath.Join(cwd, logPDFName))
}

func wantSeriesOrderError(t *testing.T, err error) {
	t.Helper()
	msg := err.Error()
	for _, s := range []string{"duplicate", "unsorted", "chronological", "strictly", "not after"} {
		if strings.Contains(msg, s) {
			return
		}
	}
	t.Fatalf("error %q does not reject a duplicate or unsorted series", err)
}

type rawDay struct {
	day     string
	weight  *float64
	sleep   float64
	steps   int
	workout bool
	note    string
}

// seedDuplicateDays rebuilds daily_log without the day primary key.
// Upsert cannot store two rows for one day. store.All still returns both.
func seedDuplicateDays(t *testing.T, dbPath string, rows []rawDay) {
	t.Helper()
	store, err := dailylog.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`DROP TABLE daily_log`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE daily_log (
		day TEXT,
		weight REAL,
		sleep_hours REAL NOT NULL DEFAULT 0,
		steps INTEGER NOT NULL DEFAULT 0,
		workout INTEGER NOT NULL DEFAULT 0,
		note TEXT NOT NULL DEFAULT ''
	)`); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		var weight any
		if row.weight != nil {
			weight = *row.weight
		}
		workout := 0
		if row.workout {
			workout = 1
		}
		if _, err := db.Exec(`INSERT INTO daily_log (day, weight, sleep_hours, steps, workout, note)
VALUES (?, ?, ?, ?, ?, ?)`, row.day, weight, row.sleep, row.steps, workout, row.note); err != nil {
			t.Fatal(err)
		}
	}
}

func cliLogHome(t *testing.T) (dir, pdf string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	dir = filepath.Join(root, "hdtools")
	pdf = filepath.Join(dir, logPDFName)
	return dir, pdf
}

func seedLog(t *testing.T, dbPath string, day time.Time, weight *float64, sleep float64, steps int, workout bool, note string) {
	t.Helper()
	store, err := dailylog.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	log, err := dailylog.New(day, weight, sleep, steps, workout, note)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
}

func kgPtr(v float64) *float64 { return &v }

func wantNoLogTemp(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".hdtools-log-") {
			t.Fatalf("leftover temp %s", entry.Name())
		}
	}
}

func wantPortraitBox(t *testing.T, path string) {
	t.Helper()
	w, h, ok := pdfMediaBox(readPDF(t, path))
	if !ok {
		t.Fatal("missing MediaBox")
	}
	if math.Abs(w-612) > 0.5 || math.Abs(h-792) > 0.5 {
		t.Fatalf("MediaBox = %.2f×%.2f, want 612×792", w, h)
	}
}

func wantOnePage(t *testing.T, raw []byte) {
	t.Helper()
	if n := pdfPageCount(raw); n != 1 {
		t.Fatalf("pages = %d, want 1", n)
	}
}

func readPDF(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func scanSheet(t *testing.T, raw []byte) pdfDraw {
	t.Helper()
	return scanPDF(pdfStreamContent(t, raw))
}

func wantCell(t *testing.T, spans []pdfSpan, day, col, want string) {
	t.Helper()
	got, ok := rowCells(spans, day)[col]
	if !ok {
		t.Fatalf("day %s %s has no text, want %q", day, col, want)
	}
	if got != want {
		t.Fatalf("day %s %s = %q, want %q", day, col, got, want)
	}
}

func wantBlank(t *testing.T, spans []pdfSpan, day, col string) {
	t.Helper()
	if !hasDay(spans, day) {
		t.Fatalf("day %s is not a row", day)
	}
	if got, ok := rowCells(spans, day)[col]; ok {
		t.Fatalf("day %s %s = %q, want no text", day, col, got)
	}
}

func wantNoLiteral(t *testing.T, spans []pdfSpan, lit string) {
	t.Helper()
	for _, s := range spans {
		if s.text == lit {
			t.Fatalf("literal %q is on the page", lit)
		}
	}
}

func hasLiteral(spans []pdfSpan, lit string) bool {
	for _, s := range spans {
		if s.text == lit {
			return true
		}
	}
	return false
}

func hasDay(spans []pdfSpan, day string) bool {
	return rowCells(spans, day) != nil
}

func spanTexts(spans []pdfSpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.text)
	}
	return out
}

func rowCells(spans []pdfSpan, day string) map[string]string {
	var y float64
	found := false
	for _, s := range spans {
		col, ok := columnOf(s.x)
		if ok && col == "day" && s.text == day {
			y = s.y
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	cells := map[string]string{}
	for _, s := range spans {
		if math.Abs(s.y-y) >= 0.8 || s.text == "" {
			continue
		}
		col, ok := columnOf(s.x)
		if !ok {
			continue
		}
		if prev, exists := cells[col]; exists {
			cells[col] = prev + "|" + s.text
		} else {
			cells[col] = s.text
		}
	}
	return cells
}

// columnOf maps a text baseline onto the sheet geometry: 36 pt margin,
// then day 28, weekday 36, weight 40, trend 48, delta 44, sleep 36,
// steps 52, workout 36, note 220. A blank cell draws nothing, so the
// column is which band the baseline sits in.
func columnOf(x float64) (string, bool) {
	cols := []struct {
		name string
		w    float64
	}{
		{"day", 28},
		{"weekday", 36},
		{"weight", 40},
		{"trend", 48},
		{"delta", 44},
		{"sleep", 36},
		{"steps", 52},
		{"workout", 36},
		{"note", 220},
	}
	left := 36.0
	for _, c := range cols {
		right := left + c.w
		if x >= left-1 && x < right-0.5 {
			return c.name, true
		}
		left = right
	}
	return "", false
}

type pdfSpan struct {
	text string
	x, y float64
	pos  bool
}

type pdfDraw struct {
	spans       []pdfSpan
	horiz, vert int
	rects       int
}

func scanPDF(content []byte) pdfDraw {
	var d pdfDraw
	var nums []float64
	var x, y float64
	var havePos bool
	var cx, cy float64
	var haveC bool
	for i := 0; i < len(content); {
		c := content[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == 0:
			i++
		case c == '%':
			for i < len(content) && content[i] != '\n' && content[i] != '\r' {
				i++
			}
		case c == '(':
			s, next := readPDFString(content, i)
			i = next
			j := skipPDFSpace(content, i)
			if hasPDFOp(content[j:], "Tj") || hasPDFOp(content[j:], "TJ") {
				op := "Tj"
				if hasPDFOp(content[j:], "TJ") && !hasPDFOp(content[j:], "Tj") {
					op = "TJ"
				}
				d.spans = append(d.spans, pdfSpan{text: s, x: x, y: y, pos: havePos})
				i = j + len(op)
			}
		case c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9'):
			n, next, ok := readPDFNum(content, i)
			if !ok {
				i++
				continue
			}
			nums = append(nums, n)
			i = next
		case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			start := i
			i++
			for i < len(content) && isPDFOpByte(content[i]) {
				i++
			}
			applyPDFOp(&d, string(content[start:i]), &nums, &x, &y, &havePos, &cx, &cy, &haveC)
		case c == '/':
			i++
			for i < len(content) && !isPDFDelim(content[i]) {
				i++
			}
		case c == '<':
			i = skipPDFAngle(content, i)
		default:
			i++
		}
	}
	return d
}

func applyPDFOp(d *pdfDraw, op string, nums *[]float64, x, y *float64, havePos *bool, cx, cy *float64, haveC *bool) {
	switch op {
	case "Td", "TD":
		if len(*nums) >= 2 {
			*x = (*nums)[len(*nums)-2]
			*y = (*nums)[len(*nums)-1]
			*havePos = true
		}
	case "Tm":
		if len(*nums) >= 6 {
			*x = (*nums)[len(*nums)-2]
			*y = (*nums)[len(*nums)-1]
			*havePos = true
		}
	case "m":
		if len(*nums) >= 2 {
			*cx = (*nums)[len(*nums)-2]
			*cy = (*nums)[len(*nums)-1]
			*haveC = true
		}
	case "l":
		if *haveC && len(*nums) >= 2 {
			x2 := (*nums)[len(*nums)-2]
			y2 := (*nums)[len(*nums)-1]
			if math.Abs(y2-*cy) < 0.2 && math.Abs(x2-*cx) > 1 {
				d.horiz++
			}
			if math.Abs(x2-*cx) < 0.2 && math.Abs(y2-*cy) > 1 {
				d.vert++
			}
			*cx, *cy = x2, y2
		}
	case "re":
		d.rects++
	}
	*nums = nil
}

func readPDFString(content []byte, i int) (string, int) {
	i++
	var b strings.Builder
	for i < len(content) {
		if content[i] == '\\' && i+1 < len(content) {
			i++
			c := content[i]
			switch c {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case '(', ')', '\\':
				b.WriteByte(c)
			default:
				if c >= '0' && c <= '7' {
					oct := int(c - '0')
					for n := 0; n < 2 && i+1 < len(content) && content[i+1] >= '0' && content[i+1] <= '7'; n++ {
						i++
						oct = oct*8 + int(content[i]-'0')
					}
					b.WriteByte(byte(oct))
				} else {
					b.WriteByte(c)
				}
			}
			i++
			continue
		}
		if content[i] == ')' {
			return b.String(), i + 1
		}
		b.WriteByte(content[i])
		i++
	}
	return b.String(), i
}

func readPDFNum(content []byte, i int) (float64, int, bool) {
	start := i
	if content[i] == '+' || content[i] == '-' {
		i++
		if i >= len(content) {
			return 0, start, false
		}
	}
	saw := false
	for i < len(content) && content[i] >= '0' && content[i] <= '9' {
		saw = true
		i++
	}
	if i < len(content) && content[i] == '.' {
		i++
		for i < len(content) && content[i] >= '0' && content[i] <= '9' {
			saw = true
			i++
		}
	}
	if !saw {
		return 0, start, false
	}
	v, err := strconv.ParseFloat(string(content[start:i]), 64)
	if err != nil {
		return 0, start, false
	}
	return v, i, true
}

func skipPDFSpace(content []byte, i int) int {
	for i < len(content) {
		switch content[i] {
		case ' ', '\t', '\n', '\r', '\f', 0:
			i++
		default:
			return i
		}
	}
	return i
}

func hasPDFOp(content []byte, op string) bool {
	if !bytes.HasPrefix(content, []byte(op)) {
		return false
	}
	if len(content) == len(op) {
		return true
	}
	return isPDFDelim(content[len(op)])
}

func isPDFOpByte(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '*'
}

func isPDFDelim(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\f', 0, '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

func skipPDFAngle(content []byte, i int) int {
	if i+1 < len(content) && content[i+1] == '<' {
		i += 2
		for i+1 < len(content) && !(content[i] == '>' && content[i+1] == '>') {
			i++
		}
		if i+1 < len(content) {
			return i + 2
		}
		return len(content)
	}
	i++
	for i < len(content) && content[i] != '>' {
		i++
	}
	if i < len(content) {
		i++
	}
	return i
}

var mediaBoxRe = regexp.MustCompile(`/MediaBox\s*\[\s*[\d.]+\s+[\d.]+\s+([\d.]+)\s+([\d.]+)\s*\]`)

func pdfMediaBox(raw []byte) (w, h float64, ok bool) {
	m := mediaBoxRe.FindSubmatch(raw)
	if m == nil {
		return 0, 0, false
	}
	w, err1 := strconv.ParseFloat(string(m[1]), 64)
	h, err2 := strconv.ParseFloat(string(m[2]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return w, h, true
}

var pageTypeRe = regexp.MustCompile(`/Type\s*/Page\b`)

func pdfPageCount(raw []byte) int {
	return len(pageTypeRe.FindAll(raw, -1))
}

func pdfStreamContent(t *testing.T, raw []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	rest := raw
	for {
		i := bytes.Index(rest, []byte("stream"))
		if i < 0 {
			break
		}
		dictStart := bytes.LastIndex(rest[:i], []byte("<<"))
		dict := []byte{}
		if dictStart >= 0 {
			dict = rest[dictStart:i]
		}
		rest = rest[i+len("stream"):]
		if len(rest) > 0 && rest[0] == '\r' {
			rest = rest[1:]
		}
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		j := bytes.Index(rest, []byte("endstream"))
		if j < 0 {
			break
		}
		data := rest[:j]
		if bytes.HasSuffix(data, []byte("\r\n")) {
			data = data[:len(data)-2]
		} else if bytes.HasSuffix(data, []byte("\n")) {
			data = data[:len(data)-1]
		}
		if bytes.Contains(dict, []byte("FlateDecode")) {
			dec, err := zlib.NewReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("inflate content: %v", err)
			}
			plain, err := io.ReadAll(dec)
			_ = dec.Close()
			if err != nil {
				t.Fatalf("inflate content: %v", err)
			}
			data = plain
		}
		buf.Write(data)
		buf.WriteByte('\n')
		rest = rest[j+len("endstream"):]
	}
	return buf.Bytes()
}
