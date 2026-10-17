# Record grok usage on a pull request

**Date**: 2026-10-17
**Status**: approved
**Issue**: [#111](https://github.com/jglueckstein/hdtools/issues/111)
**Backlog**: contributor workflow, not a product slice in
[`idea.md`](../../../idea.md)

One Grok session in this repo lasts across many pull requests.
`grok usage <session-id>` can name each turn's tokens and priced
ticks. The session total cannot. The weekly SuperGrok percentage
cannot. This command reads the turns that belong to one branch and
fills the per-PR actuals figures. It does not invent a per-stage
split, and it does not write a quarterly cost snapshot.

## User stories

### US1 — Mark the first turn

As a person starting a branch, I want the session id and the next
turn stored, so a later capture does not take the whole session.

### US2 — Sum only that range

As a person opening the pull request, I want `tokens_total` and
`cost_usd` from those turns, so the actuals file has observed
figures.

### US3 — Ask when the range is unknown

As a person who did not mark the start, I want clear instructions
and no written figures, so a weekly percentage is not typed in as
if it were tokens.

## Acceptance

- `mark` writes `observability/costs/spans/<branch>.json` with the
  session id, the branch, and `from_turn`. Without `--from-turn`,
  `from_turn` is one past the last turn in the usage JSON. An empty
  turn list marks turn 1.
- `capture` sums `totalTokens` for turns from `from_turn` through
  `through_turn`, inclusive. The first capture sets `through_turn`
  to the last turn in the file that is in range, and writes it back,
  so a later session turn is not added on a re-run.
- The session object's `totalTokens` is ignored.
- Every `tokens_by_stage` value stays `unavailable`.
- `cost_usd` is the sum of `costUsdTicks` divided by 1e10 when every
  included turn has that field. If any included turn omits it,
  `cost_usd` is `unavailable` and `tokens_total` is still written.
- A turn in range that omits `totalTokens` writes nothing. The
  command tells the user which turn.
- No span file writes nothing. The message tells the user to mark,
  how to choose `--from-turn`, and that the weekly percentage is
  not a token count.
- An actuals file whose `figures_source` is already
  `human-supplied` is left unchanged unless `--replace` is set.
- `figures_source` becomes `human-supplied` only when the figures
  were read from usage JSON.
- `scripts/usage-capture` is the same command. It exits 2 when a
  person still has to choose a number, and 1 when the tool fails.
  `go run` reports both as exit status 1.

## Out of scope

- A quarterly `observability/costs/*-costs.md` snapshot.
- Mapping turns onto spec-writer, tdd-agent, implementer,
  code-reviewer, or integration-agent.
- Teaching the cost estimator to price `grok-4.6-build` or
  `grok-4.7-build`.
- Backfilling older per-PR files.
