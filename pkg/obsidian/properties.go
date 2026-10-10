package obsidian

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type PropertyType int

const (
	Text PropertyType = iota
	Numeric
	Boolean
)

type Property struct {
	Name    string
	Content string
	Value   int
	Unit    string
	Type    PropertyType
}

// GetProperties parses the YAML frontmatter of a note into properties, in the
// order they are declared.
func GetProperties(text string) ([]Property, error) {
	frontmatter, err := GetFrontmatter(text)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(frontmatter) == "" {
		return nil, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		return nil, fmt.Errorf("invalid frontmatter: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("invalid frontmatter: expected key/value pairs")
	}

	var properties []Property
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		name := strings.TrimSpace(mapping.Content[i].Value)
		properties = append(properties, parsePropertyValue(name, nodeContent(mapping.Content[i+1])))
	}

	return properties, nil
}

// nodeContent flattens a YAML value into the string form used by properties.
// Lists are joined with ", "; nested mappings are not supported and yield "".
func nodeContent(n *yaml.Node) string {
	switch n.Kind {
	case yaml.ScalarNode:
		if n.Tag == "!!null" {
			return ""
		}
		return n.Value
	case yaml.SequenceNode:
		var items []string
		for _, item := range n.Content {
			if c := nodeContent(item); c != "" {
				items = append(items, c)
			}
		}
		return strings.Join(items, ", ")
	case yaml.AliasNode:
		return nodeContent(n.Alias)
	}
	return ""
}

// GetFrontmatter returns the YAML between the opening "---" line and the next
// line consisting only of "---" (or "..."). Text without frontmatter returns "".
func GetFrontmatter(text string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if strings.TrimRight(lines[0], " \t") != "---" {
		return "", nil
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "---" || line == "..." {
			return strings.Join(lines[1:i], "\n"), nil
		}
	}
	return "", fmt.Errorf("error: closing frontmatter delimiter not found")
}

// durationPart matches one "<number> <unit>" part at the start of a duration,
// such as "1.5 hours" or "30m".
var durationPart = regexp.MustCompile(`^(\d+(?:\.\d+)?)(\s*)([a-z]+)\s*`)

// durationUnits are the seconds in each duration unit. Single letters only
// count when attached to the number ("30m"), so "500 m" stays in meters.
var durationUnits = map[string]float64{
	"hour": 3600, "hours": 3600, "hr": 3600, "hrs": 3600, "h": 3600,
	"minute": 60, "minutes": 60, "min": 60, "mins": 60, "m": 60,
	"second": 1, "seconds": 1, "sec": 1, "secs": 1, "s": 1,
}

// parseDuration parses content made only of duration parts, such as
// "1 hour 30 minutes", "1.5 hours" or "1h30m", into seconds.
func parseDuration(content string) (int, bool) {
	var seconds float64
	for rest := content; rest != ""; {
		m := durationPart.FindStringSubmatch(rest)
		if m == nil {
			return 0, false
		}
		unit, ok := durationUnits[m[3]]
		if !ok || (len(m[3]) == 1 && m[2] != "") {
			return 0, false
		}
		n, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, false
		}
		seconds += n * unit
		rest = rest[len(m[0]):]
	}
	return int(math.Round(seconds)), true
}

func parsePropertyValue(name, content string) Property {
	content = strings.TrimSpace(content)
	if content == "" {
		return Property{Name: name}
	}

	content = strings.ToLower(content)
	if content == "true" || content == "false" {
		var value int
		if content == "true" {
			value = 1
		}
		return Property{Name: name, Content: content, Value: value, Type: Boolean}
	}

	if seconds, ok := parseDuration(content); ok {
		return Property{Name: name, Content: content, Value: seconds, Unit: "seconds", Type: Numeric}
	}

	parts := strings.Split(content, " ")
	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return Property{Name: name, Content: content, Type: Text}
	}
	var unit string
	if len(parts) > 1 {
		unit = parts[1]
	}
	return Property{Name: name, Content: content, Value: value, Unit: unit, Type: Numeric}
}
