from abc import ABC, abstractmethod
from test_framework import test, expect

# Test 1: __subclasshook__ returning True
def test_subclasshook_true(t):
    class Drawable(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            if hasattr(C, "draw"):
                return True
            return NotImplemented

        @abstractmethod
        def draw(self):
            pass

    class Circle:
        def draw(self):
            return "drawing circle"

    expect(issubclass(Circle, Drawable)).to_be(True)
    expect(isinstance(Circle(), Drawable)).to_be(True)

test("__subclasshook__ returning True", test_subclasshook_true)

# Test 2: __subclasshook__ returning NotImplemented falls through
def test_subclasshook_notimplemented(t):
    class MyABC(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            return NotImplemented

    class Unrelated:
        pass

    expect(issubclass(Unrelated, MyABC)).to_be(False)

test("__subclasshook__ returning NotImplemented falls through", test_subclasshook_notimplemented)

# Test 3: __subclasshook__ returning False
def test_subclasshook_false(t):
    class Exclusive(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            return False

    class Child(Exclusive):
        pass

    # Even actual subclass fails because hook returns False
    expect(issubclass(Child, Exclusive)).to_be(False)

test("__subclasshook__ returning False", test_subclasshook_false)

# Test 4: isinstance with __subclasshook__
def test_isinstance_with_hook(t):
    class Iterable(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            if hasattr(C, "__iter__"):
                return True
            return NotImplemented

        @abstractmethod
        def __iter__(self):
            pass

    class MyCollection:
        def __iter__(self):
            return iter([])

    class NotIterable:
        pass

    expect(isinstance(MyCollection(), Iterable)).to_be(True)
    expect(isinstance(NotIterable(), Iterable)).to_be(False)

test("isinstance uses __subclasshook__", test_isinstance_with_hook)

# Test 5: __subclasshook__ without ABC metaclass has no effect
def test_hook_without_abcmeta(t):
    class RegularClass:
        @classmethod
        def __subclasshook__(cls, C):
            return True

    class Other:
        pass

    # Regular classes don't use __subclasshook__
    expect(issubclass(Other, RegularClass)).to_be(False)

test("__subclasshook__ without ABCMeta has no effect", test_hook_without_abcmeta)

# Test 6: NotImplemented builtin exists and has correct str
def test_notimplemented_builtin(t):
    expect(str(NotImplemented)).to_be("NotImplemented")

test("NotImplemented builtin", test_notimplemented_builtin)

# Test 7: ABC.register still works alongside __subclasshook__
def test_register_with_hook(t):
    class Serializable(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            return NotImplemented  # always defer

    class MyData:
        pass

    Serializable.register(MyData)
    expect(issubclass(MyData, Serializable)).to_be(True)
    expect(isinstance(MyData(), Serializable)).to_be(True)

test("ABC.register works alongside __subclasshook__", test_register_with_hook)

# Test 8: __subclasshook__ checking for multiple methods
def test_hook_multiple_methods(t):
    class Mapping(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            if hasattr(C, "__getitem__") and hasattr(C, "__len__"):
                return True
            return NotImplemented

    class MyDict:
        def __getitem__(self, key):
            pass
        def __len__(self):
            return 0

    class Incomplete:
        def __getitem__(self, key):
            pass

    expect(issubclass(MyDict, Mapping)).to_be(True)
    expect(issubclass(Incomplete, Mapping)).to_be(False)

test("__subclasshook__ checking for multiple methods", test_hook_multiple_methods)

# Test 9: Actual ABC subclass still works
def test_actual_subclass(t):
    class Base(ABC):
        @classmethod
        def __subclasshook__(cls, C):
            return NotImplemented

        @abstractmethod
        def method(self):
            pass

    class Impl(Base):
        def method(self):
            return 42

    expect(issubclass(Impl, Base)).to_be(True)
    expect(isinstance(Impl(), Base)).to_be(True)

test("actual ABC subclass still works", test_actual_subclass)
