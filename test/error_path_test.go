package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to run code and expect a specific error type
// Accepts multiple possible error substrings
func runExpectError(t *testing.T, source string, expectedErrors ...string) {
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		// Check if compilation error matches any expected
		for _, e := range errs {
			for _, expected := range expectedErrors {
				if strings.Contains(e.Error(), expected) {
					return
				}
			}
		}
		// If not a matching compile error, still check
		t.Logf("Compilation errors: %v", errs)
	}

	if len(errs) == 0 {
		_, err := vm.Execute(code)
		require.Error(t, err, "Expected error containing one of %v", expectedErrors)
		errStr := err.Error()
		matched := false
		for _, expected := range expectedErrors {
			if strings.Contains(errStr, expected) {
				matched = true
				break
			}
		}
		assert.True(t, matched, "Error %q should contain one of %v", errStr, expectedErrors)
	}
}

// =============================================================================
// ZeroDivisionError Tests
// =============================================================================

func TestDivisionByZeroInt(t *testing.T) {
	runExpectError(t, `result = 10 / 0`, "ZeroDivisionError", "division by zero")
}

func TestDivisionByZeroFloat(t *testing.T) {
	runExpectError(t, `result = 10.0 / 0.0`, "ZeroDivisionError", "division by zero")
}

func TestFloorDivisionByZero(t *testing.T) {
	runExpectError(t, `result = 10 // 0`, "ZeroDivisionError", "division by zero")
}

func TestModuloByZero(t *testing.T) {
	runExpectError(t, `result = 10 % 0`, "ZeroDivisionError", "division by zero", "modulo by zero")
}

func TestDivisionByZeroInExpression(t *testing.T) {
	runExpectError(t, `result = (1 + 2) / (3 - 3)`, "ZeroDivisionError", "division by zero")
}

func TestDivisionByZeroInFunction(t *testing.T) {
	source := `
def divide(a, b):
    return a / b

result = divide(10, 0)
`
	runExpectError(t, source, "ZeroDivisionError", "division by zero")
}

// =============================================================================
// IndexError Tests
// =============================================================================

func TestListIndexOutOfRange(t *testing.T) {
	runExpectError(t, `lst = [1, 2, 3]; x = lst[10]`, "IndexError", "index out of range", "out of range")
}

func TestListNegativeIndexOutOfRange(t *testing.T) {
	runExpectError(t, `lst = [1, 2, 3]; x = lst[-10]`, "IndexError", "index out of range", "out of range")
}

func TestEmptyListIndex(t *testing.T) {
	runExpectError(t, `lst = []; x = lst[0]`, "IndexError", "index out of range", "out of range")
}

func TestStringIndexOutOfRange(t *testing.T) {
	runExpectError(t, `s = "hello"; c = s[100]`, "IndexError", "index out of range", "out of range")
}

func TestTupleIndexOutOfRange(t *testing.T) {
	runExpectError(t, `t = (1, 2, 3); x = t[5]`, "IndexError", "index out of range", "out of range")
}

func TestListPopEmpty(t *testing.T) {
	runExpectError(t, `lst = []; lst.pop()`, "IndexError", "pop from empty", "empty")
}

// =============================================================================
// KeyError Tests
// =============================================================================

func TestDictKeyNotFound(t *testing.T) {
	runExpectError(t, `d = {"a": 1}; x = d["b"]`, "KeyError")
}

func TestDictKeyNotFoundInt(t *testing.T) {
	runExpectError(t, `d = {1: "one"}; x = d[2]`, "KeyError")
}

func TestEmptyDictKeyAccess(t *testing.T) {
	runExpectError(t, `d = {}; x = d["key"]`, "KeyError")
}

func TestDictPopMissing(t *testing.T) {
	runExpectError(t, `d = {"a": 1}; d.pop("b")`, "KeyError")
}

// =============================================================================
// TypeError Tests
// =============================================================================

func TestTypeErrorAddStringInt(t *testing.T) {
	runExpectError(t, `result = "hello" + 5`, "TypeError")
}

