# Plan — Filled log-sheet PDF

**Spec**:
[`docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`](../specs/2026-10-17-filled-log-sheet-pdf.md)
**Slice**: S1 of
[`filled-log-sheet-pdf.md`](../slices/filled-log-sheet-pdf.md)
**Follows**:
[`2026-10-03-pdf-destination.md`](../plans/2026-10-03-pdf-destination.md)
**Status**: approved

No production code until the spec's scenarios exist as failing
tests.

This slice adds `hdtools -log-pdf YYYY-MM`: one portrait page of
the filled month, file `YYYY-MM-log.pdf`, same directory rule as
the chart. It does not add a TUI key. Issue #113 stays out. Do
not edit `internal/chartpdf`. Do not edit `internal/tui`. Do not
edit `internal/chartspan`. Do not import `fpdf` from the TUI, and
do not import the new writer from the TUI. `cmd/hdtools` only
wires the flag.

The `idea.md` change-record link is already in place. Do not
rewrite the historical sentences that say filled log sheets were
not in the chart-PDF slice or the PDF-destination slice. Do not
describe a TUI key there.

Do not extract the month-sheet row builder or the delta formatter
into a shared package. Copy the rules into the writer. The tests
are what keep the page on the sheet.

## Module structure

| File | Why |
| --- | --- |
| `internal/logpdf/logpdf.go` | New package. One concern: paint one filled month as a portrait page. It does not open SQLite, resolve `pdf_dir`, read `[colors]` or `NO_COLOR`, clip at today, or import `internal/tui`, `internal/chartpdf`, or `internal/chartspan`. |
| `internal/logpdf/logpdf_test.go` | Page box, rules, font, heading, calendar rows including day 31's baseline, carry, a series with no weigh-in, delta, blanks, notes measured after skipped characters, and the temp-file failure. |
| `cmd/hdtools/main.go` | Flag `-log-pdf`. Empty `-o` joins `YYYY-MM-log.pdf` onto `ResolvePDFDir`. Non-empty `-o` is the file path and does not call the resolver. Both export flags set is an error before config, database, or PDF. Bad `YYYY-MM` fails before those too, same as `-chart-pdf`. Success prints the path. |
| `cmd/hdtools/main_test.go` | Directory, `-o`, failures, bad month, both flags before config and the database, `-o` alone, chart flag unchanged, stored kilograms, carry through the command, empty log, a month with no weigh-in, duplicate day, unsorted days. |
| `docs/reference/cli.md` | Synopsis and flag table gain `-log-pdf`. The `-o` line covers both exports and still does not say the working directory is the default. Success and failure mention the log sheet. `NO_COLOR` does not grey either PDF; the log sheet is black, and the chart PDF stays in colour. |
| `docs/reference/config.md` | One sentence on `pdf_dir`: the filled log-sheet PDF uses this same directory. Do not rewrite the rest of the section. |
| `docs/how-to/monthly-log-pdf.md` | New how-to for the command only. Directory, `-o`, bad month, both flags. No TUI key. Not medical advice and no loss numbers. |
| `docs/how-to/README.md` | One bullet for that how-to, beside the chart PDF how-to. |
| `CHANGELOG.md` | When the behaviour ships, one Unreleased bullet. Not in the failing-test step. |

Do not edit `docs/how-to/monthly-chart-pdf.md` or
`docs/reference/keys.md`. Do not add a module. `github.com/go-pdf/fpdf`
is already required. Do not edit the slicing record.

`internal/logpdf/logpdf.go` opens with a literate preamble: why
the sheet is its own writer, that the rows match the month sheet
without sharing its package, and what the file does not do
(directory rule, chart clip, colours, SQLite, a second page).
Keep the file at or under 500 lines. If the painter would pass
that, put the `0600` install in another file in this package
before adding more paint. Do not pre-split.

The how-to, the reference edits, and `CHANGELOG.md` land with the
implementation, after the tests are green. The failing-test step
does not write them.

