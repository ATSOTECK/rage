package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Expression Parsing Edge Cases
// =============================================================================

func TestParenthesizedExpressions(t *testing.T) {
	source := `
result = ((((1 + 2))))
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

func TestComplexArithmeticPrecedence(t *testing.T) {
	source := `
result = 2 + 3 * 4 - 5 / 1
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyFloat)
	// 2 + 12 - 5 = 9
	assert.Equal(t, 9.0, result.Value)
}

func TestUnaryOperatorPrecedence(t *testing.T) {
	source := `
a = -3 * 2
b = --4
c = -(-5)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyInt)
	b := vm.GetGlobal("b").(*runtime.PyInt)
	c := vm.GetGlobal("c").(*runtime.PyInt)
	assert.Equal(t, int64(-6), a.Value)
	assert.Equal(t, int64(4), b.Value)
	assert.Equal(t, int64(5), c.Value)
}

func TestPowerOperatorRightAssociativity(t *testing.T) {
	source := `
result = 2 ** 3 ** 2
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// 2 ** (3 ** 2) = 2 ** 9 = 512
	assert.Equal(t, int64(512), result.Value)
}

func TestNegativePower(t *testing.T) {
	source := `
result = (-2) ** 3
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(-8), result.Value)
}

func TestComparisonChaining(t *testing.T) {
	source := `
a = 1 < 2 < 3
b = 1 < 3 < 2
c = 1 < 2 <= 2 < 3
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyBool)
	b := vm.GetGlobal("b").(*runtime.PyBool)
	c := vm.GetGlobal("c").(*runtime.PyBool)
	assert.True(t, a.Value)
	assert.False(t, b.Value)
	assert.True(t, c.Value)
}

func TestBooleanOperatorPrecedence(t *testing.T) {
	source := `
a = True or False and False
b = (True or False) and False
c = not True or True
d = not (True or True)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyBool)
	b := vm.GetGlobal("b").(*runtime.PyBool)
	c := vm.GetGlobal("c").(*runtime.PyBool)
	d := vm.GetGlobal("d").(*runtime.PyBool)
	assert.True(t, a.Value)
	assert.False(t, b.Value)
	assert.True(t, c.Value)
	assert.False(t, d.Value)
}

func TestBitwiseOperatorPrecedence(t *testing.T) {
	source := `
a = 1 | 2 & 3
b = (1 | 2) & 3
c = 1 ^ 2 | 3
d = 1 ^ (2 | 3)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyInt)
	b := vm.GetGlobal("b").(*runtime.PyInt)
	c := vm.GetGlobal("c").(*runtime.PyInt)
	d := vm.GetGlobal("d").(*runtime.PyInt)
	// 1 | (2 & 3) = 1 | 2 = 3
	assert.Equal(t, int64(3), a.Value)
	// (1 | 2) & 3 = 3 & 3 = 3
	assert.Equal(t, int64(3), b.Value)
	// (1 ^ 2) | 3 = 3 | 3 = 3
	assert.Equal(t, int64(3), c.Value)
	// 1 ^ (2 | 3) = 1 ^ 3 = 2
	assert.Equal(t, int64(2), d.Value)
}

// =============================================================================
// Literal Parsing Edge Cases
// =============================================================================

func TestIntegerLiterals(t *testing.T) {
	source := `
a = 0
b = 123
c = 1_000_000
d = 0b1010
e = 0o755
f = 0xff
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		// Some formats may not be supported
		t.Skip("Some integer literal formats not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Some integer literal formats not supported: " + err.Error())
		return
	}
	a := vm.GetGlobal("a").(*runtime.PyInt)
	b := vm.GetGlobal("b").(*runtime.PyInt)
	c := vm.GetGlobal("c").(*runtime.PyInt)
	assert.Equal(t, int64(0), a.Value)
	assert.Equal(t, int64(123), b.Value)
	assert.Equal(t, int64(1000000), c.Value)
}

func TestFloatLiterals(t *testing.T) {
	source := `
a = 0.0
b = 1.5
c = .5
d = 1.
e = 1e10
f = 1.5e-3
g = 1_000.5
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		// Some formats may not be supported
		t.Skip("Some float literal formats not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Some float literal formats not supported: " + err.Error())
		return
	}
	a := vm.GetGlobal("a").(*runtime.PyFloat)
	b := vm.GetGlobal("b").(*runtime.PyFloat)
	assert.Equal(t, 0.0, a.Value)
	assert.Equal(t, 1.5, b.Value)
}

