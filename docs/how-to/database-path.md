# How to use another database

By default hdtools opens `$XDG_DATA_HOME/hdtools/hdtools.db` (typically
`~/.local/share/hdtools/hdtools.db`).

To open a different file for this run:

```bash
go run ./cmd/hdtools -db /path/to/logs.db
```

Or set `$HDTOOLS_DB` to the same path. The flag wins if both are set.

The parent directory is created if needed. Config stays in
`config.toml`; it is not stored in the database.

A remote database is not implemented yet.