## Algorithm notes

-   `Write` receives a full path and already-trended logs. It does
    not call `ApplyTrend` and it does not call `ResolvePDFDir`.

    ```go
    type Options struct {
        Year  int
        Month time.Month
        Logs  []dailylog.DailyLog
        Unit  units.Unit
        Today time.Time
    }

    func Write(path string, opts Options) error
    ```

    `Today` is how a test proves the sheet does not clip. Every
    calendar day of `Year`/`Month` is a row, including days after
    `Today`. Do not call `chartspan.Clip`. A zero `Today` must
    not drop days either.
-   Page: `fpdf.New("P", "pt", "Letter", "")`, auto page break
    off. Media box 612 by 792. One page. Black text and black
    rules. No filled title box. Font family `Helvetica` only.
    Do not call `AddFont`. Heading is Helvetica 12. Header and
    body are Helvetica 8. Heading text is
    `fmt.Sprintf("%s %d (%s)", month, year, unit)` with the unit
    token `kg`, `lb`, or `st`.
-   Geometry, so 31 rows fit and the note tests are stable.
    Margins 36 pt. Content width 540. Row pitch 18 pt, so the
    heading, the header, and 31 day rows sit on the page with
    room below. One line per row: no multi-line cell. Day 31's
    baseline is inside the media box. A test that only extracts
    the text `31` does not pass. Do not add a second page. Column
    widths, left to right: day 28, weekday 36, weight 40,
    trend 48, delta 44, sleep 36, steps 52, workout 36, note the
    remainder (220). Weekday is 36 pt because Helvetica 8 draws
    that word at 31.57 pt, and a 28 pt column crosses into weight.
    Those 8 pt come from weight. Each header word's drawn end
    falls before the next vertical rule. Vertical rules sit on
    the column boundaries.
    A horizontal rule under the header and under each day row.
    Each cell is its own text operation, in that column order,
    so tests read the literals in reading order.
-   Weekday is `day.Weekday().String()[:2]` in UTC, the same two
    letters as the month sheet. Day text is the day number with
    no zero pad.
-   Carry matches `buildMonthSheet`: a logged day has a trend
    when its weight is non-nil or its trend is non-zero. A day
    with no row uses the last log on or before that day that
    passes the same test. Do not import the TUI. Days before the
    first such log have no trend.
-   Display numbers match the month sheet, not a kilogram
    residual. Convert with `units.FromKG`. Paint a cell with
    `%.1f`. Delta parses those two one-decimal texts back to
    numbers, subtracts, and formats `%.1f`. If that text is
    `0.0` or `-0.0`, the cell is `0.0`. If the painted weight is
    greater than the painted trend, prefix `+`. Otherwise use the
    formatted text, which already has the minus. Blank when the
    weight is nil or the day has no trend. Do not import
    `formatDelta`.
-   A day with a row: sleep `%.1f` (so `0.0` and `8.0`), steps
    `%d` with the full integer, workout `yes` or `no`. A day
    with no row: sleep, steps, workout, and note have no text.
    A row with a nil weight still prints sleep, steps, and
    workout. A blank cell draws no glyph. It is not `-`, not an
    em dash, and not the word blank. An empty stored note is
    that same empty cell. A minus sign on a negative delta stays.
-   Notes: drop runes the core font cannot draw before measuring.
    Printable ASCII (U+0020 through U+007E) stays. A rune that
    cp1252 maps to a glyph other than `.notdef` stays, encoded
    as that byte, so `é` is the cp1252 byte `0xE9` in the content
    stream. Every other rune is omitted, including controls and
    newlines. Do not use fpdf's translator substitution of `.`.
    Do not substitute `?`. If nothing remains, the cell text is
    empty. Then, if the Helvetica width of the remaining
    note is at most the note column's inner width (column width
    minus 4 pt), print it unchanged. Otherwise print the longest
    rune prefix such that the prefix plus `...` fits that inner
    width. The ellipsis is three ASCII periods, not U+2026.
    The note column is wider than `...`. Parentheses stay;
    `Text` escapes them. Do not trim the note.
