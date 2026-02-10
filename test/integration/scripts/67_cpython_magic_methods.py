# Test: CPython Magic/Dunder Method Patterns
# Adapted from CPython tests - covers dunder method patterns
# beyond 27_dunder_methods.py using test_framework

from test_framework import test, expect

# =============================================================================
# __str__ and __repr__
# =============================================================================

class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __str__(self):
        return "(" + str(self.x) + ", " + str(self.y) + ")"

    def __repr__(self):
        return "Point(" + str(self.x) + ", " + str(self.y) + ")"

def test_str_repr():
    p = Point(3, 4)
    expect(str(p)).to_be("(3, 4)")
    expect(repr(p)).to_be("Point(3, 4)")

test("str_repr", test_str_repr)

def test_str_repr_different():
    p = Point(0, 0)
    expect(str(p)).to_be("(0, 0)")
    expect(repr(p)).to_be("Point(0, 0)")
    # str and repr give different results
    expect(str(p) != repr(p)).to_be(True)

test("str_repr_different", test_str_repr_different)

# =============================================================================
# __len__ and __bool__
# =============================================================================

class Bag:
    def __init__(self, items):
        self._items = items

    def __len__(self):
        return len(self._items)

    def __bool__(self):
        return len(self._items) > 0

def test_len():
    b = Bag([1, 2, 3])
    expect(len(b)).to_be(3)
    b2 = Bag([])
    expect(len(b2)).to_be(0)

test("len", test_len)

def test_bool():
    b = Bag([1, 2])
    expect(bool(b)).to_be(True)
    b2 = Bag([])
    expect(bool(b2)).to_be(False)

test("bool", test_bool)

# =============================================================================
# __bool__ falls back to __len__ when __bool__ not defined
# =============================================================================

class LenOnly:
    def __init__(self, n):
        self._n = n

    def __len__(self):
        return self._n

def test_bool_fallback_to_len():
    expect(bool(LenOnly(5))).to_be(True)
    expect(bool(LenOnly(0))).to_be(False)

test("bool_fallback_to_len", test_bool_fallback_to_len)

# =============================================================================
# __getitem__ and __setitem__
# =============================================================================

class DictLike:
    def __init__(self):
        self._data = {}

    def __getitem__(self, key):
        return self._data[key]

    def __setitem__(self, key, value):
        self._data[key] = value

    def __len__(self):
        return len(self._data)

def test_getitem_setitem():
    d = DictLike()
    d["a"] = 1
    d["b"] = 2
    expect(d["a"]).to_be(1)
    expect(d["b"]).to_be(2)
    expect(len(d)).to_be(2)
    d["a"] = 10
    expect(d["a"]).to_be(10)

test("getitem_setitem", test_getitem_setitem)

# =============================================================================
# __contains__
# =============================================================================

class WordSet:
    def __init__(self, words):
        self._words = words

    def __contains__(self, item):
        for w in self._words:
            if w == item:
                return True
        return False

def test_contains():
    ws = WordSet(["hello", "world", "python"])
    expect("hello" in ws).to_be(True)
    expect("java" in ws).to_be(False)
    expect("world" in ws).to_be(True)

test("contains", test_contains)

# =============================================================================
# __iter__ and __next__
# =============================================================================

class CountUp:
    def __init__(self, limit):
        self._limit = limit
        self._current = 0

    def __iter__(self):
        return self

    def __next__(self):
        if self._current >= self._limit:
            raise StopIteration()
        val = self._current
        self._current = self._current + 1
        return val

def test_iter_next():
    result = []
    for x in CountUp(5):
        result.append(x)
    expect(result).to_be([0, 1, 2, 3, 4])

test("iter_next", test_iter_next)

def test_iter_to_list():
    expect(list(CountUp(4))).to_be([0, 1, 2, 3])
    expect(list(CountUp(0))).to_be([])

test("iter_to_list", test_iter_to_list)

# =============================================================================
# __call__ (callable objects)
# =============================================================================

class Multiplier:
    def __init__(self, factor):
        self.factor = factor

    def __call__(self, x):
        return x * self.factor

def test_call():
    double = Multiplier(2)
    triple = Multiplier(3)
    expect(double(5)).to_be(10)
    expect(triple(5)).to_be(15)
    expect(double(0)).to_be(0)

test("call", test_call)

def test_callable_as_higher_order():
    """Callable objects can be passed like functions."""
    double = Multiplier(2)
    items = [1, 2, 3, 4]
    result = list(map(double, items))
    expect(result).to_be([2, 4, 6, 8])

test("callable_as_higher_order", test_callable_as_higher_order)

