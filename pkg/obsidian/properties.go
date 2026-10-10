package obsidian

import (
	"fmt"
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
		property, err := parsePropertyValue(name, nodeContent(mapping.Content[i+1]))
		if err != nil {
			return nil, err
		}
		properties = append(properties, property)
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

func ParseProperty(s string) (Property, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Property{}, fmt.Errorf("empty string")
	}

	name, content, _ := strings.Cut(s, ":")
	return parsePropertyValue(strings.TrimSpace(name), content)
}

func parsePropertyValue(name, content string) (Property, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return Property{Name: name}, nil
	}

	var value int
	content = strings.ToLower(content)
	if content == "true" || content == "false" {
		contentBool, err := strconv.ParseBool(content)
		if err != nil {
			return Property{}, fmt.Errorf("invalid format: %s", content)
		}
		if contentBool {
			value = 1
		}
		return Property{
			Name:    name,
			Content: content,
			Value:   value,
			Type:    Boolean,
		}, nil
	}

	re := regexp.MustCompile(`(\d+)\s*(hour|minute|second)s?`)
	match := re.FindStringSubmatch(content)

	parts := strings.Split(content, " ")
	if match == nil {
		value, err := strconv.Atoi(parts[0])
		if err != nil {
			return Property{
				Name:    name,
				Content: content,
				Value:   value,
				Type:    Text,
			}, nil
		}
		var unit string
		if len(parts) > 1 {
			unit = parts[1]
		}
		return Property{
			Name:    name,
			Content: content,
			Value:   value,
			Unit:    unit,
			Type:    Numeric,
		}, nil
	}

	// TODO: write function for converting value to seconds
	if len(match) == 3 {
		var err error
		value, err = strconv.Atoi(match[1])
		if err != nil {
			return Property{}, fmt.Errorf("invalid number: %s", match[1])
		}

		switch match[2] {
		case "hour":
			value = value * 3600
		case "minute":
			value = value * 60
		case "second":
			// Do nothing
		default:
			return Property{}, fmt.Errorf("unknown unit: %s", match[2])
		}

	}

	return Property{
		Name:    name,
		Content: content,
		Value:   int(value),
		Unit:    "seconds",
		Type:    Numeric,
	}, nil
}

func NormalizeProperties(properties []Property) ([]Property, error) {
	for i, p := range properties {
		if p.Type != Numeric {
			continue
		}
		if p.Unit == "" {
			continue
		}
		re := regexp.MustCompile(`(\d+)\s*(hour|minute|second)s?`)
		match := re.FindStringSubmatch(p.Unit)
		if match == nil {
			properties[i].Unit = ""
			continue
		}

	}
	return properties, nil
}
