---
spec: docs/superpowers/specs/2026-10-07-charm-v2-upgrade.md
date: 2026-10-08
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: >
      Bracketed paste never becomes text in a month cell or a day-form
      field, because Update handles only KeyPressMsg while every view
      leaves bracketed paste on.
    evidence: >
      internal/tui/app.go switches on tea.KeyPressMsg and otherwise
      returns. View returns tea.NewView with no
      DisableBracketedPasteMode. Bubble Tea v2's renderer then writes
      bracketed-paste mode, and a paste arrives as PasteMsg, which
      textinput would insert and this Update never forwards.
    disposition: accepted
    disposition_rationale: >
      Accepted. Bracketed paste arrives as tea.PasteMsg, and
      App.Update drops it, so an open day-form field and an open
      month cell never receive the text. On the previous stack that
      paste was a key message and the field inserted the runes.
      Forward PasteMsg to the focused day-form field, and to the
      month cell while it is editing. A paste of one rune on an idle
      non-workout cell starts that edit. A longer paste on an idle
      cell does nothing, which is what len(msg.Runes) == 1 already
      did. A test drives tea.PasteMsg into an open field and into an
      idle note cell.
  - id: O2
    category: risk
    severity: high
    claim: >
      NO_COLOR strips chroma only in the palette. The day form still
      paints the text field's 256-color placeholder and blurred-text
      styles into the model text, and the NO_COLOR test never opens
      the form.
    evidence: >
      internal/tui/style.go skips Foreground only when NO_COLOR is
      non-empty. internal/tui/form.go calls textinput.New and never
      SetStyles. bubbles v2 DefaultDarkStyles sets Placeholder to
      lipgloss.Color("240") and blurred text to Color("245") or
      Color("7"). TestNoColorStripsChromaKeepsStructure reads the
      list view only.
    disposition: accepted
    disposition_rationale: >
      Accepted. Decision 9 and S13 say a non-empty NO_COLOR leaves
      the model text without chromatic colour. textinput.New still
      installs the widget's default styles, so an empty field's
      placeholder is colour 240 (38;5;) in View().Content, and a
      blurred value uses the blurred text colour. The list test
      never opens the form. The previous widget already used
      placeholder colour 240; this slice's rule covers that model
      text. When NO_COLOR is set, the day-form fields and the
      month-cell field use styles with no foreground. A test opens
      the day form under NO_COLOR and requires no chromatic SGR,
      with the labels and the > mark still present. With NO_COLOR
      empty, the widget's default colours stay.
  - id: O3
    category: risk
    severity: medium
    claim: >
      An idle non-workout month cell starts an edit only when the key
      text is one byte, so a non-ASCII character never opens the cell.
    evidence: >
      internal/tui/app.go updateMonth default branch:
      len(msg.Text) == 1 then beginEdit(msg.Text). The new space-bar
      test types only Text "8".
    disposition: accepted
    disposition_rationale: >
      Accepted. The previous gate was len(msg.Runes) == 1. The port
      checks len(msg.Text) == 1, a byte count, so é does not open an
      idle note cell. S6's one character is that rune, and the edit
      shows it. The gate becomes one rune. A test types é on an idle
      note cell and expects the edit to show é. A two-rune sequence
      stays outside the gate, as the rune-length check did.
---

## O1 — implementation — high

### Claim

Bracketed paste never becomes text in a month cell or a day-form field.
`Update` handles only `KeyPressMsg`, and every view leaves bracketed
paste enabled, which is how this stack delivers a paste.

### Evidence

`App.Update` has no paste case. Anything that is not a load, a save, a
resize, or a key press is dropped:

```go
case tea.KeyPressMsg:
    switch a.screen {
    case screenForm:
        return a.updateForm(msg)
    case screenMonth:
        return a.updateMonth(msg)
    // ...
    }
}
return a, nil
```

`internal/tui/app.go` around lines 216–229. `updateForm` and
`updateMonth` are therefore only ever called with a `KeyPressMsg`.

`View` returns `tea.NewView(...)` and sets nothing else
(`internal/tui/app.go` around lines 406–418). `tea.NewView` only calls
`SetContent` (`charm.land/bubbletea/v2@v2.0.10/tea.go` lines 76–79), so
`DisableBracketedPasteMode` stays false. The renderer turns the mode
on:

```go
if !view.DisableBracketedPasteMode {
    _, _ = s.scr.WriteString(ansi.SetModeBracketedPaste)
}
```

`charm.land/bubbletea/v2@v2.0.10/cursed_renderer.go` lines 366–368.
The input translator does not turn that sequence into a key press:

```go
case uv.PasteEvent:
    return PasteMsg(e)
```

