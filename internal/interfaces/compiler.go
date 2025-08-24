// Package interfaces defines the core interfaces for the StaticLang compiler components
package interfaces

import (
	"io"

	"github.com/sokoide/llvm5/internal/domain"
)

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Value    string
	Location domain.SourcePosition
}

type TokenType int

const (
	// Literals
	TokenInt TokenType = iota
	TokenFloat
	TokenString
	TokenBool
	TokenIdentifier

	// Keywords
	TokenFunc
	TokenStruct
	TokenVar
	TokenIf
	TokenElse
	TokenWhile
	TokenFor
	TokenReturn
	TokenTrue
	TokenFalse

	// Operators
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAnd
	TokenOr
	TokenNot
	TokenAssign

	// Delimiters
	TokenLeftParen
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenSemicolon
	TokenComma
	TokenDot
	TokenColon
	TokenArrow

	// Special
	TokenEOF
	TokenError
)

// Lexer interface defines the lexical analyzer
type Lexer interface {
	// NextToken returns the next token from the input
	NextToken() Token

	// Peek returns the next token without consuming it
	Peek() Token

	// SetInput sets the input source for lexing
	SetInput(filename string, input io.Reader) error

	// GetCurrentPosition returns the current position in the input
	GetCurrentPosition() domain.SourcePosition
}

// Parser interface defines the syntax analyzer
type Parser interface {
	// Parse parses the input and returns an AST
	Parse(lexer Lexer) (*domain.Program, error)

	// SetErrorReporter sets the error reporter for the parser
	SetErrorReporter(reporter domain.ErrorReporter)
}

// SemanticAnalyzer interface defines the semantic analyzer
type SemanticAnalyzer interface {
	// Analyze performs semantic analysis on the AST
	Analyze(ast *domain.Program) error

	// SetTypeRegistry sets the type registry
	SetTypeRegistry(registry domain.TypeRegistry)

	// SetSymbolTable sets the symbol table
	SetSymbolTable(symbolTable SymbolTable)

	// SetErrorReporter sets the error reporter
	SetErrorReporter(reporter domain.ErrorReporter)
}

// CodeGenerator interface defines the code generator
type CodeGenerator interface {
	// Generate generates code for the given AST
	Generate(ast *domain.Program) error

	// SetOutput sets the output destination
	SetOutput(output io.Writer)

	// SetOptions sets code generation options
	SetOptions(options CodeGenOptions)

	// SetErrorReporter sets the error reporter
	SetErrorReporter(reporter domain.ErrorReporter)
}

// CodeGenOptions holds code generation configuration
type CodeGenOptions struct {
	OptimizationLevel int
	DebugInfo         bool
	TargetTriple      string
}

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name     string
	Type     domain.Type
	Kind     SymbolKind
	Location domain.SourceRange
	Scope    *Scope
}

type SymbolKind int

const (
	VariableSymbol SymbolKind = iota
	FunctionSymbol
	ParameterSymbol
	StructSymbol
	FieldSymbol
)

// Scope represents a lexical scope
type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
	Level    int
}

// SymbolTable interface defines symbol table operations
type SymbolTable interface {
	// EnterScope creates a new scope
	EnterScope() *Scope

	// ExitScope exits the current scope
	ExitScope()

	// GetCurrentScope returns the current scope
	GetCurrentScope() *Scope

	// DeclareSymbol declares a symbol in the current scope
	DeclareSymbol(name string, symbolType domain.Type, kind SymbolKind, location domain.SourceRange) (*Symbol, error)

	// LookupSymbol looks up a symbol in the current scope chain
	LookupSymbol(name string) (*Symbol, bool)

	// LookupSymbolInScope looks up a symbol in a specific scope only
	LookupSymbolInScope(name string, scope *Scope) (*Symbol, bool)
}

// MemoryManager interface defines memory management operations
type MemoryManager interface {
	// AllocateNode allocates memory for an AST node
	AllocateNode(nodeType string, size int) (interface{}, error)

	// AllocateString allocates memory for a string
	AllocateString(s string) (interface{}, error)

	// FreeAll frees all allocated memory
	FreeAll()

	// GetStats returns memory usage statistics
	GetStats() MemoryStats
}

type MemoryStats struct {
	NodesAllocated   int
	StringsAllocated int
	TotalMemoryUsed  int
}

// CompilerPipeline interface defines the overall compilation process
type CompilerPipeline interface {
	// Compile compiles a source file through the entire pipeline
	Compile(filename string, input io.Reader, output io.Writer) error

	// SetLexer sets the lexer implementation
	SetLexer(lexer Lexer)

	// SetParser sets the parser implementation
	SetParser(parser Parser)

	// SetSemanticAnalyzer sets the semantic analyzer implementation
	SetSemanticAnalyzer(analyzer SemanticAnalyzer)

	// SetCodeGenerator sets the code generator implementation
	SetCodeGenerator(generator CodeGenerator)

	// SetErrorReporter sets the error reporter
	SetErrorReporter(reporter domain.ErrorReporter)

	// SetOptions sets compilation options
	SetOptions(options domain.CompilationOptions)
}

