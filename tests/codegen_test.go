package tests

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/codegen"
	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/infrastructure"
)

func TestCodeGenBasicProgram(t *testing.T) {
	generator := codegen.NewGenerator()

	// Create a mock LLVM backend
	backend := infrastructure.NewMockLLVMBackend()
	symbolTable := infrastructure.NewSymbolTable()
	typeRegistry := domain.NewTypeRegistry()
	errorReporter := infrastructure.NewConsoleErrorReporter(nil)

	generator.SetLLVMBackend(backend)
	generator.SetSymbolTable(symbolTable)
	generator.SetTypeRegistry(typeRegistry)
	generator.SetErrorReporter(errorReporter)

	// Create a simple program AST
	mainFunc := &domain.FunctionDecl{
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{
				&domain.ReturnStmt{
					Value: &domain.LiteralExpr{
						Type_: domain.NewIntType(),
						Value: "0",
					},
				},
			},
		},
	}

	program := &domain.Program{
		Declarations: []domain.Declaration{mainFunc},
	}

	// Set types for expressions
	mainFunc.Body.Statements[0].(*domain.ReturnStmt).Value.(*domain.LiteralExpr).SetType(domain.NewIntType())

	// Generate code
	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check that the result contains expected LLVM IR elements
	if !strings.Contains(result, "define i32 @main()") {
		t.Error("Generated code should contain main function definition")
	}

	if !strings.Contains(result, "ret i32 0") {
		t.Error("Generated code should contain return statement")
	}

	if !strings.Contains(result, "declare i32 @printf(i8*, ...)") {
		t.Error("Generated code should contain printf declaration")
	}
}

func TestCodeGenVariableDeclaration(t *testing.T) {
	generator := codegen.NewGenerator()

	// Setup mocks
	backend := infrastructure.NewMockLLVMBackend()
	symbolTable := infrastructure.NewSymbolTable()
	typeRegistry := domain.NewTypeRegistry()
	errorReporter := infrastructure.NewConsoleErrorReporter(nil)

	generator.SetLLVMBackend(backend)
	generator.SetSymbolTable(symbolTable)
	generator.SetTypeRegistry(typeRegistry)
	generator.SetErrorReporter(errorReporter)

	// Create variable declaration
	varDecl := &domain.VarDeclStmt{
		Name:  "x",
		Type_: domain.NewIntType(),
		Initializer: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: "42",
		},
	}
	varDecl.Initializer.(*domain.LiteralExpr).SetType(domain.NewIntType())

	mainFunc := &domain.FunctionDecl{
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{
				varDecl,
				&domain.ReturnStmt{
					Value: &domain.LiteralExpr{
						Type_: domain.NewIntType(),
						Value: "0",
					},
				},
			},
		},
	}

	program := &domain.Program{
		Declarations: []domain.Declaration{mainFunc},
	}

	// Set types
	mainFunc.Body.Statements[1].(*domain.ReturnStmt).Value.(*domain.LiteralExpr).SetType(domain.NewIntType())

	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check for variable allocation
	if !strings.Contains(result, "%x = alloca i32") {
		t.Error("Generated code should contain variable allocation")
	}
}

func TestCodeGenBinaryExpression(t *testing.T) {
	generator := codegen.NewGenerator()

	// Setup mocks
	backend := infrastructure.NewMockLLVMBackend()
	symbolTable := infrastructure.NewSymbolTable()
	typeRegistry := domain.NewTypeRegistry()
	errorReporter := infrastructure.NewConsoleErrorReporter(nil)

	generator.SetLLVMBackend(backend)
	generator.SetSymbolTable(symbolTable)
	generator.SetTypeRegistry(typeRegistry)
	generator.SetErrorReporter(errorReporter)

	// Create binary expression: 5 + 3
	leftExpr := &domain.LiteralExpr{
		Type_: domain.NewIntType(),
		Value: "5",
	}
	leftExpr.SetType(domain.NewIntType())

	rightExpr := &domain.LiteralExpr{
		Type_: domain.NewIntType(),
		Value: "3",
	}
	rightExpr.SetType(domain.NewIntType())

	binaryExpr := &domain.BinaryExpr{
		Left:     leftExpr,
		Operator: domain.Add,
		Right:    rightExpr,
	}
	binaryExpr.SetType(domain.NewIntType())

	returnStmt := &domain.ReturnStmt{
		Value: binaryExpr,
	}

	mainFunc := &domain.FunctionDecl{
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{returnStmt},
		},
	}

	program := &domain.Program{
		Declarations: []domain.Declaration{mainFunc},
	}

	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check for addition instruction
	if !strings.Contains(result, "add i32") {
		t.Error("Generated code should contain integer addition")
	}
}

