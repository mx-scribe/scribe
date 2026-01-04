package rules

import (
	"strings"
	"testing"
)

// TestHTTPStatusSeverity tests the HTTP status code to severity mapping
func TestHTTPStatusSeverity(t *testing.T) {
	if HTTPStatusSeverity == nil {
		t.Fatal("HTTPStatusSeverity should not be nil")
	}

	tests := []struct {
		statusCode       string
		expectedSeverity string
	}{
		// Success codes
		{"200", "success"},
		{"201", "success"},
		{"202", "success"},
		{"204", "success"},

		// Redirect codes
		{"301", "info"},
		{"302", "info"},
		{"304", "info"},

		// Client error codes
		{"400", "warning"},
		{"401", "error"},
		{"403", "error"},
		{"404", "warning"},
		{"408", "warning"},
		{"409", "warning"},
		{"429", "warning"},

		// Server error codes
		{"500", "error"},
		{"501", "error"},
		{"502", "error"},
		{"503", "critical"},
		{"504", "critical"},
		{"507", "critical"},
		{"511", "error"},
	}

	for _, tt := range tests {
		t.Run("Status_"+tt.statusCode, func(t *testing.T) {
			severity, exists := HTTPStatusSeverity[tt.statusCode]
			if !exists {
				t.Errorf("Status code %s not found in mapping", tt.statusCode)
				return
			}
			if severity != tt.expectedSeverity {
				t.Errorf("Status %s: expected severity '%s', got '%s'",
					tt.statusCode, tt.expectedSeverity, severity)
			}
		})
	}
}

func TestHTTPStatusSeverity_Coverage(t *testing.T) {
	count := len(HTTPStatusSeverity)
	if count < 20 {
		t.Errorf("Expected at least 20 HTTP status codes, got %d", count)
	}

	validSeverities := map[string]bool{
		"success":  true,
		"info":     true,
		"warning":  true,
		"error":    true,
		"critical": true,
	}

	for code, severity := range HTTPStatusSeverity {
		if !validSeverities[severity] {
			t.Errorf("Status code %s has invalid severity: %s", code, severity)
		}
	}
}

// TestErrorKeywords tests error keyword patterns
func TestErrorKeywords(t *testing.T) {
	if ErrorKeywords == nil {
		t.Fatal("ErrorKeywords should not be nil")
	}

	if len(ErrorKeywords) == 0 {
		t.Fatal("ErrorKeywords should not be empty")
	}

	expectedKeywords := []string{
		"error", "failed", "failure", "fatal", "critical",
		"exception", "timeout", "crash", "panic",
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range ErrorKeywords {
			if strings.Contains(keyword, expected) || strings.Contains(expected, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error keyword '%s' not found", expected)
		}
	}
}

func TestErrorKeywords_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, keyword := range ErrorKeywords {
		if seen[keyword] {
			t.Errorf("Duplicate error keyword found: %s", keyword)
		}
		seen[keyword] = true
	}
}

func TestErrorKeywords_Lowercase(t *testing.T) {
	for _, keyword := range ErrorKeywords {
		if keyword != strings.ToLower(keyword) {
			t.Errorf("Error keyword should be lowercase: %s", keyword)
		}
	}
}

// TestWarningKeywords tests warning keyword patterns
func TestWarningKeywords(t *testing.T) {
	if WarningKeywords == nil {
		t.Fatal("WarningKeywords should not be nil")
	}

	if len(WarningKeywords) == 0 {
		t.Fatal("WarningKeywords should not be empty")
	}

	expectedKeywords := []string{
		"warning", "warn", "deprecated", "slow", "retry",
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range WarningKeywords {
			if strings.Contains(keyword, expected) || strings.Contains(expected, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected warning keyword '%s' not found", expected)
		}
	}
}

func TestWarningKeywords_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, keyword := range WarningKeywords {
		if seen[keyword] {
			t.Errorf("Duplicate warning keyword found: %s", keyword)
		}
		seen[keyword] = true
	}
}

// TestSuccessKeywords tests success keyword patterns
func TestSuccessKeywords(t *testing.T) {
	if SuccessKeywords == nil {
		t.Fatal("SuccessKeywords should not be nil")
	}

	if len(SuccessKeywords) == 0 {
		t.Fatal("SuccessKeywords should not be empty")
	}

	expectedKeywords := []string{
		"success", "successful", "complete", "ok", "done",
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range SuccessKeywords {
			if strings.Contains(keyword, expected) || strings.Contains(expected, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected success keyword '%s' not found", expected)
		}
	}
}

func TestSuccessKeywords_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, keyword := range SuccessKeywords {
		if seen[keyword] {
			t.Errorf("Duplicate success keyword found: %s", keyword)
		}
		seen[keyword] = true
	}
}

