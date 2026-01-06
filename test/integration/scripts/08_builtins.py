# Test: Built-in Functions
# Tests commonly used built-in functions

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
    expect(3, int(3.7))
    expect(42, int("42"))
    expect(1, int(True))
    expect(42.0, float(42))
    expect(3.14, float("3.14"))
    expect("42", str(42))
    expect("3.14", str(3.14))
    expect("True", str(True))

def test_bool_conversion():
    expect(True, bool(1))
    expect(False, bool(0))
    expect(True, bool("hello"))
    expect(False, bool(""))
    expect(True, bool([1]))
    expect(False, bool([]))

def test_list_tuple_conversion():
    expect([1, 2, 3], list((1, 2, 3)))
    expect(["a", "b", "c"], list("abc"))
    expect([0, 1, 2, 3, 4], list(range(5)))
    expect((1, 2, 3), tuple([1, 2, 3]))
    expect(("a", "b", "c"), tuple("abc"))

def test_len():
    expect(5, len("hello"))
    expect(5, len([1, 2, 3, 4, 5]))
    expect(3, len((1, 2, 3)))
    expect(2, len({"a": 1, "b": 2}))

def test_min_max():
    expect(1, min(5, 2, 8, 1, 9))
    expect(9, max(5, 2, 8, 1, 9))
    expect(1, min([5, 2, 8, 1, 9]))
    expect(9, max([5, 2, 8, 1, 9]))

def test_sum():
    expect(15, sum([1, 2, 3, 4, 5]))
    expect(45, sum(range(10)))
    expect(0, sum([]))

def test_abs():
    expect(42, abs(42))
    expect(42, abs(-42))
    expect(3.14, abs(-3.14))
    expect(0, abs(0))

def test_ord_chr():
    expect(97, ord("a"))
    expect(65, ord("A"))
    expect("a", chr(97))
    expect("A", chr(65))

def test_isinstance():
    expect(True, isinstance(42, int))
    expect(True, isinstance("hello", str))
    expect(True, isinstance([1, 2], list))

def test_range():
    expect([0, 1, 2, 3, 4], list(range(5)))
    expect([2, 3, 4, 5, 6], list(range(2, 7)))
    expect([0, 2, 4, 6, 8], list(range(0, 10, 2)))
    expect([10, 9, 8, 7, 6, 5, 4, 3, 2, 1], list(range(10, 0, -1)))
    expect([], list(range(5, 2)))

def test_enumerate():
    expect([(0, "a"), (1, "b"), (2, "c")], list(enumerate(["a", "b", "c"])))
    expect([(1, "x"), (2, "y")], list(enumerate(["x", "y"], 1)))
    expect([], list(enumerate([])))
    expect([(0, "h"), (1, "i")], list(enumerate("hi")))

def test_zip():
    expect([(1, "a"), (2, "b"), (3, "c")], list(zip([1, 2, 3], ["a", "b", "c"])))
    expect([(1, "a"), (2, "b")], list(zip([1, 2], ["a", "b", "c", "d"])))
    expect([], list(zip()))
    expect([(1,), (2,), (3,)], list(zip([1, 2, 3])))
    expect([(1, "a", True), (2, "b", False)], list(zip([1, 2], ["a", "b"], [True, False])))

def test_map():
    expect([2, 4, 6, 8], list(map(double, [1, 2, 3, 4])))
    expect(["1", "2", "3"], list(map(str, [1, 2, 3])))
    expect([11, 22, 33], list(map(add, [1, 2, 3], [10, 20, 30])))
    expect([], list(map(double, [])))

def test_filter():
    expect([2, 4, 6], list(filter(is_even, [1, 2, 3, 4, 5, 6])))
    expect([1, 2], list(filter(is_positive, [-2, -1, 0, 1, 2])))
    expect([1, "hello", [1]], list(filter(None, [0, 1, "", "hello", [], [1]])))
    expect([], list(filter(is_even, [])))
    expect([], list(filter(is_even, [1, 3, 5])))

