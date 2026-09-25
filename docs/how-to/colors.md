# How to change TUI colours

Colours live in the same config file as display units. Edit the file and
restart; you do not rebuild.

1.  Open typically `~/.config/hdtools/config.toml`.
2.  Add or edit a `[colors]` table. Each key is a role; each value is a
    16-color name, `0`–`15`, or hex (`#rrggbb` or `#rgb`):

    ```toml
    [colors]
    weight = "green"
    trend = "#c22"
    title = "magenta"
    delta-pos = "yellow"
    delta-neg = "green"
    delta-zero = "white"
    ```

3.  Save and restart hdtools.

Omitted roles keep the built-in default (blue daily weight, red trend,
yellow / green / white for the weight−trend delta).
A typo in one role falls back to that default and does not prevent the
log from opening.

To turn chromatic colour off for a session:

```bash
NO_COLOR=1 go run ./cmd/hdtools
```

Headers, the `>` mark, bold trend, and reverse selection stay. The full
list of roles and accepted values is in
[config reference](../reference/config.md).
