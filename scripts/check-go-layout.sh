#!/bin/bash
# Fail when cmd/hdtools/main.go is missing, or a Go file lives
# outside cmd/hdtools, internal/, and scripts/.

set -euo pipefail

cd "$(dirname "$0")/.."
exec go run ./scripts/harnesscheck layout
