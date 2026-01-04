package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Server ServerConfig `json:"server"`

	// Database settings
	Database DatabaseConfig `json:"database"`

	// Logging settings
	Logging LoggingConfig `json:"logging"`

	// Output settings
	Output OutputConfig `json:"output"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Path          string `json:"path"`
	RetentionDays int    `json:"retention_days"`
}

// LoggingConfig holds logging defaults.
type LoggingConfig struct {
	DefaultSeverity string `json:"default_severity"`
	DefaultSource   string `json:"default_source"`
}

// OutputConfig holds output settings.
type OutputConfig struct {
	Format     string `json:"format"`
	NoColor    bool   `json:"no_color"`
	Verbose    bool   `json:"verbose"`
	TimeFormat string `json:"time_format"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  15,
			WriteTimeout: 15,
		},
		Database: DatabaseConfig{
			Path:          filepath.Join(homeDir, ".scribe", "scribe.db"),
			RetentionDays: 90,
		},
		Logging: LoggingConfig{
			DefaultSeverity: "info",
			DefaultSource:   "",
		},
		Output: OutputConfig{
			Format:     "table",
			NoColor:    false,
			Verbose:    false,
			TimeFormat: "2006-01-02 15:04:05",
		},
	}
}

// LoadConfig loads configuration from file and environment variables.
// Priority: flags > environment variables > config file > defaults
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load from config file if specified
	if configPath != "" {
		if err := loadConfigFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Try default config locations
		defaultPaths := getDefaultConfigPaths()
		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				if err := loadConfigFile(config, path); err != nil {
					return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
				}
				break
			}
		}
	}

	// Override with environment variables
	loadEnvConfig(config)

	return config, nil
}

// getDefaultConfigPaths returns paths to check for config files.
func getDefaultConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	return []string{
		"scribe.json",
		".scribe.json",
		filepath.Join(homeDir, ".scribe", "config.json"),
		filepath.Join(homeDir, ".config", "scribe", "config.json"),
		"/etc/scribe/config.json",
	}
}

// loadConfigFile loads configuration from a JSON file.
func loadConfigFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// loadEnvConfig loads configuration from environment variables.
func loadEnvConfig(config *Config) {
	// Server
	if v := os.Getenv("SCRIBE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Server.Port = port
		}
	}
	if v := os.Getenv("SCRIBE_HOST"); v != "" {
		config.Server.Host = v
	}

	// Database
	if v := os.Getenv("SCRIBE_DB_PATH"); v != "" {
		config.Database.Path = v
	}
	if v := os.Getenv("SCRIBE_RETENTION_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil {
			config.Database.RetentionDays = days
		}
	}

	// Logging
	if v := os.Getenv("SCRIBE_DEFAULT_SEVERITY"); v != "" {
		config.Logging.DefaultSeverity = v
	}
	if v := os.Getenv("SCRIBE_DEFAULT_SOURCE"); v != "" {
		config.Logging.DefaultSource = v
	}

	// Output
	if v := os.Getenv("SCRIBE_OUTPUT_FORMAT"); v != "" {
		config.Output.Format = v
	}
	if v := os.Getenv("SCRIBE_NO_COLOR"); v != "" {
		config.Output.NoColor = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("SCRIBE_VERBOSE"); v != "" {
		config.Output.Verbose = strings.EqualFold(v, "true") || v == "1"
	}
}

// SaveConfig saves configuration to a file.
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil { //nolint:gosec // Config file should be readable
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// global config instance
var appConfig *Config

// GetConfig returns the loaded configuration.
func GetConfig() *Config {
	if appConfig == nil {
		appConfig = DefaultConfig()
	}
	return appConfig
}

// SetConfig sets the global configuration.
func SetConfig(config *Config) {
	appConfig = config
}
