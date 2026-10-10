package render

import (
	"testing"
	"time"
)

func TestHeatmapPlain(t *testing.T) {
	// One level per month, blank after 2025-10-10 as if that were today.
	today := time.Date(2025, 10, 10, 0, 0, 0, 0, time.Local)
	levels := func(day time.Time) int {
		if day.After(today) {
			return -1
		}
		return int(day.Month()) % 5
	}
	got := Heatmap(2025, levels, Style{})
	want := `    Jan Feb Mar  Apr May Jun  Jul Aug  Sep Oct Nov  Dec
Mon  ░░░░▒▒▒▒▓▓▓▓▓████····░░░░░▒▒▒▒▓▓▓▓█████·
     ░░░░▒▒▒▒▓▓▓▓█████····░░░░▒▒▒▒▒▓▓▓▓█████·
Wed ░░░░░▒▒▒▒▓▓▓▓█████····░░░░▒▒▒▒▒▓▓▓▓████··
    ░░░░░▒▒▒▒▓▓▓▓████·····░░░░▒▒▒▒▒▓▓▓▓████··
Fri ░░░░░▒▒▒▒▓▓▓▓████·····░░░░▒▒▒▒▓▓▓▓▓████··
    ░░░░▒▒▒▒▓▓▓▓▓████·····░░░░▒▒▒▒▓▓▓▓▓████·
    ░░░░▒▒▒▒▓▓▓▓▓████····░░░░░▒▒▒▒▓▓▓▓▓████·
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
