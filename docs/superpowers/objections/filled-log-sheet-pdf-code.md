---
spec: docs/superpowers/specs/2026-10-17-filled-log-sheet-pdf.md
date: 2026-10-28
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: "The 8 pt header word weekday is wider than its 28 pt column, so the last letters cross the column rule and sit on weight, while the header test still passes because it classifies a cell by where the baseline starts."
    evidence: "internal/logpdf/logpdf.go sets bodyPt = 8.0, cellPad = 2.0, and colWidth weekday 28, then SetFont(\"Helvetica\", \"\", bodyPt) and pdf.Text(colX(i), headerY, header) for \"weekday\" before the vertical pdf.Line rules. colX(1) is 66 and the next rule is at 92. Core Helvetica Cw (font_embed/helvetica.json, the widths GetStringWidth sums) gives weekday 3946/1000 em, 31.57 pt at 8 pt, ending at x 97.57. logpdf_test.go columnOf treats the column as the band the baseline starts in (x >= left-1 && x < right-0.5); cmd/hdtools/main_test.go copies that helper."
    disposition: accepted
    disposition_rationale: >
      Accepted. `weekday` at Helvetica 8 is 31.57 pt in a 28 pt
      column, so the ink crosses the rule into `weight`. The
      header word stays `weekday`. Weekday becomes 36 pt and
      weight becomes 40 pt. The word ends at 97.57 and the rule
      moves to 100. `176.4` still fits the weight column. The
      note column stays 220 pt. The plan's widths and both column
      scanners follow the writer, and a header's drawn end must
      fall before the next rule.
---

## O1 — implementation — high

### Claim

The 8 pt header word `weekday` is wider than its 28 pt column, so
the last letters cross the column rule and sit on `weight`. The
export still exits 0, and the header test still passes, because a
cell is whatever band the baseline starts in, not where the ink
ends.

### Evidence

The header is body type, not the 12 pt title, and the weekday
column is the second 28 pt band. Text is inset by `cellPad` on the
left only. The rules are stroked after the words, so the stroke
paints over anything that crosses the boundary.

```go
const (
	margin    = 36.0
	bodyPt    = 8.0
	cellPad   = 2.0
)
var colWidth = [...]float64{28, 28, 48, 48, 44, 36, 52, 36, 220}
```

```go
pdf.SetFont("Helvetica", "", bodyPt)
headers := []string{"day", "weekday", "weight", "trend", "delta", "sleep", "steps", "workout", "note"}
for i, header := range headers {
	pdf.Text(colX(i), headerY, header)
}
```

```go
x := margin
pdf.Line(x, top, x, bot)
for _, w := range colWidth {
	x += w
	pdf.Line(x, top, x, bot)
}
```

`colX(1)` is `36 + 28 + 2` = 66. The rule between weekday and
weight is at 92. The word has 26 pt before that rule.

`SetFont("Helvetica")` loads core Helvetica. For a non-UTF-8 core
font, `GetStringWidth` sums `Cw` from
`font_embed/helvetica.json`. Those entries are w 722, e 556,
e 556, k 500, d 556, a 556, y 500. The sum is 3946. At 8 pt that
is 31.57 pt. The word ends at x = 97.57.

The rule at 92 cuts `a` (`a` runs from about 89.1 to 93.6). `y`
starts at about 93.6, entirely past the rule. `weight` starts at
94, so `y` covers the first 3.6 pt of that `w`. The same `Cw` sum
for the other eight headers ends before each of their rules.
`workout` is the next tightest, about 28.0 pt of ink in a 36 pt
column. This is one word, not the whole grid.

Notes go through `GetStringWidth` and `noteInner`. The header does
not. Nothing in `paint` asks whether `weekday` fits.

The test that would have to catch this classifies a cell by the
origin of the baseline:

```go
// columnOf maps a text baseline onto the sheet geometry: 36 pt margin,
// then day 28, weekday 28, weight 48, ...
// A blank cell draws nothing, so the
// column is which band the baseline sits in.
if x >= left-1 && x < right-0.5 {
	return c.name, true
}
```

`weekday` starts at 66, inside the weekday band, so `wantHeader`
passes. `cmd/hdtools/main_test.go` copies the same helper. A run
that only extracts the strings `weekday` and `weight` cannot see
the overlap.

### Why this matters

Every successful page paints this header, including an empty month.
The vertical rule slices `a`, and `y` is in the weight cell on top
of `w`. US1 is a sheet someone prints. The acceptance check reads
the text back and sorts baselines, which is how a layout can be
green and still wrong on paper. Day 31 was given an explicit
inside-the-box check because extraction of `31` was not enough.
The header has the same hole in the horizontal direction, and no
scenario measures a string against its rule.

This is the assumption the widths encode and never check: that
8 pt Helvetica `weekday` fits in 28 pt. Notes are measured. The
header is not.

- **accept-as-stated** — the overlap is acceptable. Write that
  down, so 28 pt is not inherited as a fit.
- **revise-spec** — the word or the width is wrong for this font
  size. The column table or the header label changes.
- **add-test** — each header string's drawn end must fall before
  the next rule. The fit becomes a requirement.
- **consciously-carry** — ship the collision on purpose, because
  the column table stays and the two words can still be told
  apart.

## Explicitly not objecting to

- **`ErrNoTrendStart`:** Decision 5 says it is not an export
  failure. `trendFromFirstWeighIn` starts `ApplyTrend` at the
  first non-nil weight, and the swallow branch matches that. This
  does not reopen spec-mode O2.
- **Day 31 off the page:** The day-31 baseline is at API y 624
  (`headerY` 66 plus 31 times 18). In the content stream that is
  y 168, inside the 792 pt box, and the January test checks the
  box rather than the literal `31` alone. Spec-mode O1 is met.
- **Drop, then measure:** `encodeNote` runs before `fitNote`. A
  snowman-only note stays an empty cell and the page is still
  written. That is accepted O3 and rejected O5, not a new hole.
- **Both flags and a bad month:** `run` returns before
  `config.Ensure` and before `Open` when both export flags are
  set, and a bad `YYYY-MM` is rejected in that same stretch.
  Spec-mode O4 is implemented.
- **Bare `return err`:** The helvetica, paint-day, weight, trend,
  and trend-suffix returns in the tree are wrapped with `%w`.
  There is no remaining bare return on those paths to object to.
- **A reversed pair of valid dates:** `store.All` is `ORDER BY day`,
  and `time.Parse` of `2006-01-02` only accepts ISO dates, which
  sort chronologically. `rejectUnsorted` then sees an ordered
  series. The sheet is not walked backwards. S34's non-zero exit
  for insertion order would reject a page that is already in
  order.
- **The other eight headers, the note column, and the spec's
  sample numbers:** The same `Cw` sum fits `day`, `weight`,
  `trend`, `delta`, `sleep`, `steps`, `workout`, and `note`.
  `fitNote` keeps a note inside `noteInner` (216 pt), leaving the
  2 pt right pad. `100000`, `176.4`, and `+0.3` fit their columns.
  The miss is the header word `weekday`, not the grid.
- **Heading length, unit wiring, and the 0600 install:** A 12 pt
  `September 1990 (st)` heading is on the order of 110 pt in a
  540 pt content width. `exportLogPDF` passes `cfg.DisplayUnit`
  through. `installOwnerOnly` still chmods a same-directory temp
  and renames it. None of those is the collision above.
- **Chart export, `ApplyTrend`'s formula, and a TUI key:** Those
  were left alone on purpose. This record does not reopen them.
