---
spec: docs/superpowers/specs/2026-09-12-pdf-monthly-charts.md
date: 2026-09-14
mode: code
diaboli_model: grok-4.6
objections:
  - id: O1
    category: risk
    severity: high
    claim: "Write creates the PDF with os.Create (0666 & umask) and chmods 0600 only after a successful OutputFileAndClose, so a failed or interrupted write leaves a world-readable weight chart and a failed overwrite destroys the previous 0600 file."
    evidence: "internal/chartpdf/chartpdf.go Write: pdf.OutputFileAndClose(path) then os.Chmod(path, 0o600) only on success. github.com/go-pdf/fpdf OutputFileAndClose: pdfFile, err := os.Create(fileStr) (truncates, mode 0666). Contrast internal/dailylog/store.go ensureOwnerOnlyFile: os.OpenFile(..., os.O_RDWR|os.O_CREATE, 0o600) then chmod."
    disposition: accepted
    disposition_rationale: "Write into a same-dir temp with mode 0600, then Rename onto the destination. Do not os.Create the final path (truncates, 0666) and chmod after. Failed writes leave the previous file; the new file is owner-only from the first byte, matching ensureOwnerOnlyFile."
  - id: O2
    category: implementation
    severity: medium
    claim: "internal/tui imports internal/chartpdf (and therefore github.com/go-pdf/fpdf), so Update writes PDFs and the TUI package grew the PDF dependency Decision 2 / FR7 / accepted spec-time O7 forbade."
    evidence: "internal/tui/chart.go: 'this file does not import a PDF library' then import github.com/jglueckstein/hdtools/internal/chartpdf and chartpdf.Write from writeChartPDF, called synchronously from updateChart case p. Spec Decision 2 / FR7; spec-time O7 disposition in docs/superpowers/objections/pdf-monthly-charts.md."
    disposition: rejected
    disposition_rationale: "FR7 / spec-time O7 forbid importing the PDF library (fpdf), not the writer. The approved plan puts p in chart.go calling chartpdf.Write. internal/tui does not import fpdf, same as it does not import database/sql while still importing dailylog."
  - id: O3
    category: risk
    severity: medium
    claim: "A bad -chart-pdf value is parsed only after config.Ensure and dailylog.Open, so 1990-13 / banana still create config.toml and an empty SQLite file before the non-zero exit."
    evidence: "cmd/hdtools/main.go run(): config.Ensure, os.MkdirAll(filepath.Dir(path), 0o755), dailylog.Open(path), then if *chartPDF != \"\" { exportChartPDF(...) }. parseMonth lives inside exportChartPDF, after the store is open. FR5 / S8: non-zero exit, no TUI, no file."
    disposition: accepted
    disposition_rationale: "parseMonth runs immediately after flag parse, before Ensure / Open. Invalid -chart-pdf exits non-zero with no config, no DB, no PDF. A valid month still loads config and the store."
  - id: O4
    category: risk
    severity: medium
    claim: "Decision 9's CLI write-failure path (non-zero exit, no TUI) has no cmd test, so a regression that swallows Write errors or starts Bubble Tea after a failed export would not be caught."
    evidence: "cmd/hdtools/main_test.go covers TestParseMonthRejects and TestChartPDFFlagSkipsTUI (success: %PDF, mode 0600, returns before 5s). No CLI test that makes Write fail. Plan: 'invalid month is non-zero and creates no file; CLI write error is non-zero.' TUI S3a is tested; CLI is not."
    disposition: accepted
    disposition_rationale: "Add a cmd test: unwritable -o (directory, like TUI S3a), non-zero exit, no TUI (returns before the 5s timeout), no success-path PDF."
  - id: O5
    category: risk
    severity: low
    claim: "S11's clip inspector treats any extractable integer 11–31 as a day-number label, so a correct current-month PDF whose Daily deficit (or another literal) is in that range fails CI — a requirement only the test states."
    evidence: "internal/chartpdf/chartpdf_test.go TestWritePDFCurrentMonthClipsAtToday: for _, s := range pdfLiteralStrings(...) { n, err := strconv.Atoi(s); if n > 10 && n <= 31 { t.Fatalf(\"day-number label %d after today\") } }. Spec S11 only forbids a day-number label after 10."
    disposition: accepted
    disposition_rationale: "Stop treating every integer 11-31 as a day label. Assert day 10 is labeled, November's last calendar day 30 is not, and keep the divisor-10 check (772 cal). Clip itself stays in chartspan tests."
---

## O1 — risk — high

### Claim

