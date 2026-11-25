# Typed food amounts

**Date**: 2026-11-24
**Status**: draft
**Issue**: not filed
**Backlog**: [`idea.md`](../../../idea.md) (Meal Planning)
**Slice**: S3 of
[`meal-planning.md`](../slices/meal-planning.md)
**Follows**:
[`2026-11-20-meal-allowance.md`](2026-11-20-meal-allowance.md)
(approved, shipped),
[`2026-10-30-meal-day.md`](2026-10-30-meal-day.md)
(approved, shipped), and
[`2026-11-14-optional-food-macros.md`](2026-11-14-optional-food-macros.md).
This file does not amend those specs. The meal screen they
describe is the screen this slice extends. This file still
does not edit them.

For this amount field, meal-day Decision 16, S24, S25, and
S31 are stale, as are the `grams` label and the error
`grams: want a number greater than zero`. The
optional-macros sentence that a stored name asks only for
grams is stale too. Those scenarios are no longer in force
for this field. The error string is
`amount: want a number greater than zero`.

The amount field accepts a typed US amount and stores grams.
The person uses that field on add and on Enter. There is no
new key. A bare number is still grams. A token after the
number selects ounces, pounds, or a volume. Volume needs one
density on the food, in grams per milliliter. A food with no
density still accepts weight and still refuses volume. The
idle line shows the food name and the stored grams to one
decimal. It does not show the token.

The factors are the accepted nutrition-label group. They are
the factors in use, not proposals. The customary cup is not
an alternative.

Volume and density are this slice. Recipes, the starter set,
and renaming or deleting a food stay later.

No production code until these scenarios exist as failing
tests.

## Amount field

The field is the one that used to be labeled `grams`. On add
and on the line edit its label is `amount`. Idle help
replaces `enter grams` with `enter amount`. `a`, Enter, `e`,
`x`, `b`, `[`, `]`, Esc, and `q` stay what they are. `f` on
the daily list remains the only door. The key table does not
gain a key.

## User stories

### US1 — Type a weight as grams, ounces, or pounds

As a person logging a meal, I want to type grams, ounces, or
pounds in the amount field and have the line store grams, so
that I do not convert by hand.

### US2 — Type a volume when the food has a density

As a person logging a liquid, I want a cup, a fluid ounce, a
tablespoon, a teaspoon, or milliliters to become grams
through that food's density, so that the line stays in
grams.

### US3 — Keep a food with no density on weight only

As a person who has not set a density, I want a volume to be
refused and a weight to save, so that a cup does not invent
a gram weight.

### US4 — Change the density later

As a person correcting a food, I want a new density to
affect the next volume only, so that grams already stored
stay put.

### US5 — Leave body weight and the allowance alone

As a person who also keeps a weight log, I want a food
amount to stay grams on this screen, so that ounces are not
body weight and the allowance is not a cup.

## Decisions

1.  **The amount field is the old grams field.** There is
    no new key. The label is `amount` on add and on the
    line edit. The idle help line is:

    ```text
    a add   enter amount   e food   x remove   b allowance   [ ] day   esc list   q quit
    ```

    The add form, the amount edit, and the food edit keep
    their help line:

    ```text
    enter save   esc cancel   tab next field
    ```

    The allowance editor keeps its help line, without Tab.
    Idle `a`, Enter, `e`, `x`, `b`, `[`, `]`, Esc, and `q`
    stay as they are. Idle `f` still does not open another
    screen. `f` on the daily list remains the only door.
2.  **A bare number is grams. A token may follow it.**
    Surrounding space is removed first, the same trim the
    other meal fields use. The number is either a decimal
    or one slash fraction. The decimal shape is the shape
    already accepted for grams: an optional leading `+`,
    then digits with at most one decimal point, and at
    least one digit. `.5`, `+100`, `100`, `100.0`, and
    `1.` are decimals. A fraction is two positive integers
    against one slash, such as `1/4` or `3/4`, with no
    sign, no decimal point, and no space around the slash.
    Each side fits in 64 bits and is at least 1. `2/4` is
    allowed. It does not have to be in lowest terms. A
    leading `+` on a fraction is refused. After the
    number, the remainder may be empty. If it is not, the
    whitespace at the start of the remainder is removed
    and what is left must be exactly one token. That gap
    may be more than one space. Whitespace inside the
    token is not removed. The tokens, compared without
    regard to case, are `g`, `oz`, `lb`, `cup`, `fl oz`,
    `tbsp`, `tsp`, and `ml`. `oz` is weight. `fl oz` is
    volume. The token `fl oz` is the two words with one
    single space between `fl` and `oz`. That space is the
    ordinary space character. `1 fl oz` and `1  fl oz`
    work. `1 fl  oz` is refused. `1g` and `1 g` both
    work. A bare `c` is not a cup. Plurals and full words
    are not tokens, including `cups`, `ounces`, `pounds`,
    `tablespoons`, `teaspoons`, `milliliters`, `floz`,
    `gram`, `grams`, `ounce`, `pound`, `stone`, `st`,
    `kg`, and `kilogram`. A token before the number, a
    second token, a mixed number such as `1 1/2`, more
    than one slash, and a zero denominator are refused.
3.  **Weight uses the accepted factors, and the line's
    tenth is not the factor check.** One ounce is
    28.349523125 g. One pound is 16 of those ounces, so
    one pound stores 453.59237 g. `16 oz` stores that
    same number. The stored value is the full product,
    not a kitchen rounding. The idle line paints stored
    grams to one decimal with the same half-even rule as
    the calorie line and the weight column. `1.25` paints
    `1.2`. `0.25` paints `0.2`. `28.25` paints `28.2`.
    `28.349523125` is below the halfway point, so one
    ounce paints `28.3`. `28.35` is halfway and the
    tenths digit is odd, so it paints `28.4`. The
    slicing record said one decimal does not distinguish
    `28.349523125` from `28.35`. That is not how this
    print behaves. A painted `28.3` is not the factor
    check, and a stored `28.35` must not pass because
    the line says `28.3`. The check is the stored grams.
    Enter on that line shows `28.349523125`, not `28.3`
    and not `28.35`. One pound paints `453.6`. One
    ounce and a typed `28.3` both paint `28.3`. Enter
    is how they differ: `28.349523125` against `28.3`.
    At 200 calories per 100 g the totals paint `56.7`
    and `56.6`. At 100 calories per 100 g both totals
    paint `28.3`, because the residue is under half a
    tenth. The idle line stays one decimal. The
    screen does not show the token.
