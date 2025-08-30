package domain

import (
	"testing"
)

func TestBaseNode(t *testing.T) {
	node := BaseNode{
		Location: SourceRange{
			Start: SourcePosition{Line: 1, Column: 1},
			End:   SourcePosition{Line: 1, Column: 10},
		},
	}

	if node.GetLocation().Start.Line != 1 {
		t.Errorf("Expected line 1, got %d", node.GetLocation().Start.Line)
	}

	if node.GetLocation().Start.Column != 1 {
		t.Errorf("Expected column 1, got %d", node.GetLocation().Start.Column)
	}
}

func TestLiteralExpr(t *testing.T) {
	// Test int literal
	intLiteral := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    int64(42),
	}

	intType := NewIntType()
	intLiteral.SetType(intType)

	if intLiteral.GetType() != intType {
		t.Error("Int literal type not set correctly")
	}

	if intLiteral.Value.(int64) != 42 {
		t.Errorf("Expected value 42, got %v", intLiteral.Value)
	}

	// Test string literal
	strLiteral := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    "hello",
	}

	strType := NewStringType()
	strLiteral.SetType(strType)

	if strLiteral.GetType() != strType {
		t.Error("String literal type not set correctly")
	}
}

func TestIdentifierExpr(t *testing.T) {
	ident := &IdentifierExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Name:     "testVar",
	}

	intType := NewIntType()
	ident.SetType(intType)

	if ident.Name != "testVar" {
		t.Errorf("Expected name 'testVar', got '%s'", ident.Name)
	}

	if ident.GetType() != intType {
		t.Error("Identifier type not set correctly")
	}
}

func TestBinaryExpr(t *testing.T) {
	left := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    int64(5),
	}

	right := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    int64(3),
	}

	binaryExpr := &BinaryExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Left:     left,
		Operator: Add,
		Right:    right,
	}

	if binaryExpr.Left != left {
		t.Error("Left operand not set correctly")
	}

	if binaryExpr.Right != right {
		t.Error("Right operand not set correctly")
	}

	if binaryExpr.Operator != Add {
		t.Error("Operator not set correctly")
	}
}

func TestUnaryExpr(t *testing.T) {
	operand := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    int64(10),
	}

	unaryExpr := &UnaryExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Operator: Neg,
		Operand:  operand,
	}

	if unaryExpr.Operator != Neg {
		t.Error("Unary operator not set correctly")
	}

	if unaryExpr.Operand != operand {
		t.Error("Unary operand not set correctly")
	}
}

func TestCallExpr(t *testing.T) {
	function := &IdentifierExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Name:     "print",
	}

	args := []Expression{
		&LiteralExpr{
			BaseNode: BaseNode{Location: SourceRange{}},
			Value:    "test",
		},
	}

	callExpr := &CallExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Function: function,
		Args:     args,
	}

	if callExpr.Function != function {
		t.Error("Function not set correctly")
	}

	if len(callExpr.Args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(callExpr.Args))
	}
}

func TestVarDeclStmt(t *testing.T) {
	varDecl := &VarDeclStmt{
		BaseNode: BaseNode{Location: SourceRange{}},
		Name:     "testVar",
		Type_:    NewIntType(),
		Initializer: &LiteralExpr{
			BaseNode: BaseNode{Location: SourceRange{}},
			Value:    int64(42),
		},
	}

	if varDecl.Name != "testVar" {
		t.Errorf("Expected name 'testVar', got '%s'", varDecl.Name)
	}

	if varDecl.Type_.String() != "int" {
		t.Errorf("Expected type 'int', got '%s'", varDecl.Type_.String())
	}

	if varDecl.Initializer == nil {
		t.Error("Initializer should not be nil")
	}
}

func TestAssignStmt(t *testing.T) {
	target := &IdentifierExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Name:     "x",
	}

	value := &LiteralExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    int64(100),
	}

	assignStmt := &AssignStmt{
		BaseNode: BaseNode{Location: SourceRange{}},
		Target:   target,
		Value:    value,
	}

	if assignStmt.Target != target {
		t.Error("Target not set correctly")
	}

	if assignStmt.Value != value {
		t.Error("Value not set correctly")
	}
}

