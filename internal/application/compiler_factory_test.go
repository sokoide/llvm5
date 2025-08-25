package application

import (
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

// TestDefaultCompilerConfig tests the default configuration
func TestDefaultCompilerConfig(t *testing.T) {
	config := DefaultCompilerConfig()

	if config.UseMockComponents {
		t.Error("Default config should not use mock components")
	}

	if config.MemoryManagerType != PooledMemoryManager {
		t.Errorf("Expected PooledMemoryManager, got %v", config.MemoryManagerType)
	}

	if config.ErrorReporterType != ConsoleErrorReporter {
		t.Errorf("Expected ConsoleErrorReporter, got %v", config.ErrorReporterType)
	}
}

// TestCompilerFactoryCreation tests factory creation
func TestCompilerFactoryCreation(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	if factory == nil {
		t.Error("NewCompilerFactory should return non-nil factory")
	}

	if factory.config.MemoryManagerType != config.MemoryManagerType {
		t.Error("Factory should preserve config")
	}
}

// TestCreateLexer tests lexer creation
func TestCreateLexer(t *testing.T) {
	tests := []struct {
		name    string
		useMock bool
	}{
		{"real_lexer", false},
		{"mock_lexer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCompilerConfig()
			config.UseMockComponents = tt.useMock
			factory := NewCompilerFactory(config)

			lexer := factory.CreateLexer()
			if lexer == nil {
				t.Error("CreateLexer should return non-nil lexer")
			}

			// Test that mock lexer responds correctly
			if tt.useMock {
				if _, ok := lexer.(*MockLexer); !ok {
					t.Error("Expected MockLexer when UseMockComponents is true")
				}
			}
		})
	}
}

// TestCreateParser tests parser creation
func TestCreateParser(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	parser := factory.CreateParser()
	if parser == nil {
		t.Error("CreateParser should return non-nil parser")
	}
}

// TestCreateSemanticAnalyzer tests semantic analyzer creation
func TestCreateSemanticAnalyzer(t *testing.T) {
	tests := []struct {
		name    string
		useMock bool
	}{
		{"real_analyzer", false},
		{"mock_analyzer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCompilerConfig()
			config.UseMockComponents = tt.useMock
			factory := NewCompilerFactory(config)

			analyzer := factory.CreateSemanticAnalyzer()
			if analyzer == nil {
				t.Error("CreateSemanticAnalyzer should return non-nil analyzer")
			}

			if tt.useMock {
				if _, ok := analyzer.(*MockSemanticAnalyzer); !ok {
					t.Error("Expected MockSemanticAnalyzer when UseMockComponents is true")
				}
			}
		})
	}
}

// TestCreateCodeGenerator tests code generator creation
func TestCreateCodeGenerator(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	generator := factory.CreateCodeGenerator()
	if generator == nil {
		t.Error("CreateCodeGenerator should return non-nil generator")
	}
}

// TestCreateTypeRegistry tests type registry creation
func TestCreateTypeRegistry(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	registry := factory.CreateTypeRegistry()
	if registry == nil {
		t.Error("CreateTypeRegistry should return non-nil registry")
	}

	// Test basic types are available
	intType := registry.GetBuiltinType(domain.IntType)
	if intType == nil {
		t.Error("Type registry should have int type")
	}
}

// TestCreateSymbolTable tests symbol table creation
func TestCreateSymbolTable(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	symbolTable := factory.CreateSymbolTable()
	if symbolTable == nil {
		t.Error("CreateSymbolTable should return non-nil symbol table")
	}

	// Test initial state
	scope := symbolTable.GetCurrentScope()
	if scope.Level != 0 {
		t.Error("Initial scope should be global (level 0)")
	}
}

// TestCreateErrorReporter tests error reporter creation
func TestCreateErrorReporter(t *testing.T) {
	tests := []struct {
		name         string
		reporterType ErrorReporterType
	}{
		{"console_reporter", ConsoleErrorReporter},
		{"sorted_reporter", SortedErrorReporter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCompilerConfig()
			config.ErrorReporterType = tt.reporterType
			factory := NewCompilerFactory(config)

			reporter := factory.CreateErrorReporter()
			if reporter == nil {
				t.Error("CreateErrorReporter should return non-nil reporter")
			}

			// Test basic functionality
			if reporter.HasErrors() {
				t.Error("New reporter should have no errors")
			}
		})
	}
}

// TestCreateMemoryManager tests memory manager creation
func TestCreateMemoryManager(t *testing.T) {
	tests := []struct {
		name        string
		managerType MemoryManagerType
	}{
		{"pooled_manager", PooledMemoryManager},
		{"compact_manager", CompactMemoryManager},
		{"tracking_manager", TrackingMemoryManager},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCompilerConfig()
			config.MemoryManagerType = tt.managerType
			factory := NewCompilerFactory(config)

			manager := factory.CreateMemoryManager()
			if manager == nil {
				t.Error("CreateMemoryManager should return non-nil manager")
			}

			// Test basic functionality
			stats := manager.GetStats()
			if stats.TotalMemoryUsed < 0 {
				t.Error("Memory stats should be valid")
			}
		})
	}
}

