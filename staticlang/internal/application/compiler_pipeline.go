// Package application contains the main application logic and pipeline
package application

import (
	"fmt"
	"io"

	"github.com/sokoide/llvm5/staticlang/internal/domain"
	"github.com/sokoide/llvm5/staticlang/internal/interfaces"
)

// DefaultCompilerPipeline implements the CompilerPipeline interface
type DefaultCompilerPipeline struct {
	lexer            interfaces.Lexer
	parser           interfaces.Parser
	semanticAnalyzer interfaces.SemanticAnalyzer
	codeGenerator    interfaces.CodeGenerator
	errorReporter    domain.ErrorReporter
	options          domain.CompilationOptions
	typeRegistry     domain.TypeRegistry
	symbolTable      interfaces.SymbolTable
	memoryManager    interfaces.MemoryManager
}

// NewDefaultCompilerPipeline creates a new compiler pipeline with default components
func NewDefaultCompilerPipeline() *DefaultCompilerPipeline {
	return &DefaultCompilerPipeline{
		options: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "",
			WarningsAsErrors:  false,
		},
	}
}

// Compile compiles a source file through the entire pipeline
func (cp *DefaultCompilerPipeline) Compile(filename string, input io.Reader, output io.Writer) error {
	// Validate components are set
	if err := cp.validateComponents(); err != nil {
		return fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Clear previous errors
	cp.errorReporter.Clear()

	// Phase 1: Lexical Analysis
	if err := cp.lexer.SetInput(filename, input); err != nil {
		return fmt.Errorf("failed to set lexer input: %w", err)
	}

	// Phase 2: Syntax Analysis
	ast, err := cp.parser.Parse(cp.lexer)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}

	// Check for parsing errors
	if cp.errorReporter.HasErrors() {
		return fmt.Errorf("compilation failed with %d error(s)", len(cp.errorReporter.GetErrors()))
	}

	// Phase 3: Semantic Analysis
	if err := cp.semanticAnalyzer.Analyze(ast); err != nil {
		return fmt.Errorf("semantic analysis failed: %w", err)
	}

	// Check for semantic errors
	if cp.errorReporter.HasErrors() {
		return fmt.Errorf("compilation failed with %d error(s)", len(cp.errorReporter.GetErrors()))
	}

	// Convert warnings to errors if requested
	if cp.options.WarningsAsErrors && cp.errorReporter.HasWarnings() {
		for _, warning := range cp.errorReporter.GetWarnings() {
			warning.Type = domain.TypeCheckError
			cp.errorReporter.ReportError(warning)
		}
		return fmt.Errorf("compilation failed: warnings treated as errors")
	}

	// Phase 4: Code Generation
	cp.codeGenerator.SetOutput(output)
	cp.codeGenerator.SetOptions(interfaces.CodeGenOptions{
		OptimizationLevel: cp.options.OptimizationLevel,
		DebugInfo:         cp.options.DebugInfo,
		TargetTriple:      cp.options.TargetTriple,
	})

	if err := cp.codeGenerator.Generate(ast); err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	// Check for code generation errors
	if cp.errorReporter.HasErrors() {
		return fmt.Errorf("compilation failed with %d error(s)", len(cp.errorReporter.GetErrors()))
	}

	return nil
}

// SetLexer sets the lexer implementation
func (cp *DefaultCompilerPipeline) SetLexer(lexer interfaces.Lexer) {
	cp.lexer = lexer
}

// SetParser sets the parser implementation
func (cp *DefaultCompilerPipeline) SetParser(parser interfaces.Parser) {
	cp.parser = parser
	if cp.errorReporter != nil {
		parser.SetErrorReporter(cp.errorReporter)
	}
}

// SetSemanticAnalyzer sets the semantic analyzer implementation
func (cp *DefaultCompilerPipeline) SetSemanticAnalyzer(analyzer interfaces.SemanticAnalyzer) {
	cp.semanticAnalyzer = analyzer
	if cp.errorReporter != nil {
		analyzer.SetErrorReporter(cp.errorReporter)
	}
	if cp.typeRegistry != nil {
		analyzer.SetTypeRegistry(cp.typeRegistry)
	}
	if cp.symbolTable != nil {
		analyzer.SetSymbolTable(cp.symbolTable)
	}
}

// SetCodeGenerator sets the code generator implementation
func (cp *DefaultCompilerPipeline) SetCodeGenerator(generator interfaces.CodeGenerator) {
	cp.codeGenerator = generator
	if cp.errorReporter != nil {
		generator.SetErrorReporter(cp.errorReporter)
	}
}

// SetErrorReporter sets the error reporter
func (cp *DefaultCompilerPipeline) SetErrorReporter(reporter domain.ErrorReporter) {
	cp.errorReporter = reporter

	// Propagate to components that are already set
	if cp.parser != nil {
		cp.parser.SetErrorReporter(reporter)
	}
	if cp.semanticAnalyzer != nil {
		cp.semanticAnalyzer.SetErrorReporter(reporter)
	}
	if cp.codeGenerator != nil {
		cp.codeGenerator.SetErrorReporter(reporter)
	}
}

// SetOptions sets compilation options
func (cp *DefaultCompilerPipeline) SetOptions(options domain.CompilationOptions) {
	cp.options = options
}

// SetTypeRegistry sets the type registry
func (cp *DefaultCompilerPipeline) SetTypeRegistry(registry domain.TypeRegistry) {
	cp.typeRegistry = registry
	if cp.semanticAnalyzer != nil {
		cp.semanticAnalyzer.SetTypeRegistry(registry)
	}
}

