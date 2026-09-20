# Goto Today

**Date**: 2026-09-19
**Status**: approved
**Backlog**: [`idea.md`](../../../idea.md) (Daily Log, Monthly Log)
**Follows**:
[`2026-09-17-closest-to-today-startup.md`](2026-09-17-closest-to-today-startup.md)
**Objections**:
[`goto-today.md`](../objections/goto-today.md)
(O1, O2, O4–O6 accepted; O3 rejected)

Startup already lands the daily list on the closest existing row to
today. After the user arrows away, there is no key to return. `idea.md`
asks for Goto Today on the **list** and the **month sheet**, without
creating a log for today if none exists.

## User stories

### US1 — Return in the list

As a person who has scrolled the daily list, I want a key that selects
the log closest to today so I do not arrow back through history.

### US2 — Return in the month sheet

As a person filling a month sheet that is not this month (or not
today's day), I want the same key to show today's calendar month and
focus today's day.

### US3 — Do not invent today

As a person who has not logged today, I do not want Goto Today to
insert a row. `n` still creates today.

## Decisions

1.  **Key is `t`.** Bound on the daily list and on the month sheet
    when not editing. Not Goto Today on the form, monthly chart, or
    long-term chart. On the form, `t` is text in the focused field.
2.  **List uses the startup closest rule.** Same calendar-day distance
    as
    [2026-09-17-closest-to-today-startup.md](2026-09-17-closest-to-today-startup.md)
    (today wins; ties prefer on-or-before today; nearer future beats
    a farther past). Reuse `closestLogIndex`. Empty list: no-op, no
    row created.
3.  **Month sheet uses civil today, not closest log.** `t` shows
    `localToday()`'s year and month and focuses that calendar day,
    **weight column** (`newMonth(today)`, not clamp-in-place). The
    sheet has a cell for every day of the month, even when SQLite has
    no row and even when the series is empty (O4, O5). Switching from
    June to November is required when today is in November.
4.  **No insert.** `t` does not `Upsert` and does not open the day
    form. Row count is unchanged.
5.  **Orthogonal gotos (O2).** Month `t` does not change the list
    cursor. List `t` does not change the month sheet. Esc after month
    `t` leaves the list where it was.
6.  **Editing.** While a month cell is being edited, `t` is text in
    the input (same as other printable keys), not Goto Today.
7.  **Help and docs.** List and month help mention `t`.
    `docs/reference/keys.md` lists it.
8.  **Out of scope.** Creating today. Charts. CLI. Startup selection
    (already shipped). Viewport scrolling. Coupling list and month
    cursors.

## Acceptance scenarios

Freeze today at 10 November 1990 when the date matters.

### S1 — List jumps to closest

**Given** logs on 1 June 1990 and 10 November 1990
**And** today is 10 November 1990
**And** the list is selected on 1 June 1990
**When** `t` is pressed
**Then** the selected list row is 10 November 1990
**And** no new log row is created

### S2 — List does not create today

**Given** logs on 1 November 1990 and 8 November 1990
**And** today is 10 November 1990
**And** the list is selected on 1 November 1990
**When** `t` is pressed
**Then** the selected list row is 8 November 1990
**And** there is still no 10 November 1990 row

### S10 — List nearest future (O1)

**Given** logs on 12 November 1990 and 20 November 1990
**And** today is 10 November 1990
**And** the list is selected on 20 November 1990
**When** `t` is pressed
**Then** the selected list row is 12 November 1990

### S11 — List tie prefers past (O1)

**Given** logs on 9 November 1990 and 11 November 1990
**And** today is 10 November 1990
**And** the list is selected on 11 November 1990
**When** `t` is pressed
**Then** the selected list row is 9 November 1990

### S12 — List nearer future beats last past (O1)

**Given** logs on 1 November 1990 and 11 November 1990
**And** today is 10 November 1990
**And** the list is selected on 1 November 1990
**When** `t` is pressed
**Then** the selected list row is 11 November 1990

### S3 — List empty is a no-op

**Given** no logs
**And** the daily list
**When** `t` is pressed
**Then** the screen is still the empty list
**And** no log row is created

### S4 — Month sheet jumps to today's day (O4)

**Given** logs on 1 June 1990
**And** today is 10 November 1990
**And** the month sheet is June 1990, focused on a non-weight column
**When** `t` is pressed
**Then** the sheet is November 1990
**And** the focused day is the 10th
**And** the focused column is weight
**And** no new log row is created

### S5 — Month sheet does not create today (O4)

**Given** logs on 8 November 1990
**And** today is 10 November 1990
**And** the month sheet is November 1990, focused on the 8th, note
column
**When** `t` is pressed
**Then** the focused day is the 10th
**And** the focused column is weight
**And** there is still no 10 November 1990 row

### S13 — Empty series, month `t` still goes to today (O5)

**Given** no logs
**And** the month sheet (opened with `m`)
**And** `[` has moved the sheet to October 1990
**When** `t` is pressed
**Then** the sheet is November 1990
**And** the focused day is the 10th
**And** the focused column is weight
**And** no log row is created

### S6 — Month `t` while editing is text

**Given** the month sheet with a cell being edited
**When** `t` is pressed
**Then** `t` is text in the input
**And** the focused day does not change to today as a Goto

### S7 — Form `t` is text

**Given** the day form, note field
**When** `t` is pressed
**Then** `t` is text in the field
**And** the list selection is unchanged after cancel

### S8 — Charts ignore `t` as Goto Today

**Given** the monthly chart or the long-term chart
**When** `t` is pressed
**Then** the screen is still that chart
**And** no log row is created

### S9 — Month `t` does not move the list (O2)

**Given** logs on 1 June 1990 and 10 November 1990
**And** today is 10 November 1990
**And** the list is selected on 1 June 1990
**And** the month sheet is June 1990
**When** `t` is pressed
**And** Esc returns to the list
**Then** the selected list row is still 1 June 1990

### S14 — Help mentions `t` (O6)

**Given** the daily list
**When** the screen is rendered
**Then** the help line contains `t`
**Given** the month sheet (not editing)
**When** the screen is rendered
**Then** the help line contains `t`

## Functional requirements

-   **FR1.** On the daily list, `t` selects the closest existing row
    to local today (startup distance rule, including nearer future
    and past-preferring ties) and does not create a log. Empty list:
    no change.
-   **FR2.** On the month sheet (not editing), `t` shows today's
    calendar month and focuses today's day, weight column, and does
    not create a log. This holds when the series is empty.
-   **FR3.** Month `t` does not change the list cursor. List `t` does
    not change the month sheet.
-   **FR4.** `t` is not Goto Today on the form or charts. On the form
    and while editing a month cell, `t` is typed text.
-   **FR5.** Help on list and month, and `docs/reference/keys.md`,
    mention `t`.

## Out of scope

-   Creating a missing today row
-   Goto Today on charts or CLI
-   Changing startup selection
-   Syncing list cursor from month `t`
