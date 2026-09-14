# Reflection Log

<!-- GENERATED FILE — do not edit by hand.

     This file is a deterministic aggregate of the per-entry fragments in
     reflections/active/. Add a reflection with /reflect (which writes a
     fragment and regenerates this file); never append here directly.
     Regenerate with: scripts/regenerate-reflection-log.sh

     Each entry below mirrors one fragment. Entry format:

     ---

     - **Date**: YYYY-MM-DD
     - **Agent**: integration-agent
     - **Task**: [one-sentence summary]
     - **Surprise**: [anything unexpected]
     - **Proposal**: [pattern or gotcha for AGENTS.md, or "none"]
     - **Improvement**: [what would make the pipeline smoother]
     - **Signal**: [context | instruction | workflow | failure | none]
     - **Constraint**: [proposed constraint, or "none"]
-->

---

- **Date**: 2026-09-08
- **Agent**: assessor (via /assess)
- **Task**: AI literacy assessment (Level 3 — Habitat Engineering)
- **Surprise**: `/assess` habitat discovery found two `HARNESS.md` files (living root vs unfilled `.claude/` template) and had to stop for a canonical pick. `/cost-capture` is a dashboard interview shaped for Anthropic/OpenAI; Grok Build's `/usage` and console.x.ai are the real numbers, and `/cost-estimate` dollar binding is Claude-family only.
- **Proposal**: Keep a banner on leftover plugin templates so discovery never forks. Treat Grok `/usage` + console.x.ai as the cost source until the plugin lists xAI.
- **Improvement**: Plugin `/cost-capture` should accept xAI/Grok as a first-class provider; the estimator binding table needs Grok family stems or an explicit "omit dollars on this stack" path.
- **Signal**: workflow
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown (Grok Build, grok-4.6)
  - Pipeline stages completed: /assess (scan, questions, document, workflow recs, literacy-improvements)
  - Agent delegation: manual

---

- **Date**: 2026-09-09
- **Agent**: Grok 4.6
- **Task**: Completed the full harness promotion cycle: all 14 constraints promoted to deterministic enforcement, 0600 file modes fixed, error wrapping applied, convention files synced, first snapshot captured, ONBOARDING.md generated and linked, all PRs merged.
- **Surprise**: The number of small mechanical fixes (bare returns, 0600 chmod, TUI import) that surfaced only after deterministic checks were added.
- **Proposal**: none
- **Improvement**: Future agents should know that every new deterministic constraint must be accompanied by a passing check script and a test that exercises the exact failure mode the constraint targets.
- **Signal**: failure
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: unknown
  - Agent delegation: manual

---

- **Date**: 2026-09-10
- **Agent**: Grok 4.6
- **Task**: Habitat loop with origin live: GC (6 findings), governance constraint User-facing copy is not medical, first governance audit and dashboard, observatory-verify, health snapshot with Trends, convention-sync, ONBOARDING regen, harness-sync no-op.
- **Surprise**: Plugin agent types (`ai-literacy-superpowers:harness-gc`, `governance-auditor`, and the rest) are listed as available but spawn as unknown; we had to use `general-purpose` with the agent brief. Same session: failure action is block merge while `harness.yml` does not run harness-enforcer; Observatory footer says 81 signals / 39 snapshot rows, the tables are 40 + 14.
- **Proposal**: AGENTS.md GOTCHA: plugin-prefixed spawn_subagent types may fail with "Unknown subagent type" even when advertised. Fall back to general-purpose and paste the agent brief.
- **Improvement**: Close the enforcer/CI split (dispatch harness-enforcer or retarget failure action). For dated snapshots, pick one clock (UTC vs local) so /harness-health does not fork "today".
- **Signal**: workflow
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown (Grok Build, grok-4.6)
  - Pipeline stages completed: /harness-gc, /governance-constrain, /governance-audit, /governance-health, /harness-audit, /convention-sync, /harness-onboarding, /observatory-verify, /harness-health, /harness-sync
  - Agent delegation: partial

---

- **Date**: 2026-09-14
- **Agent**: integration-agent
- **Task**: Add PDF export of the on-screen monthly chart (CLI -chart-pdf and TUI p).
- **Surprise**: Spec and plan drafted in an earlier session were uncommitted and had to be recovered from rewind points after a branch switch. Code-mode diaboli caught `fpdf.OutputFileAndClose` (`os.Create` 0666 then chmod) as a real 0600 hole; temp+rename was the store's pattern. `ai-literacy-superpowers:advocatus-diaboli` spawn failed as unknown even when listed; `tdd-agent` and `code-reviewer` spawned by short name.
- **Proposal**: AGENTS.md GOTCHA: a print artefact that must be 0600 cannot use `os.Create` then chmod; write a same-dir temp at 0600 and `Rename`.
- **Improvement**: Keep spec/plan on the feature branch before switching away. Dispatch diaboli as `general-purpose` with the skill when the plugin-prefixed type is unknown.
- **Signal**: workflow
- **Constraint**: none
