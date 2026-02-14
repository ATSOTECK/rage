# Test: Tuple Unpacking
# Tests implicit tuple creation and unpacking in assignments, returns, and yields

from test_framework import test, expect

def test_basic_unpacking():
    x, y = 1, 2
    expect(x).to_be(1)
    expect(y).to_be(2)

def test_three_values():
    a, b, c = 1, 2, 3
    expect(a).to_be(1)
    expect(b).to_be(2)
    expect(c).to_be(3)

def test_swap():
    a = 10
    b = 20
    a, b = b, a
    expect(a).to_be(20)
    expect(b).to_be(10)

def test_unpack_list():
    x, y = [1, 2]
    expect(x).to_be(1)
    expect(y).to_be(2)

def test_unpack_explicit_tuple():
    x, y = (1, 2)
    expect(x).to_be(1)
    expect(y).to_be(2)

def test_return_implicit_tuple():
    def pair():
        return 1, 2
    result = pair()
    expect(result[0]).to_be(1)
    expect(result[1]).to_be(2)

def test_return_unpack():
    def pair():
        return 10, 20
    x, y = pair()
    expect(x).to_be(10)
    expect(y).to_be(20)

def test_return_three():
    def triple():
        return "a", "b", "c"
    a, b, c = triple()
    expect(a).to_be("a")
    expect(b).to_be("b")
    expect(c).to_be("c")

def test_yield_tuple():
    def gen():
        yield 1, 2
        yield 3, 4
    g = gen()
    first = next(g)
    expect(first[0]).to_be(1)
    expect(first[1]).to_be(2)
    second = next(g)
    expect(second[0]).to_be(3)
    expect(second[1]).to_be(4)

def test_unpack_string_values():
    name, age = "Alice", 30
    expect(name).to_be("Alice")
    expect(age).to_be(30)

def test_nested_return_unpack():
    def get_pair():
        return 100, 200
    def use_pair():
        a, b = get_pair()
        return a + b
    expect(use_pair()).to_be(300)

test("basic tuple unpacking", test_basic_unpacking)
test("three value unpacking", test_three_values)
test("swap values", test_swap)
test("unpack list", test_unpack_list)
test("unpack explicit tuple", test_unpack_explicit_tuple)
test("return implicit tuple", test_return_implicit_tuple)
test("return and unpack", test_return_unpack)
test("return three values", test_return_three)
test("yield tuple", test_yield_tuple)
test("unpack with string values", test_unpack_string_values)
test("nested return unpack", test_nested_return_unpack)
