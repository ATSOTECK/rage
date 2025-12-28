package test

import (
	"testing"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicProperty tests a simple property getter
func TestBasicProperty(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

c = Circle(5)
result = c.radius
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, int64(5), result.(*runtime.PyInt).Value)
}

// TestPropertyWithSetter tests property with getter and setter
func TestPropertyWithSetter(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

    @radius.setter
    def radius(self, value):
        self._radius = value

c = Circle(5)
c.radius = 10
result = c.radius
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result := vm.GetGlobal("result")
	assert.Equal(t, int64(10), result.(*runtime.PyInt).Value)
}

// TestComputedProperty tests a property that computes a value
func TestComputedProperty(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Rectangle:
    def __init__(self, width, height):
        self._width = width
        self._height = height

    @property
    def area(self):
        return self._width * self._height

    @property
    def perimeter(self):
        return 2 * (self._width + self._height)

r = Rectangle(3, 4)
area = r.area
perimeter = r.perimeter
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	area := vm.GetGlobal("area")
	perimeter := vm.GetGlobal("perimeter")
	assert.Equal(t, int64(12), area.(*runtime.PyInt).Value)
	assert.Equal(t, int64(14), perimeter.(*runtime.PyInt).Value)
}

// TestClassmethod tests basic classmethod usage
func TestClassmethod(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Counter:
    count = 0

    @classmethod
    def increment(cls):
        cls.count = cls.count + 1
        return cls.count

result1 = Counter.increment()
result2 = Counter.increment()
result3 = Counter().increment()
final_count = Counter.count
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result1 := vm.GetGlobal("result1")
	result2 := vm.GetGlobal("result2")
	result3 := vm.GetGlobal("result3")
	finalCount := vm.GetGlobal("final_count")
	assert.Equal(t, int64(1), result1.(*runtime.PyInt).Value)
	assert.Equal(t, int64(2), result2.(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), result3.(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), finalCount.(*runtime.PyInt).Value)
}

// TestClassmethodFactory tests classmethod as factory method
func TestClassmethodFactory(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    @classmethod
    def origin(cls):
        return cls(0, 0)

    @classmethod
    def from_tuple(cls, t):
        return cls(t[0], t[1])

p1 = Point.origin()
p2 = Point.from_tuple((3, 4))
x1 = p1.x
y1 = p1.y
x2 = p2.x
y2 = p2.y
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	x1 := vm.GetGlobal("x1")
	y1 := vm.GetGlobal("y1")
	x2 := vm.GetGlobal("x2")
	y2 := vm.GetGlobal("y2")
	assert.Equal(t, int64(0), x1.(*runtime.PyInt).Value)
	assert.Equal(t, int64(0), y1.(*runtime.PyInt).Value)
	assert.Equal(t, int64(3), x2.(*runtime.PyInt).Value)
	assert.Equal(t, int64(4), y2.(*runtime.PyInt).Value)
}

// TestStaticmethod tests basic staticmethod usage
func TestStaticmethod(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Math:
    @staticmethod
    def add(a, b):
        return a + b

    @staticmethod
    def multiply(a, b):
        return a * b

# Call on class
result1 = Math.add(2, 3)
result2 = Math.multiply(4, 5)

# Call on instance
m = Math()
result3 = m.add(10, 20)
result4 = m.multiply(6, 7)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	result1 := vm.GetGlobal("result1")
	result2 := vm.GetGlobal("result2")
	result3 := vm.GetGlobal("result3")
	result4 := vm.GetGlobal("result4")
	assert.Equal(t, int64(5), result1.(*runtime.PyInt).Value)
	assert.Equal(t, int64(20), result2.(*runtime.PyInt).Value)
	assert.Equal(t, int64(30), result3.(*runtime.PyInt).Value)
	assert.Equal(t, int64(42), result4.(*runtime.PyInt).Value)
}

