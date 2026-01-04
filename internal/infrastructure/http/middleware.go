package http

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Metrics tracks server metrics.
type Metrics struct {
	TotalRequests   uint64
	ActiveRequests  int64
	TotalErrors     uint64
	RequestDuration sync.Map
}

var serverMetrics = &Metrics{}

// GetMetrics returns the server metrics.
func GetMetrics() *Metrics {
	return serverMetrics
}

// setupMiddleware configures all middleware for the server.
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(metricsMiddleware)
	s.router.Use(requestLogger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(rateLimiter(100, time.Second))
	s.router.Use(corsMiddleware)
	s.router.Use(middleware.SetHeader("Content-Type", "application/json"))
}

// requestLogger logs each request with timing.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		log.Printf("%s %s %d %s",
			r.Method,
			r.URL.Path,
			ww.Status(),
			time.Since(start).Round(time.Millisecond),
		)
	})
}

// corsMiddleware handles CORS headers for browser requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// metricsMiddleware tracks request metrics.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		atomic.AddInt64(&serverMetrics.ActiveRequests, 1)
		defer atomic.AddInt64(&serverMetrics.ActiveRequests, -1)
		atomic.AddUint64(&serverMetrics.TotalRequests, 1)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		if ww.Status() >= 400 {
			atomic.AddUint64(&serverMetrics.TotalErrors, 1)
		}

		duration := time.Since(start)
		path := r.URL.Path
		if existing, ok := serverMetrics.RequestDuration.Load(path); ok {
			durations := existing.([]time.Duration)
			if len(durations) >= 100 {
				durations = durations[1:]
			}
			serverMetrics.RequestDuration.Store(path, append(durations, duration))
		} else {
			serverMetrics.RequestDuration.Store(path, []time.Duration{duration})
		}
	})
}

// rateLimiter implements a simple token bucket rate limiter.
func rateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	var (
		mu       sync.Mutex
		tokens   = limit
		lastTime = time.Now()
	)

	go func() {
		ticker := time.NewTicker(window / time.Duration(limit))
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			if tokens < limit {
				tokens++
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()

			now := time.Now()
			elapsed := now.Sub(lastTime)
			refill := int(elapsed / (window / time.Duration(limit)))
			if refill > 0 {
				tokens += refill
				if tokens > limit {
					tokens = limit
				}
				lastTime = now
			}

			if tokens <= 0 {
				mu.Unlock()
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
