package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/faker"
)

var (
	fakerMinDelay   int
	fakerMaxDelay   int
	fakerDuration   int
	fakerCount      int
	fakerChaos      bool
	fakerStress     bool
	fakerRate       int
	fakerEndpoint   string
	fakerDryRun     bool
	fakerSeed       int64
	fakerCategories string
	fakerQuiet      bool
)

var fakerCmd = &cobra.Command{
	Use:   "faker",
	Short: "Generate fake logs for testing",
	Long: `Generate realistic fake logs to test SCRIBE dashboard,
SSE real-time updates, and smart pattern matching.

Examples:
  scribe faker                          # realistic mode, 3-30s intervals
  scribe faker --min 1 --max 5          # faster intervals (1-5 seconds)
  scribe faker --chaos                  # 50% error rate
  scribe faker --count 100              # stop after 100 logs
  scribe faker --stress --rate 500      # 500 logs/second
  scribe faker --dry-run                # print logs without sending
  scribe faker --categories http,database  # only specific categories

Categories: http, application, database, security, system, business, chaos`,
	RunE: runFaker,
}

func init() {
	fakerCmd.Flags().IntVar(&fakerMinDelay, "min", 3, "minimum interval in seconds")
	fakerCmd.Flags().IntVar(&fakerMaxDelay, "max", 30, "maximum interval in seconds")
	fakerCmd.Flags().IntVar(&fakerDuration, "duration", 0, "total duration in seconds (0 = infinite)")
	fakerCmd.Flags().IntVar(&fakerCount, "count", 0, "total logs to send (0 = infinite)")
	fakerCmd.Flags().BoolVar(&fakerChaos, "chaos", false, "50% error rate mode")
	fakerCmd.Flags().BoolVar(&fakerStress, "stress", false, "stress test mode")
	fakerCmd.Flags().IntVar(&fakerRate, "rate", 100, "logs per second (stress mode)")
	fakerCmd.Flags().StringVar(&fakerEndpoint, "endpoint", "http://localhost:8080", "SCRIBE API endpoint")
	fakerCmd.Flags().BoolVar(&fakerDryRun, "dry-run", false, "print logs without sending")
	fakerCmd.Flags().Int64Var(&fakerSeed, "seed", 0, "random seed for reproducibility (0 = random)")
	fakerCmd.Flags().StringVar(&fakerCategories, "categories", "", "comma-separated categories to generate")
	fakerCmd.Flags().BoolVarP(&fakerQuiet, "quiet", "q", false, "minimal output")

	rootCmd.AddCommand(fakerCmd)
}

func runFaker(cmd *cobra.Command, args []string) error {
	// Parse categories
	var categories []string
	if fakerCategories != "" {
		categories = strings.Split(fakerCategories, ",")
		for i, c := range categories {
			categories[i] = strings.TrimSpace(c)
		}
	}

	// Build config
	cfg := faker.Config{
		Endpoint:   fakerEndpoint,
		MinDelay:   time.Duration(fakerMinDelay) * time.Second,
		MaxDelay:   time.Duration(fakerMaxDelay) * time.Second,
		Duration:   time.Duration(fakerDuration) * time.Second,
		Count:      fakerCount,
		Chaos:      fakerChaos,
		Stress:     fakerStress,
		StressRate: fakerRate,
		DryRun:     fakerDryRun,
		Seed:       fakerSeed,
		Categories: categories,
		Quiet:      fakerQuiet,
		Verbose:    IsVerbose(),
	}

	// Create faker
	f := faker.New(cfg)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle duration limit
	if cfg.Duration > 0 {
		var durationCancel context.CancelFunc
		ctx, durationCancel = context.WithTimeout(ctx, cfg.Duration)
		defer durationCancel()
	}

	// Handle SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println() // newline after ^C
		cancel()
	}()

	// Run appropriate mode
	if cfg.Stress {
		return runStressMode(ctx, f, cfg)
	}
	return runRealisticMode(ctx, f, cfg)
}

