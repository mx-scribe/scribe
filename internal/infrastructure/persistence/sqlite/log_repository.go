package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

// LogRepository handles log persistence operations.
type LogRepository struct {
	db *Database
}

// NewLogRepository creates a new log repository.
func NewLogRepository(db *Database) *LogRepository {
	return &LogRepository{db: db}
}

// LogFilters contains filter criteria for querying logs.
type LogFilters struct {
	Search   string
	Severity string
	Source   string
	Color    string
	FromDate string
	ToDate   string
	Limit    int
	Offset   int
}

// Create inserts a new log into the database.
func (r *LogRepository) Create(log *entities.Log) error {
	bodyJSON, err := json.Marshal(log.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	result, err := r.db.Conn().Exec(`
		INSERT INTO logs (
			title, severity, source, color, description, body,
			derived_severity, derived_source, derived_category, created_at
		) VALUES (?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?)`,
		log.Header.Title,
		log.Header.Severity.String(),
		log.Header.Source,
		log.Header.Color.String(),
		log.Header.Description,
		string(bodyJSON),
		log.Metadata.DerivedSeverity,
		log.Metadata.DerivedSource,
		log.Metadata.DerivedCategory,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.ID = id
	return nil
}

// FindByID retrieves a single log by ID.
func (r *LogRepository) FindByID(id int64) (*entities.Log, error) {
	query := `
		SELECT id, title, severity, source, color, description, body, created_at,
		       derived_severity, derived_source, derived_category
		FROM logs WHERE id = ?`

	row := r.db.Conn().QueryRow(query, id)
	return r.scanLogRow(row)
}

// FindAll retrieves logs with optional filters.
func (r *LogRepository) FindAll(filters LogFilters) ([]*entities.Log, int, error) {
	// Build dynamic SQL query
	query := `
		SELECT id, title, severity, source, color, description, body, created_at,
		       derived_severity, derived_source, derived_category
		FROM logs WHERE 1=1`
	countQuery := "SELECT COUNT(*) FROM logs WHERE 1=1"
	var args []any
	var countArgs []any

	// Add search filter
	if filters.Search != "" {
		searchClause := " AND (title LIKE ? OR description LIKE ? OR body LIKE ?)"
		searchTerm := "%" + filters.Search + "%"
		query += searchClause
		countQuery += searchClause
		args = append(args, searchTerm, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm, searchTerm)
	}

	// Add severity filter
	if filters.Severity != "" {
		query += " AND severity = ?"
		countQuery += " AND severity = ?"
		args = append(args, filters.Severity)
		countArgs = append(countArgs, filters.Severity)
	}

	// Add source filter
	if filters.Source != "" {
		query += " AND source = ?"
		countQuery += " AND source = ?"
		args = append(args, filters.Source)
		countArgs = append(countArgs, filters.Source)
	}

	// Add color filter
	if filters.Color != "" {
		query += " AND color = ?"
		countQuery += " AND color = ?"
		args = append(args, filters.Color)
		countArgs = append(countArgs, filters.Color)
	}

	// Add date filters
	if filters.FromDate != "" {
		query += " AND created_at >= ?"
		countQuery += " AND created_at >= ?"
		args = append(args, filters.FromDate)
		countArgs = append(countArgs, filters.FromDate)
	}
	if filters.ToDate != "" {
		query += " AND created_at <= ?"
		countQuery += " AND created_at <= ?"
		args = append(args, filters.ToDate)
		countArgs = append(countArgs, filters.ToDate)
	}

	// Get total count
	var totalCount int
	if err := r.db.Conn().QueryRow(countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}
	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	// Execute query
	rows, err := r.db.Conn().Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var logs []*entities.Log
	for rows.Next() {
		log, err := r.scanLog(rows)
		if err != nil {
			continue // Skip malformed rows
		}
		logs = append(logs, log)
	}

	return logs, totalCount, nil
}

// Count returns the total number of logs.
func (r *LogRepository) Count() (int, error) {
	var count int
	err := r.db.Conn().QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}
	return count, nil
}

// CountLast24Hours returns the number of logs from the last 24 hours.
func (r *LogRepository) CountLast24Hours() (int, error) {
	cutoff := time.Now().Add(-24 * time.Hour)
	var count int
	err := r.db.Conn().QueryRow(
		"SELECT COUNT(*) FROM logs WHERE created_at >= ?", cutoff,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent logs: %w", err)
	}
	return count, nil
}

// CountBySeverity returns log counts grouped by effective severity (derived_severity if set, otherwise severity).
func (r *LogRepository) CountBySeverity() (map[string]int, error) {
	rows, err := r.db.Conn().Query(
		"SELECT COALESCE(NULLIF(derived_severity, ''), severity) as effective_severity, COUNT(*) FROM logs GROUP BY effective_severity",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count by severity: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			continue
		}
		counts[severity] = count
	}
	return counts, nil
}

