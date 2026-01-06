# Test: Data Types
# Tests basic Python data types

from test_framework import test, expect

def test_none():
    expect(None, None)
    expect(True, None is None)

def test_bool():
    expect(True, True)
    expect(False, False)
    expect(True, bool(1))
    expect(False, bool(0))
    expect(True, bool("hello"))
    expect(False, bool(""))

def test_int():
    expect(42, 42)
    expect(-17, -17)
    expect(0, 0)
    expect(1000000, 1000000)
    expect(3, int(3.7))
    expect(123, int("123"))

def test_float():
    expect(3.14, 3.14)
    expect(-2.5, -2.5)
    expect(0.0, 0.0)
    expect(42.0, float(42))
    expect(3.14, float("3.14"))

def test_string():
    expect("hello", 'hello')
    expect("world", "world")
    expect("", "")
    expect("hello world", "hello world")
    expect("hello world", "hello" + " " + "world")
    expect("ababab", "ab" * 3)
    expect(5, len("hello"))
    expect("e", "hello"[1])
    expect("o", "hello"[-1])

def test_list():
    expect([], [])
    expect([1, 2, 3], [1, 2, 3])
    expect([1, "two", 3.0, True], [1, "two", 3.0, True])
    expect([[1, 2], [3, 4]], [[1, 2], [3, 4]])
    expect(5, len([1, 2, 3, 4, 5]))
    expect(20, [10, 20, 30][1])
    expect([1, 2, 3, 4], [1, 2] + [3, 4])
    expect([1, 2, 1, 2], [1, 2] * 2)
    expect(3, [1, 2, 3][-1])

def test_tuple():
    expect((), ())
    expect((1,), (1,))
    expect((1, 2, 3), (1, 2, 3))
    expect((1, "two", 3.0), (1, "two", 3.0))
    expect(3, len((1, 2, 3)))
    expect(20, (10, 20, 30)[1])

def test_dict():
    expect({}, {})
    expect({"a": 1, "b": 2}, {"a": 1, "b": 2})
    expect({"int": 1, "str": "hello", "bool": True}, {"int": 1, "str": "hello", "bool": True})
    expect(3, len({"a": 1, "b": 2, "c": 3}))
    expect(10, {"x": 10, "y": 20}["x"])
    expect(99, {"a": 1}.get("b", 99))

def test_range():
    expect([0, 1, 2, 3, 4], list(range(5)))
    expect([2, 3, 4, 5, 6], list(range(2, 7)))
    expect([0, 2, 4, 6, 8], list(range(0, 10, 2)))

def test_isinstance():
    expect(True, isinstance(42, int))
    expect(True, isinstance("hello", str))
    expect(True, isinstance([1, 2], list))

test("none", test_none)
test("bool", test_bool)
test("int", test_int)
test("float", test_float)
test("string", test_string)
test("list", test_list)
test("tuple", test_tuple)
test("dict", test_dict)
test("range", test_range)
test("isinstance", test_isinstance)

print("Data types tests completed")
