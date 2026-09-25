---
spec: docs/superpowers/specs/2026-09-22-delta-column.md
date: 2026-09-22
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: critical
    claim: "S5 requires a blank delta on the first weigh-in, but Decision 1 and ApplyTrend make that day have both operands and a visible trend, so the spec cannot be implemented as written."
    evidence: "S5 Given: 'a day with a weight and no trend (first weigh-in, no prior)' Then 'the delta cell is empty'. Decision 1: both operands must be present, including 'on the list, the same ApplyTrend value already used for the trend cell'."
    disposition: accepted
    disposition_rationale: "Withdraw S5 as written. ApplyTrend gives the first weigh-in a trend equal to the weight, so the cell is 0.0. Blank only when a weight or a trend is actually missing (S4: carry, no weigh-in)."
  - id: O2
    category: premise
    severity: high
    claim: "US1 is the on-screen weight minus the on-screen trend; the spec instead converts a kilogram difference, which can disagree with the two cells the user is looking at."
    evidence: "US1: 'the difference between that day's weight and its trend so I do not subtract in my head'. Decision 1: 'Delta is computed in kilograms (weight − trend), then converted with units.FromKG'. S6: 0.45359237 kg → '+1.0'."
    disposition: accepted
    disposition_rationale: "US1 is the two on-screen cells. Delta is FromKG(weight) − FromKG(trend), then formatted. Rewrite S6 so displayed lb/st cells add up; do not convert a kg residual."
  - id: O3
    category: specification quality
    severity: high
    claim: "The one-decimal display format does not say whether zero is exact or rounded, how the sign is chosen, or what happens in stone, so implementations will print +0.0/−0.0 or collapse real daily deltas to 0.0."
    evidence: "Decision 3: 'One decimal, display unit implied by the weight column... Zero: 0.0 with no sign.' FR2: 'signed one-decimal display of FromKG(weightKG − trendKG)'. S3 is exact kg equality; S6 is the only non-kg case (exactly one pound)."
    disposition: accepted
    disposition_rationale: "Round to one decimal then pick the sign. Rounded zero is 0.0 (never +0.0/−0.0). Stone uses the same one decimal; small kg residuals may show 0.0. Add a tiny-residual → 0.0 scenario."
  - id: O4
    category: specification quality
    severity: high
    claim: "FR5's three colour roles have no acceptance scenario, and the followed colour-schemes spec still treats unknown keys as ignored, so custom delta colours can be dropped while every S1–S10 still passes."
    evidence: "FR5: '[colors] delta-pos / delta-neg / delta-zero fail-open; defaults red / green / white.' S1–S10 never assert a custom role is drawn. Followed 2026-09-09-color-schemes.md FR2: 'The supported role keys are exactly those in Color roles. Unknown keys are ignored.'"
    disposition: accepted
    disposition_rationale: "This spec amends the colour-schemes role list: delta-pos, delta-neg, delta-zero are known keys. Add: custom delta-pos is actually painted; invalid value is dropped and the log still opens."
  - id: O5
    category: scope
    severity: medium
    claim: "The spec adds three [colors] keys and is silent on user-facing documentation, so the published role list will be wrong on the day the column ships."
    evidence: "Out of scope names charts, PDFs, editing, and extra fields; it does not mention docs. Decision 5 / FR5 add delta-pos, delta-neg, delta-zero. Followed 2026-09-09-color-schemes.md: 'README and ONBOARDING must describe the [colors] keys'."
    disposition: accepted
    disposition_rationale: "Spec Documentation: docs/how-to/colors.md, docs/reference/config.md, and the colour explanation page list the three keys."
  - id: O6
    category: specification quality
    severity: medium
    claim: "Decision 5 says selection reverse covers the whole row, which contradicts the followed colour-schemes rule that the month sheet restyles only the focused cell."
    evidence: "Decision 5: 'Selection reverse still covers the whole row.' Followed 2026-09-09-color-schemes.md Decision 12: 'Month: the same restyle applies only to the focused cell; other cells on that day keep their role colors.'"
    disposition: accepted
    disposition_rationale: "Whole-row reverse is list-only. On the month sheet, delta keeps its role colour (it is never the focused cell), matching colour-schemes Decision 12."
  - id: O7
    category: alternatives
    severity: medium
    claim: "Default delta-pos is the same red as default trend, so the two adjacent cells are the same colour on the days the column is meant to flag."
    evidence: "Decision 5: 'Built-in defaults: positive red, negative green, zero white.' Followed 2026-09-09-color-schemes.md built-in default: trend is red."
    disposition: accepted
    disposition_rationale: "Default delta-pos is yellow (not red, which is trend). Keep delta-neg green, delta-zero white."
---

## O1 — specification quality — critical

### Claim

S5 requires the delta cell to be empty on the first weigh-in (“no
trend”). Decision 1 requires delta whenever the trend cell already has
an `ApplyTrend` value. Those two rules describe different products, and
the second matches the actual trend start rule.

### Evidence

S5 Given: a day with a weight and no trend (first weigh-in, no prior);
Then the delta cell is empty. Decision 1: both operands must be present,
including on the list the same `ApplyTrend` value already used for the
trend cell. `ApplyTrend` makes the first logged weight the first trend.

