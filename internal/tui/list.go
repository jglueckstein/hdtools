package tui

// The daily list is the weight log a meal day leaves and returns to.
// f is on this help line because it is the door off the list. A list
// that failed to load still offers only quit, so a database error is
// not a meal. This file does not open foods or move a meal date.

import (
	"fmt"
	"strings"

	"github.com/jglueckstein/hdtools/internal/units"
)

func (a *App) listView() string {
	p := a.pal
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", p.title.Render(fmt.Sprintf("hdtools — daily log  (weight %s)", a.cfg.DisplayUnit)))
	fmt.Fprintf(&b, "%s\n\n", p.muted.Render("db: "+a.dbPath))
	if a.err != nil {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.error.Render("error: "+a.err.Error()), p.help.Render("q quit"))
		return b.String()
	}
	if len(a.logs) == 0 {
		fmt.Fprintf(&b, "%s\n\n%s\n", p.muted.Render("(no entries yet)"), p.help.Render("n new day   m month   l long   f meal   t today   q quit"))
		return b.String()
	}
	fmt.Fprintf(&b, "%s\n", p.header.Render(listHeader()))
	for i, log := range a.logs {
		mark := "  "
		if i == a.cursor {
			mark = "> "
		}
		weight := "—"
		trend := "—"
		if log.Weight != nil {
			w, err := units.FromKG(*log.Weight, a.cfg.DisplayUnit)
			if err == nil {
				weight = fmt.Sprintf("%.1f", w)
			}
		}
		if t, err := units.FromKG(log.Trend, a.cfg.DisplayUnit); err == nil && (log.Weight != nil || log.Trend != 0) {
			trend = fmt.Sprintf("%.1f", t)
		}
		workout := "no"
		if log.Workout {
			workout = "yes"
		}
		selected := i == a.cursor
		showTrend := log.Weight != nil || log.Trend != 0
		delta := formatDelta(log.Weight, log.Trend, showTrend, a.cfg.DisplayUnit)
		var weightCell, trendCell, deltaCell string
		if selected && p.selectionFG {
			weightCell = visPad(weight, wWeight, true)
			trendCell = visPad(trend, wTrend, true)
			deltaCell = visPad(delta, wDelta, true)
		} else {
			weightCell = visPad(p.weight.Render(weight), wWeight, true)
			trendCell = visPad(p.trend.Render(trend), wTrend, true)
			deltaCell = styleDelta(delta, p)
		}
		line := visPad(mark, wMark, false) + joinCols(
			visPad(log.Day.Format("2006-01-02"), wDate, false),
			weightCell,
			trendCell,
			deltaCell,
			visPad(fmt.Sprintf("%.1f", log.SleepHours), wSleep, true),
			visPad(fmt.Sprintf("%d", log.Steps), wSteps, true),
			visPad(workout, wWorkout, false),
			log.Note,
		)
		if selected {
			line = p.selected.Render(line)
		}
		fmt.Fprintf(&b, "%s\n", line)
	}
	if a.status != "" {
		fmt.Fprintf(&b, "\n%s\n", p.status.Render(a.status))
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render("n new   enter edit   m month   c chart   l long   f meal   t today   q quit"))
	return b.String()
}
