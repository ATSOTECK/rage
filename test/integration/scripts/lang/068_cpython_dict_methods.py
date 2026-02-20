# Test: CPython Dict Methods - Deep Dive
# Adapted from CPython's test_dict.py - covers additional dict method edge cases
# beyond 10_collections.py and 32_cpython_dict.py

from test_framework import test, expect

def test_dict_update_overwrite():
    d = {"a": 1, "b": 2, "c": 3}
    d.update({"a": 10, "d": 4})
    expect(d["a"]).to_be(10)
    expect(d["b"]).to_be(2)
    expect(d["c"]).to_be(3)
    expect(d["d"]).to_be(4)

def test_dict_update_empty():
    d = {"a": 1}
    d.update({})
    expect(d).to_be({"a": 1})
    d.update()
    expect(d).to_be({"a": 1})

def test_dict_update_from_list_of_pairs():
    d = {}
    d.update([("x", 10), ("y", 20)])
    expect(d).to_be({"x": 10, "y": 20})
    # Update with duplicate keys in list - last wins
    d2 = {}
    d2.update([("a", 1), ("a", 2)])
    expect(d2["a"]).to_be(2)

def test_dict_setdefault_existing():
    d = {"a": 1, "b": 2}
    result = d.setdefault("a", 99)
    expect(result).to_be(1)
    expect(d["a"]).to_be(1)  # Unchanged

def test_dict_setdefault_new():
    d = {"a": 1}
    result = d.setdefault("b", 42)
    expect(result).to_be(42)
    expect(d["b"]).to_be(42)
    # setdefault with no default value
    result2 = d.setdefault("c")
    expect(result2).to_be(None)
    expect(d["c"]).to_be(None)

def test_dict_setdefault_none_value():
    d = {"a": None}
    result = d.setdefault("a", 99)
    expect(result).to_be(None)  # Key exists, even though value is None

def test_dict_pop_existing():
    d = {"a": 1, "b": 2, "c": 3}
    result = d.pop("b")
    expect(result).to_be(2)
    expect(len(d)).to_be(2)
    expect("b" in d).to_be(False)

def test_dict_pop_with_default():
    d = {"a": 1}
    result = d.pop("missing", "default_val")
    expect(result).to_be("default_val")
    expect(len(d)).to_be(1)
    # Default can be None
    result2 = d.pop("missing2", None)
    expect(result2).to_be(None)

def test_dict_pop_missing_no_default():
    d = {"a": 1}
    try:
        d.pop("missing")
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_dict_popitem_lifo():
    # popitem removes last inserted item (LIFO order since Python 3.7)
    d = {}
    d["first"] = 1
    d["second"] = 2
    d["third"] = 3
    item = d.popitem()
    expect(item).to_be(("third", 3))
    item2 = d.popitem()
    expect(item2).to_be(("second", 2))
    expect(d).to_be({"first": 1})

def test_dict_popitem_empty():
    d = {}
    try:
        d.popitem()
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_dict_copy_independence():
    original = {"a": 1, "b": 2}
    copied = original.copy()
    copied["c"] = 3
    expect("c" in original).to_be(False)
    expect("c" in copied).to_be(True)
    copied["a"] = 99
    expect(original["a"]).to_be(1)

def test_dict_copy_shallow():
    # Shallow copy - nested objects are shared
    original = {"a": [1, 2, 3]}
    copied = original.copy()
    copied["a"].append(4)
    expect(original["a"]).to_be([1, 2, 3, 4])  # Shared reference

def test_dict_clear_and_reuse():
    d = {"a": 1, "b": 2, "c": 3}
    d.clear()
    expect(d).to_be({})
    expect(len(d)).to_be(0)
    # Can reuse after clear
    d["x"] = 10
    expect(d).to_be({"x": 10})

def test_dict_comprehension_conditional():
    # Dict comprehension with condition
    d = {x: x * x for x in range(10) if x % 2 == 0}
    expect(d).to_be({0: 0, 2: 4, 4: 16, 6: 36, 8: 64})

