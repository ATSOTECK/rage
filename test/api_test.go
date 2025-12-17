package test

import (
	"testing"

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoFunctionRegistration tests registering Go functions that can be called from Python
func TestGoFunctionRegistration(t *testing.T) {
	vm := runtime.NewVM()

	// Register a simple Go function that doubles a number
	vm.Register("double", func(vm *runtime.VM) int {
		n := vm.CheckInt(1)
		vm.Push(runtime.NewInt(n * 2))
		return 1
	})

	source := `result = double(21)`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, int64(42), result.(*runtime.PyInt).Value)
}

// TestGoFunctionMultipleArgs tests Go functions with multiple arguments
func TestGoFunctionMultipleArgs(t *testing.T) {
	vm := runtime.NewVM()

	// Register a function that adds two numbers
	vm.Register("add", func(vm *runtime.VM) int {
		a := vm.CheckInt(1)
		b := vm.CheckInt(2)
		vm.Push(runtime.NewInt(a + b))
		return 1
	})

	source := `result = add(10, 32)`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, int64(42), result.(*runtime.PyInt).Value)
}

// TestGoFunctionMultipleReturns tests Go functions returning multiple values
func TestGoFunctionMultipleReturns(t *testing.T) {
	vm := runtime.NewVM()

	// Register a function that returns multiple values
	vm.Register("divmod", func(vm *runtime.VM) int {
		a := vm.CheckInt(1)
		b := vm.CheckInt(2)
		vm.Push(runtime.NewInt(a / b))
		vm.Push(runtime.NewInt(a % b))
		return 2
	})

	source := `
q, r = divmod(17, 5)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	q := vm.GetGlobal("q")
	r := vm.GetGlobal("r")
	assert.Equal(t, int64(3), q.(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), r.(*runtime.PyInt).Value)
}

// TestGoFunctionStringArgs tests Go functions with string arguments
func TestGoFunctionStringArgs(t *testing.T) {
	vm := runtime.NewVM()

	// Register a function that concatenates strings
	vm.Register("greet", func(vm *runtime.VM) int {
		name := vm.CheckString(1)
		vm.Push(runtime.NewString("Hello, " + name + "!"))
		return 1
	})

	source := `result = greet("World")`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "Hello, World!", result.(*runtime.PyString).Value)
}

// TestSetGetGlobal tests SetGlobal and GetGlobal
func TestSetGetGlobal(t *testing.T) {
	vm := runtime.NewVM()

	// Set some global values from Go
	vm.SetGlobal("my_int", runtime.NewInt(42))
	vm.SetGlobal("my_str", runtime.NewString("hello"))
	vm.SetGlobal("my_float", runtime.NewFloat(3.14))

	source := `
result_int = my_int * 2
result_str = my_str + " world"
result_float = my_float * 2
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, int64(84), vm.GetGlobal("result_int").(*runtime.PyInt).Value)
	assert.Equal(t, "hello world", vm.GetGlobal("result_str").(*runtime.PyString).Value)
	assert.Equal(t, 6.28, vm.GetGlobal("result_float").(*runtime.PyFloat).Value)
}

// TestRegisterFuncs tests bulk function registration
func TestRegisterFuncs(t *testing.T) {
	vm := runtime.NewVM()

	// Register multiple functions at once
	vm.RegisterFuncs(map[string]runtime.GoFunction{
		"square": func(vm *runtime.VM) int {
			n := vm.CheckInt(1)
			vm.Push(runtime.NewInt(n * n))
			return 1
		},
		"cube": func(vm *runtime.VM) int {
			n := vm.CheckInt(1)
			vm.Push(runtime.NewInt(n * n * n))
			return 1
		},
	})

	source := `
sq = square(5)
cb = cube(3)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, int64(25), vm.GetGlobal("sq").(*runtime.PyInt).Value)
	assert.Equal(t, int64(27), vm.GetGlobal("cb").(*runtime.PyInt).Value)
}

// TestUserData tests wrapping Go structs as userdata
func TestUserData(t *testing.T) {
	vm := runtime.NewVM()

	type Person struct {
		Name string
		Age  int
	}

	// Register type with methods
	vm.RegisterType("Person",
		// Constructor
		func(vm *runtime.VM) int {
			name := vm.CheckString(1)
			age := int(vm.CheckInt(2))
			person := &Person{Name: name, Age: age}
			ud := vm.NewUserDataWithMeta(person, "Person")
			vm.Push(ud)
			return 1
		},
		// Methods
		map[string]runtime.GoFunction{
			"get_name": func(vm *runtime.VM) int {
				ud := vm.CheckUserData(1, "Person")
				if ud == nil {
					vm.Push(runtime.None)
					return 1
				}
				person := ud.Value.(*Person)
				vm.Push(runtime.NewString(person.Name))
				return 1
			},
			"get_age": func(vm *runtime.VM) int {
				ud := vm.CheckUserData(1, "Person")
				if ud == nil {
					vm.Push(runtime.None)
					return 1
				}
				person := ud.Value.(*Person)
				vm.Push(runtime.NewInt(int64(person.Age)))
				return 1
			},
			"set_age": func(vm *runtime.VM) int {
				ud := vm.CheckUserData(1, "Person")
				if ud == nil {
					return 0
				}
				age := int(vm.CheckInt(2))
				person := ud.Value.(*Person)
				person.Age = age
				return 0
			},
		},
	)

	source := `
p = Person("Alice", 30)
name = p.get_name()
age = p.get_age()
p.set_age(31)
new_age = p.get_age()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, "Alice", vm.GetGlobal("name").(*runtime.PyString).Value)
	assert.Equal(t, int64(30), vm.GetGlobal("age").(*runtime.PyInt).Value)
	assert.Equal(t, int64(31), vm.GetGlobal("new_age").(*runtime.PyInt).Value)
}

