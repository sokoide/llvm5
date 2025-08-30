// Package domain contains the type system definitions
package domain

import (
	"fmt"
	"strings"
)

// Type represents a type in the StaticLang type system
type Type interface {
	String() string
	Equals(other Type) bool
	IsAssignableFrom(other Type) bool
	GetSize() int // Size in bytes
}

// BasicType represents primitive types
type BasicType struct {
	Kind BasicTypeKind
}

type BasicTypeKind int

const (
	IntType BasicTypeKind = iota
	FloatType
	BoolType
	StringType
	VoidType
)

func (bt *BasicType) String() string {
	switch bt.Kind {
	case IntType:
		return "int"
	case FloatType:
		return "float"
	case BoolType:
		return "bool"
	case StringType:
		return "string"
	case VoidType:
		return "void"
	default:
		return "unknown"
	}
}

func (bt *BasicType) Equals(other Type) bool {
	if otherBasic, ok := other.(*BasicType); ok {
		return bt.Kind == otherBasic.Kind
	}
	return false
}

func (bt *BasicType) IsAssignableFrom(other Type) bool {
	return bt.Equals(other)
}

func (bt *BasicType) GetSize() int {
	switch bt.Kind {
	case IntType:
		return 8 // 64-bit integers
	case FloatType:
		return 8 // 64-bit floats
	case BoolType:
		return 1
	case StringType:
		return 8 // pointer to string data
	case VoidType:
		return 0
	default:
		return 0
	}
}

// ArrayType represents array types
type ArrayType struct {
	ElementType Type
	Size        int // -1 for dynamic arrays
}

func (at *ArrayType) String() string {
	if at.Size == -1 {
		return fmt.Sprintf("[]%s", at.ElementType.String())
	}
	return fmt.Sprintf("[%d]%s", at.Size, at.ElementType.String())
}

func (at *ArrayType) Equals(other Type) bool {
	if otherArray, ok := other.(*ArrayType); ok {
		return at.Size == otherArray.Size && at.ElementType.Equals(otherArray.ElementType)
	}
	return false
}

func (at *ArrayType) IsAssignableFrom(other Type) bool {
	if otherArray, ok := other.(*ArrayType); ok {
		// Dynamic arrays can accept static arrays of the same element type
		if at.Size == -1 && otherArray.Size >= 0 {
			return at.ElementType.Equals(otherArray.ElementType)
		}
		return at.Equals(other)
	}
	return false
}

func (at *ArrayType) GetSize() int {
	if at.Size == -1 {
		return 8 // pointer to array data
	}
	return at.Size * at.ElementType.GetSize()
}

// StructType represents struct types
type StructType struct {
	Name   string
	Fields map[string]Type
	Order  []string // Preserve field order
}

func (st *StructType) String() string {
	if st.Name != "" {
		return st.Name
	}

	// If no name, show field details
	if len(st.Fields) == 0 {
		return "struct{}"
	}

	// Use Order slice if populated, otherwise use field map keys (order not guaranteed)
	var fieldNames []string
	if len(st.Order) > 0 {
		fieldNames = st.Order
	} else {
		for fieldName := range st.Fields {
			fieldNames = append(fieldNames, fieldName)
		}
	}

	fields := make([]string, len(fieldNames))
	for i, fieldName := range fieldNames {
		fieldType := st.Fields[fieldName]
		fields[i] = fieldName + " " + fieldType.String()
	}

	return "struct{" + strings.Join(fields, ", ") + "}"
}

func (st *StructType) Equals(other Type) bool {
	if otherStruct, ok := other.(*StructType); ok {
		return st.Name == otherStruct.Name
	}
	return false
}

func (st *StructType) IsAssignableFrom(other Type) bool {
	return st.Equals(other)
}

func (st *StructType) GetSize() int {
	size := 0
	for _, fieldName := range st.Order {
		fieldType := st.Fields[fieldName]
		size += fieldType.GetSize()
	}
	return size
}

func (st *StructType) GetField(name string) (Type, bool) {
	fieldType, exists := st.Fields[name]
	return fieldType, exists
}

// FunctionType represents function types
type FunctionType struct {
	ParameterTypes []Type
	ReturnType     Type
}

func (ft *FunctionType) String() string {
	params := make([]string, len(ft.ParameterTypes))
	for i, param := range ft.ParameterTypes {
		params[i] = param.String()
	}
	return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), ft.ReturnType.String())
}

func (ft *FunctionType) Equals(other Type) bool {
	if otherFunc, ok := other.(*FunctionType); ok {
		if len(ft.ParameterTypes) != len(otherFunc.ParameterTypes) {
			return false
		}
		for i, param := range ft.ParameterTypes {
			if !param.Equals(otherFunc.ParameterTypes[i]) {
				return false
			}
		}
		return ft.ReturnType.Equals(otherFunc.ReturnType)
	}
	return false
}

