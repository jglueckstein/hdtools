# How to export a filled monthly log sheet

The filled log sheet is one portrait page of the named month. Every
calendar day is a row, including days after today. The columns are
the month sheet's: day, weekday, weight, trend, delta, sleep, steps,
workout, and note. A missing cell has no mark. A logged zero still
prints.

## From the command line

```bash
hdtools -log-pdf 1990-11
```

writes `1990-11-log.pdf` under `$XDG_DATA_HOME/hdtools` (otherwise
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

Setting `-log-pdf` and `-chart-pdf` together exits non-zero before
config is read and before a database is opened. Neither PDF is
written. An invalid month does not change that.

A month with no weigh-in still prints. The page has no Monthly Loss
and no Daily Deficit. It is not medical advice. There is no TUI key
for this page.
