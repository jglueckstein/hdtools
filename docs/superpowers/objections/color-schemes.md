---
spec: docs/superpowers/specs/2026-09-09-color-schemes.md
date: 2026-09-09
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: "The spec does not define whether non-string or non-table [colors] values fail the entire load, so documented numeric aliases can be implemented as a startup failure that US2 exists to prevent."
    evidence: "Decision 3 lists only missing file, missing table, omitted role, unknown role key, or invalid color value as must-not-fail; S12 uses weight = \"4\" while Accepted color values lists 4 as an alias; the plan already treats a non-table [colors] as a parse error."
    disposition: accepted
    disposition_rationale: "US2 is false unless types are specified. Unquoted integers 0-15 are valid aliases. Other non-string role values fall back per role. colors that is not a table is treated as a missing table (all defaults), not a load failure. Only broken TOML or invalid display_unit fails the load."
  - id: O2
    category: specification quality
    severity: high
    claim: "The spec never states how selection composes with reverse video, with per-cell weight/trend colors, or with form focus, so FR8 and S14 can be implemented in incompatible ways."
    evidence: "Role table: \"Foreground of the selected list row or month cell, applied together with reverse video\"; S14 requires reverse, yellow foreground, and \">\" on a list row; FR8 says selection always uses reverse; Decision 7 / FR12 say form focus is \">\" plus bold, not the selection role."
    disposition: accepted
    disposition_rationale: "Match the screens that already exist. List: reverse plus optional selection foreground wraps the whole selected row and replaces inner weight/trend color on that row. Month: the same restyle on the focused cell only; > marks the day. Form: selection role unused; focus is > plus bold on the title-colored label."
  - id: O3
    category: specification quality
    severity: high
    claim: "The spec gives incompatible remainder sets for NO_COLOR, so \"honors NO_COLOR\" is not a single behaviour."
    evidence: "S7 requires no foreground or background color sequences; FR9 keeps reverse video and bold trend; Decision 6 also treats title, header, and focused-label bold as structural; US3 names pipelines as a NO_COLOR user; idea.md says NO_COLOR disables color entirely."
    disposition: accepted
    disposition_rationale: "One remainder set: non-empty NO_COLOR disables chromatic foreground and background. Bold (title, header, trend, focused label), reverse on selection, and > remain. S7 matches FR9 and Decision 6. idea.md \"entirely\" means chromatic color. Drop pipeline from US3; this is a TUI."
  - id: O4
    category: specification quality
    severity: high
    claim: "Scenarios that require a value to be \"drawn in\" a named color never define an observable, so S2 and S8 are not decidable under the project's no-TTY test rule and the spec's own exclusion of FORCE_COLOR."
    evidence: "S2 Then: \"the daily weight value is drawn in green\"; S8 Then: \"daily weight is drawn in the default blue\"; Out of scope: FORCE_COLOR; no scenario states which SGR, color profile, or TTY condition counts as success."
    disposition: accepted
    disposition_rationale: "Drawn in a named color means View() contains a foreground SGR for that ANSI index (3Xm / 9Xm or 38;5;N). Tests drive View and do not need a TTY. FORCE_COLOR stays out of scope because chroma is emitted from View() even when stdout is not a TTY."
  - id: O5
    category: specification quality
    severity: medium
    claim: "The spec constrains first-run Ensure to omit [colors] but not whether Load materializes defaults into the Config that Write already persists, so a partial scheme can be rewritten as a full table."
    evidence: "Decision 8 / FR6: Ensure writes display_unit only and must not emit [colors]; FR4: omitted roles stay on the built-in default; no FR or scenario mentions Write, round-trip, or whether resolved defaults are stored on Config."
    disposition: accepted
    disposition_rationale: "Load may apply defaults in memory for the TUI. Write persists only roles that were present and valid (sparse). Ensure/Default still omit [colors]. A partial file round-trips as a partial file. Invalid and unknown keys are dropped, not rewritten as defaults."
  - id: O6
    category: specification quality
    severity: medium
    claim: "FR7 requires the resolved scheme on list, form, and month screens, but every custom-chroma scenario is list-only, so a list-only implementation can satisfy the examples while leaving form and month on the hardcoded palette."
    evidence: "FR7: \"The TUI draws each role with the resolved scheme on list, form, and month screens\"; S2 and S14 are list-only; no scenario asserts a custom weight/trend/title color on the month sheet or form chrome."
    disposition: accepted
    disposition_rationale: "FR7 is the right requirement. Add a month-sheet chroma scenario and a form title scenario so a list-only implementation cannot pass."
  - id: O7
    category: specification quality
    severity: medium
    claim: "Invalid colors fall back with no required diagnostic, so US2's \"survive a bad scheme\" also makes a bad scheme indistinguishable from a working one."
    evidence: "US2 / Decision 3 / FR4 / S3: invalid values fall back and must not fail the load; S3 Then asserts weight uses default blue and the load succeeds, and nothing requires a warning, status line, or other signal that chartreuse was ignored."
    disposition: rejected
    disposition_rationale: "Silent fallback is Decision 3. A status or warning would be a new feature. README and ONBOARDING already must describe fallback. Record that fallback is silent at runtime; do not add a diagnostic."
