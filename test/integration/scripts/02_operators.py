# Test: Operators
# Tests all Python operators

from test_framework import test, expect

def test_arithmetic():
    expect(15, 10 + 5)
    expect(5, 10 - 5)
    expect(50, 10 * 5)
    expect(2.5, 10 / 4)
    expect(3, 10 // 3)
    expect(1, 10 % 3)
    expect(1024, 2 ** 10)
    expect(5, +5)
    expect(-5, -5)

def test_float_arithmetic():
    expect(4.0, 1.5 + 2.5)
    expect(10.0, 2.5 * 4.0)
    expect(3.5, 7.0 / 2.0)
    expect(7.5, 5 + 2.5)
    expect(7.5, 3 * 2.5)

def test_comparison():
    expect(True, 5 == 5)
    expect(False, 5 == 6)
    expect(True, 5 != 6)
    expect(False, 5 != 5)
    expect(True, 5 < 10)
    expect(False, 10 < 5)
    expect(True, 5 <= 5)
    expect(False, 6 <= 5)
    expect(True, 10 > 5)
    expect(False, 5 > 10)
    expect(True, 5 >= 5)
    expect(False, 4 >= 5)

def test_chained_comparison():
    expect(True, 1 < 5 < 10)
    expect(False, 1 < 10 < 5)
    expect(True, 1 <= 2 <= 3 <= 4)

def test_string_comparison():
    expect(True, "hello" == "hello")
    expect(True, "hello" != "world")
    expect(True, "apple" < "banana")

def test_identity():
    a = [1, 2, 3]
    b = a
    c = [1, 2, 3]
    expect(True, a is b)
    expect(False, a is c)
    expect(True, a is not c)
    expect(False, a is not b)
    expect(True, None is None)

def test_membership():
    expect(True, 2 in [1, 2, 3])
    expect(True, 5 not in [1, 2, 3])
    expect(True, "ell" in "hello")
    expect(True, "xyz" not in "hello")
    expect(True, "a" in {"a": 1, "b": 2})
    expect(True, 2 in {1, 2, 3})
    expect(True, 2 in (1, 2, 3))

def test_logical():
    expect(True, True and True)
    expect(False, True and False)
    expect(True, False or True)
    expect(False, False or False)
    expect(True, not False)
    expect(False, not True)

def test_short_circuit():
    expect(False, False and (1 / 0))
    expect(True, True or (1 / 0))

def test_logical_values():
    expect(10, 5 and 10)
    expect(10, 0 or 10)
    expect(0, 5 and 0)
    expect(5, 0 or 5)

def test_bitwise():
    expect(8, 0b1100 & 0b1010)
    expect(14, 0b1100 | 0b1010)
    expect(6, 0b1100 ^ 0b1010)
    expect(-1, ~0)
    expect(16, 1 << 4)
    expect(4, 16 >> 2)

def test_augmented_assignment():
    x = 10
    x += 5
    expect(15, x)

    x = 10
    x -= 3
    expect(7, x)

    x = 10
    x *= 2
    expect(20, x)

    x = 10
    x //= 3
    expect(3, x)

    x = 10
    x %= 3
    expect(1, x)

    x = 2
    x **= 4
    expect(16, x)

    x = 0b1111
    x &= 0b1010
    expect(10, x)

    x = 0b1100
    x |= 0b0011
    expect(15, x)

def test_ternary():
    expect("yes", "yes" if True else "no")
    expect("no", "yes" if False else "no")
    expect("even", "even" if 10 % 2 == 0 else "odd")

def test_sequence_operators():
    expect("hello world", "hello" + " " + "world")
    expect("abababab", "ab" * 4)
    expect([1, 2, 3, 4], [1, 2] + [3, 4])
    expect([1, 2, 1, 2, 1, 2], [1, 2] * 3)

test("arithmetic", test_arithmetic)
test("float_arithmetic", test_float_arithmetic)
test("comparison", test_comparison)
test("chained_comparison", test_chained_comparison)
test("string_comparison", test_string_comparison)
test("identity", test_identity)
test("membership", test_membership)
test("logical", test_logical)
test("short_circuit", test_short_circuit)
test("logical_values", test_logical_values)
test("bitwise", test_bitwise)
test("augmented_assignment", test_augmented_assignment)
test("ternary", test_ternary)
test("sequence_operators", test_sequence_operators)

print("Operators tests completed")
