# Test: CPython Sequence Operations
# Adapted from CPython's test_list.py, test_tuple.py, test_string.py

from test_framework import test, expect

# === Negative indexing ===
def test_negative_index_list():
    lst = [10, 20, 30, 40, 50]
    expect(lst[-1]).to_be(50)
    expect(lst[-2]).to_be(40)
    expect(lst[-5]).to_be(10)

def test_negative_index_string():
    s = "abcde"
    expect(s[-1]).to_be("e")
    expect(s[-3]).to_be("c")
    expect(s[-5]).to_be("a")

def test_negative_index_tuple():
    t = (10, 20, 30)
    expect(t[-1]).to_be(30)
    expect(t[-2]).to_be(20)
    expect(t[-3]).to_be(10)

# === Index out of bounds ===
def test_index_oob_list():
    lst = [1, 2, 3]
    caught = [False]
    try:
        x = lst[10]
    except IndexError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_index_oob_negative():
    lst = [1, 2, 3]
    caught = [False]
    try:
        x = lst[-10]
    except IndexError:
        caught[0] = True
    expect(caught[0]).to_be(True)

# === Concatenation with + ===
def test_list_concat():
    expect([1, 2] + [3, 4]).to_be([1, 2, 3, 4])
    expect([] + [1]).to_be([1])
    expect([1] + []).to_be([1])
    expect([] + []).to_be([])

def test_string_concat():
    expect("hello" + " " + "world").to_be("hello world")
    expect("" + "a").to_be("a")
    expect("a" + "").to_be("a")

def test_tuple_concat():
    expect((1, 2) + (3, 4)).to_be((1, 2, 3, 4))
    expect(() + (1,)).to_be((1,))
    expect((1,) + ()).to_be((1,))

# === Repetition with * operator ===
def test_list_repeat():
    expect([1, 2] * 3).to_be([1, 2, 1, 2, 1, 2])
    expect([0] * 5).to_be([0, 0, 0, 0, 0])

def test_list_repeat_zero():
    expect([1, 2, 3] * 0).to_be([])

def test_list_repeat_negative():
    expect([1, 2, 3] * -1).to_be([])

def test_string_repeat():
    expect("ab" * 3).to_be("ababab")
    expect("x" * 5).to_be("xxxxx")

def test_string_repeat_zero():
    expect("hello" * 0).to_be("")

def test_tuple_repeat():
    expect((1, 2) * 3).to_be((1, 2, 1, 2, 1, 2))
    expect((0,) * 4).to_be((0, 0, 0, 0))

# === len() on various types ===
def test_len_list():
    expect(len([])).to_be(0)
    expect(len([1, 2, 3])).to_be(3)

def test_len_string():
    expect(len("")).to_be(0)
    expect(len("hello")).to_be(5)

def test_len_tuple():
    expect(len(())).to_be(0)
    expect(len((1, 2, 3, 4))).to_be(4)

def test_len_dict():
    expect(len({})).to_be(0)
    expect(len({"a": 1, "b": 2})).to_be(2)

# === count() method ===
def test_list_count():
    lst = [1, 2, 2, 3, 2, 4]
    expect(lst.count(2)).to_be(3)
    expect(lst.count(1)).to_be(1)
    expect(lst.count(99)).to_be(0)

def test_string_count():
    s = "hello world hello"
    expect(s.count("hello")).to_be(2)
    expect(s.count("l")).to_be(5)
    expect(s.count("xyz")).to_be(0)

def test_tuple_count():
    t = (1, 2, 2, 3, 2)
    expect(t.count(2)).to_be(3)
    expect(t.count(5)).to_be(0)

# === index() method ===
def test_list_index():
    lst = [10, 20, 30, 40, 50]
    expect(lst.index(30)).to_be(2)
    expect(lst.index(10)).to_be(0)
    expect(lst.index(50)).to_be(4)

def test_tuple_index():
    t = (10, 20, 30, 40, 50)
    expect(t.index(30)).to_be(2)
    expect(t.index(10)).to_be(0)

# === Slice with step ===
def test_list_slice_step():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(lst[::2]).to_be([0, 2, 4, 6, 8])
    expect(lst[1::2]).to_be([1, 3, 5, 7, 9])
    expect(lst[::3]).to_be([0, 3, 6, 9])

