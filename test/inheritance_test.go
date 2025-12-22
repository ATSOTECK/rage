package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/RAGE/internal/compiler"
	"github.com/ATSOTECK/RAGE/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicMultipleInheritance tests basic multiple inheritance
func TestBasicMultipleInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def greet(self):
        return "Hello from A"

class B:
    def farewell(self):
        return "Goodbye from B"

class C(A, B):
    pass

c = C()
greet_result = c.greet()
farewell_result = c.farewell()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	greet := vm.GetGlobal("greet_result")
	farewell := vm.GetGlobal("farewell_result")
	assert.Equal(t, "Hello from A", greet.(*runtime.PyString).Value)
	assert.Equal(t, "Goodbye from B", farewell.(*runtime.PyString).Value)
}

// TestDiamondInheritance tests the classic diamond problem
func TestDiamondInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return "A"

class B(A):
    def method(self):
        return "B"

class C(A):
    def method(self):
        return "C"

class D(B, C):
    pass

d = D()
result = d.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// D's MRO should be [D, B, C, A], so method should return "B"
	assert.Equal(t, "B", result.(*runtime.PyString).Value)
}

// TestDiamondInheritanceC tests calling method from C in diamond
func TestDiamondInheritanceC(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return "A"

class B(A):
    pass  # No override, should use A's method through C

class C(A):
    def method(self):
        return "C"

class D(B, C):
    pass

d = D()
result = d.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// D's MRO should be [D, B, C, A], B doesn't override, so C's method is used
	assert.Equal(t, "C", result.(*runtime.PyString).Value)
}

// TestDeepDiamondInheritance tests deeper diamond inheritance
func TestDeepDiamondInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    value = "A"

class B(A):
    value = "B"

class C(A):
    value = "C"

class D(B, C):
    pass

class E(C, B):
    pass

d_value = D.value
e_value = E.value
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	dValue := vm.GetGlobal("d_value")
	eValue := vm.GetGlobal("e_value")
	// D(B, C): MRO = [D, B, C, A], so value = "B"
	// E(C, B): MRO = [E, C, B, A], so value = "C"
	assert.Equal(t, "B", dValue.(*runtime.PyString).Value)
	assert.Equal(t, "C", eValue.(*runtime.PyString).Value)
}

// TestMultipleInheritanceMethodOverride tests method override with multiple inheritance
func TestMultipleInheritanceMethodOverride(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return "A"

class B:
    def method(self):
        return "B"

class C(A, B):
    pass

class D(B, A):
    pass

c = C()
d = D()
c_result = c.method()
d_result = d.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	cResult := vm.GetGlobal("c_result")
	dResult := vm.GetGlobal("d_result")
	// C(A, B): MRO = [C, A, B], so method = "A"
	// D(B, A): MRO = [D, B, A], so method = "B"
	assert.Equal(t, "A", cResult.(*runtime.PyString).Value)
	assert.Equal(t, "B", dResult.(*runtime.PyString).Value)
}

// TestMultipleInheritanceAttributes tests attribute access with multiple inheritance
func TestMultipleInheritanceAttributes(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    x = 1

class B:
    y = 2

class C:
    z = 3

class D(A, B, C):
    pass

d = D()
x = d.x
y = d.y
z = d.z
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	x := vm.GetGlobal("x")
	y := vm.GetGlobal("y")
	z := vm.GetGlobal("z")
	assert.Equal(t, int64(1), x.(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), y.(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), z.(*runtime.PyInt).Value)
}

// TestMROWithSingleInheritance tests that single inheritance still works
func TestMROWithSingleInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Animal:
    def speak(self):
        return "..."

class Dog(Animal):
    def speak(self):
        return "Woof"

class Puppy(Dog):
    pass

p = Puppy()
result = p.speak()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "Woof", result.(*runtime.PyString).Value)
}

// TestInconsistentMROError tests that inconsistent MRO is detected
func TestInconsistentMROError(t *testing.T) {
	vm := runtime.NewVM()

	// This should fail because there's no consistent MRO
	// class A(B): pass; class B(A): pass - circular
	// A simpler case: class X(A, B) and class Y(B, A) where A and B are unrelated,
	// then class Z(X, Y) creates inconsistency
	source := `
class A:
    pass

class B:
    pass

class X(A, B):
    pass

class Y(B, A):
    pass

class Z(X, Y):
    pass
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Cannot create a consistent method resolution order"))
}

// TestThreeWayInheritance tests inheritance from three classes
func TestThreeWayInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def a(self):
        return "A"

class B:
    def b(self):
        return "B"

class C:
    def c(self):
        return "C"

class D(A, B, C):
    def d(self):
        return "D"

obj = D()
results = [obj.a(), obj.b(), obj.c(), obj.d()]
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	results := vm.GetGlobal("results").(*runtime.PyList)
	assert.Equal(t, 4, len(results.Items))
	assert.Equal(t, "A", results.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "B", results.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "C", results.Items[2].(*runtime.PyString).Value)
	assert.Equal(t, "D", results.Items[3].(*runtime.PyString).Value)
}

// TestMixinPattern tests the mixin design pattern
func TestMixinPattern(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class JsonMixin:
    def to_json(self):
        return "json"

class XmlMixin:
    def to_xml(self):
        return "xml"

class DataModel:
    def __init__(self):
        self.data = "data"

class MyModel(DataModel, JsonMixin, XmlMixin):
    pass

m = MyModel()
json_result = m.to_json()
xml_result = m.to_xml()
data_result = m.data
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	jsonResult := vm.GetGlobal("json_result")
	xmlResult := vm.GetGlobal("xml_result")
	dataResult := vm.GetGlobal("data_result")
	assert.Equal(t, "json", jsonResult.(*runtime.PyString).Value)
	assert.Equal(t, "xml", xmlResult.(*runtime.PyString).Value)
	assert.Equal(t, "data", dataResult.(*runtime.PyString).Value)
}
