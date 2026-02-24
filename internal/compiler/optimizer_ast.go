package compiler

import (
	"math"
	"strconv"
	"strings"

	"github.com/ATSOTECK/rage/internal/model"
)

// ==========================================
// Constant Folding - Evaluate constant expressions at compile time
// ==========================================

// FoldConstants attempts to evaluate constant expressions in the AST
// Returns the simplified expression, or the original if not foldable
func (o *Optimizer) FoldConstants(expr model.Expr) model.Expr {
	if !o.enabled {
		return expr
	}

	switch e := expr.(type) {
	case *model.BinaryOp:
		return o.foldBinaryOp(e)
	case *model.UnaryOp:
		return o.foldUnaryOp(e)
	case *model.BoolOp:
		return o.foldBoolOp(e)
	case *model.IfExpr:
		return o.foldIfExpr(e)
	default:
		return expr
	}
}

func (o *Optimizer) foldBinaryOp(e *model.BinaryOp) model.Expr {
	// First, recursively fold operands
	left := o.FoldConstants(e.Left)
	right := o.FoldConstants(e.Right)

	// Check if both operands are constants
	leftInt, leftIsInt := o.getIntValue(left)
	rightInt, rightIsInt := o.getIntValue(right)

	if leftIsInt && rightIsInt {
		result := o.evalIntBinaryOp(e.Op, leftInt, rightInt)
		if result != nil {
			return result
		}
	}

	leftFloat, leftIsFloat := o.getFloatValue(left)
	rightFloat, rightIsFloat := o.getFloatValue(right)

	if (leftIsFloat || leftIsInt) && (rightIsFloat || rightIsInt) {
		if !leftIsFloat {
			leftFloat = float64(leftInt)
		}
		if !rightIsFloat {
			rightFloat = float64(rightInt)
		}
		if leftIsFloat || rightIsFloat { // At least one was originally float
			result := o.evalFloatBinaryOp(e.Op, leftFloat, rightFloat)
			if result != nil {
				return result
			}
		}
	}

	leftStr, leftIsStr := o.getStringValue(left)
	rightStr, rightIsStr := o.getStringValue(right)

	if leftIsStr && rightIsStr && e.Op == model.TK_Plus {
		// String concatenation
		return &model.StringLit{Value: leftStr + rightStr}
	}

	if leftIsStr && rightIsInt && e.Op == model.TK_Star {
		// String repetition - use strings.Repeat for O(n) instead of O(n²)
		if rightInt >= 0 && rightInt <= 1000 { // Limit to prevent huge strings
			return &model.StringLit{Value: strings.Repeat(leftStr, int(rightInt))}
		}
	}

	// Return expression with folded operands
	if left != e.Left || right != e.Right {
		return &model.BinaryOp{Op: e.Op, Left: left, Right: right}
	}
	return e
}

func (o *Optimizer) evalIntBinaryOp(op model.TokenKind, left, right int64) model.Expr {
	var result int64
	switch op {
	case model.TK_Plus:
		result = left + right
	case model.TK_Minus:
		result = left - right
	case model.TK_Star:
		result = left * right
	case model.TK_DoubleSlash:
		if right == 0 {
			return nil // Division by zero, don't fold
		}
		result = left / right
	case model.TK_Percent:
		if right == 0 {
			return nil
		}
		result = left % right
	case model.TK_DoubleStar:
		if right >= 0 && right <= 63 { // Limit exponent
			f := math.Pow(float64(left), float64(right))
			if math.IsInf(f, 0) || math.IsNaN(f) || f > math.MaxInt64 || f < math.MinInt64 {
				return nil // Result doesn't fit in int64, don't fold
			}
			result = int64(f)
		} else {
			return nil
		}
	case model.TK_LShift:
		if right >= 0 && right <= 63 {
			result = left << uint(right)
		} else {
			return nil
		}
	case model.TK_RShift:
		if right >= 0 && right <= 63 {
			result = left >> uint(right)
		} else {
			return nil
		}
	case model.TK_Ampersand:
		result = left & right
	case model.TK_Pipe:
		result = left | right
	case model.TK_Caret:
		result = left ^ right
	default:
		return nil
	}
	return &model.IntLit{Value: strconv.FormatInt(result, 10)}
}

