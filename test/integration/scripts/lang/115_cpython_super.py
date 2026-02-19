# Adapted from CPython's Lib/test/test_super.py
# Tests for super() functionality
from test_framework import test, expect

# === MRO method resolution ===

def test_mro_chain():
    class A:
        def who(self):
            return "A"
    class B(A):
        def who(self):
            return "B->" + super().who()
    class C(A):
        def who(self):
            return "C->" + super().who()
    class D(B, C):
        def who(self):
            return "D->" + super().who()
    expect(D().who()).to_be("D->B->C->A")

test("super MRO method resolution chain", test_mro_chain)

# === Two-arg super skip class ===

def test_two_arg_skip():
    class X:
        def val(self):
            return 1
    class Y(X):
        def val(self):
            return 2
    class Z(Y):
        def val(self):
            return super(Y, self).val()
    expect(Z().val()).to_be(1)

test("two-arg super skips intermediate class", test_two_arg_skip)

# === super().__init__ with kwargs ===

def test_super_init_kwargs():
    class Base:
        def __init__(self, name="default"):
            self.name = name
    class Child(Base):
        def __init__(self, name, age):
            super().__init__(name=name)
            self.age = age
    c = Child("Alice", 30)
    expect(c.name).to_be("Alice")
    expect(c.age).to_be(30)

test("super().__init__ with keyword args", test_super_init_kwargs)

# === super() with @property ===

def test_super_property():
    class PropBase:
        @property
        def value(self):
            return 10
    class PropChild(PropBase):
        @property
        def value(self):
            return super().value + 5
    expect(PropChild().value).to_be(15)

test("super() with @property", test_super_property)

# === super() in __repr__ ===

def test_super_repr():
    class ReprBase:
        def __repr__(self):
            return "ReprBase"
    class ReprChild(ReprBase):
        def __repr__(self):
            return "ReprChild(" + super().__repr__() + ")"
    expect(repr(ReprChild())).to_be("ReprChild(ReprBase)")

test("super() in __repr__", test_super_repr)

# === super() in @classmethod ===

def test_super_classmethod():
    class CMBase:
        @classmethod
        def create(cls):
            return cls.__name__
    class CMChild(CMBase):
        @classmethod
        def create(cls):
            parent = super().create()
            return cls.__name__ + "<-" + parent
    expect(CMChild.create()).to_be("CMChild<-CMChild")

test("super() in @classmethod", test_super_classmethod)

# === Two-arg super in @classmethod ===

def test_two_arg_classmethod():
    class TBase:
        @classmethod
        def info(cls):
            return "TBase"
    class TChild(TBase):
        @classmethod
        def info(cls):
            return super(TChild, cls).info() + "+TChild"
    expect(TChild.info()).to_be("TBase+TChild")

test("two-arg super in @classmethod", test_two_arg_classmethod)

# === super() grandparent attribute ===

def test_grandparent():
    class GP:
        def greet(self):
            return "hello from GP"
    class P(GP):
        pass
    class GC(P):
        def greet(self):
            return super().greet() + " via GC"
    expect(GC().greet()).to_be("hello from GP via GC")

test("super() accesses grandparent method", test_grandparent)

# === object.__setattr__ via super ===

def test_super_setattr():
    class Tracked:
        def __init__(self):
            super().__setattr__("_log", [])
            self.x = 1
        def __setattr__(self, name, value):
            if hasattr(self, '_log'):
                self._log.append(name)
            super().__setattr__(name, value)
    t = Tracked()
    t.y = 2
    expect(t._log).to_be(["x", "y"])
    expect(t.x).to_be(1)
    expect(t.y).to_be(2)

test("object.__setattr__ via super()", test_super_setattr)

# === super() with __getattr__ delegation ===

def test_super_getattr():
    class DefaultBase:
        def __getattr__(self, name):
            return "default:" + name
    class DefaultChild(DefaultBase):
        def __getattr__(self, name):
            if name.startswith("x_"):
                return "child:" + name
            return super().__getattr__(name)
    d = DefaultChild()
    expect(d.x_test).to_be("child:x_test")
    expect(d.other).to_be("default:other")

test("super() delegates __getattr__", test_super_getattr)

# === Diamond classmethod super chain ===

def test_diamond_classmethod():
    class A:
        @classmethod
        def chain(cls):
            return ["A"]
    class B(A):
        @classmethod
        def chain(cls):
            return ["B"] + super().chain()
    class C(A):
        @classmethod
        def chain(cls):
            return ["C"] + super().chain()
    class D(B, C):
        @classmethod
        def chain(cls):
            return ["D"] + super().chain()
    expect(D.chain()).to_be(["D", "B", "C", "A"])

test("diamond classmethod super chain follows MRO", test_diamond_classmethod)

# === super() basic two-arg form ===

