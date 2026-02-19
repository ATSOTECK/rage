# Ported from CPython test_scope.py
# Additional scoping and closure tests not already covered by
# 16_cpython_closures.py, 17_closure_scoping.py, 18_nonlocal.py,
# 19_cpython_scopes.py, and 20_class_in_function_scope.py.

from test_framework import test, expect

# ---------------------------------------------------------------------------
# testSimpleAndRebinding (exact CPython version)
# The closure captures the *final* value of x after rebinding in the
# enclosing scope, not the value at the time the inner function was defined.
# ---------------------------------------------------------------------------
def test_simple_and_rebinding():
    def make_adder3(x):
        def adder(y):
            return x + y
        x = x + 1  # rebind x after defining adder
        return adder

    inc = make_adder3(0)
    plus10 = make_adder3(9)
    expect(inc(1)).to_be(2)
    expect(plus10(-2)).to_be(8)

test("simple and rebinding", test_simple_and_rebinding)

# ---------------------------------------------------------------------------
# testNestingGlobalNoFree
# Plain old globals accessed through deeply nested non-use scopes.
# ---------------------------------------------------------------------------
_global_x_nesting = 1

def test_nesting_global_no_free():
    def make_adder4():
        def nest():
            def nest():
                def adder(y):
                    return _global_x_nesting + y
                return adder
            return nest()
        return nest()

    adder = make_adder4()
    expect(adder(1)).to_be(2)

test("nesting global no free", test_nesting_global_no_free)

def test_nesting_global_no_free_updated():
    global _global_x_nesting
    _global_x_nesting = 10

    def make_adder4():
        def nest():
            def nest():
                def adder(y):
                    return _global_x_nesting + y
                return adder
            return nest()
        return nest()

    adder = make_adder4()
    expect(adder(-2)).to_be(8)
    _global_x_nesting = 1  # restore

test("nesting global no free (updated)", test_nesting_global_no_free_updated)

# ---------------------------------------------------------------------------
# testNestingPlusFreeRefToGlobal
# A free variable that refers to a global set in the enclosing function.
# ---------------------------------------------------------------------------
_global_nest_x = 0

def test_nesting_plus_free_ref_to_global():
    global _global_nest_x

    def make_adder6(x):
        global _global_nest_x
        def adder(y):
            return _global_nest_x + y
        _global_nest_x = x
        return adder

    inc = make_adder6(1)
    plus10 = make_adder6(10)
    # There is only one global — both closures see the latest value (10).
    expect(inc(1)).to_be(11)
    expect(plus10(-2)).to_be(8)

test("nesting plus free ref to global", test_nesting_plus_free_ref_to_global)

# ---------------------------------------------------------------------------
# testMixedFreevarsAndCellvars (exact CPython version)
# Variables that are both free in an inner scope and cells in an outer scope,
# with rebinding of y inside g().
# ---------------------------------------------------------------------------
def test_mixed_freevars_and_cellvars():
    def identity(x):
        return x

    def f(x, y, z):
        def g(a, b, c):
            a = a + x  # a becomes 3
            def h():
                # z * (b + y)  where y has been rebound to c + z = 9
                # = 3 * (4 + 9) = 39
                return identity(z * (b + y))
            y = c + z  # rebind y to 9
            return h
        return g

    g = f(1, 2, 3)
    h = g(2, 4, 6)
    expect(h()).to_be(39)

test("mixed freevars and cellvars", test_mixed_freevars_and_cellvars)

# ---------------------------------------------------------------------------
# testCellIsKwonlyArg
# Keyword-only parameter used as a cell variable in a closure.
# ---------------------------------------------------------------------------
def test_cell_is_kwonly_arg():
    def foo(*, a=17):
        def bar():
            return a + 5
        return bar() + 3

    expect(foo(a=42)).to_be(50)
    expect(foo()).to_be(25)

test("cell is keyword-only arg", test_cell_is_kwonly_arg)

