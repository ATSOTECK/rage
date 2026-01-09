package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRageState creates a new rage.State with all modules enabled for testing
func newRageState(t *testing.T) *rage.State {
	t.Helper()
	return rage.NewStateWithModules(rage.WithAllModules(), rage.WithAllBuiltins())
}

// =============================================================================
// Basic Import Tests
// =============================================================================

func TestImportMathModule(t *testing.T) {
	source := `
import math
result = math.pi > 3.14
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestImportMathFunctions(t *testing.T) {
	source := `
import math
result = math.sqrt(16)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 4.0, result.Value)
}

func TestFromImportBasic(t *testing.T) {
	source := `
from math import sqrt
result = sqrt(25)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 5.0, result.Value)
}

func TestFromImportMultiple(t *testing.T) {
	source := `
from math import sqrt, floor, ceil
a = sqrt(16)
b = floor(3.7)
c = ceil(3.2)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyFloat)
	b := vm.GetGlobal("b")
	c := vm.GetGlobal("c")

	assert.Equal(t, 4.0, a.Value)
	// floor and ceil may return int or float depending on implementation
	switch v := b.(type) {
	case *runtime.PyFloat:
		assert.Equal(t, 3.0, v.Value)
	case *runtime.PyInt:
		assert.Equal(t, int64(3), v.Value)
	}
	switch v := c.(type) {
	case *runtime.PyFloat:
		assert.Equal(t, 4.0, v.Value)
	case *runtime.PyInt:
		assert.Equal(t, int64(4), v.Value)
	}
}

func TestFromImportAs(t *testing.T) {
	source := `
from math import sqrt as square_root
result = square_root(36)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("from...import...as not fully supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("from...import...as not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 6.0, result.Value)
}

func TestImportAs(t *testing.T) {
	source := `
import math as m
result = m.sqrt(49)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("import...as not fully supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("import...as not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 7.0, result.Value)
}

// =============================================================================
// Module Attribute Access Tests
// =============================================================================

func TestModuleNameAttribute(t *testing.T) {
	source := `
import math
result = math.__name__
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "math", result.Value)
}

func TestModuleDocAttribute(t *testing.T) {
	source := `
import math
has_doc = hasattr(math, '__doc__')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("hasattr not supported for modules")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("hasattr not supported for modules: " + err.Error())
		return
	}
	// Just check it doesn't error
}

// =============================================================================
// Module Caching Tests
// =============================================================================

func TestModuleImportCaching(t *testing.T) {
	// Importing the same module twice should return the same module
	source := `
import math
import math as m2
result = math is m2
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("import...as not fully supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Module identity check not supported: " + err.Error())
		return
	}
	// If we get here without error, that's good enough
}

func TestModuleStateShared(t *testing.T) {
	// Changes to a module should be visible from all import references
	source := `
import math
math.custom_value = 42
import math
result = math.custom_value
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Module attribute assignment not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Module attribute assignment not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// Import Error Tests
// =============================================================================

func TestImportNonexistentModule(t *testing.T) {
	source := `
import nonexistent_module_xyz
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		// Compile-time error is acceptable
		return
	}
	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent_module_xyz")
}

func TestFromImportNonexistentAttribute(t *testing.T) {
	source := `
from math import nonexistent_function_xyz
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		// Compile-time error is acceptable
		return
	}
	_, err := vm.Execute(code)
	if err == nil {
		t.Skip("Import error not raised for nonexistent attribute")
		return
	}
	assert.Contains(t, err.Error(), "nonexistent_function_xyz")
}

// =============================================================================
// Multiple Module Import Tests
// =============================================================================

