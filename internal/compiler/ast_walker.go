package compiler

import "github.com/ATSOTECK/rage/internal/model"

// astWalker performs a depth-first traversal of AST nodes, calling a predicate
// on each expression. It short-circuits and returns true as soon as the predicate
// matches. This eliminates duplication between different AST scanning functions
// (e.g., usesSuperOrClass and containsYield) that share the same traversal
// structure but differ only in what they look for and which subtrees they skip.
type astWalker struct {
	// exprMatch is the predicate called on each expression node. If it returns
	// true the walk short-circuits immediately.
	exprMatch func(model.Expr) bool

	// enterComprehensionElts controls whether the walker descends into the
	// element expressions of ListComp, SetComp, DictComp, and GeneratorExpr.
	// Comprehensions create their own scope, so yield detection skips them
	// while super/__class__ detection enters them.
	enterComprehensionElts bool
}

// walkStmts walks a slice of statements, returning true on first match.
func (w *astWalker) walkStmts(stmts []model.Stmt) bool {
	for _, stmt := range stmts {
		if w.walkStmt(stmt) {
			return true
		}
	}
	return false
}

// walkStmt walks a single statement, returning true on first match.
func (w *astWalker) walkStmt(stmt model.Stmt) bool {
	switch s := stmt.(type) {
	case *model.ExprStmt:
		return w.walkExpr(s.Value)
	case *model.Return:
		if s.Value != nil {
			return w.walkExpr(s.Value)
		}
	case *model.Assign:
		for _, target := range s.Targets {
			if w.walkExpr(target) {
				return true
			}
		}
		return w.walkExpr(s.Value)
	case *model.AugAssign:
		return w.walkExpr(s.Target) || w.walkExpr(s.Value)
	case *model.AnnAssign:
		if s.Value != nil && w.walkExpr(s.Value) {
			return true
		}
	case *model.If:
		if w.walkExpr(s.Test) {
			return true
		}
		if w.walkStmts(s.Body) || w.walkStmts(s.OrElse) {
			return true
		}
	case *model.While:
		if w.walkExpr(s.Test) {
			return true
		}
		if w.walkStmts(s.Body) || w.walkStmts(s.OrElse) {
			return true
		}
	case *model.For:
		if w.walkExpr(s.Iter) {
			return true
		}
		if w.walkStmts(s.Body) || w.walkStmts(s.OrElse) {
			return true
		}
	case *model.Try:
		if w.walkStmts(s.Body) || w.walkStmts(s.OrElse) || w.walkStmts(s.FinalBody) {
			return true
		}
		for _, handler := range s.Handlers {
			if w.walkStmts(handler.Body) {
				return true
			}
		}
	case *model.With:
		for _, item := range s.Items {
			if w.walkExpr(item.ContextExpr) {
				return true
			}
		}
		return w.walkStmts(s.Body)
	case *model.Raise:
		if s.Exc != nil && w.walkExpr(s.Exc) {
			return true
		}
		if s.Cause != nil && w.walkExpr(s.Cause) {
			return true
		}
	case *model.Assert:
		if w.walkExpr(s.Test) {
			return true
		}
		if s.Msg != nil && w.walkExpr(s.Msg) {
			return true
		}
	case *model.Match:
		if w.walkExpr(s.Subject) {
			return true
		}
		for _, c := range s.Cases {
			if w.walkStmts(c.Body) {
				return true
			}
		}
		// Note: Don't descend into nested FunctionDef or ClassDef.
	}
	return false
}

// walkExpr walks a single expression, returning true on first match.
func (w *astWalker) walkExpr(expr model.Expr) bool {
	if expr == nil {
		return false
	}
	if w.exprMatch(expr) {
		return true
	}
	switch e := expr.(type) {
	case *model.Call:
		if w.walkExpr(e.Func) {
			return true
		}
		for _, arg := range e.Args {
			if w.walkExpr(arg) {
				return true
			}
		}
		for _, kw := range e.Keywords {
			if w.walkExpr(kw.Value) {
				return true
			}
		}
	case *model.Attribute:
		return w.walkExpr(e.Value)
	case *model.Subscript:
		if w.walkExpr(e.Value) {
			return true
		}
		return w.walkExpr(e.Slice)
	case *model.Slice:
		return w.walkExpr(e.Lower) || w.walkExpr(e.Upper) || w.walkExpr(e.Step)
	case *model.BinaryOp:
		return w.walkExpr(e.Left) || w.walkExpr(e.Right)
	case *model.UnaryOp:
		return w.walkExpr(e.Operand)
	case *model.BoolOp:
		for _, v := range e.Values {
			if w.walkExpr(v) {
				return true
			}
		}
	case *model.Compare:
		if w.walkExpr(e.Left) {
			return true
		}
		for _, comp := range e.Comparators {
			if w.walkExpr(comp) {
				return true
			}
		}
	case *model.IfExpr:
		return w.walkExpr(e.Test) || w.walkExpr(e.Body) || w.walkExpr(e.OrElse)
	case *model.List:
		for _, elt := range e.Elts {
			if w.walkExpr(elt) {
				return true
			}
		}
	case *model.Tuple:
		for _, elt := range e.Elts {
			if w.walkExpr(elt) {
				return true
			}
		}
	case *model.Dict:
		for _, k := range e.Keys {
			if w.walkExpr(k) {
				return true
			}
		}
		for _, v := range e.Values {
			if w.walkExpr(v) {
				return true
			}
		}
	case *model.Set:
		for _, elt := range e.Elts {
			if w.walkExpr(elt) {
				return true
			}
		}
	case *model.ListComp:
		if w.enterComprehensionElts {
			if w.walkExpr(e.Elt) {
				return true
			}
		}
	case *model.DictComp:
		if w.enterComprehensionElts {
			if w.walkExpr(e.Key) || w.walkExpr(e.Value) {
				return true
			}
		}
	case *model.SetComp:
		if w.enterComprehensionElts {
			if w.walkExpr(e.Elt) {
				return true
			}
		}
	case *model.GeneratorExpr:
		if w.enterComprehensionElts {
			if w.walkExpr(e.Elt) {
				return true
			}
		}
	case *model.Starred:
		return w.walkExpr(e.Value)
	case *model.Await:
		return w.walkExpr(e.Value)
	case *model.NamedExpr:
		return w.walkExpr(e.Value)
	}
	return false
}
