package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mx-scribe/scribe/internal/infrastructure/http/handlers"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupTestServer(t *testing.T) *Server {
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

	// Create server
	server := NewServer(db)
	return server
}

func TestRoutes_HealthEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRoutes_MetricsEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRoutes_PrometheusMetricsEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	req := httptest.NewRequest("GET", "/metrics/prometheus", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/plain; charset=utf-8', got '%s'", contentType)
	}
}

func TestRoutes_LogsEndpoints(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET logs",
			method:     "GET",
			path:       "/api/logs",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET single log (not found)",
			method:     "GET",
			path:       "/api/logs/999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "POST logs (bad request - no body)",
			method:     "POST",
			path:       "/api/logs",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "DELETE single log (not found)",
			method:     "DELETE",
			path:       "/api/logs/999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRoutes_StatsEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRoutes_ExportEndpoints(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	tests := []struct {
		name            string
		path            string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "JSON export",
			path:            "/api/export/json",
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "CSV export",
			path:            "/api/export/csv",
			wantStatus:      http.StatusOK,
			wantContentType: "text/csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != tt.wantContentType {
				t.Errorf("Expected Content-Type '%s', got '%s'", tt.wantContentType, contentType)
			}
		})
	}
}

func TestRoutes_AdminEndpoints(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	t.Run("GET retention info", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/admin/retention", nil)
		rec := httptest.NewRecorder()

		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("POST cleanup (bad request - no body)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/admin/cleanup", nil)
		rec := httptest.NewRecorder()

		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})
}

func TestRoutes_CORSHeaders(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	apiEndpoints := []string{
		"/api/logs",
		"/api/stats",
		"/api/export/csv",
		"/api/export/json",
	}

	for _, endpoint := range apiEndpoints {
		t.Run("CORS on "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			// Check for CORS headers
			if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("Expected CORS header on %s, got '%s'", endpoint, origin)
			}
		})
	}
}

func TestRoutes_PreflightRequest(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	req := httptest.NewRequest("OPTIONS", "/api/logs", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	// Should return 204 No Content for preflight
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", rec.Code)
	}

	// CORS headers should be set
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Error("CORS headers not set for OPTIONS request")
	}
}

func TestRoutes_SSEEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	// Create a request with a short timeout context to prevent blocking
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/api/events", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	// This will timeout quickly due to context, but we can verify the route exists
	server.router.ServeHTTP(rec, req)

	// The endpoint should exist (not 404) - it may return other codes due to timeout
	if rec.Code == http.StatusNotFound {
		t.Error("SSE endpoint /api/events should exist")
	}
}

func TestRoutes_AllRoutesRegistered(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	// Note: /api/events is tested separately in TestRoutes_SSEEndpoint
	// because it requires a timeout context to avoid blocking
	expectedRoutes := []string{
		"/health",
		"/metrics",
		"/metrics/prometheus",
		"/api/logs",
		"/api/stats",
		"/api/export/csv",
		"/api/export/json",
		"/api/admin/retention",
	}

	for _, route := range expectedRoutes {
		t.Run("Route registered: "+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotFound {
				t.Errorf("Route %s should be registered, got 404", route)
			}
		})
	}
}

func TestRoutes_ContentTypeJSON(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	jsonEndpoints := []string{
		"/api/logs",
		"/api/stats",
		"/health",
		"/metrics",
	}

	for _, endpoint := range jsonEndpoints {
		t.Run("JSON Content-Type on "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json' on %s, got '%s'", endpoint, contentType)
			}
		})
	}
}

func TestSSEHub_Integration(t *testing.T) {
	server := setupTestServer(t)
	defer server.db.Close()

	// SSEHub should be initialized
	if server.sseHub == nil {
		t.Error("SSEHub should be initialized")
	}

	// Should have 0 clients initially
	if server.sseHub.ClientCount() != 0 {
		t.Errorf("Expected 0 SSE clients, got %d", server.sseHub.ClientCount())
	}
}

func TestNewSSEHub(t *testing.T) {
	hub := handlers.NewSSEHub()

	if hub == nil {
		t.Fatal("NewSSEHub() returned nil")
	}

	if hub.ClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", hub.ClientCount())
	}
}
