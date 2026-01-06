# Test: Async/Await and Asyncio
# Tests async functions, await expressions, and asyncio module

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
    expect(42, asyncio.run(simple_async()))

def test_async_with_args():
    expect(42, asyncio.run(async_add(10, 32)))

def test_nested_await():
    expect(20, asyncio.run(outer_async()))

def test_chained_async():
    expect(16, asyncio.run(step3()))

def test_async_locals():
    expect(30, asyncio.run(async_locals()))

def test_async_conditional():
    expect("yes", asyncio.run(async_conditional(True)))
    expect("no", asyncio.run(async_conditional(False)))

def test_async_loop():
    expect(10, asyncio.run(async_sum_range(5)))

def test_gather():
    result = asyncio.gather(get_one(), get_two(), get_three())
    expect([1, 2, 3], result)

def test_async_returns_list():
    expect([1, 2, 3], asyncio.run(async_list()))

def test_async_returns_dict():
    expect({"a": 1, "b": 2}, asyncio.run(async_dict()))

def test_multiple_awaits():
    expect(15, asyncio.run(combine()))

def test_async_pipeline():
    expect(11, asyncio.run(process(5)))

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
