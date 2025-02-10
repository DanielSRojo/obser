package obsidian

import (
	"fmt"
	"log"
	"sort"
)

func GetYearlyStatistics(year int) []Property {
	// yearlyStatistics := GetMonthlyStatistics(year, month)
	// for month := 1; month <= 12; month++ {
	// 	// TODO: implement
	// 	for _, p := range yearlyStatistics {
	// 		yearlyStatistics = SumProperties()
	// 	}
	// }
	// return statistics

	// var result []Property
	for month := 1; month <= 12; month++ {
		// m = GetMonthlyStatistics(year, month)
		// result = SumProperties(result, GetMonthlyStatistics(year, m))
	}

	return []Property{}
}

func SumProperties(a, b Property) (*Property, error) {
	if (a.Type != Numeric) || (b.Type != Numeric) {
		return nil, fmt.Errorf("error: non numeric properties")
	}

	if a.Name != b.Name {
		return nil, fmt.Errorf("error: properties are not the same type")
	}

	// TODO: add unit detection and conversion when required
	if a.Unit != b.Unit {
		return nil, fmt.Errorf("error: properties have not the same unit")
	}

	return &Property{
		Name:  a.Name,
		Value: a.Value + b.Value,
		Unit:  a.Unit,
		Type:  a.Type,
	}, nil
}

func GetMonthlyStatistics(year, month int) []Property {
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

	var statistics []Property
	for _, name := range propertyNames {
		p := monthCount[name]
		value := int(p.Value)
		unit := p.Unit
		if unit == "seconds" && value >= 200 {
			value = value / 60
			unit = "minutes"
		}
		if unit == "minutes" && value >= 100 {
			value = value / 60
			unit = "hours"
		}
		p.Value = value
		p.Unit = unit
		statistics = append(statistics, p)
	}

	return statistics
}

func GetStatistics(year, month int) ([]Property, error) {
	if month != 0 {
		return GetMonthlyStatistics(year, month), nil
	}
	return GetYearlyStatistics(year), nil
}
