---
spec: docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md
date: 2026-10-28
state: resolved
supersedes: filled-log-sheet-pdf.md
voices:
  - voice: Month-sheet printer
    source_flag: inferred
    question: When you print the current month to take home, do you want the blank days after today on the sheet, or only the days through today?
    disposition: consulted
    outcome: The filled sheet keeps every calendar day, including days after today, and a missing cell has no mark.
  - voice: CLI and how-to keeper
    source_flag: inferred
    question: Does the new log-sheet how-to need a line on the how-to index beside the chart PDF, or do readers find the command from the CLI reference alone?
    disposition: consulted
    outcome: List the log-sheet how-to on docs/how-to/README.md beside the chart PDF.
  - voice: Logger without a weigh-in
    source_flag: inferred
    question: A month that has sleep and notes but no weigh-in now prints a sheet and exits 0 instead of failing — is that the page you wanted, or should that month still be an error?
    disposition: consulted
    outcome: A month with no weigh-in still prints, and the missing weight cells have no mark.
---

This record supersedes `filled-log-sheet-pdf.md`. The open file stays
pending. These are the answers given for the filled log sheet.

## Month-sheet printer

The filled sheet keeps every calendar day, including days after today,
and a missing cell has no mark.

## CLI and how-to keeper

List the log-sheet how-to on `docs/how-to/README.md` beside the chart
PDF.

## Logger without a weigh-in

A month with no weigh-in still prints, and the missing weight cells
have no mark.
