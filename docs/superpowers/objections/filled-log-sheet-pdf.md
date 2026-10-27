---
spec: docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md
date: 2026-10-17
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: "A 31-day month can pass every acceptance scenario while day 31 sits outside the 612 by 792 page, because the oracle checks page count and extracted text on February and November only."
    evidence: "Decision 1 requires one portrait page, box 612 by 792 points. Decision 3 and FR3 require one one-line row per calendar day. S4 checks November (30) and February (28 and 29) only. The acceptance preamble checks page count, the page box, and cell text read back from the PDF, not that a row's baseline lies inside the box."
    disposition: pending
    disposition_rationale: null
  - id: O2
    category: implementation
    severity: high
    claim: "A series of notes and sleep with no weigh-in fails the export, while a database with no rows prints a full sheet, so the filled log is the one that produces no paper."
    evidence: "Decision 5: an empty log skips ApplyTrend and is still a sheet of blank days, but if ApplyTrend rejects the series the export writes no file. S7 exits 0 for no rows. S33 exits non-zero when the only row has no weigh-in. ApplyTrend returns ErrNoTrendStart when the first row has a nil weight (internal/dailylog/trend.go). Upsert stores that row (internal/dailylog/store.go); weight is optional (internal/dailylog/dailylog.go)."
    disposition: pending
    disposition_rationale: null
  - id: O3
    category: specification quality
    severity: medium
    claim: "Decision 8 both drops characters Helvetica cannot draw and cuts on drawn width, but never says which happens first, so one note can keep or lose an ellipsis."
    evidence: "Decision 8: a note that fits the column is not cut; a longer note ends with `...`; a character the font cannot draw is left out and the rest stays in order. S16 requires `say (hi)` with no added `...`. S17 requires a snowman to be omitted. No scenario combines a fitting note with an undrawable character."
    disposition: pending
    disposition_rationale: null
  - id: O4
    category: specification quality
    severity: medium
    claim: "A bad month must not create config or a database, but both export flags set are not held to that rule, so a rejected command can still create the weight log."
    evidence: "Decision 10 and S28: a bad YYYY-MM is non-zero, no TUI, no PDF, and no config or database created. The both-flags sentence in Decision 10, FR17, and S29 say the process writes nothing and does not start the TUI, and they never repeat the config or database clause. In cmd/hdtools/main.go the chart path calls config.Ensure and Open after the month check, which is the shape a late both-flags check would copy."
    disposition: pending
    disposition_rationale: null
  - id: O5
    category: implementation
    severity: medium
    claim: "A note that draws nothing is specified to look the same as an empty stored note, and the export still exits 0, so a note outside Helvetica disappears with no mark."
    evidence: "S16: an empty stored note is empty, not `-`. S17: a note that is only a snowman is empty, not `-`, and the export writes the page. Decision 8 forbids replacing the missing character with `.` or `?`, and says that if nothing drawable remains the note cell is empty, not `-`."
    disposition: pending
    disposition_rationale: null
---

## O1 — specification quality — high

### Claim

A 31-day month can pass every acceptance scenario while day 31 sits
outside the 612 by 792 page, because the oracle checks page count and
extracted text on February and November only.

### Evidence

Decision 1:

> `hdtools -log-pdf YYYY-MM` writes one portrait US Letter page … The
> page box is 612 by 792 points.

Decision 3 and FR3 require one row per calendar day, one line high,
including a 31-day month. The only calendar fixtures are shorter.
S4 checks November 1990 (days 1 through 30), February 1990 (28), and
February 1992 (29). S1, S14, S16, and S17 lock "one page" on that
November fixture.

The acceptance preamble says a successful file has one page, a page
box of 612 by 792 points, and that cell text is read back from the
PDF. Nothing requires the baseline of the last row to fall inside
that box. A content stream can contain the literal `31` drawn past
y=792. Page count stays 1, extraction still finds the literal, and
every written scenario still passes. A viewer clips to the media box,
so the printed sheet does not.

### Why this matters

January is a normal month, not an edge. A layout fitted to the
November fixture (heading, header, and 30 day rows) is enough to go
green and still drop day 31 on paper. FR1's "one page" does not
distinguish "one page object" from "every day visible on that page,"
and the test method the spec names cannot tell them apart.

## O2 — implementation — high

### Claim

A series of notes and sleep with no weigh-in fails the export, while
a database with no rows prints a full sheet, so the filled log is the
one that produces no paper.

### Evidence

Decision 5:

> An empty log skips `ApplyTrend` and is still a sheet of blank days.
> … If `ApplyTrend` rejects the series, the export fails, prints no
> path, and writes no file.

S7 turns "a database with no rows" into exit 0 and a November page of
dashes. S33 turns "the only row has no weigh-in" into exit non-zero
and no PDF.

That rejection is `ErrNoTrendStart`. `ApplyTrend` returns it when the
first row's weight is nil and there is no carry
(`internal/dailylog/trend.go`). The schema allows that row.
`DailyLog` documents weight as optional, and `Upsert` stores a nil
weight (`internal/dailylog/dailylog.go`, `internal/dailylog/store.go`).
The month-sheet save builds that row on purpose: a missing day starts
as `DailyLog{Day: day}` and a sleep, steps, workout, or note edit is
upserted without a weigh-in (`internal/tui/monthsave.go`).

S14's "row with no weigh-in still prints sleep, steps, and workout"
never runs for that series. `ApplyTrend` fails before a page exists,
unless some other day, in some other month, has a weight.

### Why this matters