4.  **Volume needs one density, in grams per milliliter.**
    There is no stored gram weight for a cup, a
    tablespoon, a teaspoon, a fluid ounce, or any other
    measure. Milliliters are the token amount times the
    accepted factor: `1 ml` is 1, `1 cup` is 240,
    `1 fl oz` is 30, `1 tbsp` is 15, and `1 tsp` is 5.
    The stored grams are one exact ratio: the parsed
    number, times that factor, times the density's
    reopen spelling. That ratio converts to float64
    once. It is not the product of two float64 values.
    Enter shows it with the spelling the gram field
    already uses, `FormatFloat` with `'f'` and
    precision `-1`. One cup and `240 ml` of the same
    food store the same grams. One cup at density
    `1.03` reopens as `247.2`, the float64 of the
    decimal 247.2, not the product of the float64
    values 240 and 1.03. The line paints `247.2`.
    Saving that reopened text unchanged keeps the
    stored grams. On the Enter that first stores a
    density, the ratio uses the reopen spelling of
    the density being saved, not the raw characters.
    `ParseFloat` of `0.10000000000000001` spells
    `0.1`, and that text stores density `0.1`. A
    cup on that same Enter stores 24 g and reopens
    as `24`. A later cup reopens as `24`. The ratio
    of the spelling `0.1` times 240 spells `24`.
    The raw decimal times 240 spells
    `24.000000000000004`, which is not stored.
    `+1.03` spells `1.03`. `+1.03` with `1 cup` on
    the same Enter stores 247.2 g, the same as
    density `1.03`, and one cup of that spelling
    spells `247.2`. The raw decimal times the
    factor is not the stored grams when it differs
    from that spelling. A fraction the scenarios do
    not list, including `1/3`, uses the same rule.
    Its reopen spelling is that one conversion, and
    saving the spelling keeps the grams. The
    customary cup is not an alternative. A 240 ml
    cup is not mixed with any other fluid-ounce
    size.
5.  **A refused amount writes nothing.** The editor stays
    open on the amount field when that field is the
    first failure. The stored grams and the density
    stay as they were. No food row is created
    for a new name. The error is `error:` and
    `amount: want a number greater than zero`. That
    covers zero, a negative, a blank, a non-number, a
    mixed number, a bad fraction, and any text that is
    not an accepted token. It is also the error when
    stored grams are not finite or not greater than
    zero after a legal amount converts, as in
    Decision 12. The grammar is checked before
    density, and before that product check. `0 cup`
    is this error, not the missing density error. Tab
    or the next typed character
    clears the error and does not save. A weight-log
    reload does not clear it and does not close the
    form. Esc cancels, writes nothing, and returns to
    the idle meal screen. It does not leave the list
    offering only `q quit`. The density error in
    Decision 9 clears the same way. Tab or the next
    typed character clears it and does not save. A
    weight-log reload does not clear it and does not
    close the form. Esc cancels and writes nothing.
6.  **No density means weight only.** A food with no
    density still accepts grams, `g`, ounces, and
    pounds, including a fraction of those. A volume is
    refused and nothing is written. The error is
    `error:` and `amount: this food has no density`.
    The editor stays open. The stored grams and the
    density stay as they were. The density field does
    not show `0` for that food. Zero is not a stand-in
    for none.
7.  **Enter reopens the one conversion as a bare
    number.** The stored grams are one exact ratio:
    the parsed number, times the factor, times the
    density's reopen spelling when the token is a
    volume. That ratio converts to float64 once. It
    is not the product of two float64 values. The
    amount field is filled with that value, using the
    spelling the gram field already uses,
    `FormatFloat` with `'f'` and precision `-1`, not
    the painted tenth and not the token that was
    typed. Saving that reopened text unchanged keeps
    the stored grams. `100.0` reopens as `100`.
    `1.25` reopens as `1.25`. One ounce reopens as
    `28.349523125`. One cup at density `1.03` reopens
    as `247.2`, the float64 of the decimal 247.2.
    A fraction the scenarios do not list, including
    `1/3`, uses the same rule. `1/3` of a gram
    reopens as `0.3333333333333333`. Saving that
    spelling keeps the grams. Saving `1 cup`, or any
    other accepted amount, replaces the grams through
    the conversion. The idle line still shows the name
    and the grams to one decimal, and it does not show
    the density.
8.  **Density is a field on the food.** The label is
    `density g/ml`. It is shown whether or not
    `track_macros` is true. A new food asks for the four
    per-100 g numbers, then the density, then the
    amount. A known food still asks only for the amount.
    Tab rebuilds the add form from the trimmed name.
    That rebuild clears the four nutrient drafts and
    the density draft. The amount draft survives. A
    volume in that surviving amount uses the density
    of the food the form was rebuilt for. Onto a
    stored food, the form asks only for the amount.
    The density field is dropped and its typed value
    is discarded. That food's density stays as it
    was. Onto stored `Oats` at density `1`, a
    `1 cup` typed before Tab stores 240 g, not
    247.2, and `Oats` stays density `1`. Onto a new
    name, the form asks for the four per-100 g
    numbers, then an empty density, then the amount.
    Calories, protein, carbohydrate, and fat are
    empty, the density field is empty, and the
    surviving amount is still there. Enter without
    that Tab is still `name: press tab` and writes
    nothing. Density is written only by Enter on the
    food that form was built for, or by `e`.
    `e` edits the four numbers and the density. It does
    not edit the name and it does not show the amount.
    Tab order for a new food is name, calories per 100 g,
    protein per 100 g, carbohydrate per 100 g, fat per
    100 g, `density g/ml`, `amount`, then name again.
    Tab order for `e` is the four per-100 g numbers, then
    `density g/ml`, then calories again. The line edit
    is the amount field only. Tab, Enter, and Esc stay
    what they are. Letters in the amount field are text,
    including the letters of a token. Enter parses the
    field. `q`, `a`, `f`, `b`, `[`, and `]` do not quit
    or change the date while the field is open.
    `ctrl+c` still quits.
9.  **Blank density means none. Zero is refused.**
    Whitespace at either end is removed first. Empty
    after that stores none, including a field of only
    spaces. It does not store zero. None reopens empty,
    not `0`. A set density reopens on the stored number,
    not a one-decimal paint: `1.03` shows `1.03`, not
    `1.0`, and `1` shows `1`, not `1.0`. The decimal
    shape is the nutrient shape, an optional `+` and at
    most one dot, and the value must be greater than
    zero. `+1.03` and `.5` are valid. Zero, a negative
    including `-0`, a slash, a unit token, an exponent,
    `inf`, and `NaN` are refused. The error is `error:`
    and `density: want a number greater than zero`. The
    editor stays open on the density field when the
    density is the first failure. The amount error is
    not shown in that case. Tab or the next typed
    character clears this density error and does not
    save, the same way Decision 5 clears a refused
    amount. Each of those starts from the error
    showing, and each writes nothing. A weight-log
    reload does not clear it and does not close the
    form. A reload after Tab has already cleared the
    error is not that beat. Esc still cancels and
    writes nothing. The stored density, the four
    numbers, and the grams stay as they were.
    One Enter writes the four numbers and the
    density together, or it writes nothing. Clearing a
    density that was set returns to none. Lines already
    stored keep their grams. The next volume is then
    refused. Changing the density does not rewrite grams
    already stored. The next volume uses the new
    density. Changing the four per-100 g numbers still
    updates every meal date that uses the food, and
    those lines keep their grams. The key reference
    and the how-to say that split in the two sentences
    in the Documentation section. The `e` row must not
    say that a density edit rewrites stored grams, and
    it must not say that changing the density changes
    every meal date.
