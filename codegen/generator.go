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
}

// NewGenerator creates a new code generator
func NewGenerator() *Generator {
	return &Generator{
		labelCounter: 0,
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
func (g *Generator) VisitProgram(node *domain.Program) error {
	// Emit LLVM IR header
	g.emit("; ModuleID = 'staticlang'")
	g.emit("target datalayout = \"e-m:o-i64:64-f80:128-n8:16:32:64-S128\"")
	g.emit("target triple = \"x86_64-apple-macosx10.15.0\"")
	g.emit("")

	// Declare external functions (built-ins)
	g.emit("; External function declarations")
	g.emit("declare i32 @printf(i8*, ...)")
	g.emit("declare i8* @malloc(i64)")
	g.emit("declare void @free(i8*)")
	g.emit("")

	// Generate global variables
	for _, decl := range node.Declarations {
		if varDecl, ok := decl.(*domain.VarDeclStmt); ok {
			if err := g.generateGlobalVariable(varDecl); err != nil {
				return err
			}
		}
	}

	// Generate string literals
	g.emit("; String literals")
	g.emit("@.str.print = private unnamed_addr constant [4 x i8] c\"%%s\\0A\\00\", align 1")
	g.emit("@.str.print_int = private unnamed_addr constant [4 x i8] c\"%%d\\0A\\00\", align 1")
	g.emit("@.str.print_double = private unnamed_addr constant [4 x i8] c\"%%f\\0A\\00\", align 1")
	g.emit("")

	// Generate functions
	for _, decl := range node.Declarations {
		if funcDecl, ok := decl.(*domain.FunctionDecl); ok {
			if err := funcDecl.Accept(g); err != nil {
				return err
			}
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
		g.emit("store %s %%%s, %s* %%%s.addr, align %d", g.getLLVMType(param.Type), param.Name, g.getLLVMType(param.Type), param.Name, g.getTypeAlign(param.Type))
	}

	// Generate function body
	if err := node.Body.Accept(g); err != nil {
		return err
	}

	// Ensure function has a return statement
	if node.ReturnType.String() == "void" {
		g.emit("ret void")
	} else if node.Name == "main" {
		g.emit("ret i32 0")
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
		// The expression result should be in a temporary register
		// For now, assume it's in %temp_result
		g.emit("store %s %%temp_result, %s* %%%s, align %d", llvmType, llvmType, node.Name, align)
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
		g.emit("store %s %%temp_result, %s* %%%s, align %d", varType, varType, ident.Name, align)
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

	// Branch based on condition
	if node.ElseStmt != nil {
		g.emit("br i1 %%temp_result, label %%%s, label %%%s", thenLabel, elseLabel)
	} else {
		g.emit("br i1 %%temp_result, label %%%s, label %%%s", thenLabel, endLabel)
	}

	// Then block
	g.indentLevel--
	g.emit("%s:", thenLabel)
	g.indentLevel++
	if err := node.ThenStmt.Accept(g); err != nil {
		return err
	}
	g.emit("br label %%%s", endLabel)

	// Else block (if exists)
	if node.ElseStmt != nil {
		g.indentLevel--
		g.emit("%s:", elseLabel)
		g.indentLevel++
		if err := node.ElseStmt.Accept(g); err != nil {
			return err
		}
		g.emit("br label %%%s", endLabel)
	}

	// End block
	g.indentLevel--
	g.emit("%s:", endLabel)
	g.indentLevel++

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
	g.emit("br i1 %%temp_result, label %%%s, label %%%s", bodyLabel, endLabel)

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
		g.emit("br i1 %%temp_result, label %%%s, label %%%s", bodyLabel, endLabel)
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
		g.emit("ret %s %%temp_result", returnType)
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
	leftReg := "%temp_result"

	// Generate right operand
	if err := node.Right.Accept(g); err != nil {
		return err
	}
	rightReg := "%temp_result"

	// Perform operation based on operator
	resultType := g.getLLVMType(node.GetType())

	switch node.Operator {
	case domain.Add:
		if resultType == "i32" {
			g.emit("%%temp_result = add i32 %s, %s", leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%%temp_result = fadd double %s, %s", leftReg, rightReg)
		}
	case domain.Sub:
		if resultType == "i32" {
			g.emit("%%temp_result = sub i32 %s, %s", leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%%temp_result = fsub double %s, %s", leftReg, rightReg)
		}
	case domain.Mul:
		if resultType == "i32" {
			g.emit("%%temp_result = mul i32 %s, %s", leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%%temp_result = fmul double %s, %s", leftReg, rightReg)
		}
	case domain.Div:
		if resultType == "i32" {
			g.emit("%%temp_result = sdiv i32 %s, %s", leftReg, rightReg)
		} else if resultType == "double" {
			g.emit("%%temp_result = fdiv double %s, %s", leftReg, rightReg)
		}
	case domain.Eq:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp eq i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp oeq double %s, %s", leftReg, rightReg)
		}
	case domain.Ne:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp ne i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp one double %s, %s", leftReg, rightReg)
		}
	case domain.Lt:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp slt i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp olt double %s, %s", leftReg, rightReg)
		}
	case domain.Gt:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp sgt i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp ogt double %s, %s", leftReg, rightReg)
		}
	case domain.Le:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp sle i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp ole double %s, %s", leftReg, rightReg)
		}
	case domain.Ge:
		if node.Left.GetType().String() == "int" {
			g.emit("%%temp_result = icmp sge i32 %s, %s", leftReg, rightReg)
		} else if node.Left.GetType().String() == "double" {
			g.emit("%%temp_result = fcmp oge double %s, %s", leftReg, rightReg)
		}
	}

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
	// Special handling for built-in functions (e.g. print)
	if ident, ok := node.Function.(*domain.IdentifierExpr); ok && ident.Name == "print" && len(node.Args) == 1 {
		if err := node.Args[0].Accept(g); err != nil {
			return err
		}

		argType := node.Args[0].GetType().String()
		switch argType {
		case "int":
			g.emit("call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([4 x i8], [4 x i8]* @.str.print_int, i32 0, i32 0), i32 %%temp_result)")
		case "double":
			g.emit("call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([4 x i8], [4 x i8]* @.str.print_double, i32 0, i32 0), double %%temp_result)")
		case "string":
			g.emit("call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([4 x i8], [4 x i8]* @.str.print, i32 0, i32 0), i8* %%temp_result)")
		}
		g.emit("%%temp_result = i32 0") // printf returns an int
		return nil
	}

	// Generate arguments
	argTypes := ""
	args := ""
	for i, arg := range node.Args {
		if err := arg.Accept(g); err != nil {
			return err
		}
		if i > 0 {
			argTypes += ", "
			args += ", "
		}
		argType := g.getLLVMType(arg.GetType())
		argTypes += argType
		args += fmt.Sprintf("%s %%temp_result", argType)
	}

	// Determine function name (assume identifier)
	funcName := "<unknown>"
	if ident, ok := node.Function.(*domain.IdentifierExpr); ok {
		funcName = ident.Name
	}

	// Generate function call
	returnType := g.getLLVMType(node.GetType())
	if returnType == "void" {
		g.emit("call void @%s(%s)", funcName, args)
	} else {
		g.emit("%%temp_result = call %s @%s(%s)", returnType, funcName, args)
	}

	return nil
}

func (g *Generator) VisitIdentifierExpr(node *domain.IdentifierExpr) error {
	varType := g.getLLVMType(node.GetType())
	align := g.getTypeAlign(node.GetType())

	// Check if it's a global or local variable
	// For now, assume local variables use % prefix
	g.emit("%%temp_result = load %s, %s* %%%s, align %d", varType, varType, node.Name, align)

	return nil
}

func (g *Generator) VisitLiteralExpr(node *domain.LiteralExpr) error {
	switch node.GetType().String() {
	case "int":
		g.emit("%%temp_result = i32 %s", node.Value.(string))
	case "double":
		g.emit("%%temp_result = double %s", node.Value.(string))
	case "string":
		// String literals need special handling
		strValue := strings.Trim(node.Value.(string), "\"")
		length := len(strValue) + 1
		labelName := g.newLabel("str")
		g.emit("@%s = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1", labelName, length, strValue)
		g.emit("%%temp_result = i8* getelementptr inbounds ([%d x i8], [%d x i8]* @%s, i32 0, i32 0)", length, length, labelName)
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
