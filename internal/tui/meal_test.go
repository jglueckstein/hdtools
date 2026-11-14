package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/meal"
	"github.com/jglueckstein/hdtools/internal/units"
)

const (
	mealToday    = "1990-11-10"
	listHelp     = "n new   enter edit   m month   c chart   l long   f meal   t today   q quit"
	emptyHelp    = "n new day   m month   l long   f meal   t today   q quit"
	mealHelp     = "a add   enter grams   e food   x remove   [ ] day   esc list   q quit"
	mealFormHelp = "enter save   esc cancel   tab next field"
)

func freezeMealToday(t *testing.T) {
	t.Helper()
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
}

func shown(app *App) string {
	return visible(app.View().Content)
}

// openMealFixture builds the app the screen tests drive. meal.Open
// returns an error until the tables exist; UseMeals still receives
// that result so a nil store cannot panic. The missing screen is
// what the assertions fail on.
func openMealFixture(t *testing.T, cfg config.Config, days ...time.Time) (*App, *dailylog.Store, *meal.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hdtools.db")
	store, err := dailylog.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if len(days) > 0 {
		seedDays(t, store, days...)
	}
	meals, _ := meal.Open(path)
	app := New(store, path, cfg)
	app.UseMeals(meals)
	if msg := app.load(); msg != nil {
		app.Update(msg)
	}
	return app, store, meals, path
}

// mealApp runs the scenarios that enter protein, carbohydrate,
// and fat. The default config leaves track_macros unset.
func mealApp(t *testing.T, days ...time.Time) (*App, *dailylog.Store, *meal.Store, string) {
	t.Helper()
	freezeMealToday(t)
	cfg := config.Default()
	cfg.TrackMacros = true
	return openMealFixture(t, cfg, days...)
}

func closeMeal(meals *meal.Store) {
	if meals == nil {
		return
	}
	if c, ok := any(meals).(interface{ Close() error }); ok {
		_ = c.Close()
	}
}

func assertMeals(t *testing.T, app *App, day string) {
	t.Helper()
	view := shown(app)
	if !strings.Contains(view, "meals "+day) {
		t.Fatalf("view = %q, want meals %s", view, day)
	}
	if strings.Contains(view, "day form") {
		t.Fatalf("day form shown: %q", view)
	}
	if strings.Contains(view, "arrows move") {
		t.Fatalf("month sheet shown: %q", view)
	}
}

func openMeals(t *testing.T, app *App, day string) {
	t.Helper()
	press(app, "f")
	assertMeals(t, app, day)
}

func assertFocus(t *testing.T, app *App, label string) {
	t.Helper()
	view := shown(app)
	if !strings.Contains(view, "> "+label) {
		t.Fatalf("focus = %q, want %s", view, label)
	}
}

func assertTotals(t *testing.T, app *App, cal, pro, carb, fat string) {
	t.Helper()
	view := shown(app)
	for _, pair := range [][2]string{
		{"calories", cal},
		{"protein", pro},
		{"carbohydrate", carb},
		{"fat", fat},
	} {
		if !labeledNumber(view, pair[0], pair[1]) {
			t.Fatalf("missing %s %s in %q", pair[0], pair[1], view)
		}
	}
}

func assertNoSelection(t *testing.T, app *App) {
	t.Helper()
	view := shown(app)
	for _, line := range strings.Split(view, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			t.Fatalf("selected %q in %q", strings.TrimSpace(line), view)
		}
	}
}

func selectedMealLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, ">") {
			return trim
		}
	}
	return ""
}

func assertGramsOnly(t *testing.T, view, name, grams string) {
	t.Helper()
	var line string
	if sel := selectedMealLine(view); strings.Contains(sel, name) && numberToken(sel, grams) {
		line = sel
	}
	if line == "" {
		for _, row := range strings.Split(view, "\n") {
			if strings.Contains(row, name) && numberToken(row, grams) && !strings.Contains(row, "per 100") && !strings.Contains(row, "calories") && !strings.Contains(row, "protein") && !strings.Contains(row, "carbohydrate") {
				line = row
				break
			}
		}
	}
	if line == "" {
		t.Fatalf("no %s line of %s g in %q", name, grams, view)
	}
	rest := strings.Replace(line, name, "", 1)
	rest = stripNumberToken(rest, grams)
	if strings.ContainsAny(rest, "0123456789") {
		t.Fatalf("line %q shows a number besides %s", line, grams)
	}
}

// labeledNumber matches the word, a space, and that exact decimal.
// "calories 1.2" is not "calories 1.25".
func labeledNumber(view, label, value string) bool {
	return numberToken(view, label+" "+value)
}

func numberToken(s, value string) bool {
	if value == "" {
		return false
	}
	return stripNumberToken(s, value) != s
}

func stripNumberToken(s, value string) string {
	for i := 0; i+len(value) <= len(s); i++ {
		if s[i:i+len(value)] != value {
			continue
		}
		if i > 0 && isNumByte(s[i-1]) {
			continue
		}
		end := i + len(value)
		if end < len(s) && isNumByte(s[end]) {
			continue
		}
		return s[:i] + s[end:]
	}
	return s
}

func isNumByte(b byte) bool {
	return (b >= '0' && b <= '9') || b == '.'
}

func assertFieldShows(t *testing.T, app *App, label, value string) {
	t.Helper()
	view := shown(app)
	var line string
	for _, row := range strings.Split(view, "\n") {
		if strings.Contains(row, "> "+label) {
			line = row
			break
		}
	}
	if line == "" {
		t.Fatalf("no focused %s in %q", label, view)
	}
	rest := strings.Replace(line, label, "", 1)
	if !numberToken(rest, value) {
		t.Fatalf("%s = %q, want %s", label, strings.TrimSpace(line), value)
	}
}

func assertFieldError(t *testing.T, app *App, field, phrase string) {
	t.Helper()
	view := shown(app)
	for _, line := range strings.Split(view, "\n") {
		if !strings.Contains(line, "error:") {
			continue
		}
		fieldAt := strings.Index(line, field)
		phraseAt := strings.Index(line, phrase)
		if fieldAt >= 0 && phraseAt > fieldAt {
			return
		}
		t.Fatalf("error line %q, want %s then %s", line, field, phrase)
	}
	t.Fatalf("no error in %q", view)
}

func assertListRow(t *testing.T, app *App, day time.Time) {
	t.Helper()
	view := shown(app)
	if !strings.Contains(view, "daily log") {
		t.Fatalf("not the daily list: %q", view)
	}
	if strings.Contains(view, "meals ") {
		t.Fatalf("still on meals: %q", view)
	}
	assertSelectedDay(t, app, day)
	want := day.Format(time.DateOnly)
	if !strings.Contains(selectedMealLine(view), want) {
		t.Fatalf("row %s not selected in %q", want, view)
	}
}

