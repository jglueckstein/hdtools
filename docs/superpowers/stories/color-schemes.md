---
spec: docs/superpowers/specs/2026-09-09-color-schemes.md
date: 2026-09-10
mode: spec
cartographer_model: grok-4.6
stories:
  - id: 1
    lens: [forces, alternatives, patterns]
    title: Accept hex over a sixteen-color floor
    disposition: accepted
    disposition_rationale: "Hex is a value copied from existing themes. 16-color is the floor and the downshift target, not the ceiling. Importing a whole theme file stays out of scope."
  - id: 2
    lens: [patterns, defaults, consequences]
    title: Resolve sparse overlays at paint
    disposition: accepted
    disposition_rationale: "Sparse file plus TUI defaults is the Write contract from O5. Hex round-trips as written, not converted to names."
  - id: 3
    lens: [forces, alternatives, patterns]
    title: Honor NO_COLOR as user intent
    disposition: promoted
    disposition_rationale: "View emits chroma unless NO_COLOR is set; the process does not consult the TTY. Carry into AGENTS.md GOTCHAS. NO_COLOR strips hex chroma too."
  - id: 4
    lens: [forces, coherence, consequences]
    title: Fail open on colors, closed on units
    disposition: promoted
    disposition_rationale: "Display overlay fail-open, display_unit fail-closed, because only the unit can corrupt the series. Carry into AGENTS.md ARCH_DECISIONS. Invalid hex still fail-open."
  - id: 5
    lens: [patterns, alternatives, coherence]
    title: Keep bold and reverse structural
    disposition: accepted
    disposition_rationale: "Dual-coding is already in idea.md. Schemes recolor emphasis; they do not restyle it. Hex does not change that."
  - id: 6
    lens: [defaults, alternatives]
    title: Ratify shipping list-month-form paint
    disposition: accepted
    disposition_rationale: "This slice configures style.go, including list/month/form disagreement. Do not promote the disagreement; it is cowpath, not architecture. Default palette stays named."
  - id: 7
    lens: [forces, alternatives, defaults]
    title: Leave non-signal columns unthemed
    disposition: accepted
    disposition_rationale: "Follows daily columns are fixed. Sleep/steps/note stay terminal default until idea.md says otherwise."
  - id: 8
    lens: [alternatives, consequences]
    title: Offer no per-role chroma-off token
    disposition: accepted
    disposition_rationale: "The only chroma-off switch is global NO_COLOR. A default/none token would be a new value domain, not a clarification."
---

# Choice stories — Color schemes and NO_COLOR

## Story #1 — Accept hex over a sixteen-color floor

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Decisions 1 and 13, Accepted color values, Out of scope)
**Lens:** forces / alternatives / patterns
**Refs:** —

**Context.** `idea.md` required color that still works on a 16-color terminal. This slice originally read as names-and-aliases only. Decision 13 makes `#rrggbb` / `#rgb` a role value so a person can copy a swatch out of an existing theme, while the built-in default and the 16-color downshift target stay the sixteen named colors.

**Forces.** Theme-file matching (truecolor literals people already have) against the 16-color floor, against keeping `[colors]` a hand-typed table of roles rather than a second language of sequences or a theme-file parser.

**Options not taken.** Keep 16-color names as the exclusive value set. Accept 256-color indexes above 15. Parse Alacritty, iTerm, or VS Code theme files. Always emit truecolor and let the terminal approximate.

**Choice as written.** Hex is a color *value*, not a raw sequence and not an imported theme. Importing a whole theme file is out of scope; the user copies hex into `[colors]` keys. Named 16-color values still render as indexed SGR. Hex renders as `38;2;` on a truecolor profile and as the nearest of the sixteen names on a 16-color profile.

**Consequences.** One table may mix names and hex. The same file paints differently by profile. Sixteen named colors are the default vocabulary, the downshift target, and the compatibility floor — not the ceiling. Theme-file import would be a later slice.

**Pattern.** Graceful degradation: truecolor when the profile allows, named floor otherwise. Copy-and-bind rather than import.

## Story #2 — Resolve sparse overlays at paint

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Decisions 8–9, FR4, FR6, FR16, S1, S16)
**Lens:** patterns / defaults / consequences
**Refs:** O5

**Context.** First-run `Ensure` still writes `display_unit` only. Omitted or invalid roles use the built-in default in the running app. Writing a loaded config persists only roles that were present and valid.

**Forces.** A hand-editable minimal file against a materialized nine-key table that would freeze today's defaults into every subsequent save and grow the first-run config.

