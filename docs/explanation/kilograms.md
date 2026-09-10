# Why weight is stored in kilograms

Display units (`kg`, `lb`, `st`) are a preference. The number in SQLite
is always kilograms.

If the file stored whatever the TUI last showed, changing
`display_unit` would rewrite history, or the series would mix scales.
Kilograms are SI, match the gram-based meal planning the product
intends, and give one canonical series for the trend.

`display_unit` is therefore fail-closed: an unknown value must not
start the TUI, because the alternative is treating `"lbs"` as kilograms
on save. Colour values are the opposite — they cannot corrupt the
series — which is why a bad `[colors]` entry falls back instead of
refusing to open the log.

Config lives under XDG, not in the database, so a log file can move
machines without dragging display choices with it.
