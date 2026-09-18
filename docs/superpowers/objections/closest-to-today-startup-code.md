---
spec: docs/superpowers/specs/2026-09-17-closest-to-today-startup.md
date: 2026-09-18
mode: code
diaboli_model: grok-4.6
objections:
  - id: O1
    category: risk
    severity: medium
    claim: "FR3's keep and keep-miss paths never run in the new tests, so deleting them would leave the suite green while month-cell reload and a vanished selected day snap to closest."
    evidence: "placeCursor keep is else if !keep.IsZero() after want := a.selectDay; saveMonthCell returns savedMsg{}; TestLoad* are New then Update(load); TestSave* call saveForm; TestLoadOutOfRangeSelectsClosest sets app.cursor = len(app.logs) + 1 on the same two-row series."
    disposition: accepted
    disposition_rationale: "Add a test: list on a non-closest day, month-cell save (including an earlier new day that shifts indexes), cursor stays on that calendar day. Keep-miss (selected day gone) can stay untested — the TUI cannot delete a row."
  - id: O2
    category: risk
    severity: medium
    claim: "Every freezeToday injects a time.UTC clock, so localToday is exercised only as UTC Date(); a regression to UTC civil today would still pass FR1 and mis-land the cursor when local date differs from UTC."
    evidence: "freezeToday: nowFn returns time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC); localToday: now := nowFn(); y, m, d := now.Date(); production nowFn is time.Now."
    disposition: accepted
    disposition_rationale: "Add a load test that freezes a Time whose local civil date differs from UTC (e.g. US evening). First load must follow Date() in that location, not UTC."
  - id: O3
    category: implementation
    severity: medium
    claim: "selectDay is one-shot App state that loadErrMsg does not clear and a zero-day savedMsg does not overwrite, so a failed post-form load plus a later month-cell save selects the stale written day instead of keeping the list cursor."
    evidence: "Update(savedMsg): if !msg.day.IsZero() { a.selectDay = msg.day }; loadErrMsg does not touch selectDay; saveMonthCell return savedMsg{}; placeCursor: want := a.selectDay then if !want.IsZero() before keep."
    disposition: accepted
    disposition_rationale: "Clear selectDay on loadErrMsg. A failed reload must not apply a form-save day to a later month-cell load."
---

## O1 — risk — medium

### Claim

FR3's keep and keep-miss paths never run in the new tests, so deleting
them would leave the suite green while month-cell reload and a vanished
selected day snap to closest.

### Evidence

`placeCursor` only keeps a calendar day when `selectDay` is zero.
`saveMonthCell` returns `savedMsg{}` (no written list day). Every
`TestLoad*` constructs `New` and loads once. `TestSave*` call
`saveForm` and take the `want` branch. `TestLoadOutOfRangeSelectsClosest`
sets `app.cursor` past `len(logs)` on the **same** series; it does not
shorten the store. Plan item 11 asked for a shorter series.

### Why this matters

FR3's other-reload clause is keep / keep-miss. Month-cell save is the
production trigger. If keep stored an index instead of a calendar day,
or were deleted, the new tests would still pass. A month-cell save after
the user moved off today would snap the list to closest.

- **accept-as-stated** — keep exists for month-cell / defensive S8; the
  TUI cannot delete a day, so keep-miss is unreachable in product use.
- **revise-spec** — if month-cell reload should not move the list
  cursor, FR3's other-reload clause is the wrong contract.
- **add-test** — month-cell save while the list is on a non-closest day
  keeps that day (including when an earlier row is inserted); a reload
  whose selected day is gone falls back to closest.
- **consciously-carry** — ship with keep untested because month-cell
  upsert cannot remove the kept day.

## O2 — risk — medium

### Claim

Every freezeToday injects a time.UTC clock, so localToday is exercised
only as UTC Date(); a regression to UTC civil today would still pass
FR1 and mis-land the cursor when local date differs from UTC.

### Evidence

All new selection tests freeze
`time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC)`. `freezeToday`
replaces `nowFn` with that `Time` unchanged. `localToday` takes
`Date()` in that value's Location, not `time.Local`. Production
`time.Now()` has Location `Local`.

### Why this matters

Decision 1 is local civil today. `nowFn().UTC()` would keep every
frozen fixture on 10 November 1990 and stay green. A user whose local
date is still the 9th or already the 11th would get the wrong closest
row.

- **accept-as-stated** — `localToday` is correct for `time.Now()`; UTC
  noon is a same-civil-date freeze. Record that tests pin the civil
  date, not Location.
- **revise-spec** — if today may be the UTC date of `nowFn`, Decision 1
  is stronger than the tests.
- **add-test** — freeze a Time whose UTC date differs from its local
  date and assert first load follows the local civil date.
- **consciously-carry** — tests remain UTC-only; local-vs-UTC is left
  to `time.Now()` in production.

## O3 — implementation — medium

### Claim

selectDay is one-shot App state that loadErrMsg does not clear and a
zero-day savedMsg does not overwrite, so a failed post-form load plus
a later month-cell save selects the stale written day instead of
keeping the list cursor.

### Evidence

Form save sets `a.selectDay = msg.day` then reloads. `loadErrMsg` does
not clear `selectDay`. A later month-cell save returns `savedMsg{}`, so
the `if !msg.day.IsZero()` guard does not clear the stash.
`placeCursor` then treats the stale form day as `want` and skips keep.

### Why this matters

FR3 after a month-cell save is keep-or-closest, not "the last form that
upserted before a failed reload." The sticky field completes the form
save on the wrong Cmd. No test drives `loadErrMsg` after
`savedMsg{day: ...}` and then a zero-day save.

## Explicitly not objecting to

- **Goto Today (`t`) unbound**: out of this slice by spec and plan.
- **Viewport scrolling / terminal fold**: Decision 8; the list is not a
  window.
- **Month-cell save leaving `savedMsg.day` zero on the success path**:
  Decision 6; O3 is only the failed-load interaction.
- **`openChart` unchanged**: S6 is `TestChartFollowsSelectedRow`.
- **DST / `Hours()/24`**: `dateOnly` rebuilds UTC midnight, so `Sub` is
  an exact 24h multiple.
- **Creating a missing today row on startup**: `closestLogIndex` only
  scans existing rows.
- **Decision 3 "both after today, earlier wins"** for unique days:
  later future days are strictly farther; duplicates are rejected.
- **Process-global `nowFn`**: freezeToday tests are not `t.Parallel()`.
