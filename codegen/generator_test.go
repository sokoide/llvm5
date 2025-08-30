package codegen

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

// TestNewGenerator tests the constructor
func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()
	
	if generator == nil {
		t.Error("NewGenerator should return a non-nil generator")
	}
	
	if generator.labelCounter != 0 {
		t.Error("New generator should have labelCounter initialized to 0")
	}
	
	if generator.parameters == nil {
		t.Error("New generator should have parameters map initialized")
	}
}

// TestEmitMethods tests the emit helper methods
func TestEmitMethods(t *testing.T) {
	generator := NewGenerator()
	
	// Test emit method
	generator.emit("test instruction %s", "arg1")
	output := generator.output.String()
	if !strings.Contains(output, "test instruction arg1") {
		t.Error("emit should format and write the instruction")
	}
	
	// Test emitRaw method
	generator.output.Reset()
	generator.emitRaw("raw text")
	output = generator.output.String()
	if output != "raw text" {
		t.Error("emitRaw should write raw text without formatting")
	}
}

// TestNewLabel tests label generation
func TestNewLabel(t *testing.T) {
	generator := NewGenerator()
	
	label1 := generator.newLabel("test")
	label2 := generator.newLabel("test")
	
	if label1 == label2 {
		t.Error("newLabel should generate unique labels")
	}
	
	if !strings.HasPrefix(label1, "test") {
		t.Error("newLabel should use the provided prefix")
	}
	
	if generator.labelCounter != 2 {
		t.Error("newLabel should increment the counter")
	}
}

// TestGetLLVMType tests LLVM type conversion
func TestGetLLVMType(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		domainType   domain.Type
		expectedLLVM string
	}{
		{domain.NewIntType(), "i32"},
		{domain.NewBoolType(), "i1"},
		{domain.NewStringType(), "i8*"},
		{domain.NewVoidType(), "void"},
	}
	
	for _, test := range tests {
		result := generator.getLLVMType(test.domainType)
		if result != test.expectedLLVM {
			t.Errorf("getLLVMType(%s) = %s, expected %s", 
				test.domainType.String(), result, test.expectedLLVM)
		}
	}
}

// TestGetTypeAlign tests type alignment calculation
func TestGetTypeAlign(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		domainType    domain.Type
		expectedAlign int
	}{
		{domain.NewIntType(), 4},
		{domain.NewBoolType(), 1},
		{domain.NewStringType(), 8},
	}
	
	for _, test := range tests {
		result := generator.getTypeAlign(test.domainType)
		if result != test.expectedAlign {
			t.Errorf("getTypeAlign(%s) = %d, expected %d", 
				test.domainType.String(), result, test.expectedAlign)
		}
	}
}

// TestParseFormatString tests format string parsing
func TestParseFormatString(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		format   string
		expected []string
		hasError bool
	}{
		{"Hello %d", []string{"int"}, false},
		{"Value: %f", []string{"float"}, false},
		{"Name: %s", []string{"string"}, false},
		{"%d + %d = %d", []string{"int", "int", "int"}, false},
		{"Incomplete %", nil, true},
		{"Invalid %z", nil, true},
		{"Escaped %%d", []string{}, false},
	}
	
	for _, test := range tests {
		result, err := generator.parseFormatString(test.format)
		
		if test.hasError {
			if err == nil {
				t.Errorf("parseFormatString(%q) should return error", test.format)
			}
		} else {
			if err != nil {
				t.Errorf("parseFormatString(%q) returned error: %v", test.format, err)
				continue
			}
			
			if len(result) != len(test.expected) {
				t.Errorf("parseFormatString(%q) returned %d types, expected %d", 
					test.format, len(result), len(test.expected))
				continue
			}
			
			for i, expectedType := range test.expected {
				if result[i] != expectedType {
					t.Errorf("parseFormatString(%q)[%d] = %s, expected %s", 
						test.format, i, result[i], expectedType)
				}
			}
		}
	}
}


