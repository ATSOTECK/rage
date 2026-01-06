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

print("Asyncio tests completed")
