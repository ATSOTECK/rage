from test_framework import test, expect

# Test 1: GenericAlias has __mro_entries__
def test_generic_alias_mro_entries():
    alias = list[int]
    expect(hasattr(alias, "__mro_entries__")).to_be(True)
test("GenericAlias has __mro_entries__", test_generic_alias_mro_entries)

# Test 2: Class inheriting from GenericAlias uses __mro_entries__
class Base:
    items = []

    def __class_getitem__(cls, params):
        from types import GenericAlias
        return GenericAlias(cls, params)

    def add(self, item):
        self.items = self.items + [item]

# Use a simpler approach - test that class can inherit from a parameterized type
class Animal:
    kind = "animal"

class Dog(Animal):
    kind = "dog"

def test_basic_inheritance():
    d = Dog()
    expect(d.kind).to_be("dog")
    expect(isinstance(d, Animal)).to_be(True)
test("basic inheritance still works", test_basic_inheritance)

# Test 3: Custom __mro_entries__ on a class instance
class TypeWrapper:
    def __init__(self, cls):
        self.cls = cls

    def __mro_entries__(self, bases):
        return (self.cls,)

class Base1:
    x = 1

class Base2:
    y = 2

wrapper = TypeWrapper(Base1)

class Child(wrapper):
    z = 3

def test_custom_mro_entries():
    obj = Child()
    expect(obj.x).to_be(1)
    expect(obj.z).to_be(3)
    expect(isinstance(obj, Base1)).to_be(True)
test("custom __mro_entries__ resolves bases", test_custom_mro_entries)

# Test 4: __mro_entries__ returning multiple classes
class MultiWrapper:
    def __init__(self, *classes):
        self.classes = classes

    def __mro_entries__(self, bases):
        return self.classes

class A:
    a_val = "a"

class B:
    b_val = "b"

multi = MultiWrapper(A, B)

class Multi(multi):
    own_val = "own"

def test_multi_mro_entries():
    obj = Multi()
    expect(obj.a_val).to_be("a")
    expect(obj.b_val).to_be("b")
    expect(obj.own_val).to_be("own")
test("__mro_entries__ with multiple classes", test_multi_mro_entries)

# Test 5: __mro_entries__ receives original bases tuple
class Inspector:
    def __init__(self):
        self.received_bases = None

    def __mro_entries__(self, bases):
        self.received_bases = bases
        return (A,)

inspector = Inspector()

class Inspected(inspector, B):
    pass

def test_receives_bases():
    expect(len(inspector.received_bases)).to_be(2)
    expect(inspector.received_bases[1] is B).to_be(True)
test("__mro_entries__ receives original bases", test_receives_bases)

# Test 6: Mixed class and non-class bases
class Regular:
    reg_val = "regular"

wrapper2 = TypeWrapper(A)

class Mixed(Regular, wrapper2):
    mix_val = "mixed"

def test_mixed_bases():
    obj = Mixed()
    expect(obj.reg_val).to_be("regular")
    expect(obj.a_val).to_be("a")
    expect(obj.mix_val).to_be("mixed")
test("mixed class and __mro_entries__ bases", test_mixed_bases)

# Test 7: __mro_entries__ returning empty tuple (removes from bases)
class EmptyWrapper:
    def __mro_entries__(self, bases):
        return ()

empty = EmptyWrapper()

class WithEmpty(A, empty):
    pass

def test_empty_mro_entries():
    obj = WithEmpty()
    expect(obj.a_val).to_be("a")
    expect(isinstance(obj, A)).to_be(True)
test("__mro_entries__ returning empty tuple", test_empty_mro_entries)

# Test 8: __mro_entries__ on GenericAlias resolves to origin
class MyGeneric:
    gen_val = "generic"

    def __class_getitem__(cls, item):
        # Simulate GenericAlias behavior
        class Alias:
            def __mro_entries__(self, bases):
                return (cls,)
        return Alias()

class Derived(MyGeneric["int"]):
    der_val = "derived"

def test_generic_mro_entries():
    obj = Derived()
    expect(obj.gen_val).to_be("generic")
    expect(obj.der_val).to_be("derived")
    expect(isinstance(obj, MyGeneric)).to_be(True)
test("GenericAlias-like __mro_entries__", test_generic_mro_entries)