// TestGenerateWithoutBackend tests generation without LLVM backend
func TestGenerateWithoutBackend(t *testing.T) {
	generator := NewGenerator()
	// Don't set backend
	
	program := &domain.Program{
		Declarations: []domain.Declaration{
			&domain.FunctionDecl{
				Name:       "main",
				Parameters: []domain.Parameter{},
				ReturnType: domain.NewIntType(),
				Body:       &domain.BlockStmt{Statements: []domain.Statement{}},
			},
		},
	}
	
	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Generate should work without backend: %v", err)
	}
	
	if !strings.Contains(result, "define i32 @main()") {
		t.Error("Generated code should contain main function")
	}
}

// TestVisitBinaryExpr tests binary expression code generation
func TestVisitBinaryExpr(t *testing.T) {
	tests := []struct {
		name     string
		operator domain.BinaryOperator
		left     interface{}
		right    interface{}
		expected string
	}{
		{"add_int", domain.Add, int64(5), int64(3), "add i32"},
		{"sub_int", domain.Sub, int64(10), int64(2), "sub i32"},
		{"mul_int", domain.Mul, int64(4), int64(3), "mul i32"},
		{"div_int", domain.Div, int64(12), int64(3), "sdiv i32"},
		{"eq_int", domain.Eq, int64(5), int64(5), "icmp eq i32"},
		{"ne_int", domain.Ne, int64(5), int64(3), "icmp ne i32"},
		{"lt_int", domain.Lt, int64(3), int64(5), "icmp slt i32"},
		{"gt_int", domain.Gt, int64(5), int64(3), "icmp sgt i32"},
		{"le_int", domain.Le, int64(3), int64(5), "icmp sle i32"},
		{"ge_int", domain.Ge, int64(5), int64(3), "icmp sge i32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			
			left := &domain.LiteralExpr{
				Type_: domain.NewIntType(),
				Value: tt.left,
			}
			left.SetType(domain.NewIntType())
			
			right := &domain.LiteralExpr{
				Type_: domain.NewIntType(),
				Value: tt.right,
			}
			right.SetType(domain.NewIntType())
			
			binaryExpr := &domain.BinaryExpr{
				Left:     left,
				Operator: tt.operator,
				Right:    right,
			}
			binaryExpr.SetType(domain.NewIntType())
			
			err := binaryExpr.Accept(generator)
			if err != nil {
				t.Fatalf("VisitBinaryExpr failed: %v", err)
			}
			
			output := generator.output.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %q in output, got: %s", tt.expected, output)
			}
		})
	}
}

// TestVisitUnaryExpr tests unary expression code generation
func TestVisitUnaryExpr(t *testing.T) {
	tests := []struct {
		name     string
		operator domain.UnaryOperator
		operand  interface{}
		type_    domain.Type
		expected string
	}{
		{"neg_int", domain.Neg, int64(5), domain.NewIntType(), "sub i32 0"},
		{"not_bool", domain.Not, true, domain.NewBoolType(), "icmp eq i1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			
			operand := &domain.LiteralExpr{
				Type_: tt.type_,
				Value: tt.operand,
			}
			operand.SetType(tt.type_)
			
			unaryExpr := &domain.UnaryExpr{
				Operator: tt.operator,
				Operand:  operand,
			}
			unaryExpr.SetType(tt.type_)
			
			err := unaryExpr.Accept(generator)
			if err != nil {
				t.Fatalf("VisitUnaryExpr failed: %v", err)
			}
			
			output := generator.output.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %q in output, got: %s", tt.expected, output)
			}
		})
	}
}

