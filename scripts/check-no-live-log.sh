#!/usr/bin/env bash
# Fail if a real log database or machine config.toml is tracked.
# Fixtures under testdata/ are allowed.
set -euo pipefail

cd "$(dirname "$0")/.."

fail=0
while IFS= read -r f; do
	[ -z "$f" ] && continue
	case "$f" in
	testdata/* | */testdata/*) continue ;;
	esac
	echo "FAIL: tracked live data file: $f"
	fail=1
done < <(git ls-files | grep -E '\.(db|db-journal|db-wal|db-shm)$|(^|/)config\.toml$' || true)

exit "$fail"
