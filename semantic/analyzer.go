// Package semantic provides semantic analysis for the StaticLang compiler
package semantic

import (
	"fmt"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// Analyzer implements the SemanticAnalyzer interface
type Analyzer struct {
	typeRegistry         domain.TypeRegistry
	symbolTable          interfaces.SymbolTable
	errorReporter        domain.ErrorReporter
	currentFunction      *domain.FunctionDecl
	builtinsInitialized bool
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		typeRegistry: domain.NewDefaultTypeRegistry(),
	}
}

// Analyze performs semantic analysis on the AST
func (a *Analyzer) Analyze(ast *domain.Program) error {
	// Initialize builtin functions first
	if !a.builtinsInitialized {
		if err := a.initializeBuiltinFunctions(); err != nil {
			return err
		}
		a.builtinsInitialized = true
	}

	// First pass: collect all function and struct declarations
	for _, decl := range ast.Declarations {
		if err := a.declareTopLevelSymbol(decl); err != nil {
			return err
		}
	}

	// Second pass: analyze function bodies
	for _, decl := range ast.Declarations {
		if err := decl.Accept(a); err != nil {
			return err
		}
	}

	return nil
}

// SetTypeRegistry sets the type registry
func (a *Analyzer) SetTypeRegistry(registry domain.TypeRegistry) {
	a.typeRegistry = registry
}

// SetSymbolTable sets the symbol table
func (a *Analyzer) SetSymbolTable(symbolTable interfaces.SymbolTable) {
	a.symbolTable = symbolTable
}

// SetErrorReporter sets the error reporter
func (a *Analyzer) SetErrorReporter(reporter domain.ErrorReporter) {
	a.errorReporter = reporter
}

// declareTopLevelSymbol declares function and struct symbols in the global scope
func (a *Analyzer) declareTopLevelSymbol(decl domain.Declaration) error {
	switch d := decl.(type) {
	case *domain.FunctionDecl:
		// Create function type
		paramTypes := make([]domain.Type, len(d.Parameters))
		for i, param := range d.Parameters {
			paramTypes[i] = param.Type
		}

		funcType := &domain.FunctionType{
			ParameterTypes: paramTypes,
			ReturnType:     d.ReturnType,
		}

		// Declare function symbol
		_, err := a.symbolTable.DeclareSymbol(
			d.Name,
			funcType,
			interfaces.FunctionSymbol,
			d.GetLocation(),
		)
		return err

	case *domain.StructDecl:
		// Create struct type
		structType, err := a.typeRegistry.CreateStructType(d.Name, d.Fields)
		if err != nil {
			return err
		}

		// Declare struct symbol
		_, err = a.symbolTable.DeclareSymbol(
			d.Name,
			structType,
			interfaces.StructSymbol,
			d.GetLocation(),
		)
		return err

	default:
		return fmt.Errorf("unknown declaration type: %T", decl)
	}
}

// reportError reports a semantic error
func (a *Analyzer) reportError(errorType domain.ErrorType, message string, location domain.SourceRange, context string, hints []string) {
	if a.errorReporter != nil {
		err := domain.CompilerError{
			Type:     errorType,
			Message:  message,
			Location: location,
			Context:  context,
			Hints:    hints,
		}
		a.errorReporter.ReportError(err)
	}
}

// Visitor pattern implementation for semantic analysis

// VisitProgram analyzes the program node
func (a *Analyzer) VisitProgram(prog *domain.Program) error {
	for _, decl := range prog.Declarations {
		if err := decl.Accept(a); err != nil {
			return err
		}
	}
	return nil
}

// initializeBuiltinFunctions adds builtin functions to the symbol table
func (a *Analyzer) initializeBuiltinFunctions() error {
	// Define builtin function types
	builtinFunctions := map[string]*domain.FunctionType{
		"print": {
			ParameterTypes: []domain.Type{}, // Variadic - will be handled specially
			ReturnType:     domain.NewVoidType(),
		},
	}

	// Add builtin functions to symbol table
	for name, funcType := range builtinFunctions {
		_, err := a.symbolTable.DeclareSymbol(
			name,
			funcType,
			interfaces.FunctionSymbol,
			domain.SourceRange{}, // Builtin functions have no source location
		)
		if err != nil {
			return fmt.Errorf("failed to declare builtin function %s: %v", name, err)
		}
	}

	return nil
}

