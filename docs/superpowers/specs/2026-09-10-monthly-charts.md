# On-screen monthly weight chart

**Date**: 2026-09-10
**Status**: draft
**Issue**: [#34](https://github.com/jglueckstein/hdtools/issues/34)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)

The month sheet is a table. The book also plots the month: noisy daily
weight against the smoother trend, so the eye can see signal versus
noise, and writes Monthly Loss and Daily Deficit under the plot. This
slice adds that plot and those two numbers on screen for one calendar
month. Long-term charts and PDF are out of scope.

## Background

[Signal and Noise](https://www.fourmilab.ch/hackdiet/e4/signalnoise.html)
is why trend exists. Daily scale readings are dominated by water; the
10% moving average is the line that would have saved Dexter two months
of despair. The TUI already computes that series (`ApplyTrend`) and
paints daily weight blue and trend red. It does not yet draw them.

`idea.md` asks for on-screen monthly charts and long-term charts, in
that book's style, including Monthly Loss and Daily Deficit. One month
is the next sitting after the sheet: same data, same colours, a
picture instead of a table, plus the two end-of-month numbers from
[Calculating calorie deficit](https://www.fourmilab.ch/hackdiet/www/subsection1_4_1_0_3.html).

This is a tracking picture and an accounting identity (3500 kcal per
pound of fat), not a diagnosis.

## User stories

### US1 — See the month as a chart

As a person with a month of weights, I want to open a chart of that
month's daily readings and trend so I can see the line among the noise
without leaving the TUI.

### US2 — Read the chart without colour

As a person with `NO_COLOR` or a monochrome terminal, I want daily
weight and trend to remain distinct by mark, not only by colour.

### US3 — Empty and partial months

As a person who has not weighed every day, I want days without a weight
to omit a daily mark, and I want an empty month to show an empty state
rather than a crash.

### US4 — Monthly loss and daily deficit

As a person looking at a monthly chart, I want Monthly Loss in my
display unit and Daily Deficit in calories, taken from the first and
last trend of that month, so the chart matches the book's analysis
without me doing the arithmetic.

## Decisions

1.  **This slice is one month, on screen.** Long-term charts, PDF of
    charts, and PDF of sheets are out of scope.
2.  **A new screen, not a widget on the sheet.** The sheet stays a
    table. `c` opens the chart for the month in view (from the month
    sheet) or for the selected day's month (from the list). On the
    month sheet, `c` is a command only when **not editing**; while
    editing, `c` is text. Esc returns to the screen that opened it.
3.  **Same data as the sheet, clipped to today.** The chart uses the
    already-loaded, already-trended series. It does not query SQLite
    itself. Trend carry-forward onto blank days matches the sheet
    **inside the plotted span**. The last plotted day is: the last
    calendar day of a **past** month; **today** if the month contains
    today; none (empty) if the whole month is after today. Do not
    draw carried trend into the future.
4.  **Two series.** Daily weight is discrete marks (`o` or equivalent).
    Trend is a connected path (`-` / `/` / `\` or line-drawing). Daily
    uses the `weight` colour role; trend uses `trend`. Those roles
    already exist; no new `[colors]` keys. When both series fall in
    the same cell, the **daily mark wins** (the book's diamond on the
    line). They remain distinct by glyph type across the chart, not
    by putting two glyphs in one cell.
5.  **Display unit.** Axis labels and the implied scale are in
    `display_unit`. Storage remains kilograms.
6.  **Auto Y range.** The vertical scale is (min of plotted values)
    − *P* through (max of plotted values) + *P*, where *P* is 2 lb
    in `display_unit` (`units.ToKG(2, lb)` then `FromKG`). When min
    equals max, that is a 4 lb band centred on the value. An empty
    month has no scale. See
    [2026-09-12-chart-y-range.md](2026-09-12-chart-y-range.md).
7.  **X is the plotted span.** One **character column** per day from
    day 1 through the last plotted day. Y-axis labels sit to the left
    of day 1 (the gutter). Day *N* is the *N*th plot column after
    that gutter. The plot is at least 8 rows tall when there is data.
    At least day 1 and the last plotted day are labeled under the
    plot. Days without a weight have no daily mark. A carried trend
    still plots on blank days **in the span**.
8.  **`[` / `]` change month** on the chart, as on the sheet. They
    **share** year and month with the sheet. Esc to the sheet shows
    the month just viewed on the chart. Esc to the list leaves the
    list cursor unchanged.
9.  **No editing on the chart.** Type, Tab, and Space do not write.
    Enter is not a save.
10. **Not medical.** The screen is a plot of logged numbers plus the
    book's 3500 kcal/lb identity. Copy does not claim diagnosis or
    treatment, and does not tell the reader what to eat.
11. **Monthly Loss and Daily Deficit.** Computed from the first and
    last *trend* of the **plotted span** (carry-forward allowed), not
    from daily weights.
    - Monthly Loss (display unit) = first trend − last trend.
      Positive means the trend fell (loss); negative means it rose
      (gain).
    - Daily Deficit (whole calories) =
      (that loss converted to pounds) × 3500 ÷ days in the plotted
      span (day 1 through last plotted day).
      Positive means a calorie shortfall; negative means an excess.
    The 3500 kcal/lb factor is always applied in pounds, even when the
    display unit is kg or st. The divisor is not the count of weighed
    days. For a completed past month it equals the calendar length.
    If day 1 or the last plotted day has no trend, the two numbers
    are omitted. They are not written to SQLite.
12. **Empty month.** Empty means the plotted span has no daily mark
    and no trend (no carry). Then: empty-state copy, no scale, no
    analysis, no panic. A month with no log rows **but** a carried
    trend is not empty: plot the flat trend; Monthly Loss 0 and
    Daily Deficit 0.

## Acceptance scenarios

### S1 — Open from the month sheet

**Given** the month sheet for November 1990 with at least one weighed
day, not editing
**When** `c` is pressed
**Then** the chart screen is shown
**And** the title names November 1990
**And** the display unit appears
**And** Esc returns to the month sheet

### S2 — Open from the list

**Given** the daily list with the cursor on a November 1990 day
**When** `c` is pressed
**Then** the chart for November 1990 is shown
**And** Esc returns to the list

### S3 — Daily marks and trend path

**Given** a month with two or more weighed days
**When** the chart is shown
**Then** daily weights appear as discrete marks
**And** trend appears as a connected path
**And** the two series use different glyphs, not colour alone
**And** on a day where weight and trend share a cell, that cell shows
the daily mark

### S4 — Colour roles

**Given** a config with `weight = "green"` and `trend = "yellow"`
**And** a month with a weighed day
**When** the chart is shown
**Then** daily marks are drawn in green
**And** the trend path is drawn in yellow

### S5 — `NO_COLOR`

**Given** `NO_COLOR` is non-empty
**And** a month with a weighed day
**When** the chart is shown
**Then** the view has no chromatic foreground or background
**And** daily marks and the trend path are still both present
**And** they remain distinct by glyph type (daily mark wins a shared
cell)

### S6 — Blank days

**Given** November 1990 with weights on the 1st and 4th only
**When** the chart is shown
**Then** there is no daily mark for the 2nd or 3rd
**And** trend still plots on the 2nd and 3rd if the sheet would show a
carried trend there

### S7 — Empty month

**Given** a month with no daily weights and no carried trend
**When** the chart is shown
**Then** the view says the month is empty
**And** Monthly Loss and Daily Deficit are not shown as numbers
**And** the process does not crash

### S7a — Carry-only month is not empty

**Given** a past month with no log rows of its own and a carried
trend from the previous month
**When** the chart is shown
**Then** the view is not the empty state
**And** a trend path is shown
**And** Monthly Loss is 0
**And** Daily Deficit is 0

### S8 — Month keys on the chart

**Given** the chart for November 1990, opened from the November month
sheet
**When** `[` is pressed
**Then** the chart shows October 1990
**When** `]` is pressed
**Then** the chart shows November 1990 again

### S8a — Esc keeps the browsed month on the sheet

**Given** the November month sheet, then the chart, then `[` to
October
**When** Esc is pressed
**Then** the month sheet shows October 1990

### S8b — `c` while editing is text

**Given** the month sheet editing the note field
**When** `c` is pressed
**Then** the chart does not open
**And** the letter `c` is in the note input

### S9 — Display unit on the axis

**Given** `display_unit = "lb"` and a weighed day of 80 kg
**When** the chart is shown
**Then** the vertical axis is labeled in pounds
**And** the plotted daily value is the pound conversion of 80 kg, not
80

### S10 — Chart does not write the log

**Given** the chart screen
**When** letters, Tab, or Space are pressed
**Then** no row is inserted or updated

### S11 — Monthly loss and daily deficit from trend

**Given** a 30-day month whose day-1 trend is 80.0 kg and last-day
trend is 79.0 kg
**And** `display_unit = "kg"`
**When** the chart is shown
**Then** Monthly Loss is 1.0 kg
**And** Daily Deficit is 257 calories
(1.0 kg → 2.2046 lb, × 3500 / 30 ≈ 257)

### S12 — Gain is a negative loss

**Given** a 30-day month whose day-1 trend is 79.0 kg and last-day
trend is 80.0 kg
**When** the chart is shown
**Then** Monthly Loss is −1.0 kg (or the equivalent in the display
unit)
**And** Daily Deficit is −257 calories

### S13 — No analysis without both endpoints

**Given** a month with no trend on day 1 or no trend on the last
plotted day
**When** the chart is shown
**Then** Monthly Loss and Daily Deficit are not shown as numbers

### S15 — Current month ends at today

**Given** today is 10 November 1990
**And** the November 1990 chart
**When** the chart is shown
**Then** there is no column for the 11th or later
**And** Daily Deficit divides by 10, not by 30

### S16 — Flat series still has a scale

**Given** a month whose plotted values are all the same
**When** the chart is shown
**Then** the vertical scale spans 4 lb in the display unit (2 × *P*)
**And** the process does not crash

### S14 — Loss follows display unit

**Given** the same 80.0 kg → 79.0 kg month as S11
**And** `display_unit = "lb"`
**When** the chart is shown
**Then** Monthly Loss is in pounds (about 2.2 lb)
**And** Daily Deficit is still 257 calories (the factor is always
pounds × 3500)

## Observing the chart

Tests drive `App.Update` / `View` and must not require a TTY.

A **discrete mark** is a non-space glyph that is not part of a
horizontal run of the same glyph (daily: `o` or equivalent). A
**connected path** is a run of glyphs that join left-to-right across
adjacent day columns (line-drawing or ASCII `-` / `/` / `\` / `|`).
When both occupy one cell, the cell contains the daily mark.

The plot is a grid: gutter (Y labels) then one character column per
day in the plotted span. Day *N*'s column is gutter width + *N*. The
plot body is at least 8 rows when there is data. Tests locate a day's
mark or path glyph in that column of those rows.

Colour on the chart is observed as in the colour-scheme spec: SGR for
the `weight` and `trend` roles. `NO_COLOR` is the same remainder set
(no chromatic SGR; bold and structure may remain).

The title, empty-state copy, Monthly Loss, and Daily Deficit are
visible (ANSI-stripped) strings. Loss is a signed decimal in the
display unit; deficit is a signed integer followed by a calorie word
(`cal` or `calories`).

## Functional requirements

-   **FR1.** `c` from the list opens the chart for the selected day's
    month. `c` from the month sheet, when not editing, opens the
    chart for the sheet's month. While editing on the sheet, `c` is
    text and does not open the chart. With an empty list, `c` opens
    the current calendar month.
-   **FR2.** Esc from the chart returns to the screen that opened it.
    If that screen is the month sheet, it shows the chart's current
    month.
-   **FR3.** The chart plots daily weight as discrete marks and trend as
    a connected path for each day of the plotted span. When both
    series share a cell, the daily mark is shown.
-   **FR4.** Daily marks use the `weight` colour role; the trend path
    uses the `trend` colour role.
-   **FR5.** Under `NO_COLOR`, chroma is off; marks and path stay
    distinct by glyph (daily mark wins a shared cell).
-   **FR6.** Days without a weight have no daily mark. Carried trend
    still plots on blank days in the plotted span, not after today.
-   **FR7.** Empty means no daily marks and no trend in the plotted
    span: empty-state copy, no scale, no analysis, no panic. A
    carry-only span is not empty (flat trend, loss 0, deficit 0).
-   **FR8.** `[` and `]` move the chart one month, as on the sheet,
    and update the sheet's year/month to match.
-   **FR9.** Vertical labels use `display_unit`. Values are converted
    with `units.FromKG`.
-   **FR10.** The chart screen does not create or update log rows.
-   **FR11.** Copy does not present the chart as medical advice and
    does not tell the reader what to eat.
-   **FR12.** When day 1 and the last plotted day both have a trend,
    the chart shows Monthly Loss (first − last, in `display_unit`)
    and Daily Deficit ((loss in pounds) × 3500 / days in the plotted
    span, whole calories). Neither value is stored in SQLite.
-   **FR13.** If either endpoint trend is missing, those two numbers
    are omitted.
-   **FR14.** The plot is one character column per day of the plotted
    span after a Y-label gutter, at least 8 rows when there is data,
    with day 1 and the last plotted day labeled. Tests locate day *N*
    in that column.
-   **FR15.** Ymin = min − *P*, Ymax = max + *P*, with *P* = 2 lb in
    `display_unit`. When min equals max, the span is 4 lb (or kg/st
    equivalent).

## Out of scope

-   Long-term (multi-month / yearly) charts
-   PDF of charts or of monthly sheets
-   Blank sheet printing
-   Goal lines, calorie bands, or an "Eat!" indicator
-   Weekly loss as a third number (Monthly Loss and Daily Deficit
    only)
-   Linear regression / last-seven-days analysis from the online
    tools; this slice is the pencil-and-paper first-and-last trend
-   Interactive cursor on the plot
-   New config keys
-   Meal planning

## Documentation

A how-to (open the monthly chart) and a reference key line for `c`.
Explanation can point at Signal and Noise. Tutorial may mention `c`
after the month sheet. Docs stay on the Diátaxis split.
