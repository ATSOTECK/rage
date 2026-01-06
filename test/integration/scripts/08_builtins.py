# Test: Built-in Functions
# Tests commonly used built-in functions

from test_framework import test, expect

# Helper functions and classes at module level
def double(x):
    return x * 2

def add(a, b):
    return a + b

def is_even(x):
    return x % 2 == 0

def is_positive(x):
    return x > 0

def get_len(x):
    return len(x)

def test_func():
    pass

class AttrTest:
    def __init__(self):
        self.x = 10
        self.y = 20

class TestClass:
    pass

class CallableClass:
    def __call__(self):
        pass

def test_type_constructors():
    expect(int(3.7)).to_be(3)
    expect(int("42")).to_be(42)
    expect(int(True)).to_be(1)
    expect(float(42)).to_be(42.0)
    expect(float("3.14")).to_be(3.14)
    expect(str(42)).to_be("42")
    expect(str(3.14)).to_be("3.14")
    expect(str(True)).to_be("True")

def test_bool_conversion():
    expect(bool(1)).to_be(True)
    expect(bool(0)).to_be(False)
    expect(bool("hello")).to_be(True)
    expect(bool("")).to_be(False)
    expect(bool([1])).to_be(True)
    expect(bool([])).to_be(False)

def test_list_tuple_conversion():
    expect(list((1, 2, 3))).to_be([1, 2, 3])
    expect(list("abc")).to_be(["a", "b", "c"])
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(tuple([1, 2, 3])).to_be((1, 2, 3))
    expect(tuple("abc")).to_be(("a", "b", "c"))

def test_len():
    expect(len("hello")).to_be(5)
    expect(len([1, 2, 3, 4, 5])).to_be(5)
    expect(len((1, 2, 3))).to_be(3)
    expect(len({"a": 1, "b": 2})).to_be(2)

def test_min_max():
    expect(min(5, 2, 8, 1, 9)).to_be(1)
    expect(max(5, 2, 8, 1, 9)).to_be(9)
    expect(min([5, 2, 8, 1, 9])).to_be(1)
    expect(max([5, 2, 8, 1, 9])).to_be(9)

def test_sum():
    expect(sum([1, 2, 3, 4, 5])).to_be(15)
    expect(sum(range(10))).to_be(45)
    expect(sum([])).to_be(0)

def test_abs():
    expect(abs(42)).to_be(42)
    expect(abs(-42)).to_be(42)
    expect(abs(-3.14)).to_be(3.14)
    expect(abs(0)).to_be(0)

def test_ord_chr():
    expect(ord("a")).to_be(97)
    expect(ord("A")).to_be(65)
    expect(chr(97)).to_be("a")
    expect(chr(65)).to_be("A")

def test_isinstance():
    expect(isinstance(42, int)).to_be(True)
    expect(isinstance("hello", str)).to_be(True)
    expect(isinstance([1, 2], list)).to_be(True)

def test_range():
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(list(range(2, 7))).to_be([2, 3, 4, 5, 6])
    expect(list(range(0, 10, 2))).to_be([0, 2, 4, 6, 8])
    expect(list(range(10, 0, -1))).to_be([10, 9, 8, 7, 6, 5, 4, 3, 2, 1])
    expect(list(range(5, 2))).to_be([])

def test_enumerate():
    expect(list(enumerate(["a", "b", "c"]))).to_be([(0, "a"), (1, "b"), (2, "c")])
    expect(list(enumerate(["x", "y"], 1))).to_be([(1, "x"), (2, "y")])
    expect(list(enumerate([]))).to_be([])
    expect(list(enumerate("hi"))).to_be([(0, "h"), (1, "i")])

def test_zip():
    expect(list(zip([1, 2, 3], ["a", "b", "c"]))).to_be([(1, "a"), (2, "b"), (3, "c")])
    expect(list(zip([1, 2], ["a", "b", "c", "d"]))).to_be([(1, "a"), (2, "b")])
    expect(list(zip())).to_be([])
    expect(list(zip([1, 2, 3]))).to_be([(1,), (2,), (3,)])
    expect(list(zip([1, 2], ["a", "b"], [True, False]))).to_be([(1, "a", True), (2, "b", False)])

def test_map():
    expect(list(map(double, [1, 2, 3, 4]))).to_be([2, 4, 6, 8])
    expect(list(map(str, [1, 2, 3]))).to_be(["1", "2", "3"])
    expect(list(map(add, [1, 2, 3], [10, 20, 30]))).to_be([11, 22, 33])
    expect(list(map(double, []))).to_be([])

def test_filter():
    expect(list(filter(is_even, [1, 2, 3, 4, 5, 6]))).to_be([2, 4, 6])
    expect(list(filter(is_positive, [-2, -1, 0, 1, 2]))).to_be([1, 2])
    expect(list(filter(None, [0, 1, "", "hello", [], [1]]))).to_be([1, "hello", [1]])
    expect(list(filter(is_even, []))).to_be([])
    expect(list(filter(is_even, [1, 3, 5]))).to_be([])

def test_reversed():
    expect(list(reversed([1, 2, 3, 4, 5]))).to_be([5, 4, 3, 2, 1])
    expect(list(reversed("hello"))).to_be(["o", "l", "l", "e", "h"])
    expect(list(reversed((1, 2, 3)))).to_be([3, 2, 1])
    expect(list(reversed([]))).to_be([])
    expect(list(reversed([42]))).to_be([42])