# =============================================================================
# __eq__, __ne__, __lt__, __gt__, __le__, __ge__
# =============================================================================

class Score:
    def __init__(self, value):
        self.value = value

    def __eq__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value == other.value
        return self.value == other

    def __ne__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value != other.value
        return self.value != other

    def __lt__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value < other.value
        return self.value < other

    def __gt__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value > other.value
        return self.value > other

    def __le__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value <= other.value
        return self.value <= other

    def __ge__(self, other):
        if type(other).__name__ == type(self).__name__:
            return self.value >= other.value
        return self.value >= other

def test_eq_ne():
    s1 = Score(10)
    s2 = Score(10)
    s3 = Score(20)
    expect(s1 == s2).to_be(True)
    expect(s1 != s3).to_be(True)
    expect(s1 == s3).to_be(False)
    expect(s1 != s2).to_be(False)

test("eq_ne", test_eq_ne)

def test_lt_gt():
    s1 = Score(10)
    s2 = Score(20)
    expect(s1 < s2).to_be(True)
    expect(s2 > s1).to_be(True)
    expect(s2 < s1).to_be(False)
    expect(s1 > s2).to_be(False)

test("lt_gt", test_lt_gt)

def test_le_ge():
    s1 = Score(10)
    s2 = Score(10)
    s3 = Score(20)
    expect(s1 <= s2).to_be(True)
    expect(s1 >= s2).to_be(True)
    expect(s1 <= s3).to_be(True)
    expect(s3 >= s1).to_be(True)

test("le_ge", test_le_ge)

# =============================================================================
# __add__, __sub__, __mul__
# =============================================================================

class Vec2:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __add__(self, other):
        return Vec2(self.x + other.x, self.y + other.y)

    def __sub__(self, other):
        return Vec2(self.x - other.x, self.y - other.y)

    def __mul__(self, scalar):
        return Vec2(self.x * scalar, self.y * scalar)

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self.x == other.x and self.y == other.y

    def __str__(self):
        return "Vec2(" + str(self.x) + ", " + str(self.y) + ")"

def test_add():
    v1 = Vec2(1, 2)
    v2 = Vec2(3, 4)
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

test("add", test_add)

def test_sub():
    v1 = Vec2(5, 10)
    v2 = Vec2(2, 3)
    v3 = v1 - v2
    expect(v3.x).to_be(3)
    expect(v3.y).to_be(7)

test("sub", test_sub)

def test_mul():
    v = Vec2(3, 4)
    v2 = v * 2
    expect(v2.x).to_be(6)
    expect(v2.y).to_be(8)

test("mul", test_mul)

# =============================================================================
# __neg__, __pos__
# =============================================================================

class Number:
    def __init__(self, value):
        self.value = value

    def __neg__(self):
        return Number(-self.value)

    def __pos__(self):
        return Number(abs(self.value))

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self.value == other.value

def test_neg():
    n = Number(5)
    neg_n = -n
    expect(neg_n.value).to_be(-5)
    n2 = Number(-3)
    expect((-n2).value).to_be(3)

test("neg", test_neg)

def test_pos():
    n = Number(-7)
    pos_n = +n
    expect(pos_n.value).to_be(7)

test("pos", test_pos)

# =============================================================================
# __int__, __float__, __str__ conversions
# =============================================================================

class Temperature2:
    def __init__(self, celsius):
        self.celsius = celsius

    def __int__(self):
        return int(self.celsius)

    def __float__(self):
        return float(self.celsius)

    def __str__(self):
        return str(self.celsius) + "C"

def test_int_conversion():
    t = Temperature2(36.6)
    expect(int(t)).to_be(36)

test("int_conversion", test_int_conversion)

def test_float_conversion():
    t = Temperature2(100)
    # RAGE does not call __float__ via float(), so test __float__ directly
    expect(t.__float__()).to_be(100.0)

test("float_conversion", test_float_conversion)

def test_str_conversion():
    t = Temperature2(25)
    expect(str(t)).to_be("25C")

test("str_conversion", test_str_conversion)

# =============================================================================
# __hash__
# =============================================================================

class HashPoint:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __hash__(self):
        return self.x * 1000 + self.y

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self.x == other.x and self.y == other.y

def test_hash():
    p1 = HashPoint(1, 2)
    p2 = HashPoint(1, 2)
    p3 = HashPoint(3, 4)
    expect(hash(p1) == hash(p2)).to_be(True)
    expect(hash(p1) == hash(p3)).to_be(False)

test("hash", test_hash)

