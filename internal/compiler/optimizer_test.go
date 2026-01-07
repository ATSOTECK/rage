package compiler

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Constant Folding Tests
// =============================================================================

func TestConstantFoldingIntegerAddition(t *testing.T) {
	opt := NewOptimizer()

	// Create AST for: 2 + 3
	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok, "Expected IntLit, got %T", result)
	assert.Equal(t, "5", intLit.Value)
}

func TestConstantFoldingIntegerSubtraction(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Minus,
		Left:  &model.IntLit{Value: "10"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "7", intLit.Value)
}

func TestConstantFoldingIntegerMultiplication(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Star,
		Left:  &model.IntLit{Value: "4"},
		Right: &model.IntLit{Value: "5"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "20", intLit.Value)
}

func TestConstantFoldingIntegerFloorDivision(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_DoubleSlash,
		Left:  &model.IntLit{Value: "10"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "3", intLit.Value)
}

func TestConstantFoldingDivisionByZeroNotFolded(t *testing.T) {
	opt := NewOptimizer()

	// Division by zero should not be folded (handled at runtime)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleSlash,
		Left:  &model.IntLit{Value: "10"},
		Right: &model.IntLit{Value: "0"},
	}

	result := opt.FoldConstants(expr)

	// Should return original binary op, not folded
	_, ok := result.(*model.BinaryOp)
	assert.True(t, ok, "Division by zero should not be folded")
}

func TestConstantFoldingIntegerModulo(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Percent,
		Left:  &model.IntLit{Value: "10"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "1", intLit.Value)
}

func TestConstantFoldingIntegerPower(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "10"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "1024", intLit.Value)
}

func TestConstantFoldingBitwiseAnd(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Ampersand,
		Left:  &model.IntLit{Value: "15"},
		Right: &model.IntLit{Value: "7"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "7", intLit.Value)
}

func TestConstantFoldingBitwiseOr(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Pipe,
		Left:  &model.IntLit{Value: "4"},
		Right: &model.IntLit{Value: "2"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "6", intLit.Value)
}

func TestConstantFoldingBitwiseXor(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Caret,
		Left:  &model.IntLit{Value: "5"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "6", intLit.Value)
}

func TestConstantFoldingLeftShift(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_LShift,
		Left:  &model.IntLit{Value: "1"},
		Right: &model.IntLit{Value: "4"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "16", intLit.Value)
}

func TestConstantFoldingRightShift(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_RShift,
		Left:  &model.IntLit{Value: "16"},
		Right: &model.IntLit{Value: "2"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "4", intLit.Value)
}

func TestConstantFoldingFloatAddition(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.FloatLit{Value: "1.5"},
		Right: &model.FloatLit{Value: "2.5"},
	}

	result := opt.FoldConstants(expr)

	floatLit, ok := result.(*model.FloatLit)
	require.True(t, ok)
	assert.Equal(t, "4", floatLit.Value)
}

func TestConstantFoldingFloatDivision(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Slash,
		Left:  &model.FloatLit{Value: "10.0"},
		Right: &model.FloatLit{Value: "4.0"},
	}

	result := opt.FoldConstants(expr)

	floatLit, ok := result.(*model.FloatLit)
	require.True(t, ok)
	assert.Equal(t, "2.5", floatLit.Value)
}

func TestConstantFoldingMixedIntFloat(t *testing.T) {
	opt := NewOptimizer()

	// 2 + 3.5 should produce float
	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.FloatLit{Value: "3.5"},
	}

	result := opt.FoldConstants(expr)

	floatLit, ok := result.(*model.FloatLit)
	require.True(t, ok)
	assert.Equal(t, "5.5", floatLit.Value)
}

func TestConstantFoldingStringConcatenation(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.StringLit{Value: "hello"},
		Right: &model.StringLit{Value: " world"},
	}

	result := opt.FoldConstants(expr)

	strLit, ok := result.(*model.StringLit)
	require.True(t, ok)
	assert.Equal(t, "hello world", strLit.Value)
}

