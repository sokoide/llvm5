package tests

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/interfaces"
)

func TestLexerBasicTokens(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := "func main() { return 42; }"
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	// Expected sequence of tokens
	expectedTokens := []interfaces.TokenType{
		interfaces.TokenFunc,
		interfaces.TokenIdentifier,
		interfaces.TokenLeftParen,
		interfaces.TokenRightParen,
		interfaces.TokenLeftBrace,
		interfaces.TokenReturn,
		interfaces.TokenInt, // 42 should be recognized as int literal
		interfaces.TokenSemicolon,
		interfaces.TokenRightBrace,
		interfaces.TokenEOF,
	}

	expectedValues := []string{
		"func", "main", "(", ")", "{", "return", "42", ";", "}", "",
	}

	for i, expectedType := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expectedType {
			t.Errorf("Token %d: expected type %v, got %v", i, expectedType, token.Type)
		}
		if token.Value != expectedValues[i] {
			t.Errorf("Token %d: expected value '%s', got '%s'", i, expectedValues[i], token.Value)
		}
	}
}

func TestLexerStringLiteral(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := `print("Hello, World!");`
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	// First token should be identifier "print"
	token := lexer.NextToken()
	if token.Type != interfaces.TokenIdentifier {
		t.Errorf("Expected identifier, got %v", token.Type)
	}
	if token.Value != "print" {
		t.Errorf("Expected 'print', got '%s'", token.Value)
	}

	// Second token should be left paren
	token = lexer.NextToken()
	if token.Type != interfaces.TokenLeftParen {
		t.Errorf("Expected left paren, got %v", token.Type)
	}

	// Third token should be string literal
	token = lexer.NextToken()
	if token.Type != interfaces.TokenString {
		t.Errorf("Expected string literal, got %v", token.Type)
	}
	// String literals may or may not include quotes depending on lexer implementation
	// Check for the actual content (without quotes is also valid)
	expectedString := `Hello, World!`
	expectedStringWithQuotes := `"Hello, World!"`
	if token.Value != expectedString && token.Value != expectedStringWithQuotes {
		t.Errorf("Expected string literal '%s' or '%s', got '%s'", expectedString, expectedStringWithQuotes, token.Value)
	}
}

func TestLexerNumberLiterals(t *testing.T) {
	testCases := []struct {
		input         string
		expectedType  interfaces.TokenType
		expectedValue string
	}{
		{"42", interfaces.TokenInt, "42"},
		{"0", interfaces.TokenInt, "0"},
		{"123", interfaces.TokenInt, "123"},
		{"3.14", interfaces.TokenFloat, "3.14"},
		{"0.5", interfaces.TokenFloat, "0.5"},
		{"10.0", interfaces.TokenFloat, "10.0"},
	}

	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			lexer := factory.CreateLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Failed to set input: %v", err)
			}

			token := lexer.NextToken()
			if token.Type != tc.expectedType {
				t.Errorf("Expected token type %v, got %v", tc.expectedType, token.Type)
			}
			if token.Value != tc.expectedValue {
				t.Errorf("Expected value '%s', got '%s'", tc.expectedValue, token.Value)
			}
		})
	}
}

func TestLexerOperators(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := "+ - * / == != < <= > >= = && || !"
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	expectedTokens := []interfaces.TokenType{
		interfaces.TokenPlus,
		interfaces.TokenMinus,
		interfaces.TokenStar,
		interfaces.TokenSlash,
		interfaces.TokenEqual,
		interfaces.TokenNotEqual,
		interfaces.TokenLess,
		interfaces.TokenLessEqual,
		interfaces.TokenGreater,
		interfaces.TokenGreaterEqual,
		interfaces.TokenAssign,
		interfaces.TokenAnd,
		interfaces.TokenOr,
		interfaces.TokenNot,
	}

	for i, expectedType := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expectedType {
			t.Errorf("Operator token %d: expected type %v, got %v", i, expectedType, token.Type)
		}
	}
}

func TestLexerDelimiters(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := "{ } ( ) [ ] ; , ->"
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	expectedTokens := []interfaces.TokenType{
		interfaces.TokenLeftBrace,
		interfaces.TokenRightBrace,
		interfaces.TokenLeftParen,
		interfaces.TokenRightParen,
		interfaces.TokenLeftBracket,
		interfaces.TokenRightBracket,
		interfaces.TokenSemicolon,
		interfaces.TokenComma,
		interfaces.TokenArrow,
	}

	for i, expectedType := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expectedType {
			t.Errorf("Delimiter token %d: expected type %v, got %v", i, expectedType, token.Type)
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	keywords := map[string]interfaces.TokenType{
		"func":   interfaces.TokenFunc,
		"struct": interfaces.TokenStruct,
		"var":    interfaces.TokenVar,
		"if":     interfaces.TokenIf,
		"else":   interfaces.TokenElse,
		"while":  interfaces.TokenWhile,
		"for":    interfaces.TokenFor,
		"return": interfaces.TokenReturn,
		"true":   interfaces.TokenTrue,
		"false":  interfaces.TokenFalse,
	}

	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())

	for keyword, expectedType := range keywords {
		t.Run(keyword, func(t *testing.T) {
			lexer := factory.CreateLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(keyword))
			if err != nil {
				t.Fatalf("Failed to set input: %v", err)
			}

			token := lexer.NextToken()
			if token.Type != expectedType {
				t.Errorf("Keyword '%s': expected type %v, got %v", keyword, expectedType, token.Type)
			}
			if token.Value != keyword {
				t.Errorf("Keyword '%s': expected value '%s', got '%s'", keyword, keyword, token.Value)
			}
		})
	}
}

func TestLexerPositionTracking(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := `func main() {
    return 42;
}`
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	// First token should be at line 1
	token := lexer.NextToken() // func
	if token.Location.Line != 1 {
		t.Errorf("First token should be at line 1, got line %d", token.Location.Line)
	}

	// Skip to return token (should be on line 2)
	for token.Type != interfaces.TokenReturn && token.Type != interfaces.TokenEOF {
		token = lexer.NextToken()
	}

	if token.Type == interfaces.TokenReturn && token.Location.Line != 2 {
		t.Errorf("Return token should be at line 2, got line %d", token.Location.Line)
	}
}

func TestLexerPeekFunctionality(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	input := "func main"
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	// Peek should return the next token without advancing
	peeked := lexer.Peek()
	if peeked.Type != interfaces.TokenFunc {
		t.Errorf("Peek: expected TokenFunc, got %v", peeked.Type)
	}

	// NextToken should return the same token
	next := lexer.NextToken()
	if next.Type != interfaces.TokenFunc {
		t.Errorf("NextToken after Peek: expected TokenFunc, got %v", next.Type)
	}

	// Second peek should return identifier
	peeked2 := lexer.Peek()
	if peeked2.Type != interfaces.TokenIdentifier {
		t.Errorf("Second peek: expected TokenIdentifier, got %v", peeked2.Type)
	}
}

func TestLexerEmptyInput(t *testing.T) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	err := lexer.SetInput("test.sl", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}

	token := lexer.NextToken()
	if token.Type != interfaces.TokenEOF {
		t.Errorf("Expected EOF for empty input, got %v", token.Type)
	}
}