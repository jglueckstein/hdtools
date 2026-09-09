package dailylog

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStoreRoundTripAndMissingWeight(t *testing.T) {
	t.Parallel()
	s := openTestStore(t)
	ctx := context.Background()

	first, err := New(date(1989, 4, 9), ptr(145.5), 8, 1000, true, "")
	if err != nil {
		t.Fatal(err)
	}
	travel, err := New(date(1989, 4, 10), nil, 7, 0, false, "Travel")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, first); err != nil {
		t.Fatalf("Upsert first: %v", err)
	}
	if err := s.Upsert(ctx, travel); err != nil {
		t.Fatalf("Upsert travel: %v", err)
	}

	got, err := s.Get(ctx, date(1989, 4, 9))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Weight == nil || *got.Weight != 145.5 || !got.Workout || got.Steps != 1000 {
		t.Fatalf("Get first = %+v", got)
	}

	miss, err := s.Get(ctx, date(1989, 4, 10))
	if err != nil {
		t.Fatalf("Get travel: %v", err)
	}
	if miss.Weight != nil {
		t.Fatalf("travel weight = %v, want nil", *miss.Weight)
	}
	if miss.Note != "Travel" {
		t.Fatalf("travel note = %q", miss.Note)
	}

	series, err := s.Range(ctx, date(1989, 4, 9), date(1989, 4, 10))
	if err != nil {
		t.Fatalf("Range: %v", err)
	}
	trended, err := ApplyTrend(series, nil)
	if err != nil {
		t.Fatalf("ApplyTrend: %v", err)
	}
	if trended[0].Trend != 145.5 || trended[1].Trend != 145.5 {
		t.Fatalf("trends = %v, %v", trended[0].Trend, trended[1].Trend)
	}
}

func TestStoreUpsertReplacesSameDay(t *testing.T) {
	t.Parallel()
	s := openTestStore(t)
	ctx := context.Background()
	day := date(1990, 11, 4)
	a, err := New(day, ptr(171.5), 6, 1, false, "a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(day, ptr(172.0), 8, 5000, true, "b")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, a); err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, b); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != 172.0 || !got.Workout || got.Steps != 5000 || got.Note != "b" {
		t.Fatalf("replaced row = %+v", got)
	}
}

func TestStoreGetMissingDay(t *testing.T) {
	t.Parallel()
	s := openTestStore(t)
	_, err := s.Get(context.Background(), date(1990, 1, 1))
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestStoreAllOldestFirst(t *testing.T) {
	t.Parallel()
	s := openTestStore(t)
	ctx := context.Background()
	later, err := New(date(1990, 11, 2), ptr(171.5), 0, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	earlier, err := New(date(1990, 11, 1), ptr(172.5), 0, 0, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, later); err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctx, earlier); err != nil {
		t.Fatal(err)
	}
	all, err := s.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || !all[0].Day.Equal(date(1990, 11, 1)) {
		t.Fatalf("All = %+v", all)
	}
}

func TestStoreRejectsInvalidUpsert(t *testing.T) {
	t.Parallel()
	s := openTestStore(t)
	err := s.Upsert(context.Background(), DailyLog{})
	if !errors.Is(err, ErrZeroDay) {
		t.Fatalf("error = %v, want %v", err, ErrZeroDay)
	}
}

func TestMigrateAddsNoteColumn(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE daily_log (
	day TEXT PRIMARY KEY,
	weight REAL,
	sleep_hours REAL NOT NULL DEFAULT 0,
	steps INTEGER NOT NULL DEFAULT 0,
	workout INTEGER NOT NULL DEFAULT 0
);`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open old db: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	log, err := New(date(1990, 1, 1), ptr(80), 0, 0, false, "migrated")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), date(1990, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if got.Note != "migrated" {
		t.Fatalf("note = %q", got.Note)
	}
}

func TestOpenCreatesPrivateFile(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not POSIX on Windows")
	}
	path := filepath.Join(t.TempDir(), "daily.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("db mode = %o, want 0600", perm)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	// Unique file per test: :memory: is per-connection and races with t.Parallel.
	path := filepath.Join(t.TempDir(), "daily.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
