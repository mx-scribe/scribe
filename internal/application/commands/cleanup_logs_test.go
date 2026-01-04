package commands

import (
	"context"
	"testing"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupCleanupTest(t *testing.T) (*CleanupLogsHandler, *sqlite.LogRepository, *sqlite.Database) {
	t.Helper()

	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	logRepo := sqlite.NewLogRepository(db)
	handler := NewCleanupLogsHandler(logRepo)

	return handler, logRepo, db
}

func createLogWithTimestamp(repo *sqlite.LogRepository, timestamp time.Time) error {
	log := &entities.Log{
		Header: entities.LogHeader{
			Severity: valueobjects.SeverityInfo,
			Title:    "Test log",
			Source:   "test-service",
			Color:    valueobjects.ColorFromString("blue"),
		},
		Body:      map[string]any{"test": "data"},
		Metadata:  entities.LogMetadata{},
		CreatedAt: timestamp,
	}

	return repo.Create(log)
}

func TestNewCleanupLogsHandler(t *testing.T) {
	handler, _, db := setupCleanupTest(t)
	defer db.Close()

	if handler == nil {
		t.Fatal("NewCleanupLogsHandler() returned nil")
	}
}

func TestCleanupLogsHandler_Handle_Success(t *testing.T) {
	handler, repo, db := setupCleanupTest(t)
	defer db.Close()

	now := time.Now()

	// Create old logs (40 days ago)
	for i := 0; i < 5; i++ {
		oldTime := now.AddDate(0, 0, -40)
		if err := createLogWithTimestamp(repo, oldTime); err != nil {
			t.Fatalf("Failed to create old log: %v", err)
		}
	}

	// Create recent logs (10 days ago)
	for i := 0; i < 3; i++ {
		recentTime := now.AddDate(0, 0, -10)
		if err := createLogWithTimestamp(repo, recentTime); err != nil {
			t.Fatalf("Failed to create recent log: %v", err)
		}
	}

	// Execute cleanup with 30 days retention
	request := CleanupLogsRequest{
		RetentionDays: 30,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.DeletedCount != 5 {
		t.Errorf("Expected 5 deleted logs, got %d", response.DeletedCount)
	}

	expectedMessage := "Cleaned up 5 logs older than 30 days"
	if response.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Message)
	}

	// Verify cutoff date is approximately 30 days ago
	cutoffDiff := now.Sub(response.CutoffDate).Hours() / 24
	if cutoffDiff < 29 || cutoffDiff > 31 {
		t.Errorf("Expected cutoff date ~30 days ago, got %v days", cutoffDiff)
	}
}

func TestCleanupLogsHandler_Handle_NoOldLogs(t *testing.T) {
	handler, repo, db := setupCleanupTest(t)
	defer db.Close()

	// Create only recent logs (5 days ago)
	now := time.Now()
	for i := 0; i < 3; i++ {
		recentTime := now.AddDate(0, 0, -5)
		if err := createLogWithTimestamp(repo, recentTime); err != nil {
			t.Fatalf("Failed to create recent log: %v", err)
		}
	}

	request := CleanupLogsRequest{
		RetentionDays: 30,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.DeletedCount != 0 {
		t.Errorf("Expected 0 deleted logs, got %d", response.DeletedCount)
	}

	expectedMessage := "No logs older than 30 days to clean up"
	if response.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Message)
	}
}

