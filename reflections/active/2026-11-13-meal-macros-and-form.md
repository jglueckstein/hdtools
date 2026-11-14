- **Date**: 2026-11-13
- **Agent**: parent session
- **Task**: Make protein, carbohydrate, and fat optional with
  `track_macros`, and correct the meal form after testing.
- **Surprise**: The name field was 24 columns, the month-sheet
  cell width. The name has the line to itself, so a long food
  scrolled while the rest of the line was empty. A 72-column
  line, including the label and the cursor, holds it. `0.3`
  was already a valid fat per 100 g. Enter on a blank fat
  field showed `fat: want a number`, and that message stayed
  through Tab and the typed `0.3`, so it looked like the
  number was refused. The Excel Hacker's Diet does not track
  macros. They are optional here. `track_macros` defaults
  off. Calories per 100 g stay mandatory. True makes the
  three fields mandatory. A calories edit must not wipe
  macros already stored. A weight-log reload still leaves a
  meal error up.
- **Proposal**: In AGENTS.md GOTCHAS, record that the meal
  name field is a 72-column line and not the month-cell
  width, that a refused Enter clears on Tab or the next
  typed character and survives a weight-log reload, and
  that `track_macros` defaults off so only calories are
  asked unless it is true.
- **Improvement**: When a tester says a valid number was
  rejected, check whether an earlier Enter left its message
  on the form. The message outlives the key that caused it
  until the next edit clears it.
- **Signal**: context
- **Constraint**: none
- **Promoted**: 2026-11-13 → AGENTS.md GOTCHAS: "The meal name field is a 72-column line."
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: carpaccio, spec-writer,
    spec-mode diaboli, tdd-agent, implement, manual testing
  - Agent delegation: partial
