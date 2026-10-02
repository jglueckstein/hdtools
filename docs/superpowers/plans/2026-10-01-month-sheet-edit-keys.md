# Plan — Spreadsheet keys while editing a month cell

**Spec**:
[`docs/superpowers/specs/2026-10-01-month-sheet-edit-keys.md`](../specs/2026-10-01-month-sheet-edit-keys.md)
**Follows**:
[`2026-09-10-month-tab-next-cell.md`](../specs/2026-09-10-month-tab-next-cell.md)
**Objections**:
[`month-sheet-edit-keys.md`](../objections/month-sheet-edit-keys.md)
(O1–O4 accepted). Code-mode:
[`month-sheet-edit-keys-code.md`](../objections/month-sheet-edit-keys-code.md)
(O1 accepted, O2 deferred, O3 accepted)
**Status**: approved

No production code until the spec’s scenarios exist as failing tests.
"The same save" is the persist step only. Code-mode O3: the Tab spec
opening and its Enter out-of-scope sentence still said Enter stays in
place. Amend those two sentences. Decision 2 and FR1 already say Tab
does not take Enter's vertical move. Code-mode O2 is deferred: do not
clear the sticky `saved` status in this slice.

## Module structure

| File | Why |
| --- | --- |
| `internal/tui/app.go` | `updateMonth` while `a.month.editing`: Enter and Down save then move down; Up save then move up; Left/Right stay on `textinput.Update`. `savedMsg` needs a vertical day delta, not only Tab’s `advance` / `nextCell`. Workout Space still `saveMonthCell` in place. |
| `internal/tui/month.go` | A `moveDay` helper (`day += delta` then existing `clamp`) so vertical post-save does not call `nextCell` and does not wrap months. |
| `internal/tui/app_test.go` | S1–S15: drive `App.Update` with `tea.KeyMsg` like existing month Tab tests (`beginEdit`, `openStore`). |
| `internal/tui/month_test.go` | Optional unit tests for `moveDay` at day 1 and last day. Keep `TestNextCell*` green. |
| `docs/reference/keys.md` | FR13. While editing, Enter and Down accept and move down, Up accepts and moves up, Left and Right move the caret. Not-editing arrows and Enter stay as today. |
| `docs/superpowers/specs/2026-09-10-month-tab-next-cell.md` | Opening and Enter out-of-scope sentence no longer say Enter stays in place (code-mode O3). |

Do not add a SQLite column. Do not change Tab’s destination or call
`nextCell` from a vertical save.
Do not change the month help line (`arrows move`, `enter form`). It
stays the not-editing chord list.
Do not change list, form, or charts.
Do not rewrite `TestMonthTabWhileEditingSavesAndAdvances` or its
neighbours.

## Algorithm notes

-   Reuse the Tab save-then-move pattern: a Cmd that calls
    `saveMonthCell`, and on `savedMsg` applies the move. A failed
    save is still `loadErrMsg` and does not move (existing Tab
    behaviour).
-   Keep `savedMsg.advance` for Tab only. Add a vertical field (day
    delta: `+1` down, `-1` up, `0` stay). On `Update(savedMsg)`:
    `cancelEdit`; if `advance` then `nextCell`; else apply the day
    delta and `clamp`. Do not set both. Do not call `nextCell` for
    vertical moves.
-   Add `saveMonthCellAndMoveDown` / `saveMonthCellAndMoveUp` (or
    one helper with a delta) analogous to
    `saveMonthCellAndAdvance`. Enter and Down both use down. Space
    on workout still uses `saveMonthCell` with no move.
-   Vertical destination is the same clamp as not-editing Up/Down:
    pin to 1..daysInMonth in the current month. Do not call
    `prevMonth` / `nextMonth`. Last day down and day 1 up stay put
    after a successful save.
-   Left and Right while editing: no new cases. They remain the
    default `a.month.input.Update(msg)` path. Do not save, and do
    not change day or column.
