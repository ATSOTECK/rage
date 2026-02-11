from test_framework import test, expect

# Basic *args
def f(*args):
    return args

test("*args with no arguments", lambda: expect(f(), ()))
test("*args with one argument", lambda: expect(f(1), (1,)))
test("*args with multiple arguments", lambda: expect(f(1, 2, 3), (1, 2, 3)))

# *args type
test("*args is a tuple", lambda: expect(type(f(1, 2)).__name__, "tuple"))

# *args with positional params
def g(a, b, *args):
    return (a, b, args)

test("positional + *args no extras", lambda: expect(g(1, 2), (1, 2, ())))
test("positional + *args with extras", lambda: expect(g(1, 2, 3, 4), (1, 2, (3, 4))))

# *args with defaults
def h(a, b=10, *args):
    return (a, b, args)

test("*args with defaults used", lambda: expect(h(1), (1, 10, ())))
test("*args with defaults overridden", lambda: expect(h(1, 2, 3, 4), (1, 2, (3, 4))))

# *args length
def count(*args):
    return len(args)

test("len of empty *args", lambda: expect(count(), 0))
test("len of *args", lambda: expect(count(1, 2, 3), 3))

# *args iteration
def sum_all(*args):
    total = 0
    for x in args:
        total = total + x
    return total

test("iterate over *args", lambda: expect(sum_all(1, 2, 3, 4, 5), 15))

# *args indexing
def first_last(*args):
    return (args[0], args[-1])

test("index *args", lambda: expect(first_last(10, 20, 30), (10, 30)))

# *args slicing
def middle(*args):
    return args[1:-1]

test("slice *args", lambda: expect(middle(1, 2, 3, 4, 5), (2, 3, 4)))

# **kwargs basic
def kw(**kwargs):
    return kwargs

test("**kwargs with no arguments", lambda: expect(kw(), {}))
test("**kwargs with arguments", lambda: expect(kw(x=1, y=2), {"x": 1, "y": 2}))

# Combined *args and **kwargs
def both(*args, **kwargs):
    return (args, kwargs)

test("*args and **kwargs together", lambda: expect(both(1, 2, x=3), ((1, 2), {"x": 3})))
test("*args and **kwargs empty", lambda: expect(both(), ((), {})))

# Positional + *args + **kwargs
def full(a, b, *args, **kwargs):
    return (a, b, args, kwargs)

test("full signature", lambda: expect(full(1, 2, 3, 4, x=5), (1, 2, (3, 4), {"x": 5})))
test("full signature minimal", lambda: expect(full(1, 2), (1, 2, (), {})))

# *args forwarding
def inner(a, b, c):
    return a + b + c

def outer(*args):
    return inner(args[0], args[1], args[2])

test("*args forwarding", lambda: expect(outer(10, 20, 30), 60))

# *args in method
class MyClass:
    def method(self, *args):
        return args

obj = MyClass()
test("*args in method", lambda: expect(obj.method(1, 2, 3), (1, 2, 3)))
test("*args in method empty", lambda: expect(obj.method(), ()))
