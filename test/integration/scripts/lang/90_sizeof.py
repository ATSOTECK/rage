from test_framework import test, expect
import sys

# Test 1: Default __sizeof__ on plain instance
def test_default_sizeof():
    class Empty:
        pass

    obj = Empty()
    size = obj.__sizeof__()
    expect(isinstance(size, int)).to_be(True)
    expect(size > 0).to_be(True)

test("default __sizeof__ on plain instance", test_default_sizeof)

# Test 2: Default __sizeof__ grows with attributes
def test_sizeof_with_attrs():
    class Empty:
        pass

    a = Empty()
    size_empty = a.__sizeof__()

    b = Empty()
    b.x = 1
    b.y = 2
    b.z = 3
    size_with_attrs = b.__sizeof__()

    expect(size_with_attrs > size_empty).to_be(True)

test("__sizeof__ grows with attributes", test_sizeof_with_attrs)

# Test 3: Custom __sizeof__
def test_custom_sizeof():
    class Fixed:
        def __sizeof__(self):
            return 42

    expect(Fixed().__sizeof__()).to_be(42)

test("custom __sizeof__", test_custom_sizeof)

# Test 4: sys.getsizeof uses __sizeof__
def test_getsizeof_uses_sizeof():
    class Custom:
        def __sizeof__(self):
            return 100

    size = sys.getsizeof(Custom())
    expect(size).to_be(100)

test("sys.getsizeof uses __sizeof__", test_getsizeof_uses_sizeof)

# Test 5: sys.getsizeof on built-in types
def test_getsizeof_builtins():
    expect(sys.getsizeof(0) > 0).to_be(True)
    expect(sys.getsizeof("hello") > 0).to_be(True)
    expect(sys.getsizeof([]) > 0).to_be(True)
    expect(sys.getsizeof({}) > 0).to_be(True)

test("sys.getsizeof on built-in types", test_getsizeof_builtins)

# Test 6: __sizeof__ inherited from parent
def test_inherited_sizeof():
    class Base:
        def __sizeof__(self):
            return 200

    class Child(Base):
        pass

    expect(Child().__sizeof__()).to_be(200)

test("__sizeof__ inherited from parent", test_inherited_sizeof)

# Test 7: Child overrides __sizeof__
def test_override_sizeof():
    class Base:
        def __sizeof__(self):
            return 200

    class Child(Base):
        def __sizeof__(self):
            return 300

    expect(Child().__sizeof__()).to_be(300)

test("child overrides __sizeof__", test_override_sizeof)

# Test 8: __sizeof__ based on data
def test_sizeof_data_based():
    class Buffer:
        def __init__(self, n):
            self.data = [0] * n

        def __sizeof__(self):
            return 64 + len(self.data) * 8

    b = Buffer(10)
    expect(b.__sizeof__()).to_be(64 + 80)

test("__sizeof__ based on internal data", test_sizeof_data_based)

# Test 9: sys.getsizeof string size scales with length
def test_getsizeof_string_scales():
    s1 = sys.getsizeof("a")
    s2 = sys.getsizeof("a" * 100)
    expect(s2 > s1).to_be(True)

test("sys.getsizeof string size scales", test_getsizeof_string_scales)

# Test 10: sys.getsizeof list size scales with length
def test_getsizeof_list_scales():
    s1 = sys.getsizeof([])
    s2 = sys.getsizeof([1, 2, 3, 4, 5])
    expect(s2 > s1).to_be(True)

test("sys.getsizeof list size scales", test_getsizeof_list_scales)
