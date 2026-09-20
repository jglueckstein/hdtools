package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jglueckstein/hdtools/internal/config"
	"github.com/jglueckstein/hdtools/internal/dailylog"
	"github.com/jglueckstein/hdtools/internal/units"
)

func TestBuildMonthSheetFillsBlankDaysAndCarriesTrend(t *testing.T) {
	t.Parallel()
	carry := 173.6
	logs := []dailylog.DailyLog{
		{Day: dateUTC(1990, 11, 1), Weight: ptrFloat(172.5)},
		{Day: dateUTC(1990, 11, 2), Weight: ptrFloat(171.5)},
		{Day: dateUTC(1990, 11, 4), Weight: ptrFloat(171.5)},
	}
	trended, err := dailylog.ApplyTrend(logs, &carry)
	if err != nil {
		t.Fatal(err)
	}
	sheet := buildMonthSheet(trended, 1990, time.November)
	if len(sheet) != 30 {
		t.Fatalf("len = %d, want 30", len(sheet))
	}
	if !sheet[0].HasEntry || sheet[2].HasEntry {
		t.Fatalf("day 1 has entry=%v day 3 has entry=%v", sheet[0].HasEntry, sheet[2].HasEntry)
	}
	if !sheet[2].HasTrend || sheet[2].Trend != trended[1].Trend {
		t.Fatalf("day 3 trend = %v want carry %v", sheet[2].Trend, trended[1].Trend)
	}
}

func TestDaysInMonthFebruaryLeap(t *testing.T) {
	t.Parallel()
	if daysInMonth(1992, time.February) != 29 {
		t.Fatal("1992 should be a leap February")
	}
	if daysInMonth(1990, time.February) != 28 {
		t.Fatal("1990 February has 28 days")
	}
}

func TestMonthNavigationClampsDay(t *testing.T) {
	t.Parallel()
	m := newMonth(dateUTC(1990, 3, 31))
	m.prevMonth()
	if m.month != time.February || m.day != 28 {
		t.Fatalf("got %s %d", m.month, m.day)
	}
}

func TestPatchCellConvertsPounds(t *testing.T) {
	t.Parallel()
	log := dailylog.DailyLog{Day: dateUTC(1990, 11, 4)}
	got, err := patchCell(log, colWeight, "171.5", units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	want, err := units.ToKG(171.5, units.Pound)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight == nil || *got.Weight != want {
		t.Fatalf("weight = %v want %v", got.Weight, want)
	}
}

func TestPatchCellClearsWeight(t *testing.T) {
	t.Parallel()
	w := 80.0
	log := dailylog.DailyLog{Day: dateUTC(1990, 11, 4), Weight: &w}
	got, err := patchCell(log, colWeight, "", units.Kilogram)
	if err != nil {
		t.Fatal(err)
	}
	if got.Weight != nil {
		t.Fatalf("weight = %v, want nil", *got.Weight)
	}
}

func TestNextCellWalksColumnsThenNextDay(t *testing.T) {
	t.Parallel()
	m := newMonth(dateUTC(1990, 11, 4))
	m.col = colWeight
	m.nextCell()
	if m.day != 4 || m.col != colSleep {
		t.Fatalf("after weight: day %d col %d", m.day, m.col)
	}
	m.col = colNote
	m.nextCell()
	if m.day != 5 || m.col != colWeight {
		t.Fatalf("after note: day %d col %d", m.day, m.col)
	}
}

func TestNextCellStaysOnLastCellOfMonth(t *testing.T) {
	t.Parallel()
	m := newMonth(dateUTC(1990, 11, 30))
	m.col = colNote
	m.nextCell()
	if m.day != 30 || m.col != colNote {
		t.Fatalf("got day %d col %d", m.day, m.col)
	}
}

func TestGotoTodayMonthJumpsToToday(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, dateUTC(1990, 6, 1))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m")
	app.month.col = colSleep
	press(app, "t")
	if app.month.year != 1990 || app.month.month != time.November {
		t.Fatalf("sheet = %s %d, want November 1990", app.month.month, app.month.year)
	}
	if app.month.day != 10 {
		t.Fatalf("day = %d, want 10", app.month.day)
	}
	if app.month.col != colWeight {
		t.Fatalf("col = %d, want weight", app.month.col)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("rows = %d, want 1", len(logs))
	}
}

func TestGotoTodayMonthDoesNotCreateToday(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, dateUTC(1990, 11, 8))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m")
	app.month.day = 8
	app.month.col = colNote
	press(app, "t")
	if app.month.day != 10 {
		t.Fatalf("day = %d, want 10", app.month.day)
	}
	if app.month.col != colWeight {
		t.Fatalf("col = %d, want weight", app.month.col)
	}
	if _, err := store.Get(context.Background(), dateUTC(1990, 11, 10)); err == nil {
		t.Fatal("created 10 November 1990")
	}
}

func TestGotoTodayMonthEmptyStillGoesToToday(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m")
	press(app, "[")
	if app.month.month != time.October {
		t.Fatalf("after [: %s, want October", app.month.month)
	}
	press(app, "t")
	if app.month.year != 1990 || app.month.month != time.November {
		t.Fatalf("sheet = %s %d, want November 1990", app.month.month, app.month.year)
	}
	if app.month.day != 10 {
		t.Fatalf("day = %d, want 10", app.month.day)
	}
	if app.month.col != colWeight {
		t.Fatalf("col = %d, want weight", app.month.col)
	}
	logs, err := store.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 0 {
		t.Fatalf("rows = %d, want 0", len(logs))
	}
}

func TestGotoTodayMonthEditingIsText(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	seedDays(t, store, dateUTC(1990, 11, 8))
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	press(app, "m")
	app.month.col = colNote
	app.month.beginEdit("")
	press(app, "t")
	if !strings.Contains(app.month.input.Value(), "t") {
		t.Fatalf("input = %q, want t", app.month.input.Value())
	}
	if app.month.day != 8 {
		t.Fatalf("day = %d, want 8", app.month.day)
	}
}

func TestGotoTodayMonthDoesNotMoveList(t *testing.T) {
	freezeToday(t, time.Date(1990, 11, 10, 12, 0, 0, 0, time.UTC))
	store := openStore(t)
	june := dateUTC(1990, 6, 1)
	seedDays(t, store,
		june,
		dateUTC(1990, 11, 10),
	)
	app := New(store, "mem.db", config.Default())
	app.Update(app.load())
	selectDay(t, app, june)
	press(app, "m")
	if app.month.month != time.June {
		t.Fatalf("month = %s, want June", app.month.month)
	}
	press(app, "t")
	press(app, "esc")
	if app.month.editing {
		press(app, "esc")
	}
	if app.screen != screenList {
		press(app, "esc")
	}
	if app.screen != screenList {
		t.Fatalf("screen = %v, want list", app.screen)
	}
	assertSelectedDay(t, app, june)
}

func dateUTC(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func ptrFloat(w float64) *float64 {
	return &w
}