# ---------------------------------------------------------------------------
# testComplexDefinitions
# *args and **kwargs captured in closures.
# ---------------------------------------------------------------------------
def test_complex_definitions_varargs():
    def makeReturner(*lst):
        def returner():
            return lst
        return returner

    expect(makeReturner(1, 2, 3)()).to_be((1, 2, 3))

test("complex definitions varargs", test_complex_definitions_varargs)

def test_complex_definitions_kwargs():
    def makeReturner2(**kwargs):
        def returner():
            return kwargs
        return returner

    result = makeReturner2(a=11)()
    expect(result["a"]).to_be(11)

test("complex definitions kwargs", test_complex_definitions_kwargs)

# ---------------------------------------------------------------------------
# testScopeOfGlobalStmt (Cases I through IV)
# Examples from Samuele Pedroni — global statements in nested functions.
# ---------------------------------------------------------------------------
_scope_global_x = 7

def test_scope_of_global_stmt_case_I():
    global _scope_global_x
    _scope_global_x = 7

    def f():
        x = 1
        def g():
            global _scope_global_x
            def i():
                def h():
                    return _scope_global_x
                return h()
            return i()
        return g()

    expect(f()).to_be(7)
    expect(_scope_global_x).to_be(7)

test("scope of global stmt case I", test_scope_of_global_stmt_case_I)

def test_scope_of_global_stmt_case_II():
    global _scope_global_x
    _scope_global_x = 7

    def f():
        x = 1
        def g():
            x = 2
            def i():
                def h():
                    return x
                return h()
            return i()
        return g()

    expect(f()).to_be(2)
    expect(_scope_global_x).to_be(7)

test("scope of global stmt case II", test_scope_of_global_stmt_case_II)

def test_scope_of_global_stmt_case_III():
    global _scope_global_x
    _scope_global_x = 7

    def f():
        x = 1
        def g():
            global _scope_global_x
            _scope_global_x = 2
            def i():
                def h():
                    return _scope_global_x
                return h()
            return i()
        return g()

    expect(f()).to_be(2)
    expect(_scope_global_x).to_be(2)
    _scope_global_x = 7  # restore

test("scope of global stmt case III", test_scope_of_global_stmt_case_III)

def test_scope_of_global_stmt_case_IV():
    global _scope_global_x
    _scope_global_x = 7

    def f():
        x = 3
        def g():
            global _scope_global_x
            _scope_global_x = 2
            def i():
                def h():
                    return _scope_global_x
                return h()
            return i()
        return g()

    expect(f()).to_be(2)
    expect(_scope_global_x).to_be(2)
    _scope_global_x = 7  # restore

test("scope of global stmt case IV", test_scope_of_global_stmt_case_IV)

# ---------------------------------------------------------------------------
# testUnboundLocal — outer function variant
# Accessing a variable before it is assigned in the enclosing scope.
# ---------------------------------------------------------------------------
def test_unbound_local_in_outer():
    def errorInOuter():
        # y is used before assignment in this function
        result = y
        def inner():
            return y
        y = 1
        return result

    caught = False
    try:
        errorInOuter()
    except UnboundLocalError:
        caught = True
    expect(caught).to_be(True)

test("unbound local in outer", test_unbound_local_in_outer)

# ---------------------------------------------------------------------------
# testUnboundLocal — inner function variant
# Inner function references a variable that is assigned later in the
# enclosing scope.  Calling inner() before the assignment raises NameError.
# ---------------------------------------------------------------------------
def test_unbound_local_in_inner():
    def errorInInner():
        def inner():
            return y
        inner()
        y = 1

    caught = False
    try:
        errorInInner()
    except NameError:
        caught = True
    expect(caught).to_be(True)

test("unbound local in inner", test_unbound_local_in_inner)

# ---------------------------------------------------------------------------
# testUnboundLocal_AfterDel — outer
# Deleting a cell variable and then accessing it.
# ---------------------------------------------------------------------------
def test_unbound_local_after_del_outer():
    def errorInOuter():
        y = 1
        del y
        result = y
        def inner():
            return y
        return result

    caught = False
    try:
        errorInOuter()
    except UnboundLocalError:
        caught = True
    expect(caught).to_be(True)

