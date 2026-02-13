# Test: Property and Descriptor Protocol
# Tests @property decorator, setters, deleters, and custom descriptors

from test_framework import test, expect

def test_basic_property():
    """Basic @property getter"""
    class Circle:
        def __init__(self, radius):
            self._radius = radius

        @property
        def radius(self):
            return self._radius

    c = Circle(5)
    expect(c.radius).to_be(5)

def test_property_setter():
    """@property with setter"""
    class Temperature:
        def __init__(self, celsius):
            self._celsius = celsius

        @property
        def celsius(self):
            return self._celsius

        @celsius.setter
        def celsius(self, value):
            if value < -273.15:
                raise ValueError("Temperature below absolute zero")
            self._celsius = value

        @property
        def fahrenheit(self):
            return self._celsius * 9/5 + 32

    t = Temperature(100)
    expect(t.celsius).to_be(100)
    expect(t.fahrenheit).to_be(212.0)

    t.celsius = 0
    expect(t.celsius).to_be(0)
    expect(t.fahrenheit).to_be(32.0)

def test_property_validation():
    """Property setter with validation"""
    class Person:
        def __init__(self, name, age):
            self.name = name
            self.age = age  # Uses the setter

        @property
        def age(self):
            return self._age

        @age.setter
        def age(self, value):
            if value < 0:
                raise ValueError("Age cannot be negative")
            self._age = value

    p = Person("Alice", 30)
    expect(p.age).to_be(30)

    p.age = 25
    expect(p.age).to_be(25)

    caught = False
    try:
        p.age = -1
    except ValueError:
        caught = True
    expect(caught).to_be(True)

def test_computed_property():
    """Property that computes a value"""
    class Rectangle:
        def __init__(self, width, height):
            self.width = width
            self.height = height

        @property
        def area(self):
            return self.width * self.height

        @property
        def perimeter(self):
            return 2 * (self.width + self.height)

    r = Rectangle(3, 4)
    expect(r.area).to_be(12)
    expect(r.perimeter).to_be(14)

    r.width = 5
    expect(r.area).to_be(20)
    expect(r.perimeter).to_be(18)

def test_property_inheritance():
    """Properties work with inheritance"""
    class Base:
        def __init__(self, value):
            self._value = value

        @property
        def value(self):
            return self._value

    class Child(Base):
        @property
        def doubled(self):
            return self.value * 2

    c = Child(10)
    expect(c.value).to_be(10)
    expect(c.doubled).to_be(20)

def test_readonly_property():
    """Property without setter is read-only"""
    class Const:
        @property
        def pi(self):
            return 3.14159

    c = Const()
    expect(c.pi).to_be(3.14159)

    caught = False
    try:
        c.pi = 3
    except (AttributeError, Exception):
        caught = True
    expect(caught).to_be(True)

test("basic_property", test_basic_property)
test("property_setter", test_property_setter)
test("property_validation", test_property_validation)
test("computed_property", test_computed_property)
test("property_inheritance", test_property_inheritance)
test("readonly_property", test_readonly_property)
