// Package main provides the CLI interface for the StaticLang compiler
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sokoide/llvm5/grammar"
	"github.com/sokoide/llvm5/internal/application"
	"github.com/sokoide/llvm5/internal/domain"
)

// Version information
const (
	Version = "0.1.0"
	Author  = "StaticLang Team"
)

// Command line flags
var (
	inputFiles        = flag.String("i", "", "Input source files (comma-separated)")
	outputFile        = flag.String("o", "", "Output file")
	optimizeLevel     = flag.Int("O", 0, "Optimization level (0-3)")
	debugInfo         = flag.Bool("g", false, "Generate debug information")
	targetTriple      = flag.String("target", "", "Target triple for code generation")
	warningsAsErrors  = flag.Bool("Werror", false, "Treat warnings as errors")
	verbose           = flag.Bool("v", false, "Verbose output")
	showVersion       = flag.Bool("version", false, "Show version information")
	showHelp          = flag.Bool("h", false, "Show this help message")
	useMockComponents = flag.Bool("mock", false, "Use mock components for testing")
	parserDebug       = flag.Int("parser-debug", 0, "Parser debug level (0-4)")
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *showHelp || *inputFiles == "" {
		printUsage()
		return
	}

	grammar.SetDebugLevel(*parserDebug)

	// Parse input files
	files := strings.Split(*inputFiles, ",")
	for i, file := range files {
		files[i] = strings.TrimSpace(file)
	}

	// Validate input files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", file)
			os.Exit(1)
		}
	}

	// Determine output file
	output := *outputFile
	if output == "" {
		if len(files) == 1 {
			ext := filepath.Ext(files[0])
			output = files[0][:len(files[0])-len(ext)] + ".ll" // LLVM IR output
		} else {
			output = "output.ll"
		}
	}

	// Create compiler configuration
	config := application.CompilerConfig{
		UseMockComponents: *useMockComponents,
		MemoryManagerType: application.PooledMemoryManager,
		ErrorReporterType: application.ConsoleErrorReporter,
		CompilationOptions: domain.CompilationOptions{
			OptimizationLevel: *optimizeLevel,
			DebugInfo:         *debugInfo,
			TargetTriple:      *targetTriple,
			OutputPath:        output,
			WarningsAsErrors:  *warningsAsErrors,
		},
		ErrorOutput: os.Stderr,
		Verbose:     *verbose,
	}

	// Create compiler factory
	factory := application.NewCompilerFactory(config)

	// Compile the files
	var err error
	if len(files) == 1 {
		err = compileSingleFile(factory, files[0], output, config)
	} else {
		err = compileMultipleFiles(factory, files, output, config)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Compilation successful. Output written to: %s\n", output)
	}
}

func compileSingleFile(factory *application.CompilerFactory, inputFile, outputFile string, config application.CompilerConfig) error {
	// Open input file
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Create compiler pipeline
	pipeline := factory.CreateCompilerPipeline()

	if config.Verbose {
		fmt.Printf("Compiling: %s -> %s\n", inputFile, outputFile)
	}

	// Compile
	err = pipeline.Compile(inputFile, input, output)
	if err != nil {
		return err
	}

	// Print statistics if verbose
	if config.Verbose {
		if defaultPipeline, ok := pipeline.(*application.DefaultCompilerPipeline); ok {
			stats := defaultPipeline.GetStats()
			fmt.Printf("Compilation statistics:\n")
			fmt.Printf("  Errors: %d\n", stats.ErrorCount)
			fmt.Printf("  Warnings: %d\n", stats.WarningCount)
			fmt.Printf("  Memory used: %d bytes\n", stats.MemoryUsage)
			fmt.Printf("  Nodes created: %d\n", stats.NodesCreated)
		}
	}

	return nil
}

func compileMultipleFiles(factory *application.CompilerFactory, inputFiles []string, outputFile string, config application.CompilerConfig) error {
	// Open all input files
	fileReaders := make(map[string]io.Reader)
	var filesToClose []io.Closer

	defer func() {
		for _, file := range filesToClose {
			file.Close()
		}
	}()

	for _, filename := range inputFiles {
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open input file %s: %w", filename, err)
		}
		fileReaders[filename] = file
		filesToClose = append(filesToClose, file)
	}

	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Create multi-file compiler pipeline
	pipeline := factory.CreateMultiFileCompilerPipeline()

	if config.Verbose {
		fmt.Printf("Compiling multiple files: %v -> %s\n", inputFiles, outputFile)
	}

	// Compile all files
	err = pipeline.CompileFiles(fileReaders, output)
	if err != nil {
		return err
	}

	// Print statistics if verbose
	if config.Verbose {
		stats := pipeline.GetStats()
		fmt.Printf("Multi-file compilation statistics:\n")
		fmt.Printf("  Files compiled: %d\n", len(inputFiles))
		fmt.Printf("  Errors: %d\n", stats.ErrorCount)
		fmt.Printf("  Warnings: %d\n", stats.WarningCount)
		fmt.Printf("  Memory used: %d bytes\n", stats.MemoryUsage)
		fmt.Printf("  Nodes created: %d\n", stats.NodesCreated)
	}

	return nil
}

func printVersion() {
	fmt.Printf("StaticLang Compiler %s\n", Version)
	fmt.Printf("Author: %s\n", Author)
	fmt.Printf("Built with Go %s\n", "1.21+")
}

func printUsage() {
	fmt.Printf("StaticLang Compiler %s\n\n", Version)
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
	fmt.Printf("\nExamples:\n")
	fmt.Printf("  # Compile a single file\n")
	fmt.Printf("  %s -i main.sl -o main.ll\n", os.Args[0])
	fmt.Printf("\n  # Compile multiple files with optimization\n")
	fmt.Printf("  %s -i \"main.sl,lib.sl\" -o program.ll -O 2\n", os.Args[0])
	fmt.Printf("\n  # Compile with debug info and warnings as errors\n")
	fmt.Printf("  %s -i main.sl -o main.ll -g -Werror\n", os.Args[0])
	fmt.Printf("\n  # Use mock components for testing\n")
	fmt.Printf("  %s -i main.sl -o main.ll -mock -v\n", os.Args[0])
}
