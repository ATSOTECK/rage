# Test: CPython Inheritance Patterns
# Adapted from CPython class/inheritance tests - covers inheritance patterns
# beyond 05_classes.py and 24_multiple_inheritance.py

from test_framework import test, expect

# =============================================================================
# Basic single inheritance
# =============================================================================

class Vehicle:
    def __init__(self, make, model):
        self.make = make
        self.model = model

    def description(self):
        return self.make + " " + self.model

    def vehicle_type(self):
        return "vehicle"

class Car(Vehicle):
    def __init__(self, make, model, doors):
        Vehicle.__init__(self, make, model)
        self.doors = doors

    def vehicle_type(self):
        return "car"

class Truck(Vehicle):
    def __init__(self, make, model, payload):
        Vehicle.__init__(self, make, model)
        self.payload = payload

    def vehicle_type(self):
        return "truck"

class ElectricCar(Car):
    def __init__(self, make, model, doors, battery):
        Car.__init__(self, make, model, doors)
        self.battery = battery

    def vehicle_type(self):
        return "electric car"

# =============================================================================
# Multiple inheritance / MRO
# =============================================================================

class Printable:
    def to_string(self):
        return "Printable"

class Serializable:
    def serialize(self):
        return "serialized"

class Document(Printable, Serializable):
    def __init__(self, title):
        self.title = title

    def to_string(self):
        return "Document: " + self.title

# Diamond inheritance
class Base:
    def identify(self):
        return "Base"

class Left(Base):
    def identify(self):
        return "Left"

class Right(Base):
    def identify(self):
        return "Right"

class Bottom(Left, Right):
    pass

# =============================================================================
# Attribute inheritance
# =============================================================================

class Defaults:
    color = "red"
    size = 10
    items = []

    def get_color(self):
        return self.color

class CustomDefaults(Defaults):
    color = "blue"

class MoreCustom(CustomDefaults):
    size = 20

# =============================================================================
# Overriding dunder methods
# =============================================================================

class Vector:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __str__(self):
        return "Vector(" + str(self.x) + ", " + str(self.y) + ")"

    def __repr__(self):
        return "Vector(x=" + str(self.x) + ", y=" + str(self.y) + ")"

    def __eq__(self, other):
        if type(other).__name__ != "Vector":
            return False
        return self.x == other.x and self.y == other.y

    def __lt__(self, other):
        # Compare by magnitude squared (avoid sqrt)
        return (self.x * self.x + self.y * self.y) < (other.x * other.x + other.y * other.y)

    def __len__(self):
        # Return manhattan distance as int
        result = self.x + self.y
        if result < 0:
            result = -result
        return result

    def __getitem__(self, index):
        if index == 0:
            return self.x
        if index == 1:
            return self.y
        raise IndexError("Vector index out of range")

    def __add__(self, other):
        return Vector(self.x + other.x, self.y + other.y)

class Vector3D(Vector):
    def __init__(self, x, y, z):
        Vector.__init__(self, x, y)
        self.z = z

    def __str__(self):
        return "Vector3D(" + str(self.x) + ", " + str(self.y) + ", " + str(self.z) + ")"

    def __eq__(self, other):
        if type(other).__name__ != "Vector3D":
            return False
        return self.x == other.x and self.y == other.y and self.z == other.z

    def __getitem__(self, index):
        if index == 0:
            return self.x
        if index == 1:
            return self.y
        if index == 2:
            return self.z
        raise IndexError("Vector3D index out of range")

# =============================================================================
# Abstract-like pattern
# =============================================================================

class Shape:
    def area(self):
        raise NotImplementedError("Subclass must implement area()")

    def perimeter(self):
        raise NotImplementedError("Subclass must implement perimeter()")

    def describe(self):
        return "Shape with area " + str(self.area())

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return 3.14159 * self.radius * self.radius

    def perimeter(self):
        return 2 * 3.14159 * self.radius

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):
        return self.width * self.height

    def perimeter(self):
        return 2 * (self.width + self.height)

class Square(Rectangle):
    def __init__(self, side):
        Rectangle.__init__(self, side, side)

# =============================================================================
# Mixin pattern
# =============================================================================

class JsonMixin:
    def to_json_like(self):
        # Simple dict-like representation
        result = {}
        for key in self._fields():
            result[key] = getattr(self, key)
        return result

class ComparableMixin:
    def is_equal(self, other):
        return self._compare_key() == other._compare_key()

    def is_less(self, other):
        return self._compare_key() < other._compare_key()

