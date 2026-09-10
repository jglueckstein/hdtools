# Command line

```text
hdtools [-db path] [-config path]
```

| Flag | Environment | Default |
| --- | --- | --- |
| `-db` | `$HDTOOLS_DB` | `$XDG_DATA_HOME/hdtools/hdtools.db` (fallback `~/.local/share/hdtools/hdtools.db`) |
| `-config` | `$HDTOOLS_CONFIG` | `$XDG_CONFIG_HOME/hdtools/config.toml` (fallback `~/.config/hdtools/config.toml`) |

A non-empty flag wins over the environment variable. Missing
directories are created. A missing config file is created with
`display_unit = "kg"`. An invalid `display_unit` is a startup error.
Invalid `[colors]` values are not.

`$NO_COLOR`, when present and non-empty, disables chromatic colour for
that process. It is not a config key.