def test_sorted():
    expect(sorted([3, 1, 4, 1, 5, 9, 2, 6])).to_be([1, 1, 2, 3, 4, 5, 6, 9])
    expect(sorted("hello")).to_be(["e", "h", "l", "l", "o"])
    expect(sorted([3, 1, 2], reverse=True)).to_be([3, 2, 1])
    expect(sorted([])).to_be([])
    expect(sorted([42])).to_be([42])
    expect(sorted([1, 2, 3, 4, 5])).to_be([1, 2, 3, 4, 5])
    expect(sorted(["apple", "pie", "a"], key=get_len)).to_be(["a", "pie", "apple"])

def test_all():
    expect(all([True, True, True])).to_be(True)
    expect(all([True, False, True])).to_be(False)
    expect(all([])).to_be(True)
    expect(all([1, 2, 3])).to_be(True)
    expect(all([1, 0, 3])).to_be(False)
    expect(all(["a", "b", "c"])).to_be(True)
    expect(all(["a", "", "c"])).to_be(False)

def test_any():
    expect(any([False, False, True])).to_be(True)
    expect(any([False, False, False])).to_be(False)
    expect(any([])).to_be(False)
    expect(any([0, 0, 1])).to_be(True)
    expect(any([0, 0, 0])).to_be(False)
    expect(any([0, "", [], "hello"])).to_be(True)

def test_hasattr():
    obj = AttrTest()
    expect(hasattr(obj, "x")).to_be(True)
    expect(hasattr(obj, "z")).to_be(False)
    expect(hasattr(obj, "__init__")).to_be(True)

def test_getattr():
    obj = AttrTest()
    expect(getattr(obj, "x")).to_be(10)
    expect(getattr(obj, "z", 99)).to_be(99)
    expect(getattr(obj, "missing", None)).to_be(None)

def test_setattr():
    obj = AttrTest()
    setattr(obj, "z", 30)
    expect(obj.z).to_be(30)
    setattr(obj, "x", 100)
    expect(obj.x).to_be(100)

def test_delattr():
    obj = AttrTest()
    setattr(obj, "temp", 42)
    expect(hasattr(obj, "temp")).to_be(True)
    delattr(obj, "temp")
    expect(hasattr(obj, "temp")).to_be(False)

def test_pow():
    expect(pow(2, 3)).to_be(8)
    expect(pow(2, 10)).to_be(1024)
    expect(pow(2, 3, 5)).to_be(3)
    expect(pow(7, 2, 13)).to_be(10)
    expect(pow(5, 0)).to_be(1)
    expect(pow(5, 1)).to_be(5)

def test_divmod():
    expect(divmod(17, 5)).to_be((3, 2))
    expect(divmod(10, 2)).to_be((5, 0))
    expect(divmod(-17, 5)).to_be((-4, 3))
    expect(divmod(17, -5)).to_be((-4, -3))

def test_hex():
    expect(hex(255)).to_be("0xff")
    expect(hex(16)).to_be("0x10")
    expect(hex(0)).to_be("0x0")
    expect(hex(-255)).to_be("-0xff")

def test_oct():
    expect(oct(8)).to_be("0o10")
    expect(oct(64)).to_be("0o100")
    expect(oct(0)).to_be("0o0")
    expect(oct(-8)).to_be("-0o10")

def test_bin():
    expect(bin(5)).to_be("0b101")
    expect(bin(255)).to_be("0b11111111")
    expect(bin(0)).to_be("0b0")
    expect(bin(-5)).to_be("-0b101")

def test_round():
    expect(round(3.7)).to_be(4)
    expect(round(3.2)).to_be(3)
    expect(round(2.5)).to_be(2)  # Banker's rounding
    expect(round(3.5)).to_be(4)  # Banker's rounding
    expect(round(3.14159, 2)).to_be(3.14)
    expect(round(3.14159, 4)).to_be(3.1416)
    expect(round(1234, -2)).to_be(1200)
    expect(round(-2.5)).to_be(-2)
    expect(round(0.0)).to_be(0)

def test_callable():
    expect(callable(test_func)).to_be(True)
    expect(callable(TestClass)).to_be(True)
    expect(callable(CallableClass())).to_be(True)
    expect(callable(TestClass())).to_be(False)
    expect(callable(42)).to_be(False)
    expect(callable("hello")).to_be(False)
    expect(callable([1, 2])).to_be(False)
    expect(callable(print)).to_be(True)
    expect(callable(None)).to_be(False)

test("type_constructors", test_type_constructors)
test("bool_conversion", test_bool_conversion)
test("list_tuple_conversion", test_list_tuple_conversion)
test("len", test_len)
test("min_max", test_min_max)
test("sum", test_sum)
test("abs", test_abs)
test("ord_chr", test_ord_chr)
test("isinstance", test_isinstance)
test("range", test_range)
test("enumerate", test_enumerate)
test("zip", test_zip)
test("map", test_map)
test("filter", test_filter)
test("reversed", test_reversed)
test("sorted", test_sorted)
test("all", test_all)
test("any", test_any)
test("hasattr", test_hasattr)
test("getattr", test_getattr)
test("setattr", test_setattr)
test("delattr", test_delattr)
test("pow", test_pow)
test("divmod", test_divmod)
test("hex", test_hex)
test("oct", test_oct)
test("bin", test_bin)
test("round", test_round)
test("callable", test_callable)

print("Builtins tests completed")
