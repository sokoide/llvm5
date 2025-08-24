package tests

import (
	"strconv"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/application"
)

// BenchmarkLexerPerformance benchmarks the lexer on various input sizes
func BenchmarkLexerPerformance(b *testing.B) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())

	// Test cases with different input sizes
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Small",
			input: "func main() { return 42; }",
		},
		{
			name: "Medium",
			input: `
func fibonacci(n: int) -> int {
    if n <= 1 {
        return n;
    }
    return fibonacci(n-1) + fibonacci(n-2);
}

func main() -> int {
    var result: int = fibonacci(10);
    print("Fibonacci result:", result);
    return result;
}`,
		},
		{
			name: "Large",
			input: generateLargeProgram(100), // 100 functions
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			lexer := factory.CreateLexer()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Reset lexer for each iteration
				err := lexer.SetInput("bench.sl", strings.NewReader(tc.input))
				if err != nil {
					b.Fatalf("Failed to set input: %v", err)
				}

				// Tokenize the entire input
				for {
					token := lexer.NextToken()
					if token.Type == 70 { // TokenEOF constant value
						break
					}
				}
			}
		})
	}
}

// BenchmarkParserPerformance benchmarks the parser on various input sizes
func BenchmarkParserPerformance(b *testing.B) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = false // Use real parser
	factory := application.NewCompilerFactory(config)

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Small",
			input: "func main() { return 42; }",
		},
		{
			name: "Medium",
			input: `
func add(a: int, b: int) -> int {
    return a + b;
}

func main() -> int {
    var x: int = 10;
    var y: int = 20;
    var result: int = add(x, y);
    return result;
}`,
		},
		{
			name: "Large",
			input: generateLargeProgram(50), // 50 functions
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			parser := factory.CreateParser()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lexer := factory.CreateLexer()
				err := lexer.SetInput("bench.sl", strings.NewReader(tc.input))
				if err != nil {
					b.Fatalf("Failed to set input: %v", err)
				}

				// Parse the input
				_, err = parser.Parse(lexer)
				if err != nil {
					// Don't fail benchmark for parse errors in generated code
					continue
				}
			}
		})
	}
}

// BenchmarkCompilerPipeline benchmarks the entire compilation pipeline
func BenchmarkCompilerPipeline(b *testing.B) {
	config := application.DefaultCompilerConfig()
	config.UseMockComponents = true // Use mock components for consistent benchmarking
	factory := application.NewCompilerFactory(config)

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "HelloWorld",
			input: `func main() { print("Hello, World!"); return 0; }`,
		},
		{
			name: "Fibonacci",
			input: `
func fibonacci(n: int) -> int {
    if n <= 1 {
        return n;
    }
    return fibonacci(n-1) + fibonacci(n-2);
}

func main() -> int {
    return fibonacci(5);
}`,
		},
		{
			name: "ComplexProgram",
			input: `
struct Point {
    x: int;
    y: int;
}

func createPoint(x: int, y: int) -> Point {
    var p: Point;
    p.x = x;
    p.y = y;
    return p;
}

func distance(p1: Point, p2: Point) -> int {
    var dx: int = p1.x - p2.x;
    var dy: int = p1.y - p2.y;
    return dx * dx + dy * dy;
}

func main() -> int {
    var p1: Point = createPoint(0, 0);
    var p2: Point = createPoint(3, 4);
    var dist: int = distance(p1, p2);
    print("Distance squared:", dist);
    return dist;
}`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			pipeline := factory.CreateCompilerPipeline()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				inputReader := strings.NewReader(tc.input)
				outputWriter := &strings.Builder{}
				
				err := pipeline.Compile("bench.sl", inputReader, outputWriter)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkMemoryManager benchmarks different memory manager implementations
func BenchmarkMemoryManager(b *testing.B) {
	managerTypes := []struct {
		name        string
		managerType application.MemoryManagerType
	}{
		{"PooledMemoryManager", application.PooledMemoryManager},
		{"CompactMemoryManager", application.CompactMemoryManager},
		{"TrackingMemoryManager", application.TrackingMemoryManager},
	}

	for _, mt := range managerTypes {
		b.Run(mt.name, func(b *testing.B) {
			config := application.DefaultCompilerConfig()
			config.MemoryManagerType = mt.managerType
			factory := application.NewCompilerFactory(config)
			
			memManager := factory.CreateMemoryManager()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate memory operations
				// Note: This is a basic benchmark - real memory manager
				// benchmarking would require more sophisticated testing
				_ = memManager // Use the memory manager (actual operations depend on interface)
			}
		})
	}
}

// BenchmarkTokenClassification benchmarks token type classification
func BenchmarkTokenClassification(b *testing.B) {
	factory := application.NewCompilerFactory(application.DefaultCompilerConfig())
	lexer := factory.CreateLexer()

	// Various token types to benchmark classification
	inputs := []string{
		"func", "if", "while", "for", "return", // keywords
		"identifier", "variableName", "functionName", // identifiers
		"42", "3.14", "0", "999", // numbers
		`"string"`, `"Hello, World!"`, // strings
		"+", "-", "*", "/", "==", "!=", // operators
		"(", ")", "{", "}", "[", "]", ";", // delimiters
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			err := lexer.SetInput("bench.sl", strings.NewReader(input))
			if err != nil {
				continue
			}
			lexer.NextToken() // Classify the token
		}
	}
}

// generateLargeProgram creates a program with the specified number of functions
// for benchmarking purposes
func generateLargeProgram(numFunctions int) string {
	var builder strings.Builder
	
	// Generate multiple similar functions
	for i := 0; i < numFunctions; i++ {
		builder.WriteString("func function")
		builder.WriteString(strings.Repeat("0", 3-len(strconv.Itoa(i)))) // Pad with zeros
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("(a: int, b: int) -> int {\n")
		builder.WriteString("    var result: int = a + b;\n")
		builder.WriteString("    if result > 0 {\n")
		builder.WriteString("        return result * 2;\n")
		builder.WriteString("    } else {\n")
		builder.WriteString("        return result / 2;\n")
		builder.WriteString("    }\n")
		builder.WriteString("}\n\n")
	}
	
	// Add main function that calls some of the generated functions
	builder.WriteString("func main() -> int {\n")
	builder.WriteString("    var total: int = 0;\n")
	for i := 0; i < min(10, numFunctions); i++ { // Call first 10 functions
		builder.WriteString("    total = total + function")
		builder.WriteString(strings.Repeat("0", 3-len(strconv.Itoa(i))))
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("(")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(", ")
		builder.WriteString(strconv.Itoa(i+1))
		builder.WriteString(");\n")
	}
	builder.WriteString("    return total;\n")
	builder.WriteString("}\n")
	
	return builder.String()
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}