def test_reversed():
    expect([5, 4, 3, 2, 1], list(reversed([1, 2, 3, 4, 5])))
    expect(["o", "l", "l", "e", "h"], list(reversed("hello")))
    expect([3, 2, 1], list(reversed((1, 2, 3))))
    expect([], list(reversed([])))
    expect([42], list(reversed([42])))

def test_sorted():
    expect([1, 1, 2, 3, 4, 5, 6, 9], sorted([3, 1, 4, 1, 5, 9, 2, 6]))
    expect(["e", "h", "l", "l", "o"], sorted("hello"))
    expect([3, 2, 1], sorted([3, 1, 2], reverse=True))
    expect([], sorted([]))
    expect([42], sorted([42]))
    expect([1, 2, 3, 4, 5], sorted([1, 2, 3, 4, 5]))
    expect(["a", "pie", "apple"], sorted(["apple", "pie", "a"], key=get_len))

def test_all():
    expect(True, all([True, True, True]))
    expect(False, all([True, False, True]))
    expect(True, all([]))
    expect(True, all([1, 2, 3]))
    expect(False, all([1, 0, 3]))
    expect(True, all(["a", "b", "c"]))
    expect(False, all(["a", "", "c"]))

def test_any():
    expect(True, any([False, False, True]))
    expect(False, any([False, False, False]))
    expect(False, any([]))
    expect(True, any([0, 0, 1]))
    expect(False, any([0, 0, 0]))
    expect(True, any([0, "", [], "hello"]))

def test_hasattr():
    obj = AttrTest()
    expect(True, hasattr(obj, "x"))
    expect(False, hasattr(obj, "z"))
    expect(True, hasattr(obj, "__init__"))

def test_getattr():
    obj = AttrTest()
    expect(10, getattr(obj, "x"))
    expect(99, getattr(obj, "z", 99))
    expect(None, getattr(obj, "missing", None))

def test_setattr():
    obj = AttrTest()
    setattr(obj, "z", 30)
    expect(30, obj.z)
    setattr(obj, "x", 100)
    expect(100, obj.x)

def test_delattr():
    obj = AttrTest()
    setattr(obj, "temp", 42)
    expect(True, hasattr(obj, "temp"))
    delattr(obj, "temp")
    expect(False, hasattr(obj, "temp"))

def test_pow():
    expect(8, pow(2, 3))
    expect(1024, pow(2, 10))
    expect(3, pow(2, 3, 5))
    expect(10, pow(7, 2, 13))
    expect(1, pow(5, 0))
    expect(5, pow(5, 1))

def test_divmod():
    expect((3, 2), divmod(17, 5))
    expect((5, 0), divmod(10, 2))
    expect((-4, 3), divmod(-17, 5))
    expect((-4, -3), divmod(17, -5))

def test_hex():
    expect("0xff", hex(255))
    expect("0x10", hex(16))
    expect("0x0", hex(0))
    expect("-0xff", hex(-255))

def test_oct():
    expect("0o10", oct(8))
    expect("0o100", oct(64))
    expect("0o0", oct(0))
    expect("-0o10", oct(-8))

def test_bin():
    expect("0b101", bin(5))
    expect("0b11111111", bin(255))
    expect("0b0", bin(0))
    expect("-0b101", bin(-5))

def test_round():
    expect(4, round(3.7))
    expect(3, round(3.2))
    expect(2, round(2.5))  # Banker's rounding
    expect(4, round(3.5))  # Banker's rounding
    expect(3.14, round(3.14159, 2))
    expect(3.1416, round(3.14159, 4))
    expect(1200, round(1234, -2))
    expect(-2, round(-2.5))
    expect(0, round(0.0))

def test_callable():
    expect(True, callable(test_func))
    expect(True, callable(TestClass))
    expect(True, callable(CallableClass()))
    expect(False, callable(TestClass()))
    expect(False, callable(42))
    expect(False, callable("hello"))
    expect(False, callable([1, 2]))
    expect(True, callable(print))
    expect(False, callable(None))

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
