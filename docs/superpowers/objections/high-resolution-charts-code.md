---
spec: docs/superpowers/specs/2026-10-02-high-resolution-charts.md
date: 2026-10-02
mode: code
diaboli_model: grok-4.7
objections:
  - id: O1
    category: implementation
    severity: high
    claim: >-
      A monthly weigh-in that sits on the trend and a later carry
      day at that bin paint the same Braille mask, but only the
      carry day is bold under NO_COLOR.
    evidence: >-
      plotMonthly lights both trend dots even when the bins are
      equal and then draws no stem (internal/tui/braille.go).
      tone() paints that cell with the weight role and a trend-only
      cell with the trend role. style.go keeps trend bold when
      NO_COLOR is set and does not bold weight.
    disposition: accepted
    disposition_rationale: >-
      A coincident weigh-in and a later carry day share one mask,
      and under NO_COLOR only the carry day is bold. That is the
      trend role. The weigh-in stays the weight role and is not
      bold. Narrow Decision 5, FR5, and the plan's NO_COLOR bullet:
      the same mask means the same dots, and bold still marks a
      monthly carry day. The same-picture sentence stays for a
      stem-filled mark and a pure stem, which get no extra style
      bit. Add a monthly NO_COLOR test: one weigh-in on the trend,
      a later blank day, same mask, weigh-in not bold, carry bold.
  - id: O2
    category: risk
    severity: medium
    claim: >-
      A monthly cell that holds both a stem dot and a mark dot is
      stem green in tone(), but the file preamble says the mark
      wins a shared cell, and no test fails if the mark takes it
      back.
    evidence: >-
      internal/tui/braille.go preamble says the monthly mark wins
      a shared cell. tone() returns stem green whenever a stem bit
      is set, before the mark. TestChartBrailleStemInsideRow checks
      S4's dots and not the colour. TestChartBrailleStemIsGreen
      colours the trend cell, which has no mark. TestChartStemGreen
      skips any cell with a right-hand dot.
    disposition: accepted
    disposition_rationale: >-
      Stem green wins a shared cell, including S4. That is
      Decision 5. Replace the preamble sentence that says the
      monthly mark wins a shared cell. Extend
      TestChartBrailleStemInsideRow so S4's rune is stem green.
      When the mark sits on the trend there is no stem, and that
      cell keeps the weight role.
  - id: O3
    category: risk
    severity: medium
    claim: >-
      The monthly custom-weight check treats any green Braille
      cell as the weight role, and stem chrome is that same green,
      so a weight role that never changes still passes.
    evidence: >-
      TestChartUsesWeightAndTrendColors sets weight to green and
      calls hasFGOnBraille(view, 2). stemCell paints with
      lipgloss.Color("2") (internal/tui/chart.go). On the same
      twoDayApp fixture, TestChartStemGreen already requires a
      left-only Braille cell with foreground 2 while the weight
      role is still the default blue.
    disposition: accepted
    disposition_rationale: >-
      A custom weight colour is still required.
      TestChartUsesWeightAndTrendColors must not treat stem green
      as the weight role. Point weight at a colour the stem does
      not use, such as magenta, and assert it on a mark cell that
      contains no stem dot. Leave the yellow check on the trend
      role.
---

## O1 — implementation — high

### Claim

A monthly weigh-in that sits on the trend and a later carry day at
that same bin paint one Braille mask. Under `NO_COLOR` only the carry
day is bold, so the two cells are not the same picture.

### Evidence

A coincident day still lights both trend dots, and an equal bin draws
no stem:

```62:69:internal/tui/braille.go
	if hasTrend {
		trendBin = brailleBin(trendY, ymin, ymax, brailleBins)
		g.light(col, trendBin, 0, dotTrend)
		g.light(col, trendBin, 1, dotTrend)
	}
	if !hasWeight || !hasTrend || markBin == trendBin {
		return
	}
```

The union is the left dot and the right dot at that one bin. A carry
day has no mark, so it lights the same two dots and nothing else.
`tone` then sends those identical masks to different styles:

```134:142:internal/tui/braille.go
	if monthly {
		if d.bits[dotStem] != 0 {
			return toneStem
		}
		if d.bits[dotMark] != 0 {
			return toneWeight
		}
		return toneTrend
	}
```

```105:111:internal/tui/braille.go
	switch d.tone(g.monthly) {
	case toneWeight:
		return p.weight.Render(ch)
	case toneStem:
		return stemCell(ch)
	case toneTrend:
		return p.trend.Render(ch)
```

Weight is not bold. Trend is bold, and `NO_COLOR` does not remove
that bit:

