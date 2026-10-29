---
spec: docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md
date: 2026-10-28
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: >-
      After shift-P records a path, "[" , "]" , and "t" leave that
      path on the status line under a different month, so the line
      names a file the sheet is no longer showing.
    evidence: >-
      internal/tui/app.go updateMonth: "[" and "]" call prevMonth
      or nextMonth and do not assign a.status; "t" does
      a.month = newMonth(localToday()) and returns. openChart and
      openLong set a.status to "". internal/tui/logsheet.go
      writeMonthLogPDF sets a.status and a.logPDF.path to the
      file just written, and only failLogPDF or a mismatched
      noteMonthSaved clears that path. internal/tui/month.go
      view prints a.status whenever it is non-empty, under the
      title built from m.month. No test presses "]" or "t" after
      a successful P.
    disposition: accepted
    disposition_rationale: >-
      "[", "]", and "t" clear a log path when the month they land
      on is not the month that path was written for. The status
      line then does not name another month's file. A same-month
      "t" keeps the path. Arrows leave it. The file stays on disk.
      One test presses P, then "]" and a "t" that leaves the month,
      and the status line no longer names the previous file.
  - id: O2
    category: risk
    severity: high
    claim: >-
      A month-cell save started after a successful shift-P reloads
      the new rows and leaves the status on the old path, so the
      named file does not contain the save the sheet now shows.
    evidence: >-
      internal/tui/logsheet.go noteMonthSaved returns immediately
      when a.logPDF.path != "" && a.status == a.logPDF.path, and
      does not call the writer. internal/tui/app.go savedMsg then
      returns a.load. loadedMsg replaces a.logs and does not
      change a.status. Workout space and tab go through
      armLogSave and then that savedMsg. TestMonthShiftPSaveDoesNotCoverPath
      applies a savedMsg{} that does not upsert.
      Decision 14 limits the hold to a save already in flight;
      the guard has no generation, so every later save matches.
    disposition: accepted
    disposition_rationale: >-
      The path stays only for a save that was already in flight
      when shift-P wrote it. A save started after that sets the
      status to saved and clears the path. The PDF is not
      rewritten until the next shift-P, so the line stops offering
      a file that lacks the new save. S22 still covers the
      in-flight result. One test does a real workout toggle or
      weight save after P and expects the status to leave the old
      path.
  - id: O3
    category: risk
    severity: medium
    claim: >-
      A failed export is stored only on month.err, and the next
      typed character, a workout save, or a trip back through the
      list clears it while the directory failure is still true.
    evidence: >-
      failLogPDF sets a.month.err and a.status to "". beginEdit
      sets m.err to "" (internal/tui/month.go). savedMsg calls
      cancelEdit, which also sets m.err to "", before
      noteMonthSaved. openMonth does a.month = newMonth(...)
      after Esc to the list and m. assertLogWriteError reads the
      view in the same turn as the key. Idle arrows do not clear
      the field; a text character and a successful space do.
    disposition: accepted
    disposition_rationale: >-
      An export failure stays on the month sheet through a typed
      character and a workout save. The next shift-P replaces it.
      Leaving the sheet and opening the month again may clear it
      with the rest of that visit. One test uses a bad pdf_dir,
      types a character, and still sees write log pdf.
  - id: O4
    category: implementation
    severity: low
    claim: >-
      The keys reference row for P does not name YYYY-MM-log.pdf
      or the directory S20 requires, while the chart p row names
      the directory.
    evidence: >-
      docs/reference/keys.md: "| `P` | month (not editing) |
      write this month's filled log PDF |" and the p row "write
      this month's PDF in `pdf_dir` or the data directory". The
      prose under the table says the filled log for the month on
      screen and does not name the file or a directory. S20
      requires the row to say YYYY-MM-log.pdf in pdf_dir or the
      data directory. docs/how-to/monthly-log-pdf.md has that
      step; the reference row does not.
    disposition: accepted
    disposition_rationale: >-
      The P row in the key table names YYYY-MM-log.pdf in pdf_dir
      or the data directory. The p row stays the chart. The how-to
      already says this.
---

## O1 — implementation — high

### Claim

After shift-P records a path, `[`, `]`, and `t` leave that path on
the status line under a different month. The line then names a file
the sheet is no longer showing.

### Evidence

Success stores one string and nothing else records which month it
belonged to:

