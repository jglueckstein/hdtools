---
spec: docs/superpowers/specs/2026-10-14-online-caption.md
date: 2026-10-14
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "Displayed loss multiplies a rate estimated only on days that have a trend by the full plotted span, so the caption reports a loss the trend did not have."
    evidence: "Decision 5: \"Displayed loss is the daily rate times that span length.\" S6 requires \"Monthly loss: 30.0 kg   Daily deficit: 7716 cal\" for weights only on days 29 and 30, and the resolved section says that 1 kg/day rate \"spreads that rate across all 30 days.\""
    disposition: accepted
    disposition_rationale: >
      Accepted. The printed loss is the change over the days that
      entered the rate. The current month shows first minus last of
      the window and does not multiply by today's day of the month.
      A fit shows the daily rate times the gap from the first trend
      day to the last. Days with no trend stay out of that total.
      Calories stay 3500 times the unrounded daily rate in pounds.
      Fixture B is 1.0 kg and 3858 cal. S6 is 1.0 kg and 7716 cal.
      S7 is 0.5 kg and 421 cal. Fixture A stays 0.0 kg and 0 cal.
      S5 still prints 0.2 kg and 50 cal.
  - id: O2
    category: implementation
    severity: high
    claim: "The same two weigh-ins are 15.0 kg and 3858 cal on 30 November and 30.0 kg and 7716 cal once that month is finished, because the open-month rate divides by the count of trend values and the fit divides by the one-day gap."
    evidence: "Decision 3: \"The daily rate is that loss divided by that count.\" Decision 5: the current month's span \"is today's day of the month.\" S6: offsets 28 and 29 are \"one day apart, so the slope is −1 kg per day and the daily rate is 1 kg,\" and \"the displayed loss is 30.0 kg.\""
    disposition: accepted
    disposition_rationale: >
      Accepted. The current-month divisor is the number of calendar
      days from the first day in the window to the last, not the
      count of values. The window is consecutive, so the gap is one
      less than the count. Two weigh-ins one day apart have the same
      daily rate under both rules. S6's logs read on 30 November and
      on 1 December are both 1.0 kg and 7716 cal. The last seven stay
      a pencil difference. S4 is 1.0 kg and 1543 cal.
  - id: O3
    category: scope
    severity: medium
    claim: "Decision 9 forbids editing the PDF how-to, whose opening sentence still says the PDF has the same Monthly Loss and Daily Deficit copy as the screen S8 makes disagree."
    evidence: "Decision 9: \"The PDF how-to is not edited.\" S8 requires the PDF to contain \"Daily deficit: 772 cal\" and not \"Daily deficit: 0 cal\" for the logs whose screen caption in S1 is \"Daily deficit: 0 cal.\""
    disposition: accepted
    disposition_rationale: >
      Accepted. The PDF how-to is edited. It stops saying the file
      has the same Monthly Loss and Daily Deficit as the screen. It
      keeps the pencil identity and says the on-screen caption is a
      different rate. Decision 9 names that page.
  - id: O4
    category: specification quality
    severity: medium
    claim: "The pound scenarios do not separate conversion of the unrounded kilogram loss from conversion of the already-printed one-decimal kilogram figure, so that forbidden path still matches every Then line."
    evidence: "S2: \"7.3 lb is one decimal of the display conversion of the unrounded 10/3 kg. It is not a second rounding of 2572 calories, and it is not the conversion of the already-printed 3.3 kg.\" The Then line is only \"Monthly loss: 7.3 lb   Daily deficit: 2572 cal.\""
    disposition: accepted
    disposition_rationale: >
      Accepted. A finished November with 70 kg on day 1 and 60 kg on
      day 15 prints 1.4 kg. In pounds the line is 3.2 lb and 385 cal.
      Converting the printed 1.4 kg would show 3.1 lb. The unrounded
      loss is 6496/4495 kg.
  - id: O5
    category: specification quality
    severity: medium
    claim: "Decision 9's how-to requirement can be met without saying that loss is the daily rate times the whole span, so the published page can describe a fitted drop the screen does not show."
    evidence: "Decision 9: the monthly how-to \"describes the last seven trend values for the current month and the straight-line fit for every other month, still at 3500 kcal per pound, and it still says the PDF is not that rate.\""
    disposition: accepted
    disposition_rationale: >
      Accepted. The monthly and long-term how-tos say the printed
      loss is the change over the days that entered the rate, and
      that the daily deficit is 3500 times the unrounded daily rate
      in pounds. Days with no trend are not part of that change.
      The PDF sentence from O3 stays on the PDF how-to.
---

## O1 — implementation — high

### Claim

Displayed loss multiplies a rate estimated only on days that have a
trend by the full plotted span, including days with no trend. The
caption then reports a Monthly Loss the series did not have.

### Evidence

Decision 5:

