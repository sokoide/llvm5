package domain

import (
	"testing"
)

// MockVisitor for testing Accept methods
type MockVisitor struct {
	visitedNodes []Node
}

func (mv *MockVisitor) VisitProgram(node *Program) error         { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitFunctionDecl(node *FunctionDecl) error { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitStructDecl(node *StructDecl) error   { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitBlockStmt(node *BlockStmt) error     { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitVarDeclStmt(node *VarDeclStmt) error { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitAssignStmt(node *AssignStmt) error   { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitIfStmt(node *IfStmt) error           { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitWhileStmt(node *WhileStmt) error     { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitForStmt(node *ForStmt) error         { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitReturnStmt(node *ReturnStmt) error   { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitExprStmt(node *ExprStmt) error       { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitBinaryExpr(node *BinaryExpr) error   { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitUnaryExpr(node *UnaryExpr) error     { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitCallExpr(node *CallExpr) error       { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitIdentifierExpr(node *IdentifierExpr) error { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitLiteralExpr(node *LiteralExpr) error { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitIndexExpr(node *IndexExpr) error     { mv.visitedNodes = append(mv.visitedNodes, node); return nil }
func (mv *MockVisitor) VisitMemberExpr(node *MemberExpr) error   { mv.visitedNodes = append(mv.visitedNodes, node); return nil }

// TestIndexExprComplete tests IndexExpr with all methods
func TestIndexExprComplete(t *testing.T) {
	indexType := &BasicType{Kind: IntType}
	
	expr := &IndexExpr{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Object:   &IdentifierExpr{Name: "arr"},
		Index:    &LiteralExpr{Value: 0, Type_: &BasicType{Kind: IntType}},
		Type_:    indexType,
	}

	// Test GetType
	if expr.GetType() != indexType {
		t.Error("IndexExpr GetType should return correct type")
	}

	// Test SetType
	newType := &BasicType{Kind: StringType}
	expr.SetType(newType)
	if expr.GetType() != newType {
		t.Error("IndexExpr SetType should update type")
	}

	// Test Accept
	visitor := &MockVisitor{}
	expr.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != expr {
		t.Error("IndexExpr Accept should call visitor")
	}
}

// TestMemberExprComplete tests MemberExpr with all methods
func TestMemberExprComplete(t *testing.T) {
	memberType := &BasicType{Kind: IntType}
	
	expr := &MemberExpr{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Object:   &IdentifierExpr{Name: "obj"},
		Member:   "field",
		Type_:    memberType,
	}

	// Test GetType
	if expr.GetType() != memberType {
		t.Error("MemberExpr GetType should return correct type")
	}

	// Test SetType
	newType := &BasicType{Kind: StringType}
	expr.SetType(newType)
	if expr.GetType() != newType {
		t.Error("MemberExpr SetType should update type")
	}

	// Test Accept
	visitor := &MockVisitor{}
	expr.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != expr {
		t.Error("MemberExpr Accept should call visitor")
	}
}

// TestExprStmtComplete tests ExprStmt with Accept method
func TestExprStmtComplete(t *testing.T) {
	stmt := &ExprStmt{
		BaseNode:   BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Expression: &LiteralExpr{Value: 42, Type_: &BasicType{Kind: IntType}},
	}

	// Test Accept
	visitor := &MockVisitor{}
	stmt.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != stmt {
		t.Error("ExprStmt Accept should call visitor")
	}
}

// TestWhileStmtComplete tests WhileStmt with Accept method
func TestWhileStmtComplete(t *testing.T) {
	stmt := &WhileStmt{
		BaseNode:  BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Condition: &LiteralExpr{Value: true, Type_: &BasicType{Kind: BoolType}},
		Body:      &BlockStmt{Statements: []Statement{}},
	}

	// Test Accept
	visitor := &MockVisitor{}
	stmt.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != stmt {
		t.Error("WhileStmt Accept should call visitor")
	}
}

// TestForStmtComplete tests ForStmt with Accept method
func TestForStmtComplete(t *testing.T) {
	stmt := &ForStmt{
		BaseNode:  BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Init:      &VarDeclStmt{Name: "i", Type_: &BasicType{Kind: IntType}},
		Condition: &BinaryExpr{Left: &IdentifierExpr{Name: "i"}, Operator: Lt, Right: &LiteralExpr{Value: 10}},
		Update:    &AssignStmt{Target: &IdentifierExpr{Name: "i"}},
		Body:      &BlockStmt{Statements: []Statement{}},
	}

	// Test Accept
	visitor := &MockVisitor{}
	stmt.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != stmt {
		t.Error("ForStmt Accept should call visitor")
	}
}

// TestReturnStmtComplete tests ReturnStmt with Accept method
func TestReturnStmtComplete(t *testing.T) {
	stmt := &ReturnStmt{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Value:    &LiteralExpr{Value: 42, Type_: &BasicType{Kind: IntType}},
	}

	// Test Accept
	visitor := &MockVisitor{}
	stmt.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != stmt {
		t.Error("ReturnStmt Accept should call visitor")
	}
}

// TestBlockStmtComplete tests BlockStmt with Accept method
func TestBlockStmtComplete(t *testing.T) {
	stmt := &BlockStmt{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Statements: []Statement{
			&ExprStmt{Expression: &LiteralExpr{Value: 1}},
			&ExprStmt{Expression: &LiteralExpr{Value: 2}},
		},
	}

	// Test Accept
	visitor := &MockVisitor{}
	stmt.Accept(visitor)
	if len(visitor.visitedNodes) != 1 || visitor.visitedNodes[0] != stmt {
		t.Error("BlockStmt Accept should call visitor")
	}
}

// TestFunctionDeclGetName tests FunctionDecl GetName method
func TestFunctionDeclGetName(t *testing.T) {
	decl := &FunctionDecl{
		BaseNode:   BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Name:       "testFunction",
		Parameters: []Parameter{},
		ReturnType: &BasicType{Kind: VoidType},
		Body:       &BlockStmt{Statements: []Statement{}},
	}

	if decl.GetName() != "testFunction" {
		t.Error("FunctionDecl GetName should return function name")
	}
}

// TestStructDeclGetName tests StructDecl GetName method
func TestStructDeclGetName(t *testing.T) {
	decl := &StructDecl{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Name:     "TestStruct",
		Fields:   []StructField{},
	}

	if decl.GetName() != "TestStruct" {
		t.Error("StructDecl GetName should return struct name")
	}
}

// TestVarDeclStmtGetName tests VarDeclStmt GetName method
func TestVarDeclStmtGetName(t *testing.T) {
	stmt := &VarDeclStmt{
		BaseNode: BaseNode{Location: SourceRange{Start: SourcePosition{Line: 1, Column: 1}}},
		Name:     "testVar",
		Type_:    &BasicType{Kind: IntType},
		Initializer: &LiteralExpr{Value: 42},
	}

	if stmt.GetName() != "testVar" {
		t.Error("VarDeclStmt GetName should return variable name")
	}
}

// TestParameterStruct tests Parameter struct
func TestParameterStruct(t *testing.T) {
	param := Parameter{
		Name: "param1",
		Type: &BasicType{Kind: IntType},
	}

	if param.Name != "param1" {
		t.Error("Parameter should store name correctly")
	}
	
	if param.Type.String() != "int" {
		t.Error("Parameter should store type correctly")
	}
}

// TestStructFieldStruct tests StructField struct
func TestStructFieldStruct(t *testing.T) {
	field := StructField{
		Name: "field1",
		Type: &BasicType{Kind: StringType},
	}

	if field.Name != "field1" {
		t.Error("StructField should store name correctly")
	}
	
	if field.Type.String() != "string" {
		t.Error("StructField should store type correctly")
	}
}

// TestBinaryOperatorEdgeCases tests binary operator edge cases
func TestBinaryOperatorEdgeCases(t *testing.T) {
	// Test invalid operator
	invalidOp := BinaryOperator(999)
	if invalidOp.String() != "unknown" {
		t.Error("Invalid binary operator should return 'unknown'")
	}
}

// TestUnaryOperatorEdgeCases tests unary operator edge cases
func TestUnaryOperatorEdgeCases(t *testing.T) {
	// Test invalid operator
	invalidOp := UnaryOperator(999)
	if invalidOp.String() != "unknown" {
		t.Error("Invalid unary operator should return 'unknown'")
	}
}

// TestAllExpressionTypes tests type methods on all expression types
func TestAllExpressionTypes(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
	}{
		{
			"LiteralExpr",
			&LiteralExpr{Value: 42, Type_: &BasicType{Kind: IntType}},
		},
		{
			"IdentifierExpr",
			&IdentifierExpr{Name: "x", Type_: &BasicType{Kind: IntType}},
		},
		{
			"BinaryExpr",
			&BinaryExpr{
				Left:     &LiteralExpr{Value: 1},
				Operator: Add,
				Right:    &LiteralExpr{Value: 2},
				Type_:    &BasicType{Kind: IntType},
			},
		},
		{
			"UnaryExpr",
			&UnaryExpr{
				Operator: Neg,
				Operand:  &LiteralExpr{Value: 5},
				Type_:    &BasicType{Kind: IntType},
			},
		},
		{
			"CallExpr",
			&CallExpr{
				Function: &IdentifierExpr{Name: "func"},
				Args:     []Expression{},
				Type_:    &BasicType{Kind: IntType},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that all expressions support type operations
			originalType := tt.expr.GetType()
			
			newType := &BasicType{Kind: StringType}
			tt.expr.SetType(newType)
			
			if tt.expr.GetType() != newType {
				t.Errorf("%s should support SetType/GetType", tt.name)
			}
			
			// Test Accept method with visitor
			visitor := &MockVisitor{}
			tt.expr.Accept(visitor)
			
			if len(visitor.visitedNodes) != 1 {
				t.Errorf("%s Accept should call visitor once", tt.name)
			}

			// Restore original type for other tests
			tt.expr.SetType(originalType)
		})
	}
}

// TestAllStatementAccept tests Accept method on all statement types
func TestAllStatementAccept(t *testing.T) {
	statements := []struct {
		name string
		stmt Statement
	}{
		{
			"VarDeclStmt",
			&VarDeclStmt{Name: "x", Type_: &BasicType{Kind: IntType}},
		},
		{
			"AssignStmt",
			&AssignStmt{Target: &IdentifierExpr{Name: "x"}, Value: &LiteralExpr{Value: 1}},
		},
		{
			"IfStmt",
			&IfStmt{
				Condition: &LiteralExpr{Value: true, Type_: &BasicType{Kind: BoolType}},
				ThenStmt:  &BlockStmt{},
			},
		},
		{
			"ExprStmt",
			&ExprStmt{Expression: &LiteralExpr{Value: 1}},
		},
	}

	for _, tt := range statements {
		t.Run(tt.name, func(t *testing.T) {
			visitor := &MockVisitor{}
			tt.stmt.Accept(visitor)
			
			if len(visitor.visitedNodes) != 1 {
				t.Errorf("%s Accept should call visitor once", tt.name)
			}
		})
	}
}