```go
a.status = path
a.logPDF.path = path
```

Month motion does not touch it. In `updateMonth`, `[` and `]` only
call `prevMonth` or `nextMonth`. `t` replaces the model and returns:

```go
case "t":
    a.month = newMonth(localToday())
    return a, nil
```

`openChart` and `openLong` set `a.status` to `""`. The bracket keys
and `t` do not. `month.view` prints that status whenever it is
non-empty, under a title built from `m.month` and `m.year`. The path
is cleared only by `failLogPDF` or by `noteMonthSaved` when the
status is no longer that path.

No test presses `]` or `t` after a successful P. S4 and S5 press the
bracket first, then P, which overwrites the status.

### Why this matters

The help line puts `P log`, `t today`, and `[ ] month` on one row.
Print November, press `]`, and the title is December 1990 while the
status line still ends in `1990-11-log.pdf`. That line is the only
place the key tells the user which file it wrote. The filename and
the heading now disagree, and Enter on a file manager from that path
prints the month they just left. `t` does the same jump to today and
also drops `month.err` by constructing a new model, so a failure and
a success path are both abandoned without a new write.

Pressing shift-P again fixes the file. Nothing on the December sheet
says the path is leftover.

- **accept-as-stated** — the path means the last file written, even
  under another month. Write that down next to Decision 5.
- **revise-spec** — changing month or pressing `t` clears a log path,
  the way `c` and `l` already clear status.
- **add-test** — successful P, then `]` and `t`, and the status is
  empty or names the month now on screen.
- **consciously-carry** — the month in the filename is enough of a
  warning, and a second P is the refresh.

## O2 — risk — high

### Claim

A month-cell save started after a successful shift-P reloads the new
rows and leaves the status on the old path. The named file does not
contain the save the sheet now shows.

### Evidence

`noteMonthSaved` treats "status is still the path" as the whole
question:

```go
a.logPDF.reload = true
a.err = nil
if a.logPDF.path != "" && a.status == a.logPDF.path {
    return
}
a.status = "saved"
a.logPDF.path = ""
```

That return does not write. `savedMsg` then `return a, a.load`.
`loadedMsg` replaces `a.logs` and does not change `a.status`. Workout
space and Tab both `armLogSave` and finish in that `savedMsg`. The
save count is zero before the new key, so the in-flight hold does
not apply.

`TestMonthShiftPSaveDoesNotCoverPath` injects `savedMsg{}` after P.
That message does not upsert. The test never reads the PDF again.

Spec-time O2, accepted, is the race where a save already in flight
covers the path with `saved`. Decision 14 still says "already in
flight". FR14 and the plan drop those words. The guard has no
generation, so a save the user starts after the path is showing
takes the same branch.

### Why this matters

Print, then space on workout. The toggle commits, the reload paints
`yes`, and the status line is still the PDF path. The file on disk
still says `no`. The month in the name still matches the heading, so
this lie does not look like O1. Tab on a weight does the same with a
number. The user prints from the path the sheet is still offering.

The key itself, on the next shift-P, reads the store and rewrites
the file. The status line gives no reason to press it. This is the
inverse of the stale series the key was rewritten not to publish:
the screen is the new series, and the file named underneath it is
the old one.

A rejected reload makes it worse and still deterministic. Save a
weightless day before the first weigh-in. `a.load` runs `ApplyTrend`
on the whole series and returns `loadErrMsg`. That handler sets
`month.err` and does not clear `a.status`. The path remains, the new
day is in the database, the sheet did not take the reload, and the
file was not rewritten.

- **accept-as-stated** — the path is the last write, not a promise
  the file matches the sheet. Say so on the status line.
- **revise-spec** — Decision 14's "already in flight" is the whole
  hold. A save started after the key may replace the path, or the
  key's writer runs again for that save.
- **add-test** — P, then a real workout toggle or weight save, then
  the PDF bytes change or the status no longer names the old file.
- **consciously-carry** — one path slot cannot say both "saved" and
  the directory, and a second P is the refresh. The race fix stays.

## O3 — risk — medium

### Claim

A failed export is stored only on `month.err`. The next typed
character, a workout save, or a trip out to the list and back clears
it, and the directory failure is still true.

### Evidence

`failLogPDF` sets `a.month.err` and sets `a.status` to `""`. The
sheet paints `error:` plus that field.

