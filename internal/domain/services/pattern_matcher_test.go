package services

import (
	"testing"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

func TestPatternMatcher_AnalyzeLog_SecurityPatterns(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"SQL injection attempt detected", "critical"},
		{"Unauthorized access to /admin", "critical"},
		{"XSS attack blocked", "critical"},
		{"Authentication failed for user", "critical"},
		{"Token expired", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_HTTPStatus(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"Request returned status 200", "success"},
		{"HTTP 201 Created", "success"},
		{"API returned 404 not found", "warning"},
		{"Server error: status 500", "error"},
		{"Gateway timeout: HTTP 504", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_StackTrace(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"Error at line 42", "error"},
		{"Traceback (most recent call last)", "error"},
		{"panic: runtime error", "error"},
		{"goroutine 1 [running]:", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_DatabasePatterns(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"Deadlock detected in transaction", "critical"},
		{"Connection pool exhausted", "critical"},
		{"Duplicate key error on insert", "warning"},
		{"Foreign key violation", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_Keywords(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"Connection failed", "error"},
		{"Request timeout", "error"},
		{"Deprecated API called", "warning"},
		{"Slow query detected", "warning"},
		{"Payment processed successfully", "success"},
		{"User created", "success"},
		{"Debug mode enabled", "debug"},
		{"Entering function", "debug"},
		{"Processing request", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_SourceExtraction(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		name           string
		title          string
		body           map[string]any
		expectedSource string
	}{
		{
			name:           "from body service field",
			title:          "Error occurred",
			body:           map[string]any{"service": "payment"},
			expectedSource: "payment",
		},
		{
			name:           "from body source field",
			title:          "Error occurred",
			body:           map[string]any{"source": "api"},
			expectedSource: "api",
		},
		{
			name:           "from title prefix with brackets",
			title:          "[auth] Login failed",
			body:           nil,
			expectedSource: "auth",
		},
		{
			name:           "from title prefix with colon",
			title:          "database: Connection failed",
			body:           nil,
			expectedSource: "database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := entities.LogHeader{
				Title:    tt.title,
				Severity: valueobjects.SeverityInfo,
			}
			log := entities.NewLog(header, tt.body)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSource != tt.expectedSource {
				t.Errorf("got %q, want %q", meta.DerivedSource, tt.expectedSource)
			}
		})
	}
}

func createTestLog(title string) *entities.Log {
	header := entities.LogHeader{
		Title:    title,
		Severity: valueobjects.SeverityInfo,
	}
	return entities.NewLog(header, nil)
}

func TestPatternMatcher_AnalyzeLog_BusinessPatterns(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
		category string
	}{
		{"Payment failed for order 123", "error", "business"},
		{"Subscription canceled", "warning", "business"},
		{"Refund processed successfully", "info", "business"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("severity: got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
			if meta.DerivedCategory != tt.category {
				t.Errorf("category: got %q, want %q", meta.DerivedCategory, tt.category)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_PerformancePatterns(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		name     string
		title    string
		body     map[string]any
		expected string
	}{
		{
			name:     "slow query text",
			title:    "Slow query took 5 seconds",
			body:     nil,
			expected: "warning",
		},
		{
			name:     "memory leak",
			title:    "Memory leak detected in worker",
			body:     nil,
			expected: "error", // "leak" triggers error keyword
		},
		{
			name:     "duration from body float64",
			title:    "Request completed",
			body:     map[string]any{"duration_ms": float64(2500)},
			expected: "warning",
		},
		{
			name:     "duration from body int",
			title:    "Request completed",
			body:     map[string]any{"latency": 500},
			expected: "info",
		},
		{
			name:     "duration from body string",
			title:    "Request completed",
			body:     map[string]any{"elapsed_ms": "1200"},
			expected: "warning", // 1200ms > 1000ms is slow
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := entities.LogHeader{
				Title:    tt.title,
				Severity: valueobjects.SeverityInfo,
			}
			log := entities.NewLog(header, tt.body)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_AnalyzeLog_SystemErrorCodes(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"Error ECONNREFUSED on connect", "error"},
		{"Signal SIGKILL received", "error"}, // SIGKILL is not in SystemErrorCodes, triggers error keyword
		{"ENOMEM: cannot allocate", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedSeverity != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedSeverity, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_DetectCategory(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		title    string
		expected string
	}{
		{"HTTP request failed", "http"},
		{"SQL query executed", "database"},
		{"Latency spike detected", "performance"},
		{"User logged in", "business"},
		{"Process started", "system"},
		{"Random log message", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			log := createTestLog(tt.title)
			meta := pm.AnalyzeLog(log)
			if meta.DerivedCategory != tt.expected {
				t.Errorf("got %q, want %q", meta.DerivedCategory, tt.expected)
			}
		})
	}
}

func TestSourceDeriver_DeriveSource_SmartExtraction(t *testing.T) {
	sd := NewSourceDeriver()

	tests := []struct {
		name     string
		title    string
		body     map[string]any
		expected string
	}{
		{
			name:     "postgres pattern",
			title:    "PostgreSQL query failed",
			expected: "postgresql-db",
		},
		{
			name:     "mysql pattern",
			title:    "MySQL connection error",
			expected: "mysql-db",
		},
		{
			name:     "mongodb pattern",
			title:    "Mongo database connection error",
			expected: "mongodb",
		},
		{
			name:     "redis pattern",
			title:    "Redis query timeout",
			expected: "redis-cache",
		},
		{
			name:     "sqlite pattern",
			title:    "SQLite database locked",
			expected: "sqlite-db",
		},
		{
			name:     "generic database",
			title:    "Database query timeout",
			expected: "database-service",
		},
		{
			name:     "auth pattern",
			title:    "Login attempt blocked",
			expected: "auth-service",
		},
		{
			name:     "payment pattern",
			title:    "Stripe payment processed",
			expected: "payment-service",
		},
		{
			name:     "email pattern",
			title:    "Email sent via SendGrid",
			expected: "email-service",
		},
		{
			name:     "api gateway",
			title:    "API gateway /api/users request",
			expected: "api-gateway",
		},
		{
			name:     "user service",
			title:    "User profile updated",
			expected: "user-service",
		},
		{
			name:     "order service",
			title:    "Order placed in cart",
			expected: "order-service",
		},
		{
			name:     "file service",
			title:    "File upload to S3 completed",
			expected: "file-service",
		},
		{
			name:     "search service",
			title:    "Search index updated in Algolia",
			expected: "search-service",
		},
		{
			name:     "monitoring",
			title:    "Health check passed",
			expected: "monitoring-service",
		},
		{
			name:     "load balancer",
			title:    "Nginx upstream error",
			expected: "load-balancer",
		},
		{
			name:     "cache service",
			title:    "Cache invalidation",
			expected: "cache-service",
		},
		{
			name:     "queue service",
			title:    "RabbitMQ message published",
			expected: "queue-service",
		},
		{
			name:     "config service",
			title:    "Config setting updated",
			expected: "config-service",
		},
		{
			name:     "backup service",
			title:    "Backup completed successfully",
			expected: "backup-service",
		},
		{
			name:     "reporting",
			title:    "Report generated for analytics",
			expected: "reporting-service",
		},
		{
			name:     "deployment",
			title:    "Docker container deployed",
			expected: "deployment-service",
		},
		{
			name:     "cdn",
			title:    "CDN Cloudflare edge node",
			expected: "cdn-service",
		},
		{
			name:     "scheduler",
			title:    "Cron job executed",
			expected: "scheduler-service",
		},
		{
			name:     "webhook",
			title:    "Webhook callback received",
			expected: "webhook-service",
		},
		{
			name:     "service naming pattern",
			title:    "my-service: request handled",
			expected: "my-service",
		},
		{
			name:     "explicit service in text",
			title:    "Error from payment-service",
			expected: "payment-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := entities.LogHeader{
				Title:    tt.title,
				Severity: valueobjects.SeverityInfo,
			}
			log := entities.NewLog(header, tt.body)
			source := sd.DeriveSource(log)
			if source != tt.expected {
				t.Errorf("got %q, want %q", source, tt.expected)
			}
		})
	}
}

func TestSourceDeriver_DeriveSource_FromHeader(t *testing.T) {
	sd := NewSourceDeriver()

	header := entities.LogHeader{
		Title:    "Some log",
		Severity: valueobjects.SeverityInfo,
		Source:   "explicit-source",
	}
	log := entities.NewLog(header, nil)

	source := sd.DeriveSource(log)
	if source != "explicit-source" {
		t.Errorf("got %q, want explicit-source", source)
	}
}

func TestSourceDeriver_DeriveSource_FromBodyFields(t *testing.T) {
	sd := NewSourceDeriver()

	tests := []struct {
		name     string
		body     map[string]any
		expected string
	}{
		{"service field", map[string]any{"service": "my-service"}, "my-service"},
		{"source field", map[string]any{"source": "my-source"}, "my-source"},
		{"app field", map[string]any{"app": "my-app"}, "my-app"},
		{"application field", map[string]any{"application": "my-application"}, "my-application"},
		{"component field", map[string]any{"component": "my-component"}, "my-component"},
		{"module field", map[string]any{"module": "my-module"}, "my-module"},
		{"empty string ignored", map[string]any{"service": ""}, ""},
		{"non-string ignored", map[string]any{"service": 123}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := entities.LogHeader{
				Title:    "Log entry",
				Severity: valueobjects.SeverityInfo,
			}
			log := entities.NewLog(header, tt.body)
			source := sd.DeriveSource(log)
			if source != tt.expected {
				t.Errorf("got %q, want %q", source, tt.expected)
			}
		})
	}
}

func TestSourceDeriver_DeriveSource_FromTitlePrefix(t *testing.T) {
	sd := NewSourceDeriver()

	tests := []struct {
		title    string
		expected string
	}{
		{"[api] Request failed", "api"},
		{"[AUTH] Login attempt", "auth"},
		{"database: Connection error", "database"},
		{"api-gateway: Request timeout", "api-gateway"},
		{"Regular title without prefix", ""},
		{"Title with spaces: should not match", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			header := entities.LogHeader{
				Title:    tt.title,
				Severity: valueobjects.SeverityInfo,
			}
			log := entities.NewLog(header, nil)
			source := sd.DeriveSource(log)
			if source != tt.expected {
				t.Errorf("got %q, want %q", source, tt.expected)
			}
		})
	}
}
