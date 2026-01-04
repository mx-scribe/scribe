package valueobjects

import "strings"

// Color represents a Tailwind CSS color for visual categorization.
type Color string

// ValidColors contains all valid Tailwind CSS 4 colors (22 colors).
var ValidColors = []string{
	"slate", "gray", "zinc", "neutral", "stone",
	"red", "orange", "amber", "yellow", "lime", "green",
	"emerald", "teal", "cyan", "sky", "blue", "indigo",
	"violet", "purple", "fuchsia", "pink", "rose",
}

// validColorsMap for O(1) lookup.
var validColorsMap = func() map[string]bool {
	m := make(map[string]bool, len(ValidColors))
	for _, c := range ValidColors {
		m[c] = true
	}
	return m
}()

// IsValid checks if the color is a valid Tailwind CSS color.
func (c Color) IsValid() bool {
	return validColorsMap[strings.ToLower(string(c))]
}

// String returns the string representation of the color.
func (c Color) String() string {
	return string(c)
}

// ColorFromString creates a Color from a string, returns empty if invalid.
func ColorFromString(s string) Color {
	color := Color(strings.ToLower(s))
	if color.IsValid() {
		return color
	}
	return ""
}

// DefaultColor returns the default color.
func DefaultColor() Color {
	return "slate"
}

// AutoAssignColor automatically assigns a color based on severity.
func AutoAssignColor(severity Severity) Color {
	switch severity {
	case SeverityCritical, SeverityError:
		return "red"
	case SeverityWarning:
		return "yellow"
	case SeveritySuccess:
		return "green"
	case SeverityInfo:
		return "blue"
	case SeverityDebug:
		return "gray"
	default:
		return "slate"
	}
}
