from test_framework import test, expect

# Test 1: Basic __complex__ conversion
def test_basic_complex(t):
    class Impedance:
        def __init__(self, r, x):
            self.r = r
            self.x = x
        def __complex__(self):
            return complex(self.r, self.x)

    z = Impedance(3, 4)
    c = complex(z)
    expect(c).to_be((3+4j))

test("basic __complex__ conversion", test_basic_complex)

# Test 2: __complex__ with pure real
def test_complex_pure_real(t):
    class RealOnly:
        def __init__(self, val):
            self.val = val
        def __complex__(self):
            return complex(self.val, 0)

    r = RealOnly(5)
    c = complex(r)
    expect(c).to_be((5+0j))

test("__complex__ with pure real", test_complex_pure_real)

# Test 3: __complex__ with pure imaginary
def test_complex_pure_imag(t):
    class ImagOnly:
        def __init__(self, val):
            self.val = val
        def __complex__(self):
            return complex(0, self.val)

    i = ImagOnly(7)
    c = complex(i)
    expect(c).to_be(7j)

test("__complex__ with pure imaginary", test_complex_pure_imag)

# Test 4: __complex__ with negative values
def test_complex_negative(t):
    class Num:
        def __complex__(self):
            return complex(-1, -2)

    n = Num()
    c = complex(n)
    expect(c).to_be((-1-2j))

test("__complex__ with negative values", test_complex_negative)

# Test 5: __complex__ inherited from base class
def test_complex_inherited(t):
    class Base:
        def __init__(self, r, i):
            self.r = r
            self.i = i
        def __complex__(self):
            return complex(self.r, self.i)

    class Child(Base):
        pass

    c = Child(1, 2)
    result = complex(c)
    expect(result).to_be((1+2j))

test("__complex__ inherited from base class", test_complex_inherited)

# Test 6: complex() without __complex__ raises TypeError
def test_complex_no_dunder(t):
    class NoComplex:
        pass

    n = NoComplex()
    try:
        complex(n)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("complex() without __complex__ raises TypeError", test_complex_no_dunder)

# Test 7: __complex__ returning non-complex raises TypeError
def test_complex_returns_non_complex(t):
    class Bad:
        def __complex__(self):
            return 42

    b = Bad()
    try:
        complex(b)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("__complex__ returning non-complex raises TypeError", test_complex_returns_non_complex)

# Test 8: complex() two-arg with __float__ on first arg
def test_complex_two_arg_float_real(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    n = Num(3)
    c = complex(n, 4)
    expect(c).to_be((3+4j))

test("complex() two-arg with __float__ on real part", test_complex_two_arg_float_real)

# Test 9: complex() two-arg with __float__ on second arg
def test_complex_two_arg_float_imag(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    n = Num(4)
    c = complex(3, n)
    expect(c).to_be((3+4j))

test("complex() two-arg with __float__ on imag part", test_complex_two_arg_float_imag)

# Test 10: complex() two-arg with __float__ on both args
def test_complex_two_arg_float_both(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    r = Num(1)
    i = Num(2)
    c = complex(r, i)
    expect(c).to_be((1+2j))

test("complex() two-arg with __float__ on both args", test_complex_two_arg_float_both)
