# Ported from CPython test_generators.py
# Additional generator tests beyond 23_cpython_generators.py and 24_cpython_generators_advanced.py

from test_framework import test, expect

# =============================================================================
# Tutorial tests: basic generator protocol
# =============================================================================

def test_simple_generator_for_loop():
    """Simple generator iterated with for loop."""
    def f():
        yield 1
        yield 2
    result = []
    for i in f():
        result.append(i)
    expect(result).to_be([1, 2])

test("simple_generator_for_loop", test_simple_generator_for_loop)

def test_generator_next_stops():
    """Falling off the end raises StopIteration."""
    def f():
        yield 1
        yield 2
    g = f()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

test("generator_next_stops", test_generator_next_stops)

def test_return_stops_generator():
    """Return statement stops the generator, unreached yield is skipped."""
    def f():
        yield 1
        return
        yield 2  # never reached
    g = f()
    expect(next(g)).to_be(1)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

test("return_stops_generator", test_return_stops_generator)

def test_stopped_generator_stays_stopped():
    """Once stopped, a generator can't be resumed."""
    def f():
        yield 1
        return
    g = f()
    expect(next(g)).to_be(1)
    caught1 = False
    try:
        next(g)
    except StopIteration:
        caught1 = True
    expect(caught1).to_be(True)
    caught2 = False
    try:
        next(g)
    except StopIteration:
        caught2 = True
    expect(caught2).to_be(True)

test("stopped_generator_stays_stopped", test_stopped_generator_stays_stopped)

# =============================================================================
# Return vs StopIteration in try/except
# =============================================================================

def test_return_in_try_except():
    """Return in try/except exits without triggering except."""
    def g1():
        try:
            return
        except:
            yield 1
    expect(list(g1())).to_be([])

test("return_in_try_except", test_return_in_try_except)

def test_stopiteration_caught_by_except():
    """Raising StopIteration is caught by bare except (unlike return)."""
    def g2():
        try:
            raise StopIteration
        except:
            yield 42
    expect(list(g2())).to_be([42])

test("stopiteration_caught_by_except", test_stopiteration_caught_by_except)

def test_return_in_try_finally():
    """Return in try/finally still executes finally (which can yield)."""
    def g3():
        try:
            return
        finally:
            yield 1
    expect(list(g3())).to_be([1])

test("return_in_try_finally", test_return_in_try_finally)

# =============================================================================
# Alternate range() as generator
# =============================================================================

def test_yrange():
    """Generator-based range implementation."""
    def yrange(n):
        for i in range(n):
            yield i
    expect(list(yrange(5))).to_be([0, 1, 2, 3, 4])

test("yrange", test_yrange)

# =============================================================================
# Generators calling other generators
# =============================================================================

def test_generator_calls_generator():
    """Generator that iterates over another generator."""
    def yrange(n):
        for i in range(n):
            yield i
    def zrange(n):
        for i in yrange(n):
            yield i
    expect(list(zrange(5))).to_be([0, 1, 2, 3, 4])

test("generator_calls_generator", test_generator_calls_generator)

# =============================================================================
# Generator returns to most recent caller
# =============================================================================

def test_generator_returns_to_caller():
    """Generators return to the most recent caller, not the creator."""
    def yrange(n):
        for i in range(n):
            yield i

    log = []
    def creator():
        r = yrange(5)
        log.append("creator " + str(next(r)))
        return r

    def caller():
        r = creator()
        for i in r:
            log.append("caller " + str(i))

    caller()
    expect(log).to_be(["creator 0", "caller 1", "caller 2", "caller 3", "caller 4"])

test("generator_returns_to_caller", test_generator_returns_to_caller)

# =============================================================================
# Exception propagation through generators (PEP tests)
# =============================================================================

def test_exception_propagation():
    """Exceptions propagate out of generators and stop them."""
    def f():
        return 1 // 0
    def g():
        yield f()
        yield 42  # never reached
    k = g()
    caught = False
    try:
        next(k)
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)
    # Generator is now dead
    stopped = False
    try:
        next(k)
    except StopIteration:
        stopped = True
    expect(stopped).to_be(True)

test("exception_propagation", test_exception_propagation)

# =============================================================================
# Try/Except/Finally in generators
# =============================================================================

def test_try_except_finally_in_generator():
    """Complex try/except/finally with re-raise in a generator."""
    def f():
        try:
            yield 1
            try:
                yield 2
                1 // 0
                yield 3  # never get here
            except ZeroDivisionError:
                yield 4
                yield 5
                raise
            except:
                yield 6
            yield 7  # the "raise" above stops this
        except:
            yield 8
        yield 9
        try:
            x = 12
        finally:
            yield 10
        yield 11
    expect(list(f())).to_be([1, 2, 4, 5, 8, 9, 10, 11])

test("try_except_finally_in_generator", test_try_except_finally_in_generator)

# =============================================================================
# Yielding None
# =============================================================================

