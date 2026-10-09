---
spec: docs/superpowers/specs/2026-10-08-monthly-online-chart.md
date: 2026-10-09
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: scope
    severity: high
    claim: "The slice withdraws the month-year box and the shared ±2 lb margin and leaves the PDF on the old page, but it never says what happens to the how-tos or the idea.md paragraphs that still state all three."
    evidence: "The spec points at idea.md as the TUI backlog. Decision 5 removes the coloured Excel-style title box. Decision 6 replaces the fixed two-pound margin with one display-unit step. Decision 1 and FR12 keep the PDF on the vector page it has now. Out of scope defers the PDF redraw, the long-term picture, and the Online caption, and does not mention docs."
    disposition: accepted
    disposition_rationale: >
      Accepted. This slice names the sentences it makes false. The
      monthly how-to drops the month-year box and the ±2 lb margin,
      and describes the plain heading and the one-step scale. The
      long-term how-to keeps the ±2 lb margin and stops saying it
      matches the monthly chart. The PDF how-to keeps the existing
      vector page, including its title box, and stops saying the PDF
      is the on-screen chart. The idea.md change record says the
      monthly screen drops the box and the shared margin, the
      long-term screen keeps both until #78, and the PDF keeps the
      old page. Older change records stay the history of those
      slices.
  - id: O2
    category: specification quality
    severity: high
    claim: "Plot height, the smallest gap that must still show a stem, and when extra ticks are dropped are all unspecified, and the stem scenarios only separate a gap that is most of its own scale."
    evidence: "Decision 12 and FR16 fix every plot row at 67 visible columns and never set a row count. S3 requires a stem for 80 kg then 100 kg and for 80 kg then 70 kg. S7 requires ten stone labels, and each tick gutter is one four-column label plus a hyphen. S8 allows a tick between 39 and 121 to be omitted. FR5 requires every tick only for the S6 and S7 ranges."
    disposition: accepted
    disposition_rationale: >
      Accepted. The plot is 16 rows, fixed, and it does not follow
      the terminal. A tick never shares a row with another tick. S6
      and S7 have fewer ticks than rows, so every tick is shown. S8
      has more ticks than rows, so the end ticks stay and the middle
      ticks may be dropped. A stem is required when the weight and
      the trend land on different rows. S3's 18 kg and 9 kg gaps do.
      Two different values that land on the same row have no stem. A
      1 kg gap on a tall scale is not a separate guarantee.
  - id: O3
    category: specification quality
    severity: medium
    claim: "\"Existing title style\" does not choose between the palette title role and bare bold, and S5 checks neither colour, bold, nor centering."
    evidence: "Decision 5: the heading is the month name and the year, in the existing title style, with no border and no filled background, and the coloured Excel-style title box is gone. S5 checks the words November 1990, no border, no filled background, and that the heading does not say hdtools. Decision 8 assigns colour to trend, weight, and the green stem, not the heading."
    disposition: accepted
    disposition_rationale: >
      Accepted. The heading uses the existing title role: bold,
      default cyan, and a [colors] title entry changes it. It has no
      border and no filled background, and it is not the old
      yellow-on-blue box. It is centered on the 67-column plot line.
      NO_COLOR keeps the words and the bold. S5 checks the role, the
      bold, the centering, and the override.
---

## O1 — scope — high

### Claim

The slice withdraws the month-year box and the shared ±2 lb margin and
leaves the PDF on the old page, but it never says what happens to the
how-tos or the idea.md paragraphs that still state all three.

### Evidence

The spec's backlog pointer is `idea.md` (TUI). Decision 5: "The coloured
Excel-style title box is gone from this screen." Decision 6 replaces
"its fixed two-pound margin" with one step past the plotted min and
max. Decision 1 and FR12: the PDF "stays the vector chart it is,
including its title," and "This slice does not redraw that page."