def test_basic_two_arg():
    class A:
        def f(self):
            return "A.f"
    class B(A):
        def f(self):
            return super(B, self).f()
    expect(B().f()).to_be("A.f")

test("basic two-arg super(Type, self)", test_basic_two_arg)

# === super() with multiple methods ===

def test_multiple_methods():
    class Base:
        def method1(self):
            return "base1"
        def method2(self):
            return "base2"
    class Child(Base):
        def method1(self):
            return "child1+" + super().method1()
        def method2(self):
            return "child2+" + super().method2()
    c = Child()
    expect(c.method1()).to_be("child1+base1")
    expect(c.method2()).to_be("child2+base2")

test("super() with multiple overridden methods", test_multiple_methods)

# === super() in __init__ chain ===

def test_init_chain():
    result = []
    class A:
        def __init__(self):
            result.append("A")
    class B(A):
        def __init__(self):
            result.append("B")
            super().__init__()
    class C(A):
        def __init__(self):
            result.append("C")
            super().__init__()
    class D(B, C):
        def __init__(self):
            result.append("D")
            super().__init__()
    D()
    expect(result).to_be(["D", "B", "C", "A"])

test("super().__init__ chain follows MRO in diamond", test_init_chain)

# === super() with *args forwarding ===

def test_args_forwarding():
    class A:
        def __init__(self, x, y):
            self.x = x
            self.y = y
    class B(A):
        def __init__(self, x, y, z):
            super().__init__(x, y)
            self.z = z
    b = B(1, 2, 3)
    expect(b.x).to_be(1)
    expect(b.y).to_be(2)
    expect(b.z).to_be(3)

test("super().__init__ with positional args forwarding", test_args_forwarding)

# === super() type inquiry (single arg) ===

def test_type_check():
    expect(type(1).__name__).to_be("int")
    expect(type("hello").__name__).to_be("str")
    expect(type([]).__name__).to_be("list")
    expect(type({}).__name__).to_be("dict")

test("type() single-arg type inquiry", test_type_check)

# === super() with cooperative multiple inheritance ===

def test_cooperative():
    class Base:
        def action(self):
            return ["Base"]
    class Mixin1(Base):
        def action(self):
            return ["Mixin1"] + super().action()
    class Mixin2(Base):
        def action(self):
            return ["Mixin2"] + super().action()
    class Combined(Mixin1, Mixin2):
        def action(self):
            return ["Combined"] + super().action()
    expect(Combined().action()).to_be(["Combined", "Mixin1", "Mixin2", "Base"])

test("cooperative multiple inheritance via super()", test_cooperative)

# === super() preserves self through MRO ===

def test_self_preserved():
    class A:
        def who_am_i(self):
            return type(self).__name__
    class B(A):
        def who_am_i(self):
            return super().who_am_i()
    class C(B):
        def who_am_i(self):
            return super().who_am_i()
    expect(C().who_am_i()).to_be("C")

test("super() preserves self through MRO chain", test_self_preserved)

# === super() with __eq__ ===

def test_super_eq():
    class Base:
        def __eq__(self, other):
            return True
    class Child(Base):
        def __eq__(self, other):
            if other == 42:
                return super().__eq__(other)
            return False
    c = Child()
    expect(c == 42).to_be(True)
    expect(c == 99).to_be(False)

test("super().__eq__ delegation", test_super_eq)

# === super() with __str__ ===

def test_super_str():
    class Base:
        def __str__(self):
            return "base_str"
    class Child(Base):
        def __str__(self):
            return "child+" + super().__str__()
    expect(str(Child())).to_be("child+base_str")

test("super().__str__ delegation", test_super_str)

# === two-arg super with type as second arg ===

def test_two_arg_type():
    class A:
        @classmethod
        def name(cls):
            return "A"
    class B(A):
        @classmethod
        def name(cls):
            return super(B, B).name()
    expect(B.name()).to_be("A")

test("two-arg super(Type, Type) for classmethods", test_two_arg_type)

# === super() deep inheritance chain ===

def test_deep_chain():
    class L0:
        def depth(self):
            return 0
    class L1(L0):
        def depth(self):
            return super().depth() + 1
    class L2(L1):
        def depth(self):
            return super().depth() + 1
    class L3(L2):
        def depth(self):
            return super().depth() + 1
    class L4(L3):
        def depth(self):
            return super().depth() + 1
    expect(L4().depth()).to_be(4)

test("super() through deep inheritance chain", test_deep_chain)

# === super() attribute from base not overridden ===

def test_inherited_attr():
    class A:
        def only_in_a(self):
            return "from A"
    class B(A):
        pass
    class C(B):
        def test(self):
            return super().only_in_a()
    expect(C().test()).to_be("from A")

test("super() finds non-overridden base method", test_inherited_attr)
