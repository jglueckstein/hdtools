- **Date**: 2026-10-08
- **Agent**: parent session
- **Task**: Close the Charm v2 upgrade after the merged pull request.
- **Surprise**: The integration-agent file still checks out `main`,
  writes a dated changelog, and commits the reflection after the
  merge. This repo uses `master`, an Unreleased changelog, and a
  reflection on the feature branch before the squash. The dispatch
  prompt had to override the agent file. Code review passed before
  the paste, `NO_COLOR`, and one-rune fixes, and there was no second
  review.
- **Proposal**: In AGENTS.md GOTCHAS, say that integration stays on
  the feature branch, targets `master`, and keeps CHANGELOG under
  Unreleased. The agent file's `main` checkout loses.
- **Improvement**: Put that override in the integration prompt every
  time, and say in the merge note when a PASS predates a code-mode
  fix. `HARNESS.md` still names the v1 Charm modules; leave it for
  a later harness sync.
- **Signal**: workflow
- **Constraint**: none
- **Promoted**: 2026-10-08 → AGENTS.md GOTCHAS: "Integration stays on the feature branch until squash-merge."
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: carpaccio, spec-writer,
    advocatus-diaboli, tdd-agent, implementer, code-reviewer,
    integration-agent
  - Agent delegation: partial
