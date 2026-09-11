---
spec: docs/superpowers/specs/2026-09-10-monthly-charts.md
date: 2026-09-10
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: critical
    claim: "Empty month is three incompatible predicates, so a month with no rows but a carried trend cannot both show the empty state and plot or analyse that carry."
    evidence: "S7 Given: 'a month with no log rows' Then: 'the view says the month is empty'; FR7: 'A month with no values'; Decision 7: 'A carried trend still plots'; Decision 11: 'carry-forward allowed' and 'empty month, or a month with no series yet'."
    disposition: accepted
    disposition_rationale: "Empty only when the plotted span has no daily marks and no trend (including no carry). A no-row month with carry is not empty: plot the flat trend; first equals last so Monthly Loss 0 and Daily Deficit 0."
  - id: O2
    category: premise
    severity: high
    claim: "First-and-last calendar-day trend with carry-forward, divided by the full month length, treats unoccurred future days as blank log days, so the in-progress month reports a diluted Daily Deficit and draws a flat trend into the future."
    evidence: "Decision 11: 'first and last trend of the calendar month (carry-forward allowed)'; 'Days in the month is the calendar length (28–31), not the count of weighed days'; Decision 3: 'Trend carry-forward onto blank days matches the sheet'; Decision 7: 'A carried trend still plots'."
    disposition: accepted
    disposition_rationale: "A month that contains today is plotted and analysed only through today. Divisor is days from day 1 through the last plotted day. Do not draw carried trend into the future. Past months use the full calendar month."
  - id: O3
    category: premise
    severity: high
    claim: "Daily marks and a trend path that share one character cell cannot remain distinct by mark, so US2, S3, and S5 fail on the many days where 10% smoothing maps weight and trend onto the same row."
    evidence: "US2: 'daily weight and trend to remain distinct by mark, not only by colour'; Decision 4: 'Daily weight is discrete marks. Trend is a connected path'; Observing: a discrete mark 'is not part of a horizontal run of the same glyph' and a connected path 'join[s] left-to-right across adjacent day columns'."
    disposition: accepted
    disposition_rationale: "Daily mark and path use different glyphs. When they share a cell the daily mark wins (diamond on the line). Distinction is glyph type across the chart, not two glyphs in one cell."
  - id: O4
    category: specification quality
    severity: high
    claim: "Auto Y range is min/max plus padding, which is undefined when min equals max, so a single weigh-in or a flat series has no specified scale and may panic or collapse."
    evidence: "Decision 6: 'The vertical scale is the min and max of values plotted that month (daily weights and trend), with padding so the series is not glued to the frame. An empty month has no scale.'"
    disposition: accepted
    disposition_rationale: "When min equals max, use a minimum Y span of 1.0 in the display unit so a flat or single-point series still draws."
  - id: O5
    category: specification quality
    severity: high
    claim: "S6 and the observing rules require identifiable day columns, but the spec specifies neither plot height, column width, axis gutter, nor how to locate day N, so two layouts can both claim to pass."
    evidence: "Decision 7: 'One column per day of the month, left to right, day 1 through the last day'; S6: 'there is no daily mark for the 2nd or 3rd'; Observing: a connected path 'join[s] left-to-right across adjacent day columns'."
    disposition: accepted
    disposition_rationale: "Testable geometry: one character column per calendar day; Y labels left of day 1; at least day 1 and last day labeled; plot at least 8 rows when there is data. Day N is the Nth plot column after the gutter."
  - id: O6
    category: specification quality
    severity: high
    claim: "Binding c on the month sheet is specified without saying whether it preempts type-to-edit, so either a cell can no longer be started with the letter c, or c never opens the chart from the sheet."
    evidence: "Decision 2: 'c opens the chart for the month in view (from the month sheet) or for the selected day's month (from the list)'; Decision 9 constrains Type/Tab/Space only on the chart; existing month sheet: any non-workout rune begins an edit."
    disposition: accepted
    disposition_rationale: "c is a command only when not editing. While editing, c is text (note cardio). Same as other command keys on the sheet."
  - id: O7
    category: specification quality
    severity: high
    claim: "Esc returns to the opening screen and brackets change the chart's month, with no rule for whether the month sheet follows, so Esc after browsing months lands on two different months."
    evidence: "Decision 2: 'Esc returns to the screen that opened it'; Decision 8: '`[` / `]` change month on the chart, as on the sheet'; S8 asserts only the chart title; S1/S2 assert only which screen Esc restores."
    disposition: accepted
    disposition_rationale: "Chart [ / ] share year/month with the sheet. Esc to the sheet shows the month you browsed to. Esc to the list leaves the list cursor unchanged."
  - id: O8
    category: specification quality
    severity: medium
    claim: "S8 requires ] twice to get from October back to November, which contradicts Decision 8's one-month step and cannot be implemented as written."
    evidence: "S8: Given November 1990, When `[` Then October 1990, When `]` is pressed twice Then November 1990 again; Decision 8: '`[` / `]` change month on the chart, as on the sheet'."
    disposition: accepted
    disposition_rationale: "S8 is a typo. After [ to October, one ] returns to November."
