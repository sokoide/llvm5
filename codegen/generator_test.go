package codegen

import (
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
)

// TestNewGenerator tests the constructor
func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()
	
	if generator == nil {
		t.Error("NewGenerator should return a non-nil generator")
	}
	
	if generator.labelCounter != 0 {
		t.Error("New generator should have labelCounter initialized to 0")
	}
	
	if generator.parameters == nil {
		t.Error("New generator should have parameters map initialized")
	}
}

// TestEmitMethods tests the emit helper methods
func TestEmitMethods(t *testing.T) {
	generator := NewGenerator()
	
	// Test emit method
	generator.emit("test instruction %s", "arg1")
	output := generator.output.String()
	if !strings.Contains(output, "test instruction arg1") {
		t.Error("emit should format and write the instruction")
	}
	
	// Test emitRaw method
	generator.output.Reset()
	generator.emitRaw("raw text")
	output = generator.output.String()
	if output != "raw text" {
		t.Error("emitRaw should write raw text without formatting")
	}
}

// TestNewLabel tests label generation
func TestNewLabel(t *testing.T) {
	generator := NewGenerator()
	
	label1 := generator.newLabel("test")
	label2 := generator.newLabel("test")
	
	if label1 == label2 {
		t.Error("newLabel should generate unique labels")
	}
	
	if !strings.HasPrefix(label1, "test") {
		t.Error("newLabel should use the provided prefix")
	}
	
	if generator.labelCounter != 2 {
		t.Error("newLabel should increment the counter")
	}
}

// TestGetLLVMType tests LLVM type conversion
func TestGetLLVMType(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		domainType   domain.Type
		expectedLLVM string
	}{
		{domain.NewIntType(), "i32"},
		{domain.NewBoolType(), "i1"},
		{domain.NewStringType(), "i8*"},
		{domain.NewVoidType(), "void"},
	}
	
	for _, test := range tests {
		result := generator.getLLVMType(test.domainType)
		if result != test.expectedLLVM {
			t.Errorf("getLLVMType(%s) = %s, expected %s", 
				test.domainType.String(), result, test.expectedLLVM)
		}
	}
}

// TestGetTypeAlign tests type alignment calculation
func TestGetTypeAlign(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		domainType    domain.Type
		expectedAlign int
	}{
		{domain.NewIntType(), 4},
		{domain.NewBoolType(), 1},
		{domain.NewStringType(), 8},
	}
	
	for _, test := range tests {
		result := generator.getTypeAlign(test.domainType)
		if result != test.expectedAlign {
			t.Errorf("getTypeAlign(%s) = %d, expected %d", 
				test.domainType.String(), result, test.expectedAlign)
		}
	}
}

// TestParseFormatString tests format string parsing
func TestParseFormatString(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		format   string
		expected []string
		hasError bool
	}{
		{"Hello %d", []string{"int"}, false},
		{"Value: %f", []string{"float"}, false},
		{"Name: %s", []string{"string"}, false},
		{"%d + %d = %d", []string{"int", "int", "int"}, false},
		{"Incomplete %", nil, true},
		{"Invalid %z", nil, true},
		{"Escaped %%d", []string{}, false},
	}
	
	for _, test := range tests {
		result, err := generator.parseFormatString(test.format)
		
		if test.hasError {
			if err == nil {
				t.Errorf("parseFormatString(%q) should return error", test.format)
			}
		} else {
			if err != nil {
				t.Errorf("parseFormatString(%q) returned error: %v", test.format, err)
				continue
			}
			
			if len(result) != len(test.expected) {
				t.Errorf("parseFormatString(%q) returned %d types, expected %d", 
					test.format, len(result), len(test.expected))
				continue
			}
			
			for i, expectedType := range test.expected {
				if result[i] != expectedType {
					t.Errorf("parseFormatString(%q)[%d] = %s, expected %s", 
						test.format, i, result[i], expectedType)
				}
			}
		}
	}
}

// TestValidateFormatArguments tests format argument validation
func TestValidateFormatArguments(t *testing.T) {
	generator := NewGenerator()
	
	tests := []struct {
		format   string
		argTypes []string
		hasError bool
	}{
		{"Hello %d", []string{"int"}, false},
		{"Value: %f", []string{"float"}, false},
		{"Wrong type %d", []string{"string"}, true},
		{"Too many args %d", []string{"int", "int"}, true},
		{"Too few args %d %d", []string{"int"}, true},
	}
	
	for _, test := range tests {
		err := generator.validateFormatArguments(test.format, test.argTypes)
		
		if test.hasError {
			if err == nil {
				t.Errorf("validateFormatArguments(%q, %v) should return error", 
					test.format, test.argTypes)
			}
		} else {
			if err != nil {
				t.Errorf("validateFormatArguments(%q, %v) returned error: %v", 
					test.format, test.argTypes, err)
			}
		}
	}
}

// TestGenerateWithoutBackend tests generation without LLVM backend
func TestGenerateWithoutBackend(t *testing.T) {
	generator := NewGenerator()
	// Don't set backend
	
	program := &domain.Program{
		Declarations: []domain.Declaration{
			&domain.FunctionDecl{
				Name:       "main",
				Parameters: []domain.Parameter{},
				ReturnType: domain.NewIntType(),
				Body:       &domain.BlockStmt{Statements: []domain.Statement{}},
			},
		},
	}
	
	result, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Generate should work without backend: %v", err)
	}
	
	if !strings.Contains(result, "define i32 @main()") {
		t.Error("Generated code should contain main function")
	}
}