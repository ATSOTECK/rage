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

# =============================================================================
# CPython test_sort.py - Adapted Tests
# =============================================================================

# --- Identity, reversed, and random-ish sorts at various sizes ---

def test_cpython_sort_identity():
    """Sorting an already-sorted list is identity"""
    for n in [0, 1, 2, 3, 4, 7, 8, 15, 16, 31, 32, 63, 64]:
        x = list(range(n))
        s = x[:]
        s.sort()
        expect(s).to_be(x)

test("cpython_sort_identity", test_cpython_sort_identity)

def test_cpython_sort_reversed():
    """Sorting a reversed list produces sorted order"""
    for n in [0, 1, 2, 3, 4, 7, 8, 15, 16, 31, 32]:
        x = list(range(n))
        s = list(range(n - 1, -1, -1))
        s.sort()
        expect(s).to_be(x)

test("cpython_sort_reversed", test_cpython_sort_reversed)

# --- Stability tests (from CPython TestBase.testStressfully) ---

def test_cpython_sort_stability_basic():
    """Equal elements maintain their relative order (stability)"""
    # Pairs of (key, original_index)
    data = [(2, 0), (1, 1), (2, 2), (1, 3), (0, 4), (2, 5), (0, 6)]
    data.sort(key=lambda t: t[0])

    # Within each key group, original indices should be in order
    expect(data).to_be([
        (0, 4), (0, 6),
        (1, 1), (1, 3),
        (2, 0), (2, 2), (2, 5),
    ])

test("cpython_sort_stability_basic", test_cpython_sort_stability_basic)

def test_cpython_sort_stability_key():
    """Stability: using key= to sort on first field matches full sort"""
    data = [(3, 0), (1, 1), (2, 2), (1, 3), (3, 4), (2, 5)]
    copy = data[:]

    data.sort(key=lambda t: t[0])
    copy.sort()  # sort on both fields forces stability

    expect(data).to_be(copy)

test("cpython_sort_stability_key", test_cpython_sort_stability_key)

def test_cpython_sort_stability_small_exhaustive():
    """Exhaustively test stability across small list sizes with few distinct elements"""
    # Adapted from CPython's TestBase.test_small_stability
    NELTS = 3
    MAXSIZE = 5  # Keep small for RAGE

    # Generate all permutations of length `length` over `NELTS` elements
    def product_lists(nelts, length):
        """Generate all tuples of given length from range(nelts)"""
        if length == 0:
            return [()]
        result = []
        sub = product_lists(nelts, length - 1)
        for i in range(nelts):
            for s in sub:
                result.append((i,) + s)
        return result

    for length in range(MAXSIZE + 1):
        for t in product_lists(NELTS, length):
            xs = list(zip(t, range(length)))
            # Stability forced by including index in each element
            forced = sorted(xs)
            # Use key= to hide the index from comparisons
            native = sorted(xs, key=lambda x: x[0])
            expect(native).to_be(forced)

test("cpython_sort_stability_small_exhaustive", test_cpython_sort_stability_small_exhaustive)

# --- Reverse parameter (from CPython TestDecorateSortUndecorate.test_reverse) ---

def test_cpython_sort_reverse_range():
    """Reverse sort produces descending order"""
    data = list(range(100))
    data.sort(reverse=True)
    expect(data).to_be(list(range(99, -1, -1)))

test("cpython_sort_reverse_range", test_cpython_sort_reverse_range)

# --- Reverse stability (from CPython TestDecorateSortUndecorate.test_reverse_stability) ---

def test_cpython_sort_reverse_stability():
    """Reverse sort maintains stability among equal elements"""
    # Pairs of (key, original_index)
    data = [(3, 0), (1, 1), (2, 2), (1, 3), (3, 4), (2, 5), (1, 6)]
    copy = data[:]

    # Sort by key descending
    data.sort(key=lambda x: x[0], reverse=True)
    # Verify descending key order and stable within groups
    keys = [x[0] for x in data]
    expect(keys).to_be([3, 3, 2, 2, 1, 1, 1])

    # Within each key group, original indices should still be in ascending order
    group3 = [x[1] for x in data if x[0] == 3]
    group2 = [x[1] for x in data if x[0] == 2]
    group1 = [x[1] for x in data if x[0] == 1]
    expect(group3).to_be([0, 4])
    expect(group2).to_be([2, 5])
    expect(group1).to_be([1, 3, 6])

test("cpython_sort_reverse_stability", test_cpython_sort_reverse_stability)

# --- Key function that raises (from CPython TestDecorateSortUndecorate.test_key_with_exception) ---

