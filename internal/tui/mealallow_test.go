// Allowance scenarios drive the meal screen from the approved plan.
// They fail until that screen stores one calorie number for the meal
// date and paints the comparison. This file does not change the food
// form, and it does not treat a chart deficit as the allowance.
package tui

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/meal"
	"github.com/jglueckstein/hdtools/internal/units"

	_ "modernc.org/sqlite"
)

const mealAllowanceEditHelp = "enter save   esc cancel"

func TestMealUnsetShowsNoAllowance(t *testing.T) {
	freezeMealToday(t)
	app, _, _, _ := openMealFixture(t, config.Default())
	openMeals(t, app, mealToday)
	view := shown(app)
	if !strings.Contains(view, "calories 0.0") || !strings.Contains(view, "no allowance") {
		t.Fatalf("unset = %q", view)
	}
	for _, word := range []string{"protein", "carbohydrate", "fat", "remaining"} {
		if strings.Contains(view, word) {
			t.Fatalf("unset shows %s: %q", word, view)
		}
	}
	saveNew(app, "Oats", "200", "", "", "", "100")
	view = shown(app)
	if !strings.Contains(view, "calories 200.0") || !strings.Contains(view, "no allowance") {
		t.Fatalf("after food = %q", view)
	}
	for _, word := range []string{"protein", "carbohydrate", "fat", "remaining"} {
		if strings.Contains(view, word) {
			t.Fatalf("after food shows %s: %q", word, view)
		}
	}
}

func TestMealTrackedMacrosStillShowNoAllowance(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	assertTotals(t, app, "0.0", "0.0", "0.0", "0.0")
	if !strings.Contains(shown(app), "no allowance") || strings.Contains(shown(app), "remaining") {
		t.Fatalf("empty = %q", shown(app))
	}
	saveNew(app, "Oats", "200", "10", "20", "5", "100")
	assertTotals(t, app, "200.0", "10.0", "20.0", "5.0")
	if !strings.Contains(shown(app), "no allowance") || strings.Contains(shown(app), "remaining") {
		t.Fatalf("tracked = %q", shown(app))
	}
}

func TestMealRemainingUnder(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	press(app, "b", " 500 ", "enter")
	assertRemaining(t, app, "300.0")
	if strings.Contains(shown(app), "no allowance") {
		t.Fatalf("under still unset: %q", shown(app))
	}
	press(app, "b")
	replaceField(app, "+500")
	press(app, "enter")
	assertRemaining(t, app, "300.0")
}

func TestMealRemainingOver(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	enterAllowance(app, "100")
	assertRemaining(t, app, "-100.0")
}

func TestMealRemainingExactZero(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	enterAllowance(app, "200")
	assertRemaining(t, app, "0.0")
	view := shown(app)
	if strings.Contains(view, "-0.0") || strings.Contains(view, "no allowance") {
		t.Fatalf("exact = %q", view)
	}
}

func TestMealRemainingHalfEven(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	enterAllowance(app, "1.25")
	assertRemaining(t, app, "1.2")
	view := shown(app)
	if strings.Contains(view, "remaining 1.25") || strings.Contains(view, "remaining 1.3") {
		t.Fatalf("half-even = %q", view)
	}
	press(app, "b")
	assertFieldShows(t, app, "allowance", "1.25")
	press(app, "enter")
	assertRemaining(t, app, "1.2")
	press(app, "b")
	assertFieldShows(t, app, "allowance", "1.25")
	press(app, "esc")
	enterAllowance(app, "0.25")
	assertRemaining(t, app, "0.2")
	if strings.Contains(shown(app), "remaining 0.3") {
		t.Fatalf("0.25 painted 0.3: %q", shown(app))
	}
	press(app, "]")
	saveNew(app, "Pea", "1.25", "0", "0", "0", "100")
	if !labeledNumber(shown(app), "calories", "1.2") {
		t.Fatalf("calories = %q", shown(app))
	}
	enterAllowance(app, "2.5")
	assertRemaining(t, app, "1.2")
	if strings.Contains(shown(app), "remaining 1.3") {
		t.Fatalf("painted tenth subtracted: %q", shown(app))
	}
	press(app, "[")
	assertRemaining(t, app, "0.2")
}

