package services

import (
	"encoding/json"
	"strings"

	"github.com/mx-scribe/scribe/internal/domain/entities"
)

// SourceDeriver intelligently derives service names from log content.
type SourceDeriver struct{}

// NewSourceDeriver creates a new source deriver service.
func NewSourceDeriver() *SourceDeriver {
	return &SourceDeriver{}
}

// DeriveSource intelligently extracts or derives the source/service name from log content.
func (sd *SourceDeriver) DeriveSource(log *entities.Log) string {
	// 1. Check if header already has a source
	if log.Header.Source != "" {
		return log.Header.Source
	}

	// 2. Check common body fields (service, source, app, application)
	if source := sd.extractFromBodyFields(log.Body); source != "" {
		return source
	}

	// 3. Try to extract from title prefix
	if source := sd.extractFromTitlePrefix(log.Header.Title); source != "" {
		return source
	}

	// 4. Smart extraction from all content
	allText := sd.getSearchableText(log)
	return sd.smartSourceExtraction(allText)
}

// extractFromBodyFields checks common body fields for source information.
func (sd *SourceDeriver) extractFromBodyFields(body map[string]any) string {
	// Priority order of fields to check
	sourceFields := []string{"service", "source", "app", "application", "component", "module"}

	for _, field := range sourceFields {
		if val, ok := body[field]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return strVal
			}
		}
	}

	return ""
}

// extractFromTitlePrefix extracts source from title prefix like "[api]" or "auth:".
func (sd *SourceDeriver) extractFromTitlePrefix(title string) string {
	// Check for [bracket] prefix
	if strings.HasPrefix(title, "[") {
		if idx := strings.Index(title, "]"); idx > 1 {
			return strings.ToLower(title[1:idx])
		}
	}

	// Check for colon prefix
	if idx := strings.Index(title, ":"); idx > 0 && idx < 20 {
		prefix := strings.TrimSpace(title[:idx])
		if !strings.Contains(prefix, " ") {
			return strings.ToLower(prefix)
		}
	}

	return ""
}

// getSearchableText combines all text content from the log.
func (sd *SourceDeriver) getSearchableText(log *entities.Log) string {
	var parts []string

	parts = append(parts, log.Header.Title)

	if log.Header.Description != "" {
		parts = append(parts, log.Header.Description)
	}

	if len(log.Body) > 0 {
		if bodyJSON, err := json.Marshal(log.Body); err == nil {
			parts = append(parts, string(bodyJSON))
		}
	}

	return strings.Join(parts, " ")
}

// smartSourceExtraction intelligently derives service names from log content.
func (sd *SourceDeriver) smartSourceExtraction(allText string) string {
	textLower := strings.ToLower(allText)

	// Database-related patterns
	if sd.containsAny(textLower, []string{"database", "sql", "query", "table"}) {
		if strings.Contains(textLower, "postgres") {
			return "postgresql-db"
		} else if strings.Contains(textLower, "mysql") {
			return "mysql-db"
		} else if strings.Contains(textLower, "mongo") {
			return "mongodb"
		} else if strings.Contains(textLower, "redis") {
			return "redis-cache"
		} else if strings.Contains(textLower, "sqlite") {
			return "sqlite-db"
		}
		return "database-service"
	}

	// Authentication/Security patterns
	if sd.containsAny(textLower, []string{"login", "auth", "token", "session", "jwt", "oauth"}) {
		return "auth-service"
	}

	// Payment processing patterns
	if sd.containsAny(textLower, []string{"payment", "stripe", "paypal", "billing", "invoice", "checkout"}) {
		return "payment-service"
	}

	// Email/Notification patterns
	if sd.containsAny(textLower, []string{"email", "smtp", "notification", "mailgun", "sendgrid", "push notification"}) {
		return "email-service"
	}

	// API Gateway patterns
	if sd.containsAny(textLower, []string{"api gateway", "endpoint", "route", "/api/"}) {
		return "api-gateway"
	}

	// User management patterns
	if strings.Contains(textLower, "user") && sd.containsAny(textLower, []string{"profile", "register", "account", "signup"}) {
		return "user-service"
	}

	// Order/Shopping patterns
	if sd.containsAny(textLower, []string{"order", "cart", "inventory", "product", "catalog"}) {
		return "order-service"
	}

	// File/Storage patterns
	if sd.containsAny(textLower, []string{"file", "upload", "download", "s3", "storage", "blob", "bucket"}) {
		return "file-service"
	}

	// Search patterns
	if sd.containsAny(textLower, []string{"search", "elasticsearch", "solr", "algolia", "opensearch"}) {
		return "search-service"
	}

	// Monitoring/Health patterns
	if sd.containsAny(textLower, []string{"health", "monitor", "metrics", "prometheus", "grafana", "datadog"}) {
		return "monitoring-service"
	}

	// Load balancer patterns
	if sd.containsAny(textLower, []string{"load balan", "nginx", "haproxy", "upstream", "reverse proxy"}) {
		return "load-balancer"
	}

	// Cache patterns
	if strings.Contains(textLower, "cache") && !strings.Contains(textLower, "redis") {
		return "cache-service"
	}

	// Queue/Message patterns
	if sd.containsAny(textLower, []string{"queue", "rabbitmq", "kafka", "sqs", "pubsub", "message broker"}) {
		return "queue-service"
	}

	// Configuration patterns
	if sd.containsAny(textLower, []string{"config", "setting", "environment", "feature flag"}) {
		return "config-service"
	}

	// Backup patterns
	if sd.containsAny(textLower, []string{"backup", "restore", "archive", "snapshot"}) {
		return "backup-service"
	}

	// Reporting/Analytics patterns
	if sd.containsAny(textLower, []string{"report", "analytics", "dashboard", "bi ", "business intelligence"}) {
		return "reporting-service"
	}

	// Deployment/CI/CD patterns
	if sd.containsAny(textLower, []string{"deploy", "build", "pipeline", "docker", "kubernetes", "k8s", "ci/cd"}) {
		return "deployment-service"
	}

	// CDN patterns
	if sd.containsAny(textLower, []string{"cdn", "cloudflare", "fastly", "akamai", "static asset"}) {
		return "cdn-service"
	}

	// Scheduler/Cron patterns
	if sd.containsAny(textLower, []string{"cron", "scheduler", "scheduled task", "job runner"}) {
		return "scheduler-service"
	}

	// Webhook patterns
	if sd.containsAny(textLower, []string{"webhook", "callback", "hook endpoint"}) {
		return "webhook-service"
	}

	// Try to extract from common service naming patterns
	words := strings.Fields(textLower)
	for _, word := range words {
		if strings.Contains(word, "-service") || strings.Contains(word, "_service") {
			cleanWord := strings.Trim(word, ".,!?:;\"'()[]{}")
			if len(cleanWord) > 2 {
				return cleanWord
			}
		}
	}

	// Default - don't return anything if we can't determine
	return ""
}

// containsAny checks if text contains any of the patterns.
func (sd *SourceDeriver) containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
