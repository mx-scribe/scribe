package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/mx-scribe/scribe/internal/application/commands"
	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// CreateLogRequest represents the request body for creating a log.
type CreateLogRequest struct {
	Header struct {
		Title       string `json:"title"`
		Severity    string `json:"severity,omitempty"`
		Source      string `json:"source,omitempty"`
		Color       string `json:"color,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"header"`
	Body map[string]any `json:"body,omitempty"`
}

// LogResponse represents a log in API responses.
type LogResponse struct {
	ID        int64          `json:"id"`
	Header    HeaderResponse `json:"header"`
	Body      map[string]any `json:"body"`
	Metadata  MetaResponse   `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at"`
}

// HeaderResponse represents the log header in responses.
type HeaderResponse struct {
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Source      string `json:"source,omitempty"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// MetaResponse represents the log metadata in responses.
type MetaResponse struct {
	DerivedSeverity string `json:"derived_severity,omitempty"`
	DerivedSource   string `json:"derived_source,omitempty"`
	DerivedCategory string `json:"derived_category,omitempty"`
}

// ListLogsResponse represents the paginated logs response.
type ListLogsResponse struct {
	Logs  []LogResponse `json:"logs"`
	Total int           `json:"total"`
	Limit int           `json:"limit"`
	Page  int           `json:"page"`
}

// CreateLog handles POST /api/logs.
func CreateLog(db *sqlite.Database) http.HandlerFunc {
	return CreateLogWithSSE(db, nil)
}

// CreateLogWithSSE handles POST /api/logs with SSE broadcast support.
func CreateLogWithSSE(db *sqlite.Database, hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateLogRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Header.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}

		repo := sqlite.NewLogRepository(db)
		handler := commands.NewCreateLogHandler(repo)

		input := commands.CreateLogInput{
			Title:       req.Header.Title,
			Severity:    req.Header.Severity,
			Source:      req.Header.Source,
			Color:       req.Header.Color,
			Description: req.Header.Description,
			Body:        req.Body,
		}

		output, err := handler.Handle(input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Broadcast to SSE clients if hub is available
		if hub != nil {
			log, _ := repo.FindByID(output.ID)
			if log != nil {
				hub.BroadcastLogCreated(log)
			}
		}

		response := map[string]any{
			"id":         output.ID,
			"title":      output.Title,
			"severity":   output.Severity,
			"created_at": output.CreatedAt,
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// DeleteLog handles DELETE /api/logs/{id}.
func DeleteLog(db *sqlite.Database) http.HandlerFunc {
	return DeleteLogWithSSE(db, nil)
}

// DeleteLogWithSSE handles DELETE /api/logs/{id} with SSE broadcast support.
func DeleteLogWithSSE(db *sqlite.Database, hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid log ID")
			return
		}

		repo := sqlite.NewLogRepository(db)

		// Check if log exists
		_, err = repo.FindByID(id)
		if err != nil {
			if err == entities.ErrLogNotFound {
				writeError(w, http.StatusNotFound, "log not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Delete the log
		if err := repo.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Broadcast to SSE clients if hub is available
		if hub != nil {
			hub.BroadcastLogDeleted(id)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteLogs handles DELETE /api/logs (bulk delete).
func DeleteLogs(db *sqlite.Database) http.HandlerFunc {
	return DeleteLogsWithSSE(db, nil)
}

// DeleteLogsWithSSE handles DELETE /api/logs (bulk delete) with SSE broadcast.
func DeleteLogsWithSSE(db *sqlite.Database, hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IDs []int64 `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if len(req.IDs) == 0 {
			writeError(w, http.StatusBadRequest, "ids are required")
			return
		}

		repo := sqlite.NewLogRepository(db)
		deleted := 0

		for _, id := range req.IDs {
			if err := repo.Delete(id); err == nil {
				deleted++
				if hub != nil {
					hub.BroadcastLogDeleted(id)
				}
			}
		}

		_ = json.NewEncoder(w).Encode(map[string]int{"deleted": deleted})
	}
}

// ListLogs handles GET /api/logs.
func ListLogs(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit

		filters := sqlite.LogFilters{
			Limit:    limit,
			Offset:   offset,
			Severity: r.URL.Query().Get("severity"),
			Source:   r.URL.Query().Get("source"),
			Search:   r.URL.Query().Get("search"),
			FromDate: r.URL.Query().Get("from"),
			ToDate:   r.URL.Query().Get("to"),
		}

		repo := sqlite.NewLogRepository(db)
		logs, total, err := repo.FindAll(filters)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := ListLogsResponse{
			Logs:  make([]LogResponse, 0, len(logs)),
			Total: total,
			Limit: limit,
			Page:  page,
		}

		for _, log := range logs {
			response.Logs = append(response.Logs, logToResponse(log))
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetLog handles GET /api/logs/{id}.
func GetLog(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid log ID")
			return
		}

		repo := sqlite.NewLogRepository(db)
		log, err := repo.FindByID(id)
		if err != nil {
			if err == entities.ErrLogNotFound {
				writeError(w, http.StatusNotFound, "log not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_ = json.NewEncoder(w).Encode(logToResponse(log))
	}
}

// logToResponse converts a Log entity to a LogResponse.
func logToResponse(log *entities.Log) LogResponse {
	return LogResponse{
		ID: log.ID,
		Header: HeaderResponse{
			Title:       log.Header.Title,
			Severity:    string(log.EffectiveSeverity()),
			Source:      log.Header.Source,
			Color:       string(log.EffectiveColor()),
			Description: log.Header.Description,
		},
		Body: log.Body,
		Metadata: MetaResponse{
			DerivedSeverity: log.Metadata.DerivedSeverity,
			DerivedSource:   log.Metadata.DerivedSource,
			DerivedCategory: log.Metadata.DerivedCategory,
		},
		CreatedAt: log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
