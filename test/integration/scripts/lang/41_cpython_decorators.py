# Test: CPython Decorator Patterns
# Adapted from CPython decorator tests - covers advanced decorator patterns
# beyond 12_decorators.py

from test_framework import test, expect

# =============================================================================
# Basic function decorators
# =============================================================================

_logging_calls = [0]

def logging_decorator(func):
    def wrapper(x):
        _logging_calls[0] = _logging_calls[0] + 1
        result = func(x)
        return result
    return wrapper

def preserve_return(func):
    def wrapper(x):
        return func(x)
    return wrapper

# Decorator with closure state (counter)
def counting_decorator(func):
    count = [0]
    def wrapper():
        count[0] = count[0] + 1
        return (count[0], func())
    return wrapper

# Multiple decorators stacked
def add_prefix(func):
    def wrapper(s):
        return "prefix_" + func(s)
    return wrapper

def add_suffix(func):
    def wrapper(s):
        return func(s) + "_suffix"
    return wrapper

# Decorator that modifies arguments
def double_arg(func):
    def wrapper(x):
        return func(x * 2)
    return wrapper

def negate_arg(func):
    def wrapper(x):
        return func(-x)
    return wrapper

# Identity decorator
def identity(func):
    def wrapper(x):
        return func(x)
    return wrapper

# Decorator factory (decorator that takes arguments)
def multiply_result(factor):
    def decorator(func):
        def wrapper(x):
            return func(x) * factor
        return wrapper
    return decorator

def add_result(amount):
    def decorator(func):
        def wrapper(x):
            return func(x) + amount
        return wrapper
    return decorator

# Class decorator pattern (class with __call__)
class Memoize:
    def __init__(self, func):
        self.func = func
        self.cache = {}

    def __call__(self, x):
        if x not in self.cache:
            self.cache[x] = self.func(x)
        return self.cache[x]

class CallTracker:
    def __init__(self, func):
        self.func = func
        self.calls = 0

    def __call__(self, x):
        self.calls = self.calls + 1
        return self.func(x)

# =============================================================================
# Apply decorators
# =============================================================================

@logging_decorator
def compute(x):
    return x * x

@preserve_return
def triple(x):
    return x * 3

@counting_decorator
def greet():
    return "hello"

@add_prefix
@add_suffix
def process_string(s):
    return s

@double_arg
def square(x):
    return x * x

@negate_arg
def return_val(x):
    return x

@identity
def passthrough(x):
    return x + 10

@multiply_result(3)
def base_func(x):
    return x + 1

@add_result(100)
@multiply_result(2)
def chained_factory(x):
    return x

@Memoize
def fibonacci(n):
    if n < 2:
        return n
    return fibonacci(n - 1) + fibonacci(n - 2)

@CallTracker
def tracked_func(x):
    return x + 1

# =============================================================================
# Classes with decorators
# =============================================================================

class Temperature:
    def __init__(self, celsius):
        self._celsius = celsius

    @property
    def celsius(self):
        return self._celsius

    @celsius.setter
    def celsius(self, value):
        self._celsius = value

    @property
    def fahrenheit(self):
        return self._celsius * 9 / 5 + 32

class MathHelper:
    @staticmethod
    def square(x):
        return x * x

    @staticmethod
    def cube(x):
        return x * x * x

    @classmethod
    def describe(cls):
        return cls.__name__

class Registry:
    items = []

    @classmethod
    def register(cls, item):
        cls.items.append(item)
        return len(cls.items)

    @classmethod
    def get_all(cls):
        return cls.items

    @classmethod
    def clear(cls):
        cls.items = []

# =============================================================================
# Conditional decorator application
# =============================================================================

def conditional_decorator(condition, decorator):
    def apply(func):
        if condition:
            return decorator(func)
        return func
    return apply

def double_result_dec(func):
    def wrapper(x):
        return func(x) * 2
    return wrapper

@conditional_decorator(True, double_result_dec)
def maybe_doubled(x):
    return x + 1

@conditional_decorator(False, double_result_dec)
def not_doubled(x):
    return x + 1

# =============================================================================
# Tests
# =============================================================================

def test_basic_logging_decorator():
    _logging_calls[0] = 0
    result = compute(5)
    expect(result).to_be(25)
    result2 = compute(3)
    expect(result2).to_be(9)
    expect(_logging_calls[0]).to_be(2)

def test_preserve_return():
    expect(triple(7)).to_be(21)
    expect(triple(0)).to_be(0)
    expect(triple(-3)).to_be(-9)

def test_closure_state_decorator():
    r1 = greet()
    expect(r1).to_be((1, "hello"))
    r2 = greet()
    expect(r2).to_be((2, "hello"))
    r3 = greet()
    expect(r3).to_be((3, "hello"))

def test_stacked_decorators():
    # add_prefix applied last (outermost), add_suffix applied first (innermost)
    # So: add_prefix(add_suffix(process_string))
    # process_string("test") -> "test" -> add_suffix -> "test_suffix" -> add_prefix -> "prefix_test_suffix"
    expect(process_string("test")).to_be("prefix_test_suffix")

