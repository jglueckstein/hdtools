# Plan — Online caption

**Spec**:
[`docs/superpowers/specs/2026-10-14-online-caption.md`](../specs/2026-10-14-online-caption.md)
**Slice**: S3 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

This slice changes the on-screen caption only. The monthly Online
picture, the long-term Braille-cell plot, and the PDF pencil caption
stay. Do not add `github.com/sajari/regression`. Do not edit `go.mod`
or the `go 1.27.0` line. Do not change `ApplyTrend`.

## Module structure

| File | Why |
| --- | --- |
| `internal/dailylog/analysis.go` | The new rates live beside `MonthlyBalance`, so the monthly screen and the long-term screen share them. Keep `MonthlyBalance` as the pencil identity. Do not import `units`, `chartspan`, or the TUI. |
| `internal/dailylog/analysis_test.go` | The fixture arithmetic, without a screen. Existing `MonthlyBalance` tests stay. |
| `internal/tui/chart.go` | The monthly screen stops calling `span.Analysis` for its caption. The current month uses the last-seven rate. Any other month uses the fit. Update the preamble that says the on-screen caption lives in `chartspan`. |
| `internal/tui/onlinechart_test.go` | Screen captions for S1 through S6, the 30 November reading of S6, S10, and the monthly half of S9. Retarget the two pencil caption assertions. |
| `internal/tui/onlinechart.go` | No edit. The picture stays. |
| `internal/tui/longchart.go` | `longAnalysis` uses the fit on the day span, before bucketing. Keep the `Loss:` words. Do not put the fit loop in this file. It is already about 418 lines. |
| `internal/tui/longchart_test.go` | S7, including its pound line, and the long-term half of S9. Retarget `TestLongChartLossAndDeficit`. |
| `internal/tui/chart_test.go` | No new numbers. Leave the word checks and the carry-only picture check. |
| `internal/chartspan/chartspan.go` | No edit. `Analysis` stays the pencil function. |
| `internal/chartspan/chartspan_test.go` | No edit. `TestSpanAnalysisBothEndpoints` still wants the pencil line. |
| `internal/chartpdf/chartpdf.go` | No edit. It keeps calling `span.Analysis`. |
| `internal/chartpdf/chartpdf_test.go` | Existing pencil tests stay, including 257, 772, and the carry-only zero line. Add the S8 fixture test. |
| `internal/dailylog/trend.go` | No edit. |
| `docs/how-to/monthly-chart.md` | When the behaviour is implemented. See below. |
| `docs/how-to/long-term-chart.md` | When the behaviour is implemented. See below. |
| `docs/how-to/monthly-chart-pdf.md` | When the behaviour is implemented. See below. |

`cmd/hdtools/main_test.go` already rejects
`github.com/sajari/regression` and locks `go 1.27.0`. Do not edit it
except to keep those checks if a module edit is ever proposed. Do not
edit `docs/reference` or `docs/reference/keys.md`. Do not edit the
slicing record, earlier specs, or earlier plans. The `idea.md` change
record for this draft is already in place. Do not rewrite the
monthly-picture paragraph that leaves the pencil identity until #79.

When the behaviour is implemented, `CHANGELOG.md` gets an Unreleased
bullet. Not before the tests are green.

The `analysis.go` preamble says why both rates live here (one
arithmetic owner for the screen, the pencil path left for the PDF),
that calories are the unrounded daily pound rate times 3500 with no
second division, and what the file does not do (plot, persist,
convert display units, fit raw weights, import a regression module).

## Algorithm notes

-   Names in `internal/dailylog`. Oldest trend first. `x` is the
    0-based offset from the first day of the span. `ok` is false
    when fewer than two values are available, when the gap from the
    first x to the last x is not positive, or when the fit
    denominator is zero. `WindowRate` has no x. Its gap is one less
    than the count, so two or more values already have a positive
    gap.

    ```go
    func LastSeven(logs []DailyLog, end time.Time) []float64
    func WindowRate(trends []float64) (lossKG float64, kcal int, ok bool)
    func FitRate(x, trendKG []float64) (lossKG float64, kcal int, ok bool)
    ```

-   `LastSeven` walks one calendar day at a time, backward from
    `end`, and collects at most seven trends. Compare dates, not
    clock times. The screen passes `localToday()`, already a UTC
    date. A day with no row still counts when an earlier log carries
    a trend onto it: the last log on or before that day with a
    weight or a non-zero trend, which is the carry the month sheet
    already uses. Copy that predicate here. Do not import the TUI,
    and do not recompute `ApplyTrend`. Stop at seven values or at
    the first day with no trend. Return oldest first. Days before
    the first weight are not skipped in search of an earlier one.
