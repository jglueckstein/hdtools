---
spec: docs/superpowers/specs/2026-10-01-month-sheet-edit-keys.md
date: 2026-10-01
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: "The amendment updates only the Tab spec's out-of-scope line, leaving its rule that Tab saves the same way Enter does aimed at an Enter that now moves down a day."
    evidence: "Amends: that spec's out of scope (Enter while editing no longer saves in place). FR8: Tab still saves and moves to the next cell as in the Tab spec. Followed spec FR1: Tab saves that cell the same way Enter does, then moves to the next cell. Followed spec Decision 2: Tab does the same save as Enter, then moves."
    disposition: accepted
    disposition_rationale: "The Tab spec's same save as Enter means the persist step only (display unit, kilograms, failed save does not move). Enter's vertical move belongs to this spec. Amend the Tab spec's opening line, Decision 2, FR1, and the out-of-scope bullet that says Enter still saves without advancing, and point this spec's Amends line at those sentences. Tab's destination stays the FR8 parenthetical and S10. The vertical save must not call Tab's next-cell move."
  - id: O2
    category: scope
    severity: medium
    claim: "Editing-mode Enter and arrows change, but the spec neither requires nor rules out updates to the month-sheet help line and docs/reference/keys.md, which still describe arrows as moving cells and Enter as opening the form."
    evidence: "Decision 1: It saves, then moves down in the current column. Decision 10: When not editing, arrows and Enter are unchanged (arrows move cells; Enter still opens the form). Out of scope lists List, form, charts, CLI, and PDF, Goto Today, Shift+Tab, Changing Tab, and Space, and does not mention the help line or keys.md."
    disposition: accepted
    disposition_rationale: "docs/reference/keys.md is in scope. While editing, Enter and Down accept and move down, Up accepts and moves up, and Left and Right move the caret. Not-editing arrows and Enter stay as they are. The month help line stays the not-editing chord list (arrows move, enter form). Say that in the spec so a build cannot treat the footer as a second contract."
  - id: O3
    category: specification quality
    severity: medium
    claim: "S6's second step says Given the same setup after Enter has already ended editing, so Down on the last day can pass without being pressed during an edit."
    evidence: "S6 Then: editing has ended, and the focus is still the weight cell on 30 November. Then: Given the same setup. When Down is pressed. Then: editing has ended, and the focus is still the weight cell on 30 November."
    disposition: accepted
    disposition_rationale: "Replace S6's Given the same setup with a full Given, including editing 80.0, the way S7 already does. Down on 30 November has to be pressed during that edit."
  - id: O4
    category: risk
    severity: medium
    claim: "A second Enter after a successful accept opens the day form on the destination day, because the accept ends editing and a not-editing Enter opens the form, and no scenario covers that sequence."
    evidence: "Enter currently saves in place. Decision 1: It no longer saves in place. Decision 5: After a successful save, editing has ended. The destination cell is focused, not in edit mode. FR10: Enter still opens the day form. S11 covers Enter only from an already not-editing cell."
    disposition: accepted
    disposition_rationale: "After a successful Enter, Down, or Up, the next Enter is FR10 on the cell that is now focused: it opens the day form there and writes nothing. On the last day that cell is the same day. Record that as one scenario. Do not swallow the second Enter, and do not leave the destination in edit mode."
---

## O1 — specification quality — high

### Claim

The amendment updates only the Tab spec's out-of-scope line, leaving its
rule that Tab saves the same way Enter does aimed at an Enter that now
moves down a day.

### Evidence

This spec cites the Tab spec as what it follows, and limits the amendment
to one bullet:

> **Amends**: that spec’s out of scope (Enter while editing no longer
> saves in place)

It then defines Enter's persistence as Tab's, and Tab's behavior as the
followed spec's:

> **Same save as today.** Accept stores the cell the same way Tab
> already does (display unit, kilograms in the log).

> **FR8.** While editing, Tab still saves and moves to the next cell as
> in [2026-09-10-month-tab-next-cell.md](2026-09-10-month-tab-next-cell.md)
> (weight → sleep → steps → workout → note → next day’s weight; last
> cell of the month stays).

The followed spec's normative sentences, which this amendment does not
replace, are:

> **Save then advance.** While editing, Tab does the same save as Enter,
> then moves.

> **FR1.** While editing a month cell, Tab saves that cell the same way
> Enter does, then moves to the next cell.

> Changing Enter (it still saves without advancing)

### Why this matters

Enter's save and Enter's navigation are one action in this spec
(Decision 1, FR1). The Tab spec still says Tab saves the same way Enter
does and then does its own move. Read together, weight on day 4 can
become sleep on day 5: save, move down, then next cell. S10 and the FR8
parenthetical say the opposite (sleep on the same day), so a careful
reader of this file alone is steered back. The hazard is the pair. FR8
tells the implementer to obey the Tab spec, and that spec is now false:
Enter no longer saves without advancing, but only its out-of-scope line
is amended. A later edit that "aligns Tab with Enter" by calling the new
Enter path will double-move, and both documents will look authoritative.
The amendment has to reword Tab Decision 2 and FR1 so "the same way
Enter does" means the patch-and-upsert only, not Enter's new vertical
move. Accepting this spec as written leaves an accepted spec citing a
contradicted one.

## O2 — scope — medium

### Claim

Editing-mode Enter and arrows change, but the spec neither requires nor
rules out updates to the month-sheet help line and
`docs/reference/keys.md`, which still describe arrows as moving cells
and Enter as opening the form.

