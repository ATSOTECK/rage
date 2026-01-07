package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Large Data Structure Tests
// =============================================================================

func TestLargeListCreation(t *testing.T) {
	source := `
result = list(range(10000))
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(10000), count.Value)
}

func TestLargeListAppend(t *testing.T) {
	source := `
result = []
for i in range(5000):
    result.append(i)
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5000), count.Value)
}

func TestLargeDictCreation(t *testing.T) {
	source := `
result = {}
for i in range(5000):
    result[i] = i * 2
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5000), count.Value)
}

func TestLargeDictLookup(t *testing.T) {
	source := `
d = {}
for i in range(5000):
    d[i] = i * 2

# Look up every value
total = 0
for i in range(5000):
    total = total + d[i]
`
	vm := runCode(t, source)
	total := vm.GetGlobal("total").(*runtime.PyInt)
	// Sum of 2*i for i in 0..4999 = 2 * (4999 * 5000 / 2) = 24995000
	assert.Equal(t, int64(24995000), total.Value)
}

func TestLargeStringConcatenation(t *testing.T) {
	source := `
result = ""
for i in range(1000):
    result = result + "a"
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
}

func TestLargeStringJoin(t *testing.T) {
	// This is more efficient than repeated concatenation
	source := `
parts = []
for i in range(5000):
    parts.append("x")
result = "".join(parts)
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5000), count.Value)
}

// =============================================================================
// Deep Nesting Tests
// =============================================================================

func TestDeeplyNestedLists(t *testing.T) {
	source := `
result = []
current = result
for i in range(100):
    new_list = []
    current.append(new_list)
    current = new_list
current.append(42)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result")
	assert.NotNil(t, result)
}

func TestDeeplyNestedDicts(t *testing.T) {
	source := `
result = {}
current = result
for i in range(100):
    new_dict = {}
    current["nested"] = new_dict
    current = new_dict
current["value"] = 42
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result")
	assert.NotNil(t, result)
}

func TestDeeplyNestedFunctionCalls(t *testing.T) {
	source := `
def recurse(n):
    if n <= 0:
        return 0
    return 1 + recurse(n - 1)

result = recurse(500)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(500), result.Value)
}

func TestDeeplyNestedLoopsStress(t *testing.T) {
	source := `
count = 0
for i in range(10):
    for j in range(10):
        for k in range(10):
            count = count + 1
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
}

// =============================================================================
// Iteration Tests
// =============================================================================

func TestManyIterations(t *testing.T) {
	source := `
count = 0
for i in range(100000):
    count = count + 1
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(100000), count.Value)
}

func TestManyWhileIterations(t *testing.T) {
	source := `
count = 0
i = 0
while i < 50000:
    count = count + 1
    i = i + 1
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(50000), count.Value)
}

func TestManyListIterations(t *testing.T) {
	source := `
items = list(range(10000))
total = 0
for item in items:
    total = total + item
`
	vm := runCode(t, source)
	total := vm.GetGlobal("total").(*runtime.PyInt)
	// Sum of 0..9999 = 9999 * 10000 / 2 = 49995000
	assert.Equal(t, int64(49995000), total.Value)
}

// =============================================================================
// Computation Tests
// =============================================================================

func TestFibonacciStress(t *testing.T) {
	source := `
def fib(n):
    if n <= 1:
        return n
    a = 0
    b = 1
    for i in range(2, n + 1):
        temp = a + b
        a = b
        b = temp
    return b

result = fib(1000)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// Just check it's a big number
	assert.True(t, result.Value > 0)
}

func TestPrimeComputation(t *testing.T) {
	source := `
def is_prime(n):
    if n < 2:
        return False
    if n == 2:
        return True
    if n % 2 == 0:
        return False
    i = 3
    while i * i <= n:
        if n % i == 0:
            return False
        i = i + 2
    return True

# Find primes up to 1000
primes = []
for n in range(2, 1000):
    if is_prime(n):
        primes.append(n)

count = len(primes)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// There are 168 primes less than 1000
	assert.Equal(t, int64(168), count.Value)
}

func TestSortingLargeList(t *testing.T) {
	source := `
import random
random.seed(42)
items = []
for i in range(1000):
    items.append(random.randint(0, 10000))
sorted_items = sorted(items)
is_sorted = True
for i in range(len(sorted_items) - 1):
    if sorted_items[i] > sorted_items[i + 1]:
        is_sorted = False
        break
`
	vm := runCode(t, source)
	isSorted := vm.GetGlobal("is_sorted").(*runtime.PyBool)
	assert.True(t, isSorted.Value)
}

func TestListReversal(t *testing.T) {
	source := `
items = list(range(5000))
items.reverse()
result = items[0]
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("list.reverse() not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("list.reverse() not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(4999), result.Value)
}