test("unbound local after del in outer", test_unbound_local_after_del_outer)

# ---------------------------------------------------------------------------
# testUnboundLocal_AfterDel — inner
# ---------------------------------------------------------------------------
def test_unbound_local_after_del_inner():
    def errorInInner():
        def inner():
            return y
        y = 1
        del y
        inner()

    caught = False
    try:
        errorInInner()
    except NameError:
        caught = True
    expect(caught).to_be(True)

test("unbound local after del in inner", test_unbound_local_after_del_inner)

# ---------------------------------------------------------------------------
# testNestedNonLocal
# nonlocal declarations at multiple nesting levels.
# ---------------------------------------------------------------------------
def test_nested_nonlocal():
    def f(x):
        def g():
            nonlocal x
            x -= 2
            def h():
                nonlocal x
                x += 4
                return x
            return h
        return g

    g = f(1)
    h = g()   # g() runs: x becomes 1-2 = -1
    expect(h()).to_be(3)  # h() runs: x becomes -1+4 = 3

test("nested nonlocal", test_nested_nonlocal)

# ---------------------------------------------------------------------------
# testGlobalInParallelNestedFunctions
# A global statement in one nested function must not leak to a sibling.
# ---------------------------------------------------------------------------
_parallel_y = 9

def test_global_in_parallel_nested():
    global _parallel_y
    _parallel_y = 9

    def f():
        y = 1
        def g():
            global _parallel_y
            return _parallel_y
        def h():
            return y + 1  # should see local y=1, not global
        return g, h

    g, h = f()
    expect(g()).to_be(9)
    expect(h()).to_be(2)

test("global in parallel nested functions", test_global_in_parallel_nested)

# ---------------------------------------------------------------------------
# testLambdas — additional lambda closure patterns from CPython
# ---------------------------------------------------------------------------
def test_lambda_nested_closure():
    # lambda returning lambda — closure through non-use scope
    f2 = lambda x: (lambda: lambda y: x + y)()
    inc = f2(1)
    plus10 = f2(10)
    expect(inc(1)).to_be(2)
    expect(plus10(5)).to_be(15)

test("lambda nested closure", test_lambda_nested_closure)

def test_lambda_with_global():
    global _lambda_global_x
    _lambda_global_x = 1

    f3 = lambda x: lambda y: _lambda_global_x + y
    inc = f3(None)
    expect(inc(2)).to_be(3)

test("lambda with global", test_lambda_with_global)

def test_lambda_deeply_nested():
    # Three levels of lambda with mixed free/cell vars
    f8 = lambda x, y, z: lambda a, b, c: lambda: z * (b + y)
    g = f8(1, 2, 3)
    h = g(2, 4, 6)
    expect(h()).to_be(18)

test("lambda deeply nested", test_lambda_deeply_nested)

# ---------------------------------------------------------------------------
# testFreeVarInMethod — module-level class variant
# When the class is at module level (not nested in a function), the method
# still sees the enclosing module variable, not the method with the same name.
# ---------------------------------------------------------------------------
_method_and_var = "var"

def test_free_var_in_method_module_level():
    class Test:
        def method_and_var(self):
            return "method"
        def test_it(self):
            return _method_and_var
        def actual_global(self):
            return str("global")
    t = Test()
    expect(t.test_it()).to_be("var")
    expect(t.method_and_var()).to_be("method")
    expect(t.actual_global()).to_be("global")

test("free var in method at module level", test_free_var_in_method_module_level)

# ---------------------------------------------------------------------------
# testFreeVarInMethod — nested function variant (full CPython version)
# Inside a function: method_and_var is both a free variable and a method name.
# The method body sees the enclosing variable, not the method.
# ---------------------------------------------------------------------------
def test_free_var_in_method_nested():
    def make():
        method_and_var = "var"
        class Test:
            def method_and_var(self):
                return "method"
            def test_it(self):
                return method_and_var
            def actual_global(self):
                return str("global")
            def str_method(self):
                return str(self)
        return Test()

    t = make()
    expect(t.test_it()).to_be("var")
    expect(t.method_and_var()).to_be("method")
    expect(t.actual_global()).to_be("global")

