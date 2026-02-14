# Test: CPython Advanced Class Features
# Adapted from CPython's test_class.py, test_descr.py - covers advanced
# class patterns beyond 42_cpython_classes.py and 64_cpython_inheritance.py

from test_framework import test, expect

# =============================================================================
# Class variables shared across instances
# =============================================================================

class SharedState:
    instances = []
    count = 0

    def __init__(self, name):
        self.name = name
        SharedState.count = SharedState.count + 1
        SharedState.instances.append(name)

def test_class_variable_shared():
    SharedState.instances = []
    SharedState.count = 0
    a = SharedState("alpha")
    b = SharedState("beta")
    c = SharedState("gamma")
    expect(SharedState.count).to_be(3)
    expect(SharedState.instances).to_be(["alpha", "beta", "gamma"])
    expect(a.count).to_be(3)
    expect(b.count).to_be(3)

test("class_variable_shared", test_class_variable_shared)

# =============================================================================
# Class variable modification from instance vs class
# =============================================================================

class Tracker:
    value = 10

def test_class_variable_modification():
    a = Tracker()
    b = Tracker()
    expect(a.value).to_be(10)
    expect(b.value).to_be(10)
    # Modify via class - affects all instances
    Tracker.value = 20
    expect(a.value).to_be(20)
    expect(b.value).to_be(20)
    # Modify via instance - creates instance variable (shadows class var)
    a.value = 99
    expect(a.value).to_be(99)
    expect(b.value).to_be(20)
    expect(Tracker.value).to_be(20)
    # Restore
    Tracker.value = 10

test("class_variable_modification", test_class_variable_modification)

# =============================================================================
# Static methods
# =============================================================================

class MathUtils:
    @staticmethod
    def add(x, y):
        return x + y

    @staticmethod
    def multiply(x, y):
        return x * y

    @staticmethod
    def is_even(n):
        return n % 2 == 0

def test_static_methods():
    expect(MathUtils.add(3, 4)).to_be(7)
    expect(MathUtils.multiply(5, 6)).to_be(30)
    expect(MathUtils.is_even(4)).to_be(True)
    expect(MathUtils.is_even(7)).to_be(False)
    # Also callable from instances
    m = MathUtils()
    expect(m.add(10, 20)).to_be(30)

test("static_methods", test_static_methods)

# =============================================================================
# Class methods and factory methods
# =============================================================================

class Color:
    def __init__(self, r, g, b):
        self.r = r
        self.g = g
        self.b = b

    @classmethod
    def red(cls):
        return cls(255, 0, 0)

    @classmethod
    def green(cls):
        return cls(0, 255, 0)

    @classmethod
    def blue(cls):
        return cls(0, 0, 255)

    @classmethod
    def from_hex_list(cls, vals):
        return cls(vals[0], vals[1], vals[2])

    def to_list(self):
        return [self.r, self.g, self.b]

def test_classmethod_factory():
    r = Color.red()
    expect(r.r).to_be(255)
    expect(r.g).to_be(0)
    expect(r.b).to_be(0)
    g = Color.green()
    expect(g.to_list()).to_be([0, 255, 0])
    b = Color.blue()
    expect(b.to_list()).to_be([0, 0, 255])

test("classmethod_factory", test_classmethod_factory)

def test_classmethod_from_data():
    c = Color.from_hex_list([128, 64, 32])
    expect(c.r).to_be(128)
    expect(c.g).to_be(64)
    expect(c.b).to_be(32)

test("classmethod_from_data", test_classmethod_from_data)

# =============================================================================
# __str__ and __repr__ on custom classes
# =============================================================================

class Box:
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def __str__(self):
        return "Box(" + str(self.width) + "x" + str(self.height) + ")"

    def __repr__(self):
        return "Box(width=" + str(self.width) + ", height=" + str(self.height) + ")"

def test_str_repr_on_class():
    b = Box(10, 20)
    expect(str(b)).to_be("Box(10x20)")
    expect(repr(b)).to_be("Box(width=10, height=20)")

test("str_repr_on_class", test_str_repr_on_class)

# =============================================================================
# Dynamic attribute setting
# =============================================================================

class Flexible:
    pass

def test_dynamic_attributes():
    obj = Flexible()
    obj.x = 10
    obj.y = 20
    obj.name = "test"
    expect(obj.x).to_be(10)
    expect(obj.y).to_be(20)
    expect(obj.name).to_be("test")

test("dynamic_attributes", test_dynamic_attributes)

# =============================================================================
# hasattr / getattr / setattr patterns
# =============================================================================

