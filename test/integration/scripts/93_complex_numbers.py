# Test: Complex Numbers
# Tests complex number type, arithmetic, builtins, and attributes

from test_framework import test, expect

# === Literal creation ===
def test_imaginary_literal():
    z = 1j
    expect(z).to_be(1j)
    z2 = 2.5j
    expect(z2).to_be(2.5j)

def test_complex_expression():
    z = 1 + 2j
    expect(z).to_be(1+2j)
    z2 = 3 - 4j
    expect(z2).to_be(3-4j)

# === complex() builtin ===
def test_complex_builtin_zero_args():
    expect(complex()).to_be(0j)

def test_complex_builtin_one_arg():
    expect(complex(3)).to_be(3+0j)
    expect(complex(3.5)).to_be(3.5+0j)
    expect(complex(1+2j)).to_be(1+2j)

def test_complex_builtin_two_args():
    expect(complex(1, 2)).to_be(1+2j)
    expect(complex(0, 0)).to_be(0j)
    expect(complex(3.5, -1.5)).to_be(3.5-1.5j)

def test_complex_from_string():
    expect(complex("1+2j")).to_be(1+2j)
    expect(complex("3j")).to_be(3j)
    expect(complex("-1-2j")).to_be(-1-2j)
    expect(complex("5")).to_be(5+0j)
    expect(complex("  3+4j  ")).to_be(3+4j)

# === Arithmetic ===
def test_complex_add():
    expect((1+2j) + (3+4j)).to_be(4+6j)
    expect((1+2j) + 3).to_be(4+2j)
    expect(3 + (1+2j)).to_be(4+2j)
    expect((1+2j) + 1.5).to_be(2.5+2j)

def test_complex_sub():
    expect((5+3j) - (2+1j)).to_be(3+2j)
    expect((1+2j) - 1).to_be(0+2j)
    expect(5 - (1+2j)).to_be(4-2j)

def test_complex_mul():
    expect((1+2j) * (3+4j)).to_be(-5+10j)
    expect((2+3j) * 2).to_be(4+6j)
    expect(2 * (2+3j)).to_be(4+6j)

def test_complex_div():
    expect((4+2j) / (1+1j)).to_be(3-1j)
    expect((10+0j) / 2).to_be(5+0j)

def test_complex_power():
    # (1j) ** 2 == -1 (with floating point tolerance)
    result = 1j ** 2
    expect(result.real).to_be(-1.0)
    expect(abs(result.imag) < 1e-10).to_be(True)

def test_complex_unary():
    z = 3 + 4j
    expect(-z).to_be(-3-4j)
    expect(+z).to_be(3+4j)

# === abs() ===
def test_complex_abs():
    expect(abs(3+4j)).to_be(5.0)
    expect(abs(0j)).to_be(0.0)

# === Attributes ===
def test_complex_real_imag():
    z = 3 + 4j
    expect(z.real).to_be(3.0)
    expect(z.imag).to_be(4.0)

def test_complex_conjugate():
    z = 3 + 4j
    expect(z.conjugate()).to_be(3-4j)

# === Equality ===
def test_complex_equality():
    expect(1+2j == 1+2j).to_be(True)
    expect(1+2j != 3+4j).to_be(True)
    # Cross-type equality
    expect(1+0j == 1).to_be(True)
    expect(1+0j == 1.0).to_be(True)
    expect(1+1j == 1).to_be(False)

# === Ordering raises TypeError ===
def test_complex_ordering_error():
    try:
        result = (1+2j) < (3+4j)
        expect(False).to_be(True)  # Should not reach here
    except TypeError:
        expect(True).to_be(True)

# === Truthiness ===
def test_complex_truthiness():
    expect(bool(0j)).to_be(False)
    expect(bool(1j)).to_be(True)
    expect(bool(1+0j)).to_be(True)
    expect(bool(0+0j)).to_be(False)

# === isinstance and type ===
def test_complex_isinstance():
    expect(isinstance(1j, complex)).to_be(True)
    expect(isinstance(1+2j, complex)).to_be(True)
    expect(isinstance(1, complex)).to_be(False)

def test_complex_type():
    t = type(1j)
    expect(t.__name__).to_be("complex")

# === Hash ===
def test_complex_hash():
    # Complex with zero imag should hash equal to the equivalent real number
    expect(hash(1+0j) == hash(1)).to_be(True)
    expect(hash(2+0j) == hash(2)).to_be(True)
    # Complex as dict key
    d = {1+2j: "hello"}
    expect(d[1+2j]).to_be("hello")

# === str/repr formatting ===
def test_complex_str_repr():
    expect(str(1+2j)).to_be("(1+2j)")
    expect(str(1-2j)).to_be("(1-2j)")
    expect(str(3j)).to_be("3j")
    expect(str(-3j)).to_be("(-0-3j)")
    expect(repr(1+2j)).to_be("(1+2j)")
    expect(str(0j)).to_be("0j")

# === int()/float() on complex raises TypeError ===
def test_complex_to_int_error():
    try:
        int(1+2j)
        expect(False).to_be(True)  # Should not reach here
    except TypeError:
        expect(True).to_be(True)

def test_complex_to_float_error():
    try:
        float(1+2j)
        expect(False).to_be(True)  # Should not reach here
    except TypeError:
        expect(True).to_be(True)

# === Floor div and mod raise TypeError ===
def test_complex_floordiv_error():
    try:
        (1+2j) // (3+4j)
        expect(False).to_be(True)
    except TypeError:
        expect(True).to_be(True)

def test_complex_mod_error():
    try:
        (1+2j) % (3+4j)
        expect(False).to_be(True)
    except TypeError:
        expect(True).to_be(True)

test("imaginary literal", test_imaginary_literal)
test("complex expression", test_complex_expression)
test("complex() zero args", test_complex_builtin_zero_args)
test("complex() one arg", test_complex_builtin_one_arg)
test("complex() two args", test_complex_builtin_two_args)
test("complex() from string", test_complex_from_string)
test("complex addition", test_complex_add)
test("complex subtraction", test_complex_sub)
test("complex multiplication", test_complex_mul)
test("complex division", test_complex_div)
test("complex power", test_complex_power)
test("complex unary ops", test_complex_unary)
test("complex abs()", test_complex_abs)
test("complex .real/.imag", test_complex_real_imag)
test("complex .conjugate()", test_complex_conjugate)
test("complex equality", test_complex_equality)
test("complex ordering error", test_complex_ordering_error)
test("complex truthiness", test_complex_truthiness)
test("complex isinstance", test_complex_isinstance)
test("complex type()", test_complex_type)
test("complex hash", test_complex_hash)
test("complex str/repr", test_complex_str_repr)
test("complex to int error", test_complex_to_int_error)
test("complex to float error", test_complex_to_float_error)
test("complex floordiv error", test_complex_floordiv_error)
test("complex mod error", test_complex_mod_error)
