%{
// Package grammar contains the Yacc parser for StaticLang
package grammar

import (
	"strconv"
	"fmt"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

%}

%union {
	token      interfaces.Token
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
	str        string
	num        int64
	float      float64
	boolean    bool
}

// Token types
%token <token> INT FLOAT STRING BOOL IDENTIFIER
%token <token> FUNC STRUCT VAR IF ELSE WHILE FOR RETURN TRUE FALSE
%token <token> PLUS MINUS STAR SLASH PERCENT
%token <token> EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL
%token <token> AND OR NOT
%token <token> ASSIGN
%token <token> LEFT_PAREN RIGHT_PAREN LEFT_BRACE RIGHT_BRACE LEFT_BRACKET RIGHT_BRACKET
%token <token> SEMICOLON COMMA DOT COLON
%token <token> ARROW
%token <token> EOF

// Non-terminal types
%type <program> program
%type <decl> declaration function_decl struct_decl global_var_decl main_function
%type <decls> declaration_list
%type <stmt> statement var_decl_stmt assign_stmt if_stmt while_stmt for_stmt return_stmt expr_stmt block_stmt
%type <stmts> statement_list
%type <expr> expression primary_expr call_expr unary_expr binary_expr
%type <exprs> argument_list
%type <param> parameter
%type <params> parameter_list
%type <field> struct_field
%type <fields> struct_field_list
%type <typ> type
%type <str> identifier

// Token precedence for dangling else resolution
%nonassoc LOWER_THAN_ELSE
%nonassoc ELSE

// Operator precedence (lowest to highest)
%left OR
%left AND
%left EQUAL NOT_EQUAL
%left LESS LESS_EQUAL GREATER GREATER_EQUAL
%left PLUS MINUS
%left STAR SLASH PERCENT
%right UNARY_MINUS NOT

%%

program:
	declaration_list {
		$$ = &domain.Program{
			BaseNode:     domain.BaseNode{Location: getLocation($1)},
			Declarations: $1,
		}
		yylex.(*Parser).result = $$
	}
	| /* empty */ {
		$$ = &domain.Program{
			BaseNode:     domain.BaseNode{},
			Declarations: []domain.Declaration{},
		}
		yylex.(*Parser).result = $$
	}

declaration_list:
	declaration {
		$$ = []domain.Declaration{$1}
	}
	| declaration_list declaration {
		$$ = append($1, $2)
	}

declaration:
 	function_decl { $$ = $1 }
 	| struct_decl { $$ = $1 }
 	| global_var_decl { $$ = $1 }
 	| main_function { $$ = $1 }

global_var_decl:
 	type identifier SEMICOLON {
 		$$ = &domain.VarDeclStmt{
 			BaseNode:    domain.BaseNode{Location: $1.GetLocation()},
 			Name:        $2,
 			Type_:       $1,
 			Initializer: nil,
 		}
 	}
 	| type identifier ASSIGN expression SEMICOLON {
 		$$ = &domain.VarDeclStmt{
 			BaseNode:    domain.BaseNode{Location: $1.GetLocation()},
 			Name:        $2,
 			Type_:       $1,
 			Initializer: $4,
 		}
 	}

main_function:
 	type identifier LEFT_PAREN RIGHT_PAREN block_stmt {
 		$$ = &domain.FunctionDecl{
 			BaseNode:   domain.BaseNode{Location: $1.GetLocation()},
 			Name:       $2,
 			Parameters: []domain.Parameter{},
 			ReturnType: $1,
 			Body:       $5.(*domain.BlockStmt),
 		}
 	}

function_decl:
 	FUNC identifier LEFT_PAREN parameter_list RIGHT_PAREN type block_stmt {
 		$$ = &domain.FunctionDecl{
 			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
 			Name:       $2,
 			Parameters: $4,
 			ReturnType: $6,
 			Body:       $7.(*domain.BlockStmt),
 		}
 	}
 	| FUNC identifier LEFT_PAREN RIGHT_PAREN type block_stmt {
 		$$ = &domain.FunctionDecl{
 			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
 			Name:       $2,
 			Parameters: []domain.Parameter{},
 			ReturnType: $5,
 			Body:       $6.(*domain.BlockStmt),
 		}
 	}
 	| FUNC identifier LEFT_PAREN parameter_list RIGHT_PAREN ARROW type block_stmt {
 		$$ = &domain.FunctionDecl{
 			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
 			Name:       $2,
 			Parameters: $4,
 			ReturnType: $7,
 			Body:       $8.(*domain.BlockStmt),
 		}
 	}
 	| FUNC identifier LEFT_PAREN RIGHT_PAREN ARROW type block_stmt {
 		$$ = &domain.FunctionDecl{
 			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
 			Name:       $2,
 			Parameters: []domain.Parameter{},
 			ReturnType: $6,
 			Body:       $7.(*domain.BlockStmt),
 		}
 	}

struct_decl:
	STRUCT identifier LEFT_BRACE struct_field_list RIGHT_BRACE {
		$$ = &domain.StructDecl{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Name:     $2,
			Fields:   $4,
		}
	}
	| STRUCT identifier LEFT_BRACE RIGHT_BRACE {
		$$ = &domain.StructDecl{
			BaseNode: domain.BaseNode{Location: getLocationFromToken($1)},
			Name:     $2,
			Fields:   []domain.StructField{},
		}
	}

parameter_list:
	parameter {
		$$ = []domain.Parameter{$1}
	}
	| parameter_list COMMA parameter {
		$$ = append($1, $3)
	}

parameter:
	identifier type {
		$$ = domain.Parameter{
			Name: $1,
			Type: $2,
		}
	}

struct_field_list:
	struct_field {
		$$ = []domain.StructField{$1}
	}
	| struct_field_list struct_field {
		$$ = append($1, $2)
	}

struct_field:
	identifier type SEMICOLON {
		$$ = domain.StructField{
			Name: $1,
			Type: $2,
		}
	}

type:
	identifier {
		reg := yylex.(*Parser).typeRegistry
		if t, exists := reg.GetType($1); exists {
			$$ = t
		} else {
			$$ = &domain.TypeError{Message: fmt.Sprintf("unknown type: %s", $1)}
		}
	}
	| LEFT_BRACKET INT RIGHT_BRACKET type {
		size, _ := strconv.ParseInt($2.Value, 10, 32)
		$$ = &domain.ArrayType{
			ElementType: $4,
			Size:        int(size),
		}
	}
	| LEFT_BRACKET RIGHT_BRACKET type {
		$$ = &domain.ArrayType{
			ElementType: $3,
			Size:        -1, // dynamic array
		}
	}

statement_list:
	statement {
		$$ = []domain.Statement{$1}
	}
	| statement_list statement {
		$$ = append($1, $2)
	}

statement:
	var_decl_stmt   { $$ = $1 }
	| assign_stmt   { $$ = $1 }
	| if_stmt       { $$ = $1 }
	| while_stmt    { $$ = $1 }
	| for_stmt      { $$ = $1 }
	| return_stmt   { $$ = $1 }
	| expr_stmt     { $$ = $1 }
	| block_stmt    { $$ = $1 }

var_decl_stmt:
	VAR identifier type SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($1)},
			Name:        $2,
			Type_:       $3,
			Initializer: nil,
		}
	}
	| VAR identifier type ASSIGN expression SEMICOLON {
		$$ = &domain.VarDeclStmt{
			BaseNode:    domain.BaseNode{Location: getLocationFromToken($1)},
			Name:        $2,
			Type_:       $3,
			Initializer: $5,
		}
	}

