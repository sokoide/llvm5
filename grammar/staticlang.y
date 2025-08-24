%{
// Package grammar contains the Yacc parser for StaticLang
//
// This grammar defines the syntax for the StaticLang programming language,
// a statically-typed language with functions, structs, and control flow.
//
// Grammar Structure:
//   1. Headers and imports
//   2. Union and token definitions  
//   3. Operator precedence rules
//   4. Grammar productions organized by:
//      - Program structure (program, declarations)
//      - Type system (types, parameters, fields)
//      - Statements (control flow, assignments)
//      - Expressions (binary, unary, primary)
//   5. Helper functions for location tracking

package grammar

import (
	"strconv"
	"fmt"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

%}

// =============================================================================
// UNION AND TOKEN DEFINITIONS
// =============================================================================

// Union defines the possible types for grammar symbols
%union {
	// Tokens and primitives
	token      interfaces.Token
	str        string
	num        int64
	float      float64
	boolean    bool

	// AST node types
	program    *domain.Program
	decl       domain.Declaration
	decls      []domain.Declaration
	stmt       domain.Statement
	stmts      []domain.Statement
	expr       domain.Expression
	exprs      []domain.Expression
	param      domain.Parameter
	params     []domain.Parameter
	field      domain.StructField
	fields     []domain.StructField
	typ        domain.Type
}

// =============================================================================
// TOKEN DECLARATIONS
// =============================================================================

// Literal tokens
%token <token> INT FLOAT STRING BOOL IDENTIFIER

// Keywords
%token <token> FUNC STRUCT VAR IF ELSE WHILE FOR RETURN TRUE FALSE

// Arithmetic operators
%token <token> PLUS MINUS STAR SLASH PERCENT

// Comparison operators
%token <token> EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL

// Logical operators
%token <token> AND OR NOT

// Assignment operator
%token <token> ASSIGN

// Delimiters
%token <token> LEFT_PAREN RIGHT_PAREN LEFT_BRACE RIGHT_BRACE LEFT_BRACKET RIGHT_BRACKET

// Punctuation
%token <token> SEMICOLON COMMA DOT COLON ARROW

// =============================================================================
// NON-TERMINAL TYPE DECLARATIONS
// =============================================================================

// Program structure
%type <program> program
%type <decl> declaration function_decl struct_decl global_var_decl
%type <decls> declaration_list

// Statements
%type <stmt> statement var_decl_stmt assign_stmt if_stmt while_stmt for_stmt return_stmt expr_stmt block_stmt
%type <stmts> statement_list

// Expressions
%type <expr> expression primary_expr call_expr unary_expr binary_expr
%type <exprs> argument_list

// Type system
%type <param> parameter
%type <params> parameter_list
%type <field> struct_field
%type <fields> struct_field_list
%type <typ> type

// Utilities
%type <token> identifier

// =============================================================================
// OPERATOR PRECEDENCE AND ASSOCIATIVITY
// =============================================================================

// Dangling else resolution
%nonassoc LOWER_THAN_ELSE
%nonassoc ELSE

// Expression operators (lowest to highest precedence)
%left OR                                    // Logical OR
%left AND                                   // Logical AND
%left EQUAL NOT_EQUAL                      // Equality operators
%left LESS LESS_EQUAL GREATER GREATER_EQUAL // Relational operators
%left PLUS MINUS                           // Additive operators
%left STAR SLASH PERCENT                   // Multiplicative operators
%right UNARY_MINUS NOT                     // Unary operators (highest precedence)

%%

// =============================================================================
// PROGRAM STRUCTURE PRODUCTIONS
// =============================================================================

// Top-level program: a sequence of declarations
program:
	declaration_list {
		ret := &domain.Program{
			BaseNode:     domain.BaseNode{Location: getLocation($1)},
			Declarations: $1,
		}
		yylex.(*Parser).result = ret
		$$ = ret
	}
	| /* empty program */ {
		ret := &domain.Program{
			BaseNode:     domain.BaseNode{},
			Declarations: []domain.Declaration{},
		}
		yylex.(*Parser).result = ret
		$$ = ret
	}

// List of one or more declarations
declaration_list:
	declaration {
		$$ = []domain.Declaration{$1}
	}
	| declaration_list declaration {
		$$ = append($1, $2)
	}

// Top-level declarations: functions, structs, or global variables
declaration:
	function_decl   { $$ = $1 }
	| struct_decl   { $$ = $1 }
	| global_var_decl { $$ = $1 }

// =============================================================================
// GLOBAL VARIABLE DECLARATIONS
// =============================================================================

// Global variable declaration with optional initialization
global_var_decl:
	type identifier SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($2)},
			Name:        $2.Value,
			Type_:       $1,
			Initializer: nil,
		}
	}
	| type identifier ASSIGN expression SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($2)},
			Name:        $2.Value,
			Type_:       $1,
			Initializer: $4,
		}
	}

