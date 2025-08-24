package semantic

import (
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/infrastructure"
)

// TestAnalyzer_Creation tests analyzer creation
func TestAnalyzer_Creation(t *testing.T) {
	analyzer := NewAnalyzer()
	if analyzer == nil {
		t.Error("NewAnalyzer should return non-nil analyzer")
	}

	if analyzer.typeRegistry == nil {
		t.Error("Analyzer should have type registry initialized")
	}
}

// TestAnalyzer_SetTypeRegistry tests type registry setter
func TestAnalyzer_SetTypeRegistry(t *testing.T) {
	analyzer := NewAnalyzer()
	registry := domain.NewDefaultTypeRegistry()

	analyzer.SetTypeRegistry(registry)

	// We can't directly access private fields, but this tests the setter doesn't panic
	if analyzer.typeRegistry != registry {
		t.Error("SetTypeRegistry should set the type registry")
	}
}

// TestAnalyzer_SetSymbolTable tests symbol table setter
func TestAnalyzer_SetSymbolTable(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()

	analyzer.SetSymbolTable(symbolTable)

	// We can't directly access private fields, but this tests the setter doesn't panic
	if analyzer.symbolTable != symbolTable {
		t.Error("SetSymbolTable should set the symbol table")
	}
}

// TestAnalyzer_SetErrorReporter tests error reporter setter
func TestAnalyzer_SetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a simple mock error reporter
	mockReporter := &MockErrorReporter{}

	analyzer.SetErrorReporter(mockReporter)

	// Test doesn't panic and sets the reporter
	if analyzer.errorReporter != mockReporter {
		t.Error("SetErrorReporter should set the error reporter")
	}
}

// TestAnalyzer_AnalyzeEmptyProgram tests analyzing an empty program
func TestAnalyzer_AnalyzeEmptyProgram(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create empty program
	program := &domain.Program{
		Declarations: []domain.Declaration{},
	}

	err := analyzer.Analyze(program)
	if err != nil {
		t.Errorf("Analyzing empty program should not fail: %v", err)
	}
}

// TestAnalyzer_AnalyzeSimpleFunction tests analyzing a simple function
func TestAnalyzer_AnalyzeSimpleFunction(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Create simple function: func main() -> void { }
	voidType := &domain.BasicType{Kind: domain.VoidType}

	functionDecl := &domain.FunctionDecl{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Name:       "main",
		Parameters: []domain.Parameter{},
		ReturnType: voidType,
		Body: &domain.BlockStmt{
			BaseNode: domain.BaseNode{
				Location: domain.SourceRange{},
			},
			Statements: []domain.Statement{},
		},
	}

	program := &domain.Program{
		BaseNode: domain.BaseNode{
			Location: domain.SourceRange{},
		},
		Declarations: []domain.Declaration{functionDecl},
	}

	err := analyzer.Analyze(program)
	if err != nil {
		t.Errorf("Analyzing simple function should not fail: %v", err)
	}
}

// TestAnalyzer_BasicTypeValidation tests basic type validation
func TestAnalyzer_BasicTypeValidation(t *testing.T) {
	analyzer := NewAnalyzer()
	symbolTable := infrastructure.NewSymbolTable()
	analyzer.SetSymbolTable(symbolTable)

	// Test that type registry has basic types
	registry := analyzer.typeRegistry

	intType := registry.GetBuiltinType(domain.IntType)
	if intType == nil {
		t.Error("Type registry should have int type")
	}

	floatType := registry.GetBuiltinType(domain.FloatType)
	if floatType == nil {
		t.Error("Type registry should have float type")
	}

	boolType := registry.GetBuiltinType(domain.BoolType)
	if boolType == nil {
		t.Error("Type registry should have bool type")
	}

	stringType := registry.GetBuiltinType(domain.StringType)
	if stringType == nil {
		t.Error("Type registry should have string type")
	}
}

// MockErrorReporter is a simple mock for testing
type MockErrorReporter struct {
	errors   []domain.CompilerError
	warnings []domain.CompilerError
}

func (m *MockErrorReporter) ReportError(err domain.CompilerError) {
	m.errors = append(m.errors, err)
}

func (m *MockErrorReporter) ReportWarning(warning domain.CompilerError) {
	m.warnings = append(m.warnings, warning)
}

func (m *MockErrorReporter) HasErrors() bool {
	return len(m.errors) > 0
}

func (m *MockErrorReporter) HasWarnings() bool {
	return len(m.warnings) > 0
}

func (m *MockErrorReporter) GetErrors() []domain.CompilerError {
	return m.errors
}

func (m *MockErrorReporter) GetWarnings() []domain.CompilerError {
	return m.warnings
}

func (m *MockErrorReporter) Clear() {
	m.errors = nil
	m.warnings = nil
}