---

## O1 — specification quality — critical

### Claim

Empty month is three incompatible predicates, so a month with no rows
but a carried trend cannot both show the empty state and plot or
analyse that carry.

### Evidence

S7:

> **Given** a month with no log rows
> **When** the chart is shown
> **Then** the view says the month is empty

FR7: "A month with no values shows an empty state and does not panic."

Decision 3: "The chart uses the already-loaded, already-trended
series. […] Trend carry-forward onto blank days matches the sheet."

Decision 7: "Days without a weight have no daily mark. A carried
trend still plots."

Decision 11 computes Monthly Loss and Daily Deficit from the first
and last trend "of the calendar month (carry-forward allowed)", and
omits the numbers "If day 1 or the last day has no trend […].
(empty month, or a month with no series yet)."

The sheet already fills every calendar day from the full series: a
month with no rows still has `HasTrend` on every day when an earlier
weigh-in exists (`buildMonthSheet` / `lastTrendOnOrBefore`).

### Why this matters

"No log rows", "no values", and "no endpoint trend" are different
sets. The common case — `]` into a month not yet logged, or any month
after the last weigh-in — has no rows and a full carried trend. One
implementation shows S7's empty state and omits the numbers; another
plots a flat line and prints Monthly Loss 0 and Daily Deficit 0. Both
can cite the spec. Until empty is one predicate, S7, FR7, Decision 6
("An empty month has no scale"), Decision 7, and Decision 11 cannot
be implemented together.

## O2 — premise — high

### Claim

First-and-last calendar-day trend with carry-forward, divided by the
full month length, treats unoccurred future days as blank log days,
so the in-progress month reports a diluted Daily Deficit and draws a
flat trend into the future.

### Evidence

Decision 11:

> Computed from the first and last *trend* of the calendar month
> (carry-forward allowed), not from daily weights.
>
> Daily Deficit (whole calories) = (that loss converted to pounds)
> × 3500 ÷ days in the month.
>
> Days in the month is the calendar length (28–31), not the count of
> weighed days.

Decision 3 requires carry-forward onto blank days to match the sheet.
Decision 7 plots that carried trend. US4 says the numbers are "taken
from the first and last trend of that month, so the chart matches the
book's analysis."

The cited pencil page works a *completed* July: first trend, last
trend, divide by 31. The book's computer tools end a current-month
chart on today, not on the last calendar day. Every numbered scenario
is a completed historical month (November 1990, a 30-day month with
both endpoints). None is the in-progress month. Out of scope rejects
the online last-seven-days fit, not "end at today."

### Why this matters

The month a person actually opens is the current one. Matching the
sheet means days after today already receive carried trend. On the
10th of a 30-day month the chart then draws a flat line through the
30th and reports Daily Deficit ≈ (loss so far × 3500) / 30 — about
one third of the rate over the days that have happened. That is not
the book's completed-page identity; it is that identity applied to
fabricated future blanks. The spec never distinguishes past blank
days (carry is the sheet) from days that have not occurred. An
implementation that omits analysis until the month is complete, or
that ends the series on today and divides by elapsed days, would
match the book's intent and fail S11's calendar rule if applied to
an in-progress month. The premise that this slice "matches the
book's analysis" is only true for finished months, which the
acceptance list exclusively tests.

