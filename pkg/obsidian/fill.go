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

// LoadDailyNotesConfig reads the Daily notes configuration of the vault. Unlike
// JournalDir, it requires the file to exist and to name a template.
func LoadDailyNotesConfig() (DailyNotesConfig, error) {
	cfg, err := readDailyNotesConfig()
	if err != nil {
		return cfg, err
	}
	if cfg.Template == "" {
		return cfg, fmt.Errorf("%s: no template configured", dailyNotesConfigFile())
	}
	if _, err := cfg.journalDir(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// JournalDir returns the folder that holds the journal entries: the folder
// configured in the Daily notes plugin, or the vault root when there is no
// configuration or no folder is set.
func JournalDir() (string, error) {
	cfg, err := readDailyNotesConfig()
	if errors.Is(err, fs.ErrNotExist) {
		cfg, err = DailyNotesConfig{}, nil
	}
	if err != nil {
		return "", err
	}
	return cfg.journalDir()
}

func dailyNotesConfigFile() string {
	return filepath.Join(vaultDir, ".obsidian", "daily-notes.json")
}

// readDailyNotesConfig parses the Daily notes configuration and checks the
// date format. A missing file is returned as an fs.ErrNotExist error.
func readDailyNotesConfig() (DailyNotesConfig, error) {
	var cfg DailyNotesConfig
	file := dailyNotesConfigFile()
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
	return cfg, nil
}

// journalDir returns the configured journal folder and checks that it exists.
func (cfg DailyNotesConfig) journalDir() (string, error) {
	dir := filepath.Join(vaultDir, cfg.Folder)
	info, err := os.Stat(dir)
	if err != nil {
		if cfg.Folder == "" {
			return "", fmt.Errorf("vault %s: %w", dir, err)
		}
		return "", fmt.Errorf("journal folder %q: %w", cfg.Folder, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("journal folder %q: not a directory", cfg.Folder)
	}
	return dir, nil
}

// LastJournalDate returns the date of the most recent daily note on or before
// the given day, or the zero time if there is none.
func LastJournalDate(cfg DailyNotesConfig, until time.Time) (time.Time, error) {
	dir, err := cfg.journalDir()
	if err != nil {
		return time.Time{}, err
	}
	files, err := os.ReadDir(dir)
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
	dir, err := cfg.journalDir()
	if err != nil {
		return nil, err
	}
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
		name := filepath.Join(dir, day.Format(journalDateLayout)+".md")
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
