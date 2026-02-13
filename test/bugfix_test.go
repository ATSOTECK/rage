package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Modulo Operator Tests (Python semantics: result has same sign as divisor)
// =============================================================================

func TestModuloPositiveNumbers(t *testing.T) {
	vm := runCode(t, `result = 7 % 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestModuloNegativeDividend(t *testing.T) {
	// Python: -7 % 3 = 2 (result has same sign as divisor, which is positive)
	vm := runCode(t, `result = -7 % 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func TestModuloNegativeDivisor(t *testing.T) {
	// Python: 7 % -3 = -2 (result has same sign as divisor, which is negative)
	vm := runCode(t, `result = 7 % -3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-2), result.Value)
}

func TestModuloBothNegative(t *testing.T) {
	// Python: -7 % -3 = -1 (result has same sign as divisor, which is negative)
	vm := runCode(t, `result = -7 % -3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-1), result.Value)
}

func TestModuloZeroResult(t *testing.T) {
	vm := runCode(t, `result = -6 % 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(0), result.Value)
}

// =============================================================================
// Floor Division Tests (Python semantics: rounds toward negative infinity)
// =============================================================================

func TestFloorDivPositiveNumbers(t *testing.T) {
	vm := runCode(t, `result = 7 // 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func TestFloorDivNegativeDividend(t *testing.T) {
	// Python: -7 // 3 = -3 (rounds toward negative infinity)
	vm := runCode(t, `result = -7 // 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-3), result.Value)
}

func TestFloorDivNegativeDivisor(t *testing.T) {
	// Python: 7 // -3 = -3 (rounds toward negative infinity)
	vm := runCode(t, `result = 7 // -3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-3), result.Value)
}

func TestFloorDivBothNegative(t *testing.T) {
	// Python: -7 // -3 = 2 (rounds toward negative infinity, but result is positive)
	vm := runCode(t, `result = -7 // -3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func TestFloorDivExact(t *testing.T) {
	vm := runCode(t, `result = -6 // 3`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-2), result.Value)
}

// =============================================================================
// Shift Operation Tests (bounds checking)
// =============================================================================

func TestLeftShiftPositive(t *testing.T) {
	vm := runCode(t, `result = 1 << 4`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(16), result.Value)
}

func TestRightShiftPositive(t *testing.T) {
	vm := runCode(t, `result = 16 >> 2`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(4), result.Value)
}

func TestLeftShiftNegativeCount(t *testing.T) {
	runCodeExpectError(t, `result = 1 << -1`, "ValueError: negative shift count")
}

func TestRightShiftNegativeCount(t *testing.T) {
	runCodeExpectError(t, `result = 16 >> -1`, "ValueError: negative shift count")
}

func TestLeftShiftLargeCount(t *testing.T) {
	// Very large left shift should return 0 (overflow)
	vm := runCode(t, `result = 1 << 100`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(0), result.Value)
}

func TestRightShiftLargeCount(t *testing.T) {
	// Very large right shift of positive number should return 0
	vm := runCode(t, `result = 100 >> 100`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(0), result.Value)
}

func TestRightShiftNegativeNumberLargeCount(t *testing.T) {
	// Very large right shift of negative number should return -1
	vm := runCode(t, `result = -100 >> 100`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-1), result.Value)
}

// =============================================================================
// Power Operation Tests (integer precision)
// =============================================================================

func TestPowerSmall(t *testing.T) {
	vm := runCode(t, `result = 2 ** 10`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1024), result.Value)
}

func TestPowerZeroExponent(t *testing.T) {
	vm := runCode(t, `result = 5 ** 0`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestPowerLargerExponent(t *testing.T) {
	// 3^20 = 3486784401
	vm := runCode(t, `result = 3 ** 20`)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3486784401), result.Value)
}

func TestPowerNegativeExponent(t *testing.T) {
	// Negative exponent should return float
	vm := runCode(t, `result = 2 ** -1`)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 0.5, result.Value)
}

// =============================================================================
// Cycle Detection Tests (equality on self-referential structures)
// =============================================================================

func TestSelfReferentialListEquality(t *testing.T) {
	// Create a list that contains itself, compare with itself
	// Should not cause stack overflow
	vm := runCode(t, `
a = [1, 2]
a.append(a)
result = a == a
`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestMutualReferentialListEquality(t *testing.T) {
	// Two lists that reference each other
	vm := runCode(t, `
a = [1]
b = [1]
a.append(b)
b.append(a)
result = a == a
`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestNestedListEquality(t *testing.T) {
	// Normal nested lists should still compare correctly
	vm := runCode(t, `
a = [[1, 2], [3, 4]]
b = [[1, 2], [3, 4]]
c = [[1, 2], [3, 5]]
result1 = a == b
result2 = a == c
`)
	result1 := vm.GetGlobal("result1").(*runtime.PyBool)
	result2 := vm.GetGlobal("result2").(*runtime.PyBool)
	assert.True(t, result1.Value)
	assert.False(t, result2.Value)
}

// =============================================================================
// Hashable Type Validation Tests
// =============================================================================

func TestDictWithListKeyFails(t *testing.T) {
	runCodeExpectError(t, `d = {[1, 2]: "value"}`, "TypeError: unhashable type: 'list'")
}

func TestDictWithDictKeyFails(t *testing.T) {
	runCodeExpectError(t, `d = {{"a": 1}: "value"}`, "TypeError: unhashable type: 'dict'")
}

func TestDictWithSetKeyFails(t *testing.T) {
	runCodeExpectError(t, `d = {{1, 2}: "value"}`, "TypeError: unhashable type: 'set'")
}

func TestDictAssignmentWithListKeyFails(t *testing.T) {
	runCodeExpectError(t, `
d = {}
d[[1, 2]] = "value"
`, "TypeError: unhashable type: 'list'")
}

func TestSetWithListFails(t *testing.T) {
	runCodeExpectError(t, `s = {[1, 2]}`, "TypeError: unhashable type: 'list'")
}

func TestSetFunctionWithListFails(t *testing.T) {
	// Using set() builtin with a list containing unhashable items
	runCodeExpectError(t, `s = set([[1, 2]])`, "TypeError: unhashable type: 'list'")
}

func TestDictWithTupleKeySucceeds(t *testing.T) {
	// Tuples are hashable and should work as dict keys
	vm := runCode(t, `
d = {(1, 2): "value"}
result = d[(1, 2)]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "value", result.Value)
}

func TestDictWithStringKeySucceeds(t *testing.T) {
	vm := runCode(t, `
d = {"key": "value"}
result = d["key"]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "value", result.Value)
}

func TestDictWithIntKeySucceeds(t *testing.T) {
	vm := runCode(t, `
d = {42: "value"}
result = d[42]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "value", result.Value)
}

// =============================================================================
// String Repetition Limit Tests
// =============================================================================

func TestStringRepetitionNormal(t *testing.T) {
	vm := runCode(t, `result = "ab" * 3`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "ababab", result.Value)
}

func TestStringRepetitionZero(t *testing.T) {
	vm := runCode(t, `result = "hello" * 0`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "", result.Value)
}

func TestStringRepetitionNegative(t *testing.T) {
	vm := runCode(t, `result = "hello" * -5`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "", result.Value)
}

func TestStringRepetitionReversed(t *testing.T) {
	vm := runCode(t, `result = 3 * "ab"`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "ababab", result.Value)
}

func TestStringRepetitionTooLarge(t *testing.T) {
	// Attempting to create a string larger than 100MB should fail
	runCodeExpectError(t, `result = "x" * 200000000`, "MemoryError")
}

// =============================================================================
// List Repetition Limit Tests
// =============================================================================

func TestListRepetitionNormal(t *testing.T) {
	vm := runCode(t, `result = [1, 2] * 3`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 6, len(result.Items))
}

func TestListRepetitionZero(t *testing.T) {
	vm := runCode(t, `result = [1, 2, 3] * 0`)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Equal(t, 0, len(result.Items))
}

func TestListRepetitionTooLarge(t *testing.T) {
	// Attempting to create a list with more than 10M items should fail
	runCodeExpectError(t, `result = [1] * 20000000`, "MemoryError")
}

// =============================================================================
// UTF-8 String Indexing Tests
// =============================================================================

func TestASCIIStringIndexing(t *testing.T) {
	vm := runCode(t, `
s = "hello"
result = s[1]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "e", result.Value)
}

func TestUnicodeStringIndexing(t *testing.T) {
	// Test with multi-byte UTF-8 characters
	vm := runCode(t, `
s = "héllo"
result = s[1]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "é", result.Value)
}

func TestEmojiStringIndexing(t *testing.T) {
	// Emojis are multi-byte UTF-8 characters
	vm := runCode(t, `
s = "a😀b"
result0 = s[0]
result1 = s[1]
result2 = s[2]
`)
	result0 := vm.GetGlobal("result0").(*runtime.PyString)
	result1 := vm.GetGlobal("result1").(*runtime.PyString)
	result2 := vm.GetGlobal("result2").(*runtime.PyString)
	assert.Equal(t, "a", result0.Value)
	assert.Equal(t, "😀", result1.Value)
	assert.Equal(t, "b", result2.Value)
}

func TestChineseStringIndexing(t *testing.T) {
	vm := runCode(t, `
s = "你好世界"
result = s[2]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "世", result.Value)
}

func TestNegativeStringIndex(t *testing.T) {
	vm := runCode(t, `
s = "hello"
result = s[-1]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "o", result.Value)
}

func TestNegativeUnicodeStringIndex(t *testing.T) {
	vm := runCode(t, `
s = "你好"
result = s[-1]
`)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "好", result.Value)
}

// =============================================================================
// String Containment Tests (optimized with strings.Contains)
// =============================================================================

func TestStringContainmentTrue(t *testing.T) {
	vm := runCode(t, `result = "ell" in "hello"`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestStringContainmentFalse(t *testing.T) {
	vm := runCode(t, `result = "xyz" in "hello"`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.False(t, result.Value)
}

func TestStringContainmentEmpty(t *testing.T) {
	// Empty string is always contained in any string
	vm := runCode(t, `result = "" in "hello"`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestStringContainmentUnicode(t *testing.T) {
	vm := runCode(t, `result = "世" in "你好世界"`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestStringNotContainment(t *testing.T) {
	vm := runCode(t, `result = "xyz" not in "hello"`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestListContainment(t *testing.T) {
	vm := runCode(t, `result = 2 in [1, 2, 3]`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestDictKeyContainment(t *testing.T) {
	vm := runCode(t, `result = "a" in {"a": 1, "b": 2}`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Combined Tests (ensure fixes work together)
// =============================================================================

func TestModuloFloorDivConsistency(t *testing.T) {
	// Python invariant: a == (a // b) * b + (a % b)
	vm := runCode(t, `
a = -7
b = 3
quotient = a // b
remainder = a % b
reconstructed = quotient * b + remainder
result = a == reconstructed
`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestModuloFloorDivConsistencyNegativeDivisor(t *testing.T) {
	vm := runCode(t, `
a = 7
b = -3
quotient = a // b
remainder = a % b
reconstructed = quotient * b + remainder
result = a == reconstructed
`)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}
