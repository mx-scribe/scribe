package rules

// PerformanceThresholds defines performance thresholds in milliseconds.
var PerformanceThresholds = map[string]int{
	"fast":     100,
	"normal":   1000,
	"slow":     3000,
	"critical": 5000,
}

// CategorizePerformance categorizes duration into performance level.
func CategorizePerformance(durationMs int) string {
	if durationMs < PerformanceThresholds["fast"] {
		return "fast"
	} else if durationMs < PerformanceThresholds["normal"] {
		return "normal"
	} else if durationMs < PerformanceThresholds["slow"] {
		return "slow"
	}
	return "critical"
}

// PerformanceSeverity maps performance level to severity.
func PerformanceSeverity(level string) string {
	switch level {
	case "fast":
		return "success"
	case "normal":
		return "info"
	case "slow":
		return "warning"
	case "critical":
		return "error"
	default:
		return "info"
	}
}

// PerformancePatterns maps performance-related patterns to severity.
var PerformancePatterns = map[string]string{
	"slow query":         "warning",
	"query timeout":      "error",
	"request timeout":    "error",
	"connection timeout": "error",
	"high latency":       "warning",
	"memory pressure":    "warning",
	"cpu spike":          "warning",
	"disk full":          "critical",
	"rate limited":       "warning",
	"throttled":          "warning",
	"bottleneck":         "warning",
	"degraded":           "warning",
}
