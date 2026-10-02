# Spreadsheet keys while editing a month cell

**Date**: 2026-10-01
**Status**: approved
**Objections**:
[`month-sheet-edit-keys.md`](../objections/month-sheet-edit-keys.md)
(O1–O4 accepted). Code-mode:
[`month-sheet-edit-keys-code.md`](../objections/month-sheet-edit-keys-code.md)
(O1 accepted, O2 deferred, O3 accepted).
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Follows**:
[`2026-09-10-month-tab-next-cell.md`](2026-09-10-month-tab-next-cell.md)
**Amends**: that spec's opening paragraph, Decision 2, FR1, and the
out-of-scope line on Enter. "The same save" is the persist step only
(display unit, kilograms in the log, failed save does not move).
Enter's vertical move is this spec. Tab's destination stays FR8 and
S10. The vertical save must not call Tab's next-cell move.

The monthly sheet is for entering a month in one sitting. Tab while
editing already saves the cell and moves to the next. Enter currently
saves in place. This slice makes Enter and Down accept the edit and
move down the column, Up accept and move up, and Left and Right move
the caret in the text, as `idea.md` asks.

Delta is not a focusable column (see
[2026-09-22-delta-column.md](2026-09-22-delta-column.md)); Tab still
skips it. This is not the Goto Today slice.

Scenarios use November 1990 (30 days). Day 4 is a typical mid-month
cell; day 30 is last; day 1 is first.

## User stories

### US1 — Fill a column downward

As a person entering a month of weights (or sleep, or steps), I want
Enter and Down to accept the cell I just typed and move to the same
column on the next day so I can fill a column without Tabbing across
the row.

### US2 — Step back up the column

As a person who needs the previous day’s cell in the same column, I
want Up to accept the current edit and move up one day.

### US3 — Fix text without leaving the cell

As a person correcting a typo while editing, I want Left and Right to
move the caret in the value, not save or leave the cell.

## Decisions

1.  **Enter while editing changes.** It saves, then moves down in the
    current column (the same destination as Down). It no longer saves
    in place.
2.  **Down while editing** is the same accept-and-move-down as Enter.
3.  **Up while editing** saves, then moves up in the current column.
4.  **Failed save.** Invalid text (the same failure as Tab) stays
    editing, on the same cell and column, with nothing stored. No
    move.
5.  **After a successful save**, editing has ended. The destination
    cell is focused, not in edit mode. The column is unchanged.
6.  **Edges do not wrap months.** On the last day of the month, a
    successful Enter or Down saves and stays on that day (same
    column, not editing). On day 1, a successful Up saves and stays
    on day 1. Same spirit as Tab on the last cell of the month (Tab
    spec FR4).
7.  **Left and Right while editing** move the caret in the text.
    They never save, never leave the cell, and never change day or
    column. At the start of the text, Left stays in the cell with
    the caret at the start. At the end, Right stays in the cell with
    the caret at the end.
8.  **Tab while editing is unchanged** (persist, then next cell:
    weight → sleep → steps → workout → note → next day’s weight;
    last cell of the month stays). The vertical save must not call
    Tab's next-cell move.
9.  **Esc still cancels** without save.
10. **When not editing, arrows and Enter are unchanged** (arrows
    move cells; Enter still opens the form).
11. **Month sheet only.** List, form, charts, CLI, PDF, Goto Today,
    extra daily fields, and delta (not a focusable column) are out
    of scope.
12. **Workout** is not text-edited; do not change Space or the
    workout toggle.
13. **Same save as today.** Accept stores the cell the same way Tab
    already does (display unit, kilograms in the log). Do not invent
    a different save.
14. **`docs/reference/keys.md` is in scope.** While editing, Enter
    and Down accept and move down, Up accepts and moves up, and
    Left and Right move the caret. Not-editing arrows and Enter
    stay as they are.
15. **The month help line stays the not-editing chord list**
    (`arrows move`, `enter form`). It is not a second contract for
    the editing keys.
16. **The accept is the cell at the key.** Enter, Down, and Up
    capture the day, column, and text when the key is handled. The
    save writes that cell, not wherever the cursor is when the
    command later runs. A second Enter, Down, or Up before that
    save is applied does not move another day. Esc before the
    command runs cancels it: nothing is stored and the day does
    not change. The sticky `saved` status line is unchanged (code
    objection O2 deferred).

