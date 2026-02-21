package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/internal/stdlib"
	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Recursion Depth Limit Tests
// =============================================================================

func TestResourceLimitsRecursionNormalPasses(t *testing.T) {
	source := `
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

result = recurse(50)
`
	vm := runtime.NewVM()
	vm.SetMaxRecursionDepth(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(50), result.Value)
}

func TestResourceLimitsRecursionExceedsLimit(t *testing.T) {
	source := `
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

result = recurse(200)
`
	vm := runtime.NewVM()
	vm.SetMaxRecursionDepth(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RecursionError")
}

func TestResourceLimitsRecursionZeroUnlimited(t *testing.T) {
	source := `
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

result = recurse(200)
`
	vm := runtime.NewVM()
	// Default is 0 = unlimited
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(200), result.Value)
}

func TestResourceLimitsMutualRecursionHitsLimit(t *testing.T) {
	source := `
def ping(n):
    if n <= 0:
        return "done"
    return pong(n - 1)

def pong(n):
    if n <= 0:
        return "done"
    return ping(n - 1)

result = ping(200)
`
	vm := runtime.NewVM()
	vm.SetMaxRecursionDepth(50)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RecursionError")
}

func TestResourceLimitsRecursionErrorCatchable(t *testing.T) {
	source := `
def infinite():
    return infinite()

caught = False
try:
    infinite()
except RecursionError:
    caught = True
`
	vm := runtime.NewVM()
	vm.SetMaxRecursionDepth(50)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	caught := vm.GetGlobal("caught").(*runtime.PyBool)
	assert.True(t, caught.Value)
}

// =============================================================================
// Memory Limit Tests
// =============================================================================

func TestResourceLimitsMemoryFrameAllocation(t *testing.T) {
	source := `
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

result = recurse(500)
`
	vm := runtime.NewVM()
	vm.SetMaxMemoryBytes(4096) // Very small
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
}

func TestResourceLimitsMemoryStringRepetition(t *testing.T) {
	source := `
s = "x" * 1000000
`
	vm := runtime.NewVM()
	vm.SetMaxMemoryBytes(50000)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
}

func TestResourceLimitsMemoryZeroUnlimited(t *testing.T) {
	source := `
s = "hello " * 100
result = len(s)
`
	vm := runtime.NewVM()
	// Default is 0 = unlimited
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(600), result.Value)
}

// =============================================================================
// Collection Size Limit Tests
// =============================================================================

func TestResourceLimitsCollectionListAppend(t *testing.T) {
	source := `
lst = []
for i in range(200):
    lst.append(i)
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
	assert.Contains(t, err.Error(), "list")
}

func TestResourceLimitsCollectionDictInsertion(t *testing.T) {
	source := `
d = {}
for i in range(200):
    d[i] = i
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
	assert.Contains(t, err.Error(), "dict")
}

func TestResourceLimitsCollectionSetAdd(t *testing.T) {
	source := `
s = set()
for i in range(200):
    s.add(i)
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
	assert.Contains(t, err.Error(), "set")
}

func TestResourceLimitsCollectionListComprehension(t *testing.T) {
	source := `
result = [i for i in range(200)]
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
	assert.Contains(t, err.Error(), "list")
}

func TestResourceLimitsCollectionListRepetition(t *testing.T) {
	source := `
result = [0] * 200
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MemoryError")
	assert.Contains(t, err.Error(), "list")
}

func TestResourceLimitsCollectionZeroUnlimited(t *testing.T) {
	source := `
lst = [i for i in range(500)]
result = len(lst)
`
	vm := runtime.NewVM()
	// Default is 0 = unlimited
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(500), result.Value)
}

func TestResourceLimitsCollectionUnderLimit(t *testing.T) {
	source := `
lst = [i for i in range(50)]
result = len(lst)
`
	vm := runtime.NewVM()
	vm.SetMaxCollectionSize(100)
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(50), result.Value)
}

// =============================================================================
// sys.setrecursionlimit Sync Test
// =============================================================================

func TestResourceLimitsSysSetRecursionLimit(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitSysModule()

	source := `
import sys
sys.setrecursionlimit(30)

def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

# Should fail since we exceed the limit
try:
    recurse(100)
    caught = False
except RecursionError:
    caught = True

# Reset to default so we don't pollute other tests
sys.setrecursionlimit(1000)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)
	caught := vm.GetGlobal("caught").(*runtime.PyBool)
	assert.True(t, caught.Value)
}

// =============================================================================
// Public API StateOption Tests
// =============================================================================

func TestResourceLimitsStateOptionRecursion(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithAllModules(),
		rage.WithMaxRecursionDepth(50),
	)
	defer state.Close()

	_, err := state.Run(`
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

result = recurse(100)
`)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "RecursionError"))
}

func TestResourceLimitsStateOptionMemory(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithAllModules(),
		rage.WithMaxMemoryBytes(1024),
	)
	defer state.Close()

	_, err := state.Run(`s = "x" * 100000`)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "MemoryError"))
}

func TestResourceLimitsStateOptionCollection(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithAllModules(),
		rage.WithMaxCollectionSize(50),
	)
	defer state.Close()

	_, err := state.Run(`lst = [i for i in range(100)]`)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "MemoryError"))
}

func TestResourceLimitsStateSetters(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetMaxRecursionDepth(30)
	_, err := state.Run(`
def recurse(n):
    if n <= 0:
        return 0
    return recurse(n - 1) + 1

try:
    recurse(100)
    caught = False
except RecursionError:
    caught = True
`)
	require.NoError(t, err)

	caught := state.GetGlobal("caught")
	require.NotNil(t, caught)
}

func TestResourceLimitsAllocatedBytes(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithMaxMemoryBytes(1024 * 1024), // 1MB
	)
	defer state.Close()

	_, err := state.Run(`x = "hello " * 10`)
	require.NoError(t, err)

	// Should have tracked some bytes
	assert.Greater(t, state.AllocatedBytes(), int64(0))
}
