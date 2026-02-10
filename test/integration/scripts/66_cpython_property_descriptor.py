# Test: CPython Property and Descriptor Patterns
# Adapted from CPython's test_property.py and descriptor tests
# Covers @property, @x.setter, descriptors

from test_framework import test, expect

# =============================================================================
# Basic @property getter
# =============================================================================

class Circle:
    def __init__(self, radius):
        self._radius = radius

    @property
    def radius(self):
        return self._radius

    @property
    def area(self):
        return 3.14159 * self._radius * self._radius

    @property
    def diameter(self):
        return self._radius * 2

def test_property_getter():
    c = Circle(5)
    expect(c.radius).to_be(5)
    expect(c.diameter).to_be(10)

test("property_getter", test_property_getter)

def test_property_computed():
    c = Circle(10)
    area = c.area
    expect(area > 314.0).to_be(True)
    expect(area < 315.0).to_be(True)

test("property_computed", test_property_computed)

# =============================================================================
# @property with setter
# =============================================================================

class Temperature:
    def __init__(self, celsius):
        self._celsius = celsius

    @property
    def celsius(self):
        return self._celsius

    @celsius.setter
    def celsius(self, value):
        self._celsius = value

    @property
    def fahrenheit(self):
        return self._celsius * 9 / 5 + 32

def test_property_setter():
    t = Temperature(0)
    expect(t.celsius).to_be(0)
    t.celsius = 100
    expect(t.celsius).to_be(100)

test("property_setter", test_property_setter)

def test_property_computed_fahrenheit():
    t = Temperature(100)
    expect(t.fahrenheit).to_be(212.0)
    t.celsius = 0
    expect(t.fahrenheit).to_be(32.0)

test("property_computed_fahrenheit", test_property_computed_fahrenheit)

# =============================================================================
# Read-only property (no setter) - try to set raises an error
# =============================================================================

class ReadOnly:
    def __init__(self, x):
        self._x = x

    @property
    def x(self):
        return self._x

def test_property_read_only():
    r = ReadOnly(42)
    expect(r.x).to_be(42)
    # Writing to a read-only property should raise an error
    error_raised = False
    try:
        r.x = 99
    except Exception:
        error_raised = True
    expect(error_raised).to_be(True)
    # Value should be unchanged
    expect(r.x).to_be(42)

test("property_read_only", test_property_read_only)

# =============================================================================
# Property with validation (setter that validates)
# =============================================================================

class BoundedValue:
    def __init__(self, value, min_val, max_val):
        self._min = min_val
        self._max = max_val
        self._value = value

    @property
    def value(self):
        return self._value

    @value.setter
    def value(self, val):
        if val < self._min:
            self._value = self._min
        elif val > self._max:
            self._value = self._max
        else:
            self._value = val

def test_property_validation():
    b = BoundedValue(5, 0, 10)
    expect(b.value).to_be(5)
    b.value = 15
    expect(b.value).to_be(10)  # clamped to max
    b.value = -5
    expect(b.value).to_be(0)  # clamped to min
    b.value = 7
    expect(b.value).to_be(7)

test("property_validation", test_property_validation)

# =============================================================================
# Property caching pattern
# =============================================================================

class ExpensiveComputation:
    def __init__(self, data):
        self._data = data
        self._result = None

    @property
    def result(self):
        if self._result is None:
            # simulate expensive computation
            total = 0
            for item in self._data:
                total = total + item
            self._result = total
        return self._result

def test_property_caching():
    e = ExpensiveComputation([1, 2, 3, 4, 5])
    expect(e._result).to_be(None)
    expect(e.result).to_be(15)
    expect(e._result).to_be(15)
    # Second access uses cached value
    expect(e.result).to_be(15)

test("property_caching", test_property_caching)

# =============================================================================
# Multiple properties on same class
# =============================================================================

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
    def height(self):
        return self._height

    @height.setter
    def height(self, value):
        self._height = value

    @property
    def area(self):
        return self._width * self._height

    @property
    def perimeter(self):
        return 2 * (self._width + self._height)

def test_multiple_properties():
    r = Rectangle(3, 4)
    expect(r.width).to_be(3)
    expect(r.height).to_be(4)
    expect(r.area).to_be(12)
    expect(r.perimeter).to_be(14)
    r.width = 5
    expect(r.area).to_be(20)
    expect(r.perimeter).to_be(18)

test("multiple_properties", test_multiple_properties)

# =============================================================================
# Property inheritance
# =============================================================================

