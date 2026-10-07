---
spec: docs/superpowers/specs/2026-10-03-pdf-destination.md
date: 2026-10-04
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: risk
    severity: medium
    claim: "ResolvePDFDir can return a relative directory when XDG_DATA_HOME or HOME is relative, and the export then creates the chart in the working directory and reports success."
    evidence: "internal/config/config.go ResolvePDFDir returns DataDir() unchanged when pdf_dir is empty, and returns filepath.Join(home, rest) for a ~/ prefix, with no absolute check. DataDir uses os.Getenv(\"XDG_DATA_HOME\") whenever it is non-empty. os.UserHomeDir returns any non-empty HOME. os.UserConfigDir errors when XDG_CONFIG_HOME is relative."
    disposition: accepted
    disposition_rationale: "A resolved PDF directory that is not absolute fails the export and writes nothing. The path is not made absolute against the working directory. DataDir is unchanged, so the default database path does not move. Tests cover a relative XDG_DATA_HOME and pdf_dir ~/charts with a relative HOME."
  - id: O2
    category: risk
    severity: medium
    claim: "The config reference says an omitted pdf_dir is the same directory as the database, but the export never reads the database path, so -db or $HDTOOLS_DB leaves the chart somewhere else."
    evidence: "docs/reference/config.md: \"the same directory as the database.\" cmd/hdtools/main.go exportChartPDF calls config.ResolvePDFDir(cfg) and does not take the database path. TestChartPDFDatabasePathDoesNotMovePDF expects the PDF in the data directory and not beside -db or $HDTOOLS_DB."
    disposition: accepted
    disposition_rationale: "The config reference and the how-to say the omitted directory is the XDG data directory, which is where the database lives when -db and $HDTOOLS_DB are unset. Those two do not move the PDF. The export code is unchanged."
---

## O1 — risk — medium

### Claim

`ResolvePDFDir` can return a relative directory. Both writers then
create the chart under the process working directory and report
success. That happens when `XDG_DATA_HOME` is relative, and when
`pdf_dir` begins with `~/` while `HOME` is relative. Neither case is
an error.

### Evidence

`ResolvePDFDir` returns `DataDir` unchanged when `pdf_dir` is empty,
and returns `filepath.Join` of the home directory and the remainder
for a `~/` prefix. Neither result is checked to be absolute.

```104:122:internal/config/config.go
func ResolvePDFDir(cfg Config) (string, error) {
	if cfg.pdfDirBad {
		return "", fmt.Errorf("pdf_dir must be a string")
	}
	dir := strings.TrimSpace(cfg.PDFDir)
	if dir == "" {
		return DataDir()
	}
	if filepath.IsAbs(dir) {
		return dir, nil
	}
	if rest, ok := strings.CutPrefix(dir, "~/"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory: %w", err)
		}
		return filepath.Join(home, rest), nil
	}
	return "", fmt.Errorf("pdf_dir %q must be absolute or start with ~/", dir)
}
```

`DataDir` treats any non-empty `XDG_DATA_HOME` as the base, relative
or not:

```68:77:internal/config/config.go
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
```

`os.UserHomeDir` returns whatever non-empty `HOME` it finds. It does
not require that value to be absolute. The config helper beside this
code does the opposite for the sibling variable. `os.UserConfigDir`
returns an error when `$XDG_CONFIG_HOME` is relative and does not use
that path.

The writers join the filename onto the returned directory and treat a
completed write as success. The CLI prints that path. The TUI stores
it as the status line.

```130:148:cmd/hdtools/main.go
	if out == "" {
		dir, err := config.ResolvePDFDir(cfg)
		if err != nil {
			return fmt.Errorf("chart-pdf: %w", err)
		}
		out = filepath.Join(dir, fmt.Sprintf("%04d-%02d-chart.pdf", year, month))
	}
	// ...
	fmt.Println(out)
```

```73:96:internal/tui/chart.go
func (a *App) writeChartPDF() {
	dir, err := config.ResolvePDFDir(a.cfg)
	// ...
	path := filepath.Join(dir, name)
	err = chartpdf.Write(path, chartpdf.Options{
```

`chartpdf.Write` creates any parent other than `.` with `MkdirAll`. A
relative parent is created in the working directory.

The tests that lock the default and the `~/` expansion set `HOME` and
`XDG_DATA_HOME` only to `t.TempDir()` or to empty. An empty `HOME` is
the unavailable-home failure, which is a different case.

### Why this matters

A relative `pdf_dir` such as `charts` or `.` now fails the export, and
nothing is written. The bases those rules sit on are not held to the
same bar. `XDG_DATA_HOME=data` writes
`data/hdtools/YYYY-MM-chart.pdf` in whatever directory the process was
started from, exits 0, and prints that relative path as the one-line
success report. The next launch from another directory writes another
file. `pdf_dir = "~/charts"` with `HOME` set to a relative value does
the same thing through the new expansion.

