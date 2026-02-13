# Test: Extended Unpacking
# Tests starred/extended unpacking in assignments, loops, and edge cases

from test_framework import test, expect

def test_star_in_middle():
    a, *rest, b = [1, 2, 3, 4]
    expect(a).to_be(1)
    expect(rest).to_be([2, 3])
    expect(b).to_be(4)

def test_star_at_end():
    a, *rest = [1, 2, 3]
    expect(a).to_be(1)
    expect(rest).to_be([2, 3])

def test_star_at_start():
    *rest, a = [1, 2, 3]
    expect(rest).to_be([1, 2])
    expect(a).to_be(3)

def test_multiple_after_star():
    a, *rest, b, c = [1, 2, 3, 4, 5]
    expect(a).to_be(1)
    expect(rest).to_be([2, 3])
    expect(b).to_be(4)
    expect(c).to_be(5)

def test_multiple_before_star():
    a, b, *rest, c = [1, 2, 3, 4, 5]
    expect(a).to_be(1)
    expect(b).to_be(2)
    expect(rest).to_be([3, 4])
    expect(c).to_be(5)

def test_star_gets_empty_list():
    a, *rest, b = [1, 2]
    expect(a).to_be(1)
    expect(rest).to_be([])
    expect(b).to_be(2)

def test_star_from_tuple():
    a, *rest, b = (10, 20, 30, 40)
    expect(a).to_be(10)
    expect(rest).to_be([20, 30])
    expect(b).to_be(40)

def test_star_from_string():
    a, *rest, b = "hello"
    expect(a).to_be("h")
    expect(rest).to_be(["e", "l", "l"])
    expect(b).to_be("o")

def test_star_from_range():
    a, *rest, b = range(1, 6)
    expect(a).to_be(1)
    expect(rest).to_be([2, 3, 4])
    expect(b).to_be(5)

def test_star_single_rest():
    a, *rest, b = [1, 2, 3]
    expect(a).to_be(1)
    expect(rest).to_be([2])
    expect(b).to_be(3)

def test_star_in_for_loop():
    results = []
    for a, *rest in [[1, 2, 3], [4, 5, 6], [7, 8, 9]]:
        results.append((a, rest))
    expect(results[0]).to_be((1, [2, 3]))
    expect(results[1]).to_be((4, [5, 6]))
    expect(results[2]).to_be((7, [8, 9]))

def test_star_in_for_loop_middle():
    results = []
    for a, *rest, b in [[1, 2, 3, 4], [5, 6, 7, 8]]:
        results.append((a, rest, b))
    expect(results[0]).to_be((1, [2, 3], 4))
    expect(results[1]).to_be((5, [6, 7], 8))

def test_star_many_elements():
    a, *rest, b = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    expect(a).to_be(1)
    expect(rest).to_be([2, 3, 4, 5, 6, 7, 8, 9])
    expect(b).to_be(10)

def test_not_enough_values():
    try:
        a, *rest, b, c = [1, 2]
        expect(True).to_be(False)  # Should not reach here
    except ValueError:
        expect(True).to_be(True)

test("star in middle", test_star_in_middle)
test("star at end", test_star_at_end)
test("star at start", test_star_at_start)
test("multiple after star", test_multiple_after_star)
test("multiple before star", test_multiple_before_star)
test("star gets empty list", test_star_gets_empty_list)
test("star from tuple", test_star_from_tuple)
test("star from string", test_star_from_string)
test("star from range", test_star_from_range)
test("star single rest element", test_star_single_rest)
test("star in for loop", test_star_in_for_loop)
test("star in for loop middle", test_star_in_for_loop_middle)
test("star many elements", test_star_many_elements)
test("not enough values error", test_not_enough_values)
