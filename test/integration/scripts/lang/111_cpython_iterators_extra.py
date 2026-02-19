# Test: CPython Iterator Protocol Extra Tests
# Adapted from CPython's Lib/test/test_iter.py
# Covers: two-arg iter(), custom __iter__/__next__, StopIteration,
# iterator idempotency, builtin consumption, zip/map/filter with iterators,
# unpacking, in/not in with iterators

from test_framework import test, expect

# ============================================================================
# Helper classes (adapted from CPython test_iter.py)
# ============================================================================

class SequenceClass:
    """Class implementing __getitem__ for old-style iteration."""
    def __init__(self, n):
        self.n = n
    def __getitem__(self, i):
        if 0 <= i < self.n:
            return i
        raise IndexError

class IteratingSequenceClass:
    """Class implementing __iter__ + __next__."""
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

class IntsFrom:
    """Infinite iterator starting from a given value."""
    def __init__(self, start):
        self.i = start
    def __iter__(self):
        return self
    def __next__(self):
        i = self.i
        self.i = i + 1
        return i

# ============================================================================
# Sentinel-style iteration via manual implementation
# (RAGE's iter() only supports 1 arg; simulate two-arg iter behavior)
# ============================================================================

def test_sentinel_iteration_manual():
    """Simulate iter(callable, sentinel) with a generator."""
    def iter_sentinel(func, sentinel):
        while True:
            val = func()
            if val == sentinel:
                break
            yield val

    state = [0]
    def counter():
        val = state[0]
        state[0] = val + 1
        return val
    result = list(iter_sentinel(counter, 5))
    expect(result).to_be([0, 1, 2, 3, 4])

test("sentinel iteration (manual)", test_sentinel_iteration_manual)

def test_sentinel_immediate_stop():
    """Sentinel iteration stops immediately when first call returns sentinel."""
    def iter_sentinel(func, sentinel):
        while True:
            val = func()
            if val == sentinel:
                break
            yield val

    def always_zero():
        return 0
    result = list(iter_sentinel(always_zero, 0))
    expect(result).to_be([])

test("sentinel iteration immediate stop", test_sentinel_immediate_stop)

def test_sentinel_with_callable_class():
    """Sentinel iteration with __call__-able class."""
    def iter_sentinel(func, sentinel):
        while True:
            val = func()
            if val == sentinel:
                break
            yield val

    class CallableCounter:
        def __init__(self):
            self.i = 0
        def __call__(self):
            val = self.i
            self.i = val + 1
            return val
    result = list(iter_sentinel(CallableCounter(), 10))
    expect(result).to_be([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])

test("sentinel with callable class", test_sentinel_with_callable_class)

def test_sentinel_string():
    """Sentinel iteration with string sentinel."""
    def iter_sentinel(func, sentinel):
        while True:
            val = func()
            if val == sentinel:
                break
            yield val

    items = ["a", "b", "c", "STOP", "d"]
    idx = [0]
    def getter():
        val = items[idx[0]]
        idx[0] = idx[0] + 1
        return val
    result = list(iter_sentinel(getter, "STOP"))
    expect(result).to_be(["a", "b", "c"])

test("sentinel iteration with string sentinel", test_sentinel_string)

# ============================================================================
# Custom __iter__ and __next__ classes
# ============================================================================

def test_iterating_sequence_class():
    """IteratingSequenceClass produces correct values."""
    expect(list(IteratingSequenceClass(5))).to_be([0, 1, 2, 3, 4])

test("IteratingSequenceClass iteration", test_iterating_sequence_class)

def test_iterating_sequence_empty():
    """IteratingSequenceClass(0) produces empty."""
    expect(list(IteratingSequenceClass(0))).to_be([])

test("IteratingSequenceClass empty", test_iterating_sequence_empty)

def test_custom_iter_next_protocol():
    """Custom class with separate iterable and iterator."""
    class MyIterable:
        def __init__(self, data):
            self.data = data
        def __iter__(self):
            return MyIterator(self.data)

    class MyIterator:
        def __init__(self, data):
            self.data = data
            self.idx = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.idx >= len(self.data):
                raise StopIteration
            val = self.data[self.idx]
            self.idx = self.idx + 1
            return val

    it = MyIterable([10, 20, 30])
    expect(list(it)).to_be([10, 20, 30])
    # Can iterate multiple times because __iter__ returns new iterator
    expect(list(it)).to_be([10, 20, 30])

test("custom iterable with separate iterator", test_custom_iter_next_protocol)