-   Do not change the not-editing branch: arrows move cells; Enter
    still `openForm`. After a successful Enter, Down, or Up, editing
    has ended on the destination, so the next Enter is that same
    branch (FR12): it opens the form on the focused cell and writes
    nothing. On the last day the focused cell is that same day. Do
    not swallow the second Enter. Do not leave the destination in
    edit mode.
-   Same cell save as today: `patchCell`, display unit, `Upsert`.
    Do not invent a new save path. Enter, Down, and Up capture the
    day, column, and text when the key is handled (FR14). The
    command writes that snapshot. Esc before the command runs
    cancels it: nothing stored, day unchanged. A second vertical
    key's save message does not move another day.
-   Do not clear `a.status` on a later failed edit or on arrow
    keys (code-mode O2 deferred).
-   `TestMonthDownOnLastDayStays` starts from `editing 80.0` on day
    30. It does not press Down after an Enter that has already ended
    editing.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestMonthEnterWhileEditingSavesAndMovesDown` |
| FR2 | `TestMonthDownWhileEditingSavesAndMovesDown` |
| FR3 | `TestMonthUpWhileEditingSavesAndMovesUp` |
| FR4 | `TestMonthEnterInvalidDoesNotMove`, `TestMonthDownInvalidDoesNotMove`, `TestMonthUpInvalidDoesNotMove` |
| FR5 | `TestMonthEnterOnLastDayStays`, `TestMonthDownOnLastDayStays` |
| FR6 | `TestMonthUpOnFirstDayStays` |
| FR7 | `TestMonthLeftRightWhileEditingStayInCell`, `TestMonthLeftAtStartRightAtEndStayInCell` |
| FR8 | `TestMonthTabWhileEditingSavesAndAdvances` (existing; keep green) |
| FR9 | `TestMonthEscWhileEditingCancels` |
| FR10 | `TestMonthEnterWhenNotEditingOpensForm`, `TestMonthDownUpWhenNotEditingMoveDay`, `TestMonthLeftRightWhenNotEditingMoveColumn` |
| FR11 | `TestMonthEnterPreservesColumn` |
| FR12 | `TestMonthSecondEnterAfterEnterOpensForm`, `TestMonthEnterAfterDownOpensForm`, `TestMonthEnterAfterUpOpensForm`, `TestMonthSecondEnterOnLastDayOpensForm` |
| FR13 | `docs/reference/keys.md` (no new Go test) |
| FR14 | `TestMonthEnterSnapshotIgnoresLaterCursor`, `TestMonthSecondVerticalKeyBeforeSaveDoesNotSkip`, `TestMonthEscBeforeSaveCancels` |

## Test case list

November 1990. Set the sheet with `newMonth(time.Date(1990, 11, …))`
as in `TestMonthTabWhileEditingSavesAndAdvances`. Drive keys with
`tea.KeyMsg` (`KeyEnter`, `KeyDown`, `KeyUp`, `KeyLeft`, `KeyRight`,
`KeyTab`, `KeyEsc`); `press()` does not send arrows. After a save
Cmd, `app.Update(cmd())` as in that Tab test; run
`app.Update(app.load())` when the test needs the reloaded series
(existing Tab tests assert the store and focus without that extra
load). `openStore` / `seedDays` / `beginEdit` as today. Assert
caret with `month.input.Position()` (`beginEdit` already
`CursorEnd`). Do not `t.Parallel()` tests that call `freezeToday`;
these tests should not need `freezeToday`.

Existing Tab tests to keep green (do not rewrite):
`TestMonthTabWhileEditingSavesAndAdvances`,
`TestMonthTabInvalidDoesNotAdvance`,
`TestMonthTabWhenNotEditingMoves`,
`TestMonthTabSkipsDelta`,
`TestNextCellWalksColumnsThenNextDay`,
`TestNextCellStaysOnLastCellOfMonth`.

1.  `TestMonthEnterWhileEditingSavesAndMovesDown` — S1: day 4
    weight `80.0`; Enter; stored 80 kg; not editing; weight day 5.
2.  `TestMonthDownWhileEditingSavesAndMovesDown` — S2: same as S1
    with Down.
3.  `TestMonthUpWhileEditingSavesAndMovesUp` — S3: day 4 weight
    `80.0`; Up; stored 80 kg; not editing; weight day 3.
4.  `TestMonthEnterInvalidDoesNotMove` — S4: `nope`; Enter; still
    editing weight; no row stored.
5.  `TestMonthDownInvalidDoesNotMove` — S5: `nope`; Down; still
    editing weight; no row stored.
6.  `TestMonthUpInvalidDoesNotMove` — S5b: `nope`; Up; still
    editing weight; no row stored.
7.  `TestMonthEnterOnLastDayStays` — S6: day 30 weight `80.0`;
    Enter; stored; day 30 weight; not editing; still November.
8.  `TestMonthDownOnLastDayStays` — S6: day 30 weight `80.0`; Down;
    stored; day 30 weight; not editing; still November.
9.  `TestMonthUpOnFirstDayStays` — S7: day 1 weight `80.0`; Up;
    stored; day 1 weight; not editing; still November.
10. `TestMonthLeftRightWhileEditingStayInCell` — S8: `80.0` caret
    at end; Left; `Position()` decreased; still editing day 4
    weight; no row. Caret at start; Right; `Position()` increased;
    still editing day 4 weight; no row.
11. `TestMonthLeftAtStartRightAtEndStayInCell` — S9: caret at
    start; Left; `Position()` still 0; still editing day 4 weight;
    no row. Caret at end; Right; `Position()` still at end; still
    editing day 4 weight; no row.
12. `TestMonthTabWhileEditingSavesAndAdvances` — S10: existing
    test; keep green.
13. `TestMonthEnterWhenNotEditingOpensForm` — S11: not editing;
    Enter; `screenForm`; no log row from that Enter (`openStore`
    required because `openForm` reads the store).
14. `TestMonthDownUpWhenNotEditingMoveDay` — S12: not editing day
    4; Down → day 5 same column, nothing written; Up from day 4 →
    day 3 same column, nothing written.
15. `TestMonthEnterPreservesColumn` — S13: sleep `8` on day 4;
    Enter; sleep stored; not editing; sleep day 5.
16. `TestMonthEscWhileEditingCancels` — S14: `80.0`; Esc; not
    editing; still weight day 4; no row.
17. `TestMonthLeftRightWhenNotEditingMoveColumn` — S15: not
    editing; Right from weight → sleep day 4; Left from sleep →
    weight day 4; nothing written.
18. `TestMonthSecondEnterAfterEnterOpensForm` — S16: day 4 weight
    `80.0`; Enter; stored 80 kg; not editing; weight day 5; Enter
    again; day form for 5 November; that second Enter writes
    nothing.
19. `TestMonthEnterAfterDownOpensForm` — S16: day 4 weight
    `80.0`; Down; not editing; weight day 5; Enter; day form for
    5 November; that Enter writes nothing.
20. `TestMonthEnterAfterUpOpensForm` — S16: day 4 weight `80.0`;
    Up; not editing; weight day 3; Enter; day form for 3 November;
    that Enter writes nothing.
21. `TestMonthSecondEnterOnLastDayOpensForm` — S16: day 30 weight
    `80.0`; Enter; not editing; still weight day 30; Enter again;
    day form for 30 November; that second Enter writes nothing.
22. `TestMonthEnterSnapshotIgnoresLaterCursor` — S17/FR14: day 4
    weight `80.0`; Enter returns a command; mutate the live cursor
    and buffer before the command runs; stored 80 kg on 4 November;
    not day 10; focus weight day 5.
23. `TestMonthSecondVerticalKeyBeforeSaveDoesNotSkip` — S17: Enter
    then Down before either save is applied; stored 80 kg on day 4;
    no row on day 5; not editing; weight day 5, not day 6.
24. `TestMonthEscBeforeSaveCancels` — S18: Enter returns a command;
    Esc before it runs; not editing; still weight day 4; no row.

Optional in `month_test.go`: `TestMoveDayClampsAtMonthEnds` — day
30 `+1` stays 30; day 1 `-1` stays 1; month unchanged.
