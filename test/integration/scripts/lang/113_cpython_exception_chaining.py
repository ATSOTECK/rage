# Test: CPython Exception Chaining
# Adapted from CPython's test_exceptions.py
# Tests __cause__, __context__, __suppress_context__, raise X from Y

from test_framework import test, expect

# === New exceptions have None for chaining attributes ===
def test_chaining_attrs_defaults():
    try:
        raise Exception("hello")
    except Exception as e:
        expect(e.__cause__).to_be(None)
        expect(e.__context__).to_be(None)
        expect(e.__suppress_context__).to_be(False)

test("chaining_attrs_defaults", test_chaining_attrs_defaults)

# === raise X from Y sets __cause__ and __suppress_context__ ===
def test_explicit_cause_basic():
    try:
        try:
            raise ValueError("original")
        except ValueError as orig:
            raise TypeError("wrapped") from orig
    except TypeError as e:
        expect(e.__cause__.__class__.__name__).to_be("ValueError")
        expect(str(e.__cause__)).to_be("original")
        expect(e.__suppress_context__).to_be(True)

test("explicit_cause_basic", test_explicit_cause_basic)

# === raise X from Y with exception instance ===
def test_explicit_cause_from_instance():
    try:
        raise TypeError("wrapped") from ValueError("the cause")
    except TypeError as e:
        expect(e.__cause__.__class__.__name__).to_be("ValueError")
        expect(str(e.__cause__)).to_be("the cause")
        expect(e.__suppress_context__).to_be(True)

test("explicit_cause_from_instance", test_explicit_cause_from_instance)

# === raise X from None suppresses context ===
def test_raise_from_none():
    try:
        try:
            raise ValueError("original")
        except ValueError:
            raise TypeError("replacement") from None
    except TypeError as e:
        expect(e.__cause__).to_be(None)
        expect(e.__suppress_context__).to_be(True)

test("raise_from_none", test_raise_from_none)

# === Implicit __context__ set when raising in except block ===
def test_implicit_context_basic():
    try:
        try:
            raise ValueError("first")
        except ValueError:
            raise TypeError("second")
    except TypeError as e:
        expect(e.__context__.__class__.__name__).to_be("ValueError")
        expect(str(e.__context__)).to_be("first")
        expect(e.__suppress_context__).to_be(False)

test("implicit_context_basic", test_implicit_context_basic)

# === __context__ is not set when not in except block ===
def test_no_context_outside_except():
    try:
        raise ValueError("standalone")
    except ValueError as e:
        expect(e.__context__).to_be(None)

test("no_context_outside_except", test_no_context_outside_except)

# === raise X from Y also sets __context__ when in except block ===
def test_explicit_cause_also_sets_context():
    try:
        try:
            raise ValueError("original")
        except ValueError:
            raise TypeError("new") from RuntimeError("the cause")
    except TypeError as e:
        # __cause__ is the explicit one
        expect(e.__cause__.__class__.__name__).to_be("RuntimeError")
        # __context__ is the implicitly active exception
        expect(e.__context__.__class__.__name__).to_be("ValueError")
        expect(e.__suppress_context__).to_be(True)

test("explicit_cause_also_sets_context", test_explicit_cause_also_sets_context)

# === Context threading through nested except blocks ===
def test_context_chain_depth():
    try:
        try:
            try:
                raise ValueError("first")
            except ValueError:
                raise TypeError("second")
        except TypeError:
            raise RuntimeError("third")
    except RuntimeError as e:
        # third.__context__ is second
        expect(e.__context__.__class__.__name__).to_be("TypeError")
        # second.__context__ is first
        expect(e.__context__.__context__.__class__.__name__).to_be("ValueError")

test("context_chain_depth", test_context_chain_depth)

# === Context in finally block ===
def test_context_in_finally():
    try:
        try:
            raise ValueError("in try")
        finally:
            raise TypeError("in finally")
    except TypeError as e:
        expect(e.__context__.__class__.__name__).to_be("ValueError")
        expect(str(e.__context__)).to_be("in try")

test("context_in_finally", test_context_in_finally)

# === Exception in else clause has no implicit context ===
def test_context_in_else():
    try:
        try:
            pass
        except ValueError:
            pass
        else:
            raise TypeError("in else")
    except TypeError as e:
        # No active exception in else clause
        expect(e.__context__).to_be(None)

test("context_in_else", test_context_in_else)

# === raise from in except block: cause and context are set ===
def test_raise_from_in_except():
    try:
        try:
            raise ValueError("original")
        except ValueError:
            raise TypeError("new") from RuntimeError("the cause")
    except TypeError as e:
        expect(e.__cause__.__class__.__name__).to_be("RuntimeError")
        expect(str(e.__cause__)).to_be("the cause")
        expect(e.__context__.__class__.__name__).to_be("ValueError")

