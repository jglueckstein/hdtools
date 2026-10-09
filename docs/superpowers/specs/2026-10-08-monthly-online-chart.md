# Monthly Online chart

**Date**: 2026-10-08
**Status**: approved
**Objections**:
[`monthly-online-chart.md`](../objections/monthly-online-chart.md)
(O1–O3 accepted)
**Issue**: [#77](https://github.com/jglueckstein/hdtools/issues/77)
**Backlog**: [`idea.md`](../../../idea.md) (TUI)
**Slice**: S2 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)

The on-screen monthly chart becomes the Hacker's Diet Online
picture. The trend is a Braille line of the derived trend. Each
weighed day is an open diamond, and a green stem joins that day's
weight to the trend on that same day. A plain heading names the
month and year. Whole-number ticks run down the left in the display
unit, one step beyond the data, and day labels run along the bottom.
The plot is 62 columns and 16 rows for every month. It does not
follow the terminal, and days after today stay empty slots on that
width.

This is the monthly screen only. Monthly Loss and Daily Deficit stay
the pencil identity. The long-term chart stays the Braille-cell plot.
The PDF keeps the vector page it has now. This draft does not revise
the earlier chart specs; those stay the history of the picture this
slice replaces, and their status stays approved.

The picture is drawn with ntcharts v2
(`github.com/NimbleMarkets/ntcharts/v2`), the current v2 release. No
patch release is pinned. The Go version stays 1.27.0.

## User stories

### US1 — See the month as the Online picture

As a person reading a monthly chart, I want a Braille trend, an open
diamond on each weighed day, and a green stem from that weight to
that day's trend, so I can see the day's weight against the trend
the way the Online chart does.

### US2 — Read the scale in the display unit

As a person who shows weight in kilograms, pounds, or stone, I want
whole-number ticks in that unit, one step beyond the data, so the
monthly scale is not a fixed two-pound margin.

### US3 — Keep colour without new settings

As a person with a colour config or `NO_COLOR`, I want the trend and
the diamond to follow the colours I already have, and the stem to
stay green, so I do not add config keys to read the chart.

### US4 — Leave the numbers and the other pictures

As a person who already reads Monthly Loss, Daily Deficit, the
long-term chart, and the PDF, I want those left as they are, so this
change is only the on-screen monthly picture.

### US5 — Compare months on one fixed width

As a person reading one month and then another, I want every month
on the same 62-column plot, each calendar day sharing that width and
days after today left empty, so a short month is not a narrower
picture and the current month does not draw the future.

## Decisions

1.  **Monthly screen only.** The on-screen monthly chart is the
    Online picture below. The long-term chart stays the Braille-cell
    plot, including its fixed two-pound margin and its four windows.
    The PDF stays the vector chart it is, including its title. Keys
    already on the chart stay. List, month sheet, day form, paste,
    field colours, and the space bar are unchanged.
2.  **The series are the derived trend and the weighed days.** The
    trend is the same derived series the month sheet shows. It is not
    stored, and it is not a straight line through the raw weights.
    Drawing it needs the full chronological log, so a month keeps the
    previous month's trend. A carried trend still draws on a day with
    no weight. It does not draw after today.
3.  **Open diamond.** "Open" means the glyph U+25C7. Every plotted
    day with a weight gets one. A day with no weight gets none. A
    filled diamond (U+25C6) is not used. Workout does not change the
    glyph, does not add a mark, and does not colour the diamond.
    There is no flag and no exercise-rung line.
4.  **The stem joins one day.** For a weighed day, the stem is the
    line from that day's weight to the trend on that same day. Float
    means the weight is above the trend. Sinker means it is below.
    The stem is green either way. The marks of the stem are not
    Braille and are not the diamond, so the three stay distinct
    without colour. When the weight and the trend are the same value,
    there is no stem. When they differ and land on different rows,
    the stem is required. The 18 kg gap and the 9 kg gap in S3 land
    on different rows. When two different values land on the same
    row, there is no stem. A 1 kg gap on a tall scale is not a
    separate guarantee. Draw the trend, then the stems, then the
    diamonds, so the diamond sits on top of the stem.
