# Config file

Path: `$XDG_CONFIG_HOME/hdtools/config.toml` (typically
`~/.config/hdtools/config.toml`). Override with `-config` or
`$HDTOOLS_CONFIG`. First run writes `display_unit = "kg"` only. Mode
`0600`.

Unknown keys are ignored.

## `display_unit`

| Value | Meaning |
| --- | --- |
| `kg` | kilograms (default) |
| `lb` | pounds |
| `st` | stone |

Any other value fails the load. Storage is always kilograms.

## `pdf_dir`

Optional directory for chart PDFs. Omitted, empty, or whitespace
means `$XDG_DATA_HOME/hdtools` (otherwise `~/.local/share/hdtools`).
That is where the database file lives when `-db` and `$HDTOOLS_DB`
are unset. Those two do not move the PDF. An absolute path is used
as written. A value that begins with `~/` uses the home directory
for that prefix. Any other relative value fails the export and
writes nothing. A directory that is not absolute after that
resolution fails the export and writes nothing. `$` is not expanded.
A value that is not a string does not fail the load; an export that
would use it fails and names `pdf_dir`. `-o` is a file path and
ignores this key. The first-run file does not contain `pdf_dir`.

## `[colors]`

Optional table. Keys are roles. Values are a 16-color name (case
insensitive; hyphens and spaces are equivalent), an integer `0`–`15`,
or hex `#rrggbb` / `#rgb`. Unknown keys are ignored. An omitted or
invalid value for a role uses that role's default. A `colors` value
that is not a table is treated as a missing table. Fallback is silent.

| Key | Where it appears | Default |
| --- | --- | --- |
| `title` | Screen titles and form labels | `cyan` |
| `muted` | Database path, empty-state, form hints | `bright-black` |
| `header` | Column headers | `white` |
| `help` | Keybinding lines | `bright-black` |
| `weight` | Daily weight values | `blue` |
| `trend` | Trend values | `red` |
| `delta-pos` | Weight above trend | `yellow` |
| `delta-neg` | Weight below trend | `green` |
| `delta-zero` | Weight on trend | `white` |
| `selection` | Extra foreground on reverse video (list row or focused month cell). Omitted: reverse only. Unused on the form. | none |
| `error` | Error lines | `red` |
| `status` | Status lines | `green` |

16-color names: `black`, `red`, `green`, `yellow`, `blue`, `magenta`,
`cyan`, `white`, `bright-black` (`gray`, `grey`), `bright-red`,
`bright-green`, `bright-yellow`, `bright-blue`, `bright-magenta`,
`bright-cyan`, `bright-white`.

Hex on a truecolor terminal is used as given. On a 16-color terminal it
maps to the nearest of those sixteen. Eight-digit hex and hex without
`#` are invalid.

Bold (title, header, trend, focused form label) and reverse (selection)
are not scheme keys.

Writing the config persists only roles that were present and valid.
First-run `Ensure` does not emit `[colors]`.
