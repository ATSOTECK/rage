package compiler

import (
	"fmt"
	"strconv"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// CompileError represents a compilation error
type CompileError struct {
	Pos     model.Position
	Message string
}

func (e CompileError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.Filename, e.Pos.Line, e.Pos.Column, e.Message)
}

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
	// Will be resolved later
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
	for _, stmt := range stmts {
		if usesSuperOrClassStmt(stmt) {
			return true
		}
	}
	return false
}

func usesSuperOrClassStmt(stmt model.Stmt) bool {
	switch s := stmt.(type) {
	case *model.ExprStmt:
		return usesSuperOrClassExpr(s.Value)
	case *model.Return:
		if s.Value != nil {
			return usesSuperOrClassExpr(s.Value)
		}
	case *model.Assign:
		for _, target := range s.Targets {
			if usesSuperOrClassExpr(target) {
				return true
			}
		}
		return usesSuperOrClassExpr(s.Value)
	case *model.AugAssign:
		return usesSuperOrClassExpr(s.Target) || usesSuperOrClassExpr(s.Value)
	case *model.AnnAssign:
		if s.Value != nil && usesSuperOrClassExpr(s.Value) {
			return true
		}
	case *model.If:
		if usesSuperOrClassExpr(s.Test) {
			return true
		}
		if usesSuperOrClass(s.Body) || usesSuperOrClass(s.OrElse) {
			return true
		}
	case *model.While:
		if usesSuperOrClassExpr(s.Test) {
			return true
		}
		if usesSuperOrClass(s.Body) || usesSuperOrClass(s.OrElse) {
			return true
		}
	case *model.For:
		if usesSuperOrClassExpr(s.Iter) {
			return true
		}
		if usesSuperOrClass(s.Body) || usesSuperOrClass(s.OrElse) {
			return true
		}
	case *model.Try:
		if usesSuperOrClass(s.Body) || usesSuperOrClass(s.OrElse) || usesSuperOrClass(s.FinalBody) {
			return true
		}
		for _, handler := range s.Handlers {
			if usesSuperOrClass(handler.Body) {
				return true
			}
		}
	case *model.With:
		for _, item := range s.Items {
			if usesSuperOrClassExpr(item.ContextExpr) {
				return true
			}
		}
		return usesSuperOrClass(s.Body)
	case *model.Raise:
		if s.Exc != nil && usesSuperOrClassExpr(s.Exc) {
			return true
		}
		if s.Cause != nil && usesSuperOrClassExpr(s.Cause) {
			return true
		}
	case *model.Assert:
		if usesSuperOrClassExpr(s.Test) {
			return true
		}
		if s.Msg != nil && usesSuperOrClassExpr(s.Msg) {
			return true
		}
	}
	return false
}

func usesSuperOrClassExpr(expr model.Expr) bool {
	switch e := expr.(type) {
	case *model.Identifier:
		return e.Name == "super" || e.Name == "__class__"
	case *model.Call:
		if usesSuperOrClassExpr(e.Func) {
			return true
		}
		for _, arg := range e.Args {
			if usesSuperOrClassExpr(arg) {
				return true
			}
		}
		for _, kw := range e.Keywords {
			if usesSuperOrClassExpr(kw.Value) {
				return true
			}
		}
	case *model.Attribute:
		return usesSuperOrClassExpr(e.Value)
	case *model.Subscript:
		if usesSuperOrClassExpr(e.Value) {
			return true
		}
		return usesSuperOrClassExpr(e.Slice)
	case *model.BinaryOp:
		return usesSuperOrClassExpr(e.Left) || usesSuperOrClassExpr(e.Right)
	case *model.UnaryOp:
		return usesSuperOrClassExpr(e.Operand)
	case *model.BoolOp:
		for _, v := range e.Values {
			if usesSuperOrClassExpr(v) {
				return true
			}
		}
	case *model.Compare:
		if usesSuperOrClassExpr(e.Left) {
			return true
		}
		for _, comp := range e.Comparators {
			if usesSuperOrClassExpr(comp) {
				return true
			}
		}
	case *model.IfExpr:
		return usesSuperOrClassExpr(e.Test) || usesSuperOrClassExpr(e.Body) || usesSuperOrClassExpr(e.OrElse)
	case *model.List:
		for _, elt := range e.Elts {
			if usesSuperOrClassExpr(elt) {
				return true
			}
		}
	case *model.Tuple:
		for _, elt := range e.Elts {
			if usesSuperOrClassExpr(elt) {
				return true
			}
		}
	case *model.Dict:
		for _, k := range e.Keys {
			if k != nil && usesSuperOrClassExpr(k) {
				return true
			}
		}
		for _, v := range e.Values {
			if usesSuperOrClassExpr(v) {
				return true
			}
		}
	case *model.Set:
		for _, elt := range e.Elts {
			if usesSuperOrClassExpr(elt) {
				return true
			}
		}
	case *model.ListComp:
		if usesSuperOrClassExpr(e.Elt) {
			return true
		}
	case *model.DictComp:
		if usesSuperOrClassExpr(e.Key) || usesSuperOrClassExpr(e.Value) {
			return true
		}
	case *model.SetComp:
		if usesSuperOrClassExpr(e.Elt) {
			return true
		}
	case *model.GeneratorExpr:
		if usesSuperOrClassExpr(e.Elt) {
			return true
		}
	case *model.Starred:
		return usesSuperOrClassExpr(e.Value)
	case *model.Await:
		return usesSuperOrClassExpr(e.Value)
	case *model.NamedExpr:
		return usesSuperOrClassExpr(e.Value)
	}
	return false
}

// Compiler compiles AST to bytecode
type Compiler struct {
	code           *runtime.CodeObject
	symbolTable    *SymbolTable
	errors         []CompileError
	loopStack      []loopInfo
	filename       string
	optimizer      *Optimizer
	currentLine    int // Current source line being compiled
	lineStartOffset int // Bytecode offset where current line started
}

type loopInfo struct {
	startOffset   int
	breakJumps    []int
	continueJumps []int
	isForLoop     bool // true for 'for' loops (iterator on stack), false for 'while' loops
}

// NewCompiler creates a new compiler
func NewCompiler(filename string) *Compiler {
	code := &runtime.CodeObject{
		Name:      "<module>",
		Filename:  filename,
		FirstLine: 1,
	}
	return &Compiler{
		code:        code,
		symbolTable: NewSymbolTable(ScopeModule, nil),
		filename:    filename,
		optimizer:   NewOptimizer(),
	}
}

// Compile compiles a module to bytecode
func (c *Compiler) Compile(module *model.Module) (*runtime.CodeObject, []CompileError) {
	stmts := module.Body

	for _, stmt := range stmts {
		c.compileStmt(stmt)
	}

	// Add implicit return None at end of module
	c.emit(runtime.OpLoadNone) // Use optimized opcode
	c.emit(runtime.OpReturn)

	// Finalize line number table
	c.finishLineTable()

	// Build names and varnames lists
	c.finalizeCode()

	// Apply peephole optimizations
	c.optimizer.PeepholeOptimize(c.code)

	return c.code, c.errors
}

func (c *Compiler) error(pos model.Position, format string, args ...any) {
	c.errors = append(c.errors, CompileError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	})
}

// Bytecode emission helpers

func (c *Compiler) emit(op runtime.Opcode) int {
	offset := len(c.code.Code)
	c.code.Code = append(c.code.Code, byte(op))
	return offset
}

func (c *Compiler) emitArg(op runtime.Opcode, arg int) int {
	// Check if argument fits in 16 bits (signed or unsigned)
	// Allow -32768 to 65535 range: negative values wrap to unsigned 16-bit
	// (e.g., -1 becomes 65535/0xFFFF for special sentinel values)
	if arg < -32768 || arg > 65535 {
		// For very large arguments, extended arg support could be added here
		// For now, report an error for overflow cases
		c.errors = append(c.errors, CompileError{
			Message: fmt.Sprintf("bytecode argument %d exceeds 16-bit limit", arg),
		})
		// Clamp to valid range to avoid corrupted bytecode
		if arg < -32768 {
			arg = -32768
		} else {
			arg = 65535
		}
	}
	offset := len(c.code.Code)
	c.code.Code = append(c.code.Code, byte(op), byte(arg), byte(arg>>8))
	return offset
}

func (c *Compiler) emitJump(op runtime.Opcode) int {
	return c.emitArg(op, 0) // Placeholder, will be patched
}

func (c *Compiler) patchJump(offset int, target int) {
	c.code.Code[offset+1] = byte(target)
	c.code.Code[offset+2] = byte(target >> 8)
}

func (c *Compiler) currentOffset() int {
	return len(c.code.Code)
}

func (c *Compiler) addConstant(value any) int {
	// Skip deduplication for slice types (they can't be compared with ==)
	switch value.(type) {
	case []string, []any, []int, []float64:
		// Don't deduplicate slices, just add them
	default:
		for i, v := range c.code.Constants {
			if v == value {
				return i
			}
		}
	}
	c.code.Constants = append(c.code.Constants, value)
	return len(c.code.Constants) - 1
}

func (c *Compiler) addName(name string) int {
	for i, n := range c.code.Names {
		if n == name {
			return i
		}
	}
	c.code.Names = append(c.code.Names, name)
	return len(c.code.Names) - 1
}