def test_yielding_none():
    """Yielding None explicitly and returning."""
    def g():
        for i in range(3):
            yield None
        yield None
        return
    expect(list(g())).to_be([None, None, None, None])

test("yielding_none", test_yielding_none)

# =============================================================================
# StopIteration in try/except yields
# =============================================================================

def test_stopiteration_in_try_except():
    """Explicitly raising StopIteration acts like any other exception in try/except."""
    def g():
        yield 1
        try:
            raise StopIteration
        except:
            yield 2
        yield 3
    expect(list(g())).to_be([1, 2, 3])

test("stopiteration_in_try_except", test_stopiteration_in_try_except)

# =============================================================================
# Recursive generator: combinations
# =============================================================================

def test_recursive_combinations():
    """Recursive generator for combinations."""
    def gcomb(x, k):
        if k > len(x):
            return
        if k == 0:
            yield []
        else:
            first = x[0]
            rest = x[1:]
            for c in gcomb(rest, k - 1):
                c.insert(0, first)
                yield c
            for c in gcomb(rest, k):
                yield c

    seq = list(range(1, 5))
    result_0 = list(gcomb(seq, 0))
    expect(result_0).to_be([[]])
    result_1 = list(gcomb(seq, 1))
    expect(result_1).to_be([[1], [2], [3], [4]])
    result_2 = list(gcomb(seq, 2))
    expect(result_2).to_be([[1, 2], [1, 3], [1, 4], [2, 3], [2, 4], [3, 4]])
    result_5 = list(gcomb(seq, 5))
    expect(result_5).to_be([])

test("recursive_combinations", test_recursive_combinations)

# =============================================================================
# Generator type checks
# =============================================================================

def test_iter_of_generator_is_self():
    """iter(g) is g for generators."""
    def g():
        yield 1
    i = g()
    expect(iter(i) is i).to_be(True)

test("iter_of_generator_is_self", test_iter_of_generator_is_self)

# =============================================================================
# Yield by itself yields None
# =============================================================================

def test_bare_yield():
    """Yield by itself yields None."""
    def f():
        yield
    expect(list(f())).to_be([None])

test("bare_yield", test_bare_yield)

# =============================================================================
# Unreachable yield still makes it a generator
# =============================================================================

def test_unreachable_yield_still_generator():
    """A function with unreachable yield is still a generator."""
    def f():
        if 0:
            yield
    g = f()
    result = list(g)
    expect(result).to_be([])

test("unreachable_yield_still_generator", test_unreachable_yield_still_generator)

def test_unreachable_yield_value_still_generator():
    """A function with unreachable yield value is still a generator."""
    def f():
        if 0:
            yield 1
    g = f()
    result = list(g)
    expect(result).to_be([])

test("unreachable_yield_value_still_generator", test_unreachable_yield_value_still_generator)

# =============================================================================
# Continue in try/finally with yield
# =============================================================================

def test_continue_in_try_finally_with_yield():
    """Yield in finally block of try/finally with continue."""
    def f():
        for i in range(3):
            try:
                continue
            finally:
                yield i
    g = f()
    expect(next(g)).to_be(0)
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

test("continue_in_try_finally_with_yield", test_continue_in_try_finally_with_yield)

# =============================================================================
# Lambda generator
# =============================================================================

def test_lambda_generator():
    """Lambda with yield produces a generator."""
    x = lambda: (yield 1)
    expect(list(x())).to_be([1])

test("lambda_generator", test_lambda_generator)

def test_lambda_multiple_yields():
    """Lambda with multiple yields."""
    x = lambda: ((yield 1), (yield 2))
    expect(list(x())).to_be([1, 2])

test("lambda_multiple_yields", test_lambda_multiple_yields)

# =============================================================================
# Send values into running generator (coroutine pattern)
# =============================================================================

def test_send_value_into_generator():
    """Sending a value into a started generator."""
    def f():
        result = yield 1
        yield result
    g = f()
    expect(next(g)).to_be(1)
    expect(g.send(42)).to_be(42)

test("send_value_into_generator", test_send_value_into_generator)

def test_send_non_none_to_new_raises():
    """Sending a non-None value to a just-started generator raises TypeError."""
    def f():
        yield 1
    g = f()
    caught = False
    try:
        g.send("foo")
    except TypeError:
        caught = True
    expect(caught).to_be(True)

test("send_non_none_to_new_raises", test_send_non_none_to_new_raises)

# =============================================================================
# Yield augmented assignment (coroutine pattern)
# =============================================================================

def test_yield_augmented_assignment():
    """Yield with augmented assignment in coroutine pattern."""
    def coroutine(seq):
        count = 0
        while count < 200:
            count += yield
            seq.append(count)
    seq = []
    c = coroutine(seq)
    next(c)
    expect(seq).to_be([])
    c.send(10)
    expect(seq).to_be([10])
    c.send(10)
    expect(seq).to_be([10, 20])
    c.send(10)
    expect(seq).to_be([10, 20, 30])

