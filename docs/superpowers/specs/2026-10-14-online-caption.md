# Online caption

**Date**: 2026-10-14
**Status**: draft
**Objections**:
[`online-caption.md`](../objections/online-caption.md)
(O1–O5 accepted)
**Issue**: [#79](https://github.com/jglueckstein/hdtools/issues/79)
**Backlog**: [`idea.md`](../../../idea.md) (TUI)
**Slice**: S3 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)

The on-screen Monthly Loss and Daily Deficit change. The current
calendar month, the month that contains local today, takes its rate
from the last seven trend values. Every other month on the monthly
screen, and every long-term window, takes its rate from a
straight-line fit through the trend. The drawn trend stays the
derived trend. It is not a fit through the raw weights.

The monthly picture from
[the monthly Online chart](2026-10-08-monthly-online-chart.md)
stays: Braille trend, open diamond, green stem, plain heading,
62-by-16 plot, five-column gutter, day labels, colours, and
`NO_COLOR`. The long-term plot stays the Braille-cell picture it
has now, including its ±2 lb margin. This slice does not draw the
long-term Online picture. The PDF keeps the pencil identity it has
now. The numbers are not stored.

The Go version stays 1.27.0. No regression library is added.

## User stories

### US1 — Rate the open month from the last seven trends

As a person reading the current month, I want Monthly Loss and
Daily Deficit from the last seven trend values, so a recent flat
trend does not still show the whole month's first-to-last drop.

### US2 — Rate every other chart from a straight line

As a person reading a finished month or a long-term window, I want
those two numbers from a straight-line fit to the trend, so the
rate is the line through the span and not only the first and last
points.

### US3 — Keep the pictures and the PDF numbers

As a person who already reads the monthly picture, the long-term
plot, and the PDF, I want the pictures left as they are and the
PDF left on the pencil numbers, so this change is only the
on-screen caption.

### US4 — Show a gain as negative

As a person whose trend rose, I want a negative loss and a negative
calorie number, so a gain is not shown as a loss.

## Decisions

1.  **On-screen captions only.** This slice changes the Monthly
    Loss and Daily Deficit on the monthly screen and the Loss and
    Daily Deficit on the long-term screen. The PDF keeps the pencil
    identity it has now: first trend minus last trend of the plotted
    span, that loss in pounds times 3500, divided by the number of
    days in that span, rounded to a whole number of calories. The
    PDF keeps calling that pencil path. The pencil numbers are not
    replaced underneath the PDF.
2.  **The drawn series stays the derived trend.** The line on the
    chart is still the trend the month sheet shows. The first
    weight's trend equals that weight. Each later weight moves the
    trend by one tenth of the gap, and that tenth is rounded half
    away from zero to one decimal of a kilogram before it is added.
    A day with no weight keeps the previous trend. The caption does
    not fit the raw weights, and it does not change that formula.
    No regression library is added. The Go version stays 1.27.0.
3.  **Current month: the last seven trend values.** The current
    month is the month that contains local today, the same today
    the chart already uses to stop the plot. Walk backward from
    today through calendar days that have a trend, including days
    in the previous month. A day with no row still counts when the
    trend carries onto it. Take up to seven. If fewer than seven
    exist, use those that exist. Stop at the first day with no
    trend; once the trend has started, later days keep it, so
    nothing earlier remains. The earliest value in the window is
    first. The latest is last. This window is the pencil identity,
    not a least-squares fit. The loss in kilograms is first minus
    last. The day count is the number of calendar days from the
    first day in the window to the last, not the count of values
    and not the length of the plotted span. The window is a
    consecutive run, so that gap is one less than the count of
    values. The daily rate is that loss divided by that gap. Two
    values on consecutive days have a gap of one, and the daily
    rate is the whole loss. Carried days from the previous month
    feed the rate only. They are not plotted, and the chart still
    ends at today. A weight logged after today is not in the window.
