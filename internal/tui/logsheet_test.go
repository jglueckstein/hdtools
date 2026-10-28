package tui

import (
	"bytes"
	"compress/zlib"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	logPDFNovember = "1990-11-log.pdf"
	logPDFDecember = "1990-12-log.pdf"
	logPDFOctober  = "1990-10-log.pdf"
	monthLogHelp   = "arrows move   type edit   tab next   space workout   enter form   c chart   l long   P log   t today   [ ] month   esc list"
)

// applyShiftP mirrors applyP. The write returns no command; a quit is a failure.
func applyShiftP(t *testing.T, app *App) {
	t.Helper()
	_, cmd := app.Update(tea.KeyPressMsg{Code: 'P', Text: "P"})
	if cmd == nil {
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Fatal("P quit")
	}
	if msg != nil {
		app.Update(msg)
	}
}

// applyReload runs the load command savedMsg returned. During-reload does not call this.
func applyReload(t *testing.T, app *App, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("save did not start a reload")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Fatal("reload quit")
	}
	if msg != nil {
		app.Update(msg)
	}
}

func isolatePDFHome(t *testing.T) (cwd, dataDir string) {
	t.Helper()
	cwd = t.TempDir()
	t.Chdir(cwd)
	dataDir, _ = useDataHome(t)
	return cwd, dataDir
}

func openNovemberCell(t *testing.T, app *App, day, col int) {
	t.Helper()
	press(app, "m")
	if app.screen != screenMonth || app.month.year != 1990 || app.month.month != time.November {
		t.Fatalf("sheet = %s %d screen %v, want November 1990", app.month.month, app.month.year, app.screen)
	}
	app.month.day = day
	app.month.col = col
}

func novemberSheet(t *testing.T) (cwd, dataDir string, app *App) {
	t.Helper()
	cwd, dataDir = isolatePDFHome(t)
	app = twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	return cwd, dataDir, app
}

func mustLogs(t *testing.T, store *dailylog.Store) []dailylog.DailyLog {
	t.Helper()
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	return logs
}

func cloneLogs(logs []dailylog.DailyLog) []dailylog.DailyLog {
	out := make([]dailylog.DailyLog, len(logs))
	for i, log := range logs {
		out[i] = log
		if log.Weight != nil {
			w := *log.Weight
			out[i].Weight = &w
		}
	}
	return out
}

func sameLogs(a, b []dailylog.DailyLog) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Day.Equal(b[i].Day) || a[i].Trend != b[i].Trend || a[i].SleepHours != b[i].SleepHours || a[i].Steps != b[i].Steps || a[i].Workout != b[i].Workout || a[i].Note != b[i].Note {
			return false
		}
		if (a[i].Weight == nil) != (b[i].Weight == nil) {
			return false
		}
		if a[i].Weight != nil && *a[i].Weight != *b[i].Weight {
			return false
		}
	}
	return true
}

