# Plan — Goto Today

**Spec**:
[`docs/superpowers/specs/2026-09-19-goto-today.md`](../specs/2026-09-19-goto-today.md)
**Follows**:
[`2026-09-17-closest-to-today-startup.md`](../specs/2026-09-17-closest-to-today-startup.md)
**Objections**:
[`goto-today.md`](../objections/goto-today.md)
(O1, O2, O4–O6 accepted; O3 rejected)
**Status**: approved

No production code until the spec’s scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/app.go` | List `t` → `closestLogIndex(localToday())`. Month `t` (not editing) → `newMonth(localToday())` (weight column); do not change list cursor. Help lines include `t`. |
| `internal/tui/select.go` | Reuse `closestLogIndex`. No new distance rule. Do not set list cursor from month `t`. |
| `internal/tui/app_test.go` | S1–S3, S7, S10–S12, S14 list help: freeze today; `press("t")`; assert cursor day and row count. |
| `internal/tui/month_test.go` | S4–S6, S9, S13, S14 month help: sheet after `t`; day 10; weight column; empty series; Esc list unchanged; editing types `t`. |
| `internal/tui/chart_test.go` | S8: `t` on monthly and long-term chart does not write logs or leave the chart. |
| `docs/reference/keys.md` | `t` on list and month (not editing). |

Do not `Upsert` in the `t` path. Do not bind `t` as Goto on form or charts.
Do not copy the list empty no-op onto the month sheet.

## Algorithm notes

-   List: if `len(logs)==0`, return. Else
    `cursor = closestLogIndex(logs, localToday())`.
-   Month (not editing): `a.month = newMonth(localToday())`. That
    resets year, month, day, and column to weight. Do not assign
    `cursor`. Works when `logs` is empty.
-   Editing month cell / form: existing input path; `t` is a rune.

## Test case list

Freeze today at 10 November 1990.

1.  `TestGotoTodayListJumpsToClosest` — June + 10 Nov, cursor on June;
    `t`; cursor is 10 Nov; row count unchanged.
2.  `TestGotoTodayListDoesNotCreateToday` — 1 and 8 Nov; `t`; cursor
    is the 8th; no 10 Nov row.
3.  `TestGotoTodayListNearestFuture` — 12 and 20 Nov, cursor on the
    20th; `t`; cursor is the 12th.
4.  `TestGotoTodayListTiePrefersPast` — 9 and 11 Nov, cursor on the
    11th; `t`; cursor is the 9th.
5.  `TestGotoTodayListNearerFutureBeatsLastPast` — 1 and 11 Nov,
    cursor on the 1st; `t`; cursor is the 11th.
6.  `TestGotoTodayListEmptyDoesNothing` — no logs; `t`; still empty;
    help mentions `t`.
6a. `TestGotoTodayListDoesNotChangeMonthModel` — month `[` to October,
    Esc, list `t`; `a.month` still October.
7.  `TestGotoTodayMonthJumpsToToday` — June sheet, non-weight column;
    `t`; November, day 10, weight column; row count unchanged.
8.  `TestGotoTodayMonthDoesNotCreateToday` — Nov sheet on the 8th,
    note column; `t`; day 10, weight; no 10 Nov row.
9.  `TestGotoTodayMonthEmptyStillGoesToToday` — no logs; month sheet;
    `[` to October; `t`; November, day 10, weight; still no rows.
10. `TestGotoTodayMonthEditingIsText` — editing a cell; `t` in input;
    day unchanged.
11. `TestGotoTodayFormIsText` — note field; `t` in input; after Esc,
    list cursor unchanged.
12. `TestGotoTodayIgnoredOnChart` — monthly chart `[` to October;
    `t`; still October and still the chart. Long-term: `]` then `t`;
    kind unchanged.
13. `TestGotoTodayMonthDoesNotMoveList` — list on June, month June;
    `t` then Esc; list still on 1 June (today has a log).
14. `TestGotoTodayHelpMentionsT` — list `View()` help contains `t`;
    month `View()` help contains `t`.

Do not `t.Parallel()` tests that call `freezeToday`.
