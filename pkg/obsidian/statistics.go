package obsidian

import (
	"time"
)

// GetYearlyStatistics returns each habit's total for the year: its amount in
// the habit's unit, or the days done for habits without amounts.
func GetYearlyStatistics(year int) ([]Property, error) {
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	return getStatistics(Range{from, from.AddDate(1, 0, -1)})
}

// GetMonthlyStatistics is GetYearlyStatistics for a single month.
func GetMonthlyStatistics(year, month int) ([]Property, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	return getStatistics(Range{from, from.AddDate(0, 1, -1)})
}

func getStatistics(r Range) ([]Property, error) {
	j, err := LoadJournal(time.Now())
	if err != nil {
		return nil, err
	}

	var statistics []Property
	for _, name := range j.HabitNames() {
		s := j.Stats(name, r)
		value := s.DaysDone
		if s.HasTotal {
			value = s.Total
		}
		if value == 0 {
			continue
		}
		statistics = append(statistics, Property{
			Name:  name,
			Value: value,
			Unit:  j.Habits[name].Unit,
			Type:  Numeric,
		})
	}
	return statistics, nil
}

func GetStatistics(year, month int) ([]Property, error) {
	if month != 0 {
		return GetMonthlyStatistics(year, month)
	}
	return GetYearlyStatistics(year)
}
