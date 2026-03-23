from test_framework import test, expect
import contextlib

# =====================================================
# contextlib.contextmanager
# =====================================================

def test_contextmanager_basic():
    @contextlib.contextmanager
    def my_ctx():
        yield 42

    with my_ctx() as val:
        expect(val).to_be(42)

test("contextmanager basic yield value", test_contextmanager_basic)

def test_contextmanager_setup_teardown():
    log = []

    @contextlib.contextmanager
    def tracked():
        log.append("enter")
        yield "value"
        log.append("exit")

    with tracked() as v:
        log.append("body")
        expect(v).to_be("value")

    expect(log).to_be(["enter", "body", "exit"])

test("contextmanager setup/teardown order", test_contextmanager_setup_teardown)

def test_contextmanager_with_args():
    @contextlib.contextmanager
    def greet(name):
        yield "Hello, " + name + "!"

    with greet("World") as msg:
        expect(msg).to_be("Hello, World!")

test("contextmanager with arguments", test_contextmanager_with_args)

def test_contextmanager_exception_in_body():
    log = []

    @contextlib.contextmanager
    def cleanup():
        log.append("enter")
        try:
            yield
        finally:
            log.append("cleanup")

    caught = False
    try:
        with cleanup():
            log.append("body")
            raise ValueError("boom")
    except ValueError:
        caught = True

    expect(caught).to_be(True)
    expect(log).to_be(["enter", "body", "cleanup"])

test("contextmanager cleanup on exception", test_contextmanager_exception_in_body)

def test_contextmanager_suppress_exception():
    @contextlib.contextmanager
    def suppressor():
        try:
            yield
        except ValueError:
            pass  # Swallow the exception

    # This should NOT raise
    with suppressor():
        raise ValueError("suppressed")

    # If we get here, the exception was suppressed
    expect(True).to_be(True)

test("contextmanager suppresses exception", test_contextmanager_suppress_exception)

def test_contextmanager_yield_none():
    @contextlib.contextmanager
    def no_value():
        yield

    with no_value() as v:
        expect(v).to_be(None)

test("contextmanager yield None", test_contextmanager_yield_none)

def test_contextmanager_try_finally():
    cleaned = []

    @contextlib.contextmanager
    def resource():
        cleaned.append("acquired")
        try:
            yield "resource"
        finally:
            cleaned.append("released")

    with resource() as r:
        expect(r).to_be("resource")

    expect(cleaned).to_be(["acquired", "released"])

test("contextmanager try/finally pattern", test_contextmanager_try_finally)

def test_contextmanager_nested():
    log = []

    @contextlib.contextmanager
    def outer():
        log.append("outer_enter")
        yield "outer"
        log.append("outer_exit")

    @contextlib.contextmanager
    def inner():
        log.append("inner_enter")
        yield "inner"
        log.append("inner_exit")

    with outer() as o:
        with inner() as i:
            log.append("body")
            expect(o).to_be("outer")
            expect(i).to_be("inner")

    expect(log).to_be(["outer_enter", "inner_enter", "body", "inner_exit", "outer_exit"])

test("contextmanager nested", test_contextmanager_nested)

# =====================================================
# contextlib.closing
# =====================================================

def test_closing_basic():
    class Resource:
        def __init__(self):
            self.closed = False
        def close(self):
            self.closed = True

    r = Resource()
    with contextlib.closing(r) as val:
        expect(val.closed).to_be(False)

    expect(r.closed).to_be(True)

test("closing calls close() on exit", test_closing_basic)

def test_closing_exception():
    class Resource:
        def __init__(self):
            self.closed = False
        def close(self):
            self.closed = True

    r = Resource()
    caught = False
    try:
        with contextlib.closing(r):
            raise ValueError("error")
    except ValueError:
        caught = True

    expect(caught).to_be(True)
    expect(r.closed).to_be(True)

test("closing calls close() even on exception", test_closing_exception)

# =====================================================
# contextlib.suppress
# =====================================================

def test_suppress_single():
    with contextlib.suppress(ValueError):
        raise ValueError("suppressed")
    expect(True).to_be(True)

test("suppress single exception type", test_suppress_single)

def test_suppress_multiple():
    with contextlib.suppress(ValueError, TypeError):
        raise TypeError("also suppressed")
    expect(True).to_be(True)

test("suppress multiple exception types", test_suppress_multiple)

def test_suppress_no_exception():
    with contextlib.suppress(ValueError):
        x = 42
    expect(x).to_be(42)

test("suppress with no exception raised", test_suppress_no_exception)

def test_suppress_wrong_exception():
    caught = False
    try:
        with contextlib.suppress(ValueError):
            raise TypeError("not suppressed")
    except TypeError:
        caught = True
    expect(caught).to_be(True)

test("suppress does not suppress non-matching exception", test_suppress_wrong_exception)

def test_suppress_subclass():
    # KeyError is a subclass of LookupError
    with contextlib.suppress(LookupError):
        raise KeyError("suppressed via parent")
    expect(True).to_be(True)

test("suppress matches exception subclasses", test_suppress_subclass)

def test_suppress_empty():
    # suppress() with no arguments suppresses nothing
    caught = False
    try:
        with contextlib.suppress():
            raise ValueError("not suppressed")
    except ValueError:
        caught = True
    expect(caught).to_be(True)

test("suppress with no args suppresses nothing", test_suppress_empty)

# =====================================================
# contextlib.nullcontext
# =====================================================

