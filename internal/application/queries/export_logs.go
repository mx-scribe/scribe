package queries

import (
	"context"
	"fmt"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// ExportFormat represents the format for log export.
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

// ExportLogsHandler handles export of logs in various formats.
type ExportLogsHandler struct {
	logRepo *sqlite.LogRepository
}

// NewExportLogsHandler creates a new ExportLogsHandler.
func NewExportLogsHandler(logRepo *sqlite.LogRepository) *ExportLogsHandler {
	return &ExportLogsHandler{
		logRepo: logRepo,
	}
}

// ExportLogsRequest represents the input for exporting logs.
type ExportLogsRequest struct {
	Format   ExportFormat `json:"format"`
	Search   string       `json:"search,omitempty"`
	Severity string       `json:"severity,omitempty"`
	Source   string       `json:"source,omitempty"`
	Color    string       `json:"color,omitempty"`
	FromDate string       `json:"from_date,omitempty"`
	ToDate   string       `json:"to_date,omitempty"`
	Limit    int          `json:"limit,omitempty"`
}

// ExportLogsResponse represents the output of log export.
type ExportLogsResponse struct {
	Logs   []*entities.Log `json:"logs"`
	Format ExportFormat    `json:"format"`
	Count  int             `json:"count"`
}

// Handle retrieves logs for export with optional filters.
func (h *ExportLogsHandler) Handle(ctx context.Context, request ExportLogsRequest) (*ExportLogsResponse, error) {
	// Validate format
	if request.Format != ExportFormatCSV && request.Format != ExportFormatJSON {
		return nil, fmt.Errorf("invalid export format: %s (must be csv or json)", request.Format)
	}

	// Set default limit for exports
	if request.Limit <= 0 {
		request.Limit = 10000 // Higher limit for exports
	}
	if request.Limit > 100000 {
		request.Limit = 100000 // Maximum export limit
	}

	// Build filters
	filters := sqlite.LogFilters{
		Search:   request.Search,
		Severity: request.Severity,
		Source:   request.Source,
		Color:    request.Color,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
		Offset:   0, // Exports always start from beginning
	}

	// Retrieve logs
	logs, _, err := h.logRepo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs for export: %w", err)
	}

	// Build response
	response := &ExportLogsResponse{
		Logs:   logs,
		Format: request.Format,
		Count:  len(logs),
	}

	return response, nil
}
