#!/bin/bash
#
# Fail if daily-log persistence mentions preference fields that belong
# in config.toml (display unit, colour scheme).

set -euo pipefail

cd "$(dirname "$0")/.."

hits="$(grep -nE 'display_unit|DisplayUnit|color_scheme|ColorScheme' \
  internal/dailylog/*.go 2>/dev/null || true)"
if [[ -n "${hits}" ]]; then
  echo "FAIL: preferences must not live in the log database:" >&2
  echo "${hits}" >&2
  exit 1
fi
exit 0
