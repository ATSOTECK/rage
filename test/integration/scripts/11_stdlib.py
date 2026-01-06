# Test: Standard Library Modules
# Tests math, random, string, time, sys

from test_framework import test, expect

import math
import random
import string
import time
import sys

def test_math_constants():
    expect(math.pi > 3.14).to_be(True)
    expect(math.e > 2.71).to_be(True)

def test_math_functions():
    expect(math.sqrt(16)).to_be(4.0)
    expect(math.pow(2, 10)).to_be(1024.0)

def test_math_rounding():
    expect(math.ceil(3.2)).to_be(4)
    expect(math.floor(3.8)).to_be(3)

def test_math_other():
    expect(math.fabs(-5.5)).to_be(5.5)
    expect(math.factorial(5)).to_be(120)

def test_random_basic():
    random.seed(42)
    r = random.random()
    expect(0 <= r < 1).to_be(True)

def test_random_randint():
    random.seed(42)
    val = random.randint(1, 10)
    expect(1 <= val <= 10).to_be(True)

def test_random_choice():
    random.seed(42)
    val = random.choice([10, 20, 30, 40, 50])
    expect(val in [10, 20, 30, 40, 50]).to_be(True)

def test_random_shuffle():
    random.seed(42)
    items = [1, 2, 3, 4, 5]
    random.shuffle(items)
    expect(len(items)).to_be(5)
    expect(sum(items)).to_be(15)

def test_string_constants():
    expect(string.ascii_lowercase).to_be("abcdefghijklmnopqrstuvwxyz")
    expect(string.ascii_uppercase).to_be("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    expect(string.digits).to_be("0123456789")

def test_time_module():
    t = time.time()
    expect(t > 0).to_be(True)

def test_time_perf_counter():
    pc1 = time.perf_counter()
    pc2 = time.perf_counter()
    expect(pc2 >= pc1).to_be(True)

def test_time_monotonic():
    m1 = time.monotonic()
    m2 = time.monotonic()
    expect(m2 >= m1).to_be(True)

def test_sys_module():
    expect(len(sys.version) > 0).to_be(True)
    expect(len(sys.platform) > 0).to_be(True)

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
