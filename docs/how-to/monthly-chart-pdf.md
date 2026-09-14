# How to export a monthly chart PDF

The monthly chart PDF is the on-screen monthly chart on paper: title
box, daily marks, stems, trend, Y margin, and the same Monthly Loss
and Daily Deficit copy. It is not a screenshot of the TUI.

## From the command line

```bash
hdtools -chart-pdf 1990-11
```

writes `1990-11-chart.pdf` in the current directory and does not start
the TUI. `-o path` sets the output path. `-db` and `-config` work as
usual.

A bad month (`1990-13`, `banana`) exits non-zero and writes no file.

## From the TUI

On the monthly chart (`c`), press `p`. The file `YYYY-MM-chart.pdf` is
written in the current directory. The status line shows the path. `p`
does not write from the list, month sheet, or long-term chart.

The current month ends at today, matching the on-screen chart. An
empty month (no daily marks and no carried trend) still writes a
one-page PDF that says `(empty month)`.

The PDF is always in colour. `NO_COLOR` does not apply. Copy is the
book's 3500 kcal-per-pound identity; it is not medical advice.
