---
task: >
  Upgrade hdtools from Bubble Tea v1 and Lip Gloss v1 to Charm v2
  first, without changing the current screens. After that, draw the
  on-screen charts with ntcharts v2 in the Hacker's Diet Online
  picture. The monthly chart is a Braille trend line with an open
  diamond and a green stem for each weighed day. The long-term chart
  is a thin weight line and a Braille trend line. Stems join the day's
  weight to the trend at that day, drawn after the trend and before
  the diamonds. A plain heading names the month or the window. Round
  weight ticks run down the left in the display unit, one step of
  padding beyond the data, and day or month labels run along the
  bottom. The caption uses ApplyTrend: the last seven trend values
  for the current month, and a straight-line fit to the trend for
  every other chart, with calories at 3500 kcal per pound. Colours
  default to a red trend, a green stem, and an open diamond. A
  [colors] entry can override the named part, and NO_COLOR keeps the
  shapes. This does not redraw the PDF, add a flag or an exercise
  line, fit a line through the raw weights, or change the four
  long-term windows.
task_slug: charm-v2-then-online-charts
date: 2026-10-07
carpaccio_model: grok-4.7
inseparable: false
progressed_slice: S1
slices:
  - id: S1
    title: Upgrade Charm to v2 and leave every screen as it is
    scope: >
      Move the TUI to Charm v2: Bubble Tea, Lip Gloss, and bubbles.
      The month sheet and the day form import bubbles textinput, so
      the upgrade includes that module. The daily list, month sheet,
      day form, monthly chart, long-term chart, and monthly PDF keep
      the behaviour they have now, including the Braille-cell charts
      already on master. This slice does not add ntcharts and does
      not draw the Online picture.
    decision_focus: >
      Does the Charm v2 upgrade land by itself, with every current
      screen still behaving as it does today, before any ntcharts
      work?
    lens_used: decision-boundary
    sequencing_note: >
      First. S2 and S4 need this stack. S3 does not need it to
      compute a rate, and still waits so the upgrade stays free of
      the caption change.
    disposition: accepted
    disposition_rationale: >
      Accepted. Charm v2, including bubbles because the month sheet
      and the day form use textinput, lands with every current screen
      left as it is. The Online picture waits until that stack is in.
      This is the slice in progress.
    file_as_issue: false
    issue_url: null
    merged_into: null
  - id: S2
    title: Draw the monthly chart in the Online picture
    scope: >
      On the monthly chart only, replace the Braille-cell plot with
      the Hacker's Diet Online picture drawn by ntcharts v2. The
      trend is a Braille line of ApplyTrend across the plotted days.
      Each day with a weight gets an open diamond (U+25C7). A green
      stem joins that weight to the trend at that day. Draw the
      trend, then the stems, then the diamonds. A day with no weight
      has no diamond. Workout does not change the diamond. A plain
      heading names the month and year. Whole-number weight ticks run
      down the left in the display unit, from one step below the
      minimum to one step above the maximum. Stone uses a 1 lb step
      in the stone display the TUI already uses. Crowded labels may
      be skipped; the end ticks stay. Day labels run along the
      bottom. There is no grid, legend, mouse, or zoom. Defaults are
      a red trend, a green stem, and an open diamond. A [colors]
      entry overrides the part it names. NO_COLOR keeps those shapes
      uncolored, and the Braille trend, the diamond, and the stem
      still tell the series apart. Monthly Loss and Daily Deficit
      stay the pencil identity. The current month still ends at
      today. The long-term chart and the PDF stay as they are.
    decision_focus: >
      Does the monthly Online picture ship while Monthly Loss and
      Daily Deficit stay the pencil identity, so the number change
      is its own later engagement?
    lens_used: decision-boundary
    sequencing_note: >
      After S1. Before S3 and S4. S4 reuses this axis, heading,
      colour rule, and ntcharts canvas.
    disposition: accepted
    disposition_rationale: >
      Accepted and filed for later. The monthly picture lands while
      Monthly Loss and Daily Deficit stay the pencil identity, so the
      pixels are judged apart from the calorie figure.
    file_as_issue: true
    issue_url: https://github.com/jglueckstein/hdtools/issues/77
    merged_into: null
  - id: S3
    title: Replace the pencil caption with the Online analysis
    scope: >
      Change the on-screen Monthly Loss and Daily Deficit. The drawn
      trend stays ApplyTrend. For the current calendar month the rate
      uses the last seven trend values, carrying into the previous
      month when this month has fewer than seven. That carry feeds
      the rate only. It does not plot the earlier days, and the chart
      still ends at today. Every other monthly chart, and every
      long-term window, uses a straight-line fit to the trend values
      in the span. The calorie figure is 3500 kcal per pound of that
      rate. A current month and a past month on the monthly screen
      show the two rates. The long-term caption uses the fit even
      while that screen still shows the pre-S4 picture. The PDF keeps
      the pencil identity.
    decision_focus: >
      Are the published Monthly Loss and Daily Deficit the Online
      rates (last seven trend values for the current month, a
      straight-line fit to the trend otherwise) rather than the first
      trend minus the last trend of the whole span?
    lens_used: decision-boundary
    sequencing_note: >
      After S2, so the new numbers sit under the new monthly chart.
      Before S4, so the long-term picture is specified with the
      caption it will show. The formula does not need ntcharts.
    disposition: accepted
    disposition_rationale: >
      Accepted and filed for later. The current month uses the last
      seven trend values, and every other chart uses a straight-line
      fit to the trend. The long-term caption updates while that
      screen still shows the old plot. The PDF keeps the pencil
      identity.
    file_as_issue: true
    issue_url: https://github.com/jglueckstein/hdtools/issues/79
    merged_into: null
  - id: S4
    title: Draw the long-term chart in the Online picture
    scope: >
      Replace the long-term plot with ntcharts v2. The four windows
      stay: Quarterly, Semiannual, Annual, and Complete. The picture
      is a thin weight line and a Braille trend line. There are no
      diamonds and no stems. A plain heading shows the window name.
      Weight ticks follow the monthly rule from S2. Month labels run
      along the bottom. The caption uses the S3 fit. There is no
      custom date range. The monthly chart and the PDF stay as they
      are.
    decision_focus: >
      Does the long-term screen take the Online picture (thin weight
      line, Braille trend, no diamonds or stems) on the four existing
      windows, reusing the monthly axis and the S3 caption?
    lens_used: decision-boundary
    sequencing_note: >
      Last. Depends on S1 for the stack, S2 for the axis, heading,
      colours, and canvas, and S3 for the caption rate.
    disposition: accepted
    disposition_rationale: >
      Accepted and filed for later. The long-term series are a thin
      weight line and a Braille trend, with no diamonds or stems, on
      the four windows that already exist. The axis comes from S2 and
      the caption from S3.
    file_as_issue: true
    issue_url: https://github.com/jglueckstein/hdtools/issues/78
    merged_into: null