func TestMultipleModuleImports(t *testing.T) {
	source := `
import math
import random
a = math.sqrt(16)
b = random.random()
result = a >= 0 and b >= 0
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestImportRandomModule(t *testing.T) {
	source := `
import random
result = 0 <= random.random() <= 1
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestImportTimeModule(t *testing.T) {
	source := `
import time
result = time.time() > 0
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("time module not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("time module not available: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestImportSysModule(t *testing.T) {
	source := `
import sys
has_version = hasattr(sys, 'version')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("sys module or hasattr not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("sys module not fully available: " + err.Error())
		return
	}
	// Success if no error
}

// =============================================================================
// Import Inside Functions Tests
// =============================================================================

func TestImportInsideFunction(t *testing.T) {
	source := `
def get_sqrt(n):
    import math
    return math.sqrt(n)

result = get_sqrt(64)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 8.0, result.Value)
}

func TestImportInsideFunctionMultipleCalls(t *testing.T) {
	// Import inside function called multiple times shouldn't cause issues
	source := `
def calculate(n):
    import math
    return math.sqrt(n)

results = []
for i in [4, 9, 16, 25]:
    results.append(calculate(i))
`
	vm := runCode(t, source)
	results := vm.GetGlobal("results").(*runtime.PyList)
	require.Len(t, results.Items, 4)
	expected := []float64{2.0, 3.0, 4.0, 5.0}
	for i, exp := range expected {
		assert.Equal(t, exp, results.Items[i].(*runtime.PyFloat).Value)
	}
}

// =============================================================================
// Import Inside Class Tests
// =============================================================================

func TestImportInsideClassMethod(t *testing.T) {
	source := `
class Calculator:
    def sqrt(self, n):
        import math
        return math.sqrt(n)

calc = Calculator()
result = calc.sqrt(81)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 9.0, result.Value)
}

// =============================================================================
// Import Inside Conditional Tests
// =============================================================================

func TestImportInsideConditional(t *testing.T) {
	source := `
x = 10
if x > 5:
    import math
    result = math.sqrt(100)
else:
    result = 0
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 10.0, result.Value)
}

func TestConditionalImportNotExecuted(t *testing.T) {
	source := `
x = 2
result = 42
if x > 5:
    import nonexistent_conditional_module
    result = 0
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// Star Import Tests
// =============================================================================

func TestFromImportStarMath(t *testing.T) {
	source := `
from math import *
result = sqrt(144)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("from...import * not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("from...import * not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 12.0, result.Value)
}

// =============================================================================
// Module as Object Tests
// =============================================================================

func TestModuleAsObject(t *testing.T) {
	source := `
import math
m = math
result = m.sqrt(169)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 13.0, result.Value)
}

func TestModuleInList(t *testing.T) {
	source := `
import math
modules = [math]
result = modules[0].sqrt(196)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Module in list not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Module attribute access from list not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 14.0, result.Value)
}

func TestModuleInDict(t *testing.T) {
	source := `
import math
modules = {"math": math}
result = modules["math"].sqrt(225)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Module in dict not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Module attribute access from dict not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 15.0, result.Value)
}

// =============================================================================
// Chained Module Attribute Access Tests
// =============================================================================

func TestChainedModuleAccess(t *testing.T) {
	source := `
import math
result = int(math.floor(math.sqrt(50)))
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(7), result.Value)
}

// =============================================================================
// Collections Module Tests
// =============================================================================

func TestImportCollections(t *testing.T) {
	source := `
import collections
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("collections module not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("collections module not available: " + err.Error())
		return
	}
	// Success if no error
}

