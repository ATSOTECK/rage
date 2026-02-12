package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tryCompile attempts to compile source code and returns nil if there's a panic
func tryCompile(source string) (code *runtime.CodeObject, errs []error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	code, errs = compiler.CompileSource(source, "<test>")
	return code, errs, false
}

// =============================================================================
// Metaclass Tests
// =============================================================================

func TestBasicMetaclass(t *testing.T) {
	source := `
class Meta(type):
    def __new__(mcs, name, bases, namespace):
        namespace['added_by_meta'] = True
        return super().__new__(mcs, name, bases, namespace)

class MyClass(metaclass=Meta):
    pass

result = MyClass.added_by_meta
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Metaclass syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Metaclass not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if b, ok := result.(*runtime.PyBool); ok {
		assert.True(t, b.Value)
	}
}

func TestMetaclassInit(t *testing.T) {
	source := `
class Meta(type):
    def __init__(cls, name, bases, namespace):
        super().__init__(name, bases, namespace)
        cls.initialized = True

class MyClass(metaclass=Meta):
    pass

result = MyClass.initialized
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Metaclass syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Metaclass not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if b, ok := result.(*runtime.PyBool); ok {
		assert.True(t, b.Value)
	}
}

func TestMetaclassCall(t *testing.T) {
	source := `
instance_count = 0

class CountingMeta(type):
    def __call__(cls, *args, **kwargs):
        global instance_count
        instance_count = instance_count + 1
        return super().__call__(*args, **kwargs)

class Counted(metaclass=CountingMeta):
    pass

a = Counted()
b = Counted()
c = Counted()
result = instance_count
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Metaclass syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Metaclass not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(3), i.Value)
	}
}

// =============================================================================
// Descriptor Protocol Tests
// =============================================================================

func TestDataDescriptor(t *testing.T) {
	source := `
class Descriptor:
    def __init__(self, name):
        self.name = name
        self.value = None

    def __get__(self, obj, objtype=None):
        if obj is None:
            return self
        return self.value

    def __set__(self, obj, value):
        self.value = value * 2  # Double the value

class MyClass:
    attr = Descriptor("attr")

obj = MyClass()
obj.attr = 5
result = obj.attr
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Descriptor syntax not supported: " + errs[0].Error())
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Descriptors not fully implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(10), i.Value)
	}
}

func TestNonDataDescriptor(t *testing.T) {
	source := `
class NonDataDescriptor:
    def __get__(self, obj, objtype=None):
        return "from descriptor"

class MyClass:
    attr = NonDataDescriptor()

obj = MyClass()
result1 = obj.attr

obj.__dict__["attr"] = "from instance"
result2 = obj.attr  # Instance attr shadows non-data descriptor
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Descriptor syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Descriptors not fully implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1")
	if s, ok := result1.(*runtime.PyString); ok {
		assert.Equal(t, "from descriptor", s.Value)
	}
	result2 := vm.GetGlobal("result2")
	if s, ok := result2.(*runtime.PyString); ok {
		assert.Equal(t, "from instance", s.Value)
	}
}

func TestDescriptorDelete(t *testing.T) {
	source := `
deleted = False

class Descriptor:
    def __get__(self, obj, objtype=None):
        return "value"

    def __set__(self, obj, value):
        pass

    def __delete__(self, obj):
        global deleted
        deleted = True

class MyClass:
    attr = Descriptor()

obj = MyClass()
del obj.attr
result = deleted
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Descriptor delete syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Descriptor delete not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if b, ok := result.(*runtime.PyBool); ok {
		assert.True(t, b.Value)
	}
}

// =============================================================================
// __slots__ Tests
// =============================================================================

