package faker

import (
	"context"
	"testing"
	"time"
)

func TestGenerator_CategoryDistribution(t *testing.T) {
	g := NewGenerator(12345, false)

	counts := make(map[string]int)
	total := 10000

	for i := 0; i < total; i++ {
		log := g.Generate()
		source := log.Header.Source
		if source == "" {
			source = "chaos"
		}

		// Map sources to categories
		switch source {
		case "nginx":
			counts["http"]++
		case "auth-service", "payment-service", "notification-service", "job-worker", "api-server", "frontend", "worker", "backend":
			counts["application"]++
		case "postgresql":
			counts["database"]++
		case "security":
			counts["security"]++
		case "docker", "system", "monitor":
			counts["system"]++
		case "deploy":
			counts["business"]++
		default:
			counts["chaos"]++
		}
	}

	// Check distribution is roughly correct (with 5% tolerance)
	expected := map[string]int{
		"http":        WeightHTTP * total / 100,
		"application": WeightApplication * total / 100,
		"database":    WeightDatabase * total / 100,
		"security":    WeightSecurity * total / 100,
		"system":      WeightSystem * total / 100,
		"business":    WeightBusiness * total / 100,
		"chaos":       WeightChaos * total / 100,
	}

	tolerance := 0.05 * float64(total)
	for cat, exp := range expected {
		actual := counts[cat]
		if float64(actual) < float64(exp)-tolerance || float64(actual) > float64(exp)+tolerance {
			t.Errorf("Category %s: expected ~%d, got %d", cat, exp, actual)
		}
	}
}

func TestGenerator_SeededReproducibility(t *testing.T) {
	g1 := NewGenerator(42, false)
	g2 := NewGenerator(42, false)

	for i := 0; i < 100; i++ {
		log1 := g1.Generate()
		log2 := g2.Generate()

		if log1.Header.Title != log2.Header.Title {
			t.Errorf("Seeded generators produced different logs at iteration %d", i)
		}
	}
}

func TestGenerator_DifferentSeeds(t *testing.T) {
	g1 := NewGenerator(1, false)
	g2 := NewGenerator(2, false)

	// Generate several logs and check they're not all identical
	same := 0
	for i := 0; i < 100; i++ {
		log1 := g1.Generate()
		log2 := g2.Generate()
		if log1.Header.Title == log2.Header.Title {
			same++
		}
	}

	// Some might be the same by chance, but not most
	if same > 50 {
		t.Errorf("Different seeds produced too many identical logs: %d/100", same)
	}
}

func TestGenerator_HTTPLogs(t *testing.T) {
	g := NewGenerator(12345, false)

	for i := 0; i < 100; i++ {
		log := g.GenerateHTTP()

		if log.Header.Source != "nginx" {
			t.Errorf("HTTP log should have source 'nginx', got %s", log.Header.Source)
		}

		body, ok := log.Body.(map[string]any)
		if !ok {
			t.Error("HTTP log body should be a map")
			continue
		}

		// Check required fields
		requiredFields := []string{"method", "path", "status", "response_time_ms"}
		for _, field := range requiredFields {
			if _, exists := body[field]; !exists {
				t.Errorf("HTTP log missing field: %s", field)
			}
		}
	}
}

func TestGenerator_DatabaseLogs(t *testing.T) {
	g := NewGenerator(12345, false)

	for i := 0; i < 100; i++ {
		log := g.GenerateDatabase()

		if log.Header.Source != "postgresql" {
			t.Errorf("Database log should have source 'postgresql', got %s", log.Header.Source)
		}
	}
}

func TestGenerator_SecurityLogs(t *testing.T) {
	g := NewGenerator(12345, false)

	for i := 0; i < 100; i++ {
		log := g.GenerateSecurity()

		if log.Header.Source != "security" {
			t.Errorf("Security log should have source 'security', got %s", log.Header.Source)
		}
	}
}

