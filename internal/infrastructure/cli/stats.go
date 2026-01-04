package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/application/queries"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show log statistics",
	Long:  `Show statistics about logs in the local SCRIBE database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to database
		db, err := sqlite.NewDatabase(GetDBPath())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		// Run migrations
		if err := sqlite.RunMigrations(db.Conn()); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		// Get stats
		repo := sqlite.NewLogRepository(db)
		handler := queries.NewGetStatsHandler(repo)

		stats, err := handler.Handle()
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		// Print stats
		fmt.Println("=== SCRIBE Statistics ===")
		fmt.Println()
		fmt.Printf("Total logs:     %d\n", stats.Total)
		fmt.Printf("Last 24 hours:  %d\n", stats.Last24Hours)

		if len(stats.BySeverity) > 0 {
			fmt.Println("\nBy Severity:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			// Sort severities for consistent output
			severities := make([]string, 0, len(stats.BySeverity))
			for s := range stats.BySeverity {
				severities = append(severities, s)
			}
			sort.Strings(severities)
			for _, s := range severities {
				fmt.Fprintf(w, "  %s:\t%d\n", s, stats.BySeverity[s])
			}
			w.Flush()
		}

		if len(stats.BySource) > 0 {
			fmt.Println("\nBy Source:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			// Sort sources for consistent output
			sources := make([]string, 0, len(stats.BySource))
			for s := range stats.BySource {
				sources = append(sources, s)
			}
			sort.Strings(sources)
			for _, s := range sources {
				fmt.Fprintf(w, "  %s:\t%d\n", s, stats.BySource[s])
			}
			w.Flush()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
