# Test: Parser, compiler, and VM regressions
# Each test covers a construct that previously parsed wrong, compiled to an
# unbalanced stack, or raised the wrong exception type.

from test_framework import test, expect

# === Parser ===

def test_elif_chain():
    def classify(x):
        if x == 1:
            return "one"
        elif x == 2:
            return "two"
        elif x == 3:
            return "three"
        else:
            return "other"
    expect([classify(i) for i in range(5)]).to_be(["other", "one", "two", "three", "other"])

def test_elif_chain_no_else():
    def classify(x):
        r = "none"
        if x < 0:
            r = "neg"
        elif x == 0:
            r = "zero"
        elif x < 10:
            r = "small"
        return r
    expect([classify(v) for v in (-1, 0, 5, 50)]).to_be(["neg", "zero", "small", "none"])

def test_single_line_suites():
    def f(): return 1
    out = []
    for i in range(4):
        if i == 2: continue
        if i == 3: break
        out.append(i)
    while True: break
    try: x = 1
    except: x = 2
    if x: a = 1; b = 2
    class K:
        def m(self): self.v = 5; return self.v
    expect((f(), out, x, a, b, K().m())).to_be((1, [0, 1], 1, 1, 2, 5))

def test_comments_in_odd_places():
    if True:  # trailing comment after colon
        v = 1  # trailing
    # comment between branches
    else:  # another
        v = 2
    items = [1,  # one
             2]  # two
    d = {
        # leading
        "a": 1,  # a
    }
    def g(p,  # first
          q):  # second
        return p + q
    expect((v, items, d, g(1, 2))).to_be((1, [1, 2], {"a": 1}, 3))

def test_semicolons():
    a = 1; b = 2; c = a + b
    expect((a, b, c)).to_be((1, 2, 3))

def test_unpacking_displays():
    a = [1, 2]
    d = {"x": 1}
    expect([*a, *a, 3]).to_be([1, 2, 1, 2, 3])
    expect((*a, 3)).to_be((1, 2, 3))
    expect({*a, 5}).to_be({1, 2, 5})
    expect({**d, "y": 2}).to_be({"x": 1, "y": 2})
    expect({**d, **{"x": 9}}).to_be({"x": 9})

def test_call_unpacking():
    def f(*args, **kw):
        return args, kw
    expect(f(1, *[2, 3], 4, z=5, **{"y": 6})).to_be(((1, 2, 3, 4), {"z": 5, "y": 6}))
    expect(f(**{"a": 1}, **{"b": 2})).to_be(((), {"a": 1, "b": 2}))
    caught = None
    try:
        f(a=1, **{"a": 2})
    except TypeError:
        caught = "TypeError"
    expect(caught).to_be("TypeError")

def test_lambda_signatures():
    f = lambda a, *b, c=3, **d: (a, b, c, d)
    expect(f(1, 2, c=4, e=5)).to_be((1, (2,), 4, {"e": 5}))
    g = lambda *, k: k
    expect(g(k=2)).to_be(2)
    h = lambda a, /, b: (a, b)
    expect(h(1, b=2)).to_be((1, 2))

def test_positional_only_params():
    def f(a, b, /, c, *, d=4):
        return a, b, c, d
    expect(f(1, 2, 3)).to_be((1, 2, 3, 4))
    expect(f(1, 2, c=3, d=5)).to_be((1, 2, 3, 5))

def test_kwargs_keep_call_order():
    def f(**kw):
        return list(kw)
    expect(f(e=1, b=2, a=3, d=4, c=5)).to_be(["e", "b", "a", "d", "c"])
    expect(list(dict(z=3, y=2, x=1))).to_be(["z", "y", "x"])

# === Compiler / peephole optimizer ===

def test_short_circuit_in_call_args():
    def g(*args):
        return args
    def f(a, b, c):
        return g(a and b, c), g(a or b, c)
    expect(f(0, 2, 3)).to_be(((0, 3), (2, 3)))
    expect(f(1, 2, 3)).to_be(((2, 3), (1, 3)))

def test_typed_except_inside_loop():
    total = 0
    for i in range(5):
        try:
            if i % 2:
                raise ValueError(i)
            total += i
        except ValueError:
            total += 100
    expect(total).to_be(206)

