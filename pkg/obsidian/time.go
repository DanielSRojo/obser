package obsidian

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func SumMinutes(timeStrings []string) (int, error) {
	var totalMinutes int

	for _, timeStr := range timeStrings {
		parts := strings.Split(timeStr, " ")
		if len(parts) < 2 {
			return 0, fmt.Errorf("invalid format: %s", timeStr)
		}

		_, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("failed to parse minutes in %s: %w", timeStr, err)
		}
	}

	return totalMinutes, nil
}

func ConvertDate(date string) (string, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		return "", err
	}

	convertedDate := t.Format("Mon, 02 Jan")

	return convertedDate, nil
}