func (o *Optimizer) evalFloatBinaryOp(op model.TokenKind, left, right float64) model.Expr {
	var result float64
	switch op {
	case model.TK_Plus:
		result = left + right
	case model.TK_Minus:
		result = left - right
	case model.TK_Star:
		result = left * right
	case model.TK_Slash:
		if right == 0 {
			return nil
		}
		result = left / right
	case model.TK_DoubleStar:
		result = math.Pow(left, right)
	default:
		return nil
	}
	return &model.FloatLit{Value: strconv.FormatFloat(result, 'g', -1, 64)}
}

func (o *Optimizer) foldUnaryOp(e *model.UnaryOp) model.Expr {
	operand := o.FoldConstants(e.Operand)

	if intVal, ok := o.getIntValue(operand); ok {
		switch e.Op {
		case model.TK_Minus:
			return &model.IntLit{Value: strconv.FormatInt(-intVal, 10)}
		case model.TK_Plus:
			return operand
		case model.TK_Tilde:
			return &model.IntLit{Value: strconv.FormatInt(^intVal, 10)}
		}
	}

	if floatVal, ok := o.getFloatValue(operand); ok {
		switch e.Op {
		case model.TK_Minus:
			return &model.FloatLit{Value: strconv.FormatFloat(-floatVal, 'g', -1, 64)}
		case model.TK_Plus:
			return operand
		}
	}

	if boolVal, ok := o.getBoolValue(operand); ok {
		if e.Op == model.TK_Not {
			return &model.BoolLit{Value: !boolVal}
		}
	}

	if operand != e.Operand {
		return &model.UnaryOp{Op: e.Op, Operand: operand}
	}
	return e
}

func (o *Optimizer) foldBoolOp(e *model.BoolOp) model.Expr {
	// Fold each value
	values := make([]model.Expr, len(e.Values))
	allConst := true
	for i, v := range e.Values {
		values[i] = o.FoldConstants(v)
		if _, ok := o.getBoolValue(values[i]); !ok {
			allConst = false
		}
	}

	if allConst {
		if e.Op == model.TK_And {
			for _, v := range values {
				if boolVal, _ := o.getBoolValue(v); !boolVal {
					return &model.BoolLit{Value: false}
				}
			}
			return &model.BoolLit{Value: true}
		} else { // Or
			for _, v := range values {
				if boolVal, _ := o.getBoolValue(v); boolVal {
					return &model.BoolLit{Value: true}
				}
			}
			return &model.BoolLit{Value: false}
		}
	}

	return &model.BoolOp{Op: e.Op, Values: values}
}

func (o *Optimizer) foldIfExpr(e *model.IfExpr) model.Expr {
	test := o.FoldConstants(e.Test)

	if boolVal, ok := o.getBoolValue(test); ok {
		if boolVal {
			return o.FoldConstants(e.Body)
		}
		return o.FoldConstants(e.OrElse)
	}

	return &model.IfExpr{
		Test:   test,
		Body:   o.FoldConstants(e.Body),
		OrElse: o.FoldConstants(e.OrElse),
	}
}

// Helper functions to extract constant values
func (o *Optimizer) getIntValue(expr model.Expr) (int64, bool) {
	if lit, ok := expr.(*model.IntLit); ok {
		val, err := strconv.ParseInt(lit.Value, 0, 64)
		if err == nil {
			return val, true
		}
	}
	return 0, false
}

func (o *Optimizer) getFloatValue(expr model.Expr) (float64, bool) {
	if lit, ok := expr.(*model.FloatLit); ok {
		val, err := strconv.ParseFloat(lit.Value, 64)
		if err == nil {
			return val, true
		}
	}
	return 0, false
}

func (o *Optimizer) getStringValue(expr model.Expr) (string, bool) {
	if lit, ok := expr.(*model.StringLit); ok {
		return lit.Value, true
	}
	return "", false
}

func (o *Optimizer) getBoolValue(expr model.Expr) (bool, bool) {
	if lit, ok := expr.(*model.BoolLit); ok {
		return lit.Value, true
	}
	return false, false
}