`chartpdf.Write` does not create the PDF owner-only. It lets
`fpdf.OutputFileAndClose` `os.Create` the path (mode `0666` masked by
umask, typically `0644`), then `chmod 0600` only if that call returns
nil. A failed or interrupted write therefore leaves a world-readable
personal weight chart; a failed overwrite truncates the previous
`0600` file first and may replace it with `0644` garbage. The
database already avoids this by `OpenFile(..., 0600)`.

### Evidence

`internal/chartpdf/chartpdf.go` `Write`:

> `pdf.OutputFileAndClose(path)` then `os.Chmod(path, 0o600)` only on
> success.

`fpdf.OutputFileAndClose` (v0.9.0) uses `os.Create` (`O_RDWR|O_CREATE|O_TRUNC`,
mode `0666`). Truncation happens before any bytes are written. `Chmod`
is skipped when `OutputFileAndClose` returns an error, including after
a successful `Create`.

The log store, which Decision 9 said the PDF must match, creates with
`0600` from the first byte (`internal/dailylog/store.go`
`ensureOwnerOnlyFile`: `os.OpenFile(..., os.O_RDWR|os.O_CREATE, 0o600)`
then chmod). Tests only `Stat` mode after a successful write.

### Why this matters

Decision 9 / FR8 treat `0600` as load-bearing for the same reason as
the database: a monthly weight chart is a personal series. On a shared
Unix host the success path has a world-readable window; the failure
path has no `chmod` at all. TUI `p` overwrite of an existing
`YYYY-MM-chart.pdf` is non-atomic: `Create` truncates first, so a
disk-full or kill during `Output` destroys the previous chart and can
leave a truncated `0644` file while `writeChartPDF` sets an error
status and does not claim saved. The caller is told the write failed;
the world-readable remnant is still there.

## O2 — implementation — medium

### Claim

`internal/tui` imports the PDF writer and calls it from `Update`, so
the TUI package grew a compile-time `fpdf` dependency and a
filesystem side-effect. That contradicts Decision 2, FR7, and the
accepted spec-time O7 disposition that `internal/tui` must not import
a PDF library. `chartspan` fixed data-rule duplication; the write
boundary was not kept.

### Evidence

`internal/tui/chart.go` says “this file does not import a PDF library”
then imports `internal/chartpdf` and calls `chartpdf.Write` from
`writeChartPDF`, invoked synchronously from `updateChart` on `p`.
`internal/chartpdf` imports `github.com/go-pdf/fpdf`. Same package
`tui`, `app.go` still says this package must not import a PDF library.

Spec Decision 2: the writer lives under `internal/`, “not in
`internal/tui`, which must stay SQL-free and must not grow a PDF
dependency.” FR7: “not in `internal/tui` as a PDF library import.”
Spec-time O7 (accepted): “internal/tui still must not import a PDF
library.” This is not a re-litigation of that disposition; the code
does not honour it.

### Why this matters

`go test ./internal/tui` now compiles `fpdf`. `p` is a blocking
`os.Getwd` + write inside `Update`, not a `tea.Cmd`, so a hung cwd
(NFS) freezes the TUI and any test that sends `p` without `t.Chdir`
writes `YYYY-MM-chart.pdf` into the working tree. The SQL analogy in
`app.go` was “tests inject a store”; there is no injected writer. A
later change can keep putting PDF concerns into `tui` because the
boundary the spec named is already gone.

## O3 — risk — medium

### Claim

Invalid `-chart-pdf` values fail only after first-run filesystem
mutation. `1990-13` and `banana` still `Ensure` a `config.toml` and
`Open` (creating) the SQLite file, then exit non-zero. FR5 / S8’s
“no file” is implemented as “no PDF,” not “no writes.”

### Evidence

`cmd/hdtools/main.go` `run()`: `config.Ensure`, `os.MkdirAll` of the
database directory, `dailylog.Open`, then if `-chart-pdf` is set,
`exportChartPDF`. `parseMonth` lives inside `exportChartPDF`, after
the store is open.

`TestParseMonthRejects` only `Stat`s the `-o` PDF path. `cliPaths`
has already written `config.toml`; `Open` creates `t.db` in the same
temp dir even on the failing month.

### Why this matters

A CLI-only typo is enough to create `~/.config/hdtools/config.toml`
and `~/.local/share/hdtools/hdtools.db` (or whatever `-db` named).
The process then prints `chart-pdf: month "banana": want YYYY-MM` and
never mentions those files. S8 / FR5 promised non-zero, no TUI, no
file. The TUI path creates those files on a deliberate launch;
`-chart-pdf banana` is not a launch. `parseMonth` can run immediately
after flag parse; it does not.

