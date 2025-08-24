package tests

import (
	"testing"

	"github.com/sokoide/llvm5/staticlang/internal/domain"
	"github.com/sokoide/llvm5/staticlang/lexer"
)

func TestLexerBasicTokens(t *testing.T) {
	input := "int x = 42;"

	l := lexer.NewLexer()
	l.SetInput(input)

	// Expected tokens: INT, IDENTIFIER, ASSIGN, INTEGER_LITERAL, SEMICOLON, EOF
	expectedTokens := []domain.TokenType{
		domain.INT,
		domain.IDENTIFIER,
		domain.ASSIGN,
		domain.INTEGER_LITERAL,
		domain.SEMICOLON,
		domain.EOF,
	}

	expectedValues := []string{
		"int",
		"x",
		"=",
		"42",
		";",
		"",
	}

	for i, expected := range expectedTokens {
		token := l.NextToken()
		if token.Type != expected {
			t.Errorf("Token %d: expected type %v, got %v", i, expected, token.Type)
		}
		if token.Literal != expectedValues[i] {
			t.Errorf("Token %d: expected literal '%s', got '%s'", i, expectedValues[i], token.Literal)
		}
	}
}

func TestLexerStringLiteral(t *testing.T) {
	input := `string message = "Hello, World!";`

	l := lexer.NewLexer()
	l.SetInput(input)

	// Get tokens
	tokens := []domain.Token{}
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == domain.EOF {
			break
		}
	}

	// Check string literal
	if len(tokens) < 4 {
		t.Fatalf("Expected at least 4 tokens, got %d", len(tokens))
	}

	stringLiteral := tokens[3]
	if stringLiteral.Type != domain.STRING_LITERAL {
		t.Errorf("Expected STRING_LITERAL, got %v", stringLiteral.Type)
	}
	if stringLiteral.Literal != "\"Hello, World!\"" {
		t.Errorf("Expected string literal with quotes, got %s", stringLiteral.Literal)
	}
}

func TestLexerDoubleLiteral(t *testing.T) {
	input := "double pi = 3.14159;"

	l := lexer.NewLexer()
	l.SetInput(input)

	tokens := []domain.Token{}
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == domain.EOF {
			break
		}
	}

	// Check double literal
	if len(tokens) < 4 {
		t.Fatalf("Expected at least 4 tokens, got %d", len(tokens))
	}

	doubleLiteral := tokens[3]
	if doubleLiteral.Type != domain.DOUBLE_LITERAL {
		t.Errorf("Expected DOUBLE_LITERAL, got %v", doubleLiteral.Type)
	}
	if doubleLiteral.Literal != "3.14159" {
		t.Errorf("Expected '3.14159', got %s", doubleLiteral.Literal)
	}
}

func TestLexerKeywords(t *testing.T) {
	keywords := map[string]domain.TokenType{
		"int":      domain.INT,
		"double":   domain.DOUBLE,
		"string":   domain.STRING,
		"function": domain.FUNCTION,
		"if":       domain.IF,
		"else":     domain.ELSE,
		"while":    domain.WHILE,
		"for":      domain.FOR,
		"return":   domain.RETURN,
	}

	for keyword, expectedType := range keywords {
		l := lexer.NewLexer()
		l.SetInput(keyword)

		token := l.NextToken()
		if token.Type != expectedType {
			t.Errorf("Keyword '%s': expected type %v, got %v", keyword, expectedType, token.Type)
		}
		if token.Literal != keyword {
			t.Errorf("Keyword '%s': expected literal '%s', got '%s'", keyword, keyword, token.Literal)
		}
	}
}

func TestLexerOperators(t *testing.T) {
	input := "+ - * / == != < <= > >= = && || !"

	l := lexer.NewLexer()
	l.SetInput(input)

	expectedTokens := []domain.TokenType{
		domain.PLUS,
		domain.MINUS,
		domain.MULTIPLY,
		domain.DIVIDE,
		domain.EQ,
		domain.NE,
		domain.LT,
		domain.LE,
		domain.GT,
		domain.GE,
		domain.ASSIGN,
		domain.AND,
		domain.OR,
		domain.NOT,
		domain.EOF,
	}

	for i, expected := range expectedTokens {
		token := l.NextToken()
		if token.Type != expected {
			t.Errorf("Operator token %d: expected type %v, got %v", i, expected, token.Type)
		}
	}
}