**Options not taken.** Fill all nine roles onto `Config` at `Load` so `Write` dumps a complete table. Emit a commented `[colors]` template from `Ensure`. Store a resolved scheme in config and a sparse map beside it.

**Choice as written.** Sparse on disk, complete at paint. `Load` keeps present valid role→color entries; the TUI applies built-in defaults when it builds the palette. Invalid values and unknown keys are dropped, not rewritten as defaults. A partial file round-trips as a partial file.

**Consequences.** Omitted keys follow a future default-palette change. `Load`→`Write` is lossy for unknown keys, invalid values, and (under canonical marshal) comments. The config package owns the overlay; the TUI owns the complete palette.

**Pattern.** Sparse overlay / late binding: defaults live at the consumer, not in the store. The plan names this; the spec describes it without naming it.

## Story #3 — Honor NO_COLOR as user intent

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Decision 4, Observing color, FR9, FR10, FR17, Out of scope)
**Lens:** forces / alternatives / patterns
**Refs:** O3, O4

**Context.** A non-empty `NO_COLOR` disables chromatic foreground and background and wins over any scheme. Empty or unset leaves the resolved scheme in place. `View()` still emits chroma when `NO_COLOR` is unset, even if stdout is not a TTY. `FORCE_COLOR` and a config key are out of scope.

**Forces.** The user said no chroma (no-color.org) against terminal capability (`isatty`, `TERM`, color profile) against tests and captured output that must observe SGR without a TTY. `idea.md`'s "entirely" against keeping structure.

**Options not taken.** Strip color when stdout is not a TTY. Honor `FORCE_COLOR`. Add `no_color` to `config.toml`. Treat `NO_COLOR=0` as enabling color. Strip bold and reverse along with chroma.

**Choice as written.** `NO_COLOR` is environment intent, not capability detection. Profile still chooses how hex is encoded (truecolor versus downshift); that is a different axis from whether chroma is allowed at all.

**Consequences.** Non-TTY and test `View()` output still contains SGR when the user did not opt out. `NO_COLOR=0` disables color. Hex chroma is stripped with named chroma. There is no override that restores color while `NO_COLOR` is set.

**Pattern.** [no-color.org](https://no-color.org/): intent over capability. The spec uses both this axis and profile-based downshift (#1) without naming the distinction.

## Story #4 — Fail open on colors, closed on units

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Background, Decisions 3, 10, 11, FR4, FR5, S3, S4)
**Lens:** forces / coherence / consequences
**Refs:** O1, O7

**Context.** Invalid `display_unit` still fails the load so a typo cannot store pounds as kilograms. Omitted, unknown, empty, or invalid color values — including wrong TOML types and a non-table `colors` — must not fail the load. Fallback produces no error, status, or warning.

**Forces.** Never lock someone out of the log over a color typo against never persisting a wrong unit. Availability against diagnosis.

**Options not taken.** Fail the load on any invalid color, symmetric with units. Surface a warning or status line on fallback (rejected as a new feature in O7). Fail-open `display_unit` too. Treat a non-table `colors` as broken TOML.

**Choice as written.** Colors fail-open and silent; `display_unit` and unparseable TOML fail-closed. Unquoted integers 0–15 are aliases; any other role type is per-role fallback; `colors` that is not a table is a missing table.

**Consequences.** Two keys in one file, two policies — the unit error path cannot be reused for colors. Detection of a bad scheme is documentation plus the visible default palette. A later in-app editor that called `Write` after `Load` would already have dropped the bad keys (#2).

**Pattern.** Fail-open versus fail-closed by blast radius: aesthetic overlay versus stored measurement.

## Story #5 — Keep bold and reverse structural

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Decisions 4 and 6, Observing color, Out of scope)
**Lens:** patterns / alternatives / coherence
**Refs:** O3

**Context.** Each `[colors]` key is a foreground. Bold on title, header, trend, and focused form label, and reverse on selection, are not scheme keys. Background colors, configurable bold, reverse, and underline are out of scope.

**Forces.** Color must not be the only cue against theming "the whole look." `NO_COLOR` needs a remainder set that still marks selection, trend, and focus after chroma is gone.

**Options not taken.** Background as scheme keys. Per-role `bold` / `reverse` flags. Selection as a background color instead of reverse. Trend distinguished by color only.

**Choice as written.** Scheme keys are chromatic foregrounds. Emphasis is code, not config. Selection's optional foreground sits on reverse, not instead of it. Reverse (`7`) and bold (`1`) are defined as structural, not chromatic.