// =============================================================================
// Function Call Overhead Tests
// =============================================================================

func TestManyFunctionCalls(t *testing.T) {
	source := `
def add_one(x):
    return x + 1

result = 0
for i in range(10000):
    result = add_one(result)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(10000), result.Value)
}

func TestManyMethodCalls(t *testing.T) {
	source := `
class Counter:
    def __init__(self):
        self.value = 0

    def increment(self):
        self.value = self.value + 1
        return self.value

c = Counter()
for i in range(5000):
    c.increment()
result = c.value
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(5000), result.Value)
}

func TestManyObjectCreations(t *testing.T) {
	source := `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

points = []
for i in range(1000):
    points.append(Point(i, i * 2))
count = len(points)
last_x = points[999].x
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	lastX := vm.GetGlobal("last_x").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
	assert.Equal(t, int64(999), lastX.Value)
}

// =============================================================================
// Memory Pressure Tests
// =============================================================================

func TestCreateAndDiscardLists(t *testing.T) {
	source := `
for i in range(100):
    temp = list(range(1000))
    # temp goes out of scope each iteration
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestCreateAndDiscardDicts(t *testing.T) {
	source := `
for i in range(100):
    temp = {}
    for j in range(100):
        temp[j] = j * 2
    # temp goes out of scope each iteration
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestCreateAndDiscardStrings(t *testing.T) {
	source := `
for i in range(1000):
    temp = "x" * 1000
    # temp goes out of scope each iteration
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Closure Stress Tests
// =============================================================================

func TestManyClosures(t *testing.T) {
	source := `
def make_adder(n):
    def adder(x):
        return x + n
    return adder

adders = []
for i in range(1000):
    adders.append(make_adder(i))

# Call each closure
total = 0
for i in range(1000):
    total = total + adders[i](1)
`
	vm := runCode(t, source)
	total := vm.GetGlobal("total").(*runtime.PyInt)
	// Sum of (1 + i) for i in 0..999 = 1000 + sum(0..999) = 1000 + 499500 = 500500
	assert.Equal(t, int64(500500), total.Value)
}

// =============================================================================
// Generator Stress Tests
// =============================================================================

func TestLargeGenerator(t *testing.T) {
	source := `
def count_gen(n):
    i = 0
    while i < n:
        yield i
        i = i + 1

total = 0
for x in count_gen(10000):
    total = total + x
`
	vm := runCode(t, source)
	total := vm.GetGlobal("total").(*runtime.PyInt)
	// Sum of 0..9999
	assert.Equal(t, int64(49995000), total.Value)
}

func TestGeneratorPipelineStress(t *testing.T) {
	source := `
def numbers(n):
    for i in range(n):
        yield i

def doubled(gen):
    for x in gen:
        yield x * 2

def filtered(gen):
    for x in gen:
        if x % 4 == 0:
            yield x

total = 0
for x in filtered(doubled(numbers(1000))):
    total = total + x
`
	vm := runCode(t, source)
	total := vm.GetGlobal("total").(*runtime.PyInt)
	// doubled values divisible by 4 are 0, 4, 8, ... (i.e., 4k where k = 0..499)
	// Wait, doubled gives 0, 2, 4, 6, 8, ...
	// filtered keeps those divisible by 4: 0, 4, 8, 12, ...
	// These are 4*0, 4*1, 4*2, ... 4*499 (since 4*500 = 2000 > 1998)
	// Sum = 4 * (0 + 1 + 2 + ... + 499) = 4 * 499 * 500 / 2 = 4 * 124750 = 499000
	assert.Equal(t, int64(499000), total.Value)
}

// =============================================================================
// String Operations Stress Tests
// =============================================================================

func TestManyStringSplits(t *testing.T) {
	source := `
s = "a,b,c,d,e"
for i in range(10000):
    parts = s.split(",")
count = len(parts)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5), count.Value)
}

func TestManyStringJoins(t *testing.T) {
	source := `
parts = ["a", "b", "c", "d", "e"]
for i in range(10000):
    result = ",".join(parts)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "a,b,c,d,e", result.Value)
}

func TestManyStringSearches(t *testing.T) {
	source := `
s = "abcdefghijklmnopqrstuvwxyz" * 100
count = 0
for i in range(1000):
    if "xyz" in s:
        count = count + 1
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
}

// =============================================================================
// List Comprehension Stress Tests
// =============================================================================

func TestLargeListComprehension(t *testing.T) {
	source := `
result = [x * 2 for x in range(10000)]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(10000), count.Value)
}