// TestVisitLiteralExpr tests literal expression code generation
func TestVisitLiteralExpr(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		type_    domain.Type
		expected string
	}{
		{"int_literal", int64(42), domain.NewIntType(), "42"},
		{"string_literal", "hello", domain.NewStringType(), "hello"},
		{"bool_literal", true, domain.NewBoolType(), "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			
			literal := &domain.LiteralExpr{
				Type_: tt.type_,
				Value: tt.value,
			}
			literal.SetType(tt.type_)
			
			err := literal.Accept(generator)
			if err != nil {
				t.Fatalf("VisitLiteralExpr failed: %v", err)
			}
			
			if tt.name == "string_literal" {
				output := generator.output.String()
				if !strings.Contains(output, "hello") {
					t.Errorf("Expected string literal processing")
				}
			}
		})
	}
}

// TestVisitIdentifierExpr tests identifier expression code generation
func TestVisitIdentifierExpr(t *testing.T) {
	generator := NewGenerator()
	
	// Set up a parameter so the identifier resolution works
	generator.parameters["x"] = true
	
	identifier := &domain.IdentifierExpr{
		Name:  "x",
		Type_: domain.NewIntType(),
	}
	identifier.SetType(domain.NewIntType())
	
	err := identifier.Accept(generator)
	if err != nil {
		t.Fatalf("VisitIdentifierExpr failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "load i32") {
		t.Error("Expected load instruction for identifier")
	}
	if !strings.Contains(output, "%x.addr") {
		t.Error("Expected parameter addressing")
	}
}

// TestVisitReturnStmt tests return statement code generation
func TestVisitReturnStmt(t *testing.T) {
	tests := []struct {
		name      string
		hasValue  bool
		value     interface{}
		valueType domain.Type
		expected  string
	}{
		{"return_void", false, nil, nil, "ret void"},
		{"return_int", true, int64(42), domain.NewIntType(), "ret i32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			
			var returnStmt *domain.ReturnStmt
			if tt.hasValue {
				value := &domain.LiteralExpr{
					Type_: tt.valueType,
					Value: tt.value,
				}
				value.SetType(tt.valueType)
				returnStmt = &domain.ReturnStmt{Value: value}
			} else {
				returnStmt = &domain.ReturnStmt{Value: nil}
			}
			
			err := returnStmt.Accept(generator)
			if err != nil {
				t.Fatalf("VisitReturnStmt failed: %v", err)
			}
			
			output := generator.output.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %q in output, got: %s", tt.expected, output)
			}
		})
	}
}

// TestVisitVarDeclStmt tests variable declaration code generation
func TestVisitVarDeclStmt(t *testing.T) {
	generator := NewGenerator()
	
	initializer := &domain.LiteralExpr{
		Type_: domain.NewIntType(),
		Value: int64(42),
	}
	initializer.SetType(domain.NewIntType())
	
	varDecl := &domain.VarDeclStmt{
		Name:        "x",
		Type_:       domain.NewIntType(),
		Initializer: initializer,
	}
	
	err := varDecl.Accept(generator)
	if err != nil {
		t.Fatalf("VisitVarDeclStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "%x = alloca i32") {
		t.Error("Expected variable allocation")
	}
	if !strings.Contains(output, "store i32") {
		t.Error("Expected variable initialization")
	}
}

// TestVisitAssignStmt tests assignment statement code generation
func TestVisitAssignStmt(t *testing.T) {
	generator := NewGenerator()
	
	target := &domain.IdentifierExpr{
		Name:  "x",
		Type_: domain.NewIntType(),
	}
	target.SetType(domain.NewIntType())
	
	value := &domain.LiteralExpr{
		Type_: domain.NewIntType(),
		Value: int64(42),
	}
	value.SetType(domain.NewIntType())
	
	assignStmt := &domain.AssignStmt{
		Target: target,
		Value:  value,
	}
	
	err := assignStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitAssignStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "store i32") {
		t.Error("Expected store instruction for assignment")
	}
}