> Displayed loss is the daily rate times that span length. … The
> plotted span length … includes days with no trend.

S2 prints `Monthly loss: 3.3 kg` for a window loss of 1.0 kg over
three trend values, spread across 10 plotted days. S6 is the same
rule on a finished month: 80 kg on day 29 and 70 kg on day 30, no
earlier logs, trend 80 kg then 79 kg. The slope is −1 kg per day.
The required caption is `Monthly loss: 30.0 kg   Daily deficit:
7716 cal`. The resolved section says this on purpose: the rate "spreads
that rate across all 30 days." S7 does it on the long-term screen:
a 1 kg trend step becomes `Loss: 3.9 kg` because the slope is
multiplied by 71 days, 61 of which have no trend. S4 prints
`Monthly loss: 0.7 kg` for a November whose own trend never leaves
79 kg.

US2's reason is "so the rate is the line through the span and not
only the first and last points." Days with no trend are not on that
line. Decision 4 leaves them out of the fit. Decision 5 then puts
them back as a multiplier.

### Why this matters

The slice's phrase "the loss that rate implies across the span" is
what Decision 5 implements. This spec is the first place that phrase
becomes a number a person will read as the month's loss: 30.0 kg and
7716 cal/day for a trend that moved 1 kg, or 3.3 kg for a 1 kg step
in fixture B. The label stays `Monthly loss` (Decision 6), which
until now meant the actual first-to-last change. A loss taken only
on the days that have a trend — the change along the fitted line
from the first trend day to the last — would still be a straight-line
fit, and for S6 it would be 1.0 kg, not 30.0 kg. The spec records
the spread and does not weigh that cut. Shipping S6 as written
teaches the open month and every finished sparse month to invent a
total.

## O2 — implementation — high

### Claim

The same two weigh-ins are a 15.0 kg and 3858 cal caption on
30 November and a 30.0 kg and 7716 cal caption once that month is no
longer current. The open-month rate divides by the count of trend
values. The fit divides by the one-day calendar gap. Nothing in the
log changes overnight.

### Evidence

Decision 3, for the month that contains local today:

> The loss in kilograms is first minus last. The day count is the
> count of values in the window, not the length of the plotted span.
> The daily rate is that loss divided by that count.

Decision 5: that current-month span "is today's day of the month."
Decision 4, for every other month: "Two points are the line through
them." S6 fixes those two points:

> Those calendar offsets are 28 and 29, one day apart, so the slope
> is −1 kg per day and the daily rate is 1 kg. The plotted span is
> 30 days, so the displayed loss is 30.0 kg.

The deficit on that rate is 7716 cal. No scenario asks for the
caption of those same logs on 30 November.

### Why this matters

On 30 November the walk finds two trend values, 80 kg and 79 kg, and
stops because 28 November has no trend. The window loss is 1.0 kg.
Dividing by the count of values gives 0.5 kg/day. Times today's day
of the month, 30, the caption is 15.0 kg and
round(0.5 / 0.45359237 × 3500) = 3858 cal. On 1 December the
November sample is still those two days. There is no later day for
carry to fill. The fit's slope is −1 kg per calendar day, and S6's
caption is 30.0 kg and 7716 cal. The chart still ends on day 30,
with the same two diamonds.

US2's reason for preferring a fit does not apply here. With two
points, the line through them is the first and the last. The
formulas still disagree by two, because one divides by the number of
samples and the other by the number of days between them. A week of
perfectly linear samples disagrees by 7/6 for the same reason. The
spec never places the two captions on one series, so the doubling is
not a resolved trade. It is what the two rules do when the month
rolls over. Using the fit for the seven-day window as well, or
dividing the pencil window by the number of days between the first
and last value, would remove the jump. Neither is weighed.

## O3 — scope — medium

### Claim

Decision 9 forbids editing the PDF how-to. That page still says the
PDF carries the same Monthly Loss and Daily Deficit copy as the
on-screen chart. S8 makes those numbers different, and the monthly
how-to is the only page this spec updates to say so.

### Evidence

Decision 9:

> The PDF how-to is not edited.

`docs/how-to/monthly-chart-pdf.md` opens with:

> title box, daily marks, stems, trend, Y margin, and the same
> Monthly Loss and Daily Deficit copy. It keeps that page when the
> on-screen chart changes.

S1's screen caption for 80 kg on day 1 and 70 kg on day 2, read on
10 November, is `Daily deficit: 0 cal`. S8 requires the PDF of those
logs to contain `Daily deficit: 772 cal` and not `Daily deficit:
0 cal`. The finished-month pair is 50 cal on screen (S5) and 257 cal
on the PDF (S8).

### Why this matters

