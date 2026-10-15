---
spec: docs/superpowers/specs/2026-10-14-online-caption.md
date: 2026-10-15
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "FitRate negates a positive-zero slope numerator, so a flat finished month and a flat long-term window print a loss of -0.0 while calories stay 0."
    evidence: "internal/dailylog/analysis.go: daily := -(nf*sumXY - sumX*sumY) / den; return daily * gap, kcalFromDailyKG(daily), true. internal/tui/chart.go prints that loss with fmt.Sprintf(\"Monthly loss: %.1f %s   Daily deficit: %d cal\", loss, unit, kcal). Decision 7 requires a flat finished month to show 0.0 and 0 calories. MonthlyBalance uses firstTrendKG - lastTrendKG, and TestWritePDFCarryOnlyNotEmpty locks Monthly loss: 0.0 kg on that path."
    disposition: accepted
    disposition_rationale: >
      Accepted. A flat fit prints -0.0 because the slope numerator
      is an exact +0 and the unary minus stores a negative zero.
      The %.1f format prints that sign, and rounding that zero is
      0 calories. Decision 7 asks for 0.0 and 0 calories. Clear the
      sign when the fitted loss is zero, and lock a flat fit so the
      caption is 0.0. WindowRate already prints 0.0 for equal
      trends. The PDF stays on its path.
  - id: O2
    category: risk
    severity: medium
    claim: "The shared one-decimal loss format turns fixture C's 0.187 kg into 0.0 st, so a stone caption shows no loss beside a 50 cal deficit."
    evidence: "internal/tui/chart.go: return fmt.Sprintf(\"Monthly loss: %.1f %s   Daily deficit: %d cal\", loss, unit, kcal). internal/tui/longchart.go uses the same %.1f for Loss. internal/units/units.go converts stone as kg / (poundsPerStone * poundInKG) with poundsPerStone = 14. Fixture C's printed loss is 841/4495 kg."
    disposition: rejected
    disposition_rationale: >
      Rejected. Stone is already one decimal, and the calorie
      figure stays on the unrounded pound rate when the display
      unit changes. Fixture C is 841/4495 kg, about 0.029 st, so
      one decimal is 0.0 st beside 50 cal. The kilogram line stays
      0.2 and the pound line stays 0.4. There is no stone-only
      format, and the calories are not recomputed from the rounded
      loss.
  - id: O3
    category: risk
    severity: medium
    claim: "AGENTS.md and the standing idea.md paragraphs still require the pencil identity and call this rate unshipped, which contradicts the caption the screen now prints."
    evidence: "AGENTS.md: \"Decision: Monthly Loss is first minus last trend of the plotted span; Daily Deficit is that loss in pounds × 3500 ÷ days in the span.\" idea.md still says a monthly chart takes those numbers \"from the first and last trend of the plotted span\" and that the Online-caption spec \"is drafted, not shipped.\" internal/tui/chart.go monthCaption instead calls WindowRate(LastSeven(...)) or FitRate and does not call span.Analysis."
    disposition: accepted
    disposition_rationale: >
      Accepted. AGENTS.md still states the pencil identity as the
      current architecture decision. Replace that decision with the
      on-screen rate: last seven for the current month, a
      straight-line fit otherwise, calories from the unrounded
      daily pound rate, and the PDF left on the pencil identity.
      idea.md stays as it is until merge. The caption record
      already describes the new rate and says it is not shipped,
      and the older paragraphs stay history.
---

## O1 — implementation — high

### Claim

FitRate negates a positive-zero slope numerator, so a flat finished
month and a flat long-term window print a loss of -0.0 while
calories stay 0.

### Evidence

The slope step in `internal/dailylog/analysis.go` is:

```go
daily := -(nf*sumXY - sumX*sumY) / den
return daily * gap, kcalFromDailyKG(daily), true
```

When every trend value is the same number, `nf*sumXY - sumX*sumY` is
an exact positive zero. For a finished November of carried 80 kg,
the sums are integers (`Σx = 435`, `Σy = 2400`, `Σxy = 34800`) and
that difference is `+0`. The unary minus stores a negative zero.
Multiplying by the positive gap (`29`) keeps it. `units.FromKG`
returns kilograms unchanged, and dividing by the pound keeps the
sign for pounds and stone.

Both screens print that value with `%.1f`:

```go
fmt.Sprintf("Monthly loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal)
```

`%.1f` of a negative zero is `-0.0`. `kcalFromDailyKG` then does
`int(math.Round(...))` on a negative zero, which is `0`, so the
calorie side does not take the minus. The line is `Monthly loss:
-0.0 kg   Daily deficit: 0 cal`. The long-term screen uses the same
`FitRate` result and the same verb, so a flat quarter does it too.

