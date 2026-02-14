# Test: CPython Exception Hierarchy and Edge Cases
# Adapted from CPython's test_exceptions.py

from test_framework import test, expect

# === Catching parent catches child ===
def test_exception_hierarchy_valueerror():
    caught = [False]
    try:
        raise ValueError("bad value")
    except Exception:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_exception_hierarchy_typeerror():
    caught = [False]
    try:
        raise TypeError("bad type")
    except Exception:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_exception_hierarchy_keyerror():
    caught = [False]
    try:
        d = {}
        x = d["missing"]
    except Exception:
        caught[0] = True
    expect(caught[0]).to_be(True)

# === Multiple exception types in single except ===
def test_except_multiple_types_first():
    caught = [False]
    try:
        raise ValueError("val")
    except (ValueError, KeyError):
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_except_multiple_types_second():
    caught = [False]
    try:
        raise KeyError("key")
    except (ValueError, KeyError):
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_except_multiple_types_miss():
    caught_right = [False]
    try:
        raise TypeError("type")
    except (ValueError, KeyError):
        caught_right[0] = False
    except TypeError:
        caught_right[0] = True
    expect(caught_right[0]).to_be(True)

# === Exception message preservation ===
def test_exception_message_preserved():
    msg = [""]
    try:
        raise ValueError("hello world")
    except ValueError as e:
        msg[0] = str(e)
    expect("hello world" in msg[0]).to_be(True)

def test_exception_message_via_str():
    msg = [""]
    try:
        raise ValueError("test message")
    except ValueError as e:
        msg[0] = str(e)
    expect("test message" in msg[0]).to_be(True)

# === Re-raising exceptions with explicit raise ===
def test_reraise_explicit():
    caught_outer = [False]
    try:
        try:
            raise ValueError("inner")
        except ValueError as e:
            raise ValueError("re-raised: " + str(e))
    except ValueError as e:
        caught_outer[0] = True
        expect("re-raised" in str(e)).to_be(True)
    expect(caught_outer[0]).to_be(True)

# === Custom exception classes ===
def test_custom_exception():
    class MyError(Exception):
        pass
    caught = [False]
    try:
        raise MyError("custom")
    except MyError as e:
        caught[0] = True
        expect("custom" in str(e)).to_be(True)
    expect(caught[0]).to_be(True)

def test_custom_exception_caught_by_parent():
    class MyError(Exception):
        pass
    caught = [False]
    try:
        raise MyError("custom")
    except Exception:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_custom_exception_hierarchy():
    class AppError(Exception):
        pass
    class DatabaseError(AppError):
        pass
    caught = [False]
    try:
        raise DatabaseError("db failed")
    except AppError as e:
        caught[0] = True
        expect("db failed" in str(e)).to_be(True)
    expect(caught[0]).to_be(True)

# === Built-in exception types ===
def test_value_error_int_conversion():
    # int() with invalid string raises ValueError
    caught = [False]
    try:
        x = int("not_a_number")
    except ValueError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_index_error():
    caught = [False]
    try:
        lst = [1, 2, 3]
        x = lst[10]
    except IndexError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_key_error():
    caught = [False]
    try:
        d = {"a": 1}
        x = d["b"]
    except KeyError:
        caught[0] = True
    expect(caught[0]).to_be(True)

def test_runtime_error():
    caught = [False]
    try:
        raise RuntimeError("runtime issue")
    except RuntimeError as e:
        caught[0] = True
        expect("runtime issue" in str(e)).to_be(True)
    expect(caught[0]).to_be(True)

# === Nested try/except/finally ===
def test_nested_try_except():
    results = []
    try:
        results.append("outer try")
        try:
            results.append("inner try")
            raise ValueError("inner")
        except ValueError:
            results.append("inner except")
        results.append("after inner")
    except ValueError:
        results.append("outer except")
    results.append("done")
    expect(results).to_be(["outer try", "inner try", "inner except", "after inner", "done"])

def test_finally_runs_no_exception():
    results = []
    try:
        results.append("try")
    finally:
        results.append("finally")
    results.append("after")
    expect(results).to_be(["try", "finally", "after"])

def test_finally_runs_with_exception():
    results = []
    try:
        try:
            results.append("try")
            raise ValueError("err")
        finally:
            results.append("finally")
    except ValueError:
        results.append("caught")
    expect(results).to_be(["try", "finally", "caught"])

# === try/except/else clause ===
def test_try_else_no_exception():
    results = []
    try:
        results.append("try")
    except ValueError:
        results.append("except")
    else:
        results.append("else")
    expect(results).to_be(["try", "else"])

def test_try_else_with_exception():
    results = []
    try:
        results.append("try")
        raise ValueError("err")
    except ValueError:
        results.append("except")
    else:
        results.append("else")
    expect(results).to_be(["try", "except"])

# === try/except/else/finally ===
def test_try_except_else_finally_no_error():
    results = []
    try:
        results.append("try")
    except ValueError:
        results.append("except")
    else:
        results.append("else")
    finally:
        results.append("finally")
    expect(results).to_be(["try", "else", "finally"])

def test_try_except_else_finally_with_error():
    results = []
    try:
        results.append("try")
        raise ValueError("err")
    except ValueError:
        results.append("except")
    else:
        results.append("else")
    finally:
        results.append("finally")
    expect(results).to_be(["try", "except", "finally"])

# Register all tests
test("exception_hierarchy_valueerror", test_exception_hierarchy_valueerror)
test("exception_hierarchy_typeerror", test_exception_hierarchy_typeerror)
test("exception_hierarchy_keyerror", test_exception_hierarchy_keyerror)
test("except_multiple_types_first", test_except_multiple_types_first)
test("except_multiple_types_second", test_except_multiple_types_second)
test("except_multiple_types_miss", test_except_multiple_types_miss)
test("exception_message_preserved", test_exception_message_preserved)
test("exception_message_via_str", test_exception_message_via_str)
test("reraise_explicit", test_reraise_explicit)
test("custom_exception", test_custom_exception)
test("custom_exception_caught_by_parent", test_custom_exception_caught_by_parent)
test("custom_exception_hierarchy", test_custom_exception_hierarchy)
test("value_error_int_conversion", test_value_error_int_conversion)
test("index_error", test_index_error)
test("key_error", test_key_error)
test("runtime_error", test_runtime_error)
test("nested_try_except", test_nested_try_except)
test("finally_runs_no_exception", test_finally_runs_no_exception)
test("finally_runs_with_exception", test_finally_runs_with_exception)
test("try_else_no_exception", test_try_else_no_exception)
test("try_else_with_exception", test_try_else_with_exception)
test("try_except_else_finally_no_error", test_try_except_else_finally_no_error)
test("try_except_else_finally_with_error", test_try_except_else_finally_with_error)

print("CPython exception edge tests completed")
