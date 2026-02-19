# Test: CPython Unary Operators and Comparisons
# Adapted from CPython's test_unary.py and test_compare.py

from test_framework import test, expect

# === Unary minus ===
def test_unary_minus():
    expect(-1).to_be(-1)
    expect(-0).to_be(0)
    expect(-(-1)).to_be(1)
    expect(-(3.14)).to_be(-3.14)
    expect(-True).to_be(-1)
    expect(-False).to_be(0)

# === Unary plus ===
def test_unary_plus():
    expect(+1).to_be(1)
    expect(+0).to_be(0)
    expect(+(-1)).to_be(-1)
    expect(+(3.14)).to_be(3.14)

# === Bitwise NOT ===
def test_bitwise_not():
    expect(~0).to_be(-1)
    expect(~1).to_be(-2)
    expect(~(-1)).to_be(0)
    expect(~42).to_be(-43)
    expect(~~42).to_be(42)

# === Not ===
def test_not_operator():
    expect(not True).to_be(False)
    expect(not False).to_be(True)
    expect(not 0).to_be(True)
    expect(not 1).to_be(False)
    expect(not "").to_be(True)
    expect(not "hello").to_be(False)
    expect(not []).to_be(True)
    expect(not [1]).to_be(False)
    expect(not None).to_be(True)

# === Chained comparisons ===
def test_chained_comparison():
    expect(1 < 2 < 3).to_be(True)
    expect(1 < 2 < 2).to_be(False)
    expect(1 <= 1 <= 1).to_be(True)
    expect(3 > 2 > 1).to_be(True)
    expect(3 > 2 > 2).to_be(False)
    expect(1 < 2 <= 2 < 3).to_be(True)

# === Comparison operators ===
def test_int_comparisons():
    expect(1 < 2).to_be(True)
    expect(2 < 1).to_be(False)
    expect(1 > 2).to_be(False)
    expect(2 > 1).to_be(True)
    expect(1 <= 1).to_be(True)
    expect(1 >= 1).to_be(True)
    expect(1 <= 2).to_be(True)
    expect(2 >= 1).to_be(True)

def test_float_comparisons():
    expect(1.0 < 2.0).to_be(True)
    expect(1.0 == 1.0).to_be(True)
    expect(1.0 != 2.0).to_be(True)
    expect(1.5 > 1.4).to_be(True)

def test_mixed_comparisons():
    expect(1 == 1.0).to_be(True)
    expect(1 < 1.5).to_be(True)
    expect(2.0 > 1).to_be(True)
    expect(0 == 0.0).to_be(True)

# === String comparisons ===
def test_string_comparisons():
    expect("a" < "b").to_be(True)
    expect("b" > "a").to_be(True)
    expect("abc" < "abd").to_be(True)
    expect("abc" < "abcd").to_be(True)
    expect("" < "a").to_be(True)
    expect("abc" == "abc").to_be(True)
    expect("abc" != "abd").to_be(True)

# === List comparisons ===
def test_list_comparisons():
    expect([1, 2] < [1, 3]).to_be(True)
    expect([1, 2] < [1, 2, 3]).to_be(True)
    expect([] < [1]).to_be(True)
    expect([1, 2, 3] == [1, 2, 3]).to_be(True)
    expect([1, 2] != [1, 3]).to_be(True)
    expect([1, 2] > [1, 1]).to_be(True)

# === Tuple comparisons ===
def test_tuple_comparisons():
    expect((1, 2) < (1, 3)).to_be(True)
    expect((1, 2) < (1, 2, 3)).to_be(True)
    expect(()) .to_be(())
    expect((1, 2) == (1, 2)).to_be(True)
    expect((1, 2) != (1, 3)).to_be(True)

# === Identity vs equality ===
def test_is_operator():
    a = [1, 2, 3]
    b = a
    c = [1, 2, 3]
    expect(a is b).to_be(True)
    expect(a is c).to_be(False)
    expect(a is not c).to_be(True)
    expect(None is None).to_be(True)

# === None comparisons ===
def test_none_comparisons():
    expect(None == None).to_be(True)
    expect(None is None).to_be(True)
    expect(None != 0).to_be(True)
    expect(None != "").to_be(True)
    expect(None != False).to_be(True)

# === Boolean as int ===
def test_bool_as_int():
    expect(True + True).to_be(2)
    expect(False + 1).to_be(1)
    expect(True * 5).to_be(5)
    expect(False * 100).to_be(0)

# === Bitwise operators ===
def test_bitwise_and():
    expect(0xFF & 0x0F).to_be(0x0F)
    expect(12 & 10).to_be(8)
    expect(0 & 0xFFFF).to_be(0)

