# Test: CPython Lambda Edge Cases
# Adapted from CPython's test_syntax.py and test_grammar.py lambda tests

from test_framework import test, expect

# === Lambda with default arguments ===
def test_lambda_default_args():
    f = lambda x, y=10: x + y
    expect(f(1)).to_be(11)
    expect(f(1, 2)).to_be(3)
    expect(f(5, 5)).to_be(10)

def test_lambda_multiple_defaults():
    f = lambda a=1, b=2, c=3: a + b + c
    expect(f()).to_be(6)
    expect(f(10)).to_be(15)
    expect(f(10, 20)).to_be(33)
    expect(f(10, 20, 30)).to_be(60)

# === Nested lambdas ===
def test_nested_lambda():
    f = lambda x: lambda y: x + y
    add5 = f(5)
    expect(add5(3)).to_be(8)
    expect(add5(10)).to_be(15)
    expect(f(1)(2)).to_be(3)

def test_deeply_nested_lambda():
    f = lambda x: lambda y: lambda z: x + y + z
    expect(f(1)(2)(3)).to_be(6)
    expect(f(10)(20)(30)).to_be(60)

# === Lambda as arguments to builtins ===
def test_lambda_with_map():
    result = list(map(lambda x: x * 2, [1, 2, 3, 4]))
    expect(result).to_be([2, 4, 6, 8])

def test_lambda_with_filter():
    result = list(filter(lambda x: x > 3, [1, 2, 3, 4, 5, 6]))
    expect(result).to_be([4, 5, 6])

def test_lambda_with_sorted():
    data = [[3, "c"], [1, "a"], [2, "b"]]
    result = sorted(data, key=lambda item: item[0])
    expect(result).to_be([[1, "a"], [2, "b"], [3, "c"]])

def test_lambda_with_sorted_reverse():
    nums = [3, 1, 4, 1, 5, 9]
    result = sorted(nums, key=lambda x: -x)
    expect(result).to_be([9, 5, 4, 3, 1, 1])

# === Lambda returning lambda ===
def test_lambda_returning_lambda():
    compose = lambda f, g: lambda x: f(g(x))
    double = lambda x: x * 2
    inc = lambda x: x + 1
    double_then_inc = compose(inc, double)
    expect(double_then_inc(5)).to_be(11)
    inc_then_double = compose(double, inc)
    expect(inc_then_double(5)).to_be(12)

# === Lambda with conditional expressions ===
def test_lambda_conditional():
    absolute = lambda x: x if x >= 0 else -x
    expect(absolute(5)).to_be(5)
    expect(absolute(-5)).to_be(5)
    expect(absolute(0)).to_be(0)

def test_lambda_conditional_nested():
    classify = lambda x: "positive" if x > 0 else ("zero" if x == 0 else "negative")
    expect(classify(5)).to_be("positive")
    expect(classify(0)).to_be("zero")
    expect(classify(-3)).to_be("negative")

# === Lambda capturing variables ===
def test_lambda_capture_loop_default():
    # Use default arg to capture current value
    funcs = []
    for i in range(5):
        funcs.append(lambda i=i: i * i)
    expect(funcs[0]()).to_be(0)
    expect(funcs[1]()).to_be(1)
    expect(funcs[2]()).to_be(4)
    expect(funcs[3]()).to_be(9)
    expect(funcs[4]()).to_be(16)

def test_lambda_capture_outer():
    x = 100
    f = lambda: x
    expect(f()).to_be(100)
    x = 200
    # Lambda sees the current value of x (late binding)
    expect(f()).to_be(200)

# === Lambda in list comprehensions with default arg capture ===
def test_lambda_in_comprehension():
    funcs = [lambda x, n=n: x + n for n in range(4)]
    expect(funcs[0](10)).to_be(10)
    expect(funcs[1](10)).to_be(11)
    expect(funcs[2](10)).to_be(12)
    expect(funcs[3](10)).to_be(13)

def test_lambda_comprehension_squared():
    squares = [lambda n=n: n ** 2 for n in range(5)]
    results = [f() for f in squares]
    expect(results).to_be([0, 1, 4, 9, 16])

# === Immediately invoked lambda ===
def test_iife_lambda():
    result = (lambda: 42)()
    expect(result).to_be(42)

def test_iife_lambda_with_args():
    result = (lambda x, y: x * y)(6, 7)
    expect(result).to_be(42)

# === Lambda with multiple args ===
def test_lambda_multiple_args():
    f = lambda a, b, c: a * b + c
    expect(f(2, 3, 4)).to_be(10)
    expect(f(0, 100, 1)).to_be(1)

# === Lambda returning None ===
def test_lambda_returns_none():
    f = lambda: None
    expect(f()).to_be(None)
    expect(f() is None).to_be(True)

def test_lambda_side_effect_returns_none():
    # list.append returns None
    lst = []
    f = lambda x: lst.append(x)
    result = f(42)
    expect(result).to_be(None)
    expect(lst).to_be([42])

# Register all tests
test("lambda_default_args", test_lambda_default_args)
test("lambda_multiple_defaults", test_lambda_multiple_defaults)
test("nested_lambda", test_nested_lambda)
test("deeply_nested_lambda", test_deeply_nested_lambda)
test("lambda_with_map", test_lambda_with_map)
test("lambda_with_filter", test_lambda_with_filter)
test("lambda_with_sorted", test_lambda_with_sorted)
test("lambda_with_sorted_reverse", test_lambda_with_sorted_reverse)
test("lambda_returning_lambda", test_lambda_returning_lambda)
test("lambda_conditional", test_lambda_conditional)
test("lambda_conditional_nested", test_lambda_conditional_nested)
test("lambda_capture_loop_default", test_lambda_capture_loop_default)
test("lambda_capture_outer", test_lambda_capture_outer)
test("lambda_in_comprehension", test_lambda_in_comprehension)
test("lambda_comprehension_squared", test_lambda_comprehension_squared)
test("iife_lambda", test_iife_lambda)
test("iife_lambda_with_args", test_iife_lambda_with_args)
test("lambda_multiple_args", test_lambda_multiple_args)
test("lambda_returns_none", test_lambda_returns_none)
test("lambda_side_effect_returns_none", test_lambda_side_effect_returns_none)

print("CPython lambda tests completed")