5.  **Plain heading.** The heading is the month name and the year
    (`November 1990`). It uses the existing `title` role: bold,
    default cyan, and a `[colors]` title entry changes it. It has
    no border and no filled background. It is not the old
    yellow-on-blue box. The words are centered across the 67-column
    plot line, with space on both sides. The heading line is not a
    plot row. `NO_COLOR` keeps the words and the bold and drops the
    chromatic colour. The line that names the database and the
    display unit stays. The help line stays.
6.  **Display-unit scale.** The canvas is the display unit, not the
    stored kilograms. Plotted values are each day's weight, when it
    has one, and each day's trend, when it has one. A stem does not
    push the scale. The step is 1 kg, 1 lb, or one pound on the stone
    scale. One pound is 1/14 stone. The scale runs from exactly one
    step below the minimum plotted value to exactly one step above
    the maximum. Tick labels are the multiples of that step that fall
    inside the scale, including a multiple that lands on an end.
    Kilogram and pound labels are those whole numbers, with no
    decimal. Stone labels are one decimal of stones, the same
    rounding a stone weight already uses on screen, not a
    "N st M lb" pair. Two pound-steps that round to the same decimal
    are still two labels. The high end tick is at the top. The low
    end tick is at the bottom. Those two end labels are always shown.
    The ranges in S6 and S7 show every tick label. A tick never
    shares a row with another tick. The plot is 16 rows
    (Decision 13). S6 and S7 have fewer ticks than rows, so every
    tick is shown. S8 has more ticks than rows, so the end ticks
    stay and at least one tick between them is omitted.
7.  **Day labels.** Day-of-month numbers run under the plot. The
    first plotted day and the last plotted day are always labeled.
    Days between may be labeled. A day that is not plotted is not
    labeled. The current calendar month's last plotted day is today.
8.  **Colours stay on the roles that already exist.** The trend uses
    the `trend` role. The default is red, and the role stays bold. A
    `[colors]` trend entry overrides it. The diamond uses the
    `weight` role. The default is blue, and the role stays not bold.
    A `[colors]` weight entry overrides it. There is no `diamond`
    key. The stem stays green chrome, not bold. There is no `stem`
    key. An unknown key is ignored. An invalid colour is ignored and
    the default is used. `NO_COLOR`, when the variable is non-empty,
    drops chromatic colour and keeps the Braille trend, the diamond,
    the stem, and the trend's bold. The program does not decide this
    by whether a terminal is attached. Bold, reverse, and `>`
    elsewhere stay.
9.  **16-color is the floor of colour written out.** The app does
    not downsample the view text. A hex colour written for a
    truecolor profile is `38;2`. The same hex written for a 16-color
    profile is an indexed colour and is not `38;2`. The view text may
    stay full fidelity. `NO_COLOR`, bold, reverse, and `>` are read
    on that text, not after a second conversion.
10. **The caption stays the pencil identity.** Monthly Loss is the
    first trend minus the last trend of the plotted span, in the
    display unit. Daily Deficit is that loss in pounds times 3500,
    divided by the number of days in the plotted span, rounded to a
    whole number of calories. It is not the last seven trend values,
    and it is not a straight-line fit. The numbers are not stored.
