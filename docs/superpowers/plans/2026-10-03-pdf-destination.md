# Plan — PDF destination

**Spec**:
[`docs/superpowers/specs/2026-10-03-pdf-destination.md`](../specs/2026-10-03-pdf-destination.md)
**Follows**:
[`2026-09-12-pdf-monthly-charts.md`](../specs/2026-09-12-pdf-monthly-charts.md)
**Objections**:
[`pdf-destination.md`](../objections/pdf-destination.md)
(O1–O6 accepted)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

This slice changes where the monthly chart PDF is written. It does
not change the picture, the page box, colours, the clip, the
filename, or file mode `0600`. Do not add filled-sheet, long-term, or
blank PDF writers. Do not edit `idea.md` (the change-record link is
added outside this slice). Do not rewrite the 2026-09-12 spec.

`chartpdf.Write` already creates a missing parent at `0755` and
installs the file at `0600` via a temp file plus `Rename`. Keep that.
Pass it a path that already includes the directory. Do not add a
second mkdir, and do not change paint.

## Module structure

| File | Why |
| --- | --- |
| `internal/config/config.go` | `PDFDir` on `Config` (`toml:"pdf_dir,omitempty"`). `ResolvePDFDir` is the one directory choice for CLI and TUI. A non-string `pdf_dir` does not fail `Load`; the resolver reports it when the export uses the key. |
| `internal/config/config_test.go` | S3 trim and absolute, S4 omitted/blank/unknown/`[colors]`/first-run, S11 non-string still loads, S14 `~/` expansion and other relative failures. |
| `cmd/hdtools/main.go` | Empty `-o` joins `YYYY-MM-chart.pdf` onto `ResolvePDFDir`. Non-empty `-o` is the file path and does not call the resolver. Success prints that path as one line on stdout. Usage text for `-o` must not say the working directory is the default. |
| `cmd/hdtools/main_test.go` | S1, S2, S3 file location, S4 blank `pdf_dir`, S5, S6, S9 CLI, S10, S11 CLI, S13 CLI, S15 CLI. |
| `internal/tui/chart.go` | `writeChartPDF` stops using `os.Getwd` for the default path. It joins the filename onto `ResolvePDFDir`. |
| `internal/tui/chart_test.go` | Retarget the cwd `Stat("1990-11-chart.pdf")` tests. S7, S8, S9 TUI, S12, S13 TUI, S15 TUI. |
| `docs/reference/cli.md` | Default path is the data directory or `pdf_dir`. `-o` is still a file and wins. Success prints the path on stdout. |
| `docs/reference/config.md` | Top-level `pdf_dir`: absolute, or a leading `~/` for the home directory. Other relative values fail the export. Omitted means the data directory. A non-string does not fail the read. Absent from the first-run file. |
| `docs/reference/keys.md` | `p` writes into that directory, not the current directory. The status line shows the path. |
| `docs/how-to/monthly-chart-pdf.md` | CLI and TUI sections match the directory rule, including `pdf_dir` and `-o`. |
| `CHANGELOG.md` | Unreleased **Changed**: default monthly-chart PDF directory, `pdf_dir`, `-o` still the file path. |

Do not import `fpdf` from `internal/tui`. Do not change
`internal/chartpdf` paint, page size, or the `0600` install.
`internal/chartpdf` tests pass an explicit path; leave them.
Do not change the chart help string (`p pdf` does not name a
directory). Do not change list, form, month, or long-term key
handling. `p` on the month sheet can still start or continue a text
edit. That is not a PDF write.

## Algorithm notes

-   One resolver, `config.ResolvePDFDir(cfg) (string, error)`.
    Trim surrounding whitespace on `pdf_dir`. Empty after trim calls
    `DataDir` and returns that error unchanged (`home directory`
    when home and `XDG_DATA_HOME` are both unusable). An absolute
    path is returned as trimmed. A value that begins with `~/`
    replaces that prefix with the home directory from
    `os.UserHomeDir` and returns the join; if home cannot be
    resolved, return that error and do not fall back. Any other
    relative value, including `.` and `charts`, returns an error
    that names `pdf_dir`. Do not expand `$`. No `filepath.Abs`.
    `filepath.Join` is only for the `~/` prefix plus the remainder,
    and later for the filename onto the directory.
