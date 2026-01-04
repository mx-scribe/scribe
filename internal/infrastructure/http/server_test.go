package http

import (
	"net/http"
	"testing"

	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

func setupServerTest(t *testing.T) (*Server, *sqlite.Database) {
	t.Helper()

	db, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := sqlite.RunMigrations(db.Conn()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	server := NewServer(db)
	return server, db
}

func TestNewServer(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.router == nil {
		t.Fatal("Server router should be initialized")
	}

	if server.db == nil {
		t.Fatal("Server db should be initialized")
	}

	if server.sseHub == nil {
		t.Fatal("Server sseHub should be initialized")
	}
}

func TestServer_Router(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	router := server.Router()

	if router == nil {
		t.Fatal("Router() returned nil")
	}

	// Router should be the same instance
	if router != server.router {
		t.Error("Router() should return the same router instance")
	}
}

func TestServer_DB(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	dbResult := server.DB()

	if dbResult == nil {
		t.Fatal("DB() returned nil")
	}

	// DB should be the same instance
	if dbResult != db {
		t.Error("DB() should return the same database instance")
	}
}

func TestServer_SSEHub(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	hub := server.SSEHub()

	if hub == nil {
		t.Fatal("SSEHub() returned nil")
	}

	// SSEHub should be the same instance
	if hub != server.sseHub {
		t.Error("SSEHub() should return the same hub instance")
	}
}

func TestServer_RouterCustomization(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	router := server.Router()

	// Register a test handler
	testHandlerCalled := false
	router.HandleFunc("/custom-test", func(w http.ResponseWriter, r *http.Request) {
		testHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Simulate a request
	req, _ := http.NewRequest("GET", "/custom-test", nil)
	rec := &testResponseWriter{headers: make(http.Header)}

	router.ServeHTTP(rec, req)

	if !testHandlerCalled {
		t.Error("Custom handler should have been called")
	}
}

// testResponseWriter is a minimal ResponseWriter for testing.
type testResponseWriter struct {
	headers    http.Header
	statusCode int
	body       []byte
}

func (w *testResponseWriter) Header() http.Header {
	return w.headers
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestServer_MultipleInstances(t *testing.T) {
	db1, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database 1: %v", err)
	}
	defer db1.Close()

	if err := sqlite.RunMigrations(db1.Conn()); err != nil {
		t.Fatalf("Failed to run migrations 1: %v", err)
	}

	db2, err := sqlite.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database 2: %v", err)
	}
	defer db2.Close()

	if err := sqlite.RunMigrations(db2.Conn()); err != nil {
		t.Fatalf("Failed to run migrations 2: %v", err)
	}

	server1 := NewServer(db1)
	server2 := NewServer(db2)

	if server1.router == server2.router {
		t.Error("Different server instances should have different routers")
	}

	if server1.db == server2.db {
		t.Error("Different server instances should have different databases")
	}

	if server1.sseHub == server2.sseHub {
		t.Error("Different server instances should have different SSE hubs")
	}
}

func TestServer_MiddlewareApplied(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	// Test that middleware is applied by checking response headers
	req, _ := http.NewRequest("GET", "/health", nil)
	rec := &testResponseWriter{headers: make(http.Header)}

	server.router.ServeHTTP(rec, req)

	// Should have gotten a response (middleware didn't block)
	if rec.statusCode == 0 {
		rec.statusCode = 200 // Default if not explicitly set
	}

	if rec.statusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.statusCode)
	}
}

func TestServer_RoutesRegistered(t *testing.T) {
	server, db := setupServerTest(t)
	defer db.Close()

	routes := []string{
		"/health",
		"/metrics",
		"/api/logs",
		"/api/stats",
	}

	for _, route := range routes {
		t.Run("Route: "+route, func(t *testing.T) {
			req, _ := http.NewRequest("GET", route, nil)
			rec := &testResponseWriter{headers: make(http.Header)}

			server.router.ServeHTTP(rec, req)

			// Status 0 means WriteHeader was never called, which implies 200
			if rec.statusCode == 0 {
				rec.statusCode = 200
			}

			// Route should exist (not 404)
			if rec.statusCode == http.StatusNotFound {
				t.Errorf("Route %s should be registered, got 404", route)
			}
		})
	}
}
