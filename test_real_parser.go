package main

import (
	"fmt"
	"strings"

	"github.com/sokoide/llvm5/grammar"
	"github.com/sokoide/llvm5/internal/interfaces"
	"github.com/sokoide/llvm5/lexer"
)

func main() {
	// Test the real parser with the failing program
	input := "func main() { return 42; }"
	fmt.Printf("Testing parser with input: %s\n", input)

	// First, let's see what tokens the lexer produces
	fmt.Printf("\n=== Lexer Analysis ===\n")
	lex1 := lexer.NewLexer()
	err := lex1.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		fmt.Printf("Lexer setup error: %v\n", err)
		return
	}

	tokens := []string{}
	for {
		token := lex1.NextToken()
		fmt.Printf("Token: Type=%d (%s), Value='%s'\n", int(token.Type), getTokenName(token.Type), token.Value)
		tokens = append(tokens, fmt.Sprintf("%s('%s')", getTokenName(token.Type), token.Value))
		if token.Type == 70 { // TokenEOF
			break
		}
	}
	fmt.Printf("Token sequence: %v\n", tokens)

	// Now test the parser
	fmt.Printf("\n=== Parser Analysis ===\n")
	lex := lexer.NewLexer()
	err = lex.SetInput("test.sl", strings.NewReader(input))
	if err != nil {
		fmt.Printf("Lexer setup error: %v\n", err)
		return
	}

	parser := grammar.NewRecursiveDescentParser()
	program, err := parser.Parse(lex)
	if err != nil {
		fmt.Printf("Parser error: %v\n", err)
		return
	}

	if program == nil {
		fmt.Printf("Parser returned nil program\n")
		return
	}

	fmt.Printf("Parser succeeded! Got %d declarations\n", len(program.Declarations))
	for i, decl := range program.Declarations {
		fmt.Printf("Declaration %d: %T\n", i, decl)
	}
}

func getTokenName(t interfaces.TokenType) string {
	switch t {
	case 0:
		return "INT"
	case 1:
		return "FLOAT"
	case 2:
		return "STRING"
	case 3:
		return "BOOL"
	case 4:
		return "IDENTIFIER"
	case 5:
		return "FUNC"
	case 6:
		return "STRUCT"
	case 7:
		return "VAR"
	case 8:
		return "IF"
	case 9:
		return "ELSE"
	case 10:
		return "WHILE"
	case 11:
		return "FOR"
	case 12:
		return "RETURN"
	case 13:
		return "TRUE"
	case 14:
		return "FALSE"
	default:
		return fmt.Sprintf("Token%d", int(t))
	}
}