10. **The allowance and the nutrient fields do not take
    these tokens.** `1 cup` in the allowance field is
    still `allowance: want a number`. `100 g` in
    calories is still `calories: want a number`. A blank
    macro is still zero when `track_macros` is false,
    and still refused when it is true. A blank density
    is none in either case. Density is not mandatory
    when tracking is on.
11. **Food mass stays grams on this screen.**
    `display_unit` does not scale it, in kilograms,
    pounds, or stone. The allowance, Daily Deficit,
    `ApplyTrend`, and both charts stay as they are. A
    meal still does not write a `daily_log` row. Stone
    is not a food unit. A kilogram token is refused
    because body weight already uses kilograms for a
    different quantity. An existing food that was stored
    before density existed still has none. Opening the
    meal tables does not invent a density for it and
    does not change its lines' grams.
12. **Every 64-bit slash fraction stays legal.** There
    is no kitchen-fraction list and no maximum gram
    weight. The typed number is checked first, as in
    Decision 5. After the one conversion in Decision 7,
    a product that is not finite, or that is not
    greater than zero, is refused with `error:` and
    `amount: want a number greater than zero`, and
    nothing is written. That second check is on the
    stored grams. The density and the grams already
    stored stay as they were. `1/3` of a gram is
    finite and greater than zero, so it saves. It is
    not this refusal. A fraction at either 64-bit
    end saves too, including
    `18446744073709551615/1` and
    `1/18446744073709551615` of a gram. A painted
    tenth of `0.0` is not this refusal when the
    stored grams are still greater than zero.
13. **One Enter shows the first failure and leaves the
    focus there.** The check walks only the fields on
    the form. A hidden field is not a failure and
    does not take focus. On the new-food add form
    every named field is on screen, and that order is
    calories, protein, carbohydrate, fat, density,
    then amount. S25 is that form. A line edit checks
    the amount and does not check calories, the
    macros, or the density. `e` checks calories,
    protein, carbohydrate, fat, then density, and
    does not check the amount. A blank macro with
    `track_macros` off is still zero. A blank macro
    with `track_macros` on is an error in that order.
    A blank density is still none. Neither a blank
    macro stored as zero nor a blank density is an
    error, so a later field can still fail. A blank
    calories field, with every other field valid,
    shows `calories: want a number` and the focus
    stays on calories per 100 g. Blank calories wins
    over a blank protein, a density of `0`, and a bad
    amount. Density `0` with amount `1 cup` shows
    `density: want a number greater than zero` on the
    density field and does not show the amount error.
    Nothing is written. A cold line edit, opened from
    the idle screen, fills the amount with the stored
    grams and saves that text. A cold `e` fills the
    four numbers and the density from the stored food
    and saves them. Empty hidden widgets do not block
    either save. Those saves keep the grams.

## Acceptance scenarios

Today in these scenarios is 10 November 1990 unless a
scenario says otherwise. The weight log, when it has
rows, has 1 November 1990 selected unless the scenario
says otherwise. `track_macros` is omitted unless the
scenario sets it. The default leaves that key unset.

### S1 — A bare number and g store the same grams

**Given** the meal screen for an empty 1990-11-10
**When** a new food `Oats` is saved with 200 calories per
100 g, blank macros, a blank density, and amount `100`
**Then** the idle line shows `Oats` and `100.0`
**And** the idle screen does not show `grams` as a field
label
**And** the idle screen does not show a unit token
**When** Enter is pressed
**Then** the field is labeled `amount`
**And** the field shows `100`, not `100.0`
**When** Enter is pressed without a change
**Then** the line still shows `100.0`
**When** a known `Oats` is added with amount `100 g`, and
again with `100g`, `100.0`, `+100`, and `.5`
**Then** the `100 g`, `100g`, `100.0`, and `+100` lines
each show `100.0`
**And** Enter on each of those shows `100`, not `100.0`
**And** the `.5` line shows `0.5`
**And** Enter on that line shows `0.5`
**And** the known-food add did not ask for calories or
for density

### S2 — Ounces and pounds store the full product

**Given** a food with 200 calories per 100 g and no
density
**When** the amount is `1 oz`
**Then** the idle line paints `28.3`
**And** the idle line does not show `28.35`, `28.4`, or
`oz`
**And** the screen shows `calories 56.7`
**And** the screen does not show `calories 56.6`
**When** Enter is pressed
**Then** the field shows `28.349523125`
**And** the field does not show `28.3` or `1 oz`
**When** another line of that food is `1 lb`
**Then** that line paints `453.6`
**And** Enter on it shows `453.59237`
**When** another line of that food is `16 oz`
**Then** that line paints `453.6`
**And** Enter on it shows `453.59237`
**And** the pound line and the `16 oz` line store the
same grams

### S3 — A slash fraction is an amount

**Given** a food with no density
**When** the amount is `1/4`
**Then** the line paints `0.2`
**And** the line does not show `0.3` or `1/4`
**When** Enter is pressed
**Then** the field shows `0.25`, not `0.2`
**When** the amount on a new line is `3/4`
**Then** the line paints `0.8`
**And** Enter shows `0.75`
**When** the amount on a new line is `2/4`
**Then** the line paints `0.5`
**And** Enter shows `0.5`
**When** the amount on a new line is `1/4 oz`
**Then** the line paints `7.1`
**And** Enter shows `7.08738078125`
**And** the field does not show `1/4 oz`

### S4 — Volume uses the density and the label factors

**Given** a new food saved with density `1.03`, 100
calories per 100 g, and amount `1 cup`
**Then** the line paints `247.2`
**And** the screen shows `calories 247.2`
**And** Enter shows `247.2`
**And** the idle line does not show `cup`
**When** a known line of that food is `240 ml`
**Then** that line paints `247.2`
**And** Enter shows `247.2`
**And** the two lines store the same grams
**Given** a food whose density is `1`
**When** the amounts are `1 cup`, `240 ml`, `8 fl oz`,
`1 fl oz`, `1 tbsp`, `1 tsp`, and `1 ml`, on separate
lines
**Then** those lines paint `240.0`, `240.0`, `240.0`,
`30.0`, `15.0`, `5.0`, and `1.0`
**And** Enter on the cup line and on the `240 ml` line
shows `240`
**And** Enter on the tablespoon line shows `15`
**When** a known line of the density `1` food is
`1/4 cup`
**Then** that line paints `60.0`
**And** Enter shows `60`

### S5 — Case and the gap before the token

**Given** a food whose density is `1`
**When** each of `1 fl oz`, `1fl oz`, `1  fl oz`,
`1 FL OZ`, and `1 Fl Oz` is saved on its own line
**Then** each line paints `30.0`
**And** Enter on each shows `30`
**When** the amounts `1g` and `1 G` are saved
**Then** each line paints `1.0`
**And** Enter on each shows `1`
**When** the amount is `2 CUP`
**Then** the line paints `480.0`
**And** Enter shows `480`
**When** the amount `1 fl  oz` is entered
**Then** the editor stays open
**And** the screen shows `error:` and
`amount: want a number greater than zero`
**And** no line was written for that amount

