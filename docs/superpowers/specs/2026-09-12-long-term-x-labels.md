# Long-term chart: month-start X labels

**Date**: 2026-09-12
**Status**: draft
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Amends**:
[`2026-09-12-long-term-charts.md`](2026-09-12-long-term-charts.md)
Decision 6 (X resolution).

The long-term plot currently labels the axis like the monthly chart
(day 1 and last plotted day as numbers). For a quarter or a year that
is the wrong grain. Labels belong at **month starts**, in the book's
`Aug 26` style (abbreviated month + two-digit year), and they must
not collide when buckets are wider than a month's column gap.

This slice is **long-term charts only**. The monthly chart keeps
numeric day labels.

## Decision

1.  **Where.** Place a label at the plot column that contains the
    **first day of each calendar month** in the span. If the span
    starts mid-month (complete history from 15 April), that first
    plotted day still gets that month's label (column 0 is `Apr 89`).
2.  **Text.** English three-letter month + two-digit year:
    `Jan` … `Dec` and `89` for 1989, `26` for 2026. Example:
    `Aug 26` is August 2026, not 26 August.
3.  **Fit, in this order, left to right:**
    1.  **One line** `Aug 26` (six columns including the space) if
        that run of columns does not overlap the next *placed*
        label.
    2.  Else **two lines**, `Aug` over `26` (three columns: the
        longer of the two tokens). Same overlap rule.
    3.  Else **skip months** until the next two-line label fits.
        Example: `Aug`/`26` then the next placed label is
        `Nov`/`26`, not `Sep` or `Oct`.
4.  **Overlap.** A label occupies columns
    `[start, start + width)`. It overlaps another if those ranges
    intersect. The gutter is not part of the plot; labels sit under
    plot columns only. Labels may extend to the right of the last
    column only if clipped (do not wrap onto a third line).
5.  **Skip is greedy.** Walk month starts from the left. Place the
    richest form that fits against the previous *placed* label
    (and the plot's remaining width). Skipped months have no label.
    The first month in the span is always placed, using two-line if
    one-line does not fit in the gap to the next month start (or to
    the plot end).
6.  **Empty span.** No X labels.

## User stories

### US1 — Read the year at a glance

As a person on a quarterly or annual chart, I want month names under
the plot (`Sep 90`) so I am not counting day numbers.

### US2 — Dense history still labeled

As a person on complete history with wide buckets, I want fewer
labels (`Aug` over `26`, then `Nov` over `26`) rather than overlapping
garbage.

## Acceptance scenarios

November 1990 fixtures: freeze today at 10 November 1990 unless noted.

### S1 — Quarterly one-line

**Given** a quarterly chart ending November 1990 (start 1 September)
**And** one column per day (*D* ≤ *W*)
**When** the chart is shown
**Then** X labels include `Sep 90`, `Oct 90`, and `Nov 90`
**And** each sits under the column for that month's first day
**And** those labels are on one line

### S2 — Two-line when one-line will not fit

**Given** a span whose month-start columns are fewer than 6 columns
apart and at least 3 apart
**When** the chart is shown
**Then** each placed label is two lines (`Sep` above `90`)
**And** no one-line `Sep 90` is used

### S3 — Skip months when two-line will not fit

**Given** a complete-history chart whose month-start columns are
fewer than 3 columns apart
**When** the chart is shown
**Then** some months have no X label
**And** each placed label is two lines
**And** consecutive placed labels do not overlap

### S4 — Mid-month start

**Given** complete history whose first log is 15 April 1989
**When** the chart is shown
**Then** the first X label is `Apr 89` at column 0

### S5 — Empty

**Given** an empty long-term span
**When** the chart is shown
**Then** there are no `Jan`…`Dec` X labels

## Observing labels

Strip ANSI. A **one-line** label is the substring `Mon YY` on a
single row under the plot (after the Y gutter). A **two-line** label
is `Mon` on one row and `YY` on the next, sharing the same starting
column. Tests must not require a TTY.

## Functional requirements

-   **FR1.** Label month starts (and a mid-month span start) as
    `Mon YY`.
-   **FR2.** Prefer one line; if that overlaps, two lines; if that
    overlaps, skip months. Greedy left to right.
-   **FR3.** The monthly chart is unchanged (numeric day labels).

## Out of scope

-   Localised month names
-   Four-digit years
-   Labels at mid-month or week starts other than the rules above
-   PDF

This replaces any implication in the long-term spec that X labels
are day numbers at the ends of the plot.
