from test_framework import test, expect

# Test 1: __index__ used for list indexing
class MyIdx:
    def __init__(self, val):
        self.val = val
    def __index__(self):
        return self.val

def test_list_indexing():
    lst = [10, 20, 30, 40, 50]
    idx = MyIdx(2)
    expect(lst[idx]).to_be(30)
test("__index__ for list indexing", test_list_indexing)

# Test 2: __index__ used for tuple indexing
def test_tuple_indexing():
    tup = (10, 20, 30, 40, 50)
    idx = MyIdx(3)
    expect(tup[idx]).to_be(40)
test("__index__ for tuple indexing", test_tuple_indexing)

# Test 3: __index__ used for string indexing
def test_string_indexing():
    s = "hello"
    idx = MyIdx(1)
    expect(s[idx]).to_be("e")
test("__index__ for string indexing", test_string_indexing)

# Test 4: __index__ with negative index
def test_negative_index():
    lst = [10, 20, 30]
    idx = MyIdx(-1)
    expect(lst[idx]).to_be(30)
test("__index__ with negative index", test_negative_index)

# Test 5: __index__ used in hex()
def test_hex():
    idx = MyIdx(255)
    expect(hex(idx)).to_be("0xff")
test("__index__ in hex()", test_hex)

# Test 6: __index__ used in oct()
def test_oct():
    idx = MyIdx(8)
    expect(oct(idx)).to_be("0o10")
test("__index__ in oct()", test_oct)

# Test 7: __index__ used in bin()
def test_bin():
    idx = MyIdx(10)
    expect(bin(idx)).to_be("0b1010")
test("__index__ in bin()", test_bin)

# Test 8: __index__ for list assignment
def test_list_assignment():
    lst = [10, 20, 30]
    idx = MyIdx(1)
    lst[idx] = 99
    expect(lst[1]).to_be(99)
test("__index__ for list assignment", test_list_assignment)

# Test 9: __index__ for list deletion
def test_list_deletion():
    lst = [10, 20, 30, 40]
    idx = MyIdx(1)
    del lst[idx]
    expect(len(lst)).to_be(3)
    expect(lst[0]).to_be(10)
    expect(lst[1]).to_be(30)
test("__index__ for list deletion", test_list_deletion)

# Test 10: __index__ returning non-int raises TypeError
def test_index_non_int():
    class BadIdx:
        def __index__(self):
            return "not an int"
    try:
        lst = [1, 2, 3]
        _ = lst[BadIdx()]
        expect("should have raised").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)
test("__index__ returning non-int raises TypeError", test_index_non_int)

# Test 11: No __index__ raises TypeError
def test_no_index():
    class NoIdx:
        pass
    try:
        lst = [1, 2, 3]
        _ = lst[NoIdx()]
        expect("should have raised").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)
test("no __index__ raises TypeError", test_no_index)

# Test 12: __index__ in operator.index()
from operator import index as op_index

def test_operator_index():
    idx = MyIdx(42)
    expect(op_index(idx)).to_be(42)
test("operator.index()", test_operator_index)

# Test 13: operator.index() with plain int
def test_operator_index_int():
    expect(op_index(5)).to_be(5)
test("operator.index() with int", test_operator_index_int)

# Test 14: operator.index() with bool
def test_operator_index_bool():
    expect(op_index(True)).to_be(1)
    expect(op_index(False)).to_be(0)
test("operator.index() with bool", test_operator_index_bool)

# Test 15: __index__ with inheritance
class DerivedIdx(MyIdx):
    pass

def test_inherited_index():
    idx = DerivedIdx(4)
    lst = [0, 1, 2, 3, 4, 5]
    expect(lst[idx]).to_be(4)
test("__index__ with inheritance", test_inherited_index)
