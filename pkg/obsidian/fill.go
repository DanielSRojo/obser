package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const journalDateLayout = "2006-01-02"

// DailyNotesConfig is the configuration of Obsidian's Daily notes core plugin,
// stored in .obsidian/daily-notes.json inside the vault.
type DailyNotesConfig struct {
	Folder   string `json:"folder"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

// LoadDailyNotesConfig reads the Daily notes configuration of the vault.
func LoadDailyNotesConfig() (DailyNotesConfig, error) {
	var cfg DailyNotesConfig
	file := filepath.Join(vaultDir, ".obsidian", "daily-notes.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return cfg, fmt.Errorf("daily notes config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", file, err)
	}
	if cfg.Format != "" && cfg.Format != "YYYY-MM-DD" {
		return cfg, fmt.Errorf("%s: unsupported daily note format %q", file, cfg.Format)
	}
	if cfg.Template == "" {
		return cfg, fmt.Errorf("%s: no template configured", file)
	}
	return cfg, nil
}

// LastJournalDate returns the date of the most recent daily note on or before
// the given day, or the zero time if there is none.
func LastJournalDate(cfg DailyNotesConfig, until time.Time) (time.Time, error) {
	files, err := os.ReadDir(filepath.Join(vaultDir, cfg.Folder))
	if err != nil {
		return time.Time{}, err
	}

	var last time.Time
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		date, err := time.Parse(journalDateLayout, strings.TrimSuffix(f.Name(), ".md"))
		if err != nil || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		if date.After(until) {
			continue
		}
		if date.After(last) {
			last = date
		}
	}
	return last, nil
}

// FillJournal creates a daily note from the template for every day between
// from and to (inclusive) that doesn't have one yet, and returns the paths of
// the notes it created. With dryRun set, it only reports what it would create.
func FillJournal(cfg DailyNotesConfig, from, to time.Time, dryRun bool) ([]string, error) {
	templateFile := filepath.Join(vaultDir, cfg.Template)
	if filepath.Ext(templateFile) != ".md" {
		templateFile += ".md"
	}
	template, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("journal template: %w", err)
	}

	var created []string
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		name := filepath.Join(vaultDir, cfg.Folder, day.Format(journalDateLayout)+".md")
		if _, err := os.Stat(name); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return created, err
		}

		content, err := RenderTemplate(string(template), day)
		if err != nil {
			return created, fmt.Errorf("%s: %w", templateFile, err)
		}

		if !dryRun {
			f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				return created, err
			}
			_, err = f.WriteString(content)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				return created, fmt.Errorf("%s: %w", name, err)
			}
		}
		created = append(created, name)
	}
	return created, nil
}