The XDG base-directory rule is that a relative path in these variables
is invalid and should be ignored. This process already does that for
`$XDG_CONFIG_HOME`, by failing the config path. The PDF path, added so
the chart would stop following the launch directory, accepts a
relative data base and reports success.

`filepath.Abs` is the wrong repair. Turning the relative value into an
absolute path against the working directory is the behaviour that was
rejected. A relative result should fail the export the way any other
relative `pdf_dir` fails.

`DataDir` is also where a default database is created. Rejecting a
relative `XDG_DATA_HOME` inside `DataDir` would move that database,
which this slice does not do. The export can refuse a non-absolute
directory on its own and leave the database path alone. The `~/`
branch is new in this slice and does not have that constraint.

- **accept-as-stated** — a relative `XDG_DATA_HOME` or `HOME` is used
  as given, including when the chart then follows the working
  directory. Write that down, including that it matches the database
  for the data-directory case, so the next reader inherits a decision.
- **revise-spec** — a resolved PDF directory that is not absolute
  fails the export and writes nothing. The `~/` case is the part this
  slice introduced. Changing `DataDir` itself would also move the
  default database and is a wider change than this slice.
- **add-test** — a relative `XDG_DATA_HOME` and a relative `HOME` with
  `pdf_dir = "~/charts"` are requirements. A test that expects either
  the working-directory file or a failed export is what makes the
  choice one.
- **consciously-carry** — known, wrong for a non-absolute base, and
  shipped anyway, because the PDF directory stays the directory
  `DataDir` already returns for the database.

## O2 — risk — medium

### Claim

The config reference says an omitted `pdf_dir` is the same directory
as the database. The export never reads the database path. `-db` and
`$HDTOOLS_DB` leave the chart in the XDG data directory, not beside
the database file.

### Evidence

`docs/reference/config.md` says an omitted, empty, or whitespace
`pdf_dir` means the XDG data directory, then equates that with the
database:

> Omitted, empty, or whitespace means `$XDG_DATA_HOME/hdtools`
> (otherwise `~/.local/share/hdtools`), the same directory as the
> database.

`exportChartPDF` chooses the directory only from
`config.ResolvePDFDir(cfg)` when `-o` is empty. The store and the
database path are not inputs to that choice. `resolveDBPath` stays a
separate function and is not passed in (`cmd/hdtools/main.go`).

`TestChartPDFDatabasePathDoesNotMovePDF` locks the split. With `-db`
or `$HDTOOLS_DB` pointing outside the data directory, the PDF is at
the data-directory path and is absent beside that database file
(`cmd/hdtools/main_test.go`).

### Why this matters

The path in that sentence is the default database directory, and that
reading matches the code. The appositive does not. `-db` and
`$HDTOOLS_DB` exist so the log can live somewhere else. A person who
moved it, and who reads "the same directory as the database," looks
next to that file. The chart is still under the default data
directory. The file is mode `0600`, and it is not where the reference
says it is.

The CLI prints the real path, and the chart screen shows the database
path next to a status line, so the moment of export can be checked.
The reference is what is left after that line scrolls off. The config
page never says that the database path does not move the PDF. The
how-to says `-db` and `-config` "work as usual," which does not
correct the sentence.

## Explicitly not objecting to

- **No TUI test that opens with `pdf_dir` of `charts` or `.`:**
  `writeChartPDF` calls `ResolvePDFDir` before it creates anything,
  and `TestChartPDFRelativeDirFails` already locks those values as a
  failed export that writes nothing. Opening the chart does not
  resolve the directory, so a missing TUI case is not a second writer.
- **`filepath.Join` cleaning `..`:** callers join the filename onto
  the directory, and Join cleans. The status line and stdout show
  that cleaned path. A lexical `..` and a symlink disagreeing is not
  a case the scenarios ask for.
- **`~` alone and `~user`:** only a leading `~/` expands. Any other
  tilde fails the export and names `pdf_dir`. That is the accepted
  rule, not a silent directory named `~`.
- **The code ignoring `-db` and `$HDTOOLS_DB`:** Decision 2 and
  `TestChartPDFDatabasePathDoesNotMovePDF` are the behaviour. O2 is
  the reference sentence that describes it as the database's own
  directory, not the split itself.
- **The command table omitting `~/.local/share/hdtools`:** the how-to
  states that fallback, and a successful export prints the absolute
  path. The table is incomplete, not a second location.
- **Directory mode `0755`, silent overwrite, and the `0600` install:**
  this slice keeps `chartpdf.Write`'s temp file plus `Rename`, and
  the tests lock the file mode. Confirm-before-overwrite stayed out
  of scope.
- **CLI `time.Now()` versus TUI `localToday()`:** `LastPlottedDay`
  uses `Date()` on either value, and both are the local calendar day.
  That is not a destination failure.
- **Spec-time O1–O6:** those dispositions were accepted and the spec
  was amended. This pass does not reopen them.
