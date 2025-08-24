// Package infrastructure contains LLVM backend implementation
package infrastructure

import (
	"fmt"
	"io"
	"sync"

	"github.com/sokoide/llvm5/codegen"
	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// MockLLVMBackend provides a mock implementation for testing and development
type MockLLVMBackend struct {
	mutex        sync.RWMutex
	initialized  bool
	targetTriple string
	modules      map[string]*MockLLVMModule
	moduleCount  int
}

// NewMockLLVMBackend creates a new mock LLVM backend
func NewMockLLVMBackend() *MockLLVMBackend {
	return &MockLLVMBackend{
		modules: make(map[string]*MockLLVMModule),
	}
}

// Initialize initializes the LLVM backend
func (backend *MockLLVMBackend) Initialize(targetTriple string) error {
	backend.mutex.Lock()
	defer backend.mutex.Unlock()

	if backend.initialized {
		return fmt.Errorf("backend already initialized")
	}

	backend.targetTriple = targetTriple
	backend.initialized = true

	return nil
}

// CreateModule creates a new LLVM module
func (backend *MockLLVMBackend) CreateModule(name string) (interfaces.LLVMModule, error) {
	backend.mutex.Lock()
	defer backend.mutex.Unlock()

	if !backend.initialized {
		return nil, fmt.Errorf("backend not initialized")
	}

	module := &MockLLVMModule{
		name:      name,
		backend:   backend,
		functions: make(map[string]*MockLLVMFunction),
		globals:   make(map[string]*MockLLVMValue),
		structs:   make(map[string]*MockLLVMType),
		verified:  false,
	}

	backend.modules[name] = module
	backend.moduleCount++

	return module, nil
}

// Optimize optimizes the given module
func (backend *MockLLVMBackend) Optimize(module interfaces.LLVMModule, level int) error {
	backend.mutex.RLock()
	defer backend.mutex.RUnlock()

	// In a real implementation, this would run LLVM optimization passes
	mockModule, ok := module.(*MockLLVMModule)
	if !ok {
		return fmt.Errorf("invalid module type")
	}

	mockModule.optimizationLevel = level
	return nil
}

// EmitObject emits object code to the writer
func (backend *MockLLVMBackend) EmitObject(module interfaces.LLVMModule, output io.Writer) error {
	backend.mutex.RLock()
	defer backend.mutex.RUnlock()

	mockModule, ok := module.(*MockLLVMModule)
	if !ok {
		return fmt.Errorf("invalid module type")
	}

	// Generate mock object code
	_, err := fmt.Fprintf(output, "; Mock object code for module %s\n", mockModule.name)
	return err
}

// EmitAssembly emits assembly code to the writer
func (backend *MockLLVMBackend) EmitAssembly(module interfaces.LLVMModule, output io.Writer) error {
	backend.mutex.RLock()
	defer backend.mutex.RUnlock()

	mockModule, ok := module.(*MockLLVMModule)
	if !ok {
		return fmt.Errorf("invalid module type")
	}

	// Generate mock assembly code
	_, err := fmt.Fprintf(output, "; Mock assembly code for module %s\n", mockModule.name)
	return err
}

// Dispose disposes of backend resources
func (backend *MockLLVMBackend) Dispose() {
	backend.mutex.Lock()
	defer backend.mutex.Unlock()

	for _, module := range backend.modules {
		module.Dispose()
	}
	backend.modules = make(map[string]*MockLLVMModule)
	backend.initialized = false
}

// MockLLVMModule implements LLVMModule for testing
type MockLLVMModule struct {
	name              string
	backend           *MockLLVMBackend
	functions         map[string]*MockLLVMFunction
	globals           map[string]*MockLLVMValue
	structs           map[string]*MockLLVMType
	verified          bool
	optimizationLevel int
}

// CreateFunction creates a new function in the module
func (module *MockLLVMModule) CreateFunction(name string, funcType domain.Type) (interfaces.LLVMFunction, error) {
	if _, exists := module.functions[name]; exists {
		return nil, fmt.Errorf("function %s already exists", name)
	}

	ft, ok := funcType.(*domain.FunctionType)
	if !ok {
		return nil, fmt.Errorf("invalid function type")
	}

	function := &MockLLVMFunction{
		name:       name,
		funcType:   ft,
		module:     module,
		parameters: make([]*MockLLVMValue, len(ft.ParameterTypes)),
		blocks:     make(map[string]*MockLLVMBasicBlock),
	}

	// Create parameter values
	for i, paramType := range ft.ParameterTypes {
		function.parameters[i] = &MockLLVMValue{
			name:     fmt.Sprintf("param%d", i),
			typ:      &MockLLVMType{domainType: paramType},
			function: function,
		}
	}

	module.functions[name] = function
	return function, nil
}

// CreateGlobalVariable creates a global variable
func (module *MockLLVMModule) CreateGlobalVariable(name string, varType domain.Type) (interfaces.LLVMValue, error) {
	if _, exists := module.globals[name]; exists {
		return nil, fmt.Errorf("global variable %s already exists", name)
	}

	global := &MockLLVMValue{
		name: name,
		typ:  &MockLLVMType{domainType: varType},
	}

	module.globals[name] = global
	return global, nil
}

// CreateStruct creates a struct type
func (module *MockLLVMModule) CreateStruct(name string, structType *domain.StructType) (interfaces.LLVMType, error) {
	if _, exists := module.structs[name]; exists {
		return nil, fmt.Errorf("struct %s already exists", name)
	}

	llvmType := &MockLLVMType{
		name:       name,
		domainType: structType,
		isStruct:   true,
	}

	module.structs[name] = llvmType
	return llvmType, nil
}

// GetFunction gets a function by name
func (module *MockLLVMModule) GetFunction(name string) (interfaces.LLVMFunction, bool) {
	function, exists := module.functions[name]
	return function, exists
}

// Verify verifies the module
func (module *MockLLVMModule) Verify() error {
	// Mock verification - check for basic consistency
	for _, function := range module.functions {
		if len(function.blocks) == 0 {
			return fmt.Errorf("function %s has no basic blocks", function.name)
		}

		// Check that each block is properly terminated
		for blockName, block := range function.blocks {
			if !block.terminated {
				return fmt.Errorf("basic block %s in function %s is not terminated", blockName, function.name)
			}
		}
	}

	module.verified = true
	return nil
}

// Print prints the module IR
func (module *MockLLVMModule) Print(output io.Writer) {
	fmt.Fprintf(output, "; Module: %s\n", module.name)
	fmt.Fprintf(output, "target triple = \"%s\"\n\n", module.backend.targetTriple)

	// Print struct declarations
	for structName, structType := range module.structs {
		fmt.Fprintf(output, "%%struct.%s = type { ", structName)
		if st, ok := structType.domainType.(*domain.StructType); ok {
			for i, fieldName := range st.Order {
				if i > 0 {
					fmt.Fprintf(output, ", ")
				}
				fmt.Fprintf(output, "%s", st.Fields[fieldName].String())
			}
		}
		fmt.Fprintf(output, " }\n")
	}

	// Print global variables
	for globalName, global := range module.globals {
		if mockType, ok := global.typ.(*MockLLVMType); ok {
			fmt.Fprintf(output, "@%s = global %s\n", globalName, mockType.domainType.String())
		}
	}

	// Print functions
	for _, function := range module.functions {
		function.print(output)
	}
}

// Dispose disposes of the module
func (module *MockLLVMModule) Dispose() {
	for _, function := range module.functions {
		function.dispose()
	}
	module.functions = make(map[string]*MockLLVMFunction)
	module.globals = make(map[string]*MockLLVMValue)
	module.structs = make(map[string]*MockLLVMType)
}

// MockLLVMFunction implements LLVMFunction
type MockLLVMFunction struct {
	name       string
	funcType   *domain.FunctionType
	module     *MockLLVMModule
	parameters []*MockLLVMValue
	blocks     map[string]*MockLLVMBasicBlock
	blockCount int
}

// CreateBasicBlock creates a basic block in the function
func (function *MockLLVMFunction) CreateBasicBlock(name string) interfaces.LLVMBasicBlock {
	if name == "" {
		name = fmt.Sprintf("bb%d", function.blockCount)
	}

	block := &MockLLVMBasicBlock{
		name:       name,
		function:   function,
		terminated: false,
	}

	function.blocks[name] = block
	function.blockCount++

	return block
}

// GetParameter gets a parameter by index
func (function *MockLLVMFunction) GetParameter(index int) interfaces.LLVMValue {
	if index < 0 || index >= len(function.parameters) {
		return nil
	}
	return function.parameters[index]
}

// GetParameterCount gets the number of parameters
func (function *MockLLVMFunction) GetParameterCount() int {
	return len(function.parameters)
}

// SetName sets the function name
func (function *MockLLVMFunction) SetName(name string) {
	function.name = name
}

func (function *MockLLVMFunction) print(output io.Writer) {
	// Print function signature
	fmt.Fprintf(output, "define %s @%s(",
		function.funcType.ReturnType.String(),
		function.name)

	for i, paramType := range function.funcType.ParameterTypes {
		if i > 0 {
			fmt.Fprintf(output, ", ")
		}
		fmt.Fprintf(output, "%s %%param%d", paramType.String(), i)
	}
	fmt.Fprintf(output, ") {\n")

	// Print basic blocks
	for _, block := range function.blocks {
		fmt.Fprintf(output, "%s:\n", block.name)
		// In a real implementation, we'd print the instructions
		fmt.Fprintf(output, "  ; block content\n")
	}

	fmt.Fprintf(output, "}\n\n")
}

func (function *MockLLVMFunction) dispose() {
	for range function.blocks {
		// Clean up block resources
	}
	function.blocks = make(map[string]*MockLLVMBasicBlock)
}

// MockLLVMBasicBlock implements LLVMBasicBlock
type MockLLVMBasicBlock struct {
	name       string
	function   *MockLLVMFunction
	terminated bool
}

// GetName gets the block name
func (block *MockLLVMBasicBlock) GetName() string {
	return block.name
}

// IsTerminated checks if the block is terminated
func (block *MockLLVMBasicBlock) IsTerminated() bool {
	return block.terminated
}

// MockLLVMValue implements LLVMValue
type MockLLVMValue struct {
	name     string
	typ      interfaces.LLVMType
	function *MockLLVMFunction
}

// GetType gets the value type
func (value *MockLLVMValue) GetType() interfaces.LLVMType {
	return value.typ
}

// SetName sets the value name
func (value *MockLLVMValue) SetName(name string) {
	value.name = name
}

// GetName gets the value name
func (value *MockLLVMValue) GetName() string {
	return value.name
}

// MockLLVMType implements LLVMType
type MockLLVMType struct {
	name       string
	domainType domain.Type
	isStruct   bool
}

// IsInteger checks if the type is an integer
func (typ *MockLLVMType) IsInteger() bool {
	if basic, ok := typ.domainType.(*domain.BasicType); ok {
		return basic.Kind == domain.IntType
	}
	return false
}

// IsFloat checks if the type is a float
func (typ *MockLLVMType) IsFloat() bool {
	if basic, ok := typ.domainType.(*domain.BasicType); ok {
		return basic.Kind == domain.FloatType
	}
	return false
}

// IsPointer checks if the type is a pointer
func (typ *MockLLVMType) IsPointer() bool {
	// In this mock implementation, we'll consider string types as pointers
	if basic, ok := typ.domainType.(*domain.BasicType); ok {
		return basic.Kind == domain.StringType
	}
	return false
}

// IsStruct checks if the type is a struct
func (typ *MockLLVMType) IsStruct() bool {
	return typ.isStruct
}

// MockLLVMBuilder implements LLVMBuilder for instruction generation
type MockLLVMBuilder struct {
	currentBlock interfaces.LLVMBasicBlock
	instructions []MockInstruction
}

type MockInstruction struct {
	opcode   string
	operands []string
	result   string
}

// NewMockLLVMBuilder creates a new mock LLVM builder
func NewMockLLVMBuilder() *MockLLVMBuilder {
	return &MockLLVMBuilder{
		instructions: make([]MockInstruction, 0),
	}
}

// PositionAtEnd positions the builder at the end of a basic block
func (builder *MockLLVMBuilder) PositionAtEnd(block interfaces.LLVMBasicBlock) {
	builder.currentBlock = block
}

// CreateAlloca creates an alloca instruction
func (builder *MockLLVMBuilder) CreateAlloca(t interfaces.LLVMType, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode: "alloca",
		result: name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{
		name: name,
		typ:  t,
	}
}

// CreateStore creates a store instruction
func (builder *MockLLVMBuilder) CreateStore(value, ptr interfaces.LLVMValue) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "store",
		operands: []string{value.GetName(), ptr.GetName()},
	}
	builder.instructions = append(builder.instructions, instruction)

	return nil // Store returns void
}

