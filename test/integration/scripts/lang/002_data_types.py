# Test: Data Types
# Tests basic Python data types

from test_framework import test, expect

def test_none():
    expect(None).to_be(None)
    expect(None is None).to_be(True)

def test_bool():
    expect(True).to_be(True)
    expect(False).to_be(False)
    expect(bool(1)).to_be(True)
    expect(bool(0)).to_be(False)
    expect(bool("hello")).to_be(True)
    expect(bool("")).to_be(False)

def test_int():
    expect(42).to_be(42)
    expect(-17).to_be(-17)
    expect(0).to_be(0)
    expect(1000000).to_be(1000000)
    expect(int(3.7)).to_be(3)
    expect(int("123")).to_be(123)

def test_float():
    expect(3.14).to_be(3.14)
    expect(-2.5).to_be(-2.5)
    expect(0.0).to_be(0.0)
    expect(float(42)).to_be(42.0)
    expect(float("3.14")).to_be(3.14)

def test_string():
    expect('hello').to_be("hello")
    expect("world").to_be("world")
    expect("").to_be("")
    expect("hello world").to_be("hello world")
    expect("hello" + " " + "world").to_be("hello world")
    expect("ab" * 3).to_be("ababab")
    expect(len("hello")).to_be(5)
    expect("hello"[1]).to_be("e")
    expect("hello"[-1]).to_be("o")

def test_list():
    expect([]).to_be([])
    expect([1, 2, 3]).to_be([1, 2, 3])
    expect([1, "two", 3.0, True]).to_be([1, "two", 3.0, True])
    expect([[1, 2], [3, 4]]).to_be([[1, 2], [3, 4]])
    expect(len([1, 2, 3, 4, 5])).to_be(5)
    expect([10, 20, 30][1]).to_be(20)
    expect([1, 2] + [3, 4]).to_be([1, 2, 3, 4])
    expect([1, 2] * 2).to_be([1, 2, 1, 2])
    expect([1, 2, 3][-1]).to_be(3)

def test_tuple():
    expect(()).to_be(())
    expect((1,)).to_be((1,))
    expect((1, 2, 3)).to_be((1, 2, 3))
    expect((1, "two", 3.0)).to_be((1, "two", 3.0))
    expect(len((1, 2, 3))).to_be(3)
    expect((10, 20, 30)[1]).to_be(20)

def test_dict():
    expect({}).to_be({})
    expect({"a": 1, "b": 2}).to_be({"a": 1, "b": 2})
    expect({"int": 1, "str": "hello", "bool": True}).to_be({"int": 1, "str": "hello", "bool": True})
    expect(len({"a": 1, "b": 2, "c": 3})).to_be(3)
    expect({"x": 10, "y": 20}["x"]).to_be(10)
    expect({"a": 1}.get("b", 99)).to_be(99)

def test_range():
    expect(list(range(5))).to_be([0, 1, 2, 3, 4])
    expect(list(range(2, 7))).to_be([2, 3, 4, 5, 6])
    expect(list(range(0, 10, 2))).to_be([0, 2, 4, 6, 8])

def test_isinstance():
    expect(isinstance(42, int)).to_be(True)
    expect(isinstance("hello", str)).to_be(True)
    expect(isinstance([1, 2], list)).to_be(True)

test("none", test_none)
test("bool", test_bool)
test("int", test_int)
test("float", test_float)
test("string", test_string)
test("list", test_list)
test("tuple", test_tuple)
test("dict", test_dict)
test("range", test_range)
test("isinstance", test_isinstance)

print("Data types tests completed")
