# Test: Exceptions
# Tests try/except/finally, raise, exception types, inheritance

from test_framework import test, expect

# Helper functions at module level
def raise_in_func():
    raise ValueError

def inner_raise():
    raise KeyError

def outer_call():
    inner_raise()

gen_finally_ran = False
def gen_with_finally():
    global gen_finally_ran
    try:
        yield 1
        yield 2
    finally:
        gen_finally_ran = True

def gen_catches_throw():
    try:
        yield 1
        yield 2
    except ValueError:
        yield "caught"
    yield 3

def gen_no_catch():
    yield 1
    yield 2
    yield 3

def gen_internal_except():
    try:
        yield 1
        raise ValueError
    except ValueError:
        yield "internal_caught"
    yield 3

def simple_gen():
    yield 1

def test_basic_try_except():
    result = "not caught"
    try:
        raise ValueError
    except ValueError:
        result = "caught"
    expect(result).to_be("caught")

def test_exception_as_binding():
    caught = False
    try:
        raise ValueError
    except ValueError as e:
        caught = True
    expect(caught).to_be(True)

def test_multiple_except():
    # KeyError case
    result = ""
    try:
        raise KeyError
    except ValueError:
        result = "value"
    except KeyError:
        result = "key"
    except TypeError:
        result = "type"
    expect(result).to_be("key")

    # TypeError case
    result2 = ""
    try:
        raise TypeError
    except ValueError:
        result2 = "value"
    except KeyError:
        result2 = "key"
    except TypeError:
        result2 = "type"
    expect(result2).to_be("type")

def test_bare_except():
    result = ""
    try:
        raise RuntimeError
    except:
        result = "caught"
    expect(result).to_be("caught")

def test_finally_no_exception():
    finally_ran = False
    try:
        x = 1
    finally:
        finally_ran = True
    expect(finally_ran).to_be(True)

def test_finally_with_caught_exception():
    finally_ran = False
    caught = False
    try:
        raise ValueError
    except ValueError:
        caught = True
    finally:
        finally_ran = True
    expect(finally_ran).to_be(True)
    expect(caught).to_be(True)

def test_finally_propagates():
    outer_finally_ran = False
    inner_finally_ran = False
    outer_caught = False
    try:
        try:
            raise ValueError
        finally:
            inner_finally_ran = True
    except ValueError:
        outer_caught = True
    finally:
        outer_finally_ran = True
    expect(inner_finally_ran).to_be(True)
    expect(outer_finally_ran).to_be(True)
    expect(outer_caught).to_be(True)

def test_else_clause():
    # Runs when no exception
    else_ran = False
    try:
        x = 1
    except ValueError:
        pass
    else:
        else_ran = True
    expect(else_ran).to_be(True)

    # Doesn't run when exception
    else_ran2 = False
    try:
        raise ValueError
    except ValueError:
        pass
    else:
        else_ran2 = True
    expect(else_ran2).to_be(False)

def test_reraise():
    caught_outer = False
    try:
        try:
            raise ValueError
        except ValueError:
            raise
    except ValueError:
        caught_outer = True
    expect(caught_outer).to_be(True)

def test_exception_inheritance():
    # ValueError is Exception
    caught1 = False
    try:
        raise ValueError
    except Exception:
        caught1 = True
    expect(caught1).to_be(True)

    # KeyError is Exception
    caught2 = False
    try:
        raise KeyError
    except Exception:
        caught2 = True
    expect(caught2).to_be(True)

    # IndexError is Exception
    caught3 = False
    try:
        raise IndexError
    except Exception:
        caught3 = True
    expect(caught3).to_be(True)

    # ZeroDivisionError is Exception
    caught4 = False
    try:
        raise ZeroDivisionError
    except Exception:
        caught4 = True
    expect(caught4).to_be(True)

def test_nested_try():
    inner_caught = False
    outer_caught = False
    try:
        try:
            raise KeyError
        except ValueError:
            inner_caught = True
    except KeyError:
        outer_caught = True
    expect(inner_caught).to_be(False)
    expect(outer_caught).to_be(True)