// TestFromGoValue tests converting Go values to Python values
func TestFromGoValue(t *testing.T) {
	// Test basic types
	assert.Equal(t, int64(42), runtime.FromGoValue(42).(*runtime.PyInt).Value)
	assert.Equal(t, int64(42), runtime.FromGoValue(int64(42)).(*runtime.PyInt).Value)
	assert.Equal(t, 3.14, runtime.FromGoValue(3.14).(*runtime.PyFloat).Value)
	assert.Equal(t, "hello", runtime.FromGoValue("hello").(*runtime.PyString).Value)
	assert.Equal(t, true, runtime.FromGoValue(true).(*runtime.PyBool).Value)

	// Test slice -> list
	slice := []int{1, 2, 3}
	list := runtime.FromGoValue(slice).(*runtime.PyList)
	assert.Equal(t, 3, len(list.Items))
	assert.Equal(t, int64(1), list.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), list.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), list.Items[2].(*runtime.PyInt).Value)

	// Test nil -> None
	assert.True(t, runtime.IsNone(runtime.FromGoValue(nil)))
}

// TestToGoValue tests converting Python values to Go values
func TestToGoValue(t *testing.T) {
	assert.Equal(t, int64(42), runtime.ToGoValue(runtime.NewInt(42)))
	assert.Equal(t, 3.14, runtime.ToGoValue(runtime.NewFloat(3.14)))
	assert.Equal(t, "hello", runtime.ToGoValue(runtime.NewString("hello")))
	assert.Equal(t, true, runtime.ToGoValue(runtime.NewBool(true)))
	assert.Nil(t, runtime.ToGoValue(runtime.None))
}

// TestTypeCheckers tests the Is* type checking functions
func TestTypeCheckers(t *testing.T) {
	assert.True(t, runtime.IsNone(runtime.None))
	assert.True(t, runtime.IsInt(runtime.NewInt(42)))
	assert.True(t, runtime.IsFloat(runtime.NewFloat(3.14)))
	assert.True(t, runtime.IsString(runtime.NewString("hello")))
	assert.True(t, runtime.IsBool(runtime.NewBool(true)))
	assert.True(t, runtime.IsList(runtime.NewList(nil)))
	assert.True(t, runtime.IsDict(runtime.NewDict()))
	assert.True(t, runtime.IsTuple(runtime.NewTuple(nil)))
	assert.True(t, runtime.IsUserData(runtime.NewUserData(nil)))
}

// TestGoFunctionNoReturn tests Go functions that return nothing
func TestGoFunctionNoReturn(t *testing.T) {
	vm := runtime.NewVM()

	var sideEffect int
	vm.Register("set_value", func(vm *runtime.VM) int {
		sideEffect = int(vm.CheckInt(1))
		return 0
	})

	source := `set_value(42)`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	assert.Equal(t, 42, sideEffect)
}

// TestGoFunctionWithList tests Go functions that work with lists
func TestGoFunctionWithList(t *testing.T) {
	vm := runtime.NewVM()

	vm.Register("sum_list", func(vm *runtime.VM) int {
		list := vm.CheckList(1)
		if list == nil {
			vm.Push(runtime.NewInt(0))
			return 1
		}
		var sum int64
		for _, item := range list.Items {
			if i, ok := item.(*runtime.PyInt); ok {
				sum += i.Value
			}
		}
		vm.Push(runtime.NewInt(sum))
		return 1
	})

	source := `result = sum_list([1, 2, 3, 4, 5])`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, int64(15), result.(*runtime.PyInt).Value)
}

// TestGoFunctionReturningList tests Go functions that return lists
func TestGoFunctionReturningList(t *testing.T) {
	vm := runtime.NewVM()

	vm.Register("make_range", func(vm *runtime.VM) int {
		start := vm.CheckInt(1)
		end := vm.CheckInt(2)
		items := make([]runtime.Value, 0, end-start)
		for i := start; i < end; i++ {
			items = append(items, runtime.NewInt(i))
		}
		vm.Push(runtime.NewList(items))
		return 1
	})

	source := `
nums = make_range(0, 5)
total = sum(nums)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	total := vm.GetGlobal("total")
	assert.Equal(t, int64(10), total.(*runtime.PyInt).Value)
}

// TestCallableCheck tests IsCallable
func TestCallableCheck(t *testing.T) {
	vm := runtime.NewVM()

	vm.Register("test_func", func(vm *runtime.VM) int {
		return 0
	})

	assert.True(t, runtime.IsCallable(runtime.NewGoFunction("test", func(vm *runtime.VM) int { return 0 })))
	assert.False(t, runtime.IsCallable(runtime.NewInt(42)))
	assert.False(t, runtime.IsCallable(runtime.NewString("hello")))
}
