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
display units (`[colors]` table). The user names roles (`title`,
`muted`, `header`, `help`, `weight`, `trend`, `selection`, `error`,
`status`), not raw terminal sequences. A role value is a 16-color name
or a hex color (`#rrggbb` / `#rgb`) so the file can copy definitions
from existing terminal and editor themes. A built-in default scheme is
the 16-color palette that ships with the app (blue daily weight, red
trend, matching the book's charts). Omitted or invalid colors fall back
to that default. Hex still works on a 16-color terminal by mapping down
to the nearest of those sixteen. `NO_COLOR` disables chromatic color;
structure (headers, `>` marks, bold trend, reverse selection) must
remain readable. Changing scheme must not require a rebuild.

Change record: [docs/superpowers/specs/2026-09-09-color-schemes.md](docs/superpowers/specs/2026-09-09-color-schemes.md).

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

To the right of trend, the daily list and the month sheet should
show a calculated delta: daily weight minus trend, in the display
unit. Color-code positive, negative, and zero. The delta is derived,
not stored. Color must not be the only way to read the sign.

Change record:
[docs/superpowers/specs/2026-09-22-delta-column.md](docs/superpowers/specs/2026-09-22-delta-column.md).

Tracking calories is a separate tool, see below.

It would be good to add user configuration to the daily logging. For
example, individuals may want to track a sleep quality number, water
intake, mood, stress, resting HR, HRV.

The TUI should make it easy to create and edit daily logs. On
startup the daily list selects the log row whose date is closest to
today, not the first row. Opening the monthly chart from the list
still uses that selected row's month.

Change record:
[docs/superpowers/specs/2026-09-17-closest-to-today-startup.md](docs/superpowers/specs/2026-09-17-closest-to-today-startup.md).

A Goto Today key should work in the daily list and in the monthly
log. In the list it selects the closest existing row to today. In
the month sheet it shows today's calendar month and focuses today's
day. It must not create today if that day has no log. That key is
not in the startup-selection slice.

Change record:
[docs/superpowers/specs/2026-09-19-goto-today.md](docs/superpowers/specs/2026-09-19-goto-today.md).

### Monthly Log

We should support the printing of blank sheets for recording daily
numbers. The TUI should make it easy to enter a whole month's data in
one sitting. Tab accepts the current cell and moves to the next, so a
row can be filled without Enter plus arrows. The TUI should make a nice
display of an entire month's data sheet.

Change record:
[docs/superpowers/specs/2026-09-10-month-tab-next-cell.md](docs/superpowers/specs/2026-09-10-month-tab-next-cell.md).

While a month-sheet cell is being edited, keys should work as a
spreadsheet: Tab already accepts and moves to the next cell. Enter
and Down should accept the edit and move down in the current column.
Up should accept the edit and move up in the current column. Left
and Right should move the caret in the text being edited, not leave
the cell. This is not the Goto Today slice.

The tool should be able to generate pdf of the monthly log sheets
(filled in, not the blank sheets mentioned above), and pdfs of the
monthly charts.

PDF monthly charts are the on-screen monthly chart on paper. CLI
`-chart-pdf YYYY-MM` and TUI `p` write that picture. Filled log-sheet
PDFs and long-term PDFs are not in this slice.

Change record:
[docs/superpowers/specs/2026-09-12-pdf-monthly-charts.md](docs/superpowers/specs/2026-09-12-pdf-monthly-charts.md).

The TUI should be able to display on screen the monthly charts and the
long term charts. Every on-screen chart (monthly and all four
long-term kinds) uses the same vertical scale: (min of plotted
values) − 2 lb through (max) + 2 lb, with the equivalent margin in
kg and stone. A monthly chart also shows Monthly Loss in the
display weight unit and Daily Deficit in calories, from the first and
last trend of the plotted span, as in the book's pencil-and-paper
analysis (3500 kcal per pound). The current month ends at today.

Monthly charts (TUI and later PDF) should match the Excel monthly
chart's floats-and-sinkers look:

*   A text box centered above the plot names the month and year
    (for example `June 1990`). In Excel that box has a red border,
    blue background, and yellow foreground. The TUI should match
    those colours as far as the terminal allows. Under `NO_COLOR`,
    the month and year stay centered and readable without colour.
*   Each daily weight is a float or sinker: a mark tied to the trend
    with a thin vertical stem (green in Excel). The stem is part of
    the look, not an extra logged series.

PDF monthly charts should use the same title box and stems, not a
stripped-down plot.

The style of the charts is described more fully [in this
section](https://www.fourmilab.ch/hackdiet/e4/signalnoise.html)

Change record (monthly, on screen):
[docs/superpowers/specs/2026-09-10-monthly-charts.md](docs/superpowers/specs/2026-09-10-monthly-charts.md).
Long-term charts and PDF layout are not in that slice.

Change record (title box and stems):
[docs/superpowers/specs/2026-09-11-monthly-chart-floats-sinkers.md](docs/superpowers/specs/2026-09-11-monthly-chart-floats-sinkers.md).

Long-term charts follow the book's WEIGHT menu: four views
(quarterly, semiannual, annual, complete history), always ending at
the latest data, with daily weight as a thin line and trend as a
thick line — not the monthly floats-and-sinkers form. `l` opens that
screen; `[` / `]` cycle the four kinds. Loss and Daily Deficit use
the same first-and-last trend identity over the window's day count.

Change record (long-term, on screen):
[docs/superpowers/specs/2026-09-12-long-term-charts.md](docs/superpowers/specs/2026-09-12-long-term-charts.md).
PDF is not in that slice.

Change record (Y range, every on-screen chart):
[docs/superpowers/specs/2026-09-12-chart-y-range.md](docs/superpowers/specs/2026-09-12-chart-y-range.md).

Long-term X labels sit at month starts as `Aug 26` (month + two-digit
year). If that will not fit, use two lines (`Aug` over `26`); if that
will not fit, skip months.

Change record (long-term X labels):
[docs/superpowers/specs/2026-09-12-long-term-x-labels.md](docs/superpowers/specs/2026-09-12-long-term-x-labels.md).

On-screen monthly and long-term charts should look at least as good
as the book's Excel charts. Hacker's Diet Online is a higher bar the
book does not cover; match it where the terminal allows. Paint may
use high-resolution glyphs so curves and stems are not one character
per day in a handful of rows. Data identity is unchanged: title box,
marks, stems, trend, Y margin, clip, Loss / Daily Deficit, `[colors]`,
16-color floor, `NO_COLOR`. PDF stays vector. The TUI chart need not
be a character-cell twin of the PDF. This is not the Goto Today
slice.

## Meal Planning

The tool should implement the meal planning system described
[in the Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/planningmeals.html#Fa147).

We should have an initial database of foods. We should have protein and
carbs in addition to calories.

We should prefer weight over volume measurements. Internally, we should
try to work consistently with grams. We will support U.S. units.

Besides the ingredients, we should save groups of ingredients as recipes
or meals. Then the user can enter 1/4 of this recipe or 245 g of that recipe.

## Documentation

User-facing documentation follows
[Diátaxis](https://diataxis.fr): tutorials, how-to guides, reference,
and explanation, kept distinct. The README is the map, not a dump of
all four. Habitat files (harness, onboarding, agent memory, dated
specs) are for people working on the code and are not part of that
split.

Markdown source follows the main points of
[Google's Markdown Style Guide](https://google.github.io/styleguide/docguide/style.html):
80-character wrap (except links, tables, headings, and code blocks);
ATX headings with a single H1; fenced code with a language; nested
lists indented 4 spaces; no trailing whitespace; informative link
text. GitHub does not honour Gitiles `[TOC]`; `../` links are allowed
because GitHub has no repo-root Markdown paths.

Shell scripts follow the main points of
[Google's Shell Style Guide](https://google.github.io/styleguide/shellguide.html):
`#!/bin/bash` and `set -euo pipefail`; 2-space indent; 80-column wrap;
quoted `"${var}"`; `$(...)` not backticks; `[[ ... ]]` not `[ ... ]`;
errors on STDERR; file header comments; `scripts/lib/*.sh` are not
executable. ShellCheck is recommended.

Go source follows the main points of
[Google's Go Style Guide](https://google.github.io/styleguide/go/guide):
`gofmt`, MixedCaps, no fixed line length, package comments, and
package `testing` only. The 80-column wrap used for Markdown and
shell does not apply to `.go` files.
