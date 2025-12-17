package test

import (
	"testing"

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/runtime"
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