test("free var in method nested", test_free_var_in_method_nested)

# ---------------------------------------------------------------------------
# Closure over loop variable — classic late-binding trap
# Without default argument trick, all closures see the final loop value.
# ---------------------------------------------------------------------------
def test_closure_late_binding_trap():
    funcs = []
    for i in range(5):
        def f():
            return i
        funcs.append(f)
    # All closures see the final value of i
    results = [f() for f in funcs]
    expect(results).to_be([4, 4, 4, 4, 4])

test("closure late binding trap", test_closure_late_binding_trap)

# ---------------------------------------------------------------------------
# Closure with default arg vs. closure — default captures immediately.
# ---------------------------------------------------------------------------
def test_closure_default_arg_capture():
    funcs = []
    for i in range(5):
        def f(x=i):
            return x
        funcs.append(f)
    results = [f() for f in funcs]
    expect(results).to_be([0, 1, 2, 3, 4])

test("closure default arg capture", test_closure_default_arg_capture)

# ---------------------------------------------------------------------------
# testListCompLocalVars
# List comprehension variables are local to the comprehension.
# ---------------------------------------------------------------------------
def test_listcomp_local_vars():
    # 'bad' is not defined outside the comprehension
    def x():
        result = [bad for s in ["a b"] for bad in s.split()]
        return result

    expect(x()).to_be(["a", "b"])

test("listcomp local vars", test_listcomp_local_vars)

# ---------------------------------------------------------------------------
# Comprehension does not leak iteration variable into enclosing scope.
# ---------------------------------------------------------------------------
def test_comprehension_no_leak():
    x = 42
    result = [x for x in range(5)]
    # In Python 3, the comprehension variable does not leak
    expect(x).to_be(42)
    expect(result).to_be([0, 1, 2, 3, 4])

test("comprehension no variable leak", test_comprehension_no_leak)

# ---------------------------------------------------------------------------
# nonlocal with inc and dec — shared cell from two closures
# (Full CPython testNonLocalFunction)
# ---------------------------------------------------------------------------
def test_nonlocal_inc_dec():
    def f(x):
        def inc():
            nonlocal x
            x += 1
            return x
        def dec():
            nonlocal x
            x -= 1
            return x
        return inc, dec

    inc, dec = f(0)
    expect(inc()).to_be(1)
    expect(inc()).to_be(2)
    expect(dec()).to_be(1)
    expect(dec()).to_be(0)

test("nonlocal inc dec", test_nonlocal_inc_dec)

# ---------------------------------------------------------------------------
# nonlocal in class body — the nonlocal x modifies enclosing scope,
# and x should NOT appear in the class __dict__.
# ---------------------------------------------------------------------------
def test_nonlocal_class_not_in_dict():
    def f(x):
        class c:
            nonlocal x
            x += 1
            def get(self):
                return x
        return c

    C = f(0)
    obj = C()
    expect(obj.get()).to_be(1)
    # x should not be a class attribute
    has_x = "x" in C.__dict__
    expect(has_x).to_be(False)

test("nonlocal class not in dict", test_nonlocal_class_not_in_dict)

# ---------------------------------------------------------------------------
# Scope of global statement in class body
# global x inside a class means set() creates a local, get() reads global.
# ---------------------------------------------------------------------------
_class_global_x = 12

def test_class_and_global():
    global _class_global_x
    _class_global_x = 12

    class Global:
        global _class_global_x
        _class_global_x = 13
        def set_val(self, val):
            x = val  # local variable, does not affect global
        def get_val(self):
            return _class_global_x

    g = Global()
    expect(g.get_val()).to_be(13)
    g.set_val(15)
    expect(g.get_val()).to_be(13)

test("class and global", test_class_and_global)

# ---------------------------------------------------------------------------
# Deeply nested closures — four levels deep
# ---------------------------------------------------------------------------
def test_four_level_closure():
    def level1(a):
        def level2(b):
            def level3(c):
                def level4(d):
                    return a + b + c + d
                return level4
            return level3
        return level2

    f = level1(1)(2)(3)
    expect(f(4)).to_be(10)

