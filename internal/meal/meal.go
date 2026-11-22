// Package meal is one day's foods in the same SQLite file as the
// daily log. A line stores grams and which food, not a copy of the
// four per-100 g numbers, so editing the food changes every day
// that uses it. The same file holds one calorie allowance per
// civil date. No row means that date has none. A stored zero is
// an allowance of zero, not an empty day.
//
// Open does not create the database. The daily log owns that file,
// including mode 0600, and a missing path must stay missing. This
// package does not draw a screen, convert display units, or store
// a density or a recipe.
package meal

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Store is the food, meal_line, and meal_allowance tables in the
// daily-log file. A nil Store is what the TUI holds when those
// tables are not open. One connection keeps the foreign-key pragma
// on: the pool would otherwise open a second connection that had
// not run it.
type Store struct {
	db *sql.DB
}

// Food is one saved food. The four numbers are per 100 g.
type Food struct {
	ID           int64
	Name         string
	Calories     float64
	Protein      float64
	Carbohydrate float64
	Fat          float64
}

// Line is one weighed food on a civil day. The four numbers come
// from the food row at read time, so a later edit is already in
// the next read. Grams are the stored weight, not the tenth the
// screen paints.
type Line struct {
	ID           int64
	FoodID       int64
	Name         string
	Grams        float64
	Calories     float64
	Protein      float64
	Carbohydrate float64
	Fat          float64
}

// Open opens the meal tables in an existing daily-log file at path.
// A path that is not there yet is an error and is not created.
func Open(path string) (*Store, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("open meals %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("open meals %s: is a directory", path)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open meals %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open meals %s: %w", path, err)
	}
	return s, nil
}

// Close releases the connection. The daily log closes its own.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close meals: %w", err)
	}
	return nil
}

func (s *Store) migrate() error {
	// Foreign keys are per connection. The line points at the food
	// so a later delete of a food can be refused by the recipe
	// slice; this slice never deletes a food.
	if _, err := s.db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	const schema = `
CREATE TABLE IF NOT EXISTS food (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	calories REAL NOT NULL,
	protein REAL NOT NULL,
	carbohydrate REAL NOT NULL,
	fat REAL NOT NULL
);
CREATE TABLE IF NOT EXISTS meal_line (
	id INTEGER PRIMARY KEY,
	day TEXT NOT NULL,
	food_id INTEGER NOT NULL REFERENCES food(id),
	grams REAL NOT NULL
);
CREATE TABLE IF NOT EXISTS meal_allowance (
	day TEXT PRIMARY KEY,
	calories REAL NOT NULL
);`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate meals: %w", err)
	}
	return nil
}

// Lines returns the foods logged on day, oldest first. day is a
// civil date, YYYY-MM-DD. An unknown day is an empty slice.
func (s *Store) Lines(day string) ([]Line, error) {
	const q = `
SELECT meal_line.id, food.id, food.name, meal_line.grams,
	food.calories, food.protein, food.carbohydrate, food.fat
FROM meal_line
JOIN food ON food.id = meal_line.food_id
WHERE meal_line.day = ?
ORDER BY meal_line.id`
	rows, err := s.db.Query(q, day)
	if err != nil {
		return nil, fmt.Errorf("list meals %s: %w", day, err)
	}
	defer rows.Close()
	var lines []Line
	for rows.Next() {
		var ln Line
		if err := rows.Scan(&ln.ID, &ln.FoodID, &ln.Name, &ln.Grams, &ln.Calories, &ln.Protein, &ln.Carbohydrate, &ln.Fat); err != nil {
			return nil, fmt.Errorf("list meals %s: %w", day, err)
		}
		lines = append(lines, ln)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list meals %s: %w", day, err)
	}
	return lines, nil
}

