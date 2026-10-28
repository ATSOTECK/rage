package compiler

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// Expression compilation

func (c *Compiler) compileExpr(expr model.Expr) {
	if expr == nil {
		c.errors = append(c.errors, CompileError{Message: "unexpected nil expression"})
		c.emit(runtime.OpLoadNone)
		return
	}

	switch e := expr.(type) {
	case *model.IntLit:
		val, err := strconv.ParseInt(e.Value, 0, 64)
		if err != nil {
			// Overflow: use big.Int
			bi := new(big.Int)
			s := e.Value
			// Handle base prefixes for big.Int
			base := 10
			if len(s) > 2 {
				switch s[:2] {
				case "0x", "0X":
					base = 16
					s = s[2:]
				case "0o", "0O":
					base = 8
					s = s[2:]
				case "0b", "0B":
					base = 2
					s = s[2:]
				}
			}
			bi.SetString(s, base)
			c.emitLoadConst(runtime.MakeBigInt(bi))
		} else {
			c.emitLoadConst(val)
		}

	case *model.FloatLit:
		val, _ := strconv.ParseFloat(e.Value, 64)
		c.emitLoadConst(val)

	case *model.ImaginaryLit:
		s := strings.TrimRight(e.Value, "jJ")
		imag, _ := strconv.ParseFloat(s, 64)
		c.emitLoadConst(runtime.MakeComplex(0, imag))

	case *model.StringLit:
		c.emitLoadConst(e.Value)

	case *model.FStringLit:
		c.compileFString(e)

	case *model.BytesLit:
		// Convert rune-by-rune to preserve raw byte values (e.g., \xff -> 0xFF, not UTF-8 encoded)
		runes := []rune(e.Value)
		raw := make([]byte, len(runes))
		for i, r := range runes {
			raw[i] = byte(r)
		}
		c.emitLoadConst(raw)

	case *model.BoolLit:
		c.emitLoadConst(e.Value)

	case *model.NoneLit:
		c.emitLoadConst(nil)

	case *model.Ellipsis:
		c.emitLoadConst(runtime.Ellipsis)

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
		c.compileStarredList(e.Elts)

	case *model.Tuple:
		if hasStarred(e.Elts) {
			c.compileStarredList(e.Elts)
			c.emit(runtime.OpListToTuple)
			break
		}
		for _, elt := range e.Elts {
			c.compileExpr(elt)
		}
		c.emitArg(runtime.OpBuildTuple, len(e.Elts))

	case *model.Dict:
		// Leading key/value pairs go into BUILD_MAP; once a **mapping appears,
		// the rest are added in order with DICT_UPDATE / MAP_ADD.
		n := 0
		for n < len(e.Keys) && e.Keys[n] != nil {
			c.compileExpr(e.Keys[n])
			c.compileExpr(e.Values[n])
			n++
		}
		c.emitArg(runtime.OpBuildMap, n)
		for i := n; i < len(e.Keys); i++ {
			if e.Keys[i] == nil {
				c.compileExpr(e.Values[i])
				c.emitArg(runtime.OpDictUpdate, 0)
			} else {
				c.compileExpr(e.Keys[i])
				c.compileExpr(e.Values[i])
				c.emitArg(runtime.OpMapAdd, 0)
			}
		}

	case *model.Set:
		n := 0
		for n < len(e.Elts) && !isStarred(e.Elts[n]) {
			c.compileExpr(e.Elts[n])
			n++
		}
		c.emitArg(runtime.OpBuildSet, n)
		for _, elt := range e.Elts[n:] {
			if starred, ok := elt.(*model.Starred); ok {
				c.compileExpr(starred.Value)
				c.emitArg(runtime.OpSetUpdate, 0)
			} else {
				c.compileExpr(elt)
				c.emitArg(runtime.OpSetAdd, 0)
			}
		}

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
				derefIdx := sym.Index
				if sym.Scope == ScopeFree {
					derefIdx = len(c.symbolTable.cellSyms) + sym.Index
				}
				c.emitArg(runtime.OpStoreDeref, derefIdx)
			}
		} else {
			c.compileStore(e.Target)
		}

	default:
		c.error(expr.Pos(), "unsupported expression type: %T", expr)
	}
}

