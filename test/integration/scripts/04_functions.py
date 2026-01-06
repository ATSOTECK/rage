# Test: Functions
# Tests function definitions, arguments, and recursion

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
    expect(42, simple())

def test_function_args():
    def add(a, b):
        return a + b
    expect(30, add(10, 20))

def test_default_args():
    def greet(name, greeting="Hello"):
        return greeting + ", " + name
    expect("Hello, World", greet("World"))
    expect("Hi, World", greet("World", "Hi"))

def test_multi_default_args():
    def make_point(x=0, y=0, z=0):
        return [x, y, z]
    expect([0, 0, 0], make_point())
    expect([1, 2, 0], make_point(1, 2))
    expect([1, 2, 3], make_point(1, 2, 3))

def test_recursion():
    expect(1, factorial(0))
    expect(120, factorial(5))

def test_fibonacci():
    expect(0, fib(0))
    expect(1, fib(1))
    expect(55, fib(10))

def test_nested_functions():
    def outer(x):
        def inner(y):
            return y * 2
        return inner(x) + 10
    expect(20, outer(5))

def test_lambda():
    square = lambda x: x * x
    expect(25, square(5))

    add_lambda = lambda a, b: a + b
    expect(7, add_lambda(3, 4))

def test_higher_order():
    def apply(func, value):
        return func(value)
    expect(42, apply(lambda x: x * 2, 21))

def test_global_access():
    global_var = 100
    def use_global():
        return global_var
    expect(100, use_global())

def test_early_return():
    def find_first_even(numbers):
        for n in numbers:
            if n % 2 == 0:
                return n
        return None
    expect(6, find_first_even([1, 3, 5, 6, 7]))
    expect(None, find_first_even([1, 3, 5, 7]))

def test_func_as_arg():
    def double(x):
        return x * 2
    def apply_to_list(func, lst):
        result = []
        for item in lst:
            result.append(func(item))
        return result
    expect([2, 4, 6, 8, 10], apply_to_list(double, [1, 2, 3, 4, 5]))

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
