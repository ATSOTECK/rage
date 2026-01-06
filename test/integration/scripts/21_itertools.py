# Test: itertools module
# Tests the itertools standard library module

from test_framework import test, expect

import itertools

def test_count():
    counter = itertools.count(10, 2)
    expect(list(itertools.islice(counter, 5))).to_be([10, 12, 14, 16, 18])

def test_cycle():
    cycler = itertools.cycle([1, 2, 3])
    expect(list(itertools.islice(cycler, 7))).to_be([1, 2, 3, 1, 2, 3, 1])

def test_repeat():
    expect(list(itertools.repeat("x", 4))).to_be(["x", "x", "x", "x"])
    expect(list(itertools.repeat(42, 1))).to_be([42])

def test_accumulate():
    expect(list(itertools.accumulate([1, 2, 3, 4, 5]))).to_be([1, 3, 6, 10, 15])
    expect(list(itertools.accumulate([1, 2, 3, 4], lambda x, y: x * y))).to_be([1, 2, 6, 24])
    expect(list(itertools.accumulate([1, 2, 3], None, 10))).to_be([10, 11, 13, 16])
    expect(list(itertools.accumulate(["a", "b", "c"]))).to_be(["a", "ab", "abc"])

def test_chain():
    expect(list(itertools.chain([1, 2], [3, 4], [5, 6]))).to_be([1, 2, 3, 4, 5, 6])
    expect(list(itertools.chain("ab", "cd"))).to_be(["a", "b", "c", "d"])
    expect(list(itertools.chain([], [1], []))).to_be([1])
    expect(list(itertools.chain([1, 2, 3]))).to_be([1, 2, 3])

def test_compress():
    expect(list(itertools.compress("ABCDEF", [1, 0, 1, 0, 1, 1]))).to_be(["A", "C", "E", "F"])
    expect(list(itertools.compress([1, 2, 3, 4], [True, False, True, False]))).to_be([1, 3])
    expect(list(itertools.compress([1, 2, 3, 4, 5], [1, 1]))).to_be([1, 2])

def test_dropwhile():
    expect(list(itertools.dropwhile(lambda x: x < 5, [1, 4, 6, 3, 8]))).to_be([6, 3, 8])
    expect(list(itertools.dropwhile(lambda x: x > 10, [1, 2, 3]))).to_be([1, 2, 3])
    expect(list(itertools.dropwhile(lambda x: x < 10, [1, 2, 3]))).to_be([])

def test_filterfalse():
    expect(list(itertools.filterfalse(lambda x: x % 2, [1, 2, 3, 4, 5, 6]))).to_be([2, 4, 6])
    expect(list(itertools.filterfalse(lambda x: x > 0, [-2, -1, 0, 1, 2]))).to_be([-2, -1, 0])

def test_groupby():
    data = [1, 1, 2, 2, 2, 3, 1, 1]
    groups = []
    for key, group in itertools.groupby(data):
        groups.append([key, list(group)])
    expect(groups).to_be([[1, [1, 1]], [2, [2, 2, 2]], [3, [3]], [1, [1, 1]]])

def test_groupby_with_key():
    words = ["apple", "apricot", "banana", "berry", "cherry"]
    groups2 = []
    for key, group in itertools.groupby(words, lambda x: x[0]):
        groups2.append([key, list(group)])
    expect(groups2).to_be([["a", ["apple", "apricot"]], ["b", ["banana", "berry"]], ["c", ["cherry"]]])

def test_islice():
    expect(list(itertools.islice(range(10), 5))).to_be([0, 1, 2, 3, 4])
    expect(list(itertools.islice(range(10), 2, 7))).to_be([2, 3, 4, 5, 6])
    expect(list(itertools.islice(range(10), 1, 9, 2))).to_be([1, 3, 5, 7])
    expect(list(itertools.islice([1, 2, 3], 10))).to_be([1, 2, 3])

def test_pairwise():
    expect(list(itertools.pairwise([1, 2, 3, 4, 5]))).to_be([(1, 2), (2, 3), (3, 4), (4, 5)])
    expect(list(itertools.pairwise("ABCD"))).to_be([("A", "B"), ("B", "C"), ("C", "D")])
    expect(list(itertools.pairwise([1]))).to_be([])
    expect(list(itertools.pairwise([]))).to_be([])

