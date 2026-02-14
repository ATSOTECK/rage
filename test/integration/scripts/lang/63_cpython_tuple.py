# Test: CPython Tuple Edge Cases
# Adapted from CPython's test_tuple.py - covers edge cases beyond 01_data_types.py

from test_framework import test, expect

def test_tuple_construction():
    expect(tuple()).to_be(())
    expect(tuple([1, 2, 3])).to_be((1, 2, 3))
    expect(tuple("abc")).to_be(("a", "b", "c"))
    expect(tuple(range(5))).to_be((0, 1, 2, 3, 4))

def test_tuple_single_element():
    expect((1,)).to_be((1,))
    expect(len((1,))).to_be(1)
    # (1) is just 1, not a tuple
    expect(type((1,)).__name__).to_be("tuple")

def test_tuple_indexing():
    t = (10, 20, 30, 40, 50)
    expect(t[0]).to_be(10)
    expect(t[-1]).to_be(50)
    expect(t[2]).to_be(30)
    try:
        t[99]
        expect("no error").to_be("IndexError")
    except IndexError:
        expect(True).to_be(True)

def test_tuple_slicing():
    t = (0, 1, 2, 3, 4, 5)
    expect(t[1:3]).to_be((1, 2))
    expect(t[::2]).to_be((0, 2, 4))
    expect(t[::-1]).to_be((5, 4, 3, 2, 1, 0))
    expect(t[-2:]).to_be((4, 5))

def test_tuple_immutability():
    t = (1, 2, 3)
    try:
        t[0] = 99
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_tuple_concat():
    expect((1, 2) + (3, 4)).to_be((1, 2, 3, 4))
    expect(() + (1,)).to_be((1,))
    expect((1,) + ()).to_be((1,))

def test_tuple_repeat():
    expect((1, 2) * 3).to_be((1, 2, 1, 2, 1, 2))
    expect(3 * (1, 2)).to_be((1, 2, 1, 2, 1, 2))
    expect((1,) * 0).to_be(())
    expect((1,) * -1).to_be(())

def test_tuple_contains():
    t = (1, 2, 3, "hello")
    expect(1 in t).to_be(True)
    expect(4 not in t).to_be(True)
    expect("hello" in t).to_be(True)

def test_tuple_count():
    t = (1, 2, 1, 3, 1)
    expect(t.count(1)).to_be(3)
    expect(t.count(2)).to_be(1)
    expect(t.count(99)).to_be(0)

def test_tuple_index():
    t = (1, 2, 3, 4, 5)
    expect(t.index(3)).to_be(2)
    expect(t.index(1)).to_be(0)
    try:
        t.index(99)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_tuple_comparison():
    expect((1, 2) < (1, 3)).to_be(True)
    expect((1,) < (1, 2)).to_be(True)
    expect((1, 2) == (1, 2)).to_be(True)
    expect((1, 2) != (1, 3)).to_be(True)
    expect((2,) > (1, 2, 3)).to_be(True)
    expect(() < (1,)).to_be(True)

def test_tuple_bool():
    expect(bool(())).to_be(False)
    expect(bool((0,))).to_be(True)
    expect(bool((None,))).to_be(True)

def test_tuple_unpacking():
    a, b, c = (1, 2, 3)
    expect(a).to_be(1)
    expect(b).to_be(2)
    expect(c).to_be(3)
    # Starred unpacking
    a, *b = (1, 2, 3, 4)
    expect(a).to_be(1)
    expect(b).to_be([2, 3, 4])
    *a, b = (1, 2, 3, 4)
    expect(a).to_be([1, 2, 3])
    expect(b).to_be(4)

def test_tuple_nested():
    t = ((1, 2), (3, 4))
    expect(t[0][1]).to_be(2)
    expect(t[1][0]).to_be(3)
    expect(len(t)).to_be(2)

def test_tuple_hash():
    # Tuples of equal value should have the same hash
    expect(hash((1, 2)) == hash((1, 2))).to_be(True)
    # Tuples can be used as dict keys
    d = {(1, 2): "a", (3, 4): "b"}
    expect(d[(1, 2)]).to_be("a")
    expect(d[(3, 4)]).to_be("b")

# Register all tests
test("tuple_construction", test_tuple_construction)
test("tuple_single_element", test_tuple_single_element)
test("tuple_indexing", test_tuple_indexing)
test("tuple_slicing", test_tuple_slicing)
test("tuple_immutability", test_tuple_immutability)
test("tuple_concat", test_tuple_concat)
test("tuple_repeat", test_tuple_repeat)
test("tuple_contains", test_tuple_contains)
test("tuple_count", test_tuple_count)
test("tuple_index", test_tuple_index)
test("tuple_comparison", test_tuple_comparison)
test("tuple_bool", test_tuple_bool)
test("tuple_unpacking", test_tuple_unpacking)
test("tuple_nested", test_tuple_nested)
test("tuple_hash", test_tuple_hash)

print("CPython tuple tests completed")
