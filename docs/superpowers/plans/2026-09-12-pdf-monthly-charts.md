# Plan — PDF of the monthly chart

**Spec**:
[`docs/superpowers/specs/2026-09-12-pdf-monthly-charts.md`](../specs/2026-09-12-pdf-monthly-charts.md)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md),
[`2026-09-11-monthly-chart-floats-sinkers.md`](2026-09-11-monthly-chart-floats-sinkers.md),
[`2026-09-12-chart-y-range.md`](2026-09-12-chart-y-range.md)
**Objections**:
[`pdf-monthly-charts.md`](../objections/pdf-monthly-charts.md)
(O1–O8 accepted)
**Status**: approved

No production code until the spec’s scenarios exist as failing tests.

## Module structure

| File | Why |
| --- | --- |
| `internal/chartspan/chartspan.go` | Shared owner (O7): last-plotted-day clip, empty predicate, Y pad, analysis-omission wrapping `MonthlyBalance`. No SQLite, no PDF, no `tui`. |
| `internal/chartspan/chartspan_test.go` | Clip at today; empty vs carry-only; omit analysis without both endpoints; Y pad in kg/lb/st. |
| `internal/chartpdf/chartpdf.go` | Writer: landscape Letter, title box, marks, stems, trend, day numbers, TUI copy. Calls `chartspan`. Does not open SQLite or import `internal/tui`. |
| `internal/chartpdf/color.go` | Hex as RGB; named 16-color values → VGA table; title `#0000FF`/`#FF0000`/`#FFFF00`; stems `#008000`. |
| `internal/chartpdf/chartpdf_test.go` | S1–S2, S5–S9, S7a, S11–S12: write under `t.TempDir()`, `%PDF`, page box, extractable title/loss/empty, vectors from the file, mode `0600`, lb axis. |
| `cmd/hdtools/main.go` | `-chart-pdf YYYY-MM` and `-o`; load **all** logs, `ApplyTrend`, same local today as the TUI; write PDF and return (no Bubble Tea). |
| `cmd/hdtools/main_test.go` | Flag path: valid month writes a file; invalid month is non-zero and creates no file; CLI write error is non-zero. |
| `internal/tui/chart.go` / `app.go` | Call `chartspan` instead of local clip/empty/Y/analysis. `p` on the monthly chart only; cwd; `0600`; overwrite; success path in status; write failure is error status, do not quit. Help line includes `p`. |
| `internal/tui/chart_test.go` | S3, S3a, S4, S10: `p` writes; write failure; list/sheet/long/editing do not (except `p` as text while editing). Existing chart tests stay green. |
| `docs/how-to/monthly-chart-pdf.md` | Export from CLI and from `p`. |
| `docs/reference/` | `-chart-pdf` / `p`. |
| `CHANGELOG.md` | Added. |
| `go.mod` | A Go PDF writer (plan: `github.com/go-pdf/fpdf`). |

`internal/tui` must not import the PDF library.
`internal/chartpdf` must not import `tui`.
Do not add `internal/chartpdf/span.go`.

## Algorithm notes

-   **Span (O1, O3, O7).** `chartspan.LastPlottedDay(today, year, month)`:
    today if the month contains today, else last calendar day of a
    past month, else empty. CLI and TUI pass the same local today.
    Load the full series, `ApplyTrend`, then clip. Carry in the span;
    no trend after last plotted day.
-   **Empty (O1, O5).** No daily marks and no trend (including no
    carry) → title + `(empty month)`, no axes, no analysis. Carry-only
    is not empty: flat trend, `Monthly loss: 0.0 … Daily deficit: 0 cal`.
-   **Analysis (O4, O5).** `MonthlyBalance` only when both endpoints
    have trend. Copy:
    `Monthly loss: %.1f %s   Daily deficit: %d cal`.
-   **Y (O7).** Shared pad: `units.ToKG(2, lb)` then `FromKG` into
    `display_unit`; Ymin = min − *P*, Ymax = max + *P*. Observed as
    extractable Y labels (one decimal), from the PDF (O6).
-   **Page (O6).** Landscape Letter, page box 792 × 612 pt, read from
    the file. Title box: fill `#0000FF`, stroke `#FF0000`, text
    `#FFFF00`. Stems `#008000`. Weight/trend: hex as RGB; names from
    the VGA table. `NO_COLOR` ignored.
-   **X.** One position per day; label 1 … last plotted day.
-   **Paint order.** Trend path, then stems, then daily marks on top.
-   **CLI.** Parse `YYYY-MM`; load config + **all** store rows; same
    today; `ApplyTrend`; write mode `0600`; exit. Write error: non-zero,
    no TUI. Tests `chdir` to `t.TempDir()` or pass `-o`.
-   **TUI `p` (O8).** `os.Getwd()` +
    `fmt.Sprintf("%04d-%02d-chart.pdf", year, month)`. Overwrite.
    Mode `0600`. Success: status contains the path. Failure: error
    status, do not quit, do not claim saved.

## Test case list

Freeze today at 10 November 1990 when the clip matters. `t.TempDir()`
for files. Inspect the PDF (page box, text, paths, mode), not a dump
that can pass without the file.

1.  `TestWritePDFHeaderAndOnePage` — November 1990 logs; `%PDF`; one
    page; page box 792 × 612 pt; mode `0600`.
2.  `TestWritePDFTitle` — extractable `November 1990`; no `hdtools —`.
3.  `TestWritePDFYRangeKG` — 80–81 kg; extractable Y labels are
    80.0−*P* and 81.0+*P*.
4.  `TestWritePDFLossAndDeficit` — 80 kg → 79 kg over 30 days; text
    has `Monthly loss: 1.0 kg   Daily deficit: 257 cal`.
5.  `TestWritePDFEmptyMonth` — June 1990 no logs and no carry; one
    page; `June 1990`; `(empty month)`; no `Monthly loss:`.
5a. `TestWritePDFCarryOnlyNotEmpty` — May log, empty June; no
    `(empty month)`; trend path; `Monthly loss: 0.0 kg` and
    `Daily deficit: 0 cal`.
6.  `TestWritePDFDisplayUnitLB` — axis text in pounds.
7.  `TestParseMonthRejects` — `1990-13`, `banana`; error; no file.
8.  `TestChartPDFFlagSkipsTUI` — CLI `-chart-pdf 1990-11 -o …`
    writes and returns (cmd test; stub or temp db).
9.  `TestChartPWritesPDF` — monthly chart, `p`; file in cwd (test
    `chdir` temp); mode `0600`; status contains path.
9a. `TestChartPWriteFailure` — unwritable path; error status; still
    on the chart; process does not quit.
10. `TestChartPIgnoredOnList` — no file.
11. `TestChartPWhileEditingIsText` — note field; no file; `p` in input.
12. `TestChartPDoesNotWriteLogs` — row count unchanged.
13. `TestWritePDFCurrentMonthClipsAtToday` — today 10 Nov 1990;
    no day label after 10; deficit divisor 10.
14. `TestWritePDFOmitsAnalysisWithoutEndpoints` — no trend on day 1
    or last plotted day; no `Monthly loss:`.
15. `TestWritePDFVectorsFromFile` — two-point month; PDF has a trend
    stroke and a stem stroke.
16. `TestColorToRGB` — `#c0a080` and `#rgb`; `blue` → `#000080`;
    `bright-red` → `#FF0000`.

CI must not require a TTY or `pdftoppm`. Optional human check of
title RGB with `pdftoppm` is not a test.

Existing monthly-chart TUI tests stay green after the `chartspan`
extraction.

Do not implement long-term or log-sheet PDFs.
