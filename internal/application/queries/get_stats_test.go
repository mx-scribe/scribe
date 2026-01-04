package queries

import (
	"testing"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupGetStatsTest(t *testing.T) (*GetStatsHandler, *sqlite.LogRepository, *sqlite.Database) {
	t.Helper()

	// Create in-memory database
	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repository (LogRepository implements StatsRepository interface)
	logRepo := sqlite.NewLogRepository(db)

	// Create handler
	handler := NewGetStatsHandler(logRepo)

	return handler, logRepo, db
}

func createStatsTestLog(t *testing.T, repo *sqlite.LogRepository, severity, source string) {
	t.Helper()

	log := entities.NewLog(entities.LogHeader{
		Title:    "Test log",
		Severity: valueobjects.Severity(severity),
		Source:   source,
	}, nil)

	if err := repo.Create(log); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
}

func TestNewGetStatsHandler(t *testing.T) {
	handler, _, db := setupGetStatsTest(t)
	defer db.Close()

	if handler == nil {
		t.Fatal("NewGetStatsHandler() returned nil")
	}
}

func TestGetStatsHandler_Handle_EmptyDatabase(t *testing.T) {
	handler, _, db := setupGetStatsTest(t)
	defer db.Close()

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output == nil {
		t.Fatal("Expected output, got nil")
	}

	if output.Total != 0 {
		t.Errorf("Expected 0 total logs, got %d", output.Total)
	}

	if output.Last24Hours != 0 {
		t.Errorf("Expected 0 last 24 hours, got %d", output.Last24Hours)
	}
}

func TestGetStatsHandler_Handle_WithLogs(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create test logs with different severities
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "info", "database")
	createStatsTestLog(t, logRepo, "warning", "auth")

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.Total != 4 {
		t.Errorf("Expected 4 total logs, got %d", output.Total)
	}
}

func TestGetStatsHandler_Handle_BySeverity(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create logs with specific severities
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "warning", "auth")
	createStatsTestLog(t, logRepo, "warning", "auth")
	createStatsTestLog(t, logRepo, "info", "database")

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check severity breakdown exists
	if output.BySeverity == nil {
		t.Fatal("Expected BySeverity, got nil")
	}

	if output.BySeverity["error"] != 3 {
		t.Errorf("Expected 3 error logs, got %d", output.BySeverity["error"])
	}

	if output.BySeverity["warning"] != 2 {
		t.Errorf("Expected 2 warning logs, got %d", output.BySeverity["warning"])
	}

	if output.BySeverity["info"] != 1 {
		t.Errorf("Expected 1 info log, got %d", output.BySeverity["info"])
	}
}

func TestGetStatsHandler_Handle_BySource(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create logs with specific sources
	createStatsTestLog(t, logRepo, "error", "api-service")
	createStatsTestLog(t, logRepo, "error", "api-service")
	createStatsTestLog(t, logRepo, "info", "api-service")
	createStatsTestLog(t, logRepo, "warning", "database")
	createStatsTestLog(t, logRepo, "warning", "database")
	createStatsTestLog(t, logRepo, "info", "auth")

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check BySource exists
	if output.BySource == nil {
		t.Fatal("Expected BySource, got nil")
	}

	if output.BySource["api-service"] != 3 {
		t.Errorf("Expected 3 api-service logs, got %d", output.BySource["api-service"])
	}

	if output.BySource["database"] != 2 {
		t.Errorf("Expected 2 database logs, got %d", output.BySource["database"])
	}

	if output.BySource["auth"] != 1 {
		t.Errorf("Expected 1 auth log, got %d", output.BySource["auth"])
	}
}

func TestGetStatsHandler_Handle_MultipleCalls(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create initial logs
	for i := 0; i < 5; i++ {
		createStatsTestLog(t, logRepo, "info", "service")
	}

	// First call
	output1, err := handler.Handle()
	if err != nil {
		t.Fatalf("First call error: %v", err)
	}

	if output1.Total != 5 {
		t.Errorf("First call: Expected 5 logs, got %d", output1.Total)
	}

	// Add more logs
	for i := 0; i < 3; i++ {
		createStatsTestLog(t, logRepo, "error", "service")
	}

	// Second call - should reflect new logs
	output2, err := handler.Handle()
	if err != nil {
		t.Fatalf("Second call error: %v", err)
	}

	if output2.Total != 8 {
		t.Errorf("Second call: Expected 8 logs, got %d", output2.Total)
	}
}

func TestGetStatsHandler_Handle_Last24Hours(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create logs (they should all be within last 24 hours)
	createStatsTestLog(t, logRepo, "info", "service")
	createStatsTestLog(t, logRepo, "error", "service")

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// All recent logs should be in last 24 hours count
	if output.Last24Hours != 2 {
		t.Errorf("Expected 2 logs in last 24 hours, got %d", output.Last24Hours)
	}
}

func TestGetStatsHandler_Handle_ResponseStructure(t *testing.T) {
	handler, logRepo, db := setupGetStatsTest(t)
	defer db.Close()

	// Create sample logs
	createStatsTestLog(t, logRepo, "error", "api")
	createStatsTestLog(t, logRepo, "warning", "database")
	createStatsTestLog(t, logRepo, "info", "auth")

	output, err := handler.Handle()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify response structure
	if output.Total != 3 {
		t.Errorf("Expected Total 3, got %d", output.Total)
	}

	if output.BySeverity == nil {
		t.Error("Expected BySeverity to be non-nil")
	}

	if output.BySource == nil {
		t.Error("Expected BySource to be non-nil")
	}
}

func TestStatsOutput_Fields(t *testing.T) {
	output := &StatsOutput{
		Total:       100,
		Last24Hours: 50,
		BySeverity:  map[string]int{"error": 30, "warning": 20, "info": 50},
		BySource:    map[string]int{"api": 60, "database": 40},
	}

	if output.Total != 100 {
		t.Errorf("Expected Total 100, got %d", output.Total)
	}

	if output.Last24Hours != 50 {
		t.Errorf("Expected Last24Hours 50, got %d", output.Last24Hours)
	}

	if output.BySeverity["error"] != 30 {
		t.Errorf("Expected 30 errors, got %d", output.BySeverity["error"])
	}

	if output.BySource["api"] != 60 {
		t.Errorf("Expected 60 api logs, got %d", output.BySource["api"])
	}
}