def test_hash_as_dict_key():
    p1 = HashPoint(1, 2)
    p2 = HashPoint(1, 2)
    d = {}
    d[p1] = "first"
    d[p2] = "second"  # should overwrite since p1 == p2
    expect(len(d)).to_be(1)
    expect(d[p1]).to_be("second")

test("hash_as_dict_key", test_hash_as_dict_key)

# =============================================================================
# __delitem__
# =============================================================================

class SimpleDict:
    def __init__(self):
        self._data = {}

    def __getitem__(self, key):
        return self._data[key]

    def __setitem__(self, key, value):
        self._data[key] = value

    def __delitem__(self, key):
        del self._data[key]

    def __len__(self):
        return len(self._data)

    def __contains__(self, key):
        return key in self._data

def test_delitem():
    d = SimpleDict()
    d["x"] = 1
    d["y"] = 2
    d["z"] = 3
    expect(len(d)).to_be(3)
    del d["y"]
    expect(len(d)).to_be(2)
    expect("y" in d).to_be(False)
    expect("x" in d).to_be(True)

test("delitem", test_delitem)

# =============================================================================
# __init__ patterns
# =============================================================================

class DefaultInit:
    def __init__(self):
        self.x = 0
        self.y = 0
        self.name = "default"

class ParamInit:
    def __init__(self, x, y, name="custom"):
        self.x = x
        self.y = y
        self.name = name

def test_init_default():
    d = DefaultInit()
    expect(d.x).to_be(0)
    expect(d.y).to_be(0)
    expect(d.name).to_be("default")

test("init_default", test_init_default)

def test_init_params():
    p = ParamInit(1, 2)
    expect(p.x).to_be(1)
    expect(p.y).to_be(2)
    expect(p.name).to_be("custom")
    p2 = ParamInit(3, 4, "special")
    expect(p2.name).to_be("special")

test("init_params", test_init_params)

# =============================================================================
# Custom container using multiple dunders
# =============================================================================

class Matrix:
    def __init__(self, rows):
        self._rows = rows
        self._nrows = len(rows)
        self._ncols = len(rows[0]) if len(rows) > 0 else 0

    def __getitem__(self, row):
        return self._rows[row]

    def __len__(self):
        return self._nrows

    def __contains__(self, value):
        for row in self._rows:
            for item in row:
                if item == value:
                    return True
        return False

    def __eq__(self, other):
        if type(other).__name__ != type(self).__name__:
            return False
        return self._rows == other._rows

    def __str__(self):
        lines = []
        for row in self._rows:
            parts = []
            for item in row:
                parts.append(str(item))
            lines.append("[" + ", ".join(parts) + "]")
        return "[" + ", ".join(lines) + "]"

def test_custom_container():
    m = Matrix([[1, 2, 3], [4, 5, 6], [7, 8, 9]])
    expect(len(m)).to_be(3)
    expect(m[0]).to_be([1, 2, 3])
    expect(m[1]).to_be([4, 5, 6])
    expect(m[2][2]).to_be(9)
    expect(5 in m).to_be(True)
    expect(10 in m).to_be(False)

test("custom_container", test_custom_container)

def test_custom_container_equality():
    m1 = Matrix([[1, 2], [3, 4]])
    m2 = Matrix([[1, 2], [3, 4]])
    m3 = Matrix([[5, 6], [7, 8]])
    expect(m1 == m2).to_be(True)
    expect(m1 == m3).to_be(False)

test("custom_container_equality", test_custom_container_equality)

# =============================================================================
# __add__ chaining
# =============================================================================

def test_add_chaining():
    v1 = Vec2(1, 1)
    v2 = Vec2(2, 2)
    v3 = Vec2(3, 3)
    result = v1 + v2 + v3
    expect(result.x).to_be(6)
    expect(result.y).to_be(6)

test("add_chaining", test_add_chaining)

# =============================================================================
# Comparison with same values
# =============================================================================

def test_comparison_same_values():
    s1 = Score(10)
    s2 = Score(10)
    expect(s1 == s2).to_be(True)
    expect(s1 <= s2).to_be(True)
    expect(s1 >= s2).to_be(True)
    expect(s1 < s2).to_be(False)
    expect(s1 > s2).to_be(False)

test("comparison_same_values", test_comparison_same_values)

# =============================================================================
# __str__ used in string concatenation
# =============================================================================

def test_str_in_concatenation():
    p = Point(1, 2)
    result = "Point is: " + str(p)
    expect(result).to_be("Point is: (1, 2)")

test("str_in_concatenation", test_str_in_concatenation)

print("CPython magic methods tests completed")
