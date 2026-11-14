- **Date**: 2026-11-14
- **Agent**: parent session
- **Task**: Find why Grok rejects the harness-gc subagent id.
- **Surprise**: On Grok 1.0.77, `grok inspect` registers
  `agents/harness-gc.agent.md` as
  `ai-literacy-superpowers:harness-gc`. `spawn_subagent` lists
  that id and rejects it. Passing
  `ai-literacy-superpowers:harness-gc.agent` comes back in the
  error with `.agent` already removed. The same call launches
  `academic-writing-agents:brainstormer`, whose file is
  `brainstormer.md`. Omitting `subagent_type` runs the pasted
  brief on the general-purpose agent, and that run finished
  the eight GC rules.
- **Proposal**: Replace the Grok 1.0.46 spawn gotcha in
  AGENTS.md. The listed `plugin:<name>` id is still rejected.
  The `.agent` suffix is stripped before lookup, so it is not
  a usable id. Plugin agents whose files are `<name>.md`
  launch as `plugin:<name>`. Project agents in
  `.claude/agents/<name>.md` launch under the short name.
  When the specialist id is rejected, omit `subagent_type`
  and paste the brief. Do not rename the installed
  `*.agent.md` files; a plugin update overwrites them.
- **Improvement**: Stop retrying the id the error lists. One
  rejected call is enough to switch to the general-purpose
  brief.
- **Signal**: workflow
- **Constraint**: none
- **Promoted**: 2026-11-14 → AGENTS.md GOTCHAS: "omit `subagent_type` and paste"
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: harness-gc via general-purpose,
    spawn probes, parent session
  - Agent delegation: partial