def test_argument_modifying_decorator():
    # double_arg makes square(3) become square(6)
    expect(square(3)).to_be(36)
    expect(square(5)).to_be(100)

def test_negate_arg_decorator():
    expect(return_val(5)).to_be(-5)
    expect(return_val(-3)).to_be(3)

def test_identity_decorator():
    expect(passthrough(5)).to_be(15)
    expect(passthrough(0)).to_be(10)

def test_decorator_factory():
    # multiply_result(3) wraps base_func
    # base_func(4) = 5, then * 3 = 15
    expect(base_func(4)).to_be(15)
    expect(base_func(0)).to_be(3)

def test_chained_factory_decorators():
    # chained_factory(5): multiply_result(2) applied first -> 5*2=10, then add_result(100) -> 10+100=110
    expect(chained_factory(5)).to_be(110)
    expect(chained_factory(0)).to_be(100)

def test_class_decorator_memoize():
    # fibonacci with memoization
    expect(fibonacci(0)).to_be(0)
    expect(fibonacci(1)).to_be(1)
    expect(fibonacci(10)).to_be(55)
    expect(fibonacci(15)).to_be(610)
    # Check cache was populated
    expect(0 in fibonacci.cache).to_be(True)
    expect(10 in fibonacci.cache).to_be(True)

def test_class_decorator_call_tracker():
    tracked_func(1)
    tracked_func(2)
    tracked_func(3)
    expect(tracked_func.calls).to_be(3)
    expect(tracked_func(10)).to_be(11)
    expect(tracked_func.calls).to_be(4)

def test_property_computed():
    t = Temperature(100)
    expect(t.celsius).to_be(100)
    expect(t.fahrenheit).to_be(212.0)
    t.celsius = 0
    expect(t.celsius).to_be(0)
    expect(t.fahrenheit).to_be(32.0)

def test_staticmethod_patterns():
    expect(MathHelper.square(4)).to_be(16)
    expect(MathHelper.cube(3)).to_be(27)
    m = MathHelper()
    expect(m.square(5)).to_be(25)

def test_classmethod_patterns():
    expect(MathHelper.describe()).to_be("MathHelper")
    Registry.clear()
    Registry.register("item1")
    Registry.register("item2")
    expect(Registry.get_all()).to_be(["item1", "item2"])
    expect(Registry.register("item3")).to_be(3)
    Registry.clear()

def test_conditional_decorator_applied():
    # condition=True, so double_result_dec is applied
    expect(maybe_doubled(4)).to_be(10)  # (4+1)*2 = 10

def test_conditional_decorator_not_applied():
    # condition=False, so decorator is NOT applied
    expect(not_doubled(4)).to_be(5)  # 4+1 = 5

def test_decorator_on_lambda_equivalent():
    # Apply decorator manually (since @decorator on lambda isn't valid syntax)
    inc = double_result_dec(lambda x: x + 1)
    expect(inc(5)).to_be(12)  # (5+1)*2 = 12

def test_decorator_preserves_different_types():
    def type_preserving(func):
        def wrapper(x):
            return func(x)
        return wrapper

    @type_preserving
    def return_list(x):
        return [x, x + 1]

    @type_preserving
    def return_dict(x):
        return {"val": x}

    @type_preserving
    def return_str(x):
        return str(x) + "!"

    expect(return_list(1)).to_be([1, 2])
    expect(return_dict(42)).to_be({"val": 42})
    expect(return_str(99)).to_be("99!")

def test_decorator_with_default_args():
    def with_default(func):
        def wrapper(x, y=10):
            return func(x, y)
        return wrapper

    @with_default
    def add(x, y):
        return x + y

    expect(add(5, 3)).to_be(8)
    expect(add(5)).to_be(15)

# =============================================================================
# Run all tests
# =============================================================================

test("basic_logging_decorator", test_basic_logging_decorator)
test("preserve_return", test_preserve_return)
test("closure_state_decorator", test_closure_state_decorator)
test("stacked_decorators", test_stacked_decorators)
test("argument_modifying_decorator", test_argument_modifying_decorator)
test("negate_arg_decorator", test_negate_arg_decorator)
test("identity_decorator", test_identity_decorator)
test("decorator_factory", test_decorator_factory)
test("chained_factory_decorators", test_chained_factory_decorators)
test("class_decorator_memoize", test_class_decorator_memoize)
test("class_decorator_call_tracker", test_class_decorator_call_tracker)
test("property_computed", test_property_computed)
test("staticmethod_patterns", test_staticmethod_patterns)
test("classmethod_patterns", test_classmethod_patterns)
test("conditional_decorator_applied", test_conditional_decorator_applied)
test("conditional_decorator_not_applied", test_conditional_decorator_not_applied)
test("decorator_on_lambda_equivalent", test_decorator_on_lambda_equivalent)
test("decorator_preserves_different_types", test_decorator_preserves_different_types)
test("decorator_with_default_args", test_decorator_with_default_args)

print("CPython decorator tests completed")
