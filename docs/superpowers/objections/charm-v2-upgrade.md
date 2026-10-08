---
spec: docs/superpowers/specs/2026-10-07-charm-v2-upgrade.md
date: 2026-10-07
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "Sixteen-color downshift is required in the shown text and locked by the view-string suite, but Charm v2 no longer applies the profile inside that string."
    evidence: "Decision 10: \"A hex colour on a truecolor profile is emitted as 38;2. The same hex on a 16-color profile is an indexed colour and is not 38;2.\" S13: \"the text does not contain 38;2 for that hex.\" The introduction says the existing suite already locks S13."
    disposition: accepted
    disposition_rationale: >
      Accepted. The 16-color floor is the colour written to the
      terminal, not the string inside the screen model. Hex on a
      truecolor profile is still 38;2 in that output. Hex on a
      16-color profile is an indexed colour and is not 38;2. The
      model text may stay full fidelity. The writer downsamples
      once, and the app does not downsample the view string itself.
      NO_COLOR, bold, reverse, and > stay properties of the model
      text. Profile-dependent colour tests read the written output.
      lipgloss.Writer.Profile is not treated as something that
      changes Style.Render.
  - id: O2
    category: implementation
    severity: medium
    claim: "Starting with the model alone does not keep today's terminal, and the spec waives the scenario that would notice."
    evidence: "Decision 12: \"The program is started with the screen model alone, as it is today. It does not switch the terminal to an alternate screen and it does not turn on mouse reporting.\" The introduction then says S16 \"is starting the program with the screen model alone\" and \"needs a new test\" is false for it."
    disposition: accepted
    disposition_rationale: >
      Accepted. Decision 12 stays no alternate screen and no mouse.
      Charm v2 enables modifyOtherKeys and Kitty disambiguation for
      any program that reads keys, and this slice does not try to
      turn that off. S16 is not a promise that the wire protocol
      matches v1, and it still needs no new test. Enter, Tab, Esc,
      the arrows, q, and ctrl+c must still match. If one of those
      arrives under a different printed name, stop and report. Do
      not invent new key names.
  - id: O3
    category: specification quality
    severity: medium
    claim: "S1 and FR1 cannot fail a wrong Charm v2 module, because they never name the paths or a release."
    evidence: "S1 and FR1: \"it depends on Charm v2 of the three libraries the TUI already uses (Bubble Tea, Lip Gloss, and bubbles).\" The spec points at the idea.md change record for the backlog, and that paragraph names module paths for only two of the three."
    disposition: accepted
    disposition_rationale: >
      Accepted. S1 and FR1 require charm.land/bubbletea/v2,
      charm.land/lipgloss/v2, and charm.land/bubbles/v2, and still
      forbid ntcharts. No patch release is pinned. The plan still
      fetches the current v2 releases. The idea.md change record
      names the bubbles module next to the other two.
---

## O1 — implementation — high

### Claim

Decision 10 and S13 require the shown text to downshift a hex colour off
`38;2` on a 16-color profile, and the introduction makes the existing suite
the proof of that. Charm v2 no longer downshifts inside the string that
suite reads. A correct upgrade therefore cannot satisfy both the scenario
and the libraries.

### Evidence

Decision 10:

> 16-color is still the floor and the downshift target. A hex colour on a
> truecolor profile is emitted as `38;2`. The same hex on a 16-color
> profile is an indexed colour and is not `38;2`.

S13 repeats the oracle: given the same hex and a 16-color profile, "the
weight downshifts to an indexed colour" and "the text does not contain
`38;2` for that hex." The introduction says S10 through S18, which includes
S13, are behaviour the suite already locks, and that they are not a second
product. S10 binds chart colour the same way: "the glyphs, title, and
colours the current chart tests expect."

### Why this matters

Those tests read the view string. Lip Gloss v1 applied the profile inside
`Style.Render` (`s.r.ColorProfile()` in v1.1.0 `style.go`). The Lip Gloss v2
upgrade guide says `Render` always emits full-fidelity ANSI and that
downsampling happens on print. `Style.Render` in lipgloss v2.0.6 writes the
colour with no profile consult. `lipgloss.Writer.Profile` affects `Sprint`
and `Println`, not that string. Bubble Tea v2 then downsamples in its own
writer; the same guide says a Bubble Tea app needs no further change.

An import bump therefore leaves `38;2` in the view on a 16-color profile.
S13 stays red, and so does the locked list test. The companion plan says to
set `lipgloss.Writer.Profile` and not to change the expected SGR. That
field does not affect `Render`, so the instruction cannot turn the tests
green. Downsampling in the app, so the view string matches S13, hands an
already indexed sequence to Bubble Tea's writer, which converts it again.
The spec never chooses between "the view stays full fidelity and the writer
is the floor" and "the view string is still the floor." Until it does,
there is no green state that satisfies both S13 and these libraries. S10
has the same split, because the chart colours the current tests expect are
also SGR in the view string. FR11 then requires the changelog to call the
result behaviour-preserving.