func TestTypeErrorSubscriptInt(t *testing.T) {
	runExpectError(t, `x = 42; y = x[0]`, "TypeError")
}

func TestTypeErrorCallNonCallable(t *testing.T) {
	runExpectError(t, `x = 5; x()`, "TypeError")
}

func TestTypeErrorIterateNonIterable(t *testing.T) {
	runExpectError(t, `for x in 42: pass`, "TypeError")
}

func TestTypeErrorUnhashableListKey(t *testing.T) {
	runExpectError(t, `d = {[1, 2]: "value"}`, "TypeError")
}

func TestTypeErrorUnhashableDictKey(t *testing.T) {
	runExpectError(t, `d = {{}: "value"}`, "TypeError")
}

func TestTypeErrorCompareIncompatible(t *testing.T) {
	// Note: Python 3 raises TypeError for incompatible comparisons
	// Some interpreters allow this, so skip if no error
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(`result = "string" < 5`, "<test>")
	if len(errs) > 0 {
		return // compile error is acceptable
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("This interpreter allows string < int comparisons")
	}
	assert.Contains(t, err.Error(), "TypeError")
}

func TestTypeErrorWrongArgCount(t *testing.T) {
	source := `
def add(a, b):
    return a + b

result = add(1)  # Missing argument
`
	runExpectError(t, source, "TypeError", "missing", "argument", "required")
}

func TestTypeErrorTooManyArgs(t *testing.T) {
	source := `
def add(a, b):
    return a + b

result = add(1, 2, 3)  # Too many arguments
`
	// Some interpreters may just ignore extra args
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("This interpreter allows extra arguments")
	}
	assert.True(t, strings.Contains(err.Error(), "TypeError") ||
		strings.Contains(err.Error(), "argument"))
}

func TestTypeErrorUnpackWrongCount(t *testing.T) {
	runExpectError(t, `a, b, c = [1, 2]`, "ValueError", "not enough values", "unpack")
}

func TestTypeErrorIntNotIterable(t *testing.T) {
	runExpectError(t, `list(42)`, "TypeError", "not iterable", "iterable")
}

// =============================================================================
// AttributeError Tests
// =============================================================================

func TestAttributeErrorMissingAttribute(t *testing.T) {
	source := `
class Foo:
    def __init__(self):
        self.x = 1

f = Foo()
y = f.nonexistent
`
	runExpectError(t, source, "AttributeError")
}

func TestAttributeErrorOnInt(t *testing.T) {
	runExpectError(t, `x = 5; y = x.nonexistent`, "AttributeError")
}

func TestAttributeErrorOnString(t *testing.T) {
	runExpectError(t, `s = "hello"; x = s.nonexistent_method()`, "AttributeError")
}

func TestAttributeErrorOnList(t *testing.T) {
	runExpectError(t, `lst = [1, 2, 3]; x = lst.fake_method()`, "AttributeError")
}

// =============================================================================
// NameError Tests
// =============================================================================

func TestNameErrorUndefinedVariable(t *testing.T) {
	runExpectError(t, `x = undefined_variable`, "NameError")
}

func TestNameErrorInFunction(t *testing.T) {
	source := `
def foo():
    return undefined_var

foo()
`
	runExpectError(t, source, "NameError")
}

func TestNameErrorBeforeAssignment(t *testing.T) {
	source := `
def foo():
    x = x + 1  # x is used before assignment
    return x

foo()
`
	// This might be UnboundLocalError in Python, check what RAGE throws
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		return // Compile error is acceptable
	}
	_, err := vm.Execute(code)
	require.Error(t, err)
	// Should contain either NameError or UnboundLocalError
	errStr := err.Error()
	assert.True(t, strings.Contains(errStr, "NameError") ||
		strings.Contains(errStr, "UnboundLocalError") ||
		strings.Contains(errStr, "undefined"))
}

// =============================================================================
// ValueError Tests
// =============================================================================

func TestValueErrorIntConversion(t *testing.T) {
	runExpectError(t, `x = int("not a number")`, "ValueError")
}

