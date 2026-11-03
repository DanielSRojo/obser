package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeDailyNotesConfig(t *testing.T, json string) {
	t.Helper()
	dir := filepath.Join(vaultDir, ".obsidian")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "daily-notes.json"), []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestJournalDir(t *testing.T) {
	tests := []struct {
		name    string
		config  string // "" means no daily-notes.json
		mkdir   string
		want    string // relative to the vault
		wantErr string
	}{
		{name: "no config", want: "."},
		{name: "empty folder", config: `{"template": "Journal"}`, want: "."},
		{name: "configured folder", config: `{"folder": "Journal"}`, mkdir: "Journal", want: "Journal"},
		{name: "no template is fine", config: `{"folder": "Journal", "format": "YYYY-MM-DD"}`, mkdir: "Journal", want: "Journal"},
		{name: "missing folder", config: `{"folder": "Daily"}`, wantErr: "Daily"},
		{name: "year subfolder format", config: `{"folder": "Journal", "format": "YYYY/YYYY-MM-DD"}`, mkdir: "Journal", want: "Journal"},
		{name: "unsupported token", config: `{"format": "dddd, MMMM Do YYYY"}`, wantErr: "dddd, MMMM Do YYYY"},
		{name: "format without day", config: `{"format": "YYYY-MM"}`, wantErr: "YYYY-MM"},
		{name: "folder without date", config: `{"format": "Daily/YYYY-MM-DD"}`, wantErr: "Daily/YYYY-MM-DD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupVault(t, nil)
			if tt.config != "" {
				writeDailyNotesConfig(t, tt.config)
			}
			if tt.mkdir != "" {
				if err := os.Mkdir(filepath.Join(vaultDir, tt.mkdir), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			got, err := JournalDir()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want an error naming %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if want := filepath.Join(vaultDir, tt.want); got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func TestLoadDailyNotesConfigRequiresTemplate(t *testing.T) {
	setupVault(t, nil)
	writeDailyNotesConfig(t, `{"folder": ""}`)
	if _, err := LoadDailyNotesConfig(); err == nil {
		t.Error("expected an error when no template is configured")
	}
}

func TestJournalFolderIsSharedByCommands(t *testing.T) {
	setupVault(t, map[string]string{
		"2024-03-01.md": "---\ntype: journal\n---\n",
		"Journal.md":    journalTemplate,
	})
	writeDailyNotesConfig(t, `{"folder": "Journal", "template": "Journal"}`)
	if err := os.Mkdir(filepath.Join(vaultDir, "Journal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Journal", "2024-03-03.md"), []byte("---\ntype: journal\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := GetJournalEntries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0] != "2024-03-03.md" {
		t.Errorf("journal entries = %v, want only Journal/2024-03-03.md", entries)
	}

	cfg, err := LoadDailyNotesConfig()
	if err != nil {
		t.Fatal(err)
	}
	last, err := LastJournalDate(cfg, date(t, "2024-03-31"))
	if err != nil {
		t.Fatal(err)
	}
	if !last.Equal(date(t, "2024-03-03")) {
		t.Errorf("last journal date = %v, want 2024-03-03", last)
	}
	created, err := FillJournal(cfg, date(t, "2024-03-04"), date(t, "2024-03-04"), true)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(vaultDir, "Journal", "2024-03-04.md"); len(created) != 1 || created[0] != want {
		t.Errorf("created %v, want [%s]", created, want)
	}
}

func TestJournalInYearSubfolders(t *testing.T) {
	setupVault(t, map[string]string{"Journal.md": journalTemplate})
	writeDailyNotesConfig(t, `{"folder": "Journal", "format": "YYYY/YYYY-MM-DD", "template": "Journal"}`)
	for name, content := range map[string]string{
		"Journal/2023/2023-12-31.md":      "---\ntype: journal\nguitar: 30 minutes\n---\n",
		"Journal/2024/2024-01-01.md":      "---\ntype: journal\nguitar:\n---\n",
		"Journal/2024-01-02.md":           "---\ntype: journal\nguitar: 1 hour\n---\n", // not in its year folder
		"Journal/2023/2024-01-03.md":      "---\ntype: journal\nguitar: 1 hour\n---\n", // wrong year folder
		"Journal/2024/sub/2024-01-04.md":  "---\ntype: journal\nguitar: 1 hour\n---\n", // too deep
		"Journal/.trash/2024-01-05.md":    "---\ntype: journal\nguitar: 1 hour\n---\n",
		"Journal/2024/2024-01-01 copy.md": "---\ntype: journal\nguitar: 1 hour\n---\n",
	} {
		path := filepath.Join(vaultDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := GetJournalEntries()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"2023/2023-12-31.md", "2024/2024-01-01.md"}; !reflect.DeepEqual(entries, want) {
		t.Errorf("journal entries = %v, want %v", entries, want)
	}

	j, err := LoadJournal(time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(j.Days) != 2 {
		t.Errorf("habits read %d days, want 2", len(j.Days))
	}

	results, skipped, err := TidyJournal(time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local), true)
	if err != nil || len(skipped) != 0 {
		t.Fatal(err, skipped)
	}
	if want := filepath.Join(vaultDir, "Journal", "2024", "2024-01-01.md"); len(results) != 1 || results[0].Path != want {
		t.Errorf("tidy results = %v, want only %s", results, want)
	}

	cfg, err := LoadDailyNotesConfig()
	if err != nil {
		t.Fatal(err)
	}
	last, err := LastJournalDate(cfg, date(t, "2024-01-31"))
	if err != nil {
		t.Fatal(err)
	}
	if !last.Equal(date(t, "2024-01-01")) {
		t.Errorf("last journal date = %v, want 2024-01-01", last)
	}
	created, err := FillJournal(cfg, date(t, "2024-12-31"), date(t, "2025-01-01"), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(vaultDir, "Journal", "2024", "2024-12-31.md"),
		filepath.Join(vaultDir, "Journal", "2025", "2025-01-01.md"),
	}
	if !reflect.DeepEqual(created, want) {
		t.Errorf("created %v, want %v", created, want)
	}
	for _, name := range want {
		if _, err := os.Stat(name); err != nil {
			t.Error(err)
		}
	}
}
