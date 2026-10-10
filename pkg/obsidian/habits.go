package obsidian

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Value is a habit's value on one day, normalized to the habit's unit.
type Value struct {
	Done      bool
	Amount    int    // seconds for durations, else the raw count
	Unit      string // "seconds", "words", "" for unitless counts
	HasAmount bool
}

// Day is one journal entry.
type Day struct {
	Date   time.Time        // local midnight
	Values map[string]Value // habit -> value, present when the entry has the key
}

// Habit is a property that is done on at least one day.
type Habit struct {
	Name       string
	Unit       string    // resolved unit, "" for counts and booleans
	Start      time.Time // tracking start: the first done day
	HasAmounts bool
}

// Journal holds every journal entry up to Today, with normalized habit values.
type Journal struct {
	Today    time.Time
	Days     []Day // sorted by date
	Habits   map[string]Habit
	Warnings []string

	index map[time.Time]*Day
}

var journalFileName = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)

// metadataProperties are never habits.
var metadataProperties = map[string]bool{"type": true, "aliases": true, "tags": true}

// LoadJournal reads every journal entry dated on or before today from the
// journal folder. Any YYYY-MM-DD.md note directly in the folder is an entry,
// with or without type: journal.
func LoadJournal(today time.Time) (*Journal, error) {
	today = Midnight(today)
	dir, err := JournalDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var days []Day
	var properties [][]Property
	for _, f := range files {
		if f.IsDir() || !journalFileName.MatchString(f.Name()) {
			continue
		}
		date, err := time.ParseInLocation(journalDateLayout, strings.TrimSuffix(f.Name(), ".md"), time.Local)
		if err != nil || date.After(today) {
			continue
		}
		note := &Note{Directory: dir, Title: f.Name()}
		if err := note.LoadProperties(); err != nil {
			return nil, err
		}
		days = append(days, Day{Date: date})
		properties = append(properties, note.Properties)
	}

	return newJournal(today, days, properties), nil
}

func isJournal(properties []Property) bool {
	for _, p := range properties {
		if strings.ToLower(p.Name) == "type" && strings.ToLower(p.Content) == "journal" {
			return true
		}
	}
	return false
}

// newJournal builds a journal from the days, sorted by date, and the parsed
// properties of each day.
func newJournal(today time.Time, days []Day, properties [][]Property) *Journal {
	j := &Journal{Today: Midnight(today), Habits: map[string]Habit{}}

	isHabit := map[string]bool{}
	for i := range days {
		days[i].Values = map[string]Value{}
		for _, p := range properties[i] {
			if p.Name == "" || metadataProperties[strings.ToLower(p.Name)] {
				continue
			}
			v := parseValue(p)
			days[i].Values[p.Name] = v
			if v.Done && p.Type != Text {
				isHabit[p.Name] = true
			}
		}
	}
	for i := range days {
		for name := range days[i].Values {
			if !isHabit[name] {
				delete(days[i].Values, name)
			}
		}
	}

	sort.Slice(days, func(a, b int) bool { return days[a].Date.Before(days[b].Date) })
	j.Days = days
	j.index = make(map[time.Time]*Day, len(days))
	for i := range j.Days {
		j.index[j.Days[i].Date] = &j.Days[i]
	}

	names := make([]string, 0, len(isHabit))
	for name := range isHabit {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		j.Habits[name] = j.resolveHabit(name)
	}
	return j
}

// parseValue applies the done-day rule to a parsed property. Unit resolution
// happens later, once all the habit's values are known.
func parseValue(p Property) Value {
	switch {
	case p.Content == "":
		return Value{}
	case p.Type == Boolean:
		return Value{Done: p.Value == 1}
	case p.Type == Numeric:
		if p.Value <= 0 {
			return Value{}
		}
		return Value{Done: true, Amount: p.Value, Unit: p.Unit, HasAmount: true}
	default:
		return Value{Done: true}
	}
}

// resolveHabit finds the habit's tracking start and unit, and normalizes its
// values to that unit: bare numbers take the unit when it is the only one
// (counting as minutes when it is a duration), and values in other units lose
// their amount.
func (j *Journal) resolveHabit(name string) Habit {
	h := Habit{Name: name}
	units := map[string]int{}
	for _, d := range j.Days {
		v := d.Values[name]
		if v.Done && h.Start.IsZero() {
			h.Start = d.Date
		}
		if v.HasAmount && v.Unit != "" {
			units[v.Unit]++
		}
	}

	for unit, n := range units {
		if h.Unit == "" || n > units[h.Unit] || (n == units[h.Unit] && unit < h.Unit) {
			h.Unit = unit
		}
	}
	if len(units) > 1 {
		var found []string
		for unit := range units {
			found = append(found, unit)
		}
		sort.Strings(found)
		j.Warnings = append(j.Warnings, fmt.Sprintf(
			"habit %q has values in several units (%s): only %s are counted in totals",
			name, strings.Join(found, ", "), h.Unit))
	}

	for i := range j.Days {
		v, ok := j.Days[i].Values[name]
		if !ok || !v.HasAmount {
			continue
		}
		switch {
		case v.Unit == h.Unit:
		case v.Unit == "" && len(units) == 1:
			v.Unit = h.Unit
			if h.Unit == "seconds" {
				v.Amount *= 60
			}
		default:
			v.Amount, v.Unit, v.HasAmount = 0, "", false
		}
		h.HasAmounts = h.HasAmounts || v.HasAmount
		j.Days[i].Values[name] = v
	}
	return h
}

// Value returns the habit's value on the given day, which is the zero Value
// when there is no entry or the entry doesn't have the habit.
func (j *Journal) Value(habit string, day time.Time) Value {
	d := j.index[Midnight(day)]
	if d == nil {
		return Value{}
	}
	return d.Values[habit]
}

// HabitNames returns the habit names, sorted.
func (j *Journal) HabitNames() []string {
	names := make([]string, 0, len(j.Habits))
	for name := range j.Habits {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Midnight returns the start of t's day in local time.
func Midnight(t time.Time) time.Time {
	t = t.In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
