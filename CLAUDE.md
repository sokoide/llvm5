# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

StaticLang is a production-ready compiler for a statically-typed programming language built with Go and LLVM. The project follows Clean Architecture principles with clear separation of concerns across four layers: Application, Interface, Domain, and Infrastructure.

**Key Status**: The compiler now features real LLVM IR code generation (not mocks) and produces valid LLVM Intermediate Representation with proper syntax, functions, and control flow.

## Essential Commands

### Building and Testing
- `make build` - Build the compiler for current platform
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make all` - Complete pipeline (fmt + vet + test + build)

### Code Quality
- `make fmt` - Format code using gofmt
- `make vet` - Run go vet static analysis  
- `make lint` - Run golangci-lint (if installed)

### Development
- `make deps` - Install/update dependencies
- `make dev-setup` - Install development tools (goyacc, golangci-lint)
- `make generate-parser` - Generate parser from grammar using goyacc
- `make run-example` - Build and run example compilation

### Running the Compiler
- `./build/staticlang -i hello.sl -o hello.ll -v` - Compile with verbose output (produces real LLVM IR)
- `./build/staticlang -i hello.sl -o hello.ll -mock -v` - Use mock backend for development

## Architecture Overview

### Clean Architecture Layers
```
Application Layer  (CLI, Pipeline, Factory)
      ↓
Interface Layer    (Component Contracts)  
      ↓
Domain Layer       (AST, Types, Core Logic)
      ↓
Infrastructure     (LLVM, Symbol Tables, I/O)
```

### Key Directories
- `cmd/staticlang/` - CLI application entry point
- `internal/application/` - Application layer (pipeline, factory)
- `internal/interfaces/` - Interface contracts  
- `internal/domain/` - Core domain logic (AST, types)
- `internal/infrastructure/` - External concerns (LLVM, I/O)
- `grammar/` - Parser grammar and generated code
- `lexer/` - Lexical analysis
- `semantic/` - Semantic analysis
- `codegen/` - Code generation
- `examples/` - Example programs
- `tests/` - Test files

### Compilation Pipeline
```
Input Source → Lexer → Parser → Semantic Analyzer → Code Generator → LLVM IR
              ↓         ↓           ↓                    ↓
           Tokens    AST       Typed AST         LLVM Module
```

## Important Architectural Notes

### Interface Design
- All major components use interfaces for dependency inversion
- Mock implementations available for testing (use `-mock` flag)
- Real LLVM backend implemented and working (default mode)

### AST and Type System
- Visitor pattern used for AST traversal
- Strong static typing with comprehensive type checking
- Symbol table with hierarchical scope management
- All AST nodes implement `Node` interface with `GetLocation()` and `Accept(visitor)`

### Error Handling
- Structured errors with source location tracking
- Multiple error reporter implementations (Console, Sorted, Tracking)
- Error recovery in parser when possible

### Memory Management
- Memory pools for efficient allocation
- String deduplication with reference counting
- Automatic cleanup after compilation phases

## Known Issues and Constraints

### Current Architectural Issues
- Some interface inconsistencies between layers (see architectural_issues memory)
- Parser integration requires careful handling of generated code
- Field naming inconsistencies in AST nodes (some use `Type_` to avoid Go keywords)

### Development Constraints
- Go 1.21+ required
- LLVM 15+ optional (mock backend available)
- Use `make generate-parser` when modifying grammar files
- Always run `make fmt vet` before committing

## Language Features Supported

- **Basic Types**: `int`, `float`, `bool`, `string`, `void`
- **Functions**: Parameters, return values, local variables
- **Structs**: User-defined composite types with member access
- **Arrays**: Static and dynamic arrays with indexing
- **Control Flow**: `if/else`, `while`, `for` loops
- **Expressions**: Arithmetic, logical, comparison operations

## Testing Strategy

- Unit tests with mock implementations for isolation
- Integration tests for end-to-end compilation
- Real LLVM IR output verification
- Memory usage and performance benchmarking
- Use `make test-coverage` to ensure comprehensive test coverage

## Building and Deployment

### Multi-Platform Build
- `make build-all` - Build for Linux, macOS, Windows
- Cross-compilation supported via GOOS/GOARCH
- Docker support available (`make docker-build`, `make docker-run`)

### Development Tools
The project uses standard Go tooling plus:
- `goyacc` for parser generation from grammar
- `golangci-lint` for comprehensive linting  
- Native Go testing framework (no external test dependencies)