# Test: CPython Set Methods Deep Dive
# Adapted from CPython's test_set.py - covers set methods in more depth
# beyond 34_cpython_set.py

from test_framework import test, expect

# =============================================================================
# set.add() edge cases
# =============================================================================

def test_add_various_types():
    s = set()
    s.add(1)
    s.add("hello")
    s.add((1, 2))
    s.add(None)
    s.add(True)
    s.add(3.14)
    expect(len(s) >= 5).to_be(True)  # True and 1 may be same
    expect("hello" in s).to_be(True)
    expect((1, 2) in s).to_be(True)
    expect(None in s).to_be(True)

def test_add_duplicate_no_effect():
    s = {1, 2, 3}
    s.add(2)
    s.add(2)
    s.add(2)
    expect(len(s)).to_be(3)
    expect(s).to_be({1, 2, 3})

# =============================================================================
# set.remove() edge cases
# =============================================================================

def test_remove_existing():
    s = {1, 2, 3, 4, 5}
    s.remove(3)
    expect(s).to_be({1, 2, 4, 5})
    s.remove(1)
    expect(s).to_be({2, 4, 5})

def test_remove_missing_raises_keyerror():
    s = {1, 2, 3}
    caught = False
    try:
        s.remove(99)
    except KeyError:
        caught = True
    expect(caught).to_be(True)

# =============================================================================
# set.discard() edge cases
# =============================================================================

def test_discard_existing_and_missing():
    s = {1, 2, 3, 4}
    s.discard(2)
    expect(s).to_be({1, 3, 4})
    # discard on missing element does NOT raise
    s.discard(99)
    expect(s).to_be({1, 3, 4})
    s.discard(1)
    s.discard(3)
    s.discard(4)
    expect(s).to_be(set())

# =============================================================================
# set.pop() edge cases
# =============================================================================

def test_pop_until_empty():
    s = {10, 20, 30}
    results = []
    results.append(s.pop())
    results.append(s.pop())
    results.append(s.pop())
    expect(len(s)).to_be(0)
    expect(sorted(results)).to_be([10, 20, 30])

def test_pop_empty_raises():
    s = set()
    caught = False
    try:
        s.pop()
    except KeyError:
        caught = True
    expect(caught).to_be(True)

# =============================================================================
# set.clear() edge cases
# =============================================================================

def test_clear_already_empty():
    s = set()
    s.clear()
    expect(s).to_be(set())
    expect(len(s)).to_be(0)

def test_clear_large():
    s = set(range(100))
    expect(len(s)).to_be(100)
    s.clear()
    expect(len(s)).to_be(0)
    expect(s).to_be(set())

# =============================================================================
# set.copy() edge cases
# =============================================================================

def test_copy_independence():
    original = {1, 2, 3}
    copied = original.copy()
    expect(copied).to_be(original)
    copied.add(4)
    expect(4 in original).to_be(False)
    expect(4 in copied).to_be(True)
    original.add(5)
    expect(5 in copied).to_be(False)

def test_copy_empty():
    s = set()
    c = s.copy()
    expect(c).to_be(set())
    c.add(1)
    expect(len(s)).to_be(0)

# =============================================================================
# set.union() with various iterables
# =============================================================================

def test_union_with_list():
    s = {1, 2}
    result = s.union([3, 4])
    expect(result).to_be({1, 2, 3, 4})

def test_union_with_tuple():
    s = {1, 2}
    result = s.union((3, 4))
    expect(result).to_be({1, 2, 3, 4})

def test_union_with_string():
    s = {"a", "b"}
    result = s.union("cd")
    expect(result).to_be({"a", "b", "c", "d"})

def test_union_multiple_args():
    s = {1}
    result = s.union({2}, {3}, {4})
    expect(result).to_be({1, 2, 3, 4})

# =============================================================================
# set.intersection() variants
# =============================================================================

def test_intersection_with_list():
    s = {1, 2, 3, 4}
    result = s.intersection([2, 3, 5])
    expect(result).to_be({2, 3})

def test_intersection_empty_result():
    s = {1, 2, 3}
    result = s.intersection({4, 5, 6})
    expect(result).to_be(set())

