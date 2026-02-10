# Test: CPython Builtin Function Edge Cases
# Adapted from CPython's test_builtin.py

from test_framework import test, expect

# === abs ===
def test_abs():
    expect(abs(0)).to_be(0)
    expect(abs(-1)).to_be(1)
    expect(abs(1)).to_be(1)
    expect(abs(-1234)).to_be(1234)
    expect(abs(-3.14)).to_be(3.14)
    expect(abs(3.14)).to_be(3.14)
    expect(abs(0.0)).to_be(0.0)

# === all ===
def test_all():
    expect(all([])).to_be(True)
    expect(all([True])).to_be(True)
    expect(all([True, True])).to_be(True)
    expect(all([True, False])).to_be(False)
    expect(all([1, 2, 3])).to_be(True)
    expect(all([1, 0, 3])).to_be(False)
    expect(all(x > 0 for x in [1, 2, 3])).to_be(True)
    expect(all(x > 0 for x in [1, -1, 3])).to_be(False)

# === any ===
def test_any():
    expect(any([])).to_be(False)
    expect(any([False])).to_be(False)
    expect(any([False, False])).to_be(False)
    expect(any([False, True])).to_be(True)
    expect(any([0, 0, 0])).to_be(False)
    expect(any([0, 1, 0])).to_be(True)
    expect(any(x > 5 for x in [1, 2, 3])).to_be(False)
    expect(any(x > 2 for x in [1, 2, 3])).to_be(True)

# === bin ===
def test_bin():
    expect(bin(0)).to_be("0b0")
    expect(bin(1)).to_be("0b1")
    expect(bin(10)).to_be("0b1010")
    expect(bin(-1)).to_be("-0b1")
    expect(bin(255)).to_be("0b11111111")

# === callable ===
def test_callable():
    expect(callable(len)).to_be(True)
    expect(callable(lambda: None)).to_be(True)
    expect(callable(42)).to_be(False)
    expect(callable("hello")).to_be(False)
    expect(callable([])).to_be(False)
    class C:
        def __call__(self):
            pass
    expect(callable(C())).to_be(True)
    class D:
        pass
    expect(callable(D)).to_be(True)  # classes are callable
    expect(callable(D())).to_be(False)

# === chr/ord ===
def test_chr_ord():
    expect(chr(65)).to_be("A")
    expect(chr(97)).to_be("a")
    expect(chr(48)).to_be("0")
    expect(ord("A")).to_be(65)
    expect(ord("a")).to_be(97)
    expect(ord("0")).to_be(48)
    # Round-trip
    for i in range(128):
        expect(ord(chr(i))).to_be(i)

# === divmod ===
def test_divmod():
    expect(divmod(12, 7)).to_be((1, 5))
    expect(divmod(7, 3)).to_be((2, 1))
    expect(divmod(-12, 7)).to_be((-2, 2))
    expect(divmod(12, -7)).to_be((-2, -2))
    expect(divmod(0, 5)).to_be((0, 0))
    try:
        divmod(1, 0)
        expect("no error").to_be("ZeroDivisionError")
    except ZeroDivisionError:
        expect(True).to_be(True)

# === enumerate ===
def test_enumerate():
    expect(list(enumerate([]))).to_be([])
    expect(list(enumerate(["a", "b", "c"]))).to_be([(0, "a"), (1, "b"), (2, "c")])
    expect(list(enumerate("abc"))).to_be([(0, "a"), (1, "b"), (2, "c")])
    expect(list(enumerate(["x", "y"], 1))).to_be([(1, "x"), (2, "y")])
    expect(list(enumerate(["x"], 10))).to_be([(10, "x")])

# === filter ===
def test_filter():
    expect(list(filter(None, [0, 1, 2, "", "a", None, True, False]))).to_be([1, 2, "a", True])
    expect(list(filter(lambda x: x > 0, [-1, 0, 1, 2, -3]))).to_be([1, 2])
    expect(list(filter(lambda x: x % 2 == 0, range(10)))).to_be([0, 2, 4, 6, 8])
    expect(list(filter(None, []))).to_be([])

