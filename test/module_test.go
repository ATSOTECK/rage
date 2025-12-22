package test

import (
	"math"
	"testing"

	"github.com/ATSOTECK/RAGE/internal/compiler"
	"github.com/ATSOTECK/RAGE/internal/runtime"
	"github.com/ATSOTECK/RAGE/internal/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Initialize standard library modules for testing
	stdlib.InitAllModules()
}

// TestImportModule tests basic module import
func TestImportModule(t *testing.T) {
	// Reset modules for clean test
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import math
result = math.pi
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	require.NotNil(t, result)
	assert.InDelta(t, math.Pi, result.(*runtime.PyFloat).Value, 0.0001)
}

// TestFromImport tests "from module import name"
func TestFromImport(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from math import pi, e
result_pi = pi
result_e = e
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	resultPi := vm.GetGlobal("result_pi")
	resultE := vm.GetGlobal("result_e")
	assert.InDelta(t, math.Pi, resultPi.(*runtime.PyFloat).Value, 0.0001)
	assert.InDelta(t, math.E, resultE.(*runtime.PyFloat).Value, 0.0001)
}

// TestFromImportStar tests "from module import *"
func TestFromImportStar(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from math import *
result = pi + e
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	expected := math.Pi + math.E
	assert.InDelta(t, expected, result.(*runtime.PyFloat).Value, 0.0001)
}

// TestImportModuleFunction tests calling functions from imported module
func TestImportModuleFunction(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import math
result = math.sqrt(16)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.InDelta(t, 4.0, result.(*runtime.PyFloat).Value, 0.0001)
}

// TestFromImportFunction tests calling function imported with "from"
func TestFromImportFunction(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from math import sqrt, cos
result_sqrt = sqrt(25)
result_cos = cos(0)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	resultSqrt := vm.GetGlobal("result_sqrt")
	resultCos := vm.GetGlobal("result_cos")
	assert.InDelta(t, 5.0, resultSqrt.(*runtime.PyFloat).Value, 0.0001)
	assert.InDelta(t, 1.0, resultCos.(*runtime.PyFloat).Value, 0.0001)
}

// TestMathModuleFunctions tests various math module functions
func TestMathModuleFunctions(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	tests := []struct {
		name     string
		source   string
		expected float64
	}{
		{"sqrt", "from math import sqrt\nresult = sqrt(9)", 3.0},
		{"pow", "from math import pow\nresult = pow(2, 3)", 8.0},
		{"sin_0", "from math import sin\nresult = sin(0)", 0.0},
		{"cos_0", "from math import cos\nresult = cos(0)", 1.0},
		{"ceil", "from math import ceil\nresult = ceil(3.2)", 4.0},
		{"floor", "from math import floor\nresult = floor(3.8)", 3.0},
		{"fabs", "from math import fabs\nresult = fabs(-5.5)", 5.5},
		{"log", "from math import log, e\nresult = log(e)", 1.0},
		{"degrees", "from math import degrees, pi\nresult = degrees(pi)", 180.0},
		{"radians", "from math import radians\nresult = radians(180)", math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.ResetModules()
			stdlib.InitAllModules()

			vm := runtime.NewVM()
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			_, err := vm.Execute(code)
			require.NoError(t, err)

			result := vm.GetGlobal("result")
			assert.InDelta(t, tt.expected, result.(*runtime.PyFloat).Value, 0.0001)
		})
	}
}

