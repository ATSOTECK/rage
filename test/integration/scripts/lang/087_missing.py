from test_framework import test, expect

# Test 1: Basic __missing__ called on failed key lookup
def test_basic_missing():
    class MyDict:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            return "default"

        def __setitem__(self, key, value):
            self._data[key] = value

    d = MyDict()
    d["a"] = 10
    expect(d["a"]).to_be(10)
    expect(d["b"]).to_be("default")

test("basic __missing__ on failed lookup", test_basic_missing)

# Test 2: __missing__ receives the key
def test_missing_receives_key():
    class KeyTracker:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            return key

    d = KeyTracker()
    expect(d["hello"]).to_be("hello")
    expect(d[42]).to_be(42)

test("__missing__ receives the key", test_missing_receives_key)

# Test 3: __missing__ not called when key exists
def test_missing_not_called_when_found():
    calls = []

    class MyDict:
        def __init__(self):
            self._data = {"x": 1}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            calls.append(key)
            return None

    d = MyDict()
    d["x"]
    expect(calls).to_be([])

test("__missing__ not called when key exists", test_missing_not_called_when_found)

# Test 4: __missing__ inherited from parent
def test_inherited_missing():
    class Base:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            return "from_base"

    class Child(Base):
        pass

    d = Child()
    expect(d["anything"]).to_be("from_base")

test("__missing__ inherited from parent", test_inherited_missing)

# Test 5: Child overrides __missing__
def test_override_missing():
    class Base:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            return "base"

    class Child(Base):
        def __missing__(self, key):
            return "child"

    d = Child()
    expect(d["x"]).to_be("child")

test("child overrides __missing__", test_override_missing)

# Test 6: __missing__ can raise its own error
def test_missing_raises():
    class StrictDict:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            raise ValueError("no such key: " + str(key))

    d = StrictDict()
    try:
        d["nope"]
        expect(True).to_be(False)
    except ValueError as e:
        expect(str(e)).to_be("no such key: nope")

test("__missing__ can raise its own error", test_missing_raises)

# Test 7: Without __missing__, KeyError propagates normally
def test_no_missing_keyerror():
    class PlainDict:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

    d = PlainDict()
    try:
        d["nope"]
        expect(True).to_be(False)
    except KeyError:
        expect(True).to_be(True)

test("without __missing__ KeyError propagates", test_no_missing_keyerror)

# Test 8: __missing__ with defaultdict-like auto-vivification
def test_defaultdict_pattern():
    class DefaultDict:
        def __init__(self, factory):
            self._data = {}
            self._factory = factory

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            value = self._factory()
            self._data[key] = value
            return value

        def __setitem__(self, key, value):
            self._data[key] = value

    d = DefaultDict(list)
    d["a"].append(1)
    d["a"].append(2)
    d["b"].append(3)
    expect(d["a"]).to_be([1, 2])
    expect(d["b"]).to_be([3])

test("defaultdict-like pattern with __missing__", test_defaultdict_pattern)

# Test 9: __missing__ with counting
def test_counting_missing():
    class CountingDict:
        def __init__(self):
            self._data = {}
            self.miss_count = 0

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            self.miss_count += 1
            return None

    d = CountingDict()
    d["a"]
    d["b"]
    d["c"]
    expect(d.miss_count).to_be(3)

test("__missing__ counting misses", test_counting_missing)

# Test 10: __missing__ returning different types
def test_missing_return_types():
    class FlexDict:
        def __init__(self):
            self._data = {}

        def __getitem__(self, key):
            if key in self._data:
                return self._data[key]
            raise KeyError(key)

        def __missing__(self, key):
            if key == "int":
                return 0
            if key == "str":
                return ""
            if key == "list":
                return []
            return None

    d = FlexDict()
    expect(d["int"]).to_be(0)
    expect(d["str"]).to_be("")
    expect(d["list"]).to_be([])
    expect(d["other"]).to_be(None)

test("__missing__ returning different types", test_missing_return_types)
