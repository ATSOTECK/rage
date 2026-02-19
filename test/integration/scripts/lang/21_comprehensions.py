# Test: Comprehensions
# Tests list comprehensions, dict comprehensions, and set comprehensions

from test_framework import test, expect

# Walrus helper function
def walrus_func():
    inner_result = [(walrus_z := j + 100) for j in range(3)]
    return walrus_z

def test_list_comp_simple():
    expect([x for x in range(5)]).to_be([0, 1, 2, 3, 4])

def test_list_comp_expr():
    expect([x * 2 for x in range(5)]).to_be([0, 2, 4, 6, 8])

def test_list_comp_squares():
    expect([x * x for x in range(6)]).to_be([0, 1, 4, 9, 16, 25])

def test_list_comp_filter():
    expect([x for x in range(10) if x % 2 == 0]).to_be([0, 2, 4, 6, 8])

def test_list_comp_multi_filter():
    expect([x for x in range(30) if x % 2 == 0 if x % 3 == 0]).to_be([0, 6, 12, 18, 24])

def test_list_comp_expr_filter():
    expect([x * x for x in range(10) if x % 2 == 1]).to_be([1, 9, 25, 49, 81])

def test_list_comp_string():
    expect([c for c in "hello"]).to_be(["h", "e", "l", "l", "o"])

def test_list_comp_list():
    expect([x + 10 for x in [1, 2, 3, 4, 5]]).to_be([11, 12, 13, 14, 15])

def test_list_comp_empty():
    expect([x for x in []]).to_be([])

def test_list_comp_filter_all():
    expect([x for x in range(5) if x > 100]).to_be([])

def test_list_comp_ternary():
    expect(["even" if x % 2 == 0 else "odd" for x in range(5)]).to_be(["even", "odd", "even", "odd", "even"])

def test_list_comp_negative():
    expect([-x for x in range(5)]).to_be([0, -1, -2, -3, -4])

def test_list_comp_tuples():
    expect([(x, x * x) for x in range(4)]).to_be([(0, 0), (1, 1), (2, 4), (3, 9)])

def test_dict_comp_simple():
    expect({x: x * x for x in range(5)}).to_be({0: 0, 1: 1, 2: 4, 3: 9, 4: 16})

def test_dict_comp_filter():
    expect({x: x * 2 for x in range(10) if x % 2 == 0}).to_be({0: 0, 2: 4, 4: 8, 6: 12, 8: 16})

def test_dict_comp_str_keys():
    expect({str(x): x for x in range(4)}).to_be({"0": 0, "1": 1, "2": 2, "3": 3})

def test_dict_comp_empty():
    expect({x: x for x in []}).to_be({})

def test_set_comp_simple():
    set_result = {x % 3 for x in range(10)}
    expect(len(set_result)).to_be(3)

def test_set_comp_filter():
    set_filtered = {x for x in range(10) if x > 5}
    expect(len(set_filtered)).to_be(4)

def test_list_comp_nested():
    matrix = [[1, 2, 3], [4, 5, 6]]
    expect([x for row in matrix for x in row]).to_be([1, 2, 3, 4, 5, 6])

def test_list_comp_nested_expr():
    expect([x * y for x in [1, 2, 3] for y in [10, 100]]).to_be([10, 100, 20, 200, 30, 300])

def test_walrus_basic():
    expect((walrus_x := 42)).to_be(42)

def test_walrus_expr():
    expect((walrus_y := 10) + 5).to_be(15)
    expect(walrus_y).to_be(10)

def test_walrus_if():
    result = None
    if (walrus_n := 7) > 5:
        result = walrus_n
    expect(result).to_be(7)

def test_walrus_while():
    walrus_count = 0
    walrus_sum = 0
    while (walrus_val := walrus_count) < 3:
        walrus_sum = walrus_sum + walrus_val
        walrus_count = walrus_count + 1
    expect(walrus_sum).to_be(3)
    expect(walrus_val).to_be(3)

def test_walrus_list_comp():
    result = [(walrus_i := i * 2) for i in range(4)]
    expect(result).to_be([0, 2, 4, 6])
    expect(walrus_i).to_be(6)

def test_walrus_dict_comp():
    result = {k: (walrus_v := k * 3) for k in range(3)}
    expect(result).to_be({0: 0, 1: 3, 2: 6})
    expect(walrus_v).to_be(6)

