package queries

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupExportLogsTest(t *testing.T) (*ExportLogsHandler, *sqlite.LogRepository, *sqlite.Database) {
	t.Helper()

	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	logRepo := sqlite.NewLogRepository(db)
	handler := NewExportLogsHandler(logRepo)

	return handler, logRepo, db
}

func createExportTestLog(repo *sqlite.LogRepository, severity valueobjects.Severity, title string, color valueobjects.Color) error {
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

func TestNewExportLogsHandler(t *testing.T) {
	handler, _, db := setupExportLogsTest(t)
	defer db.Close()

	if handler == nil {
		t.Fatal("NewExportLogsHandler() returned nil")
	}
}

func TestExportLogsHandler_Handle_CSV_Success(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create test logs
	for i := 0; i < 5; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := ExportLogsRequest{
		Format: ExportFormatCSV,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Format != ExportFormatCSV {
		t.Errorf("Expected CSV format, got %s", response.Format)
	}

	if len(response.Logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(response.Logs))
	}

	if response.Count != 5 {
		t.Errorf("Expected count 5, got %d", response.Count)
	}
}

func TestExportLogsHandler_Handle_JSON_Success(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create test logs
	for i := 0; i < 3; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityError, "Test error", valueobjects.ColorFromString("red")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := ExportLogsRequest{
		Format: ExportFormatJSON,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Format != ExportFormatJSON {
		t.Errorf("Expected JSON format, got %s", response.Format)
	}

	if len(response.Logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(response.Logs))
	}

	if response.Count != 3 {
		t.Errorf("Expected count 3, got %d", response.Count)
	}
}

func TestExportLogsHandler_Handle_InvalidFormat(t *testing.T) {
	handler, _, db := setupExportLogsTest(t)
	defer db.Close()

	request := ExportLogsRequest{
		Format: "xml", // Invalid format
	}

	_, err := handler.Handle(context.Background(), request)

	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}

	if !strings.Contains(err.Error(), "invalid export format") {
		t.Errorf("Expected error about invalid format, got: %v", err)
	}
}

func TestExportLogsHandler_Handle_EmptyDatabase(t *testing.T) {
	handler, _, db := setupExportLogsTest(t)
	defer db.Close()

	request := ExportLogsRequest{
		Format: ExportFormatCSV,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(response.Logs))
	}

	if response.Count != 0 {
		t.Errorf("Expected count 0, got %d", response.Count)
	}
}

func TestExportLogsHandler_Handle_DefaultLimit(t *testing.T) {
	handler, _, db := setupExportLogsTest(t)
	defer db.Close()

	request := ExportLogsRequest{
		Format: ExportFormatCSV,
		Limit:  0, // Should default to 10000
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Error("Expected response, got nil")
	}
}

func TestExportLogsHandler_Handle_CustomLimit(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create 20 logs
	for i := 0; i < 20; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	request := ExportLogsRequest{
		Format: ExportFormatCSV,
		Limit:  10,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 10 {
		t.Errorf("Expected 10 logs (limit), got %d", len(response.Logs))
	}

	if response.Count != 10 {
		t.Errorf("Expected count 10, got %d", response.Count)
	}
}

func TestExportLogsHandler_Handle_MaxLimit(t *testing.T) {
	handler, _, db := setupExportLogsTest(t)
	defer db.Close()

	request := ExportLogsRequest{
		Format: ExportFormatCSV,
		Limit:  200000, // Should be capped at 100000
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Error("Expected response, got nil")
	}
}

func TestExportLogsHandler_Handle_WithSearch(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create logs with different titles
	if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Database error", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createExportTestLog(repo, valueobjects.SeverityInfo, "API success", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Database query", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := ExportLogsRequest{
		Format: ExportFormatJSON,
		Search: "database",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 2 {
		t.Errorf("Expected 2 logs with 'database', got %d", len(response.Logs))
	}
}

func TestExportLogsHandler_Handle_WithSeverityFilter(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create logs with different severities
	for i := 0; i < 3; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityError, "Error log", valueobjects.ColorFromString("red")); err != nil {
			t.Fatalf("Failed to create error log: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Info log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create info log: %v", err)
		}
	}

	request := ExportLogsRequest{
		Format:   ExportFormatCSV,
		Severity: "error",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 3 {
		t.Errorf("Expected 3 error logs, got %d", len(response.Logs))
	}
}

func TestExportLogsHandler_Handle_WithColorFilter(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create logs with different colors
	for i := 0; i < 4; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityError, "Error log", valueobjects.ColorFromString("red")); err != nil {
			t.Fatalf("Failed to create red log: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Info log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create blue log: %v", err)
		}
	}

	request := ExportLogsRequest{
		Format: ExportFormatJSON,
		Color:  "red",
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 4 {
		t.Errorf("Expected 4 red logs, got %d", len(response.Logs))
	}
}

func TestExportLogsHandler_Handle_WithMultipleFilters(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create diverse logs
	if err := createExportTestLog(repo, valueobjects.SeverityError, "Database error", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createExportTestLog(repo, valueobjects.SeverityError, "API error", valueobjects.ColorFromString("red")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
	if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Database query", valueobjects.ColorFromString("blue")); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	request := ExportLogsRequest{
		Format:   ExportFormatCSV,
		Search:   "database",
		Severity: "error",
		Color:    "red",
		Limit:    100,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 1 {
		t.Errorf("Expected 1 log matching all filters, got %d", len(response.Logs))
	}
}

func TestExportLogsHandler_Handle_CSVvsJSON(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create identical logs
	for i := 0; i < 5; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Export as CSV
	csvRequest := ExportLogsRequest{
		Format: ExportFormatCSV,
	}

	csvResponse, err := handler.Handle(context.Background(), csvRequest)
	if err != nil {
		t.Fatalf("CSV export error: %v", err)
	}

	// Export as JSON
	jsonRequest := ExportLogsRequest{
		Format: ExportFormatJSON,
	}

	jsonResponse, err := handler.Handle(context.Background(), jsonRequest)
	if err != nil {
		t.Fatalf("JSON export error: %v", err)
	}

	// Both should have same count
	if csvResponse.Count != jsonResponse.Count {
		t.Errorf("CSV count (%d) != JSON count (%d)", csvResponse.Count, jsonResponse.Count)
	}

	// Both should have same number of logs
	if len(csvResponse.Logs) != len(jsonResponse.Logs) {
		t.Errorf("CSV logs (%d) != JSON logs (%d)", len(csvResponse.Logs), len(jsonResponse.Logs))
	}
}

func TestExportLogsHandler_Handle_OffsetAlwaysZero(t *testing.T) {
	handler, repo, db := setupExportLogsTest(t)
	defer db.Close()

	// Create logs
	for i := 0; i < 10; i++ {
		if err := createExportTestLog(repo, valueobjects.SeverityInfo, "Test log", valueobjects.ColorFromString("blue")); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Exports always start from beginning (offset=0)
	request := ExportLogsRequest{
		Format: ExportFormatCSV,
		Limit:  10,
	}

	response, err := handler.Handle(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(response.Logs) != 10 {
		t.Errorf("Expected all 10 logs from beginning, got %d", len(response.Logs))
	}
}
