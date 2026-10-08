package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
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

// TestModuleRequiresCharmV2 is S1. Docs name ntcharts in prose, so the
// scan stays on go.mod plus Go files under cmd and internal. The module
// root is this file's tree: other tests in this package change directory.
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
		if strings.Contains(mod, "ntcharts") {
			problems = append(problems, "require block names ntcharts module "+mod)
		}
	}
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
		if strings.Contains(imp.path, "ntcharts") {
			problems = append(problems, imp.file+" imports ntcharts "+imp.path)
		}
	}
	if len(problems) > 0 {
		t.Fatalf("charm v2 module contract:\n%s", strings.Join(problems, "\n"))
	}
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