11. **The library line.** The program requires
    `github.com/NimbleMarkets/ntcharts/v2` and still requires
    `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, and
    `charm.land/bubbles/v2`. It does not require ntcharts v1 or the
    v1 modules `github.com/charmbracelet/bubbletea`,
    `github.com/charmbracelet/lipgloss`, or
    `github.com/charmbracelet/bubbles`. The Go version stays 1.27.0.
    The program is still started with the screen model alone: no
    alternate screen and no mouse. The chart has no grid, no legend,
    no zoom, and no pointer highlight.
12. **Fixed width.** The monthly plot is 62 terminal columns, not
    counting the tick labels. It does not grow or shrink with the
    terminal. Every calendar day of that month gets one contiguous
    share of those 62 columns, in order from the left, with no inset.
    February and a 31-day month are the same plot width. A 31-day
    month is exactly two columns per day. Other month lengths divide
    62 as evenly as whole columns allow. Which days receive a
    remainder column is not fixed. Slot boundaries depend only on how
    many days the month has, not on today. A November day is the same
    width on the 10th as in a finished November. The line, the stems,
    and the diamonds stop at today. Days after today are empty slots:
    no trend, no stem, and no diamond. They still occupy their share
    of the 62 columns. The tick gutter is five columns and stays
    fixed, so the plot's left edge does not move. Four columns hold
    the tick, right-aligned. The fifth is the hyphen the chart
    already uses. The monthly gutter is not the long-term `%5.1f-`
    form. `  79-`, ` 121-`, and `12.6-` fit. A plot row is 67 visible
    columns, five plus 62. Width is visible columns, not bytes,
    because the plot may contain colour. The long-term chart keeps
    its current width and its `%5.1f-` gutter. This slice does not
    make that chart 62 or 67 columns. An empty month stays the empty
    state: no scale and no 62-column plot. A carry-only month has a
    trend, so it uses the 62-column plot.
13. **Fixed height.** The monthly plot is 16 rows. It does not grow
    or shrink with the terminal. An empty month has no plot, so it
    has no 16-row canvas. A carry-only month uses the 16-row plot.
14. **Published sentences.** This slice amends the sentences it
    makes false. `docs/how-to/monthly-chart.md` drops the month-year
    box and the ±2 lb margin, and describes the plain heading and
    the one-step scale. `docs/how-to/long-term-chart.md` keeps the
    ±2 lb margin and stops saying that margin matches the monthly
    chart. `docs/how-to/monthly-chart-pdf.md` keeps the existing
    vector page, including its title box, and stops saying the PDF
    is the on-screen chart. The `idea.md` change record for this
    draft says the monthly screen drops the box and the shared
    margin, the long-term screen keeps both until #78, and the PDF
    keeps the old page. Older change records stay the history of
    those slices. This slice does not redraw the PDF or the
    long-term chart.

## Acceptance scenarios

November 1990 is the fixture month unless a scenario says otherwise.
Weights in the log are kilograms. The display unit is kilograms
unless the scenario says otherwise. A finished month is read with
today on 15 December 1990, so the plot runs through day 30. The
derived trend of a new log starts at the first weight. After that,
each new weight moves the trend one tenth of the way toward that
weight, with the tenth rounded to one decimal of a kilogram before
it is added. Later blank days keep that trend.

Tests drive the screen from messages and read the view. They do not
need a terminal. A test that freezes today or changes directory does
not run beside other tests. A test that needs colour clears
`NO_COLOR`. The database file is temporary, not an in-memory shared
database.

An open diamond is U+25C7. A filled diamond is U+25C6. The trend is
Braille (U+2800 through U+28FF). The stem is a plot mark that is
neither of those. With colour on, stem marks are green, the default
diamond is blue, and the default trend is red. The diamond's own
cell is the weight colour, not green, when the weight role is not
green. Read the trend's red and its bold on a blank day, because the
stem is drawn after the trend and may cover the trend at the one
point where they meet. Left-axis labels are the numbers beside the
plot. Day labels are the day numbers under the plot. A plot row is a
scale line, not the heading, the database line, the day-label line,
the caption, or the help line. The first five visible columns of a
plot row are the tick gutter. The next 62 are the plot. Plot column
1 is the first column after the gutter. Width is visible columns,
not bytes. A plot column has none of the series when it has no
diamond, no stem, and no Braille. The caption is the loss line. Do
not treat one of those as another.

Profile-dependent colour is the bytes written through a colour
profile. `NO_COLOR` is read on the view text.

### S1 — Diamond and Braille trend

**Given** a finished November with 80 kg on day 1 and 82 kg on day 30
**And** no earlier logs
**When** the monthly chart is shown
**Then** there are two open diamonds, day 1's then day 30's, from
left to right
**And** the trend is Braille across the plotted days, including the
days between the two weights
**And** the view has no filled diamond

### S2 — A blank day has trend and no diamond

**Given** the month in S1
**When** the chart is shown
**Then** a day with no weight, such as day 2, has Braille trend
**And** it has no diamond and no stem

### S3 — The stem joins that day, and the diamond is on top

**Given** a finished November with 80 kg on day 1 and 100 kg on day 2
**And** no earlier logs
**When** the chart is shown
**Then** day 1's weight and trend are the same, and that day has no
stem
**And** day 2's weight is above the trend (a float), and a stem joins
them
**And** the stem marks are green, are not Braille, and are not the
diamond
**And** the stem occupies a row between the diamond and the trend
**And** the diamond on day 2 is blue, not green

**Given** a finished November with 80 kg on day 1 and 70 kg on day 2
**And** no earlier logs
**When** the chart is shown
**Then** day 2's weight is below the trend (a sinker)
**And** the stem is green, the same green as the float
**And** the stem occupies a row between the diamond and the trend
**And** the diamond is blue and sits on top of the stem

### S4 — Workout does not change the diamond

**Given** a finished November with a weighed day whose workout flag
is set, and another weighed day whose flag is clear
**When** the chart is shown
**Then** both days show the same open diamond in the weight colour
**And** the workout day has no extra mark and no filled diamond

### S5 — Plain heading

**Given** the November chart
**When** it is shown with colour on
**Then** the heading names November 1990
**And** it is bold and uses the title colour, cyan by default
**And** it has no border and no filled background
**And** it is not yellow on blue
**And** it does not say "hdtools"
**And** the words are centered across the plot, with space on both
sides

**Given** `[colors]` sets `title` to magenta
**When** the chart is shown
**Then** the heading is magenta
**And** the trend stays red

**When** `NO_COLOR` is non-empty
**Then** the same month and year are still there
**And** the heading is still bold
**And** the view has no chromatic colour

### S6 — Kilogram and pound ticks

The weighed values below are the minimum and the maximum plotted, and
the trend stays between them. The high tick is the top label. The low
tick is the bottom label. The plot is 16 rows. These ranges have
fewer ticks than rows, so every listed tick is shown, each on its
own row.

**Given** a finished month in kilograms whose plotted values run from
80 to 82
**When** the monthly chart is shown
**Then** the left-axis labels are 79, 80, 81, 82, and 83
**And** the 79 tick's gutter is the five columns `  79-`

**Given** a finished month in kilograms whose only plotted value is 80
**When** the monthly chart is shown
**Then** the labels are 79, 80, and 81
**And** the scale is not a four-pound band

**Given** a finished month in kilograms whose plotted values run from
80.4 to 82.4
**When** the monthly chart is shown
**Then** the labels are 80, 81, 82, and 83
**And** 79 and 84 are not axis labels

**Given** `display_unit` is pounds
**And** the plotted values run from 170 lb to 175 lb
**When** the monthly chart is shown
**Then** the labels are the whole numbers 169 through 176
**And** they have no decimal part

### S7 — Stone ticks are one pound, labeled as stones

**Given** `display_unit` is stone
**And** a finished month whose plotted values run from 12.0 st to
12.5 st
**When** the monthly chart is shown
**Then** the ticks are one pound apart, which is 1/14 stone
**And** from the bottom to the top the labels are 11.9, 12.0, 12.1,
12.1, 12.2, 12.3, 12.4, 12.4, 12.5, and 12.6
**And** 12.1 appears twice, and 12.4 appears twice
**And** no label is a stones-and-pounds pair such as "12 st 0 lb"
**And** the 12.6 tick's gutter is the five columns `12.6-`

### S8 — Crowded tick labels keep the ends

**Given** a finished month in kilograms whose plotted values run from
40 to 120
**When** the monthly chart is shown
**Then** the bottom end label is 39 and the top end label is 121
**And** the 121 tick's gutter is the five columns ` 121-`
**And** the plot is 16 rows, fewer than the step ticks, so at least
one tick between the ends is omitted
**And** no two tick labels share a row
**And** 38 and 122 are not axis labels

### S9 — Day labels, and the current month ends today

**Given** a finished November
**When** the monthly chart is shown
**Then** the day labels under the plot include 1 and 30
**And** the plot is 62 columns and the day slots use that full width

**Given** today is 10 November 1990
**And** November has 80 kg on day 1, 70 kg on day 2, and 60 kg on
day 15
**When** the monthly chart is shown
**Then** there are two diamonds, not three
**And** the day labels include 1 and 10
**And** day 15 and day 30 are not labeled and are not plotted
**And** each plot row is 67 visible columns
**And** plot columns 23 through 62 contain no diamond, no stem, and
no trend
**And** Monthly Loss is 1.0 kg
**And** Daily Deficit is 772 calories, not 257

772 is the pencil figure for a 1.0 kg loss over 10 days. 257 would be
the same loss over 30 days. A 30-day month has two remainder
columns. Ten plotted days occupy at most 22 of the 62 when both
remainder columns fall on days 1 through 10, so columns 23 through
62 stay empty whichever days receive a remainder. Days 11 through 30
still take their share of the 62. A November day's width on the 10th
matches a finished November. Day 30 is an empty slot and has no
label.

### S10 — The trend keeps the previous month

**Given** today is 15 December 1990
**And** 31 October has 90 kg and 1 November has 80 kg
**And** November has no other weights
**When** the November chart is shown
**Then** 1 November has a diamond and a stem
**And** a later blank November day has Braille trend and no diamond
**And** the left-axis end labels are 79 and 90
**And** 91 is not an axis label
**And** October is not plotted on this chart

The carried trend on 1 November is 89 kg, not 80. The scale therefore
runs from 79 to 90. A trend computed from November alone would have
no stem and would not reach 90.

### S11 — The trend is not a line through the weights

**Given** a finished November with 80 kg on day 1 and 100 kg on day 2
**And** no earlier logs
**When** the chart is shown
**Then** day 1 has a diamond and no stem
**And** day 2 has a diamond and a stem
**And** the left-axis end labels are 79 and 101

A line through the two weights would put both diamonds on the line,
so neither day would have a stem. A scale that saw only the trend
would not reach 101.

### S12 — Colour roles, and no new keys

**Given** a weighed day whose weight differs from the trend, and a
later blank day
**And** colour is on and no colour config is set
**When** the chart is shown
**Then** the diamond is blue
**And** the blank day's trend is red
**And** the stem is green

**Given** `weight` is magenta and `trend` is yellow
**When** the chart is shown
**Then** the diamond is magenta
**And** the blank day's trend is yellow
**And** the stem is still green

**Given** a config that sets `diamond` and `stem` and an invalid
`trend` value
**When** the chart is shown
**Then** the diamond stays the weight colour
**And** the stem stays green
**And** the trend stays the default red

### S13 — `NO_COLOR` keeps the three shapes

**Given** `NO_COLOR` is non-empty
**And** a weighed day whose weight differs from the trend, and a
later blank day
**When** the chart is shown
**Then** the view text has no chromatic colour
**And** the diamond, the Braille trend, and the stem are all present
**And** the blank day's trend is still bold
**And** the diamond is not bold

### S14 — Hex colour and the 16-color floor

**Given** `trend` is the hex colour `#123456`
**And** `NO_COLOR` is empty
**When** the view text is written for a truecolor profile
**Then** that output contains `38;2;18;52;86`
**When** the same view text is written for a 16-color profile
**Then** that output does not contain `38;2`
**And** the app has not already replaced the hex colour in the view
with the 16-color result

