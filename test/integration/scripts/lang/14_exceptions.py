# Test: Exceptions
# Tests try/except/finally, raise, exception types, inheritance

from test_framework import test, expect

# Helper functions at module level
def raise_in_func():
    raise ValueError

def inner_raise():
    raise KeyError

def outer_call():
    inner_raise()

gen_finally_ran = False
def gen_with_finally():
    global gen_finally_ran
    try:
        yield 1
        yield 2
    finally:
        gen_finally_ran = True

def gen_catches_throw():
    try:
        yield 1
        yield 2
    except ValueError:
        yield "caught"
    yield 3

def gen_no_catch():
    yield 1
    yield 2
    yield 3

def gen_internal_except():
    try:
        yield 1
        raise ValueError
    except ValueError:
        yield "internal_caught"
    yield 3

def simple_gen():
    yield 1

def test_basic_try_except():
    result = "not caught"
    try:
        raise ValueError
    except ValueError:
        result = "caught"
    expect(result).to_be("caught")

def test_exception_as_binding():
    caught = False
    try:
        raise ValueError
    except ValueError as e:
        caught = True
    expect(caught).to_be(True)

def test_multiple_except():
    # KeyError case
    result = ""
    try:
        raise KeyError
    except ValueError:
        result = "value"
    except KeyError:
        result = "key"
    except TypeError:
        result = "type"
    expect(result).to_be("key")

    # TypeError case
    result2 = ""
    try:
        raise TypeError
    except ValueError:
        result2 = "value"
    except KeyError:
        result2 = "key"
    except TypeError:
        result2 = "type"
    expect(result2).to_be("type")

def test_bare_except():
    result = ""
    try:
        raise RuntimeError
    except:
        result = "caught"
    expect(result).to_be("caught")

def test_finally_no_exception():
    finally_ran = False
    try:
        x = 1
    finally:
        finally_ran = True
    expect(finally_ran).to_be(True)

def test_finally_with_caught_exception():
    finally_ran = False
    caught = False
    try:
        raise ValueError
    except ValueError:
        caught = True
    finally:
        finally_ran = True
    expect(finally_ran).to_be(True)
    expect(caught).to_be(True)

def test_finally_propagates():
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
    expect(inner_finally_ran).to_be(True)
    expect(outer_finally_ran).to_be(True)
    expect(outer_caught).to_be(True)

def test_else_clause():
    # Runs when no exception
    else_ran = False
    try:
        x = 1
    except ValueError:
        pass
    else:
        else_ran = True
    expect(else_ran).to_be(True)

    # Doesn't run when exception
    else_ran2 = False
    try:
        raise ValueError
    except ValueError:
        pass
    else:
        else_ran2 = True
    expect(else_ran2).to_be(False)

def test_reraise():
    caught_outer = False
    try:
        try:
            raise ValueError
        except ValueError:
            raise
    except ValueError:
        caught_outer = True
    expect(caught_outer).to_be(True)

def test_exception_inheritance():
    # ValueError is Exception
    caught1 = False
    try:
        raise ValueError
    except Exception:
        caught1 = True
    expect(caught1).to_be(True)

    # KeyError is Exception
    caught2 = False
    try:
        raise KeyError
    except Exception:
        caught2 = True
    expect(caught2).to_be(True)

    # IndexError is Exception
    caught3 = False
    try:
        raise IndexError
    except Exception:
        caught3 = True
    expect(caught3).to_be(True)

    # ZeroDivisionError is Exception
    caught4 = False
    try:
        raise ZeroDivisionError
    except Exception:
        caught4 = True
    expect(caught4).to_be(True)

def test_nested_try():
    inner_caught = False
    outer_caught = False
    try:
        try:
            raise KeyError
        except ValueError:
            inner_caught = True
    except KeyError:
        outer_caught = True
    expect(inner_caught).to_be(False)
    expect(outer_caught).to_be(True)

def test_full_try_except_else_finally():
    # No exception case
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
    expect(caught).to_be(False)
    expect(else_ran).to_be(True)
    expect(finally_ran).to_be(True)

    # With exception case
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
    expect(caught2).to_be(True)
    expect(else_ran2).to_be(False)
    expect(finally_ran2).to_be(True)

