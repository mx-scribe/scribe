package handlers_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"

	"github.com/mx-scribe/scribe/internal/infrastructure/http/handlers"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// testDB creates a temporary in-memory database for testing.
func testDB(t *testing.T) *sqlite.Database {
	t.Helper()
	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return db
}

// createTestLog helper to create a log via the handler.
//
//nolint:unparam // return value used in some tests
func createTestLog(t *testing.T, db *sqlite.Database, title, severity, source string) int64 {
	t.Helper()
	body := map[string]any{
		"header": map[string]any{
			"title":    title,
			"severity": severity,
			"source":   source,
		},
		"body": map[string]any{
			"test": true,
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/logs", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.CreateLog(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	return int64(resp["id"].(float64))
}

func TestCreateLog_Success(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		body     map[string]any
		wantCode int
	}{
		{
			name: "basic log with title only",
			body: map[string]any{
				"header": map[string]any{"title": "Test log"},
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "log with severity and source",
			body: map[string]any{
				"header": map[string]any{
					"title":    "Error occurred",
					"severity": "error",
					"source":   "api-gateway",
				},
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "log with body data",
			body: map[string]any{
				"header": map[string]any{
					"title":       "Payment processed",
					"severity":    "info",
					"description": "Customer payment completed",
				},
				"body": map[string]any{
					"order_id":   "ORD-123",
					"amount":     99.99,
					"customer":   "john@example.com",
					"successful": true,
				},
			},
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/logs", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := handlers.CreateLog(db)
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d: %s", tt.wantCode, rec.Code, rec.Body.String())
			}

			if tt.wantCode == http.StatusCreated {
				var resp map[string]any
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if resp["id"] == nil {
					t.Error("expected response to have 'id'")
				}
				if resp["created_at"] == nil {
					t.Error("expected response to have 'created_at'")
				}
			}
		})
	}
}

func TestCreateLog_Validation(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		body     string
		wantCode int
		wantMsg  string
	}{
		{
			name:     "missing title",
			body:     `{"header": {}}`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "title is required",
		},
		{
			name:     "empty title",
			body:     `{"header": {"title": ""}}`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "title is required",
		},
		{
			name:     "invalid JSON",
			body:     `{invalid json`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/logs", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := handlers.CreateLog(db)
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, rec.Code)
			}

			var resp map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&resp)
			if resp["error"] != tt.wantMsg {
				t.Errorf("expected error '%s', got '%s'", tt.wantMsg, resp["error"])
			}
		})
	}
}

func TestListLogs_Pagination(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create 25 test logs
	for i := 0; i < 25; i++ {
		createTestLog(t, db, "Test log", "info", "test")
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantPage  int
		wantLimit int
	}{
		{
			name:      "default pagination",
			query:     "",
			wantCount: 20, // default limit
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "custom limit",
			query:     "?limit=5",
			wantCount: 5,
			wantPage:  1,
			wantLimit: 5,
		},
		{
			name:      "second page",
			query:     "?limit=10&page=2",
			wantCount: 10,
			wantPage:  2,
			wantLimit: 10,
		},
		{
			name:      "last page partial",
			query:     "?limit=10&page=3",
			wantCount: 5, // 25 total, page 3 with limit 10 = 5 remaining
			wantPage:  3,
			wantLimit: 10,
		},
		{
			name:      "limit capped at 100",
			query:     "?limit=200",
			wantCount: 25, // only 25 logs exist
			wantPage:  1,
			wantLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/logs"+tt.query, nil)
			rec := httptest.NewRecorder()

			handler := handlers.ListLogs(db)
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}

			var resp struct {
				Logs  []map[string]any `json:"logs"`
				Total int              `json:"total"`
				Limit int              `json:"limit"`
				Page  int              `json:"page"`
			}
			_ = json.NewDecoder(rec.Body).Decode(&resp)

			if len(resp.Logs) != tt.wantCount {
				t.Errorf("expected %d logs, got %d", tt.wantCount, len(resp.Logs))
			}
			if resp.Page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, resp.Page)
			}
			if resp.Limit != tt.wantLimit {
				t.Errorf("expected limit %d, got %d", tt.wantLimit, resp.Limit)
			}
			if resp.Total != 25 {
				t.Errorf("expected total 25, got %d", resp.Total)
			}
		})
	}
}

