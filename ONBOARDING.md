<!-- Generated from HARNESS.md, AGENTS.md, and REFLECTION_LOG.md.
     Do not edit directly — regenerate with /harness-onboarding. -->

# Welcome to hdtools

hdtools is a terminal UI for daily and monthly weight logs,
inspired by
[The Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/welcome.html).
It tracks and analyses. It does not diagnose or treat. You will
work in Go. Bubble Tea draws the screen and SQLite holds the log.
Product intent lives in `idea.md`. A behaviour change also gets a
dated spec under `docs/superpowers/specs/` before the code.
Expect a living harness: formatting, tests, and secret scanning
run on every pull request, and a new Go file opens with a short
note on why it exists.

---

## Tech Stack

- **Go 1.27.0** (`github.com/jglueckstein/hdtools`). One module at
  the repo root. `go build ./...` and `go test ./...` are the
  toolchain. There are no containers. Run the binary, or
  `go run ./cmd/hdtools`, on your machine.
- **Charm v2.** Import `charm.land/bubbletea/v2`,
  `charm.land/lipgloss/v2`, and `charm.land/bubbles/v2`.
  `cmd/hdtools` opens the store and runs `internal/tui`. Charts
  use `github.com/NimbleMarkets/ntcharts/v2`. Call
  `tea.NewProgram` with the model only. Do not pass alternate
  screen, mouse capture, or bubblezone.
- **Colour.** The palette is config `[colors]`: 16-color names,
  `0`–`15`, or hex. Hex downshifts on a 16-color terminal.
  `NO_COLOR` strips chromatic colour. Bold, reverse, and `>`
  stay. A bad colour is dropped. A bad `display_unit` fails the
  load.
- **SQLite** through `database/sql` and `modernc.org/sqlite`, with
  no CGO. The default file is
  `$XDG_DATA_HOME/hdtools/hdtools.db` (usually
  `~/.local/share/hdtools/hdtools.db`). `-db` or `$HDTOOLS_DB`
  overrides that path. Weight is stored in kilograms. There is no
  remote database.
- **Config in TOML** at
  `$XDG_CONFIG_HOME/hdtools/config.toml`. `display_unit` is `kg`
  until the file says `lb` or `st`. Preferences never go in
  SQLite. Monthly chart PDFs use `github.com/go-pdf/fpdf` and
  default to the XDG data directory.

---

## How We Write Code

The day-to-day copy of these rules is `CONVENTIONS.md`. Each
section there names the check that covers it.

**Names.** Follow the main points of Google's Go style. Exported
names are MixedCaps. Unexported names are mixedCaps. Initialisms
stay a whole word in caps: `HTTPServer`, `userID`. Do not prefix
a new API with `Get`. Packages are short, lowercase, and have no
underscores. Files are lowercase with no underscores except
`_test.go`. `context.Context` is the first parameter when a
function needs one. Tests use package `testing` only.

**Layout.** `cmd/hdtools/main.go` is wiring: flags, paths, open
the store, run the TUI. Application code lives under `internal/`
by concern (`config`, `dailylog`, `tui`, `units`, and the chart
packages). Trend stays in `dailylog`. Tests sit next to the code.
If a type is the package's main export, the file is named after
that type. `dailylog` does not import `tui`, `config`, or
`units`. `units` does not import `tui`, `dailylog`, or `config`.
`config` may import `units` and does not import `tui` or
`dailylog`. The TUI does not import `database/sql` or a SQLite
driver. A non-test `.go` file stays at or under 500 lines.

When the same code is already wrong in two places, extract a
helper. `internal/tui/layout.go` (`visPad`) is the model. That
is guidance, not a merge check.

**Errors.** Handle the error, or wrap it with `%w` and a short
context (`fmt.Errorf("load daily log: %w", err)`). A bare
`return err` is not allowed. `panic` is only for an impossible
programmer mistake, never for I/O or SQLite. Compare sentinels
with `errors.Is` or `errors.As`.

**Comments.** A new or rewritten `.go` file, other than a test,
opens with why it exists, the decisions it made, and what it
does not do. Function comments explain the reasoning. An inline
comment explains why, not the next line. Every exported name has
a doc comment whose first sentence says what it does or returns.
A comment that only restates the signature counts as missing.
Each package has a package comment.

**Docs and shell.** User-facing docs follow Diátaxis under
`docs/`. The README is the map. Habitat files such as this one
stay outside that split. User-facing Markdown and shell scripts
follow Google's style for those languages, including an
80-column wrap. Go has no fixed line length. `gofmt` is the
format.