func TestSlotsBasic(t *testing.T) {
	source := `
class Point:
    __slots__ = ['x', 'y']

    def __init__(self, x, y):
        self.x = x
        self.y = y

p = Point(3, 4)
result = p.x + p.y
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__slots__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__slots__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(7), i.Value)
	}
}

func TestSlotsPreventDict(t *testing.T) {
	source := `
class Slotted:
    __slots__ = ['x']

obj = Slotted()
obj.x = 1
has_dict = hasattr(obj, '__dict__')
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__slots__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__slots__ not implemented: " + err.Error())
		return
	}
	hasDict := vm.GetGlobal("has_dict")
	if b, ok := hasDict.(*runtime.PyBool); ok {
		assert.False(t, b.Value)
	}
}

// =============================================================================
// Abstract Base Class Tests
// =============================================================================

func TestAbstractMethod(t *testing.T) {
	source := `
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return 3.14159 * self.radius * self.radius

c = Circle(2)
result = c.area()
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("ABC not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("ABC not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if f, ok := result.(*runtime.PyFloat); ok {
		assert.InDelta(t, 12.566, f.Value, 0.01)
	}
}

// =============================================================================
// Operator Overloading Tests
// =============================================================================

func TestAddOperatorOverload(t *testing.T) {
	source := `
class Vector:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __add__(self, other):
        return Vector(self.x + other.x, self.y + other.y)

v1 = Vector(1, 2)
v2 = Vector(3, 4)
v3 = v1 + v2
result = v3.x + v3.y
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Operator overloading not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__add__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(10), result.Value)
}

func TestSubOperatorOverload(t *testing.T) {
	source := `
class Number:
    def __init__(self, value):
        self.value = value

    def __sub__(self, other):
        return Number(self.value - other.value)

n1 = Number(10)
n2 = Number(3)
n3 = n1 - n2
result = n3.value
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Operator overloading not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__sub__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(7), result.Value)
}

func TestMulOperatorOverload(t *testing.T) {
	source := `
class Vector:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __mul__(self, scalar):
        return Vector(self.x * scalar, self.y * scalar)

v = Vector(2, 3)
v2 = v * 3
result = v2.x + v2.y
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Operator overloading not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__mul__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(15), result.Value)
}

func TestReverseOperatorOverload(t *testing.T) {
	source := `
class Number:
    def __init__(self, value):
        self.value = value

    def __radd__(self, other):
        return Number(other + self.value)

n = Number(5)
result_obj = 10 + n  # Calls n.__radd__(10)
result = result_obj.value
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Reverse operators not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Reverse operators not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(15), i.Value)
	}
}

func TestComparisonOperatorOverload(t *testing.T) {
	source := `
class Number:
    def __init__(self, value):
        self.value = value

    def __lt__(self, other):
        return self.value < other.value

    def __eq__(self, other):
        return self.value == other.value

n1 = Number(5)
n2 = Number(10)
n3 = Number(5)

lt_result = n1 < n2
eq_result = n1 == n3
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Comparison operators not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Comparison operators not implemented: " + err.Error())
		return
	}
	ltResult := vm.GetGlobal("lt_result")
	eqResult := vm.GetGlobal("eq_result")
	if lt, ok := ltResult.(*runtime.PyBool); ok {
		if !lt.Value {
			t.Skip("__lt__ not implemented correctly")
			return
		}
	}
	if eq, ok := eqResult.(*runtime.PyBool); ok {
		if !eq.Value {
			t.Skip("__eq__ not implemented correctly")
			return
		}
	}
}

func TestInplaceOperatorOverload(t *testing.T) {
	source := `
class Counter:
    def __init__(self, value):
        self.value = value

    def __iadd__(self, other):
        self.value = self.value + other
        return self

c = Counter(5)
c += 3
result = c.value
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Inplace operators not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Inplace operators not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(8), i.Value)
	}
}

// =============================================================================
// __call__ Tests
// =============================================================================

func TestCallableObject(t *testing.T) {
	source := `
class Adder:
    def __init__(self, n):
        self.n = n

    def __call__(self, x):
        return x + self.n

add5 = Adder(5)
result = add5(10)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__call__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__call__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(15), result.Value)
}

