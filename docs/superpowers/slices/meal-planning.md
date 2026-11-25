---
task: >
  Meal planning from the Hacker's Diet chapter cited in idea.md.
  One list is both the plan and the log; the original plan is not
  kept. Each food has calories, protein, carbohydrate, and fat.
  The day has a calorie allowance the person sets, compared with
  the list's calorie total. That allowance is not the chart's
  Daily Deficit, and there is no macro allowance. A line is a
  weighed food, a fraction of a saved recipe, or a gram weight of
  a recipe. A recipe is ingredients, which may be foods or other
  recipes, and it cannot contain itself. Amounts are stored in
  grams. Typed weight is grams, ounces, or pounds. Typed volume
  is cups, fluid ounces, tablespoons, teaspoons, or milliliters,
  converted through one grams-per-milliliter density when the
  food has one. A food with no density is weight only. The meal
  date is not the weight-log cursor. Stone is not a food unit.
  The starter set, the unset-allowance display, and the TUI door
  are not settled. Nutrition figures and the cup-to-milliliter
  factor are not settled either.
task_slug: meal-planning
date: 2026-10-30
carpaccio_model: grok-4.7
inseparable: false
progressed_slice: S3
slices:
  - id: S1
    title: Open a day's weighed foods from the list with f
    scope: >
      From the daily list, f opens a meal screen for one date.
      It is not a column on the month sheet and not a field on
      the day form. Those columns stay weight, sleep, steps,
      workout, and note. f on the month sheet still starts an
      edit. f on either chart does nothing. n, m, c, l, p, P, q,
      and t keep the actions in docs/reference/keys.md. Esc
      returns to the list and leaves the selected log row where
      it was. The list help names f. The key reference gains
      that row. A how-to at docs/how-to/meal-day.md starts with
      this screen, and docs/how-to/README.md lists it.
      The meal date is the screen's own date. f opens today,
      even when the selected log row is another day, and even
      when today has no daily_log row. On this screen, [ and ]
      move that meal date by one day. They do not move the list
      selection and they do not change the month sheet. Opening
      or editing a meal date does not insert a daily_log row
      and does not change weight, sleep, steps, workout, note,
      trend, or either chart.
      There is one list for that date. Nothing stores a plan
      beside it. Adding, changing, or removing a line is the
      log. A line in this slice is one food and a weight in
      grams. Ounces, pounds, volume, and stone are refused
      here. Zero and negative grams are refused. The four
      numbers on a food are calories, protein, carbohydrate,
      and fat, and a food missing any of them is refused. The
      proposal, not a settled rule, is that those numbers are
      per 100 g and that a line contributes each number times
      grams divided by 100. Calories are kilocalories. The line
      stores the grams and which food. It does not store a copy
      of the four numbers and it does not store a calorie
      column. Totals are computed. The proposal is that changing
      a food's numbers changes the totals of every meal date
      that uses it. The screen shows the lines and the four
      totals to one decimal. An empty date shows 0.0 for each
      total. display_unit does not scale these grams. Food mass
      is not body-weight kilograms.
      The person can add a food that is not stored yet. This
      slice ships no starter catalog. A name is unique after
      surrounding space is trimmed, and the match is
      case-sensitive. A blank name is refused. Inner keys are a
      proposal: a adds a line, Enter edits the selected line's
      grams, e edits that food's four numbers, and x removes
      the line. There is no allowance on this screen yet.
      Meals live in the same SQLite file as daily_log, in new
      tables, not in new columns on daily_log. The file stays
      mode 0600. Application code stays under internal/.
      cmd/hdtools stays wiring. No AltScreen and no mouse.
    decision_focus: >
      Does the meal list open as its own screen from the daily
      list with f, on a date that is not the weight-log cursor,
      instead of a log column, a form field, or another key?
    lens_used: decision-boundary
    sequencing_note: >
      First. S2, S3, and S4 need this screen and a gram line.
      Nothing in the repo has to land before it.
    disposition: accepted
    disposition_rationale: >
      Accepted. The meal list is its own screen. f on the daily
      list opens today, and [ and ] move that meal date. The
      weight log keeps its columns and its cursor. Nutrition is
      per 100 g. A line stores the food, so editing that food
      updates every day that uses it. On that screen, a adds a
      line, Enter edits grams, e edits the food, and x removes
      the line. This slice is grams only. S1 has landed.
    file_as_issue: false
    issue_url: https://github.com/jglueckstein/hdtools/issues/123
    merged_into: null
  - id: S2
    title: Compare that day's calories with an allowance
    scope: >
      The meal screen from S1 gains a calorie allowance for
      that meal date only. The person types it. It is stored
      with the meal date in the same SQLite file. It is not a
      key in config.toml, not a column on daily_log, and not
      the chart's Daily Deficit. Saving it does not call
      ApplyTrend and does not change a chart caption. Protein,
      carbohydrate, and fat do not gain allowances. Their
      totals stay the S1 totals.
      When the allowance is set, the screen shows remaining
      calories as the allowance minus the day's calorie total,
      to one decimal. Under the allowance, remaining is
      positive. Over it, remaining is negative. Zero is a set
      allowance: remaining is minus the calorie total, not the
      unset state. A negative allowance is refused and the
      stored value stays as it was. A non-number is refused the
      same way. An allowance may be set when the day has no
      foods; remaining is then the allowance.
      When no allowance is set, the screen still shows the
      calorie total and the three macro totals, and the
      comparison is the words "no allowance". It does not show
      remaining 0.0, does not treat the total as over, and does
      not copy Daily Deficit, another day's allowance, or a
      config default. Clearing the field returns to that unset
      state. Foods still save while it is unset. Another meal
      date does not inherit the number. The meal how-to gains
      this step. The key table does not gain a second door.
    decision_focus: >
      When the meal date has no calorie allowance, does the day
      show the calorie total and the macro totals with the
      words "no allowance", rather than zero, a carried-forward
      number, or the chart's Daily Deficit?
    lens_used: decision-boundary
    sequencing_note: >
      After S1. It does not need S3, S4, or S5. Dropping it
      leaves the day list and the four totals.
    disposition: accepted
    disposition_rationale: >
      Accepted. An empty allowance shows the four totals and the
      words "no allowance". A stored zero is an allowance of
      zero. When an allowance is set, remaining is that allowance
      minus the day's calorie total, positive when the day is
      under it. The chart's Daily Deficit stays the chart's
      number. This slice waits until S1 has landed.
    file_as_issue: true
    issue_url: https://github.com/jglueckstein/hdtools/issues/140
    merged_into: null
  - id: S3
    title: Turn typed US amounts into stored grams
    scope: >
      A food line's typed amount is converted to grams before
      it is stored. Typed weight is grams, ounces, or pounds.
      Typed volume is cups, fluid ounces, tablespoons,
      teaspoons, or milliliters. Stone is rejected. A kilogram
      token is rejected: it is not one of those weights, and
      body weight already uses kilograms for a different
      quantity. The proposed tokens are g, oz, lb, cup, fl oz,
      tbsp, tsp, and ml. oz is weight. fl oz is volume. A bare
      c is not a cup. Tokens are case-insensitive. A decimal or
      a single slash fraction such as 1/4 or 3/4 is allowed.
      A mixed number such as 1 1/2 is refused.
      The food may store one density, in grams per milliliter.
      If it has one, any volume above becomes milliliters, then
      grams, through that density. If it has none, a volume is
      refused and no line is written; a weight still saves.
      There is no stored gram weight for a cup, a tablespoon,
      or any other measure. Editing the density changes later
      conversions. It does not rewrite grams already stored on
      a line.
      Every factor below is a proposal. Weight, as one group:
      one ounce is 28.349523125 g, and one pound is 16 of those
      ounces. Volume, as one group, the nutrition-label set,
      which is the US legal cup: one milliliter is 1 ml, one
      cup is 240 ml, one fluid ounce is 30 ml, one tablespoon
      is 15 ml, and one teaspoon is 5 ml. Do not mix that cup
      with the customary-cup group. The stored gram value is
      the full product, not a kitchen rounding. The saved line
      shows grams to one decimal. One cup and 240 ml of the
      same food store the same grams only if this cup proposal
      is the one that lands. One cup of a food with density
      1.03 then stores 247.2 g. One ounce stores 28.349523125 g,
      which one decimal on the line does not distinguish from
      28.35; the stored value is the check. The meal how-to
      names the tokens and these factors as proposals until
      this slice is accepted.
    decision_focus: >
      Do typed ounces, pounds, and volumes become stored grams
      through one grams-per-milliliter density and the proposed
      factor group, with volume refused when the food has no
      density?
    lens_used: decision-boundary
    sequencing_note: >
      After S1. It does not need the allowance. S4 can type a
      volume ingredient after this. S5's liquids need this
      density field.
    disposition: accepted
    disposition_rationale: >
      Accepted. Use the nutrition-label group. One cup is 240 ml,
      one fluid ounce is 30 ml, one tablespoon is 15 ml, and one
      teaspoon is 5 ml. One ounce is 28.349523125 g, and one
      pound is 16 of those ounces. Tokens are g, oz, lb, cup,
      fl oz, tbsp, tsp, and ml. The stored grams are the full
      product. A food with no density still accepts weight and
      still refuses volume. Changing the density does not rewrite
      grams already stored on a line. This slice waits until S1
      has landed.
    file_as_issue: true
    issue_url: null
    merged_into: null
  - id: S4
    title: Log a fraction or a gram weight of a recipe
    scope: >
      The person can save a recipe that is not in any starter
      set. A recipe has a name and ingredients. An ingredient
      is a food and a gram amount, or another saved recipe and
      either a fraction or a gram weight. Gram entry works
      even if S3 was dropped. If S3 has landed, an ingredient
      uses that same typed-amount parser, and the grams are
      what is stored. The name rule is the S1 rule. This slice
      does not ship the starter recipes.
      A recipe cannot contain itself, directly or through
      another recipe. That save is refused and no recipe row is
      written. The proposal for deletion: a food or recipe that
      is still an ingredient of another recipe cannot be
      deleted until it is removed there. No cascade.
      The day's one list gains a line that is a fraction of a
      recipe or a gram weight of a recipe, beside weighed
      foods. The line stays one row. It does not expand into
      ingredient rows. Totals add the scaled calories, protein,
      carbohydrate, and fat. Logging the recipe does not keep a
      second plan. Changing a food's four numbers changes the
      contribution, the same live reference as S1. Changing the
      recipe's ingredients changes days that log it.
      The yield proposal: yield is the sum of ingredient grams
      after nested recipes are expanded to foods. There is no
      servings field. One quarter scales every expanded food by
      0.25. A gram weight scales by that weight divided by the
      yield. Yield zero cannot be logged. A fraction must be
      greater than zero, and a fraction above one is allowed,
      so twice the recipe is 2. The outer scale multiplies the
      inner scale once.
      What you can check: a recipe of 40 g of one food and
      200 g of another has yield 240 g. Logging 1/4 adds a
      quarter of those two foods to the day. Logging 245 g
      adds 245/240 of them. A recipe that lists itself does
      not save. The meal how-to gains this step.
    decision_focus: >
      Is a recipe's yield the sum of its ingredient grams, so
      that a fraction and a gram weight are two scales of that
      mass, with no servings field?
    lens_used: decision-boundary
    sequencing_note: >
      After S1. After S3 if an ingredient should be typed as a
      volume. Gram recipes do not need S2 or S3. S5's two
      recipes need this yield.
    disposition: accepted
    disposition_rationale: >
      Accepted. Yield is the sum of ingredient grams after nested
      recipes expand to foods. A fraction and a gram weight are
      two scales of that mass. There is no servings field. A
      fraction above one is allowed. A food or recipe that is
      still an ingredient cannot be deleted until it is removed
      there. This slice waits until S1 has landed.
    file_as_issue: true
    issue_url: null
    merged_into: null
  - id: S5
    title: Ship a short starter set on top of your foods
    scope: >
      A database gains the proposed foods and two recipes below.
      A name that is already stored is skipped. Its numbers,
      its density, and its ingredients are not replaced. A name
      the person added that is not in the set is left in place.
      Opening the app again does not insert a second copy. The
      person can still add a food or a recipe that is not in
      the set. The figures assume S1's per-100 g proposal and
      S4's mass yield. Densities are grams per milliliter and
      do not bake in the cup factor. If the per-100 g basis is
      revised, this list is void and has to be rewritten.
      Every figure is a proposal. They are rounded so the list
      can be reviewed in one sitting. They are not a USDA
      import, and they are not accepted by being written here.
      The four numbers are not forced to match 4, 4, and 9
      kilocalories per gram. Correct any figure before
      accepting the slice.
      Foods, per 100 g, as calories, protein, carbohydrate,
      fat, and density if any. Chicken breast, cooked: 165,
      31.0, 0.0, 3.6, no density. Egg, whole: 143, 12.6, 0.7,
      9.5, no density. Oats, dry: 389, 16.9, 66.3, 6.9, no
      density. Milk, whole: 61, 3.2, 4.8, 3.3, density 1.03.
      Olive oil: 884, 0.0, 0.0, 100.0, density 0.91. Banana:
      89, 1.1, 22.8, 0.3, no density. Rice, white, cooked:
      130, 2.7, 28.2, 0.3, no density. Broccoli, raw: 34, 2.8,
      6.6, 0.4, no density. Bread, whole wheat: 247, 13.0,
      41.0, 4.2, no density. Peanut butter: 588, 25.0, 20.0,
      50.0, no density.
      Recipes, flat, ingredients by grams. Oats with milk:
      40 g Oats, dry and 240 g Milk, whole. Chicken and rice:
      150 g Chicken breast, cooked and 200 g Rice, white,
      cooked. Neither recipe contains another recipe.
      What you can check on a new database: f shows those
      names. 100 g of Oats, dry shows that row's four numbers
      as the day's totals. One quarter of Oats with milk shows
      a quarter of 40 g of those oats plus 240 g of that milk.
      Add a food named my stew and it remains. A later open
      does not reset a starter name whose numbers you edited,
      and it does not overwrite my stew. The meal how-to names
      the set as the proposal that shipped.
    decision_focus: >
      Does the database gain this proposed set of foods and two
      recipes, inserting a name only when it is absent, so a
      food you already added is left alone?
    lens_used: decision-boundary
    sequencing_note: >
      Last. It needs the food line from S1, density from S3,
      and recipe yield from S4. It does not need the allowance.
      Dropping it leaves foods and recipes you add yourself.
    disposition: accepted
    disposition_rationale: >
      Accepted. Ship the ten foods and two recipes written in
      this slice: cooked chicken breast, whole egg, dry oats,
      whole milk, olive oil, banana, cooked white rice, raw
      broccoli, whole-wheat bread, and peanut butter. Oats with
      milk is 40 g oats and 240 g milk. Chicken and rice is 150 g
      chicken and 200 g rice. Insert a name only when it is
      absent. A later open does not reset a starter you edited
      and does not overwrite a food you added. This slice waits
      until S1, S3, and S4 have landed.
    file_as_issue: true
    issue_url: null
    merged_into: null
