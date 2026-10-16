- **Date**: 2026-10-15
- **Agent**: integration-agent
- **Task**: Rate on-screen loss from the last seven trends or a line
  fit.
- **Surprise**: FitRate's unary minus turns an exact zero slope into a
  negative zero, and %.1f prints -0.0. The last-seven tests compared
  0.0 and never opened a finished flat month.
- **Proposal**: none
- **Improvement**: A flat finished-month caption belongs in the first
  red tests for a rate change. The pencil path and the fit do not
  share a sign for zero.
- **Signal**: failure
- **Constraint**: A zero on-screen loss prints 0.0, never -0.0.
- **Promoted**: 2026-10-16 → AGENTS.md GOTCHAS: "A zero on-screen loss prints `0.0`."
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: spec-writer, tdd-agent, implementer,
    code-reviewer, integration-agent
  - Agent delegation: partial
