from test_framework import test, expect

# Test 1: LookupError catches KeyError
def test_lookup_catches_key():
    try:
        raise KeyError("k")
    except LookupError as e:
        return True
    return False

test("LookupError catches KeyError", lambda: expect(test_lookup_catches_key()).to_be(True))

# Test 2: LookupError catches IndexError
def test_lookup_catches_index():
    try:
        raise IndexError("i")
    except LookupError as e:
        return True
    return False

test("LookupError catches IndexError", lambda: expect(test_lookup_catches_index()).to_be(True))

# Test 3: ArithmeticError catches ZeroDivisionError
def test_arith_catches_zerodiv():
    try:
        1 / 0
    except ArithmeticError:
        return True
    return False

test("ArithmeticError catches ZeroDivisionError", lambda: expect(test_arith_catches_zerodiv()).to_be(True))

# Test 4: ArithmeticError catches OverflowError
def test_arith_catches_overflow():
    try:
        raise OverflowError("too big")
    except ArithmeticError:
        return True
    return False

test("ArithmeticError catches OverflowError", lambda: expect(test_arith_catches_overflow()).to_be(True))

# Test 5: ArithmeticError catches FloatingPointError
def test_arith_catches_fpe():
    try:
        raise FloatingPointError("fpe")
    except ArithmeticError:
        return True
    return False

test("ArithmeticError catches FloatingPointError", lambda: expect(test_arith_catches_fpe()).to_be(True))

# Test 6: OSError catches TimeoutError
def test_os_catches_timeout():
    try:
        raise TimeoutError("timed out")
    except OSError:
        return True
    return False

test("OSError catches TimeoutError", lambda: expect(test_os_catches_timeout()).to_be(True))

# Test 7: OSError catches ConnectionError
def test_os_catches_connection():
    try:
        raise ConnectionError("conn failed")
    except OSError:
        return True
    return False

test("OSError catches ConnectionError", lambda: expect(test_os_catches_connection()).to_be(True))

# Test 8: ConnectionError catches ConnectionRefusedError
def test_conn_catches_refused():
    try:
        raise ConnectionRefusedError("refused")
    except ConnectionError:
        return True
    return False

test("ConnectionError catches ConnectionRefusedError", lambda: expect(test_conn_catches_refused()).to_be(True))

# Test 9: ConnectionError catches ConnectionResetError
def test_conn_catches_reset():
    try:
        raise ConnectionResetError("reset")
    except ConnectionError:
        return True
    return False

test("ConnectionError catches ConnectionResetError", lambda: expect(test_conn_catches_reset()).to_be(True))

# Test 10: OSError catches ConnectionRefusedError (grandchild)
def test_os_catches_conn_refused():
    try:
        raise ConnectionRefusedError("refused")
    except OSError:
        return True
    return False

test("OSError catches ConnectionRefusedError (grandchild)", lambda: expect(test_os_catches_conn_refused()).to_be(True))

# Test 11: ValueError catches UnicodeError
def test_value_catches_unicode():
    try:
        raise UnicodeError("bad unicode")
    except ValueError:
        return True
    return False

test("ValueError catches UnicodeError", lambda: expect(test_value_catches_unicode()).to_be(True))

# Test 12: UnicodeError catches UnicodeDecodeError
def test_unicode_catches_decode():
    try:
        raise UnicodeDecodeError("decode fail")
    except UnicodeError:
        return True
    return False

test("UnicodeError catches UnicodeDecodeError", lambda: expect(test_unicode_catches_decode()).to_be(True))

# Test 13: ValueError catches UnicodeDecodeError (grandchild)
def test_value_catches_decode():
    try:
        raise UnicodeDecodeError("decode fail")
    except ValueError:
        return True
    return False

test("ValueError catches UnicodeDecodeError (grandchild)", lambda: expect(test_value_catches_decode()).to_be(True))

# Test 14: Warning catches DeprecationWarning
def test_warning_catches_deprecation():
    try:
        raise DeprecationWarning("old")
    except Warning:
        return True
    return False

test("Warning catches DeprecationWarning", lambda: expect(test_warning_catches_deprecation()).to_be(True))

# Test 15: Warning catches RuntimeWarning
def test_warning_catches_runtime():
    try:
        raise RuntimeWarning("runtime")
    except Warning:
        return True
    return False

test("Warning catches RuntimeWarning", lambda: expect(test_warning_catches_runtime()).to_be(True))

# Test 16: Warning catches UserWarning
def test_warning_catches_user():
    try:
        raise UserWarning("user")
    except Warning:
        return True
    return False

test("Warning catches UserWarning", lambda: expect(test_warning_catches_user()).to_be(True))

# Test 17: Exception catches Warning (Warning inherits from Exception)
def test_exception_catches_warning():
    try:
        raise FutureWarning("future")
    except Exception:
        return True
    return False

test("Exception catches Warning subclass", lambda: expect(test_exception_catches_warning()).to_be(True))

# Test 18: BaseException catches everything
def test_base_catches_all():
    count = 0
    try:
        raise KeyError("test")
    except BaseException:
        count = count + 1
    try:
        raise IndexError("test")
    except BaseException:
        count = count + 1
    try:
        raise ZeroDivisionError("test")
    except BaseException:
        count = count + 1
    try:
        raise TimeoutError("test")
    except BaseException:
        count = count + 1
    try:
        raise UnicodeError("test")
    except BaseException:
        count = count + 1
    try:
        raise DeprecationWarning("test")
    except BaseException:
        count = count + 1
    try:
        raise SyntaxError("test")
    except BaseException:
        count = count + 1
    return count

