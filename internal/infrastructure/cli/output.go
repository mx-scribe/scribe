package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// OutputFormat represents the output format type.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatPlain OutputFormat = "plain"
)

// Output handles formatted output to the terminal.
type Output struct {
	Writer     io.Writer
	Format     OutputFormat
	NoColor    bool
	TimeFormat string
}

// NewOutput creates a new Output with the current settings.
func NewOutput() *Output {
	config := GetConfig()
	return &Output{
		Writer:     os.Stdout,
		Format:     OutputFormat(GetOutputFormat()),
		NoColor:    IsNoColor(),
		TimeFormat: config.Output.TimeFormat,
	}
}

// Print outputs data in the configured format.
func (o *Output) Print(data any) error {
	switch o.Format {
	case FormatJSON:
		return o.printJSON(data)
	case FormatPlain:
		return o.printPlain(data)
	default:
		return o.printTable(data)
	}
}

// printJSON outputs data as JSON.
func (o *Output) printJSON(data any) error {
	encoder := json.NewEncoder(o.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printPlain outputs data as plain text.
func (o *Output) printPlain(data any) error {
	switch v := data.(type) {
	case []TableRow:
		for _, row := range v {
			fmt.Fprintln(o.Writer, strings.Join(row.Values, " "))
		}
	case TableData:
		for _, row := range v.Rows {
			fmt.Fprintln(o.Writer, strings.Join(row.Values, " "))
		}
	case string:
		fmt.Fprintln(o.Writer, v)
	default:
		// Fall back to JSON for complex types
		return o.printJSON(data)
	}
	return nil
}

// printTable outputs data as a formatted table.
func (o *Output) printTable(data any) error {
	switch v := data.(type) {
	case TableData:
		return o.writeTable(v.Headers, v.Rows)
	case []TableRow:
		// No headers, just rows
		return o.writeTable(nil, v)
	case string:
		fmt.Fprintln(o.Writer, v)
	default:
		// Fall back to JSON for complex types
		return o.printJSON(data)
	}
	return nil
}

// writeTable writes a formatted table.
func (o *Output) writeTable(headers []string, rows []TableRow) error {
	w := tabwriter.NewWriter(o.Writer, 0, 0, 2, ' ', 0)

	// Write headers if provided
	if len(headers) > 0 {
		headerLine := strings.Join(headers, "\t")
		if !o.NoColor {
			headerLine = colorize(headerLine, ColorBold)
		}
		fmt.Fprintln(w, headerLine)
	}

	// Write rows
	for _, row := range rows {
		values := row.Values
		if !o.NoColor && row.Color != "" {
			// Apply color to the entire row
			for i, v := range values {
				values[i] = colorize(v, row.Color)
			}
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	return w.Flush()
}

// TableData represents data to be displayed as a table.
type TableData struct {
	Headers []string
	Rows    []TableRow
}

// TableRow represents a single row in a table.
type TableRow struct {
	Values []string
	Color  string // Optional color for the row
}

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"
)

// SeverityColor returns the ANSI color code for a severity level.
func SeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return ColorRed
	case "error":
		return ColorRed
	case "warning":
		return ColorYellow
	case "success":
		return ColorGreen
	case "info":
		return ColorBlue
	case "debug":
		return ColorGray
	default:
		return ""
	}
}

// colorize wraps text with ANSI color codes.
func colorize(text, color string) string {
	if color == "" {
		return text
	}
	return color + text + ColorReset
}

// Success prints a success message.
func (o *Output) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !o.NoColor {
		msg = colorize("✓ "+msg, ColorGreen)
	} else {
		msg = "✓ " + msg
	}
	fmt.Fprintln(o.Writer, msg)
}

// Error prints an error message.
func (o *Output) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !o.NoColor {
		msg = colorize("✗ "+msg, ColorRed)
	} else {
		msg = "✗ " + msg
	}
	fmt.Fprintln(o.Writer, msg)
}

// Info prints an info message.
func (o *Output) Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !o.NoColor {
		msg = colorize("ℹ "+msg, ColorBlue)
	} else {
		msg = "ℹ " + msg
	}
	fmt.Fprintln(o.Writer, msg)
}

// Warning prints a warning message.
func (o *Output) Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !o.NoColor {
		msg = colorize("⚠ "+msg, ColorYellow)
	} else {
		msg = "⚠ " + msg
	}
	fmt.Fprintln(o.Writer, msg)
}

// Verbose prints a message only if verbose mode is enabled.
func (o *Output) Verbose(format string, args ...any) {
	if IsVerbose() {
		msg := fmt.Sprintf(format, args...)
		if !o.NoColor {
			msg = colorize(msg, ColorGray)
		}
		fmt.Fprintln(o.Writer, msg)
	}
}