// =============================================================================
// FUNCTION DECLARATIONS
// =============================================================================

// Function declaration with various parameter and return type combinations
function_decl:
	// Function with parameters and explicit return type
	FUNC identifier LEFT_PAREN parameter_list RIGHT_PAREN ARROW type block_stmt {
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: $4,
			ReturnType: $7,
			Body:       $8.(*domain.BlockStmt),
		}
	}
	// Function without parameters but with explicit return type
	| FUNC identifier LEFT_PAREN RIGHT_PAREN ARROW type block_stmt {
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: []domain.Parameter{},
			ReturnType: $6,
			Body:       $7.(*domain.BlockStmt),
		}
	}
	// Function with parameters and legacy return type syntax (no arrow)
	| FUNC identifier LEFT_PAREN parameter_list RIGHT_PAREN type block_stmt {
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: $4,
			ReturnType: $6,
			Body:       $7.(*domain.BlockStmt),
		}
	}
	// Function without parameters and legacy return type syntax
	| FUNC identifier LEFT_PAREN RIGHT_PAREN type block_stmt {
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: []domain.Parameter{},
			ReturnType: $5,
			Body:       $6.(*domain.BlockStmt),
		}
	}
	// Function with parameters, default return type (int)
	| FUNC identifier LEFT_PAREN parameter_list RIGHT_PAREN block_stmt {
		reg := yylex.(*Parser).typeRegistry
		intType, _ := reg.GetType("int")
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: $4,
			ReturnType: intType,
			Body:       $6.(*domain.BlockStmt),
		}
	}
	// Function without parameters, default return type (int)
	| FUNC identifier LEFT_PAREN RIGHT_PAREN block_stmt {
		reg := yylex.(*Parser).typeRegistry
		intType, _ := reg.GetType("int")
		$$ = &domain.FunctionDecl{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Name:       $2.Value,
			Parameters: []domain.Parameter{},
			ReturnType: intType,
			Body:       $5.(*domain.BlockStmt),
		}
	}

// =============================================================================
// STRUCT DECLARATIONS
// =============================================================================

// Struct declaration with optional field list
struct_decl:
	STRUCT identifier LEFT_BRACE struct_field_list RIGHT_BRACE {
		$$ = &domain.StructDecl{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Name:     $2.Value,
			Fields:   $4,
		}
	}
	| STRUCT identifier LEFT_BRACE RIGHT_BRACE {
		$$ = &domain.StructDecl{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Name:     $2.Value,
			Fields:   []domain.StructField{},
		}
	}

// =============================================================================
// TYPE SYSTEM PRODUCTIONS
// =============================================================================

// Type expressions: basic types and array types
type:
	identifier {
		reg := yylex.(*Parser).typeRegistry
		if t, exists := reg.GetType($1.Value); exists {
			$$ = t
		} else {
			$$ = &domain.TypeError{Message: fmt.Sprintf("unknown type: %s", $1.Value)}
		}
	}
	// Fixed-size array: [size]type
	| LEFT_BRACKET INT RIGHT_BRACKET type {
		size, _ := strconv.ParseInt($2.Value, 10, 32)
		$$ = &domain.ArrayType{
			ElementType: $4,
			Size:        int(size),
		}
	}
	// Dynamic array: []type
	| LEFT_BRACKET RIGHT_BRACKET type {
		$$ = &domain.ArrayType{
			ElementType: $3,
			Size:        -1, // -1 indicates dynamic array
		}
	}

// Function parameter list
parameter_list:
	parameter {
		$$ = []domain.Parameter{$1}
	}
	| parameter_list COMMA parameter {
		$$ = append($1, $3)
	}

// Single function parameter
parameter:
	identifier type {
		$$ = domain.Parameter{
			Name: $1.Value,
			Type: $2,
		}
	}

// Struct field list
struct_field_list:
	struct_field {
		$$ = []domain.StructField{$1}
	}
	| struct_field_list struct_field {
		$$ = append($1, $2)
	}

