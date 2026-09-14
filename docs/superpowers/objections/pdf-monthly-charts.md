---
spec: docs/superpowers/specs/2026-09-12-pdf-monthly-charts.md
date: 2026-09-14
mode: spec
diaboli_model: grok-4.6
objections:
  - id: O1
    category: specification quality
    severity: high
    claim: "S7 treats a month with no rows as empty while Decision 4 requires on-screen carry-forward, and the CLI never states that the writer receives the full ApplyTrend series."
    evidence: "Decision 4: 'Carry-forward in the span, no trend into the future. … Empty span: title + empty copy, no axes.' S7 Given: 'no logs in June 1990' Then: 'extractable text includes `June 1990` and empty-state copy'. Decision 2: 'flags, open store, call the writer.'"
    disposition: accepted
    disposition_rationale: "Empty means the plotted span has no daily marks and no trend (including no carry). A June with no rows but a May trend is a flat-trend PDF, not empty. Both CLI and TUI load the full series, ApplyTrend, then clip. Add an S7a analogue."
  - id: O2
    category: specification quality
    severity: high
    claim: "Decision 6 maps only 16-color names to RGB, so configured hex weight/trend roles cannot be honoured, and no RGB table is given for those names, Excel title chrome, or stem green."
    evidence: "Decision 6: 'Daily marks and the trend path use the configured `weight` and `trend` roles, mapped from the same 16-color names as the TUI to RGB. Stems are green (Excel default, not a `[colors]` key).' Decision 5: 'Fill blue, stroke red, text yellow — Excel chrome, not `[colors]` roles.'"
    disposition: accepted
    disposition_rationale: "Hex #rrggbb / #rgb is RGB as written. Named 16-color values use a fixed VGA table (the TUI names). Title chrome and stem green are explicit RGB (title fill #0000FF, stroke #FF0000, text #FFFF00; stems #008000). NO_COLOR still does not grey the PDF."
  - id: O3
    category: specification quality
    severity: high
    claim: "Decision 4 clips a current month at today, but no acceptance scenario freezes today or forbids columns after today, so a current-month PDF can plot the rest of the calendar month."
    evidence: "Decision 4: 'Plotted span is day 1 through last plotted day (today if the month contains today; last calendar day if past).' S1–S10 never freeze today; S1 uses November 1990 as a past fixture; S7 uses June 1990. The companion plan lists no current-month clip test."
    disposition: accepted
    disposition_rationale: "Keep Decision 4. Add a current-month scenario (today 10 Nov 1990: no column after the 10th; Daily Deficit divides by 10). Freeze today in that test. CLI and TUI use the same local today."
  - id: O4
    category: specification quality
    severity: high
    claim: "Decision 4 assigns loss and deficit to MonthlyBalance on the plotted span, which does not omit numbers when an endpoint has no trend, so a PDF can print analysis the on-screen chart would hide."
    evidence: "Decision 4: 'Monthly Loss and Daily Deficit use `MonthlyBalance` on the plotted span. Empty span: title + empty copy, no axes.' FR3: 'match the on-screen monthly chart’s *data* rules.' Parent spec 2026-09-10 Decision 11 / FR13: omit the two numbers if day 1 or the last plotted day has no trend."
    disposition: accepted
    disposition_rationale: "If day 1 or the last plotted day has no trend, omit Loss and Daily Deficit (parent FR13). MonthlyBalance only after both endpoints have trend. Empty span still has no analysis."
  - id: O5
    category: specification quality
    severity: medium
    claim: "Empty-state copy and the analysis line are not frozen strings, and S6’s wording does not match the on-screen analysis line, so TUI and PDF copy will diverge under the same data rules."
    evidence: "S7: 'empty-state copy' with no string. S6: 'extractable text includes Monthly Loss 1.0 kg and Daily Deficit 257 calories'. On-screen analysis is `Monthly loss: %.1f %s   Daily deficit: %d cal`. US3 only requires that the PDF 'says the month is empty'."
    disposition: accepted
    disposition_rationale: "Match the TUI strings: empty is (empty month); analysis is Monthly loss: %.1f %s   Daily deficit: %d cal. S6 must not freeze calories."
  - id: O6
    category: specification quality
    severity: medium
    claim: "S1–S2 and the Observing section lock `%PDF`, page count, and extractable title, not Letter landscape, numeric Y bounds, or title fill colour, and they allow a debug dump in place of the PDF."
    evidence: "Decision 1 / FR1: landscape US Letter. S1 Then: 'exists, starts with `%PDF`, and has one page' — no size or orientation. S5 states Ymin/Ymax with no observation method. Observing: 'tests may use the writer’s debug dump or a small inspector, not a human screenshot, for CI. A rendered PNG (`pdftoppm`) is optional for a human check, not a CI gate.'"
    disposition: accepted
    disposition_rationale: "S1 locks landscape US Letter from the PDF page box (no pdftoppm). Y bounds and vector presence come from the PDF or a specified inspector of that file, not an optional dump that can pass without a real path. Title RGB stays a human PNG check, not a CI gate."
  - id: O7
    category: alternatives
    severity: medium
    claim: "The spec requires identical clip, Y, empty, and analysis rules while placing the writer outside internal/tui with no shared owner for those rules, so a second implementation is the default and will drift."
    evidence: "Decision 4 / FR3: same data rules as the on-screen monthly chart. Decision 2 / FR7: writer under `internal/`, 'not in `internal/tui`, which must stay SQL-free and must not grow a PDF dependency.' Companion plan: `internal/chartpdf/span.go` copies last-plotted-day clip because 'copy is cheaper than pulling PDF into `tui`.'"
    disposition: accepted
    disposition_rationale: "Do not copy clip, empty, Y pad, or analysis-omission into chartpdf. Put those next to MonthlyBalance (or a small helper both packages call). internal/tui still must not import a PDF library."
  - id: O8
    category: risk
    severity: medium
    claim: "TUI `p` writes `YYYY-MM-chart.pdf` in the process cwd and overwrites without confirmation, with no write-failure behaviour and no file mode, so a keystroke can clobber a file or leave a shareable weight chart in an unexpected directory."
    evidence: "Decision 9: 'Overwrite if the path exists.' Decision 10: 'Current directory for TUI `p` is the process working directory, not XDG.' S3 Then: file written and 'the status line contains that path' — success only. S8 covers only a bad `-chart-pdf` value, not I/O errors."
    disposition: accepted
    disposition_rationale: "Keep cwd and overwrite. On write failure: error status, do not quit, do not claim saved. File mode 0600 (same personal series as the DB)."