test("yield_augmented_assignment", test_yield_augmented_assignment)

# =============================================================================
# Throw exception into generator
# =============================================================================

def test_throw_valueerror():
    """Throw ValueError into a generator that catches it."""
    def f():
        while True:
            try:
                val = yield
                # val is unused in catch path
            except ValueError:
                yield "caught ValueError"
    g = f()
    next(g)
    result = g.throw(ValueError)
    expect(result).to_be("caught ValueError")

test("throw_valueerror", test_throw_valueerror)

def test_throw_valueerror_with_message():
    """Throw ValueError with a message into generator."""
    def f():
        while True:
            try:
                val = yield
            except ValueError as v:
                yield "caught: " + str(v)
    g = f()
    next(g)
    result = g.throw(ValueError("xyz"))
    expect(result).to_be("caught: xyz")

test("throw_valueerror_with_message", test_throw_valueerror_with_message)

# =============================================================================
# Close generator with GeneratorExit
# =============================================================================

def test_close_with_print_in_except():
    """Close triggers GeneratorExit which can be caught."""
    log = []
    def f():
        try:
            yield
        except GeneratorExit:
            log.append("exiting")
    g = f()
    next(g)
    g.close()
    expect(log).to_be(["exiting"])
    # Closing again is a no-op
    g.close()
    expect(log).to_be(["exiting"])

test("close_with_except_generatorexit", test_close_with_print_in_except)

def test_close_fresh_generator():
    """Closing a never-started generator is fine."""
    def f():
        yield
    f().close()
    # No exception raised
    expect(True).to_be(True)

test("close_fresh_generator", test_close_fresh_generator)

def test_close_exhausted_generator():
    """Closing an already-exhausted generator is fine."""
    def f():
        yield
    g = f()
    next(g)
    caught = False
    try:
        next(g)
    except StopIteration:
        caught = True
    expect(caught).to_be(True)
    g.close()
    expect(True).to_be(True)

test("close_exhausted_generator", test_close_exhausted_generator)

def test_close_with_finally():
    """Close triggers finally clause."""
    log = []
    def f():
        try:
            yield
        finally:
            log.append("exiting")
    g = f()
    next(g)
    g.close()
    expect(log).to_be(["exiting"])

test("close_with_finally", test_close_with_finally)

def test_close_with_except_and_finally():
    """Close triggers finally, not except Exception."""
    log = []
    def f():
        try:
            yield
        except Exception:
            log.append("except")
        finally:
            log.append("finally")
    g = f()
    next(g)
    g.close()
    expect(log).to_be(["finally"])

test("close_with_except_and_finally", test_close_with_except_and_finally)

# =============================================================================
# Sieve of Eratosthenes as recursive generator
# =============================================================================

def test_sieve_of_eratosthenes():
    """Build up to a recursive Sieve of Eratosthenes generator."""
    def firstn(g, n):
        return [next(g) for i in range(n)]

    def intsfrom(i):
        while True:
            yield i
            i += 1

    expect(firstn(intsfrom(5), 7)).to_be([5, 6, 7, 8, 9, 10, 11])

    def exclude_multiples(n, ints):
        for i in ints:
            if i % n:
                yield i

    expect(firstn(exclude_multiples(3, intsfrom(1)), 6)).to_be([1, 2, 4, 5, 7, 8])

    def sieve(ints):
        prime = next(ints)
        yield prime
        not_divisible_by_prime = exclude_multiples(prime, ints)
        for p in sieve(not_divisible_by_prime):
            yield p

    primes = sieve(intsfrom(2))
    expect(firstn(primes, 20)).to_be([2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71])

test("sieve_of_eratosthenes", test_sieve_of_eratosthenes)

# =============================================================================
# Hamming numbers (2^i * 3^j * 5^k) using merge
# =============================================================================

def test_merge_generators():
    """Merge two sorted infinite generators."""
    def firstn(g, n):
        return [next(g) for i in range(n)]

    def intsfrom(i):
        while True:
            yield i
            i += 1

    def times(n, g):
        for i in g:
            yield n * i

    expect(firstn(times(10, intsfrom(1)), 10)).to_be([10, 20, 30, 40, 50, 60, 70, 80, 90, 100])

    def merge(g, h):
        ng = next(g)
        nh = next(h)
        while True:
            if ng < nh:
                yield ng
                ng = next(g)
            elif ng > nh:
                yield nh
                nh = next(h)
            else:
                yield ng
                ng = next(g)
                nh = next(h)

    def m235():
        yield 1
        me_times2 = times(2, m235())
        me_times3 = times(3, m235())
        me_times5 = times(5, m235())
        for i in merge(merge(me_times2, me_times3), me_times5):
            yield i

    result = m235()
    expect(firstn(result, 15)).to_be([1, 2, 3, 4, 5, 6, 8, 9, 10, 12, 15, 16, 18, 20, 24])
    expect(firstn(result, 15)).to_be([25, 27, 30, 32, 36, 40, 45, 48, 50, 54, 60, 64, 72, 75, 80])