def test_tuple_except():
    # Catching ValueError
    result1 = ""
    try:
        raise ValueError
    except (ValueError, KeyError):
        result1 = "caught"
    expect(result1).to_be("caught")

    # Catching KeyError
    result2 = ""
    try:
        raise KeyError
    except (ValueError, KeyError):
        result2 = "caught"
    expect(result2).to_be("caught")

def test_exception_classes():
    count = 0

    try:
        raise ValueError
    except ValueError:
        count = count + 1

    try:
        raise TypeError
    except TypeError:
        count = count + 1

    try:
        raise KeyError
    except KeyError:
        count = count + 1

    try:
        raise RuntimeError
    except RuntimeError:
        count = count + 1

    try:
        raise StopIteration
    except StopIteration:
        count = count + 1

    expect(count).to_be(5)

def test_deeply_nested():
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
    expect(depth_reached).to_be(3)

def test_var_preserved():
    x = 10
    try:
        x = 20
        raise ValueError
    except ValueError:
        pass
    expect(x).to_be(20)

def test_finally_after_handler_raises():
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
    expect(finally_after_reraise).to_be(True)
    expect(outer_finally).to_be(True)

def test_func_exception():
    func_exc_caught = False
    try:
        raise_in_func()
    except ValueError:
        func_exc_caught = True
    expect(func_exc_caught).to_be(True)

def test_nested_func_exception():
    nested_func_exc_caught = False
    try:
        outer_call()
    except KeyError:
        nested_func_exc_caught = True
    expect(nested_func_exc_caught).to_be(True)

def test_generator_throw_caught():
    g = gen_catches_throw()
    gen_throw_results = []
    for v in g:
        gen_throw_results.append(v)
        if v == 1:
            gen_throw_results.append(g.throw(ValueError, "test"))
            break
    expect(gen_throw_results[0]).to_be(1)
    expect(gen_throw_results[1]).to_be("caught")

def test_generator_throw_propagates():
    g2 = gen_no_catch()
    gen_throw_propagated = False
    for v in g2:
        if v == 1:
            try:
                g2.throw(RuntimeError, "uncaught")
            except RuntimeError:
                gen_throw_propagated = True
            break
    expect(gen_throw_propagated).to_be(True)

def test_generator_close_finally():
    global gen_finally_ran
    gen_finally_ran = False
    g3 = gen_with_finally()
    for v in g3:
        break
    g3.close()
    expect(gen_finally_ran).to_be(True)

def test_generator_internal_except():
    gen_internal_results = []
    for v in gen_internal_except():
        gen_internal_results.append(v)
    expect(gen_internal_results[0]).to_be(1)
    expect(gen_internal_results[1]).to_be("internal_caught")
    expect(gen_internal_results[2]).to_be(3)

def test_throw_into_closed_gen():
    g4 = simple_gen()
    for v in g4:
        pass  # Exhaust generator

    throw_into_closed_raised = False
    try:
        g4.throw(ValueError, "to closed")
    except:
        throw_into_closed_raised = True
    expect(throw_into_closed_raised).to_be(True)

test("basic_try_except", test_basic_try_except)
test("exception_as_binding", test_exception_as_binding)
test("multiple_except", test_multiple_except)
test("bare_except", test_bare_except)
test("finally_no_exception", test_finally_no_exception)
test("finally_with_caught_exception", test_finally_with_caught_exception)
test("finally_propagates", test_finally_propagates)
test("else_clause", test_else_clause)
test("reraise", test_reraise)
test("exception_inheritance", test_exception_inheritance)
test("nested_try", test_nested_try)
test("full_try_except_else_finally", test_full_try_except_else_finally)
test("tuple_except", test_tuple_except)
test("exception_classes", test_exception_classes)
test("deeply_nested", test_deeply_nested)
test("var_preserved", test_var_preserved)
test("finally_after_handler_raises", test_finally_after_handler_raises)
test("func_exception", test_func_exception)
test("nested_func_exception", test_nested_func_exception)
test("generator_throw_caught", test_generator_throw_caught)
test("generator_throw_propagates", test_generator_throw_propagates)
test("generator_close_finally", test_generator_close_finally)
test("generator_internal_except", test_generator_internal_except)
test("throw_into_closed_gen", test_throw_into_closed_gen)

print("Exceptions tests completed")
