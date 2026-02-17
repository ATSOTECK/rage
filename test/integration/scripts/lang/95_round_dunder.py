from test_framework import test, expect

# Test 1: Basic __round__ without ndigits
def test_basic_round(t):
    class Measurement:
        def __init__(self, val):
            self.val = val
        def __round__(self):
            return round(self.val)

    m = Measurement(3.7)
    expect(round(m)).to_be(4)

test("basic __round__ without ndigits", test_basic_round)

# Test 2: __round__ with ndigits
def test_round_ndigits(t):
    class Measurement:
        def __init__(self, val):
            self.val = val
        def __round__(self, ndigits=0):
            return round(self.val, ndigits)

    m = Measurement(3.14159)
    expect(round(m, 2)).to_be(3.14)

test("__round__ with ndigits", test_round_ndigits)

# Test 3: __round__ returning int
def test_round_returns_int(t):
    class Score:
        def __init__(self, val):
            self.val = val
        def __round__(self):
            if self.val >= 0:
                return int(self.val + 0.5)
            return int(self.val - 0.5)

    s = Score(7.6)
    expect(round(s)).to_be(8)

test("__round__ returning int", test_round_returns_int)

# Test 4: __round__ with negative value
def test_round_negative(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __round__(self):
            return round(self.val)

    n = Num(-2.3)
    expect(round(n)).to_be(-2)

test("__round__ with negative value", test_round_negative)

# Test 5: __round__ inherited from base class
def test_round_inherited(t):
    class Base:
        def __init__(self, val):
            self.val = val
        def __round__(self):
            return round(self.val)

    class Child(Base):
        pass

    c = Child(5.5)
    expect(round(c)).to_be(6)

test("__round__ inherited from base class", test_round_inherited)

# Test 6: round() on regular float still works
def test_round_float(t):
    expect(round(3.7)).to_be(4)
    expect(round(3.14159, 2)).to_be(3.14)
    expect(round(2.5)).to_be(2)  # banker's rounding

test("round() on regular float still works", test_round_float)

# Test 7: round() on regular int still works
def test_round_int(t):
    expect(round(5)).to_be(5)

test("round() on regular int still works", test_round_int)

# Test 8: __round__ without method raises TypeError
def test_round_no_dunder(t):
    class NoRound:
        pass

    n = NoRound()
    try:
        round(n)
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("round() without __round__ raises TypeError", test_round_no_dunder)

# Test 9: __round__ with ndigits=0
def test_round_ndigits_zero(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __round__(self, ndigits=0):
            return round(self.val, ndigits)

    n = Num(3.14159)
    expect(round(n, 0)).to_be(3.0)

test("__round__ with ndigits=0", test_round_ndigits_zero)

# Test 10: __round__ with negative ndigits
def test_round_negative_ndigits(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __round__(self, ndigits=0):
            return round(self.val, ndigits)

    n = Num(1234.5)
    expect(round(n, -2)).to_be(1200)

test("__round__ with negative ndigits", test_round_negative_ndigits)