test("merge_generators", test_merge_generators)

# =============================================================================
# Binary tree traversal with generators
# =============================================================================

def test_binary_tree_inorder():
    """Recursive generator for in-order tree traversal."""
    class Tree:
        def __init__(self, label, left=None, right=None):
            self.label = label
            self.left = left
            self.right = right

    def tree_from_list(lst):
        n = len(lst)
        if n == 0:
            return []
        i = n // 2
        return Tree(lst[i], tree_from_list(lst[:i]), tree_from_list(lst[i+1:]))

    def inorder(t):
        if t:
            for x in inorder(t.left):
                yield x
            yield t.label
            for x in inorder(t.right):
                yield x

    t = tree_from_list("ABCDEFG")
    result = list(inorder(t))
    expect(result).to_be(["A", "B", "C", "D", "E", "F", "G"])

test("binary_tree_inorder", test_binary_tree_inorder)

# =============================================================================
# Non-recursive tree traversal with generator
# =============================================================================

def test_binary_tree_inorder_nonrecursive():
    """Non-recursive generator for in-order tree traversal using a stack."""
    class Tree:
        def __init__(self, label, left=None, right=None):
            self.label = label
            self.left = left
            self.right = right

    def tree_from_list(lst):
        n = len(lst)
        if n == 0:
            return []
        i = n // 2
        return Tree(lst[i], tree_from_list(lst[:i]), tree_from_list(lst[i+1:]))

    def inorder(node):
        stack = []
        while node:
            while node.left:
                stack.append(node)
                node = node.left
            yield node.label
            while not node.right:
                try:
                    node = stack.pop()
                except IndexError:
                    return
                yield node.label
            node = node.right

    t = tree_from_list("ABCDEFG")
    result = list(inorder(t))
    expect(result).to_be(["A", "B", "C", "D", "E", "F", "G"])

test("binary_tree_inorder_nonrecursive", test_binary_tree_inorder_nonrecursive)

# =============================================================================
# Yield from: basic delegation
# =============================================================================

def test_yield_from_basic_list():
    """yield from with a list iterable."""
    def gen():
        yield from [1, 2, 3]
    expect(list(gen())).to_be([1, 2, 3])

test("yield_from_basic_list", test_yield_from_basic_list)

def test_yield_from_generator():
    """yield from with a sub-generator."""
    def sub():
        yield "a"
        yield "b"
    def main():
        yield "start"
        yield from sub()
        yield "end"
    expect(list(main())).to_be(["start", "a", "b", "end"])

test("yield_from_generator", test_yield_from_generator)

def test_yield_from_range():
    """yield from with range."""
    def gen():
        yield from range(5)
    expect(list(gen())).to_be([0, 1, 2, 3, 4])

test("yield_from_range", test_yield_from_range)

def test_yield_from_string():
    """yield from with a string."""
    def gen():
        yield from "abc"
    expect(list(gen())).to_be(["a", "b", "c"])

test("yield_from_string", test_yield_from_string)

def test_yield_from_tuple():
    """yield from with a tuple."""
    def gen():
        yield from (10, 20, 30)
    expect(list(gen())).to_be([10, 20, 30])

test("yield_from_tuple", test_yield_from_tuple)

def test_yield_from_empty():
    """yield from with an empty iterable."""
    def gen():
        yield from []
    expect(list(gen())).to_be([])

test("yield_from_empty", test_yield_from_empty)

# =============================================================================
# Yield from: chaining multiple
# =============================================================================

def test_yield_from_chained():
    """Multiple yield from in sequence."""
    def gen():
        yield from [1, 2]
        yield from [3, 4]
        yield from [5]
    expect(list(gen())).to_be([1, 2, 3, 4, 5])

test("yield_from_chained", test_yield_from_chained)

# =============================================================================
# Yield from: return value propagation
# =============================================================================

def test_yield_from_return_value():
    """yield from captures the return value of the sub-generator."""
    def sub():
        yield 1
        yield 2
        return "result"
    def main():
        val = yield from sub()
        yield val
    expect(list(main())).to_be([1, 2, "result"])

test("yield_from_return_value", test_yield_from_return_value)

# =============================================================================
# Yield from: nested 3 levels
# =============================================================================

def test_yield_from_three_levels():
    """Three levels of yield from delegation."""
    def level3():
        yield "c"
        return "from3"
    def level2():
        result = yield from level3()
        yield result
        return "from2"
    def level1():
        result = yield from level2()
        yield result
    expect(list(level1())).to_be(["c", "from3", "from2"])

test("yield_from_three_levels", test_yield_from_three_levels)

# =============================================================================
# Yield from: send to sub-generator
# =============================================================================

def test_yield_from_send():
    """send() passes through yield from to sub-generator."""
    def sub():
        x = yield "ready"
        yield "got " + str(x)
    def main():
        yield from sub()
    g = main()
    expect(next(g)).to_be("ready")
    expect(g.send(42)).to_be("got 42")