// CreateLoad creates a load instruction
func (builder *MockLLVMBuilder) CreateLoad(ptr interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "load",
		operands: []string{ptr.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{
		name: name,
		typ:  ptr.GetType(),
	}
}

// Implement other builder methods with similar mock patterns...
func (builder *MockLLVMBuilder) CreateAdd(lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "add",
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{name: name, typ: lhs.GetType()}
}

func (builder *MockLLVMBuilder) CreateSub(lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "sub",
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{name: name, typ: lhs.GetType()}
}

func (builder *MockLLVMBuilder) CreateMul(lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "mul",
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{name: name, typ: lhs.GetType()}
}

func (builder *MockLLVMBuilder) CreateSDiv(lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "sdiv",
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{name: name, typ: lhs.GetType()}
}

func (builder *MockLLVMBuilder) CreateICmp(pred interfaces.IntPredicate, lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   fmt.Sprintf("icmp %d", int(pred)),
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	// Return boolean type
	return &MockLLVMValue{
		name: name,
		typ: &MockLLVMType{
			domainType: &domain.BasicType{Kind: domain.BoolType},
		},
	}
}

func (builder *MockLLVMBuilder) CreateFCmp(pred interfaces.FloatPredicate, lhs, rhs interfaces.LLVMValue, name string) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   fmt.Sprintf("fcmp %d", int(pred)),
		operands: []string{lhs.GetName(), rhs.GetName()},
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{
		name: name,
		typ: &MockLLVMType{
			domainType: &domain.BasicType{Kind: domain.BoolType},
		},
	}
}