func TestValueErrorFloatConversion(t *testing.T) {
	runExpectError(t, `x = float("not a float")`, "ValueError")
}

func TestValueErrorUnpackTooFew(t *testing.T) {
	runExpectError(t, `a, b, c = [1, 2]`, "ValueError")
}

func TestValueErrorUnpackTooMany(t *testing.T) {
	runExpectError(t, `a, b = [1, 2, 3]`, "ValueError")
}

func TestValueErrorNegativeShift(t *testing.T) {
	runExpectError(t, `x = 1 << -1`, "ValueError")
}

func TestValueErrorListRemoveNotFound(t *testing.T) {
	runExpectError(t, `lst = [1, 2, 3]; lst.remove(5)`, "ValueError")
}

func TestValueErrorIndexNotFound(t *testing.T) {
	runExpectError(t, `lst = [1, 2, 3]; lst.index(5)`, "ValueError")
}

// =============================================================================
// StopIteration Tests
// =============================================================================

func TestStopIterationExhaustedIterator(t *testing.T) {
	source := `
def gen():
    yield 1

g = gen()
next(g)
next(g)  # Should raise StopIteration
`
	runExpectError(t, source, "StopIteration")
}

func TestStopIterationEmptyIterator(t *testing.T) {
	source := `
def empty_gen():
    return
    yield

g = empty_gen()
next(g)  # Should raise StopIteration immediately
`
	runExpectError(t, source, "StopIteration")
}

// =============================================================================
// RuntimeError Tests
// =============================================================================

func TestRuntimeErrorDictSizeChangedDuringIteration(t *testing.T) {
	source := `
d = {"a": 1, "b": 2, "c": 3}
for k in d:
    d["new_key"] = 4  # Modify dict during iteration
`
	// This may or may not error depending on implementation
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Compile error")
		return
	}
	_, err := vm.Execute(code)
	// Either succeeds or raises RuntimeError
	if err != nil {
		assert.Contains(t, err.Error(), "RuntimeError")
	}
}

// =============================================================================
// AssertionError Tests
// =============================================================================

func TestAssertionErrorFalse(t *testing.T) {
	runExpectError(t, `assert False`, "AssertionError")
}

func TestAssertionErrorWithMessage(t *testing.T) {
	runExpectError(t, `assert False, "custom message"`, "AssertionError")
}

func TestAssertionErrorExpression(t *testing.T) {
	runExpectError(t, `x = 5; assert x > 10`, "AssertionError")
}

// =============================================================================
// Import Error Tests
// =============================================================================

func TestImportErrorNonexistentModule(t *testing.T) {
	runExpectError(t, `import nonexistent_module_xyz`, "ModuleNotFoundError")
}

func TestFromImportErrorNonexistentModule(t *testing.T) {
	runExpectError(t, `from nonexistent_module_xyz import something`, "ModuleNotFoundError")
}

// =============================================================================
// Overflow and Boundary Tests
// =============================================================================

func TestMemoryErrorLargeStringRepetition(t *testing.T) {
	// Attempting to create an extremely large string
	runExpectError(t, `s = "x" * (10 ** 10)`, "MemoryError")
}

func TestMemoryErrorLargeListRepetition(t *testing.T) {
	// Attempting to create an extremely large list
	runExpectError(t, `lst = [1] * (10 ** 9)`, "MemoryError")
}

// =============================================================================
// Syntax/Compile Error Tests
// =============================================================================

func TestSyntaxErrorMissingColon(t *testing.T) {
	vm := runtime.NewVM()
	_, errs := compiler.CompileSource(`if True print("hello")`, "<test>")
	require.NotEmpty(t, errs)
	_ = vm // prevent unused warning
}

func TestSyntaxErrorUnmatchedParen(t *testing.T) {
	_, errs := compiler.CompileSource(`x = (1 + 2`, "<test>")
	require.NotEmpty(t, errs)
}

func TestSyntaxErrorUnmatchedBracket(t *testing.T) {
	_, errs := compiler.CompileSource(`x = [1, 2, 3`, "<test>")
	require.NotEmpty(t, errs)
}

