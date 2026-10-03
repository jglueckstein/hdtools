# High-resolution on-screen charts

**Date**: 2026-10-02
**Status**: approved
**Objections**:
[`high-resolution-charts.md`](../objections/high-resolution-charts.md)
(O1–O3, O5, O6 accepted; O4 rejected). Code-mode:
[`high-resolution-charts-code.md`](../objections/high-resolution-charts-code.md)
(O1–O3 accepted).
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Amends**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md)
Decision 4, Decision 7, S3, the Observing section, FR3, FR5, and
FR14 (paint and shared-cell collision);
[`2026-09-12-long-term-charts.md`](2026-09-12-long-term-charts.md)
Decision 5, Decision 6, S5, the Observing section, and FR4;
[`2026-09-11-monthly-chart-floats-sinkers.md`](2026-09-11-monthly-chart-floats-sinkers.md)
Decision 2 (a shared cell is a shared dot; one color per rune).
**Follows**:
[`2026-09-12-chart-y-range.md`](2026-09-12-chart-y-range.md),
[`2026-09-12-long-term-x-labels.md`](2026-09-12-long-term-x-labels.md).
PDF monthly charts stay as in
[`2026-09-12-pdf-monthly-charts.md`](2026-09-12-pdf-monthly-charts.md).

The book's Excel charts, and Hacker's Diet Online where a terminal
can go further, are why this slice exists. They are not a second
test. The contract is the paint below: Unicode Braille inside the
character cells the charts already use, so a curve is not one dot
per day in eight rows. Data identity is unchanged. Samples are the
plotted weights and the trend. Nothing is splined or filled. PDF
stays vector.

Until this spec ships, the amended sentences still describe the
paint. This spec replaces them when it ships. It does not replace
the Y range, the clip, which series are drawn, or the long-term
bucket rule. No chart library is added. The plot stays in the TUI.

## Decisions

1.  **The glyph is Braille.** Each plot cell is one Unicode Braille
    rune, U+2800 through U+28FF. The rune is a 2×4 dot grid. Tests
    read the dots from `View()` by the bits below. No other glyph
    set is the plot. A mapping that cannot tell the positions in
    the scenarios apart does not meet this spec.

    | Dot | Bit | Place in the cell |
    | --- | --- | --- |
    | 1 | `0x01` | left, top |
    | 2 | `0x02` | left, second |
    | 3 | `0x04` | left, third |
    | 7 | `0x40` | left, bottom |
    | 4 | `0x08` | right, top |
    | 5 | `0x10` | right, second |
    | 6 | `0x20` | right, third |
    | 8 | `0x80` | right, bottom |

    Top to bottom, the left column is dots 1, 2, 3, 7 and the right
    column is dots 4, 5, 6, 8. Dot row 0 is the top of the cell.
2.  **A day stays one slot.** On the monthly chart, day *N* of the
    plotted span is still the *N*th character column after the Y
    gutter. That column is one Braille cell wide. Its left dots and
    its right dots are the two horizontal paint positions, and no
    dot belongs to two days. They are not extra terminal columns.