### S6 — A bad amount writes nothing

**Given** one line of 100 g of a food whose density is
`1.03`
**And** the line paints `100.0`
**When** each of `0`, `0 g`, `0.0`, `0 cup`, `-1`,
`-1 oz`, `-0`, `-1/4`, a blank field, a field of only
spaces, `1 1/2`, `1 1/2 cup`, `1/2/3`, `1/0`, `0/1`,
`0/0`, `1 / 4`, `1/ 4`, `+1/4`, `1e2`, `1e2 g`, `inf`,
`NaN`, `c`, `1 c`, `1 C`, `1 cups`, `1 ounces`,
`1 pounds`, `1 tablespoons`, `1 teaspoons`,
`1 milliliters`, `1 floz`, `1 gram`, `1 grams`,
`1 ounce`, `1 pound`, `1 stone`, `1 st`, `1 kg`,
`1 kilogram`, `oz 1`, `1 oz g`, and `cup` is entered
with Enter, one at a time
**Then** each Enter leaves the editor open
**And** each shows `error:` and
`amount: want a number greater than zero`
**And** none of them shows `no density`
**When** Esc is pressed after each refusal
**Then** the only line still paints `100.0`
**And** the density is still `1.03`
**And** no other line was written

### S7 — No density accepts weight and refuses volume

**Given** one line of 100 g of a food with no density
**When** each of `1 cup`, `1 CUP`, `1 fl oz`, `1 tbsp`,
`1 tsp`, and `240 ml` is entered with Enter, one at a
time
**Then** each Enter leaves the editor open
**And** each shows `error:` and
`amount: this food has no density`
**And** none of them shows `greater than zero`
**When** Esc is pressed after each refusal
**Then** the only line still paints `100.0`
**And** opening the food still shows an empty density
**And** the density field does not show `0`
**When** the amounts `1 oz`, `1 lb`, `1 g`, and `50`
are saved as further lines
**Then** those lines paint `28.3`, `453.6`, `1.0`, and
`50.0`
**And** the original line still paints `100.0`

### S8 — Enter reopens stored grams, not the token

**Given** a line stored from `1 oz` on a food with no
density
**When** Enter is pressed
**Then** the field shows `28.349523125`
**And** the field does not show `28.3` or `oz`
**When** Enter is pressed without a change
**Then** the line still paints `28.3`
**And** opening the field again shows `28.349523125`
**Given** a line stored from `1 cup` of a food whose
density is `1.03`
**When** Enter is pressed
**Then** the field shows `247.2`
**When** that bare `247.2` is saved
**Then** the line still paints `247.2`
**And** the field still shows `247.2`
**When** the amount is replaced with `1 oz`
**Then** the line paints `28.3`
**And** Enter shows `28.349523125`
**And** the density is still `1.03`

### S9 — A new food asks for density; a known food does not

**Given** the add form for a name that is not stored
**Then** the fields are name, calories per 100 g,
protein per 100 g, carbohydrate per 100 g, fat per
100 g, `density g/ml`, and `amount`, in that order
**And** the density field is empty
**When** Tab walks from the name through every field
**Then** the focus returns to the name
**And** nothing was stored
**When** the food is saved with calories `100`, the
three macros `0`, density `1.03`, and amount `1 cup`
**Then** the line paints `247.2`
**When** `e` is pressed
**Then** the density field shows `1.03`
**And** the form does not show `amount`
**And** the name is not a field
**When** Esc is pressed and the known name is added
**Then** the form asks only for the amount
**And** it does not show `density g/ml`
**And** it does not show calories per 100 g

### S10 — Blank is none; zero and negative density are refused

**Given** a food with no density and one line of 100 g
**When** `e` is pressed
**Then** the density field is empty
**And** it does not show `0`
**When** `0` is entered and Enter is pressed
**Then** the editor stays open
**And** the screen shows `error:` and
`density: want a number greater than zero`
**When** Esc is pressed and `e` is pressed
**Then** the density field is empty
**And** the line still paints `100.0`
**When** each of `0.0`, `-1`, `-0`, `1 g`, `1e2`,
`inf`, `NaN`, and `1/4` is entered on the density
field
**Then** each is that same density error
**And** the density stays unset
**When** the density is saved as `1.03`
**Then** `e` shows `1.03`, not `1.0`
**When** Enter is pressed without a change
**Then** `e` still shows `1.03`
**When** `0` is entered
**Then** the editor stays open with the density error
**When** Esc is pressed
**Then** `e` shows `1.03`
**And** the line still paints `100.0`
**When** the density is saved as `+1.03`
**Then** `e` shows `1.03`
**When** the density is saved as `.5`
**Then** `e` shows `0.5`

### S11 — Clearing density keeps the stored grams

**Given** a line stored from `1 cup` of a food whose
density is `1.03`, so the line paints `247.2`
**When** the density field is cleared and Enter is
pressed
**Then** the line still paints `247.2`
**And** Enter on the line shows `247.2`
**When** `e` is pressed
**Then** the density field is empty
**When** a known line is entered as `1 cup`
**Then** the editor stays open
**And** the screen shows
`amount: this food has no density`
**And** the cup line was not added
**And** the original line still paints `247.2`
**When** a known line is entered as `100`
**Then** that new line paints `100.0`
**And** the original line still paints `247.2`

### S12 — Changing density does not rewrite grams

**Given** density `1` and 100 calories per 100 g
**And** 1990-11-10 has a line stored from `1 cup`
**And** 1990-11-11 has a line stored from `1 cup`
**Then** each line paints `240.0`
**And** Enter on each shows `240`
**When** the density is changed to `1.03`
**Then** 1990-11-10 still paints `240.0`
**And** 1990-11-11 still paints `240.0`
**And** Enter on each old line shows `240`
**And** `e` shows `1.03`
**When** a new line on 1990-11-10 is `1 cup`
**Then** that line paints `247.2`
**And** Enter on it shows `247.2`
**And** the older line on 1990-11-10 still paints
`240.0`
**And** 1990-11-11 still paints `240.0`

### S13 — The four per-100 g numbers still update every date

**Given** the two dates and the lines from S12, after
the density is `1.03`
**When** calories per 100 g are changed from `100` to
`200`
**Then** 1990-11-10 shows `calories 974.4`
**And** the older line still paints `240.0`
**And** the newer line still paints `247.2`
**And** 1990-11-11 shows `calories 480.0`
**And** that line still paints `240.0`
**And** the density is still `1.03`

### S14 — track_macros does not hide or require density