def test_cpython_sort_key_exception():
    """Sort with key function that raises preserves the list"""
    data = list(range(-2, 2))
    dup = data[:]
    got_error = False
    try:
        data.sort(key=lambda x: 1 // x)
    except ZeroDivisionError:
        got_error = True
    expect(got_error).to_be(True)
    # After exception, list should still be a permutation of the original
    expect(sorted(data)).to_be(sorted(dup))

test("cpython_sort_key_exception", test_cpython_sort_key_exception)

# --- TypeError for uncomparable types (from CPython TestOptimizedCompares) ---

def test_cpython_sort_type_error_heterogeneous():
    """Sorting heterogeneous uncomparable types raises TypeError"""
    got_error = False
    try:
        [0, "foo"].sort()
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_sort_type_error_heterogeneous", test_cpython_sort_type_error_heterogeneous)

def test_cpython_sort_type_error_int_str():
    """Sorting mixed int and str raises TypeError"""
    got_error = False
    try:
        sorted([1, "a", 2, "b"])
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_sort_type_error_int_str", test_cpython_sort_type_error_int_str)

def test_cpython_sort_type_error_tuple_mixed():
    """Sorting tuples with mixed types in corresponding positions raises TypeError"""
    got_error = False
    try:
        [("a", 1), (1, "a")].sort()
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_sort_type_error_tuple_mixed", test_cpython_sort_type_error_tuple_mixed)

# --- None in tuples (from CPython TestOptimizedCompares.test_none_in_tuples) ---

def test_cpython_sort_none_in_tuples():
    """Tuples with None compare correctly when first elements are equal"""
    expected = [(None, 1), (None, 2)]
    actual = sorted([(None, 2), (None, 1)])
    expect(actual).to_be(expected)

test("cpython_sort_none_in_tuples", test_cpython_sort_none_in_tuples)

# --- list.sort() returns None (from CPython) ---

def test_cpython_sort_returns_none():
    """list.sort() returns None (sorts in place)"""
    a = [3, 2, 1]
    result = a.sort()
    expect(result).to_be(None)
    expect(a).to_be([1, 2, 3])

test("cpython_sort_returns_none", test_cpython_sort_returns_none)

# --- sorted() vs list.sort() equivalence ---

def test_cpython_sorted_vs_list_sort():
    """sorted() and list.sort() produce the same result"""
    data = [5, 3, 8, 1, 9, 2, 7, 4, 6, 0]
    sorted_copy = sorted(data)
    data.sort()
    expect(data).to_be(sorted_copy)

test("cpython_sorted_vs_list_sort", test_cpython_sorted_vs_list_sort)

# --- Key function sorting with lambda ---

def test_cpython_sort_key_lambda():
    """Sort with lambda key function"""
    data = "The quick Brown fox Jumped over The lazy Dog".split()
    data.sort(key=lambda s: s.lower())
    expect(data).to_be(["Brown", "Dog", "fox", "Jumped", "lazy", "over", "quick", "The", "The"])

test("cpython_sort_key_lambda", test_cpython_sort_key_lambda)

# --- Key + reverse combined ---

def test_cpython_sort_key_reverse_combined():
    """Sort with both key and reverse parameters"""
    words = ["banana", "pie", "Washington", "book"]
    result = sorted(words, key=len, reverse=True)
    expect(result).to_be(["Washington", "banana", "book", "pie"])

test("cpython_sort_key_reverse_combined", test_cpython_sort_key_reverse_combined)

# --- Sort with negative key ---

def test_cpython_sort_negative_key():
    """Sort ascending by negated key gives descending order"""
    data = [3, 1, 4, 1, 5, 9, 2, 6]
    result = sorted(data, key=lambda x: -x)
    expect(result).to_be([9, 6, 5, 4, 3, 2, 1, 1])

test("cpython_sort_negative_key", test_cpython_sort_negative_key)

# --- Sort with complex key function ---

def test_cpython_sort_complex_key():
    """Sort with a key function that returns a tuple"""
    data = [(1, "b"), (2, "a"), (1, "a"), (2, "b")]
    # Sort by second element first, then first element
    result = sorted(data, key=lambda x: (x[1], x[0]))
    expect(result).to_be([(1, "a"), (2, "a"), (1, "b"), (2, "b")])

test("cpython_sort_complex_key", test_cpython_sort_complex_key)

# --- Duplicates sort ---

def test_cpython_sort_many_duplicates():
    """Sort a list with many duplicate values"""
    data = [2, 1, 3, 1, 2, 3, 1, 2, 3]
    expect(sorted(data)).to_be([1, 1, 1, 2, 2, 2, 3, 3, 3])

test("cpython_sort_many_duplicates", test_cpython_sort_many_duplicates)

# --- Two-element sort ---

def test_cpython_sort_two_elements():
    """Sort all permutations of two elements"""
    expect(sorted([1, 2])).to_be([1, 2])
    expect(sorted([2, 1])).to_be([1, 2])
    expect(sorted([1, 1])).to_be([1, 1])

test("cpython_sort_two_elements", test_cpython_sort_two_elements)

# --- Sort with bool key ---

def test_cpython_sort_bool_key():
    """Sort using a boolean key function partitions by predicate"""
    data = [1, -2, 3, -4, 5, -6]
    # False (0) < True (1), so negatives come first
    result = sorted(data, key=lambda x: x > 0)
    expect(result).to_be([-2, -4, -6, 1, 3, 5])

test("cpython_sort_bool_key", test_cpython_sort_bool_key)

# --- Sort with class objects using __lt__ ---

def test_cpython_sort_custom_lt():
    """Sort objects that define __lt__"""
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def __lt__(self, other):
            return (self.x, self.y) < (other.x, other.y)
        def __eq__(self, other):
            return self.x == other.x and self.y == other.y
        def __repr__(self):
            return "Point(" + str(self.x) + ", " + str(self.y) + ")"

    points = [Point(2, 1), Point(1, 2), Point(1, 1), Point(2, 2)]
    points.sort()
    expect(points[0].x).to_be(1)
    expect(points[0].y).to_be(1)
    expect(points[1].x).to_be(1)
    expect(points[1].y).to_be(2)
    expect(points[2].x).to_be(2)
    expect(points[2].y).to_be(1)
    expect(points[3].x).to_be(2)
    expect(points[3].y).to_be(2)

test("cpython_sort_custom_lt", test_cpython_sort_custom_lt)

# --- Sort stability with custom objects ---

def test_cpython_sort_stability_custom_objects():
    """Stability with custom objects: equal keys preserve insertion order"""
    class Item:
        def __init__(self, key, index):
            self.key = key
            self.index = index
        def __lt__(self, other):
            return self.key < other.key
        def __eq__(self, other):
            return self.key == other.key and self.index == other.index

    items = [Item(2, 0), Item(1, 1), Item(2, 2), Item(1, 3), Item(0, 4)]
    items.sort()

    expect(items[0].key).to_be(0)
    expect(items[0].index).to_be(4)
    expect(items[1].key).to_be(1)
    expect(items[1].index).to_be(1)
    expect(items[2].key).to_be(1)
    expect(items[2].index).to_be(3)
    expect(items[3].key).to_be(2)
    expect(items[3].index).to_be(0)
    expect(items[4].key).to_be(2)
    expect(items[4].index).to_be(2)

test("cpython_sort_stability_custom_objects", test_cpython_sort_stability_custom_objects)

# --- Sort of strings by different keys ---

def test_cpython_sort_strings_by_last_char():
    """Sort strings by their last character"""
    words = ["hello", "world", "python", "code"]
    result = sorted(words, key=lambda s: s[-1])
    # 'code' -> 'e', 'hello' -> 'o', 'python' -> 'n', 'world' -> 'd'
    expect(result).to_be(["world", "code", "python", "hello"])

test("cpython_sort_strings_by_last_char", test_cpython_sort_strings_by_last_char)

# --- Mixed int/float sorting ---

def test_cpython_sort_int_float_mixed():
    """Sorting mixed ints and floats works correctly"""
    data = [3, 1.5, 2, 0.5, 4.0, 1]
    expect(sorted(data)).to_be([0.5, 1, 1.5, 2, 3, 4.0])

test("cpython_sort_int_float_mixed", test_cpython_sort_int_float_mixed)

# --- Large reversed list ---

def test_cpython_sort_large_reversed():
    """Sort a large reversed list"""
    n = 200
    data = list(range(n - 1, -1, -1))
    data.sort()
    expect(data).to_be(list(range(n)))

test("cpython_sort_large_reversed", test_cpython_sort_large_reversed)

# --- Sorted preserves original ---

def test_cpython_sorted_preserves_original():
    """sorted() does not modify the original list"""
    original = [5, 3, 1, 4, 2]
    result = sorted(original)
    expect(original).to_be([5, 3, 1, 4, 2])
    expect(result).to_be([1, 2, 3, 4, 5])

test("cpython_sorted_preserves_original", test_cpython_sorted_preserves_original)

# --- Sorted with iterable (tuple, string) ---

def test_cpython_sorted_from_tuple():
    """sorted() works on tuples"""
    expect(sorted((3, 1, 2))).to_be([1, 2, 3])

test("cpython_sorted_from_tuple", test_cpython_sorted_from_tuple)

def test_cpython_sorted_from_string():
    """sorted() works on strings, returning a list of characters"""
    expect(sorted("dcba")).to_be(["a", "b", "c", "d"])

test("cpython_sorted_from_string", test_cpython_sorted_from_string)

# --- Exception from __lt__ during sort ---

def test_cpython_sort_exception_from_lt():
    """Exception in __lt__ propagates from sort"""
    class BadCompare:
        def __init__(self, val):
            self.val = val
        def __lt__(self, other):
            if self.val == 999:
                raise RuntimeError("bad compare")
            return self.val < other.val

    items = [BadCompare(1), BadCompare(999), BadCompare(2)]
    got_error = False
    try:
        items.sort()
    except RuntimeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_sort_exception_from_lt", test_cpython_sort_exception_from_lt)

# --- Sort empty and single-element edge cases ---

def test_cpython_sort_edge_cases():
    """Edge cases: empty, single, all same"""
    a = []
    a.sort()
    expect(a).to_be([])

    b = [42]
    b.sort()
    expect(b).to_be([42])

    c = [7, 7, 7, 7, 7]
    c.sort()
    expect(c).to_be([7, 7, 7, 7, 7])

    d = [7, 7, 7, 7, 7]
    c.sort(reverse=True)
    expect(c).to_be([7, 7, 7, 7, 7])

test("cpython_sort_edge_cases", test_cpython_sort_edge_cases)

print("CPython sort tests completed")
