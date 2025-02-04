package cmd

import (
	"log"
	"sort"

	"github.com/danielsrojo/obser/pkg/obsidian"
)

func GetYearlyStatistics(year int) []obsidian.Property {
	var properties []obsidian.Property
	for month := 1; month <= 12; month++ {
		// TODO: implement
		// GetMonthlyStatistics(year, month)
	}

	return properties
}

func GetMonthlyStatistics(year, month int) []obsidian.Property {
	monthCount, err := obsidian.AggregateMonthlyProperties(year, month)
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

	// fmt.Printf("%s\n---\n", month)
	var properties []obsidian.Property
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
	}

	return properties
}

func formatStatistics() {
	// fmt.Printf("%d\n\n", year)
}
