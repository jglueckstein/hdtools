# Plan — Monthly Online chart

**Spec**:
[`docs/superpowers/specs/2026-10-08-monthly-online-chart.md`](../specs/2026-10-08-monthly-online-chart.md)
**Slice**: S2 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)
**Status**: approved
**Objections**:
[`monthly-online-chart.md`](../objections/monthly-online-chart.md)
(O1–O3 accepted)

No production code until the spec's scenarios exist as failing tests.

This slice replaces the on-screen monthly Braille-cell plot with the
Online picture. It does not change the pencil caption, the long-term
Braille plot, or the PDF. Fetch the current v2 release of
`github.com/NimbleMarkets/ntcharts/v2`. Do not pin a patch in the
spec. Do not raise the `go 1.27.0` line. That module's own go line
is 1.26.8, which this module can build. If the fetched release will
not build on go 1.27.0, stop and report. Do not bump the go line to
make it build.

## Module structure

| File | Why |
| --- | --- |
| `go.mod`, `go.sum` | Fetch the current `github.com/NimbleMarkets/ntcharts/v2`. Do not change `go 1.27.0`. BubbleZone may arrive as an indirect dependency. Do not import it. |
| `cmd/hdtools/main.go` | No edit. `tea.NewProgram` still receives the model and nothing else. No `AltScreen`. No `MouseMode`. |
| `cmd/hdtools/main_test.go` | Retarget `TestModuleRequiresCharmV2` as below. Add `TestProgramIsModelOnly`. |
| `internal/tui/onlinechart.go` | New. One concern: paint the monthly Online picture on a fixed 62-column, 16-row plot with a 5-column gutter. Narrative preamble. |
| `internal/tui/onlinechart_test.go` | New scenarios. Co-located. |
| `internal/tui/chart.go` | Call the new painter. Drop the monthly Braille plot and the title box. Keep keys, PDF, the caption, the database line, and the help line. |
| `internal/tui/chart_test.go` | Retarget the monthly paint assertions listed below. Leave PDF, keys, and the pencil tests that already match. |
| `internal/tui/braille.go` | Preamble only: the long-term chart still uses this grid; the monthly chart does not. Do not change the bin rules. |
| `internal/tui/style.go` | No new role. The heading uses the existing `title` style. The painter uses the existing `weight` and `trend` styles. Stem green stays chrome. |
| `docs/how-to/monthly-chart.md` | Drop the month-year box and the ±2 lb margin. Describe the plain heading and the one-step scale. |
| `docs/how-to/long-term-chart.md` | Keep the ±2 lb margin. Stop saying it matches the monthly chart. |
| `docs/how-to/monthly-chart-pdf.md` | Keep the existing vector page and its title box. Stop saying the PDF is the on-screen chart. |

`app.go` is 493 lines. The 500-line check flags non-test Go files.
Do not put this painter in `app.go`. `chart.go` is about 238 lines
and should shrink once `renderPlot` and the title box leave.
`longchart.go` is 416 lines. Do not put the monthly painter there.
`onlinechart.go` has one concern. If it would pass 500 lines, stop
and split a second concern instead of growing it.

The new file's preamble says why it exists (the monthly Online
picture, so `chart.go` and `app.go` stay under the line limit and
the long-term painter stays put), the decisions (display-unit
canvas, fixed 62-column plot, five-column gutter, one share of that
plot per calendar day, trend then stems then diamonds, stem from two
points, no mouse), and what it does not do (resize with the
terminal, change the long-term width or its `%5.1f-` gutter, PDF,
caption formula, regression, new colour keys, grid, legend, zoom).

Do not edit `internal/dailylog` (trend or analysis),
`internal/chartspan` (`YRange` and `YPad` stay the ±2 lb rule the
PDF and the tests of that package use), `internal/chartpdf`,
`internal/units`, `internal/config/colors.go`, `internal/tui/longchart.go`
paint, `docs/reference`, or `keys.md`. Do not add
`github.com/sajari/regression`. When the behaviour is implemented,
`CHANGELOG.md` gets an Unreleased bullet. Do not edit the slicing
record. Do not edit the earlier chart specs.

## Algorithm notes

