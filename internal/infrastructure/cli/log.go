package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/application/commands"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

var (
	logSeverity    string
	logSource      string
	logColor       string
	logDescription string
	logBody        string
)

var logCmd = &cobra.Command{
	Use:   "log <title>",
	Short: "Send a log entry",
	Long:  `Send a log entry to the local SCRIBE database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Ensure database directory exists
		dbPath := GetDBPath()
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}

		// Connect to database
		db, err := sqlite.NewDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		// Run migrations
		if err := sqlite.RunMigrations(db.Conn()); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		// Parse body JSON if provided
		var body map[string]any
		if logBody != "" {
			if err := json.Unmarshal([]byte(logBody), &body); err != nil {
				return fmt.Errorf("invalid JSON body: %w", err)
			}
		}

		// Create handler and execute
		repo := sqlite.NewLogRepository(db)
		handler := commands.NewCreateLogHandler(repo)

		input := commands.CreateLogInput{
			Title:       title,
			Severity:    logSeverity,
			Source:      logSource,
			Color:       logColor,
			Description: logDescription,
			Body:        body,
		}

		output, err := handler.Handle(input)
		if err != nil {
			return fmt.Errorf("failed to create log: %w", err)
		}

		out := NewOutput()
		if GetOutputFormat() == "json" {
			return out.Print(output)
		}
		out.Success("Log created: #%d [%s] %s", output.ID, output.Severity, output.Title)
		return nil
	},
}

func init() {
	logCmd.Flags().StringVarP(&logSeverity, "severity", "s", "", "log severity (critical, error, warning, success, info, debug)")
	logCmd.Flags().StringVar(&logSource, "source", "", "log source (e.g., api, database, auth)")
	logCmd.Flags().StringVarP(&logColor, "color", "c", "", "log color (tailwind color name)")
	logCmd.Flags().StringVarP(&logDescription, "description", "d", "", "log description")
	logCmd.Flags().StringVarP(&logBody, "body", "b", "", "log body as JSON")

	rootCmd.AddCommand(logCmd)
}
