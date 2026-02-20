from test_framework import test, expect

# Test e.args on single-arg exception
def test_single_arg():
    try:
        raise ValueError("something went wrong")
    except ValueError as e:
        expect(e.args).to_equal(("something went wrong",))
        expect(e.args[0]).to_equal("something went wrong")
        expect(len(e.args)).to_equal(1)

test("single arg exception - e.args", test_single_arg)

# Test e.args on multi-arg exception
def test_multi_arg():
    try:
        raise ValueError("error", 42, "extra")
    except ValueError as e:
        expect(e.args).to_equal(("error", 42, "extra"))
        expect(e.args[0]).to_equal("error")
        expect(e.args[1]).to_equal(42)
        expect(len(e.args)).to_equal(3)

test("multi arg exception - e.args", test_multi_arg)

# Test e.args on zero-arg exception
def test_zero_arg():
    try:
        raise ValueError
    except ValueError as e:
        expect(e.args).to_equal(())
        expect(len(e.args)).to_equal(0)

test("zero arg exception - e.args", test_zero_arg)

# Test e.args on different exception types
def test_type_error():
    try:
        raise TypeError("bad type")
    except TypeError as e:
        expect(e.args).to_equal(("bad type",))

test("TypeError - e.args", test_type_error)

def test_key_error():
    try:
        raise KeyError("missing_key")
    except KeyError as e:
        expect(e.args[0]).to_equal("missing_key")

test("KeyError - e.args[0]", test_key_error)

def test_runtime_error():
    try:
        raise RuntimeError("runtime issue")
    except RuntimeError as e:
        expect(e.args).to_equal(("runtime issue",))

test("RuntimeError - e.args", test_runtime_error)

# Test __cause__ with raise X from Y
def test_cause_chained():
    try:
        try:
            raise ValueError("original")
        except ValueError as orig:
            raise TypeError("wrapper") from orig
    except TypeError as e:
        expect(e.__cause__.args).to_equal(("original",))

test("__cause__ with raise from", test_cause_chained)

# Test __cause__ is None when not chained
def test_cause_none():
    try:
        raise ValueError("standalone")
    except ValueError as e:
        expect(e.__cause__).to_equal(None)

test("__cause__ is None when not chained", test_cause_none)

# Test __traceback__ returns None
def test_traceback():
    try:
        raise ValueError("tb test")
    except ValueError as e:
        expect(e.__traceback__).to_equal(None)

test("__traceback__ returns None", test_traceback)

# Test args indexing and iteration
def test_args_iteration():
    try:
        raise ValueError("a", "b", "c")
    except ValueError as e:
        items = []
        for arg in e.args:
            items.append(arg)
        expect(items).to_equal(["a", "b", "c"])

test("e.args iteration", test_args_iteration)
