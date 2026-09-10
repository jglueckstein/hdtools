# Project Conventions

High-signal rules for agents. Stack, full constraints, and GC live in
`HARNESS.md` — do not duplicate that file here. User-facing docs follow
Diátaxis (`docs/tutorials`, `how-to`, `reference`, `explanation`);
README is the map. Markdown source follows Google's Markdown Style
Guide (80-col wrap, ATX headings, one H1, fenced code with a
language).

## Literate Programming

All code follows Don Knuth's literate programming principles as declared
in `HARNESS.md` (Documentation convention and the "Literate preamble on
new files" constraint).

When creating a new source file or significantly rewriting one:

1. Open with a narrative preamble — why it exists, key design decisions,
   what it deliberately does NOT do
2. Function comments explain reasoning, not signatures
3. Order of presentation follows understanding — orchestration before detail
4. Each file has one clearly stated concern
5. Inline comments explain WHY, not WHAT — restating the next line is prohibited

## CUPID Code Review

When reviewing or refactoring, apply CUPID: Composable, Unix philosophy,
Predictable, Idiomatic, Domain-based.

## Workflow

### Spec-First Change Discipline

Behaviour changes update `idea.md` **and** a dated spec at
`docs/superpowers/specs/YYYY-MM-DD-<slug>.md` before implementation.
`idea.md` is the product backlog; the dated file is the change record.

1. Update `idea.md` if the product intent changed
2. Add or revise the dated spec (user stories, acceptance, FRs)
3. Write failing tests from the spec — confirm red
4. Implement until green
5. Refactor while tests stay green

### Test-Driven Development

Red-green-refactor. No production code without a failing test first,
except wiring that has no behaviour of its own (`cmd/hdtools` flag
resolution).

### Session close

After any session that changes behaviour, run `/reflect` before
stopping so a fragment is written under `reflections/active/`. An
empty `REFLECTION_LOG.md` is not compound learning.

### Branch Discipline

Do not commit directly to `master`. Open an issue, use a hyphenated
branch name, and merge via PR. Origin is `git@github.com:jglueckstein/hdtools.git`.

### Commit Messages

Concise: what changed and why. No postamble, no attribution lines.

### CHANGELOG and PR checks

When a GitHub remote and PRs exist: update CHANGELOG.md before the PR,
and watch `gh pr checks` until green. Do not invent a CHANGELOG for
local-only commits.

## Build and Test

    # Build
    go build ./...

    # Test
    go test ./...

    # Format check (CI treats any listed path as failure; `gofmt -l` itself exits 0)
    test -z "$(gofmt -l .)"

    # Format
    gofmt -w .

## Project Constraints

- Weight is stored in kilograms. Display units (kg/lb/st) are config-only.
- Trend is derived (`ApplyTrend`), never a SQLite column.
- Config is XDG (`$XDG_CONFIG_HOME/hdtools/config.toml`), not the database.
- Application code lives under `internal/`. `cmd/hdtools` is wiring only.
- Errors are wrapped with `%w` and context; no bare `return err`.
- See `HARNESS.md` for the full constraint list and enforcement.

## Learnings

REFLECTION_LOG.md is a generated aggregate of `reflections/active/`
fragments. Read recent entries before starting work. Do not edit
REFLECTION_LOG.md by hand. Humans curate AGENTS.md.

## Reflection Log Curation

Write reflections as fragments under
`reflections/active/<YYYY-MM-DD>-<slug>.md` via `/reflect`. Regenerate
the aggregate with `scripts/regenerate-reflection-log.sh` when that
script is present.

When promoting into AGENTS.md or HARNESS.md, add a `Promoted` line on
the fragment in the same commit.

## Monthly Operations

1. `/harness-sync` if convention files may have drifted
2. Scan REFLECTION_LOG.md for entries worth promoting to AGENTS.md
3. `/harness-audit` when constraints change

## Conventions (extracted 2026-09-08)

- Extract a helper when the same code is already wrong in two places.
  `internal/tui/layout.go` (`visPad`, shared column widths for list and
  month) is the model of an extraction that was due. A layout framework
  (or the same helper for a single screen) is over-engineered.
