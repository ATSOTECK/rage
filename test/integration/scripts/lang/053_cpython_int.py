# Test: CPython Int Edge Cases
# Adapted from CPython's test_int.py - covers edge cases beyond 01_data_types.py

from test_framework import test, expect

def test_int_from_float_truncation():
    expect(int(3.14)).to_be(3)
    expect(int(-3.14)).to_be(-3)
    expect(int(3.9)).to_be(3)
    expect(int(-3.9)).to_be(-3)
    expect(int(0.999)).to_be(0)
    expect(int(-0.999)).to_be(0)

def test_int_from_string():
    expect(int("314")).to_be(314)
    expect(int("-3")).to_be(-3)
    expect(int(" -3 ")).to_be(-3)
    try:
        int("")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_from_string_bases():
    expect(int("10", 16)).to_be(16)
    expect(int("0x123", 16)).to_be(291)
    expect(int("0o123", 8)).to_be(83)
    expect(int("0b100", 2)).to_be(4)

def test_int_base_zero_autodetect():
    expect(int("0x123", 0)).to_be(291)
    expect(int("0o123", 0)).to_be(83)
    expect(int("0b100", 0)).to_be(4)
    expect(int("000", 0)).to_be(0)

def test_int_invalid_signs():
    try:
        int("+")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        int("-")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        int("- 1")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        int("+ 1")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_no_args():
    expect(int()).to_be(0)

def test_int_keyword_args():
    expect(int("100", base=2)).to_be(4)
    try:
        int(base=10)
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_int_base_limits():
    # Bases 2-36 should accept "0"
    for base in range(2, 37):
        expect(int("0", base)).to_be(0)
    # Base 1 should raise ValueError
    try:
        int("0", 1)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    # Base 37 should raise ValueError
    try:
        int("0", 37)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_float_base():
    # int("0", 5.5) should raise TypeError
    try:
        int("0", 5.5)
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

def test_int_string_float():
    # int("1.2") should raise ValueError
    try:
        int("1.2")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_underscore_literals():
    expect(int("1_000")).to_be(1000)
    expect(int("0xff_ff", 16)).to_be(65535)

def test_int_large_numbers():
    expect(int("1" * 20)).to_be(11111111111111111111)
    expect(int(1e15)).to_be(1000000000000000)

def test_int_bool_is_int():
    expect(int(True)).to_be(1)
    expect(int(False)).to_be(0)
    expect(isinstance(True, int)).to_be(True)
    expect(isinstance(False, int)).to_be(True)

def test_int_base_all_valid():
    # Base 16 should accept 0-9 and a-f
    expect(int("a", 16)).to_be(10)
    expect(int("f", 16)).to_be(15)
    expect(int("z", 36)).to_be(35)
    expect(int("A", 16)).to_be(10)
    expect(int("Z", 36)).to_be(35)

def test_int_leading_zeros():
    # Base 10: leading zeros are ok
    expect(int("0123")).to_be(123)
    # Base 0: leading zeros (non-octal) should raise ValueError
    try:
        int("010", 0)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_uppercase_prefixes():
    expect(int("0X1A", 16)).to_be(26)
    expect(int("0O17", 8)).to_be(15)
    expect(int("0B101", 2)).to_be(5)

def test_int_negative_bases():
    try:
        int("0", -1)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        int("0", 100)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_int_dunder_int():
    class MyInt:
        def __int__(self):
            return 42
    expect(int(MyInt())).to_be(42)

def test_int_dunder_index():
    class MyIndex:
        def __index__(self):
            return 8
    # __index__ should work as base argument
    expect(int("10", MyIndex())).to_be(8)

def test_int_divmod():
    expect(divmod(7, 2)).to_be((3, 1))
    expect(divmod(-7, 2)).to_be((-4, 1))
    expect(divmod(7, -2)).to_be((-4, -1))
    expect(divmod(-7, -2)).to_be((3, -1))

def test_int_bit_operations():
    expect(bin(10)).to_be("0b1010")
    expect(hex(255)).to_be("0xff")
    expect(oct(8)).to_be("0o10")
    # Round-trips
    expect(int(bin(42), 2)).to_be(42)
    expect(int(hex(42), 16)).to_be(42)
    expect(int(oct(42), 8)).to_be(42)

def test_int_pow_three_arg():
    expect(pow(2, 10, 1000)).to_be(24)
    expect(pow(3, 4, 5)).to_be(1)
    expect(pow(7, 3, 13)).to_be(5)

def test_int_abs():
    expect(abs(-42)).to_be(42)
    expect(abs(42)).to_be(42)
    expect(abs(0)).to_be(0)

def test_int_comparison_mixed():
    expect(1 == 1.0).to_be(True)
    expect(2 > 1.5).to_be(True)
    expect(1 < 1.5).to_be(True)
    expect(2 >= 2.0).to_be(True)
    expect(2 <= 2.0).to_be(True)

def test_int_conversion_edge():
    expect(int(True)).to_be(1)
    expect(int(False)).to_be(0)
    expect(int(0.999)).to_be(0)
    expect(int(-0.999)).to_be(0)
    expect(int(1e0)).to_be(1)

# Register all tests
test("int_from_float_truncation", test_int_from_float_truncation)
test("int_from_string", test_int_from_string)
test("int_from_string_bases", test_int_from_string_bases)
test("int_base_zero_autodetect", test_int_base_zero_autodetect)
test("int_invalid_signs", test_int_invalid_signs)
test("int_no_args", test_int_no_args)
test("int_keyword_args", test_int_keyword_args)
test("int_base_limits", test_int_base_limits)
test("int_float_base", test_int_float_base)
test("int_string_float", test_int_string_float)
test("int_underscore_literals", test_int_underscore_literals)
test("int_large_numbers", test_int_large_numbers)
test("int_bool_is_int", test_int_bool_is_int)
test("int_base_all_valid", test_int_base_all_valid)
test("int_leading_zeros", test_int_leading_zeros)
test("int_uppercase_prefixes", test_int_uppercase_prefixes)
test("int_negative_bases", test_int_negative_bases)
test("int_dunder_int", test_int_dunder_int)
test("int_dunder_index", test_int_dunder_index)
test("int_divmod", test_int_divmod)
test("int_bit_operations", test_int_bit_operations)
test("int_pow_three_arg", test_int_pow_three_arg)
test("int_abs", test_int_abs)
test("int_comparison_mixed", test_int_comparison_mixed)
test("int_conversion_edge", test_int_conversion_edge)

print("CPython int tests completed")
