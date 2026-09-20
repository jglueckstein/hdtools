---
spec: docs/superpowers/specs/2026-09-19-goto-today.md
date: 2026-09-19
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: "List Goto Today asserts the full startup closest rule, but S1–S2 only lock today-if-present else latest past, so tests from the scenarios will accept a different selection than closestLogIndex."
    evidence: "Decision 2: 'Same calendar-day distance as 2026-09-17-closest-to-today-startup.md (today wins; ties prefer on-or-before today). Reuse closestLogIndex.' Acceptance S1 (today exists) and S2 (1 and 8 Nov 1990; no future rows). No analogue of the followed spec's S3 (nearest future), S4 (tie prefers past), or S10 (nearer future beats last past)."
    disposition: accepted
    disposition_rationale: "List t must match closestLogIndex. Add the missing shapes: nearest future only, yesterday/tomorrow tie (prefer past), nearer future vs last past (1 Nov + 11 Nov → 11 Nov)."
  - id: O2
    category: alternatives
    severity: high
    claim: "Decision 5's optional list-cursor coupling is not in idea.md or the user stories, and the hybrid both clobbers the list breadcrumb when today exists and leaves Esc far from today when it does not."
    evidence: "Decision 5 / FR3: 'If a row dated today exists, month-sheet t also selects that list row so Esc back to the list is on today. If today has no row, the list cursor is unchanged.' idea.md and US2 require only that the month sheet show today's month and focus today's day. S9 tests only the exists path."
    disposition: accepted
    disposition_rationale: "Orthogonal gotos, matching idea.md. Month t only moves the sheet. It does not change the list cursor. Drop Decision 5 / FR3 / S9. Esc after month t leaves the list where it was."
  - id: O3
    category: specification quality
    severity: medium
    claim: "FR3's else branch (month t must not move the list cursor when today has no row) has no acceptance scenario, so always applying list closest from the month sheet stays green."
    evidence: "FR3: 'If a log exists for today, month-sheet t also selects that list row. If not, the list cursor is unchanged.' S9 Given includes 10 November 1990. No scenario takes S5's fixture (8 Nov only), presses t, Esc, and asserts the list row is still 8 November (or June, or whatever it was)."
    disposition: rejected
    disposition_rationale: "Superseded by O2: there is no FR3 else-branch to test."
  - id: O4
    category: specification quality
    severity: medium
    claim: "FR2 requires landing on the weight column after month t, but S4 and S5 only assert the calendar day, so keeping sleep, steps, or note still passes."
    evidence: "Decision 3 / FR2: 'focuses that calendar day, weight column.' S4 Then: 'the sheet is November 1990' / 'the focused day is the 10th.' S5 Then: 'the focused day is the 10th.' Neither names the column. Plan algorithm notes: 'newMonth(today) (or set year/month/day and clamp)' — the or-branch leaves col unchanged."
    disposition: accepted
    disposition_rationale: "S4 and S5 Then include the weight column (not sleep/steps/note). Plan: newMonth(today), not clamp-in-place."
  - id: O5
    category: specification quality
    severity: medium
    claim: "Month-sheet t is specified for a civil calendar that exists without SQLite rows, but every month scenario has at least one log, so hoisting S3's empty no-op to both screens leaves an empty series stuck after [."
    evidence: "Decision 2 empty no-op is list-only: 'Empty list: no-op, no row created' (S3 Given: no logs, the daily list). Decision 3: 'The sheet has a cell for every day of the month, even when SQLite has no row.' S4–S6 and S9 all Given at least one log. FR2 has no empty-series exception."
    disposition: accepted
    disposition_rationale: "Empty series: month t still shows today's month and day 10. Do not copy the list's empty no-op. Add: no logs, month sheet, [ then t → November 1990, day 10, still no rows."
  - id: O6
    category: specification quality
    severity: low
    claim: "FR5 requires list/month help and keys.md to mention t, but no scenario fails if those are omitted."
    evidence: "Decision 7 / FR5: 'Help on list and month, and docs/reference/keys.md, mention t.' Acceptance S1–S9 never inspect help or the reference table. Plan test list items 1–9 are S1–S9 only; keys.md is a module-structure row, not a failing test."
    disposition: accepted
    disposition_rationale: "A test that list and month View() help mention t. keys.md stays a docs file in the plan."
---

## O1 — specification quality — high

### Claim

List Goto Today asserts the full startup closest rule, but S1–S2 only
lock “today if present, else latest past,” so tests written from the
scenarios will accept a different selection than `closestLogIndex`.

### Evidence

Decision 2:

> Same calendar-day distance as
> 2026-09-17-closest-to-today-startup.md (today wins; ties prefer
> on-or-before today). Reuse `closestLogIndex`.

The followed spec needed extra scenarios after review: S3 (nearest
future), S4 (tie prefers past), S10 (nearer future beats last past).
This spec’s list cases are S1 (10 November 1990 exists) and S2 (1 and
8 November 1990, both before today). The plan test list copies those
two plus empty.

S1 is satisfied by any rule that prefers today. S2 is satisfied by
“last on-or-before today.” That helper disagrees with Decision 2 when
every row is after today, when yesterday and tomorrow tie, and when
11 November is closer than 1 November.

### Why this matters

