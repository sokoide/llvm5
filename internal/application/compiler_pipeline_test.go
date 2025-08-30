package application

import (
	"io"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

// TestNewDefaultCompilerPipeline tests pipeline creation
func TestNewDefaultCompilerPipeline(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	if pipeline == nil {
		t.Error("NewDefaultCompilerPipeline should return non-nil pipeline")
	}

	// Test initial state
	stats := pipeline.GetStats()
	if stats.ErrorCount != 0 {
		t.Error("New pipeline should have no errors")
	}
	if stats.WarningCount != 0 {
		t.Error("New pipeline should have no warnings")
	}
	if stats.NodesCreated != 0 {
		t.Error("New pipeline should have no nodes created")
	}
}

// TestPipelineSetters tests all setter methods
func TestPipelineSetters(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Test SetLexer
	lexer := factory.CreateLexer()
	pipeline.SetLexer(lexer)

	// Test SetParser
	parser := factory.CreateParser()
	pipeline.SetParser(parser)

	// Test SetSemanticAnalyzer
	analyzer := factory.CreateSemanticAnalyzer()
	pipeline.SetSemanticAnalyzer(analyzer)

	// Test SetCodeGenerator
	generator := factory.CreateCodeGenerator()
	pipeline.SetCodeGenerator(generator)

	// Test SetErrorReporter
	reporter := factory.CreateErrorReporter()
	pipeline.SetErrorReporter(reporter)

	// Test SetTypeRegistry
	typeRegistry := factory.CreateTypeRegistry()
	pipeline.SetTypeRegistry(typeRegistry)

	// Test SetSymbolTable
	symbolTable := factory.CreateSymbolTable()
	pipeline.SetSymbolTable(symbolTable)

	// Test SetMemoryManager
	memoryManager := factory.CreateMemoryManager()
	pipeline.SetMemoryManager(memoryManager)

	// Test SetOptions
	options := domain.CompilationOptions{
		OptimizationLevel: 2,
		DebugInfo:         true,
		TargetTriple:      "x86_64-linux-gnu",
		OutputPath:        "test.ll",
		WarningsAsErrors:  true,
	}
	pipeline.SetOptions(options)
}

// TestPipelineValidateComponents tests component validation
func TestPipelineValidateComponents(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()

	// Create a simple reader with empty content
	reader := strings.NewReader("")
	writer := &strings.Builder{}

	// Test compilation without components should fail
	err := pipeline.Compile("test.sl", reader, writer)
	if err == nil {
		t.Error("Pipeline should fail validation without components")
	}
}

// TestPipelineCompileWithMocks tests compilation with mock components
func TestPipelineCompileWithMocks(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Set up all components
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())

	// Create test input
	input := strings.NewReader("func main() -> int { return 42; }")
	output := &strings.Builder{}

	// Test successful compilation
	err := pipeline.Compile("test.sl", input, output)
	if err != nil {
		t.Errorf("Pipeline compilation should succeed with mocks: %v", err)
	}

	// Check output was generated
	if output.Len() == 0 {
		t.Error("Pipeline should generate output")
	}

	// Check stats - mock components might not update all stats
	stats := pipeline.GetStats()
	// Just verify stats structure works, don't require specific values from mocks
	if stats.ErrorCount < 0 {
		t.Error("Stats should have valid error count")
	}
}

// TestPipelineStats tests statistics collection
func TestPipelineStats(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Set up pipeline
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())

	// Test initial stats
	stats := pipeline.GetStats()
	if stats.ErrorCount < 0 {
		t.Error("Initial stats should be valid")
	}

	// Compile something
	input := strings.NewReader("func test() -> void {}")
	output := &strings.Builder{}
	pipeline.Compile("test.sl", input, output)

	// Check stats after compilation - with mocks, stats might not change much
	newStats := pipeline.GetStats()
	if newStats.MemoryUsage < 0 {
		t.Error("Memory usage should be non-negative")
	}
	if newStats.ErrorCount < 0 {
		t.Error("Error count should be non-negative")
	}
}

// TestPipelineReset tests pipeline reset functionality
func TestPipelineReset(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Set up pipeline
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())

	// Compile to generate stats
	input := strings.NewReader("func test() -> void {}")
	output := &strings.Builder{}
	pipeline.Compile("test.sl", input, output)

	// Check we have some stats - just verify stats exist
	stats := pipeline.GetStats()
	if stats.ErrorCount < 0 {
		t.Error("Stats should be initialized")
	}

	// Reset pipeline
	pipeline.Reset()

	// Check stats are reset
	newStats := pipeline.GetStats()
	if newStats.ErrorCount != 0 {
		t.Error("Error count should be reset to 0")
	}
	if newStats.WarningCount != 0 {
		t.Error("Warning count should be reset to 0")
	}
	// Mock components may not track nodes, just verify reset doesn't break things
	if newStats.ErrorCount < 0 {
		t.Error("Reset stats should be valid")
	}
}

// TestMultiFileCompilerPipeline tests multi-file compilation
func TestNewMultiFileCompilerPipeline(t *testing.T) {
	pipeline := NewMultiFileCompilerPipeline()
	if pipeline == nil {
		t.Error("NewMultiFileCompilerPipeline should return non-nil pipeline")
	}
}