class Person(JsonMixin, ComparableMixin):
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def _fields(self):
        return ["name", "age"]

    def _compare_key(self):
        return self.age

# =============================================================================
# Class attributes vs instance attributes
# =============================================================================

class Config:
    debug = False
    version = "1.0"
    max_retries = 3

    def __init__(self, name):
        self.name = name

class AppConfig(Config):
    debug = True

# =============================================================================
# issubclass-like patterns
# =============================================================================

class Animal:
    kind = "animal"

class Mammal(Animal):
    kind = "mammal"

class Reptile(Animal):
    kind = "reptile"

class Dog(Mammal):
    kind = "dog"

class Cat(Mammal):
    kind = "cat"

class Snake(Reptile):
    kind = "snake"

# =============================================================================
# Tests
# =============================================================================

def test_basic_single_inheritance():
    car = Car("Toyota", "Camry", 4)
    expect(car.make).to_be("Toyota")
    expect(car.model).to_be("Camry")
    expect(car.doors).to_be(4)
    expect(car.description()).to_be("Toyota Camry")
    expect(car.vehicle_type()).to_be("car")

def test_method_override():
    v = Vehicle("Generic", "Vehicle")
    c = Car("Honda", "Civic", 4)
    t = Truck("Ford", "F150", 1000)
    expect(v.vehicle_type()).to_be("vehicle")
    expect(c.vehicle_type()).to_be("car")
    expect(t.vehicle_type()).to_be("truck")

def test_calling_parent_method():
    car = Car("BMW", "X5", 4)
    # description() is inherited from Vehicle
    expect(car.description()).to_be("BMW X5")

def test_deep_inheritance_chain():
    ec = ElectricCar("Tesla", "Model 3", 4, 75)
    expect(ec.make).to_be("Tesla")
    expect(ec.model).to_be("Model 3")
    expect(ec.doors).to_be(4)
    expect(ec.battery).to_be(75)
    expect(ec.vehicle_type()).to_be("electric car")
    expect(ec.description()).to_be("Tesla Model 3")

def test_multiple_inheritance_basic():
    doc = Document("My Report")
    expect(doc.to_string()).to_be("Document: My Report")
    expect(doc.serialize()).to_be("serialized")
    expect(doc.title).to_be("My Report")

def test_diamond_inheritance_mro():
    b = Bottom()
    # MRO: Bottom -> Left -> Right -> Base
    # Left.identify() should be called (first in MRO after Bottom)
    expect(b.identify()).to_be("Left")

def test_init_inheritance():
    # Truck has its own __init__ calling parent explicitly
    t = Truck("Chevy", "Silverado", 2000)
    expect(t.make).to_be("Chevy")
    expect(t.payload).to_be(2000)

def test_attribute_inheritance():
    d = Defaults()
    expect(d.color).to_be("red")
    expect(d.size).to_be(10)

    cd = CustomDefaults()
    expect(cd.color).to_be("blue")
    expect(cd.size).to_be(10)  # inherited from Defaults

    mc = MoreCustom()
    expect(mc.color).to_be("blue")  # inherited from CustomDefaults
    expect(mc.size).to_be(20)  # overridden

def test_class_vs_instance_attributes():
    c1 = Config("app1")
    c2 = Config("app2")
    expect(c1.debug).to_be(False)
    expect(c1.name).to_be("app1")
    expect(c2.name).to_be("app2")
    # Modifying instance doesn't affect class
    c1.debug = True
    expect(c1.debug).to_be(True)
    expect(c2.debug).to_be(False)
    expect(Config.debug).to_be(False)

def test_subclass_class_attributes():
    ac = AppConfig("myapp")
    expect(ac.debug).to_be(True)
    expect(ac.version).to_be("1.0")  # inherited
    expect(ac.max_retries).to_be(3)  # inherited

def test_type_name_checking():
    car = Car("Honda", "Civic", 4)
    truck = Truck("Ford", "F150", 1000)
    expect(type(car).__name__).to_be("Car")
    expect(type(truck).__name__).to_be("Truck")

def test_hierarchy_type_checking():
    dog = Dog()
    cat = Cat()
    snake = Snake()
    expect(dog.kind).to_be("dog")
    expect(cat.kind).to_be("cat")
    expect(snake.kind).to_be("snake")
    expect(type(dog).__name__).to_be("Dog")

def test_override_str():
    v = Vector(3, 4)
    expect(str(v)).to_be("Vector(3, 4)")

