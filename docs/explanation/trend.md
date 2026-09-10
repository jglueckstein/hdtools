# Why trend is not a column in SQLite

The daily log shows two numbers for weight: the noisy scale reading, and
a smoother **trend**. The trend is a 10% exponential moving average of
the chronological series, the same idea as in [The Hacker's Diet weight
monitoring](https://www.fourmilab.ch/hackdiet/e4/weightmonitor.html).

It is computed on load (`ApplyTrend`), never written. A stored moving
average would go stale the moment someone edited a backdated weight. The
function needs the full series: loading only the current month would
drop carry-forward from the previous month, so the first days of a sheet
would disagree with the end of the last one.

You cannot type the trend. The TUI shows it in a separate column, in
bold, so it is still distinct when colour is off.
