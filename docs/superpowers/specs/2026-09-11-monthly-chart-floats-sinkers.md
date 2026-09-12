# Monthly chart title box and float/sinker stems

**Date**: 2026-09-11
**Status**: approved
**Issue**: [#36](https://github.com/jglueckstein/hdtools/issues/36)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Follows**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md)

The on-screen monthly chart already plots daily marks and a trend
path. Excel's monthly chart also has a month-year title box above the
plot and green vertical stems from each daily mark to the trend
(floats and sinkers). Those belong in the TUI now, and in PDF monthly
charts when those exist.

## Decisions

1.  **Title box.** Centered above the plot, text is the month name and
    year (`June 1990`), not `hdtools — June 1990`. Excel: red border,
    blue background, yellow foreground. The TUI maps those to the
    16-color set (`red` / `blue` / `yellow`) as far as the terminal
    allows. They are chart chrome, not new `[colors]` roles in this
    slice (defaults only). Under `NO_COLOR`, the same text stays
    centered with no chromatic SGR.
2.  **Stems.** For each day with a daily mark, draw a thin vertical
    stem from that mark to the trend on that day (Excel: green). If
    the mark sits on the trend, no stem (zero length). The stem does
    not replace the daily mark or the trend path. Daily mark still
    wins the shared cell. Under `NO_COLOR`, stems stay as a distinct
    glyph from mark and path.
3.  **PDF.** When PDF monthly charts are built, they use this title
    box and these stems, not a plot without them.

## Out of scope

-   User-configurable title-box or stem colours
-   Implementing PDF in this follow-on (intent only)
-   Long-term charts