`charm.land/bubbletea/v2@v2.0.10/input.go` lines 36–37.

The text field already knows what to do with that message, and this
call path never sends it:

```go
case tea.PasteMsg:
    m.insertRunesFromUserInput([]rune(msg.Content))
```

`charm.land/bubbles/v2@v2.2.1/textinput/textinput.go` lines 654–655.
`formModel.update` would forward a non-key message
(`internal/tui/form.go` lines 128–133), but `App.Update` never calls
it with one.

On the stack this slice replaced, the same terminal mode was a key
message, not a separate type. Bubble Tea v1 returned
`KeyMsg{Type: KeyRunes, Paste: true}` from `detectBracketedPaste`
(`charm.land/bubbletea@v1.3.10/key_sequences.go` lines 109–118). No
test in `internal/tui` constructs a `tea.PasteMsg`. The fixtures are
`KeyPressMsg` values, including the new space bar
(`internal/tui/app_test.go` lines 627–631).

The widget's own ctrl+v binding is a second drop, not the regression.
It returns the clipboard `Paste` command, whose result is an
unexported `pasteMsg`, and `App.Update` does not send that result
back to the field either.

### Why this matters

Pasting a weight, a date, or a note used to arrive as runes on the
key path the month sheet already treated as typed text. It now hits
`return a, nil` and the cell does not change. Typing still works, so
the suite stays green, and the changelog says the screens behave as
before (`CHANGELOG.md` lines 67–68). A one-character paste is the
same hole: it is a `PasteMsg`, not a one-byte `KeyPressMsg`, so the
edit-start gate in O3 never sees it.

## O2 — risk — high

### Claim

`NO_COLOR` removes chroma from the palette only. The day form still
puts the text field's own 256-color styles into the model text, and
the test that locks `NO_COLOR` never opens the form.

### Evidence

The palette is the only place the program consults `NO_COLOR` before
painting:

```go
noColor := os.Getenv("NO_COLOR") != ""
// ...
if noColor {
    return s
}
return s.Foreground(lipglossColor(roleColor(cfg, role)))
```

`internal/tui/style.go` lines 52–61. The chart title does the same
check before it sets a foreground (`internal/tui/chart.go` lines
103–110).

The day form does not. `newForm` builds each field with
`textinput.New`, sets a placeholder, and never calls `SetStyles`
(`internal/tui/form.go` lines 41–52). `textinput.New` installs
`DefaultDarkStyles` (`charm.land/bubbles/v2@v2.2.1/textinput/textinput.go`
lines 157–166). Those styles are chromatic on both focus states:

```go
Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
// ...
Text: lipgloss.NewStyle().Foreground(
    lightDark(lipgloss.Color("245"), lipgloss.Color("7"))),
```

`charm.land/bubbles/v2@v2.2.1/textinput/styles.go` lines 17–27.
`LightDark(true)` keeps the dark argument, `Color("7")`
(`charm.land/lipgloss/v2@v2.0.6/color.go` lines 205–210).
`Color("240")` is an ANSI-256 colour, index 240
(`color.go` lines 87–90). Lip Gloss v2 `Render` emits that as
`38;5;` (the v2 style tests expect `38;5;234`).

`formModel.view` prints `inputs[i].View()` next to the palette label
(`internal/tui/form.go` lines 165–170). An empty field takes
`placeholderView`, which renders that placeholder style
(`textinput.go` lines 685–687 and 751–772). A blurred field that
already has a value takes the blurred text colour instead.

`TestNoColorStripsChromaKeepsStructure` sets `NO_COLOR=1`, builds the
weighed list, and reads `app.View().Content` (`internal/tui/color_test.go`
lines 219–227). That screen has no text field.

### Why this matters

S13's `NO_COLOR` clause is about the model text, not about the list
alone, and not about the writer. The list satisfies it. The day form
does not: `YYYY-MM-DD`, `optional`, `hours`, and `count` carry
`38;5;240` in `View().Content` whether or not `NO_COLOR` is set.
`hasChromaticSGR` matches `38;`.

The terminal sometimes hides this. Bubble Tea calls
`colorprofile.Detect`, and a TTY with a boolean `NO_COLOR` (the
documented `NO_COLOR=1`) selects the ASCII profile, which clears
foreground on the way out. `Detect` uses `strconv.ParseBool`
(`github.com/charmbracelet/colorprofile@v0.4.3/env.go` lines
115–117). The palette treats any non-empty value as off
(`style.go` line 52). `NO_COLOR=yes` is non-empty, so the list is
plain, and it is not a bool, so `Detect` does not strip the form.
The grey placeholders are then what the terminal shows. The suite
cannot fail either case.

