package obsidian

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// AggregateMonthlyProperties aggregates numeric properties for given month
func AggregateMonthlyProperties(year, month int) (map[string]Property, error) {
	journalFiles, err := os.ReadDir(vaultDir)
	if err != nil {
		return nil, err
	}

	totals := make(map[string]Property)
	for _, journalFile := range journalFiles {
		prefix := fmt.Sprintf("%04d-%02d", year, month)
		if !strings.HasPrefix(journalFile.Name(), prefix) {
			continue
		}

		note := &Note{
			Title:     journalFile.Name(),
			Directory: vaultDir,
		}
		if err := note.LoadProperties(); err != nil {
			return nil, err
		}

		for _, property := range note.Properties {
			if totals[property.Name].Unit != "" {
				property.Unit = totals[property.Name].Unit
			}
			if property.Unit == "hours" {
				property.Value = property.Value * 60
				property.Unit = "minutes"
			}
			if property.Unit == "minutes" {
				property.Value = property.Value * 60
				property.Unit = "seconds"
			}
			totals[property.Name] = Property{
				Name:  property.Name,
				Value: totals[property.Name].Value + property.Value,
				Unit:  property.Unit,
				Type:  Numeric,
			}
		}
	}

	return totals, nil
}

func AggregateProperty(propertyName string) ([]int, error) {
	journalFiles, err := os.ReadDir(vaultDir)
	if err != nil {
		return nil, err
	}

	totalCount := make([]int, 12)
	var propertyTotal Property

	for i := 0; i <= 11; i++ { // 12 months
		var totals int

		// TODO: change this var to be an input
		year := 2024
		for _, journalFile := range journalFiles {
			prefix := fmt.Sprintf("%04d-%02d", year, i+1)
			if !strings.HasPrefix(journalFile.Name(), prefix) {
				continue
			}

			note := Note{
				Directory: vaultDir,
				Title:     journalFile.Name(),
			}
			if err := note.LoadProperties(); err != nil {
				return nil, err
			}

			for _, p := range note.Properties {
				if p.Name != propertyName {
					continue
				}

				if p.Unit != propertyTotal.Unit && propertyTotal.Unit != "" {
					// fmt.Printf("warning: diferent units:\n%s\n%s\n", p.Unit, propertyTotal.Unit)
					// fmt.Println(journalFile.Name())
					continue
				}

				totals += p.Value
			}
		}
		totalCount[i] = totals
	}

	return totalCount, nil
}

func GetPropertiesNames() ([]string, error) {
	journalFiles, err := GetJournalFiles()
	if err != nil {
		return nil, err
	}
	var propertyList []string
	seen := make(map[string]bool)
	for _, journalFile := range journalFiles {
		note := Note{
			Directory: vaultDir,
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

// contains checks if a string is in a slice
func contains(slice []string, target string) bool {
	index := sort.SearchStrings(slice, target)
	return index < len(slice) && slice[index] == target
}

func GetJournalFiles() ([]fs.DirEntry, error) {
	files, err := os.ReadDir(vaultDir)
	if err != nil {
		return nil, err
	}

	var journalFiles []fs.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		note := &Note{
			Directory: vaultDir,
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
