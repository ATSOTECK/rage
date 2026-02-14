# Test: CPython Comprehension Edge Cases
# Adapted from CPython comprehension tests - covers advanced patterns
# beyond 07_comprehensions.py

from test_framework import test, expect

# =============================================================================
# Helper functions
# =============================================================================

def is_even(x):
    return x % 2 == 0

def double(x):
    return x * 2

def square(x):
    return x * x

# =============================================================================
# Tests
# =============================================================================

def test_list_comp_complex_condition():
    # Numbers divisible by both 3 and 5
    result = [x for x in range(100) if x % 3 == 0 if x % 5 == 0]
    expect(result).to_be([0, 15, 30, 45, 60, 75, 90])

def test_list_comp_with_function_calls():
    result = [double(x) for x in range(6)]
    expect(result).to_be([0, 2, 4, 6, 8, 10])
    result2 = [square(x) for x in range(5) if is_even(x)]
    expect(result2).to_be([0, 4, 16])

def test_list_comp_nested_2d_flatten():
    matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
    flat = [x for row in matrix for x in row]
    expect(flat).to_be([1, 2, 3, 4, 5, 6, 7, 8, 9])

def test_list_comp_multiple_for_clauses():
    result = [x + y for x in [1, 2] for y in [10, 20, 30]]
    expect(result).to_be([11, 21, 31, 12, 22, 32])

def test_list_comp_multiple_if_clauses():
    result = [x for x in range(50) if x % 2 == 0 if x % 3 == 0 if x > 0]
    expect(result).to_be([6, 12, 18, 24, 30, 36, 42, 48])

def test_list_comp_with_ternary():
    result = ["even" if x % 2 == 0 else "odd" for x in range(6)]
    expect(result).to_be(["even", "odd", "even", "odd", "even", "odd"])

def test_list_comp_string_operations():
    words = ["hello", "world", "python"]
    # Use explicit string method calls
    uppers = [w.upper() for w in words]
    expect(uppers).to_be(["HELLO", "WORLD", "PYTHON"])
    lengths = [len(w) for w in words]
    expect(lengths).to_be([5, 5, 6])

def test_list_comp_with_method_calls():
    data = ["  hello  ", "  world  ", "  test  "]
    stripped = [s.strip() for s in data]
    expect(stripped).to_be(["hello", "world", "test"])

def test_list_comp_with_enumerate():
    words = ["a", "b", "c", "d"]
    indexed = [(i, w) for i, w in enumerate(words)]
    expect(indexed).to_be([(0, "a"), (1, "b"), (2, "c"), (3, "d")])
    # Filter with enumerate
    even_indexed = [w for i, w in enumerate(words) if i % 2 == 0]
    expect(even_indexed).to_be(["a", "c"])

def test_list_comp_over_strings():
    vowels = [c for c in "hello world" if c in "aeiou"]
    expect(vowels).to_be(["e", "o", "o"])
    consonants = [c for c in "hello" if c not in "aeiou"]
    expect(consonants).to_be(["h", "l", "l"])

def test_list_comp_empty_results():
    result = [x for x in range(10) if x > 100]
    expect(result).to_be([])
    result2 = [x for x in []]
    expect(result2).to_be([])
    result3 = [x for x in range(0)]
    expect(result3).to_be([])

def test_list_comp_nested_output():
    # Create a list of lists using comprehension
    # Avoid referencing outer comp variable in inner comp (RAGE closure limitation)
    result = [[j for j in range(i)] for i in range(4)]
    expect(len(result)).to_be(4)
    expect(result[0]).to_be([])
    expect(result[1]).to_be([0])
    expect(result[2]).to_be([0, 1])
    expect(result[3]).to_be([0, 1, 2])

def test_list_comp_with_boolean_ops():
    result = [x for x in range(10) if x > 2 if x < 7]
    expect(result).to_be([3, 4, 5, 6])

def test_dict_comp_basic():
    result = {x: x * x for x in range(5)}
    expect(result).to_be({0: 0, 1: 1, 2: 4, 3: 9, 4: 16})

def test_dict_comp_with_condition():
    result = {x: x * 2 for x in range(10) if x % 2 == 0}
    expect(result).to_be({0: 0, 2: 4, 4: 8, 6: 12, 8: 16})

def test_dict_comp_from_list_of_pairs():
    pairs = [(1, "one"), (2, "two"), (3, "three")]
    result = {k: v for k, v in pairs}
    expect(result).to_be({1: "one", 2: "two", 3: "three"})

def test_dict_comp_swap_keys_values():
    original = {"a": 1, "b": 2, "c": 3}
    swapped = {v: k for k, v in original.items()}
    expect(swapped).to_be({1: "a", 2: "b", 3: "c"})