test("yield_from_send", test_yield_from_send)

# =============================================================================
# Nested generators (not yield from, manual iteration)
# =============================================================================

def test_nested_generator_manual():
    """Manually iterating nested generators."""
    def outer():
        def inner(n):
            for i in range(n):
                yield i
        for val in inner(3):
            yield val * 10
        for val in inner(2):
            yield val * 100
    expect(list(outer())).to_be([0, 10, 20, 0, 100])

test("nested_generator_manual", test_nested_generator_manual)

# =============================================================================
# Generator with complex control flow
# =============================================================================

def test_generator_complex_control_flow():
    """Generator with if/elif/else and loops."""
    def classify(items):
        for item in items:
            if item > 0:
                yield "pos"
            elif item < 0:
                yield "neg"
            else:
                yield "zero"
    expect(list(classify([3, -1, 0, 5, -2]))).to_be(["pos", "neg", "zero", "pos", "neg"])

test("generator_complex_control_flow", test_generator_complex_control_flow)

# =============================================================================
# Generator with break in consumer
# =============================================================================

def test_generator_with_break():
    """Breaking out of a for loop over a generator."""
    def infinite():
        i = 0
        while True:
            yield i
            i += 1
    result = []
    for val in infinite():
        if val >= 5:
            break
        result.append(val)
    expect(result).to_be([0, 1, 2, 3, 4])

test("generator_with_break", test_generator_with_break)

# =============================================================================
# Generator with default argument capture
# =============================================================================

def test_generator_default_arg_capture():
    """Generator factory using default argument to capture loop variable."""
    generators = []
    for i in range(3):
        def gen(n=i):
            yield n
        generators.append(gen)
    result = [next(g()) for g in generators]
    expect(result).to_be([0, 1, 2])

test("generator_default_arg_capture", test_generator_default_arg_capture)

# =============================================================================
# Generator with closures
# =============================================================================

def test_generator_with_closure():
    """Generator accessing variables from enclosing scope."""
    def make_gen(multiplier):
        def gen():
            for i in range(5):
                yield i * multiplier
        return gen
    g2 = make_gen(2)
    g3 = make_gen(3)
    expect(list(g2())).to_be([0, 2, 4, 6, 8])
    expect(list(g3())).to_be([0, 3, 6, 9, 12])

test("generator_with_closure", test_generator_with_closure)

# =============================================================================
# Generator try/except with ZeroDivisionError
# =============================================================================

def test_generator_yields_in_try_except():
    """Try/except inside for loop in generator."""
    def safe_div(items, divisor):
        for item in items:
            try:
                yield item // divisor
            except ZeroDivisionError:
                yield -1
    expect(list(safe_div([10, 20, 30], 5))).to_be([2, 4, 6])

test("generator_yields_in_try_except", test_generator_yields_in_try_except)

# =============================================================================
# Generator with nested try/finally
# =============================================================================

def test_generator_nested_try_finally():
    """Nested try/finally in generator."""
    log = []
    def gen():
        try:
            try:
                yield 1
            finally:
                log.append("inner finally")
            yield 2
        finally:
            log.append("outer finally")
        yield 3
    result = list(gen())
    expect(result).to_be([1, 2, 3])
    expect(log).to_be(["inner finally", "outer finally"])

test("generator_nested_try_finally", test_generator_nested_try_finally)

# =============================================================================
# Generator with multiple except clauses
# =============================================================================

def test_generator_multiple_except():
    """Generator with multiple except clauses."""
    def gen():
        try:
            yield 1
            raise ValueError("test")
        except TypeError:
            yield "type error"
        except ValueError:
            yield "value error"
        yield "done"
    expect(list(gen())).to_be([1, "value error", "done"])

test("generator_multiple_except", test_generator_multiple_except)

# =============================================================================
# Queens solver (simplified N=4 test)
# =============================================================================

def test_nqueens_generator():
    """N-Queens solver using generators (N=4)."""
    def queens(n):
        def safe(queen, queens):
            for i in range(len(queens)):
                q = queens[i]
                if q == queen:
                    return False
                diff = len(queens) - i
                if q == queen + diff or q == queen - diff:
                    return False
            return True

        def solve(n, queens_so_far):
            if len(queens_so_far) == n:
                yield list(queens_so_far)
                return
            for col in range(n):
                if safe(col, queens_so_far):
                    queens_so_far.append(col)
                    for solution in solve(n, queens_so_far):
                        yield solution
                    queens_so_far.pop()

        for solution in solve(n, []):
            yield solution

    solutions = list(queens(4))
    expect(len(solutions)).to_be(2)
    expect(solutions[0]).to_be([1, 3, 0, 2])
    expect(solutions[1]).to_be([2, 0, 3, 1])

test("nqueens_generator", test_nqueens_generator)

# =============================================================================
# Generator with map/filter patterns from CPython tests
# =============================================================================

