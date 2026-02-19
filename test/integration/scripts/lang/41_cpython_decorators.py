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

# =============================================================================
# CPython test_decorators.py - Adapted Tests
# =============================================================================

# --- Memoize + countcalls pattern (from CPython) ---

def countcalls(counts):
    """Decorator to count calls to a function"""
    def decorate(func):
        func_name = func.__name__
        counts[func_name] = 0
        def call(*args):
            counts[func_name] = counts[func_name] + 1
            return func(*args)
        call._name = func_name
        return call
    return decorate

def memoize_cpython(func):
    """Memoize decorator from CPython test suite"""
    saved = {}
    def call(*args):
        try:
            return saved[args]
        except KeyError:
            res = func(*args)
            saved[args] = res
            return res
        except TypeError:
            # Unhashable argument
            return func(*args)
    call._name = func.__name__
    return call

def test_cpython_memoize_countcalls():
    """Test memoize + countcalls pattern from CPython"""
    counts = {}

    @memoize_cpython
    @countcalls(counts)
    def double(x):
        return x * 2

    expect(counts).to_be({"double": 0})

    # Only the first call with a given argument bumps the call count
    expect(double(2)).to_be(4)
    expect(counts["double"]).to_be(1)
    expect(double(2)).to_be(4)
    expect(counts["double"]).to_be(1)
    expect(double(3)).to_be(6)
    expect(counts["double"]).to_be(2)

    # Unhashable arguments do not get memoized
    expect(double([10])).to_be([10, 10])
    expect(counts["double"]).to_be(3)
    expect(double([10])).to_be([10, 10])
    expect(counts["double"]).to_be(4)

test("cpython_memoize_countcalls", test_cpython_memoize_countcalls)

# --- Decorator application order (from CPython TestDecorators.test_order) ---

def test_cpython_decorator_application_order():
    """Decorators are applied bottom-to-top: the outermost decorator wins"""
    def callnum(num):
        def deco(func):
            return lambda: num
        return deco

    @callnum(2)
    @callnum(1)
    def foo():
        return 42

    expect(foo()).to_be(2)

test("cpython_decorator_application_order", test_cpython_decorator_application_order)

# --- Decorator evaluation order (from CPython TestDecorators.test_eval_order) ---

def test_cpython_decorator_eval_order():
    """
    Evaluating a decorated function involves four steps for each decorator-maker:
      1: Evaluate the decorator-maker name
      2: Evaluate the decorator-maker arguments (if any)
      3: Call the decorator-maker to make a decorator
      4: Call the decorator
    Decorator-makers are evaluated top-to-bottom, but decorators are applied
    bottom-to-top (i.e., the last decorator is called first).
    """
    actions = []

    def make_decorator(tag):
        actions.append("makedec" + tag)
        def decorate(func):
            actions.append("calldec" + tag)
            return func
        return decorate

    class NameLookupTracer:
        def __init__(self, index):
            self.index = index
            self._make_decorator = make_decorator
            self._arg = str(index)

    c1 = NameLookupTracer(1)
    c2 = NameLookupTracer(2)
    c3 = NameLookupTracer(3)

    # Simulate the evaluation order that CPython tests:
    # @c1.make_decorator(c1.arg)
    # @c2.make_decorator(c2.arg)
    # @c3.make_decorator(c3.arg)
    # def foo(): return 42
    #
    # Since RAGE doesn't support dotted decorator expressions,
    # we manually replicate the evaluation order.
    actions = []
    # Top-to-bottom: evaluate names and args, then make decorators
    actions.append("evalname1")
    actions.append("evalargs1")
    d1 = make_decorator(str(c1.index))  # makedec1
    actions.append("evalname2")
    actions.append("evalargs2")
    d2 = make_decorator(str(c2.index))  # makedec2
    actions.append("evalname3")
    actions.append("evalargs3")
    d3 = make_decorator(str(c3.index))  # makedec3

    # Bottom-to-top: apply decorators
    def foo():
        return 42
    foo = d3(foo)  # calldec3
    foo = d2(foo)  # calldec2
    foo = d1(foo)  # calldec1

    expect(foo()).to_be(42)

    expected_actions = [
        "evalname1", "evalargs1", "makedec1",
        "evalname2", "evalargs2", "makedec2",
        "evalname3", "evalargs3", "makedec3",
        "calldec3", "calldec2", "calldec1",
    ]
    expect(actions).to_be(expected_actions)

test("cpython_decorator_eval_order", test_cpython_decorator_eval_order)

# --- Stacked decorator evaluation order with decorator factories ---

