---
spec: docs/superpowers/specs/2026-09-10-monthly-charts.md
date: 2026-09-10
mode: spec
cartographer_model: grok-4.6
stories:
  - id: 1
    lens: [forces, alternatives]
    title: Limit Charts to One On-Screen Month
    disposition: accepted
    disposition_rationale: "Intentional slice. Long-term and PDF stay in idea.md. Do not freeze month-shaped X as architecture."
  - id: 2
    lens: [patterns, alternatives]
    title: Open Chart as Sibling Screen
    disposition: accepted
    disposition_rationale: "Matches list/form/month. c only when not editing is cowpath, not a lasting keybinding law."
  - id: 3
    lens: [forces, consequences, coherence]
    title: Clip the Plotted Span at Today
    disposition: promoted
    disposition_rationale: "Period-to-date for the current month. PDF and long-term charts must not draw carry into the future. Carry into AGENTS.md DESIGN_DECISIONS."
  - id: 4
    lens: [defaults, consequences]
    title: Define Empty as Both Series Absent
    disposition: accepted
    disposition_rationale: "Empty versus zero for this plot. Later charts may use a different empty."
  - id: 5
    lens: [patterns, alternatives, defaults]
    title: Paint Two Series with Daily on Top
    disposition: accepted
    disposition_rationale: "Book diamond-on-the-line; existing weight/trend roles. Terminal z-order is not a PDF contract."
  - id: 6
    lens: [patterns, defaults]
    title: Map Days onto a Character Grid
    disposition: accepted
    disposition_rationale: "Testable TUI geometry. One character per day does not apply to PDF."
  - id: 7
    lens: [forces, alternatives, consequences]
    title: Compute Deficit from Endpoint Trends
    disposition: promoted
    disposition_rationale: "First-last trend, 3500 kcal/lb in pounds, divisor is plotted days, never SQLite. Carry into AGENTS.md ARCH_DECISIONS."
  - id: 8
    lens: [forces, alternatives]
    title: Keep Chart Copy Non-Prescriptive
    disposition: accepted
    disposition_rationale: "Already a harness constraint (no medical claims). No Eat! or goal overlay until idea.md has a goal."
---

