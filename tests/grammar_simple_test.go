package tests

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/domain"
)

// TestGrammarBasicPrograms tests basic program parsing
func TestGrammarBasicPrograms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "simple main function",
			input:    "func main() { return 42; }",
			hasError: false,
		},
		{
			name:     "function with return type",
			input:    "func test() -> int { return 0; }",
			hasError: false,
		},
		{
			name:     "function with parameter",
			input:    "func square(x int) -> int { return x * x; }",
			hasError: false,
		},
		{
			name:     "global variable",
			input:    "int x = 42;",
			hasError: false,
		},
		{
			name:     "function with local variable",
			input:    "func test() -> int { var x int = 42; return x; }",
			hasError: false,
		},
		{
			name:     "function with if statement",
			input:    "func test() -> int { if (x > 0) { return 1; } return 0; }",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
			lexer := factory.CreateLexer()
			parser := factory.CreateParser()

			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Lexer SetInput failed: %v", err)
			}

			program, err := parser.Parse(lexer)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected parsing error: %v", err)
				} else if program == nil {
					t.Errorf("Program should not be nil")
				} else if len(program.Declarations) == 0 {
					t.Errorf("Expected at least one declaration")
				}
			}
		})
	}
}

// TestGrammarExpressionParsing tests basic expressions
func TestGrammarExpressionParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "arithmetic",
			input: "func test() -> int { return 1 + 2 * 3; }",
		},
		{
			name:  "comparison",
			input: "func test() -> bool { return x > y; }",
		},
		{
			name:  "parentheses",
			input: "func test() -> int { return (1 + 2) * 3; }",
		},
		{
			name:  "function call",
			input: "func test() -> int { return add(1, 2); }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
			lexer := factory.CreateLexer()
			parser := factory.CreateParser()

			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Lexer SetInput failed: %v", err)
			}

			program, err := parser.Parse(lexer)
			if err != nil {
				t.Errorf("Unexpected parsing error: %v", err)
			} else if program == nil {
				t.Errorf("Program should not be nil")
			}
		})
	}
}

// TestGrammarVariableDeclarations tests variable declarations
func TestGrammarVariableDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "global int variable",
			input: "int x = 42;",
		},
		{
			name:  "global string variable",
			input: "string message = \"Hello\";",
		},
		{
			name:  "local variable in function",
			input: "func test() -> int { var y int = 10; return y; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
			lexer := factory.CreateLexer()
			parser := factory.CreateParser()

			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Lexer SetInput failed: %v", err)
			}

			program, err := parser.Parse(lexer)
			if err != nil {
				t.Errorf("Unexpected parsing error: %v", err)
			} else if program == nil || len(program.Declarations) == 0 {
				t.Errorf("Expected at least one declaration")
			} else {
				// Check if it's a variable declaration
				if _, ok := program.Declarations[0].(*domain.VarDeclStmt); !ok {
					if _, ok := program.Declarations[0].(*domain.FunctionDecl); !ok {
						t.Errorf("Expected VarDeclStmt or FunctionDecl, got %T", program.Declarations[0])
					}
				}
			}
		})
	}
}

// TestGrammarCompleteProgram tests a complete program similar to existing examples
func TestGrammarCompleteProgram(t *testing.T) {
	input := `
int globalCounter = 0;

func testFunction(n int) -> int {
    var x int = n * 2;
    if (x > 10) {
        return x;
    } else {
        return 0;
    }
}

func main() -> int {
    var result int = testFunction(5);
    globalCounter = result;
    return globalCounter;
}
`

	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()
	parser := factory.CreateParser()

	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Lexer SetInput failed: %v", err)
	}

	program, err := parser.Parse(lexer)
	if err != nil {
		t.Fatalf("Parser failed: %v", err)
	}

	if program == nil {
		t.Fatal("Program should not be nil")
	}

	if len(program.Declarations) != 3 {
		t.Errorf("Expected 3 declarations, got %d", len(program.Declarations))
	}

	// Verify first declaration is a global variable
	if _, ok := program.Declarations[0].(*domain.VarDeclStmt); !ok {
		t.Errorf("First declaration should be VarDeclStmt, got %T", program.Declarations[0])
	}

	// Verify second and third declarations are functions
	for i := 1; i < len(program.Declarations); i++ {
		if _, ok := program.Declarations[i].(*domain.FunctionDecl); !ok {
			t.Errorf("Declaration %d should be FunctionDecl, got %T", i, program.Declarations[i])
		}
	}
}
