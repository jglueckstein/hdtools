# Compound Learning

<!-- This file is the project's persistent memory across AI sessions.
     Review new entries with the same scepticism you would apply to any
     generated content. Entries should reflect observed reality.
     Coding and style rules live in CONVENTIONS.md. Read that file
     before writing or reviewing code. -->

## STYLE

Coding and style rules live in `CONVENTIONS.md`. Read that file
before writing or reviewing code. Do not add style rules here.

## GOTCHAS

- `gofmt -l .` exits 0 even when it prints paths. CI and humans must
  treat non-empty output as failure (`test -z "$(gofmt -l .)"`).
- `ApplyTrend` needs the full chronological series. Loading only the
  current month drops carry-forward from the previous month.
- A chart of the current calendar month ends at today. Do not plot
  carried trend into the future.
- The filled log-sheet PDF lists every calendar day of the named
  month, including days after today. A later day with no weigh-in
  still shows the carried trend. Do not apply the chart clip to
  that page.
- Display-unit conversion (`units.FromKG`) is for the TUI only. Saving
  must go through `units.ToKG` so SQLite always stores kilograms.
- `View()` emits chroma unless `NO_COLOR` is non-empty; the process does
  not consult the TTY. Tests that need color must `t.Setenv("NO_COLOR",
  "")`. Hex is `38;2;` on truecolor and downshifts on a 16-color profile.
- Integration stays on the feature branch until squash-merge. The
  default branch is `master`, not `main`, and CHANGELOG stays under
  Unreleased. Commit the reflection on that feature branch before the
  merge. The integration-agent file checks out `main` and writes a
  dated changelog; that file loses to this note.
- Every harness GC rule is auto-fix false, so findings become issues.
  A same-package unused name is not dead code until a module-wide
  search agrees.
- On Grok 1.0.46, `spawn_subagent` rejects
  `ai-literacy-superpowers:<name>` even when the error lists that
  string as available. Plugin files are `agents/<name>.agent.md`, and
  the spawn id is `ai-literacy-superpowers:<name>.agent`. Pass that
  id. The five `.claude/agents/*.md` files (`code-reviewer`,
  `integration-agent`, `orchestrator`, `spec-writer`, `tdd-agent`)
  spawn under those short names. If the `.agent` id is itself
  rejected, the host mapping has changed; use the id the error
  accepts. Do not edit the installed plugin copy except to re-apply
  the Stop-hook patch below; a plugin update overwrites it.
- A plugin update restores Stop hooks that Grok rejects. The message
  is `hook reply is not valid JSON: control character (\u0000-\u001F)
  found while parsing a string at line 2 column 0`. Grok parses a
  reply that starts with `{` as JSON. Re-apply the patch to both
  trees after every update:
  `~/.grok/installed-plugins/ai-literacy-superpowers-*/hooks/scripts/`
  and the same files under
  `~/.grok/marketplace-cache/*/ai-literacy-superpowers/hooks/scripts/`.
  In `governance-drift-check.sh`, `drift-check.sh`, and
  `framework-change-prompt.sh`, replace the `echo -e` command
  substitution with `printf '%s' "$message" | sed 's/"/\\"/g'`, so
  the line is `printf '{"systemMessage": "%s"}' "$(printf '%s' "$message" | sed 's/"/\\"/g')"`.
  The scripts store breaks as the two characters `\n`. `echo -e`
  turns those into a raw newline inside the JSON string. On Linux,
  `governance-drift-check.sh` must set
  `audit_epoch=$(date -d "$audit_date" "+%s" 2>/dev/null || date -j -f "%Y-%m-%d" "$audit_date" "+%s" 2>/dev/null || true)`
  and compute the age only when that value is non-empty. Do not
  substitute epoch 0: that reported
  `observability/governance/audit-2026-09-09.md` as about 20754
  days old. The audit is under the 90-day threshold, so the hook
  should print nothing and exit 0. Leave `markdownlint-check.sh`
  alone: it runs `echo -e` and then `tr '\n' ' '` before printing.
  Do not edit any other plugin file.
- `gh pr checks` can exit at once with "no checks reported" while
  the runs are still queued.
  Wait for both required jobs on the current head commit.
  A commit pushed while that run is in progress is not the commit
  the job checked out.
- A file that must be mode 0600 is not written with `os.Create`
  or `fpdf.OutputFileAndClose` on the destination. Those truncate
  the file and start from 0666.
  Write a temp in the same directory and `Rename` it into place.
  The PDF path is `installOwnerOnly`.
  The harness rule names only the database and `config.toml`.
- A derived display column subtracts the `%.1f` strings the
  adjacent cells already paint. Converting a kilogram difference
  disagrees with those cells: 80.0 kg and 79.9 kg show as `+0.3`
  pounds. A test that the first day is `0.0` only in kilograms
  hides `ApplyTrend` rounding. At 176.5 lb that day is `-0.1`.
- A zero on-screen loss prints `0.0`. `FitRate`'s unary minus
  turns an exact zero slope into a negative zero, and `%.1f`
  prints `-0.0`.

## ARCH_DECISIONS

- Decision: store body weight in kilograms; kg/lb/st are display-only
  in `config.toml`. Reason: a unit change must not rewrite history.
  Alternatives: store pounds (rejected — SI and meal-planning grams);
  store the display unit in each row (rejected — mixed series).
- Decision: do not persist trend. Reason: a backdated weight edit would
  desync a stored moving average. `ApplyTrend` is the source of truth.
- Decision: On-screen Monthly Loss for the current month is first
  minus last of the last seven trend values. Every other month and
  every long-term window uses a straight-line fit, and the printed
  loss is the daily rate times the gap from the first trend day to
  the last. Daily Deficit is that unrounded daily rate in pounds
  times 3500. The PDF keeps the pencil identity: first minus last
  trend of the plotted span, divided by the days in the span.
  Neither number is stored.
- Decision: config file under XDG, database under XDG data, not
  `~/.hdtools`. Reason: spec-compliant Unix paths; config is
  hand-editable and must not live in SQLite.
- Decision: `[colors]` fail-open, `display_unit` fail-closed. Reason:
  a bad color cannot corrupt the log series; a bad unit would store
  pounds as kilograms. Invalid or omitted colors are dropped silently
  and the TUI uses the built-in default.

## TEST_STRATEGY

Test layout, fixtures, and TUI test rules live in `CONVENTIONS.md`.

## DESIGN_DECISIONS

- Daily log columns are fixed: weight, sleep, steps, workout, note.
  No extras bag until `idea.md` says otherwise.
- Workout is a boolean, not an exercise rung.
- Default display unit is kg until `config.toml` says otherwise.
- `[colors]` is a sparse overlay of named 16-color or hex values.
  16-color is the floor and downshift target. `NO_COLOR` strips
  chromatic color only; bold, reverse, and `>` remain.
