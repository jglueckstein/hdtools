---
spec: docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md
date: 2026-10-28
mode: spec
cartographer_model: grok-4.7
stories:
  - id: 1
    lens: [defaults, consequences]
    title: Letter P, not the shift bit
    disposition: accepted
    disposition_rationale: >-
      The idle month sheet matches keys on their text, so the key
      is the letter P. Caps Lock P is that key. The spec does not
      read the shift bit. Idle p still starts an edit, and P is
      text only while a cell is being edited.
  - id: 2
    lens: [forces, alternatives]
    title: Only the idle month sheet prints
    disposition: accepted
    disposition_rationale: >-
      Only an idle month sheet writes the log PDF. P on the list,
      the form, and both charts writes nothing and leaves the
      process running. Chart p still writes only the chart PDF. A
      later "P means the month in view" would have to amend S14
      through S18.
  - id: 3
    lens: [forces, alternatives]
    title: One filename, no prompt, no backup
    disposition: accepted
    disposition_rationale: >-
      The key uses the command's directory and the name
      YYYY-MM-log.pdf. A second press overwrites that file. There
      is no path prompt, no backup, and no working-directory
      fallback. Mode 0600 is that file's mode. The harness rule
      still names only the database and config.toml.
  - id: 4
    lens: [forces, consequences]
    title: Saved and the path share one line
    disposition: accepted
    disposition_rationale: >-
      Success and failure share the one status line, and the last
      finished attempt wins. A failed write clears a sticky saved
      after the reload has been applied. An in-flight save stays
      objection O2's hold. Other keys still leave sticky saved
      alone.
  - id: 5
    lens: [patterns, consequences]
    title: Drawn text, the earlier spec stays
    disposition: accepted
    disposition_rationale: >-
      The 2026-10-17 spec stays the page contract, including the
      sentence that this key comes later. The match is drawn text
      against -log-pdf, not identical bytes. The how-to gains the
      key when the key is implemented. The chart how-to stays as
      it is.
  - id: 6
    lens: [defaults, consequences]
    title: Own the keys this slice must not take
    disposition: accepted
    disposition_rationale: >-
      Idle p, chart p, P while editing, and P on the other screens
      are acceptance scenarios. A later change to those keys has
      to satisfy those scenarios as well as the earlier specs. The
      chart how-to stays the chart's instructions.
---

## Story #1 — Letter P, not the shift bit

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(Decision 1, Decision 8, Decision 9, Decision 10, US2)
**Lens:** defaults, consequences
**Refs:** —

**Context.** Idle `p` on a text cell starts an edit. Chart `p` writes
the chart. The slice already chose a shifted letter so neither of
those bindings moves. This spec still has to say whether "shifted"
means the letter P or the shift modifier.

**Forces.** The sheet's main job is typing. A print key on `p` would
take the first character of a cell. A key that is not a letter would
need a modifier test this screen does not have. Capital P is a rarer
first character than `s`, and it leaves chart `p` and idle `p` alone.
The cost is that the same character is also how Caps Lock, and a
legacy terminal, report Shift.

**Options not taken.** Require the shift modifier bit, and let a `P`
that arrives without that bit start an edit. Bind a lowercase letter
such as `s`, the slice's other candidate, and accept that it becomes
the first character of an idle cell. Treat only a chord with no
letter, or a function key, as print, so every typed letter stays
text.

**Choice as written.** The key is the letter P when the month sheet
is not editing. The spec calls that the shifted letter and equates
shift-P with the letter P. It is not a second meaning for `p`. A
chord that carries no letter is not this key. The spec never asks
for the shift modifier bit. Idle `p` still starts an edit on a text
cell and still does nothing on the workout column. While a cell is
being edited, P is text, including the steps cell and the note.
Idle P on the workout column exports, because that column is not an
edit and the letter is still the key.

**Consequences.** Caps Lock P on an idle sheet is this key. An idle
text cell cannot be opened with a leading P. P is text only after
an edit has already started, and the day form's note can still take
a leading P. Any other character that already starts an edit still
does, and idle `p` still starts one whose text is `p`. Requiring
the shift bit would not hand that leading P back on a terminal that
reports the byte P as Shift.

**Pattern.** Modal binding. The sheet already treats a typed key as
a command when idle and as text while editing. This slice adds one
command on the capital letter in the idle state only. The month
sheet already matches idle keys on their text. Chart `p` is the
text `p`. This key is the text `P` by that same habit.

## Story #2 — Only the idle month sheet prints

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(US3, Decision 4, Decision 11, S11, S14–S18, FR11)
**Lens:** forces, alternatives
**Refs:** #1

**Context.** Chart `p` lives on the monthly chart because that screen
does not take typed text. The log key follows that locality and then
stops. The letter P does not become "print the PDF for whatever is
on screen," and it does not become a shifted alias of chart `p`.