// TestVisitIfStmt tests if statement code generation
func TestVisitIfStmt(t *testing.T) {
	generator := NewGenerator()
	generator.indentLevel = 1
	
	condition := &domain.LiteralExpr{
		Type_: domain.NewBoolType(),
		Value: true,
	}
	condition.SetType(domain.NewBoolType())
	
	thenStmt := &domain.ReturnStmt{
		Value: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(1),
		},
	}
	thenStmt.Value.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	elseStmt := &domain.ReturnStmt{
		Value: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(0),
		},
	}
	elseStmt.Value.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	ifStmt := &domain.IfStmt{
		Condition: condition,
		ThenStmt:  thenStmt,
		ElseStmt:  elseStmt,
	}
	
	err := ifStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitIfStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "br i1") {
		t.Error("Expected conditional branch")
	}
	if !strings.Contains(output, "if.then") {
		t.Error("Expected then label")
	}
	if !strings.Contains(output, "if.else") {
		t.Error("Expected else label")
	}
}

// TestVisitWhileStmt tests while loop code generation
func TestVisitWhileStmt(t *testing.T) {
	generator := NewGenerator()
	generator.indentLevel = 1
	
	condition := &domain.LiteralExpr{
		Type_: domain.NewBoolType(),
		Value: true,
	}
	condition.SetType(domain.NewBoolType())
	
	body := &domain.BlockStmt{
		Statements: []domain.Statement{},
	}
	
	whileStmt := &domain.WhileStmt{
		Condition: condition,
		Body:      body,
	}
	
	err := whileStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitWhileStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "while.cond") {
		t.Error("Expected while condition label")
	}
	if !strings.Contains(output, "while.body") {
		t.Error("Expected while body label")
	}
	if !strings.Contains(output, "while.end") {
		t.Error("Expected while end label")
	}
}

// TestVisitForStmt tests for loop code generation
func TestVisitForStmt(t *testing.T) {
	generator := NewGenerator()
	generator.indentLevel = 1
	
	// for (int i = 0; i < 10; i++)
	init := &domain.VarDeclStmt{
		Name:  "i",
		Type_: domain.NewIntType(),
		Initializer: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(0),
		},
	}
	init.Initializer.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	condition := &domain.BinaryExpr{
		Left: &domain.IdentifierExpr{
			Name:  "i",
			Type_: domain.NewIntType(),
		},
		Operator: domain.Lt,
		Right: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(10),
		},
	}
	condition.Left.(*domain.IdentifierExpr).SetType(domain.NewIntType())
	condition.Right.(*domain.LiteralExpr).SetType(domain.NewIntType())
	condition.SetType(domain.NewBoolType())
	
	update := &domain.AssignStmt{
		Target: &domain.IdentifierExpr{
			Name:  "i",
			Type_: domain.NewIntType(),
		},
		Value: &domain.BinaryExpr{
			Left: &domain.IdentifierExpr{
				Name:  "i",
				Type_: domain.NewIntType(),
			},
			Operator: domain.Add,
			Right: &domain.LiteralExpr{
				Type_: domain.NewIntType(),
				Value: int64(1),
			},
		},
	}
	update.Target.(*domain.IdentifierExpr).SetType(domain.NewIntType())
	update.Value.(*domain.BinaryExpr).Left.(*domain.IdentifierExpr).SetType(domain.NewIntType())
	update.Value.(*domain.BinaryExpr).Right.(*domain.LiteralExpr).SetType(domain.NewIntType())
	update.Value.(*domain.BinaryExpr).SetType(domain.NewIntType())
	
	body := &domain.BlockStmt{
		Statements: []domain.Statement{},
	}
	
	forStmt := &domain.ForStmt{
		Init:      init,
		Condition: condition,
		Update:    update,
		Body:      body,
	}
	
	err := forStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitForStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "for.cond") {
		t.Error("Expected for condition label")
	}
	if !strings.Contains(output, "for.body") {
		t.Error("Expected for body label")
	}
	if !strings.Contains(output, "for.inc") {
		t.Error("Expected for increment label")
	}
	if !strings.Contains(output, "for.end") {
		t.Error("Expected for end label")
	}
}

