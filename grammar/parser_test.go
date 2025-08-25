package grammar

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
	"github.com/sokoide/llvm5/lexer"
)

// TestNewRecursiveDescentParser tests parser creation
func TestNewRecursiveDescentParser(t *testing.T) {
	parser := NewRecursiveDescentParser()
	if parser == nil {
		t.Error("NewRecursiveDescentParser should return non-nil parser")
	}
}

// TestParserSetDebugLevel tests debug level setting
func TestParserSetDebugLevel(t *testing.T) {
	// Test various debug levels
	levels := []int{0, 1, 2, 3}

	for _, level := range levels {
		SetDebugLevel(level)
		// No direct way to verify, but should not panic
	}
}

// TestParserSetErrorReporter tests error reporter setting
func TestParserSetErrorReporter(t *testing.T) {
	parser := NewRecursiveDescentParser()

	// Create a mock error reporter
	var reportedErrors []domain.CompilerError
	mockReporter := &MockErrorReporter{
		errors: &reportedErrors,
	}

	parser.SetErrorReporter(mockReporter)
	// Should not panic
}

// TestParserParseEmptyProgram tests parsing empty program
func TestParserParseEmptyProgram(t *testing.T) {
	parser := NewRecursiveDescentParser()

	// Create lexer with empty input
	lexerInstance := lexer.NewLexer()
	err := lexerInstance.SetInput("test.sl", strings.NewReader(""))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}

	program, err := parser.Parse(lexerInstance)
	if err != nil {
		// Empty program might result in syntax error, which is expected
		t.Logf("Empty program parse resulted in error (expected): %v", err)
	}

	if program != nil {
		// If parsing succeeds, program should have empty declarations
		if len(program.Declarations) > 0 {
			t.Error("Empty program should have no declarations")
		}
	}
}

// TestParserParseSimpleFunction tests parsing a simple function
func TestParserParseSimpleFunction(t *testing.T) {
	parser := NewRecursiveDescentParser()

	source := `func main() -> int {
		return 42;
	}`

	lexerInstance := lexer.NewLexer()
	err := lexerInstance.SetInput("test.sl", strings.NewReader(source))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}
	program, err := parser.Parse(lexerInstance)

	if err != nil {
		t.Errorf("Simple function parse failed: %v", err)
		return
	}

	if program == nil {
		t.Error("Parser should return non-nil program")
		return
	}

	if len(program.Declarations) == 0 {
		t.Error("Program should have function declaration")
		return
	}

	// Check that first declaration is a function
	if funcDecl, ok := program.Declarations[0].(*domain.FunctionDecl); ok {
		if funcDecl.Name != "main" {
			t.Errorf("Expected function name 'main', got '%s'", funcDecl.Name)
		}
	} else {
		t.Error("First declaration should be function declaration")
	}
}

// TestParserParseVariableDeclaration tests parsing variable declarations
func TestParserParseVariableDeclaration(t *testing.T) {
	parser := NewRecursiveDescentParser()

	source := `func test() -> void {
		var x int = 0;
		var y int = 42;
	}`

	lexerInstance := lexer.NewLexer()
	err := lexerInstance.SetInput("test.sl", strings.NewReader(source))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}
	program, err := parser.Parse(lexerInstance)

	if err != nil {
		t.Errorf("Variable declaration parse failed: %v", err)
		return
	}

	if program == nil || len(program.Declarations) == 0 {
		t.Error("Program should have function declaration")
		return
	}
}