// Single struct field
struct_field:
	identifier type SEMICOLON {
		$$ = domain.StructField{
			Name: $1.Value,
			Type: $2,
		}
	}

// =============================================================================
// STATEMENT PRODUCTIONS
// =============================================================================

// Statement list (can be empty)
statement_list:
	/* empty */ {
		$$ = []domain.Statement{}
	}
	| statement_list statement {
		$$ = append($1, $2)
	}

// All statement types
statement:
	var_decl_stmt { $$ = $1 }
	| assign_stmt { $$ = $1 }
	| if_stmt     { $$ = $1 }
	| while_stmt  { $$ = $1 }
	| for_stmt    { $$ = $1 }
	| return_stmt { $$ = $1 }
	| expr_stmt   { $$ = $1 }
	| block_stmt  { $$ = $1 }

// Local variable declaration
var_decl_stmt:
	VAR identifier type SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($1)},
			Name:        $2.Value,
			Type_:       $3,
			Initializer: nil,
		}
	}
	| VAR identifier type ASSIGN expression SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($1)},
			Name:        $2.Value,
			Type_:       $3,
			Initializer: $5,
		}
	}

// Assignment statement
assign_stmt:
	expression ASSIGN expression SEMICOLON {
		$$ = &domain.AssignStmt{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Target:   $1,
			Value:    $3,
		}
	}

// If statement with optional else clause
if_stmt:
	IF LEFT_PAREN expression RIGHT_PAREN statement %prec LOWER_THAN_ELSE {
		$$ = &domain.IfStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Condition: $3,
			ThenStmt:  $5,
			ElseStmt:  nil,
		}
	}
	| IF LEFT_PAREN expression RIGHT_PAREN statement ELSE statement {
		$$ = &domain.IfStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Condition: $3,
			ThenStmt:  $5,
			ElseStmt:  $7,
		}
	}

// While loop
while_stmt:
	WHILE LEFT_PAREN expression RIGHT_PAREN statement {
		$$ = &domain.WhileStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Condition: $3,
			Body:      $5,
		}
	}

// For loop with optional init, condition, and update
for_stmt:
	FOR LEFT_PAREN statement expression SEMICOLON statement RIGHT_PAREN statement {
		$$ = &domain.ForStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Init:      $3,
			Condition: $4,
			Update:    $6,
			Body:      $8,
		}
	}
	// For loop with no init statement
	| FOR LEFT_PAREN SEMICOLON expression SEMICOLON statement RIGHT_PAREN statement {
		$$ = &domain.ForStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Init:      nil,
			Condition: $4,
			Update:    $6,
			Body:      $8,
		}
	}

// Return statement with optional value
return_stmt:
	RETURN SEMICOLON {
		$$ = &domain.ReturnStmt{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    nil,
		}
	}
	| RETURN expression SEMICOLON {
		$$ = &domain.ReturnStmt{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    $2,
		}
	}

// Expression statement
expr_stmt:
	expression SEMICOLON {
		$$ = &domain.ExprStmt{
			BaseNode:   domain.BaseNode{Location: $1.GetLocation()},
			Expression: $1,
		}
	}

// Block statement (compound statement)
block_stmt:
	LEFT_BRACE statement_list RIGHT_BRACE {
		$$ = &domain.BlockStmt{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Statements: $2,
		}
	}

// =============================================================================
// EXPRESSION PRODUCTIONS
// =============================================================================

// Top-level expression
expression:
	binary_expr { $$ = $1 }

// Binary expressions with proper precedence handling
binary_expr:
	unary_expr { $$ = $1 }
	
	// Arithmetic operators
	| binary_expr PLUS binary_expr {
		$$ = createBinaryExpr($1, domain.Add, $3)
	}
	| binary_expr MINUS binary_expr {
		$$ = createBinaryExpr($1, domain.Sub, $3)
	}
	| binary_expr STAR binary_expr {
		$$ = createBinaryExpr($1, domain.Mul, $3)
	}
	| binary_expr SLASH binary_expr {
		$$ = createBinaryExpr($1, domain.Div, $3)
	}
	| binary_expr PERCENT binary_expr {
		$$ = createBinaryExpr($1, domain.Mod, $3)
	}
	
	// Comparison operators
	| binary_expr EQUAL binary_expr {
		$$ = createBinaryExpr($1, domain.Eq, $3)
	}
	| binary_expr NOT_EQUAL binary_expr {
		$$ = createBinaryExpr($1, domain.Ne, $3)
	}
	| binary_expr LESS binary_expr {
		$$ = createBinaryExpr($1, domain.Lt, $3)
	}
	| binary_expr LESS_EQUAL binary_expr {
		$$ = createBinaryExpr($1, domain.Le, $3)
	}
	| binary_expr GREATER binary_expr {
		$$ = createBinaryExpr($1, domain.Gt, $3)
	}
	| binary_expr GREATER_EQUAL binary_expr {
		$$ = createBinaryExpr($1, domain.Ge, $3)
	}
	
	// Logical operators
	| binary_expr AND binary_expr {
		$$ = createBinaryExpr($1, domain.And, $3)
	}
	| binary_expr OR binary_expr {
		$$ = createBinaryExpr($1, domain.Or, $3)
	}

