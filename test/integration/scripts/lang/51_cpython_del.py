# Test: CPython del Statement Tests
# Adapted from CPython's test_del.py and related tests

from test_framework import test, expect

# === Del variable and reassign ===
def test_del_and_reassign():
    x = 10
    del x
    x = 20
    expect(x).to_be(20)

# === Del variable in various contexts ===
def test_del_variable_basic():
    x = 42
    expect(x).to_be(42)
    del x
    # After del, reassign works fine
    x = 100
    expect(x).to_be(100)

# === Del list element by index ===
def test_del_list_element():
    lst = [1, 2, 3, 4, 5]
    del lst[2]
    expect(lst).to_be([1, 2, 4, 5])

def test_del_list_first():
    lst = [10, 20, 30]
    del lst[0]
    expect(lst).to_be([20, 30])

def test_del_list_last():
    lst = [10, 20, 30]
    del lst[2]
    expect(lst).to_be([10, 20])

# === Del list with negative index ===
def test_del_negative_index():
    lst = [1, 2, 3, 4, 5]
    del lst[-1]
    expect(lst).to_be([1, 2, 3, 4])

def test_del_negative_index_middle():
    lst = [10, 20, 30, 40]
    del lst[-2]
    expect(lst).to_be([10, 20, 40])

# === Del list slice ===
def test_del_list_slice():
    lst = [1, 2, 3, 4, 5]
    del lst[1:3]
    expect(lst).to_be([1, 4, 5])

def test_del_list_slice_from_start():
    lst = [1, 2, 3, 4, 5]
    del lst[:2]
    expect(lst).to_be([3, 4, 5])

def test_del_list_slice_to_end():
    lst = [1, 2, 3, 4, 5]
    del lst[3:]
    expect(lst).to_be([1, 2, 3])

def test_del_list_slice_all():
    lst = [1, 2, 3]
    del lst[:]
    expect(lst).to_be([])

# === Del dict key ===
def test_del_dict_key():
    d = {"a": 1, "b": 2, "c": 3}
    del d["b"]
    expect("b" in d).to_be(False)
    expect(len(d)).to_be(2)
    expect(d["a"]).to_be(1)
    expect(d["c"]).to_be(3)

def test_del_dict_int_key():
    d = {1: "one", 2: "two", 3: "three"}
    del d[2]
    expect(2 in d).to_be(False)
    expect(len(d)).to_be(2)

# === Del missing dict key raises KeyError ===
def test_del_missing_dict_key():
    d = {"a": 1}
    try:
        del d["z"]
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

# === Del attribute using delattr builtin ===
def test_delattr_basic():
    class Obj:
        pass
    o = Obj()
    o.x = 42
    expect(o.x).to_be(42)
    delattr(o, "x")
    expect(hasattr(o, "x")).to_be(False)

def test_delattr_then_access():
    class Obj:
        pass
    o = Obj()
    o.name = "hello"
    delattr(o, "name")
    try:
        y = o.name
        expect("no error").to_be("error")
    except Exception:
        # Attribute access after delattr raises an error
        expect(True).to_be(True)

# === Del nonexistent attribute raises AttributeError ===
def test_delattr_nonexistent():
    class Obj:
        pass
    o = Obj()
    try:
        delattr(o, "missing")
        expect("no error").to_be("error")
    except Exception:
        # Deleting nonexistent attribute raises an error
        expect(True).to_be(True)

# === Custom __delitem__ ===
def test_custom_delitem():
    class MyContainer:
        def __init__(self):
            self.deleted_keys = []
            self.data = {"a": 1, "b": 2}
        def __delitem__(self, key):
            self.deleted_keys.append(key)
            if key in self.data:
                # manually remove from data dict
                new_data = {}
                for k in self.data:
                    if k != key:
                        new_data[k] = self.data[k]
                self.data = new_data
    c = MyContainer()
    del c["a"]
    expect(c.deleted_keys).to_be(["a"])
    expect("a" in c.data).to_be(False)
    expect(c.data["b"]).to_be(2)

