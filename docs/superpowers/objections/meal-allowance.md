---
spec: docs/superpowers/specs/2026-11-20-meal-allowance.md
date: 2026-11-20
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: medium
    claim: "A real overage smaller than half a tenth has no sign on screen, because the calorie-line printer this spec adopts turns that overage into the same `remaining 0.0` an exact match prints, and the sign is the only overage signal allowed."
    evidence: "Decision 4: \"Over it, remaining is negative. There is no separate overage word. An exact zero prints `0.0`, never `-0.0`.\" Out of scope: \"The sign and the words `no allowance` are the comparison.\""
    disposition: accepted
    disposition_rationale: >
      A difference whose tenth rounds to zero prints
      `remaining 0.0`. That covers an exact zero, a surplus
      smaller than half a tenth, and an overage smaller than
      half a tenth. "Over is negative" means the printed tenth
      has a minus. This line still turns a printed `-0.0` into
      `0.0`, the same way the weight column does. One scenario
      sets the allowance to 0 against an unrounded total of
      0.04 and expects `remaining 0.0`, and the screen does
      not show `no allowance`.
  - id: O2
    category: specification quality
    severity: medium
    claim: "The how-to text this slice requires states an unrounded subtraction, and it never mentions the one-decimal rule that Decision 4 and S6 make the digits on screen."
    evidence: "S19: \"otherwise shows remaining calories as the allowance minus the day's calorie total.\""
    disposition: accepted
    disposition_rationale: >
      The how-to says remaining is the allowance minus the
      unrounded calorie total, shown to one decimal the same
      way as the calorie line. It says 2.5 against a total of
      1.25 shows 1.2. It does not tell the reader to subtract
      the painted calorie text. S19 requires that sentence.
  - id: O3
    category: specification quality
    severity: medium
    claim: "Rejecting proposed `b` is specified only for the idle help line and \"the key row,\" while the scenarios, FR8, FR9, and the documentation section are satisfiable only by shipping the letter `b`."
    evidence: "Key proposal: \"`b` is not a settled decision.\" \"If the human rejects `b`, the idle help line and the key row change to the chosen letter.\""
    disposition: accepted
    disposition_rationale: >
      `b` is the allowance key. The key proposal, the idle
      help, the scenarios, FR8, FR9, the key reference, and
      the how-to all use `b`. The sentences that say the
      letter is unsettled, and that a rejection changes only
      the help line and the key row, come out.
  - id: O4
    category: specification quality
    severity: medium
    claim: "While the allowance field is open, the spec hides the line list and does not say whether the calorie total and the comparison line stay, so the target can be typed with or without the total it is judged against."
    evidence: "Decision 6: \"The line list is not shown while it is open.\""
    disposition: accepted
    disposition_rationale: >
      The allowance editor replaces the idle body. The line
      list, the calorie total, the macro totals, and the
      comparison line are not shown while the field is open.
      The heading keeps the date. Unset opens empty. Set
      opens on the stored number. S17 checks that `calories`,
      `no allowance`, and `remaining` are absent while the
      field is open.
---

## O1 — implementation — medium

### Claim

A real overage smaller than half a tenth has no sign on screen, because the calorie-line printer this spec adopts turns that overage into the same `remaining 0.0` an exact match prints, and the sign is the only overage signal allowed.

### Evidence

Decision 4 requires both a signed comparison and the calorie line's tenth printer, and it limits the `-0.0` ban to an exact zero:

> Remaining is the allowance minus the day's calorie total. The subtraction uses the unrounded total, the same sum the calorie line is drawn from, and then the calorie line's one-decimal half-even rule once. It does not subtract the painted calorie text. Under the allowance, remaining is positive. Over it, remaining is negative. There is no separate overage word. An exact zero prints `0.0`, never `-0.0`.

S5 then teaches the exact-match picture, and forbids the only other spelling of a rounded zero:

> Then the screen shows `remaining 0.0`
> And the screen does not contain `-0.0`

Out of scope removes every other way to see the comparison:

> A colour role for remaining calories is not added. The sign and the words `no allowance` are the comparison.

The same one-decimal print that Decision 4 cites for `1.25` → `1.2` and `0.25` → `0.2` renders a difference of `-0.04` as `-0.0`. The plan commits the implementation to that printer and to a second rewrite: "`formatTenth` is `%.1f`, and `-0.0` becomes `0.0`."

### Why this matters

The calorie line this rule comes from cannot be a small negative. Remaining can. Take a stored zero, which Decision 4 says is set, and an unrounded total of `0.04`. The tenth rule paints `calories 0.0`. The difference is `-0.04`, so the day is over, but `%.1f` yields `-0.0` and the committed rewrite yields `remaining 0.0`. That is the same screen as S5's exact match and as S7's empty day. A surplus of `0.04` also prints `0.0`, so "under is positive" fails in the same band. No scenario enters a difference in `(-0.05, 0.05)` other than exact zero, so a green S5 does not catch it. The spec says to apply the tenth rule once; the plan's clamp is a second step that deletes a true minus. Until the spec names the string for this band, "over is negative" and "never `-0.0`" cannot both be implemented.

## O2 — specification quality — medium

### Claim

The how-to text this slice requires states an unrounded subtraction, and it never mentions the one-decimal rule that Decision 4 and S6 make the digits on screen.

### Evidence

S19 and the documentation section require this sentence, and no other sentence about the arithmetic:

> And that how-to says the idle screen shows `no allowance` when none is set, and otherwise shows remaining calories as the allowance minus the day's calorie total

Decision 4 and S6 require a different result whenever the unrounded difference is a halfway tenth. S6, with no foods:

