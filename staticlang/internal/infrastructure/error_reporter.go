// Package infrastructure contains error reporting implementation
package infrastructure

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/sokoide/llvm5/staticlang/internal/domain"
)

// ConsoleErrorReporter implements ErrorReporter for console output
type ConsoleErrorReporter struct {
	errors      []domain.CompilerError
	warnings    []domain.CompilerError
	output      io.Writer
	sourceMap   map[string][]byte // filename -> source content
	maxErrors   int
	maxWarnings int
}

// NewConsoleErrorReporter creates a new console error reporter
func NewConsoleErrorReporter(output io.Writer) *ConsoleErrorReporter {
	if output == nil {
		output = os.Stderr
	}

	return &ConsoleErrorReporter{
		errors:      make([]domain.CompilerError, 0),
		warnings:    make([]domain.CompilerError, 0),
		output:      output,
		sourceMap:   make(map[string][]byte),
		maxErrors:   100, // Limit to prevent spam
		maxWarnings: 50,
	}
}

// SetSourceContent sets the source content for a file (for better error reporting)
func (er *ConsoleErrorReporter) SetSourceContent(filename string, content []byte) {
	er.sourceMap[filename] = content
}

// SetMaxErrors sets the maximum number of errors to report
func (er *ConsoleErrorReporter) SetMaxErrors(max int) {
	er.maxErrors = max
}

// SetMaxWarnings sets the maximum number of warnings to report
func (er *ConsoleErrorReporter) SetMaxWarnings(max int) {
	er.maxWarnings = max
}

// ReportError reports a compilation error
func (er *ConsoleErrorReporter) ReportError(err domain.CompilerError) {
	if len(er.errors) < er.maxErrors {
		er.errors = append(er.errors, err)
		er.printError(err, "error")
	}
}

// ReportWarning reports a compilation warning
func (er *ConsoleErrorReporter) ReportWarning(warning domain.CompilerError) {
	if len(er.warnings) < er.maxWarnings {
		er.warnings = append(er.warnings, warning)
		er.printError(warning, "warning")
	}
}

// HasErrors returns true if any errors have been reported
func (er *ConsoleErrorReporter) HasErrors() bool {
	return len(er.errors) > 0
}

// HasWarnings returns true if any warnings have been reported
func (er *ConsoleErrorReporter) HasWarnings() bool {
	return len(er.warnings) > 0
}

// GetErrors returns all reported errors
func (er *ConsoleErrorReporter) GetErrors() []domain.CompilerError {
	// Return a copy to prevent modification
	errors := make([]domain.CompilerError, len(er.errors))
	copy(errors, er.errors)
	return errors
}

// GetWarnings returns all reported warnings
func (er *ConsoleErrorReporter) GetWarnings() []domain.CompilerError {
	// Return a copy to prevent modification
	warnings := make([]domain.CompilerError, len(er.warnings))
	copy(warnings, er.warnings)
	return warnings
}

// Clear clears all errors and warnings
func (er *ConsoleErrorReporter) Clear() {
	er.errors = er.errors[:0]
	er.warnings = er.warnings[:0]
}

// PrintSummary prints a summary of all errors and warnings
func (er *ConsoleErrorReporter) PrintSummary() {
	if er.HasErrors() || er.HasWarnings() {
		fmt.Fprintf(er.output, "\n")

		if er.HasErrors() {
			fmt.Fprintf(er.output, "Found %d error(s)\n", len(er.errors))
		}

		if er.HasWarnings() {
			fmt.Fprintf(er.output, "Found %d warning(s)\n", len(er.warnings))
		}
	}
}

// printError prints a formatted error or warning message
func (er *ConsoleErrorReporter) printError(err domain.CompilerError, severity string) {
	// Print basic error information
	fmt.Fprintf(er.output, "%s: %s: %s\n", err.Location.String(), severity, err.Message)

	// Print source context if available
	if content, exists := er.sourceMap[err.Location.Start.Filename]; exists {
		er.printSourceContext(content, err.Location)
	}

	// Print additional context if available
	if err.Context != "" {
		fmt.Fprintf(er.output, "  Context: %s\n", err.Context)
	}

	// Print hints if available
	for _, hint := range err.Hints {
		fmt.Fprintf(er.output, "  Hint: %s\n", hint)
	}

	fmt.Fprintf(er.output, "\n")
}

