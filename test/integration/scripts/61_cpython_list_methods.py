# Test: CPython List Methods - Deep Dive
# Adapted from CPython's test_list.py - covers additional list method edge cases
# beyond 10_collections.py and 31_cpython_list.py

from test_framework import test, expect

def test_insert_negative_index():
    lst = [1, 2, 3, 4, 5]
    lst.insert(-1, 99)
    expect(lst).to_be([1, 2, 3, 4, 99, 5])
    lst2 = [1, 2, 3]
    lst2.insert(-100, 0)  # Very negative -> inserts at beginning
    expect(lst2).to_be([0, 1, 2, 3])

def test_insert_beyond_length():
    lst = [1, 2, 3]
    lst.insert(100, 4)  # Beyond length -> appends
    expect(lst).to_be([1, 2, 3, 4])
    lst.insert(0, 0)  # At beginning
    expect(lst).to_be([0, 1, 2, 3, 4])

def test_insert_into_empty():
    lst = []
    lst.insert(0, "a")
    expect(lst).to_be(["a"])
    lst.insert(0, "b")
    expect(lst).to_be(["b", "a"])

def test_remove_first_occurrence():
    lst = [1, 2, 3, 2, 1]
    lst.remove(2)
    expect(lst).to_be([1, 3, 2, 1])
    lst.remove(2)
    expect(lst).to_be([1, 3, 1])

def test_remove_not_found():
    lst = [1, 2, 3]
    try:
        lst.remove(99)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_remove_different_types():
    lst = [1, "hello", 3.14, True]
    lst.remove("hello")
    expect(len(lst)).to_be(3)
    lst.remove(3.14)
    expect(len(lst)).to_be(2)

def test_pop_default_last():
    lst = [1, 2, 3, 4, 5]
    result = lst.pop()
    expect(result).to_be(5)
    expect(lst).to_be([1, 2, 3, 4])
    result2 = lst.pop()
    expect(result2).to_be(4)
    expect(lst).to_be([1, 2, 3])

def test_pop_with_index():
    lst = [10, 20, 30, 40, 50]
    result = lst.pop(0)
    expect(result).to_be(10)
    expect(lst).to_be([20, 30, 40, 50])
    result2 = lst.pop(2)
    expect(result2).to_be(40)
    expect(lst).to_be([20, 30, 50])
    result3 = lst.pop(-1)
    expect(result3).to_be(50)
    expect(lst).to_be([20, 30])

def test_pop_empty_raises():
    lst = []
    try:
        lst.pop()
        expect("no error").to_be("IndexError")
    except IndexError:
        expect(True).to_be(True)

def test_clear_and_reuse():
    lst = [1, 2, 3, 4, 5]
    lst.clear()
    expect(lst).to_be([])
    expect(len(lst)).to_be(0)
    # Can reuse after clear
    lst.append(99)
    expect(lst).to_be([99])

def test_copy_independence():
    original = [1, 2, 3]
    copied = original.copy()
    copied.append(4)
    expect(original).to_be([1, 2, 3])
    expect(copied).to_be([1, 2, 3, 4])

def test_copy_shallow_nested():
    original = [[1, 2], [3, 4]]
    copied = original.copy()
    # Shallow copy means inner lists are shared
    copied[0].append(99)
    expect(original[0]).to_be([1, 2, 99])
    # But replacing an element in the copy doesn't affect original
    copied[1] = [5, 6]
    expect(original[1]).to_be([3, 4])

def test_extend_with_list():
    lst = [1, 2]
    lst.extend([3, 4, 5])
    expect(lst).to_be([1, 2, 3, 4, 5])

def test_extend_with_tuple():
    lst = [1, 2]
    lst.extend((3, 4))
    expect(lst).to_be([1, 2, 3, 4])

def test_extend_with_string():
    lst = ["a", "b"]
    lst.extend("cd")
    expect(lst).to_be(["a", "b", "c", "d"])

def test_extend_with_range():
    lst = []
    lst.extend(range(5))
    expect(lst).to_be([0, 1, 2, 3, 4])

