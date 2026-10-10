// Package render draws tables and charts for the terminal from plain values.
package render

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// levelColors is a green scale similar to GitHub contributions, from level 0
// (not done) to level 4.
var levelColors = [5]lipgloss.AdaptiveColor{
	{Light: "#d0d7de", Dark: "#30363d"},
	{Light: "#9be9a8", Dark: "#0e4429"},
	{Light: "#40c463", Dark: "#006d32"},
	{Light: "#30a14e", Dark: "#26a641"},
	{Light: "#216e39", Dark: "#39d353"},
}

// plainLevels draws levels 0–4 without color.
var plainLevels = [5]string{"·", "░", "▒", "▓", "█"}

// Style decides whether output is colored and holds the palette.
type Style struct {
	Enabled bool
}

// DetectStyle enables color when stdout is a terminal and NO_COLOR is unset.
func DetectStyle() Style {
	fd := os.Stdout.Fd()
	tty := isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
	return Style{Enabled: tty && os.Getenv("NO_COLOR") == ""}
}

// Level draws one heatmap cell. Levels outside 0–4 are blank.
func (s Style) Level(level int) string {
	if level < 0 || level > 4 {
		return " "
	}
	if !s.Enabled {
		return plainLevels[level]
	}
	return lipgloss.NewStyle().Foreground(levelColors[level]).Render("■")
}

// Shade gives text a background whose color gets stronger with the level
// (0–4). Without color, the text is returned unchanged.
func (s Style) Shade(text string, level int) string {
	if !s.Enabled || level < 0 || level > 4 {
		return text
	}
	fg := lipgloss.AdaptiveColor{Light: "#24292f", Dark: "#e6edf3"}
	if level >= 3 {
		fg = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"}
	}
	return lipgloss.NewStyle().Background(levelColors[level]).Foreground(fg).Render(text)
}

// Bold renders text in bold when color is enabled.
func (s Style) Bold(text string) string {
	if !s.Enabled {
		return text
	}
	return lipgloss.NewStyle().Bold(true).Render(text)
}

// Faint renders text dimmed when color is enabled.
func (s Style) Faint(text string) string {
	if !s.Enabled {
		return text
	}
	return lipgloss.NewStyle().Faint(true).Render(text)
}
