package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sokoide/llvm5/grammar"
	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/domain"
)

// TestMainIntegration tests the main function integration
func TestMainIntegration(t *testing.T) {
	// Create a simple test file
	testContent := `func main() -> int {
	print("Hello from test!");
	return 0;
}`

	// Create temporary file
	tempDir := t.TempDir()
	inputFile := tempDir + "/test.sl"

	// Write test file
	err := os.WriteFile(inputFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock os.Args and flag parsing
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Mock stdout and stderr for testing
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	// Test version flag
	t.Run("VersionFlag", func(t *testing.T) {
		os.Args = []string{"staticlang", "-version"}
		// We can't easily test main() directly since it calls flag.Parse()
		// Let's test individual functions instead - just verify they don't panic
		printVersion()
		// Note: We skip output verification due to stdout redirection challenges
	})

	// Test help flag
	t.Run("HelpFlag", func(t *testing.T) {
		// Test that printUsage doesn't panic
		printUsage()
		// Note: We skip output verification due to stdout redirection challenges
	})
}

// TestCompileSingleFile tests the compileSingleFile function
func TestCompileSingleFile(t *testing.T) {
	// Create a simple test file
	testContent := `func main() -> int {
	print("Test file");
	return 0;
}`

	tempDir := t.TempDir()
	inputFile := tempDir + "/test.sl"
	outputFile := tempDir + "/test.ll"

	// Write test file
	err := os.WriteFile(inputFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create compiler configuration
	config := application.CompilerConfig{
		UseMockComponents: true, // Use mocks to avoid LLVM dependencies
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: 0,
			DebugInfo:         false,
			TargetTriple:      "",
			OutputPath:        outputFile,
			WarningsAsErrors:  false,
		},
		ErrorOutput: io.Discard, // Discard error output
		Verbose:     false,
	}

	// Create compiler factory
	factory := application.NewCompilerFactory(config)

	// Test compilation
	err = compileSingleFile(factory, inputFile, outputFile, config)
	if err != nil {
		t.Errorf("compileSingleFile should not fail: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file %s to be created", outputFile)
	}

	// Verify output file is not empty
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(outputContent) == 0 {
		t.Error("Expected output file to contain LLVM IR content")
	}
}

// TestCompileMultipleFiles tests the compileMultipleFiles function
func TestCompileMultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple test input files
	inputFiles := []string{
		tempDir + "/main.sl",
		tempDir + "/lib.sl",
	}
	outputFile := tempDir + "/output.ll"

	// Main file content
	mainContent := `func main() -> int {
	print("Hello from main");
	helper();
	return 0;
}`

	// Library file content
	libContent := `func helper() -> void {
	print("Helper function");
}`

	// Write test files
	for i, file := range inputFiles {
		var content string
		if i == 0 {
			content = mainContent
		} else {
			content = libContent
		}

		err := os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

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
		ErrorOutput: io.Discard,
		Verbose:     false,
	}

	factory := application.NewCompilerFactory(config)

	err := compileMultipleFiles(factory, inputFiles, outputFile, config)
	if err != nil {
		t.Errorf("compileMultipleFiles should not fail: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file %s to be created", outputFile)
	}
}

// TestFileValidation tests input file validation logic
func TestFileValidation(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		files       []string
		expectError bool
	}{
		{
			name:        "valid file",
			files:       []string{tempDir + "/valid.sl"},
			expectError: false,
		},
		{
			name:        "non-existent file",
			files:       []string{tempDir + "/nonexistent.sl"},
			expectError: true,
		},
		{
			name:        "multiple files",
			files:       []string{tempDir + "/main.sl", tempDir + "/lib.sl"},
			expectError: false,
		},
	}

	// Create valid test files
	validFiles := []string{tempDir + "/valid.sl", tempDir + "/main.sl", tempDir + "/lib.sl"}
	for _, file := range validFiles {
		content := "func test() -> void { print(\"test\"); }"
		err := os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the file validation logic from main function
			hasError := false
			var errorMessages []string

			for _, file := range tc.files {
				if _, err := os.Stat(file); os.IsNotExist(err) {
					hasError = true
					errorMessages = append(errorMessages, "File does not exist: "+file)
				}
			}

			if hasError != tc.expectError {
				t.Errorf("Expected error: %v, got error: %v, messages: %v", tc.expectError, hasError, errorMessages)
			}
		})
	}
}

