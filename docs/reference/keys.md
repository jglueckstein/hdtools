# Keyboard

| Key | Screen | Action |
| --- | --- | --- |
| `n` | list | new day (form) |
| `f` | list | open today's meal screen |
| Enter | list | edit selected day |
| `m` | list | monthly sheet |
| `c` | list or month (not editing) | monthly chart |
| `l` | list, month (not editing), or monthly chart | long-term chart |
| `p` | monthly chart | write this month's PDF in `pdf_dir` or the data directory |
| `P` | month (not editing) | write `YYYY-MM-log.pdf` in `pdf_dir` or the data directory |
| `q` | list, month, chart, or meal | quit |
| Tab | form | next field |
| Enter | form | save |
| Esc | form, month, chart, or meal | cancel / back |
| Arrows | month (not editing) | move cells |
| Enter | month (not editing) | open the day form |
| Enter, Down | month (editing) | accept cell and move down the column |
| Up | month (editing) | accept cell and move up the column |
| Left, Right | month (editing) | move the caret |
| Tab | month | accept cell (if editing) and move to the next |
| `t` | list | select the log closest to today |
| `t` | month (not editing) | today's month and day, weight column |
| Type | month | edit cell |
| Space | month | toggle workout |
| `[` / `]` | month or monthly chart | previous / next month |
| `[` / `]` | long-term chart | previous / next kind |
| `a` | meal | add a food line |
| Enter | meal | edit grams of the selected line |
| `e` | meal | edit that food; saving changes every meal date that uses it |
| `x` | meal | remove the selected line |
| `[` / `]` | meal | previous / next meal day |
| `Up`, `Down` | meal | move the selected line |
| Tab | meal form | next field |
| Enter | meal form | save |
| Esc | meal form | cancel |

The selected list row, month cell, or form field is marked with `>`.

Idle `p` on the month sheet still starts an edit. `P` while
editing is text. Idle `P` writes the filled log PDF for the month
on screen, including after `[` and `]`.
