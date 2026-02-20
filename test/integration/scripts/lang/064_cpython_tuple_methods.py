# Test: CPython Tuple Operations Deep Dive
# Adapted from CPython's test_tuple.py - covers tuple methods and operations
# beyond 33_cpython_tuple.py

from test_framework import test, expect

# =============================================================================
# Tuple creation variants
# =============================================================================

def test_tuple_from_list():
    expect(tuple([1, 2, 3])).to_be((1, 2, 3))
    expect(tuple([])).to_be(())

def test_tuple_from_string():
    expect(tuple("hello")).to_be(("h", "e", "l", "l", "o"))
    expect(tuple("")).to_be(())

def test_tuple_from_range():
    expect(tuple(range(5))).to_be((0, 1, 2, 3, 4))
    expect(tuple(range(0))).to_be(())

def test_tuple_from_set():
    # Sets are unordered, so sort to verify contents
    result = sorted(tuple({3, 1, 2}))
    expect(result).to_be([1, 2, 3])

def test_tuple_from_generator():
    result = tuple(x * 2 for x in range(4))
    expect(result).to_be((0, 2, 4, 6))

# =============================================================================
# Tuple immutability
# =============================================================================

def test_immutability_setitem():
    t = (1, 2, 3)
    caught = False
    try:
        t[0] = 99
    except TypeError:
        caught = True
    expect(caught).to_be(True)

def test_immutability_delitem():
    t = (1, 2, 3)
    caught = False
    try:
        del t[0]
    except TypeError:
        caught = True
    expect(caught).to_be(True)

# =============================================================================
# Tuple concatenation and repetition
# =============================================================================

def test_concat_basic():
    expect((1, 2) + (3, 4)).to_be((1, 2, 3, 4))

def test_concat_empty():
    expect(() + (1, 2)).to_be((1, 2))
    expect((1, 2) + ()).to_be((1, 2))
    expect(() + ()).to_be(())

def test_concat_chain():
    result = (1,) + (2,) + (3,) + (4,)
    expect(result).to_be((1, 2, 3, 4))

def test_repetition_basic():
    expect((1, 2) * 3).to_be((1, 2, 1, 2, 1, 2))
    expect(3 * (1, 2)).to_be((1, 2, 1, 2, 1, 2))

def test_repetition_zero():
    expect((1, 2, 3) * 0).to_be(())
    expect(0 * (1, 2, 3)).to_be(())

def test_repetition_negative():
    expect((1, 2) * -5).to_be(())

def test_repetition_one():
    expect((1, 2, 3) * 1).to_be((1, 2, 3))

# =============================================================================
# Tuple comparison (lexicographic)
# =============================================================================

def test_comparison_less_than():
    expect((1, 2, 3) < (1, 2, 4)).to_be(True)
    expect((1, 2, 3) < (1, 3, 0)).to_be(True)
    expect((1,) < (2,)).to_be(True)

def test_comparison_shorter_prefix():
    expect((1, 2) < (1, 2, 3)).to_be(True)
    expect((1, 2, 3) > (1, 2)).to_be(True)

def test_comparison_equal():
    expect((1, 2, 3) == (1, 2, 3)).to_be(True)
    expect(() == ()).to_be(True)

def test_comparison_not_equal():
    expect((1, 2) != (1, 3)).to_be(True)
    expect((1,) != (1, 2)).to_be(True)

def test_comparison_greater():
    expect((2, 0) > (1, 9, 9)).to_be(True)
    expect((1, 2, 3) >= (1, 2, 3)).to_be(True)
    expect((1, 2, 4) >= (1, 2, 3)).to_be(True)

def test_comparison_le_ge():
    expect((1, 2) <= (1, 2)).to_be(True)
    expect((1, 2) <= (1, 3)).to_be(True)
    expect((1, 3) >= (1, 2)).to_be(True)

# =============================================================================
# Tuple as dict key
# =============================================================================

def test_tuple_as_dict_key():
    d = {}
    d[(1, 2)] = "a"
    d[(3, 4)] = "b"
    d[("hello", "world")] = "c"
    expect(d[(1, 2)]).to_be("a")
    expect(d[(3, 4)]).to_be("b")
    expect(d[("hello", "world")]).to_be("c")

def test_tuple_dict_key_lookup():
    d = {(0, 0): "origin", (1, 0): "right", (0, 1): "up"}
    expect(d[(0, 0)]).to_be("origin")
    expect((1, 0) in d).to_be(True)
    expect((2, 2) in d).to_be(False)

# =============================================================================
# Single element tuple (trailing comma)
# =============================================================================