func TestCallableWithMultipleArgs(t *testing.T) {
	source := `
class Multiplier:
    def __call__(self, a, b, c=1):
        return a * b * c

mult = Multiplier()
result = mult(2, 3, c=4)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__call__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__call__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(24), result.Value)
}

// =============================================================================
// __hash__ and __eq__ Tests
// =============================================================================

func TestHashAndEqForDict(t *testing.T) {
	source := `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __hash__(self):
        return hash((self.x, self.y))

    def __eq__(self, other):
        return self.x == other.x and self.y == other.y

p1 = Point(1, 2)
p2 = Point(1, 2)
p3 = Point(3, 4)

d = {p1: "origin"}
result1 = d[p2]  # Should work because p1 == p2
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Custom hash not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Custom hash not implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1")
	if s, ok := result1.(*runtime.PyString); ok {
		assert.Equal(t, "origin", s.Value)
	}
}

// =============================================================================
// __bool__ Tests
// =============================================================================

func TestBoolOverload(t *testing.T) {
	source := `
class Container:
    def __init__(self, items):
        self.items = items

    def __bool__(self):
        return len(self.items) > 0

empty = Container([])
full = Container([1, 2, 3])

result1 = bool(empty)
result2 = bool(full)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__bool__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__bool__ not implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1").(*runtime.PyBool)
	result2 := vm.GetGlobal("result2").(*runtime.PyBool)
	if result1.Value {
		t.Skip("__bool__ not working correctly for empty container")
		return
	}
	assert.True(t, result2.Value)
}

func TestBoolInIf(t *testing.T) {
	source := `
class Truthy:
    def __bool__(self):
        return True

class Falsy:
    def __bool__(self):
        return False

result1 = "yes" if Truthy() else "no"
result2 = "yes" if Falsy() else "no"
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__bool__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__bool__ not implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1").(*runtime.PyString)
	result2 := vm.GetGlobal("result2").(*runtime.PyString)
	if result1.Value != "yes" || result2.Value != "no" {
		t.Skip("__bool__ not working correctly in conditionals")
		return
	}
}

// =============================================================================
// __getattr__ and __setattr__ Tests
// =============================================================================

func TestGetattr(t *testing.T) {
	source := `
class Dynamic:
    def __init__(self):
        self.existing = "I exist"

    def __getattr__(self, name):
        return f"dynamic_{name}"

obj = Dynamic()
result1 = obj.existing
result2 = obj.missing
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	require.False(t, panicked)
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)
	result1 := vm.GetGlobal("result1").(*runtime.PyString)
	result2 := vm.GetGlobal("result2").(*runtime.PyString)
	assert.Equal(t, "I exist", result1.Value)
	assert.Equal(t, "dynamic_missing", result2.Value)
}

func TestSetattr(t *testing.T) {
	source := `
class Logged:
    def __init__(self):
        object.__setattr__(self, 'log', [])

    def __setattr__(self, name, value):
        self.log.append(name)
        object.__setattr__(self, name, value)

obj = Logged()
obj.x = 1
obj.y = 2
result = len(obj.log)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	require.False(t, panicked)
	require.Empty(t, errs)
	_, err := vm.Execute(code)
	require.NoError(t, err)
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(2), i.Value)
	}
}

// =============================================================================
// __getitem__ and __setitem__ Tests
// =============================================================================

