# Test: CPython Float Edge Cases
# Adapted from CPython's test_float.py - covers edge cases beyond 01_data_types.py

from test_framework import test, expect

def test_float_from_string():
    expect(float("3.14")).to_be(3.14)
    expect(float("-1e10")).to_be(-1e10)
    expect(float("inf")).to_be(float("inf"))
    expect(float("-inf")).to_be(float("-inf"))
    # NaN is not equal to itself
    x = float("nan")
    expect(x != x).to_be(True)

def test_float_from_int():
    expect(float(42)).to_be(42.0)
    expect(float(-1)).to_be(-1.0)
    expect(float(0)).to_be(0.0)

def test_float_from_bool():
    expect(float(True)).to_be(1.0)
    expect(float(False)).to_be(0.0)

def test_float_no_args():
    expect(float()).to_be(0.0)

def test_float_special_values():
    expect(float("inf") > 1e308).to_be(True)
    expect(float("-inf") < -1e308).to_be(True)
    nan = float("nan")
    expect(nan != nan).to_be(True)
    expect(nan == nan).to_be(False)

def test_float_arithmetic():
    expect(7.0 / 2.0).to_be(3.5)
    expect(7.0 // 2.0).to_be(3.0)
    expect(7.0 % 2.0).to_be(1.0)
    expect(-7.0 // 2.0).to_be(-4.0)
    expect(-7.0 % 2.0).to_be(1.0)

def test_float_rounding():
    # Python uses banker's rounding (round half to even)
    expect(round(0.5)).to_be(0)
    expect(round(1.5)).to_be(2)
    expect(round(2.5)).to_be(2)
    expect(round(3.5)).to_be(4)
    expect(round(4.5)).to_be(4)
    expect(round(3.14159, 2)).to_be(3.14)

def test_float_int_conversion():
    expect(int(3.9)).to_be(3)
    expect(int(-3.9)).to_be(-3)
    expect(int(0.0)).to_be(0)
    expect(int(-0.0)).to_be(0)

def test_float_comparison():
    # Floating point precision
    expect(0.1 + 0.2 == 0.3).to_be(False)
    expect(1.0 == 1).to_be(True)
    inf = float("inf")
    expect(inf == inf).to_be(True)
    expect(inf > 0).to_be(True)
    expect(-inf < 0).to_be(True)

def test_float_string_repr():
    expect(str(1.0)).to_be("1.0")
    expect(str(0.5)).to_be("0.5")
    expect(str(-3.14)).to_be("-3.14")

def test_float_bool():
    expect(bool(0.0)).to_be(False)
    expect(bool(1.0)).to_be(True)
    expect(bool(-0.5)).to_be(True)
    nan = float("nan")
    expect(bool(nan)).to_be(True)

def test_float_negative_zero():
    expect(-0.0 == 0.0).to_be(True)
    expect(float("-0.0") == 0.0).to_be(True)

def test_float_overflow_string():
    expect(float("1e500")).to_be(float("inf"))
    expect(float("-1e500")).to_be(float("-inf"))

def test_float_invalid_string():
    try:
        float("abc")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)
    try:
        float("")
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_float_whitespace():
    expect(float(" 3.14 ")).to_be(3.14)
    expect(float("  -1.5  ")).to_be(-1.5)

def test_float_underscore():
    expect(float("1_000.5")).to_be(1000.5)
    expect(float("1_000_000")).to_be(1000000.0)

def test_float_methods():
    expect((1.0).is_integer()).to_be(True)
    expect((1.5).is_integer()).to_be(False)
    expect((0.0).is_integer()).to_be(True)
    expect((-2.0).is_integer()).to_be(True)

def test_float_abs():
    expect(abs(-3.14)).to_be(3.14)
    expect(abs(3.14)).to_be(3.14)
    expect(abs(0.0)).to_be(0.0)
    inf = float("inf")
    expect(abs(-inf)).to_be(inf)

def test_float_divmod():
    result = divmod(3.5, 1.5)
    expect(result[0]).to_be(2.0)
    expect(result[1]).to_be(0.5)

def test_float_power():
    expect(2.0 ** 10).to_be(1024.0)
    expect(0.5 ** 2).to_be(0.25)
    expect((-1.0) ** 2).to_be(1.0)

# Register all tests
test("float_from_string", test_float_from_string)
test("float_from_int", test_float_from_int)
test("float_from_bool", test_float_from_bool)
test("float_no_args", test_float_no_args)
test("float_special_values", test_float_special_values)
test("float_arithmetic", test_float_arithmetic)
test("float_rounding", test_float_rounding)
test("float_int_conversion", test_float_int_conversion)
test("float_comparison", test_float_comparison)
test("float_string_repr", test_float_string_repr)
test("float_bool", test_float_bool)
test("float_negative_zero", test_float_negative_zero)
test("float_overflow_string", test_float_overflow_string)
test("float_invalid_string", test_float_invalid_string)
test("float_whitespace", test_float_whitespace)
test("float_underscore", test_float_underscore)
test("float_methods", test_float_methods)
test("float_abs", test_float_abs)
test("float_divmod", test_float_divmod)
test("float_power", test_float_power)

print("CPython float tests completed")
