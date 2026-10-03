# Plan — High-resolution on-screen charts

**Spec**:
[`docs/superpowers/specs/2026-10-02-high-resolution-charts.md`](../specs/2026-10-02-high-resolution-charts.md)
**Objections**:
[`high-resolution-charts.md`](../objections/high-resolution-charts.md)
(O1–O3, O5, O6 accepted; O4 rejected). Code-mode:
[`high-resolution-charts-code.md`](../objections/high-resolution-charts-code.md)
(O1–O3 accepted).
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

This slice paints the existing monthly and long-term charts with
Unicode Braille inside the character columns they already use. It
does not add ntcharts, and it does not move the app to Bubble Tea v2
or Lip Gloss v2. That upgrade is later work in `idea.md`.

The spec's bin index increases with the value. Bin group 0 is the
bottom of the plot, toward Ymin, so weight still increases upward.
S4's "bin group 2" is floor(bin / 4) == 2, the third group up from
the bottom, not plot line 2 counted from the top of `View`.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/braille.go` | New. Dot bits, the bin function, the cell grid, rune assembly, and the one-color rule. Both plots call it. `longchart.go` is already near the 500-line limit, so the painter does not live there. |
| `internal/tui/braille_test.go` | Bin numbers for S2 and S4, Ymax in the last bin, and the eight bit masks. No TTY. |
| `internal/tui/chart.go` | `renderPlot` fills a Braille grid instead of `o`, `-` / `/` / `\`, and `\|`. `plotRows` stays 8. Y labels, day labels, title, loss, keys, and `p` stay. |
| `internal/tui/longchart.go` | `renderLongPlot` uses the same painter. Bucket count and which series are drawn stay. A trend cell is bold. |
| `internal/tui/chart_test.go` | S1–S5, S9, S10. Retarget tests that look for `o` or `\|`. `plotColumn` must recognize Braille, not `o-/\\|`. |
| `internal/tui/longchart_test.go` | S6–S8. Retarget `TestLongChartTwoLinesNoMarksOrStems`, which looks for `/\-`. |

Do not import a chart module. Do not change `internal/chartpdf`,
`internal/chartspan` (Y range, clip, buckets), help lines, or
`docs/reference/keys.md`. Do not spline and do not fill under a line.

## Algorithm notes

-   `plotRows` stays 8. *N* = 32. One character column per day or
    per bucket, one Braille rune per column.
-   Bin of value *v*, with Ymin and Ymax from the existing padded
    range: *N* − 1 when *v* ≥ Ymax, otherwise
    clamp(floor((*v* − Ymin) / (Ymax − Ymin) × *N*), 0, *N* − 1).
    The old `Round` onto `plotRows-1` inclusive ticks goes away.
    If Ymin equals Ymax, use bin *N* / 2. A real padded range is
    not equal.
-   Screen row from the top is (*R* − 1) − floor(bin / 4). Inside
    the cell, dot row from the top is 3 − (bin mod 4). Left bits
    top to bottom are `0x01`, `0x02`, `0x04`, `0x40`. Right bits
    top to bottom are `0x08`, `0x10`, `0x20`, `0x80`.
-   A mark sets the left dot at its bin. A stem sets the left dots
    strictly between the mark bin and that day's trend bin. Equal
    bins set no stem dot. The trend sets the left dot and the right
    dot at the trend bin in that column. Do not light dots in the
    columns between two days. On a shared dot the monthly mark
    wins and the long-term trend wins.
-   One foreground for the whole rune. Monthly: a stem dot paints
    the cell stem green (the chrome already used for `|`, not a
    new `[colors]` role), even when a mark or trend dot shares
    it. Otherwise mark dot present → `weight`, else `trend`.
    S4's cell is stem green because bins 9 and 10 are in it. A
    cell with no stem keeps the mark, then the trend. Long-term:
    trend dot present → `trend` and bold; else `weight`, not
    bold.
-   Under `NO_COLOR`, strip chromatic SGR and keep the rune. Bold
    stays on the trend role, including a monthly carry day, and
    stays off a coincident weigh-in that shares that mask. A
    stem-filled mark and a pure stem with the same mask are the
    same picture. There is no third style bit.
-   Y labels stay `%5.1f-`, eight of them, top label at Ymax.
    Gutter stays 6 runes. Day labels under the monthly plot stay
    day 1 and the last plotted day. Long-term X labels stay month
    starts.
-   Empty span still returns the empty state and does not build a
    grid. PDF still calls `chartpdf` with the same span. The
    Braille string is not the PDF.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestChartBrailleBothColumnsNovember` |
| FR2 | `TestChartBrailleEightByFour`, `TestBrailleBinLastIsYMax` |
| FR3 | `TestChartBrailleCloseMarksBins`, `TestBrailleBinS2Marks` |
| FR4 | `TestChartBrailleStemInsideRow`, `TestBrailleBinS4Stem`, updated `TestChartNoStemWhenMarkOnTrend` |
| FR5 | `TestChartBrailleNoColorKeepsDots`, `TestChartBrailleNoColorCarryIsBold`, `TestChartBrailleStemInsideRow`, updated `TestChartStemGreen`, updated `TestChartDailyMarkWinsSharedCell`, `TestLongChartBrailleNoColorBoldTrend` |
| FR6 | `TestLongChartBrailleKeepsDayColumns`, `TestLongChartBrailleStaysBucketed` |
| FR7 | Existing title, loss, Y-range, key, and help tests stay |
| FR8 | Existing `TestChartEmptyMonth` stays and must not panic |
| FR9 | Existing `TestChartPWritesPDF` stays; `chartpdf` is untouched |