func TestMealRemainingZeroTenth(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Dust", "0.04", "0", "0", "0", "100")
	enterAllowance(app, "0")
	if !labeledNumber(shown(app), "calories", "0.0") {
		t.Fatalf("calories = %q", shown(app))
	}
	assertRemaining(t, app, "0.0")
	view := shown(app)
	if strings.Contains(view, "no allowance") || strings.Contains(view, "-0.0") {
		t.Fatalf("sub-tenth = %q", view)
	}
}

func TestMealZeroAllowanceIsSet(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "40", "0", "0", "0", "100")
	for _, text := range []string{"0", "0.0", "+0"} {
		enterAllowance(app, text)
		assertRemaining(t, app, "-40.0")
		if strings.Contains(shown(app), "no allowance") {
			t.Fatalf("%s cleared the allowance: %q", text, shown(app))
		}
	}
	press(app, "b")
	assertFieldShows(t, app, "allowance", "0")
	line := lineContaining(shown(app), "> allowance")
	if strings.Contains(line, "0.0") {
		t.Fatalf("field = %q", line)
	}
	press(app, "esc", "x")
	if !labeledNumber(shown(app), "calories", "0.0") {
		t.Fatalf("after remove = %q", shown(app))
	}
	assertRemaining(t, app, "0.0")
	view := shown(app)
	if strings.Contains(view, "no allowance") || strings.Contains(view, "-0.0") {
		t.Fatalf("zero day = %q", view)
	}
}

func TestMealNegativeAllowanceRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	enterAllowance(app, "500")
	assertRemaining(t, app, "300.0")
	press(app, "b", "-1", "enter")
	assertFieldError(t, app, "allowance", "want a number that is not negative")
	assertFocus(t, app, "allowance")
	press(app, "5")
	if strings.Contains(shown(app), "error:") {
		t.Fatalf("typing left the error: %q", shown(app))
	}
	assertFocus(t, app, "allowance")
	press(app, "esc")
	assertRemaining(t, app, "300.0")

	fresh, _, _, _ := mealApp(t)
	openMeals(t, fresh, mealToday)
	press(fresh, "b", "-0", "enter")
	assertFieldError(t, fresh, "allowance", "want a number that is not negative")
	press(fresh, "esc")
	if !strings.Contains(shown(fresh), "no allowance") {
		t.Fatalf("unset after -0 = %q", shown(fresh))
	}
}

func TestMealBadAllowanceRefused(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	enterAllowance(app, "500")
	bad := []string{
		"2000 kcal", "2000 cal", "2000 g", "2000g", "1 kg", "4 oz",
		"1 lb", "1 st", "1 cup", "8 fl oz", "1 tbsp", "1 tsp",
		"250 ml", "1/4", "1e2", "inf", "NaN",
	}
	for _, text := range bad {
		press(app, "b")
		replaceField(app, text)
		press(app, "enter")
		assertFieldError(t, app, "allowance", "want a number")
		if strings.Contains(shown(app), "not negative") {
			t.Fatalf("%s used the negative error: %q", text, shown(app))
		}
		assertFocus(t, app, "allowance")
		press(app, "esc")
		assertRemaining(t, app, "300.0")
	}
}

func TestMealBlankAllowanceClears(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	enterAllowance(app, "500")
	for _, text := range []string{"", "   "} {
		press(app, "b")
		replaceField(app, text)
		press(app, "enter")
		view := shown(app)
		if !strings.Contains(view, "no allowance") || strings.Contains(view, "remaining") || strings.Contains(view, "want a number") {
			t.Fatalf("clear %q = %q", text, view)
		}
		if !strings.Contains(view, "Oats") {
			t.Fatalf("clear removed the food: %q", view)
		}
		press(app, "b")
		assertAllowanceEmpty(t, app)
		press(app, "esc")
		if text == "" {
			enterAllowance(app, "500")
		}
	}
}

