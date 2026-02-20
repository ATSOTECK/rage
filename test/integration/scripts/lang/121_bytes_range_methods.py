from test_framework import test, expect

# =====================
# bytes methods
# =====================

# === removeprefix ===
test("bytes removeprefix match", lambda: expect(b"HelloWorld".removeprefix(b"Hello")).to_be(b"World"))
test("bytes removeprefix no match", lambda: expect(b"HelloWorld".removeprefix(b"World")).to_be(b"HelloWorld"))
test("bytes removeprefix empty", lambda: expect(b"Hello".removeprefix(b"")).to_be(b"Hello"))
test("bytes removeprefix full", lambda: expect(b"Hello".removeprefix(b"Hello")).to_be(b""))

# === removesuffix ===
test("bytes removesuffix match", lambda: expect(b"HelloWorld".removesuffix(b"World")).to_be(b"Hello"))
test("bytes removesuffix no match", lambda: expect(b"HelloWorld".removesuffix(b"Hello")).to_be(b"HelloWorld"))
test("bytes removesuffix empty", lambda: expect(b"Hello".removesuffix(b"")).to_be(b"Hello"))
test("bytes removesuffix full", lambda: expect(b"Hello".removesuffix(b"Hello")).to_be(b""))

# === isascii ===
test("bytes isascii all ascii", lambda: expect(b"hello".isascii()).to_be(True))
test("bytes isascii empty", lambda: expect(b"".isascii()).to_be(True))
test("bytes isascii with high byte", lambda: expect(b"\xff".isascii()).to_be(False))
test("bytes isascii mixed", lambda: expect(b"\x00\x7f".isascii()).to_be(True))

# === istitle ===
test("bytes istitle yes", lambda: expect(b"Hello World".istitle()).to_be(True))
test("bytes istitle no", lambda: expect(b"hello world".istitle()).to_be(False))
test("bytes istitle all upper", lambda: expect(b"HELLO".istitle()).to_be(False))
test("bytes istitle empty", lambda: expect(b"".istitle()).to_be(False))
test("bytes istitle single", lambda: expect(b"Hello".istitle()).to_be(True))
test("bytes istitle with number", lambda: expect(b"Hello 123 World".istitle()).to_be(True))

# === maketrans ===
test("bytes maketrans basic", lambda: expect(b"hello".translate(bytes.maketrans(b"helo", b"HELO"))).to_be(b"HELLO"))

# === translate ===
test("bytes translate basic", lambda: expect(b"abc".translate(bytes.maketrans(b"abc", b"xyz"))).to_be(b"xyz"))
test("bytes translate no match", lambda: expect(b"hello".translate(bytes.maketrans(b"xyz", b"XYZ"))).to_be(b"hello"))

# translate with delete (second arg)
test("bytes translate delete", lambda: expect(b"hello world".translate(None, b" ")).to_be(b"helloworld"))
test("bytes translate map and delete", lambda: expect(b"hello!".translate(bytes.maketrans(b"h", b"H"), b"!")).to_be(b"Hello"))

# === Error cases ===
def test_bytes_removeprefix_type():
    try:
        b"hello".removeprefix("hello")
        return False
    except TypeError:
        return True

test("bytes removeprefix type error", lambda: expect(test_bytes_removeprefix_type()).to_be(True))

def test_bytes_maketrans_unequal():
    try:
        bytes.maketrans(b"ab", b"xyz")
        return False
    except ValueError:
        return True

test("bytes maketrans unequal length", lambda: expect(test_bytes_maketrans_unequal()).to_be(True))

# =====================
# range methods
# =====================

# === Properties ===
test("range start", lambda: expect(range(1, 10, 2).start).to_be(1))
test("range stop", lambda: expect(range(1, 10, 2).stop).to_be(10))
test("range step", lambda: expect(range(1, 10, 2).step).to_be(2))
test("range start default", lambda: expect(range(5).start).to_be(0))
test("range stop default", lambda: expect(range(5).stop).to_be(5))
test("range step default", lambda: expect(range(5).step).to_be(1))
test("range negative step", lambda: expect(range(10, 0, -1).step).to_be(-1))

# === count ===
test("range count present", lambda: expect(range(10).count(5)).to_be(1))
test("range count absent", lambda: expect(range(10).count(10)).to_be(0))
test("range count negative absent", lambda: expect(range(10).count(-1)).to_be(0))
test("range count step", lambda: expect(range(0, 10, 2).count(4)).to_be(1))
test("range count step miss", lambda: expect(range(0, 10, 2).count(3)).to_be(0))
test("range count non-int", lambda: expect(range(10).count("hello")).to_be(0))

# === index ===
test("range index first", lambda: expect(range(10).index(0)).to_be(0))
test("range index middle", lambda: expect(range(10).index(5)).to_be(5))
test("range index last", lambda: expect(range(10).index(9)).to_be(9))
test("range index step", lambda: expect(range(0, 10, 2).index(6)).to_be(3))
test("range index negative step", lambda: expect(range(10, 0, -2).index(8)).to_be(1))

def test_range_index_error():
    try:
        range(10).index(10)
        return False
    except ValueError:
        return True

test("range index not found", lambda: expect(test_range_index_error()).to_be(True))

def test_range_index_non_int():
    try:
        range(10).index("hello")
        return False
    except ValueError:
        return True

test("range index non-int error", lambda: expect(test_range_index_non_int()).to_be(True))

# === range attribute error ===
def test_range_no_attr():
    try:
        range(10).nonexistent
        return False
    except AttributeError:
        return True

test("range no attribute", lambda: expect(test_range_no_attr()).to_be(True))
