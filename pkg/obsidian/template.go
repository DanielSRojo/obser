package obsidian

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	templaterTag = regexp.MustCompile(`(?s)<%(.*?)%>`)

	// momentExpr matches the subset of Templater used for daily notes:
	// [name =] moment(tp.file.title, '<fmt>')[.add|.subtract(N, '<unit>')]*.format('<fmt>')
	momentExpr  = regexp.MustCompile(`^(?:\w+\s*=\s*)?moment\(\s*tp\.file\.title\s*,\s*'[^']*'\s*\)((?:\.(?:add|subtract)\(\s*-?\d+\s*,\s*'\w+'\s*\))*)\.format\(\s*'([^']*)'\s*\)$`)
	momentShift = regexp.MustCompile(`\.(add|subtract)\(\s*(-?\d+)\s*,\s*'(\w+)'\s*\)`)

	momentTokens = strings.NewReplacer(
		"YYYY", "2006",
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"dddd", "Monday",
		"ddd", "Mon",
		"DD", "02",
	)
)

// RenderTemplate renders the Templater tags of a daily note template for the
// note of the given date. Only date expressions based on the note title are
// supported; any other tag is reported as an error.
func RenderTemplate(template string, date time.Time) (string, error) {
	var renderErr error
	rendered := templaterTag.ReplaceAllStringFunc(template, func(tag string) string {
		if renderErr != nil {
			return tag
		}
		expr := strings.TrimSpace(templaterTag.FindStringSubmatch(tag)[1])
		value, err := evalMoment(expr, date)
		if err != nil {
			renderErr = err
			return tag
		}
		return value
	})
	if renderErr != nil {
		return "", renderErr
	}
	return rendered, nil
}

func evalMoment(expr string, date time.Time) (string, error) {
	m := momentExpr.FindStringSubmatch(expr)
	if m == nil {
		return "", fmt.Errorf("unsupported template expression: %s", expr)
	}

	for _, shift := range momentShift.FindAllStringSubmatch(m[1], -1) {
		n, err := strconv.Atoi(shift[2])
		if err != nil {
			return "", fmt.Errorf("invalid amount in %s: %w", expr, err)
		}
		if shift[1] == "subtract" {
			n = -n
		}
		switch shift[3] {
		case "d", "day", "days":
			date = date.AddDate(0, 0, n)
		case "w", "week", "weeks":
			date = date.AddDate(0, 0, 7*n)
		case "M", "month", "months":
			date = date.AddDate(0, n, 0)
		case "y", "year", "years":
			date = date.AddDate(n, 0, 0)
		default:
			return "", fmt.Errorf("unsupported unit %q in %s", shift[3], expr)
		}
	}

	return date.Format(momentTokens.Replace(m[2])), nil
}