### S15 — Pencil loss and deficit stay

**Given** today is 15 December 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**And** the display unit is kilograms
**When** the chart is shown
**Then** Monthly Loss is 1.0 kg
**And** Daily Deficit is 257 calories
**And** the caption is not a rate from only the last seven trend
values

**Given** the same logs and `display_unit` pounds
**When** the chart is shown
**Then** Monthly Loss is 2.2 lb
**And** Daily Deficit is still 257 calories

Day 1's trend is 80 kg and the carried last trend is 79 kg. Over 30
days that is 257 calories. The same loss in pounds, to one decimal,
is 2.2 lb.

### S16 — The long-term chart stays

**Given** the pound series in S6
**When** the long-term chart is opened from the monthly chart
**Then** it is still the Braille-cell plot
**And** it has no open diamonds and no green stems
**And** its vertical scale still runs from 168 lb to 177 lb
**And** its width and its `%5.1f-` gutter are unchanged
**And** it is not drawn as a 62-column or 67-column plot
**And** the four windows are otherwise unchanged

### S17 — The PDF stays

**Given** the monthly chart
**When** the PDF command is used
**Then** the PDF file is still written
**And** the page is still the existing vector chart, with its
existing title
**And** the on-screen heading is still the plain month and year

### S18 — Keys stay, and nothing pointer-driven is added

