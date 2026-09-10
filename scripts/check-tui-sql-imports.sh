#!/bin/bash
#
# Fail if internal/tui imports database/sql or a SQLite driver.

set -euo pipefail

cd "$(dirname "$0")/.."

imports="$(go list -f '{{range .Imports}}{{.}} {{end}}' ./internal/tui)"
fail=0
for pkg in database/sql github.com/mattn/go-sqlite3 modernc.org/sqlite; do
  # shellcheck disable=SC2086  # word-split go list import paths
  if printf '%s\n' ${imports} | grep -qx "${pkg}"; then
    echo "FAIL: ./internal/tui imports ${pkg}" >&2
    fail=1
  fi
done
exit "${fail}"