This is the assumption that the palette is the only chroma in the
model. The list fixture is now the whole `NO_COLOR` requirement.

- **accept-as-stated** — widget chrome may stay coloured in the
  model. Write that down, including that the terminal strip depends
  on `Detect` and a boolean `NO_COLOR`.
- **revise-spec** — `NO_COLOR` applies to the palette and the chart
  title, not to bubbles' default field styles.
- **add-test** — open the day form under a non-empty `NO_COLOR` and
  require no chromatic SGR in the model text.
- **consciously-carry** — known, wrong for the form, shipped because
  `NO_COLOR=1` on a TTY is stripped later.

## O3 — risk — medium

### Claim

An idle non-workout month cell starts an edit only when `Text` is one
byte. A non-ASCII character does not open the cell.

### Evidence

After the space-bar check, the default branch is:

```go
default:
    if a.month.col != colWorkout && len(msg.Text) == 1 {
        a.month.beginEdit(msg.Text)
    }
```

`internal/tui/app.go` lines 395–398. `len` on a string is a byte
count. The only new typed-character fixture is
`tea.KeyPressMsg{Code: '8', Text: "8"}` (`internal/tui/app_test.go`
lines 631 and 693–705).

Once the edit is open, the field inserts runes, not bytes:

```go
m.insertRunesFromUserInput([]rune(msg.Text))
```

`charm.land/bubbles/v2@v2.2.1/textinput/textinput.go` line 647. The
byte gate is only the first character of an idle cell. The day form
does not have this gate; it forwards the key.

### Why this matters

`é` is one character and two bytes. `£` and a CJK character are one
rune and more than one byte. A two-rune grapheme is longer still.
On a note, weight, sleep, or steps cell that is not editing, that
key does not match `len(msg.Text) == 1`, does not match a command
name, and does not start an edit. The sheet does not move and
nothing is written. The operator can only begin the edit with an
ASCII byte. The space bar is already excluded above this branch, so
the `== 1` test is not what keeps the space bar from editing. The
fixture's `"8"` is now the definition of a typed character.

- **accept-as-stated** — month-cell edits are opened by an ASCII
  byte. Write that down.
- **revise-spec** — S6's "one character" means one byte, with `8`
  as the whole contract.
- **add-test** — a non-ASCII rune, and a two-rune grapheme, on an
  idle note cell either open an edit containing that text or are
  explicitly rejected.
- **consciously-carry** — known, wrong for a non-ASCII note, shipped
  because weights are ASCII.

## Explicitly not objecting to

- **modifyOtherKeys, Kitty disambiguation, and S16 wire parity**:
  accepted in `charm-v2-upgrade.md` O2. `tea.NewProgram` still
  receives only the model (`cmd/hdtools/main.go` line 89), and
  `NewView` does not set an alternate screen or a mouse mode.
- **Where 16-color downshift is judged**: accepted in O1. Profile
  tests write the view through `colorprofile.Writer`
  (`internal/tui/color_test.go` lines 44–56). The program does not
  set `lipgloss.Writer.Profile`, and the v2 renderer downsamples
  with the same `colorprofile` v0.4.3.
- **The three module paths and `go 1.27.0`**: `go.mod` requires
  `charm.land/bubbletea/v2` v2.0.10, `lipgloss/v2` v2.0.6, and
  `bubbles/v2` v2.2.1, and the `go` line is still 1.27.0. That is
  the accepted O3 gate, not a new miss.
- **The space bar itself**: `spaceBar` matches `Code == KeySpace` or
  `Text == " "` before the one-byte edit gate (`internal/tui/app.go`
  lines 323–353), and S4–S9 assert the flag, the inserted space, and
  the absence of the word "space".
- **`ctrl+c` still spelling `ctrl+c`**: a raw 0x03 is decoded as
  `Code: 'c', Mod: ModCtrl` with empty text
  (`ultraviolet` `decoder.go` lines 1158–1161), and `String()` then
  returns `ctrl+c`, which the quit cases still name. The missing
  test was an accepted spec choice, not a renamed key.
- **A 24-wide month field while editing**: `SetWidth(24)` plus
  `highlightCell`'s `TrimSpace` of `input.View()`
  (`internal/tui/month.go` lines 54 and 238–242). Bubbles v1's
  `View` padded to `Width` the same way. This slice did not add it.
- **`HARNESS.md` still naming `github.com/charmbracelet/bubbletea`**:
  the spec did not rewrite habitat files, and
  `TestModuleRequiresCharmV2` fails a v1 import, so the stale
  sentence does not by itself put the old module back in the binary.
