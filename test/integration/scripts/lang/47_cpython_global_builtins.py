# Test: CPython Global Built-in Functions
# Adapted from CPython's test_builtin.py - covers built-in functions
# not already covered by other test files

from test_framework import test, expect

# =============================================================================
# all() and any()
# =============================================================================

def test_all_true():
    expect(all([True, True, True])).to_be(True)
    expect(all([1, 2, 3])).to_be(True)
    expect(all(["a", "b", "c"])).to_be(True)

def test_all_false():
    expect(all([True, False, True])).to_be(False)
    expect(all([1, 0, 2])).to_be(False)
    expect(all(["a", "", "c"])).to_be(False)

def test_all_empty():
    expect(all([])).to_be(True)

def test_any_true():
    expect(any([False, True, False])).to_be(True)
    expect(any([0, 0, 1])).to_be(True)
    expect(any(["", "", "a"])).to_be(True)

def test_any_false():
    expect(any([False, False, False])).to_be(False)
    expect(any([0, 0, 0])).to_be(False)
    expect(any(["", "", ""])).to_be(False)

def test_any_empty():
    expect(any([])).to_be(False)

def test_all_any_generators():
    expect(all(x > 0 for x in [1, 2, 3])).to_be(True)
    expect(all(x > 0 for x in [1, -1, 3])).to_be(False)
    expect(any(x > 5 for x in [1, 2, 6])).to_be(True)
    expect(any(x > 5 for x in [1, 2, 3])).to_be(False)

# =============================================================================
# chr() and ord()
# =============================================================================

def test_chr_basic():
    expect(chr(65)).to_be("A")
    expect(chr(97)).to_be("a")
    expect(chr(48)).to_be("0")
    expect(chr(32)).to_be(" ")
    expect(chr(10)).to_be("\n")

def test_ord_basic():
    expect(ord("A")).to_be(65)
    expect(ord("a")).to_be(97)
    expect(ord("0")).to_be(48)
    expect(ord(" ")).to_be(32)

def test_chr_ord_roundtrip():
    for i in range(128):
        expect(ord(chr(i))).to_be(i)

# =============================================================================
# bin(), oct(), hex()
# =============================================================================

def test_bin():
    expect(bin(0)).to_be("0b0")
    expect(bin(1)).to_be("0b1")
    expect(bin(10)).to_be("0b1010")
    expect(bin(255)).to_be("0b11111111")
    expect(bin(-1)).to_be("-0b1")
    expect(bin(-10)).to_be("-0b1010")

def test_oct():
    expect(oct(0)).to_be("0o0")
    expect(oct(8)).to_be("0o10")
    expect(oct(64)).to_be("0o100")
    expect(oct(-8)).to_be("-0o10")

def test_hex():
    expect(hex(0)).to_be("0x0")
    expect(hex(16)).to_be("0x10")
    expect(hex(255)).to_be("0xff")
    expect(hex(-255)).to_be("-0xff")

# =============================================================================
# callable()
# =============================================================================

def test_callable_function():
    def foo():
        pass
    expect(callable(foo)).to_be(True)

def test_callable_lambda():
    f = lambda x: x + 1
    expect(callable(f)).to_be(True)

def test_callable_builtin():
    expect(callable(len)).to_be(True)
    expect(callable(print)).to_be(True)
    expect(callable(int)).to_be(True)

def test_callable_not_callable():
    expect(callable(42)).to_be(False)
    expect(callable("hello")).to_be(False)
    expect(callable([1, 2])).to_be(False)
    expect(callable(None)).to_be(False)

def test_callable_class():
    class Foo:
        pass
    expect(callable(Foo)).to_be(True)

# =============================================================================
# Object identity (id() not available, test via 'is' operator)
# =============================================================================

def test_identity_same_object():
    a = [1, 2, 3]
    b = a
    expect(a is b).to_be(True)

def test_identity_different_objects():
    a = [1, 2, 3]
    b = [1, 2, 3]
    expect(a is b).to_be(False)

def test_identity_none():
    x = None
    expect(x is None).to_be(True)

# =============================================================================
# hash() basics
# =============================================================================

def test_hash_int():
    expect(hash(42) == hash(42)).to_be(True)

def test_hash_string():
    expect(hash("hello") == hash("hello")).to_be(True)

def test_hash_tuple():
    expect(hash((1, 2)) == hash((1, 2))).to_be(True)

def test_hash_unhashable():
    caught = False
    try:
        hash([1, 2, 3])
    except TypeError:
        caught = True
    expect(caught).to_be(True)

# =============================================================================
# pow() with 2 and 3 args
# =============================================================================

def test_pow_two_args():
    expect(pow(2, 3)).to_be(8)
    expect(pow(2, 0)).to_be(1)
    expect(pow(5, 2)).to_be(25)
    expect(pow(10, 3)).to_be(1000)