func (c *Compiler) emitLoadConst(value any) {
	idx := c.addConstant(value)
	c.emitArg(runtime.OpLoadConst, idx)
}

// setLine updates the current line number being compiled.
// When the line changes, it records the previous line's bytecode range.
func (c *Compiler) setLine(line int) {
	if line <= 0 {
		return
	}
	if line != c.currentLine {
		// Finish the previous line's entry if there was one
		if c.currentLine > 0 {
			currentOffset := len(c.code.Code)
			if currentOffset > c.lineStartOffset {
				c.code.LineNoTab = append(c.code.LineNoTab, runtime.LineEntry{
					StartOffset: c.lineStartOffset,
					EndOffset:   currentOffset,
					Line:        c.currentLine,
				})
			}
		}
		c.currentLine = line
		c.lineStartOffset = len(c.code.Code)
	}
}

// finishLineTable finalizes the line number table with the last entry
func (c *Compiler) finishLineTable() {
	if c.currentLine > 0 {
		currentOffset := len(c.code.Code)
		if currentOffset > c.lineStartOffset {
			c.code.LineNoTab = append(c.code.LineNoTab, runtime.LineEntry{
				StartOffset: c.lineStartOffset,
				EndOffset:   currentOffset,
				Line:        c.currentLine,
			})
		}
	}
}

// Statement compilation

func (c *Compiler) compileStmt(stmt model.Stmt) {
	// Track line number for this statement
	if stmt != nil {
		c.setLine(stmt.Pos().Line)
	}

	switch s := stmt.(type) {
	case *model.ExprStmt:
		c.compileExpr(s.Value)
		c.emit(runtime.OpPop)

	case *model.Assign:
		c.compileExpr(s.Value)
		// Duplicate value for multiple targets
		for i := 0; i < len(s.Targets)-1; i++ {
			c.emit(runtime.OpDup)
		}
		for _, target := range s.Targets {
			c.compileStore(target)
		}

	case *model.AugAssign:
		c.compileAugAssign(s)

	case *model.AnnAssign:
		if s.Value != nil {
			c.compileExpr(s.Value)
			c.compileStore(s.Target)
		}
		// Type annotations are not emitted at runtime

	case *model.If:
		c.compileIf(s)

	case *model.While:
		c.compileWhile(s)

	case *model.For:
		c.compileFor(s)

	case *model.FunctionDef:
		c.compileFunctionDef(s)

	case *model.ClassDef:
		c.compileClassDef(s)

	case *model.Return:
		if s.Value != nil {
			c.compileExpr(s.Value)
		} else {
			c.emitLoadConst(nil)
		}
		c.emit(runtime.OpReturn)

	case *model.Pass:
		// No-op

	case *model.Break:
		if len(c.loopStack) == 0 {
			c.error(s.StartPos, "'break' outside loop")
			return
		}
		// Pop the iterator from stack if breaking from a for loop
		loop := c.loopStack[len(c.loopStack)-1]
		if loop.isForLoop {
			c.emit(runtime.OpPop)
		}
		jump := c.emitJump(runtime.OpJump)
		c.loopStack[len(c.loopStack)-1].breakJumps = append(
			c.loopStack[len(c.loopStack)-1].breakJumps, jump)

	case *model.Continue:
		if len(c.loopStack) == 0 {
			c.error(s.StartPos, "'continue' outside loop")
			return
		}
		jump := c.emitJump(runtime.OpJump)
		c.loopStack[len(c.loopStack)-1].continueJumps = append(
			c.loopStack[len(c.loopStack)-1].continueJumps, jump)

	case *model.Global:
		for _, name := range s.Names {
			c.symbolTable.DefineGlobal(name.Name)
		}

	case *model.Nonlocal:
		for _, name := range s.Names {
			c.symbolTable.DefineNonlocal(name.Name)
		}

	case *model.Import:
		for _, alias := range s.Names {
			nameIdx := c.addName(alias.Name.Name)
			c.emitLoadConst(0)   // level
			c.emitLoadConst(nil) // fromlist
			c.emitArg(runtime.OpImportName, nameIdx)

			// For "import outer.inner", store as "outer" (root module)
			// For "import outer.inner as x", store as "x"
			storeName := alias.Name.Name
			if alias.AsName != nil {
				storeName = alias.AsName.Name
			} else {
				// Extract just the first part of a dotted name
				if dotIdx := indexByte(storeName, '.'); dotIdx >= 0 {
					storeName = storeName[:dotIdx]
				}
			}
			storeIdx := c.addName(storeName)
			c.emitArg(runtime.OpStoreName, storeIdx)
		}

	case *model.ImportFrom:
		level := s.Level
		c.emitLoadConst(level)

		// Build fromlist
		fromNames := make([]string, len(s.Names))
		for i, alias := range s.Names {
			fromNames[i] = alias.Name.Name
		}
		c.emitLoadConst(fromNames)

		var modName string
		if s.Module != nil {
			modName = s.Module.Name
		}
		nameIdx := c.addName(modName)
		c.emitArg(runtime.OpImportName, nameIdx)

		for _, alias := range s.Names {
			if alias.Name.Name == "*" {
				c.emit(runtime.OpImportStar)
			} else {
				fromIdx := c.addName(alias.Name.Name)
				c.emitArg(runtime.OpImportFrom, fromIdx)

				storeName := alias.Name.Name
				if alias.AsName != nil {
					storeName = alias.AsName.Name
				}
				storeIdx := c.addName(storeName)
				c.emitArg(runtime.OpStoreName, storeIdx)
			}
		}
		c.emit(runtime.OpPop) // Pop the module

	case *model.Raise:
		argc := 0
		if s.Exc != nil {
			c.compileExpr(s.Exc)
			argc = 1
			if s.Cause != nil {
				c.compileExpr(s.Cause)
				argc = 2
			}
		}
		c.emitArg(runtime.OpRaiseVarargs, argc)

	case *model.Try:
		c.compileTry(s)

	case *model.With:
		c.compileWith(s)

	case *model.Assert:
		c.compileExpr(s.Test)
		skipJump := c.emitJump(runtime.OpPopJumpIfTrue)

		// Load AssertionError
		nameIdx := c.addName("AssertionError")
		c.emitArg(runtime.OpLoadGlobal, nameIdx)

		if s.Msg != nil {
			c.compileExpr(s.Msg)
			c.emitArg(runtime.OpCall, 1)
		} else {
			c.emitArg(runtime.OpCall, 0)
		}
		c.emitArg(runtime.OpRaiseVarargs, 1)

		c.patchJump(skipJump, c.currentOffset())

	case *model.Delete:
		for _, target := range s.Targets {
			c.compileDelete(target)
		}

	case *model.Match:
		c.compileMatch(s)

	default:
		c.error(stmt.Pos(), "unsupported statement type: %T", stmt)
	}
}

// Expression compilation

