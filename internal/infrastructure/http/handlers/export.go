package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// ExportJSON handles GET /api/export/json.
func ExportJSON(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := getAllLogs(db, r)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Set download headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=scribe-logs.json")

		// Convert to response format
		response := make([]LogResponse, 0, len(logs))
		for _, log := range logs {
			response = append(response, logToResponse(log))
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// ExportCSV handles GET /api/export/csv.
func ExportCSV(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := getAllLogs(db, r)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Set download headers
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=scribe-logs.csv")

		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		// Header
		_ = csvWriter.Write([]string{"id", "severity", "source", "title", "description", "created_at"})

		// Rows
		for _, log := range logs {
			row := []string{
				strconv.FormatInt(log.ID, 10),
				string(log.EffectiveSeverity()),
				log.Header.Source,
				log.Header.Title,
				log.Header.Description,
				log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			_ = csvWriter.Write(row)
		}
	}
}

// getAllLogs retrieves all logs with optional filters.
func getAllLogs(db *sqlite.Database, r *http.Request) ([]*entities.Log, error) {
	filters := sqlite.LogFilters{
		Limit:    10000, // Max export limit
		Severity: r.URL.Query().Get("severity"),
		Source:   r.URL.Query().Get("source"),
		Search:   r.URL.Query().Get("search"),
		FromDate: r.URL.Query().Get("from"),
		ToDate:   r.URL.Query().Get("to"),
	}

	repo := sqlite.NewLogRepository(db)
	logs, _, err := repo.FindAll(filters)
	return logs, err
}