### Why this matters

An implementer who follows Decision 1 shows `0.0` next to the first
trend and S5 fails. An implementer who follows S5 blanks a cell whose
two neighbours already display numbers. TDD from S5 first will lock in
the blank. Until S5 or Decision 1 is withdrawn, the pipeline should not
proceed.

## O2 — premise — high

### Claim

The stated problem is not having to subtract the two numbers on the row.
The spec solves a different problem: convert `(weightKG − trendKG)` with
`FromKG`. In `lb` and `st` those are not the same number.

### Evidence

US1: the difference between that day's weight and its trend so I do not
subtract in my head. Decision 1: delta is computed in kilograms then
`FromKG`. S6 only checks a kilogram delta of exactly one pound.

A counter-example: weight 80.0 kg displays as 176.4 lb; trend 79.9 kg as
176.1 lb; cells subtract to 0.3; the spec formula yields 0.1 kg → 0.2 lb
→ `+0.2`.

### Why this matters

US1 is verification by eye. If the three cells do not add, the column
does the opposite of “do not subtract in my head.” S1–S3 are kilogram
cases where `FromKG` is identity, so the happy path hides the bug.

## O3 — specification quality — high

### Claim

Decision 3 and FR2 require a signed one-decimal display and `0.0` with
no sign for zero, but they never define zero after rounding, when the
sign is taken, or what one decimal means in stone.

### Evidence

S3 is exact kilogram equality. S6 is a delta constructed to be exactly
one pound. `display_unit = "st"` is a valid config value and does not
appear in any scenario.

### Why this matters

`%+.1f` of a tiny residual is `+0.0` or `-0.0`. One stone is ~6.35 kg,
so a typical daily residual becomes `0.0` st. Until rounding, sign, and
`st` are specified, implementers invent the format.

## O4 — specification quality — high

### Claim

FR5 adds configurable delta colour roles, but no scenario requires a
custom value to be painted, and the followed colour-schemes spec still
says unknown keys are ignored. Hardcoded defaults satisfy S1–S10.

### Evidence

S9 only requires no chromatic ANSI under `NO_COLOR`. Nothing is of the
form: given `delta-pos = "yellow"`, `+0.5` is drawn in yellow. The
followed spec’s Color roles table does not include the three new keys.

### Why this matters

If this slice does not add the three names as known roles,
`delta-pos = "yellow"` is ignored. An implementation that never reads
`[colors]` for delta still passes S1–S10.

## O5 — scope — medium

### Claim

The spec adds three user-visible `[colors]` keys and does not require
any documentation update. The followed colour-schemes spec already made
README and ONBOARDING describe the role list.

### Evidence

Out of scope names charts, PDFs, editing, and extra fields; it does not
mention docs. There is no Documentation section. User-facing pages that
enumerate roles are outside this spec’s in/out lists.

### Why this matters

Spec-first TDD can go green without touching docs. A person copying a
theme into `[colors]` has no documented keys for the new column.

## O6 — specification quality — medium

### Claim

Decision 5 requires selection reverse to cover the whole row. The
followed colour-schemes spec requires the month sheet to reverse only
the focused cell. Delta is never the focused cell.

### Evidence

Colour-schemes Decision 12: list reverse wraps the entire selected row;
month restyle applies only to the focused cell. This spec Decision 4:
delta is not an editable month-sheet column.

### Why this matters

A literal Decision 5 extends whole-row reverse to the month sheet,
breaking colour-schemes S21. A colour-schemes reading leaves delta in
role colour while weight is reverse-selected. The sentence needs to be
list-only, or the month composition needs its own rule.

## O7 — alternatives — medium

### Claim

The built-in default for a positive delta is `red`, which is already the
built-in default for trend. On days the column exists to flag, the two
adjacent cells are the same colour.

### Evidence

Decision 5: positive `red`, negative `green`, zero `white`. Colour-schemes
built-in: trend is `red`, weight is `blue`.

### Why this matters

US2 already puts `+` in the text, so this is not unreadability. First-run
and omitted-role paths paint delta identically to its neighbour. Config
can override later; the default scheme cannot.

## Explicitly not objecting to

- **Charts and PDFs out of scope**: `idea.md` asks for the column on the
  daily list and month sheet only; the monthly chart already encodes the
  residual as floats-and-sinkers.
- **Not storing delta**: the same reason trend is not a SQLite column.
- **Red = above trend, green = below**: reducing-diet polarity; US2 puts
  the sign in the text.
- **Blank cell rather than `—`**: Decision 2 considered the missing-weight
  glyph and rejected it; US3 only requires not printing a fake zero.
- **Header `delta` immediately right of `trend`**: that is what `idea.md`
  asked for.
- **Shared `layout.go` widths**: Decision 4 plus S10 is enough without a
  separate month alignment scenario.
- **Tab skips delta / not editable**: matches trend; S8 is the right check.
- **Fail-open invalid colours**: inherited from colour-schemes.
- **Non-zero text format (`+0.5` / `-0.5`)**: the right answer to
  colour-must-not-be-the-only-sign; the objection is rounding/zero/`st`.
