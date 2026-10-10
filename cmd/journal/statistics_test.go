package journal

import (
	"testing"

	"github.com/danielsrojo/obser/pkg/obsidian"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		value     int
		unit      string
		wantValue string
		wantUnit  string
	}{
		{45, "seconds", "45", "seconds"},
		{199, "seconds", "199", "seconds"},
		{200, "seconds", "3", "minutes"},
		{5400, "seconds", "90", "minutes"},
		{5969, "seconds", "99", "minutes"},
		{5970, "seconds", "1.7", "hours"},
		{12000, "seconds", "3.3", "hours"},
		{36000, "seconds", "10", "hours"},
		{470, "words", "470", "words"},
		{12, "", "12", ""},
	}
	for _, tt := range tests {
		value, unit := formatValue(obsidian.Property{Value: tt.value, Unit: tt.unit})
		if value != tt.wantValue || unit != tt.wantUnit {
			t.Errorf("formatValue(%d %s) = %s %s, want %s %s", tt.value, tt.unit, value, unit, tt.wantValue, tt.wantUnit)
		}
	}
}
