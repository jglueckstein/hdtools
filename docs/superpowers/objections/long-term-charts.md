---
spec: docs/superpowers/specs/2026-09-12-long-term-charts.md
date: 2026-09-12
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: critical
    claim: "End month is defined both as calendar-today and as the latest log's month, so a database whose last weigh-in is in the past either plots last-N-months-of-data or an empty window in the present."
    evidence: "Decision 1: 'today if any loaded log's month contains today or we are viewing “now”; otherwise the month of the latest log' and 'these four charts are always “last N months” / all history ending at current data'; Decision 2: 'empty if the end month is entirely after today'."
    disposition: accepted
    disposition_rationale: "End month is the latest loaded log's month. If that month contains today, last plotted day is today (same clip as monthly). Never use a calendar month after the latest log. A 1990 database in 2026 plots 1990, not an empty 2026 window."
  - id: O2
    category: premise
    severity: high
    claim: "Last-weight-in-bucket is not a daily-weight line, so annual and complete charts (which always bucket on an 80-column terminal) cannot satisfy US2's thin daily line."
    evidence: "US2: 'daily weight as a thin line'; Decision 5: 'Daily weight is a thin connected path'; Decision 6: 'If D > W, W equal-time buckets … last daily weight in the bucket'; Background: 'A year of days will not fit one character per day on an 80-column terminal.'"
    disposition: accepted
    disposition_rationale: "Addressed with O8: when D > W omit the daily path (trend only). When D ≤ W keep the thin daily line. Annual/complete on 80 columns are trend-only."
  - id: O3
    category: implementation
    severity: high
    claim: "`[` / `]` shift the monthly chart's month and cycle long-term kind, and `l` is reachable from the monthly chart, so the same keys move two different axes with no specified cue which axis is live."
    evidence: "Decision 3: '`l` works from … the monthly chart'; Decision 4: '`[` / `]` cycle the four kinds' and 'They do not change calendar month on the month sheet'; monthly-charts Decision 8: '`[` / `]` change month on the chart'."
    disposition: accepted
    disposition_rationale: "Keep [ ] cycling kinds. Specified cue: the title box always names the kind; the help line on this screen is `[ ] kind`, not `[ ] month`."
  - id: O4
    category: specification quality
    severity: high
    claim: "Equal-time buckets do not specify how leftover days are spread, so two 72-column annual plots can bucket different days and both pass S15."
    evidence: "Decision 6: 'W equal-time buckets covering the span'; S15: 'the plot has 72 columns (not ~365)'; FR7: 'else W equal-time buckets'."
    disposition: accepted
    disposition_rationale: "Bucket i (0-based) covers days [floor(i*D/W), floor((i+1)*D/W)) of the span. Remainders spread; S15 still requires W columns."
  - id: O5
    category: specification quality
    severity: medium
    claim: "Decision 2's 'end month entirely after today' clause conflicts with Decision 1's fallback to the latest log, so last-plotted-day is not a single function."
    evidence: "Decision 2: 'today if the end month contains today; last calendar day if the end month is past; empty if the end month is entirely after today'; Decision 1: 'otherwise the month of the latest log'."
    disposition: accepted
    disposition_rationale: "Last plotted day uses the monthly clip on the chosen end month only: today if that month contains today, else last calendar day if past. Drop 'empty if the end month is entirely after today'; this screen does not select a future end month."
  - id: O6
    category: specification quality
    severity: medium
    claim: "S14 allows the complete-history start to be either the first log's date or merely 'the first bucket includes that day', so a title that says April 1989 can still drop 15 April from the span."
    evidence: "S14: 'the span's first day is 15 April 1989 (or the first bucket includes that day)'; Decision 1: 'the date of the first loaded log through that last day'."
    disposition: accepted
    disposition_rationale: "Complete history starts on the first log's calendar date. Drop S14's parenthetical. Buckets cover that span; they do not start earlier or skip the first log day."
  - id: O7
    category: specification quality
    severity: medium
    claim: "S8's Given is two different empty cases joined by or, so an empty list and a list whose quarterly window has no trend are one scenario and can hide a crash in one of them."
    evidence: "S8 Given: 'no logs (or no trend in the quarterly window)'; When: '`l` is pressed from an empty list'."
    disposition: accepted
    disposition_rationale: "Split S8 (empty list) and S8a (logs exist, quarterly window has no trend). Both: empty-state copy, no Loss numbers, no panic."
  - id: O8
    category: alternatives
    severity: medium
    claim: "The book already says not to plot daily weights on long-term charts by hand; omitting the daily path when D > W is simpler and avoids O2, and the spec does not acknowledge it."
    evidence: "Background: 'The book notes that by hand you would skip daily weights because they make the graph busy; Excel plots them because it is free. This TUI follows Excel and plots both.'; Weight Monitoring: 'If you do decide to plot long term charts, don't bother with the daily weights.'"
    disposition: accepted
    disposition_rationale: "When D > W, omit the daily path (trend only) — the book's hand rule, because a last-in-bucket sample is not a daily line. When D ≤ W, plot the thin daily line as Excel does."
---

## O1 — specification quality — critical

### Claim

End month is defined both as calendar-today and as the latest log's
month, so a database whose last weigh-in is in the past either plots
last-N-months-of-data or an empty window in the present.

### Evidence

Decision 1:

> The **end month** is the month of the most recent day that the
> monthly chart would plot: today if any loaded log's month contains
> today or we are viewing “now”; otherwise the month of the latest
> log.
>
> these four charts are always “last N months” / all history ending at
> current data

Decision 2:

> empty if the end month is entirely after today

S16 freezes today at 10 November 1990, so the two readings coincide
in tests and will not catch the split.