## Story #1 — Limit Charts to One On-Screen Month

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md` (opening, Decision 1, Out of scope); `idea.md` (Monthly Log)
**Lens:** forces / alternatives
**Refs:** —

**Context.** `idea.md` still wants on-screen monthly charts, on-screen
long-term charts, PDF of filled monthly sheets, and PDF of monthly
charts, all in the book's style. The dated spec takes only the monthly
on-screen sitting.

**Forces.** The next sitting after the sheet is the same month, already
loaded and trended, with the same colour roles. A character-per-day X
axis fits a calendar month on a typical terminal and does not fit a
year. PDF and multi-month views need a different encoding and a
different output path.

**Options not taken.** Ship long-term charts in the same slice. Emit
PDF of the plot or of the sheet. Draw a sparkline on the sheet instead
of a month-scale picture. Defer Monthly Loss / Daily Deficit until PDF.

**Choice as written.** This slice is one calendar month, on screen.
Long-term charts, PDF of charts, PDF of sheets, and blank-sheet
printing stay out of scope. The two end-of-month numbers travel with
the on-screen plot, not with a later export.

**Consequences.** A later PDF sitting is expected to reuse the same
identity; the plan already hosts that arithmetic in `internal/dailylog`
for that reason, which the spec never names. Long-term charts cannot
inherit this X geometry. The backlog in `idea.md` remains larger than
the spec.

**Pattern.** Thin vertical slice against a larger book-shaped backlog.

## Story #2 — Open Chart as Sibling Screen

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decisions 2, 8, 9; FR1, FR2, FR8, FR10)
**Lens:** patterns / alternatives
**Refs:** O6, O7

**Context.** The month sheet is already a full-height table of every
calendar day. `c` is not a command on that sheet today; any
non-workout rune starts an edit. List and sheet already have an
Esc-return path via `afterSave` for the form.

**Forces.** A plot of at least eight rows plus title, axes, and two
analysis numbers will not sit under thirty-one table rows on a 24-row
terminal. Stealing `c` while a note is being typed would regress
"cardio". Chart and sheet must agree on which month is in view, or Esc
after `[` lands on the month you left rather than the month you
browsed.

**Options not taken.** A widget or split under the sheet. A different
key that never collides with type-to-edit. A chart-local year/month
that leaves the sheet behind. `c` as text even when not editing, so
the chart opens only from the list. Interactive cursor on the plot.
Opening from an empty list refused, or pointed at the last logged
month.

**Choice as written.** The chart is a new screen, not a widget. `c`
opens it for the sheet's month, or for the selected list day's month,
and only when the sheet is not editing; while editing, `c` is text. An
empty list opens the current calendar month. Esc returns to the
opener. `[` / `]` on the chart share year and month with the sheet, so
Esc to the sheet shows the browsed month; Esc to the list leaves the
list cursor unchanged. Type, Tab, and Space do not write.

**Consequences.** The month cursor is shared even when the chart was
opened from the list, so list → `c` → `[` → Esc → `m` can show a month
the list cursor is not on. The plan adds `afterChart` beside
`afterSave` so form-return and chart-return do not overwrite each
other — a choice that lives only in the plan.

**Pattern.** Sibling screen over a shared month navigator, with a
previous-screen pointer; command keys yield to an in-progress edit.

## Story #3 — Clip the Plotted Span at Today

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decisions 3, 7, 11; FR6, FR14; S15)
**Lens:** forces / consequences / coherence
**Refs:** O2, #2

**Context.** The book's pencil page works a completed month. Its
computer tools end the current month on today. `buildMonthSheet`
already fills every calendar day, including days after today, and
carries trend onto those rows so the table matches a paper log.

**Forces.** Carrying trend past today fabricates a flat tail and, if
used as the last analysis day, dilutes Daily Deficit by days that have
not occurred. Matching the sheet "inside the plotted span" is not the
same as matching the sheet's full calendar.

**Options not taken.** Always plot day 1 through the last calendar
day, including the future. Omit analysis until the month completes.
Clamp `]` so a future month cannot be opened. End the current month
on the last weighed day rather than today.

**Choice as written.** Last plotted day is the last calendar day of a
past month, today if the month contains today, and none if the whole
month is after today. Do not draw carried trend into the future. The
deficit divisor is day 1 through that last plotted day, not the count
of weighed days.

**Consequences.** For the in-progress month the chart is not the sheet
rotated: the sheet still shows post-today rows with carried trend; the
chart has no columns there. `]` into a future month is allowed and
yields the empty state, not a clamp. Past blank days still carry, so
"blank" and "not yet occurred" are different.

**Pattern.** Period-to-date clipping on an otherwise complete calendar
series.

## Story #4 — Define Empty as Both Series Absent

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decisions 6, 7, 11, 12; FR7; S7, S7a)
**Lens:** defaults / consequences
**Refs:** O1, #3

**Context.** `ApplyTrend` plus carry-forward mean a month with no log
rows still has trend on every day once any earlier weigh-in exists.
"No rows", "no daily marks", and "no endpoint trend" are different
sets. The common `]` after the last weigh-in is the second of those,
not the first.

**Forces.** Calling that month empty would hide the carry the sheet
already shows. Calling any no-row month populated would print a flat
line and 0/0 where there is no series at all.

**Options not taken.** Empty iff no SQLite rows in the month. Empty
iff no daily marks, ignoring carry. Empty iff analysis endpoints are
missing. Refuse to open months with no rows.

**Choice as written.** Empty means the plotted span has no daily mark
and no trend, including no carry: empty-state copy, no scale, no
analysis numbers, no panic. A no-row month with carried trend is not
empty: plot the flat trend; Monthly Loss 0 and Daily Deficit 0.

**Consequences.** First-log and pre-first-weigh-in months are empty;
post-last-weigh-in months are a flat carry at 0/0. Missing an endpoint
still omits analysis even when the span is not empty. Empty has no Y
scale; carry-only uses the 1.0-unit floor because min equals max.

**Pattern.** Empty versus zero: absence of a series, not absence of
rows.

## Story #5 — Paint Two Series with Daily on Top

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decision 4; FR3–FR5; S3–S5; Observing)
**Lens:** patterns / alternatives / defaults
**Refs:** O3

**Context.** The sheet already paints daily weight with the `weight`
role and trend with `trend`. `ApplyTrend` starts equal to the first
weight and often rounds onto the same 0.1 kg as the daily reading, so
the two values share a character cell on a short plot. Colour-only
encoding is forbidden; `NO_COLOR` already strips chroma.

**Forces.** One cell holds one glyph. The book's floats-and-sinkers
put a diamond on the line when the series coincide. US2 needs the two
series distinct by mark across the chart, not two glyphs in one cell.

**Options not taken.** New `[colors]` keys. Stems or a third "both"
glyph. Colour-only distinction. Side-by-side columns per day. Trend
on top. Drop daily marks on collision so the path stays connected
through that cell.

**Choice as written.** Daily weight is a discrete mark (`o` or
equivalent); trend is a connected path. Daily uses `weight`; trend
uses `trend`; no new config keys. When both fall in one cell, the
daily mark wins. Distinction is glyph type across the chart. Under
`NO_COLOR`, skip chromatic foreground; both series remain.

**Consequences.** A collision breaks the path glyph in that column;
tests must still see a connected path across adjacent columns and a
daily mark in the shared cell. Palette typos still fail-open to the
built-in default. PDF later can reuse the same two roles; it cannot
reuse a terminal z-order.

**Pattern.** Overplotting / painter's algorithm; the book's
diamond-on-the-line without stems.

## Story #6 — Map Days onto a Character Grid

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decisions 5–7; FR9, FR14, FR15; S9, S16; Observing)
**Lens:** patterns / defaults
**Refs:** O4, O5, #1, #3, #5

**Context.** Tests drive `View()` with no TTY and must locate day *N*.
Auto min/max plus fractional padding is undefined when every plotted
value is the same. Storage is kilograms; the axis is not.

**Forces.** A no-TTY assertion needs a named cell for each day. A
single weigh-in or a carry-only flat line has min = max, so a span of
zero is a divide-by-zero. Display unit is the only scale the reader
has configured. One character per day is what makes a month-shaped
slice fit.

**Options not taken.** Variable-width day columns. Scale Y to trend
only. Flush the series to the frame with no padding. Collapse or
refuse a flat series. Label every day. Fill the terminal height
instead of a minimum.

**Choice as written.** X is one character column per day from day 1
through the last plotted day, after a Y-label gutter; day *N* is the
*N*th plot column; at least day 1 and the last plotted day are
labeled; plot height is at least 8 rows when there is data. Y is min
and max of converted values in the span, padded so the series is not
glued to the frame. When min equals max, the span is 1.0 in
`display_unit`, centred on that value. An empty month has no scale.
Axis labels use `units.FromKG`.

**Consequences.** The padding fraction is unnamed, so two paddings
both pass until a test pins it. Height is a floor, not a window-fill.
Future days have no columns. Flat and single-point months draw on a
1.0-unit band.

**Pattern.** Typewriter / character-raster plot with a minimum scale
span.

## Story #7 — Compute Deficit from Endpoint Trends

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Decision 11; FR12, FR13; S11–S14; Out of scope)
**Lens:** forces / alternatives / consequences
**Refs:** #3, #4

**Context.** The cited pencil page takes first trend, last trend,
converts the drop to pounds, multiplies by 3500, and divides by the
month's day count. The online tools also offer a last-seven-days fit,
which this slice rejects. Trend is derived and must not be stored.
Display unit may be kg or st.

**Forces.** The 3500 kcal/lb identity is defined in avoirdupois
pounds; an SI restatement would miss the book's 257-calorie worked
example. Daily weights are the noise the trend exists to ignore.
Dividing by weighed days would treat skipped days as if they were not
part of the span. A current month must divide by elapsed plotted
days, or the identity is applied to days that have not happened.

**Options not taken.** Linear regression or last-seven-days analysis.
First and last daily weight. Divisor = count of weighed days. Convert
3500 into kcal per kg and skip the pound step. Persist the two
numbers. Print Weekly Loss as a third figure.

**Choice as written.** When day 1 and the last plotted day both have a
trend (carry allowed), Monthly Loss is first − last in `display_unit`
(positive = fell). Daily Deficit is that loss converted to pounds ×
3500 ÷ days in the plotted span, whole calories (positive =
shortfall). The 3500 factor is always pounds. Missing either endpoint
omits both numbers. They are not written to SQLite.

**Consequences.** The same kilogram drop always yields the same
calorie figure regardless of `display_unit`; only the loss label
changes. A sparse month still divides by calendar or elapsed length,
not by how often the user stepped on the scale. Carry-only spans are
0 and 0.

**Pattern.** Pencil-and-paper accounting identity rather than a
statistical estimator.

## Story #8 — Keep Chart Copy Non-Prescriptive

**Source:** `docs/superpowers/specs/2026-09-10-monthly-charts.md`
(Background, Decision 10, FR11, Out of scope)
**Lens:** forces / alternatives
**Refs:** —

**Context.** The book plots an "Eat!" band once a goal weight exists.
This project's daily log has no goal and no calorie budget. Background
already frames the screen as a tracking picture and an accounting
identity, not a diagnosis.

**Forces.** There is no goal in the data model, so an Eat! indicator
or goal line would invent a target the log cannot store.
Diagnosis/treatment language would outrun what first-and-last trend ×
3500 is.

**Options not taken.** Goal lines. Calorie bands. An "Eat!" indicator.
Copy that diagnoses or prescribes. Weekly Loss as a third coaching
number.

**Choice as written.** The screen is logged numbers plus the 3500
kcal/lb identity. Copy does not claim diagnosis or treatment and does
not tell the reader what to eat. Goal lines, calorie bands, and Eat!
are out of scope for this slice.

**Consequences.** Title, empty-state, Monthly Loss, and Daily Deficit
must be reviewable as strings without medical or dietetic verbs. A
later goal feature would be a schema change, not a chart overlay on
this slice.

**Pattern.** Descriptive display versus prescriptive overlay.
