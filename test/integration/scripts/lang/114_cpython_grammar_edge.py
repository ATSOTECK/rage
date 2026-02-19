# Test: CPython Grammar Edge Cases
# Adapted from CPython's test_grammar.py and test_syntax.py
# Tests assert args, ternary precedence, comprehension edge cases, etc.

from test_framework import test, expect

# === Assert statement: zero-arg AssertionError ===
def test_assert_false_zero_args():
    try:
        assert False
    except AssertionError as e:
        expect(len(e.args)).to_be(0)
        expect(e.args).to_be(())

test("assert_false_zero_args", test_assert_false_zero_args)

# === Assert statement: message becomes args[0] ===
def test_assert_message_in_args():
    try:
        assert False, "custom message"
    except AssertionError as e:
        expect(e.args[0]).to_be("custom message")
        expect(len(e.args)).to_be(1)

test("assert_message_in_args", test_assert_message_in_args)

# === Assert with expression as message ===
def test_assert_expression_message():
    x = 42
    try:
        assert x < 0, "x must be negative, got " + str(x)
    except AssertionError as e:
        expect(e.args[0]).to_be("x must be negative, got 42")

test("assert_expression_message", test_assert_expression_message)

# === Assert with zero integer message ===
def test_assert_zero_message():
    try:
        assert 0, 0
    except AssertionError as e:
        expect(e.args[0]).to_be(0)

test("assert_zero_message", test_assert_zero_message)

# === Assert with tuple message ===
def test_assert_tuple_message():
    try:
        assert False, (1, 2, 3)
    except AssertionError as e:
        expect(e.args[0]).to_be((1, 2, 3))

test("assert_tuple_message", test_assert_tuple_message)

# === Ternary operator: parsed as (5 and 6) if 0 else 1 ===
def test_ternary_and_precedence():
    expect(5 and 6 if 0 else 1).to_be(1)

test("ternary_and_precedence", test_ternary_and_precedence)

# === Ternary: (5 and (6 if 1 else 1)) ===
def test_ternary_and_inner():
    expect(5 and (6 if 1 else 1)).to_be(6)

test("ternary_and_inner", test_ternary_and_inner)

# === Ternary: (not 5) if 1 else 1 ===
def test_ternary_not_precedence():
    expect(not 5 if 1 else 1).to_be(False)

test("ternary_not_precedence", test_ternary_not_precedence)

# === Ternary: (6 + 1) if 1 else 2 ===
def test_ternary_addition():
    expect(6 + 1 if 1 else 2).to_be(7)

test("ternary_addition", test_ternary_addition)

# === Ternary: (6 / 2) if 1 else 3 ===
def test_ternary_division():
    expect(6 / 2 if 1 else 3).to_be(3.0)

test("ternary_division", test_ternary_division)

# === Ternary: (6 < 4) if 0 else 2 ===
def test_ternary_comparison():
    expect(6 < 4 if 0 else 2).to_be(2)

test("ternary_comparison", test_ternary_comparison)

# === Nested ternary ===
def test_nested_ternary():
    def classify(x):
        return "positive" if x > 0 else ("zero" if x == 0 else "negative")
    expect(classify(5)).to_be("positive")
    expect(classify(0)).to_be("zero")
    expect(classify(-3)).to_be("negative")

test("nested_ternary", test_nested_ternary)

# === Ternary in assignment ===
def test_ternary_assignment():
    x = 10
    y = "big" if x > 5 else "small"
    expect(y).to_be("big")
    z = "big" if x < 5 else "small"
    expect(z).to_be("small")

test("ternary_assignment", test_ternary_assignment)

# === Ternary with function calls ===
def test_ternary_with_calls():
    def double(x):
        return x * 2
    result = double(3) if True else double(4)
    expect(result).to_be(6)
    result = double(3) if False else double(4)
    expect(result).to_be(8)

test("ternary_with_calls", test_ternary_with_calls)

# === Ternary short-circuit: false branch not evaluated ===
def test_ternary_short_circuit():
    evaluated = []
    def side_effect(val):
        evaluated.append(val)
        return val
    result = side_effect(1) if True else side_effect(2)
    expect(result).to_be(1)
    expect(evaluated).to_be([1])

test("ternary_short_circuit", test_ternary_short_circuit)

# === Chained comparisons ===
def test_chained_comparison_basic():
    x = 5
    expect(1 < x < 10).to_be(True)
    expect(1 < x > 10).to_be(False)
    expect(0 < 1 < 2 < 3).to_be(True)
    expect(0 < 1 < 2 > 3).to_be(False)

test("chained_comparison_basic", test_chained_comparison_basic)

# === Chained comparison short-circuits ===
def test_chained_comparison_short_circuit():
    calls = []
    def track(val):
        calls.append(val)
        return val
    # 10 < 5 is False, so 5 < 20 should NOT be evaluated
    result = track(10) < track(5) < track(20)
    expect(result).to_be(False)
    expect(calls).to_be([10, 5])

