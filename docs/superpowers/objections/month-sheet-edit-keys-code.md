---
spec: docs/superpowers/specs/2026-10-01-month-sheet-edit-keys.md
date: 2026-10-02
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: risk
    severity: high
    claim: "A vertical save does not snapshot the cell or end the edit until an asynchronous command returns, so a second key in that window can write the wrong day, ignore Esc, or move the cursor again."
    evidence: "internal/tui/app.go:304-305 returns a.saveMonthCellAndMoveDown while editing stays true; saveMonthCell then reads a.month.cursorDay and a.month.input.Value (app.go:370-382); savedMsg applies dayDelta with no generation check (app.go:161-171). cancelEdit does not clear the buffer (month.go:254-257). flushMonthKey runs cmd() inline (app_test.go:1242-1246). Bubble Tea runs the command on another goroutine (bubbletea tea.go handleCommands)."
    disposition: accepted
    disposition_rationale: "Snapshot the day, column, and text when Enter, Down, or Up is handled. The save writes that cell. A second vertical key before the save message is applied does not move another day. Esc before the command runs cancels the save: nothing is stored and the day does not change."
  - id: O2
    category: risk
    severity: low
    claim: "After a successful Enter, Down, or Up, the month sheet keeps the status line \"saved\" through a later failed edit and through ordinary navigation."
    evidence: "savedMsg sets a.status = \"saved\" (internal/tui/app.go:172). loadErrMsg on the month screen sets a.month.err and returns without clearing a.status (app.go:184-187). The sheet renders the error and the status together (month.go:228-233). beginEdit clears m.err only (month.go:246-251)."
    disposition: deferred
    disposition_rationale: "The sticky saved line stays as Tab left it. The error line still shows a failed edit. This slice does not clear status on later keys."
  - id: O3
    category: specification quality
    severity: medium
    claim: "The Tab spec still says Enter while editing persists in place, which contradicts this slice's Enter handler and can be read as a reason to put that move back."
    evidence: "docs/superpowers/specs/2026-09-10-month-tab-next-cell.md opening: \"Enter while editing persists the cell (display unit, kilograms in the log) and stays on that cell.\" Out of scope: \"This spec's Enter persists in place.\" The handler is case \"enter\", \"down\": return a, a.saveMonthCellAndMoveDown (internal/tui/app.go:304-305)."
    disposition: accepted
    disposition_rationale: "The Tab spec opening no longer says Enter stays on the cell. The out-of-scope line no longer says this spec's Enter persists in place."
---

## O1 — risk — high

### Claim

A vertical save does not snapshot the cell or end the edit until an
asynchronous command returns, so a second key in that window can write
the wrong day, ignore Esc, or move the cursor again.

### Evidence

Enter and Down schedule the save and leave the edit open:

```304:305:internal/tui/app.go
		case "enter", "down":
			return a, a.saveMonthCellAndMoveDown
```

The command reads the live cursor and the live buffer when it runs, not
the values at the key:

```369:382:internal/tui/app.go
func (a *App) saveMonthCell() tea.Msg {
	day := a.month.cursorDay()
	// ...
		log, err = patchCell(log, a.month.col, a.month.input.Value(), a.cfg.DisplayUnit)
```

`dayDelta` is a bare `+1` or `-1` (`saveMonthCellAndMoveDown` /
`saveMonthCellAndMoveUp`). On `savedMsg` the handler always ends the
edit, applies that delta to whatever the cursor is now, and forces the
screen back to `afterSave`:

```161:171:internal/tui/app.go
	case savedMsg:
		a.month.cancelEdit()
		if msg.advance {
			a.month.nextCell()
		} else if msg.dayDelta != 0 {
			a.month.moveDay(msg.dayDelta)
		}
		// ...
		a.screen = a.afterSave
```

Esc only clears the flag. It does not drop the command, and it does not
clear the text the command will still upsert:

```301:303:internal/tui/app.go
		case "esc":
			a.month.cancelEdit()
			return a, nil
```

```254:257:internal/tui/month.go
func (m *monthModel) cancelEdit() {
	m.editing = false
	m.input.Blur()
	m.err = ""
}
```