---

This record slices the Charm upgrade and the on-screen Hacker's Diet
Online charts into four pieces. The look and the stack were settled in
the grill. Each slice is one engagement: the upgrade, the monthly
picture, the caption numbers, then the long-term picture. All four
slices are accepted. S1 is the slice in progress. S2 is #77, S3 is
#79, and S4 is #78.

## S1 — Upgrade Charm to v2 and leave every screen as it is — decision-boundary

### Context

The app is on Bubble Tea v1.3.10, Lip Gloss v1.1.0, and bubbles v1.0.0.
ntcharts v2 needs the Charm v2 modules, and the month sheet and the day
form both use bubbles textinput. This slice moves those three modules
and leaves the screens alone. Opening the daily list, the month sheet,
the day form, either chart, or the monthly PDF still shows today's
behaviour. The charts remain the Braille-cell plots already on master.
The Online diamonds, stems, headings, and ticks are not in this slice.

What you can see when it lands: the same screens, on the v2 stack, with
the existing tests still describing that behaviour.

### Decision content

The engagement is the upgrade boundary. Charm v2, Lip Gloss v2, and
bubbles v2 land together, and no screen changes. The alternative, doing
the upgrade and the Online chart in one slice, would make the human
judge a library move and a new picture at the same time. ntcharts v1 on
the current stack is the other alternative, and this slice does not take
it.