## Acceptance scenarios

### S1 — Enter while editing saves and moves down

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Enter is pressed
**Then** the 4 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is the weight cell on 5 November
**And** that cell is not being edited

### S2 — Down while editing is the same as Enter

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Down is pressed
**Then** the 4 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is the weight cell on 5 November
**And** that cell is not being edited

### S3 — Up while editing saves and moves up

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Up is pressed
**Then** the 4 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is the weight cell on 3 November
**And** that cell is not being edited

### S4 — Invalid Enter does not move

**Given** the month sheet editing weight with `nope`
**When** Enter is pressed
**Then** the cell is still being edited
**And** the column is still weight
**And** the day has not changed
**And** no weight is stored

### S5 — Invalid Down does not move

**Given** the month sheet editing weight with `nope`
**When** Down is pressed
**Then** the cell is still being edited
**And** the column is still weight
**And** the day has not changed
**And** no weight is stored

### S5b — Invalid Up does not move

**Given** the month sheet editing weight with `nope`
**When** Up is pressed
**Then** the cell is still being edited
**And** the column is still weight
**And** the day has not changed
**And** no weight is stored

### S6 — Last day of the month does not wrap

**Given** the month sheet for November 1990, focused on 30 November,
weight column, editing `80.0`
**When** Enter is pressed
**Then** the 30 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is still the weight cell on 30 November
**And** the sheet is still November 1990

**Given** the month sheet for November 1990, focused on 30 November,
weight column, editing `80.0`
**When** Down is pressed
**Then** the 30 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is still the weight cell on 30 November
**And** the sheet is still November 1990

### S7 — Up on day 1 does not wrap

**Given** the month sheet for November 1990, focused on 1 November,
weight column, editing `80.0`
**When** Up is pressed
**Then** the 1 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is still the weight cell on 1 November
**And** the sheet is still November 1990

### S8 — Left and Right while editing stay in the cell

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`, with the caret at the end of the text
**When** Left is pressed
**Then** the caret has moved toward the start of the text
**And** editing continues
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`, with the caret at the start of the text
**When** Right is pressed
**Then** the caret has moved toward the end of the text
**And** editing continues
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

### S9 — Left at the start and Right at the end stay in the cell

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`, with the caret at the start of the text
**When** Left is pressed
**Then** the caret is still at the start of the text
**And** editing continues
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`, with the caret at the end of the text
**When** Right is pressed
**Then** the caret is still at the end of the text
**And** editing continues
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