// TestCreateLLVMBackend tests LLVM backend creation
func TestCreateLLVMBackend(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	backend := factory.CreateLLVMBackend()
	if backend == nil {
		t.Error("CreateLLVMBackend should return non-nil backend")
	}

	// Test initialization
	err := backend.Initialize("test-module")
	if err != nil {
		t.Errorf("Backend initialization failed: %v", err)
	}

	// Cleanup
	backend.Dispose()
}

// TestCreateCompilerPipeline tests pipeline creation
func TestCreateCompilerPipeline(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	pipeline := factory.CreateCompilerPipeline()
	if pipeline == nil {
		t.Error("CreateCompilerPipeline should return non-nil pipeline")
	}
}

// TestCreateMultiFileCompilerPipeline tests multi-file pipeline creation
func TestCreateMultiFileCompilerPipeline(t *testing.T) {
	config := DefaultCompilerConfig()
	factory := NewCompilerFactory(config)

	pipeline := factory.CreateMultiFileCompilerPipeline()
	if pipeline == nil {
		t.Error("CreateMultiFileCompilerPipeline should return non-nil pipeline")
	}
}

// TestMockComponents tests mock component functionality
func TestMockComponents(t *testing.T) {
	t.Run("mock_lexer", func(t *testing.T) {
		lexer := NewMockLexer()
		if lexer == nil {
			t.Error("NewMockLexer should return non-nil lexer")
		}

		token := lexer.NextToken()
		if token.Type == 0 && token.Value == "" {
			t.Error("Mock lexer should return valid tokens")
		}

		peeked := lexer.Peek()
		if peeked.Type != token.Type {
			t.Error("Peek should return same token as NextToken")
		}
	})

	t.Run("mock_parser", func(t *testing.T) {
		parser := NewMockParser()
		if parser == nil {
			t.Error("NewMockParser should return non-nil parser")
		}

		// Test parsing empty program (using mock lexer)
		lexer := NewMockLexer()
		program, err := parser.Parse(lexer)
		if err != nil {
			t.Errorf("Mock parser should not fail: %v", err)
		}
		if program == nil {
			t.Error("Mock parser should return non-nil program")
		}
	})

	t.Run("mock_semantic_analyzer", func(t *testing.T) {
		analyzer := NewMockSemanticAnalyzer()
		if analyzer == nil {
			t.Error("NewMockSemanticAnalyzer should return non-nil analyzer")
		}

		// Test with empty program
		program := &domain.Program{
			BaseNode:     domain.BaseNode{Location: domain.SourceRange{}},
			Declarations: []domain.Declaration{},
		}

		err := analyzer.Analyze(program)
		if err != nil {
			t.Errorf("Mock analyzer should not fail: %v", err)
		}
	})

	t.Run("mock_code_generator", func(t *testing.T) {
		generator := NewMockCodeGenerator()
		if generator == nil {
			t.Error("NewMockCodeGenerator should return non-nil generator")
		}

		// Test with empty program
		program := &domain.Program{
			BaseNode:     domain.BaseNode{Location: domain.SourceRange{}},
			Declarations: []domain.Declaration{},
		}

		err := generator.Generate(program)
		if err != nil {
			t.Errorf("Mock generator should not fail: %v", err)
		}
	})
}

// TestFactoryIntegration tests component integration through factory
func TestFactoryIntegration(t *testing.T) {
	config := DefaultCompilerConfig()
	config.UseMockComponents = true
	factory := NewCompilerFactory(config)

	// Create all components
	lexer := factory.CreateLexer()
	parser := factory.CreateParser()
	analyzer := factory.CreateSemanticAnalyzer()
	generator := factory.CreateCodeGenerator()
	reporter := factory.CreateErrorReporter()
	typeRegistry := factory.CreateTypeRegistry()
	symbolTable := factory.CreateSymbolTable()

	// Test that components can be configured
	parser.SetErrorReporter(reporter)
	analyzer.SetTypeRegistry(typeRegistry)
	analyzer.SetSymbolTable(symbolTable)
	analyzer.SetErrorReporter(reporter)
	generator.SetErrorReporter(reporter)

	// Test basic workflow simulation
	token := lexer.NextToken()
	if token.Type == 0 && token.Value == "" {
		t.Error("Integrated lexer should produce valid tokens")
	}

	program, err := parser.Parse(lexer)
	if err != nil || program == nil {
		t.Error("Integrated parser should succeed")
	}

	err = analyzer.Analyze(program)
	if err != nil {
		t.Error("Integrated analyzer should succeed")
	}

	err = generator.Generate(program)
	if err != nil {
		t.Error("Integrated generator should succeed")
	}

	if reporter.HasErrors() {
		t.Error("No errors should be reported in successful workflow")
	}
}
