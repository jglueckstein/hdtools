# Optional food macros

`track_macros` no longer hides protein, carbohydrate, and
fat on the food form. This revises Decision 24 of
[the meal-day spec](2026-10-30-meal-day.md).

Calories per 100 g stay mandatory. `track_macros` true
makes the three macros mandatory: a blank is refused and
zero is allowed. False or omitted still shows the three
fields. A blank one is stored as zero. A non-number or a
negative is refused either way.

`e` prefills the stored numbers. Leaving them and changing
calories keeps them. Clearing a macro when tracking is off
stores zero.

The idle meal screen shows the three day totals only when
`track_macros` is true. A name already stored still asks
only for grams.