def test_walrus_nested():
    result_a = None
    result_b = None
    if (walrus_a := (walrus_b := 5) + 10) > 10:
        result_a = walrus_a
        result_b = walrus_b
    expect(result_a).to_be(15)
    expect(result_b).to_be(5)

def test_walrus_string():
    expect(len(walrus_s := "hello")).to_be(5)
    expect(walrus_s).to_be("hello")

def test_walrus_in_func():
    expect(walrus_func()).to_be(102)

test("list_comp_simple", test_list_comp_simple)
test("list_comp_expr", test_list_comp_expr)
test("list_comp_squares", test_list_comp_squares)
test("list_comp_filter", test_list_comp_filter)
test("list_comp_multi_filter", test_list_comp_multi_filter)
test("list_comp_expr_filter", test_list_comp_expr_filter)
test("list_comp_string", test_list_comp_string)
test("list_comp_list", test_list_comp_list)
test("list_comp_empty", test_list_comp_empty)
test("list_comp_filter_all", test_list_comp_filter_all)
test("list_comp_ternary", test_list_comp_ternary)
test("list_comp_negative", test_list_comp_negative)
test("list_comp_tuples", test_list_comp_tuples)
test("dict_comp_simple", test_dict_comp_simple)
test("dict_comp_filter", test_dict_comp_filter)
test("dict_comp_str_keys", test_dict_comp_str_keys)
test("dict_comp_empty", test_dict_comp_empty)
test("set_comp_simple", test_set_comp_simple)
test("set_comp_filter", test_set_comp_filter)
test("list_comp_nested", test_list_comp_nested)
test("list_comp_nested_expr", test_list_comp_nested_expr)
test("walrus_basic", test_walrus_basic)
test("walrus_expr", test_walrus_expr)
test("walrus_if", test_walrus_if)
test("walrus_while", test_walrus_while)
test("walrus_list_comp", test_walrus_list_comp)
test("walrus_dict_comp", test_walrus_dict_comp)
test("walrus_nested", test_walrus_nested)
test("walrus_string", test_walrus_string)
test("walrus_in_func", test_walrus_in_func)

# =============================================================================
# Tests ported from CPython's test_listcomps.py
# =============================================================================

# --- Lambda capture of iteration variable as default ---

def test_lambda_capture_iter_var_default():
    """Lambda with default arg captures iteration variable at each step"""
    items = [(lambda i=i: i) for i in range(5)]
    y = [x() for x in items]
    expect(y).to_be([0, 1, 2, 3, 4])

test("lambda capture iter var as default", test_lambda_capture_iter_var_default)

# --- Lambda with free variable (late binding) ---

def test_lambda_free_var_late_binding():
    """Lambda with free variable gets the final value (late binding)"""
    items = [(lambda: i) for i in range(5)]
    y = [x() for x in items]
    expect(y).to_be([4, 4, 4, 4, 4])

test("lambda free var late binding", test_lambda_free_var_late_binding)

# --- Inner variable doesn't shadow outer ---

def test_inner_var_doesnt_shadow_outer():
    """Comprehension iteration variable does not leak to outer scope"""
    i = 20
    result = [i * i for i in range(5)]
    expect(result).to_be([0, 1, 4, 9, 16])
    expect(i).to_be(20)

test("inner var doesnt shadow outer", test_inner_var_doesnt_shadow_outer)

def test_inner_cell_shadows_outer():
    """Lambda in comp captures comp's cell, not outer"""
    items = [(lambda: i) for i in range(5)]
    i = 20
    y = [x() for x in items]
    expect(y).to_be([4, 4, 4, 4, 4])
    expect(i).to_be(20)

test("inner cell shadows outer", test_inner_cell_shadows_outer)

def test_inner_cell_shadows_outer_redefined():
    """Outer variable with same name is not affected by comp"""
    y = 10
    items = [(lambda: y) for y in range(5)]
    x = y
    y = 20
    out = [z() for z in items]
    expect(x).to_be(10)
    expect(out).to_be([4, 4, 4, 4, 4])

test("inner cell shadows outer redefined", test_inner_cell_shadows_outer_redefined)

# --- Closure can jump over comp scope ---

