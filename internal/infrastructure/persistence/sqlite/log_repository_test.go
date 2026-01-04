package sqlite

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

func setupTestDB(t *testing.T) (*Database, func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create database: %v", err)
	}

	// Run migrations
	if err := RunMigrations(db.Conn()); err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func createTestLog(title string, severity valueobjects.Severity) *entities.Log {
	header := entities.LogHeader{
		Title:    title,
		Severity: severity,
	}
	return entities.NewLog(header, make(map[string]any))
}

func TestLogRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	log := createTestLog("Test log", valueobjects.SeverityInfo)
	log.Header.Source = "test-source"
	log.Header.Description = "Test description"
	log.Body["key"] = "value"

	err := repo.Create(log)
	if err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	if log.ID == 0 {
		t.Error("expected log ID to be set after create")
	}
}

func TestLogRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create a log first
	original := createTestLog("Find me", valueobjects.SeverityWarning)
	original.Header.Source = "test"
	original.Body["test"] = "data"
	if err := repo.Create(original); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	// Find it
	found, err := repo.FindByID(original.ID)
	if err != nil {
		t.Fatalf("failed to find log: %v", err)
	}

	if found.Header.Title != original.Header.Title {
		t.Errorf("title mismatch: got %q, want %q", found.Header.Title, original.Header.Title)
	}
	if found.Header.Severity != original.Header.Severity {
		t.Errorf("severity mismatch: got %v, want %v", found.Header.Severity, original.Header.Severity)
	}
	if found.Body["test"] != "data" {
		t.Errorf("body mismatch: got %v, want %v", found.Body["test"], "data")
	}
}

func TestLogRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	_, err := repo.FindByID(99999)
	if err != entities.ErrLogNotFound {
		t.Errorf("expected ErrLogNotFound, got %v", err)
	}
}