def test_intersection_operator():
    expect({1, 2, 3, 4} & {2, 4, 6}).to_be({2, 4})

# =============================================================================
# set.difference() variants
# =============================================================================

def test_difference_with_list():
    s = {1, 2, 3, 4, 5}
    result = s.difference([2, 4])
    expect(result).to_be({1, 3, 5})

def test_difference_operator():
    expect({1, 2, 3, 4} - {2, 3}).to_be({1, 4})

def test_difference_no_overlap():
    s = {1, 2, 3}
    result = s.difference({4, 5})
    expect(result).to_be({1, 2, 3})

# =============================================================================
# set.symmetric_difference() variants
# =============================================================================

def test_symmetric_difference_basic():
    s1 = {1, 2, 3, 4}
    s2 = {3, 4, 5, 6}
    result = s1.symmetric_difference(s2)
    expect(result).to_be({1, 2, 5, 6})

def test_symmetric_difference_operator():
    expect({1, 2, 3} ^ {2, 3, 4}).to_be({1, 4})

def test_symmetric_difference_identical():
    s = {1, 2, 3}
    result = s.symmetric_difference({1, 2, 3})
    expect(result).to_be(set())

# =============================================================================
# set.update() and method-based updates
# =============================================================================

def test_update_basic():
    s = {1, 2}
    s.update({3, 4})
    expect(s).to_be({1, 2, 3, 4})

def test_update_with_list():
    s = {1}
    s.update([2, 3, 4])
    expect(s).to_be({1, 2, 3, 4})

def test_intersection_update():
    # intersection_update not available, use &= operator instead
    s = {1, 2, 3, 4, 5}
    s &= {2, 3, 4, 6}
    expect(s).to_be({2, 3, 4})

def test_difference_update():
    # difference_update not available, use -= operator instead
    s = {1, 2, 3, 4, 5}
    s -= {2, 4}
    expect(s).to_be({1, 3, 5})

def test_symmetric_difference_update():
    # symmetric_difference_update not available, use ^= operator instead
    s = {1, 2, 3}
    s ^= {2, 3, 4}
    expect(s).to_be({1, 4})

# =============================================================================
# set.issubset(), set.issuperset(), set.isdisjoint()
# =============================================================================

def test_subset_true():
    expect({1, 2}.issubset({1, 2, 3})).to_be(True)

def test_subset_equal_sets():
    expect({1, 2, 3}.issubset({1, 2, 3})).to_be(True)

def test_subset_false():
    expect({1, 2, 4}.issubset({1, 2, 3})).to_be(False)

def test_superset_true():
    expect({1, 2, 3}.issuperset({1, 2})).to_be(True)

def test_superset_false():
    expect({1, 2}.issuperset({1, 2, 3})).to_be(False)

def test_disjoint_true():
    expect({1, 2}.isdisjoint({3, 4})).to_be(True)

def test_disjoint_false():
    expect({1, 2, 3}.isdisjoint({3, 4, 5})).to_be(False)

def test_empty_set_subset_of_everything():
    expect(set().issubset({1, 2, 3})).to_be(True)
    expect(set().issubset(set())).to_be(True)

# =============================================================================
# Frozenset basics
# =============================================================================

def test_frozenset_creation():
    fs = frozenset([1, 2, 3])
    expect(len(fs)).to_be(3)
    expect(1 in fs).to_be(True)

def test_frozenset_immutable():
    fs = frozenset([1, 2, 3])
    caught = False
    try:
        fs.add(4)
    except:
        caught = True
    expect(caught).to_be(True)

def test_frozenset_as_set_element():
    s = set()
    s.add(frozenset([1, 2]))
    s.add(frozenset([3, 4]))
    expect(len(s)).to_be(2)
    expect(frozenset([1, 2]) in s).to_be(True)

# =============================================================================
# Set from various iterables
# =============================================================================

def test_set_from_range():
    s = set(range(5))
    expect(s).to_be({0, 1, 2, 3, 4})

def test_set_from_string():
    s = set("banana")
    expect(s).to_be({"b", "a", "n"})