> When the allowance is `1.25`
> Then the screen shows `remaining 1.2`

And with a painted total of `1.2` from an unrounded `1.25`:

> When that date's allowance is `2.5`
> Then that date shows `remaining 1.2`
> And that date does not show `remaining 1.3`

Decision 3 keeps the allowance itself off the idle screen, so those two tenths are all the person can add.

### Why this matters

A how-to that follows S19 literally is false for the case the spec uses as its rounding contract. `2.5` minus the painted `1.2` is `1.3`, which S6 forbids, and `2.5` minus `1.25` is `1.25`, which the screen does not show. The idle screen also does not show `2.5`, so the person cannot recover the stored allowance from the two lines the how-to will have taught them to subtract. Implementers can satisfy S19 and S6 with two artifacts that disagree, and no acceptance check forces the how-to to mention one decimal, half-even, or "do not subtract the painted calorie text."

## O3 — specification quality — medium

### Claim

Rejecting proposed `b` is specified only for the idle help line and "the key row," while the scenarios, FR8, FR9, and the documentation section are satisfiable only by shipping the letter `b`.

### Evidence

The key proposal says the letter is not a decision, and it names the only things that change if it is rejected:

> This draft proposes `b` on the idle meal screen, and the scenarios below use that letter. `b` is not a settled decision. It is waiting for the human.

> If the human rejects `b`, the idle help line and the key row change to the chosen letter. This draft does not name another letter.

The acceptance contract does not follow that sentence. S19:

> Then the key reference has a meal row for `b` that edits that day's calorie allowance

> And `docs/how-to/meal-day.md` says `b` edits the allowance, Enter saves, Esc cancels, and a blank field clears it

FR9:

> `docs/reference/keys.md` gains one meal row for proposed `b` and does not gain a second door. `docs/how-to/meal-day.md` gains the allowance step in S19

The documentation section repeats the same letter, with no "or the letter the human chose" clause. Scenarios S3, S6, S7, S8, S9, S10, S11, S15, S17, and S18 all press `b`. The preamble says there is no production code until those scenarios exist as failing tests.

### Why this matters

This is not a request to pick a different letter. `b` can stand. The failure is that the escape hatch does not reach the artifacts a rejection would make false. An implementation that is green on S17, S19, FR8, and FR9 has already written `b` into the key reference and the how-to. The plan follows that reading: "The tests and the docs use that letter. Do not invent another key while this plan is draft." After that, "the key row" in the key proposal cannot mean `docs/reference/keys.md`, because FR9 has already fixed that row as `b`. A later human rejection has no specified edit list, and a literal reading of the whole spec cannot both change the letter and pass S19.

## O4 — specification quality — medium

### Claim

While the allowance field is open, the spec hides the line list and does not say whether the calorie total and the comparison line stay, so the target can be typed with or without the total it is judged against.

### Evidence

Decision 6 names the list and does not name the totals:

> The line list is not shown while it is open. Unset opens empty, not `0`. Set opens on the stored number, not the idle tenth.

S17, the scenario that opens the editor, checks the field, the heading, and the help line only:

> Then the field labeled `allowance` is focused
> And the field is empty
> And the heading still contains the meal date
> And the help line is the editor line in Decision 6

The idle comparison is specified only for the idle screen. Decision 3: "The comparison line is either `no allowance` or `remaining` and its number, after the totals that are shown." Nothing says those lines are removed when the field opens. The food form this screen extends does replace the whole idle body, but this spec does not say the allowance editor copies that layout. It says "the line list," which is not the totals block.

### Why this matters

Setting the allowance is the action US1 exists for. One layout hides `calories 200.0` while the person types `500`. The other leaves the total, and possibly `no allowance` or `remaining`, on screen under a field labeled `allowance`. Both pass S17. A careful reading of "the line list" keeps the totals; the existing non-idle branch drops them. The spec does not say which failure it is willing to ship: typing a target blind, or an editor that still shows the idle comparison the field is about to replace.

## Explicitly not objecting to

- **Proposing `b`:** Idle `b` is free on that screen and the proposal says it is not a second door. The objection is the rejection path, not the letter.
- **No carry-forward and no `config.toml` default:** Decision 1 and US3 require each date to keep only the number typed for it. Inheriting a target is the alternative the slice already refused, not an undetected failure.
- **A negative `remaining` when the tenth survives:** S4 locks `remaining -100.0` and forbids another overage word. That choice is specified. The hole is only the sub-tenth band in O1.
- **Leaving the allowance number off the idle screen:** Decision 3 forbids a second total, and Decision 4 shows the stored number in the editor, so Enter on `1.25` does not rewrite it as `1.2`.
- **The stale four-totals sentence in the slice:** The preamble calls that sentence stale, and S1 and S2 lock the shipped `track_macros` gate. The slice file is history, not the screen contract.
- **Refusing `kcal`, `cal`, and the other unit tokens:** They are outside the nutrient shape Decision 5 names, and this field is not the later amount parser.
- **Refusing `-0` and `-0.0`:** S8 and Decision 5 already require the negative error and `no allowance` after Esc, so a negative zero cannot become a stored zero.
- **A blank clearing without `want a number`:** Decision 5 and S10 state that on purpose, including a field of only spaces, as the opposite of a food number.
- **Weight-log isolation:** S12, S13, S14, S16, and S20 already require the daily log, both captions, `display_unit`, the in-flight save, and the reload behavior to stay as they are.
- **Caret keys and the selected line:** Backspace and arrows are how a one-field editor meets S10 and S6. The specified allowance keys do not include idle Up or Down, and nothing in those scenarios changes which line `>` marks.
