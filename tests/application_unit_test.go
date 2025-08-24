package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/domain"
)

func TestCompilerFactory(t *testing.T) {
	config := application.DefaultCompilerConfig()
	factory := application.NewCompilerFactory(config)

	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	// Test that we can create all components
	lexer := factory.CreateLexer()
	if lexer == nil {
		t.Error("CreateLexer should not return nil")
	}

	parser := factory.CreateParser()
	if parser == nil {
		t.Error("CreateParser should not return nil")
	}

	analyzer := factory.CreateSemanticAnalyzer()
	if analyzer == nil {
		t.Error("CreateSemanticAnalyzer should not return nil")
	}

	generator := factory.CreateCodeGenerator()
	if generator == nil {
		t.Error("CreateCodeGenerator should not return nil")
	}

	errorReporter := factory.CreateErrorReporter()
	if errorReporter == nil {
		t.Error("CreateErrorReporter should not return nil")
	}

	typeRegistry := factory.CreateTypeRegistry()
	if typeRegistry == nil {
		t.Error("CreateTypeRegistry should not return nil")
	}

	symbolTable := factory.CreateSymbolTable()
	if symbolTable == nil {
		t.Error("CreateSymbolTable should not return nil")
	}

	memoryManager := factory.CreateMemoryManager()
	if memoryManager == nil {
		t.Error("CreateMemoryManager should not return nil")
	}
}

func TestCompilerFactoryMockComponents(t *testing.T) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := application.NewCompilerFactory(config)

	// Create a pipeline with mock components
	pipeline := factory.CreateCompilerPipeline()
	if pipeline == nil {
		t.Fatal("Pipeline should not be nil")
	}

	// Test that mock components work with a simple compilation
	// This is mainly to verify the mock components are properly wired
	mockInput := "func main() { return 0; }"
	inputReader := strings.NewReader(mockInput)
	outputWriter := &strings.Builder{}
	
	err := pipeline.Compile("test.sl", inputReader, outputWriter)
	if err != nil {
		t.Errorf("Mock compilation failed: %v", err)
	}
	
	// Mock should write some result to output
	result := outputWriter.String()
	if result == "" {
		t.Error("Mock compilation should write non-empty result to output")
	}
}

func TestDefaultCompilerConfig(t *testing.T) {
	config := application.DefaultCompilerConfig()

	// Test default values
	if config.UseMockComponents {
		t.Error("Default config should not use mock components")
	}

	if config.MemoryManagerType != application.PooledMemoryManager {
		t.Error("Default config should use PooledMemoryManager")
	}

	if config.ErrorReporterType != application.ConsoleErrorReporter {
		t.Error("Default config should use ConsoleErrorReporter")
	}

	if config.CompilationOptions.OptimizationLevel != 0 {
		t.Error("Default optimization level should be 0")
	}

	if config.CompilationOptions.DebugInfo {
		t.Error("Debug info should be false by default")
	}

	if config.CompilationOptions.WarningsAsErrors {
		t.Error("Warnings as errors should be false by default")
	}

	if config.ErrorOutput != os.Stderr {
		t.Error("Error output should default to stderr")
	}

	if config.Verbose {
		t.Error("Verbose should be false by default")
	}
}

func TestCompilerPipelineConfiguration(t *testing.T) {
	config := application.CompilerConfig{
		UseMockComponents:     false,
		MemoryManagerType:     application.PooledMemoryManager,
		ErrorReporterType:     application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 2,
			DebugInfo:         true,
			TargetTriple:      "x86_64-unknown-linux-gnu",
			OutputPath:        "test.ll",
			WarningsAsErrors:  true,
		},
		ErrorOutput: os.Stderr,
		Verbose:     true,
	}

	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	// Test that configuration is propagated
	// (This mainly tests that the pipeline doesn't crash when configured)
	if pipeline == nil {
		t.Error("Configured pipeline should not be nil")
	}
	
	// Test basic functionality with configured pipeline
	mockInput := "func main() { return 0; }"
	inputReader := strings.NewReader(mockInput)
	outputWriter := &strings.Builder{}
	
	// This should not crash with the configuration
	err := pipeline.Compile("test.sl", inputReader, outputWriter)
	// We don't expect this to succeed with real components in unit tests,
	// but it should not crash due to configuration issues
	_ = err // Ignore error for this configuration test
}

