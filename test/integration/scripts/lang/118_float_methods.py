from test_framework import test, expect

# === is_integer ===
test("is_integer 1.0", lambda: expect((1.0).is_integer()).to_be(True))
test("is_integer 1.5", lambda: expect((1.5).is_integer()).to_be(False))
test("is_integer 0.0", lambda: expect((0.0).is_integer()).to_be(True))
test("is_integer -2.0", lambda: expect((-2.0).is_integer()).to_be(True))
test("is_integer -2.5", lambda: expect((-2.5).is_integer()).to_be(False))

# === conjugate ===
test("conjugate 3.14", lambda: expect((3.14).conjugate()).to_be(3.14))
test("conjugate -1.5", lambda: expect((-1.5).conjugate()).to_be(-1.5))

# === Properties ===
test("real property", lambda: expect((3.14).real).to_be(3.14))
test("imag property", lambda: expect((3.14).imag).to_be(0.0))
test("negative real", lambda: expect((-2.5).real).to_be(-2.5))
test("negative imag", lambda: expect((-2.5).imag).to_be(0.0))

# === hex ===
test("hex 0.0", lambda: expect((0.0).hex()).to_be("0x0.0000000000000p+0"))
test("hex 1.0", lambda: expect((1.0).hex()).to_be("0x1.0000000000000p+0"))
test("hex -1.0", lambda: expect((-1.0).hex()).to_be("-0x1.0000000000000p+0"))
test("hex 1.5", lambda: expect((1.5).hex()).to_be("0x1.8000000000000p+0"))
test("hex 12.5", lambda: expect((12.5).hex()).to_be("0x1.9000000000000p+3"))

# === fromhex (class method) ===
test("float.fromhex 0x0p+0", lambda: expect(float.fromhex("0x0.0p+0")).to_be(0.0))
test("float.fromhex 0x1.0p+0", lambda: expect(float.fromhex("0x1.0p+0")).to_be(1.0))
test("float.fromhex 0x1.8p+0", lambda: expect(float.fromhex("0x1.8p+0")).to_be(1.5))
test("float.fromhex 0x1.9p+3", lambda: expect(float.fromhex("0x1.9p+3")).to_be(12.5))
test("float.fromhex -0x1.0p+0", lambda: expect(float.fromhex("-0x1.0p+0")).to_be(-1.0))

# Instance fromhex also works
test("instance fromhex", lambda: expect((0.0).fromhex("0x1.0p+0")).to_be(1.0))

# fromhex with whitespace
test("fromhex whitespace", lambda: expect(float.fromhex("  0x1.0p+0  ")).to_be(1.0))

# === as_integer_ratio ===
test("as_integer_ratio 0.5", lambda: expect((0.5).as_integer_ratio()).to_be((1, 2)))
test("as_integer_ratio 2.5", lambda: expect((2.5).as_integer_ratio()).to_be((5, 2)))
test("as_integer_ratio 1.0", lambda: expect((1.0).as_integer_ratio()).to_be((1, 1)))
test("as_integer_ratio 0.0", lambda: expect((0.0).as_integer_ratio()).to_be((0, 1)))
test("as_integer_ratio -0.5", lambda: expect((-0.5).as_integer_ratio()).to_be((-1, 2)))
test("as_integer_ratio 10.0", lambda: expect((10.0).as_integer_ratio()).to_be((10, 1)))
test("as_integer_ratio 0.25", lambda: expect((0.25).as_integer_ratio()).to_be((1, 4)))
test("as_integer_ratio 0.75", lambda: expect((0.75).as_integer_ratio()).to_be((3, 4)))

# as_integer_ratio for non-simple fractions (0.1 is not exactly representable)
def test_ratio_0_1():
    n, d = (0.1).as_integer_ratio()
    return abs(n / d - 0.1) < 1e-15

test("as_integer_ratio 0.1 roundtrip", lambda: expect(test_ratio_0_1()).to_be(True))

# Error cases for as_integer_ratio
def test_ratio_inf():
    try:
        x = 1e309  # infinity
        x.as_integer_ratio()
        return False
    except OverflowError:
        return True

test("as_integer_ratio inf raises OverflowError", lambda: expect(test_ratio_inf()).to_be(True))

# === Dunder methods ===
test("__abs__ positive", lambda: expect((3.14).__abs__()).to_be(3.14))
test("__abs__ negative", lambda: expect((-3.14).__abs__()).to_be(3.14))
test("__abs__ zero", lambda: expect((0.0).__abs__()).to_be(0.0))

test("__bool__ nonzero", lambda: expect((1.0).__bool__()).to_be(True))
test("__bool__ zero", lambda: expect((0.0).__bool__()).to_be(False))
test("__bool__ negative", lambda: expect((-1.0).__bool__()).to_be(True))

test("__ceil__ 1.1", lambda: expect((1.1).__ceil__()).to_be(2))
test("__ceil__ 1.9", lambda: expect((1.9).__ceil__()).to_be(2))
test("__ceil__ -1.1", lambda: expect((-1.1).__ceil__()).to_be(-1))
test("__ceil__ 2.0", lambda: expect((2.0).__ceil__()).to_be(2))

test("__floor__ 1.1", lambda: expect((1.1).__floor__()).to_be(1))
test("__floor__ 1.9", lambda: expect((1.9).__floor__()).to_be(1))
test("__floor__ -1.1", lambda: expect((-1.1).__floor__()).to_be(-2))
test("__floor__ 2.0", lambda: expect((2.0).__floor__()).to_be(2))

test("__trunc__ 1.9", lambda: expect((1.9).__trunc__()).to_be(1))
test("__trunc__ -1.9", lambda: expect((-1.9).__trunc__()).to_be(-1))
test("__trunc__ 2.0", lambda: expect((2.0).__trunc__()).to_be(2))

test("__int__ 3.7", lambda: expect((3.7).__int__()).to_be(3))
test("__int__ -3.7", lambda: expect((-3.7).__int__()).to_be(-3))

test("__float__ identity", lambda: expect((3.14).__float__()).to_be(3.14))

# === __round__ ===
# No args: return int with banker's rounding
test("__round__ no args 2.5", lambda: expect((2.5).__round__()).to_be(2))
test("__round__ no args 3.5", lambda: expect((3.5).__round__()).to_be(4))
test("__round__ no args 1.5", lambda: expect((1.5).__round__()).to_be(2))
test("__round__ no args 0.5", lambda: expect((0.5).__round__()).to_be(0))
test("__round__ no args 4.5", lambda: expect((4.5).__round__()).to_be(4))
test("__round__ no args 2.7", lambda: expect((2.7).__round__()).to_be(3))
test("__round__ no args -0.5", lambda: expect((-0.5).__round__()).to_be(0))

# With ndigits: return float
test("__round__ ndigits=1", lambda: expect((2.75).__round__(1)).to_be(2.8))
test("__round__ ndigits=0", lambda: expect((2.5).__round__(0)).to_be(2.0))
test("__round__ ndigits=2", lambda: expect((1.005).__round__(2)).to_be(1.0))  # floating point...
test("__round__ ndigits=-1", lambda: expect((25.0).__round__(-1)).to_be(20.0))

# None arg acts like no args
test("__round__ None", lambda: expect((2.5).__round__(None)).to_be(2))

# === Error attribute ===
def test_no_attr():
    try:
        (1.0).nonexistent
        return False
    except AttributeError:
        return True

test("float no attribute raises AttributeError", lambda: expect(test_no_attr()).to_be(True))
