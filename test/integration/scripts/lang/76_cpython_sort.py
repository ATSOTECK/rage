# Test: CPython Sorting Edge Cases
# Adapted from CPython's test_sort.py

from test_framework import test, expect

# === Basic sorting ===
def test_sort_basic():
    a = [3, 1, 4, 1, 5, 9, 2, 6]
    expect(sorted(a)).to_be([1, 1, 2, 3, 4, 5, 6, 9])
    # Original unchanged
    expect(a).to_be([3, 1, 4, 1, 5, 9, 2, 6])

def test_sort_empty():
    expect(sorted([])).to_be([])

def test_sort_single():
    expect(sorted([42])).to_be([42])

def test_sort_already_sorted():
    expect(sorted([1, 2, 3, 4, 5])).to_be([1, 2, 3, 4, 5])

def test_sort_reverse_sorted():
    expect(sorted([5, 4, 3, 2, 1])).to_be([1, 2, 3, 4, 5])

# === Reverse ===
def test_sort_reverse():
    expect(sorted([3, 1, 2], reverse=True)).to_be([3, 2, 1])

def test_sort_reverse_empty():
    expect(sorted([], reverse=True)).to_be([])

# === Key function ===
def test_sort_key():
    expect(sorted([-3, 1, -2, 4], key=abs)).to_be([1, -2, -3, 4])

def test_sort_key_len():
    words = ["banana", "pie", "Washington", "book"]
    expect(sorted(words, key=len)).to_be(["pie", "book", "banana", "Washington"])

def test_sort_key_reverse():
    expect(sorted([1, -2, 3, -4], key=abs, reverse=True)).to_be([-4, 3, -2, 1])

# === Stability ===
def test_sort_stability():
    # Stability: equal elements maintain their relative order
    data = [(1, "a"), (2, "b"), (1, "c"), (2, "d")]
    # Sort by first element only
    result = sorted(data, key=lambda x: x[0])
    expect(result).to_be([(1, "a"), (1, "c"), (2, "b"), (2, "d")])

# === String sorting ===
def test_sort_strings():
    expect(sorted(["banana", "apple", "cherry"])).to_be(["apple", "banana", "cherry"])

def test_sort_strings_case():
    # Case-insensitive sort
    expect(sorted(["Banana", "apple", "Cherry"], key=lambda s: s.lower())).to_be(["apple", "Banana", "Cherry"])

# === list.sort() ===
def test_list_sort_inplace():
    a = [3, 1, 2]
    result = a.sort()
    expect(result).to_be(None)
    expect(a).to_be([1, 2, 3])

def test_list_sort_reverse():
    a = [1, 2, 3]
    a.sort(reverse=True)
    expect(a).to_be([3, 2, 1])

def test_list_sort_key():
    a = ["banana", "pie", "book"]
    a.sort(key=len)
    expect(a).to_be(["pie", "book", "banana"])

# === Numeric sorting ===
def test_sort_mixed_int():
    expect(sorted([0, -1, 1, -2, 2])).to_be([-2, -1, 0, 1, 2])

def test_sort_floats():
    expect(sorted([1.5, 1.1, 1.9, 1.3])).to_be([1.1, 1.3, 1.5, 1.9])

# === All equal ===
def test_sort_all_equal():
    expect(sorted([5, 5, 5, 5])).to_be([5, 5, 5, 5])

# === Large list ===
def test_sort_large():
    data = list(range(100, 0, -1))
    expect(sorted(data)).to_be(list(range(1, 101)))

# === Boolean sorting ===
def test_sort_booleans():
    expect(sorted([True, False, True, False])).to_be([False, False, True, True])

# === Nested sort ===
def test_sort_tuples():
    data = [(2, "b"), (1, "a"), (2, "a"), (1, "b")]
    expect(sorted(data)).to_be([(1, "a"), (1, "b"), (2, "a"), (2, "b")])

# Register all tests
test("sort_basic", test_sort_basic)
test("sort_empty", test_sort_empty)
test("sort_single", test_sort_single)
test("sort_already_sorted", test_sort_already_sorted)
test("sort_reverse_sorted", test_sort_reverse_sorted)
test("sort_reverse", test_sort_reverse)
test("sort_reverse_empty", test_sort_reverse_empty)
test("sort_key", test_sort_key)
test("sort_key_len", test_sort_key_len)
test("sort_key_reverse", test_sort_key_reverse)
test("sort_stability", test_sort_stability)
test("sort_strings", test_sort_strings)
test("sort_strings_case", test_sort_strings_case)
test("list_sort_inplace", test_list_sort_inplace)
test("list_sort_reverse", test_list_sort_reverse)
test("list_sort_key", test_list_sort_key)
test("sort_mixed_int", test_sort_mixed_int)
test("sort_floats", test_sort_floats)
test("sort_all_equal", test_sort_all_equal)
test("sort_large", test_sort_large)
test("sort_booleans", test_sort_booleans)
test("sort_tuples", test_sort_tuples)

print("CPython sort tests completed")