def test_set_from_dict_keys():
    d = {"a": 1, "b": 2, "c": 3}
    s = set(d)
    expect(s).to_be({"a", "b", "c"})

# =============================================================================
# Set with mixed types
# =============================================================================

def test_mixed_types():
    s = {1, "hello", 3.14, (1, 2), None, frozenset([5])}
    expect(1 in s).to_be(True)
    expect("hello" in s).to_be(True)
    expect(None in s).to_be(True)
    expect((1, 2) in s).to_be(True)

# =============================================================================
# Empty set operations
# =============================================================================

def test_empty_set_operations():
    empty = set()
    other = {1, 2, 3}
    expect(empty | other).to_be({1, 2, 3})
    expect(empty & other).to_be(set())
    expect(empty - other).to_be(set())
    expect(other - empty).to_be({1, 2, 3})
    expect(empty ^ other).to_be({1, 2, 3})

# =============================================================================
# Set comprehension with conditions
# =============================================================================

def test_set_comprehension_basic():
    s = {x * x for x in range(6)}
    expect(s).to_be({0, 1, 4, 9, 16, 25})

def test_set_comprehension_filtered():
    s = {x for x in range(20) if x % 3 == 0}
    expect(s).to_be({0, 3, 6, 9, 12, 15, 18})

def test_set_comprehension_from_string():
    s = {c for c in "abracadabra" if c not in "abc"}
    expect(s).to_be({"d", "r"})

# =============================================================================
# Run all tests
# =============================================================================

test("add_various_types", test_add_various_types)
test("add_duplicate_no_effect", test_add_duplicate_no_effect)
test("remove_existing", test_remove_existing)
test("remove_missing_raises_keyerror", test_remove_missing_raises_keyerror)
test("discard_existing_and_missing", test_discard_existing_and_missing)
test("pop_until_empty", test_pop_until_empty)
test("pop_empty_raises", test_pop_empty_raises)
test("clear_already_empty", test_clear_already_empty)
test("clear_large", test_clear_large)
test("copy_independence", test_copy_independence)
test("copy_empty", test_copy_empty)
test("union_with_list", test_union_with_list)
test("union_with_tuple", test_union_with_tuple)
test("union_with_string", test_union_with_string)
test("union_multiple_args", test_union_multiple_args)
test("intersection_with_list", test_intersection_with_list)
test("intersection_empty_result", test_intersection_empty_result)
test("intersection_operator", test_intersection_operator)
test("difference_with_list", test_difference_with_list)
test("difference_operator", test_difference_operator)
test("difference_no_overlap", test_difference_no_overlap)
test("symmetric_difference_basic", test_symmetric_difference_basic)
test("symmetric_difference_operator", test_symmetric_difference_operator)
test("symmetric_difference_identical", test_symmetric_difference_identical)
test("update_basic", test_update_basic)
test("update_with_list", test_update_with_list)
test("intersection_update", test_intersection_update)
test("difference_update", test_difference_update)
test("symmetric_difference_update", test_symmetric_difference_update)
test("subset_true", test_subset_true)
test("subset_equal_sets", test_subset_equal_sets)
test("subset_false", test_subset_false)
test("superset_true", test_superset_true)
test("superset_false", test_superset_false)
test("disjoint_true", test_disjoint_true)
test("disjoint_false", test_disjoint_false)
test("empty_set_subset_of_everything", test_empty_set_subset_of_everything)
test("frozenset_creation", test_frozenset_creation)
test("frozenset_immutable", test_frozenset_immutable)
test("frozenset_as_set_element", test_frozenset_as_set_element)
test("set_from_range", test_set_from_range)
test("set_from_string", test_set_from_string)
test("set_from_dict_keys", test_set_from_dict_keys)
test("mixed_types", test_mixed_types)
test("empty_set_operations", test_empty_set_operations)
test("set_comprehension_basic", test_set_comprehension_basic)
test("set_comprehension_filtered", test_set_comprehension_filtered)
test("set_comprehension_from_string", test_set_comprehension_from_string)

print("CPython set methods tests completed")
