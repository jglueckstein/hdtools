# Plan — Month-sheet key for the filled log PDF

**Spec**:
[`docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`](../specs/2026-10-28-month-sheet-log-pdf.md)
**Slice**: S2 of
[`filled-log-sheet-pdf.md`](../slices/filled-log-sheet-pdf.md)
**Issue**: [#113](https://github.com/jglueckstein/hdtools/issues/113)
**Follows**:
[`2026-10-17-filled-log-sheet-pdf.md`](../plans/2026-10-17-filled-log-sheet-pdf.md)
(approved). Do not edit that plan or its spec.
**Status**: approved

No production code until the spec's scenarios exist as failing
tests.

S2 is only the month-sheet key. The page, the filename, and the
directory rule are shipped. The key loads and trends the way
`-log-pdf` does, then calls the existing writer. Do not restate
geometry, note cutting, or delta arithmetic in tests. Accepted
objections O1, O2, and O3 in
`docs/superpowers/objections/month-sheet-log-pdf.md` are already
folded into the spec.

The `idea.md` change-record link is already in place. Do not
rewrite the historical sentences. Blank sheets stay the earlier
backlog sentence.

## Module structure

| File | Why |
| --- | --- |
| `internal/dailylog` | Move the two helpers `exportLogPDF` already uses, the order check and trend-from-first-weigh-in, so the command and the key call one function each. `ApplyTrend` does not change. |
| `cmd/hdtools/main.go` | Wiring only. Call those helpers. The command's page does not change. |
| `internal/tui/app.go` | One idle-branch case in `updateMonth`, before the single-rune edit, and a call from `savedMsg` so a late save does not cover the PDF path. `app.go` is 493 lines. Each site is a call and a return. Do not add the write or the load, and do not add an import, or the file passes 500 lines. Do not handle `P` in the editing branch. |
| `internal/tui/logsheet.go` | New file. One concern: if no save or reload is in flight, load the store, trend with the shared helpers, choose the directory, and call the existing writer. Literate preamble: the chart key cannot be this key, this file does not paint, and it does not import a PDF library. |
| `internal/tui/month.go` | The help line only. Do not change columns, carry, or cell editing. |
| `internal/tui/logsheet_test.go` | New tests. Same package as the chart helpers. |
| `docs/reference/keys.md` | FR13. A `P` row. The `p` row stays the chart. |
| `docs/how-to/monthly-log-pdf.md` | FR13. A month-sheet step. Remove the sentence that there is no TUI key. |
| `CHANGELOG.md` | When the behaviour ships, one Unreleased bullet. Not in the failing-test step. |

Do not edit `internal/logpdf`, `internal/chartpdf`,
`internal/tui/chart.go`, `internal/tui/longchart.go`,
`internal/tui/form.go`, or `docs/how-to/monthly-chart-pdf.md`.
Do not edit the slicing record. Do not import
`github.com/go-pdf/fpdf` from the TUI. The TUI already uses the
store. `internal/logpdf` does not import the TUI, so the TUI
may import `logpdf` the way `chart.go` imports `chartpdf`.

The existing export path is `exportLogPDF` in
`cmd/hdtools/main.go`: load `store.All`, reject a duplicate or
a backwards series, trend from the first weigh-in, and on an
empty `-o` join `YYYY-MM-log.pdf` onto `config.ResolvePDFDir`,
then `logpdf.Write`. The key uses that preparation, that
resolver, and that `Write`. It does not call `exportLogPDF`,
and it does not grow a second painter.

## Algorithm notes

-   The message is the letter P, built the way `press` builds a
    rune and the way chart `p` is built. `msg.String()` is `P`.
    Match that string in the idle branch only. Do not match
    `shift+p`. Do not require the shift modifier bit. An empty
    text with a shift modifier is not this key; leave it on
    the path that already ignores it.
-   Idle `p` stays the default branch: a text cell calls
    `beginEdit` with `p`. The workout column still skips that
    branch. Do not add a `p` case.
-   While `month.editing` is set, `P` falls through to
    `textinput.Update` with the other typed letters. Do not
    export from that branch. Steps is `colSteps`. The note is
    `colNote`. Both are text. Workout is `colWorkout` and is
    never editing; idle `P` there is the export case.
-   The write is synchronous inside the key, then
    `return a, nil`, same as `writeChartPDF`. Success and
    failure do not return `tea.Quit`.
-   Directory and name follow `writeChartPDF` with a different
    suffix: `config.ResolvePDFDir`, then
    `%04d-%02d-log.pdf` joined onto that directory. Do not
    call `os.Getwd`. Do not trim or interpret `-o`. Mode
    `0600` and overwrite stay inside `logpdf.Write`
    (`installOwnerOnly`). Do not `os.Create` the destination
    from the TUI. Do not assert the created directory's mode.
-   `logpdf.Write` receives the month on screen
    (`month.year`, `month.month`), the series from the shared
    helpers, `cfg.DisplayUnit`, and `localToday()`. `Today` is
    ignored by the painter; pass it so the call matches the
    sheet's clock. Do not pass `a.logs`. Do not call
    `ApplyTrend` on the whole series. An empty series and a
    series with no weigh-in are passed through. A leading run
    with no weigh-in stays untrended, and the trend starts at
    the first weigh-in. `ErrNoTrendStart` is not an export
    failure. Do not replace the rows on screen.
-   The month view paints `month.err` and the status
    argument. It does not paint `a.err`. A failure stored
    only on `a.err` is invisible, and a leftover `a.err`
    replaces the list after Esc. Put the error string on
    `month.err`, set `a.err` to nil, and set `a.status` to
    `""`. Success clears `month.err` and `a.err`, and sets
    `a.status` to the full path `Write` received. That path
    is absolute after `ResolvePDFDir`, so a `~/` value does
    not remain in the status line.
-   Wrap the resolver error and the write error with
    `fmt.Errorf("write log pdf: %w", err)` before storing
    `Error()` on `month.err`. The inner text stays visible
    (`pdf_dir must be a string`, `must be absolute`,
    `home directory`, `not absolute`). The tests require the
    prefix `write log pdf` and reject `write chart pdf`. Do
    not `return err`.
-   A month-cell save command marks the save in flight until
    `savedMsg`, `loadErrMsg`, or the dropped-save message.
    `savedMsg` marks the reload in flight until `loadedMsg`
    or `loadErrMsg`. Shift-P returns immediately, and does
    not change status, while either mark is set. The string
    `saved` alone is not the mark: a sticky `saved` after the
    reload has finished is Decision 6, and a failed write
    clears it. Other keys stay as they are. Code-mode O2 on
    the edit-keys record stays deferred.
-   `savedMsg` must not set status to `saved` when this key
    has already set the PDF path. The reload still runs. That
    guard is a function in `logsheet.go` so `app.go` stays at
    or under 500 lines.
-   Chart `updateChart` still matches only `p`. Do not add
    `P` there. Long-term keys and the form are unchanged.
-   Help text is the Decision 12 line, still one
    `p.help.Render` string in `month.view`. Assert it with
    `visible`, which strips the colour.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestMonthShiftPWritesLogPDF` |
| FR2 | `TestMonthShiftPUsesPDFDir`, `TestMonthShiftPHomePrefix` |
| FR3 | `TestMonthShiftPAfterNextMonth`, `TestMonthShiftPAfterPrevMonth` |
| FR4 | `TestMonthShiftPOverwrite` |
| FR5 | `TestMonthShiftPMatchesCommand` |
| FR6 | `TestMonthShiftPWriteFailure`, `TestMonthShiftPPDFDirIsFile`, `TestMonthShiftPNonStringPDFDir`, `TestMonthShiftPRelativeDirFails`, `TestMonthShiftPNoHomeFails`, `TestMonthShiftPRelativeDataHomeFails` |
| FR7 | `TestMonthPStartsEditWritesNoPDF` |
| FR8 | `TestChartPWritesChartNotLogPDF` |
| FR9 | `TestMonthShiftPWhileEditingIsText` |
| FR10 | `TestMonthShiftPOnWorkoutWrites` |
| FR11 | `TestMonthShiftPIgnoredOnList`, `TestMonthShiftPOnFormNoteIsText`, `TestMonthShiftPOnFormWorkoutWritesNothing`, `TestMonthShiftPIgnoredOnChart`, `TestMonthShiftPIgnoredOnLong` |
| FR12 | `TestMonthHelpNamesLogKey` |
| FR13 | `docs/reference/keys.md`, `docs/how-to/monthly-log-pdf.md` (no Go test) |
| FR14 | `TestMonthShiftPDuringReload`, `TestMonthShiftPSaveDoesNotCoverPath` |

## Test list

November 1990 via `twoDayApp` unless the test moves the month.
Shift-P is `tea.KeyPressMsg{Code: 'P', Text: "P"}`. Chart and
idle `p` are `tea.KeyPressMsg{Code: 'p', Text: "p"}`, the body
of `applyP`. `[` and `]` are the `press` messages
`tea.KeyPressMsg{Code: '[', Text: "["}` and
`tea.KeyPressMsg{Code: ']', Text: "]"}`. Do not use `tea.KeyMsg`.
Mirror `applyP`: run a returned command and fail on
`tea.QuitMsg`. `TestMonthShiftPDuringReload` does not. It
applies the save result, leaves the load command unrun, and
then presses P.

`t.Chdir` to a temp working directory. `useDataHome`,
`assertNoPDF`, `assertPDFNotCwd`, `appWithConfig`,
`configWithPDFDir`, and `visible` already exist. No
`t.Parallel` on these tests. After `press(app, "m")`, set
`month.day` and `month.col` before the key so the cell does
not depend on the clock. Filename helpers:
`1990-11-log.pdf`, and `1990-12-log.pdf` / `1990-10-log.pdf`
after the brackets. Do not open the real home.

On current master the export tests and
`TestMonthHelpNamesLogKey` fail because idle `P` starts an
edit (or, on workout, does nothing) and the help line has no
`P log`. The locks that already pass — idle `p`, chart `p`,
`P` while editing, and `P` on the list, the form, and both
charts — stay in this list so the key cannot take those
paths. Do not change production code to make a lock fail.
Keep `TestChartPIgnoredOnList`, `TestChartPWhileEditingIsText`,
`TestChartPIgnoredOnForm`, and `TestChartPWritesPDF` green
and do not retarget them.

Docs and `CHANGELOG.md` are not part of the failing-test
step.

-   `TestMonthShiftPWritesLogPDF` — S1: idle P on weight
    writes `1990-11-log.pdf` under the data directory, mode
    `0600`, status contains that directory and name, error
    line clear, same cell, not editing, no cwd file, no chart
    PDF, row count unchanged, command is not quit.
-   `TestMonthShiftPUsesPDFDir` — S2: absolute `pdf_dir`,
    file only there, status contains that directory and the
    name, still the same month cell.
-   `TestMonthShiftPHomePrefix` — S3: `pdf_dir` is `~/charts`
    under the temp home from `useDataHome`; status is the
    expanded directory and the name and has no `~`; mode
    `0600`; no cwd file.
-   `TestMonthShiftPAfterNextMonth` — S4: `]` then P writes
    `1990-12-log.pdf` only, December sheet, same day and
    column, not editing.
-   `TestMonthShiftPAfterPrevMonth` — S5: `[` then P writes
    `1990-10-log.pdf` only, October sheet, same day and
    column, not editing.
-   `TestMonthShiftPOverwrite` — S6: destination is a regular
    non-PDF file; after P it is one file, mode `0600`, bytes
    start with `%PDF`, no backup and no chart PDF.
-   `TestMonthShiftPMatchesCommand` — S7: display unit `lb`;
    store a leading day with no weigh-in ahead of the
    weighed days; leave the sheet showing the older series.
    The command's file and the key's file have the same page
    content streams. Ignore `/CreationDate` and `/ModDate`.
    Do not compare raw bytes. Do not build the oracle from
    `a.logs`. The sheet's rows are unchanged. Do not assert a
    media box, a column width, or a delta.
-   `TestMonthShiftPWriteFailure` — S8: the log path is a
    directory, the reload has been applied, and status is
    the sticky `saved`; after P the view shows an error that
    contains `write log pdf` and not `write chart pdf`,
    status is empty, the directory remains, no cwd file,
    same cell, command is not quit.
-   `TestMonthShiftPPDFDirIsFile` — S9: `pdf_dir` is a file;
    error contains `write log pdf` and not `write chart
    pdf`, status empty, that path stays a file, no log PDF
    in the cwd or the data directory, same cell, not quit.
-   `TestMonthShiftPNonStringPDFDir` — S9: `pdf_dir = 3`;
    error contains `write log pdf` and `pdf_dir must be a
    string`, and not `write chart pdf`; status empty; no
    log PDF.
-   `TestMonthShiftPRelativeDirFails` — S9: `pdf_dir` is
    `charts`; error contains `write log pdf`, `pdf_dir`, and
    `must be absolute`, and not `write chart pdf`; status
    empty; no cwd log PDF.
-   `TestMonthShiftPNoHomeFails` — S9: empty `HOME` and
    `XDG_DATA_HOME`, database path set to a temp file; error
    contains `write log pdf` and `home directory`, and not
    `write chart pdf`; status empty; no cwd log PDF.
-   `TestMonthShiftPRelativeDataHomeFails` — S9:
    `XDG_DATA_HOME` is `data`; error contains `write log
    pdf` and `not absolute`, and not `write chart pdf`;
    status empty; no log PDF in the cwd or under that
    relative path.
-   `TestMonthShiftPDuringReload` — S21: apply the save
    result so status is `saved`, do not run the load
    command, then press P. No log PDF. Status is still
    `saved`. Same cell.
-   `TestMonthShiftPSaveDoesNotCoverPath` — S22: after P has
    set the PDF path, apply a `savedMsg` from a save that
    was already in flight. Status is still that path and
    does not say `saved`.
-   `TestMonthPStartsEditWritesNoPDF` — S10: P's lowercase
    message on weight starts an edit whose text contains
    `p`, and on `colWorkout` it does not start an edit;
    neither PDF in the cwd, the data directory, or `pdf_dir`.
-   `TestChartPWritesChartNotLogPDF` — S11: on the monthly
    chart, the lowercase `p` message writes
    `1990-11-chart.pdf` only, mode `0600`, and no
    `1990-11-log.pdf`; the screen stays the chart.
-   `TestMonthShiftPWhileEditingIsText` — S12: `beginEdit`
    on weight, on `colSteps`, and on `colNote`; P leaves
    `P` in the field, that edit open, and no log PDF.
-   `TestMonthShiftPOnWorkoutWrites` — S13: idle P on
    `colWorkout` writes the log PDF as in S1, `editing`
    stays false, column stays workout.
-   `TestMonthShiftPIgnoredOnList` — S14: P on the list
    writes neither PDF; screen stays the list.
-   `TestMonthShiftPOnFormNoteIsText` — S15: the note,
    reached as in `TestChartPIgnoredOnForm`, gains `P`; no
    log PDF; screen stays the form.
-   `TestMonthShiftPOnFormWorkoutWritesNothing` — S16: four
    Tabs from the date field, then P; yes/no unchanged, no
    `P` inserted, no log PDF, screen stays the form.
-   `TestMonthShiftPIgnoredOnChart` — S17: P on the monthly
    chart writes neither PDF and stays on the chart.
-   `TestMonthShiftPIgnoredOnLong` — S18: P on the long-term
    chart writes neither PDF and stays on that chart.
-   `TestMonthHelpNamesLogKey` — S19: `visible` of the month
    sheet help line is the Decision 12 string.
