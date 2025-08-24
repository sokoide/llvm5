// Package application contains factory patterns for compiler components
package application

import (
	"fmt"
	"io"
	"os"

	"github.com/sokoide/llvm5/grammar"
	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/infrastructure"
	"github.com/sokoide/llvm5/internal/interfaces"
	"github.com/sokoide/llvm5/lexer"
	"github.com/sokoide/llvm5/semantic"
)

// CompilerConfig holds configuration for the compiler
type CompilerConfig struct {
	// Component configurations
	UseMockComponents bool
	MemoryManagerType MemoryManagerType
	ErrorReporterType ErrorReporterType

	// Compilation options
	CompilationOptions domain.CompilationOptions

	// Output configuration
	ErrorOutput io.Writer
	Verbose     bool
}

// MemoryManagerType specifies the type of memory manager to use
type MemoryManagerType int

const (
	PooledMemoryManager MemoryManagerType = iota
	CompactMemoryManager
	TrackingMemoryManager
)

// ErrorReporterType specifies the type of error reporter to use
type ErrorReporterType int

const (
	ConsoleErrorReporter ErrorReporterType = iota
	SortedErrorReporter
)

// DefaultCompilerConfig returns a default compiler configuration
func DefaultCompilerConfig() CompilerConfig {
	return CompilerConfig{
		UseMockComponents: false,
		MemoryManagerType: PooledMemoryManager,
		ErrorReporterType: ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}
}

// CompilerFactory creates configured compiler components
type CompilerFactory struct {
	config CompilerConfig
}

// NewCompilerFactory creates a new compiler factory with the given configuration
func NewCompilerFactory(config CompilerConfig) *CompilerFactory {
	return &CompilerFactory{
		config: config,
	}
}

// CreateCompilerPipeline creates a fully configured compiler pipeline
func (factory *CompilerFactory) CreateCompilerPipeline() interfaces.CompilerPipeline {
	pipeline := NewDefaultCompilerPipeline()

	// Set up components
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())
	pipeline.SetOptions(factory.config.CompilationOptions)

	return pipeline
}

// CreateMultiFileCompilerPipeline creates a multi-file compiler pipeline
func (factory *CompilerFactory) CreateMultiFileCompilerPipeline() *MultiFileCompilerPipeline {
	pipeline := NewMultiFileCompilerPipeline()

	// Set up components
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())
	pipeline.SetOptions(factory.config.CompilationOptions)

	return pipeline
}

// CreateLexer creates a lexer component
func (factory *CompilerFactory) CreateLexer() interfaces.Lexer {
	if factory.config.UseMockComponents {
		return NewMockLexer()
	}
	// Return the real lexer implementation
	return lexer.NewLexer()
}

// CreateParser creates a parser component
func (factory *CompilerFactory) CreateParser() interfaces.Parser {
	if factory.config.UseMockComponents {
		return NewMockParser()
	}
	// Use the recursive descent parser instead of yacc-generated parser
	return grammar.NewRecursiveDescentParser()
}

// CreateSemanticAnalyzer creates a semantic analyzer component
func (factory *CompilerFactory) CreateSemanticAnalyzer() interfaces.SemanticAnalyzer {
	if factory.config.UseMockComponents {
		return NewMockSemanticAnalyzer()
	}
	// Return the real semantic analyzer implementation
	return semantic.NewAnalyzer()
}

// CreateCodeGenerator creates a code generator component
func (factory *CompilerFactory) CreateCodeGenerator() interfaces.CodeGenerator {
	if factory.config.UseMockComponents {
		return NewMockCodeGenerator()
	}
	// Return the real LLVM IR generator
	return infrastructure.NewRealLLVMIRGenerator()
}

// CreateErrorReporter creates an error reporter
func (factory *CompilerFactory) CreateErrorReporter() domain.ErrorReporter {
	var baseReporter domain.ErrorReporter

	switch factory.config.ErrorReporterType {
	case ConsoleErrorReporter:
		baseReporter = infrastructure.NewConsoleErrorReporter(factory.config.ErrorOutput)
	case SortedErrorReporter:
		consoleReporter := infrastructure.NewConsoleErrorReporter(factory.config.ErrorOutput)
		baseReporter = infrastructure.NewSortedErrorReporter(consoleReporter)
	default:
		baseReporter = infrastructure.NewConsoleErrorReporter(factory.config.ErrorOutput)
	}

	return baseReporter
}

// CreateTypeRegistry creates a type registry
func (factory *CompilerFactory) CreateTypeRegistry() domain.TypeRegistry {
	return domain.NewDefaultTypeRegistry()
}

// CreateSymbolTable creates a symbol table
func (factory *CompilerFactory) CreateSymbolTable() interfaces.SymbolTable {
	return infrastructure.NewDefaultSymbolTable()
}

// CreateMemoryManager creates a memory manager
func (factory *CompilerFactory) CreateMemoryManager() interfaces.MemoryManager {
	var baseManager interfaces.MemoryManager

	switch factory.config.MemoryManagerType {
	case PooledMemoryManager:
		baseManager = infrastructure.NewPooledMemoryManager()
	case CompactMemoryManager:
		baseManager = infrastructure.NewCompactMemoryManager()
	case TrackingMemoryManager:
		pooled := infrastructure.NewPooledMemoryManager()
		baseManager = infrastructure.NewTrackingMemoryManager(pooled)
	default:
		baseManager = infrastructure.NewPooledMemoryManager()
	}

	return baseManager
}

// CreateLLVMBackend creates an LLVM backend
func (factory *CompilerFactory) CreateLLVMBackend() interfaces.LLVMBackend {
	// For now, always return the mock backend
	// In a real implementation, this would create the actual LLVM backend
	return infrastructure.NewMockLLVMBackend()
}