---

## What's Enforced

### At commit time

These are fast checks. CI also runs them on the pull request.

- **Formatting.** Every `.go` file must match `gofmt`.
  `gofmt -l .` listing any path is a failure even though the
  command exits 0. Use `test -z "$(gofmt -l .)"`.
- **No secrets.** gitleaks must not find keys, tokens, or
  passwords in the tree.

### At PR time

GitHub Actions (Harness Constraints and AI Literacy) must be
green. A red check blocks the merge.

**CI gates:**

- **Tests.** `go test ./...` with zero failures. That run also
  covers private `0600` files and the XDG default paths.
- **Preferences stay out of SQLite.** No display unit or colour
  fields in `internal/dailylog`.
- **The TUI has no SQL imports.** Missing days are
  `dailylog.ErrNotFound`.
- **No skip-to-green.** Do not call `t.Skip` or comment out a
  test to make CI pass. `testing.Short()` and a real
  `GOOS`/`GOARCH` skip are fine.
- **No live log in git.** Do not commit a real database or a
  machine `config.toml`. A small fixture under `testdata/` is
  allowed.
- **Import direction.** Domain and config packages do not import
  upward into the TUI.
- **File size.** No non-test `.go` file is over 500 lines.
- **Conventions stay bound.** Every section of `CONVENTIONS.md`
  names a real harness check, or says it is not a merge gate.
- **Go names.** Non-test `.go` files and package names are
  lowercase with no underscores. An external test package may
  be the directory name plus `_test`.
- **Go layout.** `cmd/hdtools/main.go` exists. Production Go
  lives under `cmd/hdtools` or `internal/`. `scripts/` may hold
  check tools. A test is `*_test.go` in those trees.
- **Wrapped errors.** Non-test Go under `cmd/` and `internal/`
  does not return the identifier `err` unwrapped.
- **Exported docs.** Every exported function, type, and const
  has a doc comment.
- **File preamble.** Every non-test `.go` file opens with a
  comment, before `package` or between `package` and the
  imports.

**Reviewed by an agent, not by a CI job today:**

Medical copy is the remaining agent pull-request constraint.
`harness.yml` does not run harness-enforcer, so a green build
does not prove it. Still follow it. User-facing copy in the
TUI, the README, this file, and `idea.md` must not present
hdtools as diagnosing, treating, prescribing, or giving medical
advice. Paraphrase counts. Trend and meal planning are product
behaviour, not a violation. The build does not fail on this
constraint.

### On schedule

Weekly garbage collection looks for stale docs, dependency lag,
convention drift, dead code, a working gitleaks install, import
direction, files over 500 lines, and a coverage drop. Every GC
rule is auto-fix false, so a finding becomes an issue. A monthly
`/harness-health` snapshot is the observability cadence. Quarterly
work is `/harness-audit`, `/assess`, and a cost capture when you
have numbers. Promote reflections about monthly.

Import direction and the 500-line limit also run on the pull
request, so the weekly pass is a second look, not the only one.

---

## Common Pitfalls

- **`gofmt -l .` looks green when it is not.** It exits 0 even
  when it prints paths. Treat any listed file as a failure.
- **Trend needs the full series.** `ApplyTrend` must see every
  day in order. Loading only the current month drops
  carry-forward from the previous month. A chart of the current
  calendar month ends at today. Do not plot carried trend into
  the future.
- **Display units are for the screen.** Use `units.FromKG` in
  the TUI. Save through `units.ToKG`. SQLite always stores
  kilograms.
- **Colour does not ask the terminal.** `View()` emits chroma
  unless `NO_COLOR` is set. A test that needs colour sets
  `NO_COLOR` to empty with `t.Setenv`. Hex is truecolor
  `38;2;` and downshifts on a 16-color profile. Pad columns
  with `lipgloss.Width` (`visPad`), never byte length, once
  colour codes are in the string.
- **The default branch is `master`.** Stay on the feature branch
  until the squash-merge. Keep `CHANGELOG.md` under Unreleased.
  Commit a reflection on that branch before the merge. An
  integration prompt that checks out `main` and writes a dated
  changelog is wrong for this repo. `gh pr checks` can say no
  checks are reported while the runs are still queued. Wait for
  both required jobs on the current head commit.
- **Garbage collection does not edit the tree.** A name that
  looks unused inside one file may have a caller in another
  package. Search the module before deleting it. `plotMonthly`
  in `braille.go` is unreferenced on purpose.

---

## Architecture Decisions

These are already decided. Do not reopen one without a dated spec.

