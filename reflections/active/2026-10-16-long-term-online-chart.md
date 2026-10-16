- **Date**: 2026-10-16
- **Agent**: parent session
- **Task**: Draw the long-term chart as the Online picture.
- **Surprise**: ntcharts draws no line from a single point, so a lone
  weigh-in stays visible only because the painter strokes a horizontal
  rune on that cell before joining later points. The weight stroke is
  painted after the Braille trend and wins a shared cell, so a width-1
  day that is the only trend and the only weigh-in shows a line and no
  Braille. `monthShares` gives later days width 0 once the span passes
  62 columns and drops them. A six-column month label past column 56
  does not fit, so the last label slides left. A vertical join on a
  weight change is the polyline, and a stem check that treats every
  non-Braille rune as a stem flags that join.
- **Proposal**: In AGENTS.md GOTCHAS, record that a one-point ntcharts
  line draws nothing, that the weight stroke overwrites Braille on a
  shared cell, and that `monthShares` drops days past column 62.
- **Improvement**: When a slice replaces a picture, invert the tests
  that lock the old gutter before writing the painter. A one-cell
  weigh-in shows only the weight stroke, so the both-series check
  does not apply there.
- **Signal**: context
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: single agent (spec, failing tests,
    implement, pull request)
  - Agent delegation: manual