Out of scope names the Online caption (#79), the long-term picture
(#78), and "Redrawing the PDF." It does not mention
`docs/how-to/monthly-chart.md`, `docs/how-to/monthly-chart-pdf.md`,
`docs/how-to/long-term-chart.md`, or the earlier normative paragraphs
in `idea.md`. The plan bars edits to `docs/reference` and `keys.md` and
never assigns the how-tos. Those reference pages are not where the
withdrawn claims are published.

`idea.md` still says every on-screen chart, monthly and all four
long-term kinds, uses "(min of plotted values) − 2 lb through (max) +
2 lb", that a text box above the plot has the Excel colours, and that
"PDF monthly charts are the on-screen monthly chart on paper" and
"should use the same title box and stems." The draft note at the
bottom says this spec replaces the Braille-cell picture. It does not
say the box or the shared margin is withdrawn.

`docs/how-to/monthly-chart.md` still says "A month-year box sits above
the plot" and "The vertical scale is the plotted min and max plus 2 lb
of margin." `docs/how-to/long-term-chart.md` says that same margin is
"as on the monthly chart." `docs/how-to/monthly-chart-pdf.md` opens
with "The monthly chart PDF is the on-screen monthly chart on paper:
title box, daily marks, stems, trend, Y margin."

### Why this matters

On the day this ships, the monthly screen no longer has the box or the
±2 lb margin, and the long-term screen still has the ±2 lb margin.
The long-term how-to's claim that the two scales match becomes false
even though this slice does not draw the long-term chart. The PDF
how-to's claim that the file is the on-screen chart becomes false
because the spec keeps the old vector page on purpose. The next reader
of `idea.md` still has a standing requirement to put the box and the
shared margin back. The fix is to say, in this spec, which of those
sentences are amended now and which stay until a later slice. That is
not a request to redraw the PDF or the long-term chart here.

## O2 — specification quality — high

### Claim

Plot height, the smallest gap that must still show a stem, and when
extra ticks are dropped are all unspecified, and the stem scenarios
only separate a gap that is most of its own scale.

### Evidence

Decision 12 and FR16 specify the horizontal contract in detail: a
62-column plot, a five-column gutter, a 67-column plot row, and no
resize with the terminal. Neither they nor the functional requirements
state a plot height in rows, Braille dots, or terminal lines.

S3 is the stem contract. For 80 kg then 100 kg, day 1's trend is 80
and day 2's trend is 82 (one tenth of the 20 kg gap, already a whole
tenth). The stem is 100 against 82, an 18 kg gap, on a scale from 79
to 101 (22 kg). That gap is about four-fifths of the scale. For 80 kg
then 70 kg, day 2's trend is 79, so the stem is 9 kg on a scale from
69 to 81 (12 kg), three-quarters of the scale. No scenario asks for a
stem when the gap is a kilogram or less inside a wider month. S1's
finished 80 kg to 82 kg month is the ordinary shape, and it does not
require a stem at all.

S7 requires ten labels from 11.9 through 12.6, including both copies
of 12.1 and of 12.4. The gutter that holds each of them is four
columns plus a hyphen (`12.6-` is the whole five-column gutter), so
two of those labels cannot share a row. Showing every S7 tick
therefore needs at least ten plot rows, unless the height rule is
something else. S6 likewise requires five kilogram labels, every one
of them. S8 runs from 40 kg to 120 kg, so the closed tick list is 39
through 121 (83 labels), and "a tick between those ends may be
omitted." FR5 repeats that split: every tick for the S6 and S7
ranges, middles optional only for the S8 range. Nothing says what a
15 kg month does, or what "crowded" means in rows.

The plan's only vertical instruction is "Pick a height on which an
18 kg gap and a 9 kg gap leave a visible stem mark between the diamond
and the trend." That is the S3 bar, not a smaller one.

### Why this matters

A height that puts one green cell inside an 18 kg gap on a 22 kg scale
passes S3 and can still put a 1 kg float on the same row as the trend.
Decision 4 then says there is no stem, because the two values landed
on one row. The chart's job, in US1, is that stem: the day's weight
against the trend. The fixtures never make the gap small relative to
the month, so they cannot catch a coarse canvas.

The same missing number leaves the tick rule with two implementations
that both pass. One grows the plot to one row per tick and is 83 rows
tall on S8, which the view-string tests accept because they do not use
a terminal. Another keeps a short fixed height, shows all ten S7
labels, and drops S8's middles. A third shows all 83 labels inside a
short canvas; S8 allows that, and the four-column gutter does not stop
two labels from being painted on top of each other. US2 is to read the
scale. None of the three is ruled out.

## O3 — specification quality — medium

### Claim

"Existing title style" does not choose between the palette title role
and bare bold, and S5 checks neither colour, bold, nor centering.

### Evidence

Decision 5 requires "the month name and the year (`November 1990`), in
the existing title style, with no border and no filled background."
The next sentence says the coloured Excel-style title box is gone.
S5 checks that the heading names November 1990, has no border and no
filled background, and does not say "hdtools", with and without
`NO_COLOR`. It does not check a foreground, bold, or centering.
Decision 8 names the `trend` role, the `weight` role, and green stem
chrome. It does not say the heading uses the `title` role, and it does
not say a `[colors]` title entry is ignored there.

The screen today does not use that role for the chart. The box is
hard-coded yellow on blue with a red border. The palette's `title`
style, used for other chrome, is bold cyan and follows `[colors]`.
The plan unpacks the phrase as "bold, no border, no background" and
still names no colour and no role. S4 is supposed to reuse this plain
heading.

### Why this matters

After the box is removed, two readings still pass S5. One paints the
palette title role, so a configured title colour changes the heading
and S4 inherits that. The other paints bare bold in the default
foreground, so a title entry does nothing and the heading does not
match the rest of the chrome. A third keeps the box's yellow
foreground once the background and border are stripped. The slice's
decision to drop the Excel box is clear. Which style replaces it is
not, and the scenario that accepts the heading cannot tell them apart.

## Explicitly not objecting to

- **The fixed 62-column width:** US5 and Decision 12 want every month
  on one width, with days after today left as empty slots. That is a
  product choice, not an accident. The horizontal slot math for a
  31-day month and the 10 November empty tail (columns 23 through 62)
  is checked.
- **Which days get a remainder column:** The spec leaves that unfixed
  on purpose, and S9's empty-tail bound holds whichever two days of a
  30-day month are wider. Same-day X for the stem is the ordinary
  reading of "the line from that day's weight to the trend on that
  same day," not a second layout.
- **Deferring the PDF page and the long-term picture:** US4, FR11,
  FR12, and slice S2 already separate those engagements. O1 is only
  that the published sentences equating them are left in place.
- **772 versus 2.2 × 3500 / 10:** S9 and FR10 require 772 calories for
  a 1.0 kg loss over 10 days. That is the unrounded pound conversion
  the pencil path already uses. The 2.2 lb figure is the one-decimal
  display of the same loss, and the 30-day case rounds to 257 either
  way.
- **Duplicate stone labels:** Decision 6 and S7 require both pound
  steps even when both format as 12.1 or as 12.4. That matches the
  one-decimal stone display already on the sheet.
- **Stem green with no new `[colors]` key:** US3 and the resolved
  section close this. Decision 4 and S13 keep the three shapes
  distinct with colour off, and an invalid colour still fails open.
- **A log that starts after day 1:** No fixture opens that month, but
  Decision 2 does not invent a trend, and Decision 7 already refuses
  a label on a day that is not plotted. The future empty day is the
  stated example of that rule.
- **Leaving the earlier chart specs approved:** The draft says they
  stay the history of the picture this slice replaces, not a second
  live contract.
- **An unpinned ntcharts patch:** The spec requires the v2 module and
  Go 1.27.0, and it refuses to freeze a patch number in the spec.
  The plan stops if the fetched release cannot draw on that Go
  version. The picture scenarios, not S20, are what lock the drawing.
- **The stem glyph:** Any mark that is not Braille and not U+25C7
  keeps the three shapes apart under `NO_COLOR`, which is the
  requirement. The plan's U+2502 is one such mark, not a second
  contract.