def test_iterator_missing_next():
    """iter() on class with __iter__ but iterator has no __next__ raises error."""
    class BadIter:
        def __iter__(self):
            return object()
    try:
        for x in BadIter():
            pass
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("iterator missing __next__ raises error", test_iterator_missing_next)

# ============================================================================
# StopIteration from __next__
# ============================================================================

def test_stopiteration_from_next():
    """StopIteration from __next__ terminates iteration."""
    class StopsAtThree:
        def __init__(self):
            self.i = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.i >= 3:
                raise StopIteration
            val = self.i
            self.i = self.i + 1
            return val

    result = []
    for x in StopsAtThree():
        result.append(x)
    expect(result).to_be([0, 1, 2])

test("StopIteration from __next__ terminates for-loop", test_stopiteration_from_next)

def test_stopiteration_with_value():
    """StopIteration can carry a value."""
    class ValIter:
        def __init__(self):
            self.done = False
        def __iter__(self):
            return self
        def __next__(self):
            if self.done:
                raise StopIteration("finished")
            self.done = True
            return 42

    it = ValIter()
    expect(next(it)).to_be(42)
    try:
        next(it)
        expect("no error").to_be("StopIteration")
    except StopIteration:
        expect(True).to_be(True)

test("StopIteration with value", test_stopiteration_with_value)

def test_exception_during_iteration():
    """Non-StopIteration exception propagates out of for-loop."""
    class ErrorIter:
        def __init__(self):
            self.i = 0
        def __iter__(self):
            return self
        def __next__(self):
            if self.i == 3:
                raise RuntimeError("boom")
            val = self.i
            self.i = self.i + 1
            return val

    result = []
    got_error = False
    try:
        for x in ErrorIter():
            result.append(x)
    except RuntimeError:
        got_error = True
    expect(result).to_be([0, 1, 2])
    expect(got_error).to_be(True)

test("non-StopIteration exception propagates from for-loop", test_exception_during_iteration)

# ============================================================================
# Iterator idempotency: iter(iter(x)) is iter(x)
# ============================================================================

def test_iter_idempotency_list():
    """iter(iter(list)) should return the same iterator object."""
    seq = [1, 2, 3]
    it = iter(seq)
    it2 = iter(it)
    expect(it is it2).to_be(True)

test("iter(iter(list)) is same object", test_iter_idempotency_list)

def test_iter_idempotency_custom():
    """iter() on a custom iterator returns itself."""
    it = IteratingSequenceClass(5)
    it2 = iter(it)
    expect(it is it2).to_be(True)

test("iter(custom_iterator) returns itself", test_iter_idempotency_custom)

def test_iter_idempotency_generator():
    """iter() on a generator returns itself."""
    def gen():
        yield 1
    g = gen()
    g2 = iter(g)
    expect(g is g2).to_be(True)

test("iter(generator) returns itself", test_iter_idempotency_generator)

# ============================================================================
# For-loop over custom iterators
# ============================================================================

def test_for_loop_custom_class():
    """for-loop over IteratingSequenceClass."""
    result = []
    for x in IteratingSequenceClass(5):
        result.append(x)
    expect(result).to_be([0, 1, 2, 3, 4])

test("for-loop over IteratingSequenceClass", test_for_loop_custom_class)

def test_nested_for_loops():
    """Nested for-loops produce correct cartesian product."""
    result = []
    for i in IteratingSequenceClass(3):
        for j in IteratingSequenceClass(3):
            result.append((i, j))
    expected = [(0, 0), (0, 1), (0, 2), (1, 0), (1, 1), (1, 2), (2, 0), (2, 1), (2, 2)]
    expect(result).to_be(expected)

test("nested for-loops with custom iterators", test_nested_for_loops)

def test_for_loop_independence():
    """Independent iterations over range."""
    result = []
    seq = range(3)
    for i in iter(seq):
        for j in iter(seq):
            result.append((i, j))
    expected = [(0, 0), (0, 1), (0, 2), (1, 0), (1, 1), (1, 2), (2, 0), (2, 1), (2, 2)]
    expect(result).to_be(expected)

test("for-loop independence with range", test_for_loop_independence)

def test_for_loop_break():
    """Break exits custom iterator for-loop."""
    result = []
    for x in IteratingSequenceClass(10):
        if x == 5:
            break
        result.append(x)
    expect(result).to_be([0, 1, 2, 3, 4])

test("for-loop break with custom iterator", test_for_loop_break)