func upsertLog(t *testing.T, store *dailylog.Store, day time.Time, weight *float64, sleep float64, steps int, workout bool, note string) {
	t.Helper()
	log, err := dailylog.New(day, weight, sleep, steps, workout, note)
	if err != nil {
		t.Fatalf("new log %s: %v", day.Format(time.DateOnly), err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatalf("upsert %s: %v", day.Format(time.DateOnly), err)
	}
}

func assertSameCell(t *testing.T, app *App, month time.Month, day, col int) {
	t.Helper()
	if app.screen != screenMonth {
		t.Fatalf("screen = %v, want month", app.screen)
	}
	if app.month.editing {
		t.Fatalf("idle P started an edit (%q)", app.month.input.Value())
	}
	if app.month.year != 1990 || app.month.month != month || app.month.day != day || app.month.col != col {
		t.Fatalf("cell = %d %s %d col %d, want 1990 %s %d col %d", app.month.year, app.month.month, app.month.day, app.month.col, month, day, col)
	}
	view := visible(app.View().Content)
	label := month.String() + " 1990"
	if !strings.Contains(view, label) || !strings.Contains(view, "arrows move") {
		t.Fatalf("left the month sheet, want %s day %d col %d", label, day, col)
	}
}

func assertIdleWrote(t *testing.T, app *App, path, dir, name string, month time.Month, day, col int) {
	t.Helper()
	if app.month.editing {
		t.Fatalf("idle P started an edit (%q); %s was not written", app.month.input.Value(), name)
	}
	info, err := os.Stat(path)
	if err != nil {
		if col == colWorkout {
			t.Fatalf("idle P on workout did nothing: %s was not written: %v", name, err)
		}
		t.Fatalf("log pdf %s: %v", path, err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("%s is not a regular file", path)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatalf("%s does not start with %%PDF", path)
	}
	if !strings.Contains(app.status, dir) || !strings.Contains(app.status, name) {
		t.Fatalf("status = %q, want %s and %s", app.status, dir, name)
	}
	view := visible(app.View().Content)
	if !strings.Contains(view, dir) || !strings.Contains(view, name) {
		t.Fatalf("status line = %q, want %s and %s", app.status, dir, name)
	}
	if app.month.err != "" || app.err != nil || strings.Contains(view, "error:") {
		t.Fatalf("error line %q app error %v", app.month.err, app.err)
	}
	assertSameCell(t, app, month, day, col)
}

func assertLogWriteError(t *testing.T, app *App, parts ...string) {
	t.Helper()
	view := visible(app.View().Content)
	if strings.Contains(view, "write chart pdf") {
		t.Fatalf("month error %q uses the chart writer prefix", app.month.err)
	}
	var missing []string
	if !strings.Contains(view, "write log pdf") {
		missing = append(missing, "write log pdf")
	}
	for _, part := range parts {
		if !strings.Contains(view, part) {
			missing = append(missing, part)
		}
	}
	if len(missing) > 0 {
		want := strings.Join(missing, ", ")
		if app.month.editing {
			t.Fatalf("idle P started an edit (%q); month error %q status %q, want write log pdf (%s)", app.month.input.Value(), app.month.err, app.status, want)
		}
		t.Fatalf("month error %q status %q, want write log pdf (%s)", app.month.err, app.status, want)
	}
	if app.status != "" {
		t.Fatalf("status = %q, want empty", app.status)
	}
	if app.err != nil {
		t.Fatalf("app error = %v, want nil", app.err)
	}
}

func assertNoPDFFiles(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
				t.Fatalf("wrote %s", filepath.Join(dir, entry.Name()))
			}
		}
	}
}

func assertOnlyLogPDF(t *testing.T, dir, name string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var got []string
	for _, entry := range entries {
		n := entry.Name()
		if n == name || strings.HasSuffix(n, ".pdf") || strings.HasSuffix(n, ".bak") || strings.HasSuffix(n, "~") || strings.HasPrefix(n, ".hdtools-") {
			got = append(got, n)
		}
	}
	if len(got) != 1 || got[0] != name {
		t.Fatalf("pdf files = %v, want [%s]", got, name)
	}
}

func monthHelpLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "arrows move") {
			return line
		}
	}
	return ""
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("find test file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root %s: %v", root, err)
	}
	return root
}

// buildHDTools compiles the real command before the test changes HOME.
// A later go run would put the module cache under that temp home, and
// TempDir could not remove the read-only cache files.
func buildHDTools(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "hdtools")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/hdtools")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build hdtools: %v\n%s", err, out)
	}
	return bin
}

// runLogPDFCommand is hdtools -log-pdf with an empty -o, against dbPath.
// The page is the command's page. It is not logpdf.Write of the sheet's rows.
func runLogPDFCommand(t *testing.T, bin, dbPath, cfgPath string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-log-pdf", "1990-11", "-db", dbPath, "-config", cfgPath)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hdtools -log-pdf: %v\n%s", err, out)
	}
	return string(out)
}

// pdfContentStreams is the page content, not the file bytes.
// CreationDate and ModDate live outside the stream; a copy inside a stream is dropped.
func pdfContentStreams(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var buf bytes.Buffer
	rest := raw
	for {
		i := bytes.Index(rest, []byte("stream"))
		if i < 0 {
			break
		}
		dictStart := bytes.LastIndex(rest[:i], []byte("<<"))
		var dict []byte
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
			dec, zerr := zlib.NewReader(bytes.NewReader(data))
			if zerr != nil {
				t.Fatalf("inflate %s: %v", path, zerr)
			}
			plain, rerr := io.ReadAll(dec)
			_ = dec.Close()
			if rerr != nil {
				t.Fatalf("inflate %s: %v", path, rerr)
			}
			data = plain
		}
		buf.Write(stripPDFDates(data))
		buf.WriteByte('\n')
		rest = rest[j+len("endstream"):]
	}
	if buf.Len() == 0 {
		t.Fatalf("%s has no page content stream", path)
	}
	return buf.Bytes()
}

