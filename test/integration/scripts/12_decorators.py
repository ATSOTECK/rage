# Test: Decorators and Closures
# Tests decorator syntax, closures, and nested functions

results = {}

# Basic closure
def make_counter():
    count = 0
    def counter():
        return count
    return counter

c = make_counter()
results["basic_closure"] = c()

# Closure with outer variable
def outer_with_val(x):
    def inner():
        return x * 2
    return inner

f = outer_with_val(21)
results["closure_captures_param"] = f()

# Nested closures
def outer_nested(x):
    def middle():
        def inner():
            return x
        return inner
    return middle

fn = outer_nested(42)()()
results["nested_closure"] = fn

# Basic decorator
def double_result(func):
    def wrapper():
        return func() * 2
    return wrapper

@double_result
def get_five():
    return 5

results["basic_decorator"] = get_five()

# Decorator with wrapped function args
def log_args(func):
    def wrapper(a, b):
        return func(a, b)
    return wrapper

@log_args
def add(a, b):
    return a + b

results["decorator_with_args"] = add(10, 20)

# Multiple decorators
def add_one(func):
    def wrapper():
        return func() + 1
    return wrapper

def double(func):
    def wrapper():
        return func() * 2
    return wrapper

@add_one
@double
def get_three():
    return 3

# (3 * 2) + 1 = 7
results["multiple_decorators"] = get_three()

# Decorator factory (decorator with arguments)
def repeat(n):
    def decorator(func):
        def wrapper():
            result = []
            for i in range(n):
                result.append(func())
            return result
        return wrapper
    return decorator

@repeat(3)
def say_hi():
    return "hi"

results["decorator_factory"] = say_hi()

# Decorator that wraps return value
def make_list(func):
    def wrapper():
        return [func()]
    return wrapper

@make_list
def get_value():
    return 42

results["wrapper_modifies_result"] = get_value()

# Closure with mutable state (via list)
def make_accumulator():
    total = [0]  # Use list for mutable state
    def add(x):
        total[0] = total[0] + x
        return total[0]
    return add

acc = make_accumulator()
acc(5)
acc(10)
results["closure_mutable_state"] = acc(3)

# Decorator that preserves function behavior
def identity_decorator(func):
    def wrapper(x):
        return func(x)
    return wrapper

@identity_decorator
def square(x):
    return x * x

results["identity_decorator"] = square(7)

# ============================================
# @property decorator tests
# ============================================

# Basic property getter
class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

c = Circle(5)
results["property_basic_getter"] = c.radius

# Property with setter
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

r = Rectangle(3, 4)
results["property_computed"] = r.area
r.width = 5
results["property_after_setter"] = r.area

# Property inheritance
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

sq = Square()
tr = Triangle()
results["property_override"] = sq.name
results["property_inherited"] = tr.name

# ============================================
# @classmethod decorator tests
# ============================================

# Basic classmethod
class Counter:
    count = 0

    @classmethod
    def increment(cls):
        cls.count = cls.count + 1
        return cls.count

    @classmethod
    def reset(cls):
        cls.count = 0

results["classmethod_call1"] = Counter.increment()
results["classmethod_call2"] = Counter.increment()
results["classmethod_on_instance"] = Counter().increment()
results["classmethod_final_count"] = Counter.count

Counter.reset()
results["classmethod_after_reset"] = Counter.count

# Classmethod as factory
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
results["classmethod_factory_x"] = p1.x
results["classmethod_factory_y"] = p1.y
results["classmethod_factory2_x"] = p2.x
results["classmethod_factory2_y"] = p2.y

# Classmethod with inheritance
class Animal:
    species = "Animal"

    @classmethod
    def get_species(cls):
        return cls.species

class Dog(Animal):
    species = "Dog"

class Cat(Animal):
    pass

results["classmethod_inherit_animal"] = Animal.get_species()
results["classmethod_inherit_dog"] = Dog.get_species()
results["classmethod_inherit_cat"] = Cat.get_species()

# ============================================
# @staticmethod decorator tests
# ============================================

# Basic staticmethod
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

# Call on class
results["staticmethod_add"] = MathUtils.add(2, 3)
results["staticmethod_multiply"] = MathUtils.multiply(4, 5)

# Call on instance
m = MathUtils()
results["staticmethod_on_instance"] = m.add(10, 20)
results["staticmethod_bool_true"] = MathUtils.is_positive(5)
results["staticmethod_bool_false"] = MathUtils.is_positive(-3)

# ============================================
# Mixed decorators in one class
# ============================================

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

obj = MyClass(5)
results["mixed_property_get"] = obj.x
obj.x = 15
results["mixed_property_set"] = obj.x
results["mixed_classmethod"] = MyClass.get_class_value()
results["mixed_staticmethod"] = MyClass.helper(10, 20)

print("Decorators tests completed")