func (c *Compiler) compileExpr(expr model.Expr) {

	switch e := expr.(type) {
	case *model.IntLit:
		val, _ := strconv.ParseInt(e.Value, 0, 64)
		c.emitLoadConst(val)

	case *model.FloatLit:
		val, _ := strconv.ParseFloat(e.Value, 64)
		c.emitLoadConst(val)

	case *model.StringLit:
		c.emitLoadConst(e.Value)

	case *model.FStringLit:
		c.compileFString(e)

	case *model.BytesLit:
		c.emitLoadConst([]byte(e.Value))

	case *model.BoolLit:
		c.emitLoadConst(e.Value)

	case *model.NoneLit:
		c.emitLoadConst(nil)

	case *model.Ellipsis:
		c.emitLoadConst("...") // Placeholder for Ellipsis object

	case *model.Identifier:
		c.compileLoad(e.Name)

	case *model.UnaryOp:
		c.compileExpr(e.Operand)
		switch e.Op {
		case model.TK_Plus:
			c.emit(runtime.OpUnaryPositive)
		case model.TK_Minus:
			c.emit(runtime.OpUnaryNegative)
		case model.TK_Not:
			c.emit(runtime.OpUnaryNot)
		case model.TK_Tilde:
			c.emit(runtime.OpUnaryInvert)
		}

	case *model.BinaryOp:
		c.compileExpr(e.Left)
		c.compileExpr(e.Right)
		c.emitBinaryOp(e.Op)

	case *model.BoolOp:
		c.compileBoolOp(e)

	case *model.Compare:
		c.compileCompare(e)

	case *model.Call:
		c.compileCall(e)

	case *model.Attribute:
		c.compileExpr(e.Value)
		nameIdx := c.addName(e.Attr.Name)
		c.emitArg(runtime.OpLoadAttr, nameIdx)

	case *model.Subscript:
		c.compileExpr(e.Value)
		c.compileExpr(e.Slice)
		c.emit(runtime.OpBinarySubscr)

	case *model.Slice:
		// Load slice function first, then arguments
		nameIdx := c.addName("slice")
		c.emitArg(runtime.OpLoadGlobal, nameIdx)
		if e.Lower != nil {
			c.compileExpr(e.Lower)
		} else {
			c.emitLoadConst(nil)
		}
		if e.Upper != nil {
			c.compileExpr(e.Upper)
		} else {
			c.emitLoadConst(nil)
		}
		if e.Step != nil {
			c.compileExpr(e.Step)
		} else {
			c.emitLoadConst(nil)
		}
		c.emitArg(runtime.OpCall, 3)

	case *model.List:
		for _, elt := range e.Elts {
			c.compileExpr(elt)
		}
		c.emitArg(runtime.OpBuildList, len(e.Elts))

	case *model.Tuple:
		for _, elt := range e.Elts {
			c.compileExpr(elt)
		}
		c.emitArg(runtime.OpBuildTuple, len(e.Elts))

	case *model.Dict:
		for i := range e.Keys {
			if e.Keys[i] != nil {
				c.compileExpr(e.Keys[i])
				c.compileExpr(e.Values[i])
			} else {
				// **unpacking
				c.compileExpr(e.Values[i])
			}
		}
		c.emitArg(runtime.OpBuildMap, len(e.Keys))

	case *model.Set:
		for _, elt := range e.Elts {
			c.compileExpr(elt)
		}
		c.emitArg(runtime.OpBuildSet, len(e.Elts))

	case *model.IfExpr:
		c.compileExpr(e.Test)
		falseJump := c.emitJump(runtime.OpPopJumpIfFalse)
		c.compileExpr(e.Body)
		endJump := c.emitJump(runtime.OpJump)
		c.patchJump(falseJump, c.currentOffset())
		c.compileExpr(e.OrElse)
		c.patchJump(endJump, c.currentOffset())

	case *model.Lambda:
		c.compileLambda(e)

	case *model.ListComp:
		c.compileListComp(e)

	case *model.SetComp:
		c.compileSetComp(e)

	case *model.DictComp:
		c.compileDictComp(e)

	case *model.GeneratorExpr:
		c.compileGeneratorExpr(e)

	case *model.Yield:
		// Yield expression: push value and suspend
		if e.Value != nil {
			c.compileExpr(e.Value)
		} else {
			c.emit(runtime.OpLoadNone)
		}
		c.emit(runtime.OpYieldValue)

	case *model.YieldFrom:
		// Yield from: delegate to sub-iterator
		c.compileExpr(e.Value)
		c.emit(runtime.OpGetIter)
		c.emit(runtime.OpYieldFrom)

	case *model.Await:
		// Await: get awaitable and yield from it
		c.compileExpr(e.Value)
		c.emit(runtime.OpGetAwaitable)
		c.emit(runtime.OpYieldFrom)

	case *model.Starred:
		c.compileExpr(e.Value)
		// Starred unpacking handled by context

	case *model.NamedExpr:
		c.compileExpr(e.Value)
		c.emit(runtime.OpDup)
		// In comprehension scope, walrus operator stores to enclosing scope (PEP 572)
		if c.symbolTable.scopeType == ScopeComprehension {
			// Find the enclosing non-comprehension scope type
			enclosingType := c.symbolTable.GetEnclosingScopeType()
			if enclosingType == ScopeModule {
				// For module-level, store as global
				idx := c.addName(e.Target.Name)
				c.emitArg(runtime.OpStoreGlobal, idx)
			} else {
				// For function-level, use cells/deref
				sym := c.symbolTable.DefineInEnclosingScope(e.Target.Name)
				c.emitArg(runtime.OpStoreDeref, sym.Index)
			}
		} else {
			c.compileStore(e.Target)
		}

	default:
		c.error(expr.Pos(), "unsupported expression type: %T", expr)
	}
}

func (c *Compiler) emitBinaryOp(op model.TokenKind) {
	switch op {
	case model.TK_Plus:
		c.emit(runtime.OpBinaryAdd)
	case model.TK_Minus:
		c.emit(runtime.OpBinarySubtract)
	case model.TK_Star:
		c.emit(runtime.OpBinaryMultiply)
	case model.TK_Slash:
		c.emit(runtime.OpBinaryDivide)
	case model.TK_DoubleSlash:
		c.emit(runtime.OpBinaryFloorDiv)
	case model.TK_Percent:
		c.emit(runtime.OpBinaryModulo)
	case model.TK_DoubleStar:
		c.emit(runtime.OpBinaryPower)
	case model.TK_At:
		c.emit(runtime.OpBinaryMatMul)
	case model.TK_LShift:
		c.emit(runtime.OpBinaryLShift)
	case model.TK_RShift:
		c.emit(runtime.OpBinaryRShift)
	case model.TK_Ampersand:
		c.emit(runtime.OpBinaryAnd)
	case model.TK_Pipe:
		c.emit(runtime.OpBinaryOr)
	case model.TK_Caret:
		c.emit(runtime.OpBinaryXor)
	}
}

func (c *Compiler) emitCompareOp(op model.TokenKind) {
	switch op {
	case model.TK_Equal:
		c.emit(runtime.OpCompareEq)
	case model.TK_NotEqual:
		c.emit(runtime.OpCompareNe)
	case model.TK_Less:
		c.emit(runtime.OpCompareLt)
	case model.TK_LessEqual:
		c.emit(runtime.OpCompareLe)
	case model.TK_Greater:
		c.emit(runtime.OpCompareGt)
	case model.TK_GreaterEqual:
		c.emit(runtime.OpCompareGe)
	case model.TK_Is:
		c.emit(runtime.OpCompareIs)
	case model.TK_IsNot:
		c.emit(runtime.OpCompareIsNot)
	case model.TK_In:
		c.emit(runtime.OpCompareIn)
	case model.TK_NotIn:
		c.emit(runtime.OpCompareNotIn)
	}
}

func (c *Compiler) compileBoolOp(e *model.BoolOp) {
	if e.Op == model.TK_And {
		// Short-circuit and: if any is false, skip rest
		var jumpOffsets []int
		for _, val := range e.Values[:len(e.Values)-1] {
			c.compileExpr(val)
			jumpOffsets = append(jumpOffsets, c.emitJump(runtime.OpJumpIfFalseOrPop))
		}
		c.compileExpr(e.Values[len(e.Values)-1])
		endOffset := c.currentOffset()
		for _, offset := range jumpOffsets {
			c.patchJump(offset, endOffset)
		}
	} else {
		// Short-circuit or: if any is true, skip rest
		var jumpOffsets []int
		for _, val := range e.Values[:len(e.Values)-1] {
			c.compileExpr(val)
			jumpOffsets = append(jumpOffsets, c.emitJump(runtime.OpJumpIfTrueOrPop))
		}
		c.compileExpr(e.Values[len(e.Values)-1])
		endOffset := c.currentOffset()
		for _, offset := range jumpOffsets {
			c.patchJump(offset, endOffset)
		}
	}
}

func (c *Compiler) compileCompare(e *model.Compare) {
	// Handle chained comparisons: a < b < c becomes (a < b) and (b < c)
	c.compileExpr(e.Left)

	if len(e.Ops) == 1 {
		// Simple case: just one comparison
		c.compileExpr(e.Comparators[0])
		c.emitCompareOp(e.Ops[0])
		return
	}

	// Chained comparisons need special handling
	var jumpOffsets []int
	for i, op := range e.Ops {
		c.compileExpr(e.Comparators[i])
		if i < len(e.Ops)-1 {
			c.emit(runtime.OpDup)
			c.emit(runtime.OpRot3)
		}
		c.emitCompareOp(op)
		if i < len(e.Ops)-1 {
			jumpOffsets = append(jumpOffsets, c.emitJump(runtime.OpJumpIfFalseOrPop))
		}
	}

	// Patch all jump targets to end
	endOffset := c.currentOffset()
	for _, offset := range jumpOffsets {
		c.patchJump(offset, endOffset)
	}
}

func (c *Compiler) compileCall(e *model.Call) {
	c.compileExpr(e.Func)

	// Compile positional arguments
	for _, arg := range e.Args {
		c.compileExpr(arg)
	}

	if len(e.Keywords) > 0 {
		// Compile keyword argument values
		kwNames := make([]string, 0, len(e.Keywords))
		for _, kw := range e.Keywords {
			c.compileExpr(kw.Value)
			if kw.Arg != nil {
				kwNames = append(kwNames, kw.Arg.Name)
			}
		}
		// Push tuple of keyword names at the end
		c.emitLoadConst(kwNames)
		c.emitArg(runtime.OpCallKw, len(e.Args)+len(e.Keywords))
	} else {
		c.emitArg(runtime.OpCall, len(e.Args))
	}
}

func (c *Compiler) compileFString(e *model.FStringLit) {
	if len(e.Parts) == 0 {
		// Empty f-string
		c.emitLoadConst("")
		return
	}

	// Compile each part
	for i, part := range e.Parts {
		if part.IsExpr {
			// Load str builtin
			strIdx := c.addName("str")
			c.emitArg(runtime.OpLoadGlobal, strIdx)

			// Compile the expression
			c.compileExpr(part.Expr)

			// Call str(expr)
			c.emitArg(runtime.OpCall, 1)
		} else {
			// Load literal string
			c.emitLoadConst(part.Value)
		}

		// Concatenate with previous part
		if i > 0 {
			c.emit(runtime.OpBinaryAdd)
		}
	}
}

// Variable access compilation

func (c *Compiler) compileLoad(name string) {
	sym, _ := c.symbolTable.Resolve(name)
	switch sym.Scope {
	case ScopeLocal:
		c.emitArg(runtime.OpLoadFast, sym.Index)
	case ScopeGlobal, ScopeBuiltin:
		idx := c.addName(name)
		c.emitArg(runtime.OpLoadGlobal, idx)
	case ScopeFree:
		c.emitArg(runtime.OpLoadDeref, sym.Index)
	case ScopeCell:
		c.emitArg(runtime.OpLoadDeref, sym.Index)
	}
}

