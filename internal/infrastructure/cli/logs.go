package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

var (
	logsLimit    int
	logsOffset   int
	logsSeverity string
	logsSource   string
	logsSearch   string
	logsFormat   string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "List log entries",
	Long:  `List log entries from the local SCRIBE database with optional filters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to database
		db, err := sqlite.NewDatabase(GetDBPath())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		// Run migrations (ensures table exists)
		if err := sqlite.RunMigrations(db.Conn()); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		// Query logs
		repo := sqlite.NewLogRepository(db)
		filters := sqlite.LogFilters{
			Limit:    logsLimit,
			Offset:   logsOffset,
			Severity: logsSeverity,
			Source:   logsSource,
			Search:   logsSearch,
		}

		logs, total, err := repo.FindAll(filters)
		if err != nil {
			return fmt.Errorf("failed to query logs: %w", err)
		}

		if len(logs) == 0 {
			if logsFormat == "json" {
				fmt.Println("[]")
			} else {
				fmt.Println("No logs found.")
			}
			return nil
		}

		// Output based on format
		switch logsFormat {
		case "json":
			return outputLogsJSON(logs)
		case "csv":
			return outputLogsCSV(logs)
		default:
			return outputLogsTable(logs, total)
		}
	},
}

func init() {
	logsCmd.Flags().IntVarP(&logsLimit, "limit", "l", 20, "maximum number of logs to show")
	logsCmd.Flags().IntVarP(&logsOffset, "offset", "o", 0, "number of logs to skip")
	logsCmd.Flags().StringVarP(&logsSeverity, "severity", "s", "", "filter by severity")
	logsCmd.Flags().StringVar(&logsSource, "source", "", "filter by source")
	logsCmd.Flags().StringVar(&logsSearch, "search", "", "search in title and body")
	logsCmd.Flags().StringVarP(&logsFormat, "format", "f", "table", "output format (table, json, csv)")

	rootCmd.AddCommand(logsCmd)
}

//nolint:unparam // error return for consistency with outputLogsJSON/CSV
func outputLogsTable(logs []*entities.Log, total int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tSEVERITY\tSOURCE\tTITLE\tCREATED")
	_, _ = fmt.Fprintln(w, "--\t--------\t------\t-----\t-------")

	for _, log := range logs {
		source := log.Header.Source
		if source == "" {
			source = "-"
		}
		created := log.CreatedAt.Format("2006-01-02 15:04:05")
		title := log.Header.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			log.ID,
			log.EffectiveSeverity(),
			source,
			title,
			created,
		)
	}
	_ = w.Flush()

	// Show pagination info
	showing := len(logs)
	if logsOffset > 0 || total > showing {
		fmt.Printf("\nShowing %d-%d of %d logs\n", logsOffset+1, logsOffset+showing, total)
	}

	return nil
}

func outputLogsJSON(logs []*entities.Log) error {
	output, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func outputLogsCSV(logs []*entities.Log) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"id", "severity", "source", "title", "description", "created_at"}); err != nil {
		return err
	}

	// Rows
	for _, log := range logs {
		source := log.Header.Source
		if source == "" {
			source = ""
		}
		row := []string{
			strconv.FormatInt(log.ID, 10),
			string(log.EffectiveSeverity()),
			source,
			log.Header.Title,
			log.Header.Description,
			log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}
