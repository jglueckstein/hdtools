---
spec: docs/superpowers/specs/2026-10-08-monthly-online-chart.md
date: 2026-10-09
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: >
      A vertical stem is DrawLinePoints half-tips, so the bar slits at
      the trend and a one-row gap does not join the diamond to the
      trend at all.
    evidence: >
      internal/tui/onlinechart.go paintStems draws
      graph.DrawLinePoints on graph.GetLinePoints of the two cells in
      column c0. For a vertical run that maps to Down-only and
      Up-only ends, ntcharts emits U+2577 and U+2575, and its own
      candlestick notes say those tips leave a half-cell slit unless
      merged to U+2502. paintDiamonds then replaces the weight cell.
      stemBetween only accepts a rune strictly between the diamond
      row and a Braille row, which is the middle bar on a long gap
      and nothing on a one-row gap.
    disposition: accepted
    disposition_rationale: >
      Accepted. Decision 4 and FR3 require a stem when that day's
      weight and trend land on different rows. The library line
      leaves half-tips, and the diamond then covers the weight
      cell, so the trend cell keeps a half-tip and a one-row gap
      keeps no bar. paintStems sets every cell of that segment to
      U+2502 before the diamond is drawn. A test places 80 kg and
      80.1 kg on adjacent rows and requires U+2502 in the
      diamond's column on the trend row.
  - id: O2
    category: risk
    severity: medium
    claim: >
      The open diamond is one canvas cell, but U+25C7 is East Asian
      Ambiguous, so a 67-column row and the day labels hold only when
      that glyph is one column wide.
    evidence: >
      internal/tui/onlinechart.go sets diamondRune to U+25C7 and
      writes that rune once per column. Day labels are one rune per
      plot column. charm.land/lipgloss Width goes through ansi, and
      ansi leaves EastAsianWidth false unless RUNEWIDTH_EASTASIAN is
      true, in which case an Ambiguous rune is two columns.
      onlineScaleLine treats a row as 67 columns under that default.
    disposition: accepted
    disposition_rationale: >
      Accepted. The open diamond stays U+25C7. That rune is East
      Asian Ambiguous, and lipgloss.Width counts it as two
      columns only when RUNEWIDTH_EASTASIAN is true. This program
      does not set that variable. The 67-column row is the
      default width. The glyph stays, and this slice does not
      special-case the flag.
---

## O1 — implementation — high

### Claim

A vertical stem is `DrawLinePoints` half-tips, so the bar slits at the
trend and a one-row gap does not join the diamond to the trend at all.

### Evidence

The painter draws the stem with the two-point line, then paints the
diamond over the weight cell:

```go
paintTrend(&m, samples, share, lo, hi, p.trend)
paintStems(&m, samples, share, lo, hi)
paintDiamonds(&m, samples, share, lo, hi, p.weight)
```

`internal/tui/onlinechart.go` lines 89–91. The stem itself is only
that library line, in the diamond's column:

```go
// GetLinePoints plus DrawLinePoints is the library's line
// between two points. A vertical helper would also paint past the
// segment. The same row means no stem, including equal values.
from := canvas.Point{X: c0, Y: weightRow}
to := canvas.Point{X: c0, Y: trendRow}
graph.DrawLinePoints(m, graph.GetLinePoints(from, to), runes.ThinLineStyle, style)
```

`internal/tui/onlinechart.go` lines 295–316. `GetLinePoints` on a
vertical pair emits one point per row. `DrawLinePoints` then marks the
upper cell Down-only and the lower cell Up-only, and the cells in
between both (`canvas/graph/graph.go` lines 331–335). Thin style turns
those into U+2577 (`╷`), U+2502 (`│`), and U+2575 (`╵`)
(`canvas/runes/runes.go` lines 403–411).

The same library treats an unmerged tip as a slit against the bar it
is supposed to continue. The candlestick path says so and upgrades the
tip to a full vertical:

```go
// The line style merges this tip's ╷ with the body's upward segment
// into │ ... so the wick doesn't leave a slit against the body.
lr = runes.LineVertical // merged ╷ → │
```

and the top wick merges `╵` into `│` for the same reason
(`canvas/graph/graph.go` lines 1047–1054 and 1090–1097). `paintStems`
never merges. On a gap of three or more rows the middle is `│`, and
the cell that remains on the trend row is the unmerged tip. On a gap
of one row there is no `│` at all: the two cells are the two tips, and
the diamond replaces the weight tip, so the only mark left is a
half-tip on the trend cell.

The trend glyph that tip is supposed to meet is the top dot of that
same character cell, not a full-height bar:

```go
// One dot is enough for the rune. The top of the character cell
// is the row the tick scale uses.
grid.Set(canvas.Point{X: col * 2, Y: row * 4})
```

`internal/tui/onlinechart.go` lines 290–292. The other columns of the
day keep that dot. The tip on the diamond's column does not fill the
cell up to it.

Decision 4 and FR3 require a stem whenever the weight and the trend
land on different rows. S3's 18 kg and 9 kg gaps do have rows between,
and the tests only ask for a non-Braille rune strictly between the
diamond row and a Braille row:

