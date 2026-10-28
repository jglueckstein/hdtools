---
spec: docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md
date: 2026-10-28
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: premise
    severity: high
    claim: "The spec treats the open sheet's series as the page -log-pdf writes, but a sheet left open after a rejected reload is not that page, and the plan forbids the path that would make it so."
    evidence: "Decision 7: \"The key does not read the database again and does not change ApplyTrend. For a sheet the program could open, that series is the one the command writes.\" S7 and FR5: the drawn text matches \"-log-pdf\" for that log with an empty \"-o\". Plan: \"Do not call ApplyTrend. Do not re-read the store. Do not copy trendFromFirstWeighIn.\" The S7 test copies a.logs and compares logpdf.Write of that copy; it does not run the command."
    disposition: accepted
    disposition_rationale: >-
      The slice is done when shift-P writes the file -log-pdf
      writes. Decision 7 and the plan refuse the load that would
      make that true, and a rejected reload leaves the old series
      on screen. The key loads the store and trends it the way the
      command does, including a leading run with no weigh-in, then
      calls the existing writer. The rows on screen stay as they
      are. The match test compares the key's file with the
      command's file. Comparing logpdf.Write of a.logs with itself
      does not count.
  - id: O2
    category: implementation
    severity: medium
    claim: "Shift-P writes whatever series is already in memory, but a month-cell save shows saved and only then reloads, and the planned tests finish each command before the next key, so that gap stays green."
    evidence: "Decision 5 and S1: success puts the directory and the filename on the status line, and the page is the month on screen. Plan: \"The write is synchronous inside the key, then return a, nil\" and \"Mirror applyP: run a returned command\" before the next key. The same plan leaves other keys' sticky saved status alone."
    disposition: accepted
    disposition_rationale: >-
      A month-cell save shows saved and only then reloads, and a
      key in that gap can publish the pre-save series or be covered
      later by saved. Shift-P writes only when no month-cell save
      or reload is in flight. While the status is saved and the
      load has not been applied, the key does not write and does
      not clear saved. A later savedMsg does not replace a path
      this key already set. One test presses the key in that gap
      without finishing the load command first. Draining every
      command before the next key is not that test.
  - id: O3
    category: specification quality
    severity: medium
    claim: "Decision 6 requires the chart key's wording for a directory failure, while the plan requires a write log pdf wrapper, and S9's substrings accept both."
    evidence: "Decision 6: \"The wording of a directory failure is the wording chart p already shows for that cause.\" Plan: wrap with fmt.Errorf(\"write log pdf: %w\", err) and store Error() on month.err. S9 checks pdf_dir must be a string, must be absolute, home directory, and not absolute, and does not mention the wrapper."
    disposition: accepted
    disposition_rationale: >-
      The cause text stays the resolver's: pdf_dir must be a
      string, an absolute path, and the home directory. The noun
      in front of it is write log pdf, the plan's wrapper. It is
      not write chart pdf. The failure scenarios name that prefix,
      so either wrapper no longer passes.
---

## O1 — premise — high

### Claim

The spec treats the open sheet's series as the page `-log-pdf`
writes, but a sheet left open after a rejected reload is not that
page, and the plan forbids the path that would make it so.

### Evidence

Decision 7:

> The key does not read the database again and does not change
> `ApplyTrend`. For a sheet the program could open, that series is
> the one the command writes.

S7's title is "The page is the command's page." FR5 says the drawn
text matches `-log-pdf YYYY-MM` for the log the sheet is showing,
with an empty `-o`.

The plan says the opposite mechanism:

> Do not call `ApplyTrend`. Do not re-read the store. Do not copy
> `trendFromFirstWeighIn`.

`TestMonthShiftPMatchesCommand` copies `a.logs` and compares
`logpdf.Write` of that copy. It does not run the command. The plan
also says not to edit `cmd/hdtools`, which is where
`trendFromFirstWeighIn` lives, unexported.

### Why this matters

The command does not trend the way the sheet loads. `exportLogPDF`
keeps a leading run with no weigh-in on the page and runs
`ApplyTrend` only from the first weight. The TUI runs `ApplyTrend`
on the whole series and, on failure, leaves `a.logs` as it was and
sets `month.err`. An empty log and a first row that has a weight do
match. The sheet can already store the series that does not: sleep,
steps, workout, or a note, and no weigh-in. Two open sheets then
disagree with `-log-pdf` no matter how long the user waits.

- The database is empty, so the sheet opens. The user saves one
  day with no weigh-in. Reload fails. `a.logs` stays empty.
  Shift-P writes a blank month. The command writes that day's
  sleep, steps, and note.
- The sheet is showing a trended series. The user saves a
  weightless day before the first weigh-in. Reload fails. Shift-P
  writes the old series. The command writes the new day and trends
  from the first weight.

Success, as the plan specifies it, clears `month.err`. The load
error disappears and the status line names the file. The planned
oracle builds both PDFs from the same memory copy, so this stays
green. An implementer who follows the plan cannot make FR5 true.
The spec has to say which page wins when the screen and the
command disagree. "The series on screen" and "the command's page"
are not the same requirement.