test("four level closure", test_four_level_closure)

# ---------------------------------------------------------------------------
# Closure over boolean and None
# ---------------------------------------------------------------------------
def test_closure_over_none_and_bool():
    def make():
        n = None
        t = True
        f = False
        def get_none():
            return n
        def get_true():
            return t
        def get_false():
            return f
        return get_none, get_true, get_false

    gn, gt, gf = make()
    expect(gn()).to_be(None)
    expect(gt()).to_be(True)
    expect(gf()).to_be(False)

test("closure over None and bool", test_closure_over_none_and_bool)

# ---------------------------------------------------------------------------
# Multiple closures returning from a single scope — each sees same cell
# ---------------------------------------------------------------------------
def test_multiple_closures_same_cell():
    def f():
        x = 0
        def a():
            nonlocal x
            x += 1
            return x
        def b():
            nonlocal x
            x += 10
            return x
        def c():
            return x
        return a, b, c

    a, b, c = f()
    expect(a()).to_be(1)
    expect(b()).to_be(11)
    expect(a()).to_be(12)
    expect(c()).to_be(12)

test("multiple closures same cell", test_multiple_closures_same_cell)

# ---------------------------------------------------------------------------
# Cell variable with keyword-only arg and default
# ---------------------------------------------------------------------------
def test_cell_kwonly_with_default():
    def outer(*, value=10):
        def inner():
            return value * 2
        return inner

    expect(outer()()).to_be(20)
    expect(outer(value=7)()).to_be(14)

test("cell kwonly with default", test_cell_kwonly_with_default)

# ---------------------------------------------------------------------------
# Cell variable that is a regular argument escaping into closure
# ---------------------------------------------------------------------------
def test_cell_is_arg_and_escapes():
    def spam(arg):
        def eggs():
            return arg
        return eggs

    eggs = spam(42)
    expect(eggs()).to_be(42)

    # The closure captures the value, a second call gets a fresh cell
    eggs2 = spam(99)
    expect(eggs2()).to_be(99)
    expect(eggs()).to_be(42)  # original still holds 42

test("cell is arg and escapes", test_cell_is_arg_and_escapes)

# ---------------------------------------------------------------------------
# Cell variable that is a local (not arg) escaping into closure
# ---------------------------------------------------------------------------
def test_cell_is_local_and_escapes():
    def spam(arg):
        cell = arg
        def eggs():
            return cell
        return eggs

    eggs = spam(42)
    expect(eggs()).to_be(42)

test("cell is local and escapes", test_cell_is_local_and_escapes)

# ---------------------------------------------------------------------------
# Recursive closure inside a function
# ---------------------------------------------------------------------------
def test_recursive_closure():
    def f(x):
        def fact(n):
            if n == 0:
                return 1
            else:
                return n * fact(n - 1)
        if x >= 0:
            return fact(x)
        else:
            raise ValueError("x must be >= 0")

    expect(f(6)).to_be(720)

test("recursive closure", test_recursive_closure)

def test_recursive_closure_negative():
    def f(x):
        def fact(n):
            if n == 0:
                return 1
            else:
                return n * fact(n - 1)
        if x >= 0:
            return fact(x)
        else:
            raise ValueError("x must be >= 0")

    caught = False
    try:
        f(-1)
    except ValueError:
        caught = True
    expect(caught).to_be(True)

test("recursive closure negative raises", test_recursive_closure_negative)

# ---------------------------------------------------------------------------
# Generator with nonlocal — yields values that depend on nonlocal mutation
# ---------------------------------------------------------------------------
def test_generator_nonlocal_accumulate():
    def f(start):
        def gen(n):
            nonlocal start
            for i in range(n):
                start += i
                yield start
        return gen

    g = f(100)
    expect(list(g(5))).to_be([100, 101, 103, 106, 110])

test("generator nonlocal accumulate", test_generator_nonlocal_accumulate)

