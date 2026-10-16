#!/bin/bash
# Fail when an exported function, type, or const has no doc
# comment. The words of that comment are a review.

set -euo pipefail

cd "$(dirname "$0")/.."
exec go run ./scripts/harnesscheck docs
