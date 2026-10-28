# Month-sheet key for the filled log PDF

**Date**: 2026-10-28
**Status**: approved
**Issue**: [#113](https://github.com/jglueckstein/hdtools/issues/113)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Slice**: S2 of
[`filled-log-sheet-pdf.md`](../slices/filled-log-sheet-pdf.md)
**Follows**:
[`2026-10-17-filled-log-sheet-pdf.md`](2026-10-17-filled-log-sheet-pdf.md)
(approved, shipped). This file does not amend that spec, its plan,
or the page. The filename and the directory rule stay there. The
sentence in that spec that issue #113 is the later key stays.

The command already writes one filled month. This slice is only
the key on the month sheet. Shift-P, the letter P, writes that
same file for the month on screen. Idle `p` still starts an edit.
Chart `p` still writes only the chart.

No production code until these scenarios exist as failing tests.

## User stories

### US1 — Print the month on screen

As a person looking at the month sheet, I want shift-P to write
that month's filled log sheet into the usual PDF directory and
leave me on the same cell, so that I can print without a path
prompt and without leaving the sheet.

### US2 — Keep typing

As a person entering a month, I want idle `p` and a `P` typed
into a cell to stay text, so that printing does not take the
keys I use to write.

### US3 — Print only from this sheet

As a person on the list, the form, or a chart, I want `P` to
leave the log sheet unwritten, so that only an idle month sheet
prints it, and chart `p` still prints only the chart.

## Decisions

1.  **The key is the letter P** on the month sheet when no cell is
    being edited. It is the shifted letter, not a second meaning
    for `p`. A chord that carries no letter is not this key.
    Today that idle letter starts an edit on a text cell. This
    slice is the change: idle P exports, and it does not start
    an edit.
2.  **The file is the one the command already writes** for the
    month on screen. The key calls that writer. It does not add
    a painter. Page geometry, note cutting, and delta arithmetic
    stay in the 2026-10-17 spec.
3.  **The directory is TUI chart `p` with an empty `-o`.**
    `pdf_dir` when it is set, otherwise the XDG data directory.
    The name is `YYYY-MM-log.pdf`. Mode `0600`. A second press
    overwrites that one name. There is no working-directory
    fallback, no path prompt, and no second destination. `-o`
    stays a command flag. The database path does not move the
    file.
4.  **The month is the one on screen**, including after `[` and
    `]`. The key does not change day, column, or month. The
    screen stays the month sheet, not editing.
5.  **Success.** The status line contains the directory and
    `YYYY-MM-log.pdf`. A `~/` directory is shown expanded, not
    as a tilde. The month sheet's error line is clear. The
    process does not quit. Stored logs are unchanged.
6.  **Failure.** An error is shown on the month sheet. The
    status line is cleared, including a sticky `saved` left
    after a reload has already been applied, so the file is
    not claimed as saved. The process does not quit. Nothing
    is written in the working directory. The sheet stays on
    the same cell. The error text starts with `write log pdf`
    and then the cause. The cause is the one chart `p`
    already shows for that directory problem: `pdf_dir must
    be a string`, `pdf_dir` and `must be absolute`, `home
    directory`, or `not absolute`. The prefix is not `write
    chart pdf`.
7.  **The page is the command's page.** The key reads the
    database and trends that series the way `-log-pdf` does.
    A leading run with no weigh-in stays on the page, and the
    trend starts at the first weigh-in. A series with no
    weigh-in is still written. The key does not change
    `ApplyTrend`, and it does not pass the series the sheet
    has in memory. The rows on screen stay as they are,
    including after a reload that was rejected. Success still
    clears the error line and sets the path. The drawn text
    matches the command for that database, that month, and
    the display unit, with an empty `-o`. The files need not
    be byte-identical.
8.  **Idle `p` is unchanged.** On a text cell it still starts
    an edit whose text is `p` and it writes no PDF. On the
    workout column it still does not start an edit and still
    writes no PDF.
9.  **While a cell is being edited, P is text** and writes no
    PDF. That includes the steps cell beside workout, and the
    note. Editing continues on that cell.
10. **Idle P on the workout column exports.** That column is
    not a text edit, so the letter is the key, not a typed
    character.
11. **Other screens do not write a log PDF.** The list, the
    form, the monthly chart, and the long-term chart stay
    where they are. On the form, P is typed text where the
    field already accepts text, and it is not text on the
    workout field. `P` on the monthly chart does not write
    the chart PDF either. Chart `p` still writes only
    `YYYY-MM-chart.pdf`.
12. **The month sheet help line names the key** and keeps
    every chord it has now. The line is:

    ```text
    arrows move   type edit   tab next   space workout   enter form   c chart   l long   P log   t today   [ ] month   esc list
    ```

13. **The key reference and the log-sheet how-to gain this
    step.** The chart how-to stays the chart page.
14. **A save or reload still in flight is not a page.**
    Shift-P writes only when no month-cell save and no reload
    of that save is in flight. While the status is `saved`
    and that reload has not been applied, the key writes no
    PDF and does not clear `saved`. After this key has set
    the PDF path, a save result that was already in flight
    does not replace that path with `saved`. The reload may
    still update the rows. A sticky `saved` left after the
    reload has been applied is not in flight. Decision 6
    still clears that leftover on a failure.

## Acceptance scenarios

November 1990 unless the scenario moves the month. The filename
for that month is `1990-11-log.pdf`. The working directory is
not the data directory.

### S1 — Idle shift-P writes the sheet and stays

**Given** the month sheet for November 1990, not editing, on a
text cell
**And** no `pdf_dir`
**And** the data directory is the XDG data directory
**And** the database file is not in that directory
**When** shift-P is pressed
**Then** `1990-11-log.pdf` exists in that data directory, mode
`0600`
**And** the status line contains that directory and
`1990-11-log.pdf`
**And** the error line is clear
**And** the screen is still the month sheet, on the same day
and column, not editing
**And** there was no path prompt
**And** the working directory has no `1990-11-log.pdf` and no
`1990-11-chart.pdf`
**And** the data directory has no `1990-11-chart.pdf`
**And** the stored logs are unchanged
**And** the process does not quit

### S2 — `pdf_dir` is the directory

**Given** the month sheet for November 1990, not editing
**And** `pdf_dir` is an absolute directory
**When** shift-P is pressed
**Then** `1990-11-log.pdf` exists only in that directory, mode
`0600`
**And** the status line contains that directory and
`1990-11-log.pdf`
**And** the data directory and the working directory have no
`1990-11-log.pdf`
**And** the screen is still that month sheet, on the same cell

### S3 — A leading `~/` is shown expanded

**Given** the month sheet for November 1990, not editing
**And** `pdf_dir` begins with `~/`
**When** shift-P is pressed
**Then** the file is in the directory that prefix expands to,
mode `0600`
**And** the status line contains that expanded directory and
`1990-11-log.pdf`
**And** the status line does not contain `~`
**And** the working directory has no `1990-11-log.pdf`

### S4 — `]` writes the next month

**Given** the month sheet for November 1990, not editing, on
a chosen day and column
**When** `]` is pressed and then shift-P
**Then** `1990-12-log.pdf` exists in the PDF directory
**And** `1990-11-log.pdf` does not
**And** the screen is still the month sheet for December 1990,
on that same day and column, not editing
**And** the status line contains `1990-12-log.pdf`

### S5 — `[` writes the previous month

**Given** the month sheet for November 1990, not editing, on
a chosen day and column
**When** `[` is pressed and then shift-P
**Then** `1990-10-log.pdf` exists in the PDF directory
**And** `1990-11-log.pdf` does not
**And** the screen is still the month sheet for October 1990,
on that same day and column, not editing
**And** the status line contains `1990-10-log.pdf`

### S6 — Shift-P overwrites that one name

**Given** the month sheet for November 1990, not editing
**And** the destination path already holds a regular file
that is not a PDF
**When** shift-P is pressed
**Then** that path is one file, mode `0600`, and the bytes
begin with `%PDF`
**And** the directory has no second log PDF, no backup, and
no `1990-11-chart.pdf`
**And** the working directory has no `1990-11-log.pdf`

### S7 — The page is the command's page

**Given** the database holds a November 1990 log in pounds,
including a leading day with no weigh-in
**And** the month sheet is showing an older series than that
database
**When** shift-P writes `1990-11-log.pdf`
**Then** the drawn text is the drawn text the command
`-log-pdf 1990-11` writes for that database and display unit
with an empty `-o`
**And** the two files need not be byte-identical
**And** the rows on screen are unchanged
**And** this slice does not describe the columns, the note
cut, or the delta

### S8 — A write failure is shown and does not quit

**Given** the month sheet for November 1990, not editing, on
a chosen cell
**And** the reload has been applied
**And** the status line still says `saved`
**And** the destination path exists as a directory
**When** shift-P is pressed
**Then** the screen shows an error
**And** the error contains `write log pdf`
**And** the error does not contain `write chart pdf`
**And** the status line is clear, so it does not say `saved`
and it is not the PDF path
**And** that destination is still a directory
**And** the working directory has no `1990-11-log.pdf`
**And** the screen is still the month sheet, on the same
cell, not editing
**And** the process does not quit

### S9 — An unusable directory writes nothing

Each cause below is its own run. The screen starts as the
November 1990 month sheet, not editing, on a chosen cell.
Shift-P does not quit, stays on that cell, writes no
`1990-11-log.pdf` in the working directory, and does not
present the path as a saved file.

**Given** `pdf_dir` is an existing file
**When** shift-P is pressed
**Then** an error is shown
**And** the error contains `write log pdf`
**And** the error does not contain `write chart pdf`
**And** that path is still a file
**And** the data directory has no `1990-11-log.pdf`

**Given** `pdf_dir` is not a string
**When** shift-P is pressed
**Then** the error contains `write log pdf` and `pdf_dir must
be a string`
**And** the error does not contain `write chart pdf`

**Given** `pdf_dir` is a relative path other than a leading
`~/`
**When** shift-P is pressed
**Then** the error contains `write log pdf`, `pdf_dir`, and
`must be absolute`
**And** the error does not contain `write chart pdf`

**Given** no `pdf_dir`, and the home directory is unavailable
**When** shift-P is pressed
**Then** the error contains `write log pdf` and `home
directory`
**And** the error does not contain `write chart pdf`

**Given** no `pdf_dir`, and the data directory is not absolute
**When** shift-P is pressed
**Then** the error contains `write log pdf` and `not absolute`
**And** the error does not contain `write chart pdf`
**And** no `1990-11-log.pdf` is written under that relative
data path

### S10 — Idle `p` still edits

**Given** the month sheet, not editing, on a text cell
**When** `p` is pressed
**Then** an edit starts whose text contains `p`
**And** neither PDF is written in the working directory, the
data directory, or `pdf_dir`

**Given** the month sheet, not editing, on the workout column
**When** `p` is pressed
**Then** an edit does not start
**And** neither PDF is written

### S11 — Chart `p` still writes only the chart

**Given** the monthly chart for November 1990
**When** `p` is pressed
**Then** `1990-11-chart.pdf` exists in the PDF directory,
mode `0600`
**And** `1990-11-log.pdf` does not exist in that directory,
the data directory, or the working directory
**And** the screen is still the monthly chart

### S12 — Shift-P while editing is text

**Given** the month sheet, editing the weight cell
**When** shift-P is pressed
**Then** the field text contains `P`
**And** that same cell is still being edited
**And** no log PDF is written

**Given** the month sheet, editing the steps cell beside
workout
**When** shift-P is pressed
**Then** the field text contains `P`, that cell is still
being edited, and no log PDF is written

**Given** the month sheet, editing the note
**When** shift-P is pressed
**Then** the field text contains `P`, that cell is still
being edited, and no log PDF is written

### S13 — Idle shift-P on workout still writes

**Given** the month sheet, not editing, on the workout
column
**When** shift-P is pressed
**Then** the log PDF is written as in S1
**And** an edit does not start
**And** the column is still workout

### S14 — `P` on the list writes nothing

**Given** the daily list
**When** shift-P is pressed
**Then** no log PDF and no chart PDF is written
**And** the screen is still the list

### S15 — `P` on the form note is text

**Given** the day form, on the note field
**When** shift-P is pressed
**Then** the note contains `P`
**And** no log PDF is written
**And** the screen is still the form

### S16 — `P` on the form workout is not text

**Given** the day form, on the workout field
**When** shift-P is pressed
**Then** the workout value is unchanged
**And** the field does not gain the letter `P`
**And** no log PDF is written
**And** the screen is still the form

### S17 — `P` on the monthly chart writes neither PDF

**Given** the monthly chart for November 1990
**When** shift-P is pressed
**Then** no `1990-11-log.pdf` and no `1990-11-chart.pdf` is
written
**And** the screen is still the monthly chart

### S18 — `P` on the long-term chart writes nothing

**Given** the long-term chart
**When** shift-P is pressed
**Then** no log PDF and no chart PDF is written
**And** the screen is still the long-term chart

### S19 — The help line names the key

**Given** the month sheet
**When** the sheet is shown
**Then** the help line is the line in Decision 12

### S20 — The reference and the how-to name the key

**Given** the key reference and the filled-log how-to
**When** this slice is documented
**Then** the key reference has a row for `P` on the month
sheet when not editing, writing `YYYY-MM-log.pdf` in
`pdf_dir` or the data directory
**And** the `p` row is still only the monthly chart
**And** the reference says idle `p` on the month sheet still
edits, and that `P` while editing is text
**And** the how-to tells the reader to press shift-P on the
idle month sheet, including after `[` and `]`
**And** the how-to no longer says there is no TUI key
**And** the chart how-to is unchanged

### S21 — Shift-P waits out the reload

**Given** a month-cell save has set the status to `saved`
**And** the reload from that save has not been applied
**When** shift-P is pressed
**Then** no log PDF is written
**And** the status line still says `saved`
**And** the screen is still the month sheet, on the same
cell

### S22 — A late save does not cover the path

**Given** shift-P has set the status line to the log PDF
path
**And** a month-cell save that was already in flight then
completes
**When** that save result is applied
**Then** the status line is still that PDF path
**And** the status line does not say `saved`

## Functional requirements

-   **FR1.** Idle shift-P on the month sheet writes
    `YYYY-MM-log.pdf` for the month on screen into the XDG
    data directory when `pdf_dir` is omitted, mode `0600`.
    The status line contains that directory and the filename.
    The error line is clear. The screen stays the month sheet
    on the same cell, not editing. There is no path prompt.
    The working directory gains neither PDF. The chart PDF is
    not written. Stored logs are unchanged. The process does
    not quit. The database path does not move the file.
-   **FR2.** An absolute `pdf_dir`, and a `pdf_dir` that
    begins with `~/`, are the directory chart `p` already
    uses. The status line shows the expanded directory and
    `YYYY-MM-log.pdf`, not a tilde. The data directory and
    the working directory do not gain the file.
-   **FR3.** After `]` the file is the next month's
    `YYYY-MM-log.pdf`. After `[` it is the previous month's.
    The day and column stay where those keys left them.
-   **FR4.** A second press, or a press onto a regular file
    that is not a PDF, leaves one file of that name, mode
    `0600`, beginning with `%PDF`. No backup and no chart PDF.
-   **FR5.** The drawn text matches the command
    `-log-pdf YYYY-MM` for the database and the display unit,
    with an empty `-o`, including a leading day with no
    weigh-in. The files need not be byte-identical. The key
    reads the database and trends it the way that command
    does. It does not pass the series on screen, and the rows
    on screen do not change. The key adds no painter and does
    not change `ApplyTrend` or the page rules.
-   **FR6.** When the destination path is a directory, or the
    directory cannot be used, an error is shown, the status
    line is clear (a sticky `saved` left after a finished
    reload is gone), the process does not quit, the working
    directory gains no PDF, and the sheet stays on the same
    cell. The error starts with `write log pdf` and does not
    contain `write chart pdf`. A `pdf_dir` that is a file
    stays a file. A non-string `pdf_dir` says `pdf_dir must
    be a string`. Another relative `pdf_dir` says `pdf_dir`
    and `must be absolute`. No usable home says `home
    directory`. A data directory that is not absolute says
    `not absolute` and writes nothing under that relative
    path.
-   **FR7.** Idle `p` on a text cell still starts an edit
    whose text contains `p` and writes no PDF. Idle `p` on
    the workout column still does not start an edit and
    writes no PDF.
-   **FR8.** Chart `p` still writes only `YYYY-MM-chart.pdf`
    in the same directory, mode `0600`, and does not write
    `YYYY-MM-log.pdf`. The screen stays the monthly chart.
-   **FR9.** While weight, steps, or the note is being
    edited, shift-P inserts `P`, leaves that edit open, and
    writes no PDF.
-   **FR10.** Idle shift-P on the workout column writes the
    log PDF and does not start an edit. The column stays
    workout.
-   **FR11.** Shift-P on the list, the form, the monthly
    chart, and the long-term chart writes no log PDF and
    does not change screen. The form note gains `P`. The
    form workout value does not change and does not gain
    `P`. The monthly chart does not write the chart PDF for
    this key.
-   **FR12.** The month sheet help line is the line in
    Decision 12.
-   **FR13.** `docs/reference/keys.md` and
    `docs/how-to/monthly-log-pdf.md` gain the step in S20.
    The chart how-to is unchanged.
-   **FR14.** While a month-cell save or its reload is in
    flight, shift-P writes no PDF. When the status is `saved`
    and that reload has not been applied, the status stays
    `saved`. A save result that arrives after the PDF path is
    set does not replace that path with `saved`. A sticky
    `saved` after the reload has been applied is not this
    hold.

## Out of scope

-   Blank log sheets
-   Long-term chart PDFs
-   A second PDF directory, a new config key, or a path
    prompt
-   A Unicode font, or any change to which characters the
    page can draw
-   Issues
    [#114](https://github.com/jglueckstein/hdtools/issues/114)
    and
    [#115](https://github.com/jglueckstein/hdtools/issues/115)
-   Changing idle `p`, including on the workout column
-   Changing chart `p`, the chart painter, or the chart page
-   Changing on-screen columns, `ApplyTrend`, or the
    long-term chart
-   Importing a second PDF painter, or restating page
    geometry, note cutting, or delta arithmetic
-   Using `-o` from the TUI
-   Clearing the sticky `saved` status on keys other than
    this export
-   Amending
    [`2026-10-17-filled-log-sheet-pdf.md`](2026-10-17-filled-log-sheet-pdf.md)

## Documentation

Add a `P` row to `docs/reference/keys.md` for the month sheet
when not editing. Keep the `p` row as the monthly chart.
State that idle `p` on the month sheet still edits, and that
`P` while editing is text.

Add a month-sheet step to `docs/how-to/monthly-log-pdf.md`:
shift-P on the idle sheet, the month on screen including
after `[` and `]`, the same directory as the command, no
path prompt, and `-o` only on the command. Remove the
sentence that there is no TUI key. Do not edit
`docs/how-to/monthly-chart-pdf.md`.

Those edits land with the implementation, not with the
failing tests. Blank sheets stay the earlier backlog
sentence under Monthly Log.
