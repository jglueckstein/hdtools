# Plan — Monthly chart title box and float/sinker stems

**Spec**:
[`docs/superpowers/specs/2026-09-11-monthly-chart-floats-sinkers.md`](../specs/2026-09-11-monthly-chart-floats-sinkers.md)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/chart.go` | Title box (month year, Excel colours). Stems (`\|`) from daily mark to trend in that column. Paint order: path, stems, daily mark. |
| `internal/tui/chart_test.go` | Title line is `November 1990` not `hdtools — …`; box SGR; stems; `NO_COLOR`; coincident day has no stem. |
| `internal/tui/style.go` | Optional local styles for box and stem. Not new `[colors]` keys. |
| `docs/how-to/monthly-chart.md` | Mention the month-year box and stems. |
| `CHANGELOG.md` | Added. |

`internal/dailylog` does not change. PDF is not implemented; the spec only records that PDF must reuse this look.

## Algorithm notes

-   **Title box.** One centered line above the plot: `time.Month.String()`
    plus year (`November 1990`). Do not prefix `hdtools —`. Lipgloss:
    foreground yellow, background blue, border red (16-color names).
    Center over the plot including gutter so the box sits in the
    chart, not the left margin. Display unit stays off the box (muted
    line or Y-axis, as now). Under `NO_COLOR`, skip chromatic SGR;
    keep the centered month-year text. Empty months still show the
    box.
-   **Stems.** After the trend path, for each day with a daily weight
    and a trend: map both to rows. If the rows differ, put `\|` in
    the cells strictly between them (not on the mark, not on the
    path). Colour those glyphs green (`lipgloss.Color("2")`). If the
    rows are equal, draw nothing. Then paint `o` so the daily mark
    still wins the shared cell.
-   **Z-order.** Path, then stems, then daily marks. A stem glyph is
    not `-` `/` `\` or `o`.
-   **No TTY.** Reuse SGR helpers. Yellow `33`/`93`, blue background
    `44`, red border `31`, stem green `32`.
-   **PDF.** No code in this slice. When a PDF sitting starts, this
    plan is the look to copy (title box + stems), not the current
    `hdtools —` header without stems.

## Test case list

1.  `TestChartTitleIsMonthYear` — twoDayApp, `c`; visible title line
    is `November 1990` (no `hdtools —` on that line).
2.  `TestChartTitleBoxColours` — chroma on; that line has yellow
    foreground and blue background (and red if the border is
    visible in SGR).
3.  `TestChartTitleSurvivesNoColor` — `NO_COLOR=1`; month-year text
    still present and centered; no chromatic SGR on that line.
4.  `TestChartStemJoinsMarkToTrend` — a day whose weight and trend
    map to different rows; that day's column contains `\|` between
    them; `o` still present.
5.  `TestChartNoStemWhenMarkOnTrend` — coincident day (first weigh-in);
    that column has `o` and no `\|`.
6.  `TestChartStemGreen` — chroma on; stem cells use green SGR.
7.  Existing chart tests stay green (open/Esc, `[` `]`, analysis,
    empty, clip at today).

Do not require a real terminal. Continue to drive `App.Update` /
`View`.
