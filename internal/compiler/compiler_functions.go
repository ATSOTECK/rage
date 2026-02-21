package compiler

import (
	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// Function and class compilation

func (c *Compiler) compileFunctionDef(s *model.FunctionDef) {
	// Compile decorators
	for _, dec := range s.Decorators {
		c.compileExpr(dec)
	}

	// Create a new compiler for the function body
	funcCompiler := c.newChildCompiler(s.Name.Name, s.StartPos.Line, ScopeFunction, 0)

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

	// Pre-scan: identify all assigned names in the function body and pre-define
	// them as locals. This ensures that variables assigned anywhere in the function
	// are treated as local for ALL references (matching Python's scoping rules).
	// Without this, `x = x + 1` without `nonlocal` would incorrectly capture x
	// from the enclosing scope instead of raising UnboundLocalError.
	predefineAssignedLocals(funcCompiler.symbolTable, s.Body)

	// Pre-scan for captured variables: identify locals that are referenced by
	// inner functions/lambdas/comprehensions and mark them as cell variables
	// BEFORE body compilation. This ensures assignments to captured variables
	// emit OpStoreDeref from the start, enabling proper closure sharing.
	prescanCapturedVariables(funcCompiler.symbolTable, s.Body)

	// Compile function body
	for _, stmt := range s.Body {
		funcCompiler.compileStmt(stmt)
	}

	// Add implicit return None
	funcCompiler.emitLoadConst(nil)
	funcCompiler.emit(runtime.OpReturn)
	c.finalizeAndOptimize(funcCompiler)

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

	// Load kwonly defaults as a dict
	hasKwDefaults := false
	if s.Args != nil && len(s.Args.KwDefaults) > 0 {
		// Build a dict of kwonly param name -> default value
		count := 0
		for i, def := range s.Args.KwDefaults {
			if def != nil && i < len(s.Args.KwOnlyArgs) {
				c.emitLoadConst(s.Args.KwOnlyArgs[i].Arg.Name)
				c.compileExpr(def)
				count++
			}
		}
		if count > 0 {
			c.emitArg(runtime.OpBuildMap, count)
			hasKwDefaults = true
		}
	}

	// Load code object and make function
	c.emitLoadConst(funcCode)
	c.emitLoadConst(s.Name.Name)

	flags := 0
	if s.Args != nil && len(s.Args.Defaults) > 0 {
		flags |= 1 // Has defaults
	}
	if hasKwDefaults {
		flags |= 2 // Has kwonly defaults
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
	classCompiler := c.newChildCompiler(s.Name.Name, s.StartPos.Line, ScopeClass, 0)

	// If any method uses super(), define __class__ as a cell in the class scope
	if needsClassCell {
		classCompiler.symbolTable.Define("__class__")
		classCompiler.symbolTable.MarkAsCell("__class__")
	}

	// Check if any statement is an annotated assignment and emit SETUP_ANNOTATIONS
	for _, stmt := range s.Body {
		if _, ok := stmt.(*model.AnnAssign); ok {
			classCompiler.emit(runtime.OpSetupAnnotations)
			break
		}
	}

	// Compile class body
	for _, stmt := range s.Body {
		classCompiler.compileStmt(stmt)
	}
	classCompiler.emit(runtime.OpLoadLocals)
	classCompiler.emit(runtime.OpReturn)
	c.finalizeAndOptimize(classCompiler)

	c.emitLoadConst(classCompiler.code)
	c.emitLoadConst(s.Name.Name)
	c.emitArg(runtime.OpMakeFunction, 0)

	// Class name
	c.emitLoadConst(s.Name.Name)

	// Compile bases
	for _, base := range s.Bases {
		c.compileExpr(base)
	}

	// Compile keywords (values only, then keyword names tuple)
	if len(s.Keywords) > 0 {
		kwNames := make([]string, 0, len(s.Keywords))
		for _, kw := range s.Keywords {
			c.compileExpr(kw.Value)
			if kw.Arg != nil {
				kwNames = append(kwNames, kw.Arg.Name)
			}
		}
		c.emitLoadConst(kwNames)
		c.emitArg(runtime.OpCallKw, 2+len(s.Bases)+len(s.Keywords))
	} else {
		c.emitArg(runtime.OpCall, 2+len(s.Bases))
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
	lambdaCompiler := c.newChildCompiler("<lambda>", e.StartPos.Line, ScopeFunction, 0)

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
	c.finalizeAndOptimize(lambdaCompiler)

	lambdaCode := lambdaCompiler.code
	if e.Args != nil {
		lambdaCode.ArgCount = len(e.Args.Args)
	}

	// Check if the lambda body contains yield (making it a generator lambda)
	if containsYieldExpr(e.Body) {
		lambdaCode.Flags |= runtime.FlagGenerator
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

// compileComprehension is the shared implementation for list/set/dict comprehensions
// and generator expressions. It creates a child compiler, compiles the comprehension
// body, finalizes the code, and emits the function call.
func (c *Compiler) compileComprehension(
	name string,
	firstLine int,
	flags runtime.CodeFlags,
	generators []*model.Comprehension,
	firstIter model.Expr,
	eltExprs []model.Expr, // element expressions for prescan (capture detection)
	setupBody func(cc *Compiler, stackOffset int),
) {
	cc := c.newChildCompiler(name, firstLine, ScopeComprehension, flags)
	cc.symbolTable.Define(".0")

	// Pre-define generator target variables so prescan can mark them as cells
	for _, gen := range generators {
		assigned := make(map[string]bool)
		collectNamesFromExpr(gen.Target, assigned)
		for aName := range assigned {
			if _, exists := cc.symbolTable.symbols[aName]; !exists {
				cc.symbolTable.Define(aName)
			}
		}
	}

	// Prescan element expressions and conditions for lambdas/inner functions
	// that capture comprehension-local variables
	refs := make(map[string]bool)
	for _, elt := range eltExprs {
		captureWalkExpr(elt, refs, false)
	}
	for _, gen := range generators {
		for _, cond := range gen.Ifs {
			captureWalkExpr(cond, refs, false)
		}
	}
	for rName := range refs {
		if sym, ok := cc.symbolTable.symbols[rName]; ok && sym.Scope == ScopeLocal {
			cc.symbolTable.MarkAsCell(rName)
		}
	}

	stackOffset := len(generators)
	setupBody(cc, stackOffset)

	if flags&runtime.FlagGenerator != 0 {
		cc.emitLoadConst(nil)
	}
	cc.emit(runtime.OpReturn)

	c.finalizeAndOptimize(cc)
	cc.code.ArgCount = 1

	// Make function and call with iterator
	c.compileExpr(firstIter)
	c.emit(runtime.OpGetIter)
	c.emitLoadConst(cc.code)
	c.emitLoadConst(name)
	c.emitArg(runtime.OpMakeFunction, 0)
	c.emit(runtime.OpRot2)
	c.emitArg(runtime.OpCall, 1)
}

func (c *Compiler) compileListComp(e *model.ListComp) {
	c.compileComprehension("<listcomp>", e.StartPos.Line, 0, e.Generators, e.Generators[0].Iter,
		[]model.Expr{e.Elt},
		func(cc *Compiler, stackOffset int) {
			cc.emitArg(runtime.OpBuildList, 0)
			c.compileComprehensionGenerators(cc, e.Generators, func() {
				cc.compileExpr(e.Elt)
				cc.emitArg(runtime.OpListAppend, stackOffset)
			}, 0)
		})
}

func (c *Compiler) compileSetComp(e *model.SetComp) {
	c.compileComprehension("<setcomp>", e.StartPos.Line, 0, e.Generators, e.Generators[0].Iter,
		[]model.Expr{e.Elt},
		func(cc *Compiler, stackOffset int) {
			cc.emitArg(runtime.OpBuildSet, 0)
			c.compileComprehensionGenerators(cc, e.Generators, func() {
				cc.compileExpr(e.Elt)
				cc.emitArg(runtime.OpSetAdd, stackOffset)
			}, 0)
		})
}

func (c *Compiler) compileDictComp(e *model.DictComp) {
	c.compileComprehension("<dictcomp>", e.StartPos.Line, 0, e.Generators, e.Generators[0].Iter,
		[]model.Expr{e.Key, e.Value},
		func(cc *Compiler, stackOffset int) {
			cc.emitArg(runtime.OpBuildMap, 0)
			c.compileComprehensionGenerators(cc, e.Generators, func() {
				cc.compileExpr(e.Key)
				cc.compileExpr(e.Value)
				cc.emitArg(runtime.OpMapAdd, stackOffset)
			}, 0)
		})
}

func (c *Compiler) compileGeneratorExpr(e *model.GeneratorExpr) {
	c.compileComprehension("<genexpr>", e.StartPos.Line, runtime.FlagGenerator, e.Generators, e.Generators[0].Iter,
		[]model.Expr{e.Elt},
		func(cc *Compiler, stackOffset int) {
			c.compileComprehensionGenerators(cc, e.Generators, func() {
				cc.compileExpr(e.Elt)
				cc.emit(runtime.OpYieldValue)
				cc.emit(runtime.OpPop)
			}, 0)
		})
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

	// Define and store target (only if not already defined by prescan)
	switch t := gen.Target.(type) {
	case *model.Identifier:
		if _, exists := comp.symbolTable.symbols[t.Name]; !exists {
			comp.symbolTable.Define(t.Name)
		}
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

// Match statement compilation

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
