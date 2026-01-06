package compiler

import (
	"math"
	"strconv"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// Optimizer performs various compile-time optimizations
type Optimizer struct {
	enabled bool
}

// NewOptimizer creates a new optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{enabled: true}
}

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
		// String repetition
		if rightInt >= 0 && rightInt <= 1000 { // Limit to prevent huge strings
			result := ""
			for i := int64(0); i < rightInt; i++ {
				result += leftStr
			}
			return &model.StringLit{Value: result}
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
			result = int64(math.Pow(float64(left), float64(right)))
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

		// x ** 2 -> x * x
		if e.Op == model.TK_DoubleStar {
			if rightInt, ok := o.getIntValue(right); ok {
				switch rightInt {
				case 0:
					return &model.IntLit{Value: "1"}
				case 1:
					return left
				case 2:
					return &model.BinaryOp{Op: model.TK_Star, Left: left, Right: left}
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

// ==========================================
// Peephole Optimizer - Post-process bytecode
// ==========================================

// PeepholeOptimize optimizes the generated bytecode
func (o *Optimizer) PeepholeOptimize(code *runtime.CodeObject) {
	if !o.enabled || len(code.Code) == 0 {
		return
	}

	// Parse bytecode into instructions
	instrs := o.parseInstructions(code)

	// Apply peephole patterns
	changed := true
	for changed {
		changed = false

		// Remove redundant LOAD/POP pairs
		changed = o.removeLoadPop(instrs) || changed

		// Remove DUP/POP pairs
		changed = o.removeDupPop(instrs) || changed

		// Optimize jumps
		changed = o.optimizeJumps(instrs, code) || changed

		// Convert LOAD_FAST to specialized versions
		changed = o.specializeLoadFast(instrs) || changed

		// Convert STORE_FAST to specialized versions
		changed = o.specializeStoreFast(instrs) || changed

		// Convert constant loads to specialized versions
		changed = o.specializeLoadConst(instrs, code) || changed

		// Detect and convert increment patterns
		changed = o.detectIncrementPattern(instrs, code) || changed

		// Detect and convert decrement patterns
		changed = o.detectDecrementPattern(instrs, code) || changed

		// Detect and convert negate-in-place patterns (sign = -sign)
		changed = o.detectNegatePattern(instrs, code) || changed

		// Detect and convert add-const patterns (x = x + const)
		changed = o.detectAddConstPattern(instrs, code) || changed

		// Detect LOAD_FAST LOAD_FAST superinstruction
		changed = o.detectLoadFastLoadFast(instrs) || changed

		// Detect LOAD_FAST LOAD_CONST superinstruction
		changed = o.detectLoadFastLoadConst(instrs, code) || changed

		// Detect LOAD_CONST LOAD_FAST superinstruction
		changed = o.detectLoadConstLoadFast(instrs, code) || changed

		// Detect STORE_FAST LOAD_FAST superinstruction
		changed = o.detectStoreFastLoadFast(instrs) || changed

		// Optimize empty collection building
		changed = o.optimizeEmptyCollections(instrs) || changed

		// Optimize compare+jump sequences
		changed = o.optimizeCompareJump(instrs) || changed

		// Detect compare-local-jump fusion (for while loop conditions)
		changed = o.detectCompareLtLocalJump(instrs, code) || changed

		// Store-load elimination (STORE x; LOAD x -> DUP; STORE x)
		changed = o.eliminateStoreLoad(instrs) || changed

		// Jump threading
		changed = o.threadJumps(instrs) || changed

		// Optimize len() calls on known types
		changed = o.optimizeLenCalls(instrs, code) || changed

		// Specialize binary operations for integers
		changed = o.SpecializeBinaryOps(instrs, code) || changed
	}

	// Rebuild bytecode from instructions
	o.rebuildBytecode(code, instrs)
}

type instruction struct {
	op             runtime.Opcode
	arg            int
	removed        bool
	originalHadArg bool // Track if original instruction had an argument (for offset calculation)
}

func (o *Optimizer) parseInstructions(code *runtime.CodeObject) []*instruction {
	var instrs []*instruction
	offset := 0
	for offset < len(code.Code) {
		op := runtime.Opcode(code.Code[offset])
		hadArg := op.HasArg() && offset+2 < len(code.Code)
		instr := &instruction{op: op, arg: -1, originalHadArg: hadArg}
		if hadArg {
			instr.arg = int(code.Code[offset+1]) | int(code.Code[offset+2])<<8
			offset += 3
		} else {
			offset++
		}
		instrs = append(instrs, instr)
	}
	return instrs
}

func (o *Optimizer) removeLoadPop(instrs []*instruction) bool {
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}
		// Pattern: LOAD_FAST/LOAD_CONST followed by POP
		if (instrs[i].op == runtime.OpLoadFast || instrs[i].op == runtime.OpLoadConst ||
			instrs[i].op == runtime.OpLoadFast0 || instrs[i].op == runtime.OpLoadFast1 ||
			instrs[i].op == runtime.OpLoadFast2 || instrs[i].op == runtime.OpLoadFast3 ||
			instrs[i].op == runtime.OpLoadNone || instrs[i].op == runtime.OpLoadTrue ||
			instrs[i].op == runtime.OpLoadFalse) &&
			instrs[i+1].op == runtime.OpPop {
			instrs[i].removed = true
			instrs[i+1].removed = true
			changed = true
		}
	}
	return changed
}

func (o *Optimizer) removeDupPop(instrs []*instruction) bool {
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}
		if instrs[i].op == runtime.OpDup && instrs[i+1].op == runtime.OpPop {
			instrs[i].removed = true
			instrs[i+1].removed = true
			changed = true
		}
	}
	return changed
}

