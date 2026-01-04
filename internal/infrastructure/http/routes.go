package http

import (
	"io/fs"

	"github.com/go-chi/chi/v5"

	"github.com/mx-scribe/scribe/internal/infrastructure/http/handlers"
)

// setupRoutes configures API routes for the server.
func (s *Server) setupRoutes() {
	s.router.Get("/health", handlers.Health)

	getMetrics := func() (uint64, int64, uint64) {
		m := GetMetrics()
		return m.TotalRequests, m.ActiveRequests, m.TotalErrors
	}
	s.router.Get("/metrics", handlers.MetricsHandler(getMetrics, s.sseHub))
	s.router.Get("/metrics/prometheus", handlers.PrometheusMetricsHandler(getMetrics, s.sseHub))

	s.router.Route("/api", func(r chi.Router) {
		r.Post("/logs", handlers.CreateLogWithSSE(s.db, s.sseHub))
		r.Get("/logs", handlers.ListLogs(s.db))
		r.Get("/logs/{id}", handlers.GetLog(s.db))
		r.Delete("/logs/{id}", handlers.DeleteLogWithSSE(s.db, s.sseHub))
		r.Delete("/logs", handlers.DeleteLogsWithSSE(s.db, s.sseHub))

		r.Get("/stats", handlers.GetStats(s.db))

		r.Get("/export/json", handlers.ExportJSON(s.db))
		r.Get("/export/csv", handlers.ExportCSV(s.db))

		r.Get("/events", handlers.SSEHandler(s.sseHub))

		r.Route("/admin", func(r chi.Router) {
			r.Get("/retention", handlers.GetRetentionInfo(s.db))
			r.Post("/cleanup", handlers.CleanupLogs(s.db))
		})
	})
}

// SetStaticFS sets the embedded filesystem for serving static files.
func (s *Server) SetStaticFS(staticFS fs.FS) {
	s.staticFS = staticFS
	spaHandler := handlers.NewSPAHandler(staticFS, "dist")
	s.router.Handle("/*", spaHandler)
}