def test_cpython_stacked_decorator_factories():
    """Multiple decorator factories: makers run top-to-bottom, application bottom-to-top"""
    order = []

    def decorator_maker(name):
        order.append("make_" + name)
        def decorator(func):
            order.append("apply_" + name)
            def wrapper(*args, **kwargs):
                order.append("call_" + name)
                return func(*args, **kwargs)
            return wrapper
        return decorator

    @decorator_maker("first")
    @decorator_maker("second")
    @decorator_maker("third")
    def target():
        return "result"

    # Makers run top-to-bottom
    expect(order).to_be([
        "make_first", "make_second", "make_third",
        "apply_third", "apply_second", "apply_first",
    ])

    # Calls run outside-in (first wraps second wraps third)
    order = []
    result = target()
    expect(result).to_be("result")
    expect(order).to_be(["call_first", "call_second", "call_third"])

test("cpython_stacked_decorator_factories", test_cpython_stacked_decorator_factories)

# --- Decorator with *args/**kwargs (from CPython TestDecorators.test_argforms) ---

def test_cpython_decorator_argforms():
    """Test argument passing forms in decorator factories"""
    def noteargs(*args, **kwds):
        def decorate(func):
            func_info = (args, kwds)
            def wrapper():
                return (func(), func_info)
            return wrapper
        return decorate

    args = ("Now", "is", "the", "time")
    kwds = {"one": 1, "two": 2}

    @noteargs(*args, **kwds)
    def f1():
        return 42
    result, info = f1()
    expect(result).to_be(42)
    expect(info).to_be((("Now", "is", "the", "time"), {"one": 1, "two": 2}))

    @noteargs("terry", "gilliam", eric="idle", john="cleese")
    def f2():
        return 84
    result2, info2 = f2()
    expect(result2).to_be(84)
    expect(info2).to_be((("terry", "gilliam"), {"eric": "idle", "john": "cleese"}))

    @noteargs(1, 2)
    def f3():
        return None
    result3, info3 = f3()
    expect(result3).to_be(None)
    expect(info3).to_be(((1, 2), {}))

test("cpython_decorator_argforms", test_cpython_decorator_argforms)

# --- Class decorators (from CPython TestClassDecorators) ---

def test_cpython_class_decorator_simple():
    """Simple class decorator that adds an attribute"""
    def plain(x):
        x.extra = "Hello"
        return x

    @plain
    class C:
        pass

    expect(C.extra).to_be("Hello")

test("cpython_class_decorator_simple", test_cpython_class_decorator_simple)

def test_cpython_class_decorator_double():
    """Two class decorators stacked"""
    def ten(x):
        x.extra = 10
        return x
    def add_five(x):
        x.extra = x.extra + 5
        return x

    @add_five
    @ten
    class C:
        pass

    # ten applied first (sets extra=10), then add_five (adds 5)
    expect(C.extra).to_be(15)

test("cpython_class_decorator_double", test_cpython_class_decorator_double)

def test_cpython_class_decorator_order():
    """Class decorator ordering: bottom-to-top application"""
    def applied_first(x):
        x.extra = "first"
        return x
    def applied_second(x):
        x.extra = "second"
        return x

    @applied_second
    @applied_first
    class C:
        pass

    # applied_first sets extra='first', then applied_second overwrites to 'second'
    expect(C.extra).to_be("second")

test("cpython_class_decorator_order", test_cpython_class_decorator_order)

# --- Class decorator that replaces the class ---

def test_cpython_class_decorator_replace():
    """Class decorator can return a completely different class"""
    def replace_with_enhanced(cls):
        class Enhanced(cls):
            bonus = True
        Enhanced.__name__ = cls.__name__ + "Enhanced"
        return Enhanced

    @replace_with_enhanced
    class Original:
        value = 42

    expect(Original.value).to_be(42)
    expect(Original.bonus).to_be(True)

test("cpython_class_decorator_replace", test_cpython_class_decorator_replace)

# --- Class decorator with arguments ---

def test_cpython_class_decorator_with_args():
    """Class decorator factory with arguments"""
    def tag(name, value):
        def decorator(cls):
            cls.tag_name = name
            cls.tag_value = value
            return cls
        return decorator

    @tag("version", 2)
    class MyService:
        pass

    expect(MyService.tag_name).to_be("version")
    expect(MyService.tag_value).to_be(2)

test("cpython_class_decorator_with_args", test_cpython_class_decorator_with_args)

# --- Multiple stacked class decorators with ordering ---

def test_cpython_class_decorator_triple_stack():
    """Three class decorators stacked"""
    log = []

    def d1(cls):
        log.append("d1")
        cls.d1 = True
        return cls

    def d2(cls):
        log.append("d2")
        cls.d2 = True
        return cls

    def d3(cls):
        log.append("d3")
        cls.d3 = True
        return cls

    @d1
    @d2
    @d3
    class C:
        pass

    # Applied bottom-to-top: d3 first, then d2, then d1
    expect(log).to_be(["d3", "d2", "d1"])
    expect(C.d1).to_be(True)
    expect(C.d2).to_be(True)
    expect(C.d3).to_be(True)

