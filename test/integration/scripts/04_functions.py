# Test: Functions
# Tests function definitions, arguments, and recursion

results = {}

# Simple function
def simple():
    return 42

results["simple_func"] = simple()

# Function with arguments
def add(a, b):
    return a + b

results["func_args"] = add(10, 20)

# Function with default arguments
def greet(name, greeting="Hello"):
    return greeting + ", " + name

results["default_arg_used"] = greet("World")
results["default_arg_override"] = greet("World", "Hi")

# Function with multiple default arguments
def make_point(x=0, y=0, z=0):
    return [x, y, z]

results["multi_default_none"] = make_point()
results["multi_default_some"] = make_point(1, 2)
results["multi_default_all"] = make_point(1, 2, 3)

# Recursive function
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

results["factorial_0"] = factorial(0)
results["factorial_5"] = factorial(5)

# Recursive Fibonacci
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

results["fib_0"] = fib(0)
results["fib_1"] = fib(1)
results["fib_10"] = fib(10)

# Nested functions (not closures)
def outer(x):
    def inner(y):
        return y * 2
    return inner(x) + 10

results["nested_func"] = outer(5)

# Lambda expressions
square = lambda x: x * x
results["lambda_simple"] = square(5)

add_lambda = lambda a, b: a + b
results["lambda_two_args"] = add_lambda(3, 4)

# Higher-order function with lambda
def apply(func, value):
    return func(value)

results["higher_order_lambda"] = apply(lambda x: x * 2, 21)

# Global variable access
global_var = 100

def use_global():
    return global_var

results["access_global"] = use_global()

# Early return
def find_first_even(numbers):
    for n in numbers:
        if n % 2 == 0:
            return n
    return None

results["early_return_found"] = find_first_even([1, 3, 5, 6, 7])
results["early_return_none"] = find_first_even([1, 3, 5, 7])

# Function as argument
def double(x):
    return x * 2

def apply_to_list(func, lst):
    result = []
    for item in lst:
        result.append(func(item))
    return result

results["func_as_arg"] = apply_to_list(double, [1, 2, 3, 4, 5])

print("Functions tests completed")
