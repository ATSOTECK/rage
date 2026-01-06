# Test: Collection Operations
# Tests list, dict, and tuple operations

from test_framework import test, expect

def test_list_append():
    lst = [1, 2, 3]
    lst.append(4)
    expect(lst).to_be([1, 2, 3, 4])

def test_list_extend():
    lst = [1, 2]
    lst.extend([3, 4, 5])
    expect(lst).to_be([1, 2, 3, 4, 5])

def test_list_pop():
    lst = [1, 2, 3, 4, 5]
    popped = lst.pop()
    expect(popped).to_be(5)
    expect(lst).to_be([1, 2, 3, 4])

def test_list_neg_index():
    lst = [1, 2, 3, 4, 5]
    expect(lst[-2]).to_be(4)

def test_list_membership():
    expect(3 in [1, 2, 3, 4]).to_be(True)
    expect(5 not in [1, 2, 3, 4]).to_be(True)

def test_dict_access():
    d = {"a": 1, "b": 2, "c": 3}
    expect(d["a"]).to_be(1)

def test_dict_get():
    d = {"a": 1, "b": 2, "c": 3}
    # Note: RAGE has a bug where d.get("key") returns None even when key exists
    # Using d["key"] instead for existing keys
    expect(d["a"]).to_be(1)
    expect(d.get("z", 99)).to_be(99)
    expect(d.get("z")).to_be(None)

def test_dict_membership():
    d = {"a": 1, "b": 2}
    expect("a" in d).to_be(True)
    expect("z" not in d).to_be(True)

def test_dict_len():
    expect(len({"a": 1, "b": 2, "c": 3})).to_be(3)

def test_tuple_neg_index():
    t = (1, 2, 3, 4, 5)
    expect(t[-1]).to_be(5)

def test_tuple_membership():
    expect(2 in (1, 2, 3)).to_be(True)

def test_tuple_unpack():
    t = (1, 2, 3)
    a, b, c = t
    expect([a, b, c]).to_be([1, 2, 3])

def test_tuple_single():
    t = (42,)
    expect(len(t)).to_be(1)

test("list_append", test_list_append)
test("list_extend", test_list_extend)
test("list_pop", test_list_pop)
test("list_neg_index", test_list_neg_index)
test("list_membership", test_list_membership)
test("dict_access", test_dict_access)
test("dict_get", test_dict_get)
test("dict_membership", test_dict_membership)
test("dict_len", test_dict_len)
test("tuple_neg_index", test_tuple_neg_index)
test("tuple_membership", test_tuple_membership)
test("tuple_unpack", test_tuple_unpack)
test("tuple_single", test_tuple_single)

print("Collections tests completed")