// TestCustomModuleRegistration tests registering a custom module from Go
func TestCustomModuleRegistration(t *testing.T) {
	runtime.ResetModules()

	// Register a custom module
	runtime.NewModuleBuilder("mymodule").
		Doc("My custom module").
		Const("VERSION", runtime.NewString("1.0.0")).
		Const("MAGIC_NUMBER", runtime.NewInt(42)).
		Func("greet", func(vm *runtime.VM) int {
			name := vm.CheckString(1)
			vm.Push(runtime.NewString("Hello, " + name + "!"))
			return 1
		}).
		Func("add", func(vm *runtime.VM) int {
			a := vm.CheckInt(1)
			b := vm.CheckInt(2)
			vm.Push(runtime.NewInt(a + b))
			return 1
		}).
		Register()

	vm := runtime.NewVM()

	source := `
import mymodule
version = mymodule.VERSION
magic = mymodule.MAGIC_NUMBER
greeting = mymodule.greet("World")
sum_result = mymodule.add(10, 32)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", vm.GetGlobal("version").(*runtime.PyString).Value)
	assert.Equal(t, int64(42), vm.GetGlobal("magic").(*runtime.PyInt).Value)
	assert.Equal(t, "Hello, World!", vm.GetGlobal("greeting").(*runtime.PyString).Value)
	assert.Equal(t, int64(42), vm.GetGlobal("sum_result").(*runtime.PyInt).Value)
}

// TestCustomModuleFromImport tests "from custom_module import ..."
func TestCustomModuleFromImport(t *testing.T) {
	runtime.ResetModules()

	// Register a custom module
	runtime.NewModuleBuilder("constants").
		Const("PI", runtime.NewFloat(3.14159)).
		Const("E", runtime.NewFloat(2.71828)).
		Const("GOLDEN_RATIO", runtime.NewFloat(1.61803)).
		Register()

	vm := runtime.NewVM()

	source := `
from constants import PI, GOLDEN_RATIO
result = PI + GOLDEN_RATIO
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	expected := 3.14159 + 1.61803
	assert.InDelta(t, expected, result.(*runtime.PyFloat).Value, 0.0001)
}

// TestModuleNotFound tests error handling for non-existent modules
func TestModuleNotFound(t *testing.T) {
	runtime.ResetModules()

	vm := runtime.NewVM()

	source := `import nonexistent_module`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No module named 'nonexistent_module'")
}

// TestImportNameNotFound tests error handling for non-existent names in module
func TestImportNameNotFound(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `from math import nonexistent_function`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot import name 'nonexistent_function'")
}

// TestModuleWithType tests registering a module with a custom type
func TestModuleWithType(t *testing.T) {
	runtime.ResetModules()

	type Counter struct {
		Value int
	}

	// Register a module with a type
	runtime.NewModuleBuilder("counter").
		Type("Counter",
			// Constructor
			func(vm *runtime.VM) int {
				initial := int(vm.ToInt(1)) // Optional initial value
				ud := vm.NewUserDataWithMeta(&Counter{Value: initial}, "Counter")
				vm.Push(ud)
				return 1
			},
			// Methods
			map[string]runtime.GoFunction{
				"get": func(vm *runtime.VM) int {
					ud := vm.CheckUserData(1, "Counter")
					counter := ud.Value.(*Counter)
					vm.Push(runtime.NewInt(int64(counter.Value)))
					return 1
				},
				"increment": func(vm *runtime.VM) int {
					ud := vm.CheckUserData(1, "Counter")
					counter := ud.Value.(*Counter)
					counter.Value++
					return 0
				},
				"add": func(vm *runtime.VM) int {
					ud := vm.CheckUserData(1, "Counter")
					n := int(vm.CheckInt(2))
					counter := ud.Value.(*Counter)
					counter.Value += n
					return 0
				},
			},
		).
		Register()

	vm := runtime.NewVM()

	source := `
from counter import Counter
c = Counter(10)
initial = c.get()
c.increment()
after_inc = c.get()
c.add(5)
final = c.get()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, int64(10), vm.GetGlobal("initial").(*runtime.PyInt).Value)
	assert.Equal(t, int64(11), vm.GetGlobal("after_inc").(*runtime.PyInt).Value)
	assert.Equal(t, int64(16), vm.GetGlobal("final").(*runtime.PyInt).Value)
}

// TestImportAlias tests "import module as alias"
func TestImportAlias(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
import math as m
result = m.sqrt(16)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.InDelta(t, 4.0, result.(*runtime.PyFloat).Value, 0.0001)
}

// TestFromImportAlias tests "from module import name as alias"
func TestFromImportAlias(t *testing.T) {
	runtime.ResetModules()
	stdlib.InitAllModules()

	vm := runtime.NewVM()

	source := `
from math import sqrt as square_root
result = square_root(25)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.InDelta(t, 5.0, result.(*runtime.PyFloat).Value, 0.0001)
}
