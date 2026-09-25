# Plan — Weight−trend delta column

**Spec**:
[`docs/superpowers/specs/2026-09-22-delta-column.md`](../specs/2026-09-22-delta-column.md)
**Objections**:
[`delta-column.md`](../objections/delta-column.md)
(O1–O7 accepted)
**Status**: approved

No production code until the spec’s scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/layout.go` | `wDelta`, header `delta` after `trend` on list and month. |
| `internal/tui/delta.go` | `FromKG(w) − FromKG(t)`, round 1 decimal, then sign. |
| `internal/tui/app.go` | List row: delta cell; whole-row reverse includes it. |
| `internal/tui/month.go` | Sheet row: same cell, role colour; `colCount` unchanged. |
| `internal/config/colors.go` | Known roles `delta-pos`, `delta-neg`, `delta-zero`; fail-open. |
| `internal/tui/style.go` | Palette fields; defaults yellow / green / white. |
| `internal/tui/layout_test.go` | S10: header/value column ends for `delta`. |
| `internal/tui/app_test.go` / `month_test.go` | S1–S9, S11: strip ANSI; Tab skip; custom colour. |
| `internal/config/config_test.go` | Invalid `delta-pos` dropped; valid key kept. |
| `docs/how-to/colors.md` | Three keys. |
| `docs/reference/config.md` | Role table. |
| `docs/explanation/color.md` | Delta as a data role. |
| `CHANGELOG.md` | Added. |

Do not add a SQLite column. Do not add a month `colDelta`.
Do not plot delta on charts.

## Algorithm notes

-   Need `Weight != nil` and a displayed trend (same condition as the
    trend cell). Else `""`. First weigh-in: trend equals weight →
    `0.0`.
-   Paint each `FromKG` with `%.1f` (same as the cells), parse,
    subtract, then `%.1f` again. Never `math.Round`.
-   `v == 0` → `0.0` and `delta-zero`; `v > 0` → `+%.1f` and
    `delta-pos`; `v < 0` → `%.1f` and `delta-neg`.
-   `visPad` the styled string; `lipgloss.Width`.
-   List selected row: reverse the whole line. Month: style delta
    with the role even when another cell on that day is focused.

## Test case list

1.  `TestListDeltaPositive` — 80.5 vs 80.0 kg; visible `+0.5`.
2.  `TestListDeltaNegative` — 79.5 vs 80.0 kg; visible `-0.5`.
3.  `TestListDeltaZero` — equal; visible `0.0`; no `+0.0`.
3a. `TestListDeltaRoundedZero` — 80.04 vs 80.00 kg; `0.0`; no `+`.
4.  `TestMonthDeltaBlankWithoutWeight` — carry trend, no weigh-in;
    trend shown; delta empty.
5.  `TestListDeltaFirstWeighInIsZero` — first weigh-in kg; `0.0`.
5a. `TestListDeltaHalfwayMatchesPaintedCells` — 80.05 vs 80.00 kg;
    `0.0` (`%.1f`, not `math.Round` `+0.1`).
5b. `TestListDeltaFirstWeighInLBIsDisplayedSubtraction` — 176.5 lb;
    `-0.1` (176.5−176.6).
6.  `TestListDeltaDisplayedCellsAddUp` — 80.0 kg / 79.9 kg in lb;
    weight `176.4`, trend `176.1`, delta `+0.3`.
7.  `TestMonthDeltaMatchesList` — same day `+0.5` on the sheet.
8.  `TestMonthTabSkipsDelta` — from weight, Tab lands on sleep.
9.  `TestListDeltaNOCOLORKeepsSign` — `NO_COLOR` set; `+0.5`; no
    `38;2;` / `38;5;` in that cell.
10. `TestDeltaColumnAligns` — header `delta` and value share a column
    end (ANSI stripped).
11. `TestListDeltaCustomColorPainted` — `delta-pos = magenta`; `+0.5`
    uses magenta (tests `t.Setenv("NO_COLOR", "")`).
11a. `TestListDeltaCustomNegColorPainted` / `Zero` — magenta on
    `-0.5` and `0.0`.
11b. `TestInvalidDeltaNegColorDropped` — bad `delta-neg`; default
    green on `-0.5`.
11c. `TestListDeltaSelectionIncludesReverse` — SGR 7 on the `+0.5`
    line.
11d. `TestMonthDeltaKeepsRoleWhenWeightFocused` — weight focused;
    delta still magenta.
12. `TestInvalidDeltaColorDropped` — bad `delta-pos` omitted; log
    still opens; default yellow on `+0.5`.

`t.Setenv("NO_COLOR", "1")` or `""` as needed. Do not `t.Parallel()`
tests that share process-global color if they conflict; `t.Setenv` is
isolated per test in Go 1.17+.