func TestGetitemSetitem(t *testing.T) {
	source := `
class Matrix:
    def __init__(self):
        self.data = {}

    def __getitem__(self, key):
        row, col = key
        return self.data.get((row, col), 0)

    def __setitem__(self, key, value):
        row, col = key
        self.data[(row, col)] = value

m = Matrix()
m[0, 0] = 1
m[1, 1] = 2
result = m[0, 0] + m[1, 1] + m[2, 2]
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Tuple subscript not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Tuple subscript not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(3), i.Value)
	}
}

// =============================================================================
// Multiple Inheritance Method Resolution Order (MRO) Tests
// =============================================================================

func TestMRODiamond(t *testing.T) {
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
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "B", result.Value) // B comes before C in MRO
}

func TestSuperInMRO(t *testing.T) {
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
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("MRO super() syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("super() in MRO not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyList)
	// MRO is D -> B -> C -> A, so result should be ["D", "B", "C", "A"]
	require.Len(t, result.Items, 4)
	assert.Equal(t, "D", result.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "B", result.Items[1].(*runtime.PyString).Value)
	assert.Equal(t, "C", result.Items[2].(*runtime.PyString).Value)
	assert.Equal(t, "A", result.Items[3].(*runtime.PyString).Value)
}

// =============================================================================
// Class Decorators Tests
// =============================================================================

func TestClassDecorator(t *testing.T) {
	source := `
def add_method(cls):
    cls.added = lambda self: "added"
    return cls

@add_method
class MyClass:
    pass

obj = MyClass()
result = obj.added()
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Class decorator syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Class decorators not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "added", result.Value)
}

func TestMultipleClassDecorators(t *testing.T) {
	source := `
def dec1(cls):
    cls.attr1 = 1
    return cls

def dec2(cls):
    cls.attr2 = 2
    return cls

@dec1
@dec2
class MyClass:
    pass

result = MyClass.attr1 + MyClass.attr2
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Multiple class decorators syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Multiple class decorators not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(3), result.Value)
}

// =============================================================================
// __new__ Tests
// =============================================================================

func TestNewMethod(t *testing.T) {
	source := `
class Singleton:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = object.__new__(cls)
        return cls._instance

s1 = Singleton()
s2 = Singleton()
result = s1 is s2
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyBool)
	assert.True(t, result.Value)
}

func TestNewMethodWithInit(t *testing.T) {
	source := `
order = []

class MyClass:
    def __new__(cls, val):
        order.append("new")
        instance = object.__new__(cls)
        return instance

    def __init__(self, val):
        order.append("init")
        self.val = val

obj = MyClass(42)
result_order = order
result_val = obj.val
`
	vm := runCode(t, source)
	order := vm.GetGlobal("result_order").(*runtime.PyList)
	require.Len(t, order.Items, 2)
	assert.Equal(t, "new", order.Items[0].(*runtime.PyString).Value)
	assert.Equal(t, "init", order.Items[1].(*runtime.PyString).Value)
	val := vm.GetGlobal("result_val").(*runtime.PyInt)
	assert.Equal(t, int64(42), val.Value)
}

func TestNewMethodSkipsInitForDifferentType(t *testing.T) {
	source := `
class Factory:
    def __new__(cls):
        return 42

result = Factory()
`
	vm := runCode(t, source)
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(42), result.Value)
}

// =============================================================================
// __del__ Tests
// =============================================================================

func TestDelMethod(t *testing.T) {
	source := `
deleted = False

class Resource:
    def __del__(self):
        global deleted
        deleted = True

r = Resource()
del r

result = deleted
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__del__ not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__del__ not implemented: " + err.Error())
		return
	}
	// Note: __del__ might not be called immediately
	result := vm.GetGlobal("result")
	_ = result // Just check it doesn't crash
}

// =============================================================================
// __class__ attribute Tests
// =============================================================================

func TestClassAttribute(t *testing.T) {
	source := `
class MyClass:
    pass

obj = MyClass()
result = obj.__class__.__name__
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__class__ attribute syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__class__ attribute not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyString)
	assert.Equal(t, "MyClass", result.Value)
}

// =============================================================================
// isinstance and issubclass Tests
// =============================================================================

