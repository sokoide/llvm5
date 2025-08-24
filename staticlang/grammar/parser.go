// Package grammar provides the parser for StaticLang
package grammar

import (
	"fmt"
	"io"

	"github.com/sokoide/llvm5/staticlang/internal/domain"
	"github.com/sokoide/llvm5/staticlang/internal/interfaces"
)

// Parser implements the Parser interface using the Yacc-generated parser
type Parser struct {
	lexer         interfaces.Lexer
	errorReporter domain.ErrorReporter
	typeRegistry  domain.TypeRegistry
	result        *domain.Program
	currentToken  interfaces.Token
}

// NewParser creates a new StaticLang parser
func NewParser() *Parser {
	return &Parser{
		typeRegistry: domain.NewDefaultTypeRegistry(),
	}
}

// Parse parses the input using the provided lexer and returns an AST
func (p *Parser) Parse(lexer interfaces.Lexer) (*domain.Program, error) {
	p.lexer = lexer
	p.result = nil
	p.currentToken = lexer.NextToken()

	// Call the Yacc-generated parser
	result := yyParse(p)

	if result != 0 {
		return nil, fmt.Errorf("parse error")
	}

	if p.result == nil {
		return nil, fmt.Errorf("no result from parser")
	}

	return p.result, nil
}

// SetErrorReporter sets the error reporter for the parser
func (p *Parser) SetErrorReporter(reporter domain.ErrorReporter) {
	p.errorReporter = reporter
}

// SetTypeRegistry sets the type registry for the parser
func (p *Parser) SetTypeRegistry(registry domain.TypeRegistry) {
	p.typeRegistry = registry
}

// Lexer interface implementation for Yacc
func (p *Parser) Lex(lval *yySymType) int {
	if p.currentToken.Type == interfaces.TokenEOF {
		return 0 // EOF
	}

	lval.token = p.currentToken
	tokenType := p.mapTokenType(p.currentToken.Type)

	// Advance to next token
	p.currentToken = p.lexer.NextToken()

	return tokenType
}

// Error handling for Yacc
func (p *Parser) Error(s string) {
	if p.errorReporter != nil {
		err := domain.CompilerError{
			Type:    domain.SyntaxError,
			Message: s,
			Location: domain.SourceRange{
				Start: p.currentToken.Location,
				End:   p.currentToken.Location,
			},
			Context: fmt.Sprintf("at token: %s", p.currentToken.Value),
		}
		p.errorReporter.ReportError(err)
	}
}

// mapTokenType maps our token types to Yacc token constants
func (p *Parser) mapTokenType(tokenType interfaces.TokenType) int {
	switch tokenType {
	case interfaces.TokenInt:
		return INT
	case interfaces.TokenFloat:
		return FLOAT
	case interfaces.TokenString:
		return STRING
	case interfaces.TokenBool:
		return BOOL
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
		return TRUE
	case interfaces.TokenFalse:
		return FALSE
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
	case interfaces.TokenEOF:
		return EOF
	case interfaces.TokenError:
		if p.errorReporter != nil {
			err := domain.CompilerError{
				Type:    domain.LexicalError,
				Message: p.currentToken.Value,
				Location: domain.SourceRange{
					Start: p.currentToken.Location,
					End:   p.currentToken.Location,
				},
			}
			p.errorReporter.ReportError(err)
		}
		return 0 // Skip error tokens
	default:
		return 0 // Unknown token
	}
}

// RecursiveDescentParser provides an alternative parser implementation
// that doesn't depend on Yacc (useful for testing or when Yacc is unavailable)
type RecursiveDescentParser struct {
	lexer         interfaces.Lexer
	errorReporter domain.ErrorReporter
	typeRegistry  domain.TypeRegistry
	currentToken  interfaces.Token
	peekToken     interfaces.Token
}

// NewRecursiveDescentParser creates a new recursive descent parser
func NewRecursiveDescentParser() *RecursiveDescentParser {
	return &RecursiveDescentParser{
		typeRegistry: domain.NewDefaultTypeRegistry(),
	}
}

// Parse parses the input using recursive descent
func (p *RecursiveDescentParser) Parse(lexer interfaces.Lexer) (*domain.Program, error) {
	p.lexer = lexer
	p.nextToken()
	p.nextToken() // Fill both current and peek

	return p.parseProgram()
}

// SetErrorReporter sets the error reporter
func (p *RecursiveDescentParser) SetErrorReporter(reporter domain.ErrorReporter) {
	p.errorReporter = reporter
}

// Helper methods for recursive descent parsing
func (p *RecursiveDescentParser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *RecursiveDescentParser) expectToken(tokenType interfaces.TokenType) error {
	if p.currentToken.Type != tokenType {
		return fmt.Errorf("expected %s, got %s", tokenType, p.currentToken.Type)
	}
	p.nextToken()
	return nil
}

func (p *RecursiveDescentParser) parseProgram() (*domain.Program, error) {
	var declarations []domain.Declaration

	for p.currentToken.Type != interfaces.TokenEOF {
		decl, err := p.parseDeclaration()
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, decl)
	}

	return &domain.Program{
		BaseNode:     domain.BaseNode{},
		Declarations: declarations,
	}, nil
}

