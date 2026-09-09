<!-- Generated from HARNESS.md, AGENTS.md, and REFLECTION_LOG.md.
     Do not edit directly — regenerate with /harness-onboarding. -->

# Welcome to hdtools

hdtools is a terminal UI for [The Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/welcome.html): daily and monthly weight logs, a 10% EMA trend, and (later) meal planning and charts. You will work in Go, with Bubble Tea on the screen and SQLite on disk. Product intent lives in `idea.md`; behaviour changes also get a dated spec under `docs/superpowers/specs/`. Expect a living harness: `gofmt`, tests, and gitleaks run on every PR, and new files need a short "why this exists" preamble.

---

## Tech Stack

- **Go 1.27** (`github.com/jglueckstein/hdtools`) — one module at the repo root. `go build ./...` and `go test ./...` are the whole toolchain.
- **Bubble Tea + lipgloss** — the TUI. `cmd/hdtools` opens the store and runs `internal/tui`. Colour is 16-color ANSI so a basic terminal still works.
- **SQLite via `database/sql` and `modernc.org/sqlite`** — no CGO. Default database is `$XDG_DATA_HOME/hdtools/hdtools.db` (usually `~/.local/share/hdtools/hdtools.db`). Weight is stored in kilograms. There is no remote database yet.
- **Config in TOML** — `$XDG_CONFIG_HOME/hdtools/config.toml`. `display_unit` is `kg` (first run), `lb`, or `st`. Prefs never go in SQLite.
- **No containers** — you run the binary (or `go run ./cmd/hdtools`) on your machine.

---

## How We Write Code

**Names.** Exported identifiers are PascalCase; unexported are camelCase. Initialisms stay caps as a word (`HTTPServer`, `userID`). Packages are short and lowercase. File names are one word (`trend.go`) or `snake_case.go` for a phrase.

**Layout.** `cmd/hdtools/main.go` is wiring only: flags, XDG paths, open the store, run the TUI. Application code lives under `internal/` by concern (`config`, `dailylog`, `tui`, `units`). Trend stays in `dailylog`, not its own package. Tests sit next to the code as `*_test.go`. If a type is the package's main export, the file is named after that type.

**Errors.** Handle the error or wrap it with `%w` and a short context (`fmt.Errorf("load daily log: %w", err)`). Bare `return err` is not allowed. `panic` is only for impossible programmer mistakes, never I/O or SQLite. Compare sentinels with `errors.Is` / `errors.As`.

**Comments.** New `.go` files (not `_test.go`) open with a preamble: why the file exists, key decisions, what it deliberately does not do. Function comments explain reasoning, not the signature. Inline comments explain why, not the next line. Every exported symbol has a doc comment whose first sentence says what it does or returns.

**Extract helpers when the same code is already wrong in two places.** `internal/tui/layout.go` (`visPad`) is the model. A layout framework for one screen is not.

---

## What's Enforced

### At commit time

These are fast checks. CI also runs them on the PR.

- **Formatting** — every `.go` file must match `gofmt`. `gofmt -l .` listing any path is a failure even though the command exits 0.
- **No secrets** — gitleaks must not find keys, tokens, or passwords in the tree.

### At PR time

GitHub Actions (Harness Constraints + AI Literacy) must be green.

- **Tests** — `go test ./...` with zero failures.
- **Go naming and layout** — reviewed against the conventions above.
- **Wrapped errors and literate preambles** — same as How We Write Code.
- **Exported docs** — exported symbols need a real first sentence, not a restatement of the signature.
- **Prefs stay out of SQLite** — no `display_unit` / colour-scheme fields in `internal/dailylog`.
- **TUI has no SQL imports** — `internal/tui` must not import `database/sql` or a SQLite driver. Missing days are `dailylog.ErrNotFound`.
- **New DB and config files are `0600`** — owner read/write only.
- **Default paths are XDG** — unless the user passed `-db` / `-config` or the env vars.
- **No skip-to-green** — do not `t.Skip` or comment out a test to make CI pass. `testing.Short()` and real `GOOS`/`GOARCH` skips are fine.
- **No live log in git** — do not commit a real `.db` or a machine `config.toml`. Tiny fixtures under `testdata/` are allowed.

### On schedule