### Evidence

> **Enter while editing changes.** It saves, then moves down in the
> current column (the same destination as Down). It no longer saves in
> place.

> **When not editing, arrows and Enter are unchanged** (arrows move
> cells; Enter still opens the form).

The out-of-scope list names List, form, charts, CLI, and PDF, Goto
Today, extra daily fields, delta, Shift+Tab, Changing Tab, Space, and
wrapping to another month. It does not mention the keyboard reference
or the on-screen help. No functional requirement does either.

### Why this matters

`docs/reference/keys.md` says `Arrows | month | move cells` and has no
month-sheet Enter row. The month sheet help line is `arrows move   type
edit   tab next   space workout   enter form   ...`. Those sentences
have no editing qualifier. Not-editing behavior stays true (Decision
10). While editing, this slice makes Up and Down save and change day,
and makes Enter leave the cell. Left and Right were already caret
motion, so "arrows move" was already a not-editing summary; Up, Down,
and Enter were not this kind of contradiction. Two spec-compliant
builds both pass S1–S15: one updates the hint and the reference, one
ships the keys and leaves the only in-product instructions denying
them. The Goto Today spec treated help text and `docs/reference/keys.md`
as an FR for a smaller key change. This slice needs the same decision,
or an explicit out-of-scope line, or the false hint ships because
nothing in the contract mentions it.

## O3 — specification quality — medium

### Claim

S6's second step says Given the same setup after Enter has already
ended editing, so Down on the last day can pass without being pressed
during an edit.

### Evidence

> **Given** the month sheet for November 1990, focused on 30 November,
> weight column, editing `80.0`
> **When** Enter is pressed
> **Then** the 30 November weight is stored as 80 kg
> **And** editing has ended
> **And** the focus is still the weight cell on 30 November
> **And** the sheet is still November 1990
>
> **Given** the same setup
> **When** Down is pressed
> **Then** the 30 November weight is stored as 80 kg
> **And** editing has ended
> **And** the focus is still the weight cell on 30 November
> **And** the sheet is still November 1990

S7, for Up on day 1, repeats the full Given, including `editing 80.0`.
S6 does not. "The same setup" appears nowhere else in the specs.

### Why this matters

FR5 requires both Enter and Down, while editing, on the last day, to
save and stay. One reading of "the same setup" resets to `editing
80.0`. The other continues: Enter has ended editing and stayed on 30
November, then Down is the not-editing Down that already clamps on the
last day. That second run stores nothing new, and every Then is already
true from the Enter half. An Enter-only implementation passes. The
phrase is the only acceptance text for Down at the edge. Repeat the
S7-style Given, and say the cell is still being edited when Down is
pressed.

## O4 — risk — medium

### Claim

A second Enter after a successful accept opens the day form on the
destination day, because the accept ends editing and a not-editing
Enter opens the form, and no scenario covers that sequence.

### Evidence

> Enter currently saves in place. This slice makes Enter and Down accept
> the edit and move down the column

> **Enter while editing changes.** It saves, then moves down in the
> current column (the same destination as Down). It no longer saves in
> place.

> **After a successful save**, editing has ended. The destination cell
> is focused, not in edit mode. The column is unchanged.

> **FR10.** When not editing, Left and Right still change column, Up and
> Down still change day, and none of those keys write. Enter still opens
> the day form.

S11 presses Enter only on a cell that was not editing to begin with.
Nothing presses Enter twice.

### Why this matters

The two rules compose, and the composition is a change from the
behavior this spec names. Today Enter saves in place, so a second Enter
opens the form on the day just saved. After Decision 1 and Decision 5,
the first Enter leaves that day, and the second opens the form on the
next day, often an empty one, and column entry is over. US1 types a
value between Enters ("the cell I just typed"), so the happy path is
unaffected. The unstated path is confirmation, key repeat, or a
spreadsheet habit of Enter to walk. All of those now open the form one
row below the value just stored. On the last day the first Enter
already stays (FR5), so the second Enter opens the form on that same
day — a different two-press result from mid-month, also unstated. Write
the sequence down, including whether it is intended, or the footgun
ships as an accident of Decision 5 plus FR10.

## Explicitly not objecting to

- **Column-fill premise**: While a cell is editing, Up and Down do not
  move the day, and `idea.md` asks for Enter, Down, and Up to accept
  and move, so this slice is aimed at a real gap.
- **Decision 5 ending the edit on arrival**: S1 already forbids arriving
  in edit mode, and that matches Tab. The next value is the existing
  type-to-edit behavior, which this slice does not retarget.
- **Moving Enter rather than only Down**: The backlog asks for Enter to
  move. The objections are the unfinished Tab-spec amendment and the
  unstated second Enter, not a veto of that request.
- **Caret step size**: US3 and FR7 require the caret to move and the
  cell to stay. That is the current text input, not a second caret
  design.
- **Blank versus invalid text**: Decision 4 and Decision 13 delegate
  failure and persistence to today's Tab save, including a blank weight
  clearing instead of a new validator.
- **Last-day save that ends editing**: FR5 and FR6 match Tab staying on
  the last cell without changing month. November 1990 is the fixture;
  "last day" and "day 1" are the rule.
- **Shift+Tab, Space, and delta**: They are named out of scope, workout
  is not text-edited, and FR8's focus order already skips delta.
