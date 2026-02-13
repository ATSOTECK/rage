# Test: ABC Module
# Tests abstract base classes, abstractmethod, register, and deprecated decorators

from test_framework import test, expect
from abc import ABC, ABCMeta, abstractmethod, abstractclassmethod, abstractstaticmethod, abstractproperty

# === Basic ABC with abstractmethod ===

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

    @abstractmethod
    def perimeter(self):
        pass

def test_cannot_instantiate_abstract():
    try:
        s = Shape()
        expect(True).to_be(False)  # Should not reach here
    except TypeError:
        expect(True).to_be(True)

test("cannot instantiate abstract class", test_cannot_instantiate_abstract)

# === Concrete subclass ===

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return 3.14159 * self.radius * self.radius

    def perimeter(self):
        return 2 * 3.14159 * self.radius

def test_concrete_subclass():
    c = Circle(5)
    expect(c.area()).to_be(3.14159 * 25)
    expect(c.perimeter()).to_be(2 * 3.14159 * 5)

test("concrete subclass works", test_concrete_subclass)

# === Partial implementation still abstract ===

class HalfShape(Shape):
    def area(self):
        return 0

def test_partial_implementation():
    try:
        h = HalfShape()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("partial implementation still abstract", test_partial_implementation)

# === isinstance and issubclass with ABC ===

def test_isinstance_with_abc():
    c = Circle(3)
    expect(isinstance(c, Shape)).to_be(True)
    expect(isinstance(c, Circle)).to_be(True)

test("isinstance with ABC hierarchy", test_isinstance_with_abc)

def test_issubclass_with_abc():
    expect(issubclass(Circle, Shape)).to_be(True)
    expect(issubclass(Circle, ABC)).to_be(True)

test("issubclass with ABC hierarchy", test_issubclass_with_abc)

# === Virtual subclass registration ===

class MyABC(ABC):
    @abstractmethod
    def do_something(self):
        pass

class Unrelated:
    def do_something(self):
        return "done"

MyABC.register(Unrelated)

def test_register_isinstance():
    u = Unrelated()
    expect(isinstance(u, MyABC)).to_be(True)

test("isinstance with registered subclass", test_register_isinstance)

def test_register_issubclass():
    expect(issubclass(Unrelated, MyABC)).to_be(True)

test("issubclass with registered subclass", test_register_issubclass)

# === register returns the class (can be used as decorator) ===

class AnotherABC(ABC):
    @abstractmethod
    def method(self):
        pass

class Registered:
    def method(self):
        return 42

result = AnotherABC.register(Registered)

def test_register_returns_class():
    expect(result is Registered).to_be(True)

test("register returns the registered class", test_register_returns_class)

# === Multiple abstract methods ===

class Vehicle(ABC):
    @abstractmethod
    def start(self):
        pass

    @abstractmethod
    def stop(self):
        pass

    @abstractmethod
    def fuel_type(self):
        pass

class Car(Vehicle):
    def start(self):
        return "Car started"

    def stop(self):
        return "Car stopped"

    def fuel_type(self):
        return "gasoline"

def test_multiple_abstract_methods():
    c = Car()
    expect(c.start()).to_be("Car started")
    expect(c.stop()).to_be("Car stopped")
    expect(c.fuel_type()).to_be("gasoline")

test("multiple abstract methods", test_multiple_abstract_methods)

# === Missing one of multiple abstract methods ===

class IncompleteCar(Vehicle):
    def start(self):
        return "Started"

    def stop(self):
        return "Stopped"
    # Missing fuel_type

def test_missing_one_abstract():
    try:
        ic = IncompleteCar()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("missing one of multiple abstract methods", test_missing_one_abstract)

# === abstractclassmethod (deprecated) ===

class WithClassMethod(ABC):
    @abstractclassmethod
    def create(cls):
        pass

class ConcreteWithCM(WithClassMethod):
    @classmethod
    def create(cls):
        return "created"

def test_abstractclassmethod():
    # Can't instantiate abstract
    try:
        WithClassMethod()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)
    # Concrete works
    expect(ConcreteWithCM.create()).to_be("created")

test("abstractclassmethod decorator", test_abstractclassmethod)

# === abstractstaticmethod (deprecated) ===

class WithStaticMethod(ABC):
    @abstractstaticmethod
    def utility():
        pass

class ConcreteWithSM(WithStaticMethod):
    @staticmethod
    def utility():
        return "useful"

def test_abstractstaticmethod():
    try:
        WithStaticMethod()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)
    expect(ConcreteWithSM.utility()).to_be("useful")

test("abstractstaticmethod decorator", test_abstractstaticmethod)

# === abstractproperty (deprecated) ===

class WithProperty(ABC):
    @abstractproperty
    def value(self):
        pass

class ConcreteWithProp(WithProperty):
    @property
    def value(self):
        return 42

def test_abstractproperty():
    try:
        WithProperty()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)
    c = ConcreteWithProp()
    expect(c.value).to_be(42)

test("abstractproperty decorator", test_abstractproperty)

# === Diamond inheritance with ABCs ===

class Drawable(ABC):
    @abstractmethod
    def draw(self):
        pass

class Resizable(ABC):
    @abstractmethod
    def resize(self, factor):
        pass

class Widget(Drawable, Resizable):
    def draw(self):
        return "drawing widget"

    def resize(self, factor):
        return "resizing by " + str(factor)

def test_diamond_abc():
    w = Widget()
    expect(w.draw()).to_be("drawing widget")
    expect(w.resize(2)).to_be("resizing by 2")
    expect(isinstance(w, Drawable)).to_be(True)
    expect(isinstance(w, Resizable)).to_be(True)

test("diamond inheritance with ABCs", test_diamond_abc)

# === metaclass=ABCMeta usage ===

class MyInterface(metaclass=ABCMeta):
    @abstractmethod
    def execute(self):
        pass

class Implementation(MyInterface):
    def execute(self):
        return "executed"

def test_metaclass_abcmeta():
    try:
        MyInterface()
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)
    impl = Implementation()
    expect(impl.execute()).to_be("executed")

test("metaclass=ABCMeta usage", test_metaclass_abcmeta)
