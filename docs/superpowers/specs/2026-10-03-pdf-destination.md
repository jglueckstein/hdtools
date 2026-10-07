# Where the monthly chart PDF is written

**Date**: 2026-10-03
**Status**: approved
**Objections**:
[`pdf-destination.md`](../objections/pdf-destination.md)
(O1–O6 accepted)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Follows**:
[`2026-09-12-pdf-monthly-charts.md`](2026-09-12-pdf-monthly-charts.md)
**Amends**: that spec's Decision 2 (the current-directory sentences
for the CLI and for TUI `p`), Decision 10, and S3 ("working
directory"). Decision 11 below also covers that spec's FR1 cwd
default and the "XDG as the PDF destination" half of its out-of-scope
line. Confirm-before-overwrite stays out of scope. The 2026-09-12
file remains the record of the shipped picture; this file is the
path change.

The monthly chart PDF already exists. CLI `-chart-pdf` and TUI `p`
write it in the process working directory. The backlog wants PDFs in
the same data directory as the database, with a config override, and
with `-o` still naming the file. This slice changes only that path.
The page does not change.

Filled log sheets, long-term chart PDFs, and blank sheets are not
produced here. When they are added later, they use this same
directory rule.

## User stories

### US1 — Keep the chart with the log

As a person who exports a monthly chart, I want the PDF in the same
data directory as the database, whether I use the command or the
chart key, so that it is not left in whatever directory the program
was started from.

### US2 — Put charts in another folder

As a person who files charts somewhere else, I want to set that
directory in the config file so I do not pass a path on every export.

### US3 — Send one PDF to a chosen file

As a person who needs a single chart at a particular path, I want
`-o` to name that file and ignore the usual directory so a one-off
export does not change where the next one goes.

## Decisions

1.  **Default directory** is the directory the database already uses
    when `-db` and `$HDTOOLS_DB` are not set:
    `$XDG_DATA_HOME/hdtools` when `XDG_DATA_HOME` is non-empty,
    otherwise `~/.local/share/hdtools`. The file is
    `YYYY-MM-chart.pdf` in that directory, beside `hdtools.db`, not
    in a subdirectory. There is no environment variable for the PDF
    directory.
2.  **`-db` and `$HDTOOLS_DB` do not move the PDF.** A database opened
    somewhere else still uses the data directory from Decision 1
    unless `pdf_dir` or `-o` says otherwise.
3.  **Config key** is top-level `pdf_dir`, a directory string, next
    to `display_unit`. Not under `[colors]`. A `pdf_dir` entry inside
    `[colors]` is an unknown colour and does not set the directory.
    If the key is omitted, empty, or only whitespace, the directory
    is Decision 1. Surrounding whitespace on a real path is removed.
    The first-run config file does not contain `pdf_dir`.
    Unknown keys stay ignored.
4.  **A non-string `pdf_dir` does not fail the config read** and does
    not stop the TUI from opening (O3). The bad value is not a
    directory. `-o` still writes, because that flag ignores
    `pdf_dir`. An export that would use the key fails, names
    `pdf_dir`, and writes nothing. A string is not checked for
    existence at load. A missing or unwritable directory does not
    stop the TUI from opening. Only the export fails. This key is
    not fail-closed like `display_unit`: a bad directory cannot
    store pounds as kilograms.
5.  **`pdf_dir` does not follow the launch directory** (O2). An
    absolute path is used as written. A trailing slash is the same
    directory. A value that begins with `~/` expands that prefix to
    the home directory and leaves the rest of the path as written.
    `$` is not expanded. Any other relative value, including `.`
    and `charts`, fails the export: the TUI still opens, nothing is
    written, and there is no working-directory fallback. If the home
    directory is unavailable, a `~/` value fails the export the same
    way and writes nothing. A resolved directory that is not absolute
    fails the export and writes nothing (code-mode O1). That includes
    a relative `XDG_DATA_HOME` when `pdf_dir` is empty, and `~/…` when
    `HOME` is relative. The value is not made absolute against the
    working directory. The database directory helper is unchanged.
    `-o` and `-db` stay working-directory relative when their own
    paths are relative. That rule is for those flags, not for
    `pdf_dir`.
6.  **`-o` is still a file path, not a directory.** When non-empty it
    wins completely. It is not joined onto `pdf_dir` or the data
    directory. `-o out.pdf` is the working directory. `-o /tmp/out.pdf`
    is that path. Config `pdf_dir` is ignored for that invocation,
    and a missing `pdf_dir` is not created. On success the CLI prints
    that path as one line on standard output (O1). A failed `-o`
    write does not fall back to `pdf_dir`, the data directory, or
    another name in the working directory, and it prints no path.
    Only a zero-length `-o` counts as empty; the flag is not trimmed.
7.  **When `-o` is empty**, the CLI writes `YYYY-MM-chart.pdf` inside
    the chosen directory (`pdf_dir` if set, else Decision 1), creates
    that directory if missing (mode `0755` before the process umask,
    the same as today's parent directories for this PDF), file mode
    `0600`, and overwrites if the file exists. On success the CLI
    prints that path as one line on standard output (O1). No
    working-directory fallback if that write fails, and failure
    prints no path. Parent directories of the file are created the
    same way they are today.
8.  **TUI `p`** uses the same directory rule as an empty `-o`. There
    is no per-press path. On success the status line contains the
    path written, including its directory and `YYYY-MM-chart.pdf`,
    not only the filename. An absolute directory is shown as that
    absolute path. A `~/` value is shown after that prefix expands
    to the home directory. On failure: an error is shown, the screen
    stays the monthly chart, the process does not quit, the file is
    not claimed as saved, and no working-directory copy is written.
    The error may name the path that failed. `p` writes no PDF on
    the list, the form, the month sheet, or the long-term chart:
    nothing in the working directory, nothing under the data
    directory, and nothing under `pdf_dir` (O5). This slice does not
    change what those screens already do with the `p` key (on the
    month sheet and the form it can still be typed as text).
9.  **Home and `XDG_DATA_HOME` both unusable** (no home directory, and
    `XDG_DATA_HOME` unset or empty): the export fails with the same
    home-directory failure the data directory already reports. It
    does not fall back to the working directory. An explicit `-db`
    (or `$HDTOOLS_DB`) still lets the log open; only the export
    needs the data directory when `pdf_dir` and `-o` are empty.
10. **`pdf_dir` that names an existing file** (not a directory): the
    export fails. No working-directory fallback. The same applies
    when the file path that would be written already exists as a
    directory.
11. **This slice amends** the 2026-09-12 cwd sentences in Decision 2,
    Decision 10, S3's "working directory", and FR1's "default is
    `YYYY-MM-chart.pdf` in the cwd". It also takes the "XDG as the
    PDF destination" item out of that spec's leftover list. It does
    not change the picture, the page box, colours, the clip, or
    `-o` as a full file path. Confirm-before-overwrite stays out of
    scope. The older spec file is not rewritten.
12. **Out of scope:** filled log-sheet PDFs, long-term PDFs, blank
    sheets, ntcharts, Bubble Tea v2, moving the database, changing
    the filename pattern, a confirm-before-overwrite prompt, and an
    environment variable for the PDF directory. Later PDF kinds must
    use this same directory rule; this slice does not implement
    them.

## Acceptance scenarios

November 1990 is the fixture. The file name is `1990-11-chart.pdf`.
The working directory and the data directory are different temporary
directories unless a scenario says the file belongs in the working
directory. Tests set `XDG_DATA_HOME`, or a stand-in home, to a
temporary directory. They must not use the developer's real home.
Unset `XDG_DATA_HOME` and an empty value are the same.

A successful file starts with `%PDF` and is mode `0600`. The page
inside it is unchanged from the 2026-09-12 spec; these scenarios do
not re-check the picture.

Every successful `-chart-pdf` prints the path it wrote as one line
on standard output (O1). A failed export prints no path.

### S1 — CLI default is the data directory

**Given** a log in November 1990
**And** no `-o` and no `pdf_dir`
**And** `XDG_DATA_HOME` points at a temporary directory
**And** the working directory is a different temporary directory
**When** `-chart-pdf 1990-11` runs
**Then** the process exits 0 and does not start the TUI
**And** `$XDG_DATA_HOME/hdtools/1990-11-chart.pdf` exists
**And** that directory was created if it was missing
**And** the file is mode `0600` and starts with `%PDF`
**And** the file is beside where `hdtools.db` would live, not in a
subdirectory
**And** the working directory has no `1990-11-chart.pdf`
**And** standard output is one line, that same path

### S2 — No `XDG_DATA_HOME` uses a stand-in home

**Given** a log in November 1990
**And** no `-o` and no `pdf_dir`
**And** `XDG_DATA_HOME` is unset or empty
**And** the home directory is a temporary stand-in, not the
developer's real home
**And** the working directory is a different temporary directory
**When** `-chart-pdf 1990-11` runs
**Then** the process exits 0 and does not start the TUI
**And** `~/.local/share/hdtools/1990-11-chart.pdf` under that
stand-in home exists, is mode `0600`, and starts with `%PDF`
**And** the working directory has no `1990-11-chart.pdf`

### S3 — Absolute `pdf_dir` replaces the data directory

**Given** a log in November 1990
**And** `pdf_dir` is an absolute directory that does not yet exist
**And** `XDG_DATA_HOME` points at a temporary directory
**And** no `-o`
**When** `-chart-pdf 1990-11` runs
**Then** the process exits 0
**And** `1990-11-chart.pdf` exists in that directory, mode `0600`
**And** it does not exist under `$XDG_DATA_HOME/hdtools`
**And** a `pdf_dir` with a trailing slash, or with surrounding
whitespace, is that same directory

**Given** the same log, an absolute `pdf_dir`, and no `-o`
**And** `-db` and `-config` point at temporary paths (O4)
**And** `XDG_DATA_HOME` is empty and the home directory is
unavailable
**When** `-chart-pdf 1990-11` runs
**Then** the process exits 0 and the PDF is in that `pdf_dir`

### S4 — Omitted or blank `pdf_dir` is the data directory

**Given** a log in November 1990 and no `-o`
**And** `XDG_DATA_HOME` points at a temporary directory
**When** the config omits `pdf_dir`, or sets it to `""`, or sets it
to whitespace only (each of those files on its own)
**Then** the PDF is written at
`$XDG_DATA_HOME/hdtools/1990-11-chart.pdf` and not in the working
directory

**Given** a config with an unknown top-level key and no `pdf_dir`
**When** the config is read and the chart is exported
**Then** the read succeeds and the PDF uses the data directory

**Given** `pdf_dir` only as a key inside `[colors]`
**When** the chart is exported with no `-o`
**Then** that key does not set the directory
**And** the PDF uses the data directory

**Given** a missing config file
**When** the program creates the first-run config
**Then** that file does not contain a `pdf_dir` key

### S5 — Absolute `-o` wins

**Given** a log in November 1990
**And** `pdf_dir` is an absolute directory that does not exist
**And** `XDG_DATA_HOME` is set
**When** `-chart-pdf 1990-11 -o` is given an absolute file path
**Then** the process exits 0
**And** the PDF exists only at that `-o` path, mode `0600`
**And** it is not under `pdf_dir` and not under the data directory
**And** the missing `pdf_dir` is still missing

**Given** the same log and no usable home or `XDG_DATA_HOME`
**And** `-db` and `-config` point at temporary paths (O4)
**And** `pdf_dir` is an absolute directory that does not exist
**When** `-chart-pdf 1990-11 -o` is an absolute file path
**Then** the process exits 0 and the PDF exists only at that path

### S6 — Relative `-o` stays in the working directory

**Given** a log in November 1990
**And** `pdf_dir` is an absolute directory
**And** `XDG_DATA_HOME` points at a temporary directory
**When** `-chart-pdf 1990-11 -o out.pdf` runs
**Then** `out.pdf` exists in the working directory, mode `0600`
**And** standard output is one line, `out.pdf`
**And** it does not exist inside `pdf_dir`
**And** `1990-11-chart.pdf` does not exist under the data directory
or inside `pdf_dir`

### S7 — TUI `p` uses the data directory

**Given** the monthly chart for November 1990
**And** no `pdf_dir`
**And** `XDG_DATA_HOME` points at a temporary directory
**And** the working directory is a different temporary directory
**When** `p` is pressed
**Then** `$XDG_DATA_HOME/hdtools/1990-11-chart.pdf` exists, mode
`0600`
**And** the status line contains that directory and
`1990-11-chart.pdf`
**And** the screen is still the monthly chart
**And** the working directory has no `1990-11-chart.pdf`

### S8 — TUI `p` uses `pdf_dir`

**Given** the monthly chart for November 1990
**And** `pdf_dir` is an absolute directory
**When** `p` is pressed
**Then** `1990-11-chart.pdf` exists in that directory, mode `0600`
**And** the status line contains that directory and
`1990-11-chart.pdf`
**And** the screen is still the monthly chart
**And** the file is not under the data directory
**And** the working directory has no `1990-11-chart.pdf`

### S9 — Export failure does not copy into the working directory

**Given** a log in November 1990 and no `-o`
**And** the path `$XDG_DATA_HOME/hdtools/1990-11-chart.pdf` already
exists as a directory
**When** `-chart-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** that path is still a directory
**And** the working directory has no `1990-11-chart.pdf`

**Given** `pdf_dir` is an existing file, not a directory, and no `-o`
**When** `-chart-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** no PDF is written in the working directory or under the data
directory

**Given** the monthly chart for November 1990 is already open
**And** the default file path already exists as a directory
**When** `p` is pressed
**Then** the screen shows an error and does not present the file as
saved
**And** the process does not quit
**And** the screen is still the monthly chart
**And** the working directory has no `1990-11-chart.pdf`
**And** the path that was a directory is still a directory

**Given** the monthly chart is open and `pdf_dir` names an existing
file
**When** `p` is pressed
**Then** the screen shows an error, stays on the monthly chart, and
does not quit
**And** no PDF is written in the working directory or under the data
directory

### S10 — The database path does not move the PDF

**Given** a log in November 1990
**And** no `-o` and no `pdf_dir`
**And** `XDG_DATA_HOME` points at a temporary directory
**When** `-db` points at a database outside that data directory
**Then** the PDF is `$XDG_DATA_HOME/hdtools/1990-11-chart.pdf`
**And** no PDF is written beside that database file
**And** the working directory has no `1990-11-chart.pdf`

**Given** the same logs, no `-db`, and no `pdf_dir`
**When** `$HDTOOLS_DB` points at a database outside the data
directory and `-chart-pdf 1990-11` runs
**Then** the PDF is in the data directory, not beside that database,
and not in the working directory

### S11 — Non-string `pdf_dir` does not block the log

**Given** a config file whose `pdf_dir` is a number (`pdf_dir = 3`)
**And** `-db` and `-config` point at temporary paths
**And** `XDG_DATA_HOME` points at another temporary directory
**When** the config is read
**Then** the read succeeds and the TUI can open (O3)

**When** `-chart-pdf 1990-11` runs with no `-o`
**Then** the process exits non-zero and does not start the TUI
**And** the error names `pdf_dir`
**And** no PDF is written in the working directory or under the data
directory
**And** standard output has no path

**When** `-chart-pdf 1990-11 -o` names a file
**Then** the process exits 0
**And** the PDF exists only at that `-o` path
**And** standard output is one line, that path

**Given** the monthly chart is already open with that config
**When** `p` is pressed
**Then** the screen shows an error that names `pdf_dir`, stays on
the monthly chart, and does not quit
**And** no PDF is written in the working directory or under the data
directory

### S12 — `p` still writes nothing from other screens

**Given** the daily list for a November 1990 log
**And** `pdf_dir` is a third temporary directory, different from
the working directory and the data directory (O5)
**And** `XDG_DATA_HOME` points at a temporary directory
**And** the working directory is a different temporary directory
**When** `p` is pressed on the list
**Then** no PDF is created in the working directory, under the data
directory, or in `pdf_dir`
**And** the screen is still the list

**When** `p` is pressed on the month sheet
**Then** no PDF is created in those three directories
**And** while a month cell is being edited, `p` is still text in
the field
**And** this slice does not change `p` starting an edit when the
month sheet was not editing

**When** `p` is pressed on the long-term chart
**Then** no PDF is created in those three directories
**And** the screen is still the long-term chart

**When** `p` is pressed on the form
**Then** no PDF is created in those three directories
**And** where the field already accepts text, `p` is still typed
into that field

### S13 — A second export overwrites the same file

**Given** the default data-directory path already contains the bytes
`old`
**And** no `-o` and no `pdf_dir`
**When** `-chart-pdf 1990-11` runs, and then runs again
**Then** that directory has exactly one entry named
`1990-11-chart.pdf`
**And** the file starts with `%PDF` and is mode `0600`
**And** there is no backup file and no `1990-11-chart.pdf` in the
working directory

**Given** the monthly chart for November 1990 and no `pdf_dir`
**When** `p` is pressed twice
**Then** the data directory has exactly one entry named
`1990-11-chart.pdf`, mode `0600`
**And** the working directory has no `1990-11-chart.pdf`

### S14 — `~/` expands; other relative values fail

**Given** a log in November 1990 and no `-o`
**And** the home directory is a temporary stand-in
**And** the working directory is a different temporary directory
**When** `pdf_dir` is `~/charts`
**Then** the PDF is `charts/1990-11-chart.pdf` under that stand-in
home, mode `0600` (O2)
**And** it is not under the data directory
**And** it is not in the working directory
**And** standard output is one line, that path

**When** `pdf_dir` is `charts`, `.`, `charts/`, or
`$XDG_DATA_HOME/charts`
**Then** the export fails and the TUI can still open
**And** no PDF is written in the working directory, under the home
directory, or under the data directory
**And** `$` is not expanded

**Given** no home directory and empty `XDG_DATA_HOME`
**And** `-db` and `-config` point at temporary paths
**And** `pdf_dir` is `~/charts` and there is no `-o`
**When** `-chart-pdf 1990-11` runs
**Then** the process exits non-zero and writes no PDF
**And** standard output has no path

### S15 — No data directory means no export

**Given** a log opened with `-db` pointing outside the data
directory
**And** no `-o` and no `pdf_dir`
**And** `XDG_DATA_HOME` is empty and the home directory is
unavailable
**When** `-chart-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** the error is the existing home-directory failure used when
the data directory cannot be resolved (the message names the home
directory)
**And** the working directory has no `1990-11-chart.pdf`

**Given** the monthly chart is open in that same environment, with
the database path given explicitly so opening the log did not need
the data directory
**When** `p` is pressed
**Then** the screen shows an error, stays on the monthly chart, and
does not quit
**And** the working directory has no `1990-11-chart.pdf`
**And** the file is not claimed as saved

## Functional requirements

-   **FR1.** With no `-o` and no `pdf_dir`, `-chart-pdf YYYY-MM`
    writes `YYYY-MM-chart.pdf` in the data directory from Decision 1,
    creates that directory if needed, mode `0600`, exits 0, does not
    start the TUI, and does not write that filename in the working
    directory. Standard output is one line, that path.
-   **FR2.** A `pdf_dir` string that is absolute, or that begins with
    `~/` after that prefix expands to the home directory, is that
    directory. Surrounding whitespace is trimmed first. An absolute
    path is used as written. A trailing slash is the same directory.
    The directory is created if missing. The file is not also written
    under the data directory. Export does not require a resolvable
    data directory when that `pdf_dir` is usable. The home-less cases
    in S3 and S5 open the log with explicit `-db` and `-config`.
-   **FR3.** Omitted, empty, or whitespace-only `pdf_dir` uses the
    data directory. Unknown keys stay ignored. `pdf_dir` inside
    `[colors]` does not set the directory. The first-run config file
    does not contain `pdf_dir`.
-   **FR4.** A value that begins with `~/` expands that prefix to the
    home directory. `$` is not expanded. Any other relative value,
    including `"."` and `"charts"`, fails the export, writes nothing,
    and does not stop the TUI from opening. `"~/charts"` with no home
    directory fails the same way. A resolved directory that is not
    absolute fails the export and writes nothing. The database
    directory helper is unchanged.
-   **FR5.** A non-empty `-o` is the whole file path. It is not joined
    to `pdf_dir` or the data directory. Relative `-o` is the working
    directory. Absolute `-o` is that path. `pdf_dir` is ignored, and
    a missing `pdf_dir` is not created. On success, standard output
    is one line, that path. Failure of that path does not fall back
    anywhere else and prints no path. `-o` is not trimmed. The
    home-less `-o` case opens the log with explicit `-db` and
    `-config`.
-   **FR6.** TUI `p` on the monthly chart uses the same directory as
    an empty `-o`. Success: status line contains the directory and
    `YYYY-MM-chart.pdf`, the screen stays the monthly chart, mode
    `0600`. A `~/` value is shown after that prefix expands.
-   **FR7.** Export failure does not write a working-directory copy
    and prints no path. Covered failures: the destination path exists
    as a directory; `pdf_dir` names an existing file; the data
    directory cannot be resolved because there is no home and
    `XDG_DATA_HOME` is empty; `pdf_dir` is relative and does not
    begin with `~/`; `pdf_dir` begins with `~/` and the home
    directory is unavailable; `pdf_dir` is not a string and the
    export would use it; the resolved directory is not absolute.
    The CLI exits non-zero and does not start
    the TUI. TUI `p` shows an error, stays on the monthly chart, does
    not quit, and does not claim the file was saved. A missing or
    unwritable directory, a relative `pdf_dir`, or a non-string
    `pdf_dir` does not stop the TUI from opening when the database
    path is usable.
-   **FR8.** `-db` and `$HDTOOLS_DB` do not change the PDF directory.
-   **FR9.** A `pdf_dir` value that is not a string does not fail the
    config read and does not stop the TUI from opening. `-o` still
    writes that file. An export that would use the key fails, names
    `pdf_dir`, and writes nothing. A string is not checked for
    existence at load.
-   **FR10.** `p` on the list, form, month sheet, or long-term chart
    writes no PDF in the working directory, the data directory, or
    `pdf_dir`. S12 sets `pdf_dir` to a third directory and checks it.
    Text entry of `p` on the month sheet and the form stays as it is.
-   **FR11.** A second export to the same default path overwrites
    that file, leaves a single file of that name, mode `0600`, and
    does not write a second copy in the working directory. This holds
    for the CLI and for TUI `p`.
-   **FR12.** The command reference, the config reference, the key
    reference, and the monthly-chart-PDF how-to describe this
    directory, `pdf_dir`, `-o`, and the one-line path on standard
    output. They no longer say the default is the current directory.
    They say a leading `~/` is the home directory and that other
    relative `pdf_dir` values fail the export.
-   **FR13.** Picture, page size, colours, clip, filename pattern,
    and `NO_COLOR` not applying stay as in the 2026-09-12 spec. File
    mode stays `0600`. This slice does not add other PDF kinds.

## Out of scope

-   PDF of filled or blank monthly log sheets
-   PDF of long-term charts
-   ntcharts, or moving the TUI to Bubble Tea v2
-   Moving the database, or naming the PDF from the database path
-   Changing `YYYY-MM-chart.pdf`
-   Confirm-before-overwrite
-   An environment variable for the PDF directory
-   Shell expansion of `$` inside `pdf_dir`, and of `~` other than a
    leading `~/` prefix
-   Changing what `p` types on the month sheet or the form
-   Rewriting the 2026-09-12 spec file

## Documentation

Update the how-to for exporting a monthly chart PDF, and the
reference lines for `-chart-pdf`, `-o`, `pdf_dir`, and `p`, so they
match this directory rule. Diátaxis split unchanged.
