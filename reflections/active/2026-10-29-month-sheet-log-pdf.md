- **Date**: 2026-10-29
- **Agent**: parent session
- **Task**: Add shift-P on the idle month sheet for the filled log
  PDF, and correct the status line after that key.
- **Surprise**: The log path is only the status string. `[`, `]`,
  and `t` change the month and leave that string up, so the line
  names a file for a month the sheet is no longer showing. A
  month-cell save started after the key reloads the new rows and
  was keeping that path, so the line offered a file that did not
  contain the save. The hold is only for a save already armed
  when the path was written. `beginEdit` and `cancelEdit` clear
  `month.err`, so the next typed character hid a `pdf_dir`
  failure that was still true. Shift-P writes nothing while the
  save count is above zero, so the in-flight test has to release
  that count without applying the message before it can write.
  One boolean for the count let a stale save release a newer one.
  `app.go` was at 499 lines, so those rules live in `logsheet.go`.
- **Proposal**: In AGENTS.md GOTCHAS, record that the month-sheet
  log path belongs to the month it was written for, that a later
  cell save sets the status to `saved` and does not rewrite the
  file, that `write log pdf` survives the next cell edit, and
  that the in-flight hold is a count of armed saves.
- **Improvement**: When a decision says "already in flight" and a
  later requirement drops those words, a test that injects a bare
  `savedMsg` after the path locks the broader reading. Arm the
  in-flight save, and give a save started after the path its own
  test.
- **Signal**: context
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: spec, stories, consultation,
    tdd-agent, implement, code-reviewer, code-mode diaboli,
    manual pull request
  - Agent delegation: partial
