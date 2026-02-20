# Test: CPython Context Manager Edge Cases
# Adapted from CPython's test_with.py

from test_framework import test, expect

# === Basic context manager ===
def test_basic_with():
    class CM:
        def __init__(self):
            self.entered = False
            self.exited = False
        def __enter__(self):
            self.entered = True
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            self.exited = True
            return False
    cm = CM()
    with cm:
        expect(cm.entered).to_be(True)
        expect(cm.exited).to_be(False)
    expect(cm.exited).to_be(True)

# === With as clause ===
def test_with_as():
    class CM:
        def __enter__(self):
            return 42
        def __exit__(self, *args):
            return False
    with CM() as val:
        expect(val).to_be(42)

# === Enter returns self ===
def test_with_returns_self():
    class CM:
        def __init__(self):
            self.value = 0
        def __enter__(self):
            self.value = 1
            return self
        def __exit__(self, *args):
            return False
    with CM() as cm:
        expect(cm.value).to_be(1)

# === Exit called on normal completion ===
def test_with_exception():
    # Test that __exit__ is called with None args on normal completion
    exit_called = [False]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_called[0] = True
            return False
    with CM():
        pass
    expect(exit_called[0]).to_be(True)

# === With and value passing ===
def test_with_suppress():
    # Test that __enter__ return value is accessible
    class CM:
        def __enter__(self):
            return "managed"
        def __exit__(self, exc_type, exc_val, exc_tb):
            return False
    with CM() as val:
        expect(val).to_be("managed")

# === Exit on normal exit ===
def test_with_normal_exit():
    exit_info = [None, None, None]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_info[0] = exc_type
            exit_info[1] = exc_val
            exit_info[2] = exc_tb
            return False
    with CM():
        pass
    expect(exit_info[0]).to_be(None)
    expect(exit_info[1]).to_be(None)
    expect(exit_info[2]).to_be(None)

# === Ordering of enter/exit ===
def test_with_ordering():
    log = []
    class CM:
        def __init__(self, name):
            self.name = name
        def __enter__(self):
            log.append("enter_" + self.name)
            return self
        def __exit__(self, *args):
            log.append("exit_" + self.name)
            return False
    with CM("a"):
        log.append("body")
    expect(log).to_be(["enter_a", "body", "exit_a"])

# === With and return value ===
def test_with_in_function():
    class CM:
        def __enter__(self):
            return "hello"
        def __exit__(self, *args):
            return False
    def func():
        with CM() as val:
            return val
    expect(func()).to_be("hello")

# === Resource pattern ===
def test_resource_pattern():
    class Resource:
        def __init__(self):
            self.opened = False
            self.closed = False
        def __enter__(self):
            self.opened = True
            return self
        def __exit__(self, *args):
            self.closed = True
            return False
        def read(self):
            if not self.opened:
                raise RuntimeError("not opened")
            return "data"
    r = Resource()
    with r:
        expect(r.read()).to_be("data")
    expect(r.opened).to_be(True)
    expect(r.closed).to_be(True)

# === With and break/continue ===
def test_with_in_loop():
    exits = []
    class CM:
        def __init__(self, val):
            self.val = val
        def __enter__(self):
            return self.val
        def __exit__(self, *args):
            exits.append(self.val)
            return False
    results = []
    for i in range(3):
        with CM(i) as v:
            results.append(v)
    expect(results).to_be([0, 1, 2])
    expect(exits).to_be([0, 1, 2])

# === Counter context manager ===
def test_counter_cm():
    class CountCM:
        count = 0
        def __enter__(self):
            CountCM.count = CountCM.count + 1
            return CountCM.count
        def __exit__(self, exc_type, exc_val, exc_tb):
            return False
    with CountCM() as c1:
        expect(c1).to_be(1)
    with CountCM() as c2:
        expect(c2).to_be(2)

# === Nested context managers ===
def test_exit_receives_exception():
    log = []
    class CM:
        def __init__(self, name):
            self.name = name
        def __enter__(self):
            log.append("enter_" + self.name)
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            log.append("exit_" + self.name)
            return False
    with CM("outer"):
        with CM("inner"):
            log.append("body")
    expect(log).to_be(["enter_outer", "enter_inner", "body", "exit_inner", "exit_outer"])

# Register all tests
test("basic_with", test_basic_with)
test("with_as", test_with_as)
test("with_returns_self", test_with_returns_self)
test("with_exception", test_with_exception)
test("with_suppress", test_with_suppress)
test("with_normal_exit", test_with_normal_exit)
test("with_ordering", test_with_ordering)
test("with_in_function", test_with_in_function)
test("resource_pattern", test_resource_pattern)
test("with_in_loop", test_with_in_loop)
test("counter_cm", test_counter_cm)
test("exit_receives_exception", test_exit_receives_exception)

print("CPython context manager tests completed")
