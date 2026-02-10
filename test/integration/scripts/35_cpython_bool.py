# Test: CPython Bool Edge Cases
# Adapted from CPython's test_bool.py

from test_framework import test, expect

def test_bool_repr():
    expect(repr(True)).to_be("True")
    expect(repr(False)).to_be("False")

def test_bool_str():
    expect(str(True)).to_be("True")
    expect(str(False)).to_be("False")

def test_bool_int_conversion():
    expect(int(True)).to_be(1)
    expect(int(False)).to_be(0)
    expect(True + 1).to_be(2)
    expect(False + 1).to_be(1)
    expect(True * 3).to_be(3)
    expect(False * 3).to_be(0)

def test_bool_float_conversion():
    expect(float(True)).to_be(1.0)
    expect(float(False)).to_be(0.0)

def test_bool_constructor():
    expect(bool(0)).to_be(False)
    expect(bool(1)).to_be(True)
    expect(bool(-1)).to_be(True)
    expect(bool(0.0)).to_be(False)
    expect(bool(1.0)).to_be(True)
    expect(bool("")).to_be(False)
    expect(bool("x")).to_be(True)
    expect(bool([])).to_be(False)
    expect(bool([1])).to_be(True)
    expect(bool(())).to_be(False)
    expect(bool((1,))).to_be(True)
    expect(bool({})).to_be(False)
    expect(bool({1: 2})).to_be(True)
    expect(bool(set())).to_be(False)
    expect(bool({1})).to_be(True)
    expect(bool(None)).to_be(False)

def test_bool_isinstance():
    expect(isinstance(True, bool)).to_be(True)
    expect(isinstance(False, bool)).to_be(True)
    expect(isinstance(True, int)).to_be(True)
    expect(isinstance(False, int)).to_be(True)
    expect(isinstance(1, bool)).to_be(False)
    expect(isinstance(0, bool)).to_be(False)

def test_bool_issubclass():
    expect(issubclass(bool, int)).to_be(True)
    expect(issubclass(int, bool)).to_be(False)
    expect(issubclass(bool, bool)).to_be(True)

def test_bool_comparison():
    expect(True == 1).to_be(True)
    expect(False == 0).to_be(True)
    expect(True != 0).to_be(True)
    expect(False != 1).to_be(True)
    expect(True > False).to_be(True)
    expect(False < True).to_be(True)
    expect(True >= True).to_be(True)
    expect(False <= False).to_be(True)

def test_bool_logical_ops():
    expect(True and True).to_be(True)
    expect(True and False).to_be(False)
    expect(False and True).to_be(False)
    expect(False and False).to_be(False)
    expect(True or True).to_be(True)
    expect(True or False).to_be(True)
    expect(False or True).to_be(True)
    expect(False or False).to_be(False)
    expect(not True).to_be(False)
    expect(not False).to_be(True)

def test_bool_bitwise():
    expect(True & True).to_be(True)
    expect(True & False).to_be(False)
    expect(True | False).to_be(True)
    expect(False | False).to_be(False)
    expect(True ^ True).to_be(False)
    expect(True ^ False).to_be(True)

def test_bool_hash():
    expect(hash(True)).to_be(hash(1))
    expect(hash(False)).to_be(hash(0))

def test_bool_as_index():
    lst = ["a", "b"]
    expect(lst[True]).to_be("b")
    expect(lst[False]).to_be("a")

def test_bool_in_containers():
    expect(True in [1, 2, 3]).to_be(True)
    expect(False in [0, 1, 2]).to_be(True)
    d = {True: "yes", False: "no"}
    expect(d[True]).to_be("yes")
    expect(d[False]).to_be("no")

def test_bool_string_format():
    expect(f"{True}").to_be("True")
    expect(f"{False}").to_be("False")

def test_bool_arithmetic_types():
    # Bool + bool = int
    result = True + True
    expect(result).to_be(2)
    result = True - True
    expect(result).to_be(0)
    result = True * True
    expect(result).to_be(1)

def test_bool_truthiness_custom():
    class AlwaysTrue:
        def __bool__(self):
            return True
    class AlwaysFalse:
        def __bool__(self):
            return False
    expect(bool(AlwaysTrue())).to_be(True)
    expect(bool(AlwaysFalse())).to_be(False)

def test_bool_truthiness_len():
    class Empty:
        def __len__(self):
            return 0
    class NonEmpty:
        def __len__(self):
            return 5
    expect(bool(Empty())).to_be(False)
    expect(bool(NonEmpty())).to_be(True)

# Register all tests
test("bool_repr", test_bool_repr)
test("bool_str", test_bool_str)
test("bool_int_conversion", test_bool_int_conversion)
test("bool_float_conversion", test_bool_float_conversion)
test("bool_constructor", test_bool_constructor)
test("bool_isinstance", test_bool_isinstance)
test("bool_issubclass", test_bool_issubclass)
test("bool_comparison", test_bool_comparison)
test("bool_logical_ops", test_bool_logical_ops)
test("bool_bitwise", test_bool_bitwise)
test("bool_hash", test_bool_hash)
test("bool_as_index", test_bool_as_index)
test("bool_in_containers", test_bool_in_containers)
test("bool_string_format", test_bool_string_format)
test("bool_arithmetic_types", test_bool_arithmetic_types)
test("bool_truthiness_custom", test_bool_truthiness_custom)
test("bool_truthiness_len", test_bool_truthiness_len)

print("CPython bool tests completed")
