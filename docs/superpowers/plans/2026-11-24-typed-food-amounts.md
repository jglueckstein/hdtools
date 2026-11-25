# Plan — Typed food amounts

**Spec**:
[`docs/superpowers/specs/2026-11-24-typed-food-amounts.md`](../specs/2026-11-24-typed-food-amounts.md)
**Slice**: S3 of
[`meal-planning.md`](../slices/meal-planning.md)
**Issue**: not filed
**Follows**:
[`2026-11-20-meal-allowance.md`](2026-11-20-meal-allowance.md)
(approved, shipped) and its plan. Do not edit that plan
or its spec. Do not edit
[`2026-10-30-meal-day.md`](../plans/2026-10-30-meal-day.md)
or
[`2026-11-14-optional-food-macros.md`](../specs/2026-11-14-optional-food-macros.md).
**Status**: draft

No production code until the spec's scenarios exist as
failing tests.

S3 turns a typed amount into stored grams on the shipped
meal screen. Volume and density are this slice. Recipes,
the starter set, and renaming or deleting a food stay
later. Do not implement them. Do not edit the slicing
record. `idea.md` already points at this spec. Do not
rewrite the other sections.

The factors are the accepted nutrition-label group, not
proposals. One cup is 240 ml. Do not offer the customary
cup. The stored grams are one ratio converted to float64
once. Do not multiply two float64 values.
`float64(240) * float64(1.03)` is not the product.

A test that does not compile is not red. The red step may
add the store operations the new tests call, a parser the
new tests call, and an amount-field branch that does not
save a converted amount. The parser may still accept a
bare positive decimal, because that is the shipped gram
field. It must not accept a token, a slash fraction, or a
volume, and it must not return a converted product. Those
additions must not pass a behavioural test. No successful
ounce, pound, fraction, or volume save in that step. No
successful density save in that step.

## Module structure

Line counts below are the files as this plan was written.
A non-test file stays at or under 500 lines. If a named
file would pass 500, do the split named on that row first.
Do not pre-split.

| File | Why |
| --- | --- |
| `internal/meal/amount.go` | New. The tokens and the factors live here so a later recipe can use the same conversion without importing the TUI. This is its own concern, not a split of `meal.go`. Preamble after the package clause, not a second package comment. |
| `internal/meal/amount_test.go` | Parser tests and the old-schema density test. Package `meal`. |
| `internal/meal/meal.go` | 269 lines. Nullable density on the food row, scanned as unset when NULL. Extend `AddFoodLine` and `UpdateFood` with a `*float64` density. Nil stores none. Update the preamble: this package stores an optional density and still does not draw a screen, convert display units, or store a recipe. |
| `internal/meal/meal_test.go` | 265 lines. Revise `TestMealMigrationKeepsDailyLog` so `food` has `density` and an insert that omits it leaves NULL. `meal_line` gains no column. `daily_log` columns stay. |
| `internal/tui/meal.go` | 473 lines. The label `amount`, the idle help string, and the preamble. The preamble may say a volume becomes grams through the food's density. Do not put the factor table here. If the edit would pass 500, move the allowance editor and `parseAllowance` to `internal/tui/mealallow.go` first. Do not pre-split. |
| `internal/tui/mealform.go` | 376 lines. The density field, tab order, and the call that parses the amount. If that work would pass 500, move the amount submit path to `internal/tui/mealamount.go` in this package first. Do not pre-split. |
| `internal/tui/meal_test.go` | Revise the tests listed below. The file is already past 500 lines; the limit is non-test files. `saveNew` gains one blank density tab in the same change that adds the field, not in the red step. |
| `internal/tui/mealamount_test.go` | New screen scenarios. Package `tui`, so the meal helpers stay in the package. |
| `internal/tui/mealallow_test.go` | `TestMealAllowanceKey` uses `mealHelp` and asserts the grams label. Revising the shared help constant covers the help line. Change that one focus assertion to `amount`. Do not change the allowance grammar. |
| `internal/tui/app.go` | 464 lines. No new case. `updateMeal` already owns meal keys. A `loadedMsg` must not clear a meal-form error and must not call `loadMealDay`. Do not import `database/sql` or a SQLite driver. |
| `cmd/hdtools/main.go` | 261 lines. No change. `meal.Open` already migrates the daily-log path. |
| `docs/reference/keys.md` | FR11. The meal Enter row and the meal `e` row. No new key. |
| `docs/how-to/meal-day.md` | FR11. Tokens and factors in use. Drop the sentence that a volume is later work. Keep recipes, the starter set, and renaming or deleting a food as later work. |
| `docs/how-to/README.md` | No new page. The meal-day link stays. |
| `docs/reference/config.md` | No density key and no amount key. |
| `idea.md` | Do not edit. |
| `CHANGELOG.md` | When the behaviour ships, one Unreleased bullet. Not in the failing-test step. |

`internal/meal` must not import `internal/tui`,
`internal/dailylog`, `internal/config`, or
`internal/units`. The TUI calls `ParseAmount` and the store
the way it already calls `Lines`. It does not import
`database/sql`. `cmd/hdtools` does not parse meal keys and
does not embed the new SQL.

Do not edit `internal/dailylog`, `internal/units`,
`internal/tui/chart.go`, or `internal/tui/longchart.go`.
Do not call `ApplyTrend` from an amount save. Do not add a
`[colors]` role. Do not pass `AltScreen` or a mouse option.
Do not `os.Create` the database.