The new tests never see this window. `flushMonthKey` runs `cmd()` on
the calling goroutine and applies `savedMsg` before the test can send
another key (`internal/tui/app_test.go` lines 1242–1246). Bubble Tea
does not. `handleCommands` starts
`go func() { msg := cmd(); p.Send(msg) }()`
(`github.com/charmbracelet/bubbletea@v1.3.10/tea.go` around lines
354–365). Until that message is applied, `Update` on the event loop can
still mutate `a.month` while the command reads it. That is a data race
on the model, including the text input's rune slice (`Value` does
`string(m.value)`).

Tab, Space, and the form already return commands that read `a` from a
goroutine. This slice is what puts a non-idempotent day move on Enter,
Down, and Up. Down and Up while editing did not previously save. Enter
used to save in place, so a late `savedMsg` did not move the cursor.

### Why this matters

The spec's scenarios are sequential: the save finishes, editing has
ended, and only then does the next Enter open the form. The code does
that only if nothing else is handled first.

A second Enter, Down, or Up while `editing` is still true schedules
another save of the same buffer and another `dayDelta`. Whichever
`savedMsg` is applied after the cursor has already moved calls
`moveDay` again, so the focus skips a day. If that command reads the
cursor after the first move, it upserts the stale buffer onto the
destination day. The user did not type that day.

Esc in the same window does not win. The upsert still commits,
`moveDay` still runs, and `a.screen = a.afterSave` sends the user back
to the month sheet if they had already Esc'd to the list (`afterSave`
stays `screenMonth` from `openMonth`). A typed digit in the window
either joins the value being stored or is thrown away when `cancelEdit`
runs, instead of starting the next cell.

A single Enter followed by a pause is fine: one local upsert usually
finishes before the next key. The failure is a key already queued, key
repeat on Down or Enter, or Esc immediately after Enter. `go test` does
not exercise that interleaving, so a green race detector on these tests
does not cover it.

- **accept-as-stated** — the round trip finishes before the next key.
  Write that down, including that Esc does not cancel an Enter whose
  command has already been returned.
- **revise-spec** — if Esc, a second vertical key, and a following
  digit must not see this window, the spec should say the edit session
  closes synchronously and the command carries a snapshot of the day,
  column, and text.
- **add-test** — send Esc or a second Enter before applying `savedMsg`,
  and run the command on another goroutine under the race detector. The
  inline `flushMonthKey` helper cannot be the only description of the
  ordering.
- **consciously-carry** — known, wrong for a repeated or immediately
  followed key, shipped because the Tab command already works this way
  and the local upsert is short.

## O2 — risk — low

### Claim

After a successful Enter, Down, or Up, the month sheet keeps the status
line "saved" through a later failed edit and through ordinary
navigation.

### Evidence

Every successful month-cell save, including the new vertical ones, sets
the sticky line and does not retire it:

```172:173:internal/tui/app.go
		a.status = "saved"
		a.err = nil
```

A failed Enter, Down, or Up is the existing `loadErrMsg` path. On the
month screen it records `a.month.err` and leaves `a.status` alone:

```184:187:internal/tui/app.go
		if a.screen == screenMonth {
			a.month.err = msg.err.Error()
			a.err = nil
			return a, nil
		}
```

The sheet prints both:

```228:233:internal/tui/month.go
	if m.err != "" {
		fmt.Fprintf(&b, "\n%s\n", p.error.Render("error: "+m.err))
	}
	if status != "" {
		fmt.Fprintf(&b, "\n%s\n", p.status.Render(status))
	}
```

Starting the next cell clears the month error and not the status
(`beginEdit`, `month.go` lines 246–251). Arrow keys, Left, and Right do
not clear it either. `openChart` does (`chart.go` sets `a.status = ""`).
Nothing in the new tests reads the status line.

This handler already did the same for Tab and for an in-place workout
save. Column fill is the new path that sets the line on every row.

### Why this matters

FR4 keeps the cursor on the cell when the text is invalid, which is
right, but the status line still says the previous commit saved.
Someone scanning that line rather than the error line can leave the
cell believing `nope` was stored. It was not. The omission is silent
after they move on. The same word stays up while they only arrow
around, so "saved" does not mean "the last key saved."

