package semantic

import (
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/infrastructure"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// TestAnalyzer_Creation tests analyzer creation
func TestAnalyzer_Creation(t *testing.T) {
	analyzer := NewAnalyzer()
	if analyzer == nil {
		t.Error("NewAnalyzer should return non-nil analyzer")
	}

	if analyzer.typeRegistry == nil {
		t.Error("Analyzer should have type registry initialized")
	}
}

// TestAnalyzer_SetTypeRegistry tests type registry setter
func TestAnalyzer_SetTypeRegistry(t *testing.T) {
	analyzer := NewAnalyzer()
	registry := domain.NewDefaultTypeRegistry()

	analyzer.SetTypeRegistry(registry)

	// We can't directly access private fields, but this tests the setter doesn't panic
	if analyzer.typeRegistry != registry {
		t.Error("SetTypeRegistry should set the type registry")
	}
}

// TestAnalyzer_SetSymbolTable tests symbol table setter
func TestAnalyzer_SetSymbolTable(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()

	analyzer.SetSymbolTable(symbolTable)

	// We can't directly access private fields, but this tests the setter doesn't panic
	if analyzer.symbolTable != symbolTable {
		t.Error("SetSymbolTable should set the symbol table")
	}
}

// TestAnalyzer_SetErrorReporter tests error reporter setter
func TestAnalyzer_SetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a simple mock error reporter
	mockReporter := &MockErrorReporter{}

	analyzer.SetErrorReporter(mockReporter)

	// Test doesn't panic and sets the reporter
	if analyzer.errorReporter != mockReporter {
		t.Error("SetErrorReporter should set the error reporter")
	}
}

// TestAnalyzer_AnalyzeEmptyProgram tests analyzing an empty program
func TestAnalyzer_AnalyzeEmptyProgram(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create empty program
	program := &domain.Program{
		Declarations: []domain.Declaration{},
	}

	err := analyzer.Analyze(program)
	if err != nil {
		t.Errorf("Analyzing empty program should not fail: %v", err)
	}
}

// TestAnalyzer_AnalyzeSimpleFunction tests analyzing a simple function
func TestAnalyzer_AnalyzeSimpleFunction(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create simple function: func main() -> void { }
	voidType := &domain.BasicType{Kind: domain.VoidType}

	functionDecl := &domain.FunctionDecl{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: voidType,
		Body: &domain.BlockStmt{
			BaseNode: domain.BaseNode{
				Location: domain.SourceRange{},
			},
			Statements: []domain.Statement{},
		},
	}

	program := &domain.Program{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Declarations: []domain.Declaration{functionDecl},
	}

	err := analyzer.Analyze(program)
	if err != nil {
		t.Errorf("Analyzing simple function should not fail: %v", err)
	}
}

// TestAnalyzer_BasicTypeValidation tests basic type validation
func TestAnalyzer_BasicTypeValidation(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Test that type registry has basic types
	registry := analyzer.typeRegistry

	intType := registry.GetBuiltinType(domain.IntType)
	if intType == nil {
		t.Error("Type registry should have int type")
	}

	floatType := registry.GetBuiltinType(domain.FloatType)
	if floatType == nil {
		t.Error("Type registry should have float type")
	}

	boolType := registry.GetBuiltinType(domain.BoolType)
	if boolType == nil {
		t.Error("Type registry should have bool type")
	}

	stringType := registry.GetBuiltinType(domain.StringType)
	if stringType == nil {
		t.Error("Type registry should have string type")
	}
}

// MockErrorReporter is a simple mock for testing
type MockErrorReporter struct {
	errors   []domain.CompilerError
	warnings []domain.CompilerError
}

func (m *MockErrorReporter) ReportError(err domain.CompilerError) {
	m.errors = append(m.errors, err)
}

func (m *MockErrorReporter) ReportWarning(warning domain.CompilerError) {
	m.warnings = append(m.warnings, warning)
}

func (m *MockErrorReporter) HasErrors() bool {
	return len(m.errors) > 0
}

func (m *MockErrorReporter) HasWarnings() bool {
	return len(m.warnings) > 0
}

func (m *MockErrorReporter) GetErrors() []domain.CompilerError {
	return m.errors
}

func (m *MockErrorReporter) GetWarnings() []domain.CompilerError {
	return m.warnings
}