def test_firstn_helper():
    """firstn helper pattern used in CPython tests."""
    def firstn(g, n):
        return [next(g) for i in range(n)]

    def intsfrom(i):
        while True:
            yield i
            i += 1

    expect(firstn(intsfrom(5), 7)).to_be([5, 6, 7, 8, 9, 10, 11])

test("firstn_helper", test_firstn_helper)

def test_exclude_multiples():
    """Exclude multiples pattern from CPython sieve test."""
    def intsfrom(i):
        while True:
            yield i
            i += 1

    def exclude_multiples(n, ints):
        for i in ints:
            if i % n:
                yield i

    def firstn(g, n):
        return [next(g) for i in range(n)]

    expect(firstn(exclude_multiples(3, intsfrom(1)), 6)).to_be([1, 2, 4, 5, 7, 8])

test("exclude_multiples", test_exclude_multiples)

# =============================================================================
# Generator with multiple independent instances
# =============================================================================

def test_multiple_generator_instances():
    """Multiple independent instances of the same generator function."""
    def counter(start, step):
        n = start
        while True:
            yield n
            n += step

    g1 = counter(0, 1)
    g2 = counter(10, 5)
    g3 = counter(100, -10)

    r1 = [next(g1) for _ in range(5)]
    r2 = [next(g2) for _ in range(5)]
    r3 = [next(g3) for _ in range(5)]

    expect(r1).to_be([0, 1, 2, 3, 4])
    expect(r2).to_be([10, 15, 20, 25, 30])
    expect(r3).to_be([100, 90, 80, 70, 60])

test("multiple_generator_instances", test_multiple_generator_instances)

# =============================================================================
# Generator with tuple unpacking in yield
# =============================================================================

def test_generator_tuple_unpacking():
    """Generator yielding tuples, consumed with unpacking."""
    def pairs():
        yield (1, "one")
        yield (2, "two")
        yield (3, "three")

    nums = []
    names = []
    for n, name in pairs():
        nums.append(n)
        names.append(name)

    expect(nums).to_be([1, 2, 3])
    expect(names).to_be(["one", "two", "three"])

test("generator_tuple_unpacking", test_generator_tuple_unpacking)

# =============================================================================
# Generator expression in various contexts
# =============================================================================

def test_genexpr_in_sum():
    """Generator expression passed to sum()."""
    result = sum(x * x for x in range(10))
    expect(result).to_be(285)

test("genexpr_in_sum", test_genexpr_in_sum)

def test_genexpr_in_any():
    """Generator expression passed to any()."""
    expect(any(x > 5 for x in range(10))).to_be(True)
    expect(any(x > 100 for x in range(10))).to_be(False)

test("genexpr_in_any", test_genexpr_in_any)

def test_genexpr_in_all():
    """Generator expression passed to all()."""
    expect(all(x >= 0 for x in range(10))).to_be(True)
    expect(all(x > 5 for x in range(10))).to_be(False)

test("genexpr_in_all", test_genexpr_in_all)

def test_genexpr_in_min_max():
    """Generator expression passed to min() and max()."""
    expect(min(x * x for x in range(1, 6))).to_be(1)
    expect(max(x * x for x in range(1, 6))).to_be(25)

test("genexpr_in_min_max", test_genexpr_in_min_max)

def test_genexpr_in_list():
    """Generator expression passed to list()."""
    expect(list(x + 1 for x in range(5))).to_be([1, 2, 3, 4, 5])

test("genexpr_in_list", test_genexpr_in_list)

def test_genexpr_in_tuple():
    """Generator expression passed to tuple()."""
    expect(tuple(x * 2 for x in range(4))).to_be((0, 2, 4, 6))

test("genexpr_in_tuple", test_genexpr_in_tuple)

def test_genexpr_in_set():
    """Generator expression passed to set()."""
    result = set(x % 3 for x in range(10))
    expect(len(result)).to_be(3)
    expect(0 in result).to_be(True)
    expect(1 in result).to_be(True)
    expect(2 in result).to_be(True)

test("genexpr_in_set", test_genexpr_in_set)

def test_genexpr_in_dict():
    """Dict comprehension-like generator with dict()."""
    result = dict((x, x * x) for x in range(5))
    expect(result).to_be({0: 0, 1: 1, 2: 4, 3: 9, 4: 16})

test("genexpr_in_dict", test_genexpr_in_dict)

def test_genexpr_in_join():
    """Generator expression passed to str.join()."""
    result = ", ".join(str(x) for x in range(5))
    expect(result).to_be("0, 1, 2, 3, 4")

test("genexpr_in_join", test_genexpr_in_join)

# =============================================================================
# Generator with return value captured via StopIteration
# =============================================================================

