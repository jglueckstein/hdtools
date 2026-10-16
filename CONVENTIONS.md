# Coding conventions

Agents read this file before writing or reviewing code. It collects
the standards already declared in `HARNESS.md`, `CLAUDE.md`,
`AGENTS.md`, and `idea.md`. `HARNESS.md` is what CI and review
check. `AGENTS.md` keeps gotchas and decisions, and points here.
`idea.md` is the product backlog.

Each section starts with `Checked by`. Those names are `###`
headings under Constraints in `HARNESS.md`. A section that has no
merge check says `not a merge gate` and nothing else. A long list
continues on the next lines, indented two spaces.

The Context conventions in `HARNESS.md` are named here. Naming is
Go. File structure is Layout. Error handling is Errors.
Documentation is Comments, Documentation, and Shell.

## Go

Checked by: Consistent formatting; Idiomatic Go naming; Tests must pass

Go is 1.27.0. Source follows the main points of
[Google's Go Style Guide](https://google.github.io/styleguide/go/guide).

- `gofmt` is the format. There is no fixed line length. The
  80-column wrap used for Markdown and shell does not apply to
  `.go` files. A long line is refactored rather than wrapped by
  hand.
- `gofmt -l .` exits 0 even when it prints paths. A listed path is
  a failure: `test -z "$(gofmt -l .)"`.
- Exported identifiers are MixedCaps. Unexported identifiers are
  mixedCaps. Initialisms stay all-caps as words: `HTTPServer`,
  `userID`, `ServeHTTP`. New APIs do not use a `Get` prefix.
- Packages are short, lowercase, and contain no underscores.
  Files are lowercase with no underscores except `_test.go`.
- `context.Context` is the first parameter when a function needs
  a context.
- Tests use package `testing` only. No assertion library.
- `panic` is only for unreachable programmer errors, never for
  I/O, parse, or database failures. That rule is also checked
  under Errors.

Build and test:

```bash
go build ./...
go test ./...
gofmt -w .
test -z "$(gofmt -l .)"
```

## Layout

Checked by: Standard Go layout; TUI has no SQL imports;
  Import direction; File size

One module at the repo root. `cmd/hdtools` is wiring only: flag
resolution, paths, open the store, run the TUI. Application code
lives under `internal/` and is not importable from outside the
module. One package directory per concern (`config`, `dailylog`,
`tui`, `units`, and the chart packages). Weight trend stays in
`dailylog`. If a type is the package's main export, its file is
named after that type.

Import direction, checked by `scripts/check-import-direction.sh`:

- `internal/dailylog` does not import `tui`, `config`, or `units`.
- `internal/units` does not import `tui`, `dailylog`, or `config`.
- `internal/config` may import `units`. It does not import `tui`
  or `dailylog`.
- `cmd/hdtools` and `internal/tui` import downward.
- `internal/tui` does not import `database/sql` or a SQLite
  driver. The command and tests open the database and inject a
  store.

A non-test `.go` file stays at or under 500 lines. Extract a
helper when the same code is already wrong in two places.
`internal/tui/layout.go` (`visPad`, shared column widths) is the
model. A framework for a single screen is not. That extraction
is written guidance and is not a merge gate.

## Errors

Checked by: Errors wrapped with context

Handle the error, or wrap it with `%w` and a short context
(`fmt.Errorf("load daily log: %w", err)`). Bare `return err` is
not permitted. Sentinel errors live in the package that defines
them and are compared with `errors.Is` and `errors.As`.

## Comments

Checked by: Literate preamble on new files; Exported API documentation

New or substantially rewritten `.go` files, except `_test.go`,
follow literate programming:

1. Open with a narrative preamble: why the file exists, the key
   design decisions, and what it deliberately does not do.
2. Function comments explain the reasoning, not the signature.
3. Order of presentation follows understanding. Orchestration
   comes before detail.
4. Each file has one concern.
5. Inline comments explain why, not what. A comment that restates
   the next line is prohibited.

Every exported function, type, and const has a doc comment. The
first sentence says what it does or returns, not how. A comment
that only restates the signature counts as missing. Unexported
code is commented only when the reason is non-obvious: a
constraint, a workaround, or a domain rule. Each package has a
package comment, on one file or in `doc.go`.

When reviewing or refactoring, apply CUPID: Composable, Unix
philosophy, Predictable, Idiomatic, Domain-based. CUPID is
written guidance and is not a merge gate.

## Tests

Checked by: No skip-to-green; No live log in git

- Tests sit next to the code as `*_test.go`.
- Validation and conversion matrices are table-driven (`units`,
  `dailylog` constructors).
- SQLite tests use `t.TempDir()`. `:memory:` races under
  `t.Parallel`.
- Do not call `t.Parallel` in a test that calls `freezeToday`,
  `t.Chdir`, or `t.Setenv`. `freezeToday` replaces a
  process-global clock.
- Do not call `t.Skip`, and do not comment out a test, to make
  the suite pass. `testing.Short()` and a genuine `GOOS` or
  `GOARCH` skip are allowed.
- Do not commit a real log database or a machine `config.toml`.
  Small fixtures under `testdata/` are allowed.
- TUI tests do not need a real terminal. Drive `App.Update` and
  `View` with messages.
- Alignment tests strip ANSI and compare header and value column
  ends (`internal/tui/layout_test.go`).
- Tests that need colour set `NO_COLOR` to empty with
  `t.Setenv`. The process does not consult the TTY.

Table-driven tests, `t.TempDir`, the `t.Parallel` ban, driving
`Update` and `View`, alignment checks, and `NO_COLOR` are written
guidance and are not a merge gate.

## TUI

Checked by: Tests must pass

`cmd/hdtools` calls `tea.NewProgram` with the model only. No
`AltScreen`. No `MouseMode`. No file under `cmd/` or `internal/`
imports a path containing `bubblezone`, and the program does not
set a zone manager.

The module requires Charm v2:
`charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, and
`charm.land/bubbles/v2`. Do not require or import the
`github.com/charmbracelet/` v1 module paths. Do not add
`github.com/sajari/regression`.

Pad columns with `lipgloss.Width` (`visPad`), never byte length,
once ANSI colour is in the string. `View()` emits chromatic
colour unless `NO_COLOR` is non-empty. `NO_COLOR` strips
chromatic colour only; bold, reverse, and `>` remain. 16-colour
is the floor and the downshift target. Hex is `38;2;` on
truecolor. `visPad` is written guidance and is not a merge
gate. Charm v2, `tea.NewProgram`, and `bubblezone` are locked by
those tests.

## Paths and files

Checked by: Default paths are XDG; DB and config files are 0600;
  Prefs out of SQLite; Tests must pass

Default config is `$XDG_CONFIG_HOME/hdtools/config.toml`. Default
data is `$XDG_DATA_HOME/hdtools/hdtools.db`. The fallbacks are
`~/.config` and `~/.local/share`, not `~/.hdtools`. `-config`,
`-db`, `$HDTOOLS_CONFIG`, and `$HDTOOLS_DB` override only when
the user named the path. New database files and `config.toml`
files are created mode `0600`. Preferences stay in the config
file, not in SQLite.

Weight is stored in kilograms. Display conversion (`units.FromKG`)
is for the TUI. A save goes through `units.ToKG`. Trend is
`ApplyTrend` over the full chronological series, never a SQLite
column. `ApplyTrend` on the current month alone drops
carry-forward from the previous month. Kilograms and
`ApplyTrend` are locked by those tests.

## Documentation

Checked by: User-facing copy is not medical

User-facing docs follow
[Diátaxis](https://diataxis.fr): tutorials, how-to guides,
reference, and explanation, each in its own page under `docs/`.
The README is the map. Do not mix a lesson with a key table, or
a recipe with a design discussion. Habitat files (`HARNESS.md`,
`ONBOARDING.md`, `AGENTS.md`, `CONVENTIONS.md`,
`docs/superpowers/`) are a different audience and stay outside
that split.

User-facing Markdown follows the main points of
[Google's Markdown Style Guide](https://google.github.io/styleguide/docguide/style.html):

- 80-column wrap, except links, tables, headings, and code.
- ATX headings. One H1.
- Fenced code with a language.
- Nested lists indented 4 spaces.
- No trailing whitespace. Markdown rather than HTML.
- Informative link text.
- No Gitiles `[TOC]`. `../` links are allowed on GitHub.

Diátaxis and the Markdown style are written guidance and are not
a merge gate.

User-facing copy in TUI chrome and help, `README.md`,
`ONBOARDING.md`, and `idea.md` product claims does not present
the program as diagnosing, treating, prescribing, or giving
medical advice. Paraphrase counts. Trend and meal planning are
product features.

## Shell

Checked by: not a merge gate

Shell scripts follow the main points of
[Google's Shell Style Guide](https://google.github.io/styleguide/shellguide.html):

- `#!/bin/bash` and `set -euo pipefail`.
- 2-space indent, no tabs, 80-column wrap.
- `; then` and `; do` on the same line as `if`, `for`, and
  `while`.
- Quoted `"${var}"`. `$(...)` rather than backticks.
- `[[ ... ]]` rather than `[ ... ]`.
- Errors on STDERR. A file header comment says why the script
  exists.
- `scripts/lib/*.sh` are not executable.

ShellCheck is recommended.

## Changes

Checked by: not a merge gate

A behaviour change updates `idea.md` and a dated spec at
`docs/superpowers/specs/YYYY-MM-DD-<slug>.md` before
implementation. `idea.md` is the backlog. The dated file is the
change record.

1. Update `idea.md` when the product intent changed.
2. Add or revise the dated spec.
3. Write a failing test from the spec and confirm it is red.
4. Implement until the test is green.
5. Refactor while the tests stay green.

No production code without a failing test first, except wiring
that has no behaviour of its own (`cmd/hdtools` flag resolution).

Do not commit directly to `master`. Open an issue, use a
hyphenated branch, and merge via a pull request. When a GitHub
remote and pull requests exist, update `CHANGELOG.md` under
Unreleased before the pull request. Do not invent a changelog
for a local-only commit. Commit messages say what changed and
why, with no postamble and no attribution line.

`REFLECTION_LOG.md` is generated from `reflections/active/`.
Do not edit it by hand. Regenerate it with
`scripts/regenerate-reflection-log.sh`.
