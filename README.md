# StaticLang Compiler

A compiler project for the StaticLang programming language built with Go, following clean architecture principles with comprehensive error handling. **Currently in development with mock components - not yet production-ready.**

## ⚠️ Development Status

**Important**: This project is currently in active development. The compiler uses mock implementations by default and does not yet perform real LLVM code generation. All compilation currently outputs mock results for development and testing purposes.

## Features

- **Clean Architecture**: Layered design with clear interfaces and dependency inversion
- **Comprehensive Type System**: Strong static typing with type inference
- **Advanced Error Reporting**: Detailed error messages with source context
- **Memory Management**: Efficient memory pools and tracking
- **Mock LLVM Backend**: Complete mock implementation for development and testing
- **Extensible Design**: Plugin-ready architecture for future enhancements

## 🚧 Current Limitations

- **Mock Components**: All core components (lexer, parser, semantic analyzer, code generator) use mock implementations by default
- **No Real LLVM**: Code generation produces mock output, not actual LLVM IR
- **Development Only**: Not suitable for production use
- **Limited Language Support**: Basic language features implemented, advanced features pending

## Quick Start

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)
- LLVM 15+ (optional, for future real LLVM integration - currently uses mocks)

### Installation

```bash
# Clone the repository
git clone https://github.com/sokoide/llvm5.git
cd staticlang

# Install dependencies
make deps

# Build the compiler
make build

# Install to system (optional)
make install
```

### Basic Usage

```bash
# Compile a single file (uses mock components by default)
./build/staticlang -i hello.sl -o hello.ll -v

# Compile with explicit mock flag (recommended for development)
./build/staticlang -i hello.sl -o hello.ll -mock -v

# Compile multiple files (all examples use mock components)
./build/staticlang -i "main.sl,lib.sl" -o program.ll -mock -O 2

# Enable debug info and verbose output
./build/staticlang -i main.sl -o main.ll -mock -g -v
```

## Architecture Overview

The StaticLang compiler follows a layered architecture pattern:

```
Application Layer  (CLI, Pipeline, Factory)
      ↓
Interface Layer    (Component Contracts)
      ↓
Domain Layer       (AST, Types, Core Logic)
      ↓
Infrastructure     (LLVM, Symbol Tables, I/O)
```

### Key Components

- **Lexer**: Tokenizes source code with position tracking (real implementation)
- **Parser**: Builds AST using recursive descent parsing (real implementation)
- **Semantic Analyzer**: Type checking and symbol resolution (real implementation)
- **Code Generator**: LLVM IR generation with optimization (currently mock)
- **Error Reporter**: Advanced error reporting with source context (real implementation)

## Language Features

StaticLang supports:

- **Basic Types**: `int`, `float`, `bool`, `string`
- **Functions**: First-class functions with parameters and return values
- **Structs**: User-defined composite types
- **Arrays**: Static and dynamic arrays
- **Control Flow**: `if/else`, `while`, `for` loops
- **Expressions**: Arithmetic, logical, and comparison operations

### Example Program

```staticlang
struct Point {
    x: float
    y: float
}

func distance(p1: Point, p2: Point) -> float {
    dx := p1.x - p2.x
    dy := p1.y - p2.y
    return sqrt(dx*dx + dy*dy)
}

func main() -> int {
    origin := Point{x: 0.0, y: 0.0}
    point := Point{x: 3.0, y: 4.0}

    dist := distance(origin, point)
    print("Distance: ", dist)

    return 0
}
```

## Development

### Project Structure

```
staticlang/
├── cmd/staticlang/              # CLI application
├── internal/
│   ├── application/             # Application services
│   ├── domain/                  # Core domain logic
│   ├── interfaces/              # Interface definitions
│   └── infrastructure/          # External concerns
├── examples/                    # Example programs
├── tests/                       # Test files
└── docs/                        # Documentation
```

### Building from Source

```bash
# Development setup
make dev-setup

# Format and lint code
make fmt vet lint

# Run tests with coverage
make test-coverage

# Build for all platforms
make build-all

# Run benchmarks
make bench
```

### Testing

The project includes comprehensive testing:

```bash
# Unit tests
make test

# Integration tests
go test -tags=integration ./...

# Benchmark tests
make bench

# Test with mock components
./build/staticlang -i examples/hello.sl -mock -v
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the architecture patterns
4. Add tests for new functionality
5. Run the full test suite (`make all`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Architecture Details

### Clean Architecture Principles

The compiler follows clean architecture with:

- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Interface Segregation**: Focused, cohesive interfaces
- **Single Responsibility**: Each component has one reason to change
- **Open/Closed**: Open for extension, closed for modification

### Error Handling Strategy

- **Structured Errors**: Typed errors with source location tracking
- **Error Recovery**: Parser continues after syntax errors when possible
- **Helpful Messages**: Context and suggestions for common errors
- **Multiple Formats**: Console output with syntax highlighting

### Memory Management

- **Memory Pools**: Type-specific allocation pools for efficiency
- **Reference Counting**: String deduplication with reference counting
- **Automatic Cleanup**: Memory freed after compilation phases
- **Statistics**: Detailed memory usage tracking and reporting

### LLVM Integration

- **Current Status**: Mock implementation only - no real LLVM integration yet
- **Architecture**: LLVM functionality abstracted through interfaces (ready for real implementation)
- **Mock Backend**: Complete mock implementation for development and testing
- **Future Plans**: Real LLVM integration planned for future versions
- **Optimization**: Configurable optimization levels (0-3) - currently simulated

## Performance

### Current Status

**Performance metrics not available**: The compiler currently uses mock implementations and does not perform real compilation. Performance benchmarking will be available once real LLVM integration is implemented.

### Planned Optimizations

The architecture includes several optimization strategies planned for future implementation:

- **Memory Pooling**: Reduces allocation overhead (partially implemented)
- **String Interning**: Deduplicates string literals (planned)
- **AST Caching**: Reuses parsed AST nodes when possible (planned)
- **Parallel Processing**: Multi-threaded compilation phases (planned)

## Extending the Compiler

### Adding New Language Features

1. **Lexer**: Add token types in `interfaces/compiler.go`
2. **Parser**: Extend grammar and AST nodes in `domain/ast.go`
3. **Type System**: Add types in `domain/type_system.go`
4. **Semantic Analysis**: Implement type checking rules
5. **Code Generation**: Add visitor methods for new AST nodes

### Custom Error Reporting

```go
// Custom error reporter implementation
type MyErrorReporter struct {
    // Custom fields
}

func (er *MyErrorReporter) ReportError(err domain.CompilerError) {
    // Custom error handling logic
}

// Use in factory
config := CompilerConfig{
    ErrorReporterType: CustomErrorReporter,
}
```

### Plugin Architecture

The interface-based design supports plugins:

```go
// Custom code generator plugin
type MyCodeGenerator struct {
    // Plugin implementation
}

func (cg *MyCodeGenerator) Generate(ast *domain.Program) error {
    // Custom code generation logic
    return nil
}
```

## Docker Support

```bash
# Build Docker image
make docker-build

# Run in container
make docker-run

# Development with Docker
docker run --rm -v $(pwd):/workspace staticlang:latest -i hello.sl -mock
```

## Troubleshooting

### Common Issues

**Q: "lexer not set" error**
A: Ensure all pipeline components are configured through the factory.

**Q: LLVM linking errors**
A: Use `-mock` flag for development without LLVM dependencies.

**Q: Memory usage too high**
A: Try the `CompactMemoryManager` for smaller memory footprint.

### Debug Mode

```bash
# Build debug version
make debug

# Run with verbose logging
./build/staticlang-debug -i main.sl -v

# Enable all debug output
STATICLANG_DEBUG=1 ./build/staticlang -i main.sl
```

## Roadmap

### Current Version (0.1.0)
- ✅ Clean Architecture implementation
- ✅ Basic compiler pipeline with mock components
- ✅ Core language features (types, functions, control flow)
- ✅ Comprehensive error reporting
- ❌ Real LLVM integration (planned)
- ❌ Production-ready compilation (planned)

### Version 0.2.0 - LLVM Integration
- [ ] Complete LLVM integration (replacing mock)
- [ ] Real code generation and optimization
- [ ] Performance benchmarking
- [ ] Production-ready compilation pipeline

### Version 0.3.0 - Language Features
- [ ] Package system
- [ ] Standard library
- [ ] Advanced language features
- [ ] Incremental compilation

### Version 1.0.0 - Production Ready
- [ ] Full production stability
- [ ] Language server protocol
- [ ] IDE integration
- [ ] Comprehensive documentation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- LLVM Project for the code generation backend
- Go team for the excellent tooling and runtime
- Clean Architecture community for design principles

## Contact

- **Repository**: https://github.com/sokoide/llvm5
- **Issues**: https://github.com/sokoide/llvm5/issues
- **Discussions**: https://github.com/sokoide/llvm5/discussions

---

*Built with ❤️ using Go and Clean Architecture principles.*