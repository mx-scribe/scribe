package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/mx-scribe/scribe/internal/application/commands"
	"github.com/mx-scribe/scribe/internal/domain/services"
	"github.com/mx-scribe/scribe/internal/infrastructure/http/handlers"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// TestServer represents a test server setup for integration tests.
type TestServer struct {
	db     *sqlite.Database
	router *chi.Mux
}

// setupTestServer creates a complete test server with all dependencies.
func setupTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create in-memory database
	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create router
	router := chi.NewRouter()

	// Setup routes
	router.Get("/health", handlers.Health)
	router.Post("/api/logs", handlers.CreateLog(db))
	router.Get("/api/logs", handlers.ListLogs(db))
	router.Get("/api/logs/{id}", handlers.GetLog(db))
	router.Delete("/api/logs/{id}", handlers.DeleteLog(db))
	router.Get("/api/stats", handlers.GetStats(db))
	router.Get("/api/export/json", handlers.ExportJSON(db))
	router.Get("/api/export/csv", handlers.ExportCSV(db))

	return &TestServer{
		db:     db,
		router: router,
	}
}

func (ts *TestServer) Close() {
	ts.db.Close()
}

// TestEndToEnd_CompleteLogLifecycle tests the complete lifecycle of a log.
func TestEndToEnd_CompleteLogLifecycle(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// 1. Check health
	t.Run("Health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		if result["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%v'", result["status"])
		}
	})

	// 2. Get initial stats (empty database)
	t.Run("Initial stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/stats", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		if result["total"].(float64) != 0 {
			t.Errorf("Expected total 0, got %v", result["total"])
		}
	})

	var createdLogID float64

	// 3. Create a log
	t.Run("Create log", func(t *testing.T) {
		logData := map[string]any{
			"header": map[string]any{
				"title":    "Database connection failed",
				"severity": "error",
				"source":   "api-service",
			},
			"body": map[string]any{
				"error":    "connection timeout",
				"database": "main",
			},
		}

		body, _ := json.Marshal(logData)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		createdLogID = result["id"].(float64)
		if createdLogID == 0 {
			t.Error("Expected log ID to be set")
		}
	})

	// 4. List logs
	t.Run("List logs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		logs := result["logs"].([]any)
		if len(logs) != 1 {
			t.Errorf("Expected 1 log, got %d", len(logs))
		}
	})

	// 5. Get stats after creating log
	t.Run("Stats after create", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/stats", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		if result["total"].(float64) != 1 {
			t.Errorf("Expected total 1, got %v", result["total"])
		}
	})

	// 6. Export JSON
	t.Run("Export JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/export/json", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
	})

	// 7. Export CSV
	t.Run("Export CSV", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/export/csv", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/csv" {
			t.Errorf("Expected Content-Type text/csv, got %s", contentType)
		}
	})

	// 8. Delete log
	t.Run("Delete log", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/logs/1", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 9. Verify deletion
	t.Run("Verify deletion", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/stats", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		if result["total"].(float64) != 0 {
			t.Errorf("Expected total 0 after deletion, got %v", result["total"])
		}
	})
}

// TestPatternMatching tests automatic severity/source derivation.
func TestPatternMatching(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	tests := []struct {
		name             string
		title            string
		expectedSeverity string
	}{
		{"Error pattern", "Database connection failed", "error"},
		{"Success pattern", "Payment completed successfully", "success"},
		{"Warning pattern", "Deprecated API call detected", "warning"},
		{"HTTP 500 pattern", "HTTP 500 Internal Server Error", "error"},
		{"HTTP 404 pattern", "HTTP 404 Not Found", "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logData := map[string]any{
				"header": map[string]any{
					"title": tt.title,
				},
			}

			body, _ := json.Marshal(logData)
			req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			server.router.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
				return
			}

			var result map[string]any
			_ = json.NewDecoder(rec.Body).Decode(&result)

			// CreateLog returns a flat structure: {id, title, severity, created_at}
			severity, ok := result["severity"].(string)
			if !ok {
				t.Errorf("Expected severity field in response, got: %v", result)
				return
			}
			if severity != tt.expectedSeverity {
				t.Errorf("Expected severity %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

// TestLogFiltering tests log filtering capabilities.
func TestLogFiltering(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Create test logs
	logs := []map[string]any{
		{"header": map[string]any{"title": "Error log 1", "severity": "error", "source": "api"}},
		{"header": map[string]any{"title": "Error log 2", "severity": "error", "source": "database"}},
		{"header": map[string]any{"title": "Info log 1", "severity": "info", "source": "api"}},
		{"header": map[string]any{"title": "Warning log 1", "severity": "warning", "source": "auth"}},
	}

	for _, logData := range logs {
		body, _ := json.Marshal(logData)
		req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
	}

	// Test severity filter
	t.Run("Filter by severity", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?severity=error", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		logs := result["logs"].([]any)
		if len(logs) != 2 {
			t.Errorf("Expected 2 error logs, got %d", len(logs))
		}
	})

	// Test source filter
	t.Run("Filter by source", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?source=api", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		logs := result["logs"].([]any)
		if len(logs) != 2 {
			t.Errorf("Expected 2 api logs, got %d", len(logs))
		}
	})

	// Test search
	t.Run("Search logs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?search=Error", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		logs := result["logs"].([]any)
		if len(logs) != 2 {
			t.Errorf("Expected 2 logs with 'Error', got %d", len(logs))
		}
	})

	// Test limit
	t.Run("Limit results", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/logs?limit=2", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		var result map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&result)
		logs := result["logs"].([]any)
		if len(logs) != 2 {
			t.Errorf("Expected 2 logs with limit, got %d", len(logs))
		}
	})
}

// TestDomainServices tests domain services directly.
func TestDomainServices(t *testing.T) {
	t.Run("PatternMatcher", func(t *testing.T) {
		matcher := services.NewPatternMatcher()

		// PatternMatcher.AnalyzeLog returns metadata with derived severity
		// We test by creating logs and checking the metadata
		if matcher == nil {
			t.Error("NewPatternMatcher should return non-nil matcher")
		}
	})

	t.Run("SourceDeriver", func(t *testing.T) {
		deriver := services.NewSourceDeriver()

		// SourceDeriver.DeriveSource derives source from log
		if deriver == nil {
			t.Error("NewSourceDeriver should return non-nil deriver")
		}
	})
}

// TestCommandHandler tests the command handler directly.
func TestCommandHandler(t *testing.T) {
	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	repo := sqlite.NewLogRepository(db)
	handler := commands.NewCreateLogHandler(repo)

	input := commands.CreateLogInput{
		Title:       "Test log",
		Severity:    "info",
		Source:      "test",
		Description: "A test log entry",
	}

	output, err := handler.Handle(input)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	if output.ID == 0 {
		t.Error("Expected log ID to be set")
	}
	if output.Title != "Test log" {
		t.Errorf("Expected title 'Test log', got '%s'", output.Title)
	}
}