**Given** `track_macros` is omitted
**When** a new food is added
**Then** the form shows protein, carbohydrate, fat,
and `density g/ml`
**When** the three macros and the density are left
blank and the amount is `100`
**Then** the line is saved
**And** the density is none
**And** the idle screen does not show `protein`,
`carbohydrate`, or `fat`
**Given** `track_macros` is true
**And** the add form has been built for a new name
**When** calories are `100`, carbohydrate and fat are
`0`, the density is blank, the amount is `100`, and
protein is left blank, and Enter is pressed
**Then** the editor stays open
**And** the focus stays on protein per 100 g
**And** the screen shows `error:` and
`protein: want a number`
**And** the screen does not show the density error
**And** the screen does not show the amount error
**And** nothing is written
**When** the three macros are `0`, the density is
blank, and the amount is `100`
**Then** the line is saved
**And** the density is none
**And** the screen shows `protein 0.0`

### S15 — The old keys stay, and there is no new key

**Given** the idle meal screen with one food line
**Then** the help line is the idle line in Decision 1
**When** `a` is pressed
**Then** the add form opens on the name
**When** Esc is pressed and Enter is pressed
**Then** the amount field opens on the stored grams
**When** Esc is pressed and `e` is pressed
**Then** the food edit shows `density g/ml`
**And** it does not show `amount`
**And** the help line is the form line in Decision 1
**When** calories are replaced with `100 g` and Enter
is pressed
**Then** the screen shows `calories: want a number`
**And** the food's calories are unchanged
**When** Esc is pressed and `x` is pressed
**Then** the line is gone
**And** adding the same name asks only for the amount
**When** `b` is pressed
**Then** the allowance editor's help line does not
name Tab
**When** `1 cup` is entered and Enter is pressed
**Then** the editor stays open
**And** the screen shows `allowance: want a number`
**And** the screen does not show `no density`
**When** Esc is pressed
**Then** the allowance is unchanged
**And** idle `[` and `]` still move the meal date
**And** idle Esc still returns to the same list row
**And** idle `f` does not open another screen
**And** `f` on the daily list is still what opens
today's meal screen

### S16 — Letters in the amount field are text

**Given** the amount field is open and empty on
1990-11-10
**When** `q`, `a`, `f`, `b`, `[`, `]`, `o`, and `z`
are typed
**Then** those characters are in the field
**And** the date is still 1990-11-10
**And** the process does not quit
**And** no line was written
**When** Esc is pressed
**Then** nothing was written
**Given** the amount field is open again
**When** `ctrl+c` is pressed
**Then** the process quits

### S17 — A refused amount survives a reload

**Given** the amount editor is showing
`amount: want a number greater than zero`
**When** a weight-log reload finishes
**Then** that error is still shown
**And** the editor is still open
**And** the screen is still the meal screen
**When** Tab is pressed
**Then** the error is gone
**And** nothing was saved
**When** the same refusal is entered again
**And** the next character is typed
**Then** the error is gone
**Given** the editor is showing
`amount: this food has no density`
**When** a weight-log reload finishes
**Then** that error is still shown
**And** the editor is still open
**When** Esc is pressed
**Then** the idle meal screen shows the grams from
before the refusal
**And** the screen is not a list that offers only
`q quit`
**When** Esc is pressed again
**Then** the daily list shows its normal help

### S18 — Pounds do not scale food mass

**Given** `display_unit` is pounds
**And** a food with 200 calories per 100 g and no
density
**When** the amount is `1 oz` and the allowance is
`100`
**Then** the line paints `28.3`
**And** the screen shows `calories 56.7`
**And** the screen shows `remaining 43.3`
**And** the screen does not show those figures
converted as body weight
**And** the screen does not show `Daily deficit`
**When** the amount `1 st` is entered
**Then** it is refused
**And** the ounce line is unchanged

### S19 — The weight log and an older food stay

**Given** a daily list with at least two weighed days
**And** both chart captions contain `Daily deficit:`
**And** a food and a line that were stored before a
density existed
**When** the meal tables are opened
**Then** that food still has no density
**And** the old line's grams are unchanged
**And** a volume of that food is refused
**And** a weight of that food still saves
**When** the captions are noted again after that save
**Then** each caption is the same text as before
**And** the list's weight and trend text is unchanged
**And** a meal date with no daily log row still has
none
**And** an existing daily log row is unchanged
**And** the daily log gains no meal column
**And** the line stores grams and which food, not a
token and not a density
**And** the file is still mode `0600`
**And** no second database file is created

### S20 — The key reference and the how-to

**Given** the key reference, the meal how-to, and the
config reference
**When** this slice is documented
**Then** the meal Enter row says the selected line's
amount, which is stored as grams
**And** the meal `e` row contains these two sentences,
not one sentence that claims both:

```text
The four per-100 g numbers change every meal date that uses the food.
Density is on that edit, and changing it does not rewrite stored grams, and the next volume uses the new density.
```

**And** the `e` row has no sentence in which `density`
and `every meal date` both appear
**And** the `e` row does not say that a density edit
rewrites stored grams
**And** the key table does not gain a key
**And** the `f` row is still the list row that opens
today's meal screen
**And** the list does not gain a `[` or `]` row
**And** `docs/how-to/meal-day.md` names the tokens
`g`, `oz`, `lb`, `cup`, `fl oz`, `tbsp`, `tsp`, and
`ml`
**And** that how-to gives 28.349523125 g per ounce,
16 of those ounces per pound, 240 ml per cup, 30 ml
per fluid ounce, 15 ml per tablespoon, and 5 ml per
teaspoon, as the factors in use
**And** that how-to does not call those factors
proposals
**And** that how-to says a blank density means weight
only
**And** that how-to contains the same two sentences as
the meal `e` row
**And** that how-to says the line shows grams
**And** that how-to no longer says that a volume is
later work
**And** recipes, the starter set, and renaming or
deleting a food stay later work
**And** `docs/how-to/README.md` still lists that
how-to
**And** the config reference does not gain a density
key or an amount key

### S21 — One ounce and typed 28.3 share a tenth

**Given** a food with 200 calories per 100 g and no
density, and that date has no other line
**When** the amount is `1 oz`
**Then** the idle line paints `28.3`
**And** the screen shows `calories 56.7`
**And** Enter shows `28.349523125`
**When** that amount is replaced with `28.3`
**Then** the idle line paints `28.3`
**And** the screen shows `calories 56.6`
**And** Enter shows `28.3`
**Given** a food with 100 calories per 100 g and no
density, and that date has no other line
**When** the amount is `1 oz`
**Then** the idle line paints `28.3`
**And** the screen shows `calories 28.3`
**When** that amount is replaced with `28.3`
**Then** the idle line paints `28.3`
**And** the screen shows `calories 28.3`
**And** Enter shows `28.3`
**And** neither idle line shows `28.35` or
`28.349523125`

### S22 — One third of a gram reopens as one conversion

**Given** a food with no density
**When** the amount is `1/3`
**Then** the line paints `0.3`
**And** the line does not show `1/3`
**When** Enter is pressed
**Then** the field shows `0.3333333333333333`
**When** that text is saved unchanged
**Then** the line still paints `0.3`
**And** Enter shows `0.3333333333333333`

### S23 — Tab keeps the amount and clears the drafts

