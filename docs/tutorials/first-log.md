# Log your first day

This lesson walks you through running hdtools and saving one daily log.
Follow it in order. You do not need to configure anything first.

You need [Go](https://go.dev/dl/) 1.27 or later, and this repository.

## Run the program

From the repository root:

```bash
go run ./cmd/hdtools
```

The first time, hdtools creates a config file and an empty log database
under your XDG directories. You should see the daily log list and the
line `(no entries yet)`.

## Add a day

1.  Press `n`. The day form opens.
2.  Leave the date as today, or type a date as `YYYY-MM-DD`.
3.  Tab to **weight** and type a number (kilograms unless you have
    already changed display units).
4.  Optionally fill sleep (hours), steps, and toggle workout with space.
5.  Press Enter to save.

You should be back on the list. The new day is there. If you entered a
weight, a **trend** number appears in the next column. That number is
computed from the series; you do not type it.

## Look around, then quit

*   Arrow keys move the selection. `>` marks the selected row.
*   Press `m` for the monthly sheet of the selected day, then Esc to
    return.
*   Press `c` for that month's chart (daily marks and trend), then Esc
    to return.
*   Press `q` to quit.

Your log is in the SQLite file created on first run. Where that file
lives, and how to point at another one, is in
[CLI reference](../reference/cli.md) and
[How to use another database](../how-to/database-path.md).
