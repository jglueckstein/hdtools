---
spec: docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md
date: 2026-10-27
state: open
supersedes: null
voices:
  - voice: Month-sheet printer
    source_flag: inferred
    question: When you print the current month to take home, do you want the blank days after today on the sheet, or only the days through today?
    disposition: pending
    outcome: null
  - voice: CLI and how-to keeper
    source_flag: inferred
    question: Does the new log-sheet how-to need a line on the how-to index beside the chart PDF, or do readers find the command from the CLI reference alone?
    disposition: pending
    outcome: null
  - voice: Logger without a weigh-in
    source_flag: inferred
    question: A month that has sleep and notes but no weigh-in now prints a sheet and exits 0 instead of failing — is that the page you wanted, or should that month still be an error?
    disposition: pending
    outcome: null
---

HARNESS.md declares no ## Stakeholders section. The project declares
no stakeholder map, so every voice below is inferred from the change
rather than named by the repo.

## Month-sheet printer

When you print the current month to take home, do you want the blank
days after today on the sheet, or only the days through today?

The command copies the on-screen month sheet, which lists every
calendar day, and it deliberately does not use the chart PDF's
stop-at-today clip. Only the person who prints this page knows
whether the later blank rows are part of the sheet they take home.

## CLI and how-to keeper

Does the new log-sheet how-to need a line on the how-to index beside
the chart PDF, or do readers find the command from the CLI reference
alone?

The spec adds `docs/how-to/monthly-log-pdf.md` and lists that page
on `docs/how-to/README.md` beside the chart PDF how-to. The chart
how-to itself stays the chart page. The index is hand-kept, and it
is the list readers use to find the chart export.

## Logger without a weigh-in

A month that has sleep and notes but no weigh-in now prints a sheet
and exits 0 instead of failing — is that the page you wanted, or
should that month still be an error?

That series is one the month sheet can store and still refuses to
load, while an empty database already prints a page of blanks. Only
the person who logs sleep and a note before any weigh-in knows
whether a successful sheet, with the missing weights left empty, is
the month they wanted on paper.
