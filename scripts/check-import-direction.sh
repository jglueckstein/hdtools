#!/usr/bin/env bash
# Fail if domain or config packages import upward into the TUI, or across
# the declared direction. cmd/hdtools and internal/tui may import downward.
set -euo pipefail

cd "$(dirname "$0")/.."

mod="$(go list -m)"
fail=0

check() {
	local pkg="$1"
	shift
	local imports
	imports="$(go list -f '{{range .Imports}}{{.}} {{end}}' "$pkg")"
	local forbidden
	for forbidden in "$@"; do
		if printf '%s\n' $imports | grep -qx "${mod}/${forbidden}"; then
			echo "FAIL: ${pkg} imports ${mod}/${forbidden}"
			fail=1
		fi
	done
}

check ./internal/dailylog internal/tui internal/config internal/units
check ./internal/units internal/tui internal/dailylog internal/config
check ./internal/config internal/tui internal/dailylog

exit "$fail"