The project writes failing tests from acceptance scenarios, then
implements until green. Decision 2’s “reuse `closestLogIndex`” is not
a red test. A Goto Today that lands on the latest past row will pass
S1–S3 and still be wrong for series shapes the previous slice already
treated as in-scope.

## O2 — alternatives — high

### Claim

Decision 5’s optional list-cursor coupling is not in `idea.md` or the
user stories, and the hybrid both clobbers the list breadcrumb when
today exists and leaves Esc far from today when it does not.

### Evidence

`idea.md`:

> In the list it selects the closest existing row to today. In
> the month sheet it shows today's calendar month and focuses today's
> day.

US1 is list-only. US2 is month-only. Neither mentions Esc or the
other screen’s cursor.

Decision 5 / FR3:

> If a row dated today exists, month-sheet `t` also selects that list
> row so Esc back to the list is on today. If today has no row, the
> list cursor is unchanged.

S9 covers only the exists path.

### Why this matters

The two screens already have different legal landings: the list can
only select an existing row; the sheet has a cell for a missing
today. Decision 5 tries to align them only when they can name the
same day. That produces two failure classes, both user-visible:

1. Today has a log. List is on 1 June, month is June, user presses
   `t` to glance at today, then Esc. The June breadcrumb is gone.
2. Today has no log. Same gesture leaves the list on June. Esc after
   “Goto Today” is not near today.

Two coherent designs exist: (a) orthogonal gotos — month `t` only
moves the sheet; (b) month `t` always applies list closest as well,
so Esc is always as near today as the list can get. The hybrid
inherits both designs’ failure modes.

## O3 — specification quality — medium

### Claim

FR3’s else branch (month `t` must not move the list cursor when today
has no row) has no acceptance scenario, so always applying list
closest from the month sheet stays green.

### Evidence

FR3:

> If a log exists for today, month-sheet `t` also selects that list
> row. If not, the list cursor is unchanged.

S9 Given includes 10 November 1990. There is no Then that, after
month `t` with no 10 November row, Esc still shows the pre-`t` list
row.

### Why this matters

If Decision 5 stands, the surprising branch is the else: same key,
same Esc, different list outcome depending on whether today was
logged. An implementer who always sets
`cursor = closestLogIndex(...)` on month `t` still passes S9 and
ships the opposite of FR3 when today has no row.

## O4 — specification quality — medium

### Claim

FR2 requires landing on the weight column after month `t`, but S4 and
S5 only assert the calendar day, so keeping sleep, steps, or note
still passes.

### Evidence

Decision 3 / FR2: “focuses that calendar day, weight column.” S4 Then:
the sheet is November 1990 and the focused day is the 10th. S5 Then:
the focused day is the 10th. Neither names the column. The plan’s
or-branch “set year/month/day and clamp” leaves `col` unchanged.

### Why this matters

Tab-next-cell fills a row from weight. A TDD pass from S4–S5 can leave
the user on today’s note after `t`. The plan’s or-branch makes that
the path of least mutation.

## O5 — specification quality — medium

### Claim

Month-sheet `t` is specified for a civil calendar that exists without
SQLite rows, but every month scenario has at least one log, so
hoisting S3’s empty no-op to both screens leaves an empty series stuck
after `[`.

### Evidence

Decision 2 empty no-op is list-only (S3 Given: no logs, the daily
list). Decision 3: the sheet has a cell for every day even when
SQLite has no row. FR2 has no empty-series exception. S4–S6 and S9
all Given at least one log.

### Why this matters

Empty-list help already offers `m`. `openMonth` on an empty series
lands on `localToday()`, and `[` / `]` still move months. A shared
`if len(logs)==0 { return }` copied from the list makes `t` a no-op
on that sheet. S4 proves a jump to a month that has no row; it does
not prove a jump when there are no rows at all.

## O6 — specification quality — low

### Claim

FR5 requires list/month help and `keys.md` to mention `t`, but no
scenario fails if those are omitted.

### Evidence

Decision 7 / FR5. S1–S9 never inspect a help line or the reference
table. The plan lists `keys.md` as a file to touch, not as a test
that starts red.

### Why this matters

`t` is undiscoverable without chrome. A behaviour-green slice can
ship a key nobody is told about.

## Explicitly not objecting to

- **The premise that a mid-session return key is needed**: startup
  already selects closest on first load; US1 is after the user has
  arrowed away, and nothing in the current key table does that
  without creating today (`n`).
- **List closest vs month civil-today**: `idea.md` states both; the
  sheet has a cell for a missing today and the list does not.
- **Key `t` despite month type-to-edit**: the sheet already consumes
  `n`, `c`, `l`, and `q` before `beginEdit`; `t` is the usual
  calendar “today” binding.
- **Not creating today, and not opening the day form**: US3 and
  `idea.md` forbid inventing a row; `n` remains the create path.
- **Charts, CLI, startup selection, and viewport scrolling as out of
  scope**: `idea.md` names list and month sheet only.
- **Reusing `closestLogIndex` by name**: that avoids rule drift,
  provided the distinguishing scenarios exist (O1).
- **Form `t` as text (S7)**: load-bearing; not a form-level Goto.
- **Weight as the landing column (the choice)**: starting a day at
  weight matches Tab-next-cell; the objection is that S4/S5 do not
  lock it (O4).