---

## O1 — specification quality — high

### Claim

The spec does not define whether a `[colors]` value that is not a table, or a role value that is not a string, fails the entire config load. Because numeric aliases are documented as `4` and the only typed-invalid examples are quoted strings, an implementation can treat `weight = 4` or `colors = "red"` as a parse error and refuse to start the TUI — the lockout US2 is there to prevent.

### Evidence

Decision 3:

> A missing file, a missing `[colors]` table, an omitted role, an unknown role key, or an invalid color value must not fail config load. Invalid `display_unit` still fails.

FR4 repeats that list (missing file, missing table, omitted role, empty value, invalid color) and does not mention wrong TOML types.

Accepted color values lists `4` as an alias for `blue`. S12 is the only numeric scenario, and it uses a quoted string:

```toml
weight = "4"
title = "6"
```

S3's invalid example is also a string (`"chartreuse"`). S6 covers hex and escape *strings*. Nothing states whether an unquoted TOML integer, a boolean, an array, or a non-table `colors` key is an "invalid color value" (per-role fallback) or a load failure.

The accompanying plan already fills the silence in one direction: "A `[colors]` value that is not a table is a parse error (broken TOML shape), same as a non-string `display_unit`." That reading would fail the whole file, including a valid `display_unit`, and keep the user out of the log.

Existing `config.Load` unmarshals into a string-typed `display_unit` and returns a wrapped parse error on unmarshal failure. The same pattern applied to `[colors]` as `map[string]string` would reject `weight = 4` even though `4` is a documented alias.

### Why this matters

US2 and the Background rationale are the slice's safety claim: a color typo must not block the log. The typos people will actually type, given a table of numeric aliases in a TOML file, are unquoted numbers and a malformed `[colors]` shape. If those fail the load, the safety claim is false. If they fall back, the plan's parse-error rule is false. The spec currently licenses both.

## O2 — specification quality — high

### Claim

The spec never states how the `selection` role composes with reverse video, with per-cell `weight`/`trend` colors already on the same glyphs, or with form focus. FR8 and S14 can therefore be implemented as whole-row restyle, nested SGR, cell-only reverse, or form reverse, and still claim compliance.

### Evidence

Color roles:

> `selection` — Foreground of the selected list row or month cell, applied together with reverse video

FR8:

> Selection always uses reverse video plus the resolved `selection` foreground (none extra when omitted).

S14 (list only):

> that row is reverse video
> And its foreground is yellow
> And the `>` mark is still present

Decision 7 and FR12 say form labels use the `title` role and that focus is the `>` mark plus bold, "not a new role." S9 still calls the focused field "selected" and requires a `>` mark on list, month, and form.

The current paint paths already disagree with each other, and the spec does not pick one:

- List (`internal/tui/app.go`): `weightStyle` / `trendStyle` color cells, then `selectedStyle.Render(line)` wraps the entire row, including `>`.
- Month (`internal/tui/month.go`): `>` marks the day; `selectedStyle` wraps only the focused cell, which may already contain `weightStyle` / `trendStyle` output.
- Form (`internal/tui/form.go`): focus is `>` plus `focusLabelStyle` (bold), with no reverse and no `selection` role.

S14's three simultaneous properties (reverse, yellow foreground, `>` present) do not say whether yellow replaces inner weight/trend colors, whether reverse is SGR 7 or a fg/bg swap, or whether "foreground is yellow" after reverse means yellow glyphs or a yellow fill. Decision 6's "Foreground only" does not resolve that, because reverse is a background-affecting attribute.