**Forces.** US3 asks that only an idle month sheet write the log, and
that chart `p` still write only the chart. A log key on the chart
would file a sheet the chart does not show. A shifted alias of `p`
would give the chart two print keys, and would make case mean
"modifier" there and "different command" on the sheet. The command
exits after it writes. The person at the sheet is still entering a
month, so a print that quit would drop the cell they were on.

**Options not taken.** Shift-P on the monthly chart writes
`YYYY-MM-log.pdf` for the month the chart is showing. Shift-P on
that chart aliases `p` and writes the chart PDF. Shift-P on the list
or the form writes the month of the day in view. After a successful
write, leave the TUI the way `-log-pdf` leaves the process.

**Choice as written.** The list, the form, the monthly chart, and the
long-term chart write no log PDF and do not change screen. Shift-P
on the monthly chart writes no chart PDF either. Chart `p` still
writes only `YYYY-MM-chart.pdf`. On the form, P is typed into a note
and is not typed into workout, and workout does not export. The key
itself does not change day, column, or month. After `[` or `]` the
file is the month now on screen, and the cell stays where those keys
left it. The process does not quit.

**Consequences.** Chart muscle memory does not transfer. P on the
chart prints nothing, and `p` there still prints only the chart.
The form is where a note can start with P, and form workout stays a
dead letter for P. A later "P means the log of the month in view"
has to amend S14 through S18. Printing stays a side effect of the
idle sheet, not a navigation command and not an exit.

**Pattern.** The chart key's locality: the command sits on the screen
it writes. This spec applies that once and does not generalize it
into a print-this-screen binding.

## Story #3 — One filename, no prompt, no backup

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(Decision 3, S1, S2, S6, FR1, FR2, FR4; out of scope, `-o`)
**Lens:** forces, alternatives
**Refs:** —

**Context.** The command already files `YYYY-MM-log.pdf` through
`ResolvePDFDir`, with `-o` as a whole-file override. Chart `p` is
that rule with an empty `-o`. The page spec already refused a
confirm-before-overwrite. This key is pressed in a session that has
a working directory and an open database, both easy to mistake for
a destination.

**Forces.** US1 is print without a path prompt and without leaving
the cell. A confirm, or a TUI `-o`, would be another screen. The
filename is the month's identity, so a second press means "write
that name again," not "keep both." The process cwd is not where
chart PDFs go, and `-db` already does not move them. An interactive
key could have grown a safer overwrite than the command. It does
not.

**Options not taken.** A path prompt, or any TUI use of `-o`. Refuse
to replace a regular file that is not already a PDF, or write a
backup beside it. Fall back to the working directory, or file the
sheet next to the database, when `pdf_dir` is omitted.

**Choice as written.** The directory is `pdf_dir` when set, otherwise
the XDG data directory. The name is `YYYY-MM-log.pdf`. There is no
working-directory fallback, no path prompt, and no second
destination. `-o` stays a command flag. The database path does not
move the file. A second press overwrites that one name. A regular
file that is not a PDF is replaced: one file, mode `0600`, bytes
beginning with `%PDF`, no backup, and no chart PDF. Mode `0600` is
the command's file mode. It is not the harness rule, which names
only the database and `config.toml`.

**Consequences.** A one-off path still requires the command. A file
that already uses the log name is destroyed without a prompt and
without a backup. A directory at that path is outside this
overwrite: the write fails, the directory stays, and the working
directory gains nothing. The key does not grow an interactive safety
policy the command lacks. Creating a missing destination directory
stays the writer's rule; this spec does not add another.

**Pattern.** A stable filename as the identity of the month's sheet,
overwritten in place. The same filing rule as chart `p` with an
empty `-o`, not a new destination policy.

## Story #4 — Saved and the path share one line

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(Decision 5, Decision 6, Decision 14, S3, S8, S22; out of scope)
**Lens:** forces, consequences
**Refs:** O2

**Context.** The month sheet has one status string. Cell saves put
`saved` there. Chart `p` puts the written path there. This key does
both jobs on that same string, and the sheet also has an error
line. The in-flight ordering between a save and this key is O2. It
is not reopened here.

**Forces.** `saved` is the word for a cell write. Left up after a
failed print, it reads as "the PDF was saved." A path and an error
together would claim a file the failure just refused. A second
status line could show the cell save and the PDF at once. The sheet
does not have that line, and this slice does not add one. Clearing
`saved` from every later key would fix the sticky word for the
whole sheet. That fix is explicitly deferred.

**Options not taken.** A second line, so `saved` and the PDF path can
both stay. Leave a previous path in place when a later press fails.
Clear sticky `saved` from every key, not only from this export's
failure. Show the configured `~/` text on success instead of the
expanded directory.

**Choice as written.** Success puts the expanded directory and
`YYYY-MM-log.pdf` on the status line, with no tilde, and clears the
error line. Failure shows the error and clears the status line,
including a sticky `saved` left after a reload has already been
applied, so the file is not claimed as saved. A `saved` that is
still in flight is O2's hold: the key does not write and does not
clear it. After the path is set, a save result that was already in
flight does not put `saved` back. Other keys do not clear sticky
`saved`.