-   `WindowRate` is first minus last. That difference is the
    unrounded kilogram loss. The daily rate is that loss divided by
    one less than the count of `trends`. The loss is not rounded
    inside this function, and it is not multiplied by a span length.
-   `FitRate` is ordinary least squares. Slope `b` is
    `(n*Σxy − Σx*Σy) / (n*Σx² − (Σx)²)` with `n = len(x)`. The daily
    rate is `-b` kilograms per day. The unrounded kilogram loss is
    that rate times the gap from the first x to the last x. `n` is
    not that gap. Do not renumber the points that were passed in.
    Callers pass the span index of each day that has a trend, not a
    fresh 0..n-1 sequence, and not a raw weight.
-   Both rates share one calorie step: `math.Round(dailyKG /
    poundInKG * kcalPerPound)`. Use the constants already in this
    file. Do not divide by the gap or by the window count in that
    step. `math.Round` is half away from zero, including for a gain.
-   Display stays at the call site. `units.FromKG` of the unrounded
    kilogram loss, then the existing format. Monthly:
    `Monthly loss: %.1f %s   Daily deficit: %d cal`. Long-term:
    `Loss: %.1f %s   Daily deficit: %d cal`. Three spaces before
    `Daily`. Do not format from the calorie figure, and do not
    convert the already-printed one-decimal loss back to calories.
-   Monthly screen. `chartspan.Clip` is unchanged and still supplies
    the plotted span. If the viewed month is local today's year and
    month, call `LastSeven` on `a.logs` with that today, then
    `WindowRate`. Otherwise build `x` and `y` from points with
    `HasTrend`, `x` equal to the point index, and call `FitRate`.
    Do not pass the plotted length as a multiplier. Do not call
    `span.Analysis` for this caption. `writeChartPDF` stays on
    `chartpdf`, which still calls `Analysis`. An empty month returns
    before the caption, as it does now.
-   Long-term screen. Build `x` and `y` from the `sheetRange` days,
    before `bucketSpan`. `x` is the index in that day slice,
    including leading days with no trend, and those days are left
    out of the sample rather than renumbered. Call `FitRate` for
    every window, including one that ends today. The printed loss
    uses the gap from the first trend day's x to the last. Delete
    the guard that omits the caption unless both endpoints have a
    trend. A bucketed plot must not become the regression sample.
    Do not change `renderLongPlot`, the ±2 lb scale, the title box,
    or the `%5.1f-` gutter.