test("raise_from_in_except", test_raise_from_in_except)

# === Setting __cause__ programmatically sets __suppress_context__ ===
def test_setting_cause_sets_suppress():
    try:
        raise TypeError("err")
    except TypeError as e:
        expect(e.__suppress_context__).to_be(False)
        e.__cause__ = ValueError("cause")
        expect(e.__suppress_context__).to_be(True)

test("setting_cause_sets_suppress", test_setting_cause_sets_suppress)

# === Setting __cause__ to None still sets __suppress_context__ ===
def test_setting_cause_none_sets_suppress():
    try:
        raise TypeError("err")
    except TypeError as e:
        e.__cause__ = None
        expect(e.__suppress_context__).to_be(True)

test("setting_cause_none_sets_suppress", test_setting_cause_none_sets_suppress)

# === __suppress_context__ can be manually reset ===
def test_suppress_context_manual_reset():
    try:
        raise TypeError("err")
    except TypeError as e:
        e.__cause__ = ValueError("cause")
        expect(e.__suppress_context__).to_be(True)
        e.__suppress_context__ = False
        expect(e.__suppress_context__).to_be(False)

test("suppress_context_manual_reset", test_suppress_context_manual_reset)

# === Context set when raising a different exception in except ===
def test_context_with_different_raise():
    try:
        try:
            raise ValueError("original")
        except ValueError:
            raise TypeError("new")
    except TypeError as e:
        expect(str(e.__context__)).to_be("original")
        expect(e.__context__.__class__.__name__).to_be("ValueError")

test("context_with_different_raise", test_context_with_different_raise)

# === Bare raise preserves the exception (no new context) ===
def test_bare_raise_no_new_context():
    try:
        try:
            raise ValueError("original")
        except ValueError:
            raise  # bare re-raise
    except ValueError as e:
        expect(str(e)).to_be("original")

test("bare_raise_no_new_context", test_bare_raise_no_new_context)

# === Custom exceptions support chaining ===
def test_custom_exception_chaining():
    class AppError(Exception):
        pass
    class DBError(AppError):
        pass
    try:
        try:
            raise DBError("db failure")
        except DBError:
            raise AppError("app failure")
    except AppError as e:
        expect(e.__context__.__class__.__name__).to_be("DBError")

test("custom_exception_chaining", test_custom_exception_chaining)

# === Bare raise does not alter __context__ of the original exception ===
def test_bare_raise_preserves_original_context():
    try:
        try:
            raise ValueError("v")
        except ValueError:
            try:
                raise TypeError("t")
            except TypeError:
                pass
            raise  # re-raises ValueError
    except ValueError as e:
        # Bare re-raise should not add context
        expect(e.__context__).to_be(None)

test("bare_raise_preserves_original_context", test_bare_raise_preserves_original_context)

# === Multiple except clauses with different contexts ===
def test_multiple_except_different_context():
    results = []
    try:
        try:
            raise ValueError("v")
        except ValueError:
            try:
                raise TypeError("t")
            except TypeError as te:
                results.append(te.__context__.__class__.__name__)
            raise KeyError("k")
    except KeyError as ke:
        results.append(ke.__context__.__class__.__name__)
    expect(results).to_be(["ValueError", "ValueError"])

test("multiple_except_different_context", test_multiple_except_different_context)

# === raise from with pre-existing exception instance ===
def test_raise_from_existing_instance():
    try:
        raise ValueError("the effect") from RuntimeError("the cause")
    except ValueError as e:
        expect(str(e.__cause__)).to_be("the cause")
        expect(e.__cause__.__class__.__name__).to_be("RuntimeError")
        expect(e.__suppress_context__).to_be(True)

test("raise_from_existing_instance", test_raise_from_existing_instance)

# === __class__ accessible on caught exceptions ===
def test_exception_class_access():
    try:
        raise ValueError("test")
    except ValueError as e:
        expect(e.__class__.__name__).to_be("ValueError")

test("exception_class_access", test_exception_class_access)

# === Chained cause is itself a proper exception with attributes ===
def test_cause_has_attributes():
    try:
        try:
            raise ValueError("inner")
        except ValueError as inner:
            raise TypeError("outer") from inner
    except TypeError as e:
        expect(str(e)).to_be("outer")
        expect(str(e.__cause__)).to_be("inner")
        expect(e.__cause__.__cause__).to_be(None)
        expect(e.__cause__.__context__).to_be(None)

test("cause_has_attributes", test_cause_has_attributes)

print("CPython exception chaining tests completed")
