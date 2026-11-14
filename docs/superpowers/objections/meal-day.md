---
spec: docs/superpowers/specs/2026-10-30-meal-day.md
date: 2026-10-30
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "A weight-log save still in flight sets the screen, the list cursor, and the list error, and the plan limits app.go to the f key, the screen switch, and the view branch, so the meal screen inherits those handlers."
    evidence: "FR2 and Decision 2: Esc from the idle meal screen returns to the list and leaves the selected row. Plan, app.go: \"Only the list f case, the screen switch, and the view branch, each a call into meal.go.\" Plan tests: \"press already sends esc, tab, enter, [, ], and a space.\""
    disposition: accepted
    disposition_rationale: >
      A weight-log save that finishes while the meal screen is up
      does not change that screen and does not move the list row
      Esc will restore. The save still reloads the log. A form or
      month error that arrives on the meal screen is kept on that
      form or sheet, so the list does not become the failed-log
      view. A meal error is not stored where a log reload clears
      it. The plan may edit those handlers. One test leaves a
      day-form save unfinished, presses Esc, presses f, and then
      delivers the save. One does the same for a month-cell save.
      One delivers a form parse error while meals are open and
      then leaves to the list.
  - id: O2
    category: specification quality
    severity: high
    claim: "Stored food numbers stay unrounded, but only the grams editor must survive an unchanged Enter, so a one-decimal food field rewrites every day that uses the food."
    evidence: "Decision 10: \"Stored grams and stored food numbers are not rounded to one decimal.\" The unchanged-Enter sentence is only \"Opening grams again shows the stored grams, and accepting that text unchanged keeps them.\" Decision 18: other dates follow a successful edit, and a refused number does not replace the stored numbers. Plan: the nutrient editor uses strconv.FormatFloat precision -1, but TestMealDisplayOneDecimal round-trips grams of 1.25, and S32 changes calories 200 to 100."
    disposition: accepted
    disposition_rationale: >
      The food editor shows the stored numbers at full precision,
      the same way the grams editor does. Enter on an unchanged
      field keeps them. A blank nutrient on an edit refuses the
      whole edit, writes nothing, and does not store zero. A
      fractional nutrient is allowed. A test stores 1.25, opens
      the food, presses Enter, and the totals stay.
  - id: O3
    category: specification quality
    severity: high
    claim: "Tab is what switches the add form between a known food and a new one, and Enter saves from any field, but nothing says which food is written when the name and the fields disagree."
    evidence: "Decision 15: \"If the trimmed name matches a stored food, the only other field is grams.\" \"Changing the name and pressing Tab follows the new name. Tab does not write. Enter saves from any field.\" S17 asks only for grams after Tab. Plan: \"A match reuses that food and does not write its four numbers. A miss inserts the food and the line in one save.\""
    disposition: accepted
    disposition_rationale: >
      Enter saves only when the fields match the trimmed name. If
      the name changed and Tab has not been pressed since, Enter
      refuses, the form stays open, nothing is written, and the
      screen shows error: and name: press tab. After Tab, a known
      name logs grams against that food and does not write the
      four numbers. A new name inserts the food and the line
      together.
  - id: O4
    category: alternatives
    severity: medium
    claim: "Idle tenths copy the weight column's half-even format, which exists so a delta matches painted cells, but this screen has no such identity and never states that 1.25 paints as 1.2."
    evidence: "Decision 10: shown totals and grams use \"one decimal the way a body weight is shown, including a halfway value.\" S27 checks \"the one-decimal text a body weight uses for 1.25\" and for 0.25, and does not state the characters. Plan: idle figures use \"fmt one decimal, the same call the weight column uses.\""
    disposition: accepted
    disposition_rationale: >
      The idle tenth stays the weight column's one-decimal format.
      S27 states the characters: 1.25 grams paints as 1.2, and
      0.25 calories paints as 0.2. Opening grams shows 1.25 again.
  - id: O5
    category: scope
    severity: medium
    claim: "This slice can store a food it cannot rename or delete, and the later recipe slice only forbids deleting a food that is still an ingredient, so a typo name keeps matching."
    evidence: "Decision 13: the match is case-sensitive and the stored name is the trimmed name. Decision 14: x removes the line and \"The food remains.\" Decision 18: \"The name is shown and cannot be changed from this screen.\" The out-of-scope list does not defer rename or delete. Slice S4: \"a food or recipe that is still an ingredient of another recipe cannot be deleted until it is removed there.\""
    disposition: deferred
    disposition_rationale: >
      This slice does not rename or delete a food. idea.md records
      that those are later work, and that the recipe slice's rule
      about an ingredient is not that operation. The meal-day
      out-of-scope list says the same. No new key is added here.
  - id: O6
    category: risk
    severity: medium
    claim: "Enter on a food edit replaces the four numbers on every date, the idle line never showed them, and the help and how-to requirements never say that other days change."
    evidence: "Decision 18: \"Other lines and other dates that use the food follow a successful edit.\" Decision 9: the line does not show that line's own four figures. Decision 20's idle help is \"e food\". The how-to instructions say \"edit food\" and do not require saying that other days change. S32 checks the new totals, not that the old numbers can be read back."
    disposition: accepted
    disposition_rationale: >
      The e row in the key reference and the meal how-to say that
      saving changes every meal date that uses the food. The edit
      shows the stored numbers before Enter. There is no confirm.
      The idle line still shows only the name and the grams.
  - id: O7
    category: specification quality
    severity: medium
    claim: "Grams have a closed grammar, but a nutrient is only a non-negative number, so the same text is a stored value or a refusal depending on who parses it."
    evidence: "Decision 16 and S24 refuse a gram exponent, slash fraction, and unit token, and name 1e2. Decision 17: a missing nutrient or \"a non-number\" is refused, and a negative is refused, without that grammar. Plan: \"Nutrients use the same trim. Empty after trim is missing, not zero. Negative is refused. Zero is allowed.\" It does not apply the gram grammar."
    disposition: accepted
    disposition_rationale: >
      A nutrient uses the gram grammar, and zero is allowed. A
      leading + and one decimal point are allowed. An exponent,
      inf, NaN, a slash, and a unit token are refused on add and
      on edit. Empty is missing, not zero. A negative is refused.