func (c *Compiler) compileStore(target model.Expr) {
	switch t := target.(type) {
	case *model.Identifier:
		sym, found := c.symbolTable.Resolve(t.Name)
		// In function scope, if variable not found or resolved as global but not explicitly declared global,
		// define it as a local variable
		if c.symbolTable.scopeType == ScopeFunction || c.symbolTable.scopeType == ScopeComprehension {
			if !found || (sym.Scope == ScopeGlobal && !c.symbolTable.globals[t.Name]) {
				sym = c.symbolTable.Define(t.Name)
			}
		}
		switch sym.Scope {
		case ScopeLocal:
			if sym.Index < 0 {
				// Define new local
				sym = c.symbolTable.Define(t.Name)
			}
			c.emitArg(runtime.OpStoreFast, sym.Index)
		case ScopeGlobal:
			idx := c.addName(t.Name)
			c.emitArg(runtime.OpStoreGlobal, idx)
		case ScopeFree, ScopeCell:
			c.emitArg(runtime.OpStoreDeref, sym.Index)
		default:
			// New variable in module scope
			idx := c.addName(t.Name)
			c.emitArg(runtime.OpStoreName, idx)
		}

	case *model.Attribute:
		c.compileExpr(t.Value)
		idx := c.addName(t.Attr.Name)
		c.emitArg(runtime.OpStoreAttr, idx)

	case *model.Subscript:
		c.compileExpr(t.Value)
		c.compileExpr(t.Slice)
		c.emit(runtime.OpStoreSubscr)

	case *model.Tuple, *model.List:
		var elts []model.Expr
		if tup, ok := t.(*model.Tuple); ok {
			elts = tup.Elts
		} else {
			elts = t.(*model.List).Elts
		}
		c.emitArg(runtime.OpUnpackSequence, len(elts))
		for _, elt := range elts {
			c.compileStore(elt)
		}

	case *model.Starred:
		// Starred in unpacking context
		c.compileStore(t.Value)
	}
}

func (c *Compiler) compileDelete(target model.Expr) {
	switch t := target.(type) {
	case *model.Identifier:
		sym, _ := c.symbolTable.Resolve(t.Name)
		switch sym.Scope {
		case ScopeLocal:
			c.emitArg(runtime.OpDeleteFast, sym.Index)
		case ScopeGlobal:
			idx := c.addName(t.Name)
			c.emitArg(runtime.OpDeleteGlobal, idx)
		default:
			idx := c.addName(t.Name)
			c.emitArg(runtime.OpDeleteName, idx)
		}

	case *model.Attribute:
		c.compileExpr(t.Value)
		idx := c.addName(t.Attr.Name)
		c.emitArg(runtime.OpDeleteAttr, idx)

	case *model.Subscript:
		c.compileExpr(t.Value)
		c.compileExpr(t.Slice)
		c.emit(runtime.OpDeleteSubscr)
	}
}

func (c *Compiler) compileAugAssign(s *model.AugAssign) {
	// Track if this is a subscript (needs special handling to avoid double evaluation)
	_, isSubscript := s.Target.(*model.Subscript)

	// Load target
	switch t := s.Target.(type) {
	case *model.Identifier:
		c.compileLoad(t.Name)
	case *model.Attribute:
		c.compileExpr(t.Value)
		c.emit(runtime.OpDup)
		idx := c.addName(t.Attr.Name)
		c.emitArg(runtime.OpLoadAttr, idx)
	case *model.Subscript:
		// For subscript: push object and index, duplicate both, then get value
		// Stack sequence: [obj, idx] -> [obj, idx, obj, idx] -> [obj, idx, value]
		c.compileExpr(t.Value)
		c.compileExpr(t.Slice)
		c.emit(runtime.OpDup2) // Duplicate top two: [obj, idx, obj, idx]
		c.emit(runtime.OpBinarySubscr) // Get value: [obj, idx, value]
	}

	// Compile the value
	c.compileExpr(s.Value)

	// Emit inplace operation
	switch s.Op {
	case model.TK_PlusAssign:
		c.emit(runtime.OpInplaceAdd)
	case model.TK_MinusAssign:
		c.emit(runtime.OpInplaceSubtract)
	case model.TK_StarAssign:
		c.emit(runtime.OpInplaceMultiply)
	case model.TK_SlashAssign:
		c.emit(runtime.OpInplaceDivide)
	case model.TK_DoubleSlashAssign:
		c.emit(runtime.OpInplaceFloorDiv)
	case model.TK_PercentAssign:
		c.emit(runtime.OpInplaceModulo)
	case model.TK_DoubleStarAssign:
		c.emit(runtime.OpInplacePower)
	case model.TK_AtAssign:
		c.emit(runtime.OpInplaceMatMul)
	case model.TK_LShiftAssign:
		c.emit(runtime.OpInplaceLShift)
	case model.TK_RShiftAssign:
		c.emit(runtime.OpInplaceRShift)
	case model.TK_AmpersandAssign:
		c.emit(runtime.OpInplaceAnd)
	case model.TK_PipeAssign:
		c.emit(runtime.OpInplaceOr)
	case model.TK_CaretAssign:
		c.emit(runtime.OpInplaceXor)
	}

	// Store result
	if isSubscript {
		// For subscript: stack is [obj, idx, result], need [result, obj, idx] for StoreSubscr
		c.emit(runtime.OpRot3)       // [obj, idx, result] -> [result, obj, idx]
		c.emit(runtime.OpStoreSubscr) // Store and pop all three
	} else {
		c.compileStore(s.Target)
	}
}

// Control flow compilation

func (c *Compiler) compileIf(s *model.If) {
	c.compileExpr(s.Test)
	falseJump := c.emitJump(runtime.OpPopJumpIfFalse)

	// Compile body
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	if len(s.OrElse) > 0 {
		endJump := c.emitJump(runtime.OpJump)
		c.patchJump(falseJump, c.currentOffset())
		for _, stmt := range s.OrElse {
			c.compileStmt(stmt)
		}
		c.patchJump(endJump, c.currentOffset())
	} else {
		c.patchJump(falseJump, c.currentOffset())
	}
}

func (c *Compiler) compileWhile(s *model.While) {
	loopStart := c.currentOffset()

	c.loopStack = append(c.loopStack, loopInfo{startOffset: loopStart})

	c.compileExpr(s.Test)
	exitJump := c.emitJump(runtime.OpPopJumpIfFalse)

	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}
	c.emitArg(runtime.OpJump, loopStart)

	loopEnd := c.currentOffset()
	c.patchJump(exitJump, loopEnd)

	// Handle else clause
	if len(s.OrElse) > 0 {
		for _, stmt := range s.OrElse {
			c.compileStmt(stmt)
		}
	}

	// Patch break and continue jumps
	loop := c.loopStack[len(c.loopStack)-1]
	for _, jump := range loop.breakJumps {
		c.patchJump(jump, c.currentOffset())
	}
	for _, jump := range loop.continueJumps {
		c.patchJump(jump, loopStart)
	}
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

