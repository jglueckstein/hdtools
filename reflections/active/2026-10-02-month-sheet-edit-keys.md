- **Date**: 2026-10-02
- **Agent**: tdd-agent and implementer
- **Task**: Month-sheet Enter, Down, and Up save the edit and move one
  day in the same column.
- **Surprise**: Nine of the new tests failed before production code
  changed. The rest passed. Last-day Enter already saved in place, so
  a second Enter already opened that day's form. Left and Right
  already moved the caret. Invalid Down and Up never reached the
  saver, because the text input ignored them. Those green tests are
  the locks that must stay green once Enter moves on other days.
- **Proposal**: A vertical month save sets `savedMsg.dayDelta` and
  calls `moveDay`. Tab alone sets `advance` and calls `nextCell`. Do
  not set both. An acceptance test that passes before implementation
  can still be the right test.
- **Improvement**: Tell the tdd-agent which scenarios already match
  current behaviour, so a green result is reported as a lock rather
  than a bad test.
- **Signal**: workflow
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: advocatus-diaboli, tdd-agent,
    implementer (manual)
  - Agent delegation: partial