def test_for_loop_continue():
    """Continue skips values in custom iterator for-loop."""
    result = []
    for x in IteratingSequenceClass(6):
        if x % 2 == 0:
            continue
        result.append(x)
    expect(result).to_be([1, 3, 5])

test("for-loop continue with custom iterator", test_for_loop_continue)

# ============================================================================
# Builtin consumption: list(), tuple(), sorted(), sum(), min(), max()
# ============================================================================

def test_list_from_custom_iterator():
    """list() consumes a custom iterator."""
    expect(list(IteratingSequenceClass(5))).to_be([0, 1, 2, 3, 4])
    expect(list(IteratingSequenceClass(0))).to_be([])

test("list() from custom iterator", test_list_from_custom_iterator)

def test_tuple_from_custom_iterator():
    """tuple() consumes a custom iterator."""
    expect(tuple(IteratingSequenceClass(5))).to_be((0, 1, 2, 3, 4))
    expect(tuple(IteratingSequenceClass(0))).to_be(())

test("tuple() from custom iterator", test_tuple_from_custom_iterator)

def test_sorted_from_custom_iterator():
    """sorted() consumes a custom iterator and sorts."""
    class ReverseIter:
        def __init__(self, n):
            self.n = n
            self.i = n - 1
        def __iter__(self):
            return self
        def __next__(self):
            if self.i < 0:
                raise StopIteration
            val = self.i
            self.i = self.i - 1
            return val
    expect(sorted(ReverseIter(5))).to_be([0, 1, 2, 3, 4])

test("sorted() from custom iterator", test_sorted_from_custom_iterator)

def test_sum_from_custom_iterator():
    """sum() consumes a custom iterator."""
    expect(sum(IteratingSequenceClass(5))).to_be(10)
    expect(sum(IteratingSequenceClass(0))).to_be(0)

test("sum() from custom iterator", test_sum_from_custom_iterator)

def test_min_max_from_custom_iterator():
    """min() and max() consume custom iterators."""
    expect(min(IteratingSequenceClass(5))).to_be(0)
    expect(max(IteratingSequenceClass(5))).to_be(4)
    expect(max(IteratingSequenceClass(1))).to_be(0)
    expect(min(IteratingSequenceClass(1))).to_be(0)

test("min/max from custom iterator", test_min_max_from_custom_iterator)

def test_list_type_error():
    """list() on non-iterable raises an error."""
    try:
        list(42)
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("list() on non-iterable raises error", test_list_type_error)

def test_tuple_type_error():
    """tuple() on non-iterable raises an error."""
    try:
        tuple(42)
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("tuple() on non-iterable raises error", test_tuple_type_error)

def test_list_from_dict():
    """list() on dict returns keys."""
    d = {"one": 1, "two": 2, "three": 3}
    result = list(d)
    expect("one" in result).to_be(True)
    expect("two" in result).to_be(True)
    expect("three" in result).to_be(True)
    expect(len(result)).to_be(3)

test("list() from dict returns keys", test_list_from_dict)

def test_tuple_from_string():
    """tuple() on string returns chars."""
    expect(tuple("abc")).to_be(("a", "b", "c"))
    expect(tuple("")).to_be(())

test("tuple() from string", test_tuple_from_string)

def test_sum_from_generator():
    """sum() from a generator expression."""
    result = sum(x * x for x in range(5))
    expect(result).to_be(30)  # 0+1+4+9+16

test("sum() from generator expression", test_sum_from_generator)

def test_min_max_from_dict():
    """min/max on dict compares keys."""
    d = {"one": 1, "two": 2, "three": 3}
    expect(max(d)).to_be("two")
    expect(min(d)).to_be("one")

test("min/max from dict compares keys", test_min_max_from_dict)

def test_min_max_from_dict_values():
    """min/max on dict.values()."""
    d = {"one": 1, "two": 2, "three": 3}
    expect(max(d.values())).to_be(3)
    expect(min(d.values())).to_be(1)

test("min/max from dict.values()", test_min_max_from_dict_values)

# ============================================================================
# zip() with iterators
# ============================================================================

def test_zip_empty():
    """zip() with no args returns empty."""
    expect(list(zip())).to_be([])

test("zip() empty", test_zip_empty)

def test_zip_single_iterable():
    """zip(iterable) wraps elements in 1-tuples."""
    expect(list(zip(range(3)))).to_be([(0,), (1,), (2,)])

test("zip() single iterable", test_zip_single_iterable)

