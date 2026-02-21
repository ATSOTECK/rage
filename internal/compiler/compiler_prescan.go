package compiler

import "github.com/ATSOTECK/rage/internal/model"

// Function and class compilation

// predefineAssignedLocals scans a function body to find all names that are assigned
// and pre-defines them as locals in the symbol table. This ensures Python-correct scoping:
// a variable assigned anywhere in a function is local throughout that function, so reading
// it before assignment raises UnboundLocalError instead of capturing from an outer scope.
func predefineAssignedLocals(st *SymbolTable, stmts []model.Stmt) {
	// Collect global and nonlocal declarations first
	globals := make(map[string]bool)
	nonlocals := make(map[string]bool)
	collectDeclarations(stmts, globals, nonlocals)

	// Collect all assigned names
	assigned := make(map[string]bool)
	for _, stmt := range stmts {
		collectAssignedNames(stmt, assigned)
	}

	// Pre-define as locals: names that are assigned but not declared global/nonlocal
	// and not already defined (e.g., as a parameter)
	for name := range assigned {
		if globals[name] || nonlocals[name] {
			continue
		}
		if _, exists := st.symbols[name]; exists {
			continue // Already defined (parameter, etc.)
		}
		st.Define(name)
	}
}

// collectDeclarations finds all global and nonlocal declarations in statements.
// Does not descend into nested functions, classes, or comprehensions.
func collectDeclarations(stmts []model.Stmt, globals, nonlocals map[string]bool) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *model.Global:
			for _, name := range s.Names {
				globals[name.Name] = true
			}
		case *model.Nonlocal:
			for _, name := range s.Names {
				nonlocals[name.Name] = true
			}
		case *model.If:
			collectDeclarations(s.Body, globals, nonlocals)
			collectDeclarations(s.OrElse, globals, nonlocals)
		case *model.For:
			collectDeclarations(s.Body, globals, nonlocals)
			collectDeclarations(s.OrElse, globals, nonlocals)
		case *model.While:
			collectDeclarations(s.Body, globals, nonlocals)
			collectDeclarations(s.OrElse, globals, nonlocals)
		case *model.Try:
			collectDeclarations(s.Body, globals, nonlocals)
			for _, handler := range s.Handlers {
				collectDeclarations(handler.Body, globals, nonlocals)
			}
			collectDeclarations(s.OrElse, globals, nonlocals)
			collectDeclarations(s.FinalBody, globals, nonlocals)
		case *model.With:
			collectDeclarations(s.Body, globals, nonlocals)
		}
		// Don't descend into FunctionDef, ClassDef — they have their own scope
	}
}

