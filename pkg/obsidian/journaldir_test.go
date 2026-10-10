package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		{name: "unsupported format", config: `{"format": "DD-MM-YYYY"}`, wantErr: "DD-MM-YYYY"},
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