### Why this matters

In 2026 a log that ends in 1990 is a realistic fixture and a realistic
user. If end month is calendar today, quarterly is Jul–Sep 2026 and
always empty. If end month is the latest log, quarterly is the last
three months of data. Those are different products. The feature should
not proceed until one predicate is the only one.

## O2 — premise — high

### Claim

Last-weight-in-bucket is not a daily-weight line, so annual and
complete charts (which always bucket on an 80-column terminal) cannot
satisfy US2's thin daily line.

### Evidence

US2 asks for “daily weight as a thin line”. Decision 5 requires a
thin connected path. Decision 6, when *D* > *W*, takes the **last**
daily weight in each bucket. Background already admits a year will
not fit one column per day. Annual is ~365 days against *W* = 72, so
the views the book illustrates as year and complete history are
always subsampled.

### Why this matters

A bucket that ends on a water spike plots a rise while the omitted
days fell. The thin line is then not daily weight; it is an aliased
sample. US2, S5, and the claim to follow Excel's two-line long-term
chart fail on the two kinds that need long-term most.

## O3 — implementation — high

### Claim

`[` / `]` shift the monthly chart's month and cycle long-term kind,
and `l` is reachable from the monthly chart, so the same keys move
two different axes with no specified cue which axis is live.

### Evidence

Decision 3 binds `l` from the monthly chart. Decision 4 binds `[` /
`]` to cycle kinds and says they do not change the month sheet.
The monthly-chart spec binds the same keys to change month. S4 only
asserts kind names; S9 only asserts the sheet month after Esc. Neither
requires the long-term screen to say that `[` / `]` are cycling
kinds.

### Why this matters

`c` then `]` is next month; `l` then `]` is semiannual. Both can pass
their tests while the user cannot tell which axis they are on. That
is a class of wrong-chart failures, not a taste in keycaps.

## O4 — specification quality — high

### Claim

Equal-time buckets do not specify how leftover days are spread, so
two 72-column annual plots can bucket different days and both pass
S15.

### Evidence

Decision 6 and FR7 say “*W* equal-time buckets covering the span”.
S15 only requires 72 columns, not 365. 365 / 72 is not an integer.
The spec never says whether remainders go to the first buckets, the
last, or are spread.

### Why this matters

Loss uses first and last *day* of the span, but the daily path uses
last-in-bucket. Different remainder rules put different days in the
end buckets, so the thin line's last point is not a stable function
of the logs. Implementers can ship two annual charts that both
“have 72 columns”.

## O5 — specification quality — medium

### Claim

Decision 2's “end month entirely after today” clause conflicts with
Decision 1's fallback to the latest log, so last-plotted-day is not
a single function.

### Evidence

Decision 2's third clause empties a future end month. Decision 1
says if we are not “viewing now”, end month is the latest log (a
past month). Both cannot govern the same App state. S16 only
exercises the case where today lies inside the end month.

### Why this matters

`lastPlottedDay` on the monthly chart already has this tri-state.
Copying it without saying which Decision 1 branch selected the end
month reopens the in-progress-month bug the monthly diaboli already
caught.

## O6 — specification quality — medium

### Claim

S14 allows the complete-history start to be either the first log's
date or merely “the first bucket includes that day”, so a title that
says April 1989 can still drop 15 April from the span.

### Evidence

S14 Then: “the span's first day is 15 April 1989 (or the first
bucket includes that day)”. Decision 1: “the date of the first
loaded log through that last day”.

### Why this matters

The parenthetical lets a bucket that begins 1 April satisfy the
scenario while the plotted span starts before the first weight, or
a bucket that begins 16 April while 15 April is omitted. Complete
history then does not mean complete.

## O7 — specification quality — medium

### Claim

S8's Given is two different empty cases joined by or, so an empty
list and a list whose quarterly window has no trend are one
scenario and can hide a crash in one of them.

### Evidence

S8 Given: “no logs (or no trend in the quarterly window)”. When:
“`l` is pressed from an empty list”. Decision 13 distinguishes no
trend from carried-only trend.

### Why this matters

The When clause only drives the empty-list path. A store with logs
whose quarterly window has no trend is never executed. US4 asked
for both.

## O8 — alternatives — medium

### Claim

The book already says not to plot daily weights on long-term charts
by hand; omitting the daily path when *D* > *W* is simpler and
avoids O2, and the spec does not acknowledge it.

### Evidence

Background quotes the book's hand advice (skip daily weights) and
then chooses Excel's “plot both because it is free”. Weight
Monitoring: “If you do decide to plot long term charts, don't
bother with the daily weights.” Decision 6 still requires a daily
path through last-in-bucket samples.

### Why this matters

The TUI is not Excel: extra series are not free once they require
aliasing. Trend-only when bucketed is the book's own cheaper
long-term picture. Leaving it unlisted makes “follow Excel” look
like the only option.

## Explicitly not objecting to

- **Four WEIGHT-menu kinds**: Quarterly, semiannual, annual, and
  complete history are what the cited chapter lists; the first draft's
  single 12-month view was the error, not this list.
- **First-and-last trend vs least-squares**: The book tells hand
  users to use endpoints and says Excel's extra fit is optional; the
  spec already reuses `MonthlyBalance`.
- **No stems on the long-term plot**: The book switches to two lines
  for these charts; stems stay on the monthly sitting.
- **Key `l` vs `C`**: A free letter is a choice, not a failure class,
  as long as editing-on-the-sheet is specified (it is).
- **Default plot width 72**: Tests need a number; 72 is enough to
  force annual bucketing and to fit an 80-column guttered plot.
- **PDF and diet-plan overlay**: Honest out-of-scope; the monthly
  specs already deferred PDF.
