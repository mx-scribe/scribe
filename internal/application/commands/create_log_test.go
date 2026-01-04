package commands

import (
	"testing"

	"github.com/mx-scribe/scribe/internal/domain/entities"
)

// mockLogRepository implements LogRepository for testing.
type mockLogRepository struct {
	logs    []*entities.Log
	nextID  int64
	lastLog *entities.Log
}

func newMockLogRepository() *mockLogRepository {
	return &mockLogRepository{
		logs:   make([]*entities.Log, 0),
		nextID: 1,
	}
}

func (m *mockLogRepository) Create(log *entities.Log) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	m.lastLog = log
	return nil
}

func TestCreateLogHandler_Handle(t *testing.T) {
	repo := newMockLogRepository()
	handler := NewCreateLogHandler(repo)

	input := CreateLogInput{
		Title:       "Test log",
		Severity:    "info",
		Source:      "test",
		Description: "A test log entry",
		Body: map[string]any{
			"key": "value",
		},
	}

	output, err := handler.Handle(input)
	if err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	if output.ID != 1 {
		t.Errorf("expected ID 1, got %d", output.ID)
	}
	if output.Title != "Test log" {
		t.Errorf("expected title 'Test log', got %q", output.Title)
	}
	if output.Severity != "info" {
		t.Errorf("expected severity 'info', got %q", output.Severity)
	}

	// Verify the log was stored
	if repo.lastLog == nil {
		t.Fatal("expected log to be stored")
	}
	if repo.lastLog.Body["key"] != "value" {
		t.Errorf("expected body key=value, got %v", repo.lastLog.Body["key"])
	}
}

func TestCreateLogHandler_Handle_MinimalInput(t *testing.T) {
	repo := newMockLogRepository()
	handler := NewCreateLogHandler(repo)

	input := CreateLogInput{
		Title: "Minimal log",
	}

	output, err := handler.Handle(input)
	if err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	if output.Title != "Minimal log" {
		t.Errorf("expected title 'Minimal log', got %q", output.Title)
	}
	// Default severity should be info
	if output.Severity != "info" {
		t.Errorf("expected default severity 'info', got %q", output.Severity)
	}
}

func TestCreateLogHandler_Handle_MissingTitle(t *testing.T) {
	repo := newMockLogRepository()
	handler := NewCreateLogHandler(repo)

	input := CreateLogInput{
		Severity: "error",
	}

	_, err := handler.Handle(input)
	if err != entities.ErrMissingTitle {
		t.Errorf("expected ErrMissingTitle, got %v", err)
	}
}

func TestCreateLogHandler_Handle_WithColor(t *testing.T) {
	repo := newMockLogRepository()
	handler := NewCreateLogHandler(repo)

	input := CreateLogInput{
		Title: "Colored log",
		Color: "blue",
	}

	output, err := handler.Handle(input)
	if err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	if output.Title != "Colored log" {
		t.Errorf("expected title 'Colored log', got %q", output.Title)
	}

	// Verify color was set
	if repo.lastLog.Header.Color.String() != "blue" {
		t.Errorf("expected color 'blue', got %q", repo.lastLog.Header.Color.String())
	}
}
