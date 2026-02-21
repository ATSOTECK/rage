package compiler

import (
	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

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
		// In class scope, store annotation in __annotations__ dict
		if c.symbolTable.scopeType == ScopeClass {
			if ident, ok := s.Target.(*model.Identifier); ok {
				// Stack order for STORE_SUBSCR: val, obj, index
				c.compileExpr(s.Annotation)         // val: the annotation type
				annIdx := c.addName("__annotations__")
				c.emitArg(runtime.OpLoadName, annIdx) // obj: __annotations__ dict
				c.emitLoadConst(ident.Name)           // index: field name string
				c.emit(runtime.OpStoreSubscr)
			}
		}

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
		if c.finallyDepth > 0 {
			// Inside a try/finally — use OpContinueLoop so the generator can
			// run the finally block before jumping to the loop target
			jump := c.emitJump(runtime.OpContinueLoop)
			c.loopStack[len(c.loopStack)-1].continueJumps = append(
				c.loopStack[len(c.loopStack)-1].continueJumps, jump)
		} else {
			jump := c.emitJump(runtime.OpJump)
			c.loopStack[len(c.loopStack)-1].continueJumps = append(
				c.loopStack[len(c.loopStack)-1].continueJumps, jump)
		}

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
			c.compileStore(&model.Identifier{Name: storeName})
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
				c.compileStore(&model.Identifier{Name: storeName})
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
	if binOp, ok := augAssignToOp[s.Op]; ok {
		if entry, ok := opMapping[binOp]; ok {
			c.emit(entry.inplace)
		}
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
	// Check if any handler uses except* syntax
	for _, h := range s.Handlers {
		if h.IsStar {
			c.compileTryStar(s)
			return
		}
	}

	hasFinally := len(s.FinalBody) > 0
	hasExcept := len(s.Handlers) > 0

	// If we have a finally block, set it up first (it wraps everything)
	var finallyJump int
	if hasFinally {
		finallyJump = c.emitJump(runtime.OpSetupFinally)
		c.finallyDepth++
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

				// Mark end of except handler (pops excHandlerStack)
				c.emit(runtime.OpPopExceptHandler)

				handlerEnds = append(handlerEnds, c.emitJump(runtime.OpJump))
				c.patchJump(nextHandler, c.currentOffset())
			} else {
				// Bare except - catches everything
				c.emit(runtime.OpClearException) // Clear the current exception state
				c.emit(runtime.OpPop)
				for _, stmt := range handler.Body {
					c.compileStmt(stmt)
				}
				// Mark end of except handler (pops excHandlerStack)
				c.emit(runtime.OpPopExceptHandler)
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
		c.finallyDepth--

		// Pop the BlockFinally from the block stack in the normal (no-exception) flow.
		// The exception flow enters at the handler address below, skipping this PopBlock,
		// because handleException already popped the block.
		c.emit(runtime.OpPopBlock)

		// Patch the finally setup to jump here (exception flow entry point)
		c.patchJump(finallyJump, c.currentOffset())

		// Compile finally body
		for _, stmt := range s.FinalBody {
			c.compileStmt(stmt)
		}

		// End finally - will re-raise exception if one was active
		c.emit(runtime.OpEndFinally)
	}
}

func (c *Compiler) compileTryStar(s *model.Try) {
	hasFinally := len(s.FinalBody) > 0

	// If we have a finally block, set it up first (it wraps everything)
	var finallyJump int
	if hasFinally {
		finallyJump = c.emitJump(runtime.OpSetupFinally)
	}

	// Setup except* handler
	handlerJump := c.emitJump(runtime.OpSetupExceptStar)

	// Compile try body
	for _, stmt := range s.Body {
		c.compileStmt(stmt)
	}

	// Pop except* handler (try body completed normally)
	c.emit(runtime.OpPopExcept)
	successJump := c.emitJump(runtime.OpJump)

	// Handler entry: stack has [eg]
	c.patchJump(handlerJump, c.currentOffset())

	// Each except* handler checks its type against the exception group
	// Stack invariant: [eg] at start and end of each handler
	for _, handler := range s.Handlers {
		// DUP the eg for EXCEPT_STAR_MATCH to consume: [eg, eg]
		c.emit(runtime.OpDup)
		// Compile the exception type: [eg, eg, type]
		c.compileExpr(handler.Type)
		// EXCEPT_STAR_MATCH pops eg_dup + type, pushes result: [eg, result]
		c.emit(runtime.OpExceptStarMatch)

		// DUP result for test: [eg, result, result]
		c.emit(runtime.OpDup)
		// POP_JUMP_IF_FALSE pops copy: [eg, result]
		skipJump := c.emitJump(runtime.OpPopJumpIfFalse)

		// Matched — clear exception state and store/pop subgroup
		c.emit(runtime.OpClearException)
		if handler.Name != nil {
			c.compileStore(handler.Name) // pops result: [eg]
		} else {
			c.emit(runtime.OpPop) // pops result: [eg]
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

		afterBodyJump := c.emitJump(runtime.OpJump)

		// skip: no match — pop the None result: [eg]
		c.patchJump(skipJump, c.currentOffset())
		c.emit(runtime.OpPop) // pop None: [eg]

		// next_handler:
		c.patchJump(afterBodyJump, c.currentOffset())
	}

	// EXCEPT_STAR_RERAISE: re-raise remaining or clean up
	c.emit(runtime.OpExceptStarReraise)

	// After handler completes, skip over the else body
	var skipElseJump int
	if len(s.OrElse) > 0 {
		skipElseJump = c.emitJump(runtime.OpJump)
	}

	// Else clause (runs if no exception was raised)
	c.patchJump(successJump, c.currentOffset())
	for _, stmt := range s.OrElse {
		c.compileStmt(stmt)
	}

	if len(s.OrElse) > 0 {
		c.patchJump(skipElseJump, c.currentOffset())
	}

	// Finally clause
	if hasFinally {
		c.patchJump(finallyJump, c.currentOffset())
		for _, stmt := range s.FinalBody {
			c.compileStmt(stmt)
		}
		c.emit(runtime.OpEndFinally)
	}
}

func (c *Compiler) compileWith(s *model.With) {
	c.compileWithItem(s.Items, 0, s.Body)
}

func (c *Compiler) compileWithItem(items []*model.WithItem, idx int, body []model.Stmt) {
	item := items[idx]

	// Compile the context expression: stack: [..., cm]
	c.compileExpr(item.ContextExpr)

	// DUP so we have cm for both __enter__ and later __exit__
	c.emit(runtime.OpDup) // stack: [..., cm, cm]

	// Call __enter__
	enterIdx := c.addName("__enter__")
	c.emitArg(runtime.OpLoadMethod, enterIdx)
	c.emitArg(runtime.OpCallMethod, 0) // stack: [..., cm, enter_result]

	// Store or pop the __enter__ result BEFORE setting up the block,
	// so block.Level captures SP with only cm on the stack.
	if item.OptionalVar != nil {
		c.compileStore(item.OptionalVar) // stack: [..., cm]
	} else {
		c.emit(runtime.OpPop) // stack: [..., cm]
	}

	// Setup with block (push BlockWith onto block stack)
	// block.Level = SP here, with cm at stack[SP-1]
	cleanupJump := c.emitJump(runtime.OpSetupWith) // stack: [..., cm]

	// Compile body (or next nested with item)
	if idx < len(items)-1 {
		c.compileWithItem(items, idx+1, body)
	} else {
		for _, stmt := range body {
			c.compileStmt(stmt)
		}
	}

	// Normal exit: pop the BlockWith block
	c.emit(runtime.OpPopExcept) // stack: [..., cm]

	// Jump over the exception cleanup handler
	normalJump := c.emitJump(runtime.OpJump)

	// === Exception cleanup handler ===
	// handleException pushes the exception and jumps here
	// stack: [..., cm, exception]
	c.patchJump(cleanupJump, c.currentOffset())

	c.emit(runtime.OpWithCleanup) // calls __exit__(exc_type, exc_val, exc_tb), may suppress
	c.emit(runtime.OpEndFinally)  // re-raises if currentException still set
	// If we reach here, exception was suppressed — skip normal exit
	skipJump := c.emitJump(runtime.OpJump)

	// === Normal exit path ===
	c.patchJump(normalJump, c.currentOffset())
	// stack: [..., cm] — call __exit__(None, None, None)
	exitIdx := c.addName("__exit__")
	c.emitArg(runtime.OpLoadMethod, exitIdx)
	c.emitLoadConst(nil) // exc_type
	c.emitLoadConst(nil) // exc_val
	c.emitLoadConst(nil) // exc_tb
	c.emitArg(runtime.OpCallMethod, 3)
	c.emit(runtime.OpPop) // discard __exit__ return value

	c.patchJump(skipJump, c.currentOffset())
}
