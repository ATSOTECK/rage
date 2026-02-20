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

# Ported from CPython test_contains.py

# === __contains__ protocol on class hierarchy ===
def test_contains_class_hierarchy():
    """Test __contains__ with inheritance and __getitem__ fallback (CPython test_common_tests)"""
    class base_set:
        def __init__(self, el):
            self.el = el

    class myset(base_set):
        def __contains__(self, el):
            return self.el == el

    class seq(base_set):
        def __getitem__(self, n):
            return [self.el][n]

    b = myset(1)
    c = seq(1)
    # myset uses __contains__
    expect(1 in b).to_be(True)
    expect(0 in b).to_be(False)
    expect(1 not in b).to_be(False)
    expect(0 not in b).to_be(True)
    # seq uses __getitem__ fallback
    expect(1 in c).to_be(True)
    expect(0 in c).to_be(False)

test("contains_class_hierarchy", test_contains_class_hierarchy)

# === TypeError for non-iterable ===
def test_contains_typeerror_non_iterable():
    """Test that 'in' raises TypeError for objects with no __contains__/__iter__/__getitem__"""
    class base_set:
        def __init__(self, el):
            self.el = el

    a = base_set(1)
    raised = False
    try:
        1 in a
    except TypeError:
        raised = True
    expect(raised).to_be(True)

    raised2 = False
    try:
        1 not in a
    except TypeError:
        raised2 = True
    expect(raised2).to_be(True)

test("contains_typeerror_non_iterable", test_contains_typeerror_non_iterable)

# === None in string raises TypeError ===
def test_none_in_string_typeerror():
    """Test that None in 'abc' raises TypeError (CPython test_common_tests)"""
    raised = False
    try:
        None in "abc"
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("none_in_string_typeerror", test_none_in_string_typeerror)

# === __contains__ returning non-bool ===
def test_contains_nonbool_return():
    """Test that __contains__ returning truthy/falsy non-bool values works correctly"""
    class TruthyContainer:
        def __contains__(self, item):
            # Return a non-bool truthy value
            return 42

    class FalsyContainer:
        def __contains__(self, item):
            # Return a non-bool falsy value
            return 0

    class StringContainer:
        def __contains__(self, item):
            return "yes"

    class EmptyStringContainer:
        def __contains__(self, item):
            return ""

    tc = TruthyContainer()
    fc = FalsyContainer()
    sc = StringContainer()
    esc = EmptyStringContainer()
    # 'in' should coerce to bool
    expect(1 in tc).to_be(True)
    expect(1 in fc).to_be(False)
    expect(1 in sc).to_be(True)
    expect(1 in esc).to_be(False)
    # 'not in' should also work
    expect(1 not in tc).to_be(False)
    expect(1 not in fc).to_be(True)

test("contains_nonbool_return", test_contains_nonbool_return)

# === Range membership ===
def test_range_membership():
    """Test membership in range objects (CPython test_builtin_sequence_types)"""
    a = range(10)
    for i in a:
        expect(i in a).to_be(True)
    expect(16 in a).to_be(False)
    expect(-1 in a).to_be(False)

test("range_membership", test_range_membership)

# === Tuple membership from range ===
def test_tuple_from_range_membership():
    """Test membership in tuple created from range (CPython test_builtin_sequence_types)"""
    a = tuple(range(10))
    for i in a:
        expect(i in a).to_be(True)
    expect(16 in a).to_be(False)

test("tuple_from_range_membership", test_tuple_from_range_membership)

# === __contains__ = None blocks fallback ===
def test_contains_none_blocks_fallback():
    """Test that __contains__ = None blocks iteration fallback (CPython test_block_fallback)"""
    class ByContains:
        def __contains__(self, other):
            return False

    c = ByContains()
    expect(0 in c).to_be(False)

    class BlockContains(ByContains):
        def __iter__(self):
            while False:
                yield None
        __contains__ = None

    bc = BlockContains()
    # list(bc) should work since __iter__ is defined
    expect(list(bc)).to_be([])
    # But 'in' should raise TypeError because __contains__ = None blocks fallback
    raised = False
    try:
        0 in bc
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("contains_none_blocks_fallback", test_contains_none_blocks_fallback)

# === __getitem__ fallback with IndexError ===
def test_getitem_fallback_membership():
    """Test that __getitem__ fallback works correctly for membership (stops at IndexError)"""
    class SeqContainer:
        def __init__(self, *items):
            self.items = list(items)
        def __getitem__(self, index):
            return self.items[index]

    sc = SeqContainer(10, 20, 30)
    expect(10 in sc).to_be(True)
    expect(20 in sc).to_be(True)
    expect(30 in sc).to_be(True)
    expect(40 in sc).to_be(False)
    expect(10 not in sc).to_be(False)
    expect(40 not in sc).to_be(True)

test("getitem_fallback_membership", test_getitem_fallback_membership)

# === __contains__ with identity check ===
def test_contains_identity():
    """Test that 'in' works with identity (same object) in lists and tuples"""
    class AlwaysNotEqual:
        def __eq__(self, other):
            return False
    obj = AlwaysNotEqual()
    # Even though __eq__ always returns False, identity check should find it
    expect(obj in [obj]).to_be(True)
    expect(obj in (obj,)).to_be(True)
    expect(obj in [1, 2, obj, 3]).to_be(True)

test("contains_identity", test_contains_identity)

# === Membership with boolean/int equivalence ===
def test_contains_bool_int_equivalence():
    """Test membership with True==1 and False==0 equivalence"""
    expect(True in [1]).to_be(True)
    expect(1 in [True]).to_be(True)
    expect(False in [0]).to_be(True)
    expect(0 in [False]).to_be(True)
    expect(True in {1}).to_be(True)
    expect(1 in {True}).to_be(True)

test("contains_bool_int_equivalence", test_contains_bool_int_equivalence)

print("CPython membership tests completed")
