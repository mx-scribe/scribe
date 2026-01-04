package services

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
	"github.com/mx-scribe/scribe/internal/patterns/rules"
)

// PatternMatcher analyzes log content and extracts intelligent metadata.
type PatternMatcher struct {
	sourceDeriver *SourceDeriver
}

// NewPatternMatcher creates a new pattern matcher service.
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		sourceDeriver: NewSourceDeriver(),
	}
}

// AnalyzeLog performs comprehensive pattern matching on a log entry.
func (pm *PatternMatcher) AnalyzeLog(log *entities.Log) entities.LogMetadata {
	// Combine all searchable text
	allText := pm.getSearchableText(log)
	textLower := strings.ToLower(allText)

	metadata := entities.LogMetadata{}

	// 1. Detect category first
	metadata.DerivedCategory = pm.detectCategory(textLower).String()

	// 2. Check for security issues first (highest priority - critical)
	if pm.detectSecurityIssue(textLower) {
		metadata.DerivedSeverity = "critical"
		metadata.DerivedCategory = valueobjects.CategorySecurity.String()
		metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
		return metadata
	}

	// 3. Check business patterns
	if severity := pm.checkBusinessPatterns(textLower); severity != "" {
		metadata.DerivedSeverity = severity
		metadata.DerivedCategory = valueobjects.CategoryBusiness.String()
		metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
		return metadata
	}

	// 4. Check performance patterns
	if severity := pm.checkPerformancePatterns(log, textLower); severity != "" {
		metadata.DerivedSeverity = severity
		metadata.DerivedCategory = valueobjects.CategoryPerformance.String()
		metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
		return metadata
	}

	// 5. Extract HTTP status code and map to severity
	if statusCode := pm.extractHTTPStatusCode(allText); statusCode != "" {
		if severity, ok := rules.HTTPStatusSeverity[statusCode]; ok {
			metadata.DerivedSeverity = severity
			metadata.DerivedCategory = valueobjects.CategoryHTTP.String()
			metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
			return metadata
		}
	}

	// 6. Check for stack traces (indicates error)
	if pm.hasStackTrace(allText) {
		metadata.DerivedSeverity = "error"
		metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
		return metadata
	}

	// 7. Check database patterns
	for pattern, severity := range rules.DatabasePatterns {
		if strings.Contains(textLower, pattern) {
			metadata.DerivedSeverity = severity
			metadata.DerivedCategory = valueobjects.CategoryDatabase.String()
			metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
			return metadata
		}
	}

	// 8. Check system error codes
	for code, severity := range rules.SystemErrorCodes {
		if strings.Contains(allText, code) {
			metadata.DerivedSeverity = severity
			metadata.DerivedCategory = valueobjects.CategorySystem.String()
			metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)
			return metadata
		}
	}

	// 9. Check keyword-based severity detection
	metadata.DerivedSeverity = pm.detectSeverityFromKeywords(textLower)

	// 10. Extract source from content
	metadata.DerivedSource = pm.sourceDeriver.DeriveSource(log)

	return metadata
}

// detectCategory detects the log category from content.
func (pm *PatternMatcher) detectCategory(textLower string) valueobjects.Category {
	// Security patterns
	if pm.detectSecurityIssue(textLower) {
		return valueobjects.CategorySecurity
	}

	// HTTP patterns
	httpPatterns := []string{"http", "request", "response", "status", "endpoint", "api", "rest", "graphql"}
	for _, pattern := range httpPatterns {
		if strings.Contains(textLower, pattern) {
			return valueobjects.CategoryHTTP
		}
	}

	// Database patterns
	dbPatterns := []string{"database", "sql", "query", "table", "postgres", "mysql", "mongo", "redis", "sqlite"}
	for _, pattern := range dbPatterns {
		if strings.Contains(textLower, pattern) {
			return valueobjects.CategoryDatabase
		}
	}

	// Performance patterns
	perfPatterns := []string{"slow", "timeout", "latency", "performance", "memory", "cpu", "disk", "duration"}
	for _, pattern := range perfPatterns {
		if strings.Contains(textLower, pattern) {
			return valueobjects.CategoryPerformance
		}
	}

	// Business patterns
	bizPatterns := []string{"payment", "order", "invoice", "subscription", "user", "login", "checkout", "cart"}
	for _, pattern := range bizPatterns {
		if strings.Contains(textLower, pattern) {
			return valueobjects.CategoryBusiness
		}
	}

	// System patterns
	sysPatterns := []string{"system", "kernel", "process", "signal", "daemon", "service", "cron", "scheduler"}
	for _, pattern := range sysPatterns {
		if strings.Contains(textLower, pattern) {
			return valueobjects.CategorySystem
		}
	}

	return valueobjects.CategoryGeneral
}

