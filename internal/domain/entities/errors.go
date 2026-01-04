package entities

import "errors"

var (
	// ErrMissingTitle is returned when a log is created without a title.
	ErrMissingTitle = errors.New("log title is required")

	// ErrLogNotFound is returned when a log cannot be found.
	ErrLogNotFound = errors.New("log not found")
)