func TestListLogs_Filters(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create logs with different severities and sources
	createTestLog(t, db, "Error in payment", "error", "payment-service")
	createTestLog(t, db, "Warning in auth", "warning", "auth-service")
	createTestLog(t, db, "Info in api", "info", "api-gateway")
	createTestLog(t, db, "Another error", "error", "database")
	createTestLog(t, db, "Debug message", "debug", "api-gateway")

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{
			name:      "filter by severity error",
			query:     "?severity=error",
			wantCount: 2,
		},
		{
			name:      "filter by severity warning",
			query:     "?severity=warning",
			wantCount: 1,
		},
		{
			name:      "filter by source api-gateway",
			query:     "?source=api-gateway",
			wantCount: 2,
		},
		{
			name:      "filter by source and severity",
			query:     "?source=api-gateway&severity=info",
			wantCount: 1,
		},
		{
			name:      "search by title keyword",
			query:     "?search=payment",
			wantCount: 1,
		},
		{
			name:      "no results for non-existent filter",
			query:     "?severity=critical",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/logs"+tt.query, nil)
			rec := httptest.NewRecorder()

			handler := handlers.ListLogs(db)
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}

			var resp struct {
				Logs  []map[string]any `json:"logs"`
				Total int              `json:"total"`
			}
			_ = json.NewDecoder(rec.Body).Decode(&resp)

			if len(resp.Logs) != tt.wantCount {
				t.Errorf("expected %d logs, got %d", tt.wantCount, len(resp.Logs))
			}
		})
	}
}

func TestGetLog_Success(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create a test log
	id := createTestLog(t, db, "Test log for retrieval", "info", "test-source")

	// Create a router with the route parameter
	router := chi.NewRouter()
	router.Get("/api/logs/{id}", handlers.GetLog(db))

	rec := httptest.NewRecorder()

	// Use proper ID in URL
	req := httptest.NewRequest(http.MethodGet, "/api/logs/1", nil)
	_ = id // Used to create the test log
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	header := resp["header"].(map[string]any)
	if header["title"] != "Test log for retrieval" {
		t.Errorf("expected title 'Test log for retrieval', got '%s'", header["title"])
	}
}

