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

var journalDir string

func init() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to determine user home directory: " + err.Error())
	}
	journalDir = filepath.Join(userHome, ".obsidian")
}

func GetJournalEntries() ([]string, error) {
	var entries []string
	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			note := &Note{
				Content:    "",
				Directory:  journalDir,
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
	//TODO: do not take into consideration non countable properties (e.g. type)
	journalFiles, err := os.ReadDir(journalDir)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	monthSum := make(map[string]Property)
	for _, journalFile := range journalFiles {
		prefix := fmt.Sprintf("%04d-%02d", year, month)
		if !strings.HasPrefix(journalFile.Name(), prefix) {
			continue
		}

		note := &Note{
			Title:     journalFile.Name(),
			Directory: journalDir,
		}
		err = note.LoadProperties()
		if err != nil || len(note.Properties) < 2 {
			continue
		}

		for _, property := range note.Properties {
			var unit string
			if monthSum[property.Name].Unit != "" {
				unit = monthSum[property.Name].Unit
			} else if monthSum[property.Name].Unit == "" && property.Unit != "" {
				unit = property.Unit
			}
			monthSum[property.Name] = Property{
				Name:  property.Name,
				Value: monthSum[property.Name].Value + property.Value,
				Unit:  unit,
			}
		}
	}

	return monthSum, nil
}

func AggregateProperty(propertyName string) ([]int, error) {
	journalFiles, err := os.ReadDir(journalDir)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	totalCount := make([]int, 12)
	var propertyCount Property

	for i := 0; i <= 11; i++ { // 12 months
		var monthSum int

		year := 2024
		for _, journalFile := range journalFiles {
			prefix := fmt.Sprintf("%04d-%02d", year, i+1)
			if !strings.HasPrefix(journalFile.Name(), prefix) {
				continue
			}

			note := Note{
				Directory: journalDir,
				Title:     journalFile.Name(),
			}
			if err := note.LoadProperties(); err != nil {
				return nil, err
			}

			for _, p := range note.Properties {
				if p.Name != propertyName {
					continue
				}

				if p.Unit != propertyCount.Unit && propertyCount.Unit != "" {
					fmt.Printf("warning: diferent units:\n%s\n%s\n", p.Unit, propertyCount.Unit)
					fmt.Println(journalFile.Name())
					continue
				}

				monthSum += p.Value
			}
		}
		totalCount[i] = monthSum
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
			Directory: journalDir,
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
	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, err
	}

	var journalFiles []fs.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		note := &Note{
			Directory: journalDir,
			Title:     f.Name(),
		}
		if !note.IsJournal() {
			continue
		}

		journalFiles = append(journalFiles, f)
	}

	return journalFiles, nil
}
