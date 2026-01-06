# Test: itertools module
# Tests the itertools standard library module

import itertools

results = {}

# ===================================
# Infinite Iterators (with islice)
# ===================================

# count - generates consecutive integers
counter = itertools.count(10, 2)
results["count_with_islice"] = list(itertools.islice(counter, 5))

# cycle - cycles through an iterable
cycler = itertools.cycle([1, 2, 3])
results["cycle_with_islice"] = list(itertools.islice(cycler, 7))

# repeat - repeats an object
results["repeat_with_times"] = list(itertools.repeat("x", 4))
results["repeat_single"] = list(itertools.repeat(42, 1))

# ===================================
# Terminating Iterators
# ===================================

# accumulate - running totals
results["accumulate_simple"] = list(itertools.accumulate([1, 2, 3, 4, 5]))
results["accumulate_with_func"] = list(itertools.accumulate([1, 2, 3, 4], lambda x, y: x * y))
results["accumulate_with_initial"] = list(itertools.accumulate([1, 2, 3], None, 10))
results["accumulate_strings"] = list(itertools.accumulate(["a", "b", "c"]))

# chain - chains iterables
results["chain_simple"] = list(itertools.chain([1, 2], [3, 4], [5, 6]))
results["chain_strings"] = list(itertools.chain("ab", "cd"))
results["chain_empty"] = list(itertools.chain([], [1], []))
results["chain_single"] = list(itertools.chain([1, 2, 3]))

# compress - filters by selector
results["compress_simple"] = list(itertools.compress("ABCDEF", [1, 0, 1, 0, 1, 1]))
results["compress_bools"] = list(itertools.compress([1, 2, 3, 4], [True, False, True, False]))
results["compress_shorter_selector"] = list(itertools.compress([1, 2, 3, 4, 5], [1, 1]))

# dropwhile - drops while predicate is true
results["dropwhile_simple"] = list(itertools.dropwhile(lambda x: x < 5, [1, 4, 6, 3, 8]))
results["dropwhile_all_false"] = list(itertools.dropwhile(lambda x: x > 10, [1, 2, 3]))
results["dropwhile_all_true"] = list(itertools.dropwhile(lambda x: x < 10, [1, 2, 3]))

# filterfalse - elements where predicate is false
results["filterfalse_even"] = list(itertools.filterfalse(lambda x: x % 2, [1, 2, 3, 4, 5, 6]))
results["filterfalse_positive"] = list(itertools.filterfalse(lambda x: x > 0, [-2, -1, 0, 1, 2]))

# groupby - group consecutive elements
data = [1, 1, 2, 2, 2, 3, 1, 1]
groups = []
for key, group in itertools.groupby(data):
    groups.append([key, list(group)])
results["groupby_simple"] = groups

# groupby with key function
words = ["apple", "apricot", "banana", "berry", "cherry"]
groups2 = []
for key, group in itertools.groupby(words, lambda x: x[0]):
    groups2.append([key, list(group)])
results["groupby_with_key"] = groups2

# islice - slice an iterator
results["islice_stop_only"] = list(itertools.islice(range(10), 5))
results["islice_start_stop"] = list(itertools.islice(range(10), 2, 7))
results["islice_with_step"] = list(itertools.islice(range(10), 1, 9, 2))
results["islice_beyond_length"] = list(itertools.islice([1, 2, 3], 10))

# pairwise - consecutive pairs
results["pairwise_simple"] = list(itertools.pairwise([1, 2, 3, 4, 5]))
results["pairwise_string"] = list(itertools.pairwise("ABCD"))
results["pairwise_short"] = list(itertools.pairwise([1]))
results["pairwise_empty"] = list(itertools.pairwise([]))

# starmap - map with unpacked arguments
results["starmap_pow"] = list(itertools.starmap(pow, [(2, 3), (3, 2), (10, 2)]))
results["starmap_add"] = list(itertools.starmap(lambda a, b: a + b, [(1, 2), (3, 4), (5, 6)]))

# takewhile - take while predicate is true
results["takewhile_simple"] = list(itertools.takewhile(lambda x: x < 5, [1, 4, 6, 3, 8]))
results["takewhile_all_true"] = list(itertools.takewhile(lambda x: x < 10, [1, 2, 3]))
results["takewhile_all_false"] = list(itertools.takewhile(lambda x: x > 10, [1, 2, 3]))

# zip_longest - zip to longest iterable
results["zip_longest_simple"] = list(itertools.zip_longest([1, 2, 3], [4, 5]))
results["zip_longest_three"] = list(itertools.zip_longest([1], [2, 3], [4, 5, 6]))
results["zip_longest_equal"] = list(itertools.zip_longest([1, 2], [3, 4]))

# ===================================
# Combinatoric Iterators
# ===================================

# product - Cartesian product
results["product_two_lists"] = list(itertools.product([1, 2], [3, 4]))
results["product_three_lists"] = list(itertools.product([1, 2], ["a", "b"], [True]))
results["product_single"] = list(itertools.product([1, 2, 3]))
results["product_string"] = list(itertools.product("AB", "xy"))

# permutations - r-length permutations
results["permutations_full"] = list(itertools.permutations([1, 2, 3]))
results["permutations_r2"] = list(itertools.permutations([1, 2, 3], 2))
results["permutations_r1"] = list(itertools.permutations([1, 2, 3], 1))
results["permutations_string"] = list(itertools.permutations("AB"))

# combinations - r-length combinations
results["combinations_r2"] = list(itertools.combinations([1, 2, 3, 4], 2))
results["combinations_r3"] = list(itertools.combinations([1, 2, 3, 4], 3))
results["combinations_r1"] = list(itertools.combinations([1, 2, 3], 1))
results["combinations_r0"] = list(itertools.combinations([1, 2, 3], 0))
results["combinations_string"] = list(itertools.combinations("ABCD", 2))

# combinations_with_replacement
results["cwr_r2"] = list(itertools.combinations_with_replacement([1, 2, 3], 2))
results["cwr_r3"] = list(itertools.combinations_with_replacement([1, 2], 3))
results["cwr_r1"] = list(itertools.combinations_with_replacement([1, 2, 3], 1))
results["cwr_string"] = list(itertools.combinations_with_replacement("AB", 2))

# ===================================
# Edge Cases
# ===================================

# Empty iterables
results["chain_all_empty"] = list(itertools.chain([], [], []))
results["product_empty"] = list(itertools.product([1, 2], []))
results["combinations_empty"] = list(itertools.combinations([], 2))
results["permutations_empty"] = list(itertools.permutations([], 2))

# Single element
results["product_single_elem"] = list(itertools.product([1]))
results["combinations_single"] = list(itertools.combinations([1], 1))
results["permutations_single"] = list(itertools.permutations([1], 1))

print("itertools tests completed")
