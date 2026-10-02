# Tab accepts a month cell and moves to the next

**Date**: 2026-09-10
**Status**: accepted
**Issue**: [#28](https://github.com/jglueckstein/hdtools/issues/28)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)

The monthly sheet is for entering a month in one sitting. Enter while
editing persists the cell (display unit, kilograms in the log). Tab
repeats that persist step, then moves to the next cell so a row can
be filled. A failed save does not move. Enter then moves vertically;
that move belongs to
[2026-10-01-month-sheet-edit-keys.md](2026-10-01-month-sheet-edit-keys.md).

## User stories

### US1 — Tab through a row

As a person filling a monthly sheet, I want Tab to accept the value I
just typed and move to the next cell so I can enter weight, sleep, and
steps in sequence.

### US2 — Tab without editing

As a person moving around the sheet, I want Tab to advance one cell when
I am not editing, like Right arrow.

## Decisions

1.  **Month sheet only.** Form Tab still moves between form fields
    without saving the whole form.
2.  **Save then advance.** While editing, Tab persists the cell the
    same way Enter's save does (display unit, kilograms in the log),
    then moves to the next cell. A failed save does not advance. Tab
    does not take Enter's vertical move.
3.  **Next cell is the next column**, wrapping to the next day's first
    column (weight). The last cell of the month does not wrap.
4.  **Workout is a cell.** Tab lands on it; Space still toggles. Tab
    does not toggle.
5.  **No Shift+Tab** in this slice.

## Acceptance scenarios

### S1 — Tab while editing saves and advances

**Given** the month sheet focused on 4 November, weight column, editing
`80.0`
**When** Tab is pressed
**Then** the day's weight is stored as 80 kg
**And** editing has ended
**And** the focus is the sleep cell on the same day

### S2 — Invalid Tab does not advance

**Given** the month sheet editing weight with `nope`
**When** Tab is pressed
**Then** the cell is still being edited
**And** the column is still weight
**And** no weight is stored

### S3 — Tab when not editing moves

**Given** the month sheet focused on weight, not editing
**When** Tab is pressed
**Then** the focus is the sleep cell
**And** nothing is written

### S4 — Tab wraps to the next day

**Given** the month sheet focused on the note cell of day 4, not editing
**When** Tab is pressed
**Then** the focus is the weight cell of day 5

### S5 — Last cell stays

**Given** the month sheet focused on the note cell of the last day of
the month
**When** Tab is pressed
**Then** the focus is still that cell

## Functional requirements

-   **FR1.** While editing a month cell, Tab persists that cell the
    same way Enter's save does (display unit, kilograms in the log;
    a failed save does not move), then moves to the next cell. Tab
    does not follow Enter's vertical destination.
-   **FR2.** If that save fails, focus and editing stay put.
-   **FR3.** When not editing, Tab moves to the next cell without
    writing.
-   **FR4.** Next cell is the next column in
    weight, sleep, steps, workout, note; after note, the next day's
    weight. The last cell of the month does not wrap.
-   **FR5.** Form Tab is unchanged.

## Out of scope

-   Shift+Tab
-   Tab on the list screen
-   Enter's vertical move (owned by
    [2026-10-01-month-sheet-edit-keys.md](2026-10-01-month-sheet-edit-keys.md)).
    This spec covers Enter's persist step only.
