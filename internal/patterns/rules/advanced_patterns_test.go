package rules

import (
	"testing"
)

// TestDatabasePatterns_Detailed tests database error pattern mapping in detail.
func TestDatabasePatterns_Detailed(t *testing.T) {
	tests := []struct {
		pattern          string
		expectedSeverity string
	}{
		{"deadlock", "critical"},
		{"connection pool exhausted", "critical"},
		{"too many connections", "critical"},
		{"duplicate key", "warning"},
		{"constraint violation", "error"},
		{"foreign key violation", "error"},
		{"sqlite_corrupt", "critical"},
		{"sqlite_busy", "warning"},
	}

	for _, tt := range tests {
		t.Run("Pattern: "+tt.pattern, func(t *testing.T) {
			severity, exists := DatabasePatterns[tt.pattern]
			if !exists {
				t.Errorf("Pattern '%s' not found in mapping", tt.pattern)
				return
			}
			if severity != tt.expectedSeverity {
				t.Errorf("Pattern '%s': expected severity '%s', got '%s'",
					tt.pattern, tt.expectedSeverity, severity)
			}
		})
	}
}

func TestDatabasePatterns_CriticalErrors(t *testing.T) {
	criticalPatterns := []string{"deadlock", "connection pool exhausted", "sqlite_corrupt"}

	for _, pattern := range criticalPatterns {
		severity, exists := DatabasePatterns[pattern]
		if !exists {
			t.Errorf("Critical database pattern '%s' not found", pattern)
			continue
		}
		if severity != "critical" {
			t.Errorf("Pattern '%s' should be critical, got '%s'", pattern, severity)
		}
	}
}

func TestDatabasePatterns_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for pattern := range DatabasePatterns {
		if seen[pattern] {
			t.Errorf("Duplicate database pattern found: %s", pattern)
		}
		seen[pattern] = true
	}
}

// TestSystemErrorCodes_Detailed tests system error code mapping in detail.
func TestSystemErrorCodes_Detailed(t *testing.T) {
	tests := []struct {
		code             string
		expectedSeverity string
	}{
		{"ECONNREFUSED", "error"},
		{"ETIMEDOUT", "error"},
		{"ENOTFOUND", "error"},
		{"ECONNRESET", "error"},
		{"EMFILE", "critical"},
		{"ENOMEM", "critical"},
		{"ENOSPC", "critical"},
		{"ENOENT", "warning"},
	}

	for _, tt := range tests {
		t.Run("Code: "+tt.code, func(t *testing.T) {
			severity, exists := SystemErrorCodes[tt.code]
			if !exists {
				t.Errorf("Error code '%s' not found in mapping", tt.code)
				return
			}
			if severity != tt.expectedSeverity {
				t.Errorf("Code '%s': expected severity '%s', got '%s'",
					tt.code, tt.expectedSeverity, severity)
			}
		})
	}
}

func TestSystemErrorCodes_CriticalErrors(t *testing.T) {
	criticalCodes := []string{"EMFILE", "ENOMEM", "ENOSPC", "EIO"}

	for _, code := range criticalCodes {
		severity, exists := SystemErrorCodes[code]
		if !exists {
			t.Errorf("Critical error code '%s' not found", code)
			continue
		}
		if severity != "critical" {
			t.Errorf("Code '%s' should be critical, got '%s'", code, severity)
		}
	}
}

func TestSystemErrorCodes_AllUppercase(t *testing.T) {
	for code := range SystemErrorCodes {
		hasLowercase := false
		for _, char := range code {
			if char >= 'a' && char <= 'z' {
				hasLowercase = true
				break
			}
		}
		if hasLowercase {
			t.Errorf("System error code should be uppercase: %s", code)
		}
	}
}

// TestPerformanceThresholds_Detailed tests performance threshold configuration.
func TestPerformanceThresholds_Detailed(t *testing.T) {
	expectedLevels := []string{"fast", "normal", "slow", "critical"}
	for _, level := range expectedLevels {
		if _, exists := PerformanceThresholds[level]; !exists {
			t.Errorf("Performance level '%s' not found", level)
		}
	}

	if PerformanceThresholds["slow"] >= PerformanceThresholds["critical"] {
		t.Error("'slow' threshold should be less than 'critical'")
	}
}

func TestPerformanceThresholds_ReasonableValues(t *testing.T) {
	if PerformanceThresholds["fast"] <= 0 {
		t.Error("'fast' threshold should be positive")
	}
	if PerformanceThresholds["critical"] > 60000 {
		t.Error("'critical' threshold seems too high (> 1 minute)")
	}
}