func (builder *MockLLVMBuilder) CreateBr(dest interfaces.LLVMBasicBlock) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "br",
		operands: []string{dest.GetName()},
	}
	builder.instructions = append(builder.instructions, instruction)

	// Mark current block as terminated
	if mockBlock, ok := builder.currentBlock.(*MockLLVMBasicBlock); ok {
		mockBlock.terminated = true
	}

	return nil
}

func (builder *MockLLVMBuilder) CreateCondBr(cond interfaces.LLVMValue, then, else_ interfaces.LLVMBasicBlock) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "br",
		operands: []string{cond.GetName(), then.GetName(), else_.GetName()},
	}
	builder.instructions = append(builder.instructions, instruction)

	// Mark current block as terminated
	if mockBlock, ok := builder.currentBlock.(*MockLLVMBasicBlock); ok {
		mockBlock.terminated = true
	}

	return nil
}

func (builder *MockLLVMBuilder) CreateRet(value interfaces.LLVMValue) interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode:   "ret",
		operands: []string{value.GetName()},
	}
	builder.instructions = append(builder.instructions, instruction)

	// Mark current block as terminated
	if mockBlock, ok := builder.currentBlock.(*MockLLVMBasicBlock); ok {
		mockBlock.terminated = true
	}

	return nil
}

