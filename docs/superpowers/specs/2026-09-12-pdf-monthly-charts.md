# PDF of the monthly chart

**Date**: 2026-09-12
**Status**: approved
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md),
[`2026-09-11-monthly-chart-floats-sinkers.md`](2026-09-11-monthly-chart-floats-sinkers.md),
[`2026-09-12-chart-y-range.md`](2026-09-12-chart-y-range.md)
**Objections**:
[`pdf-monthly-charts.md`](../objections/pdf-monthly-charts.md)
(O1–O8 accepted)

`idea.md` asks for PDFs of the monthly charts. The on-screen chart
already has the book's look: month-year title box, daily marks, green
stems, trend path, ±2 lb Y, Monthly Loss and Daily Deficit. This
slice writes **that** picture to a PDF, not a dump of terminal
glyphs and not a stripped sparkline.

Filled monthly **log sheets** and **blank** sheets are a different
sitting. Long-term chart PDFs are out of scope.

## Background

Excel printed a landscape monthly chart. The TUI cannot be the print
artefact: character cells, 16-color ANSI, and `NO_COLOR` are screen
rules. PDF is vector: real lines, a filled title box, colours even
when the terminal is monochrome.

Data and arithmetic stay the same as the on-screen monthly chart
(same month clip, `ApplyTrend`, `MonthlyBalance`, `display_unit`).
Geometry does **not**: one character per day does not apply.

## User stories

### US1 — Take a month home

As a person who has a month of logs, I want a one-page PDF of that
month's chart so I can print or file it without a screenshot of the
TUI.

### US2 — Same look as the screen, on paper

As a person who knows the on-screen chart, I want the PDF to show the
title box, floats-and-sinkers, trend, Y margin, and loss numbers, not
a different plot.

### US3 — Empty month

As a person who exports a month with no daily marks and no trend, I
want a PDF that says `(empty month)` rather than a crash or a
zero-byte file.

## Decisions

1.  **One page, landscape US Letter.** One month, one page. Not A4
    in this slice (the book tools were US Excel). The PDF page box is
    792 × 612 pt (Letter landscape). Tests read that box from the
    file; they do not rasterise.
2.  **How to run.** Two entry points share one writer:
    - **CLI:** `hdtools -chart-pdf YYYY-MM` writes
      `YYYY-MM-chart.pdf` in the current directory and **does not**
      start the TUI. `-o path` overrides the output path.
    - **TUI:** on the monthly chart, `p` writes the same PDF for the
      chart's month into the current directory and sets a status
      line with the path. `p` does not write while on the list,
      form, month sheet, or long-term chart.
    `cmd/hdtools` stays wiring: flags, open store, call the writer.
    The writer lives under `internal/` (not in `internal/tui`, which
    must stay SQL-free and must not grow a PDF dependency).
    Both paths load the **full** log series, `ApplyTrend`, then clip
    to the month (O1). CLI and TUI use the same local today (O3).
3.  **Vector, not a TUI screenshot.** Lines, a filled rectangle for
    the title box, and text. Not a raster of the terminal, not
    Braille/block cells.
4.  **Same data as the on-screen monthly chart.** Plotted span is
    day 1 through last plotted day (today if the month contains
    today; last calendar day if past). Daily marks, trend path,
    stems (zero-length omitted), carry-forward in the span, no
    trend into the future. Y is min − *P* through max + *P* with
    *P* = 2 lb in `display_unit`
    ([2026-09-12-chart-y-range.md](2026-09-12-chart-y-range.md)).
    If day 1 and the last plotted day both have trend, Monthly Loss
    and Daily Deficit use `MonthlyBalance` on that span; if either
    endpoint has no trend, omit both numbers (O4). Empty means the
    plotted span has no daily marks and no trend, including no
    carry (O1). Then: title + `(empty month)`, no axes, no
    analysis. A month with no log rows **but** a carried trend is
    not empty: plot the flat trend; Monthly Loss 0 and Daily
    Deficit 0.
5.  **Title box.** Centered above the plot. Text is the month name
    and year (`November 1990`), not `hdtools — …`. Excel chrome,
    not `[colors]` roles, as explicit RGB (O2): fill `#0000FF`,
    stroke `#FF0000`, text `#FFFF00`. PDF is a print artefact:
    **always colour**; `NO_COLOR` does not apply. Title RGB is a
    human PNG check, not a CI gate (O6).
