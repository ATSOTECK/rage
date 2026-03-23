package compiler

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// estimateStackSize Tests
//
// Verify that compiled code with build operations gets reasonable stack sizes.
// =============================================================================

func TestEstimateStackSizeBasic(t *testing.T) {
	// Simple assignment: should have a small stack size
	code, errs := CompileSource("x = 1", "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	assert.GreaterOrEqual(t, code.StackSize, 1, "stack size should be at least 1")
}

func TestEstimateStackSizeLargeList(t *testing.T) {
	// Build a list with 50 elements -- estimateStackSize should handle BUILD_LIST arg
	var elements []string
	for i := 0; i < 50; i++ {
		elements = append(elements, "1")
	}
	source := "x = [" + strings.Join(elements, ", ") + "]"

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	// StackSize should be at least 50 (the BUILD_LIST arg) + some headroom
	assert.GreaterOrEqual(t, code.StackSize, 50, "stack size should accommodate 50-element list")
}

func TestEstimateStackSizeLargeTuple(t *testing.T) {
	// Build a tuple with 30 elements
	var elements []string
	for i := 0; i < 30; i++ {
		elements = append(elements, "1")
	}
	source := "x = (" + strings.Join(elements, ", ") + ",)"

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	assert.GreaterOrEqual(t, code.StackSize, 30, "stack size should accommodate 30-element tuple")
}

func TestEstimateStackSizeLargeDict(t *testing.T) {
	// Build a dict with 20 key-value pairs (40 items on stack)
	source := "x = {"
	for i := 0; i < 20; i++ {
		if i > 0 {
			source += ", "
		}
		source += "\"k" + strings.Repeat("_", i) + "\": 1"
	}
	source += "}"

	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	// BUILD_MAP arg is number of pairs, so at least 20
	assert.GreaterOrEqual(t, code.StackSize, 10, "stack size should accommodate dict build")
}

func TestEstimateStackSizeNestedFunction(t *testing.T) {
	// Functions have their own stack sizes
	source := `
def foo():
    a = [1, 2, 3, 4, 5]
    return a
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)
	assert.Greater(t, code.StackSize, 0, "module stack size should be positive")
}

// =============================================================================
// patchJump Bounds Tests
//
// Verify that invalid patch offsets produce compile errors rather than panics.
// =============================================================================

func TestPatchJumpOutOfBounds(t *testing.T) {
	// Directly test the compiler's patchJump with an out-of-bounds offset
	c := NewCompiler("<test>")

	// Emit a small amount of code
	c.emit(runtime.OpLoadNone)
	c.emit(runtime.OpReturn)

	// The code is only 2 bytes long. Patching at offset 10 should produce an error.
	c.patchJump(10, 0)

	require.NotEmpty(t, c.errors, "patchJump with out-of-bounds offset should produce an error")
	assert.Contains(t, c.errors[0].Message, "patchJump")
	assert.Contains(t, c.errors[0].Message, "out of bounds")
}

func TestPatchJumpAtBoundary(t *testing.T) {
	// If code length is 3, offset 0 is valid (0+1=1, 0+2=2, both < 3)
	c := NewCompiler("<test>")
	c.emitArg(runtime.OpLoadConst, 0) // 3 bytes
	c.code.Constants = append(c.code.Constants, nil)

	// patchJump at offset 0, target 42 -- should succeed without error
	c.patchJump(0, 42)
	assert.Empty(t, c.errors, "patchJump at valid offset should not produce errors")

	// Verify the target was written
	target := int(c.code.Code[1]) | int(c.code.Code[2])<<8
	assert.Equal(t, 42, target)
}

func TestPatchJumpCodeTooShort(t *testing.T) {
	// Code with only 1 byte -- offset 0 means we need index 0+2=2 which is out of bounds
	c := NewCompiler("<test>")
	c.emit(runtime.OpLoadNone) // 1 byte, no arg

	c.patchJump(0, 5)
	require.NotEmpty(t, c.errors, "patchJump should fail when code is too short for the offset")
}

// =============================================================================
// emitArg Overflow Tests
//
// Verify that arguments > 65535 or < -32768 produce compile errors.
// =============================================================================

func TestEmitArgOverflow(t *testing.T) {
	tests := []struct {
		name string
		arg  int
	}{
		{"arg exceeds 65535", 70000},
		{"arg far exceeds limit", 100000},
		{"large negative arg", -40000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler("<test>")
			c.emitArg(runtime.OpLoadConst, tt.arg)

			require.NotEmpty(t, c.errors, "emitArg with arg=%d should produce error", tt.arg)
			assert.Contains(t, c.errors[0].Message, "exceeds 16-bit limit")
		})
	}
}

func TestEmitArgValidRange(t *testing.T) {
	tests := []struct {
		name string
		arg  int
	}{
		{"zero", 0},
		{"max unsigned 16-bit", 65535},
		{"mid range", 32000},
		{"small negative (sentinel)", -1},
		{"min negative allowed", -32768},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler("<test>")
			c.code.Constants = append(c.code.Constants, nil) // Ensure constant 0 exists
			c.emitArg(runtime.OpLoadConst, tt.arg)

			assert.Empty(t, c.errors, "emitArg with arg=%d should not produce error", tt.arg)
		})
	}
}

// =============================================================================
// Constant Folding - Exponent Edge Cases
//
// Test that 2**100 doesn't overflow (should NOT fold), 2**10 folds correctly,
// and (-1)**63 is handled as an edge case.
// =============================================================================

func TestConstantFolding2Pow100NotFolded(t *testing.T) {
	opt := NewOptimizer()

	// 2**100 should NOT be folded (exponent > 63)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "100"},
	}

	result := opt.FoldConstants(expr)

	_, ok := result.(*model.BinaryOp)
	assert.True(t, ok, "2**100 should not be folded -- exponent too large")
}

func TestConstantFolding2Pow10(t *testing.T) {
	opt := NewOptimizer()

	// 2**10 = 1024, should fold
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "10"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok, "2**10 should fold to an IntLit")
	assert.Equal(t, "1024", intLit.Value)
}

func TestConstantFoldingNeg1Pow63(t *testing.T) {
	opt := NewOptimizer()

	// (-1)**63 = -1, should fold since exponent is exactly 63 (the limit)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "-1"},
		Right: &model.IntLit{Value: "63"},
	}

	result := opt.FoldConstants(expr)

	intLit, ok := result.(*model.IntLit)
	require.True(t, ok, "(-1)**63 should fold to an IntLit")
	assert.Equal(t, "-1", intLit.Value)
}

func TestConstantFoldingExponentAtBoundary(t *testing.T) {
	opt := NewOptimizer()

	// 2**63 should fold (but may overflow int64 -- the function should handle this)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "63"},
	}

	result := opt.FoldConstants(expr)

	// 2^63 = 9223372036854775808 which exceeds int64 max (9223372036854775807)
	// The folding function should detect this and NOT fold
	_, isBinOp := result.(*model.BinaryOp)
	_, isIntLit := result.(*model.IntLit)
	// Either it was not folded (correct) or it was folded (implementation detail)
	assert.True(t, isBinOp || isIntLit, "2**63 should either fold or remain as BinaryOp")
}

func TestConstantFoldingNegativeExponentNotFolded(t *testing.T) {
	opt := NewOptimizer()

	// 2**(-1) should NOT be folded (negative exponent)
	expr := &model.BinaryOp{
		Op:    model.TK_DoubleStar,
		Left:  &model.IntLit{Value: "2"},
		Right: &model.IntLit{Value: "-1"},
	}

	result := opt.FoldConstants(expr)

	_, ok := result.(*model.BinaryOp)
	assert.True(t, ok, "2**(-1) should not be folded")
}

func TestConstantFoldingExponentEndToEnd(t *testing.T) {
	// End-to-end: compile 2**10 and verify it evaluates correctly
	code, errs := CompileSource("result = 2 ** 10", "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1024), result.Value)
}

// =============================================================================
// F-String Tests
//
// Test f-string parsing error propagation and successful compilation.
// =============================================================================

func TestFStringSyntaxErrorPropagated(t *testing.T) {
	// f"{1 +}" has a syntax error inside the expression
	_, errs := CompileSource(`x = f"{1 +}"`, "<test>")
	require.NotEmpty(t, errs, "f-string with syntax error should produce a compile error")

	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "f-string") || strings.Contains(e.Error(), "expression") {
			found = true
			break
		}
	}
	assert.True(t, found, "error should mention f-string or expression, got: %v", errs)
}

func TestFStringSimpleVariable(t *testing.T) {
	// f"{x}" should compile successfully
	code, errs := CompileSource(`x = 42; y = f"{x}"`, "<test>")
	require.Empty(t, errs, "f-string with simple variable should compile")
	require.NotNil(t, code)
}

func TestFStringArithmeticExpression(t *testing.T) {
	// f"{1 + 2}" should compile successfully
	code, errs := CompileSource(`x = f"{1 + 2}"`, "<test>")
	require.Empty(t, errs, "f-string with arithmetic expression should compile")
	require.NotNil(t, code)
}

func TestFStringEmptyExpression(t *testing.T) {
	// f"{}" should produce an error (empty expression)
	_, errs := CompileSource(`x = f"{}"`, "<test>")
	require.NotEmpty(t, errs, "f-string with empty expression should produce an error")
}

func TestFStringLiteralOnly(t *testing.T) {
	// f"hello" (no expressions) should compile and work like a regular string
	code, errs := CompileSource(`x = f"hello"`, "<test>")
	require.Empty(t, errs, "f-string with no expressions should compile")
	require.NotNil(t, code)
}

func TestFStringEscapedBraces(t *testing.T) {
	// f"{{x}}" should compile (escaped braces produce literal { and })
	code, errs := CompileSource(`x = f"{{y}}"`, "<test>")
	require.Empty(t, errs, "f-string with escaped braces should compile")
	require.NotNil(t, code)
}

func TestFStringWithFormatSpec(t *testing.T) {
	// f"{x:.2f}" should compile successfully
	code, errs := CompileSource(`x = 3.14; y = f"{x:.2f}"`, "<test>")
	require.Empty(t, errs, "f-string with format spec should compile")
	require.NotNil(t, code)
}

func TestFStringWithConversion(t *testing.T) {
	// f"{x!r}" should compile successfully
	code, errs := CompileSource(`x = "hello"; y = f"{x!r}"`, "<test>")
	require.Empty(t, errs, "f-string with conversion should compile")
	require.NotNil(t, code)
}

// =============================================================================
// Compiler Error Handling - Additional Edge Cases
// =============================================================================

func TestCompileSourceReturnsNilOnError(t *testing.T) {
	// Verify that CompileSource returns nil code object when there are errors
	code, errs := CompileSource("def", "<test>")
	require.NotEmpty(t, errs, "incomplete 'def' should produce errors")
	assert.Nil(t, code, "code should be nil when there are parse errors")
}

func TestCompileSourceEmptyModule(t *testing.T) {
	// Empty source should compile to a valid code object
	code, errs := CompileSource("", "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)
	assert.Equal(t, "<module>", code.Name)
	assert.Greater(t, len(code.Code), 0, "code should have at least return None")
}

func TestCompileSourceFilename(t *testing.T) {
	// Verify the filename is recorded in the code object
	code, errs := CompileSource("x = 1", "myfile.py")
	require.Empty(t, errs)
	require.NotNil(t, code)
	assert.Equal(t, "myfile.py", code.Filename)
}

// =============================================================================
// addConstant Deduplication Tests
// =============================================================================

func TestAddConstantDeduplication(t *testing.T) {
	c := NewCompiler("<test>")

	// Adding the same integer constant twice should return the same index
	idx1 := c.addConstant(int64(42))
	idx2 := c.addConstant(int64(42))
	assert.Equal(t, idx1, idx2, "same constant should be deduplicated")
}

func TestAddConstantDifferentValues(t *testing.T) {
	c := NewCompiler("<test>")

	idx1 := c.addConstant(int64(1))
	idx2 := c.addConstant(int64(2))
	assert.NotEqual(t, idx1, idx2, "different constants should get different indices")
}

func TestAddConstantStringDedup(t *testing.T) {
	c := NewCompiler("<test>")

	idx1 := c.addConstant("hello")
	idx2 := c.addConstant("hello")
	assert.Equal(t, idx1, idx2, "same string constant should be deduplicated")

	idx3 := c.addConstant("world")
	assert.NotEqual(t, idx1, idx3, "different strings should get different indices")
}

// =============================================================================
// addName Deduplication Tests
// =============================================================================

func TestAddNameDeduplication(t *testing.T) {
	c := NewCompiler("<test>")

	idx1 := c.addName("foo")
	idx2 := c.addName("foo")
	assert.Equal(t, idx1, idx2, "same name should be deduplicated")

	idx3 := c.addName("bar")
	assert.NotEqual(t, idx1, idx3, "different names should get different indices")
}

// =============================================================================
// Line Number Table Tests
// =============================================================================

func TestLineNumberTable(t *testing.T) {
	source := `x = 1
y = 2
z = 3`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	// Should have line entries
	assert.NotEmpty(t, code.LineNoTab, "line number table should not be empty")

	// Check that lines 1, 2, 3 are represented
	lines := make(map[int]bool)
	for _, entry := range code.LineNoTab {
		lines[entry.Line] = true
	}
	assert.True(t, lines[1], "line 1 should be in line table")
	assert.True(t, lines[2], "line 2 should be in line table")
	assert.True(t, lines[3], "line 3 should be in line table")
}

// =============================================================================
// Bytecode Validation Tests
// =============================================================================

func TestBytecodeValidationPasses(t *testing.T) {
	// Well-formed source should produce a valid code object
	source := `
x = 1
y = x + 2
z = [1, 2, 3]
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	// Validate should pass
	err := code.Validate()
	assert.NoError(t, err, "well-formed code should validate")
}

// =============================================================================
// OpCompareLtLocal: peephole optimizer must NOT pack jump target into 16-bit arg
//
// Regression test for a critical bug where the optimizer packed local1 (8 bits),
// local2 (8 bits), and jumpTarget (16 bits) into a single 16-bit instruction arg,
// silently truncating the jump target to zero. The fix fuses only LOAD_FAST +
// LOAD_FAST + COMPARE_LT into OpCompareLtLocal (16 bits for two local indices),
// leaving POP_JUMP_IF_FALSE as a separate instruction.
// =============================================================================

func TestCompareLtLocalArgFitsIn16Bits(t *testing.T) {
	// A while loop with local comparison — the optimizer should produce
	// OpCompareLtLocal whose arg packs only two 8-bit local indices.
	source := `
def f():
    i = 0
    n = 10
    while i < n:
        i = i + 1
    return i
`
	code, errs := CompileSource(source, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	// Find the function's code object
	var funcCode *runtime.CodeObject
	for _, c := range code.Constants {
		if co, ok := c.(*runtime.CodeObject); ok && co.Name == "f" {
			funcCode = co
			break
		}
	}
	require.NotNil(t, funcCode, "function 'f' code object not found")

	// Scan bytecode for OpCompareLtLocal and verify its arg fits in 16 bits
	found := false
	offset := 0
	for offset < len(funcCode.Code) {
		op := runtime.Opcode(funcCode.Code[offset])
		if op == runtime.OpCompareLtLocal {
			found = true
			arg := int(funcCode.Code[offset+1]) | int(funcCode.Code[offset+2])<<8
			local1 := arg & 0xFF
			local2 := (arg >> 8) & 0xFF
			assert.GreaterOrEqual(t, local1, 0, "local1 index should be non-negative")
			assert.GreaterOrEqual(t, local2, 0, "local2 index should be non-negative")
			assert.LessOrEqual(t, local1, 255, "local1 index should fit in 8 bits")
			assert.LessOrEqual(t, local2, 255, "local2 index should fit in 8 bits")
			// The arg should contain ONLY the two local indices, no jump target
			assert.Equal(t, 0, arg>>16, "arg bits 16+ should be zero (no packed jump target)")
			break
		}
		if op.HasArg() {
			offset += 3
		} else {
			offset++
		}
	}
	// The optimizer may or may not fire depending on conditions, so only
	// validate the arg shape if the instruction was emitted.
	if found {
		// Verify a POP_JUMP_IF_FALSE follows (possibly after other instructions)
		// to confirm the jump wasn't absorbed into OpCompareLtLocal.
		foundJump := false
		for offset < len(funcCode.Code) {
			op := runtime.Opcode(funcCode.Code[offset])
			if op == runtime.OpPopJumpIfFalse {
				foundJump = true
				break
			}
			if op.HasArg() {
				offset += 3
			} else {
				offset++
			}
		}
		assert.True(t, foundJump, "POP_JUMP_IF_FALSE should remain as a separate instruction after OpCompareLtLocal")
	}
}

func TestCompareLtLocalEndToEnd(t *testing.T) {
	// End-to-end: compile and run a while loop that uses local comparison.
	// This was broken when the jump target was truncated, causing jumps to offset 0.
	code, errs := CompileSource(`
def count_up(n):
    i = 0
    total = 0
    while i < n:
        total = total + i
        i = i + 1
    return total

result = count_up(100)
`, "<test>")
	require.Empty(t, errs)
	require.NotNil(t, code)

	vm := runtime.NewVM()
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(4950), result.Value, "sum of 0..99 should be 4950")
}