func TestMealAllowanceOnAnEmptyDay(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	view := shown(app)
	if !strings.Contains(view, "calories 0.0") || !strings.Contains(view, "no allowance") {
		t.Fatalf("empty = %q", view)
	}
	enterAllowance(app, "1800")
	assertRemaining(t, app, "1800.0")
	enterAllowance(app, ".5")
	assertRemaining(t, app, "0.5")
	press(app, "b")
	assertFieldShows(t, app, "allowance", "0.5")
}

func TestMealAllowanceDoesNotCarry(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, store, meals, path := mealApp(t, day)
	openMeals(t, app, mealToday)
	enterAllowance(app, "1800")
	assertRemaining(t, app, "1800.0")
	press(app, "]")
	assertMeals(t, app, "1990-11-11")
	view := shown(app)
	if !strings.Contains(view, "calories 0.0") || !strings.Contains(view, "no allowance") {
		t.Fatalf("next day = %q", view)
	}
	if strings.Contains(view, "1800") || strings.Contains(view, "remaining") {
		t.Fatalf("allowance carried: %q", view)
	}
	press(app, "[")
	assertRemaining(t, app, "1800.0")

	got, err := store.Get(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != 80 || got.Note != "" {
		t.Fatalf("daily log changed: %+v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 0600", info.Mode().Perm())
	}
	assertOneDatabase(t, path)
	assertDailyLogColumns(t, path)

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
	t.Cleanup(func() { closeMeal(again) })
	cfg := config.Default()
	cfg.TrackMacros = true
	next := New(reopened, path, cfg)
	next.UseMeals(again)
	if msg := next.load(); msg != nil {
		next.Update(msg)
	}
	openMeals(t, next, mealToday)
	assertRemaining(t, next, "1800.0")
	press(next, "]")
	if !strings.Contains(shown(next), "no allowance") || strings.Contains(shown(next), "remaining") {
		t.Fatalf("reopened next day = %q", shown(next))
	}
}

func TestMealAllowanceLeavesTheChartCaption(t *testing.T) {
	day1 := dateUTC(1990, time.November, 1)
	day2 := dateUTC(1990, time.November, 2)
	app, store, _, _ := mealApp(t, day1, day2)
	list := shown(app)
	row1 := rowText(list, "1990-11-01")
	row2 := rowText(list, "1990-11-02")
	press(app, "c")
	monthly := deficitLine(t, app)
	press(app, "esc", "l")
	longTerm := deficitLine(t, app)
	press(app, "esc", "f")
	assertMeals(t, app, mealToday)
	enterAllowance(app, "1800")
	assertRemaining(t, app, "1800.0")
	if strings.Contains(shown(app), "Daily deficit") {
		t.Fatalf("meal screen has a deficit: %q", shown(app))
	}
	press(app, "esc", "c")
	if got := deficitLine(t, app); got != monthly {
		t.Fatalf("monthly caption = %q, want %q", got, monthly)
	}
	press(app, "esc", "l")
	if got := deficitLine(t, app); got != longTerm {
		t.Fatalf("long caption = %q, want %q", got, longTerm)
	}
	press(app, "esc")
	after := shown(app)
	if rowText(after, "1990-11-01") != row1 || rowText(after, "1990-11-02") != row2 {
		t.Fatalf("list changed\nbefore %q\n%q\nafter %q", row1, row2, after)
	}
	assertNoDay(t, store, dateUTC(1990, time.November, 10))
}

func TestMealAllowanceIgnoresDisplayUnit(t *testing.T) {
	freezeMealToday(t)
	cfg := config.Config{DisplayUnit: units.Pound}
	app, _, _, _ := openMealFixture(t, cfg)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "100", "", "", "", "100")
	enterAllowance(app, "2000")
	view := shown(app)
	assertGramsOnly(t, view, "Oats", "100.0")
	if !labeledNumber(view, "calories", "100.0") || !labeledNumber(view, "remaining", "1900.0") {
		t.Fatalf("pounds = %q", view)
	}
	for _, kg := range []float64{100, 1900} {
		lb, err := units.FromKG(kg, units.Pound)
		if err != nil {
			t.Fatal(err)
		}
		scaled := fmt.Sprintf("%.1f", lb)
		if strings.Contains(view, scaled) {
			t.Fatalf("scaled %v kg to %s: %q", kg, scaled, view)
		}
	}
}