---

## O1 — implementation — high

### Claim

A weight-log save still in flight sets the screen, the list
cursor, and the list error, and the plan limits `app.go` to
the `f` key, the screen switch, and the view branch, so the
meal screen inherits those handlers.

### Evidence

FR2 and Decision 2: Esc from the idle meal screen returns to
the list and leaves the selected row where it was. S3 checks
that row only when nothing else is in flight.

The plan's `app.go` row is:

> Only the list `f` case, the screen switch, and the view
> branch, each a call into `meal.go`.

The plan's screen tests send keys with `press`, which
"already sends `esc`, `tab`, `enter`, `[`, `]`, and a
space." No meal scenario leaves a save unfinished.

### Why this matters

The day form does not wait for its write. Enter returns
`saveForm` as a command and stays on the form. Esc sets the
screen back to the list while that command is still running.
`f` then opens meals. `savedMsg` does not look at which
screen is up. It sets `a.screen` to `afterSave`, may set
`selectDay` to the day the form wrote, and returns `load`.
The meal screen is gone.

`n` opens that form on today, not on the selected row. The
following `loadedMsg` runs `placeCursor`. The row Esc was
supposed to restore is no longer that row. A month-sheet
Tab save is the same shape with one more Esc. Esc ends the
edit, Esc leaves the sheet, `f` opens meals, and `savedMsg`
puts the screen back on the month sheet. Tab's save does
not watch the cancel count Esc bumps. `armLogSave` already
says the next key is not ordered behind the command.

A failed form parse is `loadErrMsg`. That handler stores
the text on the day form or the month sheet, and otherwise
on `a.err`. The meal screen is neither. `a.err` is the list
failure view: the log is replaced by the error and `q quit`,
which is Decision 20's failed list. A typo on the day form,
then Esc and `f`, turns the next list into that failure.
The same `loadedMsg` clears `a.err`. A meal error kept
there disappears when a log reload lands.