6.  **Series colours.** Daily marks use the configured `weight`
    role, the trend path the `trend` role. Hex `#rrggbb` / `#rgb`
    is RGB as written (`#rgb` doubles digits). Named 16-color
    values (and 0–15 aliases) use this VGA table (O2):

    | Name | RGB |
    | --- | --- |
    | `black` | `#000000` |
    | `red` | `#800000` |
    | `green` | `#008000` |
    | `yellow` | `#808000` |
    | `blue` | `#000080` |
    | `magenta` | `#800080` |
    | `cyan` | `#008080` |
    | `white` | `#C0C0C0` |
    | `bright-black` | `#808080` |
    | `bright-red` | `#FF0000` |
    | `bright-green` | `#00FF00` |
    | `bright-yellow` | `#FFFF00` |
    | `bright-blue` | `#0000FF` |
    | `bright-magenta` | `#FF00FF` |
    | `bright-cyan` | `#00FFFF` |
    | `bright-white` | `#FFFFFF` |

    Stems are `#008000` (Excel green, not a `[colors]` key). Daily
    mark is drawn on top of the trend where they meet (book diamond
    on the line).
7.  **X axis.** One column per day of the plotted span (real
    coordinates, not character cells). Label **every day number**
    under its column (1 … last plotted day). PDF has room; do not
    use long-term `Sep 90` labels.
8.  **Copy is not medical.** No diagnosis, treatment, prescription,
    or medical advice. Loss / deficit stay the book's 3500 kcal/lb
    identity. Extractable strings match the TUI (O5): empty is
    `(empty month)`; analysis is
    `Monthly loss: %.1f %s   Daily deficit: %d cal`.
9.  **File.** Valid PDF (`%PDF` header), one page, mode `0600`
    (same personal series as the database). Tests write under
    `t.TempDir()`. Overwrite if the path exists. On write failure:
    CLI exits non-zero and does not start the TUI; TUI `p` sets an
    error status, does not quit, and does not claim the file was
    saved (O8).
10. **Current directory** for TUI `p` is the process working
    directory, not XDG. Config and the database stay XDG.
11. **One owner for data rules (O7).** Last-plotted-day clip, empty
    predicate, Y pad, and analysis-omission live in one helper both
    the TUI and the PDF writer call. The writer does not copy those
    rules. `internal/tui` still must not import a PDF library.

## Acceptance scenarios

November 1990 is a fixture. The title is the **actual** month.
Freeze today at 10 November 1990 when the clip matters.

### S1 — CLI writes a PDF and skips the TUI

**Given** logs in November 1990
**When** the CLI writes `-chart-pdf 1990-11 -o <tmp>/out.pdf`
with `-db` and `-config` under tmp
**Then** the process exits 0
**And** `<tmp>/out.pdf` exists, starts with `%PDF`, and has one page
**And** the page box is 792 × 612 pt (landscape US Letter)
**And** the file mode is `0600`
**And** the process does not require a TTY

### S2 — Title and look

**Given** that PDF
**When** text is extracted from the file
**Then** it contains `November 1990` and does not contain `hdtools —`
**And** the PDF contains vector paths for daily marks, a trend, and
at least one stem (not a screenshot of `o` / `|` cells)
**And** a human PNG may check title fill `#0000FF` / stroke
`#FF0000` / text `#FFFF00`; that is not a CI gate

### S3 — TUI `p`

**Given** the monthly chart for November 1990
**When** `p` is pressed
**Then** `1990-11-chart.pdf` is written in the working directory
**And** the file mode is `0600`
**And** the status line contains that path
**And** the screen is still the monthly chart

### S3a — TUI `p` write failure

**Given** the monthly chart
**And** the output path cannot be written
**When** `p` is pressed
**Then** the status line is an error (not a saved-path claim)
**And** the process does not quit
**And** the screen is still the monthly chart

### S4 — `p` is not a write on other screens

**Given** the daily list (or month sheet, or long-term chart)
**When** `p` is pressed
**Then** no PDF is created
**And** on the month sheet while editing, `p` is text in the field

### S5 — Y range

**Given** plotted values from 80.0 kg to 81.0 kg and `display_unit = "kg"`
**When** the PDF is written
**Then** extractable Y-axis labels include 80.0 − *P* and 81.0 + *P*
with *P* = 2 × 0.45359237 kg (one decimal, from the PDF)

### S6 — Loss numbers