"The same … copy" was true while both surfaces used the pencil
identity. This slice is the change that breaks it. Decision 9
updates `docs/how-to/monthly-chart.md` so that page "still says the
PDF is not that rate," and it explicitly leaves the PDF how-to
alone. A person who exports with `p` and then reads the export
how-to is told the file has the same loss and deficit copy as the
chart they are looking at. Fixture A is 0 cal on the screen and
772 cal in the file. The two how-tos will disagree on the day this
ships, and no acceptance scenario reads the PDF how-to.

## O4 — specification quality — medium

### Claim

The pound scenarios do not separate conversion of the unrounded
kilogram loss from conversion of the already-printed one-decimal
kilogram figure. The path Decision 5 forbids still matches every
Then line.

### Evidence

Decision 5: the displayed loss "is converted to the display unit on
the existing display path" and "A one-decimal figure can hide the
unrounded kilogram loss." S2 states the distinction and then checks
only the printed line:

> 7.3 lb is one decimal of the display conversion of the unrounded
> 10/3 kg. It is not a second rounding of 2572 calories, and it is
> not the conversion of the already-printed 3.3 kg.

The Then line is `Monthly loss: 7.3 lb   Daily deficit: 2572 cal`.
S5's pound line is `Monthly loss: 0.4 lb   Daily deficit: 50 cal`
for the unrounded 0.193548… kg. S7, where the two conversions
diverge, is specified only in kilograms.

### Why this matters

10/3 kg prints as 3.3 kg. 10/3 ÷ 0.45359237 prints as 7.3 lb, and
3.3 ÷ 0.45359237 also prints as 7.3 lb. The unrounded 0.193548… kg
prints as 0.4 lb, and the printed 0.2 kg also prints as 0.4 lb.
Tests driven from the Then lines, which is how this repo writes
tests, stay green if the screen rounds the kilogram loss to one
decimal and converts that. The calorie side is actually locked: S5's
50 cal is not the figure from the printed 0.2 kg, which is about
51. The loss line in another unit is not. S7 would have separated
them. 213/55 kg is 8.5 lb to one decimal, and the printed 3.9 kg is
8.6 lb. S7 never switches the display unit. A later pound or stone
reading can show the rounded kilogram figure rather than the
unrounded loss, and every scenario in this spec still passes.

## O5 — specification quality — medium

### Claim

Decision 9 tells the how-tos to name the last-seven rate and the
straight-line fit. It does not require them to say that the printed
loss is that daily rate times the whole span, including days with no
trend. A page that meets the decision can describe a loss the screen
does not show.

### Evidence

Decision 9:

> `docs/how-to/monthly-chart.md` stops saying the on-screen numbers
> are the book's first-and-last identity. It describes the last seven
> trend values for the current month and the straight-line fit for
> every other month, still at 3500 kcal per pound, and it still says
> the PDF is not that rate.

The long-term how-to "describes the straight-line fit." FR10 repeats
that list and adds no further sentence. The span multiplier lives in
Decision 5 and in S6, not in the documentation requirement.

### Why this matters

A how-to that says the finished month uses a straight-line fit at
3500 kcal per pound, and that the PDF does not, satisfies every
clause of Decision 9 and FR10. The ordinary reading of "the loss
from that fit" is the change along the line from the first trend day
to the last, about 1 kg in S6, or the slope times the gap between
those days. The screen shows 30.0 kg. Nothing in the doc contract
makes that spread part of the published page. The code scenarios and
the how-to can both be "done" while describing different quantities,
which is the divergence Decision 9 was added to prevent.

## Explicitly not objecting to

- **Pencil on the last seven, as against a seven-day fit:** The
  accepted slice already separates "last seven trend values" from "a
  straight-line fit," and a flat week is zero either way. The cost of
  running both estimators is O2, not the bare choice.
- **A long-term window that ends today staying on the fit:** S7 and
  the resolved ambiguities require the quarterly caption not to be
  the last-seven zero. That screen split is specified.
- **S7's x = 0..9 for a window that starts on 1 September:** The spec
  says a constant shift does not change the slope, and both origins
  print 3.9 kg and 421 cal.
- **Leaving the PDF on the pencil numbers:** US3, Decision 1, and S8
  require 772 and 257 on paper. O3 is only the how-to sentence that
  still calls that copy the same as the screen.
- **A finished flat carry showing 0.0 and 0 cal:** Decision 7 already
  says so, and it is the same line the pencil path shows now.
- **"Nothing earlier remains" against the October carry in S4:** The
  resolved walk stops at the first day with no trend, and S4 requires
  the October values in the window.
- **The old long-term picture, the ±2 lb gutters, Go 1.27, and no
  regression library:** S7's ` 80.9-` and ` 69.1-` are the existing
  margin around the 70 kg weight and the 80 kg trend, and the slope
  is written out in the scenarios.
- **Fitting days rather than plotted columns:** Decision 4 already
  says the fit uses the days of the window, not the columns those
  days may have been bucketed into. S7 not forcing a narrow width is
  a test gap, not a second rule.
