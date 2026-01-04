package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// MetricsData holds collected metrics.
type MetricsData struct {
	TotalRequests  uint64 `json:"total_requests"`
	ActiveRequests int64  `json:"active_requests"`
	TotalErrors    uint64 `json:"total_errors"`
	ErrorRate      string `json:"error_rate"`
	Uptime         string `json:"uptime"`
	GoRoutines     int    `json:"go_routines"`
	MemoryMB       uint64 `json:"memory_mb"`
	SSEClients     int    `json:"sse_clients,omitempty"`
}

// MetricsCollector interface for getting metrics from the server.
type MetricsCollector interface {
	GetTotalRequests() uint64
	GetActiveRequests() int64
	GetTotalErrors() uint64
}

var startTime = time.Now()

// MetricsHandler handles GET /metrics.
func MetricsHandler(getMetrics func() (uint64, int64, uint64), sseHub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalReqs, activeReqs, totalErrs := getMetrics()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		errorRate := "0.00%"
		if totalReqs > 0 {
			rate := float64(totalErrs) / float64(totalReqs) * 100
			errorRate = formatPercent(rate)
		}

		data := MetricsData{
			TotalRequests:  totalReqs,
			ActiveRequests: activeReqs,
			TotalErrors:    totalErrs,
			ErrorRate:      errorRate,
			Uptime:         formatDuration(time.Since(startTime)),
			GoRoutines:     runtime.NumGoroutine(),
			MemoryMB:       memStats.Alloc / 1024 / 1024,
		}

		if sseHub != nil {
			data.SSEClients = sseHub.ClientCount()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	}
}

// PrometheusMetricsHandler handles GET /metrics/prometheus.
func PrometheusMetricsHandler(getMetrics func() (uint64, int64, uint64), sseHub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalReqs, activeReqs, totalErrs := getMetrics()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		sseClients := 0
		if sseHub != nil {
			sseClients = sseHub.ClientCount()
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Prometheus format
		_, _ = w.Write([]byte("# HELP scribe_http_requests_total Total number of HTTP requests\n"))
		_, _ = w.Write([]byte("# TYPE scribe_http_requests_total counter\n"))
		writeMetric(w, "scribe_http_requests_total", totalReqs)

		_, _ = w.Write([]byte("# HELP scribe_http_requests_active Current number of active HTTP requests\n"))
		_, _ = w.Write([]byte("# TYPE scribe_http_requests_active gauge\n"))
		writeMetricInt(w, "scribe_http_requests_active", activeReqs)

		_, _ = w.Write([]byte("# HELP scribe_http_errors_total Total number of HTTP errors (4xx and 5xx)\n"))
		_, _ = w.Write([]byte("# TYPE scribe_http_errors_total counter\n"))
		writeMetric(w, "scribe_http_errors_total", totalErrs)

		_, _ = w.Write([]byte("# HELP scribe_uptime_seconds Server uptime in seconds\n"))
		_, _ = w.Write([]byte("# TYPE scribe_uptime_seconds gauge\n"))
		writeMetricFloat(w, "scribe_uptime_seconds", time.Since(startTime).Seconds())

		_, _ = w.Write([]byte("# HELP scribe_goroutines Current number of goroutines\n"))
		_, _ = w.Write([]byte("# TYPE scribe_goroutines gauge\n"))
		writeMetricInt(w, "scribe_goroutines", int64(runtime.NumGoroutine()))

		_, _ = w.Write([]byte("# HELP scribe_memory_bytes Current memory usage in bytes\n"))
		_, _ = w.Write([]byte("# TYPE scribe_memory_bytes gauge\n"))
		writeMetric(w, "scribe_memory_bytes", memStats.Alloc)

		_, _ = w.Write([]byte("# HELP scribe_sse_clients Current number of SSE clients\n"))
		_, _ = w.Write([]byte("# TYPE scribe_sse_clients gauge\n"))
		writeMetricInt(w, "scribe_sse_clients", int64(sseClients))
	}
}

func writeMetric(w http.ResponseWriter, name string, value uint64) {
	_, _ = w.Write([]byte(name + " " + formatUint(value) + "\n"))
}

func writeMetricInt(w http.ResponseWriter, name string, value int64) {
	_, _ = w.Write([]byte(name + " " + formatInt(value) + "\n"))
}

func writeMetricFloat(w http.ResponseWriter, name string, value float64) {
	_, _ = w.Write([]byte(name + " " + formatFloat(value) + "\n"))
}

func formatUint(v uint64) string {
	return formatInt(int64(v)) //nolint:gosec // Metrics won't overflow int64
}

func formatInt(v int64) string {
	if v == 0 {
		return "0"
	}
	var result []byte
	negative := v < 0
	if negative {
		v = -v
	}
	for v > 0 {
		result = append([]byte{byte('0' + v%10)}, result...)
		v /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}

func formatFloat(v float64) string {
	// Simple float formatting
	intPart := int64(v)
	fracPart := int64((v - float64(intPart)) * 1000)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	return formatInt(intPart) + "." + padLeft(formatInt(fracPart), 3, '0')
}

func formatPercent(v float64) string {
	return formatFloat(v) + "%"
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return formatInt(int64(hours)) + "h " + formatInt(int64(minutes)) + "m " + formatInt(int64(seconds)) + "s"
	}
	if minutes > 0 {
		return formatInt(int64(minutes)) + "m " + formatInt(int64(seconds)) + "s"
	}
	return formatInt(int64(seconds)) + "s"
}

func padLeft(s string, length int, pad byte) string {
	for len(s) < length {
		s = string(pad) + s
	}
	return s
}

// SimpleMetrics is a simple implementation of metrics for standalone use.
type SimpleMetrics struct {
	totalRequests  uint64
	activeRequests int64
	totalErrors    uint64
}

// IncrementRequests increments the total request count.
func (m *SimpleMetrics) IncrementRequests() {
	atomic.AddUint64(&m.totalRequests, 1)
}

// IncrementErrors increments the error count.
func (m *SimpleMetrics) IncrementErrors() {
	atomic.AddUint64(&m.totalErrors, 1)
}

// GetTotalRequests returns total requests.
func (m *SimpleMetrics) GetTotalRequests() uint64 {
	return atomic.LoadUint64(&m.totalRequests)
}

// GetActiveRequests returns active requests.
func (m *SimpleMetrics) GetActiveRequests() int64 {
	return atomic.LoadInt64(&m.activeRequests)
}

// GetTotalErrors returns total errors.
func (m *SimpleMetrics) GetTotalErrors() uint64 {
	return atomic.LoadUint64(&m.totalErrors)
}