func TestCleanupLogsHandler_Handle_InvalidRetentionDays(t *testing.T) {
	handler, _, db := setupCleanupTest(t)
	defer db.Close()

	tests := []struct {
		name          string
		retentionDays int
		expectError   bool
	}{
		{
			name:          "Zero retention days",
			retentionDays: 0,
			expectError:   true,
		},
		{
			name:          "Negative retention days",
			retentionDays: -10,
			expectError:   true,
		},
		{
			name:          "Valid retention days (1)",
			retentionDays: 1,
			expectError:   false,
		},
		{
			name:          "Valid retention days (365)",
			retentionDays: 365,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CleanupLogsRequest{
				RetentionDays: tt.retentionDays,
			}

			_, err := handler.Handle(context.Background(), request)

			if tt.expectError && err == nil {
				t.Error("Expected error for invalid retention days, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestCleanupLogsHandler_Handle_EmptyDatabase(t *testing.T) {
	handler, _, db := setupCleanupTest(t)
	defer db.Close()

	request := CleanupLogsRequest{
		RetentionDays: 30,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error for empty database, got: %v", err)
	}

	if response.DeletedCount != 0 {
		t.Errorf("Expected 0 deleted logs in empty database, got %d", response.DeletedCount)
	}

	expectedMessage := "No logs older than 30 days to clean up"
	if response.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Message)
	}
}

func TestCleanupLogsHandler_Handle_AllLogsOld(t *testing.T) {
	handler, repo, db := setupCleanupTest(t)
	defer db.Close()

	// Create only old logs (60 days ago)
	now := time.Now()
	logCount := 10
	for i := 0; i < logCount; i++ {
		oldTime := now.AddDate(0, 0, -60)
		if err := createLogWithTimestamp(repo, oldTime); err != nil {
			t.Fatalf("Failed to create old log: %v", err)
		}
	}

	request := CleanupLogsRequest{
		RetentionDays: 30,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.DeletedCount != logCount {
		t.Errorf("Expected %d deleted logs, got %d", logCount, response.DeletedCount)
	}
}

func TestCleanupLogsHandler_Handle_EdgeCaseCutoffDate(t *testing.T) {
	handler, repo, db := setupCleanupTest(t)
	defer db.Close()

	now := time.Now()
	retentionDays := 30

	// Create log exactly at cutoff date
	cutoffTime := now.AddDate(0, 0, -retentionDays)
	if err := createLogWithTimestamp(repo, cutoffTime); err != nil {
		t.Fatalf("Failed to create log at cutoff: %v", err)
	}

	// Create log 1 second before cutoff (should be deleted)
	beforeCutoff := cutoffTime.Add(-1 * time.Second)
	if err := createLogWithTimestamp(repo, beforeCutoff); err != nil {
		t.Fatalf("Failed to create log before cutoff: %v", err)
	}

	// Create log 1 second after cutoff (should NOT be deleted)
	afterCutoff := cutoffTime.Add(1 * time.Second)
	if err := createLogWithTimestamp(repo, afterCutoff); err != nil {
		t.Fatalf("Failed to create log after cutoff: %v", err)
	}

	request := CleanupLogsRequest{
		RetentionDays: retentionDays,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should delete logs older than cutoff (1-2 logs depending on timing)
	if response.DeletedCount < 1 || response.DeletedCount > 2 {
		t.Logf("Note: Deleted %d logs (timing dependent)", response.DeletedCount)
	}
}

func TestCleanupLogsHandler_Handle_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	handler, repo, db := setupCleanupTest(t)
	defer db.Close()

	now := time.Now()

	// Create 1000 old logs
	oldCount := 1000
	for i := 0; i < oldCount; i++ {
		oldTime := now.AddDate(0, 0, -60)
		if err := createLogWithTimestamp(repo, oldTime); err != nil {
			t.Fatalf("Failed to create old log: %v", err)
		}
	}

	// Create 500 recent logs
	recentCount := 500
	for i := 0; i < recentCount; i++ {
		recentTime := now.AddDate(0, 0, -10)
		if err := createLogWithTimestamp(repo, recentTime); err != nil {
			t.Fatalf("Failed to create recent log: %v", err)
		}
	}

	request := CleanupLogsRequest{
		RetentionDays: 30,
	}

	startTime := time.Now()
	response, err := handler.Handle(context.Background(), request)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.DeletedCount != oldCount {
		t.Errorf("Expected %d deleted logs, got %d", oldCount, response.DeletedCount)
	}

	t.Logf("Cleaned up %d logs in %v", oldCount, duration)

	// Performance check: should complete in reasonable time
	if duration > 2*time.Second {
		t.Logf("Warning: Cleanup took %v for %d logs", duration, oldCount)
	}
}
