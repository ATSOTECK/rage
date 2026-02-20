from test_framework import test, expect

# Test 1: Basic __reversed__ returning an iterator
def test_basic_reversed(t):
    class MySeq:
        def __init__(self, items):
            self.items = items
        def __reversed__(self):
            return iter(list(reversed(self.items)))

    s = MySeq([1, 2, 3])
    result = list(reversed(s))
    expect(result).to_be([3, 2, 1])

test("basic __reversed__ returning iterator", test_basic_reversed)

# Test 2: __reversed__ returning a list (iterable)
def test_reversed_returns_list(t):
    class MySeq:
        def __init__(self, items):
            self.items = items
        def __reversed__(self):
            return list(reversed(self.items))

    s = MySeq([10, 20, 30])
    result = list(reversed(s))
    expect(result).to_be([30, 20, 10])

test("__reversed__ returning a list", test_reversed_returns_list)

# Test 3: Falls back to sequence protocol when __reversed__ not defined
def test_fallback_no_reversed(t):
    result = list(reversed([1, 2, 3]))
    expect(result).to_be([3, 2, 1])

test("falls back to sequence protocol without __reversed__", test_fallback_no_reversed)

# Test 4: __reversed__ inherited from base class
def test_reversed_inherited(t):
    class Base:
        def __init__(self, items):
            self.items = items
        def __reversed__(self):
            return iter(list(reversed(self.items)))

    class Child(Base):
        pass

    c = Child([4, 5, 6])
    result = list(reversed(c))
    expect(result).to_be([6, 5, 4])

test("__reversed__ inherited from base class", test_reversed_inherited)

# Test 5: __reversed__ with empty sequence
def test_reversed_empty(t):
    class MySeq:
        def __init__(self, items):
            self.items = items
        def __reversed__(self):
            return iter(list(reversed(self.items)))

    s = MySeq([])
    result = list(reversed(s))
    expect(result).to_be([])

test("__reversed__ with empty sequence", test_reversed_empty)

# Test 6: __reversed__ with single element
def test_reversed_single(t):
    class MySeq:
        def __init__(self, items):
            self.items = items
        def __reversed__(self):
            return iter(list(reversed(self.items)))

    s = MySeq([42])
    result = list(reversed(s))
    expect(result).to_be([42])

test("__reversed__ with single element", test_reversed_single)

# Test 7: __reversed__ with strings
def test_reversed_strings(t):
    class Words:
        def __init__(self, words):
            self.words = words
        def __reversed__(self):
            return iter(list(reversed(self.words)))

    w = Words(["hello", "world", "foo"])
    result = list(reversed(w))
    expect(result).to_be(["foo", "world", "hello"])

test("__reversed__ with string elements", test_reversed_strings)

# Test 8: __reversed__ used in for loop
def test_reversed_in_for(t):
    class MyRange:
        def __init__(self, n):
            self.n = n
        def __reversed__(self):
            return iter(list(reversed(list(range(self.n)))))

    collected = []
    for x in reversed(MyRange(5)):
        collected.append(x)
    expect(collected).to_be([4, 3, 2, 1, 0])

test("__reversed__ used in for loop", test_reversed_in_for)

# Test 9: reversed() on tuple (builtin, no __reversed__)
def test_reversed_tuple(t):
    result = list(reversed((1, 2, 3)))
    expect(result).to_be([3, 2, 1])

test("reversed() on tuple", test_reversed_tuple)

# Test 10: reversed() on string
def test_reversed_string(t):
    result = list(reversed("abc"))
    expect(result).to_be(["c", "b", "a"])

test("reversed() on string", test_reversed_string)