-   Imports. Fetch the current v2 release. The import path is
    `github.com/NimbleMarkets/ntcharts/v2` or a package under it.
    There is no function named `DrawLine`. The stem is
    `canvas/graph` `GetLinePoints` on the two points, then
    `DrawLinePoints`. Do not use `DrawVerticalLineDown`,
    `DrawVerticalLineUp`, or `DrawLineSequenceLeftToRight` (that one
    assumes the next X is adjacent and does not draw the first
    point). The Braille trend may use `BrailleGrid`,
    `DrawBraillePatterns`, `DrawBrailleRune`, or a timeseries
    `DrawBrailleAll`. If a fetched release does not provide the
    primitive you need, stop and report. Do not substitute `DrawLine`
    or the three forbidden helpers.
-   Do not copy a quickstart that enables mouse, zoom, or BubbleZone.
    BubbleZone is only a library dependency of ntcharts. No file
    under `cmd/` or `internal/` imports a path containing
    `bubblezone`. If the only widget that can draw the series forces
    mouse, zoom, a grid, or a legend, use the lower-level canvas
    instead. If that still cannot draw without those, stop and
    report. `cmd/hdtools/main.go` stays `tea.NewProgram(model)` only.
-   Series. X is the plotted day index, the same X for that day's
    weight and that day's trend, so the stem is the line between
    those two points and not a line to a neighbour. Y is the display
    unit from `units.FromKG`, not stored kilograms. The trend point
    is the `ApplyTrend` value already on the log, including carry
    from before the month. Do not recompute the trend from the
    month's rows alone. Do not fit a line through `WeightKg`. Skip
    days after today; `chartspan.Clip` already does that. Do not
    change `Clip`.
-   Draw order. Braille trend for every plotted day that has a
    trend, including blank days. Then stems. Then the diamond rune
    U+25C7 at each weighed day's weight. The stem rune is not
    Braille (not U+2800 through U+28FF) and is not U+25C7, so the
    three shapes differ with colour off. U+2502 is a fitting rune.
    When the display-unit weight and the display-unit trend are the
    same value, do not draw a stem. Do not draw a stem on a blank
    day.
-   Colour. Trend uses the palette's `trend` style (default red,
    bold). Diamond uses the palette's `weight` style (default blue,
    not bold). Stem green is chrome, the same green as today's
    `stemCell` (`lipgloss.Color("2")`), not bold, and not a palette
    role. When `NO_COLOR` is non-empty, skip chromatic foreground on
    the stem the way the palette already skips it, and keep the
    trend bold. Do not add `diamond` or `stem` to
    `internal/config/colors.go` or to `defaultColors`. Unknown keys
    are already dropped by the loader. An invalid `trend` never
    reaches the palette; the default red remains.
-   Scale. Do not call `chartspan.YRange` for this picture. That
    function still adds the ±2 lb pad for the PDF. Let `lo` and `hi`
    be the min and max of plotted display-unit weights and trends.
    The step is 1 for kg and lb, and 1/14 stone for st. The canvas
    runs from `lo - step` to `hi + step`, even when those ends are
    not on a tick. Ticks are the multiples of `step` inside that
    closed interval. Do the bounds in the unit's own step, in
    pounds for stone (integer pounds), not by adding `1/14` in a
    loop. For stone, the lowest pound is `ceil(minLb - 1)` and the
    highest is `floor(maxLb + 1)`. Label each pound as one decimal
    of stones with the same `%.1f` the sheet uses. Kg and lb labels
    are integers with no `.0`. The high tick is the top label. End
    labels are always drawn. S6 and S7 ranges show every tick,
    including both stone labels that format as `12.1` and both that
    format as `12.4`. S8 (40 kg through 120 kg) has more ticks than
    the 16 plot rows, so it omits at least one tick between 39 and
    121. A tick never
    shares a row with another tick. S6 and S7 fit, so they show
    every tick.