// opMapping maps binary operator token kinds to their binary and inplace opcodes.
var opMapping = map[model.TokenKind]struct{ binary, inplace runtime.Opcode }{
	model.TK_Plus:        {runtime.OpBinaryAdd, runtime.OpInplaceAdd},
	model.TK_Minus:       {runtime.OpBinarySubtract, runtime.OpInplaceSubtract},
	model.TK_Star:        {runtime.OpBinaryMultiply, runtime.OpInplaceMultiply},
	model.TK_Slash:       {runtime.OpBinaryDivide, runtime.OpInplaceDivide},
	model.TK_DoubleSlash: {runtime.OpBinaryFloorDiv, runtime.OpInplaceFloorDiv},
	model.TK_Percent:     {runtime.OpBinaryModulo, runtime.OpInplaceModulo},
	model.TK_DoubleStar:  {runtime.OpBinaryPower, runtime.OpInplacePower},
	model.TK_At:          {runtime.OpBinaryMatMul, runtime.OpInplaceMatMul},
	model.TK_LShift:      {runtime.OpBinaryLShift, runtime.OpInplaceLShift},
	model.TK_RShift:      {runtime.OpBinaryRShift, runtime.OpInplaceRShift},
	model.TK_Ampersand:   {runtime.OpBinaryAnd, runtime.OpInplaceAnd},
	model.TK_Pipe:        {runtime.OpBinaryOr, runtime.OpInplaceOr},
	model.TK_Caret:       {runtime.OpBinaryXor, runtime.OpInplaceXor},
}

// augAssignToOp maps augmented assignment token kinds to their corresponding binary operator token kind.
var augAssignToOp = map[model.TokenKind]model.TokenKind{
	model.TK_PlusAssign:        model.TK_Plus,
	model.TK_MinusAssign:       model.TK_Minus,
	model.TK_StarAssign:        model.TK_Star,
	model.TK_SlashAssign:       model.TK_Slash,
	model.TK_DoubleSlashAssign: model.TK_DoubleSlash,
	model.TK_PercentAssign:     model.TK_Percent,
	model.TK_DoubleStarAssign:  model.TK_DoubleStar,
	model.TK_AtAssign:          model.TK_At,
	model.TK_LShiftAssign:      model.TK_LShift,
	model.TK_RShiftAssign:      model.TK_RShift,
	model.TK_AmpersandAssign:   model.TK_Ampersand,
	model.TK_PipeAssign:        model.TK_Pipe,
	model.TK_CaretAssign:       model.TK_Caret,
}