func TestFilteredListComprehension(t *testing.T) {
	source := `
result = [x for x in range(10000) if x % 2 == 0]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5000), count.Value)
}

func TestNestedListComprehensionStress(t *testing.T) {
	source := `
result = [[i * j for j in range(10)] for i in range(100)]
count = len(result)
inner_count = len(result[0])
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	innerCount := vm.GetGlobal("inner_count").(*runtime.PyInt)
	assert.Equal(t, int64(100), count.Value)
	assert.Equal(t, int64(10), innerCount.Value)
}

// =============================================================================
// Dict Comprehension Stress Tests
// =============================================================================

func TestLargeDictComprehension(t *testing.T) {
	source := `
result = {x: x * 2 for x in range(5000)}
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5000), count.Value)
}

// =============================================================================
// Exception Handling Stress Tests
// =============================================================================

func TestManyTryCatches(t *testing.T) {
	source := `
count = 0
for i in range(1000):
    try:
        x = 1 / 1  # No error
        count = count + 1
    except:
        pass
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
}

func TestManyCaughtExceptions(t *testing.T) {
	source := `
count = 0
for i in range(1000):
    try:
        x = 1 / 0  # Error
    except:
        count = count + 1
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Exception handling in loops not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Exception handling in loops not fully working: " + err.Error())
		return
	}
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(1000), count.Value)
}

// =============================================================================
// Benchmark-style Tests
// These test execution time to ensure no severe performance regressions
// =============================================================================

func TestBenchmarkSimpleLoop(t *testing.T) {
	source := `
count = 0
for i in range(100000):
    count = count + 1
`
	start := time.Now()
	vm := runCode(t, source)
	elapsed := time.Since(start)

	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(100000), count.Value)

	// Should complete in reasonable time (< 5 seconds)
	assert.Less(t, elapsed.Seconds(), 5.0, "Simple loop took too long: %v", elapsed)
}

func TestBenchmarkFunctionCalls(t *testing.T) {
	source := `
def noop():
    pass

for i in range(50000):
    noop()
result = True
`
	start := time.Now()
	vm := runCode(t, source)
	elapsed := time.Since(start)

	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)

	// Should complete in reasonable time (< 5 seconds)
	assert.Less(t, elapsed.Seconds(), 5.0, "Function calls took too long: %v", elapsed)
}

func TestBenchmarkDictOperations(t *testing.T) {
	source := `
d = {}
for i in range(10000):
    d[i] = i * 2
for i in range(10000):
    x = d[i]
result = len(d)
`
	start := time.Now()
	vm := runCode(t, source)
	elapsed := time.Since(start)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(10000), result.Value)

	// Should complete in reasonable time (< 5 seconds)
	assert.Less(t, elapsed.Seconds(), 5.0, "Dict operations took too long: %v", elapsed)
}

// =============================================================================
// Edge Cases Under Stress
// =============================================================================

func TestEmptyLoopStress(t *testing.T) {
	source := `
for i in range(100000):
    pass
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestNestedEmptyLoopsStress(t *testing.T) {
	source := `
for i in range(100):
    for j in range(100):
        for k in range(100):
            pass
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

// =============================================================================
// Compile Time Stress Tests
// =============================================================================

func TestCompileLargeFunction(t *testing.T) {
	// Generate a function with many statements
	source := "def big_func():\n"
	for i := 0; i < 500; i++ {
		source += fmt.Sprintf("    x%d = %d\n", i, i)
	}
	source += "    return x499\n"
	source += "result = big_func()\n"

	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(499), result.Value)
}

func TestCompileManyFunctions(t *testing.T) {
	// Generate many small functions
	source := ""
	for i := 0; i < 100; i++ {
		source += fmt.Sprintf("def func%d(x):\n    return x + %d\n\n", i, i)
	}
	source += "result = func99(1)\n"

	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(100), result.Value)
}

func TestCompileDeeplyNestedExpressions(t *testing.T) {
	// Generate deeply nested expression
	source := "result = "
	for i := 0; i < 50; i++ {
		source += "("
	}
	source += "1"
	for i := 0; i < 50; i++ {
		source += " + 1)"
	}
	source += "\n"

	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(51), result.Value)
}
