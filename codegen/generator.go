package codegen

import (
	"fmt"
	"strings"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// Generator implements the CodeGenerator interface for LLVM IR generation
type Generator struct {
	backend       interfaces.LLVMBackend
	symbolTable   interfaces.SymbolTable
	typeRegistry  domain.TypeRegistry
	errorReporter domain.ErrorReporter
	output        strings.Builder
	indentLevel   int
	labelCounter  int
	functionName  string
	currentValue  string          // Holds the current expression result value
	currentType   string          // Holds the current expression result type
	parameters    map[string]bool // Track which identifiers are function parameters
}

// NewGenerator creates a new code generator
func NewGenerator() *Generator {
	return &Generator{
		labelCounter: 0,
		parameters:   make(map[string]bool),
	}
}

// SetLLVMBackend sets the LLVM backend
func (g *Generator) SetLLVMBackend(backend interfaces.LLVMBackend) {
	g.backend = backend
}

// SetSymbolTable sets the symbol table
func (g *Generator) SetSymbolTable(table interfaces.SymbolTable) {
	g.symbolTable = table
}

// SetTypeRegistry sets the type registry
func (g *Generator) SetTypeRegistry(registry domain.TypeRegistry) {
	g.typeRegistry = registry
}

// SetErrorReporter sets the error reporter
func (g *Generator) SetErrorReporter(reporter domain.ErrorReporter) {
	g.errorReporter = reporter
}

// Generate generates LLVM IR for the given AST
func (g *Generator) Generate(node domain.Node) (string, error) {
	g.output.Reset()
	g.indentLevel = 0
	g.labelCounter = 0

	// Initialize LLVM backend
	if g.backend != nil {
		if err := g.backend.Initialize("x86_64-apple-macosx10.15.0"); err != nil {
			return "", fmt.Errorf("failed to initialize LLVM backend: %v", err)
		}
		defer g.backend.Dispose()
	}

	// Generate the code
	err := node.Accept(g)
	if err != nil {
		return "", err
	}

	return g.output.String(), nil
}

// Helper methods for code generation
func (g *Generator) emit(format string, args ...interface{}) {
	indent := strings.Repeat("  ", g.indentLevel)
	g.output.WriteString(indent)
	g.output.WriteString(fmt.Sprintf(format, args...))
	g.output.WriteString("\n")
}

func (g *Generator) emitRaw(text string) {
	g.output.WriteString(text)
}

func (g *Generator) newLabel(prefix string) string {
	g.labelCounter++
	return fmt.Sprintf("%s%d", prefix, g.labelCounter)
}

// Visitor pattern implementation for AST nodes
func (g *Generator) VisitProgram(prog *domain.Program) error {
	// Emit LLVM module header
	g.emit("; ModuleID = 'staticlang'")
	g.emit("target datalayout = \"e-m:o-i64:64-f80:128-n8:16:32:64-S128\"")
	g.emit("target triple = \"x86_64-apple-macosx10.15.0\"")
	g.emit("")

	// Emit external function declarations
	g.emit("; External function declarations")
	g.emit("declare i32 @printf(i8*, ...)")
	g.emit("declare i8* @malloc(i64)")
	g.emit("declare void @free(i8*)")
	g.emit("")

	// Emit StaticLang builtin functions
	g.emit("; StaticLang builtin functions")
	g.emit("declare void @sl_print_int(i32)")
	g.emit("declare void @sl_print_double(double)")
	g.emit("declare void @sl_print_string(i8*)")
	g.emit("declare i8* @sl_alloc_string(i8*)")
	g.emit("declare i8* @sl_concat_string(i8*, i8*)")
	g.emit("declare i32 @sl_compare_string(i8*, i8*)")
	g.emit("declare i8* @sl_alloc_array(i64, i64)")
	g.emit("")

	// Process all declarations
	for _, decl := range prog.Declarations {
		if err := decl.Accept(g); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateGlobalVariable(varDecl *domain.VarDeclStmt) error {
	if varDecl.Initializer != nil {
		// Initialize with value
		if lit, ok := varDecl.Initializer.(*domain.LiteralExpr); ok {
			switch varDecl.Type_.String() {
			case "int":
				g.emit("@%s = global i32 %s, align 4", varDecl.Name, lit.Value)
			case "double":
				g.emit("@%s = global double %s, align 8", varDecl.Name, lit.Value)
			case "string":
				// String literals need special handling
				strValue := strings.Trim(lit.Value.(string), "\"")
				length := len(strValue) + 1
				g.emit("@%s.str = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1", varDecl.Name, length, strValue)
				g.emit("@%s = global i8* getelementptr inbounds ([%d x i8], [%d x i8]* @%s.str, i32 0, i32 0), align 8", varDecl.Name, length, length, varDecl.Name)
			}
		} else {
			// Initialize with zero
			switch varDecl.Type_.String() {
			case "int":
				g.emit("@%s = global i32 0, align 4", varDecl.Name)
			case "double":
				g.emit("@%s = global double 0.0, align 8", varDecl.Name)
			case "string":
				g.emit("@%s = global i8* null, align 8", varDecl.Name)
			}
		}
	} else {
		// Initialize with zero/null
		switch varDecl.Type_.String() {
		case "int":
			g.emit("@%s = global i32 0, align 4", varDecl.Name)
		case "double":
			g.emit("@%s = global double 0.0, align 8", varDecl.Name)
		case "string":
			g.emit("@%s = global i8* null, align 8", varDecl.Name)
		}
	}

	return nil
}

func (g *Generator) VisitFunctionDecl(node *domain.FunctionDecl) error {
	g.functionName = node.Name
	returnType := g.getLLVMType(node.ReturnType)

	// Clear and track parameters for this function
	g.parameters = make(map[string]bool)
	for _, param := range node.Parameters {
		g.parameters[param.Name] = true
	}

	// Generate function signature
	paramStr := ""
	for i, param := range node.Parameters {
		if i > 0 {
			paramStr += ", "
		}
		paramStr += fmt.Sprintf("%s %%%s", g.getLLVMType(param.Type), param.Name)
	}

	g.emit("define %s @%s(%s) {", returnType, node.Name, paramStr)
	g.emit("entry:")
	g.indentLevel++

	// Allocate parameters on stack
	for _, param := range node.Parameters {
		g.emit("%%%s.addr = alloca %s, align %d", param.Name, g.getLLVMType(param.Type), g.getTypeAlign(param.Type))
		g.emit("store %s %%%s, ptr %%%s.addr, align %d", g.getLLVMType(param.Type), param.Name, param.Name, g.getTypeAlign(param.Type))
	}

	// Generate function body
	if err := node.Body.Accept(g); err != nil {
		return err
	}

	// Check if the function already has a return statement by examining the output
	// If not, add a default return
	outputStr := g.output.String()
	lines := strings.Split(outputStr, "\n")
	hasReturn := false
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || line == "}" {
			continue
		}
		if strings.HasPrefix(line, "ret ") {
			hasReturn = true
			break
		}
		// If we encounter any other instruction, we assume no return
		if line != "" {
			break
		}
	}

	// Only add default return if there's no explicit return
	if !hasReturn {
		if node.ReturnType.String() == "void" {
			g.emit("ret void")
		} else if node.Name == "main" {
			g.emit("ret i32 0")
		}
	}

	g.indentLevel--
	g.emit("}")
	g.emit("")

	return nil
}

func (g *Generator) VisitStructDecl(node *domain.StructDecl) error {
	// Not implemented for now
	return fmt.Errorf("struct declarations not yet implemented")
}

func (g *Generator) VisitBlockStmt(node *domain.BlockStmt) error {
	for _, stmt := range node.Statements {
		if err := stmt.Accept(g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) VisitVarDeclStmt(node *domain.VarDeclStmt) error {
	llvmType := g.getLLVMType(node.Type_)
	align := g.getTypeAlign(node.Type_)

	// Allocate local variable
	g.emit("%%%s = alloca %s, align %d", node.Name, llvmType, align)

	// Initialize if there's an initializer
	if node.Initializer != nil {
		if err := node.Initializer.Accept(g); err != nil {
			return err
		}
		// The expression result should be in g.currentValue
		g.emit("store %s %s, ptr %%%s, align %d", llvmType, g.currentValue, node.Name, align)
	}

	return nil
}

func (g *Generator) VisitAssignStmt(node *domain.AssignStmt) error {
	// Generate code for the expression (value)
	if err := node.Value.Accept(g); err != nil {
		return err
	}

	// Store the result into the target
	varType := g.getLLVMType(node.Target.GetType())
	align := g.getTypeAlign(node.Target.GetType())

	if ident, ok := node.Target.(*domain.IdentifierExpr); ok {
		g.emit("store %s %s, ptr %%%s, align %d", varType, g.currentValue, ident.Name, align)
	}

	return nil
}

func (g *Generator) VisitIfStmt(node *domain.IfStmt) error {
	thenLabel := g.newLabel("if.then")
	elseLabel := g.newLabel("if.else")
	endLabel := g.newLabel("if.end")

	// Generate condition
	if err := node.Condition.Accept(g); err != nil {
		return err
	}
	conditionReg := g.currentValue

	// Branch based on condition
	if node.ElseStmt != nil {
		g.emit("br i1 %s, label %%%s, label %%%s", conditionReg, thenLabel, elseLabel)
	} else {
		g.emit("br i1 %s, label %%%s, label %%%s", conditionReg, thenLabel, endLabel)
	}

	// Then block
	g.indentLevel--
	g.emit("%s:", thenLabel)
	g.indentLevel++
	if err := node.ThenStmt.Accept(g); err != nil {
		return err
	}

	// Check if the then block ends with a return statement
	// If so, don't add a branch to avoid unreachable code
	outputStr := g.output.String()
	lines := strings.Split(outputStr, "\n")
	lastNonEmptyLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			lastNonEmptyLine = line
			break
		}
	}

	thenHasReturn := strings.HasPrefix(lastNonEmptyLine, "ret ")
	if !thenHasReturn {
		g.emit("br label %%%s", endLabel)
	}

	// Else block (if exists)
	if node.ElseStmt != nil {
		g.indentLevel--
		g.emit("%s:", elseLabel)
		g.indentLevel++
		if err := node.ElseStmt.Accept(g); err != nil {
			return err
		}

		// Check if the else block ends with a return statement
		outputStr = g.output.String()
		lines = strings.Split(outputStr, "\n")
		lastNonEmptyLine = ""
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line != "" {
				lastNonEmptyLine = line
				break
			}
		}

		elseHasReturn := strings.HasPrefix(lastNonEmptyLine, "ret ")
		if !elseHasReturn {
			g.emit("br label %%%s", endLabel)
		}
	}

	// Only emit the end block if it's reachable
	// (i.e., if at least one branch doesn't end with return)
	needsEndBlock := !thenHasReturn || (node.ElseStmt != nil && !strings.HasPrefix(lastNonEmptyLine, "ret "))
	if node.ElseStmt == nil {
		needsEndBlock = true // Always need end block if there's no else
	}

	if needsEndBlock {
		g.indentLevel--
		g.emit("%s:", endLabel)
		g.indentLevel++
	}

	return nil
}

func (g *Generator) VisitWhileStmt(node *domain.WhileStmt) error {
	condLabel := g.newLabel("while.cond")
	bodyLabel := g.newLabel("while.body")
	endLabel := g.newLabel("while.end")

	// Jump to condition
	g.emit("br label %%%s", condLabel)

	// Condition block
	g.indentLevel--
	g.emit("%s:", condLabel)
	g.indentLevel++
	if err := node.Condition.Accept(g); err != nil {
		return err
	}
	conditionReg := g.currentValue
	g.emit("br i1 %s, label %%%s, label %%%s", conditionReg, bodyLabel, endLabel)

	// Body block
	g.indentLevel--
	g.emit("%s:", bodyLabel)
	g.indentLevel++
	if err := node.Body.Accept(g); err != nil {
		return err
	}
	g.emit("br label %%%s", condLabel)

	// End block
	g.indentLevel--
	g.emit("%s:", endLabel)
	g.indentLevel++

	return nil
}

func (g *Generator) VisitForStmt(node *domain.ForStmt) error {
	// Initialize
	if node.Init != nil {
		if err := node.Init.Accept(g); err != nil {
			return err
		}
	}

	condLabel := g.newLabel("for.cond")
	bodyLabel := g.newLabel("for.body")
	incLabel := g.newLabel("for.inc")
	endLabel := g.newLabel("for.end")

	// Jump to condition
	g.emit("br label %%%s", condLabel)

	// Condition block
	g.indentLevel--
	g.emit("%s:", condLabel)
	g.indentLevel++
	if node.Condition != nil {
		if err := node.Condition.Accept(g); err != nil {
			return err
		}
		conditionReg := g.currentValue
		g.emit("br i1 %s, label %%%s, label %%%s", conditionReg, bodyLabel, endLabel)
	} else {
		g.emit("br label %%%s", bodyLabel)
	}

	// Body block
	g.indentLevel--
	g.emit("%s:", bodyLabel)
	g.indentLevel++
	if err := node.Body.Accept(g); err != nil {
		return err
	}
	g.emit("br label %%%s", incLabel)

	// Increment block
	g.indentLevel--
	g.emit("%s:", incLabel)
	g.indentLevel++
	if node.Update != nil {
		if err := node.Update.Accept(g); err != nil {
			return err
		}
	}
	g.emit("br label %%%s", condLabel)

	// End block
	g.indentLevel--
	g.emit("%s:", endLabel)
	g.indentLevel++

	return nil
}

func (g *Generator) VisitReturnStmt(node *domain.ReturnStmt) error {
	if node.Value != nil {
		if err := node.Value.Accept(g); err != nil {
			return err
		}
		returnType := g.getLLVMType(node.Value.GetType())
		g.emit("ret %s %s", returnType, g.currentValue)
	} else {
		g.emit("ret void")
	}
	return nil
}

func (g *Generator) VisitExprStmt(node *domain.ExprStmt) error {
	return node.Expression.Accept(g)
}

func (g *Generator) VisitBinaryExpr(node *domain.BinaryExpr) error {
	// Generate left operand
	if err := node.Left.Accept(g); err != nil {
		return err
	}
	leftReg := g.currentValue

	// Generate right operand
	if err := node.Right.Accept(g); err != nil {
		return err
	}
	rightReg := g.currentValue

	// Generate unique temporary register
	tempReg := fmt.Sprintf("%%temp_%d", g.labelCounter)
	g.labelCounter++

	// Perform operation based on operator
	resultType := g.getLLVMType(node.GetType())

	switch node.Operator {
	case domain.Add:
		if resultType == "i32" {
			g.emit("%s = add i32 %s, %s", tempReg, leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%s = fadd double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Sub:
		if resultType == "i32" {
			g.emit("%s = sub i32 %s, %s", tempReg, leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%s = fsub double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Mul:
		if resultType == "i32" {
			g.emit("%s = mul i32 %s, %s", tempReg, leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%s = fmul double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Div:
		if resultType == "i32" {
			g.emit("%s = sdiv i32 %s, %s", tempReg, leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%s = fdiv double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Eq:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp eq i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp oeq double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Ne:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp ne i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp one double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Lt:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp slt i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp olt double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Gt:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp sgt i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp ogt double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Le:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp sle i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp ole double %s, %s", tempReg, leftReg, rightReg)
		}
	case domain.Ge:
		if node.Left.GetType().String() == "int" {
			g.emit("%s = icmp sge i32 %s, %s", tempReg, leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%s = fcmp oge double %s, %s", tempReg, leftReg, rightReg)
		}
	}

	// Update current value for parent expressions
	g.currentValue = tempReg
	g.currentType = resultType

	return nil
}

func (g *Generator) VisitUnaryExpr(node *domain.UnaryExpr) error {
	if err := node.Operand.Accept(g); err != nil {
		return err
	}

	switch node.Operator {
	case domain.Neg:
		if node.GetType().String() == "int" {
			g.emit("%%temp_result = sub i32 0, %%temp_result")
		} else if node.GetType().String() == "double" {
			g.emit("%%temp_result = fsub double 0.0, %%temp_result")
		}
	case domain.Not:
		g.emit("%%temp_result = icmp eq i1 %%temp_result, false")
	}

	return nil
}

func (g *Generator) VisitCallExpr(node *domain.CallExpr) error {
	// Special handling for built-in print function
	if ident, ok := node.Function.(*domain.IdentifierExpr); ok && ident.Name == "print" {
		return g.handlePrintFunction(node)
	}

	// Generate arguments for regular function calls
	var argValues []string
	var argTypes []string
	for _, arg := range node.Args {
		if err := arg.Accept(g); err != nil {
			return err
		}
		argType := g.getLLVMType(arg.GetType())
		argTypes = append(argTypes, argType)
		argValues = append(argValues, fmt.Sprintf("%s %s", argType, g.currentValue))
	}

	// Determine function name (assume identifier)
	funcName := "<unknown>"
	if ident, ok := node.Function.(*domain.IdentifierExpr); ok {
		funcName = ident.Name
	}

	// Generate unique temporary register for the result
	tempReg := fmt.Sprintf("%%temp_%d", g.labelCounter)
	g.labelCounter++

	// Generate function call
	returnType := g.getLLVMType(node.GetType())
	argsStr := ""
	if len(argValues) > 0 {
		argsStr = fmt.Sprintf("%s", strings.Join(argValues, ", "))
	}

	if returnType == "void" {
		g.emit("call void @%s(%s)", funcName, argsStr)
		g.currentValue = ""
		g.currentType = "void"
	} else {
		g.emit("%s = call %s @%s(%s)", tempReg, returnType, funcName, argsStr)
		g.currentValue = tempReg
		g.currentType = returnType
	}

	return nil
}

// handlePrintFunction handles both simple print(value) and formatted print("format", args...)
func (g *Generator) handlePrintFunction(node *domain.CallExpr) error {
	if len(node.Args) == 0 {
		return fmt.Errorf("print function requires at least one argument")
	}

	// Single argument: use builtin sl_print_* functions
	if len(node.Args) == 1 {
		if err := node.Args[0].Accept(g); err != nil {
			return err
		}

		argType := node.Args[0].GetType().String()
		switch argType {
		case "int":
			g.emit("call void @sl_print_int(i32 %s)", g.currentValue)
		case "double":
			g.emit("call void @sl_print_double(double %s)", g.currentValue)
		case "string":
			g.emit("call void @sl_print_string(i8* %s)", g.currentValue)
		default:
			return fmt.Errorf("unsupported type for print: %s", argType)
		}

		// print functions are void, set current value accordingly
		g.currentValue = ""
		g.currentType = "void"
		return nil
	}

	// Multiple arguments: formatted printing with printf
	// First argument should be format string
	if err := node.Args[0].Accept(g); err != nil {
		return err
	}

	if node.Args[0].GetType().String() != "string" {
		return fmt.Errorf("first argument to print must be a string for formatted printing")
	}

	formatValue := g.currentValue

	// Get the actual format string for validation (if it's a literal)
	var formatStr string
	if lit, ok := node.Args[0].(*domain.LiteralExpr); ok {
		if strVal, ok := lit.Value.(string); ok {
			formatStr = strings.Trim(strVal, "\"")
		}
	}

	// Generate remaining arguments
	var argValues []string
	var argTypes []string
	for i := 1; i < len(node.Args); i++ {
		if err := node.Args[i].Accept(g); err != nil {
			return err
		}
		argType := g.getLLVMType(node.Args[i].GetType())
		typeStr := node.Args[i].GetType().String()
		argTypes = append(argTypes, typeStr)
		argValues = append(argValues, fmt.Sprintf("%s %s", argType, g.currentValue))
	}

	// Validate format string if we have it as a literal
	if formatStr != "" {
		if err := g.validateFormatArguments(formatStr, argTypes); err != nil {
			return fmt.Errorf("format string validation failed: %v", err)
		}
	}

	// Generate printf call
	tempReg := fmt.Sprintf("%%temp_%d", g.labelCounter)
	g.labelCounter++

	argsStr := fmt.Sprintf("i8* %s", formatValue)
	if len(argValues) > 0 {
		argsStr += ", " + strings.Join(argValues, ", ")
	}

	g.emit("%s = call i32 (i8*, ...) @printf(%s)", tempReg, argsStr)
	g.currentValue = tempReg
	g.currentType = "i32"

	return nil
}

func (g *Generator) VisitIdentifierExpr(node *domain.IdentifierExpr) error {
	varType := g.getLLVMType(node.GetType())
	align := g.getTypeAlign(node.GetType())

	// Generate a unique temporary register name
	tempReg := fmt.Sprintf("%%temp_%d", g.labelCounter)
	g.labelCounter++

	// Determine whether to use .addr suffix based on whether it's a parameter
	var varName string
	if g.parameters[node.Name] {
		// This is a function parameter, use .addr suffix
		varName = fmt.Sprintf("%%%s.addr", node.Name)
	} else {
		// This is a local variable, use direct name
		varName = fmt.Sprintf("%%%s", node.Name)
	}

	g.emit("%s = load %s, ptr %s, align %d", tempReg, varType, varName, align)

	// Store the result for use by parent expressions
	g.currentValue = tempReg
	g.currentType = varType

	return nil
}

func (g *Generator) VisitLiteralExpr(node *domain.LiteralExpr) error {
	switch node.GetType().String() {
	case "int":
		if val, ok := node.Value.(int64); ok {
			// For integer literals, we can directly use the value in ret statement
			// But if we need a temp variable, we should allocate and store
			g.currentValue = fmt.Sprintf("%d", val)
			g.currentType = "i32"
		} else {
			// Fallback for safety, though parser should ensure int64
			g.currentValue = fmt.Sprintf("%s", node.Value)
			g.currentType = "i32"
		}
	case "double":
		if val, ok := node.Value.(float64); ok {
			g.currentValue = fmt.Sprintf("%f", val)
			g.currentType = "double"
		} else {
			// Fallback for safety
			g.currentValue = fmt.Sprintf("%s", node.Value)
			g.currentType = "double"
		}
	case "string":
		// String literals need special handling
		strValue := strings.Trim(node.Value.(string), "\"")
		length := len(strValue) + 1
		labelName := g.newLabel("str")
		g.emit("@%s = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1", labelName, length, strValue)
		g.currentValue = fmt.Sprintf("getelementptr inbounds ([%d x i8], [%d x i8]* @%s, i32 0, i32 0)", length, length, labelName)
		g.currentType = "i8*"
	}
	return nil
}

func (g *Generator) VisitIndexExpr(node *domain.IndexExpr) error {
	// Not implemented for now
	return fmt.Errorf("index expressions not yet implemented")
}

func (g *Generator) VisitMemberExpr(node *domain.MemberExpr) error {
	// Not implemented for now
	return fmt.Errorf("member expressions not yet implemented")
}

// Helper functions
func (g *Generator) getLLVMType(t domain.Type) string {
	switch t.String() {
	case "int":
		return "i32"
	case "double":
		return "double"
	case "string":
		return "i8*"
	case "bool":
		return "i1"
	case "void":
		return "void"
	default:
		return "i32" // fallback
	}
}

func (g *Generator) getTypeAlign(t domain.Type) int {
	switch t.String() {
	case "int":
		return 4
	case "double":
		return 8
	case "string":
		return 8
	case "bool":
		return 1
	default:
		return 4
	}
}

// parseFormatString analyzes a printf-style format string and returns expected argument types
func (g *Generator) parseFormatString(formatStr string) ([]string, error) {
	var expectedTypes []string

	for i := 0; i < len(formatStr); i++ {
		if formatStr[i] == '%' {
			if i+1 >= len(formatStr) {
				return nil, fmt.Errorf("incomplete format specifier at end of string")
			}

			switch formatStr[i+1] {
			case 'd', 'i':
				expectedTypes = append(expectedTypes, "int")
			case 'f', 'g', 'e':
				expectedTypes = append(expectedTypes, "float") // StaticLang uses float, not double
			case 's':
				expectedTypes = append(expectedTypes, "string")
			case 'c':
				expectedTypes = append(expectedTypes, "int") // char is passed as int
			case '%':
				// %% is escaped %, no argument needed
			default:
				return nil, fmt.Errorf("unsupported format specifier: %%%c", formatStr[i+1])
			}
			i++ // Skip the format character
		}
	}

	return expectedTypes, nil
}

// validateFormatArguments checks if the provided arguments match the format string expectations
func (g *Generator) validateFormatArguments(formatStr string, argTypes []string) error {
	expectedTypes, err := g.parseFormatString(formatStr)
	if err != nil {
		return err
	}

	if len(expectedTypes) != len(argTypes) {
		return fmt.Errorf("format string expects %d arguments, got %d", len(expectedTypes), len(argTypes))
	}

	for i, expected := range expectedTypes {
		actual := argTypes[i]
		if expected != actual {
			return fmt.Errorf("format argument %d: expected %s, got %s", i+1, expected, actual)
		}
	}

	return nil
}
