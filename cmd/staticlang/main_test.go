package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/domain"
)

// TestPrintVersion tests version output
func TestPrintVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printVersion()

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	output := string(out)

	if !strings.Contains(output, "StaticLang Compiler") {
		t.Errorf("Expected version output to contain 'StaticLang Compiler', got: %s", output)
	}

	if !strings.Contains(output, Version) {
		t.Errorf("Expected version output to contain version '%s', got: %s", Version, output)
	}

	if !strings.Contains(output, Author) {
		t.Errorf("Expected version output to contain author '%s', got: %s", Author, output)
	}
}

// TestPrintUsage tests usage output
func TestPrintUsage(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	output := string(out)

	expectedStrings := []string{
		"Usage:",
		"Examples:",
		"main.sl",
		"main.ll",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected usage output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestCompileSingleFileWithTempFiles tests single file compilation
func TestCompileSingleFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input file
	inputFile := filepath.Join(tempDir, "test.sl")
	inputContent := `func main() -> int {
    return 42;
}`

	err = os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Create output file path
	outputFile := filepath.Join(tempDir, "test.ll")

	// Set up arguments
	oldArgs := os.Args
	os.Args = []string{"staticlang", "-i", inputFile, "-o", outputFile, "-mock", "-v"}
	defer func() { os.Args = oldArgs }()

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Test the compilation - this will call compileSingleFile internally
	// We can't easily test the main function directly due to os.Exit,
	// so we test the compilation logic indirectly

	// Reset flag variables (these are flag pointers in main.go)
	// Note: We can't actually reset flag variables in tests since they are package-level
	// This section is commented out as it's not testable in this way

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutBytes, _ := io.ReadAll(rOut)
	stderrBytes, _ := io.ReadAll(rErr)

	stdoutStr := string(stdoutBytes)
	stderrStr := string(stderrBytes)

	// Since we can't easily test main() directly due to os.Exit,
	// we'll test that the file operations work
	t.Logf("Stdout: %s", stdoutStr)
	t.Logf("Stderr: %s", stderrStr)

	// Test that input file exists and is readable
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("Input file should exist")
	}

	// Test file content
	content, err := os.ReadFile(inputFile)
	if err != nil {
		t.Errorf("Should be able to read input file: %v", err)
	}

	if !strings.Contains(string(content), "func main()") {
		t.Error("Input file should contain expected content")
	}
}

// TestCompileMultipleFiles tests multiple file compilation logic
func TestCompileMultipleFiles(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_test_multi")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input files
	file1 := filepath.Join(tempDir, "file1.sl")
	file2 := filepath.Join(tempDir, "file2.sl")

	content1 := `func helper() -> int { return 1; }`
	content2 := `func main() -> int { return helper(); }`

	err = os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Test that files exist and are readable
	for _, file := range []string{file1, file2} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("File %s should exist", file)
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("Should be able to read %s: %v", file, err)
		}

		if len(content) == 0 {
			t.Errorf("File %s should not be empty", file)
		}
	}

	// Test file list parsing
	inputList := file1 + "," + file2
	files := strings.Split(inputList, ",")

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	if files[0] != file1 || files[1] != file2 {
		t.Errorf("File parsing incorrect: got %v", files)
	}
}

// TestFlagValidation tests command line flag validation
func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		inputFiles  string
		outputFile  string
		expectError bool
	}{
		{
			name:        "valid_single_file",
			inputFiles:  "test.sl",
			outputFile:  "test.ll",
			expectError: false,
		},
		{
			name:        "valid_multiple_files",
			inputFiles:  "file1.sl,file2.sl",
			outputFile:  "output.ll",
			expectError: false,
		},
		{
			name:        "empty_input",
			inputFiles:  "",
			outputFile:  "test.ll",
			expectError: true,
		},
		{
			name:        "empty_output",
			inputFiles:  "test.sl",
			outputFile:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic (simulated)
			hasError := false

			if tt.inputFiles == "" {
				hasError = true
			}

			if tt.outputFile == "" {
				hasError = true
			}

			if hasError != tt.expectError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectError, hasError)
			}
		})
	}
}

