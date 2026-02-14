# Test: CPython Set Edge Cases
# Adapted from CPython's test_set.py - covers edge cases beyond 26_frozenset.py

from test_framework import test, expect

def test_set_construction():
    expect(set()).to_be(set())
    expect(set([1, 2, 2, 3])).to_be({1, 2, 3})
    expect(len(set("hello"))).to_be(4)  # h, e, l, o

def test_set_add_discard():
    s = {1, 2, 3}
    s.add(4)
    expect(4 in s).to_be(True)
    s.add(3)  # Already exists, no effect
    expect(len(s)).to_be(4)
    s.discard(2)
    expect(2 in s).to_be(False)
    s.discard(99)  # Missing, no error
    expect(len(s)).to_be(3)

def test_set_remove():
    s = {1, 2, 3}
    s.remove(2)
    expect(2 in s).to_be(False)
    try:
        s.remove(99)
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_set_pop():
    s = {1}
    item = s.pop()
    expect(item).to_be(1)
    expect(len(s)).to_be(0)
    try:
        s.pop()
        expect("no error").to_be("KeyError")
    except KeyError:
        expect(True).to_be(True)

def test_set_clear():
    s = {1, 2, 3}
    s.clear()
    expect(s).to_be(set())
    expect(len(s)).to_be(0)

def test_set_copy():
    s = {1, 2, 3}
    c = s.copy()
    expect(c).to_be({1, 2, 3})
    c.add(4)
    expect(4 in s).to_be(False)  # Independent

def test_set_union():
    s1 = {1, 2, 3}
    s2 = {3, 4, 5}
    expect(s1 | s2).to_be({1, 2, 3, 4, 5})
    expect(s1.union(s2)).to_be({1, 2, 3, 4, 5})
    # Union with multiple args
    expect(s1.union(s2, {6})).to_be({1, 2, 3, 4, 5, 6})

def test_set_intersection():
    s1 = {1, 2, 3, 4}
    s2 = {3, 4, 5, 6}
    expect(s1 & s2).to_be({3, 4})
    expect(s1.intersection(s2)).to_be({3, 4})

def test_set_difference():
    s1 = {1, 2, 3, 4}
    s2 = {3, 4, 5}
    expect(s1 - s2).to_be({1, 2})
    expect(s1.difference(s2)).to_be({1, 2})

def test_set_symmetric_difference():
    s1 = {1, 2, 3}
    s2 = {2, 3, 4}
    expect(s1 ^ s2).to_be({1, 4})
    expect(s1.symmetric_difference(s2)).to_be({1, 4})

def test_set_subset_superset():
    s1 = {1, 2}
    s2 = {1, 2, 3, 4}
    expect(s1 <= s2).to_be(True)
    expect(s1.issubset(s2)).to_be(True)
    expect(s2 >= s1).to_be(True)
    expect(s2.issuperset(s1)).to_be(True)
    # Proper subset
    expect(s1 < s2).to_be(True)
    expect(s1 < s1).to_be(False)
    expect(s1 <= s1).to_be(True)

def test_set_disjoint():
    expect({1, 2}.isdisjoint({3, 4})).to_be(True)
    expect({1, 2}.isdisjoint({2, 3})).to_be(False)

def test_set_update_ops():
    s = {1, 2, 3}
    s |= {4, 5}
    expect(s).to_be({1, 2, 3, 4, 5})
    s &= {1, 2, 4}
    expect(s).to_be({1, 2, 4})
    s -= {1}
    expect(s).to_be({2, 4})
    s ^= {2, 3}
    expect(s).to_be({3, 4})

def test_set_contains():
    s = {1, 2, 3}
    expect(1 in s).to_be(True)
    expect(4 not in s).to_be(True)

def test_set_len():
    expect(len(set())).to_be(0)
    expect(len({1, 2, 3})).to_be(3)

def test_set_bool():
    expect(bool(set())).to_be(False)
    expect(bool({1})).to_be(True)

def test_set_iteration():
    s = {3, 1, 2}
    result = sorted(s)
    expect(result).to_be([1, 2, 3])

def test_set_unhashable():
    try:
        {[1, 2]}
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_set_frozenset_interop():
    s = {1, 2}
    fs = frozenset({3, 4})
    result = s | fs
    expect(result).to_be({1, 2, 3, 4})
    # Frozenset as set element (hashable)
    s2 = {frozenset({1, 2}), frozenset({3, 4})}
    expect(len(s2)).to_be(2)
    expect(frozenset({1, 2}) in s2).to_be(True)

def test_set_comparison():
    expect({1, 2} == {2, 1}).to_be(True)
    expect({1, 2} != {1, 3}).to_be(True)
    expect({1, 2, 3} == {1, 2, 3}).to_be(True)

# Register all tests
test("set_construction", test_set_construction)
test("set_add_discard", test_set_add_discard)
test("set_remove", test_set_remove)
test("set_pop", test_set_pop)
test("set_clear", test_set_clear)
test("set_copy", test_set_copy)
test("set_union", test_set_union)
test("set_intersection", test_set_intersection)
test("set_difference", test_set_difference)
test("set_symmetric_difference", test_set_symmetric_difference)
test("set_subset_superset", test_set_subset_superset)
test("set_disjoint", test_set_disjoint)
test("set_update_ops", test_set_update_ops)
test("set_contains", test_set_contains)
test("set_len", test_set_len)
test("set_bool", test_set_bool)
test("set_iteration", test_set_iteration)
test("set_unhashable", test_set_unhashable)
test("set_frozenset_interop", test_set_frozenset_interop)
test("set_comparison", test_set_comparison)

print("CPython set tests completed")