def test_single_element_tuple():
    t = (42,)
    expect(type(t).__name__).to_be("tuple")
    expect(len(t)).to_be(1)
    expect(t[0]).to_be(42)

def test_parenthesized_expression_not_tuple():
    x = (42)
    expect(type(x).__name__).to_be("int")

# =============================================================================
# Tuple methods: count() and index()
# =============================================================================

def test_count_basic():
    t = (1, 2, 3, 2, 1, 2)
    expect(t.count(2)).to_be(3)
    expect(t.count(1)).to_be(2)
    expect(t.count(3)).to_be(1)

def test_count_missing():
    t = (1, 2, 3)
    expect(t.count(99)).to_be(0)

def test_count_empty():
    expect(().count(1)).to_be(0)

def test_count_various_types():
    t = (1, "a", 1, "a", True, None, None)
    expect(t.count("a")).to_be(2)
    expect(t.count(None)).to_be(2)

def test_index_basic():
    t = (10, 20, 30, 40, 50)
    expect(t.index(10)).to_be(0)
    expect(t.index(30)).to_be(2)
    expect(t.index(50)).to_be(4)

def test_index_first_occurrence():
    t = (1, 2, 3, 2, 1)
    expect(t.index(2)).to_be(1)
    expect(t.index(1)).to_be(0)

def test_index_missing_raises():
    t = (1, 2, 3)
    caught = False
    try:
        t.index(99)
    except ValueError:
        caught = True
    expect(caught).to_be(True)

# =============================================================================
# Nested tuples
# =============================================================================

def test_nested_access():
    t = ((1, 2), (3, 4), (5, 6))
    expect(t[0][0]).to_be(1)
    expect(t[1][1]).to_be(4)
    expect(t[2][0]).to_be(5)

def test_deeply_nested():
    t = ((1, (2, (3, 4))),)
    expect(t[0][0]).to_be(1)
    expect(t[0][1][0]).to_be(2)
    expect(t[0][1][1][0]).to_be(3)
    expect(t[0][1][1][1]).to_be(4)

def test_tuple_of_different_containers():
    t = ([1, 2], {"a": 1}, {3, 4})
    expect(t[0]).to_be([1, 2])
    expect(t[1]).to_be({"a": 1})
    # Set comparison
    expect(3 in t[2]).to_be(True)

# =============================================================================
# Tuple equality and identity
# =============================================================================

def test_equality_same_values():
    t1 = (1, 2, 3)
    t2 = (1, 2, 3)
    expect(t1 == t2).to_be(True)

def test_equality_different_values():
    expect((1, 2) == (1, 3)).to_be(False)
    expect((1, 2) == (1, 2, 3)).to_be(False)

def test_equality_different_types():
    expect((1, 2, 3) == [1, 2, 3]).to_be(False)

# =============================================================================
# Empty tuple
# =============================================================================

def test_empty_tuple():
    t = ()
    expect(len(t)).to_be(0)
    expect(bool(t)).to_be(False)
    expect(list(t)).to_be([])

def test_empty_tuple_from_constructor():
    t = tuple()
    expect(t).to_be(())
    expect(len(t)).to_be(0)

# =============================================================================
# Tuple in membership testing
# =============================================================================

def test_membership_basic():
    t = (1, 2, 3, "hello", None)
    expect(1 in t).to_be(True)
    expect("hello" in t).to_be(True)
    expect(None in t).to_be(True)
    expect(99 not in t).to_be(True)

def test_membership_nested():
    t = ((1, 2), (3, 4))
    expect((1, 2) in t).to_be(True)
    expect((3, 4) in t).to_be(True)
    expect((1, 3) in t).to_be(False)

# =============================================================================
# Tuple sorting (as elements in a list)
# =============================================================================

def test_sort_tuples_in_list():
    data = [(3, "c"), (1, "a"), (2, "b")]
    result = sorted(data)
    expect(result).to_be([(1, "a"), (2, "b"), (3, "c")])

def test_sort_tuples_by_second_element():
    # Tuples sort lexicographically, first element first
    data = [(1, "z"), (1, "a"), (1, "m")]
    result = sorted(data)
    expect(result).to_be([(1, "a"), (1, "m"), (1, "z")])

# =============================================================================
# Tuple with mixed types
# =============================================================================

def test_mixed_types_in_tuple():
    t = (1, "two", 3.0, True, None, (5, 6))
    expect(len(t)).to_be(6)
    expect(t[0]).to_be(1)
    expect(t[1]).to_be("two")
    expect(t[4]).to_be(None)
    expect(t[5]).to_be((5, 6))

# =============================================================================
# Tuple conversion
# =============================================================================

def test_list_to_tuple():
    expect(tuple([1, 2, 3])).to_be((1, 2, 3))