def test_closure_jumps_over_comp_scope():
    """Lambda inside comp can reference outer variable defined later"""
    items = [(lambda: y) for i in range(5)]
    y = 2
    z = [x() for x in items]
    expect(z).to_be([2, 2, 2, 2, 2])

test("closure can jump over comp scope", test_closure_jumps_over_comp_scope)

# --- Walrus operator (:=) in comprehensions ---

def test_walrus_in_comp_assigns_to_outer():
    """Walrus operator in comprehension assigns to outer scope"""
    x = -1
    items = [(x := y) for y in range(3)]
    expect(items).to_be([0, 1, 2])
    expect(x).to_be(2)

test("walrus in comp assigns to outer", test_walrus_in_comp_assigns_to_outer)

def test_walrus_accumulator():
    """Walrus operator used as accumulator in comprehension"""
    b = 0
    res = [b := i + b for i in range(5)]
    expect(res).to_be([0, 1, 3, 6, 10])
    expect(b).to_be(10)

test("walrus accumulator", test_walrus_accumulator)

def test_walrus_in_comp_with_function():
    """Walrus inside comprehension calling function"""
    def spam(a):
        return a
    res = [[y := spam(x), x / y] for x in range(1, 5)]
    expect(res).to_be([[1, 1.0], [2, 1.0], [3, 1.0], [4, 1.0]])
    expect(y).to_be(4)

test("walrus in comp with function", test_walrus_in_comp_with_function)

def test_walrus_in_comp_filter():
    """Walrus in comprehension filter clause"""
    def spam(a):
        return a
    input_data = [1, 2, 3]
    res = [(x, y, x / y) for x in input_data if (y := spam(x)) > 0]
    expect(res).to_be([(1, 1, 1.0), (2, 2, 1.0), (3, 3, 1.0)])
    expect(y).to_be(3)

test("walrus in comp filter", test_walrus_in_comp_filter)

def test_walrus_nested_comps():
    """Walrus in nested comprehensions"""
    res = [[spam := i for i in range(3)] for j in range(2)]
    expect(res).to_be([[0, 1, 2], [0, 1, 2]])
    expect(spam).to_be(2)

test("walrus nested comps", test_walrus_nested_comps)

def test_walrus_nested_comp_with_list():
    """Walrus in nested comp creating inner lists"""
    res = [b := [a := 1 for i in range(2)] for j in range(2)]
    expect(res).to_be([[1, 1], [1, 1]])
    expect(a).to_be(1)
    expect(b).to_be([1, 1])

test("walrus nested comp with list", test_walrus_nested_comp_with_list)

def test_walrus_simple_rename():
    """Walrus simple rename in comprehension"""
    res = [j := i for i in range(5)]
    expect(res).to_be([0, 1, 2, 3, 4])
    expect(j).to_be(4)

test("walrus simple rename", test_walrus_simple_rename)

# --- Nested comprehensions with free variables ---

def test_nested_comp_dependent():
    """Nested comprehension where inner depends on outer"""
    l = [2, 3]
    y = [[x ** 2 for x in range(x)] for x in l]
    expect(y).to_be([[0, 1], [0, 1, 4]])

test("nested comp dependent", test_nested_comp_dependent)

def test_nested_comp_with_lambda():
    """Nested comprehension with lambda"""
    f = [(z, lambda y: [(x, y, z) for x in [3]]) for z in [1]]
    (z, func) = f[0]
    out = func(2)
    expect(z).to_be(1)
    expect(out).to_be([(3, 2, 1)])

test("nested comp with lambda", test_nested_comp_with_lambda)

def test_shadow_comp_iterable_name():
    """Comprehension can iterate over a variable it shadows"""
    x = [1]
    y = [x for x in x]
    expect(x).to_be([1])

test("shadow comp iterable name", test_shadow_comp_iterable_name)

# --- Simple nesting (from doctests) ---

def test_simple_nesting():
    """Simple nesting from CPython doctests"""
    result = [(i, j) for i in range(3) for j in range(4)]
    expected = [(0, 0), (0, 1), (0, 2), (0, 3),
                (1, 0), (1, 1), (1, 2), (1, 3),
                (2, 0), (2, 1), (2, 2), (2, 3)]
    expect(result).to_be(expected)

test("simple nesting", test_simple_nesting)