func runRealisticMode(ctx context.Context, f *faker.Faker, cfg faker.Config) error {
	// Print header
	if !cfg.Quiet {
		mode := "realistic"
		if cfg.DryRun {
			mode = "DRY RUN"
		}
		if cfg.Chaos {
			mode = "chaos"
		}

		fmt.Println()
		fmt.Println("🎭 SCRIBE Faker starting...")
		fmt.Printf("   Endpoint:  %s\n", cfg.Endpoint)
		fmt.Printf("   Interval:  %ds - %ds\n", int(cfg.MinDelay.Seconds()), int(cfg.MaxDelay.Seconds()))
		fmt.Printf("   Mode:      %s\n", mode)
		if cfg.Count > 0 {
			fmt.Printf("   Limit:     %d logs\n", cfg.Count)
		}
		if cfg.Duration > 0 {
			fmt.Printf("   Duration:  %ds\n", int(cfg.Duration.Seconds()))
		}
		fmt.Println()
	}

	// Run
	err := f.Run(ctx, func(log faker.LogEntry, nextDelay time.Duration, sendErr error) {
		if cfg.DryRun && !cfg.Quiet {
			// Print full JSON in dry-run mode
			data, _ := json.MarshalIndent(log, "", "  ")
			fmt.Println(string(data))
			fmt.Println()
			return
		}

		if cfg.Quiet {
			return
		}

		// Print log line
		timestamp := time.Now().Format("15:04:05")
		status := "→"
		if sendErr != nil {
			status = "✗"
		}

		severity := ""
		if log.Header.Severity != "" {
			severity = fmt.Sprintf(" (%s)", log.Header.Severity)
		}

		source := log.Header.Source
		if source == "" {
			source = "unknown"
		}

		title := log.Header.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		fmt.Printf("[%s] %s POST %s %q%s [wait %.1fs]\n",
			timestamp, status, source, title, severity, nextDelay.Seconds())
	})

	// Print summary
	stats := f.Stats()
	if !cfg.Quiet {
		fmt.Println()
		fmt.Println("📊 Summary:")
		fmt.Printf("   Duration:  %s\n", time.Since(stats.StartTime).Truncate(time.Second))
		fmt.Printf("   Sent:      %d logs\n", stats.Sent.Load())
		fmt.Printf("   Errors:    %d failed requests\n", stats.Errors.Load())
		fmt.Printf("   Rate:      %.2f logs/s average\n", stats.Rate())
	}

	if err == context.Canceled || err == context.DeadlineExceeded {
		return nil
	}
	return err
}

func runStressMode(ctx context.Context, f *faker.Faker, cfg faker.Config) error {
	// Print header
	if !cfg.Quiet {
		fmt.Println()
		fmt.Println("🔥 SCRIBE Faker STRESS TEST")
		fmt.Printf("   Endpoint:  %s\n", cfg.Endpoint)
		fmt.Printf("   Rate:      %d logs/s\n", cfg.StressRate)
		if cfg.Duration > 0 {
			fmt.Printf("   Duration:  %ds\n", int(cfg.Duration.Seconds()))
			fmt.Printf("   Target:    %d logs\n", cfg.StressRate*int(cfg.Duration.Seconds()))
		}
		if cfg.Count > 0 {
			fmt.Printf("   Target:    %d logs\n", cfg.Count)
		}
		fmt.Println()
	}

	// Run
	err := f.RunStress(ctx, func(sent, errors int64, rate float64, p95 time.Duration) {
		if cfg.Quiet {
			return
		}

		// Progress line (overwrite previous)
		elapsed := time.Since(f.Stats().StartTime).Truncate(time.Second)
		fmt.Printf("\r[%s] sent: %d | %.0f/s | errors: %d | p95: %s    ",
			elapsed, sent, rate, errors, p95.Truncate(time.Millisecond))
	})

	// Print final summary
	stats := f.Stats()
	if !cfg.Quiet {
		fmt.Println() // newline after progress
		fmt.Println()
		fmt.Println("📊 Results:")
		fmt.Printf("   Duration:    %s\n", time.Since(stats.StartTime).Truncate(time.Second))
		fmt.Printf("   Total sent:  %d logs\n", stats.Sent.Load())

		total := stats.Sent.Load() + stats.Errors.Load()
		if total > 0 {
			successRate := float64(stats.Sent.Load()) / float64(total) * 100
			fmt.Printf("   Success:     %d (%.1f%%)\n", stats.Sent.Load(), successRate)
			fmt.Printf("   Failed:      %d (%.1f%%)\n", stats.Errors.Load(), 100-successRate)
		}

		fmt.Printf("   Rate:        %.1f logs/s average\n", stats.Rate())
		fmt.Println("   Latency:")
		fmt.Printf("     p50:  %s\n", stats.Percentile(50).Truncate(time.Millisecond))
		fmt.Printf("     p95:  %s\n", stats.Percentile(95).Truncate(time.Millisecond))
		fmt.Printf("     p99:  %s\n", stats.Percentile(99).Truncate(time.Millisecond))
		fmt.Printf("     max:  %s\n", stats.Max().Truncate(time.Millisecond))
	}

	if err == context.Canceled || err == context.DeadlineExceeded {
		return nil
	}
	return err
}