test("chained_comparison_short_circuit", test_chained_comparison_short_circuit)

# === Chained equality ===
def test_chained_equality():
    expect(1 == 1 == 1).to_be(True)
    expect(1 == 1 == 2).to_be(False)

test("chained_equality", test_chained_equality)

# === Chained mixed operators ===
def test_chained_mixed_ops():
    expect(1 < 2 <= 2 < 3).to_be(True)
    expect(1 < 2 <= 2 < 2).to_be(False)

test("chained_mixed_ops", test_chained_mixed_ops)

# === Single-element tuple unpacking in list comprehension ===
def test_single_tuple_unpack_listcomp():
    result = [x for x, in [(4,), (5,), (6,)]]
    expect(result).to_be([4, 5, 6])

test("single_tuple_unpack_listcomp", test_single_tuple_unpack_listcomp)

# === Single-element tuple unpacking in generator expression ===
def test_single_tuple_unpack_genexp():
    result = list(x for x, in [(7,), (8,), (9,)])
    expect(result).to_be([7, 8, 9])

test("single_tuple_unpack_genexp", test_single_tuple_unpack_genexp)

# === Generator expression: outermost iterable is precomputed ===
def test_genexp_outermost_precomputed():
    x = 10
    g = (i for i in range(x))
    x = 5
    expect(len(list(g))).to_be(10)

test("genexp_outermost_precomputed", test_genexp_outermost_precomputed)

# === List comprehension: outermost iterable is precomputed ===
def test_listcomp_outermost_precomputed():
    x = 5
    result = [i for i in range(x)]
    x = 100
    expect(result).to_be([0, 1, 2, 3, 4])

test("listcomp_outermost_precomputed", test_listcomp_outermost_precomputed)

# === Nested generator expressions ===
def test_nested_genexp():
    a = [x for x in range(10)]
    b = (x for x in (y for y in a))
    expect(sum(b)).to_be(sum(range(10)))

test("nested_genexp", test_nested_genexp)

# === Triple-nested generator expressions ===
def test_triple_nested_genexp():
    result = sum(x for x in (y for y in (z for z in range(10))))
    expect(result).to_be(45)

test("triple_nested_genexp", test_triple_nested_genexp)

# === Genexp scope isolation: loop var doesn't leak ===
def test_genexp_scope_isolation():
    i = 20
    total = sum(i * i for i in range(100))
    expect(i).to_be(20)

test("genexp_scope_isolation", test_genexp_scope_isolation)

# === Listcomp scope isolation ===
def test_listcomp_scope_isolation():
    x = 99
    result = [x for x in range(5)]
    expect(x).to_be(99)
    expect(result).to_be([0, 1, 2, 3, 4])

test("listcomp_scope_isolation", test_listcomp_scope_isolation)

# === Multiple assignment targets ===
def test_multiple_assignment():
    a = b = c = 42
    expect(a).to_be(42)
    expect(b).to_be(42)
    expect(c).to_be(42)

test("multiple_assignment", test_multiple_assignment)

# === Multiple assignment with mutable ===
def test_multiple_assignment_mutable():
    a = b = [1, 2, 3]
    a.append(4)
    # Both refer to the same list
    expect(b).to_be([1, 2, 3, 4])

test("multiple_assignment_mutable", test_multiple_assignment_mutable)

# === Lambda with default argument ===
def test_lambda_default():
    f = lambda x, y=10: x + y
    expect(f(1)).to_be(11)
    expect(f(1, 2)).to_be(3)

test("lambda_default", test_lambda_default)

# === Lambda nested with defaults ===
def test_lambda_nested_defaults():
    # Nested lambda where inner uses default from outer scope
    f = lambda x=1: (lambda y=2: x + y)
    expect(f()(3)).to_be(4)
    expect(f(10)()).to_be(12)

test("lambda_nested_defaults", test_lambda_nested_defaults)

# === Deeply nested lambda with defaults ===
def test_lambda_deep_nested():
    f = lambda x=lambda y=lambda z=1: z: y(): x()
    expect(f()).to_be(1)

test("lambda_deep_nested", test_lambda_deep_nested)

# === Comprehension with complex filter ===
def test_comp_complex_filter():
    # Multiple conditions as separate if clauses
    result = [x for x in range(30) if x % 2 == 0 if x % 3 == 0]
    expect(result).to_be([0, 6, 12, 18, 24])

test("comp_complex_filter", test_comp_complex_filter)

# === Comprehension with ternary in body ===
def test_comp_ternary_body():
    result = ["even" if x % 2 == 0 else "odd" for x in range(5)]
    expect(result).to_be(["even", "odd", "even", "odd", "even"])

test("comp_ternary_body", test_comp_ternary_body)

# === Dict comprehension from enumerate ===
def test_dictcomp_enumerate():
    result = {i: c for i, c in enumerate("abc")}
    expect(result).to_be({0: "a", 1: "b", 2: "c"})