def test_bitwise_or():
    expect(0xF0 | 0x0F).to_be(0xFF)
    expect(12 | 10).to_be(14)
    expect(0 | 0).to_be(0)

def test_bitwise_xor():
    expect(0xFF ^ 0x0F).to_be(0xF0)
    expect(12 ^ 10).to_be(6)
    expect(42 ^ 42).to_be(0)

# === Shift operators ===
def test_left_shift():
    expect(1 << 0).to_be(1)
    expect(1 << 1).to_be(2)
    expect(1 << 8).to_be(256)
    expect(5 << 3).to_be(40)

def test_right_shift():
    expect(256 >> 8).to_be(1)
    expect(16 >> 1).to_be(8)
    expect(1 >> 1).to_be(0)
    expect(100 >> 2).to_be(25)

# === Floor division ===
def test_floor_division():
    expect(7 // 3).to_be(2)
    expect(-7 // 3).to_be(-3)
    expect(7 // -3).to_be(-3)
    expect(-7 // -3).to_be(2)
    expect(10 // 3).to_be(3)

# === Modulo ===
def test_modulo():
    expect(7 % 3).to_be(1)
    expect(-7 % 3).to_be(2)
    expect(7 % -3).to_be(-2)
    expect(10 % 5).to_be(0)

# === Power ===
def test_power():
    expect(2 ** 0).to_be(1)
    expect(2 ** 10).to_be(1024)
    expect((-1) ** 2).to_be(1)
    expect((-1) ** 3).to_be(-1)

# === Augmented assignment ===
def test_augmented_assignment():
    x = 10
    x += 5
    expect(x).to_be(15)
    x -= 3
    expect(x).to_be(12)
    x *= 2
    expect(x).to_be(24)
    x //= 5
    expect(x).to_be(4)
    x **= 3
    expect(x).to_be(64)
    x %= 10
    expect(x).to_be(4)

# === Custom __neg__ and __pos__ ===
def test_custom_unary():
    class MyNum:
        def __init__(self, val):
            self.val = val
        def __neg__(self):
            return type(self)(-self.val)
        def __pos__(self):
            return type(self)(abs(self.val))
    n = MyNum(5)
    expect((-n).val).to_be(-5)
    expect((+MyNum(-3)).val).to_be(3)

# === Custom __lt__, __gt__, __le__, __ge__ ===
def test_custom_comparison():
    class Num:
        def __init__(self, val):
            self.val = val
        def __lt__(self, other):
            return self.val < other.val
        def __le__(self, other):
            return self.val <= other.val
        def __gt__(self, other):
            return self.val > other.val
        def __ge__(self, other):
            return self.val >= other.val
        def __eq__(self, other):
            return self.val == other.val
    a = Num(1)
    b = Num(2)
    c = Num(1)
    expect(a < b).to_be(True)
    expect(b > a).to_be(True)
    expect(a <= c).to_be(True)
    expect(a >= c).to_be(True)
    expect(a == c).to_be(True)
    expect(a < c).to_be(False)

# Register all tests
test("unary_minus", test_unary_minus)
test("unary_plus", test_unary_plus)
test("bitwise_not", test_bitwise_not)
test("not_operator", test_not_operator)
test("chained_comparison", test_chained_comparison)
test("int_comparisons", test_int_comparisons)
test("float_comparisons", test_float_comparisons)
test("mixed_comparisons", test_mixed_comparisons)
test("string_comparisons", test_string_comparisons)
test("list_comparisons", test_list_comparisons)
test("tuple_comparisons", test_tuple_comparisons)
test("is_operator", test_is_operator)
test("none_comparisons", test_none_comparisons)
test("bool_as_int", test_bool_as_int)
test("bitwise_and", test_bitwise_and)
test("bitwise_or", test_bitwise_or)
test("bitwise_xor", test_bitwise_xor)
test("left_shift", test_left_shift)
test("right_shift", test_right_shift)
test("floor_division", test_floor_division)
test("modulo", test_modulo)
test("power", test_power)
test("augmented_assignment", test_augmented_assignment)
test("custom_unary", test_custom_unary)
test("custom_comparison", test_custom_comparison)

# Ported from CPython test_unary.py

# === Negation equivalence (CPython test_negative) ===
def test_cpython_negative():
    """Negation matches subtraction from zero"""
    expect(-2 == 0 - 2).to_be(True)
    expect(-0).to_be(0)
    expect(--2).to_be(2)
    expect(-2.0 == 0 - 2.0).to_be(True)
    expect(-2j == 0 - 2j).to_be(True)

test("cpython_negative", test_cpython_negative)

# === Positive equivalence (CPython test_positive) ===
def test_cpython_positive():
    """Positive operator is identity for numeric types"""
    expect(+2).to_be(2)
    expect(+0).to_be(0)
    expect(++2).to_be(2)
    expect(+2.0).to_be(2.0)
    expect(+2j).to_be(2j)

test("cpython_positive", test_cpython_positive)

# === Invert identity (CPython test_invert) ===
def test_cpython_invert():
    """Bitwise invert: ~x == -(x+1)"""
    expect(~2 == -(2 + 1)).to_be(True)
    expect(~0).to_be(-1)
    expect(~~2).to_be(2)

test("cpython_invert", test_cpython_invert)

# === Negation of exponentiation / precedence (CPython test_negation_of_exponentiation) ===
def test_cpython_negation_of_exponentiation():
    """Make sure ** binds tighter than unary minus (SourceForge bug #456756)"""
    # -2 ** 3 means -(2**3), not (-2)**3
    expect(-2 ** 3).to_be(-8)
    expect((-2) ** 3).to_be(-8)
    # -2 ** 4 means -(2**4) = -16, but (-2)**4 = 16
    expect(-2 ** 4).to_be(-16)
    expect((-2) ** 4).to_be(16)

test("cpython_negation_of_exponentiation", test_cpython_negation_of_exponentiation)

# === Bad types raise TypeError (CPython test_bad_types) ===
def test_cpython_bad_types_unary_plus():
    """Unary + on string raises TypeError"""
    raised = False
    try:
        +"a"
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("cpython_bad_types_unary_plus", test_cpython_bad_types_unary_plus)

def test_cpython_bad_types_unary_minus():
    """Unary - on string raises TypeError"""
    raised = False
    try:
        -"a"
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("cpython_bad_types_unary_minus", test_cpython_bad_types_unary_minus)

def test_cpython_bad_types_invert_string():
    """Bitwise ~ on string raises TypeError"""
    raised = False
    try:
        ~"a"
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("cpython_bad_types_invert_string", test_cpython_bad_types_invert_string)

def test_cpython_bad_types_invert_complex():
    """Bitwise ~ on complex raises TypeError"""
    raised = False
    try:
        ~2j
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("cpython_bad_types_invert_complex", test_cpython_bad_types_invert_complex)

def test_cpython_bad_types_invert_float():
    """Bitwise ~ on float raises TypeError"""
    raised = False
    try:
        ~2.0
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("cpython_bad_types_invert_float", test_cpython_bad_types_invert_float)

# === No overflow with moderately large ints (within int64 range) ===
def test_cpython_no_overflow():
    """Unary ops on large (but int64-safe) integers"""
    big = 999999999999999999  # within int64 range
    expect(+big).to_be(999999999999999999)
    expect(-big).to_be(-999999999999999999)
    expect(~big).to_be(-1000000000000000000)

test("cpython_no_overflow", test_cpython_no_overflow)

# === Custom __invert__ ===
def test_custom_invert():
    """Test custom __invert__ dunder method"""
    class MyInt:
        def __init__(self, val):
            self.val = val
        def __invert__(self):
            return MyInt(~self.val)
    n = MyInt(42)
    expect((~n).val).to_be(-43)
    expect((~~n).val).to_be(42)

test("custom_invert", test_custom_invert)

# === Float unary operations ===
def test_float_unary():
    """Float negation and positive"""
    expect(-0.0).to_be(0.0)
    expect(-1.5).to_be(-1.5)
    expect(--1.5).to_be(1.5)
    expect(+(-1.5)).to_be(-1.5)

test("float_unary", test_float_unary)

# === Complex unary operations ===
def test_complex_unary():
    """Complex negation and positive"""
    expect(-1j).to_be(-1j)
    expect(-(1 + 2j)).to_be(-1 - 2j)
    expect(+(1 + 2j)).to_be(1 + 2j)
    expect(--1j).to_be(1j)

test("complex_unary", test_complex_unary)

# === Unary plus on list/dict/None raises TypeError ===
def test_bad_types_unary_plus_list():
    """Unary + on list raises TypeError"""
    raised = False
    try:
        +[]
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("bad_types_unary_plus_list", test_bad_types_unary_plus_list)

def test_bad_types_unary_minus_none():
    """Unary - on None raises TypeError"""
    raised = False
    try:
        -None
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("bad_types_unary_minus_none", test_bad_types_unary_minus_none)

def test_bad_types_invert_list():
    """Bitwise ~ on list raises TypeError"""
    raised = False
    try:
        ~[]
    except TypeError:
        raised = True
    expect(raised).to_be(True)

test("bad_types_invert_list", test_bad_types_invert_list)

print("CPython unary/comparison tests completed")