// TestCategorizePerformance_Detailed tests the performance categorization function.
func TestCategorizePerformance_Detailed(t *testing.T) {
	tests := []struct {
		durationMs int
		expected   string
	}{
		{50, "fast"},
		{99, "fast"},
		{100, "normal"},
		{500, "normal"},
		{999, "normal"},
		{1000, "slow"},
		{2000, "slow"},
		{2999, "slow"},
		{3000, "critical"},
		{5000, "critical"},
		{10000, "critical"},
	}

	for _, tt := range tests {
		result := CategorizePerformance(tt.durationMs)
		if result != tt.expected {
			t.Errorf("CategorizePerformance(%d) = %s, expected %s",
				tt.durationMs, result, tt.expected)
		}
	}
}

func TestCategorizePerformance_Boundaries(t *testing.T) {
	tests := []struct {
		durationMs int
		expected   string
	}{
		{PerformanceThresholds["fast"] - 1, "fast"},
		{PerformanceThresholds["fast"], "normal"},
		{PerformanceThresholds["normal"] - 1, "normal"},
		{PerformanceThresholds["normal"], "slow"},
		{PerformanceThresholds["slow"] - 1, "slow"},
		{PerformanceThresholds["slow"], "critical"},
	}

	for _, tt := range tests {
		result := CategorizePerformance(tt.durationMs)
		if result != tt.expected {
			t.Errorf("Boundary test: CategorizePerformance(%d) = %s, expected %s",
				tt.durationMs, result, tt.expected)
		}
	}
}

func TestCategorizePerformance_ZeroAndNegative(t *testing.T) {
	if result := CategorizePerformance(0); result != "fast" {
		t.Errorf("CategorizePerformance(0) should return 'fast', got '%s'", result)
	}

	if result := CategorizePerformance(-100); result != "fast" {
		t.Errorf("CategorizePerformance(-100) should return 'fast', got '%s'", result)
	}
}

// TestBusinessPatterns_Detailed tests business logic pattern mapping.
func TestBusinessPatterns_Detailed(t *testing.T) {
	tests := []struct {
		pattern          string
		expectedSeverity string
	}{
		{"payment failed", "error"},
		{"payment successful", "success"},
		{"payment pending", "info"},
		{"order completed", "success"},
		{"order canceled", "warning"},
		{"subscription expired", "warning"},
		{"login successful", "success"},
		{"login failed", "warning"},
	}

	for _, tt := range tests {
		t.Run("Pattern: "+tt.pattern, func(t *testing.T) {
			severity, exists := BusinessPatterns[tt.pattern]
			if !exists {
				t.Errorf("Pattern '%s' not found in mapping", tt.pattern)
				return
			}
			if severity != tt.expectedSeverity {
				t.Errorf("Pattern '%s': expected severity '%s', got '%s'",
					tt.pattern, tt.expectedSeverity, severity)
			}
		})
	}
}

func TestBusinessPatterns_PaymentCoverage(t *testing.T) {
	paymentPatterns := []string{"payment failed", "payment successful", "payment pending"}

	found := 0
	for _, pattern := range paymentPatterns {
		if _, exists := BusinessPatterns[pattern]; exists {
			found++
		}
	}

	if found < len(paymentPatterns) {
		t.Errorf("Expected all %d payment patterns, found %d", len(paymentPatterns), found)
	}
}

func TestBusinessPatterns_OrderCoverage(t *testing.T) {
	orderPatterns := []string{"order completed", "order canceled", "order failed"}

	found := 0
	for _, pattern := range orderPatterns {
		if _, exists := BusinessPatterns[pattern]; exists {
			found++
		}
	}

	if found < len(orderPatterns) {
		t.Errorf("Expected all %d order patterns, found %d", len(orderPatterns), found)
	}
}

func TestBusinessPatterns_ValidSeverities(t *testing.T) {
	validSeverities := map[string]bool{
		"success": true,
		"info":    true,
		"warning": true,
		"error":   true,
	}

	for pattern, severity := range BusinessPatterns {
		if !validSeverities[severity] {
			t.Errorf("Business pattern '%s' has invalid severity: %s", pattern, severity)
		}
	}
}

func TestBusinessPatterns_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for pattern := range BusinessPatterns {
		if seen[pattern] {
			t.Errorf("Duplicate business pattern found: %s", pattern)
		}
		seen[pattern] = true
	}
}

// TestAllPatterns_ComprehensiveCoverage tests overall pattern coverage.
func TestAllPatterns_ComprehensiveCoverage(t *testing.T) {
	totalPatterns := len(HTTPStatusSeverity) + len(DatabasePatterns) +
		len(SystemErrorCodes) + len(BusinessPatterns)

	t.Logf("Total pattern mappings: %d", totalPatterns)

	if totalPatterns < 50 {
		t.Errorf("Expected at least 50 pattern mappings, got %d", totalPatterns)
	}
}
