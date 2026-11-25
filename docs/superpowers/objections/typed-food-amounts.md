---
spec: docs/superpowers/specs/2026-11-24-typed-food-amounts.md
date: 2026-11-24
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: >
      Tab rebuild drops a typed density and never says
      whether the amount draft or the four nutrient
      drafts survive, so S23 still passes if a cup
      typed for one food is saved against another
      food's density.
    evidence: >
      "The density field is dropped and its typed
      value is discarded. That food's density stays
      as it was." S23 then says "When the amount is
      `40` and Enter is pressed" and, on the way back
      to a new name, "When the four numbers are `1`".
    disposition: accepted
    disposition_rationale: >
      Tab rebuild clears the four nutrient drafts and the
      density draft. The amount draft survives. A volume in
      that surviving amount uses the density of the food the
      form was rebuilt for. Onto stored `Oats` at density
      `1`, a `1 cup` typed before Tab stores 240 g, not
      247.2, and `Oats` stays density `1`. Onto a new name,
      calories, protein, carbohydrate, and fat are empty,
      the density field is empty, and the surviving amount
      is still there. S23 looks at those fields before it
      types over them.
  - id: O2
    category: specification quality
    severity: high
    claim: >
      One Enter checks calories through amount even
      on a form that does not show those fields, and
      no scenario saves a line edit or a food edit
      from a cold screen whose hidden inputs are
      empty.
    evidence: >
      "The order is calories, protein, carbohydrate,
      fat, density, then amount." "`e` edits the four
      numbers and the density. It does not edit the
      name and it does not show the amount. The line
      edit is the amount field only."
    disposition: accepted
    disposition_rationale: >
      Decision 13 walks only the fields on the form. A line
      edit checks the amount and does not check calories,
      the macros, or the density. `e` checks calories,
      protein, carbohydrate, fat, then density, and does
      not check the amount. A hidden field is not a failure
      and does not take focus. A cold line edit, opened
      from the idle screen, fills the amount with the
      stored grams and saves that text. A cold `e` fills
      the four numbers and the density from the stored food
      and saves them. Empty hidden widgets do not block
      either save. S25's order stays the new-food add form,
      where every named field is on screen.
  - id: O3
    category: specification quality
    severity: medium
    claim: >
      Volume grams use the density's reopen spelling,
      but every volume scenario types a density that
      is already that spelling, so the creating Enter
      can multiply the raw decimal and not match the
      next cup.
    evidence: >
      "times the density's reopen spelling. That
      ratio converts to float64 once." The volume
      scenarios type `1.03`, `1`, `+1.03`, and `.5`.
      "+1.03" "reopens as `1.03`" is a density edit
      with no cup on that Enter.
    disposition: accepted
    disposition_rationale: >
      On the Enter that first stores a density, the ratio
      uses the reopen spelling of the density being saved,
      not the raw characters. `0.10000000000000001` stores
      density `0.1`. A cup on that same Enter stores 24 g
      and reopens as `24`. A later cup reopens as `24`.
      `+1.03` with `1 cup` on the same Enter stores 247.2 g,
      the same as density `1.03`. The raw decimal times the
      factor is not the stored grams when it differs from
      that spelling.
  - id: O4
    category: specification quality
    severity: medium
    claim: >
      A refused amount clears when the person tabs or
      types, and a refused density has no such rule
      and no scenario.
    evidence: >
      "Tab or the next typed character clears the
      error and does not save." FR9: "This holds for
      both amount errors." Density is specified as
      "density: want a number greater than zero" and
      S10 only leaves that error by Esc.
    disposition: accepted
    disposition_rationale: >
      A refused density clears on Tab or the next typed
      character and does not save, the same way a refused
      amount does. A weight-log reload does not clear it
      and does not close the form. Esc still cancels and
      writes nothing. One scenario shows density `0`, then
      Tab, then a typed correction, then a reload that
      keeps `density: want a number greater than zero`.
---

# Typed food amounts

## O1 — specification quality — high

### Claim

Tab rebuild drops a typed density and never says whether
the amount draft or the four nutrient drafts survive, so
S23 still passes if a cup typed for one food is saved
against another food's density.

### Evidence

Decision 8:

> Tab rebuilds the add form from the trimmed name. Onto
> a stored food, the form asks only for the amount. The
> density field is dropped and its typed value is
> discarded. That food's density stays as it was. Onto a
> new name, the form asks for the four per-100 g
> numbers, then an empty density, then the amount.

S23 types `1.03` in the density field, changes the name
to stored `Oats`, and after Tab "the amount is `40`".
On the way back to new `Rye`, Tab shows "an empty
`density g/ml`", and the scenario then sets "the four
numbers" to `1` and the amount to `100`. Nothing says
the amount widget is cleared, kept, or replaced, and
nothing says the four nutrient widgets are empty rather
than leftover text.

### Why this matters

A person can type density `1.03` and amount `1 cup` for
a new name, change the name to a stored food, and press
Tab. The spec requires the typed density to be thrown
away. If the cup stays in the amount field, Enter logs
that cup at the stored food's density. `Oats` at density
`1` stores 240 g. The discarded density would have
stored 247.2 g. S23 never sees that cup, because it
types `40` after the rebuild. The same hole runs the
other way. Tab onto `Rye` requires an empty density and
does not require empty calories. Leftover per-100 g
numbers can be stored as the new food. Setting them to
`1` in the scenario hides that. Two implementations
pass S23 and write different grams.

## O2 — specification quality — high

