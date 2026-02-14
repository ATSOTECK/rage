# Test: CPython Dict Edge Cases
# Adapted from CPython's test_dict.py - covers edge cases beyond 10_collections.py

from test_framework import test, expect

def test_dict_construction():
    expect(dict()).to_be({})
    expect(dict(a=1, b=2)).to_be({"a": 1, "b": 2})
    expect(dict([("a", 1), ("b", 2)])).to_be({"a": 1, "b": 2})
    expect(dict({"a": 1, "b": 2})).to_be({"a": 1, "b": 2})

def test_dict_fromkeys():
    expect(dict.fromkeys(["a", "b", "c"])).to_be({"a": None, "b": None, "c": None})
    expect(dict.fromkeys(["a", "b"], 0)).to_be({"a": 0, "b": 0})
    expect(dict.fromkeys([])).to_be({})

def test_dict_setdefault():
    d = {"a": 1}
    expect(d.setdefault("a", 99)).to_be(1)
    expect(d.setdefault("b", 2)).to_be(2)
    expect(d).to_be({"a": 1, "b": 2})
    expect(d.setdefault("c")).to_be(None)
    expect(d["c"]).to_be(None)

def test_dict_update_dict():
    d = {"a": 1, "b": 2}
    d.update({"c": 3, "b": 20})
    expect(d).to_be({"a": 1, "b": 20, "c": 3})

def test_dict_update_kwargs():
    d = {"a": 1}
    d.update(b=2, c=3)
    expect(d["b"]).to_be(2)
    expect(d["c"]).to_be(3)

def test_dict_update_pairs():
    d = {"a": 1}
    d.update([("b", 2), ("c", 3)])
    expect(d).to_be({"a": 1, "b": 2, "c": 3})

def test_dict_pop():
    d = {"a": 1, "b": 2}
    expect(d.pop("a")).to_be(1)
    expect(d).to_be({"b": 2})
    expect(d.pop("z", 99)).to_be(99)
    try:
        d.pop("z")
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_dict_popitem():
    d = {"a": 1}
    item = d.popitem()
    expect(item).to_be(("a", 1))
    expect(d).to_be({})
    try:
        d.popitem()
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_dict_clear():
    d = {"a": 1, "b": 2, "c": 3}
    d.clear()
    expect(d).to_be({})
    expect(len(d)).to_be(0)

def test_dict_copy():
    d = {"a": 1, "b": [2, 3]}
    c = d.copy()
    expect(c).to_be({"a": 1, "b": [2, 3]})
    c["a"] = 99
    expect(d["a"]).to_be(1)  # Independent
    # Shallow: inner list is shared
    c["b"].append(4)
    expect(d["b"]).to_be([2, 3, 4])

def test_dict_keys_view():
    d = {"a": 1, "b": 2, "c": 3}
    keys = list(d.keys())
    expect(len(keys)).to_be(3)
    expect("a" in d.keys()).to_be(True)
    expect("z" in d.keys()).to_be(False)

def test_dict_values_view():
    d = {"a": 1, "b": 2, "c": 3}
    vals = sorted(list(d.values()))
    expect(vals).to_be([1, 2, 3])

def test_dict_items_view():
    d = {"a": 1, "b": 2}
    items = sorted(list(d.items()))
    expect(items).to_be([("a", 1), ("b", 2)])
    expect(("a", 1) in d.items()).to_be(True)
    expect(("a", 2) in d.items()).to_be(False)

def test_dict_iteration_order():
    # Insertion order is preserved (Python 3.7+)
    d = {}
    d["b"] = 2
    d["a"] = 1
    d["c"] = 3
    expect(list(d.keys())).to_be(["b", "a", "c"])

def test_dict_reversed():
    d = {"a": 1, "b": 2, "c": 3}
    expect(list(reversed(d))).to_be(["c", "b", "a"])