func (c *Compiler) compileFor(s *model.For) {
	// Compile iterator
	c.compileExpr(s.Iter)
	c.emit(runtime.OpGetIter)

	loopStart := c.currentOffset()
	c.loopStack = append(c.loopStack, loopInfo{startOffset: loopStart, isForLoop: true})

	// FOR_ITER jumps to end when exhausted
	exitJump := c.emitJump(runtime.OpForIter)

	// Store loop variable
	c.compileStore(s.Target)

	// Compile body
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	c.emitArg(runtime.OpJump, loopStart)

	loopEnd := c.currentOffset()
	c.patchJump(exitJump, loopEnd)

	// Handle else clause
	if len(s.OrElse) > 0 {
		for _, stmt := range s.OrElse {
			c.compileStmt(stmt)
		}
	}

	// Patch break and continue jumps
	loop := c.loopStack[len(c.loopStack)-1]
	for _, jump := range loop.breakJumps {
		c.patchJump(jump, c.currentOffset())
	}
	for _, jump := range loop.continueJumps {
		c.patchJump(jump, loopStart)
	}
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

func (c *Compiler) compileTry(s *model.Try) {
	hasFinally := len(s.FinalBody) > 0
	hasExcept := len(s.Handlers) > 0

	// If we have a finally block, set it up first (it wraps everything)
	var finallyJump int
	if hasFinally {
		finallyJump = c.emitJump(runtime.OpSetupFinally)
	}

	// Setup exception handler only if we have except clauses
	var handlerJump int
	if hasExcept {
		handlerJump = c.emitJump(runtime.OpSetupExcept)
	}

	// Compile try body
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	// Pop exception handler if we set one up
	if hasExcept {
		c.emit(runtime.OpPopExcept)
	}
	successJump := c.emitJump(runtime.OpJump)

	// Exception handlers start here (only if we have handlers)
	var handlerEnds []int
	if hasExcept {
		c.patchJump(handlerJump, c.currentOffset())

		for _, handler := range s.Handlers {
			if handler.Type != nil {
				// Check exception type using OpExceptionMatch
				c.emit(runtime.OpDup)
				c.compileExpr(handler.Type)
				c.emit(runtime.OpExceptionMatch)
				nextHandler := c.emitJump(runtime.OpPopJumpIfFalse)

				// Exception matched - clear the current exception state
				// Use OpClearException instead of OpPopExcept because the block was
				// already popped by handleException in the VM
				c.emit(runtime.OpClearException)

				// Pop the exception if we're not storing it
				if handler.Name != nil {
					c.compileStore(handler.Name)
				} else {
					c.emit(runtime.OpPop)
				}

				// Compile handler body
				for _, stmt := range handler.Body {
					c.compileStmt(stmt)
				}

				// Clear exception variable after handler (Python 3 semantics)
				if handler.Name != nil {
					c.emit(runtime.OpLoadNone)
					c.compileStore(handler.Name)
				}

				handlerEnds = append(handlerEnds, c.emitJump(runtime.OpJump))
				c.patchJump(nextHandler, c.currentOffset())
			} else {
				// Bare except - catches everything
				c.emit(runtime.OpClearException) // Clear the current exception state
				c.emit(runtime.OpPop)
				for _, stmt := range handler.Body {
					c.compileStmt(stmt)
				}
				handlerEnds = append(handlerEnds, c.emitJump(runtime.OpJump))
			}
		}

		// Re-raise if no handler matched (will be caught by finally if present)
		c.emitArg(runtime.OpRaiseVarargs, 0)
	}

	// Else clause (runs if no exception was raised)
	c.patchJump(successJump, c.currentOffset())
	for _, stmt := range s.OrElse {
		c.compileStmt(stmt)
	}

	// All handler ends jump here
	endOfHandlers := c.currentOffset()
	for _, jump := range handlerEnds {
		c.patchJump(jump, endOfHandlers)
	}

	// Finally clause
	if hasFinally {
		// Patch the finally setup to jump here
		c.patchJump(finallyJump, c.currentOffset())

		// Compile finally body
		for _, stmt := range s.FinalBody {
			c.compileStmt(stmt)
		}

		// End finally - will re-raise exception if one was active
		c.emit(runtime.OpEndFinally)
	}
}

func (c *Compiler) compileWith(s *model.With) {
	for i, item := range s.Items {
		c.compileExpr(item.ContextExpr)

		// Call __enter__
		c.emit(runtime.OpDup)
		enterIdx := c.addName("__enter__")
		c.emitArg(runtime.OpLoadMethod, enterIdx)
		c.emitArg(runtime.OpCallMethod, 0)

		if item.OptionalVar != nil {
			c.compileStore(item.OptionalVar)
		} else {
			c.emit(runtime.OpPop)
		}

		// Setup cleanup
		if i == len(s.Items)-1 {
			// Last item, compile body
			for _, stmt := range s.Body {
				c.compileStmt(stmt)
			}
		}
	}

	// Call __exit__ for each item (in reverse)
	for range s.Items {
		exitIdx := c.addName("__exit__")
		c.emitArg(runtime.OpLoadMethod, exitIdx)
		c.emitLoadConst(nil) // exc_type
		c.emitLoadConst(nil) // exc_val
		c.emitLoadConst(nil) // exc_tb
		c.emitArg(runtime.OpCallMethod, 3)
		c.emit(runtime.OpPop)
	}
}

// Function and class compilation

// containsYield checks if statements contain yield or yield from expressions
func containsYield(stmts []model.Stmt) bool {
	for _, stmt := range stmts {
		if containsYieldInStmt(stmt) {
			return true
		}
	}
	return false
}

func containsYieldInStmt(stmt model.Stmt) bool {
	switch s := stmt.(type) {
	case *model.ExprStmt:
		return containsYieldInExpr(s.Value)
	case *model.Assign:
		return containsYieldInExpr(s.Value)
	case *model.AugAssign:
		return containsYieldInExpr(s.Value)
	case *model.AnnAssign:
		if s.Value != nil {
			return containsYieldInExpr(s.Value)
		}
	case *model.Return:
		if s.Value != nil {
			return containsYieldInExpr(s.Value)
		}
	case *model.If:
		if containsYieldInExpr(s.Test) || containsYield(s.Body) || containsYield(s.OrElse) {
			return true
		}
	case *model.While:
		if containsYieldInExpr(s.Test) || containsYield(s.Body) || containsYield(s.OrElse) {
			return true
		}
	case *model.For:
		if containsYieldInExpr(s.Iter) || containsYield(s.Body) || containsYield(s.OrElse) {
			return true
		}
	case *model.With:
		for _, item := range s.Items {
			if containsYieldInExpr(item.ContextExpr) {
				return true
			}
		}
		return containsYield(s.Body)
	case *model.Try:
		if containsYield(s.Body) || containsYield(s.OrElse) || containsYield(s.FinalBody) {
			return true
		}
		for _, handler := range s.Handlers {
			if containsYield(handler.Body) {
				return true
			}
		}
	case *model.Match:
		if containsYieldInExpr(s.Subject) {
			return true
		}
		for _, c := range s.Cases {
			if containsYield(c.Body) {
				return true
			}
		}
		// Note: Don't descend into nested FunctionDef or ClassDef
	}
	return false
}

func containsYieldInExpr(expr model.Expr) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *model.Yield, *model.YieldFrom:
		return true
	case *model.BinaryOp:
		return containsYieldInExpr(e.Left) || containsYieldInExpr(e.Right)
	case *model.UnaryOp:
		return containsYieldInExpr(e.Operand)
	case *model.BoolOp:
		for _, v := range e.Values {
			if containsYieldInExpr(v) {
				return true
			}
		}
	case *model.Compare:
		if containsYieldInExpr(e.Left) {
			return true
		}
		for _, c := range e.Comparators {
			if containsYieldInExpr(c) {
				return true
			}
		}
	case *model.Call:
		if containsYieldInExpr(e.Func) {
			return true
		}
		for _, arg := range e.Args {
			if containsYieldInExpr(arg) {
				return true
			}
		}
		for _, kw := range e.Keywords {
			if containsYieldInExpr(kw.Value) {
				return true
			}
		}
	case *model.IfExpr:
		return containsYieldInExpr(e.Test) || containsYieldInExpr(e.Body) || containsYieldInExpr(e.OrElse)
	case *model.Attribute:
		return containsYieldInExpr(e.Value)
	case *model.Subscript:
		return containsYieldInExpr(e.Value) || containsYieldInExpr(e.Slice)
	case *model.Slice:
		return containsYieldInExpr(e.Lower) || containsYieldInExpr(e.Upper) || containsYieldInExpr(e.Step)
	case *model.List:
		for _, el := range e.Elts {
			if containsYieldInExpr(el) {
				return true
			}
		}
	case *model.Tuple:
		for _, el := range e.Elts {
			if containsYieldInExpr(el) {
				return true
			}
		}
	case *model.Dict:
		for _, k := range e.Keys {
			if containsYieldInExpr(k) {
				return true
			}
		}
		for _, v := range e.Values {
			if containsYieldInExpr(v) {
				return true
			}
		}
	case *model.Set:
		for _, el := range e.Elts {
			if containsYieldInExpr(el) {
				return true
			}
		}
	case *model.Starred:
		return containsYieldInExpr(e.Value)
	case *model.NamedExpr:
		return containsYieldInExpr(e.Value)
	case *model.Await:
		return containsYieldInExpr(e.Value)
		// ListComp, SetComp, DictComp, GeneratorExpr create their own scope, don't check
	}
	return false
}

func (c *Compiler) compileFunctionDef(s *model.FunctionDef) {
	// Compile decorators
	for _, dec := range s.Decorators {
		c.compileExpr(dec)
	}

	// Create a new compiler for the function body
	funcCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      s.Name.Name,
			Filename:  c.filename,
			FirstLine: s.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeFunction, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	// Check if we're inside a class and this method uses super() or __class__
	// If so, we need to capture __class__ as a free variable
	if classScope, isInClass := funcCompiler.symbolTable.IsInsideClass(); isInClass {
		if usesSuperOrClass(s.Body) {
			// Ensure __class__ is marked as a cell in the class scope
			classScope.MarkAsCell("__class__")
			// Resolve __class__ in the function's symbol table to create the free variable
			funcCompiler.symbolTable.Resolve("__class__")
		}
	}

	// Define parameters
	if s.Args != nil {
		for _, arg := range s.Args.Args {
			funcCompiler.symbolTable.Define(arg.Arg.Name)
		}
		for _, arg := range s.Args.PosOnlyArgs {
			funcCompiler.symbolTable.Define(arg.Arg.Name)
		}
		for _, arg := range s.Args.KwOnlyArgs {
			funcCompiler.symbolTable.Define(arg.Arg.Name)
		}
		if s.Args.VarArg != nil {
			funcCompiler.symbolTable.Define(s.Args.VarArg.Arg.Name)
		}
		if s.Args.KwArg != nil {
			funcCompiler.symbolTable.Define(s.Args.KwArg.Arg.Name)
		}
	}

	// Compile function body
	for _, stmt := range s.Body {
		funcCompiler.compileStmt(stmt)
	}

	// Add implicit return None
	funcCompiler.emitLoadConst(nil)
	funcCompiler.emit(runtime.OpReturn)
	funcCompiler.finishLineTable()
	funcCompiler.finalizeCode()

	// Apply peephole optimizations to function body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(funcCompiler.code)
	}

	// Set up code object
	funcCode := funcCompiler.code
	if s.Args != nil {
		funcCode.ArgCount = len(s.Args.Args) + len(s.Args.PosOnlyArgs)
		funcCode.KwOnlyArgCount = len(s.Args.KwOnlyArgs)
		if s.Args.VarArg != nil {
			funcCode.Flags |= runtime.FlagVarArgs
		}
		if s.Args.KwArg != nil {
			funcCode.Flags |= runtime.FlagVarKeywords
		}
	}

	// Check if function is a generator or coroutine
	isGenerator := containsYield(s.Body)
	if s.IsAsync {
		if isGenerator {
			funcCode.Flags |= runtime.FlagAsyncGenerator
		} else {
			funcCode.Flags |= runtime.FlagCoroutine
		}
	} else if isGenerator {
		funcCode.Flags |= runtime.FlagGenerator
	}

	// Load defaults
	if s.Args != nil && len(s.Args.Defaults) > 0 {
		for _, def := range s.Args.Defaults {
			c.compileExpr(def)
		}
		c.emitArg(runtime.OpBuildTuple, len(s.Args.Defaults))
	}

	// Load code object and make function
	c.emitLoadConst(funcCode)
	c.emitLoadConst(s.Name.Name)

	flags := 0
	if s.Args != nil && len(s.Args.Defaults) > 0 {
		flags |= 1 // Has defaults
	}
	c.emitArg(runtime.OpMakeFunction, flags)

	// Apply decorators (in reverse order)
	for i := len(s.Decorators) - 1; i >= 0; i-- {
		c.emitArg(runtime.OpCall, 1)
	}

	// Store function
	c.compileStore(s.Name)
}