func TestMealFoodsSaveWhileUnset(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	saveNew(app, "Oats", "200", "0", "0", "0", "100")
	view := shown(app)
	if !strings.Contains(view, "Oats") || !labeledNumber(view, "calories", "200.0") || !strings.Contains(view, "no allowance") {
		t.Fatalf("unset save = %q", view)
	}
	if strings.Contains(view, "remaining") {
		t.Fatalf("unset shows remaining: %q", view)
	}
	press(app, "b")
	assertAllowanceEmpty(t, app)
}

func TestMealAllowanceKey(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	if !strings.Contains(shown(app), mealHelp) {
		t.Fatalf("idle help = %q", shown(app))
	}
	press(app, "b")
	assertAllowanceEditor(t, app, mealToday)
	assertAllowanceEmpty(t, app)
	press(app, "esc")
	if !strings.Contains(shown(app), "no allowance") {
		t.Fatalf("esc stored an allowance: %q", shown(app))
	}

	saveNew(app, "Oats", "100", "0", "0", "0", "100")
	enterAllowance(app, "500")
	press(app, "b")
	assertAllowanceEditor(t, app, mealToday)
	assertFieldShows(t, app, "allowance", "500")
	if strings.Contains(lineContaining(shown(app), "> allowance"), "500.0") {
		t.Fatalf("field shows a tenth: %q", shown(app))
	}
	press(app, "enter")
	press(app, "b")
	assertFieldShows(t, app, "allowance", "500")
	press(app, "esc")

	press(app, "a")
	assertFocus(t, app, "name")
	if !strings.Contains(shown(app), mealFormHelp) {
		t.Fatalf("add help = %q", shown(app))
	}
	press(app, "esc", "enter")
	assertFocus(t, app, "grams")
	press(app, "esc", "f")
	assertMeals(t, app, mealToday)
	if strings.Count(shown(app), "meals ") != 1 {
		t.Fatalf("f opened another screen: %q", shown(app))
	}
	press(app, "esc")
	openMeals(t, app, mealToday)
}

func TestMealAllowanceFieldIsText(t *testing.T) {
	app, _, _, _ := mealApp(t)
	openMeals(t, app, mealToday)
	press(app, "b")
	for _, key := range []string{"q", "a", "f", "b", "[", "]"} {
		_, cmd := app.Update(letter(rune(key[0])))
		if commandQuits(cmd) {
			t.Fatalf("%s quit", key)
		}
	}
	line := lineContaining(shown(app), "> allowance")
	for _, key := range []string{"q", "a", "f", "b", "[", "]"} {
		if !strings.Contains(line, key) {
			t.Fatalf("field %q missing %s in %q", line, key, shown(app))
		}
	}
	if !strings.Contains(shown(app), "meals "+mealToday) {
		t.Fatalf("date moved: %q", shown(app))
	}
	press(app, "esc")
	if !strings.Contains(shown(app), "no allowance") || strings.Contains(shown(app), "remaining") {
		t.Fatalf("esc wrote the field: %q", shown(app))
	}
	press(app, "b")
	_, cmd := app.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if !commandQuits(cmd) {
		t.Fatal("ctrl+c did not quit")
	}
}