func TestGenerator_ChaosMode(t *testing.T) {
	gNormal := NewGenerator(12345, false)
	gChaos := NewGenerator(12345, true)

	errorCountNormal := 0
	errorCountChaos := 0
	total := 1000

	for i := 0; i < total; i++ {
		logNormal := gNormal.Generate()
		logChaos := gChaos.Generate()

		if logNormal.Header.Severity == "error" || logNormal.Header.Severity == "critical" {
			errorCountNormal++
		}
		if logChaos.Header.Severity == "error" || logChaos.Header.Severity == "critical" {
			errorCountChaos++
		}
	}

	// Chaos mode should have significantly more errors
	if errorCountChaos <= errorCountNormal {
		t.Errorf("Chaos mode should produce more errors: normal=%d, chaos=%d", errorCountNormal, errorCountChaos)
	}
}

func TestGenerator_AllCategoriesValid(t *testing.T) {
	g := NewGenerator(12345, false)

	categories := []string{"http", "application", "database", "security", "system", "business", "chaos"}

	for _, cat := range categories {
		log := g.GenerateCategory(cat)
		if log.Header.Title == "" {
			t.Errorf("Category %s generated empty title", cat)
		}
	}
}

func TestFaker_DryRun(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DryRun = true
	cfg.Count = 10
	cfg.MinDelay = 1 * time.Millisecond
	cfg.MaxDelay = 2 * time.Millisecond
	cfg.Seed = 12345

	f := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := f.Run(ctx, nil)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Dry run should not fail: %v", err)
	}

	// In dry run, no actual requests are made but stats should be tracked
	if f.Stats().Sent.Load() != 10 {
		t.Errorf("Expected 10 logs sent in dry run, got %d", f.Stats().Sent.Load())
	}
}

func TestFaker_CountLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DryRun = true
	cfg.Count = 5
	cfg.MinDelay = 1 * time.Millisecond
	cfg.MaxDelay = 2 * time.Millisecond
	cfg.Seed = 12345

	f := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := f.Run(ctx, nil)
	if err != nil {
		t.Errorf("Count limited run should not fail: %v", err)
	}

	if f.Stats().Sent.Load() != 5 {
		t.Errorf("Expected exactly 5 logs, got %d", f.Stats().Sent.Load())
	}
}

func TestFaker_IntervalRange(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DryRun = true
	cfg.Count = 5
	cfg.MinDelay = 10 * time.Millisecond
	cfg.MaxDelay = 20 * time.Millisecond
	cfg.Seed = 12345

	f := New(cfg)

	var delays []time.Duration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := f.Run(ctx, func(log LogEntry, delay time.Duration, sendErr error) {
		delays = append(delays, delay)
	})
	if err != nil {
		t.Errorf("Run should not fail: %v", err)
	}

	for i, d := range delays {
		if d < cfg.MinDelay || d > cfg.MaxDelay {
			t.Errorf("Delay %d out of range: %v (expected %v-%v)", i, d, cfg.MinDelay, cfg.MaxDelay)
		}
	}
}

func TestStats_Percentile(t *testing.T) {
	s := &Stats{}

	// Add some latencies
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond, // outlier
	}

	for _, l := range latencies {
		s.AddLatency(l)
	}

	p50 := s.Percentile(50)
	p95 := s.Percentile(95)
	max := s.Max()

	if p50 < 20*time.Millisecond || p50 > 40*time.Millisecond {
		t.Errorf("P50 should be around 30ms, got %v", p50)
	}

	if p95 < 50*time.Millisecond {
		t.Errorf("P95 should be >= 50ms, got %v", p95)
	}

	if max != 100*time.Millisecond {
		t.Errorf("Max should be 100ms, got %v", max)
	}
}

func TestRandomHelpers(t *testing.T) {
	g := NewGenerator(12345, false)

	// Test randomIP
	ip := randomIP(g.rng)
	if ip == "" || len(ip) < 7 {
		t.Errorf("Invalid IP: %s", ip)
	}

	// Test randomID
	id := randomID(g.rng, "usr_")
	if len(id) != 12 { // prefix (4) + id (8)
		t.Errorf("Invalid ID length: %s", id)
	}

	// Test randomEmail
	email := randomEmail(g.rng)
	if email == "" || len(email) < 5 {
		t.Errorf("Invalid email: %s", email)
	}

	// Test randomDuration
	d := randomDuration(g.rng, 10, 100)
	if d < 10 || d >= 100 {
		t.Errorf("Duration out of range: %d", d)
	}
}
