package compiler

import (
	"fmt"
	"strconv"

	"github.com/ATSOTECK/oink/internal/model"
	"github.com/ATSOTECK/oink/internal/runtime"
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
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable tracks variables in a scope
type SymbolTable struct {
	outer     *SymbolTable
	symbols   map[string]*Symbol
	freeSyms  []*Symbol
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
	sym := &Symbol{Name: name, Scope: ScopeLocal, Index: st.numDefs}
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

		// Create a free variable
		free := &Symbol{Name: name, Scope: ScopeFree, Index: len(st.freeSyms)}
		st.freeSyms = append(st.freeSyms, free)
		st.symbols[name] = free
		return free, true
	}

	return &Symbol{Name: name, Scope: ScopeGlobal, Index: -1}, true
}

// Compiler compiles AST to bytecode
type Compiler struct {
	code        *runtime.CodeObject
	symbolTable *SymbolTable
	errors      []CompileError
	loopStack   []loopInfo
	filename    string
	optimizer   *Optimizer
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

	// Build names and varnames lists
	c.finalizeCode()

	// Apply peephole optimizations
	c.optimizer.PeepholeOptimize(c.code)

	return c.code, c.errors
}

func (c *Compiler) error(pos model.Position, format string, args ...interface{}) {
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

func (c *Compiler) addConstant(value interface{}) int {
	for i, v := range c.code.Constants {
		if v == value {
			return i
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

func (c *Compiler) emitLoadConst(value interface{}) {
	idx := c.addConstant(value)
	c.emitArg(runtime.OpLoadConst, idx)
}

// Statement compilation

func (c *Compiler) compileStmt(stmt model.Stmt) {
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

			storeName := alias.Name.Name
			if alias.AsName != nil {
				storeName = alias.AsName.Name
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
		// Build slice object
		nameIdx := c.addName("slice")
		c.emitArg(runtime.OpLoadGlobal, nameIdx)
		c.emit(runtime.OpRot2) // Swap slice function with first arg
		// Need to rotate all 4 items to get: slice, lower, upper, step
		// This is simplified; actual implementation needs BUILD_SLICE opcode
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
		if e.Value != nil {
			c.compileExpr(e.Value)
		} else {
			c.emitLoadConst(nil)
		}
		// Yield is handled specially in generator functions
		// For now, this is a placeholder

	case *model.YieldFrom:
		c.compileExpr(e.Value)
		// Yield from is handled specially

	case *model.Await:
		c.compileExpr(e.Value)
		// Await is handled specially in async functions

	case *model.Starred:
		c.compileExpr(e.Value)
		// Starred unpacking handled by context

	case *model.NamedExpr:
		c.compileExpr(e.Value)
		c.emit(runtime.OpDup)
		c.compileStore(e.Target)

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
		for i, val := range e.Values[:len(e.Values)-1] {
			c.compileExpr(val)
			if i < len(e.Values)-1 {
				jumpOffsets = append(jumpOffsets, c.emitJump(runtime.OpJumpIfFalseOrPop))
			}
		}
		c.compileExpr(e.Values[len(e.Values)-1])
		endOffset := c.currentOffset()
		for _, offset := range jumpOffsets {
			c.patchJump(offset, endOffset)
		}
	} else {
		// Short-circuit or: if any is true, skip rest
		var jumpOffsets []int
		for i, val := range e.Values[:len(e.Values)-1] {
			c.compileExpr(val)
			if i < len(e.Values)-1 {
				jumpOffsets = append(jumpOffsets, c.emitJump(runtime.OpJumpIfTrueOrPop))
			}
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
		// Compile keyword arguments
		for _, kw := range e.Keywords {
			if kw.Arg != nil {
				c.emitLoadConst(kw.Arg.Name)
			}
			c.compileExpr(kw.Value)
		}
		c.emitArg(runtime.OpCallKw, len(e.Args)+len(e.Keywords)*2)
	} else {
		c.emitArg(runtime.OpCall, len(e.Args))
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
		c.compileExpr(t.Value)
		c.compileExpr(t.Slice)
		c.emit(runtime.OpDup)
		c.emit(runtime.OpDup)
		c.emit(runtime.OpBinarySubscr)
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
	c.compileStore(s.Target)
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
	// Setup exception handler
	handlerJump := c.emitJump(runtime.OpSetupExcept)

	// Compile try body
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}
	c.emit(runtime.OpPopExcept)
	successJump := c.emitJump(runtime.OpJump)

	// Exception handlers
	c.patchJump(handlerJump, c.currentOffset())

	var handlerEnds []int
	for _, handler := range s.Handlers {
		if handler.Type != nil {
			// Check exception type
			c.emit(runtime.OpDup)
			c.compileExpr(handler.Type)
			// isinstance check would go here
			c.emit(runtime.OpCompareIs) // Simplified
			nextHandler := c.emitJump(runtime.OpPopJumpIfFalse)

			if handler.Name != nil {
				c.compileStore(handler.Name)
			} else {
				c.emit(runtime.OpPop)
			}

			for _, stmt := range handler.Body {
				c.compileStmt(stmt)
			}
			handlerEnds = append(handlerEnds, c.emitJump(runtime.OpJump))

			c.patchJump(nextHandler, c.currentOffset())
		} else {
			// Bare except
			c.emit(runtime.OpPop)
			for _, stmt := range handler.Body {
				c.compileStmt(stmt)
			}
			handlerEnds = append(handlerEnds, c.emitJump(runtime.OpJump))
		}
	}

	// Re-raise if no handler matched
	c.emitArg(runtime.OpRaiseVarargs, 0)

	// Else clause (runs if no exception)
	c.patchJump(successJump, c.currentOffset())
	for _, stmt := range s.OrElse {
		c.compileStmt(stmt)
	}

	// Patch all handler ends
	for _, jump := range handlerEnds {
		c.patchJump(jump, c.currentOffset())
	}

	// Finally clause
	if len(s.FinalBody) > 0 {
		for _, stmt := range s.FinalBody {
			c.compileStmt(stmt)
		}
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

	// Compile class body
	for _, stmt := range s.Body {
		classCompiler.compileStmt(stmt)
	}
	classCompiler.emit(runtime.OpLoadLocals)
	classCompiler.emit(runtime.OpReturn)
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
	lambdaCompiler.compileExpr(e.Body)
	lambdaCompiler.emit(runtime.OpReturn)
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
	switch p := pattern.(type) {
	case *model.MatchValue:
		c.compileExpr(p.Value)
		c.emit(runtime.OpCompareEq)

	case *model.MatchSingleton:
		c.compileExpr(p.Value)
		c.emit(runtime.OpCompareIs)

	case *model.MatchAs:
		if p.Pattern != nil {
			c.emit(runtime.OpDup)
			c.compilePattern(p.Pattern)
			// If pattern matches, bind name
			matchJump := c.emitJump(runtime.OpPopJumpIfFalse)
			c.emit(runtime.OpDup)
			if p.Name != nil {
				c.compileStore(p.Name)
			}
			c.emitLoadConst(true)
			endJump := c.emitJump(runtime.OpJump)
			c.patchJump(matchJump, c.currentOffset())
			c.emitLoadConst(false)
			c.patchJump(endJump, c.currentOffset())
		} else {
			// Wildcard - always matches
			if p.Name != nil {
				c.emit(runtime.OpDup)
				c.compileStore(p.Name)
			}
			c.emitLoadConst(true)
		}

	case *model.MatchSequence:
		// Simplified sequence matching
		c.emitLoadConst(len(p.Patterns))
		// Length check would go here
		c.emit(runtime.OpCompareEq)

	case *model.MatchMapping:
		// Simplified mapping matching
		c.emitLoadConst(true)

	case *model.MatchOr:
		var orJumps []int
		for _, subPattern := range p.Patterns[:len(p.Patterns)-1] {
			c.emit(runtime.OpDup)
			c.compilePattern(subPattern)
			orJumps = append(orJumps, c.emitJump(runtime.OpJumpIfTrueOrPop))
		}
		c.compilePattern(p.Patterns[len(p.Patterns)-1])
		for _, jump := range orJumps {
			c.patchJump(jump, c.currentOffset())
		}

	default:
		c.emitLoadConst(true)
	}
}

func (c *Compiler) finalizeCode() {
	// Build VarNames list
	for name, sym := range c.symbolTable.symbols {
		if sym.Scope == ScopeLocal && sym.Index >= 0 {
			// Ensure VarNames has enough capacity
			for len(c.code.VarNames) <= sym.Index {
				c.code.VarNames = append(c.code.VarNames, "")
			}
			c.code.VarNames[sym.Index] = name
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