// VisitFunctionDecl analyzes function declarations
func (a *Analyzer) VisitFunctionDecl(decl *domain.FunctionDecl) error {
	a.currentFunction = decl
	defer func() { a.currentFunction = nil }()

	// Enter function scope
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Declare parameters in function scope
	for _, param := range decl.Parameters {
		_, err := a.symbolTable.DeclareSymbol(
			param.Name,
			param.Type,
			interfaces.ParameterSymbol,
			decl.GetLocation(),
		)
		if err != nil {
			a.reportError(
				domain.SemanticError,
				fmt.Sprintf("duplicate parameter: %s", param.Name),
				decl.GetLocation(),
				"in function declaration",
				[]string{"parameter names must be unique within a function"},
			)
		}
	}

	// Analyze function body
	if decl.Body != nil {
		return decl.Body.Accept(a)
	}

	return nil
}

// VisitStructDecl analyzes struct declarations
func (a *Analyzer) VisitStructDecl(decl *domain.StructDecl) error {
	// Struct analysis is mostly done during declaration phase
	// Additional validation can be added here if needed
	return nil
}

// VisitBlockStmt analyzes block statements
func (a *Analyzer) VisitBlockStmt(stmt *domain.BlockStmt) error {
	// Enter new scope for block
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	for _, s := range stmt.Statements {
		if err := s.Accept(a); err != nil {
			return err
		}
	}
	return nil
}

// VisitVarDeclStmt analyzes variable declarations
func (a *Analyzer) VisitVarDeclStmt(stmt *domain.VarDeclStmt) error {
	// Check if initializer exists and type check it
	if stmt.Initializer != nil {
		if err := stmt.Initializer.Accept(a); err != nil {
			return err
		}

		// Type check assignment
		initType := stmt.Initializer.GetType()
		if !stmt.Type_.IsAssignableFrom(initType) {
			a.reportError(
				domain.TypeCheckError,
				fmt.Sprintf("cannot assign %s to variable of type %s", initType.String(), stmt.Type_.String()),
				stmt.GetLocation(),
				"in variable declaration",
				[]string{"ensure the initializer expression matches the declared type"},
			)
		}
	}

	// Declare variable symbol
	_, err := a.symbolTable.DeclareSymbol(
		stmt.Name,
		stmt.Type_,
		interfaces.VariableSymbol,
		stmt.GetLocation(),
	)
	if err != nil {
		a.reportError(
			domain.SemanticError,
			fmt.Sprintf("variable '%s' already declared", stmt.Name),
			stmt.GetLocation(),
			"in variable declaration",
			[]string{"variable names must be unique within a scope"},
		)
	}

	return nil
}

// VisitAssignStmt analyzes assignment statements
func (a *Analyzer) VisitAssignStmt(stmt *domain.AssignStmt) error {
	// Analyze target and value expressions
	if err := stmt.Target.Accept(a); err != nil {
		return err
	}
	if err := stmt.Value.Accept(a); err != nil {
		return err
	}

	// Type check assignment
	targetType := stmt.Target.GetType()
	valueType := stmt.Value.GetType()

	if !targetType.IsAssignableFrom(valueType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot assign %s to %s", valueType.String(), targetType.String()),
			stmt.GetLocation(),
			"in assignment",
			[]string{"ensure both sides of assignment have compatible types"},
		)
	}

	return nil
}

