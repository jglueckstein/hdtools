- **Date**: 2026-10-13
- **Agent**: parent session
- **Task**: Run harness garbage collection and file findings #86, #87,
  and #88.
- **Surprise**: A same-package name count marks ResolvePDFDir and
  App.Init unused because their callers sit in other files. It also
  marks plotMonthly, which braille.go keeps so the bin rule stays
  for the long-term grid. The documentation finding is the Charm v1
  stack line already named on 2026-10-08, still present, now
  issue #86.
- **Proposal**: In AGENTS.md GOTCHAS, record that every harness GC
  rule is auto-fix false, so findings become issues. A same-package
  unused name is not dead code until a module-wide search agrees.
  plotMonthly in braille.go is unreferenced on purpose; the
  long-term chart still calls newBrailleGrid with monthly false.
- **Improvement**: Write the per-package coverage figures into the
  GC snapshot. This run left them in /tmp, so the next run still
  cannot tell a decline from a first measurement.
- **Signal**: workflow
- **Constraint**: none
- **Promoted**: 2026-10-14 → AGENTS.md GOTCHAS: "Every harness GC rule is auto-fix false, so findings become issues."
- **Session metadata**:
  - Duration: unknown
  - Model tiers used: unknown
  - Pipeline stages completed: harness-gc
  - Agent delegation: partial
