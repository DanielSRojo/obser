package obsidian

import (
	"fmt"
	"log"
	"sort"
)

func GetYearlyStatistics(year int) []Property {
	totalProperties := make(map[string]Property)
	for month := 0; month <= 11; month++ {
		statistics := GetMonthlyStatistics(year, month)
		for _, p := range statistics {
			if totalProperties[p.Name].Value == 0 {
				totalProperties[p.Name] = p
				continue
			}
			psum, err := SumProperties(totalProperties[p.Name], p)
			if err != nil {
				log.Printf("error summing properties: %s\n", err)
				// fmt.Printf("%v\n", totalProperties[p.Name])
				continue
			}
			totalProperties[p.Name] = *psum
		}
	}

	var keys []string
	for key := range totalProperties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	totalStatistics := make([]Property, 0, len(totalProperties))
	for _, key := range keys {
		totalStatistics = append(totalStatistics, totalProperties[key])
	}

	return totalStatistics
}

func SumProperties(a, b Property) (*Property, error) {
	if (a.Type != Numeric) || (b.Type != Numeric) {
		return nil, fmt.Errorf("error: non numeric properties")
	}

	if a.Name != b.Name {
		return nil, fmt.Errorf("error: properties are not the same type")
	}

	if a.Unit != b.Unit && a.Value != 0 && b.Value != 0 {
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
	monthTotals, err := AggregateMonthlyProperties(year, month)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	if len(monthTotals) == 0 {
		return nil
	}

	var propertyNames []string
	for i := range monthTotals {
		propertyNames = append(propertyNames, i)
	}
	sort.Strings(propertyNames)

	var statistics []Property
	for _, name := range propertyNames {
		p := monthTotals[name]
		p.Value = int(p.Value)
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
