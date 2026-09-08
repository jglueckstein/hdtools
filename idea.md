# hacker's diet tools

Inspired by [The Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/welcome.html)

## TUI

The project should use a modern tui framework. The original Hacker's
Diet tools were excel spreadsheets or hacker's diet online.

The TUI should look good, not like a dump of aligned columns. It should
use color so the eye can tell structure and meaning at a glance: chrome
(titles, help) vs data, the selected cell or row, and the trend line
versus noisy daily weights. Color should still work on a 16-color
terminal and must not be the only way to read a value (no
color-only encoding).

[charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) is
one way to do that, since the app already uses Bubble Tea. Another
styling library is fine if it fits the same stack.

Color schemes should be user-configurable in the same config file as
display units. The user names roles (at least: title, muted chrome,
header, help, daily weight, trend, selection, error, status), not
raw terminal sequences. A built-in default scheme is the 16-color
palette that ships with the app (blue daily weight, red trend, matching
the book's charts). Omitted or invalid colors fall back to that
default. `NO_COLOR` disables color entirely; structure must remain
readable. Changing scheme must not require a rebuild.

## Database

The data should be stored in a database. The database may be local
(default), or if the user has access to a server, the TUI should be able
to connect to a remote database.

## Weight Monitoring

The tool should implement the weight monitoring system described
[in the Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/weightmonitor.html).

### Daily Log

By default, the daily log should track

- weight
- sleep (hours)
- steps
- whether a workout happened or not

A trend number should be calculated from the daily weight.

Tracking calories is a separate tool, see below.

It would be good to add user configuration to the daily logging. For
example, individuals may want to track a sleep quality number, water
intake, mood, stress, resting HR, HRV.

The TUI should make it easy to create and edit daily logs.

### Monthly Log

We should support the printing of blank sheets for recording daily
numbers. The TUI should make it easy to enter a whole month's data in
one sitting. The TUI should make a nice display of an entire month's
data sheet.

The tool should be able to generate pdf of the monthly log sheets
(filled in, not the blank sheets mentioned above), and pdfs of the
monthly charts.

The TUI should be able to display on screen the monthly charts and the
long term charts.

The style of the charts is described more fully [in this
section](https://www.fourmilab.ch/hackdiet/e4/signalnoise.html)

## Meal Planning

The tool should implement the meal planning system described
[in the Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/planningmeals.html#Fa147).

We should have an initial database of foods. We should have protein and
carbs in addition to calories.

We should prefer weight over volume measurements. Internally, we should
try to work consistently with grams. We will support U.S. units.

Besides the ingredients, we should save groups of ingredients as recipes
or meals. Then the user can enter 1/4 of this recipe or 245 g of that recipe.
