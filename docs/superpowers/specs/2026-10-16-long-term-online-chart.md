# Long-term Online chart

**Date**: 2026-10-16
**Status**: approved
**Issue**: [#78](https://github.com/jglueckstein/hdtools/issues/78)
**Backlog**: [`idea.md`](../../../idea.md) (TUI)
**Slice**: S4 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)

The on-screen long-term chart becomes the Hacker's Diet Online
picture. The four windows stay: Quarterly, Semiannual, Annual, and
Complete. The series are a thin weight line and a Braille trend.
There are no diamonds and no stems. A plain heading names the
window. Weight ticks follow the monthly one-step rule. Month labels
run along the bottom. The caption stays the straight-line fit.

The monthly chart and the PDF stay as they are. This draft does not
revise the earlier chart specs. Those stay the history of the
picture this slice replaces.

## User stories

### US1 — See the window as the Online picture

As a person reading a long-term chart, I want a thin line through
the weighed days and a Braille trend, with no diamonds and no
stems, so the window is the Online picture rather than the
Braille-cell plot.

### US2 — Read the same scale as the monthly chart

As a person who shows weight in kilograms, pounds, or stone, I
want the monthly tick rule on this screen, so the long-term scale
is not a two-pound margin.

### US3 — Name the window and the months

As a person cycling the four windows, I want the heading to be the
window name and the months labeled under the plot, so I can tell
Quarterly from Complete without a title box.

### US4 — Keep the rate, the keys, and the other pictures

As a person who already reads the fit caption, the monthly chart,
and the PDF, I want those left as they are, so this change is only
the on-screen long-term picture.

## Decisions

1.  **Long-term screen only.** The four windows, their ranges, and
    the keys stay. The window still ends at the latest log, and at
    today when that log is in the current month. There is no custom
    date range. The monthly chart stays the Online month. The PDF
    stays the vector page.
2.  **The series are a thin weight line and a Braille trend.** The
    trend is the derived series already on the log, drawn as
    Braille, including carry onto days with no weight. Each weighed
    day is a point on one thin line, in the `weight` role. The line
    joins weighed days in order and is not a diamond and not a
    stem. A day with no weight adds no point. The trend is drawn
    first. The weight line is drawn after it.
3.  **Plain heading.** The heading is the window name
    (`Quarterly`, `Semiannual`, `Annual`, or `Complete`). It uses
    the `title` role, centered on the same 67-column line as the
    monthly heading. It has no border and no filled background.
    The date range is not in the heading. `NO_COLOR` keeps the
    words and the bold and drops the chromatic colour. The
    database line and the help line stay.
4.  **The monthly scale.** Plotted values are each day's weight,
    when it has one, and each day's trend, when it has one. The
    canvas, the one-step pad, the tick rule, the 16 rows, and the
    five-column gutter are the monthly ones. The `%5.1f-` gutter
    and the two-pound margin are gone. The plot is 62 columns. It
    does not follow the terminal.
5.  **Days share the 62 columns.** The days of the window, from
    its first day through its last day, take the 62 columns in
    order. A window of 62 days or fewer uses the monthly share
    rule. A longer window maps each day into those columns, so
    later days are not dropped and the weight line is not removed.
    Days after today are not in the window.
6.  **Month labels.** One line under the plot names a month as
    `Sep 90`, at that month's column. The first month and the last
    month are labeled when the last label fits. A label that would
    overlap is skipped. There is no second label line. An empty
    window has no month labels.
7.  **The caption stays the fit.** Loss and Daily Deficit stay the
    straight-line fit already on this screen. The numbers are not
    stored. An empty window and a window with fewer than two
    trend values omit the caption.
8.  **Colour stays on the existing roles.** The trend uses the
    `trend` role and stays bold. The weight line uses the `weight`
    role and stays not bold. `NO_COLOR` drops chromatic colour and
    keeps the Braille trend, the weight line, and the trend's
    bold.

## Published sentences

`docs/how-to/long-term-chart.md` describes the thin line, the
Braille trend, the plain window name, the one-step scale, and the
one-line month labels. `idea.md` records that this screen no
longer keeps the title box or the two-pound margin.
