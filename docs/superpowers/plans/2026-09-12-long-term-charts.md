# Plan — On-screen long-term weight charts

**Spec**:
[`docs/superpowers/specs/2026-09-12-long-term-charts.md`](../specs/2026-09-12-long-term-charts.md)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md),
[`2026-09-11-monthly-chart-floats-sinkers.md`](2026-09-11-monthly-chart-floats-sinkers.md)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/longchart.go` | New screen: four kinds, two line series when *D* ≤ *W*, trend-only buckets when *D* > *W*, Loss / Daily deficit. Reuses `monthYearBox` (generalise it to take the title string) and `MonthlyBalance`. Does not query SQLite. Help: ``[ ] kind``. |
| `internal/tui/longchart_test.go` | Open/Esc from list, sheet, monthly chart; `l` while editing; cycle kinds; no `o`/`\|`; colour and bold; `NO_COLOR`; empty; sheet month unchanged; display unit; no writes; Loss arithmetic; title kind+span; complete starts at first log; annual bucketed to 72 cols; quarterly clips at today. |
| `internal/tui/app.go` | `screenLong`, `afterLong`, `l` from list; help line. |
| `internal/tui/month.go` | `l` when not editing; help line. |
| `internal/tui/chart.go` | `l` from the monthly chart; `monthYearBox` accepts a string so both screens share Excel chrome. `[` / `]` on the monthly chart stay month shifts. |
| `docs/how-to/long-term-chart.md` | How to open and cycle the four kinds. |
| `docs/reference/` (keys) | `l` line. |
| `CHANGELOG.md` | Added. |

`internal/dailylog` does not change (`MonthlyBalance` already takes a day count). PDF is not implemented.

## Algorithm notes

-   **Kind.** `quarterly`, `semiannual`, `annual`, `complete`. Default on `l`: quarterly. `[` / `]` wrap in that order. Help on this screen: ``[ ] kind``. End month is the latest loaded log’s month, not the sheet’s month and not calendar-today if that is after the latest log.
-   **Window.** Quarterly: day 1 of endMonth−2 through last plotted day. Semiannual: −5 months. Annual: −11 months. Complete: date of the first loaded log through last plotted day. Last plotted day: today if the end month contains today, else last calendar day of that past month.
-   **Columns.** Plot width *W* = 72 until `tea.WindowSizeMsg` is stored (tests never send a resize). *D* = days in the span. If *D* ≤ *W*, one column per day. Else *W* buckets: day *i* covers `[floor(i·D/W), floor((i+1)·D/W))`; last trend in the bucket (carry counts).
-   **Paint.** If *D* ≤ *W*: daily path first (not bold, `weight` role), then trend path (bold, `trend` role) so trend wins a shared cell. If *D* > *W*: trend only (O8). Glyphs `-` `/` `\`. No `o`, no `\|`.
-   **Title.** `monthYearBox` with text like `Quarterly  September 1990–November 1990`. Kind word + two English month-years. Same yellow / blue / red chrome. Unit stays on the muted line.
-   **Analysis.** `MonthlyBalance(firstTrend, lastTrend, D)` when both endpoints have trend. Labels `Loss` and `Daily deficit`.
-   **Esc.** `afterLong` is the screen that opened `l`. Cycling kinds must not assign the month sheet’s year/month.
-   **No TTY.** Reuse SGR helpers. Trend bold is `1`; daily not bold.

## Test case list

Freeze today at 10 November 1990 when the span must be known. Build logs with `t.TempDir()` stores as now.

1.  `TestLongChartOpensFromListAndEscapes` — `l`; quarterly; Esc to list.
2.  `TestLongChartOpensFromMonthAndEscapes` — `m`, `l`; Esc to sheet.
3.  `TestLongChartLWhileEditingIsText` — note field; `l` does not open.
4.  `TestLongChartOpensFromMonthlyChartAndEscapes` — `c`, `l`; Esc to monthly chart.
5.  `TestLongChartCyclesKinds` — `]` ×3: semiannual, annual, complete; `]` wrap to quarterly; `[` to complete.
6.  `TestLongChartTwoLinesNoMarksOrStems` — quarterly (*D* ≤ *W*): both paths; no `o`; no `\|`.
7.  `TestLongChartWeightAndTrendColors` — quarterly; `weight=green`, `trend=yellow`; daily green not bold; trend yellow bold.
8.  `TestLongChartNoColorKeepsTwoLines` — quarterly; `NO_COLOR=1`; no chromatic SGR; both paths; trend bold.
9.  `TestLongChartEmpty` — no logs; empty copy; no Loss numbers.
9a. `TestLongChartQuarterlyEmptyDespiteLogs` — logs only in June 1990, today 10 Nov 1990; quarterly empty copy.
10. `TestLongChartBracketsDoNotChangeSheetMonth` — November sheet, `l`, `]`, Esc; sheet still November.
11. `TestLongChartAxisUsesDisplayUnit` — `lb`.
12. `TestLongChartKeysDoNotWrite`.
13. `TestLongChartLossAndDeficit` — 80 kg → 79 kg over 90 days; Loss 1.0 kg; deficit 86 cal.
14. `TestLongChartTitleKindAndSpan` — annual; title has `Annual`, `December 1989`, `November 1990`; no `hdtools —`.
15. `TestLongChartCompleteStartsAtFirstLog` — first log 15 April 1989; span starts that day; complete title has `April 1989`.
16. `TestLongChartAnnualIsBucketed` — annual; plot column count is 72, not ~365; no daily path.
17. `TestLongChartQuarterlyClipsAtToday` — today 10 Nov 1990; quarterly starts 1 September 1990, ends 10 November 1990.
18. `TestLongChartPastDatabaseUsesLatestLog` — today 12 Sep 2026, latest log 10 Nov 1990; quarterly title is Sep–Nov 1990, not empty.

Existing monthly chart tests stay green (`c`, `[` `]` still change month on that screen).

Do not require a real terminal. Drive `App.Update` / `View`.
