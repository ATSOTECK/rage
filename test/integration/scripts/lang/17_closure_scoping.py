from test_framework import test, expect

# UnboundLocalError: assignment makes variable local
def test_unbound():
    x = 10
    def inner():
        x = x + 1  # No nonlocal — x is local, read before assignment
        return x
    return inner

caught = False
try:
    test_unbound()()
except:
    caught = True
test("assignment without nonlocal raises error", lambda: expect(caught, True))

# With nonlocal, rebinding works
def test_nonlocal_rebind():
    x = 0
    def inc():
        nonlocal x
        x = x + 1
        return x
    return inc

c = test_nonlocal_rebind()
test("nonlocal rebind call 1", lambda: expect(c(), 1))
test("nonlocal rebind call 2", lambda: expect(c(), 2))
test("nonlocal rebind call 3", lambda: expect(c(), 3))

# Read-only closure doesn't need nonlocal
def test_readonly():
    val = 42
    def get():
        return val
    return get

test("read-only closure works", lambda: expect(test_readonly()(), 42))

# Simple unbound local in function
def test_simple_unbound():
    caught = False
    def f():
        y = y + 1
    try:
        f()
    except:
        caught = True
    return caught

test("simple unbound local", lambda: expect(test_simple_unbound(), True))

# for loop variable is properly local
def test_for_local():
    result = []
    for i in range(3):
        result = result + [i]
    return result

test("for loop variable is local", lambda: expect(test_for_local(), [0, 1, 2]))

# import inside function works
def test_import():
    import math
    return math.sqrt(16)

test("import inside function", lambda: expect(test_import(), 4.0))

# from import inside function works
def test_from_import():
    from math import floor
    return floor(3.7)

test("from import inside function", lambda: expect(test_from_import(), 3))

# Assignment in except handler
def test_except_local():
    def f():
        try:
            x = 1 / 0
        except:
            x = -1
        return x
    return f()

test("except handler assignment", lambda: expect(test_except_local(), -1))

# Global declaration prevents local treatment
x_global = 100
def test_global():
    global x_global
    x_global = 200
    return x_global

test("global declaration", lambda: expect(test_global(), 200))
test("global actually modified", lambda: expect(x_global, 200))

# Multiple closures with nonlocal sharing state
def test_shared_nonlocal():
    value = 0
    def inc():
        nonlocal value
        value = value + 1
    def get():
        return value
    inc()
    inc()
    inc()
    return get()

test("shared nonlocal state", lambda: expect(test_shared_nonlocal(), 3))
