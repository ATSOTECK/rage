package test

import (
	"fmt"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVMArithmetic(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"addition", "result = 1 + 2", 3},
		{"subtraction", "result = 5 - 3", 2},
		{"multiplication", "result = 4 * 5", 20},
		{"floor division", "result = 10 // 3", 3},
		{"modulo", "result = 10 % 3", 1},
		{"power", "result = 2 ** 3", 8},
		{"complex expression", "result = 2 + 3 * 4", 14},
		{"parentheses", "result = (2 + 3) * 4", 20},
		{"negative", "result = -5 + 10", 5},
		{"bitwise and", "result = 7 & 3", 3},
		{"bitwise or", "result = 4 | 2", 6},
		{"bitwise xor", "result = 5 ^ 3", 6},
		{"left shift", "result = 1 << 3", 8},
		{"right shift", "result = 8 >> 2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyInt)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMFloatArithmetic(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   float64
	}{
		{"float addition", "result = 1.5 + 2.5", 4.0},
		{"float subtraction", "result = 5.5 - 3.0", 2.5},
		{"float multiplication", "result = 2.0 * 3.5", 7.0},
		{"true division", "result = 7 / 2", 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyFloat)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMComparisons(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"equal true", "result = 1 == 1", true},
		{"equal false", "result = 1 == 2", false},
		{"not equal true", "result = 1 != 2", true},
		{"not equal false", "result = 1 != 1", false},
		{"less than true", "result = 1 < 2", true},
		{"less than false", "result = 2 < 1", false},
		{"less equal true", "result = 2 <= 2", true},
		{"greater than true", "result = 3 > 2", true},
		{"greater equal true", "result = 3 >= 3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyBool)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMBooleanOps(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"and true", "result = True and True", true},
		{"and false", "result = True and False", false},
		{"or true", "result = False or True", true},
		{"or false", "result = False or False", false},
		{"not true", "result = not False", true},
		{"not false", "result = not True", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyBool)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMStrings(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"string literal", `result = "hello"`, "hello"},
		{"string concat", `result = "hello" + " world"`, "hello world"},
		{"string repeat", `result = "ab" * 3`, "ababab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyString)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMLists(t *testing.T) {
	t.Run("list creation", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = [1, 2, 3]", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyList)
		require.True(t, ok)
		require.Len(t, result.Items, 3)
		assert.Equal(t, int64(1), result.Items[0].(*runtime.PyInt).Value)
		assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
		assert.Equal(t, int64(3), result.Items[2].(*runtime.PyInt).Value)
	})

	t.Run("list indexing", func(t *testing.T) {
		code, errs := compiler.CompileSource("x = [10, 20, 30]\nresult = x[1]", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(20), result.Value)
	})

	t.Run("list negative indexing", func(t *testing.T) {
		code, errs := compiler.CompileSource("x = [10, 20, 30]\nresult = x[-1]", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(30), result.Value)
	})
}

func TestVMTuples(t *testing.T) {
	t.Run("tuple creation", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = (1, 2, 3)", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyTuple)
		require.True(t, ok)
		require.Len(t, result.Items, 3)
	})
}

func TestVMDicts(t *testing.T) {
	t.Run("dict creation", func(t *testing.T) {
		code, errs := compiler.CompileSource(`result = {"a": 1, "b": 2}`, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyDict)
		require.True(t, ok)
		require.Len(t, result.Items, 2)
	})
}

func TestVMIfStatement(t *testing.T) {
	t.Run("if true", func(t *testing.T) {
		source := `
if True:
    result = 1
else:
    result = 2
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(1), result.Value)
	})

	t.Run("if false", func(t *testing.T) {
		source := `
if False:
    result = 1
else:
    result = 2
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(2), result.Value)
	})

	t.Run("if elif else", func(t *testing.T) {
		source := `
x = 2
if x == 1:
    result = "one"
elif x == 2:
    result = "two"
else:
    result = "other"
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyString)
		require.True(t, ok)
		assert.Equal(t, "two", result.Value)
	})
}

func TestVMWhileLoop(t *testing.T) {
	t.Run("while loop", func(t *testing.T) {
		source := `
result = 0
i = 0
while i < 5:
    result = result + i
    i = i + 1
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(10), result.Value) // 0+1+2+3+4 = 10
	})

	t.Run("while with break", func(t *testing.T) {
		source := `
result = 0
i = 0
while True:
    result = result + i
    i = i + 1
    if i >= 5:
        break
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(10), result.Value)
	})
}

func TestVMForLoop(t *testing.T) {
	t.Run("for loop over list", func(t *testing.T) {
		source := `
result = 0
for x in [1, 2, 3, 4, 5]:
    result = result + x
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(15), result.Value)
	})
}

