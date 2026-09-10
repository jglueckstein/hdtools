# Compound Learning

<!-- This file is the project's persistent memory across AI sessions.
     Review new entries with the same scepticism you would apply to any
     generated content. Entries should reflect observed reality. -->

## STYLE

- Co-located `*_test.go` next to the code they cover — that is how
  `internal/dailylog` and `internal/tui` are already laid out.
- Table-driven tests for validation and conversion matrices (`units`,
  `dailylog` constructors).
- Pad TUI columns with `lipgloss.Width` (`visPad`), never byte length,
  once ANSI color is in the string.

## GOTCHAS

- `gofmt -l .` exits 0 even when it prints paths. CI and humans must
  treat non-empty output as failure (`test -z "$(gofmt -l .)"`).
- `ApplyTrend` needs the full chronological series. Loading only the
  current month drops carry-forward from the previous month.
- Display-unit conversion (`units.FromKG`) is for the TUI only. Saving
  must go through `units.ToKG` so SQLite always stores kilograms.
- `View()` emits chroma unless `NO_COLOR` is non-empty; the process does
  not consult the TTY. Tests that need color must `t.Setenv("NO_COLOR",
  "")`. Hex is `38;2;` on truecolor and downshifts on a 16-color profile.

## ARCH_DECISIONS

- Decision: store body weight in kilograms; kg/lb/st are display-only
  in `config.toml`. Reason: a unit change must not rewrite history.
  Alternatives: store pounds (rejected — SI and meal-planning grams);
  store the display unit in each row (rejected — mixed series).
- Decision: do not persist trend. Reason: a backdated weight edit would
  desync a stored moving average. `ApplyTrend` is the source of truth.
- Decision: config file under XDG, database under XDG data, not
  `~/.hdtools`. Reason: spec-compliant Unix paths; config is
  hand-editable and must not live in SQLite.
- Decision: `[colors]` fail-open, `display_unit` fail-closed. Reason:
  a bad color cannot corrupt the log series; a bad unit would store
  pounds as kilograms. Invalid or omitted colors are dropped silently
  and the TUI uses the built-in default.

## TEST_STRATEGY

- Unit tests live beside source as `_test.go`.
- Use `t.TempDir()` for SQLite; `:memory:` races under `t.Parallel`.
- Alignment tests strip ANSI and compare header/value column ends
  (`internal/tui/layout_test.go`).
- Do not require a real terminal for TUI tests: drive `App.Update` /
  `View` with messages.

## DESIGN_DECISIONS

- Daily log columns are fixed: weight, sleep, steps, workout, note.
  No extras bag until `idea.md` says otherwise.
- Workout is a boolean, not an exercise rung.
- Default display unit is kg until `config.toml` says otherwise.
- `[colors]` is a sparse overlay of named 16-color or hex values.
  16-color is the floor and downshift target. `NO_COLOR` strips
  chromatic color only; bold, reverse, and `>` remain.
- User-facing docs follow Diátaxis under `docs/` (tutorials, how-to,
  reference, explanation). README is the map. Habitat files are not
  that split.
- User-facing Markdown follows Google's Markdown Style Guide: 80-col
  wrap (except links, tables, headings, code); ATX headings; one H1;
  fenced code with a language; 4-space nested lists; no trailing
  whitespace. No Gitiles `[TOC]`; `../` links are allowed on GitHub.
- Shell scripts follow Google's Shell Style Guide: `#!/bin/bash`,
  `set -euo pipefail`, 2-space indent, `"${var}"`, `[[ ... ]]`,
  errors on STDERR. `scripts/lib/*.sh` are not executable.