func TestMultiFileCompilerPipeline(t *testing.T) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = true // Use mocks for this test
	factory := application.NewCompilerFactory(config)

	pipeline := factory.CreateMultiFileCompilerPipeline()
	if pipeline == nil {
		t.Fatal("MultiFile pipeline should not be nil")
	}

	// Test that multi-file pipeline can be created and accessed
	// We'll test basic functionality without complex operations
	if pipeline == nil {
		t.Error("MultiFile pipeline creation failed")
	}
	
	// Test basic interface availability (pipeline embeds DefaultCompilerPipeline)
	mockInput := "func main() { return 0; }"
	inputReader := strings.NewReader(mockInput)
	outputWriter := &strings.Builder{}
	
	// Should not crash (even if compilation fails, structure should be sound)
	_ = pipeline.Compile("test.sl", inputReader, outputWriter)
}

func TestMemoryManagerTypes(t *testing.T) {
	testCases := []struct {
		name         string
		managerType  application.MemoryManagerType
		description  string
	}{
		{
			name:         "PooledMemoryManager",
			managerType:  application.PooledMemoryManager,
			description:  "Should create pooled memory manager",
		},
		{
			name:         "CompactMemoryManager", 
			managerType:  application.CompactMemoryManager,
			description:  "Should create compact memory manager",
		},
		{
			name:         "TrackingMemoryManager",
			managerType:  application.TrackingMemoryManager,
			description:  "Should create tracking memory manager",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := application.DefaultCompilerConfig()
			config.MemoryManagerType = tc.managerType
			factory := application.NewCompilerFactory(config)

			manager := factory.CreateMemoryManager()
			if manager == nil {
				t.Errorf("%s: Memory manager should not be nil", tc.description)
			}
		})
	}
}

func TestErrorReporterTypes(t *testing.T) {
	testCases := []struct {
		name         string
		reporterType application.ErrorReporterType
		description  string
	}{
		{
			name:         "ConsoleErrorReporter",
			reporterType: application.ConsoleErrorReporter,
			description:  "Should create console error reporter",
		},
		{
			name:         "SortedErrorReporter",
			reporterType: application.SortedErrorReporter,
			description:  "Should create sorted error reporter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := application.DefaultCompilerConfig()
			config.ErrorReporterType = tc.reporterType
			factory := application.NewCompilerFactory(config)

			reporter := factory.CreateErrorReporter()
			if reporter == nil {
				t.Errorf("%s: Error reporter should not be nil", tc.description)
			}
		})
	}
}

func TestCompilerPipelineBasicOperation(t *testing.T) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = true // Use mocks for predictable behavior
	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	if pipeline == nil {
		t.Fatal("Pipeline should not be nil")
	}

	// Perform a compilation to test basic operation
	mockInput := "func main() { return 0; }"
	inputReader := strings.NewReader(mockInput)
	outputWriter := &strings.Builder{}
	
	err := pipeline.Compile("test.sl", inputReader, outputWriter)
	if err != nil {
		t.Errorf("Mock compilation failed: %v", err)
	}
	
	// Mock should produce some output
	result := outputWriter.String()
	if result == "" {
		t.Error("Mock compilation should produce output")
	}
}

func TestPipelineComponentSetting(t *testing.T) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	// Test that we can set individual components
	// (This verifies the pipeline interface works correctly)
	lexer := factory.CreateLexer()
	parser := factory.CreateParser()
	analyzer := factory.CreateSemanticAnalyzer()
	generator := factory.CreateCodeGenerator()
	reporter := factory.CreateErrorReporter()
	
	// These should not panic
	pipeline.SetLexer(lexer)
	pipeline.SetParser(parser)
	pipeline.SetSemanticAnalyzer(analyzer)
	pipeline.SetCodeGenerator(generator)
	pipeline.SetErrorReporter(reporter)
	
	options := domain.CompilationOptions{
		OptimizationLevel: 1,
		DebugInfo:         false,
	}
	pipeline.SetOptions(options)
	
	// Pipeline should still work after component setting
	mockInput := "func main() { return 0; }"
	inputReader := strings.NewReader(mockInput)
	outputWriter := &strings.Builder{}
	
	err := pipeline.Compile("test.sl", inputReader, outputWriter)
	if err != nil {
		t.Errorf("Compilation after component setting failed: %v", err)
	}
}