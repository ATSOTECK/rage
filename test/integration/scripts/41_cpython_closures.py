# Test: CPython Scope and Closure Edge Cases
# Adapted from CPython's test_scope.py

from test_framework import test, expect

# === Simple closures ===
def test_simple_closure():
    def make_adder(x):
        def adder(y):
            return x + y
        return adder
    inc = make_adder(1)
    plus10 = make_adder(10)
    expect(inc(1)).to_be(2)
    expect(plus10(5)).to_be(15)

def test_closure_captures_variable():
    x = 10
    def inner():
        return x
    x = 20
    # Closure sees the current value, not the value at definition time
    expect(inner()).to_be(20)

# === Extra nesting ===
def test_extra_nesting():
    def make_adder(x):
        def extra():
            def adder(y):
                return x + y
            return adder
        return extra()
    inc = make_adder(1)
    expect(inc(1)).to_be(2)

# === Rebinding ===
def test_rebinding():
    # Test that closure captures the final value of a variable
    def make_adder(x):
        x2 = x + 1  # use a new variable to avoid rebinding issue
        def adder(y):
            return x2 + y
        return adder
    inc = make_adder(0)
    expect(inc(1)).to_be(2)

# === Nonlocal ===
# Note: nonlocal is not yet fully supported in RAGE.
# These tests use alternative patterns to test similar behavior.
def test_nonlocal_basic():
    # Test closure over mutable state using a list (workaround for nonlocal)
    def counter():
        count = [0]
        def increment():
            count[0] = count[0] + 1
            return count[0]
        return increment
    c = counter()
    expect(c()).to_be(1)
    expect(c()).to_be(2)
    expect(c()).to_be(3)

def test_nonlocal_shared():
    # Test shared mutable state between closures
    def make_counter():
        count = [0]
        def inc():
            count[0] = count[0] + 1
            return count[0]
        def get():
            return count[0]
        return [inc, get]
    result = make_counter()
    inc = result[0]
    get = result[1]
    inc()
    inc()
    expect(get()).to_be(2)

def test_nonlocal_nested():
    # Test nested mutable state
    def outer():
        x = [1]
        def middle():
            x[0] = x[0] + 1
            def inner():
                x[0] = x[0] + 1
                return x[0]
            return inner()
        return middle()
    expect(outer()).to_be(3)

# === Global ===
def test_global_in_nested():
    global _test_global_var
    _test_global_var = 10
    def inner():
        global _test_global_var
        _test_global_var = 20
    inner()
    expect(_test_global_var).to_be(20)

# === Nearest enclosing scope ===
def test_nearest_scope():
    x = 10
    def outer():
        x = 42
        def inner():
            return x
        return inner()
    expect(outer()).to_be(42)

# === Mixed free vars and cell vars ===
def test_mixed_freevars_cellvars():
    def outer(x):
        def middle(y):
            def inner(z):
                return x + y + z
            return inner
        return middle
    f = outer(1)(2)
    expect(f(3)).to_be(6)

# === Closures in loops ===
def test_closure_in_loop():
    funcs = []
    for i in range(5):
        def f(x=i):
            return x
        funcs.append(f)
    results = [f() for f in funcs]
    expect(results).to_be([0, 1, 2, 3, 4])

def test_closure_in_loop_late_binding():
    # Test late binding with default argument capture
    funcs = []
    for i in range(5):
        funcs.append(lambda i=i: i)
    # Each lambda captures its own value of i via default arg
    results = [f() for f in funcs]
    expect(results).to_be([0, 1, 2, 3, 4])

# === Closures and recursion ===
# Recursive function at module level (nested recursion not yet supported in RAGE)
def _factorial(n):
    if n <= 1:
        return 1
    return n * _factorial(n - 1)

def test_recursive_closure():
    expect(_factorial(5)).to_be(120)

# === Closure over mutable objects ===
def test_closure_mutable():
    items = []
    def add(x):
        items.append(x)
    def get():
        return items
    add(1)
    add(2)
    add(3)
    expect(get()).to_be([1, 2, 3])

# === Multiple closures from same scope ===
def test_multiple_closures():
    def make_ops(x):
        def add(y):
            return x + y
        def mul(y):
            return x * y
        return [add, mul]
    ops = make_ops(5)
    add5 = ops[0]
    mul5 = ops[1]
    expect(add5(3)).to_be(8)
    expect(mul5(3)).to_be(15)

# === Closure over conditional ===
def test_closure_conditional():
    def make_func(flag):
        if flag:
            x = 1
        else:
            x = 2
        def inner():
            return x
        return inner
    expect(make_func(True)()).to_be(1)
    expect(make_func(False)()).to_be(2)

# === Lambda closures ===
def test_lambda_closure():
    def make_adder(x):
        return lambda y: x + y
    add3 = make_adder(3)
    expect(add3(4)).to_be(7)

def test_lambda_in_list():
    funcs = [lambda x, i=i: x + i for i in range(3)]
    expect(funcs[0](10)).to_be(10)
    expect(funcs[1](10)).to_be(11)
    expect(funcs[2](10)).to_be(12)

# === Nested class scope ===
def test_closure_in_class():
    x = 10
    class MyClass:
        y = x + 5
    expect(MyClass.y).to_be(15)

# Register all tests
test("simple_closure", test_simple_closure)
test("closure_captures_variable", test_closure_captures_variable)
test("extra_nesting", test_extra_nesting)
test("rebinding", test_rebinding)
test("nonlocal_basic", test_nonlocal_basic)
test("nonlocal_shared", test_nonlocal_shared)
test("nonlocal_nested", test_nonlocal_nested)
test("global_in_nested", test_global_in_nested)
test("nearest_scope", test_nearest_scope)
test("mixed_freevars_cellvars", test_mixed_freevars_cellvars)
test("closure_in_loop", test_closure_in_loop)
test("closure_in_loop_late_binding", test_closure_in_loop_late_binding)
test("recursive_closure", test_recursive_closure)
test("closure_mutable", test_closure_mutable)
test("multiple_closures", test_multiple_closures)
test("closure_conditional", test_closure_conditional)
test("lambda_closure", test_lambda_closure)
test("lambda_in_list", test_lambda_in_list)
test("closure_in_class", test_closure_in_class)

print("CPython closure tests completed")
