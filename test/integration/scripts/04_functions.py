# Test: Functions
# Tests function definitions, arguments, and recursion

from test_framework import test, expect

# Define helper functions at module level so recursion works
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

def test_simple_function():
    def simple():
        return 42
    expect(simple()).to_be(42)

def test_function_args():
    def add(a, b):
        return a + b
    expect(add(10, 20)).to_be(30)

def test_default_args():
    def greet(name, greeting="Hello"):
        return greeting + ", " + name
    expect(greet("World")).to_be("Hello, World")
    expect(greet("World", "Hi")).to_be("Hi, World")

def test_multi_default_args():
    def make_point(x=0, y=0, z=0):
        return [x, y, z]
    expect(make_point()).to_be([0, 0, 0])
    expect(make_point(1, 2)).to_be([1, 2, 0])
    expect(make_point(1, 2, 3)).to_be([1, 2, 3])

def test_recursion():
    expect(factorial(0)).to_be(1)
    expect(factorial(5)).to_be(120)

def test_fibonacci():
    expect(fib(0)).to_be(0)
    expect(fib(1)).to_be(1)
    expect(fib(10)).to_be(55)

def test_nested_functions():
    def outer(x):
        def inner(y):
            return y * 2
        return inner(x) + 10
    expect(outer(5)).to_be(20)

def test_lambda():
    square = lambda x: x * x
    expect(square(5)).to_be(25)

    add_lambda = lambda a, b: a + b
    expect(add_lambda(3, 4)).to_be(7)

def test_higher_order():
    def apply(func, value):
        return func(value)
    expect(apply(lambda x: x * 2, 21)).to_be(42)

def test_global_access():
    global_var = 100
    def use_global():
        return global_var
    expect(use_global()).to_be(100)

def test_early_return():
    def find_first_even(numbers):
        for n in numbers:
            if n % 2 == 0:
                return n
        return None
    expect(find_first_even([1, 3, 5, 6, 7])).to_be(6)
    expect(find_first_even([1, 3, 5, 7])).to_be(None)

def test_func_as_arg():
    def double(x):
        return x * 2
    def apply_to_list(func, lst):
        result = []
        for item in lst:
            result.append(func(item))
        return result
    expect(apply_to_list(double, [1, 2, 3, 4, 5])).to_be([2, 4, 6, 8, 10])

test("simple_function", test_simple_function)
test("function_args", test_function_args)
test("default_args", test_default_args)
test("multi_default_args", test_multi_default_args)
test("recursion", test_recursion)
test("fibonacci", test_fibonacci)
test("nested_functions", test_nested_functions)
test("lambda", test_lambda)
test("higher_order", test_higher_order)
test("global_access", test_global_access)
test("early_return", test_early_return)
test("func_as_arg", test_func_as_arg)

print("Functions tests completed")
