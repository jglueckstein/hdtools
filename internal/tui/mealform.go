package tui

// The meal form is the add, the grams edit, and the food edit. It is
// separate from the idle list so a refused number can stay on the
// field that failed. Tab is the only lookup of a name. Enter writes
// only when that lookup still matches. This file does not move the
// meal date and does not scale grams by the display unit.

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (a *App) beginMealAdd() {
	m := &a.mealUI
	m.mode = mealAdd
	m.err = ""
	m.prevSelected = m.selected
	m.known = false
	m.foodID = 0
	m.resolved = ""
	m.resolvedOK = false
	m.focus = mealFieldName
	for i := range m.inputs {
		m.inputs[i].SetValue("")
		m.inputs[i].CursorEnd()
	}
	m.focusMeal()
}

func (a *App) beginMealGrams() {
	m := &a.mealUI
	if m.selected < 0 {
		return
	}
	m.mode = mealEditGrams
	m.err = ""
	m.prevSelected = m.selected
	m.focus = mealFieldGrams
	m.inputs[mealFieldGrams].SetValue(strconv.FormatFloat(m.lines[m.selected].Grams, 'f', -1, 64))
	m.inputs[mealFieldGrams].CursorEnd()
	m.focusMeal()
}

func (a *App) beginMealFood() {
	m := &a.mealUI
	if m.selected < 0 {
		return
	}
	m.mode = mealEditFood
	m.err = ""
	m.prevSelected = m.selected
	line := m.lines[m.selected]
	vals := []float64{line.Calories, line.Protein, line.Carbohydrate, line.Fat}
	for i, v := range vals {
		m.inputs[mealFieldCal+i].SetValue(strconv.FormatFloat(v, 'f', -1, 64))
		m.inputs[mealFieldCal+i].CursorEnd()
	}
	m.focus = mealFieldCal
	m.focusMeal()
}

func (a *App) updateMealForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.cancelMealForm()
		return a, nil
	case "enter":
		a.submitMealForm()
		return a, nil
	}
	// Tab and a typed character answer the refused Enter. The
	// message stays through a weight-log reload, which does not
	// come through here, and comes back if the next Enter is
	// still refused.
	a.mealUI.err = ""
	if msg.String() == "tab" {
		a.tabMealForm()
		return a, nil
	}
	// Brackets, q, and f are characters here. They must not move
	// the date, quit, or open another screen.
	idx := a.mealUI.focus
	var cmd tea.Cmd
	a.mealUI.inputs[idx], cmd = a.mealUI.inputs[idx].Update(msg)
	return a, cmd
}

func (a *App) cancelMealForm() {
	a.mealUI.mode = mealIdle
	a.mealUI.err = ""
	a.mealUI.selected = a.mealUI.prevSelected
}

func (a *App) tabMealForm() {
	m := &a.mealUI
	if m.mode == mealAdd && m.focus == mealFieldName {
		a.resolveMealName()
	}
	fields := m.activeFields()
	cur := 0
	for i, field := range fields {
		if field == m.focus {
			cur = i
			break
		}
	}
	m.focus = fields[(cur+1)%len(fields)]
	m.focusMeal()
}

// resolveMealName rebuilds the add form from the trimmed name. Tab
// is the only time the name is looked up, so Enter can refuse a
// name that changed after that lookup without writing a row.
func (a *App) resolveMealName() {
	m := &a.mealUI
	name := strings.TrimSpace(m.inputs[mealFieldName].Value())
	if m.resolvedOK && name == m.resolved {
		return
	}
	m.resolvedOK = true
	m.resolved = name
	m.known = false
	m.foodID = 0
	for i := mealFieldCal; i <= mealFieldFat; i++ {
		m.inputs[i].SetValue("")
	}
	if name == "" || a.meals == nil {
		return
	}
	food, ok, err := a.meals.FoodByName(name)
	if err != nil {
		m.err = err.Error()
		return
	}
	if !ok {
		return
	}
	m.known = true
	m.foodID = food.ID
}

func (m mealModel) activeFields() []int {
	switch m.mode {
	case mealEditFood:
		return []int{mealFieldCal, mealFieldPro, mealFieldCarb, mealFieldFat}
	case mealEditGrams:
		return []int{mealFieldGrams}
	default:
		if m.known {
			return []int{mealFieldName, mealFieldGrams}
		}
		return []int{mealFieldName, mealFieldCal, mealFieldPro, mealFieldCarb, mealFieldFat, mealFieldGrams}
	}
}

func (m *mealModel) focusMeal() {
	active := map[int]bool{}
	for _, field := range m.activeFields() {
		active[field] = true
	}
	for i := range m.inputs {
		if i == m.focus && active[i] {
			m.inputs[i].Focus()
			continue
		}
		m.inputs[i].Blur()
	}
}

func (a *App) submitMealForm() {
	switch a.mealUI.mode {
	case mealAdd:
		a.submitMealAdd()
	case mealEditGrams:
		a.submitMealGrams()
	case mealEditFood:
		a.submitMealFood()
	}
}

