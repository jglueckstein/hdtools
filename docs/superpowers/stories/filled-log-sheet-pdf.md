---
spec: docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md
date: 2026-10-27
mode: spec
cartographer_model: grok-4.7
stories:
  - id: 1
    lens: [forces, consequences]
    title: Portrait sheet, not chart landscape
    disposition: accepted
    disposition_rationale: "One portrait page, 612 by 792, with day 31 inside that box. The chart PDF stays landscape. This slice does not define a combined packet."
  - id: 2
    lens: [forces, alternatives]
    title: Whole month, including after today
    disposition: promoted
    disposition_rationale: "The filled log is every calendar day of the named month, including days after today, and a later blank day still shows the carried trend. The chart clip stays on the chart. Carry that split into AGENTS.md beside the gotcha that a chart of the current month ends at today."
  - id: 3
    lens: [defaults, consequences]
    title: Delta subtracts the painted cells
    disposition: accepted
    disposition_rationale: "The page uses the rule already in AGENTS.md: subtract the one-decimal texts the weight and trend cells show. It does not store a delta, and it does not print Monthly Loss or Daily Deficit."
  - id: 4
    lens: [alternatives, defaults]
    title: No mark for a missing cell
    disposition: accepted
    disposition_rationale: "A missing cell on this page is empty. A logged zero still prints as 0.0, 0, or no. The on-screen em dash stays on the screen."
  - id: 5
    lens: [patterns, defaults, consequences]
    title: Built-in Helvetica, no new font
    disposition: accepted
    disposition_rationale: "Use the chart's built-in Helvetica. Drop a character it cannot draw, then measure what remains. A note with nothing drawable is an empty cell, and the export still writes the page. A font file is a later slice."
  - id: 6
    lens: [forces, alternatives]
    title: Skip ApplyTrend instead of changing it
    disposition: accepted
    disposition_rationale: "Do not change ApplyTrend. A series with no weigh-in still prints. A duplicate day or an unsorted series still fails the export and writes nothing. The TUI load error stays on the TUI."
  - id: 7
    lens: [patterns, consequences]
    title: New package, not the chart painter
    disposition: accepted
    disposition_rationale: "The sheet gets its own package under internal/. cmd/hdtools only wires the flag. This slice does not edit the chart painter, and the TUI does not import a PDF library. Issue #113 can call this package later."
  - id: 8
    lens: [patterns, coherence]
    title: Copy the sheet instead of sharing it
    disposition: accepted
    disposition_rationale: "Leave the month-sheet row builder where it is. The page must match the sheet's days, columns, display-unit figures, and carried trend. It must not share a package while the blank cell, the note, and the weightless path differ."
---

## Story #1 — Portrait sheet, not chart landscape

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 1, Decision 3, S4, S31, FR1)
**Lens:** forces, consequences
**Refs:** O1

**Context.** The monthly chart PDF is one landscape US Letter page, 792 by 612
points, because that picture reads across. This command writes the month sheet:
a ruled table with one line per calendar day. The spec keeps that table on a
single portrait page and does not amend the chart picture.

**Forces.** Thirty-one day rows need vertical room; the chart needs horizontal
room. Filing `YYYY-MM-log.pdf` beside `YYYY-MM-chart.pdf` tempts one shared page
box, so a packet prints without rotating a sheet. One page is the print promise
in US3. A long note must be cut rather than open a second page. Ink is the same
split: the chart keeps its coloured title box, and this page stays black, so
`[colors]` and `NO_COLOR` do not change the sheet.

**Options not taken.** Landscape Letter, the chart's 792 by 612 box. A second
page when day 31 or a long note does not fit. A4, which the chart spec already
set aside because the book tools were US Excel. Stroking rules in the configured
weight colour.

**Choice as written.** `-log-pdf` writes one portrait page, box 612 by 792
points, vector rules, black text, no embedded image and no embedded font. Every
day row's baseline, including day 31 of a 31-day month, lies inside that box. A
content stream that draws past the media box fails even when the extracted text
contains `31`. The page count stays one. Absent `-log-pdf`, `-chart-pdf` is
still 792 by 612 and does not write the log file.

**Consequences.** The two files for one month do not share an orientation. A
later packet that staples chart and sheet has to rotate one of them; this spec
does not define a combined document. Fitting January on the portrait page is
part of the contract, not an accident of the November fixture. Page two, and a
landscape sheet, are closed for this slice.

**Pattern.** Two print artefacts that share a directory and not a page setup.
Not a strategy variant of the chart painter: orientation and ink differ, so one
page setup would couple documents that only share a filing rule.

