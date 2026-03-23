# Test: Async/Await and Asyncio
# Tests async functions, await expressions, and asyncio module

from test_framework import test, expect

import asyncio

async def simple_async():
    return 42

async def async_add(a, b):
    return a + b

async def inner_async():
    return 10

async def outer_async():
    val = await inner_async()
    return val * 2

async def step1():
    return 5

async def step2():
    x = await step1()
    return x + 3

async def step3():
    y = await step2()
    return y * 2

async def async_locals():
    x = 10
    y = 20
    z = x + y
    return z

async def async_conditional(flag):
    if flag:
        return "yes"
    else:
        return "no"

async def async_sum_range(n):
    total = 0
    for i in range(n):
        total = total + i
    return total

async def get_one():
    return 1

async def get_two():
    return 2

async def get_three():
    return 3

async def async_list():
    return [1, 2, 3]

async def async_dict():
    return {"a": 1, "b": 2}

async def compute_a():
    return 5

async def compute_b():
    return 10

async def combine():
    a = await compute_a()
    b = await compute_b()
    return a + b

async def double(x):
    return x * 2

async def add_one(x):
    return x + 1

async def process(x):
    d = await double(x)
    result = await add_one(d)
    return result

def test_basic_async():
    expect(asyncio.run(simple_async())).to_be(42)

def test_async_with_args():
    expect(asyncio.run(async_add(10, 32))).to_be(42)

def test_nested_await():
    expect(asyncio.run(outer_async())).to_be(20)

def test_chained_async():
    expect(asyncio.run(step3())).to_be(16)

def test_async_locals():
    expect(asyncio.run(async_locals())).to_be(30)

def test_async_conditional():
    expect(asyncio.run(async_conditional(True))).to_be("yes")
    expect(asyncio.run(async_conditional(False))).to_be("no")

def test_async_loop():
    expect(asyncio.run(async_sum_range(5))).to_be(10)

def test_gather():
    result = asyncio.gather(get_one(), get_two(), get_three())
    expect(result).to_be([1, 2, 3])

def test_async_returns_list():
    expect(asyncio.run(async_list())).to_be([1, 2, 3])

def test_async_returns_dict():
    expect(asyncio.run(async_dict())).to_be({"a": 1, "b": 2})

def test_multiple_awaits():
    expect(asyncio.run(combine())).to_be(15)

def test_async_pipeline():
    expect(asyncio.run(process(5))).to_be(11)

# Tests for coroutine.throw
def test_coroutine_throw_into_closed():
    # Test throwing into an already closed coroutine
    coro = simple_async()
    # Run it to completion
    asyncio.run(coro)
    # Now try to throw - should re-raise the exception
    threw_error = False
    try:
        coro.throw(ValueError, "test error")
    except ValueError:
        threw_error = True
    expect(threw_error).to_be(True)

def test_coroutine_throw_into_new():
    # Test throwing into a just-created coroutine (never started)
    coro = simple_async()
    threw_error = False
    try:
        coro.throw(RuntimeError, "early throw")
    except RuntimeError:
        threw_error = True
    expect(threw_error).to_be(True)

def test_coroutine_throw_method_exists():
    # Verify coroutine has throw method
    coro = simple_async()
    has_throw = hasattr(coro, "throw")
    # Clean up - run the coroutine to avoid warnings
    try:
        asyncio.run(coro)
    except:
        pass
    expect(has_throw).to_be(True)

def test_coroutine_throw_exception_type():
    # Test that exception type is preserved when throwing
    coro = simple_async()
    exception_type = None
    try:
        coro.throw(KeyError, "test key error")
    except KeyError:
        exception_type = "KeyError"
    except:
        exception_type = "other"
    expect(exception_type).to_be("KeyError")

test("basic_async", test_basic_async)
test("async_with_args", test_async_with_args)
test("nested_await", test_nested_await)
test("chained_async", test_chained_async)
test("async_locals", test_async_locals)
test("async_conditional", test_async_conditional)
test("async_loop", test_async_loop)
test("gather", test_gather)
test("async_returns_list", test_async_returns_list)
test("async_returns_dict", test_async_returns_dict)
test("multiple_awaits", test_multiple_awaits)
test("async_pipeline", test_async_pipeline)
test("coroutine_throw_into_closed", test_coroutine_throw_into_closed)
test("coroutine_throw_into_new", test_coroutine_throw_into_new)
test("coroutine_throw_method_exists", test_coroutine_throw_method_exists)
test("coroutine_throw_exception_type", test_coroutine_throw_exception_type)

# --- Coroutine exception state isolation ---

def test_coroutine_internal_exception():
    async def might_fail():
        try:
            raise ValueError("inner")
        except ValueError:
            pass
        return "ok"
    expect(asyncio.run(might_fail())).to_be("ok")

def test_coroutine_in_except_handler():
    async def inner():
        return 42
    caught = False
    try:
        raise ValueError("outer")
    except ValueError:
        result = asyncio.run(inner())
        caught = True
    expect(result).to_be(42)
    expect(caught).to_be(True)

def test_coroutine_nested_exceptions():
    async def inner():
        try:
            raise KeyError("k")
        except KeyError:
            return "caught_inner"

    async def outer():
        try:
            r = await inner()
            return r
        except Exception:
            return "wrong"

    expect(asyncio.run(outer())).to_be("caught_inner")

def test_coroutine_try_finally():
    async def with_finally():
        result = []
        try:
            result.append("try")
            return result
        finally:
            result.append("finally")

    result = asyncio.run(with_finally())
    expect(len(result)).to_be(2)
    expect(result[0]).to_be("try")
    expect(result[1]).to_be("finally")

test("coroutine_internal_exception", test_coroutine_internal_exception)
test("coroutine_in_except_handler", test_coroutine_in_except_handler)
test("coroutine_nested_exceptions", test_coroutine_nested_exceptions)
test("coroutine_try_finally", test_coroutine_try_finally)

# --- Coroutine close ---

def test_coroutine_close_not_started():
    async def never_started():
        return 1
    c = never_started()
    c.close()
    closed_ok = True
    expect(closed_ok).to_be(True)

def test_coroutine_close_already_finished():
    async def simple():
        return 1
    c = simple()
    result = asyncio.run(c)
    c.close()
    c.close()  # double close
    expect(result).to_be(1)

def test_coroutine_close_with_finally():
    finally_ran = [False]

    async def coro():
        try:
            x = 1
        finally:
            finally_ran[0] = True
        return x

    result = asyncio.run(coro())
    expect(finally_ran[0]).to_be(True)
    expect(result).to_be(1)

test("coroutine_close_not_started", test_coroutine_close_not_started)
test("coroutine_close_already_finished", test_coroutine_close_already_finished)
test("coroutine_close_with_finally", test_coroutine_close_with_finally)

print("Asyncio tests completed")