# ---------------------------------------------------------------------------
# Nested classes — inner class method accesses outer function variable
# ---------------------------------------------------------------------------
def test_nested_class_closure():
    def outer(x):
        class A:
            class B:
                def get(self):
                    return x
            def get(self):
                return x + 1
        return A(), A.B()

    a, b = outer(10)
    expect(a.get()).to_be(11)
    expect(b.get()).to_be(10)

test("nested class closure", test_nested_class_closure)

# ---------------------------------------------------------------------------
# Closure interacting with *args parameter
# ---------------------------------------------------------------------------
def test_closure_with_starargs():
    def outer(*args):
        def inner():
            return args
        return inner

    f = outer(1, 2, 3)
    expect(f()).to_be((1, 2, 3))

test("closure with starargs", test_closure_with_starargs)

# ---------------------------------------------------------------------------
# Closure interacting with **kwargs parameter
# ---------------------------------------------------------------------------
def test_closure_with_starkwargs():
    def outer(**kwargs):
        def inner():
            return kwargs
        return inner

    f = outer(x=1, y=2)
    result = f()
    expect(result["x"]).to_be(1)
    expect(result["y"]).to_be(2)

test("closure with starkwargs", test_closure_with_starkwargs)

# ---------------------------------------------------------------------------
# Closure with mixed positional, keyword-only, *args, **kwargs
# ---------------------------------------------------------------------------
def test_closure_mixed_params():
    def outer(a, b, *args, key=99, **kwargs):
        def inner():
            return (a, b, args, key, kwargs)
        return inner

    f = outer(1, 2, 3, 4, key=5, extra=6)
    a, b, args, key, kwargs = f()
    expect(a).to_be(1)
    expect(b).to_be(2)
    expect(args).to_be((3, 4))
    expect(key).to_be(5)
    expect(kwargs["extra"]).to_be(6)

test("closure mixed params", test_closure_mixed_params)

# ---------------------------------------------------------------------------
# UnboundLocalError from augmented assignment without nonlocal
# ---------------------------------------------------------------------------
def test_unbound_local_aug_assign():
    x = 1
    def f():
        x += 1  # x is local due to augmented assignment, but never assigned first
    caught = False
    try:
        f()
    except UnboundLocalError:
        caught = True
    expect(caught).to_be(True)

test("unbound local aug assign", test_unbound_local_aug_assign)

# ---------------------------------------------------------------------------
# Global variable unaffected by local assignment in nested function
# ---------------------------------------------------------------------------
_outer_val = "global"

def test_global_unaffected_by_local():
    def f():
        _outer_val = "local"
        return _outer_val
    expect(f()).to_be("local")
    expect(_outer_val).to_be("global")

test("global unaffected by local", test_global_unaffected_by_local)

# ---------------------------------------------------------------------------
# Closure over class instance — method returns closure
# ---------------------------------------------------------------------------
def test_method_returns_closure():
    class Counter:
        def __init__(self):
            self.count = 0
        def make_incrementer(self):
            def inc():
                self.count += 1
                return self.count
            return inc

    c = Counter()
    inc = c.make_incrementer()
    expect(inc()).to_be(1)
    expect(inc()).to_be(2)
    expect(c.count).to_be(2)

test("method returns closure", test_method_returns_closure)

# ---------------------------------------------------------------------------
# Closure in list comprehension with enclosing variable
# ---------------------------------------------------------------------------
def test_closure_in_listcomp():
    def make(n):
        return [lambda x, i=i: x + i + n for i in range(3)]

    funcs = make(100)
    expect(funcs[0](0)).to_be(100)
    expect(funcs[1](0)).to_be(101)
    expect(funcs[2](0)).to_be(102)

test("closure in listcomp", test_closure_in_listcomp)

# ---------------------------------------------------------------------------
# Multiple nonlocal declarations in deeply nested chain
# ---------------------------------------------------------------------------
def test_deep_nonlocal_chain():
    def a():
        x = 0
        def b():
            nonlocal x
            x += 1
            def c():
                nonlocal x
                x += 10
                def d():
                    nonlocal x
                    x += 100
                    return x
                return d
            return c
        return b

    b = a()
    c = b()   # x = 1
    d = c()   # x = 11
    expect(d()).to_be(111)