**Given** the monthly chart, opened from the month sheet
**When** `[` is pressed, then `]`
**Then** the chart moves to the previous month and back
**When** Esc is pressed
**Then** the screen that opened the chart returns
**And** `q` still quits
**And** the help line still offers back, month, the long-term chart,
the PDF, and quit
**And** the chart has no legend, no zoom control, and no ruled grid
**And** the program is still started with the screen model alone, not
an alternate screen and not a mouse mode

### S19 — Empty and carry-only stay

**Given** a month with no weights and no carried trend
**When** the chart is shown
**Then** the view says the month is empty
**And** there is no scale and no loss numbers
**And** there is no 62-column plot
**And** the process does not crash

**Given** a month with no weights of its own and a trend carried from
the previous month
**When** the chart is shown
**Then** the view is not the empty state
**And** the trend is shown on the 62-column plot and there is no
diamond
**And** Monthly Loss is 0 and Daily Deficit is 0

### S20 — ntcharts v2, not v1

**Given** the built program
**When** its module requirements and its source imports are read
**Then** it requires `github.com/NimbleMarkets/ntcharts/v2`
**And** it still requires `charm.land/bubbletea/v2`,
`charm.land/lipgloss/v2`, and `charm.land/bubbles/v2`
**And** the Go version is 1.27.0
**And** it does not require an ntcharts v1 module
**And** it does not require `github.com/charmbracelet/bubbletea`,
`github.com/charmbracelet/lipgloss`, or
`github.com/charmbracelet/bubbles`
**And** no scanned source file imports those v1 modules or an
ntcharts v1 path