New text keeps a literate preamble: why the file exists,
what it decided, and what it does not do. Returned store
errors use `%w` and a context. Do not bare-return `err`.
The amount and density refusals are the sentences the form
already shows. Do not wrap those sentences a second time.

## Algorithm notes

-   None is NULL. Zero is not none, because zero is
    refused. The column is nullable and has no DEFAULT
    and no NOT NULL. Existing rows must not be rewritten
    to a fake density. `CREATE TABLE IF NOT EXISTS` does
    not add a column to a table that already exists, so
    the new-database statement includes the column and
    the migrate also adds it when `pragma_table_info`
    has no `density` row:

    ```sql
    ALTER TABLE food ADD COLUMN density REAL
    ```

    Do not `UPDATE` after that. Do not `ALTER`
    `daily_log` or `meal_line`. `meal_line` stays
    `id`, `day`, `food_id`, and `grams`. There is no
    token column and no gram weight per spoon. A
    missing path still returns an error. Mode `0600`
    stays the daily log's file. Leave the foreign-key
    pragma on.
-   Scan density with `sql.NullFloat64`. `Food` carries
    `Density float64` and `HasDensity bool`. False means
    NULL, including when the numeric value would have
    been zero. `Lines` and `FoodByName` both scan it.
    The TUI must not treat a false flag as density 0.
-   `AddFoodLine` and `UpdateFood` take `density *float64`
    at the end. Nil writes NULL. A non-nil value is the
    number the form already accepted. One statement
    writes the four per-100 g numbers and the density.
    Neither call updates `meal_line.grams`. The red step
    may add the parameter and pass nil from the existing
    call sites so those tests still compile. Passing nil
    must leave existing foods weight-only.
-   `ParseAmount(raw string, density float64, hasDensity bool) (float64, error)`
    lives in `internal/meal`. The flag comes from
    `HasDensity`, not from a zero. Weight ignores the
    flag. Volume with the flag false returns
    `amount: this food has no density` and does not
    return grams. The TUI does not call it with the
    flag true and a non-positive density. The red stub
    may parse a bare positive decimal and return it
    unchanged. Any token or slash returns
    `amount: want a number greater than zero` until the
    green step.
-   Trim the field the way the other meal fields trim.
    Then take either the existing decimal shape or one
    fraction. A slash means the fraction path, so
    `1.5/2` is refused and is not a decimal plus junk.
    Fraction sides are `ParseUint` in base 10 and 64
    bits, value at least 1, no sign and no dot, and the
    digits sit against the slash. The fraction is that
    ratio, not a float64 divide and not integer
    division. The rest of the string, after
    trimming only its leading whitespace, must be empty
    or exactly one token. Compare tokens with
    `EqualFold`. Do not collapse whitespace inside the
    token. `fl oz` is one ordinary space. `1  fl oz`
    works. `1 fl  oz` does not.
-   Factors are exact ratios next to the parser, not
    float64 constants multiplied together. One ounce
    is 28349523125/1000000000. One pound is 16 of
    those ounces. One cup is 240 ml, one fluid ounce
    is 30, one tablespoon is 15, one teaspoon is 5,
    and one milliliter is 1. `g` is 1. Use
    `math/big.Rat`. Do not add a dependency. Grams
    are the number times the weight factor, or the
    number times the milliliter factor times the
    density, as one ratio, then one conversion to
    float64. Do not round before that conversion.
    Do not multiply two float64 values.
    `float64(240) * float64(1.03)` formats as
    `247.20000000000002` with `strconv.FormatFloat`
    precision `-1`, so Enter would not show `247.2`.
    The density in the product is the decimal
    spelling the field already reopens: `FormatFloat`
    with `'f'` and precision `-1`, parsed as a ratio.
    For a stored `1.03` that spelling is `1.03`, and
    one cup then stores the float64 of the decimal
    247.2, which reopens as `247.2`. On the Enter
    that first stores a density, that spelling is
    the reopen spelling of the float64 being saved,
    not the raw characters.
    `strconv.ParseFloat("0.10000000000000001", 64)`
    spells `0.1`. The ratio of that spelling times
    240 is the float64 literal 24, which reopens as
    `24`. The exact decimal of those raw characters
    times 240 spells `24.000000000000004`, and that
    float64 is not 24. Do not store it.
    `strconv.ParseFloat("+1.03", 64)` spells `1.03`,
    and one cup of that spelling is the float64
    literal 247.2. `ParseAmount` receives the
    float64 and spells it. The form must not turn
    the raw density text into the ratio. The same
    path stores one ounce as 28.349523125, one pound and
    `16 oz` as 453.59237, and `1/4 oz` as
    7.08738078125. Tests compare those reopen
    spellings and the float64 literal `247.2`. They
    do not compare a painted tenth to `28.35`, and
    they do not assert
    `float64(240) * float64(1.03) == 247.2`.
    `1/3` is the ratio 1/3, one `Float64`, and the
    `FormatFloat` spelling `0.3333333333333333`.
    Saving that spelling keeps the same float64.
    `TestMealAmountReopensStoredGrams` already saves
    the cup's bare `247.2`. Do not add a second test
    whose only job is that save.
