package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Server defaults
	if config.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Server.Port)
	}
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", config.Server.Host)
	}
	if config.Server.ReadTimeout != 15 {
		t.Errorf("expected read timeout 15, got %d", config.Server.ReadTimeout)
	}

	// Database defaults
	if config.Database.RetentionDays != 90 {
		t.Errorf("expected retention 90 days, got %d", config.Database.RetentionDays)
	}

	// Logging defaults
	if config.Logging.DefaultSeverity != "info" {
		t.Errorf("expected severity info, got %s", config.Logging.DefaultSeverity)
	}

	// Output defaults
	if config.Output.Format != "table" {
		t.Errorf("expected format table, got %s", config.Output.Format)
	}
	if config.Output.NoColor != false {
		t.Error("expected NoColor false")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Should return defaults when no config file exists
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", config.Server.Port)
	}
}

func TestLoadConfig_WithFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {
			"port": 9090,
			"host": "127.0.0.1"
		},
		"database": {
			"retention_days": 30
		},
		"output": {
			"format": "json",
			"no_color": true
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil { //nolint:gosec // Test file
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.Server.Port)
	}
	if config.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", config.Server.Host)
	}
	if config.Database.RetentionDays != 30 {
		t.Errorf("expected retention 30, got %d", config.Database.RetentionDays)
	}
	if config.Output.Format != "json" {
		t.Errorf("expected format json, got %s", config.Output.Format)
	}
	if config.Output.NoColor != true {
		t.Error("expected NoColor true")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil { //nolint:gosec // Test file
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadEnvConfig(t *testing.T) {
	config := DefaultConfig()

	// Set environment variables
	os.Setenv("SCRIBE_PORT", "3000")
	os.Setenv("SCRIBE_HOST", "localhost")
	os.Setenv("SCRIBE_DB_PATH", "/tmp/test.db")
	os.Setenv("SCRIBE_RETENTION_DAYS", "7")
	os.Setenv("SCRIBE_DEFAULT_SEVERITY", "debug")
	os.Setenv("SCRIBE_OUTPUT_FORMAT", "plain")
	os.Setenv("SCRIBE_NO_COLOR", "true")
	os.Setenv("SCRIBE_VERBOSE", "1")
	defer func() {
		os.Unsetenv("SCRIBE_PORT")
		os.Unsetenv("SCRIBE_HOST")
		os.Unsetenv("SCRIBE_DB_PATH")
		os.Unsetenv("SCRIBE_RETENTION_DAYS")
		os.Unsetenv("SCRIBE_DEFAULT_SEVERITY")
		os.Unsetenv("SCRIBE_OUTPUT_FORMAT")
		os.Unsetenv("SCRIBE_NO_COLOR")
		os.Unsetenv("SCRIBE_VERBOSE")
	}()

	loadEnvConfig(config)

	if config.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", config.Server.Port)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", config.Server.Host)
	}
	if config.Database.Path != "/tmp/test.db" {
		t.Errorf("expected db path /tmp/test.db, got %s", config.Database.Path)
	}
	if config.Database.RetentionDays != 7 {
		t.Errorf("expected retention 7, got %d", config.Database.RetentionDays)
	}
	if config.Logging.DefaultSeverity != "debug" {
		t.Errorf("expected severity debug, got %s", config.Logging.DefaultSeverity)
	}
	if config.Output.Format != "plain" {
		t.Errorf("expected format plain, got %s", config.Output.Format)
	}
	if config.Output.NoColor != true {
		t.Error("expected NoColor true")
	}
	if config.Output.Verbose != true {
		t.Error("expected Verbose true")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.json")

	config := DefaultConfig()
	config.Server.Port = 5000

	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.Server.Port != 5000 {
		t.Errorf("expected port 5000, got %d", loaded.Server.Port)
	}
}

func TestGetSetConfig(t *testing.T) {
	// Reset global config
	appConfig = nil

	// Get should return default
	config := GetConfig()
	if config.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", config.Server.Port)
	}

	// Set custom config
	custom := &Config{
		Server: ServerConfig{Port: 4000},
	}
	SetConfig(custom)

	// Get should return custom
	config = GetConfig()
	if config.Server.Port != 4000 {
		t.Errorf("expected port 4000, got %d", config.Server.Port)
	}

	// Cleanup
	appConfig = nil
}

func TestGetDefaultConfigPaths(t *testing.T) {
	paths := getDefaultConfigPaths()

	if len(paths) == 0 {
		t.Error("expected at least one default config path")
	}

	// Should include current directory options
	hasCurrentDir := false
	for _, p := range paths {
		if p == "scribe.json" || p == ".scribe.json" {
			hasCurrentDir = true
			break
		}
	}
	if !hasCurrentDir {
		t.Error("expected current directory config paths")
	}
}