// SetSymbolTable sets the symbol table
func (cp *DefaultCompilerPipeline) SetSymbolTable(symbolTable interfaces.SymbolTable) {
	cp.symbolTable = symbolTable
	if cp.semanticAnalyzer != nil {
		cp.semanticAnalyzer.SetSymbolTable(symbolTable)
	}
}

// SetMemoryManager sets the memory manager
func (cp *DefaultCompilerPipeline) SetMemoryManager(memoryManager interfaces.MemoryManager) {
	cp.memoryManager = memoryManager
}

// GetStats returns compilation statistics
func (cp *DefaultCompilerPipeline) GetStats() CompilationStats {
	stats := CompilationStats{}

	if cp.errorReporter != nil {
		stats.ErrorCount = len(cp.errorReporter.GetErrors())
		stats.WarningCount = len(cp.errorReporter.GetWarnings())
	}

	if cp.memoryManager != nil {
		memStats := cp.memoryManager.GetStats()
		stats.MemoryUsage = memStats.TotalMemoryUsed
		stats.NodesCreated = memStats.NodesAllocated
	}

	return stats
}

// CompilationStats holds statistics about the compilation process
type CompilationStats struct {
	ErrorCount   int
	WarningCount int
	MemoryUsage  int
	NodesCreated int
}

// validateComponents ensures all required components are set
func (cp *DefaultCompilerPipeline) validateComponents() error {
	if cp.lexer == nil {
		return fmt.Errorf("lexer not set")
	}
	if cp.parser == nil {
		return fmt.Errorf("parser not set")
	}
	if cp.semanticAnalyzer == nil {
		return fmt.Errorf("semantic analyzer not set")
	}
	if cp.codeGenerator == nil {
		return fmt.Errorf("code generator not set")
	}
	if cp.errorReporter == nil {
		return fmt.Errorf("error reporter not set")
	}
	if cp.typeRegistry == nil {
		return fmt.Errorf("type registry not set")
	}
	if cp.symbolTable == nil {
		return fmt.Errorf("symbol table not set")
	}
	return nil
}

// Reset resets the pipeline state for a new compilation
func (cp *DefaultCompilerPipeline) Reset() {
	if cp.errorReporter != nil {
		cp.errorReporter.Clear()
	}
	if cp.symbolTable != nil {
		// Reset symbol table if it has a Reset method
		if resettable, ok := cp.symbolTable.(interface{ Reset() }); ok {
			resettable.Reset()
		}
	}
	if cp.memoryManager != nil {
		cp.memoryManager.FreeAll()
	}
}

// MultiFileCompilerPipeline extends DefaultCompilerPipeline for multi-file compilation
type MultiFileCompilerPipeline struct {
	*DefaultCompilerPipeline
	fileContexts map[string]*domain.CompilationContext
	linkOrder    []string
}

// NewMultiFileCompilerPipeline creates a new multi-file compiler pipeline
func NewMultiFileCompilerPipeline() *MultiFileCompilerPipeline {
	return &MultiFileCompilerPipeline{
		DefaultCompilerPipeline: NewDefaultCompilerPipeline(),
		fileContexts:            make(map[string]*domain.CompilationContext),
		linkOrder:               make([]string, 0),
	}
}

// CompileFiles compiles multiple source files
func (mcp *MultiFileCompilerPipeline) CompileFiles(files map[string]io.Reader, output io.Writer) error {
	// Phase 1: Parse all files and build global symbol table
	asts := make(map[string]*domain.Program)

	for filename, input := range files {
		// Create file-specific context
		context := &domain.CompilationContext{
			SourceFiles:   make(map[string][]byte),
			ErrorReporter: mcp.errorReporter,
			Options:       mcp.options,
		}
		mcp.fileContexts[filename] = context

		// Parse the file
		if err := mcp.lexer.SetInput(filename, input); err != nil {
			return fmt.Errorf("failed to set input for %s: %w", filename, err)
		}

		ast, err := mcp.parser.Parse(mcp.lexer)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", filename, err)
		}

		asts[filename] = ast
		mcp.linkOrder = append(mcp.linkOrder, filename)
	}

	// Check for parsing errors
	if mcp.errorReporter.HasErrors() {
		return fmt.Errorf("parsing failed with %d error(s)", len(mcp.errorReporter.GetErrors()))
	}

	// Phase 2: Semantic analysis across all files
	for _, filename := range mcp.linkOrder {
		ast := asts[filename]
		if err := mcp.semanticAnalyzer.Analyze(ast); err != nil {
			return fmt.Errorf("semantic analysis failed for %s: %w", filename, err)
		}
	}

	// Check for semantic errors
	if mcp.errorReporter.HasErrors() {
		return fmt.Errorf("semantic analysis failed with %d error(s)", len(mcp.errorReporter.GetErrors()))
	}

	// Phase 3: Code generation for all files
	mcp.codeGenerator.SetOutput(output)
	mcp.codeGenerator.SetOptions(interfaces.CodeGenOptions{
		OptimizationLevel: mcp.options.OptimizationLevel,
		DebugInfo:         mcp.options.DebugInfo,
		TargetTriple:      mcp.options.TargetTriple,
	})

	for _, filename := range mcp.linkOrder {
		ast := asts[filename]
		if err := mcp.codeGenerator.Generate(ast); err != nil {
			return fmt.Errorf("code generation failed for %s: %w", filename, err)
		}
	}

	// Check for code generation errors
	if mcp.errorReporter.HasErrors() {
		return fmt.Errorf("code generation failed with %d error(s)", len(mcp.errorReporter.GetErrors()))
	}

	return nil
}

// GetFileContext returns the compilation context for a specific file
func (mcp *MultiFileCompilerPipeline) GetFileContext(filename string) (*domain.CompilationContext, bool) {
	context, exists := mcp.fileContexts[filename]
	return context, exists
}
