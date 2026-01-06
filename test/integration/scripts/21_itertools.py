# Test: itertools module
# Tests the itertools standard library module

from test_framework import test, expect

import itertools

def test_count():
    counter = itertools.count(10, 2)
    expect([10, 12, 14, 16, 18], list(itertools.islice(counter, 5)))

def test_cycle():
    cycler = itertools.cycle([1, 2, 3])
    expect([1, 2, 3, 1, 2, 3, 1], list(itertools.islice(cycler, 7)))

def test_repeat():
    expect(["x", "x", "x", "x"], list(itertools.repeat("x", 4)))
    expect([42], list(itertools.repeat(42, 1)))

def test_accumulate():
    expect([1, 3, 6, 10, 15], list(itertools.accumulate([1, 2, 3, 4, 5])))
    expect([1, 2, 6, 24], list(itertools.accumulate([1, 2, 3, 4], lambda x, y: x * y)))
    expect([10, 11, 13, 16], list(itertools.accumulate([1, 2, 3], None, 10)))
    expect(["a", "ab", "abc"], list(itertools.accumulate(["a", "b", "c"])))

def test_chain():
    expect([1, 2, 3, 4, 5, 6], list(itertools.chain([1, 2], [3, 4], [5, 6])))
    expect(["a", "b", "c", "d"], list(itertools.chain("ab", "cd")))
    expect([1], list(itertools.chain([], [1], [])))
    expect([1, 2, 3], list(itertools.chain([1, 2, 3])))

def test_compress():
    expect(["A", "C", "E", "F"], list(itertools.compress("ABCDEF", [1, 0, 1, 0, 1, 1])))
    expect([1, 3], list(itertools.compress([1, 2, 3, 4], [True, False, True, False])))
    expect([1, 2], list(itertools.compress([1, 2, 3, 4, 5], [1, 1])))

def test_dropwhile():
    expect([6, 3, 8], list(itertools.dropwhile(lambda x: x < 5, [1, 4, 6, 3, 8])))
    expect([1, 2, 3], list(itertools.dropwhile(lambda x: x > 10, [1, 2, 3])))
    expect([], list(itertools.dropwhile(lambda x: x < 10, [1, 2, 3])))

def test_filterfalse():
    expect([2, 4, 6], list(itertools.filterfalse(lambda x: x % 2, [1, 2, 3, 4, 5, 6])))
    expect([-2, -1, 0], list(itertools.filterfalse(lambda x: x > 0, [-2, -1, 0, 1, 2])))

def test_groupby():
    data = [1, 1, 2, 2, 2, 3, 1, 1]
    groups = []
    for key, group in itertools.groupby(data):
        groups.append([key, list(group)])
    expect([[1, [1, 1]], [2, [2, 2, 2]], [3, [3]], [1, [1, 1]]], groups)

def test_groupby_with_key():
    words = ["apple", "apricot", "banana", "berry", "cherry"]
    groups2 = []
    for key, group in itertools.groupby(words, lambda x: x[0]):
        groups2.append([key, list(group)])
    expect([["a", ["apple", "apricot"]], ["b", ["banana", "berry"]], ["c", ["cherry"]]], groups2)

def test_islice():
    expect([0, 1, 2, 3, 4], list(itertools.islice(range(10), 5)))
    expect([2, 3, 4, 5, 6], list(itertools.islice(range(10), 2, 7)))
    expect([1, 3, 5, 7], list(itertools.islice(range(10), 1, 9, 2)))
    expect([1, 2, 3], list(itertools.islice([1, 2, 3], 10)))

def test_pairwise():
    expect([(1, 2), (2, 3), (3, 4), (4, 5)], list(itertools.pairwise([1, 2, 3, 4, 5])))
    expect([("A", "B"), ("B", "C"), ("C", "D")], list(itertools.pairwise("ABCD")))
    expect([], list(itertools.pairwise([1])))
    expect([], list(itertools.pairwise([])))

def test_starmap():
    expect([8, 9, 100], list(itertools.starmap(pow, [(2, 3), (3, 2), (10, 2)])))
    expect([3, 7, 11], list(itertools.starmap(lambda a, b: a + b, [(1, 2), (3, 4), (5, 6)])))

