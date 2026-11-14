- **Date**: 2026-11-14
- **Agent**: integration-agent
- **Task**: Show food macro fields when track_macros is off and
  store a blank macro as zero.
- **Surprise**: `assertFieldShows` matches only the focused row
  (`> ` plus the label). A macro field that is on screen still
  fails that assertion until Tab lands on it.
- **Proposal**: none
- **Improvement**: none
- **Signal**: none
- **Constraint**: none
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: spec, tests, and implementation in the parent session; integration-agent
  - Agent delegation: manual
