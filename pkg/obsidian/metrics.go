package obsidian

import (
	"math"
	"sort"
	"time"
)

// Range is an inclusive range of days.
type Range struct{ From, To time.Time }

// Stats are a habit's metrics over a range.
type Stats struct {
	DaysDone, TrackedDays int
	Rate                  float64
	Total                 int
	HasTotal              bool
	Avg                   float64 // per done day with an amount
	CurrentStreak         int
	LongestStreak         int
}

// Period is the length of the periods a range is split into.
type Period int

const (
	Month Period = iota
	Year
)

// tracked clips the range to the days the habit is tracked: from its tracking
// start through today. ok is false when no day is left.
func (j *Journal) tracked(habit string, r Range) (from, to time.Time, ok bool) {
	h, exists := j.Habits[habit]
	if !exists {
		return from, to, false
	}
	from, to = Midnight(r.From), Midnight(r.To)
	if from.Before(h.Start) {
		from = h.Start
	}
	if to.After(j.Today) {
		to = j.Today
	}
	return from, to, !from.After(to)
}

// Stats computes the habit's metrics over the range. Days without an entry
// count as tracked but not done. The current streak ignores the range.
func (j *Journal) Stats(habit string, r Range) Stats {
	var s Stats
	s.HasTotal = j.Habits[habit].HasAmounts
	s.CurrentStreak = j.currentStreak(habit)

	from, to, ok := j.tracked(habit, r)
	if !ok {
		return s
	}
	var withAmount, streak int
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		s.TrackedDays++
		v := j.Value(habit, day)
		if !v.Done {
			streak = 0
			continue
		}
		s.DaysDone++
		streak++
		s.LongestStreak = max(s.LongestStreak, streak)
		if v.HasAmount {
			s.Total += v.Amount
			withAmount++
		}
	}
	s.Rate = float64(s.DaysDone) / float64(s.TrackedDays)
	if withAmount > 0 {
		s.Avg = float64(s.Total) / float64(withAmount)
	}
	return s
}

// currentStreak counts the consecutive done days ending today, or ending
// yesterday when today is not done (yet).
func (j *Journal) currentStreak(habit string) int {
	day := j.Today
	if !j.Value(habit, day).Done {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for ; j.Value(habit, day).Done; day = day.AddDate(0, 0, -1) {
		n++
	}
	return n
}

// Periods splits the range into calendar months or years. The first and last
// periods are clipped to the range.
func Periods(by Period, r Range) []Range {
	from, to := Midnight(r.From), Midnight(r.To)
	var periods []Range
	for start := from; !start.After(to); {
		var next time.Time
		if by == Year {
			next = time.Date(start.Year()+1, 1, 1, 0, 0, 0, 0, time.Local)
		} else {
			next = time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.Local)
		}
		end := next.AddDate(0, 0, -1)
		if end.After(to) {
			end = to
		}
		periods = append(periods, Range{start, end})
		start = next
	}
	return periods
}

// WeekdayRates returns the habit's completion rate on each weekday within
// the range, Monday first.
func (j *Journal) WeekdayRates(habit string, r Range) [7]float64 {
	var rates [7]float64
	from, to, ok := j.tracked(habit, r)
	if !ok {
		return rates
	}
	var done, tracked [7]int
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		wd := MondayIndex(day)
		tracked[wd]++
		if j.Value(habit, day).Done {
			done[wd]++
		}
	}
	for i := range rates {
		if tracked[i] > 0 {
			rates[i] = float64(done[i]) / float64(tracked[i])
		}
	}
	return rates
}

// MondayIndex returns the day's weekday counting from Monday = 0.
func MondayIndex(t time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

// Levels returns a heatmap level for every day of the range: -1 for days
// before the tracking start or after today, 0 when not done, and 1–4 for done
// days by quartile of the habit's amounts on done days in the range. Done
// days of a habit without amounts, or whose amounts don't vary, are level 4.
// A done day without an amount in a habit that has them is level 1.
func (j *Journal) Levels(habit string, r Range) map[time.Time]int {
	levels := map[time.Time]int{}
	from, to := Midnight(r.From), Midnight(r.To)
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		levels[day] = -1
	}
	tFrom, tTo, ok := j.tracked(habit, r)
	if !ok {
		return levels
	}

	var values []int
	for day := tFrom; !day.After(tTo); day = day.AddDate(0, 0, 1) {
		if v := j.Value(habit, day); v.Done && v.HasAmount {
			values = append(values, v.Amount)
		}
	}
	sort.Ints(values)
	quartile := func(p float64) int {
		return values[int(math.Ceil(p*float64(len(values))))-1]
	}
	vary := len(values) > 0 && values[0] != values[len(values)-1]

	for day := tFrom; !day.After(tTo); day = day.AddDate(0, 0, 1) {
		v := j.Value(habit, day)
		switch {
		case !v.Done:
			levels[day] = 0
		case !vary:
			if v.HasAmount || !j.Habits[habit].HasAmounts {
				levels[day] = 4
			} else {
				levels[day] = 1
			}
		case !v.HasAmount:
			levels[day] = 1
		case v.Amount <= quartile(0.25):
			levels[day] = 1
		case v.Amount <= quartile(0.5):
			levels[day] = 2
		case v.Amount <= quartile(0.75):
			levels[day] = 3
		default:
			levels[day] = 4
		}
	}
	return levels
}