def test_nesting_inner_depends_outer():
    """Nesting with inner expression dependent on outer"""
    result = [(i, j) for i in range(4) for j in range(i)]
    expect(result).to_be([(1, 0), (2, 0), (2, 1), (3, 0), (3, 1), (3, 2)])

test("nesting inner depends outer", test_nesting_inner_depends_outer)

def test_temp_var_idiom():
    """Temporary variable assignment idiom in comprehensions"""
    result = [j * j for i in range(4) for j in [i + 1]]
    expect(result).to_be([1, 4, 9, 16])

test("temp var idiom", test_temp_var_idiom)

def test_temp_var_idiom_two_vars():
    """Temporary variable assignment with two derived vars"""
    result = [j * k for i in range(4) for j in [i + 1] for k in [j + 1]]
    expect(result).to_be([2, 6, 12, 20])

test("temp var idiom two vars", test_temp_var_idiom_two_vars)

def test_temp_var_tuple_unpack():
    """Temporary variable assignment with tuple unpacking"""
    result = [j * k for i in range(4) for j, k in [(i + 1, i + 2)]]
    expect(result).to_be([2, 6, 12, 20])

test("temp var tuple unpack", test_temp_var_tuple_unpack)

def test_none_values():
    """Comprehension producing None values"""
    result = [None for i in range(5)]
    expect(result).to_be([None, None, None, None, None])

test("none values in comp", test_none_values)

def test_sum_with_condition():
    """Sum of squares of odd numbers"""
    result = sum([i * i for i in range(100) if i & 1 == 1])
    expect(result).to_be(166650)

test("sum with condition", test_sum_with_condition)

def test_frange_with_comp():
    """Function returning a range-like list via comprehension"""
    def frange(n):
        return [i for i in range(n)]
    expect(frange(10)).to_be([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])

test("frange with comp", test_frange_with_comp)

def test_lambda_range():
    """Lambda that returns a range-like list via comprehension"""
    lrange = lambda n: [i for i in range(n)]
    expect(lrange(10)).to_be([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])

test("lambda range", test_lambda_range)

# --- Comp in function scope ---

def test_comp_in_function_scope():
    """Comprehension variable doesn't leak in function scope"""
    def f():
        x = 10
        result = [x for x in range(5)]
        return x, result
    outer_x, result = f()
    expect(outer_x).to_be(10)
    expect(result).to_be([0, 1, 2, 3, 4])

test("comp in function scope", test_comp_in_function_scope)

def test_walrus_in_func_comp():
    """Walrus operator in comprehension inside function"""
    def f():
        total = 0
        partial_sums = [total := total + v for v in range(5)]
        return total, partial_sums
    total, partial_sums = f()
    expect(partial_sums).to_be([0, 1, 3, 6, 10])
    expect(total).to_be(10)

test("walrus in func comp", test_walrus_in_func_comp)

def test_walrus_in_func_with_calls():
    """Walrus in comprehension with function calls inside function"""
    def spam(a):
        return a
    def eggs(b):
        return b * 2
    res = [spam(a := eggs(b := h)) for h in range(2)]
    expect(res).to_be([0, 2])
    expect(a).to_be(2)
    expect(b).to_be(1)

test("walrus in func with calls", test_walrus_in_func_with_calls)

def test_walrus_reassign_same_name():
    """Walrus operator reassigning same name twice"""
    def spam(a):
        return a
    def eggs(b):
        return b * 2
    res = [spam(a := eggs(a := h)) for h in range(2)]
    expect(res).to_be([0, 2])
    expect(a).to_be(2)

test("walrus reassign same name", test_walrus_reassign_same_name)

# --- Nested free var in comprehension filter ---

def test_nested_has_free_var():
    """Nested comp with free var in filter"""
    items = [a for a in [1] if [a for _ in [0]]]
    expect(items).to_be([1])

test("nested has free var", test_nested_has_free_var)

def test_nested_free_var_in_expr():
    """Nested comprehension with free var in expression"""
    items = [(_C, [x for x in [1] if _C]) for _C in [0, 1]]
    expect(items).to_be([(0, []), (1, [1])])

test("nested free var in expr", test_nested_free_var_in_expr)

# --- Comprehension in try/except ---

def test_comp_in_try_except_no_exception():
    """Comprehension in try block, no exception raised"""
    value = ["ab"]
    result = None
    snapshot = None
    try:
        result = [len(value) for value in value]
    except ValueError:
        snapshot = value
    expect(value).to_be(["ab"])
    expect(result).to_be([2])
    expect(snapshot).to_be(None)