// isPureExpr checks if an expression is pure (has no side effects).
// Pure expressions include: literals, identifiers, and operators with pure operands.
// Impure expressions include: function calls, subscripts, attribute access (could have __getattr__).
func (o *Optimizer) isPureExpr(expr model.Expr) bool {
	switch e := expr.(type) {
	case *model.IntLit, *model.FloatLit, *model.StringLit, *model.BytesLit,
		*model.BoolLit, *model.NoneLit, *model.Identifier:
		return true
	case *model.BinaryOp:
		return o.isPureExpr(e.Left) && o.isPureExpr(e.Right)
	case *model.UnaryOp:
		return o.isPureExpr(e.Operand)
	case *model.Tuple:
		for _, elt := range e.Elts {
			if !o.isPureExpr(elt) {
				return false
			}
		}
		return true
	case *model.List:
		for _, elt := range e.Elts {
			if !o.isPureExpr(elt) {
				return false
			}
		}
		return true
	default:
		// Function calls, subscripts, attribute access, etc. are not pure
		return false
	}
}

// isIntegerExpr checks if an expression is known to produce an integer
func (o *Optimizer) isIntegerExpr(expr model.Expr) bool {
	switch e := expr.(type) {
	case *model.IntLit:
		return true
	case *model.BinaryOp:
		// Bit operations always produce ints (in Python semantics)
		switch e.Op {
		case model.TK_LShift, model.TK_RShift, model.TK_Ampersand, model.TK_Pipe, model.TK_Caret:
			return true
		case model.TK_DoubleSlash:
			// Floor division produces int if both operands are int
			return o.isIntegerExpr(e.Left) && o.isIntegerExpr(e.Right)
		case model.TK_Plus, model.TK_Minus, model.TK_Star, model.TK_Percent:
			// These produce int only if both operands are int
			return o.isIntegerExpr(e.Left) && o.isIntegerExpr(e.Right)
		}
	case *model.UnaryOp:
		switch e.Op {
		case model.TK_Tilde:
			return true // Bit inversion always produces int
		case model.TK_Minus, model.TK_Plus:
			return o.isIntegerExpr(e.Operand)
		}
	}
	return false
}

// ==========================================
// Strength Reduction - Replace expensive ops with cheaper ones
// ==========================================

// ReduceStrength applies strength reduction optimizations
func (o *Optimizer) ReduceStrength(expr model.Expr) model.Expr {
	if !o.enabled {
		return expr
	}

	switch e := expr.(type) {
	case *model.BinaryOp:
		left := o.ReduceStrength(e.Left)
		right := o.ReduceStrength(e.Right)

		// Check if left operand is known to be an integer (for safe shift conversion)
		leftIsInt := o.isIntegerExpr(left)
		rightIsInt := o.isIntegerExpr(right)

		// x * 2 -> x << 1 (only for integers!)
		if e.Op == model.TK_Star {
			if rightInt, ok := o.getIntValue(right); ok {
				switch rightInt {
				case 0:
					return &model.IntLit{Value: "0"}
				case 1:
					return left
				case 2:
					if leftIsInt {
						return &model.BinaryOp{Op: model.TK_LShift, Left: left, Right: &model.IntLit{Value: "1"}}
					}
				case 4:
					if leftIsInt {
						return &model.BinaryOp{Op: model.TK_LShift, Left: left, Right: &model.IntLit{Value: "2"}}
					}
				case 8:
					if leftIsInt {
						return &model.BinaryOp{Op: model.TK_LShift, Left: left, Right: &model.IntLit{Value: "3"}}
					}
				case 16:
					if leftIsInt {
						return &model.BinaryOp{Op: model.TK_LShift, Left: left, Right: &model.IntLit{Value: "4"}}
					}
				}
			}
			if leftInt, ok := o.getIntValue(left); ok {
				switch leftInt {
				case 0:
					return &model.IntLit{Value: "0"}
				case 1:
					return right
				case 2:
					if rightIsInt {
						return &model.BinaryOp{Op: model.TK_LShift, Left: right, Right: &model.IntLit{Value: "1"}}
					}
				}
			}
		}

		// x ** 2 -> x * x (only if x is pure, to avoid evaluating side effects twice)
		if e.Op == model.TK_DoubleStar {
			if rightInt, ok := o.getIntValue(right); ok {
				switch rightInt {
				case 0:
					return &model.IntLit{Value: "1"}
				case 1:
					return left
				case 2:
					// Only optimize if left is a pure expression (no side effects)
					if o.isPureExpr(left) {
						return &model.BinaryOp{Op: model.TK_Star, Left: left, Right: left}
					}
				}
			}
		}

		// x + 0, x - 0 -> x
		if e.Op == model.TK_Plus || e.Op == model.TK_Minus {
			if rightInt, ok := o.getIntValue(right); ok && rightInt == 0 {
				return left
			}
		}
		if e.Op == model.TK_Plus {
			if leftInt, ok := o.getIntValue(left); ok && leftInt == 0 {
				return right
			}
		}

		// x // 1 -> x, x % 1 -> 0
		if e.Op == model.TK_DoubleSlash {
			if rightInt, ok := o.getIntValue(right); ok && rightInt == 1 {
				return left
			}
		}
		if e.Op == model.TK_Percent {
			if rightInt, ok := o.getIntValue(right); ok && rightInt == 1 {
				return &model.IntLit{Value: "0"}
			}
		}

		if left != e.Left || right != e.Right {
			return &model.BinaryOp{Op: e.Op, Left: left, Right: right}
		}
		return e

	default:
		return expr
	}
}