func TestLexerDelimiters(t *testing.T) {
	input := "{ } ( ) [ ] ; , ->"

	l := lexer.NewLexer()
	l.SetInput(input)

	expectedTokens := []domain.TokenType{
		domain.LBRACE,
		domain.RBRACE,
		domain.LPAREN,
		domain.RPAREN,
		domain.LBRACKET,
		domain.RBRACKET,
		domain.SEMICOLON,
		domain.COMMA,
		domain.ARROW,
		domain.EOF,
	}

	for i, expected := range expectedTokens {
		token := l.NextToken()
		if token.Type != expected {
			t.Errorf("Delimiter token %d: expected type %v, got %v", i, expected, token.Type)
		}
	}
}

func TestLexerComments(t *testing.T) {
	input := `// This is a comment
int x = 42; // End of line comment
/* Multi-line
   comment */
int y = 24;`

	l := lexer.NewLexer()
	l.SetInput(input)

	// Comments should be skipped, so we should only get tokens for the actual code
	expectedTokens := []domain.TokenType{
		domain.INT,             // int
		domain.IDENTIFIER,      // x
		domain.ASSIGN,          // =
		domain.INTEGER_LITERAL, // 42
		domain.SEMICOLON,       // ;
		domain.INT,             // int
		domain.IDENTIFIER,      // y
		domain.ASSIGN,          // =
		domain.INTEGER_LITERAL, // 24
		domain.SEMICOLON,       // ;
		domain.EOF,
	}

	for i, expected := range expectedTokens {
		token := l.NextToken()
		if token.Type != expected {
			t.Errorf("Token %d: expected type %v, got %v", i, expected, token.Type)
		}
	}
}

func TestLexerPositionTracking(t *testing.T) {
	input := `int x = 42;
string name = "test";`

	l := lexer.NewLexer()
	l.SetInput(input)

	// First token should be at line 1, column 1
	token := l.NextToken()
	pos := l.GetCurrentPosition()
	if pos.Line != 1 || pos.Column != 1 {
		t.Errorf("First token position: expected (1,1), got (%d,%d)", pos.Line, pos.Column)
	}

	// Skip to the token on the second line
	for token.Type != domain.STRING {
		token = l.NextToken()
	}

	pos = l.GetCurrentPosition()
	if pos.Line != 2 {
		t.Errorf("Second line token: expected line 2, got line %d", pos.Line)
	}
}

func TestLexerErrorHandling(t *testing.T) {
	// Test with invalid characters
	input := "int x = @#$;"

	l := lexer.NewLexer()
	l.SetInput(input)

	// Should get tokens up to the invalid character
	token1 := l.NextToken() // int
	if token1.Type != domain.INT {
		t.Errorf("Expected INT token, got %v", token1.Type)
	}

	token2 := l.NextToken() // x
	if token2.Type != domain.IDENTIFIER {
		t.Errorf("Expected IDENTIFIER token, got %v", token2.Type)
	}

	token3 := l.NextToken() // =
	if token3.Type != domain.ASSIGN {
		t.Errorf("Expected ASSIGN token, got %v", token3.Type)
	}

	// Next token should be ILLEGAL or EOF
	token4 := l.NextToken()
	if token4.Type != domain.ILLEGAL && token4.Type != domain.EOF {
		t.Errorf("Expected ILLEGAL or EOF token for invalid character, got %v", token4.Type)
	}
}

func TestLexerEmptyInput(t *testing.T) {
	l := lexer.NewLexer()
	l.SetInput("")

	token := l.NextToken()
	if token.Type != domain.EOF {
		t.Errorf("Expected EOF for empty input, got %v", token.Type)
	}
}

func TestLexerPeekFunctionality(t *testing.T) {
	input := "int x"

	l := lexer.NewLexer()
	l.SetInput(input)

	// Peek should return the next token without advancing
	peeked := l.Peek()
	if peeked.Type != domain.INT {
		t.Errorf("Peek: expected INT, got %v", peeked.Type)
	}

	// NextToken should return the same token
	next := l.NextToken()
	if next.Type != domain.INT {
		t.Errorf("NextToken after Peek: expected INT, got %v", next.Type)
	}

	// Second peek should return IDENTIFIER
	peeked2 := l.Peek()
	if peeked2.Type != domain.IDENTIFIER {
		t.Errorf("Second peek: expected IDENTIFIER, got %v", peeked2.Type)
	}
}
