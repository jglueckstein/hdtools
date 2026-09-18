---
spec: docs/superpowers/specs/2026-09-17-closest-to-today-startup.md
date: 2026-09-17
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: premise
    severity: high
    claim: "The typical first open of the day has no today row, so closest-plus-Enter edits yesterday while creating today is out of scope, which is not the daily weigh-in US1 describes."
    evidence: "US1: 'land on the log closest to today so I do not arrow through months of history'; S2 Then: 'the selected list row is 8 November 1990' with today 10 November; Out of scope: 'Creating a missing \"today\" row on startup'."
    disposition: rejected
    disposition_rationale: "idea.md asked for the closest existing row, not a new today. S2 (yesterday) is the intended morning path. n still creates today. US1 is don't page through history, not land on a blank today."
  - id: O2
    category: specification quality
    severity: high
    claim: "Every dated scenario stays inside November 1990, so a current-month-only closest search passes FR1 while US1 is about months of history."
    evidence: "Acceptance scenarios: 'Freeze today at 10 November 1990 when the date matters'; S1–S4 and S6–S8 use only 1/8/9/10/11/12/20 November 1990; US1: 'I do not arrow through months of history'."
    disposition: accepted
    disposition_rationale: "Add a cross-month load: e.g. 1 June 1990 and 10 November 1990, today 10 Nov → select 10 Nov, not June (index 0)."
  - id: O3
    category: specification quality
    severity: high
    claim: "S1–S4 are also satisfied by last-on-or-before-today else first-after, so Decision 2's minimum absolute day count is not locked."
    evidence: "Decision 2: 'Closest is minimum absolute day count'; S2 only past; S3 only future; S4 equal-distance past and future; no mixed series with a strictly nearer future than the last past."
    disposition: accepted
    disposition_rationale: "Add a mixed series where the future is strictly nearer: 1 Nov and 11 Nov, today 10 Nov → select 11 Nov (1 day vs 9). That locks min absolute distance, not last-on-or-before."
  - id: O4
    category: specification quality
    severity: high
    claim: "FR3's keep-calendar-day reload is not forced by S7 or S8: always-FR1 and keep-index-with-OOB-closest both pass."
    evidence: "FR3: 'if the previously selected calendar day is still present, it stays selected'; S7 Given: 'the list selected on 10 November 1990' (frozen today); S7 Then still 10 November; S8 Given: 'selected index is no longer valid'."
    disposition: accepted
    disposition_rationale: "S7 must keep a non-closest day: selected 1 Nov, save that form, still 1 Nov while today is 10 Nov. That rules out always-FR1 on load."
  - id: O5
    category: specification quality
    severity: medium
    claim: "S6 cannot fail if the chart uses the first row, the selected row, or localToday's month, because every S1 date and frozen today is November 1990."
    evidence: "S6: 'Given S1's logs and selection When c is pressed Then the monthly chart title is November 1990'; S1 logs are 1, 10, and 20 November 1990 with today 10 November 1990."
    disposition: accepted
    disposition_rationale: "S6 uses the O2 fixture: first row June, selected 10 Nov → c title is November 1990, not June 1990."
  - id: O6
    category: specification quality
    severity: medium
    claim: "Decision 5 and S8 assert 'not index 0' / 'not the first row' as if that were an invariant, but S3's closest row is index 0."
    evidence: "Decision 5: 'the list cursor is that closest row, not index 0'; S8 Then: 'the closest remaining day to today, not the first row'; S3 Then: '12 November 1990' from logs on 12 and 20 November (first row)."
    disposition: accepted
    disposition_rationale: "Reword Decision 5 and S8: the cursor is the closest row, which may be index 0. Drop not the first row as a Then."
  - id: O7
    category: implementation
    severity: medium
    claim: "FR3 keep-day on every reload means saving a new today from n leaves the list on the previous closest row, so the next Enter edits the old day."
    evidence: "Decision 6: 'Keep the cursor on the same calendar day if that day is still in the series'; FR3: 'it stays selected'; S7 only saves 'that day's form' (the already-selected day)."
    disposition: accepted
    disposition_rationale: "After a successful save, select the calendar day that was written (n or Enter). First load is still FR1. A reload with no just-saved day keeps FR3's keep-or-closest rule."
---

## O1 — premise — high

### Claim

The typical first open of the day has no today row, so
closest-plus-Enter edits yesterday while creating today is out of
scope, which is not the daily weigh-in US1 describes.

### Evidence

US1:

> As a person opening the TUI, I want the daily list to land on the
> log closest to today so I do not arrow through months of history.

S2 freezes today at 10 November 1990, logs 1 and 8 November, and
requires the selected row to be 8 November. Out of scope:

> Creating a missing "today" row on startup

The list's primary key on a non-empty series is Enter, documented as
edit selected day.

### Why this matters

A Hacker's Diet session usually starts before today is logged. Under
this spec that session is S2, not S1: the cursor is yesterday.
Enter then overwrites the last weigh-in. `n` still creates today, but
US1 is written as landing where the user will work, and the spec
refuses to put today under that cursor unless the row already exists.
Shipping S1 as the happy path hides the modal case.

## O2 — specification quality — high

### Claim

Every dated scenario stays inside November 1990, so a current-month-only
closest search passes FR1 while US1 is about months of history.

### Evidence

The acceptance block opens with:

> Freeze today at 10 November 1990 when the date matters.

S1–S4 and S6–S8 only name 1, 8, 9, 10, 11, 12, and 20 November 1990.
S5 is empty. US1's reason is "months of history." Decision 1 says
year/month/day, but no scenario has a log outside that November.

### Why this matters

