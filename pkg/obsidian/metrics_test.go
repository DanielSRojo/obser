package obsidian

import (
	"fmt"
	"testing"
	"time"
)

func dateRange(t *testing.T, from, to string) Range {
	t.Helper()
	return Range{localDate(t, from), localDate(t, to)}
}

func TestStatsRateIgnoresDaysBeforeTrackingStart(t *testing.T) {
	j := testJournal(t, "2026-10-10", map[string]string{
		"2025-05-01": "dopamineFast:",
		"2025-06-01": "dopamineFast: true",
		"2025-06-02": "dopamineFast: true",
	})
	s := j.Stats("dopamineFast", dateRange(t, "2025-01-01", "2025-12-31"))
	if s.TrackedDays != 214 || s.DaysDone != 2 {
		t.Errorf("tracked %d, done %d; want 214 and 2", s.TrackedDays, s.DaysDone)
	}
	if s.HasTotal {
		t.Error("a boolean habit has no total")
	}
}

func TestStatsCurrentStreakTodayNotLogged(t *testing.T) {
	entries := map[string]string{
		"2025-03-04": "guitar:",
		"2025-03-10": "guitar:",
	}
	for d := 5; d <= 9; d++ {
		entries[fmt.Sprintf("2025-03-%02d", d)] = "guitar: 30 minutes"
	}
	j := testJournal(t, "2025-03-10", entries)
	s := j.Stats("guitar", dateRange(t, "2025-03-01", "2025-03-31"))
	if s.CurrentStreak != 5 {
		t.Errorf("current streak = %d, want 5", s.CurrentStreak)
	}

	// Logging today extends the streak.
	entries["2025-03-10"] = "guitar: 10 minutes"
	j = testJournal(t, "2025-03-10", entries)
	if s := j.Stats("guitar", dateRange(t, "2025-03-01", "2025-03-31")); s.CurrentStreak != 6 {
		t.Errorf("current streak = %d, want 6", s.CurrentStreak)
	}
}

func TestStatsMissingEntriesAreNotDone(t *testing.T) {
	j := testJournal(t, "2025-01-31", map[string]string{
		"2025-01-01": "guitar: 30 minutes",
		"2025-01-02": "guitar: 1 hour",
		"2025-01-04": "guitar: 30 minutes",
	})
	s := j.Stats("guitar", dateRange(t, "2025-01-01", "2025-01-10"))
	want := Stats{
		DaysDone: 3, TrackedDays: 10, Rate: 0.3,
		Total: 7200, HasTotal: true, Avg: 2400,
		LongestStreak: 2,
	}
	if s != want {
		t.Errorf("got %+v\nwant %+v", s, want)
	}
}

func TestStatsLongestStreakIsWithinRange(t *testing.T) {
	j := testJournal(t, "2025-03-31", map[string]string{
		"2025-01-30": "chain: true",
		"2025-01-31": "chain: true",
		"2025-02-01": "chain: true",
		"2025-02-02": "chain: true",
		"2025-02-10": "chain: true",
	})
	if s := j.Stats("chain", dateRange(t, "2025-02-01", "2025-02-28")); s.LongestStreak != 2 {
		t.Errorf("longest streak = %d, want 2", s.LongestStreak)
	}
}

func TestPeriods(t *testing.T) {
	got := Periods(Month, dateRange(t, "2025-11-15", "2026-02-10"))
	want := []Range{
		dateRange(t, "2025-11-15", "2025-11-30"),
		dateRange(t, "2025-12-01", "2025-12-31"),
		dateRange(t, "2026-01-01", "2026-01-31"),
		dateRange(t, "2026-02-01", "2026-02-10"),
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("months:\ngot  %v\nwant %v", got, want)
	}

	got = Periods(Year, dateRange(t, "2023-11-01", "2025-06-30"))
	want = []Range{
		dateRange(t, "2023-11-01", "2023-12-31"),
		dateRange(t, "2024-01-01", "2024-12-31"),
		dateRange(t, "2025-01-01", "2025-06-30"),
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("years:\ngot  %v\nwant %v", got, want)
	}
}

func TestWeekdayRates(t *testing.T) {
	// 2025-01-06 is a Monday. Done on both Tuesdays and one Thursday.
	j := testJournal(t, "2025-01-19", map[string]string{
		"2025-01-06": "calisthenics:",
		"2025-01-07": "calisthenics: 30 minutes",
		"2025-01-09": "calisthenics: 30 minutes",
		"2025-01-14": "calisthenics: 45 minutes",
	})
	got := j.WeekdayRates("calisthenics", dateRange(t, "2025-01-01", "2025-12-31"))
	want := [7]float64{0, 1, 0, 0.5, 0, 0, 0}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLevelsQuartiles(t *testing.T) {
	entries := map[string]string{}
	for d := 1; d <= 8; d++ {
		entries[fmt.Sprintf("2025-01-%02d", d)] = fmt.Sprintf("pushups: %d", d)
	}
	entries["2025-01-09"] = "pushups: 0"
	j := testJournal(t, "2025-01-10", entries)
	levels := j.Levels("pushups", dateRange(t, "2024-12-31", "2025-01-11"))

	// Nearest-rank quartiles of 1..8 are 2, 4 and 6.
	want := map[string]int{
		"2024-12-31": -1, // before tracking start
		"2025-01-01": 1, "2025-01-02": 1,
		"2025-01-03": 2, "2025-01-04": 2,
		"2025-01-05": 3, "2025-01-06": 3,
		"2025-01-07": 4, "2025-01-08": 4,
		"2025-01-09": 0,
		"2025-01-10": 0,  // today, no entry
		"2025-01-11": -1, // after today
	}
	for d, w := range want {
		if got := levels[localDate(t, d)]; got != w {
			t.Errorf("%s: level %d, want %d", d, got, w)
		}
	}
	if len(levels) != len(want) {
		t.Errorf("got %d days, want %d", len(levels), len(want))
	}
}

func TestLevelsBooleanHabit(t *testing.T) {
	j := testJournal(t, "2025-01-03", map[string]string{
		"2025-01-01": "chain: true",
		"2025-01-02": "chain: false",
		"2025-01-03": "chain: true",
	})
	levels := j.Levels("chain", dateRange(t, "2025-01-01", "2025-01-05"))
	want := []int{4, 0, 4, -1, -1}
	for i, w := range want {
		day := localDate(t, "2025-01-01").AddDate(0, 0, i)
		if got := levels[day]; got != w {
			t.Errorf("%s: level %d, want %d", day.Format(time.DateOnly), got, w)
		}
	}
}
