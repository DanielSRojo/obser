package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	if _, err := cfg.dateLayout(); err != nil {
		return cfg, fmt.Errorf("%s: %w", file, err)
	}
	return cfg, nil
}

// dateLayout returns the Go time layout of a daily note's path relative to the
// journal folder, without the .md extension. Of Obsidian's moment.js format,
// only the YYYY, MM and DD tokens are supported, separated by -, _, . or
// spaces, and by / for subfolders, as in YYYY/YYYY-MM-DD.
func (cfg DailyNotesConfig) dateLayout() (string, error) {
	if cfg.Format == "" {
		return journalDateLayout, nil
	}
	unsupported := fmt.Errorf("unsupported daily note format %q", cfg.Format)
	for _, token := range []string{"YYYY", "MM", "DD"} {
		if !strings.Contains(cfg.Format, token) {
			return "", unsupported
		}
	}
	layout := strings.NewReplacer("YYYY", "2006", "MM", "01", "DD", "02").Replace(cfg.Format)
	for _, segment := range strings.Split(cfg.Format, "/") {
		rest := strings.NewReplacer("YYYY", "", "MM", "", "DD", "").Replace(segment)
		if rest == segment || strings.Trim(rest, "-_. ") != "" {
			return "", unsupported
		}
	}
	return layout, nil
}

// journalEntry is a daily note in the journal folder.
type journalEntry struct {
	Name string    // path relative to the journal folder, with / separators
	Date time.Time // local midnight
}

// journalEntries returns the journal folder and the daily notes in it, sorted
// by date: the .md files whose path relative to the folder is exactly the
// configured format for a valid date. Only the subfolders the format names are
// searched, and hidden ones never are.
func (cfg DailyNotesConfig) journalEntries() (string, []journalEntry, error) {
	dir, err := cfg.journalDir()
	if err != nil {
		return "", nil, err
	}
	layout, err := cfg.dateLayout()
	if err != nil {
		return "", nil, err
	}
	depth := strings.Count(layout, "/")

	var entries []journalEntry
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if IsHidden(rel) || strings.Count(rel, "/") >= depth {
				return fs.SkipDir
			}
			return nil
		}
		name, ok := strings.CutSuffix(rel, ".md")
		if !ok {
			return nil
		}
		date, err := time.ParseInLocation(layout, name, time.Local)
		if err != nil || date.Format(layout) != name {
			return nil
		}
		entries = append(entries, journalEntry{Name: rel, Date: date})
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Date.Before(entries[j].Date) })
	return dir, entries, nil
}

// journalEntries returns the configured journal folder and its daily notes.
func journalEntries() (string, []journalEntry, error) {
	cfg, err := readDailyNotesConfig()
	if errors.Is(err, fs.ErrNotExist) {
		cfg, err = DailyNotesConfig{}, nil
	}
	if err != nil {
		return "", nil, err
	}
	return cfg.journalEntries()
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
	_, entries, err := cfg.journalEntries()
	if err != nil {
		return time.Time{}, err
	}

	var last time.Time
	for _, e := range entries {
		date := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, time.UTC)
		if !date.After(until) {
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
	layout, err := cfg.dateLayout()
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
		name := filepath.Join(dir, filepath.FromSlash(day.Format(layout))+".md")
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
			if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
				return created, err
			}
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
