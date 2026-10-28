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

test("elif_chain", test_elif_chain)
test("single_line_suites", test_single_line_suites)
test("comments_in_odd_places", test_comments_in_odd_places)
test("semicolons", test_semicolons)
test("soft_keywords", test_soft_keywords)
test("fstring_forms", test_fstring_forms)

print("Parser/VM regression tests completed")
