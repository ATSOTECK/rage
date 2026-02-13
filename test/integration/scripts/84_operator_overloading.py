# Test: Operator Overloading
# Tests custom __add__, __sub__, __mul__, __eq__, __lt__, etc.

from test_framework import test, expect

def test_add_sub():
    """__add__ and __sub__"""
    class Vector:
        def __init__(self, x, y):
            self.x = x
            self.y = y

        def __add__(self, other):
            return Vector(self.x + other.x, self.y + other.y)

        def __sub__(self, other):
            return Vector(self.x - other.x, self.y - other.y)

        def __eq__(self, other):
            return self.x == other.x and self.y == other.y

    v1 = Vector(1, 2)
    v2 = Vector(3, 4)
    v3 = v1 + v2
    expect(v3.x).to_be(4)
    expect(v3.y).to_be(6)

    v4 = v2 - v1
    expect(v4.x).to_be(2)
    expect(v4.y).to_be(2)

def test_mul():
    """__mul__ for scalar multiplication"""
    class Vector:
        def __init__(self, x, y):
            self.x = x
            self.y = y

        def __mul__(self, scalar):
            return Vector(self.x * scalar, self.y * scalar)

        def __rmul__(self, scalar):
            return self.__mul__(scalar)

    v = Vector(3, 4)
    v2 = v * 2
    expect(v2.x).to_be(6)
    expect(v2.y).to_be(8)

def test_comparison():
    """__lt__, __le__, __gt__, __ge__"""
    class Money:
        def __init__(self, amount):
            self.amount = amount

        def __lt__(self, other):
            return self.amount < other.amount

        def __le__(self, other):
            return self.amount <= other.amount

        def __gt__(self, other):
            return self.amount > other.amount

        def __ge__(self, other):
            return self.amount >= other.amount

        def __eq__(self, other):
            return self.amount == other.amount

    a = Money(10)
    b = Money(20)
    c = Money(10)

    expect(a < b).to_be(True)
    expect(b > a).to_be(True)
    expect(a <= c).to_be(True)
    expect(a >= c).to_be(True)
    expect(a == c).to_be(True)
    expect(a < c).to_be(False)

def test_neg_pos():
    """__neg__ and __pos__"""
    class Number:
        def __init__(self, value):
            self.value = value

        def __neg__(self):
            return Number(-self.value)

        def __pos__(self):
            return Number(abs(self.value))

    n = Number(5)
    neg = -n
    expect(neg.value).to_be(-5)

    n2 = Number(-3)
    pos = +n2
    expect(pos.value).to_be(3)

def test_str_repr():
    """__str__ and __repr__"""
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y

        def __str__(self):
            return "(" + str(self.x) + ", " + str(self.y) + ")"

        def __repr__(self):
            return "Point(" + str(self.x) + ", " + str(self.y) + ")"

    p = Point(1, 2)
    expect(str(p)).to_be("(1, 2)")
    expect(repr(p)).to_be("Point(1, 2)")

def test_len_bool():
    """__len__ and __bool__"""
    class Stack:
        def __init__(self):
            self._items = []

        def push(self, item):
            self._items.append(item)

        def __len__(self):
            return len(self._items)

        def __bool__(self):
            return len(self._items) > 0

    s = Stack()
    expect(len(s)).to_be(0)
    expect(bool(s)).to_be(False)

    s.push(1)
    s.push(2)
    expect(len(s)).to_be(2)
    expect(bool(s)).to_be(True)

def test_contains():
    """__contains__ for 'in' operator"""
    class Range:
        def __init__(self, start, end):
            self.start = start
            self.end = end

        def __contains__(self, item):
            return self.start <= item < self.end

    r = Range(1, 10)
    expect(5 in r).to_be(True)
    expect(10 in r).to_be(False)
    expect(0 in r).to_be(False)

def test_getitem_setitem():
    """__getitem__ and __setitem__"""
    class Matrix:
        def __init__(self, rows, cols):
            self.data = {}
            self.rows = rows
            self.cols = cols

        def __getitem__(self, key):
            return self.data.get(key, 0)

        def __setitem__(self, key, value):
            self.data[key] = value

    m = Matrix(3, 3)
    m[(0, 0)] = 1
    m[(1, 1)] = 5
    expect(m[(0, 0)]).to_be(1)
    expect(m[(1, 1)]).to_be(5)
    expect(m[(2, 2)]).to_be(0)

def test_iter():
    """__iter__ and __next__"""
    class Countdown:
        def __init__(self, start):
            self.start = start

        def __iter__(self):
            self.current = self.start
            return self

        def __next__(self):
            if self.current <= 0:
                raise StopIteration
            val = self.current
            self.current -= 1
            return val

    result = list(Countdown(5))
    expect(result).to_be([5, 4, 3, 2, 1])

def test_call():
    """__call__ makes objects callable"""
    class Multiplier:
        def __init__(self, factor):
            self.factor = factor

        def __call__(self, x):
            return x * self.factor

    double = Multiplier(2)
    triple = Multiplier(3)
    expect(double(5)).to_be(10)
    expect(triple(5)).to_be(15)

test("add_sub", test_add_sub)
test("mul", test_mul)
test("comparison", test_comparison)
test("neg_pos", test_neg_pos)
test("str_repr", test_str_repr)
test("len_bool", test_len_bool)
test("contains", test_contains)
test("getitem_setitem", test_getitem_setitem)
test("iter", test_iter)
test("call", test_call)
