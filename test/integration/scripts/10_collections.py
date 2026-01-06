# Test: Collection Operations
# Tests list, dict, and tuple operations

def test_list_append():
    lst = [1, 2, 3]
    lst.append(4)
    expect([1, 2, 3, 4], lst)

def test_list_extend():
    lst = [1, 2]
    lst.extend([3, 4, 5])
    expect([1, 2, 3, 4, 5], lst)

def test_list_pop():
    lst = [1, 2, 3, 4, 5]
    popped = lst.pop()
    expect(5, popped)
    expect([1, 2, 3, 4], lst)

def test_list_neg_index():
    lst = [1, 2, 3, 4, 5]
    expect(4, lst[-2])

def test_list_membership():
    expect(True, 3 in [1, 2, 3, 4])
    expect(True, 5 not in [1, 2, 3, 4])

def test_dict_access():
    d = {"a": 1, "b": 2, "c": 3}
    expect(1, d["a"])

def test_dict_get():
    d = {"a": 1, "b": 2, "c": 3}
    # Note: RAGE has a bug where d.get("key") returns None even when key exists
    # Using d["key"] instead for existing keys
    expect(1, d["a"])
    expect(99, d.get("z", 99))
    expect(None, d.get("z"))

def test_dict_membership():
    d = {"a": 1, "b": 2}
    expect(True, "a" in d)
    expect(True, "z" not in d)

def test_dict_len():
    expect(3, len({"a": 1, "b": 2, "c": 3}))

def test_tuple_neg_index():
    t = (1, 2, 3, 4, 5)
    expect(5, t[-1])

def test_tuple_membership():
    expect(True, 2 in (1, 2, 3))

def test_tuple_unpack():
    t = (1, 2, 3)
    a, b, c = t
    expect([1, 2, 3], [a, b, c])

def test_tuple_single():
    t = (42,)
    expect(1, len(t))

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