-   After that one `Float64`, if the value is not
    finite or is not greater than zero, `ParseAmount`
    returns
    `amount: want a number greater than zero` and no
    grams. Grammar still fails first, so `0`, a blank,
    and `0 cup` never reach this check. There is no
    kitchen-fraction list and no maximum gram weight.
    The parser cannot be given a stub product. The
    real inputs are the text and the density.
    `1/4 ml` with density `math.SmallestNonzeroFloat64`
    is the ratio 1/4 times 1 times that density's
    `'f'` `-1` spelling, parsed as a `Rat`, then one
    `Float64`. That value is 0, so it is refused.
    `1/2 ml` with the same density is the ratio 1/2
    times that same spelling. Its one `Float64` equals
    `math.SmallestNonzeroFloat64`, so it saves. Do not
    compute it as
    `float64(1) / 2 * math.SmallestNonzeroFloat64`:
    that product is 0 and is not the ratio.
    `1/4` with the flag false still returns 0.25.
    `1 cup` with density `math.MaxFloat64` is the
    ratio 240 times that density's `'f'` `-1`
    spelling, and the one `Float64` is not finite, so
    it is refused. Do not define that product as
    `float64(240) * math.MaxFloat64`. Weight ignores
    density, so `1 g` on that food still returns 1.
    `1/18446744073709551615` with the flag false
    returns the float64 whose `'f'` `-1` spelling is
    `0.00000000000000000005421010862427522`.
    `18446744073709551615/1` returns the float64 whose
    spelling is `18446744073709552000`. The idle
    `%.1f` paints of those two are `0.0` and
    `18446744073709551616.0`. Both saves keep the
    grams when the spelling is saved unchanged.
    `1/3` is finite and must save. It is not either
    refusal. The screen tests set those densities
    through `AddFoodLine`'s `*float64`. They do not
    type a 300-digit decimal into the density field.
-   The line still uses `formatTenth`, which is `%.1f`.
    That print is half-even, the same rule the allowance
    spec already requires. `1.25` is `1.2`, `0.25` is
    `0.2`, and `28.25` is `28.2`. It is not half away
    from zero. `28.349523125` paints `28.3` because it
    is below the halfway point. `28.35` paints `28.4`.
    Do not change `formatTenth` to make the slicing
    note true, and do not store a rounded tenth. One
    ounce at 200 calories per 100 g paints
    `calories 56.7`. Scaling the painted `28.3` by 2
    would paint `56.6`. That gap is the check that the
    total uses the stored product. Do not change
    `formatTenth`. One ounce and a typed `28.3` both
    paint `28.3` through that `%.1f`. At 200 calories
    per 100 g the totals paint `56.7` and `56.6`. At
    100 calories per 100 g both paint `28.3`, because
    the ounce's residue is under half a tenth. The
    huge fraction in the product bullet paints
    `18446744073709551616.0` while Enter shows
    `18446744073709552000`. The tiny fraction paints
    `0.0`. Do not change `formatTenth` to make those
    match the reopen spelling.
-   The amount field keeps its input index. Change the
    label string to `amount`. Add a density input
    before it. Labels are the strings in Decision 8.
    `activeFields` for a new food includes the density.
    A known food is still name and amount. The line
    edit is still the amount only. `e` is the four
    numbers and the density. `beginMealGrams` still
    fills the field with `strconv.FormatFloat` of the
    stored grams, format `'f'` and precision `-1`.
    `beginMealFood` uses that same spelling when
    `HasDensity` is true, and an empty field when it
    is false. A cold line edit is that fill of the
    stored grams. A cold `e` is that fill of the four
    numbers and the density. The widgets the form
    does not show stay empty and are not checked.
-   Tab on the add form trims the name and rebuilds
    the inputs from that name. Clear the four
    nutrient drafts and the density draft. Keep the
    amount draft, including `1 cup`. A stored name
    keeps name and amount only. Drop the density
    input and discard its text. Do not write that
    text onto the stored food. A volume in the
    surviving amount uses that food's stored
    density. `Oats` at density `1` stores 240 g for
    that cup, not 247.2, and the density stays `1`.
    A new name gets the four nutrients empty, an
    empty density, and the surviving amount. Do not
    copy the previous food's calories, macros, or
    density into those fields. Enter when the
    trimmed name does not match the form Tab last
    built returns `name: press tab` and does not
    call the store. Density is written only by Enter
    on that built form, or by `e`.
-   On one Enter, walk only the fields on the form,
    in the order calories, protein, carbohydrate,
    fat, density, then amount. Skip a field the form
    does not show. A hidden field is not a failure
    and does not take focus. A line edit checks the
    amount only. `e` checks the four numbers, then
    density, and does not check the amount. The
    new-food add form shows every named field, and
    S25 is that form. Show the first error and focus
    that field. Do not collect later errors. With
    `track_macros` off, a blank macro is zero and is
    not an error. A blank density is nil and is not
    an error, so a later field can still fail.
    Density `0` is
    `density: want a number greater than zero` and
    stops before the amount, including when the amount
    is `1 cup`. S14's blank protein, with tracking on,
    has calories `100`, carbohydrate `0`, fat `0`, a
    blank density, and amount `100`, so the error is
    `protein: want a number` and the focus is
    `protein per 100 g`. The focused row is the `>`
    prefix on that label.
-   Parse density in the form, not in `ParseAmount`.
    Blank after trim is nil. A positive nutrient-shaped
    decimal is the pointer. Anything else, including
    zero, a leading `-`, a slash, and a unit, is
    `density: want a number greater than zero`. If a
    shown field among the four numbers or the
    density fails, do not call the store. A hidden
    field does not fail this check. Unchanged Enter
    on `e`, with an empty density, still saves the
    four numbers and writes NULL. That is what keeps
    `TestMealFoodEditRoundTrip` green.