### Claim

One Enter checks calories through amount even on a form
that does not show those fields, and no scenario saves a
line edit or a food edit from a cold screen whose
hidden inputs are empty.

### Evidence

Decision 13:

> One Enter shows the first failure and leaves the
> focus there. The order is calories, protein,
> carbohydrate, fat, density, then amount.

Decision 8:

> `e` edits the four numbers and the density. It does
> not edit the name and it does not show the amount.
> The line edit is the amount field only.

S25's order is only "the add form has been built for a
new name", where every field in that list is on screen.
S6, S8, S10, and S12 do open the line edit or `e`, but
each Given is either a food just saved through the add
form or a food whose widgets the scenario never
describes.

### Why this matters

The ordinary path is to open the meal screen and press
Enter or `e` on a line that is already stored. The
amount widget is empty until the line edit fills it.
`e` prefills the four numbers and the density. It does
not fill the amount. A validator that always walks
Decision 13 hits an empty amount on `e`, or an empty
calories field on a line edit, and focuses a field the
form does not show. The person cannot correct it. A
harness that adds the food through the form in the same
process leaves valid text in those widgets, so S6 and
S10 still pass. The spec never says to skip a field
that is not on the form, and it never says those
widgets are empty. Both readings satisfy the scenarios
that are written.

## O3 — specification quality — medium

### Claim

Volume grams use the density's reopen spelling, but
every volume scenario types a density that is already
that spelling, so the creating Enter can multiply the
raw decimal and not match the next cup.

### Evidence

Decision 4:

> The stored grams are one exact ratio: the parsed
> number, times that factor, times the density's
> reopen spelling. That ratio converts to float64
> once. It is not the product of two float64 values.

S4, S9, and S12 type `1.03` or `1`. S10 saves `+1.03`
and `.5` and checks the density field, not a cup on
that Enter. Decision 9 says `+1.03` "reopens as
`1.03`". The grammar that accepts those values also
accepts a longer decimal with one dot.

### Why this matters

"Reopen spelling" is defined for a number that is
already stored. On the creating Enter the density is
only what was typed. Parsing `0.10000000000000001`
yields the float whose spelling is `0.1`. The exact
decimal times 240, converted once, is
`24.000000000000004`. The spelling `0.1` times 240 is
`24`. A later cup uses the stored spelling and stores
`24`. The first line does not. Both paint `24.0` at
one decimal. Enter is how they differ. No scenario
types a density that is not already its reopen
spelling and then types a volume, so both ratios pass
S4 and S9. Decision 4's rule that one cup and `240 ml`
of the same food store the same grams is then false
across those two saves.

## O4 — specification quality — medium

### Claim

A refused amount clears when the person tabs or types,
and a refused density has no such rule and no scenario.

### Evidence

Decision 5:

> Tab or the next typed character clears the error and
> does not save.

FR9: "This holds for both amount errors." S17 covers
both amount strings, including Tab, the next
character, and a weight-log reload. Density's rule is
the message "density: want a number greater than zero"
and, in S10, Esc. S10 never tabs, never types a
correction, and never reloads while that error is up.

### Why this matters

The amount rule exists so a refusal does not stick
across the next key and does not vanish on reload.
The density field is a new error on the same form.
Nothing says Tab or the next digit clears it. An
implementation can leave `density: want a number
greater than zero` on screen after the person has
replaced `0` with `1.03`, and can clear it only on
Esc. S10 still passes, because Esc is the only exit
it uses. A weight-log reload of a meal error is
already covered by meal-day Decision 23. This hole is
the clear on the next key, which this spec states
only for the two amount errors.

## Explicitly not objecting to

- **Ounce factor, half-even tenths, and the 28.3
  alias:** Decision 3, S2, and S21 pin the stored
  product, the paint, and the calorie split, and
  those strings match one `Float64` plus `%.1f`.
- **One-conversion spellings that the revision
  added:** `247.2`, `0.3333333333333333`,
  `18446744073709552000`, and the tiny fraction's
  spelling are what `FormatFloat` with `'f'` and
  precision `-1` actually returns. S22 and S24 are
  not wrong oracles.
- **Smallest-density `1/2 ml` versus `1/4 ml`:** Half
  of that magnitude's reopen spelling converts to the
  same magnitude. A quarter converts to zero and is
  the greater-than-zero refusal. That is the check
  the revision added, not a missing one.
- **Stale meal-day sentences already named:** The
  preamble retires Decision 16, S24, S25, S31, the
  `grams` label, the old error string, and the
  optional-macros sentence. Re-listing meal-day's
  out-of-scope density bullet would repeat that
  objection. S9 requires the field anyway.
- **Mixed numbers and the customary cup:** `1 1/2` is
  refused in S6, and one cup is 240 ml with no
  alternate. Both are choices with a specified
  result, not an untested failure.
- **Grammar before density for `0 cup`:** Decision 5
  states it. S6's food happens to have a density, but
  the sentence is not two readings.
- **`e` and `x` while a number field is open:**
  Meal-day Decision 19 still says they do not run
  their idle actions. No accepted token needs either
  letter.
- **A kitchen-fraction list or a maximum gram
  weight:** Decision 12 refuses both limits, and S24
  is the behavior that choice produces.
- **Tab onto a new name starts from an empty
  density:** Decision 8 says the field is empty, and
  the person can see that it was cleared.
- **Add-form first-failure order:** S25 pins blank
  calories over a blank protein, density `0`, and a
  bad amount, and density `0` over `1 cup`.