func saveNew(app *App, name, cal, pro, carb, fat, grams string) {
	press(app, "a")
	if name != "" {
		press(app, name)
	}
	press(app, "tab")
	if cal != "" {
		press(app, cal)
	}
	press(app, "tab")
	if pro != "" {
		press(app, pro)
	}
	press(app, "tab")
	if carb != "" {
		press(app, carb)
	}
	press(app, "tab")
	if fat != "" {
		press(app, fat)
	}
	press(app, "tab")
	if grams != "" {
		press(app, grams)
	}
	press(app, "enter")
}

func saveGrams(app *App, name, grams string) {
	press(app, "a")
	press(app, name)
	press(app, "tab")
	if grams != "" {
		press(app, grams)
	}
	press(app, "enter")
}

// replaceField clears the focused field, then types value.
// The caret sits at the end of the old text, and ctrl+u
// deletes back to the start, so the new text is not appended.
func replaceField(app *App, value string) {
	app.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	if value != "" {
		press(app, value)
	}
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func commandQuits(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func follow(t *testing.T, app *App, cmd tea.Cmd) {
	t.Helper()
	for i := 0; i < 3 && cmd != nil; i++ {
		msg := cmd()
		if msg == nil {
			return
		}
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Fatal("quit")
		}
		_, cmd = app.Update(msg)
	}
}

func assertNameIsNew(t *testing.T, app *App, name string) {
	t.Helper()
	press(app, "a")
	press(app, name)
	press(app, "tab")
	assertFocus(t, app, "calories per 100 g")
	press(app, "esc")
}

func assertNoPDFs(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				assertNoPDFs(t, filepath.Join(dir, entry.Name()))
				continue
			}
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
				t.Fatalf("wrote %s", filepath.Join(dir, entry.Name()))
			}
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func readRepo(t *testing.T, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(body)
}

func TestMealFOpensTodayNotTheSelectedRow(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	assertSelectedDay(t, app, dateUTC(1990, time.November, 1))
	press(app, "f")
	assertMeals(t, app, mealToday)
}

func TestMealFOpensTodayWithNoLogRow(t *testing.T) {
	app, store, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	openMeals(t, app, mealToday)
	press(app, "esc")
	assertNoDay(t, store, dateUTC(1990, time.November, 10))
	assertListRow(t, app, dateUTC(1990, time.November, 1))
}

func TestMealEscReturnsToTheSameRow(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	openMeals(t, app, mealToday)
	_, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if commandQuits(cmd) {
		t.Fatal("esc quit")
	}
	assertListRow(t, app, dateUTC(1990, time.November, 1))
}

func TestMealNextDayLeavesTheListAndTheMonth(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	openMeals(t, app, mealToday)
	press(app, "]")
	assertMeals(t, app, "1990-11-11")
	press(app, "esc")
	assertListRow(t, app, dateUTC(1990, time.November, 1))
	press(app, "m")
	if !strings.Contains(shown(app), "November 1990") {
		t.Fatalf("month = %q, want November 1990", shown(app))
	}
	press(app, "esc")
	openMeals(t, app, mealToday)
	for i := 0; i < 21; i++ {
		press(app, "]")
	}
	assertMeals(t, app, "1990-12-01")
	press(app, "esc")
	press(app, "m")
	if !strings.Contains(shown(app), "November 1990") {
		t.Fatalf("month after December meals = %q, want November 1990", shown(app))
	}
}

func TestMealPrevDayIsCalendar(t *testing.T) {
	freezeToday(t, time.Date(1992, 3, 1, 12, 0, 0, 0, time.UTC))
	app, _, _, _ := openMealFixture(t, config.Default())
	openMeals(t, app, "1992-03-01")
	press(app, "[")
	assertMeals(t, app, "1992-02-29")
	press(app, "]")
	assertMeals(t, app, "1992-03-01")

	freezeToday(t, time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC))
	year, _, _, _ := openMealFixture(t, config.Default())
	openMeals(t, year, "1990-01-01")
	press(year, "[")
	assertMeals(t, year, "1989-12-31")
}

func TestListBracketsDoNotOpenMeals(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	press(app, "m")
	if !strings.Contains(shown(app), "November 1990") {
		t.Fatalf("month = %q", shown(app))
	}
	press(app, "esc", "[", "]")
	assertListRow(t, app, dateUTC(1990, time.November, 1))
	if strings.Contains(shown(app), "meals ") {
		t.Fatalf("brackets opened meals: %q", shown(app))
	}
	press(app, "m")
	if !strings.Contains(shown(app), "November 1990") {
		t.Fatalf("month moved: %q", shown(app))
	}
}

func TestMealFDoesNotChangeTheDate(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, _, _, _ := mealApp(t, day)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "]")
	assertMeals(t, app, "1990-11-11")
	if strings.Contains(shown(app), "Oats") {
		t.Fatalf("line moved onto the next day: %q", shown(app))
	}
	press(app, "f")
	assertMeals(t, app, "1990-11-11")
	if strings.Count(shown(app), "meals ") != 1 || strings.Contains(shown(app), "Oats") {
		t.Fatalf("f changed the meal: %q", shown(app))
	}
	press(app, "esc")
	assertListRow(t, app, day)
	openMeals(t, app, mealToday)
	assertGramsOnly(t, shown(app), "Oats", "40.0")
}

func TestMealIgnoresListKeys(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()
	pdfDir := t.TempDir()
	cwd := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Chdir(cwd)
	freezeMealToday(t)
	cfg := config.Default()
	cfg.PDFDir = pdfDir
	cfg.TrackMacros = true
	app, _, _, dbPath := openMealFixture(t, cfg)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	before := shown(app)
	for _, key := range []rune{'n', 'm', 'c', 'l', 'p', 'P', 't'} {
		_, cmd := app.Update(letter(key))
		// follow is the only call. press would drop the command,
		// and calling it twice would drop the message that writes.
		follow(t, app, cmd)
		view := shown(app)
		if !strings.Contains(view, "meals "+mealToday) || !strings.Contains(view, "Oats") {
			t.Fatalf("%s left the meal: %q", string(key), view)
		}
		if view != before {
			t.Fatalf("%s changed the meal\nbefore %q\nafter  %q", string(key), before, view)
		}
	}
	assertNoPDFs(t, home, data, pdfDir, cwd, filepath.Dir(dbPath))
}

func TestMonthFStillEdits(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	press(app, "m", "f")
	if !app.month.editing || app.month.input.Value() != "f" {
		t.Fatalf("edit = %v %q, want text f", app.month.editing, app.month.input.Value())
	}
	if strings.Contains(shown(app), "meals ") {
		t.Fatalf("month f opened meals: %q", shown(app))
	}
	press(app, "esc")
	for app.month.col != colWorkout {
		app.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	}
	press(app, "f")
	if app.month.editing {
		t.Fatal("workout f started an edit")
	}
	if strings.Contains(shown(app), "meals ") {
		t.Fatalf("workout f opened meals: %q", shown(app))
	}
}

