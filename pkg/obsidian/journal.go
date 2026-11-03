package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var vaultDir string

// DefaultVaultDir returns the vault used when none is configured: ~/.obsidian.
func DefaultVaultDir() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user home directory: %w", err)
	}
	return filepath.Join(userHome, ".obsidian"), nil
}

// SetVaultDir sets the vault directory all other functions operate on.
func SetVaultDir(dir string) {
	vaultDir = dir
}

// GetJournalEntries returns the paths, relative to the journal folder, of the
// daily notes with type journal.
func GetJournalEntries() ([]string, error) {
	_, notes, err := typedJournalNotes()
	if err != nil {
		return nil, err
	}
	var entries []string
	for _, n := range notes {
		entries = append(entries, filepath.ToSlash(n.Title))
	}
	return entries, nil
}

func GetPropertiesNames() ([]string, error) {
	_, notes, err := typedJournalNotes()
	if err != nil {
		return nil, err
	}
	var propertyList []string
	seen := make(map[string]bool)
	for _, note := range notes {
		for _, p := range note.Properties {
			if p.Name == "" || seen[p.Name] {
				continue
			}
			propertyList = append(propertyList, p.Name)
			seen[p.Name] = true
		}
	}
	sort.Strings(propertyList)

	return propertyList, nil
}

// typedJournalNotes returns the journal folder and the daily notes in it with
// type journal, with their properties loaded.
func typedJournalNotes() (string, []Note, error) {
	dir, entries, err := journalEntries()
	if err != nil {
		return "", nil, err
	}

	var notes []Note
	for _, e := range entries {
		note := Note{Directory: dir, Title: filepath.FromSlash(e.Name)}
		if err := note.LoadProperties(); err != nil {
			return "", nil, err
		}
		if isJournal(note.Properties) {
			notes = append(notes, note)
		}
	}
	return dir, notes, nil
}