def test_zip_two_iterables():
    """zip(a, b) pairs elements."""
    expect(list(zip([1, 2, 3], ["a", "b", "c"]))).to_be([(1, "a"), (2, "b"), (3, "c")])

test("zip() two iterables", test_zip_two_iterables)

def test_zip_unequal_length():
    """zip() stops at shortest."""
    expect(list(zip([1, 2], [10, 20, 30]))).to_be([(1, 10), (2, 20)])
    expect(list(zip([1, 2, 3], [10]))).to_be([(1, 10)])

test("zip() unequal length stops at shortest", test_zip_unequal_length)

def test_zip_custom_iterators():
    """zip() with custom iterators."""
    expect(list(zip(IteratingSequenceClass(3)))).to_be([(0,), (1,), (2,)])

test("zip() with custom iterators", test_zip_custom_iterators)

def test_zip_with_infinite():
    """zip() with an infinite iterator stops at shorter."""
    result = list(zip(IntsFrom(0), [10, 20, 30]))
    expect(result).to_be([(0, 10), (1, 20), (2, 30)])

test("zip() with infinite iterator", test_zip_with_infinite)

def test_zip_three_iterables():
    """zip() with three iterables."""
    result = list(zip([1, 2], ["a", "b"], [True, False]))
    expect(result).to_be([(1, "a", True), (2, "b", False)])

test("zip() three iterables", test_zip_three_iterables)

def test_zip_type_error():
    """zip() with non-iterable raises an error."""
    try:
        list(zip(42))
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("zip() non-iterable raises error", test_zip_type_error)

def test_zip_dict_items():
    """zip(dict.keys(), dict.values()) matches dict.items()."""
    d = {"a": 1, "b": 2}
    keys = list(d.keys())
    vals = list(d.values())
    zipped = list(zip(keys, vals))
    items = list(d.items())
    expect(zipped).to_be(items)

test("zip(keys, values) matches items", test_zip_dict_items)

# ============================================================================
# map() and filter() with iterators
# ============================================================================

def test_map_custom_iterator():
    """map() over a custom iterator."""
    expect(list(map(lambda x: x + 1, IteratingSequenceClass(5)))).to_be([1, 2, 3, 4, 5])

test("map() over custom iterator", test_map_custom_iterator)

def test_map_multiple_iterables():
    """map() with two iterables."""
    result = list(map(lambda x, y: x + y, [1, 2, 3], [10, 20, 30]))
    expect(result).to_be([11, 22, 33])

test("map() with multiple iterables", test_map_multiple_iterables)

def test_map_empty():
    """map() with empty iterable."""
    expect(list(map(lambda x: x, []))).to_be([])

test("map() empty iterable", test_map_empty)

def test_filter_none():
    """filter(None, iterable) removes falsy values."""
    expect(list(filter(None, [0, 1, 2, 0, 3, 0]))).to_be([1, 2, 3])
    expect(list(filter(None, ["", "a", "", "b"]))).to_be(["a", "b"])

test("filter(None) removes falsy", test_filter_none)

def test_filter_custom_iterator():
    """filter() with a predicate over custom iterator."""
    result = list(filter(lambda x: x % 2 == 0, IteratingSequenceClass(6)))
    expect(result).to_be([0, 2, 4])

test("filter() over custom iterator", test_filter_custom_iterator)

def test_filter_with_bool_class():
    """filter() with class implementing __bool__."""
    class Boolean:
        def __init__(self, truth):
            self.truth = truth
        def __bool__(self):
            return self.truth
        def __eq__(self, other):
            if isinstance(other, Boolean):
                return self.truth == other.truth
            return False
        def __repr__(self):
            return "Boolean(" + str(self.truth) + ")"

    items = [Boolean(True), Boolean(False), Boolean(True), Boolean(False)]
    result = list(filter(None, items))
    expect(len(result)).to_be(2)
    expect(result[0].truth).to_be(True)
    expect(result[1].truth).to_be(True)

test("filter(None) with __bool__ class", test_filter_with_bool_class)

def test_filter_empty():
    """filter() on empty iterable."""
    expect(list(filter(lambda x: True, []))).to_be([])

test("filter() empty iterable", test_filter_empty)

def test_map_with_generator():
    """map() over a generator."""
    def gen():
        yield 1
        yield 2
        yield 3
    expect(list(map(lambda x: x * 10, gen()))).to_be([10, 20, 30])

test("map() over generator", test_map_with_generator)

