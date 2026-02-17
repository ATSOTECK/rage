from test_framework import test, expect

# Test 1: __instancecheck__ on metaclass
def test_basic_instancecheck(t):
    class MyMeta(type):
        def __instancecheck__(cls, instance):
            # Accept anything with a 'quack' attribute
            return hasattr(instance, "quack")

    class Duck(metaclass=MyMeta):
        def __init__(self):
            self.quack = True

    class NotADuck:
        pass

    class QuackingThing:
        def __init__(self):
            self.quack = True

    d = Duck()
    n = NotADuck()
    q = QuackingThing()
    expect(isinstance(d, Duck)).to_be(True)   # has quack
    expect(isinstance(n, Duck)).to_be(False)   # no quack
    expect(isinstance(q, Duck)).to_be(True)    # has quack via __instancecheck__

test("__instancecheck__ on metaclass", test_basic_instancecheck)

# Test 2: __subclasscheck__ on metaclass
def test_basic_subclasscheck(t):
    class ProtocolMeta(type):
        def __subclasscheck__(cls, subclass):
            # Accept any class that has a 'serialize' method
            return "serialize" in subclass.__dict__

    class Serializable(metaclass=ProtocolMeta):
        def serialize(self):
            pass

    class HasSerialize:
        def serialize(self):
            pass

    class NoSerialize:
        pass

    expect(issubclass(HasSerialize, Serializable)).to_be(True)
    expect(issubclass(NoSerialize, Serializable)).to_be(False)

test("__subclasscheck__ on metaclass", test_basic_subclasscheck)

# Test 3: isinstance without metaclass still works
def test_isinstance_normal(t):
    class Animal:
        pass

    class Dog(Animal):
        pass

    d = Dog()
    expect(isinstance(d, Dog)).to_be(True)
    expect(isinstance(d, Animal)).to_be(True)
    expect(isinstance(d, int)).to_be(False)

test("isinstance without metaclass works normally", test_isinstance_normal)

# Test 4: issubclass without metaclass still works
def test_issubclass_normal(t):
    class Base:
        pass

    class Child(Base):
        pass

    expect(issubclass(Child, Base)).to_be(True)
    expect(issubclass(Base, Child)).to_be(False)
    expect(issubclass(Child, Child)).to_be(True)

test("issubclass without metaclass works normally", test_issubclass_normal)

# Test 5: __instancecheck__ returning False overrides normal check
def test_instancecheck_override(t):
    class StrictMeta(type):
        def __instancecheck__(cls, instance):
            # Only exact class matches, no subclass instances
            return type(instance) == cls

    class Base(metaclass=StrictMeta):
        pass

    class Child(Base):
        pass

    b = Base()
    c = Child()
    expect(isinstance(b, Base)).to_be(True)
    expect(isinstance(c, Base)).to_be(False)  # overridden: Child is not exact Base

test("__instancecheck__ can override normal behavior", test_instancecheck_override)

# Test 6: __instancecheck__ with built-in types
def test_instancecheck_builtins(t):
    class AcceptAllMeta(type):
        def __instancecheck__(cls, instance):
            return True

    class Anything(metaclass=AcceptAllMeta):
        pass

    expect(isinstance(42, Anything)).to_be(True)
    expect(isinstance("hello", Anything)).to_be(True)
    expect(isinstance([1, 2], Anything)).to_be(True)

test("__instancecheck__ works with built-in type instances", test_instancecheck_builtins)

# Test 7: __subclasscheck__ with actual subclass
def test_subclasscheck_override(t):
    class NoSubclassMeta(type):
        def __subclasscheck__(cls, subclass):
            return False

    class Sealed(metaclass=NoSubclassMeta):
        pass

    class Derived(Sealed):
        pass

    expect(issubclass(Derived, Sealed)).to_be(False)  # metaclass says no

test("__subclasscheck__ can override normal behavior", test_subclasscheck_override)

# Test 8: isinstance with tuple still works
def test_isinstance_tuple(t):
    expect(isinstance(42, (str, int))).to_be(True)
    expect(isinstance("hi", (str, int))).to_be(True)
    expect(isinstance([], (str, int))).to_be(False)

test("isinstance with tuple of types", test_isinstance_tuple)

# Test 9: issubclass with built-in types still works
def test_issubclass_builtins(t):
    expect(issubclass(bool, int)).to_be(True)
    expect(issubclass(int, object)).to_be(True)
    expect(issubclass(str, int)).to_be(False)

test("issubclass with built-in types", test_issubclass_builtins)

# Test 10: Inherited __instancecheck__ from meta base
def test_inherited_meta(t):
    class BaseMeta(type):
        def __instancecheck__(cls, instance):
            return hasattr(instance, "valid") and instance.valid

    class DerivedMeta(BaseMeta):
        pass

    class Validated(metaclass=DerivedMeta):
        pass

    class Good:
        def __init__(self):
            self.valid = True

    class Bad:
        def __init__(self):
            self.valid = False

    expect(isinstance(Good(), Validated)).to_be(True)
    expect(isinstance(Bad(), Validated)).to_be(False)

test("inherited __instancecheck__ from meta base", test_inherited_meta)