def test_takewhile():
    expect([1, 4], list(itertools.takewhile(lambda x: x < 5, [1, 4, 6, 3, 8])))
    expect([1, 2, 3], list(itertools.takewhile(lambda x: x < 10, [1, 2, 3])))
    expect([], list(itertools.takewhile(lambda x: x > 10, [1, 2, 3])))

def test_zip_longest():
    expect([(1, 4), (2, 5), (3, None)], list(itertools.zip_longest([1, 2, 3], [4, 5])))
    expect([(1, 2, 4), (None, 3, 5), (None, None, 6)], list(itertools.zip_longest([1], [2, 3], [4, 5, 6])))
    expect([(1, 3), (2, 4)], list(itertools.zip_longest([1, 2], [3, 4])))

def test_product():
    expect([(1, 3), (1, 4), (2, 3), (2, 4)], list(itertools.product([1, 2], [3, 4])))
    expect([(1, "a", True), (1, "b", True), (2, "a", True), (2, "b", True)], list(itertools.product([1, 2], ["a", "b"], [True])))
    expect([(1,), (2,), (3,)], list(itertools.product([1, 2, 3])))
    expect([("A", "x"), ("A", "y"), ("B", "x"), ("B", "y")], list(itertools.product("AB", "xy")))

def test_permutations():
    expect([(1, 2, 3), (1, 3, 2), (2, 1, 3), (2, 3, 1), (3, 1, 2), (3, 2, 1)], list(itertools.permutations([1, 2, 3])))
    expect([(1, 2), (1, 3), (2, 1), (2, 3), (3, 1), (3, 2)], list(itertools.permutations([1, 2, 3], 2)))
    expect([(1,), (2,), (3,)], list(itertools.permutations([1, 2, 3], 1)))
    expect([("A", "B"), ("B", "A")], list(itertools.permutations("AB")))

def test_combinations():
    expect([(1, 2), (1, 3), (1, 4), (2, 3), (2, 4), (3, 4)], list(itertools.combinations([1, 2, 3, 4], 2)))
    expect([(1, 2, 3), (1, 2, 4), (1, 3, 4), (2, 3, 4)], list(itertools.combinations([1, 2, 3, 4], 3)))
    expect([(1,), (2,), (3,)], list(itertools.combinations([1, 2, 3], 1)))
    expect([()], list(itertools.combinations([1, 2, 3], 0)))
    expect([("A", "B"), ("A", "C"), ("A", "D"), ("B", "C"), ("B", "D"), ("C", "D")], list(itertools.combinations("ABCD", 2)))

def test_combinations_with_replacement():
    expect([(1, 1), (1, 2), (1, 3), (2, 2), (2, 3), (3, 3)], list(itertools.combinations_with_replacement([1, 2, 3], 2)))
    expect([(1, 1, 1), (1, 1, 2), (1, 2, 2), (2, 2, 2)], list(itertools.combinations_with_replacement([1, 2], 3)))
    expect([(1,), (2,), (3,)], list(itertools.combinations_with_replacement([1, 2, 3], 1)))
    expect([("A", "A"), ("A", "B"), ("B", "B")], list(itertools.combinations_with_replacement("AB", 2)))

def test_edge_cases():
    expect([], list(itertools.chain([], [], [])))
    expect([], list(itertools.product([1, 2], [])))
    expect([], list(itertools.combinations([], 2)))
    expect([], list(itertools.permutations([], 2)))
    expect([(1,)], list(itertools.product([1])))
    expect([(1,)], list(itertools.combinations([1], 1)))
    expect([(1,)], list(itertools.permutations([1], 1)))

test("count", test_count)
test("cycle", test_cycle)
test("repeat", test_repeat)
test("accumulate", test_accumulate)
test("chain", test_chain)
test("compress", test_compress)
test("dropwhile", test_dropwhile)
test("filterfalse", test_filterfalse)
test("groupby", test_groupby)
test("groupby_with_key", test_groupby_with_key)
test("islice", test_islice)
test("pairwise", test_pairwise)
test("starmap", test_starmap)
test("takewhile", test_takewhile)
test("zip_longest", test_zip_longest)
test("product", test_product)
test("permutations", test_permutations)
test("combinations", test_combinations)
test("combinations_with_replacement", test_combinations_with_replacement)
test("edge_cases", test_edge_cases)

print("itertools tests completed")