func TestChartFDoesNothing(t *testing.T) {
	pdfDir := t.TempDir()
	cwd := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Chdir(cwd)
	freezeMealToday(t)
	cfg := config.Default()
	cfg.PDFDir = pdfDir
	app, _, _, _ := openMealFixture(t, cfg, dateUTC(1990, time.November, 1))
	press(app, "c")
	if !strings.Contains(shown(app), "November 1990") {
		t.Fatalf("chart = %q", shown(app))
	}
	_, cmd := app.Update(letter('f'))
	follow(t, app, cmd)
	view := shown(app)
	if !strings.Contains(view, "November 1990") || strings.Contains(view, "meals ") || strings.Contains(view, "daily log") {
		t.Fatalf("f left the monthly chart: %q", view)
	}
	press(app, "esc", "l")
	if !strings.Contains(shown(app), "Quarterly") {
		t.Fatalf("long chart = %q", shown(app))
	}
	_, cmd = app.Update(letter('f'))
	follow(t, app, cmd)
	view = shown(app)
	if !strings.Contains(view, "Quarterly") || strings.Contains(view, "meals ") {
		t.Fatalf("f left the long-term chart: %q", view)
	}
	assertNoPDFs(t, pdfDir, cwd)
}

func TestFormFIsText(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	press(app, "n", "tab", "tab", "tab", "tab", "tab", "f")
	if app.screen != screenForm {
		t.Fatalf("screen = %v, want the day form", app.screen)
	}
	if got := app.form.inputs[inputIndex(fieldNote)].Value(); got != "f" {
		t.Fatalf("note = %q, want f", got)
	}
	if strings.Contains(shown(app), "meals ") {
		t.Fatalf("note f opened meals: %q", shown(app))
	}
	press(app, "esc", "n", "tab", "tab", "tab", "tab")
	if app.form.focus != fieldWorkout {
		t.Fatalf("focus = %d, want workout", app.form.focus)
	}
	press(app, "f")
	if app.form.workout {
		t.Fatal("workout f toggled the workout")
	}
	if app.screen != screenForm || strings.Contains(shown(app), "meals ") {
		t.Fatalf("workout f opened meals: %q", shown(app))
	}
}

func TestMealDoesNotChangeTheWeightLog(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, store, _, _ := mealApp(t, day)
	before := rowText(shown(app), "1990-11-01")
	if before == "" {
		t.Fatalf("no list row in %q", shown(app))
	}
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "50")
	assertGramsOnly(t, shown(app), "Oats", "50.0")
	press(app, "esc")
	if got := rowText(shown(app), "1990-11-01"); got != before {
		t.Fatalf("list row = %q, want %q", got, before)
	}
	header := headerLine(shown(app))
	for _, label := range []string{"date", "weight", "trend", "delta", "sleep", "steps", "workout", "note"} {
		if !strings.Contains(header, label) {
			t.Fatalf("header %q missing %s", header, label)
		}
	}
	assertStoredWeight(t, store, day, 80)
	assertNoDay(t, store, dateUTC(1990, time.November, 10))
	press(app, "c")
	if strings.Contains(shown(app), "Oats") {
		t.Fatalf("monthly chart shows the food: %q", shown(app))
	}
	press(app, "esc", "l")
	if strings.Contains(shown(app), "Oats") {
		t.Fatalf("long-term chart shows the food: %q", shown(app))
	}
}

func TestMealEmptyDateIsZero(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	assertNoSelection(t, app)
	view := shown(app)
	if strings.Contains(view, "no allowance") {
		t.Fatalf("allowance copy shown: %q", view)
	}
	if strings.Contains(view, "Oats") {
		t.Fatalf("starter name shown: %q", view)
	}
}

func TestMealEmptySelectionIgnoresEditKeys(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "x", "enter", "e")
	assertMeals(t, app, mealToday)
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	assertNoSelection(t, app)
	if strings.Contains(shown(app), "per 100 g") || strings.Contains(shown(app), "> grams") {
		t.Fatalf("edit opened on an empty date: %q", shown(app))
	}
	press(app, "a")
	assertFocus(t, app, "name")
}

func TestMealAddStartsOnTheName(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "a")
	view := shown(app)
	if !strings.Contains(view, "meals "+mealToday) {
		t.Fatalf("heading lost the date: %q", view)
	}
	assertFocus(t, app, "name")
	if strings.Contains(view, "Oats") {
		t.Fatalf("line list still shown: %q", view)
	}
}

func TestMealLongNameStaysOnTheLine(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	name := "Pompeian Extra Virgin Olive Oil"
	press(app, "a", name)
	line := ""
	for _, row := range strings.Split(shown(app), "\n") {
		if strings.Contains(row, "> name") {
			line = row
			break
		}
	}
	if line == "" || !strings.Contains(line, name) {
		t.Fatalf("name scrolled off: %q", line)
	}
	if w := lipgloss.Width(line); w > 72 {
		t.Fatalf("name line is %d columns", w)
	}
}

func TestMealAddTabOrderForANewFood(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "a", "Rye")
	assertFocus(t, app, "name")
	press(app, "tab")
	assertFocus(t, app, "calories per 100 g")
	press(app, "tab")
	assertFocus(t, app, "protein per 100 g")
	press(app, "tab")
	assertFocus(t, app, "carbohydrate per 100 g")
	press(app, "tab")
	assertFocus(t, app, "fat per 100 g")
	press(app, "tab")
	assertFocus(t, app, "grams")
	press(app, "tab")
	assertFocus(t, app, "name")
	press(app, "esc")
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	assertNoSelection(t, app)
	if strings.Contains(shown(app), "Rye") {
		t.Fatalf("tab wrote Rye: %q", shown(app))
	}
}

func TestMealAddExistingFoodAsksOnlyGrams(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	press(app, "a", " Oats ", "tab")
	assertFocus(t, app, "grams")
	view := shown(app)
	for _, label := range []string{"calories per 100 g", "protein per 100 g", "carbohydrate per 100 g", "fat per 100 g"} {
		if strings.Contains(view, label) {
			t.Fatalf("known name asked for %s: %q", label, view)
		}
	}
	press(app, "40", "enter")
	assertTotals(t, app, "280.0", "14.0", "28.0", "7.0")
	assertGramsOnly(t, shown(app), "Oats", "40.0")
}

func TestMealAddSavesScaledTotals(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "50")
	view := shown(app)
	assertGramsOnly(t, view, "Oats", "50.0")
	if !strings.Contains(selectedMealLine(view), "Oats") {
		t.Fatalf("selected = %q, want Oats", selectedMealLine(view))
	}
	assertTotals(t, app, "100.0", "5.0", "10.0", "2.5")
	press(app, "]")
	saveGrams(app, "Oats", "100")
	assertMeals(t, app, "1990-11-11")
	assertTotals(t, app, "200.0", "10.0", "20.0", "5.0")
	assertGramsOnly(t, shown(app), "Oats", "100.0")
}