Decision 7 requires the other sign: a finished month whose carried
trend is flat "still shows 0.0 and 0 calories". The PDF still does.
`MonthlyBalance` sets `lossKG = firstTrendKG - lastTrendKG`, and
equal endpoints are positive zero. `TestWritePDFCarryOnlyNotEmpty`
locks `Monthly loss: 0.0 kg   Daily deficit: 0 cal` for an 80 kg
carry into June. `TestChartCarryOnlyIsNotEmpty` builds the same
shape (80 kg on 31 October, November chart, no November weight) and
never reads the caption. With today outside November that chart
takes the fit, not the last seven. No fit test passes a flat
series. The last-seven path is a different expression:
`trends[0] - trends[n-1]` of equals is positive zero, which is why
fixture A prints `0.0`.

### Why this matters

US4 makes the sign the difference between a loss and a gain. A flat
carry is the ordinary steady state, not a corner fixture: one
earlier weigh-in, then blank days. After the month is no longer
current, the screen shows a gain's sign and the PDF of that month
shows `0.0`. The calorie figure stays `0`, so the line also breaks
the pair "negative loss and negative calories." Every numbered
scenario still passes, because none of them is a flat fit. Shipping
that way leaves Decision 7 true on paper and false on the screen
that the decision was written for.

## O2 — risk — medium

### Claim

The shared one-decimal loss format turns fixture C's 0.187 kg into
0.0 st, so a stone caption shows no loss beside a 50 cal deficit.

### Evidence

The monthly screen formats the converted loss with one decimal and
does not format the calories from that rounded figure:

```go
loss, err := units.FromKG(lossKG, unit)
return fmt.Sprintf("Monthly loss: %.1f %s   Daily deficit: %d cal", loss, unit, kcal), true
```

The long-term line is the same `%.1f` on `Loss:`. Stone conversion
is one division, not a special caption rule:

```go
case Stone:
    return kg / (poundsPerStone * poundInKG), nil
```

`poundsPerStone` is 14, so a stone is 6.35029318 kg. Fixture C's
unrounded loss is `29 * 435/67425 = 841/4495` kg, about 0.1871 kg.
That is about 0.0295 st. One decimal of 0.0295 is 0.0. The deficit
stays on the pound rate, which the spec fixes at 50. The stone line
is `Monthly loss: 0.0 st   Daily deficit: 50 cal`.

S7 does not collapse. `27/55` kg is about 0.077 st and prints 0.1.
The failure is a loss under half a stone-tenth, about 0.32 kg.
Fixture C is that case. The kilogram line prints 0.2 and the pound
line prints 0.4, and those are the only display units the tests set.
Nothing calls `monthCaption` or `longAnalysis` with `units.Stone`.

### Why this matters

Fixture C exists to show that the fit is 0.2 kg, not the flat 0.0
and not the pencil 1.0. In stone those two different losses print as
the same `0.0`, and the deficit still says the trend fell. The
how-to says the printed loss is the change over the days that
entered the rate, and that the deficit is 3500 kcal per pound of the
unrounded daily rate. A stone reader of fixture C sees a zero change
and a 50 cal deficit on the same line. S10 locks the pound path
(3.2 lb, not 3.1) and does not see this, because a pound tenth is
fine enough for 0.187 kg and a stone tenth is not. The assumption
sits in the shared format, not in a stone scenario.

- **accept-as-stated** — one decimal is the stone rule too. Write
  down that a loss under about 0.32 kg prints `0.0 st` beside a
  non-zero deficit.
- **revise-spec** — the one-decimal weight print is the wrong grain
  for this loss. The spec names a stone format that keeps fixture C
  off zero.
- **add-test** — whatever figure is chosen, a stone fixture C locks
  it, including whether `0.0 st` with `50 cal` is allowed.
- **consciously-carry** — stone users get a zero loss on small
  months, kilograms and pounds stay as specified, and that ships on
  purpose.

## O3 — risk — medium

### Claim

AGENTS.md and the standing idea.md paragraphs still require the
pencil identity and call this rate unshipped, which contradicts the
caption the screen now prints.

### Evidence

`monthCaption` does not call `span.Analysis`. For the month that
contains today it uses the last seven. Every other month uses the
fit:

```go
if year == ty && month == tm {
    lossKG, kcal, ok = dailylog.WindowRate(dailylog.LastSeven(logs, today))
} else {
    // x is the point index of each day with HasTrend
    lossKG, kcal, ok = dailylog.FitRate(x, y)
}
```

`longAnalysis` always calls `FitRate`. The PDF still calls
`span.Analysis`, and the three how-tos describe that split.

