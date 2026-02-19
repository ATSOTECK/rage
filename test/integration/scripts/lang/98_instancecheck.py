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

# =============================================================================
# Tests ported from CPython's test_isinstance.py
# =============================================================================

# --- isinstance with tuple of types ---

def test_isinstance_tuple_of_types(t):
    """isinstance with a tuple as the second argument"""
    expect(isinstance(42, (str, int))).to_be(True)
    expect(isinstance("hi", (str, int))).to_be(True)
    expect(isinstance(3.14, (str, int))).to_be(False)
    expect(isinstance([], (str, int))).to_be(False)
    expect(isinstance(True, (str, int))).to_be(True)  # bool is subclass of int

test("isinstance with tuple of types", test_isinstance_tuple_of_types)

def test_isinstance_nested_tuple(t):
    """isinstance with nested tuples as the second argument"""
    expect(isinstance(42, (str, (float, int)))).to_be(True)
    expect(isinstance("hi", (str, (float, int)))).to_be(True)
    expect(isinstance(3.14, (str, (float, int)))).to_be(True)
    expect(isinstance([], (str, (float, int)))).to_be(False)

test("isinstance with nested tuple of types", test_isinstance_nested_tuple)

def test_isinstance_empty_tuple(t):
    """isinstance with empty tuple always returns False"""
    expect(isinstance(42, ())).to_be(False)
    expect(isinstance("hi", ())).to_be(False)
    expect(isinstance(None, ())).to_be(False)

test("isinstance with empty tuple", test_isinstance_empty_tuple)

# --- issubclass basic behavior ---

def test_issubclass_basic(t):
    """issubclass basic checks"""
    class Super:
        pass
    class Child(Super):
        pass

    expect(issubclass(Super, Super)).to_be(True)
    expect(issubclass(Child, Child)).to_be(True)
    expect(issubclass(Child, Super)).to_be(True)
    expect(issubclass(Super, Child)).to_be(False)

test("issubclass basic behavior", test_issubclass_basic)

def test_issubclass_tuple(t):
    """issubclass with tuple as second argument"""
    class Super:
        pass
    class Child(Super):
        pass

    expect(issubclass(Child, (Child,))).to_be(True)
    expect(issubclass(Child, (Super,))).to_be(True)
    expect(issubclass(Super, (Child,))).to_be(False)
    expect(issubclass(Super, (Child, Super))).to_be(True)
    expect(issubclass(Child, ())).to_be(False)

test("issubclass with tuple", test_issubclass_tuple)

def test_issubclass_nested_tuple(t):
    """issubclass with nested tuples"""
    expect(issubclass(int, (int, (float, int)))).to_be(True)
    expect(issubclass(str, (str, (int, str)))).to_be(True)

    class Super:
        pass
    class Child(Super):
        pass

    expect(issubclass(Super, (Child, (Super,)))).to_be(True)
    expect(issubclass(Child, (int, (str, Super)))).to_be(True)

test("issubclass with nested tuple", test_issubclass_nested_tuple)

def test_issubclass_builtin_hierarchy(t):
    """issubclass with built-in type hierarchy"""
    expect(issubclass(bool, int)).to_be(True)
    expect(issubclass(int, object)).to_be(True)
    expect(issubclass(str, object)).to_be(True)
    expect(issubclass(list, object)).to_be(True)
    expect(issubclass(int, str)).to_be(False)
    expect(issubclass(str, int)).to_be(False)

test("issubclass builtin hierarchy", test_issubclass_builtin_hierarchy)

# --- isinstance/issubclass with custom __instancecheck__/__subclasscheck__ ---

def test_custom_instancecheck_accepts_all(t):
    """Custom __instancecheck__ that accepts everything"""
    class AcceptAllMeta(type):
        def __instancecheck__(cls, instance):
            return True

    class Anything(metaclass=AcceptAllMeta):
        pass

    expect(isinstance(42, Anything)).to_be(True)
    expect(isinstance("hello", Anything)).to_be(True)
    expect(isinstance(None, Anything)).to_be(True)
    expect(isinstance([], Anything)).to_be(True)

test("custom __instancecheck__ accepts all", test_custom_instancecheck_accepts_all)

def test_custom_instancecheck_rejects_all(t):
    """Custom __instancecheck__ that rejects everything"""
    class RejectAllMeta(type):
        def __instancecheck__(cls, instance):
            return False

    class Nothing(metaclass=RejectAllMeta):
        pass

    n = Nothing()
    expect(isinstance(n, Nothing)).to_be(False)  # even its own instances
    expect(isinstance(42, Nothing)).to_be(False)

