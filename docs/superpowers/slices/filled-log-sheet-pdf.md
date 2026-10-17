---
task: >
  Generate a PDF of one month's filled log sheet. The on-screen
  month sheet already exists. Monthly chart PDFs already exist
  (CLI -chart-pdf YYYY-MM and TUI p). This task is the filled log
  sheet on paper, not the chart. Filled log-sheet PDFs are separate
  from blank sheets and from monthly chart PDFs. Blank sheets and
  long-term chart PDFs stay out. The file uses the existing PDF
  directory: $XDG_DATA_HOME/hdtools (default ~/.local/share/hdtools),
  config pdf_dir, and CLI -o overrides the path. Do not invent a
  second destination.
task_slug: filled-log-sheet-pdf
date: 2026-10-16
carpaccio_model: grok-4.7
inseparable: false
progressed_slice: S1
slices:
  - id: S1
    title: Write one portrait page of the filled month with -log-pdf
    scope: >
      Add hdtools -log-pdf YYYY-MM. It writes one portrait US Letter
      page (612 by 792 points) of that month's log sheet and does not
      start the TUI. The page is a ruled vector table, not a
      screenshot of the terminal and not the chart. The heading names
      the month, the year, and the display unit, for example
      "November 1990 (lb)". It does not print the database path, the
      key help, Monthly Loss, or Daily Deficit. Columns, in the month
      sheet's order, are day, weekday, weight, trend, delta, sleep,
      steps, workout, and note. One row per calendar day, one line
      high, including days with no database row and days after today
      when the month is the current one. An empty month is still
      those rows, not the chart's "(empty month)" page. The full
      series is loaded and ApplyTrend runs before the month is built,
      so a blank day shows the carried trend the month sheet shows.
      An empty log skips ApplyTrend and is a sheet of blank days.
      This slice does not use the chart clip. Weight, trend, and
      delta are the display unit to one decimal. Delta is the weight
      cell minus the trend cell after that rounding, not a kilogram
      gap converted afterwards. A positive delta has a leading plus.
      An exact zero prints 0.0, not -0.0. Delta is blank when the day
      has no weigh-in or no trend. Stone is the single decimal the
      TUI already shows. Kilograms stay in the database. A day with a
      row prints sleep to one decimal, steps as an integer, and
      workout as yes or no, including zeros. A day with no row leaves
      those cells blank. Blank cells use a dash the font can draw.
      The note is the stored note, cut with an ellipsis when it does
      not fit the line. A note that fits is not cut. The page count
      stays one. The font is the chart's core Helvetica. A character
      it cannot draw is skipped, and the export still writes the
      page. There is no new font file. The file name is
      YYYY-MM-log.pdf. The directory is the chart rule: ResolvePDFDir,
      pdf_dir, and -o as a whole file path. -o is not joined to the
      directory. Success prints that path as one line on standard
      output. Failure prints no path, writes nothing in the working
      directory, and does not fall back to another directory. The
      file is mode 0600 via a same-directory temp renamed into place.
      A second run overwrites that one name and does not write
      YYYY-MM-chart.pdf. A bad YYYY-MM fails the same way -chart-pdf
      does. If -log-pdf and -chart-pdf are both set, the process
      exits non-zero, writes nothing, and does not start the TUI.
      -o alone, with neither export flag, still only starts the TUI.
      Absent -log-pdf, -chart-pdf is unchanged. The writer is a new
      package under internal/. internal/chartpdf's page is not
      edited. cmd/hdtools only wires the flag. The TUI does not gain
      a key in this slice and does not import a PDF library. The CLI
      reference, the -o line, and a one-line clause on pdf_dir
      describe the new file. A how-to at
      docs/how-to/monthly-log-pdf.md says how to run the command.
      The chart how-to stays the chart page.
    decision_focus: >
      Does -log-pdf YYYY-MM write one portrait US Letter page of
      every calendar day in that month — the month sheet's columns,
      in the display unit, including blank days, days after today,
      and carried trend — as YYYY-MM-log.pdf in the existing PDF
      directory, with a note that does not fit the row cut short,
      and with the chart PDF left untouched?
    lens_used: decision-boundary
    sequencing_note: >
      First. S2 needs this writer, this filename, and this directory
      rule. The chart export does not.
    disposition: accepted
    disposition_rationale: >
      Accepted. The page is the month sheet on paper, and the
      command is enough to check it. A dump of only stored rows, a
      clip at today, a second page, landscape, or kilograms
      regardless of display_unit would be a different sheet. This
      is the slice in progress. Issue #110 already covers it.
    file_as_issue: false
    issue_url: null
    merged_into: null
  - id: S2
    title: Export that page from the month sheet with shift-p
    scope: >
      On the month sheet, when no cell is being edited, shift-p (P)
      writes the S1 file for the month on screen. The directory rule
      is the same as TUI chart p with an empty -o: pdf_dir or the
      data directory, filename YYYY-MM-log.pdf, mode 0600, overwrite,
      no working-directory fallback. There is no path prompt and no
      second destination. Success: the status line contains the
      directory and YYYY-MM-log.pdf, and the screen stays the month
      sheet on the same cell. Failure: an error is shown, the process
      does not quit, the file is not claimed as saved, and no
      working-directory copy is written. Idle p on the month sheet
      still starts an edit whose text is p and writes no PDF. Chart
      p still writes only YYYY-MM-chart.pdf. While a cell is being
      edited, P is text in the field and writes no PDF, including on
      the workout column's neighbours and on the note. P on the list,
      the form, the monthly chart, and the long-term chart writes no
      PDF. On the form it stays typed text where the field already
      accepts text. The month sheet help line names the key. The key
      reference and the log-sheet how-to from S1 gain this step. The
      on-screen columns, the chart painter, and the long-term chart
      do not change.
    decision_focus: >
      Does the month sheet gain shift-p as the export key for the
      file S1 writes, so that idle p still starts an edit and chart
      p still writes only the chart?
    lens_used: decision-boundary
    sequencing_note: >
      After S1. Dropping this slice leaves the command in place. It
      must not land before the writer and the filename exist.
    disposition: accepted
    disposition_rationale: >
      Accepted. The month sheet's main job is typing, so the key
      is a separate decision from the page. The chart already has
      a key on the screen you are looking at. The log sheet should
      too, after the file exists. Merging the key into S1 would
      make one review accept the table and the typing change
      together. This slice waits until S1 has landed.
    file_as_issue: true
    issue_url: null
    merged_into: null