func stripPDFDates(b []byte) []byte {
	if !bytes.Contains(b, []byte("/CreationDate")) && !bytes.Contains(b, []byte("/ModDate")) {
		return b
	}
	var out bytes.Buffer
	for _, line := range bytes.Split(b, []byte("\n")) {
		if bytes.Contains(line, []byte("/CreationDate")) || bytes.Contains(line, []byte("/ModDate")) {
			continue
		}
		out.Write(line)
		out.WriteByte('\n')
	}
	return out.Bytes()
}

func TestMonthShiftPWritesLogPDF(t *testing.T) {
	cwd, dataDir, app := novemberSheet(t)
	dbDir := t.TempDir()
	app.dbPath = filepath.Join(dbDir, "hdtools.db")
	day, col := app.month.day, app.month.col
	stored := cloneLogs(mustLogs(t, app.store))
	applyShiftP(t, app)
	path := filepath.Join(dataDir, logPDFNovember)
	assertIdleWrote(t, app, path, dataDir, logPDFNovember, time.November, day, col)
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, filepath.Join(dataDir, chartPDFFile))
	assertNoPDF(t, filepath.Join(dbDir, logPDFNovember))
	if !sameLogs(mustLogs(t, app.store), stored) {
		t.Fatal("stored logs changed")
	}
}

