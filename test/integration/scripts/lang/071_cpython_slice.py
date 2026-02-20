# Test: CPython Slice Operations
# Adapted from CPython's test_slice.py and test_listobject.py

from test_framework import test, expect

# === Negative step slicing (reverse) ===
def test_reverse_slice():
    lst = [1, 2, 3, 4, 5]
    expect(lst[::-1]).to_be([5, 4, 3, 2, 1])

def test_reverse_string():
    s = "hello"
    expect(s[::-1]).to_be("olleh")

def test_reverse_empty():
    expect([][::-1]).to_be([])
    expect(""[::-1]).to_be("")

# === Out-of-bounds slicing ===
def test_oob_slice_list():
    lst = [1, 2, 3]
    expect(lst[0:100]).to_be([1, 2, 3])
    expect(lst[-100:100]).to_be([1, 2, 3])
    expect(lst[100:200]).to_be([])
    expect(lst[-100:-200]).to_be([])

def test_oob_slice_string():
    s = "abc"
    expect(s[0:100]).to_be("abc")
    expect(s[-100:100]).to_be("abc")
    expect(s[100:200]).to_be("")

# === Step variations ===
def test_step_two():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(lst[::2]).to_be([0, 2, 4, 6, 8])
    expect(lst[1::2]).to_be([1, 3, 5, 7, 9])

def test_step_three():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(lst[::3]).to_be([0, 3, 6, 9])
    expect(lst[1::3]).to_be([1, 4, 7])

def test_negative_step_two():
    lst = [0, 1, 2, 3, 4, 5]
    expect(lst[::-2]).to_be([5, 3, 1])

def test_step_with_start_stop():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(lst[1:8:2]).to_be([1, 3, 5, 7])
    expect(lst[8:1:-2]).to_be([8, 6, 4, 2])

# === Slice assignment ===
def test_slice_assign_basic():
    lst = [1, 2, 3, 4, 5]
    lst[1:3] = [20, 30]
    expect(lst).to_be([1, 20, 30, 4, 5])

def test_slice_assign_different_length():
    lst = [1, 2, 3, 4, 5]
    lst[1:3] = [20, 30, 40, 50]
    expect(lst).to_be([1, 20, 30, 40, 50, 4, 5])

def test_slice_assign_shorter():
    lst = [1, 2, 3, 4, 5]
    lst[1:4] = [99]
    expect(lst).to_be([1, 99, 5])

def test_slice_assign_empty():
    lst = [1, 2, 3, 4, 5]
    lst[1:4] = []
    expect(lst).to_be([1, 5])

def test_slice_assign_insert():
    lst = [1, 2, 3]
    lst[1:1] = [10, 20]
    expect(lst).to_be([1, 10, 20, 2, 3])

# === Del with slices ===
def test_del_slice():
    lst = [1, 2, 3, 4, 5]
    del lst[1:3]
    expect(lst).to_be([1, 4, 5])

def test_del_slice_all():
    lst = [1, 2, 3, 4, 5]
    del lst[:]
    expect(lst).to_be([])

def test_del_slice_step():
    lst = [0, 1, 2, 3, 4, 5]
    del lst[::2]
    expect(lst).to_be([1, 3, 5])

# === String slicing edge cases ===
def test_string_basic_slice():
    s = "Hello, World!"
    expect(s[0:5]).to_be("Hello")
    expect(s[7:12]).to_be("World")
    expect(s[-1:]).to_be("!")

def test_string_step_slice():
    s = "abcdefgh"
    expect(s[::2]).to_be("aceg")
    expect(s[1::2]).to_be("bdfh")

def test_string_negative_indices():
    s = "python"
    expect(s[-3:]).to_be("hon")
    expect(s[:-3]).to_be("pyt")
    expect(s[-4:-1]).to_be("tho")

# === Tuple slicing ===
def test_tuple_basic_slice():
    t = (1, 2, 3, 4, 5)
    expect(t[1:3]).to_be((2, 3))
    expect(t[:2]).to_be((1, 2))
    expect(t[3:]).to_be((4, 5))

def test_tuple_reverse():
    t = (1, 2, 3, 4, 5)
    expect(t[::-1]).to_be((5, 4, 3, 2, 1))

def test_tuple_step():
    t = (0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
    expect(t[::2]).to_be((0, 2, 4, 6, 8))
    expect(t[1::3]).to_be((1, 4, 7))

# === Empty slices ===
def test_empty_slice_list():
    lst = [1, 2, 3]
    expect(lst[2:1]).to_be([])
    expect(lst[5:10]).to_be([])
    expect(lst[1:1]).to_be([])

def test_empty_slice_string():
    s = "hello"
    expect(s[3:1]).to_be("")
    expect(s[10:20]).to_be("")
    expect(s[2:2]).to_be("")

# === Negative indices ===
def test_negative_index_list():
    lst = [10, 20, 30, 40, 50]
    expect(lst[-1]).to_be(50)
    expect(lst[-2]).to_be(40)
    expect(lst[-5]).to_be(10)
    expect(lst[-3:]).to_be([30, 40, 50])
    expect(lst[:-2]).to_be([10, 20, 30])

def test_negative_index_string():
    s = "abcde"
    expect(s[-1]).to_be("e")
    expect(s[-3:]).to_be("cde")
    expect(s[:-2]).to_be("abc")

# Register all tests
test("reverse_slice", test_reverse_slice)
test("reverse_string", test_reverse_string)
test("reverse_empty", test_reverse_empty)
test("oob_slice_list", test_oob_slice_list)
test("oob_slice_string", test_oob_slice_string)
test("step_two", test_step_two)
test("step_three", test_step_three)
test("negative_step_two", test_negative_step_two)
test("step_with_start_stop", test_step_with_start_stop)
test("slice_assign_basic", test_slice_assign_basic)
test("slice_assign_different_length", test_slice_assign_different_length)
test("slice_assign_shorter", test_slice_assign_shorter)
test("slice_assign_empty", test_slice_assign_empty)
test("slice_assign_insert", test_slice_assign_insert)
test("del_slice", test_del_slice)
test("del_slice_all", test_del_slice_all)
test("del_slice_step", test_del_slice_step)
test("string_basic_slice", test_string_basic_slice)
test("string_step_slice", test_string_step_slice)
test("string_negative_indices", test_string_negative_indices)
test("tuple_basic_slice", test_tuple_basic_slice)
test("tuple_reverse", test_tuple_reverse)
test("tuple_step", test_tuple_step)
test("empty_slice_list", test_empty_slice_list)
test("empty_slice_string", test_empty_slice_string)
test("negative_index_list", test_negative_index_list)
test("negative_index_string", test_negative_index_string)

print("CPython slice tests completed")
