# Test: Standard Library Modules
# Tests math, random, string, time, sys

from test_framework import test, expect

import math
import random
import string
import time
import sys

def test_math_constants():
    expect(True, math.pi > 3.14)
    expect(True, math.e > 2.71)

def test_math_functions():
    expect(4.0, math.sqrt(16))
    expect(1024.0, math.pow(2, 10))

def test_math_rounding():
    expect(4, math.ceil(3.2))
    expect(3, math.floor(3.8))

def test_math_other():
    expect(5.5, math.fabs(-5.5))
    expect(120, math.factorial(5))

def test_random_basic():
    random.seed(42)
    r = random.random()
    expect(True, 0 <= r < 1)

def test_random_randint():
    random.seed(42)
    val = random.randint(1, 10)
    expect(True, 1 <= val <= 10)

def test_random_choice():
    random.seed(42)
    val = random.choice([10, 20, 30, 40, 50])
    expect(True, val in [10, 20, 30, 40, 50])

def test_random_shuffle():
    random.seed(42)
    items = [1, 2, 3, 4, 5]
    random.shuffle(items)
    expect(5, len(items))
    expect(15, sum(items))

def test_string_constants():
    expect("abcdefghijklmnopqrstuvwxyz", string.ascii_lowercase)
    expect("ABCDEFGHIJKLMNOPQRSTUVWXYZ", string.ascii_uppercase)
    expect("0123456789", string.digits)

def test_time_module():
    t = time.time()
    expect(True, t > 0)

def test_time_perf_counter():
    pc1 = time.perf_counter()
    pc2 = time.perf_counter()
    expect(True, pc2 >= pc1)

def test_time_monotonic():
    m1 = time.monotonic()
    m2 = time.monotonic()
    expect(True, m2 >= m1)

def test_sys_module():
    expect(True, len(sys.version) > 0)
    expect(True, len(sys.platform) > 0)

test("math_constants", test_math_constants)
test("math_functions", test_math_functions)
test("math_rounding", test_math_rounding)
test("math_other", test_math_other)
test("random_basic", test_random_basic)
test("random_randint", test_random_randint)
test("random_choice", test_random_choice)
test("random_shuffle", test_random_shuffle)
test("string_constants", test_string_constants)
test("time_module", test_time_module)
test("time_perf_counter", test_time_perf_counter)
test("time_monotonic", test_time_monotonic)
test("sys_module", test_sys_module)

print("Standard library tests completed")