def test_dict_comp_string_keys():
    result = {str(i): i * i for i in range(4)}
    expect(result).to_be({"0": 0, "1": 1, "2": 4, "3": 9})

def test_set_comp_basic():
    result = {x % 3 for x in range(10)}
    # Should deduplicate: 0, 1, 2
    expect(len(result)).to_be(3)
    expect(0 in result).to_be(True)
    expect(1 in result).to_be(True)
    expect(2 in result).to_be(True)

def test_set_comp_with_condition():
    result = {x for x in range(20) if x % 4 == 0}
    expect(len(result)).to_be(5)
    expect(0 in result).to_be(True)
    expect(4 in result).to_be(True)
    expect(16 in result).to_be(True)

def test_set_comp_deduplication():
    words = ["hello", "world", "hello", "python", "world"]
    unique_lengths = {len(w) for w in words}
    # "hello"=5, "world"=5, "python"=6 -> {5, 6}
    expect(len(unique_lengths)).to_be(2)
    expect(5 in unique_lengths).to_be(True)
    expect(6 in unique_lengths).to_be(True)

def test_generator_expr_with_sum():
    result = sum(x * x for x in range(5))
    expect(result).to_be(30)  # 0+1+4+9+16

def test_generator_expr_with_min_max():
    data = [3, 1, 4, 1, 5, 9, 2, 6]
    result_min = min(x for x in data)
    result_max = max(x for x in data)
    expect(result_min).to_be(1)
    expect(result_max).to_be(9)

def test_generator_expr_with_list():
    result = list(x * 2 for x in range(5))
    expect(result).to_be([0, 2, 4, 6, 8])

def test_comprehension_scope_isolation():
    x = 999
    result = [x for x in range(5)]
    # In Python 3, comprehension variable doesn't leak
    expect(result).to_be([0, 1, 2, 3, 4])
    # x should still be 999 (comprehension has its own scope)
    expect(x).to_be(999)

def test_comprehension_with_type_conversion():
    strings = ["1", "2", "3", "4", "5"]
    nums = [int(s) for s in strings]
    expect(nums).to_be([1, 2, 3, 4, 5])
    back = [str(n) for n in nums]
    expect(back).to_be(["1", "2", "3", "4", "5"])

def test_comprehension_building_dicts():
    keys = ["a", "b", "c"]
    vals = [1, 2, 3]
    n = len(keys)
    pairs = [(keys[i], vals[i]) for i in range(n)]
    expect(pairs).to_be([("a", 1), ("b", 2), ("c", 3)])

# =============================================================================
# Run all tests
# =============================================================================

test("list_comp_complex_condition", test_list_comp_complex_condition)
test("list_comp_with_function_calls", test_list_comp_with_function_calls)
test("list_comp_nested_2d_flatten", test_list_comp_nested_2d_flatten)
test("list_comp_multiple_for_clauses", test_list_comp_multiple_for_clauses)
test("list_comp_multiple_if_clauses", test_list_comp_multiple_if_clauses)
test("list_comp_with_ternary", test_list_comp_with_ternary)
test("list_comp_string_operations", test_list_comp_string_operations)
test("list_comp_with_method_calls", test_list_comp_with_method_calls)
test("list_comp_with_enumerate", test_list_comp_with_enumerate)
test("list_comp_over_strings", test_list_comp_over_strings)
test("list_comp_empty_results", test_list_comp_empty_results)
test("list_comp_nested_output", test_list_comp_nested_output)
test("list_comp_with_boolean_ops", test_list_comp_with_boolean_ops)
test("dict_comp_basic", test_dict_comp_basic)
test("dict_comp_with_condition", test_dict_comp_with_condition)
test("dict_comp_from_list_of_pairs", test_dict_comp_from_list_of_pairs)
test("dict_comp_swap_keys_values", test_dict_comp_swap_keys_values)
test("dict_comp_string_keys", test_dict_comp_string_keys)
test("set_comp_basic", test_set_comp_basic)
test("set_comp_with_condition", test_set_comp_with_condition)
test("set_comp_deduplication", test_set_comp_deduplication)
test("generator_expr_with_sum", test_generator_expr_with_sum)
test("generator_expr_with_min_max", test_generator_expr_with_min_max)
test("generator_expr_with_list", test_generator_expr_with_list)
test("comprehension_scope_isolation", test_comprehension_scope_isolation)
test("comprehension_with_type_conversion", test_comprehension_with_type_conversion)
test("comprehension_building_dicts", test_comprehension_building_dicts)

print("CPython comprehension tests completed")