## Story #2 — Whole month, including after today

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 2, Decision 4, Decision 5, S5, S6, S7, FR4)
**Lens:** forces, alternatives
**Refs:** —

**Context.** The on-screen month sheet has a row for every calendar day of the
open month. The chart, on screen and in the chart PDF, ends the current month at
today and does not draw carried trend past that day. US2 asks for the sheet's
days, including a blank day's carried trend.

**Forces.** A log sheet is a form. Days you have not reached are still rows, and
a weigh-in dated later in the named month is data. A chart is a plot of what has
happened; the project rule is that a chart of the current month ends at today
and does not plot carried trend into the future. This page is the form, so the
form wins, including when today is 10 November 1990 and the 15th has a weigh-in.

**Options not taken.** The chart clip: stop at today and do not carry past it.
Rows only for days stored in SQLite, dropping the blank calendar. Later days
listed, but with trend left blank so the page shows no future carry.

**Choice as written.** The named month is one row per calendar day, including
days with no database row. Days after today stay. A weigh-in on a later day is
on the page. Carried trend still shows on a day after today, from the full
series, not from the named month alone. The chart clip is not used. The page
does not print `(empty month)`; an empty log is still that calendar of blank
days.

**Consequences.** Paper can show a trend on a date the chart PDF of the same
month will not plot. The two files disagree on span on purpose. The sheet also
has no chart-style empty state: a month with nothing logged still prints every
day. Clipping this export to today would be a different document.

**Pattern.** —

## Story #3 — Delta subtracts the painted cells

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 6, S8–S13, FR6, FR7)
**Lens:** defaults, consequences
**Refs:** —

**Context.** Delta is not a SQLite column. On the month sheet it is the weight
cell minus the trend cell after each is shown to one decimal in the display
unit. That rule is already a project gotcha: 80.0 kg against a 79.9 kg trend
paints as `+0.3` pounds, not the `+0.2` a converted kilogram gap gives, and a
first weigh-in entered as 176.5 lb paints trend `176.6` and delta `-0.1`. An
exact zero must print `0.0`, not `-0.0`.

**Forces.** Someone holding the paper next to the screen subtracts the two
numbers they see. A kilogram-true residual, or the chart PDF's pencil identity
for Monthly Loss, would be a third number that matches neither cell. Display
unit pulls the same way. Stone has to stay one decimal or the two cells are no
longer texts you can subtract.

**Options not taken.** Subtract kilograms, then convert the difference. Store a
delta. Print stone as stones and pounds, which breaks the one-decimal
subtraction. Put Monthly Loss and Daily Deficit on the page using the chart's
pencil identity.

**Choice as written.** Weight and trend are the display unit to one decimal, the
same text the month sheet paints. Delta is that weight text minus that trend
text. A positive value has a leading `+`. When the painted cells are equal,
delta is `0.0`, never `+0.0` or `-0.0`. The cell is blank when the day has no
weigh-in or no trend. 80 kg in stone is `12.6`. Stored kilograms are not
rewritten. Monthly Loss, Daily Deficit, and `(empty month)` are not on the page.

**Consequences.** The printed delta will not sum back to a kilogram change
without re-rounding, and it is not the chart's monthly-loss figure. A later
"correction" that subtracts stored kilograms breaks S9 and S10 on purpose. Stone
on this page cannot grow a stones-and-pounds column without a new subtraction
rule. The chart's analysis numbers stay off the sheet.

**Pattern.** The project's derived-display rule, not a stored column: subtract
the one-decimal texts the adjacent cells already show. Recorded in AGENTS.md
from the delta-column work. No GoF name.

## Story #4 — No mark for a missing cell

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 7, S12, S14, S15, S16, FR8, FR9)
**Lens:** alternatives, defaults
**Refs:** —

**Context.** The month sheet paints a missing weight, trend, sleep, steps,
or workout as an em dash. A missing delta is an empty string. The filled
sheet is the paper record after the month is entered, and a missing value
on that paper is a blank.

**Forces.** A logged zero is real data (`0.0`, `0`, `no`) and still has to
print, or a day that was logged as zero would look like a day that was
skipped. A hyphen would keep a missing cell visible in the text stream, and
it would also invent a mark the paper record does not use. The screen's em
dash is the same kind of mark.

**Options not taken.** The screen's em dash. A hyphen-minus, so extraction
can see a token for "no value." The word blank.

**Choice as written.** A blank cell has no mark. It is not a hyphen, not an
em dash, and not the word blank. A day with no row leaves sleep, steps,
workout, and note empty. A missing weight, trend, or delta is empty. An
empty stored note is that same empty cell. A logged zero is still `0.0`,
`0`, or `no`.