-   Callers format the existing name `%04d-%02d-chart.pdf` and
    `filepath.Join` it onto that directory. `Join` is what makes a
    trailing slash the same directory. `.` joined with the filename
    is the working directory.
-   `-o` is not trimmed and is not passed through the resolver. A
    non-empty `-o` is the file path `chartpdf.Write` receives, even
    when `pdf_dir` is not a string or is a relative value the
    resolver would reject. Do not join it to `pdf_dir` or the data
    directory. If that write fails, stop. No second attempt. On
    success, print the path as one line on stdout. On failure,
    print no path. The same stdout line applies when `-o` is empty
    and the write succeeds.
-   Parent creation stays inside `chartpdf.Write` (`MkdirAll` at
    `0755`, then the existing temp plus `Rename`). Tests assert the
    directory exists and the file is mode `0600`. Do not assert the
    directory's permission bits; umask masks `0755`, and
    `syscall.Umask` is process-global.
-   `Load` decodes `pdf_dir` as `any`, not as a string field, so a
    number is not a generic document failure. `nil` or a string that
    trims to empty leaves `PDFDir` empty. Any other type leaves
    `PDFDir` empty and sets a flag the resolver reports as
    `pdf_dir must be a string` without failing `Load`. Do not read
    `pdf_dir` out of the `[colors]` table. A string is not stat'd
    at load. `-o` does not call the resolver, so a non-string key
    does not block that write.
-   `PDFDir` is `omitempty`. `Ensure` writes `Default()`, so the
    first-run file has no `pdf_dir` key. Unknown top-level keys stay
    ignored by the existing struct decode.
-   TUI `p` still writes inside the key handler, before `Update`
    returns. Resolver or write error: set `err`, clear `status`,
    stay on the chart, do not quit, do not write a working-directory
    copy. Success: `status` is the path passed to `Write` (directory
    plus filename). Show a `~/` value after that prefix expands.
    Do not leave the tilde in the status line.