### S10 — Tab while editing still advances to the next column

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`
**When** Tab is pressed
**Then** the day’s weight is stored as 80 kg
**And** editing has ended
**And** the focus is the sleep cell on the same day

### S11 — Enter when not editing still opens the form

**Given** the month sheet focused on 4 November, weight column, not
editing
**When** Enter is pressed
**Then** the day form is open
**And** that Enter did not write a log row

### S12 — Down and Up when not editing still move the day

**Given** the month sheet focused on 4 November, weight column, not
editing
**When** Down is pressed
**Then** the focus is 5 November, same column
**And** nothing is written

**Given** the month sheet focused on 4 November, weight column, not
editing
**When** Up is pressed
**Then** the focus is 3 November, same column
**And** nothing is written

### S13 — Column is preserved on a vertical move

**Given** the month sheet focused on 4 November, sleep column,
editing `8`
**When** Enter is pressed
**Then** the 4 November sleep is stored as 8 hours
**And** editing has ended
**And** the focus is the sleep cell on 5 November
**And** that cell is not being edited

### S14 — Esc still cancels without save

**Given** the month sheet focused on 4 November, weight column,
editing `80.0`
**When** Esc is pressed
**Then** editing has ended
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

### S15 — Left and Right when not editing still move columns

**Given** the month sheet focused on 4 November, weight column, not
editing
**When** Right is pressed
**Then** the focus is the sleep cell on 4 November
**And** nothing is written

**Given** the month sheet focused on 4 November, sleep column, not
editing
**When** Left is pressed
**Then** the focus is the weight cell on 4 November
**And** nothing is written

### S16 — The next Enter opens the form on the focused cell

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Enter is pressed
**Then** the 4 November weight is stored as 80 kg
**And** editing has ended
**And** the focus is the weight cell on 5 November
**And** that cell is not being edited
**When** Enter is pressed again
**Then** the day form is open for 5 November
**And** that second Enter did not write a log row

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Down is pressed
**Then** the focus is the weight cell on 5 November
**And** that cell is not being edited
**When** Enter is pressed
**Then** the day form is open for 5 November
**And** that Enter did not write a log row

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Up is pressed
**Then** the focus is the weight cell on 3 November
**And** that cell is not being edited
**When** Enter is pressed
**Then** the day form is open for 3 November
**And** that Enter did not write a log row

**Given** the month sheet for November 1990, focused on 30 November,
weight column, editing `80.0`
**When** Enter is pressed
**Then** editing has ended
**And** the focus is still the weight cell on 30 November
**And** that cell is not being edited
**When** Enter is pressed again
**Then** the day form is open for 30 November
**And** that second Enter did not write a log row

### S17 — A second vertical key does not skip a day

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Enter is pressed and, before that save is applied, Down is
pressed
**Then** the 4 November weight is stored as 80 kg
**And** 5 November has no log row
**And** editing has ended
**And** the focus is the weight cell on 5 November
**And** the focus is not 6 November

### S18 — Esc before the save runs cancels it

**Given** the month sheet for November 1990, focused on 4 November,
weight column, editing `80.0`
**When** Enter is pressed and, before that save command runs, Esc is
pressed
**Then** editing has ended
**And** the focus is still the weight cell on 4 November
**And** nothing is stored

## Functional requirements

-   **FR1.** While editing a month cell, Enter saves that cell the
    same way Tab already does, then focuses the same column on the
    next day. Editing has ended. The destination is not being
    edited.
-   **FR2.** While editing a month cell, Down does the same as Enter
    (FR1).
-   **FR3.** While editing a month cell, Up saves that cell the same
    way Tab already does, then focuses the same column on the
    previous day. Editing has ended. The destination is not being
    edited.
-   **FR4.** If that save fails, focus, day, column, and editing stay
    put, and nothing is stored. This holds for Enter, Down, and Up.
-   **FR5.** On the last day of the month, a successful Enter or Down
    saves and leaves focus on that day and column, not editing. The
    sheet does not change month.
-   **FR6.** On day 1 of the month, a successful Up saves and leaves
    focus on day 1 and the same column, not editing. The sheet does
    not change month.
-   **FR7.** While editing, Left and Right move the caret in the
    text. They do not save, do not end editing, and do not change
    day or column. Left at the start of the text stays at the start.
    Right at the end stays at the end.
-   **FR8.** While editing, Tab still persists the cell and moves
    to the next cell as in
    [2026-09-10-month-tab-next-cell.md](2026-09-10-month-tab-next-cell.md)
    (weight → sleep → steps → workout → note → next day’s weight;
    last cell of the month stays). Enter, Down, and Up must not
    call that next-cell move.
-   **FR9.** While editing, Esc cancels the edit without saving.
-   **FR10.** When not editing, Left and Right still change column,
    Up and Down still change day, and none of those keys write.
    Enter still opens the day form.
-   **FR11.** A vertical move after a successful save keeps the
    column (sleep stays sleep).
-   **FR12.** After a successful Enter, Down, or Up, the next Enter
    is FR10 on the cell now focused. It opens the day form there
    and writes nothing. On the last day that cell is the same day.
    That Enter is not swallowed. The destination of the first key
    is not left in edit mode.
-   **FR13.** `docs/reference/keys.md` records the editing chords
    (Enter and Down accept and move down, Up accepts and moves up,
    Left and Right move the caret) and the unchanged not-editing
    chords (arrows move cells; Enter opens the form).
-   **FR14.** Enter, Down, and Up write the day, column, and text
    captured when the key was handled. A second of those keys
    before the save is applied does not move another day. Esc
    before that command runs stores nothing and does not change
    the day.

## Out of scope

-   List, form, charts, CLI, and PDF
-   Goto Today
-   Extra daily fields
-   Delta as a focusable column (Tab still skips it)
-   Shift+Tab
-   Changing Tab's destination
-   Changing the month help line (it stays the not-editing chord
    list: `arrows move`, `enter form`)
-   Space or the workout toggle
-   Wrapping to another month