func TestVMFunctions(t *testing.T) {
	t.Run("simple function", func(t *testing.T) {
		source := `
def add(a, b):
    return a + b

result = add(3, 4)
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(7), result.Value)
	})

	t.Run("recursive function", func(t *testing.T) {
		source := `
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

result = factorial(5)
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(120), result.Value)
	})

	t.Run("function with default args", func(t *testing.T) {
		source := `
def greet(name, greeting="Hello"):
    return greeting + " " + name

result = greet("World")
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyString)
		require.True(t, ok)
		assert.Equal(t, "Hello World", result.Value)
	})
}

func TestVMBuiltins(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = len([1, 2, 3])", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(3), result.Value)
	})

	t.Run("abs positive", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = abs(-5)", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(5), result.Value)
	})

	t.Run("min", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = min([3, 1, 4, 1, 5])", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(1), result.Value)
	})

	t.Run("max", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = max([3, 1, 4, 1, 5])", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(5), result.Value)
	})

	t.Run("sum", func(t *testing.T) {
		code, errs := compiler.CompileSource("result = sum([1, 2, 3, 4, 5])", "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(15), result.Value)
	})
}

func TestVMAugmentedAssignment(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"plus equal", "x = 5\nx += 3\nresult = x", 8},
		{"minus equal", "x = 5\nx -= 3\nresult = x", 2},
		{"times equal", "x = 5\nx *= 3\nresult = x", 15},
		{"floor div equal", "x = 10\nx //= 3\nresult = x", 3},
		{"mod equal", "x = 10\nx %= 3\nresult = x", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyInt)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMTernaryExpr(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int64
	}{
		{"true branch", "result = 1 if True else 2", 1},
		{"false branch", "result = 1 if False else 2", 2},
		{"with condition", "x = 5\nresult = 10 if x > 3 else 20", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyInt)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMUnpacking(t *testing.T) {
	t.Run("tuple unpacking", func(t *testing.T) {
		source := `
a, b, c = (1, 2, 3)
result = a + b + c
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(6), result.Value)
	})
}

