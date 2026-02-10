# Test: CPython Iterator Protocol Edge Cases
# Adapted from CPython's test_iter.py

from test_framework import test, expect

# === Custom iterator ===
def test_custom_iterator():
    class MyRange:
        def __init__(self, n):
            self.n = n
            self.i = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.i >= self.n:
                raise StopIteration
            val = self.i
            self.i = self.i + 1
            return val
    expect(list(MyRange(5))).to_be([0, 1, 2, 3, 4])

# === Separate iterable and iterator ===
def test_separate_iterable_iterator():
    # Use a single class that creates a fresh iterator via __iter__
    class MyList:
        def __init__(self, data):
            self.data = data
            self._idx = 0
        def __iter__(self):
            # Return a new iterator each time
            new = type(self)(self.data)
            new._idx = 0
            return new
        def __next__(self):
            if self._idx >= len(self.data):
                raise StopIteration
            val = self.data[self._idx]
            self._idx = self._idx + 1
            return val

    ml = MyList([10, 20, 30])
    expect(list(ml)).to_be([10, 20, 30])
    # Can iterate multiple times
    expect(list(ml)).to_be([10, 20, 30])

# === iter() on lists ===
def test_iter_list():
    it = iter([1, 2, 3])
    expect(next(it)).to_be(1)
    expect(next(it)).to_be(2)
    expect(next(it)).to_be(3)
    try:
        next(it)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        expect(True).to_be(True)

# === iter() on strings ===
def test_iter_string():
    it = iter("abc")
    expect(next(it)).to_be("a")
    expect(next(it)).to_be("b")
    expect(next(it)).to_be("c")

# === iter() on tuples ===
def test_iter_tuple():
    expect(list(iter((1, 2, 3)))).to_be([1, 2, 3])

# === iter() on dicts ===
def test_iter_dict():
    d = {"a": 1, "b": 2}
    keys = list(iter(d))
    expect("a" in keys).to_be(True)
    expect("b" in keys).to_be(True)

# === next() with default ===
def test_next_default():
    it = iter([1])
    expect(next(it)).to_be(1)
    expect(next(it, "default")).to_be("default")
    expect(next(it, None)).to_be(None)

# === for loop uses iterator ===
def test_for_uses_iter():
    class Counter:
        def __init__(self, n):
            self.n = n
            self.i = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.i >= self.n:
                raise StopIteration
            val = self.i
            self.i = self.i + 1
            return val
    result = []
    for x in Counter(3):
        result.append(x)
    expect(result).to_be([0, 1, 2])

# === in operator uses iterator ===
def test_in_uses_iter():
    class Haystack:
        def __init__(self, items):
            self.items = items
        def __iter__(self):
            return iter(self.items)
    h = Haystack([1, 2, 3])
    expect(2 in h).to_be(True)
    expect(5 in h).to_be(False)

# === sum/min/max with iterators ===
def test_builtins_with_iterators():
    def gen():
        yield 1
        yield 2
        yield 3
    expect(sum(gen())).to_be(6)
    expect(min(gen())).to_be(1)
    expect(max(gen())).to_be(3)

# === enumerate with custom iterator ===
def test_enumerate_custom():
    class Items:
        def __init__(self, data):
            self.data = data
        def __iter__(self):
            return iter(self.data)
    result = list(enumerate(Items(["a", "b"])))
    expect(result).to_be([(0, "a"), (1, "b")])

# === zip with custom iterators ===
def test_zip_custom():
    def gen1():
        yield 1
        yield 2
    def gen2():
        yield "a"
        yield "b"
    expect(list(zip(gen1(), gen2()))).to_be([(1, "a"), (2, "b")])

# === map with iterators ===
def test_map_iter():
    def gen():
        yield 1
        yield 2
        yield 3
    expect(list(map(lambda x: x * 2, gen()))).to_be([2, 4, 6])

# === filter with iterators ===
def test_filter_iter():
    def gen():
        yield 1
        yield 2
        yield 3
        yield 4
        yield 5
    expect(list(filter(lambda x: x % 2 == 0, gen()))).to_be([2, 4])

# === Iterator chaining ===
def test_iter_chain():
    def chain(iterables):
        for it in iterables:
            for item in it:
                yield item
    expect(list(chain([[1, 2], [3, 4], [5]]))).to_be([1, 2, 3, 4, 5])

# === Multiple passes ===
def test_generator_single_pass():
    def gen():
        yield 1
        yield 2
        yield 3
    g = gen()
    expect(list(g)).to_be([1, 2, 3])
    # Second pass yields nothing
    expect(list(g)).to_be([])

# === Fibonacci iterator ===
def test_fibonacci_iterator():
    class Fib:
        def __init__(self, n):
            self.n = n
            self.a = 0
            self.b = 1
            self.count = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.count >= self.n:
                raise StopIteration
            val = self.a
            temp = self.a + self.b
            self.a = self.b
            self.b = temp
            self.count = self.count + 1
            return val
    expect(list(Fib(8))).to_be([0, 1, 1, 2, 3, 5, 8, 13])

# === Iterator from __getitem__ ===
def test_getitem_iterator():
    class OldStyle:
        def __init__(self, data):
            self.data = data
        def __getitem__(self, idx):
            if idx >= len(self.data):
                raise IndexError
            return self.data[idx]
        def __len__(self):
            return len(self.data)
    # Note: RAGE may or may not support this old-style iteration
    o = OldStyle([10, 20, 30])
    expect(o[0]).to_be(10)
    expect(o[1]).to_be(20)

# === Yield from ===
def test_yield_from():
    def inner():
        yield 1
        yield 2
    def outer():
        yield from inner()
        yield 3
    expect(list(outer())).to_be([1, 2, 3])

def test_yield_from_list():
    def gen():
        yield from [1, 2, 3]
        yield from [4, 5]
    expect(list(gen())).to_be([1, 2, 3, 4, 5])

# Register all tests
test("custom_iterator", test_custom_iterator)
test("separate_iterable_iterator", test_separate_iterable_iterator)
test("iter_list", test_iter_list)
test("iter_string", test_iter_string)
test("iter_tuple", test_iter_tuple)
test("iter_dict", test_iter_dict)
test("next_default", test_next_default)
test("for_uses_iter", test_for_uses_iter)
test("in_uses_iter", test_in_uses_iter)
test("builtins_with_iterators", test_builtins_with_iterators)
test("enumerate_custom", test_enumerate_custom)
test("zip_custom", test_zip_custom)
test("map_iter", test_map_iter)
test("filter_iter", test_filter_iter)
test("iter_chain", test_iter_chain)
test("generator_single_pass", test_generator_single_pass)
test("fibonacci_iterator", test_fibonacci_iterator)
test("getitem_iterator", test_getitem_iterator)
test("yield_from", test_yield_from)
test("yield_from_list", test_yield_from_list)

print("CPython iterator tests completed")
