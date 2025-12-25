# Test: Async/Await and Asyncio
# Tests async functions, await expressions, and asyncio module

import asyncio

results = {}

# Basic async function
async def simple_async():
    return 42

results["basic_async"] = asyncio.run(simple_async())

# Async function with computation
async def async_add(a, b):
    return a + b

results["async_with_args"] = asyncio.run(async_add(10, 32))

# Async function calling another async function
async def inner_async():
    return 10

async def outer_async():
    val = await inner_async()
    return val * 2

results["nested_await"] = asyncio.run(outer_async())

# Async chain of calls
async def step1():
    return 5

async def step2():
    x = await step1()
    return x + 3

async def step3():
    y = await step2()
    return y * 2

results["chained_async"] = asyncio.run(step3())

# Async with local variables
async def async_locals():
    x = 10
    y = 20
    z = x + y
    return z

results["async_locals"] = asyncio.run(async_locals())

# Async with conditionals
async def async_conditional(flag):
    if flag:
        return "yes"
    else:
        return "no"

results["async_conditional_true"] = asyncio.run(async_conditional(True))
results["async_conditional_false"] = asyncio.run(async_conditional(False))

# Async with loop
async def async_sum_range(n):
    total = 0
    for i in range(n):
        total = total + i
    return total

results["async_loop"] = asyncio.run(async_sum_range(5))

# asyncio.gather with multiple coroutines
async def get_one():
    return 1

async def get_two():
    return 2

async def get_three():
    return 3

# Gather runs coroutines and returns a list of results
results["gather_results"] = asyncio.gather(get_one(), get_two(), get_three())

# Async function returning list
async def async_list():
    return [1, 2, 3]

results["async_returns_list"] = asyncio.run(async_list())

# Async function returning dict
async def async_dict():
    return {"a": 1, "b": 2}

results["async_returns_dict"] = asyncio.run(async_dict())

# Async function with multiple awaits
async def compute_a():
    return 5

async def compute_b():
    return 10

async def combine():
    a = await compute_a()
    b = await compute_b()
    return a + b

results["multiple_awaits"] = asyncio.run(combine())

# Nested async calls with computation
async def double(x):
    return x * 2

async def add_one(x):
    return x + 1

async def process(x):
    d = await double(x)
    result = await add_one(d)
    return result

results["async_pipeline"] = asyncio.run(process(5))

print("Asyncio tests completed")