def test_nullcontext_default():
    with contextlib.nullcontext() as val:
        expect(val).to_be(None)

test("nullcontext default enter_result is None", test_nullcontext_default)

def test_nullcontext_with_value():
    with contextlib.nullcontext(42) as val:
        expect(val).to_be(42)

test("nullcontext with positional enter_result", test_nullcontext_with_value)

def test_nullcontext_kwarg():
    with contextlib.nullcontext(enter_result="hello") as val:
        expect(val).to_be("hello")

test("nullcontext with keyword enter_result", test_nullcontext_kwarg)

def test_nullcontext_no_suppress():
    caught = False
    try:
        with contextlib.nullcontext():
            raise ValueError("not suppressed")
    except ValueError:
        caught = True
    expect(caught).to_be(True)

test("nullcontext does not suppress exceptions", test_nullcontext_no_suppress)

# =====================================================
# contextlib.ExitStack
# =====================================================

def test_exitstack_basic():
    log = []

    with contextlib.ExitStack() as stack:
        stack.callback(lambda: log.append("callback1"))
        stack.callback(lambda: log.append("callback2"))

    # Callbacks should be called in LIFO order
    expect(log).to_be(["callback2", "callback1"])

test("ExitStack callbacks in LIFO order", test_exitstack_basic)

def test_exitstack_enter_context():
    log = []

    class CM:
        def __init__(self, name):
            self.name = name
        def __enter__(self):
            log.append(self.name + "_enter")
            return self.name
        def __exit__(self, *args):
            log.append(self.name + "_exit")
            return False

    with contextlib.ExitStack() as stack:
        a = stack.enter_context(CM("a"))
        b = stack.enter_context(CM("b"))
        expect(a).to_be("a")
        expect(b).to_be("b")

    expect(log).to_be(["a_enter", "b_enter", "b_exit", "a_exit"])

test("ExitStack enter_context", test_exitstack_enter_context)

def test_exitstack_callback_with_args():
    result = []

    with contextlib.ExitStack() as stack:
        stack.callback(result.append, "hello")

    expect(result).to_be(["hello"])

test("ExitStack callback with args", test_exitstack_callback_with_args)

def test_exitstack_close():
    log = []
    stack = contextlib.ExitStack()
    stack.__enter__()
    stack.callback(lambda: log.append("closed"))
    stack.close()
    expect(log).to_be(["closed"])

test("ExitStack close()", test_exitstack_close)

def test_exitstack_pop_all():
    log = []
    with contextlib.ExitStack() as stack:
        stack.callback(lambda: log.append("cb"))
        new_stack = stack.pop_all()

    # Original stack should be empty, so no callbacks called yet
    expect(log).to_be([])

    # Now close the new stack
    new_stack.close()
    expect(log).to_be(["cb"])

test("ExitStack pop_all", test_exitstack_pop_all)

def test_exitstack_suppress():
    class SuppressCM:
        def __enter__(self):
            return self
        def __exit__(self, *args):
            return True  # Suppress all exceptions

    with contextlib.ExitStack() as stack:
        stack.enter_context(SuppressCM())
        raise ValueError("should be suppressed")

    expect(True).to_be(True)

test("ExitStack with suppressing context manager", test_exitstack_suppress)

# =====================================================
# AbstractContextManager
# =====================================================

def test_abstract_cm():
    # AbstractContextManager provides default __enter__ returning self
    cm = contextlib.AbstractContextManager
    expect(cm.__name__).to_be("AbstractContextManager")

test("AbstractContextManager exists", test_abstract_cm)

# =====================================================
# Practical patterns
# =====================================================

def test_temporary_value():
    @contextlib.contextmanager
    def temporary_value(lst, val):
        lst.append(val)
        try:
            yield
        finally:
            lst.remove(val)

    items = [1, 2, 3]
    with temporary_value(items, 99):
        expect(99 in items).to_be(True)
        expect(len(items)).to_be(4)

    expect(99 in items).to_be(False)
    expect(len(items)).to_be(3)

test("practical: temporary value pattern", test_temporary_value)

def test_timer_pattern():
    @contextlib.contextmanager
    def timer():
        result = {"elapsed": 0}
        yield result
        result["elapsed"] = 42  # Simulated

    with timer() as t:
        pass
    expect(t["elapsed"]).to_be(42)

test("practical: timer/result pattern", test_timer_pattern)

def test_multiple_suppress():
    errors = []
    for exc_type in [ValueError, TypeError, KeyError]:
        try:
            with contextlib.suppress(exc_type):
                if exc_type == ValueError:
                    raise ValueError("v")
                elif exc_type == TypeError:
                    raise TypeError("t")
                else:
                    raise KeyError("k")
        except:
            errors.append(exc_type)

    expect(len(errors)).to_be(0)

test("practical: suppress in loop", test_multiple_suppress)

def test_exitstack_dynamic():
    log = []

    @contextlib.contextmanager
    def make_cm(name):
        log.append(name + "_enter")
        yield name
        log.append(name + "_exit")

    names = ["a", "b", "c"]
    with contextlib.ExitStack() as stack:
        values = [stack.enter_context(make_cm(n)) for n in names]
        expect(values).to_be(["a", "b", "c"])

    expect(log).to_be(["a_enter", "b_enter", "c_enter", "c_exit", "b_exit", "a_exit"])

test("practical: ExitStack with dynamic context managers", test_exitstack_dynamic)
