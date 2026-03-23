package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Coroutine exception state isolation
// =============================================================================

func TestCoroutineExceptionStateIsolation(t *testing.T) {
	// A coroutine handling exceptions internally should not corrupt the
	// caller's exception state
	source := `
import asyncio

async def might_fail():
    try:
        raise ValueError("inner")
    except ValueError:
        pass
    return "ok"

result = asyncio.run(might_fail())
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "ok", result.Value)
}

func TestCoroutineExceptionInHandler(t *testing.T) {
	// Running a coroutine inside an except handler should not clobber
	// the handler's exception state
	source := `
import asyncio

async def inner():
    return 42

caught = False
try:
    raise ValueError("outer")
except ValueError:
    result = asyncio.run(inner())
    caught = True
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
	caught := vm.GetGlobal("caught")
	assert.Equal(t, runtime.True, caught)
}

func TestCoroutineTryFinally(t *testing.T) {
	// try/finally inside a coroutine should work correctly
	source := `
import asyncio

async def with_finally():
    result = []
    try:
        result.append("try")
        return result
    finally:
        result.append("finally")

result = asyncio.run(with_finally())
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "try", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "finally", result.Items[1].(*runtime.PyString).Value)
}

func TestCoroutineNestedExceptionHandling(t *testing.T) {
	// Nested coroutines with exception handling should not corrupt state
	source := `
import asyncio

async def inner():
    try:
        raise KeyError("k")
    except KeyError:
        return "caught_inner"

async def outer():
    try:
        r = await inner()
        return r
    except Exception:
        return "wrong"

result = asyncio.run(outer())
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "caught_inner", result.Value)
}

// =============================================================================
// Coroutine close
// =============================================================================

func TestCoroutineCloseRunsFinally(t *testing.T) {
	// A coroutine with try/finally runs finally when completing normally
	// Note: RAGE has no event loop, so coroutines can't be suspended
	// mid-execution. The finally block runs during asyncio.run().
	source := `
import asyncio

finally_ran = False

async def coro():
    global finally_ran
    try:
        x = 1
    finally:
        finally_ran = True
    return x

result = asyncio.run(coro())
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("finally_ran")
	assert.Equal(t, runtime.True, result)
	intResult := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), intResult.Value)
}

func TestCoroutineCloseAlreadyClosed(t *testing.T) {
	// Closing an already-closed coroutine should be a no-op
	source := `
import asyncio

async def simple():
    return 1

c = simple()
result = asyncio.run(c)
c.close()  # should not error
c.close()  # double close should also not error
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestCoroutineCloseNotStarted(t *testing.T) {
	// Closing a not-yet-started coroutine should be safe
	source := `
import asyncio

async def never_started():
    return 1

c = never_started()
c.close()
closed_ok = True
`
	vm := runCodeWithStdlib(t, source)
	result := vm.GetGlobal("closed_ok")
	assert.Equal(t, runtime.True, result)
}
