package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Note struct {
	Content    string
	Directory  string
	Properties []Property
	Title      string
	Type       string
}

func (n *Note) LoadProperties() error {
	file := filepath.Join(n.Directory, n.Title)
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	n.Properties, err = GetProperties(string(data))
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}

func GetNotesNames() ([]string, error) {
	var names []string
	files, err := os.ReadDir(vaultDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || IsHidden(file.Name()) {
			continue
		}
		names = append(names, file.Name())
	}

	return names, nil
}

func (n *Note) IsJournal() (bool, error) {
	if err := n.LoadProperties(); err != nil {
		return false, err
	}

	for _, p := range n.Properties {
		if strings.ToLower(p.Name) == "type" && strings.ToLower(p.Content) == "journal" {
			return true, nil
		}
	}

	return false, nil
}

func IsHidden(path string) bool {
	return filepath.Base(path)[0] == '.'
}