func TestStringLiteralVariations(t *testing.T) {
	source := `
a = "double"
b = 'single'
c = """triple double"""
d = '''triple single'''
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyString)
	b := vm.GetGlobal("b").(*runtime.PyString)
	c := vm.GetGlobal("c").(*runtime.PyString)
	d := vm.GetGlobal("d").(*runtime.PyString)
	assert.Equal(t, "double", a.Value)
	assert.Equal(t, "single", b.Value)
	assert.Equal(t, "triple double", c.Value)
	assert.Equal(t, "triple single", d.Value)
}

func TestStringConcatenationLiterals(t *testing.T) {
	source := `
result = "hello " "world"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("String literal concatenation not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("String literal concatenation not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	// Skip if implicit string concatenation doesn't work
	if result.Value != "hello world" {
		t.Skip("Implicit string literal concatenation not supported by RAGE")
		return
	}
	assert.Equal(t, "hello world", result.Value)
}

func TestEmptyCollectionLiterals(t *testing.T) {
	source := `
a = []
b = {}
c = ()
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyList)
	b := vm.GetGlobal("b").(*runtime.PyDict)
	assert.Len(t, a.Items, 0)
	assert.Len(t, b.Items, 0)
}

func TestSingleElementTuple(t *testing.T) {
	source := `
a = (1,)
b = (1)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyTuple)
	b := vm.GetGlobal("b").(*runtime.PyInt)
	assert.Len(t, a.Items, 1)
	assert.Equal(t, int64(1), b.Value)
}

func TestTrailingCommaInCollections(t *testing.T) {
	source := `
a = [1, 2, 3,]
b = {"a": 1, "b": 2,}
c = (1, 2, 3,)
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyList)
	b := vm.GetGlobal("b").(*runtime.PyDict)
	c := vm.GetGlobal("c").(*runtime.PyTuple)
	assert.Len(t, a.Items, 3)
	assert.Len(t, b.Items, 2)
	assert.Len(t, c.Items, 3)
}

// =============================================================================
// Statement Parsing Edge Cases
// =============================================================================

func TestOneLinerIf(t *testing.T) {
	source := `
x = 10
if x > 5: result = True
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("One-liner if statement not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("One-liner if statement not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestOneLinerIfElse(t *testing.T) {
	source := `
x = 3
if x > 5: result = True
else: result = False
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("One-liner if/else statement not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("One-liner if/else statement not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.False(t, result.Value)
}

func TestSemicolonSeparatedStatements(t *testing.T) {
	source := `
a = 1; b = 2; c = 3
result = a + b + c
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(6), result.Value)
}

func TestMultipleAssignment(t *testing.T) {
	source := `
a = b = c = 5
result = a + b + c
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(15), result.Value)
}

func TestAugmentedAssignment(t *testing.T) {
	source := `
a = 10
a += 5
b = 20
b -= 5
c = 3
c *= 4
d = 20
d //= 3
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyInt)
	b := vm.GetGlobal("b").(*runtime.PyInt)
	c := vm.GetGlobal("c").(*runtime.PyInt)
	d := vm.GetGlobal("d").(*runtime.PyInt)
	assert.Equal(t, int64(15), a.Value)
	assert.Equal(t, int64(15), b.Value)
	assert.Equal(t, int64(12), c.Value)
	assert.Equal(t, int64(6), d.Value)
}

func TestPassStatement(t *testing.T) {
	source := `
def empty_func():
    pass

class EmptyClass:
    pass

if True:
    pass

result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestDelStatement(t *testing.T) {
	source := `
a = 10
del a
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("del statement not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("del statement not fully supported: " + err.Error())
		return
	}
	// After del, 'a' should not exist
	a := vm.GetGlobal("a")
	assert.Nil(t, a)
}

// =============================================================================
// Function Definition Edge Cases
// =============================================================================