def test_extend_empty():
    lst = [1, 2]
    lst.extend([])
    expect(lst).to_be([1, 2])

def test_reverse_in_place():
    lst = [1, 2, 3, 4, 5]
    lst.reverse()
    expect(lst).to_be([5, 4, 3, 2, 1])
    # Reverse again to get back original
    lst.reverse()
    expect(lst).to_be([1, 2, 3, 4, 5])

def test_reverse_single_and_empty():
    lst = [42]
    lst.reverse()
    expect(lst).to_be([42])
    empty = []
    empty.reverse()
    expect(empty).to_be([])

def test_count_basic():
    lst = [1, 2, 2, 3, 2, 4, 2]
    expect(lst.count(2)).to_be(4)
    expect(lst.count(1)).to_be(1)
    expect(lst.count(99)).to_be(0)

def test_count_different_types():
    lst = [1, "1", 1.0, True]
    # 1 == 1.0 == True in Python
    expect(lst.count(1)).to_be(3)  # 1, 1.0, and True all == 1
    expect(lst.count("1")).to_be(1)

def test_index_basic():
    lst = ["a", "b", "c", "d", "e"]
    expect(lst.index("a")).to_be(0)
    expect(lst.index("c")).to_be(2)
    expect(lst.index("e")).to_be(4)

def test_index_with_start():
    lst = [1, 2, 3, 1, 2, 3]
    expect(lst.index(1)).to_be(0)
    expect(lst.index(1, 1)).to_be(3)
    expect(lst.index(2, 2)).to_be(4)

def test_index_not_found():
    lst = [1, 2, 3]
    try:
        lst.index(99)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_list_multiply_edge():
    expect([0] * 5).to_be([0, 0, 0, 0, 0])
    expect([1, 2] * 0).to_be([])
    expect([1, 2] * -1).to_be([])
    expect([] * 100).to_be([])
    expect([None] * 3).to_be([None, None, None])

def test_list_as_stack():
    stack = []
    stack.append("a")
    stack.append("b")
    stack.append("c")
    expect(stack.pop()).to_be("c")
    expect(stack.pop()).to_be("b")
    expect(stack.pop()).to_be("a")
    expect(stack).to_be([])

def test_list_as_queue():
    # Using pop(0) as dequeue (not efficient but works)
    queue = []
    queue.append("first")
    queue.append("second")
    queue.append("third")
    expect(queue.pop(0)).to_be("first")
    expect(queue.pop(0)).to_be("second")
    expect(queue.pop(0)).to_be("third")
    expect(queue).to_be([])

def test_nested_list_operations():
    matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
    expect(matrix[0][0]).to_be(1)
    expect(matrix[1][1]).to_be(5)
    expect(matrix[2][2]).to_be(9)
    # Modify nested
    matrix[1][1] = 99
    expect(matrix[1]).to_be([4, 99, 6])
    # Append to nested
    matrix[0].append(10)
    expect(matrix[0]).to_be([1, 2, 3, 10])

def test_list_equality_vs_identity():
    a = [1, 2, 3]
    b = [1, 2, 3]
    c = a
    expect(a == b).to_be(True)   # Equal
    expect(a == c).to_be(True)   # Same object
    # Modify c affects a (same reference)
    c.append(4)
    expect(a).to_be([1, 2, 3, 4])
    expect(b).to_be([1, 2, 3])  # b is independent

def test_list_equality_nested():
    a = [[1, 2], [3, 4]]
    b = [[1, 2], [3, 4]]
    expect(a == b).to_be(True)
    b[0][0] = 99
    expect(a == b).to_be(False)

def test_list_sort_stability():
    # Python's sort is stable - equal elements maintain their relative order
    # We can test this with a key function
    lst = ["banana", "apple", "cherry", "date", "fig"]
    lst.sort(key=len)
    # "fig" (3) and "date" (4) maintain order among same-length strings
    expect(lst[0]).to_be("fig")  # length 3
    expect(lst[1]).to_be("date")  # length 4