// VisitIfStmt analyzes if statements
func (a *Analyzer) VisitIfStmt(stmt *domain.IfStmt) error {
	// Analyze condition
	if err := stmt.Condition.Accept(a); err != nil {
		return err
	}

	// Check condition type
	condType := stmt.Condition.GetType()
	boolType := a.typeRegistry.GetBuiltinType(domain.BoolType)
	if !condType.Equals(boolType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("if condition must be bool, got %s", condType.String()),
			stmt.Condition.GetLocation(),
			"in if statement",
			[]string{"use a boolean expression as the condition"},
		)
	}

	// Analyze then statement
	if err := stmt.ThenStmt.Accept(a); err != nil {
		return err
	}

	// Analyze else statement if present
	if stmt.ElseStmt != nil {
		if err := stmt.ElseStmt.Accept(a); err != nil {
			return err
		}
	}

	return nil
}

// VisitWhileStmt analyzes while statements
func (a *Analyzer) VisitWhileStmt(stmt *domain.WhileStmt) error {
	// Analyze condition
	if err := stmt.Condition.Accept(a); err != nil {
		return err
	}

	// Check condition type
	condType := stmt.Condition.GetType()
	boolType := a.typeRegistry.GetBuiltinType(domain.BoolType)
	if !condType.Equals(boolType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("while condition must be bool, got %s", condType.String()),
			stmt.Condition.GetLocation(),
			"in while statement",
			[]string{"use a boolean expression as the condition"},
		)
	}

	// Analyze body
	return stmt.Body.Accept(a)
}

// VisitForStmt analyzes for statements
func (a *Analyzer) VisitForStmt(stmt *domain.ForStmt) error {
	// Enter new scope for for loop
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Analyze init statement
	if stmt.Init != nil {
		if err := stmt.Init.Accept(a); err != nil {
			return err
		}
	}

	// Analyze condition
	if stmt.Condition != nil {
		if err := stmt.Condition.Accept(a); err != nil {
			return err
		}

		// Check condition type
		condType := stmt.Condition.GetType()
		boolType := a.typeRegistry.GetBuiltinType(domain.BoolType)
		if !condType.Equals(boolType) {
			a.reportError(
				domain.TypeCheckError,
				fmt.Sprintf("for condition must be bool, got %s", condType.String()),
				stmt.Condition.GetLocation(),
				"in for statement",
				[]string{"use a boolean expression as the condition"},
			)
		}
	}

	// Analyze update statement
	if stmt.Update != nil {
		if err := stmt.Update.Accept(a); err != nil {
			return err
		}
	}

	// Analyze body
	return stmt.Body.Accept(a)
}

// VisitReturnStmt analyzes return statements
func (a *Analyzer) VisitReturnStmt(stmt *domain.ReturnStmt) error {
	if a.currentFunction == nil {
		a.reportError(
			domain.SemanticError,
			"return statement outside function",
			stmt.GetLocation(),
			"",
			[]string{"return statements can only be used inside functions"},
		)
		return nil
	}

	expectedReturnType := a.currentFunction.ReturnType

	if stmt.Value == nil {
		// Void return
		voidType := a.typeRegistry.GetBuiltinType(domain.VoidType)
		if !expectedReturnType.Equals(voidType) {
			a.reportError(
				domain.TypeCheckError,
				fmt.Sprintf("function expects return value of type %s", expectedReturnType.String()),
				stmt.GetLocation(),
				"in return statement",
				[]string{"add a return value or change function return type to void"},
			)
		}
	} else {
		// Analyze return value
		if err := stmt.Value.Accept(a); err != nil {
			return err
		}

		valueType := stmt.Value.GetType()
		if !expectedReturnType.IsAssignableFrom(valueType) {
			a.reportError(
				domain.TypeCheckError,
				fmt.Sprintf("cannot return %s from function expecting %s", valueType.String(), expectedReturnType.String()),
				stmt.GetLocation(),
				"in return statement",
				[]string{"ensure return value matches function return type"},
			)
		}
	}

	return nil
}

// VisitExprStmt analyzes expression statements
func (a *Analyzer) VisitExprStmt(stmt *domain.ExprStmt) error {
	return stmt.Expression.Accept(a)
}