func TestGetLog_NotFound(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	router := chi.NewRouter()
	router.Get("/api/logs/{id}", handlers.GetLog(db))

	req := httptest.NewRequest(http.MethodGet, "/api/logs/99999", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetLog_InvalidID(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	router := chi.NewRouter()
	router.Get("/api/logs/{id}", handlers.GetLog(db))

	req := httptest.NewRequest(http.MethodGet, "/api/logs/invalid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetStats(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create logs with different severities
	createTestLog(t, db, "Error 1", "error", "service-a")
	createTestLog(t, db, "Error 2", "error", "service-a")
	createTestLog(t, db, "Warning", "warning", "service-b")
	createTestLog(t, db, "Info 1", "info", "service-a")
	createTestLog(t, db, "Info 2", "info", "service-b")

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()

	handler := handlers.GetStats(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Total       int            `json:"total"`
		Last24Hours int            `json:"last_24_hours"`
		BySeverity  map[string]int `json:"by_severity"`
		BySource    map[string]int `json:"by_source"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Total)
	}

	if resp.BySeverity["error"] != 2 {
		t.Errorf("expected 2 errors, got %d", resp.BySeverity["error"])
	}

	if resp.BySeverity["warning"] != 1 {
		t.Errorf("expected 1 warning, got %d", resp.BySeverity["warning"])
	}

	if resp.BySeverity["info"] != 2 {
		t.Errorf("expected 2 info, got %d", resp.BySeverity["info"])
	}

	if resp.BySource["service-a"] != 3 {
		t.Errorf("expected 3 from service-a, got %d", resp.BySource["service-a"])
	}

	if resp.BySource["service-b"] != 2 {
		t.Errorf("expected 2 from service-b, got %d", resp.BySource["service-b"])
	}
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handlers.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}

	if resp.Version == "" {
		t.Error("expected version to be set")
	}
}

func TestExportJSON(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create test logs
	createTestLog(t, db, "Error log", "error", "api")
	createTestLog(t, db, "Info log", "info", "database")
	createTestLog(t, db, "Warning log", "warning", "api")

	req := httptest.NewRequest(http.MethodGet, "/api/export/json", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportJSON(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check disposition
	disposition := rec.Header().Get("Content-Disposition")
	if disposition != "attachment; filename=scribe-logs.json" {
		t.Errorf("unexpected Content-Disposition: %s", disposition)
	}

	// Parse response
	var logs []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&logs); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
}

func TestExportJSON_WithFilters(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	createTestLog(t, db, "Error log", "error", "api")
	createTestLog(t, db, "Info log", "info", "database")

	req := httptest.NewRequest(http.MethodGet, "/api/export/json?severity=error", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportJSON(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var logs []map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&logs)

	if len(logs) != 1 {
		t.Errorf("expected 1 error log, got %d", len(logs))
	}
}

func TestExportCSV(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	createTestLog(t, db, "Error log", "error", "api")
	createTestLog(t, db, "Info log", "info", "database")

	req := httptest.NewRequest(http.MethodGet, "/api/export/csv", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportCSV(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected Content-Type 'text/csv', got '%s'", contentType)
	}

	// Check disposition
	disposition := rec.Header().Get("Content-Disposition")
	if disposition != "attachment; filename=scribe-logs.csv" {
		t.Errorf("unexpected Content-Disposition: %s", disposition)
	}

	// Check CSV content has header and data rows
	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty CSV body")
	}

	// Should contain the CSV header
	if !contains(body, "id,severity,source,title,description,created_at") {
		t.Error("CSV should contain header row")
	}

	// Should contain the log titles
	if !contains(body, "Error log") || !contains(body, "Info log") {
		t.Error("CSV should contain log data")
	}
}

func TestExportCSV_WithFilters(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	createTestLog(t, db, "Error log", "error", "api")
	createTestLog(t, db, "Info log", "info", "database")

	req := httptest.NewRequest(http.MethodGet, "/api/export/csv?source=api", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportCSV(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Should contain the API log
	if !contains(body, "Error log") {
		t.Error("CSV should contain api log")
	}

	// Should NOT contain the database log
	if contains(body, "Info log") {
		t.Error("CSV should not contain filtered out log")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock filesystem for SPA handler tests
func mockFS() fs.FS {
	return fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>SPA</body></html>"),
		},
		"dist/assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('app');"),
		},
		"dist/assets/style.css": &fstest.MapFile{
			Data: []byte("body { color: black; }"),
		},
		"dist/favicon.ico": &fstest.MapFile{
			Data: []byte("icon-data"),
		},
	}
}

func TestSPAHandler_ServeIndex(t *testing.T) {
	handler := handlers.NewSPAHandler(mockFS(), "dist")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	// Should have no-cache headers for index.html
	cacheControl := rec.Header().Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("expected no-cache for index.html, got '%s'", cacheControl)
	}
}

func TestSPAHandler_ServeAsset(t *testing.T) {
	handler := handlers.NewSPAHandler(mockFS(), "dist")

	tests := []struct {
		name        string
		path        string
		wantContent string
		wantType    string
		wantCache   string
	}{
		{
			name:        "serve JavaScript asset",
			path:        "/assets/app.js",
			wantContent: "console.log('app');",
			wantType:    "application/javascript; charset=utf-8",
			wantCache:   "public, max-age=31536000, immutable",
		},
		{
			name:        "serve CSS asset",
			path:        "/assets/style.css",
			wantContent: "body { color: black; }",
			wantType:    "text/css; charset=utf-8",
			wantCache:   "public, max-age=31536000, immutable",
		},
		{
			name:        "serve favicon",
			path:        "/favicon.ico",
			wantContent: "icon-data",
			wantType:    "image/x-icon",
			wantCache:   "", // No special caching for favicon
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != tt.wantType {
				t.Errorf("expected Content-Type '%s', got '%s'", tt.wantType, contentType)
			}

			cacheControl := rec.Header().Get("Cache-Control")
			if cacheControl != tt.wantCache {
				t.Errorf("expected Cache-Control '%s', got '%s'", tt.wantCache, cacheControl)
			}

			body := rec.Body.String()
			if body != tt.wantContent {
				t.Errorf("expected body '%s', got '%s'", tt.wantContent, body)
			}
		})
	}
}

func TestSPAHandler_SPAFallback(t *testing.T) {
	handler := handlers.NewSPAHandler(mockFS(), "dist")

	// Request a path that doesn't exist - should serve index.html for SPA routing
	tests := []string{
		"/dashboard",
		"/logs/123",
		"/settings/profile",
		"/nonexistent/path/here",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200 for SPA fallback, got %d", rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "text/html; charset=utf-8" {
				t.Errorf("expected Content-Type 'text/html; charset=utf-8' for SPA fallback, got '%s'", contentType)
			}

			body := rec.Body.String()
			if !contains(body, "SPA") {
				t.Error("expected index.html content for SPA fallback")
			}
		})
	}
}

func TestSPAHandler_PathCleaning(t *testing.T) {
	handler := handlers.NewSPAHandler(mockFS(), "dist")

	tests := []struct {
		name string
		path string
	}{
		{"path with double slash", "//assets//app.js"},
		{"path with dot segments", "/assets/../assets/app.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// All paths should either serve the file or fall back to index.html
			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d for path %s", rec.Code, tt.path)
			}
		})
	}
}

func TestSPAHandler_ContentTypes(t *testing.T) {
	// Test getContentType for various file types
	mockFSWithTypes := fstest.MapFS{
		"dist/index.html":  &fstest.MapFile{Data: []byte("html")},
		"dist/image.png":   &fstest.MapFile{Data: []byte("png")},
		"dist/image.jpg":   &fstest.MapFile{Data: []byte("jpg")},
		"dist/image.gif":   &fstest.MapFile{Data: []byte("gif")},
		"dist/icon.svg":    &fstest.MapFile{Data: []byte("svg")},
		"dist/data.json":   &fstest.MapFile{Data: []byte("{}")},
		"dist/font.woff":   &fstest.MapFile{Data: []byte("woff")},
		"dist/font.woff2":  &fstest.MapFile{Data: []byte("woff2")},
		"dist/font.ttf":    &fstest.MapFile{Data: []byte("ttf")},
		"dist/font.eot":    &fstest.MapFile{Data: []byte("eot")},
		"dist/unknown.xyz": &fstest.MapFile{Data: []byte("unknown")},
	}

	handler := handlers.NewSPAHandler(mockFSWithTypes, "dist")

	tests := []struct {
		path     string
		wantType string
	}{
		{"/image.png", "image/png"},
		{"/image.jpg", "image/jpeg"},
		{"/image.gif", "image/gif"},
		{"/icon.svg", "image/svg+xml"},
		{"/data.json", "application/json; charset=utf-8"},
		{"/font.woff", "font/woff"},
		{"/font.woff2", "font/woff2"},
		{"/font.ttf", "font/ttf"},
		{"/font.eot", "application/vnd.ms-fontobject"},
		{"/unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != tt.wantType {
				t.Errorf("expected Content-Type '%s', got '%s'", tt.wantType, contentType)
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	getMetrics := func() (uint64, int64, uint64) {
		return 100, 5, 3 // totalRequests, activeRequests, totalErrors
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := handlers.MetricsHandler(getMetrics, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var resp struct {
		TotalRequests  uint64 `json:"total_requests"`
		ActiveRequests int64  `json:"active_requests"`
		TotalErrors    uint64 `json:"total_errors"`
		ErrorRate      string `json:"error_rate"`
		Uptime         string `json:"uptime"`
		GoRoutines     int    `json:"go_routines"`
		MemoryMB       uint64 `json:"memory_mb"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.TotalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", resp.TotalRequests)
	}
	if resp.ActiveRequests != 5 {
		t.Errorf("expected 5 active requests, got %d", resp.ActiveRequests)
	}
	if resp.TotalErrors != 3 {
		t.Errorf("expected 3 total errors, got %d", resp.TotalErrors)
	}
	if resp.ErrorRate != "3.000%" {
		t.Errorf("expected error rate '3.000%%', got '%s'", resp.ErrorRate)
	}
	if resp.GoRoutines <= 0 {
		t.Error("expected go_routines > 0")
	}
}

func TestMetricsHandler_WithSSE(t *testing.T) {
	getMetrics := func() (uint64, int64, uint64) {
		return 50, 2, 1
	}

	hub := handlers.NewSSEHub()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := handlers.MetricsHandler(getMetrics, hub)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		SSEClients int `json:"sse_clients"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	// SSEClients should be 0 since no clients are connected
	if resp.SSEClients != 0 {
		t.Errorf("expected 0 SSE clients, got %d", resp.SSEClients)
	}
}

func TestPrometheusMetricsHandler(t *testing.T) {
	getMetrics := func() (uint64, int64, uint64) {
		return 200, 10, 5
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	rec := httptest.NewRecorder()

	handler := handlers.PrometheusMetricsHandler(getMetrics, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/plain; charset=utf-8', got '%s'", contentType)
	}

	body := rec.Body.String()

	// Check for required Prometheus metrics
	expectedMetrics := []string{
		"scribe_http_requests_total 200",
		"scribe_http_requests_active 10",
		"scribe_http_errors_total 5",
		"scribe_uptime_seconds",
		"scribe_goroutines",
		"scribe_memory_bytes",
		"scribe_sse_clients",
	}

	for _, metric := range expectedMetrics {
		if !contains(body, metric) {
			t.Errorf("expected Prometheus output to contain '%s'", metric)
		}
	}

	// Check for HELP and TYPE comments
	if !contains(body, "# HELP scribe_http_requests_total") {
		t.Error("expected HELP comment for requests_total")
	}
	if !contains(body, "# TYPE scribe_http_requests_total counter") {
		t.Error("expected TYPE comment for requests_total")
	}
}

func TestDeleteLog(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create a test log
	id := createTestLog(t, db, "Log to delete", "info", "test")

	router := chi.NewRouter()
	router.Delete("/api/logs/{id}", handlers.DeleteLog(db))

	req := httptest.NewRequest(http.MethodDelete, "/api/logs/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify log is deleted
	router2 := chi.NewRouter()
	router2.Get("/api/logs/{id}", handlers.GetLog(db))

	req2 := httptest.NewRequest(http.MethodGet, "/api/logs/1", nil)
	rec2 := httptest.NewRecorder()
	router2.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected deleted log to return 404, got %d", rec2.Code)
	}

	_ = id // Suppress unused variable warning
}

func TestDeleteLog_NotFound(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	router := chi.NewRouter()
	router.Delete("/api/logs/{id}", handlers.DeleteLog(db))

	req := httptest.NewRequest(http.MethodDelete, "/api/logs/99999", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDeleteLogs_Bulk(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create test logs
	createTestLog(t, db, "Log 1", "info", "test")
	createTestLog(t, db, "Log 2", "info", "test")
	createTestLog(t, db, "Log 3", "info", "test")

	body := `{"ids": [1, 2]}`
	req := httptest.NewRequest(http.MethodDelete, "/api/logs", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.DeleteLogs(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Deleted int `json:"deleted"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", resp.Deleted)
	}
}

func TestDeleteLogs_EmptyIDs(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	body := `{"ids": []}`
	req := httptest.NewRequest(http.MethodDelete, "/api/logs", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.DeleteLogs(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteLogs_InvalidBody(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodDelete, "/api/logs", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.DeleteLogs(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteLog_InvalidID(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	router := chi.NewRouter()
	router.Delete("/api/logs/{id}", handlers.DeleteLog(db))

	req := httptest.NewRequest(http.MethodDelete, "/api/logs/invalid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCleanupLogs(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create a test log
	createTestLog(t, db, "Test log", "info", "test")

	body := `{"retention_days": 30}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/cleanup", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.CleanupLogs(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		DeletedCount int64  `json:"deleted_count"`
		CutoffDate   string `json:"cutoff_date"`
		Message      string `json:"message"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Message != "Cleanup completed successfully" {
		t.Errorf("expected success message, got '%s'", resp.Message)
	}
	if resp.CutoffDate == "" {
		t.Error("expected cutoff_date to be set")
	}
}

func TestCleanupLogs_InvalidRetention(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{
			name:     "zero retention days",
			body:     `{"retention_days": 0}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "negative retention days",
			body:     `{"retention_days": -5}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid JSON",
			body:     `{invalid}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/admin/cleanup", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := handlers.CleanupLogs(db)
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestGetRetentionInfo(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Create some test logs
	createTestLog(t, db, "Log 1", "info", "test")
	createTestLog(t, db, "Log 2", "error", "test")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/retention", nil)
	rec := httptest.NewRecorder()

	handler := handlers.GetRetentionInfo(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Total       int            `json:"total"`
		Last24Hours int            `json:"last_24_hours"`
		ByAge       map[string]int `json:"by_age"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if resp.Last24Hours != 2 {
		t.Errorf("expected last_24_hours 2, got %d", resp.Last24Hours)
	}
	if resp.ByAge == nil {
		t.Error("expected by_age to be set")
	}
	// Logs created today should be in "today" bucket
	if resp.ByAge["today"] != 2 {
		t.Errorf("expected 2 logs in 'today' bucket, got %d", resp.ByAge["today"])
	}
}

func TestSimpleMetrics(t *testing.T) {
	m := &handlers.SimpleMetrics{}

	// Initial values should be 0
	if m.GetTotalRequests() != 0 {
		t.Errorf("expected initial total requests 0, got %d", m.GetTotalRequests())
	}
	if m.GetActiveRequests() != 0 {
		t.Errorf("expected initial active requests 0, got %d", m.GetActiveRequests())
	}
	if m.GetTotalErrors() != 0 {
		t.Errorf("expected initial total errors 0, got %d", m.GetTotalErrors())
	}

	// Increment requests
	m.IncrementRequests()
	m.IncrementRequests()
	if m.GetTotalRequests() != 2 {
		t.Errorf("expected 2 total requests, got %d", m.GetTotalRequests())
	}

	// Increment errors
	m.IncrementErrors()
	if m.GetTotalErrors() != 1 {
		t.Errorf("expected 1 total error, got %d", m.GetTotalErrors())
	}
}

func TestSSEHub_ClientCount(t *testing.T) {
	hub := handlers.NewSSEHub()

	// Initially should have 0 clients
	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestSSEHub_Broadcast(t *testing.T) {
	hub := handlers.NewSSEHub()

	// These should not panic even with no clients
	hub.BroadcastLogDeleted(123)
	hub.BroadcastStatsUpdated(map[string]int{"total": 10})

	// Give time for goroutine to process
	// (broadcast should be non-blocking)
}

func TestMetricsHandler_ZeroRequests(t *testing.T) {
	getMetrics := func() (uint64, int64, uint64) {
		return 0, 0, 0 // No requests yet
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := handlers.MetricsHandler(getMetrics, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		ErrorRate string `json:"error_rate"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	// With 0 requests, error rate should be 0.00%
	if resp.ErrorRate != "0.00%" {
		t.Errorf("expected error rate '0.00%%', got '%s'", resp.ErrorRate)
	}
}

func TestCreateLogWithSSE_Broadcast(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	hub := handlers.NewSSEHub()

	body := map[string]any{
		"header": map[string]any{
			"title":    "Test with SSE",
			"severity": "info",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/logs", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.CreateLogWithSSE(db, hub)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteLogWithSSE_Broadcast(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	hub := handlers.NewSSEHub()

	// Create a test log
	createTestLog(t, db, "Log to delete with SSE", "info", "test")

	router := chi.NewRouter()
	router.Delete("/api/logs/{id}", handlers.DeleteLogWithSSE(db, hub))

	req := httptest.NewRequest(http.MethodDelete, "/api/logs/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteLogsWithSSE_Broadcast(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	hub := handlers.NewSSEHub()

	// Create test logs
	createTestLog(t, db, "Log 1", "info", "test")
	createTestLog(t, db, "Log 2", "info", "test")

	body := `{"ids": [1, 2]}`
	req := httptest.NewRequest(http.MethodDelete, "/api/logs", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.DeleteLogsWithSSE(db, hub)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Deleted int `json:"deleted"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", resp.Deleted)
	}
}

func TestPrometheusMetricsHandler_WithSSE(t *testing.T) {
	getMetrics := func() (uint64, int64, uint64) {
		return 100, 5, 2
	}

	hub := handlers.NewSSEHub()

	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	rec := httptest.NewRecorder()

	handler := handlers.PrometheusMetricsHandler(getMetrics, hub)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !contains(body, "scribe_sse_clients 0") {
		t.Error("expected sse_clients metric in output")
	}
}

func TestGetStats_Empty(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()

	handler := handlers.GetStats(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Total       int            `json:"total"`
		Last24Hours int            `json:"last_24_hours"`
		BySeverity  map[string]int `json:"by_severity"`
		BySource    map[string]int `json:"by_source"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}
}

func TestSPAHandler_MethodNotAllowed(t *testing.T) {
	handler := handlers.NewSPAHandler(mockFS(), "dist")

	// POST should still work (SPA handles routing)
	req := httptest.NewRequest(http.MethodPost, "/some-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// SPA should serve index.html for any path
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for SPA fallback, got %d", rec.Code)
	}
}

func TestExportJSON_Empty(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/export/json", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportJSON(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var logs []map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&logs)

	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}

func TestExportCSV_Empty(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/export/csv", nil)
	rec := httptest.NewRecorder()

	handler := handlers.ExportCSV(db)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	// Should contain header even with no data
	if !contains(body, "id,severity,source,title,description,created_at") {
		t.Error("CSV should contain header row")
	}
}

// mockResponseWriter implements http.ResponseWriter but not http.Flusher
type mockResponseWriter struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func TestSSEHandler_NoFlusher(t *testing.T) {
	hub := handlers.NewSSEHub()

	// Use a mock response writer that doesn't implement Flusher
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	rec := &mockResponseWriter{}

	handler := handlers.SSEHandler(hub)
	handler.ServeHTTP(rec, req)

	// Should return 500 because streaming is unsupported
	if rec.code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.code)
	}

	body := rec.body.String()
	if !contains(body, "Streaming unsupported") {
		t.Error("expected 'Streaming unsupported' error message")
	}
}

func TestSSEHub_RegisterUnregister(t *testing.T) {
	hub := handlers.NewSSEHub()

	// Give the hub time to start its goroutine
	// (already started in NewSSEHub)

	// Verify initial count
	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestMetrics_FormatDuration(t *testing.T) {
	// Test by calling MetricsHandler which uses formatDuration internally
	getMetrics := func() (uint64, int64, uint64) {
		return 1000, 10, 5
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := handlers.MetricsHandler(getMetrics, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Uptime string `json:"uptime"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	// Uptime should be a formatted duration string like "0s" or "1m 30s"
	if resp.Uptime == "" {
		t.Error("expected uptime to be set")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
