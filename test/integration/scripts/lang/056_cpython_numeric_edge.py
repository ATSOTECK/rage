# Test: CPython Numeric Edge Cases
# Adapted from CPython's test_int.py, test_float.py, test_complex.py

from test_framework import test, expect

# === Floor division edge cases: negative numbers ===
def test_floor_div_positive():
    expect(7 // 2).to_be(3)
    expect(10 // 3).to_be(3)
    expect(1 // 1).to_be(1)

def test_floor_div_negative():
    expect(-7 // 2).to_be(-4)
    expect(7 // -2).to_be(-4)
    expect(-7 // -2).to_be(3)

def test_floor_div_float():
    expect(7.0 // 2.0).to_be(3.0)
    expect(-7.0 // 2.0).to_be(-4.0)

# === Modulo edge cases: negative numbers ===
def test_modulo_positive():
    expect(7 % 3).to_be(1)
    expect(10 % 5).to_be(0)

def test_modulo_negative():
    # Python modulo: result has same sign as divisor
    expect(-7 % 3).to_be(2)
    expect(7 % -3).to_be(-2)
    expect(-7 % -3).to_be(-1)

def test_modulo_float():
    expect(abs(7.5 % 2.5) < 0.0001).to_be(True)
    expect(abs(7.0 % 3.0 - 1.0) < 0.0001).to_be(True)

# === Large integer arithmetic ===
def test_large_int_add():
    a = 1000000000
    b = 2000000000
    expect(a + b).to_be(3000000000)

def test_large_int_multiply():
    a = 1000000000
    b = 1000000000
    expect(a * b).to_be(1000000000000000000)

def test_large_int_power():
    expect(2 ** 32).to_be(4294967296)
    expect(2 ** 20).to_be(1048576)

# === Mixed int/float arithmetic: type promotion ===
def test_int_float_add():
    result = 1 + 2.0
    expect(result).to_be(3.0)

def test_int_float_multiply():
    result = 3 * 2.5
    expect(result).to_be(7.5)

def test_int_float_divide():
    # Division always returns float in Python 3
    result = 10 / 2
    expect(result).to_be(5.0)

def test_int_float_floor_div():
    result = 10 // 3.0
    expect(result).to_be(3.0)

# === Power edge cases ===
def test_power_zero_zero():
    expect(0 ** 0).to_be(1)

def test_power_negative_base():
    expect((-2) ** 3).to_be(-8)
    expect((-2) ** 2).to_be(4)

def test_power_one():
    expect(1 ** 1000).to_be(1)
    expect((-1) ** 2).to_be(1)
    expect((-1) ** 3).to_be(-1)

def test_power_float():
    expect(abs(2.0 ** 0.5 - 1.41421356) < 0.0001).to_be(True)
    expect(4.0 ** 0.5).to_be(2.0)

# === Integer from various bases ===
def test_int_from_hex():
    expect(int("ff", 16)).to_be(255)
    expect(int("FF", 16)).to_be(255)
    expect(int("10", 16)).to_be(16)
    expect(int("0", 16)).to_be(0)

def test_int_from_octal():
    expect(int("77", 8)).to_be(63)
    expect(int("10", 8)).to_be(8)

def test_int_from_binary():
    expect(int("11", 2)).to_be(3)
    expect(int("1010", 2)).to_be(10)
    expect(int("0", 2)).to_be(0)

def test_int_from_decimal():
    expect(int("42")).to_be(42)
    expect(int("-42")).to_be(-42)
    expect(int("0")).to_be(0)

# === Float precision edge cases ===
def test_float_equality():
    # 0.1 + 0.2 != 0.3 in floating point
    expect(0.1 + 0.2 == 0.3).to_be(False)
    expect(abs(0.1 + 0.2 - 0.3) < 1e-10).to_be(True)

def test_float_from_string():
    expect(float("3.14")).to_be(3.14)
    expect(float("-1.5")).to_be(-1.5)
    expect(float("0.0")).to_be(0.0)

# === Numeric comparison edge cases ===
def test_int_float_equal():
    expect(1 == 1.0).to_be(True)
    expect(0 == 0.0).to_be(True)
    expect(-5 == -5.0).to_be(True)
    expect(1000000 == 1000000.0).to_be(True)

def test_int_float_comparison():
    expect(1 < 1.5).to_be(True)
    expect(2.5 > 2).to_be(True)
    expect(3 >= 3.0).to_be(True)
    expect(3.0 <= 3).to_be(True)

# === abs() with various types ===
def test_abs_int():
    expect(abs(0)).to_be(0)
    expect(abs(42)).to_be(42)
    expect(abs(-42)).to_be(42)

def test_abs_float():
    expect(abs(0.0)).to_be(0.0)
    expect(abs(3.14)).to_be(3.14)
    expect(abs(-3.14)).to_be(3.14)

# === round() with various inputs ===
def test_round_basic():
    expect(round(3.7)).to_be(4)
    expect(round(3.3)).to_be(3)
    expect(round(-3.7)).to_be(-4)
    expect(round(-3.3)).to_be(-3)

def test_round_integer():
    expect(round(5)).to_be(5)
    expect(round(-3)).to_be(-3)

# === min/max edge cases ===
def test_min_basic():
    expect(min(1, 2, 3)).to_be(1)
    expect(min(-1, -2, -3)).to_be(-3)
    expect(min(0, 0, 0)).to_be(0)

def test_max_basic():
    expect(max(1, 2, 3)).to_be(3)
    expect(max(-1, -2, -3)).to_be(-1)
    expect(max(0, 0, 0)).to_be(0)

def test_min_max_mixed():
    expect(min(1, 2.5, 3)).to_be(1)
    expect(max(1, 2.5, 3)).to_be(3)

def test_min_max_list():
    expect(min([5, 3, 8, 1, 9])).to_be(1)
    expect(max([5, 3, 8, 1, 9])).to_be(9)

# === Boolean arithmetic ===
def test_bool_arithmetic():
    expect(True + True).to_be(2)
    expect(True * 5).to_be(5)
    expect(False + 1).to_be(1)
    expect(True + 0).to_be(1)

# Register all tests
test("floor_div_positive", test_floor_div_positive)
test("floor_div_negative", test_floor_div_negative)
test("floor_div_float", test_floor_div_float)
test("modulo_positive", test_modulo_positive)
test("modulo_negative", test_modulo_negative)
test("modulo_float", test_modulo_float)
test("large_int_add", test_large_int_add)
test("large_int_multiply", test_large_int_multiply)
test("large_int_power", test_large_int_power)
test("int_float_add", test_int_float_add)
test("int_float_multiply", test_int_float_multiply)
test("int_float_divide", test_int_float_divide)
test("int_float_floor_div", test_int_float_floor_div)
test("power_zero_zero", test_power_zero_zero)
test("power_negative_base", test_power_negative_base)
test("power_one", test_power_one)
test("power_float", test_power_float)
test("int_from_hex", test_int_from_hex)
test("int_from_octal", test_int_from_octal)
test("int_from_binary", test_int_from_binary)
test("int_from_decimal", test_int_from_decimal)
test("float_equality", test_float_equality)
test("float_from_string", test_float_from_string)
test("int_float_equal", test_int_float_equal)
test("int_float_comparison", test_int_float_comparison)
test("abs_int", test_abs_int)
test("abs_float", test_abs_float)
test("round_basic", test_round_basic)
test("round_integer", test_round_integer)
test("min_basic", test_min_basic)
test("max_basic", test_max_basic)
test("min_max_mixed", test_min_max_mixed)
test("min_max_list", test_min_max_list)
test("bool_arithmetic", test_bool_arithmetic)

print("CPython numeric edge tests completed")
