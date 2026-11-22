# Meal allowance

**Date**: 2026-11-20
**Status**: approved
**Issue**: [#140](https://github.com/jglueckstein/hdtools/issues/140)
**Backlog**: [`idea.md`](../../../idea.md) (Meal Planning)
**Slice**: S2 of
[`meal-planning.md`](../slices/meal-planning.md)
**Follows**:
[`2026-10-30-meal-day.md`](2026-10-30-meal-day.md)
(approved, shipped) and
[`2026-11-14-optional-food-macros.md`](2026-11-14-optional-food-macros.md).
This file does not amend those specs. The meal screen they
describe is the screen this slice extends.

The meal screen gains a calorie allowance for the meal date
on screen. The person types it. It is stored with that date.
Remaining calories are that allowance minus the day's calorie
total. An unset allowance is the words "no allowance", not
zero and not the chart's Daily Deficit.

The slice text that an unset day shows the four totals is
stale. The idle screen shows the calorie total either way.
It shows the protein, carbohydrate, and fat day totals only
when `track_macros` is true. This slice does not change the
food form.

No production code until these scenarios exist as failing
tests.

## Allowance key

`b` on the idle meal screen opens the allowance for the
meal date on screen. The idle help, the scenarios, FR8,
FR9, the key reference, and the how-to use that letter.

Idle `n`, `m`, `c`, `l`, `p`, `P`, `t`, and `f` do
nothing. `a`, Enter, `e`, `x`, `[`, `]`, Esc, and `q`
stay what they are. `b` is not a second way to open
meals. `f` on the daily list remains the only door. The
key table does not gain a second door.

## User stories

### US1 — Set a calorie allowance for this date

As a person logging a meal, I want to type a calorie
allowance for this meal date and see the calories that
remain, so that the day's total has a target I chose.

### US2 — Leave the target unset

As a person who has not chosen an allowance, I want the
day to say "no allowance" and still show the calorie
total, so that an empty target is not zero and is not
the chart's Daily Deficit.

### US3 — Keep each date's own number

As a person moving between meal dates, I want each date
to keep only the allowance I typed for it, so that
another day does not inherit the number.

### US4 — Leave the weight log alone

As a person who also keeps a weight log, I want saving
an allowance to leave Daily Deficit, the charts, and
the display unit alone, so that a meal target is not a
body weight.

## Decisions

1.  **The allowance belongs to one meal date.** The person
    types it. It is stored with that date in the same
    SQLite file as the meals. It is not a key in
    `config.toml`, not a column on the daily log, and not
    the chart's Daily Deficit. Another meal date does not
    inherit it. `[` and `]` show the date they land on,
    including that date's own allowance. A meal date with
    no weight-log row can still have one.
2.  **Only calories have an allowance.** Protein,
    carbohydrate, and fat do not. The idle screen shows
    the calorie total either way. It shows those three
    day totals only when `track_macros` is true. Omitted
    or false leaves them hidden. The food form is
    unchanged: it still asks for the three, and a blank
    one is still stored as zero when tracking is off.
3.  **Unset is the words `no allowance`.** The screen
    does not show remaining `0.0`, does not treat the
    calorie total as over, and does not copy Daily
    Deficit, another day's allowance, or a config
    default. Foods still save. Clearing the field
    returns to this state. The idle screen does not
    also print the allowance as its own total. The
    comparison line is either `no allowance` or
    `remaining` and its number, after the totals that
    are shown.
4.  **Remaining is the allowance minus the day's
    calorie total.** The subtraction uses the unrounded
    total, the same sum the calorie line is drawn from,
    and then the calorie line's one-decimal half-even
    rule once. It does not subtract the painted calorie
    text. Under the allowance, remaining is positive.
    Over it, remaining is negative. "Over is negative"
    means the printed tenth has a minus. There is no
    separate overage word. A difference whose tenth
    rounds to zero prints `remaining 0.0`. That covers
    an exact zero, a surplus smaller than half a tenth,
    and an overage smaller than half a tenth. A printed
    `-0.0` becomes `0.0`, the same way the weight
    column does. A stored zero is set: remaining is
    minus the calorie total, not the unset state. A day
    with no foods can hold an allowance, and remaining
    is then that allowance. `1.25` prints `1.2`. `0.25`
    prints `0.2`. One hundred grams of a food at `1.25`
    calories per 100 g paints `calories 1.2`; an
    allowance of `2.5` on that day paints
    `remaining 1.2`, not `remaining 1.3`.
5.  **Blank clears. Zero is a number. A negative or a
    non-number is refused.** Whitespace at either end
    is removed first. Empty after that clear, including
    a field of only spaces, deletes the allowance and
    returns to Decision 3. It does not show
    `want a number`. That is the opposite of a food
    number, where blank means the number is missing.
    A stored zero stays set. The accepted shape is the
    food-nutrient shape: an optional leading `+`, then
    digits with at most one decimal point. `500`,
    ` 500 `, `+500`, `500.0`, `0`, `+0`, `0.0`, `.5`,
    and `1.25` are valid. A leading `-` is negative,
    including `-0`, `-0.0`, and `-1`, and is refused.
    It is not stored as zero. An exponent, `inf`,
    `NaN`, a slash fraction, a middle space, and a unit
    token are refused. Unit tokens include a trailing
    `g`, `kg`, `oz`, `lb`, `st`, `cup`, `fl oz`,
    `tbsp`, `tsp`, `ml`, `kcal`, and `cal`, with or
    without a space before the token. A refused Enter
    leaves the editor open, writes nothing, and the
    stored value stays as it was. A negative shows
    `error:` and
    `allowance: want a number that is not negative`.
    Any other refusal shows `error:` and
    `allowance: want a number`. The next typed
    character clears that error. A weight-log reload
    does not.
6.  **`b` on the idle meal screen opens the allowance.**
    `b` opens one allowance field for the current meal
    date. The heading keeps the date. The editor
    replaces the idle body. The line list, the calorie
    total, the macro totals, and the comparison line
    are not shown while the field is open. The field
    is labeled `allowance` and is focused, marked with
    `>`. Unset opens empty, not `0`. Set opens on the
    stored number, not the idle tenth: `1.25` shows
    `1.25`, and `0` shows `0`. Enter on that text
    unchanged keeps the stored number. Enter saves
    without Tab.
    Esc cancels and writes nothing. Tab is not
    required, does not save, does not cancel, and does
    not move to another field. Letters are text,
    including `q`, `a`, `f`, and `b`. `[` and `]` are
    text and do not change the meal date. `ctrl+c`
    still quits. The idle help line is:

    ```text
    a add   enter grams   e food   x remove   b allowance   [ ] day   esc list   q quit
    ```

    The allowance editor's help line is:

    ```text
    enter save   esc cancel
    ```

    The add form, the grams edit, and the food edit
    keep their help line, including Tab. Idle `a`,
    Enter, `e`, `x`, `[`, `]`, Esc, and `q` stay as
    they are. Idle `f` still does not open another
    screen.
7.  **Saving an allowance does not touch the weight
    log.** It does not change weight, sleep, steps,
    workout, note, or the trend text on the list. It
    does not change either chart caption. Daily Deficit
    stays the chart's number. The meal screen does not
    read `display_unit` and does not scale the
    allowance, the remaining line, or the grams. An
    in-flight weight-log save still does not take the
    meal screen, as in meal-day Decision 23. The list
    row Esc will restore stays the row selected when
    the meal screen opened. A refused allowance error
    is a meal error: the reload does not clear it, and
    Esc afterward shows the log, not the failed-load
    view that offers only `q quit`.

## Acceptance scenarios

Today in these scenarios is 10 November 1990 unless a
scenario says otherwise. The weight log, when it has
rows, has 1 November 1990 selected unless the scenario
says otherwise. `track_macros` is omitted unless the
scenario sets it. The default leaves that key unset.

### S1 — Unset shows the calorie total and no macros

**Given** `track_macros` is omitted
**And** the meal screen for an empty 1990-11-10
**And** no allowance is stored
**Then** the screen shows `calories 0.0`
**And** the screen shows `no allowance`
**And** the screen does not show `protein`,
`carbohydrate`, `fat`, or `remaining`
**When** 100 grams of `Oats` are saved, with calories
`200` per 100 g and the three macro fields left blank
**Then** the screen shows `calories 200.0`
**And** the screen shows `no allowance`
**And** the screen does not show `protein`,
`carbohydrate`, `fat`, or `remaining`
**And** the calorie line is not a negative number

### S2 — Tracking still shows the three totals

**Given** `track_macros` is true
**And** the meal screen for an empty 1990-11-10
**And** no allowance is stored
**Then** the screen shows `calories 0.0`
**And** the screen shows `protein 0.0`,
`carbohydrate 0.0`, and `fat 0.0`
**And** the screen shows `no allowance`
**And** the screen does not show `remaining`
**When** 100 grams of a food are saved with calories
`200`, protein `10`, carbohydrate `20`, and fat `5`
per 100 g
**Then** the screen shows `calories 200.0` and
`protein 10.0`
**And** the screen shows `carbohydrate 20.0` and
`fat 5.0`
**And** the screen shows `no allowance`
**And** the screen does not show `remaining`
**And** the screen does not show an allowance for
protein, carbohydrate, or fat

### S3 — Under the allowance, remaining is positive

**Given** the meal screen shows `calories 200.0` from
100 grams of a food at 200 calories per 100 g
**When** `b` is pressed, ` 500 ` is entered, and
Enter is pressed without Tab
**Then** the idle screen shows `remaining 300.0`
**And** the idle screen shows `calories 200.0`
**And** the idle screen does not show `no allowance`
**When** the allowance is changed to `+500`
**Then** the idle screen still shows `remaining 300.0`

### S4 — Over the allowance, remaining is negative

**Given** the meal screen shows `calories 200.0`
**When** the allowance is `100`
**Then** the screen shows `remaining -100.0`
**And** the screen does not show `no allowance`
**And** the screen does not use a word other than
`remaining` for being over

### S5 — An exact match prints 0.0

**Given** the meal screen shows `calories 200.0`
**When** the allowance is `200`
**Then** the screen shows `remaining 0.0`
**And** the screen does not contain `-0.0`
**And** the screen does not show `no allowance`

### S6 — Remaining uses the calorie total's tenth

**Given** no foods on 1990-11-10
**When** the allowance is `1.25`
**Then** the screen shows `remaining 1.2`
**And** the screen does not show `remaining 1.25`
**And** the screen does not show `remaining 1.3`
**When** `b` is pressed
**Then** the field shows `1.25`, not `1.2`
**When** Enter is pressed without a change
**Then** opening the field again shows `1.25`
**And** the screen shows `remaining 1.2`
**When** the allowance is changed to `0.25`
**Then** the screen shows `remaining 0.2`
**And** the screen does not show `remaining 0.3`
**Given** another meal date with 100 grams of a food
whose calories are `1.25` per 100 g
**And** that date shows `calories 1.2`
**When** that date's allowance is `2.5`
**Then** that date shows `remaining 1.2`
**And** that date does not show `remaining 1.3`
**And** 1990-11-10 still shows `remaining 0.2`

### S7 — A stored zero is set

**Given** the meal screen shows `calories 40.0`
**When** the allowance is `0`
**Then** the screen shows `remaining -40.0`
**And** the screen does not show `no allowance`
**When** the allowance is `0.0`, and again `+0`
**Then** the screen still shows `remaining -40.0`
**When** `b` is pressed
**Then** the field shows `0`
**And** the field is not empty
**When** the food line is removed
**Then** the screen shows `calories 0.0`
**And** the screen shows `remaining 0.0`
**And** the screen does not show `no allowance`
**And** the screen does not contain `-0.0`

### S8 — A negative allowance is refused

**Given** the allowance is `500` and the screen shows
`calories 200.0` and `remaining 300.0`
**When** `b` is pressed, `-1` is entered, and Enter
is pressed
**Then** the editor stays open
**And** the screen shows `error:` and
`allowance: want a number that is not negative`
**When** the next character is typed
**Then** that error line is gone
**And** the editor is still open
**When** Esc is pressed
**Then** the screen shows `remaining 300.0`
**And** the food line is unchanged
**Given** no allowance is stored
**When** `b` is pressed, `-0` is entered, and Enter
is pressed
**Then** the editor stays open with that same
negative error
**When** Esc is pressed
**Then** the screen shows `no allowance`
**And** the screen does not show `remaining`

### S9 — A non-number is refused

**Given** the allowance is `500` and the screen shows
`remaining 300.0`
**When** each of `2000 kcal`, `2000 cal`, `2000 g`,
`2000g`, `1 kg`, `4 oz`, `1 lb`, `1 st`, `1 cup`,
`8 fl oz`, `1 tbsp`, `1 tsp`, `250 ml`, `1/4`,
`1e2`, `inf`, and `NaN` is entered with `b` and
Enter, one at a time
**Then** each Enter leaves the editor open
**And** each shows `error:` and
`allowance: want a number`
**And** none of them shows `not negative`
**When** Esc is pressed after each refusal
**Then** the idle screen still shows
`remaining 300.0`
**And** the stored allowance is still `500`

### S10 — Blank clears back to no allowance

**Given** the allowance is `500` and a food line is
on 1990-11-10
**When** `b` is pressed, the field is cleared, and
Enter is pressed
**Then** the idle screen shows `no allowance`
**And** the screen does not show `remaining`
**And** the screen does not show `want a number`
**And** the food line is still there
**When** `b` is pressed
**Then** the field is empty
**When** the field is only spaces and Enter is
pressed
**Then** the screen shows `no allowance` again

### S11 — A day with no foods can hold an allowance

**Given** 1990-11-10 has no food lines
**And** the screen shows `calories 0.0` and
`no allowance`
**When** the allowance is `1800`
**Then** the screen shows `calories 0.0`
**And** the screen shows `remaining 1800.0`
**And** no food line is shown
**When** the allowance is changed to `.5`
**Then** the screen shows `remaining 0.5`
**When** `b` is pressed
**Then** the field shows `0.5`

### S12 — Another date does not inherit it

**Given** a daily log row exists for 1990-11-01
**And** 1990-11-10 has no foods and its allowance
is `1800`
**And** the screen shows `remaining 1800.0`
**When** `]` is pressed
**Then** the screen shows 1990-11-11
**And** the screen shows `calories 0.0` and
`no allowance`
**And** the screen does not show `1800`
**And** the screen does not show `remaining`
**When** `[` is pressed
**Then** 1990-11-10 shows `remaining 1800.0`
**When** the program quits and the same database is
opened again
**Then** 1990-11-10 still shows `remaining 1800.0`
**And** 1990-11-11 still shows `no allowance`
**And** an existing daily log row is unchanged
**And** the daily log gains no allowance column
**And** the file is still mode `0600`
**And** no second database file is created

### S13 — The chart caption stays

**Given** a daily list with at least two weighed days
**And** the monthly chart caption contains
`Daily deficit:`
**And** the long-term chart caption contains
`Daily deficit:`
**When** those two captions are noted, an allowance
is saved on the meal screen, and each chart is
opened again from the list
**Then** each caption is the same text as before
**And** the meal screen does not contain
`Daily deficit`
**And** the list's weight and trend text for those
days is unchanged
**And** saving the allowance does not insert a daily
log row for a day that had none

### S14 — Pounds do not scale the allowance

**Given** `display_unit` is pounds
**And** the meal screen shows 100 grams of a food
with 100 calories per 100 g
**When** the allowance is `2000`
**Then** the line shows `100.0`
**And** the screen shows `calories 100.0`
**And** the screen shows `remaining 1900.0`
**And** the screen does not show those figures
converted as body weight

### S15 — Foods still save while unset

**Given** no allowance is stored
**When** a food line is saved
**Then** that line is on the idle screen
**And** the calorie total matches the line
**And** the screen shows `no allowance`
**And** the screen does not show `remaining`
**When** `b` is pressed
**Then** the field is empty

### S16 — A weight-log save in flight leaves meals up

**Given** the daily list with 1990-11-01 selected
**And** today is 1990-11-10
**When** `n` opens the day form, Enter starts a save,
Esc returns toward the list, and `f` opens meals
before that save finishes
**And** the save then finishes
**Then** the meal screen is still showing
**And** the heading contains `meals` and `1990-11-10`
**And** Esc shows the daily list with 1990-11-01
still selected
**And** the weight log has reloaded

### S17 — `b` opens the allowance and is not a door

**Given** the idle meal screen with no allowance
**Then** the help line is the idle line in Decision 6
**When** `b` is pressed
**Then** the field labeled `allowance` is focused
**And** the field is empty
**And** the heading still contains the meal date
**And** the help line is the editor line in
Decision 6
**And** the screen does not show `calories`
**And** the screen does not show `no allowance`
**And** the screen does not show `remaining`
**When** Esc is pressed
**Then** the screen shows `no allowance`
**And** nothing was stored
**Given** the allowance is `500`
**When** `b` is pressed
**Then** the field shows `500`, not `500.0`
**And** the screen does not show `calories`
**And** the screen does not show `no allowance`
**And** the screen does not show `remaining`
**When** Enter is pressed without Tab
**Then** the allowance is still `500`
**And** idle `a` still opens the add form on the name
**And** idle Enter still edits grams of the selected
line
**And** idle `f` does not open another screen
**And** the add form's help line still names Tab
**And** `f` on the daily list is still what opens
today's meal screen

### S18 — Letters in the allowance field are text

**Given** the allowance field is open on 1990-11-10
**When** `q`, `a`, `f`, `b`, `[`, and `]` are typed
**Then** those characters are in the field
**And** the date is still 1990-11-10
**And** the process does not quit
**When** Esc is pressed
**Then** nothing was written
**Given** the allowance field is open again
**When** `ctrl+c` is pressed
**Then** the process quits

### S19 — The key reference and the how-to

**Given** the key reference, the meal how-to, and the
config reference
**When** this slice is documented
**Then** the key reference has a meal row for `b`
that edits that day's calorie allowance
**And** that row does not open the meal screen
**And** the `f` row is still the list row that opens
today's meal screen
**And** the list does not gain a `[` or `]` row
**And** the month, monthly chart, and long-term `[`
and `]` rows stay as they are
**And** `docs/how-to/meal-day.md` says `b` edits the
allowance, Enter saves, Esc cancels, and a blank
field clears it
**And** that how-to says the idle screen shows
`no allowance` when none is set, and otherwise
shows remaining calories as the allowance minus
the unrounded calorie total, shown to one decimal
the same way as the calorie line
**And** that how-to says `2.5` against a total of
`1.25` shows `1.2`
**And** that how-to does not tell the reader to
subtract the painted calorie text
**And** that how-to says the allowance is not the
chart's Daily Deficit and not a key in
`config.toml`
**And** that how-to says `display_unit` does not
scale it, and that protein, carbohydrate, and fat
have no allowance
**And** that how-to no longer says that an
allowance, a volume, a recipe, and a starter set
are later slices
**And** volume, recipes, the starter set, and
renaming or deleting a food stay later work
**And** `docs/how-to/README.md` still lists that
how-to
**And** the config reference does not gain an
allowance key

### S20 — A refused allowance survives a reload

**Given** the allowance editor is showing
`allowance: want a number that is not negative`
**When** a weight-log reload finishes
**Then** that error is still shown
**And** the editor is still open
**And** the screen is still the meal screen
**And** Esc afterward shows the daily list and its
normal help
**And** the list does not offer only `q quit`

### S21 — A sub-tenth overage prints 0.0

**Given** 100 grams of a food at `0.04` calories per
100 g, so the unrounded calorie total is `0.04`
**When** the allowance is `0`
**Then** the screen shows `calories 0.0`
**And** the screen shows `remaining 0.0`
**And** the screen does not show `no allowance`
**And** the screen does not contain `-0.0`

## Functional requirements

-   **FR1.** With no allowance stored, the idle screen
    shows the calorie total and the words
    `no allowance`. It does not show `remaining`. It
    does not treat the total as over. It does not
    copy Daily Deficit, another day, or a config
    default. Protein, carbohydrate, and fat day
    totals are shown only when `track_macros` is
    true. Foods still save, and saving one does not
    create an allowance.
-   **FR2.** With an allowance stored, remaining is
    that allowance minus the unrounded calorie total,
    shown with the calorie total's one-decimal
    half-even rule. Under is positive. Over is
    negative, which means the printed tenth has a
    minus. A difference whose tenth rounds to zero
    prints `remaining 0.0`. That covers an exact
    zero, a surplus smaller than half a tenth, and
    an overage smaller than half a tenth. A printed
    `-0.0` becomes `0.0`. `1.25` shows `1.2`.
    `0.25` shows `0.2`. An allowance of `2.5`
    against a total of `1.25` shows `remaining 1.2`,
    not `remaining 1.3`. The editor shows the stored
    number, and Enter on that text unchanged keeps
    it.
-   **FR3.** A stored zero is set, including `0`,
    `0.0`, and `+0`. Remaining is minus the calorie
    total. On a day with no foods, remaining is
    `0.0` and the screen does not show
    `no allowance`. The field shows `0`, not empty.
    A day with no foods can also hold a non-zero
    allowance, and remaining is then that allowance,
    including `.5` shown as `remaining 0.5` and a
    field of `0.5`.
-   **FR4.** After trim, blank clears the allowance
    and returns to `no allowance` without
    `want a number`. A leading `-` is refused,
    including `-0`, and the previous value stays.
    The error is `error:` and
    `allowance: want a number that is not negative`.
    A non-number, including the unit tokens in
    Decision 5, is refused with `error:` and
    `allowance: want a number`, and the previous
    value stays. The next typed character clears
    the error. ` 500 ` and `+500` are the number
    500.
-   **FR5.** `[` and `]` show that date's own
    allowance. Another date does not inherit the
    number. Quit and open of the same file shows
    the same allowances. An existing daily log row
    is unchanged. The daily log gains no allowance
    column. The file stays mode `0600`. Meals do
    not create a second database file.
-   **FR6.** Saving an allowance does not change
    either chart caption or the Daily Deficit
    those captions already show. The list's weight
    and trend text stay. A day with no daily log
    row does not gain one. `display_unit` pounds
    does not scale grams, the calorie total, the
    allowance, or remaining. One hundred grams at
    100 calories per 100 g with an allowance of
    `2000` shows `100.0`, `calories 100.0`, and
    `remaining 1900.0`.
-   **FR7.** A day-form save that finishes while
    the meal screen is up leaves that screen up
    and does not move the list row Esc will
    restore. The save still reloads the log. A
    refused allowance error survives that reload,
    and Esc then shows the ordinary list help.
-   **FR8.** Idle `b` opens the allowance field. The
    editor replaces the idle body. The line list, the
    calorie total, the macro totals, and the
    comparison line are not shown while the field is
    open. The heading keeps the date. It is not a
    second door.
    The idle help line and the editor help line
    are the lines in Decision 6. Unset opens an
    empty field. Enter saves without Tab. Esc
    writes nothing. Tab does not save, does not
    cancel, and does not move to another field.
    Letters, `[`, and `]` are text and do not
    quit or change the date. `ctrl+c` still quits.
    Idle `a`, Enter, `e`, `x`, and `f` stay as
    they are. The food-form help line still names
    Tab.
-   **FR9.** `docs/reference/keys.md` gains one
    meal row for `b` and does not gain a
    second door. `docs/how-to/meal-day.md` gains
    the allowance step in S19 and stops calling
    the allowance a later slice. Volume, recipes,
    the starter set, and renaming or deleting a
    food stay later work.
    `docs/how-to/README.md` still lists that
    how-to. The config reference does not gain an
    allowance key.

## Out of scope

-   Protein, carbohydrate, and fat do not gain
    allowances.
-   The food form is unchanged. This slice does not
    hide those three fields, and it does not change
    when a blank macro is zero.
-   Volume, density, ounces, and pounds as a food
    amount are not this slice.
-   Recipes are not this slice.
-   The starter set is not this slice.
-   Renaming a food is not this slice.
-   Deleting a food is not this slice. Removing a
    meal line does not delete the food.
-   No key other than `b` is added.
    `b` is not a second door onto the meal screen.
-   The allowance is not a `config.toml` key and
    not a column on the daily log.
-   ApplyTrend, Daily Deficit, and both chart
    captions stay as they are.
-   `display_unit` is not read by this screen.
-   A colour role for remaining calories is not
    added. The comparison is the printed sign and
    the words `no allowance`. A tenth that prints
    `0.0` has no minus.
-   Kilojoules are not shown.
-   The alternate screen and the mouse stay off.

## Documentation

Add one meal row to `docs/reference/keys.md`: `b`
edits that day's calorie allowance. Do not add a row
that opens the meal screen. The `f` row stays the
list. Do not add a list row for `[` or `]`. Leave
the month, monthly chart, and long-term bracket rows
as they are.

In `docs/how-to/meal-day.md`, replace the sentence
that an allowance is a later slice. Say that `b` on
the idle meal screen edits the calorie allowance for
the meal date on screen, that Enter saves, that Esc
cancels, and that a blank field clears it. Say that
the screen shows `no allowance` when none is set, and
otherwise shows remaining calories as that allowance
minus the unrounded calorie total, shown to one
decimal the same way as the calorie line. Say that
`2.5` against a total of `1.25` shows `1.2`. Do not
tell the reader to subtract the painted calorie text.
Say that this number is not the chart's Daily Deficit
and not a key in `config.toml`. Say that
`display_unit` does not scale it. Say that protein,
carbohydrate, and fat have no allowance. Keep volume,
recipes, the starter set, and renaming or deleting a
food as later work.

Do not add a how-to page.
`docs/how-to/README.md` already lists the meal page.
Do not add an allowance key to
`docs/reference/config.md`.