def test_match_in_loop():
    out = []
    for s in [0, 5, "abc", [1, 2], {"k": 1}, 2.5, None]:
        match s:
            case 0:
                out.append("zero")
            case int(n) if n > 3:
                out.append("big")
            case str() as t if len(t) > 2:
                out.append("lstr")
            case [a, *rest]:
                out.append("seq")
            case {"k": v, **others}:
                out.append("k")
            case float() | int():
                out.append("num")
            case _:
                out.append("other")
    expect(out).to_be(["zero", "big", "lstr", "seq", "k", "num", "other"])

def test_match_in_generator():
    def g(items):
        for it in items:
            match it:
                case [x, *_]:
                    yield x
                case _:
                    yield None
    expect(list(g([[1, 2], 3, (4,)]))).to_be([1, None, 4])

# === Exceptions ===

def test_builtin_errors_have_python_types():
    def kind(fn):
        try:
            fn()
        except TypeError:
            return "TypeError"
        except NameError:
            return "NameError"
        except AttributeError:
            return "AttributeError"
        return "none"
    expect(kind(lambda: None * 2)).to_be("TypeError")
    expect(kind(lambda: len(5))).to_be("TypeError")
    expect(kind(lambda: undefined_name_xyz)).to_be("NameError")
    expect(kind(lambda: int.no_such_attr)).to_be("AttributeError")

def test_raw_errors_caught_in_same_frame():
    caught = []
    try:
        undefined_name_xyz
    except NameError:
        caught.append("name")
    try:
        x = 1 + None
    except TypeError:
        caught.append("type")
    expect(caught).to_be(["name", "type"])

def test_custom_exception_attributes_survive_raise():
    class E(Exception):
        def __init__(self, code):
            super().__init__("code %d" % code)
            self.code = code
        def describe(self):
            return "E%d" % self.code
    try:
        raise E(42)
    except E as e:
        expect((e.code, e.describe(), str(e), e.args)).to_be((42, "E42", "code 42", ("code 42",)))

def test_inherited_exception_str():
    class Base(Exception):
        def __str__(self):
            return "custom"
    class Child(Base):
        pass
    try:
        raise Child("m")
    except Exception as e:
        expect(str(e)).to_be("custom")

def test_super_init_reaches_object():
    class A:
        def __init__(self):
            super().__init__()
            self.x = 1
    class B:
        pass
    class C(B):
        def __init__(self):
            B.__init__(self)
            self.y = 2
    expect((A().x, C().y)).to_be((1, 2))

def test_stop_iteration_value():
    def gen():
        yield 1
        return "done"
    g = gen()
    next(g)
    try:
        next(g)
    except StopIteration as e:
        expect(e.value).to_be("done")
    def outer():
        r = yield from gen()
        yield r
    expect(list(outer())).to_be([1, "done"])

def test_finally_state_is_per_frame():
    def finally_return():
        try:
            return 1
        finally:
            return 2
    def finally_break():
        r = []
        for i in range(3):
            try:
                if i == 1:
                    break
                r.append(i)
            finally:
                r.append("f%d" % i)
        return r
    # finally_return used to leave a stale pending return that the next
    # unrelated finally block resumed
    expect((finally_return(), finally_break())).to_be((2, [0, "f0", "f1"]))

def test_call_inside_finally_keeps_callers_return():
    def inner():
        try:
            return "inner"
        finally:
            pass
    def outer():
        try:
            return "outer"
        finally:
            inner()
    expect(outer()).to_be("outer")

def test_pending_return_vs_exceptions():
    def caught_inside_finally():
        try:
            return 1
        finally:
            try:
                raise ValueError
            except ValueError:
                pass
    def escaped_finally():
        try:
            try:
                return 1
            finally:
                raise ValueError
        except ValueError:
            pass
        try:
            pass
        finally:
            pass
        return 2
    def break_cancels_return():
        for i in range(3):
            try:
                return "ret"
            finally:
                break
        return "after"
    expect((caught_inside_finally(), escaped_finally(), break_cancels_return())).to_be((1, 2, "after"))

def test_soft_keywords():
    import re
    match = re.match("a", "abc")
    case = 3
    def f(match):
        return match
    expect((match is not None, case, f(7))).to_be((True, 3, 7))
    match case:
        case 3:
            hit = "matched"
        case _:
            hit = "no"
    expect(hit).to_be("matched")