// TestStaticmethodUtility tests staticmethod as utility function
func TestStaticmethodUtility(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Validator:
    @staticmethod
    def is_positive(n):
        return n > 0

    @staticmethod
    def is_even(n):
        return n % 2 == 0

pos = Validator.is_positive(5)
neg = Validator.is_positive(-3)
even = Validator.is_even(4)
odd = Validator.is_even(7)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	pos := vm.GetGlobal("pos")
	neg := vm.GetGlobal("neg")
	even := vm.GetGlobal("even")
	odd := vm.GetGlobal("odd")
	assert.Equal(t, true, pos.(*runtime.PyBool).Value)
	assert.Equal(t, false, neg.(*runtime.PyBool).Value)
	assert.Equal(t, true, even.(*runtime.PyBool).Value)
	assert.Equal(t, false, odd.(*runtime.PyBool).Value)
}

// TestPropertyInheritance tests property inheritance
func TestPropertyInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Shape:
    @property
    def name(self):
        return "Shape"

class Circle(Shape):
    @property
    def name(self):
        return "Circle"

class Square(Shape):
    pass

c = Circle()
s = Square()
circle_name = c.name
square_name = s.name
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	circleName := vm.GetGlobal("circle_name")
	squareName := vm.GetGlobal("square_name")
	assert.Equal(t, "Circle", circleName.(*runtime.PyString).Value)
	assert.Equal(t, "Shape", squareName.(*runtime.PyString).Value)
}

// TestClassmethodInheritance tests classmethod with inheritance
func TestClassmethodInheritance(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class Animal:
    name = "Animal"

    @classmethod
    def get_name(cls):
        return cls.name

class Dog(Animal):
    name = "Dog"

class Cat(Animal):
    pass

# Test that cls is correctly bound to the calling class
animal_name = Animal.get_name()
dog_name = Dog.get_name()
cat_name = Cat.get_name()
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	animalName := vm.GetGlobal("animal_name")
	dogName := vm.GetGlobal("dog_name")
	catName := vm.GetGlobal("cat_name")
	assert.Equal(t, "Animal", animalName.(*runtime.PyString).Value)
	assert.Equal(t, "Dog", dogName.(*runtime.PyString).Value)
	assert.Equal(t, "Animal", catName.(*runtime.PyString).Value) // Cat inherits from Animal
}

// TestMixedDecorators tests using property, classmethod, and staticmethod together
func TestMixedDecorators(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class MyClass:
    _value = 100

    def __init__(self, x):
        self._x = x

    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value

    @classmethod
    def get_value(cls):
        return cls._value

    @staticmethod
    def helper(a, b):
        return a + b

obj = MyClass(5)
prop_result = obj.x
obj.x = 15
prop_result2 = obj.x
class_result = MyClass.get_value()
static_result = MyClass.helper(10, 20)
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	propResult := vm.GetGlobal("prop_result")
	propResult2 := vm.GetGlobal("prop_result2")
	classResult := vm.GetGlobal("class_result")
	staticResult := vm.GetGlobal("static_result")
	assert.Equal(t, int64(5), propResult.(*runtime.PyInt).Value)
	assert.Equal(t, int64(15), propResult2.(*runtime.PyInt).Value)
	assert.Equal(t, int64(100), classResult.(*runtime.PyInt).Value)
	assert.Equal(t, int64(30), staticResult.(*runtime.PyInt).Value)
}

// TestPropertyOnClassAccess tests that accessing property on class returns property object
func TestPropertyOnClassAccess(t *testing.T) {
	vm := runtime.NewVM()

	source := `
class MyClass:
    @property
    def get_value(self):
        return 42

# Accessing property on class should return property object, not invoke it
prop = MyClass.get_value
`
	code, errs := compiler.CompileSource(source, "<test>")
	require.Empty(t, errs)

	_, err := vm.Execute(code)
	require.NoError(t, err)

	prop := vm.GetGlobal("prop")
	// Property object should be returned when accessing on class
	_, ok := prop.(*runtime.PyProperty)
	assert.True(t, ok, "Expected property object, got %T", prop)
}