3.  **Vertical bins.** The Y span is still (min − *P*) through
    (max + *P*) from the Y-range spec. Let *R* be the character
    rows of the plot body and *N* = *R* × 4. *R* is at least 8, so
    *N* is at least 32. Scenarios that name a bin use *R* = 8 and
    *N* = 32. A value *v* has bin

    -   *N* − 1 when *v* ≥ Ymax
    -   otherwise clamp(floor((*v* − Ymin) / (Ymax − Ymin) × *N*),
        0, *N* − 1)

    The bin group is floor(bin / 4). Group 0 sits at the bottom of
    the plot, toward Ymin. Group *R* − 1 sits at the top. Inside a
    cell, bin mod 4 = 0 is the bottom dot and bin mod 4 = 3 is the
    top dot (Decision 1's dot row 0). Two values share a vertical
    paint position only when they have the same bin. The coarse
    grid used in S4 is the same function with 8 bins, one per
    character row.
4.  **What sets a dot.** A daily mark sets the left-column dot at
    its bin. A stem sets the left-column dots strictly between the
    mark's bin and that day's trend bin. Equal bins draw no stem.
    The trend path sets both the left and the right dot at the
    trend bin in that day or bucket column, so the second
    horizontal position is visible. On a shared dot the monthly
    daily mark wins and the long-term trend wins. The stem uses the
    mark's horizontal column.
5.  **One color per cell.** A cell has one foreground color. Dots
    do not. On the monthly chart, a cell that contains a stem dot
    uses the stem green (chart chrome, not a `[colors]` role),
    including when that cell also holds a mark or a trend dot. The
    stem is one colour along its length. Otherwise a mark dot uses
    the `weight` role, and otherwise the `trend` role. On S2 day 2
    the stem shares the trend's cell and that rune is stem green.
    The mark sits in the next cell and keeps the weight role. A
    day with no weigh-in keeps the trend role. On the long-term
    chart, a trend dot uses the `trend` role and is bold;
    otherwise the `weight` role, not bold. Under `NO_COLOR`
    chromatic color is off and the rune remains. Bold stays on
    the trend role, including a monthly carry day, and stays off
    a coincident weigh-in. Those two days share one mask. The
    same mask means the same dots. A filled mark cell and a pure
    stem cell are the same picture and get no third style bit.
    A shared cell is one rune, not three.
6.  **Long-term fit is unchanged.** *D*, *W* (default 72 when the
    TUI has not seen a resize), and the equal-time buckets stay as
    in the long-term spec. Finer paint does not add character
    columns and must not push a span that fits (*D* ≤ *W*) into
    buckets. Each day column or bucket column is one Braille cell
    and follows Decisions 3 and 4. When the span is bucketed, the
    daily path is still omitted, and there are still no monthly
    marks or stems.
7.  **Identity stays.** Title box, which series are plotted, Y
    margin, clip (the current month and a long-term end month that
    contains today both end at today), Loss and Daily Deficit,
    `[colors]` roles, and the 16-color floor are unchanged. No new
    `[colors]` roles. Empty means what the existing specs say:
    empty-state copy, no scale, no analysis, no panic. Keys and
    help lines are unchanged.
8.  **PDF is not this paint.** `p` and the chart-PDF command still
    write the vector page from the PDF spec. This slice does not
    redraw that page and does not change where the file is written.

## User stories

### US1 — A monthly chart that is not a coarse grid

As a person looking at a monthly chart, I want marks, stems, and the
trend drawn as Braille dots inside each day column, so a small
change in weight is visible.

### US2 — The same fineness on long-term lines

As a person looking at a long-term chart, I want that Braille paint
on the lines that are already drawn, including when the span is
bucketed, without changing which days or series appear.

### US3 — Still readable without color

As a person with `NO_COLOR` or a monochrome terminal, I want the
Braille still there, and the long-term trend still bold, so the
chart can be read without color.

### US4 — Numbers and the PDF stay put

As a person who already reads Loss, Daily Deficit, and the PDF, I
want those unchanged by the finer on-screen paint.

## Acceptance scenarios

Tests drive `App.Update` / `View` and must not require a TTY. Today
may be frozen. A paint position is a Braille dot from Decision 1,
not a count of terminal columns. Named bins use *R* = 8.

November 1990 has 30 days. September has 30, October 31.

*P* for kilograms is 2 × 0.45359237. For the fixtures whose plotted
minimum is 80.00 kg and maximum is 82.00 kg, Ymin is 80 − *P* and
Ymax is 82 + *P*.

### S1 — Each November day uses both Braille columns

**Given** today is 15 December 1990
**And** November 1990 has a weight on day 1 and day 30
**When** the monthly chart for November 1990 is shown
**Then** the plot has 30 day columns in order after the Y gutter
**And** each column is one Braille rune
**And** each of those runes has at least one left-column dot and
one right-column dot set
**And** no dot belongs to two days

### S2 — Two close marks do not share a vertical dot

**Given** today is 15 December 1990
**And** `display_unit` is `kg`
**And** there are no logs before November 1990
**And** November 1990 has daily weights 80.00 kg on day 1, 80.55 kg
on day 2, 80.85 kg on day 3, and 82.00 kg on day 30
**When** the monthly chart is shown
**Then** the plotted minimum is 80.00 kg and the plotted maximum is
82.00 kg
**And** the day 2 mark is bin 12 and the day 3 mark is bin 14
**And** those marks do not share a vertical paint position

Trend on those weights stays inside 80.00 kg through 82.00 kg (day 1
trend 80.0, day 2 trend 80.1, day 3 trend 80.2, day 30 trend 80.4).

### S3 — Eight rows, four dots each

**Given** the monthly chart in S2
**When** it is shown
**Then** the plot body is 8 character rows
**And** each of those rows is one Braille cell with 4 vertical dots
**And** the Y span therefore has 32 vertical paint positions

### S4 — A stem survives inside one character row

**Given** today is 15 December 1990
**And** `display_unit` is `kg`
**And** there are no logs before November 1990
**And** November 1990 has daily weights 80.00 kg on day 1, 80.50 kg
on day 2, and 82.00 kg on day 30
**When** the monthly chart is shown
**Then** day 1's mark and trend are both bin 7, and that day has no
stem dot
**And** day 2's trend is 80.1 kg at bin 8
**And** day 2's mark is bin 11
**And** bins 8 and 11 are both in bin group 2
**And** the left column of that day has stem dots at bins 9 and 10

On 8 equal bins of the same padded span, 80.50 kg and 80.1 kg share
one bin, and a one-dot-per-row plot would omit the stem. This slice
still draws it.

### S5 — Monthly chart without color

**Given** `NO_COLOR` is set
**And** the monthly chart in S2
**When** it is shown
**Then** the view has no chromatic color
**And** day 2's mark and that day's trend are in different character
rows, so they are different cells
**And** both cells still contain their Braille dots
**And** the title still shows the month and year

### S6 — Long-term columns stay columns

**Given** today is 10 November 1990
**And** the latest loaded log is on 10 November 1990
**And** the TUI is at the default plot width (*W* = 72)
**When** the quarterly chart is shown
**Then** the span is 1 September through 10 November (71 days)
**And** 71 ≤ 72, so each of those days has its own character column
**And** the 10 November column is one Braille cell whose trend path
sets a left dot and a right dot
**And** the chart does not use fewer day columns in order to make
room for that paint

### S7 — A bucketed long-term chart stays bucketed

**Given** today is 15 December 1990
**And** the latest loaded log is on 30 November 1990
**And** *W* = 72
**When** the quarterly chart is shown
**Then** the span is 1 September through 30 November (91 days)
**And** 91 > 72, so the plot uses 72 equal-time buckets, not one
column per day
**And** the daily path is omitted
**And** the trend path uses Decision 3
**And** the bucket that contains 30 November is one Braille cell
with a left dot and a right dot set for the trend

### S8 — Long-term lines without color

**Given** `NO_COLOR` is set
**And** the quarterly chart in S6 (both paths are drawn)
**When** it is shown
**Then** the view has no chromatic color
**And** a cell that contains a trend dot is bold
**And** a cell that contains only the daily path is not bold

### S9 — Empty month

**Given** a month with no daily mark and no trend in the plotted span
**When** the monthly chart is shown
**Then** the empty state is shown, with no scale and no panic

### S10 — The current month still ends at today

**Given** today is 10 November 1990
**And** November 1990 has a weight on day 1 and day 20
**When** the monthly chart for November 1990 is shown
**Then** the last day column is 10 November, not 20 or 30
**And** no Braille dot is placed after today

### S11 — A coincident weigh-in without color

**Given** `NO_COLOR` is set
**And** today is 15 December 1990
**And** November 1990 has one weight, 80 kg on day 1
**When** the monthly chart is shown
**Then** the view has no chromatic color
**And** day 1 and day 2 paint the same Braille mask
**And** day 1 is not bold
**And** day 2 is bold

## Observing the chart

Tests drive `App.Update` / `View` and must not require a TTY.

Strip ANSI for structure. A plot cell is one Braille rune. Its eight
bits are Decision 1. A missing plot cell is not a Braille rune and
has no dots. Bin numbers are Decision 3 with *R* = 8 unless a
scenario says otherwise.

Color is observed as in the color-scheme spec: one foreground SGR
for the whole rune. `NO_COLOR` leaves no chromatic SGR. Bold stays
on a trend cell, including a monthly carry day, and stays off a
coincident weigh-in that shares that cell's mask.

The title, empty-state copy, Loss, and Daily Deficit stay the
ANSI-stripped strings the monthly and long-term specs already
require.

## Functional requirements

-   **FR1.** The monthly plot is one Braille column per day of the
    plotted span after the Y gutter, in order. Each column has a
    left dot column and a right dot column. No dot belongs to two
    days. The columns are not extra terminal columns.
-   **FR2.** *R* ≥ 8 character rows when there is data. Each row has
    4 vertical dots, so *N* = *R* × 4 ≥ 32. Bins are Decision 3,
    half-open on the padded span, with *v* ≥ Ymax in the last bin.
-   **FR3.** Two plotted values share a vertical paint position only
    when Decision 3 gives them the same bin. S2 is the worked
    example: bins 12 and 14 at *R* = 8.
-   **FR4.** A mark sets the left dot at its bin. A stem sets the
    left dots strictly between the mark and that day's trend. Equal
    bins draw no stem. The trend path sets both horizontal dots at
    the trend bin. The monthly mark wins a shared dot. S4 is the
    worked example: stem dots at bins 9 and 10.
-   **FR5.** A cell has one foreground color, chosen by Decision 5.
    A stem dot makes the cell stem green. Otherwise mark, then
    trend. Under `NO_COLOR` the Braille remains and chromatic
    color is off. Bold stays on the trend role, including a
    monthly carry day, and stays off a coincident weigh-in that
    shares its mask. The same mask means the same dots. A
    stem-filled mark and a pure stem get no extra style bit.
    Distinctness does not require three runes in one cell.
-   **FR6.** Long-term *D* ≤ *W* still assigns one character column
    per day, and *D* > *W* still assigns *W* buckets and omits the
    daily path. Each of those columns is one Braille cell. Decision
    3 applies to all four long-term kinds. The trend wins a shared
    dot.
-   **FR7.** Title box, series, Y margin, today-clip, Loss, Daily
    Deficit, `[colors]`, and the 16-color floor are unchanged. No
    new `[colors]` role. Keys and help lines are unchanged. No
    spline and no area fill.
-   **FR8.** An empty plotted span shows the empty state, no scale,
    and does not panic. The fine grid is not required when there is
    nothing to place.
-   **FR9.** PDF generation is unchanged: same picture, same path
    rules, same command. The Braille grid is not the PDF.

## Out of scope

-   Kitty, Sixel, iTerm2, and any other image protocol
-   ntcharts, Pulse, ChartGo, bubble-plot, hl-tickers, and any other
    chart widget
-   Half-block glyphs as the plot
-   Splines and fills under the line
-   Redrawing PDF charts, filled log sheets, or blank sheets
-   Where PDF files are written
-   Extra daily-log fields
-   Changing which days or series are plotted, the bucket formula, or
    the Y-range arithmetic
-   Goto Today, and any change to keys or help lines