-   How-tos, only once the tests are green.
    `docs/how-to/monthly-chart.md` replaces the sentence that the
    two on-screen numbers are the first-and-last identity. The
    current month uses the last seven trend values, carrying into
    the previous month when this month has fewer than seven. Every
    other month uses a straight-line fit to the trend. The printed
    loss is the change over the days that entered the rate. Days
    with no trend are not part of that change. Calories stay
    3500 kcal per pound of the unrounded daily rate. The PDF is not
    that rate. They are not medical advice.
    `docs/how-to/long-term-chart.md` keeps the ±2 lb margin and the
    current picture. Loss and Daily Deficit use the straight-line
    fit the same way: the change along the line from the first
    trend day to the last, not the monthly last-seven rate and not
    the PDF pencil identity. Not medical advice.
    `docs/how-to/monthly-chart-pdf.md` stops saying the PDF has the
    same Monthly Loss and Daily Deficit copy as the screen. It
    keeps the pencil identity and says the on-screen caption is a
    different rate.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestLastSevenFixtureA`, `TestOnlineFixtureA`, retargeted `TestOnlineCurrentMonthEndsToday` |
| FR2 | `TestWindowRateFixtureB`, `TestOnlineFixtureB` |
| FR3 | `TestWindowRateGain`, `TestOnlineGainIsNegative` |
| FR4 | `TestLastSevenCarry`, `TestWindowRateCarry`, `TestOnlineCarryFeedsRateOnly` |
| FR5 | `TestFitRateFixtureC`, `TestOnlineFixtureC` |
| FR6 | `TestFitRateLateStart`, `TestOnlineFitWithoutFirstDay`, `TestLastSevenLateStart`, `TestWindowRateConsecutivePair`, `TestOnlineCurrentMonthLateStart` |
| FR7 | `TestFitRateQuarterlySpan`, `TestLongChartFitWhileOldPicture` |
| FR8 | `TestWindowRateOmitsFewerThanTwo`, `TestFitRateOmitsFewerThanTwo`, `TestOnlineOmitsShortSeries`, `TestLongChartOmitsShortSeries` |
| FR9 | `TestWritePDFOnlineFixturesStayPencil`, existing `TestWritePDFLossAndDeficit`, `TestWritePDFCurrentMonthClipsAtToday`, `TestSpanAnalysisBothEndpoints` |
| FR10 | The three how-to edits. No new Go test. Existing `TestModuleRequiresCharmV2` stays the go-line and regression-module guard. |
| FR11 | `TestFitRateUnroundedLoss`, `TestOnlineUnroundedPounds` |

## Test list

New `internal/dailylog` tests are table-free and may run in parallel.
They call `ApplyTrend` rather than hand-writing a carried series,
except where a test passes trend values straight into `WindowRate` or
`FitRate`. Assert the whole calorie count exactly. Assert
`fmt.Sprintf("%.1f", lossKG)` for the one-decimal kilogram figure. Do
not require the kilogram loss to be bit-identical to a decimal
literal. Do not call `units.FromKG` from this package.

Screen tests drive `App.Update` and `View`. No real terminal. SQLite
uses `t.TempDir()` through `openStore`, not `:memory:`. A test that
calls `freezeToday`, `t.Chdir`, or `t.Setenv` stays off `t.Parallel`.
Match the whole caption line, including the three spaces. Monthly
fixtures use `novemberPlot` unless the test has to open November
from an October log. Colour is irrelevant to these captions.

-   `TestLastSevenFixtureA` — S1 walk. 80 kg on 1 November and 70 kg
    on 2 November. Ending 10 November returns seven values, all 79.
    The 80 on day 1 is not in the window. The logs have only those
    two rows; the blank days still count.
-   `TestWindowRateFixtureA` — seven 79s. Loss formats `0.0`,
    calorie count 0. Not the pencil 772.
-   `TestOnlineFixtureA` — S1 screen. Today 10 November 1990. No
    weight after day 2. Caption
    `Monthly loss: 0.0 kg   Daily deficit: 0 cal`. Not 772, not 421,
    not `1.0 kg`, not `0.5 kg`.
-   `TestOnlineCurrentMonthEndsToday` — keep the picture assertions.
    The caption becomes the same zero line. Not 772 and not 421. The
    day-15 weight stays unplotted.
-   `TestWindowRateFixtureB` — trends 80, 79, 79. Loss formats
    `1.0`, calorie count 3858. Not 257 and not 2572.
-   `TestOnlineFixtureB` — S2. Today 10 November 1990, 80 kg on day 8
    and 70 kg on day 9 only. Kilograms:
    `Monthly loss: 1.0 kg   Daily deficit: 3858 cal`. Pounds:
    `Monthly loss: 2.2 lb   Daily deficit: 3858 cal`, and the loss
    is not in kilograms. Not 257 calories, not 3.3 kg, and not
    7.3 lb.
-   `TestWindowRateGain` — trends 70, 71, 71. Loss formats `-1.0`,
    calorie count −3858.
-   `TestOnlineGainIsNegative` — S3. 70 kg on day 8 and 80 kg on day
    9. Caption
    `Monthly loss: -1.0 kg   Daily deficit: -3858 cal`.
-   `TestLastSevenCarry` — S4 walk. 80 kg on 30 October and 70 kg on
    31 October. Ending 4 November returns 80 then five 79s. No
    seventh value.
-   `TestWindowRateCarry` — those six values. Loss formats `1.0`,
    calorie count 1543. Not 0.7 and not 1286.
-   `TestOnlineCarryFeedsRateOnly` — S4 screen. Freeze today at 4
    November 1990. The only logs are in October, so `c` opens
    October. Press `]` once and read November. October's own fit is
    not this caption. Caption
    `Monthly loss: 1.0 kg   Daily deficit: 1543 cal`. Heading says
    November 1990. No open diamond. Not `0.0 kg`, not `0.7 kg`, not
    1286 calories, and not 0 calories.
-   `TestFitRateFixtureC` — x = 0..29, y = 80 then 79 twenty-nine
    times. Loss formats `0.2`, calorie count 50. The unrounded loss
    is below 0.19, so the multiplier is the 29-day gap and not 30.
    Not 257. Pass all 30 points. Two weighed days alone are the
    wrong sample. Do not pass a span length.
-   `TestOnlineFixtureC` — S5. Replaces `TestOnlinePencilLoss`. Today
    15 December 1990, 80 kg on 1 November and 70 kg on 2 November.
    Kilograms: `Monthly loss: 0.2 kg   Daily deficit: 50 cal`.
    Pounds: `Monthly loss: 0.4 lb   Daily deficit: 50 cal`. Not
    `1.0 kg`, not `2.2 lb`, not 257, not the zero line. Do not leave
    a test that still expects the pencil caption on this screen.
-   `TestFitRateLateStart` — x = 28 and 29, y = 80 and 79. Loss
    formats `1.0`, calorie count 7716. Not 30.0.
-   `TestOnlineFitWithoutFirstDay` — S6. Today 15 December 1990. 80
    kg on 29 November and 70 kg on 30 November only. Caption
    `Monthly loss: 1.0 kg   Daily deficit: 7716 cal`. The line is
    present. Not 30.0 kg.
-   `TestLastSevenLateStart` — the same two logs, ending 30 November.
    Returns 80 then 79. It does not walk onto day 28.
-   `TestWindowRateConsecutivePair` — trends 80, 79. Loss formats
    `1.0`, calorie count 7716. Not 3858. The divisor is the one-day
    gap, not the count of two values.
-   `TestOnlineCurrentMonthLateStart` — S6 read while November is
    still current. Freeze today at 30 November 1990. The chart opens
    on November. Caption
    `Monthly loss: 1.0 kg   Daily deficit: 7716 cal`. Not 15.0 kg
    and not 3858 calories. The line is present.
-   `TestFitRateQuarterlySpan` — the ten November trends as x = 0..9
    with y = 80 then nine 79s. Loss formats `0.5` and the unrounded
    loss is below 0.5, so it is the nine-day gap and not the
    ten-trend-day product. Calorie count 421. A 71-day product
    would format `3.9` and is wrong.
-   `TestLongChartFitWhileOldPicture` — S7. Replaces the body of
    `TestLongChartLossAndDeficit`. Today 10 November 1990, 80 kg on
    1 November and 70 kg on 2 November. Press `l` and stay on
    Quarterly. Do not set a narrow width. Caption
    `Loss: 0.5 kg   Daily deficit: 421 cal`. In pounds,
    `Loss: 1.1 lb   Daily deficit: 421 cal`, not 1.2 lb and not
    8.5 lb. The view names Quarterly, September 1990, and November
    1990, and does not say `Monthly loss:`. Braille is present.
    U+25C7 is absent. The top gutter is ` 80.9-` and the bottom
    gutter is ` 69.1-`. Not 3.9 kg. Do not leave the old 1.0 kg and
    86 cal expectation.
-   `TestWindowRateOmitsFewerThanTwo` — one trend value. `ok` is
    false.
-   `TestFitRateOmitsFewerThanTwo` — one point, a zero gap, or a zero
    denominator. `ok` is false.
-   `TestOnlineOmitsShortSeries` — S9 monthly. Today 1 November 1990,
    only 80 kg that day. The view has no `Monthly loss:` and no
    `Daily deficit:`. Not the zero pencil line. The chart is still
    shown.
-   `TestLongChartOmitsShortSeries` — S9 quarterly. The same single
    day. The view has no `Loss:` and no `Daily deficit:`.
-   `TestFitRateUnroundedLoss` — S10 arithmetic. x = 0..29, y is 70
    fourteen times then 69 sixteen times. Loss formats `1.4`,
    calorie count 385. The unrounded loss is above 1.44, so it is
    not the printed tenth treated as exact, and a 30-day product
    would format `1.5`.
-   `TestOnlineUnroundedPounds` — S10 screen. Today 15 December 1990.
    70 kg on 1 November and 60 kg on 15 November only. Kilograms:
    `Monthly loss: 1.4 kg   Daily deficit: 385 cal`. Pounds:
    `Monthly loss: 3.2 lb   Daily deficit: 385 cal`. Not 3.1 lb.
-   `TestWritePDFOnlineFixturesStayPencil` — S8. Fixture C logs,
    today 15 December 1990, PDF text contains
    `Monthly loss: 1.0 kg   Daily deficit: 257 cal` and not
    `Daily deficit: 50 cal`. Fixture A logs, today 10 November 1990,
    PDF text contains
    `Monthly loss: 1.0 kg   Daily deficit: 772 cal` and not
    `Daily deficit: 0 cal`. Use `ApplyTrend` logs, not hand-set
    endpoint trends.

Leave these assertions alone. They already describe behaviour this
slice does not change:

-   `TestMonthlyBalanceLossAndDeficit`, `TestMonthlyBalanceGainIsNegative`,
    `TestMonthlyBalanceElapsedDays`, and
    `TestMonthlyBalanceRejectsNonPositiveDays`.
-   `TestSpanAnalysisBothEndpoints` and the other `chartspan` tests.
-   `TestWritePDFLossAndDeficit`, `TestWritePDFCurrentMonthClipsAtToday`,
    and `TestWritePDFCarryOnlyNotEmpty`.
-   `TestChartShowsMonthlyLossAndDeficit` and
    `TestChartLossUsesDisplayUnit`. They only check that the words
    are present and that pounds are not labelled `kg`.
-   `TestChartCarryOnlyIsNotEmpty`. It checks the picture, not a zero
    caption. A finished flat carry still shows 0 through the fit;
    this test does not need a new number.
-   `TestOnlineLongTermKeepsPad` and the other picture tests. Do not
    retarget `chartYBounds`, `isPlotGutter`, or `mustPlot`.