- **Store weight in kilograms.** kg, lb, and st are display-only
  in `config.toml`. A unit change must not rewrite history.
  Storing pounds was rejected. Storing the display unit on each
  row was rejected, because the series would mix units.
- **Do not persist trend.** A backdated weight edit would desync
  a stored moving average. `ApplyTrend` is the source of truth
  after every read.
- **On-screen loss uses the Online rate.** The current month is
  the first minus the last of the last seven trend values. Every
  other month and every long-term window uses a straight-line
  fit, and the printed loss is the daily rate times the gap from
  the first trend day to the last. Daily deficit is that
  unrounded daily rate in pounds times 3500. The PDF keeps the
  pencil identity: first minus last trend of the plotted span,
  divided by the days in the span. Neither number is stored.
- **Config under XDG, database under XDG data**, not
  `~/.hdtools`. Config is hand-editable and must not live in
  SQLite.
- **Colours fail open. `display_unit` fails closed.** A bad
  colour cannot corrupt the log. A bad unit would store pounds
  as kilograms.

Related product choices: daily columns are fixed (weight, sleep,
steps, workout, note). Workout is a boolean. The first-run
display unit is kg. `[colors]` is a sparse overlay. 16-color is
the floor. `NO_COLOR` strips chromatic colour only.

---

## How We Test

The layout rules live in `CONVENTIONS.md`. Run `go test ./...`.

Tests sit next to the code as `_test.go`. Use `t.TempDir()` for
SQLite. `:memory:` races under `t.Parallel` because it is one
database per connection. Do not call `t.Parallel` in a test that
calls `freezeToday`, `t.Chdir`, or `t.Setenv`.

Table-driven tests fit validation and unit conversion (`units`,
`dailylog` constructors). There is no assertion library.

TUI tests do not need a real terminal. Drive `Update` and `View`
with messages. Alignment tests strip ANSI and compare where the
header and value columns end (`internal/tui/layout_test.go`).

The pull-request scripts under `scripts/` are safe to run
locally: Go names, layout, wrapped errors, exported docs, file
preambles, import direction, file size, skip-to-green, live
logs, TUI SQL imports, preferences out of SQLite, and the
conventions binding.

---

## How the Harness Works

- **Advisory loop.** There are no edit-time hooks yet. Format
  and secrets still run in CI.
- **Strict loop.** A pull request runs
  `.github/workflows/harness.yml` and `ai-literacy.yml`. A red
  check blocks the merge. That is the test, `gofmt`, gitleaks,
  and script set above. Medical copy is still an agent review.
  The build does not fail on it.
- **Investigative loop.** Weekly GC looks for slow drift. It
  does not rewrite the code. A finding becomes an issue.

Snapshots are monthly (`/harness-health`). Audits, literacy
assessments, and cost capture are quarterly. Reflection review
is monthly. After a session that changes behaviour, run
`/reflect` and commit the fragment on the feature branch.
`REFLECTION_LOG.md` is generated. Do not edit it by hand.

---

## Your First PR Checklist

1. Branch off `master` with a hyphenated name. Do not commit on
   `master`.
2. If behaviour changes, update `idea.md` and add
   `docs/superpowers/specs/YYYY-MM-DD-<slug>.md` before the code.
   Write a failing test and confirm it is red.
3. Run `gofmt -w .`, then `test -z "$(gofmt -l .)"`.
4. Run `go test ./...`.
5. New `.go` files, other than tests, get a narrative preamble.
   Wrap every returned error. Keep SQL out of `internal/tui`.
   Keep non-test files at or under 500 lines, with no
   underscores in the name.
6. Do not commit a `*.db` or a machine `config.toml`.
7. Do not claim diagnosis, treatment, prescription, or medical
   advice in TUI chrome, help, the README, this file, or
   `idea.md`.
8. Add a line under Unreleased in `CHANGELOG.md`.
9. Push the branch and open a pull request. Wait until Habitat
   checks and Enforce PR constraints succeed on that head
   commit.

---

## Where to Learn More

- [HARNESS.md](HARNESS.md) — conventions, constraints, and GC
- [CONVENTIONS.md](CONVENTIONS.md) — the coding page agents read
- [AGENTS.md](AGENTS.md) — gotchas and decisions
- [REFLECTION_LOG.md](REFLECTION_LOG.md) — session learnings,
  generated from `reflections/active/`
- [idea.md](idea.md) — product backlog
- [README.md](README.md) — how to run the TUI, and the doc map
- [assessments/2026-09-08-assessment.md](assessments/2026-09-08-assessment.md)
  — latest AI literacy assessment
