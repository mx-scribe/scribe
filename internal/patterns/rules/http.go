package rules

// HTTPStatusSeverity maps HTTP status codes to severity levels.
var HTTPStatusSeverity = map[string]string{
	// 2xx Success
	"200": "success", "201": "success", "202": "success", "204": "success",
	// 3xx Redirect
	"301": "info", "302": "info", "304": "info",
	// 4xx Client errors
	"400": "warning", "401": "error", "403": "error", "404": "warning",
	"405": "warning", "408": "warning", "409": "warning", "410": "warning",
	"413": "warning", "422": "warning", "429": "warning",
	// 5xx Server errors
	"500": "error", "501": "error", "502": "error", "503": "critical",
	"504": "critical", "507": "critical", "511": "error",
}