# === hasattr after delattr ===
def test_hasattr_after_delattr():
    class Obj:
        pass
    o = Obj()
    o.x = 10
    o.y = 20
    expect(hasattr(o, "x")).to_be(True)
    expect(hasattr(o, "y")).to_be(True)
    delattr(o, "x")
    expect(hasattr(o, "x")).to_be(False)
    expect(hasattr(o, "y")).to_be(True)
    expect(o.y).to_be(20)

# === Del multiple dict keys sequentially ===
def test_del_multiple_dict_keys():
    d = {"a": 1, "b": 2, "c": 3, "d": 4}
    del d["a"]
    del d["c"]
    expect(len(d)).to_be(2)
    expect("a" in d).to_be(False)
    expect("c" in d).to_be(False)
    expect(d["b"]).to_be(2)
    expect(d["d"]).to_be(4)

# === Del in loop ===
def test_del_in_loop():
    d = {"a": 1, "b": 2, "c": 3}
    keys_to_delete = ["a", "c"]
    for k in keys_to_delete:
        del d[k]
    expect(len(d)).to_be(1)
    expect(d["b"]).to_be(2)

# === Del list element updates length ===
def test_del_updates_length():
    lst = [10, 20, 30, 40, 50]
    expect(len(lst)).to_be(5)
    del lst[0]
    expect(len(lst)).to_be(4)
    del lst[0]
    expect(len(lst)).to_be(3)
    expect(lst[0]).to_be(30)

# === Del slice preserves surrounding elements ===
def test_del_list_slice_middle():
    lst = [0, 1, 2, 3, 4, 5, 6, 7]
    del lst[2:6]
    expect(lst).to_be([0, 1, 6, 7])

# === Del from dict then re-add ===
def test_del_dict_readd():
    d = {"x": 1}
    del d["x"]
    expect("x" in d).to_be(False)
    d["x"] = 99
    expect(d["x"]).to_be(99)

# === Del multiple attributes with delattr ===
def test_delattr_multiple():
    class Obj:
        pass
    o = Obj()
    o.a = 1
    o.b = 2
    o.c = 3
    delattr(o, "a")
    delattr(o, "c")
    expect(hasattr(o, "a")).to_be(False)
    expect(hasattr(o, "b")).to_be(True)
    expect(hasattr(o, "c")).to_be(False)
    expect(o.b).to_be(2)

# Register all tests
test("del_and_reassign", test_del_and_reassign)
test("del_variable_basic", test_del_variable_basic)
test("del_list_element", test_del_list_element)
test("del_list_first", test_del_list_first)
test("del_list_last", test_del_list_last)
test("del_negative_index", test_del_negative_index)
test("del_negative_index_middle", test_del_negative_index_middle)
test("del_list_slice", test_del_list_slice)
test("del_list_slice_from_start", test_del_list_slice_from_start)
test("del_list_slice_to_end", test_del_list_slice_to_end)
test("del_list_slice_all", test_del_list_slice_all)
test("del_dict_key", test_del_dict_key)
test("del_dict_int_key", test_del_dict_int_key)
test("del_missing_dict_key", test_del_missing_dict_key)
test("delattr_basic", test_delattr_basic)
test("delattr_then_access", test_delattr_then_access)
test("delattr_nonexistent", test_delattr_nonexistent)
test("custom_delitem", test_custom_delitem)
test("hasattr_after_delattr", test_hasattr_after_delattr)
test("del_multiple_dict_keys", test_del_multiple_dict_keys)
test("del_in_loop", test_del_in_loop)
test("del_updates_length", test_del_updates_length)
test("del_list_slice_middle", test_del_list_slice_middle)
test("del_dict_readd", test_del_dict_readd)
test("delattr_multiple", test_delattr_multiple)

print("CPython del tests completed")
