from test_framework import test, expect

# Test 1: __iadd__ is called when defined
def test_iadd_called(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            self.val += other
            return self

    a = Acc(10)
    original = a
    a += 5
    expect(a.val).to_be(15)
    expect(a is original).to_be(True)  # same object, mutated in-place

test("__iadd__ is called and mutates in place", test_iadd_called)

# Test 2: Falls back to __add__ when __iadd__ not defined
def test_fallback_to_add(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __add__(self, other):
            return Num(self.val + other)

    a = Num(10)
    original = a
    a += 5
    expect(a.val).to_be(15)
    expect(a is original).to_be(False)  # new object from __add__

test("falls back to __add__ when __iadd__ not defined", test_fallback_to_add)

# Test 3: __isub__
def test_isub(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __isub__(self, other):
            self.val -= other
            return self

    a = Acc(10)
    a -= 3
    expect(a.val).to_be(7)

test("__isub__", test_isub)

# Test 4: __imul__
def test_imul(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imul__(self, other):
            self.val *= other
            return self

    a = Acc(5)
    a *= 3
    expect(a.val).to_be(15)

test("__imul__", test_imul)

# Test 5: __itruediv__
def test_itruediv(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __itruediv__(self, other):
            self.val = self.val / other
            return self

    a = Acc(10)
    a /= 2
    expect(a.val).to_be(5.0)

test("__itruediv__", test_itruediv)

# Test 6: __ifloordiv__
def test_ifloordiv(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ifloordiv__(self, other):
            self.val = self.val // other
            return self

    a = Acc(10)
    a //= 3
    expect(a.val).to_be(3)

test("__ifloordiv__", test_ifloordiv)

# Test 7: __imod__
def test_imod(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imod__(self, other):
            self.val = self.val % other
            return self

    a = Acc(10)
    a %= 3
    expect(a.val).to_be(1)

test("__imod__", test_imod)

# Test 8: __ipow__
def test_ipow(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ipow__(self, other):
            self.val = self.val ** other
            return self

    a = Acc(2)
    a **= 10
    expect(a.val).to_be(1024)

test("__ipow__", test_ipow)

# Test 9: __imatmul__
def test_imatmul(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __imatmul__(self, other):
            self.val = self.val + other  # just test it's called
            return self

    a = Acc(1)
    a @= 2
    expect(a.val).to_be(3)

test("__imatmul__", test_imatmul)

# Test 10: __iand__
def test_iand(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __iand__(self, other):
            self.val = self.val & other
            return self

    a = Acc(0b1111)
    a &= 0b1010
    expect(a.val).to_be(0b1010)

test("__iand__", test_iand)

# Test 11: __ior__
def test_ior(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ior__(self, other):
            self.val = self.val | other
            return self

    a = Acc(0b1010)
    a |= 0b0101
    expect(a.val).to_be(0b1111)

test("__ior__", test_ior)

# Test 12: __ixor__
def test_ixor(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ixor__(self, other):
            self.val = self.val ^ other
            return self

    a = Acc(0b1111)
    a ^= 0b1010
    expect(a.val).to_be(0b0101)

test("__ixor__", test_ixor)

# Test 13: __ilshift__
def test_ilshift(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __ilshift__(self, other):
            self.val = self.val << other
            return self

    a = Acc(1)
    a <<= 4
    expect(a.val).to_be(16)

test("__ilshift__", test_ilshift)

# Test 14: __irshift__
def test_irshift(t):
    class Acc:
        def __init__(self, val):
            self.val = val
        def __irshift__(self, other):
            self.val = self.val >> other
            return self

    a = Acc(16)
    a >>= 4
    expect(a.val).to_be(1)

test("__irshift__", test_irshift)

# Test 15: List-like mutation pattern with __iadd__
def test_list_like_iadd(t):
    class MyList:
        def __init__(self):
            self.items = []
        def __iadd__(self, other):
            for item in other:
                self.items.append(item)
            return self

    lst = MyList()
    lst += [1, 2, 3]
    lst += [4, 5]
    expect(lst.items).to_be([1, 2, 3, 4, 5])

test("list-like __iadd__ mutation pattern", test_list_like_iadd)

# Test 16: Inherited inplace dunder
def test_inherited_iadd(t):
    class Base:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            self.val += other
            return self

    class Child(Base):
        pass

    c = Child(10)
    c += 5
    expect(c.val).to_be(15)

test("inherited __iadd__ from base class", test_inherited_iadd)

# Test 17: __iadd__ returning None makes variable None (no fallback)
def test_iadd_returns_none(t):
    class Num:
        def __init__(self, val):
            self.val = val
        def __iadd__(self, other):
            return None

    a = Num(10)
    a += 5
    expect(a).to_be(None)

test("__iadd__ returning None assigns None", test_iadd_returns_none)