def test_filter_with_generator():
    """filter() over a generator."""
    def gen():
        yield 1
        yield 2
        yield 3
        yield 4
        yield 5
    expect(list(filter(lambda x: x > 3, gen()))).to_be([4, 5])

test("filter() over generator", test_filter_with_generator)

# ============================================================================
# Unpacking iterators
# ============================================================================

def test_unpack_custom_iterator():
    """Unpack a custom iterator into variables."""
    a, b, c = IteratingSequenceClass(3)
    expect((a, b, c)).to_be((0, 1, 2))

test("unpack custom iterator", test_unpack_custom_iterator)

def test_unpack_too_many():
    """Unpacking more values than variables raises ValueError."""
    try:
        a, b = IteratingSequenceClass(3)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

test("unpack too many values raises ValueError", test_unpack_too_many)

def test_unpack_too_few():
    """Unpacking fewer values than variables raises ValueError."""
    try:
        a, b, c = IteratingSequenceClass(2)
        expect("no error").to_be("ValueError")
    except ValueError:
        expect(True).to_be(True)

test("unpack too few values raises ValueError", test_unpack_too_few)

def test_unpack_non_iterable():
    """Unpacking non-iterable raises an error."""
    try:
        a, b, c = 42
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("unpack non-iterable raises error", test_unpack_non_iterable)

def test_unpack_dict_values():
    """Unpack dict values."""
    a, b, c = {1: 42, 2: 42, 3: 42}.values()
    expect(a).to_be(42)
    expect(b).to_be(42)
    expect(c).to_be(42)

test("unpack dict values", test_unpack_dict_values)

def test_unpack_nested():
    """Nested unpacking with iterators."""
    (a, b), (c,) = IteratingSequenceClass(2), [42]
    expect((a, b, c)).to_be((0, 1, 42))

test("nested unpacking with iterators", test_unpack_nested)

def test_unpack_star():
    """Star unpacking with iterators."""
    a, *b, c = [1, 2, 3, 4, 5]
    expect(a).to_be(1)
    expect(b).to_be([2, 3, 4])
    expect(c).to_be(5)

test("star unpacking", test_unpack_star)

def test_unpack_star_generator():
    """Star unpacking with a generator."""
    def gen():
        yield 10
        yield 20
        yield 30
        yield 40
    a, *mid, z = gen()
    expect(a).to_be(10)
    expect(mid).to_be([20, 30])
    expect(z).to_be(40)

test("star unpacking with generator", test_unpack_star_generator)

def test_unpack_star_empty_middle():
    """Star unpacking with empty middle."""
    a, *b, c = [1, 2]
    expect(a).to_be(1)
    expect(b).to_be([])
    expect(c).to_be(2)

test("star unpacking empty middle", test_unpack_star_empty_middle)

# ============================================================================
# in / not in with iterators
# ============================================================================

def test_in_custom_iterator():
    """'in' with custom iterating class."""
    for i in range(5):
        expect(i in IteratingSequenceClass(5)).to_be(True)
    expect(5 in IteratingSequenceClass(5)).to_be(False)
    expect(-1 in IteratingSequenceClass(5)).to_be(False)

test("in with custom iterator", test_in_custom_iterator)

def test_not_in_custom_iterator():
    """'not in' with custom iterating class."""
    expect(5 not in IteratingSequenceClass(5)).to_be(True)
    expect(3 not in IteratingSequenceClass(5)).to_be(False)

test("not in with custom iterator", test_not_in_custom_iterator)

def test_in_non_iterable():
    """'in' with non-iterable raises an error."""
    try:
        result = 3 in 12
        expect("no error").to_be("error")
    except Exception:
        expect(True).to_be(True)

test("in non-iterable raises error", test_in_non_iterable)

def test_in_dict_keys():
    """'in' on dict checks keys."""
    d = {"one": 1, "two": 2, "three": 3}
    for k in d:
        expect(k in d).to_be(True)
    expect("four" in d).to_be(False)
    expect(1 in d).to_be(False)

test("in dict checks keys", test_in_dict_keys)

def test_in_dict_values():
    """'in' on dict.values() checks values."""
    d = {"one": 1, "two": 2, "three": 3}
    expect(1 in d.values()).to_be(True)
    expect(4 in d.values()).to_be(False)

test("in dict.values()", test_in_dict_values)