// VisitBinaryExpr analyzes binary expressions
func (a *Analyzer) VisitBinaryExpr(expr *domain.BinaryExpr) error {
	// Analyze operands
	if err := expr.Left.Accept(a); err != nil {
		return err
	}
	if err := expr.Right.Accept(a); err != nil {
		return err
	}

	leftType := expr.Left.GetType()
	rightType := expr.Right.GetType()

	// Check if operation is valid
	if !domain.CanApplyBinaryOperator(expr.Operator, leftType, rightType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot apply operator %s to %s and %s", expr.Operator.String(), leftType.String(), rightType.String()),
			expr.GetLocation(),
			"in binary expression",
			[]string{"ensure operands have compatible types for the operator"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid binary operation"})
		return nil
	}

	// Determine result type
	var resultType domain.Type
	switch expr.Operator {
	case domain.Add, domain.Sub, domain.Mul, domain.Div, domain.Mod:
		resultType = leftType // Arithmetic operations preserve type
	case domain.Eq, domain.Ne, domain.Lt, domain.Le, domain.Gt, domain.Ge:
		resultType = a.typeRegistry.GetBuiltinType(domain.BoolType) // Comparison operations return bool
	case domain.And, domain.Or:
		resultType = a.typeRegistry.GetBuiltinType(domain.BoolType) // Logical operations return bool
	default:
		resultType = &domain.TypeError{Message: "unknown binary operator"}
	}

	expr.SetType(resultType)
	return nil
}

// VisitUnaryExpr analyzes unary expressions
func (a *Analyzer) VisitUnaryExpr(expr *domain.UnaryExpr) error {
	// Analyze operand
	if err := expr.Operand.Accept(a); err != nil {
		return err
	}

	operandType := expr.Operand.GetType()

	// Check if operation is valid
	if !domain.CanApplyUnaryOperator(expr.Operator, operandType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot apply operator %s to %s", expr.Operator.String(), operandType.String()),
			expr.GetLocation(),
			"in unary expression",
			[]string{"ensure operand has compatible type for the operator"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid unary operation"})
		return nil
	}

	// Result type is same as operand for unary operations
	expr.SetType(operandType)
	return nil
}

// VisitCallExpr analyzes function call expressions
func (a *Analyzer) VisitCallExpr(expr *domain.CallExpr) error {
	// Analyze function expression
	if err := expr.Function.Accept(a); err != nil {
		return err
	}

	funcType, ok := expr.Function.GetType().(*domain.FunctionType)
	if !ok {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot call non-function type %s", expr.Function.GetType().String()),
			expr.GetLocation(),
			"in function call",
			[]string{"ensure the expression is a function"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid function call"})
		return nil
	}

	// Special handling for builtin functions
	if identExpr, ok := expr.Function.(*domain.IdentifierExpr); ok {
		if identExpr.Name == "print" {
			// Special handling for print function - it's variadic
			return a.handlePrintFunction(expr)
		}
	}

	// Check argument count for regular functions
	if len(expr.Args) != len(funcType.ParameterTypes) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("function expects %d arguments, got %d", len(funcType.ParameterTypes), len(expr.Args)),
			expr.GetLocation(),
			"in function call",
			[]string{"provide the correct number of arguments"},
		)
	}

	// Analyze and type check arguments
	for i, arg := range expr.Args {
		if err := arg.Accept(a); err != nil {
			return err
		}

		if i < len(funcType.ParameterTypes) {
			expectedType := funcType.ParameterTypes[i]
			actualType := arg.GetType()

			if !expectedType.IsAssignableFrom(actualType) {
				a.reportError(
					domain.TypeCheckError,
					fmt.Sprintf("argument %d: cannot pass %s to parameter of type %s", i+1, actualType.String(), expectedType.String()),
					arg.GetLocation(),
					"in function call",
					[]string{"ensure argument types match parameter types"},
				)
			}
		}
	}

	expr.SetType(funcType.ReturnType)
	return nil
}

// handlePrintFunction performs special validation for the print builtin function
func (a *Analyzer) handlePrintFunction(expr *domain.CallExpr) error {
	// Analyze all arguments
	for _, arg := range expr.Args {
		if err := arg.Accept(a); err != nil {
			return err
		}
	}

	// Print function validation logic
	if len(expr.Args) == 0 {
		a.reportError(
			domain.SemanticError,
			"print function requires at least one argument",
			expr.GetLocation(),
			"in print function call",
			[]string{"provide at least one argument to print"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid print call"})
		return nil
	}

	// For now, accept any argument types for print function
	// The code generator will handle type-specific formatting
	expr.SetType(domain.NewVoidType())
	return nil
}

// VisitIdentifierExpr analyzes identifier expressions
func (a *Analyzer) VisitIdentifierExpr(expr *domain.IdentifierExpr) error {
	// Look up symbol
	symbol, found := a.symbolTable.LookupSymbol(expr.Name)
	if !found {
		a.reportError(
			domain.SemanticError,
			fmt.Sprintf("undefined identifier: %s", expr.Name),
			expr.GetLocation(),
			"",
			[]string{"ensure the identifier is declared before use"},
		)
		expr.SetType(&domain.TypeError{Message: "undefined identifier"})
		return nil
	}

	expr.SetType(symbol.Type)
	return nil
}

// VisitLiteralExpr analyzes literal expressions
func (a *Analyzer) VisitLiteralExpr(expr *domain.LiteralExpr) error {
	// Determine type based on value
	var literalType domain.Type

	switch v := expr.Value.(type) {
	case int64:
		literalType = a.typeRegistry.GetBuiltinType(domain.IntType)
	case float64:
		literalType = a.typeRegistry.GetBuiltinType(domain.FloatType)
	case string:
		literalType = a.typeRegistry.GetBuiltinType(domain.StringType)
	case bool:
		literalType = a.typeRegistry.GetBuiltinType(domain.BoolType)
	default:
		literalType = &domain.TypeError{Message: fmt.Sprintf("unknown literal type: %T", v)}
	}

	expr.SetType(literalType)
	return nil
}

// VisitIndexExpr analyzes array index expressions
func (a *Analyzer) VisitIndexExpr(expr *domain.IndexExpr) error {
	// Analyze object and index
	if err := expr.Object.Accept(a); err != nil {
		return err
	}
	if err := expr.Index.Accept(a); err != nil {
		return err
	}

	objectType := expr.Object.GetType()
	indexType := expr.Index.GetType()

	// Check if object is an array
	arrayType, ok := objectType.(*domain.ArrayType)
	if !ok {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot index non-array type %s", objectType.String()),
			expr.GetLocation(),
			"in index expression",
			[]string{"ensure the object is an array"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid index operation"})
		return nil
	}

	// Check if index is integer
	intType := a.typeRegistry.GetBuiltinType(domain.IntType)
	if !indexType.Equals(intType) {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("array index must be int, got %s", indexType.String()),
			expr.Index.GetLocation(),
			"in index expression",
			[]string{"use an integer expression as the index"},
		)
	}

	expr.SetType(arrayType.ElementType)
	return nil
}

// VisitMemberExpr analyzes struct member access expressions
func (a *Analyzer) VisitMemberExpr(expr *domain.MemberExpr) error {
	// Analyze object
	if err := expr.Object.Accept(a); err != nil {
		return err
	}

	objectType := expr.Object.GetType()

	// Check if object is a struct
	structType, ok := objectType.(*domain.StructType)
	if !ok {
		a.reportError(
			domain.TypeCheckError,
			fmt.Sprintf("cannot access member of non-struct type %s", objectType.String()),
			expr.GetLocation(),
			"in member access",
			[]string{"ensure the object is a struct"},
		)
		expr.SetType(&domain.TypeError{Message: "invalid member access"})
		return nil
	}

	// Check if member exists
	memberType, exists := structType.GetField(expr.Member)
	if !exists {
		a.reportError(
			domain.SemanticError,
			fmt.Sprintf("struct %s has no member %s", structType.Name, expr.Member),
			expr.GetLocation(),
			"in member access",
			[]string{fmt.Sprintf("available members: %v", structType.Order)},
		)
		expr.SetType(&domain.TypeError{Message: "undefined member"})
		return nil
	}

	expr.SetType(memberType)
	return nil
}
