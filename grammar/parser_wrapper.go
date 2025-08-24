package grammar

import (
	"fmt"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// Parser is a wrapper used by the generated Yacc parser. It adapts the
// project's Lexer (interfaces.Lexer) to the lexer interface expected by the
// generated parser and stores the parse result and a type registry.
type Parser struct {
	lexer        interfaces.Lexer
	result       *domain.Program
	typeRegistry domain.TypeRegistry
	errors       []string
}

// SetDebugLevel sets the parser debug level (0-4)
func SetDebugLevel(level int) {
	if level >= 0 && level <= 4 {
		yyDebug = level
	}
}

// NewRecursiveDescentParser returns a new Parser that implements
// interfaces.Parser. The name matches the factory usage in the codebase.
func NewRecursiveDescentParser() interfaces.Parser {
	return &Parser{}
}

// Parse runs the generated parser against the provided lexer and returns
// the resulting AST program or an error.
func (p *Parser) Parse(lex interfaces.Lexer) (*domain.Program, error) {
	p.lexer = lex
	p.typeRegistry = domain.NewDefaultTypeRegistry()
	p.errors = nil

	// The generated parser expects an object that implements Lex/Error
	// (the yyLexer interface). Our Parser implements those methods below,
	// so we can pass it directly to yyParse.
	rc := yyParse(p)
	if rc != 0 {
		return nil, fmt.Errorf("parse error: %v", p.errors)
	}

	if p.result == nil {
		return nil, fmt.Errorf("no AST produced")
	}
	return p.result, nil
}

// SetErrorReporter satisfies the interfaces.Parser API (no-op for now).
func (p *Parser) SetErrorReporter(reporter domain.ErrorReporter) {
	// Not used by the generated parser directly; parser rules may write
	// errors into p.errors via Error.
}

// Lex implements the lexer interface expected by the generated parser.
// It pulls tokens from the underlying interfaces.Lexer and maps them to
// the token constants generated in parser.go. It also stores the token
// into the semantic value union so parser actions can access it.
func (p *Parser) Lex(lval *yySymType) int {
	tok := p.lexer.NextToken()
	lval.token = tok

	switch tok.Type {
	case interfaces.TokenInt:
		// Map numeric literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenFloat:
		// Map float literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenString:
		// Map string literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenBool:
		// Map boolean literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenIdentifier:
		return IDENTIFIER
	case interfaces.TokenFunc:
		return FUNC
	case interfaces.TokenStruct:
		return STRUCT
	case interfaces.TokenVar:
		return VAR
	case interfaces.TokenIf:
		return IF
	case interfaces.TokenElse:
		return ELSE
	case interfaces.TokenWhile:
		return WHILE
	case interfaces.TokenFor:
		return FOR
	case interfaces.TokenReturn:
		return RETURN
	case interfaces.TokenTrue:
		// Map boolean literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenFalse:
		// Map boolean literals to IDENTIFIER since they're handled as identifiers in grammar
		return IDENTIFIER
	case interfaces.TokenPlus:
		return PLUS
	case interfaces.TokenMinus:
		return MINUS
	case interfaces.TokenStar:
		return STAR
	case interfaces.TokenSlash:
		return SLASH
	case interfaces.TokenPercent:
		return PERCENT
	case interfaces.TokenEqual:
		return EQUAL
	case interfaces.TokenNotEqual:
		return NOT_EQUAL
	case interfaces.TokenLess:
		return LESS
	case interfaces.TokenLessEqual:
		return LESS_EQUAL
	case interfaces.TokenGreater:
		return GREATER
	case interfaces.TokenGreaterEqual:
		return GREATER_EQUAL
	case interfaces.TokenAnd:
		return AND
	case interfaces.TokenOr:
		return OR
	case interfaces.TokenNot:
		return NOT
	case interfaces.TokenAssign:
		return ASSIGN
	case interfaces.TokenLeftParen:
		return LEFT_PAREN
	case interfaces.TokenRightParen:
		return RIGHT_PAREN
	case interfaces.TokenLeftBrace:
		return LEFT_BRACE
	case interfaces.TokenRightBrace:
		return RIGHT_BRACE
	case interfaces.TokenLeftBracket:
		return LEFT_BRACKET
	case interfaces.TokenRightBracket:
		return RIGHT_BRACKET
	case interfaces.TokenSemicolon:
		return SEMICOLON
	case interfaces.TokenComma:
		return COMMA
	case interfaces.TokenDot:
		return DOT
	case interfaces.TokenColon:
		return COLON
	case interfaces.TokenArrow:
		return ARROW
	case interfaces.TokenEOF:
		return EOF
	default:
		// Unknown token -> signal error token
		return 0
	}
}

// Error records a parse error message.
func (p *Parser) Error(s string) {
	p.errors = append(p.errors, s)
}
