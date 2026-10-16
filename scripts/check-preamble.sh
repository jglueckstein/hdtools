#!/bin/bash
# Fail when a non-test Go file has no opening comment. The comment
# may sit before the package clause or between that clause and
# the imports. What the comment explains is a review.

set -euo pipefail

cd "$(dirname "$0")/.."
exec go run ./scripts/harnesscheck preamble