// TestVisitCallExpr tests function call code generation
func TestVisitCallExpr(t *testing.T) {
	generator := NewGenerator()
	
	// Test print function call
	printFunc := &domain.IdentifierExpr{
		Name:  "print",
		Type_: domain.NewVoidType(),
	}
	printFunc.SetType(domain.NewVoidType())
	
	arg := &domain.LiteralExpr{
		Type_: domain.NewStringType(),
		Value: "hello",
	}
	arg.SetType(domain.NewStringType())
	
	callExpr := &domain.CallExpr{
		Function: printFunc,
		Args:     []domain.Expression{arg},
	}
	callExpr.SetType(domain.NewVoidType())
	
	err := callExpr.Accept(generator)
	if err != nil {
		t.Fatalf("VisitCallExpr failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "call void @sl_print_string") {
		t.Error("Expected print function call")
	}
}

// TestVisitExprStmt tests expression statement code generation
func TestVisitExprStmt(t *testing.T) {
	generator := NewGenerator()
	
	expr := &domain.LiteralExpr{
		Type_: domain.NewIntType(),
		Value: int64(42),
	}
	expr.SetType(domain.NewIntType())
	
	exprStmt := &domain.ExprStmt{
		Expression: expr,
	}
	
	err := exprStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitExprStmt failed: %v", err)
	}
	
	// ExprStmt should generate code for its expression
	if generator.currentValue != "42" {
		t.Error("Expression statement should evaluate its expression")
	}
}

// TestVisitBlockStmt tests block statement code generation
func TestVisitBlockStmt(t *testing.T) {
	generator := NewGenerator()
	
	stmt1 := &domain.VarDeclStmt{
		Name:  "x",
		Type_: domain.NewIntType(),
		Initializer: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(5),
		},
	}
	stmt1.Initializer.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	stmt2 := &domain.VarDeclStmt{
		Name:  "y",
		Type_: domain.NewIntType(),
		Initializer: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(10),
		},
	}
	stmt2.Initializer.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	blockStmt := &domain.BlockStmt{
		Statements: []domain.Statement{stmt1, stmt2},
	}
	
	err := blockStmt.Accept(generator)
	if err != nil {
		t.Fatalf("VisitBlockStmt failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "%x = alloca i32") {
		t.Error("Block should contain first variable")
	}
	if !strings.Contains(output, "%y = alloca i32") {
		t.Error("Block should contain second variable")
	}
}

// TestVisitIndexExpr tests index expression error handling
func TestVisitIndexExpr(t *testing.T) {
	generator := NewGenerator()
	
	indexExpr := &domain.IndexExpr{
		Object: &domain.IdentifierExpr{Name: "arr"},
		Index:  &domain.LiteralExpr{Value: int64(0)},
	}
	
	err := indexExpr.Accept(generator)
	if err == nil {
		t.Error("VisitIndexExpr should return error for unimplemented feature")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Error("Error should indicate not implemented")
	}
}

// TestVisitMemberExpr tests member expression error handling
func TestVisitMemberExpr(t *testing.T) {
	generator := NewGenerator()
	
	memberExpr := &domain.MemberExpr{
		Object: &domain.IdentifierExpr{Name: "obj"},
		Member: "field",
	}
	
	err := memberExpr.Accept(generator)
	if err == nil {
		t.Error("VisitMemberExpr should return error for unimplemented feature")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Error("Error should indicate not implemented")
	}
}

// TestGenerateGlobalVariable tests global variable generation
func TestGenerateGlobalVariable(t *testing.T) {
	generator := NewGenerator()
	
	globalVar := &domain.VarDeclStmt{
		Name:  "global_x",
		Type_: domain.NewIntType(),
		Initializer: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: int64(42),
		},
	}
	globalVar.Initializer.(*domain.LiteralExpr).SetType(domain.NewIntType())
	
	err := generator.generateGlobalVariable(globalVar)
	if err != nil {
		t.Fatalf("generateGlobalVariable failed: %v", err)
	}
	
	output := generator.output.String()
	if !strings.Contains(output, "@global_x = global i32") {
		t.Error("Expected global variable declaration")
	}
	if !strings.Contains(output, "42") {
		t.Error("Expected global variable value")
	}
}

