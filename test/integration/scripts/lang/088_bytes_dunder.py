from test_framework import test, expect

# Test 1: Basic __bytes__ called by bytes()
def test_basic_bytes():
    class MyObj:
        def __bytes__(self):
            return b"hello"

    obj = MyObj()
    result = bytes(obj)
    expect(result).to_be(b"hello")

test("basic __bytes__ called by bytes()", test_basic_bytes)

# Test 2: __bytes__ returns empty bytes
def test_empty_bytes():
    class Empty:
        def __bytes__(self):
            return b""

    expect(bytes(Empty())).to_be(b"")

test("__bytes__ returns empty bytes", test_empty_bytes)

# Test 3: __bytes__ with instance state
def test_bytes_with_state():
    class Data:
        def __init__(self, values):
            self.values = values

        def __bytes__(self):
            return bytes(self.values)

    d = Data([72, 101, 108, 108, 111])
    expect(bytes(d)).to_be(b"Hello")

test("__bytes__ with instance state", test_bytes_with_state)

# Test 4: __bytes__ inherited from parent
def test_inherited_bytes():
    class Base:
        def __bytes__(self):
            return b"base"

    class Child(Base):
        pass

    expect(bytes(Child())).to_be(b"base")

test("__bytes__ inherited from parent", test_inherited_bytes)

# Test 5: Child overrides __bytes__
def test_override_bytes():
    class Base:
        def __bytes__(self):
            return b"base"

    class Child(Base):
        def __bytes__(self):
            return b"child"

    expect(bytes(Child())).to_be(b"child")

test("child overrides __bytes__", test_override_bytes)

# Test 6: __bytes__ returning non-bytes raises TypeError
def test_non_bytes_return():
    class Bad:
        def __bytes__(self):
            return "not bytes"

    try:
        bytes(Bad())
        expect(True).to_be(False)
    except TypeError:
        expect(True).to_be(True)

test("__bytes__ returning non-bytes raises TypeError", test_non_bytes_return)

# Test 7: Without __bytes__, falls back to iteration
def test_fallback_to_iter():
    class Iterable:
        def __iter__(self):
            return iter([65, 66, 67])

    result = bytes(Iterable())
    expect(result).to_be(b"ABC")

test("without __bytes__ falls back to iteration", test_fallback_to_iter)

# Test 8: __bytes__ takes priority over __iter__
def test_bytes_over_iter():
    class Both:
        def __bytes__(self):
            return b"bytes"

        def __iter__(self):
            return iter([1, 2, 3])

    expect(bytes(Both())).to_be(b"bytes")

test("__bytes__ takes priority over __iter__", test_bytes_over_iter)

# Test 9: __bytes__ with binary data
def test_binary_data():
    class BinData:
        def __bytes__(self):
            return bytes([0, 1, 255, 128])

    result = bytes(BinData())
    expect(len(result)).to_be(4)
    expect(result[0]).to_be(0)
    expect(result[1]).to_be(1)
    expect(result[2]).to_be(255)
    expect(result[3]).to_be(128)

test("__bytes__ with binary data", test_binary_data)

# Test 10: __bytes__ error propagates
def test_bytes_error_propagates():
    class Broken:
        def __bytes__(self):
            raise ValueError("broken")

    try:
        bytes(Broken())
        expect(True).to_be(False)
    except ValueError as e:
        expect(str(e)).to_be("broken")

test("__bytes__ error propagates", test_bytes_error_propagates)