---

This record slices meal planning. idea.md points at the
[Hacker's Diet meal chapter](https://www.fourmilab.ch/hackdiet/e4/planningmeals.html#Fa147)
and asks for a food database, protein and carbohydrate as well as
calories, grams rather than volume, and recipes you can log by
fraction or by gram weight. A later handoff is the authority. It
settled one editable list, a calorie allowance that is not the
chart's Daily Deficit, fat as a fourth total with no macro targets,
nested recipes that cannot loop, and a meal date that is not the
weight-log cursor. It left three choices open: which foods ship,
what an unset allowance shows, and where the list sits. The
cup-to-milliliter factor was not accepted either, and a recipe still
needs a yield before "1/4" and "245 g" are a number. Those five are
the slices. Settled behaviour sits in the slice that first has to
show it. Every nutrition figure and every conversion factor below is
a proposal.

## S1 — Open a day's weighed foods from the list with f — decision-boundary

### Context

The daily list, the day form, and the month sheet are the weight
log. The month sheet's columns are fixed: weight, sleep, steps,
workout, and note. A letter that is not a command on that sheet
starts an edit. Letter f is not in the key table. n, m, c, l, p, P,
q, and t already do something on the screens that table names.

There is no foods table. The schema today is `daily_log` in
`internal/dailylog/store.go`: one row per day, weight in kilograms,
no trend column. ApplyTrend and the chart's Daily Deficit belong to
that series. They are not a meal. `display_unit` changes how body
weight is painted. It must not scale food.

What you can see when it lands: from the list, f opens today, you
add a food, you enter grams, and the screen shows calorie, protein,
carbohydrate, and fat totals. Esc returns to the same list row.
Selecting another log row and pressing f does not change the meal
date. The charts do not move.

### Decision content

The engagement is the door. The proposal is an own screen, opened
with f from the daily list only, on a date that is not the selected
log row. On that screen, [ and ] move the meal date and leave the
list selection where it is.

The alternatives are a column or a note on the day form, a new
month-sheet column, f from every screen, a command flag, or opening
whichever day the weight log has selected. A form field cannot hold
a list of lines. A new log column breaks the fixed daily log. A flag
is a second product; `cmd/hdtools` is wiring. Binding the date to
the cursor makes a meal for a day you cannot open unless that day
already has a weight row.

Two proposals are bundled so the first line means something. They
are not settled, and they are not separate slices. Revise this slice
if you want different ones. The four numbers are per 100 g, and the
line scales by grams divided by 100. The line stores grams and which
food, not a snapshot of the four numbers, so editing the food
changes past totals. The alternative to per 100 g is a declared
serving. The alternative to a live food is freezing the four numbers
onto the line. Inner keys are also a proposal: a adds, Enter edits
grams, e edits the food, x removes the line.

### Dependencies

Nothing in the repo has to land first. The list, the key table, and
the single SQLite file already exist. S2 adds the allowance on this
screen. S3 adds units on this line. S4 adds a recipe line beside the
food line. S5 fills names into the store this slice leaves empty.

### Rationale

Without a door, a food row in SQLite is not something you can log.
The open choice is where that door is. The allowance wording, the
cup factor, and the starter list are different acceptances. Putting
them here would make one yes cover four decisions.

Gram-only is enough to see a total. Per 100 g rides along because a
line has to scale, and a basis with no line on screen ships nothing.
The live food reference rides along for the same reason: the first
saved line has to say what a later edit means. Daily Deficit is not
on this screen. The weight log's columns and ApplyTrend stay as they
are.

## S2 — Compare that day's calories with an allowance — decision-boundary

### Context

After S1, a day has lines and four totals, and no target. The
handoff says the person sets a calorie allowance for that day, and
the list's calorie total is compared with it. The chart's Daily
Deficit is a different sum: a weight trend, first-minus-last or a
fit, times 3500. It is not an allowance, and this screen must not
read it.

What you can see when it lands: a day with food and an empty
allowance shows the totals and the words "no allowance". You set
2000, and remaining is the allowance minus the calorie total. You
clear the field, and the words come back. Another date still has no
allowance. The monthly chart's Daily Deficit line is unchanged.

### Decision content

The engagement is the unset day. The proposal is: foods still save,
the four totals still show, and the comparison reads "no allowance".
Zero is not that state. A stored zero is a set allowance, so
remaining is minus the calorie total.

The alternatives are to refuse foods until an allowance is set, to
treat unset as zero, to carry yesterday's allowance forward, to hide
the calorie total until it is set, or to drop the chart's Daily
Deficit into the gap. Carry-forward invents a target you did not
type. The chart figure moves when the weight trend moves, which is
the wrong input for a meal.

When an allowance is set, the comparison proposal is remaining
equals allowance minus calorie total, to one decimal, positive when
you are under. Revise this slice if the set day should show overage
instead. That sign is part of making the comparison visible. It is
not a second door.

### Dependencies

S1 must land first. The field sits on that screen and uses that meal
date. S3, S4, and S5 do not have to land first. If this slice is
dropped, the day list and the four totals from S1 still work.

### Rationale

The allowance is settled. What the day shows when you have not set
one is not. That display is easy to rubber-stamp if it shares a
review with the key, because a wrong word sits next to a new screen
that otherwise looks right.

The slice is done when an empty allowance and a stored zero look
different, when another date does not inherit the number, and when
the chart caption does not change. Macro totals stay totals. They do
not grow targets in this slice.

## S3 — Turn typed US amounts into stored grams — decision-boundary

### Context

S1 stores a gram weight. The handoff also allows ounces and pounds,
and volume when the food has a density. Density is one number, grams
per milliliter. A food with no density is weight only. A separate
gram weight per cup or spoon is already ruled out. The factor that
turns a cup into milliliters was not accepted. Stone is not a food
unit. Body weight stays kilograms in `daily_log` and does not become
the food unit.

What you can see when it lands: you put a density of 1.03 on a food
and enter a cup. If the proposal stands, that cup stores the same
grams as 240 ml, which is 247.2 g. You enter one ounce and the
stored grams are 28.349523125. You enter a cup of a food with no
density and nothing is written. One hundred grams of that food still
saves.

### Decision content

The engagement is the factor group. The model is settled: one
density, every allowed volume uses it, missing density refuses
volume. The numbers are not settled. Every figure here is a
proposal.

Weight, taken together: one ounce is 28.349523125 g, and one pound
is 16 of those ounces. The alternative is a rounded ounce of 28.35 g,
still with a pound of 16 ounces. Volume, taken together, the
nutrition-label set, which is the US legal cup: one cup is 240 ml,
one fluid ounce is 30 ml, one tablespoon is 15 ml, and one teaspoon
is 5 ml. One milliliter is 1 ml. The alternative is the US customary
group, also taken together: one cup is 236.588 ml, and the spoon and
fluid-ounce figures are the ones that match that cup. A mix of a
240 ml cup and a 29.5735 ml fluid ounce is not a proposal.

Revise the group in this slice if you want the customary cup. Do not
treat 240 as already agreed because it appears here. Token spelling
is a proposal attached to that group: g, oz, lb, cup, fl oz, tbsp,
tsp, and ml, with oz as weight and fl oz as volume.

### Dependencies

S1 must land first. The converter needs a food line to land on. The
allowance does not. S4 should follow if a recipe ingredient is typed
as a cup; a gram ingredient does not need this slice. S5's two
liquids store a density that this slice is what makes a cup mean.

### Rationale

The cup factor changes stored grams for every volume line after it.
Folding it into the key, or into the starter list, hides a bad
constant inside a screen or a food name you otherwise want.

The slice is done when one cup and 240 ml store the same grams under
this proposal, when one ounce stores the full ounce factor, and when
a food with no density still accepts grams and still refuses a cup.
A table of gram weights, one per spoon, is not an alternative inside
this slice. That shape was ruled out.

## S4 — Log a fraction or a gram weight of a recipe — decision-boundary

### Context

idea.md says you can enter 1/4 of a recipe or 245 g of that recipe.
The handoff says a recipe is a group of ingredients, an ingredient
may be a food or another recipe, and a recipe cannot contain itself,
including through a loop. The day still has one list. A recipe line
is a line, not a second plan and not an exploded pile of foods.

What "1/4" multiplies is not stated. A quarter of a declared serving
and a quarter of the batch are different lunches.

What you can see when it lands: you save a recipe of 40 g of one
food and 200 g of another. You log 1/4 on a date, and the totals are
a quarter of those two foods. You log 245 g on another date, and the
totals are 245/240 of them. The line stays one row. A recipe that
lists itself does not save.

### Decision content

The engagement is the yield. The proposal is that yield is the sum
of ingredient grams after nested recipes expand to foods, and that
there is no servings field. A fraction scales that mass. A gram
weight scales by grams divided by that yield. Yield zero cannot be
logged.

The alternative is a servings count. Then "1/4" might mean a quarter
of one serving, or you would log a count of servings instead of a
fraction of the recipe. That needs a field the handoff does not
have, and it makes 245 g a second definition of the same batch.
idea.md pairs the fraction and the gram weight as two ways into the
same recipe. The mass yield is the reading that makes those two
phrases one scale. Revise this slice if you want servings instead.
Do not add servings beside the mass yield. Two yields would make
1/4 ambiguous again.

Deletion is a proposal attached here, not a second decision: you
cannot delete a food or a recipe that is still an ingredient until
you remove it. The alternative, cascade, drops lines you did not
open. A fraction above one is allowed, so 2 means twice the batch.

### Dependencies

S1 must land first. The recipe line is a line on that list, and the
ingredients are foods this screen can store. S3 is not required for
gram ingredients. If S3 has landed, ingredient amounts use that
parser. S2 is not required. S5's two recipes use this yield, so they
wait until it is accepted.

### Rationale

Recipes are settled as a shape and unsettled as a scale. You cannot
check 1/4 against 245 g until the yield is chosen. Reviewing that
choice inside the starter list would tie a definition to a set of
names. You should be able to reject the starter foods and still save
a recipe of your own.

The slice is done when the quarter and the 245 g line match the same
expanded foods, when the row does not split into ingredients, and
when a cycle is refused and the day is unchanged. The plan is still
not a second list.

## S5 — Ship a short starter set on top of your foods — decision-boundary

### Context

The handoff says a starter set ships, and that you can add a food or
a recipe that is not in it. Which names, and which numbers, are the
open choice. Draft nutrition numbers are not accepted by being
drafted. S1's per-100 g rule and S4's mass yield are what make a
shipped recipe mean a total you can read. S3's density is what makes
a shipped liquid usable by the cup. The densities themselves do not
include the cup factor.

What you can see when it lands: on a new database, f lists these
foods and the two recipes. One hundred grams of a starter food shows
that row's four numbers. One quarter of Oats with milk matches a
quarter of its two ingredients. A food you added is still there on
the next open, and a starter name you already changed is not reset.

### Decision content

The engagement is the list. The proposal is the ten foods and two
recipes in the scope. Every calorie, gram, and density there is a
proposal. They are rounded so you can read them in one sitting.
Correct a figure rather than accepting it as a measurement. They are
not checked against 4, 4, and 9 kilocalories per gram. Olive oil's
884 kcal beside 100 g of fat is that kind of proposal, not a claim
that the arithmetic closes.

A name is inserted only when it is absent. That is part of this
decision. The alternative is to overwrite on every open, which would
destroy a correction, or to seed only a completely empty database
and skip the set forever once you have added one food.

The other alternatives are an empty ship, so you type every food, or
a full USDA import. An empty ship drops the settled "a starter set
ships" without putting a reviewable list in front of you. An import
is a different product. Revise the names and the figures in this
slice if the short list is wrong. Do not treat a number as agreed
because the slice contains it.

### Dependencies

S1, S3, and S4 must land first. The set needs a food line, a density
field, and a recipe yield. S2 is not required to seed or to log a
starter recipe. If S1's per-100 g proposal is revised, every figure
here has to be rewritten before this slice can land. Dropping this
slice leaves the foods and recipes you add yourself.

### Rationale

The starter list is the remaining open choice from the handoff, and
it is the easiest to accept unread. A short list with round figures
is small enough to correct. A USDA dump would make acceptance cheaper
than reading.

It is last so a recipe in the set can be logged and a liquid can be
entered by the cup. Seeding earlier would ship names you cannot yet
weigh as a recipe, or densities nothing reads. The skip-if-present
rule is in this slice because the earlier slices already let you add
foods. A seed that overwrites those foods would punish that.

## Sequencing recommendation

Land S1, then S2, then S3, then S4, then S5.

S1 is the door and the first gram line. You can see a total without
an allowance, a cup, or a recipe. The weight log is still easy to
check, because that slice does not write `daily_log` and does not
call ApplyTrend.

S2 is only the allowance on that screen. It should not wait for
units or recipes, and it should not be merged into S1. The key and
the words "no allowance" are separate yeses. Dropping S2 leaves the
day list.

S3 is the factor table. It can land any time after S1. It is placed
after S2 so the allowance review is not mixed with the cup factor.
Dropping S3 leaves gram entry. Do not merge it into S1. A wrong cup
would then ride along with a key you meant to accept.

S4 needs foods from S1. It does not need the allowance. It is after
S3 so an ingredient can be typed in the same units as a food line.
Gram recipes still work if S3 is dropped. Do not merge S4 into S5.
You should be able to save your own recipe if you reject the starter
list.

S5 is last. The figures assume per 100 g, one density, and a mass
yield. If any of those was revised, rewrite the list before seeding
it. Dropping S5 leaves user-added foods and recipes in place.

S2 and S3 may be swapped after S1. The order above is the one that
adds a single decision to a screen you have already seen.

## Explicitly not slicing on

- Files or layers. `cmd/hdtools`, a new package under `internal/`,
  and `internal/dailylog` are not decision boundaries. Each slice is
  a behaviour on the meal screen or in the stored meal. The schema
  rides inside the slice that first stores that behaviour.
  `daily_log` does not gain meal columns. A foods table with no line
  on screen is not a slice.
- Keeping a plan beside what was eaten. One list is settled. A
  second planned copy, and a "commit the plan" step, are not slices.
- Using Daily Deficit as the allowance. The chart figure stays the
  weight-trend rate times 3500. It is not a meal target and not a
  candidate for S2. ApplyTrend is not opened.
- Macro allowances, macro percentages of calories, and a
  kilojoule display. Protein, carbohydrate, and fat are totals in
  grams. Targets for them, and a second energy unit, are another
  sitting.
- A gram column per volume measure. One density in grams per
  milliliter is settled. Cup, tablespoon, teaspoon, and fluid ounce
  do not get their own stored gram weights.
- Stone as a food unit, and kilograms as a food unit. Stone stays a
  body-weight display unit. Typed food weight is grams, ounces, or
  pounds. Body-weight kilograms stay in `daily_log`.
- Binding the meal date to the weight-log cursor, and creating a
  `daily_log` row because a meal date was opened. The meal date is
  its own date. That coupling is not a slice.
- A week planner, a seven-day grid, or dragging a meal onto another
  date. The settled sitting is one day and one list.
- A full USDA import, a barcode lookup, or a network food search.
  S5 is a short list you can read. An import is a different task.
- A meal PDF, plus the blank log sheets and long-term chart PDFs
  still in idea.md. Those are other tasks. This record does not
  slice them.
- Splitting the per-100 g basis, the unset-allowance wording, or the
  letter f into slices that ship nothing on their own. Per 100 g is
  a proposal inside S1, because a line has to scale. The unset
  wording is S2's decision, and S2 also ships the allowance. The key
  is S1's decision, and S1 also ships the day list. The inner keys
  for add, edit, and remove ride with S1 for the same reason.
- A separate slice for snapshot versus live food numbers, and a
  separate slice for cascade versus refuse-when-in-use. The first
  saved line and the first saved recipe have to mean something, so
  those are proposals in S1 and S4. Revise those slices to change
  them.
- A second SQLite file, a `config.toml` default allowance, an
  Atwater check that rewrites a food's calories, fibre, sodium, or
  other micronutrients. Food mass stays grams in the existing
  database. The four numbers are the ones the handoff names.
- Colour, `NO_COLOR`, and a `[colors]` role for macros or for
  remaining calories. Sign and the words "no allowance" are how the
  comparison is read.
- An AltScreen or a mouse binding. The TUI stays
  `tea.NewProgram(model)` as it is.
- Writing the dated spec, editing idea.md, or filing issues in this
  record. The how-to and the key row travel with the slice that
  makes them true. The spec comes after a slice is accepted.
