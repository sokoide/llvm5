package infrastructure

import (
	"os"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

// TestConsoleErrorReporter tests console error reporter functionality
func TestConsoleErrorReporter(t *testing.T) {
	reporter := NewConsoleErrorReporter(os.Stderr)
	if reporter == nil {
		t.Fatal("NewConsoleErrorReporter should return non-nil reporter")
	}

	// Test initial state
	if reporter.HasErrors() {
		t.Error("New reporter should have no errors")
	}

	if reporter.HasWarnings() {
		t.Error("New reporter should have no warnings")
	}

	// Create test error
	testError := domain.CompilerError{
		Type:    domain.SyntaxError,
		Message: "Test syntax error",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 5},
		},
	}

	// Report error
	reporter.ReportError(testError)

	// Test state after error
	if !reporter.HasErrors() {
		t.Error("Reporter should have errors after reporting")
	}

	errors := reporter.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].Message != "Test syntax error" {
		t.Errorf("Expected 'Test syntax error', got '%s'", errors[0].Message)
	}

	// Create test warning (using LexicalError as warning type)
	testWarning := domain.CompilerError{
		Type:    domain.LexicalError,
		Message: "Test warning",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 5},
		},
	}

	// Report warning
	reporter.ReportWarning(testWarning)

	// Test state after warning
	if !reporter.HasWarnings() {
		t.Error("Reporter should have warnings after reporting")
	}

	warnings := reporter.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	// Test clear
	reporter.Clear()

	if reporter.HasErrors() {
		t.Error("Reporter should have no errors after clear")
	}

	if reporter.HasWarnings() {
		t.Error("Reporter should have no warnings after clear")
	}
}

// TestConsoleErrorReporterLimits tests error and warning limits
func TestConsoleErrorReporterLimits(t *testing.T) {
	reporter := NewConsoleErrorReporter(os.Stderr)

	// Set limits
	reporter.SetMaxErrors(2)
	reporter.SetMaxWarnings(1)

	// Create multiple errors
	for i := 0; i < 5; i++ {
		testError := domain.CompilerError{
			Type:    domain.SyntaxError,
			Message: "Test error",
			Location: domain.SourceRange{
				Start: domain.SourcePosition{Filename: "test.sl", Line: i + 1, Column: 1},
				End:   domain.SourcePosition{Filename: "test.sl", Line: i + 1, Column: 5},
			},
		}
		reporter.ReportError(testError)
	}

	// Should only have 2 errors due to limit
	errors := reporter.GetErrors()
	if len(errors) > 2 {
		t.Errorf("Expected at most 2 errors due to limit, got %d", len(errors))
	}

	// Create multiple warnings
	for i := 0; i < 3; i++ {
		testWarning := domain.CompilerError{
			Type:    domain.LexicalError,
			Message: "Test warning",
			Location: domain.SourceRange{
				Start: domain.SourcePosition{Filename: "test.sl", Line: i + 1, Column: 1},
				End:   domain.SourcePosition{Filename: "test.sl", Line: i + 1, Column: 5},
			},
		}
		reporter.ReportWarning(testWarning)
	}

	// Should only have 1 warning due to limit
	warnings := reporter.GetWarnings()
	if len(warnings) > 1 {
		t.Errorf("Expected at most 1 warning due to limit, got %d", len(warnings))
	}
}

// TestSortedErrorReporter tests sorted error reporter functionality
func TestSortedErrorReporter(t *testing.T) {
	baseReporter := NewConsoleErrorReporter(os.Stderr)
	reporter := NewSortedErrorReporter(baseReporter)
	if reporter == nil {
		t.Fatal("NewSortedErrorReporter should return non-nil reporter")
	}

	// Test initial state
	if reporter.HasErrors() {
		t.Error("New reporter should have no errors")
	}

	// Create test errors in different order
	error1 := domain.CompilerError{
		Type:    domain.SyntaxError,
		Message: "First error",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 3, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 3, Column: 5},
		},
	}

	error2 := domain.CompilerError{
		Type:    domain.TypeCheckError,
		Message: "Second error",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 5},
		},
	}

	error3 := domain.CompilerError{
		Type:    domain.SyntaxError,
		Message: "Third error",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 5},
		},
	}

	// Report errors in non-sorted order
	reporter.ReportError(error1)
	reporter.ReportError(error2)
	reporter.ReportError(error3)

	// Test that errors are present
	if !reporter.HasErrors() {
		t.Error("Reporter should have errors after reporting")
	}

	// Get errors before flush to check sorting
	errors := reporter.GetErrors()
	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errors))
	}

	// Test sorting - errors should be sorted by line number after flush
	// Test flush functionality (this sorts the errors and reports them to underlying)
	reporter.Flush()

	// After flush, check the underlying reporter has the errors
	underlyingErrors := baseReporter.GetErrors()
	if len(underlyingErrors) >= 3 {
		if underlyingErrors[0].Location.Start.Line > underlyingErrors[1].Location.Start.Line {
			t.Error("Errors should be sorted by line number")
		}

		if underlyingErrors[1].Location.Start.Line > underlyingErrors[2].Location.Start.Line {
			t.Error("Errors should be sorted by line number")
		}
	}
}