---

## O1 — specification quality — high

### Claim

S7 treats a month with no log rows as empty, while Decision 4 requires the on-screen carry-forward rule, and the CLI path never says the writer is given the full chronological `ApplyTrend` series. A June with no rows but a May trend can therefore be an empty PDF from `-chart-pdf` and a flat-trend chart from `c`.

### Evidence

Decision 4:

> Same data as the on-screen monthly chart. … Carry-forward in the span, no trend into the future. … Empty span: title + empty copy, no axes.

S7:

> **Given** no logs in June 1990
> **When** `-chart-pdf 1990-06` runs
> **Then** a one-page PDF is written
> **And** extractable text includes `June 1990` and empty-state copy

The parent monthly-chart spec (Decision 12 / S7a) already had to say that a month with no log rows **but** a carried trend is not empty. This spec does not restate that, and Decision 2 only says `cmd/hdtools` will “open store, call the writer” — not `All()` then `ApplyTrend` on the whole book versus `Range` of June alone. Loading only the target month drops carry-forward.

### Why this matters

“Same data as the on-screen monthly chart” fails on the exact trap this project already documented: a month with no rows is not always empty. S7 as written is a passing test for the wrong empty definition. Two entry points that share a writer still disagree if CLI wiring passes a different series than `a.logs`.

## O2 — specification quality — high

### Claim

Decision 6 maps configured `weight` and `trend` roles from 16-color **names** to RGB. Those roles already accept hex (`#rrggbb` / `#rgb`) and the spec gives no RGB table for names, for Excel title chrome, or for stem green. Implementers cannot honour the configured roles, and two palettes can both claim compliance.

### Evidence

Decision 6:

> Daily marks and the trend path use the configured `weight` and `trend` roles, mapped from the same 16-color names as the TUI to RGB. Stems are green (Excel default, not a `[colors]` key).

Decision 5:

> Fill blue, stroke red, text yellow — Excel chrome, not `[colors]` roles. PDF is a print artefact: **always colour**; `NO_COLOR` does not apply.

US2 wants the PDF to show the same picture the user already knows from the screen. Config `[colors]` is a sparse overlay of names **or hex**; hex is stored as `#rrggbb` and never becomes a 16-color name. FR6 says `NO_COLOR` does not grey the PDF, but no scenario checks that a colour PDF is still produced under `NO_COLOR`.

### Why this matters

