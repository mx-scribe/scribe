package faker

import "time"

// Config holds the configuration for the faker.
type Config struct {
	// Connection
	Endpoint string

	// Realistic mode
	MinDelay time.Duration
	MaxDelay time.Duration

	// Limits
	Duration time.Duration
	Count    int

	// Modes
	Chaos   bool
	Stress  bool
	DryRun  bool
	Quiet   bool
	Verbose bool

	// Stress mode
	StressRate int

	// Reproducibility
	Seed int64

	// Filtering
	Categories []string
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Endpoint:   "http://localhost:8080",
		MinDelay:   3 * time.Second,
		MaxDelay:   30 * time.Second,
		Duration:   0, // infinite
		Count:      0, // infinite
		Chaos:      false,
		Stress:     false,
		StressRate: 100,
		DryRun:     false,
		Seed:       0, // random
		Categories: nil,
		Quiet:      false,
		Verbose:    false,
	}
}

// Category distribution weights (must sum to 100).
const (
	WeightHTTP        = 25
	WeightApplication = 25
	WeightDatabase    = 15
	WeightSecurity    = 10
	WeightSystem      = 10
	WeightBusiness    = 10
	WeightChaos       = 5
)

// Severity distribution weights for normal mode (must sum to 100).
var SeverityWeightsNormal = map[string]int{
	"debug":    10,
	"info":     45,
	"success":  15,
	"warning":  15,
	"error":    12,
	"critical": 3,
}

// Severity distribution weights for chaos mode (must sum to 100).
var SeverityWeightsChaos = map[string]int{
	"debug":    5,
	"info":     20,
	"success":  10,
	"warning":  25,
	"error":    30,
	"critical": 10,
}
