package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Duration formats seconds in the largest fitting units, using at most two
// parts, rounded to the smallest part shown: 42h 30m, 12m 5s, 45s.
func Duration(seconds int) string {
	if seconds >= 3600 {
		minutes := (seconds + 30) / 60
		h, m := minutes/60, minutes%60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if seconds >= 60 {
		m, s := seconds/60, seconds%60
		if s == 0 {
			return fmt.Sprintf("%dm", m)
		}
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", seconds)
}

// Amount formats an amount in a habit's unit: a duration for seconds, the
// number alone for unitless counts, and "12340 words" otherwise.
func Amount(n int, unit string) string {
	switch unit {
	case "seconds":
		return Duration(n)
	case "":
		return strconv.Itoa(n)
	}
	return fmt.Sprintf("%d %s", n, unit)
}

// AmountF formats an average, rounded to the nearest whole unit.
func AmountF(x float64, unit string) string {
	return Amount(int(math.Round(x)), unit)
}

var sparks = []rune("▁▂▃▄▅▆▇█")

// Sparkline draws one character per value, each between 0 and 1. NaN values
// are drawn as a space.
func Sparkline(values []float64) string {
	var b strings.Builder
	for _, v := range values {
		if math.IsNaN(v) {
			b.WriteRune(' ')
			continue
		}
		i := int(math.Round(math.Max(0, math.Min(1, v)) * float64(len(sparks)-1)))
		b.WriteRune(sparks[i])
	}
	return b.String()
}

var eighths = []rune(" ▏▎▍▌▋▊▉")

// Bar draws a horizontal bar of the given fraction (0–1) of width cells,
// padded with spaces to width.
func Bar(fraction float64, width int) string {
	fraction = math.Max(0, math.Min(1, fraction))
	n := int(math.Round(fraction * float64(width) * 8))
	full, part := n/8, n%8
	bar := strings.Repeat("█", full)
	if part > 0 {
		bar += string(eighths[part])
		full++
	}
	return bar + strings.Repeat(" ", width-full)
}

// Table lays out rows in columns padded to their visible width, so multibyte
// characters and ANSI styles don't break the alignment.
type Table struct {
	Header []string
	Rows   [][]string
	Right  map[int]bool // right-aligned columns
}

func (t Table) String() string {
	rows := append([][]string{t.Header}, t.Rows...)
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if i == len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
	}

	var b strings.Builder
	for _, row := range rows {
		if row == nil {
			continue
		}
		var line strings.Builder
		for i, cell := range row {
			if i > 0 {
				line.WriteString("  ")
			}
			pad := strings.Repeat(" ", widths[i]-lipgloss.Width(cell))
			if t.Right[i] {
				line.WriteString(pad + cell)
			} else {
				line.WriteString(cell + pad)
			}
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}