test("dictcomp_enumerate", test_dictcomp_enumerate)

# === Set comprehension deduplication ===
def test_setcomp_dedup():
    result = sorted(list({x % 3 for x in range(10)}))
    expect(result).to_be([0, 1, 2])

test("setcomp_dedup", test_setcomp_dedup)

# === Global statement in nested function ===
def test_global_in_nested():
    global _test_global_var
    _test_global_var = 10
    def outer():
        def inner():
            global _test_global_var
            _test_global_var = 99
        inner()
    outer()
    expect(_test_global_var).to_be(99)

test("global_in_nested", test_global_in_nested)

# === Del statement on list elements ===
def test_del_list_elements():
    a = [1, 2, 3, 4, 5]
    del a[1]
    expect(a).to_be([1, 3, 4, 5])
    del a[0:2]
    expect(a).to_be([4, 5])

test("del_list_elements", test_del_list_elements)

# === Del statement on dict keys ===
def test_del_dict_keys():
    d = {"a": 1, "b": 2, "c": 3}
    del d["b"]
    expect(sorted(d.keys())).to_be(["a", "c"])

test("del_dict_keys", test_del_dict_keys)

# === Del variable at module scope ===
def test_del_variable():
    # del removes names; re-assignment after del works
    result = []
    result.append(1)
    result.append(2)
    del result[0]
    expect(result).to_be([2])

test("del_variable", test_del_variable)

# === For/else: else runs when no break ===
def test_for_else_no_break():
    result = []
    for i in range(3):
        result.append(i)
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2, "else"])

test("for_else_no_break", test_for_else_no_break)

# === For/else: else skipped when break ===
def test_for_else_with_break():
    result = []
    for i in range(5):
        if i == 2:
            break
        result.append(i)
    else:
        result.append("else")
    expect(result).to_be([0, 1])

test("for_else_with_break", test_for_else_with_break)

# === While/else: else runs when condition becomes false ===
def test_while_else():
    result = []
    i = 0
    while i < 3:
        result.append(i)
        i = i + 1
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2, "else"])

test("while_else", test_while_else)

# === While/else: else skipped on break ===
def test_while_else_break():
    result = []
    i = 0
    while i < 5:
        if i == 2:
            break
        result.append(i)
        i = i + 1
    else:
        result.append("else")
    expect(result).to_be([0, 1])

test("while_else_break", test_while_else_break)

# === Nested loops: break only exits innermost ===
def test_nested_break():
    result = []
    for i in range(3):
        for j in range(3):
            if j == 1:
                break
            result.append((i, j))
    expect(result).to_be([(0, 0), (1, 0), (2, 0)])

test("nested_break", test_nested_break)

# === Pass statement in various contexts ===
def test_pass_contexts():
    # pass in for
    for i in range(3):
        pass
    expect(i).to_be(2)

    # pass in if
    x = 5
    if x > 10:
        pass
    else:
        x = 0
    expect(x).to_be(0)

    # pass in function
    def noop():
        pass
    expect(noop()).to_be(None)

    # pass in class
    class Empty:
        pass
    expect(type(Empty()).__name__).to_be("Empty")

test("pass_contexts", test_pass_contexts)

# === Starred assignment ===
def test_starred_assignment():
    a, *b, c = [1, 2, 3, 4, 5]
    expect(a).to_be(1)
    expect(b).to_be([2, 3, 4])
    expect(c).to_be(5)

test("starred_assignment", test_starred_assignment)

# === Starred with empty middle ===
def test_starred_empty_middle():
    a, *b, c = [1, 2]
    expect(a).to_be(1)
    expect(b).to_be([])
    expect(c).to_be(2)

test("starred_empty_middle", test_starred_empty_middle)

# === Genexp as sole function argument (no extra parens) ===
def test_genexp_as_argument():
    result = sum(x * x for x in range(5))
    expect(result).to_be(30)
    result = list(x + 1 for x in range(3))
    expect(result).to_be([1, 2, 3])

test("genexp_as_argument", test_genexp_as_argument)

# === Comprehension in try/except ===
def test_comp_in_try():
    try:
        result = [1/x for x in [1, 2, 0, 4]]
    except ZeroDivisionError:
        result = "caught"
    expect(result).to_be("caught")

test("comp_in_try", test_comp_in_try)

# === Boolean operator short-circuit in complex expression ===
def test_bool_short_circuit():
    calls = []
    def f(x):
        calls.append(x)
        return x

    # and short-circuits on falsy
    result = f(0) and f(1)
    expect(result).to_be(0)
    expect(calls).to_be([0])

    calls.clear()
    # or short-circuits on truthy
    result = f(1) or f(2)
    expect(result).to_be(1)
    expect(calls).to_be([1])

test("bool_short_circuit", test_bool_short_circuit)

print("CPython grammar edge case tests completed")