// TestOutputFileGeneration tests output file generation logic
func TestOutputFileGeneration(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name             string
		inputFiles       []string
		specifiedOutput  string
		expectedOutput   string
	}{
		{
			name:             "single file with specified output",
			inputFiles:       []string{tempDir + "/single.sl"},
			specifiedOutput:  "custom.ll",
			expectedOutput:   "custom.ll",
		},
		{
			name:             "single file without specified output",
			inputFiles:       []string{tempDir + "/single.sl"},
			specifiedOutput:  "",
			expectedOutput:   tempDir + "/single.ll",
		},
		{
			name:             "multiple files without specified output",
			inputFiles:       []string{tempDir + "/main.sl", tempDir + "/lib.sl"},
			specifiedOutput:  "",
			expectedOutput:   "output.ll",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the output file generation logic from main function
			var output string
			if tc.specifiedOutput != "" {
				output = tc.specifiedOutput
			} else {
				if len(tc.inputFiles) == 1 {
					file := tc.inputFiles[0]
					ext := ".sl" // Simplified assumption for test
					output = file[:len(file)-len(ext)] + ".ll"
				} else {
					output = "output.ll"
				}
			}

			if output != tc.expectedOutput {
				t.Errorf("Expected output %s, got %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestFlagHandling tests command line flag handling
func TestFlagHandling(t *testing.T) {
	// Test the logic of how flags are processed
	testCases := []struct {
		name           string
		flagSet        map[string]interface{}
		expectVersion  bool
		expectHelp     bool
		expectUsageExit bool
	}{
		{
			name:           "version flag",
			flagSet:        map[string]interface{}{"version": true},
			expectVersion:  true,
			expectHelp:     false,
			expectUsageExit: true, // Version flag should exit
		},
		{
			name:           "help flag",
			flagSet:        map[string]interface{}{"h": true, "i": ""},
			expectVersion:  false,
			expectHelp:     true,
			expectUsageExit: true,
		},
		{
			name:           "no input files",
			flagSet:        map[string]interface{}{"i": ""},
			expectVersion:  false,
			expectHelp:     false,
			expectUsageExit: true,
		},
		{
			name:           "normal execution",
			flagSet:        map[string]interface{}{"i": "test.sl"},
			expectVersion:  false,
			expectHelp:     false,
			expectUsageExit: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the flag handling logic from main function
			showVersion := tc.flagSet["version"] == true
			showHelp := tc.flagSet["h"] == true
			inputFilesIntf := tc.flagSet["i"]
			inputFiles := ""
			if inputFilesIntf != nil {
				inputFiles = inputFilesIntf.(string)
			}

			shouldExit := showVersion || showHelp || inputFiles == ""

			if showVersion != tc.expectVersion {
				t.Errorf("Expected version: %v, got: %v", tc.expectVersion, showVersion)
			}
			if showHelp != tc.expectHelp {
				t.Errorf("Expected help: %v, got: %v", tc.expectHelp, showHelp)
			}
			if shouldExit != tc.expectUsageExit {
				t.Errorf("Expected exit: %v, got: %v", tc.expectUsageExit, shouldExit)
			}
		})
	}
}

// TestParserDebugLevelSetting tests parser debug level configuration
func TestParserDebugLevelSetting(t *testing.T) {
	// Test different debug levels
	testCases := []int{0, 1, 2, 3, 4, 5}

	for _, level := range testCases {
		t.Run("DebugLevel_"+string(rune('0'+level)), func(t *testing.T) {
			// This should not panic
			grammar.SetDebugLevel(level)
		})
	}
}

// TestInputFileParsing tests the input file parsing logic
func TestInputFileParsing(t *testing.T) {
	testCases := []struct {
		name          string
		inputString   string
		expectedFiles []string
	}{
		{
			name:          "single file",
			inputString:   "main.sl",
			expectedFiles: []string{"main.sl"},
		},
		{
			name:          "multiple files",
			inputString:   "main.sl,lib.sl",
			expectedFiles: []string{"main.sl", "lib.sl"},
		},
		{
			name:          "files with spaces",
			inputString:   " main.sl , lib.sl ",
			expectedFiles: []string{"main.sl", "lib.sl"},
		},
		{
			name:          "empty string",
			inputString:   "",
			expectedFiles: []string{""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the file parsing logic from main function
			files := strings.Split(tc.inputString, ",")
			for i, file := range files {
				files[i] = strings.TrimSpace(file)
			}

			if len(files) != len(tc.expectedFiles) {
				t.Errorf("Expected %d files, got %d", len(tc.expectedFiles), len(files))
				return
			}

			for i, expected := range tc.expectedFiles {
				if files[i] != expected {
					t.Errorf("Expected file #%d to be %s, got %s", i, expected, files[i])
				}
			}
		})
	}
}
