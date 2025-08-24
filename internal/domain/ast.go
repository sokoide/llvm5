// Package domain contains AST node definitions and visitor interfaces
package domain

// Node represents the base interface for all AST nodes
type Node interface {
	GetLocation() SourceRange
	Accept(visitor Visitor) error
}

// Visitor defines the visitor pattern interface for AST traversal
type Visitor interface {
	// Expressions
	VisitLiteralExpr(expr *LiteralExpr) error
	VisitIdentifierExpr(expr *IdentifierExpr) error
	VisitBinaryExpr(expr *BinaryExpr) error
	VisitUnaryExpr(expr *UnaryExpr) error
	VisitCallExpr(expr *CallExpr) error
	VisitIndexExpr(expr *IndexExpr) error
	VisitMemberExpr(expr *MemberExpr) error

	// Statements
	VisitExprStmt(stmt *ExprStmt) error
	VisitVarDeclStmt(stmt *VarDeclStmt) error
	VisitAssignStmt(stmt *AssignStmt) error
	VisitIfStmt(stmt *IfStmt) error
	VisitWhileStmt(stmt *WhileStmt) error
	VisitForStmt(stmt *ForStmt) error
	VisitReturnStmt(stmt *ReturnStmt) error
	VisitBlockStmt(stmt *BlockStmt) error

	// Declarations
	VisitFunctionDecl(decl *FunctionDecl) error
	VisitStructDecl(decl *StructDecl) error
	VisitProgram(prog *Program) error
}

// Expression interface for all expression nodes
type Expression interface {
	Node
	GetType() Type
	SetType(t Type)
}

// Statement interface for all statement nodes
type Statement interface {
	Node
}

// Declaration interface for all declaration nodes
type Declaration interface {
	Node
	GetName() string
}

// BaseNode provides common functionality for all AST nodes
type BaseNode struct {
	Location SourceRange
}

func (b *BaseNode) GetLocation() SourceRange {
	return b.Location
}

// Expression nodes
type LiteralExpr struct {
	BaseNode
	Value interface{}
	Type_ Type
}

func (e *LiteralExpr) Accept(visitor Visitor) error { return visitor.VisitLiteralExpr(e) }
func (e *LiteralExpr) GetType() Type                { return e.Type_ }
func (e *LiteralExpr) SetType(t Type)               { e.Type_ = t }

type IdentifierExpr struct {
	BaseNode
	Name  string
	Type_ Type
}

func (e *IdentifierExpr) Accept(visitor Visitor) error { return visitor.VisitIdentifierExpr(e) }
func (e *IdentifierExpr) GetType() Type                { return e.Type_ }
func (e *IdentifierExpr) SetType(t Type)               { e.Type_ = t }

type BinaryExpr struct {
	BaseNode
	Left     Expression
	Operator BinaryOperator
	Right    Expression
	Type_    Type
}

func (e *BinaryExpr) Accept(visitor Visitor) error { return visitor.VisitBinaryExpr(e) }
func (e *BinaryExpr) GetType() Type                { return e.Type_ }
func (e *BinaryExpr) SetType(t Type)               { e.Type_ = t }

type UnaryExpr struct {
	BaseNode
	Operator UnaryOperator
	Operand  Expression
	Type_    Type
}

func (e *UnaryExpr) Accept(visitor Visitor) error { return visitor.VisitUnaryExpr(e) }
func (e *UnaryExpr) GetType() Type                { return e.Type_ }
func (e *UnaryExpr) SetType(t Type)               { e.Type_ = t }

type CallExpr struct {
	BaseNode
	Function Expression
	Args     []Expression
	Type_    Type
}

func (e *CallExpr) Accept(visitor Visitor) error { return visitor.VisitCallExpr(e) }
func (e *CallExpr) GetType() Type                { return e.Type_ }
func (e *CallExpr) SetType(t Type)               { e.Type_ = t }

type IndexExpr struct {
	BaseNode
	Object Expression
	Index  Expression
	Type_  Type
}

