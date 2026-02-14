# Test: CPython Type Conversion Edge Cases
# Adapted from CPython's test_builtin.py - covers type constructor edge cases

from test_framework import test, expect

# === str() with various types ===
def test_str_from_int():
    expect(str(0)).to_be("0")
    expect(str(42)).to_be("42")
    expect(str(-7)).to_be("-7")
    expect(str(1000000)).to_be("1000000")

def test_str_from_float():
    expect(str(0.0)).to_be("0.0")
    expect(str(3.14)).to_be("3.14")
    expect(str(-2.5)).to_be("-2.5")
    expect(str(1.0)).to_be("1.0")

def test_str_from_bool():
    expect(str(True)).to_be("True")
    expect(str(False)).to_be("False")

def test_str_from_none():
    expect(str(None)).to_be("None")

def test_str_from_list():
    expect(str([])).to_be("[]")
    expect(str([1, 2, 3])).to_be("[1, 2, 3]")

def test_str_from_dict():
    expect(str({})).to_be("{}")

def test_str_from_tuple():
    expect(str(())).to_be("()")
    expect(str((1,))).to_be("(1,)")
    expect(str((1, 2, 3))).to_be("(1, 2, 3)")

# === Custom __str__ and __repr__ ===
def test_custom_str():
    class MyObj:
        def __str__(self):
            return "custom_str"
    expect(str(MyObj())).to_be("custom_str")

def test_custom_repr():
    class MyObj:
        def __repr__(self):
            return "custom_repr"
    expect(repr(MyObj())).to_be("custom_repr")

def test_str_falls_back_to_repr():
    class MyObj:
        def __repr__(self):
            return "from_repr"
    # str() should use __repr__ if no __str__
    expect(str(MyObj())).to_be("from_repr")

# === list() from various types ===
def test_list_from_string():
    expect(list("abc")).to_be(["a", "b", "c"])
    expect(list("")).to_be([])
    expect(list("x")).to_be(["x"])

def test_list_from_tuple():
    expect(list((1, 2, 3))).to_be([1, 2, 3])
    expect(list(())).to_be([])

def test_list_from_range():
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(list(range(0))).to_be([])
    expect(list(range(2, 5))).to_be([2, 3, 4])

def test_list_from_generator():
    def gen():
        yield 10
        yield 20
        yield 30
    expect(list(gen())).to_be([10, 20, 30])

def test_list_from_dict():
    # list(dict) gives keys
    d = {"a": 1, "b": 2}
    result = list(d)
    expect(len(result)).to_be(2)
    expect("a" in result).to_be(True)
    expect("b" in result).to_be(True)

# === tuple() from various types ===
def test_tuple_from_list():
    expect(tuple([1, 2, 3])).to_be((1, 2, 3))
    expect(tuple([])).to_be(())

def test_tuple_from_string():
    expect(tuple("abc")).to_be(("a", "b", "c"))
    expect(tuple("")).to_be(())

def test_tuple_from_range():
    expect(tuple(range(4))).to_be((0, 1, 2, 3))

# === set() from various types ===
def test_set_from_list():
    expect(set([1, 2, 2, 3, 3, 3])).to_be({1, 2, 3})

def test_set_from_tuple():
    expect(set((1, 2, 2, 3))).to_be({1, 2, 3})

def test_set_from_string():
    s = set("hello")
    expect(len(s)).to_be(4)  # h, e, l, o
    expect("h" in s).to_be(True)
    expect("l" in s).to_be(True)

# === bool() truthiness ===
def test_bool_empty_collections():
    expect(bool([])).to_be(False)
    expect(bool({})).to_be(False)
    expect(bool(())).to_be(False)
    expect(bool(set())).to_be(False)
    expect(bool("")).to_be(False)

def test_bool_nonempty_collections():
    expect(bool([1])).to_be(True)
    expect(bool({"a": 1})).to_be(True)
    expect(bool((1,))).to_be(True)
    expect(bool({1})).to_be(True)
    expect(bool("x")).to_be(True)

def test_bool_numbers():
    expect(bool(0)).to_be(False)
    expect(bool(0.0)).to_be(False)
    expect(bool(1)).to_be(True)
    expect(bool(-1)).to_be(True)
    expect(bool(0.1)).to_be(True)

def test_bool_none():
    expect(bool(None)).to_be(False)

def test_bool_custom_bool():
    class Truthy:
        def __bool__(self):
            return True
    class Falsy:
        def __bool__(self):
            return False
    expect(bool(Truthy())).to_be(True)
    expect(bool(Falsy())).to_be(False)

def test_bool_custom_len():
    class Empty:
        def __len__(self):
            return 0
    class NonEmpty:
        def __len__(self):
            return 3
    expect(bool(Empty())).to_be(False)
    expect(bool(NonEmpty())).to_be(True)

# === int() conversions ===
def test_int_from_string():
    expect(int("42")).to_be(42)
    expect(int("-7")).to_be(-7)
    expect(int("0")).to_be(0)

def test_int_from_float():
    expect(int(3.7)).to_be(3)
    expect(int(-2.9)).to_be(-2)
    expect(int(0.0)).to_be(0)

def test_int_from_bool():
    expect(int(True)).to_be(1)
    expect(int(False)).to_be(0)

# === float() conversions ===
def test_float_from_string():
    expect(float("3.14")).to_be(3.14)
    expect(float("-2.5")).to_be(-2.5)
    expect(float("0")).to_be(0.0)
    expect(float("1")).to_be(1.0)

def test_float_from_int():
    expect(float(42)).to_be(42.0)
    expect(float(0)).to_be(0.0)
    expect(float(-7)).to_be(-7.0)

def test_float_from_bool():
    expect(float(True)).to_be(1.0)
    expect(float(False)).to_be(0.0)

# Register all tests
test("str_from_int", test_str_from_int)
test("str_from_float", test_str_from_float)
test("str_from_bool", test_str_from_bool)
test("str_from_none", test_str_from_none)
test("str_from_list", test_str_from_list)
test("str_from_dict", test_str_from_dict)
test("str_from_tuple", test_str_from_tuple)
test("custom_str", test_custom_str)
test("custom_repr", test_custom_repr)
test("str_falls_back_to_repr", test_str_falls_back_to_repr)
test("list_from_string", test_list_from_string)
test("list_from_tuple", test_list_from_tuple)
test("list_from_range", test_list_from_range)
test("list_from_generator", test_list_from_generator)
test("list_from_dict", test_list_from_dict)
test("tuple_from_list", test_tuple_from_list)
test("tuple_from_string", test_tuple_from_string)
test("tuple_from_range", test_tuple_from_range)
test("set_from_list", test_set_from_list)
test("set_from_tuple", test_set_from_tuple)
test("set_from_string", test_set_from_string)
test("bool_empty_collections", test_bool_empty_collections)
test("bool_nonempty_collections", test_bool_nonempty_collections)
test("bool_numbers", test_bool_numbers)
test("bool_none", test_bool_none)
test("bool_custom_bool", test_bool_custom_bool)
test("bool_custom_len", test_bool_custom_len)
test("int_from_string", test_int_from_string)
test("int_from_float", test_int_from_float)
test("int_from_bool", test_int_from_bool)
test("float_from_string", test_float_from_string)
test("float_from_int", test_float_from_int)
test("float_from_bool", test_float_from_bool)

print("CPython type conversion tests completed")