def test_fstring_forms():
    x = 1
    s = "hi"
    w = 6
    d = {"a:b": 5}
    expect((f"{x=}", f"{ x = }", f"{s=}", f"{x=:>4}", f"{x==1=}")).to_be(("x=1", " x = 1", "s='hi'", "x=   1", "x==1=True"))
    expect((f"{'{'}", f"{'}'}", f"{d['a:b']}", f"{x:'>5}", f"{x:{w}}|")).to_be(("{", "}", "5", "''''1", "     1|"))

def test_nested_with_outer_suppresses():
    log = []
    class CM:
        def __init__(self, n):
            self.n = n
        def __enter__(self):
            return self
        def __exit__(self, et, ev, tb):
            log.append(self.n)
            return self.n == 1
    def f():
        with CM(1):
            with CM(2):
                raise KeyError
        return "after"
    expect((f(), log)).to_be(("after", [2, 1]))

def test_compare_in_finally_during_unwind():
    def g(a, b):
        out = []
        try:
            try:
                raise ValueError("v")
            finally:
                if a < b:
                    out.append("lt")
                if "x" in "xyz":
                    out.append("in")
        except ValueError as e:
            out.append(str(e))
        return out
    expect(g(1, 2)).to_be(["lt", "in", "v"])

def test_function_attributes():
    import functools
    def f(a, b=2, *, c=3):
        return a
    def deco(fn):
        @functools.wraps(fn)
        def inner(*a):
            return fn(*a)
        return inner
    @deco
    def h(x):
        return x
    expect((f.__defaults__, f.__kwdefaults__, h.__name__, h.__wrapped__(5))).to_be(((2,), {"c": 3}, "h", 5))

def test_bare_name_of_deleted_local_raises():
    def f():
        x = 5
        del x
        try:
            x
        except UnboundLocalError:
            return "unbound"
        return "no error"
    expect(f()).to_be("unbound")

def test_instance_class_attribute():
    class C:
        pass
    expect(C().__class__ is C).to_be(True)

def test_generator_recursion_is_catchable():
    # Unbounded recursion through generators used to overflow the Go stack
    # and kill the host process.
    def g(n):
        yield from g(n + 1)
    caught = False
    try:
        list(g(0))
    except RecursionError:
        caught = True
    expect(caught).to_be(True)

test("elif_chain", test_elif_chain)
test("elif_chain_no_else", test_elif_chain_no_else)
test("single_line_suites", test_single_line_suites)
test("comments_in_odd_places", test_comments_in_odd_places)
test("semicolons", test_semicolons)
test("unpacking_displays", test_unpacking_displays)
test("call_unpacking", test_call_unpacking)
test("lambda_signatures", test_lambda_signatures)
test("positional_only_params", test_positional_only_params)
test("kwargs_keep_call_order", test_kwargs_keep_call_order)
test("short_circuit_in_call_args", test_short_circuit_in_call_args)
test("typed_except_inside_loop", test_typed_except_inside_loop)
test("match_in_loop", test_match_in_loop)
test("match_in_generator", test_match_in_generator)
test("builtin_errors_have_python_types", test_builtin_errors_have_python_types)
test("raw_errors_caught_in_same_frame", test_raw_errors_caught_in_same_frame)
test("custom_exception_attributes_survive_raise", test_custom_exception_attributes_survive_raise)
test("inherited_exception_str", test_inherited_exception_str)
test("super_init_reaches_object", test_super_init_reaches_object)
test("stop_iteration_value", test_stop_iteration_value)
test("finally_state_is_per_frame", test_finally_state_is_per_frame)
test("call_inside_finally_keeps_callers_return", test_call_inside_finally_keeps_callers_return)
test("pending_return_vs_exceptions", test_pending_return_vs_exceptions)
test("soft_keywords", test_soft_keywords)
test("fstring_forms", test_fstring_forms)
test("nested_with_outer_suppresses", test_nested_with_outer_suppresses)
test("compare_in_finally_during_unwind", test_compare_in_finally_during_unwind)
test("function_attributes", test_function_attributes)
test("bare_name_of_deleted_local_raises", test_bare_name_of_deleted_local_raises)
test("instance_class_attribute", test_instance_class_attribute)
test("generator_recursion_is_catchable", test_generator_recursion_is_catchable)

print("Parser/VM regression tests completed")
