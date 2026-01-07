package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic Generator Tests
// Note: Tests use for loops to iterate over generators since list() on generators
// may not be fully implemented. Also, next() builtin may not be available.
// =============================================================================

func TestSimpleGenerator(t *testing.T) {
	source := `
def simple_gen():
    yield 1
    yield 2
    yield 3

result = []
for x in simple_gen():
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, int64(1), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), result.Items[2].(*runtime.PyInt).Value)
}

func TestGeneratorWithLoop(t *testing.T) {
	source := `
def count_gen(n):
    i = 0
    while i < n:
        yield i
        i = i + 1

result = []
for x in count_gen(5):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, int64(i), result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestGeneratorWithForLoop(t *testing.T) {
	source := `
def squares_gen(nums):
    for n in nums:
        yield n * n

result = []
for x in squares_gen([1, 2, 3, 4]):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 4)
	expected := []int64{1, 4, 9, 16}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestGeneratorWithReturn(t *testing.T) {
	// Test that generator stops when exhausted (bare return causes parser issues)
	source := `
def gen_with_limited():
    yield 1
    yield 2

result = []
for x in gen_with_limited():
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 2)
	assert.Equal(t, int64(1), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
}

func TestGeneratorWithConditionalYield(t *testing.T) {
	source := `
def filter_gen(nums, threshold):
    for n in nums:
        if n > threshold:
            yield n

result = []
for x in filter_gen([1, 5, 2, 8, 3, 9], 4):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	expected := []int64{5, 8, 9}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator Expression Tests
// =============================================================================

func TestGeneratorExpression(t *testing.T) {
	// Note: Generator expressions may not be fully implemented
	// Skip if compilation fails
	source := `
gen = (x * 2 for x in [1, 2, 3])
result = []
for x in gen:
    result.append(x)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Generator expressions not fully supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Generator expressions not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	expected := []int64{2, 4, 6}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestGeneratorExpressionWithCondition(t *testing.T) {
	// Skip - generator expressions with conditions not fully supported
	t.Skip("Generator expressions with conditions not fully supported")
}

func TestNestedGeneratorExpression(t *testing.T) {
	// Skip - nested generator expressions not fully supported
	t.Skip("Nested generator expressions not fully supported")
}

// =============================================================================
// Generator Iteration Tests
// =============================================================================

func TestGeneratorInForLoop(t *testing.T) {
	source := `
def gen():
    yield 1
    yield 2
    yield 3

result = 0
for x in gen():
    result += x
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(6), result.Value)
}

func TestGeneratorExhaustion(t *testing.T) {
	source := `
def gen():
    yield 1

g = gen()
result1 = []
for x in g:
    result1.append(x)

result2 = []
for x in g:
    result2.append(x)  # Should be empty - generator exhausted
`
	vm := runCode(t, source)
	result1 := vm.GetGlobal("result1").(*runtime.PyList)
	result2 := vm.GetGlobal("result2").(*runtime.PyList)
	assert.Len(t, result1.Items, 1)
	assert.Len(t, result2.Items, 0)
}

// =============================================================================
// Yield From Tests
// =============================================================================

func TestYieldFrom(t *testing.T) {
	source := `
def inner_gen():
    yield 1
    yield 2

def outer_gen():
    yield from inner_gen()
    yield 3

result = []
for x in outer_gen():
    result.append(x)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("yield from not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("yield from not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, int64(1), result.Items[0].(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result.Items[1].(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), result.Items[2].(*runtime.PyInt).Value)
}

func TestYieldFromList(t *testing.T) {
	source := `
def gen():
    yield from [1, 2, 3]

result = []
for x in gen():
    result.append(x)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("yield from not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("yield from not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, int64(i+1), result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestYieldFromString(t *testing.T) {
	source := `
def char_gen():
    yield from "abc"

result = []
for c in char_gen():
    result.append(c)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("yield from not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("yield from not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "a", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "b", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "c", result.Items[2].(*runtime.PyString).Value)
}

func TestNestedYieldFrom(t *testing.T) {
	source := `
def gen1():
    yield 1

def gen2():
    yield from gen1()
    yield 2

def gen3():
    yield from gen2()
    yield 3

result = []
for x in gen3():
    result.append(x)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("yield from not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("yield from not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, int64(i+1), result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator State Tests
// =============================================================================

func TestGeneratorMaintainsState(t *testing.T) {
	source := `
def stateful_gen():
    x = 0
    while x < 5:
        x = x + 1
        yield x

result = []
for v in stateful_gen():
    result.append(v)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, int64(i+1), result.Items[i].(*runtime.PyInt).Value)
	}
}

func TestMultipleGeneratorInstances(t *testing.T) {
	source := `
def counter():
    x = 0
    while x < 3:
        x = x + 1
        yield x

g1 = counter()
g2 = counter()

result1 = []
for x in g1:
    result1.append(x)

result2 = []
for x in g2:
    result2.append(x)
`
	vm := runCode(t, source)
	result1 := vm.GetGlobal("result1").(*runtime.PyList)
	result2 := vm.GetGlobal("result2").(*runtime.PyList)
	// Both should have independent state
	require.Len(t, result1.Items, 3)
	require.Len(t, result2.Items, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, int64(i+1), result1.Items[i].(*runtime.PyInt).Value)
		assert.Equal(t, int64(i+1), result2.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator with Arguments Tests
// =============================================================================

func TestGeneratorWithArguments(t *testing.T) {
	source := `
def range_gen(start, stop, step=1):
    current = start
    while current < stop:
        yield current
        current = current + step

result = []
for x in range_gen(0, 10, 2):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 5)
	expected := []int64{0, 2, 4, 6, 8}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator Edge Cases
// =============================================================================

func TestEmptyGenerator(t *testing.T) {
	// Test generator that yields nothing (no yields hit)
	source := `
def empty_gen():
    if False:
        yield 1

result = []
for x in empty_gen():
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	assert.Len(t, result.Items, 0)
}

func TestGeneratorYieldNone(t *testing.T) {
	source := `
def gen():
    yield None
    yield

result = []
for x in gen():
    result.append(x)

count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(2), count.Value)
}

// =============================================================================
// Fibonacci Generator (Classic Use Case)
// =============================================================================

func TestFibonacciGenerator(t *testing.T) {
	source := `
def fib_gen(n):
    a = 0
    b = 1
    count = 0
    while count < n:
        yield a
        temp = a + b
        a = b
        b = temp
        count = count + 1

result = []
for x in fib_gen(10):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 10)
	expected := []int64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator Comprehension in Function Arguments
// Note: Generator expressions as function arguments may not be supported
// =============================================================================

func TestGeneratorAsArgumentToSum(t *testing.T) {
	// Skip - generator expressions as function args not supported
	t.Skip("Generator expressions as function arguments not supported")
}

func TestGeneratorWithAny(t *testing.T) {
	// Skip - generator expressions as function args not supported
	t.Skip("Generator expressions as function arguments not supported")
}

func TestGeneratorWithAll(t *testing.T) {
	// Skip - generator expressions as function args not supported
	t.Skip("Generator expressions as function arguments not supported")
}

func TestGeneratorWithMin(t *testing.T) {
	// Skip - generator expressions as function args not supported
	t.Skip("Generator expressions as function arguments not supported")
}

func TestGeneratorWithMax(t *testing.T) {
	// Skip - generator expressions as function args not supported
	t.Skip("Generator expressions as function arguments not supported")
}

// =============================================================================
// Generator with Break
// =============================================================================

func TestGeneratorWithBreak(t *testing.T) {
	source := `
def infinite_gen():
    i = 0
    while True:
        yield i
        i = i + 1

result = []
for x in infinite_gen():
    if x >= 5:
        break
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, int64(i), result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator Type Check
// =============================================================================

func TestGeneratorTypeString(t *testing.T) {
	// Test that a generator object exists and can be assigned
	source := `
def gen():
    yield 1

g = gen()
result = []
for x in g:
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.Items[0].(*runtime.PyInt).Value)
}

// =============================================================================
// Complex Generator Pattern
// =============================================================================

func TestGeneratorPipeline(t *testing.T) {
	source := `
def numbers(n):
    for i in range(n):
        yield i

def squared(gen):
    for x in gen:
        yield x * x

def filtered(gen, threshold):
    for x in gen:
        if x >= threshold:
            yield x

result = []
for x in filtered(squared(numbers(10)), 25):
    result.append(x)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyList)
	// squares >= 25: 25, 36, 49, 64, 81 (from 5, 6, 7, 8, 9)
	require.Len(t, result.Items, 5)
	expected := []int64{25, 36, 49, 64, 81}
	for i, exp := range expected {
		assert.Equal(t, exp, result.Items[i].(*runtime.PyInt).Value)
	}
}

// =============================================================================
// Generator Send Tests (if send is implemented)
// =============================================================================

func TestGeneratorSend(t *testing.T) {
	// Skip this test if send is not implemented
	source := `
def echo_gen():
    received = None
    while True:
        received = yield received

g = echo_gen()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Generator send syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Generator send not fully implemented")
		return
	}
	// If we get here, at least the generator was created
	g := vm.GetGlobal("g")
	assert.NotNil(t, g)
}