func (m *MockErrorReporter) Clear() {
	m.errors = nil
	m.warnings = nil
}

// TestAnalyzer_VisitLiteralExpr tests literal expression analysis
func TestAnalyzer_VisitLiteralExpr(t *testing.T) {
	analyzer := NewAnalyzer()

	// Test integer literal
	intLiteral := &domain.LiteralExpr{
		Value: int64(42), // Must be int64 explicitly
	}
	err := analyzer.VisitLiteralExpr(intLiteral)
	if err != nil {
		t.Errorf("VisitLiteralExpr should not fail: %v", err)
	}

	if intLiteral.GetType().String() != "int" {
		t.Errorf("Expected int type, got %s", intLiteral.GetType().String())
	}

	// Test float literal
	floatLiteral := &domain.LiteralExpr{
		Value: 3.14, // This will be float64 by default
	}
	err = analyzer.VisitLiteralExpr(floatLiteral)
	if err != nil {
		t.Errorf("VisitLiteralExpr should not fail: %v", err)
	}

	if floatLiteral.GetType().String() != "float" {
		t.Errorf("Expected float type, got %s", floatLiteral.GetType().String())
	}

	// Test string literal
	strLiteral := &domain.LiteralExpr{
		Value: "hello",
	}
	err = analyzer.VisitLiteralExpr(strLiteral)
	if err != nil {
		t.Errorf("VisitLiteralExpr should not fail: %v", err)
	}

	if strLiteral.GetType().String() != "string" {
		t.Errorf("Expected string type, got %s", strLiteral.GetType().String())
	}

	// Test boolean literal
	boolLiteral := &domain.LiteralExpr{
		Value: true,
	}
	err = analyzer.VisitLiteralExpr(boolLiteral)
	if err != nil {
		t.Errorf("VisitLiteralExpr should not fail: %v", err)
	}

	if boolLiteral.GetType().String() != "bool" {
		t.Errorf("Expected bool type, got %s", boolLiteral.GetType().String())
	}
}

// TestAnalyzer_VisitBinaryExpr tests binary expression analysis
func TestAnalyzer_VisitBinaryExpr(t *testing.T) {
	analyzer := NewAnalyzer()

	// Test valid arithmetic expression: 1 + 2
	left := &domain.LiteralExpr{Value: int64(1)}
	right := &domain.LiteralExpr{Value: int64(2)}

	err := analyzer.VisitLiteralExpr(left)
	if err != nil {
		t.Errorf("Setting up left operand failed: %v", err)
	}

	err = analyzer.VisitLiteralExpr(right)
	if err != nil {
		t.Errorf("Setting up right operand failed: %v", err)
	}

	binaryExpr := &domain.BinaryExpr{
		Left:     left,
		Operator: domain.Add,
		Right:    right,
	}

	err = analyzer.VisitBinaryExpr(binaryExpr)
	if err != nil {
		t.Errorf("VisitBinaryExpr should not fail: %v", err)
	}

	if binaryExpr.GetType().String() != "int" {
		t.Errorf("Expected int result type, got %s", binaryExpr.GetType().String())
	}

	// Test comparison expression: 1 == 2
	compExpr := &domain.BinaryExpr{
		Left:     &domain.LiteralExpr{Value: int64(1)},
		Operator: domain.Eq,
		Right:    &domain.LiteralExpr{Value: int64(2)},
	}

	err = analyzer.VisitBinaryExpr(compExpr)
	if err != nil {
		t.Errorf("VisitBinaryExpr should not fail: %v", err)
	}

	if compExpr.GetType().String() != "bool" {
		t.Errorf("Expected bool result type, got %s", compExpr.GetType().String())
	}
}