**Consequences.** A missing cell and an empty stored note look the same on
paper. A logged zero does not. The text stream has no token that means
"this cell was blank," so a test has to show the absence. The screen and
the page disagree on how a gap is drawn.

**Pattern.** —

## Story #5 — Built-in Helvetica, no new font

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 8, S16, S17, FR10, FR11)
**Lens:** patterns, defaults, consequences
**Refs:** O3, O5, #4

**Context.** The chart PDF already uses built-in Helvetica and does not embed a
font program. This page takes that font. Notes are free text: an accent, a
parenthesis, a snowman, a newline. The note stays on one line so the month stays
on one page.

**Forces.** A font file would keep every Unicode scalar, and the slice lists
that as a different sitting. Replacing an undrawable character with `.` or `?`
would leave a mark and would change the measured width. Measuring the stored
string before the drop spends column width on a character that will not be
drawn, and can add `...` the remaining text did not earn. The page keeps every
character Helvetica can draw, in order, and ellipsizes only that remainder.

**Options not taken.** Embed a Unicode font so the snowman survives. Replace an
undrawable character with `.` or `?`. Measure the stored note, cut, then drop —
the other order, which still passes S16 and S17 taken separately. Wrap the note,
or move the cut tail onto a second page.

**Choice as written.** The font is the chart's built-in Helvetica. There is no
new font file. Characters it cannot draw, including a newline, are dropped first
and are not replaced. The column then measures what remains. A remainder that
fits is not cut and gains no ellipsis: `say (hi)` followed by a snowman, once
the snowman is gone and the rest fits, is `say (hi)`. A remainder that does not
fit is shortened on that line and ends with `...`. The tail is not elsewhere on
the page. `é` is kept. A snowman is not. If nothing drawable remains, the note
cell is empty, not `-`, and the export still writes the page. Numeric cells are
not cut.

**Consequences.** The paper note is a projection onto Helvetica, not a copy of
the stored string. A note with nothing drawable is the same empty cell as a note
stored empty, and the export still exits 0. That collision was raised and left
in place; it is the same empty cell as a missing value (#4). A later font
file would change the repertoire and every width the ellipsis uses. This slice
leaves no hook for that file.

**Pattern.** A two-step pipeline: project onto the font's repertoire, then
single-line truncation with an ellipsis. Not font fallback, and not the Unicode
replacement character. The ellipsis applies only after the projection, so a
dropped character cannot by itself force it.

## Story #6 — Skip ApplyTrend instead of changing it

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 5, S6, S7, S12, S33, S34, FR5)
**Lens:** forces, alternatives
**Refs:** O2, #2

**Context.** `ApplyTrend` is the only trend math: full chronological series, no
stored column. It rejects a series that does not start with a weight, and it
rejects a duplicate or unsorted day. The month sheet can save sleep, steps,
workout, or a note with no weigh-in. The TUI load path treats a series with no
starting weight as a load error and does not keep those rows. US2 still wants
carried trend on a blank day, including after today (#2), which needs the full
series when a weigh-in exists.

**Forces.** Changing `ApplyTrend` so a leading weightless run is legal would
also change the screen's load error, and this slice does not own that screen.
Failing the export copies that error onto a command whose job is the filled
sheet: the log with notes would produce no paper, and a database with no rows
still would (O2). Disorder is a different force. A trend on a duplicate or
unsorted series would not be the sheet's number, with or without a weigh-in.

**Options not taken.** Change `ApplyTrend` to skip leading days with no
weigh-in. Fail the export on `ErrNoTrendStart` and put the TUI load error on the
page. Use `ApplyTrend`'s error set as the policy, so "no starting weight" and
"bad order" are whatever that function returns. Sort a copy, or keep the last of
a duplicate day, and print anyway.

**Choice as written.** The command loads the full series. A duplicate day or a
day out of order fails the export, prints no path, and writes no file, whether
or not any row has a weigh-in. An empty log skips `ApplyTrend` and is still a
sheet of blank days. A series that has rows and no weigh-in also skips
`ApplyTrend` and still writes the page: sleep, steps, workout, and note print,
and weight, trend, and delta are `-`. Otherwise `ApplyTrend` runs on the suffix
that starts at the first weigh-in and does not see the leading days with no
weigh-in. `ErrNoTrendStart` is not an export failure. This command does not copy
the TUI's load error onto the page. `ApplyTrend` is not changed. A day before
the first weigh-in has no trend. A later day with no weigh-in shows the carried
trend.

**Consequences.** Order is this export's own precondition, including on the path
that does not call `ApplyTrend`. A missing starting weight is not a failure.
Screen and paper diverge for a weightless series until something else changes
the TUI. Leading days stay without a trend, which is the unchanged function's
result, not a new smoothing rule. Editing `ApplyTrend` so this PDF can succeed
is out of bounds.

**Pattern.** This looks like a call-site Adapter (Gamma, Helm, Johnson,
Vlissides, *Design Patterns*, 1994): the export hands `ApplyTrend` a suffix the
function already accepts, and the function stays closed. The ordering check
cannot live only inside that call, or the weightless path would not run it.

## Story #7 — New package, not the chart painter

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 11, S31, FR18; out of scope)
**Lens:** patterns, consequences
**Refs:** #1