func TestConstantFoldingStringRepetition(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Star,
		Left:  &model.StringLit{Value: "ab"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	strLit, ok := result.(*model.StringLit)
	require.True(t, ok)
	assert.Equal(t, "ababab", strLit.Value)
}

func TestConstantFoldingNestedExpressions(t *testing.T) {
	opt := NewOptimizer()

	// (2 + 3) * 4
	expr := &model.BinaryOp{
		Op: model.TK_Star,
		Left: &model.BinaryOp{
			Op:    model.TK_Plus,
			Left:  &model.IntLit{Value: "2"},
			Right: &model.IntLit{Value: "3"},
		},
		Right: &model.IntLit{Value: "4"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "20", intLit.Value)
}

func TestConstantFoldingUnaryNegation(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.UnaryOp{
		Op:      model.TK_Minus,
		Operand: &model.IntLit{Value: "5"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "-5", intLit.Value)
}

func TestConstantFoldingUnaryPlus(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.UnaryOp{
		Op:      model.TK_Plus,
		Operand: &model.IntLit{Value: "5"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "5", intLit.Value)
}

func TestConstantFoldingUnaryBitwiseNot(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.UnaryOp{
		Op:      model.TK_Tilde,
		Operand: &model.IntLit{Value: "0"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "-1", intLit.Value)
}

func TestConstantFoldingUnaryNot(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.UnaryOp{
		Op:      model.TK_Not,
		Operand: &model.BoolLit{Value: true},
	}

	result := opt.FoldConstants(expr)

	boolLit, ok := result.(*model.BoolLit)
	require.True(t, ok)
	assert.False(t, boolLit.Value)
}

func TestConstantFoldingBoolOpAnd(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BoolOp{
		Op:     model.TK_And,
		Values: []model.Expr{&model.BoolLit{Value: true}, &model.BoolLit{Value: true}},
	}

	result := opt.FoldConstants(expr)

	boolLit, ok := result.(*model.BoolLit)
	require.True(t, ok)
	assert.True(t, boolLit.Value)
}

func TestConstantFoldingBoolOpAndFalse(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BoolOp{
		Op:     model.TK_And,
		Values: []model.Expr{&model.BoolLit{Value: true}, &model.BoolLit{Value: false}},
	}

	result := opt.FoldConstants(expr)

	boolLit, ok := result.(*model.BoolLit)
	require.True(t, ok)
	assert.False(t, boolLit.Value)
}

func TestConstantFoldingBoolOpOr(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BoolOp{
		Op:     model.TK_Or,
		Values: []model.Expr{&model.BoolLit{Value: false}, &model.BoolLit{Value: true}},
	}

	result := opt.FoldConstants(expr)

	boolLit, ok := result.(*model.BoolLit)
	require.True(t, ok)
	assert.True(t, boolLit.Value)
}

func TestConstantFoldingIfExprTrue(t *testing.T) {
	opt := NewOptimizer()

	// 1 if True else 2
	expr := &model.IfExpr{
		Test:   &model.BoolLit{Value: true},
		Body:   &model.IntLit{Value: "1"},
		OrElse: &model.IntLit{Value: "2"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "1", intLit.Value)
}

func TestConstantFoldingIfExprFalse(t *testing.T) {
	opt := NewOptimizer()

	// 1 if False else 2
	expr := &model.IfExpr{
		Test:   &model.BoolLit{Value: false},
		Body:   &model.IntLit{Value: "1"},
		OrElse: &model.IntLit{Value: "2"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "2", intLit.Value)
}

func TestConstantFoldingWithVariable(t *testing.T) {
	opt := NewOptimizer()

	// x + 3 - should not be folded since x is not a constant
	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "3"},
	}

	result := opt.FoldConstants(expr)

	// Should remain as binary op
	binOp, ok := result.(*model.BinaryOp)
	require.True(t, ok)
	assert.NotNil(t, binOp.Left)
	assert.NotNil(t, binOp.Right)
}

// =============================================================================
// Strength Reduction Tests
// =============================================================================

func TestStrengthReductionMultiplyByZero(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Star,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "0"},
	}

	result := opt.ReduceStrength(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "0", intLit.Value)
}

func TestStrengthReductionMultiplyByOne(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Star,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "1"},
	}

	result := opt.ReduceStrength(expr)

	ident, ok := result.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "x", ident.Name)
}

func TestStrengthReductionMultiplyByTwoToShift(t *testing.T) {
	opt := NewOptimizer()

	// x * 2 -> x << 1 (only for integer x)
	expr := &model.BinaryOp{
		Op:    model.TK_Star,
		Left:  &model.IntLit{Value: "5"}, // Known integer
		Right: &model.IntLit{Value: "2"},
	}

	result := opt.ReduceStrength(expr)

	binOp, ok := result.(*model.BinaryOp)
	require.True(t, ok)
	assert.Equal(t, model.TK_LShift, binOp.Op)
}

func TestStrengthReductionPowerZero(t *testing.T) {
	opt := NewOptimizer()

	// x ** 0 -> 1
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "0"},
	}

	result := opt.ReduceStrength(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "1", intLit.Value)
}