**Given** a stored food `Oats` with density `1` and
100 calories per 100 g
**And** an add form for a new name, not yet stored
**When** `1.03` is typed in the density field, the
amount is `1 cup`, the name is changed to `Oats`,
and Enter is pressed with no Tab after that name
change
**Then** the form stays open
**And** the screen shows `error:` and
`name: press tab`
**And** nothing is written
**And** `Oats` still has density `1`
**When** Tab is pressed
**Then** the form asks only for the amount
**And** it does not show `density g/ml`
**And** it does not show `1.03`
**And** it does not show `calories per 100 g`
**And** it does not show protein per 100 g,
carbohydrate per 100 g, or fat per 100 g
**And** the amount field still shows `1 cup`
**When** Enter is pressed on that amount
**Then** the line uses `Oats`
**And** the line stores 240 g, not 247.2
**And** the line paints `240.0`
**And** Enter on the line shows `240`, not `247.2`
**And** `Oats` still has density `1`
**And** the four per-100 g numbers are unchanged
**Given** an add form that asks only for the amount
of `Oats`
**When** the amount is `40`, the name is changed to
`Rye`, and Enter is pressed without Tab
**Then** the screen shows `error:` and
`name: press tab`
**And** no food is stored under `Rye`
**And** `Oats` still has density `1`
**When** Tab is pressed
**Then** the form asks for the four per-100 g
numbers, then an empty `density g/ml`, then the
amount
**And** calories, protein, carbohydrate, and fat
are empty
**And** those four values are not `0` and not `100`
**And** the density field is empty
**And** the density field does not show `1` or
`1.03`
**And** the amount field still shows `40`
**When** the four numbers are `1`, the density is
`1.5`, and the amount is replaced with `100`
**And** Enter is pressed
**Then** `Rye` and the line are stored
**And** `Rye` has density `1.5`
**And** `Oats` still has density `1`

### S24 — A bad product is refused; a long fraction is not

**Given** a food with no density
**When** the amount is `18446744073709551615/1`
**Then** the line is saved
**And** the idle line paints `18446744073709551616.0`
**And** Enter shows `18446744073709552000`
**When** that reopened text is saved unchanged
**Then** Enter still shows `18446744073709552000`
**When** a new line of that food is
`1/18446744073709551615`
**Then** that line is saved
**And** the idle line paints `0.0`
**And** Enter shows
`0.00000000000000000005421010862427522`
**When** that reopened text is saved unchanged
**Then** Enter still shows
`0.00000000000000000005421010862427522`
**Given** a food whose density is already the smallest
positive magnitude a density can be
**When** the amount `1/2 ml` is entered
**Then** the line is saved
**And** the stored grams are that same smallest
positive magnitude
**And** Enter shows its reopen spelling
**When** that reopened text is saved unchanged
**Then** the stored grams stay
**When** the amount `1/4 ml` is entered
**Then** the editor stays open
**And** the screen shows `error:` and
`amount: want a number greater than zero`
**And** the `1/2 ml` line is unchanged
**And** no line was written for `1/4 ml`
**And** the density is unchanged
**And** the screen does not show `no density`
**Given** a food whose density is already the largest
finite magnitude a density can be
**When** the amount `1 cup` is entered
**Then** the editor stays open
**And** the screen shows `error:` and
`amount: want a number greater than zero`
**And** no line was written for that cup
**And** the density is unchanged
**When** the amount `1 g` is entered
**Then** that line is saved
**And** it paints `1.0`
**And** Enter shows `1`

### S25 — The first bad field keeps the focus

**Given** `track_macros` is true
**And** the add form has been built for a new name
**When** calories are blank, protein, carbohydrate,
and fat are `0`, the density is blank, the amount is
`100`, and Enter is pressed
**Then** the focus stays on calories per 100 g
**And** the screen shows `error:` and
`calories: want a number`
**And** nothing is written
**When** calories stay blank, protein is blank,
carbohydrate and fat are `0`, the density is `0`,
the amount is `0`, and Enter is pressed
**Then** the focus stays on calories per 100 g
**And** the screen shows `error:` and
`calories: want a number`
**And** the screen does not show `protein:`
**And** the screen does not show `density:`
**And** the screen does not show `amount:`
**And** nothing is written
**Given** `track_macros` is omitted
**And** the add form has been built for a new name
**When** calories are `100`, the three macros are
blank, the density is `0`, the amount is `1 cup`,
and Enter is pressed
**Then** the focus stays on `density g/ml`
**And** the screen shows `error:` and
`density: want a number greater than zero`
**And** the screen does not show `amount:`
**And** the screen does not show `protein:`
**And** nothing is written
**When** calories stay `100`, the three macros stay
blank, the density is cleared, the amount is `0`,
and Enter is pressed
**Then** the focus stays on `amount`
**And** the screen shows `error:` and
`amount: want a number greater than zero`
**And** the screen does not show `protein:`
**And** nothing is written

### S26 — A cold edit saves only the fields it shows

**Given** the meal screen is opened cold on a day
that already has one stored line
**And** this opening did not type into the add form
**And** the line is `Oats` at 100 g
**And** `Oats` is 200 calories, 10 protein, 30
carbohydrate, and 4 fat per 100 g, with density
`1.03`
**When** Enter is pressed
**Then** the field is labeled `amount`
**And** the field shows `100`
**And** the form does not show `calories per 100 g`
**And** it does not show protein per 100 g,
carbohydrate per 100 g, fat per 100 g, or
`density g/ml`
**When** Enter is pressed without a change
**Then** the line still paints `100.0`
**And** the screen does not show `error:`
**And** the grams are still 100
**When** Enter is pressed
**Then** the field shows `100`
**When** Esc is pressed and `e` is pressed
**Then** calories per 100 g show `200`
**And** protein per 100 g shows `10`
**And** carbohydrate per 100 g shows `30`
**And** fat per 100 g shows `4`
**And** the density field shows `1.03`
**And** the form does not show `amount`
**When** Enter is pressed without a change
**Then** the line still paints `100.0`
**And** the screen does not show `error:`
**And** the grams are still 100
**When** `e` is pressed
**Then** those four numbers and the density show
the same five values

### S27 — The creating Enter uses the density spelling

**Given** a new food, not yet stored
**When** it is saved with 100 calories per 100 g,
blank macros, density `0.10000000000000001`, and
amount `1 cup`, on that same Enter
**Then** the line paints `24.0`
**And** Enter shows `24`
**And** Enter does not show `24.000000000000004`
**When** `e` is pressed
**Then** the density field shows `0.1`
**And** it does not show `0.10000000000000001`
**When** Esc is pressed and a known line of that
food is `1 cup`
**Then** that line paints `24.0`
**And** Enter shows `24`
**Given** another new food
**When** it is saved with 100 calories per 100 g,
blank macros, density `+1.03`, and amount `1 cup`,
on that same Enter
**Then** the line paints `247.2`
**And** Enter shows `247.2`
**And** the stored grams are the same as one cup
at density `1.03`
**When** `e` is pressed
**Then** the density field shows `1.03`

