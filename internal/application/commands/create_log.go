package commands

import (
	"github.com/mx-scribe/scribe/internal/domain/entities"
	"github.com/mx-scribe/scribe/internal/domain/services"
	"github.com/mx-scribe/scribe/internal/domain/valueobjects"
)

// CreateLogInput represents the input for creating a log.
type CreateLogInput struct {
	Title       string         `json:"title"`
	Severity    string         `json:"severity,omitempty"`
	Source      string         `json:"source,omitempty"`
	Color       string         `json:"color,omitempty"`
	Description string         `json:"description,omitempty"`
	Body        map[string]any `json:"body,omitempty"`
}

// CreateLogOutput represents the output after creating a log.
type CreateLogOutput struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Severity  string `json:"severity"`
	CreatedAt string `json:"created_at"`
}

// LogRepository defines the interface for log persistence.
type LogRepository interface {
	Create(log *entities.Log) error
}

// CreateLogHandler handles the create log command.
type CreateLogHandler struct {
	repo LogRepository
}

// NewCreateLogHandler creates a new create log handler.
func NewCreateLogHandler(repo LogRepository) *CreateLogHandler {
	return &CreateLogHandler{repo: repo}
}

// Handle executes the create log command.
func (h *CreateLogHandler) Handle(input CreateLogInput) (*CreateLogOutput, error) {
	// Build header
	header := entities.LogHeader{
		Title:       input.Title,
		Severity:    valueobjects.SeverityFromString(input.Severity),
		Source:      input.Source,
		Color:       valueobjects.ColorFromString(input.Color),
		Description: input.Description,
	}

	// Build body
	body := input.Body
	if body == nil {
		body = make(map[string]any)
	}

	// Create log entity
	log := entities.NewLog(header, body)

	// Validate
	if err := log.Validate(); err != nil {
		return nil, err
	}

	// Run pattern matching to derive metadata
	matcher := services.NewPatternMatcher()
	metadata := matcher.AnalyzeLog(log)

	// Apply derived metadata only if not already set
	if log.Header.Severity == "" || log.Header.Severity == valueobjects.SeverityInfo {
		if metadata.DerivedSeverity != "" && metadata.DerivedSeverity != "info" {
			log.Metadata.DerivedSeverity = metadata.DerivedSeverity
		}
	}
	if log.Header.Source == "" && metadata.DerivedSource != "" {
		log.Metadata.DerivedSource = metadata.DerivedSource
	}
	if metadata.DerivedCategory != "" {
		log.Metadata.DerivedCategory = metadata.DerivedCategory
	}

	// Persist
	if err := h.repo.Create(log); err != nil {
		return nil, err
	}

	// Return output
	return &CreateLogOutput{
		ID:        log.ID,
		Title:     log.Header.Title,
		Severity:  log.EffectiveSeverity().String(),
		CreatedAt: log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