// LLVMBackend interface defines the LLVM abstraction layer
type LLVMBackend interface {
	// Initialize initializes the LLVM backend
	Initialize(targetTriple string) error

	// CreateModule creates a new LLVM module
	CreateModule(name string) (LLVMModule, error)

	// Optimize optimizes the given module
	Optimize(module LLVMModule, level int) error

	// EmitObject emits object code to the writer
	EmitObject(module LLVMModule, output io.Writer) error

	// EmitAssembly emits assembly code to the writer
	EmitAssembly(module LLVMModule, output io.Writer) error

	// Dispose disposes of resources
	Dispose()
}

// LLVMModule interface represents an LLVM module
type LLVMModule interface {
	// CreateFunction creates a new function in the module
	CreateFunction(name string, funcType domain.Type) (LLVMFunction, error)

	// CreateGlobalVariable creates a global variable
	CreateGlobalVariable(name string, varType domain.Type) (LLVMValue, error)

	// CreateStruct creates a struct type
	CreateStruct(name string, structType *domain.StructType) (LLVMType, error)

	// GetFunction gets a function by name
	GetFunction(name string) (LLVMFunction, bool)

	// Verify verifies the module
	Verify() error

	// Print prints the module IR
	Print(output io.Writer)

	// Dispose disposes of the module
	Dispose()
}

// LLVMFunction interface represents an LLVM function
type LLVMFunction interface {
	// CreateBasicBlock creates a basic block in the function
	CreateBasicBlock(name string) LLVMBasicBlock

	// GetParameter gets a parameter by index
	GetParameter(index int) LLVMValue

	// GetParameterCount gets the number of parameters
	GetParameterCount() int

	// SetName sets the function name
	SetName(name string)
}

// LLVMBasicBlock interface represents an LLVM basic block
type LLVMBasicBlock interface {
	// GetName gets the block name
	GetName() string

	// IsTerminated checks if the block is terminated
	IsTerminated() bool
}

// LLVMValue interface represents an LLVM value
type LLVMValue interface {
	// GetType gets the value type
	GetType() LLVMType

	// SetName sets the value name
	SetName(name string)

	// GetName gets the value name
	GetName() string
}

// LLVMType interface represents an LLVM type
type LLVMType interface {
	// IsInteger checks if the type is an integer
	IsInteger() bool

	// IsFloat checks if the type is a float
	IsFloat() bool

	// IsPointer checks if the type is a pointer
	IsPointer() bool

	// IsStruct checks if the type is a struct
	IsStruct() bool
}

// LLVMBuilder interface represents an LLVM IR builder
type LLVMBuilder interface {
	// PositionAtEnd positions the builder at the end of a basic block
	PositionAtEnd(block LLVMBasicBlock)

	// CreateAlloca creates an alloca instruction
	CreateAlloca(t LLVMType, name string) LLVMValue

	// CreateStore creates a store instruction
	CreateStore(value, ptr LLVMValue) LLVMValue

	// CreateLoad creates a load instruction
	CreateLoad(ptr LLVMValue, name string) LLVMValue

	// CreateAdd creates an add instruction
	CreateAdd(lhs, rhs LLVMValue, name string) LLVMValue

	// CreateSub creates a sub instruction
	CreateSub(lhs, rhs LLVMValue, name string) LLVMValue

	// CreateMul creates a mul instruction
	CreateMul(lhs, rhs LLVMValue, name string) LLVMValue

	// CreateSDiv creates a signed division instruction
	CreateSDiv(lhs, rhs LLVMValue, name string) LLVMValue

	// CreateICmp creates an integer comparison
	CreateICmp(pred IntPredicate, lhs, rhs LLVMValue, name string) LLVMValue

	// CreateFCmp creates a float comparison
	CreateFCmp(pred FloatPredicate, lhs, rhs LLVMValue, name string) LLVMValue

	// CreateBr creates an unconditional branch
	CreateBr(dest LLVMBasicBlock) LLVMValue

	// CreateCondBr creates a conditional branch
	CreateCondBr(cond LLVMValue, then, else_ LLVMBasicBlock) LLVMValue

	// CreateRet creates a return instruction
	CreateRet(value LLVMValue) LLVMValue

	// CreateRetVoid creates a void return instruction
	CreateRetVoid() LLVMValue

	// CreateCall creates a function call
	CreateCall(fn LLVMFunction, args []LLVMValue, name string) LLVMValue

	// CreateGEP creates a getelementptr instruction
	CreateGEP(ptr LLVMValue, indices []LLVMValue, name string) LLVMValue

	// Dispose disposes of the builder
	Dispose()
}

// IntPredicate represents integer comparison predicates
type IntPredicate int

const (
	IntEQ  IntPredicate = iota // equal
	IntNE                      // not equal
	IntSLT                     // signed less than
	IntSLE                     // signed less or equal
	IntSGT                     // signed greater than
	IntSGE                     // signed greater or equal
)

// FloatPredicate represents float comparison predicates
type FloatPredicate int

const (
	FloatOEQ FloatPredicate = iota // ordered and equal
	FloatONE                       // ordered and not equal
	FloatOLT                       // ordered and less than
	FloatOLE                       // ordered and less than or equal
	FloatOGT                       // ordered and greater than
	FloatOGE                       // ordered and greater than or equal
)
