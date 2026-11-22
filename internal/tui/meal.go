package tui

// The meal screen is one day's foods, not a weight-log column. A
// line keeps grams and which food, so the four per-100 g numbers
// live on the food and an edit shows up on every date that uses it.
// The same date can hold one calorie allowance. Remaining is that
// number minus the unrounded calorie total. This file does not
// convert display units and does not accept a volume. A nil meal
// store leaves f on the daily list.

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/meal"
)

const (
	mealIdleHelp      = "a add   enter grams   e food   x remove   b allowance   [ ] day   esc list   q quit"
	mealEditHelp      = "enter save   esc cancel   tab next field"
	mealAllowanceHelp = "enter save   esc cancel"
)

// mealScreenCols is the working width of a meal form line. The
// name is the only long value, and it has the line to itself.
// A 24-column box is a month-sheet cell. Using it here left the
// rest of the line empty, so a name scrolled while the screen
// still had room. The cursor cell sits outside the value.
const mealScreenCols = 72

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
	// mealAllowance is the one field b opens. Keys are text.
	// Enter writes or clears the number. Esc writes nothing.
	mealAllowance
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
	// trackMacros is copied when the screen opens. True makes
	// protein, carbohydrate, and fat mandatory and shows their
	// day totals. False still shows the fields. A blank one
	// then stores zero, and the day totals stay off.
	trackMacros bool
	// allowance is the number stored for this meal date.
	// allowanceSet is false when that date has no row. A stored
	// zero is set, so a false bool is not remaining 0.0.
	allowance    float64
	allowanceSet bool
	// allowanceInput is not one of the six food fields. Tab on
	// the food form still walks those, and this field has no
	// next field of its own.
	allowanceInput textinput.Model
	// allowanceEscList is set when a weight-log reload finishes
	// while this field is open. Esc then returns to the daily
	// list. Closing the field before that reload returns to the
	// idle meal screen.
	allowanceEscList bool
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
		day:            localToday(),
		selected:       -1,
		inputs:         newMealInputs(),
		trackMacros:    a.cfg.TrackMacros,
		allowanceInput: newAllowanceInput(),
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
		// A number is short. Padding it out to the line would push
		// "carbohydrate per 100 g" past an 80-column terminal.
		ti.SetWidth(24)
		if i == mealFieldName {
			ti.SetWidth(mealNameWidth())
		}
		quietInput(&ti)
		inputs[i] = ti
	}
	return inputs
}

func newAllowanceInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 80
	ti.SetWidth(24)
	quietInput(&ti)
	return ti
}