func TestMealAddEscWritesNothing(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "a", "Rye", "tab", "200", "tab", "10", "tab", "20", "tab", "5", "tab", "50", "esc")
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	assertNoSelection(t, app)
	if strings.Contains(shown(app), "Rye") {
		t.Fatalf("esc stored Rye: %q", shown(app))
	}
	assertNameIsNew(t, app, "Rye")
}

func TestMealNameTrimAndCase(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "a", " Oats ", "tab")
	assertFocus(t, app, "grams")
	press(app, "10", "enter")
	// The same totals are the trim. A stored " Oats " would not
	// share the 200/10/20/5 row, and "> Oats" still contains a space.
	assertTotals(t, app, "100.0", "5.0", "10.0", "2.5")
	press(app, "a", "oats", "tab")
	assertFocus(t, app, "calories per 100 g")
	press(app, "1", "tab", "1", "tab", "1", "tab", "1", "tab", "100", "enter")
	view := shown(app)
	if !strings.Contains(view, "oats") {
		t.Fatalf("lowercase name was not its own food: %q", view)
	}
	assertTotals(t, app, "101.0", "6.0", "11.0", "3.5")
	press(app, "a", "Oat s", "tab")
	assertFocus(t, app, "calories per 100 g")
}

func TestMealBlankNameRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	for _, name := range []string{"", "   "} {
		press(app, "a")
		if name != "" {
			press(app, name)
		}
		press(app, "enter")
		view := shown(app)
		if !strings.Contains(view, "error:") || !strings.Contains(view, "name: blank") {
			t.Fatalf("blank name %q: %q", name, view)
		}
		press(app, "esc")
		assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
		assertNoSelection(t, app)
	}
}

func TestMealMissingNutrientRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	fields := []string{"calories", "protein", "carbohydrate", "fat"}
	for i, field := range fields {
		name := "Food" + string(rune('A'+i))
		values := []string{"1", "1", "1", "1"}
		values[i] = ""
		press(app, "a", name)
		for _, value := range values {
			press(app, "tab")
			if value != "" {
				press(app, value)
			}
		}
		press(app, "tab", "100", "enter")
		assertFieldError(t, app, field, "want a number")
		errLine := errorLine(shown(app))
		if strings.Contains(errLine, "not negative") || strings.Contains(errLine, "greater than zero") {
			t.Fatalf("missing %s used the wrong phrase: %q", field, errLine)
		}
		press(app, "esc")
		assertNameIsNew(t, app, name)
		assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	}
}

func TestMealNegativeNutrientRefusedZeroAllowed(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	fields := []string{"calories", "protein", "carbohydrate", "fat"}
	for i, field := range fields {
		name := "Neg" + string(rune('A'+i))
		values := []string{"1", "1", "1", "1"}
		values[i] = "-1"
		press(app, "a", name)
		for _, value := range values {
			press(app, "tab", value)
		}
		press(app, "tab", "1", "enter")
		assertFieldError(t, app, field, "want a number that is not negative")
		press(app, "esc")
		assertNameIsNew(t, app, name)
	}
	saveNew(app, "Fig", "1", "1", "1", "0", "100")
	assertTotals(t, app, "1.0", "1.0", "1.0", "0.0")
	if strings.Contains(shown(app), "-0.0") {
		t.Fatalf("negative zero: %q", shown(app))
	}
}

func TestMealFractionalGramsAllowed(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "1", "1", "1", "1", "0.5")
	assertGramsOnly(t, shown(app), "Oats", "0.5")
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
}

func TestMealBadGramsRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	bad := []string{"0", "-1", "1/4", "1e2", ""}
	for i, grams := range bad {
		name := "Bad" + string(rune('A'+i))
		press(app, "a", name, "tab", "1", "tab", "1", "tab", "1", "tab", "1", "tab")
		if grams != "" {
			press(app, grams)
		}
		press(app, "enter")
		view := shown(app)
		if !strings.Contains(view, "error:") || !strings.Contains(view, "grams: want a number greater than zero") {
			t.Fatalf("grams %q: %q", grams, view)
		}
		press(app, "esc")
		assertNameIsNew(t, app, name)
		assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	}
}

func TestMealUnitTokensRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "]")
	tokens := []string{"100 g", "100g", "1 kg", "4 oz", "1 lb", "1 st", "1 cup", "8 fl oz", "1 tbsp", "1 tsp", "250 ml"}
	for _, token := range tokens {
		press(app, "a", "Oats", "tab")
		assertFocus(t, app, "grams")
		press(app, token, "enter")
		view := shown(app)
		if !strings.Contains(view, "error:") || !strings.Contains(view, "grams: want a number greater than zero") {
			t.Fatalf("token %q: %q", token, view)
		}
		press(app, "esc")
		assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
		assertNoSelection(t, app)
	}
	saveGrams(app, "Oats", "100")
	saveGrams(app, "Oats", "+100")
	saveGrams(app, "Oats", ".5")
	// nutrient * grams / 100, summed in float64, then %.1f.
	// 100 g + 100 g + 0.5 g of 200/10/20/5 prints 401.0, 20.1, 40.1, 10.0.
	assertTotals(t, app, "401.0", "20.1", "40.1", "10.0")
	press(app, "[")
	assertTotals(t, app, "80.0", "4.0", "8.0", "2.0")
	press(app, "]")
	for i := 0; i < 20; i++ {
		app.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	press(app, "enter")
	assertFieldShows(t, app, "grams", "100")
	press(app, "esc")
	for i := 0; i < 20; i++ {
		app.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	press(app, "enter")
	assertFieldShows(t, app, "grams", "0.5")
}

func TestMealRefusedGramsDoNotCreateAFood(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "a", "Rye", "tab", "1", "tab", "1", "tab", "1", "tab", "1", "tab", "0", "enter")
	view := shown(app)
	if !strings.Contains(view, "error:") || !strings.Contains(view, "grams: want a number greater than zero") {
		t.Fatalf("view = %q", view)
	}
	assertFocus(t, app, "grams")
	press(app, "esc")
	assertNameIsNew(t, app, "Rye")
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
}

func TestMealDisplayOneDecimal(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "1.25")
	view := shown(app)
	assertGramsOnly(t, view, "Oats", "1.2")
	if strings.Contains(view, "1.25") {
		t.Fatalf("idle screen shows stored grams: %q", view)
	}
	assertTotals(t, app, "2.5", "0.0", "0.0", "0.0")
	if strings.Contains(view, "-0.0") {
		t.Fatalf("negative zero: %q", view)
	}
	press(app, "enter")
	assertFieldShows(t, app, "grams", "1.25")
	press(app, "enter")
	view = shown(app)
	assertTotals(t, app, "2.5", "0.0", "0.0", "0.0")
	assertGramsOnly(t, view, "Oats", "1.2")
	press(app, "]")
	saveNew(app, "Pea", "1", "0", "0", "0", "25")
	assertTotals(t, app, "0.2", "0.0", "0.0", "0.0")
}