The files agents are told to follow do not. `AGENTS.md` states a
present-tense architecture decision:

> Decision: Monthly Loss is first minus last trend of the plotted
> span; Daily Deficit is that loss in pounds × 3500 ÷ days in the
> span. Reason: the book's pencil identity.

`idea.md` still specifies the same formula as product behaviour, not
as history: a monthly chart shows Monthly Loss and Daily Deficit
"from the first and last trend of the plotted span", and long-term
Loss and Daily Deficit "use the same first-and-last trend identity
over the window's day count." The Online-caption change record in
that file says "This spec is drafted, not shipped." `CHANGELOG.md`
under Unreleased already describes the last-seven rate and the fit.

`CLAUDE.md` says behaviour changes update `idea.md` when the product
intent changes, and that humans curate `AGENTS.md`. This branch
changes the intent. The standing paragraphs were not replaced, and
the change record still says the work is not shipped.

### Why this matters

The next session is instructed to follow `AGENTS.md` and to treat
`idea.md` as the backlog. Both still say the screen's numbers are
the pencil identity. The how-tos say otherwise, but they are not the
files the agent preamble loads as decisions. A later edit can "fix"
`monthCaption` back to `span.Analysis` and match every standing
decision while undoing S1, S5, and S7. The "not shipped" sentence
makes that look like finishing the backlog rather than reverting the
branch. The spec named the how-tos, the PDF how-to, and left
`docs/reference` and `keys.md` alone. It did not say these two
files would keep ordering the old rate after the code landed.

- **accept-as-stated** — record the new rate in `AGENTS.md`, and
  make the `idea.md` change record say the standing pencil paragraphs
  are history. The next reader inherits a decision.
- **revise-spec** — the screen should still be the pencil identity.
  Then `monthCaption` is the defect, not the docs.
- **add-test** — a test cannot keep a prose decision from
  contradicting the code. This framing does not close it.
- **consciously-carry** — the how-tos and the changelog are the
  user-facing record, `AGENTS.md` stays on the pencil identity until
  a human promotes a reflection, and the contradiction ships on
  purpose.

## Explicitly not objecting to

- **The PDF staying on `span.Analysis`:** `chartpdf.Write` still
  calls it, and `TestWritePDFOnlineFixturesStayPencil` locks 257 and
  772 against 50 and 0. That split is the accepted product behaviour.
- **The last-seven gap and the OLS gap:** `WindowRate` divides by
  `n-1`, and `FitRate` multiplies by `x[n-1] - x[0]`. The tests lock
  3858, 7716, 1.0 rather than 30.0, and 0.5 rather than 3.9. Those
  are the accepted resolutions of spec-time O1 and O2.
- **Calorie rounding:** `kcalFromDailyKG` is `math.Round` of the
  daily kilogram rate over `0.45359237` times 3500, with no second
  division. The gain fixture locks -3858. The duplicated
  `poundInKG` matches `internal/units` and the plan forbids that
  import.
- **A long-term window that grows with today:** trailing carry days
  enter the fit because `sheetRange` marks them `HasTrend` and
  `longEnd` clips `lastDay` to today only inside the current month.
  Decision 4 already counts a blank day that has a trend. S7 is that
  rule on 10 November.
- **Fitting buckets instead of days:** `longView` passes the
  unbucketed `span` to `longAnalysis` and the bucketed `cols` only
  to `renderLongPlot`. S7's quarter is 71 days on a 72-column
  default, so that test does not exercise a bucket, but the call
  site does not use one.
- **Renumbering trend days:** both callers append the span index,
  not `len(x)`. A leading hole does not change S6's gap of 1 either
  way. Carry fills every day after the first weight, so a middle
  hole does not arise on an `ApplyTrend` series.
- **The other three long-term kinds:** they call the same
  `longAnalysis`. For the S7 logs the trend days are consecutive, so
  the gap stays 9 on a longer offset.
- **`trendOn` requiring chronological logs:** the comment says so,
  `ApplyTrend` rejects unsorted input, and `app.load` replaces
  `a.logs` with that result. An unsorted slice would also break
  `chartspan`'s existing carry walk.
- **The how-to's "fewer than seven":** with carry, a previous month
  contributes only when today is early enough that the seven
  calendar days cross the boundary. A series that starts this month
  has nothing earlier to include. The sentence matches that walk.
- **Stale package preambles:** `chartspan` still calls `Analysis`
  "the TUI copy", and `chartpdf` still says "the same loss copy as
  the TUI." The plan left both files unedited. The how-tos are the
  correction the spec required. O1 is the flat caption those
  preambles would hide, not a second doc pass.