test("custom __instancecheck__ rejects all", test_custom_instancecheck_rejects_all)

def test_custom_subclasscheck_accepts_all(t):
    """Custom __subclasscheck__ that accepts everything"""
    class AcceptAllMeta(type):
        def __subclasscheck__(cls, subclass):
            return True

    class Universal(metaclass=AcceptAllMeta):
        pass

    expect(issubclass(int, Universal)).to_be(True)
    expect(issubclass(str, Universal)).to_be(True)

test("custom __subclasscheck__ accepts all", test_custom_subclasscheck_accepts_all)

def test_custom_subclasscheck_rejects_all(t):
    """Custom __subclasscheck__ that rejects everything"""
    class RejectAllMeta(type):
        def __subclasscheck__(cls, subclass):
            return False

    class Sealed(metaclass=RejectAllMeta):
        pass

    class Derived(Sealed):
        pass

    expect(issubclass(Derived, Sealed)).to_be(False)

test("custom __subclasscheck__ rejects all", test_custom_subclasscheck_rejects_all)

def test_instancecheck_with_attribute_check(t):
    """Custom __instancecheck__ that checks for a specific attribute"""
    class HasLenMeta(type):
        def __instancecheck__(cls, instance):
            return hasattr(instance, "__len__")

    class Sized(metaclass=HasLenMeta):
        pass

    expect(isinstance([], Sized)).to_be(True)
    expect(isinstance("hello", Sized)).to_be(True)
    expect(isinstance({}, Sized)).to_be(True)
    expect(isinstance(42, Sized)).to_be(False)

test("__instancecheck__ with attribute check", test_instancecheck_with_attribute_check)

def test_subclasscheck_with_method_check(t):
    """Custom __subclasscheck__ that checks for a method"""
    class HasIterMeta(type):
        def __subclasscheck__(cls, subclass):
            return hasattr(subclass, "__iter__")

    class Iterable(metaclass=HasIterMeta):
        pass

    expect(issubclass(list, Iterable)).to_be(True)
    expect(issubclass(str, Iterable)).to_be(True)
    expect(issubclass(dict, Iterable)).to_be(True)
    expect(issubclass(int, Iterable)).to_be(False)

test("__subclasscheck__ with method check", test_subclasscheck_with_method_check)

# --- TypeError for invalid arguments ---

def test_isinstance_non_class_raises_typeerror(t):
    """isinstance raises TypeError for non-class second arg"""
    raised = False
    try:
        isinstance(42, 42)
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("isinstance raises TypeError for non-class", test_isinstance_non_class_raises_typeerror)

def test_issubclass_non_class_first_arg_raises(t):
    """issubclass raises TypeError when first arg is not a class"""
    raised = False
    try:
        issubclass(42, int)
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("issubclass raises TypeError for non-class first arg", test_issubclass_non_class_first_arg_raises)

def test_issubclass_non_class_second_arg_raises(t):
    """issubclass raises TypeError when second arg is not a class"""
    raised = False
    try:
        issubclass(int, 42)
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("issubclass raises TypeError for non-class second arg", test_issubclass_non_class_second_arg_raises)

# --- isinstance normal with inheritance ---

def test_isinstance_normal_inheritance(t):
    """isinstance with normal class hierarchy (from CPython test)"""
    class Super:
        pass
    class Child(Super):
        pass

    expect(isinstance(Super(), Super)).to_be(True)
    expect(isinstance(Super(), Child)).to_be(False)
    expect(isinstance(Child(), Super)).to_be(True)
    expect(isinstance(Child(), Child)).to_be(True)

test("isinstance normal inheritance", test_isinstance_normal_inheritance)

def test_isinstance_with_multiple_inheritance(t):
    """isinstance with multiple inheritance"""
    class A:
        pass
    class B:
        pass
    class C(A, B):
        pass

    c = C()
    expect(isinstance(c, A)).to_be(True)
    expect(isinstance(c, B)).to_be(True)
    expect(isinstance(c, C)).to_be(True)
    expect(isinstance(c, object)).to_be(True)

test("isinstance with multiple inheritance", test_isinstance_with_multiple_inheritance)

def test_issubclass_with_multiple_inheritance(t):
    """issubclass with multiple inheritance"""
    class A:
        pass
    class B:
        pass
    class C(A, B):
        pass

    expect(issubclass(C, A)).to_be(True)
    expect(issubclass(C, B)).to_be(True)
    expect(issubclass(C, object)).to_be(True)
    expect(issubclass(A, C)).to_be(False)
    expect(issubclass(B, C)).to_be(False)

