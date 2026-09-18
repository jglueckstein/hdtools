# Closest-to-today startup selection

**Date**: 2026-09-17
**Status**: approved
**Backlog**: [`idea.md`](../../../idea.md) (Daily Log)
**Objections**:
[`closest-to-today-startup.md`](../objections/closest-to-today-startup.md)
(O1 rejected; O2–O7 accepted)

On startup the daily list currently selects the first row (oldest,
because the store is chronological). `idea.md` asks for the row whose
date is closest to today. The monthly chart already follows the
selected row; this slice does not change that. Goto Today (`t`) is
backlog, not this slice.

## User stories

### US1 — Open near today

As a person opening the TUI, I want the daily list to land on the log
closest to today so I do not arrow through months of history.

### US2 — Chart still follows the list

As a person who then opens the monthly chart, I want that chart to
show the selected row's month, as it already does.

## Decisions

1.  **Calendar dates, local today.** Distance is whole days between
    log `Day` and local today (`localToday`). Clock time is ignored.
2.  **Closest is minimum absolute day count.** A row dated today
    always wins.
3.  **Ties prefer on-or-before today.** If two rows are equally far
    (yesterday and tomorrow), select the one that is not after today.
    If both are after today, select the earlier of those.
4.  **Empty list.** No row to select. The empty-state screen is
    unchanged.
5.  **First load.** After the initial store load, the list cursor is
    the closest row. That row may be index 0 (O6).
6.  **Reload.** After a successful save, select the calendar day that
    was written (`n` or Enter) if it is in the series (O7). Any other
    reload: if the previously selected calendar day is still present,
    keep it; otherwise select closest to today (O4). Closest may be
    index 0.
7.  **One cursor.** `>` and reverse highlight stay the same index.
    This slice does not add a second selection.
8.  **Out of scope.** Viewport scrolling (the list is not a window).
    Changing sort order. Selecting a month-sheet cell. CLI
    `-chart-pdf`. Creating a missing today row on startup (O1).
    Goto Today.

## Acceptance scenarios

Freeze today at 10 November 1990 when the date matters.

### S1 — Today is in the log

**Given** logs on 1 November 1990, 10 November 1990, and 20 November
1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 10 November 1990

### S2 — Nearest past

**Given** logs on 1 November 1990 and 8 November 1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 8 November 1990

### S3 — Nearest future

**Given** logs on 12 November 1990 and 20 November 1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 12 November 1990
(that row is index 0)

### S4 — Tie prefers past

**Given** logs on 9 November 1990 and 11 November 1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 9 November 1990

### S9 — Cross-month history (O2)

**Given** logs on 1 June 1990 and 10 November 1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 10 November 1990
**And** it is not 1 June 1990 (index 0)

### S10 — Nearer future beats last past (O3)

**Given** logs on 1 November 1990 and 11 November 1990
**And** today is 10 November 1990
**When** the TUI finishes loading
**Then** the selected list row is 11 November 1990

### S5 — Empty list

**Given** no logs
**When** the TUI finishes loading
**Then** the screen is the empty list
**And** there is no selected row

### S6 — Chart follows the selected row (O5)

**Given** S9's logs and selection (first row June, selected 10
November)
**When** `c` is pressed
**Then** the monthly chart title is `November 1990`
**And** it is not `June 1990`

### S7 — Save keeps a non-closest day (O4)

**Given** today is 10 November 1990
**And** the list selected on 1 November 1990
**When** that day's form is saved
**Then** after reload the selected list row is 1 November 1990
**And** it is not 10 November 1990

### S7a — Save from `n` selects the written day (O7)

**Given** today is 10 November 1990
**And** the list selected on 8 November 1990
**When** a new day form for 10 November 1990 is saved (`n`)
**Then** after reload the selected list row is 10 November 1990

### S8 — Out-of-range cursor falls back to closest (O6)

**Given** a loaded series whose selected index is no longer valid
(the series is shorter, or empty of that day)
**And** this reload is not a successful save
**When** the series is loaded again
**Then** the selected row is the closest remaining day to today
**And** if the series is empty, S5 applies

## Functional requirements

-   **FR1.** After the first successful load of a non-empty series,
    the list cursor is the row closest to local today (Decisions
    1–3). That row may be index 0.
-   **FR2.** An empty series has no selected row.
-   **FR3.** After a successful save, the list cursor is the calendar
    day that was written, if it is in the series. After any other
    reload, if the previously selected calendar day is still present,
    it stays selected; otherwise FR1 applies.
-   **FR4.** Opening the monthly chart from the list still uses the
    selected row's month.

## Out of scope

-   Jumping the month sheet to today independently of the list
-   Long-term chart start position
-   Creating a missing "today" row on startup
-   Goto Today (`t`) in the daily list
