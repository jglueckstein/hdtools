#!/bin/bash
# Fail when non-test code under cmd/ or internal/ returns the
# identifier err without wrapping it.

set -euo pipefail

cd "$(dirname "$0")/.."
exec go run ./scripts/harnesscheck errors
