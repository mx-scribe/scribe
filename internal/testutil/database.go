package testutil

import (
	"testing"

	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// SetupTestDB creates an in-memory SQLite database for testing.
func SetupTestDB(t *testing.T) *sqlite.Database {
	t.Helper()

	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// CleanupTestDB closes and cleans up the test database.
func CleanupTestDB(t *testing.T, db *sqlite.Database) {
	t.Helper()

	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// TruncateLogs removes all logs from the test database.
func TruncateLogs(t *testing.T, db *sqlite.Database) {
	t.Helper()

	_, err := db.Conn().Exec("DELETE FROM logs")
	if err != nil {
		t.Fatalf("Failed to truncate logs table: %v", err)
	}
}

// GetLogCount returns the number of logs in the database.
func GetLogCount(t *testing.T, db *sqlite.Database) int {
	t.Helper()

	var count int
	err := db.Conn().QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to get log count: %v", err)
	}

	return count
}