func (c *Compiler) compileClassDef(s *model.ClassDef) {
	// Compile decorators
	for _, dec := range s.Decorators {
		c.compileExpr(dec)
	}

	c.emit(runtime.OpLoadBuildClass)

	// Check if any method in the class uses super() or __class__
	// If so, we need to set up __class__ as a cell variable
	needsClassCell := classNeedsClassCell(s.Body)

	// Create class body function
	classCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      s.Name.Name,
			Filename:  c.filename,
			FirstLine: s.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeClass, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	// If any method uses super(), define __class__ as a cell in the class scope
	if needsClassCell {
		classCompiler.symbolTable.Define("__class__")
		classCompiler.symbolTable.MarkAsCell("__class__")
	}

	// Compile class body
	for _, stmt := range s.Body {
		classCompiler.compileStmt(stmt)
	}
	classCompiler.emit(runtime.OpLoadLocals)
	classCompiler.emit(runtime.OpReturn)
	classCompiler.finishLineTable()
	classCompiler.finalizeCode()

	// Apply peephole optimizations to class body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(classCompiler.code)
	}

	c.emitLoadConst(classCompiler.code)
	c.emitLoadConst(s.Name.Name)
	c.emitArg(runtime.OpMakeFunction, 0)

	// Class name
	c.emitLoadConst(s.Name.Name)

	// Compile bases
	for _, base := range s.Bases {
		c.compileExpr(base)
	}

	// Compile keywords
	for _, kw := range s.Keywords {
		if kw.Arg != nil {
			c.emitLoadConst(kw.Arg.Name)
		}
		c.compileExpr(kw.Value)
	}

	argc := 2 + len(s.Bases) + len(s.Keywords)*2
	if len(s.Keywords) > 0 {
		c.emitArg(runtime.OpCallKw, argc)
	} else {
		c.emitArg(runtime.OpCall, argc)
	}

	// Apply decorators
	for i := len(s.Decorators) - 1; i >= 0; i-- {
		c.emitArg(runtime.OpCall, 1)
	}

	// Store class
	c.compileStore(s.Name)
}

func (c *Compiler) compileLambda(e *model.Lambda) {
	// Create lambda function code
	lambdaCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      "<lambda>",
			Filename:  c.filename,
			FirstLine: e.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeFunction, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	// Define parameters
	if e.Args != nil {
		for _, arg := range e.Args.Args {
			lambdaCompiler.symbolTable.Define(arg.Arg.Name)
		}
	}

	// Compile body expression and return it
	lambdaCompiler.setLine(e.Body.Pos().Line)
	lambdaCompiler.compileExpr(e.Body)
	lambdaCompiler.emit(runtime.OpReturn)
	lambdaCompiler.finishLineTable()
	lambdaCompiler.finalizeCode()

	// Apply peephole optimizations to lambda body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(lambdaCompiler.code)
	}

	lambdaCode := lambdaCompiler.code
	if e.Args != nil {
		lambdaCode.ArgCount = len(e.Args.Args)
	}

	// Load defaults
	if e.Args != nil && len(e.Args.Defaults) > 0 {
		for _, def := range e.Args.Defaults {
			c.compileExpr(def)
		}
		c.emitArg(runtime.OpBuildTuple, len(e.Args.Defaults))
	}

	c.emitLoadConst(lambdaCode)
	c.emitLoadConst("<lambda>")

	flags := 0
	if e.Args != nil && len(e.Args.Defaults) > 0 {
		flags |= 1
	}
	c.emitArg(runtime.OpMakeFunction, flags)
}

// Comprehension compilation

func (c *Compiler) compileListComp(e *model.ListComp) {
	// Create comprehension function
	compCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      "<listcomp>",
			Filename:  c.filename,
			FirstLine: e.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeComprehension, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	// The outermost iterable is passed as argument
	compCompiler.symbolTable.Define(".0")

	// Build empty list
	compCompiler.emitArg(runtime.OpBuildList, 0)

	// The stack offset for LIST_APPEND is the number of generators
	// because after the loops, stack is [list, iter1, iter2, ..., iterN, element]
	// After popping element: [list, iter1, iter2, ..., iterN]
	// peek(N) finds the list at the correct position
	stackOffset := len(e.Generators)
	c.compileComprehensionGenerators(compCompiler, e.Generators, func() {
		compCompiler.compileExpr(e.Elt)
		compCompiler.emitArg(runtime.OpListAppend, stackOffset)
	}, 0)

	compCompiler.emit(runtime.OpReturn)
	compCompiler.finishLineTable()
	compCompiler.finalizeCode()

	// Apply peephole optimizations to comprehension body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(compCompiler.code)
	}

	// Make function and call with iterator
	c.compileExpr(e.Generators[0].Iter)
	c.emit(runtime.OpGetIter)
	c.emitLoadConst(compCompiler.code)
	c.emitLoadConst("<listcomp>")
	c.emitArg(runtime.OpMakeFunction, 0)
	c.emit(runtime.OpRot2)
	c.emitArg(runtime.OpCall, 1)
}

func (c *Compiler) compileSetComp(e *model.SetComp) {
	compCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      "<setcomp>",
			Filename:  c.filename,
			FirstLine: e.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeComprehension, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	compCompiler.symbolTable.Define(".0")
	compCompiler.emitArg(runtime.OpBuildSet, 0)

	stackOffset := len(e.Generators)
	c.compileComprehensionGenerators(compCompiler, e.Generators, func() {
		compCompiler.compileExpr(e.Elt)
		compCompiler.emitArg(runtime.OpSetAdd, stackOffset)
	}, 0)

	compCompiler.emit(runtime.OpReturn)
	compCompiler.finishLineTable()
	compCompiler.finalizeCode()

	// Apply peephole optimizations to comprehension body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(compCompiler.code)
	}

	c.compileExpr(e.Generators[0].Iter)
	c.emit(runtime.OpGetIter)
	c.emitLoadConst(compCompiler.code)
	c.emitLoadConst("<setcomp>")
	c.emitArg(runtime.OpMakeFunction, 0)
	c.emit(runtime.OpRot2)
	c.emitArg(runtime.OpCall, 1)
}

func (c *Compiler) compileDictComp(e *model.DictComp) {
	compCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      "<dictcomp>",
			Filename:  c.filename,
			FirstLine: e.StartPos.Line,
		},
		symbolTable: NewSymbolTable(ScopeComprehension, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	compCompiler.symbolTable.Define(".0")
	compCompiler.emitArg(runtime.OpBuildMap, 0)

	stackOffset := len(e.Generators)
	c.compileComprehensionGenerators(compCompiler, e.Generators, func() {
		compCompiler.compileExpr(e.Key)
		compCompiler.compileExpr(e.Value)
		compCompiler.emitArg(runtime.OpMapAdd, stackOffset)
	}, 0)

	compCompiler.emit(runtime.OpReturn)
	compCompiler.finishLineTable()
	compCompiler.finalizeCode()

	// Apply peephole optimizations to comprehension body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(compCompiler.code)
	}

	c.compileExpr(e.Generators[0].Iter)
	c.emit(runtime.OpGetIter)
	c.emitLoadConst(compCompiler.code)
	c.emitLoadConst("<dictcomp>")
	c.emitArg(runtime.OpMakeFunction, 0)
	c.emit(runtime.OpRot2)
	c.emitArg(runtime.OpCall, 1)
}

