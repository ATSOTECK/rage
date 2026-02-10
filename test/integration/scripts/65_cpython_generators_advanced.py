# Test: CPython Advanced Generator Patterns
# Adapted from CPython's test_generators.py - covers advanced generator patterns
# beyond 37_cpython_generators.py

from test_framework import test, expect

# =============================================================================
# Generator with try/finally ensuring cleanup
# =============================================================================

def test_generator_try_finally_complete():
    """Generator that completes naturally triggers finally."""
    log = []
    def gen():
        try:
            yield 1
            yield 2
        finally:
            log.append("cleanup")
    g = gen()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    try:
        next(g)
    except StopIteration:
        pass
    expect(log).to_be(["cleanup"])

test("generator_try_finally_complete", test_generator_try_finally_complete)

# =============================================================================
# Generator with try/finally on close
# =============================================================================

def test_generator_try_finally_close():
    """Closing a generator triggers finally."""
    log = []
    def gen():
        try:
            yield 1
            yield 2
            yield 3
        finally:
            log.append("closed")
    g = gen()
    expect(next(g)).to_be(1)
    g.close()
    expect(log).to_be(["closed"])

test("generator_try_finally_close", test_generator_try_finally_close)

# =============================================================================
# Generator with multiple yields in sequence
# =============================================================================

def test_generator_many_yields():
    """Generator yielding many values in sequence."""
    def gen():
        yield "a"
        yield "b"
        yield "c"
        yield "d"
        yield "e"
    result = []
    for item in gen():
        result.append(item)
    expect(result).to_be(["a", "b", "c", "d", "e"])

test("generator_many_yields", test_generator_many_yields)

# =============================================================================
# Sending values to generators
# =============================================================================

def test_generator_send_values():
    """Send values into a generator to control behavior."""
    def accumulator():
        total = 0
        while True:
            value = yield total
            if value is None:
                break
            total = total + value
    g = accumulator()
    next(g)  # prime
    expect(g.send(10)).to_be(10)
    expect(g.send(20)).to_be(30)
    expect(g.send(5)).to_be(35)

test("generator_send_values", test_generator_send_values)

# =============================================================================
# Generator send with echo pattern
# =============================================================================

def test_generator_send_echo():
    """Generator that echoes back what is sent."""
    def echo():
        value = yield "ready"
        while True:
            value = yield "echo: " + str(value)
    g = echo()
    expect(next(g)).to_be("ready")
    expect(g.send("hello")).to_be("echo: hello")
    expect(g.send(42)).to_be("echo: 42")
    expect(g.send("world")).to_be("echo: world")

test("generator_send_echo", test_generator_send_echo)

# =============================================================================
# Generator pipelines (chaining generators)
# =============================================================================

def test_generator_pipeline():
    """Chain generators to form a processing pipeline."""
    def numbers(n):
        for i in range(n):
            yield i

    def doubled(source):
        for item in source:
            yield item * 2

    def filtered(source):
        for item in source:
            if item > 4:
                yield item

    pipe = filtered(doubled(numbers(10)))
    result = list(pipe)
    expect(result).to_be([6, 8, 10, 12, 14, 16, 18])

test("generator_pipeline", test_generator_pipeline)

# =============================================================================
# Generator as data producer
# =============================================================================

def test_generator_data_producer():
    """Generator producing structured data."""
    def records():
        yield {"name": "Alice", "age": 30}
        yield {"name": "Bob", "age": 25}
        yield {"name": "Charlie", "age": 35}

    names = []
    for rec in records():
        names.append(rec["name"])
    expect(names).to_be(["Alice", "Bob", "Charlie"])

test("generator_data_producer", test_generator_data_producer)

# =============================================================================
# Infinite generator with take-first-N pattern
# =============================================================================

def test_generator_take_n():
    """Take first N items from an infinite generator."""
    def naturals():
        n = 0
        while True:
            yield n
            n = n + 1

    def take(n, gen):
        result = []
        count = 0
        for item in gen:
            if count >= n:
                break
            result.append(item)
            count = count + 1
        return result

    expect(take(5, naturals())).to_be([0, 1, 2, 3, 4])
    expect(take(0, naturals())).to_be([])

test("generator_take_n", test_generator_take_n)

# =============================================================================
# Generator state preservation between yields
# =============================================================================