func TestStrengthReductionPowerOne(t *testing.T) {
	opt := NewOptimizer()

	// x ** 1 -> x
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "1"},
	}

	result := opt.ReduceStrength(expr)

	ident, ok := result.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "x", ident.Name)
}

func TestStrengthReductionPowerTwo(t *testing.T) {
	opt := NewOptimizer()

	// x ** 2 -> x * x (only for pure expressions)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "2"},
	}

	result := opt.ReduceStrength(expr)

	binOp, ok := result.(*model.BinaryOp)
	require.True(t, ok)
	assert.Equal(t, model.TK_Star, binOp.Op)
}

func TestStrengthReductionAddZero(t *testing.T) {
	opt := NewOptimizer()

	// x + 0 -> x
	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "0"},
	}

	result := opt.ReduceStrength(expr)

	ident, ok := result.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "x", ident.Name)
}

func TestStrengthReductionSubtractZero(t *testing.T) {
	opt := NewOptimizer()

	// x - 0 -> x
	expr := &model.BinaryOp{
		Op:    model.TK_Minus,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "0"},
	}

	result := opt.ReduceStrength(expr)

	ident, ok := result.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "x", ident.Name)
}

func TestStrengthReductionFloorDivByOne(t *testing.T) {
	opt := NewOptimizer()

	// x // 1 -> x
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleSlash,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "1"},
	}

	result := opt.ReduceStrength(expr)

	ident, ok := result.(*model.Identifier)
	require.True(t, ok)
	assert.Equal(t, "x", ident.Name)
}

func TestStrengthReductionModByOne(t *testing.T) {
	opt := NewOptimizer()

	// x % 1 -> 0
	expr := &model.BinaryOp{
		Op:    model.TK_Percent,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "1"},
	}

	result := opt.ReduceStrength(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok)
	assert.Equal(t, "0", intLit.Value)
}

// =============================================================================
// Dead Code Elimination Tests
// =============================================================================

func TestDeadCodeAfterReturn(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.ExprStmt{Value: &model.IntLit{Value: "1"}},
		&model.Return{Value: &model.IntLit{Value: "2"}},
		&model.ExprStmt{Value: &model.IntLit{Value: "3"}}, // Dead code
	}

	result := opt.EliminateDeadCode(stmts)

	assert.Len(t, result, 2) // Third statement should be removed
}

func TestDeadCodeAfterRaise(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.ExprStmt{Value: &model.IntLit{Value: "1"}},
		&model.Raise{Exc: &model.Identifier{Name: "ValueError"}},
		&model.ExprStmt{Value: &model.IntLit{Value: "2"}}, // Dead code
	}

	result := opt.EliminateDeadCode(stmts)

	assert.Len(t, result, 2)
}

func TestDeadCodeAfterBreak(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.ExprStmt{Value: &model.IntLit{Value: "1"}},
		&model.Break{},
		&model.ExprStmt{Value: &model.IntLit{Value: "2"}}, // Dead code
	}

	result := opt.EliminateDeadCode(stmts)

	assert.Len(t, result, 2)
}

func TestDeadCodeAfterContinue(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.ExprStmt{Value: &model.IntLit{Value: "1"}},
		&model.Continue{},
		&model.ExprStmt{Value: &model.IntLit{Value: "2"}}, // Dead code
	}

	result := opt.EliminateDeadCode(stmts)

	assert.Len(t, result, 2)
}

func TestNoDeadCode(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.ExprStmt{Value: &model.IntLit{Value: "1"}},
		&model.ExprStmt{Value: &model.IntLit{Value: "2"}},
		&model.ExprStmt{Value: &model.IntLit{Value: "3"}},
	}

	result := opt.EliminateDeadCode(stmts)

	assert.Len(t, result, 3) // All statements retained
}

// =============================================================================
// If Statement Optimization Tests
// =============================================================================

func TestOptimizeIfAlwaysTrue(t *testing.T) {
	opt := NewOptimizer()

	ifStmt := &model.If{
		Test:   &model.BoolLit{Value: true},
		Body:   []model.Stmt{&model.ExprStmt{Value: &model.IntLit{Value: "1"}}},
		OrElse: []model.Stmt{&model.ExprStmt{Value: &model.IntLit{Value: "2"}}},
	}

	result := opt.OptimizeIf(ifStmt)

	// Should return just the body
	require.Len(t, result, 1)
	_, ok := result[0].(*model.ExprStmt)
	assert.True(t, ok)
}

