package tui

// The day form is create and edit for a single calendar day. Month-cell
// edits reuse parseForm/patchCell so a typed weight still converts to kg
// through the same path. Date is editable: saving a new date upserts that
// day and leaves the old row in place (the paper log never deletes a line
// by changing its date).

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	fieldDate = iota
	fieldWeight
	fieldSleep
	fieldSteps
	fieldWorkout
	fieldNote
	fieldCount
)

type formModel struct {
	inputs  []textinput.Model
	focus   int
	workout bool
	err     string
	unit    units.Unit
}

func newForm(unit units.Unit) formModel {
	placeholders := []string{"YYYY-MM-DD", "optional", "hours", "count", "optional"}
	inputs := make([]textinput.Model, 5)
	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.Prompt = ""
		ti.Width = 32
		ti.CharLimit = 200
		if i == 0 {
			ti.CharLimit = 10
			ti.Width = 12
		}
		inputs[i] = ti
	}
	f := formModel{inputs: inputs, unit: unit}
	f.setFocus(fieldDate)
	return f
}

func (f *formModel) load(log dailylog.DailyLog) {
	f.inputs[0].SetValue(log.Day.Format(time.DateOnly))
	if log.Weight != nil {
		w, err := units.FromKG(*log.Weight, f.unit)
		if err == nil {
			f.inputs[1].SetValue(formatWeight(w))
		}
	} else {
		f.inputs[1].SetValue("")
	}
	if log.SleepHours != 0 {
		f.inputs[2].SetValue(strconv.FormatFloat(log.SleepHours, 'f', -1, 64))
	} else {
		f.inputs[2].SetValue("")
	}
	if log.Steps != 0 {
		f.inputs[3].SetValue(strconv.Itoa(log.Steps))
	} else {
		f.inputs[3].SetValue("")
	}
	f.workout = log.Workout
	f.inputs[4].SetValue(log.Note)
	f.err = ""
	f.setFocus(fieldDate)
}

func (f *formModel) loadNew(day time.Time) {
	f.load(dailylog.DailyLog{Day: day})
}

func (f *formModel) setFocus(i int) {
	f.focus = i
	for idx := range f.inputs {
		if idx == inputIndex(i) && i != fieldWorkout {
			f.inputs[idx].Focus()
		} else {
			f.inputs[idx].Blur()
		}
	}
}

func inputIndex(focus int) int {
	if focus == fieldNote {
		return 4
	}
	if focus == fieldWorkout {
		return -1
	}
	return focus
}

func (f *formModel) update(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "tab", "down":
			f.setFocus((f.focus + 1) % fieldCount)
			return nil
		case "shift+tab", "up":
			f.setFocus((f.focus - 1 + fieldCount) % fieldCount)
			return nil
		case " ":
			if f.focus == fieldWorkout {
				f.workout = !f.workout
				return nil
			}
		}
	}
	idx := inputIndex(f.focus)
	if idx < 0 {
		return nil
	}
	var cmd tea.Cmd
	f.inputs[idx], cmd = f.inputs[idx].Update(msg)
	return cmd
}

func (f formModel) parse() (dailylog.DailyLog, error) {
	return parseForm(
		strings.TrimSpace(f.inputs[0].Value()),
		strings.TrimSpace(f.inputs[1].Value()),
		strings.TrimSpace(f.inputs[2].Value()),
		strings.TrimSpace(f.inputs[3].Value()),
		strings.TrimSpace(f.inputs[4].Value()),
		f.workout,
		f.unit,
	)
}

func (f formModel) view(p palette) string {
	workout := "no"
	if f.workout {
		workout = "yes"
	}
	label := func(i int, name string) string {
		mark := " "
		style := p.label
		if f.focus == i {
			mark = ">"
			style = p.focusLabel
		}
		return mark + " " + style.Render(fmt.Sprintf("%-8s", name))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", p.title.Render(fmt.Sprintf("  day form  (weight in %s)", f.unit)))
	fmt.Fprintf(&b, "%s %s\n", label(fieldDate, "date"), f.inputs[0].View())
	fmt.Fprintf(&b, "%s %s\n", label(fieldWeight, "weight"), f.inputs[1].View())
	fmt.Fprintf(&b, "%s %s\n", label(fieldSleep, "sleep"), f.inputs[2].View())
	fmt.Fprintf(&b, "%s %s\n", label(fieldSteps, "steps"), f.inputs[3].View())
	fmt.Fprintf(&b, "%s %s  %s\n", label(fieldWorkout, "workout"), workout, p.muted.Render("(space to toggle)"))
	fmt.Fprintf(&b, "%s %s\n", label(fieldNote, "note"), f.inputs[4].View())
	if f.err != "" {
		fmt.Fprintf(&b, "\n%s\n", p.error.Render("error: "+f.err))
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render("enter save   esc cancel   tab next field"))
	return b.String()
}

// parseForm turns form strings into a DailyLog. Empty weight means no scale
// reading. Display-unit weight is converted to kilograms for storage.
func parseForm(date, weight, sleep, steps, note string, workout bool, unit units.Unit) (dailylog.DailyLog, error) {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return dailylog.DailyLog{}, fmt.Errorf("date: want YYYY-MM-DD")
	}
	var kg *float64
	if weight != "" {
		v, err := strconv.ParseFloat(weight, 64)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("weight: not a number")
		}
		stored, err := units.ToKG(v, unit)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("weight: %w", err)
		}
		kg = &stored
	}
	var sleepHours float64
	if sleep != "" {
		sleepHours, err = strconv.ParseFloat(sleep, 64)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("sleep: not a number")
		}
	}
	var stepCount int
	if steps != "" {
		stepCount, err = strconv.Atoi(steps)
		if err != nil {
			return dailylog.DailyLog{}, fmt.Errorf("steps: not a whole number")
		}
	}
	log, err := dailylog.New(day, kg, sleepHours, stepCount, workout, note)
	if err != nil {
		return dailylog.DailyLog{}, fmt.Errorf("parse form: %w", err)
	}
	return log, nil
}

func formatWeight(w float64) string {
	return strconv.FormatFloat(w, 'f', 1, 64)
}

func localToday() time.Time {
	now := time.Now()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