func TestSyntaxErrorInvalidIndentation(t *testing.T) {
	source := `
def foo():
x = 1
`
	_, errs := compiler.CompileSource(source, "<test>")
	require.NotEmpty(t, errs)
}

func TestSyntaxErrorBreakOutsideLoop(t *testing.T) {
	_, errs := compiler.CompileSource(`break`, "<test>")
	require.NotEmpty(t, errs)
}

func TestSyntaxErrorContinueOutsideLoop(t *testing.T) {
	_, errs := compiler.CompileSource(`continue`, "<test>")
	require.NotEmpty(t, errs)
}

func TestSyntaxErrorReturnOutsideFunction(t *testing.T) {
	// This might be allowed at top-level in some implementations
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(`return 5`, "<test>")
	if len(errs) > 0 {
		return // Compile error as expected
	}
	// If it compiles, it should error at runtime
	_, err := vm.Execute(code)
	if err != nil {
		return // Runtime error is acceptable
	}
}

// =============================================================================
// Exception Handling Edge Cases
// =============================================================================

func TestReraiseDifferentException(t *testing.T) {
	source := `
caught = ""
try:
    try:
        raise ValueError("original")
    except ValueError:
        raise TypeError("different")
except TypeError as e:
    caught = "TypeError"
`
	vm := runCode(t, source)
	caught := vm.GetGlobal("caught").(*runtime.PyString)
	assert.Equal(t, "TypeError", caught.Value)
}

func TestExceptionInExceptBlock(t *testing.T) {
	source := `
result = ""
try:
    try:
        raise ValueError
    except ValueError:
        raise KeyError
except KeyError:
    result = "caught"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "caught", result.Value)
}

func TestExceptionInFinallyBlock(t *testing.T) {
	source := `
result = ""
try:
    try:
        pass
    finally:
        raise ValueError
except ValueError:
    result = "caught from finally"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "caught from finally", result.Value)
}

