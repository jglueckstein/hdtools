# Plan — Color schemes and NO_COLOR

**Spec**: [`docs/superpowers/specs/2026-09-09-color-schemes.md`](../specs/2026-09-09-color-schemes.md)
**Issue**: [#22](https://github.com/jglueckstein/hdtools/issues/22)
**Status**: approved

No production code until the spec's scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/config/config.go` | Parse `[colors]`, accept 16-color names, 0–15 aliases, and `#rrggbb` / `#rgb`, keep `display_unit` failure behaviour. `Default()` and `Ensure` stay free of a `[colors]` table. |
| `internal/config/config_test.go` | Table-driven parse, fallback, case folding, numeric aliases, hex, invalid unit still errors, first-run file has no `[colors]`. |
| `internal/tui/style.go` | Replace package-level style variables with a palette built from a resolved scheme. Bold and reverse stay here, not in config. |
| `internal/tui/app.go` | Hold the palette on `App` so tests inject a scheme. List view uses it. |
| `internal/tui/form.go` | Form labels use the title role; focus is `>` plus bold. |
| `internal/tui/month.go` | Month chrome, weight, trend, selection, help, error, status use the same palette. |
| `internal/tui/style_test.go` | New: palette resolution at the TUI boundary (ANSI for a role, `NO_COLOR` strips chroma). |
| `internal/tui/layout_test.go` | Existing alignment tests must stay green with default and custom schemes. |
| `internal/tui/app_test.go` | View tests for `>` marks and `NO_COLOR` leaving structure in place. |
| `cmd/hdtools/main.go` | No behaviour change: it already passes `config.Config` into `tui.New`. |
| `README.md`, `ONBOARDING.md` | Document `[colors]`, names, hex, fallback, 16-color downshift, `NO_COLOR`. |
| `AGENTS.md` | Drop the "specified but not implemented" line once the tests are green. |
| `HARNESS.md` Context | Mention `[colors]` next to `display_unit` after implementation. |

`internal/dailylog` does not change. Prefs stay out of SQLite.

## Algorithm notes

- **Sparse on disk, resolved in the TUI.** `config.Load` keeps only
  present, valid role→color entries. The TUI fills built-in defaults
  when it builds the palette. `Write` marshals that sparse map;
  `omitempty` (or equivalent) so an empty map emits no `[colors]` table.
- **Invalid color ≠ invalid file.** Per-role: if the value is not in the
  accepted set (16-color names, `#rrggbb` / `#rgb`, integers 0–15),
  that role is omitted from the stored map (TUI uses the default) and
  load continues. Unquoted integers 0–15 are aliases. Hex without `#`,
  eight-digit hex, and SGR strings are invalid. Unknown `[colors]`
  keys are dropped. If `colors` is present but not a table, treat it as
  a missing table (all defaults). Unparseable TOML and invalid
  `display_unit` still fail.
- **Palette is data on `App`.** Package-level `var titleStyle = ...` cannot
  honour a per-process config or `NO_COLOR` in tests. Build lipgloss styles
  from the resolved scheme once in `tui.New` (and rebuild if tests construct
  an `App` with a different `Config`).
- **`NO_COLOR` is read at palette build.** A non-empty value means
  chromatic foreground/background are not applied (`3Xm`/`9Xm`/`38;`/
  `4Xm`/`10Xm`/`48;`). Bold (trend, title, header, focused label) and
  reverse (selection) stay. Do not add a config key. `View()` must still
  emit chroma when `NO_COLOR` is unset, even if stdout is not a TTY.
- **Selection composition.** List: reverse plus optional foreground wraps
  the entire selected row and replaces inner weight/trend colors.
  Month: the same restyle on the focused cell only; `>` marks the day.
  Form: selection unused; focus is `>` plus bold on the title style.
  Default selection has no extra foreground.
- **First-run marshal.** `Default()` must encode without a `[colors]` table
  (`omitempty` or equivalent) so `Ensure` does not grow the file.
- **Load→Write is sparse.** Invalid and unknown keys are not rewritten as
  defaults. Omitted roles stay omitted. The running palette still uses
  defaults for those roles. Fallback is silent (no status or warning).
- **Lipgloss profile.** Tests that assert chroma must not assume a TTY.
  Drive `View` as today. Named colors: assert 16-color `3Xm` / `9Xm` or
  256-color `38;5;N`. Hex under a truecolor profile: assert `38;2;r;g;b`
  matching the hex bytes. Hex under a 16-color profile: nearest named
  index, no `38;2;` for that cell. For `NO_COLOR`, assert no `3Xm` /
  `9Xm` / `38;` / `48;` / `4Xm` in the view. Downshift can use
  Euclidean distance in RGB to the sixteen named colors; lipgloss /
  termenv conversion is acceptable if tests pin the expected name for
  a few fixtures (`#00ff00` → `green`).
- **Do not live-reload.** Palette is fixed for the process after `New`.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1, FR2, FR3 | `TestLoadColorsTable`, `TestLoadColorNames` (table: names, aliases, case, unquoted 0–15, `#rrggbb`, `#rgb`) |
| FR4 | `TestLoadInvalidColorFallsBack`, `TestLoadMissingColorsUsesDefaults`, `TestLoadEmptyColorsTable`, `TestLoadUnknownColorKeyIgnored`, `TestLoadNonStringRoleFallsBack`, `TestLoadColorsNotATableUsesDefaults` |
| FR5 | `TestLoadRejectsUnknownUnit` (existing) plus a case with `[colors]` present |
| FR6 | `TestEnsureDoesNotWriteColorsTable` |
| FR7, FR12 | `TestViewUsesSchemeOnList`, `TestViewUsesSchemeOnMonth`, `TestFormLabelsUseTitleRole` |
| FR8, FR14 | `TestSelectionReverseWithOptionalForeground`, `TestMonthSelectionRestylesFocusedCellOnly`, `TestFormFocusNotSelectionRole` |
| FR9, FR10 | `TestNoColorStripsChromaKeepsStructure`, `TestEmptyNoColorKeepsDefaultChroma` |
| FR11 | `TestTrendIsBoldWithAndWithoutColor` |
| FR13 | covered by existing prefs-out-of-SQLite check; no dailylog change |
| FR15 | existing `TestListHeaderAlignsWithWeightAndTrend` / month equivalent, plus one run with a custom scheme |
| FR16 | `TestWriteRoundTripIsSparse` |
| FR17 | chroma assertions in the View tests (SGR index or `38;2;`, no TTY) |
| FR18 | `TestViewUsesHexOnTruecolor`, `TestHexDownshiftsOnSixteenColor`, `TestNamedAndHexMix` |

## Test case list

Config (`internal/config/config_test.go`):

1. `TestLoadMissingFileUsesKilograms` — still default unit; colors all default (extend existing).
2. `TestLoadMissingColorsUsesDefaults` — file with only `display_unit`; every role default.
3. `TestLoadEmptyColorsTable` — `[colors]` present and empty; every role default.
4. `TestLoadColorNames` — table: each canonical name, quoted and unquoted numeric alias, `gray`/`grey`, mixed case, `bright black` vs `bright-black`, `#00ff00`, `#0f0`, `#0D47A1`.
5. `TestLoadInvalidColorFallsBack` — `chartreuse`, `ff0000` (no `#`), `#00ff00ff`, escape sequence, `""`; sibling valid role still applied.
5a. `TestLoadNonStringRoleFallsBack` — `weight = true`, `trend = "yellow"`; load succeeds; weight default; trend yellow.
5b. `TestLoadColorsNotATableUsesDefaults` — `colors = "red"`; load succeeds; all roles default.
6. `TestLoadUnknownColorKeyIgnored` — `accent` dropped; `title` applied.
7. `TestLoadRejectsUnknownUnit` — still errors when `[colors]` is valid.
8. `TestEnsureDoesNotWriteColorsTable` — first-run file has no `[colors]`.
9. `TestWriteRoundTripCustomColors` — Write/Load preserves valid roles; still 0600.
9a. `TestWriteRoundTripIsSparse` — file with `weight = green`, `trend = chartreuse`, `accent = cyan`; after Load+Write the file has only `weight`; in-memory/TUI `trend` is still default red.

TUI:

10. `TestViewUsesSchemeOnList` — `weight = green`, `trend = yellow`; list view SGR matches those indexes; header still default white.
10a. `TestViewUsesSchemeOnMonth` — same scheme; month sheet weight/trend SGR match; unselected cells keep role colors.
11. `TestFormLabelsUseTitleRole` — `title = magenta`; form title and field labels contain that foreground; focused field still has `>`.
12. `TestSelectionReverseWithOptionalForeground` — `weight = green`, `selection = yellow`; selected list row is reverse with yellow foreground including the weight cell; `>` present.
12a. `TestMonthSelectionRestylesFocusedCellOnly` — weight green, trend cyan, selection yellow; focused weight cell reverse+yellow; trend cell still cyan; `>` on the day.
12b. `TestFormFocusNotSelectionRole` — title magenta, selection yellow; labels magenta, not reverse; focused field has `>` plus bold.
13. `TestNoColorStripsChromaKeepsStructure` — `NO_COLOR=1` with a custom scheme; no chromatic SGR; headers, values, `>`, reverse, and title/header/trend/focus bold present.
14. `TestEmptyNoColorKeepsDefaultChroma` — `NO_COLOR=""`; default blue weight / red trend.
15. `TestTrendIsBoldWithAndWithoutColor` — trend SGR includes bold (`1`) with color on and with `NO_COLOR`.
16. `TestListHeaderAlignsWithWeightAndTrend` — existing; still green.
17. `TestMonthHeaderAlignsWithWeightAndTrend` — existing; still green.
18. `TestListAlignsWithCustomWeightColor` — custom `weight`/`trend`; visible columns still right-aligned.
19. `TestViewUsesHexOnTruecolor` — `weight = #00ff00`, `trend = #0D47A1`; truecolor profile; `38;2;` matches those RGB bytes.
20. `TestHexDownshiftsOnSixteenColor` — `weight = #00ff00`; 16-color profile; drawn as green; no `38;2;` for that cell.
21. `TestNamedAndHexMix` — `weight = blue`, `trend = #ff0000`; truecolor profile; weight indexed blue, trend `38;2;255;0;0`.

Do not require a real terminal. Continue to drive `App.Update` / `View`.
Named colors: foreground SGR for that ANSI index (`3Xm` / `9Xm` or
`38;5;N`). Hex: `38;2;r;g;b` on truecolor, nearest named index on
16-color. `View()` must emit chroma when `NO_COLOR` is unset even if
stdout is not a TTY.