-   `-db` and `$HDTOOLS_DB` stay on the database path only. They are
    not arguments to `ResolvePDFDir`.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestChartPDFDefaultWritesDataDir`, `TestChartPDFDefaultUsesStandInHome` |
| FR2 | `TestPDFDirAbsolute`, `TestChartPDFDirAbsolute`, `TestChartPDFDirWithoutHome` |
| FR3 | `TestPDFDirOmittedOrBlankUsesDataDir`, `TestEnsureOmitsPDFDir`, `TestChartPDFBlankPDFDirWritesDataDir` |
| FR4 | `TestPDFDirHomePrefix`, `TestChartPDFHomePrefix`, `TestChartPDFRelativeDirFails` |
| FR5 | `TestChartPDFOAbsoluteWins`, `TestChartPDFRelativeOStaysInCwd`, existing `TestChartPDFFlagWriteFailure` |
| FR6 | `TestChartPWritesPDF`, `TestChartPUsesPDFDir` |
| FR7 | `TestChartPDFDefaultFailureNoCwdCopy`, `TestChartPWriteFailure`, `TestChartPPDFDirIsFile`, `TestChartPDFNoHomeFails`, `TestChartPNoHomeFails`, `TestChartPDFRelativeDirFails`, `TestChartPDFHomePrefixNoHome` |
| FR8 | `TestChartPDFDatabasePathDoesNotMovePDF` |
| FR9 | `TestPDFDirNonStringStillLoads`, `TestChartPDFNonStringPDFDir`, `TestChartPNonStringPDFDir` |
| FR10 | `TestChartPIgnoredOnList`, `TestChartPWhileEditingIsText`, `TestChartPIgnoredOnForm` |
| FR11 | `TestChartPDFOverwriteDefaultPath`, `TestChartPOverwrite` |
| FR12 | doc files in the module table (no Go test) |
| FR13 | existing `internal/chartpdf` tests stay; this slice adds no paint assertion |

`TestChartPDoesNotWriteLogs` stays a log-count test. Point its PDF
stat at the data-directory path so it still proves `p` wrote the
chart and did not insert a row.

## Test list

November 1990, filename `1990-11-chart.pdf`. Set
`XDG_DATA_HOME` or `HOME` with `t.Setenv` to a `t.TempDir`. Never
read the real home. `t.Chdir` for the working directory. Do not call
`t.Parallel` in a test that uses `t.Setenv` or `t.Chdir`. Follow
`TestChartPWritesPDF` (`t.Chdir`, no `t.Parallel`) and
`TestDefaultDBPathFallsBackToLocalShare` (`HOME` plus empty
`XDG_DATA_HOME`, no `t.Parallel`). A non-parallel test does not
overlap the package's parallel tests. Config `Load` cases that do
not touch the environment may use `t.Parallel`.

Write `pdf_dir` as toml text (`pdf_dir = "..."`), not only as a
struct field, so the on-disk key is what `Load` accepts.

Existing `-o` tests stay on an explicit `-o` and must remain green:
`TestChartPDFFlagSkipsTUI`, `TestChartPDFFlagWriteFailure`,
`TestParseMonthRejects`. Do not point them at the data directory.

Retarget, do not leave green against the old directory. A
`Stat("1990-11-chart.pdf")` in the working directory stays green if
`p` on the list, the form, or an edit writes only under the data
directory. Those tests must also stat
`$XDG_DATA_HOME/hdtools/1990-11-chart.pdf`. Success tests must stat
that data path and assert the working-directory name is absent. Do
not satisfy the old assertion by writing the working directory.

-   `TestPDFDirAbsolute` — trimmed absolute path, trailing slash, and
    surrounding spaces. `ResolvePDFDir` returns that trimmed string,
    not `DataDir`.
-   `TestPDFDirOmittedOrBlankUsesDataDir` — missing key, `""`,
    whitespace, an unknown top-level key, and `pdf_dir` under
    `[colors]`. With `XDG_DATA_HOME` set, the directory is
    `$XDG_DATA_HOME/hdtools`.
-   `TestEnsureOmitsPDFDir` — first-run file text does not contain
    `pdf_dir`.
-   `TestPDFDirNonStringStillLoads` — `pdf_dir = 3`. `Load` returns
    nil. `ResolvePDFDir` errors and the text names `pdf_dir`.
-   `TestPDFDirHomePrefix` — `~/charts` with `HOME` set returns
    `$HOME/charts`. `charts`, `charts/`, `.`, and
    `$XDG_DATA_HOME/charts` error and are not joined to the working
    directory or `DataDir`. `$` is not expanded.
-   `TestChartPDFDefaultWritesDataDir` — S1. `-db` and `-config`
    under the working directory. File only at the data path, mode
    `0600`, header `%PDF`, exit nil, directory created. Stdout is
    one line, that path. No TUI (the existing `runCLI` timeout).
-   `TestChartPDFBlankPDFDirWritesDataDir` — S4 write. Each of
    omitted `pdf_dir`, `pdf_dir = ""`, and whitespace-only
    `pdf_dir` writes the data-directory file and not the working
    directory. Unknown-key and `[colors]` cases stay in
    `TestPDFDirOmittedOrBlankUsesDataDir` (same directory string).
-   `TestChartPDFDefaultUsesStandInHome` — S2. `HOME` is a temp
    dir, `XDG_DATA_HOME` empty. File at
    `$HOME/.local/share/hdtools/1990-11-chart.pdf`. Not in the
    working directory.
-   `TestChartPDFDirAbsolute` — S3. Absolute `pdf_dir`, including a
    trailing-slash / padded value as a subtest. File there, not
    under `XDG_DATA_HOME/hdtools`. Missing directory is created.
-   `TestChartPDFDirWithoutHome` — S3 second case. Absolute
    `pdf_dir`, empty `XDG_DATA_HOME`, empty `HOME`, explicit `-db`
    and `-config`. Exit nil, file in `pdf_dir`. Stdout is that path.
-   `TestChartPDFOAbsoluteWins` — S5. `-o` absolute, `pdf_dir`
    missing, `XDG_DATA_HOME` set. File only at `-o`. `pdf_dir` still
    absent. Stdout is that path. Subtest: same with empty `HOME`,
    empty `XDG_DATA_HOME`, and explicit `-db` and `-config`.
-   `TestChartPDFRelativeOStaysInCwd` — S6. `-o out.pdf` with
    `pdf_dir` set. File is `out.pdf` in the working directory only.
    Stdout is one line, `out.pdf`.
-   `TestChartPDFDefaultFailureNoCwdCopy` — S9 CLI. Subtest: the
    data-directory file path is a directory, still a directory
    after, non-zero, no working-directory PDF. Subtest: `pdf_dir`
    is an existing file, non-zero, no PDF in the working directory
    or the data directory.
-   `TestChartPDFDatabasePathDoesNotMovePDF` — S10. Subtest `-db`
    outside the data directory. Subtest `$HDTOOLS_DB` and no `-db`.
    PDF at the data path only.
-   `TestChartPDFNonStringPDFDir` — S11. `pdf_dir = 3`. `Load`
    succeeds. No `-o`: non-zero, error names `pdf_dir`, no PDF in
    the working directory or the data directory, stdout empty.
    With `-o`: exit nil, file only at `-o`, stdout is that path.
-   `TestChartPDFOverwriteDefaultPath` — S13 CLI. Seed the default
    path with `old`, export twice. Exactly one
    `1990-11-chart.pdf`, header `%PDF`, mode `0600`, nothing in
    the working directory.
-   `TestChartPDFNoHomeFails` — S15 CLI. `-db` outside, no
    `pdf_dir`, no `-o`, empty `HOME`, empty `XDG_DATA_HOME`.
    Non-zero, error contains `home directory`, no working-directory
    PDF.
-   `TestChartPDFHomePrefix` — S14. `pdf_dir = "~/charts"` and
    `HOME` set. File is `$HOME/charts/1990-11-chart.pdf`, mode
    `0600`, not in the working directory or the data directory.
    Stdout is that path.
-   `TestChartPDFRelativeDirFails` — S14. `charts`, `.`, `charts/`,
    and `$XDG_DATA_HOME/charts` each exit non-zero, write no PDF in
    the working directory, under `HOME`, or under the data
    directory, and print no path. The TUI is not started.
-   `TestChartPDFHomePrefixNoHome` — S14. `pdf_dir = "~/charts"`,
    empty `HOME`, empty `XDG_DATA_HOME`, explicit `-db` and
    `-config`, no `-o`. Non-zero, no PDF, stdout empty.

TUI. `twoDayApp` stays on `config.Default()`; layout tests use it.
Set `XDG_DATA_HOME` in the PDF test before `p`. For `pdf_dir`,
construct `Config` in that test. Do not add a PDF-library import.

-   `TestChartPWritesPDF` — S7. Stat the data-directory path, mode
    `0600`. Status contains that directory and `1990-11-chart.pdf`.
    Still the November chart. Working directory has no
    `1990-11-chart.pdf`.
-   `TestChartPUsesPDFDir` — S8. Absolute `pdf_dir`. File and status
    use it. Not the data directory, not the working directory.
-   `TestChartPWriteFailure` — S9. The data-directory file path is
    a directory (not a cwd directory of that name). Error is shown,
    path is not a saved-file claim, still the monthly chart, cwd
    has no PDF, the directory is still a directory.
-   `TestChartPPDFDirIsFile` — S9. `pdf_dir` is an existing file.
    Error, still the chart, no PDF in cwd or the data directory.
    Opening the chart did not require that directory to be writable.
-   `TestChartPIgnoredOnList` — S12 list, month sheet, long-term.
    `pdf_dir` is a third temp directory. No file in cwd, the data
    directory, or `pdf_dir`. Do not require the month sheet to
    ignore `p` as text.
-   `TestChartPWhileEditingIsText` — S12. `p` still lands in the
    field. Also stat the data-directory path and assert it is
    absent.
-   `TestChartPIgnoredOnForm` — S12 form. No PDF in cwd, the data
    directory, or `pdf_dir`. A text field still receives `p`.
-   `TestChartPNonStringPDFDir` — S11 TUI. `pdf_dir = 3` still opens
    the chart. `p` shows an error that names `pdf_dir`, stays on
    the chart, and writes no PDF in cwd or the data directory.
-   `TestChartPOverwrite` — S13 TUI. `p` twice. Exactly one
    `1990-11-chart.pdf` in the data directory, mode `0600`, none
    in the working directory.
-   `TestChartPNoHomeFails` — S15 TUI. Empty `HOME`, empty
    `XDG_DATA_HOME`, chart opened on an explicit store. `p` shows
    an error containing `home directory`, stays on the chart, no
    working-directory PDF.
-   `TestChartPDoesNotWriteLogs` — stat the data-directory PDF, row
    count still 2.

Leave paint, Y-range, loss, empty-month, and `NO_COLOR` chart tests
on their current assertions. They do not press `p`.