test("issubclass with multiple inheritance", test_issubclass_with_multiple_inheritance)

# --- Abstract base classes with isinstance ---

def test_isinstance_with_abc(t):
    """isinstance with abstract base classes"""
    from abc import ABC, abstractmethod

    class Drawable(ABC):
        @abstractmethod
        def draw(self):
            pass

    class Circle(Drawable):
        def draw(self):
            return "circle"

    c = Circle()
    expect(isinstance(c, Drawable)).to_be(True)
    expect(isinstance(c, Circle)).to_be(True)
    expect(issubclass(Circle, Drawable)).to_be(True)

test("isinstance with ABC", test_isinstance_with_abc)

def test_isinstance_abc_with_register(t):
    """isinstance with ABC.register"""
    from abc import ABC

    class MyABC(ABC):
        pass

    class NotDirectSubclass:
        pass

    MyABC.register(NotDirectSubclass)
    expect(issubclass(NotDirectSubclass, MyABC)).to_be(True)
    expect(isinstance(NotDirectSubclass(), MyABC)).to_be(True)

test("isinstance with ABC.register", test_isinstance_abc_with_register)

def test_isinstance_abc_subclasshook(t):
    """isinstance uses __subclasshook__ for ABCs"""
    from abc import ABC

    class Printable(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            if hasattr(C, "__str__"):
                return True
            return NotImplemented

    # All objects have __str__ in Python
    expect(isinstance(42, Printable)).to_be(True)
    expect(isinstance("hi", Printable)).to_be(True)
    expect(issubclass(int, Printable)).to_be(True)

test("isinstance with ABC __subclasshook__", test_isinstance_abc_subclasshook)

# --- isinstance/issubclass with builtin types ---

def test_isinstance_builtin_types(t):
    """isinstance with all common builtin types"""
    expect(isinstance(42, int)).to_be(True)
    expect(isinstance(3.14, float)).to_be(True)
    expect(isinstance("hello", str)).to_be(True)
    expect(isinstance(True, bool)).to_be(True)
    expect(isinstance(True, int)).to_be(True)  # bool is subclass of int
    expect(isinstance(None, type(None))).to_be(True)
    expect(isinstance([], list)).to_be(True)
    expect(isinstance({}, dict)).to_be(True)
    expect(isinstance(set(), set)).to_be(True)
    expect(isinstance((1,), tuple)).to_be(True)

test("isinstance with builtin types", test_isinstance_builtin_types)

def test_issubclass_builtin_types(t):
    """issubclass with all common builtin types"""
    expect(issubclass(bool, int)).to_be(True)
    expect(issubclass(int, float)).to_be(False)
    expect(issubclass(float, int)).to_be(False)
    expect(issubclass(int, object)).to_be(True)
    expect(issubclass(str, object)).to_be(True)
    expect(issubclass(list, object)).to_be(True)
    expect(issubclass(type, object)).to_be(True)

test("issubclass with builtin types", test_issubclass_builtin_types)

# --- isinstance with __instancecheck__ that raises ---

def test_instancecheck_exception_propagates(t):
    """Exceptions from __instancecheck__ should propagate"""
    class ErrorMeta(type):
        def __instancecheck__(cls, instance):
            raise ValueError("custom error from __instancecheck__")

    class Broken(metaclass=ErrorMeta):
        pass

    raised = False
    msg = ""
    try:
        isinstance(42, Broken)
    except ValueError as e:
        raised = True
        msg = str(e)
    expect(raised).to_be(True)
    expect(msg).to_be("custom error from __instancecheck__")

test("__instancecheck__ exception propagates", test_instancecheck_exception_propagates)

def test_subclasscheck_exception_propagates(t):
    """Exceptions from __subclasscheck__ should propagate"""
    class ErrorMeta(type):
        def __subclasscheck__(cls, subclass):
            raise ValueError("custom error from __subclasscheck__")

    class Broken(metaclass=ErrorMeta):
        pass

    raised = False
    msg = ""
    try:
        issubclass(int, Broken)
    except ValueError as e:
        raised = True
        msg = str(e)
    expect(raised).to_be(True)
    expect(msg).to_be("custom error from __subclasscheck__")

test("__subclasscheck__ exception propagates", test_subclasscheck_exception_propagates)