-   Heading and chrome. Replace `monthYearBox` / `titleBox` with the
    plain `monthYearLabel` in the existing `title` style: bold,
    default cyan, no border, no background. A `[colors]` title entry
    changes the heading and does not change the trend. Center the
    words across the 67-column plot line, with space on both sides.
    The heading line is not a plot row. `NO_COLOR` keeps the words
    and the bold. Do not keep the yellow-on-blue box. Keep the
    database line and `chartHelp`.
    Keep `chartspan.Analysis` for the caption. Do not change its
    format string. Day labels sit under the plot. Always label the
    first and last plotted day numbers. A day that is not plotted is
    not labeled, including the empty last calendar day. Other plotted
    days may be labeled. The plot is 62 terminal columns, not
    counting the gutter, for every month length. Do not grow or
    shrink it with the terminal. Do not wait for a resize, and do
    not add a resize test. Each calendar day gets one contiguous
    share of those 62 columns, in order from plot column 1 through
    plot column 62, with no inset. A 31-day month is exactly two
    columns per day. Other lengths divide 62 as evenly as whole
    columns allow. Do not pin which days receive a remainder column.
    Slot boundaries depend only on the number of days in the month,
    not on today. The line, the stems, and the diamonds stop at
    today. Days after today are empty slots and still occupy their
    share. The gutter is five visible columns on every plot row and
    stays fixed: four right-aligned tick columns and a hyphen. The
    plot row is 67 visible columns. Measure with `lipgloss.Width`,
    not bytes. `  79-`, ` 121-`, and `12.6-` are the gutter shapes.
    The long-term chart keeps its current width and its `%5.1f-`
    gutter. Do not make it 62 or 67 columns, and do not retarget
    `chartYBounds` or `isPlotGutter`. An empty month has no plot. A
    carry-only month uses the 62-column plot. The plot is 16 rows
    whenever it has a picture, including a carry-only month. An
    empty month has no plot. Do not follow the terminal. The S3
    gaps, 18 kg and 9 kg, each leave a stem row between the diamond
    and the trend. Two different values that land on the same row
    have no stem. Do not add a fixture whose only job is a 1 kg gap.
    The `idea.md` change record already says the monthly screen
    drops the box and the shared margin. Do not rewrite the older
    backlog paragraphs. Edit the three how-tos as the module table
    says. Do not redraw the PDF or the long-term chart.