// TestAnalyzer_VisitUnaryExpr tests unary expression analysis
func TestAnalyzer_VisitUnaryExpr(t *testing.T) {
	analyzer := NewAnalyzer()

	// Test negation: -5
	operand := &domain.LiteralExpr{Value: int64(5)}
	unaryExpr := &domain.UnaryExpr{
		Operator: domain.Neg,
		Operand:  operand,
	}

	err := analyzer.VisitUnaryExpr(unaryExpr)
	if err != nil {
		t.Errorf("VisitUnaryExpr should not fail: %v", err)
	}

	if unaryExpr.GetType().String() != "int" {
		t.Errorf("Expected int type, got %s", unaryExpr.GetType().String())
	}

	// Test logical not: !true
	boolOperand := &domain.LiteralExpr{Value: true}
	notExpr := &domain.UnaryExpr{
		Operator: domain.Not,
		Operand:  boolOperand,
	}

	err = analyzer.VisitUnaryExpr(notExpr)
	if err != nil {
		t.Errorf("VisitUnaryExpr should not fail: %v", err)
	}

	if notExpr.GetType().String() != "bool" {
		t.Errorf("Expected bool type, got %s", notExpr.GetType().String())
	}
}

// TestAnalyzer_VisitIdentifierExpr tests identifier expression analysis
func TestAnalyzer_VisitIdentifierExpr(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	errReporter := &MockErrorReporter{}
	analyzer.SetErrorReporter(errReporter)

	// Declare a variable first
	testType := &domain.BasicType{Kind: domain.IntType}
	_, err := symbolTable.DeclareSymbol(
		"x",
		testType,
		interfaces.VariableSymbol,
		domain.SourceRange{},
	)
	if err != nil {
		t.Fatalf("Failed to declare test symbol: %v", err)
	}

	// Now test identifier lookup
	identExpr := &domain.IdentifierExpr{
		Name: "x",
	}

	err = analyzer.VisitIdentifierExpr(identExpr)
	if err != nil {
		t.Errorf("VisitIdentifierExpr should not fail: %v", err)
	}

	if identExpr.GetType().String() != "int" {
		t.Errorf("Expected int type, got %s", identExpr.GetType().String())
	}

	// Test undefined identifier
	undefinedExpr := &domain.IdentifierExpr{
		Name: "undefined_var",
	}

	err = analyzer.VisitIdentifierExpr(undefinedExpr)
	if err != nil {
		t.Errorf("VisitIdentifierExpr should handle undefined identifiers gracefully: %v", err)
	}

	// Should have error reported
	if len(errReporter.errors) == 0 {
		t.Error("Expected error for undefined identifier")
	}
}

// TestAnalyzer_VisitCallExpr tests function call expression analysis
func TestAnalyzer_VisitCallExpr(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create function type for "testFunc"
	paramType := &domain.BasicType{Kind: domain.IntType}
	returnType := &domain.BasicType{Kind: domain.VoidType}

	funcType := &domain.FunctionType{
		ParameterTypes: []domain.Type{paramType},
		ReturnType:     returnType,
	}

	_, err := symbolTable.DeclareSymbol(
		"testFunc",
		funcType,
		interfaces.FunctionSymbol,
		domain.SourceRange{},
	)
	if err != nil {
		t.Fatalf("Failed to declare function symbol: %v", err)
	}

	// Test function call with correct args
	callExpr := &domain.CallExpr{
		Function: &domain.IdentifierExpr{Name: "testFunc"},
		Args:     []domain.Expression{&domain.LiteralExpr{Value: int64(42)}},
	}

	err = analyzer.VisitCallExpr(callExpr)
	if err != nil {
		t.Errorf("VisitCallExpr should not fail: %v", err)
	}

	if callExpr.GetType().String() != "void" {
		t.Errorf("Expected void return type, got %s", callExpr.GetType().String())
	}
}

// TestAnalyzer_VisitIfStmt tests if statement analysis
func TestAnalyzer_VisitIfStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	errReporter := &MockErrorReporter{}

	analyzer.SetSymbolTable(symbolTable)
	analyzer.SetErrorReporter(errReporter)

	// Test valid if statement
	thenBlock := &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Statements: []domain.Statement{},
	}

	ifStmt := &domain.IfStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Condition: &domain.LiteralExpr{Value: true},
		ThenStmt:  thenBlock,
	}

	err := analyzer.VisitIfStmt(ifStmt)
	if err != nil {
		t.Errorf("VisitIfStmt should not fail: %v", err)
	}

	// Test if with else
	elseBlock := &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Statements: []domain.Statement{},
	}

	ifElseStmt := &domain.IfStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Condition: &domain.LiteralExpr{Value: false},
		ThenStmt:  thenBlock,
		ElseStmt:  elseBlock,
	}

	err = analyzer.VisitIfStmt(ifElseStmt)
	if err != nil {
		t.Errorf("VisitIfStmt with else should not fail: %v", err)
	}
}

