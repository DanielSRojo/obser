package render

import (
	"strings"
	"time"
)

// Heatmap draws a calendar of the year with one row per weekday, Monday
// first, and one column per week, with month names above the week each month
// starts in. levels returns each day's level, 0–4, or -1 for a blank day.
func Heatmap(year int, levels func(time.Time) int, style Style) string {
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	next := jan1.AddDate(1, 0, 0)
	offset := (int(jan1.Weekday()) + 6) % 7 // Monday = 0
	column := func(day time.Time) int { return (day.YearDay() - 1 + offset) / 7 }
	weeks := column(next.AddDate(0, 0, -1)) + 1

	labels := []rune(strings.Repeat(" ", weeks+3))
	free := 0
	for m := time.January; m <= time.December; m++ {
		col := column(time.Date(year, m, 1, 0, 0, 0, 0, time.Local))
		if col < free {
			continue
		}
		copy(labels[col:], []rune(m.String()[:3]))
		free = col + 4
	}

	var cells [7][]string
	for wd := range cells {
		cells[wd] = make([]string, weeks)
		for c := range cells[wd] {
			cells[wd][c] = " "
		}
	}
	for day := jan1; day.Before(next); day = day.AddDate(0, 0, 1) {
		cells[(int(day.Weekday())+6)%7][column(day)] = style.Level(levels(day))
	}

	var b strings.Builder
	b.WriteString(strings.TrimRight("    "+string(labels), " ") + "\n")
	for wd, name := range []string{"Mon", "", "Wed", "", "Fri", "", ""} {
		line := (name + "    ")[:4] + strings.Join(cells[wd], "")
		b.WriteString(strings.TrimRight(line, " ") + "\n")
	}
	return b.String()
}
