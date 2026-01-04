package queries

import (
	"context"
	"fmt"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// GetLogsHandler handles retrieval of logs with filtering.
type GetLogsHandler struct {
	logRepo *sqlite.LogRepository
}

// NewGetLogsHandler creates a new GetLogsHandler.
func NewGetLogsHandler(logRepo *sqlite.LogRepository) *GetLogsHandler {
	return &GetLogsHandler{
		logRepo: logRepo,
	}
}

// GetLogsRequest represents the input for retrieving logs.
type GetLogsRequest struct {
	Search   string `json:"search,omitempty"`
	Severity string `json:"severity,omitempty"`
	Source   string `json:"source,omitempty"`
	Color    string `json:"color,omitempty"`
	FromDate string `json:"from_date,omitempty"`
	ToDate   string `json:"to_date,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// GetLogsResponse represents the output of log retrieval.
type GetLogsResponse struct {
	Logs       []*entities.Log `json:"logs"`
	TotalCount int             `json:"total_count"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
}

// Handle retrieves logs with optional filters.
func (h *GetLogsHandler) Handle(ctx context.Context, request GetLogsRequest) (*GetLogsResponse, error) {
	if request.Limit <= 0 {
		request.Limit = 100
	}
	if request.Limit > 1000 {
		request.Limit = 1000
	}
	if request.Offset < 0 {
		request.Offset = 0
	}

	filters := sqlite.LogFilters{
		Search:   request.Search,
		Severity: request.Severity,
		Source:   request.Source,
		Color:    request.Color,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
		Offset:   request.Offset,
	}

	logs, totalCount, err := h.logRepo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs: %w", err)
	}

	response := &GetLogsResponse{
		Logs:       logs,
		TotalCount: totalCount,
		Limit:      request.Limit,
		Offset:     request.Offset,
	}

	return response, nil
}