### S28 — A refused density clears like a refused amount

**Given** a food with no density and one line of
100 g
**When** `e` is pressed, the density is `0`, and
Enter is pressed
**Then** the editor stays open
**And** the screen shows `error:` and
`density: want a number greater than zero`
**When** Tab is pressed
**Then** the error is gone
**And** nothing was saved
**And** the stored density is still unset
**When** Esc is pressed, `e` is pressed, the density
is `0`, and Enter is pressed
**Then** that density error is showing
**When** `1` is typed
**Then** the error is gone
**And** nothing was saved
**And** the stored density is still unset
**When** Esc is pressed, `e` is pressed, the density
is `0`, and Enter is pressed
**Then** that density error is showing
**When** a weight-log reload finishes
**Then** that error is still shown
**And** the editor is still open
**And** the screen is still the meal screen
**And** nothing was written
**When** Esc is pressed
**Then** the idle meal screen shows the line at
`100.0`
**And** the density is still unset
**And** the screen is not a list that offers only
`q quit`

## Functional requirements

-   **FR1.** The amount label replaces `grams` on add
    and on the line edit. A bare number and `g` store
    the same grams, including `100`, `100 g`, `100g`,
    `+100`, and `.5`. The idle line shows the name and
    the stored grams to one decimal and does not show
    the token. Enter reopens the full stored grams as
    a bare number: `100.0` reopens as `100`, and `.5`
    reopens as `0.5`. A known food asks only for the
    amount. The idle help line is the line in
    Decision 1.
-   **FR2.** One ounce stores 28.349523125 g and the
    line paints `28.3`. Enter shows `28.349523125`,
    not `28.3` and not `28.35`. A food at 200
    calories per 100 g then shows `calories 56.7`, not
    `calories 56.6`. One pound and `16 oz` each store
    453.59237 g and paint `453.6`. `1/4` stores 0.25
    and paints `0.2`, and Enter shows `0.25`. `3/4`
    stores 0.75 and paints `0.8`. `2/4` stores 0.5.
    `1/4 oz` stores 7.08738078125 g and paints `7.1`.
-   **FR3.** Volume is milliliters times density, the
    full product. The factors are 240, 30, 15, 5, and
    1 for cup, fl oz, tbsp, tsp, and ml. One cup and
    `240 ml` of density `1.03` each store 247.2 g,
    paint `247.2`, and reopen as `247.2`. At density
    `1`, one cup, `240 ml`, and `8 fl oz` store 240,
    and one fl oz, tbsp, tsp, and ml store 30, 15, 5,
    and 1. `1/4 cup` at density `1` stores 60. `1g`,
    `1 g`, and `1 G` store 1. `1 fl oz`, `1fl oz`,
    `1  fl oz`, and `1 FL OZ` of density `1` store
    30. `2 CUP` stores 480. `1 fl  oz` is refused.
-   **FR4.** A refused amount leaves the editor open,
    writes no line, and leaves the stored grams and
    the density as they were. The error is `error:`
    and `amount: want a number greater than zero`
    for the inputs in S6. Grammar is checked before
    density, so `0 cup` uses that error. A volume of
    a food with no density uses `error:` and
    `amount: this food has no density`, writes
    nothing, and does not show the density as `0`.
    That food still accepts `1 oz`, `1 lb`, `1 g`,
    and a bare number.
-   **FR5.** Saving the bare number Enter showed
    stores those grams. Saving `1 oz` in place of a
    stored cup replaces the grams through the weight
    factor and leaves the density as it was. The idle
    line still does not show the token or the
    density.
-   **FR6.** A new food's fields are the four per-100 g
    numbers, then `density g/ml`, then `amount`. `e`
    shows the four numbers and the density, not the
    amount. A known food asks only for the amount.
    Blank density means none and reopens empty. Zero,
    a negative, and a non-number are refused with
    `density: want a number greater than zero`, and
    the stored density stays. `+1.03` reopens as
    `1.03`. `.5` reopens as `0.5`. One Enter writes
    the four numbers and the density together, or
    writes nothing. When more than one field is wrong,
    Enter shows the first failure in Decision 13 and
    writes nothing. `track_macros` does not hide the
    density and does not make it mandatory. Blank
    macros stay as they are today.
-   **FR7.** Clearing a set density, or changing it,
    does not rewrite grams already stored. After a
    clear, a later cup is refused and a later bare
    number still saves. After a change from `1` to
    `1.03`, an old cup still stores 240 and a new cup
    stores 247.2. Changing calories per 100 g from
    `100` to `200` on the S12 lines shows
    `calories 974.4` on 1990-11-10 and
    `calories 480.0` on 1990-11-11, and the gram
    paints stay `240.0` and `247.2`.
-   **FR8.** Idle `a`, Enter, `e`, `x`, `b`, `[`, `]`,
    Esc, and `q` stay as they are. `x` removes the
    line and keeps the food. `e` edits the density and
    the four per-100 g numbers. Those four numbers
    still change every date that uses the food.
    Changing the density does not rewrite stored
    grams. The next volume uses the new density.
    The allowance field still refuses `1 cup` with
    `allowance: want a number`. Calories still refuse
    `100 g`. Letters in the amount field are text.
    `ctrl+c` still quits. There is no new key. `f`
    on the list remains the only door. The form help
    line still names Tab. The allowance help line
    does not.
-   **FR9.** Tab or the next typed character clears a
    refused amount and a refused density. A
    weight-log reload does not clear either, and it
    does not close the form. Esc cancels to the idle
    meal screen and writes nothing. A further Esc
    shows the ordinary list help, not a list that
    offers only `q quit`. This holds for both amount
    errors and for
    `density: want a number greater than zero`. Tab
    and the next typed character are separate beats.
    Each starts from that error showing, and each
    writes nothing. The reload beat starts while the
    density error is showing. It does not follow a
    Tab that already cleared the error.
-   **FR10.** `display_unit` pounds does not scale the
    ounce in S2. With an allowance of `100` the screen
    shows `28.3`, `calories 56.7`, and
    `remaining 43.3`, and not `Daily deficit`. `1 st`
    is refused. Saving an amount does not change
    either chart caption, the list's weight or trend
    text, or a daily log row. A day with no daily log
    row does not gain one. The file stays mode `0600`.
    Meals do not create a second database file. A food
    stored before density existed still has none, its
    old grams are unchanged, volume is refused, and
    weight still saves. The line does not store the
    token or a copy of the density.
-   **FR11.** `docs/reference/keys.md` updates the meal
    Enter row and the meal `e` row as in S20 and does
    not gain a key. The `e` row contains S20's two
    sentences, not one sentence that claims both, and
    it does not say that a density edit rewrites
    stored grams. `docs/how-to/meal-day.md` names the
    tokens and the factors as the factors in use,
    contains those same two sentences, says a blank
    density means weight only, says the line shows
    grams, and stops calling a volume later work.
    Recipes, the starter set, and renaming or deleting
    a food stay later work. The how-to index still
    lists that page. The config reference does not
    gain a density key or an amount key.
