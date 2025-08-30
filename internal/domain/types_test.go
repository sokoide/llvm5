package domain

import (
	"testing"
)

func TestSourcePosition(t *testing.T) {
	pos := SourcePosition{Line: 5, Column: 12}

	if pos.Line != 5 {
		t.Errorf("Expected line 5, got %d", pos.Line)
	}

	if pos.Column != 12 {
		t.Errorf("Expected column 12, got %d", pos.Column)
	}
}

func TestSourceRange(t *testing.T) {
	srcRange := SourceRange{
		Start: SourcePosition{Line: 1, Column: 1},
		End:   SourcePosition{Line: 1, Column: 10},
	}

	if srcRange.Start.Line != 1 || srcRange.Start.Column != 1 {
		t.Error("Start position not set correctly")
	}

	if srcRange.End.Line != 1 || srcRange.End.Column != 10 {
		t.Error("End position not set correctly")
	}
}

func TestCompilationOptions(t *testing.T) {
	options := CompilationOptions{
		OptimizationLevel: 2,
		DebugInfo:         true,
		TargetTriple:      "x86_64-apple-macosx10.15.0",
		OutputPath:        "test_output.ll",
		WarningsAsErrors:  true,
	}

	if options.OptimizationLevel != 2 {
		t.Errorf("Expected optimization level 2, got %d", options.OptimizationLevel)
	}

	if !options.DebugInfo {
		t.Error("Expected DebugInfo to be true")
	}

	if options.TargetTriple != "x86_64-apple-macosx10.15.0" {
		t.Errorf("Expected target triple 'x86_64-apple-macosx10.15.0', got '%s'", options.TargetTriple)
	}

	if options.OutputPath != "test_output.ll" {
		t.Errorf("Expected output path 'test_output.ll', got '%s'", options.OutputPath)
	}

	if !options.WarningsAsErrors {
		t.Error("Expected WarningsAsErrors to be true")
	}
}

func TestIntType(t *testing.T) {
	intType := NewIntType()

	if intType.String() != "int" {
		t.Errorf("Expected type string 'int', got '%s'", intType.String())
	}

	if intType.GetSize() != 8 {
		t.Errorf("Expected size 8, got %d", intType.GetSize())
	}

	// Test type equality
	intType2 := NewIntType()
	if !intType.Equals(intType2) {
		t.Error("Int types should be equal")
	}

	if !intType.IsAssignableFrom(intType2) {
		t.Error("Int type should be assignable from int type")
	}

	// Test assignment compatibility
	floatType := NewFloatType()
	if intType.IsAssignableFrom(floatType) {
		t.Error("Int type should not be assignable from float type")
	}
}

func TestFloatType(t *testing.T) {
	floatType := NewFloatType()

	if floatType.String() != "float" {
		t.Errorf("Expected type string 'float', got '%s'", floatType.String())
	}

	if floatType.GetSize() != 8 {
		t.Errorf("Expected size 8, got %d", floatType.GetSize())
	}

	// Test type equality
	floatType2 := NewFloatType()
	if !floatType.Equals(floatType2) {
		t.Error("Float types should be equal")
	}

	if !floatType.IsAssignableFrom(floatType2) {
		t.Error("Float type should be assignable from float type")
	}

	// Test numeric type compatibility
	intType := NewIntType()
	if floatType.IsAssignableFrom(intType) {
		t.Error("Float type should not be assignable from int type (different sizes)")
	}
}

func TestStringType(t *testing.T) {
	stringType := NewStringType()

	if stringType.String() != "string" {
		t.Errorf("Expected type string 'string', got '%s'", stringType.String())
	}

	if stringType.GetSize() != 8 {
		t.Errorf("Expected size 8, got %d", stringType.GetSize())
	}

	stringType2 := NewStringType()
	if !stringType.Equals(stringType2) {
		t.Error("String types should be equal")
	}

	if !stringType.IsAssignableFrom(stringType2) {
		t.Error("String type should be assignable from string type")
	}

	// Test incompatibility with other types
	intType := NewIntType()
	if stringType.IsAssignableFrom(intType) {
		t.Error("String type should not be assignable from int type")
	}
}

func TestBoolType(t *testing.T) {
	boolType := NewBoolType()

	if boolType.String() != "bool" {
		t.Errorf("Expected type string 'bool', got '%s'", boolType.String())
	}

	if boolType.GetSize() != 1 {
		t.Errorf("Expected size 1, got %d", boolType.GetSize())
	}

	boolType2 := NewBoolType()
	if !boolType.Equals(boolType2) {
		t.Error("Bool types should be equal")
	}

	if !boolType.IsAssignableFrom(boolType2) {
		t.Error("Bool type should be assignable from bool type")
	}
}

