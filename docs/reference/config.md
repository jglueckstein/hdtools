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

## `[colors]`

Optional table. Keys are roles. Values are a 16-color name (case
insensitive; hyphens and spaces are equivalent), an integer `0`–`15`, or
hex `#rrggbb` / `#rgb`. Unknown keys are ignored. An omitted or invalid
value for a role uses that role's default. A `colors` value that is not
a table is treated as a missing table. Fallback is silent.

| Key | Where it appears | Default |
| --- | --- | --- |
| `title` | Screen titles and form labels | `cyan` |
| `muted` | Database path, empty-state, form hints | `bright-black` |
| `header` | Column headers | `white` |
| `help` | Keybinding lines | `bright-black` |
| `weight` | Daily weight values | `blue` |
| `trend` | Trend values | `red` |
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
