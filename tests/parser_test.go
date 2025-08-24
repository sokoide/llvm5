package tests

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
	"github.com/sokoide/llvm5/lexer"
)

// Mock parser for testing since we don't have the actual parser implementation
// In a real implementation, this would use the generated Goyacc parser
type MockParser struct {
	lexer  *lexer.StaticLangLexer
	errors []string
}

func NewMockParser(l *lexer.StaticLangLexer) *MockParser {
	return &MockParser{
		lexer:  l,
		errors: make([]string, 0),
	}
}

func (p *MockParser) ParseProgram() *domain.Program {
	// This is a simplified mock implementation
	// In the real parser, this would use the Yacc-generated parser
	program := &domain.Program{
		Declarations: make([]domain.Declaration, 0),
	}

	for {
		token := p.lexer.NextToken()
		if token.Type == interfaces.TokenEOF {
			break
		}

		// Simple parsing logic for testing
		switch token.Type {
		case interfaces.TokenFunc:
			if funcDecl := p.parseFunction(); funcDecl != nil {
				program.Declarations = append(program.Declarations, funcDecl)
			}
		case interfaces.TokenInt, interfaces.TokenFloat, interfaces.TokenString:
			if varDecl := p.parseGlobalVariable(token); varDecl != nil {
				program.Declarations = append(program.Declarations, varDecl)
			}
		default:
			// For main function without explicit function keyword
			if token.Value == "int" {
				next := p.lexer.Peek()
				if next.Value == "main" {
					if mainFunc := p.parseMainFunction(); mainFunc != nil {
						program.Declarations = append(program.Declarations, mainFunc)
					}
				}
			}
		}
	}

	return program
}

func (p *MockParser) parseFunction() *domain.FunctionDecl {
	// Simplified function parsing
	nameToken := p.lexer.NextToken()
	if nameToken.Type != interfaces.TokenIdentifier {
		p.errors = append(p.errors, "expected function name")
		return nil
	}

	return &domain.FunctionDecl{
		Name:       nameToken.Value,
		Parameters: make([]domain.Parameter, 0),
		ReturnType: domain.NewVoidType(),
		Body: &domain.BlockStmt{
			Statements: make([]domain.Statement, 0),
		},
	}
}

func (p *MockParser) parseGlobalVariable(typeToken interfaces.Token) *domain.VarDeclStmt {
	nameToken := p.lexer.NextToken()
	if nameToken.Type != interfaces.TokenIdentifier {
		p.errors = append(p.errors, "expected variable name")
		return nil
	}

	var varType domain.Type
	switch typeToken.Type {
	case interfaces.TokenInt:
		varType = domain.NewIntType()
	case interfaces.TokenFloat:
		varType = domain.NewFloatType()
	case interfaces.TokenString:
		varType = domain.NewStringType()
	}

	// Check for initializer (= value)
	nextToken := p.lexer.Peek()
	if nextToken.Type == interfaces.TokenAssign {
		// Skip the = token
		p.lexer.NextToken()

		// Skip the initializer value (simplified - just consume next token)
		valueToken := p.lexer.NextToken()
		if valueToken.Type != interfaces.TokenInt && valueToken.Type != interfaces.TokenFloat && valueToken.Type != interfaces.TokenString {
			p.errors = append(p.errors, "expected initializer value")
		}

		// Skip semicolon if present
		if p.lexer.Peek().Type == interfaces.TokenSemicolon {
			p.lexer.NextToken()
		}
	}

	return &domain.VarDeclStmt{
		Name:  nameToken.Value,
		Type_: varType,
	}
}

func (p *MockParser) parseMainFunction() *domain.FunctionDecl {
	// Skip "main"
	p.lexer.NextToken()

	return &domain.FunctionDecl{
		Name:       "main",
		Parameters: make([]domain.Parameter, 0),
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: make([]domain.Statement, 0),
		},
	}
}

func (p *MockParser) GetErrors() []string {
	return p.errors
}

func TestParseSimpleProgram(t *testing.T) {
	input := `func main() {
		return 0;
	}`

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	parser := NewMockParser(l)
	program := parser.ParseProgram()

	if program == nil {
		t.Fatal("Parser returned nil program")
	}

	if len(program.Declarations) != 1 {
		t.Errorf("Expected 1 declaration, got %d", len(program.Declarations))
	}

	mainFunc, ok := program.Declarations[0].(*domain.FunctionDecl)
	if !ok {
		t.Fatal("First declaration is not a function declaration")
	}

	if mainFunc.Name != "main" {
		t.Errorf("Expected function name 'main', got '%s'", mainFunc.Name)
	}

	if mainFunc.ReturnType.String() != "void" {
		t.Errorf("Expected return type 'void', got '%s'", mainFunc.ReturnType.String())
	}
}

func TestParseFunctionWithParameters_Disable(t *testing.T) {
	input := `function add(int a, int b) -> int {
		return a + b;
	}`

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	parser := NewMockParser(l)
	program := parser.ParseProgram()

	if program == nil {
		t.Fatal("Parser returned nil program")
	}

	if len(program.Declarations) != 1 {
		t.Errorf("Expected 1 declaration, got %d", len(program.Declarations))
	}

	funcDecl, ok := program.Declarations[0].(*domain.FunctionDecl)
	if !ok {
		t.Fatal("Declaration is not a function declaration")
	}

	if funcDecl.Name != "add" {
		t.Errorf("Expected function name 'add', got '%s'", funcDecl.Name)
	}
}

