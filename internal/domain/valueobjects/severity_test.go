package valueobjects

import "testing"

func TestSeverity_IsValid(t *testing.T) {
	tests := []struct {
		severity Severity
		want     bool
	}{
		{SeverityInfo, true},
		{SeverityError, true},
		{Severity("custom"), true},
		{Severity("p1"), true},
		{Severity(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if got := tt.severity.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverity_IsStandard(t *testing.T) {
	tests := []struct {
		severity Severity
		want     bool
	}{
		{SeverityInfo, true},
		{SeverityError, true},
		{SeverityCritical, true},
		{Severity("custom"), false},
		{Severity("p1"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if got := tt.severity.IsStandard(); got != tt.want {
				t.Errorf("IsStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverityFromString(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{"error", SeverityError},
		{"info", SeverityInfo},
		{"custom", Severity("custom")},
		{"", SeverityInfo}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SeverityFromString(tt.input); got != tt.want {
				t.Errorf("SeverityFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSeverity(t *testing.T) {
	if got := DefaultSeverity(); got != SeverityInfo {
		t.Errorf("DefaultSeverity() = %v, want %v", got, SeverityInfo)
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityInfo, "info"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
		{SeverityWarning, "warning"},
		{SeveritySuccess, "success"},
		{SeverityDebug, "debug"},
		{Severity("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
