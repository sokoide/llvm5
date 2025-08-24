// Package lexer provides lexical analysis for the StaticLang compiler
package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// StaticLangLexer implements the Lexer interface for StaticLang
type StaticLangLexer struct {
	input       *bufio.Reader
	filename    string
	line        int
	column      int
	offset      int
	current     rune
	next        rune
	hasNext     bool
	peekedToken *interfaces.Token
}

// Keywords maps keyword strings to their token types
var keywords = map[string]interfaces.TokenType{
	"func":     interfaces.TokenFunc,
	"function": interfaces.TokenFunc, // Alternative syntax
	"struct":   interfaces.TokenStruct,
	"var":      interfaces.TokenVar,
	"if":       interfaces.TokenIf,
	"else":     interfaces.TokenElse,
	"while":    interfaces.TokenWhile,
	"for":      interfaces.TokenFor,
	"return":   interfaces.TokenReturn,
	"true":     interfaces.TokenTrue,
	"false":    interfaces.TokenFalse,
	// Type names like "int", "double", "string", "bool" should be identifiers
	// resolved by the type system, not special tokens
	"print": interfaces.TokenIdentifier, // Built-in function
}

// NewLexer creates a new StaticLang lexer
func NewLexer() *StaticLangLexer {
	return &StaticLangLexer{
		line:   1,
		column: 0,
		offset: 0,
	}
}

// SetInput sets the input source for lexing
func (l *StaticLangLexer) SetInput(filename string, input io.Reader) error {
	l.filename = filename
	l.input = bufio.NewReader(input)
	l.line = 1
	l.column = 0
	l.offset = 0
	l.peekedToken = nil

	// Read the first two characters
	if err := l.readChar(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("failed to read input: %w", err)
		}
		l.current = 0
	}
	if err := l.readChar(); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read input: %w", err)
	}

	return nil
}

// NextToken returns the next token from the input
func (l *StaticLangLexer) NextToken() interfaces.Token {
	if l.peekedToken != nil {
		token := *l.peekedToken
		l.peekedToken = nil
		return token
	}

	l.skipWhitespace()

	position := l.getCurrentPosition()

	if l.current == 0 {
		return interfaces.Token{
			Type:     interfaces.TokenEOF,
			Value:    "",
			Location: position,
		}
	}

	// Single-character tokens
	switch l.current {
	case '+':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenPlus, Value: "+", Location: position}
	case '-':
		if l.next == '>' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenArrow, Value: "->", Location: position}
		}
		l.advance()
		return interfaces.Token{Type: interfaces.TokenMinus, Value: "-", Location: position}
	case '*':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenStar, Value: "*", Location: position}
	case '/':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenSlash, Value: "/", Location: position}
	case '%':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenPercent, Value: "%", Location: position}
	case '(':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenLeftParen, Value: "(", Location: position}
	case ')':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenRightParen, Value: ")", Location: position}
	case '{':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenLeftBrace, Value: "{", Location: position}
	case '}':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenRightBrace, Value: "}", Location: position}
	case '[':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenLeftBracket, Value: "[", Location: position}
	case ']':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenRightBracket, Value: "]", Location: position}
	case ';':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenSemicolon, Value: ";", Location: position}
	case ',':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenComma, Value: ",", Location: position}
	case '.':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenDot, Value: ".", Location: position}
	case ':':
		l.advance()
		return interfaces.Token{Type: interfaces.TokenColon, Value: ":", Location: position}
	}

	// Two-character tokens
	switch l.current {
	case '=':
		if l.next == '=' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenEqual, Value: "==", Location: position}
		}
		l.advance()
		return interfaces.Token{Type: interfaces.TokenAssign, Value: "=", Location: position}
	case '!':
		if l.next == '=' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenNotEqual, Value: "!=", Location: position}
		}
		l.advance()
		return interfaces.Token{Type: interfaces.TokenNot, Value: "!", Location: position}
	case '<':
		if l.next == '=' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenLessEqual, Value: "<=", Location: position}
		}
		l.advance()
		return interfaces.Token{Type: interfaces.TokenLess, Value: "<", Location: position}
	case '>':
		if l.next == '=' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenGreaterEqual, Value: ">=", Location: position}
		}
		l.advance()
		return interfaces.Token{Type: interfaces.TokenGreater, Value: ">", Location: position}
	case '&':
		if l.next == '&' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenAnd, Value: "&&", Location: position}
		}
	case '|':
		if l.next == '|' {
			l.advance()
			l.advance()
			return interfaces.Token{Type: interfaces.TokenOr, Value: "||", Location: position}
		}
	}

	// String literals
	if l.current == '"' {
		return l.readString(position)
	}

	// Numeric literals
	if unicode.IsDigit(l.current) {
		return l.readNumber(position)
	}

	// Identifiers and keywords
	if unicode.IsLetter(l.current) || l.current == '_' {
		return l.readIdentifier(position)
	}

	// Unknown character
	char := string(l.current)
	l.advance()
	return interfaces.Token{
		Type:     interfaces.TokenError,
		Value:    fmt.Sprintf("unexpected character: %s", char),
		Location: position,
	}
}