// collectAssignedNames finds all names that are targets of assignment in statements.
// Does not descend into nested functions, classes, or comprehensions (they have their own scope).
func collectAssignedNames(stmt model.Stmt, names map[string]bool) {
	switch s := stmt.(type) {
	case *model.Assign:
		for _, target := range s.Targets {
			collectNamesFromExpr(target, names)
		}
	case *model.AugAssign:
		collectNamesFromExpr(s.Target, names)
	case *model.AnnAssign:
		if s.Value != nil {
			collectNamesFromExpr(s.Target, names)
		}
	case *model.For:
		collectNamesFromExpr(s.Target, names)
		for _, bodyStmt := range s.Body {
			collectAssignedNames(bodyStmt, names)
		}
		for _, elseStmt := range s.OrElse {
			collectAssignedNames(elseStmt, names)
		}
	case *model.While:
		for _, bodyStmt := range s.Body {
			collectAssignedNames(bodyStmt, names)
		}
		for _, elseStmt := range s.OrElse {
			collectAssignedNames(elseStmt, names)
		}
	case *model.If:
		for _, bodyStmt := range s.Body {
			collectAssignedNames(bodyStmt, names)
		}
		for _, elseStmt := range s.OrElse {
			collectAssignedNames(elseStmt, names)
		}
	case *model.With:
		for _, item := range s.Items {
			if item.OptionalVar != nil {
				collectNamesFromExpr(item.OptionalVar, names)
			}
		}
		for _, bodyStmt := range s.Body {
			collectAssignedNames(bodyStmt, names)
		}
	case *model.Try:
		for _, bodyStmt := range s.Body {
			collectAssignedNames(bodyStmt, names)
		}
		for _, handler := range s.Handlers {
			if handler.Name != nil {
				names[handler.Name.Name] = true
			}
			for _, bodyStmt := range handler.Body {
				collectAssignedNames(bodyStmt, names)
			}
		}
		for _, elseStmt := range s.OrElse {
			collectAssignedNames(elseStmt, names)
		}
		for _, finalStmt := range s.FinalBody {
			collectAssignedNames(finalStmt, names)
		}
	case *model.FunctionDef:
		// The function name itself is assigned in this scope
		names[s.Name.Name] = true
		// Don't descend into function body — it has its own scope
	case *model.ClassDef:
		// The class name itself is assigned in this scope
		names[s.Name.Name] = true
		// Don't descend into class body — it has its own scope
	case *model.Import:
		for _, alias := range s.Names {
			if alias.AsName != nil {
				names[alias.AsName.Name] = true
			} else {
				// import foo.bar -> binds 'foo'
				parts := alias.Name.Name
				for i, ch := range parts {
					if ch == '.' {
						names[parts[:i]] = true
						break
					}
					if i == len(parts)-1 {
						names[parts] = true
					}
				}
			}
		}
	case *model.ImportFrom:
		for _, alias := range s.Names {
			if alias.AsName != nil {
				names[alias.AsName.Name] = true
			} else {
				names[alias.Name.Name] = true
			}
		}
	case *model.Match:
		for _, mc := range s.Cases {
			for _, bodyStmt := range mc.Body {
				collectAssignedNames(bodyStmt, names)
			}
		}
	}
}

// collectNamesFromExpr extracts simple identifier names from assignment targets.
func collectNamesFromExpr(expr model.Expr, names map[string]bool) {
	switch e := expr.(type) {
	case *model.Identifier:
		names[e.Name] = true
	case *model.Tuple:
		for _, elt := range e.Elts {
			collectNamesFromExpr(elt, names)
		}
	case *model.List:
		for _, elt := range e.Elts {
			collectNamesFromExpr(elt, names)
		}
	case *model.Starred:
		collectNamesFromExpr(e.Value, names)
	}
	// Attribute, Subscript — not simple names, skip
}

// prescanCapturedVariables pre-marks local variables as cell variables if they
// are referenced by any inner scope (function, lambda, class, comprehension).
// This must run after predefineAssignedLocals but before body compilation,
// so that assignments to captured variables emit OpStoreDeref from the start.
func prescanCapturedVariables(st *SymbolTable, stmts []model.Stmt) {
	refs := make(map[string]bool)
	for _, stmt := range stmts {
		captureWalkStmt(stmt, refs, false)
	}
	for name := range refs {
		if sym, ok := st.symbols[name]; ok && sym.Scope == ScopeLocal {
			st.MarkAsCell(name)
		}
	}
}