// checkBusinessPatterns checks for business-related patterns.
func (pm *PatternMatcher) checkBusinessPatterns(textLower string) string {
	for pattern, severity := range rules.BusinessPatterns {
		if strings.Contains(textLower, pattern) {
			return severity
		}
	}
	return ""
}

// checkPerformancePatterns checks for performance-related patterns.
func (pm *PatternMatcher) checkPerformancePatterns(log *entities.Log, textLower string) string {
	// Check for performance patterns in text
	for pattern, severity := range rules.PerformancePatterns {
		if strings.Contains(textLower, pattern) {
			return severity
		}
	}

	// Check for duration in body
	if len(log.Body) > 0 {
		durationFields := []string{"duration", "duration_ms", "elapsed", "elapsed_ms", "time_ms", "latency", "latency_ms"}
		for _, field := range durationFields {
			if val, ok := log.Body[field]; ok {
				var durationMs int
				switch v := val.(type) {
				case float64:
					durationMs = int(v)
				case int:
					durationMs = v
				case string:
					if parsed, err := strconv.Atoi(v); err == nil {
						durationMs = parsed
					}
				}
				if durationMs > 0 {
					level := rules.CategorizePerformance(durationMs)
					return rules.PerformanceSeverity(level)
				}
			}
		}
	}

	return ""
}

// getSearchableText combines all text content from the log.
func (pm *PatternMatcher) getSearchableText(log *entities.Log) string {
	var parts []string

	parts = append(parts, log.Header.Title)

	if log.Header.Description != "" {
		parts = append(parts, log.Header.Description)
	}

	// Convert body to string if not empty
	if len(log.Body) > 0 {
		if bodyJSON, err := json.Marshal(log.Body); err == nil {
			parts = append(parts, string(bodyJSON))
		}
	}

	return strings.Join(parts, " ")
}

// extractHTTPStatusCode extracts HTTP status codes from text.
func (pm *PatternMatcher) extractHTTPStatusCode(text string) string {
	patterns := []string{
		`(?i)(?:status|http|code)[\s:=]*(\d{3})`,
		`(?i)returned\s+(\d{3})`,
		`(?i)\b(\d{3})\s+(?:error|ok|found|not found)`,
		`(?i)"status"[\s:]+["']?(\d{3})`,
		`(?i)"status_code"[\s:]+["']?(\d{3})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// hasStackTrace detects if text contains a stack trace.
func (pm *PatternMatcher) hasStackTrace(text string) bool {
	stackIndicators := []string{
		" at line ", " at Object.", "Traceback", "goroutine ",
		"panic:", ".java:", ".py:", ".js:", ".go:",
		"at /", "File \"", " line ", "in <module>",
		"Exception in thread", "Caused by:", "\n\tat ",
		"Call Stack:", "Stack trace:", "at Function.",
	}

	textLower := strings.ToLower(text)
	for _, indicator := range stackIndicators {
		if strings.Contains(textLower, strings.ToLower(indicator)) {
			return true
		}
	}
	return false
}

// detectSecurityIssue checks for security-related patterns.
func (pm *PatternMatcher) detectSecurityIssue(textLower string) bool {
	for _, pattern := range rules.SecurityPatterns {
		if strings.Contains(textLower, pattern) {
			return true
		}
	}
	return false
}

// detectSeverityFromKeywords detects severity from keyword analysis.
func (pm *PatternMatcher) detectSeverityFromKeywords(textLower string) string {
	// Check for error keywords (highest priority)
	for _, keyword := range rules.ErrorKeywords {
		if strings.Contains(textLower, keyword) {
			return "error"
		}
	}

	// Check for warning keywords
	for _, keyword := range rules.WarningKeywords {
		if strings.Contains(textLower, keyword) {
			return "warning"
		}
	}

	// Check for success keywords
	for _, keyword := range rules.SuccessKeywords {
		if strings.Contains(textLower, keyword) {
			return "success"
		}
	}

	// Check for debug keywords
	for _, keyword := range rules.DebugKeywords {
		if strings.Contains(textLower, keyword) {
			return "debug"
		}
	}

	// Default to info if no match
	return "info"
}