-   `TestModuleRequiresCharmV2`. Keep the go line at `1.27.0`. Keep
    the three `charm.land` v2 requirements. Keep the failures for
    `github.com/charmbracelet/bubbletea`,
    `github.com/charmbracelet/lipgloss`, and
    `github.com/charmbracelet/bubbles`, as requirements and as
    imports. Require `github.com/NimbleMarkets/ntcharts/v2`, and
    require at least one import under `cmd/` or `internal/` of that
    module or a subpackage. Fail any other module or import whose
    path contains `ntcharts` (v1 has no `/v2` suffix). Fail a
    requirement on `github.com/sajari/regression`. Fail an import
    whose path contains `bubblezone`. Do not scan docs; they name
    the module in prose. `TestProgramIsModelOnly` reads `main.go`
    and fails if `tea.NewProgram` is passed more than the model, or
    if the file mentions `AltScreen` or `MouseMode`.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestOnlineDiamondAndTrend`, `TestOnlineBlankDayHasTrendNoDiamond`, `TestOnlineWorkoutSameDiamond` |
| FR2 | `TestOnlineDiamondAndTrend`, `TestOnlineBlankDayHasTrendNoDiamond`, `TestOnlineCarryFromPreviousMonth`, `TestOnlineTrendNotRawFit` |
| FR3 | `TestOnlineStemFloatAndSinker`, `TestOnlineBlankDayHasTrendNoDiamond` |
| FR4 | `TestOnlinePlainHeading`, existing `TestChartTitleIsMonthYear`, `TestChartTitleSurvivesNoColor`. Retarget `TestChartTitleBoxColours` off yellow-on-blue. |
| FR5 | `TestOnlineAxisKilograms`, `TestOnlineAxisFlat`, `TestOnlineAxisOffGrid`, `TestOnlineAxisPounds`, `TestOnlineAxisStone`, `TestOnlineCrowdedEndTicks` |
| FR6 | `TestOnlineDayLabels`, `TestOnlineCurrentMonthEndsToday`, existing `TestChartCurrentMonthDividesByElapsedDays` |
| FR7 | `TestOnlineDefaultColours`, `TestOnlineColourOverrides`, `TestOnlineIgnoredColourKeys` |
| FR8 | `TestOnlineNoColorKeepsShapes` |
| FR9 | `TestOnlineHexDownshift` |
| FR10 | `TestOnlinePencilLoss`, existing `TestChartShowsMonthlyLossAndDeficit`, `TestChartLossUsesDisplayUnit` |
| FR11 | `TestOnlineLongTermKeepsPad`, existing long-term tests |
| FR12 | Existing `TestChartP*` tests. Do not edit `chartpdf`. |
| FR13 | Existing key tests, `TestProgramIsModelOnly` |
| FR14 | Existing `TestChartEmptyMonth`, `TestChartCarryOnlyIsNotEmpty` |
| FR15 | Updated `TestModuleRequiresCharmV2` |
| FR16 | `TestOnlineFebruaryAnd31DayPlotWidth`, `TestOnlineCurrentMonthEndsToday`, `TestOnlineDayLabels`, `TestOnlineAxisKilograms`, `TestOnlineAxisStone`, `TestOnlineCrowdedEndTicks`, `TestOnlineLongTermKeepsPad`, existing `TestChartEmptyMonth`, `TestChartCarryOnlyIsNotEmpty` |
| FR17 | `TestOnlinePlotHeight`, `TestOnlineCrowdedEndTicks`, `TestOnlineAxisKilograms`, `TestOnlineAxisStone` |
| FR18 | The three how-to edits. No Go test. The `idea.md` change record is already updated. |

## Test list

New tests live in `internal/tui/onlinechart_test.go` unless named
below. Drive `App.Update` and `View`. No real terminal. SQLite uses
`t.TempDir()` through `openStore`, not `:memory:`. Any test that
calls `freezeToday`, `t.Chdir`, or `t.Setenv` stays off
`t.Parallel`. Freeze today for every new monthly fixture. Colour
tests set `NO_COLOR` empty via `enableChroma`. `NO_COLOR` cases set
it non-empty. Do not retarget `mustPlot`, `isPlotGutter`, or
`chartYBounds`; the long-term plot still uses the `%5.1f-` gutter.
Do not widen `isPlotGutter` to the monthly five-column form. Read
monthly ticks with a new helper that does not treat a day number or
the caption as an axis label. That helper measures visible columns
with `lipgloss.Width`, not bytes. A monthly plot row is 67 visible
columns: a 5-column gutter, then 62 plot columns. Plot column 1 is
the first column after the gutter. The long-term chart keeps its
current width and its `%5.1f-` gutter. Do not treat it as 62 or 67
columns.

-   `TestOnlineDiamondAndTrend` — S1. Today 15 December 1990. 80 kg
    on 1 November and 82 kg on 30 November. Two U+25C7, left then
    right. Braille on the days between. No U+25C6.
-   `TestOnlineBlankDayHasTrendNoDiamond` — S2. That chart's day 2
    has Braille, no diamond, and no non-Braille stem rune.
-   `TestOnlineStemFloatAndSinker` — S3. 80 then 100 kg: day 1 has
    no stem, day 2's stem is green and not Braille, and the diamond
    cell is blue not green. 80 then 70 kg: day 2's stem is the same
    green and the diamond is blue. Each of those stems has a row
    between the diamond and the trend. Do not add a 1 kg fixture.
-   `TestOnlineWorkoutSameDiamond` — S4. One workout day and one
    clear day. Same U+25C7, weight colour, no second mark, no
    U+25C6.
-   `TestOnlinePlainHeading` — S5. "November 1990" on the heading,
    bold, default cyan, no background SGR, no border runes, no
    "hdtools", space on both sides of the words. A `[colors]` title
    of magenta changes the heading and leaves the trend red. Not
    yellow on blue. `NO_COLOR` keeps the words and the bold and has
    no chromatic SGR.
-   `TestOnlineAxisKilograms` — S6. Plotted 80 through 82. Labels
    79, 80, 81, 82, 83. High tick on top. The 79 gutter is the five
    visible columns `  79-`.
-   `TestOnlineAxisFlat` — S6. Only 80 kg. Labels 79, 80, 81. Not
    the old ±2 lb span.
-   `TestOnlineAxisOffGrid` — S6. Plotted 80.4 through 82.4. Labels
    80, 81, 82, 83. Not 79 or 84.
-   `TestOnlineAxisPounds` — S6. 170 lb through 175 lb. Labels 169
    through 176, no decimal. Does not change the long-term half.
-   `TestOnlineAxisStone` — S7. 12.0 st through 12.5 st. Bottom to
    top: 11.9, 12.0, 12.1, 12.1, 12.2, 12.3, 12.4, 12.4, 12.5, 12.6.
    No "st" and "lb" pair on a tick. The 12.6 gutter is the five
    visible columns `12.6-`.
-   `TestOnlineCrowdedEndTicks` — S8. 40 kg through 120 kg. End
    labels 39 and 121. The 121 gutter is the five visible columns
    ` 121-`. The plot is 16 rows, so at least one middle tick is
    omitted. No two tick labels share a row. Fail if 38 or 122 is
    an axis label.
-   `TestOnlineDayLabels` — S9 finished month. Day labels include 1
    and 30. The plot is 62 visible columns and the day slots use
    that full width.
-   `TestOnlineCurrentMonthEndsToday` — S9. Today 10 November 1990.
    Weights 80, 70, and 60 kg on days 1, 2, and 15. Two diamonds.
    Day labels include 1 and 10, and do not include 15 or 30. Each
    plot row is 67 visible columns. Plot columns 23 through 62 are
    spaces: no diamond, no stem, and no Braille. Do not require ink
    in columns 1 through 22. Caption is 1.0 kg and 772 calories, not
    257.
-   `TestOnlineFebruaryAnd31DayPlotWidth` — S21. Today 15 March 1990.
    1990 is not a leap year. February has 80 kg on day 1 only.
    January has 80 kg on day 1 and day 31. Both plots are 62 visible
    columns, not counting the gutter. Do not assert which February
    days are wider. January: day 1's diamond is in plot columns 1
    through 2, day 31's diamond is in plot columns 61 through 62,
    and day 31's column is day 1's column plus 60. Do not add a
    resize test.
-   `TestOnlinePlotHeight` — S22. Any pictured month, including the
    S6, S7, and S8 charts and a carry-only month, is 16 plot rows.
    No two tick labels share a row. S6 and S7 show every tick. S8
    omits at least one middle tick. An empty month has no plot. Do
    not add a resize test.
-   `TestOnlineCarryFromPreviousMonth` — S10. 90 kg on 31 October,
    80 kg on 1 November, today 15 December 1990. Stem on 1 November.
    A later blank day has Braille and no diamond. End labels 79 and
    90, not 91.
-   `TestOnlineTrendNotRawFit` — S11. 80 then 100 kg. Day 1 no stem,
    day 2 has a stem. End labels 79 and 101.
-   `TestOnlineDefaultColours` — S12. Diamond blue (index 4), blank
    day trend red (index 1), stem green (index 2). Read indexes on
    the 16-color write.
-   `TestOnlineColourOverrides` — S12. `weight` magenta, `trend`
    yellow. Diamond magenta, blank-day trend yellow, stem green.
-   `TestOnlineIgnoredColourKeys` — S12. TOML sets `diamond`,
    `stem`, and an invalid `trend`. Diamond stays the weight colour,
    stem stays green, trend stays red.
-   `TestOnlineNoColorKeepsShapes` — S13. `NO_COLOR` set. No
    chromatic SGR on the view text. Diamond, Braille, and stem rune
    present. Blank-day Braille is bold. Diamond is not bold.
-   `TestOnlineHexDownshift` — S14. `trend = "#123456"`. Truecolor
    write contains `38;2;18;52;86`. 16-color write does not contain
    `38;2`. Do not assert that the view was pre-indexed.
-   `TestOnlinePencilLoss` — S15. 80 then 70 kg, today 15 December
    1990. Kilograms: 1.0 kg and 257 cal. Pounds: 2.2 lb and 257 cal.
    Not a seven-day caption.
-   `TestOnlineLongTermKeepsPad` — S16. After the pound monthly
    chart, `l` still shows Braille, no U+25C7, no stem rune, and
    `chartYBounds` still reads about 168 through 177. The long-term
    width and the `%5.1f-` gutter stay as they are. The long-term
    plot is not 62 or 67 columns. Do not retarget `chartYBounds`.
-   `TestProgramIsModelOnly` — S18. `main.go` calls
    `tea.NewProgram` with the model only, and does not mention
    `AltScreen` or `MouseMode`.

Retarget, do not keep the old monthly assertion:

-   `TestChartDailyMarksAndTrendPath` — diamonds plus Braille, not
    left and right dots in every cell.
-   `TestChartOmitsDailyMarkOnBlankDay` — behaviour stays: carried
    trend, no diamond, no stem. Drop the braille-bin check.
-   `TestChartUsesWeightAndTrendColors` — weight colour on the
    diamond, trend colour on a blank day's Braille. The stem does
    not win the diamond's cell.
-   `TestChartNoColorKeepsTwoSeries` — no chroma; diamond and
    Braille still present.
-   `TestChartStemJoinsMarkToTrend` — stem between that day's weight
    and trend, not interior braille bins.
-   `TestChartNoStemWhenMarkOnTrend` — equal values, no stem rune.
-   `TestChartStemGreen` — stem rune is green; it does not have to
    be the only colour in a shared Braille cell.
-   `TestChartDailyMarkWinsSharedCell` — the diamond glyph and the
    weight colour sit on the weight. Drop the shared-dot mask.
-   `TestChartFlatSeriesHasScale` — labels 79, 80, 81, not a 2×`yPad`
    span.
-   `TestChartYRangePounds` — monthly labels 169 through 176. The
    long-term half stays 168 through 177 via `chartYBounds`.
-   `TestChartYRangeKilograms` — monthly labels for 80 and 81 kg are
    79 through 82. The long-term half stays on `yPad`.
-   `TestChartYRangeStone` — monthly labels as S7, not `yPad`.
-   `TestChartTitleBoxColours` — the box is gone: no background, no
    border. Do not require yellow on blue.
-   `TestChartBrailleBothColumnsNovember`,
    `TestChartBrailleCloseMarksBins`,
    `TestChartBrailleEightByFour`,
    `TestChartBrailleStemInsideRow`,
    `TestChartBrailleStemIsGreen`,
    `TestChartBrailleNoColorKeepsDots` — drop dot-mask, one-column,
    and eight-by-four expectations. Coverage moves to the Online
    tests. Do not leave a test that requires the old monthly grid.
-   `TestChartBrailleNoColorCarryIsBold` — Braille trend stays bold
    on a blank day; the diamond is a different glyph and is not
    bold. Drop the shared-mask rule.

Leave these on their current assertions:

-   PDF: `TestChartPWritesPDF`, `TestChartPUsesPDFDir`,
    `TestChartPWriteFailure`, `TestChartPPDFDirIsFile`,
    `TestChartPIgnoredOnList`, `TestChartPWhileEditingIsText`,
    `TestChartPIgnoredOnForm`, `TestChartPNonStringPDFDir`,
    `TestChartPOverwrite`, `TestChartPNoHomeFails`,
    `TestChartPRelativeDataHomeFails`, `TestChartPDoesNotWriteLogs`.
-   Pencil presence and unit: `TestChartShowsMonthlyLossAndDeficit`,
    `TestChartLossUsesDisplayUnit`. The 257 and 772 figures are the
    new tests above. Do not change `chartspan.Analysis`.
-   Current month clip string: `TestChartCurrentMonthDividesByElapsedDays`.
    It still fails if the view contains both `" 11"` and `" 30"`.
    Day labels for a current month must not invent those.
-   Keys: `TestChartOpensFromMonthAndEscapes`,
    `TestChartOpensFromListAndEscapes`, `TestChartFollowsSelectedRow`,
    `TestChartBracketChangesMonth`, `TestChartEscKeepsBrowsedMonth`,
    `TestChartCWhileEditingIsText`, `TestChartKeysDoNotWrite`,
    `TestGotoTodayIgnoredOnChart`.
-   Empty and carry-only: `TestChartEmptyMonth`,
    `TestChartCarryOnlyIsNotEmpty`. An empty month has no scale and
    no 62-column plot. A carry-only month uses the 62-column plot,
    with Braille and no diamond. If the current body cannot see that
    without the old gutter, extend it without changing the
    empty-state words or the zero caption.
-   Heading words: `TestChartTitleIsMonthYear`,
    `TestChartTitleSurvivesNoColor`. Retarget
    `TestChartTitleBoxColours`: the monthly heading is the `title`
    role, not yellow on blue.
-   Unit name on the database line: `TestChartAxisUsesDisplayUnit`.
-   Long-term, including every `TestLongChart*` Braille test.
    `TestBrailleBinS2Marks`, `TestBrailleBinS4Stem`,
    `TestBrailleBinLastIsYMax`, and `TestBrailleBits` stay. They
    lock the long-term grid, not the monthly picture. They keep the
    current long-term width and the `%5.1f-` gutter. Do not retarget
    them to 62 or 67 columns.
-   `internal/chartspan` tests and `internal/chartpdf` tests. The
    PDF Y range stays the ±2 lb pad.
