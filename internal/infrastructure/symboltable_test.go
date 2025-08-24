package infrastructure

import (
	"testing"

	"github.com/sokoide/llvm5/internal/domain"
	"github.com/sokoide/llvm5/internal/interfaces"
)

// TestSymbolTable_BasicOperations tests basic symbol table operations
func TestSymbolTable_BasicOperations(t *testing.T) {
	st := NewSymbolTable()

	// Create a test symbol
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Test declaration
	symbol, err := st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("DeclareSymbol failed: %v", err)
	}

	if symbol.Name != "x" {
		t.Errorf("Symbol name: got %q, expected %q", symbol.Name, "x")
	}

	// Test lookup
	foundSymbol, found := st.LookupSymbol("x")
	if !found {
		t.Error("Should find declared symbol")
	}

	if foundSymbol.Name != "x" {
		t.Errorf("Symbol name: got %q, expected %q", foundSymbol.Name, "x")
	}

	if foundSymbol.Type != intType {
		t.Error("Symbol type should match declared type")
	}

	// Test lookup of non-existent symbol
	_, notFound := st.LookupSymbol("nonexistent")
	if notFound {
		t.Error("Should not find non-existent symbol")
	}
}

// TestSymbolTable_ScopeManagement tests scope entering and exiting
func TestSymbolTable_ScopeManagement(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Test initial scope level
	if st.GetCurrentScope().Level != 0 {
		t.Error("Initial scope should be 0 (global)")
	}

	// Declare symbol in global scope
	_, err := st.DeclareSymbol("global", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare global symbol: %v", err)
	}

	// Enter a new scope
	st.EnterScope()
	if st.GetCurrentScope().Level != 1 {
		t.Error("Should be in scope level 1")
	}

	// Declare symbol in local scope
	_, err = st.DeclareSymbol("local", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local symbol: %v", err)
	}

	// Test that both symbols are accessible
	_, found := st.LookupSymbol("global")
	if !found {
		t.Error("Should find global symbol from inner scope")
	}

	_, found = st.LookupSymbol("local")
	if !found {
		t.Error("Should find local symbol in current scope")
	}

	// Exit scope
	st.ExitScope()
	if st.GetCurrentScope().Level != 0 {
		t.Error("Should be back to global scope")
	}

	// Test that local symbol is no longer accessible
	_, found = st.LookupSymbol("local")
	if found {
		t.Error("Should not find local symbol after exiting scope")
	}

	// But global symbol should still be accessible
	_, found = st.LookupSymbol("global")
	if !found {
		t.Error("Should still find global symbol")
	}
}

// TestSymbolTable_Shadowing tests variable shadowing across scopes
func TestSymbolTable_Shadowing(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	stringType := &domain.BasicType{Kind: domain.StringType}
	location := domain.SourceRange{}

	// Declare 'x' in global scope as int
	_, err := st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare global x: %v", err)
	}

	// Enter inner scope
	st.EnterScope()

	// Declare 'x' in local scope as string (shadowing)
	_, err = st.DeclareSymbol("x", stringType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local x: %v", err)
	}

	// Lookup should find the local (shadowing) symbol
	foundX, found := st.LookupSymbol("x")
	if !found {
		t.Error("Should find symbol x")
	}

	if foundX.Type != stringType {
		t.Error("Should find local (string) version of x, not global (int)")
	}

	// Exit scope
	st.ExitScope()

	// Now should find global version again
	foundX, found = st.LookupSymbol("x")
	if !found {
		t.Error("Should find global symbol x after exiting scope")
	}

	if foundX.Type != intType {
		t.Error("Should find global (int) version of x after exiting scope")
	}
}

// TestSymbolTable_RedeclarationError tests redeclaration error handling
func TestSymbolTable_RedeclarationError(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Declare symbol
	_, err := st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("First declaration should succeed: %v", err)
	}

	// Try to redeclare same symbol in same scope
	_, err = st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err == nil {
		t.Error("Redeclaration in same scope should fail")
	}

	// But redeclaration in different scope should work
	st.EnterScope()
	_, err = st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Error("Redeclaration in different scope should succeed")
	}
}