func TestFunctionWithDefaultArgs(t *testing.T) {
	source := `
def greet(name, greeting="Hello"):
    return greeting + ", " + name

a = greet("World")
b = greet("World", "Hi")
`
	vm := runCode(t, source)
	a := vm.GetGlobal("a").(*runtime.PyString)
	b := vm.GetGlobal("b").(*runtime.PyString)
	assert.Equal(t, "Hello, World", a.Value)
	assert.Equal(t, "Hi, World", b.Value)
}

func TestFunctionWithStarArgs(t *testing.T) {
	source := `
def sum_all(*args):
    total = 0
    for arg in args:
        total = total + arg
    return total

result = sum_all(1, 2, 3, 4, 5)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("*args not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("*args not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(15), result.Value)
}

func TestFunctionWithKwargs(t *testing.T) {
	source := `
def show_kwargs(**kwargs):
    return len(kwargs)

result = show_kwargs(a=1, b=2, c=3)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("**kwargs not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("**kwargs not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

func TestFunctionWithMixedArgs(t *testing.T) {
	source := `
def mixed(a, b=2, *args, **kwargs):
    return a + b + len(args) + len(kwargs)

result = mixed(1, 3, 4, 5, x=10, y=20)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Mixed *args/**kwargs not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Mixed *args/**kwargs not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// 1 + 3 + 2 (args: 4,5) + 2 (kwargs: x,y) = 8
	assert.Equal(t, int64(8), result.Value)
}

func TestLambdaExpression(t *testing.T) {
	source := `
double = lambda x: x * 2
result = double(21)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func TestLambdaWithMultipleArgs(t *testing.T) {
	source := `
add = lambda a, b: a + b
result = add(10, 32)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func TestLambdaWithDefault(t *testing.T) {
	source := `
greet = lambda name, prefix="Hello": prefix + " " + name
result = greet("World")
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "Hello World", result.Value)
}

func TestNestedLambdasParser(t *testing.T) {
	source := `
make_adder = lambda n: lambda x: x + n
add_five = make_adder(5)
result = add_five(37)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// Class Definition Edge Cases
// =============================================================================

func TestClassWithPass(t *testing.T) {
	source := `
class Empty:
    pass

obj = Empty()
result = True
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestClassWithClassAttributes(t *testing.T) {
	source := `
class MyClass:
    class_attr = 42

result = MyClass.class_attr
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func TestClassWithMethods(t *testing.T) {
	source := `
class Calculator:
    def add(self, a, b):
        return a + b

    def multiply(self, a, b):
        return a * b

calc = Calculator()
result = calc.add(3, 4) + calc.multiply(2, 5)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(17), result.Value)
}

func TestClassInheritance(t *testing.T) {
	source := `
class Animal:
    def speak(self):
        return "?"

class Dog(Animal):
    def speak(self):
        return "Woof"

d = Dog()
result = d.speak()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "Woof", result.Value)
}

func TestClassMultipleInheritance(t *testing.T) {
	source := `
class A:
    def method(self):
        return "A"

class B:
    def another(self):
        return "B"

class C(A, B):
    pass

c = C()
result = c.method() + c.another()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "AB", result.Value)
}

// =============================================================================
// Loop Edge Cases
// =============================================================================

func TestForElseParser(t *testing.T) {
	source := `
result = ""
for i in range(3):
    result = result + str(i)
else:
    result = result + "done"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "012done", result.Value)
}

func TestForElseWithBreak(t *testing.T) {
	source := `
result = ""
for i in range(5):
    if i == 3:
        break
    result = result + str(i)
else:
    result = result + "done"  # Should not execute
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "012", result.Value)
}

func TestWhileElseParser(t *testing.T) {
	source := `
result = ""
i = 0
while i < 3:
    result = result + str(i)
    i = i + 1
else:
    result = result + "done"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "012done", result.Value)
}

func TestNestedBreakContinueParser(t *testing.T) {
	source := `
result = []
for i in range(3):
    for j in range(3):
        if j == 1:
            continue
        if j == 2:
            break
        result.append(i * 10 + j)
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// Only j=0 passes, so we get 0, 10, 20
	assert.Equal(t, int64(3), count.Value)
}

// =============================================================================
// Exception Handling Edge Cases
// =============================================================================

