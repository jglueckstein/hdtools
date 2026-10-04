# Command line

```text
hdtools [-db path] [-config path] [-chart-pdf YYYY-MM] [-o path]
```

| Flag | Environment | Default |
| --- | --- | --- |
| `-db` | `$HDTOOLS_DB` | `$XDG_DATA_HOME/hdtools/hdtools.db` (fallback `~/.local/share/hdtools/hdtools.db`) |
| `-config` | `$HDTOOLS_CONFIG` | `$XDG_CONFIG_HOME/hdtools/config.toml` (fallback `~/.config/hdtools/config.toml`) |
| `-chart-pdf` | | write a one-page monthly chart PDF for `YYYY-MM` and exit |
| `-o` | | output file for `-chart-pdf`. When omitted, `YYYY-MM-chart.pdf` is written in `pdf_dir` or `$XDG_DATA_HOME/hdtools` |

A non-empty flag wins over the environment variable. Missing
directories are created. A missing config file is created with
`display_unit = "kg"`. An invalid `display_unit` is a startup error.
Invalid `[colors]` values are not.

A successful `-chart-pdf` prints the path it wrote as one line on
standard output. A failed export prints no path.

`$NO_COLOR`, when present and non-empty, disables chromatic colour for
that process. It is not a config key. It does not grey `-chart-pdf`
output; a PDF is always in colour.
