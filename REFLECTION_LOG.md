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
