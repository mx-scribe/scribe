package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewOutput(t *testing.T) {
	// Reset global config
	appConfig = nil
	defer func() { appConfig = nil }()

	out := NewOutput()
	if out == nil {
		t.Fatal("NewOutput returned nil")
	}
	if out.Writer == nil {
		t.Error("expected non-nil Writer")
	}
}

func TestOutput_PrintJSON(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer: &buf,
		Format: FormatJSON,
	}

	data := map[string]string{"key": "value"}
	if err := out.Print(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result)
	}
}

func TestOutput_PrintPlain_String(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer: &buf,
		Format: FormatPlain,
	}

	if err := out.Print("hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(buf.String()) != "hello world" {
		t.Errorf("expected 'hello world', got %q", buf.String())
	}
}

func TestOutput_PrintPlain_TableData(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer: &buf,
		Format: FormatPlain,
	}

	data := TableData{
		Headers: []string{"ID", "Name"},
		Rows: []TableRow{
			{Values: []string{"1", "Alice"}},
			{Values: []string{"2", "Bob"}},
		},
	}

	if err := out.Print(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1 Alice") {
		t.Errorf("expected '1 Alice' in output, got %q", output)
	}
	if !strings.Contains(output, "2 Bob") {
		t.Errorf("expected '2 Bob' in output, got %q", output)
	}
}

func TestOutput_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		Format:  FormatTable,
		NoColor: true,
	}

	data := TableData{
		Headers: []string{"ID", "Name"},
		Rows: []TableRow{
			{Values: []string{"1", "Alice"}},
			{Values: []string{"2", "Bob"}},
		},
	}

	if err := out.Print(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID") {
		t.Errorf("expected header 'ID' in output, got %q", output)
	}
	if !strings.Contains(output, "Name") {
		t.Errorf("expected header 'Name' in output, got %q", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected 'Alice' in output, got %q", output)
	}
}

func TestOutput_PrintTable_WithColor(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		Format:  FormatTable,
		NoColor: false,
	}

	data := TableData{
		Headers: []string{"Status"},
		Rows: []TableRow{
			{Values: []string{"error"}, Color: ColorRed},
		},
	}

	if err := out.Print(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain ANSI codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("expected ANSI color codes in output")
	}
}

func TestOutput_Success(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: true,
	}

	out.Success("operation completed")

	if !strings.Contains(buf.String(), "✓") {
		t.Error("expected checkmark in success message")
	}
	if !strings.Contains(buf.String(), "operation completed") {
		t.Error("expected message in output")
	}
}

func TestOutput_Error(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: true,
	}

	out.Error("something failed")

	if !strings.Contains(buf.String(), "✗") {
		t.Error("expected X mark in error message")
	}
	if !strings.Contains(buf.String(), "something failed") {
		t.Error("expected message in output")
	}
}

func TestOutput_Info(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: true,
	}

	out.Info("information")

	if !strings.Contains(buf.String(), "ℹ") {
		t.Error("expected info symbol in message")
	}
}

func TestOutput_Warning(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: true,
	}

	out.Warning("be careful")

	if !strings.Contains(buf.String(), "⚠") {
		t.Error("expected warning symbol in message")
	}
}

func TestOutput_Verbose(t *testing.T) {
	// Reset verbose flag
	verbose = false
	defer func() { verbose = false }()

	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: true,
	}

	// Verbose disabled - should not print
	out.Verbose("debug info")
	if buf.Len() > 0 {
		t.Error("expected no output when verbose disabled")
	}

	// Enable verbose
	verbose = true
	out.Verbose("debug info")
	if !strings.Contains(buf.String(), "debug info") {
		t.Error("expected message when verbose enabled")
	}
}

func TestOutput_Success_WithColor(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		NoColor: false,
	}

	out.Success("colored success")

	if !strings.Contains(buf.String(), ColorGreen) {
		t.Error("expected green color in success message")
	}
}

func TestSeverityColor(t *testing.T) {
	tests := []struct {
		severity string
		color    string
	}{
		{"critical", ColorRed},
		{"error", ColorRed},
		{"warning", ColorYellow},
		{"success", ColorGreen},
		{"info", ColorBlue},
		{"debug", ColorGray},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			color := SeverityColor(tt.severity)
			if color != tt.color {
				t.Errorf("SeverityColor(%s) = %q, want %q", tt.severity, color, tt.color)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	result := colorize("test", ColorRed)
	if !strings.HasPrefix(result, ColorRed) {
		t.Error("expected color prefix")
	}
	if !strings.HasSuffix(result, ColorReset) {
		t.Error("expected color reset suffix")
	}
	if !strings.Contains(result, "test") {
		t.Error("expected text content")
	}

	// Empty color should return text unchanged
	result = colorize("plain", "")
	if result != "plain" {
		t.Errorf("expected 'plain', got %q", result)
	}
}

func TestOutput_PrintTable_Rows(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer:  &buf,
		Format:  FormatTable,
		NoColor: true,
	}

	rows := []TableRow{
		{Values: []string{"a", "b"}},
		{Values: []string{"c", "d"}},
	}

	if err := out.Print(rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "a") {
		t.Error("expected 'a' in output")
	}
}

func TestOutput_PrintPlain_Rows(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer: &buf,
		Format: FormatPlain,
	}

	rows := []TableRow{
		{Values: []string{"x", "y"}},
	}

	if err := out.Print(rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "x y") {
		t.Errorf("expected 'x y', got %q", buf.String())
	}
}

func TestOutput_PrintPlain_FallbackJSON(t *testing.T) {
	var buf bytes.Buffer
	out := &Output{
		Writer: &buf,
		Format: FormatPlain,
	}

	// Complex type should fall back to JSON
	data := struct {
		Field string `json:"field"`
	}{Field: "value"}

	if err := out.Print(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "field") {
		t.Error("expected JSON field in output")
	}
}
