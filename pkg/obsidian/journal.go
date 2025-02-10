package obsidian

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var vaultDir string

func init() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to determine user home directory: " + err.Error())
	}
	vaultDir = filepath.Join(userHome, ".obsidian")
}

func GetJournalEntries() ([]string, error) {
	var entries []string
	files, err := os.ReadDir(vaultDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			note := &Note{
				Content:    "",
				Directory:  vaultDir,
				Properties: []Property{},
				Title:      file.Name(),
			}
			if note.IsJournal() {
				entries = append(entries, file.Name())
			}
		}
	}
	return entries, nil
}

// AggregateMonthlyProperties aggregates numeric properties for given month
func AggregateMonthlyProperties(year, month int) (map[string]Property, error) {
	journalFiles, err := os.ReadDir(vaultDir)
	if err != nil {
		log.Fatal(err)
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
		err = note.LoadProperties()
		if err != nil || len(note.Properties) < 1 {
			continue
		}

		for _, property := range note.Properties {
			var unit string
			if totals[property.Name].Unit != "" {
				unit = totals[property.Name].Unit
			} else if totals[property.Name].Unit == "" && property.Unit != "" {
				unit = property.Unit
			}
			totals[property.Name] = Property{
				Name:  property.Name,
				Value: totals[property.Name].Value + property.Value,
				Unit:  unit,
			}
		}
	}

	return totals, nil
}

func AggregateProperty(propertyName string) ([]int, error) {
	journalFiles, err := os.ReadDir(vaultDir)
	if err != nil {
		log.Fatal(err)
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
					fmt.Printf("warning: diferent units:\n%s\n%s\n", p.Unit, propertyTotal.Unit)
					fmt.Println(journalFile.Name())
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
			log.Println(err)
			continue
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
		if !note.IsJournal() {
			continue
		}

		journalFiles = append(journalFiles, f)
	}

	return journalFiles, nil
}
