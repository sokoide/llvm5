// Package infrastructure contains implementations of infrastructure concerns
package infrastructure

import (
	"fmt"

	"github.com/sokoide/llvm5/staticlang/internal/domain"
	"github.com/sokoide/llvm5/staticlang/internal/interfaces"
)

// DefaultSymbolTable implements the SymbolTable interface
type DefaultSymbolTable struct {
	currentScope *interfaces.Scope
	globalScope  *interfaces.Scope
}

// NewDefaultSymbolTable creates a new symbol table with a global scope
func NewDefaultSymbolTable() *DefaultSymbolTable {
	globalScope := &interfaces.Scope{
		Parent:   nil,
		Symbols:  make(map[string]*interfaces.Symbol),
		Children: make([]*interfaces.Scope, 0),
		Level:    0,
	}

	return &DefaultSymbolTable{
		currentScope: globalScope,
		globalScope:  globalScope,
	}
}

// EnterScope creates a new scope as a child of the current scope
func (st *DefaultSymbolTable) EnterScope() *interfaces.Scope {
	newScope := &interfaces.Scope{
		Parent:   st.currentScope,
		Symbols:  make(map[string]*interfaces.Symbol),
		Children: make([]*interfaces.Scope, 0),
		Level:    st.currentScope.Level + 1,
	}

	st.currentScope.Children = append(st.currentScope.Children, newScope)
	st.currentScope = newScope

	return newScope
}

// ExitScope exits the current scope and returns to the parent scope
func (st *DefaultSymbolTable) ExitScope() {
	if st.currentScope.Parent != nil {
		st.currentScope = st.currentScope.Parent
	}
}

// GetCurrentScope returns the current scope
func (st *DefaultSymbolTable) GetCurrentScope() *interfaces.Scope {
	return st.currentScope
}

// DeclareSymbol declares a new symbol in the current scope
func (st *DefaultSymbolTable) DeclareSymbol(name string, symbolType domain.Type, kind interfaces.SymbolKind, location domain.SourceRange) (*interfaces.Symbol, error) {
	// Check if symbol already exists in current scope
	if _, exists := st.currentScope.Symbols[name]; exists {
		return nil, fmt.Errorf("symbol '%s' already declared in current scope", name)
	}

	symbol := &interfaces.Symbol{
		Name:     name,
		Type:     symbolType,
		Kind:     kind,
		Location: location,
		Scope:    st.currentScope,
	}

	st.currentScope.Symbols[name] = symbol
	return symbol, nil
}

// LookupSymbol looks up a symbol in the current scope chain
func (st *DefaultSymbolTable) LookupSymbol(name string) (*interfaces.Symbol, bool) {
	scope := st.currentScope
	for scope != nil {
		if symbol, exists := scope.Symbols[name]; exists {
			return symbol, true
		}
		scope = scope.Parent
	}
	return nil, false
}

// LookupSymbolInScope looks up a symbol in a specific scope only
func (st *DefaultSymbolTable) LookupSymbolInScope(name string, scope *interfaces.Scope) (*interfaces.Symbol, bool) {
	symbol, exists := scope.Symbols[name]
	return symbol, exists
}

// GetGlobalScope returns the global scope
func (st *DefaultSymbolTable) GetGlobalScope() *interfaces.Scope {
	return st.globalScope
}

// GetSymbolsInScope returns all symbols in the given scope
func (st *DefaultSymbolTable) GetSymbolsInScope(scope *interfaces.Scope) []*interfaces.Symbol {
	symbols := make([]*interfaces.Symbol, 0, len(scope.Symbols))
	for _, symbol := range scope.Symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// GetAllSymbols returns all symbols in all scopes
func (st *DefaultSymbolTable) GetAllSymbols() []*interfaces.Symbol {
	var symbols []*interfaces.Symbol
	st.collectSymbolsFromScope(st.globalScope, &symbols)
	return symbols
}

func (st *DefaultSymbolTable) collectSymbolsFromScope(scope *interfaces.Scope, symbols *[]*interfaces.Symbol) {
	for _, symbol := range scope.Symbols {
		*symbols = append(*symbols, symbol)
	}
	for _, child := range scope.Children {
		st.collectSymbolsFromScope(child, symbols)
	}
}

// Reset resets the symbol table to its initial state
func (st *DefaultSymbolTable) Reset() {
	st.globalScope = &interfaces.Scope{
		Parent:   nil,
		Symbols:  make(map[string]*interfaces.Symbol),
		Children: make([]*interfaces.Scope, 0),
		Level:    0,
	}
	st.currentScope = st.globalScope
}
