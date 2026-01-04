package valueobjects

// Severity represents the severity level of a log entry.
// Custom severities are allowed - the standard ones are provided for convenience.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityError    Severity = "error"
	SeverityWarning  Severity = "warning"
	SeveritySuccess  Severity = "success"
	SeverityInfo     Severity = "info"
	SeverityDebug    Severity = "debug"
)

// standardSeverities maps standard severities for quick lookup.
var standardSeverities = map[Severity]bool{
	SeverityCritical: true,
	SeverityError:    true,
	SeverityWarning:  true,
	SeveritySuccess:  true,
	SeverityInfo:     true,
	SeverityDebug:    true,
}

// IsValid checks if the severity is non-empty (all custom severities are valid).
func (s Severity) IsValid() bool {
	return s != ""
}

// IsStandard checks if the severity is one of the standard predefined severities.
func (s Severity) IsStandard() bool {
	return standardSeverities[s]
}

// String returns the string representation of the severity.
func (s Severity) String() string {
	return string(s)
}

// DefaultSeverity returns the default severity when none is specified.
func DefaultSeverity() Severity {
	return SeverityInfo
}

// SeverityFromString creates a Severity from a string.
// Returns the severity as-is (custom severities are allowed).
// Returns default only if empty.
func SeverityFromString(s string) Severity {
	if s == "" {
		return DefaultSeverity()
	}
	return Severity(s)
}