def test_in_consumes_iterator():
    """'in' may short-circuit but consumes from the iterator."""
    it = IteratingSequenceClass(10)
    expect(3 in it).to_be(True)
    # After finding 3, the iterator has advanced past 3
    # Remaining values depend on implementation, but next should be > 3
    remaining = list(it)
    expect(3 not in remaining).to_be(True)

test("in consumes from iterator (short-circuit)", test_in_consumes_iterator)

# ============================================================================
# Comprehensions with iterators
# ============================================================================

def test_listcomp_custom_iterator():
    """List comprehension with custom iterator."""
    result = [x * 2 for x in IteratingSequenceClass(5)]
    expect(result).to_be([0, 2, 4, 6, 8])

test("list comprehension with custom iterator", test_listcomp_custom_iterator)

def test_nested_comprehension_iterators():
    """Nested comprehension with iterators over range."""
    seq = range(3)
    result = [(i, j) for i in iter(seq) for j in iter(seq)]
    expect(result).to_be([(0, 0), (0, 1), (0, 2),
                          (1, 0), (1, 1), (1, 2),
                          (2, 0), (2, 1), (2, 2)])

test("nested comprehension with iterators", test_nested_comprehension_iterators)

def test_setcomp_custom_iterator():
    """Set comprehension with custom iterator."""
    result = {x % 3 for x in IteratingSequenceClass(9)}
    expect(result).to_be({0, 1, 2})

test("set comprehension with custom iterator", test_setcomp_custom_iterator)

def test_dictcomp_custom_iterator():
    """Dict comprehension with custom iterator."""
    result = {x: x * x for x in IteratingSequenceClass(4)}
    expect(result).to_be({0: 0, 1: 1, 2: 4, 3: 9})

test("dict comprehension with custom iterator", test_dictcomp_custom_iterator)

# ============================================================================
# Generator-based iterator patterns
# ============================================================================

def test_generator_as_iterator():
    """Generator function returns a proper iterator."""
    def squares(n):
        for i in range(n):
            yield i * i
    result = list(squares(5))
    expect(result).to_be([0, 1, 4, 9, 16])

test("generator as iterator", test_generator_as_iterator)

def test_generator_send():
    """Generator send() method."""
    def accumulator():
        total = 0
        while True:
            val = yield total
            if val is None:
                break
            total = total + val

    gen = accumulator()
    expect(next(gen)).to_be(0)
    expect(gen.send(10)).to_be(10)
    expect(gen.send(20)).to_be(30)
    expect(gen.send(5)).to_be(35)

test("generator send()", test_generator_send)

def test_chained_generators():
    """Chained generators using yield from."""
    def gen1():
        yield 1
        yield 2
    def gen2():
        yield 3
        yield 4
    def chained():
        yield from gen1()
        yield from gen2()
    expect(list(chained())).to_be([1, 2, 3, 4])

test("chained generators with yield from", test_chained_generators)

# ============================================================================
# Edge cases
# ============================================================================

def test_single_element_iteration():
    """Iterator that yields exactly one element."""
    class SingleIter:
        def __init__(self):
            self.yielded = False
        def __iter__(self):
            return self
        def __next__(self):
            if self.yielded:
                raise StopIteration
            self.yielded = True
            return 42
    expect(list(SingleIter())).to_be([42])

test("single element iterator", test_single_element_iteration)

def test_empty_iterator():
    """Iterator that yields nothing."""
    class EmptyIter:
        def __iter__(self):
            return self
        def __next__(self):
            raise StopIteration
    expect(list(EmptyIter())).to_be([])

test("empty iterator", test_empty_iterator)

def test_next_default_on_exhausted():
    """next() with default on exhausted iterator."""
    it = iter([1])
    expect(next(it)).to_be(1)
    expect(next(it, "default")).to_be("default")
    expect(next(it, "still default")).to_be("still default")

test("next() default on exhausted iterator", test_next_default_on_exhausted)

def test_bool_from_iterator():
    """bool() does not consume iterators, but any/all do."""
    def gen():
        yield 0
        yield 1
        yield 2
    # any() finds first truthy
    expect(any(gen())).to_be(True)
    # all() finds first falsy
    expect(all(gen())).to_be(False)

test("any/all consume iterators correctly", test_bool_from_iterator)

def test_enumerate_with_custom_start():
    """enumerate() with start parameter on custom iterator."""
    result = list(enumerate(IteratingSequenceClass(3), 10))
    expect(result).to_be([(10, 0), (11, 1), (12, 2)])

test("enumerate with start parameter", test_enumerate_with_custom_start)

print("CPython iterator extra tests completed")