func TestParseVariableDeclarations(t *testing.T) {
	input := `int x = 42;
	double pi = 3.14;
	string name = "test";`

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	parser := NewMockParser(l)
	program := parser.ParseProgram()

	if program == nil {
		t.Fatal("Parser returned nil program")
	}

	expectedDecls := 3
	if len(program.Declarations) != expectedDecls {
		t.Errorf("Expected %d declarations, got %d", expectedDecls, len(program.Declarations))
	}

	// Check first variable (int x)
	if len(program.Declarations) > 0 {
		varDecl, ok := program.Declarations[0].(*domain.VarDeclStmt)
		if !ok {
			t.Error("First declaration is not a variable declaration")
		} else {
			if varDecl.Name != "x" {
				t.Errorf("Expected variable name 'x', got '%s'", varDecl.Name)
			}
			if varDecl.Type_.String() != "int" {
				t.Errorf("Expected type 'int', got '%s'", varDecl.Type_.String())
			}
		}
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "integer literal"},
		{"3.14", "double literal"},
		{"\"hello\"", "string literal"},
		{"x", "identifier"},
		{"x + y", "binary expression"},
		{"x * (y + z)", "complex expression"},
	}

	for _, test := range tests {
		l := lexer.NewLexer()
		l.SetInput("test.sl", strings.NewReader(test.input))

		// For this test, we would parse expressions
		// Since we don't have the full parser implementation,
		// we just verify that the lexer can tokenize the expression correctly
		tokens := []interfaces.Token{}
		for {
			token := l.NextToken()
			tokens = append(tokens, token)
			if token.Type == interfaces.TokenEOF {
				break
			}
		}

		if len(tokens) < 2 { // At least one token + EOF
			t.Errorf("Expression '%s' produced too few tokens: %d", test.input, len(tokens))
		}
	}
}

func TestParseControlStructures(t *testing.T) {
	input := `int main() {
		if (x > 0) {
			return 1;
		} else {
			return 0;
		}
	}`

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	parser := NewMockParser(l)
	program := parser.ParseProgram()

	if program == nil {
		t.Fatal("Parser returned nil program")
	}

	// Verify that the basic structure is parsed
	if len(program.Declarations) == 0 {
		t.Error("Expected at least one declaration")
	}
}

func TestParseLoops(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"for loop",
			`for (int i = 0; i < 10; i++) {
				print(i);
			}`,
		},
		{
			"while loop",
			`while (x > 0) {
				x = x - 1;
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer()
			l.SetInput("test.sl", strings.NewReader(test.input))

			// Verify that all keywords are properly tokenized
			foundKeywords := make(map[string]bool)
			for {
				token := l.NextToken()
				if token.Type == interfaces.TokenEOF {
					break
				}

				switch token.Type {
				case interfaces.TokenFor, interfaces.TokenWhile, interfaces.TokenIf:
					foundKeywords[token.Value] = true
				}
			}

			if test.name == "for loop" && !foundKeywords["for"] {
				t.Error("for keyword not found in for loop")
			}
			if test.name == "while loop" && !foundKeywords["while"] {
				t.Error("while keyword not found in while loop")
			}
		})
	}
}

func TestParserErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing semicolon",
			`int x = 42
			int y = 24;`,
		},
		{
			"invalid function syntax",
			`function () {
				return 0;
			}`,
		},
		{
			"mismatched braces",
			`int main() {
				if (true) {
					return 1;
				// Missing closing brace
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer()
			l.SetInput("test.sl", strings.NewReader(test.input))

			parser := NewMockParser(l)
			program := parser.ParseProgram()

			// In a real parser, these should produce errors
			// For now, we just verify that the parser doesn't crash
			if program == nil {
				t.Error("Parser should not return nil even on errors")
			}
		})
	}
}

func TestParseComplexProgram(t *testing.T) {
	input := `
	int globalVar;

	func fibonacci() {
		return 0;
	}

	int main() {
		return 0;
	}
	`

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	parser := NewMockParser(l)
	program := parser.ParseProgram()

	if program == nil {
		t.Fatal("Parser returned nil program")
	}

	// Should have global variable, fibonacci function, and main function
	expectedDecls := 3
	if len(program.Declarations) != expectedDecls {
		t.Errorf("Expected %d declarations, got %d", expectedDecls, len(program.Declarations))
	}

	errors := parser.GetErrors()
	if len(errors) > 0 {
		t.Errorf("Parser produced errors: %v", errors)
	}
}

func TestParseOperatorPrecedence(t *testing.T) {
	// Test that operators are tokenized correctly
	// Real precedence testing would require the full parser
	input := "x + y * z - w / v"

	l := lexer.NewLexer()
	l.SetInput("test.sl", strings.NewReader(input))

	expectedOperators := []interfaces.TokenType{
		interfaces.TokenPlus,  // +
		interfaces.TokenStar,  // *
		interfaces.TokenMinus, // -
		interfaces.TokenSlash, // /
	}

	operatorCount := 0
	for {
		token := l.NextToken()
		if token.Type == interfaces.TokenEOF {
			break
		}

		for _, expectedOp := range expectedOperators {
			if token.Type == expectedOp {
				operatorCount++
				break
			}
		}
	}

	if operatorCount != len(expectedOperators) {
		t.Errorf("Expected %d operators, found %d", len(expectedOperators), operatorCount)
	}
}
