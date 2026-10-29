---
spec: docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md
date: 2026-10-28
state: open
supersedes: null
voices:
  - voice: Command-line log exporter
    source_flag: inferred
    question: When you take the file from `-log-pdf`, you copy the path it prints; is the same path on the month sheet's status line enough to take, or do you still need that printed line?
    disposition: pending
    outcome: null
  - voice: Export-failure reader
    source_flag: inferred
    question: When a bad `pdf_dir` fails `-log-pdf`, you read the terminal; when that same failure shows only on the month sheet's error line, will you see it there?
    disposition: pending
    outcome: null
  - voice: Month-sheet footer reader
    source_flag: inferred
    question: On the month-sheet help line, does `P log` mean write the filled PDF, or the sheet already on screen?
    disposition: pending
    outcome: null
  - voice: Key-table reader
    source_flag: inferred
    question: The key table will document `P` only on the idle month sheet; when you are on the chart or the list, do you treat the missing row as doing nothing, or do you try `P`?
    disposition: pending
    outcome: null
---

HARNESS.md declares no ## Stakeholders section. The project declares
no stakeholder map, so every voice below is inferred from the change
rather than named by the repo.

## Command-line log exporter

When you take the file from `-log-pdf`, you copy the path it prints;
is the same path on the month sheet's status line enough to take, or
do you still need that printed line?

`-log-pdf` is still the only shipped way to write this file. The
command reference and the log how-to both teach the path as one line
of output, and the process then exits. The key stays on the cell and
puts the expanded path on the status line. Only the person who
copies that line today knows whether the status line is a path they
can take.

## Export-failure reader

When a bad `pdf_dir` fails `-log-pdf`, you read the terminal; when
that same failure shows only on the month sheet's error line, will
you see it there?

From the command, a bad directory prints a cause and no path. This
key puts that cause on the sheet's error line and clears the status
line. Only the person who fixes `pdf_dir` knows whether they look at
that line, or only at the terminal, where this key prints nothing.

## Month-sheet footer reader

On the month-sheet help line, does `P log` mean write the filled PDF,
or the sheet already on screen?

That line is the only place the sheet names the key. It is written
like `c chart` and `l long`, and "log" was chosen because the file
is a log PDF. Only the person reading the footer knows whether "log"
means that file or the month already on screen. If it means the
sheet, the label does not name a print.

## Key-table reader

The key table will document `P` only on the idle month sheet; when
you are on the chart or the list, do you treat the missing row as
doing nothing, or do you try `P`?

The reference gains a `P` row for the idle month sheet, and the `p`
row stays the chart. It does not say what `P` does on the chart, the
list, or the form. That table is one row per action. Only a reader
of it knows whether a missing row means the key does nothing there.
