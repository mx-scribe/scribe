package http

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testHandler is a simple handler for testing middleware.
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// TestCORSMiddleware tests the CORS middleware.
func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(testHandler))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check CORS headers
	headers := rec.Header()

	if origin := headers.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
	}

	if methods := headers.Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Access-Control-Allow-Methods header not set")
	}

	if allowHeaders := headers.Get("Access-Control-Allow-Headers"); allowHeaders == "" {
		t.Error("Access-Control-Allow-Headers header not set")
	}

	// Should call next handler
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(testHandler))

	// OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should return 204 No Content for preflight
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, rec.Code)
	}

	// CORS headers should still be set
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Error("CORS headers not set for OPTIONS request")
	}

	// Should NOT call next handler (body should be empty)
	if rec.Body.String() == "OK" {
		t.Error("Next handler should not be called for OPTIONS request")
	}
}

func TestCORSMiddleware_AllMethods(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(testHandler))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method: "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// All methods should get CORS headers
			if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("CORS headers not set for %s request", method)
			}
		})
	}
}

// TestMetricsMiddleware tests the metrics tracking middleware.
func TestMetricsMiddleware(t *testing.T) {
	// Reset metrics before test
	atomic.StoreUint64(&serverMetrics.TotalRequests, 0)
	atomic.StoreInt64(&serverMetrics.ActiveRequests, 0)
	atomic.StoreUint64(&serverMetrics.TotalErrors, 0)

	handler := metricsMiddleware(http.HandlerFunc(testHandler))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check that request was tracked
	if atomic.LoadUint64(&serverMetrics.TotalRequests) != 1 {
		t.Errorf("Expected 1 total request, got %d", serverMetrics.TotalRequests)
	}

	// Active requests should be 0 after request completes
	if atomic.LoadInt64(&serverMetrics.ActiveRequests) != 0 {
		t.Errorf("Expected 0 active requests, got %d", serverMetrics.ActiveRequests)
	}

	// No errors for successful request
	if atomic.LoadUint64(&serverMetrics.TotalErrors) != 0 {
		t.Errorf("Expected 0 errors, got %d", serverMetrics.TotalErrors)
	}
}

func TestMetricsMiddleware_TrackErrors(t *testing.T) {
	// Reset metrics
	atomic.StoreUint64(&serverMetrics.TotalRequests, 0)
	atomic.StoreUint64(&serverMetrics.TotalErrors, 0)

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := metricsMiddleware(errorHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check that error was tracked
	if atomic.LoadUint64(&serverMetrics.TotalErrors) != 1 {
		t.Errorf("Expected 1 error, got %d", serverMetrics.TotalErrors)
	}
}

func TestMetricsMiddleware_Track4xxErrors(t *testing.T) {
	// Reset metrics
	atomic.StoreUint64(&serverMetrics.TotalErrors, 0)

	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := metricsMiddleware(notFoundHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// 4xx should also be tracked as errors
	if atomic.LoadUint64(&serverMetrics.TotalErrors) != 1 {
		t.Errorf("Expected 1 error for 4xx, got %d", serverMetrics.TotalErrors)
	}
}

func TestMetricsMiddleware_MultipleRequests(t *testing.T) {
	// Reset metrics
	atomic.StoreUint64(&serverMetrics.TotalRequests, 0)

	handler := metricsMiddleware(http.HandlerFunc(testHandler))

	// Make 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if atomic.LoadUint64(&serverMetrics.TotalRequests) != 5 {
		t.Errorf("Expected 5 total requests, got %d", serverMetrics.TotalRequests)
	}
}

// TestRateLimiter tests the rate limiting middleware.
func TestRateLimiter(t *testing.T) {
	// Create a rate limiter with low limit for testing
	limiter := rateLimiter(5, time.Second)
	handler := limiter(http.HandlerFunc(testHandler))

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rec.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 (rate limited), got %d", rec.Code)
	}

	// Check Retry-After header
	if rec.Header().Get("Retry-After") != "1" {
		t.Error("Expected Retry-After header")
	}
}

func TestRateLimiter_RefillTokens(t *testing.T) {
	// Create a rate limiter that refills quickly
	limiter := rateLimiter(2, 100*time.Millisecond)
	handler := limiter(http.HandlerFunc(testHandler))

	// Use up all tokens
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Should be rate limited now
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected rate limit, got %d", rec.Code)
	}

	// Wait for tokens to refill
	time.Sleep(150 * time.Millisecond)

	// Should be able to make request again
	req = httptest.NewRequest("GET", "/test", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 after refill, got %d", rec.Code)
	}
}

// TestRequestLogger tests the request logging middleware.
func TestRequestLogger(t *testing.T) {
	handler := requestLogger(http.HandlerFunc(testHandler))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// This test just ensures the middleware doesn't panic
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestRequestLogger_AllMethods(t *testing.T) {
	handler := requestLogger(http.HandlerFunc(testHandler))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method: "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}
		})
	}
}

// TestGetMetrics tests the GetMetrics function.
func TestGetMetrics(t *testing.T) {
	metrics := GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics() returned nil")
	}

	if metrics != serverMetrics {
		t.Error("GetMetrics() should return serverMetrics")
	}
}

// TestMetrics_Duration tests that request durations are tracked.
func TestMetrics_Duration(t *testing.T) {
	// Reset duration tracking
	serverMetrics.RequestDuration = sync.Map{}

	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := metricsMiddleware(slowHandler)

	req := httptest.NewRequest("GET", "/slow-endpoint", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check that duration was recorded
	if existing, ok := serverMetrics.RequestDuration.Load("/slow-endpoint"); ok {
		durations := existing.([]time.Duration)
		if len(durations) == 0 {
			t.Error("Expected at least one duration to be recorded")
		}
		if durations[0] < 10*time.Millisecond {
			t.Error("Duration should be at least 10ms")
		}
	} else {
		t.Error("Expected duration to be recorded for /slow-endpoint")
	}
}

// TestMetrics_DurationCap tests that durations are capped at 100.
func TestMetrics_DurationCap(t *testing.T) {
	// Reset duration tracking
	serverMetrics.RequestDuration = sync.Map{}

	handler := metricsMiddleware(http.HandlerFunc(testHandler))

	// Make 105 requests
	for i := 0; i < 105; i++ {
		req := httptest.NewRequest("GET", "/capped-endpoint", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Check that durations are capped at 100
	if existing, ok := serverMetrics.RequestDuration.Load("/capped-endpoint"); ok {
		durations := existing.([]time.Duration)
		if len(durations) > 100 {
			t.Errorf("Expected max 100 durations, got %d", len(durations))
		}
	}
}
