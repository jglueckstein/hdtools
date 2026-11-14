package meal

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/dailylog"

	_ "modernc.org/sqlite"
)

func TestMealMigrationKeepsDailyLog(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "hdtools.db")
	store, err := dailylog.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(1990, 11, 1, 15, 4, 0, 0, time.UTC)
	weight := 80.5
	log, err := dailylog.New(day, &weight, 7.5, 1234, true, "kept")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 0600", before.Mode().Perm())
	}
	namesBefore := dirNames(t, filepath.Dir(path))

	meals, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { closeStore(meals) })

	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if after.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 0600", after.Mode().Perm())
	}
	for _, name := range dirNames(t, filepath.Dir(path)) {
		if !sameDBFile(name, namesBefore) {
			t.Fatalf("extra database file %s", name)
		}
	}

	got, err := store.Get(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	if got.Day.Format(time.DateOnly) != "1990-11-01" || got.Weight == nil || *got.Weight != 80.5 ||
		got.SleepHours != 7.5 || got.Steps != 1234 || !got.Workout || got.Note != "kept" {
		t.Fatalf("daily log = %+v", got)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	gotCols := pragmaColumns(t, db, "daily_log")
	wantCols := []string{"day", "weight", "sleep_hours", "steps", "workout", "note"}
	if !sameStrings(gotCols, wantCols) {
		t.Fatalf("daily_log columns = %v, want %v", gotCols, wantCols)
	}
	tables := userTables(t, db)
	wantTables := []string{"daily_log", "food", "meal_line"}
	if !sameStrings(tables, wantTables) {
		t.Fatalf("tables = %v, want %v", tables, wantTables)
	}

	// A food edit writes the food row. The line keeps grams and the
	// food id, and it has no nutrient column to hold a stale copy.
	foodCols := pragmaColumns(t, db, "food")
	for _, name := range []string{"id", "name", "calories", "protein", "carbohydrate", "fat"} {
		if !hasString(foodCols, name) {
			t.Fatalf("food columns = %v, missing %s", foodCols, name)
		}
	}
	lineCols := pragmaColumns(t, db, "meal_line")
	for _, name := range []string{"id", "day", "food_id", "grams"} {
		if !hasString(lineCols, name) {
			t.Fatalf("meal_line columns = %v, missing %s", lineCols, name)
		}
	}
	for _, name := range []string{"calories", "protein", "carbohydrate", "fat"} {
		if hasString(lineCols, name) {
			t.Fatalf("meal_line copies %s: %v", name, lineCols)
		}
	}
	res, err := db.Exec(`INSERT INTO food (name, calories, protein, carbohydrate, fat) VALUES (?, ?, ?, ?, ?)`,
		"Oats", 200.0, 10.0, 20.0, 5.0)
	if err != nil {
		t.Fatal(err)
	}
	foodID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO meal_line (day, food_id, grams) VALUES (?, ?, ?)`,
		"1990-11-10", foodID, 100.0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO meal_line (day, food_id, grams) VALUES (?, ?, ?)`,
		"1990-11-11", foodID, 50.0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE food SET calories = ? WHERE id = ?`, 100.0, foodID); err != nil {
		t.Fatal(err)
	}
	var calories float64
	if err := db.QueryRow(`SELECT calories FROM food WHERE id = ?`, foodID).Scan(&calories); err != nil {
		t.Fatal(err)
	}
	if calories != 100 {
		t.Fatalf("food calories = %v, want 100", calories)
	}
	rows, err := db.Query(`SELECT food_id, grams FROM meal_line ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var lines [][2]float64
	for rows.Next() {
		var id, grams float64
		if err := rows.Scan(&id, &grams); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, [2]float64{id, grams})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != [2]float64{float64(foodID), 100} || lines[1] != [2]float64{float64(foodID), 50} {
		t.Fatalf("lines = %v, want the same food id and the original grams", lines)
	}
}

func TestMealOpenMissingFileCreatesNothing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "missing.db")
	_, err := Open(path)
	if err == nil {
		t.Fatal("Open succeeded on a missing file")
	}
	// A blanket refusal is not "this file is not there". The stat
	// error has to come through, and the file still must not exist.
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Open = %v, want the missing file", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("Stat = %v, want no file", statErr)
	}
}

func closeStore(store *Store) {
	if store == nil {
		return
	}
	closer, ok := any(store).(interface{ Close() error })
	if ok {
		_ = closer.Close()
	}
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func sameDBFile(name string, before []string) bool {
	if hasString(before, name) {
		return true
	}
	return name == "hdtools.db" || name == "hdtools.db-wal" || name == "hdtools.db-shm" || name == "hdtools.db-journal"
}

func pragmaColumns(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return names
}

func userTables(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return names
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func hasString(list []string, needle string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}
	return false
}