### Dependencies

Nothing in the repo has to land first. S2 and S4 need this stack because
they draw with ntcharts v2. S3's rate can be computed on the current
stack; it still follows S2 so the upgrade does not also change the
published numbers.

### Rationale

The user asked for the upgrade first, with current behaviour kept, and
for the chart work to start after it. The three Charm modules share that
one decision. Splitting bubbles off because only two files import it
would be a file cut. The slice is cohesive because a person can run the
app and confirm that nothing visible moved except the dependency line.

## S2 — Draw the monthly chart in the Online picture — decision-boundary

### Context

After S1, the monthly chart becomes the Hacker's Diet Online picture.
The trend is a Braille line of the existing ApplyTrend series. Each
weighed day is an open diamond, U+25C7, and a green stem joins that
day's weight to the trend at the same day. Float means the weight is
above the trend. Sinker means it is below. The stem is green either way.
The draw order is trend, stems, diamonds. A blank day has no diamond.
Workout does not add a marker. There is no flag and no exercise-rung
line.

The coloured title box is gone. A plain heading names the month and
year. Weight numbers run down the left in the display unit: whole
pounds, whole kilograms, or a 1 lb step shown in the stone display the
TUI already uses. The axis starts one step below the minimum and ends
one step above the maximum. The fixed ±2 lb margin is gone. Crowded
labels may be skipped, and the end ticks stay on those whole numbers.
The canvas scale is the display unit. Day labels run along the bottom.
There is no grid, legend, mouse highlight, zoom, or bordered title.

Defaults are a red trend, a green stem, and an open diamond. A [colors]
entry overrides the part it names. NO_COLOR leaves the Braille trend,
the diamond, and the stem in place and drops the colour.

Monthly Loss and Daily Deficit stay the pencil identity: first trend
minus last trend of the plotted span, then that loss in pounds times
3500 divided by the days in the span. The current calendar month still
ends at today. The long-term chart and the PDF vector page stay as they
are.

### Decision content

The engagement is the monthly picture as one change, with the published
numbers left alone. Heading, ticks, colours, diamonds, stems, and the
Braille trend share that picture. The alternative is to change the
caption in the same slice, so a person could not tell a wrong stem from
a wrong calorie figure. This slice asks you to accept the picture first.

### Dependencies

S1 must land first. ntcharts v2 is not added before the Charm upgrade.
S3 follows this slice and is the one that changes the caption. S4 reuses
the axis, the plain heading, the colour rule, and the canvas, and does
not redraw the monthly screen.

### Rationale

The diamond, the stem, and the trend are one picture. A person reads
them together: the stem is the gap between the diamond and the trend at
that day, and the draw order keeps the diamond on top of the stem. An
axis slice with no series, or a colour slice with no marks, would not be
a chart anyone can look at. Keeping the pencil caption makes the new
pixels the only thing this slice asks you to judge.

## S3 — Replace the pencil caption with the Online analysis — decision-boundary

### Context

The on-screen caption stops using the pencil identity. The drawn series
is still ApplyTrend, including the Braille trend from S2. It is not a
regression through the stored weights, and this slice does not add a
regression library.

For the current calendar month, the rate uses the last seven ApplyTrend
values. When the month has fewer than seven, the window carries into the
previous month. Those carried days feed the rate only. They do not
appear on the monthly plot, and the plot still ends at today. For every
other monthly chart, and for every long-term window, the rate is a
straight-line fit to the trend values in the span. Daily Deficit is
3500 kcal per pound on that rate. Monthly Loss is the loss that rate
implies across the span.

You can see it on the monthly screen: the open month and a finished
month publish different rates from the same kind of series. The
long-term caption updates in this slice too, while that screen still
shows the pre-S4 plot. The PDF keeps the pencil identity.

### Decision content

The engagement is the published pair of numbers. The alternative is to
keep first-trend minus last-trend for every span, which is what the app
does now. Another alternative is to fit the raw weights. This slice uses
ApplyTrend values only: the last seven for the current month, and a
straight-line fit to the trend for every other chart.

### Dependencies

