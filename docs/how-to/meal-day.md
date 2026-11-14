# How to log a day's foods

`f` on the daily list opens today's foods. The selected weight-log
row does not choose the date. Today does not need a weight-log row,
and opening the screen does not add one. `[` and `]` move one meal
day. Esc returns to the same list row.

`a` adds a line. A new food asks for calories per 100 g,
then a gram weight. Protein, carbohydrate, and fat are
asked as well when `track_macros` is true, and a blank
one of those is then refused. A name already stored asks
only for grams.
Enter on a line edits grams. `e` edits that food.
Saving changes every meal date that uses the food. The edit
is written as soon as Enter accepts the numbers. `x` removes
the line and keeps the food.

Grams are the food's mass. They are not body weight. `display_unit`
does not scale them.

An allowance, a volume, a recipe, and a starter set are later
slices. Renaming a food and deleting a food are later work.