### S21 — A 28-day month and a 31-day month share the width

**Given** today is 15 March 1990
**And** February 1990, which has 28 days, has 80 kg on day 1 and no
other weights
**And** January 1990, which has 31 days, has 80 kg on day 1 and 80 kg
on day 31
**When** each monthly chart is shown
**Then** each plot is 62 columns
**And** each January day occupies two plot columns
**And** day 1's diamond lies in plot columns 1 through 2
**And** day 31's diamond lies in plot columns 61 through 62
**And** February is the same plot width, not a narrower one
**And** neither chart takes its width from the terminal

February divides 62 as evenly as whole columns allow. Which of its
days receive a remainder column is not fixed. January has no
remainder, so every day is two columns, not only the ends.

### S22 — The plot is 16 rows

**Given** a monthly chart that has a picture
**When** it is shown
**Then** the plot is 16 rows
**And** the height does not follow the terminal
**And** no two tick labels share a row

**Given** the short kilogram range in S6 and the stone range in S7
**Then** every tick in those ranges is shown

**Given** the wide range in S8
**Then** the end labels 39 and 121 are shown
**And** at least one tick between them is omitted

An empty month has no plot, so it is not 16 rows. A carry-only
month uses the 16-row plot.

## Functional requirements

-   **FR1.** A plotted day with a weight shows one open diamond
    (U+25C7). A plotted day without a weight shows none. The view
    never shows a filled diamond (U+25C6). Workout does not change
    the diamond.
-   **FR2.** The trend is a Braille line of the derived trend on
    every plotted day, including a blank day and a day carried from
    the previous month. It is not a line fitted through the raw
    weights. It is not drawn after today.
-   **FR3.** A weighed day whose weight and trend differ and land on
    different rows has a stem between those two values for that day
    only. The S3 gaps do. The stem is green for a float and for a
    sinker. Its marks are not Braille and are not the diamond. The
    diamond is drawn on top and keeps the weight colour. A day whose
    weight equals the trend has no stem. Two different values that
    land on the same row have no stem. A 1 kg gap on a tall scale is
    not a separate guarantee.
-   **FR4.** The on-screen heading is the month name and the year. It
    uses the `title` role: bold, default cyan, and a `[colors]`
    title entry changes it. It has no border and no filled
    background, and it is not yellow on blue. The words are centered
    across the plot, with space on both sides. `NO_COLOR` keeps the
    words and the bold. The database line and the help line stay.
