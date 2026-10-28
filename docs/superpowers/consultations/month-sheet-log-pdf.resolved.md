---
spec: docs/superpowers/specs/2026-10-28-month-sheet-log-pdf.md
date: 2026-10-28
state: resolved
supersedes: month-sheet-log-pdf.md
voices:
  - voice: Command-line log exporter
    source_flag: inferred
    question: When you take the file from `-log-pdf`, you copy the path it prints; is the same path on the month sheet's status line enough to take, or do you still need that printed line?
    disposition: consulted
    outcome: The status line shows the expanded directory and `YYYY-MM-log.pdf`, and that path is the file to take. `-log-pdf` still prints its own line and exits.
  - voice: Export-failure reader
    source_flag: inferred
    question: When a bad `pdf_dir` fails `-log-pdf`, you read the terminal; when that same failure shows only on the month sheet's error line, will you see it there?
    disposition: consulted
    outcome: A bad `pdf_dir` shows on the month sheet's error line as `write log pdf` plus the resolver's cause, and the status line is cleared.
  - voice: Month-sheet footer reader
    source_flag: inferred
    question: On the month-sheet help line, does `P log` mean write the filled PDF, or the sheet already on screen?
    disposition: consulted
    outcome: `P log` on the help line means write the filled log PDF for the month on screen.
  - voice: Key-table reader
    source_flag: inferred
    question: The key table will document `P` only on the idle month sheet; when you are on the chart or the list, do you treat the missing row as doing nothing, or do you try `P`?
    disposition: consulted
    outcome: On the chart and the list, where the key table has no `P` row, P does nothing.
---

This record supersedes `month-sheet-log-pdf.md`. The open file stays
pending. These are the answers given for the month-sheet key.

## Command-line log exporter

The status line shows the expanded directory and `YYYY-MM-log.pdf`,
and that path is the file to take. `-log-pdf` still prints its own
line and exits.

## Export-failure reader

A bad `pdf_dir` shows on the month sheet's error line as `write log
pdf` plus the resolver's cause, and the status line is cleared.

## Month-sheet footer reader

`P log` on the help line means write the filled log PDF for the month
on screen.

## Key-table reader

On the chart and the list, where the key table has no `P` row, P does
nothing.