func (builder *MockLLVMBuilder) CreateRetVoid() interfaces.LLVMValue {
	instruction := MockInstruction{
		opcode: "ret void",
	}
	builder.instructions = append(builder.instructions, instruction)

	// Mark current block as terminated
	if mockBlock, ok := builder.currentBlock.(*MockLLVMBasicBlock); ok {
		mockBlock.terminated = true
	}

	return nil
}

func (builder *MockLLVMBuilder) CreateCall(fn interfaces.LLVMFunction, args []interfaces.LLVMValue, name string) interfaces.LLVMValue {
	operands := make([]string, len(args)+1)
	operands[0] = "@" + fn.(*MockLLVMFunction).name
	for i, arg := range args {
		operands[i+1] = arg.GetName()
	}

	instruction := MockInstruction{
		opcode:   "call",
		operands: operands,
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	// Return type based on function return type
	return &MockLLVMValue{
		name: name,
		typ: &MockLLVMType{
			domainType: fn.(*MockLLVMFunction).funcType.ReturnType,
		},
	}
}

func (builder *MockLLVMBuilder) CreateGEP(ptr interfaces.LLVMValue, indices []interfaces.LLVMValue, name string) interfaces.LLVMValue {
	operands := make([]string, len(indices)+1)
	operands[0] = ptr.GetName()
	for i, index := range indices {
		operands[i+1] = index.GetName()
	}

	instruction := MockInstruction{
		opcode:   "getelementptr",
		operands: operands,
		result:   name,
	}
	builder.instructions = append(builder.instructions, instruction)

	return &MockLLVMValue{
		name: name,
		typ:  ptr.GetType(), // Simplified - would be element type
	}
}

func (builder *MockLLVMBuilder) Dispose() {
	builder.instructions = builder.instructions[:0]
	builder.currentBlock = nil
}

// RealLLVMIRGenerator implements interfaces.CodeGenerator using the existing codegen.Generator
type RealLLVMIRGenerator struct {
	generator     *codegen.Generator
	output        io.Writer
	options       interfaces.CodeGenOptions
	errorReporter domain.ErrorReporter
}

// NewRealLLVMIRGenerator creates a new real LLVM IR generator
func NewRealLLVMIRGenerator() *RealLLVMIRGenerator {
	return &RealLLVMIRGenerator{
		generator: codegen.NewGenerator(),
	}
}

// Generate generates LLVM IR for the given AST using the real code generator
func (cg *RealLLVMIRGenerator) Generate(ast *domain.Program) error {
	if cg.output == nil {
		return fmt.Errorf("output not set")
	}

	// Generate LLVM IR using the existing codegen.Generator
	llvmIR, err := cg.generator.Generate(ast)
	if err != nil {
		return fmt.Errorf("failed to generate LLVM IR: %v", err)
	}

	// Write the generated LLVM IR to the output
	_, err = cg.output.Write([]byte(llvmIR))
	if err != nil {
		return fmt.Errorf("failed to write LLVM IR: %v", err)
	}

	return nil
}

// SetOutput sets the output destination
func (cg *RealLLVMIRGenerator) SetOutput(output io.Writer) {
	cg.output = output
}

// SetOptions sets code generation options
func (cg *RealLLVMIRGenerator) SetOptions(options interfaces.CodeGenOptions) {
	cg.options = options

	// Set the target triple in the generator if supported
	// The codegen.Generator currently uses a fixed target triple
	// but we could extend it to be configurable in the future
}

// SetErrorReporter sets the error reporter
func (cg *RealLLVMIRGenerator) SetErrorReporter(reporter domain.ErrorReporter) {
	cg.errorReporter = reporter
	cg.generator.SetErrorReporter(reporter)
}
