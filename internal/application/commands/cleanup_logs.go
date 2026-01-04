package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// CleanupLogsHandler handles the cleanup of old log entries based on retention policy.
type CleanupLogsHandler struct {
	logRepo *sqlite.LogRepository
}

// NewCleanupLogsHandler creates a new CleanupLogsHandler.
func NewCleanupLogsHandler(logRepo *sqlite.LogRepository) *CleanupLogsHandler {
	return &CleanupLogsHandler{
		logRepo: logRepo,
	}
}

// CleanupLogsRequest represents the input for cleanup operation.
type CleanupLogsRequest struct {
	RetentionDays int `json:"retention_days"`
}

// CleanupLogsResponse represents the output of cleanup operation.
type CleanupLogsResponse struct {
	DeletedCount int       `json:"deleted_count"`
	CutoffDate   time.Time `json:"cutoff_date"`
	Message      string    `json:"message"`
}

// Handle performs the log cleanup operation.
func (h *CleanupLogsHandler) Handle(ctx context.Context, request CleanupLogsRequest) (*CleanupLogsResponse, error) {
	// Validate retention days
	if request.RetentionDays < 1 {
		return nil, fmt.Errorf("retention_days must be at least 1")
	}

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -request.RetentionDays)

	// Delete old logs
	deletedCount, err := h.logRepo.DeleteOlderThan(cutoffDate)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old logs: %w", err)
	}

	// Build response
	message := fmt.Sprintf("Cleaned up %d logs older than %d days", deletedCount, request.RetentionDays)
	if deletedCount == 0 {
		message = fmt.Sprintf("No logs older than %d days to clean up", request.RetentionDays)
	}

	response := &CleanupLogsResponse{
		DeletedCount: int(deletedCount),
		CutoffDate:   cutoffDate,
		Message:      message,
	}

	return response, nil
}