test("comp in try except no exception", test_comp_in_try_except_no_exception)

def test_comp_in_try_except_with_exception():
    """Comprehension in try block, exception raised"""
    value = ["ab"]
    result = None
    snapshot = None
    raised = False
    try:
        result = [int(value) for value in value]
    except ValueError:
        snapshot = value
        raised = True
    expect(value).to_be(["ab"])
    expect(result).to_be(None)
    expect(snapshot).to_be(["ab"])
    expect(raised).to_be(True)

test("comp in try except with exception", test_comp_in_try_except_with_exception)

# --- Assignment expression in comprehension ---

def test_walrus_assign_expr():
    """Assignment expression scope (from CPython)"""
    x = -1
    items = [(x := y) for y in range(3)]
    expect(items).to_be([0, 1, 2])
    expect(x).to_be(2)

test("walrus assignment expr", test_walrus_assign_expr)

# =============================================================================
# Tests ported from CPython's test_named_expressions.py
# =============================================================================

# --- Basic walrus in various contexts ---

def test_named_expr_basic_assign():
    """Basic named expression assignment"""
    (a := 10)
    expect(a).to_be(10)

test("named expr basic assign", test_named_expr_basic_assign)

def test_named_expr_self_assign():
    """Named expression self-assignment"""
    a = 20
    (a := a)
    expect(a).to_be(20)

test("named expr self assign", test_named_expr_self_assign)

def test_named_expr_arithmetic():
    """Named expression with arithmetic"""
    (total := 1 + 2)
    expect(total).to_be(3)

test("named expr arithmetic", test_named_expr_arithmetic)

def test_named_expr_tuple():
    """Named expression with tuple"""
    (info := (1, 2, 3))
    expect(info).to_be((1, 2, 3))

test("named expr tuple", test_named_expr_tuple)

def test_named_expr_chained():
    """Chained named expressions"""
    (z := (y := (x := 0)))
    expect(x).to_be(0)
    expect(y).to_be(0)
    expect(z).to_be(0)

test("named expr chained", test_named_expr_chained)

# --- Walrus in if conditions ---

def test_walrus_in_if_string():
    """Walrus operator in if condition with string"""
    result = None
    if spam := "eggs":
        result = spam
    expect(result).to_be("eggs")

test("walrus in if string", test_walrus_in_if_string)

def test_walrus_in_if_and():
    """Walrus operator in if with 'and'"""
    result = None
    if True and (spam := True):
        result = spam
    expect(result).to_be(True)

test("walrus in if and", test_walrus_in_if_and)

def test_walrus_in_if_comparison():
    """Walrus operator in if with comparison"""
    result = None
    if (m := 10) == 10:
        result = m
    expect(result).to_be(10)

test("walrus in if comparison", test_walrus_in_if_comparison)

# --- Walrus in while loops ---

def test_walrus_in_while_false():
    """Walrus operator in while that is immediately false"""
    body_executed = False
    while a := False:
        body_executed = True
    expect(body_executed).to_be(False)
    expect(a).to_be(False)

test("walrus in while false", test_walrus_in_while_false)

