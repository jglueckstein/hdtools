---
spec: docs/superpowers/specs/2026-10-02-high-resolution-charts.md
date: 2026-10-02
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: premise
    severity: high
    claim: "The backlog problem is visual parity with the book's Excel charts and, where the terminal allows, Hacker's Diet Online, but the enforceable contract is only a minimum subdivision of the existing character grid."
    evidence: "Intro: 'should look at least as good as the book's Excel charts. Hacker's Diet Online is a higher bar; match it where the terminal allows.' Decision 3 requires 'at least 16 vertical paint positions' and 'at least two' per character row. Out of scope: 'Choosing or naming the glyph set (the plan does that).' idea.md: 'look at least as good as the book's Excel charts' and 'match it where the terminal allows.'"
    disposition: accepted
    disposition_rationale: "The contract is the Braille floor. Excel and Hacker's Diet Online stay the reason in the intro, not a second test. Kitty and other image protocols are out of scope."
  - id: O2
    category: scope
    severity: high
    claim: "The amends list does not cover the monthly and long-term rules that still give the whole character cell to one series, so those rules still forbid the same-row stem and a second long-term line inside that cell."
    evidence: "Amends monthly Decision 7 and FR14, long-term Decision 6, and floats-and-sinkers Decision 2 only. Monthly Decision 4: 'When both series fall in the same cell, the daily mark wins' and 'not by putting two glyphs in one cell.' Monthly S3: 'that cell shows the daily mark.' Long-term Decision 5 and S5: 'on a shared cell the trend glyph is shown.' This spec's S4 is titled 'A stem survives inside one coarse row.'"
    disposition: accepted
    disposition_rationale: "Monthly Decision 4, S3, Observing, FR3, and FR5, and long-term Decision 5, S5, Observing, and FR4, now point here. A shared dot is what one series wins. One Braille rune may hold more than one series."
  - id: O3
    category: implementation
    severity: high
    claim: "The minimum grid puts S4's stem in one character cell, so one glyph, while FR5 also requires the mark, the stem, and the trend to stay distinct glyphs under NO_COLOR."
    evidence: "Decision 1: a glyph 'maps to paint positions inside its cell.' Decision 3: the body is 'at least 8 character rows' and 'each of those rows holds at least two vertical paint positions.' S4: 'A stem survives inside one coarse row' and 'a stem joins those two paint positions.' FR5: 'the monthly mark, stem, and trend path stay distinct by glyph.' Decision 4: 'Mark, stem, and trend stay distinct glyphs under NO_COLOR.'"
    disposition: accepted
    disposition_rationale: "A Braille cell is one rune and one color. Mark, stem, and trend are told apart across the chart. FR5 does not require three runes in the cell that S4 shares."
  - id: O4
    category: alternatives
    severity: medium
    claim: "A taller one-glyph plot with an explicit quantizer would separate the worked examples without a second coordinate system, and the spec does not weigh it against intra-cell paint."
    evidence: "Decision 3 keeps 'at least 8 character rows' and moves the new resolution inside the row ('at least two vertical paint positions' per row). Decision 1 refuses to name glyphs. Monthly Observing already locates 'a day's mark or path glyph in that column of those rows.'"
    disposition: rejected
    disposition_rationale: "idea.md asks for high-resolution glyphs inside the existing columns. Braille is that alphabet. A taller one-glyph plot does not give two horizontal positions per day."
  - id: O5
    category: specification quality
    severity: high
    claim: "Paint positions are not defined as something View shows, and no scenario requires the second horizontal position in a day slot to change the picture."
    evidence: "'Tests drive App.Update / View'. 'Paint positions are the positions from Decision 1, not a count of terminal columns.' Decision 1: 'The spec does not name the glyphs. The plan chooses the glyph set'. S1's Then is only that each slot 'owns at least two horizontal paint positions'. Out of scope: 'Choosing or naming the glyph set (the plan does that).'"
    disposition: accepted
    disposition_rationale: "The glyph is Braille U+2800–U+28FF with the standard dot bits. Tests decode those dots from View. S1 requires both horizontal columns of each November day to be lit."
  - id: O6
    category: specification quality
    severity: high
    claim: "Sixteen vertical positions, separation by one sixteenth of the span, and S4's claim that an eight-position grid drops the stem are not the same grid, and S2 does not force one."
    evidence: "FR2: 'The Y span has at least 16 vertical paint positions.' FR3: values 'differing by at least one sixteenth of the span, do not share a vertical paint position. S2 is the worked example.' S4: they 'fall in the same eighth of the Y span (min 80.00 kg, max 82.00 kg)' and 'On a grid of 8 positions those day-2 endpoints would share one position and the stem would be omitted.' Decision 3 defines that span as '(min − P) through (max + P)'."
    disposition: accepted
    disposition_rationale: "One quantizer: half-open equal bins of the padded span, four per character row. S2 names bins 12 and 14 at 8 rows. S4's coarse grid is eight equal bins of that same span."
