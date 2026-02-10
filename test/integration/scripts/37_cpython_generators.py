# Test: CPython Generator Edge Cases
# Adapted from CPython's test_generators.py

from test_framework import test, expect

def test_generator_basic():
    def gen():
        yield 1
        yield 2
        yield 3
    expect(list(gen())).to_be([1, 2, 3])

def test_generator_empty():
    def gen():
        return
        yield  # Makes it a generator
    expect(list(gen())).to_be([])

def test_generator_single():
    def gen():
        yield 42
    g = gen()
    expect(next(g)).to_be(42)
    try:
        next(g)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        expect(True).to_be(True)

def test_generator_send():
    def gen():
        x = yield 1
        yield x * 2
    g = gen()
    expect(next(g)).to_be(1)
    expect(g.send(5)).to_be(10)

def test_generator_send_none():
    def gen():
        x = yield 1
        yield x
    g = gen()
    expect(next(g)).to_be(1)
    expect(g.send(None)).to_be(None)

def test_generator_close():
    closed = [False]
    def gen():
        try:
            yield 1
            yield 2
        finally:
            closed[0] = True
    g = gen()
    next(g)
    g.close()
    expect(closed[0]).to_be(True)

def test_generator_throw():
    def gen():
        try:
            yield 1
        except ValueError:
            yield "caught"
    g = gen()
    expect(next(g)).to_be(1)
    expect(g.throw(ValueError)).to_be("caught")

def test_generator_return_value():
    def gen():
        yield 1
        return "done"
    g = gen()
    next(g)
    try:
        next(g)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        expect(True).to_be(True)

def test_generator_expression():
    g = (x * 2 for x in range(5))
    expect(list(g)).to_be([0, 2, 4, 6, 8])

def test_generator_expression_filter():
    g = (x for x in range(10) if x % 2 == 0)
    expect(list(g)).to_be([0, 2, 4, 6, 8])

def test_generator_expression_nested():
    g = (x * y for x in range(3) for y in range(3))
    expect(list(g)).to_be([0, 0, 0, 0, 1, 2, 0, 2, 4])

def test_generator_multiple_yields():
    def gen():
        for i in range(5):
            yield i
    expect(list(gen())).to_be([0, 1, 2, 3, 4])

def test_generator_with_accumulator():
    def running_sum():
        total = 0
        while True:
            x = yield total
            if x is None:
                break
            total = total + x
    g = running_sum()
    next(g)  # prime
    expect(g.send(1)).to_be(1)
    expect(g.send(2)).to_be(3)
    expect(g.send(3)).to_be(6)

def test_generator_yield_in_try():
    results = []
    def gen():
        try:
            yield 1
            yield 2
        finally:
            results.append("finally")
    g = gen()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    try:
        next(g)
    except StopIteration:
        pass
    expect("finally" in results).to_be(True)

def test_generator_exhausted():
    def gen():
        yield 1
    g = gen()
    expect(next(g)).to_be(1)
    # After exhaustion, always raises StopIteration
    try:
        next(g)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        pass
    try:
        next(g)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        expect(True).to_be(True)

def test_generator_as_iterator():
    def gen():
        yield 1
        yield 2
    g = gen()
    expect(iter(g) is g).to_be(True)

def test_generator_in_for_loop():
    def gen():
        yield "a"
        yield "b"
        yield "c"
    result = ""
    for ch in gen():
        result += ch
    expect(result).to_be("abc")

def test_generator_with_conditional():
    def gen(flag):
        if flag:
            yield "yes"
        else:
            yield "no"
    expect(list(gen(True))).to_be(["yes"])
    expect(list(gen(False))).to_be(["no"])

def test_generator_fibonacci():
    def fib():
        a = 0
        b = 1
        while True:
            yield a
            temp = a + b
            a = b
            b = temp
    g = fib()
    result = []
    for _ in range(10):
        result.append(next(g))
    expect(result).to_be([0, 1, 1, 2, 3, 5, 8, 13, 21, 34])

def test_generator_sum():
    def gen():
        yield 1
        yield 2
        yield 3
    expect(sum(gen())).to_be(6)

def test_generator_in_list_constructor():
    def gen():
        yield 1
        yield 2
        yield 3
    expect(list(gen())).to_be([1, 2, 3])

def test_generator_in_tuple_constructor():
    def gen():
        yield 1
        yield 2
    expect(tuple(gen())).to_be((1, 2))

def test_generator_in_set_constructor():
    def gen():
        yield 1
        yield 2
        yield 1
    expect(set(gen())).to_be({1, 2})

def test_generator_chained():
    def gen1():
        yield 1
        yield 2
    def gen2():
        yield 3
        yield 4
    result = list(gen1()) + list(gen2())
    expect(result).to_be([1, 2, 3, 4])

def test_generator_yield_none():
    def gen():
        yield None
        yield None
    expect(list(gen())).to_be([None, None])

# Register all tests
test("generator_basic", test_generator_basic)
test("generator_empty", test_generator_empty)
test("generator_single", test_generator_single)
test("generator_send", test_generator_send)
test("generator_send_none", test_generator_send_none)
test("generator_close", test_generator_close)
test("generator_throw", test_generator_throw)
test("generator_return_value", test_generator_return_value)
test("generator_expression", test_generator_expression)
test("generator_expression_filter", test_generator_expression_filter)
test("generator_expression_nested", test_generator_expression_nested)
test("generator_multiple_yields", test_generator_multiple_yields)
test("generator_with_accumulator", test_generator_with_accumulator)
test("generator_yield_in_try", test_generator_yield_in_try)
test("generator_exhausted", test_generator_exhausted)
test("generator_as_iterator", test_generator_as_iterator)
test("generator_in_for_loop", test_generator_in_for_loop)
test("generator_with_conditional", test_generator_with_conditional)
test("generator_fibonacci", test_generator_fibonacci)
test("generator_sum", test_generator_sum)
test("generator_in_list_constructor", test_generator_in_list_constructor)
test("generator_in_tuple_constructor", test_generator_in_tuple_constructor)
test("generator_in_set_constructor", test_generator_in_set_constructor)
test("generator_chained", test_generator_chained)
test("generator_yield_none", test_generator_yield_none)

print("CPython generator tests completed")