US1 is a filled sheet you can print without the TUI. The first days
of a log are often sleep, steps, and a note before any scale reading.
That file is valid, and it is the file S33 refuses to print. An empty
database, which has less on it, prints. The TUI load path already
drops the same series (`internal/tui/app.go` returns `loadErrMsg` and
does not keep the logs), so the screen does not show those notes
either. Copying that gate into the new command means those notes have
no paper path at all. S33 is explicit; the hole is the cliff against
S7 and against a row the editor can save.

## O3 — specification quality — medium

### Claim

Decision 8 both drops characters Helvetica cannot draw and cuts on
drawn width, but never says which happens first, so one note can keep
or lose an ellipsis.

### Evidence

Decision 8:

> A note whose drawn width fits the note column is not cut and gains
> no ellipsis. A longer note is shortened on that same line and ends
> with three periods (`...`). … A character that font cannot draw is
> left out of the cell, not replaced with `.` or `?`. The rest of the
> note stays, in order.

S16 requires the cell `say (hi)` and forbids an added `...` when that
note fits. S17 requires `a` + snowman + `b` to become `ab`, not
`a.b` or `a?b`. No given combines those rules.

Two readings both satisfy every scenario:

- Drop undrawable runes, then measure, then cut. `say (hi)` plus a
  snowman stays `say (hi)` with no ellipsis.
- Measure the stored note, cut to the column, then drop undrawable
  runes from what remains. The snowman counts toward the width (as
  itself, or as the `.` the Helvetica encoder would otherwise
  substitute). The fitting note gains `...`, or a visible tail is
  dropped, and what is left no longer matches S16's uncut cell.

### Why this matters

The page is supposed to keep every character the font can draw, in
order, and add `...` only when that drawable text does not fit. The
two rules interact on ordinary notes (a parenthesis note, an emoji, a
newline). Shipping either order passes S16 and S17, and the two PDFs
disagree on whether the note was cut.

## O4 — specification quality — medium

### Claim

A bad month must not create config or a database, but both export
flags set are not held to that rule, so a rejected command can still
create the weight log.

### Evidence

Decision 10 and S28, for a bad month: non-zero, no TUI, no PDF, "no
config file and no database are created," and `-o` is not created.

The both-flags rule in the same decision, in FR17, and in S29 is
narrower: exit non-zero, write no log PDF and no chart PDF, do not
create `-o`, do not start the TUI, including when one month string is
`banana`. They do not repeat the config or database sentence.

"Writes nothing" in this spec is the PDF-failure phrase. Decision 9
uses it for "nothing is written in the working directory" and "no
fallback directory," not for "do not call `config.Ensure`." The chart
command parses the month, then ensures config and opens the store
(`cmd/hdtools/main.go`). A both-flags check placed there, after a
valid pair of months, creates the config file and the database and
still passes S29.

### Why this matters

S28 exists so a typo cannot leave a new weight log behind. The new
failure, both flags set, is the one that can still do that if it is
wired like the export rather than like the month check. Two
implementations both print no path and write no PDF. Only one leaves
`config.toml` and `hdtools.db` behind on a command that was rejected.

## O5 — implementation — medium

### Claim

A note that draws nothing is specified to look the same as an empty
stored note, and the export still exits 0, so a note outside
Helvetica disappears with no mark.

### Evidence

S16:

> **Given** a stored note that is empty … the note cell is empty, not
> `-`.

S17:

> **Given** a note that is only a snowman … the note cell is empty,
> not `-`, and the export writes the page.

Decision 8 forbids `.` and `?`, and says that if nothing drawable
remains the cell is empty, not `-`. A day with no row is the only
note state that prints `-` (S15). Exit 0 is required whenever the
page is written (S17, FR11).

### Why this matters

Cutting a long note leaves `...`. Dropping one snowman inside `ab`
leaves `ab`, which is at least a visible change from `a?b`. A note
whose every character is outside core Helvetica becomes the same
cell as a note the person left blank, on a command that succeeded.
The prohibition on `.` and `?` removes the only in-cell signal the
spec considers. Holding the page next to the screen cannot show that
the stored note was lost. This is not a request for a new font file.
It is the collision of the two empty states Decision 8 defines
separately.

## Explicitly not objecting to

- **No TUI key:** Issue #113 and slice S2 are out of this spec on
  purpose. Shift-p is not a missing part of S1.
- **The chart painter and the chart clip:** Decision 4, S5, and S31
  keep days after today on the log sheet and leave `-chart-pdf` on
  the 792 by 612 chart page. That split is already made.
- **The PDF directory, `-o`, and mode 0600:** FR12–FR15 and S18–S27
  follow the 2026-10-03 destination rule. This record does not reopen
  relative `pdf_dir`, `~/`, or confirm-before-overwrite.
- **The hyphen as the blank mark:** Decision 7 rejects an empty gap
  and the screen's em dash. The month sheet's empty delta
  (`formatDelta` returns `""`) is a different glyph on purpose, not
  an underspecified cell.
- **Painted-cell delta:** S8–S11 match `formatDelta` and the known
  traps: 80.0 kg against a 79.9 kg trend is `+0.3` lb, a first
  176.5 lb weigh-in is `-0.1`, and an exact zero is `0.0`.
- **Not sharing the month-sheet package:** Decision 11 keeps the row
  builder in the TUI. The check is the literals above, not a second
  copy of `buildMonthSheet` hidden in the writer.
- **Header words `day` and `weekday`:** S4 locks those strings even
  though `monthHeader` prints `date` and `wd`. That is a paper label,
  not a second set of columns, and it does not change a figure.
- **Staying on core Helvetica:** The slice already refused a Unicode
  font file. O5 is only the identical empty cell, not that refusal.
- **Today's clock:** The sheet does not clip, so a UTC versus local
  today does not add or drop a row.