### Why this matters

Selection is one of the three things `idea.md` says color exists to show. Without a composition rule, two correct implementations can show a selected weighed row as (a) a yellow reversed bar that wipes blue/red, (b) reverse of leftover inner colors with yellow ignored, or (c) reverse on a month cell while the rest of the day stays scheme-colored. FR8's "always" also licenses reverse video on form fields, which Decision 7 forbids. That is not a later styling tweak; it is the behaviour S14 claims to accept.

## O3 — specification quality — high

### Claim

The spec uses several incompatible definitions of what remains when `NO_COLOR` is set, so "honors `NO_COLOR`" is not one behaviour.

### Evidence

S7 Then:

> the view contains no foreground or background color sequences
> And column headers, values, the `>` selection mark, and help text are still present
> And the selected list row or month cell still has the `>` mark

FR9:

> When `NO_COLOR` is present and non-empty, renders emit no foreground or background color. Custom schemes are ignored for chroma. `>` marks, headers, values, help, bold trend, and reverse video on selection remain.

Decision 6:

> Bold (title, header, trend, focused form label) and reverse video (selection) are structural, not scheme keys.

US3:

> As a person in a monochrome terminal, a screen reader, or a pipeline, I want `NO_COLOR` to turn color off so structure stays readable without chromatic encoding.

`idea.md`:

> `NO_COLOR` disables color entirely; structure must remain readable.

FR9 names only trend among the bolds that survive. Decision 6 also names title, header, and focused form label. S7 forbids "background color sequences" while FR9 requires reverse video, which terminals and libraries may emit as SGR 7 or as explicit fg/bg. US3 includes pipelines, which consume leftover SGR 1 / SGR 7 as noise. `idea.md`'s "entirely" is stricter than US3's "chromatic."

### Why this matters

This slice's second feature is `NO_COLOR`. If implementers follow S7 literally, reverse may have to go. If they follow FR9, reverse stays and some "background" encodings fail S7. If they follow Decision 6, header/title/focus bold stay; if they follow FR9's list, only trend bold is required. Pipeline users named in US3 still receive ANSI. Those are different products, and the spec treats them as one acceptance bar.

## O4 — specification quality — high

### Claim

Scenarios that require a value to be "drawn in" a named color never define an observable, so S2 and S8 cannot be judged pass or fail under the project's rule that TUI tests must not require a real terminal, especially with `FORCE_COLOR` explicitly out of scope.

### Evidence

S2 Then: "the daily weight value is drawn in green" and "the trend value is drawn in yellow."

S8 Then: "daily weight is drawn in the default blue" and "trend is drawn in the default red."

S11/S12 talk about load results ("`weight` is blue") without saying whether that is an in-memory name, a lipgloss style, or bytes in `View()`.

Out of scope includes `FORCE_COLOR`. The spec never mentions TTY detection, `TERM`, color profile, or which SGR (`3Xm`, `9Xm`, `38;5;N`) counts as the named color.

Project test strategy (AGENTS.md): "Do not require a real terminal for TUI tests: drive `App.Update` / `View` with messages." Existing layout tests strip ANSI and assert geometry; they do not assert that chroma is present. Lipgloss may emit no foreground at all when stdout is not a TTY.

The plan already invents the missing observable ("Assert on SGR for foreground… treat either encoding of the same index as a pass"), which is evidence the spec does not contain one.

### Why this matters

Configurable color is the slice. If "drawn in green" is a visual claim, it is not testable as this repo tests TUIs. If it is an SGR claim, the spec does not say so, and excluding `FORCE_COLOR` without specifying a profile leaves S2/S8 environment-dependent. Implementations and tests can both be "correct" and disagree on whether the feature exists.

## O5 — specification quality — medium

### Claim

The spec says first-run `Ensure` must not emit `[colors]`, but it does not say whether `Load` fills omitted roles onto the `Config` value that `Write` already persists. A partial scheme can therefore be rewritten as a fully materialized table the next time anything saves preferences.

### Evidence

Decision 8 / FR6: `Ensure` still writes `display_unit` only; no `[colors]` table on first run. S1 repeats that the created file does not contain `[colors]`.

FR4: omitted roles stay on the built-in default in the running app.

There is no FR, decision, or scenario for `Write`, for Load→Write round-trip, or for whether the in-memory scheme is "sparse overrides" or "all nine roles resolved."