`press` calls `Update` and drops the command it returns.
Every meal scenario in the plan is a `press` sequence.
None of them leaves a weight-log save in flight, so FR2
stays green. An implementer who obeys the `app.go` limit
cannot exempt the meal screen.

## O2 — specification quality — high

### Claim

Stored food numbers stay unrounded, but only the grams
editor must survive an unchanged Enter, so a one-decimal
food field rewrites every day that uses the food.

### Evidence

Decision 10:

> Stored grams and stored food numbers are not rounded to
> one decimal.

The round-trip exception in that decision is only for
grams:

> Opening grams again shows the stored grams, and
> accepting that text unchanged keeps them.

Decision 18 says `e` edits the four numbers, Enter
accepts, and other lines and other dates follow a
successful edit. It does not say the editor shows the
stored numbers, or that accepting them unchanged keeps
them. A refused number does not replace what is stored.
Decision 17's missing-number rule is headed "A new food
needs all four numbers," not "an edit."

The plan says the nutrient editor uses
`strconv.FormatFloat` with precision `-1`, so an unchanged
Enter stores the same number. `TestMealDisplayOneDecimal`
shows `1.25` in the grams editor and keeps `calories 2.5`.
S32 changes calories from 200 to 100. Every nutrient in
the plan's tests is an integer or zero.

### Why this matters

Add can store `1.25`. The person types it, and Decision 10
forbids rounding it on the way in. No scenario does that
for a nutrient, but nothing refuses a fractional one.
Shown totals and shown grams are one decimal. The only
place that says the editor must not use that tenth is the
grams sentence.

A food form that paints the idle tenth, then parses it on
Enter, turns stored `1.25` into `1.2` or `1.3` and writes
that back. Decision 18 then updates every date that uses
the food. Gram weights stay. The totals move, and the old
number is not on the idle line to copy. S32 still passes,
because 200 and 100 are unchanged at one decimal. The
plan's `-1` note is the fix, and no named test fails if
the note is skipped.

A blank nutrient is the same editor, and it is also
unspecified. On a new food, empty is a refusal and is not
stored as zero. On `e`, that heading does not apply. A
blank can refuse the whole edit, leave that one number,
or store zero. S32 never clears a field, so all three
stay green. Zero would also paint as `0.0` on every date
that uses the food.

## O3 — specification quality — high

### Claim

Tab is what switches the add form between a known food
and a new one, and Enter saves from any field, but nothing
says which food is written when the name and the fields
disagree.

### Evidence

Decision 15:

> If the trimmed name matches a stored food, the only
> other field is grams. If it does not, the fields are
> calories, protein, carbohydrate, and fat, each labeled
> per 100 g, then grams. Changing the name and pressing
> Tab follows the new name. Tab does not write. Enter
> saves from any field.

S16 walks Tab for a name that is not stored. S17 asks
only for grams after Tab on a name that is. No scenario
changes the name after those fields are chosen and then
presses Enter.

The plan resolves the save, not the fields on screen:

> A match reuses that food and does not write its four
> numbers. A miss inserts the food and the line in one
> save.

### Why this matters

The usual correction is to fix the spelling and press
Enter. The person types a new name, Tabs through the four
numbers and grams, Tabs back to the name, changes it to a
food that already exists, and presses Enter. They do not
press Tab again. Decision 15 updates the fields on Tab,
not on Enter.

Two readings both pass S16 and S17.

-   Save-time match, as the plan describes it, ignores
    the numbers just typed and logs grams against the
    existing food. There is no error. The day's totals
    are the old food, not the figures on the form.
-   The field set stays what the last Tab chose. Enter
    then tries to insert a second food under a name that
    Decision 13 says is unique. The spec has no text for
    that failure. The other resolution is to overwrite
    the existing food, which updates every day, the
    opposite of "does not write its four numbers."

The reverse is the same hole. The form is on grams for a
known food. The person changes the name to a new one and
presses Enter without Tab. The plan's miss has no four
numbers to store. The sticky grams form can insert a food
with none, which Decision 17 forbids, or with zeros,
which that decision also forbids. S19 only checks Esc.
S22 checks a missing nutrient when the new-food fields
are already showing.