// TestAnalyzer_VisitWhileStmt tests while statement analysis
func TestAnalyzer_VisitWhileStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	bodyBlock := &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Statements: []domain.Statement{},
	}

	whileStmt := &domain.WhileStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Condition: &domain.LiteralExpr{Value: true},
		Body:      bodyBlock,
	}

	err := analyzer.VisitWhileStmt(whileStmt)
	if err != nil {
		t.Errorf("VisitWhileStmt should not fail: %v", err)
	}
}

// TestAnalyzer_VisitForStmt tests for statement analysis
func TestAnalyzer_VisitForStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	bodyBlock := &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Statements: []domain.Statement{},
	}

	// Test for with init, condition, and update
	initStmt := &domain.VarDeclStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Name:       "i",
		Type_:      &domain.BasicType{Kind: domain.IntType},
		Initializer: &domain.LiteralExpr{Value: int64(0)},
	}

	condition := &domain.BinaryExpr{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Left:     &domain.IdentifierExpr{Name: "i"},
		Operator: domain.Lt,
		Right:    &domain.LiteralExpr{Value: int64(10)},
	}

	updateStmt := &domain.AssignStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Target: &domain.IdentifierExpr{Name: "i"},
		Value: &domain.BinaryExpr{
			BaseNode: domain.BaseNode{
				Location: domain.SourceRange{},
			},
			Left:     &domain.IdentifierExpr{Name: "i"},
			Operator: domain.Add,
			Right:    &domain.LiteralExpr{Value: int64(1)},
		},
	}

	forStmt := &domain.ForStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Init:      initStmt,
		Condition: condition,
		Update:    updateStmt,
		Body:      bodyBlock,
	}

	err := analyzer.VisitForStmt(forStmt)
	if err != nil {
		t.Errorf("VisitForStmt should not fail: %v", err)
	}
}

// TestAnalyzer_VisitReturnStmt tests return statement analysis
func TestAnalyzer_VisitReturnStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	errReporter := &MockErrorReporter{}
	analyzer.SetErrorReporter(errReporter)

	// Test return without value in void function
	returnStmt := &domain.ReturnStmt{
		Value: nil,
	}

	err := analyzer.VisitReturnStmt(returnStmt)
	if err != nil {
		t.Errorf("VisitReturnStmt should not fail even outside function context: %v", err)
	}

	// Check that error was reported for being outside function
	if len(errReporter.errors) == 0 {
		t.Error("Expected error when return statement is used outside function")
	}
}

// TestAnalyzer_VisitVarDeclStmt tests variable declaration statement analysis
func TestAnalyzer_VisitVarDeclStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)
	errReporter := &MockErrorReporter{}
	analyzer.SetErrorReporter(errReporter)

	// Test valid variable declaration
	varDeclStmt := &domain.VarDeclStmt{
		Name:       "testVar",
		Type_:      &domain.BasicType{Kind: domain.IntType},
		Initializer: &domain.LiteralExpr{Value: int64(42)},
	}

	err := analyzer.VisitVarDeclStmt(varDeclStmt)
	if err != nil {
		t.Errorf("VisitVarDeclStmt should not fail: %v", err)
	}

	// Test variable declaration without initializer
	varDeclStmt2 := &domain.VarDeclStmt{
		Name:       "testVar2",
		Type_:      &domain.BasicType{Kind: domain.FloatType},
		Initializer: nil,
	}

	err = analyzer.VisitVarDeclStmt(varDeclStmt2)
	if err != nil {
		t.Errorf("VisitVarDeclStmt without initializer should not fail: %v", err)
	}
}

// TestAnalyzer_VisitAssignStmt tests assignment statement analysis
func TestAnalyzer_VisitAssignStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Declare a variable first
	testType := &domain.BasicType{Kind: domain.IntType}
	symbolTable.DeclareSymbol("target", testType, interfaces.VariableSymbol, domain.SourceRange{})

	// Test valid assignment
	assignStmt := &domain.AssignStmt{
		Target: &domain.IdentifierExpr{Name: "target"},
		Value:  &domain.LiteralExpr{Value: int64(100)},
	}

	err := analyzer.VisitAssignStmt(assignStmt)
	if err != nil {
		t.Errorf("VisitAssignStmt should not fail: %v", err)
	}
}

