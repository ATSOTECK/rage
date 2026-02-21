package compiler

import "github.com/ATSOTECK/rage/internal/model"

// Scope types for variable resolution
type ScopeType int

const (
	ScopeModule ScopeType = iota
	ScopeFunction
	ScopeClass
	ScopeComprehension
)

// SymbolScope indicates where a variable is stored
type SymbolScope int

const (
	ScopeLocal SymbolScope = iota
	ScopeGlobal
	ScopeBuiltin
	ScopeFree
	ScopeCell
)

// Symbol represents a variable in a scope
type Symbol struct {
	Name          string
	Scope         SymbolScope
	Index         int
	OriginalIndex int // Original local index before becoming a cell (-1 if N/A)
}

// SymbolTable tracks variables in a scope
type SymbolTable struct {
	outer     *SymbolTable
	symbols   map[string]*Symbol
	freeSyms  []*Symbol
	cellSyms  []*Symbol // Variables captured by inner functions
	numDefs   int
	scopeType ScopeType
	globals   map[string]bool
	nonlocals map[string]bool
}

func NewSymbolTable(scopeType ScopeType, outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer:     outer,
		symbols:   make(map[string]*Symbol),
		scopeType: scopeType,
		globals:   make(map[string]bool),
		nonlocals: make(map[string]bool),
	}
}

func (st *SymbolTable) Define(name string) *Symbol {
	sym := &Symbol{Name: name, Scope: ScopeLocal, Index: st.numDefs, OriginalIndex: -1}
	st.symbols[name] = sym
	st.numDefs++
	return sym
}

func (st *SymbolTable) DefineGlobal(name string) *Symbol {
	st.globals[name] = true
	sym := &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}
	st.symbols[name] = sym
	return sym
}

func (st *SymbolTable) DefineNonlocal(name string) *Symbol {
	st.nonlocals[name] = true

	// Resolve through outer scopes to find the variable and set up cell/free linkage
	if st.outer != nil {
		outerSym, ok := st.outer.Resolve(name)
		if ok && outerSym.Scope != ScopeGlobal && outerSym.Scope != ScopeBuiltin {
			// Mark it as a cell in the outer scope if it's a local
			if outerSym.Scope == ScopeLocal {
				st.outer.MarkAsCell(name)
			}

			// Create a free variable in our scope
			free := &Symbol{Name: name, Scope: ScopeFree, Index: len(st.freeSyms)}
			st.freeSyms = append(st.freeSyms, free)
			st.symbols[name] = free
			return free
		}
	}

	// Fallback: couldn't resolve, create with -1 index (will error at runtime)
	sym := &Symbol{Name: name, Scope: ScopeFree, Index: -1}
	st.symbols[name] = sym
	return sym
}

// GetEnclosingScopeType returns the scope type of the first non-comprehension enclosing scope.
func (st *SymbolTable) GetEnclosingScopeType() ScopeType {
	enclosing := st.outer
	for enclosing != nil && enclosing.scopeType == ScopeComprehension {
		enclosing = enclosing.outer
	}
	if enclosing == nil {
		return ScopeModule // Default to module if no enclosing scope
	}
	return enclosing.scopeType
}

// IsInsideClass returns true if this scope or any ancestor is a class scope.
// Also returns the class scope's SymbolTable if found.
func (st *SymbolTable) IsInsideClass() (*SymbolTable, bool) {
	current := st.outer // Skip the current scope (which would be the function)
	for current != nil {
		if current.scopeType == ScopeClass {
			return current, true
		}
		current = current.outer
	}
	return nil, false
}

