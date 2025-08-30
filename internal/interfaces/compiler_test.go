package interfaces

import (
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

func TestTokenString(t *testing.T) {
	// Test all token types for String() method
	testCases := []struct {
		token    TokenType
		expected string
	}{
		{TokenEOF, "EOF"},
		{TokenError, "Error"},
		{TokenIdentifier, "Identifier"},
		{TokenInt, "Int"},
		{TokenFloat, "Float"},
		{TokenString, "String"},
		{TokenBool, "Bool"},
		{TokenTrue, "True"},
		{TokenFalse, "False"},
		{TokenFunc, "Func"},
		{TokenStruct, "Struct"},
		{TokenVar, "Var"},
		{TokenIf, "If"},
		{TokenElse, "Else"},
		{TokenWhile, "While"},
		{TokenFor, "For"},
		{TokenReturn, "Return"},
		{TokenPlus, "Plus"},
		{TokenMinus, "Minus"},
		{TokenStar, "Star"},
		{TokenSlash, "Slash"},
		{TokenPercent, "Percent"},
		{TokenEqual, "Equal"},
		{TokenNotEqual, "NotEqual"},
		{TokenLess, "Less"},
		{TokenLessEqual, "LessEqual"},
		{TokenGreater, "Greater"},
		{TokenGreaterEqual, "GreaterEqual"},
		{TokenAnd, "And"},
		{TokenOr, "Or"},
		{TokenNot, "Not"},
		{TokenAssign, "Assign"},
		{TokenLeftParen, "LeftParen"},
		{TokenRightParen, "RightParen"},
		{TokenLeftBrace, "LeftBrace"},
		{TokenRightBrace, "RightBrace"},
		{TokenLeftBracket, "LeftBracket"},
		{TokenRightBracket, "RightBracket"},
		{TokenSemicolon, "Semicolon"},
		{TokenComma, "Comma"},
		{TokenDot, "Dot"},
		{TokenArrow, "Arrow"},
		{TokenColon, "Colon"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.token.String()
			if result != tc.expected {
				t.Errorf("Expected token string '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestTokenLocationString(t *testing.T) {
	// Test token location formatting
	location := domain.SourcePosition{
		Filename: "test.sl",
		Line:      5,
		Column:   12,
		Offset:   47,
	}

	expected := "test.sl:5:12"
	result := location.String()
	if result != expected {
		t.Errorf("Expected location string '%s', got '%s'", expected, result)
	}
}

// TestTokenTypesValues verifies token type constants have expected values
func TestTokenTypesValues(t *testing.T) {
	// Test a few specific token type values
	if TokenInt != 0 {
		t.Errorf("TokenInt should have value 0, got %d", TokenInt)
	}

	if TokenIdentifier < TokenInt {
		t.Error("TokenIdentifier should have value greater than TokenInt")
	}

	if TokenFunc <= TokenIdentifier {
		t.Error("TokenFunc should have value greater than TokenIdentifier")
	}

	// Test that all keyword tokens are properly defined
	keywords := []TokenType{
		TokenFunc, TokenStruct, TokenVar, TokenIf, TokenElse,
		TokenWhile, TokenFor, TokenReturn, TokenTrue, TokenFalse,
	}

	for _, keyword := range keywords {
		if keyword <= 0 {
			t.Errorf("Keyword token %v should have positive value", keyword)
		}
	}
}

func TestSourcePosition(t *testing.T) {
	pos := domain.SourcePosition{
		Filename: "main.sl",
		Line:      10,
		Column:   25,
		Offset:   150,
	}

	if pos.Filename != "main.sl" {
		t.Errorf("Expected filename 'main.sl', got '%s'", pos.Filename)
	}

	if pos.Line != 10 {
		t.Errorf("Expected line 10, got %d", pos.Line)
	}

	if pos.Column != 25 {
		t.Errorf("Expected column 25, got %d", pos.Column)
	}

	if pos.Offset != 150 {
		t.Errorf("Expected offset 150, got %d", pos.Offset)
	}

	expectedString := "main.sl:10:25"
	if pos.String() != expectedString {
		t.Errorf("Expected string '%s', got '%s'", expectedString, pos.String())
	}
}

func TestSourceRange(t *testing.T) {
	start := domain.SourcePosition{Filename: "test.sl", Line: 5, Column: 10, Offset: 50}
	end := domain.SourcePosition{Filename: "test.sl", Line: 5, Column: 25, Offset: 65}

	srcRange := domain.SourceRange{
		Start: start,
		End:   end,
	}

	if srcRange.Start != start {
		t.Error("SourceRange Start not set correctly")
	}

	if srcRange.End != end {
		t.Error("SourceRange End not set correctly")
	}

	expectedString := "test.sl:5:10-25"
	if srcRange.String() != expectedString {
		t.Errorf("Expected range string '%s', got '%s'", expectedString, srcRange.String())
	}

	// Test multi-line range
	endMultiLine := domain.SourcePosition{Filename: "test.sl", Line: 7, Column: 5, Offset: 85}
	srcRangeMulti := domain.SourceRange{Start: start, End: endMultiLine}

	expectedMultiString := "test.sl:5:10-7:5"
	if srcRangeMulti.String() != expectedMultiString {
		t.Errorf("Expected multi-line range string '%s', got '%s'", expectedMultiString, srcRangeMulti.String())
	}
}

// TestTokenTypeConstants ensures all token type constants are properly defined
func TestTokenTypeConstants(t *testing.T) {
	// Collect all token type constants
	allTokens := []TokenType{
		TokenEOF, TokenError, TokenIdentifier, TokenInt, TokenFloat, TokenString, TokenBool,
		TokenTrue, TokenFalse, TokenFunc, TokenStruct, TokenVar, TokenIf, TokenElse,
		TokenWhile, TokenFor, TokenReturn, TokenPlus, TokenMinus, TokenStar, TokenSlash, TokenPercent,
		TokenEqual, TokenNotEqual, TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual,
		TokenAnd, TokenOr, TokenNot, TokenAssign, TokenLeftParen, TokenRightParen,
		TokenLeftBrace, TokenRightBrace, TokenLeftBracket, TokenRightBracket,
		TokenSemicolon, TokenComma, TokenDot, TokenArrow, TokenColon,
	}

	// Ensure no duplicates by checking uniqueness
	seen := make(map[TokenType]bool)
	for _, token := range allTokens {
		if seen[token] {
			t.Errorf("Duplicate token type constant: %v", token)
		}
		seen[token] = true
	}

	// Ensure all constants are defined (TokenInt is legitimately 0 as the first iota)
	for i, token := range allTokens {
		if i > 0 && i != 3 && token == 0 { // Allow TokenInt (index 3) to be 0
			t.Errorf("Token constant at index %d should not be 0", i)
		}
	}
}