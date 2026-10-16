#!/bin/bash
#
# Fail when a CONVENTIONS.md section has no Checked by line, or when
# that line names a constraint HARNESS.md does not declare. The whole
# value "not a merge gate" is the only binding that is not a
# constraint name. A long value may continue on lines indented by
# exactly two spaces.
#
# Optional args are the conventions path and the harness path. The
# defaults are the copies at the repo root.

set -euo pipefail

cd "$(dirname "$0")/.."

conventions="${1:-CONVENTIONS.md}"
harness="${2:-HARNESS.md}"

CONVENTIONS_PATH="${conventions}" HARNESS_PATH="${harness}" python3 - <<'PY'
import os
import sys

conventions_path = os.environ["CONVENTIONS_PATH"]
harness_path = os.environ["HARNESS_PATH"]
failures = []


def fail(message):
    failures.append(message)


def read_lines(path):
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read().splitlines()
    except OSError as err:
        fail(f"FAIL: cannot read {path}: {err}")
        return []


def constraint_names(lines):
    names = []
    in_constraints = False
    for line in lines:
        is_h2 = line.startswith("## ") and not line.startswith("### ")
        if is_h2:
            if line.strip() == "## Constraints":
                in_constraints = True
                continue
            if in_constraints:
                break
        if in_constraints and line.startswith("### "):
            names.append(line[4:].strip())
    if not names:
        fail(f"FAIL: {harness_path}: no ### headings under ## Constraints")
    return set(names)


def section_headings(lines):
    headings = []
    in_fence = False
    for index, line in enumerate(lines):
        if line.strip().startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if line.startswith("## ") and not line.startswith("### "):
            headings.append((index, line[3:].strip()))
    if not headings:
        fail(f"FAIL: {conventions_path}: no ## sections")
    return headings


def binding_value(lines, heading_index):
    """Return the Checked by value, or None when the section is unbound.

    Blank lines and fenced blocks between the heading and the binding
    are skipped. A fenced ## heading is not a section.
    """
    cursor = heading_index + 1
    in_fence = False
    while cursor < len(lines):
        raw = lines[cursor]
        if raw.strip().startswith("```"):
            in_fence = not in_fence
            cursor += 1
            continue
        if in_fence or raw.strip() == "":
            cursor += 1
            continue
        break
    else:
        return None
    first = lines[cursor]
    if not first.startswith("Checked by:"):
        return None
    value = first[len("Checked by:") :].strip()
    cursor += 1
    while cursor < len(lines):
        raw = lines[cursor]
        # Exactly two spaces. A deeper indent is a nested list, not a wrap.
        if (
            raw.startswith("  ")
            and not raw.startswith("   ")
            and raw.strip() != ""
        ):
            value = f"{value.rstrip()} {raw.strip()}"
            cursor += 1
            continue
        break
    return value.strip()


harness_lines = read_lines(harness_path)
conventions_lines = read_lines(conventions_path)
if failures:
    for message in failures:
        print(message, file=sys.stderr)
    sys.exit(1)

known = constraint_names(harness_lines)
for index, title in section_headings(conventions_lines):
    line_no = index + 1
    value = binding_value(conventions_lines, index)
    if value is None:
        fail(
            f"FAIL: {conventions_path}:{line_no}: "
            f"section {title!r} has no Checked by line"
        )
        continue
    if value == "not a merge gate":
        continue
    parts = [part.strip() for part in value.split(";")]
    if any(part == "" for part in parts):
        fail(
            f"FAIL: {conventions_path}:{line_no}: "
            f"section {title!r} has an empty Checked by name"
        )
        continue
    for part in parts:
        if part == "not a merge gate":
            fail(
                f"FAIL: {conventions_path}:{line_no}: "
                f"section {title!r} mixes 'not a merge gate' "
                "with a constraint"
            )
        elif part not in known:
            fail(
                f"FAIL: {conventions_path}:{line_no}: "
                f"section {title!r} names unknown constraint {part!r}"
            )

if failures:
    for message in failures:
        print(message, file=sys.stderr)
    sys.exit(1)
PY