def test_list_slice_negative_step():
    lst = [0, 1, 2, 3, 4, 5]
    expect(lst[::-1]).to_be([5, 4, 3, 2, 1, 0])
    expect(lst[::-2]).to_be([5, 3, 1])
    expect(lst[4:1:-1]).to_be([4, 3, 2])

def test_string_slice_step():
    s = "abcdefgh"
    expect(s[::2]).to_be("aceg")
    expect(s[::-1]).to_be("hgfedcba")

# === reversed() builtin ===
def test_reversed_list():
    lst = [1, 2, 3, 4, 5]
    expect(list(reversed(lst))).to_be([5, 4, 3, 2, 1])

def test_reversed_empty():
    expect(list(reversed([]))).to_be([])

# === Tuple immutability ===
def test_tuple_immutable():
    t = (1, 2, 3)
    caught = [False]
    try:
        t[0] = 99
    except TypeError:
        caught[0] = True
    expect(caught[0]).to_be(True)

# === String immutability ===
def test_string_immutable():
    s = "hello"
    caught = [False]
    try:
        s[0] = "H"
    except TypeError:
        caught[0] = True
    expect(caught[0]).to_be(True)

# === List copy behavior (shallow copy via [:]) ===
def test_list_shallow_copy():
    original = [1, 2, 3, 4, 5]
    copy = original[:]
    expect(copy).to_be([1, 2, 3, 4, 5])
    copy.append(6)
    expect(len(original)).to_be(5)  # original unchanged
    expect(len(copy)).to_be(6)

def test_list_shallow_copy_nested():
    original = [[1, 2], [3, 4]]
    copy = original[:]
    # Shallow copy: inner lists are shared
    copy[0].append(99)
    expect(original[0]).to_be([1, 2, 99])

# === Sequence comparison (lexicographic) ===
def test_list_comparison():
    expect([1, 2, 3] < [1, 2, 4]).to_be(True)
    expect([1, 2, 3] < [1, 3, 0]).to_be(True)
    expect([1, 2] < [1, 2, 3]).to_be(True)
    expect([1, 2, 3] == [1, 2, 3]).to_be(True)
    expect([1, 2, 4] > [1, 2, 3]).to_be(True)

def test_string_comparison():
    expect("abc" < "abd").to_be(True)
    expect("abc" < "abcd").to_be(True)
    expect("abc" == "abc").to_be(True)
    expect("abd" > "abc").to_be(True)
    expect("b" > "a").to_be(True)

def test_tuple_comparison():
    expect((1, 2, 3) < (1, 2, 4)).to_be(True)
    expect((1, 2) < (1, 2, 3)).to_be(True)
    expect((1, 2, 3) == (1, 2, 3)).to_be(True)

# Register all tests
test("negative_index_list", test_negative_index_list)
test("negative_index_string", test_negative_index_string)
test("negative_index_tuple", test_negative_index_tuple)
test("index_oob_list", test_index_oob_list)
test("index_oob_negative", test_index_oob_negative)
test("list_concat", test_list_concat)
test("string_concat", test_string_concat)
test("tuple_concat", test_tuple_concat)
test("list_repeat", test_list_repeat)
test("list_repeat_zero", test_list_repeat_zero)
test("list_repeat_negative", test_list_repeat_negative)
test("string_repeat", test_string_repeat)
test("string_repeat_zero", test_string_repeat_zero)
test("tuple_repeat", test_tuple_repeat)
test("len_list", test_len_list)
test("len_string", test_len_string)
test("len_tuple", test_len_tuple)
test("len_dict", test_len_dict)
test("list_count", test_list_count)
test("string_count", test_string_count)
test("tuple_count", test_tuple_count)
test("list_index", test_list_index)
test("tuple_index", test_tuple_index)
test("list_slice_step", test_list_slice_step)
test("list_slice_negative_step", test_list_slice_negative_step)
test("string_slice_step", test_string_slice_step)
test("reversed_list", test_reversed_list)
test("reversed_empty", test_reversed_empty)
test("tuple_immutable", test_tuple_immutable)
test("string_immutable", test_string_immutable)
test("list_shallow_copy", test_list_shallow_copy)
test("list_shallow_copy_nested", test_list_shallow_copy_nested)
test("list_comparison", test_list_comparison)
test("string_comparison", test_string_comparison)
test("tuple_comparison", test_tuple_comparison)

print("CPython sequence ops tests completed")
