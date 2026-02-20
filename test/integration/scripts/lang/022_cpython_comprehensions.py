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

# =============================================================================
# Ported from CPython test_genexps.py
# =============================================================================

def test_genexp_sum_with_condition():
    """sum of squares of odd numbers from 0..99"""
    result = sum(i * i for i in range(100) if i & 1 == 1)
    expect(result).to_be(166650)

def test_genexp_simple_nesting():
    """Cartesian product with two for clauses."""
    result = list((i, j) for i in range(3) for j in range(4))
    expect(result).to_be([
        (0, 0), (0, 1), (0, 2), (0, 3),
        (1, 0), (1, 1), (1, 2), (1, 3),
        (2, 0), (2, 1), (2, 2), (2, 3),
    ])

def test_genexp_inner_depends_on_outer():
    """Inner for clause depends on outer loop variable."""
    result = list((i, j) for i in range(4) for j in range(i))
    expect(result).to_be([(1, 0), (2, 0), (2, 1), (3, 0), (3, 1), (3, 2)])

def test_genexp_temp_variable_idiom():
    """Temporary variable assignment idiom using inner for-in-list."""
    result = list(j * j for i in range(4) for j in [i + 1])
    expect(result).to_be([1, 4, 9, 16])

def test_genexp_temp_variable_two_levels():
    """Two levels of temporary variable assignment."""
    result = list(j * k for i in range(4) for j in [i + 1] for k in [j + 1])
    expect(result).to_be([2, 6, 12, 20])

def test_genexp_temp_variable_tuple_unpack():
    """Temporary variable via tuple unpacking in inner for."""
    result = list(j * k for i in range(4) for j, k in [(i + 1, i + 2)])
    expect(result).to_be([2, 6, 12, 20])

def test_genexp_scope_no_leak():
    """Generator expression loop variable does not leak into outer scope."""
    i = 20
    result = sum(i * i for i in range(100))
    expect(result).to_be(328350)
    expect(i).to_be(20)

def test_genexp_first_class():
    """Generator expression is a first-class object."""
    g = (i * i for i in range(4))
    expect(list(g)).to_be([0, 1, 4, 9])

def test_genexp_next_calls():
    """Direct calls to next() on a generator expression."""
    g = (i * i for i in range(3))
    expect(next(g)).to_be(0)
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(4)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

def test_genexp_stays_stopped():
    """Once exhausted, genexp keeps raising StopIteration."""
    g = (i * i for i in range(3))
    expect(next(g)).to_be(0)
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(4)
    caught1 = False
    try:
        next(g)
    except StopIteration:
        caught1 = True
    expect(caught1).to_be(True)
    # Still stopped
    caught2 = False
    try:
        next(g)
    except StopIteration:
        caught2 = True
    expect(caught2).to_be(True)
    expect(list(g)).to_be([])

def test_genexp_defining_function_out_of_scope():
    """Genexp works even after defining function is out of scope."""
    def f(n):
        return (i * i for i in range(n))
    expect(list(f(10))).to_be([0, 1, 4, 9, 16, 25, 36, 49, 64, 81])

def test_genexp_nested_from_function():
    """Nested genexp returned from function."""
    def f(n):
        return ((i, j) for i in range(3) for j in range(n))
    expect(list(f(4))).to_be([
        (0, 0), (0, 1), (0, 2), (0, 3),
        (1, 0), (1, 1), (1, 2), (1, 3),
        (2, 0), (2, 1), (2, 2), (2, 3),
    ])

def test_genexp_with_filter_from_function():
    """Genexp with filter from function, called with different args."""
    def f(n):
        return ((i, j) for i in range(3) for j in range(4) if j in range(n))
    expect(list(f(4))).to_be([
        (0, 0), (0, 1), (0, 2), (0, 3),
        (1, 0), (1, 1), (1, 2), (1, 3),
        (2, 0), (2, 1), (2, 2), (2, 3),
    ])
    expect(list(f(2))).to_be([
        (0, 0), (0, 1),
        (1, 0), (1, 1),
        (2, 0), (2, 1),
    ])

def test_genexp_early_binding_outermost():
    """Outermost for-expression is evaluated eagerly (early binding)."""
    x = 10
    g = (i * i for i in range(x))
    x = 5
    # range(x) was captured when x=10, so we get 10 items not 5
    expect(list(g)).to_be([0, 1, 4, 9, 16, 25, 36, 49, 64, 81])

def test_genexp_late_binding_if():
    """Outermost if-expression is evaluated lazily (late binding)."""
    include = (2, 4, 6, 8)
    g = (i * i for i in range(10) if i in include)
    include = (1, 3, 5, 7, 9)
    # The if-condition uses the NEW value of include
    expect(list(g)).to_be([1, 9, 25, 49, 81])

def test_genexp_late_binding_inner_for():
    """Innermost for-expression is evaluated lazily (late binding)."""
    x = 5
    g = ((i, j) for i in range(3) for j in range(x))
    x = 4
    # Inner range(x) uses x=4 (late binding)
    expect(list(g)).to_be([
        (0, 0), (0, 1), (0, 2), (0, 3),
        (1, 0), (1, 1), (1, 2), (1, 3),
        (2, 0), (2, 1), (2, 2), (2, 3),
    ])

def test_genexp_lambda_yrange():
    """Lambda-based range using generator expression (from CPython tests)."""
    yrange = lambda n: (i for i in range(n))
    expect(list(yrange(10))).to_be([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])

def test_genexp_creator_caller():
    """Genexp returns to most recent caller, not creator."""
    yrange = lambda n: (i for i in range(n))
    log = []
    def creator():
        r = yrange(5)
        log.append("creator " + str(next(r)))
        return r
    def caller():
        r = creator()
        for i in r:
            log.append("caller " + str(i))
    caller()
    expect(log).to_be(["creator 0", "caller 1", "caller 2", "caller 3", "caller 4"])