def test_full_try_except_else_finally():
    # No exception case
    caught = False
    else_ran = False
    finally_ran = False
    try:
        x = 1
    except ValueError:
        caught = True
    else:
        else_ran = True
    finally:
        finally_ran = True
    expect(caught).to_be(False)
    expect(else_ran).to_be(True)
    expect(finally_ran).to_be(True)

    # With exception case
    caught2 = False
    else_ran2 = False
    finally_ran2 = False
    try:
        raise ValueError
    except ValueError:
        caught2 = True
    else:
        else_ran2 = True
    finally:
        finally_ran2 = True
    expect(caught2).to_be(True)
    expect(else_ran2).to_be(False)
    expect(finally_ran2).to_be(True)

def test_tuple_except():
    # Catching ValueError
    result1 = ""
    try:
        raise ValueError
    except (ValueError, KeyError):
        result1 = "caught"
    expect(result1).to_be("caught")

    # Catching KeyError
    result2 = ""
    try:
        raise KeyError
    except (ValueError, KeyError):
        result2 = "caught"
    expect(result2).to_be("caught")

def test_exception_classes():
    count = 0

    try:
        raise ValueError
    except ValueError:
        count = count + 1

    try:
        raise TypeError
    except TypeError:
        count = count + 1

    try:
        raise KeyError
    except KeyError:
        count = count + 1

    try:
        raise RuntimeError
    except RuntimeError:
        count = count + 1

    try:
        raise StopIteration
    except StopIteration:
        count = count + 1

    expect(count).to_be(5)

def test_deeply_nested():
    depth_reached = 0
    try:
        depth_reached = 1
        try:
            depth_reached = 2
            try:
                depth_reached = 3
                raise ValueError
            except TypeError:
                depth_reached = -1
        except KeyError:
            depth_reached = -2
    except ValueError:
        pass
    expect(depth_reached).to_be(3)

def test_var_preserved():
    x = 10
    try:
        x = 20
        raise ValueError
    except ValueError:
        pass
    expect(x).to_be(20)

def test_finally_after_handler_raises():
    finally_after_reraise = False
    outer_finally = False
    try:
        try:
            raise ValueError
        except ValueError:
            raise KeyError
        finally:
            finally_after_reraise = True
    except KeyError:
        pass
    finally:
        outer_finally = True
    expect(finally_after_reraise).to_be(True)
    expect(outer_finally).to_be(True)

def test_func_exception():
    func_exc_caught = False
    try:
        raise_in_func()
    except ValueError:
        func_exc_caught = True
    expect(func_exc_caught).to_be(True)

def test_nested_func_exception():
    nested_func_exc_caught = False
    try:
        outer_call()
    except KeyError:
        nested_func_exc_caught = True
    expect(nested_func_exc_caught).to_be(True)

def test_generator_throw_caught():
    g = gen_catches_throw()
    gen_throw_results = []
    for v in g:
        gen_throw_results.append(v)
        if v == 1:
            gen_throw_results.append(g.throw(ValueError, "test"))
            break
    expect(gen_throw_results[0]).to_be(1)
    expect(gen_throw_results[1]).to_be("caught")

def test_generator_throw_propagates():
    g2 = gen_no_catch()
    gen_throw_propagated = False
    for v in g2:
        if v == 1:
            try:
                g2.throw(RuntimeError, "uncaught")
            except RuntimeError:
                gen_throw_propagated = True
            break
    expect(gen_throw_propagated).to_be(True)

def test_generator_close_finally():
    global gen_finally_ran
    gen_finally_ran = False
    g3 = gen_with_finally()
    for v in g3:
        break
    g3.close()
    expect(gen_finally_ran).to_be(True)

def test_generator_internal_except():
    gen_internal_results = []
    for v in gen_internal_except():
        gen_internal_results.append(v)
    expect(gen_internal_results[0]).to_be(1)
    expect(gen_internal_results[1]).to_be("internal_caught")
    expect(gen_internal_results[2]).to_be(3)