// captureWalkStmt walks an AST subtree. When collecting is true, all identifier
// references are gathered. When false, it recurses through control flow but only
// starts collecting when it hits an inner scope boundary (function/class/lambda/comprehension).
func captureWalkStmt(stmt model.Stmt, refs map[string]bool, collecting bool) {
	switch s := stmt.(type) {
	case *model.FunctionDef:
		// Inner scope boundary — collect all ident refs from body
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, true)
		}
		// Default values run in our scope, not inner
		if s.Args != nil {
			for _, d := range s.Args.Defaults {
				captureWalkExpr(d, refs, collecting)
			}
			for _, d := range s.Args.KwDefaults {
				if d != nil {
					captureWalkExpr(d, refs, collecting)
				}
			}
		}
		// Decorators run in our scope
		for _, dec := range s.Decorators {
			captureWalkExpr(dec, refs, collecting)
		}
	case *model.ClassDef:
		// Class body is an inner scope
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, true)
		}
		for _, dec := range s.Decorators {
			captureWalkExpr(dec, refs, collecting)
		}
		for _, base := range s.Bases {
			captureWalkExpr(base, refs, collecting)
		}
	case *model.ExprStmt:
		captureWalkExpr(s.Value, refs, collecting)
	case *model.Assign:
		for _, t := range s.Targets {
			captureWalkExpr(t, refs, collecting)
		}
		captureWalkExpr(s.Value, refs, collecting)
	case *model.AugAssign:
		captureWalkExpr(s.Target, refs, collecting)
		captureWalkExpr(s.Value, refs, collecting)
	case *model.AnnAssign:
		if s.Value != nil {
			captureWalkExpr(s.Value, refs, collecting)
		}
	case *model.Return:
		if s.Value != nil {
			captureWalkExpr(s.Value, refs, collecting)
		}
	case *model.For:
		captureWalkExpr(s.Target, refs, collecting)
		captureWalkExpr(s.Iter, refs, collecting)
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, collecting)
		}
		for _, elseStmt := range s.OrElse {
			captureWalkStmt(elseStmt, refs, collecting)
		}
	case *model.While:
		captureWalkExpr(s.Test, refs, collecting)
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, collecting)
		}
		for _, elseStmt := range s.OrElse {
			captureWalkStmt(elseStmt, refs, collecting)
		}
	case *model.If:
		captureWalkExpr(s.Test, refs, collecting)
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, collecting)
		}
		for _, elseStmt := range s.OrElse {
			captureWalkStmt(elseStmt, refs, collecting)
		}
	case *model.With:
		for _, item := range s.Items {
			captureWalkExpr(item.ContextExpr, refs, collecting)
			if item.OptionalVar != nil {
				captureWalkExpr(item.OptionalVar, refs, collecting)
			}
		}
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, collecting)
		}
	case *model.Try:
		for _, bodyStmt := range s.Body {
			captureWalkStmt(bodyStmt, refs, collecting)
		}
		for _, handler := range s.Handlers {
			for _, bodyStmt := range handler.Body {
				captureWalkStmt(bodyStmt, refs, collecting)
			}
		}
		for _, elseStmt := range s.OrElse {
			captureWalkStmt(elseStmt, refs, collecting)
		}
		for _, finalStmt := range s.FinalBody {
			captureWalkStmt(finalStmt, refs, collecting)
		}
	case *model.Raise:
		if s.Exc != nil {
			captureWalkExpr(s.Exc, refs, collecting)
		}
		if s.Cause != nil {
			captureWalkExpr(s.Cause, refs, collecting)
		}
	case *model.Assert:
		if s.Test != nil {
			captureWalkExpr(s.Test, refs, collecting)
		}
		if s.Msg != nil {
			captureWalkExpr(s.Msg, refs, collecting)
		}
	case *model.Delete:
		for _, target := range s.Targets {
			captureWalkExpr(target, refs, collecting)
		}
	case *model.Match:
		captureWalkExpr(s.Subject, refs, collecting)
		for _, c := range s.Cases {
			for _, bodyStmt := range c.Body {
				captureWalkStmt(bodyStmt, refs, collecting)
			}
		}
	}
}

