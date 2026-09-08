# AI Literacy Assessment — hdtools

**Date**: 2026-09-08
**Assessed by**: assessor (via `/assess`)
**Assessed level**: Level 3 — Habitat Engineering

Canonical habitat documents (user-confirmed): `HARNESS.md` at repo root;
`AGENTS.md`; `CLAUDE.md`. `.claude/HARNESS.md` is a leftover unfilled
template from `/superpowers-init` and was **not** used for scoring.

---

## Observable Evidence

### Repository Signals

| Signal | Found | Level indicator |
| --- | --- | --- |
| CI workflows | yes — `.github/workflows/harness.yml`, `.github/workflows/ai-literacy.yml` | L2 |
| Test coverage enforcement | no | L2 |
| Vulnerability scanning | partial — gitleaks in CI/`HARNESS.md`; no govulncheck | L2 |
| Mutation testing | no | L2 |
| CLAUDE.md | yes — 104 lines | L3 |
| HARNESS.md | yes — 8 constraints (3 deterministic, 5 agent), 8 GC rules (3 fitness functions added this session) | L3 |
| AGENTS.md | yes — STYLE 3, GOTCHAS 3, ARCH_DECISIONS 3 | L3 |
| MODEL_ROUTING.md | yes | L3 |
| Custom skills | no (plugin skills used from marketplace, none vendored in-repo) | L3 |
| Custom agents | yes — 5 under `.claude/agents/` | L3 |
| Custom commands | no in-repo `.claude/commands/` | L3 |
| Hooks configured | no `.claude/settings.json` / hooks | L3 |
| REFLECTION_LOG.md | yes — 1 fragment (this assessment) | L3 |
| Parallel-tool config | yes — `.cursor/rules/`, `.github/copilot-instructions.md`, `.windsurf/rules/` | L3 |
| Specifications directory | yes — `docs/superpowers/specs/` created this session; no dated spec files yet (`idea.md` still the product backlog) | L4 |
| Implementation plans | no `plan.md` | L4 |
| Orchestrator with safety gates | file present (`MAX_REVIEW_CYCLES=3` in `.claude/agents/orchestrator.md`); not operated | L4 |
| Plugin/platform tooling | no | L5 |
| OTel configuration | no | L5 |

### Evidence Summary

The repo has a living harness, synced convention files for three AI
tools, CI that actually runs `gofmt` / `go test` / gitleaks, and a
copied five-agent team. That is a real L3 habitat. It is thin where L4
lives: no dated specs, no implementation plans, orchestrator unused,
reflection log empty of sessions. Verification is disciplined tests-in-CI
without coverage or mutation. Sophistication marker (cited, no level
bump): the orchestrator agent text is state-based (review-cycle
guardrail) but is a copied artefact, not an operated pipeline.

## Clarifying Responses

| Topic | Answer |
| --- | --- |
| Spec before code | **1** — spec / `idea.md` updated before code most of the time |
| Verification | **1** — systematic; red tests/format/CI are a fail even if the diff looks right |
| Cost | **2** — rough order of magnitude, not a number they would stand behind |
| Learning capture | **1** — written into the habitat (`/reflect`, `AGENTS.md`, or `REFLECTION_LOG.md`) |

Self-report of spec-first and habitat reflections is **ahead of the
artefacts**: `idea.md` is the spec stand-in (no `docs/superpowers/specs/`),
and `REFLECTION_LOG.md` has zero fragments. AGENTS.md *was* seeded with
real gotchas from this collaboration, which supports the learning claim
partially.

## Level Assessment

### Primary Level: 3 — Habitat Engineering

Meets the L3 bar: `CLAUDE.md` + eight enforced harness constraints +
custom agents. Not L4: no spec directory, no plan files, agent pipeline
not run with a plan-approval gate. Not L2-only: context and constraints
are written and promoted (tests went unverified → deterministic; error
wrapping audited and fixed).

### Discipline Maturity

| Discipline | Strength (1-5) | Evidence |
| --- | --- | --- |
| Context Engineering | 3 | `CLAUDE.md`, `HARNESS.md` Context (stack, conventions including literate programming), Cursor/Copilot/Windsurf sync |
| Architectural Constraints | 3 | 8/8 enforced; 3 deterministic in CI; promotion ladder used |
| Guardrail Design | 3 | CI gates for format/tests/secrets; agent PR constraints declared but not dispatched; no hooks |

### The Weakest Discipline

**Guardrail design** is the ceiling for L4: the orchestrator and
review-cycle text exist, but CI cannot fail agent rules, and there are
no session hooks. Context and constraints are already L3-shaped.

## Operational Axes (ALCI Part D)

Placement mode: **evidence-first** (40-statement survey offered, not taken).