test("BaseException catches all exception types", lambda: expect(test_base_catches_all()).to_be(7))

# Test 19: isinstance checks for hierarchy
def test_isinstance_hierarchy():
    results = []
    results.append(isinstance(KeyError("k"), LookupError))
    results.append(isinstance(IndexError("i"), LookupError))
    results.append(isinstance(ZeroDivisionError("z"), ArithmeticError))
    results.append(isinstance(OverflowError("o"), ArithmeticError))
    results.append(isinstance(TimeoutError("t"), OSError))
    results.append(isinstance(ConnectionRefusedError("r"), ConnectionError))
    results.append(isinstance(ConnectionRefusedError("r"), OSError))
    results.append(isinstance(UnicodeDecodeError("d"), UnicodeError))
    results.append(isinstance(UnicodeDecodeError("d"), ValueError))
    results.append(isinstance(DeprecationWarning("d"), Warning))
    return all(results)

test("isinstance respects exception hierarchy", lambda: expect(test_isinstance_hierarchy()).to_be(True))

# Test 20: issubclass checks
def test_issubclass_hierarchy():
    results = []
    results.append(issubclass(KeyError, LookupError))
    results.append(issubclass(IndexError, LookupError))
    results.append(issubclass(ZeroDivisionError, ArithmeticError))
    results.append(issubclass(OverflowError, ArithmeticError))
    results.append(issubclass(FloatingPointError, ArithmeticError))
    results.append(issubclass(TimeoutError, OSError))
    results.append(issubclass(ConnectionError, OSError))
    results.append(issubclass(ConnectionRefusedError, ConnectionError))
    results.append(issubclass(ConnectionResetError, ConnectionError))
    results.append(issubclass(BrokenPipeError, ConnectionError))
    results.append(issubclass(UnicodeError, ValueError))
    results.append(issubclass(UnicodeDecodeError, UnicodeError))
    results.append(issubclass(DeprecationWarning, Warning))
    results.append(issubclass(RuntimeWarning, Warning))
    results.append(issubclass(Warning, Exception))
    results.append(issubclass(SyntaxError, Exception))
    results.append(issubclass(EOFError, Exception))
    results.append(issubclass(StopAsyncIteration, BaseException))
    return all(results)

test("issubclass respects exception hierarchy", lambda: expect(test_issubclass_hierarchy()).to_be(True))

# Test 21: StopAsyncIteration is BaseException but NOT Exception
def test_stop_async_not_exception():
    try:
        raise StopAsyncIteration()
    except Exception:
        return False
    except BaseException:
        return True
    return False

test("StopAsyncIteration is BaseException, not Exception", lambda: expect(test_stop_async_not_exception()).to_be(True))

# Test 22: SyntaxError can be raised and caught
def test_syntax_error():
    try:
        raise SyntaxError("invalid syntax")
    except SyntaxError as e:
        return True
    return False

test("SyntaxError can be raised and caught", lambda: expect(test_syntax_error()).to_be(True))

# Test 23: Multiple except clauses with hierarchy
def test_specific_before_general():
    try:
        raise KeyError("k")
    except KeyError:
        return "KeyError"
    except LookupError:
        return "LookupError"
    except Exception:
        return "Exception"

test("specific except matched before general", lambda: expect(test_specific_before_general()).to_be("KeyError"))

# Test 24: OSError subclass names
def test_os_subclasses():
    count = 0
    try:
        raise IsADirectoryError("test")
    except OSError:
        count = count + 1
    try:
        raise NotADirectoryError("test")
    except OSError:
        count = count + 1
    try:
        raise InterruptedError("test")
    except OSError:
        count = count + 1
    try:
        raise BlockingIOError("test")
    except OSError:
        count = count + 1
    try:
        raise ChildProcessError("test")
    except OSError:
        count = count + 1
    try:
        raise ProcessLookupError("test")
    except OSError:
        count = count + 1
    return count

test("all OSError subclasses caught by OSError", lambda: expect(test_os_subclasses()).to_be(6))

# Test 25: Warning subclass names
def test_warning_subclasses():
    count = 0
    try:
        raise DeprecationWarning("test")
    except Warning:
        count = count + 1
    try:
        raise PendingDeprecationWarning("test")
    except Warning:
        count = count + 1
    try:
        raise RuntimeWarning("test")
    except Warning:
        count = count + 1
    try:
        raise SyntaxWarning("test")
    except Warning:
        count = count + 1
    try:
        raise UserWarning("test")
    except Warning:
        count = count + 1
    try:
        raise FutureWarning("test")
    except Warning:
        count = count + 1
    try:
        raise ImportWarning("test")
    except Warning:
        count = count + 1
    try:
        raise UnicodeWarning("test")
    except Warning:
        count = count + 1
    try:
        raise BytesWarning("test")
    except Warning:
        count = count + 1
    try:
        raise ResourceWarning("test")
    except Warning:
        count = count + 1
    try:
        raise EncodingWarning("test")
    except Warning:
        count = count + 1
    return count

test("all Warning subclasses caught by Warning", lambda: expect(test_warning_subclasses()).to_be(11))
