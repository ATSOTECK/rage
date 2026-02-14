# Test: CPython List Edge Cases
# Adapted from CPython's test_list.py - covers edge cases beyond 10_collections.py

from test_framework import test, expect

def test_list_construction():
    expect(list()).to_be([])
    expect(list("abc")).to_be(["a", "b", "c"])
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(list((1, 2, 3))).to_be([1, 2, 3])

def test_list_insert():
    lst = [1, 2, 3]
    lst.insert(0, 0)
    expect(lst).to_be([0, 1, 2, 3])
    lst.insert(len(lst), 4)
    expect(lst).to_be([0, 1, 2, 3, 4])
    lst.insert(-1, 99)
    expect(lst).to_be([0, 1, 2, 3, 99, 4])
    lst2 = [1, 2, 3]
    lst2.insert(999, 4)
    expect(lst2).to_be([1, 2, 3, 4])

def test_list_remove():
    lst = [1, 2, 3, 2, 1]
    lst.remove(2)
    expect(lst).to_be([1, 3, 2, 1])
    try:
        lst.remove(99)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_list_clear():
    lst = [1, 2, 3]
    lst.clear()
    expect(lst).to_be([])

def test_list_index():
    lst = [1, 2, 3, 4, 5, 2]
    expect(lst.index(2)).to_be(1)
    expect(lst.index(2, 2)).to_be(5)
    try:
        lst.index(99)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_list_count():
    lst = [1, 2, 2, 3, 2, 4]
    expect(lst.count(2)).to_be(3)
    expect(lst.count(1)).to_be(1)
    expect(lst.count(99)).to_be(0)

def test_list_reverse():
    lst = [1, 2, 3, 4, 5]
    lst.reverse()
    expect(lst).to_be([5, 4, 3, 2, 1])
    empty = []
    empty.reverse()
    expect(empty).to_be([])

def test_list_sort_basic():
    lst = [3, 1, 4, 1, 5, 9, 2, 6]
    lst.sort()
    expect(lst).to_be([1, 1, 2, 3, 4, 5, 6, 9])
    lst.sort(reverse=True)
    expect(lst).to_be([9, 6, 5, 4, 3, 2, 1, 1])

def test_list_sort_key():
    lst = ["banana", "apple", "cherry", "date"]
    lst.sort(key=len)
    expect(lst).to_be(["date", "apple", "banana", "cherry"])

def test_list_sorted_builtin():
    expect(sorted([3, 1, 2])).to_be([1, 2, 3])
    expect(sorted([3, 1, 2], reverse=True)).to_be([3, 2, 1])
    expect(sorted("hello")).to_be(["e", "h", "l", "l", "o"])

def test_list_copy():
    lst = [1, 2, [3, 4]]
    cp = lst.copy()
    expect(cp).to_be([1, 2, [3, 4]])
    cp[0] = 99
    expect(lst[0]).to_be(1)  # Independent
    # Shallow: inner list is shared
    cp[2].append(5)
    expect(lst[2]).to_be([3, 4, 5])

def test_list_slice_read():
    lst = [0, 1, 2, 3, 4, 5]
    expect(lst[1:3]).to_be([1, 2])
    expect(lst[::2]).to_be([0, 2, 4])
    expect(lst[::-1]).to_be([5, 4, 3, 2, 1, 0])
    expect(lst[-2:]).to_be([4, 5])

def test_list_slice_assign():
    lst = [0, 1, 2, 3, 4]
    lst[1:3] = [10, 20, 30]
    expect(lst).to_be([0, 10, 20, 30, 3, 4])
    lst2 = [0, 1, 2, 3, 4]
    lst2[1:4] = []
    expect(lst2).to_be([0, 4])

def test_list_slice_delete():
    lst = [0, 1, 2, 3, 4, 5]
    del lst[1:3]
    expect(lst).to_be([0, 3, 4, 5])

def test_list_iadd():
    lst = [1, 2]
    lst += [3, 4]
    expect(lst).to_be([1, 2, 3, 4])

def test_list_imul():
    lst = [1, 2]
    lst *= 3
    expect(lst).to_be([1, 2, 1, 2, 1, 2])

def test_list_concat():
    a = [1, 2]
    b = [3, 4]
    c = a + b
    expect(c).to_be([1, 2, 3, 4])
    expect(a).to_be([1, 2])  # Originals unchanged

def test_list_repeat():
    expect([1, 2] * 3).to_be([1, 2, 1, 2, 1, 2])
    expect(3 * [1, 2]).to_be([1, 2, 1, 2, 1, 2])
    expect([1] * 0).to_be([])
    expect([1] * -1).to_be([])