4.  **Every other chart: a straight-line fit.** Every month on the
    monthly screen that is not the current month, and every
    long-term window (quarterly, semiannual, annual, and complete),
    uses ordinary least squares through the trend values in the
    span. The long-term screen uses the fit even when the window
    ends today, and even while that screen still draws the pre-S4
    picture. x is the calendar day, one step per day, measured as
    the offset from the first day of the span, starting at 0. A day
    with no trend is left out of the fit and does not renumber the
    days that remain. A weighed day is not the only point: a blank
    day that has a trend is a point, and the point is the trend,
    not the raw weight. The fit uses the days of the window, not
    the columns the plot may have bucketed those days into. Slope
    is kilograms per day. A falling trend has a negative slope. The
    daily rate is the negation of that slope. If the span's first
    day has no trend, the caption is still shown when two later
    days have a trend. Two points are the line through them.
5.  **The printed loss is the change over the rate's own days.**
    For the current month the printed loss is first minus last of
    the window. It is not multiplied by today's day of the month.
    For a fit the printed loss is the daily rate times the gap from
    the first trend day's offset to the last trend day's offset.
    Days with no trend are not part of that change. A finished
    month that has a trend on day 1 uses the offset of the last
    day, which is one less than the day count. The loss is
    converted to the display unit on the existing display path and
    printed to one decimal. The conversion uses the unrounded
    kilogram loss. A one-decimal figure can hide that loss, and
    converting the already-printed tenth is the wrong path. Daily
    Deficit is the unrounded daily rate in pounds times 3500,
    rounded half away from zero to a whole number of calories. The
    pound is the international pound the app already uses,
    0.45359237 kg. Do not divide by the plotted span a second time.
    Do not recompute the calories from the already-rounded loss
    line. Positive loss and positive calories mean the trend fell.
    A gain is a negative loss and a negative calorie number.
    Changing the display unit changes the loss line only. The
    calorie figure stays on the pound rate.
6.  **The words stay.** The monthly screen still says
    `Monthly loss: %.1f %s   Daily deficit: %d cal`, with three
    spaces before `Daily`. The long-term screen still says
    `Loss: %.1f %s   Daily deficit: %d cal`, with those same three
    spaces. It does not say `Monthly loss:`.
7.  **Fewer than two trend values omits the caption.** The current
    month omits it when the window has fewer than two values. Every
    other chart omits it when the span has fewer than two days with
    a trend. The zero line is not shown in place of an omitted
    caption. An empty month still has no caption and no plot. A
    finished month whose carried trend is flat still shows 0.0 and
    0 calories, because that fit has a slope of zero and it has at
    least two trend values.