## O3 — premise — high

### Claim

Daily marks and a trend path that share one character cell cannot
remain distinct by mark, so US2, S3, and S5 fail on the many days
where 10% smoothing maps weight and trend onto the same row.

### Evidence

US2: "daily weight and trend to remain distinct by mark, not only by
colour."

Decision 4: "Daily weight is discrete marks. Trend is a connected
path."

S3 and S5 require both series present and distinct by mark, including
under `NO_COLOR`.

Observing the chart:

> A **discrete mark** is a non-space glyph that is not part of a
> horizontal run of the same glyph. A **connected path** is a run of
> glyphs that join left-to-right across adjacent day columns
> (line-drawing or ASCII `-` / `/` / `\` / `|` is fine).

`ApplyTrend` starts trend equal to the first weight and then moves
10% of the gap, rounded to 0.1 kg. On day one, and on any day the
increment rounds to zero, the two values are the same number.
Decision 6 scales Y to that month's min/max. A typical month spans a
kilogram or two; a terminal plot has on the order of ten rows.
Differences of 0.1 kg share a cell.

### Why this matters

A character cell holds one glyph. If that glyph is the daily mark,
the path is broken; if it is a path stroke, the daily mark is absent;
if it is a third "both" glyph, that glyph is unspecified and is not
"not part of a horizontal run of the same glyph." Colour cannot break
the tie: US2 and S5 forbid colour-only distinction, and `NO_COLOR`
removes chroma. The book's floats-and-sinkers exist specifically to
show deviation when the two series coincide (diamond plus stem). This
spec forbids stems, forbids colour-only encoding, and overlays two
series on a one-glyph grid. US1's "line among the noise" is then
unreadable on the days the 10% average is doing its job. S3 can be
made green only by picking a month where weight and trend never
collide — which is not the series the feature is for.

## O4 — specification quality — high

### Claim

Auto Y range is min/max plus padding, which is undefined when min
equals max, so a single weigh-in or a flat series has no specified
scale and may panic or collapse.

### Evidence

Decision 6:

> The vertical scale is the min and max of values plotted that month
> (daily weights and trend), with padding so the series is not glued
> to the frame. An empty month has no scale.

US3 and S6 require a chart of a partial month. A month with one
weighed day, or with only carried trend, or with several days at the
same rounded trend, has min = max. Padding described as a fraction of
(max − min) is then zero. The only empty case given a defined scale
behaviour is "empty month" (already contradictory; see O1). No
scenario covers a flat or single-point series. Mapping a value onto a
plot with span zero is a divide-by-zero unless an implementer invents
a minimum span.

### Why this matters

Partial months are in scope. The first weigh-in of a log, a
maintenance month that does not move 0.1 kg, and a no-row month with
carried trend are all min = max. One implementation panics (fails
S7's "does not crash" if that month is also called empty); another
refuses to draw; another invents a band. The plan already had to
invent "a minimum span so a flat month is still drawn" — evidence the
spec does not decide it. Until zero-span Y is specified, FR3/FR6/FR7
are not a single behaviour.

## O5 — specification quality — high

### Claim

S6 and the observing rules require identifiable day columns, but the
spec specifies neither plot height, column width, axis gutter, nor
how to locate day N, so two layouts can both claim to pass.

### Evidence

Decision 7: "One column per day of the month, left to right, day 1
through the last day."

S6: "there is no daily mark for the 2nd or the 3rd" and "trend still
plots on the 2nd and 3rd if the sheet would show a carried trend
there."

Observing the chart defines a path as glyphs that "join left-to-right
across adjacent day columns." Tests "must not require a TTY" and
drive `App.Update` / `View`. The current App does not record window
size. Nothing states how many cells a day occupies, where column 1
starts relative to Y labels, whether day numbers appear, or how tall
the plot is.

### Why this matters

S6 is the blank-day scenario. Without a rule that names the cells for
day 2 and day 3, a test can pass by finding any gap and any path,
including a path that is not on those days. A 1-cell-per-day plot and
a 2-cell-per-day plot both satisfy "one column per day." A 3-row
sparkline and a 40-row chart both satisfy "discrete marks" and
"connected path." FR3, FR6, S3, and S6 are then not decidable under
the project's no-TTY rule. Geometry here is not decoration; it is
the observable.

## O6 — specification quality — high

### Claim

Binding `c` on the month sheet is specified without saying whether it
preempts type-to-edit, so either a cell can no longer be started with
the letter c, or `c` never opens the chart from the sheet.

### Evidence

Decision 2: "`c` opens the chart for the month in view (from the
month sheet) or for the selected day's month (from the list)."

S1 is that path. Decision 9 says type, Tab, and Space do not write
*on the chart*; it does not mention the sheet.

The month sheet already starts an edit on any single rune except
workout (`updateMonth` default: `beginEdit`). Notes, and any
non-numeric field reached in a sitting, are entered that way. `c` is
not currently a command key.

### Why this matters

Two implementations both match the spec: (1) steal `c` on the sheet,
so "cardio" cannot be started by typing `c` and the month-sitting
workflow regresses; (2) leave `c` in type-to-edit, so S1 is false
unless the user is on the workout column or already at the list. The
spec never says `c` is a command only when not editing, only on the
list, or only on numeric columns. That collision is a class of
failures, not a keybinding taste.

## O7 — specification quality — high

### Claim

Esc returns to the opening screen and brackets change the chart's
month, with no rule for whether the month sheet follows, so Esc after
browsing months lands on two different months.

### Evidence

Decision 2: "Esc returns to the screen that opened it."

Decision 8: "`[` / `]` change month on the chart, as on the sheet."

S1: Esc from a November chart opened on the month sheet returns to
the month sheet. S8: `[` then the *chart* shows October. Neither S1
nor S8 says which month the sheet shows after that Esc. "As on the
sheet" can mean shared `monthModel` (brackets mutate the sheet) or
only the same key meaning (chart holds its own year/month).

### Why this matters

Open November sheet, `c`, `[`, Esc. Implementation A: November sheet
(chart month is separate). Implementation B: October sheet (chart
reused month stepping). Both satisfy S1 and S8. The next `c` then
opens a different month. "The screen that opened it" names a screen,
not a month. Until that coupling is specified, FR2 and FR8 together
have two behaviours.

## O8 — specification quality — medium

### Claim

S8 requires `]` twice to get from October back to November, which
contradicts Decision 8's one-month step and cannot be implemented as
written.

### Evidence

S8:

> **Given** the chart for November 1990
> **When** `[` is pressed
> **Then** the chart shows October 1990
> **When** `]` is pressed twice
> **Then** the chart shows November 1990 again

Decision 8: "`[` / `]` change month on the chart, as on the sheet."
On the sheet, one `]` is one month forward. Sequential reading of S8:
`[` then `]` `]` lands on December, not November. Independent reading
of the second When from the original Given: `]` twice from November
is January, not "November again." No reading is a round-trip of one
month each way.

### Why this matters

S8 is the only acceptance scenario for FR8. An implementation of
Decision 8 (one press, one month) fails S8 as written. An
implementation of S8 as written fails Decision 8. The scenario cannot
be transcribed into a test without silently rewriting "twice."

## Explicitly not objecting to

-   **One-month on-screen slice**: long-term charts and PDF stay out
    of scope, as `idea.md` already records.
-   **3500 kcal/lb in pounds**: even when `display_unit` is kg or st,
    and converting the plotted series with `units.FromKG`.
-   **Derived trend**: Monthly Loss and Daily Deficit are not written
    to SQLite.
-   **Existing colour roles**: reusing `weight` and `trend` instead of
    new `[colors]` keys.
-   **No-TTY tests**: driving the chart through `App.Update` / `View`.
-   **Medical copy**: forbidding diagnosis language, goal lines,
    calorie bands, and an "Eat!" indicator in this slice.
