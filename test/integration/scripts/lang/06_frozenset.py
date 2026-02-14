# Test: Frozenset Operations
# Tests frozenset type - immutable, hashable set

from test_framework import test, expect

def test_frozenset_creation_empty():
    fs = frozenset()
    expect(len(fs)).to_be(0)

def test_frozenset_creation_from_list():
    fs = frozenset([1, 2, 3])
    expect(len(fs)).to_be(3)
    expect(1 in fs).to_be(True)
    expect(4 in fs).to_be(False)

def test_frozenset_creation_from_string():
    fs = frozenset("abc")
    expect(len(fs)).to_be(3)
    expect("a" in fs).to_be(True)

def test_frozenset_membership():
    fs = frozenset([1, 2, 3, 4, 5])
    expect(3 in fs).to_be(True)
    expect(6 in fs).to_be(False)
    expect(6 not in fs).to_be(True)

def test_frozenset_len():
    expect(len(frozenset())).to_be(0)
    expect(len(frozenset([1, 2, 3]))).to_be(3)
    expect(len(frozenset("hello"))).to_be(4)  # h, e, l, o (l appears twice)

def test_frozenset_equality():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([3, 2, 1])
    expect(fs1 == fs2).to_be(True)
    expect(frozenset([1, 2]) == frozenset([1, 2, 3])).to_be(False)

def test_frozenset_set_equality():
    # Frozenset and set with same elements should be equal
    fs = frozenset([1, 2, 3])
    s = {1, 2, 3}
    expect(fs == s).to_be(True)
    expect(s == fs).to_be(True)

def test_frozenset_hashable():
    # Frozenset can be used as dict key
    d = {frozenset([1, 2]): "value1", frozenset([3, 4]): "value2"}
    expect(d[frozenset([1, 2])]).to_be("value1")
    expect(d[frozenset([3, 4])]).to_be("value2")

def test_frozenset_in_set():
    # Frozenset can be a member of a set
    s = {frozenset([1, 2]), frozenset([3, 4])}
    expect(len(s)).to_be(2)
    expect(frozenset([1, 2]) in s).to_be(True)

def test_frozenset_union():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([3, 4, 5])
    result = fs1.union(fs2)
    expect(len(result)).to_be(5)
    expect(1 in result).to_be(True)
    expect(5 in result).to_be(True)

def test_frozenset_union_operator():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([3, 4, 5])
    result = fs1 | fs2
    expect(len(result)).to_be(5)
    # Verify it's a frozenset by using it as dict key (only hashable types work)
    d = {result: "test"}
    expect(d[result]).to_be("test")

def test_frozenset_intersection():
    fs1 = frozenset([1, 2, 3, 4])
    fs2 = frozenset([3, 4, 5, 6])
    result = fs1.intersection(fs2)
    expect(len(result)).to_be(2)
    expect(3 in result).to_be(True)
    expect(4 in result).to_be(True)
    expect(1 in result).to_be(False)

def test_frozenset_intersection_operator():
    fs1 = frozenset([1, 2, 3, 4])
    fs2 = frozenset([3, 4, 5, 6])
    result = fs1 & fs2
    expect(len(result)).to_be(2)
    # Verify it's a frozenset by using it as dict key
    d = {result: "test"}
    expect(d[result]).to_be("test")

def test_frozenset_difference():
    fs1 = frozenset([1, 2, 3, 4])
    fs2 = frozenset([3, 4, 5])
    result = fs1.difference(fs2)
    expect(len(result)).to_be(2)
    expect(1 in result).to_be(True)
    expect(2 in result).to_be(True)
    expect(3 in result).to_be(False)

def test_frozenset_difference_operator():
    fs1 = frozenset([1, 2, 3, 4])
    fs2 = frozenset([3, 4, 5])
    result = fs1 - fs2
    expect(len(result)).to_be(2)
    # Verify it's a frozenset by using it as dict key
    d = {result: "test"}
    expect(d[result]).to_be("test")

def test_frozenset_symmetric_difference():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([2, 3, 4])
    result = fs1.symmetric_difference(fs2)
    expect(len(result)).to_be(2)
    expect(1 in result).to_be(True)
    expect(4 in result).to_be(True)
    expect(2 in result).to_be(False)

def test_frozenset_symmetric_difference_operator():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([2, 3, 4])
    result = fs1 ^ fs2
    expect(len(result)).to_be(2)
    # Verify it's a frozenset by using it as dict key
    d = {result: "test"}
    expect(d[result]).to_be("test")

def test_frozenset_issubset():
    fs1 = frozenset([1, 2])
    fs2 = frozenset([1, 2, 3, 4])
    expect(fs1.issubset(fs2)).to_be(True)
    expect(fs2.issubset(fs1)).to_be(False)
    expect(fs1.issubset(fs1)).to_be(True)

def test_frozenset_issuperset():
    fs1 = frozenset([1, 2, 3, 4])
    fs2 = frozenset([1, 2])
    expect(fs1.issuperset(fs2)).to_be(True)
    expect(fs2.issuperset(fs1)).to_be(False)
    expect(fs1.issuperset(fs1)).to_be(True)

def test_frozenset_isdisjoint():
    fs1 = frozenset([1, 2, 3])
    fs2 = frozenset([4, 5, 6])
    fs3 = frozenset([3, 4, 5])
    expect(fs1.isdisjoint(fs2)).to_be(True)
    expect(fs1.isdisjoint(fs3)).to_be(False)

def test_frozenset_copy():
    fs1 = frozenset([1, 2, 3])
    fs2 = fs1.copy()
    expect(fs1 == fs2).to_be(True)

def test_frozenset_iteration():
    fs = frozenset([1, 2, 3])
    total = 0
    for item in fs:
        total = total + item
    expect(total).to_be(6)

def test_frozenset_bool():
    expect(bool(frozenset())).to_be(False)
    expect(bool(frozenset([1]))).to_be(True)

def test_frozenset_with_set_operator():
    # Mixed set/frozenset operations return set
    fs = frozenset([1, 2, 3])
    s = {3, 4, 5}
    result = fs | s
    expect(len(result)).to_be(5)

# Register all tests
test("frozenset_creation_empty", test_frozenset_creation_empty)
test("frozenset_creation_from_list", test_frozenset_creation_from_list)
test("frozenset_creation_from_string", test_frozenset_creation_from_string)
test("frozenset_membership", test_frozenset_membership)
test("frozenset_len", test_frozenset_len)
test("frozenset_equality", test_frozenset_equality)
test("frozenset_set_equality", test_frozenset_set_equality)
test("frozenset_hashable", test_frozenset_hashable)
test("frozenset_in_set", test_frozenset_in_set)
test("frozenset_union", test_frozenset_union)
test("frozenset_union_operator", test_frozenset_union_operator)
test("frozenset_intersection", test_frozenset_intersection)
test("frozenset_intersection_operator", test_frozenset_intersection_operator)
test("frozenset_difference", test_frozenset_difference)
test("frozenset_difference_operator", test_frozenset_difference_operator)
test("frozenset_symmetric_difference", test_frozenset_symmetric_difference)
test("frozenset_symmetric_difference_operator", test_frozenset_symmetric_difference_operator)
test("frozenset_issubset", test_frozenset_issubset)
test("frozenset_issuperset", test_frozenset_issuperset)
test("frozenset_isdisjoint", test_frozenset_isdisjoint)
test("frozenset_copy", test_frozenset_copy)
test("frozenset_iteration", test_frozenset_iteration)
test("frozenset_bool", test_frozenset_bool)
test("frozenset_with_set_operator", test_frozenset_with_set_operator)

print("Frozenset tests completed")
