#!/bin/bash
# Fail when a non-test Go file name contains an underscore, or a
# package name has an underscore or a capital. An external test
# package may end in _test.

set -euo pipefail

cd "$(dirname "$0")/.."
exec go run ./scripts/harnesscheck naming
