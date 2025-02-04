package obsidian

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Property struct {
	Name    string
	Content string
	Value   int
	Unit    string
}

func GetProperties(text string) ([]Property, error) {
	text, _ = GetFrontmatter(text)

	lines := strings.Split(text, "\n")
	properties := make([]Property, len(lines))

	var err error
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		properties[i], err = ParseProperty(line)
		if err != nil {
			return properties, err
		}
	}

	return properties, nil
}

func GetFrontmatter(text string) (string, error) {
	if !strings.HasPrefix(text, "---") {
		return "", nil
	}
	parts := strings.Split(text, "---")
	if len(parts) > 1 {
		text = parts[1]
	} else {
		return "", fmt.Errorf("error: YAML delimiters not found\n")
	}
	lines := strings.Split(text, "\n")
	lines = lines[1 : len(lines)-1]
	text = strings.Join(lines, "\n")

	return text, nil
}

func ParseProperty(s string) (Property, error) {

	s = strings.TrimSpace(s)
	if s == "" {
		return Property{}, fmt.Errorf("empty string")
	}

	parts := strings.Split(s, ":")
	name := strings.TrimSpace(parts[0])

	var content string
	if len(parts) < 2 {
		return Property{Name: name}, nil
	}
	content = strings.TrimSpace(parts[1])

	if content == "" {
		return Property{Name: name}, nil
	}

	var value int
	if content == "true" || content == "false" {
		boolValue, err := strconv.ParseBool(content)
		if err != nil {
			return Property{}, fmt.Errorf("invalid format: %s", content)
		}
		if boolValue {
			value = 1
		}
		return Property{
			Name:    name,
			Content: content,
			Value:   value,
		}, nil
	}

	re := regexp.MustCompile(`(\d+)\s*(hour|minute|second)s?`)
	match := re.FindStringSubmatch(content)

	parts = strings.Split(content, " ")
	if match == nil {
		value, err := strconv.Atoi(parts[0])
		if err != nil {
			// return Property{}, fmt.Errorf("invalid format: %s", content)
			// Values can be just strings also with no numeric value.
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
		}, nil
	}

	var unit string
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

		unit = "seconds"
	}

	return Property{
		Name:    name,
		Content: content,
		Value:   int(value),
		Unit:    unit,
	}, nil
}
