package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func setupVault(t *testing.T, files map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	prev := vaultDir
	SetVaultDir(dir)
	t.Cleanup(func() { SetVaultDir(prev) })
}

func TestGetJournalEntries(t *testing.T) {
	setupVault(t, map[string]string{
		"2024-03-01.md": "---\ntype: journal\n---\n",
		"2024-03-02.md": "---\ntype: Journal\nreading: 1 hour\n---\n",
		"idea.md":       "---\ntype: note\n---\n",
		"plain.md":      "no frontmatter",
	})

	got, err := GetJournalEntries()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"2024-03-01.md", "2024-03-02.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestStatsMonthAndYearTotals(t *testing.T) {
	setupVault(t, map[string]string{
		"2024-03-01.md": "---\ntype: journal\nreading: 1 hour\npages: 10\n---\n",
		"2024-03-02.md": "---\ntype: journal\nreading: 30 minutes\npages: 5\n---\n",
		"2024-12-31.md": "---\ntype: journal\npages: 7\n---\n",
	})
	j, err := LoadJournal(time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}

	march := Range{time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local), time.Date(2024, 3, 31, 0, 0, 0, 0, time.Local)}
	if got := j.Stats("reading", march).Total; got != 5400 {
		t.Errorf("March reading = %d, want 5400", got)
	}
	if got := j.Stats("pages", march).Total; got != 15 {
		t.Errorf("March pages = %d, want 15", got)
	}

	year := Range{time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)}
	if got := j.Stats("pages", year).Total; got != 22 {
		t.Errorf("2024 pages = %d, want 22 (December must be included)", got)
	}
}

func TestMissingVaultReturnsError(t *testing.T) {
	prev := vaultDir
	SetVaultDir(filepath.Join(t.TempDir(), "missing"))
	t.Cleanup(func() { SetVaultDir(prev) })

	if _, err := GetJournalEntries(); err == nil {
		t.Error("expected error for missing vault")
	}
	if _, err := LoadJournal(time.Now()); err == nil {
		t.Error("expected error for missing vault")
	}
}

func TestMonthStatsUseJournalEntriesOnly(t *testing.T) {
	setupVault(t, map[string]string{
		"2023-11-28.md":        "---\ntype: journal\nguitar: 30 minutes\njournaling: 150 words\nchain: true\n---\n",
		"2023-11-29.md":        "---\ntype: journal\nguitar: 30 minutes\njournaling: 120\nchain: false\n---\n",
		"2023-11-29-guitar.md": "---\ntype: journal\nguitar: 60 minutes\n---\n# Practice",
	})
	j, err := LoadJournal(time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}

	november := Range{time.Date(2023, 11, 1, 0, 0, 0, 0, time.Local), time.Date(2023, 11, 30, 0, 0, 0, 0, time.Local)}
	if s, unit := j.Stats("guitar", november), j.Habits["guitar"].Unit; s.Total != 3600 || unit != "seconds" {
		t.Errorf("guitar = %d %s, want 3600 seconds (the -guitar note must not count)", s.Total, unit)
	}
	if s, unit := j.Stats("journaling", november), j.Habits["journaling"].Unit; s.Total != 270 || unit != "words" {
		t.Errorf("journaling = %d %s, want 270 words", s.Total, unit)
	}
	if s := j.Stats("chain", november); s.DaysDone != 1 || s.HasTotal {
		t.Errorf("chain = %d days done (has total %v), want 1 day without a total", s.DaysDone, s.HasTotal)
	}
}
