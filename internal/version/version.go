package version

// Version is the current version of SCRIBE.
// This is updated by the release script.
const Version = "1.0.0"

// BuildDate is set at compile time.
var BuildDate = "unknown"

// GitCommit is set at compile time.
var GitCommit = "unknown"

// Info returns version information as a formatted string.
func Info() string {
	return Version
}

// Full returns full version information including build details.
func Full() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