## Test list

New tests. Any test that calls `freezeToday` stays off `t.Parallel`.

-   `TestBrailleBinS2Marks` — display kg, plotted min 80 and max 82,
    bins 12 and 14 for 80.55 and 80.85.
-   `TestBrailleBinS4Stem` — 80.00 → 7, 80.10 → 8, 80.50 → 11. The
    8-bin coarse index of 80.10 and 80.50 is 2 for both.
-   `TestBrailleBinLastIsYMax` — *v* = Ymax is bin 31; *v* = Ymin
    is bin 0.
-   `TestBrailleBits` — the eight masks in Decision 1, top to
    bottom, left then right.
-   `TestChartBrailleBothColumnsNovember` — S1. Today 15 December
    1990. Weights on 1 Nov and 30 Nov. Thirty Braille columns after
    the gutter. Each rune has a left dot and a right dot.
-   `TestChartBrailleCloseMarksBins` — S2. Today 15 December 1990.
    No logs before November. Weights 80.00, 80.55, 80.85, 82.00 kg
    on days 1, 2, 3, and 30. Day 2's left dot is bin 12. Day 3's
    left dot is bin 14.
-   `TestChartBrailleStemIsGreen` — the day-2 cell holding bin 8
    is stem green. The mark in the next cell keeps the weight
    role. A day with no weigh-in keeps the trend role.
-   `TestChartBrailleEightByFour` — S3. That chart has 8 plot rows.
    Each row's rune has four vertical dot places.
-   `TestChartBrailleStemInsideRow` — S4. Today 15 December 1990.
    Weights 80.00, 80.50, 82.00 on days 1, 2, and 30. Day 1 has no
    stem dot (bin 7 for both). Day 2's left column has stem dots at
    bins 9 and 10, and the mark and trend are bins 11 and 8, both
    in group 2. That rune is stem green.
-   `TestChartBrailleNoColorKeepsDots` — S5. `NO_COLOR` set on the
    S2 chart. No chromatic SGR. Day 2's mark and that day's trend
    are different plot rows. Both runes still have their dots.
    Title still contains the month and year.
-   `TestLongChartBrailleKeepsDayColumns` — S6. Today 10 November
    1990, latest log that day, *W* = 72. Quarterly span is 71 day
    columns, not fewer. The 10 November rune has a left trend dot
    and a right trend dot.
-   `TestLongChartBrailleStaysBucketed` — S7. Today 15 December
    1990, latest log 30 November 1990. Quarterly plot has 72
    columns. The bucket that contains 30 November has a left trend
    dot and a right trend dot. No daily-mark-only pattern is
    required; the daily path is omitted.
-   `TestLongChartBucketOmitsWeightDot` — annual buckets. A column
    is only its trend bin. The 90 kg weigh-in does not add a dot
    at a different bin.
-   `TestLongChartBrailleNoColorBoldTrend` — S8. `NO_COLOR` on the
    S6 chart. No chromatic SGR. A cell with a trend dot is bold. A
    cell with only the daily path is not bold.
-   `TestChartBrailleNoColorCarryIsBold` — S11. One weigh-in on
    the trend and a later blank day. `NO_COLOR`. Same mask. The
    weigh-in is not bold. The carry day is bold.

Retarget, do not delete:

-   `TestChartDailyMarksAndTrendPath` — 8 plot rows of Braille, not
    a search for `o` or `-/\\`.
-   `TestChartOmitsDailyMarkOnBlankDay` — a blank day has the
    carried trend's two dots and no mark dot. The digits 1 and 30
    are not that check.
-   `TestChartUsesWeightAndTrendColors` — weight is magenta on a
    mark cell with no stem dot. Trend is yellow on a Braille
    cell. Stem green is not the weight check.
-   `TestChartNoColorKeepsTwoSeries` — no chroma; both series still
    present as dots.
-   `TestChartStemJoinsMarkToTrend` — stem dots between mark and
    trend, or fold into `TestChartBrailleStemInsideRow` if it
    becomes the same assertion.
-   `TestChartNoStemWhenMarkOnTrend` — equal bins, left dot set, no
    interior stem dot.
-   `TestChartStemGreen` — a cell with only stem dots is green.
-   `TestChartDailyMarkWinsSharedCell` — the shared dot is set, and
    that cell uses the weight color.
-   `TestLongChartTwoLinesNoMarksOrStems` — no `o` and no `|` is
    still true because those runes are gone. Assert Braille dots
    for both paths when they fit, and trend color on a shared cell.
-   `plotColumn` — select plot lines by the `%5.1f-` gutter, then
    read rune 6 + day index. Do not require the line to contain
    `o-/\\|`.

Leave these on their current assertions: open and Esc, brackets,
loss and deficit, Y-range bounds, display unit, empty month, keys
do not write, `p` writes the PDF, long-term kind cycling, and
long-term X labels.
