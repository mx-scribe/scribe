// Package testutil provides test utilities, fixtures, and helpers for SCRIBE tests.
package testutil

import (
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

// NewTestLog creates a test log with sensible defaults.
func NewTestLog() *entities.Log {
	return &entities.Log{
		ID: 1,
		Header: entities.LogHeader{
			Title:       "Test error log",
			Severity:    valueobjects.SeverityError,
			Source:      "test-service",
			Color:       "red",
			Description: "This is a test error description",
		},
		Body: map[string]any{
			"error":   "test error message",
			"code":    500,
			"details": "Something went wrong",
		},
		Metadata: entities.LogMetadata{
			DerivedSeverity: "error",
			DerivedSource:   "test-service",
			DerivedCategory: "system",
		},
		CreatedAt: time.Now(),
	}
}

// NewTestLogWithSeverity creates a test log with specific severity.
func NewTestLogWithSeverity(severity valueobjects.Severity) *entities.Log {
	log := NewTestLog()
	log.Header.Severity = severity
	log.Header.Color = valueobjects.AutoAssignColor(severity)
	return log
}

// NewTestLogWithTitle creates a test log with specific title.
func NewTestLogWithTitle(title string) *entities.Log {
	log := NewTestLog()
	log.Header.Title = title
	return log
}

// NewTestLogWithBody creates a test log with specific body.
func NewTestLogWithBody(body map[string]any) *entities.Log {
	log := NewTestLog()
	log.Body = body
	return log
}

// NewTestLogMinimal creates a minimal valid log (only required fields).
func NewTestLogMinimal() *entities.Log {
	return &entities.Log{
		Header: entities.LogHeader{
			Title:    "Minimal test log",
			Severity: valueobjects.SeverityInfo,
			Color:    "blue",
		},
		Body:      make(map[string]any),
		CreatedAt: time.Now(),
	}
}

// TestLogOptions allows customization of test logs.
type TestLogOptions struct {
	Severity    valueobjects.Severity
	Title       string
	Description string
	Source      string
	Color       valueobjects.Color
	Body        map[string]any
}

// NewTestLogWithOptions creates a test log with custom options.
func NewTestLogWithOptions(opts TestLogOptions) *entities.Log {
	log := NewTestLog()

	if opts.Severity != "" {
		log.Header.Severity = opts.Severity
	}
	if opts.Title != "" {
		log.Header.Title = opts.Title
	}
	if opts.Description != "" {
		log.Header.Description = opts.Description
	}
	if opts.Source != "" {
		log.Header.Source = opts.Source
	}
	if opts.Color != "" {
		log.Header.Color = opts.Color
	}
	if opts.Body != nil {
		log.Body = opts.Body
	}

	return log
}

// NewTestLogBatch creates multiple test logs.
func NewTestLogBatch(count int) []*entities.Log {
	logs := make([]*entities.Log, count)
	for i := 0; i < count; i++ {
		log := NewTestLog()
		log.ID = int64(i + 1)
		log.Header.Title = "Test log " + string(rune('A'+i))
		logs[i] = log
	}
	return logs
}