---

This record slices the filled monthly log-sheet PDF from issue 110
into two pieces. The month sheet on screen is already the paper log:
every calendar day, the fixed columns, a carried trend, and a delta
that is not stored. The monthly chart PDF already has a painter, a
flag, and a key. These slices add the sheet as a second file in the
same PDF directory. They do not add a blank form or a long-term chart
PDF.

## S1 — Write one portrait page of the filled month with -log-pdf — decision-boundary

### Context

The month sheet is the paper log. It has one row for every calendar
day of the open month, whether or not SQLite has that day. Trend
comes from ApplyTrend over the full series, never from a column. A
day with no weigh-in still shows the trend carried forward. Delta is
the on-screen weight cell minus the on-screen trend cell, in the
display unit, and it is not stored. The columns are day, weekday,
weight, trend, delta, sleep, steps, workout, and note. Workout is a
boolean, painted yes or no when the day has a row.

The chart PDF is a different artefact. `-chart-pdf` and TUI `p` on
the monthly chart write `YYYY-MM-chart.pdf` through `ResolvePDFDir`,
`pdf_dir`, and `-o`. That page is landscape, clips the current month
at today, and can say `(empty month)`. This slice does not open that
painter.

What you can see when it lands: `hdtools -log-pdf 1990-11` exits
without the TUI, prints one path, and leaves a one-page portrait
sheet. The chart file is not created or replaced by that run.

### Decision content

The engagement is the page the command writes. Every calendar day is
on it, including a day with no row and a day after today when you
export the current month. Numbers follow `display_unit`. A note that
does not fit its line is cut, and the month stays one page. The file
is `YYYY-MM-log.pdf` in the directory the chart PDF already uses.

The alternatives are a dump of only the stored rows, the chart's clip
at today, a second page so a long note is never cut, landscape
Letter, or kilograms on paper no matter what `display_unit` says.
Each of those is a different sheet. Skipping a character core
Helvetica cannot draw is part of this page. Embedding a Unicode font
so every note character survives is the alternative, and this slice
does not take it.

### Dependencies

Nothing in the repo has to land first. The destination rule and the
month sheet's row rules already exist. S2 calls this writer. The
chart export does not, and it stays on its own flag.

### Rationale

The rows, the unit, and the one-page cut are one sheet. You read a
blank day, its carried trend, and its empty delta together. A column
slice, or a kilogram page that waits to learn pounds, would ship
paper you cannot hold next to the screen. The known delta trap is in
this slice on purpose: converting a kilogram residual disagrees with
the cells, and a zero slope must not print as `-0.0`.

The command is the edge that makes the sheet observable. A path on
stdout and a file that is not `YYYY-MM-chart.pdf` are the check. The
month-sheet key is a different engagement, because that screen turns
an unbound letter into an edit. Putting the key here would make one
review accept the page and a typing change. The destination failures
stay the chart's rule, so this slice does not reopen `pdf_dir`.

