- **Date**: 2026-10-07
- **Agent**: integration-agent
- **Task**: Write monthly chart PDFs to the XDG data directory, with
  `pdf_dir` and `-o`.
- **Surprise**: `gh pr checks --watch` exited at once with "no checks
  reported" while the workflow runs were already in progress. A later
  push started a new pair, so the first green pair was not the head.
  The session date was behind the machine clock, and the actuals file
  had to be renamed to the integration day.
- **Proposal**: In AGENTS.md GOTCHAS, note that `gh pr checks --watch`
  is not a result when the rollup is empty. Wait for the runs on the
  current head SHA.
- **Improvement**: Take the actuals date from `date -u`, and watch the
  runs for that SHA instead of the first check command that returns.
- **Signal**: workflow
- **Constraint**: none