def test_tuple_to_list():
    expect(list((1, 2, 3))).to_be([1, 2, 3])

def test_string_to_tuple():
    expect(tuple("abc")).to_be(("a", "b", "c"))

# =============================================================================
# Tuple hashing (for dict keys and set elements)
# =============================================================================

def test_tuple_hash_equal():
    t1 = (1, 2, 3)
    t2 = (1, 2, 3)
    expect(hash(t1) == hash(t2)).to_be(True)

def test_tuple_in_set():
    s = {(1, 2), (3, 4), (5, 6)}
    expect((1, 2) in s).to_be(True)
    expect((7, 8) in s).to_be(False)
    expect(len(s)).to_be(3)

def test_tuple_set_dedup():
    s = {(1, 2), (1, 2), (3, 4)}
    expect(len(s)).to_be(2)

# =============================================================================
# Tuple slicing
# =============================================================================

def test_slicing_basic():
    t = (0, 1, 2, 3, 4, 5)
    expect(t[1:4]).to_be((1, 2, 3))
    expect(t[:3]).to_be((0, 1, 2))
    expect(t[3:]).to_be((3, 4, 5))

def test_slicing_step():
    t = (0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
    expect(t[::2]).to_be((0, 2, 4, 6, 8))
    expect(t[1::3]).to_be((1, 4, 7))

def test_slicing_reverse():
    t = (1, 2, 3, 4, 5)
    expect(t[::-1]).to_be((5, 4, 3, 2, 1))

def test_slicing_negative():
    t = (0, 1, 2, 3, 4)
    expect(t[-3:]).to_be((2, 3, 4))
    expect(t[:-2]).to_be((0, 1, 2))

# =============================================================================
# Run all tests
# =============================================================================

test("tuple_from_list", test_tuple_from_list)
test("tuple_from_string", test_tuple_from_string)
test("tuple_from_range", test_tuple_from_range)
test("tuple_from_set", test_tuple_from_set)
test("tuple_from_generator", test_tuple_from_generator)
test("immutability_setitem", test_immutability_setitem)
test("immutability_delitem", test_immutability_delitem)
test("concat_basic", test_concat_basic)
test("concat_empty", test_concat_empty)
test("concat_chain", test_concat_chain)
test("repetition_basic", test_repetition_basic)
test("repetition_zero", test_repetition_zero)
test("repetition_negative", test_repetition_negative)
test("repetition_one", test_repetition_one)
test("comparison_less_than", test_comparison_less_than)
test("comparison_shorter_prefix", test_comparison_shorter_prefix)
test("comparison_equal", test_comparison_equal)
test("comparison_not_equal", test_comparison_not_equal)
test("comparison_greater", test_comparison_greater)
test("comparison_le_ge", test_comparison_le_ge)
test("tuple_as_dict_key", test_tuple_as_dict_key)
test("tuple_dict_key_lookup", test_tuple_dict_key_lookup)
test("single_element_tuple", test_single_element_tuple)
test("parenthesized_expression_not_tuple", test_parenthesized_expression_not_tuple)
test("count_basic", test_count_basic)
test("count_missing", test_count_missing)
test("count_empty", test_count_empty)
test("count_various_types", test_count_various_types)
test("index_basic", test_index_basic)
test("index_first_occurrence", test_index_first_occurrence)
test("index_missing_raises", test_index_missing_raises)
test("nested_access", test_nested_access)
test("deeply_nested", test_deeply_nested)
test("tuple_of_different_containers", test_tuple_of_different_containers)
test("equality_same_values", test_equality_same_values)
test("equality_different_values", test_equality_different_values)
test("equality_different_types", test_equality_different_types)
test("empty_tuple", test_empty_tuple)
test("empty_tuple_from_constructor", test_empty_tuple_from_constructor)
test("membership_basic", test_membership_basic)
test("membership_nested", test_membership_nested)
test("sort_tuples_in_list", test_sort_tuples_in_list)
test("sort_tuples_by_second_element", test_sort_tuples_by_second_element)
test("mixed_types_in_tuple", test_mixed_types_in_tuple)
test("list_to_tuple", test_list_to_tuple)
test("tuple_to_list", test_tuple_to_list)
test("string_to_tuple", test_string_to_tuple)
test("tuple_hash_equal", test_tuple_hash_equal)
test("tuple_in_set", test_tuple_in_set)
test("tuple_set_dedup", test_tuple_set_dedup)
test("slicing_basic", test_slicing_basic)
test("slicing_step", test_slicing_step)
test("slicing_reverse", test_slicing_reverse)
test("slicing_negative", test_slicing_negative)

print("CPython tuple methods tests completed")
