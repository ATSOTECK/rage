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
test("short_circuit_in_call_args", test_short_circuit_in_call_args)
test("typed_except_inside_loop", test_typed_except_inside_loop)
test("builtin_errors_have_python_types", test_builtin_errors_have_python_types)
test("raw_errors_caught_in_same_frame", test_raw_errors_caught_in_same_frame)
test("soft_keywords", test_soft_keywords)
test("fstring_forms", test_fstring_forms)
test("nested_with_outer_suppresses", test_nested_with_outer_suppresses)
test("compare_in_finally_during_unwind", test_compare_in_finally_during_unwind)
test("bare_name_of_deleted_local_raises", test_bare_name_of_deleted_local_raises)
test("generator_recursion_is_catchable", test_generator_recursion_is_catchable)

print("Parser/VM regression tests completed")
