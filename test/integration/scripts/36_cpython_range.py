# Test: CPython Range Edge Cases
# Adapted from CPython's test_range.py

from test_framework import test, expect

def test_range_basic():
    expect(list(range(0))).to_be([])
    expect(list(range(3))).to_be([0, 1, 2])
    expect(list(range(1, 5))).to_be([1, 2, 3, 4])
    expect(list(range(0, 10, 2))).to_be([0, 2, 4, 6, 8])

def test_range_negative_step():
    expect(list(range(5, 0, -1))).to_be([5, 4, 3, 2, 1])
    expect(list(range(10, 0, -3))).to_be([10, 7, 4, 1])
    expect(list(range(-5, -10, -1))).to_be([-5, -6, -7, -8, -9])

def test_range_empty():
    expect(list(range(0))).to_be([])
    expect(list(range(5, 5))).to_be([])
    expect(list(range(5, 0))).to_be([])
    expect(list(range(0, 0, 1))).to_be([])
    expect(list(range(5, 5, -1))).to_be([])
    expect(list(range(0, 5, -1))).to_be([])

def test_range_single_element():
    expect(list(range(1))).to_be([0])
    expect(list(range(5, 6))).to_be([5])
    expect(list(range(5, 4, -1))).to_be([5])

def test_range_len():
    expect(len(range(0))).to_be(0)
    expect(len(range(10))).to_be(10)
    expect(len(range(1, 10))).to_be(9)
    expect(len(range(1, 10, 2))).to_be(5)
    expect(len(range(1, 10, 3))).to_be(3)
    expect(len(range(10, 0, -1))).to_be(10)
    expect(len(range(10, 0, -3))).to_be(4)

def test_range_contains():
    r = range(10)
    expect(0 in r).to_be(True)
    expect(5 in r).to_be(True)
    expect(9 in r).to_be(True)
    expect(10 in r).to_be(False)
    expect(-1 in r).to_be(False)

def test_range_contains_step():
    r = range(0, 20, 3)
    expect(0 in r).to_be(True)
    expect(3 in r).to_be(True)
    expect(6 in r).to_be(True)
    expect(1 in r).to_be(False)
    expect(2 in r).to_be(False)
    expect(20 in r).to_be(False)

def test_range_contains_negative():
    r = range(10, 0, -2)
    expect(10 in r).to_be(True)
    expect(8 in r).to_be(True)
    expect(2 in r).to_be(True)
    expect(1 in r).to_be(False)
    expect(0 in r).to_be(False)

def test_range_bool():
    expect(bool(range(0))).to_be(False)
    expect(bool(range(1))).to_be(True)
    expect(bool(range(5, 5))).to_be(False)
    expect(bool(range(5, 6))).to_be(True)

def test_range_iteration():
    result = []
    for i in range(5):
        result.append(i)
    expect(result).to_be([0, 1, 2, 3, 4])

def test_range_in_comprehension():
    expect([x * 2 for x in range(5)]).to_be([0, 2, 4, 6, 8])
    expect([x for x in range(10) if x % 2 == 0]).to_be([0, 2, 4, 6, 8])

def test_range_nested():
    result = []
    for i in range(3):
        for j in range(3):
            result.append((i, j))
    expect(len(result)).to_be(9)
    expect(result[0]).to_be((0, 0))
    expect(result[-1]).to_be((2, 2))

def test_range_sum():
    expect(sum(range(10))).to_be(45)
    expect(sum(range(1, 101))).to_be(5050)

def test_range_with_builtins():
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(tuple(range(3))).to_be((0, 1, 2))
    expect(min(range(5))).to_be(0)
    expect(max(range(5))).to_be(4)
    expect(sorted(range(5, 0, -1))).to_be([1, 2, 3, 4, 5])

def test_range_reversed():
    expect(list(reversed(range(5)))).to_be([4, 3, 2, 1, 0])
    expect(list(reversed(range(0, 10, 2)))).to_be([8, 6, 4, 2, 0])
    expect(list(reversed(range(5, 0, -1)))).to_be([1, 2, 3, 4, 5])

def test_range_equality():
    expect(range(10) == range(10)).to_be(True)
    expect(range(0) == range(0)).to_be(True)
    expect(range(1, 10, 2) == range(1, 10, 2)).to_be(True)
    expect(range(10) != range(5)).to_be(True)

def test_range_zero_step_error():
    try:
        range(0, 10, 0)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_range_type_error():
    try:
        range(1.0)
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_range_negative_values():
    expect(list(range(-3, 3))).to_be([-3, -2, -1, 0, 1, 2])
    expect(list(range(-10, -5))).to_be([-10, -9, -8, -7, -6])

def test_range_large_step():
    expect(list(range(0, 100, 25))).to_be([0, 25, 50, 75])
    expect(list(range(100, 0, -25))).to_be([100, 75, 50, 25])

# Register all tests
test("range_basic", test_range_basic)
test("range_negative_step", test_range_negative_step)
test("range_empty", test_range_empty)
test("range_single_element", test_range_single_element)
test("range_len", test_range_len)
test("range_contains", test_range_contains)
test("range_contains_step", test_range_contains_step)
test("range_contains_negative", test_range_contains_negative)
test("range_bool", test_range_bool)
test("range_iteration", test_range_iteration)
test("range_in_comprehension", test_range_in_comprehension)
test("range_nested", test_range_nested)
test("range_sum", test_range_sum)
test("range_with_builtins", test_range_with_builtins)
test("range_reversed", test_range_reversed)
test("range_equality", test_range_equality)
test("range_zero_step_error", test_range_zero_step_error)
test("range_type_error", test_range_type_error)
test("range_negative_values", test_range_negative_values)
test("range_large_step", test_range_large_step)

print("CPython range tests completed")