def test_generator_state_preservation():
    """Generator preserves local variable state between yields."""
    def stateful():
        x = 1
        yield x
        x = x + 10
        yield x
        x = x * 2
        yield x
        x = x - 5
        yield x

    g = stateful()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(11)
    expect(next(g)).to_be(22)
    expect(next(g)).to_be(17)

test("generator_state_preservation", test_generator_state_preservation)

# =============================================================================
# Yield from (delegating generator)
# =============================================================================

def test_yield_from_basic():
    """yield from delegates to a sub-generator."""
    def inner():
        yield 1
        yield 2

    def outer():
        yield 0
        yield from inner()
        yield 3

    expect(list(outer())).to_be([0, 1, 2, 3])

test("yield_from_basic", test_yield_from_basic)

# =============================================================================
# Nested yield from
# =============================================================================

def test_yield_from_nested():
    """Multiple levels of yield from."""
    def level3():
        yield "c"

    def level2():
        yield "b"
        yield from level3()

    def level1():
        yield "a"
        yield from level2()
        yield "d"

    expect(list(level1())).to_be(["a", "b", "c", "d"])

test("yield_from_nested", test_yield_from_nested)

# =============================================================================
# yield from with iterables (not just generators)
# =============================================================================

def test_yield_from_iterables():
    """yield from works with any iterable."""
    def gen():
        yield from [1, 2, 3]
        yield from range(4, 7)

    expect(list(gen())).to_be([1, 2, 3, 4, 5, 6])

test("yield_from_iterables", test_yield_from_iterables)

# =============================================================================
# Generator with conditional yields
# =============================================================================

def test_generator_conditional_yields():
    """Generator that conditionally yields values."""
    def conditional(items):
        for item in items:
            if item > 0:
                yield item

    expect(list(conditional([3, -1, 4, -1, 5, -9, 2]))).to_be([3, 4, 5, 2])

test("generator_conditional_yields", test_generator_conditional_yields)

# =============================================================================
# Generator comprehension equivalence
# =============================================================================

def test_genexpr_equivalence():
    """Generator expression produces same results as equivalent generator function."""
    def gen_func():
        for x in range(5):
            yield x * x

    genexpr = (x * x for x in range(5))

    expect(list(gen_func())).to_be(list(genexpr))

test("genexpr_equivalence", test_genexpr_equivalence)

# =============================================================================
# Fibonacci generator
# =============================================================================

def test_fibonacci_generator():
    """Classic Fibonacci generator."""
    def fib():
        a = 0
        b = 1
        while True:
            yield a
            temp = a + b
            a = b
            b = temp

    g = fib()
    results = []
    for _ in range(12):
        results.append(next(g))
    expect(results).to_be([0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89])

test("fibonacci_generator", test_fibonacci_generator)

# =============================================================================
# Range-like custom generator
# =============================================================================

def test_custom_range_generator():
    """Custom range-like generator with step."""
    def my_range(start, stop, step):
        current = start
        while current < stop:
            yield current
            current = current + step

    expect(list(my_range(0, 10, 2))).to_be([0, 2, 4, 6, 8])
    expect(list(my_range(1, 5, 1))).to_be([1, 2, 3, 4])
    expect(list(my_range(0, 0, 1))).to_be([])

test("custom_range_generator", test_custom_range_generator)

# =============================================================================
# Generator that yields different types
# =============================================================================

def test_generator_mixed_types():
    """Generator can yield values of different types."""
    def mixed():
        yield 1
        yield "hello"
        yield [1, 2]
        yield None
        yield True
        yield 3.14

    result = list(mixed())
    expect(result[0]).to_be(1)
    expect(result[1]).to_be("hello")
    expect(result[2]).to_be([1, 2])
    expect(result[3]).to_be(None)
    expect(result[4]).to_be(True)
    expect(result[5]).to_be(3.14)

test("generator_mixed_types", test_generator_mixed_types)

# =============================================================================
# Empty generator (return before yield)
# =============================================================================

def test_empty_generator():
    """Generator with no yields produces nothing."""
    def empty():
        return
        yield  # makes it a generator

    expect(list(empty())).to_be([])

test("empty_generator", test_empty_generator)

# =============================================================================
# Generator with return value (StopIteration carries value)
# =============================================================================