func TestVoidType(t *testing.T) {
	voidType := NewVoidType()

	if voidType.String() != "void" {
		t.Errorf("Expected type string 'void', got '%s'", voidType.String())
	}

	if voidType.GetSize() != 0 {
		t.Errorf("Expected size 0, got %d", voidType.GetSize())
	}
}

func TestArrayType(t *testing.T) {
	elementType := NewIntType()

	// Test fixed-size array
	arrayType := &ArrayType{
		ElementType: elementType,
		Size:        5,
	}

	if arrayType.String() != "[5]int" {
		t.Errorf("Expected type string '[5]int', got '%s'", arrayType.String())
	}

	if arrayType.GetSize() != 40 { // 5 elements * 8 bytes each
		t.Errorf("Expected size 40, got %d", arrayType.GetSize())
	}

	// Test dynamic array
	dynamicArray := &ArrayType{
		ElementType: elementType,
		Size:        -1,
	}

	if dynamicArray.String() != "[]int" {
		t.Errorf("Expected type string '[]int', got '%s'", dynamicArray.String())
	}

	if dynamicArray.GetSize() != 8 { // Pointer size
		t.Errorf("Expected size 8, got %d", dynamicArray.GetSize())
	}

	// Test array type compatibility
	arrayType2 := &ArrayType{
		ElementType: elementType,
		Size:        5,
	}

	if !arrayType.Equals(arrayType2) {
		t.Error("Array types with same element type and size should be equal")
	}

	if !arrayType.IsAssignableFrom(arrayType2) {
		t.Error("Compatible array types should be assignable")
	}

	// Test incompatibility
	floatArray := &ArrayType{
		ElementType: NewFloatType(),
		Size:        5,
	}

	if arrayType.IsAssignableFrom(floatArray) {
		t.Error("Int array should not be assignable from float array")
	}
}

func TestStructType(t *testing.T) {
	fields := map[string]Type{
		"name":  NewStringType(),
		"age":   NewIntType(),
		"score": NewFloatType(),
	}

	structType := &StructType{
		Name:   "", // Empty name to test field details format
		Fields: fields,
		Order:  []string{"name", "age", "score"}, // Define field order
	}

	if structType.String() != "struct{name string, age int, score float}" {
		t.Errorf("Expected struct string format, got '%s'", structType.String())
	}

	// Test field access
	nameField, exists := structType.GetField("name")
	if !exists {
		t.Error("Field 'name' should exist")
	}
	if !nameField.Equals(NewStringType()) {
		t.Error("Name field should be string type")
	}

	// Test non-existent field
	_, exists = structType.GetField("nonexistent")
	if exists {
		t.Error("Non-existent field should not be found")
	}

	// Test struct type equality
	structType2 := &StructType{
		Name:   "", // Empty name to test field details format
		Fields: fields,
		Order:  []string{"name", "age", "score"}, // Define field order
	}

	if !structType.Equals(structType2) {
		t.Error("Struct types with identical fields should be equal")
	}

	if !structType.IsAssignableFrom(structType2) {
		t.Error("Compatible struct types should be assignable")
	}
}

func TestTypeOperationsOnDifferentTypes(t *testing.T) {
	intType := NewIntType()
	floatType := NewFloatType()
	stringType := NewStringType()

	// Test cross-type incompatibilities
	testCases := []struct {
		type1     Type
		type2     Type
		shouldBeCompatible bool
		description string
	}{
		{intType, floatType, false, "int should not be assignable from float"},
		{floatType, stringType, false, "float should not be assignable from string"},
		{stringType, intType, false, "string should not be assignable from int"},
		{intType, intType, true, "int should be assignable from int"},
		{floatType, floatType, true, "float should be assignable from float"},
		{stringType, stringType, true, "string should be assignable from string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.type1.IsAssignableFrom(tc.type2)
			if result != tc.shouldBeCompatible {
				t.Errorf("Expected assignable=%t, got %t for %s", tc.shouldBeCompatible, result, tc.description)
			}
		})
	}
}

func TestErrorType(t *testing.T) {
	errorType := &TypeError{Message: "Undefined variable 'x'"}

	expectedMsg := "Undefined variable 'x'"
	if errorType.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, errorType.Message)
	}

	if errorType.String() != "<error: "+expectedMsg+">" {
		t.Errorf("Expected error type string format, got '%s'", errorType.String())
	}

	// Error types don't have a meaningful size
	if errorType.GetSize() != 0 {
		t.Errorf("Expected error type size 0, got %d", errorType.GetSize())
	}

	// Error types are not assignable to/from other types
	intType := NewIntType()
	if errorType.IsAssignableFrom(intType) {
		t.Error("Error type should not be assignable from int type")
	}
}