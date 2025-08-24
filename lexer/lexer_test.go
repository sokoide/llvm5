package lexer

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/interfaces"
)

// TestLexer_BasicTokenization tests basic token recognition
func TestLexer_BasicTokenization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interfaces.TokenType
	}{
		{
			name:  "keywords",
			input: "func var if else while for return struct true false",
			expected: []interfaces.TokenType{
				interfaces.TokenFunc, interfaces.TokenVar, interfaces.TokenIf, interfaces.TokenElse,
				interfaces.TokenWhile, interfaces.TokenFor, interfaces.TokenReturn, interfaces.TokenStruct,
				interfaces.TokenTrue, interfaces.TokenFalse, interfaces.TokenEOF,
			},
		},
		{
			name:  "operators",
			input: "+ - * / % == != < <= > >= && || ! =",
			expected: []interfaces.TokenType{
				interfaces.TokenPlus, interfaces.TokenMinus, interfaces.TokenStar, interfaces.TokenSlash,
				interfaces.TokenPercent, interfaces.TokenEqual, interfaces.TokenNotEqual, interfaces.TokenLess,
				interfaces.TokenLessEqual, interfaces.TokenGreater, interfaces.TokenGreaterEqual,
				interfaces.TokenAnd, interfaces.TokenOr, interfaces.TokenNot, interfaces.TokenAssign,
				interfaces.TokenEOF,
			},
		},
		{
			name:  "delimiters",
			input: "( ) { } [ ] ; , . : ->",
			expected: []interfaces.TokenType{
				interfaces.TokenLeftParen, interfaces.TokenRightParen, interfaces.TokenLeftBrace,
				interfaces.TokenRightBrace, interfaces.TokenLeftBracket, interfaces.TokenRightBracket,
				interfaces.TokenSemicolon, interfaces.TokenComma, interfaces.TokenDot, interfaces.TokenColon,
				interfaces.TokenArrow, interfaces.TokenEOF,
			},
		},
		{
			name:  "literals",
			input: `42 3.14 "hello" identifier`,
			expected: []interfaces.TokenType{
				interfaces.TokenInt, interfaces.TokenFloat, interfaces.TokenString,
				interfaces.TokenIdentifier, interfaces.TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			var tokens []interfaces.TokenType
			for {
				token := lexer.NextToken()
				tokens = append(tokens, token.Type)
				if token.Type == interfaces.TokenEOF {
					break
				}
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Token count mismatch. Got %d, expected %d", len(tokens), len(tt.expected))
				t.Errorf("Got tokens: %v", tokens)
				t.Errorf("Expected:   %v", tt.expected)
				return
			}

			for i, expected := range tt.expected {
				if tokens[i] != expected {
					t.Errorf("Token %d: got %v, expected %v", i, tokens[i], expected)
				}
			}
		})
	}
}

// TestLexer_TokenValues tests that token values are correctly extracted
func TestLexer_TokenValues(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  interfaces.TokenType
		expectedValue string
	}{
		{"integer", "42", interfaces.TokenInt, "42"},
		{"float", "3.14", interfaces.TokenFloat, "3.14"},
		{"string", `"hello world"`, interfaces.TokenString, "hello world"},
		{"identifier", "myVariable", interfaces.TokenIdentifier, "myVariable"},
		{"keyword_func", "func", interfaces.TokenFunc, "func"},
		{"keyword_var", "var", interfaces.TokenVar, "var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			token := lexer.NextToken()
			if token.Type != tt.expectedType {
				t.Errorf("Token type: got %v, expected %v", token.Type, tt.expectedType)
			}
			if token.Value != tt.expectedValue {
				t.Errorf("Token value: got %q, expected %q", token.Value, tt.expectedValue)
			}
		})
	}
}

