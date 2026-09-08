# hdtools

Hacker's Diet tools: a TUI for weight monitoring and meal planning, inspired by [The Hacker's Diet](https://www.fourmilab.ch/hackdiet/e4/welcome.html).

[![Harness](https://img.shields.io/badge/Harness-9%2F14_enforced-CA8A04?style=flat-square)](HARNESS.md)
[![Agent Harness Enabled](https://img.shields.io/badge/Agent_Harness-Enabled-000000?style=flat-square)](HARNESS.md)
[![AI Literacy](https://img.shields.io/badge/AI_Literacy-Level_3-20B2AA?style=flat-square)](assessments/2026-09-08-assessment.md)

Product notes live in [`idea.md`](idea.md). The living harness is [`HARNESS.md`](HARNESS.md).

## How to run

Needs [Go](https://go.dev/dl/) 1.27 or later.

```bash
go run ./cmd/hdtools
```

First run writes `$XDG_CONFIG_HOME/hdtools/config.toml` (typically `~/.config/hdtools/config.toml`) and opens SQLite at `$XDG_DATA_HOME/hdtools/hdtools.db` (typically `~/.local/share/hdtools/hdtools.db`).

```bash
go run ./cmd/hdtools -db /path/to/logs.db
go run ./cmd/hdtools -config /path/to/config.toml
```

`$HDTOOLS_DB` and `$HDTOOLS_CONFIG` override the same paths. Weight is stored in kilograms. Display units are `display_unit` in the config file: `kg` (default), `lb`, or `st`.

| Key | Where | Action |
| --- | --- | --- |
| `n` | list | new day (form) |
| Enter | list | edit selected day |
| `m` | list | monthly sheet |
| `q` | list or month | quit |
| Tab | form | next field |
| Enter | form | save |
| Esc | form or month | cancel / back to list |
| Arrows | month | move cells |
| Type | month | edit cell |
| Space | month | toggle workout |
| `[` / `]` | month | previous / next month |