func (c *Compiler) emitBinaryOp(op model.TokenKind) {
	if entry, ok := opMapping[op]; ok {
		c.emit(entry.binary)
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
	// Check if we need the extended call protocol (for *args or **kwargs unpacking)
	hasStar := false
	hasDoubleStar := false
	for _, arg := range e.Args {
		if _, ok := arg.(*model.Starred); ok {
			hasStar = true
			break
		}
	}
	for _, kw := range e.Keywords {
		if kw.Arg == nil {
			// **kwargs unpacking
			hasDoubleStar = true
			break
		}
	}

	if hasStar || hasDoubleStar {
		// Extended call: build the positional args tuple (unpacking *iterables)
		// and, if there are keywords, a kwargs dict (merging **mappings, where
		// a repeated keyword is a TypeError), then CALL_EX.
		c.compileExpr(e.Func)
		c.compileStarredList(e.Args)
		c.emit(runtime.OpListToTuple)

		flags := 0
		if len(e.Keywords) > 0 {
			flags = 1
			c.emitArg(runtime.OpBuildMap, 0)
			named := 0
			flushNamed := func() {
				if named > 0 {
					c.emitArg(runtime.OpBuildMap, named)
					c.emitArg(runtime.OpDictMerge, 0)
					named = 0
				}
			}
			for _, kw := range e.Keywords {
				if kw.Arg == nil {
					flushNamed()
					c.compileExpr(kw.Value)
					c.emitArg(runtime.OpDictMerge, 0)
				} else {
					c.emitLoadConst(kw.Arg.Name)
					c.compileExpr(kw.Value)
					named++
				}
			}
			flushNamed()
		}

		c.emitArg(runtime.OpCallEx, flags)
	} else {
		// Standard call path (no unpacking)
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
}

// Variable access compilation

func (c *Compiler) compileLoad(name string) {
	sym, _ := c.symbolTable.Resolve(name)
	switch sym.Scope {
	case ScopeLocal:
		c.emitArg(runtime.OpLoadFast, sym.Index)
	case ScopeGlobal, ScopeBuiltin:
		idx := c.addName(name)
		// In class scope, use OpLoadName for normal names (checks class namespace + enclosing globals),
		// but use OpLoadGlobal for explicitly global-declared names (checks module globals directly)
		if c.symbolTable.scopeType == ScopeClass && !c.symbolTable.globals[name] {
			c.emitArg(runtime.OpLoadName, idx)
		} else {
			c.emitArg(runtime.OpLoadGlobal, idx)
		}
	case ScopeFree:
		c.emitArg(runtime.OpLoadDeref, len(c.symbolTable.cellSyms)+sym.Index)
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
			if c.symbolTable.scopeType == ScopeClass && !c.symbolTable.globals[t.Name] {
				// In class scope, non-explicitly-global names are class attributes
				c.emitArg(runtime.OpStoreName, idx)
			} else {
				c.emitArg(runtime.OpStoreGlobal, idx)
			}
		case ScopeFree, ScopeCell:
			// In class scope, assignments create class attributes (OpStoreName),
			// not modifications to the enclosing scope (OpStoreDeref),
			// unless explicitly declared nonlocal.
			if c.symbolTable.scopeType == ScopeClass && !c.symbolTable.nonlocals[t.Name] {
				idx := c.addName(t.Name)
				c.emitArg(runtime.OpStoreName, idx)
			} else {
				derefIdx := sym.Index
				if sym.Scope == ScopeFree {
					derefIdx = len(c.symbolTable.cellSyms) + sym.Index
				}
				c.emitArg(runtime.OpStoreDeref, derefIdx)
			}
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
		// Check for starred element
		starIdx := -1
		for i, elt := range elts {
			if _, ok := elt.(*model.Starred); ok {
				starIdx = i
				break
			}
		}
		if starIdx >= 0 {
			// Emit OpUnpackEx: arg = countBefore | (countAfter << 8)
			countBefore := starIdx
			countAfter := len(elts) - starIdx - 1
			c.emitArg(runtime.OpUnpackEx, countBefore|(countAfter<<8))
		} else {
			c.emitArg(runtime.OpUnpackSequence, len(elts))
		}
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
		case ScopeFree, ScopeCell:
			derefIdx := sym.Index
			if sym.Scope == ScopeFree {
				derefIdx = len(c.symbolTable.cellSyms) + sym.Index
			}
			c.emitArg(runtime.OpDeleteDeref, derefIdx)
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

// isStarred reports whether e is a *starred element.
func isStarred(e model.Expr) bool {
	_, ok := e.(*model.Starred)
	return ok
}

// hasStarred reports whether any element is a *starred element.
func hasStarred(elts []model.Expr) bool {
	for _, e := range elts {
		if isStarred(e) {
			return true
		}
	}
	return false
}

// compileStarredList pushes a list of elts, unpacking *starred elements in
// place: leading plain elements go into BUILD_LIST, and the rest are added in
// order with LIST_EXTEND / LIST_APPEND.
func (c *Compiler) compileStarredList(elts []model.Expr) {
	n := 0
	for n < len(elts) && !isStarred(elts[n]) {
		c.compileExpr(elts[n])
		n++
	}
	c.emitArg(runtime.OpBuildList, n)
	for _, elt := range elts[n:] {
		if starred, ok := elt.(*model.Starred); ok {
			c.compileExpr(starred.Value)
			c.emitArg(runtime.OpListExtend, 0)
		} else {
			c.compileExpr(elt)
			c.emitArg(runtime.OpListAppend, 0)
		}
	}
}