assign_stmt:
	expression ASSIGN expression SEMICOLON {
		$$ = &domain.AssignStmt{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Target:   $1,
			Value:    $3,
		}
	}

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

while_stmt:
	WHILE LEFT_PAREN expression RIGHT_PAREN statement {
		$$ = &domain.WhileStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Condition: $3,
			Body:      $5,
		}
	}

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
	| FOR LEFT_PAREN SEMICOLON expression SEMICOLON statement RIGHT_PAREN statement {
		$$ = &domain.ForStmt{
			BaseNode:  domain.BaseNode{Location: getLocationFromToken($1)},
			Init:      nil,
			Condition: $4,
			Update:    $6,
			Body:      $8,
		}
	}

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

expr_stmt:
	expression SEMICOLON {
		$$ = &domain.ExprStmt{
			BaseNode:   domain.BaseNode{Location: $1.GetLocation()},
			Expression: $1,
		}
	}

block_stmt:
	LEFT_BRACE statement_list RIGHT_BRACE {
		$$ = &domain.BlockStmt{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Statements: $2,
		}
	}
	| LEFT_BRACE RIGHT_BRACE {
		$$ = &domain.BlockStmt{
			BaseNode:   domain.BaseNode{Location: getLocationFromToken($1)},
			Statements: []domain.Statement{},
		}
	}