## O4 — risk — medium

### Claim

CLI write failure is specified (Decision 9: non-zero, no TUI) and
planned as a cmd test, but only TUI S3a is tested.
`TestChartPDFFlagSkipsTUI` covers the happy path. A refactor that
ignores `chartpdf.Write` errors, or that falls through to
`tea.NewProgram` after a failed export, is undetected at the process
boundary.

### Evidence

`cmd/hdtools/main_test.go` chart-pdf cases are `TestParseMonthRejects`
(`1990-13`, `banana`) and `TestChartPDFFlagSkipsTUI` (valid month,
`%PDF`, `0600`, returns within 5s). There is no test that forces
`Write` to fail (unwritable `-o`, EISDIR, and so on).

The companion plan’s cmd row: “Flag path: valid month writes a file;
invalid month is non-zero and creates no file; CLI write error is
non-zero.” `run` does return `exportChartPDF`’s error today, so the
TUI is not started — that structure is unenforced by a failing case.

### Why this matters

S3a locked TUI behaviour: error status, still on the chart, no quit,
no saved-path claim. The CLI analogue is Decision 9’s only
process-level guarantee that a full disk or a bad `-o` does not drop
the user into Bubble Tea or exit 0 with a missing/partial file.
`runCLI`’s 5s timeout would catch an accidental TUI start, but only
if the test actually takes the write-error path. It never does.

## O5 — risk — low

### Claim

The only S11 observation of “no day-number label after 10” is “no PDF
literal that `Atoi`s to 11–31.” That makes a Daily deficit (or any
other bare integer) in 11–31 a CI failure, which the spec never
required. The fixture’s 772 cal hides it.

### Evidence

`internal/chartpdf/chartpdf_test.go`
`TestWritePDFCurrentMonthClipsAtToday` walks extractable literals and
fails if `strconv.Atoi` yields 11–31. Spec S11 only forbids a
day-number label after 10.

Y labels are `%.1f` (so `Atoi` fails). Analysis in this fixture is
`Daily deficit: 772 cal` (772 > 31). A 10-day span with a ~0.02 kg
loss yields a deficit in 11–31; that PDF is still a correct clip and
would fail this test.

### Why this matters

S11 is the current-month clip gate (no column after today; deficit
divisor 10). If the inspector is wrong, either a legal PDF is
rejected or a day-11 label that is not a bare decimal integer is
accepted. The integer heuristic is now the requirement, decided in
the test, and no document says so.

- **accept-as-stated** — Y labels stay `%.1f` and the S11 fixture
  stays 80→79 (772 cal); write that down so the next reader does not
  treat “no integer 11–31” as a product rule.
- **revise-spec** — observe day numbers as the axis labels under the
  plot, not as any extractable integer in 11–31.
- **add-test** — a nearly-flat current month whose Daily deficit is
  in 11–31 must still pass the clip check.
- **consciously-carry** — known false-red for small deficits, shipped
  because the frozen 80→79 fixture never hits it.

## Explicitly not objecting to

- **Cwd overwrite without confirmation**: spec-time O8 accepted
  keep-cwd-and-overwrite; write-failure behaviour is specified and
  the TUI path is tested (S3a).
- **VGA table versus the current TTY palette for named `[colors]`**:
  spec-time O2 accepted a fixed print table; `color.go` matches
  Decision 6.
- **US Letter landscape, not A4**: Decision 1; tests read 792×612
  from the file.
- **PDF labelling every day versus the TUI labelling only 1 and
  last**: Decision 7; geometry is allowed to differ.
- **`chartspan` as the shared owner of clip, empty, Y pad, and
  analysis omission**: that half of spec-time O7 is implemented; TUI
  `chartView` and `chartpdf.paint` both call it.
- **`Write` `MkdirAll(..., 0755)` for nested `-o` paths**: extra
  relative to the spec, but the user named the path; a typo creating
  a directory is a decision, not an untested failure class for the
  monthly chart itself.
- **English `time.Month.String()` and Helvetica core fonts**: the
  spec froze `November 1990` and ASCII analysis/empty copy; there is
  no non-ASCII series text on the page.
- **Duplicate `lastTrendOnOrBefore` in `internal/tui/month.go`**: the
  month sheet was not this slice; long-term PDFs are out of scope.
- **`NO_COLOR` ignored on the PDF**: FR6; the writer never consults
  it.
