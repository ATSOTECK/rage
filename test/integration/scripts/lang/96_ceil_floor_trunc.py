import math
from test_framework import test, expect

# Test 1: Basic __ceil__
def test_ceil_dunder(t):
    class Measurement:
        def __init__(self, val):
            self.val = val
        def __ceil__(self):
            return math.ceil(self.val)

    m = Measurement(3.2)
    expect(math.ceil(m)).to_be(4)

test("__ceil__ on custom class", test_ceil_dunder)

# Test 2: Basic __floor__
def test_floor_dunder(t):
    class Measurement:
        def __init__(self, val):
            self.val = val
        def __floor__(self):
            return math.floor(self.val)

    m = Measurement(3.7)
    expect(math.floor(m)).to_be(3)

test("__floor__ on custom class", test_floor_dunder)

# Test 3: Basic __trunc__
def test_trunc_dunder(t):
    class Measurement:
        def __init__(self, val):
            self.val = val
        def __trunc__(self):
            return math.trunc(self.val)

    m = Measurement(3.7)
    expect(math.trunc(m)).to_be(3)

test("__trunc__ on custom class", test_trunc_dunder)

# Test 4: __ceil__ with negative value
def test_ceil_negative(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __ceil__(self):
            return math.ceil(self.val)

    n = Num(-2.3)
    expect(math.ceil(n)).to_be(-2)

test("__ceil__ with negative value", test_ceil_negative)

# Test 5: __floor__ with negative value
def test_floor_negative(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __floor__(self):
            return math.floor(self.val)

    n = Num(-2.3)
    expect(math.floor(n)).to_be(-3)

test("__floor__ with negative value", test_floor_negative)

# Test 6: __trunc__ with negative value
def test_trunc_negative(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __trunc__(self):
            return math.trunc(self.val)

    n = Num(-2.7)
    expect(math.trunc(n)).to_be(-2)

test("__trunc__ with negative value", test_trunc_negative)

# Test 7: Inherited __ceil__
def test_ceil_inherited(t):
    class Base:
        def __init__(self, val):
            self.val = val
        def __ceil__(self):
            return math.ceil(self.val)

    class Child(Base):
        pass

    c = Child(1.1)
    expect(math.ceil(c)).to_be(2)

test("__ceil__ inherited from base class", test_ceil_inherited)

# Test 8: math.ceil on regular float still works
def test_ceil_float(t):
    expect(math.ceil(3.2)).to_be(4)
    expect(math.ceil(3.0)).to_be(3)
    expect(math.ceil(-1.5)).to_be(-1)

test("math.ceil on regular float", test_ceil_float)

# Test 9: math.floor on regular float still works
def test_floor_float(t):
    expect(math.floor(3.7)).to_be(3)
    expect(math.floor(3.0)).to_be(3)
    expect(math.floor(-1.5)).to_be(-2)

test("math.floor on regular float", test_floor_float)

# Test 10: math.trunc on regular float still works
def test_trunc_float(t):
    expect(math.trunc(3.7)).to_be(3)
    expect(math.trunc(-3.7)).to_be(-3)

test("math.trunc on regular float", test_trunc_float)

# Test 11: math.ceil on int still works
def test_ceil_int(t):
    expect(math.ceil(5)).to_be(5)

test("math.ceil on int", test_ceil_int)

# Test 12: __ceil__/__floor__/__trunc__ without dunder raises TypeError
def test_no_dunder_raises(t):
    class NoMath:
        pass

    n = NoMath()

    try:
        math.ceil(n)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

    try:
        math.floor(n)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

    try:
        math.trunc(n)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("no dunder raises TypeError", test_no_dunder_raises)

# Test 13: Custom rounding logic
def test_custom_rounding(t):
    class AlwaysUp:
        def __init__(self, val):
            self.val = val
        def __ceil__(self):
            return int(self.val) + 1
        def __floor__(self):
            return int(self.val)
        def __trunc__(self):
            return int(self.val)

    a = AlwaysUp(5.0)
    expect(math.ceil(a)).to_be(6)
    expect(math.floor(a)).to_be(5)
    expect(math.trunc(a)).to_be(5)

test("custom rounding logic", test_custom_rounding)
