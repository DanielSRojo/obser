package obsidian

import (
	"fmt"
	"io/fs"
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

func GetJournalEntries() ([]string, error) {
	files, err := GetJournalFiles()
	if err != nil {
		return nil, err
	}
	var entries []string
	for _, file := range files {
		entries = append(entries, file.Name())
	}
	return entries, nil
}

func GetPropertiesNames() ([]string, error) {
	dir, err := JournalDir()
	if err != nil {
		return nil, err
	}
	journalFiles, err := GetJournalFiles()
	if err != nil {
		return nil, err
	}
	var propertyList []string
	seen := make(map[string]bool)
	for _, journalFile := range journalFiles {
		note := Note{
			Directory: dir,
			Title:     journalFile.Name(),
		}
		if err := note.LoadProperties(); err != nil {
			return nil, err
		}

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

// GetJournalFiles returns the notes with type journal in the journal folder.
func GetJournalFiles() ([]fs.DirEntry, error) {
	dir, err := JournalDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var journalFiles []fs.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		note := &Note{
			Directory: dir,
			Title:     f.Name(),
		}
		isJournal, err := note.IsJournal()
		if err != nil {
			return nil, err
		}
		if !isJournal {
			continue
		}

		journalFiles = append(journalFiles, f)
	}

	return journalFiles, nil
}