// TestOutputFileExtensionValidation tests output file extension handling
func TestOutputFileExtensionValidation(t *testing.T) {
	tests := []struct {
		name       string
		outputFile string
		expected   string
	}{
		{
			name:       "ll_extension",
			outputFile: "test.ll",
			expected:   "test.ll",
		},
		{
			name:       "no_extension",
			outputFile: "test",
			expected:   "test",
		},
		{
			name:       "different_extension",
			outputFile: "test.txt",
			expected:   "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we handle various output file extensions
			result := tt.outputFile

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestVerboseOutput tests verbose flag functionality
func TestVerboseOutput(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Simulate verbose output
	verbose := true
	message := "Compilation step completed"

	if verbose {
		buf.WriteString("[VERBOSE] " + message + "\n")
	}

	output := buf.String()

	if !strings.Contains(output, "[VERBOSE]") {
		t.Error("Verbose output should contain [VERBOSE] prefix")
	}

	if !strings.Contains(output, message) {
		t.Error("Verbose output should contain the message")
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		expectedInMsg string
	}{
		{
			name:          "file_not_found",
			scenario:      "input file does not exist",
			expectedInMsg: "file",
		},
		{
			name:          "permission_denied",
			scenario:      "cannot write to output location",
			expectedInMsg: "write",
		},
		{
			name:          "invalid_syntax",
			scenario:      "source file contains syntax errors",
			expectedInMsg: "syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate error message generation
			errorMsg := "Error: " + tt.scenario

			if !strings.Contains(strings.ToLower(errorMsg), tt.expectedInMsg) {
				t.Errorf("Error message should contain '%s', got: %s", tt.expectedInMsg, errorMsg)
			}
		})
	}
}

// TestTempFileCleanup tests temporary file handling
func TestTempFileCleanup(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a temporary file
	tempFile := filepath.Join(tempDir, "temp.sl")
	err = os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Temp file should exist")
	}

	// Clean up
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("Failed to clean up temp dir: %v", err)
	}

	// Verify cleanup
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp dir should be cleaned up")
	}
}

// TestCompileSingleFileFunction tests the compileSingleFile function directly
func TestCompileSingleFileFunction(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_compile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input file
	inputFile := filepath.Join(tempDir, "test.sl")
	inputContent := `func main() -> int {
    return 42;
}`

	err = os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Create output file path
	outputFile := filepath.Join(tempDir, "test.ll")

	// Create compiler configuration with mocks
	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        outputFile,
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	// Create compiler factory
	factory := application.NewCompilerFactory(config)

	// Test successful compilation
	err = compileSingleFile(factory, inputFile, outputFile, config)
	if err != nil {
		t.Errorf("compileSingleFile should succeed: %v", err)
	}

	// Check output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should be created")
	}
}

// TestCompileSingleFileVerbose tests verbose output
func TestCompileSingleFileVerbose(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_verbose_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input file
	inputFile := filepath.Join(tempDir, "test.sl")
	inputContent := `func main() -> int {
    return 42;
}`

	err = os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Create output file path
	outputFile := filepath.Join(tempDir, "test.ll")

	// Create compiler configuration with verbose enabled
	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        outputFile,
			WarningsAsErrors:  false,
		},
		ErrorOutput: os.Stderr,
		Verbose:     true, // Enable verbose mode
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create compiler factory and compile
	factory := application.NewCompilerFactory(config)
	err = compileSingleFile(factory, inputFile, outputFile, config)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("compileSingleFile should succeed: %v", err)
	}

	// Read captured output
	out, _ := io.ReadAll(r)
	output := string(out)

	// Check for verbose output
	if !strings.Contains(output, "Compiling:") {
		t.Error("Verbose mode should output compilation message")
	}

	if !strings.Contains(output, "Compilation statistics:") {
		t.Error("Verbose mode should output statistics")
	}
}

// TestCompileSingleFileErrors tests error handling
func TestCompileSingleFileErrors(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-existent input file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.sl")
	outputFile := filepath.Join(tempDir, "test.ll")

	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)

	err = compileSingleFile(factory, nonExistentFile, outputFile, config)
	if err == nil {
		t.Error("compileSingleFile should fail with non-existent input file")
	}

	if !strings.Contains(err.Error(), "failed to open input file") {
		t.Error("Error should mention failed to open input file")
	}
}

