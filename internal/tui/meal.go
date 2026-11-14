package tui

// The meal screen is one day's foods, not a weight-log column. A
// line keeps grams and which food, so the four per-100 g numbers
// live on the food and an edit shows up on every date that uses it.
// This file does not convert display units, set an allowance, or
// accept a volume. A nil meal store leaves f on the daily list.

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/meal"
)

const (
	mealIdleHelp = "a add   enter grams   e food   x remove   [ ] day   esc list   q quit"
	mealEditHelp = "enter save   esc cancel   tab next field"
)

const (
	mealFieldName = iota
	mealFieldCal
	mealFieldPro
	mealFieldCarb
	mealFieldFat
	mealFieldGrams
	mealFieldCount
)

var mealFieldLabel = []string{
	"name",
	"calories per 100 g",
	"protein per 100 g",
	"carbohydrate per 100 g",
	"fat per 100 g",
	"grams",
}

var mealNutrientName = []string{"calories", "protein", "carbohydrate", "fat"}

type mealMode int

const (
	mealIdle mealMode = iota
	mealAdd
	mealEditGrams
	mealEditFood
)

type mealModel struct {
	day          time.Time
	lines        []meal.Line
	selected     int
	prevSelected int
	mode         mealMode
	err          string
	inputs       []textinput.Model
	focus        int
	known        bool
	foodID       int64
	resolved     string
	resolvedOK   bool
}

// UseMeals stores the meal store. It does not open the meal screen.
// A nil store leaves f on the daily list.
func (a *App) UseMeals(store *meal.Store) {
	a.meals = store
}

// openMeals shows today. The list cursor stays where it was so Esc
// can return to that row, including when today has no weight log.
func (a *App) openMeals() {
	a.mealUI = mealModel{
		day:      localToday(),
		selected: -1,
		inputs:   newMealInputs(),
	}
	a.screen = screenMeal
	a.loadMealDay(true)
}

func newMealInputs() []textinput.Model {
	inputs := make([]textinput.Model, mealFieldCount)
	for i := range inputs {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 80
		ti.SetWidth(24)
		quietInput(&ti)
		inputs[i] = ti
	}
	return inputs
}

func (a *App) loadMealDay(selectFirst bool) {
	if a.meals == nil {
		a.mealUI.lines = nil
		a.mealUI.selected = -1
		return
	}
	lines, err := a.meals.Lines(a.mealUI.day.Format(time.DateOnly))
	if err != nil {
		a.mealUI.err = err.Error()
		return
	}
	a.mealUI.lines = lines
	a.mealUI.err = ""
	if selectFirst || a.mealUI.selected >= len(lines) {
		if len(lines) == 0 {
			a.mealUI.selected = -1
		} else if selectFirst {
			a.mealUI.selected = 0
		} else {
			a.mealUI.selected = len(lines) - 1
		}
	}
}

func (a *App) updateMeal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// ctrl+c quits from a field too. Idle q is the only other quit.
	if msg.String() == "ctrl+c" {
		return a, tea.Quit
	}
	if a.mealUI.mode != mealIdle {
		return a.updateMealForm(msg)
	}
	switch msg.String() {
	case "q":
		return a, tea.Quit
	case "esc":
		a.screen = screenList
		return a, nil
	case "a":
		a.beginMealAdd()
	case "x":
		a.removeMealLine()
	case "enter":
		a.beginMealGrams()
	case "e":
		a.beginMealFood()
	case "[":
		a.stepMealDay(-1)
	case "]":
		a.stepMealDay(1)
	case "up":
		a.moveMealLine(-1)
	case "down":
		a.moveMealLine(1)
	}
	return a, nil
}

func (a *App) stepMealDay(delta int) {
	a.mealUI.day = a.mealUI.day.AddDate(0, 0, delta)
	a.mealUI.mode = mealIdle
	a.loadMealDay(true)
}

func (a *App) moveMealLine(delta int) {
	if a.mealUI.selected < 0 {
		return
	}
	next := a.mealUI.selected + delta
	if next < 0 || next >= len(a.mealUI.lines) {
		return
	}
	a.mealUI.selected = next
}

func (a *App) removeMealLine() {
	if a.mealUI.selected < 0 || a.meals == nil {
		return
	}
	id := a.mealUI.lines[a.mealUI.selected].ID
	if err := a.meals.DeleteLine(id); err != nil {
		a.mealUI.err = err.Error()
		return
	}
	// The following line slides into this index. Past the end, the
	// new last line is selected. An empty date has no selection.
	a.loadMealDay(false)
}

func (a *App) mealView() string {
	p := a.pal
	m := a.mealUI
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", p.title.Render("meals "+m.day.Format(time.DateOnly)))
	if m.mode == mealIdle {
		a.writeMealLines(&b)
		writeMealTotals(&b, m.lines)
		if m.err != "" {
			fmt.Fprintf(&b, "\n%s\n", p.error.Render("error: "+m.err))
		}
		fmt.Fprintf(&b, "\n%s\n", p.help.Render(mealIdleHelp))
		return b.String()
	}
	a.writeMealForm(&b)
	if m.err != "" {
		fmt.Fprintf(&b, "\n%s\n", p.error.Render("error: "+m.err))
	}
	fmt.Fprintf(&b, "\n%s\n", p.help.Render(mealEditHelp))
	return b.String()
}

func (a *App) writeMealLines(b *strings.Builder) {
	for i, line := range a.mealUI.lines {
		mark := "  "
		if i == a.mealUI.selected {
			mark = "> "
		}
		fmt.Fprintf(b, "%s%s  %s\n", mark, line.Name, formatTenth(line.Grams))
	}
	if len(a.mealUI.lines) > 0 {
		fmt.Fprintln(b)
	}
}

func writeMealTotals(b *strings.Builder, lines []meal.Line) {
	var cal, pro, carb, fat float64
	for _, line := range lines {
		cal += line.Calories * line.Grams / 100
		pro += line.Protein * line.Grams / 100
		carb += line.Carbohydrate * line.Grams / 100
		fat += line.Fat * line.Grams / 100
	}
	fmt.Fprintf(b, "calories %s\n", formatTenth(cal))
	fmt.Fprintf(b, "protein %s\n", formatTenth(pro))
	fmt.Fprintf(b, "carbohydrate %s\n", formatTenth(carb))
	fmt.Fprintf(b, "fat %s\n", formatTenth(fat))
}

func (a *App) writeMealForm(b *strings.Builder) {
	m := a.mealUI
	p := a.pal
	if m.mode != mealAdd && m.selected >= 0 && m.selected < len(m.lines) {
		fmt.Fprintf(b, "%s\n", m.lines[m.selected].Name)
	}
	for _, field := range m.activeFields() {
		mark := "  "
		label := p.label
		if field == m.focus {
			mark = "> "
			label = p.focusLabel
		}
		fmt.Fprintf(b, "%s%s %s\n", mark, label.Render(mealFieldLabel[field]), m.inputs[field].View())
	}
}

// formatTenth is the weight column's one-decimal print. A negative
// zero from the formatter is a text bug, not a loss.
func formatTenth(v float64) string {
	s := fmt.Sprintf("%.1f", v)
	if s == "-0.0" {
		return "0.0"
	}
	return s
}
