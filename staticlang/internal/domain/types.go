// Package domain contains the core domain types and interfaces for the StaticLang compiler
package domain

import (
	"fmt"
)

// SourcePosition represents a position in the source code
type SourcePosition struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (pos SourcePosition) String() string {
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
}

// SourceRange represents a range in the source code
type SourceRange struct {
	Start SourcePosition
	End   SourcePosition
}

func (r SourceRange) String() string {
	if r.Start.Filename == r.End.Filename {
		if r.Start.Line == r.End.Line {
			return fmt.Sprintf("%s:%d:%d-%d", r.Start.Filename, r.Start.Line, r.Start.Column, r.End.Column)
		}
		return fmt.Sprintf("%s:%d:%d-%d:%d", r.Start.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
	}
	return fmt.Sprintf("%s-%s", r.Start.String(), r.End.String())
}

// CompilerError represents different types of compilation errors
type CompilerError struct {
	Type     ErrorType
	Message  string
	Location SourceRange
	Context  string
	Hints    []string
}

type ErrorType int

const (
	LexicalError ErrorType = iota
	SyntaxError
	SemanticError
	TypeCheckError
	CodeGenError
	InternalError
)

func (e CompilerError) Error() string {
	return fmt.Sprintf("%s: %s at %s", e.Type, e.Message, e.Location)
}

func (et ErrorType) String() string {
	switch et {
	case LexicalError:
		return "Lexical Error"
	case SyntaxError:
		return "Syntax Error"
	case SemanticError:
		return "Semantic Error"
	case TypeCheckError:
		return "Type Error"
	case CodeGenError:
		return "Code Generation Error"
	case InternalError:
		return "Internal Error"
	default:
		return "Unknown Error"
	}
}

// ErrorReporter defines the interface for error reporting
type ErrorReporter interface {
	ReportError(err CompilerError)
	ReportWarning(warning CompilerError)
	HasErrors() bool
	HasWarnings() bool
	GetErrors() []CompilerError
	GetWarnings() []CompilerError
	Clear()
}

// CompilationContext holds the shared context for compilation
type CompilationContext struct {
	SourceFiles   map[string][]byte
	ErrorReporter ErrorReporter
	Options       CompilationOptions
}

// CompilationOptions holds compiler configuration
type CompilationOptions struct {
	OptimizationLevel int
	DebugInfo         bool
	TargetTriple      string
	OutputPath        string
	WarningsAsErrors  bool
}