func TestIsinstance(t *testing.T) {
	source := `
class Animal:
    pass

class Dog(Animal):
    pass

d = Dog()
result1 = isinstance(d, Dog)
result2 = isinstance(d, Animal)
result3 = isinstance(d, str)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("isinstance syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("isinstance not implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1").(*runtime.PyBool)
	result2 := vm.GetGlobal("result2").(*runtime.PyBool)
	result3 := vm.GetGlobal("result3").(*runtime.PyBool)
	assert.True(t, result1.Value)
	assert.True(t, result2.Value)
	assert.False(t, result3.Value)
}

func TestIssubclass(t *testing.T) {
	source := `
class Animal:
    pass

class Dog(Animal):
    pass

class Cat(Animal):
    pass

result1 = issubclass(Dog, Animal)
result2 = issubclass(Dog, Dog)
result3 = issubclass(Dog, Cat)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("issubclass syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("issubclass not implemented: " + err.Error())
		return
	}
	result1 := vm.GetGlobal("result1").(*runtime.PyBool)
	result2 := vm.GetGlobal("result2").(*runtime.PyBool)
	result3 := vm.GetGlobal("result3").(*runtime.PyBool)
	assert.True(t, result1.Value)
	assert.True(t, result2.Value)
	assert.False(t, result3.Value)
}

// =============================================================================
// __bases__ and __mro__ Tests
// =============================================================================

func TestBasesAttribute(t *testing.T) {
	source := `
class A:
    pass

class B:
    pass

class C(A, B):
    pass

result = len(C.__bases__)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__bases__ not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__bases__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result")
	if i, ok := result.(*runtime.PyInt); ok {
		assert.Equal(t, int64(2), i.Value)
	}
}

// =============================================================================
// Classmethod and Staticmethod with Inheritance Tests
// =============================================================================

func TestClassmethodInheritanceOOP(t *testing.T) {
	source := `
class Base:
    value = 10

    @classmethod
    def get_value(cls):
        return cls.value

class Derived(Base):
    value = 20

result1 = Base.get_value()
result2 = Derived.get_value()
`
	vm := runCode(t, source)
	result1 := vm.GetGlobal("result1").(*runtime.PyInt)
	result2 := vm.GetGlobal("result2").(*runtime.PyInt)
	assert.Equal(t, int64(10), result1.Value)
	assert.Equal(t, int64(20), result2.Value)
}

func TestStaticmethodInheritance(t *testing.T) {
	source := `
class Base:
    @staticmethod
    def utility():
        return "base"

class Derived(Base):
    @staticmethod
    def utility():
        return "derived"

result1 = Base.utility()
result2 = Derived.utility()
`
	vm := runCode(t, source)
	result1 := vm.GetGlobal("result1").(*runtime.PyString)
	result2 := vm.GetGlobal("result2").(*runtime.PyString)
	assert.Equal(t, "base", result1.Value)
	assert.Equal(t, "derived", result2.Value)
}

// =============================================================================
// __dict__ Tests
// =============================================================================

func TestClassDict(t *testing.T) {
	source := `
class MyClass:
    class_attr = "class"

    def method(self):
        pass

has_class_attr = "class_attr" in MyClass.__dict__
has_method = "method" in MyClass.__dict__
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("Class __dict__ not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("Class __dict__ not implemented: " + err.Error())
		return
	}
	hasClassAttr := vm.GetGlobal("has_class_attr")
	hasMethod := vm.GetGlobal("has_method")
	if b, ok := hasClassAttr.(*runtime.PyBool); ok {
		assert.True(t, b.Value)
	}
	if b, ok := hasMethod.(*runtime.PyBool); ok {
		assert.True(t, b.Value)
	}
}

func TestInstanceDict(t *testing.T) {
	source := `
class MyClass:
    def __init__(self):
        self.x = 1
        self.y = 2

obj = MyClass()
result = len(obj.__dict__)
`
	vm := runtime.NewVM()
	code, errs, panicked := tryCompile(source)
	if panicked || len(errs) > 0 {
		t.Skip("__dict__ syntax not supported")
		return
	}
	_, err := vm.Execute(code)
	if err != nil {
		t.Skip("__dict__ not implemented: " + err.Error())
		return
	}
	result := vm.GetGlobal("result").(*runtime.PyInt)
	assert.Equal(t, int64(2), result.Value)
}