**Consequences.** The status line is the last attempt, not a
history. A failed press forgets a path an earlier success may have
written. The file can still be on disk. A successful press forgets
`saved`, so the sheet does not say both "the cell was saved" and
"the PDF is here." Those two facts never share the screen. The
sticky word remains for every command except this export's failure
and this export's success.

**Pattern.** One status register and two producers. Each finished
outcome fills one line and blanks the other. Last attempt wins.
Not a named GoF pattern.

**Notes.** O2 already ranks an in-flight save above this key, and
ranks the path above a late `saved`. This story records the shared
register that made that ranking necessary. It does not restate the
stale-series race.

## Story #5 — Drawn text, the earlier spec stays

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(opening, Decision 2, Decision 7, S7, FR5, Documentation;
out of scope)
**Lens:** patterns, consequences
**Refs:** —

**Context.** The 2026-10-17 spec owns the page: geometry, note
cutting, delta, the filename, and the directory rule. It also says
that slice has no TUI key, and that issue #113 is the later key.
This file says it does not amend that spec, its plan, or the page.
The how-to still says there is no TUI key.

**Forces.** A second painter, or a second statement of the columns,
would fork the page. Byte identity would fail on timestamps the
drawing does not own. Spawning `-log-pdf` would match the page by
running the command, and would make the TUI parse a process that
exits. The dated spec is a change record. The how-to is the living
instruction. Updating the how-to with the failing tests would
document a key that is not there yet. Updating the old spec would
rewrite an approved slice.

**Options not taken.** Amend the 2026-10-17 spec so it no longer says
the slice has no TUI key. Restate geometry, note cutting, or delta
here, and compare bytes or a layout this slice does not own. Shell
out to `-log-pdf` and treat its stdout and exit as the key. Add a
painter in the TUI.

**Choice as written.** The key calls the existing writer. It adds no
painter and does not change `ApplyTrend` or the page rules. The
drawn text matches `-log-pdf` for that database, that month, and
the display unit, with an empty `-o`. The files need not be
byte-identical. S7 does not describe the columns, the note cut, or
the delta. The sentence in the earlier spec that issue #113 is the
later key stays, and so does the rest of that file. The key
reference and `docs/how-to/monthly-log-pdf.md` gain the step when
the key is implemented, not with the failing tests. That how-to
loses "there is no TUI key." The chart how-to is unchanged. Blank
sheets stay the earlier backlog sentence.

**Consequences.** This slice cannot catch a layout change that keeps
the words, and it does not try. Both sides use the writer the page
spec already binds. A reader of the approved 2026-10-17 spec still
sees a slice with no TUI key. That sentence is history. The how-to
is where the key becomes current, and only once the key exists.
Between the failing tests and the implementation, the how-to is
still true. Putting the how-to edit in the red-test step, or into
the approved page spec, is foreclosed.

**Pattern.** Single source of truth for the page. The earlier spec
is the contract, and this key is a client. Leaving that spec frozen
and changing the how-to later is the project's split between a
change record and a living guide.

## Story #6 — Own the keys this slice must not take

**Source:** `docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md`
(S10, S11, S12, S14–S18, FR7, FR8, FR11; out of scope)
**Lens:** defaults, consequences
**Refs:** #2

**Context.** The opening says the non-changes in one paragraph: idle
`p` still edits, and chart `p` still writes only the chart. The
out-of-scope list says them again. The acceptance section then
writes those paths as full scenarios, next to the scenarios for the
new key.

**Forces.** A spec that only describes the new press leaves the old
presses as commentary. Commentary is easy to drop when a test list
is cut. On this screen a bare letter starts an edit, so the way
this slice fails is a key that works and also steals `p`, or a
chart key that starts writing logs. Those are current behaviours.
Writing them as Given/When/Then makes a green result the contract.
That is already how this project treats a test that passes before
the change.

**Options not taken.** One out-of-scope bullet and no scenarios, so
only the new export has to go red. A preface that names the locks
and leaves the steps to the plan. Leave chart `p` out of this file,
on the grounds the chart spec already owns it.

**Choice as written.** Idle `p`, chart `p`, P while editing, and P
on the list, the form, and both charts are acceptance scenarios.
The `p` row in the key reference stays the monthly chart. The
reference also says that idle `p` on the month sheet still edits,
and that P while editing is text. Out of scope repeats the ban: do
not change idle `p`, and do not change chart `p`.

**Consequences.** This file becomes a second owner of those keys. A
later change to chart `p`, or to idle `p` on the sheet, has to
satisfy S10, S11, and S14 through S18 as well as the earlier specs.
The chart how-to is untouched, so the double ownership sits in the
acceptance record, not in the chart's instructions. A shorter spec
that left the negative space as an aside is foreclosed. A green run
of those scenarios is the pin, not a sign that the scenario was the
wrong kind of work.

**Pattern.** Characterization tests: pin behaviour the change must
not take (Michael Feathers, *Working Effectively with Legacy Code*,
2004). The out-of-scope bullets are the ban. The scenarios are the
pin.
