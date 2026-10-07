# How to export a monthly chart PDF

The monthly chart PDF is the on-screen monthly chart on paper: title
box, daily marks, stems, trend, Y margin, and the same Monthly Loss
and Daily Deficit copy. It is not a screenshot of the TUI.

## From the command line

```bash
hdtools -chart-pdf 1990-11
```

writes `1990-11-chart.pdf` under `$XDG_DATA_HOME/hdtools` (otherwise
`~/.local/share/hdtools`) and does not start the TUI. That is the
default database directory. `-db` and `$HDTOOLS_DB` do not move the
PDF. The command prints that path as one line. Set `pdf_dir` in the
config file to choose another directory. An absolute path is used as
written. A value that begins with `~/` uses the home directory. Any
other relative `pdf_dir` fails the export and writes nothing. A
directory that is not absolute after that resolution fails the same
way. `-o path` names the file and ignores `pdf_dir`. `-config` still
selects the config file. A failed export prints no path.

A bad month (`1990-13`, `banana`) exits non-zero and writes no file.

## From the TUI

On the monthly chart (`c`), press `p`. The file `YYYY-MM-chart.pdf` is
written in that same directory. The status line shows the path. `p`
does not write from the list, month sheet, or long-term chart.

The current month ends at today, matching the on-screen chart. An
empty month (no daily marks and no carried trend) still writes a
one-page PDF that says `(empty month)`.

The PDF is always in colour. `NO_COLOR` does not apply. Copy is the
book's 3500 kcal-per-pound identity; it is not medical advice.