The motivating series is many months of chronological rows. An
implementation that minimises distance only among rows in
`localToday`'s month, and otherwise leaves index 0, satisfies every
written Then. The oldest-month `c`/`m` failure US1 exists to prevent
would remain.

## O3 — specification quality — high

### Claim

S1–S4 are also satisfied by last-on-or-before-today else first-after,
so Decision 2's minimum absolute day count is not locked.

### Evidence

Decision 2:

> Closest is minimum absolute day count. A row dated today always
> wins.

S1 has today. S2 is only past (last past). S3 is only future (first
future). S4 is yesterday and tomorrow (tie prefers past, which is also
the last on-or-before). There is no mixed series where a future row is
strictly nearer than the last past row (for example 1 November and 11
November with today 10 November).

### Why this matters

A sorted-list lower bound is the usual "go to today" and will be the
first helper many implementers write. It passes S1–S4 and then, when
the month sheet has been filled ahead with a gap around today, selects
the last weigh-in instead of the nearer future row Decision 2 names —
or the reverse if they invert it. Either way FR1 is green and wrong.
The month sheet is built for entering a whole month in one sitting, so
that mix is not exotic.

## O4 — specification quality — high

### Claim

FR3's keep-calendar-day reload is not forced by S7 or S8:
always-FR1 and keep-index-with-OOB-closest both pass.

### Evidence

FR3:

> After a reload, if the previously selected calendar day is still
> present, it stays selected. Otherwise FR1 applies (not index 0).

S7 selects 10 November, saves that day's form, and still wants 10
November — the same day FR1 would pick with today frozen at 10
November. S8's Given is only an invalid index ("the series is shorter,
or empty of that day"), not "the user is on a non-closest day that
still exists at a new index."

### Why this matters

The naive hook is "on `loadedMsg`, run FR1." That snaps the cursor
back to closest after every save once the user has moved, and S7 will
not catch it. The other naive hook is "keep the index unless OOB, then
FR1," which passes S7–S8 and still loses the selected calendar day when
a new row is inserted before the cursor. FR3 is the only statement of
the intended policy, and the scenarios do not uniquely determine it.

## O5 — specification quality — medium

### Claim

S6 cannot fail if the chart uses the first row, the selected row, or
localToday's month, because every S1 date and frozen today is November
1990.

### Evidence

S6:

> **Given** S1's logs and selection
> **When** `c` is pressed
> **Then** the monthly chart title is `November 1990`

S1's logs are 1, 10, and 20 November 1990; today is 10 November 1990.
Index 0, the closest row, and `localToday()` all yield November 1990.
FR4 says the chart still uses the selected row's month.

### Why this matters

This slice's observable effect on `c` is exactly when the closest row
is in a different month from the first row (the history US1 names). S6
cannot detect a chart that still keys off index 0, and cannot detect a
chart that was "fixed" to today's month independently of the list. US2
is a non-regression that the fixtures make tautological.

## O6 — specification quality — medium

### Claim

Decision 5 and S8 assert "not index 0" / "not the first row" as if
that were an invariant, but S3's closest row is index 0.

### Evidence

Decision 5:

> After the initial store load, the list cursor is that closest row,
> not index 0.

S8 Then:

> the selected row is the closest remaining day to today, not the
> first row

S3 Given logs on 12 and 20 November 1990, today 10 November; Then the
selected row is 12 November 1990 — the first row of a chronological
series.

### Why this matters

"Not the first row" is a contrast with today's clamp-to-0, not a
postcondition. A test written from S8's Then can fail a correct
closest-is-index-0 result (S3, a one-row series, or an all-future
series). A test written from Decision 5 can treat any cursor 0 as a
bug. The plan already hedges ("not 0 unless that row is closest"); the
spec does not.

## O7 — implementation — medium

### Claim

FR3 keep-day on every reload means saving a new today from `n` leaves
the list on the previous closest row, so the next Enter edits the old
day.

### Evidence

Decision 6:

> Reload after save. Keep the cursor on the same calendar day if that
> day is still in the series.

FR3 restates that for any reload. S7 only covers saving the form of
the already-selected day. The spec never mentions saving a different
calendar day than the one the list is on.

### Why this matters

Morning path (O1): list on yesterday, `n` logs today, save reloads,
yesterday is still present so it stays selected. Today appears as
another row. Enter edits yesterday again. That is FR3 executed
faithfully. A policy of selecting the day just written would not need
a new product idea; the spec's reload rule rules it out without a
scenario that shows the `n` path.

## Explicitly not objecting to

- **The problem of cursor 0 on a chronological list**: `c`, `m`, and
  Enter currently act on the oldest row after load; that is a real
  startup miss independent of `n`.
- **Calendar dates and `localToday`, ignoring clock time**: Decision 1
  is a clear comparison rule and matches how log days are already
  stored.
- **One cursor for `>` and reverse highlight**: Decision 7 correctly
  refuses a second selection field.
- **Empty list unchanged**: Decision 4 / FR2 / S5 are consistent; there
  is no row to select.
- **Leaving sort order chronological**: newest-first would also move
  "today" toward the start of the dump, but it would break the paper
  month-sheet order the TUI already copies.
- **Last-row-only selection**: S1 already falsifies it when today is
  not the latest log.
- **Viewport scrolling**: Decision 8 is consistent with a non-window
  list; US1 is which row commands act on, not whether that row is
  clipped.
- **Creating a today row, long-term chart position, and CLI
  `-chart-pdf`**: those are named out of scope and are not required to
  make FR1 true when a closest row exists.
- **Tie prefers on-or-before today**: S4 uniquely forces that
  past-vs-future tie; it is not a slogan left untested.
- **A jump key instead of auto-select**: `idea.md` asked for startup
  selection, not an extra binding; US1 would fail if the default
  remained index 0.
