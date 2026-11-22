# How to log a day's foods

`f` on the daily list opens today's foods. The selected weight-log
row does not choose the date. Today does not need a weight-log row,
and opening the screen does not add one. `[` and `]` move one meal
day. Esc returns to the same list row.

`a` adds a line. A new food asks for calories per 100 g,
then protein, carbohydrate, and fat per 100 g, then a
gram weight. When `track_macros` is true, a blank macro
is refused. When it is false or omitted, a blank macro
is stored as zero. A name already stored asks only for
grams.
Enter on a line edits grams. `e` edits that food.
Saving changes every meal date that uses the food. The edit
is written as soon as Enter accepts the numbers. `x` removes
the line and keeps the food.

Grams are the food's mass. They are not body weight. `display_unit`
does not scale them.

`b` edits that day's calorie allowance. Enter saves it. Esc
cancels and writes nothing. A blank field clears it. The idle
screen shows `no allowance` when none is set. Otherwise it shows
remaining calories: the allowance minus the unrounded calorie
total, shown to one decimal the same way as the calorie line.
An allowance of 2.5 against a total of 1.25 shows 1.2. That
allowance is not the chart's Daily Deficit and not a key in
`config.toml`. `display_unit` does not scale it. Protein,
carbohydrate, and fat have no allowance.

A volume, a recipe, and a starter set are later work. Renaming
a food and deleting a food are later work.