func TestOptimizeIfAlwaysFalse(t *testing.T) {
	opt := NewOptimizer()

	ifStmt := &model.If{
		Test:   &model.BoolLit{Value: false},
		Body:   []model.Stmt{&model.ExprStmt{Value: &model.IntLit{Value: "1"}}},
		OrElse: []model.Stmt{&model.ExprStmt{Value: &model.IntLit{Value: "2"}}},
	}

	result := opt.OptimizeIf(ifStmt)

	// Should return just the else branch
	require.Len(t, result, 1)
	_, ok := result[0].(*model.ExprStmt)
	assert.True(t, ok)
}

func TestOptimizeIfAlwaysFalseNoElse(t *testing.T) {
	opt := NewOptimizer()

	ifStmt := &model.If{
		Test:   &model.BoolLit{Value: false},
		Body:   []model.Stmt{&model.ExprStmt{Value: &model.IntLit{Value: "1"}}},
		OrElse: nil,
	}

	result := opt.OptimizeIf(ifStmt)

	// Should be empty
	assert.Len(t, result, 0)
}

// =============================================================================
// Pure Expression Detection Tests
// =============================================================================

func TestIsPureExprLiterals(t *testing.T) {
	opt := NewOptimizer()

	assert.True(t, opt.isPureExpr(&model.IntLit{Value: "1"}))
	assert.True(t, opt.isPureExpr(&model.FloatLit{Value: "1.0"}))
	assert.True(t, opt.isPureExpr(&model.StringLit{Value: "hello"}))
	assert.True(t, opt.isPureExpr(&model.BoolLit{Value: true}))
	assert.True(t, opt.isPureExpr(&model.NoneLit{}))
}

func TestIsPureExprIdentifier(t *testing.T) {
	opt := NewOptimizer()
	assert.True(t, opt.isPureExpr(&model.Identifier{Name: "x"}))
}

func TestIsPureExprBinaryOp(t *testing.T) {
	opt := NewOptimizer()

	// Pure binary op
	pureExpr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.IntLit{Value: "1"},
		Right: &model.IntLit{Value: "2"},
	}
	assert.True(t, opt.isPureExpr(pureExpr))
}

func TestIsPureExprCall(t *testing.T) {
	opt := NewOptimizer()

	// Function call is not pure (could have side effects)
	callExpr := &model.Call{
		Func: &model.Identifier{Name: "func"},
		Args: []model.Expr{},
	}
	assert.False(t, opt.isPureExpr(callExpr))
}

// =============================================================================
// Integer Expression Detection Tests
// =============================================================================

func TestIsIntegerExprIntLit(t *testing.T) {
	opt := NewOptimizer()
	assert.True(t, opt.isIntegerExpr(&model.IntLit{Value: "5"}))
}

func TestIsIntegerExprBitwiseOps(t *testing.T) {
	opt := NewOptimizer()

	// Bitwise operations always produce integers
	shiftExpr := &model.BinaryOp{
		Op:    model.TK_LShift,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "2"},
	}
	assert.True(t, opt.isIntegerExpr(shiftExpr))
}

func TestIsIntegerExprFloorDiv(t *testing.T) {
	opt := NewOptimizer()

	// Floor division of ints produces int
	floorDivExpr := &model.BinaryOp{
		Op:    model.TK_DoubleSlash,
		Left:  &model.IntLit{Value: "10"},
		Right: &model.IntLit{Value: "3"},
	}
	assert.True(t, opt.isIntegerExpr(floorDivExpr))
}

func TestIsIntegerExprUnaryTilde(t *testing.T) {
	opt := NewOptimizer()

	tildeExpr := &model.UnaryOp{
		Op:      model.TK_Tilde,
		Operand: &model.Identifier{Name: "x"},
	}
	assert.True(t, opt.isIntegerExpr(tildeExpr))
}

// =============================================================================
// End-to-End Optimization Tests (via compilation)
// =============================================================================

func TestOptimizationEndToEndConstantFolding(t *testing.T) {
	source := `result = 2 + 3 * 4`

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(14), result.Value)
}

func TestOptimizationEndToEndDeadCode(t *testing.T) {
	source := `
def foo():
    return 1
    x = 2  # Dead code

result = foo()
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestOptimizationEndToEndStrengthReduction(t *testing.T) {
	source := `
x = 5
result = x * 2  # Could be optimized to x << 1
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(10), result.Value)
}

