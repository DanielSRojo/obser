package habits

import "testing"

func TestDelta(t *testing.T) {
	tests := []struct {
		before, after float64
		want          string
	}{
		{0.4, 0.6, "↑"},
		{0.6, 0.4, "↓"},
		{100, 110, "→"}, // exactly 10% is not more than 10%
		{100, 111, "↑"},
		{100, 90, "→"},
		{100, 89, "↓"},
		{0, 5, "↑"},
		{0, 0, "→"},
		{5, 0, "↓"},
	}
	for _, tt := range tests {
		if got := delta(tt.before, tt.after); got != tt.want {
			t.Errorf("delta(%v, %v) = %s, want %s", tt.before, tt.after, got, tt.want)
		}
	}
}

func TestShadeLevel(t *testing.T) {
	tests := []struct {
		value, highest float64
		want           int
	}{
		{0, 10, 0},
		{0, 0, 0},
		{1, 10, 1},
		{2.5, 10, 1},
		{2.6, 10, 2},
		{7.5, 10, 3},
		{10, 10, 4},
	}
	for _, tt := range tests {
		if got := shadeLevel(tt.value, tt.highest); got != tt.want {
			t.Errorf("shadeLevel(%v, %v) = %d, want %d", tt.value, tt.highest, got, tt.want)
		}
	}
}