class Config:
    def __init__(self):
        self.debug = False
        self.version = "1.0"

def test_hasattr_pattern():
    c = Config()
    expect(hasattr(c, "debug")).to_be(True)
    expect(hasattr(c, "version")).to_be(True)
    expect(hasattr(c, "missing")).to_be(False)

test("hasattr_pattern", test_hasattr_pattern)

def test_getattr_pattern():
    c = Config()
    expect(getattr(c, "debug")).to_be(False)
    expect(getattr(c, "version")).to_be("1.0")
    expect(getattr(c, "missing", "default")).to_be("default")

test("getattr_pattern", test_getattr_pattern)

def test_setattr_pattern():
    c = Config()
    setattr(c, "debug", True)
    expect(c.debug).to_be(True)
    setattr(c, "new_field", 42)
    expect(c.new_field).to_be(42)

test("setattr_pattern", test_setattr_pattern)

# =============================================================================
# Class decorator pattern
# =============================================================================

def add_greeting(cls):
    cls.greet = lambda self: "Hello from " + type(self).__name__
    return cls

@add_greeting
class Greeter:
    def __init__(self, name):
        self.name = name

def test_class_decorator():
    g = Greeter("test")
    expect(g.greet()).to_be("Hello from Greeter")
    expect(g.name).to_be("test")

test("class_decorator", test_class_decorator)

# =============================================================================
# Mixin with method resolution
# =============================================================================

class LogMixin:
    def get_log_name(self):
        return type(self).__name__

class ValidateMixin:
    def validate(self):
        return hasattr(self, "data")

class Service(LogMixin, ValidateMixin):
    def __init__(self, data):
        self.data = data

def test_mixin_method_resolution():
    s = Service("payload")
    expect(s.get_log_name()).to_be("Service")
    expect(s.validate()).to_be(True)
    expect(s.data).to_be("payload")

test("mixin_method_resolution", test_mixin_method_resolution)

# =============================================================================
# Property computed from other properties
# =============================================================================

class Rectangle:
    def __init__(self, w, h):
        self._w = w
        self._h = h

    @property
    def width(self):
        return self._w

    @property
    def height(self):
        return self._h

    @property
    def area(self):
        return self._w * self._h

    @property
    def perimeter(self):
        return 2 * (self._w + self._h)

    @property
    def is_square(self):
        return self._w == self._h

def test_property_computed_from_properties():
    r = Rectangle(3, 4)
    expect(r.width).to_be(3)
    expect(r.height).to_be(4)
    expect(r.area).to_be(12)
    expect(r.perimeter).to_be(14)
    expect(r.is_square).to_be(False)
    sq = Rectangle(5, 5)
    expect(sq.is_square).to_be(True)
    expect(sq.area).to_be(25)

test("property_computed_from_properties", test_property_computed_from_properties)

# =============================================================================
# __init__ with default values
# =============================================================================

class Connection:
    def __init__(self, host="localhost", port=8080, secure=False):
        self.host = host
        self.port = port
        self.secure = secure

    def url(self):
        protocol = "https" if self.secure else "http"
        return protocol + "://" + self.host + ":" + str(self.port)

def test_init_defaults():
    c = Connection()
    expect(c.host).to_be("localhost")
    expect(c.port).to_be(8080)
    expect(c.secure).to_be(False)
    expect(c.url()).to_be("http://localhost:8080")

test("init_defaults", test_init_defaults)

def test_init_override_defaults():
    c = Connection("example.com", 443, True)
    expect(c.host).to_be("example.com")
    expect(c.port).to_be(443)
    expect(c.secure).to_be(True)
    expect(c.url()).to_be("https://example.com:443")

test("init_override_defaults", test_init_override_defaults)

# =============================================================================
# Equality and hashing of class instances
# =============================================================================

class Token:
    def __init__(self, kind, value):
        self.kind = kind
        self.value = value

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self.kind == other.kind and self.value == other.value

    def __hash__(self):
        return hash(self.kind) + hash(self.value)

def test_class_equality():
    t1 = Token("ID", "x")
    t2 = Token("ID", "x")
    t3 = Token("NUM", "42")
    expect(t1 == t2).to_be(True)
    expect(t1 == t3).to_be(False)
    expect(t1 != t3).to_be(True)

test("class_equality", test_class_equality)

def test_class_hashing():
    t1 = Token("ID", "x")
    t2 = Token("ID", "x")
    expect(hash(t1) == hash(t2)).to_be(True)
    # Can use as dict keys
    d = {}
    d[t1] = "found"
    expect(d[t2]).to_be("found")

test("class_hashing", test_class_hashing)

