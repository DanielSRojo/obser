package obsidian

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// TidyFrontmatter removes the top-level frontmatter properties that have no
// value (empty, null, ~, "", [] or {}) and returns the new content with the
// names of the removed properties. Only the lines of those properties are
// dropped; every other byte, including line endings, is kept. The type
// property is never removed.
func TidyFrontmatter(content []byte) (out []byte, removed []string, err error) {
	lines := splitLines(content)
	if len(lines) == 0 || !isDelimiter(lines[0], "---") {
		return content, nil, nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if isDelimiter(lines[i], "---") || isDelimiter(lines[i], "...") {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, nil, fmt.Errorf("closing frontmatter delimiter not found")
	}

	frontmatter := bytes.Join(lines[1:end], nil)
	if bytes.Contains(frontmatter, []byte("<%")) {
		return nil, nil, fmt.Errorf("frontmatter contains unrendered Templater tags")
	}
	frontmatter = bytes.ReplaceAll(frontmatter, []byte("\r\n"), []byte("\n"))
	if len(bytes.TrimSpace(frontmatter)) == 0 {
		return content, nil, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(frontmatter, &doc); err != nil {
		return nil, nil, fmt.Errorf("invalid frontmatter: %w", err)
	}
	if len(doc.Content) == 0 {
		return content, nil, nil
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode || mapping.Style&yaml.FlowStyle != 0 {
		return nil, nil, fmt.Errorf("invalid frontmatter: expected one key/value pair per line")
	}

	// Frontmatter line n (1-based) is lines[n], since lines[0] is the opening
	// delimiter.
	seen := map[string]bool{}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		name := mapping.Content[i].Value
		if seen[name] {
			return nil, nil, fmt.Errorf("invalid frontmatter: duplicate property %q", name)
		}
		seen[name] = true
	}

	drop := make([]bool, len(lines))
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		if strings.EqualFold(key.Value, "type") || !isEmptyValue(value) {
			continue
		}
		next := end
		if i+2 < len(mapping.Content) {
			next = mapping.Content[i+2].Line
		}
		// Blank lines and unindented comments before the next key belong to
		// what follows, not to this property.
		for next-1 > key.Line && isSeparator(lines[next-1]) {
			next--
		}
		for l := key.Line; l < next; l++ {
			drop[l] = true
		}
		removed = append(removed, key.Value)
	}
	if len(removed) == 0 {
		return content, nil, nil
	}

	out = make([]byte, 0, len(content))
	for i, line := range lines {
		if !drop[i] {
			out = append(out, line...)
		}
	}
	return out, removed, nil
}

// isEmptyValue reports whether a property value carries no data.
func isEmptyValue(n *yaml.Node) bool {
	switch n.Kind {
	case yaml.ScalarNode:
		return n.Tag == "!!null" || (n.Tag == "!!str" && n.Value == "")
	case yaml.SequenceNode, yaml.MappingNode:
		return len(n.Content) == 0
	}
	return false
}

// splitLines splits content into lines that keep their own terminators.
func splitLines(content []byte) [][]byte {
	var lines [][]byte
	for len(content) > 0 {
		i := bytes.IndexByte(content, '\n')
		if i < 0 {
			lines = append(lines, content)
			break
		}
		lines = append(lines, content[:i+1])
		content = content[i+1:]
	}
	return lines
}

func isDelimiter(line []byte, delim string) bool {
	return string(bytes.TrimRight(line, " \t\r\n")) == delim
}

// isSeparator reports whether a frontmatter line is blank or an unindented
// comment.
func isSeparator(line []byte) bool {
	trimmed := bytes.TrimRight(line, " \t\r\n")
	return len(trimmed) == 0 || trimmed[0] == '#'
}

// TidyResult is a journal entry that was, or with a dry run would be, tidied.
type TidyResult struct {
	Path    string
	Removed []string
}

// Skipped is a journal entry that couldn't be tidied.
type Skipped struct {
	Path string
	Err  error
}

// TidyJournal removes empty properties from the journal entries dated before
// the given day. An entry is a YYYY-MM-DD.md note in the journal folder with
// type: journal. Entries that can't be processed are skipped and reported.
// Files are rewritten in place and keep their modification time. With dryRun
// set, nothing is written.
func TidyJournal(before time.Time, dryRun bool) ([]TidyResult, []Skipped, error) {
	before = Midnight(before)
	dir, err := JournalDir()
	if err != nil {
		return nil, nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	var results []TidyResult
	var skipped []Skipped
	for _, f := range files {
		if f.IsDir() || !journalFileName.MatchString(f.Name()) {
			continue
		}
		date, err := time.ParseInLocation(journalDateLayout, strings.TrimSuffix(f.Name(), ".md"), time.Local)
		if err != nil || !date.Before(before) {
			continue
		}
		path := filepath.Join(dir, f.Name())
		removed, err := tidyEntry(path, dryRun)
		if err != nil {
			skipped = append(skipped, Skipped{Path: path, Err: err})
			continue
		}
		if len(removed) > 0 {
			results = append(results, TidyResult{Path: path, Removed: removed})
		}
	}
	return results, skipped, nil
}

// tidyEntry tidies one note if it is a journal entry and returns the removed
// properties.
func tidyEntry(path string, dryRun bool) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	properties, err := GetProperties(string(data))
	if err != nil {
		return nil, err
	}
	if !isJournal(properties) {
		return nil, nil
	}
	out, removed, err := TidyFrontmatter(data)
	if err != nil || len(removed) == 0 || dryRun {
		return removed, err
	}
	if err := rewriteInPlace(path, out); err != nil {
		return nil, err
	}
	return removed, nil
}

// rewriteInPlace replaces the content of a file without recreating it, so it
// keeps its identity, creation time and mode, and then restores its
// modification time.
func rewriteInPlace(path string, content []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	// A zero access time leaves it unchanged.
	return os.Chtimes(path, time.Time{}, info.ModTime())
}