func TestIfStmt(t *testing.T) {
	condition := &BinaryExpr{
		BaseNode: BaseNode{Location: SourceRange{}},
		Left:     &IdentifierExpr{BaseNode: BaseNode{}, Name: "x"},
		Operator: Gt,
		Right:    &LiteralExpr{BaseNode: BaseNode{}, Value: int64(0)},
	}

	thenStmt := &ReturnStmt{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    &LiteralExpr{BaseNode: BaseNode{}, Value: int64(1)},
	}

	elseStmt := &ReturnStmt{
		BaseNode: BaseNode{Location: SourceRange{}},
		Value:    &LiteralExpr{BaseNode: BaseNode{}, Value: int64(0)},
	}

	ifStmt := &IfStmt{
		BaseNode:  BaseNode{Location: SourceRange{}},
		Condition: condition,
		ThenStmt:  thenStmt,
		ElseStmt:  elseStmt,
	}

	if ifStmt.Condition != condition {
		t.Error("If condition not set correctly")
	}

	if ifStmt.ThenStmt != thenStmt {
		t.Error("Then statement not set correctly")
	}

	if ifStmt.ElseStmt != elseStmt {
		t.Error("Else statement not set correctly")
	}
}

func TestFunctionDecl(t *testing.T) {
	params := []Parameter{
		{Name: "a", Type: NewIntType()},
		{Name: "b", Type: NewIntType()},
	}

	body := &BlockStmt{
		BaseNode:   BaseNode{Location: SourceRange{}},
		Statements: []Statement{},
	}

	funcDecl := &FunctionDecl{
		BaseNode:   BaseNode{Location: SourceRange{}},
		Name:       "add",
		Parameters: params,
		ReturnType: NewIntType(),
		Body:       body,
	}

	if funcDecl.Name != "add" {
		t.Errorf("Expected function name 'add', got '%s'", funcDecl.Name)
	}

	if len(funcDecl.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
	}

	if funcDecl.GetName() != "add" {
		t.Errorf("GetName() returned '%s', expected 'add'", funcDecl.GetName())
	}
}

func TestStructDecl(t *testing.T) {
	fields := []StructField{
		{Name: "name", Type: NewStringType()},
		{Name: "age", Type: NewIntType()},
	}

	structDecl := &StructDecl{
		BaseNode: BaseNode{Location: SourceRange{}},
		Name:     "Person",
		Fields:   fields,
	}

	if structDecl.Name != "Person" {
		t.Errorf("Expected struct name 'Person', got '%s'", structDecl.Name)
	}

	if len(structDecl.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structDecl.Fields))
	}

	if structDecl.GetName() != "Person" {
		t.Errorf("GetName() returned '%s', expected 'Person'", structDecl.GetName())
	}
}

func TestProgram(t *testing.T) {
	decls := []Declaration{
		&VarDeclStmt{
			BaseNode: BaseNode{Location: SourceRange{}},
			Name:     "globalVar",
			Type_:    NewIntType(),
		},
		&FunctionDecl{
			BaseNode: BaseNode{Location: SourceRange{}},
			Name:     "main",
			ReturnType: NewVoidType(),
		},
	}

	program := &Program{
		BaseNode:     BaseNode{Location: SourceRange{}},
		Declarations: decls,
	}

	if len(program.Declarations) != 2 {
		t.Errorf("Expected 2 declarations, got %d", len(program.Declarations))
	}

	if len(program.Declarations) > 0 {
		if funcDecl, ok := program.Declarations[1].(*FunctionDecl); ok {
			if funcDecl.Name != "main" {
				t.Errorf("Second declaration should be main function, got '%s'", funcDecl.Name)
			}
		}
	}
}

func TestBinaryOperatorString(t *testing.T) {
	testCases := []struct {
		operator BinaryOperator
		expected string
	}{
		{Add, "+"},
		{Sub, "-"},
		{Mul, "*"},
		{Div, "/"},
		{Mod, "%"},
		{Eq, "=="},
		{Ne, "!="},
		{Lt, "<"},
		{Le, "<="},
		{Gt, ">"},
		{Ge, ">="},
		{And, "&&"},
		{Or, "||"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.operator.String()
			if result != tc.expected {
				t.Errorf("Expected operator string '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestUnaryOperatorString(t *testing.T) {
	testCases := []struct {
		operator UnaryOperator
		expected string
	}{
		{Neg, "-"},
		{Not, "!"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.operator.String()
			if result != tc.expected {
				t.Errorf("Expected operator string '%s', got '%s'", tc.expected, result)
			}
		})
	}
}