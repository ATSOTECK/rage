package test

import (
	"strings"
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
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

// =============================================================================
// Cooperative Multiple Inheritance with super() Tests
// =============================================================================

// TestSuperZeroArgDiamond tests zero-argument super() in diamond inheritance
func TestSuperZeroArgDiamond(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return ["A"]

class B(A):
    def method(self):
        return ["B"] + super().method()

class C(A):
    def method(self):
        return ["C"] + super().method()

class D(B, C):
    def method(self):
        return ["D"] + super().method()

d = D()
result = d.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyList)
	// MRO: D -> B -> C -> A -> object
	// D.method() returns ["D"] + B.method()
	// B.method() returns ["B"] + C.method() (next in MRO after B is C, not A!)
	// C.method() returns ["C"] + A.method()
	// A.method() returns ["A"]
	// Result: ["D", "B", "C", "A"]
	require.Len(t, result.Items, 4)
	assert.Equal(t, "D", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "B", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "C", result.Items[2].(*runtime.PyString).Value)
	assert.Equal(t, "A", result.Items[3].(*runtime.PyString).Value)
}

// TestSuperCooperativeInit tests cooperative __init__ with super()
func TestSuperCooperativeInit(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    def __init__(self):
        self.base_init = True

class Mixin1(Base):
    def __init__(self):
        super().__init__()
        self.mixin1_init = True

class Mixin2(Base):
    def __init__(self):
        super().__init__()
        self.mixin2_init = True

class Final(Mixin1, Mixin2):
    def __init__(self):
        super().__init__()
        self.final_init = True

f = Final()
base_init = f.base_init
mixin1_init = f.mixin1_init
mixin2_init = f.mixin2_init
final_init = f.final_init
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	// All __init__ methods should have been called via cooperative super()
	assert.Equal(t, runtime.True, vm.GetGlobal("base_init"))
	assert.Equal(t, runtime.True, vm.GetGlobal("mixin1_init"))
	assert.Equal(t, runtime.True, vm.GetGlobal("mixin2_init"))
	assert.Equal(t, runtime.True, vm.GetGlobal("final_init"))
}

// TestSuperMethodChaining tests method chaining through super()
func TestSuperMethodChaining(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    def process(self, value):
        return value

class AddOne(Base):
    def process(self, value):
        return super().process(value) + 1

class Double(Base):
    def process(self, value):
        return super().process(value) * 2

class Combined(AddOne, Double):
    def process(self, value):
        return super().process(value)

c = Combined()
# MRO: Combined -> AddOne -> Double -> Base
# Combined.process(5) -> AddOne.process(5) -> Double.process(5+1=6) -> Base.process(6*2=12)
# Wait, let's trace more carefully:
# Combined.process(5) calls super().process(5) which is AddOne.process(5)
# AddOne.process(5) calls super().process(5) which is Double.process(5), then adds 1
# Double.process(5) calls super().process(5) which is Base.process(5), then doubles
# Base.process(5) returns 5
# So: Base returns 5, Double returns 5*2=10, AddOne returns 10+1=11
result = c.process(5)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// MRO: Combined -> AddOne -> Double -> Base
	// process(5): AddOne adds 1 to Double's result, Double doubles Base's result
	// Base.process(5) = 5
	// Double.process(5) = 5 * 2 = 10
	// AddOne.process(5) = 10 + 1 = 11
	assert.Equal(t, int64(11), result.(*runtime.PyInt).Value)
}

// TestSuperTwoArg tests two-argument super(Type, obj) form
func TestSuperTwoArg(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return "A"

class B(A):
    def method(self):
        return "B"

class C(B):
    def method(self):
        # Skip B, call A's method directly
        return super(B, self).method()

c = C()
result = c.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// super(B, self) starts searching from after B in MRO, so finds A.method
	assert.Equal(t, "A", result.(*runtime.PyString).Value)
}

// TestMROAttribute tests accessing __mro__ attribute
func TestMROAttribute(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    pass

class B(A):
    pass

class C(A):
    pass

class D(B, C):
    pass

mro = D.__mro__
mro_len = len(mro)
first_class = mro[0]
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	mroLen := vm.GetGlobal("mro_len")
	// MRO: D -> B -> C -> A -> object (5 classes)
	assert.Equal(t, int64(5), mroLen.(*runtime.PyInt).Value)

	firstClass := vm.GetGlobal("first_class")
	assert.Equal(t, "D", firstClass.(*runtime.PyClass).Name)
}

// TestSuperWithMultipleMethods tests super() with different methods
func TestSuperWithMultipleMethods(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    def method_a(self):
        return "Base.a"

    def method_b(self):
        return "Base.b"

class Child(Base):
    def method_a(self):
        return "Child.a:" + super().method_a()

    def method_b(self):
        return "Child.b:" + super().method_b()

c = Child()
result_a = c.method_a()
result_b = c.method_b()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	resultA := vm.GetGlobal("result_a")
	resultB := vm.GetGlobal("result_b")
	assert.Equal(t, "Child.a:Base.a", resultA.(*runtime.PyString).Value)
	assert.Equal(t, "Child.b:Base.b", resultB.(*runtime.PyString).Value)
}

