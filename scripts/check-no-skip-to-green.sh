#!/usr/bin/env bash
# Fail on t.Skip used to hide a failing test, or a commented-out Test
# function. testing.Short() and GOOS/GOARCH skips in the same function
# are allowed.
set -euo pipefail

cd "$(dirname "$0")/.."

python3 - <<'PY'
import os, re, sys

skip_re = re.compile(r"\bt\.Skip(f|Now)?\s*\(")
commented_test = re.compile(r"^\s*//\s*func Test")
func_re = re.compile(r"^func (Test\w+)")
allow_re = re.compile(r"testing\.Short\(\)|runtime\.GOOS|runtime\.GOARCH")

fail = 0
for root, dirs, files in os.walk("."):
    dirs[:] = [d for d in dirs if d not in (".git", "vendor")]
    for name in files:
        if not name.endswith("_test.go"):
            continue
        path = os.path.join(root, name)
        with open(path, encoding="utf-8") as f:
            lines = f.readlines()
        for i, line in enumerate(lines, 1):
            if commented_test.search(line):
                print(f"FAIL: {path}:{i}: commented-out test")
                fail = 1

        current = None
        start = 0
        body = []

        def flush():
            global fail
            if current is None:
                return
            text = "".join(body)
            if skip_re.search(text) and not allow_re.search(text):
                print(
                    f"FAIL: {path}:{start}: {current} calls t.Skip "
                    "without testing.Short or GOOS/GOARCH"
                )
                fail = 1

        for i, line in enumerate(lines, 1):
            m = func_re.match(line)
            if m:
                flush()
                current = m.group(1)
                start = i
                body = [line]
            elif current is not None:
                if line.startswith("func "):
                    flush()
                    current = None
                    body = []
                else:
                    body.append(line)
        flush()

sys.exit(fail)
PY
