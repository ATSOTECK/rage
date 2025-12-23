# Test: Decorators and Closures
# Tests decorator syntax, closures, and nested functions

results = {}

# Basic closure
def make_counter():
    count = 0
    def counter():
        return count
    return counter

c = make_counter()
results["basic_closure"] = c()

# Closure with outer variable
def outer_with_val(x):
    def inner():
        return x * 2
    return inner

f = outer_with_val(21)
results["closure_captures_param"] = f()

# Nested closures
def outer_nested(x):
    def middle():
        def inner():
            return x
        return inner
    return middle

fn = outer_nested(42)()()
results["nested_closure"] = fn

# Basic decorator
def double_result(func):
    def wrapper():
        return func() * 2
    return wrapper

@double_result
def get_five():
    return 5

results["basic_decorator"] = get_five()

# Decorator with wrapped function args
def log_args(func):
    def wrapper(a, b):
        return func(a, b)
    return wrapper

@log_args
def add(a, b):
    return a + b

results["decorator_with_args"] = add(10, 20)

# Multiple decorators
def add_one(func):
    def wrapper():
        return func() + 1
    return wrapper

def double(func):
    def wrapper():
        return func() * 2
    return wrapper

@add_one
@double
def get_three():
    return 3

# (3 * 2) + 1 = 7
results["multiple_decorators"] = get_three()

# Decorator factory (decorator with arguments)
def repeat(n):
    def decorator(func):
        def wrapper():
            result = []
            for i in range(n):
                result.append(func())
            return result
        return wrapper
    return decorator

@repeat(3)
def say_hi():
    return "hi"

results["decorator_factory"] = say_hi()

# Decorator that wraps return value
def make_list(func):
    def wrapper():
        return [func()]
    return wrapper

@make_list
def get_value():
    return 42

results["wrapper_modifies_result"] = get_value()

# Closure with mutable state (via list)
def make_accumulator():
    total = [0]  # Use list for mutable state
    def add(x):
        total[0] = total[0] + x
        return total[0]
    return add

acc = make_accumulator()
acc(5)
acc(10)
results["closure_mutable_state"] = acc(3)

# Decorator that preserves function behavior
def identity_decorator(func):
    def wrapper(x):
        return func(x)
    return wrapper

@identity_decorator
def square(x):
    return x * x

results["identity_decorator"] = square(7)

print("Decorators tests completed")