-   **FR12.** One ounce and a typed `28.3` both paint
    `28.3` on the idle line, which stays one decimal.
    Enter shows `28.349523125` for the ounce and
    `28.3` for the typed number. At 200 calories per
    100 g the totals paint `56.7` and `56.6`. At 100
    calories per 100 g both totals paint `28.3`.
-   **FR13.** `1/3` of a gram uses the one conversion
    in Decision 7. The ratio is one third. Enter shows
    `0.3333333333333333`. The idle line paints `0.3`.
    Saving that reopened text unchanged keeps the
    stored grams. A fraction the other scenarios do
    not list uses that same rule.
-   **FR14.** Tab rebuilds the add form from the
    trimmed name. That rebuild clears the four
    nutrient drafts and the density draft. The amount
    draft survives. A volume in that surviving amount
    uses the density of the food the form was rebuilt
    for. Onto a stored food, the form asks only for
    the amount, the density field is dropped, and a
    typed density is discarded. That food's density
    stays as it was. Onto stored `Oats` at density
    `1`, a `1 cup` typed before Tab stores 240 g, not
    247.2, and `Oats` stays density `1`. Onto a new
    name, the form asks for the four per-100 g
    numbers, then an empty density, then the amount.
    Calories, protein, carbohydrate, and fat are
    empty, the density field is empty, and the
    surviving amount is still there. Enter without
    that Tab shows `name: press tab` and writes
    nothing. Density is written only by Enter on the
    food that form was built for, or by `e`.
-   **FR15.** Every 64-bit slash fraction is legal.
    There is no kitchen-fraction list and no maximum
    gram weight. The typed number is checked first.
    After conversion, stored grams that are not
    finite, or are not greater than zero, are refused
    with `amount: want a number greater than zero`,
    and nothing is written. `1/3` is not that refusal.
    `18446744073709551615/1` and
    `1/18446744073709551615` of a gram are saved.
    `1/2 ml` of a food whose density is the smallest
    positive magnitude a density can be is saved as
    that magnitude. `1/4 ml` of that food is refused.
    `1 cup` of a food whose density is the largest
    finite magnitude a density can be is refused, and
    `1 g` of that food still saves.
-   **FR16.** One Enter shows the first failure and
    leaves the focus there. The order is calories,
    protein, carbohydrate, fat, density, then amount.
    That order is the fields on the form being saved.
    S25 is the new-food add form, where every named
    field is on screen. A hidden field is not a
    failure and does not take focus. A line edit and
    `e` are FR17. A blank macro with `track_macros`
    off, and a blank density, are not errors. Blank
    calories, with the other fields valid, shows
    `calories: want a number` on calories per 100 g.
    Blank calories wins over a blank protein, density
    `0`, and a bad amount. Density `0` with amount
    `1 cup` shows the density error on the density
    field and does not show the amount error. S14's
    blank protein is the protein error only, because
    the other nutrients are valid, the density is
    blank, and the amount is valid. Nothing is
    written.
-   **FR17.** Enter checks only the fields on the
    form. A line edit checks the amount and does not
    check calories, the macros, or the density. `e`
    checks calories, protein, carbohydrate, fat, then
    density, and does not check the amount. A hidden
    field is not a failure and does not take focus.
    A cold line edit, opened from the idle meal
    screen on a line already stored, fills the amount
    with the stored grams. Saving that text keeps
    those grams. A cold `e` fills the four numbers
    and the density from the stored food. Saving them
    keeps the grams. Empty hidden widgets do not
    block either save. The add form is not what
    filled those widgets.
-   **FR18.** On the Enter that first stores a
    density, the volume ratio uses the reopen
    spelling of the density being saved, not the raw
    characters. `0.10000000000000001` stores density
    `0.1`. A cup on that same Enter stores 24 g and
    reopens as `24`. A later cup reopens as `24`.
    `+1.03` with `1 cup` on the same Enter stores
    247.2 g, the same as density `1.03`. The raw
    decimal times the factor is not the stored grams
    when it differs from that spelling. S4 is
    unchanged: density `1.03` and `1 cup` still store
    247.2.

## Out of scope

-   Recipes are not this slice. An ingredient amount
    is not this slice. This slice does not log a
    fraction of a recipe or a gram weight of a
    recipe.
-   The starter set is not this slice.
-   Renaming a food is not this slice.
-   Deleting a food is not this slice. Removing a
    meal line does not delete the food.
-   The customary cup is not an alternative. One cup
    is 240 ml.
-   Stone is not a food unit. A kilogram token is
    not a food unit.
-   No key is added. `f` remains the only door.
-   The allowance grammar does not gain these tokens.
    The four per-100 g fields do not gain them.
-   `display_unit` is not read by this screen.
    `ApplyTrend`, Daily Deficit, and both chart
    captions stay as they are.
-   A meal does not write a `daily_log` row.
-   Colour and a new `[colors]` role are not added.
-   The alternate screen and the mouse stay off.
-   A second SQLite file is not created.

## Documentation

In `docs/reference/keys.md`, change the meal Enter row
so it edits the selected line's amount, which is stored
as grams. Change the meal `e` row so the two claims are
two sentences, not one sentence that claims both. The
row contains:

```text
The four per-100 g numbers change every meal date that uses the food.
Density is on that edit, and changing it does not rewrite stored grams, and the next volume uses the new density.
```

The row must not say that a density edit rewrites stored
grams, and it must not say that changing the density
changes every meal date. Do not add a key. The `f` row
stays the list. Do not add a list row for `[` or `]`.
Leave the month, monthly chart, and long-term bracket
rows as they are.

In `docs/how-to/meal-day.md`, say that the amount field
takes a bare number, which is grams, or a number and a
token. Name the tokens `g`, `oz`, `lb`, `cup`,
`fl oz`, `tbsp`, `tsp`, and `ml`. Say that case does
not matter, that `oz` is weight, and that `fl oz` is
volume. Give the factors as the factors in use, not as
proposals: one ounce is 28.349523125 g, one pound is
16 of those ounces, one milliliter is 1 ml, one cup is
240 ml, one fluid ounce is 30 ml, one tablespoon is
15 ml, and one teaspoon is 5 ml. Say that volume
multiplies those milliliters by the food's density in
grams per milliliter. Say that a blank density means
weight only, and that a volume is then refused and
nothing is written. Include, verbatim, the two
sentences in the key-row block above. Say that the
line shows grams, not the token. Say that Enter on a
line opens the stored grams as a bare number. Say that
a new food asks for the four per-100 g numbers and the
density, then the amount, and that a name already
stored asks only for the amount. Drop the sentence
that a volume is later work. Keep recipes, the starter
set, and renaming or deleting a food as later work.
Keep the statements that `display_unit` does not scale
food mass and that the allowance is not the chart's
Daily Deficit.

Do not add a how-to page. `docs/how-to/README.md`
already lists the meal page. Do not add a density key
or an amount key to `docs/reference/config.md`.
