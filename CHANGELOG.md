# Changelog

## Unreleased

### Added

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

- Harness Status and README badge 8/8 → 14/14 after `/harness-audit`
  and promoting the extracted constraints (drift yes:
  convention-file lag)
