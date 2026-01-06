# Test: Decorators and Closures
# Tests decorator syntax, closures, and nested functions

# Closure and decorator helpers at module level
def make_counter():
    count = 0
    def counter():
        return count
    return counter

def outer_with_val(x):
    def inner():
        return x * 2
    return inner

def outer_nested(x):
    def middle():
        def inner():
            return x
        return inner
    return middle

def double_result(func):
    def wrapper():
        return func() * 2
    return wrapper

def log_args(func):
    def wrapper(a, b):
        return func(a, b)
    return wrapper

def add_one(func):
    def wrapper():
        return func() + 1
    return wrapper

def double(func):
    def wrapper():
        return func() * 2
    return wrapper

def repeat(n):
    def decorator(func):
        def wrapper():
            result = []
            for i in range(n):
                result.append(func())
            return result
        return wrapper
    return decorator

def make_list(func):
    def wrapper():
        return [func()]
    return wrapper

def make_accumulator():
    total = [0]
    def add(x):
        total[0] = total[0] + x
        return total[0]
    return add

def identity_decorator(func):
    def wrapper(x):
        return func(x)
    return wrapper

@double_result
def get_five():
    return 5

@log_args
def add(a, b):
    return a + b

@add_one
@double
def get_three():
    return 3

@repeat(3)
def say_hi():
    return "hi"

@make_list
def get_value():
    return 42

@identity_decorator
def square(x):
    return x * x

class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

class Rectangle:
    def __init__(self, width, height):
        self._width = width
        self._height = height

    @property
    def width(self):
        return self._width

    @width.setter
    def width(self, value):
        self._width = value

    @property
    def area(self):
        return self._width * self._height

class Shape:
    @property
    def name(self):
        return "Shape"

class Square(Shape):
    @property
    def name(self):
        return "Square"

class Triangle(Shape):
    pass

class Counter:
    count = 0

    @classmethod
    def increment(cls):
        cls.count = cls.count + 1
        return cls.count

    @classmethod
    def reset(cls):
        cls.count = 0

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

class Animal:
    species = "Animal"

    @classmethod
    def get_species(cls):
        return cls.species

class Dog(Animal):
    species = "Dog"

class Cat(Animal):
    pass

class MathUtils:
    @staticmethod
    def add(a, b):
        return a + b

    @staticmethod
    def multiply(a, b):
        return a * b

    @staticmethod
    def is_positive(n):
        return n > 0

class MyClass:
    class_value = 100

    def __init__(self, x):
        self._x = x

    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value

    @classmethod
    def get_class_value(cls):
        return cls.class_value

    @staticmethod
    def helper(a, b):
        return a + b

def test_basic_closure():
    c = make_counter()
    expect(0, c())

def test_closure_captures_param():
    f = outer_with_val(21)
    expect(42, f())

def test_nested_closure():
    fn = outer_nested(42)()()
    expect(42, fn)

def test_basic_decorator():
    expect(10, get_five())

def test_decorator_with_args():
    expect(30, add(10, 20))

def test_multiple_decorators():
    # (3 * 2) + 1 = 7
    expect(7, get_three())

def test_decorator_factory():
    expect(["hi", "hi", "hi"], say_hi())

def test_wrapper_modifies_result():
    expect([42], get_value())

def test_closure_mutable_state():
    acc = make_accumulator()
    acc(5)
    acc(10)
    expect(18, acc(3))

def test_identity_decorator():
    expect(49, square(7))

def test_property_basic_getter():
    c = Circle(5)
    expect(5, c.radius)

def test_property_computed():
    r = Rectangle(3, 4)
    expect(12, r.area)

def test_property_after_setter():
    r = Rectangle(3, 4)
    r.width = 5
    expect(20, r.area)

def test_property_override():
    sq = Square()
    expect("Square", sq.name)

def test_property_inherited():
    tr = Triangle()
    expect("Shape", tr.name)

def test_classmethod():
    Counter.reset()
    expect(1, Counter.increment())
    expect(2, Counter.increment())
    expect(3, Counter().increment())
    expect(3, Counter.count)
    Counter.reset()
    expect(0, Counter.count)

def test_classmethod_factory():
    p1 = Point.origin()
    p2 = Point.from_tuple((3, 4))
    expect(0, p1.x)
    expect(0, p1.y)
    expect(3, p2.x)
    expect(4, p2.y)

def test_classmethod_inheritance():
    expect("Animal", Animal.get_species())
    expect("Dog", Dog.get_species())
    expect("Animal", Cat.get_species())

def test_staticmethod():
    expect(5, MathUtils.add(2, 3))
    expect(20, MathUtils.multiply(4, 5))

def test_staticmethod_on_instance():
    m = MathUtils()
    expect(30, m.add(10, 20))
    expect(True, MathUtils.is_positive(5))
    expect(False, MathUtils.is_positive(-3))

def test_mixed_decorators():
    obj = MyClass(5)
    expect(5, obj.x)
    obj.x = 15
    expect(15, obj.x)
    expect(100, MyClass.get_class_value())
    expect(30, MyClass.helper(10, 20))

test("basic_closure", test_basic_closure)
test("closure_captures_param", test_closure_captures_param)
test("nested_closure", test_nested_closure)
test("basic_decorator", test_basic_decorator)
test("decorator_with_args", test_decorator_with_args)
test("multiple_decorators", test_multiple_decorators)
test("decorator_factory", test_decorator_factory)
test("wrapper_modifies_result", test_wrapper_modifies_result)
test("closure_mutable_state", test_closure_mutable_state)
test("identity_decorator", test_identity_decorator)
test("property_basic_getter", test_property_basic_getter)
test("property_computed", test_property_computed)
test("property_after_setter", test_property_after_setter)
test("property_override", test_property_override)
test("property_inherited", test_property_inherited)
test("classmethod", test_classmethod)
test("classmethod_factory", test_classmethod_factory)
test("classmethod_inheritance", test_classmethod_inheritance)
test("staticmethod", test_staticmethod)
test("staticmethod_on_instance", test_staticmethod_on_instance)
test("mixed_decorators", test_mixed_decorators)

print("Decorators tests completed")