// TestValidateFormatArguments tests format validation
func TestValidateFormatArguments(t *testing.T) {
	generator := NewGenerator()
	
	// Test valid format
	err := generator.validateFormatArguments("Hello %s", []string{"string"})
	if err != nil {
		t.Errorf("Valid format should not return error: %v", err)
	}
	
	// Test invalid format - wrong type
	err = generator.validateFormatArguments("Hello %d", []string{"string"})
	if err == nil {
		t.Error("Invalid format should return error")
	}
	
	// Test invalid format - wrong count
	err = generator.validateFormatArguments("Hello %s %d", []string{"string"})
	if err == nil {
		t.Error("Wrong argument count should return error")
	}
}

// TestSetterMethods tests the setter methods
func TestSetterMethods(t *testing.T) {
	generator := NewGenerator()
	
	// Test SetLLVMBackend
	generator.SetLLVMBackend(nil)
	if generator.backend != nil {
		t.Error("SetLLVMBackend should set backend to nil")
	}
	
	// Test SetSymbolTable
	generator.SetSymbolTable(nil)
	if generator.symbolTable != nil {
		t.Error("SetSymbolTable should set symbolTable to nil")
	}
	
	// Test SetTypeRegistry
	generator.SetTypeRegistry(nil)
	if generator.typeRegistry != nil {
		t.Error("SetTypeRegistry should set typeRegistry to nil")
	}
	
	// Test SetErrorReporter
	generator.SetErrorReporter(nil)
	if generator.errorReporter != nil {
		t.Error("SetErrorReporter should set errorReporter to nil")
	}
}

// TestVisitProgram tests program traversal
func TestVisitProgram(t *testing.T) {
	generator := NewGenerator()
	
	// Create a program with a simple function
	program := &domain.Program{
		Declarations: []domain.Declaration{
			&domain.FunctionDecl{
				Name: "main",
				Parameters: []domain.Parameter{},
				ReturnType: &domain.BasicType{Kind: domain.VoidType},
				Body: &domain.BlockStmt{Statements: []domain.Statement{}},
			},
		},
	}
	
	// Visit the program
	generator.VisitProgram(program)
	
	output := generator.output.String()
	if !strings.Contains(output, "define void @main()") {
		t.Error("VisitProgram should generate function definition")
	}
}

// TestVisitStructDecl tests struct declaration generation
func TestVisitStructDecl(t *testing.T) {
	generator := NewGenerator()
	
	// Create a struct declaration
	structDecl := &domain.StructDecl{
		Name: "Point",
		Fields: []domain.StructField{
			{Name: "x", Type: &domain.BasicType{Kind: domain.IntType}},
			{Name: "y", Type: &domain.BasicType{Kind: domain.IntType}},
		},
	}
	
	// Visit the struct - just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("VisitStructDecl should not panic: %v", r)
		}
	}()
	generator.VisitStructDecl(structDecl)
}

// TestHandlePrintFunction tests print function handling
func TestHandlePrintFunction(t *testing.T) {
	generator := NewGenerator()
	
	// Create a print function call
	callExpr := &domain.CallExpr{
		Function: &domain.IdentifierExpr{Name: "print"},
		Args: []domain.Expression{
			&domain.LiteralExpr{Value: "Hello %s", Type_: &domain.BasicType{Kind: domain.StringType}},
			&domain.LiteralExpr{Value: "world", Type_: &domain.BasicType{Kind: domain.StringType}},
		},
	}
	
	// Test print handling
	generator.handlePrintFunction(callExpr)
	
	output := generator.output.String()
	if !strings.Contains(output, "printf") || len(output) == 0 {
		t.Errorf("handlePrintFunction should generate printf call, got: %s", output)
	}
}