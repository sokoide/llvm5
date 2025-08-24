package domain

import (
	"testing"
)

// TestBasicType_String tests string representation of basic types
func TestBasicType_String(t *testing.T) {
	tests := []struct {
		kind     BasicTypeKind
		expected string
	}{
		{IntType, "int"},
		{FloatType, "float"},
		{BoolType, "bool"},
		{StringType, "string"},
		{VoidType, "void"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			basicType := &BasicType{Kind: tt.kind}
			if got := basicType.String(); got != tt.expected {
				t.Errorf("BasicType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestBasicType_Equals tests type equality
func TestBasicType_Equals(t *testing.T) {
	intType1 := &BasicType{Kind: IntType}
	intType2 := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}

	// Same types should be equal
	if !intType1.Equals(intType2) {
		t.Error("Same basic types should be equal")
	}

	// Different types should not be equal
	if intType1.Equals(floatType) {
		t.Error("Different basic types should not be equal")
	}

	// Test with different type interface
	arrayType := &ArrayType{ElementType: intType1, Size: 10}
	if intType1.Equals(arrayType) {
		t.Error("BasicType should not equal ArrayType")
	}
}

// TestBasicType_IsAssignableFrom tests type assignment compatibility
func TestBasicType_IsAssignableFrom(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}

	// Type should be assignable from itself
	if !intType.IsAssignableFrom(intType) {
		t.Error("Type should be assignable from itself")
	}

	// Different types should not be assignable (strict typing)
	if intType.IsAssignableFrom(floatType) {
		t.Error("Int should not be assignable from float without explicit conversion")
	}
}

// TestBasicType_GetSize tests type size calculation
func TestBasicType_GetSize(t *testing.T) {
	tests := []struct {
		kind         BasicTypeKind
		expectedSize int
	}{
		{IntType, 8},    // 64-bit int
		{FloatType, 8},  // 64-bit float
		{BoolType, 1},   // 1 byte bool
		{StringType, 8}, // pointer size
		{VoidType, 0},   // void has no size
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.kind)), func(t *testing.T) {
			basicType := &BasicType{Kind: tt.kind}
			if got := basicType.GetSize(); got != tt.expectedSize {
				t.Errorf("BasicType.GetSize() = %v, want %v", got, tt.expectedSize)
			}
		})
	}
}

