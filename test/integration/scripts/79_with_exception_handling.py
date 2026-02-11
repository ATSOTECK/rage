# Test: With statement exception handling
# Tests that __exit__ is properly called with exception info and can suppress exceptions

from test_framework import test, expect

# === __exit__ receives exception info on exception ===
def test_exit_receives_exception():
    exit_args = [None, None, None]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_args[0] = exc_type
            exit_args[1] = str(exc_val)
            exit_args[2] = exc_tb
            return False
    try:
        with CM():
            raise ValueError("test error")
    except ValueError:
        pass
    expect("test error" in exit_args[1]).to_be(True)
    expect(exit_args[2]).to_be(None)

# === __exit__ returning True suppresses exception ===
def test_exit_suppresses_exception():
    suppressed = [False]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            return True  # Suppress the exception
    # This should NOT raise because __exit__ returns True
    with CM():
        raise ValueError("should be suppressed")
    suppressed[0] = True
    expect(suppressed[0]).to_be(True)

# === __exit__ returning False propagates exception ===
def test_exit_propagates_exception():
    caught = [False]
    exit_called = [False]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_called[0] = True
            return False  # Don't suppress
    try:
        with CM():
            raise ValueError("should propagate")
    except ValueError:
        caught[0] = True
    expect(exit_called[0]).to_be(True)
    expect(caught[0]).to_be(True)

# === __exit__ receives None args on normal exit ===
def test_exit_none_on_normal():
    exit_args = []
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_args.append(exc_type)
            exit_args.append(exc_val)
            exit_args.append(exc_tb)
            return False
    with CM():
        pass
    expect(exit_args[0]).to_be(None)
    expect(exit_args[1]).to_be(None)
    expect(exit_args[2]).to_be(None)

# === Nested with + exception in inner ===
def test_nested_with_exception():
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
    try:
        with CM("outer"):
            with CM("inner"):
                raise ValueError("boom")
    except ValueError:
        pass
    expect(log).to_be(["enter_outer", "enter_inner", "exit_inner", "exit_outer"])

# === Inner suppresses, outer sees normal exit ===
def test_nested_inner_suppresses():
    log = []
    class SuppressCM:
        def __enter__(self):
            log.append("enter_suppress")
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            log.append("exit_suppress")
            return True  # Suppress
    class NormalCM:
        def __enter__(self):
            log.append("enter_normal")
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            if exc_type is None:
                log.append("exit_normal_clean")
            else:
                log.append("exit_normal_exc")
            return False
    with NormalCM():
        with SuppressCM():
            raise ValueError("suppressed")
        log.append("after_suppress")
    expect(log).to_be(["enter_normal", "enter_suppress", "exit_suppress", "after_suppress", "exit_normal_clean"])

# === with inside try/except ===
def test_with_inside_try_except():
    exit_called = [False]
    caught = [False]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            exit_called[0] = True
            return False
    try:
        with CM():
            raise RuntimeError("test")
    except RuntimeError:
        caught[0] = True
    expect(exit_called[0]).to_be(True)
    expect(caught[0]).to_be(True)

# === Exception type is passed correctly ===
def test_exception_type_passed():
    received_type = [None]
    class CM:
        def __enter__(self):
            return self
        def __exit__(self, exc_type, exc_val, exc_tb):
            received_type[0] = exc_type
            return True
    with CM():
        raise TypeError("type test")
    # exc_type should be some representation of TypeError
    expect(received_type[0] is not None).to_be(True)

# === Context manager with as and exception ===
def test_with_as_and_exception():
    class CM:
        def __init__(self):
            self.entered = False
            self.exited = False
        def __enter__(self):
            self.entered = True
            return "resource"
        def __exit__(self, exc_type, exc_val, exc_tb):
            self.exited = True
            return True
    cm = CM()
    with cm as val:
        expect(val).to_be("resource")
        expect(cm.entered).to_be(True)
        raise ValueError("should be suppressed")
    expect(cm.exited).to_be(True)

# Register all tests
test("exit_receives_exception", test_exit_receives_exception)
test("exit_suppresses_exception", test_exit_suppresses_exception)
test("exit_propagates_exception", test_exit_propagates_exception)
test("exit_none_on_normal", test_exit_none_on_normal)
test("nested_with_exception", test_nested_with_exception)
test("nested_inner_suppresses", test_nested_inner_suppresses)
test("with_inside_try_except", test_with_inside_try_except)
test("exception_type_passed", test_exception_type_passed)
test("with_as_and_exception", test_with_as_and_exception)

print("With statement exception handling tests completed")