// ==========================================
// Dead Code Elimination
// ==========================================

// EliminateDeadCode removes unreachable code from statement list
func (o *Optimizer) EliminateDeadCode(stmts []model.Stmt) []model.Stmt {
	if !o.enabled {
		return stmts
	}

	result := make([]model.Stmt, 0, len(stmts))
	for i, stmt := range stmts {
		result = append(result, stmt)

		// Check for unconditional control flow
		switch stmt.(type) {
		case *model.Return, *model.Raise:
			// Everything after return/raise in the same block is dead
			if i < len(stmts)-1 {
				// Truncate here
				return result
			}
		case *model.Break, *model.Continue:
			// Everything after break/continue is dead
			if i < len(stmts)-1 {
				return result
			}
		}
	}
	return result
}

// OptimizeIf optimizes if statements, removing dead branches
func (o *Optimizer) OptimizeIf(s *model.If) []model.Stmt {
	if !o.enabled {
		return []model.Stmt{s}
	}

	test := o.FoldConstants(s.Test)

	if boolVal, ok := o.getBoolValue(test); ok {
		if boolVal {
			// Always true - just execute the body
			return s.Body
		}
		// Always false - execute else branch
		if len(s.OrElse) > 0 {
			return s.OrElse
		}
		// No else - remove entirely
		return []model.Stmt{}
	}

	return []model.Stmt{s}
}

// ==========================================
// Loop-Invariant Code Motion (LICM)
// ==========================================

// OptimizeLoop performs loop-invariant code motion on for/while loops
// It hoists expressions that don't depend on loop variables out of the loop
func (o *Optimizer) OptimizeLoop(stmts []model.Stmt) []model.Stmt {
	if !o.enabled {
		return stmts
	}

	result := make([]model.Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *model.For:
			hoisted, newFor := o.hoistLoopInvariants(s)
			result = append(result, hoisted...)
			result = append(result, newFor)
		case *model.While:
			hoisted, newWhile := o.hoistWhileInvariants(s)
			result = append(result, hoisted...)
			result = append(result, newWhile)
		default:
			result = append(result, stmt)
		}
	}
	return result
}

// hoistLoopInvariants identifies and hoists loop-invariant code from for loops
func (o *Optimizer) hoistLoopInvariants(forLoop *model.For) ([]model.Stmt, *model.For) {
	// Get the set of variables modified in the loop
	modifiedVars := o.findModifiedVars(forLoop.Body)

	// Add the loop variable to modified vars
	if ident, ok := forLoop.Target.(*model.Identifier); ok {
		modifiedVars[ident.Name] = true
	}

	// Find invariant assignments in the loop body
	var hoisted []model.Stmt
	var newBody []model.Stmt

	for _, stmt := range forLoop.Body {
		if assign, ok := stmt.(*model.Assign); ok {
			if len(assign.Targets) == 1 {
				if ident, ok := assign.Targets[0].(*model.Identifier); ok {
					// Check if the RHS is loop-invariant
					if o.isLoopInvariant(assign.Value, modifiedVars) {
						// Check if this variable is only assigned once in the loop
						// and is not used before this assignment
						if !o.isUsedBeforeDefinition(ident.Name, forLoop.Body, stmt) {
							// Hoist this assignment
							hoisted = append(hoisted, stmt)
							continue
						}
					}
				}
			}
		}
		newBody = append(newBody, stmt)
	}

	newFor := &model.For{
		Target:   forLoop.Target,
		Iter:     forLoop.Iter,
		Body:     newBody,
		OrElse:   forLoop.OrElse,
		StartPos: forLoop.StartPos,
		EndPos:   forLoop.EndPos,
	}

	return hoisted, newFor
}

