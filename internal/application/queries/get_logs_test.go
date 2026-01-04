package queries

import (
	"context"
	"testing"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupGetLogsTest(t *testing.T) (*GetLogsHandler, *sqlite.LogRepository, *sqlite.Database) {
	t.Helper()

	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	logRepo := sqlite.NewLogRepository(db)
	handler := NewGetLogsHandler(logRepo)

	return handler, logRepo, db
}

func createTestLogEntry(repo *sqlite.LogRepository, severity valueobjects.Severity, title string, color valueobjects.Color) error {
	log := &entities.Log{
		Header: entities.LogHeader{
			Severity: severity,
			Title:    title,
			Source:   "test-service",
			Color:    color,
		},
		Body:      map[string]any{"test": "data"},
		Metadata:  entities.LogMetadata{},
		CreatedAt: time.Now(),
	}

	return repo.Create(log)
}

func TestNewGetLogsHandler(t *testing.T) {
	handler, _, db := setupGetLogsTest(t)
	defer db.Close()

	if handler == nil {
		t.Fatal("NewGetLogsHandler() returned nil")
	}
}

func TestGetLogsHandler_Handle_EmptyDatabase(t *testing.T) {
	handler, _, db := setupGetLogsTest(t)
	defer db.Close()

	request := GetLogsRequest{}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(response.Logs))
	}

	if response.TotalCount != 0 {
		t.Errorf("Expected total count 0, got %d", response.TotalCount)
	}
}

func TestGetLogsHandler_Handle_WithLogs(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create test logs
	for i := 0; i < 5; i++ {
		if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := GetLogsRequest{}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(response.Logs))
	}

	if response.TotalCount != 5 {
		t.Errorf("Expected total count 5, got %d", response.TotalCount)
	}
}

func TestGetLogsHandler_Handle_DefaultLimit(t *testing.T) {
	handler, _, db := setupGetLogsTest(t)
	defer db.Close()

	request := GetLogsRequest{
		Limit: 0, // Should default to 100
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Limit != 100 {
		t.Errorf("Expected default limit 100, got %d", response.Limit)
	}
}

func TestGetLogsHandler_Handle_CustomLimit(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create 10 logs
	for i := 0; i < 10; i++ {
		if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := GetLogsRequest{
		Limit: 5,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 5 {
		t.Errorf("Expected 5 logs (limit), got %d", len(response.Logs))
	}

	if response.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", response.Limit)
	}

	if response.TotalCount != 10 {
		t.Errorf("Expected total count 10, got %d", response.TotalCount)
	}
}

func TestGetLogsHandler_Handle_MaxLimit(t *testing.T) {
	handler, _, db := setupGetLogsTest(t)
	defer db.Close()

	request := GetLogsRequest{
		Limit: 2000, // Should be capped at 1000
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Limit != 1000 {
		t.Errorf("Expected max limit 1000, got %d", response.Limit)
	}
}

func TestGetLogsHandler_Handle_WithOffset(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create 10 logs
	for i := 0; i < 10; i++ {
		if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := GetLogsRequest{
		Limit:  5,
		Offset: 3,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(response.Logs))
	}

	if response.Offset != 3 {
		t.Errorf("Expected offset 3, got %d", response.Offset)
	}

	if response.TotalCount != 10 {
		t.Errorf("Expected total count 10, got %d", response.TotalCount)
	}
}

func TestGetLogsHandler_Handle_NegativeOffset(t *testing.T) {
	handler, _, db := setupGetLogsTest(t)
	defer db.Close()

	request := GetLogsRequest{
		Offset: -5, // Should be reset to 0
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Offset != 0 {
		t.Errorf("Expected offset 0 (negative reset), got %d", response.Offset)
	}
}

func TestGetLogsHandler_Handle_WithSearch(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create logs with different titles
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test error log", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test success log", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Another error message", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := GetLogsRequest{
		Search: "error",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should find logs containing "error"
	if len(response.Logs) != 2 {
		t.Errorf("Expected 2 logs with 'error', got %d", len(response.Logs))
	}
}

func TestGetLogsHandler_Handle_WithSeverityFilter(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create logs with different severities
	if err := createTestLogEntry(repo, valueobjects.SeverityError, "Error log", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Info log", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityWarning, "Warning log", valueobjects.ColorFromString("yellow")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := GetLogsRequest{
		Severity: "error",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(response.Logs))
	}
}

func TestGetLogsHandler_Handle_WithColorFilter(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create logs with different colors
	if err := createTestLogEntry(repo, valueobjects.SeverityError, "Error log", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Info log", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Another info", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := GetLogsRequest{
		Color: "blue",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 2 {
		t.Errorf("Expected 2 blue logs, got %d", len(response.Logs))
	}
}

func TestGetLogsHandler_Handle_WithMultipleFilters(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create diverse logs
	if err := createTestLogEntry(repo, valueobjects.SeverityError, "Database error", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityError, "API error", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Database query", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := GetLogsRequest{
		Search:   "database",
		Severity: "error",
		Color:    "red",
		Limit:    10,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 1 {
		t.Errorf("Expected 1 log matching all filters, got %d", len(response.Logs))
	}
}

func TestGetLogsHandler_Handle_Pagination(t *testing.T) {
	handler, repo, db := setupGetLogsTest(t)
	defer db.Close()

	// Create 15 logs
	for i := 0; i < 15; i++ {
		if err := createTestLogEntry(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Page 1
	page1 := GetLogsRequest{
		Limit:  5,
		Offset: 0,
	}

	response1, err := handler.Handle(context.Background(), page1)
	if err != nil {
		t.Fatalf("Page 1 error: %v", err)
	}

	// Page 2
	page2 := GetLogsRequest{
		Limit:  5,
		Offset: 5,
	}

	response2, err := handler.Handle(context.Background(), page2)
	if err != nil {
		t.Fatalf("Page 2 error: %v", err)
	}

	// Page 3
	page3 := GetLogsRequest{
		Limit:  5,
		Offset: 10,
	}

	response3, err := handler.Handle(context.Background(), page3)
	if err != nil {
		t.Fatalf("Page 3 error: %v", err)
	}

	if len(response1.Logs) != 5 {
		t.Errorf("Page 1: Expected 5 logs, got %d", len(response1.Logs))
	}
	if len(response2.Logs) != 5 {
		t.Errorf("Page 2: Expected 5 logs, got %d", len(response2.Logs))
	}
	if len(response3.Logs) != 5 {
		t.Errorf("Page 3: Expected 5 logs, got %d", len(response3.Logs))
	}

	// All should have same total count
	if response1.TotalCount != 15 || response2.TotalCount != 15 || response3.TotalCount != 15 {
		t.Error("All pages should have total count 15")
	}
}