def test_pow_three_args():
    # pow(base, exp, mod) - modular exponentiation
    expect(pow(2, 10, 1000)).to_be(1024 % 1000)
    expect(pow(3, 4, 5)).to_be(81 % 5)
    expect(pow(7, 2, 10)).to_be(49 % 10)

# =============================================================================
# divmod()
# =============================================================================

def test_divmod_basic():
    result = divmod(10, 3)
    expect(result[0]).to_be(3)
    expect(result[1]).to_be(1)

def test_divmod_exact():
    result = divmod(9, 3)
    expect(result[0]).to_be(3)
    expect(result[1]).to_be(0)

def test_divmod_negative():
    result = divmod(-7, 2)
    expect(result[0]).to_be(-4)
    expect(result[1]).to_be(1)

# =============================================================================
# sum() with start value
# =============================================================================

def test_sum_basic():
    expect(sum([1, 2, 3])).to_be(6)
    expect(sum([1, 2, 3, 4, 5])).to_be(15)
    expect(sum([])).to_be(0)

def test_sum_with_start():
    expect(sum([1, 2, 3], 10)).to_be(16)
    expect(sum([], 5)).to_be(5)

def test_sum_generator():
    expect(sum(x * x for x in range(4))).to_be(14)  # 0+1+4+9

# =============================================================================
# zip() edge cases
# =============================================================================

def test_zip_basic():
    result = list(zip([1, 2, 3], ["a", "b", "c"]))
    expect(result).to_be([(1, "a"), (2, "b"), (3, "c")])

def test_zip_unequal_lengths():
    result = list(zip([1, 2], ["a", "b", "c"]))
    expect(result).to_be([(1, "a"), (2, "b")])

def test_zip_empty():
    result = list(zip([], []))
    expect(result).to_be([])

def test_zip_single():
    result = list(zip([1, 2, 3]))
    expect(result).to_be([(1,), (2,), (3,)])

def test_zip_three_iterables():
    result = list(zip([1, 2], ["a", "b"], [True, False]))
    expect(result).to_be([(1, "a", True), (2, "b", False)])

# =============================================================================
# map() basics
# =============================================================================

def test_map_basic():
    result = list(map(str, [1, 2, 3]))
    expect(result).to_be(["1", "2", "3"])

def test_map_lambda():
    result = list(map(lambda x: x * 2, [1, 2, 3]))
    expect(result).to_be([2, 4, 6])

def test_map_empty():
    result = list(map(str, []))
    expect(result).to_be([])

# =============================================================================
# filter()
# =============================================================================

def test_filter_basic():
    result = list(filter(lambda x: x > 2, [1, 2, 3, 4, 5]))
    expect(result).to_be([3, 4, 5])

def test_filter_none():
    # filter(None, ...) removes falsy values
    result = list(filter(None, [0, 1, "", "a", None, True, False, [], [1]]))
    expect(result).to_be([1, "a", True, [1]])

def test_filter_empty():
    result = list(filter(lambda x: x > 0, []))
    expect(result).to_be([])

# =============================================================================
# enumerate() with start parameter
# =============================================================================

def test_enumerate_basic():
    result = list(enumerate(["a", "b", "c"]))
    expect(result).to_be([(0, "a"), (1, "b"), (2, "c")])

def test_enumerate_start():
    result = list(enumerate(["a", "b", "c"], 1))
    expect(result).to_be([(1, "a"), (2, "b"), (3, "c")])

def test_enumerate_start_custom():
    result = list(enumerate(["x", "y"], 10))
    expect(result).to_be([(10, "x"), (11, "y")])

def test_enumerate_empty():
    result = list(enumerate([]))
    expect(result).to_be([])

# =============================================================================
# repr() vs str()
# =============================================================================

def test_repr_str_string():
    s = "hello"
    expect(str(s)).to_be("hello")
    expect(repr(s)).to_be("'hello'")

def test_repr_str_int():
    expect(str(42)).to_be("42")
    expect(repr(42)).to_be("42")

def test_repr_str_list():
    expect(str([1, 2, 3])).to_be("[1, 2, 3]")
    expect(repr([1, 2, 3])).to_be("[1, 2, 3]")

def test_repr_str_none():
    expect(str(None)).to_be("None")
    expect(repr(None)).to_be("None")

def test_repr_str_bool():
    expect(str(True)).to_be("True")
    expect(repr(True)).to_be("True")

# =============================================================================
# String formatting via f-strings and str.format()
# =============================================================================

def test_fstring_format():
    expect(f"{42}").to_be("42")
    expect(f"{'hello'}").to_be("hello")