def test_dict_merge_operator():
    d1 = {"a": 1, "b": 2}
    d2 = {"b": 3, "c": 4}
    merged = d1 | d2
    expect(merged).to_be({"a": 1, "b": 3, "c": 4})
    # Original unchanged
    expect(d1).to_be({"a": 1, "b": 2})
    # In-place merge
    d1 |= d2
    expect(d1).to_be({"a": 1, "b": 3, "c": 4})

def test_dict_eq():
    expect({"a": 1, "b": 2} == {"a": 1, "b": 2}).to_be(True)
    expect({"a": 1, "b": 2} == {"b": 2, "a": 1}).to_be(True)
    expect({"a": 1} != {"a": 2}).to_be(True)
    expect({} == {}).to_be(True)

def test_dict_bool():
    expect(bool({})).to_be(False)
    expect(bool({"a": 1})).to_be(True)

def test_dict_repr():
    expect(repr({})).to_be("{}")

def test_dict_del_item():
    d = {"a": 1, "b": 2, "c": 3}
    del d["b"]
    expect(d).to_be({"a": 1, "c": 3})
    try:
        del d["z"]
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_dict_contains():
    d = {"a": 1, "b": 2}
    expect("a" in d).to_be(True)
    expect("z" not in d).to_be(True)

def test_dict_len():
    expect(len({})).to_be(0)
    d = {"a": 1}
    expect(len(d)).to_be(1)
    d["b"] = 2
    expect(len(d)).to_be(2)
    del d["a"]
    expect(len(d)).to_be(1)

def test_dict_unhashable_key():
    try:
        {[1, 2]: "value"}
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_dict_none_key():
    d = {None: 1}
    expect(d[None]).to_be(1)

def test_dict_int_float_key_collision():
    # 1 == 1.0 so they collide as dict keys
    d = {1: "a", 1.0: "b"}
    expect(len(d)).to_be(1)
    expect(d[1]).to_be("b")

def test_dict_comprehension():
    d = {k: v for k, v in [("a", 1), ("b", 2), ("c", 3)]}
    expect(d).to_be({"a": 1, "b": 2, "c": 3})
    d2 = {x: x * x for x in range(5)}
    expect(d2).to_be({0: 0, 1: 1, 2: 4, 3: 9, 4: 16})

def test_dict_nested():
    d = {"a": {"b": {"c": 42}}}
    expect(d["a"]["b"]["c"]).to_be(42)
    d["a"]["b"]["d"] = 99
    expect(d["a"]["b"]["d"]).to_be(99)

# Register all tests
test("dict_construction", test_dict_construction)
test("dict_fromkeys", test_dict_fromkeys)
test("dict_setdefault", test_dict_setdefault)
test("dict_update_dict", test_dict_update_dict)
test("dict_update_kwargs", test_dict_update_kwargs)
test("dict_update_pairs", test_dict_update_pairs)
test("dict_pop", test_dict_pop)
test("dict_popitem", test_dict_popitem)
test("dict_clear", test_dict_clear)
test("dict_copy", test_dict_copy)
test("dict_keys_view", test_dict_keys_view)
test("dict_values_view", test_dict_values_view)
test("dict_items_view", test_dict_items_view)
test("dict_iteration_order", test_dict_iteration_order)
test("dict_reversed", test_dict_reversed)
test("dict_merge_operator", test_dict_merge_operator)
test("dict_eq", test_dict_eq)
test("dict_bool", test_dict_bool)
test("dict_repr", test_dict_repr)
test("dict_del_item", test_dict_del_item)
test("dict_contains", test_dict_contains)
test("dict_len", test_dict_len)
test("dict_unhashable_key", test_dict_unhashable_key)
test("dict_none_key", test_dict_none_key)
test("dict_int_float_key_collision", test_dict_int_float_key_collision)
test("dict_comprehension", test_dict_comprehension)
test("dict_nested", test_dict_nested)

print("CPython dict tests completed")