class PropBase:
    def __init__(self, x):
        self._x = x

    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value

class PropChild(PropBase):
    def __init__(self, x, y):
        PropBase.__init__(self, x)
        self._y = y

    @property
    def y(self):
        return self._y

def test_property_inheritance():
    c = PropChild(10, 20)
    expect(c.x).to_be(10)
    expect(c.y).to_be(20)
    c.x = 30
    expect(c.x).to_be(30)

test("property_inheritance", test_property_inheritance)

# =============================================================================
# Property override in subclass
# =============================================================================

class BaseConfig:
    def __init__(self):
        self._debug = False

    @property
    def debug(self):
        return self._debug

    @debug.setter
    def debug(self, value):
        self._debug = value

class StrictConfig(BaseConfig):
    @property
    def debug(self):
        return self._debug

    @debug.setter
    def debug(self, value):
        # Only allow True, never False once set
        if value:
            self._debug = True

def test_property_override():
    base = BaseConfig()
    expect(base.debug).to_be(False)
    base.debug = True
    expect(base.debug).to_be(True)
    base.debug = False
    expect(base.debug).to_be(False)

    strict = StrictConfig()
    expect(strict.debug).to_be(False)
    strict.debug = True
    expect(strict.debug).to_be(True)
    strict.debug = False  # should be ignored
    expect(strict.debug).to_be(True)

test("property_override", test_property_override)

# =============================================================================
# Property with default values
# =============================================================================

class Settings:
    def __init__(self):
        self._values = {}

    @property
    def timeout(self):
        if "timeout" in self._values:
            return self._values["timeout"]
        return 30  # default

    @timeout.setter
    def timeout(self, value):
        self._values["timeout"] = value

    @property
    def retries(self):
        if "retries" in self._values:
            return self._values["retries"]
        return 3  # default

    @retries.setter
    def retries(self, value):
        self._values["retries"] = value

def test_property_defaults():
    s = Settings()
    expect(s.timeout).to_be(30)
    expect(s.retries).to_be(3)
    s.timeout = 60
    expect(s.timeout).to_be(60)
    expect(s.retries).to_be(3)  # still default

test("property_defaults", test_property_defaults)

# =============================================================================
# Property with string formatting
# =============================================================================

class Color:
    def __init__(self, r, g, b):
        self._r = r
        self._g = g
        self._b = b

    @property
    def r(self):
        return self._r

    @property
    def g(self):
        return self._g

    @property
    def b(self):
        return self._b

    @property
    def hex_string(self):
        return "rgb(" + str(self._r) + "," + str(self._g) + "," + str(self._b) + ")"

def test_property_string_format():
    c = Color(255, 128, 0)
    expect(c.r).to_be(255)
    expect(c.g).to_be(128)
    expect(c.b).to_be(0)
    expect(c.hex_string).to_be("rgb(255,128,0)")

test("property_string_format", test_property_string_format)

# =============================================================================
# Property accessing other properties
# =============================================================================

class BMI:
    def __init__(self, weight_kg, height_m):
        self._weight = weight_kg
        self._height = height_m

    @property
    def weight(self):
        return self._weight

    @weight.setter
    def weight(self, value):
        self._weight = value

    @property
    def height(self):
        return self._height

    @property
    def bmi(self):
        return self._weight / (self._height * self._height)

    @property
    def category(self):
        b = self.bmi
        if b < 18.5:
            return "underweight"
        if b < 25.0:
            return "normal"
        return "overweight"

def test_property_using_other_property():
    person = BMI(70, 1.75)
    bmi = person.bmi
    expect(bmi > 22.0).to_be(True)
    expect(bmi < 23.0).to_be(True)
    expect(person.category).to_be("normal")
    person.weight = 50
    expect(person.category).to_be("underweight")

test("property_using_other_property", test_property_using_other_property)

# =============================================================================
# Property as computed attribute chain
# =============================================================================

class FullName:
    def __init__(self, first, last):
        self._first = first
        self._last = last

    @property
    def first(self):
        return self._first

    @first.setter
    def first(self, value):
        self._first = value

    @property
    def last(self):
        return self._last

    @last.setter
    def last(self, value):
        self._last = value

    @property
    def full(self):
        return self._first + " " + self._last

def test_property_chain():
    n = FullName("John", "Doe")
    expect(n.full).to_be("John Doe")
    n.first = "Jane"
    expect(n.full).to_be("Jane Doe")
    n.last = "Smith"
    expect(n.full).to_be("Jane Smith")

test("property_chain", test_property_chain)