func TestMealDisplayUnitPoundsDoesNotScale(t *testing.T) {
	freezeMealToday(t)
	day := dateUTC(1990, time.November, 1)
	lb, err := units.FromKG(80, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	pounds := fmt.Sprintf("%.1f", lb)
	wrong, err := units.FromKG(100, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	scaled := fmt.Sprintf("%.1f", wrong)

	poundApp, _, _, _ := openMealFixture(t, config.Config{DisplayUnit: units.Pound, TrackMacros: true}, day)
	if !strings.Contains(shown(poundApp), pounds) {
		t.Fatalf("list missing %s: %q", pounds, shown(poundApp))
	}
	openMeals(t, poundApp, mealToday)
	saveNew(poundApp, "Oats", "200", "10", "20", "5", "100")
	poundView := shown(poundApp)

	kgCfg := config.Default()
	kgCfg.TrackMacros = true
	kgApp, _, _, _ := openMealFixture(t, kgCfg, day)
	openMeals(t, kgApp, mealToday)
	saveNew(kgApp, "Oats", "200", "10", "20", "5", "100")
	kgView := shown(kgApp)

	for _, want := range []string{"100.0", "calories 200.0", "protein 10.0", "carbohydrate 20.0", "fat 5.0"} {
		if !strings.Contains(poundView, want) || !strings.Contains(kgView, want) {
			t.Fatalf("pound view %q\nkg view %q\nmissing %s", poundView, kgView, want)
		}
	}
	if strings.Contains(poundView, scaled) {
		t.Fatalf("meal scaled 100 g as kilograms to pounds (%s): %q", scaled, poundView)
	}
	assertGramsOnly(t, poundView, "Oats", "100.0")
}

func TestMealUpDownStopsAtEnds(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Corn", "1", "1", "1", "1", "10")
	saveNew(app, "Ale", "1", "1", "1", "1", "10")
	press(app, "]", "[")
	if !strings.Contains(selectedMealLine(shown(app)), "Corn") {
		t.Fatalf("first line not selected: %q", shown(app))
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if !strings.Contains(selectedMealLine(shown(app)), "Ale") {
		t.Fatalf("down = %q", shown(app))
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if !strings.Contains(selectedMealLine(shown(app)), "Ale") {
		t.Fatalf("down wrapped: %q", shown(app))
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(selectedMealLine(shown(app)), "Corn") {
		t.Fatalf("up = %q", shown(app))
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(selectedMealLine(shown(app)), "Corn") {
		t.Fatalf("up wrapped: %q", shown(app))
	}
	press(app, "j", "k")
	if !strings.Contains(selectedMealLine(shown(app)), "Corn") {
		t.Fatalf("j/k moved the line: %q", shown(app))
	}
}

func TestMealRemoveLineKeepsFood(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Corn", "1", "0", "0", "0", "10")
	saveNew(app, "Ale", "1", "0", "0", "0", "20")
	saveNew(app, "Bran", "1", "0", "0", "0", "30")
	app.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(selectedMealLine(shown(app)), "Ale") {
		t.Fatalf("second line = %q", shown(app))
	}
	press(app, "x")
	view := shown(app)
	if strings.Contains(view, "Ale") {
		t.Fatalf("removed line still shown: %q", view)
	}
	if !strings.Contains(selectedMealLine(view), "Bran") {
		t.Fatalf("after remove selected = %q", view)
	}
	if strings.Index(view, "Corn") > strings.Index(view, "Bran") {
		t.Fatalf("order changed: %q", view)
	}
	press(app, "x")
	if !strings.Contains(selectedMealLine(shown(app)), "Corn") {
		t.Fatalf("new last line = %q", shown(app))
	}
	press(app, "x")
	assertNoSelection(t, app)
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	press(app, "a", "Corn", "tab")
	assertFocus(t, app, "grams")
}

func TestMealEditGramsCancelAndRefuse(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "enter")
	view := shown(app)
	if !strings.Contains(view, "Oats") || strings.Contains(view, "> name") {
		t.Fatalf("grams edit = %q", view)
	}
	assertFocus(t, app, "grams")
	press(app, "esc")
	assertGramsOnly(t, shown(app), "Oats", "40.0")
	press(app, "enter")
	replaceField(app, "10")
	press(app, "enter")
	assertGramsOnly(t, shown(app), "Oats", "10.0")
	assertTotals(t, app, "20.0", "1.0", "2.0", "0.5")
	press(app, "enter")
	replaceField(app, "0")
	press(app, "enter")
	view = shown(app)
	if !strings.Contains(view, "error:") || !strings.Contains(view, "grams: want a number greater than zero") {
		t.Fatalf("bad grams = %q", view)
	}
	assertFocus(t, app, "grams")
	press(app, "esc")
	assertGramsOnly(t, shown(app), "Oats", "10.0")
	assertTotals(t, app, "20.0", "1.0", "2.0", "0.5")
}

func TestMealEditFoodUpdatesEveryDate(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	press(app, "]")
	saveGrams(app, "Oats", "50")
	press(app, "[")
	press(app, "e")
	assertFocus(t, app, "calories per 100 g")
	replaceField(app, "100")
	press(app, "enter")
	assertMeals(t, app, mealToday)
	assertTotals(t, app, "100.0", "10.0", "20.0", "5.0")
	assertGramsOnly(t, shown(app), "Oats", "100.0")
	press(app, "]")
	assertTotals(t, app, "50.0", "5.0", "10.0", "2.5")
	assertGramsOnly(t, shown(app), "Oats", "50.0")
	press(app, "[")
	saveGrams(app, "Oats", "100")
	assertTotals(t, app, "200.0", "20.0", "40.0", "10.0")
}

func TestMealEditFoodEscKeepsNumbers(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	press(app, "]")
	saveGrams(app, "Oats", "50")
	press(app, "[")
	press(app, "e")
	replaceField(app, "100")
	press(app, "esc")
	assertTotals(t, app, "200.0", "10.0", "20.0", "5.0")
	press(app, "]")
	assertTotals(t, app, "100.0", "5.0", "10.0", "2.5")
}

func TestMealEditFoodDoesNotRename(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	press(app, "e")
	view := shown(app)
	if !strings.Contains(view, "Oats") || strings.Contains(view, "> name") {
		t.Fatalf("food edit = %q", view)
	}
	assertFocus(t, app, "calories per 100 g")
	press(app, "zzz")
	press(app, "enter")
	if !strings.Contains(shown(app), "Oats") || strings.Contains(shown(app), "Oatszzz") {
		t.Fatalf("typing renamed the food: %q", shown(app))
	}
	replaceField(app, "200")
	press(app, "enter")
	assertGramsOnly(t, shown(app), "Oats", "100.0")
	if strings.Contains(shown(app), "zzz") {
		t.Fatalf("new name stored: %q", shown(app))
	}
	assertNameIsNew(t, app, "zzz")
	press(app, "a", "Oats", "tab")
	assertFocus(t, app, "grams")
}

func TestMealNewLineIsSelected(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "1", "1", "1", "1", "10")
	if !strings.Contains(selectedMealLine(shown(app)), "Oats") {
		t.Fatalf("first save selected = %q", shown(app))
	}
	saveNew(app, "Rye", "1", "1", "1", "1", "20")
	view := shown(app)
	if !strings.Contains(selectedMealLine(view), "Rye") {
		t.Fatalf("second save selected = %q", view)
	}
	if strings.Index(view, "Oats") < 0 || strings.Index(view, "Oats") > strings.Index(view, "Rye") {
		t.Fatalf("order = %q", view)
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(selectedMealLine(shown(app)), "Oats") {
		t.Fatalf("up = %q", shown(app))
	}
}

func TestMealReopenShowsTheSameLines(t *testing.T) {
	app, store, meals, path := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "50")
	assertTotals(t, app, "100.0", "5.0", "10.0", "2.5")
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	closeMeal(meals)
	reopened, err := dailylog.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	again, err := meal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	reopen := config.Default()
	reopen.TrackMacros = true
	next := New(reopened, path, reopen)
	next.UseMeals(again)
	if msg := next.load(); msg != nil {
		next.Update(msg)
	}
	openMeals(t, next, mealToday)
	assertGramsOnly(t, shown(next), "Oats", "50.0")
	assertTotals(t, next, "100.0", "5.0", "10.0", "2.5")
	press(next, "]")
	assertMeals(t, next, "1990-11-11")
	assertTotals(t, next, "0.0", "0.0", "0.0", "0.0")
	assertNoSelection(t, next)
	if strings.Contains(shown(next), "Oats") {
		t.Fatalf("next day has a line: %q", shown(next))
	}
}

func TestMealQuitAndNameFieldQ(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	_, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if commandQuits(cmd) {
		t.Fatal("esc quit")
	}
	openMeals(t, app, mealToday)
	_, cmd = app.Update(letter('q'))
	if !commandQuits(cmd) {
		t.Fatal("q did not quit")
	}

	form, _, _, _ := mealApp(t)
	openMeals(t, form, mealToday)
	press(form, "a")
	_, cmd = form.Update(letter('q'))
	if commandQuits(cmd) {
		t.Fatal("q in the name quit")
	}
	_, cmd = form.Update(letter('f'))
	if commandQuits(cmd) {
		t.Fatal("f in the name quit")
	}
	view := shown(form)
	name := lineContaining(view, "name")
	if !strings.Contains(name, "q") || !strings.Contains(name, "f") {
		t.Fatalf("name = %q in %q", name, view)
	}
	if strings.Count(view, "meals ") != 1 {
		t.Fatalf("f opened another screen: %q", view)
	}
	_, cmd = form.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if got := cmd; got == nil || !commandQuits(got) {
		t.Fatal("ctrl+c did not quit")
	}
}

func TestMealBracketsWhileAddingDoNotMove(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "a", "[", "]")
	assertMeals(t, app, mealToday)
	press(app, "esc")
	saveNew(app, "Oats", "200", "10", "20", "5", "40")
	press(app, "enter", "[", "]")
	assertMeals(t, app, mealToday)
	press(app, "esc", "e", "[", "]")
	assertMeals(t, app, mealToday)
}

func TestMealHelpAndListHelp(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	if !strings.Contains(shown(app), listHelp) {
		t.Fatalf("list help = %q", shown(app))
	}
	empty, _, _, _ := mealApp(t)
	if !strings.Contains(shown(empty), emptyHelp) {
		t.Fatalf("empty help = %q", shown(empty))
	}
	empty.Update(loadErrMsg{err: fmt.Errorf("load daily log: boom")})
	failed := shown(empty)
	if !strings.Contains(failed, "q quit") || strings.Contains(failed, "n new") || strings.Contains(failed, "f meal") {
		t.Fatalf("failed list help = %q", failed)
	}

	openMeals(t, app, mealToday)
	if !strings.Contains(shown(app), mealHelp) {
		t.Fatalf("meal help = %q", shown(app))
	}
	press(app, "a")
	if !strings.Contains(shown(app), mealFormHelp) || !strings.Contains(shown(app), "meals "+mealToday) {
		t.Fatalf("add help = %q", shown(app))
	}
	press(app, "esc")
	saveNew(app, "Oats", "1", "1", "1", "1", "10")
	press(app, "enter")
	if !strings.Contains(shown(app), mealFormHelp) || !strings.Contains(shown(app), "> grams") {
		t.Fatalf("grams help = %q", shown(app))
	}
	press(app, "esc", "e")
	if !strings.Contains(shown(app), mealFormHelp) || !strings.Contains(shown(app), "> calories per 100 g") {
		t.Fatalf("food help = %q", shown(app))
	}
}

func TestMealInFlightFormSaveKeepsTheMealScreen(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, _, _, _ := mealApp(t, day)
	before := len(app.logs)
	press(app, "n")
	_, saveCmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("enter did not start a save")
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	_, mealCmd := app.Update(letter('f'))
	follow(t, app, mealCmd)
	_, reload := app.Update(saveCmd())
	assertMeals(t, app, mealToday)
	follow(t, app, reload)
	assertMeals(t, app, mealToday)
	if len(app.logs) <= before {
		t.Fatal("weight log did not reload")
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	assertListRow(t, app, day)
}

func TestMealInFlightMonthSaveKeepsTheMealScreen(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, _, _, _ := mealApp(t, day)
	press(app, "m")
	app.Update(letter('8'))
	app.Update(letter('0'))
	_, saveCmd := app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if saveCmd == nil {
		t.Fatal("tab did not start a month-cell save")
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	want := app.logs[app.cursor].Day
	_, mealCmd := app.Update(letter('f'))
	follow(t, app, mealCmd)
	_, reload := app.Update(saveCmd())
	assertMeals(t, app, mealToday)
	follow(t, app, reload)
	assertMeals(t, app, mealToday)
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if strings.Contains(shown(app), "arrows move") {
		t.Fatalf("esc showed the month sheet: %q", shown(app))
	}
	assertListRow(t, app, want)
}

func TestMealFormErrorDoesNotFailTheList(t *testing.T) {
	app, _, _, _ := mealApp(t, dateUTC(1990, time.November, 1))
	press(app, "n", "tab", "nope")
	_, saveCmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("enter did not start a save")
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	_, mealCmd := app.Update(letter('f'))
	follow(t, app, mealCmd)
	msg := saveCmd()
	if _, ok := msg.(loadErrMsg); !ok {
		t.Fatalf("save = %T, want loadErrMsg", msg)
	}
	app.Update(msg)
	assertMeals(t, app, mealToday)
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	view := shown(app)
	if !strings.Contains(view, listHelp) {
		t.Fatalf("list help = %q", view)
	}

	grams, _, _, _ := mealApp(t)
	openMeals(t, grams, mealToday)
	saveNew(grams, "Oats", "200", "10", "20", "5", "40")
	press(grams, "enter")
	replaceField(grams, "0")
	press(grams, "enter")
	if !strings.Contains(shown(grams), "grams: want a number greater than zero") {
		t.Fatalf("grams error = %q", shown(grams))
	}
	grams.Update(loadedMsg{logs: grams.logs})
	if !strings.Contains(shown(grams), "grams: want a number greater than zero") {
		t.Fatalf("reload cleared the grams error: %q", shown(grams))
	}
}

func TestMealTypingClearsTheFormError(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "a", "Oil", "tab", "884", "tab", "0", "tab", "0", "tab", "enter")
	assertFieldError(t, app, "fat", "want a number")
	press(app, "tab")
	if strings.Contains(shown(app), "error:") {
		t.Fatalf("tab left the error up: %q", shown(app))
	}
	assertFocus(t, app, "grams")
	press(app, "enter")
	assertFieldError(t, app, "fat", "want a number")
	press(app, "0.3")
	if strings.Contains(shown(app), "error:") {
		t.Fatalf("typing left the error up: %q", shown(app))
	}
	press(app, "tab", "13", "enter")
	assertTotals(t, app, "114.9", "0.0", "0.0", "0.0")
	press(app, "e", "tab", "tab", "tab")
	assertFieldShows(t, app, "fat per 100 g", "0.3")
}

func TestMealMacrosOffShowsOptionalFields(t *testing.T) {
	freezeMealToday(t)
	app, _, _, _ := openMealFixture(t, config.Default())
	openMeals(t, app, mealToday)
	view := shown(app)
	if !strings.Contains(view, "calories 0.0") {
		t.Fatalf("totals = %q", view)
	}
	for _, word := range []string{"protein", "carbohydrate", "fat"} {
		if strings.Contains(view, word) {
			t.Fatalf("default screen shows %s: %q", word, view)
		}
	}
	press(app, "a", "Oil", "tab")
	assertFocus(t, app, "calories per 100 g")
	view = shown(app)
	for _, label := range []string{"protein per 100 g", "carbohydrate per 100 g", "fat per 100 g"} {
		if !strings.Contains(view, label) {
			t.Fatalf("form hid %s: %q", label, view)
		}
	}
	press(app, "enter")
	assertFieldError(t, app, "calories", "want a number")
	press(app, "884", "tab")
	assertFocus(t, app, "protein per 100 g")
	press(app, "12.5", "tab", "tab", "no", "enter")
	assertFieldError(t, app, "fat", "want a number")
	replaceField(app, "")
	press(app, "tab", "100", "enter")
	view = shown(app)
	if !strings.Contains(view, "calories 884.0") {
		t.Fatalf("calories = %q", view)
	}
	for _, word := range []string{"protein", "carbohydrate", "fat"} {
		if strings.Contains(view, word) {
			t.Fatalf("saved screen shows %s: %q", word, view)
		}
	}
	press(app, "e")
	assertFocus(t, app, "calories per 100 g")
	assertFieldShows(t, app, "calories per 100 g", "884")
	press(app, "tab")
	assertFieldShows(t, app, "protein per 100 g", "12.5")
	press(app, "tab")
	assertFieldShows(t, app, "carbohydrate per 100 g", "0")
	press(app, "tab")
	assertFieldShows(t, app, "fat per 100 g", "0")
}

func TestMealMacrosOffKeepsStoredMacros(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oil", "884", "10", "20", "5", "100")
	press(app, "esc")
	app.cfg.TrackMacros = false
	press(app, "f", "e", "tab")
	assertFieldShows(t, app, "protein per 100 g", "10")
	press(app, "tab")
	assertFieldShows(t, app, "carbohydrate per 100 g", "20")
	press(app, "tab")
	assertFieldShows(t, app, "fat per 100 g", "5")
	press(app, "tab")
	replaceField(app, "442")
	press(app, "enter")
	view := shown(app)
	if !strings.Contains(view, "calories 442.0") || strings.Contains(view, "protein") {
		t.Fatalf("after calories edit: %q", view)
	}
	press(app, "esc")
	app.cfg.TrackMacros = true
	press(app, "f", "e")
	assertFieldShows(t, app, "calories per 100 g", "442")
	press(app, "tab")
	assertFieldShows(t, app, "protein per 100 g", "10")
	press(app, "tab")
	assertFieldShows(t, app, "carbohydrate per 100 g", "20")
	press(app, "tab")
	assertFieldShows(t, app, "fat per 100 g", "5")
}

func TestMealFoodEditRoundTrip(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "1.25", "0", "0", "0", "100")
	press(app, "]")
	saveGrams(app, "Oats", "100")
	press(app, "[")
	assertTotals(t, app, "1.2", "0.0", "0.0", "0.0")
	press(app, "e")
	assertFieldShows(t, app, "calories per 100 g", "1.25")
	press(app, "enter")
	assertTotals(t, app, "1.2", "0.0", "0.0", "0.0")
	press(app, "e")
	assertFieldShows(t, app, "calories per 100 g", "1.25")
	press(app, "enter")
	press(app, "e")
	replaceField(app, "")
	press(app, "enter")
	assertFieldError(t, app, "calories", "want a number")
	press(app, "esc")
	assertTotals(t, app, "1.2", "0.0", "0.0", "0.0")
	press(app, "]")
	assertTotals(t, app, "1.2", "0.0", "0.0", "0.0")
	press(app, "[", "e")
	assertFieldShows(t, app, "calories per 100 g", "1.25")
}

func TestMealEnterWaitsForTab(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	press(app, "a", "Rye", "tab", "9", "tab", "9", "tab", "9", "tab", "9", "tab", "50", "tab")
	assertFocus(t, app, "name")
	replaceField(app, "Oats")
	press(app, "enter")
	view := shown(app)
	if !strings.Contains(view, "error:") || !strings.Contains(view, "name: press tab") {
		t.Fatalf("view = %q", view)
	}
	press(app, "esc")
	assertTotals(t, app, "200.0", "10.0", "20.0", "5.0")
	if strings.Contains(shown(app), "Rye") {
		t.Fatalf("refused enter wrote Rye: %q", shown(app))
	}
	press(app, "a", "Rye", "tab", "9", "tab", "9", "tab", "9", "tab", "9", "tab", "50", "tab")
	replaceField(app, "Oats")
	press(app, "enter")
	if !strings.Contains(shown(app), "name: press tab") {
		t.Fatalf("second attempt = %q", shown(app))
	}
	press(app, "tab")
	assertFocus(t, app, "grams")
	replaceField(app, "40")
	press(app, "enter")
	assertTotals(t, app, "280.0", "14.0", "28.0", "7.0")
	if strings.Contains(shown(app), "Rye") {
		t.Fatalf("Rye was stored: %q", shown(app))
	}

	press(app, "a", "Oats", "tab")
	assertFocus(t, app, "grams")
	press(app, "tab")
	assertFocus(t, app, "name")
	replaceField(app, "Rye")
	press(app, "enter")
	if !strings.Contains(shown(app), "error:") || !strings.Contains(shown(app), "name: press tab") {
		t.Fatalf("new name without tab = %q", shown(app))
	}
	press(app, "tab")
	assertFocus(t, app, "calories per 100 g")
	press(app, "1", "tab", "1", "tab", "1", "tab", "1", "tab", "100", "enter")
	if !strings.Contains(shown(app), "Rye") {
		t.Fatalf("tab then enter did not store Rye: %q", shown(app))
	}
	assertTotals(t, app, "281.0", "15.0", "29.0", "8.0")
}

func TestMealNutrientGrammar(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	bad := []string{"1e2", "inf", "NaN", "1/4", "1 cup"}
	for i, calories := range bad {
		name := "Tok" + string(rune('A'+i))
		press(app, "a", name, "tab", calories, "tab", "1", "tab", "1", "tab", "1", "tab", "1", "enter")
		assertFieldError(t, app, "calories", "want a number")
		errLine := errorLine(shown(app))
		if strings.Contains(errLine, "not negative") {
			t.Fatalf("token %q: %q", calories, errLine)
		}
		press(app, "esc")
		assertNameIsNew(t, app, name)
	}
	saveNew(app, "Fig", "+10", "1", "1", "1", "100")
	assertTotals(t, app, "10.0", "1.0", "1.0", "1.0")
	press(app, "]")
	saveNew(app, "Pea", "1.25", "0", "0", "0", "100")
	assertTotals(t, app, "1.2", "0.0", "0.0", "0.0")
	press(app, "e")
	assertFieldShows(t, app, "calories per 100 g", "1.25")
	press(app, "esc", "[", "e")
	replaceField(app, "1e2")
	press(app, "enter")
	assertFieldError(t, app, "calories", "want a number")
	press(app, "esc")
	assertTotals(t, app, "10.0", "1.0", "1.0", "1.0")
	press(app, "e")
	assertFieldShows(t, app, "calories per 100 g", "10")
	press(app, "esc")
	saveNew(app, "Oil", "884", "0", "0", "0.3", "100")
	assertTotals(t, app, "894.0", "1.0", "1.0", "1.3")
	press(app, "e", "tab", "tab", "tab")
	assertFieldShows(t, app, "fat per 100 g", "0.3")
}

func TestMealDocs(t *testing.T) {
	t.Parallel()
	keys := readRepo(t, "docs/reference/keys.md")
	fRow := tableRow(keys, "f")
	if fRow == "" || !strings.Contains(fRow, "list") || !strings.Contains(strings.ToLower(fRow), "today") || !strings.Contains(fRow, "meal") {
		t.Fatalf("f row = %q", fRow)
	}
	if row := tableRow(keys, "a"); row == "" || !strings.Contains(strings.ToLower(row), "add") {
		t.Fatalf("a row = %q", row)
	}
	if !rowHas(keys, "Enter", "grams") {
		t.Fatal("keys.md has no meal Enter row for grams")
	}
	eRow := tableRow(keys, "e")
	if eRow == "" || !strings.Contains(eRow, "every meal date") || !strings.Contains(strings.ToLower(eRow), "food") {
		t.Fatalf("e row = %q", eRow)
	}
	if strings.Contains(strings.ToLower(eRow), "confirm") {
		t.Fatalf("e row asks for a confirm: %q", eRow)
	}
	if row := tableRow(keys, "x"); row == "" || !strings.Contains(strings.ToLower(row), "remove") {
		t.Fatalf("x row = %q", row)
	}
	if !bracketRowMentions(keys, "meal") {
		t.Fatal("keys.md has no meal bracket row")
	}
	if strings.Contains(strings.ToLower(keys), "list") && bracketRowMentions(keys, "list") {
		t.Fatal("keys.md gained a list bracket row")
	}
	for _, keep := range []string{
		"| `[` / `]` | month or monthly chart | previous / next month |",
		"| `[` / `]` | long-term chart | previous / next kind |",
	} {
		if !strings.Contains(keys, keep) {
			t.Fatalf("keys.md lost %s", keep)
		}
	}
	qRow := tableRow(keys, "q")
	if qRow == "" || !strings.Contains(qRow, "meal") {
		t.Fatalf("q row = %q", qRow)
	}
	escRow := tableRow(keys, "Esc")
	if escRow == "" || !strings.Contains(escRow, "meal") {
		t.Fatalf("Esc row = %q", escRow)
	}
	if !rowHas(keys, "Tab", "meal") {
		t.Fatal("keys.md has no meal-form Tab row")
	}
	if !rowHas(keys, "Enter", "save", "meal") {
		t.Fatal("keys.md has no meal-form Enter row")
	}
	if !rowHas(keys, "Up", "meal") && !rowHas(keys, "Down", "meal") {
		t.Fatal("keys.md has no meal up/down row")
	}

	how := readRepo(t, "docs/how-to/meal-day.md")
	lower := strings.ToLower(how)
	if !strings.Contains(how, "every meal date") {
		t.Fatal("meal how-to does not say an edit changes every meal date")
	}
	if strings.Contains(lower, "confirm") {
		t.Fatal("meal how-to asks for a confirm")
	}
	for _, word := range []string{"allowance", "volume", "recipe", "starter", "later", "per 100"} {
		if !strings.Contains(lower, word) {
			t.Fatalf("meal how-to missing %s", word)
		}
	}
	if !strings.Contains(lower, "renam") || !strings.Contains(lower, "delet") {
		t.Fatal("meal how-to does not say renaming or deleting a food is later")
	}
	if !strings.Contains(lower, "display_unit") && !strings.Contains(lower, "display unit") {
		t.Fatal("meal how-to does not say display_unit leaves grams alone")
	}
	readme := readRepo(t, "docs/how-to/README.md")
	if !strings.Contains(readme, "meal-day.md") {
		t.Fatal("how-to index does not link meal-day.md")
	}
}

func rowText(view, day string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, day) && !strings.Contains(line, "meals") {
			return line
		}
	}
	return ""
}

func errorLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "error:") {
			return line
		}
	}
	return ""
}

func lineContaining(view, needle string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

func tableRow(md, key string) string {
	rows := tableRows(md, key)
	if len(rows) == 0 {
		return ""
	}
	return rows[0]
}

func tableRows(md, key string) []string {
	var rows []string
	for _, line := range strings.Split(md, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		if keyMatches(cells[1], key) {
			rows = append(rows, line)
		}
	}
	return rows
}

func keyMatches(cell, key string) bool {
	raw := strings.ReplaceAll(cell, "`", "")
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == ','
	}) {
		if strings.EqualFold(strings.TrimSpace(part), key) {
			return true
		}
	}
	return false
}

func rowHas(md, key string, parts ...string) bool {
	for _, line := range tableRows(md, key) {
		lower := strings.ToLower(line)
		ok := true
		for _, part := range parts {
			if !strings.Contains(lower, strings.ToLower(part)) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func bracketRowMentions(md, screen string) bool {
	for _, line := range strings.Split(md, "\n") {
		if !strings.Contains(line, "[") || !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		if strings.Contains(cells[1], "[") && strings.Contains(cells[2], screen) {
			return true
		}
	}
	return false
}