// TestSuperDeepHierarchy tests super() in a deep hierarchy
func TestSuperDeepHierarchy(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Level0:
    def get_level(self):
        return 0

class Level1(Level0):
    def get_level(self):
        return super().get_level() + 1

class Level2(Level1):
    def get_level(self):
        return super().get_level() + 1

class Level3(Level2):
    def get_level(self):
        return super().get_level() + 1

class Level4(Level3):
    def get_level(self):
        return super().get_level() + 1

obj = Level4()
level = obj.get_level()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	level := vm.GetGlobal("level")
	assert.Equal(t, int64(4), level.(*runtime.PyInt).Value)
}

// TestSuperInNestedClass tests super() in nested class definitions
func TestSuperInNestedClass(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Outer:
    class Inner:
        def method(self):
            return "Inner"

    class InnerChild(Inner):
        def method(self):
            return "InnerChild:" + super().method()

obj = Outer.InnerChild()
result = obj.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "InnerChild:Inner", result.(*runtime.PyString).Value)
}

// TestSuperWithArgs tests super() with methods that have arguments
func TestSuperWithArgs(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    def calculate(self, x, y):
        return x + y

class Child(Base):
    def calculate(self, x, y):
        base_result = super().calculate(x, y)
        return base_result * 2

c = Child()
result = c.calculate(3, 4)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// (3 + 4) * 2 = 14
	assert.Equal(t, int64(14), result.(*runtime.PyInt).Value)
}

// TestSuperComplexDiamond tests a more complex diamond pattern
func TestSuperComplexDiamond(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return "A"

class B(A):
    def method(self):
        return "B->" + super().method()

class C(A):
    def method(self):
        return "C->" + super().method()

class D(A):
    def method(self):
        return "D->" + super().method()

class E(B, C):
    def method(self):
        return "E->" + super().method()

class F(C, D):
    def method(self):
        return "F->" + super().method()

class G(E, F):
    def method(self):
        return "G->" + super().method()

g = G()
result = g.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	// MRO for G: G -> E -> B -> F -> C -> D -> A -> object
	// Each method calls super(), so chain is: G->E->B->F->C->D->A
	assert.Equal(t, "G->E->B->F->C->D->A", result.(*runtime.PyString).Value)
}

// TestSuperInInitWithKwargs tests super().__init__ with keyword arguments
func TestSuperInInitWithKwargs(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    def __init__(self, name="default"):
        self.name = name

class Child(Base):
    def __init__(self, name, age):
        super().__init__(name=name)
        self.age = age

c = Child("Alice", 30)
name = c.name
age = c.age
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	name := vm.GetGlobal("name")
	age := vm.GetGlobal("age")
	assert.Equal(t, "Alice", name.(*runtime.PyString).Value)
	assert.Equal(t, int64(30), age.(*runtime.PyInt).Value)
}

// TestC3LinearizationOrder verifies C3 linearization produces correct MRO
func TestC3LinearizationOrder(t *testing.T) {
	vm := runtime.NewVM()

	source := `
# Classic example from Python docs
class O:
    pass

class A(O):
    pass

class B(O):
    pass

class C(O):
    pass

class D(O):
    pass

class E(O):
    pass

class K1(A, B, C):
    pass

class K2(D, B, E):
    pass

class K3(D, A):
    pass

class Z(K1, K2, K3):
    pass

# Get MRO names
mro_names = [cls.__name__ for cls in Z.__mro__]
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	mroNames := vm.GetGlobal("mro_names").(*runtime.PyList)
	// C3 linearization for Z(K1, K2, K3) should produce:
	// Z, K1, K2, K3, D, A, B, C, E, O, object
	expected := []string{"Z", "K1", "K2", "K3", "D", "A", "B", "C", "E", "O", "object"}
	require.Len(t, mroNames.Items, len(expected))
	for i, exp := range expected {
		assert.Equal(t, exp, mroNames.Items[i].(*runtime.PyString).Value,
			"MRO mismatch at position %d", i)
	}
}

// TestSuperPartialOverride tests super() when only some classes override a method
func TestSuperPartialOverride(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class A:
    def method(self):
        return ["A"]

class B(A):
    # B does NOT override method
    pass

class C(A):
    def method(self):
        return ["C"] + super().method()

class D(B, C):
    def method(self):
        return ["D"] + super().method()

d = D()
result = d.method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result").(*runtime.PyList)
	// MRO: D -> B -> C -> A
	// D.method() -> ["D"] + B's method (but B doesn't have one, so next is C)
	// Wait, super() in D will look for method in B first. B doesn't have it,
	// so it continues to C. C.method() -> ["C"] + A.method()
	// Result: ["D", "C", "A"]
	require.Len(t, result.Items, 3)
	assert.Equal(t, "D", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "C", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "A", result.Items[2].(*runtime.PyString).Value)
}

// TestSuperWithClassmethod tests super() in a classmethod
func TestSuperWithClassmethod(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Base:
    @classmethod
    def class_method(cls):
        return "Base"

class Child(Base):
    @classmethod
    def class_method(cls):
        return "Child:" + super(Child, cls).class_method()

result = Child.class_method()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, "Child:Base", result.(*runtime.PyString).Value)
}