// hoistWhileInvariants identifies and hoists loop-invariant code from while loops
func (o *Optimizer) hoistWhileInvariants(whileLoop *model.While) ([]model.Stmt, *model.While) {
	// Get the set of variables modified in the loop
	modifiedVars := o.findModifiedVars(whileLoop.Body)

	// Find invariant assignments in the loop body
	var hoisted []model.Stmt
	var newBody []model.Stmt

	for _, stmt := range whileLoop.Body {
		if assign, ok := stmt.(*model.Assign); ok {
			if len(assign.Targets) == 1 {
				if ident, ok := assign.Targets[0].(*model.Identifier); ok {
					// Check if the RHS is loop-invariant
					if o.isLoopInvariant(assign.Value, modifiedVars) {
						// Check if this variable is only assigned once and not used in condition
						if !o.exprUsesVar(whileLoop.Test, ident.Name) &&
							!o.isUsedBeforeDefinition(ident.Name, whileLoop.Body, stmt) {
							// Hoist this assignment
							hoisted = append(hoisted, stmt)
							continue
						}
					}
				}
			}
		}
		newBody = append(newBody, stmt)
	}

	newWhile := &model.While{
		Test:     whileLoop.Test,
		Body:     newBody,
		OrElse:   whileLoop.OrElse,
		StartPos: whileLoop.StartPos,
		EndPos:   whileLoop.EndPos,
	}

	return hoisted, newWhile
}

// findModifiedVars returns a set of variable names that are assigned in the statements
func (o *Optimizer) findModifiedVars(stmts []model.Stmt) map[string]bool {
	modified := make(map[string]bool)
	for _, stmt := range stmts {
		o.collectModifiedVars(stmt, modified)
	}
	return modified
}

// collectModifiedVars recursively finds all modified variables in a statement
func (o *Optimizer) collectModifiedVars(stmt model.Stmt, modified map[string]bool) {
	switch s := stmt.(type) {
	case *model.Assign:
		for _, target := range s.Targets {
			o.collectAssignTargets(target, modified)
		}
	case *model.AugAssign:
		o.collectAssignTargets(s.Target, modified)
	case *model.AnnAssign:
		if s.Value != nil {
			o.collectAssignTargets(s.Target, modified)
		}
	case *model.For:
		o.collectAssignTargets(s.Target, modified)
		for _, bodyStmt := range s.Body {
			o.collectModifiedVars(bodyStmt, modified)
		}
	case *model.While:
		for _, bodyStmt := range s.Body {
			o.collectModifiedVars(bodyStmt, modified)
		}
	case *model.If:
		for _, bodyStmt := range s.Body {
			o.collectModifiedVars(bodyStmt, modified)
		}
		for _, elseStmt := range s.OrElse {
			o.collectModifiedVars(elseStmt, modified)
		}
	case *model.FunctionDef:
		// Function definitions don't modify outer scope (directly)
		// but the function name itself is defined
		modified[s.Name.Name] = true
	}
}

// collectAssignTargets extracts variable names from assignment targets
func (o *Optimizer) collectAssignTargets(target model.Expr, modified map[string]bool) {
	switch t := target.(type) {
	case *model.Identifier:
		modified[t.Name] = true
	case *model.Tuple:
		for _, elt := range t.Elts {
			o.collectAssignTargets(elt, modified)
		}
	case *model.List:
		for _, elt := range t.Elts {
			o.collectAssignTargets(elt, modified)
		}
	}
}