expression:
	binary_expr { $$ = $1 }

binary_expr:
	unary_expr { $$ = $1 }
	| binary_expr PLUS binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Add,
			Right:    $3,
		}
	}
	| binary_expr MINUS binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Sub,
			Right:    $3,
		}
	}
	| binary_expr STAR binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Mul,
			Right:    $3,
		}
	}
	| binary_expr SLASH binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Div,
			Right:    $3,
		}
	}
	| binary_expr PERCENT binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Mod,
			Right:    $3,
		}
	}
	| binary_expr EQUAL binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Eq,
			Right:    $3,
		}
	}
	| binary_expr NOT_EQUAL binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Ne,
			Right:    $3,
		}
	}
	| binary_expr LESS binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Lt,
			Right:    $3,
		}
	}
	| binary_expr LESS_EQUAL binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Le,
			Right:    $3,
		}
	}
	| binary_expr GREATER binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Gt,
			Right:    $3,
		}
	}
	| binary_expr GREATER_EQUAL binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Ge,
			Right:    $3,
		}
	}
	| binary_expr AND binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.And,
			Right:    $3,
		}
	}
	| binary_expr OR binary_expr {
		$$ = &domain.BinaryExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Left:     $1,
			Operator: domain.Or,
			Right:    $3,
		}
	}

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

call_expr:
	primary_expr { $$ = $1 }
	| call_expr LEFT_PAREN argument_list RIGHT_PAREN {
		$$ = &domain.CallExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Function: $1,
			Args:     $3,
		}
	}
	| call_expr LEFT_PAREN RIGHT_PAREN {
		$$ = &domain.CallExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Function: $1,
			Args:     []domain.Expression{},
		}
	}
	| call_expr LEFT_BRACKET expression RIGHT_BRACKET {
		$$ = &domain.IndexExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Object:   $1,
			Index:    $3,
		}
	}
	| call_expr DOT identifier {
		$$ = &domain.MemberExpr{
			BaseNode: domain.BaseNode{Location: $1.GetLocation()},
			Object:   $1,
			Member:   $3,
		}
	}

argument_list:
	expression {
		$$ = []domain.Expression{$1}
	}
	| argument_list COMMA expression {
		$$ = append($1, $3)
	}

primary_expr:
	identifier {
		$$ = &domain.IdentifierExpr{
			BaseNode: domain.BaseNode{Location: getLocationFromString($1)},
			Name:     $1,
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
	| LEFT_PAREN expression RIGHT_PAREN {
		$$ = $2
	}

identifier:
	IDENTIFIER { $$ = $1.Value }

%%

// Helper functions
func getLocationFromToken(token interfaces.Token) domain.SourceRange {
	pos := token.Location
	return domain.SourceRange{
		Start: pos,
		End:   pos,
	}
}

func getLocationFromString(str string) domain.SourceRange {
	// This is a placeholder - in a real implementation, we'd track positions
	return domain.SourceRange{
		Start: domain.SourcePosition{},
		End:   domain.SourcePosition{},
	}
}

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
