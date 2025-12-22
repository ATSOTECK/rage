# Test: Exceptions
# Tests try/except/finally, raise, exception types, inheritance

results = {}

# =====================================
# Basic try/except
# =====================================

result = "not caught"
try:
    raise ValueError
except ValueError:
    result = "caught"
results["basic_try_except"] = result

# =====================================
# Exception with 'as' binding
# =====================================

try:
    raise ValueError
except ValueError as e:
    results["exception_as_binding"] = True

# =====================================
# Multiple except clauses
# =====================================

result = ""
try:
    raise KeyError
except ValueError:
    result = "value"
except KeyError:
    result = "key"
except TypeError:
    result = "type"
results["multiple_except_catches_correct"] = result

# Try with TypeError
result2 = ""
try:
    raise TypeError
except ValueError:
    result2 = "value"
except KeyError:
    result2 = "key"
except TypeError:
    result2 = "type"
results["multiple_except_type"] = result2

# =====================================
# Bare except clause
# =====================================

result = ""
try:
    raise RuntimeError
except:
    result = "caught"
results["bare_except"] = result

# =====================================
# Finally block - no exception
# =====================================

finally_ran = False
try:
    x = 1
finally:
    finally_ran = True
results["finally_no_exception"] = finally_ran

# =====================================
# Finally block - with caught exception
# =====================================

finally_ran = False
caught = False
try:
    raise ValueError
except ValueError:
    caught = True
finally:
    finally_ran = True
results["finally_with_caught_exception"] = finally_ran
results["exception_was_caught"] = caught

# =====================================
# Finally block - exception propagates through
# =====================================

outer_finally_ran = False
inner_finally_ran = False
outer_caught = False
try:
    try:
        raise ValueError
    finally:
        inner_finally_ran = True
except ValueError:
    outer_caught = True
finally:
    outer_finally_ran = True
results["inner_finally_ran"] = inner_finally_ran
results["outer_finally_ran"] = outer_finally_ran
results["outer_caught_propagated"] = outer_caught

# =====================================
# Else clause - runs when no exception
# =====================================

else_ran = False
try:
    x = 1
except ValueError:
    pass
else:
    else_ran = True
results["else_runs_no_exception"] = else_ran

# =====================================
# Else clause - doesn't run when exception
# =====================================

else_ran = False
try:
    raise ValueError
except ValueError:
    pass
else:
    else_ran = True
results["else_skipped_with_exception"] = else_ran

# =====================================
# Re-raise with bare raise
# =====================================

caught_outer = False
try:
    try:
        raise ValueError
    except ValueError:
        raise
except ValueError:
    caught_outer = True
results["reraise_caught_outer"] = caught_outer

# =====================================
# Exception inheritance - ValueError is Exception
# =====================================

caught = False
try:
    raise ValueError
except Exception:
    caught = True
results["valueerror_is_exception"] = caught

# =====================================
# Exception inheritance - KeyError is Exception
# =====================================

caught = False
try:
    raise KeyError
except Exception:
    caught = True
results["keyerror_is_exception"] = caught

# =====================================
# Exception inheritance - IndexError is Exception
# =====================================

caught = False
try:
    raise IndexError
except Exception:
    caught = True
results["indexerror_is_exception"] = caught

# =====================================
# Exception inheritance - ZeroDivisionError is Exception
# =====================================

caught = False
try:
    raise ZeroDivisionError
except Exception:
    caught = True
results["zerodiv_is_exception"] = caught

# =====================================
# Nested try/except - inner doesn't catch
# =====================================

inner_caught = False
outer_caught = False
try:
    try:
        raise KeyError
    except ValueError:
        inner_caught = True
except KeyError:
    outer_caught = True
results["nested_inner_missed"] = inner_caught
results["nested_outer_caught"] = outer_caught

# =====================================
# Try/except/else/finally all together
# =====================================

caught = False
else_ran = False
finally_ran = False
try:
    x = 1
except ValueError:
    caught = True
else:
    else_ran = True
finally:
    finally_ran = True
results["full_try_caught"] = caught
results["full_try_else"] = else_ran
results["full_try_finally"] = finally_ran

# =====================================
# Try/except/else/finally with exception
# =====================================

caught2 = False
else_ran2 = False
finally_ran2 = False
try:
    raise ValueError
except ValueError:
    caught2 = True
else:
    else_ran2 = True
finally:
    finally_ran2 = True
results["full_try_exc_caught"] = caught2
results["full_try_exc_else"] = else_ran2
results["full_try_exc_finally"] = finally_ran2

# =====================================
# Multiple exception types in one except
# =====================================

# Test catching ValueError
result1 = ""
try:
    raise ValueError
except (ValueError, KeyError):
    result1 = "caught"
results["tuple_except_value"] = result1

# Test catching KeyError
result2 = ""
try:
    raise KeyError
except (ValueError, KeyError):
    result2 = "caught"
results["tuple_except_key"] = result2

# =====================================
# Standard exception classes exist
# =====================================

# Test a subset of exception classes to verify the hierarchy exists
exception_classes_count = 0

try:
    raise ValueError
except ValueError:
    exception_classes_count = exception_classes_count + 1

try:
    raise TypeError
except TypeError:
    exception_classes_count = exception_classes_count + 1

try:
    raise KeyError
except KeyError:
    exception_classes_count = exception_classes_count + 1

try:
    raise RuntimeError
except RuntimeError:
    exception_classes_count = exception_classes_count + 1

try:
    raise StopIteration
except StopIteration:
    exception_classes_count = exception_classes_count + 1

results["exception_classes_count"] = exception_classes_count

# =====================================
# Deeply nested try blocks
# =====================================

depth_reached = 0
try:
    depth_reached = 1
    try:
        depth_reached = 2
        try:
            depth_reached = 3
            raise ValueError
        except TypeError:
            depth_reached = -1
    except KeyError:
        depth_reached = -2
except ValueError:
    pass
results["deeply_nested_depth"] = depth_reached

# =====================================
# Exception doesn't affect outer scope vars
# =====================================

x = 10
try:
    x = 20
    raise ValueError
except ValueError:
    pass
results["var_preserved_after_exception"] = x

# =====================================
# Finally runs even when except raises
# =====================================

finally_after_reraise = False
outer_finally = False
try:
    try:
        raise ValueError
    except ValueError:
        raise KeyError
    finally:
        finally_after_reraise = True
except KeyError:
    pass
finally:
    outer_finally = True
results["finally_after_handler_raises"] = finally_after_reraise
results["outer_finally_after_handler_raises"] = outer_finally

print("Exceptions tests completed")