// Mock implementations for development and testing

// MockLexer provides a mock lexer for testing
type MockLexer struct {
	tokens   []interfaces.Token
	position int
}

func NewMockLexer() *MockLexer {
	return &MockLexer{
		tokens:   make([]interfaces.Token, 0),
		position: 0,
	}
}

func (l *MockLexer) NextToken() interfaces.Token {
	if l.position >= len(l.tokens) {
		return interfaces.Token{
			Type:  interfaces.TokenEOF,
			Value: "",
			Location: domain.SourcePosition{
				Filename: "mock",
				Line:     1,
				Column:   1,
			},
		}
	}

	token := l.tokens[l.position]
	l.position++
	return token
}

func (l *MockLexer) Peek() interfaces.Token {
	if l.position >= len(l.tokens) {
		return interfaces.Token{Type: interfaces.TokenEOF}
	}
	return l.tokens[l.position]
}

func (l *MockLexer) SetInput(filename string, input io.Reader) error {
	// Mock implementation - in real implementation would read from input
	l.tokens = []interfaces.Token{
		{Type: interfaces.TokenFunc, Value: "func", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 1}},
		{Type: interfaces.TokenIdentifier, Value: "main", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 6}},
		{Type: interfaces.TokenLeftParen, Value: "(", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 10}},
		{Type: interfaces.TokenRightParen, Value: ")", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 11}},
		{Type: interfaces.TokenLeftBrace, Value: "{", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 13}},
		{Type: interfaces.TokenRightBrace, Value: "}", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 14}},
		{Type: interfaces.TokenEOF, Value: "", Location: domain.SourcePosition{Filename: filename, Line: 1, Column: 15}},
	}
	l.position = 0
	return nil
}

func (l *MockLexer) GetCurrentPosition() domain.SourcePosition {
	if l.position >= len(l.tokens) {
		return domain.SourcePosition{Filename: "mock", Line: 1, Column: 1}
	}
	return l.tokens[l.position].Location
}

// MockParser provides a mock parser for testing
type MockParser struct {
	errorReporter domain.ErrorReporter
}

func NewMockParser() *MockParser {
	return &MockParser{}
}

func (p *MockParser) Parse(lexer interfaces.Lexer) (*domain.Program, error) {
	// Create a simple mock AST
	program := &domain.Program{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{
				Start: domain.SourcePosition{Filename: "mock", Line: 1, Column: 1},
				End:   domain.SourcePosition{Filename: "mock", Line: 1, Column: 15},
			},
		},
		Declarations: []domain.Declaration{
			&domain.FunctionDecl{
				BaseNode: domain.BaseNode{
					Location: domain.SourceRange{
						Start: domain.SourcePosition{Filename: "mock", Line: 1, Column: 1},
						End:   domain.SourcePosition{Filename: "mock", Line: 1, Column: 15},
					},
				},
				Name:       "main",
				Parameters: make([]domain.Parameter, 0),
				ReturnType: &domain.BasicType{Kind: domain.VoidType},
				Body: &domain.BlockStmt{
					BaseNode: domain.BaseNode{
						Location: domain.SourceRange{
							Start: domain.SourcePosition{Filename: "mock", Line: 1, Column: 13},
							End:   domain.SourcePosition{Filename: "mock", Line: 1, Column: 14},
						},
					},
					Statements: make([]domain.Statement, 0),
				},
			},
		},
	}

	return program, nil
}

func (p *MockParser) SetErrorReporter(reporter domain.ErrorReporter) {
	p.errorReporter = reporter
}

// MockSemanticAnalyzer provides a mock semantic analyzer for testing
type MockSemanticAnalyzer struct {
	typeRegistry  domain.TypeRegistry
	symbolTable   interfaces.SymbolTable
	errorReporter domain.ErrorReporter
}

func NewMockSemanticAnalyzer() *MockSemanticAnalyzer {
	return &MockSemanticAnalyzer{}
}

func (sa *MockSemanticAnalyzer) Analyze(ast *domain.Program) error {
	// Mock semantic analysis - just validate the structure exists
	if ast == nil {
		return fmt.Errorf("AST is nil")
	}
	return nil
}

func (sa *MockSemanticAnalyzer) SetTypeRegistry(registry domain.TypeRegistry) {
	sa.typeRegistry = registry
}

func (sa *MockSemanticAnalyzer) SetSymbolTable(symbolTable interfaces.SymbolTable) {
	sa.symbolTable = symbolTable
}

func (sa *MockSemanticAnalyzer) SetErrorReporter(reporter domain.ErrorReporter) {
	sa.errorReporter = reporter
}

// MockCodeGenerator provides a mock code generator for testing
type MockCodeGenerator struct {
	output        io.Writer
	options       interfaces.CodeGenOptions
	errorReporter domain.ErrorReporter
}

func NewMockCodeGenerator() *MockCodeGenerator {
	return &MockCodeGenerator{}
}

func (cg *MockCodeGenerator) Generate(ast *domain.Program) error {
	if cg.output != nil {
		// Write mock output
		_, err := cg.output.Write([]byte("; Mock generated code\n"))
		return err
	}
	return nil
}

func (cg *MockCodeGenerator) SetOutput(output io.Writer) {
	cg.output = output
}

func (cg *MockCodeGenerator) SetOptions(options interfaces.CodeGenOptions) {
	cg.options = options
}

func (cg *MockCodeGenerator) SetErrorReporter(reporter domain.ErrorReporter) {
	cg.errorReporter = reporter
}
