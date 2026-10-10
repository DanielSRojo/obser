package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

func TestGetStatistics(t *testing.T) {
	setupVault(t, map[string]string{
		"2024-03-01.md": "---\ntype: journal\nreading: 1 hour\npages: 10\n---\n",
		"2024-03-02.md": "---\ntype: journal\nreading: 30 minutes\npages: 5\n---\n",
		"2024-12-31.md": "---\ntype: journal\npages: 7\n---\n",
	})

	month, err := GetStatistics(2024, 3)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, p := range month {
		got[p.Name] = p.Value
	}
	if got["reading"] != 5400 || got["pages"] != 15 {
		t.Errorf("monthly totals = %v", got)
	}

	year, err := GetStatistics(2024, 0)
	if err != nil {
		t.Fatal(err)
	}
	got = map[string]int{}
	for _, p := range year {
		got[p.Name] = p.Value
	}
	if got["pages"] != 22 {
		t.Errorf("yearly pages = %d, want 22 (December must be included)", got["pages"])
	}
}

func TestMissingVaultReturnsError(t *testing.T) {
	prev := vaultDir
	SetVaultDir(filepath.Join(t.TempDir(), "missing"))
	t.Cleanup(func() { SetVaultDir(prev) })

	if _, err := GetJournalEntries(); err == nil {
		t.Error("expected error for missing vault")
	}
	if _, err := GetStatistics(2024, 3); err == nil {
		t.Error("expected error for missing vault")
	}
}

func TestMonthlyStatisticsUseJournalEntriesOnly(t *testing.T) {
	setupVault(t, map[string]string{
		"2023-11-28.md":        "---\ntype: journal\nguitar: 30 minutes\njournaling: 150 words\nchain: true\n---\n",
		"2023-11-29.md":        "---\ntype: journal\nguitar: 30 minutes\njournaling: 120\nchain: false\n---\n",
		"2023-11-29-guitar.md": "---\ntype: journal\nguitar: 60 minutes\n---\n# Practice",
	})

	month, err := GetStatistics(2023, 11)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Property{}
	for _, p := range month {
		got[p.Name] = p
	}
	if p := got["guitar"]; p.Value != 3600 || p.Unit != "seconds" {
		t.Errorf("guitar = %d %s, want 3600 seconds (the -guitar note must not count)", p.Value, p.Unit)
	}
	if p := got["journaling"]; p.Value != 270 || p.Unit != "words" {
		t.Errorf("journaling = %d %s, want 270 words", p.Value, p.Unit)
	}
	if p := got["chain"]; p.Value != 1 {
		t.Errorf("chain = %d, want 1 day done", p.Value)
	}
}