def test_throw_into_closed_gen():
    g4 = simple_gen()
    for v in g4:
        pass  # Exhaust generator

    throw_into_closed_raised = False
    try:
        g4.throw(ValueError, "to closed")
    except:
        throw_into_closed_raised = True
    expect(throw_into_closed_raised).to_be(True)

test("basic_try_except", test_basic_try_except)
test("exception_as_binding", test_exception_as_binding)
test("multiple_except", test_multiple_except)
test("bare_except", test_bare_except)
test("finally_no_exception", test_finally_no_exception)
test("finally_with_caught_exception", test_finally_with_caught_exception)
test("finally_propagates", test_finally_propagates)
test("else_clause", test_else_clause)
test("reraise", test_reraise)
test("exception_inheritance", test_exception_inheritance)
test("nested_try", test_nested_try)
test("full_try_except_else_finally", test_full_try_except_else_finally)
test("tuple_except", test_tuple_except)
test("exception_classes", test_exception_classes)
test("deeply_nested", test_deeply_nested)
test("var_preserved", test_var_preserved)
test("finally_after_handler_raises", test_finally_after_handler_raises)
test("func_exception", test_func_exception)
test("nested_func_exception", test_nested_func_exception)
test("generator_throw_caught", test_generator_throw_caught)
test("generator_throw_propagates", test_generator_throw_propagates)
test("generator_close_finally", test_generator_close_finally)
test("generator_internal_except", test_generator_internal_except)
test("throw_into_closed_gen", test_throw_into_closed_gen)

# --- UnboundLocalError from negation of unassigned variable ---

def test_negate_unbound_raises():
    caught = False
    try:
        def f():
            x = -x
            return x
        f()
    except UnboundLocalError:
        caught = True
    expect(caught).to_be(True)

def test_negate_normal():
    x = 42
    x = -x
    expect(x).to_be(-42)

def test_negate_float():
    x = 3.14
    x = -x
    expect(x < -3.13).to_be(True)
    expect(x > -3.15).to_be(True)

test("negate_unbound_raises", test_negate_unbound_raises)
test("negate_normal", test_negate_normal)
test("negate_float", test_negate_float)


# === return inside try/finally must run the finally body ===
# Regression: OpReturn in the main dispatch used to drop the block stack
# and pop the frame without routing through BlockFinally, so the finally
# clause was silently skipped on a `return` from inside the `try`.
def test_return_runs_finally():
    log = []
    def f():
        try:
            return 1
        finally:
            log.append("f")
    expect(f()).to_be(1)
    expect(log).to_be(["f"])


def test_return_runs_nested_finally_in_order():
    log = []
    def f():
        try:
            try:
                return "inner"
            finally:
                log.append("f1")
        finally:
            log.append("f2")
    expect(f()).to_be("inner")
    expect(log).to_be(["f1", "f2"])


def test_finally_return_overrides_try_return():
    def f():
        try:
            return 1
        finally:
            return 2
    expect(f()).to_be(2)


def test_finally_exception_replaces_return():
    def f():
        try:
            return 1
        finally:
            raise ValueError("boom")
    caught = None
    try:
        f()
    except ValueError as e:
        caught = str(e)
    expect(caught).to_be("boom")


def test_return_through_with_inside_try_finally():
    log = []

    class CM:
        def __enter__(self):
            return self
        def __exit__(self, *args):
            log.append("exit")
            return False

    def f():
        try:
            with CM():
                return "done"
        finally:
            log.append("f")

    expect(f()).to_be("done")
    # __exit__ first (innermost), then finally
    expect(log).to_be(["exit", "f"])


test("return_runs_finally", test_return_runs_finally)
test("return_runs_nested_finally_in_order", test_return_runs_nested_finally_in_order)
test("finally_return_overrides_try_return", test_finally_return_overrides_try_return)
test("finally_exception_replaces_return", test_finally_exception_replaces_return)
test("return_through_with_inside_try_finally", test_return_through_with_inside_try_finally)


