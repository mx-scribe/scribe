package entities

import (
	"time"

	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

// Log represents a complete log entry with structured header and flexible body.
type Log struct {
	ID        int64          `json:"id"`
	Header    LogHeader      `json:"header"`
	Body      map[string]any `json:"body"`
	Metadata  LogMetadata    `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// LogHeader contains structured metadata - only title is required.
type LogHeader struct {
	Title       string                `json:"title"`
	Severity    valueobjects.Severity `json:"severity,omitempty"`
	Source      string                `json:"source,omitempty"`
	Color       valueobjects.Color    `json:"color,omitempty"`
	Description string                `json:"description,omitempty"`
}

// LogMetadata contains smart derived metadata from log analysis.
type LogMetadata struct {
	DerivedSeverity string `json:"derived_severity,omitempty"`
	DerivedSource   string `json:"derived_source,omitempty"`
	DerivedCategory string `json:"derived_category,omitempty"`
}

// NewLog creates a new log entry with the given header and body.
func NewLog(header LogHeader, body map[string]any) *Log {
	return &Log{
		Header:    header,
		Body:      body,
		Metadata:  LogMetadata{},
		CreatedAt: time.Now(),
	}
}

// Validate checks if the log entry is valid.
func (l *Log) Validate() error {
	if l.Header.Title == "" {
		return ErrMissingTitle
	}
	return nil
}

// UpdateMetadata updates the log's derived metadata.
func (l *Log) UpdateMetadata(metadata LogMetadata) {
	l.Metadata = metadata
}

// EffectiveSeverity returns the most specific severity available.
func (l *Log) EffectiveSeverity() valueobjects.Severity {
	if l.Metadata.DerivedSeverity != "" {
		return valueobjects.Severity(l.Metadata.DerivedSeverity)
	}
	if l.Header.Severity != "" {
		return l.Header.Severity
	}
	return valueobjects.DefaultSeverity()
}

// EffectiveSource returns the most specific source available.
func (l *Log) EffectiveSource() string {
	if l.Metadata.DerivedSource != "" {
		return l.Metadata.DerivedSource
	}
	return l.Header.Source
}

// EffectiveColor returns the color to use for this log.
func (l *Log) EffectiveColor() valueobjects.Color {
	if l.Header.Color != "" && l.Header.Color.IsValid() {
		return l.Header.Color
	}
	return valueobjects.AutoAssignColor(l.EffectiveSeverity())
}
