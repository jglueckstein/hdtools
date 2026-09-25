---
spec: docs/superpowers/specs/2026-09-22-delta-column.md
date: 2026-09-25
mode: code
diaboli_model: grok-4.6
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "formatDelta rounds with math.Round while the weight and trend cells paint with fmt %.1f, so the column is not the subtraction of the two numbers on screen."
    evidence: "delta.go: v := round1(round1(dw) - round1(dt)) with math.Round(x*10)/10; app.go/month.go: weight/trend = fmt.Sprintf(\"%.1f\", FromKG(...))"
    disposition: accepted
    disposition_rationale: "Delta must use the same one-decimal as the painted cells: fmt.Sprintf(\"%.1f\", FromKG(...)) then parse and subtract (not math.Round). Add a halfway fixture if %.1f and math.Round disagree, and assert delta equals the painted subtraction."
  - id: O2
    category: risk
    severity: high
    claim: "S5 is locked only for an exact 80.0 kg first weigh-in, so a first lb/st entry can show a signed delta while every named test stays green."
    evidence: "TestListDeltaFirstWeighInIsZero seeds 80.0 kg under config.Default(); ApplyTrend first day is t := round1(*Weight); formatDelta then FromKG's both kg values."
    disposition: accepted
    disposition_rationale: "S5 means when the two painted cells are equal, print 0.0. First weigh-in in kg stays 0.0. Add a first weigh-in in lb (176.5); the cell is the displayed subtraction (may be signed). Do not force day-one 0.0 in lb/st."
  - id: O3
    category: risk
    severity: medium
    claim: "FR5's delta-neg and delta-zero paint and overlay paths have no test, so a swapped or dropped role still passes S11/S12."
    evidence: "styleDelta switches on s[0] == '+' / '-'; TestListDeltaCustomColorPainted and both TestInvalidDeltaColorDropped only load delta-pos."
    disposition: accepted
    disposition_rationale: "Add paint tests for delta-neg and delta-zero (custom colour on -0.5 and 0.0). Invalid delta-neg still opens with default green."
  - id: O4
    category: risk
    severity: medium
    claim: "FR7 is unimplemented-as-a-test: list reverse covering delta and month delta keeping role colour can regress without a red."
    evidence: "listDeltaCell uses visible(app.View()); TestMonthSelectionRestylesFocusedCellOnly asserts trend cyan on the focused row and never reads the delta cell."
    disposition: accepted
    disposition_rationale: "Add tests: list reverse includes the delta cell (SGR 7 on the +0.5 line); month sheet with weight focused, delta keeps role colour (not reverse)."
---

## O1 — implementation — high

### Claim

`formatDelta` does not subtract the two numbers the user can already
see. The list and month paint weight and trend with `fmt.Sprintf("%.1f",
FromKG(...))` (strconv half-to-even). Delta re-converts the kilogram
operands and rounds with `math.Round` (half away from zero), then
subtracts. The file even states the US1 intent — “Round each cell first
so delta matches the two numbers on screen” — and then uses a different
rounder than the cells.

### Evidence

`internal/tui/delta.go`: `round1` is `math.Round(x*10)/10`;
`v := round1(round1(dw) - round1(dt))`.

List and month paint: `fmt.Sprintf("%.1f", FromKG(...))`.

Spec Decision 1: take the same `FromKG` values already shown, subtract
those two display-unit numbers. S6 forbids a residual that is not that
subtraction. S6’s pair is not a halfway case, so
`TestListDeltaDisplayedCellsAddUp` stays green while US1 is false when
the first discarded digit is 5.

### Why this matters

US1 is “so I do not subtract in my head.” A 0.1 disagreement between
`(weight cell) − (trend cell)` and the delta cell is the failure the
column exists to prevent, and it is untested.

- **accept-as-stated** — `math.Round` is the project’s one-decimal
  (same as `ApplyTrend`); write down that painted cells may differ at
  half-even.
- **revise-spec** — Decision 1 means parse the two `%.1f` strings and
  subtract those decimals.
- **add-test** — a fixture where `Sprintf("%.1f", x)` and `round1(x)`
  disagree, and assert delta equals the painted subtraction.
- **consciously-carry** — a 0.1 miss at halfway is acceptable because
  sign-in-text still works.

## O2 — risk — high

### Claim