def test_generator_return_value_via_stopiteration():
    """Generator return value is carried in StopIteration.value."""
    def gen():
        yield 1
        yield 2
        return "finished"
    g = gen()
    expect(next(g)).to_be(1)
    expect(next(g)).to_be(2)
    caught = False
    return_val = None
    try:
        next(g)
    except StopIteration as e:
        caught = True
        return_val = e.args[0] if e.args else None
    expect(caught).to_be(True)
    expect(return_val).to_be("finished")

test("generator_return_value_via_stopiteration", test_generator_return_value_via_stopiteration)

# =============================================================================
# send() return value via StopIteration
# =============================================================================

def test_send_return_tuple():
    """Return of a sent tuple is accessible via StopIteration."""
    def g():
        return (yield 1)
    gen = g()
    expect(next(gen)).to_be(1)
    caught = False
    return_val = None
    try:
        gen.send((2,))
    except StopIteration as e:
        caught = True
        return_val = e.args[0] if e.args else None
    expect(caught).to_be(True)
    expect(return_val).to_be((2,))

test("send_return_tuple", test_send_return_tuple)

# =============================================================================
# Generator with dict yield pattern
# =============================================================================

def test_generator_dict_yield():
    """Generator that yields dict entries."""
    def items():
        yield "a", 1
        yield "b", 2
        yield "c", 3
    result = dict(items())
    expect(result).to_be({"a": 1, "b": 2, "c": 3})

test("generator_dict_yield", test_generator_dict_yield)

# =============================================================================
# Generator interleaving
# =============================================================================

def test_generator_interleave():
    """Interleave two generators."""
    def interleave(g1, g2):
        it1 = iter(g1)
        it2 = iter(g2)
        while True:
            try:
                yield next(it1)
            except StopIteration:
                yield from it2
                return
            try:
                yield next(it2)
            except StopIteration:
                yield from it1
                return

    def gen1():
        yield "a"
        yield "b"
        yield "c"

    def gen2():
        yield 1
        yield 2
        yield 3

    expect(list(interleave(gen1(), gen2()))).to_be(["a", 1, "b", 2, "c", 3])

test("generator_interleave", test_generator_interleave)

# =============================================================================
# Generator flatten
# =============================================================================

def test_generator_flatten():
    """Generator that flattens nested lists."""
    def flatten(lst):
        for item in lst:
            if isinstance(item, list):
                for sub in flatten(item):
                    yield sub
            else:
                yield item

    nested = [1, [2, 3], [4, [5, 6]], 7]
    expect(list(flatten(nested))).to_be([1, 2, 3, 4, 5, 6, 7])

test("generator_flatten", test_generator_flatten)

# =============================================================================
# Generator chunking
# =============================================================================

def test_generator_chunk():
    """Generator that yields chunks of n items."""
    def chunk(iterable, n):
        it = iter(iterable)
        while True:
            batch = []
            for _ in range(n):
                try:
                    batch.append(next(it))
                except StopIteration:
                    if batch:
                        yield batch
                    return
            yield batch

    expect(list(chunk(range(10), 3))).to_be([[0, 1, 2], [3, 4, 5], [6, 7, 8], [9]])
    expect(list(chunk(range(6), 2))).to_be([[0, 1], [2, 3], [4, 5]])
    expect(list(chunk([], 3))).to_be([])

test("generator_chunk", test_generator_chunk)

# =============================================================================
# Generator sliding window
# =============================================================================

def test_generator_sliding_window():
    """Generator for sliding window over a sequence."""
    def window(seq, n):
        it = iter(seq)
        buf = []
        for _ in range(n):
            try:
                buf.append(next(it))
            except StopIteration:
                return
        yield list(buf)
        for item in it:
            buf = buf[1:]
            buf.append(item)
            yield list(buf)

    expect(list(window([1, 2, 3, 4, 5], 3))).to_be([[1, 2, 3], [2, 3, 4], [3, 4, 5]])
    expect(list(window([1, 2], 3))).to_be([])
    expect(list(window([1, 2, 3], 1))).to_be([[1], [2], [3]])

test("generator_sliding_window", test_generator_sliding_window)

# =============================================================================
# Flat conjoin pattern from CPython tests
# =============================================================================

def test_conjoin_generators():
    """Conjoin generators to produce cartesian products (simple version)."""
    def simple_conjoin(gs):
        values = [None] * len(gs)
        def gen(i):
            if i >= len(gs):
                yield list(values)
            else:
                for values[i] in gs[i]():
                    for x in gen(i + 1):
                        yield x
        for x in gen(0):
            yield x

    result = list(simple_conjoin([lambda: iter((0, 1))] * 3))
    expect(len(result)).to_be(8)
    expect(result[0]).to_be([0, 0, 0])
    expect(result[-1]).to_be([1, 1, 1])

test("conjoin_generators", test_conjoin_generators)

# =============================================================================
# Generator protocol: __iter__ and __next__
# =============================================================================

def test_generator_protocol():
    """Generator implements iterator protocol."""
    def gen():
        yield 1
        yield 2
    g = gen()
    expect(iter(g) is g).to_be(True)
    expect(g.__next__()).to_be(1)
    expect(g.__next__()).to_be(2)
    caught = False
    try:
        g.__next__()
    except StopIteration:
        caught = True
    expect(caught).to_be(True)

