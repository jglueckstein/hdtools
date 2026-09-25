# Weight−trend delta column

**Date**: 2026-09-22
**Status**: approved
**Backlog**: [`idea.md`](../../../idea.md) (Daily Log)
**Follows**:
[`2026-09-09-color-schemes.md`](2026-09-09-color-schemes.md)
**Amends**: that spec's Color roles table (adds `delta-pos`,
`delta-neg`, `delta-zero`)
**Objections**:
[`delta-column.md`](../objections/delta-column.md)
(O1–O7 accepted)

`idea.md` asks for a calculated column to the right of trend: daily
weight minus trend, in the display unit, color-coded, derived not
stored, and readable without color. This slice adds that column on the
**daily list** and the **month sheet**. Charts and PDFs are unchanged.

## User stories

### US1 — See how far today's weigh-in sits from trend

As a person looking at the daily list or month sheet, I want the
difference between that day's **displayed** weight and **displayed**
trend so I do not subtract in my head.

### US2 — Read the sign without color

As a person on a 16-color terminal or with `NO_COLOR`, I want a
leading `+` or `-` (and `0.0` for zero) so the sign is in the text.

### US3 — Blank when there is nothing to subtract

As a person looking at a day with no weigh-in or no trend, I want the
delta cell empty, not a fake zero.

## Decisions

1.  **Formula (O2).** Take the same `FromKG` values already shown in
    the weight and trend cells. Subtract those two display-unit
    numbers, then format (Decision 3). Do not convert a kilogram
    residual. Both operands must be present: a logged weight and a
    displayed trend (the same `ApplyTrend` / `HasTrend` value as the
    trend cell). When the two painted cells are equal, print `0.0`.
    The first weigh-in is `0.0` in kg; in lb/st it is the displayed
    subtraction (may be signed).
2.  **Blank, not zero.** If either operand is missing, the cell is
    empty (spaces, not `—`, not `0.0`). Carry-forward trend with no
    weigh-in is S4. Days before the first weigh-in have no trend.
3.  **Text format (O3).** Round the display-unit difference to one
    decimal **then** pick the sign. If the rounded value is 0, print
    `0.0` (never `+0.0` or `-0.0`). If it is greater than 0, print
    `+` and one decimal. If it is less than 0, print `-` and one
    decimal (the minus is the sign). No unit suffix. Stone uses the
    same one decimal; a small kg residual may display as `0.0`.
4.  **Column.** Header `delta`, immediately right of `trend` on both
    list and month sheet. Shared widths in `layout.go`. Not an
    editable month-sheet column: Tab still goes weight → sleep.
    Type/Space/Enter do not edit delta.
5.  **Colour roles (O4, O7).** This spec adds known `[colors]` keys
    `delta-pos`, `delta-neg`, `delta-zero`. Invalid or omitted values
    drop silently. Built-in defaults: positive `yellow`, negative
    `green`, zero `white`. `NO_COLOR` strips chroma; `+` / `-` /
    `0.0` remain.
6.  **Selection (O6).** On the **list**, reverse covers the whole
    selected row, including delta (same as weight and trend). On the
    **month sheet**, delta is never the focused cell, so it keeps its
    role colour (colour-schemes Decision 12).
7.  **Not stored.** No SQLite column.
8.  **Out of scope.** Charts, PDFs, extra daily fields, spreadsheet
    Enter/Up/Down, high-resolution chart paint.

## Acceptance scenarios

### S1 — Positive delta

**Given** a day with weight 80.5 kg and trend 80.0 kg, `display_unit = "kg"`
**When** the daily list is shown
**Then** the delta cell is `+0.5`
**And** it sits in the `delta` column, right of `trend`

### S2 — Negative delta

**Given** a day with weight 79.5 kg and trend 80.0 kg, `display_unit = "kg"`
**When** the daily list is shown
**Then** the delta cell is `-0.5`

### S3 — Zero (exact)