8.  **The pictures stay.** The monthly screen keeps the Online
    picture named above. The long-term screen keeps the Braille-cell
    plot, the ±2 lb margin, the four windows, and the heading it
    has now. S4, issue
    [#78](https://github.com/jglueckstein/hdtools/issues/78),
    is out of scope. Keys, colours, `NO_COLOR`, and the database
    line stay.
9.  **Published sentences.** When this behaviour is implemented,
    `docs/how-to/monthly-chart.md` stops saying the on-screen
    numbers are the book's first-and-last identity. It describes
    the last seven trend values for the current month and the
    straight-line fit for every other month. It says the printed
    loss is the change over the days that entered that rate, and
    that days with no trend are not part of the change. The daily
    deficit is 3500 kcal per pound of the unrounded daily rate. It
    still says the PDF is not that rate.
    `docs/how-to/long-term-chart.md` keeps the ±2 lb margin and the
    current picture, and it stops saying Loss and Daily Deficit
    match the monthly pencil identity. It describes the straight-line
    fit the same way: the printed loss is the change along the
    line from the first trend day to the last, and the deficit is
    3500 kcal per pound of the unrounded daily rate.
    `docs/how-to/monthly-chart-pdf.md` stops saying the PDF has the
    same Monthly Loss and Daily Deficit copy as the screen. It
    keeps the pencil identity and says the on-screen caption is a
    different rate. `docs/reference` and `keys.md` are not edited.
    Older change records stay history, including the
    monthly-picture record that leaves the pencil identity in place
    until this issue.

## Acceptance scenarios

November 1990 is the fixture month unless a scenario says otherwise.
Weights in the log are kilograms. The display unit is kilograms
unless the scenario says otherwise. Today is local today. A finished
month is read with today on 15 December 1990, so the monthly plot
runs through day 30. The derived trend of a new log starts at the
first weight. After that, each new weight moves the trend one tenth
of the way toward that weight, with the tenth rounded half away
from zero to one decimal of a kilogram before it is added. Later
blank days keep that trend.

Tests drive the screen from messages and read the view. They do not
need a terminal. A test that freezes today or changes directory does
not run beside other tests. The database file is temporary, not an
in-memory shared database. The caption is the loss line. Match the
whole line, including the three spaces before `Daily`. Do not treat
a tick, a day number, or the help line as the caption.

The international pound is 0.45359237 kg. Calorie rounding is half
away from zero. The one-decimal loss is the display conversion of
the unrounded kilogram loss, not a second rounding of the calorie
figure.

### S1 — Fixture A: a flat last seven is zero

**Given** today is 10 November 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**When** the monthly chart is shown
**Then** the caption is `Monthly loss: 0.0 kg   Daily deficit: 0 cal`
**And** it is not 1.0 kg and not 772 calories
**And** it is not 0.5 kg and not 421 calories

Day 1's trend is 80 kg. Day 2's trend is 79 kg, and days 3 through
10 stay 79 kg. The last seven days that have a trend are days 4
through 10, all 79 kg, so the rate is zero. The plotted span is 10
days. The old pencil caption for this series is 1.0 kg and 772
calories. A straight-line fit through the ten plotted trends would
be 0.5 kg and 421 calories. Neither of those is the caption.

**Given** the same logs, and also 60 kg on 15 November
**When** the monthly chart is shown
**Then** the caption is unchanged
**And** day 15 is not plotted
**And** the chart still ends at day 10

### S2 — Fixture B: a short window prints its own change

**Given** today is 10 November 1990
**And** the only weights are 80 kg on day 8 and 70 kg on day 9
**When** the monthly chart is shown
**Then** the caption is
`Monthly loss: 1.0 kg   Daily deficit: 3858 cal`
**And** it is not 257 calories
**And** it is not 3.3 kg and not 2572 calories

Only three trend values exist: 80 kg on day 8, 79 kg on day 9, and
79 kg on day 10. The loss is 1.0 kg. The gap from day 8 to day 10
is two days, so the daily rate is 0.5 kg. The printed loss is that
1.0 kg. It is not spread across the 10 plotted days, which would
print 3.3 kg. The daily deficit is
round(0.5/0.45359237×3500), which is 3858 calories. Dividing the
window by three values instead of by the two-day gap would show
2572 calories.

**Given** the same logs and the display unit is pounds
**When** the monthly chart is shown
**Then** the caption is
`Monthly loss: 2.2 lb   Daily deficit: 3858 cal`
**And** the view does not show the loss in kilograms

2.2 lb is one decimal of the display conversion of the unrounded
1.0 kg. The deficit stays 3858 calories.

### S3 — A gain is negative

**Given** today is 10 November 1990
**And** the only weights are 70 kg on day 8 and 80 kg on day 9
**When** the monthly chart is shown
**Then** the caption is
`Monthly loss: -1.0 kg   Daily deficit: -3858 cal`

The three trend values are 70 kg, 71 kg, and 71 kg. The loss over
the window is −1.0 kg, and that is the printed loss. The calorie
figure is the negative of fixture B.

### S4 — Previous-month days feed the rate and are not plotted

**Given** today is 4 November 1990
**And** 30 October has 80 kg and 31 October has 70 kg
**And** November has no weights
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 1.0 kg   Daily deficit: 1543 cal`
**And** it is not 0.0 kg and not 0 calories
**And** it is not 0.7 kg and not 1286 calories
**And** October is not plotted
**And** the chart ends at day 4
**And** November has no diamond

Six trend values exist, from 30 October through 4 November: 80 kg,
then 79 kg five times. The window holds all six because seven do
not exist. The loss is 1.0 kg. The gap from 30 October to 4
November is five days, so the daily rate is 0.2 kg. The printed
loss is that 1.0 kg. Spreading the rate across the four plotted
November days would print 0.7 kg, and dividing by six values
instead of by the five-day gap would show 1286 calories. The daily
deficit is round(0.2/0.45359237×3500), which is 1543 calories.
November's own trends are flat, so a caption read only from the
plotted month would be zero.

### S5 — Fixture C: a finished month uses the fit

**Given** today is 15 December 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 0.2 kg   Daily deficit: 50 cal`
**And** it is not 1.0 kg and not 257 calories
**And** it is not 0.0 kg and not 0 calories

This is not the last-seven rule. The last seven trends are all 79
kg, which would print zero. The fit uses every day that has a
trend. Day 1's trend is 80 kg and days 2 through 30 stay 79 kg. x
runs from 0 through 29. With n = 30, the sums are Σx = 435,
Σx² = 8555, Σy = 2371, and Σxy = 34365. The slope is
(30×34365 − 435×2371) / (30×8555 − 435²) = −435/67425 kg per day.
The displayed loss is the negation of that slope times 29, the gap
from day 1 to day 30. That is 0.187096… kg and prints 0.2 kg. The
daily deficit is round((−slope / 0.45359237) × 3500), which is 50
calories. The old pencil caption is 1.0 kg and 257 calories. The
calorie figure uses the unrounded daily rate, not the printed 0.2 kg.

**Given** the same logs and the display unit is pounds
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 0.4 lb   Daily deficit: 50 cal`

0.4 lb is one decimal of the display conversion of the unrounded
0.187096… kg. The printed 0.2 kg converts to 0.4 lb as well, so
this line does not separate those two paths. The deficit is still
50 calories.

### S6 — Two trend days are enough when the month starts blank

**Given** today is 15 December 1990
**And** the only November weights are 80 kg on day 29 and 70 kg on
day 30
**And** there are no earlier logs
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 1.0 kg   Daily deficit: 7716 cal`
**And** the caption is not omitted
**And** it is not 30.0 kg

Day 1 has no trend. Days 29 and 30 do: 80 kg, then 79 kg. Those
calendar offsets are 28 and 29, one day apart, so the slope is −1
kg per day and the daily rate is 1 kg. The gap from the first trend
day to the last is one day, so the printed loss is 1.0 kg. Spreading
the rate across all 30 days would print 30.0 kg. The deficit is
round(1/0.45359237×3500), which is 7716 calories. The pencil path
omits this month because day 1 has no trend. That omission is not
the on-screen caption.

**Given** the same two logs and today is 30 November 1990
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 1.0 kg   Daily deficit: 7716 cal`
**And** it is not 15.0 kg and not 3858 calories

This month is still current, so the last-seven rule applies. The
walk finds 79 kg on day 30 and 80 kg on day 29, then stops because
day 28 has no trend. The loss is 1.0 kg. The gap is one day, so the
daily rate is the whole loss and the deficit is the same 7716
calories. Dividing by the count of two values, then multiplying by
the 30 plotted days, would print 15.0 kg and 3858 calories.

### S7 — The long-term caption uses the fit; the plot stays old

**Given** today is 10 November 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**When** the quarterly long-term chart is shown
**Then** the view names Quarterly, September 1990, and November 1990
**And** the caption is `Loss: 0.5 kg   Daily deficit: 421 cal`
**And** the view does not say `Monthly loss:`
**And** the caption is not omitted
**And** it is not 0.0 kg and not 0 calories
**And** it is not 3.9 kg
**And** the trend is Braille
**And** the view has no open diamond (U+25C7)
**And** the top tick gutter is ` 80.9-`
**And** the bottom tick gutter is ` 69.1-`

The window is 1 September through 10 November, 71 days. Only 1
through 10 November have a trend: 80 kg, then 79 kg nine times.
Their one-day steps match x = 0..9, so n = 10, Σx = 45, Σx² = 285,
Σy = 791, and Σxy = 3555. The slope is
(10×3555 − 45×791) / (10×285 − 45²) = −45/825 = −3/55 kg per day.
The displayed loss is that rate times the gap from the first trend
day to the last, which is 9: (3/55)×9 = 27/55 kg, which is
0.490909… kg and prints 0.5 kg. Multiplying by the 71-day window
would print 3.9 kg. The daily deficit is
round((3/55)/0.45359237×3500), which is 421 calories. The last
seven November trends are flat, so zero would mean the
current-month rule was used on this window. The September days
have no trend, so the old pencil caption is omitted; this caption
is not. The gutters are the ±2 lb margin around the plotted 70 kg
weight and 80 kg trend, not a one-step scale. The other three
long-term windows use this same fit on their own spans. They are
not drawn as the Online picture.

**Given** the same logs and the display unit is pounds
**When** the quarterly long-term chart is shown
**Then** the caption is `Loss: 1.1 lb   Daily deficit: 421 cal`
**And** it is not 1.2 lb and not 8.5 lb

1.1 lb is one decimal of the display conversion of the unrounded
0.490909… kg. Multiplying the rate by the ten trend days would
print 0.5 kg as well, but 1.2 lb. Multiplying by the 71-day window
would print 3.9 kg and 8.5 lb.

### S8 — The PDF stays on the pencil numbers

**Given** today is 15 December 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**When** the PDF command is used for November
**Then** the PDF contains
`Monthly loss: 1.0 kg   Daily deficit: 257 cal`
**And** it does not contain `Daily deficit: 50 cal`

**Given** today is 10 November 1990
**And** November has 80 kg on day 1 and 70 kg on day 2, and no other
weights
**When** the PDF command is used for November
**Then** the PDF contains
`Monthly loss: 1.0 kg   Daily deficit: 772 cal`
**And** it does not contain `Daily deficit: 0 cal`

Those are the pencil figures for a 1.0 kg drop over 30 days and
over 10 days. The on-screen captions for the same logs are S5 and
S1. The PDF page is otherwise unchanged.

### S9 — Fewer than two trend values omits the caption

**Given** today is 1 November 1990
**And** the only weight is 80 kg on that day
**When** the monthly chart is shown
**Then** the view does not contain `Monthly loss:`
**And** it does not contain `Daily deficit:`
**And** it does not contain
`Monthly loss: 0.0 kg   Daily deficit: 0 cal`

One trend value is not a rate. The pencil path would show the zero
line, because the first and last trend are the same day. The chart
itself is still shown.

**Given** the same logs
**When** the quarterly long-term chart is shown
**Then** the view does not contain `Loss:`
**And** it does not contain `Daily deficit:`

That window has the same single trend value. The caption is omitted
there too.

### S10 — The pound line uses the unrounded kilogram loss

**Given** today is 15 December 1990
**And** November has 70 kg on day 1 and 60 kg on day 15, and no
other weights
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 1.4 kg   Daily deficit: 385 cal`

Day 1's trend is 70 kg. Days 2 through 14 stay 70 kg. Day 15's
trend is 69 kg, and days 16 through 30 stay 69 kg. x runs from 0
through 29. With n = 30, the sums are Σx = 435, Σx² = 8555,
Σy = 2084, and Σxy = 30106. The slope is
(30×30106 − 435×2084) / (30×8555 − 435²) = −3360/67425 =
−224/4495 kg per day. The printed loss is the negation of that
slope times 29, which is 6496/4495 kg, or 1.445161… kg, and prints
1.4 kg. The daily deficit is
round((−slope / 0.45359237) × 3500), which is 385 calories.
Multiplying by 30 days would print 1.5 kg.

**Given** the same logs and the display unit is pounds
**When** the November chart is shown
**Then** the caption is
`Monthly loss: 3.2 lb   Daily deficit: 385 cal`
**And** it is not 3.1 lb

3.2 lb is one decimal of the display conversion of the unrounded
6496/4495 kg. Converting the already-printed 1.4 kg would show
3.1 lb. Multiplying the rate by 30 days would print 1.5 kg, and
that figure converts to 3.3 lb either way, so it does not replace
this lock.

## Functional requirements

-   **FR1.** For fixture A, today 10 November 1990 with 80 kg on day
    1 and 70 kg on day 2, the monthly caption is
    `Monthly loss: 0.0 kg   Daily deficit: 0 cal`. It is not the
    pencil 1.0 kg and 772 calories, and it is not the fit's 0.5 kg
    and 421 calories. A weight on day 15 does not change the caption
    and is not plotted. The chart ends at today.
-   **FR2.** For fixture B, today 10 November 1990 with 80 kg on day
    8 and 70 kg on day 9 only, the caption is
    `Monthly loss: 1.0 kg   Daily deficit: 3858 cal`. In pounds it
    is `Monthly loss: 2.2 lb   Daily deficit: 3858 cal`. The 2.2 lb
    line is the display conversion of the unrounded 1.0 kg. The
    deficit is not 257 calories, and it is not 2572 calories.
-   **FR3.** The gain mirror of fixture B, 70 kg on day 8 and 80 kg
    on day 9, shows
    `Monthly loss: -1.0 kg   Daily deficit: -3858 cal`.
-   **FR4.** With today 4 November 1990, 80 kg on 30 October, and
    70 kg on 31 October, the November caption is
    `Monthly loss: 1.0 kg   Daily deficit: 1543 cal`. October is not
    plotted. The chart ends at day 4 and has no November diamond.
    The caption is not the flat November pencil of 0.0 kg and 0
    calories, and it is not 0.7 kg and not 1286 calories.
-   **FR5.** For fixture C, a finished November with 80 kg on day 1
    and 70 kg on day 2, the monthly caption is
    `Monthly loss: 0.2 kg   Daily deficit: 50 cal`. It is not the
    pencil 1.0 kg and 257 calories, and it is not the last-seven
    zero. In pounds it is
    `Monthly loss: 0.4 lb   Daily deficit: 50 cal`. The 0.2 kg line
    hides the unrounded 0.187096… kg. The 50 calories use that
    unrounded daily rate. The printed 0.2 kg also converts to
    0.4 lb, so this fixture does not separate the two pound paths.
-   **FR6.** A finished November whose only weights are 80 kg on day
    29 and 70 kg on day 30, with no earlier logs, shows
    `Monthly loss: 1.0 kg   Daily deficit: 7716 cal`. The same two
    logs read on 30 November, while the month is still current,
    show that same line. The caption is not omitted for lack of a
    trend on day 1, and it is not 30.0 kg.
-   **FR7.** The quarterly window for 80 kg on 1 November and 70 kg
    on 2 November, read on 10 November 1990, shows
    `Loss: 0.5 kg   Daily deficit: 421 cal`. In pounds it is
    `Loss: 1.1 lb   Daily deficit: 421 cal`. It does not say
    `Monthly loss:`. It is not omitted, not the last-seven zero, and
    not 3.9 kg. The 1.1 lb line is not 1.2 lb. The plot stays
    Braille, with no open diamond. The top gutter is ` 80.9-` and
    the bottom gutter is ` 69.1-`. Semiannual, annual, and complete
    use the same fit on their own spans. None of the four windows
    is drawn as the Online picture.
-   **FR8.** A current month with only one trend value omits the
    monthly caption. It does not show the zero pencil line. The
    quarterly window of that same single day omits `Loss:` and
    `Daily deficit:` as well. An empty month still has no caption.
-   **FR9.** The PDF of fixture C still contains
    `Monthly loss: 1.0 kg   Daily deficit: 257 cal` and not
    `Daily deficit: 50 cal`. The PDF of fixture A still contains
    `Monthly loss: 1.0 kg   Daily deficit: 772 cal` and not
    `Daily deficit: 0 cal`. The PDF keeps the pencil path it has
    now.
-   **FR10.** The monthly how-to describes the last-seven rate and
    the straight-line fit. It says the printed loss is the change
    over the days that entered the rate, that days with no trend
    are not part of the change, and that the deficit is 3500 kcal
    per pound of the unrounded daily rate. It does not give the PDF
    that rate. The long-term how-to keeps the ±2 lb margin and the
    current picture, and it describes the fit the same way. The PDF
    how-to stops saying the file has the same Monthly Loss and
    Daily Deficit copy as the screen. It keeps the pencil identity
    and says the on-screen caption is a different rate.
    `docs/reference` and `keys.md` are unchanged. The derived-trend
    formula is unchanged. The Go version stays 1.27.0. No regression
    library is added. The monthly Online picture and the long-term
    Braille-cell picture stay.
-   **FR11.** A finished November with 70 kg on day 1 and 60 kg on
    day 15 shows `Monthly loss: 1.4 kg   Daily deficit: 385 cal`.
    In pounds it is `Monthly loss: 3.2 lb   Daily deficit: 385 cal`.
    The 3.2 lb line is the display conversion of the unrounded
    6496/4495 kg. Converting the printed 1.4 kg would show 3.1 lb.

## Resolved ambiguities

These are closed. They are not open questions.

x starts at 0 on the first day of the span. Numbering the calendar
days 1 through 30 would not produce fixture C's sums: Σx would be
465, not 435. Leaving a day out of the fit does not slide the later
days left. Shifting every x by the same constant does not change
the slope. A gap between two trend days does.

The last-seven walk steps one calendar day at a time. It uses the
carried trend on a day with no row. It stops when it has seven
values or when it reaches a day with no trend. First is the
earliest of those values, not the first one visited while walking
backward.

The printed loss is the change over the days that entered the rate.
For the current month that change is first minus last. The day count
is the calendar gap, one less than the count of values. For a fit
the change is the daily rate times the gap from the first trend
day's offset to the last. Days with no trend are not part of that
change. Fixture B, S4, S6, and S7 print those gaps.

A long-term window never uses the last seven, including a window
that ends today. S7 is that case: the quarterly window ends on 10
November, and the caption is the fit.

The one-decimal print can hide the kilogram loss, as 0.187096… kg
prints 0.2 kg and 6496/4495 kg prints 1.4 kg. Calories are computed
from the unrounded daily rate in pounds. Half away from zero
matches the pencil calories the app already rounds that way. The
pound line uses that same unrounded kilogram loss. S10 is the case
where converting the printed tenth would show a different pound
figure.

S6 follows the same gap rule. Two trend values on the last two days
of a 30-day month are a rate of 1 kg per day, and the printed loss
is that 1.0 kg change. It is not spread across all 30 days. Read on
30 November, while the month is still current, the last-seven
window is those same two days and the caption matches.

The required fixtures were checked against the formulas above.
Fixture A's caption is 0.0 kg and 0 calories. Its old pencil caption
is 1.0 kg and 772 calories, and a fit of the ten plotted trends
would be 0.5 kg and 421 calories. Neither is the caption. Fixture
B's deficit is 3858 calories. Fixture C's slope is −435/67425 kg
per day, the displayed loss prints 0.2 kg, and the deficit is 50
calories. S10's unrounded loss is 6496/4495 kg, which prints 1.4 kg
and 3.2 lb. Those figures stand.

## Out of scope

-   The PDF pencil identity, and any redraw of the PDF page
-   S4's long-term Online picture, issue
    [#78](https://github.com/jglueckstein/hdtools/issues/78)
-   Diamonds, stems, axes, headings, and the monthly plot size
-   The derived trend's formula
-   A regression through the raw weights
-   A regression library
-   New colour roles, flags, or exercise rungs
-   Meal planning
-   Storing Monthly Loss or Daily Deficit
