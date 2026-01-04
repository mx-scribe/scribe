package rules

// DatabasePatterns maps database error patterns to severity levels.
var DatabasePatterns = map[string]string{
	"deadlock":                  "critical",
	"connection pool exhausted": "critical",
	"too many connections":      "critical",
	"duplicate key":             "warning",
	"constraint violation":      "error",
	"foreign key violation":     "error",
	"unique constraint":         "warning",
	"table locked":              "warning",
	"database locked":           "warning",
	"sqlite_busy":               "warning",
	"sqlite_locked":             "warning",
	"sqlite_corrupt":            "critical",
	"connection refused":        "error",
	"connection timeout":        "error",
	"query timeout":             "warning",
}