```54:62:internal/tui/style.go
	noColor := os.Getenv("NO_COLOR") != ""
	fg := func(role string, bold bool) lipgloss.Style {
		s := lipgloss.NewStyle()
		if bold {
			s = s.Bold(true)
		}
		if noColor {
			return s
		}
```

```75:76:internal/tui/style.go
		weight:      fg(config.RoleWeight, false),
		trend:       fg(config.RoleTrend, true),
```

The first logged weight is the first trend, and a blank day keeps
that trend (`internal/dailylog/trend.go`). One weigh-in, then the
rest of the month, is therefore one rune repeated: the weigh-in
column unbold, every carry column bold. `renderPlot` plots every
point in the span, including those carry days
(`internal/tui/chart.go`). `TestChartNoStemWhenMarkOnTrend` builds
this month and only checks that day 1 has one left bin and a matching
right bin. `TestChartBrailleNoColorKeepsDots` uses S2, where the mark
and the trend are different rows, so the masks differ and bold is
never compared.

Long-term cells do not collide this way. A weight-only mask has no
right-hand bit, and a trend mask always does (`plotLong`).

### Why this matters

The flat month is the ordinary chart: a new logbook, or any day the
mark sits on the trend and later days are blank. With colour, the
weigh-in is the weight role and the carry is the trend role, which is
the role rule. Without colour, bold puts the series distinction back
onto two cells whose dots match. The high-resolution contract says a
long-term trend cell stays bold and that two cells with the same mask
are the same picture. This painter keeps the bold trend style on the
monthly chart, so that sentence is false for the mask a coincident
day and a carry day share. No monthly test asserts the bold bit, so
the suite stays green either way.

## O2 — risk — medium

### Claim

`tone` paints a monthly cell stem green when any stem dot is present,
including a cell that also holds the mark. The preamble of the same
file says the mark wins a shared cell, and the tests never colour the
cell where both are true.

### Evidence

The literate preamble states the other winner:

```9:12:internal/tui/braille.go
// Bin group 0 sits on the bottom row, toward Ymin, because weight
// still has to increase upward. The monthly mark wins a shared cell;
// on the long-term chart the trend wins, and that cell stays bold.
// Stems are monthly only, and only in the mark's own column. This
```

The function under it does the opposite when a stem bit is set. Stem
is checked first (`internal/tui/braille.go`, `tone`, the monthly
branch quoted in O1). The comment on `tone` matches the function, not
the preamble.

S4 is the cell that holds the mark, the stem, and the trend together
(bins 8 through 11). The test locks the dots and stops:

```1007:1014:internal/tui/chart_test.go
func TestChartBrailleStemInsideRow(t *testing.T) {
	app := openNovemberChart(t, dec151990(),
		kgOnDay{1, 80}, kgOnDay{2, 80.50}, kgOnDay{30, 82})
	rows := mustPlot(t, monthlyView(t, app))
	if !sameSet(leftBins(rows, 1), 8, 9, 10, 11) || !sameSet(rightBins(rows, 1), 8) {
		t.Fatalf("day 2 left %v right %v, want stem dots at 9 and 10 with mark 11 and trend 8 in one row; column %q",
			leftBins(rows, 1), rightBins(rows, 1), columnGlyphs(rows, 1))
	}
```

`TestChartBrailleStemIsGreen` colours S2 day 2's trend cell (bin 8,
right side) green and the mark (bin 12) with the weight role. Those
are different cells. The trend cell has no mark bit, so a `tone` that
preferred the mark would still paint it green.

`TestChartStemGreen` only accepts a left-only cell, and its comment
calls that "no right dot":

```662:669:internal/tui/chart_test.go
	found := false
	for _, row := range rows {
		for _, c := range row {
			m := maskOf(c.r)
			// A pure stem cell has left dots and no right dot.
			if !isBraille(c.r) || m&sideBits(true) == 0 || m&sideBits(false) != 0 {
				continue
			}
```

S4's shared cell has the trend's right dot, so this loop skips it.
`TestChartDailyMarkWinsSharedCell` is the equal-bin case, which draws
no stem, so the mark winning there does not pin the stem.

### Why this matters

Swapping the monthly checks so a mark bit wins before a stem bit
leaves every current colour assertion green: S2's stem cell has no
mark, a pure stem cell stays green, and a coincident day has no stem.
S4's one rune would turn the weight colour for the mark, the stem,
and the trend. The stem would not be one colour along its length in
the scenario that puts that length inside one cell. The preamble
already tells the next edit to make that swap.

