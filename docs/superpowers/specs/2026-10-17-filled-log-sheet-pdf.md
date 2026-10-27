# PDF of one filled monthly log sheet

**Date**: 2026-10-17
**Status**: draft
**Issue**: [#110](https://github.com/jglueckstein/hdtools/issues/110)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Slice**: S1 of
[`filled-log-sheet-pdf.md`](../slices/filled-log-sheet-pdf.md)
**Follows**:
[`2026-10-03-pdf-destination.md`](2026-10-03-pdf-destination.md)
for the directory rule (that spec's Decisions 1–10). The filename
here is `YYYY-MM-log.pdf`. This file does not amend the chart
picture in
[`2026-09-12-pdf-monthly-charts.md`](2026-09-12-pdf-monthly-charts.md).

The month sheet is already the paper log: every calendar day, the
fixed columns, a carried trend, and a delta that is not stored.
`hdtools -log-pdf YYYY-MM` writes that sheet as one portrait page
and exits. It does not start the TUI, and it does not write the
monthly chart.

Issue #113 is the later key on the month sheet. This slice has no
TUI key.

## User stories

### US1 — Take the month sheet home

As a person who keeps a month of logs, I want
`hdtools -log-pdf YYYY-MM` to write one portrait page of that
month's filled sheet and then exit, so that I can print the month
without a screenshot and without opening the TUI.

### US2 — Read the same days as the screen

As a person who already reads the month sheet, I want every
calendar day and the same columns and display-unit figures,
including a blank day's carried trend, so that the paper matches
the screen.

### US3 — Keep the note on one page

As a person whose note is longer than the line, I want that note
cut on the row and the month kept to one page, so that the sheet
still prints as a single page.

### US4 — File it beside the chart

As a person who already exports chart PDFs, I want
`YYYY-MM-log.pdf` in that same directory, with `-o` still naming
one whole file, so that I do not learn a second place to look.

## Decisions

1.  **Command and page.** `hdtools -log-pdf YYYY-MM` writes one
    portrait US Letter page and does not start the TUI. The page
    box is 612 by 792 points. The page is a ruled vector table:
    rules between the columns and between the day rows, text in
    the cells, black ink. It is not a screenshot of the terminal
    and not the chart. There is no embedded image and no embedded
    font program. `[colors]` and `NO_COLOR` do not change the
    page. The chart's coloured title box is not on it.
2.  **Heading.** The heading is the English month name, the year,
    and the display unit in parentheses, for example
    `November 1990 (lb)`. The unit token is `kg`, `lb`, or `st`.
    The heading does not include `hdtools` or the word weight.
    The column header is still `weight`. The page does not print
    the database path, the key help, Monthly Loss, Daily Deficit,
    or `(empty month)`.
3.  **Columns.** Left to right: day, weekday, weight, trend,
    delta, sleep, steps, workout, note. A header row uses those
    words. Then one row per calendar day, one line high. The day
    cell is the day number, not zero-padded (`1`, not `01`). The
    weekday cell is the first two letters of the English weekday
    name (`Th` for Thursday). November 1, 1990 is `Th`. Days with
    no database row are still rows. A 28-day February has no 29th
    row. A 29-day February has a 29th row. November 1990 has days
    1 through 30 and no day 31.
4.  **Do not clip at today.** Days after today are rows when the
    named month is the current one. A weigh-in on a later day in
    that month is on the page. This slice does not use the chart
    clip. Carried trend still shows on a day after today.
5.  **Trend.** The command loads the full series. When the series
    is not empty it runs `ApplyTrend` before the month is built.
    An empty log skips `ApplyTrend` and is still a sheet of blank
    days. A day with no weigh-in shows the trend carried from the
    last earlier day that has a weigh-in or a computed trend, the
    same carry the month sheet shows. A day before the first
    weigh-in has no trend. If `ApplyTrend` rejects the series, the
    export fails, prints no path, and writes no file.
6.  **Weight, trend, and delta.** Those three cells are the
    display unit, to one decimal, using the same one-decimal text
    the month sheet paints. Delta is that weight text minus that
    trend text, not a kilogram gap converted afterwards. A
    positive delta has a leading `+`. When the two painted cells
    are equal, delta is `0.0`, never `+0.0` or `-0.0`. Delta is
    blank when the day has no weigh-in or no trend. Stone is that
    same single decimal (`12.6` for 80 kg), not stones and
    pounds. Kilograms stay in the database. The export does not
    rewrite a stored weight.
7.  **Sleep, steps, workout, and blanks.** A day with a row prints
    sleep to one decimal, steps as an integer, and workout as
    `yes` or `no`, including zeros (`0.0`, `0`, and `no`). A row
    with no weigh-in still prints those three, and its weight cell
    is blank. A day with no row leaves sleep, steps, workout, and
    note blank. A blank cell is one hyphen-minus (`-`), a dash
    the font draws as that character. It is not an empty gap, not
    the word blank, and not the screen's em dash. The page is not
    a screenshot. An empty stored note is empty, not `-`, because
    the note is the stored text and an empty note fits.
8.  **Notes and the font.** The note is the stored note, not
    trimmed. A note whose drawn width fits the note column is not
    cut and gains no ellipsis. A longer note is shortened on that
    same line and ends with three periods (`...`). The dropped
    tail is not elsewhere on the page. The note does not wrap,
    and the page count stays one. Numeric cells are not cut.
    The font is the chart's built-in Helvetica. There is no new
    font file. A character that font cannot draw is left out of
    the cell, not replaced with `.` or `?`. The rest of the note
    stays, in order. A newline is one of those skipped
    characters, not a second line. `é` is kept. A snowman is not.
    If nothing drawable remains, the note cell is empty, not
    `-`, and the export still writes the page.
9.  **Filename and directory.** The file name is
    `YYYY-MM-log.pdf`. The directory is the chart rule in the
    PDF-destination spec: `ResolvePDFDir`, `pdf_dir`, and `-o` as
    a whole file path. `-o` is not joined to the directory and is
    not trimmed. On success, standard output is one line, that
    path. On failure, standard output has no path, nothing is
    written in the working directory, and there is no fallback
    directory or second name. The file is mode `0600`, written to
    a same-directory temp and renamed into place. A failed write
    does not replace an existing file's bytes and does not leave
    the temp behind. A missing destination directory is created
    the same way the chart PDF creates one. A second run
    overwrites that one name. The run does not write
    `YYYY-MM-chart.pdf`.
10. **Other flags.** A bad `YYYY-MM` fails the same way
    `-chart-pdf` does: non-zero, no TUI, no PDF, and no config or
    database created for that invocation. If `-log-pdf` and
    `-chart-pdf` are both set, the process exits non-zero, writes
    nothing, and does not start the TUI, including when `-o` is
    set and including when a month string is invalid. `-o` alone,
    with neither export flag, still only starts the TUI and
    writes no PDF. Absent `-log-pdf`, `-chart-pdf` is unchanged:
    it still writes the chart page as `YYYY-MM-chart.pdf` and
    does not write `YYYY-MM-log.pdf`.
11. **Where the code lives.** The writer is a new package under
    `internal/`. `cmd/hdtools` only wires the flag. This slice
    does not edit the chart painter. The TUI gains no key and
    does not import a PDF library. The month-sheet row builder
    stays where it is. The check is that the PDF rows match the
    sheet, not that the two screens share a package.
12. **Out of scope for the page.** Blank sheets, long-term chart
    PDFs, a second PDF directory, a new config key, a
    confirm-before-overwrite prompt, a Unicode font file, colour
    roles, and Monthly Loss or Daily Deficit on this page.

## Acceptance scenarios

November 1990 is the fixture unless a scenario names another
month. The working directory, the data directory, and any
`pdf_dir` are different temporary directories unless a scenario
says the file belongs in the working directory. Tests set
`XDG_DATA_HOME`, or a stand-in home, to a temporary directory.
They must not use the developer's real home.

A successful file starts with `%PDF`, is mode `0600`, has one
page, and has a page box of 612 by 792 points. Cell text is read
back from the PDF. Each cell is its own text, so the day number
`1` is not confused with a weight, and a blank cell is the text
`-`.

`Today`, when a scenario sets it, is 10 November 1990. The sheet
still lists the whole month.

### S1 — The command writes one portrait page and skips the TUI

**Given** a log in November 1990
**When** `-log-pdf 1990-11` runs
**Then** the process exits 0 and does not start the TUI
**And** the file is one page, page box 612 by 792 points
**And** the page has vector rules for the columns and the day
rows
**And** the page has no embedded image and no embedded font
**And** the font is the built-in Helvetica
**And** the page is still written when `NO_COLOR` is set and
`[colors]` names a weight colour

### S2 — The heading names the month, the year, and the unit

**Given** a log in November 1990 and `display_unit` `lb`
**When** `-log-pdf 1990-11` runs
**Then** the page's heading text is `November 1990 (lb)`

**Given** the same month and `display_unit` `kg`
**When** the sheet is written
**Then** the heading text is `November 1990 (kg)`

**Given** the same month and `display_unit` `st`
**When** the sheet is written
**Then** the heading text is `November 1990 (st)`

### S3 — The page is not the chart and not the screen chrome

**Given** a log in November 1990
**When** the sheet is written
**Then** the page does not contain the database path
**And** it does not contain the key help
**And** it does not contain `Monthly loss` or `Daily deficit`
**And** it does not contain `(empty month)`
**And** it does not contain `hdtools`

### S4 — One row for every calendar day

**Given** November 1990 with a row on day 1 and no row on day 3
**When** the sheet is written
**Then** the header reads day, weekday, weight, trend, delta,
sleep, steps, workout, note, in that order
**And** there is a row for each day from 1 through 30
**And** there is no day 31
**And** day 1's weekday is `Th` and day 3's weekday is `Sa`
**And** day 3 is still a row

**Given** February 1990
**When** the sheet is written
**Then** there are rows for days 1 through 28 and no day 29

**Given** February 1992
**When** the sheet is written
**Then** day 29 is a row

### S5 — Days after today stay on the sheet

**Given** today is 10 November 1990
**And** November 15 has a weigh-in and November 30 has no row
**When** the November 1990 sheet is written
**Then** November 15's weight is on the page
**And** day 30 is a row
**And** the page does not stop at day 10

### S6 — A blank day shows the carried trend

**Given** 31 October 1990 has a weigh-in of 80 kg and no other
days are logged
**And** `display_unit` is `kg`
**When** `-log-pdf 1990-11` runs
**Then** November 1's trend cell is `80.0`
**And** November 1's weight cell and delta cell are `-`
**And** the series used for that trend is the full log, not
November alone

**Given** today is 10 November 1990
**And** the only November weigh-in is 80 kg on November 10
**When** the November sheet is written
**Then** November 15's trend cell is that carried trend
**And** November 15's weight cell is `-`

### S7 — An empty log is still a sheet of days

**Given** a database with no rows
**When** `-log-pdf 1990-11` runs
**Then** the process exits 0
**And** the page has a row for each day from 1 through 30
**And** weight, trend, delta, sleep, steps, workout, and note on
those rows are `-`
**And** the page does not contain `(empty month)`

### S8 — Kilogram delta matches the painted cells

**Given** `display_unit` is `kg`
**And** a day whose weight is 80.5 kg and whose trend is 80.0 kg
**When** the sheet is written
**Then** that row's weight is `80.5`, trend is `80.0`, and delta
is `+0.5`

**Given** a day whose weight is 79.5 kg and whose trend is 80.0 kg
**When** the sheet is written
**Then** the delta cell is `-0.5`

### S9 — Pound delta subtracts the painted cells

**Given** `display_unit` is `lb`
**And** a day whose stored weight is 80.0 kg and whose trend is
79.9 kg
**When** the sheet is written
**Then** the weight cell is `176.4` and the trend cell is `176.1`
**And** the delta cell is `+0.3`
**And** the delta cell is not `+0.2`

### S10 — The first pound weigh-in is not zero

**Given** `display_unit` is `lb`
**And** the first weigh-in in the series was entered as 176.5 lb
and stored as kilograms
**When** the sheet is written
**Then** the weight cell is `176.5`, the trend cell is `176.6`,
and the delta cell is `-0.1`

### S11 — An exact zero prints 0.0

**Given** a day whose painted weight cell and painted trend cell
are the same one-decimal text
**When** the sheet is written
**Then** the delta cell is `0.0`
**And** it is not `+0.0` or `-0.0`

### S12 — Delta is blank when an operand is missing

**Given** a day with no weigh-in and a carried trend
**When** the sheet is written
**Then** the delta cell is `-`
**And** the trend cell still shows the carry

**Given** a day before the first weigh-in
**When** the sheet is written
**Then** the trend cell and the delta cell are `-`

### S13 — Stone is one decimal

**Given** `display_unit` is `st`
**And** a day whose weight and trend are both 80 kg
**When** the sheet is written
**Then** the weight cell and the trend cell are `12.6`
**And** the delta cell is `0.0`
**And** the cells do not show a stones-and-pounds split

### S14 — A logged day prints sleep, steps, and workout

**Given** a day with a row, sleep 0, steps 0, and workout false
**When** the sheet is written
**Then** sleep is `0.0`, steps is `0`, and workout is `no`

**Given** a day with sleep 7.5, steps 100000, and workout true
**When** the sheet is written
**Then** sleep is `7.5`, steps is `100000`, and workout is `yes`
**And** the step count is not shortened
**And** the page count is still one

**Given** a row with no weigh-in, sleep 8, steps 100, and
workout false
**When** the sheet is written
**Then** the weight cell is `-`
**And** sleep is `8.0`, steps is `100`, and workout is `no`

### S15 — A day with no row leaves those cells blank

**Given** a calendar day with no database row
**When** the sheet is written
**Then** sleep, steps, workout, and note are each `-`
**And** they are not `0.0`, `0`, `no`, or empty

### S16 — A long note is cut, and a note that fits is not

**Given** a note `walk` that fits the note column
**When** the sheet is written
**Then** the note cell is `walk`
**And** it does not end with an added `...`

**Given** a note `say (hi)` that fits
**When** the sheet is written
**Then** the note cell is `say (hi)`

**Given** a note of 200 `x` characters
**When** the sheet is written
**Then** the note cell ends with `...`
**And** the cell does not contain all 200 `x` characters
**And** the dropped tail is not elsewhere on the page
**And** the page count is one

**Given** a stored note that is empty
**When** the sheet is written
**Then** the note cell is empty, not `-`

### S17 — A character the font cannot draw is skipped

**Given** a note `café`
**When** the sheet is written
**Then** the note still contains `é`
**And** the export writes the page

**Given** a note made of `a`, a snowman (U+2603), and `b`
**When** the sheet is written
**Then** the note cell is `ab`
**And** it is not `a.b` or `a?b`
**And** the export writes the page

**Given** a note that is only a snowman
**When** the sheet is written
**Then** the note cell is empty, not `-`
**And** the export writes the page

**Given** a note `ab` plus a newline plus `cd`
**When** the sheet is written
**Then** the note cell is `abcd` on that day's one line
**And** the page count is one

### S18 — The default file is the log sheet in the data directory

**Given** a log in November 1990
**And** no `-o` and no `pdf_dir`
**And** `XDG_DATA_HOME` points at a temporary directory
**And** the working directory is a different temporary directory
**When** `-log-pdf 1990-11` runs
**Then** the process exits 0 and does not start the TUI
**And** `$XDG_DATA_HOME/hdtools/1990-11-log.pdf` exists
**And** that directory was created if it was missing
**And** the file is mode `0600` and starts with `%PDF`
**And** standard output is one line, that path
**And** the working directory has no `1990-11-log.pdf`
**And** no `1990-11-chart.pdf` is written
**And** no temp PDF is left beside the file

### S19 — `pdf_dir` replaces the data directory

**Given** a log in November 1990 and no `-o`
**And** `pdf_dir` is an absolute directory that does not yet exist
**When** `-log-pdf 1990-11` runs
**Then** `1990-11-log.pdf` exists in that directory, mode `0600`
**And** it does not exist under `$XDG_DATA_HOME/hdtools`
**And** standard output is one line, that path

### S20 — Absolute `-o` is the whole file

**Given** `pdf_dir` is an absolute directory that does not exist
**When** `-log-pdf 1990-11 -o` is given an absolute file path
**Then** the PDF exists only at that path, mode `0600`
**And** it is not under `pdf_dir` and not under the data directory
**And** the missing `pdf_dir` is still missing
**And** standard output is one line, that path

### S21 — Relative `-o` stays in the working directory

**Given** `pdf_dir` is an absolute directory
**When** `-log-pdf 1990-11 -o out.pdf` runs
**Then** `out.pdf` exists in the working directory, mode `0600`
**And** standard output is one line, `out.pdf`
**And** `1990-11-log.pdf` is not under the data directory or
inside `pdf_dir`

### S22 — A relative `pdf_dir` does not fall back

**Given** no `-o`
**And** `pdf_dir` is `charts`
**When** `-log-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** no PDF is written in the working directory, under the
home directory, or under the data directory
**And** standard output has no path

### S23 — A `~/` directory expands to the home directory

**Given** no `-o` and `pdf_dir` is `~/charts`
**And** the home directory is a temporary stand-in
**When** `-log-pdf 1990-11` runs
**Then** the PDF is `charts/1990-11-log.pdf` under that home
**And** it is not in the working directory or the data directory

### S24 — A non-string `pdf_dir` does not block `-o`

**Given** `pdf_dir` is a number (`pdf_dir = 3`) and no `-o`
**When** `-log-pdf 1990-11` runs
**Then** the process exits non-zero
**And** the error names `pdf_dir`
**And** no PDF is written and standard output has no path

**Given** the same non-string `pdf_dir`
**When** `-log-pdf 1990-11 -o` names a file
**Then** the process exits 0 and the PDF exists only at that path

### S25 — The database path does not move the PDF

**Given** no `-o` and no `pdf_dir`
**And** `-db` points at a database outside the data directory
**When** `-log-pdf 1990-11` runs
**Then** the PDF is in the data directory
**And** no PDF is written beside that database

### S26 — Failure writes nothing and keeps the old bytes

**Given** no `-o`
**And** the default `1990-11-log.pdf` path already exists as a
directory
**When** `-log-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** that path is still a directory
**And** standard output has no path
**And** the working directory has no `1990-11-log.pdf`

**Given** an existing log PDF whose bytes are `OLD`
**And** the destination directory cannot accept a new temp file
**When** the sheet is written
**Then** the write fails
**And** the error names the temp-file step
**And** the destination bytes are still `OLD`
**And** no temp PDF is left in that directory

### S27 — A second run overwrites the same name

**Given** the default path already contains the bytes `old`
**And** no `-o` and no `pdf_dir`
**When** `-log-pdf 1990-11` runs, and then runs again
**Then** that directory has exactly one entry named
`1990-11-log.pdf`
**And** the file starts with `%PDF` and is mode `0600`
**And** there is no backup file and no `1990-11-chart.pdf`

### S28 — A bad month fails like the chart flag

**Given** `-log-pdf` is `1990-13` or `banana`
**And** `-o` names a file that does not exist
**When** the command runs
**Then** the process exits non-zero and does not start the TUI
**And** the error says the month must be `YYYY-MM`
**And** that `-o` file is not created
**And** no config file and no database are created
**And** standard output has no path

### S29 — Both export flags write nothing

**Given** `-log-pdf` and `-chart-pdf` are both set
**When** the command runs
**Then** the process exits non-zero and does not start the TUI
**And** no log PDF and no chart PDF are written
**And** an `-o` path is not created
**And** standard output has no path

**Given** both flags are set and one month string is `banana`
**When** the command runs
**Then** the process still writes no file and does not start the
TUI

### S30 — `-o` alone still only starts the TUI

**Given** a log and `-o out.pdf`
**And** neither `-log-pdf` nor `-chart-pdf` is set
**When** the command runs
**Then** the TUI starts
**And** `out.pdf` is not written
**And** `1990-11-log.pdf` is not written in the data directory
**And** standard output has no path

### S31 — The chart flag is unchanged

**Given** a log in November 1990 and no `-log-pdf`
**When** `-chart-pdf 1990-11` runs
**Then** `1990-11-chart.pdf` is written as it is today
**And** the page box is still 792 by 612 points
**And** `1990-11-log.pdf` is not written

### S32 — Stored kilograms stay kilograms

**Given** a stored weight of 80 kg and `display_unit` `lb`
**When** `-log-pdf 1990-11` runs
**Then** the database still holds 80 kilograms for that day
**And** it does not hold 176.4

### S33 — A trend error fails the export

**Given** the only row has no weigh-in, so `ApplyTrend` cannot
start
**When** `-log-pdf 1990-11` runs
**Then** the process exits non-zero and does not start the TUI
**And** no PDF is written
**And** standard output has no path

## Functional requirements

-   **FR1.** `-log-pdf YYYY-MM` writes one portrait page, 612 by
    792 points, exits 0, and does not start the TUI. The page is
    a ruled vector table in built-in Helvetica, with no embedded
    image and no embedded font. `[colors]` and `NO_COLOR` do not
    change it.
-   **FR2.** The heading is the English month name, a space, the
    four-digit year, a space, and the display-unit token in
    parentheses. The page does not contain the database path, the
    key help, Monthly Loss, Daily Deficit, `(empty month)`, or
    `hdtools`.
-   **FR3.** The columns are day, weekday, weight, trend, delta,
    sleep, steps, workout, and note, in that order, with a header
    row of those words. There is one one-line row per calendar
    day of the named month, including days with no row. The day
    number is not zero-padded. The weekday is the first two
    letters of the English weekday name.
-   **FR4.** The sheet does not stop at today. A later day in the
    named month is still a row, including a weigh-in after today.
-   **FR5.** The command loads the full series and, when it is
    not empty, runs `ApplyTrend` before building the month. An
    empty series skips `ApplyTrend` and still writes every day as
    blank cells. A blank day shows the month sheet's carried
    trend, including on a day after today. The chart clip is not
    used. If `ApplyTrend` fails, the export fails and writes
    nothing.
-   **FR6.** Weight and trend are the display unit to one
    decimal, the same text the month sheet paints. Stone is that
    single decimal. 80 kg in stone is `12.6`.
-   **FR7.** Delta is the painted weight cell minus the painted
    trend cell. A positive value has a leading `+`. An exact zero
    is `0.0`, not `+0.0` or `-0.0`. The cell is `-` when the day
    has no weigh-in or no trend. 80.0 kg against a 79.9 kg trend,
    in pounds, is weight `176.4`, trend `176.1`, delta `+0.3`.
    A first weigh-in entered as 176.5 lb is weight `176.5`,
    trend `176.6`, delta `-0.1`.
-   **FR8.** A day with a row prints sleep to one decimal, steps
    as a full integer, and workout as `yes` or `no`, including
    zeros. A missing weigh-in on that row does not blank those
    three cells. A day with no row prints `-` for sleep, steps,
    workout, and note.
-   **FR9.** A blank cell is the hyphen-minus `-`. An empty
    stored note is empty, not `-`.
-   **FR10.** A note that fits the column is printed in full,
    including parentheses. A longer note is cut on that line and
    ends with `...`. The tail is not printed elsewhere. The page
    count stays one. Numeric cells are not cut with an ellipsis.
-   **FR11.** The font is built-in Helvetica. A character it
    cannot draw is omitted, not replaced with `.` or `?`. `é` is
    kept. A snowman is omitted. A note that is only a snowman
    leaves an empty note cell, and the page is still written. A
    newline is omitted rather than starting another line.
-   **FR12.** With no `-o` and no `pdf_dir`, the file is
    `YYYY-MM-log.pdf` in the data directory from the
    PDF-destination spec, mode `0600`, and standard output is one
    line, that path. The working directory does not gain that
    file. `YYYY-MM-chart.pdf` is not written.
-   **FR13.** An absolute `pdf_dir`, a leading `~/`, a non-string
    `pdf_dir`, a relative `pdf_dir` that is not a leading `~/`,
    and `-db` follow the PDF-destination spec. `-o` is the whole
    file path, is not joined to the directory, and is not
    trimmed. Relative `-o` stays in the working directory.
    Standard output on success is one line, that path.
-   **FR14.** Export failure prints no path, writes nothing in
    the working directory, and does not fall back. Covered
    failures: the destination path exists as a directory; the
    destination directory cannot accept the temp file, in which
    case existing bytes are kept and the error names the
    temp-file step; `pdf_dir` is not a usable directory; the
    series cannot be trended.
-   **FR15.** A second export to the same path overwrites that
    file, leaves a single file of that name, mode `0600`, and
    does not write `YYYY-MM-chart.pdf` or a backup.
-   **FR16.** A bad `YYYY-MM` fails the same way `-chart-pdf`
    does: non-zero, no TUI, no file, no new config, no new
    database, and no path on standard output.
-   **FR17.** If `-log-pdf` and `-chart-pdf` are both non-empty,
    the process exits non-zero, writes neither PDF, does not
    create `-o`, and does not start the TUI.
-   **FR18.** `-o` with neither export flag starts the TUI and
    writes no PDF. Absent `-log-pdf`, `-chart-pdf` still writes
    only the chart page, page box 792 by 612 points, and does not
    write `YYYY-MM-log.pdf`.
-   **FR19.** The export does not rewrite stored kilograms.
-   **FR20.** The command reference describes `-log-pdf` and the
    `-o` line covers both export flags. The config reference
    gains one sentence that the log-sheet PDF uses `pdf_dir`.
    A how-to at `docs/how-to/monthly-log-pdf.md` says how to run
    the command. The chart how-to stays the chart page. The key
    reference is unchanged.

## Out of scope

-   A TUI key that writes this file (issue #113)
-   Blank log sheets and long-term chart PDFs
-   Editing the chart painter, the chart page, or `-chart-pdf`
    when `-log-pdf` is absent
-   A second PDF directory, a new config key, or a
    confirm-before-overwrite prompt
-   Embedding a font, or keeping every Unicode character
-   Colour, `[colors]` roles, and `NO_COLOR` as a print switch
-   Monthly Loss and Daily Deficit on this page
-   The chart clip, and `(empty month)` as a log-sheet page
-   Moving the month-sheet row builder into a shared package
-   Rewriting the 2026-09-12 or 2026-10-03 spec files

## Documentation

Update the command reference, the `-o` line, and add one sentence
on `pdf_dir` so the log-sheet file uses the chart's directory.
Add `docs/how-to/monthly-log-pdf.md` for the command. Do not add
the command to the chart how-to, and do not document a TUI key.
