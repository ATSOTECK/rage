package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicTryExcept tests basic try/except functionality
func TestBasicTryExcept(t *testing.T) {
	vm := runtime.NewVM()

	source := `
result = "not caught"
try:
    raise ValueError
except ValueError:
    result = "caught"
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "caught", result.(*runtime.PyString).Value)
}

// TestExceptionWithMessage tests exceptions with messages
func TestExceptionWithMessage(t *testing.T) {
	vm := runtime.NewVM()

	source := `
msg = ""
try:
    raise ValueError
except ValueError as e:
    msg = "got it"
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("msg")
	assert.Equal(t, "got it", result.(*runtime.PyString).Value)
}

// TestMultipleExceptClauses tests multiple except clauses
func TestMultipleExceptClauses(t *testing.T) {
	vm := runtime.NewVM()

	source := `
result = ""
try:
    raise KeyError
except ValueError:
    result = "value"
except KeyError:
    result = "key"
except TypeError:
    result = "type"
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "key", result.(*runtime.PyString).Value)
}

// TestBareExcept tests bare except clause
func TestBareExcept(t *testing.T) {
	vm := runtime.NewVM()

	source := `
result = ""
try:
    raise RuntimeError
except:
    result = "caught"
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "caught", result.(*runtime.PyString).Value)
}

// TestFinallyBlock tests finally block execution
func TestFinallyBlock(t *testing.T) {
	vm := runtime.NewVM()

	source := `
finally_ran = False
try:
    x = 1
finally:
    finally_ran = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("finally_ran")
	assert.Equal(t, runtime.True, result)
}

// TestFinallyWithException tests finally runs even with exception
func TestFinallyWithException(t *testing.T) {
	vm := runtime.NewVM()

	source := `
finally_ran = False
caught = False
try:
    try:
        raise ValueError
    finally:
        finally_ran = True
except ValueError:
    caught = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	finallyRan := vm.GetGlobal("finally_ran")
	caught := vm.GetGlobal("caught")
	assert.Equal(t, runtime.True, finallyRan)
	assert.Equal(t, runtime.True, caught)
}

// TestElseClause tests else clause runs when no exception
func TestElseClause(t *testing.T) {
	vm := runtime.NewVM()

	source := `
else_ran = False
try:
    x = 1
except ValueError:
    pass
else:
    else_ran = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("else_ran")
	assert.Equal(t, runtime.True, result)
}

// TestElseNotRunOnException tests else doesn't run when exception occurs
func TestElseNotRunOnException(t *testing.T) {
	vm := runtime.NewVM()

	source := `
else_ran = False
try:
    raise ValueError
except ValueError:
    pass
else:
    else_ran = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("else_ran")
	assert.Equal(t, runtime.False, result)
}

// TestReRaise tests bare raise statement
func TestReRaise(t *testing.T) {
	vm := runtime.NewVM()

	source := `
caught_outer = False
try:
    try:
        raise ValueError
    except ValueError:
        raise
except ValueError:
    caught_outer = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("caught_outer")
	assert.Equal(t, runtime.True, result)
}

// TestExceptionPropagates tests uncaught exception propagates
func TestExceptionPropagates(t *testing.T) {
	vm := runtime.NewVM()

	source := `
raise ValueError
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "ValueError"))
}

// TestExceptionInheritance tests exception class inheritance
func TestExceptionInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
caught = False
try:
    raise ValueError
except Exception:
    caught = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("caught")
	assert.Equal(t, runtime.True, result)
}

// TestNestedTryExcept tests nested try/except blocks
func TestNestedTryExcept(t *testing.T) {
	vm := runtime.NewVM()

	source := `
inner_caught = False
outer_caught = False
try:
    try:
        raise KeyError
    except ValueError:
        inner_caught = True
except KeyError:
    outer_caught = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	innerCaught := vm.GetGlobal("inner_caught")
	outerCaught := vm.GetGlobal("outer_caught")
	assert.Equal(t, runtime.False, innerCaught)
	assert.Equal(t, runtime.True, outerCaught)
}

// TestExceptionClasses tests that all standard exception classes exist
func TestExceptionClasses(t *testing.T) {
	vm := runtime.NewVM()

	exceptions := []string{
		"BaseException", "Exception", "ValueError", "TypeError",
		"KeyError", "IndexError", "AttributeError", "NameError",
		"RuntimeError", "ZeroDivisionError", "AssertionError",
		"StopIteration", "NotImplementedError", "ImportError",
		"ModuleNotFoundError", "OSError", "FileNotFoundError",
	}

	for _, exc := range exceptions {
		source := `
caught = False
try:
    raise ` + exc + `
except ` + exc + `:
    caught = True
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs, "Failed to compile for %s", exc)

		_, err := vm.Execute(code)
		require.NoError(t, err, "Failed to execute for %s", exc)

		result := vm.GetGlobal("caught")
		assert.Equal(t, runtime.True, result, "Failed to catch %s", exc)
	}
}

// TestTryExceptFinally tests try/except/finally together
func TestTryExceptFinally(t *testing.T) {
	vm := runtime.NewVM()

	source := `
caught = False
finally_ran = False
try:
    raise ValueError
except ValueError:
    caught = True
finally:
    finally_ran = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	caught := vm.GetGlobal("caught")
	finallyRan := vm.GetGlobal("finally_ran")
	assert.Equal(t, runtime.True, caught)
	assert.Equal(t, runtime.True, finallyRan)
}

// TestTryElseFinally tests try/else/finally together
func TestTryElseFinally(t *testing.T) {
	vm := runtime.NewVM()

	source := `
else_ran = False
finally_ran = False
try:
    x = 1
except ValueError:
    pass
else:
    else_ran = True
finally:
    finally_ran = True
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	elseRan := vm.GetGlobal("else_ran")
	finallyRan := vm.GetGlobal("finally_ran")
	assert.Equal(t, runtime.True, elseRan)
	assert.Equal(t, runtime.True, finallyRan)
}