// TestAnalyzer_VisitIndexExpr tests array index expression analysis
func TestAnalyzer_VisitIndexExpr(t *testing.T) {
	analyzer := NewAnalyzer()
	errReporter := &MockErrorReporter{}
	analyzer.SetErrorReporter(errReporter)

	// Test array indexing
	elementType := &domain.BasicType{Kind: domain.IntType}
	arrayType := &domain.ArrayType{
		ElementType: elementType,
		Size:        10,
	}

	indexExpr := &domain.IndexExpr{
		Object: &MockArrayObject{Type: arrayType},
		Index:  &domain.LiteralExpr{Value: int64(5)},
	}

	err := analyzer.VisitIndexExpr(indexExpr)
	if err != nil {
		t.Errorf("VisitIndexExpr should not fail: %v", err)
	}

	if indexExpr.GetType().String() != "int" {
		t.Errorf("Expected int element type, got %s", indexExpr.GetType().String())
	}
}

// TestAnalyzer_VisitMemberExpr tests member access expression analysis
func TestAnalyzer_VisitMemberExpr(t *testing.T) {
	analyzer := NewAnalyzer()
	errReporter := &MockErrorReporter{}
	analyzer.SetErrorReporter(errReporter)

	// Test struct member access
	structType := &domain.StructType{
		Name: "TestStruct",
		Fields: map[string]domain.Type{
			"field": &domain.BasicType{Kind: domain.IntType},
		},
		Order: []string{"field"},
	}

	memberExpr := &domain.MemberExpr{
		Object: &MockStructObject{Type: structType},
		Member: "field",
	}

	err := analyzer.VisitMemberExpr(memberExpr)
	if err != nil {
		t.Errorf("VisitMemberExpr should not fail: %v", err)
	}

	if memberExpr.GetType().String() != "int" {
		t.Errorf("Expected int field type, got %s", memberExpr.GetType().String())
	}
}

// TestAnalyzer_VisitBlockStmt tests block statement analysis
func TestAnalyzer_VisitBlockStmt(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	errReporter := &MockErrorReporter{}

	analyzer.SetSymbolTable(symbolTable)
	analyzer.SetErrorReporter(errReporter)

	blockStmt := &domain.BlockStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Statements: []domain.Statement{
			&domain.VarDeclStmt{
				BaseNode: domain.BaseNode{
					Location: domain.SourceRange{},
				},
				Name:       "blockVar",
				Type_:      &domain.BasicType{Kind: domain.IntType},
				Initializer: &domain.LiteralExpr{Value: int64(1)},
			},
			&domain.ExprStmt{
				BaseNode: domain.BaseNode{
					Location: domain.SourceRange{},
				},
				Expression: &domain.LiteralExpr{Value: int64(2)},
			},
		},
	}

	err := analyzer.VisitBlockStmt(blockStmt)
	if err != nil {
		t.Errorf("VisitBlockStmt should not fail: %v", err)
	}
}

// TestAnalyzer_VisitExprStmt tests expression statement analysis
func TestAnalyzer_VisitExprStmt(t *testing.T) {
	analyzer := NewAnalyzer()

	exprStmt := &domain.ExprStmt{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Expression: &domain.LiteralExpr{Value: int64(42)},
	}

	err := analyzer.VisitExprStmt(exprStmt)
	if err != nil {
		t.Errorf("VisitExprStmt should not fail: %v", err)
	}
}

// TestAnalyzer_VisitProgram tests program analysis
func TestAnalyzer_VisitProgram(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	program := &domain.Program{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Declarations: []domain.Declaration{
			&domain.FunctionDecl{
				BaseNode: domain.BaseNode{
					Location: domain.SourceRange{},
				},
				Name:       "main",
				Parameters: []domain.Parameter{},
				ReturnType: &domain.BasicType{Kind: domain.VoidType},
				Body: &domain.BlockStmt{
					BaseNode: domain.BaseNode{
						Location: domain.SourceRange{},
					},
					Statements: []domain.Statement{},
				},
			},
		},
	}

	err := analyzer.VisitProgram(program)
	if err != nil {
		t.Errorf("VisitProgram should not fail: %v", err)
	}
}

