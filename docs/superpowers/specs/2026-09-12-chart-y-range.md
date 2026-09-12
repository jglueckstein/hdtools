# Chart vertical range: ±2 lb

**Date**: 2026-09-12
**Status**: approved
**Issue**: [#40](https://github.com/jglueckstein/hdtools/issues/40)
**Backlog**: [`idea.md`](../../../idea.md) (Monthly Log)
**Amends**:
[`2026-09-10-monthly-charts.md`](2026-09-10-monthly-charts.md)
Decision 6 and FR15;
[`2026-09-12-long-term-charts.md`](2026-09-12-long-term-charts.md)
Decision 9.

Monthly and long-term charts currently pad Y by 5% of the data span,
or a 1.0 display-unit band when min equals max. That glues a tight
series to the frame in kg and inflates the frame in stone. The book’s
Excel charts sit the data in a **fixed 2 lb margin** above and below.

This slice replaces that padding on **every on-screen chart**: the
monthly chart and all four long-term kinds (quarterly, semiannual,
annual, complete). It is not a long-term-only rule. PDF is still out
of scope; when PDF charts land they should use this range too.

## Decision

The vertical scale is

> (minimum of plotted values) − *P*  
> through  
> (maximum of plotted values) + *P*

*P* is **2 avoirdupois pounds** expressed in `display_unit` via
`units.ToKG(2, lb)` then `units.FromKG(…, display_unit)`:

| `display_unit` | *P* |
| --- | ---: |
| `lb` | 2 |
| `kg` | 2 × 0.45359237 ≈ 0.91 |
| `st` | 2 / 14 ≈ 0.14 |

Plotted values are already in `display_unit` (daily weight and/or
trend, converted with `FromKG`). Empty charts still have no scale.

When min equals max, this is a **4 lb** band (or kg/st equivalent)
centred on that value. There is no separate 1.0-display-unit rule and
no percentage pad.

The same *P* applies to the **monthly chart** and to **all four
long-term kinds**. Long-term still uses only the series actually
plotted (both lines when *D* ≤ *W*, trend only when *D* > *W*).
Monthly plotted values remain daily weights and trend (marks, stems
do not affect min/max).

## User stories

### US1 — Headroom in every unit

As a person viewing a chart in kg, lb, or st, I want the same 2 lb of
air above and below the data so a 2 kg swing does not fill the frame
and a stone scale does not look empty.

### US2 — Flat series still has a scale

As a person with one weigh-in or a flat trend, I want a usable Y axis
without a crash.

## Acceptance scenarios

### S1 — Pounds (monthly and long-term)

**Given** `display_unit = "lb"`
**And** plotted values from 170 lb to 175 lb
**When** the monthly chart is shown
**Then** the vertical scale runs from 168 lb to 177 lb
**When** a long-term chart with the same plotted min/max is shown
**Then** the vertical scale also runs from 168 lb to 177 lb

### S2 — Kilograms

**Given** a monthly or long-term chart whose plotted values run from
80.0 kg to 81.0 kg
**And** `display_unit = "kg"`
**When** that chart is shown
**Then** the vertical scale runs from 80.0 − *P* to 81.0 + *P* with
*P* = 2 × 0.45359237 kg

### S3 — Stone

**Given** a monthly or long-term chart whose plotted values run from
12.0 st to 12.5 st
**And** `display_unit = "st"`
**When** that chart is shown
**Then** the vertical scale runs from 12.0 − (2/14) st to 12.5 +
(2/14) st

### S4 — Flat series

**Given** every plotted value is 80.0 kg on the monthly chart (or on
a long-term chart)
**And** `display_unit = "kg"`
**When** that chart is shown
**Then** the vertical scale spans 2 × *P* (≈ 1.81 kg)
**And** the process does not crash

### S5 — Empty unchanged

**Given** an empty monthly span, or an empty long-term span
**When** that chart is shown
**Then** there is no numeric Y scale

## Functional requirements

-   **FR1.** Ymin = min(plotted) − *P*, Ymax = max(plotted) + *P*,
    with *P* = 2 lb in `display_unit` as in the table above.
-   **FR2.** The same rule applies to the monthly chart **and** every
    long-term kind. A test that only checks `l` has not satisfied
    this FR.
-   **FR3.** Empty spans have no scale.
-   **FR4.** Convert with `units.ToKG` / `units.FromKG`; do not
    hard-code a second pound.

## Out of scope

-   PDF charts (must copy this range when that sitting starts)
-   User-configurable margin
-   Changing plot row count

This replaces monthly-charts Decision 6 / FR15 / S16 and long-term
charts Decision 9’s “monthly padding rule”.
