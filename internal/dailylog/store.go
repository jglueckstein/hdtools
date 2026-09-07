package dailylog

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store persists daily logs in SQLite. Trend is not a column: readers call
// ApplyTrend on a loaded series so a backdated weight edit cannot desync the
// moving average from the scale readings.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at path and ensures the daily_log
// table exists. path may be ":memory:" for tests.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open daily log store: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close daily log store: %w", err)
	}
	return nil
}

func (s *Store) migrate() error {
	const q = `
CREATE TABLE IF NOT EXISTS daily_log (
	day TEXT PRIMARY KEY,
	weight REAL,
	sleep_hours REAL NOT NULL DEFAULT 0,
	steps INTEGER NOT NULL DEFAULT 0,
	workout INTEGER NOT NULL DEFAULT 0,
	note TEXT NOT NULL DEFAULT ''
);`
	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("migrate daily log store: %w", err)
	}
	if err := s.ensureNoteColumn(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureNoteColumn() error {
	rows, err := s.db.Query(`PRAGMA table_info(daily_log)`)
	if err != nil {
		return fmt.Errorf("inspect daily log columns: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return fmt.Errorf("inspect daily log columns: %w", err)
		}
		if name == "note" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("inspect daily log columns: %w", err)
	}
	if _, err := s.db.Exec(`ALTER TABLE daily_log ADD COLUMN note TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("add daily log note column: %w", err)
	}
	return nil
}

// Upsert inserts or replaces the log for d.Day. Trend is ignored on write.
func (s *Store) Upsert(ctx context.Context, d DailyLog) error {
	if err := d.Validate(); err != nil {
		return fmt.Errorf("upsert daily log: %w", err)
	}
	d.Day = calendarDay(d.Day)
	workout := 0
	if d.Workout {
		workout = 1
	}
	var weight any
	if d.Weight != nil {
		weight = *d.Weight
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO daily_log (day, weight, sleep_hours, steps, workout, note)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(day) DO UPDATE SET
	weight = excluded.weight,
	sleep_hours = excluded.sleep_hours,
	steps = excluded.steps,
	workout = excluded.workout,
	note = excluded.note;
`, d.Day.Format(time.DateOnly), weight, d.SleepHours, d.Steps, workout, d.Note)
	if err != nil {
		return fmt.Errorf("upsert daily log %s: %w", d.Day.Format(time.DateOnly), err)
	}
	return nil
}

// Get loads the log for day. It does not compute Trend.
func (s *Store) Get(ctx context.Context, day time.Time) (DailyLog, error) {
	day = calendarDay(day)
	row := s.db.QueryRowContext(ctx, `
SELECT day, weight, sleep_hours, steps, workout, note
FROM daily_log WHERE day = ?`, day.Format(time.DateOnly))
	d, err := scanLog(row)
	if err != nil {
		return DailyLog{}, fmt.Errorf("get daily log %s: %w", day.Format(time.DateOnly), err)
	}
	return d, nil
}

// Range returns logs from fromDay through toDay inclusive, oldest first.
func (s *Store) Range(ctx context.Context, fromDay, toDay time.Time) ([]DailyLog, error) {
	fromDay = calendarDay(fromDay)
	toDay = calendarDay(toDay)
	rows, err := s.db.QueryContext(ctx, `
SELECT day, weight, sleep_hours, steps, workout, note
FROM daily_log
WHERE day >= ? AND day <= ?
ORDER BY day ASC`, fromDay.Format(time.DateOnly), toDay.Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("range daily logs: %w", err)
	}
	defer rows.Close()

	var out []DailyLog
	for rows.Next() {
		d, err := scanLog(rows)
		if err != nil {
			return nil, fmt.Errorf("range daily logs: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("range daily logs: %w", err)
	}
	return out, nil
}

// All returns every stored log, oldest first. Trend is not computed.
func (s *Store) All(ctx context.Context) ([]DailyLog, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT day, weight, sleep_hours, steps, workout, note
FROM daily_log
ORDER BY day ASC`)
	if err != nil {
		return nil, fmt.Errorf("list daily logs: %w", err)
	}
	defer rows.Close()

	var out []DailyLog
	for rows.Next() {
		d, err := scanLog(rows)
		if err != nil {
			return nil, fmt.Errorf("list daily logs: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list daily logs: %w", err)
	}
	return out, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLog(row scanner) (DailyLog, error) {
	var (
		dayStr     string
		weight     sql.NullFloat64
		sleepHours float64
		steps      int
		workout    int
		note       string
	)
	if err := row.Scan(&dayStr, &weight, &sleepHours, &steps, &workout, &note); err != nil {
		return DailyLog{}, err
	}
	day, err := time.Parse(time.DateOnly, dayStr)
	if err != nil {
		return DailyLog{}, fmt.Errorf("parse day %q: %w", dayStr, err)
	}
	d := DailyLog{
		Day:        calendarDay(day),
		SleepHours: sleepHours,
		Steps:      steps,
		Workout:    workout != 0,
		Note:       note,
	}
	if weight.Valid {
		w := weight.Float64
		d.Weight = &w
	}
	return d, nil
}
