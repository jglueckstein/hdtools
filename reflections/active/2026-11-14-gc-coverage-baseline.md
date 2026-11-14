- **Date**: 2026-11-14
- **Agent**: integration-agent
- **Task**: Save the 2026-11-13 per-package coverage figures so
  the weekly GC rule can flag a drop.
- **Surprise**: `spawn_subagent` rejects
  `ai-literacy-superpowers:harness-gc` even though the error
  lists that id. Passing the `.agent` suffix is stripped back
  to the same id and rejected again. Omitting `subagent_type`
  runs the brief on the general-purpose agent, and that run
  completed the eight rules.
- **Proposal**: none
- **Improvement**: When the listed harness-gc id is rejected,
  omit `subagent_type` and paste the brief. Retrying the same
  id does not start the agent.
- **Signal**: workflow
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: harness-gc, integration-agent
  - Agent delegation: partial