The error text is still on screen, which is why this is not a higher
severity. The assumption the code encodes is that a status string set
by an earlier success remains true.

- **accept-as-stated** — "saved" means a save has happened this visit,
  not that the last key saved. Write that down next to the status
  field.
- **revise-spec** — the editing-key spec should say the status line
  clears when a month save fails and when the next key is not itself a
  save.
- **add-test** — after a successful Enter, an invalid Enter shows the
  error and does not show `saved`. The current tests never render the
  sheet after a vertical save.
- **consciously-carry** — the error line is enough; the sticky word is
  known and left as it was for Tab.

## O3 — specification quality — medium

### Claim

The Tab spec still says Enter while editing persists in place, which
contradicts this slice's Enter handler and can be read as a reason to
put that move back.

### Evidence

The edit-keys spec says it amends the Tab spec's opening paragraph,
Decision 2, FR1, and the out-of-scope line on Enter, and that Enter's
vertical move is this spec, not Tab's
(`docs/superpowers/specs/2026-10-01-month-sheet-edit-keys.md` lines
11–15). The plan says that wording is already amended and this slice
will not edit the Tab spec again.

The working-tree Tab spec still states the old contract in two places
(`docs/superpowers/specs/2026-09-10-month-tab-next-cell.md`):

> Enter while editing persists the cell (display unit, kilograms in the
> log) and stays on that cell.

and, in Out of scope:

> This spec's Enter persists in place.

Decision 2 and FR1 were updated to say Tab does not take Enter's
vertical destination. Those two sentences were not. The implementation
follows the edit-keys spec, not those sentences:

```304:305:internal/tui/app.go
		case "enter", "down":
			return a, a.saveMonthCellAndMoveDown
```

### Why this matters

The code and the approved edit-keys spec agree. The accepted Tab spec,
which this slice claims to have amended, still tells a later reader
that Enter stays on the cell. Someone "fixing" a drift between the two
specs can put the in-place save back and still cite the Tab spec's
opening and its last out-of-scope sentence. The new tests would catch
that only if they are run. The document those tests are checked against
would not.

## Explicitly not objecting to

- **The month help line still saying `arrows move` and `enter form`**:
  Decision 15 says that line is not a second contract and must stay the
  not-editing chord list. `docs/reference/keys.md` records the editing
  chords.
- **The next Enter opening the day form**: FR12 requires it, including
  on the last day, and the second-Enter tests lock it. A held Enter can
  then reach the form's own save; that is the specified second key, not
  a separate miss.
- **Type-to-replace on the first character**: `beginEdit` still replaces
  the cell with the typed rune. US3 is Left and Right inside the buffer
  already being edited. This slice does not claim to open the stored
  value for caret editing.
- **November-only last-day and day-1 tests**: The spec's scenarios are
  November 1990. `moveDay` then `clamp` uses `daysInMonth`, the same pin
  as not-editing Up and Down, not a hardcoded 30.
- **A vertical save calling `nextCell`**: `savedMsg` takes `advance` or
  `dayDelta`, never both from `finishMonthSave`. The Enter and Down
  tests stay on the same column, which would fail if `nextCell` ran.
- **Kilograms on the vertical path**: Enter, Down, and Up go through
  `saveMonthCell` and `patchCell` with `a.cfg.DisplayUnit`, the same
  persist step as Tab. There is no second save.
- **Reloading the full series after every cell**: `savedMsg` already
  returned `a.load` for Tab. The new tests skip that follow-up on
  purpose, matching the plan, and `loadedMsg` does not touch the month
  cursor.
- **Workout Space, not-editing arrows, and not-editing Enter**: They are
  still the old branches. Space while editing inserts a space into the
  text input, which is the editing fallthrough, not a toggle.
- **Caret blink messages never reaching the input**: `App.Update`
  ignores non-key messages, which predates this slice. `Focus` leaves
  the cursor in the shown state, and the Left and Right tests assert
  `Position()`.
- **Arrow keys not calling `moveDay`**: Both paths add to `day` and
  `clamp`. They match today; I am not claiming they will drift.