def test_walrus_in_while_counting():
    """Walrus operator in while loop for counting"""
    a = 9
    n = 2
    x = 3
    # Floor of nth root algorithm
    while a > (d := x // a ** (n - 1)):
        a = ((n - 1) * a + d) // n
    expect(a).to_be(1)

test("walrus in while counting", test_walrus_in_while_counting)

# --- Walrus in list comprehensions ---

def test_walrus_partial_sums():
    """Walrus operator for partial sums in list comprehension"""
    total = 0
    partial_sums = [total := total + v for v in range(5)]
    expect(partial_sums).to_be([0, 1, 3, 6, 10])
    expect(total).to_be(10)

test("walrus partial sums", test_walrus_partial_sums)

def test_walrus_comp_with_filter_and_func():
    """Walrus in comp with filter and function call"""
    def spam(a):
        return a
    input_data = [1, 2, 3]
    res = [(x, y, x / y) for x in input_data if (y := spam(x)) > 0]
    expect(res).to_be([(1, 1, 1.0), (2, 2, 1.0), (3, 3, 1.0)])

test("walrus comp with filter and func", test_walrus_comp_with_filter_and_func)

def test_walrus_comp_body_and_filter():
    """Walrus in list comprehension body and filter expression"""
    def spam(a):
        return a
    res = [[y := spam(x), x / y] for x in range(1, 5)]
    expect(res).to_be([[1, 1.0], [2, 1.0], [3, 1.0], [4, 1.0]])

test("walrus comp body and filter", test_walrus_comp_body_and_filter)

# --- Walrus in function arguments ---

def test_walrus_in_function_arg():
    """Walrus operator in function argument"""
    def spam(a):
        return a
    res = spam(b := 2)
    expect(res).to_be(2)
    expect(b).to_be(2)

test("walrus in function arg", test_walrus_in_function_arg)

def test_walrus_in_function_arg_parens():
    """Walrus operator in parenthesized function argument"""
    def spam(a):
        return a
    res = spam((b := 2))
    expect(res).to_be(2)
    expect(b).to_be(2)

test("walrus in function arg parens", test_walrus_in_function_arg_parens)

def test_walrus_in_keyword_arg():
    """Walrus operator in keyword argument value"""
    def spam(a):
        return a
    res = spam(a=(b := 2))
    expect(res).to_be(2)
    expect(b).to_be(2)

test("walrus in keyword arg", test_walrus_in_keyword_arg)

def test_walrus_as_positional_with_keyword():
    """Walrus operator as positional arg with keyword arg"""
    def spam(a, b):
        return a + b
    res = spam(c := 2, b=1)
    expect(res).to_be(3)
    expect(c).to_be(2)

test("walrus as positional with keyword", test_walrus_as_positional_with_keyword)

def test_walrus_in_len_call():
    """Walrus operator inside len() call"""
    length = len(lines := [1, 2])
    expect(length).to_be(2)
    expect(lines).to_be([1, 2])

test("walrus in len call", test_walrus_in_len_call)

# --- Multiple walrus in same expression ---

def test_multiple_walrus_same_expr():
    """Multiple walrus operators in same expression"""
    result = (a := 1) + (b := 2) + (c := 3)
    expect(result).to_be(6)
    expect(a).to_be(1)
    expect(b).to_be(2)
    expect(c).to_be(3)

test("multiple walrus same expr", test_multiple_walrus_same_expr)

def test_walrus_in_subscript():
    """Walrus operator used in subscript"""
    a = [1, 2, 3]
    element = a[b := 0]
    expect(b).to_be(0)
    expect(element).to_be(1)

test("walrus in subscript", test_walrus_in_subscript)

# --- Walrus scope in comprehensions ---

def test_walrus_scope_comp_leaks():
    """Walrus in comprehension leaks to enclosing scope"""
    res = [j := i for i in range(5)]
    expect(res).to_be([0, 1, 2, 3, 4])
    expect(j).to_be(4)

test("walrus scope comp leaks", test_walrus_scope_comp_leaks)

def test_walrus_scope_nested_comp():
    """Walrus in nested comprehensions"""
    res = [[spam := i for i in range(3)] for j in range(2)]
    expect(res).to_be([[0, 1, 2], [0, 1, 2]])
    expect(spam).to_be(2)

test("walrus scope nested comp", test_walrus_scope_nested_comp)

def test_walrus_scope_in_function():
    """Walrus scope inside a function"""
    def f():
        def spam(a):
            return a
        res = [[y := spam(x), x / y] for x in range(1, 5)]
        return y, res
    y, res = f()
    expect(y).to_be(4)

test("walrus scope in function", test_walrus_scope_in_function)

def test_walrus_nonlocal():
    """Walrus with nonlocal variable"""
    a = 10
    def spam():
        nonlocal a
        (a := 20)
    spam()
    expect(a).to_be(20)

test("walrus nonlocal", test_walrus_nonlocal)

# --- Walrus with dict comprehension (from CPython scope tests) ---

def test_walrus_fib_dict_comp():
    """Fibonacci via walrus in dict comprehension"""
    a, b = 1, 2
    fib = {(c := a): (a := b) + (b := a + c) - b for __ in range(6)}
    expect(fib).to_be({1: 2, 2: 3, 3: 5, 5: 8, 8: 13, 13: 21})

test("walrus fib dict comp", test_walrus_fib_dict_comp)

print("Comprehension tests completed")
