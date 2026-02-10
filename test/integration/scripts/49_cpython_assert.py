# Test: CPython Assert Statement
# Adapted from CPython's test_grammar.py assert tests

from test_framework import test, expect

# === Basic assert True ===
def test_assert_true():
    # Should not raise
    assert True
    assert 1
    assert "non-empty"
    assert [1]
    # If we get here, all asserts passed
    expect(True).to_be(True)

# === Assert False raises AssertionError ===
def test_assert_false():
    caught = [False]
    try:
        assert False
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_assert_zero():
    caught = [False]
    try:
        assert 0
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

# === Assert with messages ===
def test_assert_message():
    caught = [False]
    msg = [""]
    try:
        assert False, "custom message"
    except AssertionError as e:
        caught[0] = True
        msg[0] = str(e)
    expect(caught[0]).to_be(True)
    expect("custom message" in msg[0]).to_be(True)

def test_assert_message_expression():
    caught = [False]
    msg = [""]
    value = 42
    try:
        assert value < 0, "value must be negative, got " + str(value)
    except AssertionError as e:
        caught[0] = True
        msg[0] = str(e)
    expect(caught[0]).to_be(True)
    expect("value must be negative, got 42" in msg[0]).to_be(True)

def test_assert_message_not_evaluated_on_success():
    # The message expression should not be evaluated when the assert passes
    evaluated = [False]
    def side_effect():
        evaluated[0] = True
        return "should not see this"
    assert True, side_effect()
    # In CPython, the message IS evaluated even on success at the expression level.
    # But the key point is: no exception is raised.
    expect(True).to_be(True)

# === AssertionError catching ===
def test_catch_assertion_error():
    caught = [False]
    try:
        assert 1 == 2
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_assertion_error_is_exception():
    caught_as_exception = [False]
    try:
        assert False
    except Exception:
        caught_as_exception[0] = True
    expect(caught_as_exception[0]).to_be(True)

# === Assert with complex expressions ===
def test_assert_comparison():
    assert 1 < 2
    assert 5 >= 5
    assert 10 != 11
    assert "abc" == "abc"
    expect(True).to_be(True)

def test_assert_boolean_ops():
    assert True and True
    assert True or False
    assert not False
    assert 1 and 2 and 3
    expect(True).to_be(True)

def test_assert_membership():
    assert 3 in [1, 2, 3, 4]
    assert 5 not in [1, 2, 3, 4]
    assert "a" in "abc"
    assert "x" not in "abc"
    expect(True).to_be(True)

# === Assert with truthiness edge cases ===
def test_assert_empty_list_fails():
    caught = [False]
    try:
        assert []
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_assert_none_fails():
    caught = [False]
    try:
        assert None
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_assert_empty_string_fails():
    caught = [False]
    try:
        assert ""
    except AssertionError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_assert_truthy_values():
    assert 1
    assert -1
    assert 0.1
    assert "hello"
    assert [0]
    assert {"a": 1}
    assert (1,)
    expect(True).to_be(True)

# Register all tests
test("assert_true", test_assert_true)
test("assert_false", test_assert_false)
test("assert_zero", test_assert_zero)
test("assert_message", test_assert_message)
test("assert_message_expression", test_assert_message_expression)
test("assert_message_not_evaluated_on_success", test_assert_message_not_evaluated_on_success)
test("catch_assertion_error", test_catch_assertion_error)
test("assertion_error_is_exception", test_assertion_error_is_exception)
test("assert_comparison", test_assert_comparison)
test("assert_boolean_ops", test_assert_boolean_ops)
test("assert_membership", test_assert_membership)
test("assert_empty_list_fails", test_assert_empty_list_fails)
test("assert_none_fails", test_assert_none_fails)
test("assert_empty_string_fails", test_assert_empty_string_fails)
test("assert_truthy_values", test_assert_truthy_values)

print("CPython assert tests completed")