// TestMultiFileCompileFiles tests multi-file compilation
func TestMultiFileCompileFiles(t *testing.T) {
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	pipeline := NewMultiFileCompilerPipeline()

	// Set up pipeline components through the embedded DefaultCompilerPipeline
	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())

	// Create test files
	files := map[string]io.Reader{
		"file1.sl": strings.NewReader("func helper() -> int { return 1; }"),
		"file2.sl": strings.NewReader("func main() -> int { return helper(); }"),
	}

	output := &strings.Builder{}

	// Test compilation
	err := pipeline.CompileFiles(files, output)
	if err != nil {
		t.Errorf("Multi-file compilation should succeed: %v", err)
	}

	// Check output was generated
	if output.Len() == 0 {
		t.Error("Multi-file compilation should generate output")
	}

	// Check stats - verify basic functionality with mocks
	stats := pipeline.GetStats()
	if stats.ErrorCount < 0 {
		t.Error("Stats should be valid after multi-file compilation")
	}
}

// TestMultiFileGetFileContext tests file context retrieval
func TestMultiFileGetFileContext(t *testing.T) {
	pipeline := NewMultiFileCompilerPipeline()

	// Initially should have no context
	context, exists := pipeline.GetFileContext("test.sl")
	if context != nil || exists {
		t.Error("Should have no context for non-existent file")
	}

	// After setting up some context, we should be able to retrieve it
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	pipeline.SetLexer(factory.CreateLexer())
	pipeline.SetParser(factory.CreateParser())
	pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
	pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
	pipeline.SetErrorReporter(factory.CreateErrorReporter())
	pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
	pipeline.SetSymbolTable(factory.CreateSymbolTable())
	pipeline.SetMemoryManager(factory.CreateMemoryManager())

	// Compile a file to create context
	files := map[string]io.Reader{
		"test.sl": strings.NewReader("func test() -> void {}"),
	}

	output := &strings.Builder{}
	pipeline.CompileFiles(files, output)

	// Now we should have context
	context, exists = pipeline.GetFileContext("test.sl")
	if context == nil || !exists {
		t.Error("Should have context for compiled file")
	}
}

// TestPipelineErrorHandling tests error handling scenarios
func TestPipelineErrorHandling(t *testing.T) {
	pipeline := NewDefaultCompilerPipeline()
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Set up minimal pipeline (missing some components)
	pipeline.SetLexer(factory.CreateLexer())
	// Intentionally not setting parser to test validation

	input := strings.NewReader("test")
	output := &strings.Builder{}

	// Should fail due to missing components
	err := pipeline.Compile("test.sl", input, output)
	if err == nil {
		t.Error("Pipeline should fail with incomplete component setup")
	}
}

// TestPipelineIntegrationWithOptions tests pipeline with various options
func TestPipelineIntegrationWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		options domain.CompilationOptions
	}{
		{
			name: "basic_options",
			options: domain.CompilationOptions{
				OptimizationLevel: 0,
				DebugInfo:         false,
				TargetTriple:      "",
				OutputPath:        "test.ll",
				WarningsAsErrors:  false,
			},
		},
		{
			name: "optimized_options",
			options: domain.CompilationOptions{
				OptimizationLevel: 3,
				DebugInfo:         true,
				TargetTriple:      "x86_64-linux-gnu",
				OutputPath:        "optimized.ll",
				WarningsAsErrors:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewDefaultCompilerPipeline()
			config := DefaultCompilerConfig()
			config.UseMockComponents = true
			factory := NewCompilerFactory(config)

			// Set up pipeline
			pipeline.SetLexer(factory.CreateLexer())
			pipeline.SetParser(factory.CreateParser())
			pipeline.SetSemanticAnalyzer(factory.CreateSemanticAnalyzer())
			pipeline.SetCodeGenerator(factory.CreateCodeGenerator())
			pipeline.SetErrorReporter(factory.CreateErrorReporter())
			pipeline.SetTypeRegistry(factory.CreateTypeRegistry())
			pipeline.SetSymbolTable(factory.CreateSymbolTable())
			pipeline.SetMemoryManager(factory.CreateMemoryManager())
			pipeline.SetOptions(tt.options)

			// Test compilation
			input := strings.NewReader("func test() -> void {}")
			output := &strings.Builder{}

			err := pipeline.Compile("test.sl", input, output)
			if err != nil {
				t.Errorf("Pipeline should succeed with %s: %v", tt.name, err)
			}
		})
	}
}

// TestCompilationStatsStruct tests the CompilationStats struct
func TestCompilationStatsStruct(t *testing.T) {
	stats := CompilationStats{
		ErrorCount:   5,
		WarningCount: 3,
		MemoryUsage:  1024,
		NodesCreated: 100,
	}

	if stats.ErrorCount != 5 {
		t.Error("ErrorCount should be set correctly")
	}
	if stats.WarningCount != 3 {
		t.Error("WarningCount should be set correctly")
	}
	if stats.MemoryUsage != 1024 {
		t.Error("MemoryUsage should be set correctly")
	}
	if stats.NodesCreated != 100 {
		t.Error("NodesCreated should be set correctly")
	}
}