S5 (“first weigh-in is `0.0`”) is an artefact of the kilogram fixture,
not of the formula. `ApplyTrend` stores the first trend as
`round1(weightKG)`, then `formatDelta` converts that kilogram trend with
`FromKG`. A first weigh-in entered in lb/st can paint two different
one-decimal cells, and the delta column will show their signed
difference. No test uses that path.

### Evidence

`internal/dailylog/trend.go`: first day `t := round1(*Weight)` in kg.

`TestListDeltaFirstWeighInIsZero` seeds 80.0 kg under
`config.Default()`. 80.0 kg is already one decimal, so both cells are
`80.0`. A day-one entry of `176.5` lb stores kg that `ApplyTrend`
rounds, and the sheet can show weight `176.5`, trend `176.6`, delta
`-0.1`.

### Why this matters

Default config is kg, so CI never sees pounds-first day one. Decision 1
(displayed cells) and S5 (first day is `0.0`) cannot both hold once
kilogram `round1` meets `FromKG`. The trend column already contained
this 0.1 disagreement; delta makes it a number the user is told to
trust.

- **accept-as-stated** — S5 is only for kg; write that down.
- **revise-spec** — S5 means “when the two painted cells are equal,
  print `0.0`,” not “the first weigh-in is always `0.0`.”
- **add-test** — first weigh-in `176.5` lb; assert whatever the spec
  then decides (`0.0` or `-0.1`).
- **consciously-carry** — ship the signed day-one residual in lb/st.

## O3 — risk — medium

### Claim

FR5 names three roles with three defaults. The only colour tests load
`delta-pos`. `delta-neg` and `delta-zero` can be omitted from
`colorRoles`, wired to the wrong palette field, or fail-open
incorrectly, and S11/S12 still pass.

### Evidence

`styleDelta` switches on `s[0] == '+' / '-'`. Custom and invalid colour
tests only set `delta-pos`. `TestListDeltaNegative` and
`TestListDeltaZero` assert extractable text only.

### Why this matters

A `case s[0] == '-'` deletion paints negatives as `delta-zero`. Dropping
`RoleDeltaNeg` from `colorRoles` silently ignores a user’s
`delta-neg`. US2 puts the sign in the text, so the bug is wrong colour,
not wrong meaning — but FR5 is the colour contract.

- **accept-as-stated** — S11/S12 are the whole colour contract; pos
  stands for the three keys.
- **revise-spec** — add paint and fail-open scenarios for `delta-neg`
  and `delta-zero`.
- **add-test** — custom colour on `-0.5` and `0.0`; invalid `delta-neg`
  still opens on default green.
- **consciously-carry** — a mispainted negative or zero is acceptable.

## O4 — risk — medium

### Claim

FR7 has two sentences and neither is a test. List reverse covering
delta, and month-sheet delta keeping role colour while another cell is
focused, can both regress with the current suite green.

### Evidence

`listDeltaCell` uses `visible()` (ANSI stripped). Month
`TestMonthSelectionRestylesFocusedCellOnly` asserts focused weight
yellow and trend cyan, never the delta cell. `TestMonthDeltaMatchesList`
checks `+0.5` text only.

### Why this matters

If list reverse stops wrapping delta, or month delta picks up reverse
from a sloppy `onRow` restyle, the selection contract breaks and the
suite cannot see it.

## Explicitly not objecting to

- **Charts, PDFs, editing delta, extra fields, spreadsheet keys,
  high-res charts**: out of this slice; no `colDelta`, no SQLite
  column, no chart paint.
- **Plan `round1(dw − dt)` versus code `round1(round1(dw) − round1(dt))`**:
  the extra operand rounding is what makes S6’s `+0.3`; O1 is the
  remaining rounder mismatch with `%.1f`.
- **Tab skips delta**: `colCount` excludes delta; tests require weight
  → sleep.
- **S4 blank-on-carry**: `TestMonthDeltaBlankWithoutWeight` checks
  carried trend and empty delta.
- **Sign survives `NO_COLOR`**: `TestListDeltaNOCOLORKeepsSign`.
- **Shared `wDelta` and header right of `trend`**: `TestDeltaColumnAligns`.
- **Documentation of the three colour keys**: how-to, config, color
  explanation match spec O5.
- **Default yellow/green/white**: `defaultColors`; S12 locks default
  yellow for pos.
- **Tutorial `first-log.md` not mentioning the new column**: spec
  Documentation named only the three colour pages.