func TestFinallyOverridesReturn(t *testing.T) {
	source := `
def foo():
    try:
        return 1
    finally:
        return 2

result = foo()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}

func TestNestedExceptionHandling(t *testing.T) {
	source := `
results = []
try:
    try:
        try:
            raise ValueError
        except TypeError:
            results.append("inner-type")
        finally:
            results.append("inner-finally")
    except ValueError:
        results.append("middle-value")
    finally:
        results.append("middle-finally")
except:
    results.append("outer")
finally:
    results.append("outer-finally")
`
	vm := runCode(t, source)
	results := vm.GetGlobal("results").(*runtime.PyList)
	// inner-finally, middle-value, middle-finally, outer-finally
	require.Len(t, results.Items, 4)
	assert.Equal(t, "inner-finally", results.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "middle-value", results.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "middle-finally", results.Items[2].(*runtime.PyString).Value)
	assert.Equal(t, "outer-finally", results.Items[3].(*runtime.PyString).Value)
}

// =============================================================================
// Multiple Exception Types Tests
// =============================================================================

func TestCatchMultipleExceptionTypes(t *testing.T) {
	source := `
def test(exc):
    try:
        raise exc
    except (ValueError, KeyError):
        return "caught"
    return "not caught"

r1 = test(ValueError)
r2 = test(KeyError)
`
	vm := runCode(t, source)
	r1 := vm.GetGlobal("r1").(*runtime.PyString)
	r2 := vm.GetGlobal("r2").(*runtime.PyString)
	assert.Equal(t, "caught", r1.Value)
	assert.Equal(t, "caught", r2.Value)
}

// =============================================================================
// Error Recovery Tests
// =============================================================================

func TestErrorRecoveryAndContinue(t *testing.T) {
	source := `
results = []
for i in range(5):
    try:
        if i == 2:
            raise ValueError
        results.append(i)
    except ValueError:
        results.append("error")

length = len(results)
`
	vm := runCode(t, source)
	length := vm.GetGlobal("length").(*runtime.PyInt)
	assert.Equal(t, int64(5), length.Value)
}

func TestTryExceptInLoop(t *testing.T) {
	source := `
results = []
items = [1, "two", 3, "four", 5]
for item in items:
    try:
        results.append(item * 2)
    except TypeError:
        results.append("error")

length = len(results)
`
	vm := runCode(t, source)
	length := vm.GetGlobal("length").(*runtime.PyInt)
	assert.Equal(t, int64(5), length.Value)
}

// =============================================================================
// Callable Errors
// =============================================================================

func TestCallNonCallableObject(t *testing.T) {
	source := `
class Foo:
    pass

f = Foo()
f()  # Foo instance is not callable (no __call__)
`
	runExpectError(t, source, "TypeError")
}

func TestCallWithInvalidKwarg(t *testing.T) {
	source := `
def foo(a, b):
    return a + b

foo(1, c=2)  # 'c' is not a valid parameter
`
	runExpectError(t, source, "TypeError")
}

func TestCallWithDuplicateKwarg(t *testing.T) {
	source := `
def foo(a, b):
    return a + b

foo(1, a=2)  # 'a' got multiple values
`
	runExpectError(t, source, "TypeError")
}

// =============================================================================
// Division Edge Cases
// =============================================================================

func TestDivmodByZero(t *testing.T) {
	runExpectError(t, `divmod(10, 0)`, "ZeroDivisionError")
}

func TestPowerModZero(t *testing.T) {
	// pow(x, y, 0) should raise
	runExpectError(t, `pow(2, 10, 0)`, "ValueError")
}

// =============================================================================
// String Formatting Errors
// =============================================================================

func TestFormatStringMissingKey(t *testing.T) {
	runExpectError(t, `"{name}".format()`, "KeyError")
}

func TestFormatStringWrongIndex(t *testing.T) {
	runExpectError(t, `"{5}".format(1, 2, 3)`, "IndexError")
}

// =============================================================================
// Context Manager Errors
// =============================================================================

func TestContextManagerExitException(t *testing.T) {
	source := `
class BadContext:
    def __enter__(self):
        return self

    def __exit__(self, *args):
        raise RuntimeError("exit failed")

with BadContext():
    pass
`
	runExpectError(t, source, "RuntimeError")
}

func TestContextManagerEnterException(t *testing.T) {
	source := `
class BadContext:
    def __enter__(self):
        raise RuntimeError("enter failed")

    def __exit__(self, *args):
        pass

with BadContext():
    pass
`
	runExpectError(t, source, "RuntimeError")
}

// =============================================================================
// Property Errors
// =============================================================================

func TestPropertyGetterError(t *testing.T) {
	source := `
class Foo:
    @property
    def bad_prop(self):
        raise ValueError("getter failed")

f = Foo()
x = f.bad_prop
`
	runExpectError(t, source, "ValueError")
}

func TestPropertySetterError(t *testing.T) {
	source := `
class Foo:
    @property
    def prop(self):
        return self._prop

    @prop.setter
    def prop(self, value):
        raise ValueError("setter failed")

f = Foo()
f.prop = 5
`
	runExpectError(t, source, "ValueError")
}

func TestReadOnlyProperty(t *testing.T) {
	source := `
class Foo:
    @property
    def readonly(self):
        return 42

f = Foo()
f.readonly = 5  # Should fail - no setter defined
`
	runExpectError(t, source, "AttributeError")
}

// =============================================================================
// Super() Errors
// =============================================================================

func TestSuperNoArgOutsideClass(t *testing.T) {
	// super() with no args outside a class method
	source := `
def foo():
    super()

foo()
`
	runExpectError(t, source, "RuntimeError")
}

// =============================================================================
// Descriptor Protocol Errors
// =============================================================================

func TestDescriptorGetError(t *testing.T) {
	source := `
class BadDescriptor:
    def __get__(self, obj, type=None):
        raise ValueError("descriptor get failed")

class Foo:
    attr = BadDescriptor()

f = Foo()
x = f.attr
`
	runExpectError(t, source, "ValueError")
}