func TestMealAllowanceDocs(t *testing.T) {
	t.Parallel()
	keys := readRepo(t, "docs/reference/keys.md")
	if !rowHas(keys, "b", "meal", "allowance") {
		t.Fatal("keys.md has no meal allowance row for b")
	}
	fRows := tableRows(keys, "f")
	if len(fRows) != 1 || !strings.Contains(fRows[0], "list") {
		t.Fatalf("f rows = %v", fRows)
	}
	how := readRepo(t, "docs/how-to/meal-day.md")
	lower := strings.ToLower(how)
	if strings.Contains(how, "An allowance, a volume, a recipe, and a starter set are later") {
		t.Fatal("how-to still calls the allowance a later slice")
	}
	for _, word := range []string{"`b`", "no allowance", "unrounded", "2.5", "1.25", "1.2", "config.toml", "display_unit"} {
		if !strings.Contains(how, word) {
			t.Fatalf("how-to missing %s", word)
		}
	}
	if !strings.Contains(lower, "daily deficit") {
		t.Fatal("how-to does not name Daily Deficit")
	}
	if strings.Contains(lower, "subtract the painted") {
		t.Fatal("how-to tells the reader to subtract the painted calorie text")
	}
	for _, word := range []string{"volume", "recipe", "starter", "later", "renam", "delet"} {
		if !strings.Contains(lower, word) {
			t.Fatalf("how-to dropped %s", word)
		}
	}
	if !strings.Contains(readRepo(t, "docs/how-to/README.md"), "meal-day.md") {
		t.Fatal("how-to index does not link the meal page")
	}
	if strings.Contains(strings.ToLower(readRepo(t, "docs/reference/config.md")), "allowance") {
		t.Fatal("config reference gained an allowance key")
	}
}

func TestMealAllowanceErrorSurvivesReload(t *testing.T) {
	day := dateUTC(1990, time.November, 1)
	app, _, _, _ := mealApp(t, day)
	press(app, "n")
	_, saveCmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("enter did not start a save")
	}
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	_, mealCmd := app.Update(letter('f'))
	follow(t, app, mealCmd)
	press(app, "b", "-1", "enter")
	assertFieldError(t, app, "allowance", "want a number that is not negative")
	_, reload := app.Update(saveCmd())
	follow(t, app, reload)
	assertFieldError(t, app, "allowance", "want a number that is not negative")
	assertFocus(t, app, "allowance")
	assertMeals(t, app, mealToday)
	app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !strings.Contains(shown(app), listHelp) {
		t.Fatalf("esc help = %q", shown(app))
	}
}

func enterAllowance(app *App, value string) {
	press(app, "b")
	replaceField(app, value)
	press(app, "enter")
}

func assertRemaining(t *testing.T, app *App, value string) {
	t.Helper()
	view := shown(app)
	if !labeledNumber(view, "remaining", value) {
		t.Fatalf("remaining %s missing in %q", value, view)
	}
}

func assertAllowanceEmpty(t *testing.T, app *App) {
	t.Helper()
	line := lineContaining(shown(app), "> allowance")
	if line == "" {
		t.Fatalf("no allowance field: %q", shown(app))
	}
	rest := strings.Replace(line, "allowance", "", 1)
	if strings.ContainsAny(rest, "0123456789") {
		t.Fatalf("field not empty: %q", line)
	}
}

func assertAllowanceEditor(t *testing.T, app *App, day string) {
	t.Helper()
	view := shown(app)
	if !strings.Contains(view, "meals "+day) || !strings.Contains(view, "> allowance") {
		t.Fatalf("editor = %q", view)
	}
	if !strings.Contains(view, mealAllowanceEditHelp) || strings.Contains(view, "tab") {
		t.Fatalf("editor help = %q", view)
	}
	for _, word := range []string{"calories", "no allowance", "remaining"} {
		if strings.Contains(view, word) {
			t.Fatalf("editor shows %s: %q", word, view)
		}
	}
}

func deficitLine(t *testing.T, app *App) string {
	t.Helper()
	line := lineContaining(shown(app), "Daily deficit:")
	if line == "" {
		t.Fatalf("no deficit caption: %q", shown(app))
	}
	return line
}

func assertOneDatabase(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Base(path)
	for _, entry := range entries {
		name := entry.Name()
		switch name {
		case base, base + "-wal", base + "-shm", base + "-journal":
		default:
			t.Fatalf("extra database file %s", name)
		}
	}
}

func assertDailyLogColumns(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	rows, err := db.Query(`PRAGMA table_info(daily_log)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := []string{"day", "weight", "sleep_hours", "steps", "workout", "note"}
	if len(names) != len(want) {
		t.Fatalf("daily_log columns = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("daily_log columns = %v", names)
		}
	}
}