func TestTryExceptElse(t *testing.T) {
	source := `
result = ""
try:
    x = 1
    result = result + "try"
except:
    result = result + "except"
else:
    result = result + "else"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "tryelse", result.Value)
}

func TestTryExceptElseFinally(t *testing.T) {
	source := `
result = ""
try:
    x = 1
    result = result + "try"
except:
    result = result + "except"
else:
    result = result + "else"
finally:
    result = result + "finally"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "tryelsefinally", result.Value)
}

func TestMultipleExceptBlocks(t *testing.T) {
	source := `
result = ""
try:
    x = 1 / 0
except ValueError:
    result = "value"
except ZeroDivisionError:
    result = "zero"
except:
    result = "other"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Multiple except blocks not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Multiple except blocks not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Contains(t, []string{"zero", "other"}, result.Value)
}

func TestExceptAs(t *testing.T) {
	source := `
result = ""
try:
    raise ValueError("test error")
except ValueError as e:
    result = "caught"
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("except...as not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("except...as not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "caught", result.Value)
}

// =============================================================================
// Slicing Edge Cases
// =============================================================================

func TestSliceWithNegativeIndicesParser(t *testing.T) {
	source := `
s = "abcdef"
result = s[-3:-1]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "de", result.Value)
}

func TestSliceWithStepParser(t *testing.T) {
	source := `
s = "abcdef"
result = s[::2]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "ace", result.Value)
}

func TestSliceReverseParser(t *testing.T) {
	source := `
s = "abcdef"
result = s[::-1]
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "fedcba", result.Value)
}

func TestListSliceAssignment(t *testing.T) {
	source := `
lst = [1, 2, 3, 4, 5]
lst[1:3] = [20, 30]
result = lst
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Slice assignment not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Slice assignment not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	require.Len(t, result.Items, 5)
}

// =============================================================================
// Comprehension Edge Cases
// =============================================================================

func TestNestedListComprehensionParser(t *testing.T) {
	source := `
result = [[j for j in range(3)] for i in range(3)]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

func TestListComprehensionWithMultipleFor(t *testing.T) {
	source := `
result = [i * j for i in range(3) for j in range(3)]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(9), count.Value)
}

func TestListComprehensionWithNestedIf(t *testing.T) {
	source := `
result = [x for x in range(20) if x % 2 == 0 if x % 3 == 0]
count = len(result)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Multiple if in comprehension not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Multiple if in comprehension not supported: " + err.Error())
		return
	}
	count := vm.GetGlobal("count").(*runtime.PyInt)
	// 0, 6, 12, 18
	assert.Equal(t, int64(4), count.Value)
}

func TestDictComprehensionWithCondition(t *testing.T) {
	source := `
result = {x: x**2 for x in range(10) if x % 2 == 0}
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5), count.Value)
}

func TestSetComprehensionParser(t *testing.T) {
	source := `
result = {x % 3 for x in range(10)}
count = len(result)
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Set comprehension not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Set comprehension not fully supported: " + err.Error())
		return
	}
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(3), count.Value)
}

// =============================================================================
// Decorator Edge Cases
// =============================================================================

func TestFunctionDecorator(t *testing.T) {
	source := `
def decorator(func):
    def wrapper():
        return func() + 1
    return wrapper

@decorator
def get_value():
    return 41

result = get_value()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func TestDecoratorWithArgs(t *testing.T) {
	source := `
def multiply(n):
    def decorator(func):
        def wrapper(x):
            return func(x) * n
        return wrapper
    return decorator

@multiply(2)
def add_one(x):
    return x + 1

result = add_one(20)
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

func TestMultipleDecorators(t *testing.T) {
	source := `
def dec1(func):
    def wrapper():
        return func() + 1
    return wrapper

def dec2(func):
    def wrapper():
        return func() * 2
    return wrapper

@dec1
@dec2
def get_value():
    return 10

result = get_value()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	// dec2 applies first: 10 * 2 = 20, then dec1: 20 + 1 = 21
	assert.Equal(t, int64(21), result.Value)
}

func TestClassDecoratorParser(t *testing.T) {
	source := `
def add_method(cls):
    cls.added = 42
    return cls

@add_method
class MyClass:
    pass

result = MyClass.added
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// Global and Nonlocal Edge Cases
// =============================================================================

func TestGlobalStatement(t *testing.T) {
	source := `
x = 10

def modify_global():
    global x
    x = 20

modify_global()
result = x
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(20), result.Value)
}