Weekly garbage collection looks for stale docs, dependency lag, convention drift, dead code, a working gitleaks install, import direction (domain packages must not import the TUI), files over 500 lines, and coverage dropping by package. Monthly we take a harness-health snapshot. Quarterly: `/harness-audit`, `/assess`, and a cost capture if you have numbers.

---

## Common Pitfalls

- **`gofmt -l .` looks green when it is not.** It exits 0 even when it prints paths. Treat any listed file as a failure: `test -z "$(gofmt -l .)"`.
- **Trend needs the full series.** `ApplyTrend` must see every day in order. Loading only the current month drops carry-forward from the previous month.
- **Display units are for the screen only.** `units.FromKG` in the TUI; `units.ToKG` before save. SQLite always stores kilograms.
- **TUI columns and ANSI.** Pad with `lipgloss.Width` (`visPad`), never `len()`, once colour codes are in the string.
- **Two `HARNESS.md` files.** Root `HARNESS.md` is the living harness. `.claude/HARNESS.md` is an unfilled template — do not audit or assess against it.
- **Cost tools assume Anthropic/OpenAI.** This project is Grok Build. Use `/usage` and [console.x.ai](https://console.x.ai/team/default/usage) if you need spend; `/cost-capture` will not ground dollar estimates here yet.

---

## Architecture Decisions

The team already decided these — do not reopen them without a dated spec.

- **Store weight in kilograms.** kg/lb/st are display-only in `config.toml`. A unit change must not rewrite history. Storing pounds, or storing the display unit on each row, was rejected.
- **Do not persist trend.** A backdated weight edit would desync a stored moving average. `ApplyTrend` is the source of truth after every read.
- **Config under XDG, database under XDG data**, not `~/.hdtools`. Config is hand-editable and must not live in SQLite. A log file can move machines without dragging prefs.

Related product choices (also in AGENTS.md): daily columns are fixed (weight, sleep, steps, workout, note); workout is a boolean; first-run display is kg; `NO_COLOR` and user colour schemes are specified in `idea.md` but not implemented yet.

---

## How We Test

Tests live beside source as `_test.go`. Run them with `go test ./...`.

Use `t.TempDir()` for SQLite files. `:memory:` races under `t.Parallel` because it is per-connection.

Table-driven tests fit validation and unit conversion (`units`, `dailylog` constructors).

TUI tests do not need a real terminal: drive `App.Update` / `View` with messages. Alignment tests strip ANSI and compare header/value column ends (`internal/tui/layout_test.go`).

---

## How the Harness Works

- **Advisory loop** — we do not have edit-time hooks yet. Format and secrets still run in CI.
- **Strict loop** — pull requests run `.github/workflows/harness.yml` and `ai-literacy.yml`. A red check blocks merge.
- **Investigative loop** — weekly GC rules (docs, deps, import direction, file size, coverage trend). They do not auto-fix architecture.

Observability cadence is **monthly** snapshots (`/harness-health`). Audits and literacy assessments are quarterly. After a behaviour-changing session, run `/reflect`.

This repo's default branch is `master`. Open a hyphenated branch and a PR; do not commit on `master`. Update `CHANGELOG.md` on the PR.

---

## Your First PR Checklist

1. Branch off `master` with a hyphenated name. Do not commit on `master`.
2. If behaviour changes: update `idea.md` and add `docs/superpowers/specs/YYYY-MM-DD-<slug>.md` before the code.
3. `gofmt -w .` then `test -z "$(gofmt -l .)"`.
4. `go test ./...`.
5. New `.go` files (not tests): narrative preamble; wrap every returned error; no `database/sql` in `internal/tui`.
6. Do not commit `*.db` or a machine `config.toml`.
7. Add a line to `CHANGELOG.md` under Unreleased.
8. Push the branch and open a PR. Wait for Harness Constraints and Habitat checks.

---

## Where to Learn More

- [HARNESS.md](HARNESS.md) — conventions, constraints, and GC
- [AGENTS.md](AGENTS.md) — style, gotchas, architecture
- [REFLECTION_LOG.md](REFLECTION_LOG.md) — session learnings (generated from `reflections/active/`)
- [idea.md](idea.md) — product backlog
- [README.md](README.md) — how to run the TUI
- [assessments/2026-09-08-assessment.md](assessments/2026-09-08-assessment.md) — latest AI literacy assessment
