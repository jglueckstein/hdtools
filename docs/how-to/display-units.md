# How to change display units

Weight is always stored in kilograms. `display_unit` only changes how
the TUI shows it.

1. Open the config file (typically `~/.config/hdtools/config.toml`).
2. Set `display_unit` to `kg`, `lb`, or `st`.
3. Save the file.
4. Restart hdtools.

If the value is not one of those three, hdtools will not start. That is
deliberate: a wrong unit must not be stored as kilograms.

There is no in-app unit picker. After a restart, list, form, and month
sheet all show the new unit.