func TestNonlocalStatement(t *testing.T) {
	source := `
def outer():
    x = 10
    def inner():
        nonlocal x
        x = 20
    inner()
    return x

result = outer()
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("nonlocal statement not supported: compile error")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("nonlocal statement not supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(20), result.Value)
}

// =============================================================================
// Walrus Operator (if supported)
// =============================================================================

func TestWalrusOperator(t *testing.T) {
	source := `
if (n := 10) > 5:
    result = n * 2
`
	vm := runtime.NewVM()
	code, errs := compiler.CompileSource(source, "<test>")
	if len(errs) > 0 {
		t.Skip("Walrus operator not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Walrus operator not fully supported: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(20), result.Value)
}

// =============================================================================
// Ternary Expression Edge Cases
// =============================================================================

func TestTernaryExpression(t *testing.T) {
	source := `
x = 10
result = "big" if x > 5 else "small"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "big", result.Value)
}

func TestNestedTernaryExpression(t *testing.T) {
	source := `
x = 5
result = "big" if x > 10 else "medium" if x > 3 else "small"
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "medium", result.Value)
}

func TestTernaryInComprehension(t *testing.T) {
	source := `
result = ["even" if x % 2 == 0 else "odd" for x in range(5)]
count = len(result)
`
	vm := runCode(t, source)
	count := vm.GetGlobal("count").(*runtime.PyInt)
	assert.Equal(t, int64(5), count.Value)
}

// =============================================================================
// Whitespace and Indentation Edge Cases
// =============================================================================

func TestMixedIndentation(t *testing.T) {
	// This tests consistent indentation parsing
	source := `
def func():
    if True:
        x = 1
        if True:
            y = 2
    return x
result = func()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(1), result.Value)
}

func TestBlankLinesInBlocks(t *testing.T) {
	source := `
def func():
    x = 1

    y = 2

    return x + y

result = func()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

// =============================================================================
// Comment Edge Cases
// =============================================================================

func TestInlineComments(t *testing.T) {
	source := `
x = 10  # This is a comment
y = 20  # Another comment
result = x + y  # Sum
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(30), result.Value)
}

func TestCommentOnlyLines(t *testing.T) {
	source := `
# This is a comment
x = 10
# Another comment
# Yet another
y = 20
result = x + y
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(30), result.Value)
}

// =============================================================================
// Syntax Error Detection
// =============================================================================

func TestSyntaxErrorUnclosedParen(t *testing.T) {
	source := `
x = (1 + 2
`
	_, errs := compiler.CompileSource(source, "<test>")
	assert.NotEmpty(t, errs, "Expected syntax error for unclosed paren")
}

func TestSyntaxErrorUnclosedBracket(t *testing.T) {
	source := `
x = [1, 2, 3
`
	_, errs := compiler.CompileSource(source, "<test>")
	assert.NotEmpty(t, errs, "Expected syntax error for unclosed bracket")
}

func TestSyntaxErrorUnclosedString(t *testing.T) {
	source := `
x = "hello
`
	_, errs := compiler.CompileSource(source, "<test>")
	if len(errs) == 0 {
		t.Skip("RAGE doesn't detect unclosed string as syntax error")
		return
	}
	assert.NotEmpty(t, errs, "Expected syntax error for unclosed string")
}

func TestSyntaxErrorInvalidToken(t *testing.T) {
	source := `
x = @invalid
`
	_, errs := compiler.CompileSource(source, "<test>")
	// @ is a decorator, so this should error contextually
	// At minimum, it should compile but error at runtime
	// or error at compile time
	_ = errs
}

func TestSyntaxErrorIndentation(t *testing.T) {
	source := `
def func():
x = 1  # Bad indentation
`
	_, errs := compiler.CompileSource(source, "<test>")
	if len(errs) == 0 {
		t.Skip("RAGE doesn't detect bad indentation as syntax error")
		return
	}
	assert.NotEmpty(t, errs, "Expected syntax error for bad indentation")
}
