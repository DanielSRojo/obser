package habits

import (
	"testing"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
)

func TestMonthRange(t *testing.T) {
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.Local) }
	j := &obsidian.Journal{
		Today: day(2026, 10, 10),
		Days:  []obsidian.Day{{Date: day(2023, 6, 15)}},
	}

	tests := []struct {
		year, month int
		want        obsidian.Range
		wantErr     string
	}{
		{year: 2024, month: 3, want: obsidian.Range{From: day(2024, 3, 1), To: day(2024, 3, 31)}},
		{year: 2024, month: 2, want: obsidian.Range{From: day(2024, 2, 1), To: day(2024, 2, 29)}},
		{year: 2026, month: 10, want: obsidian.Range{From: day(2026, 10, 1), To: day(2026, 10, 10)}},
		{year: 2023, month: 6, want: obsidian.Range{From: day(2023, 6, 1), To: day(2023, 6, 30)}},
		{year: 2023, month: 5, wantErr: "no journal entries for May 2023"},
		{year: 2026, month: 11, wantErr: "no journal entries for November 2026"},
		{year: 2024, month: 13, wantErr: "invalid month 13, must be 1-12"},
		{year: 2024, month: 0, wantErr: "invalid month 0, must be 1-12"},
	}
	for _, tt := range tests {
		got, err := monthRange(j, tt.year, tt.month)
		if tt.wantErr != "" {
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("monthRange(%d, %d) error = %v, want %q", tt.year, tt.month, err, tt.wantErr)
			}
			continue
		}
		if err != nil || !got.From.Equal(tt.want.From) || !got.To.Equal(tt.want.To) {
			t.Errorf("monthRange(%d, %d) = %v, %v, want %v", tt.year, tt.month, got, err, tt.want)
		}
	}
}
