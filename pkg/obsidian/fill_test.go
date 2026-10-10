package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const journalTemplate = "---\ntype: journal\n---\n" +
	"[[<% fileDate = moment(tp.file.title, 'YYYY-MM-DD').subtract(1, 'd').format('YYYY-MM-DD') %> |<< Yesterday]] - " +
	"[[<% fileDate = moment(tp.file.title, 'YYYY-MM-DD').add(1, 'd').format('YYYY-MM-DD') %> |Tomorrow >>]]\n" +
	"# <% fileDate = moment(tp.file.title, 'YYYY-MM-DD').format('YYYY-MM-DD') %>\n" +
	"WHERE file.cday = date(\"<% moment(tp.file.title, 'YYYY-MM-DD').format('YYYY-MM-DD') %>\")"

func date(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestRenderTemplate(t *testing.T) {
	got, err := RenderTemplate(journalTemplate, date(t, "2026-03-01"))
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntype: journal\n---\n" +
		"[[2026-02-28 |<< Yesterday]] - [[2026-03-02 |Tomorrow >>]]\n" +
		"# 2026-03-01\n" +
		"WHERE file.cday = date(\"2026-03-01\")"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderTemplateUnsupported(t *testing.T) {
	if _, err := RenderTemplate("<% tp.file.cursor() %>", date(t, "2026-03-01")); err == nil {
		t.Error("expected an error for an unsupported expression")
	}
}

func TestFillJournal(t *testing.T) {
	setupVault(t, map[string]string{
		"2024-03-01.md": "existing",
		"2024-03-03.md": "existing",
		"Journal.md":    journalTemplate,
	})
	cfg := DailyNotesConfig{Template: "Journal"}

	last, err := LastJournalDate(cfg, date(t, "2024-03-02"))
	if err != nil {
		t.Fatal(err)
	}
	if !last.Equal(date(t, "2024-03-01")) {
		t.Errorf("last journal date: got %v, want 2024-03-01", last)
	}

	created, err := FillJournal(cfg, date(t, "2024-03-01"), date(t, "2024-03-04"), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(vaultDir, "2024-03-02.md"),
		filepath.Join(vaultDir, "2024-03-04.md"),
	}
	if !reflect.DeepEqual(created, want) {
		t.Errorf("created %v, want %v", created, want)
	}

	existing, err := os.ReadFile(filepath.Join(vaultDir, "2024-03-03.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(existing) != "existing" {
		t.Errorf("existing entry was overwritten: %q", existing)
	}

	note := &Note{Directory: vaultDir, Title: "2024-03-02.md"}
	if isJournal, err := note.IsJournal(); err != nil || !isJournal {
		t.Errorf("created entry is not a journal entry: %v", err)
	}
}

func TestFillJournalDryRun(t *testing.T) {
	setupVault(t, map[string]string{"Journal.md": journalTemplate})
	cfg := DailyNotesConfig{Template: "Journal"}

	created, err := FillJournal(cfg, date(t, "2024-03-01"), date(t, "2024-03-02"), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 2 {
		t.Errorf("got %d entries, want 2", len(created))
	}
	if _, err := os.Stat(filepath.Join(vaultDir, "2024-03-01.md")); err == nil {
		t.Error("dry run created a file")
	}
}