- **accept-as-stated** — stem green wins even when the mark shares
  the cell. Replace the preamble sentence so the next reader inherits
  that, not the old shared-cell rule.
- **revise-spec** — the preamble is the rule that should ship, and
  the spec's "any stem dot" sentence is too wide. Change the spec.
- **add-test** — colour S4's single rune and require stem green, not
  the weight role.
- **consciously-carry** — ship with the branch untested and the
  preamble wrong, on the record, because the painter is right today.

## O3 — risk — medium

### Claim

`TestChartUsesWeightAndTrendColors` reads any green Braille cell as
the custom weight role. Stem chrome is that same green on the fixture
the test uses, so the weight role can stay at the default and the
test still passes.

### Evidence

The assertion is any Braille cell whose foreground is index 2, after
setting the weight role to green and the trend role to yellow:

```165:182:internal/tui/chart_test.go
func TestChartUsesWeightAndTrendColors(t *testing.T) {
	enableChroma(t)
	store := openStore(t)
	app := twoDayApp(t, store)
	app.cfg = schemeCfg(map[string]string{"weight": "green", "trend": "yellow"})
	app.pal = newPalette(app.cfg)
	press(app, "c")
	view := app.View()
	if strings.Contains(visible(view), "daily log") {
		t.Fatalf("chart missing: %q", view)
	}
	if !hasFGOnBraille(view, 2) {
		t.Fatalf("missing green weight on a braille cell")
	}
	if !hasFGOnBraille(view, 3) {
		t.Fatalf("missing yellow trend on a braille cell")
	}
}
```

`hasFGOnBraille` is true for the first Braille cell with that index
(`internal/tui/braille_test.go`). It does not require a mark dot or
the weight style.

Stem chrome is hard-wired to the same index, and it does not consult
the weight role:

```118:122:internal/tui/chart.go
func stemCell(ch string) string {
	if os.Getenv("NO_COLOR") != "" {
		return ch
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(ch)
```

The same `twoDayApp` month already has a left-only Braille cell in
foreground 2 while weight is still the default. That is what
`TestChartStemGreen` requires (`internal/tui/chart_test.go`). Default
weight is blue, index 4 (`internal/tui/style.go`, `defaultColors`).
Index 2 on that fixture is the stem. Yellow on a carry cell does
check the trend override. Green does not check the weight override.

The long-term twin is not this hole. That chart does not draw stems,
so a green cell there is not stem chrome.

### Why this matters

Monthly marks are specified to follow the `weight` role, including a
custom green. The one test that retargets that sentence onto Braille
passes when every mark cell stays default blue, because the stem is
already green. A palette that stops applying `p.weight` on the
monthly chart, or a mark cell that is never the weight style, does
not go red. The trend half of the same test would still catch a
broken trend role.

- **accept-as-stated** — stem chrome and a green weight role are
  allowed to be the same index. Write down that this test does not
  prove the weight role, and name the test that does.
- **revise-spec** — custom weight colour on a Braille mark is not
  actually required. Narrow the colour sentence.
- **add-test** — set weight to a colour stem chrome does not use, and
  assert it on a mark cell that contains no stem dot.
- **consciously-carry** — leave the false green witness in place,
  knowing a weight-role regression on the monthly chart stays green.

## Explicitly not objecting to

- **Bin orientation and the named quantizer**: Group 0 is the bottom
  row, Ymax lands in the last bin, and S2's 12 and 14 and S4's 7, 8,
  and 11 are the half-open bins the painter uses.
- **Long-term weight dots inside a bucket**: `plotLong` is only given
  a weight when the span is still one day per column, and
  `TestLongChartBucketOmitsWeightDot` requires each annual column to
  be the trend bin alone.
- **PDF, keys, and the eight Y labels**: `p` still calls `chartpdf`,
  the help strings are unchanged, and the gutter is still `%5.1f-`.
  Labels that name a row rather than one of its four dots are the
  scale this slice was told to keep.
- **Stale `/\-` and `|` checks**: `TestLongChartNoColorKeepsTwoLines`
  can see the gutter dash, and `plotHas` will not see a Braille stem.
  S8's bold-versus-plain Braille check and the bucket test already
  read the dots, so those ASCII checks are not the only witness.
- **A stroke between adjacent days**: Each column lights its own bin.
  The spec does not ask for dots in the columns between two days, and
  the long-term chart still omits stems.
- **The two-day x-label clamp**: Forcing one space before a one-digit
  last day hangs the `2` one column past a two-day plot. That is the
  old label arithmetic, and the Braille columns themselves still
  follow the clipped span.
