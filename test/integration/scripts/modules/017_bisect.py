from test_framework import test, expect
import bisect

# =====================
# bisect.bisect_left
# =====================

def test_bisect_left_basic():
    a = [1, 3, 5, 7, 9]
    expect(bisect.bisect_left(a, 5)).to_be(2)
    expect(bisect.bisect_left(a, 6)).to_be(3)
    expect(bisect.bisect_left(a, 0)).to_be(0)
    expect(bisect.bisect_left(a, 10)).to_be(5)

test("bisect_left basic", test_bisect_left_basic)

def test_bisect_left_duplicates():
    a = [1, 2, 2, 2, 3, 4]
    expect(bisect.bisect_left(a, 2)).to_be(1)

test("bisect_left with duplicates finds leftmost", test_bisect_left_duplicates)

def test_bisect_left_empty():
    expect(bisect.bisect_left([], 5)).to_be(0)

test("bisect_left empty list", test_bisect_left_empty)

def test_bisect_left_lo_hi():
    a = [1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(bisect.bisect_left(a, 5, 2, 7)).to_be(4)
    expect(bisect.bisect_left(a, 1, 2, 7)).to_be(2)

test("bisect_left with lo/hi bounds", test_bisect_left_lo_hi)

# =====================
# bisect.bisect_right
# =====================

def test_bisect_right_basic():
    a = [1, 3, 5, 7, 9]
    expect(bisect.bisect_right(a, 5)).to_be(3)
    expect(bisect.bisect_right(a, 6)).to_be(3)
    expect(bisect.bisect_right(a, 0)).to_be(0)
    expect(bisect.bisect_right(a, 10)).to_be(5)

test("bisect_right basic", test_bisect_right_basic)

def test_bisect_right_duplicates():
    a = [1, 2, 2, 2, 3, 4]
    expect(bisect.bisect_right(a, 2)).to_be(4)

test("bisect_right with duplicates finds rightmost", test_bisect_right_duplicates)

def test_bisect_alias():
    a = [1, 3, 5, 7, 9]
    expect(bisect.bisect(a, 5)).to_be(3)

test("bisect is alias for bisect_right", test_bisect_alias)

# =====================
# bisect.insort_left
# =====================

def test_insort_left():
    a = [1, 3, 5, 7]
    bisect.insort_left(a, 4)
    expect(a).to_be([1, 3, 4, 5, 7])

test("insort_left basic", test_insort_left)

def test_insort_left_duplicate():
    a = [1, 2, 2, 3]
    bisect.insort_left(a, 2)
    expect(a).to_be([1, 2, 2, 2, 3])
    # The new 2 should be at index 1 (leftmost position)
    # Verify by checking bisect_left finds 1
    expect(bisect.bisect_left(a, 2)).to_be(1)

test("insort_left with duplicate", test_insort_left_duplicate)

def test_insort_left_empty():
    a = []
    bisect.insort_left(a, 5)
    expect(a).to_be([5])

test("insort_left into empty list", test_insort_left_empty)

def test_insort_left_beginning():
    a = [2, 3, 4]
    bisect.insort_left(a, 1)
    expect(a).to_be([1, 2, 3, 4])

test("insort_left at beginning", test_insort_left_beginning)

def test_insort_left_end():
    a = [1, 2, 3]
    bisect.insort_left(a, 4)
    expect(a).to_be([1, 2, 3, 4])

test("insort_left at end", test_insort_left_end)

# =====================
# bisect.insort_right
# =====================

def test_insort_right():
    a = [1, 3, 5, 7]
    bisect.insort_right(a, 4)
    expect(a).to_be([1, 3, 4, 5, 7])

test("insort_right basic", test_insort_right)

def test_insort_right_duplicate():
    a = [1, 2, 2, 3]
    bisect.insort_right(a, 2)
    expect(a).to_be([1, 2, 2, 2, 3])
    # The new 2 should be at index 3 (rightmost position of 2s)
    expect(bisect.bisect_right(a, 2)).to_be(4)

test("insort_right with duplicate", test_insort_right_duplicate)

def test_insort_alias():
    a = [1, 3, 5]
    bisect.insort(a, 4)
    expect(a).to_be([1, 3, 4, 5])

test("insort is alias for insort_right", test_insort_alias)

# =====================
# bisect with key function
# =====================

def test_bisect_left_key():
    # List of tuples sorted by second element
    a = [(0, 1), (0, 3), (0, 5), (0, 7)]
    def get_second(t):
        return t[1]
    idx = bisect.bisect_left(a, (0, 4), key=get_second)
    expect(idx).to_be(2)

test("bisect_left with key function", test_bisect_left_key)

def test_bisect_right_key():
    a = [(0, 1), (0, 3), (0, 5), (0, 7)]
    def get_second(t):
        return t[1]
    idx = bisect.bisect_right(a, (0, 5), key=get_second)
    expect(idx).to_be(3)

test("bisect_right with key function", test_bisect_right_key)

def test_insort_left_key():
    a = ["a", "ccc", "ddddd"]
    bisect.insort_left(a, "bb", key=len)
    expect(a).to_be(["a", "bb", "ccc", "ddddd"])

test("insort_left with key function", test_insort_left_key)

def test_insort_right_key():
    a = ["a", "ccc", "ddddd"]
    bisect.insort_right(a, "bb", key=len)
    expect(a).to_be(["a", "bb", "ccc", "ddddd"])

test("insort_right with key function", test_insort_right_key)

# =====================
# bisect with strings
# =====================

def test_bisect_strings():
    a = ["apple", "banana", "cherry", "date"]
    expect(bisect.bisect_left(a, "cherry")).to_be(2)
    expect(bisect.bisect_right(a, "cherry")).to_be(3)
    expect(bisect.bisect_left(a, "blueberry")).to_be(2)

test("bisect with strings", test_bisect_strings)

# =====================
# bisect with floats
# =====================

def test_bisect_floats():
    a = [1.1, 2.2, 3.3, 4.4]
    expect(bisect.bisect_left(a, 2.5)).to_be(2)
    expect(bisect.bisect_right(a, 2.2)).to_be(2)

test("bisect with floats", test_bisect_floats)

# =====================
# Grade function (classic use case)
# =====================

def test_grade_function():
    def grade(score, breakpoints=[60, 70, 80, 90], grades="FDCBA"):
        i = bisect.bisect(breakpoints, score)
        return grades[i]

    expect(grade(33)).to_be("F")
    expect(grade(60)).to_be("D")
    expect(grade(77)).to_be("C")
    expect(grade(85)).to_be("B")
    expect(grade(99)).to_be("A")

test("grade function (classic bisect example)", test_grade_function)

# =====================
# bisect_left/right with lo/hi kwargs
# =====================

def test_bisect_kwargs():
    a = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(bisect.bisect_left(a, 5, lo=3, hi=8)).to_be(5)
    expect(bisect.bisect_right(a, 5, lo=3, hi=8)).to_be(6)

test("bisect with lo/hi keyword arguments", test_bisect_kwargs)

# =====================
# insort with lo/hi
# =====================

def test_insort_with_bounds():
    a = [1, 2, 5, 6, 7]
    bisect.insort_left(a, 3, 1, 4)
    expect(a).to_be([1, 2, 3, 5, 6, 7])

test("insort with lo/hi bounds", test_insort_with_bounds)

# =====================
# Building sorted list incrementally
# =====================

def test_build_sorted():
    data = [5, 3, 8, 1, 9, 2, 7, 4, 6]
    result = []
    for x in data:
        bisect.insort(result, x)
    expect(result).to_be([1, 2, 3, 4, 5, 6, 7, 8, 9])

test("build sorted list incrementally with insort", test_build_sorted)