// TestSymbolTable_LookupSymbolInScope tests scope-specific lookup
func TestSymbolTable_LookupSymbolInScope(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Declare in global scope
	_, err := st.DeclareSymbol("global", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare global symbol: %v", err)
	}

	globalScope := st.GetCurrentScope()

	// Enter local scope and declare local symbol
	st.EnterScope()
	_, err = st.DeclareSymbol("local", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local symbol: %v", err)
	}

	currentScope := st.GetCurrentScope()

	// Test lookup in current scope only
	_, found := st.LookupSymbolInScope("local", currentScope)
	if !found {
		t.Error("Should find local symbol in current scope")
	}

	// Global symbol should not be found in current scope lookup
	_, found = st.LookupSymbolInScope("global", currentScope)
	if found {
		t.Error("Should not find global symbol in current scope only lookup")
	}

	// But should find global symbol in global scope lookup
	_, found = st.LookupSymbolInScope("global", globalScope)
	if !found {
		t.Error("Should find global symbol in global scope lookup")
	}
}

// TestSymbolTable_GetSymbolsInScope tests retrieving all symbols in a scope
func TestSymbolTable_GetSymbolsInScope(t *testing.T) {
	st := NewSymbolTable().(*DefaultSymbolTable)
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Declare multiple symbols in global scope
	symbolNames := []string{"a", "b", "c"}

	for _, name := range symbolNames {
		_, err := st.DeclareSymbol(name, intType, interfaces.VariableSymbol, location)
		if err != nil {
			t.Errorf("Failed to declare symbol %s: %v", name, err)
		}
	}

	// Get all symbols in global scope
	globalSymbols := st.GetSymbolsInScope(st.GetCurrentScope())
	if len(globalSymbols) != 3 {
		t.Errorf("Expected 3 symbols in global scope, got %d", len(globalSymbols))
	}

	// Check that all declared symbols are present
	foundNames := make(map[string]bool)
	for _, symbol := range globalSymbols {
		foundNames[symbol.Name] = true
	}

	for _, expectedName := range symbolNames {
		if !foundNames[expectedName] {
			t.Errorf("Symbol %q not found in scope symbols", expectedName)
		}
	}

	// Enter new scope and declare one symbol
	st.EnterScope()
	_, err := st.DeclareSymbol("local", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local symbol: %v", err)
	}

	// Local scope should have only one symbol
	localSymbols := st.GetSymbolsInScope(st.GetCurrentScope())
	if len(localSymbols) != 1 {
		t.Errorf("Expected 1 symbol in local scope, got %d", len(localSymbols))
	}

	if localSymbols[0].Name != "local" {
		t.Errorf("Expected local symbol, got %q", localSymbols[0].Name)
	}
}

// TestSymbolTable_GetAllSymbols tests retrieving all symbols across all scopes
func TestSymbolTable_GetAllSymbols(t *testing.T) {
	st := NewSymbolTable().(*DefaultSymbolTable)
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Declare in global scope
	_, err := st.DeclareSymbol("global1", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare global1: %v", err)
	}
	_, err = st.DeclareSymbol("global2", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare global2: %v", err)
	}

	// Enter scope and declare local symbols
	st.EnterScope()
	_, err = st.DeclareSymbol("local1", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local1: %v", err)
	}
	_, err = st.DeclareSymbol("local2", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare local2: %v", err)
	}

	// Get all symbols
	allSymbols := st.GetAllSymbols()
	if len(allSymbols) != 4 {
		t.Errorf("Expected 4 symbols total, got %d", len(allSymbols))
	}

	// Check that all symbols are present
	symbolNames := make(map[string]bool)
	for _, symbol := range allSymbols {
		symbolNames[symbol.Name] = true
	}

	expectedNames := []string{"global1", "global2", "local1", "local2"}
	for _, name := range expectedNames {
		if !symbolNames[name] {
			t.Errorf("Symbol %q not found in all symbols", name)
		}
	}
}

// TestSymbolTable_Reset tests resetting the symbol table
func TestSymbolTable_Reset(t *testing.T) {
	st := NewSymbolTable().(*DefaultSymbolTable)
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Declare some symbols
	_, err := st.DeclareSymbol("x", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare x: %v", err)
	}
	st.EnterScope()
	_, err = st.DeclareSymbol("y", intType, interfaces.VariableSymbol, location)
	if err != nil {
		t.Errorf("Failed to declare y: %v", err)
	}

	// Verify symbols exist
	_, found := st.LookupSymbol("x")
	if !found {
		t.Error("Symbol x should exist before reset")
	}
	_, found = st.LookupSymbol("y")
	if !found {
		t.Error("Symbol y should exist before reset")
	}

	if st.GetCurrentScope().Level != 1 {
		t.Error("Should be in nested scope before reset")
	}

	// Reset
	st.Reset()

	// Verify clean state
	if st.GetCurrentScope().Level != 0 {
		t.Error("Should be back to global scope after reset")
	}

	_, found = st.LookupSymbol("x")
	if found {
		t.Error("Symbol x should not exist after reset")
	}
	_, found = st.LookupSymbol("y")
	if found {
		t.Error("Symbol y should not exist after reset")
	}

	if len(st.GetAllSymbols()) != 0 {
		t.Error("Should have no symbols after reset")
	}
}