def test_override_repr():
    v = Vector(3, 4)
    expect(repr(v)).to_be("Vector(x=3, y=4)")

def test_override_eq():
    v1 = Vector(1, 2)
    v2 = Vector(1, 2)
    v3 = Vector(3, 4)
    expect(v1 == v2).to_be(True)
    expect(v1 == v3).to_be(False)

def test_override_lt():
    v1 = Vector(1, 1)   # magnitude^2 = 2
    v2 = Vector(3, 4)   # magnitude^2 = 25
    expect(v1 < v2).to_be(True)
    expect(v2 < v1).to_be(False)

def test_override_len():
    v = Vector(3, 4)
    expect(len(v)).to_be(7)

def test_override_getitem():
    v = Vector(10, 20)
    expect(v[0]).to_be(10)
    expect(v[1]).to_be(20)
    error_raised = False
    try:
        v[5]
    except IndexError:
        error_raised = True
    expect(error_raised).to_be(True)

def test_override_add():
    v1 = Vector(1, 2)
    v2 = Vector(3, 4)
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

def test_inherited_dunder_with_override():
    v = Vector3D(1, 2, 3)
    expect(str(v)).to_be("Vector3D(1, 2, 3)")
    expect(v[0]).to_be(1)
    expect(v[1]).to_be(2)
    expect(v[2]).to_be(3)
    # __eq__ is overridden in Vector3D
    v2 = Vector3D(1, 2, 3)
    expect(v == v2).to_be(True)

def test_abstract_pattern():
    s = Shape()
    error_raised = False
    try:
        s.area()
    except NotImplementedError:
        error_raised = True
    expect(error_raised).to_be(True)

def test_abstract_concrete_circle():
    c = Circle(5)
    area = c.area()
    # 3.14159 * 25 = 78.53975
    expect(area > 78.0).to_be(True)
    expect(area < 79.0).to_be(True)
    peri = c.perimeter()
    expect(peri > 31.0).to_be(True)
    expect(peri < 32.0).to_be(True)
    expect(c.describe()).to_be("Shape with area " + str(c.area()))

def test_abstract_concrete_rectangle():
    r = Rectangle(3, 4)
    expect(r.area()).to_be(12)
    expect(r.perimeter()).to_be(14)

def test_square_inherits_rectangle():
    sq = Square(5)
    expect(sq.area()).to_be(25)
    expect(sq.perimeter()).to_be(20)
    expect(sq.width).to_be(5)
    expect(sq.height).to_be(5)

def test_mixin_json_like():
    p = Person("Alice", 30)
    result = p.to_json_like()
    expect(result).to_be({"name": "Alice", "age": 30})

def test_mixin_comparable():
    p1 = Person("Alice", 30)
    p2 = Person("Bob", 25)
    p3 = Person("Charlie", 30)
    expect(p1.is_equal(p3)).to_be(True)
    expect(p1.is_equal(p2)).to_be(False)
    expect(p2.is_less(p1)).to_be(True)
    expect(p1.is_less(p2)).to_be(False)

# =============================================================================
# Run all tests
# =============================================================================

test("basic_single_inheritance", test_basic_single_inheritance)
test("method_override", test_method_override)
test("calling_parent_method", test_calling_parent_method)
test("deep_inheritance_chain", test_deep_inheritance_chain)
test("multiple_inheritance_basic", test_multiple_inheritance_basic)
test("diamond_inheritance_mro", test_diamond_inheritance_mro)
test("init_inheritance", test_init_inheritance)
test("attribute_inheritance", test_attribute_inheritance)
test("class_vs_instance_attributes", test_class_vs_instance_attributes)
test("subclass_class_attributes", test_subclass_class_attributes)
test("type_name_checking", test_type_name_checking)
test("hierarchy_type_checking", test_hierarchy_type_checking)
test("override_str", test_override_str)
test("override_repr", test_override_repr)
test("override_eq", test_override_eq)
test("override_lt", test_override_lt)
test("override_len", test_override_len)
test("override_getitem", test_override_getitem)
test("override_add", test_override_add)
test("inherited_dunder_with_override", test_inherited_dunder_with_override)
test("abstract_pattern", test_abstract_pattern)
test("abstract_concrete_circle", test_abstract_concrete_circle)
test("abstract_concrete_rectangle", test_abstract_concrete_rectangle)
test("square_inherits_rectangle", test_square_inherits_rectangle)
test("mixin_json_like", test_mixin_json_like)
test("mixin_comparable", test_mixin_comparable)

print("CPython inheritance tests completed")