test("generator_protocol", test_generator_protocol)

# =============================================================================
# Close not started generator no-op
# =============================================================================

def test_close_no_return_value():
    """close() on a fresh generator returns None."""
    def f():
        yield
    gen = f()
    result = gen.close()
    expect(result).to_be(None)

test("close_no_return_value", test_close_no_return_value)

# =============================================================================
# Yield from with send forwarding
# =============================================================================

def test_yield_from_send_forwarding():
    """send() is forwarded through yield from to the sub-generator."""
    def accumulator():
        total = 0
        while True:
            val = yield total
            if val is None:
                return total
            total += val

    def wrapper():
        result = yield from accumulator()
        yield "final: " + str(result)

    g = wrapper()
    expect(next(g)).to_be(0)
    expect(g.send(10)).to_be(10)
    expect(g.send(20)).to_be(30)
    expect(g.send(5)).to_be(35)
    expect(g.send(None)).to_be("final: 35")

test("yield_from_send_forwarding", test_yield_from_send_forwarding)

# =============================================================================
# Generator with exception in yield from
# =============================================================================

def test_yield_from_exception_propagation():
    """Exceptions from sub-generator propagate through yield from."""
    def failing_gen():
        yield 1
        raise ValueError("inner error")

    def wrapper():
        yield from failing_gen()

    g = wrapper()
    expect(next(g)).to_be(1)
    caught = False
    msg = ""
    try:
        next(g)
    except ValueError as e:
        caught = True
        msg = str(e)
    expect(caught).to_be(True)
    expect(msg).to_be("inner error")

test("yield_from_exception_propagation", test_yield_from_exception_propagation)

# =============================================================================
# Generator expression with custom __iter__ (only calls __iter__ once)
# =============================================================================

def test_genexpr_calls_iter_once():
    """Generator expression only calls __iter__ once on the iterable."""
    class Iterator:
        def __init__(self):
            self.val = 0
        def __next__(self):
            if self.val == 2:
                raise StopIteration
            self.val += 1
            return self.val

    class C:
        def __iter__(self):
            return Iterator()

    expect(list(i for i in C())).to_be([1, 2])

test("genexpr_calls_iter_once", test_genexpr_calls_iter_once)

# =============================================================================
# Generator with deeply nested yield from
# =============================================================================

def test_deep_yield_from():
    """Deeply nested yield from chain."""
    def gen(depth, val):
        if depth == 0:
            yield val
        else:
            yield from gen(depth - 1, val)

    expect(list(gen(0, "x"))).to_be(["x"])
    expect(list(gen(5, "deep"))).to_be(["deep"])
    expect(list(gen(10, 42))).to_be([42])

test("deep_yield_from", test_deep_yield_from)

# =============================================================================
# Generator: exception in finally does not suppress original
# =============================================================================

def test_generator_exception_in_finally():
    """If generator exhausts, finally runs normally."""
    log = []
    def gen():
        try:
            yield 1
            yield 2
        except GeneratorExit:
            log.append("exit")
            raise
        finally:
            log.append("finally")

    g = gen()
    expect(next(g)).to_be(1)
    g.close()
    expect("exit" in log).to_be(True)
    expect("finally" in log).to_be(True)

test("generator_exception_in_finally", test_generator_exception_in_finally)

# =============================================================================
# Generator zip pattern
# =============================================================================

def test_generator_zip_pattern():
    """Using generators with zip."""
    def evens():
        n = 0
        while True:
            yield n
            n += 2

    def odds():
        n = 1
        while True:
            yield n
            n += 2

    result = []
    e = evens()
    o = odds()
    for _ in range(5):
        result.append((next(e), next(o)))
    expect(result).to_be([(0, 1), (2, 3), (4, 5), (6, 7), (8, 9)])

test("generator_zip_pattern", test_generator_zip_pattern)

# =============================================================================
# Generator: yield in nested function is a different generator
# =============================================================================

def test_yield_in_nested_function():
    """Yield in nested function creates separate generator."""
    def outer():
        def inner():
            yield "inner"
        yield "outer"
        yield from inner()
        yield "outer again"
    expect(list(outer())).to_be(["outer", "inner", "outer again"])

test("yield_in_nested_function", test_yield_in_nested_function)

# =============================================================================
# Test issue: for loop over generator that raises in body
# =============================================================================

def test_for_loop_generator_raises():
    """Exception during for loop iteration over generator from CPython."""
    def gen_raises():
        yield
        raise ValueError()

    caught = False
    def loop():
        nonlocal caught
        try:
            for _ in gen_raises():
                if True is False:
                    return
        except ValueError:
            caught = True

    loop()
    expect(caught).to_be(True)

test("for_loop_generator_raises", test_for_loop_generator_raises)

print("CPython extra generator tests completed")
