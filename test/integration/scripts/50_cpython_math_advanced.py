# Test: CPython Math Module Advanced Tests
# Adapted from CPython's test_math.py

from test_framework import test, expect
import math

# === Constants ===
def test_constants():
    expect(abs(math.pi - 3.141592653589793) < 0.0001).to_be(True)
    expect(abs(math.e - 2.718281828459045) < 0.0001).to_be(True)
    expect(abs(math.tau - 6.283185307179586) < 0.0001).to_be(True)
    expect(math.inf > 1e308).to_be(True)
    expect(math.inf == math.inf).to_be(True)

def test_nan_constant():
    # NaN is not equal to itself
    expect(math.nan != math.nan).to_be(True)

# === Special value checks ===
def test_isnan():
    expect(math.isnan(math.nan)).to_be(True)
    expect(math.isnan(0.0)).to_be(False)
    expect(math.isnan(1.0)).to_be(False)
    expect(math.isnan(math.inf)).to_be(False)

def test_isinf():
    expect(math.isinf(math.inf)).to_be(True)
    expect(math.isinf(0.0)).to_be(False)
    expect(math.isinf(1.0)).to_be(False)
    expect(math.isinf(math.nan)).to_be(False)

def test_isfinite():
    expect(math.isfinite(0.0)).to_be(True)
    expect(math.isfinite(1.0)).to_be(True)
    expect(math.isfinite(-42.5)).to_be(True)
    expect(math.isfinite(math.inf)).to_be(False)
    expect(math.isfinite(math.nan)).to_be(False)

