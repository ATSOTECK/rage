# Test: Error Handling and Panic Fixes
# Tests ZeroDivisionError catching, while+and/or, bare raise

from test_framework import test, expect

# === ZeroDivisionError ===

def test_catch_int_division_by_zero():
    caught = False
    try:
        x = 1 / 0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch int division by zero", test_catch_int_division_by_zero)

def test_catch_float_division_by_zero():
    caught = False
    try:
        x = 1.0 / 0.0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch float division by zero", test_catch_float_division_by_zero)

def test_catch_floor_division_by_zero():
    caught = False
    try:
        x = 1 // 0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch floor division by zero", test_catch_floor_division_by_zero)

def test_catch_modulo_by_zero():
    caught = False
    try:
        x = 1 % 0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch modulo by zero", test_catch_modulo_by_zero)

def test_catch_float_floor_div_by_zero():
    caught = False
    try:
        x = 1.0 // 0.0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch float floor division by zero", test_catch_float_floor_div_by_zero)

def test_catch_float_modulo_by_zero():
    caught = False
    try:
        x = 1.0 % 0.0
    except ZeroDivisionError:
        caught = True
    expect(caught).to_be(True)

test("catch float modulo by zero", test_catch_float_modulo_by_zero)

def test_zerodiv_message():
    msg = ""
    try:
        x = 1 / 0
    except ZeroDivisionError as e:
        msg = str(e)
    expect(msg).to_be("division by zero")

test("ZeroDivisionError message", test_zerodiv_message)

def test_exception_str_no_type_prefix():
    msg = ""
    try:
        raise ValueError("test message")
    except ValueError as e:
        msg = str(e)
    expect(msg).to_be("test message")

test("str(exception) returns message without type prefix", test_exception_str_no_type_prefix)

def test_exception_str_no_args():
    msg = "not empty"
    try:
        raise RuntimeError()
    except RuntimeError as e:
        msg = str(e)
    expect(msg).to_be("")

test("str(exception) with no args returns empty string", test_exception_str_no_args)

def test_zerodiv_uncaught():
    # ZeroDivisionError should not be caught by ValueError
    caught_wrong = False
    caught_right = False
    try:
        try:
            x = 1 / 0
        except ValueError:
            caught_wrong = True
    except ZeroDivisionError:
        caught_right = True
    expect(caught_wrong).to_be(False)
    expect(caught_right).to_be(True)

test("ZeroDivisionError not caught by ValueError", test_zerodiv_uncaught)

# === while + and/or ===

def test_while_and():
    x = 0
    while x < 5 and x >= 0:
        x = x + 1
    expect(x).to_be(5)

test("while with and", test_while_and)

def test_while_or():
    x = 10
    while x > 5 or x == 3:
        x = x - 1
    # x goes 10,9,8,7,6 (all >5), then 5: 5>5=F, 5==3=F -> exit
    expect(x).to_be(5)

test("while with or", test_while_or)

def test_while_and_complex():
    x = 0
    y = 10
    while x < 5 and y > 5:
        x = x + 1
        y = y - 1
    expect(x).to_be(5)
    expect(y).to_be(5)

test("while with and (two variables)", test_while_and_complex)

def test_while_and_short_circuit():
    # Test that and short-circuits: right side never true
    x = 0
    count = 0
    while x < 3 and x >= 0:
        x = x + 1
        count = count + 1
    expect(count).to_be(3)
    expect(x).to_be(3)

test("while with and short-circuit", test_while_and_short_circuit)

# === Bare raise ===

def test_bare_raise():
    caught_type = ""
    try:
        try:
            raise ValueError("test error")
        except ValueError:
            raise  # re-raise
    except ValueError as e:
        caught_type = "ValueError"
    expect(caught_type).to_be("ValueError")

test("bare raise re-raises exception", test_bare_raise)

# === return from while in nested func ===

def test_return_from_while_nested():
    def outer():
        def inner():
            x = 0
            while x < 10:
                x = x + 1
                if x == 5:
                    return x
            return x
        return inner()
    expect(outer()).to_be(5)

test("return from while in nested func", test_return_from_while_nested)