func (o *Optimizer) optimizeJumps(instrs []*instruction, code *runtime.CodeObject) bool {
	// This is complex because we need to handle jump targets
	// For now, just handle simple cases

	// First, build a set of instruction indices that are jump targets
	// We need to be careful not to optimize patterns where the first instruction
	// is a jump target (because control flow may come from elsewhere with different stack)
	jumpTargets := make(map[int]bool)
	offset := 0
	for i, instr := range instrs {
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
		_ = i // We'll use the offset to find target instruction index below
	}

	// Map offsets to instruction indices
	offsetToIndex := make(map[int]int)
	offset = 0
	for i, instr := range instrs {
		offsetToIndex[offset] = i
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
	}

	// Mark jump targets
	for _, instr := range instrs {
		if isJumpOp(instr.op) && instr.arg >= 0 {
			if targetIdx, ok := offsetToIndex[instr.arg]; ok {
				jumpTargets[targetIdx] = true
			}
		}
	}

	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed {
			continue
		}

		// Skip if this instruction is a jump target - can't safely optimize
		// because control flow may come from elsewhere with different stack state
		if jumpTargets[i] {
			continue
		}

		// LOAD_CONST True; POP_JUMP_IF_FALSE -> remove both (never jumps)
		if instrs[i].op == runtime.OpLoadTrue || instrs[i].op == runtime.OpLoadConst {
			if instrs[i].op == runtime.OpLoadConst {
				// Check if it's loading True
				if instrs[i].arg < len(code.Constants) {
					if b, ok := code.Constants[instrs[i].arg].(bool); !ok || !b {
						continue
					}
				}
			}
			if i+1 < len(instrs) && instrs[i+1].op == runtime.OpPopJumpIfFalse {
				instrs[i].removed = true
				instrs[i+1].removed = true
				changed = true
			}
		}
		// LOAD_CONST False; POP_JUMP_IF_FALSE -> JUMP
		if instrs[i].op == runtime.OpLoadFalse || instrs[i].op == runtime.OpLoadConst {
			if instrs[i].op == runtime.OpLoadConst {
				if instrs[i].arg < len(code.Constants) {
					if b, ok := code.Constants[instrs[i].arg].(bool); !ok || b {
						continue
					}
				}
			}
			if i+1 < len(instrs) && instrs[i+1].op == runtime.OpPopJumpIfFalse {
				instrs[i].removed = true
				instrs[i+1].op = runtime.OpJump // Convert to unconditional jump
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeLoadFast(instrs []*instruction) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		if instr.op == runtime.OpLoadFast {
			switch instr.arg {
			case 0:
				instr.op = runtime.OpLoadFast0
				instr.arg = -1
				changed = true
			case 1:
				instr.op = runtime.OpLoadFast1
				instr.arg = -1
				changed = true
			case 2:
				instr.op = runtime.OpLoadFast2
				instr.arg = -1
				changed = true
			case 3:
				instr.op = runtime.OpLoadFast3
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeStoreFast(instrs []*instruction) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		if instr.op == runtime.OpStoreFast {
			switch instr.arg {
			case 0:
				instr.op = runtime.OpStoreFast0
				instr.arg = -1
				changed = true
			case 1:
				instr.op = runtime.OpStoreFast1
				instr.arg = -1
				changed = true
			case 2:
				instr.op = runtime.OpStoreFast2
				instr.arg = -1
				changed = true
			case 3:
				instr.op = runtime.OpStoreFast3
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeLoadConst(instrs []*instruction, code *runtime.CodeObject) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed || instr.op != runtime.OpLoadConst {
			continue
		}
		if instr.arg >= len(code.Constants) {
			continue
		}
		c := code.Constants[instr.arg]
		switch v := c.(type) {
		case nil:
			instr.op = runtime.OpLoadNone
			instr.arg = -1
			changed = true
		case bool:
			if v {
				instr.op = runtime.OpLoadTrue
			} else {
				instr.op = runtime.OpLoadFalse
			}
			instr.arg = -1
			changed = true
		case int64:
			if v == 0 {
				instr.op = runtime.OpLoadZero
				instr.arg = -1
				changed = true
			} else if v == 1 {
				instr.op = runtime.OpLoadOne
				instr.arg = -1
				changed = true
			}
		case int:
			if v == 0 {
				instr.op = runtime.OpLoadZero
				instr.arg = -1
				changed = true
			} else if v == 1 {
				instr.op = runtime.OpLoadOne
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) detectIncrementPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST 1, BINARY_ADD, STORE_FAST x -> INCREMENT_FAST x
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		var localIdx int
		switch instrs[i].op {
		case runtime.OpLoadFast:
			localIdx = instrs[i].arg
		case runtime.OpLoadFast0:
			localIdx = 0
		case runtime.OpLoadFast1:
			localIdx = 1
		case runtime.OpLoadFast2:
			localIdx = 2
		case runtime.OpLoadFast3:
			localIdx = 3
		default:
			continue
		}

		// Check for LOAD_CONST 1 or LOAD_ONE
		isOne := false
		if instrs[i+1].op == runtime.OpLoadOne {
			isOne = true
		} else if instrs[i+1].op == runtime.OpLoadConst {
			if instrs[i+1].arg < len(code.Constants) {
				switch v := code.Constants[instrs[i+1].arg].(type) {
				case int64:
					isOne = v == 1
				case int:
					isOne = v == 1
				}
			}
		}
		if !isOne {
			continue
		}

		// Check for BINARY_ADD
		if instrs[i+2].op != runtime.OpBinaryAdd && instrs[i+2].op != runtime.OpBinaryAddInt {
			continue
		}

		// Check for STORE_FAST to same variable
		var storeIdx int
		switch instrs[i+3].op {
		case runtime.OpStoreFast:
			storeIdx = instrs[i+3].arg
		case runtime.OpStoreFast0:
			storeIdx = 0
		case runtime.OpStoreFast1:
			storeIdx = 1
		case runtime.OpStoreFast2:
			storeIdx = 2
		case runtime.OpStoreFast3:
			storeIdx = 3
		default:
			continue
		}

		if localIdx != storeIdx {
			continue
		}

		// Convert to INCREMENT_FAST
		instrs[i].op = runtime.OpIncrementFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectDecrementPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST 1, BINARY_SUB, STORE_FAST x -> DECREMENT_FAST x
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for LOAD_CONST 1 or LOAD_ONE
		if !o.isLoadOne(instrs[i+1], code) {
			continue
		}

		// Check for BINARY_SUBTRACT
		if instrs[i+2].op != runtime.OpBinarySubtract && instrs[i+2].op != runtime.OpBinarySubtractInt {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+3])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Convert to DECREMENT_FAST
		instrs[i].op = runtime.OpDecrementFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectNegatePattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, UNARY_NEGATIVE, STORE_FAST x -> NEGATE_FAST x
	changed := false
	for i := 0; i < len(instrs)-2; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for UNARY_NEGATIVE
		if instrs[i+1].op != runtime.OpUnaryNegative {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+2])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Convert to NEGATE_FAST
		instrs[i].op = runtime.OpNegateFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectAddConstPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST c, BINARY_ADD, STORE_FAST x -> ADD_CONST_FAST (x, c)
	// Skip if c == 1 (handled by INCREMENT_FAST)
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for LOAD_CONST (but not 1, which is handled by INCREMENT_FAST)
		if instrs[i+1].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i+1].arg
		// Skip if it's loading 1 (use INCREMENT_FAST instead)
		if o.isLoadOne(instrs[i+1], code) {
			continue
		}
		// Only optimize for integer constants
		if constIdx >= len(code.Constants) {
			continue
		}
		switch code.Constants[constIdx].(type) {
		case int64, int:
			// OK, it's an integer
		default:
			continue
		}

		// Check for BINARY_ADD
		if instrs[i+2].op != runtime.OpBinaryAdd && instrs[i+2].op != runtime.OpBinaryAddInt {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+3])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Only encode if indices fit in packed format (8 bits each)
		if localIdx > 255 || constIdx > 255 {
			continue
		}

		// Convert to ADD_CONST_FAST
		instrs[i].op = runtime.OpAddConstFast
		instrs[i].arg = localIdx | (constIdx << 8)
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

// Helper to get LOAD_FAST index from instruction (handles specialized versions)
func (o *Optimizer) getLoadFastIndex(instr *instruction) int {
	if instr.removed {
		return -1
	}
	switch instr.op {
	case runtime.OpLoadFast:
		return instr.arg
	case runtime.OpLoadFast0:
		return 0
	case runtime.OpLoadFast1:
		return 1
	case runtime.OpLoadFast2:
		return 2
	case runtime.OpLoadFast3:
		return 3
	}
	return -1
}

// Helper to get STORE_FAST index from instruction (handles specialized versions)
func (o *Optimizer) getStoreFastIndex(instr *instruction) int {
	if instr.removed {
		return -1
	}
	switch instr.op {
	case runtime.OpStoreFast:
		return instr.arg
	case runtime.OpStoreFast0:
		return 0
	case runtime.OpStoreFast1:
		return 1
	case runtime.OpStoreFast2:
		return 2
	case runtime.OpStoreFast3:
		return 3
	}
	return -1
}

// Helper to check if instruction loads the constant 1
func (o *Optimizer) isLoadOne(instr *instruction, code *runtime.CodeObject) bool {
	if instr.removed {
		return false
	}
	if instr.op == runtime.OpLoadOne {
		return true
	}
	if instr.op == runtime.OpLoadConst && instr.arg < len(code.Constants) {
		switch v := code.Constants[instr.arg].(type) {
		case int64:
			return v == 1
		case int:
			return v == 1
		}
	}
	return false
}

// Helper to check if instruction is any LOAD_FAST variant
func (o *Optimizer) isLoadFastFamily(instr *instruction) bool {
	if instr.removed {
		return false
	}
	switch instr.op {
	case runtime.OpLoadFast, runtime.OpLoadFast0, runtime.OpLoadFast1,
		runtime.OpLoadFast2, runtime.OpLoadFast3:
		return true
	}
	return false
}

// Helper to check if instruction is any STORE_FAST variant
func (o *Optimizer) isStoreFastFamily(instr *instruction) bool {
	if instr.removed {
		return false
	}
	switch instr.op {
	case runtime.OpStoreFast, runtime.OpStoreFast0, runtime.OpStoreFast1,
		runtime.OpStoreFast2, runtime.OpStoreFast3:
		return true
	}
	return false
}

func (o *Optimizer) detectLoadFastLoadFast(instrs []*instruction) bool {
	// Pattern: LOAD_FAST x, LOAD_FAST y -> LOAD_FAST_LOAD_FAST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		idx1 := o.getLoadFastIndex(instrs[i])
		idx2 := o.getLoadFastIndex(instrs[i+1])

		if idx1 < 0 || idx2 < 0 {
			continue
		}

		// Both indices must fit in 8 bits each (we pack into 16-bit arg)
		if idx1 > 255 || idx2 > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = first, high byte = second
		instrs[i].op = runtime.OpLoadFastLoadFast
		instrs[i].arg = (idx2 << 8) | idx1
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectLoadFastLoadConst(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST y -> LOAD_FAST_LOAD_CONST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		fastIdx := o.getLoadFastIndex(instrs[i])
		if fastIdx < 0 || fastIdx > 255 {
			continue
		}

		// Check for LOAD_CONST (but not specialized versions - those are already optimal)
		if instrs[i+1].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i+1].arg
		if constIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = local, high byte = const
		instrs[i].op = runtime.OpLoadFastLoadConst
		instrs[i].arg = (constIdx << 8) | fastIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectLoadConstLoadFast(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_CONST x, LOAD_FAST y -> LOAD_CONST_LOAD_FAST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		// Check for LOAD_CONST (but not specialized versions)
		if instrs[i].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i].arg
		if constIdx > 255 {
			continue
		}

		fastIdx := o.getLoadFastIndex(instrs[i+1])
		if fastIdx < 0 || fastIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: high byte = const, low byte = local
		instrs[i].op = runtime.OpLoadConstLoadFast
		instrs[i].arg = (constIdx << 8) | fastIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectStoreFastLoadFast(instrs []*instruction) bool {
	// Pattern: STORE_FAST x, LOAD_FAST y -> STORE_FAST_LOAD_FAST (packed)
	// This is common for chained assignments and expression statements
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		storeIdx := o.getStoreFastIndex(instrs[i])
		if storeIdx < 0 || storeIdx > 255 {
			continue
		}

		loadIdx := o.getLoadFastIndex(instrs[i+1])
		if loadIdx < 0 || loadIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = store, high byte = load
		instrs[i].op = runtime.OpStoreFastLoadFast
		instrs[i].arg = (loadIdx << 8) | storeIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) optimizeEmptyCollections(instrs []*instruction) bool {
	// Pattern: BUILD_LIST 0 -> LOAD_EMPTY_LIST
	// Pattern: BUILD_TUPLE 0 -> LOAD_EMPTY_TUPLE
	// Pattern: BUILD_MAP 0 -> LOAD_EMPTY_DICT
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		switch instr.op {
		case runtime.OpBuildList:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyList
				instr.arg = -1
				changed = true
			}
		case runtime.OpBuildTuple:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyTuple
				instr.arg = -1
				changed = true
			}
		case runtime.OpBuildMap:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyDict
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) optimizeCompareJump(instrs []*instruction) bool {
	// Pattern: COMPARE_xx, POP_JUMP_IF_FALSE -> COMPARE_xx_JUMP
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		// Check if this is a compare followed by POP_JUMP_IF_FALSE
		if instrs[i+1].op != runtime.OpPopJumpIfFalse {
			continue
		}

		var newOp runtime.Opcode
		switch instrs[i].op {
		case runtime.OpCompareLt:
			newOp = runtime.OpCompareLtJump
		case runtime.OpCompareLe:
			newOp = runtime.OpCompareLeJump
		case runtime.OpCompareGt:
			newOp = runtime.OpCompareGtJump
		case runtime.OpCompareGe:
			newOp = runtime.OpCompareGeJump
		case runtime.OpCompareEq:
			newOp = runtime.OpCompareEqJump
		case runtime.OpCompareNe:
			newOp = runtime.OpCompareNeJump
		default:
			continue
		}

		// Convert to combined compare+jump
		instrs[i].op = newOp
		instrs[i].arg = instrs[i+1].arg // Take the jump target
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) eliminateStoreLoad(instrs []*instruction) bool {
	// Pattern: STORE_FAST x, LOAD_FAST x -> DUP, STORE_FAST x
	// This keeps the value on the stack and avoids the re-load
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		storeIdx := o.getStoreFastIndex(instrs[i])
		if storeIdx < 0 {
			continue
		}

		loadIdx := o.getLoadFastIndex(instrs[i+1])
		if loadIdx < 0 {
			continue
		}

		// Only optimize if storing and loading the same variable
		if storeIdx != loadIdx {
			continue
		}

		// Convert to DUP + STORE
		// Insert DUP before STORE, remove LOAD
		instrs[i+1].op = instrs[i].op   // Move store to second position
		instrs[i+1].arg = instrs[i].arg // Copy the argument
		instrs[i].op = runtime.OpDup
		instrs[i].arg = -1
		changed = true
	}
	return changed
}

func (o *Optimizer) threadJumps(instrs []*instruction) bool {
	// Build offset map for finding targets
	offsetToIdx := make(map[int]int)
	offset := 0
	for i, instr := range instrs {
		offsetToIdx[offset] = i
		if instr.op.HasArg() || instr.arg >= 0 {
			offset += 3
		} else {
			offset++
		}
	}

	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}

		// Only process unconditional jumps
		if instr.op != runtime.OpJump {
			continue
		}

		// Find the target instruction
		targetIdx, ok := offsetToIdx[instr.arg]
		if !ok {
			continue
		}

		// If target is also a jump, thread through it
		if targetIdx < len(instrs) && !instrs[targetIdx].removed {
			target := instrs[targetIdx]
			if target.op == runtime.OpJump {
				// Thread the jump
				instr.arg = target.arg
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) optimizeLenCalls(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_GLOBAL "len", LOAD_xxx, CALL 1 -> LEN_GENERIC
	// This avoids the function call overhead for len()
	changed := false
	for i := 0; i < len(instrs)-2; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed {
			continue
		}

		// Check for LOAD_GLOBAL with "len"
		if instrs[i].op != runtime.OpLoadGlobal {
			continue
		}

		// Verify it's loading "len"
		nameIdx := instrs[i].arg
		if nameIdx >= len(code.Names) || code.Names[nameIdx] != "len" {
			continue
		}

		// Check for CALL with 1 argument
		if instrs[i+2].op != runtime.OpCall || instrs[i+2].arg != 1 {
			continue
		}

		// Replace with inline len
		// Remove LOAD_GLOBAL len
		instrs[i].removed = true

		// Replace CALL with LEN_GENERIC (operates on whatever is on stack)
		instrs[i+2].op = runtime.OpLenGeneric
		instrs[i+2].arg = -1
		changed = true
	}
	return changed
}

// ==========================================
// Binary Operation Specialization
// ==========================================

// Helper to check if an instruction loads an integer constant
func (o *Optimizer) isIntegerLoad(instr *instruction, code *runtime.CodeObject) bool {
	if instr.removed {
		return false
	}
	switch instr.op {
	case runtime.OpLoadZero, runtime.OpLoadOne:
		return true
	case runtime.OpLoadConst:
		if instr.arg < len(code.Constants) {
			switch code.Constants[instr.arg].(type) {
			case int64, int:
				return true
			}
		}
	}
	return false
}

// Helper to check if instruction is likely loading an integer
// (heuristic: from range iteration, integer constant, or known int operations)
func (o *Optimizer) isLikelyInteger(instr *instruction, code *runtime.CodeObject) bool {
	if instr.removed {
		return false
	}
	// Direct integer loads
	if o.isIntegerLoad(instr, code) {
		return true
	}
	// LOAD_FAST after FOR_ITER (range loop variable) is likely int
	// This is a heuristic - we can't always be sure
	return false
}

// SpecializeBinaryOps converts generic binary ops to specialized int versions
// when we can prove or heuristically determine operands are integers
func (o *Optimizer) SpecializeBinaryOps(instrs []*instruction, code *runtime.CodeObject) bool {
	changed := false
	for i := 0; i < len(instrs); i++ {
		if instrs[i].removed {
			continue
		}

		// Look for patterns where we load two integer constants and operate
		// Pattern: LOAD_CONST int, LOAD_CONST int, BINARY_OP
		if i >= 2 {
			op1 := instrs[i-2]
			op2 := instrs[i-1]
			if o.isIntegerLoad(op1, code) && o.isIntegerLoad(op2, code) {
				switch instrs[i].op {
				case runtime.OpBinaryAdd:
					instrs[i].op = runtime.OpBinaryAddInt
					changed = true
				case runtime.OpBinarySubtract:
					instrs[i].op = runtime.OpBinarySubtractInt
					changed = true
				case runtime.OpBinaryMultiply:
					instrs[i].op = runtime.OpBinaryMultiplyInt
					changed = true
				case runtime.OpCompareLt:
					instrs[i].op = runtime.OpCompareLtInt
					changed = true
				case runtime.OpCompareLe:
					instrs[i].op = runtime.OpCompareLeInt
					changed = true
				case runtime.OpCompareGt:
					instrs[i].op = runtime.OpCompareGtInt
					changed = true
				case runtime.OpCompareGe:
					instrs[i].op = runtime.OpCompareGeInt
					changed = true
				case runtime.OpCompareEq:
					instrs[i].op = runtime.OpCompareEqInt
					changed = true
				case runtime.OpCompareNe:
					instrs[i].op = runtime.OpCompareNeInt
					changed = true
				}
			}
		}

		// Pattern: LOAD_FAST, LOAD_CONST int, BINARY_OP (common in loops)
		// After FOR_ITER, the loop variable is often an int (from range)
		if i >= 2 && o.getLoadFastIndex(instrs[i-2]) >= 0 && o.isIntegerLoad(instrs[i-1], code) {
			switch instrs[i].op {
			case runtime.OpCompareLt:
				instrs[i].op = runtime.OpCompareLtInt
				changed = true
			case runtime.OpCompareLe:
				instrs[i].op = runtime.OpCompareLeInt
				changed = true
			case runtime.OpCompareGt:
				instrs[i].op = runtime.OpCompareGtInt
				changed = true
			case runtime.OpCompareGe:
				instrs[i].op = runtime.OpCompareGeInt
				changed = true
			case runtime.OpCompareEq:
				instrs[i].op = runtime.OpCompareEqInt
				changed = true
			case runtime.OpCompareNe:
				instrs[i].op = runtime.OpCompareNeInt
				changed = true
			case runtime.OpBinaryAdd:
				instrs[i].op = runtime.OpBinaryAddInt
				changed = true
			case runtime.OpBinarySubtract:
				instrs[i].op = runtime.OpBinarySubtractInt
				changed = true
			case runtime.OpBinaryMultiply:
				instrs[i].op = runtime.OpBinaryMultiplyInt
				changed = true
			}
		}

		// Pattern: LOAD_CONST int, LOAD_FAST, BINARY_OP
		if i >= 2 && o.isIntegerLoad(instrs[i-2], code) && o.getLoadFastIndex(instrs[i-1]) >= 0 {
			switch instrs[i].op {
			case runtime.OpCompareLt:
				instrs[i].op = runtime.OpCompareLtInt
				changed = true
			case runtime.OpCompareLe:
				instrs[i].op = runtime.OpCompareLeInt
				changed = true
			case runtime.OpCompareGt:
				instrs[i].op = runtime.OpCompareGtInt
				changed = true
			case runtime.OpCompareGe:
				instrs[i].op = runtime.OpCompareGeInt
				changed = true
			case runtime.OpCompareEq:
				instrs[i].op = runtime.OpCompareEqInt
				changed = true
			case runtime.OpCompareNe:
				instrs[i].op = runtime.OpCompareNeInt
				changed = true
			case runtime.OpBinaryAdd:
				instrs[i].op = runtime.OpBinaryAddInt
				changed = true
			case runtime.OpBinarySubtract:
				instrs[i].op = runtime.OpBinarySubtractInt
				changed = true
			case runtime.OpBinaryMultiply:
				instrs[i].op = runtime.OpBinaryMultiplyInt
				changed = true
			}
		}

		// Always specialize BINARY_DIVIDE to BINARY_DIVIDE_FLOAT
		// (true division in Python always returns float)
		if instrs[i].op == runtime.OpBinaryDivide {
			instrs[i].op = runtime.OpBinaryDivideFloat
			changed = true
		}

		// Specialize float addition when we know one operand is float
		// Pattern: ... BINARY_ADD after float operations
		if i >= 1 && instrs[i].op == runtime.OpBinaryAdd {
			// Check if previous result was from a division (which always produces float)
			for j := i - 1; j >= 0; j-- {
				if instrs[j].removed {
					continue
				}
				if instrs[j].op == runtime.OpBinaryDivideFloat || instrs[j].op == runtime.OpBinaryDivide {
					instrs[i].op = runtime.OpBinaryAddFloat
					changed = true
				}
				break
			}
		}
	}
	return changed
}

// Simple loop unrolling for very small loops
// This optimization unrolls loops with constant small iteration counts
func (o *Optimizer) unrollSmallLoops(instrs []*instruction, code *runtime.CodeObject) bool {
	// Look for patterns like:
	// LOAD_CONST small_int
	// GET_ITER
	// FOR_ITER exit
	// ... body ...
	// JUMP back
	//
	// For now, this is a placeholder - true loop unrolling requires
	// tracking loop boundaries and duplicating the body, which is complex
	// at the bytecode level. The real benefit would come from AST-level
	// unrolling before bytecode generation.
	return false
}

func (o *Optimizer) detectCompareLtLocalJump(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_FAST y, COMPARE_LT, POP_JUMP_IF_FALSE target
	// -> COMPARE_LT_LOCAL_JUMP (x, y, target)
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Get first local index
		local1 := o.getLoadFastIndex(instrs[i])
		if local1 < 0 || local1 > 255 {
			continue
		}

		// Get second local index
		local2 := o.getLoadFastIndex(instrs[i+1])
		if local2 < 0 || local2 > 255 {
			continue
		}

		// Check for COMPARE_LT
		if instrs[i+2].op != runtime.OpCompareLt && instrs[i+2].op != runtime.OpCompareLtInt {
			continue
		}

		// Check for POP_JUMP_IF_FALSE
		if instrs[i+3].op != runtime.OpPopJumpIfFalse {
			continue
		}

		jumpTarget := instrs[i+3].arg
		// Jump target must fit in remaining bits (16 bits after two 8-bit indices)
		if jumpTarget > 0xFFFF {
			continue
		}

		// Convert to COMPARE_LT_LOCAL_JUMP
		// Pack: bits 0-7 = local1, bits 8-15 = local2, bits 16-31 = jump target
		instrs[i].op = runtime.OpCompareLtLocalJump
		instrs[i].arg = local1 | (local2 << 8) | (jumpTarget << 16)
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) rebuildBytecode(code *runtime.CodeObject, instrs []*instruction) {
	// Calculate new size
	size := 0
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		if instr.arg >= 0 {
			size += 3
		} else {
			size++
		}
	}

	// Build offset map for jump target adjustment
	oldOffsets := make([]int, len(instrs))
	newOffsets := make([]int, len(instrs))
	oldOffset := 0
	newOffset := 0
	for i, instr := range instrs {
		oldOffsets[i] = oldOffset
		newOffsets[i] = newOffset
		// Use originalHadArg to correctly calculate old bytecode offsets
		if instr.originalHadArg {
			oldOffset += 3
		} else {
			oldOffset++
		}
		if !instr.removed {
			if instr.arg >= 0 {
				newOffset += 3
			} else {
				newOffset++
			}
		}
	}

	// Rebuild bytecode
	newCode := make([]byte, 0, size)
	for i, instr := range instrs {
		if instr.removed {
			continue
		}

		// Adjust jump targets
		arg := instr.arg
		if isJumpOp(instr.op) && arg >= 0 {
			// Find which instruction the old offset points to
			for j, oldOff := range oldOffsets {
				if oldOff == arg {
					arg = newOffsets[j]
					break
				}
			}
			// Handle jumps to removed instructions - find next valid
			for j := i + 1; j < len(instrs); j++ {
				if oldOffsets[j] == instr.arg && !instrs[j].removed {
					arg = newOffsets[j]
					break
				}
			}
		}

		newCode = append(newCode, byte(instr.op))
		if arg >= 0 {
			newCode = append(newCode, byte(arg), byte(arg>>8))
		}
	}

	code.Code = newCode
}

func isJumpOp(op runtime.Opcode) bool {
	switch op {
	case runtime.OpJump, runtime.OpJumpIfTrue, runtime.OpJumpIfFalse,
		runtime.OpPopJumpIfTrue, runtime.OpPopJumpIfFalse,
		runtime.OpJumpIfTrueOrPop, runtime.OpJumpIfFalseOrPop,
		runtime.OpForIter, runtime.OpSetupExcept, runtime.OpSetupFinally,
		runtime.OpSetupWith,
		// New compare+jump superinstructions
		runtime.OpCompareLtJump, runtime.OpCompareLeJump,
		runtime.OpCompareGtJump, runtime.OpCompareGeJump,
		runtime.OpCompareEqJump, runtime.OpCompareNeJump:
		return true
	}
	return false
}