# === Square root ===
def test_sqrt():
    expect(abs(math.sqrt(4.0) - 2.0) < 0.0001).to_be(True)
    expect(abs(math.sqrt(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.sqrt(1.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.sqrt(2.0) - 1.41421356) < 0.0001).to_be(True)
    expect(abs(math.sqrt(100.0) - 10.0) < 0.0001).to_be(True)

def test_sqrt_negative():
    # sqrt of negative returns NaN
    result = math.sqrt(-1.0)
    expect(math.isnan(result)).to_be(True)

# === Power and exponential ===
def test_pow():
    expect(abs(math.pow(2.0, 3.0) - 8.0) < 0.0001).to_be(True)
    expect(abs(math.pow(2.0, 0.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.pow(2.0, -1.0) - 0.5) < 0.0001).to_be(True)
    expect(abs(math.pow(9.0, 0.5) - 3.0) < 0.0001).to_be(True)
    expect(abs(math.pow(0.0, 1.0) - 0.0) < 0.0001).to_be(True)

def test_exp():
    expect(abs(math.exp(0.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.exp(1.0) - math.e) < 0.0001).to_be(True)
    expect(abs(math.exp(2.0) - 7.38905609) < 0.0001).to_be(True)

# === Logarithms ===
def test_log():
    expect(abs(math.log(1.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.log(math.e) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.log(100.0) - 4.60517018) < 0.0001).to_be(True)

def test_log10():
    expect(abs(math.log10(1.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.log10(10.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.log10(100.0) - 2.0) < 0.0001).to_be(True)
    expect(abs(math.log10(1000.0) - 3.0) < 0.0001).to_be(True)

def test_log2():
    expect(abs(math.log2(1.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.log2(2.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.log2(8.0) - 3.0) < 0.0001).to_be(True)
    expect(abs(math.log2(1024.0) - 10.0) < 0.0001).to_be(True)

# === Trigonometric functions ===
def test_sin():
    expect(abs(math.sin(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.sin(math.pi / 2.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.sin(math.pi) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.sin(math.pi / 6.0) - 0.5) < 0.0001).to_be(True)

def test_cos():
    expect(abs(math.cos(0.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.cos(math.pi / 2.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.cos(math.pi) - (-1.0)) < 0.0001).to_be(True)
    expect(abs(math.cos(math.pi / 3.0) - 0.5) < 0.0001).to_be(True)

def test_tan():
    expect(abs(math.tan(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.tan(math.pi / 4.0) - 1.0) < 0.0001).to_be(True)

# === Inverse trigonometric functions ===
def test_asin_acos_atan():
    expect(abs(math.asin(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.asin(1.0) - math.pi / 2.0) < 0.0001).to_be(True)
    expect(abs(math.acos(1.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.acos(0.0) - math.pi / 2.0) < 0.0001).to_be(True)
    expect(abs(math.atan(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.atan(1.0) - math.pi / 4.0) < 0.0001).to_be(True)

def test_atan2():
    expect(abs(math.atan2(0.0, 1.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.atan2(1.0, 0.0) - math.pi / 2.0) < 0.0001).to_be(True)
    expect(abs(math.atan2(1.0, 1.0) - math.pi / 4.0) < 0.0001).to_be(True)

# === Hyperbolic functions ===
def test_hyperbolic():
    expect(abs(math.sinh(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.cosh(0.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.tanh(0.0) - 0.0) < 0.0001).to_be(True)
    # tanh approaches 1 for large values
    expect(abs(math.tanh(100.0) - 1.0) < 0.0001).to_be(True)

# === Rounding ===
def test_ceil():
    expect(math.ceil(0.5)).to_be(1.0)
    expect(math.ceil(1.0)).to_be(1.0)
    expect(math.ceil(-0.5)).to_be(0.0)
    expect(math.ceil(1.1)).to_be(2.0)
    expect(math.ceil(-1.1)).to_be(-1.0)

def test_floor():
    expect(math.floor(0.5)).to_be(0.0)
    expect(math.floor(1.0)).to_be(1.0)
    expect(math.floor(-0.5)).to_be(-1.0)
    expect(math.floor(1.9)).to_be(1.0)
    expect(math.floor(-1.1)).to_be(-2.0)

def test_trunc():
    expect(math.trunc(0.5)).to_be(0.0)
    expect(math.trunc(1.9)).to_be(1.0)
    expect(math.trunc(-0.5)).to_be(0.0)
    expect(math.trunc(-1.9)).to_be(-1.0)

# === fabs and copysign ===
def test_fabs():
    expect(abs(math.fabs(-1.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.fabs(1.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.fabs(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.fabs(-3.14) - 3.14) < 0.0001).to_be(True)

def test_copysign():
    expect(abs(math.copysign(1.0, -1.0) - (-1.0)) < 0.0001).to_be(True)
    expect(abs(math.copysign(-1.0, 1.0) - 1.0) < 0.0001).to_be(True)
    expect(abs(math.copysign(5.0, -3.0) - (-5.0)) < 0.0001).to_be(True)

# === Angular conversions ===
def test_degrees():
    expect(abs(math.degrees(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.degrees(math.pi) - 180.0) < 0.0001).to_be(True)
    expect(abs(math.degrees(math.pi / 2.0) - 90.0) < 0.0001).to_be(True)
    expect(abs(math.degrees(math.tau) - 360.0) < 0.0001).to_be(True)

def test_radians():
    expect(abs(math.radians(0.0) - 0.0) < 0.0001).to_be(True)
    expect(abs(math.radians(180.0) - math.pi) < 0.0001).to_be(True)
    expect(abs(math.radians(90.0) - math.pi / 2.0) < 0.0001).to_be(True)
    expect(abs(math.radians(360.0) - math.tau) < 0.0001).to_be(True)

# === Factorial ===
def test_factorial():
    expect(math.factorial(0)).to_be(1)
    expect(math.factorial(1)).to_be(1)
    expect(math.factorial(5)).to_be(120)
    expect(math.factorial(10)).to_be(3628800)

def test_factorial_negative():
    try:
        math.factorial(-1)
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

# === GCD ===
def test_gcd():
    expect(math.gcd(12, 8)).to_be(4)
    expect(math.gcd(7, 13)).to_be(1)
    expect(math.gcd(0, 5)).to_be(5)
    expect(math.gcd(5, 0)).to_be(5)
    expect(math.gcd(0, 0)).to_be(0)
    expect(math.gcd(-12, 8)).to_be(4)
    expect(math.gcd(12, -8)).to_be(4)
    expect(math.gcd(100, 75)).to_be(25)

# Register all tests
test("constants", test_constants)
test("nan_constant", test_nan_constant)
test("isnan", test_isnan)
test("isinf", test_isinf)
test("isfinite", test_isfinite)
test("sqrt", test_sqrt)
test("sqrt_negative", test_sqrt_negative)
test("pow", test_pow)
test("exp", test_exp)
test("log", test_log)
test("log10", test_log10)
test("log2", test_log2)
test("sin", test_sin)
test("cos", test_cos)
test("tan", test_tan)
test("asin_acos_atan", test_asin_acos_atan)
test("atan2", test_atan2)
test("hyperbolic", test_hyperbolic)
test("ceil", test_ceil)
test("floor", test_floor)
test("trunc", test_trunc)
test("fabs", test_fabs)
test("copysign", test_copysign)
test("degrees", test_degrees)
test("radians", test_radians)
test("factorial", test_factorial)
test("factorial_negative", test_factorial_negative)
test("gcd", test_gcd)

print("CPython math advanced tests completed")