func (ft *FunctionType) IsAssignableFrom(other Type) bool {
	return ft.Equals(other)
}

func (ft *FunctionType) GetSize() int {
	return 8 // function pointer
}

// TypeError represents an error type for type checking failures
type TypeError struct {
	Message string
}

func (et *TypeError) String() string {
	return fmt.Sprintf("<error: %s>", et.Message)
}

func (et *TypeError) Equals(other Type) bool {
	_, ok := other.(*TypeError)
	return ok
}

func (et *TypeError) IsAssignableFrom(other Type) bool {
	return false
}

func (et *TypeError) GetSize() int {
	return 0
}

// TypeRegistry manages type definitions and provides type operations
type TypeRegistry interface {
	RegisterType(name string, t Type) error
	GetType(name string) (Type, bool)
	CreateStructType(name string, fields []StructField) (*StructType, error)
	GetBuiltinType(kind BasicTypeKind) Type
}

// DefaultTypeRegistry provides the default implementation of TypeRegistry
type DefaultTypeRegistry struct {
	types    map[string]Type
	builtins map[BasicTypeKind]Type
}

func NewDefaultTypeRegistry() *DefaultTypeRegistry {
	reg := &DefaultTypeRegistry{
		types:    make(map[string]Type),
		builtins: make(map[BasicTypeKind]Type),
	}

	// Register builtin types
	reg.builtins[IntType] = &BasicType{Kind: IntType}
	reg.builtins[FloatType] = &BasicType{Kind: FloatType}
	reg.builtins[BoolType] = &BasicType{Kind: BoolType}
	reg.builtins[StringType] = &BasicType{Kind: StringType}
	reg.builtins[VoidType] = &BasicType{Kind: VoidType}

	// Also register them by name
	reg.types["int"] = reg.builtins[IntType]
	reg.types["float"] = reg.builtins[FloatType]
	reg.types["bool"] = reg.builtins[BoolType]
	reg.types["string"] = reg.builtins[StringType]
	reg.types["void"] = reg.builtins[VoidType]

	return reg
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() TypeRegistry {
	return NewDefaultTypeRegistry()
}

func (reg *DefaultTypeRegistry) RegisterType(name string, t Type) error {
	if _, exists := reg.types[name]; exists {
		return fmt.Errorf("type '%s' already registered", name)
	}
	reg.types[name] = t
	return nil
}

func (reg *DefaultTypeRegistry) GetType(name string) (Type, bool) {
	t, exists := reg.types[name]
	return t, exists
}

func (reg *DefaultTypeRegistry) CreateStructType(name string, fields []StructField) (*StructType, error) {
	if _, exists := reg.types[name]; exists {
		return nil, fmt.Errorf("type '%s' already exists", name)
	}

	structType := &StructType{
		Name:   name,
		Fields: make(map[string]Type),
		Order:  make([]string, 0, len(fields)),
	}

	for _, field := range fields {
		if _, exists := structType.Fields[field.Name]; exists {
			return nil, fmt.Errorf("duplicate field '%s' in struct '%s'", field.Name, name)
		}
		structType.Fields[field.Name] = field.Type
		structType.Order = append(structType.Order, field.Name)
	}

	reg.types[name] = structType
	return structType, nil
}

func (reg *DefaultTypeRegistry) GetBuiltinType(kind BasicTypeKind) Type {
	return reg.builtins[kind]
}

// Type checking utilities
func IsNumericType(t Type) bool {
	if basic, ok := t.(*BasicType); ok {
		return basic.Kind == IntType || basic.Kind == FloatType
	}
	return false
}

func IsComparableType(t Type) bool {
	if basic, ok := t.(*BasicType); ok {
		return basic.Kind == IntType || basic.Kind == FloatType || basic.Kind == BoolType || basic.Kind == StringType
	}
	return false
}

func CanApplyBinaryOperator(op BinaryOperator, left, right Type) bool {
	switch op {
	case Add, Sub, Mul, Div, Mod:
		return IsNumericType(left) && left.Equals(right)
	case Eq, Ne:
		return IsComparableType(left) && left.Equals(right)
	case Lt, Le, Gt, Ge:
		return (IsNumericType(left) || left.String() == "string") && left.Equals(right)
	case And, Or:
		return left.String() == "bool" && right.String() == "bool"
	default:
		return false
	}
}

// Helper functions for creating basic types
func NewIntType() Type {
	return &BasicType{Kind: IntType}
}

func NewFloatType() Type {
	return &BasicType{Kind: FloatType}
}

func NewBoolType() Type {
	return &BasicType{Kind: BoolType}
}

func NewStringType() Type {
	return &BasicType{Kind: StringType}
}

func NewVoidType() Type {
	return &BasicType{Kind: VoidType}
}

func CanApplyUnaryOperator(op UnaryOperator, operand Type) bool {
	switch op {
	case Neg:
		return IsNumericType(operand)
	case Not:
		return operand.String() == "bool"
	default:
		return false
	}
}