// Unary expressions
unary_expr:
	call_expr { $$ = $1 }
	| MINUS unary_expr %prec UNARY_MINUS {
		$$ = &domain.UnaryExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Operator: domain.Neg,
			Operand:  $2,
		}
	}
	| NOT unary_expr {
		$$ = &domain.UnaryExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Operator: domain.Not,
			Operand:  $2,
		}
	}

// Call expressions and postfix operators
call_expr:
	primary_expr { $$ = $1 }
	
	// Function call with arguments
	| call_expr LEFT_PAREN argument_list RIGHT_PAREN {
		$$ = &domain.CallExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Function: $1,
			Args:     $3,
		}
	}
	// Function call without arguments
	| call_expr LEFT_PAREN RIGHT_PAREN {
		$$ = &domain.CallExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Function: $1,
			Args:     []domain.Expression{},
		}
	}
	
	// Array indexing
	| call_expr LEFT_BRACKET expression RIGHT_BRACKET {
		$$ = &domain.IndexExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Object:   $1,
			Index:    $3,
		}
	}
	
	// Member access
	| call_expr DOT identifier {
		$$ = &domain.MemberExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Object:   $1,
			Member:   $3.Value,
		}
	}

// Function call argument list
argument_list:
	expression {
		$$ = []domain.Expression{$1}
	}
	| argument_list COMMA expression {
		$$ = append($1, $3)
	}

// Primary expressions (atoms)
primary_expr:
	identifier {
		$$ = &domain.IdentifierExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Name:     $1.Value,
		}
	}
	| INT {
		val, _ := strconv.ParseInt($1.Value, 10, 64)
		$$ = &domain.LiteralExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    val,
		}
	}
	| FLOAT {
		val, _ := strconv.ParseFloat($1.Value, 64)
		$$ = &domain.LiteralExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    val,
		}
	}
	| STRING {
		$$ = &domain.LiteralExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    $1.Value,
		}
	}
	| TRUE {
		$$ = &domain.LiteralExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    true,
		}
	}
	| FALSE {
		$$ = &domain.LiteralExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Value:    false,
		}
	}
	// Parenthesized expression
	| LEFT_PAREN expression RIGHT_PAREN {
		$$ = $2
	}

// =============================================================================
// UTILITY PRODUCTIONS
// =============================================================================

// Identifier token wrapper
identifier:
	IDENTIFIER { $$ = $1 }

%%

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// createBinaryExpr creates a binary expression node with proper location tracking
func createBinaryExpr(left domain.Expression, op domain.BinaryOperator, right domain.Expression) *domain.BinaryExpr {
	return &domain.BinaryExpr{
		BaseNode: domain.BaseNode{Location: left.GetLocation()},
		Left:     left,
		Operator: op,
		Right:    right,
	}
}

// getLocationFromToken extracts source location from a token
func getLocationFromToken(token interfaces.Token) domain.SourceRange {
	pos := token.Location
	return domain.SourceRange{
		Start: pos,
		End:   pos,
	}
}

// getLocationFromString creates a placeholder location for string-based nodes
// TODO: Implement proper position tracking for string literals
func getLocationFromString(str string) domain.SourceRange {
	return domain.SourceRange{
		Start: domain.SourcePosition{},
		End:   domain.SourcePosition{},
	}
}

// getLocation determines the source range for a list of declarations
func getLocation(decls []domain.Declaration) domain.SourceRange {
	if len(decls) == 0 {
		return domain.SourceRange{}
	}
	start := decls[0].GetLocation()
	end := decls[len(decls)-1].GetLocation()
	return domain.SourceRange{
		Start: start.Start,
		End:   end.End,
	}
}