// TestAnalyzer_VisitStructDecl tests struct declaration analysis
func TestAnalyzer_VisitStructDecl(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	structDecl := &domain.StructDecl{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Name: "TestStruct",
		Fields: []domain.StructField{
			{Name: "field1", Type: &domain.BasicType{Kind: domain.IntType}},
			{Name: "field2", Type: &domain.BasicType{Kind: domain.FloatType}},
		},
	}

	err := analyzer.VisitStructDecl(structDecl)
	if err != nil {
		t.Errorf("VisitStructDecl should not fail: %v", err)
	}
}

// TestAnalyzer_ErrorReporting tests comprehensive error reporting scenarios
func TestAnalyzer_ErrorReporting(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	errReporter := &MockErrorReporter{}

	analyzer.SetSymbolTable(symbolTable)
	analyzer.SetErrorReporter(errReporter)

	// Test binary operator error
	intLiteral := &domain.LiteralExpr{Value: int64(1)}
	stringLiteral := &domain.LiteralExpr{Value: "hello"}

	err := analyzer.VisitLiteralExpr(intLiteral)
	if err != nil {
		t.Errorf("Literal analysis failed: %v", err)
	}

	err = analyzer.VisitLiteralExpr(stringLiteral)
	if err != nil {
		t.Errorf("Literal analysis failed: %v", err)
	}

	invalidBinaryExpr := &domain.BinaryExpr{
		Left:     intLiteral,
		Operator: domain.Add,
		Right:    stringLiteral,
	}

	err = analyzer.VisitBinaryExpr(invalidBinaryExpr)
	if err != nil {
		t.Errorf("Binary expr analysis failed: %v", err)
	}

	// Should have reported error for invalid operation
	if len(errReporter.errors) == 0 {
		t.Error("Expected error for invalid binary operation")
	}

	// Reset errors for next test
	errReporter.Clear()

	// Test type assignment error
	symbolTable.DeclareSymbol("testVar", &domain.BasicType{Kind: domain.IntType}, interfaces.VariableSymbol, domain.SourceRange{})

	invalidAssignment := &domain.AssignStmt{
		Target: &domain.IdentifierExpr{Name: "testVar"},
		Value:  stringLiteral,
	}

	err = analyzer.VisitAssignStmt(invalidAssignment)
	if err != nil {
		t.Errorf("Assignment analysis failed: %v", err)
	}

	// Should have reported type error
	if len(errReporter.errors) == 0 {
		t.Error("Expected type error for invalid assignment")
	}
}

// Mock objects for testing complex expressions

type MockArrayObject struct {
	Type domain.Type
}

func (m *MockArrayObject) GetLocation() domain.SourceRange { return domain.SourceRange{} }
func (m *MockArrayObject) Accept(domain.Visitor) error      { return nil }
func (m *MockArrayObject) GetType() domain.Type            { return m.Type }
func (m *MockArrayObject) SetType(domain.Type)             {}

type MockStructObject struct {
	Type domain.Type
}

func (m *MockStructObject) GetLocation() domain.SourceRange { return domain.SourceRange{} }
func (m *MockStructObject) Accept(domain.Visitor) error      { return nil }
func (m *MockStructObject) GetType() domain.Type            { return m.Type }
func (m *MockStructObject) SetType(domain.Type)             {}

// TestAnalyzer_ComplexExpression tests complex expression structures
func TestAnalyzer_ComplexExpression(t *testing.T) {
	// Test that we can create complex expression structures without panic
	// Create: (a + b) * (c - d) expression tree

	a := &domain.IdentifierExpr{Name: "a"}
	b := &domain.IdentifierExpr{Name: "b"}
	c := &domain.IdentifierExpr{Name: "c"}
	d := &domain.IdentifierExpr{Name: "d"}

	addExpr := &domain.BinaryExpr{
		Left:     a,
		Operator: domain.Add,
		Right:    b,
	}

	subExpr := &domain.BinaryExpr{
		Left:     c,
		Operator: domain.Sub,
		Right:    d,
	}

	mulExpr := &domain.BinaryExpr{
		Left:     addExpr,
		Operator: domain.Mul,
		Right:    subExpr,
	}

	// Verify the structure is correct
	if mulExpr.Left != addExpr {
		t.Error("Left operand should be the add expression")
	}

	if mulExpr.Right != subExpr {
		t.Error("Right operand should be the sub expression")
	}

	if addExpr.Left != a || addExpr.Right != b {
		t.Error("Add expression operands incorrect")
	}

	if subExpr.Left != c || subExpr.Right != d {
		t.Error("Sub expression operands incorrect")
	}
}