// mealNameWidth is the name field's visible columns. The line is
// the focus mark, the label, a space, the field, and the cursor.
func mealNameWidth() int {
	return mealScreenCols - (2 + len(mealFieldLabel[mealFieldName]) + 1 + 1)
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
	a.loadMealAllowance()
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
	if a.mealUI.mode == mealAllowance {
		return a.updateMealAllowance(msg)
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
	case "b":
		a.beginMealAllowance()
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
	if m.mode == mealAllowance {
		a.writeMealAllowance(&b)
		return b.String()
	}
	if m.mode == mealIdle {
		a.writeMealLines(&b)
		cal := writeMealTotals(&b, m.lines, m.trackMacros)
		writeMealComparison(&b, m.allowanceSet, m.allowance, cal)
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

func writeMealTotals(b *strings.Builder, lines []meal.Line, trackMacros bool) float64 {
	var cal, pro, carb, fat float64
	for _, line := range lines {
		cal += line.Calories * line.Grams / 100
		pro += line.Protein * line.Grams / 100
		carb += line.Carbohydrate * line.Grams / 100
		fat += line.Fat * line.Grams / 100
	}
	fmt.Fprintf(b, "calories %s\n", formatTenth(cal))
	if !trackMacros {
		return cal
	}
	fmt.Fprintf(b, "protein %s\n", formatTenth(pro))
	fmt.Fprintf(b, "carbohydrate %s\n", formatTenth(carb))
	fmt.Fprintf(b, "fat %s\n", formatTenth(fat))
	return cal
}

// writeMealComparison paints the idle line under the totals. Unset
// is the words, not a remaining of zero. Set subtracts the unrounded
// calorie total once, then the same tenth the calorie line uses.
func writeMealComparison(b *strings.Builder, set bool, allowance, calories float64) {
	if !set {
		fmt.Fprintln(b, "no allowance")
		return
	}
	fmt.Fprintf(b, "remaining %s\n", formatTenth(allowance-calories))
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

// loadMealAllowance reads this date's row. A missing row is unset.
// It does not copy another date and it does not invent a zero.
func (a *App) loadMealAllowance() {
	a.mealUI.allowance = 0
	a.mealUI.allowanceSet = false
	if a.meals == nil {
		return
	}
	cal, ok, err := a.meals.Allowance(a.mealUI.day.Format(time.DateOnly))
	if err != nil {
		a.mealUI.err = err.Error()
		return
	}
	if !ok {
		return
	}
	a.mealUI.allowance = cal
	a.mealUI.allowanceSet = true
}

// beginMealAllowance replaces the idle body with one field. Unset
// opens empty. Set opens on the stored number, not on a tenth.
func (a *App) beginMealAllowance() {
	m := &a.mealUI
	m.mode = mealAllowance
	m.err = ""
	m.allowanceEscList = false
	if m.allowanceSet {
		m.allowanceInput.SetValue(strconv.FormatFloat(m.allowance, 'f', -1, 64))
	} else {
		m.allowanceInput.SetValue("")
	}
	m.allowanceInput.CursorEnd()
	m.allowanceInput.Focus()
}

func (a *App) updateMealAllowance(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.cancelMealAllowance()
		return a, nil
	case "enter":
		a.submitMealAllowance()
		return a, nil
	}
	// The next typed character answers a refused Enter. Tab stays
	// on this field: it does not save, cancel, or walk anywhere.
	a.mealUI.err = ""
	var cmd tea.Cmd
	a.mealUI.allowanceInput, cmd = a.mealUI.allowanceInput.Update(msg)
	return a, cmd
}

func (a *App) cancelMealAllowance() {
	// The reload that landed while this field was open is the
	// Esc that returns to the list. Any other Esc only drops
	// the field.
	if a.mealUI.allowanceEscList {
		a.mealUI.allowanceEscList = false
		a.mealUI.mode = mealIdle
		a.mealUI.err = ""
		a.screen = screenList
		return
	}
	a.mealUI.mode = mealIdle
	a.mealUI.err = ""
}

func (a *App) submitMealAllowance() {
	m := &a.mealUI
	if a.meals == nil {
		m.err = "allowance: meals are not open"
		return
	}
	calories, clear, err := parseAllowance(m.allowanceInput.Value())
	if err != nil {
		m.err = err.Error()
		return
	}
	day := m.day.Format(time.DateOnly)
	if clear {
		err = a.meals.ClearAllowance(day)
	} else {
		err = a.meals.SetAllowance(day, calories)
	}
	if err != nil {
		m.err = err.Error()
		return
	}
	m.mode = mealIdle
	m.allowanceEscList = false
	a.loadMealDay(false)
}

func (a *App) writeMealAllowance(b *strings.Builder) {
	m := a.mealUI
	p := a.pal
	fmt.Fprintf(b, "> %s %s\n", p.focusLabel.Render("allowance"), m.allowanceInput.View())
	if m.err != "" {
		fmt.Fprintf(b, "\n%s\n", p.error.Render("error: "+m.err))
	}
	fmt.Fprintf(b, "\n%s\n", p.help.Render(mealAllowanceHelp))
}

// parseAllowance trims, then either clears or accepts one
// non-negative decimal. A leading minus is refused before the
// grammar, so -0 is not stored as zero. Blank is a clear, which
// is the opposite of a food number.
func parseAllowance(raw string) (float64, bool, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, true, nil
	}
	// A minus anywhere is the negative refusal. The field opens
	// on the stored number, so typing -1 produces 500-1, and -0
	// must not be stored as zero.
	if strings.Contains(s, "-") {
		return 0, false, fmt.Errorf("allowance: want a number that is not negative")
	}
	if !mealDecimalText(s) {
		return 0, false, fmt.Errorf("allowance: want a number")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false, fmt.Errorf("allowance: want a number")
	}
	return v, false, nil
}