**Consequences.** A file cannot un-bold trend, un-reverse selection, or paint a role's background. `NO_COLOR` and a custom scheme share the same skeleton. Coheres with #8: chroma is optional globally, emphasis is not optional at all.

**Pattern.** Semantic color versus structural emphasis — chrominance (`3x` / `4x` / `38;` / `48;`) separated from attributes (`1`, `7`).

## Story #6 — Ratify shipping list-month-form paint

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Decision 12, Built-in default, FR7, FR8, FR12, S14, S17, S18, S21, S22)
**Lens:** defaults / alternatives
**Refs:** O2, O6

**Context.** The TUI already paints three screens from package-level styles in `internal/tui/style.go`, `app.go`, `month.go`, and `form.go`. The spec does not redesign that paint. It names it as the product: one resolved scheme on list, form, and month, composed the way those screens already compose.

**Forces.** A single `selection` role against three widgets that already disagree (whole-row list, focused-cell month, `>`-and-bold form). Consistency against matching the screens that exist.

**Options not taken.** Unify month selection to whole-row reverse. Apply the selection role to form focus. Redesign the default palette. Leave form and month on hardcoded styles while only the list reads config.

**Choice as written.** Inherit shipping composition and the shipping 16-color default (cyan title, blue weight, red trend, reverse-only selection, form labels on `title`). List: reverse plus optional selection foreground wraps the entire selected row and replaces inner weight/trend. Month: the same restyle on the focused cell only; other cells on that day keep role colors; `>` marks the day. Form: selection unused; focus is `>` plus bold on the title-colored label.

**Consequences.** "Selection" is not the same paint on every screen. A custom `selection` color never appears on the form. The default palette is a freeze of `style.go` and the book's chart colors, not a new look. Acceptance has to cover all three screens.

**Pattern.** Spec from the running system: the existing UX is the composition rule.

## Story #7 — Leave non-signal columns unthemed

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Color roles, idea.md TUI / color, Out of scope)
**Lens:** forces / alternatives / defaults
**Refs:** #6

**Context.** The role table themes chrome (`title`, `muted`, `header`, `help`), the book's two series (`weight`, `trend`), selection, and `error` / `status`. Date, weekday, sleep, steps, workout, note, and the `>` mark have no keys. Per-column schemes are out of scope.

**Forces.** The Hacker's Diet signal/noise split (trend versus noisy daily weight) against a fully themed table. Nine hand-edited keys is already a surface.

**Options not taken.** A catch-all `data` role for remaining columns. Per-column keys for sleep, steps, workout, note. A role for the `>` mark.

**Choice as written.** By the role table and by silence, only chrome, the two series, selection, and status/error are themable. Remaining glyphs use the terminal default foreground, except when list whole-row selection (#6) restyles the entire line.

**Consequences.** A scheme cannot tint sleep or steps to match weight. On the list, a selected row still paints those unthemed columns with the selection foreground. On the month sheet, unselected cells in those columns stay unthemed.

**Pattern.** Highlight the signal: chrome plus the book's two series plus selection. Suspicion: signal-versus-noise from the source text, applied to which columns get keys.

## Story #8 — Offer no per-role chroma-off token

**Source:** `docs/superpowers/specs/2026-09-09-color-schemes.md` (Accepted color values, Decisions 4 and 10, FR4, FR9)
**Lens:** alternatives / consequences
**Refs:** #3, #5

**Context.** Empty string after trim is invalid and falls back to that role's built-in *color*, not to "no color." There is no `none`, `default`, or `transparent`. `0` is black. `NO_COLOR` strips all chromatic foreground and background. Omitting a role uses the default color, except `selection`, whose default is reverse with no extra foreground.

**Forces.** Every role always has a defined paint against letting someone untheme one role (uncolored weight, colored trend) without killing all chroma.

**Options not taken.** Accept `none` / `default` as the terminal foreground. Treat empty string as unthemed rather than invalid. A per-role mute. Allow `weight = ""` to mean inherit.

**Choice as written.** The spec chose no chroma-off token by defining empty and unknown as fallback-to-default-color. The only way off is global `NO_COLOR`. The only "no extra" is omitting `selection`.

**Consequences.** Monochrome-weight / colored-trend taste cannot be expressed in the file. Turning one role off requires either `NO_COLOR` (too much) or a color that happens to match the terminal (fragile). Coheres with fail-open (#4): empty is a typo, not a request.

**Pattern.** All-or-nothing feature flag for chroma. Suspicion: the null object is the default palette, not transparency.
