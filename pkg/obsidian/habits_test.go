package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func localDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, s, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// testJournal builds a journal from frontmatter bodies keyed by date. Every
// entry gets type: journal.
func testJournal(t *testing.T, today string, entries map[string]string) *Journal {
	t.Helper()
	var days []Day
	var properties [][]Property
	for d, frontmatter := range entries {
		props, err := GetProperties("---\ntype: journal\n" + frontmatter + "\n---\n")
		if err != nil {
			t.Fatalf("%s: %v", d, err)
		}
		days = append(days, Day{Date: localDate(t, d)})
		properties = append(properties, props)
	}
	return newJournal(localDate(t, today), days, properties)
}

func journalDates(j *Journal) []string {
	var dates []string
	for _, d := range j.Days {
		dates = append(dates, d.Date.Format(time.DateOnly))
	}
	return dates
}

func TestLoadJournalSelectsEntries(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "dated note that is not a journal entry",
			files: map[string]string{
				"2023-11-29.md":        "---\ntype: journal\nguitar: 30 minutes\n---\n",
				"2023-11-29-guitar.md": "---\ntype: journal\nguitar: 60 minutes\n---\n",
			},
			want: []string{"2023-11-29"},
		},
		{
			name: "note without journal type",
			files: map[string]string{
				"2024-05-01.md": "---\nguitar: 30 minutes\n---\n",
				"2024-05-02.md": "---\ntype: Journal\nguitar: 30 minutes\n---\n",
				"2024-05-03.md": "no frontmatter",
			},
			want: []string{"2024-05-01", "2024-05-02", "2024-05-03"},
		},
		{
			name: "invalid date",
			files: map[string]string{
				"2025-02-30.md": "---\ntype: journal\nguitar: 30 minutes\n---\n",
				"2025-02-28.md": "---\ntype: journal\nguitar: 30 minutes\n---\n",
			},
			want: []string{"2025-02-28"},
		},
		{
			name: "future entry",
			files: map[string]string{
				"2026-10-10.md": "---\ntype: journal\nguitar: 30 minutes\n---\n",
				"2026-10-11.md": "---\ntype: journal\nguitar: 30 minutes\n---\n",
			},
			want: []string{"2026-10-10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupVault(t, tt.files)
			j, err := LoadJournal(localDate(t, "2026-10-10"))
			if err != nil {
				t.Fatal(err)
			}
			if got := journalDates(j); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entries = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadJournalBrokenFrontmatter(t *testing.T) {
	setupVault(t, map[string]string{"2025-01-02.md": "---\ntype: journal\nguitar: 30 minutes\n"})
	_, err := LoadJournal(localDate(t, "2026-10-10"))
	if err == nil || !strings.Contains(err.Error(), "2025-01-02.md") {
		t.Errorf("err = %v, want an error naming 2025-01-02.md", err)
	}
}

func TestLoadJournalConfiguredFolder(t *testing.T) {
	setupVault(t, map[string]string{"2025-01-01.md": "---\ntype: journal\nguitar: 30 minutes\n---\n"})
	writeDailyNotesConfig(t, `{"folder": "Journal"}`)
	if err := os.Mkdir(filepath.Join(vaultDir, "Journal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Journal", "2025-01-02.md"), []byte("---\ntype: journal\nguitar: 30 minutes\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	j, err := LoadJournal(localDate(t, "2026-10-10"))
	if err != nil {
		t.Fatal(err)
	}
	if got := journalDates(j); !reflect.DeepEqual(got, []string{"2025-01-02"}) {
		t.Errorf("entries = %v, want only Journal/2025-01-02", got)
	}
}

func TestDoneDays(t *testing.T) {
	j := testJournal(t, "2025-01-31", map[string]string{
		"2025-01-01": "guitar:\nchain: false\nmood: good\nsex: 0",
		"2025-01-02": "guitar: 30 minutes\nchain: true\nmood:\nsex: 1",
	})
	tests := []struct {
		habit, day string
		done       bool
	}{
		{"guitar", "2025-01-01", false},
		{"guitar", "2025-01-02", true},
		{"chain", "2025-01-01", false},
		{"chain", "2025-01-02", true},
		{"sex", "2025-01-01", false},
		{"sex", "2025-01-02", true},
		{"guitar", "2025-01-03", false}, // no entry
	}
	for _, tt := range tests {
		if got := j.Value(tt.habit, localDate(t, tt.day)).Done; got != tt.done {
			t.Errorf("%s on %s: done = %v, want %v", tt.habit, tt.day, got, tt.done)
		}
	}
}

func TestHabitIdentification(t *testing.T) {
	j := testJournal(t, "2025-07-31", map[string]string{
		"2025-05-01": "dopamineFast:\nchain: false\nmood: good\nwake: \"07:30\"\nread: 10 minutes",
		"2025-05-31": "dopamineFast:\nchain: false\nread: done",
		"2025-06-01": "dopamineFast: true\nchain: false",
		"2025-06-02": "dopamineFast: false\nchain: false\naliases: [x]\ntags: [daily]",
	})

	if got, want := j.HabitNames(), []string{"dopamineFast", "read"}; !reflect.DeepEqual(got, want) {
		t.Errorf("habits = %v, want %v (chain is never done, mood and wake are text only)", got, want)
	}
	if start := j.Habits["dopamineFast"].Start; !start.Equal(localDate(t, "2025-06-01")) {
		t.Errorf("dopamineFast start = %v, want 2025-06-01", start)
	}
	if !j.Value("read", localDate(t, "2025-05-31")).Done {
		t.Error("non-empty text on a habit should count as done")
	}

	// Removing the empty keys doesn't change the tracking start.
	tidy := testJournal(t, "2025-07-31", map[string]string{
		"2025-05-01": "chain: false",
		"2025-05-31": "chain: false",
		"2025-06-01": "dopamineFast: true",
		"2025-06-02": "dopamineFast: false",
	})
	if !tidy.Habits["dopamineFast"].Start.Equal(j.Habits["dopamineFast"].Start) {
		t.Error("empty keys before the first done day moved the tracking start")
	}
}

func amounts(j *Journal, habit string) (total int, withAmount, done int) {
	for _, d := range j.Days {
		v := d.Values[habit]
		if v.HasAmount {
			total += v.Amount
			withAmount++
		}
		if v.Done {
			done++
		}
	}
	return total, withAmount, done
}

func TestUnitResolution(t *testing.T) {
	j := testJournal(t, "2025-01-31", map[string]string{
		"2025-01-01": "journaling: 150 words\nvideogames: 2 hours\nread: 10 minutes\nsex: 1",
		"2025-01-02": "journaling: 120\nvideogames: 30 minutes\nread: 12 pages\nsex: 2",
		"2025-01-03": "read: 1 hour\nguitar: 25",
		"2025-01-04": "guitar: 1 hour",
	})

	if h := j.Habits["journaling"]; h.Unit != "words" || !h.HasAmounts {
		t.Errorf("journaling = %+v, want unit words", h)
	}
	if total, _, _ := amounts(j, "journaling"); total != 270 {
		t.Errorf("journaling total = %d, want 270", total)
	}

	if h := j.Habits["videogames"]; h.Unit != "seconds" {
		t.Errorf("videogames unit = %q, want seconds", h.Unit)
	}
	if total, _, _ := amounts(j, "videogames"); total != 9000 {
		t.Errorf("videogames total = %d, want 9000", total)
	}

	if total, _, _ := amounts(j, "guitar"); total != 5100 {
		t.Errorf("guitar total = %d, want 5100 (a bare 25 next to durations is 25 minutes)", total)
	}

	if h := j.Habits["sex"]; h.Unit != "" || !h.HasAmounts {
		t.Errorf("sex = %+v, want a unitless count", h)
	}

	if h := j.Habits["read"]; h.Unit != "seconds" {
		t.Errorf("read unit = %q, want seconds", h.Unit)
	}
	total, withAmount, done := amounts(j, "read")
	if total != 4200 || withAmount != 2 || done != 3 {
		t.Errorf("read: total %d over %d days, %d done; want 4200 over 2 days, 3 done", total, withAmount, done)
	}
	if len(j.Warnings) != 1 {
		t.Fatalf("warnings = %q, want one", j.Warnings)
	}
	for _, s := range []string{"read", "seconds", "pages"} {
		if !strings.Contains(j.Warnings[0], s) {
			t.Errorf("warning %q doesn't name %q", j.Warnings[0], s)
		}
	}
}

func TestUnitResolutionMostFrequentWins(t *testing.T) {
	j := testJournal(t, "2025-01-31", map[string]string{
		"2025-01-01": "read: 10 pages",
		"2025-01-02": "read: 12 pages",
		"2025-01-03": "read: 30 minutes",
		"2025-01-04": "read: 5",
	})
	if h := j.Habits["read"]; h.Unit != "pages" {
		t.Errorf("unit = %q, want pages", h.Unit)
	}
	if total, withAmount, done := amounts(j, "read"); total != 22 || withAmount != 2 || done != 4 {
		t.Errorf("total %d over %d days, %d done; want 22 over 2 days, 4 done (a bare number is not counted when units conflict)", total, withAmount, done)
	}
}