func TestVMInOperator(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"in list true", "result = 2 in [1, 2, 3]", true},
		{"in list false", "result = 4 in [1, 2, 3]", false},
		{"not in list true", "result = 4 not in [1, 2, 3]", true},
		{"in string true", `result = "ell" in "hello"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errs := compiler.CompileSource(tt.source, "<test>")
			require.Empty(t, errs)

			v := runtime.NewVM()
			_, err := v.Execute(code)
			require.NoError(t, err)

			result, ok := v.Globals["result"].(*runtime.PyBool)
			require.True(t, ok)
			assert.Equal(t, tt.want, result.Value)
		})
	}
}

func TestVMClasses(t *testing.T) {
	t.Run("simple class", func(t *testing.T) {
		source := `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def sum(self):
        return self.x + self.y

p = Point(3, 4)
result = p.sum()
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(7), result.Value)
	})

	t.Run("class with inheritance", func(t *testing.T) {
		source := `
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        return "sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

dog = Dog("Buddy")
result = dog.speak()
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyString)
		require.True(t, ok)
		assert.Equal(t, "Woof!", result.Value)
	})

	t.Run("class accessing outer variable", func(t *testing.T) {
		source := `
multiplier = 10

class Calculator:
    def __init__(self, value):
        self.value = value

    def scaled(self):
        return self.value * multiplier

calc = Calculator(5)
result = calc.scaled()
`
		code, errs := compiler.CompileSource(source, "<test>")
		require.Empty(t, errs)

		v := runtime.NewVM()
		_, err := v.Execute(code)
		require.NoError(t, err)

		result, ok := v.Globals["result"].(*runtime.PyInt)
		require.True(t, ok)
		assert.Equal(t, int64(50), result.Value)
	})
}

// =============================================================================
// ClassBuilder Integration Tests — Go-defined classes used from Python
// =============================================================================

// runPy is a helper that runs Python code in a state and returns the named global.
func runPy(t *testing.T, state *rage.State, source, globalName string) rage.Value {
	t.Helper()
	_, err := state.Run(source)
	require.NoError(t, err)
	return state.GetGlobal(globalName)
}

func TestClassBuilder_PersonEndToEnd(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	person := rage.NewClass("Person").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("name", args[0])
			self.Set("age", args[1])
		}).
		Method("greet", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String("Hello, I'm " + name)
		}).
		Str(func(s *rage.State, self rage.Object) string {
			name, _ := rage.AsString(self.Get("name"))
			age, _ := rage.AsInt(self.Get("age"))
			return fmt.Sprintf("Person(%s, %d)", name, age)
		}).
		Build(state)

	state.SetGlobal("Person", person)

	t.Run("instantiation and method call", func(t *testing.T) {
		result := runPy(t, state, `_r = Person("Alice", 30).greet()`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Hello, I'm Alice", str)
	})

	t.Run("attribute access from Python", func(t *testing.T) {
		result := runPy(t, state, `
p = Person("Bob", 25)
_r = p.name
`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Bob", str)
	})

	t.Run("str()", func(t *testing.T) {
		result := runPy(t, state, `_r = str(Person("Charlie", 40))`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Person(Charlie, 40)", str)
	})

	t.Run("isinstance", func(t *testing.T) {
		result := runPy(t, state, `_r = isinstance(Person("D", 1), Person)`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.True(t, b)
	})
}

func TestClassBuilder_InheritanceEndToEnd(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	animal := rage.NewClass("Animal").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("name", args[0])
		}).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			return rage.String("...")
		}).
		Build(state)

	dog := rage.NewClass("Dog").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Woof!")
		}).
		Build(state)

	cat := rage.NewClass("Cat").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Meow!")
		}).
		Build(state)

	state.SetGlobal("Animal", animal)
	state.SetGlobal("Dog", dog)
	state.SetGlobal("Cat", cat)

	t.Run("inherited __init__", func(t *testing.T) {
		result := runPy(t, state, `_r = Dog("Rex").name`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Rex", str)
	})

	t.Run("overridden method", func(t *testing.T) {
		result := runPy(t, state, `_r = Dog("Rex").speak()`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Rex says Woof!", str)
	})

	t.Run("isinstance with base class", func(t *testing.T) {
		result := runPy(t, state, `_r = isinstance(Dog("Rex"), Animal)`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.True(t, b)
	})

	t.Run("isinstance negative", func(t *testing.T) {
		result := runPy(t, state, `_r = isinstance(Dog("Rex"), Cat)`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.False(t, b)
	})

	t.Run("polymorphism in loop", func(t *testing.T) {
		result := runPy(t, state, `
animals = [Dog("Rex"), Cat("Whiskers"), Dog("Buddy")]
_r = [a.speak() for a in animals]
`, "_r")
		items, ok := rage.AsList(result)
		require.True(t, ok)
		require.Len(t, items, 3)
		s0, _ := rage.AsString(items[0])
		s1, _ := rage.AsString(items[1])
		s2, _ := rage.AsString(items[2])
		assert.Equal(t, "Rex says Woof!", s0)
		assert.Equal(t, "Whiskers says Meow!", s1)
		assert.Equal(t, "Buddy says Woof!", s2)
	})
}

func TestClassBuilder_DunderProtocols(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	// A container class that implements multiple dunder protocols
	container := rage.NewClass("Container").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("items", args[0])
		}).
		Len(func(s *rage.State, self rage.Object) int64 {
			items, _ := rage.AsList(self.Get("items"))
			return int64(len(items))
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) rage.Value {
			items, _ := rage.AsList(self.Get("items"))
			idx, _ := rage.AsInt(key)
			if int(idx) < len(items) {
				return items[idx]
			}
			return rage.None
		}).
		Contains(func(s *rage.State, self rage.Object, item rage.Value) bool {
			items, _ := rage.AsList(self.Get("items"))
			itemInt, ok := rage.AsInt(item)
			if !ok {
				return false
			}
			for _, v := range items {
				if n, ok := rage.AsInt(v); ok && n == itemInt {
					return true
				}
			}
			return false
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) bool {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false
			}
			selfItems, _ := rage.AsList(self.Get("items"))
			otherItems, _ := rage.AsList(otherObj.Get("items"))
			if len(selfItems) != len(otherItems) {
				return false
			}
			for i := range selfItems {
				a, _ := rage.AsInt(selfItems[i])
				b, _ := rage.AsInt(otherItems[i])
				if a != b {
					return false
				}
			}
			return true
		}).
		Bool(func(s *rage.State, self rage.Object) bool {
			items, _ := rage.AsList(self.Get("items"))
			return len(items) > 0
		}).
		Str(func(s *rage.State, self rage.Object) string {
			items, _ := rage.AsList(self.Get("items"))
			return fmt.Sprintf("Container(%d items)", len(items))
		}).
		Build(state)

	state.SetGlobal("Container", container)

	t.Run("len", func(t *testing.T) {
		result := runPy(t, state, `_r = len(Container([1, 2, 3]))`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(3), n)
	})

	t.Run("getitem", func(t *testing.T) {
		result := runPy(t, state, `_r = Container([10, 20, 30])[1]`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(20), n)
	})

	t.Run("contains true", func(t *testing.T) {
		result := runPy(t, state, `_r = 2 in Container([1, 2, 3])`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.True(t, b)
	})

	t.Run("contains false", func(t *testing.T) {
		result := runPy(t, state, `_r = 99 in Container([1, 2, 3])`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.False(t, b)
	})

	t.Run("equality", func(t *testing.T) {
		result := runPy(t, state, `_r = Container([1, 2]) == Container([1, 2])`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.True(t, b)
	})

	t.Run("inequality", func(t *testing.T) {
		result := runPy(t, state, `_r = Container([1, 2]) == Container([3, 4])`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.False(t, b)
	})

	t.Run("bool truthy", func(t *testing.T) {
		result := runPy(t, state, `_r = bool(Container([1]))`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.True(t, b)
	})

	t.Run("bool falsy", func(t *testing.T) {
		result := runPy(t, state, `_r = bool(Container([]))`, "_r")
		b, ok := rage.AsBool(result)
		require.True(t, ok)
		assert.False(t, b)
	})

	t.Run("str", func(t *testing.T) {
		result := runPy(t, state, `_r = str(Container([1, 2, 3]))`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Container(3 items)", str)
	})
}

func TestClassBuilder_Callable(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	multiplier := rage.NewClass("Multiplier").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("factor", args[0])
		}).
		Call(func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			factor, _ := rage.AsInt(self.Get("factor"))
			n, _ := rage.AsInt(args[0])
			return rage.Int(factor * n)
		}).
		Build(state)

	state.SetGlobal("Multiplier", multiplier)

	t.Run("instance as callable", func(t *testing.T) {
		result := runPy(t, state, `_r = Multiplier(3)(7)`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(21), n)
	})

	t.Run("callable in map", func(t *testing.T) {
		result := runPy(t, state, `
double = Multiplier(2)
_r = [double(x) for x in [1, 2, 3, 4, 5]]
`, "_r")
		items, ok := rage.AsList(result)
		require.True(t, ok)
		require.Len(t, items, 5)
		for i, expected := range []int64{2, 4, 6, 8, 10} {
			n, ok := rage.AsInt(items[i])
			require.True(t, ok)
			assert.Equal(t, expected, n)
		}
	})
}

func TestClassBuilder_PropertyAccess(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	rect := rage.NewClass("Rect").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("_w", args[0])
			self.Set("_h", args[1])
		}).
		Property("area", func(s *rage.State, self rage.Object) rage.Value {
			w, _ := rage.AsInt(self.Get("_w"))
			h, _ := rage.AsInt(self.Get("_h"))
			return rage.Int(w * h)
		}).
		PropertyWithSetter("width",
			func(s *rage.State, self rage.Object) rage.Value {
				return self.Get("_w")
			},
			func(s *rage.State, self rage.Object, val rage.Value) {
				self.Set("_w", val)
			},
		).
		Build(state)

	state.SetGlobal("Rect", rect)

	t.Run("read-only property", func(t *testing.T) {
		result := runPy(t, state, `_r = Rect(3, 4).area`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(12), n)
	})

	t.Run("read-write property", func(t *testing.T) {
		result := runPy(t, state, `
r = Rect(5, 6)
r.width = 10
_r = r.width
`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(10), n)
	})

	t.Run("property updates derived values", func(t *testing.T) {
		result := runPy(t, state, `
r = Rect(5, 6)
r.width = 10
_r = r.area
`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(60), n)
	})
}

func TestClassBuilder_StaticAndClassMethods(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	counter := rage.NewClass("Counter").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			if len(args) > 0 {
				self.Set("count", args[0])
			} else {
				self.Set("count", rage.Int(0))
			}
		}).
		Method("increment", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			n, _ := rage.AsInt(self.Get("count"))
			self.Set("count", rage.Int(n+1))
			return rage.None
		}).
		StaticMethod("from_string", func(s *rage.State, args ...rage.Value) rage.Value {
			str, _ := rage.AsString(args[0])
			n := int64(0)
			for _, c := range str {
				_ = c
				n++
			}
			return rage.Int(n)
		}).
		ClassMethod("class_info", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) rage.Value {
			return rage.String("class=" + cls.Name())
		}).
		Build(state)

	state.SetGlobal("Counter", counter)

	t.Run("static method on class", func(t *testing.T) {
		result := runPy(t, state, `_r = Counter.from_string("hello")`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(5), n)
	})

	t.Run("static method on instance", func(t *testing.T) {
		result := runPy(t, state, `_r = Counter(0).from_string("ab")`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(2), n)
	})

	t.Run("class method", func(t *testing.T) {
		result := runPy(t, state, `_r = Counter.class_info()`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "class=Counter", str)
	})
}

func TestClassBuilder_DunderAdd(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	vec := rage.NewClass("Vec2").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("x", args[0])
			self.Set("y", args[1])
		}).
		Dunder("__add__", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			other, ok := args[0].(rage.Object)
			if !ok {
				return rage.None
			}
			x1, _ := rage.AsInt(self.Get("x"))
			y1, _ := rage.AsInt(self.Get("y"))
			x2, _ := rage.AsInt(other.Get("x"))
			y2, _ := rage.AsInt(other.Get("y"))
			// Return a list [x1+x2, y1+y2] since we can't easily create a new Vec2 here
			return rage.List(rage.Int(x1+x2), rage.Int(y1+y2))
		}).
		Str(func(s *rage.State, self rage.Object) string {
			x, _ := rage.AsInt(self.Get("x"))
			y, _ := rage.AsInt(self.Get("y"))
			return fmt.Sprintf("Vec2(%d, %d)", x, y)
		}).
		Build(state)

	state.SetGlobal("Vec2", vec)

	t.Run("__add__", func(t *testing.T) {
		result := runPy(t, state, `_r = Vec2(1, 2) + Vec2(3, 4)`, "_r")
		items, ok := rage.AsList(result)
		require.True(t, ok)
		require.Len(t, items, 2)
		x, _ := rage.AsInt(items[0])
		y, _ := rage.AsInt(items[1])
		assert.Equal(t, int64(4), x)
		assert.Equal(t, int64(6), y)
	})

	t.Run("str", func(t *testing.T) {
		result := runPy(t, state, `_r = str(Vec2(5, 10))`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Vec2(5, 10)", str)
	})
}

func TestClassBuilder_GoDefinedUsedInPythonClass(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	// Define a Go-backed base class
	base := rage.NewClass("GoBase").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("value", args[0])
		}).
		Method("get_value", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			return self.Get("value")
		}).
		Build(state)

	state.SetGlobal("GoBase", base)

	t.Run("Python class inheriting Go class", func(t *testing.T) {
		result := runPy(t, state, `
class PyChild(GoBase):
    def doubled(self):
        return self.get_value() * 2

_r = PyChild(21).doubled()
`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(42), n)
	})
}

func TestClassBuilder_StateCall(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	cls := rage.NewClass("Greeter").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("prefix", args[0])
		}).
		Method("greet", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			prefix, _ := rage.AsString(self.Get("prefix"))
			name, _ := rage.AsString(args[0])
			return rage.String(prefix + " " + name)
		}).
		Build(state)

	t.Run("Call class from Go", func(t *testing.T) {
		result, err := state.Call(cls, rage.String("Hi"))
		require.NoError(t, err)

		obj, ok := result.(rage.Object)
		require.True(t, ok)
		assert.Equal(t, "Greeter", obj.ClassName())
	})

	t.Run("Call method from Go via State.Call", func(t *testing.T) {
		state.SetGlobal("Greeter", cls)
		_, err := state.Run(`g = Greeter("Hello")`)
		require.NoError(t, err)

		// Get the Python function and call it via State.Call
		_, err = state.Run(`
def greet_from_go(g, name):
    return g.greet(name)
`)
		require.NoError(t, err)

		fn := state.GetGlobal("greet_from_go")
		g := state.GetGlobal("g")
		result, err := state.Call(fn, g, rage.String("World"))
		require.NoError(t, err)

		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "Hello World", str)
	})
}

func TestClassBuilder_SetItemProtocol(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	store := rage.NewClass("Store").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			// nothing
		}).
		SetItem(func(s *rage.State, self rage.Object, key, val rage.Value) {
			k, _ := rage.AsString(key)
			self.Set("_item_"+k, val)
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) rage.Value {
			k, _ := rage.AsString(key)
			return self.Get("_item_" + k)
		}).
		Len(func(s *rage.State, self rage.Object) int64 {
			count := int64(0)
			// Count attributes starting with _item_
			// This is a simple approximation
			return count
		}).
		Build(state)

	state.SetGlobal("Store", store)

	t.Run("setitem and getitem", func(t *testing.T) {
		result := runPy(t, state, `
s = Store()
s["x"] = 42
s["y"] = 100
_r = s["x"] + s["y"]
`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(142), n)
	})
}

func TestClassBuilder_ReprDunder(t *testing.T) {
	state := rage.NewState()
	defer state.Close()
	state.EnableBuiltin(rage.BuiltinRepr)

	cls := rage.NewClass("Token").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("kind", args[0])
			self.Set("val", args[1])
		}).
		Repr(func(s *rage.State, self rage.Object) string {
			kind, _ := rage.AsString(self.Get("kind"))
			val, _ := rage.AsString(self.Get("val"))
			return fmt.Sprintf("Token(%s, %q)", kind, val)
		}).
		Build(state)

	state.SetGlobal("Token", cls)

	result := runPy(t, state, `_r = repr(Token("INT", "42"))`, "_r")
	str, ok := rage.AsString(result)
	require.True(t, ok)
	assert.Equal(t, `Token(INT, "42")`, str)
}

func TestClassBuilder_NewInstanceFromGo(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	cls := rage.NewClass("Config").
		Method("get", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			key, _ := rage.AsString(args[0])
			return self.Get(key)
		}).
		Build(state)

	// Create instance directly from Go without __init__
	obj := cls.NewInstance()
	obj.Set("host", rage.String("localhost"))
	obj.Set("port", rage.Int(8080))

	state.SetGlobal("config", obj)

	t.Run("use Go-created instance from Python", func(t *testing.T) {
		result := runPy(t, state, `_r = config.get("host")`, "_r")
		str, ok := rage.AsString(result)
		require.True(t, ok)
		assert.Equal(t, "localhost", str)
	})

	t.Run("Python accesses Go-set attribute", func(t *testing.T) {
		result := runPy(t, state, `_r = config.port`, "_r")
		n, ok := rage.AsInt(result)
		require.True(t, ok)
		assert.Equal(t, int64(8080), n)
	})
}
