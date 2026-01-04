package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/infrastructure/http"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
	"github.com/mx-scribe/scribe/web"
)

var (
	servePort int
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SCRIBE server",
	Long:  `Start the SCRIBE HTTP server with dashboard and API endpoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := NewOutput()
		config := GetConfig()

		// Use config values if flags weren't explicitly set
		if !cmd.Flags().Changed("port") {
			servePort = config.Server.Port
		}
		if !cmd.Flags().Changed("host") {
			serveHost = config.Server.Host
		}

		// Ensure database directory exists
		dbPath := GetDBPath()
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}

		out.Verbose("Database path: %s", dbPath)

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

		out.Verbose("Database initialized")

		// Create and start server
		server := http.NewServer(db)

		// Set embedded web assets
		server.SetStaticFS(web.DistFS)

		out.Info("Starting SCRIBE server on %s:%d", serveHost, servePort)
		out.Verbose("Read timeout: %ds, Write timeout: %ds", config.Server.ReadTimeout, config.Server.WriteTimeout)

		return server.Start(servePort)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "host to bind to")
	rootCmd.AddCommand(serveCmd)
}
