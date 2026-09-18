# Changelog

## Unreleased

### Added

- Daily list opens on the log row closest to today (not the first row)
- PDF of the monthly chart: CLI `-chart-pdf YYYY-MM` (and `-o`) and
  TUI `p` on the monthly chart write a one-page landscape Letter file
- Long-term chart X labels at month starts (`Sep 90`), then two-line,
  then skip months if bins are too narrow
- Chart Y axis is plotted min/max ± 2 lb (equivalent in kg and st)
  on monthly and long-term charts
- On-screen long-term charts (`l`): quarterly, semiannual, annual,
  and complete history; trend path, daily line when the span fits
- Monthly chart title box (month and year) and green float/sinker
  stems from each daily mark to the trend
- On-screen monthly weight chart (`c`): daily marks, trend path,
  Monthly Loss, and Daily Deficit
- Go source follows Google's Go Style Guide (`gofmt`, MixedCaps, no
  fixed line length)
- Shell scripts follow Google's Shell Style Guide (`#!/bin/bash`,
  2-space indent, `"${var}"`, `[[ ... ]]`)
- Month sheet Tab accepts the current cell and moves to the next
- Product docs follow Diátaxis under `docs/` (tutorials, how-to,
  reference, explanation); README is the map
- User-facing Markdown follows Google's Markdown Style Guide (80-col
  wrap, ATX headings, fenced code with a language)
- User-configurable TUI `[colors]` in `config.toml` (16-color names,
  `0`–`15`, or `#rrggbb` / `#rgb`), silent fallback, and `NO_COLOR`
- Harness health snapshot 2026-09-10 (15/15, Trends vs 2026-09-09) and
  README Harness Health badge
- First governance audit (`observability/governance/audit-2026-09-09.md`):
  1 constraint, 0% falsifiable, Stage 3 drift, 1 debt item (score 4)
- Governance health dashboard from that audit
  (`observability/governance/governance-dashboard.html`)
- Governance constraint **User-facing copy is not medical** (agent, PR):
  TUI chrome/help, README, ONBOARDING, and idea.md must not claim
  diagnosis, treatment, prescription, or medical advice
- Promote **DB and config files are 0600** to deterministic PR enforcement
  (`go test` private-file tests; covered by existing suite, no extra CI step)
- Promote **Prefs out of SQLite** to deterministic PR enforcement
  (`scripts/check-prefs-out-of-sqlite.sh`)
- Promote **No skip-to-green** to deterministic PR enforcement
  (`scripts/check-no-skip-to-green.sh`)
- Promote **Default paths are XDG** to deterministic PR enforcement
  (`go test` path tests; covered by existing suite, no extra CI step)
- Promote **No live log in git** to deterministic PR enforcement
  (`scripts/check-no-live-log.sh`)
- Promote **TUI has no SQL imports** to deterministic PR enforcement
  (`scripts/check-tui-sql-imports.sh`); missing days are `dailylog.ErrNotFound`
- Six unverified harness constraints from `/extract-conventions`: prefs
  out of SQLite, TUI has no SQL imports, DB/config `0600`, XDG default
  paths, no skip-to-green, no live log in git
- Convention: extract a helper when the same code is wrong in two places
  (`internal/tui/layout.go` is the model)

### Fixed

- Drop the HARNESS.md pointer at a plugin spec this tree never vendored
- Document that origin exists; stop telling agents to commit on `master`
- App and View comments include the month sheet
- `layout.go` preamble states what the file does not do
- Remove unused month-sheet `cellSeed`
- Wrap remaining bare `return err` in config, form, and store
- Create new `config.toml` and SQLite log files as `0600`
- Pin `gitleaks/gitleaks-action` to v3.0.0 so Harness Constraints CI
  can resolve the action (the old v2.3.8 SHA 404s)
- Checkout with `fetch-depth: 0` so gitleaks can see the PR commit range

### Changed

- Harness Status and README badge 14/14 → 15/15 after `/harness-audit`
- Sync Cursor, Copilot, and Windsurf copies with **User-facing copy is not medical**
- Regenerate ONBOARDING.md: fifteenth constraint, and CI vs agent-review split
- Harness Status and README badge 8/8 → 14/14 after `/harness-audit`
  and promoting the extracted constraints (drift yes:
  convention-file lag)