def test_dict_comprehension_from_dict():
    # Dict comprehension from another dict
    original = {"a": 1, "b": 2, "c": 3}
    doubled = {k: v * 2 for k, v in original.items()}
    expect(doubled).to_be({"a": 2, "b": 4, "c": 6})

def test_dict_equality_order_independent():
    d1 = {"a": 1, "b": 2, "c": 3}
    d2 = {"c": 3, "a": 1, "b": 2}
    expect(d1 == d2).to_be(True)
    expect(d1 != d2).to_be(False)

def test_dict_equality_different_values():
    d1 = {"a": 1, "b": 2}
    d2 = {"a": 1, "b": 3}
    expect(d1 == d2).to_be(False)
    expect(d1 != d2).to_be(True)

def test_dict_equality_different_keys():
    d1 = {"a": 1, "b": 2}
    d2 = {"a": 1, "c": 2}
    expect(d1 == d2).to_be(False)

def test_dict_bool_key_collision():
    # True == 1 and False == 0, so they collide as dict keys
    d = {True: "yes", 1: "one"}
    expect(len(d)).to_be(1)
    expect(d[True]).to_be("one")
    expect(d[1]).to_be("one")

    d2 = {False: "no", 0: "zero"}
    expect(len(d2)).to_be(1)
    expect(d2[False]).to_be("zero")
    expect(d2[0]).to_be("zero")

def test_dict_none_key_operations():
    d = {None: "nothing"}
    expect(d[None]).to_be("nothing")
    expect(None in d).to_be(True)
    d[None] = "updated"
    expect(d[None]).to_be("updated")
    del d[None]
    expect(None in d).to_be(False)

def test_dict_tuple_keys():
    d = {(1, 2): "a", (3, 4): "b"}
    expect(d[(1, 2)]).to_be("a")
    expect((1, 2) in d).to_be(True)
    expect((5, 6) in d).to_be(False)

def test_dict_int_str_keys_no_collision():
    # int and str keys don't collide
    d = {1: "int", "1": "str"}
    expect(len(d)).to_be(2)
    expect(d[1]).to_be("int")
    expect(d["1"]).to_be("str")

def test_dict_nested_access():
    d = {"a": {"b": {"c": {"d": 42}}}}
    expect(d["a"]["b"]["c"]["d"]).to_be(42)
    # Modify nested
    d["a"]["b"]["c"]["d"] = 99
    expect(d["a"]["b"]["c"]["d"]).to_be(99)
    # Add new nested key
    d["a"]["b"]["e"] = "new"
    expect(d["a"]["b"]["e"]).to_be("new")

def test_dict_nested_with_lists():
    d = {"users": [{"name": "Alice"}, {"name": "Bob"}]}
    expect(d["users"][0]["name"]).to_be("Alice")
    expect(d["users"][1]["name"]).to_be("Bob")
    d["users"].append({"name": "Charlie"})
    expect(len(d["users"])).to_be(3)

def test_dict_get_method():
    d = {"a": 1, "b": 2}
    expect(d.get("a")).to_be(1)
    expect(d.get("c")).to_be(None)
    expect(d.get("c", 99)).to_be(99)
    # get does not modify the dict
    expect("c" in d).to_be(False)

def test_dict_get_none_value():
    d = {"a": None}
    expect(d.get("a")).to_be(None)
    expect(d.get("a", "default")).to_be(None)  # Key exists, returns None

def test_dict_keys_values_items_empty():
    d = {}
    expect(list(d.keys())).to_be([])
    expect(list(d.values())).to_be([])
    expect(list(d.items())).to_be([])

def test_dict_iteration_during_build():
    # Build a dict by iterating over a list of pairs
    pairs = [("a", 1), ("b", 2), ("c", 3)]
    d = {}
    for k, v in pairs:
        d[k] = v
    expect(d).to_be({"a": 1, "b": 2, "c": 3})