**Context.** The chart writer is already its own package under `internal/`. The
TUI calls that package and does not import the PDF library itself. This slice
adds a second PDF. Issue #113 is the later key on the month sheet; this slice
has no TUI key.

**Forces.** A table mode inside the chart painter would put the portrait
calendar (#1) in the package that owns the landscape clip, and a chart change
could move the sheet. Painting inside `internal/tui` would put a PDF library on
the screen for a slice that does not add a key. The chart's shape — a sibling
writer, the command as wiring — is the one the repo already uses. This slice
stops at the command.

**Options not taken.** Extend the chart painter with a table page. Draw the
sheet inside `internal/tui` and import the PDF library there. Wire the month
sheet to the new writer now, the way the chart screen already calls its writer.

**Choice as written.** The writer is a new package under `internal/`.
`cmd/hdtools` only wires the flag. This slice does not edit the chart painter.
Absent `-log-pdf`, `-chart-pdf` still writes only the chart page. The TUI gains
no key and does not import a PDF library.

**Consequences.** The chart's page box, clip, and colours stay isolated from the
sheet. Issue #113 can call this package the way the chart screen calls its
writer; it should not grow a second painter or import the PDF library into the
TUI. A shared rendering toolkit for both PDFs is not part of this slice. The
chart package is not the extension point for the log sheet.

**Pattern.** The chart spec's split, repeated: the command wires, the writer
lives under `internal/`, and the TUI does not take a PDF-library import (chart
spec decision 2; the TUI package says the same). CUPID Composable (Dan North):
the page is a separate unit from the screen and from the chart.

## Story #8 — Copy the sheet instead of sharing it

**Source:** `docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md`
(Decision 3, Decision 11; out of scope)
**Lens:** patterns, coherence
**Refs:** #3, #4, #5, #6, #7

**Context.** The chart spec put clip, empty-month, and analysis omission in one
helper that both the TUI and the chart PDF call. This spec does not do that for
the month sheet. The screen's builder uses an em dash, headers `date` and `wd`,
and an empty string when delta has nothing to subtract. US2 says the paper
matches the screen. Decision 1 says the page is not a screenshot. Those agree
only if "match" means the filled grid.

**Forces.** A shared builder now would have to pretend the screen and the page
are one row. They are not: the blank glyph (#4), the note projection (#5), and
the weightless preface (#6) already differ, and the figures (#3) must not.
Importing the TUI from the new package (#7) would pull the screen into a command
that must not start it. The slice can check the page against literals without
moving the builder.

**Options not taken.** Move the month-sheet row builder into a package both
sides call, the chart's one-owner pattern. Have the PDF import the TUI's
formatters. Render the terminal view and embed that text, which is the
screenshot Decision 1 rejects.

**Choice as written.** The month-sheet row builder stays where it is. The check
is that the PDF rows match the sheet: every calendar day, the same columns, the
same display-unit figures, the same carried trend. It is not that the two
screens share a package. Header cells are the words `day` and `weekday`, not the
screen's `date` and `wd`. The blank glyph, the note projection, and the
weightless path stay the PDF's rules (#4, #5, #6). Moving the builder into a
shared package is out of scope.

**Consequences.** Two implementations of a day row will exist after this slice.
They may diverge on chrome, header words, the blank glyph, and the weightless
load error. They may not diverge on the painted figures. A later extraction has
to start from this spec's literals, not from a move of the screen builder, or it
will copy the em dash and `date`/`wd` onto the page. The acceptance text is the
only contract between the copies.

**Pattern.** Duplication over the wrong abstraction (Sandi Metz, "The Wrong
Abstraction," 2016): a shared builder would pretend the screen and the page are
the same row. This is the opposite of the chart spec's single owner for data
rules, on purpose. It is not yet an Anti-Corruption Layer (Evans, *Domain-Driven
Design*, 2003); the extraction is deferred, and the scenarios are the contract.