-   Install matches the chart's owner-only write, in this
    package, without calling `chartpdf`. `MkdirAll` the parent at
    `0755` when it is not `.`. Write the PDF bytes to a temp in
    that directory (`CreateTemp` prefix `.hdtools-log-`),
    `Chmod 0600`, close, `Rename` onto the destination. Remove
    the temp if rename does not happen. Do not `os.Create` or
    `OutputFileAndClose` on the destination. Wrap each step with
    `%w`: `create temp log`, `chmod log`, `write log`,
    `close log`, `rename log`. The outer error is
    `write log pdf %s`. Do not assert the created directory's
    mode; umask masks `0755`.
-   CLI, in order, after flag parse: if both `-log-pdf` and
    `-chart-pdf` are non-empty, return an error that names both
    flags before `config.Ensure` and before opening a database.
    Do not create config, a database, a PDF, or `-o`. An invalid
    month does not change that. If only one flag is set, reject
    a bad month before config and the database, with
    `want YYYY-MM` in the error. `-log-pdf` uses the prefix
    `log-pdf:`, then `run` wraps a later failure as `log pdf:`,
    parallel to `chart-pdf:` and `chart pdf:`. Load `store.All`.
    If any day is not strictly after the previous day, fail the
    export. A duplicate day and an unsorted series both fail,
    including when no row has a weigh-in. If the slice is empty,
    or no row has a non-nil weight, skip `ApplyTrend`. Otherwise
    call `ApplyTrend` on the suffix that begins at the first
    non-nil weight, with a nil carry. Do not pass the leading
    weightless rows. Do not change `ApplyTrend`. Do not treat
    `ErrNoTrendStart` as an export failure. Any other
    `ApplyTrend` error fails the export. Empty `-o` is
    `filepath.Join(ResolvePDFDir(cfg), YYYY-MM-log.pdf)`.
    Non-empty `-o` is the path unchanged.
    Success is `fmt.Println` of that path and no TUI. Failure
    prints no path. `-o` flag text must not say the working
    directory is the default; it names both export flags.
    Usage for `-log-pdf` says it writes the sheet and exits.