def test_str_format_method():
    expect("{} {}".format("hello", "world")).to_be("hello world")
    expect("{0} {1}".format("a", "b")).to_be("a b")

# =============================================================================
# abs() edge cases
# =============================================================================

def test_abs_int():
    expect(abs(-5)).to_be(5)
    expect(abs(5)).to_be(5)
    expect(abs(0)).to_be(0)

def test_abs_float():
    expect(abs(-3.14)).to_be(3.14)
    expect(abs(3.14)).to_be(3.14)

# =============================================================================
# round() edge cases
# =============================================================================

def test_round_basic():
    expect(round(3.7)).to_be(4)
    expect(round(3.3)).to_be(3)
    expect(round(-3.7)).to_be(-4)

def test_round_ndigits():
    expect(round(3.14159, 2)).to_be(3.14)
    expect(round(1.005, 1)).to_be(1.0)

# =============================================================================
# min() and max() edge cases
# =============================================================================

def test_min_max_basic():
    expect(min(1, 2, 3)).to_be(1)
    expect(max(1, 2, 3)).to_be(3)

def test_min_max_list():
    expect(min([5, 2, 8, 1])).to_be(1)
    expect(max([5, 2, 8, 1])).to_be(8)

def test_min_max_strings():
    expect(min("c", "a", "b")).to_be("a")
    expect(max("c", "a", "b")).to_be("c")

def test_min_max_single():
    expect(min([42])).to_be(42)
    expect(max([42])).to_be(42)

# =============================================================================
# sorted() with key
# =============================================================================

def test_sorted_reverse():
    expect(sorted([3, 1, 2], reverse=True)).to_be([3, 2, 1])

def test_sorted_strings():
    expect(sorted(["banana", "apple", "cherry"])).to_be(["apple", "banana", "cherry"])

# =============================================================================
# Run all tests
# =============================================================================

test("all_true", test_all_true)
test("all_false", test_all_false)
test("all_empty", test_all_empty)
test("any_true", test_any_true)
test("any_false", test_any_false)
test("any_empty", test_any_empty)
test("all_any_generators", test_all_any_generators)
test("chr_basic", test_chr_basic)
test("ord_basic", test_ord_basic)
test("chr_ord_roundtrip", test_chr_ord_roundtrip)
test("bin", test_bin)
test("oct", test_oct)
test("hex", test_hex)
test("callable_function", test_callable_function)
test("callable_lambda", test_callable_lambda)
test("callable_builtin", test_callable_builtin)
test("callable_not_callable", test_callable_not_callable)
test("callable_class", test_callable_class)
test("identity_same_object", test_identity_same_object)
test("identity_different_objects", test_identity_different_objects)
test("identity_none", test_identity_none)
test("hash_int", test_hash_int)
test("hash_string", test_hash_string)
test("hash_tuple", test_hash_tuple)
test("hash_unhashable", test_hash_unhashable)
test("pow_two_args", test_pow_two_args)
test("pow_three_args", test_pow_three_args)
test("divmod_basic", test_divmod_basic)
test("divmod_exact", test_divmod_exact)
test("divmod_negative", test_divmod_negative)
test("sum_basic", test_sum_basic)
test("sum_with_start", test_sum_with_start)
test("sum_generator", test_sum_generator)
test("zip_basic", test_zip_basic)
test("zip_unequal_lengths", test_zip_unequal_lengths)
test("zip_empty", test_zip_empty)
test("zip_single", test_zip_single)
test("zip_three_iterables", test_zip_three_iterables)
test("map_basic", test_map_basic)
test("map_lambda", test_map_lambda)
test("map_empty", test_map_empty)
test("filter_basic", test_filter_basic)
test("filter_none", test_filter_none)
test("filter_empty", test_filter_empty)
test("enumerate_basic", test_enumerate_basic)
test("enumerate_start", test_enumerate_start)
test("enumerate_start_custom", test_enumerate_start_custom)
test("enumerate_empty", test_enumerate_empty)
test("repr_str_string", test_repr_str_string)
test("repr_str_int", test_repr_str_int)
test("repr_str_list", test_repr_str_list)
test("repr_str_none", test_repr_str_none)
test("repr_str_bool", test_repr_str_bool)
test("fstring_format", test_fstring_format)
test("str_format_method", test_str_format_method)
test("abs_int", test_abs_int)
test("abs_float", test_abs_float)
test("round_basic", test_round_basic)
test("round_ndigits", test_round_ndigits)
test("min_max_basic", test_min_max_basic)
test("min_max_list", test_min_max_list)
test("min_max_strings", test_min_max_strings)
test("min_max_single", test_min_max_single)
test("sorted_reverse", test_sorted_reverse)
test("sorted_strings", test_sorted_strings)

print("CPython global builtins tests completed")
