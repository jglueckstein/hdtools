---
target: "docs/superpowers/slices/filled-log-sheet-pdf.md (S2: Export that page from the month sheet with shift-p)"
target_kind: slice
generated_at: 2026-10-16
generated_by: "cost-estimator / tier:unlisted"
grounding_sources:
  - path: MODEL_ROUTING.md
    kind: model-routing
  - path: observability/costs/
    kind: cost-snapshot
  - path: observability/costs/per-pr/
    kind: calibration
tokens: { low: 24000, high: 24000 }
tokens_by_stage:
  - stage: spec-writer
    tokens: { low: 8000, high: 8000 }
    model_tier: Flagship
  - stage: tdd-agent
    tokens: { low: 4000, high: 4000 }
    model_tier: Balanced
  - stage: implementer
    tokens: { low: 6000, high: 6000 }
    model_tier: unlisted
  - stage: code-reviewer
    tokens: { low: 4000, high: 4000 }
    model_tier: Balanced
  - stage: integration
    tokens: { low: 2000, high: 2000 }
    model_tier: Efficient
agent_compute_time: { low: "2.4m", high: "7.2m" }
human_gate_time: "human-gate latency dominates total wall-clock and is not estimated numerically at S1; it depends on when a human next disposes a gate, not on the work."
confidence:
  tokens: low
  time: low
failure_direction: likely-underrun
---

## Included

Slice S2 of filled-log-sheet-pdf only: on an idle month sheet, shift-p writes the S1 file (`YYYY-MM-log.pdf`) for the month on screen. Idle `p`, chart `p`, and `P` during an edit do not write that file. S1 is assumed to exist already. Its writer is not in this estimate.

Token figures are the "Suggested max tokens" column of `MODEL_ROUTING.md` Token Budget Guidance. That column is one ceiling per task type, not a low–high range, so each stage range uses the ceiling as both bounds. Tiers are the Agent Routing strings in that file.

- spec-writer, tier Flagship, Spec writing, 8000
- tdd-agent, tier Balanced, Test generation, 4000
- implementer, no Agent Routing row (`model_tier` unlisted), Implementation (per file) 6000, counted once
- code-reviewer, tier Balanced, Code review, 4000
- integration, tier Efficient from the integration-agent row, CHANGELOG + commit, 2000

The whole-record token range is 24000–24000, the sum of those five ceilings. The passes are sequential, so that sum is the total used here. No $/token widening is applied: this routing table has no split tier, and there is no snapshot rate.

Agent-compute time applies the default band of about 1–3 minutes per 10k tokens to 24000 tokens, which is 2.4–7.2 minutes. The band is an assumption, not an observed throughput.

`generated_by` uses `tier:unlisted` because the Agent Routing table has no cost-estimator row. That token records the missing row. It is not a model id and not a tier copied from another agent.

Thirteen records in `observability/costs/per-pr/`, dated 2026-09-14 through 2026-10-15, are the calibration source. Feature-shaped records run spec-writer, tdd-agent, implementer, code-reviewer, and integration-agent (the integration stage above). That structural signal is why those five stages are included. Every per-stage token figure in those records is `unavailable`.

## Excluded

cost_usd: omitted — `observability/costs/` was read and holds no quarterly `*-costs.md` snapshot, so no observed $/token is available; cost is not estimated. Token and time figures stand. Records under `per-pr/` are not a rate source.

Also excluded: human-gate latency (gates cost wall-clock, not tokens); review re-runs past one pass; advocatus-diaboli, choice-cartographer, and convener (routing rows, no Token Budget Guidance row); orchestrator planning; S1's page writer, filename, and directory rule; any implementation file beyond the single per-file ceiling of 6000; on-screen columns, the chart painter, and the long-term chart, which this slice does not change.

## Confidence rationale

tokens: low. The dispatch stated `target_kind: slice`, so the ceiling is medium, and this axis sits under it. The slice names the key and the screens it leaves alone, and it does not enumerate scenarios or files. The budget table supplies a suggested maximum only, so each range collapses to that maximum. A collapsed range is carried at low confidence.

Thirteen per-PR actuals, 2026-09-14 to 2026-10-15, were read. None supplied a numeric per-stage token count, so the ranges were not narrowed and tokens confidence was not raised. cost unchanged — calibration refines tokens only.

time: low. `agent_compute_time` is the assumed 1–3 minute band applied to a collapsed suggested maximum, not to an observed token total. Human-gate time has no numeric estimate at S1, so this axis does not qualify it.

## Failure direction

likely-underrun. The guidance calls the maxima starting points, not hard limits, and says a verbose prompt hits the limit. This slice's spec and implementation prompts include the month sheet and the existing PDF path, so actual tokens sit above a point placed on the maximum. That is the dominant driver. Counting one implementation file, and leaving advocatus-diaboli, choice-cartographer, convener, further review cycles, and human-gate latency out, push the same way. A smaller opposing driver: a short single-file generation can land under its suggested maximum, which would put that one component above the actual output. Wall-clock follows the same underrun unless the human-gate caveat is read with `agent_compute_time`.