# =============================================================================
# Comparison operators for sorting
# =============================================================================

class Student:
    def __init__(self, name, grade):
        self.name = name
        self.grade = grade

    def __lt__(self, other):
        return self.grade < other.grade

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self.grade == other.grade and self.name == other.name

    def __str__(self):
        return self.name + ":" + str(self.grade)

def test_comparison_for_sorting():
    students = [Student("Charlie", 85), Student("Alice", 92), Student("Bob", 78)]
    students.sort()
    names = []
    for s in students:
        names.append(s.name)
    expect(names).to_be(["Bob", "Charlie", "Alice"])

test("comparison_for_sorting", test_comparison_for_sorting)

# =============================================================================
# Classmethod on instance
# =============================================================================

class Counter:
    _count = 0

    @classmethod
    def increment(cls):
        cls._count = cls._count + 1
        return cls._count

    @classmethod
    def reset(cls):
        cls._count = 0

    @classmethod
    def get_count(cls):
        return cls._count

def test_classmethod_on_instance():
    Counter.reset()
    c = Counter()
    expect(c.increment()).to_be(1)
    expect(c.increment()).to_be(2)
    expect(Counter.get_count()).to_be(2)
    Counter.reset()
    expect(Counter.get_count()).to_be(0)

test("classmethod_on_instance", test_classmethod_on_instance)

# =============================================================================
# Object copying pattern (manual shallow copy)
# =============================================================================

class Settings:
    def __init__(self, theme, font_size, language):
        self.theme = theme
        self.font_size = font_size
        self.language = language

    def copy(self):
        return Settings(self.theme, self.font_size, self.language)

def test_object_copy_pattern():
    s1 = Settings("dark", 14, "en")
    s2 = s1.copy()
    expect(s2.theme).to_be("dark")
    expect(s2.font_size).to_be(14)
    s2.theme = "light"
    expect(s1.theme).to_be("dark")
    expect(s2.theme).to_be("light")

test("object_copy_pattern", test_object_copy_pattern)

# =============================================================================
# Method chaining pattern
# =============================================================================

class QueryBuilder:
    def __init__(self):
        self._table = ""
        self._conditions = []
        self._limit = None

    def from_table(self, table):
        self._table = table
        return self

    def where(self, condition):
        self._conditions.append(condition)
        return self

    def limit(self, n):
        self._limit = n
        return self

    def build(self):
        q = "SELECT FROM " + self._table
        if len(self._conditions) > 0:
            q = q + " WHERE " + " AND ".join(self._conditions)
        if self._limit is not None:
            q = q + " LIMIT " + str(self._limit)
        return q

def test_method_chaining():
    q = QueryBuilder().from_table("users").where("age>18").where("active=1").limit(10).build()
    expect(q).to_be("SELECT FROM users WHERE age>18 AND active=1 LIMIT 10")

test("method_chaining", test_method_chaining)

# =============================================================================
# Callable class instances (__call__)
# =============================================================================

class Validator:
    def __init__(self, min_val, max_val):
        self.min_val = min_val
        self.max_val = max_val

    def __call__(self, value):
        return self.min_val <= value <= self.max_val

def test_callable_instance():
    is_valid_age = Validator(0, 150)
    expect(is_valid_age(25)).to_be(True)
    expect(is_valid_age(-1)).to_be(False)
    expect(is_valid_age(200)).to_be(False)
    is_valid_score = Validator(0, 100)
    expect(is_valid_score(50)).to_be(True)
    expect(is_valid_score(101)).to_be(False)

test("callable_instance", test_callable_instance)

# =============================================================================
# Inheritance with classmethod
# =============================================================================

class Animal:
    kind = "animal"

    @classmethod
    def describe(cls):
        return cls.kind

class Dog(Animal):
    kind = "dog"

class Cat(Animal):
    kind = "cat"

def test_inherited_classmethod():
    expect(Animal.describe()).to_be("animal")
    expect(Dog.describe()).to_be("dog")
    expect(Cat.describe()).to_be("cat")

test("inherited_classmethod", test_inherited_classmethod)

# =============================================================================
# Class with __contains__
# =============================================================================

class NumberRange:
    def __init__(self, start, end):
        self.start = start
        self.end = end

    def __contains__(self, item):
        return self.start <= item <= self.end

def test_class_contains():
    r = NumberRange(1, 10)
    expect(5 in r).to_be(True)
    expect(0 in r).to_be(False)
    expect(10 in r).to_be(True)
    expect(11 in r).to_be(False)

test("class_contains", test_class_contains)

print("CPython class advanced tests completed")