`beginEdit` runs for a one-rune key on a text cell and sets
`m.err` to `""`. `savedMsg` calls `cancelEdit` before
`noteMonthSaved`, and `cancelEdit` clears the same field. Space on
workout is a save, so it takes that path. `openMonth` does
`a.month = newMonth(...)`, so Esc to the list and `m` drop the
error without showing it on the list. The list paints `a.err`, which
`failLogPDF` set to nil so a leftover would not replace the list.

`assertLogWriteError` reads the view in the same turn as the key.
Nothing presses a letter, space, or Esc afterward. Idle arrows do
not clear the field.

### Why this matters

`pdf_dir` still a file, or still relative, is not fixed by typing a
weight. The next digit is also the ordinary way to edit the cell
they are on, and it erases the only sentence that said the export
did not happen. Status is already empty, so the sheet looks like a
normal month. Shift-P fails again the same way, but only if they
press it again. Esc and `m`, which the help line calls `esc list`,
rebuilds the month model and the failure is gone for good.

A cell parse error cleared on the next edit is the same field doing
the job it had. An export failure is not that job. The condition
outlives the edit.

- **accept-as-stated** — month errors last until the next edit or
  until the sheet is opened again. The PDF failure is one of those.
- **revise-spec** — an export failure stays until the next shift-P
  or until they leave the process, and is not cleared by `beginEdit`.
- **add-test** — a bad `pdf_dir`, then a typed character and a
  workout space, and the error is still the export.
- **consciously-carry** — the error is on screen until they type,
  which is long enough, and the next P shows it again.

## O4 — implementation — low

### Claim

The keys reference row for `P` does not name `YYYY-MM-log.pdf` or
the directory. S20 requires both. The chart's `p` row names the
directory.

### Evidence

`docs/reference/keys.md`:

```text
| `p` | monthly chart | write this month's PDF in `pdf_dir` or the data directory |
| `P` | month (not editing) | write this month's filled log PDF |
```

The prose under the table says idle `P` writes the filled log for
the month on screen, including after `[` and `]`. It does not name
the file or a directory. S20's row is `P` on the month sheet when
not editing, writing `YYYY-MM-log.pdf` in `pdf_dir` or the data
directory. `docs/how-to/monthly-log-pdf.md` has the filename, the
shared directory, shift-P, and the brackets. The chart how-to is
unchanged. The reference row is the piece that does not.

### Why this matters

The how-to is right. The table is the lookup S20 names, and it is
the table that already tells `p` where the chart file goes. `P`
reads as "some PDF" with no name and no directory. Someone who
never opens the how-to cannot tell this file from the chart file,
or whether it lands next to the database. That is a missed Then,
not a wrong sentence. The behavior of the key is the one O1 and O2
describe; this row does not cause that behavior.

## Explicitly not objecting to

- **The save count:** in-flight month-cell saves are an atomic
  count. A stale `savedMsg` releases one, and a drop or a failed
  save releases itself on the command goroutine.
  `TestMonthShiftPDuringNewerSave` keeps a newer workout save armed
  across an older result. The one-boolean bug is not in this code.
- **The reload bool:** `loadedMsg` and every `loadErrMsg` clear one
  flag, so a second load can open the key while another reload is
  still queued. The key reads the store, and an upsert that has not
  finished still holds the save count, so that window does not
  write the pre-save series.
- **The doubled `write log pdf` prefix:** `logpdf.Write` already
  prefixes the write error, and `failLogPDF` prefixes again, the
  same shape as chart `p`. S8 asks for this noun and not the chart
  noun. Resolver failures are not pre-wrapped, so those stay one
  prefix.
- **The shared preparation:** the key and `exportLogPDF` both call
  `RejectUnsorted` and `TrendFromFirstWeighIn`, and
  `TestMonthShiftPMatchesCommand` runs the real command.
  `ApplyTrend` on the whole series errors when the leading day has
  no weigh-in, and that command then writes no file.
- **`Today`:** the painter does not read it. Passing `localToday`
  instead of `time.Now` cannot change a page.
- **Paste of a single `P`:** `paste` still begins an edit. Only the
  idle `P` branch exports.
- **Idle `p`, chart `p`, and `P` on the list, the form, and both
  charts:** those locks call the old paths, and the month help line
  is the Decision 12 string.
- **Mode `0600`:** it stays inside `installOwnerOnly`. The TUI does
  not `os.Create` the destination.