S2 lands first so the new numbers sit under the new monthly chart. The
arithmetic does not need ntcharts, and it does not need S1 except
through that order. S4's long-term caption consumes this rate. If this
slice is dropped, S4 keeps the pencil identity on the long-term screen.

### Rationale

The caption is a different decision from the pixels. The picture can be
right while the rate is still the old identity, and the rate can be
wrong while the diamonds are right. Putting both in S2 would make one
review cover two observables. The current-month rule and the
other-charts rule are one formula with two cases, so they stay in one
slice. You can check both cases by opening the current month and a past
month.

## S4 — Draw the long-term chart in the Online picture — decision-boundary

### Context

The long-term screen takes the Online picture on ntcharts v2. The four
windows stay the ones the keys already cycle: Quarterly, Semiannual,
Annual, and Complete. There is no custom date range. The series are a
thin weight line and a Braille trend line. There are no diamonds and no
stems. A plain heading shows the window name. Weight ticks follow the
S2 rule, in the display unit, with the same one-step padding. Month
labels run along the bottom. The caption is the S3 straight-line fit to
the trend. The monthly chart and the PDF stay as they are.

### Decision content

The engagement is the long-term picture. It reuses the monthly axis,
heading, colours, and canvas, and it reuses the S3 caption. The
alternative is a second copy of the monthly diamonds and stems, or one
slice per window. Both would repeat a decision this record already
separated. The thin weight line and the Braille trend are the whole
series change.

### Dependencies

S1 supplies the v2 stack. S2 supplies the axis, the plain heading, the
colour and NO_COLOR rule, and the ntcharts canvas. S3 supplies the
caption rate. The four windows already exist; this slice does not
invent their ranges.

### Rationale

The long-term screen is a second observable. Its series are not the
monthly series: a thin weight line replaces the diamonds, and there is
no stem. Shipping it inside S2 would hide that difference inside the
monthly review. One window per slice would repeat the same picture four
times. The slice is done when each of the four windows shows the thin
line, the Braille trend, the plain window name, and the S3 caption.

## Sequencing recommendation

Land S1, then S2, then S3, then S4.

S1 is the stack the later pictures import, and it is the only slice
whose success is "nothing visible changed." S2 is the first new
picture, judged against the pencil numbers you already know. S3 moves
those numbers once the picture is stable, including the long-term
caption while the long-term plot is still the old one. S4 then draws
the long-term series against a caption that already matches the rate.

S3 can be merged into S2 if you would rather judge the monthly picture
and the monthly numbers together. S3 can be merged into S4 if you would
rather not change the long-term caption until the long-term picture
lands; the monthly numbers would then wait for that screen. The
recommended order keeps each of those observables in its own engagement.

## Explicitly not slicing on

- One slice per file. `go.mod`, `chart.go`, `longchart.go`, and
  `braille.go` are not decision boundaries. S1 is the upgrade. S2 is
  the monthly picture. S4 is the long-term picture.
- The diamond, the stem, and the Braille trend as three slices. They
  are one monthly picture. The stem is the gap between that day's
  weight and the trend at that day, drawn after the trend and before
  the diamond.
- A colour-only slice, a NO_COLOR-only slice, or a new `[colors]` role
  as its own engagement. Colour ships inside the picture that uses it.
- The PDF page. It stays the vector chart it is. Filled log sheets,
  long-term PDFs, and blank sheets stay out of this task.
- The axis scale as its own slice. Whole-unit ticks, the one-step pad,
  the stone 1 lb step, and the display-unit canvas are part of the
  monthly frame in S2. S4 reuses them.
- One slice per long-term window. Quarterly, Semiannual, Annual, and
  Complete already exist and share one picture.
- ntcharts v1, a mouse highlight, zoom, a grid, a legend, a filled
  diamond, a diamond coloured by which side of the trend it is on, a
  flag field, an exercise-rung line, or a regression through the raw
  weights. The grill closed those. They are not later slices of this
  task.
- A Braille-cell reproduction of today's chart on the new stack. S1
  keeps that plot only until S2 and S4 replace it. The new trend is a
  Braille line in the Online picture, which is a different drawing.