```go
for _, s := range sRows {
    if s > loR && s < hiR {
        return true
    }
}
```

`internal/tui/onlinechart_test.go` lines 294–297. A middle `│`
satisfies that. The tip on the trend row does not, because that row
is the Braille row. A one-row gap has no such rune, and no test builds
one. `isStemRune` is any other rune (`onlinechart_test.go` lines
102–107), so a disconnected cap still counts.

### Why this matters

US1 is the join. A long stem still reads as a green column, but it
stops a half cell short of the trend on every gap the scenarios
actually draw, float and sinker alike. The case Decision 4 still
requires and S3 does not illustrate — different rows, nothing between
them — does not join at all, and the green suite cannot see it. That
is the boundary just above "same row, no stem," which is where a
quiet month lives.

## O2 — risk — medium

### Claim

The open diamond is one canvas cell, but U+25C7 is East Asian
Ambiguous, so a 67-column row and the day labels hold only when that
glyph is one column wide.

### Evidence

The glyph and the write are one cell per column:

```go
// Open diamond, not the filled U+25C6 the website uses for a flag.
diamondRune = '\u25c7'
```

```go
m.SetRuneWithStyle(canvas.Point{X: c0, Y: row}, diamondRune, style)
```

`internal/tui/onlinechart.go` lines 45–46 and 339. Each plot row is
the five-column gutter plus one rendered rune per canvas column
(lines 343–355). The day-label line is a buffer of `monthCols` runes,
spaces and ASCII digits, not the diamond (lines 378–395). Nothing
asks what `lipgloss.Width` is for U+25C7.

`lipgloss.Width` is `ansi.StringWidth`. That package's width condition
starts with East Asian width off, and turns it on only for the
environment variable:

```go
var wcOptions = &runewidth.Condition{
    EastAsianWidth:     false,
```

```go
if ea, err := strconv.ParseBool(os.Getenv("RUNEWIDTH_EASTASIAN")); err == nil && ea {
    wcOptions.EastAsianWidth = true
```

`github.com/charmbracelet/x/ansi@v0.11.8/method.go` lines 11–12 and
21–22. U+25C7 is Ambiguous (U+25C6..U+25C7). With the flag on, the
rune is two columns. One weighed day makes that plot row 68 visible
columns, and two diamonds on one row make it wider. The day-label
line stays 67 ASCII columns, so "10" is no longer under day 10.

The tests lock the narrow reading. `columnCells` repeats a rune once
per `lipgloss.Width`, and `onlineScaleLine` accepts the row only when
that count is 67 (`internal/tui/onlinechart_test.go` lines 109–121
and 124–132). No test sets `RUNEWIDTH_EASTASIAN`. Under the default
the diamond is one column and the contract holds. The 67-column check
is that default, written down as the requirement.

### Why this matters

FR16 says every plot row is 67 visible columns, measured as visible
width because the plot contains colour, and the day numbers sit in
that day's columns. Colour is not the only thing that changes width.
A person who sets the flag so CJK text is two columns — the switch
this stack already honors — gets a monthly row that is no longer 67
and labels that have slipped off the diamonds. The failure is silent:
the program still draws, and the suite still passes.

- **accept-as-stated** — the diamond is one column. Write that down,
  including that `RUNEWIDTH_EASTASIAN` is off for this picture, so
  the next reader inherits a decision rather than a default.
- **revise-spec** — visible width has to stay 67 even when the glyph
  is wide. The spec changes, and the painter has to measure the rune.
- **add-test** — one-column width is a requirement nobody wrote
  beside the happy-path `lipgloss.Width` check. A test that sets the
  flag makes the contract real.
- **consciously-carry** — known, wrong when Ambiguous runes are wide,
  and shipped anyway, on the record with the because.

## Explicitly not objecting to

- **One Braille dot per character cell:** Decision 13 and the
  accepted spec-mode O2 fix the plot at 16 rows and treat "same row"
  as no stem. A 4×2 sub-cell trend would reopen that, not reveal a
  bug in this painter.
- **Bucketed long-term scale includes weigh-ins:** S16's 170 lb to
  175 lb series is a quarter longer than 72 columns, so the bucket
  keeps trend only. `longYRange(..., true)` is what still yields
  about 168 lb to 177 lb. The four windows are not redrawn as the
  Online picture.
- **Remainder columns on the earliest days:** The spec leaves that
  choice open. This split still keeps 10 November's empty tail at
  plot columns 23 through 62, which is the bound S9 checks.
- **80.4 kg to 82.4 kg ticks not glued to row 0 and row 15:** Those
  ticks sit inside 79.4 to 83.4. Pinning 83 and 80 to the canvas
  ends would move them off the values. S6 asks for the labels, not
  for that placement.
- **Older `idea.md` paragraphs and "not shipped":** Decision 14
  keeps those paragraphs as history. The new change record is the
  one that drops the box and the shared margin. Issue #78 and the
  PDF redraw stay out of this slice on purpose.