def test_dict_counting_pattern():
    # Common counting pattern
    text = "abracadabra"
    counts = {}
    for ch in text:
        if ch in counts:
            counts[ch] = counts[ch] + 1
        else:
            counts[ch] = 1
    expect(counts["a"]).to_be(5)
    expect(counts["b"]).to_be(2)
    expect(counts["r"]).to_be(2)
    expect(counts["c"]).to_be(1)
    expect(counts["d"]).to_be(1)

def test_dict_grouping_pattern():
    # Common grouping pattern
    words = ["apple", "banana", "avocado", "blueberry", "cherry", "apricot"]
    groups = {}
    for w in words:
        key = w[0]
        if key in groups:
            groups[key].append(w)
        else:
            groups[key] = [w]
    expect(sorted(groups["a"])).to_be(["apple", "apricot", "avocado"])
    expect(sorted(groups["b"])).to_be(["banana", "blueberry"])
    expect(groups["c"]).to_be(["cherry"])

def test_dict_merge_precedence():
    # When merging, right side wins
    d1 = {"a": 1, "b": 2}
    d2 = {"b": 20, "c": 30}
    merged = d1 | d2
    expect(merged).to_be({"a": 1, "b": 20, "c": 30})
    # Reverse merge
    merged2 = d2 | d1
    expect(merged2).to_be({"b": 2, "c": 30, "a": 1})

def test_dict_inplace_merge():
    d = {"a": 1}
    d |= {"b": 2, "c": 3}
    expect(d).to_be({"a": 1, "b": 2, "c": 3})
    d |= {"a": 99}
    expect(d["a"]).to_be(99)

# Register all tests
test("dict_update_overwrite", test_dict_update_overwrite)
test("dict_update_empty", test_dict_update_empty)
test("dict_update_from_list_of_pairs", test_dict_update_from_list_of_pairs)
test("dict_setdefault_existing", test_dict_setdefault_existing)
test("dict_setdefault_new", test_dict_setdefault_new)
test("dict_setdefault_none_value", test_dict_setdefault_none_value)
test("dict_pop_existing", test_dict_pop_existing)
test("dict_pop_with_default", test_dict_pop_with_default)
test("dict_pop_missing_no_default", test_dict_pop_missing_no_default)
test("dict_popitem_lifo", test_dict_popitem_lifo)
test("dict_popitem_empty", test_dict_popitem_empty)
test("dict_copy_independence", test_dict_copy_independence)
test("dict_copy_shallow", test_dict_copy_shallow)
test("dict_clear_and_reuse", test_dict_clear_and_reuse)
test("dict_comprehension_conditional", test_dict_comprehension_conditional)
test("dict_comprehension_from_dict", test_dict_comprehension_from_dict)
test("dict_equality_order_independent", test_dict_equality_order_independent)
test("dict_equality_different_values", test_dict_equality_different_values)
test("dict_equality_different_keys", test_dict_equality_different_keys)
test("dict_bool_key_collision", test_dict_bool_key_collision)
test("dict_none_key_operations", test_dict_none_key_operations)
test("dict_tuple_keys", test_dict_tuple_keys)
test("dict_int_str_keys_no_collision", test_dict_int_str_keys_no_collision)
test("dict_nested_access", test_dict_nested_access)
test("dict_nested_with_lists", test_dict_nested_with_lists)
test("dict_get_method", test_dict_get_method)
test("dict_get_none_value", test_dict_get_none_value)
test("dict_keys_values_items_empty", test_dict_keys_values_items_empty)
test("dict_iteration_during_build", test_dict_iteration_during_build)
test("dict_counting_pattern", test_dict_counting_pattern)
test("dict_grouping_pattern", test_dict_grouping_pattern)
test("dict_merge_precedence", test_dict_merge_precedence)
test("dict_inplace_merge", test_dict_inplace_merge)

print("CPython dict methods tests completed")
