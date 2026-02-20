# Test: CPython Class Feature Edge Cases
# Adapted from CPython's test_class.py, test_descr.py, test_property.py

from test_framework import test, expect

# === Property ===
def test_property_basic():
    class C:
        def __init__(self):
            self._x = 0
        @property
        def x(self):
            return self._x
        @x.setter
        def x(self, val):
            self._x = val
    c = C()
    expect(c.x).to_be(0)
    c.x = 42
    expect(c.x).to_be(42)

def test_property_readonly():
    class C:
        @property
        def x(self):
            return 42
    c = C()
    expect(c.x).to_be(42)
    try:
        c.x = 10
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

# === Classmethod ===
def test_classmethod():
    class C:
        count = 0
        @classmethod
        def increment(cls):
            cls.count = cls.count + 1
            return cls.count
    expect(C.increment()).to_be(1)
    expect(C.increment()).to_be(2)
    c = C()
    expect(c.increment()).to_be(3)

# === Staticmethod ===
def test_staticmethod():
    class C:
        @staticmethod
        def add(x, y):
            return x + y
    expect(C.add(1, 2)).to_be(3)
    c = C()
    expect(c.add(3, 4)).to_be(7)

# === Super ===
def test_super_basic():
    class A:
        def method(self):
            return "A"
    class B(A):
        def method(self):
            return "B+" + super().method()
    b = B()
    expect(b.method()).to_be("B+A")

def test_super_init():
    class Base:
        def __init__(self):
            self.x = 10
    class Child(Base):
        def __init__(self):
            super().__init__()
            self.y = 20
    c = Child()
    expect(c.x).to_be(10)
    expect(c.y).to_be(20)

def test_super_chain():
    class A:
        def __init__(self):
            self.order = ["A"]
    class B(A):
        def __init__(self):
            super().__init__()
            self.order.append("B")
    class C(B):
        def __init__(self):
            super().__init__()
            self.order.append("C")
    c = C()
    expect(c.order).to_be(["A", "B", "C"])

# === MRO ===
def test_mro_diamond():
    class A:
        def method(self):
            return "A"
    class B(A):
        pass
    class C(A):
        def method(self):
            return "C"
    class D(B, C):
        pass
    d = D()
    # MRO: D -> B -> C -> A -> object
    expect(d.method()).to_be("C")

# === __str__ and __repr__ ===
def test_str_repr():
    class C:
        def __str__(self):
            return "str_C"
        def __repr__(self):
            return "repr_C"
    c = C()
    expect(str(c)).to_be("str_C")
    expect(repr(c)).to_be("repr_C")

# === __eq__ and __ne__ ===
def test_equality_dunder():
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def __eq__(self, other):
            if type(other) != type(self):
                return False
            return self.x == other.x and self.y == other.y
    p1 = Point(1, 2)
    p2 = Point(1, 2)
    p3 = Point(3, 4)
    expect(p1 == p2).to_be(True)
    expect(p1 == p3).to_be(False)
    expect(p1 != p3).to_be(True)

# === __len__ and __bool__ ===
def test_len_bool_dunder():
    class Container:
        def __init__(self, items):
            self.items = items
        def __len__(self):
            return len(self.items)
    c = Container([1, 2, 3])
    expect(len(c)).to_be(3)
    expect(bool(c)).to_be(True)
    empty = Container([])
    expect(len(empty)).to_be(0)
    expect(bool(empty)).to_be(False)

# === __getitem__ and __setitem__ ===
def test_item_dunder():
    class MyList:
        def __init__(self):
            self.data = {}
        def __getitem__(self, key):
            return self.data[key]
        def __setitem__(self, key, value):
            self.data[key] = value
    ml = MyList()
    ml[0] = "a"
    ml[1] = "b"
    expect(ml[0]).to_be("a")
    expect(ml[1]).to_be("b")

# === __contains__ ===
def test_contains_dunder():
    class MySet:
        def __init__(self, items):
            self.items = items
        def __contains__(self, item):
            return item in self.items
    s = MySet([1, 2, 3])
    expect(1 in s).to_be(True)
    expect(4 in s).to_be(False)

# === __add__ ===
def test_add_dunder():
    class Vector:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def __add__(self, other):
            return Vector(self.x + other.x, self.y + other.y)
        def __eq__(self, other):
            return self.x == other.x and self.y == other.y
    v1 = Vector(1, 2)
    v2 = Vector(3, 4)
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

# === Class variables vs instance variables ===
def test_class_vs_instance_vars():
    class C:
        shared = []
    a = C()
    b = C()
    a.shared.append(1)
    expect(b.shared).to_be([1])
    # Instance variable shadows class variable
    a.own = 42
    expect(a.own).to_be(42)
    expect(hasattr(b, "own")).to_be(False)

# === Inheritance ===
def test_method_override():
    class Animal:
        def speak(self):
            return "..."
    class Dog(Animal):
        def speak(self):
            return "Woof"
    class Cat(Animal):
        def speak(self):
            return "Meow"
    expect(Dog().speak()).to_be("Woof")
    expect(Cat().speak()).to_be("Meow")

# === isinstance with inheritance ===
def test_isinstance_inheritance():
    class A:
        pass
    class B(A):
        pass
    class C(B):
        pass
    c = C()
    expect(isinstance(c, C)).to_be(True)
    expect(isinstance(c, B)).to_be(True)
    expect(isinstance(c, A)).to_be(True)
    expect(isinstance(A(), C)).to_be(False)

# === __iter__ and __next__ ===
def test_iter_dunder():
    class Counter:
        def __init__(self, limit):
            self.limit = limit
            self.current = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.current >= self.limit:
                raise StopIteration
            val = self.current
            self.current = self.current + 1
            return val
    expect(list(Counter(5))).to_be([0, 1, 2, 3, 4])

# === __call__ ===
def test_call_dunder():
    class Multiplier:
        def __init__(self, factor):
            self.factor = factor
        def __call__(self, x):
            return x * self.factor
    double = Multiplier(2)
    triple = Multiplier(3)
    expect(double(5)).to_be(10)
    expect(triple(5)).to_be(15)

# === Class as factory ===
def test_class_factory():
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        @classmethod
        def from_tuple(cls, t):
            return cls(t[0], t[1])
        @classmethod
        def origin(cls):
            return cls(0, 0)
    p = Point.from_tuple((3, 4))
    expect(p.x).to_be(3)
    expect(p.y).to_be(4)
    o = Point.origin()
    expect(o.x).to_be(0)
    expect(o.y).to_be(0)

# Register all tests
test("property_basic", test_property_basic)
test("property_readonly", test_property_readonly)
test("classmethod", test_classmethod)
test("staticmethod", test_staticmethod)
test("super_basic", test_super_basic)
test("super_init", test_super_init)
test("super_chain", test_super_chain)
test("mro_diamond", test_mro_diamond)
test("str_repr", test_str_repr)
test("equality_dunder", test_equality_dunder)
test("len_bool_dunder", test_len_bool_dunder)
test("item_dunder", test_item_dunder)
test("contains_dunder", test_contains_dunder)
test("add_dunder", test_add_dunder)
test("class_vs_instance_vars", test_class_vs_instance_vars)
test("method_override", test_method_override)
test("isinstance_inheritance", test_isinstance_inheritance)
test("iter_dunder", test_iter_dunder)
test("call_dunder", test_call_dunder)
test("class_factory", test_class_factory)

print("CPython class tests completed")
