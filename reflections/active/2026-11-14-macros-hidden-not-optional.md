- **Date**: 2026-11-14
- **Agent**: parent session
- **Task**: Record why food macros were hidden when
  track_macros was off.
- **Surprise**: Decision 24 was headed "Protein,
  carbohydrate, and fat are optional." The next sentences
  said false hides those fields. The 2026-11-13
  reflection still said the macros are optional and that
  true makes them mandatory. Its proposal said
  `track_macros` defaults off so only calories are asked
  unless it is true, and that sentence was written into
  AGENTS.md. Later sessions followed the gotcha and left
  the fields off the form. A blank answer was never
  offered. That hid the fields against the wish that they
  stay up and a blank be valid.
- **Proposal**: In AGENTS.md GOTCHAS, record that optional
  on a form field means the field is shown and a blank is
  valid. Hiding the field is a separate decision and is
  written as hidden.
- **Improvement**: When a spec heading says optional, the
  next sentence has to keep the field on the form and
  accept a blank. A reflection proposal has to keep that
  sentence. Shortening it to "the field is not asked"
  drops the wish.
- **Signal**: failure
- **Constraint**: An optional form field stays on the form
  and a blank is valid. Hide a field only by writing that
  it is hidden. (agent)
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: meal-day spec through
    implement wrote the hide rule; this session corrected
    it in #131
  - Agent delegation: partial