// captureWalkExpr walks an expression subtree. When collecting is true, identifiers
// are gathered. Lambda and comprehension boundaries always start collecting.
func captureWalkExpr(expr model.Expr, refs map[string]bool, collecting bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *model.Identifier:
		if collecting {
			refs[e.Name] = true
		}
	case *model.Lambda:
		// Lambda body is always an inner scope
		captureWalkExpr(e.Body, refs, true)
	case *model.ListComp:
		captureWalkExpr(e.Elt, refs, true)
		for _, gen := range e.Generators {
			captureWalkExpr(gen.Iter, refs, true)
			for _, cond := range gen.Ifs {
				captureWalkExpr(cond, refs, true)
			}
		}
	case *model.SetComp:
		captureWalkExpr(e.Elt, refs, true)
		for _, gen := range e.Generators {
			captureWalkExpr(gen.Iter, refs, true)
			for _, cond := range gen.Ifs {
				captureWalkExpr(cond, refs, true)
			}
		}
	case *model.DictComp:
		captureWalkExpr(e.Key, refs, true)
		captureWalkExpr(e.Value, refs, true)
		for _, gen := range e.Generators {
			captureWalkExpr(gen.Iter, refs, true)
			for _, cond := range gen.Ifs {
				captureWalkExpr(cond, refs, true)
			}
		}
	case *model.GeneratorExpr:
		captureWalkExpr(e.Elt, refs, true)
		for _, gen := range e.Generators {
			captureWalkExpr(gen.Iter, refs, true)
			for _, cond := range gen.Ifs {
				captureWalkExpr(cond, refs, true)
			}
		}
	case *model.Call:
		captureWalkExpr(e.Func, refs, collecting)
		for _, arg := range e.Args {
			captureWalkExpr(arg, refs, collecting)
		}
		for _, kw := range e.Keywords {
			captureWalkExpr(kw.Value, refs, collecting)
		}
	case *model.BinaryOp:
		captureWalkExpr(e.Left, refs, collecting)
		captureWalkExpr(e.Right, refs, collecting)
	case *model.UnaryOp:
		captureWalkExpr(e.Operand, refs, collecting)
	case *model.BoolOp:
		for _, v := range e.Values {
			captureWalkExpr(v, refs, collecting)
		}
	case *model.Compare:
		captureWalkExpr(e.Left, refs, collecting)
		for _, comp := range e.Comparators {
			captureWalkExpr(comp, refs, collecting)
		}
	case *model.IfExpr:
		captureWalkExpr(e.Test, refs, collecting)
		captureWalkExpr(e.Body, refs, collecting)
		captureWalkExpr(e.OrElse, refs, collecting)
	case *model.Attribute:
		captureWalkExpr(e.Value, refs, collecting)
	case *model.Subscript:
		captureWalkExpr(e.Value, refs, collecting)
		captureWalkExpr(e.Slice, refs, collecting)
	case *model.Slice:
		captureWalkExpr(e.Lower, refs, collecting)
		captureWalkExpr(e.Upper, refs, collecting)
		captureWalkExpr(e.Step, refs, collecting)
	case *model.List:
		for _, elt := range e.Elts {
			captureWalkExpr(elt, refs, collecting)
		}
	case *model.Tuple:
		for _, elt := range e.Elts {
			captureWalkExpr(elt, refs, collecting)
		}
	case *model.Dict:
		for _, k := range e.Keys {
			captureWalkExpr(k, refs, collecting)
		}
		for _, v := range e.Values {
			captureWalkExpr(v, refs, collecting)
		}
	case *model.Set:
		for _, elt := range e.Elts {
			captureWalkExpr(elt, refs, collecting)
		}
	case *model.Starred:
		captureWalkExpr(e.Value, refs, collecting)
	case *model.Await:
		captureWalkExpr(e.Value, refs, collecting)
	case *model.NamedExpr:
		captureWalkExpr(e.Value, refs, collecting)
	case *model.Yield:
		if e.Value != nil {
			captureWalkExpr(e.Value, refs, collecting)
		}
	case *model.YieldFrom:
		captureWalkExpr(e.Value, refs, collecting)
	case *model.FStringLit:
		for _, part := range e.Parts {
			if part.Expr != nil {
				captureWalkExpr(part.Expr, refs, collecting)
			}
		}
	}
}

// containsYield checks if statements contain yield or yield from expressions.
// Does not descend into comprehensions or nested function/class definitions,
// as those create their own scope.
func containsYield(stmts []model.Stmt) bool {
	w := astWalker{
		exprMatch: func(expr model.Expr) bool {
			switch expr.(type) {
			case *model.Yield, *model.YieldFrom:
				return true
			}
			return false
		},
		enterComprehensionElts: false,
	}
	return w.walkStmts(stmts)
}

func containsYieldExpr(expr model.Expr) bool {
	w := astWalker{
		exprMatch: func(e model.Expr) bool {
			switch e.(type) {
			case *model.Yield, *model.YieldFrom:
				return true
			}
			return false
		},
		enterComprehensionElts: false,
	}
	return w.walkExpr(expr)
}
