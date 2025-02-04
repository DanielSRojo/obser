package obsidian

import (
	"log"
	"sort"
	"time"
)

type Statistics struct {
	Name    string
	Content string
	Value   int
	Unit    string
}

func GetYearlyStatistics(year int) []Statistics {
	var statistics []Statistics
	for month := 1; month <= 12; month++ {
		// TODO: implement
		// GetMonthlyStatistics(year, month)
	}

	return statistics
}

func GetMonthlyStatistics(year, month int) []Statistics {
	monthCount, err := AggregateMonthlyProperties(year, month)
	if err != nil {
		log.Fatal(err)
	}

	if len(monthCount) == 0 {
		return nil
	}

	var propertyNames []string
	for i := range monthCount {
		propertyNames = append(propertyNames, i)
	}
	sort.Strings(propertyNames)

	var statistics []Statistics
	for _, name := range propertyNames {
		p := monthCount[name]
		value := int(p.Value)
		if p.Unit == "seconds" && value >= 200 {
			value = value / 60
			p.Unit = "minutes"
		}
		if p.Unit == "minutes" && value >= 100 {
			value = value / 60
			p.Unit = "hours"
		}
		s := &Statistics{
			Name:    p.Name,
			Content: p.Content,
			Value:   p.Value,
			Unit:    p.Unit,
		}
		statistics = append(statistics, *s)
	}

	return statistics
}

func GetStatistics(year, month int) ([]Statistics, error) {
	now := time.Now()
	if year > 0 && month == 0 {
		return GetYearlyStatistics(year), nil
	}
	if year == 0 && month > 0 {
		year = now.Year()
	}
	if year == 0 && month == 0 {
		year = now.Year()
		month = int(now.Month())
	}
	return GetMonthlyStatistics(year, month), nil
}