-   **FR5.** The vertical scale is the display unit. It runs from one
    step below the minimum plotted value to one step above the
    maximum. The step is 1 kg, 1 lb, or 1/14 stone. Labels are the
    step multiples inside that scale. Kilogram and pound labels are
    whole numbers. Stone labels are one decimal of stones, and a
    repeated decimal is still labeled once per pound-step. The top
    and bottom end labels are always shown. A tick never shares a
    row with another tick. Short ranges in S6 and S7 show every
    tick. The wide range in S8 has more ticks than rows, so at least
    one tick between the ends is omitted.
-   **FR6.** Day labels under the plot include the first and last
    plotted days. A day that is not plotted is not labeled, including
    the empty last calendar day of the current month. The current
    month's last plotted day is today. A later logged day is not
    plotted and does not change the caption.
-   **FR7.** The trend uses the `trend` role (default red, bold). The
    diamond uses the `weight` role (default blue, not bold). The stem
    is green chrome, not bold. Config overrides `trend` and `weight`
    only. `diamond` and `stem` keys do nothing. An invalid colour
    leaves the default.
-   **FR8.** A non-empty `NO_COLOR` removes chromatic colour from the
    view and keeps the diamond, the Braille trend, the stem, and the
    trend's bold. The check is the variable, not a terminal.
-   **FR9.** Hex colour is judged on the written output. Truecolor
    output uses `38;2`. 16-color output does not. The app does not
    downsample the view text.
-   **FR10.** Monthly Loss and Daily Deficit stay the pencil
    identity in Decision 10. For the finished 80 kg then 70 kg
    November they are 1.0 kg and 257 calories. In pounds the loss is
    2.2 lb and the deficit is still 257. For the current month in S9
    the deficit is 772 calories.
-   **FR11.** The long-term chart stays the Braille-cell plot with
    the fixed two-pound margin and the four existing windows. The
    170 lb to 175 lb series still scales from 168 lb to 177 lb there.
    Its width and its `%5.1f-` gutter stay as they are. This slice
    does not make that chart 62 or 67 columns.
-   **FR12.** The PDF command still writes the existing vector page
    and its existing title. This slice does not redraw that page.
-   **FR13.** Chart keys stay: open, previous month, next month,
    back, long-term, PDF, and quit. The program is started with the
    screen model alone. The chart has no grid, legend, zoom, or
    mouse highlight.
-   **FR14.** An empty month shows the empty state, with no scale,
    no caption numbers, and no 62-column plot. A carry-only month
    shows the trend on the 62-column plot, no diamond, loss 0, and
    deficit 0.
-   **FR15.** The program requires ntcharts v2 and the three Charm v2
    modules, does not require ntcharts v1 or the Charm v1 modules,
    and keeps the Go version at 1.27.0.
-   **FR16.** A month that has a picture uses a 62-column plot, not
    counting the tick gutter, and does not grow or shrink with the
    terminal. Each calendar day gets one contiguous share of those
    columns, in order, covering plot columns 1 through 62. A 31-day
    month is two columns per day. Any other length divides 62 as
    evenly as whole columns allow. Which days receive a remainder
    column is not specified. Shares depend on how many days the
    month has, not on today. The line, the stems, and the diamonds
    stop at today. Days after today are empty slots and still take
    their share. On 10 November, plot columns 23 through 62 contain
    no diamond, no stem, and no trend. A finished November and a
    31-day month still fill all 62 columns with day slots. Every
    plot row is 67 visible columns: a five-column gutter, then the
    plot. Four gutter columns hold the tick, right-aligned, and the
    fifth is the hyphen. `  79-`, ` 121-`, and `12.6-` fit. Width is
    visible columns, not bytes. The long-term chart keeps its
    current width and its `%5.1f-` gutter. An empty month has no
    62-column plot. A carry-only month uses the 62-column plot.
-   **FR17.** A month that has a picture uses a 16-row plot. The
    height does not follow the terminal. No two tick labels share a
    row. An empty month has no 16-row plot. A carry-only month uses
    the 16-row plot.
