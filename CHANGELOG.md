# Changelog

## Unreleased

### Added

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

- Pin `gitleaks/gitleaks-action` to v3.0.0 so Harness Constraints CI
  can resolve the action (the old v2.3.8 SHA 404s)
- Checkout with `fetch-depth: 0` so gitleaks can see the PR commit range

### Changed

- Harness Status and README badge 8/8 → 13/14 after `/harness-audit`
  and promoting TUI SQL imports, no-live-log, XDG defaults,
  no-skip-to-green, and prefs-out-of-SQLite (drift yes: remaining
  bare `return err`, config `0644`, convention-file lag)