// TestSymbolTable_NestedScopes tests deeply nested scopes
func TestSymbolTable_NestedScopes(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	location := domain.SourceRange{}

	// Create nested scopes
	scopes := []string{"function", "block", "if", "while"}
	expectedDepth := 0

	for i, scopeName := range scopes {
		// Enter the next scope first
		st.EnterScope()
		expectedDepth = i + 1

		if st.GetCurrentScope().Level != expectedDepth {
			t.Errorf("Expected scope depth %d, got %d", expectedDepth, st.GetCurrentScope().Level)
		}

		// Then declare a symbol at this level
		_, err := st.DeclareSymbol(scopeName, intType, interfaces.VariableSymbol, location)
		if err != nil {
			t.Errorf("Failed to declare symbol %s: %v", scopeName, err)
		}
	}

	// All symbols should be accessible from deepest scope
	for _, scopeName := range scopes {
		_, found := st.LookupSymbol(scopeName)
		if !found {
			t.Errorf("Should find symbol %q from nested scope", scopeName)
		}
	}

	// Exit all scopes one by one
	for i := len(scopes) - 1; i >= 0; i-- {
		st.ExitScope()
		expectedDepth--

		if st.GetCurrentScope().Level != expectedDepth {
			t.Errorf("Expected scope depth %d after exiting, got %d", expectedDepth, st.GetCurrentScope().Level)
		}

		// Symbol from exited scope should no longer be accessible
		_, found := st.LookupSymbol(scopes[i])
		if found {
			t.Errorf("Should not find symbol %q after exiting its scope", scopes[i])
		}

		// But symbols from outer scopes should still be accessible
		for j := 0; j < i; j++ {
			_, found := st.LookupSymbol(scopes[j])
			if !found {
				t.Errorf("Should still find symbol %q from outer scope", scopes[j])
			}
		}
	}
}

// TestSymbolTable_SymbolKinds tests different symbol kinds
func TestSymbolTable_SymbolKinds(t *testing.T) {
	st := NewSymbolTable()
	intType := &domain.BasicType{Kind: domain.IntType}
	funcType := &domain.FunctionType{
		ParameterTypes: []domain.Type{intType},
		ReturnType:     intType,
	}
	location := domain.SourceRange{}

	symbols := []struct {
		name string
		typ  domain.Type
		kind interfaces.SymbolKind
	}{
		{"var1", intType, interfaces.VariableSymbol},
		{"func1", funcType, interfaces.FunctionSymbol},
		{"param1", intType, interfaces.ParameterSymbol},
	}

	// Declare all symbols
	for _, sym := range symbols {
		_, err := st.DeclareSymbol(sym.name, sym.typ, sym.kind, location)
		if err != nil {
			t.Errorf("Failed to declare %v symbol %q: %v", sym.kind, sym.name, err)
		}
	}

	// Verify all can be looked up
	for _, expectedSymbol := range symbols {
		found, exists := st.LookupSymbol(expectedSymbol.name)
		if !exists {
			t.Errorf("Should find %v symbol %q", expectedSymbol.kind, expectedSymbol.name)
			continue
		}

		if found.Kind != expectedSymbol.kind {
			t.Errorf("Symbol %q kind: got %v, expected %v", expectedSymbol.name, found.Kind, expectedSymbol.kind)
		}

		if found.Type != expectedSymbol.typ {
			t.Errorf("Symbol %q type mismatch", expectedSymbol.name)
		}
	}
}

// TestSymbolTable_GlobalScope tests global scope operations
func TestSymbolTable_GlobalScope(t *testing.T) {
	st := NewSymbolTable().(*DefaultSymbolTable)

	globalScope := st.GetGlobalScope()
	if globalScope == nil {
		t.Error("Global scope should not be nil")
	}

	// Global scope should initially be current scope
	if st.GetCurrentScope().Level != 0 {
		t.Error("Initial current scope should be global (0)")
	}

	// After entering and exiting scopes, global scope should remain accessible
	st.EnterScope()
	st.EnterScope()
	st.ExitScope()
	st.ExitScope()

	if st.GetCurrentScope().Level != 0 {
		t.Error("Should be back to global scope")
	}

	newGlobalScope := st.GetGlobalScope()
	if newGlobalScope != globalScope {
		t.Error("Global scope reference should remain consistent")
	}
}