// TestLexer_PositionTracking tests that source positions are correctly tracked
func TestLexer_PositionTracking(t *testing.T) {
	input := `func main() {
    var x int = 42;
    return x;
}`

	lexer := NewLexer()
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}

	// Test specific position tracking
	tests := []struct {
		expectedType   interfaces.TokenType
		expectedLine   int
		expectedColumn int
	}{
		{interfaces.TokenFunc, 1, 2},
		{interfaces.TokenIdentifier, 1, 7},  // "main"
		{interfaces.TokenLeftParen, 1, 11},  // "("
		{interfaces.TokenRightParen, 1, 12}, // ")"
		{interfaces.TokenLeftBrace, 1, 14},  // "{"
		{interfaces.TokenVar, 2, 5},         // "var"
		{interfaces.TokenIdentifier, 2, 9},  // "x"
		{interfaces.TokenIdentifier, 2, 11}, // "int"
		{interfaces.TokenAssign, 2, 15},     // "="
		{interfaces.TokenInt, 2, 17},        // "42"
		{interfaces.TokenSemicolon, 2, 19},  // ";"
	}

	for i, expected := range tests {
		token := lexer.NextToken()
		if token.Type != expected.expectedType {
			t.Errorf("Token %d type: got %v, expected %v", i, token.Type, expected.expectedType)
		}
		if token.Location.Line != expected.expectedLine {
			t.Errorf("Token %d line: got %d, expected %d", i, token.Location.Line, expected.expectedLine)
		}
		if token.Location.Column != expected.expectedColumn {
			t.Errorf("Token %d column: got %d, expected %d", i, token.Location.Column, expected.expectedColumn)
		}
	}
}