// isLoopInvariant checks if an expression doesn't depend on any modified variables
func (o *Optimizer) isLoopInvariant(expr model.Expr, modifiedVars map[string]bool) bool {
	switch e := expr.(type) {
	case *model.IntLit, *model.FloatLit, *model.StringLit,
		*model.BytesLit, *model.BoolLit, *model.NoneLit:
		return true
	case *model.Identifier:
		return !modifiedVars[e.Name]
	case *model.BinaryOp:
		return o.isLoopInvariant(e.Left, modifiedVars) &&
			o.isLoopInvariant(e.Right, modifiedVars)
	case *model.UnaryOp:
		return o.isLoopInvariant(e.Operand, modifiedVars)
	case *model.Attribute:
		// Attribute access is only invariant if the object is invariant
		// and we're not calling methods that might have side effects
		return o.isLoopInvariant(e.Value, modifiedVars)
	case *model.Subscript:
		return o.isLoopInvariant(e.Value, modifiedVars) &&
			o.isLoopInvariant(e.Slice, modifiedVars)
	case *model.Call:
		// Calls are generally not invariant (side effects)
		// Exception: some pure builtins like len()
		if ident, ok := e.Func.(*model.Identifier); ok {
			switch ident.Name {
			case "len", "abs", "min", "max", "int", "float", "str", "bool":
				// Pure functions - check if args are invariant
				for _, arg := range e.Args {
					if !o.isLoopInvariant(arg, modifiedVars) {
						return false
					}
				}
				return true
			}
		}
		return false
	case *model.List, *model.Tuple, *model.Dict, *model.Set:
		// Collections create new objects each time - not invariant
		return false
	default:
		return false
	}
}

// exprUsesVar checks if an expression uses a specific variable
func (o *Optimizer) exprUsesVar(expr model.Expr, varName string) bool {
	switch e := expr.(type) {
	case *model.Identifier:
		return e.Name == varName
	case *model.BinaryOp:
		return o.exprUsesVar(e.Left, varName) || o.exprUsesVar(e.Right, varName)
	case *model.UnaryOp:
		return o.exprUsesVar(e.Operand, varName)
	case *model.BoolOp:
		for _, val := range e.Values {
			if o.exprUsesVar(val, varName) {
				return true
			}
		}
		return false
	case *model.Compare:
		if o.exprUsesVar(e.Left, varName) {
			return true
		}
		for _, comp := range e.Comparators {
			if o.exprUsesVar(comp, varName) {
				return true
			}
		}
		return false
	case *model.Call:
		if o.exprUsesVar(e.Func, varName) {
			return true
		}
		for _, arg := range e.Args {
			if o.exprUsesVar(arg, varName) {
				return true
			}
		}
		return false
	case *model.Attribute:
		return o.exprUsesVar(e.Value, varName)
	case *model.Subscript:
		return o.exprUsesVar(e.Value, varName) || o.exprUsesVar(e.Slice, varName)
	default:
		return false
	}
}

// isUsedBeforeDefinition checks if a variable is used before a specific statement
func (o *Optimizer) isUsedBeforeDefinition(varName string, stmts []model.Stmt, defStmt model.Stmt) bool {
	for _, stmt := range stmts {
		if stmt == defStmt {
			return false // Reached the definition without finding a use
		}
		if o.stmtUsesVar(stmt, varName) {
			return true
		}
	}
	return false
}

// stmtUsesVar checks if a statement uses a specific variable
func (o *Optimizer) stmtUsesVar(stmt model.Stmt, varName string) bool {
	switch s := stmt.(type) {
	case *model.ExprStmt:
		return o.exprUsesVar(s.Value, varName)
	case *model.Assign:
		return o.exprUsesVar(s.Value, varName)
	case *model.AugAssign:
		return o.exprUsesVar(s.Target, varName) || o.exprUsesVar(s.Value, varName)
	case *model.If:
		return o.exprUsesVar(s.Test, varName)
	case *model.While:
		return o.exprUsesVar(s.Test, varName)
	case *model.For:
		return o.exprUsesVar(s.Iter, varName)
	case *model.Return:
		if s.Value != nil {
			return o.exprUsesVar(s.Value, varName)
		}
	}
	return false
}