func (e *IndexExpr) Accept(visitor Visitor) error { return visitor.VisitIndexExpr(e) }
func (e *IndexExpr) GetType() Type                { return e.Type_ }
func (e *IndexExpr) SetType(t Type)               { e.Type_ = t }

type MemberExpr struct {
	BaseNode
	Object Expression
	Member string
	Type_  Type
}

func (e *MemberExpr) Accept(visitor Visitor) error { return visitor.VisitMemberExpr(e) }
func (e *MemberExpr) GetType() Type                { return e.Type_ }
func (e *MemberExpr) SetType(t Type)               { e.Type_ = t }

// Statement nodes
type ExprStmt struct {
	BaseNode
	Expression Expression
}

func (s *ExprStmt) Accept(visitor Visitor) error { return visitor.VisitExprStmt(s) }

type VarDeclStmt struct {
	BaseNode
	Name        string
	Type_       Type
	Initializer Expression
}

func (s *VarDeclStmt) Accept(visitor Visitor) error { return visitor.VisitVarDeclStmt(s) }
func (s *VarDeclStmt) GetName() string              { return s.Name }

type AssignStmt struct {
	BaseNode
	Target Expression
	Value  Expression
}

func (s *AssignStmt) Accept(visitor Visitor) error { return visitor.VisitAssignStmt(s) }

type IfStmt struct {
	BaseNode
	Condition Expression
	ThenStmt  Statement
	ElseStmt  Statement // optional
}

func (s *IfStmt) Accept(visitor Visitor) error { return visitor.VisitIfStmt(s) }

type WhileStmt struct {
	BaseNode
	Condition Expression
	Body      Statement
}

func (s *WhileStmt) Accept(visitor Visitor) error { return visitor.VisitWhileStmt(s) }

type ForStmt struct {
	BaseNode
	Init      Statement  // optional
	Condition Expression // optional
	Update    Statement  // optional
	Body      Statement
}

func (s *ForStmt) Accept(visitor Visitor) error { return visitor.VisitForStmt(s) }

type ReturnStmt struct {
	BaseNode
	Value Expression // optional
}

func (s *ReturnStmt) Accept(visitor Visitor) error { return visitor.VisitReturnStmt(s) }

type BlockStmt struct {
	BaseNode
	Statements []Statement
}

func (s *BlockStmt) Accept(visitor Visitor) error { return visitor.VisitBlockStmt(s) }

// Declaration nodes
type Parameter struct {
	Name string
	Type Type
}

type FunctionDecl struct {
	BaseNode
	Name       string
	Parameters []Parameter
	ReturnType Type
	Body       *BlockStmt
}

func (d *FunctionDecl) Accept(visitor Visitor) error { return visitor.VisitFunctionDecl(d) }
func (d *FunctionDecl) GetName() string              { return d.Name }

type StructField struct {
	Name string
	Type Type
}

type StructDecl struct {
	BaseNode
	Name   string
	Fields []StructField
}

func (d *StructDecl) Accept(visitor Visitor) error { return visitor.VisitStructDecl(d) }
func (d *StructDecl) GetName() string              { return d.Name }

type Program struct {
	BaseNode
	Declarations []Declaration
}

func (p *Program) Accept(visitor Visitor) error { return visitor.VisitProgram(p) }

// Operator types
type BinaryOperator int

const (
	// Arithmetic
	Add BinaryOperator = iota
	Sub
	Mul
	Div
	Mod

	// Comparison
	Eq
	Ne
	Lt
	Le
	Gt
	Ge

	// Logical
	And
	Or
)

func (op BinaryOperator) String() string {
	switch op {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"
	case Mod:
		return "%"
	case Eq:
		return "=="
	case Ne:
		return "!="
	case Lt:
		return "<"
	case Le:
		return "<="
	case Gt:
		return ">"
	case Ge:
		return ">="
	case And:
		return "&&"
	case Or:
		return "||"
	default:
		return "unknown"
	}
}

type UnaryOperator int

const (
	Neg UnaryOperator = iota // -
	Not                      // !
)

func (op UnaryOperator) String() string {
	switch op {
	case Neg:
		return "-"
	case Not:
		return "!"
	default:
		return "unknown"
	}
}