def test_list_sort_reverse():
    lst = [3, 1, 4, 1, 5, 9, 2, 6]
    lst.sort(reverse=True)
    expect(lst).to_be([9, 6, 5, 4, 3, 2, 1, 1])

def test_list_comprehension_nested():
    # Flatten a 2D list
    matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
    flat = [x for row in matrix for x in row]
    expect(flat).to_be([1, 2, 3, 4, 5, 6, 7, 8, 9])

def test_list_comprehension_with_function():
    def square(x):
        return x * x
    result = [square(x) for x in range(6)]
    expect(result).to_be([0, 1, 4, 9, 16, 25])

def test_list_slice_copy():
    # Slicing creates a copy
    original = [1, 2, 3, 4, 5]
    copied = original[:]
    copied[0] = 99
    expect(original[0]).to_be(1)
    expect(copied[0]).to_be(99)

def test_list_slice_step():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    expect(lst[::2]).to_be([0, 2, 4, 6, 8])
    expect(lst[1::2]).to_be([1, 3, 5, 7, 9])
    expect(lst[::-1]).to_be([9, 8, 7, 6, 5, 4, 3, 2, 1, 0])
    expect(lst[8:2:-2]).to_be([8, 6, 4])

def test_list_in_operator():
    lst = [1, "hello", 3.14, None, True]
    expect(1 in lst).to_be(True)
    expect("hello" in lst).to_be(True)
    expect(3.14 in lst).to_be(True)
    expect(None in lst).to_be(True)
    expect("world" not in lst).to_be(True)
    expect(99 not in lst).to_be(True)

def test_list_concatenation():
    a = [1, 2, 3]
    b = [4, 5, 6]
    c = a + b
    expect(c).to_be([1, 2, 3, 4, 5, 6])
    # Original lists unchanged
    expect(a).to_be([1, 2, 3])
    expect(b).to_be([4, 5, 6])
    # Concatenate with empty
    expect(a + []).to_be([1, 2, 3])
    expect([] + b).to_be([4, 5, 6])

# Register all tests
test("insert_negative_index", test_insert_negative_index)
test("insert_beyond_length", test_insert_beyond_length)
test("insert_into_empty", test_insert_into_empty)
test("remove_first_occurrence", test_remove_first_occurrence)
test("remove_not_found", test_remove_not_found)
test("remove_different_types", test_remove_different_types)
test("pop_default_last", test_pop_default_last)
test("pop_with_index", test_pop_with_index)
test("pop_empty_raises", test_pop_empty_raises)
test("clear_and_reuse", test_clear_and_reuse)
test("copy_independence", test_copy_independence)
test("copy_shallow_nested", test_copy_shallow_nested)
test("extend_with_list", test_extend_with_list)
test("extend_with_tuple", test_extend_with_tuple)
test("extend_with_string", test_extend_with_string)
test("extend_with_range", test_extend_with_range)
test("extend_empty", test_extend_empty)
test("reverse_in_place", test_reverse_in_place)
test("reverse_single_and_empty", test_reverse_single_and_empty)
test("count_basic", test_count_basic)
test("count_different_types", test_count_different_types)
test("index_basic", test_index_basic)
test("index_with_start", test_index_with_start)
test("index_not_found", test_index_not_found)
test("list_multiply_edge", test_list_multiply_edge)
test("list_as_stack", test_list_as_stack)
test("list_as_queue", test_list_as_queue)
test("nested_list_operations", test_nested_list_operations)
test("list_equality_vs_identity", test_list_equality_vs_identity)
test("list_equality_nested", test_list_equality_nested)
test("list_sort_stability", test_list_sort_stability)
test("list_sort_reverse", test_list_sort_reverse)
test("list_comprehension_nested", test_list_comprehension_nested)
test("list_comprehension_with_function", test_list_comprehension_with_function)
test("list_slice_copy", test_list_slice_copy)
test("list_slice_step", test_list_slice_step)
test("list_in_operator", test_list_in_operator)
test("list_concatenation", test_list_concatenation)

print("CPython list methods tests completed")
