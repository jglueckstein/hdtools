# Color schemes and NO_COLOR

**Date**: 2026-09-09
**Status**: draft
**Issue**: [#22](https://github.com/jglueckstein/hdtools/issues/22)
**Backlog**: [`idea.md`](../../../idea.md) (TUI / color)

The TUI already paints chrome, daily weight, and trend in a built-in
16-color palette. This change makes that palette configurable from the
same config file as display units, accepts 16-color names and hex
values copied from existing themes, and honors `NO_COLOR`, without a
rebuild. Charts, meal planning, importing a whole theme file, and an
in-app picker are out of scope.

## Background

`idea.md` requires color so the eye can tell chrome from data, the
selected row from the rest, and the trend from noisy daily weights. Color
must still work on a 16-color terminal and must not be the only way to
read a value.

Today those colors are hardcoded in the TUI. `config.toml` only knows
`display_unit`. Invalid display units fail the load; `idea.md` treats
invalid colors the opposite way: omitted or invalid colors fall back to
the built-in default so a typo cannot keep someone out of the log.

## User stories

### US1 — Configure a scheme without rebuilding

As a person using hdtools, I want to set colors for TUI roles in
`config.toml` using 16-color names or hex values from an existing
theme so the screen matches my terminal and taste without rebuilding
the binary.

### US2 — Survive a bad or partial scheme

As a person editing config by hand, I want omitted or invalid color
values to fall back to the built-in default so a typo does not refuse to
open the log.

### US3 — Disable color from the environment

As a person in a monochrome terminal or using a screen reader, I want
`NO_COLOR` to turn chromatic color off so structure stays readable
without chromatic encoding.

### US4 — Read values without relying on color

As a person who cannot or will not use color, I want selection, column
identity, and trend versus daily weight to remain readable from marks,
headers, and emphasis that are not color.

## Decisions

These are product choices for this slice, recorded so they are not
re-litigated in the plan.

1. **Roles, not sequences.** The config names *roles*. A value is a
   16-color name (or 0–15 alias) or a hex color (`#rrggbb` / `#rgb`),
   not a raw terminal sequence. Hex exists so a person can copy a
   color out of an existing theme file. The built-in default stays
   the named 16-color palette.
2. **One table in the existing file.** Scheme lives under `[colors]` in
   the same TOML file as `display_unit`. Not a second file, not SQLite.
3. **Fallback, do not fail.** A missing file, a missing `[colors]`
   table, an omitted role, an unknown role key, or an invalid color
   value must not fail config load. Invalid `display_unit` and
   unparseable TOML still fail.
4. **`NO_COLOR` is environment-only.** It is not a config key. A
   non-empty `NO_COLOR` wins over any scheme. It disables chromatic
   foreground and background. Bold (title, header, trend, focused
   label), reverse on selection, and `>` remain. `idea.md`'s "entirely"
   means chromatic color.
5. **Restart, not rebuild, not live reload.** Editing the file and
   starting the app again is enough. Watching the file is out of scope.
6. **Foreground only.** Bold (title, header, trend, focused form label)
   and reverse video (selection) are structural, not scheme keys.
7. **Form labels share `title`.** Unfocused and focused form labels use
   the `title` role. Focus is the `>` mark plus bold, not a new role.
8. **First-run file stays minimal.** `Ensure` still writes
   `display_unit` only. No `[colors]` table is required or emitted on
   first run.
9. **Write is sparse.** The running app applies built-in defaults for
   omitted or invalid roles. Writing a loaded config persists only roles
   that were present and valid. Invalid values and unknown keys are
   dropped, not rewritten as defaults. A partial file round-trips as a
   partial file.
10. **Fallback is silent.** Invalid or unknown color values produce no
    error, status, or warning. Detection is the documentation and the
    visible default palette. Availability beats diagnosis.
11. **Value types.** A role value may be a string or an integer 0–15
    (unquoted TOML integer is the numeric alias). Any other type for a
    role is an invalid color for that role: per-role fallback, load
    succeeds. If `colors` is present but not a table, treat it as a
    missing table (all defaults); do not fail the load.
12. **Selection composition.** List: reverse plus optional selection
    foreground wraps the entire selected row and replaces inner
    weight/trend colors on that row. Month: the same restyle applies
    only to the focused cell; other cells on that day keep their role
    colors; `>` marks the day. Form: the selection role is unused;
    focus is `>` plus bold on the title-colored label.
13. **Hex is a value, not a sequence.** `#rrggbb` and `#rgb` (case
    insensitive) are accepted. Eight-digit hex, hex without `#`, and
    raw SGR are invalid. On a truecolor-capable profile, hex renders
    as given. On a 16-color profile, hex maps down to the nearest of
    the sixteen named colors so a 16-color terminal still works.

## Color roles

The `[colors]` table uses these keys. Each is a foreground for that
role.

| Key | Where it appears |
| --- | --- |
| `title` | Screen titles, and form field labels |
| `muted` | Database path, empty-state copy, form hints |
| `header` | Column headers on the list and month sheet |
| `help` | Keybinding lines |
| `weight` | Daily weight values |
| `trend` | Trend values |
| `selection` | Optional extra foreground used with reverse video on the selected list row or the focused month cell (see Decision 12) |
| `error` | Error lines |
| `status` | Status lines |

Unknown keys in `[colors]` are ignored.

### Built-in default

This is the 16-color palette the app already ships, matching the book's
charts (blue daily weight, red trend):

| Role | Default color |
| --- | --- |
| `title` | `cyan` |
| `muted` | `bright-black` |
| `header` | `white` |
| `help` | `bright-black` |
| `weight` | `blue` |
| `trend` | `red` |
| `selection` | no extra foreground (reverse video only) |
| `error` | `red` |
| `status` | `green` |

Omitting `selection` keeps reverse video with no extra foreground.

### Accepted color values

Names are case-insensitive. Hyphens and spaces are equivalent
(`bright-black`, `bright black`). Canonical names:

| Name | Also accepted |
| --- | --- |
| `black` | `0` |
| `red` | `1` |
| `green` | `2` |
| `yellow` | `3` |
| `blue` | `4` |
| `magenta` | `5` |
| `cyan` | `6` |
| `white` | `7` |
| `bright-black` | `8`, `gray`, `grey` |
| `bright-red` | `9` |
| `bright-green` | `10` |
| `bright-yellow` | `11` |
| `bright-blue` | `12` |
| `bright-magenta` | `13` |
| `bright-cyan` | `14` |
| `bright-white` | `15` |

Hex values are accepted in these forms only, case-insensitive:

| Form | Example |
| --- | --- |
| `#rrggbb` | `#00ff00`, `#0D47A1` |
| `#rgb` | `#0f0` (same as `#00ff00`) |

Unquoted integers `0` through `15` are the same aliases as the quoted
strings `"0"` through `"15"`.

Anything else — including empty string after trim, hex without `#`
(`ff0000`), eight-digit hex (`#00ff00ff`), 256-color indexes above 15,
escape sequences, booleans, floats, arrays, and integers outside 0–15 —
is invalid for that role and falls back to that role's default.

## Observing color

A value is **drawn in** a named 16-color when `View()` contains a
foreground SGR for that color's ANSI index: 16-color `3Xm` / `9Xm`, or
256-color `38;5;N` with `N` equal to the index. Either encoding counts.

A value is **drawn in** a hex color when, under a truecolor profile,
`View()` contains a foreground truecolor SGR `38;2;r;g;b` whose
components match the hex (either decimal encoding of the bytes). Tests
that assert hex use a truecolor profile and do not require a TTY.

A hex value **downshifts** when the profile is 16-color: `View()` then
contains the named-color SGR of the nearest of the sixteen accepted
names, not `38;2;`.

Tests drive `App.Update` / `View` and must not require a TTY. When
`NO_COLOR` is unset or empty, `View()` emits those sequences even if
stdout is not a TTY.

**Chromatic** sequences are foreground and background color: `3Xm`,
`9Xm`, `38;`, `4Xm`, `10Xm`, `48;`. Reverse (`7`) and bold (`1`) are
structural, not chromatic. `FORCE_COLOR` is out of scope.

## Acceptance scenarios

### S1 — First run uses the built-in scheme

**Given** no `config.toml` exists, or it exists with `display_unit` only
**When** the app starts
**Then** titles, headers, help, daily weight, trend, errors, and status
use the built-in default colors
**And** the selected row or cell is reverse video
**And** the file created on first run does not contain a `[colors]` table

### S2 — A custom role color is used

**Given** a config file with

```toml
display_unit = "kg"

[colors]
weight = "green"
trend = "yellow"
```

**When** the daily log list is shown with a weighed day
**Then** the daily weight value is drawn in green
**And** the trend value is drawn in yellow
**And** unspecified roles still use the built-in defaults

### S3 — One invalid color does not fail the load

**Given** a config file with

```toml
display_unit = "kg"

[colors]
weight = "chartreuse"
trend = "yellow"
```

**When** the config is loaded
**Then** the load succeeds
**And** `weight` uses the default (`blue`)
**And** `trend` uses `yellow`

### S4 — Invalid display unit still fails

**Given** a config file with `display_unit = "stone"` and any `[colors]`
table
**When** the config is loaded
**Then** the load fails
**And** the process does not start the TUI

### S5 — Unknown role keys are ignored

**Given** a config file whose `[colors]` table includes `accent = "cyan"`
and a valid `title = "magenta"`
**When** the config is loaded
**Then** the load succeeds
**And** `title` uses magenta
**And** `accent` has no effect

### S6 — Sequences and malformed hex are invalid

**Given** a config file with `weight = "\\x1b[34m"`, or `weight = "ff0000"`
(no `#`), or `weight = "#00ff00ff"` (eight-digit)
**When** the config is loaded
**Then** the load succeeds
**And** `weight` uses the default (`blue`)

### S7 — `NO_COLOR` disables chromatic color

**Given** `NO_COLOR` is set to a non-empty value
**And** the config may or may not contain a `[colors]` table
**When** any screen is rendered
**Then** the view contains no chromatic foreground or background
sequences (see Observing color)
**And** column headers, values, the `>` selection mark, and help text
are still present
**And** the selected list row or month cell still has the `>` mark
**And** reverse video on selection remains
**And** title, header, trend, and focused form label remain bold

### S8 — Empty `NO_COLOR` does not disable color

**Given** `NO_COLOR` is unset or set to an empty string
**And** no `[colors]` table is present
**When** the daily log list is shown with a weighed day
**Then** daily weight is drawn in the default blue
**And** trend is drawn in the default red

### S9 — Color is not the only cue

**Given** color is enabled or disabled
**When** the list, month sheet, or form is shown
**Then** the selected row, day, or field is marked with `>`
**And** columns are identified by header labels
**And** trend remains bold relative to daily weight

### S10 — Scheme is not stored in the log database

**Given** a config file with a custom `[colors]` table
**When** days are saved
**Then** the SQLite log contains no color or scheme columns or values

### S11 — Case-insensitive color names

**Given** a config file with `weight = "Blue"` and `trend = "BRIGHT-RED"`
**When** the config is loaded
**Then** `weight` is blue
**And** `trend` is bright-red

### S12 — Numeric aliases

**Given** a config file with `weight = "4"` and `title = "6"`
**When** the config is loaded
**Then** `weight` is blue
**And** `title` is cyan

### S12a — Unquoted integer aliases

**Given** a config file with unquoted `weight = 4` and `title = 6`
**When** the config is loaded
**Then** `weight` is blue
**And** `title` is cyan
**And** the load succeeds

### S13 — Empty `[colors]` table is all defaults

**Given** a config file that contains an empty `[colors]` table
**When** the config is loaded
**Then** every role uses the built-in default

### S14 — Custom selection foreground keeps reverse video

**Given** a config file with

```toml
[colors]
weight = "green"
selection = "yellow"
```

**When** a list row with a weight is selected
**Then** that row is reverse video
**And** the row's foreground is yellow, including the weight cell
**And** the `>` mark is still present

### S15 — Column alignment survives custom colors

**Given** a custom scheme that colors `weight` and `trend`
**When** the list or month sheet is shown
**Then** the visible (ANSI-stripped) weight and trend values still
right-align under their headers

### S16 — Write keeps a partial scheme partial

**Given** a config file with

```toml
display_unit = "kg"

[colors]
weight = "green"
trend = "chartreuse"
accent = "cyan"
```

**When** the config is loaded and then written back to the same path
**Then** the file's `[colors]` table contains `weight = "green"`
**And** it does not contain `trend` (invalid, dropped)
**And** it does not contain `accent` (unknown, dropped)
**And** it does not contain `title` or other omitted roles
**And** the running app still uses default `trend` (`red`)

### S17 — Month sheet uses custom weight and trend

**Given** a config file with `weight = "green"` and `trend = "yellow"`
**When** the month sheet is shown with a weighed day
**Then** that day's daily weight value is drawn in green
**And** that day's trend value is drawn in yellow

### S18 — Form labels use the title role

**Given** a config file with `title = "magenta"`
**When** the day form is shown
**Then** the form title and field labels are drawn in magenta
**And** the focused field is marked with `>`

### S19 — Non-string role value falls back

**Given** a config file with `weight = true` and `trend = "yellow"`
**When** the config is loaded
**Then** the load succeeds
**And** `weight` uses the default (`blue`)
**And** `trend` uses `yellow`

### S20 — `colors` that is not a table is a missing table

**Given** a config file with `display_unit = "kg"` and `colors = "red"`
**When** the config is loaded
**Then** the load succeeds
**And** every role uses the built-in default

### S21 — Month selection restyles only the focused cell

**Given** a config file with `weight = "green"`, `trend = "cyan"`, and
`selection = "yellow"`
**When** the month sheet is shown focused on the weight cell of a
weighed day
**Then** that weight cell is reverse video with yellow foreground
**And** the trend cell on that day is still drawn in cyan
**And** the `>` mark is on that day

### S22 — Form focus does not use the selection role

**Given** a config file with `title = "magenta"` and `selection = "yellow"`
**When** the day form is shown
**Then** field labels are drawn in magenta
**And** labels are not reverse video
**And** the focused field is marked with `>` plus bold

### S23 — Hex from a theme is used

**Given** a config file with

```toml
[colors]
weight = "#00ff00"
trend = "#0D47A1"
```

**When** the daily log list is shown with a weighed day under a
truecolor profile
**Then** the daily weight value is drawn in `#00ff00`
**And** the trend value is drawn in `#0D47A1`

### S24 — Short hex and mixed case

**Given** a config file with `weight = "#0f0"` and `trend = "#c22"`
**When** the config is loaded and the list is shown under a truecolor
profile
**Then** `weight` is `#00ff00`
**And** `trend` is `#cc2222`

### S25 — Hex downshifts on a 16-color profile

**Given** a config file with `weight = "#00ff00"`
**When** the daily log list is shown with a weighed day under a
16-color profile
**Then** the daily weight value is drawn in green (the nearest of the
sixteen named colors)
**And** the view does not contain a truecolor `38;2;` sequence for
that cell

### S26 — Named and hex roles mix in one table

**Given** a config file with `weight = "blue"` and `trend = "#ff0000"`
**When** the daily log list is shown with a weighed day under a
truecolor profile
**Then** daily weight is drawn in blue
**And** trend is drawn in `#ff0000`

## Functional requirements

- **FR1.** Config load reads an optional `[colors]` table from the same
  TOML file as `display_unit`.
- **FR2.** The supported role keys are exactly those in Color roles.
  Unknown keys are ignored.
- **FR3.** Each present, valid color value becomes that role's
  foreground. Accepted values are those in Accepted color values:
  16-color names, unquoted integers 0–15, and hex `#rrggbb` / `#rgb`.
- **FR4.** A missing file, missing table, omitted role, empty value,
  invalid color, or non-string/non-integer role value leaves that role
  on the built-in default and does not fail the load. If `colors` is
  present but not a table, treat it as a missing table. The fallback is
  silent: no error, status, or warning.
- **FR5.** Invalid `display_unit` still fails the load even if `[colors]`
  is valid.
- **FR6.** First-run `Ensure` writes a file that does not include a
  `[colors]` table.
- **FR7.** The TUI draws each role with the resolved scheme on list,
  form, and month screens.
- **FR8.** Selection uses reverse video plus the resolved `selection`
  foreground (none extra when omitted), composed as in Decision 12:
  whole selected list row, focused month cell only, unused on the form.
- **FR9.** When `NO_COLOR` is present and non-empty, renders emit no
  chromatic foreground or background (Observing color). Custom schemes
  are ignored for chroma. `>` marks, headers, values, help, reverse
  video on selection, and bold on title, header, trend, and focused
  form label remain.
- **FR10.** When `NO_COLOR` is unset or empty, the resolved scheme is
  used, including the built-in default.
- **FR11.** Trend is bold whether or not color is enabled.
- **FR12.** Form field labels use the `title` role; the focused field
  adds bold and a `>` mark.
- **FR13.** No color or scheme data is written to SQLite.
- **FR14.** Changing a scheme does not require a rebuild; starting the
  process again after editing the file is enough.
- **FR15.** Visible column alignment of weight and trend is unchanged
  when ANSI color is present.
- **FR16.** Writing a loaded config persists only color roles that were
  present and valid in the file. Omitted roles, invalid values, and
  unknown keys are not written. The running app still uses built-in
  defaults for those roles.
- **FR17.** Color in acceptance scenarios is observed as defined in
  Observing color. `View()` emits those sequences without a TTY when
  `NO_COLOR` is unset or empty.
- **FR18.** A valid hex value renders as truecolor (`38;2;r;g;b`) on a
  truecolor profile and as the nearest of the sixteen named colors on
  a 16-color profile. Named 16-color values still render as indexed
  SGR. `NO_COLOR` strips hex chroma as well.

## Out of scope

- In-app color picker or live reload while the process is running
- Importing a whole theme file (Alacritty, iTerm, VS Code, and so on);
  the user copies hex values into `[colors]` keys
- Named scheme presets beyond the single built-in default
- 256-color indexes above 15 as config values
- Background colors as scheme keys
- Configurable bold, reverse, or underline
- A `NO_COLOR` or `force_color` key in `config.toml`
- `FORCE_COLOR`
- Per-screen or per-column schemes
- Charts, PDF, and meal planning color
- Remote database

## Documentation

README and ONBOARDING must describe the `[colors]` keys, 16-color
names, hex (`#rrggbb` / `#rgb`), silent fallback (invalid values do
not fail the load and do not warn), 16-color downshift of hex, and
`NO_COLOR`. They must not present the app as medical advice.