-   **FR18.** The monthly how-to describes the plain heading and the
    one-step scale, not the month-year box or the ±2 lb margin. The
    long-term how-to keeps the ±2 lb margin and does not say that
    margin matches the monthly chart. The PDF how-to keeps the
    existing vector page, including its title box, and does not say
    the PDF is the on-screen chart. The `idea.md` change record for
    this draft says the monthly screen drops the box and the shared
    margin, the long-term screen keeps both until #78, and the PDF
    keeps the old page. Older change records are left as history.

## Resolved ambiguities

These are closed. They are not open questions.

The slice says a `[colors]` entry overrides the part it names. The
parts that already have names are `trend` and `weight`. The trend
line uses `trend` (default red). The diamond uses `weight` (default
blue); "open" is the glyph U+25C7, not a new `diamond` key. The stem
stays the green chrome the previous monthly chart used. There is no
`stem` key. Invalid colours still fail open.

The scale ends are exactly one step beyond the plotted minimum and
maximum, even when that end is not a whole number. The 80.4 kg to
82.4 kg range therefore runs from 79.4 to 83.4, and the labels are
80 through 83, not 79 through 84. Snapping the scale outward onto
the next whole number was rejected because that is more than one
step. Snapping inward was rejected because that is less than one
step of room.

Stone "whole number" means a whole pound on the stone scale the TUI
already uses. The printed label is one decimal of stones. It is not
a whole number of stones and not a stones-and-pounds pair.

Day labels follow the same end rule as crowded ticks. The first and
last plotted days are always labeled. Days between may be labeled.
The slice does not require all 30 day numbers.

The canvas runs through the last calendar day. On 10 November the
slots still cover all 30 days and all 62 columns. Day 30 is empty,
and it is unlabeled because it is not plotted. The plotted days end
at 10, so the labels include 1 and 10 and do not include 15 or 30.
No label is added on that empty last day.

Which days of a 28-day or 30-day month receive a remainder column is
not fixed. Shares are contiguous whole columns from plot column 1
through plot column 62, in calendar order, with no inset. The 10
November empty tail is plot columns 23 through 62 because ten days
of a 30-day month take at most 22 columns even when both remainder
columns fall on days 1 through 10. That bound is the check. It does
not name the days that are one column wider.

Only a scale line is a 67-column plot row. The heading, the database
line, the day-label line, the caption, and the help line are not
held to that width.

The hyphen is the separator the picture already uses. On the monthly
chart it follows a four-column tick, so the gutter is five columns.
That is not the six-column `%5.1f-` gutter the long-term chart
keeps. ` 79.0-` does not fit the monthly gutter. This slice does
not give the long-term chart the 62-column plot.

The plot is 16 rows whenever it has a picture. That height does not
follow the terminal. S6 and S7 fit in those rows, so every tick is
shown. S8 does not, so at least one middle tick is omitted, and the
ends stay. A tick never shares a row. The S3 gaps occupy different
rows and show a stem between the diamond and the trend. Two
different values on one row have no stem. A 1 kg gap on a tall
scale is not required to show one.

The heading uses the `title` role. Default cyan, bold, centered
across the plot, with no border and no filled background. A
`[colors]` title entry changes the heading and does not change the
trend. `NO_COLOR` keeps the bold. The old yellow-on-blue box is
gone from this screen.

The how-tos and the `idea.md` change record are amended as
Decision 14 says. The older backlog paragraphs stay the history of
the slices that introduced the box and the shared margin.

## Out of scope

Already accepted, and not this slice:

-   The Online caption, issue
    [#79](https://github.com/jglueckstein/hdtools/issues/79).
    Last-seven and straight-line rates stay out. The PDF caption
    stays the pencil identity too.
-   The long-term Online picture, issue
    [#78](https://github.com/jglueckstein/hdtools/issues/78).
    The four windows are not redrawn here.

Closed, and not slices to file:

-   ntcharts v1
-   Mouse highlight, zoom, a grid, or a legend
-   A filled diamond, or a diamond coloured by which side of the
    trend it is on
-   A flag field or an exercise-rung line
-   A regression through the raw weights
-   A custom date range
-   BubbleZone interaction, or a bordered title on this chart
-   Redrawing the PDF

Logging, saving, and the stored kilograms are unchanged. Display
conversion is for the chart and the rest of the screen, not for the
save.