**Given** a day with weight 80.0 kg and trend 80.0 kg
**When** the daily list is shown
**Then** the delta cell is `0.0`
**And** it does not contain `+0.0` or `-0.0`

### S3a — Rounded zero (O3)

**Given** a day whose display-unit difference rounds to 0.0 (for
example 80.04 kg vs 80.00 kg in kg)
**When** the daily list is shown
**Then** the delta cell is `0.0`
**And** it does not contain `+` or `-`

### S4 — No weight, carried trend

**Given** a month-sheet day with no weigh-in and a carried trend
**When** the month sheet is shown
**Then** the delta cell is empty
**And** the trend cell still shows the carry

### S5 — First weigh-in in kg is 0.0 (O1)

**Given** the first weigh-in in the series at 80.0 kg
**When** the daily list is shown
**Then** the delta cell is `0.0`

### S5b — First weigh-in in lb is displayed subtraction

**Given** the first weigh-in 176.5 lb (`display_unit = "lb"`)
**When** the daily list is shown
**Then** the weight cell is `176.5`, the trend cell is `176.6`, and
the delta cell is `-0.1`

### S6 — Displayed cells add up (O2)

**Given** `display_unit = "lb"`
**And** weight 80.0 kg (list shows `176.4`) and trend 79.9 kg (list
shows `176.1`)
**When** the list is shown
**Then** the delta cell is `+0.3`
**And** it is not `+0.2` (the kilogram-residual conversion)

### S7 — Month sheet shows the same number

**Given** S1's day
**When** the month sheet for that month is shown
**Then** that row's delta cell is `+0.5`

### S8 — Tab skips delta

**Given** the month sheet, cursor on weight
**When** Tab is pressed (not editing, or after accepting a cell)
**Then** the cursor is on sleep, not on delta

### S9 — `NO_COLOR` keeps the sign

**Given** S1 and `NO_COLOR` set
**When** the list is shown
**Then** extractable text includes `+0.5`
**And** the cell has no chromatic ANSI

### S10 — Header alignment

**Given** the daily list
**When** the header and a value row are compared (ANSI stripped)
**Then** the `delta` header and the delta value share a column end,
as weight and trend already do

### S11 — Custom `delta-pos` is painted (O4)

**Given** `[colors] delta-pos = "magenta"` and S1
**And** `NO_COLOR` is unset
**When** the list is shown
**Then** the `+0.5` cell uses the magenta role, not default yellow

### S12 — Invalid `delta-pos` is dropped (O4)

**Given** `[colors] delta-pos = "chartreuse"` and S1
**When** the TUI starts
**Then** it opens
**And** the `+0.5` cell uses default yellow

## Functional requirements

-   **FR1.** List and month sheet include a `delta` column immediately
    right of `trend`.
-   **FR2.** When both weight and trend exist, the cell is the signed
    one-decimal display of `FromKG(weight) − FromKG(trend)` after
    rounding (Decisions 1 and 3).
-   **FR3.** If either operand is missing, the cell is empty. The
    first weigh-in is not missing a trend.
-   **FR4.** Delta is not editable and is not a SQLite column.
-   **FR5.** `[colors]` `delta-pos` / `delta-neg` / `delta-zero` are
    known roles, fail-open; defaults yellow / green / white.
    `NO_COLOR` strips chroma only.
-   **FR6.** Sign is in the text (`+`, `-`, or `0.0`).
-   **FR7.** List reverse includes the delta cell. Month-sheet delta
    keeps role colour.

## Documentation (O5)

-   `docs/how-to/colors.md` lists the three keys.
-   `docs/reference/config.md` Color roles table includes them and
    their defaults.
-   `docs/explanation/color.md` notes delta as a coloured data role
    (sign still in the text).

## Out of scope

-   Charts and chart PDFs
-   Filled monthly log-sheet PDFs
-   Editing the delta
-   Extra logged fields