// TestLexer_ErrorHandling tests lexer behavior with invalid input
func TestLexer_ErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unterminated_string", `"unterminated string`},
		{"invalid_number", "123.456.789"},
		{"unexpected_character", "@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			// For error handling, we expect the lexer to produce some token
			// but may not be the expected type. The main thing is it shouldn't crash.
			token := lexer.NextToken()
			if token.Type == interfaces.TokenEOF && len(tt.input) > 0 {
				t.Errorf("Unexpected EOF for non-empty input")
			}
		})
	}
}

// TestLexer_CommentHandling tests that comments are properly skipped
func TestLexer_CommentHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interfaces.TokenType
	}{
		{
			name:     "single_line_comment",
			input:    "func // this is a comment\nmain",
			expected: []interfaces.TokenType{interfaces.TokenFunc, interfaces.TokenIdentifier, interfaces.TokenEOF},
		},
		{
			name:     "comment_at_end",
			input:    "var x // comment",
			expected: []interfaces.TokenType{interfaces.TokenVar, interfaces.TokenIdentifier, interfaces.TokenEOF},
		},
		{
			name:     "multiple_comments",
			input:    "// first comment\nfunc // second comment\nmain // third comment",
			expected: []interfaces.TokenType{interfaces.TokenFunc, interfaces.TokenIdentifier, interfaces.TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			var tokens []interfaces.TokenType
			for {
				token := lexer.NextToken()
				tokens = append(tokens, token.Type)
				if token.Type == interfaces.TokenEOF {
					break
				}
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Token count mismatch. Got %d, expected %d", len(tokens), len(tt.expected))
				t.Errorf("Got tokens: %v", tokens)
				t.Errorf("Expected:   %v", tt.expected)
				return
			}

			for i, expected := range tt.expected {
				if tokens[i] != expected {
					t.Errorf("Token %d: got %v, expected %v", i, tokens[i], expected)
				}
			}
		})
	}
}

// TestLexer_PeekFunctionality tests the Peek method
func TestLexer_PeekFunctionality(t *testing.T) {
	input := "func main"
	lexer := NewLexer()
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}

	// Peek should not advance the lexer
	peeked := lexer.Peek()
	if peeked.Type != interfaces.TokenFunc {
		t.Errorf("Peek: got %v, expected %v", peeked.Type, interfaces.TokenFunc)
	}

	// Next should return the same token as peek
	next := lexer.NextToken()
	if next.Type != interfaces.TokenFunc {
		t.Errorf("NextToken after Peek: got %v, expected %v", next.Type, interfaces.TokenFunc)
	}

	// Peek the next token
	peeked2 := lexer.Peek()
	if peeked2.Type != interfaces.TokenIdentifier {
		t.Errorf("Second Peek: got %v, expected %v", peeked2.Type, interfaces.TokenIdentifier)
	}
}

// TestLexer_WhitespaceHandling tests whitespace is properly handled
func TestLexer_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "func   main"},
		{"tabs", "func\t\tmain"},
		{"newlines", "func\n\nmain"},
		{"mixed", "func \t\n  main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			// Should get exactly two tokens: func and main
			token1 := lexer.NextToken()
			if token1.Type != interfaces.TokenFunc {
				t.Errorf("First token: got %v, expected %v", token1.Type, interfaces.TokenFunc)
			}

			token2 := lexer.NextToken()
			if token2.Type != interfaces.TokenIdentifier {
				t.Errorf("Second token: got %v, expected %v", token2.Type, interfaces.TokenIdentifier)
			}
			if token2.Value != "main" {
				t.Errorf("Second token value: got %q, expected %q", token2.Value, "main")
			}

			token3 := lexer.NextToken()
			if token3.Type != interfaces.TokenEOF {
				t.Errorf("Third token: got %v, expected %v", token3.Type, interfaces.TokenEOF)
			}
		})
	}
}

// TestLexer_NumberFormats tests different number formats
func TestLexer_NumberFormats(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  interfaces.TokenType
		expectedValue string
	}{
		{"integer_zero", "0", interfaces.TokenInt, "0"},
		{"integer_positive", "42", interfaces.TokenInt, "42"},
		{"integer_large", "123456789", interfaces.TokenInt, "123456789"},
		{"float_basic", "3.14", interfaces.TokenFloat, "3.14"},
		{"float_zero", "0.0", interfaces.TokenFloat, "0.0"},
		{"float_leading_zero", "0.5", interfaces.TokenFloat, "0.5"},
		{"float_trailing_zero", "5.0", interfaces.TokenFloat, "5.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			token := lexer.NextToken()
			if token.Type != tt.expectedType {
				t.Errorf("Token type: got %v, expected %v", token.Type, tt.expectedType)
			}
			if token.Value != tt.expectedValue {
				t.Errorf("Token value: got %q, expected %q", token.Value, tt.expectedValue)
			}
		})
	}
}

// TestLexer_StringEscapes tests string literal escape sequences
func TestLexer_StringEscapes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
	}{
		{"simple_string", `"hello"`, "hello"},
		{"empty_string", `""`, ""},
		{"string_with_spaces", `"hello world"`, "hello world"},
		{"string_with_newline", `"hello\nworld"`, "hello\nworld"},
		{"string_with_tab", `"hello\tworld"`, "hello\tworld"},
		{"string_with_quote", `"say \"hello\""`, `say "hello"`},
		{"string_with_backslash", `"path\\to\\file"`, `path\to\file`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer()
			err := lexer.SetInput("test.sl", strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("SetInput failed: %v", err)
			}

			token := lexer.NextToken()
			if token.Type != interfaces.TokenString {
				t.Errorf("Token type: got %v, expected %v", token.Type, interfaces.TokenString)
			}
			if token.Value != tt.expectedValue {
				t.Errorf("Token value: got %q, expected %q", token.Value, tt.expectedValue)
			}
		})
	}
}

// TestLexer_ComplexProgram tests lexing a complete small program
func TestLexer_ComplexProgram(t *testing.T) {
	input := `func fibonacci(n int) -> int {
    if (n <= 1) {
        return n;
    } else {
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}`

	lexer := NewLexer()
	err := lexer.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		t.Fatalf("SetInput failed: %v", err)
	}

	tokenCount := 0
	for {
		token := lexer.NextToken()
		tokenCount++
		if token.Type == interfaces.TokenEOF {
			break
		}
		// Ensure no error tokens
		if token.Type == interfaces.TokenError {
			t.Errorf("Error token at position %d:%d with value %q",
				token.Location.Line, token.Location.Column, token.Value)
		}
	}

	// Should have a reasonable number of tokens for this program
	if tokenCount < 30 || tokenCount > 50 {
		t.Errorf("Unexpected token count: %d (expected between 30-50)", tokenCount)
	}
}
