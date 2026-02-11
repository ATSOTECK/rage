from test_framework import test, expect

# Basic nonlocal counter
def make_counter():
    count = 0
    def increment():
        nonlocal count
        count = count + 1
        return count
    return increment

c = make_counter()
test("nonlocal counter 1", lambda: expect(c(), 1))
test("nonlocal counter 2", lambda: expect(c(), 2))
test("nonlocal counter 3", lambda: expect(c(), 3))

# Multiple nonlocals
def multi():
    x = 10
    y = 20
    def inner():
        nonlocal x, y
        x = x + 1
        y = y + 2
        return (x, y)
    return inner

f = multi()
test("multiple nonlocals first call", lambda: expect(f(), (11, 22)))
test("multiple nonlocals second call", lambda: expect(f(), (12, 24)))

# Nested nonlocal (3 levels deep)
def outer():
    x = 0
    def middle():
        def inner():
            nonlocal x
            x = x + 10
            return x
        return inner
    return middle

f = outer()()
test("nested nonlocal first", lambda: expect(f(), 10))
test("nested nonlocal second", lambda: expect(f(), 20))

# nonlocal with other parameters
def make_adder():
    total = 0
    def add(n):
        nonlocal total
        total = total + n
        return total
    return add

adder = make_adder()
test("nonlocal with param 1", lambda: expect(adder(5), 5))
test("nonlocal with param 2", lambda: expect(adder(10), 15))
test("nonlocal with param 3", lambda: expect(adder(3), 18))

# nonlocal boolean toggle
def toggle():
    state = False
    def flip():
        nonlocal state
        state = not state
        return state
    return flip

t = toggle()
test("nonlocal toggle true", lambda: expect(t(), True))
test("nonlocal toggle false", lambda: expect(t(), False))
test("nonlocal toggle true again", lambda: expect(t(), True))

# Two closures sharing nonlocal state
def shared():
    value = 0
    def inc():
        nonlocal value
        value = value + 1
        return value
    def get():
        nonlocal value
        return value
    return inc, get

inc, get = shared()
inc()
inc()
inc()
test("shared nonlocal via get", lambda: expect(get(), 3))
test("shared nonlocal via inc", lambda: expect(inc(), 4))

# nonlocal with default parameter
def with_default():
    x = 100
    def inner(amount=5):
        nonlocal x
        x = x - amount
        return x
    return inner

f = with_default()
test("nonlocal with default used", lambda: expect(f(), 95))
test("nonlocal with default overridden", lambda: expect(f(10), 85))

# Independent closures don't share state
c1 = make_counter()
c2 = make_counter()
c1()
c1()
c1()
test("independent closure 1", lambda: expect(c1(), 4))
test("independent closure 2", lambda: expect(c2(), 1))

# nonlocal string mutation
def make_greeting():
    msg = "hello"
    def update(name):
        nonlocal msg
        msg = "hello, " + name
        return msg
    def get():
        nonlocal msg
        return msg
    return update, get

update, get = make_greeting()
test("nonlocal string update", lambda: expect(update("world"), "hello, world"))
test("nonlocal string get", lambda: expect(get(), "hello, world"))

# nonlocal in loop
def make_accumulator():
    items = []
    def add(item):
        nonlocal items
        items = items + [item]
        return items
    return add

acc = make_accumulator()
acc(1)
acc(2)
test("nonlocal accumulator", lambda: expect(acc(3), [1, 2, 3]))