// DefineInEnclosingScope defines a variable in the first non-comprehension outer scope.
// This is used for walrus operator (:=) in comprehensions, where the variable should
// be accessible in the enclosing scope per PEP 572.
func (st *SymbolTable) DefineInEnclosingScope(name string) *Symbol {
	// Find the first non-comprehension scope
	enclosing := st.outer
	var intermediateScopes []*SymbolTable
	currentScope := st

	for enclosing != nil && enclosing.scopeType == ScopeComprehension {
		intermediateScopes = append(intermediateScopes, currentScope)
		currentScope = enclosing
		enclosing = enclosing.outer
	}

	if enclosing == nil {
		// No enclosing scope found, define locally
		return st.Define(name)
	}

	// Define the variable in the enclosing scope
	var enclosingSym *Symbol
	if existing, ok := enclosing.symbols[name]; ok {
		// Variable already exists in enclosing scope
		enclosingSym = existing
	} else {
		// Create new variable in enclosing scope
		enclosingSym = &Symbol{Name: name, Scope: ScopeLocal, Index: enclosing.numDefs, OriginalIndex: -1}
		enclosing.symbols[name] = enclosingSym
		enclosing.numDefs++
	}

	// Mark it as a cell in the enclosing scope if it's local
	if enclosingSym.Scope == ScopeLocal {
		enclosing.MarkAsCell(name)
	}

	// Create free variable references in all intermediate scopes (including currentScope)
	// and the original scope (st)
	allScopes := append([]*SymbolTable{currentScope}, intermediateScopes...)
	allScopes = append(allScopes, st)

	for _, scope := range allScopes {
		if _, exists := scope.symbols[name]; !exists {
			free := &Symbol{Name: name, Scope: ScopeFree, Index: len(scope.freeSyms)}
			scope.freeSyms = append(scope.freeSyms, free)
			scope.symbols[name] = free
		}
	}

	// Return the symbol from our scope (st) for storing
	return st.symbols[name]
}

func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	// Check if declared global
	if st.globals[name] {
		return &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}, true
	}

	// Check local symbols
	if sym, ok := st.symbols[name]; ok {
		return sym, true
	}

	// For module scope, treat as global
	if st.scopeType == ScopeModule {
		return &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}, true
	}

	// Check outer scopes
	if st.outer != nil {
		sym, ok := st.outer.Resolve(name)
		if !ok {
			// Not found anywhere, treat as global
			return &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}, true
		}

		if sym.Scope == ScopeGlobal || sym.Scope == ScopeBuiltin {
			return sym, true
		}

		// The outer scope has this variable (either as local, free, or cell)
		// We need to capture it as a free variable in our scope
		// And ensure the outer scope makes it available (marks it as a cell if needed)

		// If the outer scope has this as a local, it needs to become a cell
		// so inner scopes can capture it
		if sym.Scope == ScopeLocal {
			// Mark it as a cell in the outer scope so we can capture it
			st.outer.MarkAsCell(name)
		}

		// Create a free variable in our scope
		free := &Symbol{Name: name, Scope: ScopeFree, Index: len(st.freeSyms)}
		st.freeSyms = append(st.freeSyms, free)
		st.symbols[name] = free
		return free, true
	}

	return &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}, true
}

// MarkAsCell marks a variable as needing to be stored in a cell
// (because it's captured by an inner scope)
func (st *SymbolTable) MarkAsCell(name string) {
	if sym, ok := st.symbols[name]; ok {
		if sym.Scope == ScopeLocal {
			// Convert from local to cell, preserving original index
			sym.OriginalIndex = sym.Index // Remember the original local index
			sym.Scope = ScopeCell
			sym.Index = len(st.cellSyms)
			st.cellSyms = append(st.cellSyms, sym)
		} else if sym.Scope == ScopeFree {
			// This is a free variable that's also captured by an even-inner scope
			// It stays as a free variable but also needs to be passed on
			// The VM will handle this case by looking in the closure
		}
	}
}

// classNeedsClassCell checks if any method in a class body uses 'super' or '__class__'.
// This is used to determine if the class body needs a __class__ cell variable.
func classNeedsClassCell(stmts []model.Stmt) bool {
	for _, stmt := range stmts {
		if funcDef, ok := stmt.(*model.FunctionDef); ok {
			// Check if this method uses super() or __class__
			if usesSuperOrClass(funcDef.Body) {
				return true
			}
		}
	}
	return false
}

// usesSuperOrClass checks if an AST node uses 'super' or '__class__' references.
// This is used to determine if a method needs the implicit __class__ closure variable.
func usesSuperOrClass(stmts []model.Stmt) bool {
	w := astWalker{
		exprMatch: func(expr model.Expr) bool {
			if id, ok := expr.(*model.Identifier); ok {
				return id.Name == "super" || id.Name == "__class__"
			}
			return false
		},
		enterComprehensionElts: true,
	}
	return w.walkStmts(stmts)
}