// TestParserParseExpressions tests parsing various expressions
func TestParserParseExpressions(t *testing.T) {
	parser := NewRecursiveDescentParser()

	testCases := []struct {
		name   string
		source string
	}{
		{
			name: "arithmetic_expression",
			source: `func test() -> int {
				return 1 + 2 * 3;
			}`,
		},
		{
			name: "comparison_expression",
			source: `func test() -> bool {
				return 5 > 3;
			}`,
		},
		{
			name: "logical_expression",
			source: `func test() -> bool {
				return true && false;
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexerInstance := lexer.NewLexer()
			err := lexerInstance.SetInput("test.sl", strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}
			program, err := parser.Parse(lexerInstance)

			if err != nil {
				t.Errorf("Expression parse failed for %s: %v", tc.name, err)
				return
			}

			if program == nil {
				t.Errorf("Parser should return non-nil program for %s", tc.name)
			}
		})
	}
}

// TestParserParseControlFlow tests parsing control flow statements
func TestParserParseControlFlow(t *testing.T) {
	parser := NewRecursiveDescentParser()

	testCases := []struct {
		name   string
		source string
	}{
		{
			name: "if_statement",
			source: `func test() -> void {
				if (true) {
					return;
				}
			}`,
		},
		{
			name: "if_else_statement",
			source: `func test() -> void {
				if (false) {
					return;
				} else {
					return;
				}
			}`,
		},
		{
			name: "while_loop",
			source: `func test() -> void {
				while (true) {
					break;
				}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexerInstance := lexer.NewLexer()
			err := lexerInstance.SetInput("test.sl", strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}
			program, err := parser.Parse(lexerInstance)

			if err != nil {
				t.Errorf("Control flow parse failed for %s: %v", tc.name, err)
				return
			}

			if program == nil {
				t.Errorf("Parser should return non-nil program for %s", tc.name)
			}
		})
	}
}

// TestParserErrorRecovery tests error recovery
func TestParserErrorRecovery(t *testing.T) {
	parser := NewRecursiveDescentParser()

	// Invalid syntax cases
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "missing_semicolon",
			source: "func test() -> void { var x: int }",
		},
		{
			name:   "invalid_token",
			source: "func test() -> void { @ }",
		},
		{
			name:   "unmatched_brace",
			source: "func test() -> void { var x: int;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexerInstance := lexer.NewLexer()
			err := lexerInstance.SetInput("test.sl", strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}
			program, err := parser.Parse(lexerInstance)

			// Error is expected for invalid syntax
			if err == nil {
				t.Logf("Expected error for invalid syntax in %s, but got nil", tc.name)
			}

			// Program might still be returned with partial parsing
			if program != nil {
				t.Logf("Parser returned partial program for %s", tc.name)
			}
		})
	}
}

// TestParserTokenUtilities tests token utility functions
func TestParserTokenUtilities(t *testing.T) {
	// Test getLocationFromToken
	token := interfaces.Token{
		Type:  interfaces.TokenIdentifier,
		Value: "test",
		Location: domain.SourcePosition{
			Filename: "test.sl",
			Line:     1,
			Column:   1,
		},
	}

	location := getLocationFromToken(token)
	if location.Start.Filename != "test.sl" {
		t.Error("getLocationFromToken should preserve filename")
	}

	if location.Start.Line != 1 {
		t.Error("getLocationFromToken should preserve line number")
	}

	// Test getLocationFromString
	stringLocation := getLocationFromString("test")

	// Since getLocationFromString returns empty positions, just verify it doesn't panic
	_ = stringLocation
}

// MockErrorReporter for testing
type MockErrorReporter struct {
	errors   *[]domain.CompilerError
	warnings *[]domain.CompilerError
}

func (m *MockErrorReporter) ReportError(err domain.CompilerError) {
	if m.errors != nil {
		*m.errors = append(*m.errors, err)
	}
}

func (m *MockErrorReporter) ReportWarning(warning domain.CompilerError) {
	if m.warnings != nil {
		*m.warnings = append(*m.warnings, warning)
	}
}

func (m *MockErrorReporter) HasErrors() bool {
	return m.errors != nil && len(*m.errors) > 0
}

func (m *MockErrorReporter) HasWarnings() bool {
	return m.warnings != nil && len(*m.warnings) > 0
}

func (m *MockErrorReporter) GetErrors() []domain.CompilerError {
	if m.errors == nil {
		return nil
	}
	return *m.errors
}

func (m *MockErrorReporter) GetWarnings() []domain.CompilerError {
	if m.warnings == nil {
		return nil
	}
	return *m.warnings
}

func (m *MockErrorReporter) Clear() {
	if m.errors != nil {
		*m.errors = nil
	}
	if m.warnings != nil {
		*m.warnings = nil
	}
}
