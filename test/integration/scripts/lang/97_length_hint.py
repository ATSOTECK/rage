from operator import length_hint
from test_framework import test, expect

# Test 1: __length_hint__ on custom iterator
def test_basic_length_hint(t):
    class MyIter:
        def __init__(self, n):
            self.remaining = n
        def __length_hint__(self):
            return self.remaining

    it = MyIter(10)
    expect(length_hint(it)).to_be(10)

test("basic __length_hint__", test_basic_length_hint)

# Test 2: __len__ takes priority over __length_hint__
def test_len_priority(t):
    class Both:
        def __len__(self):
            return 5
        def __length_hint__(self):
            return 100

    b = Both()
    expect(length_hint(b)).to_be(5)

test("__len__ takes priority over __length_hint__", test_len_priority)

# Test 3: Falls back to default when neither defined
def test_fallback_default(t):
    class Empty:
        pass

    e = Empty()
    expect(length_hint(e)).to_be(0)
    expect(length_hint(e, 42)).to_be(42)

test("falls back to default when no dunder", test_fallback_default)

# Test 4: __length_hint__ returning 0
def test_hint_zero(t):
    class Exhausted:
        def __length_hint__(self):
            return 0

    e = Exhausted()
    expect(length_hint(e)).to_be(0)

test("__length_hint__ returning 0", test_hint_zero)

# Test 5: length_hint on built-in list uses len
def test_builtin_list(t):
    expect(length_hint([1, 2, 3])).to_be(3)

test("length_hint on built-in list", test_builtin_list)

# Test 6: length_hint on built-in tuple
def test_builtin_tuple(t):
    expect(length_hint((1, 2))).to_be(2)

test("length_hint on built-in tuple", test_builtin_tuple)

# Test 7: length_hint on built-in dict
def test_builtin_dict(t):
    expect(length_hint({"a": 1, "b": 2})).to_be(2)

test("length_hint on built-in dict", test_builtin_dict)

# Test 8: length_hint on built-in string
def test_builtin_string(t):
    expect(length_hint("hello")).to_be(5)

test("length_hint on built-in string", test_builtin_string)

# Test 9: __length_hint__ inherited
def test_inherited(t):
    class Base:
        def __length_hint__(self):
            return 7

    class Child(Base):
        pass

    c = Child()
    expect(length_hint(c)).to_be(7)

test("__length_hint__ inherited from base", test_inherited)

# Test 10: Negative __length_hint__ raises ValueError
def test_negative_hint(t):
    class Bad:
        def __length_hint__(self):
            return -1

    b = Bad()
    try:
        length_hint(b)
        expect(True).to_be(False)
    except ValueError:
        expect(True).to_be(True)

test("negative __length_hint__ raises ValueError", test_negative_hint)

# Test 11: Only __len__ defined
def test_only_len(t):
    class HasLen:
        def __len__(self):
            return 3

    h = HasLen()
    expect(length_hint(h)).to_be(3)

test("only __len__ defined", test_only_len)

# Test 12: Default value as second arg
def test_default_arg(t):
    class NoHint:
        pass

    n = NoHint()
    expect(length_hint(n, 99)).to_be(99)

test("default value as second arg", test_default_arg)