## O4 — alternatives — medium

### Claim

Idle tenths copy the weight column's half-even format,
which exists so a delta matches painted cells, but this
screen has no such identity and never states that 1.25
paints as 1.2.

### Evidence

Decision 10: shown totals and shown grams use one decimal
"the way a body weight is shown, including a halfway
value." An exact zero is `0.0`, never `-0.0`.

S27:

> the gram text is the one-decimal text a body weight
> uses for 1.25

and the calorie text a body weight uses for 0.25. It
does not say what those strings are. The plan says the
idle screen prints them with `fmt` one decimal, "the
same call the weight column uses," and the test asserts
that call rather than the characters.

### Why this matters

The weight column's one decimal is half-even, not
half away from zero. `delta.go` says `%.1f` is
half-even, not `math.Round`, so the painted weight,
trend, and delta add. The list test for 80.05 kg
expects `80.0` from `%.1f` and not `80.1` from
`math.Round`.

`1.25` and `0.25` are exact halfway cases. Half-even
keeps the even tenth: `1.25` is halfway between `1.2`
and `1.3`, and `0.25` is halfway between `0.2` and
`0.3`, so the idle texts are `1.2` and `0.2`.
`math.Round` yields `1.3` and `0.3`. S27 chose those
two inputs. A plan follower paints `1.2` grams and
`calories 0.2`, and a test written as "fmt one
decimal" stays green. The person who typed `1.25`
sees `1.2` on the line and `1.25` again only if they
open grams.

That tie break is what makes three weight cells add.
A meal line does not paint a contribution beside the
total, and Decision 9 forbids putting one there. The
spec never weighs a halfway rule that rounds 5 away
from zero, and it never writes `1.2` or `0.2`, so the
acceptance text does not show the tenth this screen
will actually use.

## O5 — scope — medium

### Claim

This slice can store a food it cannot rename or delete,
and the later recipe slice only forbids deleting a food
that is still an ingredient, so a typo name keeps
matching.

### Evidence

Decision 13: a name is unique after the ends are
trimmed, the match is case-sensitive, and the stored
name is the trimmed name. A middle space is part of
the name. S20: `oats` is a different food from `Oats`.

Decision 14: `x` removes the selected line
immediately, and the food remains. The next add of
that name asks only for grams. Decision 18: from `e`,
the name is shown and is not a field.

The out-of-scope list defers allowance, volume,
recipes, and the starter set. It does not mention
rename or delete. The slice this spec implements
points at S4 for deletion, and S4 only says:

> a food or recipe that is still an ingredient of
> another recipe cannot be deleted until it is removed
> there.

### Why this matters

The first saved line creates a permanent match key.
Removing every line that uses it does not remove the
key. The next add of that spelling skips the four
numbers and keeps whatever was stored, including a
typo in the calories. `Oats` and `oats` are two of
those keys. Neither can be merged, because the name
cannot be edited and the line does not carry its own
copy of the numbers.

S4 does not add the missing operation. It adds a case
where deletion is refused, and it still does not say
how an unused food is deleted. After both slices, a
name this screen accepted still matches, and a name
that became an ingredient is the one the record
explicitly will not remove. The acceptance scenarios
never try to undo a food. They only prove the line
can go and the food stays, which is the trap.

## O6 — risk — medium

### Claim

Enter on a food edit replaces the four numbers on every
date, the idle line never showed them, and the help and
how-to requirements never say that other days change.

### Evidence

Decision 18: a successful `e` is followed by every
other line and every other meal date that uses the
food, and the gram weights do not change. Esc keeps
the old numbers only if it is pressed before Enter.

Decision 9: the idle line shows the name and the gram
weight, not that line's own calorie, protein,
carbohydrate, or fat. Decision 20's idle help is
`e food`. S41's key row is "`e` edits the food." The
how-to instructions list "edit food" among the steps
for this screen and require the per-100 g basis.
They do not require a sentence that other days
change. `TestMealDocs` checks that later slices are
named, not that this consequence is.

### Why this matters

