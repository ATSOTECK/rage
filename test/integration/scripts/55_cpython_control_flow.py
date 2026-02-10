# Test: CPython Control Flow Edge Cases
# Adapted from CPython's test_syntax.py, test_for.py, test_while.py

from test_framework import test, expect

# === for/else ===
def test_for_else_no_break():
    result = []
    for i in range(3):
        result.append(i)
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2, "else"])

def test_for_else_with_break():
    result = []
    for i in range(5):
        if i == 3:
            break
        result.append(i)
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2])

def test_for_else_empty_iterable():
    result = []
    for i in []:
        result.append(i)
    else:
        result.append("else")
    expect(result).to_be(["else"])

# === while/else ===
def test_while_else_no_break():
    result = []
    i = 0
    while i < 3:
        result.append(i)
        i = i + 1
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2, "else"])

def test_while_else_with_break():
    result = []
    i = 0
    while i < 5:
        if i == 3:
            break
        result.append(i)
        i = i + 1
    else:
        result.append("else")
    expect(result).to_be([0, 1, 2])

def test_while_else_false_condition():
    result = []
    while False:
        result.append("body")
    else:
        result.append("else")
    expect(result).to_be(["else"])

# === Nested loop break ===
def test_nested_break_only_inner():
    result = []
    for i in range(3):
        for j in range(3):
            if j == 1:
                break
            result.append(j)
        result.append(i)
    expect(result).to_be([0, 0, 0, 1, 0, 2])

# === Continue in nested loops ===
def test_continue_in_nested():
    result = []
    for i in range(3):
        for j in range(3):
            if j == 1:
                continue
            result.append(j)
    expect(result).to_be([0, 2, 0, 2, 0, 2])

def test_continue_outer_loop():
    result = []
    for i in range(4):
        if i == 2:
            continue
        result.append(i)
    expect(result).to_be([0, 1, 3])

# === Loop variable scope after exit ===
def test_loop_var_after_for():
    for x in range(5):
        pass
    expect(x).to_be(4)

def test_loop_var_after_break():
    for x in range(10):
        if x == 3:
            break
    expect(x).to_be(3)

# === Nested loops with break/continue combo ===
def test_nested_break_continue():
    result = []
    for i in range(3):
        if i == 1:
            continue
        for j in range(3):
            if j == 2:
                break
            result.append(i * 10 + j)
    expect(result).to_be([0, 1, 20, 21])

# === while with complex conditions ===
def test_while_complex_condition():
    # Avoid 'and' directly in while condition (RAGE limitation)
    result = []
    x = 0
    y = 10
    def keep_going():
        return x < 5 and y > 5
    while keep_going():
        result.append(x)
        x = x + 1
        y = y - 1
    expect(result).to_be([0, 1, 2, 3, 4])

def test_while_or_condition():
    # Avoid 'or' directly in while condition (RAGE limitation)
    count = 0
    x = 0
    def should_continue():
        return x < 3 or count < 2
    while should_continue():
        count = count + 1
        x = x + 1
    expect(count).to_be(3)

# === Pass statement ===
def test_pass_in_for():
    for i in range(5):
        pass
    expect(i).to_be(4)

def test_pass_in_if():
    x = 5
    if x > 10:
        pass
    else:
        x = 0
    expect(x).to_be(0)

def test_pass_in_function():
    def noop():
        pass
    result = noop()
    expect(result).to_be(None)

def test_pass_in_class():
    class Empty:
        pass
    e = Empty()
    e.x = 42
    expect(e.x).to_be(42)

# === Return from nested loops ===
def test_return_from_inner_loop():
    def find_pair(target):
        for i in range(5):
            for j in range(5):
                if i + j == target:
                    return [i, j]
        return None
    expect(find_pair(3)).to_be([0, 3])
    expect(find_pair(100)).to_be(None)

def test_return_from_while():
    def first_over(lst, threshold):
        i = 0
        while i < len(lst):
            if lst[i] > threshold:
                return lst[i]
            i = i + 1
        return None
    expect(first_over([1, 5, 3, 8, 2], 4)).to_be(5)
    expect(first_over([1, 2, 3], 10)).to_be(None)

# Register all tests
test("for_else_no_break", test_for_else_no_break)
test("for_else_with_break", test_for_else_with_break)
test("for_else_empty_iterable", test_for_else_empty_iterable)
test("while_else_no_break", test_while_else_no_break)
test("while_else_with_break", test_while_else_with_break)
test("while_else_false_condition", test_while_else_false_condition)
test("nested_break_only_inner", test_nested_break_only_inner)
test("continue_in_nested", test_continue_in_nested)
test("continue_outer_loop", test_continue_outer_loop)
test("loop_var_after_for", test_loop_var_after_for)
test("loop_var_after_break", test_loop_var_after_break)
test("nested_break_continue", test_nested_break_continue)
test("while_complex_condition", test_while_complex_condition)
test("while_or_condition", test_while_or_condition)
test("pass_in_for", test_pass_in_for)
test("pass_in_if", test_pass_in_if)
test("pass_in_function", test_pass_in_function)
test("pass_in_class", test_pass_in_class)
test("return_from_inner_loop", test_return_from_inner_loop)
test("return_from_while", test_return_from_while)

print("CPython control flow tests completed")
