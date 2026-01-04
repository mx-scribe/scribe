package faker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Stats tracks faker statistics.
type Stats struct {
	Sent      atomic.Int64
	Errors    atomic.Int64
	StartTime time.Time
	mu        sync.Mutex
	latencies []time.Duration
}

// AddLatency records a request latency.
func (s *Stats) AddLatency(d time.Duration) {
	s.mu.Lock()
	s.latencies = append(s.latencies, d)
	s.mu.Unlock()
}

// Percentile returns the nth percentile latency.
func (s *Stats) Percentile(n int) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(s.latencies))
	copy(sorted, s.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	idx := len(sorted) * n / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// Max returns the maximum latency.
func (s *Stats) Max() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	var max time.Duration
	for _, d := range s.latencies {
		if d > max {
			max = d
		}
	}
	return max
}

// Rate returns logs per second.
func (s *Stats) Rate() float64 {
	elapsed := time.Since(s.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(s.Sent.Load()) / elapsed
}

// Faker generates and sends fake logs.
type Faker struct {
	config    Config
	client    *http.Client
	generator *Generator
	stats     *Stats
}

// New creates a new Faker.
func New(cfg Config) *Faker {
	return &Faker{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		generator: NewGenerator(cfg.Seed, cfg.Chaos),
		stats:     &Stats{StartTime: time.Now()},
	}
}

// Stats returns the current statistics.
func (f *Faker) Stats() *Stats {
	return f.stats
}

// Run executes the faker in realistic mode.
func (f *Faker) Run(ctx context.Context, onLog func(LogEntry, time.Duration, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check count limit
		if f.config.Count > 0 && f.stats.Sent.Load() >= int64(f.config.Count) {
			return nil
		}

		// Generate and send log
		log := f.generateLog()
		err := f.sendLog(log)

		if err != nil {
			f.stats.Errors.Add(1)
		} else {
			f.stats.Sent.Add(1)
		}

		// Calculate next delay
		delay := f.randomDelay()

		if onLog != nil {
			onLog(log, delay, err)
		}

		// Wait for next interval
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

// RunStress executes the faker in stress test mode.
func (f *Faker) RunStress(ctx context.Context, onProgress func(sent, errors int64, rate float64, p95 time.Duration)) error {
	interval := time.Second / time.Duration(f.config.StressRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 100) // limit concurrent requests

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()

		case <-progressTicker.C:
			if onProgress != nil {
				onProgress(f.stats.Sent.Load(), f.stats.Errors.Load(), f.stats.Rate(), f.stats.Percentile(95))
			}

		case <-ticker.C:
			// Check count limit
			if f.config.Count > 0 && f.stats.Sent.Load() >= int64(f.config.Count) {
				wg.Wait()
				return nil
			}

			semaphore <- struct{}{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()

				log := f.generateLog()
				start := time.Now()
				err := f.sendLog(log)
				latency := time.Since(start)

				f.stats.AddLatency(latency)

				if err != nil {
					f.stats.Errors.Add(1)
				} else {
					f.stats.Sent.Add(1)
				}
			}()
		}
	}
}

// generateLog creates a log entry based on configuration.
func (f *Faker) generateLog() LogEntry {
	if len(f.config.Categories) > 0 {
		// Pick random from allowed categories
		cat := f.config.Categories[f.generator.rng.IntN(len(f.config.Categories))]
		return f.generator.GenerateCategory(cat)
	}
	return f.generator.Generate()
}

// sendLog sends a log to the API endpoint.
func (f *Faker) sendLog(log LogEntry) error {
	if f.config.DryRun {
		return nil
	}

	body, err := json.Marshal(log)
	if err != nil {
		return err
	}

	url := f.config.Endpoint + "/api/logs"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Drain body to reuse connection
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// randomDelay returns a random delay between min and max.
func (f *Faker) randomDelay() time.Duration {
	min := f.config.MinDelay
	max := f.config.MaxDelay

	if min >= max {
		return min
	}

	delta := max - min
	randomNs := f.generator.rng.Int64N(int64(delta))
	return min + time.Duration(randomNs)
}