// Peek returns the next token without consuming it
func (l *StaticLangLexer) Peek() interfaces.Token {
	if l.peekedToken == nil {
		token := l.NextToken()
		l.peekedToken = &token
	}
	return *l.peekedToken
}

// GetCurrentPosition returns the current position in the input
func (l *StaticLangLexer) GetCurrentPosition() domain.SourcePosition {
	return l.getCurrentPosition()
}

// Helper methods

func (l *StaticLangLexer) getCurrentPosition() domain.SourcePosition {
	return domain.SourcePosition{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.column,
		Offset:   l.offset,
	}
}

func (l *StaticLangLexer) readChar() error {
	char, _, err := l.input.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.current = l.next
			l.next = 0
			l.hasNext = false
			return err
		}
		return err
	}

	l.current = l.next
	l.next = char
	l.hasNext = true

	if l.current == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.offset++

	return nil
}

func (l *StaticLangLexer) advance() {
	l.readChar()
}

func (l *StaticLangLexer) skipWhitespace() {
	for unicode.IsSpace(l.current) {
		l.advance()
	}
}

func (l *StaticLangLexer) readString(position domain.SourcePosition) interfaces.Token {
	var value strings.Builder
	l.advance() // Skip opening quote

	for l.current != '"' && l.current != 0 {
		if l.current == '\\' {
			l.advance()
			switch l.current {
			case 'n':
				value.WriteRune('\n')
			case 't':
				value.WriteRune('\t')
			case 'r':
				value.WriteRune('\r')
			case '\\':
				value.WriteRune('\\')
			case '"':
				value.WriteRune('"')
			default:
				value.WriteRune(l.current)
			}
		} else {
			value.WriteRune(l.current)
		}
		l.advance()
	}

	if l.current == 0 {
		return interfaces.Token{
			Type:     interfaces.TokenError,
			Value:    "unterminated string literal",
			Location: position,
		}
	}

	l.advance() // Skip closing quote

	return interfaces.Token{
		Type:     interfaces.TokenString,
		Value:    value.String(),
		Location: position,
	}
}

func (l *StaticLangLexer) readNumber(position domain.SourcePosition) interfaces.Token {
	var value strings.Builder
	tokenType := interfaces.TokenInt

	// Read integer part
	for unicode.IsDigit(l.current) {
		value.WriteRune(l.current)
		l.advance()
	}

	// Check for decimal point
	if l.current == '.' && unicode.IsDigit(l.next) {
		tokenType = interfaces.TokenFloat
		value.WriteRune(l.current)
		l.advance()

		// Read fractional part
		for unicode.IsDigit(l.current) {
			value.WriteRune(l.current)
			l.advance()
		}
	}

	str := value.String()

	// Validate the number
	if tokenType == interfaces.TokenInt {
		if _, err := strconv.ParseInt(str, 10, 64); err != nil {
			return interfaces.Token{
				Type:     interfaces.TokenError,
				Value:    fmt.Sprintf("invalid integer: %s", str),
				Location: position,
			}
		}
	} else {
		if _, err := strconv.ParseFloat(str, 64); err != nil {
			return interfaces.Token{
				Type:     interfaces.TokenError,
				Value:    fmt.Sprintf("invalid float: %s", str),
				Location: position,
			}
		}
	}

	return interfaces.Token{
		Type:     tokenType,
		Value:    str,
		Location: position,
	}
}