# =============================================================================
# Setter and getter with access tracking
# =============================================================================

class AccessTracker:
    def __init__(self):
        self._value = None
        self._set_count = 0
        self._get_count = 0

    @property
    def value(self):
        self._get_count = self._get_count + 1
        return self._value

    @value.setter
    def value(self, val):
        self._set_count = self._set_count + 1
        self._value = val

def test_property_access_tracking():
    t = AccessTracker()
    t.value = 10
    t.value = 20
    t.value = 30
    _ = t.value
    _ = t.value
    expect(t._set_count).to_be(3)
    expect(t._get_count).to_be(2)

test("property_access_tracking", test_property_access_tracking)

# =============================================================================
# Property returning different types
# =============================================================================

class MultiType:
    def __init__(self, val):
        self._val = val

    @property
    def as_string(self):
        return str(self._val)

    @property
    def as_list(self):
        return [self._val]

    @property
    def as_dict(self):
        return {"value": self._val}

    @property
    def is_positive(self):
        return self._val > 0

def test_property_return_types():
    m = MultiType(42)
    expect(m.as_string).to_be("42")
    expect(m.as_list).to_be([42])
    expect(m.as_dict).to_be({"value": 42})
    expect(m.is_positive).to_be(True)

test("property_return_types", test_property_return_types)

# =============================================================================
# Property with list operations
# =============================================================================

class Stack:
    def __init__(self):
        self._items = []

    @property
    def size(self):
        return len(self._items)

    @property
    def top(self):
        if len(self._items) == 0:
            return None
        return self._items[len(self._items) - 1]

    @property
    def is_empty(self):
        return len(self._items) == 0

    def push(self, item):
        self._items.append(item)

    def pop(self):
        if len(self._items) > 0:
            result = self._items[len(self._items) - 1]
            self._items = self._items[0:len(self._items) - 1]
            return result
        return None

def test_property_stack():
    s = Stack()
    expect(s.is_empty).to_be(True)
    expect(s.size).to_be(0)
    expect(s.top).to_be(None)
    s.push(10)
    s.push(20)
    s.push(30)
    expect(s.is_empty).to_be(False)
    expect(s.size).to_be(3)
    expect(s.top).to_be(30)
    val = s.pop()
    expect(val).to_be(30)
    expect(s.top).to_be(20)
    expect(s.size).to_be(2)

test("property_stack", test_property_stack)

# =============================================================================
# Property with conditional logic
# =============================================================================

class ConditionalProp:
    def __init__(self):
        self._items = []

    @property
    def count(self):
        return len(self._items)

    @property
    def is_empty(self):
        return len(self._items) == 0

    @property
    def first(self):
        if len(self._items) > 0:
            return self._items[0]
        return None

    @property
    def last(self):
        if len(self._items) > 0:
            return self._items[len(self._items) - 1]
        return None

    def add(self, item):
        self._items.append(item)

def test_property_conditional():
    c = ConditionalProp()
    expect(c.is_empty).to_be(True)
    expect(c.first).to_be(None)
    expect(c.last).to_be(None)
    expect(c.count).to_be(0)
    c.add("a")
    c.add("b")
    c.add("c")
    expect(c.is_empty).to_be(False)
    expect(c.first).to_be("a")
    expect(c.last).to_be("c")
    expect(c.count).to_be(3)

test("property_conditional", test_property_conditional)

# =============================================================================
# Stacked property and method interaction
# =============================================================================

class Account:
    def __init__(self, balance):
        self._balance = balance
        self._transactions = []

    @property
    def balance(self):
        return self._balance

    @property
    def transaction_count(self):
        return len(self._transactions)

    def deposit(self, amount):
        self._balance = self._balance + amount
        self._transactions.append("deposit:" + str(amount))

    def withdraw(self, amount):
        if amount <= self._balance:
            self._balance = self._balance - amount
            self._transactions.append("withdraw:" + str(amount))
            return True
        return False

def test_property_with_methods():
    a = Account(100)
    expect(a.balance).to_be(100)
    expect(a.transaction_count).to_be(0)
    a.deposit(50)
    expect(a.balance).to_be(150)
    expect(a.transaction_count).to_be(1)
    result = a.withdraw(30)
    expect(result).to_be(True)
    expect(a.balance).to_be(120)
    expect(a.transaction_count).to_be(2)
    result2 = a.withdraw(200)
    expect(result2).to_be(False)
    expect(a.balance).to_be(120)

test("property_with_methods", test_property_with_methods)

print("CPython property/descriptor tests completed")