-   The amount error strings are exactly
    `amount: want a number greater than zero` and
    `amount: this food has no density`. The form still
    prefixes `error:`. A `loadedMsg` must not clear
    `mealUI.err` and must not close the form. Do not
    store that error on `a.err`. Esc from the form
    still cancels to the idle meal screen. Do not set
    `allowanceEscList` for this error. Tab and the
    next typed character already clear `mealUI.err`
    before they edit. That includes
    `density: want a number greater than zero`, not
    only the two amount errors. A `loadedMsg` must
    not clear that density error either, and it must
    not close the form. Esc from the density error
    still cancels and writes nothing. The reload
    beat starts while the error is showing. Tab has
    already cleared it, so a reload after Tab is not
    the beat that keeps the error.
-   Do not call `ParseAmount` from `parseAllowance` or
    from `parseMealNutrient`. The allowance token list
    stays a refusal list. `100 g` in calories stays a
    bad nutrient.
-   `display_unit` is not read on this path. An amount
    save does not write `daily_log` and does not change
    a chart caption.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestMealAmountBareGrams`, `TestParseAmountGrams` |
| FR2 | `TestMealOunceAndPound`, `TestMealAmountFraction`, `TestParseAmountWeight` |
| FR3 | `TestMealVolumeDensity`, `TestMealAmountSpacing`, `TestParseAmountVolume` |
| FR4 | `TestMealAmountRefused`, `TestMealVolumeNeedsDensity`, `TestParseAmountRefused` |
| FR5 | `TestMealAmountReopensStoredGrams` |
| FR6 | `TestMealDensityFieldOnAdd`, `TestMealDensityBlankZeroAndNegative`, `TestMealDensityIgnoresTrackMacros` |
| FR7 | `TestMealClearDensityKeepsGrams`, `TestMealChangeDensityKeepsGrams`, `TestMealFoodNumbersStillUpdateEveryDate` |
| FR8 | `TestMealAmountHelpAndKeys`, `TestMealAmountFieldIsText` |
| FR9 | `TestMealAmountErrorSurvivesReload`, `TestMealDensityErrorClears` |
| FR10 | `TestMealAmountIgnoresDisplayUnit`, `TestMealAmountLeavesTheWeightLog`, `TestMealDensityColumnLeavesExistingFoodsUnset`, `TestMealMigrationKeepsDailyLog` |
| FR11 | `TestMealAmountDocs` |
| FR12 | `TestMealOunceAliasesTypedTenth` |
| FR13 | `TestParseAmountThird`, `TestMealThirdOfAGram` |
| FR14 | `TestMealTabRebuildsDensity` |
| FR15 | `TestParseAmountFiniteProduct`, `TestMealNonFiniteProductRefused` |
| FR16 | `TestMealFirstFieldError`, `TestMealDensityIgnoresTrackMacros` |
| FR17 | `TestMealColdEditSavesVisibleFields` |
| FR18 | `TestParseAmountDensitySpelling`, `TestMealDensitySpellingOnCreate` |

These revisions replace the stale meal-day rules for
this field. Meal-day Decision 16, S24, S25, and S31,
the `grams` label, and
`grams: want a number greater than zero` are not in
force here. The optional-macros sentence that a stored
name asks only for grams is not in force here. The new
string replaces `grams:` in these tests. Do not edit
the meal-day spec
[`2026-10-30-meal-day.md`](../specs/2026-10-30-meal-day.md),
the meal-day plan
[`2026-10-30-meal-day.md`](2026-10-30-meal-day.md),
or the optional-macros spec
[`2026-11-14-optional-food-macros.md`](../specs/2026-11-14-optional-food-macros.md).

Revise these existing tests. A label assertion changes
in the failing-test step, while the field is still
`grams`, so that assertion is red. Do not insert the
density tab in that step. Insert it in the same change
that adds the field, including the extra tab inside
`saveNew`. A tab added early types the amount into the
wrong field, and that failure is not the red this slice
wants.

-   `mealHelp` becomes the Decision 1 idle line.
    `TestMealHelpAndListHelp` expects that line and
    `> amount`. List help and `mealFormHelp` stay.
-   `TestMealAddTabOrderForANewFood` expects
    `density g/ml` between fat and `amount`, then the
    name. This revision is red until the field exists.
-   `TestMealAddExistingFoodAsksOnlyGrams` expects
    focus `amount` and still no per-100 g labels and
    no density label.
-   `TestMealBadGramsRefused` drops `1/4` from the bad
    list. `0`, `-1`, `1e2`, and blank stay, with
    `amount: want a number greater than zero`.
-   Replace `TestMealUnitTokensRefused`. Accepted
    tokens are no longer errors. The new tests cover
    them. Do not leave this test expecting `100 g` or
    `1 oz` to fail.
-   `TestMealRefusedGramsDoNotCreateAFood` expects the
    amount error and focus `amount`. Add the density
    tab when the field is added, so `0` is still the
    amount.
-   `TestMealDisplayOneDecimal` looks for the field
    `amount`. The `1.25` paint of `1.2` stays.
-   `TestMealEditGramsCancelAndRefuse`,
    `TestMealRemoveLineKeepsFood`,
    `TestMealEditFoodDoesNotRename`,
    `TestMealNameTrimAndCase`, and
    `TestMealEnterWaitsForTab` expect `amount` where
    they now expect `grams`. When the field is added,
    `TestMealNameTrimAndCase` and
    `TestMealEnterWaitsForTab` tab through a blank
    density before the amount.
-   `TestMealTypingClearsTheFormError` expects the
    first tab after a refused fat to land on
    `density g/ml`. One more tab reaches `amount`.
    That focus change is red until the field exists.
-   `TestMealFormErrorDoesNotFailTheList` expects
    `amount: want a number greater than zero` and
    still expects the reload to leave it up.
-   `TestMealMacrosOffShowsOptionalFields` tabs once
    more after a cleared fat so `100` is the amount
    and the density stays blank. It still shows the
    three macro fields when tracking is off.
-   `TestMealMacrosOffKeepsStoredMacros` tabs once
    more after fat before replacing calories with
    `442`. That value must not land in the density.
-   `TestMealAllowanceKey` expects the new idle help
    and focus `amount` on Enter. The allowance field
    and `TestMealBadAllowanceRefused` stay.
-   `TestMealMigrationKeepsDailyLog` expects a
    `density` column on `food`. An insert of the five
    old columns leaves that column NULL. `meal_line`
    columns stay `id`, `day`, `food_id`, and `grams`.
    `daily_log` columns, the table list, mode `0600`,
    and the single file stay.

These stay green and are not rewritten. They stay green
only if a blank density is a valid save and `saveNew`
leaves it blank once the field exists:

`TestMealEditFoodUpdatesEveryDate`,
`TestMealFoodEditRoundTrip`,
`TestMealEditFoodEscKeepsNumbers`,
`TestMealAddEscWritesNothing`,
`TestMealAddSavesScaledTotals`,
`TestMealDisplayUnitPoundsDoesNotScale`,
`TestMealEmptyDateIsZero`,
`TestMealIgnoresListKeys`,
`TestMealInFlightFormSaveKeepsTheMealScreen`,
`TestMealInFlightMonthSaveKeepsTheMealScreen`,
`TestMealFOpensTodayNotTheSelectedRow`,
`TestMealDocs`, `TestMealAllowanceDocs`,
`TestMealUnsetShowsNoAllowance`,
`TestMealBadAllowanceRefused`,
`TestProgramIsModelOnly`,
`TestOpenCreatesPrivateFile`,
`TestMealOpenMissingFileCreatesNothing`.

`TestMealDocs` stays green if the how-to still contains
allowance, volume, recipe, starter, later, per 100,
renaming, deleting, and the display unit, and if the
meal Enter row still contains `grams`. The new doc test
is the one that requires `amount` on that row and
forbids calling a volume later work.
`TestMealAllowanceDocs` stays green for the same reason.
It already forbids the old sentence that bundled the
allowance with a volume.

## Test list

Screen tests live in `internal/tui/mealamount_test.go`
unless a line below names another file. Build the app
from a temp-dir database: `dailylog.Open` that path,
`meal.Open` the same path, `New`, `UseMeals`, then
`load`. Do not use `:memory:`. Do not call `t.Parallel`
in a test that calls `freezeToday` or `t.Setenv`.
`press` drops the command `Update` returns. S17 must
not use `press` for the weight-log save. Call `Update`,
keep the command, open meals, then deliver the message
that command returns. Assert text with `visible`.
Today is `freezeMealToday` at 1990-11-10 12:00 UTC.
Parser tests live in `internal/meal/amount_test.go` and
call `ParseAmount`. They may use `t.Parallel`. Compare
floats to the literals in the algorithm notes, and
compare field text to the `FormatFloat` spelling those
notes name. Tests use `testing`. They may import `math`
and `strconv` for the cases that name
`math.MaxFloat64`, `math.SmallestNonzeroFloat64`, and
`FormatFloat`. Do not import another test framework.

-   `TestParseAmountGrams` — FR1: `100`, `100 g`,
    `100g`, `+100`, `.5`, and `  100 G  ` return 100
    or 0.5 with the density flag false. `100.0`
    returns 100.
-   `TestMealAmountBareGrams` — S1: label `amount`,
    idle `100.0` for `100`, `100 g`, `100g`, `100.0`,
    and `+100`, and each of those reopens `100`, not
    `100.0`. `.5` paints `0.5` and reopens `0.5`.
    The known-food form has no
    calorie label and no density label. The idle
    help is Decision 1.
-   `TestParseAmountWeight` — FR2: `1 oz` is
    28.349523125, `1 lb` and `16 oz` are 453.59237
    and equal each other, `1/4 oz` is
    7.08738078125, `1/4` is 0.25, `3/4` is 0.75,
    `2/4` is 0.5. The flag false still accepts
    these.
-   `TestMealOunceAndPound` — S2: `1 oz` at 200
    calories per 100 g paints `28.3` and
    `calories 56.7`, not `28.35`, `28.4`, `oz`, or
    `calories 56.6`. The field shows
    `28.349523125`. `1 lb` and `16 oz` paint
    `453.6` and reopen `453.59237`.
-   `TestMealAmountFraction` — S3: `1/4` paints
    `0.2` and reopens `0.25`. `3/4` paints `0.8`
    and reopens `0.75`. `2/4` paints `0.5`.
    `1/4 oz` paints `7.1` and reopens
    `7.08738078125`.
-   `TestParseAmountVolume` — FR3: density 1.03,
    `1 cup` and `240 ml` return the same grams and
    that value equals the float64 literal 247.2, not
    `float64(240) * float64(1.03)`. Density 1: cup,
    `240 ml`, and `8 fl oz` are 240; fl oz, tbsp, tsp,
    and ml are 30, 15, 5, and 1; `1/4 cup` is 60;
    `2 CUP` is 480. `1 fl oz`, `1fl oz`, `1  fl oz`,
    and `1 FL OZ` are 30. The flag false on `1 cup`
    is the no-density error and not a gram value.
-   `TestMealVolumeDensity` — S4: a new food, density
    `1.03`, amount `1 cup`, paints `247.2` and
    `calories 247.2`, and reopens `247.2`. `240 ml`
    matches it. Density `1` paints the S4 tenths,
    and `1/4 cup` paints `60.0` and reopens `60`.
-   `TestMealAmountSpacing` — S5: the five fl-oz
    spellings paint `30.0` and reopen `30`. `1g`
    and `1 G` paint `1.0`. `2 CUP` paints `480.0`.
    `1 fl  oz` stays open with the greater-than-zero
    error and writes no line.
-   `TestParseAmountRefused` — FR4: every S6 input
    returns `amount: want a number greater than zero`
    and not the no-density error, including when the
    flag is true and the density is 1.03. `1 fl  oz`
    is the same error. `1 cup` with the flag false
    returns the no-density error.
-   `TestMealAmountRefused` — S6: from a 100 g line
    at density `1.03`, each S6 input stays open with
    the greater-than-zero error and not `no density`.
    Esc leaves the one line at `100.0` and the
    density at `1.03`.
-   `TestMealVolumeNeedsDensity` — S7: `1 cup`,
    `1 CUP`, `1 fl oz`, `1 tbsp`, `1 tsp`, and
    `240 ml` report no density and write nothing.
    The density field stays empty, not `0`. Then
    `1 oz`, `1 lb`, `1 g`, and `50` paint `28.3`,
    `453.6`, `1.0`, and `50.0`.
-   `TestMealAmountReopensStoredGrams` — S8: `1 oz`
    reopens `28.349523125`, and unchanged Enter
    keeps the paint `28.3`. A cup at density `1.03`
    reopens `247.2`. Saving that bare text keeps
    it. Replacing it with `1 oz` paints `28.3`,
    reopens `28.349523125`, and leaves the density
    `1.03`.
-   `TestMealDensityFieldOnAdd` — S9: the new-food
    tab order, an empty density, nothing stored by
    Tab, then calories `100`, macros `0`, density
    `1.03`, and `1 cup` paint `247.2`. `e` shows
    `1.03` and does not show `amount`. The known
    name asks only for the amount.
-   `TestMealDensityBlankZeroAndNegative` — S10:
    empty density, not `0`. `0`, `0.0`, `-1`, `-0`,
    `1 g`, `1e2`, `inf`, `NaN`, and `1/4` use the
    density error and leave none. `1.03` reopens
    `1.03`. Unchanged Enter keeps it. `0` then
    leaves it `1.03`. `+1.03` reopens `1.03`. `.5`
    reopens `0.5`. The 100 g line stays `100.0`.
-   `TestMealDensityIgnoresTrackMacros` — S14:
    omitted `track_macros` still shows the three
    macro fields and `density g/ml`. Blank macros
    and a blank density save `100` and store no
    density, and the idle screen hides the three
    totals. True tracking, with calories `100`,
    carbohydrate `0`, fat `0`, a blank density, and
    amount `100`, refuses a blank protein. Focus is
    `protein per 100 g`. The density error and the
    amount error are absent, and nothing is written.
    Macros `0` and a blank density save, and the
    screen shows `protein 0.0`.
-   `TestMealClearDensityKeepsGrams` — S11: a cup at
    density `1.03` still paints `247.2` after the
    density is cleared, and reopens `247.2`. `e`
    shows an empty density. A later `1 cup` is the
    no-density error and adds no line. A later
    `100` paints `100.0`. The cup line stays.
-   `TestMealChangeDensityKeepsGrams` — S12: cups on
    1990-11-10 and 1990-11-11 at density `1` paint
    `240.0` and reopen `240`. After the density
    becomes `1.03`, both old lines stay `240` and
    reopen `240`. A new cup on the 10th paints
    `247.2` and reopens `247.2`. The 11th stays
    `240.0`.
-   `TestMealFoodNumbersStillUpdateEveryDate` —
    S13: from the S12 lines, calories `200` show
    `calories 974.4` on the 10th and
    `calories 480.0` on the 11th. The gram paints
    stay `240.0` and `247.2`. The density stays
    `1.03`.
-   `TestMealAmountHelpAndKeys` — S15: Decision 1
    help. `a` opens the name. Enter opens `amount`
    on the stored grams. `e` shows `density g/ml`,
    not `amount`, and the form help names Tab.
    Replacing calories with `100 g` is
    `calories: want a number`, and the stored
    calories stay. `x` removes the line and keeps the
    food. `b` then `1 cup` is
    `allowance: want a number` and not `no density`.
    `[`, `]`, Esc, and idle `f` stay. The list `f`
    is still the door. The allowance help does not
    name Tab.
-   `TestMealAmountFieldIsText` — S16: `q`, `a`,
    `f`, `b`, `[`, `]`, `o`, and `z` are characters
    on 1990-11-10. Esc writes nothing. `ctrl+c`
    from the field quits.
-   `TestMealAmountErrorSurvivesReload` — S17: the
    greater-than-zero error survives the reload.
    Tab clears it and does not save. The next
    character clears a fresh refusal. The
    no-density error also survives. Esc returns to
    the idle meal screen and the previous grams,
    not to a list that offers only `q quit`. A
    further Esc shows the ordinary list help.
-   `TestMealAmountIgnoresDisplayUnit` — S18:
    `display_unit` pounds, `1 oz` at 200 calories
    per 100 g, allowance `100`, shows `28.3`,
    `calories 56.7`, and `remaining 43.3`, not a
    body-weight conversion and not
    `Daily deficit`. `1 st` is refused and the
    ounce line stays.
-   `TestMealAmountLeavesTheWeightLog` — S19, the
    screen half: at least two weighed days, so both
    captions already contain `Daily deficit:`. Note
    those captions, save an amount, and require the
    same text. Do not assert a new deficit figure.
    Weight and trend text are unchanged. A day
    without a daily log row still has none.
-   `TestMealDensityColumnLeavesExistingFoodsUnset`
    — S19, the store half, in `amount_test.go`.
    `dailylog.Open` creates the file. Raw SQL
    creates the pre-density `food` and `meal_line`
    and inserts one food and one line. `meal.Open`
    then finds `density`, the inserted value is
    NULL, and the grams are unchanged. A volume
    parse of that food is refused. A weight parse
    still returns grams. `daily_log` columns are
    unchanged. Mode stays `0600`. No second file.
    Do not `os.Create` the database.
-   `TestMealAmountDocs` — S20: the meal Enter row
    says the selected line's amount, stored as
    grams. Collapse whitespace, then require the meal
    `e` row to contain both sentences below, as two
    sentences. Fail the row if one sentence contains
    both `density` and `every meal date`. Fail it if
    a sentence contains `density` and `rewrite` and
    does not contain `not`. That rejects a row that
    says a density edit rewrites grams or changes
    every date's grams. The how-to, after the same
    whitespace collapse, contains both sentences too.
    No new key. One list `f` door.

    ```text
    The four per-100 g numbers change every meal date that uses the food.
    Density is on that edit, and changing it does not rewrite stored grams, and the next volume uses the new density.
    ```

    The how-to names every token and the factors
    28.349523125, 16, 240, 30, 15, and 5 as factors
    in use. It does not call them proposals. It
    says a blank density means weight only, and that
    the line shows grams. The two sentences above are
    what say the density split. The how-to no longer
    says that a volume is later work.
    Recipes, the starter set, and renaming or
    deleting a food stay later. The how-to index
    still links the page. `docs/reference/config.md`
    gains no density key and no amount key.
-   `TestMealOunceAliasesTypedTenth` — S21, FR12, in
    `mealamount_test.go`. One food at 200 calories
    per 100 g, alone on its date: `1 oz` paints
    `28.3`, shows `calories 56.7`, and reopens
    `28.349523125`. Replacing that amount with `28.3`
    paints `28.3`, shows `calories 56.6`, and reopens
    `28.3`. A second food at 100 calories per 100 g,
    alone on its date: `1 oz` and then `28.3` each
    paint `28.3` and show `calories 28.3`. The typed
    line reopens `28.3`. Neither idle line shows
    `28.35` or `28.349523125`.
-   `TestParseAmountThird` — FR13, in
    `amount_test.go`. `1/3` with the flag false is
    the ratio 1/3, one `Float64`, not an error. Its
    `FormatFloat` `'f'` `-1` spelling is
    `0.3333333333333333`.
-   `TestMealThirdOfAGram` — S22, FR13. `1/3` on a
    food with no density paints `0.3`, does not show
    `1/3`, and reopens `0.3333333333333333`. Saving
    that text unchanged keeps the paint and the
    spelling.
-   `TestMealTabRebuildsDensity` — S23, FR14. Stored
    `Oats` has density `1` and 100 calories per
    100 g. `Oats` may be stored with `AddFoodLine`
    before `a`. The add form under test starts
    empty. On a new-name form, type `1.03` in
    density, set the amount to `1 cup`, change the
    name to `Oats`, and press Enter with no Tab
    after that change. The error is `name: press tab`
    and nothing is written. Tab then asks only for
    the amount, with no `density g/ml`, no `1.03`,
    and no `calories per 100 g`, protein per 100 g,
    carbohydrate per 100 g, or fat per 100 g label.
    Read the amount before changing it: it still
    shows `1 cup`. Enter on that cup saves `Oats`
    at 240 g, not 247.2. The line paints `240.0`
    and reopens `240`, not `247.2`. Density stays
    `1`. The four per-100 g numbers stay. The other
    way: the
    amount-only form for `Oats`, amount set to `40`,
    name changed to `Rye`, Enter without Tab, is
    `name: press tab` and stores no `Rye`. Tab then
    asks for the four numbers, an empty density, and
    the amount. Before typing over them, the four
    input values are empty, not `0` and not `100`.
    The label `per 100 g` is not that value. The
    density input is empty and shows neither `1`
    nor `1.03`. The amount input still shows `40`.
    Then numbers `1`, density `1.5`, and
    amount `100` store `Rye` at density `1.5` and
    leave `Oats` at density `1`.
-   `TestParseAmountFiniteProduct` — FR15, in
    `amount_test.go`. The ratios and the spellings
    are the product bullet above. `1/3` and both
    64-bit-end fractions save. `1/2 ml` at
    `math.SmallestNonzeroFloat64` equals that
    density. `1/4 ml` at that density is the
    greater-than-zero error and not a gram value.
    `1 cup` at `math.MaxFloat64` is that same error.
    `1 g` with that density returns 1. `1/4` with
    the flag false returns 0.25. Do not pass a stub
    product into `ParseAmount`.
-   `TestMealNonFiniteProductRefused` — S24, FR15.
    No-density food: `18446744073709551615/1` saves,
    paints `18446744073709551616.0`, reopens
    `18446744073709552000`, and saving that text
    keeps it. `1/18446744073709551615` saves, paints
    `0.0`, reopens
    `0.00000000000000000005421010862427522`, and
    saving that text keeps it. Smallest-density food,
    density set by `AddFoodLine`: `1/2 ml` saves as
    `math.SmallestNonzeroFloat64` and reopens that
    value's `'f'` `-1` spelling; saving it keeps the
    grams. `1/4 ml` stays open with the
    greater-than-zero error, not `no density`, and
    adds no line. Largest-density food, set the same
    way: `1 cup` is that amount error and adds no
    line. `1 g` then paints `1.0` and reopens `1`.
-   `TestMealFirstFieldError` — S25, FR16. The form
    is already built for the new name, and every
    named field is on screen. This is not the cold
    line edit and it is not `e`. Tracking on:
    blank calories, macros `0`, blank density, and
    amount `100` focus calories per 100 g and show
    `calories: want a number`, and write nothing.
    Blank calories, blank protein, the other macros
    `0`, density `0`, and amount `0` still focus
    calories and do not show `protein:`, `density:`,
    or `amount:`. Tracking omitted: calories `100`,
    blank macros, density `0`, and amount `1 cup`
    focus `density g/ml`, show the density error, and
    do not show `amount:` or `protein:`. Clearing
    that density and setting the amount to `0` focuses
    `amount` and shows the amount error, not
    `protein:`. Nothing is written in any of these.
-   `TestMealColdEditSavesVisibleFields` — S26,
    FR17, in `mealamount_test.go`. Do not press `a`
    and do not call `saveNew`. Before `New`,
    `AddFoodLine` stores `Oats` at 200 calories,
    protein 10, carbohydrate 30, fat 4, density
    `1.03`, and one line of 100 g. Open the meal
    screen from idle. Enter shows the amount field
    as `100` and does not show `calories per 100 g`,
    protein per 100 g, carbohydrate per 100 g, fat
    per 100 g, or `density g/ml`. Idle
    `calories 200.0` is not one of those labels.
    Unchanged Enter keeps the paint `100.0` and the
    grams at 100, and does not show `error:`. Enter
    again shows `100`. Esc, then `e`, shows the field
    values `200`, `10`, `30`, `4`, and `1.03`, and
    does not show `amount`. Read those inputs, not
    the idle total: `calories 200.0` is not the
    calorie field. Unchanged Enter keeps the grams
    at 100 and does not show `error:`. `e` again
    shows those same five spellings. Empty hidden
    widgets do not block either save.
-   `TestParseAmountDensitySpelling` — FR18, in
    `amount_test.go`. The density argument is the
    float64 from
    `strconv.ParseFloat("0.10000000000000001", 64)`.
    Its `'f'` `-1` spelling is `0.1`. `1 cup` with
    the flag true returns the float64 literal 24,
    whose spelling is `24`. It does not return the
    one `Float64` of the exact decimal
    `0.10000000000000001` times 240. That value
    spells `24.000000000000004` and is not 24.
    A second call uses
    `strconv.ParseFloat("+1.03", 64)`, whose
    spelling is `1.03`. `1 cup` returns the float64
    literal 247.2, whose spelling is `247.2`. Do
    not build the ratio from the raw characters.
    Keep `math/big.Rat` and one conversion to
    float64.
-   `TestMealDensitySpellingOnCreate` — S27, FR18.
    A new food, 100 calories per 100 g, blank
    macros, density text
    `0.10000000000000001`, and amount `1 cup`, on
    one Enter. The line paints `24.0` and reopens
    `24`, not `24.000000000000004`. `e` shows
    `0.1`, not the raw text. A later known `1 cup`
    paints `24.0` and reopens `24`. Another new
    food, same calories, blank macros, density
    `+1.03`, and `1 cup` on that same Enter, paints
    `247.2`, reopens `247.2`, and `e` shows `1.03`.
    Those grams equal one cup at density `1.03`.
    Do not change `TestMealVolumeDensity`: typed
    `1.03` and `1 cup` still store 247.2.
-   `TestMealDensityErrorClears` — S28, FR9. A food
    with no density and one line of 100 g. `e`,
    density `0`, Enter: the editor stays open on
    `density: want a number greater than zero`.
    Tab from that error clears it and does not
    save. The stored density stays unset. Esc, `e`,
    density `0`, Enter, so the error is showing,
    then type `1`. The error is gone and nothing is
    saved. That typed `1` is not a save. The stored
    density stays unset. Esc, `e`, density `0`,
    Enter, so the error is showing again, then
    deliver a weight-log reload the way S17 does,
    not through `press`. The error stays, the
    editor stays open, and the screen stays the
    meal screen. Esc returns to the idle line at
    `100.0` with the density still unset, not to a
    list that offers only `q quit`. Do not reload
    after Tab and expect the error.