def test_generator_return_value():
    """Generator return value is accessible via StopIteration."""
    def gen():
        yield 1
        yield 2
        return "done"

    g = gen()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

test("generator_return_value", test_generator_return_value)

# =============================================================================
# Yield in loop
# =============================================================================

def test_yield_in_for_loop():
    """Yield inside a for loop."""
    def squares(n):
        for i in range(n):
            yield i * i

    expect(list(squares(6))).to_be([0, 1, 4, 9, 16, 25])

test("yield_in_for_loop", test_yield_in_for_loop)

# =============================================================================
# Yield in while loop
# =============================================================================

def test_yield_in_while_loop():
    """Yield inside a while loop."""
    def countdown(n):
        while n > 0:
            yield n
            n = n - 1

    expect(list(countdown(5))).to_be([5, 4, 3, 2, 1])
    expect(list(countdown(0))).to_be([])

test("yield_in_while_loop", test_yield_in_while_loop)

# =============================================================================
# Generator reuse (can't iterate twice)
# =============================================================================

def test_generator_no_reuse():
    """A generator object can only be iterated once."""
    def gen():
        yield 1
        yield 2
        yield 3

    g = gen()
    first_pass = list(g)
    second_pass = list(g)
    expect(first_pass).to_be([1, 2, 3])
    expect(second_pass).to_be([])

test("generator_no_reuse", test_generator_no_reuse)

# =============================================================================
# Generator with try/except inside
# =============================================================================

def test_generator_internal_try_except():
    """Generator with try/except that catches internally."""
    def safe_convert(items):
        for item in items:
            try:
                yield int(item)
            except:
                yield -1

    data = ["10", "20", "30"]
    expect(list(safe_convert(data))).to_be([10, 20, 30])

test("generator_internal_try_except", test_generator_internal_try_except)

# =============================================================================
# Generator throw
# =============================================================================

def test_generator_throw_catches():
    """Generator can catch thrown exceptions."""
    def gen():
        try:
            yield 1
        except ValueError:
            yield "caught ValueError"
        yield "after"

    g = gen()
    expect(next(g)).to_be(1)
    expect(g.throw(ValueError)).to_be("caught ValueError")
    expect(next(g)).to_be("after")

test("generator_throw_catches", test_generator_throw_catches)

# =============================================================================
# Generator with enumerate-like pattern
# =============================================================================

def test_generator_enumerate_like():
    """Generator that adds indices like enumerate."""
    def my_enumerate(iterable):
        index = 0
        for item in iterable:
            yield [index, item]
            index = index + 1

    items = ["a", "b", "c"]
    result = list(my_enumerate(items))
    expect(result).to_be([[0, "a"], [1, "b"], [2, "c"]])

test("generator_enumerate_like", test_generator_enumerate_like)

# =============================================================================
# Generator map/filter patterns
# =============================================================================

def test_generator_map_pattern():
    """Generator used as a map function."""
    def gen_map(func, iterable):
        for item in iterable:
            yield func(item)

    def double(x):
        return x * 2

    expect(list(gen_map(double, [1, 2, 3, 4]))).to_be([2, 4, 6, 8])

test("generator_map_pattern", test_generator_map_pattern)

def test_generator_filter_pattern():
    """Generator used as a filter function."""
    def gen_filter(predicate, iterable):
        for item in iterable:
            if predicate(item):
                yield item

    def is_even(x):
        return x % 2 == 0

    expect(list(gen_filter(is_even, range(10)))).to_be([0, 2, 4, 6, 8])

test("generator_filter_pattern", test_generator_filter_pattern)

# =============================================================================
# Multiple active generators
# =============================================================================

def test_multiple_active_generators():
    """Multiple generator instances active simultaneously."""
    def counter(start):
        n = start
        while True:
            yield n
            n = n + 1

    g1 = counter(0)
    g2 = counter(100)
    g3 = counter(200)

    expect(next(g1)).to_be(0)
    expect(next(g2)).to_be(100)
    expect(next(g3)).to_be(200)
    expect(next(g1)).to_be(1)
    expect(next(g1)).to_be(2)
    expect(next(g2)).to_be(101)
    expect(next(g3)).to_be(201)

test("multiple_active_generators", test_multiple_active_generators)

print("CPython advanced generator tests completed")
