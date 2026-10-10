package render

import (
	"math"
	"testing"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{153000, "42h 30m"},
		{725, "12m 5s"},
		{0, "0s"},
		{45, "45s"},
		{60, "1m"},
		{3600, "1h"},
		{3629, "1h"},    // rounds to the minute
		{3630, "1h 1m"}, // rounds to the minute
		{7199, "2h"},    // 119m 59s rounds up to 2h
		{3599, "59m 59s"},
	}
	for _, tt := range tests {
		if got := Duration(tt.seconds); got != tt.want {
			t.Errorf("Duration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestAmount(t *testing.T) {
	tests := []struct {
		n    int
		unit string
		want string
	}{
		{12340, "words", "12340 words"},
		{153000, "seconds", "42h 30m"},
		{7, "", "7"},
	}
	for _, tt := range tests {
		if got := Amount(tt.n, tt.unit); got != tt.want {
			t.Errorf("Amount(%d, %q) = %q, want %q", tt.n, tt.unit, got, tt.want)
		}
	}
	if got := AmountF(1834.6, "seconds"); got != "30m 35s" {
		t.Errorf("AmountF = %q, want 30m 35s", got)
	}
}

func TestSparkline(t *testing.T) {
	values := []float64{math.NaN(), 0, 1.0 / 7, 2.0 / 7, 3.0 / 7, 4.0 / 7, 5.0 / 7, 6.0 / 7, 1, 1.5, math.NaN()}
	if got, want := Sparkline(values), " ▁▂▃▄▅▆▇██ "; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBar(t *testing.T) {
	tests := []struct {
		fraction float64
		width    int
		want     string
	}{
		{0, 4, "    "},
		{1, 4, "████"},
		{0.5, 4, "██  "},
		{0.5625, 4, "██▎ "},  // 2.25 cells
		{0.96875, 4, "███▉"}, // 3.875 cells
		{2, 3, "███"},
	}
	for _, tt := range tests {
		if got := Bar(tt.fraction, tt.width); got != tt.want {
			t.Errorf("Bar(%v, %d) = %q, want %q", tt.fraction, tt.width, got, tt.want)
		}
	}
}

func TestTable(t *testing.T) {
	green := "\x1b[32m█▇▆\x1b[0m"
	table := Table{
		Header: []string{"HABIT", "DAYS", "TREND"},
		Rows: [][]string{
			{"guitar", "180", "▁▂▃"},
			{"café", "7", green},
		},
		Right: map[int]bool{1: true},
	}
	want := "HABIT   DAYS  TREND\n" +
		"guitar   180  ▁▂▃\n" +
		"café       7  " + green + "\n"
	if got := table.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
