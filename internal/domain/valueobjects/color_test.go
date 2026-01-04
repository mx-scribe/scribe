package valueobjects

import "testing"

func TestColor_IsValid(t *testing.T) {
	tests := []struct {
		color Color
		want  bool
	}{
		{"red", true},
		{"blue", true},
		{"green", true},
		{"slate", true},
		{"RED", true},  // case insensitive
		{"Blue", true}, // case insensitive
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.color), func(t *testing.T) {
			if got := tt.color.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorFromString(t *testing.T) {
	tests := []struct {
		input string
		want  Color
	}{
		{"red", "red"},
		{"BLUE", "blue"},
		{"Green", "green"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ColorFromString(tt.input); got != tt.want {
				t.Errorf("ColorFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoAssignColor(t *testing.T) {
	tests := []struct {
		severity Severity
		want     Color
	}{
		{SeverityCritical, "red"},
		{SeverityError, "red"},
		{SeverityWarning, "yellow"},
		{SeveritySuccess, "green"},
		{SeverityInfo, "blue"},
		{SeverityDebug, "gray"},
		{Severity("custom"), "slate"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if got := AutoAssignColor(tt.severity); got != tt.want {
				t.Errorf("AutoAssignColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultColor(t *testing.T) {
	if got := DefaultColor(); got != "slate" {
		t.Errorf("DefaultColor() = %v, want slate", got)
	}
}

func TestColor_String(t *testing.T) {
	tests := []struct {
		color Color
		want  string
	}{
		{"red", "red"},
		{"blue", "blue"},
		{"slate", "slate"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.color.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColor_AllValidColors(t *testing.T) {
	// Test all valid Tailwind colors
	for _, colorStr := range ValidColors {
		t.Run(colorStr, func(t *testing.T) {
			color := Color(colorStr)
			if !color.IsValid() {
				t.Errorf("Color %q should be valid", colorStr)
			}
		})
	}
}