// printSourceContext prints the relevant source code context
func (er *ConsoleErrorReporter) printSourceContext(content []byte, location domain.SourceRange) {
	lines := strings.Split(string(content), "\n")

	startLine := location.Start.Line - 1 // Convert to 0-based
	endLine := location.End.Line - 1

	// Validate line numbers
	if startLine < 0 || startLine >= len(lines) {
		return
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Print a few lines before the error for context
	contextStart := max(0, startLine-2)
	contextEnd := min(len(lines)-1, endLine+2)

	lineNumWidth := len(fmt.Sprintf("%d", contextEnd+1))

	for i := contextStart; i <= contextEnd; i++ {
		lineNum := i + 1
		prefix := fmt.Sprintf("%*d | ", lineNumWidth, lineNum)

		if i >= startLine && i <= endLine {
			// This is an error line - highlight it
			fmt.Fprintf(er.output, "%s%s\n", prefix, lines[i])

			if i == startLine {
				// Add error indicator
				indicator := strings.Repeat(" ", len(prefix))
				if location.Start.Column > 0 {
					indicator += strings.Repeat(" ", location.Start.Column-1)
				}

				// Calculate length of error indicator
				indicatorLength := 1
				if startLine == endLine && location.End.Column > location.Start.Column {
					indicatorLength = location.End.Column - location.Start.Column
				}

				indicator += strings.Repeat("^", indicatorLength)
				fmt.Fprintf(er.output, "%s\n", indicator)
			}
		} else {
			// Context line
			fmt.Fprintf(er.output, "%s%s\n", prefix, lines[i])
		}
	}
}

// SortedErrorReporter wraps another ErrorReporter and sorts errors by location
type SortedErrorReporter struct {
	underlying domain.ErrorReporter
	errors     []domain.CompilerError
	warnings   []domain.CompilerError
}

// NewSortedErrorReporter creates a new sorted error reporter
func NewSortedErrorReporter(underlying domain.ErrorReporter) *SortedErrorReporter {
	return &SortedErrorReporter{
		underlying: underlying,
		errors:     make([]domain.CompilerError, 0),
		warnings:   make([]domain.CompilerError, 0),
	}
}

// ReportError collects errors for later sorted reporting
func (ser *SortedErrorReporter) ReportError(err domain.CompilerError) {
	ser.errors = append(ser.errors, err)
}

// ReportWarning collects warnings for later sorted reporting
func (ser *SortedErrorReporter) ReportWarning(warning domain.CompilerError) {
	ser.warnings = append(ser.warnings, warning)
}

// HasErrors returns true if any errors have been reported
func (ser *SortedErrorReporter) HasErrors() bool {
	return len(ser.errors) > 0
}

// HasWarnings returns true if any warnings have been reported
func (ser *SortedErrorReporter) HasWarnings() bool {
	return len(ser.warnings) > 0
}

// GetErrors returns all reported errors
func (ser *SortedErrorReporter) GetErrors() []domain.CompilerError {
	errors := make([]domain.CompilerError, len(ser.errors))
	copy(errors, ser.errors)
	return errors
}

// GetWarnings returns all reported warnings
func (ser *SortedErrorReporter) GetWarnings() []domain.CompilerError {
	warnings := make([]domain.CompilerError, len(ser.warnings))
	copy(warnings, ser.warnings)
	return warnings
}

// Clear clears all collected errors and warnings
func (ser *SortedErrorReporter) Clear() {
	ser.errors = ser.errors[:0]
	ser.warnings = ser.warnings[:0]
}

// Flush sorts and reports all collected errors and warnings
func (ser *SortedErrorReporter) Flush() {
	// Sort errors by location
	sort.Slice(ser.errors, func(i, j int) bool {
		return compareSourceRanges(ser.errors[i].Location, ser.errors[j].Location)
	})

	sort.Slice(ser.warnings, func(i, j int) bool {
		return compareSourceRanges(ser.warnings[i].Location, ser.warnings[j].Location)
	})

	// Report sorted errors
	for _, err := range ser.errors {
		ser.underlying.ReportError(err)
	}

	for _, warning := range ser.warnings {
		ser.underlying.ReportWarning(warning)
	}

	ser.Clear()
}

// compareSourceRanges compares two source ranges for sorting
func compareSourceRanges(a, b domain.SourceRange) bool {
	if a.Start.Filename != b.Start.Filename {
		return a.Start.Filename < b.Start.Filename
	}
	if a.Start.Line != b.Start.Line {
		return a.Start.Line < b.Start.Line
	}
	return a.Start.Column < b.Start.Column
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