## O2 — implementation — medium

### Claim

Shift-P writes whatever series is already in memory, but a
month-cell save shows `saved` and only then reloads, and the
planned tests finish each command before the next key, so that gap
stays green.

### Evidence

Decision 5 and S1 treat a successful shift-P as a stable result:
the status line contains the directory and `YYYY-MM-log.pdf`, and
the file is the month on screen. Stored logs are unchanged by the
key.

The plan:

> The write is synchronous inside the key, then `return a, nil`,
> same as `writeChartPDF`.

and:

> Mirror `applyP`: run a returned command and fail on
> `tea.QuitMsg`.

It also says this key may clear a sticky `saved`, and that other
keys stay as they are.

### Why this matters

A month-cell save does not update `a.logs` inside the key. Tab,
Enter, Down, Up, and workout Space return a command. When that
command's `savedMsg` is applied, the sheet sets status to `saved`
and only then starts `load`. Until `loadedMsg`, the painted rows
and the slice the key would pass to `logpdf.Write` are still the
previous series. Bubble Tea runs the command on another goroutine,
so the next key is not ordered behind that load.

The visible cue is `saved`. Shift-P at that moment writes the
pre-save page and replaces `saved` with the PDF path. `loadedMsg`
then updates the rows and does not rewrite the file. The sheet the
user reads after the status settles is not the file the status
names. This is a successful save of an ordinary weight, not the
rejected reload in O1. While the edit is still open, shift-P is
text, which the spec does say. The hole is the gap after `saved`
and before the reload.

`applyP`'s shape runs the command to completion before the next
key. Every listed test can pass. O2 on the edit-keys record, which
this plan leaves deferred, is the other direction: `saved` sticking
after a later key. It does not cover a success path that publishes
a stale series, or a late `savedMsg` that puts `saved` back over
the path this key just set.

## O3 — specification quality — medium

### Claim

Decision 6 requires the chart key's wording for a directory
failure, while the plan requires a `write log pdf` wrapper, and
S9's substrings accept both.

### Evidence

Decision 6:

> The wording of a directory failure is the wording chart `p`
> already shows for that cause.

The plan:

> Wrap the resolver error and the write error with
> `fmt.Errorf("write log pdf: %w", err)` before storing `Error()`
> on `month.err`.

S9 and FR6 require the fragments `pdf_dir must be a string`,
`pdf_dir` and `must be absolute`, `home directory`, and `not
absolute`. They do not mention a wrapper. The chart key's wrapper
is `write chart pdf`.

### Why this matters

Those fragments are the resolver's text. They appear under either
wrapper. An implementer who takes Decision 6 as "use the chart
sentence" shows `write chart pdf` for a log export. An implementer
who takes the plan shows `write log pdf`. Both pass
`TestMonthShiftPNonStringPDFDir`, `TestMonthShiftPRelativeDirFails`,
`TestMonthShiftPNoHomeFails`, and
`TestMonthShiftPRelativeDataHomeFails`. The failure the user can
see is a log write blamed on the chart, or a later "fix" that
copies the chart prefix onto this key to satisfy Decision 6. The
cause text can stay the chart's. The noun in front of it has to be
one of the two, and the scenarios do not pick.

## Explicitly not objecting to

- **Binding the letter P, including Caps Lock**: Decision 1 and the
  accepted slice already trade the first capital P for keeping idle
  `p`. On a legacy terminal the byte `P` is reported as Shift, so
  requiring the shift bit would not give Caps Lock back to the
  note. That is a recorded choice, not an undetected failure.
- **Dropping the key and leaving the command**: the slice already
  weighed no key, idle `p`, and a lowercase letter. This spec is
  that accepted key.
- **Page geometry, blank cells, and delta**: the 2026-10-17 spec
  owns them, and this file says it does not amend that page.
- **Sharing `writeChartPDF`'s `a.err`**: the month view paints
  `month.err`, not `a.err`, so the chart's error field would hide
  the failure. The plan's split is the one that shows it.
- **The fixture's `Code: 'P'`**: the idle branch already switches
  on `msg.String()`, and both that fixture and the decoder's
  shift-P (`Code` `p`, text `P`) stringify as `P`.
- **Sleep while editing**: Decision 9 is any cell being edited.
  Steps and the note are named so the workout neighbours are not
  special. The plan sends `P` through the editing branch, not a
  three-column list.
- **A missing `pdf_dir`**: `logpdf.Write` already creates it, and
  S9's failures are an enumerated list, not every path that is not
  there yet.
- **Retesting the directory matrix**: those cases lock this key to
  `ResolvePDFDir` instead of a second policy. The slice's "do not
  reopen `pdf_dir`" is about a new rule, which this spec does not
  add.
- **Help-line width**: the footer is already past 80 columns.
  Inserting `P log` does not create a new overflow.
- **Paste of a single `P`**: `PasteMsg` is not this key, and the
  plan does not retarget it, so a one-rune paste still starts an
  edit.
