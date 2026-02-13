# Test: Yield From
# Tests yield from delegation to sub-generators and iterables

from test_framework import test, expect

def test_yield_from_generator():
    """yield from delegates to another generator"""
    def inner():
        yield 1
        yield 2
        yield 3

    def outer():
        yield from inner()
        yield 4

    result = list(outer())
    expect(result).to_be([1, 2, 3, 4])

def test_yield_from_list():
    """yield from iterates over a list"""
    def gen():
        yield from [10, 20, 30]

    result = list(gen())
    expect(result).to_be([10, 20, 30])

def test_yield_from_string():
    """yield from iterates over characters of a string"""
    def gen():
        yield from "abc"

    result = list(gen())
    expect(result).to_be(["a", "b", "c"])

def test_yield_from_tuple():
    """yield from iterates over a tuple"""
    def gen():
        yield from (1, 2, 3)

    result = list(gen())
    expect(result).to_be([1, 2, 3])

def test_yield_from_range():
    """yield from iterates over a range"""
    def gen():
        yield from range(5)

    result = list(gen())
    expect(result).to_be([0, 1, 2, 3, 4])

def test_yield_from_nested():
    """Multiple levels of yield from delegation"""
    def gen1():
        yield 1

    def gen2():
        yield from gen1()
        yield 2

    def gen3():
        yield from gen2()
        yield 3

    result = list(gen3())
    expect(result).to_be([1, 2, 3])

def test_yield_from_multiple():
    """Multiple yield from in the same generator"""
    def gen():
        yield from [1, 2]
        yield from [3, 4]
        yield from [5, 6]

    result = list(gen())
    expect(result).to_be([1, 2, 3, 4, 5, 6])

def test_yield_from_with_regular_yield():
    """Mix of yield and yield from"""
    def gen():
        yield 0
        yield from [1, 2, 3]
        yield 4
        yield from [5, 6]
        yield 7

    result = list(gen())
    expect(result).to_be([0, 1, 2, 3, 4, 5, 6, 7])

def test_yield_from_empty():
    """yield from empty iterable"""
    def gen():
        yield 1
        yield from range(0)
        yield 2

    result = list(gen())
    expect(result).to_be([1, 2])

def test_yield_from_generator_expression():
    """yield from a generator expression"""
    def gen():
        yield from (x * 2 for x in range(5))

    result = list(gen())
    expect(result).to_be([0, 2, 4, 6, 8])

def test_yield_from_flatten():
    """Use yield from to flatten a list of lists"""
    def flatten(lst):
        for item in lst:
            if isinstance(item, list):
                yield from flatten(item)
            else:
                yield item

    nested = [1, [2, 3], [4, [5, 6]], 7]
    result = list(flatten(nested))
    expect(result).to_be([1, 2, 3, 4, 5, 6, 7])

test("yield_from_generator", test_yield_from_generator)
test("yield_from_list", test_yield_from_list)
test("yield_from_string", test_yield_from_string)
test("yield_from_tuple", test_yield_from_tuple)
test("yield_from_range", test_yield_from_range)
test("yield_from_nested", test_yield_from_nested)
test("yield_from_multiple", test_yield_from_multiple)
test("yield_from_with_regular_yield", test_yield_from_with_regular_yield)
test("yield_from_empty", test_yield_from_empty)
test("yield_from_generator_expression", test_yield_from_generator_expression)
test("yield_from_flatten", test_yield_from_flatten)
