package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// RetentionConfig represents log retention configuration.
type RetentionConfig struct {
	// RetentionDays is the number of days to keep logs (0 = keep forever)
	RetentionDays int `json:"retention_days"`
}

// RetentionStats represents the result of a cleanup operation.
type RetentionStats struct {
	DeletedCount int64  `json:"deleted_count"`
	CutoffDate   string `json:"cutoff_date"`
	Message      string `json:"message"`
}

// CleanupLogs handles POST /api/admin/cleanup.
// Deletes logs older than the specified retention period.
func CleanupLogs(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config RetentionConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if config.RetentionDays <= 0 {
			writeError(w, http.StatusBadRequest, "retention_days must be greater than 0")
			return
		}

		cutoffDate := time.Now().AddDate(0, 0, -config.RetentionDays)

		repo := sqlite.NewLogRepository(db)
		deleted, err := repo.DeleteOlderThan(cutoffDate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := RetentionStats{
			DeletedCount: deleted,
			CutoffDate:   cutoffDate.Format(time.RFC3339),
			Message:      "Cleanup completed successfully",
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetRetentionInfo handles GET /api/admin/retention.
// Returns information about log age distribution.
func GetRetentionInfo(db *sqlite.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := sqlite.NewLogRepository(db)

		total, err := repo.Count()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		last24h, err := repo.CountLast24Hours()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Get counts by age buckets
		ageBuckets, err := getLogAgeBuckets(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := map[string]any{
			"total":         total,
			"last_24_hours": last24h,
			"by_age":        ageBuckets,
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// getLogAgeBuckets returns log counts grouped by age.
func getLogAgeBuckets(db *sqlite.Database) (map[string]int, error) {
	now := time.Now()
	buckets := map[string]int{
		"today":      0,
		"yesterday":  0,
		"last_week":  0,
		"last_month": 0,
		"older":      0,
	}

	// Query for each bucket
	queries := []struct {
		bucket string
		from   time.Time
		to     time.Time
	}{
		{"today", now.Truncate(24 * time.Hour), now},
		{"yesterday", now.Truncate(24 * time.Hour).Add(-24 * time.Hour), now.Truncate(24 * time.Hour)},
		{"last_week", now.AddDate(0, 0, -7), now.Truncate(24 * time.Hour).Add(-24 * time.Hour)},
		{"last_month", now.AddDate(0, -1, 0), now.AddDate(0, 0, -7)},
	}

	for _, q := range queries {
		var count int
		err := db.Conn().QueryRow(
			"SELECT COUNT(*) FROM logs WHERE created_at >= ? AND created_at < ?",
			q.from, q.to,
		).Scan(&count)
		if err != nil {
			return nil, err
		}
		buckets[q.bucket] = count
	}

	// Older than a month
	var olderCount int
	err := db.Conn().QueryRow(
		"SELECT COUNT(*) FROM logs WHERE created_at < ?",
		now.AddDate(0, -1, 0),
	).Scan(&olderCount)
	if err != nil {
		return nil, err
	}
	buckets["older"] = olderCount

	return buckets, nil
}