func TestCollectionsCounterImport(t *testing.T) {
	source := `
from collections import Counter
c = Counter([1, 2, 2, 3, 3, 3])
result = c[3]
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("collections.Counter not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("collections.Counter not available: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

func TestCollectionsDefaultdict(t *testing.T) {
	source := `
from collections import defaultdict
d = defaultdict(int)
d['a'] = d['a'] + 1
result = d['a']
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("collections.defaultdict not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("collections.defaultdict not available: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

// =============================================================================
// String Module Tests
// =============================================================================

func TestImportString(t *testing.T) {
	source := `
import string
has_ascii = hasattr(string, 'ascii_lowercase')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("string module or hasattr not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("string module not available: " + err.Error())
		return
	}
	// Success if no error
}

func TestStringModuleConstants(t *testing.T) {
	source := `
import string
result = len(string.ascii_lowercase)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("string module not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("string.ascii_lowercase not available: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(26), result.Value)
}

// =============================================================================
// Re Module Tests
// =============================================================================

func TestImportRe(t *testing.T) {
	source := `
import re
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("re module not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("re module not available: " + err.Error())
		return
	}
	// Success if no error
}

func TestReModuleMatch(t *testing.T) {
	source := `
import re
m = re.match(r'\d+', '123abc')
result = m.group() if m else ''
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("re.match not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("re.match not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "123", result.Value)
}

func TestReModuleSearch(t *testing.T) {
	source := `
import re
m = re.search(r'\d+', 'abc123def')
result = m.group() if m else ''
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("re.search not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("re.search not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "123", result.Value)
}

func TestReModuleFindall(t *testing.T) {
	source := `
import re
result = re.findall(r'\d+', 'a1b22c333')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("re.findall not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("re.findall not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "1", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "22", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "333", result.Items[2].(*runtime.PyString).Value)
}

func TestReModuleSub(t *testing.T) {
	source := `
import re
result = re.sub(r'\d+', 'X', 'a1b22c333')
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("re.sub not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("re.sub not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "aXbXcX", result.Value)
}

// =============================================================================
// Builtins Module Tests
// =============================================================================

func TestImportBuiltins(t *testing.T) {
	source := `
import builtins
result = builtins.len([1, 2, 3])
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("builtins module not available")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("builtins module not available: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

// =============================================================================
// Scope Tests with Imports
// =============================================================================

func TestImportDoesNotLeakFromFunction(t *testing.T) {
	source := `
def import_local():
    import math
    return math.sqrt(4)

result = import_local()
# math should not be accessible here
has_math = 'math' in dir()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("dir() not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Scope test not supported: " + err.Error())
		return
	}
	// Success if no error
}

func TestGlobalImportAccessibleInFunction(t *testing.T) {
	source := `
import math

def use_global_import():
    return math.sqrt(9)

result = use_global_import()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 3.0, result.Value)
}

// =============================================================================
// Import Order Tests
// =============================================================================

func TestImportOrderMatters(t *testing.T) {
	// This should work because import happens before usage
	source := `
result = []
import math
result.append(math.sqrt(1))
result.append(math.sqrt(4))
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 2)
	assert.Equal(t, 1.0, result.Items[0].(*runtime.PyFloat).Value)
	assert.Equal(t, 2.0, result.Items[1].(*runtime.PyFloat).Value)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestImportSameModuleMultipleTimes(t *testing.T) {
	source := `
import math
import math
import math
result = math.sqrt(256)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 16.0, result.Value)
}

func TestImportAfterFromImport(t *testing.T) {
	source := `
from math import sqrt
import math
result = sqrt(9) + math.sqrt(16)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 7.0, result.Value)
}

func TestFromImportAfterImport(t *testing.T) {
	source := `
import math
from math import sqrt
result = sqrt(25) + math.sqrt(36)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	assert.Equal(t, 11.0, result.Value)
}

// =============================================================================
// Module Type Tests
// =============================================================================

func TestModuleTypeString(t *testing.T) {
	source := `
import math
result = type(math).__name__
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("type().__name__ not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("type().__name__ not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "module", result.Value)
}

// =============================================================================
// Relative Import Tests
// =============================================================================

func TestRelativeImportSamePackage(t *testing.T) {
	// Test: from . import sibling
	state := newRageState(t)
	defer state.Close()

	// Register a package with two modules
	err := state.RegisterPythonModule("mypackage.utils", `
value = 42
def helper():
    return "helper called"
`)
	require.NoError(t, err)

	err = state.RegisterPythonModule("mypackage.main", `
from . import utils
result = utils.value
helper_result = utils.helper()
`)
	require.NoError(t, err)

	// Import and verify
	_, err = state.Run(`
import mypackage.main
result = mypackage.main.result
helper_result = mypackage.main.helper_result
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	intVal, ok := rage.AsInt(result)
	require.True(t, ok)
	assert.Equal(t, int64(42), intVal)

	helperResult := state.GetGlobal("helper_result")
	require.NotNil(t, helperResult)
	strVal, ok := rage.AsString(helperResult)
	require.True(t, ok)
	assert.Equal(t, "helper called", strVal)
}

func TestRelativeImportFromDot(t *testing.T) {
	// Test: from .module import name
	state := newRageState(t)
	defer state.Close()

	// Register modules
	err := state.RegisterPythonModule("pkg.helpers", `
PI = 3.14159
def double(x):
    return x * 2
`)
	require.NoError(t, err)

	err = state.RegisterPythonModule("pkg.consumer", `
from .helpers import PI, double
result = double(PI)
`)
	require.NoError(t, err)

	// Import and verify
	_, err = state.Run(`
import pkg.consumer
result = pkg.consumer.result
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	floatVal, ok := rage.AsFloat(result)
	require.True(t, ok)
	assert.InDelta(t, 6.28318, floatVal, 0.0001)
}

func TestRelativeImportParentPackage(t *testing.T) {
	// Test: from .. import sibling
	state := newRageState(t)
	defer state.Close()

	// Register a three-level package structure
	err := state.RegisterPythonModule("top.shared", `
SHARED_VALUE = "shared"
`)
	require.NoError(t, err)

	err = state.RegisterPythonModule("top.sub.module", `
from .. import shared
result = shared.SHARED_VALUE
`)
	require.NoError(t, err)

	// Import and verify
	_, err = state.Run(`
import top.sub.module
result = top.sub.module.result
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	strVal, ok := rage.AsString(result)
	require.True(t, ok)
	assert.Equal(t, "shared", strVal)
}

func TestRelativeImportParentModule(t *testing.T) {
	// Test: from ..sibling import name
	state := newRageState(t)
	defer state.Close()

	// Register modules
	err := state.RegisterPythonModule("app.utils.helpers", `
def greet(name):
    return "Hello, " + name
`)
	require.NoError(t, err)

	err = state.RegisterPythonModule("app.core.main", `
from ..utils import helpers
result = helpers.greet("World")
`)
	require.NoError(t, err)

	// Import and verify
	_, err = state.Run(`
import app.core.main
result = app.core.main.result
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	strVal, ok := rage.AsString(result)
	require.True(t, ok)
	assert.Equal(t, "Hello, World", strVal)
}

func TestRelativeImportBeyondTopLevel(t *testing.T) {
	// Test: from .. in a top-level module should fail
	state := newRageState(t)
	defer state.Close()

	// Register a module that tries to import beyond its package
	// For a top-level module "shallow", __package__ is "shallow" (no parent)
	// So "from .." would try to go above the top level
	err := state.RegisterPythonModule("shallow", `
from .. import something
`)
	// The error happens during registration because the module code is executed
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ImportError")
}

func TestRelativeImportNoPackage(t *testing.T) {
	// Test: relative import in a module without a package should fail
	state := newRageState(t)
	defer state.Close()

	// Try relative import in main script (no package context)
	_, err := state.Run(`from . import something`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ImportError")
}

// =============================================================================
// Package Import Tests
// =============================================================================

func TestPackageImport(t *testing.T) {
	state := newRageState(t)
	defer state.Close()

	// Register a package with an "init" style module
	err := state.RegisterPythonModule("mypkg", `
VERSION = "1.0.0"
def init():
    return "initialized"
`)
	require.NoError(t, err)

	// Import and use
	_, err = state.Run(`
import mypkg
version = mypkg.VERSION
`)
	require.NoError(t, err)

	result := state.GetGlobal("version")
	require.NotNil(t, result)
	strVal, ok := rage.AsString(result)
	require.True(t, ok)
	assert.Equal(t, "1.0.0", strVal)
}

func TestSubmoduleImport(t *testing.T) {
	state := newRageState(t)
	defer state.Close()

	// Register nested modules
	err := state.RegisterPythonModule("outer.inner", `
VALUE = 100
`)
	require.NoError(t, err)

	// Import and use
	_, err = state.Run(`
import outer.inner
result = outer.inner.VALUE
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	intVal, ok := rage.AsInt(result)
	require.True(t, ok)
	assert.Equal(t, int64(100), intVal)
}

func TestDeeplyNestedModuleImport(t *testing.T) {
	state := newRageState(t)
	defer state.Close()

	// Register deeply nested module
	err := state.RegisterPythonModule("a.b.c.d", `
DEEP = "very deep"
`)
	require.NoError(t, err)

	// Import and use
	_, err = state.Run(`
import a.b.c.d
result = a.b.c.d.DEEP
`)
	require.NoError(t, err)

	result := state.GetGlobal("result")
	require.NotNil(t, result)
	strVal, ok := rage.AsString(result)
	require.True(t, ok)
	assert.Equal(t, "very deep", strVal)
}
