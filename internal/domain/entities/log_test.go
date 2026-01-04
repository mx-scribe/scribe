package entities

import (
	"testing"

	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

func TestNewLog(t *testing.T) {
	header := LogHeader{
		Title:    "Test log",
		Severity: valueobjects.SeverityInfo,
	}
	body := map[string]any{"key": "value"}

	log := NewLog(header, body)

	if log.Header.Title != "Test log" {
		t.Errorf("expected title 'Test log', got '%s'", log.Header.Title)
	}
	if log.Body["key"] != "value" {
		t.Errorf("expected body key 'value', got '%v'", log.Body["key"])
	}
	if log.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestLog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		header  LogHeader
		wantErr error
	}{
		{
			name:    "valid log with title",
			header:  LogHeader{Title: "Test"},
			wantErr: nil,
		},
		{
			name:    "invalid log without title",
			header:  LogHeader{},
			wantErr: ErrMissingTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewLog(tt.header, nil)
			err := log.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLog_EffectiveSeverity(t *testing.T) {
	tests := []struct {
		name     string
		header   LogHeader
		metadata LogMetadata
		want     valueobjects.Severity
	}{
		{
			name:     "uses derived severity first",
			header:   LogHeader{Title: "Test", Severity: valueobjects.SeverityInfo},
			metadata: LogMetadata{DerivedSeverity: "error"},
			want:     valueobjects.SeverityError,
		},
		{
			name:     "uses header severity if no derived",
			header:   LogHeader{Title: "Test", Severity: valueobjects.SeverityWarning},
			metadata: LogMetadata{},
			want:     valueobjects.SeverityWarning,
		},
		{
			name:     "uses default if none set",
			header:   LogHeader{Title: "Test"},
			metadata: LogMetadata{},
			want:     valueobjects.SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewLog(tt.header, nil)
			log.UpdateMetadata(tt.metadata)
			if got := log.EffectiveSeverity(); got != tt.want {
				t.Errorf("EffectiveSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_EffectiveColor(t *testing.T) {
	tests := []struct {
		name   string
		header LogHeader
		want   valueobjects.Color
	}{
		{
			name:   "uses explicit color",
			header: LogHeader{Title: "Test", Color: "purple"},
			want:   "purple",
		},
		{
			name:   "auto-assigns red for error",
			header: LogHeader{Title: "Test", Severity: valueobjects.SeverityError},
			want:   "red",
		},
		{
			name:   "auto-assigns green for success",
			header: LogHeader{Title: "Test", Severity: valueobjects.SeveritySuccess},
			want:   "green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewLog(tt.header, nil)
			if got := log.EffectiveColor(); got != tt.want {
				t.Errorf("EffectiveColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