// TestDebugKeywords tests debug keyword patterns
func TestDebugKeywords(t *testing.T) {
	if DebugKeywords == nil {
		t.Fatal("DebugKeywords should not be nil")
	}

	if len(DebugKeywords) == 0 {
		t.Fatal("DebugKeywords should not be empty")
	}

	expectedKeywords := []string{
		"debug", "trace", "verbose",
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range DebugKeywords {
			if strings.Contains(keyword, expected) || strings.Contains(expected, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected debug keyword '%s' not found", expected)
		}
	}
}

// TestSecurityPatterns tests security pattern keywords
func TestSecurityPatterns(t *testing.T) {
	if SecurityPatterns == nil {
		t.Fatal("SecurityPatterns should not be nil")
	}

	if len(SecurityPatterns) == 0 {
		t.Fatal("SecurityPatterns should not be empty")
	}

	expectedKeywords := []string{
		"unauthorized", "forbidden", "auth failed",
		"injection", "xss", "csrf", "breach",
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range SecurityPatterns {
			if keyword == expected || strings.Contains(keyword, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected security pattern '%s' not found", expected)
		}
	}
}

func TestSecurityPatterns_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, pattern := range SecurityPatterns {
		if seen[pattern] {
			t.Errorf("Duplicate security pattern found: %s", pattern)
		}
		seen[pattern] = true
	}
}

// TestKeywordSeparation tests that keywords don't overlap inappropriately
func TestKeywordSeparation(t *testing.T) {
	for _, errorKw := range ErrorKeywords {
		for _, successKw := range SuccessKeywords {
			if errorKw == successKw {
				t.Errorf("Keyword '%s' appears in both Error and Success lists", errorKw)
			}
		}
	}
}

// TestPatternCoverage tests overall pattern coverage
func TestPatternCoverage(t *testing.T) {
	totalKeywords := len(ErrorKeywords) + len(WarningKeywords) +
		len(SuccessKeywords) + len(DebugKeywords)

	if totalKeywords < 50 {
		t.Errorf("Expected at least 50 total keywords, got %d", totalKeywords)
	}
	t.Logf("Total keywords: %d", totalKeywords)
}

// TestHTTPStatusCodeRanges tests coverage of different HTTP code ranges
func TestHTTPStatusCodeRanges(t *testing.T) {
	ranges := map[string]int{
		"2xx": 0,
		"3xx": 0,
		"4xx": 0,
		"5xx": 0,
	}

	for code := range HTTPStatusSeverity {
		if len(code) != 3 {
			t.Errorf("Invalid status code length: %s", code)
			continue
		}

		switch code[0] {
		case '2':
			ranges["2xx"]++
		case '3':
			ranges["3xx"]++
		case '4':
			ranges["4xx"]++
		case '5':
			ranges["5xx"]++
		}
	}

	for rangeKey, count := range ranges {
		if count == 0 {
			t.Errorf("No HTTP status codes in range %s", rangeKey)
		} else {
			t.Logf("HTTP %s coverage: %d codes", rangeKey, count)
		}
	}
}

// TestDatabasePatterns tests database pattern keywords
func TestDatabasePatterns(t *testing.T) {
	if DatabasePatterns == nil {
		t.Fatal("DatabasePatterns should not be nil")
	}

	if len(DatabasePatterns) == 0 {
		t.Fatal("DatabasePatterns should not be empty")
	}

	// Check for common database patterns
	expectedPatterns := []string{
		"deadlock", "connection", "constraint",
	}

	for _, expected := range expectedPatterns {
		found := false
		for pattern := range DatabasePatterns {
			if strings.Contains(pattern, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected database pattern containing '%s' not found", expected)
		}
	}
}

// TestPerformancePatterns tests performance pattern keywords
func TestPerformancePatterns(t *testing.T) {
	if PerformancePatterns == nil {
		t.Fatal("PerformancePatterns should not be nil")
	}

	if len(PerformancePatterns) == 0 {
		t.Fatal("PerformancePatterns should not be empty")
	}
}

// TestPerformanceThresholds tests performance thresholds
func TestPerformanceThresholds(t *testing.T) {
	if PerformanceThresholds == nil {
		t.Fatal("PerformanceThresholds should not be nil")
	}

	if PerformanceThresholds["fast"] >= PerformanceThresholds["normal"] {
		t.Error("fast threshold should be less than normal")
	}
	if PerformanceThresholds["normal"] >= PerformanceThresholds["slow"] {
		t.Error("normal threshold should be less than slow")
	}
}

// TestCategorizePerformance tests performance categorization
func TestCategorizePerformance(t *testing.T) {
	tests := []struct {
		durationMs int
		expected   string
	}{
		{50, "fast"},
		{500, "normal"},
		{2000, "slow"},
		{6000, "critical"},
	}

	for _, tt := range tests {
		result := CategorizePerformance(tt.durationMs)
		if result != tt.expected {
			t.Errorf("CategorizePerformance(%d) = %s, want %s", tt.durationMs, result, tt.expected)
		}
	}
}

// TestPerformanceSeverity tests performance level to severity mapping
func TestPerformanceSeverity(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"fast", "success"},
		{"normal", "info"},
		{"slow", "warning"},
		{"critical", "error"},
		{"unknown", "info"},
	}

	for _, tt := range tests {
		result := PerformanceSeverity(tt.level)
		if result != tt.expected {
			t.Errorf("PerformanceSeverity(%s) = %s, want %s", tt.level, result, tt.expected)
		}
	}
}

// TestSystemErrorCodes tests system error code mappings
func TestSystemErrorCodes(t *testing.T) {
	if SystemErrorCodes == nil {
		t.Fatal("SystemErrorCodes should not be nil")
	}

	if len(SystemErrorCodes) == 0 {
		t.Fatal("SystemErrorCodes should not be empty")
	}

	// Check for common system error codes
	expectedCodes := []string{
		"ECONNREFUSED", "ETIMEDOUT", "ENOMEM",
	}

	for _, expected := range expectedCodes {
		if _, exists := SystemErrorCodes[expected]; !exists {
			t.Errorf("Expected system error code '%s' not found", expected)
		}
	}
}

// TestBusinessPatterns tests business pattern keywords
func TestBusinessPatterns(t *testing.T) {
	if BusinessPatterns == nil {
		t.Fatal("BusinessPatterns should not be nil")
	}

	if len(BusinessPatterns) == 0 {
		t.Fatal("BusinessPatterns should not be empty")
	}

	// Check for common business patterns
	expectedPatterns := []string{
		"payment", "order", "user",
	}

	for _, expected := range expectedPatterns {
		found := false
		for pattern := range BusinessPatterns {
			if strings.Contains(pattern, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected business pattern containing '%s' not found", expected)
		}
	}
}