Before Enter, the four numbers are visible only
inside `e`. After Enter, they are gone. With one
food on the day, the old per-100 g figure can be
backed out of the old total and the grams, but the
old total is no longer on screen. With two foods,
the total is a sum, so it cannot be backed out at
all. S32 uses one food and checks the new totals.
It does not keep the previous calories anywhere.

`x` says there is no confirm. `e` does not say
that, and it also does not say how many dates will
move. The help and the key row can be read as
editing the food on this line. That reading matches
the words on screen. It does not match Decision 18.
A typo in one field, then Enter, is the misuse. The
screen gives the person the new totals for the date
they are looking at, and no copy of the numbers
they just replaced.

## O7 — specification quality — medium

### Claim

Grams have a closed grammar, but a nutrient is only a
non-negative number, so the same text is a stored value
or a refusal depending on who parses it.

### Evidence

Decision 16: after trim, grams allow a leading `+` and
one decimal point, and refuse zero, a negative, a
non-number, a slash fraction, and a unit token. S24
includes `1e2` in the gram refusals. FR11 repeats the
exponent.

Decision 17: a nutrient is trimmed; a missing field or
a non-number is refused; a negative is refused; zero
is allowed. It does not mention `+`, an exponent, a
slash, or a non-finite value.

The plan gives grams a grammar and then `ParseFloat`,
and names `1e2` as refused. For nutrients it says the
same trim, empty is missing, negative is refused, and
zero is allowed. It does not say the gram grammar
applies.

### Why this matters

`ParseFloat` accepts `1e2`, `+10`, `inf`, and `NaN`.
It rejects `1/4`. A negative check does not catch
`inf` or `NaN`, because neither compares less than
zero. The gram scanner rejects all of those before
`ParseFloat`.

One implementer reuses the gram scanner and refuses
`1e2` and `inf` with `want a number`. Another uses
`ParseFloat`, stores `1e2` as 100, and stores `inf`
as a non-finite calorie. Decision 18 then puts that
value on every date that uses the food. `fmt` one
decimal of a non-finite total is not `0.0` and not
a tenth. S22 and S23 only cover an empty field and
`-1`, so both parsers pass.

The live food makes the looser parser sticky. A
later edit that round-trips the stored value, which
is what the plan's precision `-1` is for, will keep
`+Inf` or `NaN` and keep painting it. Grams cannot
get there. Nutrients can, and the spec never says
whether `1e2` is a hundred or a refusal.

## Explicitly not objecting to

- **`f` opens today, not the selected row**: That is
  the accepted S1 rule, and Decision 1, FR1, and the
  how-to all repeat it. Losing the meal date on the
  next `f` is the cost of that rule.
- **Live food numbers rather than a copy on the line**:
  The slice disposition already chose this. O2 and O6
  are the round-trip and the missing recovery, not a
  second vote on sharing the food.
- **Refusing units and slash fractions in this slice**:
  S3 owns ounces, pounds, and volume. Decision 16 and
  S25 list the tokens this screen must refuse. Pulling
  the cup factor forward is what the slicing record
  kept out.
- **Per 100 g**: The S1 rationale accepted that basis
  so a gram line has a scale. A declared serving would
  be a different slice.
- **Summing the stored products, then printing one
  tenth**: The plan already sums float64 from the
  stored grams, not from the idle text. The line does
  not show a per-line calorie the total would have to
  match. The weight delta's round-then-subtract rule
  is for three painted cells this screen does not have.
- **A second SQLite pool, and a meal migration that
  aborts startup**: The plan opens the meal tables on
  their own connection and returns if that open fails.
  Decision 21 is that an existing file still migrates
  and keeps its rows. A lock error on a second insert
  is a failed save, and a food row that already holds
  its four numbers still matches on the retry. The
  scenarios are one key sequence against one small
  file.
- **A positive gram that paints as `0.0`**: Any one
  decimal idle text does that to a value under half a
  tenth. Decision 10 shows the stored grams again when
  the grams editor opens, so the typed value is not
  only the idle tenth.
- **Paste, shift-tab, and a bracket typed into a
  field**: The scenarios do not promise paste or a
  backward key. Tab wraps to the name. S38 only
  requires that `[` and `]` leave the meal date
  alone while a form is open.