func (a *App) submitMealAdd() {
	m := &a.mealUI
	name := strings.TrimSpace(m.inputs[mealFieldName].Value())
	if name == "" {
		m.err = "name: blank"
		m.focus = mealFieldName
		m.focusMeal()
		return
	}
	if !m.resolvedOK || name != m.resolved {
		m.err = "name: press tab"
		return
	}
	if m.known {
		grams, err := parseMealGrams(m.inputs[mealFieldGrams].Value())
		if err != nil {
			m.err = err.Error()
			m.focus = mealFieldGrams
			m.focusMeal()
			return
		}
		if err := a.meals.AddLine(m.day.Format(time.DateOnly), m.foodID, grams); err != nil {
			m.err = err.Error()
			return
		}
		a.finishMealSave(true)
		return
	}
	nums, err := a.parseFoodNumbers()
	if err != nil {
		return
	}
	grams, err := parseMealGrams(m.inputs[mealFieldGrams].Value())
	if err != nil {
		m.err = err.Error()
		m.focus = mealFieldGrams
		m.focusMeal()
		return
	}
	if err := a.meals.AddFoodLine(m.day.Format(time.DateOnly), name, nums[0], nums[1], nums[2], nums[3], grams); err != nil {
		m.err = err.Error()
		return
	}
	a.finishMealSave(true)
}

func (a *App) submitMealGrams() {
	m := &a.mealUI
	grams, err := parseMealGrams(m.inputs[mealFieldGrams].Value())
	if err != nil {
		m.err = err.Error()
		m.focus = mealFieldGrams
		m.focusMeal()
		return
	}
	id := m.lines[m.selected].ID
	if err := a.meals.UpdateGrams(id, grams); err != nil {
		m.err = err.Error()
		return
	}
	a.finishMealSave(false)
}

func (a *App) submitMealFood() {
	m := &a.mealUI
	line := m.lines[m.selected]
	nums, err := a.parseFoodNumbers()
	if err != nil {
		return
	}
	if err := a.meals.UpdateFood(line.FoodID, nums[0], nums[1], nums[2], nums[3]); err != nil {
		m.err = err.Error()
		return
	}
	a.finishMealSave(false)
}

func (a *App) finishMealSave(selectNew bool) {
	a.mealUI.mode = mealIdle
	if selectNew {
		a.loadMealDay(false)
		if n := len(a.mealUI.lines); n > 0 {
			a.mealUI.selected = n - 1
		}
		return
	}
	keep := a.mealUI.selected
	a.loadMealDay(false)
	if keep >= 0 && keep < len(a.mealUI.lines) {
		a.mealUI.selected = keep
	}
}

// parseFoodNumbers reads calories and the three macros. Calories
// are always required. A blank macro is zero only when tracking
// is off; the edit form prefills stored numbers, so leaving them
// keeps those numbers and clearing one stores zero. An error is
// already on the form.
func (a *App) parseFoodNumbers() ([4]float64, error) {
	m := &a.mealUI
	cal, err := parseMealNutrient(m.inputs[mealFieldCal].Value())
	if err != nil {
		m.err = "calories: " + err.Error()
		m.focus = mealFieldCal
		m.focusMeal()
		return [4]float64{}, fmt.Errorf("calories: %w", err)
	}
	nums := [4]float64{cal}
	for i, field := range []int{mealFieldPro, mealFieldCarb, mealFieldFat} {
		raw := strings.TrimSpace(m.inputs[field].Value())
		if raw == "" && !m.trackMacros {
			continue
		}
		v, err := parseMealNutrient(raw)
		if err != nil {
			shown := err.Error()
			m.err = mealNutrientName[i+1] + ": " + shown
			m.focus = field
			m.focusMeal()
			return [4]float64{}, fmt.Errorf("%s: %w", mealNutrientName[i+1], err)
		}
		nums[i+1] = v
	}
	return nums, nil
}

func parseMealGrams(raw string) (float64, error) {
	v, err := parseMealDecimal(raw)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("grams: want a number greater than zero")
	}
	return v, nil
}

func parseMealNutrient(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("want a number")
	}
	if strings.HasPrefix(s, "-") && mealDecimalText(s[1:]) {
		return 0, fmt.Errorf("want a number that is not negative")
	}
	v, err := parseMealDecimal(s)
	if err != nil {
		return 0, fmt.Errorf("want a number")
	}
	if v < 0 {
		return 0, fmt.Errorf("want a number that is not negative")
	}
	return v, nil
}

func parseMealDecimal(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if !mealDecimalText(s) {
		return 0, fmt.Errorf("want a number")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse meal number: %w", err)
	}
	return v, nil
}

// mealDecimalText is the gram grammar: an optional plus, then digits
// and at most one dot. Exponents, units, and slash fractions never
// reach ParseFloat, which would accept 1e2 and Inf.
func mealDecimalText(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' {
		s = s[1:]
	}
	if s == "" || s == "." {
		return false
	}
	dot := false
	digits := false
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits = true
		case r == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return digits
}
