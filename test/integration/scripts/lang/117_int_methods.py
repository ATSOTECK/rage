from test_framework import test, expect

# === bit_length ===
test("bit_length(0)", lambda: expect((0).bit_length()).to_be(0))
test("bit_length(1)", lambda: expect((1).bit_length()).to_be(1))
test("bit_length(255)", lambda: expect((255).bit_length()).to_be(8))
test("bit_length(256)", lambda: expect((256).bit_length()).to_be(9))
test("bit_length(-1)", lambda: expect((-1).bit_length()).to_be(1))
test("bit_length(-255)", lambda: expect((-255).bit_length()).to_be(8))

# === bit_count ===
test("bit_count(0)", lambda: expect((0).bit_count()).to_be(0))
test("bit_count(1)", lambda: expect((1).bit_count()).to_be(1))
test("bit_count(255)", lambda: expect((255).bit_count()).to_be(8))
test("bit_count(256)", lambda: expect((256).bit_count()).to_be(1))
test("bit_count(-1)", lambda: expect((-1).bit_count()).to_be(1))
test("bit_count(7)", lambda: expect((7).bit_count()).to_be(3))

# === conjugate ===
test("conjugate(42)", lambda: expect((42).conjugate()).to_be(42))
test("conjugate(-5)", lambda: expect((-5).conjugate()).to_be(-5))

# === as_integer_ratio ===
test("as_integer_ratio(42)", lambda: expect((42).as_integer_ratio()).to_be((42, 1)))
test("as_integer_ratio(0)", lambda: expect((0).as_integer_ratio()).to_be((0, 1)))

# === Properties ===
test("real property", lambda: expect((42).real).to_be(42))
test("imag property", lambda: expect((42).imag).to_be(0))
test("numerator property", lambda: expect((42).numerator).to_be(42))
test("denominator property", lambda: expect((42).denominator).to_be(1))
test("negative real", lambda: expect((-7).real).to_be(-7))
test("negative imag", lambda: expect((-7).imag).to_be(0))

# === to_bytes ===
test("to_bytes big-endian", lambda: expect((1024).to_bytes(2, "big")).to_be(b'\x04\x00'))
test("to_bytes little-endian", lambda: expect((1024).to_bytes(2, "little")).to_be(b'\x00\x04'))
test("to_bytes zero", lambda: expect((0).to_bytes(1, "big")).to_be(b'\x00'))
test("to_bytes 255 big", lambda: expect((255).to_bytes(1, "big")).to_be(b'\xff'))
test("to_bytes 256 big", lambda: expect((256).to_bytes(2, "big")).to_be(b'\x01\x00'))

# Signed negative
test("to_bytes signed -1 big", lambda: expect((-1).to_bytes(1, "big", signed=True)).to_be(b'\xff'))
test("to_bytes signed -1 little", lambda: expect((-1).to_bytes(1, "little", signed=True)).to_be(b'\xff'))
test("to_bytes signed -128 big", lambda: expect((-128).to_bytes(1, "big", signed=True)).to_be(b'\x80'))

# Overflow error for unsigned negative
def test_to_bytes_overflow():
    try:
        (-1).to_bytes(1, "big")
        return False
    except OverflowError:
        return True

test("to_bytes unsigned negative raises OverflowError", lambda: expect(test_to_bytes_overflow()).to_be(True))

# Overflow error when number doesn't fit
def test_to_bytes_too_small():
    try:
        (256).to_bytes(1, "big")
        return False
    except OverflowError:
        return True

test("to_bytes overflow raises OverflowError", lambda: expect(test_to_bytes_too_small()).to_be(True))

# === from_bytes (instance method) ===
test("from_bytes big-endian", lambda: expect(int.from_bytes(b'\x04\x00', "big")).to_be(1024))
test("from_bytes little-endian", lambda: expect(int.from_bytes(b'\x00\x04', "little")).to_be(1024))
test("from_bytes zero", lambda: expect(int.from_bytes(b'\x00', "big")).to_be(0))
test("from_bytes 255", lambda: expect(int.from_bytes(b'\xff', "big")).to_be(255))

# Signed from_bytes
test("from_bytes signed -1", lambda: expect(int.from_bytes(b'\xff', "big", signed=True)).to_be(-1))
test("from_bytes signed -128", lambda: expect(int.from_bytes(b'\x80', "big", signed=True)).to_be(-128))
test("from_bytes signed 127", lambda: expect(int.from_bytes(b'\x7f', "big", signed=True)).to_be(127))

# Instance method from_bytes (via an int instance)
test("instance from_bytes", lambda: expect((0).from_bytes(b'\x04\x00', "big")).to_be(1024))

# === Dunder methods ===
test("__abs__ positive", lambda: expect((42).__abs__()).to_be(42))
test("__abs__ negative", lambda: expect((-42).__abs__()).to_be(42))
test("__abs__ zero", lambda: expect((0).__abs__()).to_be(0))

test("__bool__ nonzero", lambda: expect((42).__bool__()).to_be(True))
test("__bool__ zero", lambda: expect((0).__bool__()).to_be(False))

test("__ceil__", lambda: expect((42).__ceil__()).to_be(42))
test("__floor__", lambda: expect((42).__floor__()).to_be(42))
test("__trunc__", lambda: expect((42).__trunc__()).to_be(42))

test("__int__", lambda: expect((42).__int__()).to_be(42))
test("__float__", lambda: expect((42).__float__()).to_be(42.0))
test("__index__", lambda: expect((42).__index__()).to_be(42))

# === __round__ ===
test("__round__ no args", lambda: expect((42).__round__()).to_be(42))
test("__round__ None", lambda: expect((42).__round__(None)).to_be(42))
test("__round__ ndigits=0", lambda: expect((42).__round__(0)).to_be(42))
test("__round__ ndigits=1", lambda: expect((42).__round__(1)).to_be(42))

# Negative ndigits - round to nearest 10, 100, etc.
test("round -1 ndigits 25", lambda: expect((25).__round__(-1)).to_be(20))
test("round -1 ndigits 35", lambda: expect((35).__round__(-1)).to_be(40))
test("round -1 ndigits 15", lambda: expect((15).__round__(-1)).to_be(20))  # banker's: round to even
test("round -1 ndigits 45", lambda: expect((45).__round__(-1)).to_be(40))  # banker's: round to even
test("round -2 ndigits 150", lambda: expect((150).__round__(-2)).to_be(200))  # banker's: round to even
test("round -2 ndigits 250", lambda: expect((250).__round__(-2)).to_be(200))  # banker's: round to even
test("round -1 ndigits 100", lambda: expect((100).__round__(-1)).to_be(100))
test("round -2 ndigits 100", lambda: expect((100).__round__(-2)).to_be(100))

# Negative number rounding
test("round -1 negative -25", lambda: expect((-25).__round__(-1)).to_be(-20))
test("round -1 negative -35", lambda: expect((-35).__round__(-1)).to_be(-40))