func TestCodeGenFunctionCall(t *testing.T) {
	generator := codegen.NewGenerator()

	// Setup mocks
	backend := infrastructure.NewMockLLVMBackend()
	symbolTable := infrastructure.NewSymbolTable()
	typeRegistry := domain.NewTypeRegistry()
	errorReporter := infrastructure.NewConsoleErrorReporter(nil)

	generator.SetLLVMBackend(backend)
	generator.SetSymbolTable(symbolTable)
	generator.SetTypeRegistry(typeRegistry)
	generator.SetErrorReporter(errorReporter)

	// Create print function call
	arg := &domain.LiteralExpr{
		Type_: domain.NewStringType(),
		Value: "\"Hello\"",
	}
	arg.SetType(domain.NewStringType())

	printFunc := &domain.IdentifierExpr{
		Name:  "print",
		Type_: domain.NewIntType(),
	}
	printFunc.SetType(domain.NewIntType())

	callExpr := &domain.CallExpr{
		Function: printFunc,
		Args:     []domain.Expression{arg},
	}
	callExpr.SetType(domain.NewIntType()) // print returns int (like printf)

	exprStmt := &domain.ExprStmt{
		Expression: callExpr,
	}

	mainFunc := &domain.FunctionDecl{
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{
				exprStmt,
				&domain.ReturnStmt{
					Value: &domain.LiteralExpr{
						Type_: domain.NewIntType(),
						Value: "0",
					},
				},
			},
		},
	}

	// Set type for return expression
	mainFunc.Body.Statements[1].(*domain.ReturnStmt).Value.(*domain.LiteralExpr).SetType(domain.NewIntType())

	program := &domain.Program{
		Declarations: []domain.Declaration{mainFunc},
	}

	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check for printf call
	if !strings.Contains(result, "call i32 (i8*, ...) @printf") {
		t.Error("Generated code should contain printf call for print function")
	}
}

func TestCodeGenControlFlow(t *testing.T) {
	generator := codegen.NewGenerator()

	// Setup mocks
	backend := infrastructure.NewMockLLVMBackend()
	symbolTable := infrastructure.NewSymbolTable()
	typeRegistry := domain.NewTypeRegistry()
	errorReporter := infrastructure.NewConsoleErrorReporter(nil)

	generator.SetLLVMBackend(backend)
	generator.SetSymbolTable(symbolTable)
	generator.SetTypeRegistry(typeRegistry)
	generator.SetErrorReporter(errorReporter)

	// Create if statement: if (1 > 0) return 1; else return 0;
	condition := &domain.BinaryExpr{
		Left: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: "1",
		},
		Operator: domain.Gt,
		Right: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: "0",
		},
	}

	// Set types
	condition.Left.(*domain.LiteralExpr).SetType(domain.NewIntType())
	condition.Right.(*domain.LiteralExpr).SetType(domain.NewIntType())
	condition.SetType(domain.NewBoolType())

	thenStmt := &domain.ReturnStmt{
		Value: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: "1",
		},
	}
	thenStmt.Value.(*domain.LiteralExpr).SetType(domain.NewIntType())

	elseStmt := &domain.ReturnStmt{
		Value: &domain.LiteralExpr{
			Type_: domain.NewIntType(),
			Value: "0",
		},
	}
	elseStmt.Value.(*domain.LiteralExpr).SetType(domain.NewIntType())

	ifStmt := &domain.IfStmt{
		Condition: condition,
		ThenStmt:  thenStmt,
		ElseStmt:  elseStmt,
	}

	mainFunc := &domain.FunctionDecl{
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: domain.NewIntType(),
		Body: &domain.BlockStmt{
			Statements: []domain.Statement{ifStmt},
		},
	}

	program := &domain.Program{
		Declarations: []domain.Declaration{mainFunc},
	}

	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check for control flow elements
	if !strings.Contains(result, "icmp sgt i32") {
		t.Error("Generated code should contain integer comparison")
	}

	if !strings.Contains(result, "br i1") {
		t.Error("Generated code should contain conditional branch")
	}

	if !strings.Contains(result, "if.then") {
		t.Error("Generated code should contain then block label")
	}

	if !strings.Contains(result, "if.else") {
		t.Error("Generated code should contain else block label")
	}
}