// TestErrorReporterSourceContext tests source context handling
func TestErrorReporterSourceContext(t *testing.T) {
	reporter := NewConsoleErrorReporter(os.Stderr)

	// Set source content with multiple lines
	sourceContent := `func main() {
    var x int = "invalid";
    return x;
}`
	sourceBytes := []byte(sourceContent)
	reporter.SetSourceContent("test.sl", sourceBytes)

	// Create error pointing to specific location
	testError := domain.CompilerError{
		Type:    domain.TypeCheckError,
		Message: "Cannot assign string to int",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 17},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 26},
		},
	}

	reporter.ReportError(testError)

	// Test that error was recorded
	if !reporter.HasErrors() {
		t.Error("Reporter should have errors after reporting")
	}

	errors := reporter.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if !strings.Contains(errors[0].Message, "Cannot assign string to int") {
		t.Error("Error message should be preserved")
	}
}

// TestConsoleErrorReporterReportWarning tests the specific ReportWarning method coverage
func TestConsoleErrorReporterReportWarning(t *testing.T) {
	reporter := NewConsoleErrorReporter(os.Stderr)

	if reporter == nil {
		t.Fatal("NewConsoleErrorReporter should return non-nil reporter")
	}

	// Test initial state
	if reporter.HasWarnings() {
		t.Error("New reporter should have no warnings")
	}

	// Create multiple test warnings to ensure ReportWarning is covered
	testWarning1 := domain.CompilerError{
		Type:    domain.LexicalError,
		Message: "Test warning 1",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 10},
		},
	}

	testWarning2 := domain.CompilerError{
		Type:    domain.TypeCheckError,
		Message: "Test warning 2",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 10},
		},
	}

	// Report first warning
	reporter.ReportWarning(testWarning1)

	if !reporter.HasWarnings() {
		t.Error("Reporter should have warnings after reporting")
	}

	// Report second warning
	reporter.ReportWarning(testWarning2)

	// Verify both warnings are stored
	warnings := reporter.GetWarnings()
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	if warnings[0].Message != "Test warning 1" {
		t.Errorf("First warning message incorrect: got '%s'", warnings[0].Message)
	}

	if warnings[1].Message != "Test warning 2" {
		t.Errorf("Second warning message incorrect: got '%s'", warnings[1].Message)
	}
}

// TestSortedErrorReporterReportWarning tests the sorted reporter's ReportWarning method
func TestSortedErrorReporterReportWarning(t *testing.T) {
	baseReporter := NewConsoleErrorReporter(os.Stderr)
	reporter := NewSortedErrorReporter(baseReporter)

	if reporter == nil {
		t.Fatal("NewSortedErrorReporter should return non-nil reporter")
	}

	// Test initial state for warnings
	if reporter.HasWarnings() {
		t.Error("New reporter should have no warnings")
	}

	// Create test warnings
	testWarning1 := domain.CompilerError{
		Type:    domain.LexicalError,
		Message: "Warning A",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 3, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 3, Column: 10},
		},
	}

	testWarning2 := domain.CompilerError{
		Type:    domain.TypeCheckError,
		Message: "Warning B",
		Location: domain.SourceRange{
			Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 1},
			End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 10},
		},
	}

	// Report warnings in non-sorted order
	reporter.ReportWarning(testWarning1)
	reporter.ReportWarning(testWarning2)

	// Test that warnings are present
	if !reporter.HasWarnings() {
		t.Error("Reporter should have warnings after reporting")
	}

	// Get warnings before flush
	warnings := reporter.GetWarnings()
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Test that warnings can be retrieved
	if warnings[0].Message != "Warning A" {
		t.Errorf("First warning message incorrect: got '%s'", warnings[0].Message)
	}

	if warnings[1].Message != "Warning B" {
		t.Errorf("Second warning message incorrect: got '%s'", warnings[1].Message)
	}

	// Test flush functionality
	reporter.Flush()

	// After flush, check that warnings were transferred to underlying reporter
	underlyingWarnings := baseReporter.GetWarnings()
	if len(underlyingWarnings) != 2 {
		t.Errorf("Expected underlying reporter to have 2 warnings after flush, got %d", len(underlyingWarnings))
	}
}

// TestCompareSourceRanges tests source range comparison for sorting
func TestCompareSourceRanges(t *testing.T) {
	range1 := domain.SourceRange{
		Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 5},
		End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 10},
	}

	range2 := domain.SourceRange{
		Start: domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 1},
		End:   domain.SourcePosition{Filename: "test.sl", Line: 2, Column: 5},
	}

	range3 := domain.SourceRange{
		Start: domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 1},
		End:   domain.SourcePosition{Filename: "test.sl", Line: 1, Column: 5},
	}

	// Test comparison logic (compareSourceRanges returns bool: true if a < b)
	result12 := compareSourceRanges(range1, range2)
	if !result12 {
		t.Error("range1 should be less than range2 (earlier line)")
	}

	result13 := compareSourceRanges(range1, range3)
	if result13 {
		t.Error("range1 should be greater than range3 (same line, later column)")
	}

	result11 := compareSourceRanges(range1, range1)
	if result11 {
		t.Error("Same ranges should be equal")
	}
}

// TestUtilityFunctions tests min/max utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test max function
	maxResult := max(5, 10)
	if maxResult != 10 {
		t.Errorf("max(5, 10) should be 10, got %d", maxResult)
	}

	maxResultEqual := max(7, 7)
	if maxResultEqual != 7 {
		t.Errorf("max(7, 7) should be 7, got %d", maxResultEqual)
	}

	// Test min function
	minResult := min(5, 10)
	if minResult != 5 {
		t.Errorf("min(5, 10) should be 5, got %d", minResult)
	}

	minResultEqual := min(3, 3)
	if minResultEqual != 3 {
		t.Errorf("min(3, 3) should be 3, got %d", minResultEqual)
	}
}
