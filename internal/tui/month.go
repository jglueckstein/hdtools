package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	colWeight = iota
	colSleep
	colSteps
	colWorkout
	colNote
	colCount
)

// sheetDay is one row of the paper monthly log: every calendar day, whether
// or not a database row exists. Trend is taken from the last computed trend
// on or before that day so a blank weigh-in still shows the carry-forward.
type sheetDay struct {
	Day      time.Time
	Log      dailylog.DailyLog
	HasEntry bool
	Trend    float64
	HasTrend bool
}

type monthModel struct {
	year    int
	month   time.Month
	day     int
	col     int
	editing bool
	input   textinput.Model
	err     string
}

func newMonth(day time.Time) monthModel {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Width = 24
	ti.CharLimit = 200
	y, m, d := day.Date()
	return monthModel{year: y, month: m, day: d, input: ti}
}

func (m *monthModel) clamp() {
	max := daysInMonth(m.year, m.month)
	if m.day < 1 {
		m.day = 1
	}
	if m.day > max {
		m.day = max
	}
	if m.col < 0 {
		m.col = 0
	}
	if m.col >= colCount {
		m.col = colCount - 1
	}
}

func (m *monthModel) cursorDay() time.Time {
	return time.Date(m.year, m.month, m.day, 0, 0, 0, 0, time.UTC)
}

func (m *monthModel) prevMonth() {
	t := time.Date(m.year, m.month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	m.year, m.month, _ = t.Date()
	m.clamp()
}

func (m *monthModel) nextMonth() {
	t := time.Date(m.year, m.month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	m.year, m.month, _ = t.Date()
	m.clamp()
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// buildMonthSheet fills every calendar day of year/month from a full
// (already trended) log series.
func buildMonthSheet(logs []dailylog.DailyLog, year int, month time.Month) []sheetDay {
	byDay := make(map[string]dailylog.DailyLog, len(logs))
	for _, log := range logs {
		byDay[log.Day.Format(time.DateOnly)] = log
	}
	n := daysInMonth(year, month)
	out := make([]sheetDay, n)
	for i := 0; i < n; i++ {
		day := time.Date(year, month, i+1, 0, 0, 0, 0, time.UTC)
		row := sheetDay{Day: day}
		if log, ok := byDay[day.Format(time.DateOnly)]; ok {
			row.Log = log
			row.HasEntry = true
			row.Trend = log.Trend
			row.HasTrend = log.Weight != nil || log.Trend != 0
		} else if t, ok := lastTrendOnOrBefore(logs, day); ok {
			row.Trend = t
			row.HasTrend = true
		}
		out[i] = row
	}
	return out
}

func lastTrendOnOrBefore(logs []dailylog.DailyLog, day time.Time) (float64, bool) {
	var trend float64
	found := false
	for _, log := range logs {
		if log.Day.After(day) {
			break
		}
		if log.Weight != nil || log.Trend != 0 {
			trend = log.Trend
			found = true
		}
	}
	return trend, found
}

func (m monthModel) view(sheet []sheetDay, unit units.Unit, dbPath string, status string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "hdtools — %s %d  (weight %s)\n", m.month.String(), m.year, unit)
	fmt.Fprintf(&b, "db: %s\n\n", dbPath)
	fmt.Fprintf(&b, "     date  wd  weight  trend  sleep  steps  workout  note\n")
	for _, row := range sheet {
		mark := "  "
		onRow := row.Day.Day() == m.day
		if onRow {
			mark = "> "
		}
		wd := row.Day.Weekday().String()[:2]
		weight := "     —"
		if row.HasEntry && row.Log.Weight != nil {
			if w, err := units.FromKG(*row.Log.Weight, unit); err == nil {
				weight = fmt.Sprintf("%6.1f", w)
			}
		}
		trend := "    —"
		if row.HasTrend {
			if t, err := units.FromKG(row.Trend, unit); err == nil {
				trend = fmt.Sprintf("%5.1f", t)
			}
		}
		sleep, steps, workout, note := "    —", "    —", "  —", ""
		if row.HasEntry {
			sleep = fmt.Sprintf("%5.1f", row.Log.SleepHours)
			steps = fmt.Sprintf("%5d", row.Log.Steps)
			workout = "  no"
			if row.Log.Workout {
				workout = " yes"
			}
			note = row.Log.Note
		}
		cells := []string{weight, sleep, steps, workout, note}
		if onRow {
			cells[m.col] = highlightCell(cells[m.col], m.editing, m.input)
		}
		fmt.Fprintf(&b, "%s%2d %s %s  %s  %s  %s  %s  %s\n",
			mark, row.Day.Day(), wd, cells[0], trend, cells[1], cells[2], cells[3], cells[4])
	}
	if m.err != "" {
		fmt.Fprintf(&b, "\nerror: %s\n", m.err)
	}
	if status != "" {
		fmt.Fprintf(&b, "\n%s\n", status)
	}
	fmt.Fprintf(&b, "\narrows move   type edit   space workout   enter form   [ ] month   esc list\n")
	return b.String()
}

func highlightCell(value string, editing bool, input textinput.Model) string {
	if editing {
		return "[" + input.View() + "]"
	}
	return "[" + strings.TrimSpace(value) + "]"
}

func (m *monthModel) beginEdit(initial string) {
	m.editing = true
	m.input.SetValue(initial)
	m.input.CursorEnd()
	m.input.Focus()
	m.err = ""
}

func (m *monthModel) cancelEdit() {
	m.editing = false
	m.input.Blur()
	m.err = ""
}

func cellSeed(row sheetDay, col int, unit units.Unit) string {
	if !row.HasEntry {
		return ""
	}
	switch col {
	case colWeight:
		if row.Log.Weight == nil {
			return ""
		}
		w, err := units.FromKG(*row.Log.Weight, unit)
		if err != nil {
			return ""
		}
		return formatWeight(w)
	case colSleep:
		if row.Log.SleepHours == 0 {
			return ""
		}
		return strconv.FormatFloat(row.Log.SleepHours, 'f', -1, 64)
	case colSteps:
		if row.Log.Steps == 0 {
			return ""
		}
		return strconv.Itoa(row.Log.Steps)
	case colNote:
		return row.Log.Note
	default:
		return ""
	}
}

func patchCell(log dailylog.DailyLog, col int, raw string, unit units.Unit) (dailylog.DailyLog, error) {
	raw = strings.TrimSpace(raw)
	switch col {
	case colWeight:
		if raw == "" {
			log.Weight = nil
			break
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("weight: not a number")
		}
		kg, err := units.ToKG(v, unit)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("weight: %w", err)
		}
		log.Weight = &kg
	case colSleep:
		if raw == "" {
			log.SleepHours = 0
			break
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("sleep: not a number")
		}
		log.SleepHours = v
	case colSteps:
		if raw == "" {
			log.Steps = 0
			break
		}
		v, err := strconv.Atoi(raw)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("steps: not a whole number")
		}
		log.Steps = v
	case colNote:
		log.Note = raw
	case colWorkout:
		// toggled by the caller
	default:
		return dailylog.DailyLog{}, fmt.Errorf("unknown column")
	}
	return dailylog.New(log.Day, log.Weight, log.SleepHours, log.Steps, log.Workout, log.Note)
}