func TestOptimizationEndToEndIfTrue(t *testing.T) {
	source := `
if True:
    result = 1
else:
    result = 2
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestOptimizationEndToEndIfFalse(t *testing.T) {
	source := `
if False:
    result = 1
else:
    result = 2
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

// =============================================================================
// Peephole Optimization Tests (via bytecode inspection)
// =============================================================================

func TestPeepholeOptimizationIncrement(t *testing.T) {
	// Pattern: i = i + 1 should be optimized
	source := `
i = 0
i = i + 1
result = i
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestPeepholeOptimizationLoopCounter(t *testing.T) {
	source := `
result = 0
for i in range(100):
    result = result + 1
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(100), result.Value)
}

func TestPeepholeOptimizationEmptyList(t *testing.T) {
	// Empty list should use specialized opcode
	source := `result = []`

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Len(t, result.Items, 0)
}

func TestPeepholeOptimizationEmptyDict(t *testing.T) {
	source := `result = {}`

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyDict)
	assert.Len(t, result.Items, 0)
}

// =============================================================================
// Loop Invariant Code Motion Tests
// =============================================================================

func TestLoopInvariantCodeMotion(t *testing.T) {
	// y = x * 2 is invariant in the loop
	source := `
x = 5
result = 0
for i in range(10):
    y = x * 2  # Loop invariant
    result = result + y
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(100), result.Value) // 10 * 10 = 100
}

// =============================================================================
// Modified Variable Detection Tests
// =============================================================================

func TestFindModifiedVarsAssign(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.Assign{
			Targets: []model.Expr{&model.Identifier{Name: "x"}},
			Value:   &model.IntLit{Value: "1"},
		},
	}

	modified := opt.findModifiedVars(stmts)
	assert.True(t, modified["x"])
	assert.False(t, modified["y"])
}

func TestFindModifiedVarsAugAssign(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.AugAssign{
			Target: &model.Identifier{Name: "x"},
			Op:     model.TK_Plus,
			Value:  &model.IntLit{Value: "1"},
		},
	}

	modified := opt.findModifiedVars(stmts)
	assert.True(t, modified["x"])
}

func TestFindModifiedVarsTupleUnpack(t *testing.T) {
	opt := NewOptimizer()

	stmts := []model.Stmt{
		&model.Assign{
			Targets: []model.Expr{
				&model.Tuple{
					Elts: []model.Expr{
						&model.Identifier{Name: "a"},
						&model.Identifier{Name: "b"},
					},
				},
			},
			Value: &model.Tuple{
				Elts: []model.Expr{
					&model.IntLit{Value: "1"},
					&model.IntLit{Value: "2"},
				},
			},
		},
	}

	modified := opt.findModifiedVars(stmts)
	assert.True(t, modified["a"])
	assert.True(t, modified["b"])
}

// =============================================================================
// Expression Variable Usage Tests
// =============================================================================

func TestExprUsesVarIdentifier(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.Identifier{Name: "x"}
	assert.True(t, opt.exprUsesVar(expr, "x"))
	assert.False(t, opt.exprUsesVar(expr, "y"))
}

func TestExprUsesVarBinaryOp(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.BinaryOp{
		Op:    model.TK_Plus,
		Left:  &model.Identifier{Name: "x"},
		Right: &model.IntLit{Value: "1"},
	}

	assert.True(t, opt.exprUsesVar(expr, "x"))
	assert.False(t, opt.exprUsesVar(expr, "y"))
}

func TestExprUsesVarCall(t *testing.T) {
	opt := NewOptimizer()

	expr := &model.Call{
		Func: &model.Identifier{Name: "func"},
		Args: []model.Expr{&model.Identifier{Name: "x"}},
	}

	assert.True(t, opt.exprUsesVar(expr, "x"))
	assert.True(t, opt.exprUsesVar(expr, "func"))
	assert.False(t, opt.exprUsesVar(expr, "y"))
}

// =============================================================================
// Large Exponent Handling Tests
// =============================================================================

func TestConstantFoldingLargeExponentNotFolded(t *testing.T) {
	opt := NewOptimizer()

	// Large exponent should not be folded at compile time
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "1000"},
	}

	result := opt.FoldConstants(expr)

	// Should remain as binary op (not folded due to limit)
	_, ok := result.(*model.BinaryOp)
	assert.True(t, ok, "Large exponent should not be folded")
}

func TestConstantFoldingLargeShiftNotFolded(t *testing.T) {
	opt := NewOptimizer()

	// Large shift should not be folded
	expr := &model.BinaryOp{
		Op:    model.TK_LShift,
		Left:  &model.IntLit{Value: "1"},
		Right: &model.IntLit{Value: "100"},
	}

	result := opt.FoldConstants(expr)

	// Should remain as binary op (not folded due to limit)
	_, ok := result.(*model.BinaryOp)
	assert.True(t, ok, "Large shift should not be folded")
}