// TestArrayType_String tests array type string representation
func TestArrayType_String(t *testing.T) {
	intType := &BasicType{Kind: IntType}

	tests := []struct {
		name     string
		array    *ArrayType
		expected string
	}{
		{
			name:     "dynamic_array",
			array:    &ArrayType{ElementType: intType, Size: -1},
			expected: "[]int",
		},
		{
			name:     "fixed_array",
			array:    &ArrayType{ElementType: intType, Size: 10},
			expected: "[10]int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.array.String(); got != tt.expected {
				t.Errorf("ArrayType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestArrayType_Equals tests array type equality
func TestArrayType_Equals(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}

	intArray1 := &ArrayType{ElementType: intType, Size: 10}
	intArray2 := &ArrayType{ElementType: intType, Size: 10}
	intArrayDynamic := &ArrayType{ElementType: intType, Size: -1}
	floatArray := &ArrayType{ElementType: floatType, Size: 10}

	// Same array types should be equal
	if !intArray1.Equals(intArray2) {
		t.Error("Same array types should be equal")
	}

	// Different sizes should not be equal
	if intArray1.Equals(intArrayDynamic) {
		t.Error("Fixed and dynamic arrays should not be equal")
	}

	// Different element types should not be equal
	if intArray1.Equals(floatArray) {
		t.Error("Arrays with different element types should not be equal")
	}

	// Should not equal basic type
	if intArray1.Equals(intType) {
		t.Error("ArrayType should not equal BasicType")
	}
}

// TestArrayType_IsAssignableFrom tests array assignment compatibility
func TestArrayType_IsAssignableFrom(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}

	intArray10 := &ArrayType{ElementType: intType, Size: 10}
	intArray20 := &ArrayType{ElementType: intType, Size: 20}
	intArrayDynamic := &ArrayType{ElementType: intType, Size: -1}
	floatArray10 := &ArrayType{ElementType: floatType, Size: 10}

	// Same type should be assignable
	if !intArray10.IsAssignableFrom(intArray10) {
		t.Error("Array should be assignable from itself")
	}

	// Different sizes should not be assignable
	if intArray10.IsAssignableFrom(intArray20) {
		t.Error("Arrays with different sizes should not be assignable")
	}

	// Dynamic arrays should be assignable from fixed arrays of same element type
	if !intArrayDynamic.IsAssignableFrom(intArray10) {
		t.Error("Dynamic array should be assignable from fixed array of same element type")
	}

	// Different element types should not be assignable
	if intArray10.IsAssignableFrom(floatArray10) {
		t.Error("Arrays with different element types should not be assignable")
	}
}

// TestArrayType_GetSize tests array size calculation
func TestArrayType_GetSize(t *testing.T) {
	intType := &BasicType{Kind: IntType}

	tests := []struct {
		name     string
		array    *ArrayType
		expected int
	}{
		{
			name:     "fixed_array",
			array:    &ArrayType{ElementType: intType, Size: 10},
			expected: 10 * 8, // 10 elements * 8 bytes each
		},
		{
			name:     "dynamic_array",
			array:    &ArrayType{ElementType: intType, Size: -1},
			expected: 8, // pointer size for dynamic arrays
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.array.GetSize(); got != tt.expected {
				t.Errorf("ArrayType.GetSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestStructType_String tests struct type string representation
func TestStructType_String(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	stringType := &BasicType{Kind: StringType}

	structType := &StructType{
		Name: "Person",
		Fields: map[string]Type{
			"id":   intType,
			"name": stringType,
		},
		Order: []string{"id", "name"},
	}
	expected := "Person"

	if got := structType.String(); got != expected {
		t.Errorf("StructType.String() = %v, want %v", got, expected)
	}
}

// TestStructType_GetField tests struct field lookup
func TestStructType_GetField(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	stringType := &BasicType{Kind: StringType}

	structType := &StructType{
		Name: "Person",
		Fields: map[string]Type{
			"id":   intType,
			"name": stringType,
		},
		Order: []string{"id", "name"},
	}

	// Test existing field
	fieldType, found := structType.GetField("name")
	if !found {
		t.Error("Should find existing field 'name'")
	}
	if fieldType != stringType {
		t.Error("Field 'name' should have string type")
	}

	// Test non-existing field
	_, found = structType.GetField("nonexistent")
	if found {
		t.Error("Should not find non-existent field")
	}
}

// TestStructType_Equals tests struct type equality
func TestStructType_Equals(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	stringType := &BasicType{Kind: StringType}

	struct1 := &StructType{
		Name: "Person",
		Fields: map[string]Type{
			"id":   intType,
			"name": stringType,
		},
		Order: []string{"id", "name"},
	}

	struct2 := &StructType{
		Name: "Person",
		Fields: map[string]Type{
			"id":   intType,
			"name": stringType,
		},
		Order: []string{"id", "name"},
	}

	structDifferentName := &StructType{
		Name: "User",
		Fields: map[string]Type{
			"id":   intType,
			"name": stringType,
		},
		Order: []string{"id", "name"},
	}

	// Same structs should be equal
	if !struct1.Equals(struct2) {
		t.Error("Same struct types should be equal")
	}

	// Different names should not be equal
	if struct1.Equals(structDifferentName) {
		t.Error("Structs with different names should not be equal")
	}

	// Should not equal basic type
	if struct1.Equals(intType) {
		t.Error("StructType should not equal BasicType")
	}
}

// TestFunctionType_String tests function type string representation
func TestFunctionType_String(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	stringType := &BasicType{Kind: StringType}
	voidType := &BasicType{Kind: VoidType}

	tests := []struct {
		name     string
		funcType *FunctionType
		expected string
	}{
		{
			name: "no_params_void_return",
			funcType: &FunctionType{
				ParameterTypes: []Type{},
				ReturnType:     voidType,
			},
			expected: "func() void",
		},
		{
			name: "single_param",
			funcType: &FunctionType{
				ParameterTypes: []Type{intType},
				ReturnType:     intType,
			},
			expected: "func(int) int",
		},
		{
			name: "multiple_params",
			funcType: &FunctionType{
				ParameterTypes: []Type{intType, stringType},
				ReturnType:     stringType,
			},
			expected: "func(int, string) string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.funcType.String(); got != tt.expected {
				t.Errorf("FunctionType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFunctionType_Equals tests function type equality
func TestFunctionType_Equals(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	stringType := &BasicType{Kind: StringType}
	voidType := &BasicType{Kind: VoidType}

	func1 := &FunctionType{
		ParameterTypes: []Type{intType, stringType},
		ReturnType:     voidType,
	}

	func2 := &FunctionType{
		ParameterTypes: []Type{intType, stringType},
		ReturnType:     voidType,
	}

	funcDifferentParams := &FunctionType{
		ParameterTypes: []Type{stringType, intType}, // reordered
		ReturnType:     voidType,
	}

	funcDifferentReturn := &FunctionType{
		ParameterTypes: []Type{intType, stringType},
		ReturnType:     intType, // different return type
	}

	// Same function types should be equal
	if !func1.Equals(func2) {
		t.Error("Same function types should be equal")
	}

	// Different parameter order should not be equal
	if func1.Equals(funcDifferentParams) {
		t.Error("Function types with different parameter order should not be equal")
	}

	// Different return types should not be equal
	if func1.Equals(funcDifferentReturn) {
		t.Error("Function types with different return types should not be equal")
	}

	// Should not equal basic type
	if func1.Equals(intType) {
		t.Error("FunctionType should not equal BasicType")
	}
}

// TestTypeRegistry_DefaultTypes tests default type registry creation
func TestTypeRegistry_DefaultTypes(t *testing.T) {
	registry := NewDefaultTypeRegistry()

	// Test that builtin types are registered
	intType := registry.GetBuiltinType(IntType)
	if intType == nil {
		t.Error("Default registry should have int type")
	}

	floatType := registry.GetBuiltinType(FloatType)
	if floatType == nil {
		t.Error("Default registry should have float type")
	}

	boolType := registry.GetBuiltinType(BoolType)
	if boolType == nil {
		t.Error("Default registry should have bool type")
	}

	stringType := registry.GetBuiltinType(StringType)
	if stringType == nil {
		t.Error("Default registry should have string type")
	}

	voidType := registry.GetBuiltinType(VoidType)
	if voidType == nil {
		t.Error("Default registry should have void type")
	}
}

// TestTypeRegistry_CustomTypes tests custom type registration
func TestTypeRegistry_CustomTypes(t *testing.T) {
	registry := NewTypeRegistry()
	intType := &BasicType{Kind: IntType}

	// Create a custom struct type
	pointType := &StructType{
		Name: "Point",
		Fields: map[string]Type{
			"x": intType,
			"y": intType,
		},
		Order: []string{"x", "y"},
	}

	// Register the custom type
	err := registry.RegisterType("Point", pointType)
	if err != nil {
		t.Errorf("Failed to register type: %v", err)
	}

	// Retrieve the type
	retrievedType, found := registry.GetType("Point")
	if !found {
		t.Error("Should be able to retrieve registered type")
	}

	if !retrievedType.Equals(pointType) {
		t.Error("Retrieved type should equal registered type")
	}

	// Test non-existent type
	_, found = registry.GetType("NonExistent")
	if found {
		t.Error("Should return false for non-existent type")
	}
}

// TestIsNumericType tests numeric type checking
func TestIsNumericType(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}
	boolType := &BasicType{Kind: BoolType}
	stringType := &BasicType{Kind: StringType}

	// Numeric types
	if !IsNumericType(intType) {
		t.Error("Int should be numeric type")
	}

	if !IsNumericType(floatType) {
		t.Error("Float should be numeric type")
	}

	// Non-numeric types
	if IsNumericType(boolType) {
		t.Error("Bool should not be numeric type")
	}

	if IsNumericType(stringType) {
		t.Error("String should not be numeric type")
	}
}

// TestIsComparableType tests comparable type checking
func TestIsComparableType(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}
	boolType := &BasicType{Kind: BoolType}
	stringType := &BasicType{Kind: StringType}
	voidType := &BasicType{Kind: VoidType}

	// Comparable types
	if !IsComparableType(intType) {
		t.Error("Int should be comparable")
	}

	if !IsComparableType(floatType) {
		t.Error("Float should be comparable")
	}

	if !IsComparableType(boolType) {
		t.Error("Bool should be comparable")
	}

	if !IsComparableType(stringType) {
		t.Error("String should be comparable")
	}

	// Non-comparable types
	if IsComparableType(voidType) {
		t.Error("Void should not be comparable")
	}
}

// TestCanApplyBinaryOperator tests binary operator compatibility
func TestCanApplyBinaryOperator(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}
	boolType := &BasicType{Kind: BoolType}
	stringType := &BasicType{Kind: StringType}

	// Arithmetic operators on numeric types
	if !CanApplyBinaryOperator(Add, intType, intType) {
		t.Error("Should be able to add int + int")
	}

	if !CanApplyBinaryOperator(Mul, floatType, floatType) {
		t.Error("Should be able to multiply float * float")
	}

	// Logical operators on bool types
	if !CanApplyBinaryOperator(And, boolType, boolType) {
		t.Error("Should be able to apply && to bool types")
	}

	// String comparison
	if !CanApplyBinaryOperator(Lt, stringType, stringType) {
		t.Error("Should be able to compare strings")
	}

	// Invalid combinations
	if CanApplyBinaryOperator(Add, intType, stringType) {
		t.Error("Should not be able to add int + string")
	}

	if CanApplyBinaryOperator(Mul, boolType, boolType) {
		t.Error("Should not be able to multiply bool * bool")
	}
}

// TestCanApplyUnaryOperator tests unary operator compatibility
func TestCanApplyUnaryOperator(t *testing.T) {
	intType := &BasicType{Kind: IntType}
	floatType := &BasicType{Kind: FloatType}
	boolType := &BasicType{Kind: BoolType}
	stringType := &BasicType{Kind: StringType}

	// Negation on numeric types
	if !CanApplyUnaryOperator(Neg, intType) {
		t.Error("Should be able to negate int")
	}

	if !CanApplyUnaryOperator(Neg, floatType) {
		t.Error("Should be able to negate float")
	}

	// Logical not on bool type
	if !CanApplyUnaryOperator(Not, boolType) {
		t.Error("Should be able to apply ! to bool")
	}

	// Invalid combinations
	if CanApplyUnaryOperator(Neg, boolType) {
		t.Error("Should not be able to negate bool")
	}

	if CanApplyUnaryOperator(Not, intType) {
		t.Error("Should not be able to apply ! to int")
	}

	if CanApplyUnaryOperator(Neg, stringType) {
		t.Error("Should not be able to negate string")
	}
}