func TestMonthShiftPUsesPDFDir(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "charts")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	path := filepath.Join(pdfDir, logPDFNovember)
	assertIdleWrote(t, app, path, pdfDir, logPDFNovember, time.November, day, col)
	assertNoPDF(t, filepath.Join(dataDir, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
}

func TestMonthShiftPHomePrefix(t *testing.T) {
	cwd, _ := isolatePDFHome(t)
	home := os.Getenv("HOME")
	pdfDir := filepath.Join(home, "charts")
	app := appWithConfig(t, configWithPDFDir(t, "~/charts"))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	path := filepath.Join(pdfDir, logPDFNovember)
	assertIdleWrote(t, app, path, pdfDir, logPDFNovember, time.November, day, col)
	if strings.Contains(app.status, "~") || strings.Contains(visible(app.View().Content), "~") {
		t.Fatalf("status %q still has a tilde", app.status)
	}
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
}

func TestMonthShiftPAfterNextMonth(t *testing.T) {
	_, dataDir, app := novemberSheet(t)
	app.month.day = 15
	app.month.col = colSteps
	press(app, "]")
	if app.month.month != time.December || app.month.day != 15 || app.month.col != colSteps {
		t.Fatalf("after ]: %s %d day %d col %d", app.month.month, app.month.year, app.month.day, app.month.col)
	}
	applyShiftP(t, app)
	path := filepath.Join(dataDir, logPDFDecember)
	assertIdleWrote(t, app, path, dataDir, logPDFDecember, time.December, 15, colSteps)
	assertNoPDF(t, filepath.Join(dataDir, logPDFNovember))
	assertNoPDF(t, filepath.Join(dataDir, chartPDFFile))
}

func TestMonthShiftPAfterPrevMonth(t *testing.T) {
	_, dataDir, app := novemberSheet(t)
	app.month.day = 15
	app.month.col = colSteps
	press(app, "[")
	if app.month.month != time.October || app.month.day != 15 || app.month.col != colSteps {
		t.Fatalf("after [: %s %d day %d col %d", app.month.month, app.month.year, app.month.day, app.month.col)
	}
	applyShiftP(t, app)
	path := filepath.Join(dataDir, logPDFOctober)
	assertIdleWrote(t, app, path, dataDir, logPDFOctober, time.October, 15, colSteps)
	assertNoPDF(t, filepath.Join(dataDir, logPDFNovember))
	assertNoPDF(t, filepath.Join(dataDir, "1990-10-chart.pdf"))
}

func TestMonthShiftPOverwrite(t *testing.T) {
	cwd, dataDir, app := novemberSheet(t)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	dest := filepath.Join(dataDir, logPDFNovember)
	if err := os.WriteFile(dest, []byte("not a pdf"), 0o644); err != nil {
		t.Fatalf("seed destination: %v", err)
	}
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertIdleWrote(t, app, dest, dataDir, logPDFNovember, time.November, day, col)
	assertOnlyLogPDF(t, dataDir, logPDFNovember)
	assertNoPDF(t, filepath.Join(dataDir, chartPDFFile))
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
}

func TestMonthShiftPMatchesCommand(t *testing.T) {
	// The sheet keeps the series twoDayApp loaded. The database then
	// gains a weightless 31 October, ahead of those weigh-ins, and a
	// 4 November row the sheet never loaded. The command reads that
	// database with an empty -o. Painting the sheet's rows would not
	// draw 4 November, and ApplyTrend on the whole series refuses the
	// leading day.
	bin := buildHDTools(t)
	_, dataDir := isolatePDFHome(t)
	dbPath := filepath.Join(t.TempDir(), "logs.db")
	store, err := dailylog.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	app := twoDayApp(t, store)
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	screen := cloneLogs(app.logs)

	upsertLog(t, store, dateUTC(1990, 10, 31), nil, 4, 20, false, "gap")
	later := 170.0
	upsertLog(t, store, dateUTC(1990, 11, 4), &later, 7, 2500, true, "later")
	if !sameLogs(app.logs, screen) {
		t.Fatal("sheet reloaded before shift-P")
	}

	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(cfgPath, []byte("display_unit = \"lb\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DisplayUnit != units.Pound {
		t.Fatalf("display unit = %q, want lb", cfg.DisplayUnit)
	}
	app.cfg = cfg

	out := runLogPDFCommand(t, bin, dbPath, cfgPath)
	keyPath := filepath.Join(dataDir, logPDFNovember)
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("command did not write %s: %v\n%s", keyPath, err, out)
	}
	oracle := filepath.Join(t.TempDir(), "command.pdf")
	if err := os.Rename(keyPath, oracle); err != nil {
		t.Fatalf("move command pdf: %v", err)
	}
	stored := cloneLogs(mustLogs(t, store))

	applyShiftP(t, app)
	if _, err := os.Stat(keyPath); err != nil {
		if app.month.editing {
			t.Fatalf("idle P started an edit (%q) and did not write %s", app.month.input.Value(), logPDFNovember)
		}
		t.Fatalf("log pdf %s: %v", keyPath, err)
	}
	got := pdfContentStreams(t, keyPath)
	want := pdfContentStreams(t, oracle)
	if !bytes.Equal(got, want) {
		t.Fatalf("page content streams differ from -log-pdf (%d bytes vs %d)", len(got), len(want))
	}
	if !sameLogs(app.logs, screen) {
		t.Fatalf("sheet rows changed (%d to %d)", len(screen), len(app.logs))
	}
	if !sameLogs(mustLogs(t, store), stored) {
		t.Fatal("stored logs changed")
	}
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPWriteFailure(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	logPath := filepath.Join(dataDir, logPDFNovember)
	if err := os.MkdirAll(logPath, 0o700); err != nil {
		t.Fatalf("block destination: %v", err)
	}
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	_, cmd := app.Update(savedMsg{})
	applyReload(t, app, cmd)
	if app.status != "saved" {
		t.Fatalf("status = %q, want sticky saved", app.status)
	}
	applyShiftP(t, app)
	assertLogWriteError(t, app)
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("destination: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("destination directory was replaced")
	}
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPPDFDirIsFile(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(pdfDir, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed pdf_dir: %v", err)
	}
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	openNovemberCell(t, app, 15, colSleep)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertLogWriteError(t, app)
	info, err := os.Stat(pdfDir)
	if err != nil {
		t.Fatalf("pdf_dir: %v", err)
	}
	if info.IsDir() {
		t.Fatal("replaced the pdf_dir file with a directory")
	}
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(dataDir, logPDFNovember))
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPNonStringPDFDir(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := appWithConfig(t, loadConfigTOML(t, "display_unit = \"kg\"\npdf_dir = 3\n"))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertLogWriteError(t, app, "pdf_dir must be a string")
	assertNoPDFFiles(t, cwd, dataDir)
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPRelativeDirFails(t *testing.T) {
	cwd, _ := isolatePDFHome(t)
	app := appWithConfig(t, loadConfigTOML(t, "display_unit = \"kg\"\npdf_dir = \"charts\"\n"))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertLogWriteError(t, app, "pdf_dir", "must be absolute")
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, "charts", logPDFNovember))
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPNoHomeFails(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	app := twoDayApp(t, openStore(t))
	app.dbPath = filepath.Join(t.TempDir(), "explicit.db")
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertLogWriteError(t, app, "home directory")
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPRelativeDataHomeFails(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "data")
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	assertLogWriteError(t, app, "not absolute")
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, "data", "hdtools", logPDFNovember))
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPDuringReload(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	_, cmd := app.Update(savedMsg{})
	if cmd == nil {
		t.Fatal("save did not start a reload")
	}
	// Leave the reload command unrun so the reload stays in flight.
	if app.status != "saved" {
		t.Fatalf("status = %q, want saved", app.status)
	}
	day, col := app.month.day, app.month.col
	applyShiftP(t, app)
	if app.month.editing {
		t.Fatalf("shift-P started an edit (%q) while reload was in flight", app.month.input.Value())
	}
	if app.status != "saved" {
		t.Fatalf("status = %q, want saved", app.status)
	}
	assertNoPDFFiles(t, cwd, dataDir)
	assertSameCell(t, app, time.November, day, col)
}

func TestMonthShiftPSaveDoesNotCoverPath(t *testing.T) {
	_, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	want := filepath.Join(dataDir, logPDFNovember)
	applyShiftP(t, app)
	_, cmd := app.Update(savedMsg{})
	applyReload(t, app, cmd)
	if app.status != want {
		t.Fatalf("status = %q, want %q", app.status, want)
	}
	if strings.Contains(app.status, "saved") {
		t.Fatalf("status %q says saved", app.status)
	}
}

func TestMonthPStartsEditWritesNoPDF(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	weight := appWithConfig(t, configWithPDFDir(t, pdfDir))
	openNovemberCell(t, weight, 15, colWeight)
	applyP(t, weight)
	if !weight.month.editing || !strings.Contains(weight.month.input.Value(), "p") {
		t.Fatalf("input = %q editing=%v, want an edit containing p", weight.month.input.Value(), weight.month.editing)
	}
	assertNoPDFFiles(t, cwd, dataDir, pdfDir)

	workout := appWithConfig(t, configWithPDFDir(t, pdfDir))
	openNovemberCell(t, workout, 15, colWorkout)
	applyP(t, workout)
	if workout.month.editing {
		t.Fatalf("p on workout started an edit %q", workout.month.input.Value())
	}
	if workout.month.col != colWorkout {
		t.Fatalf("col = %d, want workout", workout.month.col)
	}
	assertNoPDFFiles(t, cwd, dataDir, pdfDir)
}

func TestChartPWritesChartNotLogPDF(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	press(app, "c")
	applyP(t, app)
	chartPath := filepath.Join(dataDir, chartPDFFile)
	info, err := os.Stat(chartPath)
	if err != nil {
		t.Fatalf("chart pdf: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", info.Mode().Perm())
	}
	body, err := os.ReadFile(chartPath)
	if err != nil {
		t.Fatalf("read chart pdf: %v", err)
	}
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatal("chart pdf does not start with %PDF")
	}
	assertNoPDF(t, filepath.Join(dataDir, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	view := visible(app.View().Content)
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatal("left the monthly chart")
	}
}

func TestMonthShiftPWhileEditingIsText(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	for _, col := range []int{colWeight, colSteps, colNote} {
		if app.month.editing {
			press(app, "esc")
		}
		app.month.day = 15
		app.month.col = col
		app.month.beginEdit("")
		applyShiftP(t, app)
		if !app.month.editing || app.month.col != col {
			t.Fatalf("col %d editing=%v, want that edit still open", col, app.month.editing)
		}
		if !strings.Contains(app.month.input.Value(), "P") {
			t.Fatalf("col %d input = %q, want P", col, app.month.input.Value())
		}
		assertNoPDFFiles(t, cwd, dataDir)
	}
}

func TestMonthShiftPOnWorkoutWrites(t *testing.T) {
	cwd, dataDir, app := novemberSheet(t)
	app.month.day = 15
	app.month.col = colWorkout
	applyShiftP(t, app)
	path := filepath.Join(dataDir, logPDFNovember)
	assertIdleWrote(t, app, path, dataDir, logPDFNovember, time.November, 15, colWorkout)
	assertNoPDF(t, filepath.Join(cwd, logPDFNovember))
	assertNoPDF(t, filepath.Join(cwd, chartPDFFile))
	assertNoPDF(t, filepath.Join(dataDir, chartPDFFile))
}

func TestMonthShiftPIgnoredOnList(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	applyShiftP(t, app)
	assertNoPDFFiles(t, cwd, dataDir, pdfDir)
	if app.screen != screenList || !strings.Contains(visible(app.View().Content), "daily log") {
		t.Fatalf("left the list: screen %v", app.screen)
	}
}

func TestMonthShiftPOnFormNoteIsText(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	press(app, "n")
	if app.screen != screenForm {
		t.Fatalf("screen = %v, want form", app.screen)
	}
	press(app, "tab", "tab", "tab", "tab", "tab")
	if app.form.focus != fieldNote {
		t.Fatalf("focus = %d, want note", app.form.focus)
	}
	applyShiftP(t, app)
	if !strings.Contains(app.form.inputs[4].Value(), "P") {
		t.Fatalf("note = %q, want P", app.form.inputs[4].Value())
	}
	assertNoPDFFiles(t, cwd, dataDir, pdfDir)
	if app.screen != screenForm || !strings.Contains(visible(app.View().Content), "day form") {
		t.Fatalf("left the form: screen %v", app.screen)
	}
}

func TestMonthShiftPOnFormWorkoutWritesNothing(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	pdfDir := filepath.Join(t.TempDir(), "pdfs")
	app := appWithConfig(t, configWithPDFDir(t, pdfDir))
	press(app, "n")
	if app.screen != screenForm {
		t.Fatalf("screen = %v, want form", app.screen)
	}
	press(app, "tab", "tab", "tab", "tab")
	if app.form.focus != fieldWorkout {
		t.Fatalf("focus = %d, want workout", app.form.focus)
	}
	beforeWorkout := app.form.workout
	before := make([]string, len(app.form.inputs))
	for i := range app.form.inputs {
		before[i] = app.form.inputs[i].Value()
	}
	applyShiftP(t, app)
	if app.form.workout != beforeWorkout {
		t.Fatalf("workout = %v, want %v", app.form.workout, beforeWorkout)
	}
	for i := range app.form.inputs {
		if app.form.inputs[i].Value() != before[i] {
			t.Fatalf("field %d = %q, want %q", i, app.form.inputs[i].Value(), before[i])
		}
	}
	assertNoPDFFiles(t, cwd, dataDir, pdfDir)
	if app.screen != screenForm || !strings.Contains(visible(app.View().Content), "day form") {
		t.Fatalf("left the form: screen %v", app.screen)
	}
}

func TestMonthShiftPIgnoredOnChart(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	press(app, "c")
	applyShiftP(t, app)
	assertNoPDFFiles(t, cwd, dataDir)
	view := visible(app.View().Content)
	if app.screen != screenChart || !strings.Contains(view, "November 1990") || strings.Contains(view, "arrows move") {
		t.Fatal("left the monthly chart")
	}
}

func TestMonthShiftPIgnoredOnLong(t *testing.T) {
	cwd, dataDir := isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	press(app, "l")
	applyShiftP(t, app)
	assertNoPDFFiles(t, cwd, dataDir)
	view := visible(app.View().Content)
	if app.screen != screenLong || !strings.Contains(view, "Quarterly") {
		t.Fatal("left the long-term chart")
	}
}

func TestMonthHelpNamesLogKey(t *testing.T) {
	isolatePDFHome(t)
	app := twoDayApp(t, openStore(t))
	openNovemberCell(t, app, 15, colWeight)
	got := monthHelpLine(visible(app.View().Content))
	if got != monthLogHelp {
		t.Fatalf("help = %q, want %q", got, monthLogHelp)
	}
}