test("deep nonlocal chain", test_deep_nonlocal_chain)

# ---------------------------------------------------------------------------
# Closure captures variable from for-else
# ---------------------------------------------------------------------------
def test_closure_for_else():
    def f():
        for i in range(5):
            pass
        else:
            found = True
        def inner():
            return (i, found)
        return inner()

    expect(f()).to_be((4, True))

test("closure for else", test_closure_for_else)

# ---------------------------------------------------------------------------
# Closure captures variable from try/except
# ---------------------------------------------------------------------------
def test_closure_try_except():
    def f():
        try:
            x = 1
            y = 1 / 0
        except ZeroDivisionError:
            x = -1
        def inner():
            return x
        return inner()

    expect(f()).to_be(-1)

test("closure try except", test_closure_try_except)

# ---------------------------------------------------------------------------
# Closure captures variable from with statement
# ---------------------------------------------------------------------------
def test_closure_with_stmt():
    class DummyCM:
        def __enter__(self):
            return 42
        def __exit__(self, *args):
            return False

    def f():
        with DummyCM() as val:
            pass
        def inner():
            return val
        return inner()

    expect(f()).to_be(42)

test("closure with statement", test_closure_with_stmt)

# ---------------------------------------------------------------------------
# Global statement isolation — global in one function does not affect another
# ---------------------------------------------------------------------------
_isolation_var = "original"

def test_global_isolation():
    global _isolation_var
    _isolation_var = "original"

    def writer():
        global _isolation_var
        _isolation_var = "written"

    def reader():
        # No global declaration — if the variable were somehow leaked as
        # global to this sibling, it would read the global.
        # But since there is no assignment here, it reads the global anyway.
        return _isolation_var

    writer()
    expect(reader()).to_be("written")
    _isolation_var = "original"  # restore

test("global isolation", test_global_isolation)

# ---------------------------------------------------------------------------
# Nonlocal with conditional assignment
# ---------------------------------------------------------------------------
def test_nonlocal_conditional():
    def f(flag):
        x = 0
        def set_x():
            nonlocal x
            if flag:
                x = 1
            else:
                x = 2
        set_x()
        return x

    expect(f(True)).to_be(1)
    expect(f(False)).to_be(2)

test("nonlocal conditional", test_nonlocal_conditional)

# ---------------------------------------------------------------------------
# Closure with tuple unpacking in for loop
# ---------------------------------------------------------------------------
def test_closure_tuple_unpack_loop():
    def f():
        pairs = [(1, "a"), (2, "b"), (3, "c")]
        results = []
        for k, v in pairs:
            def g(key=k, val=v):
                return (key, val)
            results.append(g())
        return results

    expect(f()).to_be([(1, "a"), (2, "b"), (3, "c")])

test("closure tuple unpack loop", test_closure_tuple_unpack_loop)

# ---------------------------------------------------------------------------
# Class body sees enclosing function's variable, method also sees it
# ---------------------------------------------------------------------------
def test_class_body_and_method_see_enclosing():
    def f(x):
        class C:
            a = x       # class body sees x
            def m(self):
                return x  # method also sees x
        return C

    C = f(77)
    expect(C.a).to_be(77)
    expect(C().m()).to_be(77)

test("class body and method see enclosing", test_class_body_and_method_see_enclosing)

# ---------------------------------------------------------------------------
# Class body assignment does NOT modify enclosing variable
# ---------------------------------------------------------------------------
def test_class_body_does_not_modify_enclosing():
    def f():
        x = 10
        class C:
            x = 99  # this creates a class attribute, not modifying enclosing x
            def m(self):
                return x  # reads enclosing x (10), not C.x (99)
        return x, C.x, C().m()

    outer_x, class_x, method_x = f()
    expect(outer_x).to_be(10)
    expect(class_x).to_be(99)
    expect(method_x).to_be(10)

test("class body does not modify enclosing", test_class_body_does_not_modify_enclosing)

print("CPython scoping tests completed")
