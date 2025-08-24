package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/staticlang/internal/application"
	"github.com/sokoide/llvm5/staticlang/internal/domain"
)

func TestCompileHelloWorldExample(t *testing.T) {
	// Create a temporary source file
	sourceCode := `
int main() {
    string message = "Hello, StaticLang!";
    print(message);
    return 0;
}
`

	tempFile := createTempFile(t, "hello.sl", sourceCode)
	defer os.Remove(tempFile)

	// Create compiler configuration
	config := application.CompilerConfig{
		UseMockComponents: true, // Use mocks for testing
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "test_output.ll",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	// Create compiler factory and pipeline
	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	// Open input file
	input, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	// Create output file
	outputFile := createTempFile(t, "output.ll", "")
	defer os.Remove(outputFile)

	output, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	// Compile
	err = pipeline.Compile(tempFile, input, output)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Read and verify output
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)

	// Verify output contains expected LLVM IR
	if !strings.Contains(outputStr, "define i32 @main()") {
		t.Error("Output should contain main function definition")
	}

	if !strings.Contains(outputStr, "@printf") {
		t.Error("Output should contain printf declaration or call")
	}
}

func TestCompileFibonacciExample(t *testing.T) {
	sourceCode := `
function fibonacci(int n) -> int {
    if (n <= 1) {
        return n;
    } else {
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}

int main() {
    int num = 10;
    int result = fibonacci(num);
    print(result);
    return 0;
}
`

	tempFile := createTempFile(t, "fibonacci.sl", sourceCode)
	defer os.Remove(tempFile)

	// Create compiler configuration
	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 1,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "fibonacci_output.ll",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	input, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	outputFile := createTempFile(t, "fibonacci_output.ll", "")
	defer os.Remove(outputFile)

	output, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	err = pipeline.Compile(tempFile, input, output)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Read and verify output
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)

	// Verify function definitions
	if !strings.Contains(outputStr, "define i32 @fibonacci(i32") {
		t.Error("Output should contain fibonacci function definition")
	}

	if !strings.Contains(outputStr, "define i32 @main()") {
		t.Error("Output should contain main function definition")
	}

	// Verify recursive calls
	if !strings.Contains(outputStr, "call i32 @fibonacci") {
		t.Error("Output should contain recursive fibonacci calls")
	}
}

func TestCompileTypesExample(t *testing.T) {
	sourceCode := `
int globalCounter = 0;

function testTypes() -> int {
    int x = 42;
    double pi = 3.14159;
    string name = "StaticLang";
    
    print(x);
    print(pi);
    print(name);
    
    int sum = x + 10;
    double product = pi * 2.0;
    
    print(sum);
    print(product);
    
    if (x > 40) {
        print("x is greater than 40");
    }
    
    return sum;
}

int main() {
    int result = testTypes();
    globalCounter = result;
    print(globalCounter);
    return 0;
}
`

	tempFile := createTempFile(t, "types.sl", sourceCode)
	defer os.Remove(tempFile)

	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         true,
			TargetTriple:      "",
			OutputPath:        "types_output.ll",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	input, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	outputFile := createTempFile(t, "types_output.ll", "")
	defer os.Remove(outputFile)

	output, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	err = pipeline.Compile(tempFile, input, output)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)

	// Verify different type operations
	if !strings.Contains(outputStr, "i32") {
		t.Error("Output should contain int type operations")
	}

	if !strings.Contains(outputStr, "double") {
		t.Error("Output should contain double type operations")
	}

	if !strings.Contains(outputStr, "i8*") {
		t.Error("Output should contain string type operations")
	}

	// Verify global variable
	if !strings.Contains(outputStr, "@globalCounter") {
		t.Error("Output should contain global variable definition")
	}
}

func TestCompileErrorHandling(t *testing.T) {
	// Test compilation with syntax error
	sourceCode := `
int main() {
    int x = 42
    // Missing semicolon should cause syntax error
    print(x);
    return 0;
}
`

	tempFile := createTempFile(t, "error.sl", sourceCode)
	defer os.Remove(tempFile)

	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "error_output.ll",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	input, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	outputFile := createTempFile(t, "error_output.ll", "")
	defer os.Remove(outputFile)

	output, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	// This should fail due to syntax error
	err = pipeline.Compile(tempFile, input, output)
	if err == nil {
		t.Error("Compilation should have failed due to syntax error")
	}
}

func TestCompileTypeErrorHandling(t *testing.T) {
	// Test type mismatch error
	sourceCode := `
int main() {
    int x = 42;
    string y = x; // Type mismatch error
    print(y);
    return 0;
}
`

	tempFile := createTempFile(t, "type_error.sl", sourceCode)
	defer os.Remove(tempFile)

	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        "type_error_output.ll",
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	pipeline := factory.CreateCompilerPipeline()

	input, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	outputFile := createTempFile(t, "type_error_output.ll", "")
	defer os.Remove(outputFile)

	output, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	// This should fail due to type error
	err = pipeline.Compile(tempFile, input, output)
	if err == nil {
		t.Error("Compilation should have failed due to type mismatch")
	}
}

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, name, content string) string {
	tempFile, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		_, err = tempFile.WriteString(content)
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	tempFile.Close()
	return tempFile.Name()
}
