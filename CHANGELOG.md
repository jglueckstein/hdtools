# Changelog

## Unreleased

### Added

- Six unverified harness constraints from `/extract-conventions`: prefs
  out of SQLite, TUI has no SQL imports, DB/config `0600`, XDG default
  paths, no skip-to-green, no live log in git
- Convention: extract a helper when the same code is wrong in two places
  (`internal/tui/layout.go` is the model)

### Fixed

- Pin `gitleaks/gitleaks-action` to v3.0.0 so Harness Constraints CI
  can resolve the action (the old v2.3.8 SHA 404s)
- Checkout with `fetch-depth: 0` so gitleaks can see the PR commit range

### Changed

- Harness Status and README badge 8/8 → 8/14 after `/harness-audit`
  (drift yes: remaining bare `return err`, TUI `database/sql` import,
  config `0644`, convention-file lag)
