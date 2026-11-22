# Plan — Meal allowance

**Spec**:
[`docs/superpowers/specs/2026-11-20-meal-allowance.md`](../specs/2026-11-20-meal-allowance.md)
**Slice**: S2 of
[`meal-planning.md`](../slices/meal-planning.md)
**Issue**: [#140](https://github.com/jglueckstein/hdtools/issues/140)
**Follows**:
[`2026-10-30-meal-day.md`](../plans/2026-10-30-meal-day.md)
(approved, shipped). Do not edit that plan or its spec.
Do not edit
[`2026-11-14-optional-food-macros.md`](../specs/2026-11-14-optional-food-macros.md).
**Status**: approved

No production code until the spec's scenarios exist as
failing tests.

S2 is a calorie allowance for one meal date on the
shipped meal screen. The idle screen already shows the
calorie total either way, and the three macro totals
only when `track_macros` is true. Do not restore four
totals on an unset day. Do not change the food form.
Volume, recipes, the starter set, and renaming or
deleting a food stay in `idea.md`. Do not implement
them. Do not edit the slicing record.

`b` is the allowance key. The tests and the docs use
that letter. Do not invent another key.

A test that does not compile is not red. The red step
may add the store operations the new tests call, and a
`b` branch that does not save and does not paint
`no allowance` or `remaining`. Those additions must not
pass a behavioural test. No successful save in that
step.

## Module structure

| File | Why |
| --- | --- |
| `internal/meal/meal.go` | The new table lives in the existing migrate, in the same SQLite file as `daily_log`. Read, replace, and clear one calorie number for a civil date. Update the preamble that says this package does not set an allowance. Still no screen, no display unit, no config, no density, and no recipe. |
| `internal/meal/meal_test.go` | Revise `TestMealMigrationKeepsDailyLog`. The table list gains `meal_allowance` only. `daily_log` columns stay. Mode stays `0600`. Opening still does not create a missing file or a second file. |
| `internal/tui/meal.go` | The idle comparison and the `b` case. Loading a meal date loads that date's allowance. Update the preamble that says this file does not set an allowance. `meal.go` is 290 lines. If the editor would put it past 500, move the one-field editor to `internal/tui/mealallow.go` in this package first. Do not pre-split. |
| `internal/tui/mealform.go` | Do not edit. The food form, Tab, and a blank macro stay there. The allowance editor must not call `updateMealForm`. |
| `internal/tui/meal_test.go` | Revise `TestMealEmptyDateIsZero` and `TestMealHelpAndListHelp` so they match this spec. Do not rewrite the food-form tests. |
| `internal/tui/mealallow_test.go` | New screen scenarios. Package `tui`, so the meal helpers stay in the package. |
| `internal/tui/app.go` | No new case. The file is 459 lines. `updateMeal` already owns meal keys. Do not import `database/sql` or a SQLite driver. |
| `cmd/hdtools/main.go` | No change. `meal.Open` already migrates the daily-log path. PDF flags still return before that open. |
| `docs/reference/keys.md` | FR9. One meal row for `b`. `f` stays the only door. |
| `docs/how-to/meal-day.md` | FR9. The allowance step. Drop the sentence that an allowance is a later slice. Keep volume, recipes, the starter set, and renaming or deleting a food as later work. |
| `docs/how-to/README.md` | No new page. The meal-day link stays. |
| `docs/reference/config.md` | No allowance key. |
| `idea.md` | Already points at this spec and at issue #140. Do not rewrite the other sections. |
| `CHANGELOG.md` | When the behaviour ships, one Unreleased bullet. Not in the failing-test step. |

`internal/meal` must not import `internal/tui`,
`internal/dailylog`, `internal/config`, or
`internal/units`. The TUI calls the meal store the way
it already calls `Lines`. `cmd/hdtools` does not parse
meal keys and does not embed the new SQL.

Do not edit `internal/dailylog`, `internal/units`,
`internal/tui/chart.go`, or `internal/tui/longchart.go`.
Do not call `ApplyTrend` from an allowance save. Do not
add a `[colors]` role. Do not pass `AltScreen` or a
mouse option.

New text in `meal.go` keeps a literate preamble: one
calorie allowance per civil date, in this file, not a
weight-log column and not a config key. Returned errors
use `%w` and a context. Do not bare-return `err`.

## Algorithm notes

-   The table is created with the existing
    `CREATE TABLE IF NOT EXISTS` migrate. Do not
    `ALTER` `daily_log`, `food`, or `meal_line`. Do
    not `os.Create` the database. `dailylog.Open`
    remains the creator, so mode `0600` stays that
    file. A missing path still returns an error.

    ```sql
    CREATE TABLE IF NOT EXISTS meal_allowance (
        day TEXT PRIMARY KEY,
        calories REAL NOT NULL
    );
    ```

    `day` is `YYYY-MM-DD`, the same string `meal_line`
    uses. One row per civil date. No row means unset.
    A row whose `calories` is `0` is a set zero. Do
    not store NULL. Do not use a sentinel in place of
    a missing row. Clear is `DELETE` for that day. A
    second save replaces that day's number. There is
    no foreign key to `daily_log`: a meal date need
    not have a weight-log row. Leave the foreign-key
    pragma on. Do not copy another day's row when `[`
    or `]` moves the date.
-   Store operations, called by the TUI:
    `Allowance(day) (float64, bool, error)` with the
    bool false when no row exists, including when the
    numeric value is zero; `SetAllowance(day, float64)`
    replaces that date; `ClearAllowance(day)` removes
    the row. A stored zero returns true and `0`. The
    TUI must not paint a false bool as remaining
    `0.0`.
-   Remaining is the stored allowance minus the
    unrounded sum already used for the calorie line
    (`calories * grams / 100` across the lines), then
    `formatTenth`. Do not subtract the painted
    `calories` text. `formatTenth` is `%.1f`, and a
    printed `-0.0` becomes `0.0`, the same way the
    weight column does. That print is half-even:
    `1.25` is `1.2` and `0.25` is `0.2`. A difference
    whose tenth rounds to zero prints
    `remaining 0.0`. That covers an exact zero, a
    surplus smaller than half a tenth, and an overage
    smaller than half a tenth. Allowance `0` against
    an unrounded total of `0.04` paints
    `remaining 0.0`, not `no allowance`, and the view
    has no `-0.0`. The S6 case is the check that the
    painted tenth was not subtracted: calories `1.25`
    per 100 g and 100 g paint `calories 1.2`;
    allowance `2.5` paints `remaining 1.2`, not
    `remaining 1.3`.
-   The idle comparison is one line after the totals
    that are already painted. Unset prints
    `no allowance`. Set prints `remaining`, a space,
    and the tenth. Do not also print the allowance as
    its own idle total. Do not paint a macro
    allowance. `track_macros` still gates only the
    three day totals.
-   The editor is one field, not one of the six food
    fields. It replaces the idle body. The line list,
    the calorie total, the macro totals, and the
    comparison line are not shown while it is open.
    The heading keeps the date. Handle it before the
    non-idle path calls `updateMealForm`. Unset shows
    an empty field, not `0`. Set shows
    `strconv.FormatFloat` with `'f'`
    and precision `-1`, so `1.25` stays `1.25`, `500`
    stays `500`, and `0` stays `0`. Enter on that
    text unchanged writes the same number. Esc drops
    the field and does not call the store.
-   Parse after `strings.TrimSpace`. Empty clears.
    Otherwise use the nutrient shape: an optional
    `+`, then digits with at most one dot. Refuse a
    leading `-` before the parse, so `-0` and `-0.0`
    are the negative error and are not stored as
    zero. Refuse `1e2`, `inf`, `NaN`, a slash, a
    middle space, and a trailing unit token. Tokens
    are the gram set plus `kcal` and `cal`: `g`,
    `kg`, `oz`, `lb`, `st`, `cup`, `fl oz`, `tbsp`,
    `tsp`, `ml`, `kcal`, and `cal`. A refused Enter
    writes nothing. Negative copy is
    `allowance: want a number that is not negative`.
    Any other refusal is `allowance: want a number`.
    The screen still prefixes `error:`. The next
    typed character clears that error. Tab does not
    save, does not cancel, and does not walk fields.
    Do not require Tab before Enter.
-   While the field is open, `[`, `]`, `q`, `a`,
    `f`, and `b` are characters. They do not move
    the date, quit, or open a screen. `ctrl+c` is
    still handled before that, as it is for the food
    form. A weight-log `loadedMsg` must not clear
    the allowance error and must not close the
    editor. Do not store that error on `a.err`.
-   Idle help becomes the Decision 6 line. The food
    form keeps `mealEditHelp`, which names Tab. The
    allowance editor uses its own line, without Tab.
    List help is unchanged. `f` on the list stays
    the only door.
-   `display_unit` is not read on this path. Body
    weight on the list still converts. An allowance
    save does not write `daily_log` and does not
    change a chart caption.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestMealUnsetShowsNoAllowance`, `TestMealTrackedMacrosStillShowNoAllowance`, `TestMealFoodsSaveWhileUnset` |
| FR2 | `TestMealRemainingUnder`, `TestMealRemainingOver`, `TestMealRemainingExactZero`, `TestMealRemainingHalfEven`, `TestMealRemainingZeroTenth` |
| FR3 | `TestMealZeroAllowanceIsSet`, `TestMealAllowanceOnAnEmptyDay` |
| FR4 | `TestMealNegativeAllowanceRefused`, `TestMealBadAllowanceRefused`, `TestMealBlankAllowanceClears` |
| FR5 | `TestMealAllowanceDoesNotCarry`, `TestMealMigrationKeepsDailyLog` |
| FR6 | `TestMealAllowanceLeavesTheChartCaption`, `TestMealAllowanceIgnoresDisplayUnit` |
| FR7 | `TestMealInFlightFormSaveKeepsTheMealScreen`, `TestMealAllowanceErrorSurvivesReload` |
| FR8 | `TestMealAllowanceKey`, `TestMealAllowanceFieldIsText` |
| FR9 | `TestMealAllowanceDocs` |

Revise these existing tests. Do not leave them asserting
the pre-allowance screen:

-   `TestMealEmptyDateIsZero` expects `no allowance`
    beside the four `0.0` totals (`mealApp` sets
    `track_macros`), no `remaining`, no selection, and
    no starter name.
-   `TestMealHelpAndListHelp` expects the Decision 6
    idle line. List help and `mealFormHelp` stay.
-   `TestMealMigrationKeepsDailyLog` expects
    `meal_allowance` in the table list and still
    expects the old `daily_log` columns, mode `0600`,
    and no second file.

These stay green and are not rewritten:
`TestMealMacrosOffShowsOptionalFields`,
`TestMealMacrosOffKeepsStoredMacros`,
`TestMealIgnoresListKeys`,
`TestMealInFlightMonthSaveKeepsTheMealScreen`,
`TestMealFormErrorDoesNotFailTheList`,
`TestMealFOpensTodayNotTheSelectedRow`,
`TestMealDisplayUnitPoundsDoesNotScale`,
`TestMealDocs`, `TestProgramIsModelOnly`,
`TestOpenCreatesPrivateFile`,
`TestMealOpenMissingFileCreatesNothing`.

`TestMealDocs` stays green only if the how-to still
contains the words that test already requires, including
allowance, volume, recipe, starter, later, and
display unit. The new doc test is the one that requires
the `b` row and forbids calling the allowance a later
slice. It also requires the one-decimal remaining
sentence, and that `2.5` against a total of `1.25`
shows `1.2`.

## Test list

Screen tests live in `internal/tui/mealallow_test.go`
unless a line below names an existing test. Build the
app from a temp-dir database: `dailylog.Open` that path,
`meal.Open` the same path, `New`, `UseMeals`, then
`load`. Do not use `:memory:`. Do not call `t.Parallel`
in a test that calls `freezeToday` or `t.Setenv`.
`press` drops the command `Update` returns. S16 and S20
must not use `press` for the weight-log save. Call
`Update`, keep the command, open meals, then deliver
the message that command returns. Assert text with
`visible`. Today is `freezeMealToday` at 1990-11-10
12:00 UTC.

-   `TestMealUnsetShowsNoAllowance` — S1: default
    config, empty 1990-11-10 shows `calories 0.0` and
    `no allowance`, and not `protein`,
    `carbohydrate`, `fat`, or `remaining`. After 100 g
    of `Oats` at 200 calories per 100 g with blank
    macros, `calories 200.0` and `no allowance`, still
    without those three words or `remaining`.
-   `TestMealTrackedMacrosStillShowNoAllowance` — S2:
    `track_macros` true. Empty day shows the four
    `0.0` totals and `no allowance`, not `remaining`.
    After 100 g at 200, 10, 20, and 5 per 100 g, those
    totals are `200.0`, `10.0`, `20.0`, and `5.0`,
    still `no allowance` and not `remaining`.
-   `TestMealRemainingUnder` — S3: calories `200.0`,
    `b`, ` 500 `, Enter with no Tab, shows
    `remaining 300.0` and not `no allowance`. `+500`
    leaves that remaining.
-   `TestMealRemainingOver` — S4: allowance `100` on
    calories `200.0` shows `remaining -100.0`.
-   `TestMealRemainingExactZero` — S5: allowance
    `200` on calories `200.0` shows `remaining 0.0`
    and the view has no `-0.0`.
-   `TestMealRemainingHalfEven` — S6: allowance
    `1.25` with no foods shows `remaining 1.2`, and
    the field shows `1.25`. Unchanged Enter keeps
    `1.25`. `0.25` shows `remaining 0.2`. On the next
    date, 100 g at `1.25` calories per 100 g shows
    `calories 1.2`, and allowance `2.5` shows
    `remaining 1.2`, not `remaining 1.3`. The first
    date still shows `remaining 0.2`.
-   `TestMealRemainingZeroTenth` — S21: 100 g at
    `0.04` calories per 100 g, so the unrounded total
    is `0.04`. Allowance `0` shows `calories 0.0` and
    `remaining 0.0`, not `no allowance`, and the view
    has no `-0.0`.
-   `TestMealZeroAllowanceIsSet` — S7: calories
    `40.0` and allowance `0`, `0.0`, or `+0` shows
    `remaining -40.0`. The field shows `0`. Removing
    the line shows `calories 0.0` and
    `remaining 0.0`, not `no allowance`, and the view
    has no `-0.0`.
-   `TestMealNegativeAllowanceRefused` — S8: from
    remaining `300.0`, `-1` stays on the editor with
    `allowance: want a number that is not negative`.
    The next character clears the error. Esc returns
    to `remaining 300.0`. From unset, `-0` is the
    same error, and Esc shows `no allowance`.
-   `TestMealBadAllowanceRefused` — S9: from
    allowance `500`, each of `2000 kcal`,
    `2000 cal`, `2000 g`, `2000g`, `1 kg`, `4 oz`,
    `1 lb`, `1 st`, `1 cup`, `8 fl oz`, `1 tbsp`,
    `1 tsp`, `250 ml`, `1/4`, `1e2`, `inf`, and
    `NaN` stays open with `allowance: want a number`
    and not `not negative`. Esc still shows
    `remaining 300.0`.
-   `TestMealBlankAllowanceClears` — S10: clearing
    the field, or entering only spaces, shows
    `no allowance`, not `remaining` and not
    `want a number`. The food line remains. The next
    `b` opens empty.
-   `TestMealAllowanceOnAnEmptyDay` — S11: no lines,
    `calories 0.0` and `no allowance`. Allowance
    `1800` shows `remaining 1800.0`. `.5` shows
    `remaining 0.5`, and the field shows `0.5`.
-   `TestMealAllowanceDoesNotCarry` — S12: `1800` on
    1990-11-10 does not appear on 1990-11-11, which
    shows `no allowance`. `[` restores
    `remaining 1800.0`. Quit and open shows the same
    pair. The daily log row, its columns, mode
    `0600`, and the single database file stay.
-   `TestMealAllowanceLeavesTheChartCaption` — S13:
    at least two weighed days, so both captions
    already contain `Daily deficit:`. Note those
    captions, save an allowance, and require the same
    text. Do not assert a new deficit figure. The
    meal screen does not contain `Daily deficit`.
    Weight and trend text are unchanged. A day
    without a daily log row still has none.
-   `TestMealAllowanceIgnoresDisplayUnit` — S14:
    `display_unit` pounds, 100 g at 100 calories per
    100 g, allowance `2000`, shows `100.0`,
    `calories 100.0`, and `remaining 1900.0`, not the
    pound conversion of those figures.
-   `TestMealFoodsSaveWhileUnset` — S15: saving a
    line while unset shows the line and its calorie
    total, still `no allowance`, not `remaining`, and
    `b` opens empty.
-   `TestMealInFlightFormSaveKeepsTheMealScreen` —
    S16, existing test, stays green: a day-form save
    that finishes after `f` leaves the meal screen
    up, and Esc returns to 1990-11-01.
-   `TestMealAllowanceKey` — S17: idle help is the
    Decision 6 line. `b` on an unset day opens an
    empty `allowance` field and the editor help, and
    the screen does not show `calories`,
    `no allowance`, or `remaining`. Esc writes
    nothing. A stored `500` shows `500`, not
    `500.0`. While that field is open the screen does
    not show `calories`, `no allowance`, or
    `remaining`. Enter without Tab keeps `500`.
    Idle `a`, Enter, and `f` stay as they are. The
    add form help still names Tab.
-   `TestMealAllowanceFieldIsText` — S18: `q`, `a`,
    `f`, `b`, `[`, and `]` are characters on
    1990-11-10. Esc writes nothing. `ctrl+c` from
    the field quits.
-   `TestMealAllowanceDocs` — S19: `keys.md` has a
    meal `b` row for that day's calorie allowance
    and still has one list `f` door. The meal how-to
    describes `b`, Enter, Esc, blank, and
    `no allowance`. It says remaining is the allowance
    minus the unrounded calorie total, shown to one
    decimal the same way as the calorie line, and that
    `2.5` against a total of `1.25` shows `1.2`. It
    does not tell the reader to subtract the painted
    calorie text. It names Daily Deficit,
    `config.toml`, and `display_unit`. It no longer
    says that an allowance, a volume, a recipe, and a
    starter set are later slices. Volume, recipes, the
    starter set, and renaming or deleting a food stay
    later. The how-to index still links the page.
    `docs/reference/config.md` gains no allowance key.
-   `TestMealAllowanceErrorSurvivesReload` — S20: a
    negative-allowance error is still showing after
    the weight-log reload, the editor is still open,
    and Esc shows the ordinary list help.