`internal/config.Write` is already the public save path ("the save path for future in-app preference edits") and is what `Ensure` calls with `Default()`. Putting resolved colors on `Config` without `omitempty` (or equivalent) would make FR6 fail; putting them on `Config` with defaults filled would make a later `Write(Load(path))` freeze today's defaults into the file and drop unknown keys that S5 says are ignored.

### Why this matters

The user-facing contract is a hand-edited sparse table: omit a key, get the built-in default, including a future default if the built-in palette changes. If Load materializes defaults and Write dumps them, omission is a one-shot property and FR4's "omitted role" behaviour cannot survive any save. FR6 only plugs the first-run hole. The persistence story for every subsequent write is unspecified.

## O6 — specification quality — medium

### Claim

FR7 requires the resolved scheme on list, form, and month screens, but every scenario that asserts a custom chromatic color is list-only. A list-only implementation can pass the spec's examples while form and month keep the hardcoded palette.

### Evidence

FR7:

> The TUI draws each role with the resolved scheme on list, form, and month screens.

S2 asserts green weight and yellow trend "when the daily log list is shown." S14 asserts custom selection foreground on "a list row." S15 checks ANSI-stripped alignment on "the list or month sheet," not that month cells actually use the custom colors.

FR12 says form labels use the `title` role, but no scenario sets `title` and then shows the form (S9 only checks `>`, headers, and bold trend). No scenario sets `weight` / `trend` and then opens the month sheet.

Today each screen has its own call sites (`listView`, `form.view`, `month.view`) over package-level styles. Missing a call site is the failure this FR exists to catch, and the scenarios would not catch it.

### Why this matters

The product already has three screens sharing one palette. If acceptance only looks at the list, the month sheet — the paper log analogue, and the place where weight vs trend is read as a table — can ship unthemed. FR7 would be an unenforceable comment.

## O7 — specification quality — medium

### Claim

Invalid colors fall back with no required diagnostic, so US2's "survive a bad scheme" also makes a bad scheme indistinguishable from a working one.

### Evidence

US2: omitted or invalid color values fall back "so a typo does not refuse to open the log."

Decision 3 / FR4 / S3: load succeeds; `weight = "chartreuse"` uses default blue; sibling `trend = "yellow"` still applies. S5 ignores unknown keys. S6 does the same for hex and sequences.

No story, scenario, or FR requires a warning, log line, status text, or other signal that a value was ignored. The Documentation section requires README/ONBOARDING to describe fallback, which is not a runtime detection path.

### Why this matters

The stated harm is "cannot open the log." The undetected harm is "opened the log with a scheme the user believes they set." `weight = "bleu"`, `titel = "magenta"`, or `weight = "chartreuse"` all look like a working config. S3 proves the load succeeded; it cannot prove the user can tell that it didn't do what they asked. That is a class of misconfiguration the spec's failure semantics make invisible by design, without recording that invisibility as a tradeoff with a detection path.

## Explicitly not objecting to

- **Roles rather than raw sequences or hex**: That matches `idea.md` and keeps the file editable without embedding terminal escapes. Truecolor/256 are recorded as out of scope, not forgotten.
- **Fallback for invalid *string* color names versus fail-on-invalid `display_unit`**: The asymmetry is an explicit product choice in both the backlog and Decision 3; the objection is the unspecified *types*, not the string-fallback policy itself.
- **`NO_COLOR` as environment-only, with empty string meaning off**: That matches no-color.org (`NO_COLOR=0` disabling color is then correct). Not having a config key is a recorded choice, not a gap.
- **First-run `Ensure` omitting `[colors]`**: A minimal file plus documented defaults is a coherent onboarding story; FR6 is specific enough for that one path.
- **Keeping scheme data out of SQLite**: That continues the existing XDG-config vs log-database split (S10 / FR13) and does not need re-litigation here.
- **Default 16-color mapping (blue weight, red trend, reverse-only selection)**: It matches the current `internal/tui/style.go` palette and the book's chart colors; this slice is configuration of that palette, not a redesign of it.
- **Form labels sharing the `title` color with focus as `>` plus bold**: Decision 7 is a closed product choice; current form labels are already cyan with bold only on focus.
- **Charts, meal planning, live reload, named presets, and background/bold as scheme keys**: These are named out of scope and are not required to make list/form/month configurable.