A user who set `weight = "#c0a080"` (the reason hex exists) gets a PDF that cannot follow Decision 6 as written. Even named colours diverge: CSS `blue`, VGA ANSI blue, and Excel title blue are different RGB triples. S2 asserts a filled blue rectangle with red stroke without defining those bytes, so CI can pass a title box that does not match the TUI or the book.

## O3 — specification quality — high

### Claim

Decision 4 clips the current month at today, but no acceptance scenario freezes today or forbids columns after today. A current-month PDF can plot the rest of the calendar month, including carried trend into the future.

### Evidence

Decision 4:

> Plotted span is day 1 through last plotted day (today if the month contains today; last calendar day if past). … no trend into the future.

S1’s fixture is November 1990 as a completed past month. S7 is June 1990. None of S1–S10 set “today”, mention remaining days of the current month, or check Daily Deficit’s divisor. The parent monthly-chart spec needed S15 for this. The companion plan says to freeze today at 10 November 1990 “when the clip matters” but its test list has no clip case.

### Why this matters

The current month is the month people export. The on-screen chart already encodes “do not plot carried trend into the future.” Without a PDF scenario, TDD from this spec’s list will not catch a writer that uses `daysInMonth` for every month. CLI and TUI can also disagree on whose clock is “today” because the spec never says the CLI uses the same local today as the TUI.

## O4 — specification quality — high

### Claim

Decision 4 sends Monthly Loss and Daily Deficit through `MonthlyBalance` on the plotted span. That function does not omit results when day 1 or the last plotted day has no trend. A PDF can therefore print numbers the on-screen chart hides.

### Evidence

Decision 4:

> Monthly Loss and Daily Deficit use `MonthlyBalance` on the plotted span. Empty span: title + empty copy, no axes.

FR3 says geometry is vector but data rules match the on-screen chart. The parent spec’s Decision 11 / FR13 / S13: if either endpoint has no trend, the two numbers are **omitted** — not printed as zero, and not computed from a missing endpoint. `MonthlyBalance` only returns `ok == false` when `days` is not positive. S6 only covers a 30-day past month whose first and last days both have trend. There is no PDF analogue of S13.

### Why this matters

A first weigh-in mid-month with no prior carry is a normal log. The TUI guards `HasTrend` on both ends, then calls `MonthlyBalance`. An implementer who follows this spec’s named function without that guard will emit a loss/deficit line from a zero or absent day-1 trend. That is a wrong number on a print artefact, not a formatting quibble.

## O5 — specification quality — medium

### Claim

Empty-state copy and the analysis line are not specified as strings, and S6’s extractable wording does not match the on-screen analysis line, so TUI and PDF will show different copy for the same month.

### Evidence

S7 requires “empty-state copy” and US3 requires that the PDF “says the month is empty.” Neither freezes the TUI’s `(empty month)`.

S6:

> extractable text includes Monthly Loss 1.0 kg and Daily Deficit 257 calories

The on-screen line is `Monthly loss: %.1f %s   Daily deficit: %d cal` (lowercase “loss”, `cal` not `calories`, different punctuation). FR3 claims data-rule parity, not copy parity; S6 then freezes different copy.

### Why this matters

Two implementers can both pass S6/S7 with different user-visible strings, and a third can copy the TUI line and fail S6 on `calories` versus `cal`. Tests that `strings.Contains` S6’s phrase will lock PDF wording that is not the screen’s wording, which is the opposite of US2.

## O6 — specification quality — medium

### Claim

S1–S2 and the Observing section lock a `%PDF` header, one page, and extractable title text. They do not lock Letter landscape, the numeric Y range, or title fill colour, and they explicitly allow a writer debug dump instead of inspecting the PDF.

### Evidence

Decision 1 and FR1 require one-page landscape US Letter. S1’s Then is only: the process exits 0, the file exists, it starts with `%PDF`, it has one page, and no TTY is required.

S2 requires a filled title rectangle (blue fill, red stroke) and vector marks/stems/trend. Observing:

> Tests must not require a TTY. A **valid PDF** starts with `%PDF` and opens as one page (Go PDF reader or `pdfinfo`). **Extractable text** is whatever the writer embeds as text (title, axis labels, loss line). Vector paths: at least one stroke for the trend and at least one for a stem on a two-point month; tests may use the writer’s debug dump or a small inspector, not a human screenshot, for CI. A rendered PNG (`pdftoppm`) is optional for a human check, not a CI gate.

S5 states Ymin = 80 − *P* through Ymax = 81 + *P* with no required labels or dump shape.

### Why this matters