# === hex ===
def test_hex():
    expect(hex(0)).to_be("0x0")
    expect(hex(16)).to_be("0x10")
    expect(hex(255)).to_be("0xff")
    expect(hex(-16)).to_be("-0x10")
    expect(hex(1)).to_be("0x1")

# === oct ===
def test_oct():
    expect(oct(0)).to_be("0o0")
    expect(oct(8)).to_be("0o10")
    expect(oct(100)).to_be("0o144")
    expect(oct(-8)).to_be("-0o10")

# === isinstance/issubclass ===
def test_isinstance():
    expect(isinstance(1, int)).to_be(True)
    expect(isinstance(1, float)).to_be(False)
    expect(isinstance(1.0, float)).to_be(True)
    expect(isinstance("", str)).to_be(True)
    expect(isinstance([], list)).to_be(True)
    expect(isinstance({}, dict)).to_be(True)
    expect(isinstance((), tuple)).to_be(True)
    expect(isinstance(True, bool)).to_be(True)
    expect(isinstance(True, int)).to_be(True)
    # Tuple of types
    expect(isinstance(1, (int, float))).to_be(True)
    expect(isinstance(1.0, (int, float))).to_be(True)
    expect(isinstance("x", (int, float))).to_be(False)

def test_issubclass():
    expect(issubclass(bool, int)).to_be(True)
    expect(issubclass(int, int)).to_be(True)
    expect(issubclass(int, bool)).to_be(False)
    class A:
        pass
    class B(A):
        pass
    class C(B):
        pass
    expect(issubclass(B, A)).to_be(True)
    expect(issubclass(C, A)).to_be(True)
    expect(issubclass(A, B)).to_be(False)
    expect(issubclass(C, B)).to_be(True)

# === len ===
def test_len():
    expect(len([])).to_be(0)
    expect(len([1, 2, 3])).to_be(3)
    expect(len("")).to_be(0)
    expect(len("hello")).to_be(5)
    expect(len(())).to_be(0)
    expect(len((1, 2))).to_be(2)
    expect(len({})).to_be(0)
    expect(len({1: 2})).to_be(1)
    expect(len(set())).to_be(0)
    expect(len({1, 2, 3})).to_be(3)
    expect(len(range(10))).to_be(10)
    expect(len(b"hello")).to_be(5)

# === map ===
def test_map():
    expect(list(map(abs, [-1, -2, 3]))).to_be([1, 2, 3])
    expect(list(map(lambda x: x * 2, [1, 2, 3]))).to_be([2, 4, 6])
    expect(list(map(lambda x, y: x + y, [1, 2], [3, 4]))).to_be([4, 6])
    expect(list(map(str, [1, 2, 3]))).to_be(["1", "2", "3"])
    expect(list(map(lambda x: x, []))).to_be([])

# === max/min ===
def test_max():
    expect(max(1, 2, 3)).to_be(3)
    expect(max([1, 2, 3])).to_be(3)
    expect(max("abc")).to_be("c")
    expect(max([1])).to_be(1)
    expect(max([], default=0)).to_be(0)
    expect(max([1, -2, 3], key=abs)).to_be(3)
    try:
        max([])
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

def test_min():
    expect(min(1, 2, 3)).to_be(1)
    expect(min([1, 2, 3])).to_be(1)
    expect(min("abc")).to_be("a")
    expect(min([1])).to_be(1)
    expect(min([], default=0)).to_be(0)
    expect(min([1, -2, 3], key=abs)).to_be(1)
    try:
        min([])
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

# === pow ===
def test_pow():
    expect(pow(2, 10)).to_be(1024)
    expect(pow(2, 0)).to_be(1)
    expect(pow(0, 0)).to_be(1)
    expect(pow(2, -1)).to_be(0.5)
    expect(pow(2, 6, 10)).to_be(4)
    expect(pow(3, 3, 5)).to_be(2)

# === reversed ===
def test_reversed():
    expect(list(reversed([1, 2, 3]))).to_be([3, 2, 1])
    expect(list(reversed([]))).to_be([])
    expect(list(reversed("abc"))).to_be(["c", "b", "a"])
    expect(list(reversed((1, 2, 3)))).to_be([3, 2, 1])
    expect(list(reversed(range(5)))).to_be([4, 3, 2, 1, 0])