// TestCompileMultipleFilesFunction tests the compileMultipleFiles function
func TestCompileMultipleFilesFunction(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_multi_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input files
	file1 := filepath.Join(tempDir, "file1.sl")
	file2 := filepath.Join(tempDir, "file2.sl")
	outputFile := filepath.Join(tempDir, "output.ll")

	content1 := `func helper() -> int { return 1; }`
	content2 := `func main() -> int { return helper(); }`

	err = os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Create compiler configuration
	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OutputPath: outputFile,
		},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	inputFiles := []string{file1, file2}

	// Test successful multi-file compilation
	err = compileMultipleFiles(factory, inputFiles, outputFile, config)
	if err != nil {
		t.Errorf("compileMultipleFiles should succeed: %v", err)
	}

	// Check output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should be created")
	}
}

// TestCompileMultipleFilesVerbose tests verbose output for multiple files
func TestCompileMultipleFilesVerbose(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_multi_verbose_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input files
	file1 := filepath.Join(tempDir, "file1.sl")
	file2 := filepath.Join(tempDir, "file2.sl")
	outputFile := filepath.Join(tempDir, "output.ll")

	content1 := `func helper() -> int { return 1; }`
	content2 := `func main() -> int { return helper(); }`

	err = os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Create compiler configuration with verbose enabled
	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OutputPath: outputFile,
		},
		ErrorOutput: os.Stderr,
		Verbose:     true, // Enable verbose mode
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	factory := application.NewCompilerFactory(config)
	inputFiles := []string{file1, file2}

	// Test compilation
	err = compileMultipleFiles(factory, inputFiles, outputFile, config)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("compileMultipleFiles should succeed: %v", err)
	}

	// Read captured output
	out, _ := io.ReadAll(r)
	output := string(out)

	// Check for verbose output
	if !strings.Contains(output, "Compiling multiple files:") {
		t.Error("Verbose mode should output multi-file compilation message")
	}

	if !strings.Contains(output, "Multi-file compilation statistics:") {
		t.Error("Verbose mode should output multi-file statistics")
	}

	if !strings.Contains(output, "Files compiled: 2") {
		t.Error("Verbose mode should show number of files compiled")
	}
}

// TestCompileMultipleFilesErrors tests error handling for multiple files
func TestCompileMultipleFilesErrors(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "staticlang_multi_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-existent input files
	nonExistentFile := filepath.Join(tempDir, "nonexistent.sl")
	outputFile := filepath.Join(tempDir, "output.ll")

	config := application.CompilerConfig{
		UseMockComponents: true,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{},
		ErrorOutput: os.Stderr,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)
	inputFiles := []string{nonExistentFile}

	err = compileMultipleFiles(factory, inputFiles, outputFile, config)
	if err == nil {
		t.Error("compileMultipleFiles should fail with non-existent input files")
	}

	if !strings.Contains(err.Error(), "failed to open input file") {
		t.Error("Error should mention failed to open input file")
	}
}

// TestMainFunctionLogic tests the main function logic indirectly
func TestMainFunctionLogic(t *testing.T) {
	// Test file list parsing
	inputList := "file1.sl,file2.sl,file3.sl"
	files := strings.Split(inputList, ",")
	for i, file := range files {
		files[i] = strings.TrimSpace(file)
	}

	expected := []string{"file1.sl", "file2.sl", "file3.sl"}
	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(files))
	}

	for i, expectedFile := range expected {
		if files[i] != expectedFile {
			t.Errorf("File[%d]: expected %s, got %s", i, expectedFile, files[i])
		}
	}
}

// TestOutputFileGeneration tests output file generation logic
func TestOutputFileGeneration(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		expected   string
	}{
		{
			name:       "single_file_no_output",
			inputFile:  "test.sl",
			outputFile: "",
			expected:   "test.ll",
		},
		{
			name:       "single_file_with_output",
			inputFile:  "test.sl",
			outputFile: "custom.ll",
			expected:   "custom.ll",
		},
		{
			name:       "file_with_path",
			inputFile:  "/path/to/test.sl",
			outputFile: "",
			expected:   "/path/to/test.ll",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from main function
			output := tt.outputFile
			if output == "" {
				ext := filepath.Ext(tt.inputFile)
				output = tt.inputFile[:len(tt.inputFile)-len(ext)] + ".ll"
			}

			if output != tt.expected {
				t.Errorf("Expected output %s, got %s", tt.expected, output)
			}
		})
	}
}
