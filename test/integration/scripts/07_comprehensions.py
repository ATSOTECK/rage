# Test: Comprehensions
# Tests list comprehensions, dict comprehensions, and set comprehensions

results = {}

# =====================================
# List Comprehensions
# =====================================

# Simple list comprehension
results["list_comp_simple"] = [x for x in range(5)]
# Expected: [0, 1, 2, 3, 4]

# List comprehension with expression
results["list_comp_expr"] = [x * 2 for x in range(5)]
# Expected: [0, 2, 4, 6, 8]

# List comprehension with squares
results["list_comp_squares"] = [x * x for x in range(6)]
# Expected: [0, 1, 4, 9, 16, 25]

# List comprehension with condition
results["list_comp_filter"] = [x for x in range(10) if x % 2 == 0]
# Expected: [0, 2, 4, 6, 8]

# List comprehension with multiple conditions
results["list_comp_multi_filter"] = [x for x in range(30) if x % 2 == 0 if x % 3 == 0]
# Expected: [0, 6, 12, 18, 24]

# List comprehension with expression and condition
results["list_comp_expr_filter"] = [x * x for x in range(10) if x % 2 == 1]
# Expected: [1, 9, 25, 49, 81]

# List comprehension over string
results["list_comp_string"] = [c for c in "hello"]
# Expected: ["h", "e", "l", "l", "o"]

# List comprehension over list
results["list_comp_list"] = [x + 10 for x in [1, 2, 3, 4, 5]]
# Expected: [11, 12, 13, 14, 15]

# Empty list comprehension
results["list_comp_empty"] = [x for x in []]
# Expected: []

# List comprehension filtering all
results["list_comp_filter_all"] = [x for x in range(5) if x > 100]
# Expected: []

# Conditional expression in comprehension
results["list_comp_ternary"] = ["even" if x % 2 == 0 else "odd" for x in range(5)]
# Expected: ["even", "odd", "even", "odd", "even"]

# List comprehension with negative numbers
results["list_comp_negative"] = [-x for x in range(5)]
# Expected: [0, -1, -2, -3, -4]

# List comprehension creating tuples
results["list_comp_tuples"] = [(x, x * x) for x in range(4)]
# Expected: [(0, 0), (1, 1), (2, 4), (3, 9)]

# =====================================
# Dict Comprehensions
# =====================================

# Simple dict comprehension
results["dict_comp_simple"] = {x: x * x for x in range(5)}
# Expected: {0: 0, 1: 1, 2: 4, 3: 9, 4: 16}

# Dict comprehension with condition
results["dict_comp_filter"] = {x: x * 2 for x in range(10) if x % 2 == 0}
# Expected: {0: 0, 2: 4, 4: 8, 6: 12, 8: 16}

# Dict comprehension with string keys
results["dict_comp_str_keys"] = {str(x): x for x in range(4)}
# Expected: {"0": 0, "1": 1, "2": 2, "3": 3}

# Empty dict comprehension
results["dict_comp_empty"] = {x: x for x in []}
# Expected: {}

# =====================================
# Set Comprehensions
# =====================================

# Simple set comprehension - result as sorted list for consistent comparison
set_result = {x % 3 for x in range(10)}
results["set_comp_simple"] = len(set_result)
# Expected: 3 (elements 0, 1, 2)

# Set comprehension with filter
set_filtered = {x for x in range(10) if x > 5}
results["set_comp_filter"] = len(set_filtered)
# Expected: 4 (elements 6, 7, 8, 9)

# =====================================
# Nested For (if supported)
# =====================================

# Nested list comprehension (flatten)
matrix = [[1, 2, 3], [4, 5, 6]]
results["list_comp_nested"] = [x for row in matrix for x in row]
# Expected: [1, 2, 3, 4, 5, 6]

# Nested with expression
results["list_comp_nested_expr"] = [x * y for x in [1, 2, 3] for y in [10, 100]]
# Expected: [10, 100, 20, 200, 30, 300]

print("Comprehension tests completed")
