# Test: Rich Comparisons
# Tests that __ne__, __le__, __ge__ are dispatched directly on user classes.

from test_framework import test, expect

class Cmp:
    def __init__(self, val):
        self.val = val

    def __eq__(self, other):
        return self.val == other.val

    def __ne__(self, other):
        # Custom __ne__: always returns "custom_ne" string (truthy)
        return "custom_ne"

    def __lt__(self, other):
        return self.val < other.val

    def __le__(self, other):
        return self.val <= other.val

    def __gt__(self, other):
        return self.val > other.val

    def __ge__(self, other):
        return self.val >= other.val


def test_ne_dispatch():
    """__ne__ is dispatched directly, not synthesized from __eq__"""
    a = Cmp(1)
    b = Cmp(2)
    # __ne__ returns "custom_ne" (truthy), so != should be True
    result = a != b
    expect(result).to_be("custom_ne")

    # Even for equal values, our __ne__ returns "custom_ne"
    c = Cmp(1)
    result2 = a != c
    expect(result2).to_be("custom_ne")

test("__ne__ dispatch", test_ne_dispatch)


def test_le_dispatch():
    """__le__ is dispatched directly"""
    a = Cmp(3)
    b = Cmp(5)
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)
    expect(a <= Cmp(3)).to_be(True)

test("__le__ dispatch", test_le_dispatch)


def test_ge_dispatch():
    """__ge__ is dispatched directly"""
    a = Cmp(5)
    b = Cmp(3)
    expect(a >= b).to_be(True)
    expect(b >= a).to_be(False)
    expect(a >= Cmp(5)).to_be(True)

test("__ge__ dispatch", test_ge_dispatch)


class OnlyNe:
    """Class that only defines __ne__, not __eq__"""
    def __init__(self, val):
        self.val = val

    def __ne__(self, other):
        return self.val != other.val


def test_ne_without_eq():
    """__ne__ works independently of __eq__"""
    a = OnlyNe(1)
    b = OnlyNe(2)
    expect(a != b).to_be(True)

    c = OnlyNe(1)
    expect(a != c).to_be(False)

test("__ne__ without __eq__", test_ne_without_eq)


class OnlyLe:
    """Class with only __le__"""
    def __init__(self, val):
        self.val = val

    def __le__(self, other):
        return self.val <= other.val


def test_le_only():
    """__le__ works without __lt__"""
    a = OnlyLe(1)
    b = OnlyLe(2)
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)

test("__le__ only", test_le_only)


class OnlyGe:
    """Class with only __ge__"""
    def __init__(self, val):
        self.val = val

    def __ge__(self, other):
        return self.val >= other.val


def test_ge_only():
    """__ge__ works without __gt__"""
    a = OnlyGe(5)
    b = OnlyGe(3)
    expect(a >= b).to_be(True)
    expect(b >= a).to_be(False)

test("__ge__ only", test_ge_only)


class ReflectedGe:
    """Class with __ge__ used as reflected __le__"""
    def __init__(self, val):
        self.val = val

    def __ge__(self, other):
        if not isinstance(other, ReflectedGe):
            return NotImplemented
        return self.val >= other.val


def test_reflected_le_via_ge():
    """a <= b tries b.__ge__(a) when a has no __le__"""
    a = ReflectedGe(2)
    b = ReflectedGe(5)
    # a <= b: a has no __le__, so tries b.__ge__(a) => 5 >= 2 => True
    expect(a <= b).to_be(True)
    expect(b <= a).to_be(False)

test("reflected __le__ via __ge__", test_reflected_le_via_ge)


def test_ne_in_if():
    """__ne__ dispatch works in if-statement (OpCompareNeJump)"""
    a = Cmp(1)
    b = Cmp(1)
    # __ne__ returns "custom_ne" (truthy), so even equal values are 'not equal'
    if a != b:
        result = "took_branch"
    else:
        result = "did_not"
    expect(result).to_be("took_branch")

test("__ne__ in if-statement", test_ne_in_if)
