package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mx-scribe/scribe/internal/application/queries"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// GetStats handles GET /api/stats.
func GetStats(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := sqlite.NewLogRepository(db)
		handler := queries.NewGetStatsHandler(repo)

		stats, err := handler.Handle()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_ = json.NewEncoder(w).Encode(stats)
	}
}
