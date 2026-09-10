# Why colour is a scheme of roles

The TUI uses colour so the eye can separate chrome from data, the
selected row from the rest, and the trend from noisy daily weights.
Those are roles, not raw terminal sequences. A `[colors]` table names
the roles so you can copy hex out of an existing theme without
rebuilding.

Sixteen named colours are the floor: the default palette, and the
downshift target when a hex value is shown on a 16-color terminal. They
are not a ceiling; `#rrggbb` is a valid value.

Colour is never the only cue. Selection keeps a `>` mark and reverse
video. Columns have headers. Trend is bold. `NO_COLOR` (environment
only, not a config key) strips chromatic foreground and background and
leaves that structure.

A bad colour must not lock you out of the log, so omitted and invalid
values fall back silently. That is the inverse of `display_unit`, which
can corrupt stored kilograms if it is wrong.

Sleep, steps, workout, and note have no colour keys. The scheme spends
its budget on the book's signal/noise split, not on a fully themed
table. There is no per-role "off" token: the only way to drop chroma is
global `NO_COLOR`.