func TestLogRepository_FindAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create multiple logs
	severities := []valueobjects.Severity{
		valueobjects.SeverityError,
		valueobjects.SeverityWarning,
		valueobjects.SeverityInfo,
	}
	titles := []string{"Log A", "Log B", "Log C"}
	for i, sev := range severities {
		log := createTestLog(titles[i], sev)
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Find all without filters
	logs, total, err := repo.FindAll(LogFilters{})
	if err != nil {
		t.Fatalf("failed to find all logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}

	// Should be ordered by created_at DESC (newest first)
	if logs[0].Header.Title != "Log C" {
		t.Errorf("expected newest log first, got %q", logs[0].Header.Title)
	}
}

func TestLogRepository_FindAll_WithFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create logs with different severities
	log1 := createTestLog("Error log", valueobjects.SeverityError)
	log1.Header.Source = "api"
	if err := repo.Create(log1); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log2 := createTestLog("Warning log", valueobjects.SeverityWarning)
	log2.Header.Source = "database"
	if err := repo.Create(log2); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log3 := createTestLog("Another error", valueobjects.SeverityError)
	log3.Header.Source = "api"
	if err := repo.Create(log3); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	// Filter by severity
	logs, total, err := repo.FindAll(LogFilters{Severity: "error"})
	if err != nil {
		t.Fatalf("failed to filter by severity: %v", err)
	}
	if len(logs) != 2 || total != 2 {
		t.Errorf("expected 2 error logs, got %d (total: %d)", len(logs), total)
	}

	// Filter by source
	logs, total, err = repo.FindAll(LogFilters{Source: "api"})
	if err != nil {
		t.Fatalf("failed to filter by source: %v", err)
	}
	if len(logs) != 2 || total != 2 {
		t.Errorf("expected 2 api logs, got %d (total: %d)", len(logs), total)
	}

	// Filter by search
	logs, total, err = repo.FindAll(LogFilters{Search: "Warning"})
	if err != nil {
		t.Fatalf("failed to filter by search: %v", err)
	}
	if len(logs) != 1 || total != 1 {
		t.Errorf("expected 1 warning log, got %d (total: %d)", len(logs), total)
	}
}

func TestLogRepository_FindAll_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create 5 logs
	for i := 0; i < 5; i++ {
		log := createTestLog("Log", valueobjects.SeverityInfo)
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Get first page
	logs, total, err := repo.FindAll(LogFilters{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("failed to get first page: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs on first page, got %d", len(logs))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	// Get second page
	logs, _, err = repo.FindAll(LogFilters{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("failed to get second page: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs on second page, got %d", len(logs))
	}

	// Get last page
	logs, _, err = repo.FindAll(LogFilters{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("failed to get last page: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log on last page, got %d", len(logs))
	}
}

func TestLogRepository_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Initially empty
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Add logs
	for i := 0; i < 3; i++ {
		log := createTestLog("Log", valueobjects.SeverityInfo)
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	count, err = repo.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestLogRepository_CountLast24Hours(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create a recent log
	log := createTestLog("Recent", valueobjects.SeverityInfo)
	if err := repo.Create(log); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	count, err := repo.CountLast24Hours()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 recent log, got %d", count)
	}
}

func TestLogRepository_DeleteOlderThan(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create some logs
	for i := 0; i < 3; i++ {
		log := createTestLog("Log", valueobjects.SeverityInfo)
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	// Delete logs older than future (should delete all)
	cutoff := time.Now().Add(1 * time.Hour)
	deleted, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if deleted != 3 {
		t.Errorf("expected 3 deleted, got %d", deleted)
	}

	// Verify count is 0
	count, _ := repo.Count()
	if count != 0 {
		t.Errorf("expected 0 logs remaining, got %d", count)
	}
}

func TestLogRepository_CountBySeverity(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create logs with different severities
	severities := []valueobjects.Severity{
		valueobjects.SeverityError,
		valueobjects.SeverityError,
		valueobjects.SeverityWarning,
		valueobjects.SeverityInfo,
		valueobjects.SeverityInfo,
		valueobjects.SeverityInfo,
	}
	for _, sev := range severities {
		log := createTestLog("Log", sev)
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	counts, err := repo.CountBySeverity()
	if err != nil {
		t.Fatalf("failed to count by severity: %v", err)
	}

	if counts["error"] != 2 {
		t.Errorf("expected 2 errors, got %d", counts["error"])
	}
	if counts["warning"] != 1 {
		t.Errorf("expected 1 warning, got %d", counts["warning"])
	}
	if counts["info"] != 3 {
		t.Errorf("expected 3 info, got %d", counts["info"])
	}
}

func TestLogRepository_CountBySource(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create logs with different sources (empty values map to "unknown")
	sources := []string{"api", "api", "database", "", ""}
	for _, src := range sources {
		log := createTestLog("Log", valueobjects.SeverityInfo)
		log.Header.Source = src
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	counts, err := repo.CountBySource()
	if err != nil {
		t.Fatalf("failed to count by source: %v", err)
	}

	if counts["api"] != 2 {
		t.Errorf("expected 2 from api, got %d", counts["api"])
	}
	if counts["database"] != 1 {
		t.Errorf("expected 1 from database, got %d", counts["database"])
	}
	if counts["unknown"] != 2 {
		t.Errorf("expected 2 unknown, got %d", counts["unknown"])
	}
}

func TestLogRepository_FindAll_ColorFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create logs with different colors
	log1 := createTestLog("Red log", valueobjects.SeverityError)
	log1.Header.Color = "red"
	if err := repo.Create(log1); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log2 := createTestLog("Blue log", valueobjects.SeverityInfo)
	log2.Header.Color = "blue"
	if err := repo.Create(log2); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log3 := createTestLog("Another red log", valueobjects.SeverityError)
	log3.Header.Color = "red"
	if err := repo.Create(log3); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	// Filter by color
	logs, total, err := repo.FindAll(LogFilters{Color: "red"})
	if err != nil {
		t.Fatalf("failed to filter by color: %v", err)
	}
	if len(logs) != 2 || total != 2 {
		t.Errorf("expected 2 red logs, got %d (total: %d)", len(logs), total)
	}
}

func TestLogRepository_FindAll_DateFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create logs
	log := createTestLog("Today's log", valueobjects.SeverityInfo)
	if err := repo.Create(log); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	// Filter with from date in the past (should include)
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05Z07:00")
	logs, _, err := repo.FindAll(LogFilters{FromDate: yesterday})
	if err != nil {
		t.Fatalf("failed to filter by from date: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log from yesterday filter, got %d", len(logs))
	}

	// Filter with from date in the future (should exclude)
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05Z07:00")
	logs, _, err = repo.FindAll(LogFilters{FromDate: tomorrow})
	if err != nil {
		t.Fatalf("failed to filter by future from date: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs from future filter, got %d", len(logs))
	}

	// Filter with to date in the future (should include)
	logs, _, err = repo.FindAll(LogFilters{ToDate: tomorrow})
	if err != nil {
		t.Fatalf("failed to filter by to date: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log from to date filter, got %d", len(logs))
	}
}

func TestLogRepository_FindAll_CombinedFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewLogRepository(db)

	// Create various logs
	log1 := createTestLog("API error", valueobjects.SeverityError)
	log1.Header.Source = "api"
	if err := repo.Create(log1); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log2 := createTestLog("API warning", valueobjects.SeverityWarning)
	log2.Header.Source = "api"
	if err := repo.Create(log2); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	log3 := createTestLog("DB error", valueobjects.SeverityError)
	log3.Header.Source = "database"
	if err := repo.Create(log3); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	// Combine source + severity
	logs, total, err := repo.FindAll(LogFilters{
		Source:   "api",
		Severity: "error",
	})
	if err != nil {
		t.Fatalf("failed with combined filters: %v", err)
	}
	if len(logs) != 1 || total != 1 {
		t.Errorf("expected 1 api error, got %d (total: %d)", len(logs), total)
	}
	if logs[0].Header.Title != "API error" {
		t.Errorf("expected 'API error', got %q", logs[0].Header.Title)
	}
}
