# Test: Operators
# Tests all Python operators

from test_framework import test, expect

def test_arithmetic():
    expect(10 + 5).to_be(15)
    expect(10 - 5).to_be(5)
    expect(10 * 5).to_be(50)
    expect(10 / 4).to_be(2.5)
    expect(10 // 3).to_be(3)
    expect(10 % 3).to_be(1)
    expect(2 ** 10).to_be(1024)
    expect(+5).to_be(5)
    expect(-5).to_be(-5)

def test_float_arithmetic():
    expect(1.5 + 2.5).to_be(4.0)
    expect(2.5 * 4.0).to_be(10.0)
    expect(7.0 / 2.0).to_be(3.5)
    expect(5 + 2.5).to_be(7.5)
    expect(3 * 2.5).to_be(7.5)

def test_comparison():
    expect(5 == 5).to_be(True)
    expect(5 == 6).to_be(False)
    expect(5 != 6).to_be(True)
    expect(5 != 5).to_be(False)
    expect(5 < 10).to_be(True)
    expect(10 < 5).to_be(False)
    expect(5 <= 5).to_be(True)
    expect(6 <= 5).to_be(False)
    expect(10 > 5).to_be(True)
    expect(5 > 10).to_be(False)
    expect(5 >= 5).to_be(True)
    expect(4 >= 5).to_be(False)

def test_chained_comparison():
    expect(1 < 5 < 10).to_be(True)
    expect(1 < 10 < 5).to_be(False)
    expect(1 <= 2 <= 3 <= 4).to_be(True)

def test_string_comparison():
    expect("hello" == "hello").to_be(True)
    expect("hello" != "world").to_be(True)
    expect("apple" < "banana").to_be(True)

def test_identity():
    a = [1, 2, 3]
    b = a
    c = [1, 2, 3]
    expect(a is b).to_be(True)
    expect(a is c).to_be(False)
    expect(a is not c).to_be(True)
    expect(a is not b).to_be(False)
    expect(None is None).to_be(True)

def test_membership():
    expect(2 in [1, 2, 3]).to_be(True)
    expect(5 not in [1, 2, 3]).to_be(True)
    expect("ell" in "hello").to_be(True)
    expect("xyz" not in "hello").to_be(True)
    expect("a" in {"a": 1, "b": 2}).to_be(True)
    expect(2 in {1, 2, 3}).to_be(True)
    expect(2 in (1, 2, 3)).to_be(True)

def test_logical():
    expect(True and True).to_be(True)
    expect(True and False).to_be(False)
    expect(False or True).to_be(True)
    expect(False or False).to_be(False)
    expect(not False).to_be(True)
    expect(not True).to_be(False)

def test_short_circuit():
    expect(False and (1 / 0)).to_be(False)
    expect(True or (1 / 0)).to_be(True)

def test_logical_values():
    expect(5 and 10).to_be(10)
    expect(0 or 10).to_be(10)
    expect(5 and 0).to_be(0)
    expect(0 or 5).to_be(5)

def test_bitwise():
    expect(0b1100 & 0b1010).to_be(8)
    expect(0b1100 | 0b1010).to_be(14)
    expect(0b1100 ^ 0b1010).to_be(6)
    expect(~0).to_be(-1)
    expect(1 << 4).to_be(16)
    expect(16 >> 2).to_be(4)

def test_augmented_assignment():
    x = 10
    x += 5
    expect(x).to_be(15)

    x = 10
    x -= 3
    expect(x).to_be(7)

    x = 10
    x *= 2
    expect(x).to_be(20)

    x = 10
    x //= 3
    expect(x).to_be(3)

    x = 10
    x %= 3
    expect(x).to_be(1)

    x = 2
    x **= 4
    expect(x).to_be(16)

    x = 0b1111
    x &= 0b1010
    expect(x).to_be(10)

    x = 0b1100
    x |= 0b0011
    expect(x).to_be(15)

def test_ternary():
    expect("yes" if True else "no").to_be("yes")
    expect("yes" if False else "no").to_be("no")
    expect("even" if 10 % 2 == 0 else "odd").to_be("even")

def test_sequence_operators():
    expect("hello" + " " + "world").to_be("hello world")
    expect("ab" * 4).to_be("abababab")
    expect([1, 2] + [3, 4]).to_be([1, 2, 3, 4])
    expect([1, 2] * 3).to_be([1, 2, 1, 2, 1, 2])

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
