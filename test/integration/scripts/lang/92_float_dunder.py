from test_framework import test, expect

# Test 1: Basic __float__ conversion
def test_basic_float(t):
    class Temperature:
        def __init__(self, celsius):
            self.celsius = celsius
        def __float__(self):
            return float(self.celsius)

    temp = Temperature(36.6)
    expect(float(temp)).to_be(36.6)

test("basic __float__ conversion", test_basic_float)

# Test 2: __float__ returning int converted to float
def test_float_from_int(t):
    class Whole:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    w = Whole(42)
    expect(float(w)).to_be(42.0)

test("__float__ from int value", test_float_from_int)

# Test 3: __float__ used in arithmetic
def test_float_in_arithmetic(t):
    class Weight:
        def __init__(self, kg):
            self.kg = kg
        def __float__(self):
            return float(self.kg)

    w = Weight(75)
    result = float(w) + 0.5
    expect(result).to_be(75.5)

test("__float__ used in arithmetic", test_float_in_arithmetic)

# Test 4: __float__ with negative value
def test_float_negative(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    n = Num(-3.14)
    expect(float(n)).to_be(-3.14)

test("__float__ with negative value", test_float_negative)

# Test 5: __float__ inherited from base class
def test_float_inherited(t):
    class Base:
        def __init__(self, val):
            self.val = val
        def __float__(self):
            return float(self.val)

    class Child(Base):
        pass

    c = Child(2.718)
    expect(float(c)).to_be(2.718)

test("__float__ inherited from base class", test_float_inherited)

# Test 6: __float__ with zero
def test_float_zero(t):
    class Zero:
        def __float__(self):
            return 0.0

    z = Zero()
    expect(float(z)).to_be(0.0)

test("__float__ with zero", test_float_zero)

# Test 7: float() without __float__ raises TypeError
def test_float_no_dunder(t):
    class NoFloat:
        pass

    n = NoFloat()
    try:
        float(n)
        expect(True).to_be(False)  # should not reach here
    except TypeError:
        expect(True).to_be(True)

test("float() without __float__ raises TypeError", test_float_no_dunder)

# Test 8: __float__ returning non-float raises TypeError
def test_float_returns_non_float(t):
    class Bad:
        def __float__(self):
            return "not a float"

    b = Bad()
    try:
        float(b)
        expect(True).to_be(False)  # should not reach here
    except TypeError:
        expect(True).to_be(True)

test("__float__ returning non-float raises TypeError", test_float_returns_non_float)