A portrait A4 page that starts `%PDF` and contains `November 1990` satisfies S1–S2’s testable parts and fails Decision 1. A debug dump that claims a stem existed can pass CI while the PDF has none. S5’s Y range is a mapping from kg to page Y; without an observation contract, it is not an acceptance test. The look US2 asked for is then a human PNG, which this spec says is not a gate.

## O7 — alternatives — medium

### Claim

The spec requires identical clip, Y, empty, and analysis rules while forbidding the PDF writer from living in `internal/tui`, and it names no shared owner for those rules. Copying them into a second package is the default and will drift.

### Evidence

Decision 4 and FR3 require the on-screen monthly chart’s data rules (span clip, carry, no future trend, ±2 lb Y, `MonthlyBalance`, empty). Decision 2 / FR7: the writer lives under `internal/`, not in `internal/tui`, which “must stay SQL-free and must not grow a PDF dependency.” The companion plan makes the copy explicit: `internal/chartpdf/span.go` holds last-plotted-day clip from already-trended logs because “copy is cheaper than pulling PDF into `tui`.”

A render-agnostic helper (span clip, empty, Y pad, analysis omission) can sit beside `dailylog.MonthlyBalance` without importing a PDF library or SQLite. The spec does not require that, and does not forbid `internal/tui` from calling such a helper.

### Why this matters

O1, O3, and O4 are how duplication fails in this sitting. Even if those scenarios are added, geometry (Y mapping, day positions, zero-length stems) can still diverge the next time one renderer is patched. “Copy is cheaper” is cheaper only until the next carry-forward or today-clip bug.

## O8 — risk — medium

### Claim

TUI `p` writes `YYYY-MM-chart.pdf` in the process working directory and overwrites if that path exists, with no specified behaviour when the write fails and no file mode. A keystroke can destroy a same-named file or leave a shareable weight chart in whatever cwd the process inherited.

### Evidence

Decision 9: “Overwrite if the path exists.” Decision 10: “Current directory for TUI `p` is the process working directory, not XDG.” S3: on success, `1990-11-chart.pdf` is written and the status line contains that path. S8 is invalid `-chart-pdf` values only (non-zero exit, no file, no TUI). FR2 does not mention errors. The database and `config.toml` are created mode 0600; this spec does not mention mode for the PDF.

### Why this matters

`p` is a single key on a screen that otherwise does not write files. Cwd for a TUI is often `$HOME`, a project root, or a launcher’s directory — not a folder the user chose. Silent overwrite is data loss; a write error with only a success status specified is a panic, a swallowed error, or a stuck “saved” line, depending on the implementer. The file is the same personal weight series the XDG database protects, now sitting in a user-visible directory with unspecified permissions.

## Explicitly not objecting to

- **Chart PDFs before filled monthly log-sheet PDFs**: `idea.md` lists both; this spec names sheets as a different sitting. US1 is over-broad (“take a month home”) but the slice title is the chart.
- **US Letter and not A4**: explicit cut, with a reason (the book tools were US Excel). Objecting only that S1 does not *test* Letter (O6), not that Letter is the wrong default for this sitting.
- **Always-colour PDF (`NO_COLOR` ignored)**: stated print-artefact policy in Decision 5 / FR6. Objecting to the incomplete series mapping (O2), not to keeping chroma on paper.
- **Excel title-box chrome outside `[colors]`**: follows `2026-09-11-monthly-chart-floats-sinkers.md`. Mixed Excel chrome plus configured series colours is a coherent print choice once RGB is defined.
- **Two entry points sharing one writer**: CLI is how S1 exports without a TTY; `p` is the chart-screen action. The failure is unspecified CLI series/errors, not the dual trigger.
- **Vector rewrite rather than rasterising the TUI**: character cells, Braille, and 16-color ANSI cannot be the print artefact. That premise holds.
- **Labelling every day number on X**: Decision 7’s “PDF has room” is plausible on landscape Letter; crowding is not a specified failure.
- **Overwrite on an explicit `-o` path**: expected for a user-supplied destination. O8 is the default TUI filename in cwd.
- **Gain as negative loss**: `MonthlyBalance` already encodes sign; a second PDF scenario would not catch a new class of failures if O4’s endpoint-omission rule is stated.
- **PDF library choice (`fpdf` vs a tiny writer)**: the spec requires a valid one-page PDF, not a dependency. That belongs in the plan.
- **In-app help line mentioning `p`**: Documentation already requires a how-to and a reference line for `-chart-pdf` / `p`. Missing chrome on the chart help row is discoverability, not an untested failure class.
- **Print spool, email, TUI preview, long-term PDFs**: named out of scope; excluding them does not make the monthly chart PDF wrong.
