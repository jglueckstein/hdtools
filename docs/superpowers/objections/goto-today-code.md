---
spec: docs/superpowers/specs/2026-09-19-goto-today.md
date: 2026-09-20
mode: code
diaboli_model: grok-4.6
objections:
  - id: O1
    category: risk
    severity: high
    claim: "TestGotoTodayIgnoredOnChart cannot detect Goto Today on the charts because the fixture is already localToday's month and the assertions only check screen and row count."
    evidence: "twoDayApp seeds November 1990; TestGotoTodayIgnoredOnChart freezeToday 10 Nov 1990, press c then t, asserts screenChart and len(logs)==2; updateChart has no case t."
    disposition: accepted
    disposition_rationale: "Add a test: monthly chart, [ to October, t, still October and still the chart. Same idea is enough for long-term if cheap (longKind / screen unchanged)."
  - id: O2
    category: risk
    severity: high
    claim: "FR3's rule that list t does not change the month sheet is untested, and New() already sets a.month to localToday(), so wiring list t to newMonth(localToday()) would stay green."
    evidence: "updateList case t: a.cursor = closestLogIndex(...); New: month: newMonth(localToday()); no list TestGotoToday* reads a.month after t."
    disposition: accepted
    disposition_rationale: "Add a test: month sheet [ to October, Esc to list, list t; a.month is still October (do not reopen with m)."
  - id: O3
    category: risk
    severity: medium
    claim: "List and month help label the key t today even when t is a no-op, selects a day that is not today, or types text into an edit."
    evidence: "empty list help: t today; populated: t today; month.view always t today; updateMonth editing forwards t to textinput.Update."
    disposition: rejected
    disposition_rationale: "today is the mnemonic, not a promise the landing day is today. FR5 only requires mentioning t. Empty-list t is a specified no-op; n remains create. Month help is one string while editing, like the other non-edit keys."
  - id: O4
    category: implementation
    severity: medium
    claim: "docs/reference/keys.md documents month t as today's month and day only, omitting the specified reset to the weight column."
    evidence: "| `t` | month (not editing) | today's month and day |"
    disposition: accepted
    disposition_rationale: "keys.md: month t is today's month and day, weight column."
  - id: O5
    category: risk
    severity: low
    claim: "S14's test never renders the empty-list help string, so removing t from that line would not fail TestGotoTodayHelpMentionsT."
    evidence: "TestGotoTodayHelpMentionsT seedDays 1990-11-10 then helpHasKey on View(); empty help is a different Render string."
    disposition: accepted
    disposition_rationale: "Assert empty-list help mentions t (e.g. in TestGotoTodayListEmptyDoesNothing)."
---

## O1 — risk — high

### Claim

`TestGotoTodayIgnoredOnChart` cannot detect Goto Today on the charts.
The fixture is already `localToday()`'s month, and the assertions only
check that the screen did not change and that no row was inserted.

### Evidence

`twoDayApp` seeds 1 and 2 November 1990. The test freezes today at 10
November 1990, opens the monthly chart (`openChart` copies November
into `a.month`), presses `t`, and checks only `screenChart` and row
count. The long-term half is the same. `updateChart` / `updateLong`
have no `"t"` case. Chart paint uses `a.month.year` / `a.month.month`.

### Why this matters

A later change can add `a.month = newMonth(localToday())` to
`updateChart` and keep showing the chart. On this fixture that
assignment is identical to ignore: the chart is already November.
S8 stays green while FR4 is violated. A user who had `[` to October
would see the chart jump to today.

- **accept-as-stated** — S8 only requires staying on the chart and not
  inserting a row.
- **revise-spec** — S8 must start from a non-today month and say the
  plotted month is unchanged.
- **add-test** — from the monthly chart, `[` to October, press `t`,
  still October (and still the chart).
- **consciously-carry** — ship S8 as screen-and-rows only.

## O2 — risk — high

### Claim

FR3’s rule that list `t` does not change the month sheet is untested.
`New()` already sets `a.month` to `localToday()`, so wiring list `t` to
`newMonth(localToday())` would stay green.

### Evidence

List `t` only assigns `a.cursor`. Startup already points `a.month` at
today. No list `TestGotoToday*` reads `a.month` after `t`. The only
orthogonality test is `TestGotoTodayMonthDoesNotMoveList` (month `t`,
then list cursor). FR3: “List `t` does not change the month sheet.”

### Why this matters

Adding `a.month = newMonth(localToday())` next to the cursor assignment
matches every list test. Asserting after a following `m` is the wrong
test: `openMonth` rebuilds from the list cursor. The missing assertion
is on `a.month` after list `t` without reopening (sheet previously
navigated off today, then Esc).

## O3 — risk — medium

### Claim

List and month help label the key `t today` even when `t` is a no-op,
selects a day that is not today, or types text into an edit.

### Evidence

Empty list help includes `t today` (S3: `t` does nothing; `n` creates
today). Populated list help is the same while S2 lands on 8 November.
Month help is one string while editing and not; while editing, `t` goes
to `textinput.Update`. `keys.md` already says closest to today.

### Why this matters

Empty-list users may take `t today` as create-today. Populated list
without a today row still says today. Editing still advertises Goto.
S14 only locks the grapheme `t`.

- **accept-as-stated** — `today` is the mnemonic; FR5 only requires
  mentioning `t`.
- **revise-spec** — help must say closest / calendar day; must not
  advertise Goto on empty list or while editing.
- **add-test** — empty-list / editing help copy.
- **consciously-carry** — ship `t today` on all three strings.

## O4 — implementation — medium

### Claim

`docs/reference/keys.md` documents month `t` as today’s month and day
only, omitting the specified reset to the weight column.

### Evidence

`| t | month (not editing) | today's month and day |`

Decision 3 / FR2 / S4 / S5 require the weight column. The code
replaces the model (`col` zero is `colWeight`). Tests assert
`colWeight`. The table does not mention the column.

### Why this matters

A user on the note column who follows the keys table will not expect
to land on weight. S4 exists because that reset is user-visible.

## O5 — risk — low

### Claim

S14’s test never renders the empty-list help string, so removing `t`
from that line would not fail `TestGotoTodayHelpMentionsT`.

### Evidence

The help test seeds 10 November 1990 (populated list). Empty help is a
different `Render` string. `TestGotoTodayListEmptyDoesNothing` does not
assert help. FR5: help on the list mentions `t`. The empty list is
still the list.

### Why this matters

Two list help strings, one test. Dropping `t` from the empty line
leaves S3 and S14 green. Harm is limited because `t` there does
nothing (O3 is the sharper empty-list problem).

## Explicitly not objecting to

- **Creating a today row / Upsert on `t`**: out of this slice; `n`
  still opens today’s form.
- **Spreadsheet Enter/Up/Down while editing, and ntcharts**: out of
  slice.
- **Case-sensitive `t` vs `T`**: same unmatched-rune rule as other
  month keys.
- **`freezeToday` injecting UTC**: inherited from startup selection.
- **Viewport scrolling after list `t`**: spec out of scope.
- **`newMonth` relying on `col` zero for weight**: S4/S5 already
  require `colWeight`.
- **Premise of key `t` and orthogonal gotos**: adjudicated at spec
  time; current code follows those dispositions.