// FoodByName returns the food whose stored name is exactly name.
// The caller trims. Case and middle spaces are part of the name.
func (s *Store) FoodByName(name string) (Food, bool, error) {
	const q = `
SELECT id, name, calories, protein, carbohydrate, fat
FROM food WHERE name = ?`
	var food Food
	err := s.db.QueryRow(q, name).Scan(&food.ID, &food.Name, &food.Calories, &food.Protein, &food.Carbohydrate, &food.Fat)
	if errors.Is(err, sql.ErrNoRows) {
		return Food{}, false, nil
	}
	if err != nil {
		return Food{}, false, fmt.Errorf("find food: %w", err)
	}
	return food, true, nil
}

// AddFoodLine inserts a new food and one line for day together.
// A failure leaves neither row. The name is stored as given.
func (s *Store) AddFoodLine(day, name string, calories, protein, carbohydrate, fat, grams float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("add food: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.Exec(`INSERT INTO food (name, calories, protein, carbohydrate, fat) VALUES (?, ?, ?, ?, ?)`,
		name, calories, protein, carbohydrate, fat)
	if err != nil {
		return fmt.Errorf("add food: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("add food: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO meal_line (day, food_id, grams) VALUES (?, ?, ?)`, day, id, grams); err != nil {
		return fmt.Errorf("add meal line: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("add food: %w", err)
	}
	return nil
}

// AddLine logs grams of a food that already exists. It does not
// write the food's four numbers.
func (s *Store) AddLine(day string, foodID int64, grams float64) error {
	if _, err := s.db.Exec(`INSERT INTO meal_line (day, food_id, grams) VALUES (?, ?, ?)`, day, foodID, grams); err != nil {
		return fmt.Errorf("add meal line: %w", err)
	}
	return nil
}

// UpdateGrams replaces one line's weight. The food row stays.
func (s *Store) UpdateGrams(id int64, grams float64) error {
	if _, err := s.db.Exec(`UPDATE meal_line SET grams = ? WHERE id = ?`, grams, id); err != nil {
		return fmt.Errorf("update grams: %w", err)
	}
	return nil
}

// UpdateFood replaces the four per-100 g numbers. Every line that
// points at this food reads the new numbers on its next load.
// The name is not a column this method writes.
func (s *Store) UpdateFood(id int64, calories, protein, carbohydrate, fat float64) error {
	const q = `
UPDATE food SET calories = ?, protein = ?, carbohydrate = ?, fat = ? WHERE id = ?`
	if _, err := s.db.Exec(q, calories, protein, carbohydrate, fat, id); err != nil {
		return fmt.Errorf("update food: %w", err)
	}
	return nil
}

// DeleteLine removes one weighed line. The food stays so another
// day, or a later add of the same name, still finds it.
func (s *Store) DeleteLine(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM meal_line WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete meal line: %w", err)
	}
	return nil
}

// Allowance reads the calorie number stored for day. The bool is
// false when that date has no row, including when the number would
// have been zero. A stored zero returns true and 0. day is
// YYYY-MM-DD, the same string a meal line uses.
func (s *Store) Allowance(day string) (float64, bool, error) {
	var calories float64
	err := s.db.QueryRow(`SELECT calories FROM meal_allowance WHERE day = ?`, day).Scan(&calories)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("read allowance %s: %w", day, err)
	}
	return calories, true, nil
}

// SetAllowance stores calories for day, replacing a number already
// there. Zero is a real allowance. The row does not require a
// weight-log day or a meal line.
func (s *Store) SetAllowance(day string, calories float64) error {
	const q = `
INSERT INTO meal_allowance (day, calories) VALUES (?, ?)
ON CONFLICT(day) DO UPDATE SET calories = excluded.calories`
	if _, err := s.db.Exec(q, day, calories); err != nil {
		return fmt.Errorf("set allowance %s: %w", day, err)
	}
	return nil
}

// ClearAllowance removes the row for day. A date with no row is
// already clear, and deleting it is not an error.
func (s *Store) ClearAllowance(day string) error {
	if _, err := s.db.Exec(`DELETE FROM meal_allowance WHERE day = ?`, day); err != nil {
		return fmt.Errorf("clear allowance %s: %w", day, err)
	}
	return nil
}
