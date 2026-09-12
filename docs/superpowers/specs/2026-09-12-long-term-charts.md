# On-screen long-term weight charts

**Date**: 2026-09-12
**Status**: approved
**Issue**: [#38](https://github.com/jglueckstein/hdtools/issues/38)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md),
[`2026-09-11-monthly-chart-floats-sinkers.md`](2026-09-11-monthly-chart-floats-sinkers.md)
**Book**:
[Weight Monitoring — Long term charts](https://www.fourmilab.ch/hackdiet/e4/weightmonitor.html)

The monthly chart is one sitting: one calendar month, diamonds (marks)
tied to the trend with stems. `idea.md` also asks for long-term
charts. The book is specific: Excel's WEIGHT menu has **four** of
them, always ending at the latest data, plotting **daily weight and
trend as two lines** (thin blue, thick red), not the monthly
floats-and-sinkers form.

This slice adds that screen to the TUI. PDF and diet-plan / exercise
overlays are out of scope.

## Background

[Long term charts](https://www.fourmilab.ch/hackdiet/e4/weightmonitor.html)
lists four WEIGHT-menu charts, illustrated with 1989 data:

1.  **Quarterly** — last three months
2.  **Semiannual** — last six months
3.  **Annual** — last twelve months
4.  **Complete history** — the entire log, from the first weight

They show only daily weight and trend: weight a thin blue line, trend
a thick red line. Thickness still tells them apart in black and
white. The book notes that by hand you would skip daily weights
because they make the graph busy; Excel plots them because it is
free. This TUI plots both **when one column per day fits**; when
the span is bucketed (*D* > *W*), it follows the hand rule and
omits the daily path (O8).

A year of days will not fit one character per day on an 80-column
terminal. The four *kinds* stay as in the book; the X resolution
shrinks to the plot width (one column per day when the span fits,
equal-time buckets otherwise).

Loss and Daily Deficit reuse `dailylog.MonthlyBalance` on the first
and last trend of the plotted span. That is the pencil-and-paper
identity, not Excel's least-squares fit, and not medical advice.

## User stories

### US1 — Pick a long-term view

As a person with several months of logs, I want the book's four
long-term charts (quarter, half year, year, all history) so I can
see the same kinds of picture the Excel WEIGHT menu produced.

### US2 — See weight and trend as lines

As a person looking at a long-term chart, I want daily weight as a
thin line and trend as a thick line when the span fits one column
per day, distinct by thickness and by colour role, not as monthly
diamonds and stems. When the span is bucketed, I want the trend
path only.

### US3 — Read it without colour

As a person with `NO_COLOR` or a monochrome terminal, I want the two
lines to remain distinct by glyph weight (trend bold, daily not),
not only by colour.

### US4 — Empty or short history

As a person who has not yet got three months of data, I still want
the screen to open: plot what exists in that window, or an empty
state if the window has no trend, without crashing.

### US5 — Loss and deficit for the window

As a person looking at a long-term chart, I want Loss in my display
unit and Daily Deficit in calories from the first and last trend of
that window.

## Decisions

1.  **Four chart kinds, as in the book.**
    - Quarterly: day 1 of (end month − 2 months) through the last
      plotted day of the end month.
    - Semiannual: day 1 of (end month − 5 months) through that last
      day.
    - Annual: day 1 of (end month − 11 months) through that last
      day.
    - Complete history: the date of the first loaded log through
      that last day. If there are no logs, the window is empty.
    The **end month** is the month of the **latest loaded log**. If
    that month contains today, the last plotted day is today (same
    clip as the monthly chart). Never use a calendar month after the
    latest log: a 1990 database opened in 2026 still plots 1990, not
    an empty 2026 window (O1). Opening from the list, month sheet, or
    monthly chart does **not** change the end to a browsed historical
    month — these four charts are always last-N-months / all history
    of the loaded series. (Sliding a window that does not end at the
    latest log is out of scope.)
2.  **Last plotted day** of the end month: today if that month
    contains today; otherwise the last calendar day of a past end
    month. This screen does not select an end month after today
    (O5).
3.  **Key `l` opens the screen.** `c` stays the monthly chart.
    Default kind is **quarterly**. `l` works from the list, from the
    month sheet when **not** editing, and from the monthly chart.
    While editing on the sheet, `l` is text. Esc returns to the
    screen that opened it.
4.  **`[` / `]` cycle the four kinds**, in this order:
    quarterly → semiannual → annual → complete → quarterly.
    `[` goes backward. They do not change calendar month on the
    month sheet. Cue (O3): the title box always names the kind;
    this screen's help line is ``[ ] kind``, not ``[ ] month``.
5.  **Two lines, not floats and sinkers — when they fit.** When
    *D* ≤ *W*, daily weight is a **thin connected path** (the
    `weight` colour role, not bold) and trend is a **thick
    connected path** (the `trend` colour role, bold). Glyphs may be
    the same `-` / `/` / `\` family as the monthly trend path. When
    both series fall in the same cell, the **trend wins**. When
    *D* > *W*, **omit the daily path**; plot trend only (O8). There
    are **no** daily `o` marks and **no** `|` stems on this screen.
6.  **X resolution fits the terminal.** Let *D* be the number of
    days in the plotted span and *W* the plot width in columns
    (terminal width minus Y gutter, default **72** when the TUI has
    not seen a resize — tests use this default and must not require
    a TTY). If *D* ≤ *W*, one character column per day. If *D* >
    *W*, *W* equal-time buckets covering the span. Bucket *i*
    (0-based) is days `[floor(i·D/W), floor((i+1)·D/W))` of the
    span (O4). Each column uses the **last trend** in that bucket
    (carry-forward counts). Complete history uses the same rule, so
    a decade still fits.
7.  **Same loaded series.** Already-loaded, already-trended logs. No
    SQLite in the TUI. Carry-forward from before the window is
    included when `ApplyTrend` has seen those days.
8.  **Title box.** Centered above the plot. Text is the kind and the
    actual span, for example `Quarterly  September 1990–November 1990`
    or `Complete  April 1989–November 1990`. Not `hdtools — …`. Same
    Excel chrome as the monthly title box (yellow on blue, red
    border; 16-color names; not new `[colors]` roles). Under
    `NO_COLOR`, the text stays centered with no chromatic SGR.
    Display unit stays off the box.
9.  **Display unit and Y range.** Axis labels use `display_unit`.
    Storage remains kilograms. Auto Y is min and max of the series
    actually plotted (both when *D* ≤ *W*, trend only when *D* > *W*),
    with the monthly padding rule (1.0 display-unit band when min
    equals max).
10. **No editing.** Type, Tab, and Space do not write.
11. **Not medical.** Copy does not claim diagnosis or treatment and
    does not tell the reader what to eat.
12. **Loss and Daily Deficit.** `MonthlyBalance` on the first and
    last *trend* of the plotted span. Labels are `Loss` and `Daily
    deficit` (not `Monthly loss`). Divisor is calendar days in the
    span. Omit both numbers if either endpoint has no trend. Not
    stored in SQLite.
13. **Empty window.** No trend in the span: empty-state copy, no
    scale, no analysis, no panic. A window with only carried trend
    is not empty. The book says you cannot *print* a quarterly chart
    until you have three months; the TUI still **opens** all four
    kinds and shows empty or a short series rather than refusing
    `l`.

## Acceptance scenarios

Month names in titles are the **actual** span, not a fixture. Tests
that need a known calendar may freeze today at 10 November 1990 and
load logs in 1990; then an annual title names December 1989–November
1990.

### S1 — Open from the list

**Given** the daily list with at least one log
**When** `l` is pressed
**Then** the long-term chart is shown
**And** the kind is quarterly
**And** Esc returns to the list

### S2 — Open from the month sheet

**Given** the month sheet, not editing
**When** `l` is pressed
**Then** the long-term chart is shown
**And** Esc returns to the month sheet

### S2a — `l` while editing is text

**Given** the month sheet editing the note field
**When** `l` is pressed
**Then** the long-term chart does not open
**And** the letter `l` is in the note input

### S3 — Open from the monthly chart

**Given** the monthly chart
**When** `l` is pressed
**Then** the long-term chart is shown
**And** Esc returns to the monthly chart

### S4 — Cycle the four kinds

**Given** the long-term chart on quarterly
**When** `]` is pressed three times
**Then** the kind is semiannual, then annual, then complete
**When** `]` is pressed once more
**Then** the kind is quarterly again
**When** `[` is pressed
**Then** the kind is complete

### S5 — Two lines, no stems

**Given** a quarterly window with weighed days and trend
**When** the chart is shown
**Then** daily weight appears as a connected path
**And** trend appears as a connected path
**And** there is no daily mark glyph `o`
**And** there is no stem glyph `|`
**And** on a shared cell the trend glyph is shown

### S6 — Colour roles and thickness

**Given** a config with `weight = "green"` and `trend = "yellow"`
**And** a window with both series
**When** the chart is shown
**Then** the daily path is green and not bold
**And** the trend path is yellow and bold

### S7 — `NO_COLOR`

**Given** `NO_COLOR` is non-empty
**And** a window with both series
**When** the chart is shown
**Then** the view has no chromatic foreground or background
**And** both paths are still present
**And** trend remains bold and daily does not
**And** the title span is still present

### S8 — Empty list

**Given** no logs
**When** `l` is pressed from an empty list
**Then** the view says the span is empty
**And** Loss and Daily Deficit are not shown as numbers
**And** the process does not crash

### S8a — Logs exist, quarterly window has no trend

**Given** logs whose dates all lie outside the quarterly window
(for example logs only in June 1990, today 10 November 1990)
**When** `l` is pressed (kind quarterly)
**Then** the view says the span is empty
**And** Loss and Daily Deficit are not shown as numbers
**And** the process does not crash

### S9 — `[` / `]` do not change the month sheet

**Given** the November 1990 month sheet, then `l`, then `]` to
annual
**When** Esc is pressed
**Then** the month sheet still shows November 1990

### S10 — Display unit

**Given** `display_unit = "lb"`
**When** the chart is shown
**Then** the vertical axis is labeled in pounds
**And** plotted values are converted with `units.FromKG`

### S11 — Chart does not write the log

**Given** the long-term chart
**When** letters, Tab, or Space are pressed
**Then** no row is inserted or updated

### S12 — Loss and deficit

**Given** a window whose first-day trend is 80.0 kg and last-day
trend is 79.0 kg
**And** that span is 90 days
**And** `display_unit = "kg"`
**When** the chart is shown
**Then** Loss is 1.0 kg
**And** Daily Deficit is 86 calories
(1.0 kg → 2.2046 lb, × 3500 / 90 ≈ 86)

### S13 — Title names kind and actual span

**Given** today is 10 November 1990
**And** the long-term chart on annual
**When** the chart is shown
**Then** the title contains `Annual`
**And** the title contains `December 1989` and `November 1990`
**And** that line does not contain `hdtools —`

### S14 — Complete history starts at the first log

**Given** the earliest log is 15 April 1989
**And** today is 10 November 1990
**And** the kind is complete
**When** the chart is shown
**Then** the title contains `April 1989` and `November 1990`
**And** the span's first day is 15 April 1989

### S15 — Wide span is bucketed

**Given** an annual window (~365 days) and the default plot width 72
**When** the chart is shown
**Then** the plot has 72 columns (not ~365)
**And** the plot body is at least 8 rows when there is data
**And** there is no daily-weight path (only trend; *D* > *W*)

### S16 — Quarterly clips at today

**Given** today is 10 November 1990
**And** the kind is quarterly
**When** the chart is shown
**Then** the last plotted day is 10 November 1990, not 30
**And** the window starts 1 September 1990

### S17 — Past database in a later calendar year

**Given** today is 12 September 2026
**And** the latest log is 10 November 1990
**When** `l` is pressed (kind quarterly)
**Then** the title contains `September 1990` and `November 1990`
**And** the view is not the empty-span state

## Observing the chart

Tests drive `App.Update` / `View` and must not require a TTY.

A **connected path** is a run of `-` / `/` / `\` across adjacent
columns. Daily and trend paths are distinguished by SGR (`weight` vs
`trend`) and by bold (trend only). Shared cells contain the trend
glyph. Tests must not require `o` or `|` on this screen.

Default plot width is 72 columns plus gutter when no resize has
been seen.

The title, empty-state copy, Loss, and Daily Deficit are
ANSI-stripped strings. Loss is a signed decimal in the display unit;
deficit is a signed integer followed by `cal` or `calories`.

## Functional requirements

-   **FR1.** `l` opens the long-term chart (default quarterly) from
    the list, from the month sheet when not editing, and from the
    monthly chart. While editing on the sheet, `l` is text.
-   **FR2.** Esc returns to the screen that opened it. The month
    sheet's month is unchanged by cycling kinds. Long-term help is
    ``[ ] kind``.
-   **FR3.** Four kinds: quarterly (3 months), semiannual (6),
    annual (12), complete (first log through last plotted day).
    `[` / `]` cycle that list, wrapping.
-   **FR4.** When *D* ≤ *W*, daily weight is a thin (not bold)
    connected path in the `weight` role; trend is a bold connected
    path in the `trend` role; trend wins a shared cell. When *D* >
    *W*, plot trend only. No `o` marks, no `|` stems.
-   **FR5.** Under `NO_COLOR`, chroma is off; plotted paths and the
    title remain; trend stays bold.
-   **FR6.** Empty means no trend in the span: empty-state copy, no
    scale, no analysis, no panic.
-   **FR7.** Columns: one per day if *D* ≤ *W*, else *W* buckets
    (*W* = 72 until a resize). Bucket *i* is days
    `[floor(i·D/W), floor((i+1)·D/W))`.
-   **FR8.** Vertical labels use `display_unit` via `units.FromKG`.
-   **FR9.** The screen does not create or update log rows.
-   **FR10.** Copy does not present the chart as medical advice.
-   **FR11.** When both endpoints have a trend, show Loss and Daily
    deficit via `MonthlyBalance` over the span's day count.
-   **FR12.** The title box names the kind and the actual start and
    end months, with the monthly title-box chrome.

## Out of scope

-   Custom start and end dates
-   Floats-and-sinkers (marks and stems) on the long-term plot
-   Exercise-rung or diet-plan overlay
-   Least-squares straight-line fit
-   PDF of long-term or monthly charts
-   Interactive cursor / open a month from a column
-   New `[colors]` keys
-   Meal planning
-   Sliding a historical window that does not end at current data

## Documentation

A how-to (open the long-term charts, cycle kinds) and a reference
key line for `l`. Tutorial may mention `l` after `c`. Docs stay on
the Diátaxis split.