def test_starmap():
    expect(list(itertools.starmap(pow, [(2, 3), (3, 2), (10, 2)]))).to_be([8, 9, 100])
    expect(list(itertools.starmap(lambda a, b: a + b, [(1, 2), (3, 4), (5, 6)]))).to_be([3, 7, 11])

def test_takewhile():
    expect(list(itertools.takewhile(lambda x: x < 5, [1, 4, 6, 3, 8]))).to_be([1, 4])
    expect(list(itertools.takewhile(lambda x: x < 10, [1, 2, 3]))).to_be([1, 2, 3])
    expect(list(itertools.takewhile(lambda x: x > 10, [1, 2, 3]))).to_be([])

def test_zip_longest():
    expect(list(itertools.zip_longest([1, 2, 3], [4, 5]))).to_be([(1, 4), (2, 5), (3, None)])
    expect(list(itertools.zip_longest([1], [2, 3], [4, 5, 6]))).to_be([(1, 2, 4), (None, 3, 5), (None, None, 6)])
    expect(list(itertools.zip_longest([1, 2], [3, 4]))).to_be([(1, 3), (2, 4)])

def test_product():
    expect(list(itertools.product([1, 2], [3, 4]))).to_be([(1, 3), (1, 4), (2, 3), (2, 4)])
    expect(list(itertools.product([1, 2], ["a", "b"], [True]))).to_be([(1, "a", True), (1, "b", True), (2, "a", True), (2, "b", True)])
    expect(list(itertools.product([1, 2, 3]))).to_be([(1,), (2,), (3,)])
    expect(list(itertools.product("AB", "xy"))).to_be([("A", "x"), ("A", "y"), ("B", "x"), ("B", "y")])

def test_permutations():
    expect(list(itertools.permutations([1, 2, 3]))).to_be([(1, 2, 3), (1, 3, 2), (2, 1, 3), (2, 3, 1), (3, 1, 2), (3, 2, 1)])
    expect(list(itertools.permutations([1, 2, 3], 2))).to_be([(1, 2), (1, 3), (2, 1), (2, 3), (3, 1), (3, 2)])
    expect(list(itertools.permutations([1, 2, 3], 1))).to_be([(1,), (2,), (3,)])
    expect(list(itertools.permutations("AB"))).to_be([("A", "B"), ("B", "A")])

def test_combinations():
    expect(list(itertools.combinations([1, 2, 3, 4], 2))).to_be([(1, 2), (1, 3), (1, 4), (2, 3), (2, 4), (3, 4)])
    expect(list(itertools.combinations([1, 2, 3, 4], 3))).to_be([(1, 2, 3), (1, 2, 4), (1, 3, 4), (2, 3, 4)])
    expect(list(itertools.combinations([1, 2, 3], 1))).to_be([(1,), (2,), (3,)])
    expect(list(itertools.combinations([1, 2, 3], 0))).to_be([()])
    expect(list(itertools.combinations("ABCD", 2))).to_be([("A", "B"), ("A", "C"), ("A", "D"), ("B", "C"), ("B", "D"), ("C", "D")])

def test_combinations_with_replacement():
    expect(list(itertools.combinations_with_replacement([1, 2, 3], 2))).to_be([(1, 1), (1, 2), (1, 3), (2, 2), (2, 3), (3, 3)])
    expect(list(itertools.combinations_with_replacement([1, 2], 3))).to_be([(1, 1, 1), (1, 1, 2), (1, 2, 2), (2, 2, 2)])
    expect(list(itertools.combinations_with_replacement([1, 2, 3], 1))).to_be([(1,), (2,), (3,)])
    expect(list(itertools.combinations_with_replacement("AB", 2))).to_be([("A", "A"), ("A", "B"), ("B", "B")])

def test_edge_cases():
    expect(list(itertools.chain([], [], []))).to_be([])
    expect(list(itertools.product([1, 2], []))).to_be([])
    expect(list(itertools.combinations([], 2))).to_be([])
    expect(list(itertools.permutations([], 2))).to_be([])
    expect(list(itertools.product([1]))).to_be([(1,)])
    expect(list(itertools.combinations([1], 1))).to_be([(1,)])
    expect(list(itertools.permutations([1], 1))).to_be([(1,)])

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
