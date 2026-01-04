package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	dbPath       string
	configPath   string
	outputFormat string
	noColor      bool
	verbose      bool
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "scribe",
	Short: "Smart logging for humans",
	Long: `SCRIBE - Smart logging for humans. Single binary. Zero dependencies.

A self-hosted logging solution that proves:
  • Simple deployment doesn't require messy code
  • Smart features don't require complex infrastructure
  • Beautiful UIs don't require JavaScript frameworks

Configuration:
  Config files are loaded from (in order of priority):
    1. --config flag
    2. ./scribe.json
    3. ./.scribe.json
    4. ~/.scribe/config.json
    5. ~/.config/scribe/config.json
    6. /etc/scribe/config.json

  Environment variables (override config file):
    SCRIBE_PORT             Server port
    SCRIBE_HOST             Server host
    SCRIBE_DB_PATH          Database file path
    SCRIBE_RETENTION_DAYS   Log retention in days
    SCRIBE_DEFAULT_SEVERITY Default log severity
    SCRIBE_DEFAULT_SOURCE   Default log source
    SCRIBE_OUTPUT_FORMAT    Output format (table, json, plain)
    SCRIBE_NO_COLOR         Disable colors (true/1)
    SCRIBE_VERBOSE          Verbose output (true/1)`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		config, err := LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Apply config values if flags weren't set explicitly
		if !cmd.Flags().Changed("db") && config.Database.Path != "" {
			dbPath = config.Database.Path
		}
		if !cmd.Flags().Changed("format") && config.Output.Format != "" {
			outputFormat = config.Output.Format
		}
		if !cmd.Flags().Changed("no-color") {
			noColor = config.Output.NoColor
		}
		if !cmd.Flags().Changed("verbose") {
			verbose = config.Output.Verbose
		}

		// Set global config
		SetConfig(config)

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Set default database path
	homeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(homeDir, ".scribe", "scribe.db")

	// Database and config
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "database file path")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")

	// Output options
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "table", "output format (table, json, plain)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// GetDBPath returns the database path from flags.
func GetDBPath() string {
	return dbPath
}

// GetConfigPath returns the config path from flags.
func GetConfigPath() string {
	return configPath
}

// GetOutputFormat returns the output format from flags.
func GetOutputFormat() string {
	return outputFormat
}

// IsNoColor returns whether color output is disabled.
func IsNoColor() bool {
	return noColor
}

// IsVerbose returns whether verbose output is enabled.
func IsVerbose() bool {
	return verbose
}