## O2 — implementation — medium

### Claim

Decision 12 treats `tea.NewProgram` with the model alone as the same
terminal as today: no alternate screen, no mouse, and the same visible
text. On Charm v2 that call still turns on keyboard modes v1 never set,
and the spec waives S16 by reducing it to the constructor arguments.

### Evidence

Decision 12: "The program is started with the screen model alone, as it is
today. It does not switch the terminal to an alternate screen and it does
not turn on mouse reporting." S16's Then clauses are those two modes plus
"the visible text is the same full text the screens already show." The
introduction equates S16 with "starting the program with the screen model
alone" and says none of that scenario needs a new test. US4 says the
upgrade does not change how the log is read or left.

### Why this matters

Alt screen and mouse do default off if the view never sets them
(`AltScreen` false, `MouseMode` zero). That is the part Decision 12 names.
It is not the whole terminal setup. Bubble Tea v2's renderer, on a
model-only program, still writes modifyOtherKeys level 2 and pushes Kitty
key disambiguation on every session (`KittyDisambiguateEscapeCodes`,
commented "always enable" in `cursed_renderer.go` of both v2.0.2 and
v2.0.9). Bubble Tea v1.3.10 has no Kitty or modifyOtherKeys path.
`cmd/hdtools/main.go` today is the model-only call the spec is preserving.

Decision 11 is the only rule that refuses a printed key name, and it
applies to the space bar alone. Enter, Tab, Esc, `q`, and ctrl+c still
match the v1 names. A physical key the new protocol reports under another
name misses those branches. Tests that build a `KeyPressMsg` directly stay
green. Clean exit pops the protocol, so this is the session, not a leftover
shell. S16 cannot see it: the introduction replaces the terminal clauses
with "the constructor received only the model." The slice can meet that
substitute and still change which keys the running program delivers.

## O3 — specification quality — medium

### Claim

S1 and FR1 are the only new module gate, and a literal reading of them
accepts any dependency the reader is willing to call Charm v2. They name
neither the module paths nor a release, including for bubbles, which this
slice adds.

### Evidence

S1: "it depends on Charm v2 of the three libraries the TUI already uses
(Bubble Tea, Lip Gloss, and bubbles)" and "it does not depend on
ntcharts." FR1 repeats that sentence. Documentation says the backlog
change record is the Charm v2 paragraph in `idea.md`. That paragraph names
`charm.land/bubbletea/v2` and `charm.land/lipgloss/v2` only. It does not
name a bubbles module.

### Why this matters

"Charm v2" does not encode a require line. Two released Bubble Tea v2
modules already exist side by side (v2.0.2 wants go 1.24.2; v2.0.9 wants
go 1.25.0). Both satisfy the sentence. So does a differently pathed major
version 2 of a similarly named module. Bubbles is the library Decision 1
refuses to leave behind, and it is the one with no path in the spec or in
the change record the spec points at. S1 is judged by "when its library
requirements and its source imports are read," but the Then clause has no
strings to read for. One implementation can land a v2 set later chart work
does not import, and FR1 still passes. The ntcharts ban is the only module
identity the scenario can actually fail.

## Explicitly not objecting to

- **Doing the upgrade at all**: S1 is the accepted slice, and moving the
  three libraries before any Online picture is the right problem. O1 and
  O2 are places that claim fails, not a reason to stay on v1.
- **A separate alternatives objection for print-time downsampling**: that
  choice is the unresolved branch inside O1, not a second failure.
- **Clipping a tall frame to the terminal height**: Bubble Tea v1's
  standard renderer already drops lines above the known height, so this is
  not a v2 regression.
- **ctrl+c sharing `q`'s case on the idle list and on both charts**: that
  is what the code does today. A second test of the same case is not the
  hole. The hole is O2's protocol, which those cases do not observe.
- **The space-bar rule itself**: treating the key as a space character,
  not the word "space", matches the month sheet and the day form as they
  work now. Bubbles v2's text field was not in the module cache, so this
  record does not assert that the widget inserts the printed name.
- **Workout space versus day-form space**: the month sheet saves the
  flipped flag in place; the day form only toggles until save. That split
  matches `updateMonth` and `formModel.update`.
- **Bubbles moving in this slice**: the month sheet and the day form both
  import its text field. Leaving it on v1 would not be a separate product
  decision.
- **Deferred Online charts**: #77, #79, and #78 stay out. Kilograms,
  ApplyTrend, the pencil caption, the PDF path, and the four long-term
  windows are restated invariants, not a new product hiding in the
  upgrade.
- **Draft status, grammar, and the 80-column wrap.**