def test_genexp_exception_propagation():
    """Exception propagation stops the genexp."""
    g = (10 // i for i in (5, 0, 2))
    expect(next(g)).to_be(2)
    caught = False
    try:
        next(g)
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)
    # After exception, genexp is dead
    stopped = False
    try:
        next(g)
    except StopIteration:
        stopped = True
    expect(stopped).to_be(True)

def test_genexp_none_values():
    """Generator expression yielding None."""
    expect(list(None for i in range(10))).to_be(
        [None, None, None, None, None, None, None, None, None, None]
    )

def test_genexp_iter_is_self():
    """iter(genexp) is genexp."""
    g = (i * i for i in range(3))
    expect(iter(g) is g).to_be(True)

def test_genexp_in_sorted():
    """Generator expression passed to sorted()."""
    result = sorted(x * x for x in [3, 1, 4, 1, 5])
    expect(result).to_be([1, 1, 9, 16, 25])

def test_genexp_in_enumerate():
    """Generator expression with enumerate."""
    g = (x * 2 for x in range(5))
    result = list(enumerate(g))
    expect(result).to_be([(0, 0), (1, 2), (2, 4), (3, 6), (4, 8)])

def test_genexp_chained_conditions():
    """Generator expression with multiple chained conditions."""
    result = list(x for x in range(50) if x % 2 == 0 if x % 5 == 0)
    expect(result).to_be([0, 10, 20, 30, 40])

def test_genexp_with_string_methods():
    """Generator expression using string methods."""
    words = ["Hello", "WORLD", "pyThOn"]
    result = list(w.lower() for w in words)
    expect(result).to_be(["hello", "world", "python"])

def test_genexp_nested_three_levels():
    """Three-level nested generator expression."""
    result = list((i, j, k) for i in range(2) for j in range(2) for k in range(2))
    expect(result).to_be([
        (0, 0, 0), (0, 0, 1), (0, 1, 0), (0, 1, 1),
        (1, 0, 0), (1, 0, 1), (1, 1, 0), (1, 1, 1),
    ])

def test_genexp_with_complex_expression():
    """Generator expression with complex transformation."""
    data = [1, 2, 3, 4, 5]
    result = list(
        "big" if x > 3 else "small" for x in data
    )
    expect(result).to_be(["small", "small", "small", "big", "big"])

def test_genexp_early_binding_with_mutation():
    """Early binding: mutating the list after genexp creation."""
    source = [1, 2, 3]
    g = (x * 10 for x in source)
    source.append(4)
    source.append(5)
    # Outermost iterable was already evaluated, so we get items from original list iterator
    # Note: list iterators see mutations, so this gets all 5 items
    result = list(g)
    expect(result).to_be([10, 20, 30, 40, 50])

def test_genexp_dict_from_genexp():
    """Build a dict from a generator expression of pairs."""
    result = dict((k, v) for k, v in [("a", 1), ("b", 2), ("c", 3)])
    expect(result).to_be({"a": 1, "b": 2, "c": 3})

def test_genexp_max_with_key_like():
    """Use genexp to find max based on derived value."""
    data = [-5, -1, 3, 2, -4]
    # Find max absolute value
    result = max(abs(x) for x in data)
    expect(result).to_be(5)

# Register all genexp tests
test("genexp_sum_with_condition", test_genexp_sum_with_condition)
test("genexp_simple_nesting", test_genexp_simple_nesting)
test("genexp_inner_depends_on_outer", test_genexp_inner_depends_on_outer)
test("genexp_temp_variable_idiom", test_genexp_temp_variable_idiom)
test("genexp_temp_variable_two_levels", test_genexp_temp_variable_two_levels)
test("genexp_temp_variable_tuple_unpack", test_genexp_temp_variable_tuple_unpack)
test("genexp_scope_no_leak", test_genexp_scope_no_leak)
test("genexp_first_class", test_genexp_first_class)
test("genexp_next_calls", test_genexp_next_calls)
test("genexp_stays_stopped", test_genexp_stays_stopped)
test("genexp_defining_function_out_of_scope", test_genexp_defining_function_out_of_scope)
test("genexp_nested_from_function", test_genexp_nested_from_function)
test("genexp_with_filter_from_function", test_genexp_with_filter_from_function)
test("genexp_early_binding_outermost", test_genexp_early_binding_outermost)
test("genexp_late_binding_if", test_genexp_late_binding_if)
test("genexp_late_binding_inner_for", test_genexp_late_binding_inner_for)
test("genexp_lambda_yrange", test_genexp_lambda_yrange)
test("genexp_creator_caller", test_genexp_creator_caller)
test("genexp_exception_propagation", test_genexp_exception_propagation)
test("genexp_none_values", test_genexp_none_values)
test("genexp_iter_is_self", test_genexp_iter_is_self)
test("genexp_in_sorted", test_genexp_in_sorted)
test("genexp_in_enumerate", test_genexp_in_enumerate)
test("genexp_chained_conditions", test_genexp_chained_conditions)
test("genexp_with_string_methods", test_genexp_with_string_methods)
test("genexp_nested_three_levels", test_genexp_nested_three_levels)
test("genexp_with_complex_expression", test_genexp_with_complex_expression)
test("genexp_early_binding_with_mutation", test_genexp_early_binding_with_mutation)
test("genexp_dict_from_genexp", test_genexp_dict_from_genexp)
test("genexp_max_with_key_like", test_genexp_max_with_key_like)

print("CPython comprehension tests completed")
