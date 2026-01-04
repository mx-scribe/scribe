package entities

// Stats represents aggregate statistics for logs
type Stats struct {
	Total              int            `json:"total"`
	TotalLast24h       int            `json:"total_last_24h"`
	ByType             []TypeCount    `json:"by_type"`
	BySource           []SourceCount  `json:"by_source"`
	ByColor            []ColorCount   `json:"by_color"`
	ErrorRate24h       string         `json:"error_rate_24h"`
	SeverityBreakdown  map[string]int `json:"severity_breakdown"`
	TopSources         []SourceCount  `json:"top_sources"`
	HourlyDistribution []HourlyCount  `json:"hourly_distribution"`
	PatternStats       *PatternStats  `json:"pattern_stats,omitempty"`
	DetectionAccuracy  string         `json:"detection_accuracy"`
	Trends             *TrendStats    `json:"trends,omitempty"`
	Alerts             []string       `json:"alerts,omitempty"`
}

// PatternStats represents smart pattern detection statistics
type PatternStats struct {
	HTTPCodesDetected int `json:"http_codes_detected"`
	StackTracesFound  int `json:"stack_traces_found"`
	SecurityIssues    int `json:"security_issues"`
	PerformanceIssues int `json:"performance_issues"`
}

// TrendStats represents trend analysis
type TrendStats struct {
	ErrorsIncreasing bool    `json:"errors_increasing"`
	ErrorChange      float64 `json:"error_change"`
	SpikeDetected    bool    `json:"spike_detected"`
}

// HourlyCount represents hourly distribution
type HourlyCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// TypeCount represents aggregated type statistics
type TypeCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SourceCount represents aggregated source statistics
type SourceCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ColorCount represents aggregated color statistics
type ColorCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Health represents the health status of the service
type Health struct {
	Status string `json:"status"`
}

// NewHealthy creates a healthy health status
func NewHealthy() *Health {
	return &Health{Status: "ok"}
}