// TestAnalyzer_ScopeHandling tests scope management
func TestAnalyzer_ScopeHandling(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Test nested scopes
	outerVarDecl := &domain.VarDeclStmt{
		Name:       "outer",
		Type_:      &domain.BasicType{Kind: domain.IntType},
		Initializer: &domain.LiteralExpr{Value: int64(1)},
	}

	blockStmt := &domain.BlockStmt{
		Statements: []domain.Statement{
			&domain.VarDeclStmt{
				Name:       "inner",
				Type_:      &domain.BasicType{Kind: domain.IntType},
				Initializer: &domain.LiteralExpr{Value: int64(2)},
			},
		},
	}

	outsideBlock := &domain.ExprStmt{
		Expression: &domain.IdentifierExpr{Name: "inner"}, // This should cause an error
	}

	// Analyze outer variable
	err := analyzer.VisitVarDeclStmt(outerVarDecl)
	if err != nil {
		t.Errorf("Outer variable declaration failed: %v", err)
	}

	// Analyze block (creates nested scope)
	err = analyzer.VisitBlockStmt(blockStmt)
	if err != nil {
		t.Errorf("Block statement analysis failed: %v", err)
	}

	// Try to use inner variable outside block
	err = analyzer.VisitExprStmt(outsideBlock)
	if err != nil {
		t.Errorf("Expression statement analysis failed: %v", err)
	}
}

// TestAnalyzer_FunctionComplexity tests function declaration and call complexity
func TestAnalyzer_FunctionComplexity(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create a complex function with parameters
	funcDecl := &domain.FunctionDecl{
		Name: "complexFunc",
		Parameters: []domain.Parameter{
			{Name: "x", Type: &domain.BasicType{Kind: domain.IntType}},
			{Name: "y", Type: &domain.BasicType{Kind: domain.FloatType}},
		},
		ReturnType: &domain.BasicType{Kind: domain.IntType},
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{
				&domain.ReturnStmt{
					Value: &domain.BinaryExpr{
						Left:     &domain.IdentifierExpr{Name: "x"},
						Operator: domain.Add,
						Right:    &domain.LiteralExpr{Value: int64(1)},
					},
				},
			},
		},
	}

	err := analyzer.VisitFunctionDecl(funcDecl)
	if err != nil {
		t.Errorf("Complex function analysis failed: %v", err)
	}
}

// TestAnalyzer_TypeChecking tests various type checking scenarios
func TestAnalyzer_TypeChecking(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	errReporter := &MockErrorReporter{}

	analyzer.SetSymbolTable(symbolTable)
	analyzer.SetErrorReporter(errReporter)

	testType := &domain.BasicType{Kind: domain.BoolType}
	symbolTable.DeclareSymbol("boolVar", testType, interfaces.VariableSymbol, domain.SourceRange{})

	// Test if condition type checking
	ifStmt := &domain.IfStmt{
		Condition: &domain.IdentifierExpr{Name: "boolVar"},
		ThenStmt:  &domain.BlockStmt{Statements: []domain.Statement{}},
	}

	err := analyzer.VisitIfStmt(ifStmt)
	if err != nil {
		t.Errorf("If statement analysis failed: %v", err)
	}

	// Should not have any errors since boolVar is boolean
	if len(errReporter.errors) > 0 {
		t.Errorf("Expected no errors, got: %v", errReporter.errors)
	}
}

// TestAnalyzer_InitializeBuiltinFunctions tests builtin function initialization
func TestAnalyzer_InitializeBuiltinFunctions(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	err := analyzer.initializeBuiltinFunctions()
	if err != nil {
		t.Errorf("initializeBuiltinFunctions should not fail: %v", err)
	}

	// Check if 'print' function was added
	printSymbol, found := symbolTable.LookupSymbol("print")
	if !found {
		t.Error("print function should be available as builtin")
	}

	if printSymbol.Kind != interfaces.FunctionSymbol {
		t.Error("print should be a function symbol")
	}
}