## S2 — Export that page from the month sheet with shift-p — decision-boundary

### Context

After S1, the same file can be written from the month you are looking
at. Chart `p` cannot be that key. On the monthly chart, `p` writes
`YYYY-MM-chart.pdf`. On an idle month sheet, `p` starts an edit whose
first character is `p` and writes nothing. While a cell is being
edited, `p` is text. The destination work locked that behaviour so a
later export would not quietly take the key.

Shift-p is the proposal because it does not take that `p`, and
because a capital letter is a rarer first character of a note than
`s`. The month on screen, including after `[` and `]`, is the month
that is written. The TUI still has no per-press path. `-o` stays a
command flag.

What you can see when it lands: from an idle month sheet, shift-p
leaves `YYYY-MM-log.pdf` in the usual directory and the status line
names that path. The screen is still the sheet. A `p` typed into a
cell still does not create a file.

### Decision content

The engagement is the key. Shift-p writes the S1 file only when the
month sheet is not editing. Idle `p` still starts an edit. Chart `p`
still writes only the chart.

The alternatives are a command with no TUI key, idle `p` as the log
export, or a lowercase letter such as `s`. No key leaves printing to
the command and leaves typing alone. Idle `p` would make one key mean
"the PDF of this screen," and it would undo the rule that `p` on the
month sheet starts an edit. A lowercase letter matches `c` and `l`,
and it takes the first character of an idle cell. Shift-p keeps `p`
for the chart and for typing, at the cost of a shifted key the other
commands do not use.

### Dependencies

S1 must land first. The key calls that writer and uses that filename.
It does not choose a new directory and it does not grow a second
painter. If this slice is dropped, the command from S1 still writes
the sheet.

### Rationale

The chart key sits on a screen that does not take typed text. The
month sheet's main job is typing. A letter bound there is a visible
change even when the PDF is right. Reviewing that change inside S1
would hide a wrong key behind a correct table, or a wrong table
behind a surprising key.

The slice is done when shift-p on an idle sheet writes the same file
the command writes, when shift-p inside a cell does not, and when
idle `p` on the sheet still writes nothing. The list, the form, and
both charts stay unable to write a log sheet.

## Sequencing recommendation

Land S1, then S2.

S1 is the file. You can export a month without a new key, and the
chart path stays easy to check because that slice does not touch it.
S2 is only the key on a screen where typing is the main job. It waits
until the filename and the writer exist, so the key cannot invent a
second path or a second page.

S2 can be merged into S1 if you would rather accept the page and the
key together. S2 can be dropped if the command is enough. The
recommended order keeps the typing change in its own engagement.

## Explicitly not slicing on

- One slice per file or package. `cmd/hdtools`, a new writer under
  `internal/`, `month.go`, and `internal/chartpdf` are not decision
  boundaries. S1 is the command's page. S2 is the key. The chart
  painter is not opened to grow a table.
- One slice per column, or one slice per display unit. Day, weekday,
  weight, trend, delta, sleep, steps, workout, and note are one
  sheet. Kilograms, pounds, and stone are that same page. Stone stays
  one decimal. Delta subtracts the printed cells.
- A second PDF directory, a new config key, or a fresh pass over
  `pdf_dir`, `~/`, relative paths, `-db`, mode 0600, or
  confirm-before-overwrite. Those rules already exist. This task
  only adds `YYYY-MM-log.pdf` beside the chart file.
- The chart page, the on-screen monthly chart, and the on-screen
  long-term chart. TUI `p` on the chart stays the chart PDF. The
  sheet does not take the chart clip, the title box, the stems, or
  the `(empty month)` line.
- Blank sheets, long-term chart PDFs, and a template reserved for a
  future blank form. Those are other sittings. This page follows the
  month sheet, not a form that does not exist yet.
- A colour slice, a `NO_COLOR` slice, or a `[colors]` role. The sheet
  is black ink. The plus and the minus are how you read delta. The
  chart PDF stays the coloured page it is.
- Embedding a Unicode font. Core Helvetica is the chart's font. A
  character it cannot draw is skipped inside S1. A font file is a
  different sitting.
- A docs-only slice, and writing the dated spec or editing `idea.md`
  in this record. The command how-to, the CLI reference, and the
  `pdf_dir` clause travel with S1. The key reference travels with
  S2. The spec comes after a slice is accepted.
- One invocation that writes both PDFs. Both flags set is a failed
  command inside S1, not its own slice. The filename token
  `YYYY-MM-log.pdf` is part of S1, not a naming slice.
- Extracting the month-sheet row builder or the delta formatter into
  a shared package. That is a code cut. The check is that the PDF
  rows match the sheet.