func (c *Compiler) compileGeneratorExpr(e *model.GeneratorExpr) {
	compCompiler := &Compiler{
		code: &runtime.CodeObject{
			Name:      "<genexpr>",
			Filename:  c.filename,
			FirstLine: e.StartPos.Line,
			Flags:     runtime.FlagGenerator,
		},
		symbolTable: NewSymbolTable(ScopeComprehension, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}

	compCompiler.symbolTable.Define(".0")

	c.compileComprehensionGenerators(compCompiler, e.Generators, func() {
		compCompiler.compileExpr(e.Elt)
		// Yield the value (simplified)
	}, 0)

	compCompiler.emitLoadConst(nil)
	compCompiler.emit(runtime.OpReturn)
	compCompiler.finishLineTable()
	compCompiler.finalizeCode()

	// Apply peephole optimizations to comprehension body
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(compCompiler.code)
	}

	c.compileExpr(e.Generators[0].Iter)
	c.emit(runtime.OpGetIter)
	c.emitLoadConst(compCompiler.code)
	c.emitLoadConst("<genexpr>")
	c.emitArg(runtime.OpMakeFunction, 0)
	c.emit(runtime.OpRot2)
	c.emitArg(runtime.OpCall, 1)
}

func (c *Compiler) compileComprehensionGenerators(
	comp *Compiler,
	generators []*model.Comprehension,
	body func(),
	depth int,
) {
	if depth >= len(generators) {
		body()
		return
	}

	gen := generators[depth]

	// First generator uses .0 argument, others compile their iter
	if depth == 0 {
		comp.emitArg(runtime.OpLoadFast, 0) // .0
	} else {
		comp.compileExpr(gen.Iter)
		comp.emit(runtime.OpGetIter)
	}

	loopStart := comp.currentOffset()
	exitJump := comp.emitJump(runtime.OpForIter)

	// Define and store target
	switch t := gen.Target.(type) {
	case *model.Identifier:
		comp.symbolTable.Define(t.Name)
	}
	comp.compileStore(gen.Target)

	// Compile if conditions
	var ifJumps []int
	for _, cond := range gen.Ifs {
		comp.compileExpr(cond)
		ifJumps = append(ifJumps, comp.emitJump(runtime.OpPopJumpIfFalse))
	}

	// Recurse to next generator or body
	c.compileComprehensionGenerators(comp, generators, body, depth+1)

	// Patch if jumps
	for _, jump := range ifJumps {
		comp.patchJump(jump, comp.currentOffset())
	}

	comp.emitArg(runtime.OpJump, loopStart)
	comp.patchJump(exitJump, comp.currentOffset())
}

func (c *Compiler) compileMatch(s *model.Match) {
	c.compileExpr(s.Subject)

	var caseEnds []int

	for _, matchCase := range s.Cases {
		// Duplicate subject for pattern matching
		c.emit(runtime.OpDup)

		// Compile pattern (simplified)
		c.compilePattern(matchCase.Pattern)

		nextCase := c.emitJump(runtime.OpPopJumpIfFalse)

		// Pop subject if matched
		c.emit(runtime.OpPop)

		// Guard
		if matchCase.Guard != nil {
			c.compileExpr(matchCase.Guard)
			guardFail := c.emitJump(runtime.OpPopJumpIfFalse)

			// Compile body
			for _, stmt := range matchCase.Body {
				c.compileStmt(stmt)
			}
			caseEnds = append(caseEnds, c.emitJump(runtime.OpJump))

			c.patchJump(guardFail, c.currentOffset())
		} else {
			// Compile body
			for _, stmt := range matchCase.Body {
				c.compileStmt(stmt)
			}
			caseEnds = append(caseEnds, c.emitJump(runtime.OpJump))
		}

		c.patchJump(nextCase, c.currentOffset())
	}

	// Pop subject if no case matched
	c.emit(runtime.OpPop)

	// Patch all case end jumps
	for _, jump := range caseEnds {
		c.patchJump(jump, c.currentOffset())
	}
}

func (c *Compiler) compilePattern(pattern model.Pattern) {
	// All patterns should:
	// - Expect the element to match on top of stack
	// - Leave the element on stack
	// - Push True/False result on top
	// After: [..., element, True/False]

	switch p := pattern.(type) {
	case *model.MatchValue:
		// Dup element, load value, compare
		c.emit(runtime.OpDup)
		c.compileExpr(p.Value)
		c.emit(runtime.OpCompareEq)

	case *model.MatchSingleton:
		// Dup element, load singleton, compare with 'is'
		c.emit(runtime.OpDup)
		c.compileExpr(p.Value)
		c.emit(runtime.OpCompareIs)

	case *model.MatchAs:
		if p.Pattern != nil {
			// Pattern with 'as' binding: pattern as name
			// First match the element against the sub-pattern
			c.compilePattern(p.Pattern) // Stack: [..., element, True/False]
			matchJump := c.emitJump(runtime.OpPopJumpIfFalse) // Stack: [..., element]
			// Pattern matched, bind name
			if p.Name != nil {
				c.emit(runtime.OpDup)
				c.compileStore(p.Name)
			}
			c.emitLoadConst(true)
			endJump := c.emitJump(runtime.OpJump)
			c.patchJump(matchJump, c.currentOffset())
			// Pattern didn't match
			c.emitLoadConst(false)
			c.patchJump(endJump, c.currentOffset())
		} else {
			// Wildcard (_) or simple capture (name)
			// Just bind to name (if present) and always match
			if p.Name != nil {
				c.emit(runtime.OpDup)
				c.compileStore(p.Name)
			}
			c.emitLoadConst(true)
		}

	case *model.MatchSequence:
		c.compileSequencePattern(p)

	case *model.MatchMapping:
		c.compileMappingPattern(p)

	case *model.MatchClass:
		c.compileClassPattern(p)

	case *model.MatchOr:
		c.compileOrPattern(p)

	case *model.MatchStar:
		// Star patterns are handled within sequence pattern compilation
		// If we get here, it's an error (star outside sequence)
		c.emitLoadConst(true)

	default:
		c.emitLoadConst(true)
	}
}

func (c *Compiler) compileSequencePattern(p *model.MatchSequence) {
	// Stack: [..., subject]
	// After: [..., subject, True/False]

	// Find if there's a star pattern and its position
	starIndex := -1
	for i, pat := range p.Patterns {
		if _, isStar := pat.(*model.MatchStar); isStar {
			starIndex = i
			break
		}
	}

	if starIndex == -1 {
		// No star pattern - exact length match
		// Check if it's a sequence with correct length
		c.emitArg(runtime.OpMatchSequence, len(p.Patterns))
		failJump := c.emitJump(runtime.OpPopJumpIfFalse)

		// Match each sub-pattern by subscripting
		var subPatternFails []int
		for i, subPattern := range p.Patterns {
			// subject[i]
			c.emit(runtime.OpDup) // Dup subject
			c.emitLoadConst(int64(i))
			c.emit(runtime.OpBinarySubscr) // Get subject[i]

			c.compilePattern(subPattern) // Match pattern, leaves element and True/False
			subPatternFails = append(subPatternFails, c.emitJump(runtime.OpPopJumpIfFalse))
			c.emit(runtime.OpPop) // Pop the element
		}

		c.emitLoadConst(true)
		successJump := c.emitJump(runtime.OpJump)

		// Failure path - each sub-pattern failure leaves 1 element on stack
		// All sub-pattern failures share a single cleanup path
		if len(subPatternFails) > 0 {
			cleanupStart := c.currentOffset()
			for _, jump := range subPatternFails {
				c.patchJump(jump, cleanupStart)
			}
			c.emit(runtime.OpPop) // Pop the leftover element
			// Fall through to common failure epilogue
		}

		// Common failure epilogue
		c.patchJump(failJump, c.currentOffset())
		c.emitLoadConst(false)

		c.patchJump(successJump, c.currentOffset())
	} else {
		// Has star pattern - variable length match
		beforeStar := starIndex
		afterStar := len(p.Patterns) - starIndex - 1
		minLen := beforeStar + afterStar

		// Check if it's a sequence with minimum length
		c.emitArg(runtime.OpMatchSequence, -1) // -1 means any length (stored as 65535)
		failJump := c.emitJump(runtime.OpPopJumpIfFalse)

		c.emitArg(runtime.OpMatchStar, minLen) // Check minimum length
		starFailJump := c.emitJump(runtime.OpPopJumpIfFalse)

		var subPatternFails []int

		// Match patterns before star
		for i := 0; i < beforeStar; i++ {
			c.emit(runtime.OpDup)
			c.emitLoadConst(int64(i))
			c.emit(runtime.OpBinarySubscr)
			c.compilePattern(p.Patterns[i])
			subPatternFails = append(subPatternFails, c.emitJump(runtime.OpPopJumpIfFalse))
			c.emit(runtime.OpPop)
		}

		// Handle star pattern - bind slice to name
		starPat := p.Patterns[starIndex].(*model.MatchStar)
		if starPat.Name != nil {
			// Extract and bind the star slice: subject[beforeStar : len(subject) - afterStar]
			c.emit(runtime.OpDup) // Dup subject
			c.emitArg(runtime.OpExtractStar, (beforeStar<<8)|afterStar)
			c.compileStore(starPat.Name)
		}

		// Match patterns after star (from the end)
		for i := 0; i < afterStar; i++ {
			c.emit(runtime.OpDup)
			// Index from end: -(afterStar - i)
			c.emitLoadConst(int64(-(afterStar - i)))
			c.emit(runtime.OpBinarySubscr)
			c.compilePattern(p.Patterns[starIndex+1+i])
			subPatternFails = append(subPatternFails, c.emitJump(runtime.OpPopJumpIfFalse))
			c.emit(runtime.OpPop)
		}

		c.emitLoadConst(true)
		successJump := c.emitJump(runtime.OpJump)

		// Failure path - each sub-pattern failure leaves 1 element on stack
		// All sub-pattern failures share a single cleanup path
		if len(subPatternFails) > 0 {
			cleanupStart := c.currentOffset()
			for _, jump := range subPatternFails {
				c.patchJump(jump, cleanupStart)
			}
			c.emit(runtime.OpPop) // Pop the leftover element
			// Fall through to common failure epilogue
		}

		// Common failure epilogue
		c.patchJump(failJump, c.currentOffset())
		c.patchJump(starFailJump, c.currentOffset())
		c.emitLoadConst(false)

		c.patchJump(successJump, c.currentOffset())
	}
}