func (l *StaticLangLexer) readIdentifier(position domain.SourcePosition) interfaces.Token {
	var value strings.Builder

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		value.WriteRune(l.current)
		l.advance()
	}

	str := value.String()
	tokenType := interfaces.TokenIdentifier

	// Check if it's a keyword
	if keywordType, isKeyword := keywords[str]; isKeyword {
		tokenType = keywordType
	}

	return interfaces.Token{
		Type:     tokenType,
		Value:    str,
		Location: position,
	}
}

// TokenTypeString returns a string representation of the token type
func TokenTypeString(t interfaces.TokenType) string {
	switch t {
	case interfaces.TokenInt:
		return "INT"
	case interfaces.TokenFloat:
		return "FLOAT"
	case interfaces.TokenString:
		return "STRING"
	case interfaces.TokenBool:
		return "BOOL"
	case interfaces.TokenIdentifier:
		return "IDENTIFIER"
	case interfaces.TokenFunc:
		return "FUNC"
	case interfaces.TokenStruct:
		return "STRUCT"
	case interfaces.TokenVar:
		return "VAR"
	case interfaces.TokenIf:
		return "IF"
	case interfaces.TokenElse:
		return "ELSE"
	case interfaces.TokenWhile:
		return "WHILE"
	case interfaces.TokenFor:
		return "FOR"
	case interfaces.TokenReturn:
		return "RETURN"
	case interfaces.TokenTrue:
		return "TRUE"
	case interfaces.TokenFalse:
		return "FALSE"
	case interfaces.TokenPlus:
		return "PLUS"
	case interfaces.TokenMinus:
		return "MINUS"
	case interfaces.TokenStar:
		return "STAR"
	case interfaces.TokenSlash:
		return "SLASH"
	case interfaces.TokenPercent:
		return "PERCENT"
	case interfaces.TokenEqual:
		return "EQUAL"
	case interfaces.TokenNotEqual:
		return "NOT_EQUAL"
	case interfaces.TokenLess:
		return "LESS"
	case interfaces.TokenLessEqual:
		return "LESS_EQUAL"
	case interfaces.TokenGreater:
		return "GREATER"
	case interfaces.TokenGreaterEqual:
		return "GREATER_EQUAL"
	case interfaces.TokenAnd:
		return "AND"
	case interfaces.TokenOr:
		return "OR"
	case interfaces.TokenNot:
		return "NOT"
	case interfaces.TokenAssign:
		return "ASSIGN"
	case interfaces.TokenLeftParen:
		return "LEFT_PAREN"
	case interfaces.TokenRightParen:
		return "RIGHT_PAREN"
	case interfaces.TokenLeftBrace:
		return "LEFT_BRACE"
	case interfaces.TokenRightBrace:
		return "RIGHT_BRACE"
	case interfaces.TokenLeftBracket:
		return "LEFT_BRACKET"
	case interfaces.TokenRightBracket:
		return "RIGHT_BRACKET"
	case interfaces.TokenSemicolon:
		return "SEMICOLON"
	case interfaces.TokenComma:
		return "COMMA"
	case interfaces.TokenDot:
		return "DOT"
	case interfaces.TokenColon:
		return "COLON"
	case interfaces.TokenArrow:
		return "ARROW"
	case interfaces.TokenEOF:
		return "EOF"
	case interfaces.TokenError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
