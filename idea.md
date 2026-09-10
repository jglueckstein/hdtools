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

Tracking calories is a separate tool, see below.

It would be good to add user configuration to the daily logging. For
example, individuals may want to track a sleep quality number, water
intake, mood, stress, resting HR, HRV.

The TUI should make it easy to create and edit daily logs.

### Monthly Log

We should support the printing of blank sheets for recording daily
numbers. The TUI should make it easy to enter a whole month's data in
one sitting. Tab accepts the current cell and moves to the next, so a
row can be filled without Enter plus arrows. The TUI should make a nice
display of an entire month's data sheet.

Change record:
[docs/superpowers/specs/2026-09-10-month-tab-next-cell.md](docs/superpowers/specs/2026-09-10-month-tab-next-cell.md).

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