func (c *Compiler) compileMappingPattern(p *model.MatchMapping) {
	// Check if it's a mapping
	c.emitArg(runtime.OpMatchMapping, len(p.Keys))
	failJump := c.emitJump(runtime.OpPopJumpIfFalse)

	// Track failure jumps with how many values need to be popped
	type failInfo struct {
		jump       int
		valuesToPop int
	}
	var subPatternFails []failInfo
	var keysFailJump int

	numValues := len(p.Patterns)
	if len(p.Keys) > 0 {
		// Push keys onto stack and check they exist
		for _, key := range p.Keys {
			c.compileExpr(key)
		}
		c.emitLoadConst(int64(len(p.Keys)))
		c.emit(runtime.OpMatchKeys)
		keysFailJump = c.emitJump(runtime.OpPopJumpIfFalse)

		// Match value patterns against extracted values
		// After OpMatchKeys success, stack has values in reverse order
		for i, pat := range p.Patterns {
			c.compilePattern(pat)
			// If this pattern fails, we need to pop remaining values (numValues - i)
			subPatternFails = append(subPatternFails, failInfo{
				jump:        c.emitJump(runtime.OpPopJumpIfFalse),
				valuesToPop: numValues - i,
			})
			c.emit(runtime.OpPop)
		}
	}

	// Handle **rest if present
	if p.Rest != nil {
		// Push keys to remove
		for _, key := range p.Keys {
			c.compileExpr(key)
		}
		c.emitLoadConst(int64(len(p.Keys)))
		c.emit(runtime.OpCopyDict)
		c.compileStore(p.Rest)
	}

	c.emitLoadConst(true)
	successJump := c.emitJump(runtime.OpJump)

	// Emit failure cleanup paths for each sub-pattern failure
	var cleanupJumps []int
	for i := 0; i < len(subPatternFails); i++ {
		info := subPatternFails[i]
		c.patchJump(info.jump, c.currentOffset())
		// Pop remaining values
		for j := 0; j < info.valuesToPop; j++ {
			c.emit(runtime.OpPop)
		}
		cleanupJumps = append(cleanupJumps, c.emitJump(runtime.OpJump))
	}

	// Common failure epilogue
	c.patchJump(failJump, c.currentOffset())
	if len(p.Keys) > 0 {
		c.patchJump(keysFailJump, c.currentOffset())
	}
	for _, jump := range cleanupJumps {
		c.patchJump(jump, c.currentOffset())
	}
	c.emitLoadConst(false)

	c.patchJump(successJump, c.currentOffset())
}

func (c *Compiler) compileClassPattern(p *model.MatchClass) {
	// Push class onto stack
	c.compileExpr(p.Cls)

	// Match class pattern with positional arguments
	c.emitArg(runtime.OpMatchClass, len(p.Patterns))
	failJump := c.emitJump(runtime.OpPopJumpIfFalse)

	// Track failure jumps with how many attrs need to be popped
	type failInfo struct {
		jump       int
		attrsToPop int
	}
	var subPatternFails []failInfo

	// Match positional patterns against extracted attributes
	// After OpMatchClass success, stack has: [..., subject, attr_{n-1}, ..., attr_0, True]
	// After PJIF, stack has: [..., subject, attr_{n-1}, ..., attr_0]
	numPositional := len(p.Patterns)
	for i, pat := range p.Patterns {
		c.compilePattern(pat)
		// If this pattern fails, we need to pop remaining attrs (numPositional - i)
		subPatternFails = append(subPatternFails, failInfo{
			jump:       c.emitJump(runtime.OpPopJumpIfFalse),
			attrsToPop: numPositional - i,
		})
		c.emit(runtime.OpPop)
	}

	// Handle keyword patterns - each fetches attribute and matches
	var kwdFails []int
	for i, attr := range p.KwdAttrs {
		// Get attribute from subject
		c.emit(runtime.OpDup) // Dup subject
		c.emitArg(runtime.OpLoadAttr, c.addName(attr.Name))

		// Match against keyword pattern
		c.compilePattern(p.KwdPatterns[i])
		kwdFails = append(kwdFails, c.emitJump(runtime.OpPopJumpIfFalse))
		c.emit(runtime.OpPop)
	}

	c.emitLoadConst(true)
	successJump := c.emitJump(runtime.OpJump)

	// Emit failure cleanup paths for each positional pattern failure
	// Each path pops the appropriate number of attrs then jumps to common epilogue
	var cleanupJumps []int
	for i := 0; i < len(subPatternFails); i++ {
		info := subPatternFails[i]
		c.patchJump(info.jump, c.currentOffset())
		// Pop remaining attrs
		for j := 0; j < info.attrsToPop; j++ {
			c.emit(runtime.OpPop)
		}
		// Jump to common failure epilogue
		cleanupJumps = append(cleanupJumps, c.emitJump(runtime.OpJump))
	}

	// Emit failure cleanup for keyword pattern failures
	// Each keyword failure leaves 1 attr value on stack
	for _, jump := range kwdFails {
		c.patchJump(jump, c.currentOffset())
		c.emit(runtime.OpPop) // Pop the leftover attr value
		cleanupJumps = append(cleanupJumps, c.emitJump(runtime.OpJump))
	}

	// Common failure epilogue
	c.patchJump(failJump, c.currentOffset())
	for _, jump := range cleanupJumps {
		c.patchJump(jump, c.currentOffset())
	}
	c.emitLoadConst(false)

	c.patchJump(successJump, c.currentOffset())
}

func (c *Compiler) compileOrPattern(p *model.MatchOr) {
	// Try each pattern, short-circuit on first match
	var orJumps []int
	for i, subPattern := range p.Patterns {
		if i < len(p.Patterns)-1 {
			c.emit(runtime.OpDup)
		}
		c.compilePattern(subPattern)
		if i < len(p.Patterns)-1 {
			orJumps = append(orJumps, c.emitJump(runtime.OpJumpIfTrueOrPop))
		}
	}

	// Patch all successful jumps to the end
	for _, jump := range orJumps {
		c.patchJump(jump, c.currentOffset())
	}
}

func (c *Compiler) finalizeCode() {
	// Build VarNames list for local variables
	for name, sym := range c.symbolTable.symbols {
		if sym.Scope == ScopeLocal && sym.Index >= 0 {
			// Ensure VarNames has enough capacity
			for len(c.code.VarNames) <= sym.Index {
				c.code.VarNames = append(c.code.VarNames, "")
			}
			c.code.VarNames[sym.Index] = name
		}
	}

	// Build CellVars list (variables captured by inner functions)
	// Also add cell variables to VarNames at their original positions
	// so the VM can match parameter names to cells
	for _, sym := range c.symbolTable.cellSyms {
		c.code.CellVars = append(c.code.CellVars, sym.Name)
		// Add to VarNames at the original position (before it became a cell)
		if sym.OriginalIndex >= 0 {
			for len(c.code.VarNames) <= sym.OriginalIndex {
				c.code.VarNames = append(c.code.VarNames, "")
			}
			c.code.VarNames[sym.OriginalIndex] = sym.Name
		}
	}

	// Build FreeVars list
	for _, sym := range c.symbolTable.freeSyms {
		c.code.FreeVars = append(c.code.FreeVars, sym.Name)
	}

	// Calculate stack size (simplified estimate)
	c.code.StackSize = c.estimateStackSize()
}

func (c *Compiler) estimateStackSize() int {
	// Conservative estimate based on code length
	maxStack := 10
	for i := 0; i < len(c.code.Code); {
		op := runtime.Opcode(c.code.Code[i])
		if op.HasArg() {
			i += 3
		} else {
			i++
		}
		// Certain ops increase stack needs
		switch op {
		case runtime.OpBuildList, runtime.OpBuildTuple, runtime.OpBuildSet, runtime.OpBuildMap:
			if i > 2 {
				arg := int(c.code.Code[i-2]) | int(c.code.Code[i-1])<<8
				if arg > maxStack {
					maxStack = arg + 10
				}
			}
		}
	}
	return maxStack
}

// CompileSource compiles Python source code to a code object
func CompileSource(source, filename string) (*runtime.CodeObject, []error) {
	parser := NewParser(source)
	module, parseErrors := parser.Parse()

	if len(parseErrors) > 0 {
		var errs []error
		for _, e := range parseErrors {
			errs = append(errs, e)
		}
		return nil, errs
	}

	compiler := NewCompiler(filename)
	code, compileErrors := compiler.Compile(module)

	if len(compileErrors) > 0 {
		var errs []error
		for _, e := range compileErrors {
			errs = append(errs, e)
		}
		return nil, errs
	}

	return code, nil
}

// indexByte returns the index of the first occurrence of c in s, or -1 if not present.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