# === continue/break through try/finally must run the finally body ===
# Regression: main dispatch's OpContinueLoop and `break` (OpJump) used to
# skip the finally body and let BlockFinally entries accumulate across
# iterations.
def test_continue_runs_finally():
    log = []
    def f():
        for i in range(3):
            try:
                if i == 1:
                    continue
                log.append(i)
            finally:
                log.append("f" + str(i))
    f()
    expect(log).to_be([0, "f0", "f1", 2, "f2"])


def test_break_runs_finally():
    log = []
    def f():
        for i in range(3):
            try:
                if i == 1:
                    break
                log.append(i)
            finally:
                log.append("f" + str(i))
    f()
    expect(log).to_be([0, "f0", "f1"])


# === continue/break through `with` must call __exit__ ===
def test_continue_runs_with_exit():
    log = []

    class CM:
        def __enter__(self):
            return self
        def __exit__(self, *a):
            log.append("exit")
            return False

    def f():
        for i in range(2):
            with CM():
                if i == 0:
                    continue
                log.append("body")
    f()
    expect(log).to_be(["exit", "body", "exit"])


def test_break_runs_with_exit():
    log = []

    class CM:
        def __enter__(self):
            return self
        def __exit__(self, *a):
            log.append("exit")
            return False

    def f():
        for i in range(3):
            with CM():
                log.append(i)
                break
    f()
    expect(log).to_be([0, "exit"])


# === continue inside try/finally inside with: finally then __exit__ ===
def test_continue_through_finally_and_with():
    log = []

    class CM:
        def __enter__(self):
            return self
        def __exit__(self, *a):
            log.append("exit")
            return False

    def f():
        for i in range(2):
            with CM():
                try:
                    if i == 0:
                        continue
                    log.append("body")
                finally:
                    log.append("f")
    f()
    expect(log).to_be(["f", "exit", "body", "f", "exit"])


# === break out of nested with blocks runs every __exit__ (LIFO) ===
# Regression: break eagerly popped the for-iterator, which actually removed an
# enclosing with's context manager and left the cleanup unwinder reading a
# dead stack slot. Now the iterator is popped at a landing pad after cleanup.
def test_break_runs_nested_with_exits():
    log = []

    class CM:
        def __init__(self, name):
            self.name = name
        def __enter__(self):
            return self
        def __exit__(self, *a):
            log.append("exit " + self.name)
            return False

    def f():
        for i in range(5):
            with CM("a"):
                with CM("b"):
                    log.append("body")
                    break
    f()
    expect(log).to_be(["body", "exit b", "exit a"])


# === break out of a for-loop skips the else clause ===
def test_break_skips_for_else():
    log = []

    class CM:
        def __enter__(self):
            return self
        def __exit__(self, *a):
            log.append("exit")
            return False

    def f():
        for i in range(3):
            with CM():
                break
        else:
            log.append("else")
    f()
    expect(log).to_be(["exit"])


# === unbounded recursion raises a catchable RecursionError ===
# Regression: with no recursion limit this overflowed the Go stack and aborted
# the whole process with an uncatchable fatal error. Now it raises a normal,
# catchable RecursionError.
def test_recursion_error_catchable():
    def recurse(n):
        return recurse(n + 1)

    caught = False
    try:
        recurse(0)
    except RecursionError:
        caught = True
    expect(caught).to_be(True)


test("continue_runs_finally", test_continue_runs_finally)
test("break_runs_finally", test_break_runs_finally)
test("continue_runs_with_exit", test_continue_runs_with_exit)
test("break_runs_with_exit", test_break_runs_with_exit)
test("continue_through_finally_and_with", test_continue_through_finally_and_with)
test("break_runs_nested_with_exits", test_break_runs_nested_with_exits)
test("break_skips_for_else", test_break_skips_for_else)
test("recursion_error_catchable", test_recursion_error_catchable)

print("Exceptions tests completed")
