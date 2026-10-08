- **Date**: 2026-10-08
- **Agent**: integration-agent
- **Task**: Upgrade the TUI to Charm v2 and leave every current screen
  behaving as it does today.
- **Surprise**: Bracketed paste arrived as PasteMsg, the idle-cell
  gate counted bytes so é did not open, and textinput's default greys
  ignored NO_COLOR. Code review had already passed. Those three
  accepted objections were fixed before this merge. `gh pr checks
  --watch` is not a reliable wait for this repo's two workflows.
- **Proposal**: none
- **Improvement**: After a code-mode objection changes code, say in
  the integration note that the PASS predates that fix.
- **Signal**: workflow
- **Constraint**: none