-   `-o` alone is the existing fall-through to `tea.NewProgram`.
    Do not treat a non-empty `-o` as an export.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestWriteLogPDFPortraitOnePage`, `TestLogPDFDefaultWritesDataDir`, `TestWriteLogPDFCalendarDays` |
| FR2 | `TestWriteLogPDFHeading`, `TestWriteLogPDFOmitsChrome` |
| FR3 | `TestWriteLogPDFCalendarDays` |
| FR4 | `TestWriteLogPDFIncludesDaysAfterToday` |
| FR5 | `TestWriteLogPDFCarriedTrend`, `TestLogPDFCarryFromEarlierMonth`, `TestWriteLogPDFEmptySeries`, `TestLogPDFEmptyLog`, `TestWriteLogPDFNoWeighIn`, `TestLogPDFNoWeighIn`, `TestLogPDFTrendError` |
| FR6 | `TestWriteLogPDFStoneOneDecimal` |
| FR7 | `TestWriteLogPDFDeltaKg`, `TestWriteLogPDFDeltaFromPaintedPounds`, `TestWriteLogPDFFirstWeighInPounds`, `TestWriteLogPDFExactZero`, `TestWriteLogPDFDeltaBlank` |
| FR8 | `TestWriteLogPDFSleepStepsWorkout`, `TestWriteLogPDFBlankDayEmpty` |
| FR9 | `TestWriteLogPDFBlankDayEmpty`, `TestWriteLogPDFNoteCut` |
| FR10 | `TestWriteLogPDFNoteCut`, `TestWriteLogPDFSkipsUndrawable` |
| FR11 | `TestWriteLogPDFSkipsUndrawable`, `TestWriteLogPDFNoteCut` |
| FR12 | `TestLogPDFDefaultWritesDataDir` |
| FR13 | `TestLogPDFDirAbsolute`, `TestLogPDFHomePrefix`, `TestLogPDFRelativeDirFails`, `TestLogPDFNonStringPDFDir`, `TestLogPDFOAbsoluteWins`, `TestLogPDFRelativeOStaysInCwd`, `TestLogPDFDatabasePathDoesNotMovePDF` |
| FR14 | `TestLogPDFFailureNoCwdCopy`, `TestWriteLogPDFFailedWriteLeavesOriginal`, `TestLogPDFRelativeDirFails`, `TestLogPDFTrendError` |
| FR15 | `TestLogPDFOverwrite` |
| FR16 | `TestLogPDFRejectsBadMonth` |
| FR17 | `TestLogPDFBothFlagsWriteNothing` |
| FR18 | `TestOAloneStartsTUIWritesNothing`, `TestChartPDFDoesNotWriteLogPDF` |
| FR19 | `TestLogPDFDoesNotRewriteKilograms` |
| FR20 | doc files in the module table (no Go test) |

Existing `-chart-pdf` tests stay on their assertions and their
`-o` paths. Do not retarget them. `TestChartPDFDoesNotWriteLogPDF`
is additive.

## Test list

Writer tests live in `package logpdf`. They may use `t.Parallel`
when they only touch `t.TempDir`. Build logs with `dailylog.New`
and set `Trend` in the test when the scenario is about painted
cells. Do not open SQLite in this package. Copy a small literal
extractor in the test file. Do not import `internal/chartpdf`.
A blank cell has no text. Heading and numbers are their
own literals.

CLI tests follow `TestChartPDFDefaultWritesDataDir`:
`isolateCLI` or `t.Chdir` plus `t.Setenv`, never `t.Parallel` on
those tests, never the real home. Reuse `wantStdoutLine` and
`wantNoStdoutPath`. Filename `1990-11-log.pdf`.

-   `TestWriteLogPDFPortraitOnePage` — S1. `%PDF`, mode `0600`,
    one page, MediaBox 612 by 792, `/Helvetica`, no `/FontFile`,
    no image, and at least one vector rule per November day row.
    `Write` takes no colours. The CLI default test also sets
    `NO_COLOR` and a `[colors]` weight and still gets that page.
-   `TestWriteLogPDFHeading` — S2. `lb`, `kg`, and `st` headings
    are exactly `November 1990 (lb)`, `November 1990 (kg)`, and
    `November 1990 (st)`.
-   `TestWriteLogPDFOmitsChrome` — S3. A filled November page has
    no database path, no `arrows move`, no `Monthly loss`, no
    `Daily deficit`, no `(empty month)`, and no `hdtools`.
-   `TestWriteLogPDFCalendarDays` — S4. November 1990 literals
    include day `1` with `Th` and day `3` with `Sa`, days `1`
    through `30`, and no literal `31`. February 1990 has `28`
    and no `29`. February 1992 has `29`. January 1991 has days
    `1` through `31`, one page, and day 31's text position has
    its baseline inside the media box (`x` in `[0, 612]`, `y` in
    `[0, 792]`). Extracting the string `31` is not that oracle.
    Header literals are `day`, `weekday`, `weight`, `trend`,
    `delta`, `sleep`, `steps`, `workout`, `note`, in that order.
-   `TestWriteLogPDFIncludesDaysAfterToday` — S5. `Today` is 10
    November 1990, a weigh-in on the 15th is on the page, and
    day `30` is a row.
-   `TestWriteLogPDFCarriedTrend` — S6 writer half. An October 31
    log at 80 kg with its trend set, and no November row: day 1
    trend literal `80.0`, and no text for weight or delta. With
    `Today` on the 10th and the only November weigh-in that day,
    day 15's trend is the carry and its weight has no text.
-   `TestWriteLogPDFEmptySeries` — S7 writer half. Nil logs:
    days 1 through 30, value cells have no text, no
    `(empty month)`.
-   `TestWriteLogPDFDeltaKg` — S8. 80.5 against trend 80.0 is
    `80.5`, `80.0`, `+0.5`. 79.5 against 80.0 is `-0.5`.
-   `TestWriteLogPDFDeltaFromPaintedPounds` — S9. Stored 80.0 kg
    and trend 79.9 kg in `lb`: `176.4`, `176.1`, `+0.3`, and the
    delta literal is not `+0.2`.
-   `TestWriteLogPDFFirstWeighInPounds` — S10. Store
    `units.ToKG(176.5, lb)`, run `ApplyTrend`, pass that series.
    Cells `176.5`, `176.6`, `-0.1`.
-   `TestWriteLogPDFExactZero` — S11. Equal painted cells: delta
    literal `0.0`, not `+0.0` or `-0.0`.
-   `TestWriteLogPDFDeltaBlank` — S12. Carried trend and no
    weight: delta has no text, trend still the number. A day
    before the first weigh-in: trend and delta have no text.
-   `TestWriteLogPDFStoneOneDecimal` — S13. 80 kg in `st` is
    weight `12.6`, trend `12.6`, delta `0.0`.
-   `TestWriteLogPDFSleepStepsWorkout` — S14. Zeros are `0.0`,
    `0`, `no`. 7.5 hours, 100000 steps, workout true are `7.5`,
    `100000`, `yes`, one page. A nil weight with sleep 8 prints
    no text for weight, sleep `8.0`, steps `100`, workout `no`.
-   `TestWriteLogPDFBlankDayEmpty` — S15. A day with no row has
    no text for sleep, steps, workout, or note. Those cells are
    not `0.0`, `0`, `no`, `-`, or an em dash.
-   `TestWriteLogPDFNoteCut` — S16. `walk` is `walk` and does not
    gain `...`. `say (hi)` is `say (hi)`. Two hundred `x`
    characters end with `...`, do not contain the whole note,
    and the page count is 1. An empty stored note is an empty
    literal. `say (hi)` plus U+2603 is `say (hi)` and
    does not contain `...`. A run of `x` that fits, with a
    snowman inserted so the run plus the width of `.` does not
    fit, is that run of `x` with no `...`. Drop the snowman
    before measuring. Measuring the raw note, or counting the
    snowman as `.`, fails this test.
-   `TestWriteLogPDFSkipsUndrawable` — S17. `café` contributes
    the byte `0xE9`, not a missing page. `a` + U+2603 + `b` is
    the literal `ab`, not `a.b` or `a?b`. A note that is only
    U+2603 has no text, and the write succeeds.
    `ab\ncd` is the literal `abcd` on one page. The same drop
    happens before `TestWriteLogPDFNoteCut` measures.
-   `TestWriteLogPDFFailedWriteLeavesOriginal` — S26 writer
    half. Destination bytes `OLD`, parent mode `0555`. Error
    contains `create temp log`. Bytes stay `OLD`. No
    `.hdtools-log-` file remains after the parent is writable
    again.

-   `TestLogPDFDefaultWritesDataDir` — S1 and S18. File only at
    the data path, mode `0600`, header `%PDF`, portrait box,
    heading present, stdout one line of that path, no TUI, no
    working-directory PDF, no `1990-11-chart.pdf`, no leftover
    temp. `NO_COLOR` is set and `[colors]` names a weight colour.
-   `TestLogPDFDirAbsolute` — S19. Absolute `pdf_dir`, file
    there, not under the data directory, missing directory
    created, stdout that path.
-   `TestLogPDFHomePrefix` — S23. `~/charts` with `HOME` set.
    File under that home, not the data directory or the working
    directory.
-   `TestLogPDFRelativeDirFails` — S22. `pdf_dir` `charts` exits
    non-zero, no PDF in the working directory, under the home,
    or under the data directory, stdout empty, no TUI.
-   `TestLogPDFNonStringPDFDir` — S24. `pdf_dir = 3` with no
    `-o` fails, the error names `pdf_dir`, stdout empty, no PDF.
    With `-o`, exit 0 and the file is only at `-o`.
-   `TestLogPDFOAbsoluteWins` — S20. Absolute `-o`, missing
    `pdf_dir` still missing, file only at `-o`, stdout that path.
-   `TestLogPDFRelativeOStaysInCwd` — S21. `-o out.pdf` is that
    name in the working directory. Stdout is `out.pdf\n`. No
    `1990-11-log.pdf` under the data directory or `pdf_dir`.
-   `TestLogPDFDatabasePathDoesNotMovePDF` — S25. `-db` outside
    the data directory. The PDF is in the data directory only.
-   `TestLogPDFFailureNoCwdCopy` — S26 CLI. The default log path
    is a directory. Non-zero, error contains `log pdf:`, path
    still a directory, stdout empty, no working-directory
    `1990-11-log.pdf`.
-   `TestLogPDFOverwrite` — S27. Seed `old`, run twice. Exactly
    one `1990-11-log.pdf`, header `%PDF`, mode `0600`, no chart
    file and no backup.
-   `TestLogPDFRejectsBadMonth` — S28. `1990-13` and `banana`,
    each with `-o`. Non-zero, error contains `want YYYY-MM`, no
    `-o` file, no config, no database, stdout empty.
-   `TestLogPDFBothFlagsWriteNothing` — S29. Both flags
    `1990-11`, and a subtest with one flag `banana`. Non-zero
    before `config.Ensure` and before the database is opened.
    No log PDF, no chart PDF, no `-o` file, no new config or
    database, stdout empty, `run` returns.
-   `TestOAloneStartsTUIWritesNothing` — S30. Call `run` in a
    goroutine with only `-o`. Do not use `runCLIOutput` (its
    timeout is a fatal). If `run` returns, the error must not
    contain `log pdf` or `chart pdf`, and it must not be nil.
    If it is still running after a second, that is the TUI.
    Either way `out.pdf` is absent, the data-directory log PDF
    is absent, and stdout has no path.
-   `TestChartPDFDoesNotWriteLogPDF` — S31. `-chart-pdf 1990-11`
    writes `1990-11-chart.pdf` with MediaBox 792 by 612 and does
    not write `1990-11-log.pdf`.
-   `TestLogPDFDoesNotRewriteKilograms` — S32. After a pound
    export, the stored weight is still 80 kg.
-   `TestLogPDFCarryFromEarlierMonth` — S6 command half. Only
    October 31 is stored, at 80 kg. The November page's day 1
    trend is `80.0`. This fails if the command trends November
    alone.
-   `TestLogPDFEmptyLog` — S7 command half. No rows. Exit 0, a
    portrait file, day `30` present, no `(empty month)`.
-   `TestWriteLogPDFNoWeighIn` — S33 writer half. Every row has a
    nil weight and trend `0`. Weight, trend, and delta have no
    text. Sleep, steps, workout, and note still print. A series
    whose first non-nil weight is not the first row: leading
    days have no text for trend or delta, and the first
    weigh-in's trend is that weight rounded to one decimal.
-   `TestLogPDFNoWeighIn` — S33 command half. Rows with sleep,
    steps, workout, and a note, and no weigh-in. Exit 0, the
    page is written, those cells print, weight, trend, and delta
    have no text, stdout is one path. This is not the TUI load
    error.
-   `TestLogPDFTrendError` — S34. Two rows on the same day, and a
    subtest with a later day stored first. Non-zero, no PDF,
    stdout empty, no TUI. A further subtest uses the same bad
    days with nil weights and still writes nothing. A single
    nil-weight row is not this failure.