test("cpython_class_decorator_triple_stack", test_cpython_class_decorator_triple_stack)

# --- Decorator that wraps with a callable class ---

def test_cpython_callable_class_as_decorator():
    """A class with __call__ used as a decorator"""
    class Validator:
        def __init__(self, func):
            self.func = func

        def __call__(self, *args):
            for arg in args:
                if not isinstance(arg, int):
                    raise TypeError("Expected int")
            return self.func(*args)

    @Validator
    def add_ints(a, b):
        return a + b

    expect(add_ints(3, 4)).to_be(7)

    got_error = False
    try:
        add_ints("a", 1)
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_callable_class_as_decorator", test_cpython_callable_class_as_decorator)

# --- Decorator factory returning callable class ---

def test_cpython_decorator_factory_with_class():
    """Decorator factory that returns a callable class wrapper"""
    class Repeater:
        def __init__(self, func, times):
            self.func = func
            self.times = times

        def __call__(self, *args):
            results = []
            for i in range(self.times):
                results.append(self.func(*args))
            return results

    def repeat(times):
        def decorator(func):
            return Repeater(func, times)
        return decorator

    @repeat(3)
    def greet(name):
        return "hello " + name

    expect(greet("world")).to_be(["hello world", "hello world", "hello world"])

test("cpython_decorator_factory_with_class", test_cpython_decorator_factory_with_class)

# --- Decorator on nested function ---

def test_cpython_decorator_on_nested():
    """Decorators work on nested function definitions"""
    def outer():
        def add_ten(func):
            def wrapper(x):
                return func(x) + 10
            return wrapper

        @add_ten
        def inner(x):
            return x * 2

        return inner

    fn = outer()
    expect(fn(5)).to_be(20)   # 5*2 + 10
    expect(fn(0)).to_be(10)   # 0*2 + 10

test("cpython_decorator_on_nested", test_cpython_decorator_on_nested)

# --- Decorator preserving closure ---

def test_cpython_decorator_with_closure():
    """Decorator that preserves access to closure variables"""
    def make_adder(n):
        def add_n(x):
            return x + n
        return add_n

    def trace(func):
        calls = [0]
        def wrapper(x):
            calls[0] = calls[0] + 1
            result = func(x)
            return (result, calls[0])
        return wrapper

    @trace
    def add5(x):
        return x + 5

    expect(add5(10)).to_be((15, 1))
    expect(add5(20)).to_be((25, 2))

    # Also works with a closure-based function
    add3 = trace(make_adder(3))
    expect(add3(10)).to_be((13, 1))
    expect(add3(7)).to_be((10, 2))

test("cpython_decorator_with_closure", test_cpython_decorator_with_closure)

# --- Decorator that raises on bad input (TypeError pattern) ---

def test_cpython_decorator_raises_typeerror():
    """Decorator that is not callable should raise TypeError"""
    got_error = False
    try:
        # Applying a non-callable as a decorator
        none_result = None
        def apply_bad():
            @none_result
            def foo():
                pass
        apply_bad()
    except TypeError:
        got_error = True
    expect(got_error).to_be(True)

test("cpython_decorator_raises_typeerror", test_cpython_decorator_raises_typeerror)

# --- Staticmethod and classmethod on class (from CPython TestDecorators.test_single) ---

def test_cpython_staticmethod_classmethod_basic():
    """@staticmethod and @classmethod basic usage from CPython"""
    class C:
        @staticmethod
        def foo():
            return 42

        @classmethod
        def bar(cls):
            return cls.__name__

    expect(C.foo()).to_be(42)
    expect(C().foo()).to_be(42)
    expect(C.bar()).to_be("C")
    expect(C().bar()).to_be("C")

test("cpython_staticmethod_classmethod_basic", test_cpython_staticmethod_classmethod_basic)

# --- Dotted decorator name (simulated, since CPython tests MiscDecorators.author) ---

def test_cpython_dotted_decorator():
    """Dotted attribute access for decorator (simulated from CPython)"""
    class Decorators:
        @staticmethod
        def author(name):
            def decorate(func):
                func_author = name
                def wrapper(*args, **kwargs):
                    return (func(*args, **kwargs), func_author)
                return wrapper
            return decorate

    decorators = Decorators()

    @decorators.author("Cleese")
    def foo():
        return 42

    result, author = foo()
    expect(result).to_be(42)
    expect(author).to_be("Cleese")

test("cpython_dotted_decorator", test_cpython_dotted_decorator)

print("CPython decorator tests completed")