// CountBySource returns log counts grouped by source.
func (r *LogRepository) CountBySource() (map[string]int, error) {
	rows, err := r.db.Conn().Query(
		"SELECT COALESCE(source, 'unknown'), COUNT(*) FROM logs GROUP BY source",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count by source: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			continue
		}
		counts[source] = count
	}
	return counts, nil
}

// Delete removes a log by ID.
func (r *LogRepository) Delete(id int64) error {
	result, err := r.db.Conn().Exec("DELETE FROM logs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrLogNotFound
	}

	return nil
}

// DeleteOlderThan deletes logs older than the specified date.
func (r *LogRepository) DeleteOlderThan(cutoffDate time.Time) (int64, error) {
	result, err := r.db.Conn().Exec(
		"DELETE FROM logs WHERE created_at < ?", cutoffDate,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// scanLog scans a row into a Log entity (for Rows).
func (r *LogRepository) scanLog(rows *sql.Rows) (*entities.Log, error) {
	var log entities.Log
	var bodyJSON string
	var severityStr string
	var source, colorStr, description sql.NullString
	var derivedSeverity, derivedSource, derivedCategory sql.NullString

	err := rows.Scan(
		&log.ID,
		&log.Header.Title,
		&severityStr,
		&source,
		&colorStr,
		&description,
		&bodyJSON,
		&log.CreatedAt,
		&derivedSeverity,
		&derivedSource,
		&derivedCategory,
	)
	if err != nil {
		return nil, err
	}

	log.Header.Severity = valueobjects.SeverityFromString(severityStr)
	log.Header.Source = source.String
	log.Header.Color = valueobjects.ColorFromString(colorStr.String)
	log.Header.Description = description.String
	log.Metadata.DerivedSeverity = derivedSeverity.String
	log.Metadata.DerivedSource = derivedSource.String
	log.Metadata.DerivedCategory = derivedCategory.String

	if bodyJSON != "" {
		if err := json.Unmarshal([]byte(bodyJSON), &log.Body); err != nil {
			log.Body = make(map[string]any)
		}
	} else {
		log.Body = make(map[string]any)
	}

	return &log, nil
}

// scanLogRow scans a single row into a Log entity (for QueryRow).
func (r *LogRepository) scanLogRow(row *sql.Row) (*entities.Log, error) {
	var log entities.Log
	var bodyJSON string
	var severityStr string
	var source, colorStr, description sql.NullString
	var derivedSeverity, derivedSource, derivedCategory sql.NullString

	err := row.Scan(
		&log.ID,
		&log.Header.Title,
		&severityStr,
		&source,
		&colorStr,
		&description,
		&bodyJSON,
		&log.CreatedAt,
		&derivedSeverity,
		&derivedSource,
		&derivedCategory,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrLogNotFound
		}
		return nil, err
	}

	log.Header.Severity = valueobjects.SeverityFromString(severityStr)
	log.Header.Source = source.String
	log.Header.Color = valueobjects.ColorFromString(colorStr.String)
	log.Header.Description = description.String
	log.Metadata.DerivedSeverity = derivedSeverity.String
	log.Metadata.DerivedSource = derivedSource.String
	log.Metadata.DerivedCategory = derivedCategory.String

	if bodyJSON != "" {
		if err := json.Unmarshal([]byte(bodyJSON), &log.Body); err != nil {
			log.Body = make(map[string]any)
		}
	} else {
		log.Body = make(map[string]any)
	}

	return &log, nil
}