# === round ===
def test_round():
    expect(round(0)).to_be(0)
    expect(round(1)).to_be(1)
    # Banker's rounding
    expect(round(0.5)).to_be(0)
    expect(round(1.5)).to_be(2)
    expect(round(2.5)).to_be(2)
    expect(round(3.5)).to_be(4)
    expect(round(3.14159, 2)).to_be(3.14)
    expect(round(-1.5)).to_be(-2)
    expect(round(-0.5)).to_be(0)

# === sorted ===
def test_sorted():
    expect(sorted([3, 1, 2])).to_be([1, 2, 3])
    expect(sorted([])).to_be([])
    expect(sorted([1])).to_be([1])
    expect(sorted([3, 1, 2], reverse=True)).to_be([3, 2, 1])
    expect(sorted(["b", "a", "c"])).to_be(["a", "b", "c"])
    expect(sorted([-1, 3, -2], key=abs)).to_be([-1, -2, 3])
    # sorted returns a new list
    original = [3, 1, 2]
    result = sorted(original)
    expect(original).to_be([3, 1, 2])
    expect(result).to_be([1, 2, 3])

# === sum ===
def test_sum():
    expect(sum([])).to_be(0)
    expect(sum([1, 2, 3])).to_be(6)
    expect(sum([1, 2, 3], 10)).to_be(16)
    expect(sum(range(10))).to_be(45)
    expect(sum([0.1, 0.2])).to_be(0.30000000000000004)
    expect(sum(x for x in range(5))).to_be(10)

# === zip ===
def test_zip():
    expect(list(zip())).to_be([])
    expect(list(zip([1, 2, 3]))).to_be([(1,), (2,), (3,)])
    expect(list(zip([1, 2], [3, 4]))).to_be([(1, 3), (2, 4)])
    expect(list(zip([1, 2, 3], [4, 5]))).to_be([(1, 4), (2, 5)])  # stops at shortest
    expect(list(zip("abc", [1, 2, 3]))).to_be([("a", 1), ("b", 2), ("c", 3)])
    expect(list(zip([], [1, 2]))).to_be([])

# === hasattr/getattr/setattr ===
def test_hasattr_getattr():
    class Obj:
        x = 10
        def __init__(self):
            self.y = 20
    o = Obj()
    expect(hasattr(o, "x")).to_be(True)
    expect(hasattr(o, "y")).to_be(True)
    expect(hasattr(o, "z")).to_be(False)
    expect(getattr(o, "x")).to_be(10)
    expect(getattr(o, "y")).to_be(20)
    expect(getattr(o, "z", 42)).to_be(42)

# === hash ===
def test_hash():
    # hash of equal values must be equal
    expect(hash(1) == hash(1.0)).to_be(True)
    expect(hash(0) == hash(0.0)).to_be(True)
    expect(hash(True) == hash(1)).to_be(True)
    expect(hash(False) == hash(0)).to_be(True)
    # Unhashable types
    try:
        hash([1, 2])
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)
    try:
        hash({1: 2})
        expect("no error").to_be("TypeError")
    except TypeError:
        expect(True).to_be(True)

# Register all tests
test("abs", test_abs)
test("all", test_all)
test("any", test_any)
test("bin", test_bin)
test("callable", test_callable)
test("chr_ord", test_chr_ord)
test("divmod", test_divmod)
test("enumerate", test_enumerate)
test("filter", test_filter)
test("hex", test_hex)
test("oct", test_oct)
test("isinstance", test_isinstance)
test("issubclass", test_issubclass)
test("len", test_len)
test("map", test_map)
test("max", test_max)
test("min", test_min)
test("pow", test_pow)
test("reversed", test_reversed)
test("round", test_round)
test("sorted", test_sorted)
test("sum", test_sum)
test("zip", test_zip)
test("hasattr_getattr", test_hasattr_getattr)
test("hash", test_hash)

print("CPython builtin tests completed")
