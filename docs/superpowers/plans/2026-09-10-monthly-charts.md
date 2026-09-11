# Plan — On-screen monthly weight chart

**Spec**:
[`docs/superpowers/specs/2026-09-10-monthly-charts.md`](../specs/2026-09-10-monthly-charts.md)
**Issue**: [#34](https://github.com/jglueckstein/hdtools/issues/34)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/chart.go` | New screen: title, axes, plot, Monthly Loss, Daily Deficit from `[]sheetDay`. Does not open SQLite. |
| `internal/tui/chart_test.go` | Plot construction: blank days, empty month, unit conversion, mark vs path, analysis line. |
| `internal/dailylog/analysis.go` | Pure first-and-last trend → loss kg and daily kcal. No SQLite. |
| `internal/tui/app.go` | `screenChart`, `c` from list and month, Esc back, `[` `]` while on the chart. |
| `internal/tui/app_test.go` | Open/Esc, month keys, `NO_COLOR`, scheme colours, no writes. |
| `internal/tui/month.go` | Reuse `buildMonthSheet` and `[` `]` month stepping; no plot code here. |
| `internal/dailylog` | Trend stays derived. Add the 3500 kcal/lb identity here so PDF can reuse it later. |
| `docs/how-to/` | How to open the monthly chart. |
| `docs/reference/keys.md` | `c` on list and month. |
| `docs/tutorials/first-log.md` | Optional one-line after the month sheet. |
| `CHANGELOG.md` | Added. |

No new config keys. Palette `weight` / `trend` already exist.

## Algorithm notes

-   **Same sheet data, clipped.** `buildMonthSheet` fills blank days
    and carries trend. The chart reads that slice through
    `lastPlottedDay`: last calendar day if the month is past; today's
    day-of-month if the month contains today; empty if the month is
    entirely after today.
-   **Open.** From list: year/month of the selected log, or `localToday`
    if the list is empty. From month, **when not editing**:
    `a.month.year` / `a.month.month`. While editing, `c` goes to the
    textinput. Remember `afterChart` separately from `afterSave`.
    Chart `[` / `]` call the same `nextMonth` / `prevMonth` on
    `a.month` so Esc to the sheet shows the browsed month.
-   **Y scale.** Min/max of present daily weights (converted) and
    present trends (converted) in the plotted span. Pad a fraction of
    the span. When min equals max, span is 1.0 in the display unit.
    Empty (no marks and no trend): no scale.
-   **X scale.** One character column per calendar day, left to right,
    after a Y-label gutter. Plot height at least 8 rows when there is
    data. Tests index day *N* as gutter + *N*.
-   **Glyphs.** Daily: `o` (or equivalent) per weighed day. Trend: a
    connected ASCII/Unicode path. When both land in one cell, draw
    `o`. Colour with the palette. Under `NO_COLOR`, skip Foreground.
-   **No TTY.** Tests inspect `View()`. Do not require a real terminal.
    Reuse the colour-scheme SGR helpers in `color_test.go` or move them
    to a test util in the same package.
-   **Do not write.** Chart `Update` ignores type/Tab/Space except as
    no-ops. `q` still quits.
-   **Monthly analysis.** First day's `Trend` and last **plotted**
    day's `Trend`. Missing either: skip the line. Loss kg = first −
    last. Display with `FromKG`. Daily kcal =
    `FromKG(lossKG, Pound)` × 3500 / days in the plotted span,
    rounded to nearest integer. Carry-only span: loss 0, deficit 0.
    Empty span: omit the line.

A plotting library is allowed if it stays in `internal/tui` and does not
import `database/sql`. A small grid in `chart.go` is enough; do not add
a layout framework.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1, FR2 | `TestChartOpensFromMonthAndEscapes`, `TestChartOpensFromListAndEscapes`, `TestChartEmptyListUsesToday`, `TestChartCWhileEditingIsText`, `TestChartEscKeepsBrowsedMonth` |
| FR3, FR6 | `TestChartDailyMarksAndTrendPath`, `TestChartOmitsDailyMarkOnBlankDay` |
| FR4 | `TestChartUsesWeightAndTrendColors` |
| FR5 | `TestChartNoColorKeepsTwoSeries` |
| FR7 | `TestChartEmptyMonth`, `TestChartCarryOnlyIsNotEmpty` |
| FR8 | `TestChartBracketChangesMonth` |
| FR9 | `TestChartAxisUsesDisplayUnit` |
| FR10 | `TestChartKeysDoNotWrite` |
| FR11 | copy review in the how-to / title string (no "diagnose", no "eat") |
| FR12, FR13 | `TestMonthlyBalanceLossAndDeficit`, `TestMonthlyBalanceGainIsNegative`, `TestChartOmitsAnalysisWithoutEndpoints`, `TestChartLossUsesDisplayUnit`, `TestChartCurrentMonthDividesByElapsedDays` |
| FR14 | `TestChartOmitsDailyMarkOnBlankDay` (day columns), plot-height assertion in `TestChartDailyMarksAndTrendPath` |
| FR15 | `TestChartFlatSeriesHasScale` |

## Test case list

1.  `TestChartOpensFromMonthAndEscapes` — twoDayApp, `m`, `c`, title
    November 1990, Esc back to month sheet.
2.  `TestChartOpensFromListAndEscapes` — twoDayApp, `c`, November 1990,
    Esc back to list (`daily log` title).
3.  `TestChartDailyMarksAndTrendPath` — two weighed days; View has
    discrete daily marks and a connected trend path.
4.  `TestChartOmitsDailyMarkOnBlankDay` — weights on 1st and 4th; no
    daily mark in the day-2/3 columns; trend still present there.
5.  `TestChartUsesWeightAndTrendColors` — `weight=green`, `trend=yellow`;
    SGR on the chart view.
6.  `TestChartNoColorKeepsTwoSeries` — `NO_COLOR=1`; no chromatic SGR;
    both series present.
7.  `TestChartEmptyMonth` — empty store, `c`; empty-state text; no
    analysis numbers; no panic.
7a. `TestChartCarryOnlyIsNotEmpty` — weight in October, chart
    November 1990; not empty; loss 0; deficit 0.
8.  `TestChartBracketChangesMonth` — November chart, `[` → October in
    the title, one `]` → November.
8a. `TestChartEscKeepsBrowsedMonth` — month sheet November, `c`, `[`,
    Esc; sheet title is October.
8b. `TestChartCWhileEditingIsText` — month sheet editing note; `c`
    does not open the chart; input contains `c`.
9.  `TestChartAxisUsesDisplayUnit` — 80 kg stored, `display_unit=lb`;
    axis/label shows pounds, not `80.0` as kg.
10. `TestChartKeysDoNotWrite` — on chart, type `x`, Tab, Space; store
    row count unchanged.
11. `TestMonthlyBalanceLossAndDeficit` — first 80 kg, last 79 kg, 30
    days; loss 1 kg, deficit 257 kcal.
12. `TestMonthlyBalanceGainIsNegative` — inverse of 11.
13. `TestChartShowsMonthlyLossAndDeficit` — View contains the loss in
    display unit and the calorie figure.
14. `TestChartOmitsAnalysisWithoutEndpoints` — no trend on day 1;
    View has no deficit number.
15. `TestChartLossUsesDisplayUnit` — same kg drop, `lb`; loss in
    pounds, deficit still 257.
16. `TestChartCurrentMonthDividesByElapsedDays` — freeze today 10
    Nov 1990; no column after the 10th; deficit divisor 10.
17. `TestChartFlatSeriesHasScale` — one weigh-in; Y span 1.0 display
    unit; no panic.
18. `TestChartDailyMarkWinsSharedCell` — a day where weight equals
    trend; that column's plot cell is the daily mark.

Do not require a real terminal. Continue to drive `App.Update` / `View`.
