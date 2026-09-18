# Plan — Closest-to-today startup selection

**Spec**:
[`docs/superpowers/specs/2026-09-17-closest-to-today-startup.md`](../specs/2026-09-17-closest-to-today-startup.md)
**Objections**:
[`closest-to-today-startup.md`](../objections/closest-to-today-startup.md)
(O1 rejected; O2–O7 accepted)
**Status**: approved

No production code until the spec’s scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/app.go` | On `loadedMsg`, first load → closest-to-today; after save → written day; other reload → keep calendar day or closest. |
| `internal/tui/list.go` or `app.go` | Unexported helper: index of the log whose date is closest to today. Calendar days only; ties prefer on-or-before today. |
| `internal/tui/app_test.go` | S1–S5, S7–S10: freeze today; drive `load` / `Update(loadedMsg)`; assert `cursor` date. |
| `internal/tui/chart_test.go` | S6: June + 10 Nov fixture; `c` title is November 1990, not June 1990. |

Do not change `openChart`. Do not import SQLite in new files.
Do not create a today row on startup. Do not bind Goto Today.

## Algorithm notes

-   Compare `log.Day` and `localToday()` as dates (year/month/day),
    not timestamps.
-   Distance is `|day.Sub(today)|` in whole days.
-   Scan once: track best index. Replace when distance is smaller, or
    equal and the candidate is on-or-before today while the current
    best is after, or equal and both after today and the candidate is
    earlier.
-   First load: no previous day and no just-saved day → FR1. Closest
    may be index 0.
-   After save: `savedMsg` carries the written calendar day; on the
    following `loadedMsg`, select that day if present, else FR1.
-   Other reload: remember `logs[cursor].Day` if the index is valid,
    then find that day in the new slice; if missing, FR1.
-   Empty slice: leave `cursor` at 0; the view does not use it.

## Test case list

Freeze today at 10 November 1990.

1.  `TestLoadSelectsToday` — 1 / 10 / 20 Nov; cursor day is the 10th.
2.  `TestLoadSelectsNearestPast` — 1 and 8 Nov; cursor day is the 8th.
3.  `TestLoadSelectsNearestFuture` — 12 and 20 Nov; cursor day is the
    12th (index 0).
4.  `TestLoadTiePrefersPast` — 9 and 11 Nov; cursor day is the 9th.
5.  `TestLoadEmptyHasNoSelection` — no logs; empty copy; cursor unused.
6.  `TestLoadSelectsAcrossMonths` — 1 June 1990 and 10 Nov; cursor day
    is 10 Nov, not June.
7.  `TestLoadNearerFutureBeatsLastPast` — 1 Nov and 11 Nov; cursor day
    is the 11th.
8.  `TestChartFollowsSelectedRow` — June + 10 Nov load, `c`; title
    `November 1990`, not `June 1990`.
9.  `TestSaveKeepsNonClosestDay` — selected 1 Nov; save that form;
    cursor day still the 1st, not the 10th.
10. `TestSaveFromNewDaySelectsWrittenDay` — selected 8 Nov; save a
    new 10 Nov form; cursor day is the 10th.
11. `TestLoadOutOfRangeSelectsClosest` — cursor past `len(logs)`;
    load a shorter series (not a save); cursor is closest to today
    (index 0 is allowed if that row is closest).
12. `TestMonthCellSaveKeepsNonClosestDay` — selected 1 Nov; month-cell
    save inserts 1 June; cursor day still the 1st, not the 10th.
13. `TestLoadSelectsLocalCivilDate` — freeze 9 Nov evening in
    America/Los_Angeles (10 Nov UTC); cursor day is the 9th.
14. `TestFailedLoadClearsSelectDay` — form save, load error, move to
    1 Nov, month-cell reload; cursor day is the 1st, not the 10th.

Do not `t.Parallel()` tests that call `freezeToday`.
Do not implement Goto Today.
