# Test: CPython Membership Testing Edge Cases
# Adapted from CPython's test_contains.py - covers in/not in across types

from test_framework import test, expect

# === String membership ===
def test_string_char_membership():
    expect("a" in "abc").to_be(True)
    expect("d" in "abc").to_be(False)
    expect("z" not in "abc").to_be(True)

def test_string_substring_membership():
    expect("bc" in "abcd").to_be(True)
    expect("ab" in "abcd").to_be(True)
    expect("cd" in "abcd").to_be(True)
    expect("abcd" in "abcd").to_be(True)
    expect("ac" in "abcd").to_be(False)

def test_string_empty_membership():
    expect("" in "hello").to_be(True)
    expect("" in "").to_be(True)
    expect("a" in "").to_be(False)

# === List membership ===
def test_list_value_membership():
    expect(1 in [1, 2, 3]).to_be(True)
    expect(4 in [1, 2, 3]).to_be(False)
    expect(4 not in [1, 2, 3]).to_be(True)

def test_list_string_membership():
    expect("hello" in ["hello", "world"]).to_be(True)
    expect("hi" in ["hello", "world"]).to_be(False)

def test_list_none_membership():
    expect(None in [1, None, 3]).to_be(True)
    expect(None in [1, 2, 3]).to_be(False)

def test_list_bool_membership():
    # True == 1 and False == 0 for membership
    expect(True in [1, 2, 3]).to_be(True)
    expect(False in [0, 1, 2]).to_be(True)

def test_nested_list_membership():
    expect([1, 2] in [[1, 2], [3, 4]]).to_be(True)
    expect([1, 3] in [[1, 2], [3, 4]]).to_be(False)
    expect([] in [[], [1]]).to_be(True)

# === Dict membership (checks keys) ===
def test_dict_key_membership():
    d = {"a": 1, "b": 2, "c": 3}
    expect("a" in d).to_be(True)
    expect("d" in d).to_be(False)
    expect("d" not in d).to_be(True)

def test_dict_value_not_checked():
    d = {"a": 1, "b": 2}
    # 'in' checks keys, not values
    expect(1 in d).to_be(False)
    expect("a" in d).to_be(True)

def test_dict_none_key():
    d = {None: "value"}
    expect(None in d).to_be(True)

# === Tuple membership ===
def test_tuple_membership():
    expect(1 in (1, 2, 3)).to_be(True)
    expect(4 in (1, 2, 3)).to_be(False)
    expect(4 not in (1, 2, 3)).to_be(True)

def test_tuple_empty_membership():
    expect(1 in ()).to_be(False)

# === Set membership ===
def test_set_membership():
    expect(1 in {1, 2, 3}).to_be(True)
    expect(4 in {1, 2, 3}).to_be(False)
    expect(4 not in {1, 2, 3}).to_be(True)

def test_set_empty_membership():
    expect(1 in set()).to_be(False)

# === Custom __contains__ ===
def test_custom_contains():
    class EvenContainer:
        def __contains__(self, item):
            return item % 2 == 0
    ec = EvenContainer()
    expect(2 in ec).to_be(True)
    expect(4 in ec).to_be(True)
    expect(3 in ec).to_be(False)
    expect(3 not in ec).to_be(True)

# === Custom __iter__ fallback ===
def test_custom_iter_fallback():
    class IterContainer:
        def __init__(self, data):
            self.data = data
        def __iter__(self):
            return iter(self.data)
    ic = IterContainer([10, 20, 30])
    expect(20 in ic).to_be(True)
    expect(40 in ic).to_be(False)

# === not in operator ===
def test_not_in_various():
    expect(5 not in [1, 2, 3]).to_be(True)
    expect(1 not in [1, 2, 3]).to_be(False)
    expect("x" not in "hello").to_be(True)
    expect("h" not in "hello").to_be(False)
    expect(5 not in {1, 2, 3}).to_be(True)
    expect(1 not in {1, 2, 3}).to_be(False)

# === Edge cases ===
def test_membership_type_mismatch():
    # Searching for wrong type does not error, just returns False
    expect("a" in [1, 2, 3]).to_be(False)
    expect(1 in ["a", "b"]).to_be(False)

# Register all tests
test("string_char_membership", test_string_char_membership)
test("string_substring_membership", test_string_substring_membership)
test("string_empty_membership", test_string_empty_membership)
test("list_value_membership", test_list_value_membership)
test("list_string_membership", test_list_string_membership)
test("list_none_membership", test_list_none_membership)
test("list_bool_membership", test_list_bool_membership)
test("nested_list_membership", test_nested_list_membership)
test("dict_key_membership", test_dict_key_membership)
test("dict_value_not_checked", test_dict_value_not_checked)
test("dict_none_key", test_dict_none_key)
test("tuple_membership", test_tuple_membership)
test("tuple_empty_membership", test_tuple_empty_membership)
test("set_membership", test_set_membership)
test("set_empty_membership", test_set_empty_membership)
test("custom_contains", test_custom_contains)
test("custom_iter_fallback", test_custom_iter_fallback)
test("not_in_various", test_not_in_various)
test("membership_type_mismatch", test_membership_type_mismatch)

print("CPython membership tests completed")