---

## O1 — premise — high

### Claim

The backlog problem is visual parity with the book's Excel charts and,
where the terminal allows, Hacker's Diet Online, but the enforceable
contract is only a minimum subdivision of the existing character grid.

### Evidence

The intro states the goal in the same words as `idea.md`:

> On-screen monthly and long-term charts should look at least as good
> as the book's Excel charts. Hacker's Diet Online is a higher bar;
> match it where the terminal allows. Curves and stems are not one
> character per day in a handful of rows.

`idea.md` adds that this match is "a higher bar the book does not
cover" and that paint "may use high-resolution glyphs." What the
decisions then require is a floor: Decision 3's "at least 16 vertical
paint positions" and "at least two" vertical positions in each of the
existing "at least 8 character rows," plus two horizontal positions
inside the existing day column. Out of scope: "Choosing or naming the
glyph set (the plan does that)." No acceptance scenario compares the
picture with Excel or with Hacker's Diet Online. S1 through S10 check
counts, ownership, clip, empty copy, and glyph distinctness.

### Why this matters

A plan can take the floor literally — two addressable positions each
way, any glyphs that a unit test can tell apart — and every scenario
passes while the chart is still an 8-row terminal plot. The sentence
that justified the slice is then unchecked. "At least" allows a finer
result, but nothing sends the plan there, and "where the terminal
allows" never says which terminal capability is in or out. The human
has to decide whether this floor *is* the product goal before the plan
treats it as one. If it is, the intro is the wrong bar to judge the
plan against. If it is not, the scenarios are the wrong contract.

## O2 — scope — high

### Claim

The amends list does not cover the monthly and long-term rules that
still give the whole character cell to one series, so those rules
still forbid the same-row stem and a second long-term line inside
that cell.

### Evidence

