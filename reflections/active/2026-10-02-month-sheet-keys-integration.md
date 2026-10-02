- **Date**: 2026-10-02
- **Agent**: integration-agent
- **Task**: Ship month-sheet Enter, Down, and Up as a save that moves
  one day.
- **Surprise**: Code review passed the live-cursor save. Code-mode
  diaboli found that the command reads the cell when it runs. The
  spec, plan, and objections stayed uncommitted until the snapshot
  commit. The branch was one commit behind master.
- **Proposal**: none
- **Improvement**: Commit the spec with the failing tests so review
  reads the contract from the branch, not the working tree.
- **Signal**: workflow
- **Constraint**: none