func (p *RecursiveDescentParser) parseDeclaration() (domain.Declaration, error) {
	switch p.currentToken.Type {
	case interfaces.TokenFunc:
		return p.parseFunctionDecl()
	case interfaces.TokenStruct:
		return p.parseStructDecl()
	default:
		return nil, fmt.Errorf("expected declaration, got %s", p.currentToken.Value)
	}
}

func (p *RecursiveDescentParser) parseFunctionDecl() (*domain.FunctionDecl, error) {
	location := p.currentToken.Location
	p.nextToken() // consume 'func'

	if p.currentToken.Type != interfaces.TokenIdentifier {
		return nil, fmt.Errorf("expected function name")
	}
	name := p.currentToken.Value
	p.nextToken()

	if err := p.expectToken(interfaces.TokenLeftParen); err != nil {
		return nil, err
	}

	var parameters []domain.Parameter
	if p.currentToken.Type != interfaces.TokenRightParen {
		params, err := p.parseParameterList()
		if err != nil {
			return nil, err
		}
		parameters = params
	}

	if err := p.expectToken(interfaces.TokenRightParen); err != nil {
		return nil, err
	}

	returnType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}

	return &domain.FunctionDecl{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{
				Start: location,
				End:   body.GetLocation().End,
			},
		},
		Name:       name,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

func (p *RecursiveDescentParser) parseStructDecl() (*domain.StructDecl, error) {
	// Implementation similar to parseFunctionDecl but for structs
	return nil, fmt.Errorf("struct declarations not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseParameterList() ([]domain.Parameter, error) {
	var params []domain.Parameter

	for {
		if p.currentToken.Type != interfaces.TokenIdentifier {
			break
		}

		name := p.currentToken.Value
		p.nextToken()

		paramType, err := p.parseType()
		if err != nil {
			return nil, err
		}

		params = append(params, domain.Parameter{
			Name: name,
			Type: paramType,
		})

		if p.currentToken.Type != interfaces.TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return params, nil
}

func (p *RecursiveDescentParser) parseType() (domain.Type, error) {
	if p.currentToken.Type != interfaces.TokenIdentifier {
		return nil, fmt.Errorf("expected type name")
	}

	typeName := p.currentToken.Value
	p.nextToken()

	if t, exists := p.typeRegistry.GetType(typeName); exists {
		return t, nil
	}

	return &domain.TypeError{Message: fmt.Sprintf("unknown type: %s", typeName)}, nil
}

func (p *RecursiveDescentParser) parseBlockStmt() (*domain.BlockStmt, error) {
	location := p.currentToken.Location

	if err := p.expectToken(interfaces.TokenLeftBrace); err != nil {
		return nil, err
	}

	var statements []domain.Statement
	for p.currentToken.Type != interfaces.TokenRightBrace && p.currentToken.Type != interfaces.TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	endLocation := p.currentToken.Location
	if err := p.expectToken(interfaces.TokenRightBrace); err != nil {
		return nil, err
	}

	return &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{
				Start: location,
				End:   endLocation,
			},
		},
		Statements: statements,
	}, nil
}

func (p *RecursiveDescentParser) parseStatement() (domain.Statement, error) {
	switch p.currentToken.Type {
	case interfaces.TokenVar:
		return p.parseVarDeclStmt()
	case interfaces.TokenIf:
		return p.parseIfStmt()
	case interfaces.TokenWhile:
		return p.parseWhileStmt()
	case interfaces.TokenFor:
		return p.parseForStmt()
	case interfaces.TokenReturn:
		return p.parseReturnStmt()
	case interfaces.TokenLeftBrace:
		return p.parseBlockStmt()
	default:
		// Try to parse as expression statement or assignment
		return p.parseExpressionOrAssignmentStmt()
	}
}

func (p *RecursiveDescentParser) parseVarDeclStmt() (*domain.VarDeclStmt, error) {
	// Implementation for variable declarations
	return nil, fmt.Errorf("variable declarations not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseIfStmt() (*domain.IfStmt, error) {
	// Implementation for if statements
	return nil, fmt.Errorf("if statements not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseWhileStmt() (*domain.WhileStmt, error) {
	// Implementation for while statements
	return nil, fmt.Errorf("while statements not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseForStmt() (*domain.ForStmt, error) {
	// Implementation for for statements
	return nil, fmt.Errorf("for statements not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseReturnStmt() (*domain.ReturnStmt, error) {
	location := p.currentToken.Location
	p.nextToken() // consume 'return'

	var value domain.Expression
	if p.currentToken.Type != interfaces.TokenSemicolon {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		value = expr
	}

	if err := p.expectToken(interfaces.TokenSemicolon); err != nil {
		return nil, err
	}

	return &domain.ReturnStmt{
		BaseNode: domain.BaseNode{Location: domain.SourceRange{Start: location, End: location}},
		Value:    value,
	}, nil
}

func (p *RecursiveDescentParser) parseExpressionOrAssignmentStmt() (domain.Statement, error) {
	// Implementation for expressions and assignments
	return nil, fmt.Errorf("expression statements not yet implemented in recursive descent parser")
}

func (p *RecursiveDescentParser) parseExpression() (domain.Expression, error) {
	// Implementation for expressions
	return nil, fmt.Errorf("expressions not yet implemented in recursive descent parser")
}
