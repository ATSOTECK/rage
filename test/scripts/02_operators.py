# Test: Operators
# Tests all Python operators

results = {}

# Arithmetic operators
results["add"] = 10 + 5
results["subtract"] = 10 - 5
results["multiply"] = 10 * 5
results["divide"] = 10 / 4
results["floor_divide"] = 10 // 3
results["modulo"] = 10 % 3
results["power"] = 2 ** 10
results["unary_positive"] = +5
results["unary_negative"] = -5

# Float arithmetic
results["float_add"] = 1.5 + 2.5
results["float_multiply"] = 2.5 * 4.0
results["float_divide"] = 7.0 / 2.0

# Mixed arithmetic
results["int_float_add"] = 5 + 2.5
results["int_float_multiply"] = 3 * 2.5

# Comparison operators
results["equal_true"] = 5 == 5
results["equal_false"] = 5 == 6
results["not_equal_true"] = 5 != 6
results["not_equal_false"] = 5 != 5
results["less_than_true"] = 5 < 10
results["less_than_false"] = 10 < 5
results["less_equal_true"] = 5 <= 5
results["less_equal_false"] = 6 <= 5
results["greater_than_true"] = 10 > 5
results["greater_than_false"] = 5 > 10
results["greater_equal_true"] = 5 >= 5
results["greater_equal_false"] = 4 >= 5

# Chained comparisons
results["chained_true"] = 1 < 5 < 10
results["chained_false"] = 1 < 10 < 5
results["chained_triple"] = 1 <= 2 <= 3 <= 4

# String comparisons
results["str_equal"] = "hello" == "hello"
results["str_not_equal"] = "hello" != "world"
results["str_less_than"] = "apple" < "banana"

# Identity operators
a = [1, 2, 3]
b = a
c = [1, 2, 3]
results["is_same"] = a is b
results["is_different"] = a is c
results["is_not_same"] = a is not c
results["is_not_different"] = a is not b
results["none_is_none"] = None is None

# Membership operators
results["in_list"] = 2 in [1, 2, 3]
results["not_in_list"] = 5 not in [1, 2, 3]
results["in_string"] = "ell" in "hello"
results["not_in_string"] = "xyz" not in "hello"
results["in_dict"] = "a" in {"a": 1, "b": 2}
results["in_set"] = 2 in {1, 2, 3}
results["in_tuple"] = 2 in (1, 2, 3)

# Logical operators
results["and_true"] = True and True
results["and_false"] = True and False
results["or_true"] = False or True
results["or_false"] = False or False
results["not_true"] = not False
results["not_false"] = not True

# Short-circuit evaluation
results["and_short_circuit"] = False and (1 / 0)  # Should not raise
results["or_short_circuit"] = True or (1 / 0)  # Should not raise

# Logical with non-booleans
results["and_values"] = 5 and 10  # Returns last truthy or first falsy
results["or_values"] = 0 or 10  # Returns first truthy or last value
results["and_with_zero"] = 5 and 0
results["or_with_zero"] = 0 or 5

# Bitwise operators
results["bit_and"] = 0b1100 & 0b1010
results["bit_or"] = 0b1100 | 0b1010
results["bit_xor"] = 0b1100 ^ 0b1010
results["bit_not"] = ~0
results["bit_left_shift"] = 1 << 4
results["bit_right_shift"] = 16 >> 2

# Augmented assignment
x = 10
x += 5
results["aug_add"] = x

x = 10
x -= 3
results["aug_sub"] = x

x = 10
x *= 2
results["aug_mul"] = x

x = 10
x //= 3
results["aug_floordiv"] = x

x = 10
x %= 3
results["aug_mod"] = x

x = 2
x **= 4
results["aug_pow"] = x

x = 0b1111
x &= 0b1010
results["aug_and"] = x

x = 0b1100
x |= 0b0011
results["aug_or"] = x

# Ternary/conditional operator
results["ternary_true"] = "yes" if True else "no"
results["ternary_false"] = "yes" if False else "no"
results["ternary_expr"] = "even" if 10 % 2 == 0 else "odd"

# String operators
results["str_concat"] = "hello" + " " + "world"
results["str_repeat"] = "ab" * 4

# List operators
results["list_concat"] = [1, 2] + [3, 4]
results["list_repeat"] = [1, 2] * 3

print("Operators tests completed")