| Axis | Placement | Evidence |
| --- | --- | --- |
| Composition | L3 | Five documented agents including a read-only code-reviewer; `MODEL_ROUTING.md`; orchestrator file present, not run as an ensemble |
| Testing | L2 | 10 `*_test.go` files; `go test ./...` in CI; no coverage threshold, no mutation |
| Observability | L2 | `HARNESS.md` Status + `/harness-audit`; empty reflection log; no cost snapshots or agent-activity metrics |
| Governance | L3 | Written `HARNESS.md`; falsifiable rules; unverified→deterministic promotion; deterministic CI. See Governance Dimension. |

Operational axes mean: L2.5

## Habitat Build Gap

```text
Level placement (from cognitive assessment): L3
Operational axes mean (Part D):              L2.5
  Composition:    L3
  Testing:        L2
  Observability:  L2
  Governance:     L3
Habitat Build Gap:                           +0.5
Interpretation:                              Ambition outpaces enablement
```

Cognitive L3 is slightly ahead of what the habitat *delivers* on testing
and observability. The useful investment is operating the artefacts
already present (reflections, dated specs, agent CI) rather than adding
more template files. Coherence is close; one more operational step
(coverage or a used reflection cadence) would land in the Coherent band.

## Governance Dimension

**Level: L3** (matches the Governance axis).

Written constitution (`HARNESS.md`, `CLAUDE.md`) with a promotion ladder
and deterministic CI for format, tests, and secrets. Agent constraints
are declared for review, not blocking in GitHub Actions. No
`/governance-constrain` institutional-frame modelling. Next ladder
rung: dispatch agent PR constraints or keep them honestly advisory.

## Strengths

- Living `HARNESS.md` with 8/8 enforced constraints and a clean 2026-09-08 audit
- Parallel-tool convention sync (Cursor, Copilot, Windsurf)
- Systematic verification habit (self-report) matching CI (`gofmt`, `go test`, gitleaks)
- Compound-learning file actually contains project-specific GOTCHAS and ARCH_DECISIONS, not empty headings
- Spec-first habit using `idea.md` before TUI/log/config work

## Gaps

- No dated spec *files* under `docs/superpowers/specs/` yet (directory and CLAUDE.md rule exist)
- Agent PR constraints cannot fail a build (no auto-enforcer)
- No coverage *threshold* or mutation testing (Testing axis still L2; weekly cover trend is GC, not a PR gate)
- No cost snapshot — `/cost-capture` deferred (Grok Build / xAI, Claude-shaped plugin)

## Recommendations

1. Write the next behaviour change as a dated spec in `docs/superpowers/specs/` — closes the L4 spec-directory gap while matching the existing spec-first habit (`idea.md`).
2. Run `/reflect` after sessions that change behaviour so the empty log becomes evidence of the claimed learning habit.
3. Add a `go test -cover` (or coverage gate) when the suite is more than a stub — Testing axis L2→L3.
4. Delete or clearly mark `.claude/HARNESS.md` as non-canonical so discovery never forks again.
5. Capture a cost snapshot (`/cost-capture`) once, even a back-of-envelope line in AGENTS.md.

## Immediate adjustments applied

- Added AI Literacy Level 3 badge to `README.md` pointing at this document.
- No Status rewrite: already 8/8, drift no, last audit 2026-09-08.
- Banner on `.claude/HARNESS.md` (workflow rec 3).
- Created `docs/superpowers/specs/` (`.gitkeep`) for dated specs.

## Workflow operation changes

| Change | Decision |
| --- | --- |
| After behaviour-changing sessions, run `/reflect` (CLAUDE.md Session close) | **accepted** |
| Next behaviour change gets a dated spec under `docs/superpowers/specs/` | **accepted** |
| Banner on `.claude/HARNESS.md`: unfilled template; root `HARNESS.md` is the only living harness | **accepted** |
| Capture a cost snapshot (`/cost-capture`) | **deferred** — Grok Build is the agentic CLI; plugin capture is Anthropic/OpenAI-shaped and `/cost-estimate` dollar binding is Claude-family only. Revisit when a Grok/xAI path exists, or run a manual snapshot from `/usage` + console.x.ai |

## Improvement Plan

- Current level: L3
- Target level: L4
- Improvements accepted: 1
- Improvements skipped: 0
- Improvements deferred: 0
- Commands executed: fitness-functions skill, /harness-gc (add)

Already present (filtered out of the L3→L4 list): agent pipeline (`.claude/agents/orchestrator.md`), `MAX_REVIEW_CYCLES=3` safety gates, convention sync (Cursor / Copilot / Windsurf), spec-first operating rule + `docs/superpowers/specs/` (first dated spec still pending next behaviour change).

### Accepted

| Gap | Action | Result |
| --- | --- | --- |
| No architectural fitness functions | fitness-functions + `/harness-gc` | Three weekly GC rules: layer import direction (`scripts/check-import-direction.sh`), file size >500 lines, per-package `go test -cover` trend. GC 5/5 → 8/8. Skipped `go-cleanarch` (layout is not those layer names). |

### Skipped

_(none)_

### Deferred

_(none)_

## Next Assessment

Suggested re-assessment date: 2026-12-08 (quarterly)

Previous assessment: first assessment