**Given** first-day trend 80.0 kg, last-day trend 79.0 kg, 30-day
November 1990 (past month)
**When** the PDF is written
**Then** extractable text includes
`Monthly loss: 1.0 kg   Daily deficit: 257 cal`

### S7 — Empty month

**Given** no logs in June 1990 and no carried trend into June
**When** `-chart-pdf 1990-06` runs
**Then** a one-page PDF is written
**And** extractable text includes `June 1990` and `(empty month)`
**And** it does not include `Monthly loss:`
**And** there is no crash

### S7a — Carry-only month is not empty

**Given** no log rows in June 1990 and a carried trend from May
**When** `-chart-pdf 1990-06` runs
**Then** extractable text does not include `(empty month)`
**And** the PDF has a trend path
**And** extractable text includes
`Monthly loss: 0.0 kg   Daily deficit: 0 cal`

### S8 — Invalid month

**Given** `-chart-pdf 1990-13` or `-chart-pdf banana`
**When** the command runs
**Then** it exits non-zero
**And** it does not write a PDF
**And** it does not start the TUI

### S9 — Display unit

**Given** `display_unit = "lb"`
**When** the PDF is written
**Then** axis labels are in pounds

### S10 — No log write

**Given** the monthly chart
**When** `p` is pressed
**Then** SQLite row count is unchanged

### S11 — Current month ends at today

**Given** today is 10 November 1990
**And** the November 1990 chart (CLI or TUI `p`)
**When** the PDF is written
**Then** there is no day-number label after 10
**And** Daily Deficit divides by 10, not by 30

### S12 — No analysis without both endpoints

**Given** a month with no trend on day 1 or no trend on the last
plotted day
**When** the PDF is written
**Then** extractable text does not include `Monthly loss:`

## Observing the PDF

Tests must not require a TTY or `pdftoppm`. Observations are of the
PDF file (O6):

-   **Valid PDF:** starts with `%PDF`.
-   **Page:** one page; page box 792 × 612 pt (landscape US Letter).
-   **Extractable text:** strings the writer embeds (title, axis
    labels, `(empty month)`, analysis line). Y bounds are the
    numeric labels at the ends of the Y axis (one decimal).
-   **Vector paths:** at least one trend stroke and at least one
    stem stroke on a two-point month, read from the PDF (content
    stream or a specified inspector **of that file**). Not an
    optional debug dump that can pass without a real path.
-   **Mode:** `0600`.
-   **Title RGB:** optional human PNG (`pdftoppm`); not a CI gate.

## Functional requirements

-   **FR1.** `-chart-pdf YYYY-MM` writes a one-page landscape Letter
    PDF of that month’s chart and does not start the TUI. `-o`
    sets the path; default is `YYYY-MM-chart.pdf` in the cwd.
    Page box is 792 × 612 pt.
-   **FR2.** `p` on the monthly chart writes that file for the
    current chart month. Success: status line contains the path.
    Write failure: error status, do not quit, do not claim saved.
-   **FR3.** Title box, stems, daily marks, trend, ±2 lb Y, day
    numbers, Loss / Daily deficit match the on-screen monthly
    chart’s *data* rules (clip, carry, empty, analysis omission).
    Geometry is vector, not character cells.
-   **FR4.** Empty month: PDF with title and `(empty month)`, exit
    0. A carry-only span is not empty.
-   **FR5.** Bad `-chart-pdf` value: non-zero exit, no TUI, no file.
-   **FR6.** Copy is not medical. `NO_COLOR` does not grey the PDF.
    Series colours follow Decision 6. Extractable analysis and
    empty strings match the TUI (Decision 8).
-   **FR7.** Application code is under `internal/`, not in
    `cmd/hdtools` beyond flag wiring, and not in `internal/tui`
    as a PDF library import. Clip, empty, Y pad, and
    analysis-omission have one owner both renderers call.
-   **FR8.** Written PDFs are mode `0600`. Overwrite if the path
    exists.

## Out of scope

-   PDF of filled or blank monthly **log sheets**
-   PDF of long-term charts
-   A4 / portrait / multi-page
-   Preview inside the TUI
-   Email or print spool
-   User-configurable title-box colours
-   Confirm-before-overwrite; XDG as the PDF destination

## Documentation

A how-to (export a monthly chart PDF) and a reference line for
`-chart-pdf` / `p`. Diátaxis split unchanged.