def test_list_contains():
    lst = [1, 2, 3, "hello"]
    expect(2 in lst).to_be(True)
    expect(4 not in lst).to_be(True)
    expect("hello" in lst).to_be(True)

def test_list_reversed():
    expect(list(reversed([1, 2, 3]))).to_be([3, 2, 1])
    expect(list(reversed([]))).to_be([])

def test_list_nested():
    lst = [[1, 2], [3, 4]]
    expect(lst[0][1]).to_be(2)
    lst[0].append(99)
    expect(lst[0]).to_be([1, 2, 99])

def test_list_comparison():
    expect([1, 2] == [1, 2]).to_be(True)
    expect([1, 2] != [1, 3]).to_be(True)
    expect([1, 2] < [1, 3]).to_be(True)
    expect([1] < [1, 2]).to_be(True)
    expect([2] > [1, 2, 3]).to_be(True)

def test_list_bool():
    expect(bool([])).to_be(False)
    expect(bool([0])).to_be(True)
    expect(bool([None])).to_be(True)

def test_list_repr():
    expect(repr([1, 2, 3])).to_be("[1, 2, 3]")
    expect(repr([])).to_be("[]")
    expect(repr(["a", "b"])).to_be("['a', 'b']")

def test_list_del_item():
    lst = [1, 2, 3, 4, 5]
    del lst[0]
    expect(lst).to_be([2, 3, 4, 5])
    del lst[-1]
    expect(lst).to_be([2, 3, 4])
    try:
        del lst[99]
        expect("no error").to_be("IndexError")
    except IndexError:
        expect(True).to_be(True)

def test_list_setitem():
    lst = [1, 2, 3]
    lst[0] = 10
    expect(lst).to_be([10, 2, 3])
    lst[-1] = 30
    expect(lst).to_be([10, 2, 30])
    try:
        lst[99] = 1
        expect("no error").to_be("IndexError")
    except IndexError:
        expect(True).to_be(True)

def test_list_pop_index():
    lst = [1, 2, 3, 4, 5]
    expect(lst.pop(0)).to_be(1)
    expect(lst).to_be([2, 3, 4, 5])
    expect(lst.pop(-1)).to_be(5)
    expect(lst).to_be([2, 3, 4])
    empty = []
    try:
        empty.pop()
        expect("no error").to_be("IndexError")
    except IndexError:
        expect(True).to_be(True)

def test_list_extend_generator():
    lst = [1, 2]
    lst.extend(x for x in range(3))
    expect(lst).to_be([1, 2, 0, 1, 2])

def test_list_iteration_basic():
    lst = [10, 20, 30]
    result = []
    for x in lst:
        result.append(x)
    expect(result).to_be([10, 20, 30])
    # enumerate
    result2 = []
    for i, v in enumerate(lst):
        result2.append((i, v))
    expect(result2).to_be([(0, 10), (1, 20), (2, 30)])

def test_list_comprehension_edge():
    expect([x * 2 for x in []]).to_be([])
    expect([x for x in range(5) if x % 2 == 0]).to_be([0, 2, 4])
    expect([[j for j in range(i)] for i in range(4)]).to_be([[], [0], [0, 1], [0, 1, 2]])

# Register all tests
test("list_construction", test_list_construction)
test("list_insert", test_list_insert)
test("list_remove", test_list_remove)
test("list_clear", test_list_clear)
test("list_index", test_list_index)
test("list_count", test_list_count)
test("list_reverse", test_list_reverse)
test("list_sort_basic", test_list_sort_basic)
test("list_sort_key", test_list_sort_key)
test("list_sorted_builtin", test_list_sorted_builtin)
test("list_copy", test_list_copy)
test("list_slice_read", test_list_slice_read)
test("list_slice_assign", test_list_slice_assign)
test("list_slice_delete", test_list_slice_delete)
test("list_iadd", test_list_iadd)
test("list_imul", test_list_imul)
test("list_concat", test_list_concat)
test("list_repeat", test_list_repeat)
test("list_contains", test_list_contains)
test("list_reversed", test_list_reversed)
test("list_nested", test_list_nested)
test("list_comparison", test_list_comparison)
test("list_bool", test_list_bool)
test("list_repr", test_list_repr)
test("list_del_item", test_list_del_item)
test("list_setitem", test_list_setitem)
test("list_pop_index", test_list_pop_index)
test("list_extend_generator", test_list_extend_generator)
test("list_iteration_basic", test_list_iteration_basic)
test("list_comprehension_edge", test_list_comprehension_edge)

print("CPython list tests completed")