This draft amends monthly Decision 7 and FR14 ("how the plot is
painted"), long-term Decision 6 ("a character column is a slot, not
the paint grid"), and floats-and-sinkers Decision 2 ("a shared cell is
a shared paint position"). It says those sentences stay the paint
until this slice ships, and that this draft replaces them.

It does not amend monthly Decision 4:

> When both series fall in the same cell, the **daily mark wins**
> (the book's diamond on the line). They remain distinct by glyph
> type across the chart, not by putting two glyphs in one cell.

Nor monthly S3 ("that cell shows the daily mark"), monthly Observing
("When both occupy one cell, the cell contains the daily mark"),
long-term Decision 5 ("When both series fall in the same cell, the
**trend wins**"), long-term S5 ("on a shared cell the trend glyph is
shown"), or long-term FR4 ("trend wins a shared cell").
Floats-and-sinkers Decision 2 retargets the word "cell" only "in this
decision," and only once this slice ships.

S4 is titled "A stem survives inside one coarse row" and requires the
stem to be drawn in the case an 8-position grid would have omitted.
Decision 3's minimum is still 8 character rows with two vertical
positions each, so that coarse row is one character cell.

### Why this matters

At the minimum this spec asks for, S4's mark and stem occupy one
character cell. The unamended monthly rules say that cell contains
the daily mark and not a second glyph. The unamended long-term rules
say a shared cell contains only the trend, so two lines that fall in
one row never show the gap the extra vertical positions were for.
Floats-and-sinkers Decision 2 does not reach those sentences. One
implementer follows this draft and draws the intra-cell stem; another
follows monthly S3 and long-term S5, which this file never replaces,
and the stem or the daily path disappears. This repo's amendment
habit is an explicit "this replaces …" list (the Y-range spec does
that for monthly Decision 6 / FR15 / S16). Silence here leaves both
contracts in force.

## O3 — implementation — high

### Claim

The minimum grid puts S4's stem in one character cell, so one glyph,
while FR5 also requires the mark, the stem, and the trend to stay
distinct glyphs under `NO_COLOR`.

### Evidence

Decision 1: the plan's glyph "maps to paint positions inside its
cell." A cell is one rune. Decision 3: "at least 8 character rows"
and "each of those rows holds at least two vertical paint positions."
S4's title is "A stem survives inside one coarse row," and its Then
says "a stem joins those two paint positions" when the endpoints
share an eighth. Decision 4: "Mark, stem, and trend stay distinct
glyphs under `NO_COLOR`." FR5: "the monthly mark, stem, and trend
path stay distinct by glyph." S5 checks that distinctness on the S2
chart, not on the same-row stem.

### Why this matters

On an 8-by-2 grid, one eighth is one character row, so S4's two
endpoints and the stem between them are one column of one row: one
code point. That code point cannot be both "the mark glyph" and "the
stem glyph." Two readings are open, and they build different charts.
If "distinct" means three pure code points each appear somewhere, a
stem that exists only as part of the mark's rune never satisfies FR5
on an S4-like month, while S5 still passes because S2's mark and
trend fall in different eighths and can occupy different rows. If
"distinct" means each painted cell is one series, S4's cell is
illegal. Monthly Decision 4 already refused an unnamed "both" glyph
("not by putting two glyphs in one cell") and resolved collision by
giving the cell to the mark. This slice puts the collision back
inside the cell and does not name the composed rune. Executing
Decision 1, S4, and FR5 together has no result that meets all three.

## O4 — alternatives — medium

### Claim

A taller one-glyph plot with an explicit quantizer would separate the
worked examples without a second coordinate system, and the spec does
not weigh it against intra-cell paint.

### Evidence

Decision 3 keeps the plot body "at least 8 character rows" and puts
the new resolution inside the row ("at least two vertical paint
positions"). Decision 1 then needs a glyph map the spec will not
name. The monthly spec's Observing section already tells a test how
to see a mark: locate "a day's mark or path glyph in that column of
those rows." Nothing in this draft compares that existing oracle
with an intra-cell coordinate system.

### Why this matters

The collisions in O2 and O3 come from packing several series into one
rune, not from separating 80.55 kg and 80.85 kg. Sixteen character
rows of the glyphs the monthly spec already allows (`o`, stem, `-` /
`/` / `\`), with a stated map from value to row, would give S2
different rows and would give S4's 80.50 kg mark and 80.1 kg trend a
stem of ordinary cells between them. Tests stay ANSI-stripped column
checks. `NO_COLOR` distinctness stays "different glyphs across the
chart," which is the rule already shipped. The cost is vertical
space, which is why someone might reject it. The draft never puts
that cost next to the cost of an unnamed sub-cell alphabet, so the
heavier mechanism is the only one on the page.

## O5 — specification quality — high

### Claim

Paint positions are not defined as something `View` shows, and no
scenario requires the second horizontal position in a day slot to
change the picture.

### Evidence

The scenarios open with:

> Tests drive `App.Update` / `View` and must not require a TTY.
> Today may be frozen. Paint positions are the positions from
> Decision 1, not a count of terminal columns.

Decision 1: "The spec does not name the glyphs. The plan chooses the
glyph set and how each glyph maps to paint positions inside its
cell." Out of scope repeats "Choosing or naming the glyph set (the
plan does that)." S1's Then is that each day slot "owns at least two
horizontal paint positions" and that "no paint position belongs to
two days." S6 and S7 repeat "has at least two horizontal paint
positions." Decision 4 uses one of them: "The stem uses the mark's
horizontal paint position." No scenario places two samples on the
two horizontal positions of one slot and requires the view to differ.
The monthly and long-term specs both have an "Observing the chart"
section. This one does not, and it does not replace theirs.

### Why this matters

`View` is a string of character cells. Two paint positions inside a
cell are one rune unless a named encoding says which rune is which.
That encoding is out of scope, so a test written from this spec
cannot decide pass or fail for S2 or S4 without inventing it. The
"mapping that cannot tell the positions in the scenarios apart"
sentence does not repair the horizontal scenarios: they assert
ownership, not a visible difference. An implementation can keep
today's one glyph per day column, attach two x-coordinates to the
column, and satisfy S1, S6, and S7. Vertical separation has the same
hole if the test is allowed to read a coordinate that `View` does not
show. The plan then becomes the test oracle, which is the artefact
this spec is supposed to constrain.

## O6 — specification quality — high

### Claim

Sixteen vertical positions, separation by one sixteenth of the span,
and S4's claim that an eight-position grid drops the stem are not the
same grid, and S2 does not force one.

### Evidence

Decision 3 and FR2: the Y span "(min − *P*) through (max + *P*)" "has
at least 16 vertical paint positions." FR3: two plotted values "in
the same eighth of the Y span," "differing by at least one sixteenth
of the span," "do not share a vertical paint position. S2 is the
worked example." S2's Then does not mention an eighth or a
sixteenth. The prose after S2 does, for 80.55 kg and 80.85 kg on a
data range of 80.00 kg to 82.00 kg. S4's Then says the day-2 mark and
the 80.1 kg trend "fall in the same eighth of the Y span (min 80.00
kg, max 82.00 kg) and differ by 0.40 kg," and then:

> On a grid of 8 positions those day-2 endpoints would share one
> position and the stem would be omitted. This slice still draws it.

### Why this matters

*P* is 2 × 0.45359237 kg ≈ 0.907 kg, so the Decision 3 span for these
fixtures is about 79.093 kg through 82.907 kg, width about 3.814 kg.
An eighth of that is about 0.477 kg; a sixteenth is about 0.238 kg.

S4's 80.50 kg and 80.1 kg differ by 0.40 kg. Measured from the padded
span they fall in the same equal eighth (both in the third of eight).
They do not share an endpoint-inclusive 8-tick grid: with *t* the
fraction of the padded span, `round(t × 7)` puts them on ticks 3 and
2, so a coarse chart that includes Ymin and Ymax already has two rows
and already draws a stem. S4's "would share … and the stem would be
omitted" is true for eight equal bins and false for eight inclusive
ticks. The parenthesis "(min 80.00 kg, max 82.00 kg)" names the data
ends, not the span. Read as the span itself, width 2 kg, 80.50 kg and
80.1 kg are not in the same eighth at all (0.50 kg is two eighths
above 80; 0.10 kg is inside the first).

FR3 is a third grid. Sixteen positions from Ymin through Ymax
inclusive are a fifteenth of the span apart, and a fifteenth is wider
than a sixteenth. Values a sixteenth apart can round to the same tick
and still sit in one eighth (for example fractions 0.435 and 0.4975
of the span both round to tick 7 of 0..15, and both lie in
[0.375, 0.5)). Equal bins of width one sixteenth do satisfy FR3.
S2 cannot settle which was meant: 80.55 kg and 80.85 kg are about
0.079 of the padded span, which is more than a fifteenth, so they
separate on the inclusive 16-tick grid that already violates FR3.
S3 only counts positions and rows. One implementation passes every
scenario with inclusive ticks and still merges pairs FR3 forbids.
Another bins the span into sixteenths and draws S4's stem only
because of that binning. The draft treats these as one rule.

## Explicitly not objecting to

- **PDF left unchanged**: `2026-09-12-pdf-monthly-charts.md` already
  requires vector geometry and rejects a Braille or block dump, and
  FR9 matches that split rather than drifting from it.
- **S6 and S7 day counts**: 1 September through 10 November is 71 days
  and through 30 November is 91 days, and both scenarios apply
  *D* ≤ *W* with *W* = 72 as the long-term spec states it.
- **The stated trend figures**: with no logs before November, 10%
  smoothing rounded to 0.1 kg, and blank days carrying the previous
  trend, S2's trends are 80.0, 80.1, 80.2, and 80.4, and S4's day-2
  trend is 80.1. The same-eighth claim for S2's two marks also holds
  on the padded span (both in the fourth of eight equal parts; 0.30 kg
  is more than one sixteenth of about 3.814 kg).
- **Identity constraints**: Y margin, today-clip, Loss, Daily Deficit,
  `[colors]`, the 16-color floor, keys, and the existing empty
  predicate are restated in Decision 6 and FR7, not quietly replaced.
  Carry-only versus empty stays where the monthly spec already put it.
- **Long-term X labels**: this spec follows the x-label spec and does
  not add terminal columns, so the six-column and three-column overlap
  rules still measure the same grid.
- **Graphics protocols as a way to "match Hacker's Diet Online"**:
  Decision 6 keeps the 16-color floor and `NO_COLOR`, which an image
  protocol would drop. That is not an open technique hiding in the
  intro.
- **A vertical stem at the mark's x**: Decision 4 can be read as
  placing that day's trend sample on the mark's horizontal position,
  so "joins those paint positions" and "uses the mark's horizontal
  paint position" describe one vertical segment. That reading does
  not by itself miss the trend.
